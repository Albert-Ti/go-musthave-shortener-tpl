package service

import (
	"context"
	"errors"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
)

type Service struct {
	repository repository.Repository
}

func NewService(r repository.Repository) *Service {
	return &Service{
		repository: r,
	}
}

func (u *Service) Get(ctx context.Context, key string) (string, error) {
	return u.repository.Get(ctx, key)
}

func (s *Service) Save(ctx context.Context, url string) (string, bool, error) {
	key := utils.GenerateUUID()

	savedKey, err := s.repository.Save(ctx, key, url)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return savedKey, false, nil
		}
		return "", false, err
	}

	return savedKey, true, nil
}

func (u *Service) BatchSave(ctx context.Context, batch []model.JSONBatchReq) ([]model.JSONBatchResp, error) {

	result := make([]model.JSONBatchResp, len(batch))
	keys := make([]string, len(batch))

	for i, v := range batch {
		key := utils.GenerateUUID()
		keys[i] = key

		result[i] = model.JSONBatchResp{
			ShortURL:      config.Envs.BaseURL + "/" + key,
			CorrelationID: v.CorrelationID,
		}
	}

	existRow, err := u.repository.BatchSave(ctx, keys, batch)

	if existRow.Key != "" {
		return []model.JSONBatchResp{{
			CorrelationID: existRow.CorrelationID,
			ShortURL:      config.Envs.BaseURL + "/" + existRow.Key,
		},
		}, err
	} else if err != nil {
		return nil, err
	}

	return result, nil
}

func (u *Service) Ping(ctx context.Context) error {
	return u.repository.Ping(ctx)
}
