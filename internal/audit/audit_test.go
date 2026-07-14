package audit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockObserver struct{}

func (m *mockObserver) Notify(_ AuditLog) {}

type slowObserver struct{ delay time.Duration }

func (s *slowObserver) Notify(_ AuditLog) {
	time.Sleep(s.delay)
}

func BenchmarkBufferSizes(b *testing.B) {
	bufSizes := []int{0, 1, 10, 100}

	for _, size := range bufSizes {
		b.Run(fmt.Sprintf("buf_%d", size), func(b *testing.B) {
			auditor := &Auditor{
				ch:     make(chan AuditLog, size),
				action: map[string]string{http.MethodGet: "follow"},
			}
			auditor.subscribe(&mockObserver{})
			go auditor.broadcast()
			defer close(auditor.ch)

			b.ResetTimer()

			for b.Loop() {
				auditor.AddLog(http.MethodGet, "/user/audit", "user-1", "http://localhost:8080")
			}
		})
	}
}

func BenchmarkSlowSubscriber(b *testing.B) {
	delays := []time.Duration{0, time.Microsecond, 10 * time.Microsecond}

	for _, delay := range delays {
		b.Run(fmt.Sprintf("delay_%v", delay), func(b *testing.B) {
			auditor := &Auditor{
				ch:     make(chan AuditLog, 20),
				action: map[string]string{http.MethodGet: "follow"},
			}
			auditor.subscribe(&slowObserver{delay: delay})
			go auditor.broadcast()
			defer close(auditor.ch)

			b.ResetTimer()

			for b.Loop() {
				auditor.AddLog(http.MethodGet, "/user/audit", "user-1", "http://localhost:8080")
			}
		})
	}
}

func BenchmarkParallel(b *testing.B) {
	auditor := &Auditor{
		ch:     make(chan AuditLog, 20),
		action: map[string]string{http.MethodGet: "follow"},
	}

	auditor.subscribe(&mockObserver{})
	go auditor.broadcast()
	defer close(auditor.ch)
	// b.RunParallel запускает несколько горутин одновременно (по умолчанию GOMAXPROCS),
	// имитируя реальную многопоточную нагрузку HTTP-сервера
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			auditor.AddLog(http.MethodGet, "/user/audit", "user-1", "http://localhost:8080")
		}
	})
}

func BenchmarkSlowHTTPObserver(b *testing.B) {
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond) // симулируем медленный внешний сервис
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	auditor := &Auditor{
		ch:     make(chan AuditLog, 20),
		action: map[string]string{http.MethodGet: "follow"},
	}
	auditor.subscribe(NewHTTPObserver(slowServer.URL))
	go auditor.broadcast()
	defer close(auditor.ch)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		auditor.AddLog(http.MethodGet, "/user/audit", "user-1", "http://localhost:8080")
	}
}
