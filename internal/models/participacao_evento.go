package models

import (
	"time"

	"gorm.io/gorm"
)

type ParticipacaoEvento struct {
	ID                  uint           `json:"id" gorm:"primaryKey"`
	EventoID            uint           `json:"evento_id" gorm:"not null;index"`
	EquinoID            uint           `json:"equino_id" gorm:"not null;index"`
	ParticipanteID      uint           `json:"participante_id" gorm:"not null;index"`
	Particularidades    string         `json:"particularidades" gorm:"type:text"`
	Resultado           string         `json:"resultado" gorm:"size:100"`
	Classificacao       *int           `json:"classificacao"`
	Compareceu          *bool          `json:"compareceu"`
	PenalizacaoAusencia *int           `json:"penalizacao_ausencia"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	Evento       *Evento         `json:"evento,omitempty" gorm:"foreignKey:EventoID"`
	Equino       *EquinoBasico   `json:"equino,omitempty" gorm:"foreignKey:EquinoID"`
	Participante *User           `json:"participante,omitempty" gorm:"foreignKey:ParticipanteID"`
}

type EquinoBasico struct {
	Equinoid string `json:"equinoid"`
	Nome     string `json:"nome"`
}

func (ParticipacaoEvento) TableName() string {
	return "participacoes_eventos"
}

type CreateParticipacaoEventoRequest struct {
	EventoID         uint   `json:"evento_id" binding:"required"`
	EquinoID         uint   `json:"equino_id" binding:"required"`
	Particularidades string `json:"particularidades"`
	Resultado        string `json:"resultado"`
	Classificacao    *int   `json:"classificacao"`
}

type UpdateParticipacaoEventoRequest struct {
	Particularidades *string `json:"particularidades"`
	Resultado        *string `json:"resultado"`
	Classificacao    *int    `json:"classificacao"`
}
