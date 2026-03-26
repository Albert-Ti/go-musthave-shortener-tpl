package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
)

func TestCreateShortUrl(t *testing.T) {
	type want struct {
		method      string
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name   string
		preset string
		want   want
	}{
		{
			name:   "case_1 Ok",
			preset: "https://yandex.ru",
			want: want{
				method:      http.MethodPost,
				code:        http.StatusCreated,
				response:    "http://example.com/key_1",
				contentType: "text/plain",
			},
		},
		{
			name:   "case_2 Method Not Allowed",
			preset: "https://yandex.ru",
			want: want{
				method:      http.MethodGet,
				code:        http.StatusMethodNotAllowed,
				response:    "Method not allowed",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case_3 No body",
			preset: "",
			want: want{
				method:      http.MethodPost,
				code:        http.StatusBadRequest,
				response:    "Invalid URL",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls := &model.ShortenedUrls{
				List:  map[uint]string{},
				Count: 1,
			}

			body := strings.NewReader("https://yandex.ru")
			if tt.preset == "" {
				body = strings.NewReader("")
			}

			r := httptest.NewRequest(tt.want.method, "/", body)
			w := httptest.NewRecorder()
			w.Header().Set("Content-Type", "text/plain")

			createShortUrlHandler := CreateShortUrl(urls)

			createShortUrlHandler(w, r)

			result := w.Result()
			defer result.Body.Close()

			responseBody, _ := io.ReadAll(result.Body)
			got := strings.TrimSpace(string(responseBody))
			want := tt.want.response

			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
			if result.StatusCode != tt.want.code {
				t.Errorf("got %d, want %d", result.StatusCode, tt.want.code)
			}
			if result.Header.Get("Content-type") != tt.want.contentType {
				t.Errorf("got %q, want %q", result.Header.Get("Content-type"), tt.want.contentType)
			}
		})
	}
}
