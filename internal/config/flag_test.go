package config_test

import (
	"flag"
	"os"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestParseFlag(t *testing.T) {
	type want struct {
		runAddr string
		baseUrl string
	}

	tests := []struct {
		name string
		args []string
		env  map[string]string
		want want
	}{
		{
			name: "case_1 add env",
			args: []string{"test"},
			env: map[string]string{
				"SERVER_ADDRESS": "localhost:8888",
				"BASE_URL":       "http://localhost:8888",
			},
			want: want{
				runAddr: "localhost:8888",
				baseUrl: "http://localhost:8888",
			},
		},
		{
			name: "case_2 add flag",
			args: []string{"test", "-a=localhost:9090", "-b=http://localhost:9090"},
			env:  map[string]string{},
			want: want{
				runAddr: "localhost:9090",
				baseUrl: "http://localhost:9090",
			},
		},
		{
			name: "case_2 default",
			args: []string{"test"},
			env:  map[string]string{},
			want: want{
				runAddr: "localhost:8080",
				baseUrl: "http://localhost:8080",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			os.Args = tt.args

			config.ParseFlag()

			assert.Equal(t, tt.want.runAddr, config.Envs.RunAddr)
			assert.Equal(t, tt.want.baseUrl, config.Envs.BaseURL)
		})
	}
}
