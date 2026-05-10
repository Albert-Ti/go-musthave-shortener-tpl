package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository/mocks"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShortenURLBatch(t *testing.T) {
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
				{CorrelationID: "ID1", OriginalURL: "https://google.com"},
				{CorrelationID: "ID2", OriginalURL: "https://yandex.ru"},
			},
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					Length().
					Return(0, nil).
					Times(1)

				mock.EXPECT().
					BatchSave(gomock.Any(), gomock.Any()).
					Return("", "", nil).
					Times(1)
			},
			statusCode: http.StatusCreated,
			response: []model.JSONBatchResp{
				{ShortURL: config.Envs.BaseURL + "/" + "key_1", CorrelationID: "ID1"},
				{ShortURL: config.Envs.BaseURL + "/" + "key_2", CorrelationID: "ID2"},
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
					Length().
					Return(0, nil).
					Times(1)

				mock.EXPECT().
					BatchSave(gomock.Any(), gomock.Any()).
					Return("", "", errors.New("database error")).
					Times(1)
			},
			statusCode: http.StatusInternalServerError,
			response:   nil,
		},
		{
			name:        "Case_3 invalid HTTP method",
			method:      http.MethodGet,
			contentType: "application/json",
			body:        []model.JSONBatchReq{},
			setupMock:   func(mock *mocks.MockRepository) {},
			statusCode:  http.StatusMethodNotAllowed,
			response:    nil,
		},
		{
			name:        "Case_4 invalid content type",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        []model.JSONBatchReq{},
			setupMock:   func(mock *mocks.MockRepository) {},
			statusCode:  http.StatusBadRequest,
			response:    nil,
		},
		{
			name:        "Case_5 Data is empty",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        []model.JSONBatchReq{},
			setupMock:   func(mock *mocks.MockRepository) {},
			statusCode:  http.StatusNoContent,
			response:    nil,
		},
		{
			name:        "Case_6 Length error",
			method:      http.MethodPost,
			contentType: "application/json",
			body: []model.JSONBatchReq{
				{CorrelationID: "ID1", OriginalURL: "https://google.com"},
			},
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					Length().
					Return(0, errors.New("DB error"))

				mock.EXPECT().
					BatchSave(gomock.Any(), gomock.Any()).
					Times(0)
			},
			statusCode: http.StatusInternalServerError,
			response:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			tt.setupMock(mockRepo)

			svc := service.NewService(mockRepo)
			handler := CreateShortenURLBatch(svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(tt.method, "/api/shorten/batch", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)

			if tt.response != nil {
				var resp []model.JSONBatchResp
				err = json.NewDecoder(rr.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, tt.response, resp)
			}

			if tt.statusCode == http.StatusConflict {
				assert.Empty(t, rr.Body.String())
			}
		})
	}
}
