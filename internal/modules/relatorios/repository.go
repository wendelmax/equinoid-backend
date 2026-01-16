package relatorios

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"gorm.io/gorm"
)

type Repository interface {
	GetDashboardStats(ctx context.Context) (*models.DashboardStats, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	if err := r.db.WithContext(ctx).Model(&models.Equino{}).Count(&stats.TotalEquinos).Error; err != nil {
		return nil, apperrors.NewDatabaseError("count_equinos", "erro ao contar equinos", err)
	}

	if err := r.db.WithContext(ctx).Model(&models.Equino{}).Where("status = ?", "ativo").Count(&stats.EquinosAtivos).Error; err != nil {
		return nil, apperrors.NewDatabaseError("count_equinos_ativos", "erro ao contar equinos ativos", err)
	}

	if err := r.db.WithContext(ctx).Model(&models.Gestacao{}).Where("status_gestacao = ?", "em_andamento").Count(&stats.GestacoesAtivas).Error; err != nil {
		return nil, apperrors.NewDatabaseError("count_gestacoes", "erro ao contar gestações", err)
	}

	if err := r.db.WithContext(ctx).Model(&models.Evento{}).Where("data_evento > NOW()").Count(&stats.EventosProximos).Error; err != nil {
		return nil, apperrors.NewDatabaseError("count_eventos", "erro ao contar eventos", err)
	}

	var valorizacaoMedia float64
	if err := r.db.WithContext(ctx).Model(&models.RegistroValorizacao{}).Select("COALESCE(AVG(pontos_valorizacao), 0)").Scan(&valorizacaoMedia).Error; err != nil {
		return nil, apperrors.NewDatabaseError("avg_valorizacao", "erro ao calcular valorização média", err)
	}
	stats.ValorizacaoMedia = int64(valorizacaoMedia)

	var totalLeiloes int64
	if err := r.db.WithContext(ctx).Model(&models.Leilao{}).Count(&totalLeiloes).Error; err == nil {
		stats.TotalLeiloes = totalLeiloes
	}

	var equinosTokenizados int64
	if err := r.db.WithContext(ctx).Model(&models.Tokenizacao{}).Where("status = ?", "ativo").Count(&equinosTokenizados).Error; err == nil {
		stats.EquinosTokenizados = equinosTokenizados
	}

	return stats, nil
}
