package service

import (
	"strconv"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
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

func (u *Service) Save(url string) (string, error) {
	len, err := u.repository.Length()
	if err != nil {
		return "", err
	}
	key := "key_" + strconv.Itoa(len+1)
	return u.repository.Save(key, url)
}

func (u *Service) BatchSave(batch []model.JSONBatchReq) ([]model.JSONBatchResp, error) {
	if len(batch) == 0 {
		return nil, nil
	}

	length, err := u.repository.Length()
	if err != nil {
		return nil, err
	}

	result := make([]model.JSONBatchResp, len(batch))
	keys := make([]string, len(batch))

	for i, v := range batch {
		key := "key_" + strconv.Itoa(length+1)
		keys[i] = key
		result[i] = model.JSONBatchResp{
			ShortURL:      config.Envs.BaseURL + "/" + key,
			CorrelationID: v.CorrelationID,
		}

		length++
	}

	existKey, existID, err := u.repository.BatchSave(keys, batch)

	if existKey != "" {
		return []model.JSONBatchResp{{
			CorrelationID: existID,
			ShortURL:      config.Envs.BaseURL + "/" + existKey,
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
