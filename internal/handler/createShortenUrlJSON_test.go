package handler

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShortURLJSON(t *testing.T) {
	tmpDir := t.TempDir()
	config.Envs.FileStoragePath = filepath.Join(tmpDir, "test.json")

	repo, e := repository.NewRepository()
	if e != nil {
		panic(e)
	}
	defer repo.Close()

	svc := service.NewService(repo)

	handler := CreateShortenURLJSON(svc)
	srv := httptest.NewServer(middleware.GzipCompress(
		middleware.AuthGuard(handler),
	))

	defer srv.Close()

	config.Envs.BaseURL = srv.URL

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
			name:        "case_2 Unsupported Content-Type",
			contentType: "text/plain",
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
			var (
				resp *http.Response
				req  *http.Request
				err  error
			)
			var bodyReader io.Reader = strings.NewReader(tt.body)

			if tt.want.method == http.MethodGet {
				req, err = http.NewRequest(http.MethodGet, srv.URL, nil)
				require.NoError(t, err)
			} else {
				req, err = http.NewRequest(tt.want.method, srv.URL, bodyReader)
				require.NoError(t, err)
			}

			req.Header.Set("Content-Type", tt.contentType)
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
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.response, gotBody)
		})
	}
}
