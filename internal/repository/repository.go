package repository

import (
	"context"
	"log/slog"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // ← для работы с PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/source/file"       // ← для чтения файлов
)

type Repository interface {
	Get(ctx context.Context, key string) (string, error)
	GetAll(ctx context.Context, userID string) ([]map[string]string, error)
	Save(ctx context.Context, key string, url string, userID string) (string, error)
	BatchDelete(ctx context.Context, keys []string, userID string) error
	Ping(ctx context.Context) error
	Close() error
}

// pattern Factory
func NewRepository() (Repository, error) {
	if config.Envs.DatabaseDSN != "" {
		slog.Info("Using database storage")
		return NewPostgresStorage(config.Envs.DatabaseDSN)
	}

	if config.Envs.FileStoragePath != "" {
		slog.Debug("Using file storage", "filePath", config.Envs.FileStoragePath)
		return NewFileStorage(config.Envs.FileStoragePath)
	}

	slog.Debug("Using memory storage")
	return NewMemoryStorage()
}
