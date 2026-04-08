package config

import (
	"flag"
	"os"
)

var (
	FlagRunAddr string
	FlagBaseURL string
)

func ParseFlag() {
	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&FlagBaseURL, "b", "http://localhost:8080", "Base URL")

	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		FlagRunAddr = envRunAddr
	}
	if envBaseUrl := os.Getenv("BASE_URL"); envBaseUrl != "" {
		FlagBaseURL = envBaseUrl
	}
}
