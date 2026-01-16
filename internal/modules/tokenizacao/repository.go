package tokenizacao

import (
	"context"
	"fmt"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.Tokenizacao, int64, error)
	FindByID(ctx context.Context, id uint) (*models.Tokenizacao, error)
	FindByEquinoID(ctx context.Context, equinoID uint) (*models.Tokenizacao, error)
	FindByEquinoid(ctx context.Context, equinoid string) (*models.Tokenizacao, error)
	Create(ctx context.Context, tokenizacao *models.Tokenizacao) error
	Update(ctx context.Context, tokenizacao *models.Tokenizacao) error
	Delete(ctx context.Context, id uint) error
	
	FindTransacoesByTokenizacaoID(ctx context.Context, tokenizacaoID uint) ([]*models.TransacaoToken, error)
	CreateTransacao(ctx context.Context, transacao *models.TransacaoToken) error
	
	FindParticipacoesByTokenizacaoID(ctx context.Context, tokenizacaoID uint) ([]*models.ParticipacaoToken, error)
	UpsertParticipacao(ctx context.Context, participacao *models.ParticipacaoToken) error
	
	CreateOferta(ctx context.Context, oferta *models.OfertaToken) error
	FindOfertasAtivasByTokenizacaoID(ctx context.Context, tokenizacaoID uint) ([]*models.OfertaToken, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.Tokenizacao, int64, error) {
	var tokenizacoes []*models.Tokenizacao
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Tokenizacao{}).
		Preload("Equino")

	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if rating, ok := filters["rating"].(string); ok && rating != "" {
		query = query.Where("rating_risco = ?", rating)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.NewDatabaseError("count_tokenizacoes", "erro ao contar tokenizações", err)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&tokenizacoes).Error; err != nil {
		return nil, 0, apperrors.NewDatabaseError("list_tokenizacoes", "erro ao listar tokenizações", err)
	}

	return tokenizacoes, total, nil
}

func (r *repository) FindByID(ctx context.Context, id uint) (*models.Tokenizacao, error) {
	var tokenizacao models.Tokenizacao
	if err := r.db.WithContext(ctx).Preload("Equino").First(&tokenizacao, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &apperrors.NotFoundError{Resource: "tokenizacao", Message: "tokenização não encontrada"}
		}
		return nil, apperrors.NewDatabaseError("find_tokenizacao", "erro ao buscar tokenização", err)
	}
	return &tokenizacao, nil
}

func (r *repository) FindByEquinoID(ctx context.Context, equinoID uint) (*models.Tokenizacao, error) {
	var tokenizacao models.Tokenizacao
	if err := r.db.WithContext(ctx).Preload("Equino").Where("equino_id = ?", equinoID).First(&tokenizacao).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &apperrors.NotFoundError{Resource: "tokenizacao", Message: "tokenização não encontrada para este equino"}
		}
		return nil, apperrors.NewDatabaseError("find_tokenizacao_by_equino", "erro ao buscar tokenização", err)
	}
	return &tokenizacao, nil
}

func (r *repository) FindByEquinoid(ctx context.Context, equinoid string) (*models.Tokenizacao, error) {
	var tokenizacao models.Tokenizacao
	if err := r.db.WithContext(ctx).
		Preload("Equino").
		Joins("JOIN equinos ON equinos.id = tokenizacoes.equino_id").
		Where("equinos.equinoid = ?", equinoid).
		First(&tokenizacao).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &apperrors.NotFoundError{Resource: "tokenizacao", Message: "tokenização não encontrada para este equino"}
		}
		return nil, apperrors.NewDatabaseError("find_tokenizacao_by_equinoid", "erro ao buscar tokenização", err)
	}
	return &tokenizacao, nil
}

func (r *repository) Create(ctx context.Context, tokenizacao *models.Tokenizacao) error {
	if err := r.db.WithContext(ctx).Create(tokenizacao).Error; err != nil {
		return apperrors.NewDatabaseError("create_tokenizacao", "erro ao criar tokenização", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, tokenizacao *models.Tokenizacao) error {
	if err := r.db.WithContext(ctx).Save(tokenizacao).Error; err != nil {
		return apperrors.NewDatabaseError("update_tokenizacao", "erro ao atualizar tokenização", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Tokenizacao{}, id)
	if result.Error != nil {
		return apperrors.NewDatabaseError("delete_tokenizacao", "erro ao deletar tokenização", result.Error)
	}
	if result.RowsAffected == 0 {
		return &apperrors.NotFoundError{Resource: "tokenizacao", Message: "tokenização não encontrada"}
	}
	return nil
}

func (r *repository) FindTransacoesByTokenizacaoID(ctx context.Context, tokenizacaoID uint) ([]*models.TransacaoToken, error) {
	var transacoes []*models.TransacaoToken
	if err := r.db.WithContext(ctx).
		Preload("Vendedor").
		Preload("Comprador").
		Where("tokenizacao_id = ?", tokenizacaoID).
		Order("data_transacao DESC").
		Find(&transacoes).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_transacoes", "erro ao buscar transações", err)
	}
	return transacoes, nil
}

func (r *repository) CreateTransacao(ctx context.Context, transacao *models.TransacaoToken) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var tokenizacao models.Tokenizacao
		if err := tx.First(&tokenizacao, transacao.TokenizacaoID).Error; err != nil {
			return &apperrors.NotFoundError{Resource: "tokenizacao", Message: "tokenização não encontrada"}
		}

		if tokenizacao.Status != models.StatusTokenAtivo {
			return &apperrors.ValidationError{Message: "tokenização não está ativa"}
		}

		if transacao.Quantidade > tokenizacao.TokensDisponiveisVenda {
			return &apperrors.ValidationError{
				Message: fmt.Sprintf("tokens insuficientes disponíveis (disponível: %d, solicitado: %d)",
					tokenizacao.TokensDisponiveisVenda, transacao.Quantidade),
			}
		}

		transacao.ValorTotal = float64(transacao.Quantidade) * transacao.PrecoUnitario
		transacao.Status = "confirmado"

		if err := tx.Create(transacao).Error; err != nil {
			return apperrors.NewDatabaseError("create_transacao", "erro ao criar transação", err)
		}

		tokenizacao.TokensVendidos += transacao.Quantidade
		tokenizacao.TokensDisponiveisVenda -= transacao.Quantidade

		if err := tx.Save(&tokenizacao).Error; err != nil {
			return apperrors.NewDatabaseError("update_tokenizacao_stock", "erro ao atualizar estoque de tokens", err)
		}

		return nil
	})
}

func (r *repository) FindParticipacoesByTokenizacaoID(ctx context.Context, tokenizacaoID uint) ([]*models.ParticipacaoToken, error) {
	var participacoes []*models.ParticipacaoToken
	if err := r.db.WithContext(ctx).
		Preload("Investidor").
		Where("tokenizacao_id = ?", tokenizacaoID).
		Order("quantidade_tokens DESC").
		Find(&participacoes).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_participacoes", "erro ao buscar participações", err)
	}
	return participacoes, nil
}

func (r *repository) UpsertParticipacao(ctx context.Context, participacao *models.ParticipacaoToken) error {
	var existing models.ParticipacaoToken
	err := r.db.WithContext(ctx).
		Where("tokenizacao_id = ? AND investidor_id = ?", participacao.TokenizacaoID, participacao.InvestidorID).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(participacao).Error
	} else if err != nil {
		return apperrors.NewDatabaseError("find_participacao", "erro ao buscar participação", err)
	}

	existing.QuantidadeTokens += participacao.QuantidadeTokens
	existing.ValorInvestido += participacao.ValorInvestido
	existing.PercentualTotal = float64(existing.QuantidadeTokens) / float64(100) * 100

	return r.db.WithContext(ctx).Save(&existing).Error
}

func (r *repository) CreateOferta(ctx context.Context, oferta *models.OfertaToken) error {
	if err := r.db.WithContext(ctx).Create(oferta).Error; err != nil {
		return apperrors.NewDatabaseError("create_oferta", "erro ao criar oferta", err)
	}
	return nil
}

func (r *repository) FindOfertasAtivasByTokenizacaoID(ctx context.Context, tokenizacaoID uint) ([]*models.OfertaToken, error) {
	var ofertas []*models.OfertaToken
	if err := r.db.WithContext(ctx).
		Preload("Vendedor").
		Where("tokenizacao_id = ? AND status = ? AND data_expiracao > NOW()", tokenizacaoID, "ativa").
		Order("preco_unitario ASC").
		Find(&ofertas).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_ofertas", "erro ao buscar ofertas", err)
	}
	return ofertas, nil
}
