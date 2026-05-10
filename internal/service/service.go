package service

import (
	"errors"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/google/uuid"
)

type Service struct {
	repository repository.Repository
}

func NewService(r repository.Repository) *Service {
	return &Service{
		repository: r,
	}
}

func (u *Service) Get(key string) (string, error) {
	return u.repository.Get(key)
}

func (s *Service) Save(url string) (string, bool, error) {
	key := GenerateUUID()

	savedKey, err := s.repository.Save(key, url)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return savedKey, false, nil
		}
		return "", false, err
	}

	return savedKey, true, nil
}

func (u *Service) BatchSave(batch []model.JSONBatchReq) ([]model.JSONBatchResp, error) {

	result := make([]model.JSONBatchResp, len(batch))
	keys := make([]string, len(batch))

	for i, v := range batch {
		key := GenerateUUID()
		keys[i] = key

		result[i] = model.JSONBatchResp{
			ShortURL:      config.Envs.BaseURL + "/" + key,
			CorrelationID: v.CorrelationID,
		}
	}

	existRow, err := u.repository.BatchSave(keys, batch)

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

func (u *Service) Ping() error {
	return u.repository.Ping()
}

var GenerateUUID = func() string {
	return uuid.NewString()[:5]
}
