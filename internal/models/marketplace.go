package models

import (
	"time"

	"gorm.io/gorm"
)

type AnuncioMarketplace struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UsuarioID uint           `json:"usuario_id" gorm:"not null"`
	Tipo      string         `json:"tipo" gorm:"size:50;not null"` // 'animal', 'semen', 'embrio', 'equipamento'
	Titulo    string         `json:"titulo" gorm:"size:200;not null"`
	Descricao string         `json:"descricao" gorm:"type:text"`
	Preco     float64        `json:"preco" gorm:"type:decimal(15,2);not null"`
	Equinoid  *string        `json:"equinoid" gorm:"size:25"`
	Fotos     JSONB          `json:"fotos" gorm:"type:jsonb"`
	Status    string         `json:"status" gorm:"size:20;default:'ativo'"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Usuario *User   `json:"vendedor,omitempty" gorm:"foreignKey:UsuarioID"`
	Equino  *Equino `json:"equino,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
}

func (AnuncioMarketplace) TableName() string {
	return "marketplace.anuncios"
}
