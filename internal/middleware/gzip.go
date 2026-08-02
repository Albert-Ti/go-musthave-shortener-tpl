package middleware

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/pool"
)

// compressWriter оборачивает http.ResponseWriter и сжимает ответ в gzip.
type compressWriter struct {
	w    http.ResponseWriter
	zw   *gzip.Writer
	pool *pool.Pool[*compressWriter]
}

var writerPool = pool.New(func() *compressWriter {
	zw, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
	return &compressWriter{zw: zw}
})

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	c := writerPool.Get()
	c.zw.Reset(w)
	c.w = w
	c.pool = writerPool
	return c
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if c.Header().Get("Content-Encoding") != "gzip" {
		return c.w.Write(p)
	}

	return c.zw.Write(p)
}

func (c *compressWriter) Reset() {
	c.zw.Reset(io.Discard) // отвязка от старого w
	c.w = nil
}

func (c *compressWriter) Close() error {
	if c.Header().Get("Content-Encoding") == "gzip" {
		err := c.zw.Close()
		c.pool.Put(c)
		return err
	}
	return nil
}

// compressReader оборачивает io.ReadCloser и распаковывает gzip при чтении.
type compressReader struct {
	zr   *gzip.Reader
	pool *pool.Pool[*compressReader]
}

var readerPool = pool.New(func() *compressReader {
	return &compressReader{} // zr == nil, инициализируется лениво при первом реальном r
})

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	c := readerPool.Get()

	if c.zr == nil {
		zr, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		c.zr = zr
	} else {
		if err := c.zr.Reset(r); err != nil {
			return nil, err
		}
	}

	c.pool = readerPool
	return c, nil
}

func (c *compressReader) Read(b []byte) (int, error) {
	return c.zr.Read(b)
}

func (c *compressReader) Reset() {
	_ = c.zr.Reset(http.NoBody) // безопасная отвязка от старого r
}

func (c *compressReader) Close() error {
	if err := c.zr.Close(); err != nil {
		return err
	}
	c.pool.Put(c)
	return nil
}

// GzipCompress - middleware, сжимающее тело запроса и ответа в gzip,
// если клиент это поддерживает.
//
// Пример использования:
//
//	r := chi.NewRouter()
//	r.Use(middleware.GzipCompress)
func GzipCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer func() {
				if err := r.Body.Close(); err != nil {
					slog.Error("Failed to close request body", "error", err)
				}
			}()

			r.Body = cr
			defer func() {
				if err := cr.Close(); err != nil {
					slog.Error("gzip: close request body reader:", "error", err)
				}
			}()
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := newCompressWriter(w)
			w = cw
			defer func() {
				if err := cw.Close(); err != nil {
					slog.Error("gzip: close response writer:", "error", err)
				}
			}()
		}

		next.ServeHTTP(w, r)
	})
}
