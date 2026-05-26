package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

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

		w.WriteHeader(http.StatusAccepted)

		ctx := context.WithoutCancel(r.Context())

		go func() {
			if err := svc.BatchDelete(ctx, keys); err != nil {
				slog.Error("batch delete failed", "error", err)
			}
		}()
	}
}
