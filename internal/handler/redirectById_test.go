package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
)

func TestRedirect(t *testing.T) {
	type want struct {
		method      string
		code        int
		location    string
		response    string
		contentType string
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
				method:      http.MethodGet,
				location:    "http://yandex.ru",
				code:        http.StatusTemporaryRedirect,
				contentType: "text/html; charset=utf-8",
			},
		},
		{
			name:   "case_2 Method Not Allowed",
			preset: "/key_1",
			want: want{
				method:      http.MethodPost,
				code:        http.StatusMethodNotAllowed,
				response:    "Method not allowed",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case_3 Url not found",
			preset: "/",
			want: want{
				method:      http.MethodGet,
				code:        http.StatusBadRequest,
				response:    "Url not found",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls := &model.ShortenedUrls{List: map[string]string{"key_1": "http://yandex.ru"}}

			r := httptest.NewRequest(tt.want.method, tt.preset, nil)
			w := httptest.NewRecorder()

			redirectHandler := RedirectById(urls)
			redirectHandler(w, r)

			result := w.Result()
			defer result.Body.Close()

			responseBody, _ := io.ReadAll(result.Body)
			gotBody := strings.TrimSpace(string(responseBody))

			if tt.want.response != "" && gotBody != tt.want.response {
				t.Errorf("got body %q, want %q", gotBody, tt.want.response)
			}
			if tt.want.location != "" && result.Header.Get("Location") != tt.want.location {
				t.Errorf("got  location %s, want %s", result.Header.Get("Location"), tt.want.location)
			}
			if result.StatusCode != tt.want.code {
				t.Errorf("got status code %d, want %d", result.StatusCode, tt.want.code)
			}
			if result.Header.Get("Content-type") != tt.want.contentType {
				t.Errorf("got content type %q, want %q", result.Header.Get("Content-type"), tt.want.contentType)
			}
		})
	}
}
