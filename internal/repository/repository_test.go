package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {

	t.Run("file storage memento rollback on encode error", func(t *testing.T) {
		tmpDir := t.TempDir()
		userID := "user-1"
		ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)

		file, err := os.OpenFile(
			filepath.Join(tmpDir, "test.json"),
			os.O_RDWR|os.O_CREATE|os.O_APPEND,
			0666,
		)
		require.NoError(t, err)

		urls := []map[string]string{
			{"key": "key-1", "url": "http://example.com", "user_id": userID},
			{"key": "key-2", "url": "http://example.com", "user_id": userID},
		}

		fs := &fileStorage{
			file:    file,
			urls:    urls,
			encoder: json.NewEncoder(file),
		}

		require.NoError(t, file.Close())

		_, err = fs.Save(ctx, "key_error", "http://error.com", userID)
		require.Error(t, err)

		require.Equal(t, urls, fs.urls)
		require.Len(t, fs.urls, 2)
	})
}
