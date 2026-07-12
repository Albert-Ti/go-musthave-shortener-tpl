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

func main() {
	cfg := config.Build()
	repo, err := repository.NewRepository(cfg)

	if err := runMigrations(cfg.DatabaseDSN); err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}
	defer repo.Close()

	svc := service.NewService(repo, cfg)
	r := chi.NewRouter()

	auditor, err := audit.NewAuditor(cfg.AuditFile, cfg.AuditURL)
	if err != nil {
		panic(err)
	}

	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(myMiddleware.WithLogging)
	r.Use(myMiddleware.GzipCompress)
	r.Use(myMiddleware.AuthGuard(cfg.JWTSecretKey))

	r.Post("/", handler.CreateShortenURL(svc, auditor, cfg.BaseURL))
	r.Get("/{id}", handler.RedirectByKeyURL(svc, auditor, cfg.BaseURL))
	r.Post("/api/shorten", handler.CreateShortenURLJSON(svc, auditor, cfg.BaseURL))
	r.Post("/api/shorten/batch", handler.CreateShortenURLBatch(svc))
	r.Get("/api/user/urls", handler.GetShortenURLs(svc))
	r.Delete("/api/user/urls", handler.DeleteShortenURLs(svc))
	r.Get("/ping", handler.PingDatabase(svc))

	if cfg.Mode == "debug" {
		slog.Info("Running server pprof", "host", "localhost:6060")
		go http.ListenAndServe("localhost:6060", nil)
	}

	slog.Info("Running server", "host", cfg.RunAddr)
	errRun := http.ListenAndServe(cfg.RunAddr, r)
	if errRun != nil {
		panic(errRun)
	}

}

func runMigrations(dsn string) error {
	if dsn == "" {
		return nil
	}
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	switch err {
	case nil:
		slog.Info("Migrations applied successfully")
	case migrate.ErrNoChange:
		slog.Info("Database schema is up-to-date")
	default:
		return err
	}
	return nil
}
