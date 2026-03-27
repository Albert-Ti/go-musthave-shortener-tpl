package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestRedirect(t *testing.T) {
	urls := &model.ShortenedUrls{List: map[string]string{"key_1": "http://yandex.ru"}}

	type want struct {
		method   string
		code     int
		location string
		response string
	}
	tests := []struct {
		name   string
		preset string
		want   want
	}{
		{
			name:   "case_1 Redirected",
			preset: "/key_1",
			want: want{
				method:   http.MethodGet,
				location: urls.List["key_1"],
				code:     http.StatusTemporaryRedirect,
			},
		},
		{
			name:   "case_2 Method Not Allowed",
			preset: "/key_1",
			want: want{
				method:   http.MethodPost,
				code:     http.StatusMethodNotAllowed,
				response: "Method not allowed",
			},
		},
		{
			name:   "case_3 Url not found",
			preset: "/unknown",
			want: want{
				method:   http.MethodGet,
				code:     http.StatusBadRequest,
				response: "Url not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest(tt.want.method, tt.preset, nil)
			w := httptest.NewRecorder()

			redirectHandler := RedirectById(urls)
			redirectHandler(w, r)

			result := w.Result()
			defer result.Body.Close()

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
