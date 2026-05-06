package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // ← для работы с PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/source/file"       // ← для чтения файлов
)

type Repository interface {
	Get(key string) (string, error)
	Save(key string, value string) error
	BatchSave(keys []string, batch []model.JSONBatchReq) error
	Length() (int, error)
	Close() error
	Ping() error
}

func NewRepository(ctx context.Context) (Repository, error) {
	fmt.Println(config.Envs.DatabaseDSN)

	if config.Envs.DatabaseDSN != "" {
		slog.Info("Using database storage")
		slog.Info("Applying migrations to the database...")

		m, err := migrate.New("file://migrations", config.Envs.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		defer m.Close()

		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			slog.Error("Migrations failed", "error", err)
			return nil, err
		}

		if err == migrate.ErrNoChange {
			slog.Info("The database schema is up-to-date and no migrations are required")
		} else {
			slog.Info("Migrations have been successfully applied")
		}

		return NewPostgresStorage(config.Envs.DatabaseDSN, ctx)
	}

	if config.Envs.FileStoragePath != "" {
		slog.Info("Using file storage", "filePath", config.Envs.FileStoragePath)
		return NewFileStorage(config.Envs.FileStoragePath)
	}

	slog.Info("Using memory storage")
	return NewMemoryStorage()
}
