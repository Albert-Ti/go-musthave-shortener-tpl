package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
)

type AuditLog struct {
	Ts     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}

type Subscriber interface {
	Notify(log AuditLog)
}

type Auditor struct {
	Ch          chan AuditLog
	action      map[string]string
	subscribers []Subscriber
}

func NewAuditor() (*Auditor, error) {
	auditor := &Auditor{
		Ch:     make(chan AuditLog, 20),
		action: map[string]string{http.MethodGet: "follow", http.MethodPost: "shorten"},
	}

	if config.Envs.AuditFile != "" {
		fileObs, err := NewFileObserver(config.Envs.AuditFile)
		if err != nil {
			return nil, err
		}
		auditor.Subscribe(fileObs)
	}

	if config.Envs.AuditURL != "" {
		auditor.Subscribe(NewHTTPObserver(config.Envs.AuditURL))
	}

	go auditor.Broadcast()

	return auditor, nil
}

func (a *Auditor) Subscribe(sub Subscriber) {
	a.subscribers = append(a.subscribers, sub)
}

func (a *Auditor) Broadcast() {
	for log := range a.Ch {
		for _, sub := range a.subscribers {
			sub.Notify(log)
		}
	}
}

func (a *Auditor) Add(log AuditLog) {
	a.Ch <- log
}

type FileObserver struct {
	file    *os.File
	encoder *json.Encoder
}

type HTTPObserver struct {
	url string
}

func NewFileObserver(path string) (*FileObserver, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &FileObserver{file: file, encoder: json.NewEncoder(file)}, nil
}

func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{url}
}

func (f *FileObserver) Notify(log AuditLog) {
	if err := f.encoder.Encode(&log); err != nil {
		slog.Error("FileObserver", "encode", err)
	}
}

func (h *HTTPObserver) Notify(log AuditLog) {
	b, err := json.Marshal(log)
	if err != nil {
		slog.Error("HTTPObserver", "marshal", err)
		return
	}

	if _, err := http.Post(h.url, "application/json", bytes.NewReader(b)); err != nil {
		slog.Error("HTTPObserver", "post", err)
	}
}

func (a *Auditor) Observer(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if config.Envs.AuditFile == "" && config.Envs.AuditURL == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		userID, err := GetAuthUserID(ctx)

		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		auditLog := AuditLog{
			Ts:     time.Now().Unix(),
			Action: a.action[r.Method],
			UserID: userID,
			URL:    config.Envs.BaseURL + r.RequestURI,
		}

		a.Add(auditLog)

		next.ServeHTTP(w, r)
	}
}
