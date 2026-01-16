package models

import (
	"time"

	"gorm.io/gorm"
)

// RegistroMidia representa um registro de mídia
type RegistroMidia struct {
	ID                     uint           `json:"id" gorm:"primaryKey"`
	RegistroValorizacaoID  uint           `json:"registro_valorizacao_id" gorm:"not null"`
	TipoMidia              TipoMidia      `json:"tipo_midia" gorm:"not null"`
	NomeProducao           string         `json:"nome_producao" gorm:"size:200"`
	Diretor                string         `json:"diretor" gorm:"size:200"`
	Estudio                string         `json:"estudio" gorm:"size:200"`
	Orcamento              *float64       `json:"orcamento" gorm:"type:decimal(15,2)"`
	Bilheteria             *float64       `json:"bilheteria" gorm:"type:decimal(15,2)"`
	Premiacoes             JSONB          `json:"premiacoes" gorm:"type:jsonb"`
	DuracaoParticipacao    string         `json:"duracao_participacao" gorm:"size:50"`
	Papel                  string         `json:"papel" gorm:"size:100"`
	SeguidoresRedesSociais *int           `json:"seguidores_redes_sociais"`
	AlcancePublicacao      *int           `json:"alcance_publicacao"`
	Engajamento            *float64       `json:"engajamento" gorm:"type:decimal(5,2)"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	RegistroValorizacao *RegistroValorizacao `json:"registro_valorizacao,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
}

// TipoMidia define o tipo de mídia
type TipoMidia string

const (
	TipoMidiaFilmeInternacional TipoMidia = "filme_internacional"
	TipoMidiaFilmeNacional      TipoMidia = "filme_nacional"
	TipoMidiaSerieTV            TipoMidia = "serie_tv"
	TipoMidiaPublicidade        TipoMidia = "publicidade"
	TipoMidiaDocumentario       TipoMidia = "documentario"
	TipoMidiaRevista            TipoMidia = "revista"
	TipoMidiaRedesSociais       TipoMidia = "redes_sociais"
	TipoMidiaProgramaTV         TipoMidia = "programa_tv"
	TipoMidiaLivro              TipoMidia = "livro"
)

// RegistroSaude representa um registro de saúde
type RegistroSaude struct {
	ID                    uint           `json:"id" gorm:"primaryKey"`
	RegistroValorizacaoID uint           `json:"registro_valorizacao_id" gorm:"not null"`
	IdadeAtual            *int           `json:"idade_atual"`
	Peso                  *float64       `json:"peso" gorm:"type:decimal(8,2)"`
	Altura                *float64       `json:"altura" gorm:"type:decimal(5,2)"`
	CondicaoFisica        CondicaoFisica `json:"condicao_fisica" gorm:"not null"`
	ProblemasSaude        JSONB          `json:"problemas_saude" gorm:"type:jsonb"`
	VacinasAtualizadas    bool           `json:"vacinas_atualizadas" gorm:"default:true"`
	ExamesRecentes        JSONB          `json:"exames_recentes" gorm:"type:jsonb"`
	CirurgiasRealizadas   JSONB          `json:"cirurgias_realizadas" gorm:"type:jsonb"`
	MedicamentosAtuais    JSONB          `json:"medicamentos_atuais" gorm:"type:jsonb"`
	Resistencia           Resistencia    `json:"resistencia" gorm:"default:'media'"`
	ExpectativaVida       string         `json:"expectativa_vida" gorm:"size:50"`
	ExameSangue           JSONB          `json:"exame_sangue" gorm:"type:jsonb"`
	ExameRaioX            JSONB          `json:"exame_raio_x" gorm:"type:jsonb"`
	ExameUltrassom        JSONB          `json:"exame_ultrassom" gorm:"type:jsonb"`
	ExameGenetico         JSONB          `json:"exame_genetico" gorm:"type:jsonb"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	RegistroValorizacao *RegistroValorizacao `json:"registro_valorizacao,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
}

// CondicaoFisica define a condição física
type CondicaoFisica string

const (
	CondicaoFisicaExcelente CondicaoFisica = "excelente"
	CondicaoFisicaBoa       CondicaoFisica = "boa"
	CondicaoFisicaRegular   CondicaoFisica = "regular"
	CondicaoFisicaRuim      CondicaoFisica = "ruim"
)

// Resistencia define a resistência
type Resistencia string

const (
	ResistenciaAlta  Resistencia = "alta"
	ResistenciaMedia Resistencia = "media"
	ResistenciaBaixa Resistencia = "baixa"
)

// RegistroEducacao representa um registro de educação
type RegistroEducacao struct {
	ID                       uint             `json:"id" gorm:"primaryKey"`
	RegistroValorizacaoID    uint             `json:"registro_valorizacao_id" gorm:"not null"`
	TipoCertificacao         TipoCertificacao `json:"tipo_certificacao" gorm:"not null"`
	Modalidade               string           `json:"modalidade" gorm:"size:100;not null"`
	Nivel                    NivelEducacao    `json:"nivel" gorm:"not null"`
	DuracaoTreinamento       string           `json:"duracao_treinamento" gorm:"size:50"`
	Instrutor                string           `json:"instrutor" gorm:"size:200"`
	Instituicao              string           `json:"instituicao" gorm:"size:200"`
	NotaFinal                *float64         `json:"nota_final" gorm:"type:decimal(3,1)"`
	HabilidadesDesenvolvidas JSONB            `json:"habilidades_desenvolvidas" gorm:"type:jsonb"`
	ProximosNiveis           JSONB            `json:"proximos_niveis" gorm:"type:jsonb"`
	NumeroCertificado        string           `json:"numero_certificado" gorm:"size:100"`
	DataCertificacao         *time.Time       `json:"data_certificacao"`
	DataValidade             *time.Time       `json:"data_validade"`
	InstituicaoCertificadora string           `json:"instituicao_certificadora" gorm:"size:200"`
	CreatedAt                time.Time        `json:"created_at"`
	UpdatedAt                time.Time        `json:"updated_at"`
	DeletedAt                gorm.DeletedAt   `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	RegistroValorizacao *RegistroValorizacao `json:"registro_valorizacao,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
}

// TipoCertificacao define o tipo de certificação
type TipoCertificacao string

const (
	TipoCertificacaoInternacional TipoCertificacao = "internacional"
	TipoCertificacaoNacional      TipoCertificacao = "nacional"
	TipoCertificacaoRegional      TipoCertificacao = "regional"
	TipoCertificacaoLocal         TipoCertificacao = "local"
)

// NivelEducacao define o nível de educação
type NivelEducacao string

const (
	NivelEducacaoBasico        NivelEducacao = "basico"
	NivelEducacaoIntermediario NivelEducacao = "intermediario"
	NivelEducacaoAvancado      NivelEducacao = "avancado"
	NivelEducacaoMestre        NivelEducacao = "mestre"
	NivelEducacaoInstrutor     NivelEducacao = "instrutor"
)

// RegistroViagem representa um registro de viagem
type RegistroViagem struct {
	ID                    uint           `json:"id" gorm:"primaryKey"`
	RegistroValorizacaoID uint           `json:"registro_valorizacao_id" gorm:"not null"`
	TipoViagem            TipoViagem     `json:"tipo_viagem" gorm:"not null"`
	Destino               string         `json:"destino" gorm:"size:200;not null"`
	PaisDestino           string         `json:"pais_destino" gorm:"size:3;not null"`
	DataPartida           time.Time      `json:"data_partida" gorm:"not null"`
	DataRetorno           *time.Time     `json:"data_retorno"`
	DuracaoDias           *int           `json:"duracao_dias"`
	MotivoViagem          string         `json:"motivo_viagem" gorm:"type:text"`
	MeioTransporte        MeioTransporte `json:"meio_transporte" gorm:"not null"`
	CustoViagem           *float64       `json:"custo_viagem" gorm:"type:decimal(15,2)"`
	Acomodacoes           string         `json:"acomodacoes" gorm:"size:200"`
	Acompanhantes         JSONB          `json:"acompanhantes" gorm:"type:jsonb"`
	ObjetivoAlcancado     bool           `json:"objetivo_alcancado" gorm:"default:true"`
	ResultadosObtidos     string         `json:"resultados_obtidos" gorm:"type:text"`
	CertificadosEmitidos  JSONB          `json:"certificados_emitidos" gorm:"type:jsonb"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	RegistroValorizacao *RegistroValorizacao `json:"registro_valorizacao,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
}

// TipoViagem define o tipo de viagem
type TipoViagem string

const (
	TipoViagemCompeticaoInternacional TipoViagem = "competicao_internacional"
	TipoViagemExportacao              TipoViagem = "exportacao"
	TipoViagemImportacao              TipoViagem = "importacao"
	TipoViagemTourMundial             TipoViagem = "tour_mundial"
	TipoViagemIntercambio             TipoViagem = "intercambio"
	TipoViagemMissaoComercial         TipoViagem = "missao_comercial"
	TipoViagemEventoInternacional     TipoViagem = "evento_internacional"
)

// MeioTransporte define o meio de transporte
type MeioTransporte string

const (
	MeioTransporteAereo     MeioTransporte = "aereo"
	MeioTransporteTerrestre MeioTransporte = "terrestre"
	MeioTransporteMaritimo  MeioTransporte = "maritimo"
)

// RegistroParceria representa um registro de parceria
type RegistroParceria struct {
	ID                    uint               `json:"id" gorm:"primaryKey"`
	RegistroValorizacaoID uint               `json:"registro_valorizacao_id" gorm:"not null"`
	TipoParceria          TipoParceria       `json:"tipo_parceria" gorm:"not null"`
	ParceiroNome          string             `json:"parceiro_nome" gorm:"size:200;not null"`
	ParceiroTipo          TipoParceiroEntity `json:"parceiro_tipo" gorm:"not null"`
	PaisParceiro          string             `json:"pais_parceiro" gorm:"size:3"`
	ObjetivoParceria      string             `json:"objetivo_parceria" gorm:"type:text;not null"`
	DuracaoParceria       string             `json:"duracao_parceria" gorm:"size:50"`
	Investimento          *float64           `json:"investimento" gorm:"type:decimal(15,2)"`
	ResultadosEsperados   string             `json:"resultados_esperados" gorm:"type:text"`
	ResultadosObtidos     string             `json:"resultados_obtidos" gorm:"type:text"`
	NumeroContrato        string             `json:"numero_contrato" gorm:"size:100"`
	DataInicio            *time.Time         `json:"data_inicio"`
	DataFim               *time.Time         `json:"data_fim"`
	StatusParceria        StatusParceria     `json:"status_parceria" gorm:"default:'ativa'"`
	CreatedAt             time.Time          `json:"created_at"`
	UpdatedAt             time.Time          `json:"updated_at"`
	DeletedAt             gorm.DeletedAt     `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	RegistroValorizacao *RegistroValorizacao `json:"registro_valorizacao,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
}

// TipoParceria define o tipo de parceria
type TipoParceria string

const (
	TipoParceriaHaras         TipoParceria = "haras"
	TipoParceriaVeterinario   TipoParceria = "veterinario"
	TipoParceriaTreinador     TipoParceria = "treinador"
	TipoParceriaComercial     TipoParceria = "comercial"
	TipoParceriaPesquisa      TipoParceria = "pesquisa"
	TipoParceriaAcademica     TipoParceria = "academica"
	TipoParceriaGovernamental TipoParceria = "governamental"
)

// TipoParceiroEntity define o tipo de entidade parceira
type TipoParceiroEntity string

const (
	TipoParceiroEntityPessoaFisica   TipoParceiroEntity = "pessoa_fisica"
	TipoParceiroEntityPessoaJuridica TipoParceiroEntity = "pessoa_juridica"
	TipoParceiroEntityInstituicao    TipoParceiroEntity = "instituicao"
	TipoParceiroEntityGoverno        TipoParceiroEntity = "governo"
)

// StatusParceria define o status da parceria
type StatusParceria string

const (
	StatusParceriaAtiva     StatusParceria = "ativa"
	StatusParceriaConcluida StatusParceria = "concluida"
	StatusParceriaSuspensa  StatusParceria = "suspensa"
	StatusParceriaCancelada StatusParceria = "cancelada"
)

// RegistroAnalise representa um registro de análise
type RegistroAnalise struct {
	ID                    uint           `json:"id" gorm:"primaryKey"`
	RegistroValorizacaoID uint           `json:"registro_valorizacao_id" gorm:"not null"`
	TipoAnalise           TipoAnalise    `json:"tipo_analise" gorm:"not null"`
	AnalistaNome          string         `json:"analista_nome" gorm:"size:200"`
	AnalistaCredencial    string         `json:"analista_credencial" gorm:"size:100"`
	InstituicaoAnalista   string         `json:"instituicao_analista" gorm:"size:200"`
	PaisAnalista          string         `json:"pais_analista" gorm:"size:3"`
	MetodologiaUsada      string         `json:"metodologia_usada" gorm:"type:text"`
	FerramentasUtilizadas JSONB          `json:"ferramentas_utilizadas" gorm:"type:jsonb"`
	AmostraAnalisada      *int           `json:"amostra_analisada"`
	PeriodoAnalise        string         `json:"periodo_analise" gorm:"size:50"`
	Conclusoes            string         `json:"conclusoes" gorm:"type:text;not null"`
	Recomendacoes         string         `json:"recomendacoes" gorm:"type:text"`
	NivelConfianca        *float64       `json:"nivel_confianca" gorm:"type:decimal(3,1)"`
	Predicoes             JSONB          `json:"predicoes" gorm:"type:jsonb"`
	Comparacoes           JSONB          `json:"comparacoes" gorm:"type:jsonb"`
	PredicaoConfirmada    *bool          `json:"predicao_confirmada"`
	DataConfirmacao       *time.Time     `json:"data_confirmacao"`
	PrecisaoPredicao      *float64       `json:"precisao_predicao" gorm:"type:decimal(5,2)"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	RegistroValorizacao *RegistroValorizacao `json:"registro_valorizacao,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
}

// TipoAnalise define o tipo de análise
type TipoAnalise string

const (
	TipoAnalisePerformance TipoAnalise = "performance"
	TipoAnaliseGenetica    TipoAnalise = "genetica"
	TipoAnaliseComercial   TipoAnalise = "comercial"
	TipoAnaliseTendencias  TipoAnalise = "tendencias"
	TipoAnalisePredicao    TipoAnalise = "predicao"
	TipoAnaliseComparativa TipoAnalise = "comparativa"
	TipoAnaliseEstatistica TipoAnalise = "estatistica"
)

// PerformanceReprodutiva representa a performance reprodutiva
type PerformanceReprodutiva struct {
	ID                    uint                       `json:"id" gorm:"primaryKey"`
	RegistroValorizacaoID uint                       `json:"registro_valorizacao_id" gorm:"not null"`
	TipoPerformance       TipoPerformanceReprodutiva `json:"tipo_performance" gorm:"not null"`
	NumeroCrias           int                        `json:"numero_crias" gorm:"default:0"`
	CriasPremiadas        int                        `json:"crias_premiadas" gorm:"default:0"`
	CriasCampeas          int                        `json:"crias_campeas" gorm:"default:0"`
	TaxaConcepcao         *float64                   `json:"taxa_concepcao" gorm:"type:decimal(5,2)"`
	IntervaloPartos       *int                       `json:"intervalo_partos"`
	ValorTotalCrias       *float64                   `json:"valor_total_crias" gorm:"type:decimal(15,2)"`
	ValorMedioCria        *float64                   `json:"valor_medio_cria" gorm:"type:decimal(15,2)"`
	RankingReprodutivo    *int                       `json:"ranking_reprodutivo"`
	PeriodoInicio         *time.Time                 `json:"periodo_inicio"`
	PeriodoFim            *time.Time                 `json:"periodo_fim"`
	CreatedAt             time.Time                  `json:"created_at"`
	DeletedAt             gorm.DeletedAt             `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	RegistroValorizacao *RegistroValorizacao `json:"registro_valorizacao,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
}

// TipoPerformanceReprodutiva define o tipo de performance reprodutiva
type TipoPerformanceReprodutiva string

const (
	TipoPerformanceReprodutor TipoPerformanceReprodutiva = "reprodutor"
	TipoPerformanceMatriz     TipoPerformanceReprodutiva = "matriz"
)

// Webhook representa um webhook registrado
type Webhook struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	URL       string         `json:"url" gorm:"not null"`
	Events    JSONB          `json:"events" gorm:"type:jsonb;not null"`
	Secret    string         `json:"secret" gorm:"not null"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// WebhookRequest representa a requisição de registro de webhook
type WebhookRequest struct {
	URL    string   `json:"url" validate:"required,url"`
	Events []string `json:"events" validate:"required,min=1"`
	Secret string   `json:"secret" validate:"required"`
}

// ChatbotQuery representa uma consulta ao chatbot
type ChatbotQuery struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	UserID         uint           `json:"user_id" gorm:"not null"`
	Equinoid       *string        `json:"equinoid" gorm:"size:25"`
	Query          string         `json:"query" gorm:"type:text;not null"`
	Response       string         `json:"response" gorm:"type:text"`
	Confidence     *float64       `json:"confidence" gorm:"type:decimal(3,2)"`
	ProcessingTime *int           `json:"processing_time"` // milliseconds
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	User   *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Equino *Equino `json:"equino,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
}

// ChatbotQueryRequest representa a requisição ao chatbot
type ChatbotQueryRequest struct {
	Equinoid *string `json:"equinoid"`
	Query    string  `json:"query" validate:"required"`
}

// ChatbotQueryResponse representa a resposta do chatbot
type ChatbotQueryResponse struct {
	Response    string   `json:"response"`
	Confidence  *float64 `json:"confidence"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// Common response structures
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination"`
}

type Pagination struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
	Pages int   `json:"pages"`
}

type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message"`
	Timestamp time.Time   `json:"timestamp"`
}

type ErrorResponse struct {
	Success   bool        `json:"success"`
	Error     string      `json:"error"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}
