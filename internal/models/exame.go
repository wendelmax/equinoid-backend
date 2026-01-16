package models

import (
	"time"

	"gorm.io/gorm"
)

// ResultadoExame define os resultados possíveis
type ResultadoExame string

const (
	ResultadoNormal       ResultadoExame = "normal"
	ResultadoAlterado     ResultadoExame = "alterado"
	ResultadoPositivo     ResultadoExame = "positivo"
	ResultadoNegativo     ResultadoExame = "negativo"
	ResultadoInconclusivo ResultadoExame = "inconclusivo"
)

// CreateExameRequest representa requisição de criação
type CreateExameRequest struct {
	Equinoid                 string  `json:"equinoid" validate:"required"`
	TipoExame                string  `json:"tipo_exame" validate:"required"`
	NomeExame                string  `json:"nome_exame" validate:"required"`
	Descricao                string  `json:"descricao"`
	VeterinarioSolicitanteID uint    `json:"veterinario_solicitante_id" validate:"required"`
	LaboratorioID            *uint   `json:"laboratorio_id"`
	Observacoes              *string `json:"observacoes"`
}

// UpdateExameRequest representa requisição de atualização
type UpdateExameRequest struct {
	Status                 *string                `json:"status"`
	DataColeta             *time.Time             `json:"data_coleta"`
	DataRecebimentoAmostra *time.Time             `json:"data_recebimento_amostra"`
	DataInicioAnalise      *time.Time             `json:"data_inicio_analise"`
	DataConclusao          *time.Time             `json:"data_conclusao"`
	Resultado              *ResultadoExame        `json:"resultado"`
	Valores                map[string]interface{} `json:"valores"`
	Laudo                  *string                `json:"laudo"`
	Observacoes            *string                `json:"observacoes"`
}

// ExameLaboratorial representa um exame laboratorial
type ExameLaboratorial struct {
	ID                       uint           `json:"id" gorm:"primaryKey"`
	Equinoid                 string         `json:"equinoid" gorm:"size:25;not null;index"`
	TipoExame                string         `json:"tipo_exame" gorm:"size:50;not null"`
	NomeExame                string         `json:"nome_exame" gorm:"size:200;not null"`
	Descricao                string         `json:"descricao" gorm:"type:text"`
	VeterinarioSolicitanteID uint           `json:"veterinario_solicitante_id" gorm:"not null;index"`
	LaboratorioID            *uint          `json:"laboratorio_id" gorm:"index"`
	Status                   string         `json:"status" gorm:"size:50;default:'solicitado'"`
	DataSolicitacao          time.Time      `json:"data_solicitacao"`
	DataColeta               *time.Time     `json:"data_coleta"`
	DataRecebimentoAmostra   *time.Time     `json:"data_recebimento_amostra"`
	DataInicioAnalise        *time.Time     `json:"data_inicio_analise"`
	DataConclusao            *time.Time     `json:"data_conclusao"`
	Resultado                *ResultadoExame `json:"resultado"`
	Valores                  JSONB          `json:"valores" gorm:"type:jsonb"`
	Laudo                    *string        `json:"laudo" gorm:"type:text"`
	Observacoes              *string        `json:"observacoes" gorm:"type:text"`
	CertificadoID            *uint          `json:"certificado_id"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Equino                  *Equino `json:"equino,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
	VeterinarioSolicitante  *User   `json:"veterinario_solicitante,omitempty" gorm:"foreignKey:VeterinarioSolicitanteID"`
	Laboratorio             *User   `json:"laboratorio,omitempty" gorm:"foreignKey:LaboratorioID"`
	Certificado             *Certificate `json:"certificado,omitempty" gorm:"foreignKey:CertificadoID"`
}

// TableName especifica o nome da tabela
func (ExameLaboratorial) TableName() string {
	return "exames_laboratoriais"
}
