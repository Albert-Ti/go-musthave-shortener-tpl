package repository

import (
	"context"
	"log/slog"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // ← для работы с PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/source/file"       // ← для чтения файлов
)

type BatchConflict struct {
	Key           string
	CorrelationID string
}

type Repository interface {
	Get(ctx context.Context, key string) (string, error)
	Save(ctx context.Context, key string, value string) (string, error)
	BatchSave(ctx context.Context, keys []string, batch []model.JSONBatchReq) (BatchConflict, error)
	Ping(ctx context.Context) error
	Close() error
}

func NewRepository() (Repository, error) {
	if config.Envs.DatabaseDSN != "" {
		slog.Debug("Using database storage")

		m, err := migrate.New("file://migrations", config.Envs.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		defer m.Close()

		err = m.Up()
		switch err {
		case nil:
			slog.Debug("Migrations have been successfully applied")
		case migrate.ErrNoChange:
			slog.Debug("The database schema is up-to-date and no migrations are required")
		default:
			slog.Error("Migrations failed", "error", err)
			return nil, err
		}

		return NewPostgresStorage(config.Envs.DatabaseDSN)
	}

	if config.Envs.FileStoragePath != "" {
		slog.Debug("Using file storage", "filePath", config.Envs.FileStoragePath)
		return NewFileStorage(config.Envs.FileStoragePath)
	}

	slog.Debug("Using memory storage")
	return NewMemoryStorage()
}
