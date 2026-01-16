package participacoes

import (
	"context"
	"errors"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"gorm.io/gorm"
)

type Repository interface {
	FindByID(ctx context.Context, id uint) (*models.ParticipacaoEvento, error)
	FindByEventoID(ctx context.Context, eventoID uint) ([]*models.ParticipacaoEvento, error)
	FindByEquinoID(ctx context.Context, equinoID uint) ([]*models.ParticipacaoEvento, error)
	Create(ctx context.Context, participacao *models.ParticipacaoEvento) error
	Update(ctx context.Context, participacao *models.ParticipacaoEvento) error
	Delete(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByID(ctx context.Context, id uint) (*models.ParticipacaoEvento, error) {
	var participacao models.ParticipacaoEvento
	if err := r.db.WithContext(ctx).
		Preload("Evento").
		Preload("Equino").
		Preload("Participante").
		Where("id = ?", id).
		First(&participacao).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "participacao", Message: "participação não encontrada", ID: id}
		}
		return nil, apperrors.NewDatabaseError("find_by_id", "erro ao buscar participação", err)
	}
	return &participacao, nil
}

func (r *repository) FindByEventoID(ctx context.Context, eventoID uint) ([]*models.ParticipacaoEvento, error) {
	var participacoes []*models.ParticipacaoEvento
	if err := r.db.WithContext(ctx).
		Preload("Equino").
		Preload("Participante").
		Where("evento_id = ?", eventoID).
		Order("created_at DESC").
		Find(&participacoes).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_by_evento", "erro ao buscar participações do evento", err)
	}
	return participacoes, nil
}

func (r *repository) FindByEquinoID(ctx context.Context, equinoID uint) ([]*models.ParticipacaoEvento, error) {
	var participacoes []*models.ParticipacaoEvento
	if err := r.db.WithContext(ctx).
		Preload("Evento").
		Preload("Participante").
		Where("equino_id = ?", equinoID).
		Order("created_at DESC").
		Find(&participacoes).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_by_equino", "erro ao buscar participações do equino", err)
	}
	return participacoes, nil
}

func (r *repository) Create(ctx context.Context, participacao *models.ParticipacaoEvento) error {
	if err := r.db.WithContext(ctx).Create(participacao).Error; err != nil {
		return apperrors.NewDatabaseError("create", "erro ao criar participação", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, participacao *models.ParticipacaoEvento) error {
	if err := r.db.WithContext(ctx).Save(participacao).Error; err != nil {
		return apperrors.NewDatabaseError("update", "erro ao atualizar participação", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.ParticipacaoEvento{}, id)
	if result.Error != nil {
		return apperrors.NewDatabaseError("delete", "erro ao deletar participação", result.Error)
	}
	if result.RowsAffected == 0 {
		return &apperrors.NotFoundError{Resource: "participacao", Message: "participação não encontrada", ID: id}
	}
	return nil
}
