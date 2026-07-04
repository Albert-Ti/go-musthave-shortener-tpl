package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/audit"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/validator"
)

func CreateShortenURL(svc *service.Service, auditor *audit.Auditor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		defer func() {
			if err := r.Body.Close(); err != nil {
				slog.Error("Failed to close request body", "error", err)
			}
		}()

		if !validator.ValidateURL(string(body)) {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		userID, err := middleware.GetAuthUserID(ctx)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		keyURL, isNew, err := svc.Save(ctx, string(body), userID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fullURL := fmt.Sprintf("%s/%s", config.Envs.BaseURL, keyURL)
		w.Header().Set("Content-type", "text/plain; charset=utf-8")

		if !isNew {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusCreated)
		}

		w.Write([]byte(fullURL))

		if !audit.IsDisabled() {
			auditor.Add(audit.AuditLog{
				Ts:     time.Now().Unix(),
				Action: auditor.Action[r.Method],
				UserID: userID,
				URL:    config.Envs.BaseURL + r.RequestURI,
			})
		}
	}
}
