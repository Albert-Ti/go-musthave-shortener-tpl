package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/handler"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository/mocks"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStats(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.NewOptions(
		config.WithFileStoragePath(filepath.Join(tmpDir, "test.json")),
		config.WithTrustedSubnet("192.168.1.0/24"),
	)

	tests := []struct {
		name       string
		method     string
		xRealIP    string
		statusCode int
		response   string
		setupMock  func(repo *mocks.MockRepository)
	}{
		{
			name:       "Case_1 OK",
			method:     http.MethodGet,
			statusCode: http.StatusOK,
			xRealIP:    "192.168.1.1",
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					GetStats(gomock.Any()).
					Return(model.StatsResp{}, nil).
					Times(1)
			},
		},

		{
			name:       "Case_2 Forbidden",
			method:     http.MethodGet,
			statusCode: http.StatusForbidden,
			xRealIP:    "10.0.0.1",
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

			ctx := context.WithValue(
				context.Background(),
				middleware.UserIDKey,
				"123",
			)
			req := httptest.NewRequestWithContext(ctx, tt.method, "/api/internal/stats", nil)
			rr := httptest.NewRecorder()
			req.Header.Set("X-Real-IP", tt.xRealIP)

			handler := handler.GetStats(svc, cfg.TrustedSubnet)

			handler.ServeHTTP(rr, req)

			res := rr.Result()
			defer func() {
				require.NoError(t, res.Body.Close())
			}()

			assert.Equal(t, tt.statusCode, res.StatusCode)
		})
	}
}
