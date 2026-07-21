package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/audit"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/handler"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository/mocks"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const savedKey = "abc123"

// ExampleRedirectByKeyURL демонстрирует редирект по короткому ключу
// (GET /{id}). Возвращает 307 и заголовок Location с исходным URL.
func ExampleRedirectByKeyURL() {
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

	// Заранее сохраняем ссылку, чтобы было куда редиректить.
	_, _ = repo.Save(context.Background(), savedKey, "http://yandex.ru", "user-1")

	svc := service.NewService(repo, config.NewOptions(config.WithBaseURL("http://localhost:8080")))
	auditor, _ := audit.NewAuditor("", "")

	req := httptest.NewRequestWithContext(
		context.WithValue(context.Background(),
			middleware.UserIDKey, "user-1"),
		http.MethodGet,
		"/abc123",
		nil,
	)
	rr := httptest.NewRecorder()

	handler.RedirectByKeyURL(svc, auditor, "http://localhost:8080")(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Header().Get("Location"))

	// Output:
	// 307
	// http://yandex.ru
}

func TestRedirectByKeyURL(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.NewOptions(
		config.WithFileStoragePath(filepath.Join(tmpDir, "test.json")),
		config.WithBaseURL("http://localhost:8080"),
	)

	type want struct {
		method   string
		code     int
		location string
	}
	tests := []struct {
		name      string
		endpoint  string
		want      want
		setupMock func(repo *mocks.MockRepository)
	}{
		{
			name:     "case_1 Redirected",
			endpoint: "/" + savedKey,
			want: want{
				method:   http.MethodGet,
				location: "http://yandex.ru",
				code:     http.StatusTemporaryRedirect,
			},
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().
					Get(gomock.Any(), savedKey).
					Return("http://yandex.ru", nil)
			},
		},
		{
			name:     "case_2 Method Not Allowed",
			endpoint: "/" + savedKey,
			want: want{
				method: http.MethodPost,
				code:   http.StatusMethodNotAllowed,
			},
		},
		{
			name:     "case_3 URL not found",
			endpoint: "/unknown",
			want: want{
				method: http.MethodGet,
				code:   http.StatusNotFound,
			},
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().
					Get(gomock.Any(), "unknown").
					Return("", repository.ErrNoRows)
			},
		},
		{
			name:     "case_4 Status Gone",
			endpoint: "/" + savedKey,
			want: want{
				method: http.MethodGet,
				code:   http.StatusGone,
			},
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().
					Get(gomock.Any(), savedKey).
					Return("", repository.ErrStatusGone)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}
			svc := service.NewService(mockRepo, cfg)

			r := httptest.NewRequestWithContext(
				context.WithValue(context.Background(),
					middleware.UserIDKey, "user-1"),
				tt.want.method,
				tt.endpoint,
				nil,
			)
			w := httptest.NewRecorder()
			auditor, _ := audit.NewAuditor("", "")

			redirectHandler := handler.RedirectByKeyURL(svc, auditor, cfg.BaseURL)
			redirectHandler(w, r)

			result := w.Result()
			defer func() {
				require.NoError(t, result.Body.Close())
			}()

			assert.Equal(t, tt.want.code, result.StatusCode)
			if tt.want.location != "" {
				assert.Equal(t, tt.want.location, result.Header.Get("Location"))
			}
		})
	}
}
