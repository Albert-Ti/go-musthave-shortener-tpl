package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShortURL(t *testing.T) {
	tmpDir := t.TempDir()
	config.Envs.FileStoragePath = filepath.Join(tmpDir, "test.json")

	repo, e := repository.NewRepository()
	if e != nil {
		panic(e)
	}
	defer repo.Close()

	svc := service.NewService(repo)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, "123")
		r = r.WithContext(ctx)
		CreateShortenURL(svc)(w, r)
	})
	srv := httptest.NewServer(handler)

	defer srv.Close()

	config.Envs.BaseURL = srv.URL

	defer utils.GenerateMockUUID()()

	type want struct {
		method   string
		code     int
		response string
	}

	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "case_1 Created",
			body: "https://yandex.ru",
			want: want{
				method:   http.MethodPost,
				code:     http.StatusCreated,
				response: srv.URL + "/key_1",
			},
		},
		{
			name: "case_2 Method Not Allowed",
			body: "",
			want: want{
				method:   http.MethodGet,
				code:     http.StatusMethodNotAllowed,
				response: "Method not allowed",
			},
		},
		{
			name: "case_3 No body",
			body: "",
			want: want{
				method:   http.MethodPost,
				code:     http.StatusBadRequest,
				response: "Invalid URL",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var (
				resp *http.Response
				err  error
			)
			if tt.want.method == http.MethodGet {
				resp, err = http.Get(srv.URL)
			} else {
				resp, err = http.Post(srv.URL, "text/plain", strings.NewReader(tt.body))
			}
			require.NoError(t, err)

			defer func() {
				if err := resp.Body.Close(); err != nil {
					slog.Error("Failed to close request body", "error", err)
				}
			}()

			responseBody, _ := io.ReadAll(resp.Body)
			gotBody := strings.TrimSpace(string(responseBody))

			assert.Equal(t, tt.want.response, gotBody)
			assert.Equal(t, tt.want.code, resp.StatusCode)
		})
	}
}
