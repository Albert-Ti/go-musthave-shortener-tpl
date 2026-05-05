package main

import (
	"log/slog"
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/handler"
	myMiddleware "github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	config.ParseFlag()

	repo, e := repository.NewShortenURLRepository()
	if e != nil {
		panic(e)
	}
	defer repo.Close()

	shortenUrlService := service.NewShortenURLService(repo)

	r := chi.NewRouter()

	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(myMiddleware.WithLogging)
	r.Use(myMiddleware.GzipCompress)

	r.Post("/", handler.CreateShortenURL(shortenUrlService))
	r.Get("/{id}", handler.RedirectByKeyURL(shortenUrlService))
	r.Get("/ping", handler.PingDatabase(shortenUrlService))
	r.Post("/api/shorten", handler.CreateShortenURLJSON(shortenUrlService))
	r.Post("/api/shorten/batch", handler.CreateShortenURLBatch(shortenUrlService))

	slog.Info("Running server", "host", config.Envs.RunAddr)

	err := http.ListenAndServe(config.Envs.RunAddr, r)

	if err != nil {
		panic(err)
	}
}
