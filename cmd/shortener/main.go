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

	repo, e := repository.NewRepository()

	if e != nil {
		panic(e)
	}
	defer repo.Close()

	svc := service.NewService(repo)

	r := chi.NewRouter()

	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(myMiddleware.WithLogging)
	r.Use(myMiddleware.GzipCompress)
	r.Use(myMiddleware.AuthGuard)

	auditor, e := audit.NewAuditor()
	if e != nil {
		panic(e)
	}

	r.Post("/", auditor.Observer(handler.CreateShortenURL(svc)))
	r.Get("/{id}", auditor.Observer(handler.RedirectByKeyURL(svc)))
	r.Post("/api/shorten", auditor.Observer(handler.CreateShortenURLJSON(svc)))

	r.Get("/ping", handler.PingDatabase(svc))
	r.Post("/api/shorten/batch", handler.CreateShortenURLBatch(svc))
	r.Get("/api/user/urls", handler.GetShortenURLs(svc))
	r.Delete("/api/user/urls", handler.DeleteShortenURLs(svc))

	slog.Info("Running server", "host", config.Envs.RunAddr)

	err := http.ListenAndServe(config.Envs.RunAddr, r)

	if err != nil {
		panic(err)
	}
}

func runMigrations(dsn string) error {
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
