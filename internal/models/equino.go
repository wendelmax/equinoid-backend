package models

import (
	"time"

	"gorm.io/gorm"
)

// ExternalID representa um identificador externo de outros sistemas
type ExternalID struct {
	Sistema string `json:"sistema" validate:"required"`
	ID      string `json:"id" validate:"required"`
}

// Equino representa um equino no sistema
type Equino struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Equinoid       string         `json:"equinoid" gorm:"uniqueIndex;size:24;not null;column:equinoid"`
	MicrochipID    string         `json:"microchip_id" gorm:"uniqueIndex;size:25;not null"`
	Nome           string         `json:"nome" gorm:"size:100;not null"`
	DataNascimento *time.Time     `json:"data_nascimento" gorm:"not null"`
	Sexo           SexoEquino     `json:"sexo" gorm:"not null"`
	Pelagem        string         `json:"pelagem" gorm:"size:50;not null"`
	Raca           string         `json:"raca" gorm:"size:100;not null"`
	PaisOrigem     string         `json:"pais_origem" gorm:"size:3;not null"`
	Genitora       string         `json:"genitora_equinoid" gorm:"size:25"`
	Genitor        string         `json:"genitor_equinoid" gorm:"size:25"`
	ProprietarioID uint           `json:"proprietario_id" gorm:"not null"`
	PropriedadeID  uint           `json:"propriedade_id"`
	ExternalIDs    JSONB          `json:"external_ids" gorm:"type:jsonb"`
	FotoPerfil     string         `json:"foto_perfil" gorm:"size:255"`
	FotosGaleria   JSONB          `json:"fotos_galeria" gorm:"type:jsonb"`
	Status         StatusEquino   `json:"status" gorm:"default:'ativo'"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Proprietario         *User                 `json:"proprietario,omitempty" gorm:"foreignKey:ProprietarioID"`
	Propriedade          *Propriedade          `json:"propriedade,omitempty" gorm:"foreignKey:PropriedadeID"`
	Eventos              []Evento              `json:"eventos,omitempty" gorm:"foreignKey:EquinoID"`
	VeterinariosNomados  []EquinoVeterinario   `json:"veterinarios,omitempty" gorm:"foreignKey:EquinoID"`
	RegistrosValorizacao []RegistroValorizacao `json:"registros_valorizacao,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
	RankingsValorizacao  []RankingValorizacao  `json:"rankings_valorizacao,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
	CoberturasComo       []Cobertura           `json:"coberturas_como_reprodutor,omitempty" gorm:"foreignKey:ReprodutorEquinoid;references:Equinoid"`
	CoberturasMatriz     []Cobertura           `json:"coberturas_como_matriz,omitempty" gorm:"foreignKey:MatrizEquinoid;references:Equinoid"`
	AvaliacoesSemen      []AvaliacaoSemen      `json:"avaliacoes_semen,omitempty" gorm:"foreignKey:ReprodutorEquinoid;references:Equinoid"`
	GestacoesComo        []Gestacao            `json:"gestacoes_como_matriz,omitempty" gorm:"foreignKey:MatrizEquinoid;references:Equinoid"`
	PerformancesMaternas []PerformanceMaterna  `json:"performances_maternas,omitempty" gorm:"foreignKey:MatrizEquinoid;references:Equinoid"`
	RankingsReprodutivos []RankingReprodutivo  `json:"rankings_reprodutivos,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
	PerfilSocial         *PerfilSocial         `json:"perfil_social,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
	PostsSociais         []PostSocial          `json:"posts_sociais,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
}

// SexoEquino define o sexo do equino
type SexoEquino string

const (
	SexoMacho SexoEquino = "macho"
	SexoFemea SexoEquino = "femea"
)

// StatusEquino define o status do equino
type StatusEquino string

const (
	StatusAtivo      StatusEquino = "ativo"
	StatusInativo    StatusEquino = "inativo"
	StatusVendido    StatusEquino = "vendido"
	StatusFalecido   StatusEquino = "falecido"
	StatusTransferir StatusEquino = "transferir"
)

// CreateEquinoRequest representa a requisição de criação de equino
// EquinoId será gerado automaticamente pelo sistema
type CreateEquinoRequest struct {
	MicrochipID      string       `json:"microchip_id" validate:"required"`
	Nome             string       `json:"nome" validate:"required"`
	DataNascimento   *time.Time   `json:"data_nascimento" validate:"required"`
	Sexo             SexoEquino   `json:"sexo" validate:"required"`
	Pelagem          string       `json:"pelagem" validate:"required"`
	Raca             string       `json:"raca" validate:"required"`
	PaisOrigem       string       `json:"pais_origem" validate:"required,len=3"`
	GenitoraEquinoid string       `json:"genitora_equinoid"`
	GenitorEquinoid  string       `json:"genitor_equinoid"`
	ProprietarioID   uint         `json:"proprietario_id" validate:"required"`
	PropriedadeID    uint         `json:"propriedade_id"`
	ExternalIDs      []ExternalID `json:"external_ids"`
	FotoPerfil       string       `json:"foto_perfil"`
}

// UpdateEquinoRequest representa a requisição de atualização de equino
type UpdateEquinoRequest struct {
	Nome          *string       `json:"nome"`
	Pelagem       *string       `json:"pelagem"`
	Raca          *string       `json:"raca"`
	PaisOrigem    *string       `json:"pais_origem" validate:"omitempty,len=3"`
	PropriedadeID *uint         `json:"propriedade_id"`
	ExternalIDs   []ExternalID  `json:"external_ids"`
	FotoPerfil    *string       `json:"foto_perfil"`
	FotosGaleria  *[]string     `json:"fotos_galeria"`
	Status        *StatusEquino `json:"status"`
}

// EquinoResponse representa a resposta de equino
type EquinoResponse struct {
	ID             uint                  `json:"id"`
	Equinoid       string                `json:"equinoid"`
	MicrochipID    string                `json:"microchip_id"`
	Nome           string                `json:"nome"`
	DataNascimento *time.Time            `json:"data_nascimento"`
	Sexo           SexoEquino            `json:"sexo"`
	Pelagem        string                `json:"pelagem"`
	Raca           string                `json:"raca"`
	PaisOrigem     string                `json:"pais_origem"`
	ExternalIDs    []ExternalID          `json:"external_ids,omitempty"`
	FotoPerfil     string                `json:"foto_perfil,omitempty"`
	FotosGaleria   []string              `json:"fotos_galeria,omitempty"`
	PropriedadeID  uint                  `json:"propriedade_id,omitempty"`
	Propriedade    *PropriedadeResponse  `json:"propriedade,omitempty"`
	Genitora       *EquinoSimpleResponse `json:"genitora,omitempty"`
	Genitor        *EquinoSimpleResponse `json:"genitor,omitempty"`
	Proprietario   *UserResponse         `json:"proprietario,omitempty"`
	Veterinarios   []EquinoVetResponse   `json:"veterinarios,omitempty"`
	Status         StatusEquino          `json:"status"`
	Eventos        []EventoResponse      `json:"eventos,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

// EquinoVeterinario representa a relação entre equino e veterinário
type EquinoVeterinario struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	EquinoID      uint      `json:"equino_id" gorm:"not null"`
	VeterinarioID uint      `json:"veterinario_id" gorm:"not null"`
	NomeadoPorID  uint      `json:"nomeado_por_id" gorm:"not null"`
	DataNomeacao  time.Time `json:"data_nomeacao" gorm:"not null"`
	IsPrincipal   bool      `json:"is_principal" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`

	// Relacionamentos
	Equino      *Equino `json:"equino,omitempty" gorm:"foreignKey:EquinoID"`
	Veterinario *User   `json:"veterinario,omitempty" gorm:"foreignKey:VeterinarioID"`
	NomeadoPor  *User   `json:"nomeado_por,omitempty" gorm:"foreignKey:NomeadoPorID"`
}

// EquinoVetResponse representa a resposta da relação equino-veterinário
type EquinoVetResponse struct {
	ID            uint          `json:"id"`
	EquinoID      uint          `json:"equino_id"`
	VeterinarioID uint          `json:"veterinario_id"`
	Veterinario   *UserResponse `json:"veterinario,omitempty"`
	NomeadoPorID  uint          `json:"nomeado_por_id"`
	NomeadoPor    *UserResponse `json:"nomeado_por,omitempty"`
	DataNomeacao  time.Time     `json:"data_nomeacao"`
	IsPrincipal   bool          `json:"is_principal"`
	CreatedAt     time.Time     `json:"created_at"`
}

// EquinoSimpleResponse representa uma resposta simplificada de equino
type EquinoSimpleResponse struct {
	Equinoid string `json:"equinoid"`
	Nome     string `json:"nome"`
}

// ToResponse converte Equino para EquinoResponse
func (e *Equino) ToResponse() *EquinoResponse {
	response := &EquinoResponse{
		ID:             e.ID,
		Equinoid:       e.Equinoid,
		MicrochipID:    e.MicrochipID,
		Nome:           e.Nome,
		DataNascimento: e.DataNascimento,
		Sexo:           e.Sexo,
		Pelagem:        e.Pelagem,
		PaisOrigem:     e.PaisOrigem,
		Raca:           e.Raca,
		Status:         e.Status,
		FotoPerfil:     e.FotoPerfil,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}

	if e.Proprietario != nil {
		response.Proprietario = e.Proprietario.ToResponse()
	}

	if e.Propriedade != nil {
		response.Propriedade = e.Propriedade.ToResponse()
	}

	if len(e.Eventos) > 0 {
		response.Eventos = make([]EventoResponse, len(e.Eventos))
		for i, evento := range e.Eventos {
			response.Eventos[i] = *evento.ToResponse()
		}
	}

	if len(e.VeterinariosNomados) > 0 {
		response.Veterinarios = make([]EquinoVetResponse, len(e.VeterinariosNomados))
		for i, v := range e.VeterinariosNomados {
			response.Veterinarios[i] = EquinoVetResponse{
				ID:            v.ID,
				EquinoID:      v.EquinoID,
				VeterinarioID: v.VeterinarioID,
				NomeadoPorID:  v.NomeadoPorID,
				DataNomeacao:  v.DataNomeacao,
				IsPrincipal:   v.IsPrincipal,
				CreatedAt:     v.CreatedAt,
			}
			if v.Veterinario != nil {
				response.Veterinarios[i].Veterinario = v.Veterinario.ToResponse()
			}
			if v.NomeadoPor != nil {
				response.Veterinarios[i].NomeadoPor = v.NomeadoPor.ToResponse()
			}
		}
	}

	if e.FotosGaleria != nil {
		response.FotosGaleria = make([]string, 0)
		// Simplesmente itera e adiciona ao slice.
		// Como as chaves no JSONB eram "0", "1", etc., podemos tentar ordenar se necessário.
		for _, v := range e.FotosGaleria {
			if str, ok := v.(string); ok {
				response.FotosGaleria = append(response.FotosGaleria, str)
			}
		}
	}

	return response
}

// ToSimpleResponse converte Equino para EquinoSimpleResponse
func (e *Equino) ToSimpleResponse() *EquinoSimpleResponse {
	return &EquinoSimpleResponse{
		Equinoid: e.Equinoid,
		Nome:     e.Nome,
	}
}

// BeforeCreate é executado antes de criar um equino
func (e *Equino) BeforeCreate(tx *gorm.DB) error {
	if e.Status == "" {
		e.Status = StatusAtivo
	}
	return nil
}

// IsMacho verifica se o equino é macho
func (e *Equino) IsMacho() bool {
	return e.Sexo == SexoMacho
}

// IsFemea verifica se o equino é fêmea
func (e *Equino) IsFemea() bool {
	return e.Sexo == SexoFemea
}

// IsAtivo verifica se o equino está ativo
func (e *Equino) IsAtivo() bool {
	return e.Status == StatusAtivo
}
