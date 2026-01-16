package equinos

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/internal/utils"
	"github.com/equinoid/backend/pkg/cache"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
)

type D4SignService interface {
	RegisterDocument(ctx context.Context, equinoid string, docType string, filePath string) (string, error)
}

type Service interface {
	List(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.Equino, int64, error)
	GetByEquinoid(ctx context.Context, equinoidID string) (*models.Equino, error)
	Create(ctx context.Context, req *models.CreateEquinoRequest, userID uint) (*models.Equino, error)
	Update(ctx context.Context, equinoidID string, req *models.UpdateEquinoRequest) (*models.Equino, error)
	Delete(ctx context.Context, equinoidID string) error
	TransferOwnership(ctx context.Context, equinoidID string, newOwnerID uint) error
}

type service struct {
	repo          Repository
	cache         cache.CacheInterface
	logger        *logging.Logger
	d4signService D4SignService
}

func NewService(repo Repository, cache cache.CacheInterface, logger *logging.Logger, d4signService D4SignService) Service {
	return &service{
		repo:          repo,
		cache:         cache,
		logger:        logger,
		d4signService: d4signService,
	}
}

func (s *service) List(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.Equino, int64, error) {
	equinos, total, err := s.repo.List(ctx, page, limit, filters)
	if err != nil {
		s.logger.LogError(err, "EquinoService.List", logging.Fields{"filters": filters})
		return nil, 0, err
	}
	return equinos, total, nil
}

func (s *service) GetByEquinoid(ctx context.Context, equinoidID string) (*models.Equino, error) {
	equino, err := s.repo.FindByEquinoid(ctx, equinoidID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "EquinoService.GetByEquinoid", logging.Fields{"equinoid": equinoidID})
		}
		return nil, err
	}
	return equino, nil
}

func (s *service) Create(ctx context.Context, req *models.CreateEquinoRequest, userID uint) (*models.Equino, error) {
	equinoidID, err := utils.GenerateEquinoId(
		req.PaisOrigem,
		req.MicrochipID,
		req.Nome,
		*req.DataNascimento,
		string(req.Sexo),
		req.Pelagem,
		req.Raca,
	)
	if err != nil {
		s.logger.LogError(err, "EquinoService.Create", logging.Fields{"action": "generate_equinoid"})
		return nil, apperrors.NewBusinessError("EQUINOID_GENERATION_FAILED", "erro ao gerar EquinoId", nil)
	}
	
	exists, err := s.repo.ExistsByEquinoid(ctx, equinoidID)
	if err != nil {
		s.logger.LogError(err, "EquinoService.Create", logging.Fields{"equinoid": equinoidID})
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrEquinoidExists.WithValue(equinoidID)
	}

	if req.MicrochipID != "" {
		exists, err := s.repo.ExistsByMicrochipID(ctx, req.MicrochipID)
		if err != nil {
			s.logger.LogError(err, "EquinoService.Create", logging.Fields{"microchip_id": req.MicrochipID})
			return nil, err
		}
		if exists {
			return nil, apperrors.ErrMicrochipExists.WithValue(req.MicrochipID)
		}
	}

	equino := &models.Equino{
		Nome:           req.Nome,
		Equinoid:       equinoidID,
		MicrochipID:    req.MicrochipID,
		Sexo:           req.Sexo,
		Raca:           req.Raca,
		Pelagem:        req.Pelagem,
		DataNascimento: req.DataNascimento,
		PaisOrigem:     req.PaisOrigem,
		ProprietarioID: req.ProprietarioID,
		Status:         "ativo",
	}

	if err := s.repo.Create(ctx, equino); err != nil {
		s.logger.LogError(err, "EquinoService.Create", logging.Fields{"equinoid": equino.Equinoid})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{
		"equinoid":        equino.Equinoid,
		"proprietario_id": userID,
	}).Info("Equino criado com sucesso")

	return equino, nil
}

func (s *service) Update(ctx context.Context, equinoidID string, req *models.UpdateEquinoRequest) (*models.Equino, error) {
	equino, err := s.repo.FindByEquinoid(ctx, equinoidID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "EquinoService.Update", logging.Fields{"equinoid": equinoidID})
		}
		return nil, err
	}

	if req.Nome != nil && *req.Nome != "" {
		equino.Nome = *req.Nome
	}
	if req.Raca != nil && *req.Raca != "" {
		equino.Raca = *req.Raca
	}
	if req.Pelagem != nil && *req.Pelagem != "" {
		equino.Pelagem = *req.Pelagem
	}

	if err := s.repo.Update(ctx, equino); err != nil {
		s.logger.LogError(err, "EquinoService.Update", logging.Fields{"equinoid": equinoidID})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{"equinoid": equinoidID}).Info("Equino atualizado com sucesso")

	return equino, nil
}

func (s *service) Delete(ctx context.Context, equinoidID string) error {
	if err := s.repo.Delete(ctx, equinoidID); err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "EquinoService.Delete", logging.Fields{"equinoid": equinoidID})
		}
		return err
	}

	s.logger.WithFields(logging.Fields{"equinoid": equinoidID}).Info("Equino deletado com sucesso")

	return nil
}

func (s *service) TransferOwnership(ctx context.Context, equinoidID string, newOwnerID uint) error {
	equino, err := s.repo.FindByEquinoid(ctx, equinoidID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "EquinoService.TransferOwnership", logging.Fields{"equinoid": equinoidID})
		}
		return err
	}

	if err := s.repo.TransferOwnership(ctx, equinoidID, newOwnerID); err != nil {
		s.logger.LogError(err, "EquinoService.TransferOwnership", logging.Fields{
			"equinoid":     equinoidID,
			"new_owner_id": newOwnerID,
		})
		return err
	}

	s.logger.WithFields(logging.Fields{
		"equinoid":     equinoidID,
		"old_owner_id": equino.ProprietarioID,
		"new_owner_id": newOwnerID,
	}).Info("Propriedade transferida com sucesso")

	return nil
}
