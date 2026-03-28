package main

import (
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/handler"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	urls := &model.ShortenedUrls{
		List:  map[string]string{},
		Count: 1,
	}

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/", handler.CreateShortUrl(urls))
	r.Get("/{id}", handler.RedirectById(urls))

	err := http.ListenAndServe("localhost:8080", r)

	if err != nil {
		panic(err)
	}
}
