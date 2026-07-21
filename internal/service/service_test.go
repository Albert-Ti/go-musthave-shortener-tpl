package service

import (
	"context"
	"strconv"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository/mocks"
	"github.com/golang/mock/gomock"
)

func BenchmarkBatchSave(b *testing.B) {
	tests := []int{1, 10, 100, 1000, 10000}

	for _, n := range tests {
		b.Run("Size_"+strconv.Itoa(n), func(b *testing.B) {
			batch := make([]model.BatchReq, 0, n)

			for i := 0; i < n; i++ {
				batch = append(batch, model.BatchReq{
					CorrelationID: "test",
					OriginalURL:   "https://example.com/" + strconv.Itoa(i),
				})
			}

			ctrl := gomock.NewController(b)
			mockRepo := mocks.NewMockRepository(ctrl)

			ctx := context.WithValue(context.Background(), middleware.UserIDKey, "user-1")
			cfg := config.NewOptions(config.WithBaseURL("http://localhost:8080"))

			svc := NewService(mockRepo, cfg)

			mockRepo.EXPECT().
				BatchSave(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, nil).AnyTimes()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if _, err := svc.BatchSave(ctx, batch, "user-1"); err != nil {
					b.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}
