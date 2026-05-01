package config

import (
	"flag"
	"os"
)

var Envs struct {
	RunAddr         string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func ParseFlag() {
	postgresConnStr := "postgres://postgres:postgres@localhost:5432/shorten_url_db"

	flag.StringVar(&Envs.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&Envs.BaseURL, "b", "http://localhost:8080", "Base URL")
	flag.StringVar(&Envs.FileStoragePath, "f", "shortenUrlList.json", "Base URL")
	flag.StringVar(&Envs.DatabaseDSN, "d", postgresConnStr, "Base URL")

	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		Envs.RunAddr = envRunAddr
	}
	if envBaseUrl := os.Getenv("BASE_URL"); envBaseUrl != "" {
		Envs.BaseURL = envBaseUrl
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		Envs.FileStoragePath = envFileStoragePath
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		Envs.DatabaseDSN = envDatabaseDSN
	}
}
