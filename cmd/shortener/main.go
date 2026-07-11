package main

import (
	"log/slog"
	"net/http"

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
	config.ParseFlag()

	if err := runMigrations(config.Envs.DatabaseDSN); err != nil {
		panic(err)
	}

	repo, err := repository.NewRepository()

	if err != nil {
		panic(err)
	}
	defer repo.Close()

	svc := service.NewService(repo)

	r := chi.NewRouter()

	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(myMiddleware.WithLogging)
	r.Use(myMiddleware.GzipCompress)
	r.Use(myMiddleware.AuthGuard)

	var auditor *audit.Auditor

	if !config.IsAuditorDisabled() {
		a, err := audit.NewAuditor(config.Envs.AuditFile, config.Envs.AuditURL)
		auditor = a
		if err != nil {
			panic(err)
		}
	}

	r.Post("/", handler.CreateShortenURL(svc, auditor))
	r.Get("/{id}", handler.RedirectByKeyURL(svc, auditor))
	r.Post("/api/shorten", handler.CreateShortenURLJSON(svc, auditor))

	r.Get("/ping", handler.PingDatabase(svc))
	r.Post("/api/shorten/batch", handler.CreateShortenURLBatch(svc))
	r.Get("/api/user/urls", handler.GetShortenURLs(svc))
	r.Delete("/api/user/urls", handler.DeleteShortenURLs(svc))

	slog.Info("Running server", "host", config.Envs.RunAddr)

	errRun := http.ListenAndServe(config.Envs.RunAddr, r)

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
