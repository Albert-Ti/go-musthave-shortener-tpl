package handler

import (
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
)

func CreateShortenURLBatch(shortenUrlService *service.ShortenURLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Write([]byte("++++"))
		w.WriteHeader(http.StatusCreated)
	}
}
