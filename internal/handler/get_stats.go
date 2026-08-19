package handler

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
)

func GetStats(svc *service.Service, trustSubnet string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if trustSubnet == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		ipXReal := r.Header.Get("X-Real-IP")

		if ipXReal == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		addr := r.RemoteAddr
		ip, _, err := net.SplitHostPort(addr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ipRemote := net.ParseIP(ip)
		fmt.Println("ipRemote", ipRemote)
		fmt.Println("ipXReal", ipXReal)

		stats, err := svc.GetStats(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
