// При Написание аудита использовался pattern Наблюдатель(Observer).
package audit

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// observer получает уведомление о каждой новой записи аудита.
type Observer interface {
	Notify(log AuditLog)
}

// AuditLog - запись аудита одного запроса.
type AuditLog struct {
	Ts     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}

// Auditor рассылает записи аудита подписанным наблюдателям (observer)
// через буферизованный канал, не блокируя обработку HTTP-запроса.
type Auditor struct {
	ch       chan AuditLog
	action   map[string]string
	observer []Observer
}

// NewAuditor создаёт Auditor и подписывает на него FileObserver (если задан
// auditFile) и HTTPObserver (если задан auditURL). Если оба параметра пустые,
// возвращает (nil, nil) - аудит считается выключенным.
//
// Пример использования:
//
//	auditor, err := audit.NewAuditor(cfg.AuditFile, cfg.AuditURL)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	auditor.AddLog(http.MethodPost, r.RequestURI, userID, baseURL)
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

	// Горутина чтобы не блокировал вызов
	go auditor.broadcast()

	return auditor, nil
}

// AddLog формирует AuditLog из параметров запроса и отправляет его всем
// подписанным наблюдателям. Вызов не блокируется, пока канал не заполнен.
func (a *Auditor) AddLog(method, requestURI, userID, baseURL string) {
	log := AuditLog{
		Ts:     time.Now().Unix(),
		Action: a.action[method],
		UserID: userID,
		URL:    baseURL + requestURI,
	}

	select {
	case a.ch <- log:
	default:
		slog.Info("Channel is full")
	}
}

func (a *Auditor) subscribe(sub Observer) {
	a.observer = append(a.observer, sub)
}

func (a *Auditor) broadcast() {
	for _, sub := range a.observer {
		for log := range a.ch {
			go sub.Notify(log) // Отдельная горутина для каждого наблюдателя
		}
	}
}

// FileObserver записывает каждую AuditLog в файл одной JSON-строкой.
type FileObserver struct {
	file    *os.File
	encoder *json.Encoder
}

// NewFileObserver открывает (или создаёт) файл по path для дозаписи логов.
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

// HTTPObserver отправляет каждую AuditLog POST-запросом на url.
type HTTPObserver struct {
	url string
}

// NewHTTPObserver создаёт HTTPObserver, отправляющий логи на url.
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
