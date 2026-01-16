package participacoes

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/pkg/cache"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
)

type Service interface {
	List(ctx context.Context, eventoID uint) ([]*models.ParticipacaoEvento, error)
	ListByEquino(ctx context.Context, equinoID uint) ([]*models.ParticipacaoEvento, error)
	GetByID(ctx context.Context, id uint) (*models.ParticipacaoEvento, error)
	Create(ctx context.Context, req *models.CreateParticipacaoEventoRequest, userID uint) (*models.ParticipacaoEvento, error)
	Update(ctx context.Context, id uint, req *models.UpdateParticipacaoEventoRequest) (*models.ParticipacaoEvento, error)
	Delete(ctx context.Context, id uint) error
	MarcarAusencia(ctx context.Context, id uint) (*models.ParticipacaoEvento, error)
	MarcarPresenca(ctx context.Context, id uint) (*models.ParticipacaoEvento, error)
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

func (s *service) List(ctx context.Context, eventoID uint) ([]*models.ParticipacaoEvento, error) {
	participacoes, err := s.repo.FindByEventoID(ctx, eventoID)
	if err != nil {
		s.logger.LogError(err, "ParticipacaoService.List", logging.Fields{"evento_id": eventoID})
		return nil, err
	}
	return participacoes, nil
}

func (s *service) ListByEquino(ctx context.Context, equinoID uint) ([]*models.ParticipacaoEvento, error) {
	participacoes, err := s.repo.FindByEquinoID(ctx, equinoID)
	if err != nil {
		s.logger.LogError(err, "ParticipacaoService.ListByEquino", logging.Fields{"equino_id": equinoID})
		return nil, err
	}
	return participacoes, nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*models.ParticipacaoEvento, error) {
	participacao, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "ParticipacaoService.GetByID", logging.Fields{"id": id})
		}
		return nil, err
	}
	return participacao, nil
}

func (s *service) Create(ctx context.Context, req *models.CreateParticipacaoEventoRequest, userID uint) (*models.ParticipacaoEvento, error) {
	participacao := &models.ParticipacaoEvento{
		EventoID:         req.EventoID,
		EquinoID:         req.EquinoID,
		ParticipanteID:   userID,
		Particularidades: req.Particularidades,
		Resultado:        req.Resultado,
		Classificacao:    req.Classificacao,
	}

	if err := s.repo.Create(ctx, participacao); err != nil {
		s.logger.LogError(err, "ParticipacaoService.Create", logging.Fields{
			"evento_id": req.EventoID,
			"equino_id": req.EquinoID,
		})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{
		"id":        participacao.ID,
		"evento_id": req.EventoID,
		"equino_id": req.EquinoID,
	}).Info("Participação criada com sucesso")

	return participacao, nil
}

func (s *service) Update(ctx context.Context, id uint, req *models.UpdateParticipacaoEventoRequest) (*models.ParticipacaoEvento, error) {
	participacao, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "ParticipacaoService.Update", logging.Fields{"id": id})
		}
		return nil, err
	}

	if req.Particularidades != nil {
		participacao.Particularidades = *req.Particularidades
	}
	if req.Resultado != nil {
		participacao.Resultado = *req.Resultado
	}
	if req.Classificacao != nil {
		participacao.Classificacao = req.Classificacao
	}

	if err := s.repo.Update(ctx, participacao); err != nil {
		s.logger.LogError(err, "ParticipacaoService.Update", logging.Fields{"id": id})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{"id": id}).Info("Participação atualizada com sucesso")

	return participacao, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "ParticipacaoService.Delete", logging.Fields{"id": id})
		}
		return err
	}

	s.logger.WithFields(logging.Fields{"id": id}).Info("Participação deletada com sucesso")

	return nil
}

func (s *service) MarcarAusencia(ctx context.Context, id uint) (*models.ParticipacaoEvento, error) {
	participacao, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "ParticipacaoService.MarcarAusencia", logging.Fields{"id": id})
		}
		return nil, err
	}

	falso := false
	participacao.Compareceu = &falso
	penalizacao := -50
	participacao.PenalizacaoAusencia = &penalizacao

	if err := s.repo.Update(ctx, participacao); err != nil {
		s.logger.LogError(err, "ParticipacaoService.MarcarAusencia", logging.Fields{"id": id})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{
		"id":           id,
		"penalizacao": penalizacao,
	}).Info("Ausência marcada com penalização")

	return participacao, nil
}

func (s *service) MarcarPresenca(ctx context.Context, id uint) (*models.ParticipacaoEvento, error) {
	participacao, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "ParticipacaoService.MarcarPresenca", logging.Fields{"id": id})
		}
		return nil, err
	}

	verdadeiro := true
	participacao.Compareceu = &verdadeiro
	participacao.PenalizacaoAusencia = nil

	if err := s.repo.Update(ctx, participacao); err != nil {
		s.logger.LogError(err, "ParticipacaoService.MarcarPresenca", logging.Fields{"id": id})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{"id": id}).Info("Presença confirmada")

	return participacao, nil
}
