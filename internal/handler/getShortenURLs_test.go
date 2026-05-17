package handler

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository/mocks"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetShortenURLs(t *testing.T) {
	tmpDir := t.TempDir()
	config.Envs.FileStoragePath = filepath.Join(tmpDir, "test.json")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		method      string
		contentType string
		statusCode  int
		response    string
		setupMock   func(repo *mocks.MockRepository)
	}{
		{
			name:        "case_1 OK",
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
					GetAll(gomock.Any()).
					Return(expectedURLs, nil).Times(1)
			},
		},

		{
			name:        "case_2 No Content",
			method:      http.MethodGet,
			contentType: "application/json",
			statusCode:  http.StatusNoContent,
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					GetAll(gomock.Any()).
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

			svc := service.NewService(mockRepo)

			req := httptest.NewRequest(
				tt.method,
				"/api/user/urls",
				nil,
			)

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
