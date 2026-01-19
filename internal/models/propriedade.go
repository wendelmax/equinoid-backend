package models

import (
	"time"

	"gorm.io/gorm"
)

// TipoPropriedade define o tipo de propriedade
type TipoPropriedade string

const (
	TipoPropriedadeHaras  TipoPropriedade = "haras"
	TipoPropriedadeHipica TipoPropriedade = "hipica"
)

// Propriedade representa um haras ou hípica no sistema
type Propriedade struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	Nome          string          `json:"nome" gorm:"size:200;not null"`
	Tipo          TipoPropriedade `json:"tipo" gorm:"size:20;not null"`
	CNPJ          string          `json:"cnpj" gorm:"size:20"`
	Endereco      string          `json:"endereco" gorm:"size:255"`
	Cidade        string          `json:"cidade" gorm:"size:100"`
	Estado        string          `json:"estado" gorm:"size:50"`
	Pais          string          `json:"pais" gorm:"size:3;default:'BRA'"`
	CEP           string          `json:"cep" gorm:"size:20"`
	Telefone      string          `json:"telefone" gorm:"size:20"`
	Email         string          `json:"email" gorm:"size:100"`
	ResponsavelID uint            `json:"responsavel_id" gorm:"not null"`
	IsActive      bool            `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Responsavel *User    `json:"responsavel,omitempty" gorm:"foreignKey:ResponsavelID"`
	Equinos     []Equino `json:"equinos,omitempty" gorm:"foreignKey:PropriedadeID"`
}

// CreatePropriedadeRequest representa a requisição para criar uma propriedade
type CreatePropriedadeRequest struct {
	Nome          string          `json:"nome" validate:"required"`
	Tipo          TipoPropriedade `json:"tipo" validate:"required"`
	CNPJ          string          `json:"cnpj"`
	Endereco      string          `json:"endereco"`
	Cidade        string          `json:"cidade"`
	Estado        string          `json:"estado"`
	Pais          string          `json:"pais"`
	CEP           string          `json:"cep"`
	Telefone      string          `json:"telefone"`
	Email         string          `json:"email"`
	ResponsavelID uint            `json:"responsavel_id"`
}

// UpdatePropriedadeRequest representa a requisição para atualizar uma propriedade
type UpdatePropriedadeRequest struct {
	Nome     *string          `json:"nome"`
	Tipo     *TipoPropriedade `json:"tipo"`
	CNPJ     *string          `json:"cnpj"`
	Endereco *string          `json:"endereco"`
	Cidade   *string          `json:"cidade"`
	Estado   *string          `json:"estado"`
	Pais     *string          `json:"pais"`
	CEP      *string          `json:"cep"`
	Telefone *string          `json:"telefone"`
	Email    *string          `json:"email"`
	IsActive *bool            `json:"is_active"`
}

// PropriedadeResponse representa a resposta de uma propriedade
type PropriedadeResponse struct {
	ID               uint            `json:"id"`
	Nome             string          `json:"nome"`
	Tipo             TipoPropriedade `json:"tipo"`
	CNPJ             string          `json:"cnpj,omitempty"`
	Endereco         string          `json:"endereco,omitempty"`
	Cidade           string          `json:"cidade,omitempty"`
	Estado           string          `json:"estado,omitempty"`
	Pais             string          `json:"pais,omitempty"`
	CEP              string          `json:"cep,omitempty"`
	Telefone         string          `json:"telefone,omitempty"`
	Email            string          `json:"email,omitempty"`
	ResponsavelID    uint            `json:"responsavel_id"`
	Responsavel      *UserResponse   `json:"responsavel,omitempty"`
	TotalEquinos     int             `json:"total_equinos"`
	EquinosEmCuidado int             `json:"equinos_em_cuidado"`
	IsActive         bool            `json:"is_active"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// ToResponse converte Propriedade para PropriedadeResponse
func (p *Propriedade) ToResponse() *PropriedadeResponse {
	response := &PropriedadeResponse{
		ID:            p.ID,
		Nome:          p.Nome,
		Tipo:          p.Tipo,
		CNPJ:          p.CNPJ,
		Endereco:      p.Endereco,
		Cidade:        p.Cidade,
		Estado:        p.Estado,
		Pais:          p.Pais,
		CEP:           p.CEP,
		Telefone:      p.Telefone,
		Email:         p.Email,
		ResponsavelID: p.ResponsavelID,
		IsActive:      p.IsActive,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		TotalEquinos:  len(p.Equinos),
	}

	if p.Responsavel != nil {
		response.Responsavel = p.Responsavel.ToResponse()
	}

	return response
}
