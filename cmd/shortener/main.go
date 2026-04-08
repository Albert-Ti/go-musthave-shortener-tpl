package main

import (
	"fmt"
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/handler"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	config.ParseFlag()

	urls := &repository.ShortenedUrls{
		List:  map[string]string{},
		Count: 1,
	}

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/", handler.CreateShortUrl(urls))
	r.Get("/{id}", handler.RedirectById(urls))

	fmt.Println("Running server on", config.FlagRunAddr)
	err := http.ListenAndServe(config.FlagRunAddr, r)

	if err != nil {
		panic(err)
	}
}
