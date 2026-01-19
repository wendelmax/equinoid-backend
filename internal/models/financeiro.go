package models

import (
	"time"

	"gorm.io/gorm"
)

// TransacaoFinanceira representa uma transação financeira
type TransacaoFinanceira struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Tipo        TipoTransacao  `json:"tipo" gorm:"size:20;not null"`
	Categoria   string         `json:"categoria" gorm:"size:100;not null"`
	Descricao   string         `json:"descricao" gorm:"size:500;not null"`
	Valor       float64        `json:"valor" gorm:"type:decimal(15,2);not null"`
	Data        time.Time      `json:"data" gorm:"not null;index"`
	EquinoID    *uint          `json:"equino_id" gorm:"index"`
	Status      StatusPagamento `json:"status" gorm:"size:20;default:'pendente'"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	Equino *Equino `json:"equino,omitempty" gorm:"foreignKey:EquinoID"`
}

// TipoTransacao define tipos de transação
type TipoTransacao string

const (
	TipoReceita TipoTransacao = "receita"
	TipoDespesa TipoTransacao = "despesa"
)

// StatusPagamento define status de pagamento
type StatusPagamento string

const (
	StatusPagamentoPendente  StatusPagamento = "pendente"
	StatusPagamentoPago      StatusPagamento = "pago"
	StatusPagamentoCancelado StatusPagamento = "cancelado"
)

// FinanceiroStats representa estatísticas financeiras
type FinanceiroStats struct {
	Balance        string  `json:"balance"`
	Revenue        string  `json:"revenue"`
	Expenses       string  `json:"expenses"`
	Profit         string  `json:"profit"`
	RevenueChange  float64 `json:"revenue_change"`
	ExpensesChange float64 `json:"expenses_change"`
}

// MonthlyData representa dados mensais
type MonthlyData struct {
	Month   string  `json:"month"`
	Receita float64 `json:"receita"`
	Despesa float64 `json:"despesa"`
}

// ExpenseBreakdown representa breakdown de despesas
type ExpenseBreakdown struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Color string  `json:"color"`
}

// RecentTransaction representa transação recente
type RecentTransaction struct {
	ID     uint            `json:"id"`
	Type   TipoTransacao   `json:"type"`
	Title  string          `json:"title"`
	Date   string          `json:"date"`
	Amount float64         `json:"amount"`
	Status StatusPagamento `json:"status"`
}

// Vencimento representa um vencimento futuro
type Vencimento struct {
	ID            uint    `json:"id"`
	Titulo        string  `json:"titulo"`
	Valor         float64 `json:"valor"`
	Vencimento    string  `json:"vencimento"`
	DiasRestantes int     `json:"dias_restantes"`
}

// CreateTransacaoRequest representa requisição de criação
type CreateTransacaoRequest struct {
	Tipo      TipoTransacao `json:"tipo" validate:"required"`
	Categoria string        `json:"categoria" validate:"required"`
	Descricao string        `json:"descricao" validate:"required"`
	Valor     float64       `json:"valor" validate:"required,gt=0"`
	Data      string        `json:"data" validate:"required"`
	EquinoID  *uint         `json:"equino_id"`
}

// TableName especifica o nome da tabela
func (TransacaoFinanceira) TableName() string {
	return "transacoes_financeiras"
}
