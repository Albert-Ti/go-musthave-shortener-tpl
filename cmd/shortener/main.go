package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "embed"
	_ "net/http/pprof"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/audit"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/cert"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/handler"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/interceptor"
	myMiddleware "github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"

	pb "github.com/Albert-Ti/go-musthave-shortener-tpl/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

//go:generate go run ../reset

func main() {
	fmt.Printf("\n  Build version: %s\n  Build date: %s\n  Build commit: %s\n \n",
		buildVersion, buildDate, buildCommit)

	opts, errCfg := config.Build()
	if errCfg != nil {
		panic(errCfg)
	}

	repo, errRepo := repository.NewRepository(opts)
	if errRepo != nil {
		panic(errRepo)
	}
	defer func() {
		if closeErr := repo.Close(); closeErr != nil {
			panic(closeErr)
		}
	}()

	if errMigrate := runMigrations(opts.DatabaseDSN); errMigrate != nil {
		panic(errMigrate)
	}

	svc := service.NewService(repo, opts)
	auditor, errAudit := audit.NewAuditor(opts.AuditFile, opts.AuditURL, 20, 100)
	if errAudit != nil {
		panic(errAudit)
	}
	defer auditor.Close()

	shortener := &shortener{}

	go shortener.pprofStart(opts.Mode)
	go shortener.httpStart(svc, auditor, opts)
	go shortener.grpcStart(svc, auditor, opts)

	idleConnsClosed := make(chan struct{})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-sigs
	signal.Stop(sigs)
	slog.Info("received", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	shortener.grpc.server.GracefulStop()
	if err := shortener.http.shutdown(ctx); err != nil {
		slog.Error("server shutdown", "error", err)
	}

	close(idleConnsClosed)

	<-idleConnsClosed

	slog.Info("server shutdown gracefully")
}

func runMigrations(dsn string) error {
	if dsn == "" {
		return nil
	}
	m, errMigrate := migrate.New("file://migrations", dsn)
	if errMigrate != nil {
		return errMigrate
	}
	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			slog.Error("failed to close migrate instance", "source_error", srcErr, "db_error", dbErr)
		}
	}()

	errMigrate = m.Up()
	switch errMigrate {
	case nil:
		slog.Info("migrations applied successfully")
	case migrate.ErrNoChange:
		slog.Info("database schema is up-to-date")
	default:
		return errMigrate
	}
	return nil
}

type shortener struct {
	http *httpServer
	grpc *grpcServer
}

func (s *shortener) httpStart(svc *service.Service, auditor *audit.Auditor, opts *config.Options) {
	r := chi.NewRouter()
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(myMiddleware.WithLogging)
	r.Use(myMiddleware.GzipCompress)
	r.Use(myMiddleware.AuthGuard(opts.JWTSecretKey, opts.DatabaseDSN == ""))

	r.Post("/", handler.CreateShortenURL(svc, auditor, opts.BaseURL))
	r.Get("/{id}", handler.RedirectByKeyURL(svc, auditor, opts.BaseURL))
	r.Post("/api/shorten", handler.CreateShortenURLJSON(svc, auditor, opts.BaseURL))
	r.Post("/api/shorten/batch", handler.CreateShortenURLBatch(svc))
	r.Get("/api/user/urls", handler.GetShortenURLs(svc))
	r.Delete("/api/user/urls", handler.DeleteShortenURLs(svc))
	r.Get("/api/internal/stats", handler.GetStats(svc, opts.TrustedSubnet))
	r.Get("/ping", handler.PingDatabase(svc))

	s.http = &httpServer{
		Server: &http.Server{Addr: opts.RunAddr,
			Handler: r},
	}

	host := "http://" + opts.RunAddr
	var errSrv error

	if opts.EnableHTTPS {
		host = "https://" + opts.RunAddr
		if !cert.IsCertValid() {
			slog.Info("certificate missing or expired, generating a new one")
			if errCert := cert.CreateCert(); errCert != nil {
				panic(errCert)
			}
		}
		slog.Info("running server", "host", host)
		errSrv = s.http.ListenAndServeTLS("cert.pem", "private.pem")
	} else {
		slog.Info("running server", "host", host)
		errSrv = s.http.ListenAndServe()
	}

	if errSrv != nil && errSrv != http.ErrServerClosed {
		panic(errSrv)
	}
}

func (s *shortener) grpcStart(svc *service.Service, auditor *audit.Auditor, opts *config.Options) {
	listen, err := net.Listen("tcp", opts.GRPCRunAddr)
	if err != nil {
		panic(err)
	}

	var srv *grpc.Server
	host := "http://" + opts.GRPCRunAddr

	if opts.EnableHTTPS {
		host = "https://" + opts.GRPCRunAddr
		transportCreds, err := credentials.NewServerTLSFromFile("cert.pem", "private.pem")
		if err != nil {
			panic(err)
		}
		srv = grpc.NewServer(grpc.UnaryInterceptor(interceptor.Auth), grpc.Creds(transportCreds))
	} else {
		srv = grpc.NewServer(grpc.UnaryInterceptor(interceptor.Auth))
	}

	s.grpc = &grpcServer{
		server:  srv,
		svc:     svc,
		auditor: auditor,
	}

	pb.RegisterShortenerServiceServer(srv, s.grpc)

	slog.Info("running server", "host", host)
	if err := srv.Serve(listen); err != nil {
		panic(err)
	}
}

func (s *shortener) pprofStart(mode string) {
	if mode == config.ModeDebug {
		slog.Info("running server pprof", "host", "localhost:6060")

		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			slog.Error("failed to Running server pprof", "error", err)
		}
	}
}

type httpServer struct {
	*http.Server
}

func (h *httpServer) shutdown(ctx context.Context) error {
	return h.Shutdown(ctx)
}

type grpcServer struct {
	pb.UnimplementedShortenerServiceServer
	server  *grpc.Server
	auditor *audit.Auditor
	svc     *service.Service
}

func (g *grpcServer) ShortenURL(ctx context.Context, in *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	// g.Svc.Save(ctx, in.GetUrl(), )
	return nil, nil
}

func (g *grpcServer) ExpandURL(ctx context.Context, in *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	return nil, nil
}

func (g *grpcServer) ListUserURLs(ctx context.Context, empty *emptypb.Empty) (*pb.UserURLsResponse, error) {
	return nil, nil
}
