package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/validator"
)

func CreateShortUrlJSON(urlService *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Content-type") != "application/json" {
			http.Error(w, "Content-type not allowed", http.StatusBadRequest)
			return
		}

		var dec model.UrlRequest
		if err := json.NewDecoder(r.Body).Decode(&dec); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !validator.ValidateUrl(string(dec.Url)) {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		keyUrl := urlService.Set(dec.Url)
		fullUrl := fmt.Sprintf("%s/%s", config.Envs.BaseURL, keyUrl)

		resp := model.UrlResponse{Result: fullUrl}

		if err := json.NewEncoder(w).Encode(&resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
