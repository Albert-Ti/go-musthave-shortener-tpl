package main

import (
	"log/slog"
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/handler"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	config.ParseFlag()

	urls := &repository.ShortenedUrls{
		List:  map[string]string{},
		Count: 1,
	}

	r := chi.NewRouter()

	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)

	r.Use(middleware.WithLogging)

	r.Post("/", handler.CreateShortUrl(urls))
	r.Get("/{id}", handler.RedirectById(urls))

	slog.Info("Running server", "host", config.Envs.RunAddr)

	err := http.ListenAndServe(config.Envs.RunAddr, r)

	if err != nil {
		panic(err)
	}
}
