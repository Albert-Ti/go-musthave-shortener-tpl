package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
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
	svc := service.NewService(repo)

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "123")

	savedKey, _, err := svc.Save(ctx, "http://yandex.ru", "123")
	if err != nil {
		t.Fatalf("Failed to save URL: %v", err)
	}

	getURL, _ := svc.Get(ctx, savedKey)

	fmt.Println(getURL)

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
			endpoint: "/" + savedKey,
			want: want{
				method:   http.MethodGet,
				location: getURL,
				code:     http.StatusTemporaryRedirect,
			},
		},
		{
			name:     "case_2 Method Not Allowed",
			endpoint: "/" + savedKey,
			want: want{
				method:   http.MethodPost,
				code:     http.StatusMethodNotAllowed,
				response: "Method not allowed\n",
			},
		},
		{
			name:     "case_3 URL not found",
			endpoint: "/unknown",
			want: want{
				method:   http.MethodGet,
				code:     http.StatusNotFound,
				response: "URL not found\n",
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
			defer result.Body.Close()

			responseBody, _ := io.ReadAll(result.Body)
			gotBody := strings.TrimSpace(string(responseBody))

			if tt.want.response != "" {
				assert.Equal(t, strings.TrimSpace(tt.want.response), gotBody)
			}
			assert.Equal(t, tt.want.code, result.StatusCode)

			if tt.want.location != "" {
				assert.Equal(t, tt.want.location, result.Header.Get("Location"))
			}
		})
	}
}
