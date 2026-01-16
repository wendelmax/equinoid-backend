package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Evento representa um evento na vida de um equino
type Evento struct {
	ID                    uint           `json:"id" gorm:"primaryKey"`
	EquinoID              uint           `json:"equino_id" gorm:"not null"`
	TipoEvento            TipoEvento     `json:"tipo_evento" gorm:"not null"`
	Categoria             string         `json:"categoria" gorm:"size:50"`
	TipoEventoCompetitivo string         `json:"tipo_evento_competitivo" gorm:"size:50"`
	TipoEventoPublico     string         `json:"tipo_evento_publico" gorm:"size:50"`
	NomeEvento            string         `json:"nome_evento" gorm:"size:200"`
	Descricao             string         `json:"descricao" gorm:"type:text"`
	DataEvento            time.Time      `json:"data_evento" gorm:"not null"`
	Local                 string         `json:"local" gorm:"size:255"`
	Organizador           string         `json:"organizador" gorm:"size:200"`
	VeterinarioID         *uint          `json:"veterinario_id"`
	Documentos            JSONB          `json:"documentos" gorm:"type:jsonb"`
	Resultados            string         `json:"resultados" gorm:"type:text"`
	Participante          bool           `json:"participante" gorm:"default:true"`
	Particularidades      string         `json:"particularidades" gorm:"type:text"`
	ValorInscricao        *float64       `json:"valor_inscricao" gorm:"type:decimal(15,2)"`
	AceitaPatrocinio      bool           `json:"aceita_patrocinio" gorm:"default:false"`
	InformacoesPatrocinio string         `json:"informacoes_patrocinio" gorm:"type:text"`
	AssinaturaDigital     string         `json:"assinatura_digital,omitempty" gorm:"type:text"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	Equino      *Equino `json:"equino,omitempty" gorm:"foreignKey:EquinoID"`
	Veterinario *User   `json:"veterinario,omitempty" gorm:"foreignKey:VeterinarioID"`
}

// TipoEvento define os tipos de evento
type TipoEvento string

const (
	TipoEventoNascimento    TipoEvento = "nascimento"
	TipoEventoVacina        TipoEvento = "vacina"
	TipoEventoExame         TipoEvento = "exame"
	TipoEventoCirurgia      TipoEvento = "cirurgia"
	TipoEventoTransferencia TipoEvento = "transferencia"
	TipoEventoObito         TipoEvento = "obito"
	TipoEventoCompeticao    TipoEvento = "competicao"
	TipoEventoReproducao    TipoEvento = "reproducao"
	TipoEventoTreinamento   TipoEvento = "treinamento"
	TipoEventoViagem        TipoEvento = "viagem"
	TipoEventoCertificacao  TipoEvento = "certificacao"
)

// JSONB representa um campo JSONB do PostgreSQL
type JSONB map[string]interface{}

// Value implementa driver.Valuer para JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implementa sql.Scanner para JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return errors.New("cannot scan non-string into JSONB")
	}
}

// CreateEventoRequest representa a requisição de criação de evento
type CreateEventoRequest struct {
	TipoEvento            TipoEvento        `json:"tipo_evento" validate:"required"`
	Categoria             string            `json:"categoria"`
	TipoEventoCompetitivo string            `json:"tipo_evento_competitivo"`
	TipoEventoPublico     string            `json:"tipo_evento_publico"`
	NomeEvento            string            `json:"nome_evento"`
	Descricao             string            `json:"descricao" validate:"required"`
	DataEvento            time.Time         `json:"data_evento" validate:"required"`
	Local                 string            `json:"local"`
	Organizador           string            `json:"organizador"`
	VeterinarioID         *uint             `json:"veterinario_id"`
	Documentos            []DocumentoEvento `json:"documentos"`
	Resultados            string            `json:"resultados"`
	Participante          bool              `json:"participante"`
	Particularidades      string            `json:"particularidades"`
	ValorInscricao        *float64          `json:"valor_inscricao"`
	AceitaPatrocinio      bool              `json:"aceita_patrocinio"`
	InformacoesPatrocinio string            `json:"informacoes_patrocinio"`
}

// DocumentoEvento representa um documento anexo ao evento
type DocumentoEvento struct {
	Nome     string `json:"nome" validate:"required"`
	URL      string `json:"url,omitempty"`
	Conteudo string `json:"conteudo,omitempty"` // Base64 encoded content
	Tipo     string `json:"tipo" validate:"required"`
	Tamanho  int64  `json:"tamanho"`
}

// EventoResponse representa a resposta de evento
type EventoResponse struct {
	ID                    uint                       `json:"id"`
	TipoEvento            TipoEvento                 `json:"tipo_evento"`
	Categoria             string                     `json:"categoria,omitempty"`
	TipoEventoCompetitivo string                     `json:"tipo_evento_competitivo,omitempty"`
	TipoEventoPublico     string                     `json:"tipo_evento_publico,omitempty"`
	NomeEvento            string                     `json:"nome_evento,omitempty"`
	Descricao             string                     `json:"descricao"`
	DataEvento            time.Time                  `json:"data_evento"`
	Local                 string                     `json:"local,omitempty"`
	Organizador           string                     `json:"organizador,omitempty"`
	Veterinario           *VeterinarioSimpleResponse `json:"veterinario,omitempty"`
	Documentos            []DocumentoEvento          `json:"documentos,omitempty"`
	Resultados            string                     `json:"resultados,omitempty"`
	Participante          bool                       `json:"participante"`
	Particularidades      string                     `json:"particularidades,omitempty"`
	ValorInscricao        *float64                   `json:"valor_inscricao,omitempty"`
	AceitaPatrocinio      bool                       `json:"aceita_patrocinio"`
	InformacoesPatrocinio string                     `json:"informacoes_patrocinio,omitempty"`
	AssinaturaDigital     string                     `json:"assinatura_digital,omitempty"`
	CreatedAt             time.Time                  `json:"created_at"`
}

// VeterinarioSimpleResponse representa uma resposta simplificada de veterinário
type VeterinarioSimpleResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// ToResponse converte Evento para EventoResponse
func (e *Evento) ToResponse() *EventoResponse {
	response := &EventoResponse{
		ID:                    e.ID,
		TipoEvento:            e.TipoEvento,
		Categoria:             e.Categoria,
		TipoEventoCompetitivo: e.TipoEventoCompetitivo,
		TipoEventoPublico:     e.TipoEventoPublico,
		NomeEvento:            e.NomeEvento,
		Descricao:             e.Descricao,
		DataEvento:            e.DataEvento,
		Local:                 e.Local,
		Organizador:           e.Organizador,
		Resultados:            e.Resultados,
		Participante:          e.Participante,
		Particularidades:      e.Particularidades,
		ValorInscricao:        e.ValorInscricao,
		AceitaPatrocinio:      e.AceitaPatrocinio,
		InformacoesPatrocinio: e.InformacoesPatrocinio,
		AssinaturaDigital:     e.AssinaturaDigital,
		CreatedAt:             e.CreatedAt,
	}

	if e.Veterinario != nil {
		response.Veterinario = &VeterinarioSimpleResponse{
			ID:   e.Veterinario.ID,
			Name: e.Veterinario.Name,
		}
	}

	// Converter JSONB para []DocumentoEvento
	if e.Documentos != nil {
		if docs, ok := e.Documentos["documentos"].([]interface{}); ok {
			response.Documentos = make([]DocumentoEvento, len(docs))
			for i, doc := range docs {
				if docMap, ok := doc.(map[string]interface{}); ok {
					response.Documentos[i] = DocumentoEvento{
						Nome: getStringFromMap(docMap, "nome"),
						URL:  getStringFromMap(docMap, "url"),
						Tipo: getStringFromMap(docMap, "tipo"),
					}
					if tamanho, ok := docMap["tamanho"].(float64); ok {
						response.Documentos[i].Tamanho = int64(tamanho)
					}
				}
			}
		}
	}

	return response
}

// getStringFromMap extrai string de um map[string]interface{}
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// IsEventoMedico verifica se o evento é médico
func (e *Evento) IsEventoMedico() bool {
	eventosmedicos := []TipoEvento{
		TipoEventoVacina,
		TipoEventoExame,
		TipoEventoCirurgia,
		TipoEventoObito,
	}

	for _, tipo := range eventosmedicos {
		if e.TipoEvento == tipo {
			return true
		}
	}

	return false
}

// RequireVeterinario verifica se o evento requer veterinário
func (e *Evento) RequireVeterinario() bool {
	return e.IsEventoMedico()
}
