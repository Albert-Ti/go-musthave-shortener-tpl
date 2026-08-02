package audit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	auditor, _ := NewAuditor("", slowServer.URL, 50, 100)
	defer auditor.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		auditor.AddLog(http.MethodGet, "/user/audit", "user-1", "http://localhost:8080")
		time.Sleep(time.Millisecond)
	}
}

func TestAuditor_WorkerCount(t *testing.T) {
	expectWorkers := 100

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond) // симулируем медленный внешний сервис
		w.WriteHeader(http.StatusOK)
	}))

	defer slowServer.Close()

	auditor, _ := NewAuditor("", slowServer.URL, 20, expectWorkers)
	defer auditor.Close()

	for i := 0; i < 100; i++ {
		auditor.AddLog(http.MethodGet, "/user/audit", "user-1", "http://localhost:8080")
		time.Sleep(time.Millisecond)
	}

	assert.Equal(t, int32(expectWorkers), auditor.active.Load())
}

type countingObserver struct {
	count atomic.Int64
}

func (c *countingObserver) Notify(_ AuditLog) {
	c.count.Add(1)
	time.Sleep(100 * time.Millisecond)
}

func TestAuditor_GoroutineCount(t *testing.T) {
	before := runtime.NumGoroutine()

	auditor := &Auditor{
		ch:     make(chan AuditLog, 20),
		action: map[string]string{http.MethodGet: "follow"},
	}

	// Добавляем 3 наблюдателя
	obs1 := &countingObserver{}
	obs2 := &countingObserver{}
	auditor.observer = []Observer{obs1, obs2}

	go auditor.broadcast()

	for range 100000 {
		auditor.AddLog(http.MethodGet, "/test", "user-1", "http://localhost")
	}

	close(auditor.ch)

	after := runtime.NumGoroutine()

	if after > before+5 {
		t.Errorf("Too many goroutines: before=%d, after=%d, diff=%d",
			before, after, after-before)
	}

	t.Logf("Goroutines: before=%d, after=%d", before, after)
}
