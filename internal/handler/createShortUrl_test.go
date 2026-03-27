package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShortUrl(t *testing.T) {
	urls := &model.ShortenedUrls{List: map[string]string{}, Count: 1}

	handler := http.HandlerFunc(CreateShortUrl(urls))
	srv := httptest.NewServer(handler)

	defer srv.Close()

	type want struct {
		method   string
		code     int
		response string
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
				method:   http.MethodPost,
				code:     http.StatusCreated,
				response: srv.URL + "/key_1",
			},
		},
		{
			name:   "case_2 Method Not Allowed",
			preset: "https://yandex.ru",
			want: want{
				method:   http.MethodGet,
				code:     http.StatusMethodNotAllowed,
				response: "Method not allowed",
			},
		},
		{
			name:   "case_3 No body",
			preset: "",
			want: want{
				method:   http.MethodPost,
				code:     http.StatusBadRequest,
				response: "Invalid URL",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			body := strings.NewReader("https://yandex.ru")
			if tt.preset == "" {
				body = strings.NewReader("")
			}

			var (
				resp *http.Response
				err  error
			)
			if tt.want.method == http.MethodGet {
				resp, err = http.Get(srv.URL)
			} else {
				resp, err = http.Post(srv.URL, "text/plain", body)
			}
			require.NoError(t, err)
			defer resp.Body.Close()

			responseBody, _ := io.ReadAll(resp.Body)
			gotBody := strings.TrimSpace(string(responseBody))

			if tt.want.response != "" {
				assert.Equal(t, tt.want.response, gotBody)
			}
			assert.Equal(t, tt.want.code, resp.StatusCode)
		})
	}
}
