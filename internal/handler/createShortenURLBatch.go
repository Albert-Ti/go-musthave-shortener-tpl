package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
)

func CreateShortenURLBatch(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if !strings.Contains(r.Header.Get("Content-type"), "application/json") {
			http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
			return
		}

		var dec []model.JSONBatchReq
		if err := json.NewDecoder(r.Body).Decode(&dec); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(dec) == 0 {
			http.Error(w, "Data is empty", http.StatusNoContent)
			return
		}

		ctx := r.Context()
		userID, err := middleware.GetAuthUserID(ctx)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		resp, err := svc.BatchSave(ctx, dec, userID)

		statusCode := http.StatusCreated

		if err != nil && errors.Is(err, repository.ErrConflict) {
			statusCode = http.StatusConflict
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(statusCode)

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
