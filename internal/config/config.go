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
	JWTSecretKey    string
	AuditFile       string
	AuditURL        string
}

func ParseFlag() {
	flag.StringVar(&Envs.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&Envs.BaseURL, "b", "http://localhost:8080", "Base URL")
	flag.StringVar(&Envs.FileStoragePath, "f", "", "file storage")
	flag.StringVar(&Envs.DatabaseDSN, "d", "", "connection string to DB")
	flag.StringVar(&Envs.AuditFile, "audit-file", "", "путь к файлу-приёмнику")
	flag.StringVar(&Envs.AuditURL, "audit-url", "", "URL удаленного сервера-приёмника")

	Envs.JWTSecretKey = "secret_key"

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

	if envJWTSecretKey := os.Getenv("JWT_SECRET_KEY"); envJWTSecretKey != "" {
		Envs.JWTSecretKey = envJWTSecretKey
	}

	if envAuditFile := os.Getenv("AUDIT_FILE"); envAuditFile != "" {
		Envs.AuditFile = envAuditFile
	}

	if envAuditURL := os.Getenv("AUDIT_URL"); envAuditURL != "" {
		Envs.AuditURL = envAuditURL
	}
}

func IsAuditorDisabled() bool {
	return Envs.AuditFile == "" && Envs.AuditURL == ""
}
