// При Написание аудита использовался pattern Наблюдатель(Observer).
package audit

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Observer получает уведомление о каждой новой записи аудита.
type Observer interface {
	Notify(log AuditLog)
}

// AuditLog запись аудита одного запроса.
// generate:reset
type AuditLog struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}

// Auditor рассылает записи аудита подписанным наблюдателям (observer)
// через буферизованный канал, не блокируя обработку HTTP-запроса.
// generate:reset
type Auditor struct {
	ch       chan AuditLog
	action   map[string]string
	observer []Observer

	bufSize int
	workers int
	wg      sync.WaitGroup
	active  atomic.Int32
}

// NewAuditor создаёт Auditor и подписывает на него FileObserver (если задан
// auditFile) и HTTPObserver (если задан auditURL). Если оба параметра пустые,
// возвращает (nil, nil) - аудит считается выключенным.
//
// Пример использования:
//
//	auditor, err := audit.NewAuditor("audit.json", "http...")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	auditor.AddLog(http.MethodPost, r.RequestURI, "user-1", "http://localhost:8080")
func NewAuditor(auditFile string, auditURL string, bufSize int, workers int) (*Auditor, error) {
	if auditFile == "" && auditURL == "" {
		return nil, nil
	}

	auditor := &Auditor{
		ch:      make(chan AuditLog, bufSize),
		action:  map[string]string{http.MethodGet: "follow", http.MethodPost: "shorten"},
		workers: workers,
	}

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
		TS:     time.Now().Unix(),
		Action: a.action[method],
		UserID: userID,
		URL:    baseURL + requestURI,
	}

	select {
	case a.ch <- log:
	default:
	}
}

// subscribe - добавляет новых наблюдателей.
func (a *Auditor) subscribe(sub Observer) {
	a.observer = append(a.observer, sub)
}

// broadcast - рассылает логи существующим наблюдателям.
func (a *Auditor) broadcast() {
	for range a.workers {
		a.active.Add(1)

		a.wg.Go(func() {
			for log := range a.ch {
				for _, sub := range a.observer {
					sub.Notify(log)
				}
			}
		})
	}
}

// FileObserver записывает каждую AuditLog в файл одной JSON-строкой.
type FileObserver struct {
	file    *os.File
	encoder *json.Encoder

	mu sync.Mutex
}

// NewFileObserver открывает (или создаёт) файл по path для дозаписи логов.
func NewFileObserver(path string) (*FileObserver, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &FileObserver{file, json.NewEncoder(file), sync.Mutex{}}, nil
}

// Notify - функция FileObserver кодировки и записи в файл логов.
func (f *FileObserver) Notify(log AuditLog) {
	f.mu.Lock()
	defer f.mu.Unlock()

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

// Notify - функция HTTPObserver отправки пост логов запроса на сторонний сервер.
func (h *HTTPObserver) Notify(log AuditLog) {
	b, err := json.Marshal(log)
	if err != nil {
		slog.Error("HTTPObserver", "marshal", err)
		return
	}

	resp, err := http.Post(h.url, "application/json", bytes.NewReader(b))
	if err != nil {
		slog.Error("HTTPObserver", "post", err)
		return
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			slog.Error("HTTPObserver", "close", errClose)
		}
	}()
}

func (a *Auditor) Close() {
	if a == nil {
		return
	}

	close(a.ch)
	a.wg.Wait()
}
