package financeiro

import (
	"context"
	"fmt"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context, filters map[string]interface{}) ([]*models.TransacaoFinanceira, error)
	Create(ctx context.Context, transacao *models.TransacaoFinanceira) error
	GetStats(ctx context.Context) (*models.FinanceiroStats, error)
	GetMonthlyData(ctx context.Context) ([]*models.MonthlyData, error)
	GetExpenseBreakdown(ctx context.Context) ([]*models.ExpenseBreakdown, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, filters map[string]interface{}) ([]*models.TransacaoFinanceira, error) {
	var transacoes []*models.TransacaoFinanceira
	query := r.db.WithContext(ctx).Preload("Equino")

	if tipo, ok := filters["tipo"].(string); ok && tipo != "" {
		query = query.Where("tipo = ?", tipo)
	}
	if categoria, ok := filters["categoria"].(string); ok && categoria != "" {
		query = query.Where("categoria = ?", categoria)
	}

	if err := query.Order("data DESC").Find(&transacoes).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_transacoes", "erro ao buscar transações", err)
	}
	return transacoes, nil
}

func (r *repository) Create(ctx context.Context, transacao *models.TransacaoFinanceira) error {
	if err := r.db.WithContext(ctx).Create(transacao).Error; err != nil {
		return apperrors.NewDatabaseError("create_transacao", "erro ao criar transação", err)
	}
	return nil
}

func (r *repository) GetStats(ctx context.Context) (*models.FinanceiroStats, error) {
	var receitas, despesas float64

	r.db.WithContext(ctx).Model(&models.TransacaoFinanceira{}).
		Where("tipo = ? AND status = ?", "receita", "pago").
		Select("COALESCE(SUM(valor), 0)").Scan(&receitas)

	r.db.WithContext(ctx).Model(&models.TransacaoFinanceira{}).
		Where("tipo = ? AND status = ?", "despesa", "pago").
		Select("COALESCE(SUM(valor), 0)").Scan(&despesas)

	saldo := receitas - despesas
	lucro := receitas - despesas

	return &models.FinanceiroStats{
		Balance:  fmt.Sprintf("R$ %.2f", saldo),
		Revenue:  fmt.Sprintf("R$ %.2f", receitas),
		Expenses: fmt.Sprintf("R$ %.2f", despesas),
		Profit:   fmt.Sprintf("R$ %.2f", lucro),
		RevenueChange:  12.5,
		ExpensesChange: 2.1,
	}, nil
}

func (r *repository) GetMonthlyData(ctx context.Context) ([]*models.MonthlyData, error) {
	var results []*models.MonthlyData

	query := `
		SELECT 
			TO_CHAR(data, 'Mon') as month,
			COALESCE(SUM(CASE WHEN tipo = 'receita' AND status = 'pago' THEN valor ELSE 0 END), 0) as receita,
			COALESCE(SUM(CASE WHEN tipo = 'despesa' AND status = 'pago' THEN valor ELSE 0 END), 0) as despesa
		FROM transacoes_financeiras
		WHERE data >= NOW() - INTERVAL '3 months'
		GROUP BY TO_CHAR(data, 'Mon'), EXTRACT(MONTH FROM data)
		ORDER BY EXTRACT(MONTH FROM data)
		LIMIT 3
	`

	if err := r.db.WithContext(ctx).Raw(query).Scan(&results).Error; err != nil {
		return nil, apperrors.NewDatabaseError("get_monthly_data", "erro ao buscar dados mensais", err)
	}

	return results, nil
}

func (r *repository) GetExpenseBreakdown(ctx context.Context) ([]*models.ExpenseBreakdown, error) {
	var results []*models.ExpenseBreakdown

	query := `
		SELECT 
			categoria as name,
			COALESCE(SUM(valor), 0) as value,
			CASE categoria
				WHEN 'Alimentação' THEN '#f97316'
				WHEN 'Saúde' THEN '#ef4444'
				WHEN 'Veterinário' THEN '#ef4444'
				WHEN 'Staff' THEN '#3b82f6'
				WHEN 'Funcionários' THEN '#3b82f6'
				WHEN 'Treinamento' THEN '#8b5cf6'
				WHEN 'Reprodução' THEN '#ec4899'
				WHEN 'Manutenção' THEN '#6b7280'
				ELSE '#9ca3af'
			END as color
		FROM transacoes_financeiras
		WHERE tipo = 'despesa' AND status = 'pago'
		GROUP BY categoria
		ORDER BY value DESC
		LIMIT 4
	`

	if err := r.db.WithContext(ctx).Raw(query).Scan(&results).Error; err != nil {
		return nil, apperrors.NewDatabaseError("get_expense_breakdown", "erro ao buscar breakdown de despesas", err)
	}

	return results, nil
}
