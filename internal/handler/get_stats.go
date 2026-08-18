package handler

import (
	"fmt"
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
)

func GetStats(svc *service.Service, trustSubnet string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if trustSubnet == "" {
			w.WriteHeader(http.StatusForbidden)
		}

		ip := r.Header.Get("X-Real-IP")

		fmt.Println(ip)
	}
}
