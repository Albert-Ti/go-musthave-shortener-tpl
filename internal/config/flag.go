package config

import "flag"

var (
	FlagRunAddr string
	FlagBaseURL string
)

func ParseFlag() {
	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&FlagBaseURL, "b", "http://localhost:8080", "Base URL")

	flag.Parse()
}
