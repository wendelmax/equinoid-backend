package gestacao

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/pkg/cache"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
)

type Service interface {
	CriarUltrassonografia(ctx context.Context, gestacaoID uint, req *models.CreateUltrassonografiaRequest) (*models.Ultrassonografia, error)
	RegistrarParto(ctx context.Context, gestacaoID uint, req *models.RegistrarPartoRequest) error
	RegistrarPerformanceMaterna(ctx context.Context, equinoid string, req *models.CreatePerformanceMaternaRequest) error
}

type service struct {
	repo   Repository
	cache  cache.CacheInterface
	logger *logging.Logger
}

func NewService(repo Repository, cache cache.CacheInterface, logger *logging.Logger) Service {
	return &service{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

func (s *service) CriarUltrassonografia(ctx context.Context, gestacaoID uint, req *models.CreateUltrassonografiaRequest) (*models.Ultrassonografia, error) {
	gestacao, err := s.repo.FindGestacaoByID(ctx, gestacaoID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "GestacaoService.CriarUltrassonografia", logging.Fields{"gestacao_id": gestacaoID})
		}
		return nil, err
	}

	if gestacao.StatusGestacao == models.StatusGestacaoConcluida {
		return nil, &apperrors.ValidationError{
			Field:   "status_gestacao",
			Message: "não é possível adicionar ultrassom a uma gestação finalizada",
		}
	}

	ultrassom := &models.Ultrassonografia{
		GestacaoID:             gestacaoID,
		DataExame:              req.DataExame,
		VeterinarioResponsavel: 1,
		IdadeGestacional:       req.IdadeGestacional,
		PresencaEmbriao:        req.PresencaEmbriao,
		NumeroEmbrioes:         req.NumeroEmbrioes,
		BatimentoCardiaco:      req.BatimentoCardiaco,
		DesenvolvimentoNormal:  req.DesenvolvimentoNormal,
		TamanhoEmbriao:         req.TamanhoEmbriao,
		FrequenciaCardiaca:     req.FrequenciaCardiaca,
		Diagnostico:            req.Diagnostico,
		Observacoes:            req.Observacoes,
		ProximoExame:           req.ProximoExame,
	}

	if err := s.repo.CreateUltrassonografia(ctx, ultrassom); err != nil {
		s.logger.LogError(err, "GestacaoService.CriarUltrassonografia", logging.Fields{"gestacao_id": gestacaoID})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{
		"ultrassom_id": ultrassom.ID,
		"gestacao_id":  gestacaoID,
	}).Info("Ultrassonografia criada com sucesso")

	return ultrassom, nil
}

func (s *service) RegistrarParto(ctx context.Context, gestacaoID uint, req *models.RegistrarPartoRequest) error {
	gestacao, err := s.repo.FindGestacaoByID(ctx, gestacaoID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "GestacaoService.RegistrarParto", logging.Fields{"gestacao_id": gestacaoID})
		}
		return err
	}

	if gestacao.StatusGestacao == models.StatusGestacaoConcluida {
		return &apperrors.ValidationError{
			Field:   "status_gestacao",
			Message: "gestação já finalizada",
		}
	}

	gestacao.StatusGestacao = models.StatusGestacaoConcluida

	if err := s.repo.UpdateGestacao(ctx, gestacao); err != nil {
		s.logger.LogError(err, "GestacaoService.RegistrarParto", logging.Fields{"gestacao_id": gestacaoID})
		return err
	}

	s.logger.WithFields(logging.Fields{
		"gestacao_id": gestacaoID,
		"data_parto":  req.DataParto,
	}).Info("Parto registrado com sucesso")

	return nil
}

func (s *service) RegistrarPerformanceMaterna(ctx context.Context, equinoid string, req *models.CreatePerformanceMaternaRequest) error {
	s.logger.WithFields(logging.Fields{
		"equinoid": equinoid,
		"performance": req,
	}).Info("Performance materna registrada")

	return nil
}
