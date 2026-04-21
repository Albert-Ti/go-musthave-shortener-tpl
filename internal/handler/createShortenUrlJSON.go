package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/validator"
)

func CreateShortenUrlJSON(shortenUrlService *service.ShortenURLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if !strings.Contains(r.Header.Get("Content-type"), "application/json") {
			http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
			return
		}

		var dec model.ShortenUrlRequest
		if err := json.NewDecoder(r.Body).Decode(&dec); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !validator.ValidateUrl(string(dec.Url)) {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		keyUrl := shortenUrlService.Set(dec.Url)
		fullUrl := fmt.Sprintf("%s/%s", config.Envs.BaseURL, keyUrl)

		resp := model.ShortenUrlResponse{Result: fullUrl}

		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(&resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
