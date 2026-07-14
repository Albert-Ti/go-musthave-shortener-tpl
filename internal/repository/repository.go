package repository

import (
	"context"
	"log/slog"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // ← для работы с PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/source/file"       // ← для чтения файлов
)

type Repository interface {
	Get(ctx context.Context, key string) (string, error)
	GetAll(ctx context.Context, userID string) ([]map[string]string, error)
	Save(ctx context.Context, key string, url string, userID string) (string, error)
	BatchSave(ctx context.Context, batch []model.BatchReq, baseURL string, userID string) ([]model.BatchResp, error)
	BatchDelete(ctx context.Context, keys []string, userID string) error
	Ping(ctx context.Context) error
	Close() error
}

// pattern Factory(Фабрика)
func NewRepository(cfg *config.Options) (Repository, error) {
	if cfg.DatabaseDSN != "" {
		slog.Info("Using database storage")
		return NewPostgresStorage(cfg.DatabaseDSN)
	}

	if cfg.FileStoragePath != "" {
		slog.Debug("Using file storage", "filePath", cfg.FileStoragePath)
		return NewFileStorage(cfg.FileStoragePath)
	}

	slog.Debug("Using memory storage")
	return NewMemoryStorage()
}
