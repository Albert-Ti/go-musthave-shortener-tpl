package middleware_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipCompress(t *testing.T) {
	requestBody := "Hello World!"
	successBody := "Zag! Zag!"

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(successBody))

		require.NoError(t, err)
	})

	gzipHandler := middleware.GzipCompress(testHandler)

	t.Run("Zip Zip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/test", buf)

		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")

		rr := httptest.NewRecorder()
		gzipHandler.ServeHTTP(rr, req)

		zr, err := gzip.NewReader(rr.Body)
		require.NoError(t, err)

		body, err := io.ReadAll(zr)

		assert.Equal(t, successBody, string(body))
	})

	t.Run("Send Zip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/test", buf)

		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "")

		rr := httptest.NewRecorder()
		gzipHandler.ServeHTTP(rr, req)

		body, err := io.ReadAll(rr.Body)
		require.NoError(t, err)

		assert.Equal(t, successBody, string(body))
	})

	t.Run("Accept Zip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", strings.NewReader(requestBody))

		req.Header.Set("Content-Encoding", "")
		req.Header.Set("Accept-Encoding", "gzip")

		rr := httptest.NewRecorder()
		gzipHandler.ServeHTTP(rr, req)

		zr, err := gzip.NewReader(rr.Body)
		require.NoError(t, err)

		body, err := io.ReadAll(zr)
		require.NoError(t, err)

		assert.Equal(t, successBody, string(body))
	})

	t.Run("No Zip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", strings.NewReader(requestBody))

		req.Header.Set("Content-Encoding", "")
		req.Header.Set("Accept-Encoding", "")

		rr := httptest.NewRecorder()
		gzipHandler.ServeHTTP(rr, req)

		body, err := io.ReadAll(rr.Body)
		require.NoError(t, err)

		assert.Equal(t, successBody, string(body))
	})
}

func BenchmarkGzipCompress(b *testing.B) {
	requestBody := "Hello World!"
	successBody := "Zag! Zag!"

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(successBody))
		require.NoError(b, err)
	})

	gzipHandler := middleware.GzipCompress(testHandler)

	b.Run("gzip alloc", func(b *testing.B) {
		for b.Loop() {
			buf := bytes.NewBuffer(nil)
			zb := gzip.NewWriter(buf)
			_, err := zb.Write([]byte(requestBody))
			require.NoError(b, err)
			err = zb.Close()
			require.NoError(b, err)

			req := httptest.NewRequest("GET", "/test", buf)

			req.Header.Set("Content-Encoding", "gzip")
			req.Header.Set("Accept-Encoding", "gzip")

			rr := httptest.NewRecorder()
			gzipHandler.ServeHTTP(rr, req)

			zr, err := gzip.NewReader(rr.Body)
			require.NoError(b, err)

			body, err := io.ReadAll(zr)

			assert.Equal(b, successBody, string(body))
		}
	})
}
