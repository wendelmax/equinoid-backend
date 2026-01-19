package models

import (
	"time"

	"gorm.io/gorm"
)

// ProgramaTreinamento representa um programa de treinamento
type ProgramaTreinamento struct {
	ID                 uint              `json:"id" gorm:"primaryKey"`
	EquinoID           uint              `json:"equino_id" gorm:"not null;index"`
	TreinadorID        uint              `json:"treinador_id" gorm:"not null;index"`
	NomePrograma       string            `json:"nome_programa" gorm:"size:200;not null"`
	Objetivo           string            `json:"objetivo" gorm:"size:200"`
	TipoPrograma       TipoPrograma      `json:"tipo_programa" gorm:"size:50;not null"`
	Intensidade        Intensidade       `json:"intensidade" gorm:"size:20;not null"`
	DuracaoSemanas     int               `json:"duracao_semanas" gorm:"not null"`
	FrequenciaSemanal  int               `json:"frequencia_semanal" gorm:"not null"`
	DuracaoSessaoMin   int               `json:"duracao_sessao_min"`
	Modalidades        JSONB             `json:"modalidades" gorm:"type:jsonb"`
	Observacoes        string            `json:"observacoes" gorm:"type:text"`
	DataInicio         time.Time         `json:"data_inicio"`
	DataFim            *time.Time        `json:"data_fim"`
	Status             StatusPrograma    `json:"status" gorm:"size:20;default:'ativo'"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
	DeletedAt          gorm.DeletedAt    `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	Equino   *Equino             `json:"equino,omitempty" gorm:"foreignKey:EquinoID"`
	Treinador *User              `json:"treinador,omitempty" gorm:"foreignKey:TreinadorID"`
	Sessoes  []SessaoTreinamento `json:"sessoes,omitempty" gorm:"foreignKey:ProgramaTreinamentoID"`
}

// TipoPrograma define os tipos de programa
type TipoPrograma string

const (
	TipoProgramaBasico      TipoPrograma = "basico"
	TipoProgramaIntermediario TipoPrograma = "intermediario"
	TipoProgramaAvancado    TipoPrograma = "avancado"
	TipoProgramaCompeticao  TipoPrograma = "competicao"
	TipoProgramaReabilitacao TipoPrograma = "reabilitacao"
)

// Intensidade define a intensidade
type Intensidade string

const (
	IntensidadeBaixa Intensidade = "baixa"
	IntensidadeMedia Intensidade = "media"
	IntensidadeAlta  Intensidade = "alta"
)

// StatusPrograma define o status do programa
type StatusPrograma string

const (
	StatusProgramaAtivo      StatusPrograma = "ativo"
	StatusProgramaPausado    StatusPrograma = "pausado"
	StatusProgramaFinalizado StatusPrograma = "finalizado"
	StatusProgramaCancelado  StatusPrograma = "cancelado"
)

// SessaoTreinamento representa uma sessão de treino
type SessaoTreinamento struct {
	ID                     uint           `json:"id" gorm:"primaryKey"`
	ProgramaTreinamentoID  *uint          `json:"programa_treinamento_id" gorm:"index"`
	EquinoID               uint           `json:"equino_id" gorm:"not null;index"`
	TreinadorID            uint           `json:"treinador_id" gorm:"not null;index"`
	DataSessao             time.Time      `json:"data_sessao" gorm:"not null;index"`
	Modalidade             string         `json:"modalidade" gorm:"size:100"`
	DuracaoMinutos         int            `json:"duracao_minutos"`
	Intensidade            Intensidade    `json:"intensidade" gorm:"size:20"`
	Distancia              *float64       `json:"distancia" gorm:"type:decimal(8,2)"`
	VelocidadeMedia        *float64       `json:"velocidade_media" gorm:"type:decimal(6,2)"`
	FrequenciaCardiacaMedia *int          `json:"frequencia_cardiaca_media"`
	CaloriasGastas         *int           `json:"calorias_gastas"`
	ExerciciosRealizados   JSONB          `json:"exercicios_realizados" gorm:"type:jsonb"`
	DesempenhoGeral        *int           `json:"desempenho_geral" gorm:"check:desempenho_geral >= 1 AND desempenho_geral <= 5"`
	Observacoes            string         `json:"observacoes" gorm:"type:text"`
	CondicoesClimaticas    string         `json:"condicoes_climaticas" gorm:"size:100"`
	TemperaturaC           *float64       `json:"temperatura_c" gorm:"type:decimal(5,2)"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	ProgramaTreinamento *ProgramaTreinamento `json:"programa_treinamento,omitempty" gorm:"foreignKey:ProgramaTreinamentoID"`
	Equino              *Equino              `json:"equino,omitempty" gorm:"foreignKey:EquinoID"`
	Treinador           *User                `json:"treinador,omitempty" gorm:"foreignKey:TreinadorID"`
}

// --- REQUEST/RESPONSE MODELS ---

// CreateProgramaTreinamentoRequest representa requisição de criação
type CreateProgramaTreinamentoRequest struct {
	Equinoid          string       `json:"equinoid" validate:"required"`
	NomePrograma      string       `json:"nome_programa" validate:"required"`
	Objetivo          string       `json:"objetivo" validate:"required"`
	TipoPrograma      TipoPrograma `json:"tipo_programa" validate:"required"`
	Intensidade       Intensidade  `json:"intensidade" validate:"required"`
	DuracaoSemanas    int          `json:"duracao_semanas" validate:"required,min=1"`
	FrequenciaSemanal int          `json:"frequencia_semanal" validate:"required,min=1,max=7"`
	DuracaoSessaoMin  int          `json:"duracao_sessao_min" validate:"required,min=15"`
	Modalidades       []string     `json:"modalidades" validate:"required"`
	Observacoes       string       `json:"observacoes"`
}

// CreateSessaoTreinamentoRequest representa requisição de criação
type CreateSessaoTreinamentoRequest struct {
	Equinoid                string       `json:"equinoid" validate:"required"`
	ProgramaTreinamentoID   *uint        `json:"programa_treinamento_id"`
	DataSessao              *time.Time   `json:"data_sessao"`
	Modalidade              string       `json:"modalidade" validate:"required"`
	DuracaoMinutos          int          `json:"duracao_minutos" validate:"required,min=5"`
	Intensidade             Intensidade  `json:"intensidade" validate:"required"`
	Distancia               *float64     `json:"distancia"`
	VelocidadeMedia         *float64     `json:"velocidade_media"`
	FrequenciaCardiacaMedia *int         `json:"frequencia_cardiaca_media"`
	CaloriasGastas          *int         `json:"calorias_gastas"`
	ExerciciosRealizados    []ExercicioRealizado `json:"exercicios_realizados"`
	DesempenhoGeral         *int         `json:"desempenho_geral" validate:"omitempty,min=1,max=5"`
	Observacoes             string       `json:"observacoes"`
	CondicoesClimaticas     string       `json:"condicoes_climaticas"`
	TemperaturaC            *float64     `json:"temperatura_c"`
}

// ExercicioRealizado representa um exercício em uma sessão
type ExercicioRealizado struct {
	Nome       string `json:"nome"`
	Series     int    `json:"series"`
	Repeticoes int    `json:"repeticoes"`
	Duracao    int    `json:"duracao_segundos"`
}

// TableName especifica o nome da tabela
func (ProgramaTreinamento) TableName() string {
	return "programas_treinamento"
}

func (SessaoTreinamento) TableName() string {
	return "sessoes_treinamento"
}
