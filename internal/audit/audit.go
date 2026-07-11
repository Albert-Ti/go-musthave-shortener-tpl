package audit

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
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
	Action      map[string]string
	Subscribers []Subscriber
}

func NewAuditor(auditFile string, auditURL string) (*Auditor, error) {
	slog.Info("Using Auditor")

	auditor := &Auditor{
		Ch:     make(chan AuditLog, 20),
		Action: map[string]string{http.MethodGet: "follow", http.MethodPost: "shorten"},
	}

	if auditFile != "" {
		fileObs, err := NewFileObserver(auditFile)
		if err != nil {
			return nil, err
		}
		auditor.Subscribe(fileObs)
	}

	if auditURL != "" {
		auditor.Subscribe(NewHTTPObserver(auditURL))
	}

	go auditor.Broadcast()

	return auditor, nil
}

func (a *Auditor) Subscribe(sub Subscriber) {
	a.Subscribers = append(a.Subscribers, sub)
}

func (a *Auditor) Broadcast() {
	for log := range a.Ch {
		for _, sub := range a.Subscribers {
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
