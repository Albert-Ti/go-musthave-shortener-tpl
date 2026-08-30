package handler

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
)

func GetStats(svc *service.Service, trustSubnet string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if trustSubnet == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		_, ipNet, err := net.ParseCIDR(trustSubnet)
		if err != nil {
			http.Error(w, "invalid trusted subnet", http.StatusInternalServerError)
			return
		}

		ipReq, err := resolveIP(r, r.Header.Get("X-Real-IP"))
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if !ipNet.Contains(ipReq) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

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

func resolveIP(r *http.Request, ipHeader string) (net.IP, error) {
	ip := net.ParseIP(ipHeader)
	if ip == nil {
		// если заголовок X-Real-IP пуст, пробуем X-Forwarded-For
		// этот заголовок содержит адреса отправителя и промежуточных прокси
		// в виде 203.0.113.195, 70.41.3.18, 150.172.238.178
		ips := r.Header.Get("X-Forwarded-For")
		// разделяем цепочку адресов
		ipStrs := strings.Split(ips, ",")
		// интересует только первый
		ipHeader = ipStrs[0]
		ip = net.ParseIP(ipHeader)
	}

	if ip != nil {
		return ip, nil
	}

	addr := r.RemoteAddr
	// метод возвращает адрес в формате host:port
	// нужна только подстрока host
	ipStr, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ipRemote := net.ParseIP(ipStr)
	if ipRemote == nil {
		return nil, fmt.Errorf("failed to parse ip from remote addr: %s", ipStr)
	}
	return ipRemote, nil
}
