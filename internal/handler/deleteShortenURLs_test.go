package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository/mocks"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDeleteShortenURLs(t *testing.T) {
	cfg := config.NewOptions()

	tests := []struct {
		name       string
		method     string
		body       any
		statusCode int
		setupMock  func(repo *mocks.MockRepository)
	}{
		{
			name:       "Case_1 Accepted",
			method:     http.MethodDelete,
			body:       []string{"key_1", "key_2"},
			statusCode: http.StatusAccepted,
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					BatchDelete(
						gomock.Any(),
						[]string{"key_1", "key_2"},
						"123",
					).
					Return(nil).
					AnyTimes()
			},
		},
		{
			name:       "Case_2 Method Not Allowed",
			method:     http.MethodGet,
			body:       nil,
			statusCode: http.StatusMethodNotAllowed,
			setupMock:  func(mock *mocks.MockRepository) {},
		},
		{
			name:       "Case_3 Invalid JSON",
			method:     http.MethodDelete,
			body:       "invalid-json",
			statusCode: http.StatusInternalServerError,
			setupMock:  func(mock *mocks.MockRepository) {},
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

			var bodyBytes []byte

			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			ctx := context.WithValue(
				context.Background(),
				middleware.UserIDKey,
				"123",
			)

			req := httptest.NewRequestWithContext(
				ctx,
				tt.method,
				"/api/user/urls",
				bytes.NewReader(bodyBytes),
			)

			rr := httptest.NewRecorder()

			handler := DeleteShortenURLs(svc)

			handler.ServeHTTP(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.statusCode, res.StatusCode)
		})
	}
}
