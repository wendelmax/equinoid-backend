package leiloes

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context, leiloeiroID *uint) ([]*models.Leilao, error)
	FindByID(ctx context.Context, id uint) (*models.Leilao, error)
	Create(ctx context.Context, leilao *models.Leilao) error
	Update(ctx context.Context, leilao *models.Leilao) error
	Delete(ctx context.Context, id uint) error
	
	FindParticipacoesByLeilaoID(ctx context.Context, leilaoID uint) ([]*models.ParticipacaoLeilao, error)
	FindParticipacaoByID(ctx context.Context, id uint) (*models.ParticipacaoLeilao, error)
	CreateParticipacao(ctx context.Context, participacao *models.ParticipacaoLeilao) error
	UpdateParticipacao(ctx context.Context, participacao *models.ParticipacaoLeilao) error
	DeleteParticipacao(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, leiloeiroID *uint) ([]*models.Leilao, error) {
	var leiloes []*models.Leilao
	query := r.db.WithContext(ctx).Preload("Leiloeiro")

	if leiloeiroID != nil {
		query = query.Where("leiloeiro_id = ?", *leiloeiroID)
	}

	if err := query.Order("data_inicio DESC").Find(&leiloes).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_leiloes", "erro ao buscar leilões", err)
	}
	return leiloes, nil
}

func (r *repository) FindByID(ctx context.Context, id uint) (*models.Leilao, error) {
	var leilao models.Leilao
	if err := r.db.WithContext(ctx).Preload("Leiloeiro").First(&leilao, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &apperrors.NotFoundError{Resource: "leilao", Message: "leilão não encontrado"}
		}
		return nil, apperrors.NewDatabaseError("find_leilao", "erro ao buscar leilão", err)
	}
	return &leilao, nil
}

func (r *repository) Create(ctx context.Context, leilao *models.Leilao) error {
	if err := r.db.WithContext(ctx).Create(leilao).Error; err != nil {
		return apperrors.NewDatabaseError("create_leilao", "erro ao criar leilão", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, leilao *models.Leilao) error {
	if err := r.db.WithContext(ctx).Save(leilao).Error; err != nil {
		return apperrors.NewDatabaseError("update_leilao", "erro ao atualizar leilão", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Leilao{}, id)
	if result.Error != nil {
		return apperrors.NewDatabaseError("delete_leilao", "erro ao deletar leilão", result.Error)
	}
	if result.RowsAffected == 0 {
		return &apperrors.NotFoundError{Resource: "leilao", Message: "leilão não encontrado"}
	}
	return nil
}

func (r *repository) FindParticipacoesByLeilaoID(ctx context.Context, leilaoID uint) ([]*models.ParticipacaoLeilao, error) {
	var participacoes []*models.ParticipacaoLeilao
	if err := r.db.WithContext(ctx).
		Preload("Equino").
		Preload("Criador").
		Preload("Comprador").
		Where("leilao_id = ?", leilaoID).
		Order("status ASC, valor_inicial DESC").
		Find(&participacoes).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_participacoes_leilao", "erro ao buscar participações", err)
	}
	return participacoes, nil
}

func (r *repository) FindParticipacaoByID(ctx context.Context, id uint) (*models.ParticipacaoLeilao, error) {
	var participacao models.ParticipacaoLeilao
	if err := r.db.WithContext(ctx).
		Preload("Leilao").
		Preload("Equino").
		Preload("Criador").
		Preload("Comprador").
		First(&participacao, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &apperrors.NotFoundError{Resource: "participacao_leilao", Message: "participação não encontrada"}
		}
		return nil, apperrors.NewDatabaseError("find_participacao_leilao", "erro ao buscar participação", err)
	}
	return &participacao, nil
}

func (r *repository) CreateParticipacao(ctx context.Context, participacao *models.ParticipacaoLeilao) error {
	if err := r.db.WithContext(ctx).Create(participacao).Error; err != nil {
		return apperrors.NewDatabaseError("create_participacao_leilao", "erro ao criar participação", err)
	}
	return nil
}

func (r *repository) UpdateParticipacao(ctx context.Context, participacao *models.ParticipacaoLeilao) error {
	if err := r.db.WithContext(ctx).Save(participacao).Error; err != nil {
		return apperrors.NewDatabaseError("update_participacao_leilao", "erro ao atualizar participação", err)
	}
	return nil
}

func (r *repository) DeleteParticipacao(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.ParticipacaoLeilao{}, id)
	if result.Error != nil {
		return apperrors.NewDatabaseError("delete_participacao_leilao", "erro ao deletar participação", result.Error)
	}
	if result.RowsAffected == 0 {
		return &apperrors.NotFoundError{Resource: "participacao_leilao", Message: "participação não encontrada"}
	}
	return nil
}
