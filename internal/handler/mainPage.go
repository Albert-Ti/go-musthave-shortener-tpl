package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type ShortenedUrls struct {
	List  map[uint]string
	Count uint
}

func (u *ShortenedUrls) getUrl(key string) string {
	for k, v := range u.List {
		if "key_"+strconv.Itoa(int(k)) == key {
			return v
		}
	}
	return ""
}
func (u *ShortenedUrls) set(url string) string {
	u.List[u.Count] = url

	k := u.Count
	u.Count++
	return "key_" + strconv.Itoa(int(k))
}

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

func createShortenedUrl(urls *ShortenedUrls, inputUrl string) (string, error) {
	if !validateUrl(inputUrl) {
		return "", errors.New("Invalid URL")
	}

	newKey := urls.set(inputUrl)

	return newKey, nil
}

func MainPage(urls *ShortenedUrls) http.HandlerFunc {
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

			fullUrl := inputUrl + "/" + shortUrl
			w.Write([]byte(fullUrl))
		} else {

			param := strings.TrimPrefix(r.URL.Path, "/")

			url := urls.getUrl(param)
			if url == "" {
				http.Error(w, "Ссылка по данному ключу не найдена", http.StatusBadRequest)
				return
			}

			w.Header().Set("Location", url)
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		}
	}
}
