package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShortUrlJSON(t *testing.T) {
	config.Envs.FileStoragePath = "test.json"
	shortenUrlStorage, e := repository.NewShortenURLStorage()
	if e != nil {
		panic(e)
	}
	defer func() {
		shortenUrlStorage.Close()
		shortenUrlStorage.Remove()
	}()
	shortenUrlService := service.NewShortenURLService(shortenUrlStorage)

	handler := CreateShortenUrlJSON(shortenUrlService)
	srv := httptest.NewServer(middleware.GzipMiddleware(handler))

	defer srv.Close()

	config.Envs.BaseURL = srv.URL

	reqJSON, err := json.Marshal(model.ShortenUrlRequest{Url: "https://yandex.ru"})
	require.NoError(t, err)

	respJSON, err := json.Marshal(model.ShortenUrlResponse{Result: srv.URL + "/key_1"})
	require.NoError(t, err)

	respJSON2, err := json.Marshal(model.ShortenUrlResponse{Result: srv.URL + "/key_2"})
	require.NoError(t, err)

	respJSON3, err := json.Marshal(model.ShortenUrlResponse{Result: srv.URL + "/key_3"})
	require.NoError(t, err)

	type want struct {
		method          string
		code            int
		contentType     string
		acceptEncoding  string
		contentEncoding string
		response        string
	}

	tests := []struct {
		name            string
		contentType     string
		acceptEncoding  string
		contentEncoding string
		body            string
		want            want
	}{
		{
			name:        "case_1 Created",
			contentType: "application/json",
			body:        string(reqJSON),
			want: want{
				method:      http.MethodPost,
				code:        http.StatusCreated,
				contentType: "application/json",
				response:    string(respJSON),
			},
		},
		{
			name: "case_2 Method Not Allowed",
			want: want{
				method:      http.MethodGet,
				code:        http.StatusMethodNotAllowed,
				contentType: "text/plain; charset=utf-8",
				response:    "Method not allowed",
			},
		},
		{
			name:            "case_4 Send gzip",
			contentType:     "application/json",
			contentEncoding: "gzip",
			acceptEncoding:  "",
			body:            string(reqJSON),
			want: want{
				method:      http.MethodPost,
				code:        http.StatusCreated,
				contentType: "application/json",
				response:    string(respJSON2),
			},
		},
		{
			name:            "case_5 Accept gzip",
			contentType:     "application/json",
			contentEncoding: "",
			acceptEncoding:  "gzip",
			body:            string(reqJSON),
			want: want{
				method:      http.MethodPost,
				code:        http.StatusCreated,
				contentType: "application/json",
				response:    string(respJSON3),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				resp *http.Response
				req  *http.Request
				err  error
			)

			if tt.want.method == http.MethodGet {
				req, err = http.NewRequest(http.MethodGet, srv.URL, nil)
				require.NoError(t, err)
			} else {
				var bodyReader io.Reader = strings.NewReader(tt.body)

				if tt.contentEncoding == "gzip" {
					buf := bytes.NewBuffer(nil)
					gzWriter := gzip.NewWriter(buf)
					_, err := gzWriter.Write([]byte(tt.body))
					require.NoError(t, err)
					err = gzWriter.Close()
					require.NoError(t, err)
					bodyReader = buf
				}

				req, err = http.NewRequest(tt.want.method, srv.URL, bodyReader)
				require.NoError(t, err)

				req.Header.Set("Content-Type", tt.contentType)
				if tt.contentEncoding != "" {
					req.Header.Set("Content-Encoding", tt.contentEncoding)
				}
				if tt.acceptEncoding != "" {
					req.Header.Set("Accept-Encoding", tt.acceptEncoding)
				}
			}

			client := &http.Client{}
			resp, err = client.Do(req)
			require.NoError(t, err)

			defer func() {
				if err := resp.Body.Close(); err != nil {
					slog.Error("Failed to close response body", "error", err)
				}
			}()

			var reader io.Reader = resp.Body
			if resp.Header.Get("Content-Encoding") == "gzip" {
				gzReader, err := gzip.NewReader(resp.Body)
				require.NoError(t, err)
				defer gzReader.Close()
				reader = gzReader
			}

			responseBody, err := io.ReadAll(reader)
			require.NoError(t, err)
			gotBody := strings.TrimSpace(string(responseBody))

			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.response, gotBody)
			assert.Equal(t, tt.want.code, resp.StatusCode)
		})
	}
}
