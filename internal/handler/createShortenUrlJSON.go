package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
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

		var req model.JSONReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !validator.ValidateURL(req.URL) {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		userID, err := middleware.GetAuthUserID(ctx)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		keyURL, isNew, err := svc.Save(ctx, req.URL, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fullURL := config.Envs.BaseURL + "/" + keyURL
		resp := model.JSONResp{Result: fullURL}

		w.Header().Set("Content-type", "application/json")

		if !isNew {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusCreated)
		}

		json.NewEncoder(w).Encode(resp)
	}
}
