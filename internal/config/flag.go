package config

import (
	"flag"
	"os"
)

var Envs struct {
	RunAddr string
	BaseURL string
}

func ParseFlag() {
	flag.StringVar(&Envs.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&Envs.BaseURL, "b", "http://localhost:8080", "Base URL")

	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		Envs.RunAddr = envRunAddr
	}
	if envBaseUrl := os.Getenv("BASE_URL"); envBaseUrl != "" {
		Envs.BaseURL = envBaseUrl
	}
}
