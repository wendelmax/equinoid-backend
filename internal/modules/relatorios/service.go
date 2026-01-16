package relatorios

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/pkg/logging"
)

type Service interface {
	GetDashboardStats(ctx context.Context) (*models.DashboardStats, error)
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

func (s *service) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	stats, err := s.repo.GetDashboardStats(ctx)
	if err != nil {
		s.logger.LogError(err, "RelatorioService.GetDashboardStats", logging.Fields{})
		return nil, err
	}

	return stats, nil
}
