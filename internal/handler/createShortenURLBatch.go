package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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

		resp, err := svc.BatchSave(dec)
		if err != nil {
			if errors.Is(err, repository.ErrConflict) {
				w.Header().Set("Content-type", "application/json")
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(err.Error()))
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(resp) == 0 {
			http.Error(w, "Data is empty", http.StatusNoContent)
			return
		}

		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
