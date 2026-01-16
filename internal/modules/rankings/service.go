package rankings

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/pkg/logging"
)

type Service interface {
	GetRankingGeral(ctx context.Context, tipo string) ([]*models.RankingItem, error)
	GetRankingsEquino(ctx context.Context, equinoid string) ([]*models.RankingItem, error)
}

type service struct {
	repo   Repository
	logger *logging.Logger
}

func NewService(repo Repository, logger *logging.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) GetRankingGeral(ctx context.Context, tipo string) ([]*models.RankingItem, error) {
	limit := 100
	if tipo == "top10" {
		limit = 10
	} else if tipo == "top50" {
		limit = 50
	}

	items, err := s.repo.GetRankingGeral(ctx, tipo, limit)
	if err != nil {
		s.logger.LogError(err, "RankingService.GetRankingGeral", logging.Fields{"tipo": tipo})
		return nil, err
	}

	return items, nil
}

func (s *service) GetRankingsEquino(ctx context.Context, equinoid string) ([]*models.RankingItem, error) {
	items, err := s.repo.GetRankingsEquino(ctx, equinoid)
	if err != nil {
		s.logger.LogError(err, "RankingService.GetRankingsEquino", logging.Fields{"equinoid": equinoid})
		return nil, err
	}

	return items, nil
}
