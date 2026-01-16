package nutricao

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"gorm.io/gorm"
)

type Repository interface {
	FindByEquinoid(ctx context.Context, equinoid string) (*models.PlanoNutricional, error)
	FindRefeicoesByEquinoid(ctx context.Context, equinoid string) ([]*models.Refeicao, error)
	Create(ctx context.Context, plano *models.PlanoNutricional) error
	CreateRefeicao(ctx context.Context, refeicao *models.Refeicao) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByEquinoid(ctx context.Context, equinoid string) (*models.PlanoNutricional, error) {
	var plano models.PlanoNutricional
	if err := r.db.WithContext(ctx).
		Preload("Equino").
		Preload("Veterinario").
		Preload("Nutricionista").
		Joins("JOIN equinos ON equinos.id = planos_nutricionais.equino_id").
		Where("equinos.equinoid = ? AND planos_nutricionais.status = ?", equinoid, models.StatusPlanoAtivo).
		First(&plano).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &apperrors.NotFoundError{Resource: "plano_nutricional", Message: "plano nutricional não encontrado"}
		}
		return nil, apperrors.NewDatabaseError("find_plano", "erro ao buscar plano", err)
	}
	return &plano, nil
}

func (r *repository) FindRefeicoesByEquinoid(ctx context.Context, equinoid string) ([]*models.Refeicao, error) {
	var refeicoes []*models.Refeicao
	if err := r.db.WithContext(ctx).
		Preload("PlanoNutricional").
		Preload("Equino").
		Preload("Usuario").
		Joins("JOIN equinos ON equinos.id = refeicoes.equino_id").
		Where("equinos.equinoid = ?", equinoid).
		Order("data_refeicao DESC").
		Limit(30).
		Find(&refeicoes).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_refeicoes", "erro ao buscar refeições", err)
	}
	return refeicoes, nil
}

func (r *repository) Create(ctx context.Context, plano *models.PlanoNutricional) error {
	if err := r.db.WithContext(ctx).Create(plano).Error; err != nil {
		return apperrors.NewDatabaseError("create_plano", "erro ao criar plano nutricional", err)
	}
	return nil
}

func (r *repository) CreateRefeicao(ctx context.Context, refeicao *models.Refeicao) error {
	if err := r.db.WithContext(ctx).Create(refeicao).Error; err != nil {
		return apperrors.NewDatabaseError("create_refeicao", "erro ao criar refeição", err)
	}
	return nil
}
