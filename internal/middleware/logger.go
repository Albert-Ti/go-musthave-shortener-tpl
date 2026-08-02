package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseData собирает данные об ответе для последующего логирования.
// generate:reset
type responseData struct {
	size   int
	status int
	body   string
}

// loggingResponseWriter оборачивает http.ResponseWriter, перехватывая
// статус и размер ответа.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (l *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	l.responseData.size += size
	l.responseData.body = string(b)
	return size, err
}

func (l *loggingResponseWriter) WriteHeader(statusCode int) {
	l.ResponseWriter.WriteHeader(statusCode)
	l.responseData.status = statusCode
}

// WithLogging - middleware, логирующее каждый запрос: метод, URI,
// длительность, статус и размер ответа. Уровень лога зависит от статуса:
// Error при 5xx, Warn при 4xx, Info в остальных случаях.
//
// Пример использования:
//
//	r := chi.NewRouter()
//	r.Use(middleware.WithLogging)
func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		switch {
		case responseData.status >= 500:
			slog.Error("",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.Duration("duration", duration),
				slog.Int("status", responseData.status),
				slog.Int("size", responseData.size),
			)
		case responseData.status >= 400:
			slog.Warn("",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.Duration("duration", duration),
				slog.Int("status", responseData.status),
				slog.Int("size", responseData.size),
			)
		default:
			slog.Info("",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.Duration("duration", duration),
				slog.Int("status", responseData.status),
				slog.Int("size", responseData.size),
			)
		}

	}
	return http.HandlerFunc(logFn)
}
