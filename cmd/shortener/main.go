package main

import (
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/handler"
)

func main() {
	urls := &handler.ShortenedUrls{
		List:  map[uint]string{},
		Count: 1,
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handler.MainPage(urls)))

	err := http.ListenAndServe("localhost:8080", mux)

	if err != nil {
		panic(err)
	}
}
