package models

import (
	"time"

	"gorm.io/gorm"
)

type AtendimentoSanitario struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Equinoid      string         `json:"equinoid" gorm:"size:25;not null"`
	Tipo          string         `json:"tipo" gorm:"size:50"` // vacina, vermifugo, consulta, odonto
	Descricao     string         `json:"descricao" gorm:"size:255;not null"`
	Data          time.Time      `json:"data"`
	VeterinarioID uint           `json:"veterinario_id"`
	ProximaData   *time.Time     `json:"proxima_data,omitempty"`
	Custo         float64        `json:"custo" gorm:"type:decimal(15,2)"`
	Observacoes   string         `json:"observacoes" gorm:"type:text"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	Equino      *Equino `json:"equino,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
	Veterinario *User   `json:"veterinario,omitempty" gorm:"foreignKey:VeterinarioID"`
}

func (AtendimentoSanitario) TableName() string {
	return "sanitario.atendimentos"
}
