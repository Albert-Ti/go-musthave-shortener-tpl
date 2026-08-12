package utils_test

import (
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID(t *testing.T) {
	t.Run("length UUID", func(t *testing.T) {
		uuid := utils.GenerateUUID()
		// 9 байт × 8 бит = 72 бита / 6 бит = 12 символов
		assert.Equal(t, 9*8/6, len(uuid))

	})

	t.Run("Uniqueness", func(t *testing.T) {
		const n = 100000
		seen := make(map[string]struct{}, n)

		for range n {
			uuid := utils.GenerateUUID()
			_, ok := seen[uuid]
			assert.False(t, ok, "collision detected: %s", uuid)
			seen[uuid] = struct{}{}
		}
	})
}
