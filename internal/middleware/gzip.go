package middleware

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/pool"
)

var (
	writerPool = pool.NewPool[*compressWriter]()
	readerPool = pool.NewPool[*compressReader]()
)

// compressWriter оборачивает http.ResponseWriter и сжимает ответ в gzip.
type compressWriter struct {
	w    http.ResponseWriter
	zw   *gzip.Writer
	pool *pool.Pool[*compressWriter]
}

func newCompressWriter(w http.ResponseWriter) (*compressWriter, error) {
	c, ok := writerPool.Get()
	if !ok {
		zw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			return nil, err
		}
		c = &compressWriter{zw: zw}
	} else {
		// обнуляет внутренние буферы и флаг closed, забывает старый w, запоминает новый w
		c.zw.Reset(w)
	}
	c.w = w
	c.pool = writerPool
	return c, nil
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
	r    io.ReadCloser
	zr   *gzip.Reader
	pool *pool.Pool[*compressReader]
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	rp, ok := readerPool.Get()

	if !ok {

		zr, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		rp = &compressReader{zr: zr}
	} else {
		if err := rp.zr.Reset(r); err != nil {
			return nil, err
		}
	}

	rp.r = r
	rp.pool = readerPool
	return rp, nil
}

func (c compressReader) Read(b []byte) (int, error) {
	return c.zr.Read(b)
}

func (c *compressReader) Reset() {
	_ = c.zr.Reset(http.NoBody) // безопасная отвязка от старого r
	c.r = nil
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
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
				slog.Error("newCompressReader", "error", err)
				return
			}

			r.Body = cr
			defer func() {
				if err := cr.Close(); err != nil {
					slog.Error("gzip: close request body reader:", "error", err)
				}
			}()
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw, err := newCompressWriter(w)
			if err != nil {
				slog.Error("newCompressWriter", "error", err)
				return
			}

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
