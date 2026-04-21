package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShortUrlJSON(t *testing.T) {
	shortenUrlStorage := repository.NewShortenURLStorage()
	shortenUrlService := service.NewShortenURLService(shortenUrlStorage)

	handler := CreateShortenUrlJSON(shortenUrlService)
	srv := httptest.NewServer(handler)

	defer srv.Close()

	config.Envs.BaseURL = srv.URL

	respJSON, errResp := json.Marshal(model.UrlResponse{Result: srv.URL + "/key_1"})
	require.NoError(t, errResp)

	reqJSON, errReq := json.Marshal(model.UrlRequest{Url: "https://yandex.ru"})
	require.NoError(t, errReq)

	type want struct {
		method      string
		code        int
		contentType string
		response    string
	}

	tests := []struct {
		name        string
		contentType string
		body        *strings.Reader
		want        want
	}{
		{
			name:        "case_1 Created",
			contentType: "application/json",
			body:        strings.NewReader(string(reqJSON)),
			want: want{
				method:      http.MethodPost,
				code:        http.StatusCreated,
				contentType: "application/json",
				response:    string(respJSON),
			},
		},
		{
			name:        "case_2 Method Not Allowed",
			contentType: "text/plain; charset=utf-8",
			body:        strings.NewReader(""),
			want: want{
				method:      http.MethodGet,
				code:        http.StatusMethodNotAllowed,
				contentType: "text/plain; charset=utf-8",
				response:    "Method not allowed",
			},
		},
		{
			name:        "case_3 Unsupported Content-Type",
			contentType: "text/plain; charset=utf-8",
			body:        strings.NewReader(""),
			want: want{
				method:      http.MethodPost,
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				response:    "Unsupported Content-Type",
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
				resp, err = http.Post(srv.URL, tt.contentType, tt.body)
			}
			require.NoError(t, err)

			defer func() {
				if err := resp.Body.Close(); err != nil {
					slog.Error("Failed to close request body", "error", err)
				}
			}()

			responseBody, _ := io.ReadAll(resp.Body)
			gotBody := strings.TrimSpace(string(responseBody))

			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-type"))
			assert.Equal(t, tt.want.response, gotBody)
			assert.Equal(t, tt.want.code, resp.StatusCode)
		})
	}
}
