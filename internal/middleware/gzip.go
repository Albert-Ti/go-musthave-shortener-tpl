package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

type gzipPool struct {
	mu   sync.Mutex
	free []*gzip.Writer
}

func newGzipPool() *gzipPool {
	return &gzipPool{}
}

func (p *gzipPool) get(w http.ResponseWriter) *gzip.Writer {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.free) == 0 {
		// gzip.BestSpeed (=1) — сжимает быстрее, но хуже (больше итоговый размер);
		// gzip.BestCompression (=9) — сжимает медленнее, но сильнее (меньше итоговый размер);
		// gzip.DefaultCompression (=-1) — компромисс, то же самое что использует NewWriter.
		newZw, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
		return newZw
	}

	gz := p.free[len(p.free)-1]     // берем
	p.free = p.free[:len(p.free)-1] // удаляет
	gz.Reset(w)                     // очистка внутреннего состояния самого объекта gz
	return gz
}

func (p *gzipPool) put(gz *gzip.Writer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.free = append(p.free, gz)
}

type compressWriter struct {
	w    http.ResponseWriter
	zw   *gzip.Writer
	pool *gzipPool
}

var gzPool = newGzipPool()

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	zw := gzPool.get(w)

	return &compressWriter{
		w:    w,
		zw:   zw,
		pool: gzPool,
	}
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

func (c *compressWriter) Close() error {
	if c.Header().Get("Content-Encoding") == "gzip" {
		err := c.zw.Close()
		c.pool.put(c.zw)
		return err
	}
	return nil
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(b []byte) (int, error) {
	return c.zr.Read(b)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer cr.Close()
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := newCompressWriter(w)

			w = cw
			defer cw.Close()
		}

		next.ServeHTTP(w, r)
	})
}
