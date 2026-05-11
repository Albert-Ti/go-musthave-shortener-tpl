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
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestRedirectByKeyURL(t *testing.T) {
	tmpDir := t.TempDir()
	config.Envs.FileStoragePath = filepath.Join(tmpDir, "test.json")

	repo, e := repository.NewRepository()
	if e != nil {
		panic(e)
	}
	defer repo.Close()

	defer utils.GenerateMockUUID()()

	svc := service.NewService(repo)

	ctx := context.Background()
	svc.Save(ctx, "http://yandex.ru")
	getURL, _ := svc.Get(ctx, "key_1")

	type want struct {
		method   string
		code     int
		location string
		response string
	}
	tests := []struct {
		name     string
		endpoint string
		want     want
	}{
		{
			name:     "case_1 Redirected",
			endpoint: "/key_1",
			want: want{
				method:   http.MethodGet,
				location: getURL,
				code:     http.StatusTemporaryRedirect,
			},
		},
		{
			name:     "case_2 Method Not Allowed",
			endpoint: "/key_1",
			want: want{
				method:   http.MethodPost,
				code:     http.StatusMethodNotAllowed,
				response: "Method not allowed",
			},
		},
		{
			name:     "case_3 URL not found",
			endpoint: "/unknown",
			want: want{
				method:   http.MethodGet,
				code:     http.StatusBadRequest,
				response: "URL not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest(tt.want.method, tt.endpoint, nil)
			w := httptest.NewRecorder()

			redirectHandler := RedirectByKeyURL(svc)
			redirectHandler(w, r)

			result := w.Result()

			defer func() {
				if err := result.Body.Close(); err != nil {
					slog.Error("Failed to close request body", "error", err)
				}
			}()

			responseBody, _ := io.ReadAll(result.Body)
			gotBody := strings.TrimSpace(string(responseBody))

			if tt.want.response != "" {
				assert.Equal(t, tt.want.response, gotBody)
			}
			assert.Equal(t, tt.want.code, result.StatusCode)

			if tt.want.location != "" {
				assert.Equal(t, tt.want.location, result.Header.Get("Location"))
			}
		})
	}
}
