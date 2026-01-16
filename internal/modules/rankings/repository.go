package rankings

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"gorm.io/gorm"
)

type Repository interface {
	GetRankingGeral(ctx context.Context, tipo string, limit int) ([]*models.RankingItem, error)
	GetRankingsEquino(ctx context.Context, equinoid string) ([]*models.RankingItem, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetRankingGeral(ctx context.Context, tipo string, limit int) ([]*models.RankingItem, error) {
	var items []*models.RankingItem
	
	query := `
		SELECT 
			e.equinoid,
			e.nome,
			ROW_NUMBER() OVER (ORDER BY COALESCE(SUM(v.pontos_valorizacao), 0) DESC) as posicao,
			COALESCE(SUM(v.pontos_valorizacao), 0) as pontuacao_total,
			CASE 
				WHEN COALESCE(SUM(v.pontos_valorizacao), 0) >= 2000 THEN 'excepcional'
				WHEN COALESCE(SUM(v.pontos_valorizacao), 0) >= 1000 THEN 'alto'
				WHEN COALESCE(SUM(v.pontos_valorizacao), 0) >= 500 THEN 'medio'
				ELSE 'basico'
			END as nivel_valorizacao
		FROM equinos e
		LEFT JOIN registros_valorizacao v ON e.id = v.equino_id
		WHERE e.deleted_at IS NULL
		GROUP BY e.id, e.equinoid, e.nome
		ORDER BY pontuacao_total DESC
		LIMIT ?
	`
	
	if err := r.db.WithContext(ctx).Raw(query, limit).Scan(&items).Error; err != nil {
		return nil, apperrors.NewDatabaseError("get_ranking", "erro ao buscar ranking", err)
	}
	
	return items, nil
}

func (r *repository) GetRankingsEquino(ctx context.Context, equinoid string) ([]*models.RankingItem, error) {
	var items []*models.RankingItem
	
	query := `
		SELECT 
			? as equinoid,
			e.nome,
			1 as posicao,
			COALESCE(SUM(v.pontos_valorizacao), 0) as pontuacao_total,
			CASE 
				WHEN COALESCE(SUM(v.pontos_valorizacao), 0) >= 2000 THEN 'excepcional'
				WHEN COALESCE(SUM(v.pontos_valorizacao), 0) >= 1000 THEN 'alto'
				WHEN COALESCE(SUM(v.pontos_valorizacao), 0) >= 500 THEN 'medio'
				ELSE 'basico'
			END as nivel_valorizacao
		FROM equinos e
		LEFT JOIN registros_valorizacao v ON e.id = v.equino_id
		WHERE e.equinoid = ? AND e.deleted_at IS NULL
		GROUP BY e.id, e.equinoid, e.nome
	`
	
	if err := r.db.WithContext(ctx).Raw(query, equinoid, equinoid).Scan(&items).Error; err != nil {
		return nil, apperrors.NewDatabaseError("get_rankings_equino", "erro ao buscar rankings do equino", err)
	}
	
	return items, nil
}
