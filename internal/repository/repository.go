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
		slog.Info("Использование хранилища базы данных")
		slog.Info("Применяем миграции к базе данных...")

		m, err := migrate.New("file://migrations", config.Envs.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		defer m.Close()

		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			slog.Error("Не удалось выполнить миграции", "error", err)
			return nil, err
		}

		if err == migrate.ErrNoChange {
			slog.Info("Схема базы данных актуальна, миграции не требуются")
		} else {
			slog.Info("Миграции успешно применены")
		}

		return NewPostgresStorage(config.Envs.DatabaseDSN, ctx)
	}

	if config.Envs.FileStoragePath != "" {
		slog.Info("Использование файлового хранилища", "filePath", config.Envs.FileStoragePath)
		return NewFileStorage(config.Envs.FileStoragePath)
	}

	slog.Info("Использование памяти")
	return NewMemoryStorage()
}
