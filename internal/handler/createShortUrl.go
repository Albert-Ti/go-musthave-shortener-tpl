package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
)

func validateUrl(inputUrl string) bool {
	inputUrl = strings.TrimSpace(inputUrl)

	if inputUrl == "" {
		return false
	}

	parsedUrl, err := url.ParseRequestURI(inputUrl)
	if err != nil {
		return false
	}

	if parsedUrl.Host == "" {
		return false
	}

	if parsedUrl.Scheme == "" {
		return false
	}

	return true
}

func CreateShortUrl(urls *model.ShortenedUrls) http.HandlerFunc {
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
		defer r.Body.Close()

		if !validateUrl(string(body)) {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		shortUrl := urls.Set(string(body))

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)

		schema := "http"
		if r.TLS != nil {
			schema = "https"
		}

		fullUrl := fmt.Sprintf("%s://%s/%s", schema, r.Host, shortUrl)
		w.Write([]byte(fullUrl))
	}
}
