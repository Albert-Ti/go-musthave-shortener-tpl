package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository/mocks"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// ExampleGetShortenURLs демонстрирует получение всех ссылок пользователя
// (GET /api/user/urls). Возвращает JSON-массив пар key/url.
func ExampleGetShortenURLs() {
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

	_, _ = repo.Save(context.Background(), "key_1", "https://yandex.ru", "123")

	svc := service.NewService(repo, config.NewOptions(config.WithBaseURL("http://localhost:8080")))

	req := httptest.NewRequestWithContext(
		context.WithValue(context.Background(),
			middleware.UserIDKey, "123"),
		http.MethodGet,
		"/api/user/urls",
		nil,
	)
	rr := httptest.NewRecorder()

	GetShortenURLs(svc)(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(strings.TrimSpace(rr.Body.String()))

	// Output:
	// 200
	// [{"short_url":"http://localhost:8080/key_1","original_url":"https://yandex.ru"}]
}

func TestGetShortenURLs(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.NewOptions(
		config.WithFileStoragePath(filepath.Join(tmpDir, "test.json")),
	)

	tests := []struct {
		name        string
		method      string
		contentType string
		statusCode  int
		response    string
		setupMock   func(repo *mocks.MockRepository)
	}{
		{
			name:        "Case_1 OK",
			method:      http.MethodGet,
			contentType: "application/json",
			statusCode:  http.StatusOK,
			setupMock: func(mock *mocks.MockRepository) {
				expectedURLs := []map[string]string{
					{
						"key": "http://localhost:8080/key_1",
						"url": "https://google.com",
					},
				}

				mock.EXPECT().
					GetAll(gomock.Any(), "123").
					Return(expectedURLs, nil).Times(1)
			},
		},

		{
			name:        "Case_2 No Content",
			method:      http.MethodGet,
			contentType: "application/json",
			statusCode:  http.StatusNoContent,
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					GetAll(gomock.Any(), "123").
					Return(nil, nil).Times(1)
			},
		},
		{
			name:        "Case_3 invalid HTTP method",
			method:      http.MethodPost,
			contentType: "application/json",
			setupMock:   func(mock *mocks.MockRepository) {},
			statusCode:  http.StatusMethodNotAllowed,
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

			ctx := context.WithValue(
				context.Background(),
				middleware.UserIDKey,
				"123",
			)
			req := httptest.NewRequestWithContext(ctx, tt.method, "/api/user/urls", nil)

			rr := httptest.NewRecorder()

			handler := GetShortenURLs(svc)

			handler.ServeHTTP(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.statusCode, res.StatusCode)
			if tt.name == "Case_1 OK" {
				assert.Equal(t, tt.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
