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
	cfg        *config.Options
}

func NewService(repo repository.Repository, cfg *config.Options) *Service {
	return &Service{repo, cfg}
}

func (s *Service) Get(ctx context.Context, key string) (string, error) {
	return s.repository.Get(ctx, key)
}

func (s *Service) GetAll(ctx context.Context, userID string) ([]model.GetAllResp, error) {
	urls, err := s.repository.GetAll(ctx, userID)
	if err != nil {
		return nil, err
	}
	results := []model.GetAllResp{}

	for i := range urls {
		results = append(results, model.GetAllResp{
			ShortURL:    s.cfg.BaseURL + "/" + urls[i]["key"],
			OriginalURL: urls[i]["url"],
		})
	}

	return results, nil
}

func (s *Service) Save(ctx context.Context, url string, userID string) (string, bool, error) {
	key := utils.GenerateUUID()

	savedKey, err := s.repository.Save(ctx, key, url, userID)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return savedKey, false, nil
		}
		return "", false, err
	}

	return savedKey, true, nil
}

// func (s *Service) BatchSave(ctx context.Context, batch []model.BatchReq, userID string) ([]model.BatchResp, error) {
// 	results := make([]model.BatchResp, len(batch))
// 	prefix := s.cfg.BaseURL + "/"

// 	for i, v := range batch {
// 		key := utils.GenerateUUID()
// 		savedKey, err := s.repository.Save(ctx, key, v.OriginalURL, userID)

// 		if err != nil && !errors.Is(err, repository.ErrConflict) {
// 			return nil, err
// 		}

// 		results[i] = model.BatchResp{
// 			CorrelationID: v.CorrelationID,
// 			ShortURL:      prefix + savedKey,
// 		}
// 	}
// 	return results, nil
// }

func (s *Service) BatchSave(ctx context.Context, batch []model.BatchReq, userID string) ([]model.BatchResp, error) {
	return s.repository.BatchSave(ctx, batch, s.cfg.BaseURL, userID)
}

func (s *Service) BatchDelete(ctx context.Context, keys []string, userID string) error {
	err := s.repository.BatchDelete(ctx, keys, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Ping(ctx context.Context) error {
	return s.repository.Ping(ctx)
}
