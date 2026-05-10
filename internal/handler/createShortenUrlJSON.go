package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/validator"
)

func CreateShortenURLJSON(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if !strings.Contains(r.Header.Get("Content-type"), "application/json") {
			http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
			return
		}

		var dec model.JSONReq
		if err := json.NewDecoder(r.Body).Decode(&dec); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !validator.ValidateURL(string(dec.URL)) {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		keyURL, err := svc.Save(dec.URL)
		fullURL := config.Envs.BaseURL + "/" + keyURL

		statusCode := http.StatusCreated

		if err != nil && errors.Is(err, repository.ErrConflict) {
			statusCode = http.StatusConflict
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(statusCode)

		if err := json.NewEncoder(w).Encode(&model.JSONResp{Result: fullURL}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
