package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
)

func DeleteShortenURLs(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var keys []string

		if err := json.NewDecoder(r.Body).Decode(&keys); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userID, err := middleware.GetAuthUserID(r.Context())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusAccepted)

		go func(keys []string, userID string) {
			if err := svc.BatchDelete(context.Background(), keys, userID); err != nil {
				slog.Error("batch delete failed", "error", err)
			}
		}(keys, userID)
	}
}
