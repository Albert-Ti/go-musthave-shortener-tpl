package handler_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/audit"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/handler"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExampleCreateShortenURLJSON демонстрирует создание короткой ссылки через JSON API
// (POST /api/shorten). Возвращает объект model.JSONResp.
func ExampleCreateShortenURLJSON() {
	dir, err := os.MkdirTemp("", "example")
	if err != nil {
		panic(err)
	}

	cfg := config.NewOptions(
		config.WithFileStoragePath(filepath.Join(dir, "example.json")),
		config.WithBaseURL("http://localhost:8080"),
	)

	repo, err := repository.NewRepository(cfg)
	if err != nil {
		panic(err)
	}

	svc := service.NewService(repo, config.NewOptions(config.WithBaseURL("http://localhost:8080")))
	auditor, _ := audit.NewAuditor("", "", 20, 100)

	defer utils.GenerateMockUUID()()

	reqBody, _ := json.Marshal(model.JSONReq{URL: "https://yandex.ru"})

	req := httptest.NewRequestWithContext(
		middleware.SetAuthUserID(context.Background(), "user-1"),
		http.MethodPost,
		"/api/shorten",
		bytes.NewReader(reqBody),
	)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.CreateShortenURLJSON(svc, auditor, "http://localhost:8080")(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(strings.TrimSpace(rr.Body.String()))

	// Output:
	// 201
	// {"result":"http://localhost:8080/key_1"}
}

func TestCreateShortURLJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.NewOptions(
		config.WithFileStoragePath(filepath.Join(tmpDir, "test.json")),
	)
	repo, e := repository.NewRepository(cfg)
	if e != nil {
		panic(e)
	}
	defer func() {
		require.NoError(t, repo.Close())
	}()

	svc := service.NewService(repo, cfg)
	auditor, _ := audit.NewAuditor("", "", 20, 100)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.SetAuthUserID(context.Background(), "user-1")
		r = r.WithContext(ctx)
		handler.CreateShortenURLJSON(svc, auditor, cfg.BaseURL)(w, r)
	})

	// e2e test
	srv := httptest.NewServer(middleware.GzipCompress(handler))
	defer srv.Close()

	cfg.BaseURL = srv.URL

	defer utils.GenerateMockUUID()()

	reqJSON, err := json.Marshal(model.JSONReq{URL: "https://yandex.ru"})
	require.NoError(t, err)

	respJSON, err := json.Marshal(model.JSONResp{Result: srv.URL + "/key_1"})
	require.NoError(t, err)
	respJSON2, err := json.Marshal(model.JSONResp{Result: srv.URL + "/key_2"})
	require.NoError(t, err)
	respJSON3, err := json.Marshal(model.JSONResp{Result: srv.URL + "/key_3"})
	require.NoError(t, err)

	type want struct {
		method      string
		code        int
		contentType string
		response    string
	}

	tests := []struct {
		name        string
		contentType string
		body        string
		want        want
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
			name:        "case_3 Unsupported Content-Type",
			contentType: "text/plain",
			body:        "test",
			want: want{
				method:      http.MethodPost,
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				response:    "Unsupported Content-Type",
			},
		},
		{
			name:        "case_4 Send gzip",
			contentType: "application/json",
			body:        string(reqJSON),
			want: want{
				method:      http.MethodPost,
				code:        http.StatusCreated,
				contentType: "application/json",
				response:    string(respJSON2),
			},
		},
		{
			name:        "case_5 Accept gzip",
			contentType: "application/json",
			body:        string(reqJSON),
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
			var bodyReader io.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			}

			req, err := http.NewRequest(tt.want.method, srv.URL, bodyReader)
			require.NoError(t, err)

			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)

			defer func() {
				if errBody := resp.Body.Close(); errBody != nil {
					slog.Error("Failed to close response body", "error", errBody)
				}
			}()

			var reader io.Reader = resp.Body
			if resp.Header.Get("Content-Encoding") == "gzip" {
				gzReader, errGzip := gzip.NewReader(resp.Body)
				require.NoError(t, errGzip)
				defer func() {
					require.NoError(t, gzReader.Close())
				}()
				reader = gzReader
			}

			responseBody, err := io.ReadAll(reader)
			require.NoError(t, err)
			gotBody := strings.TrimSpace(string(responseBody))

			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.response, gotBody)
		})
	}
}
