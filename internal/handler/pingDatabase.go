package handler

import (
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
)

func PingDatabase(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if err := svc.Ping(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
