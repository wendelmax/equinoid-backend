package treinamento

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"gorm.io/gorm"
)

type Repository interface {
	FindSessoesByEquinoid(ctx context.Context, equinoid string) ([]*models.SessaoTreinamento, error)
	FindProgramasByEquinoid(ctx context.Context, equinoid string) ([]*models.ProgramaTreinamento, error)
	CreateSessao(ctx context.Context, sessao *models.SessaoTreinamento) error
	CreatePrograma(ctx context.Context, programa *models.ProgramaTreinamento) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindSessoesByEquinoid(ctx context.Context, equinoid string) ([]*models.SessaoTreinamento, error) {
	var sessoes []*models.SessaoTreinamento
	if err := r.db.WithContext(ctx).
		Preload("Equino").
		Preload("Treinador").
		Preload("ProgramaTreinamento").
		Joins("JOIN equinos ON equinos.id = sessoes_treinamento.equino_id").
		Where("equinos.equinoid = ?", equinoid).
		Order("data_sessao DESC").
		Limit(50).
		Find(&sessoes).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_sessoes", "erro ao buscar sessões", err)
	}
	return sessoes, nil
}

func (r *repository) FindProgramasByEquinoid(ctx context.Context, equinoid string) ([]*models.ProgramaTreinamento, error) {
	var programas []*models.ProgramaTreinamento
	if err := r.db.WithContext(ctx).
		Preload("Equino").
		Preload("Treinador").
		Joins("JOIN equinos ON equinos.id = programas_treinamento.equino_id").
		Where("equinos.equinoid = ?", equinoid).
		Order("data_inicio DESC").
		Find(&programas).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_programas", "erro ao buscar programas", err)
	}
	return programas, nil
}

func (r *repository) CreateSessao(ctx context.Context, sessao *models.SessaoTreinamento) error {
	if err := r.db.WithContext(ctx).Create(sessao).Error; err != nil {
		return apperrors.NewDatabaseError("create_sessao", "erro ao criar sessão de treinamento", err)
	}
	return nil
}

func (r *repository) CreatePrograma(ctx context.Context, programa *models.ProgramaTreinamento) error {
	if err := r.db.WithContext(ctx).Create(programa).Error; err != nil {
		return apperrors.NewDatabaseError("create_programa", "erro ao criar programa de treinamento", err)
	}
	return nil
}
