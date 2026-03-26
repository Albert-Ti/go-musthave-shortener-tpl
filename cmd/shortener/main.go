package main

import (
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/handler"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
)

func main() {
	urls := &model.ShortenedUrls{
		List:  map[uint]string{},
		Count: 1,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.CreateShortUrl(urls))
	mux.HandleFunc("/{id}", handler.RedirectById(urls))

	err := http.ListenAndServe("localhost:8080", mux)

	if err != nil {
		panic(err)
	}
}
