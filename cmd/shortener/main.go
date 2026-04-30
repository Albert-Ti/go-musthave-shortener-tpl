package main

import (
	"log/slog"
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config/db"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/handler"
	myMiddleware "github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	config.ParseFlag()

	shortenURLStorage, e := repository.NewShortenURLStorage()
	if e != nil {
		panic(e)
	}
	defer shortenURLStorage.Close()

	shortenUrlService := service.NewShortenURLService(shortenURLStorage)

	r := chi.NewRouter()

	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(myMiddleware.WithLogging)
	r.Use(myMiddleware.GzipCompress)

	r.Post("/", handler.CreateShortenURL(shortenUrlService))
	r.Get("/{id}", handler.RedirectByKeyURL(shortenUrlService))
	r.Post("/api/shorten", handler.CreateShortenURLJSON(shortenUrlService))

	slog.Info("Running server", "host", config.Envs.RunAddr)

	db.Init()
	err := http.ListenAndServe(config.Envs.RunAddr, r)

	if err != nil {
		panic(err)
	}
}
