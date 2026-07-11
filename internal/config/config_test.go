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

	// все ENV-переменные, которые Build() умеет читать —
	// нужны, чтобы гарантированно зачищать их между кейсами
	allEnvKeys := []string{
		"SERVER_ADDRESS",
		"BASE_URL",
		"FILE_STORAGE_PATH",
		"DATABASE_CONN_STRING",
		"JWT_SECRET_KEY",
		"AUDIT_FILE",
		"AUDIT_URL",
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
			for _, key := range allEnvKeys {
				t.Setenv(key, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			origArgs := os.Args
			defer func() { os.Args = origArgs }()
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
