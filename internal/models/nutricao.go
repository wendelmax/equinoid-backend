package models

import (
	"time"

	"gorm.io/gorm"
)

// PlanoNutricional representa um plano alimentar
type PlanoNutricional struct {
	ID                    uint           `json:"id" gorm:"primaryKey"`
	EquinoID              uint           `json:"equino_id" gorm:"not null;index"`
	VeterinarioID         *uint          `json:"veterinario_id" gorm:"index"`
	NutricionistaID       *uint          `json:"nutricionista_id" gorm:"index"`
	TipoPlano             TipoPlano      `json:"tipo_plano" gorm:"size:50;not null"`
	Objetivo              string         `json:"objetivo" gorm:"size:200"`
	PesoAtual             float64        `json:"peso_atual" gorm:"type:decimal(6,2)"`
	PesoIdeal             float64        `json:"peso_ideal" gorm:"type:decimal(6,2)"`
	CaloriasObjetivo      int            `json:"calorias_objetivo"`
	ProteinasGramas       float64        `json:"proteinas_gramas" gorm:"type:decimal(8,2)"`
	CarboidratosGramas    float64        `json:"carboidratos_gramas" gorm:"type:decimal(8,2)"`
	GordurasGramas        float64        `json:"gorduras_gramas" gorm:"type:decimal(8,2)"`
	FibrasGramas          float64        `json:"fibras_gramas" gorm:"type:decimal(8,2)"`
	Suplementos           JSONB          `json:"suplementos" gorm:"type:jsonb"`
	Restricoes            JSONB          `json:"restricoes" gorm:"type:jsonb"`
	FrequenciaRefeicoes   int            `json:"frequencia_refeicoes" gorm:"default:3"`
	HorarioRefeicoes      JSONB          `json:"horario_refeicoes" gorm:"type:jsonb"`
	ObservacoesGerais     string         `json:"observacoes_gerais" gorm:"type:text"`
	GeradoPorIA           bool           `json:"gerado_por_ia" gorm:"default:false"`
	PromptIA              string         `json:"prompt_ia" gorm:"type:text"`
	RespostaIA            string         `json:"resposta_ia" gorm:"type:text"`
	DataInicio            time.Time      `json:"data_inicio"`
	DataFim               *time.Time     `json:"data_fim"`
	Status                StatusPlano    `json:"status" gorm:"size:20;default:'ativo'"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Equino        *Equino       `json:"equino,omitempty" gorm:"foreignKey:EquinoID"`
	Veterinario   *User         `json:"veterinario,omitempty" gorm:"foreignKey:VeterinarioID"`
	Nutricionista *User         `json:"nutricionista,omitempty" gorm:"foreignKey:NutricionistaID"`
	Refeicoes     []Refeicao    `json:"refeicoes,omitempty" gorm:"foreignKey:PlanoNutricionalID"`
}

// TipoPlano define os tipos de plano
type TipoPlano string

const (
	TipoPlanoManutencao    TipoPlano = "manutencao"
	TipoPlanoGanhoMassa    TipoPlano = "ganho_massa"
	TipoPlanoPerdaPeso     TipoPlano = "perda_peso"
	TipoPlanoRecuperacao   TipoPlano = "recuperacao"
	TipoPlanoAltoRendimento TipoPlano = "alto_rendimento"
	TipoPlanoGestacao      TipoPlano = "gestacao"
	TipoPlanoLactacao      TipoPlano = "lactacao"
)

// StatusPlano define o status do plano
type StatusPlano string

const (
	StatusPlanoAtivo    StatusPlano = "ativo"
	StatusPlanoPausado  StatusPlano = "pausado"
	StatusPlanoFinalizado StatusPlano = "finalizado"
	StatusPlanoCancelado StatusPlano = "cancelado"
)

// Refeicao representa uma refeição registrada
type Refeicao struct {
	ID                 uint           `json:"id" gorm:"primaryKey"`
	PlanoNutricionalID uint           `json:"plano_nutricional_id" gorm:"not null;index"`
	EquinoID           uint           `json:"equino_id" gorm:"not null;index"`
	DataRefeicao       time.Time      `json:"data_refeicao" gorm:"not null;index"`
	TipoRefeicao       string         `json:"tipo_refeicao" gorm:"size:50"`
	Alimentos          JSONB          `json:"alimentos" gorm:"type:jsonb"`
	QuantidadeTotal    float64        `json:"quantidade_total" gorm:"type:decimal(8,2)"`
	CaloriasConsumidas int            `json:"calorias_consumidas"`
	Observacoes        string         `json:"observacoes" gorm:"type:text"`
	RegistradoPor      uint           `json:"registrado_por" gorm:"not null"`
	CreatedAt          time.Time      `json:"created_at"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	PlanoNutricional *PlanoNutricional `json:"plano_nutricional,omitempty" gorm:"foreignKey:PlanoNutricionalID"`
	Equino           *Equino           `json:"equino,omitempty" gorm:"foreignKey:EquinoID"`
	Usuario          *User             `json:"usuario,omitempty" gorm:"foreignKey:RegistradoPor"`
}

// --- REQUEST/RESPONSE MODELS ---

// CreatePlanoNutricionalRequest representa requisição de criação
type CreatePlanoNutricionalRequest struct {
	Equinoid            string    `json:"equinoid" validate:"required"`
	TipoPlano           TipoPlano `json:"tipo_plano" validate:"required"`
	Objetivo            string    `json:"objetivo" validate:"required"`
	PesoAtual           float64   `json:"peso_atual" validate:"required,gt=0"`
	PesoIdeal           float64   `json:"peso_ideal" validate:"required,gt=0"`
	FrequenciaRefeicoes int       `json:"frequencia_refeicoes" validate:"min=2,max=6"`
	Restricoes          []string  `json:"restricoes"`
	Suplementos         []string  `json:"suplementos"`
	ObservacoesGerais   string    `json:"observacoes_gerais"`
	GerarComIA          bool      `json:"gerar_com_ia"`
}

// PlanoNutricionalResponse representa resposta de plano
type PlanoNutricionalResponse struct {
	ID                  uint        `json:"id"`
	Equinoid            string      `json:"equinoid"`
	EquinoNome          string      `json:"equino_nome"`
	TipoPlano           TipoPlano   `json:"tipo_plano"`
	Objetivo            string      `json:"objetivo"`
	PesoAtual           float64     `json:"peso_atual"`
	PesoIdeal           float64     `json:"peso_ideal"`
	CaloriasObjetivo    int         `json:"calorias_objetivo"`
	ProteinasGramas     float64     `json:"proteinas_gramas"`
	CarboidratosGramas  float64     `json:"carboidratos_gramas"`
	GordurasGramas      float64     `json:"gorduras_gramas"`
	FibrasGramas        float64     `json:"fibras_gramas"`
	Suplementos         []string    `json:"suplementos"`
	Restricoes          []string    `json:"restricoes"`
	FrequenciaRefeicoes int         `json:"frequencia_refeicoes"`
	HorarioRefeicoes    []string    `json:"horario_refeicoes"`
	ObservacoesGerais   string      `json:"observacoes_gerais"`
	GeradoPorIA         bool        `json:"gerado_por_ia"`
	Status              StatusPlano `json:"status"`
	DataInicio          time.Time   `json:"data_inicio"`
	DataFim             *time.Time  `json:"data_fim"`
	CreatedAt           time.Time   `json:"created_at"`
}

// CreateRefeicaoRequest representa requisição de registro de refeição
type CreateRefeicaoRequest struct {
	Equinoid       string                   `json:"equinoid" validate:"required"`
	TipoRefeicao   string                   `json:"tipo_refeicao" validate:"required"`
	Alimentos      []AlimentoRefeicao       `json:"alimentos" validate:"required,min=1"`
	DataRefeicao   *time.Time               `json:"data_refeicao"`
	Observacoes    string                   `json:"observacoes"`
}

// AlimentoRefeicao representa um alimento na refeição
type AlimentoRefeicao struct {
	Nome       string  `json:"nome" validate:"required"`
	Quantidade float64 `json:"quantidade" validate:"required,gt=0"`
	Unidade    string  `json:"unidade" validate:"required"`
	Calorias   int     `json:"calorias"`
}

// SugestaoIARequest representa requisição para IA
type SugestaoIARequest struct {
	Equinoid  string  `json:"equinoid" validate:"required"`
	TipoPlano TipoPlano `json:"tipo_plano" validate:"required"`
	Objetivo  string  `json:"objetivo" validate:"required"`
	PesoAtual float64 `json:"peso_atual" validate:"required,gt=0"`
}

// SugestaoIAResponse representa resposta da IA
type SugestaoIAResponse struct {
	Sugestao            string             `json:"sugestao"`
	PlanoSugerido       PlanoNutricionalIA `json:"plano_sugerido"`
	RecomendacoesExtras []string           `json:"recomendacoes_extras"`
	GeradoEm            time.Time          `json:"gerado_em"`
}

// PlanoNutricionalIA representa plano gerado por IA
type PlanoNutricionalIA struct {
	CaloriasObjetivo   int              `json:"calorias_objetivo"`
	Macronutrientes    MacronutrientesIA `json:"macronutrientes"`
	RefeicoesDetalhadas []RefeicaoIA    `json:"refeicoes_detalhadas"`
	Suplementos        []string         `json:"suplementos"`
}

// MacronutrientesIA representa macros sugeridos
type MacronutrientesIA struct {
	Proteinas    float64 `json:"proteinas_gramas"`
	Carboidratos float64 `json:"carboidratos_gramas"`
	Gorduras     float64 `json:"gorduras_gramas"`
	Fibras       float64 `json:"fibras_gramas"`
}

// RefeicaoIA representa refeição sugerida pela IA
type RefeicaoIA struct {
	Horario   string             `json:"horario"`
	Alimentos []AlimentoRefeicao `json:"alimentos"`
	Calorias  int                `json:"calorias_total"`
}

// ToResponse converte PlanoNutricional para PlanoNutricionalResponse
func (p *PlanoNutricional) ToResponse() *PlanoNutricionalResponse {
	response := &PlanoNutricionalResponse{
		ID:                  p.ID,
		TipoPlano:           p.TipoPlano,
		Objetivo:            p.Objetivo,
		PesoAtual:           p.PesoAtual,
		PesoIdeal:           p.PesoIdeal,
		CaloriasObjetivo:    p.CaloriasObjetivo,
		ProteinasGramas:     p.ProteinasGramas,
		CarboidratosGramas:  p.CarboidratosGramas,
		GordurasGramas:      p.GordurasGramas,
		FibrasGramas:        p.FibrasGramas,
		FrequenciaRefeicoes: p.FrequenciaRefeicoes,
		ObservacoesGerais:   p.ObservacoesGerais,
		GeradoPorIA:         p.GeradoPorIA,
		Status:              p.Status,
		DataInicio:          p.DataInicio,
		DataFim:             p.DataFim,
		CreatedAt:           p.CreatedAt,
	}

	if p.Equino != nil {
		response.Equinoid = p.Equino.Equinoid
		response.EquinoNome = p.Equino.Nome
	}

	if p.Suplementos != nil {
		if itens, ok := p.Suplementos["itens"].([]interface{}); ok {
			response.Suplementos = make([]string, len(itens))
			for i, s := range itens {
				if str, ok := s.(string); ok {
					response.Suplementos[i] = str
				}
			}
		}
	}

	if p.Restricoes != nil {
		if itens, ok := p.Restricoes["itens"].([]interface{}); ok {
			response.Restricoes = make([]string, len(itens))
			for i, r := range itens {
				if str, ok := r.(string); ok {
					response.Restricoes[i] = str
				}
			}
		}
	}

	if p.HorarioRefeicoes != nil {
		if itens, ok := p.HorarioRefeicoes["horarios"].([]interface{}); ok {
			response.HorarioRefeicoes = make([]string, len(itens))
			for i, h := range itens {
				if str, ok := h.(string); ok {
					response.HorarioRefeicoes[i] = str
				}
			}
		}
	}

	return response
}

// TableName especifica o nome da tabela
func (PlanoNutricional) TableName() string {
	return "planos_nutricionais"
}

func (Refeicao) TableName() string {
	return "refeicoes"
}
