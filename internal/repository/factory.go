package repository

import (
	"errors"
	"log/slog"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
)

type ShortenURLRepository interface {
	Get(key string) string
	Save(key string, value string)
	Length() int
	Close() error
}

func NewShortenURLRepository() (ShortenURLRepository, error) {
	if config.Envs.DatabaseDSN != "" {
		slog.Info("Using database storage")
		return nil, errors.New("")
	}

	if config.Envs.FileStoragePath != "" {
		slog.Info("Using file storage", "filePath", config.Envs.FileStoragePath)
		return NewFileStorage(config.Envs.FileStoragePath)
	}

	slog.Info("Using in-memory storage")
	return NewMemoryStorage()
}
