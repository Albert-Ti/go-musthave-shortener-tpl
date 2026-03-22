package handler

import (
	"errors"
	"fmt"
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

func createShortenedUrl(urls *model.ShortenedUrls, inputUrl string) (string, error) {
	if !validateUrl(inputUrl) {
		return "", errors.New("Invalid URL")
	}

	newKey := urls.Set(inputUrl)

	return newKey, nil
}

func MainPage(urls *model.ShortenedUrls) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {

			inputUrl := r.FormValue("url")
			shortUrl, err := createShortenedUrl(urls, inputUrl)

			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)

			schema := "http"
			if r.TLS != nil {
				schema = "https"
			}

			fullUrl := fmt.Sprintf("%s://%s/%s", schema, r.Host, shortUrl)
			w.Write([]byte(fullUrl))
		} else {

			param := strings.TrimPrefix(r.URL.Path, "/")

			url := urls.GetUrl(param)
			if url == "" {
				http.Error(w, "Ссылка по данному ключу не найдена", http.StatusBadRequest)
				return
			}

			w.Header().Set("Location", url)
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		}
	}
}
