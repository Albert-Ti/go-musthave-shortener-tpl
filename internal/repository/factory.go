package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
)

type ShortenURLRepository interface {
	Get(key string) (string, error)
	Save(key string, value string) error
	Length() (int, error)
	Close() error
	Ping() error
}

func NewShortenURLRepository(ctx context.Context) (ShortenURLRepository, error) {
	fmt.Println(config.Envs.DatabaseDSN)

	if config.Envs.DatabaseDSN != "" {
		slog.Info("Using database storage")
		return NewPostgresStorage(config.Envs.DatabaseDSN, ctx)
	}

	if config.Envs.FileStoragePath != "" {
		slog.Info("Using file storage", "filePath", config.Envs.FileStoragePath)
		return NewFileStorage(config.Envs.FileStoragePath)
	}

	slog.Info("Using memory storage")
	return NewMemoryStorage()
}
