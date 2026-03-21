package main

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
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

func createShortenedUrl(inputUrl string) (string, error) {
		if !validateUrl(inputUrl) {
			return "", errors.New("Invalid URL")
		}

		return inputUrl, nil
}



func mainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		inputUrl := r.FormValue("url") 

		shortUrl, err := createShortenedUrl(inputUrl)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)

		w.Write([]byte(shortUrl))
	} else {
		w.Write([]byte("GET"))
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(mainPage))

	err := http.ListenAndServe("localhost:8080", mux)

	if err != nil {
		panic(err)
	}
}
