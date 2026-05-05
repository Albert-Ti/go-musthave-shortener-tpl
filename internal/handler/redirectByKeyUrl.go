package handler

import (
	"net/http"
	"strings"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
)

func RedirectByKeyURL(shortenUrlService *service.ShortenURLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		param := strings.TrimPrefix(r.URL.Path, "/")

		url, err := shortenUrlService.Get(param)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if url == "" {
			http.Error(w, "URL not found", http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", url)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
