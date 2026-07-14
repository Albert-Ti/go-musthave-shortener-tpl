package audit

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type observer interface {
	Notify(log AuditLog)
}

type AuditLog struct {
	Ts     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}

type Auditor struct {
	ch       chan AuditLog
	action   map[string]string
	observer []observer
}

func NewAuditor(auditFile string, auditURL string) (*Auditor, error) {
	if auditFile == "" && auditURL == "" {
		return nil, nil
	}

	auditor := &Auditor{
		ch:     make(chan AuditLog, 20),
		action: map[string]string{http.MethodGet: "follow", http.MethodPost: "shorten"},
	}

	slog.Info("Using Auditor")

	if auditFile != "" {
		fileObs, err := NewFileObserver(auditFile)
		if err != nil {
			return nil, err
		}
		auditor.subscribe(fileObs)
	}

	if auditURL != "" {
		auditor.subscribe(NewHTTPObserver(auditURL))
	}

	go auditor.broadcast()

	return auditor, nil
}

func (a *Auditor) AddLog(method, requestURI, userID, baseURL string) {
	log := AuditLog{
		Ts:     time.Now().Unix(),
		Action: a.action[method],
		UserID: userID,
		URL:    baseURL + requestURI,
	}

	a.ch <- log
}

func (a *Auditor) subscribe(sub observer) {
	a.observer = append(a.observer, sub)
}

func (a *Auditor) broadcast() {
	for log := range a.ch {
		for _, sub := range a.observer {
			sub.Notify(log)
		}
	}
}

type FileObserver struct {
	file    *os.File
	encoder *json.Encoder
}

func NewFileObserver(path string) (*FileObserver, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &FileObserver{file: file, encoder: json.NewEncoder(file)}, nil
}

func (f *FileObserver) Notify(log AuditLog) {
	if err := f.encoder.Encode(&log); err != nil {
		slog.Error("FileObserver", "encode", err)
	}
}

type HTTPObserver struct {
	url string
}

func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{url}
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
