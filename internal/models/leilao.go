package models

import (
	"time"

	"gorm.io/gorm"
)

// Leilao representa um leilão de equinos
type Leilao struct {
	ID                     uint           `json:"id" gorm:"primaryKey"`
	Nome                   string         `json:"nome" gorm:"not null;size:200"`
	Descricao              string         `json:"descricao" gorm:"type:text"`
	LeiloeiroID            uint           `json:"leiloeiro_id" gorm:"not null;index"`
	TaxaComissaoPercentual float64        `json:"taxa_comissao_percentual" gorm:"type:decimal(5,2);not null"`
	TaxaFixa               *float64       `json:"taxa_fixa" gorm:"type:decimal(15,2)"`
	DataInicio             time.Time      `json:"data_inicio" gorm:"not null"`
	DataFim                time.Time      `json:"data_fim" gorm:"not null"`
	Local                  string         `json:"local" gorm:"size:255"`
	TipoLeilao             TipoLeilao     `json:"tipo_leilao" gorm:"size:20;not null"`
	Status                 StatusLeilao   `json:"status" gorm:"size:20;default:'agendado'"`
	TotalArrecadado        *float64       `json:"total_arrecadado" gorm:"type:decimal(15,2)"`
	TotalComissoes         *float64       `json:"total_comissoes" gorm:"type:decimal(15,2)"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	Leiloeiro     *User                  `json:"leiloeiro,omitempty" gorm:"foreignKey:LeiloeiroID"`
	Participacoes []ParticipacaoLeilao   `json:"participacoes,omitempty" gorm:"foreignKey:LeilaoID"`
}

// TipoLeilao define os tipos de leilão
type TipoLeilao string

const (
	TipoLeilaoPresencial TipoLeilao = "presencial"
	TipoLeilaoOnline     TipoLeilao = "online"
	TipoLeilaoHibrido    TipoLeilao = "hibrido"
)

// StatusLeilao define os status possíveis de um leilão
type StatusLeilao string

const (
	StatusLeilaoAgendado    StatusLeilao = "agendado"
	StatusLeilaoEmAndamento StatusLeilao = "em_andamento"
	StatusLeilaoEncerrado   StatusLeilao = "encerrado"
	StatusLeilaoCancelado   StatusLeilao = "cancelado"
)

// ParticipacaoLeilao representa a participação de um equino em um leilão
type ParticipacaoLeilao struct {
	ID                  uint           `json:"id" gorm:"primaryKey"`
	LeilaoID            uint           `json:"leilao_id" gorm:"not null;index"`
	EquinoID            uint           `json:"equino_id" gorm:"not null;index"`
	CriadorID           uint           `json:"criador_id" gorm:"not null;index"`
	ValorInicial        float64        `json:"valor_inicial" gorm:"type:decimal(15,2);not null"`
	ValorReserva        *float64       `json:"valor_reserva" gorm:"type:decimal(15,2)"`
	ValorFinal          *float64       `json:"valor_final" gorm:"type:decimal(15,2)"`
	ValorVendido        *float64       `json:"valor_vendido" gorm:"type:decimal(15,2)"`
	CompradorID         *uint          `json:"comprador_id" gorm:"index"`
	Status              StatusParticipacaoLeilao `json:"status" gorm:"size:20;default:'inscrito'"`
	Particularidades    string         `json:"particularidades" gorm:"type:text"`
	ComissaoLeiloeiro   *float64       `json:"comissao_leiloeiro" gorm:"type:decimal(15,2)"`
	Compareceu          *bool          `json:"compareceu"`
	PenalizacaoAusencia *int           `json:"penalizacao_ausencia"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	Leilao    *Leilao `json:"leilao,omitempty" gorm:"foreignKey:LeilaoID"`
	Equino    *Equino `json:"equino,omitempty" gorm:"foreignKey:EquinoID"`
	Criador   *User   `json:"criador,omitempty" gorm:"foreignKey:CriadorID"`
	Comprador *User   `json:"comprador,omitempty" gorm:"foreignKey:CompradorID"`
}

// StatusParticipacaoLeilao define os status de participação em leilão
type StatusParticipacaoLeilao string

const (
	StatusParticipacaoInscrito   StatusParticipacaoLeilao = "inscrito"
	StatusParticipacaoAprovado   StatusParticipacaoLeilao = "aprovado"
	StatusParticipacaoVendido    StatusParticipacaoLeilao = "vendido"
	StatusParticipacaoNaoVendido StatusParticipacaoLeilao = "nao_vendido"
	StatusParticipacaoCancelado  StatusParticipacaoLeilao = "cancelado"
)

// CreateParticipacaoLeilaoRequest representa a requisição para criar participação
type CreateParticipacaoLeilaoRequest struct {
	Equinoid         string   `json:"equinoid" validate:"required"`
	ValorInicial     float64  `json:"valor_inicial" validate:"required,gt=0"`
	ValorReserva     *float64 `json:"valor_reserva" validate:"omitempty,gt=0"`
	Particularidades string   `json:"particularidades"`
}

// RegistrarVendaRequest representa a requisição para registrar venda
type RegistrarVendaRequest struct {
	ValorVendido float64 `json:"valor_vendido" validate:"required,gt=0"`
	CompradorID  uint    `json:"comprador_id" validate:"required"`
}

// ParticipacaoLeilaoResponse representa a resposta
type ParticipacaoLeilaoResponse struct {
	ID                  uint                     `json:"id"`
	LeilaoID            uint                     `json:"leilao_id"`
	LeilaoNome          string                   `json:"leilao_nome,omitempty"`
	EquinoID            uint                     `json:"equino_id"`
	Equinoid            string                   `json:"equinoid"`
	EquinoNome          string                   `json:"equino_nome"`
	CriadorID           uint                     `json:"criador_id"`
	CriadorNome         string                   `json:"criador_nome"`
	ValorInicial        float64                  `json:"valor_inicial"`
	ValorReserva        *float64                 `json:"valor_reserva"`
	ValorFinal          *float64                 `json:"valor_final"`
	ValorVendido        *float64                 `json:"valor_vendido"`
	CompradorID         *uint                    `json:"comprador_id"`
	CompradorNome       string                   `json:"comprador_nome,omitempty"`
	Status              StatusParticipacaoLeilao `json:"status"`
	Particularidades    string                   `json:"particularidades"`
	ComissaoLeiloeiro   *float64                 `json:"comissao_leiloeiro"`
	Compareceu          *bool                    `json:"compareceu"`
	PenalizacaoAusencia *int                     `json:"penalizacao_ausencia"`
	CreatedAt           time.Time                `json:"created_at"`
}

// ToResponse converte ParticipacaoLeilao para ParticipacaoLeilaoResponse
func (p *ParticipacaoLeilao) ToResponse() *ParticipacaoLeilaoResponse {
	response := &ParticipacaoLeilaoResponse{
		ID:                  p.ID,
		LeilaoID:            p.LeilaoID,
		EquinoID:            p.EquinoID,
		CriadorID:           p.CriadorID,
		ValorInicial:        p.ValorInicial,
		ValorReserva:        p.ValorReserva,
		ValorFinal:          p.ValorFinal,
		ValorVendido:        p.ValorVendido,
		CompradorID:         p.CompradorID,
		Status:              p.Status,
		Particularidades:    p.Particularidades,
		ComissaoLeiloeiro:   p.ComissaoLeiloeiro,
		Compareceu:          p.Compareceu,
		PenalizacaoAusencia: p.PenalizacaoAusencia,
		CreatedAt:           p.CreatedAt,
	}

	if p.Leilao != nil {
		response.LeilaoNome = p.Leilao.Nome
	}
	if p.Equino != nil {
		response.Equinoid = p.Equino.Equinoid
		response.EquinoNome = p.Equino.Nome
	}
	if p.Criador != nil {
		response.CriadorNome = p.Criador.Name
	}
	if p.Comprador != nil {
		response.CompradorNome = p.Comprador.Name
	}

	return response
}

// TableName especifica o nome da tabela
func (ParticipacaoLeilao) TableName() string {
	return "participacoes_leiloes"
}
