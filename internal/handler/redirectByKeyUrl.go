package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/audit"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
)

func RedirectByKeyURL(svc *service.Service, auditor *audit.Auditor, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		param := strings.TrimPrefix(r.URL.Path, "/")

		ctx := r.Context()
		url, err := svc.Get(ctx, param)
		if err != nil {
			if errors.Is(err, repository.ErrStatusGone) {
				w.WriteHeader(http.StatusGone)
				return
			}
			if errors.Is(err, repository.ErrNoRows) {
				http.Error(w, "URL not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", url)

		userID, err := middleware.GetAuthUserID(ctx)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)

		if auditor != nil {
			auditor.AddLog(r.Method, r.RequestURI, userID, baseURL)
		}
	}
}
