package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
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

func CreateShortUrl(urls *repository.ShortenedUrls) http.HandlerFunc {
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

		w.Header().Set("Content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)

		fullUrl := fmt.Sprintf("%s/%s", config.FlagBaseURL, shortUrl)

		w.Write([]byte(fullUrl))
	}
}
