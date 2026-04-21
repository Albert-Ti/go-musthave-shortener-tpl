package handler

import (
	"net/http"
	"strings"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
)

func RedirectByKeyUrl(shortenUrlService *service.ShortenURLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		param := strings.TrimPrefix(r.URL.Path, "/")

		url := shortenUrlService.Get(param)
		if url == "" {
			http.Error(w, "Url not found", http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", url)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
