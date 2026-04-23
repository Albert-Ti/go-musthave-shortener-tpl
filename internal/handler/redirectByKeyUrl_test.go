package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestRedirect(t *testing.T) {
	shortenUrlStorage, err := repository.NewShortenURLStorage()
	if err != nil {
		panic(err)
	}
	shortenUrlService := service.NewShortenURLService(shortenUrlStorage)
	shortenUrlService.Set("http://yandex.ru")

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
				location: shortenUrlService.Get("key_1"),
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
			name:     "case_3 Url not found",
			endpoint: "/unknown",
			want: want{
				method:   http.MethodGet,
				code:     http.StatusBadRequest,
				response: "Url not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest(tt.want.method, tt.endpoint, nil)
			w := httptest.NewRecorder()

			redirectHandler := RedirectByKeyUrl(shortenUrlService)
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
