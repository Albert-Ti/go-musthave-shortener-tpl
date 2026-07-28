package main

import (
	"log/slog"
	"net/http"

	_ "net/http/pprof"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/audit"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/handler"
	myMiddleware "github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
)

//go:generate go run ../reset

func main() {
	opts := config.Build()
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

	if opts.Mode == config.ModeDebug {
		slog.Info("Running server pprof", "host", "localhost:6060")
		go func() {
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				slog.Error("failed to Running server pprof", "error", err)
			}
		}()
	}

	slog.Info("Running server", "host", opts.RunAddr, "mode", opts.Mode)
	errRun := http.ListenAndServe(opts.RunAddr, r)
	if errRun != nil {
		panic(errRun)
	}

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
		slog.Info("Migrations applied successfully")
	case migrate.ErrNoChange:
		slog.Info("Database schema is up-to-date")
	default:
		return errMigrate
	}
	return nil
}
