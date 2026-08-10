package main

import (
	"context"
	"fmt"
	"log/slog"
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
	myMiddleware "github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
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
	r := chi.NewRouter()

	auditor, errAudit := audit.NewAuditor(opts.AuditFile, opts.AuditURL, 20, 100)
	if errAudit != nil {
		panic(errAudit)
	}
	defer auditor.Close()

	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(myMiddleware.WithLogging)
	r.Use(myMiddleware.GzipCompress)
	r.Use(myMiddleware.AuthGuard(opts.JWTSecretKey))

	r.Post("/", handler.CreateShortenURL(svc, auditor, opts.BaseURL))
	r.Get("/{id}", handler.RedirectByKeyURL(svc, auditor, opts.BaseURL))
	r.Post("/api/shorten", handler.CreateShortenURLJSON(svc, auditor, opts.BaseURL))
	r.Post("/api/shorten/batch", handler.CreateShortenURLBatch(svc))
	r.Get("/api/user/urls", handler.GetShortenURLs(svc))
	r.Delete("/api/user/urls", handler.DeleteShortenURLs(svc))
	r.Get("/ping", handler.PingDatabase(svc))

	runPprof(opts.Mode)

	srv := &http.Server{Addr: opts.RunAddr}
	idleConnsClosed := make(chan struct{})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-sigs
		signal.Stop(sigs)
		slog.Info("received", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("server shutdown", "error", err)
		}

		close(idleConnsClosed)
	}()

	runServer(opts, srv)
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

func runPprof(mode string) {
	if mode == config.ModeDebug {
		slog.Info("running server pprof", "host", "localhost:6060")

		go func() {
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				slog.Error("failed to Running server pprof", "error", err)
			}
		}()
	}
}

func runServer(opts *config.Options, srv *http.Server) {
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
		errSrv = srv.ListenAndServeTLS("cert.pem", "private.pem")
	} else {
		slog.Info("running server", "host", host)
		errSrv = srv.ListenAndServe()
	}

	if errSrv != nil && errSrv != http.ErrServerClosed {
		panic(errSrv)
	}
}
