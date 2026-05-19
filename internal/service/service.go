package service

import (
	"context"
	"errors"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
)

type Service struct {
	repository repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repository: repo}
}

func (s *Service) Get(ctx context.Context, key string) (string, error) {
	return s.repository.Get(ctx, key)
}

func (s *Service) GetAll(ctx context.Context) ([]model.JSONGetAllResp, error) {
	userID := ctx.Value(middleware.UserIDKey).(string)

	urls, err := s.repository.GetAll(ctx, userID)
	if err != nil {
		return nil, err
	}
	results := []model.JSONGetAllResp{}

	for i := range urls {
		results = append(results, model.JSONGetAllResp{
			ShortURL:    config.Envs.BaseURL + "/" + urls[i]["key"],
			OriginalURL: urls[i]["url"],
		})
	}

	return results, nil
}

func (s *Service) Save(ctx context.Context, url string) (string, bool, error) {
	key := utils.GenerateUUID()
	userID := ctx.Value(middleware.UserIDKey).(string)

	savedKey, err := s.repository.Save(ctx, key, url, userID)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return savedKey, false, nil
		}
		return "", false, err
	}

	return savedKey, true, nil
}

func (s *Service) BatchSave(ctx context.Context, batch []model.JSONBatchReq) ([]model.JSONBatchResp, error) {
	results := make([]model.JSONBatchResp, 0)
	hasConflict := false

	for _, v := range batch {
		key := utils.GenerateUUID()
		userID := ctx.Value(middleware.UserIDKey).(string)

		savedKey, err := s.repository.Save(ctx, key, v.OriginalURL, userID)
		if errors.Is(err, repository.ErrConflict) {
			hasConflict = true
		}
		if err != nil && !errors.Is(err, repository.ErrConflict) {
			return nil, err
		}

		results = append(results, model.JSONBatchResp{
			CorrelationID: v.CorrelationID,
			ShortURL:      config.Envs.BaseURL + "/" + savedKey,
		})
	}
	if hasConflict {
		return results, repository.ErrConflict
	}
	return results, nil
}

func (s *Service) Ping(ctx context.Context) error {
	return s.repository.Ping(ctx)
}
