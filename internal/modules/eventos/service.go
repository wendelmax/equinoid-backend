package eventos

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
)

type Service interface {
	ListAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.EventoResponse, int64, error)
	GetByID(ctx context.Context, id uint) (*models.EventoResponse, error)
	Create(ctx context.Context, userID uint, req *models.CreateEventoRequest) (*models.EventoResponse, error)
	Update(ctx context.Context, id uint, req *models.CreateEventoRequest) (*models.EventoResponse, error)
	Delete(ctx context.Context, id uint) error
	ListByEquino(ctx context.Context, equinoid string) ([]*models.EventoResponse, error)
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

func (s *service) ListAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.EventoResponse, int64, error) {
	eventos, total, err := s.repo.ListAll(ctx, page, limit, filters)
	if err != nil {
		s.logger.LogError(err, "EventoService.ListAll", logging.Fields{"filters": filters})
		return nil, 0, err
	}
	return eventos, total, nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*models.EventoResponse, error) {
	evento, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "EventoService.GetByID", logging.Fields{"evento_id": id})
		}
		return nil, err
	}
	return evento, nil
}

func (s *service) Create(ctx context.Context, userID uint, req *models.CreateEventoRequest) (*models.EventoResponse, error) {
	evento := &models.Evento{
		TipoEvento:            req.TipoEvento,
		Categoria:             req.Categoria,
		TipoEventoCompetitivo: req.TipoEventoCompetitivo,
		TipoEventoPublico:     req.TipoEventoPublico,
		NomeEvento:            req.NomeEvento,
		Descricao:             req.Descricao,
		DataEvento:            req.DataEvento,
		Local:                 req.Local,
		Organizador:           req.Organizador,
		VeterinarioID:         req.VeterinarioID,
		Resultados:            req.Resultados,
		Participante:          req.Participante,
		Particularidades:      req.Particularidades,
		ValorInscricao:        req.ValorInscricao,
		AceitaPatrocinio:      req.AceitaPatrocinio,
		InformacoesPatrocinio: req.InformacoesPatrocinio,
	}

	if err := s.repo.Create(ctx, evento); err != nil {
		s.logger.LogError(err, "EventoService.Create", logging.Fields{"user_id": userID})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{"evento_id": evento.ID, "nome": evento.NomeEvento}).Info("Evento criado com sucesso")
	
	return s.GetByID(ctx, evento.ID)
}

func (s *service) Update(ctx context.Context, id uint, req *models.CreateEventoRequest) (*models.EventoResponse, error) {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "EventoService.Update", logging.Fields{"evento_id": id})
		}
		return nil, err
	}

	eventoUpdate := &models.Evento{
		ID:                    id,
		NomeEvento:            req.NomeEvento,
		Descricao:             req.Descricao,
		DataEvento:            req.DataEvento,
		Local:                 req.Local,
		Organizador:           req.Organizador,
		Resultados:            req.Resultados,
		ValorInscricao:        req.ValorInscricao,
		AceitaPatrocinio:      req.AceitaPatrocinio,
		InformacoesPatrocinio: req.InformacoesPatrocinio,
	}

	if err := s.repo.Update(ctx, eventoUpdate); err != nil {
		s.logger.LogError(err, "EventoService.Update", logging.Fields{"evento_id": id})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{"evento_id": id}).Info("Evento atualizado com sucesso")
	return s.GetByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "EventoService.Delete", logging.Fields{"evento_id": id})
		}
		return err
	}

	s.logger.WithFields(logging.Fields{"evento_id": id}).Info("Evento deletado com sucesso")
	return nil
}

func (s *service) ListByEquino(ctx context.Context, equinoid string) ([]*models.EventoResponse, error) {
	eventos, err := s.repo.FindByEquino(ctx, equinoid)
	if err != nil {
		s.logger.LogError(err, "EventoService.ListByEquino", logging.Fields{"equinoid": equinoid})
		return nil, err
	}
	return eventos, nil
}
