package gestacao

import (
	"context"
	"errors"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"gorm.io/gorm"
)

type Repository interface {
	FindGestacaoByID(ctx context.Context, id uint) (*models.Gestacao, error)
	CreateUltrassonografia(ctx context.Context, ultrassom *models.Ultrassonografia) error
	UpdateGestacao(ctx context.Context, gestacao *models.Gestacao) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindGestacaoByID(ctx context.Context, id uint) (*models.Gestacao, error) {
	var gestacao models.Gestacao
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&gestacao).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "gestacao", Message: "gestação não encontrada", ID: id}
		}
		return nil, apperrors.NewDatabaseError("find_by_id", "erro ao buscar gestação", err)
	}
	return &gestacao, nil
}

func (r *repository) CreateUltrassonografia(ctx context.Context, ultrassom *models.Ultrassonografia) error {
	if err := r.db.WithContext(ctx).Create(ultrassom).Error; err != nil {
		return apperrors.NewDatabaseError("create_ultrassom", "erro ao criar ultrassonografia", err)
	}
	return nil
}

func (r *repository) UpdateGestacao(ctx context.Context, gestacao *models.Gestacao) error {
	if err := r.db.WithContext(ctx).Save(gestacao).Error; err != nil {
		return apperrors.NewDatabaseError("update_gestacao", "erro ao atualizar gestação", err)
	}
	return nil
}
