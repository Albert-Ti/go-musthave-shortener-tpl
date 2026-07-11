package config_test

import (
	"os"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	type want struct {
		runAddr  string
		baseURL  string
		filePath string
		dsn      string
	}

	tests := []struct {
		name string
		args []string
		env  map[string]string
		want want
	}{
		{
			name: "env only",
			args: []string{"test"},
			env: map[string]string{
				"SERVER_ADDRESS": "localhost:8888",
				"BASE_URL":       "http://localhost:8888",
			},
			want: want{
				runAddr: "localhost:8888",
				baseURL: "http://localhost:8888",
			},
		},
		{
			name: "flag only",
			args: []string{"test", "-a=localhost:9090", "-b=http://localhost:9090"},
			want: want{
				runAddr: "localhost:9090",
				baseURL: "http://localhost:9090",
			},
		},
		{
			name: "defaults, nothing set",
			args: []string{"test"},
			want: want{
				runAddr: "localhost:8080",
				baseURL: "http://localhost:8080",
			},
		},
		{
			name: "explicit flag wins over env for the same field",
			args: []string{"test", "-a=localhost:9090"},
			env: map[string]string{
				"SERVER_ADDRESS": "localhost:8888",
			},
			want: want{
				runAddr: "localhost:9090",
				baseURL: "http://localhost:8080",
			},
		},
		{
			name: "DATABASE_CONN_STRING from env is used when -f is not passed",
			args: []string{"test"},
			env: map[string]string{
				"DATABASE_CONN_STRING": "postgres://localhost/db",
			},
			want: want{
				dsn: "postgres://localhost/db",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			os.Args = tt.args

			cfg := config.Build()

			if tt.want.runAddr != "" {
				assert.Equal(t, tt.want.runAddr, cfg.RunAddr)
			}
			if tt.want.baseURL != "" {
				assert.Equal(t, tt.want.baseURL, cfg.BaseURL)
			}
			assert.Equal(t, tt.want.filePath, cfg.FileStoragePath)
			assert.Equal(t, tt.want.dsn, cfg.DatabaseDSN)
		})
	}
}
