package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/validator"
)

func CreateShortenUrl(shortenUrlService *service.ShortenURLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		defer func() {
			if err := r.Body.Close(); err != nil {
				slog.Error("Failed to close request body", "error", err)
			}
		}()

		if !validator.ValidateUrl(string(body)) {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		keyUrl := shortenUrlService.Set(string(body))

		w.Header().Set("Content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)

		fullUrl := fmt.Sprintf("%s/%s", config.Envs.BaseURL, keyUrl)

		w.Write([]byte(fullUrl))
	}
}
