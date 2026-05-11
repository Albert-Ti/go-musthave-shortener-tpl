// internal/handler/createShortenURLBatch_test.go
package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository/mocks"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShortenURLBatch(t *testing.T) {
	config.Envs.BaseURL = "http://localhost:8080"

	defer utils.GenerateMockUUID()()

	tests := []struct {
		name        string
		method      string
		contentType string
		body        []model.JSONBatchReq
		setupMock   func(mock *mocks.MockRepository)
		statusCode  int
		response    []model.JSONBatchResp
	}{
		{
			name:        "Case_1 successful batch save",
			method:      http.MethodPost,
			contentType: "application/json",
			body: []model.JSONBatchReq{
				{CorrelationID: "ID1", OriginalURL: "https://example.com/1"},
				{CorrelationID: "ID2", OriginalURL: "https://example.com/2"},
			},
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					BatchSave(context.Background(), gomock.Any(), gomock.Any()).
					Return(repository.BatchConflict{}, nil).
					Times(1)
			},
			statusCode: http.StatusCreated,
			response: []model.JSONBatchResp{
				{ShortURL: "http://localhost:8080/key_1", CorrelationID: "ID1"},
				{ShortURL: "http://localhost:8080/key_2", CorrelationID: "ID2"},
			},
		},
		{
			name:        "Case_2 BatchSave error",
			method:      http.MethodPost,
			contentType: "application/json",
			body: []model.JSONBatchReq{
				{CorrelationID: "ID1", OriginalURL: "https://google.com"},
			},
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					BatchSave(context.Background(), gomock.Any(), gomock.Any()).
					Return(repository.BatchConflict{}, errors.New("database error")).
					Times(1)
			},
			statusCode: http.StatusInternalServerError,
			response:   nil,
		},
		{
			name:        "Case_3 BatchSave with conflict",
			method:      http.MethodPost,
			contentType: "application/json",
			body: []model.JSONBatchReq{
				{CorrelationID: "ID1", OriginalURL: "https://google.com"},
				{CorrelationID: "ID2", OriginalURL: "https://google.com"},
			},
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					BatchSave(context.Background(), gomock.Any(), gomock.Any()).
					Return(repository.BatchConflict{
						CorrelationID: "ID1",
						Key:           "key_1",
					}, repository.ErrConflict).
					Times(1)
			},
			statusCode: http.StatusConflict,
			response: []model.JSONBatchResp{
				{
					CorrelationID: "ID1",
					ShortURL:      "http://localhost:8080/key_1",
				},
			},
		},
		{
			name:        "Case_4 invalid HTTP method",
			method:      http.MethodGet,
			contentType: "application/json",
			body:        []model.JSONBatchReq{},
			setupMock:   func(mock *mocks.MockRepository) {},
			statusCode:  http.StatusMethodNotAllowed,
			response:    nil,
		},
		{
			name:        "Case_5 invalid content type",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        []model.JSONBatchReq{},
			setupMock:   func(mock *mocks.MockRepository) {},
			statusCode:  http.StatusBadRequest,
			response:    nil,
		},
		{
			name:        "Case_6 empty batch",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        []model.JSONBatchReq{},
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					BatchSave(context.Background(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			statusCode: http.StatusNoContent,
			response:   nil,
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
			handler := CreateShortenURLBatch(svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(tt.method, "/api/shorten/batch", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)

			if tt.response != nil && rr.Code == http.StatusCreated {
				var resp []model.JSONBatchResp
				err = json.NewDecoder(rr.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, tt.response, resp)
			}

			// Для конфликта проверяем ответ
			if tt.statusCode == http.StatusConflict && tt.response != nil {
				var resp []model.JSONBatchResp
				err = json.NewDecoder(rr.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, tt.response, resp)
			}
		})
	}
}
