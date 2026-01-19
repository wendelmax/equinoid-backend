package models

import (
	"time"

	"gorm.io/gorm"
)

// Cobertura representa uma cobertura realizada
type Cobertura struct {
	ID                     uint            `json:"id" gorm:"primaryKey"`
	ReprodutorEquinoid     string          `json:"reprodutor_equinoid" gorm:"size:25;not null"`
	MatrizEquinoid         string          `json:"matriz_equinoid" gorm:"size:25;not null"`
	DataCobertura          time.Time       `json:"data_cobertura" gorm:"not null"`
	TipoCobertura          TipoCobertura   `json:"tipo_cobertura" gorm:"not null"`
	MetodoCobertura        string          `json:"metodo_cobertura" gorm:"size:100"`
	VeterinarioResponsavel uint            `json:"veterinario_responsavel" gorm:"not null"`
	LaboratorioID          *uint           `json:"laboratorio_id"`
	StatusCobertura        StatusCobertura `json:"status_cobertura" gorm:"default:'pendente'"`
	DataConfirmacao        *time.Time      `json:"data_confirmacao"`
	ProbabilidadeConcepcao *float64        `json:"probabilidade_concepcao" gorm:"type:decimal(5,2)"`
	Observacoes            string          `json:"observacoes" gorm:"type:text"`
	Documentos             JSONB           `json:"documentos" gorm:"type:jsonb"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
	DeletedAt              gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Reprodutor  *Equino         `json:"reprodutor,omitempty" gorm:"foreignKey:ReprodutorEquinoid;references:Equinoid"`
	Matriz      *Equino         `json:"matriz,omitempty" gorm:"foreignKey:MatrizEquinoid;references:Equinoid"`
	Veterinario *User           `json:"veterinario,omitempty" gorm:"foreignKey:VeterinarioResponsavel"`
	Laboratorio *LaboratorioDNA `json:"laboratorio,omitempty" gorm:"foreignKey:LaboratorioID"`
	Gestacao    *Gestacao       `json:"gestacao,omitempty" gorm:"foreignKey:CoberturaID"`
}

// TipoCobertura define o tipo de cobertura
type TipoCobertura string

const (
	TipoCoberturaNatural     TipoCobertura = "natural"
	TipoCoberturaInseminacao TipoCobertura = "inseminacao"
	TipoCoberturaEmbriao     TipoCobertura = "embriao"
)

// StatusCobertura define o status da cobertura
type StatusCobertura string

const (
	StatusCoberturaPendente   StatusCobertura = "pendente"
	StatusCoberturaRealizada  StatusCobertura = "realizada"
	StatusCoberturaConfirmada StatusCobertura = "confirmada"
	StatusCoberturaFalhou     StatusCobertura = "falhou"
)

// AvaliacaoSemen representa uma avaliação de sêmen
type AvaliacaoSemen struct {
	ID                          uint               `json:"id" gorm:"primaryKey"`
	ReprodutorEquinoid          string             `json:"reprodutor_equinoid" gorm:"size:25;not null"`
	CoberturaID                 *uint              `json:"cobertura_id"`
	DataColeta                  time.Time          `json:"data_coleta" gorm:"not null"`
	DataAnalise                 time.Time          `json:"data_analise" gorm:"not null"`
	LaboratorioID               uint               `json:"laboratorio_id" gorm:"not null"`
	VolumeSemen                 *float64           `json:"volume_semen" gorm:"type:decimal(8,2)"`
	ConcentracaoEspermatozoides *float64           `json:"concentracao_espermatozoides" gorm:"type:decimal(12,2)"`
	MotiliadeProgressiva        *float64           `json:"motilidade_progressiva" gorm:"type:decimal(5,2)"`
	MotiliadeTotal              *float64           `json:"motilidade_total" gorm:"type:decimal(5,2)"`
	Viabilidade                 *float64           `json:"viabilidade" gorm:"type:decimal(5,2)"`
	MorfologiaNormal            *float64           `json:"morfologia_normal" gorm:"type:decimal(5,2)"`
	QualidadeGeral              QualidadeSemen     `json:"qualidade_geral" gorm:"not null"`
	AptidaoReprodutiva          AptidaoReprodutiva `json:"aptidao_reprodutiva" gorm:"not null"`
	DataValidade                *time.Time         `json:"data_validade"`
	TemperaturaArmazenamento    *float64           `json:"temperatura_armazenamento" gorm:"type:decimal(5,2)"`
	TecnicoResponsavel          string             `json:"tecnico_responsavel" gorm:"size:200"`
	Observacoes                 string             `json:"observacoes" gorm:"type:text"`
	Documentos                  JSONB              `json:"documentos" gorm:"type:jsonb"`
	CreatedAt                   time.Time          `json:"created_at"`
	DeletedAt                   gorm.DeletedAt     `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Reprodutor  *Equino         `json:"reprodutor,omitempty" gorm:"foreignKey:ReprodutorEquinoid;references:Equinoid"`
	Cobertura   *Cobertura      `json:"cobertura,omitempty" gorm:"foreignKey:CoberturaID"`
	Laboratorio *LaboratorioDNA `json:"laboratorio,omitempty" gorm:"foreignKey:LaboratorioID"`
}

// QualidadeSemen define a qualidade do sêmen
type QualidadeSemen string

const (
	QualidadeExcelente  QualidadeSemen = "excelente"
	QualidadeBoa        QualidadeSemen = "boa"
	QualidadeRegular    QualidadeSemen = "regular"
	QualidadeRuim       QualidadeSemen = "ruim"
	QualidadeInadequada QualidadeSemen = "inadequada"
)

// AptidaoReprodutiva define a aptidão reprodutiva
type AptidaoReprodutiva string

const (
	AptidaoAlta       AptidaoReprodutiva = "alta"
	AptidaoMedia      AptidaoReprodutiva = "media"
	AptidaoBaixa      AptidaoReprodutiva = "baixa"
	AptidaoInadequada AptidaoReprodutiva = "inadequada"
)

// Gestacao representa uma gestação
type Gestacao struct {
	ID                       uint           `json:"id" gorm:"primaryKey"`
	MatrizEquinoid           string         `json:"matriz_equinoid" gorm:"size:25;not null"`
	CoberturaID              uint           `json:"cobertura_id" gorm:"not null"`
	DataCobertura            time.Time      `json:"data_cobertura" gorm:"not null"`
	DataPrevistaParto        time.Time      `json:"data_prevista_parto" gorm:"not null"`
	DataRealParto            *time.Time     `json:"data_real_parto"`
	VeterinarioResponsavel   uint           `json:"veterinario_responsavel" gorm:"not null"`
	NumeroUltrassonografias  int            `json:"numero_ultrassonografias" gorm:"default:0"`
	UltimaUltrassonografia   *time.Time     `json:"ultima_ultrassonografia"`
	StatusGestacao           StatusGestacao `json:"status_gestacao" gorm:"default:'ativa'"`
	TipoParto                *TipoParto     `json:"tipo_parto"`
	PotroEquinoid            *string        `json:"potro_equinoid" gorm:"size:25"`
	PesoNascimento           *float64       `json:"peso_nascimento" gorm:"type:decimal(8,2)"`
	SexoPotro                *SexoEquino    `json:"sexo_potro"`
	Complicacoes             string         `json:"complicacoes" gorm:"type:text"`
	IntervencoesVeterinarias string         `json:"intervencoes_veterinarias" gorm:"type:text"`
	Observacoes              string         `json:"observacoes" gorm:"type:text"`
	Documentos               JSONB          `json:"documentos" gorm:"type:jsonb"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Matriz             *Equino             `json:"matriz,omitempty" gorm:"foreignKey:MatrizEquinoid;references:Equinoid"`
	Cobertura          *Cobertura          `json:"cobertura,omitempty" gorm:"foreignKey:CoberturaID"`
	Veterinario        *User               `json:"veterinario,omitempty" gorm:"foreignKey:VeterinarioResponsavel"`
	Potro              *Equino             `json:"potro,omitempty" gorm:"foreignKey:PotroEquinoid;references:Equinoid"`
	Ultrassonografias  []Ultrassonografia  `json:"ultrassonografias,omitempty" gorm:"foreignKey:GestacaoID"`
	PerformanceMaterna *PerformanceMaterna `json:"performance_materna,omitempty" gorm:"foreignKey:GestacaoID"`
}

// StatusGestacao define o status da gestação
type StatusGestacao string

const (
	StatusGestacaoAtiva        StatusGestacao = "ativa"
	StatusGestacaoConcluida    StatusGestacao = "concluida"
	StatusGestacaoInterrompida StatusGestacao = "interrompida"
	StatusGestacaoPerdida      StatusGestacao = "perdida"
)

// TipoParto define o tipo de parto
type TipoParto string

const (
	TipoPartoNormal    TipoParto = "normal"
	TipoPartoCesariana TipoParto = "cesariana"
	TipoPartoAssistido TipoParto = "assistido"
)

// Ultrassonografia representa um exame de ultrassonografia
type Ultrassonografia struct {
	ID                     uint           `json:"id" gorm:"primaryKey"`
	GestacaoID             uint           `json:"gestacao_id" gorm:"not null"`
	DataExame              time.Time      `json:"data_exame" gorm:"not null"`
	IdadeGestacional       *int           `json:"idade_gestacional"`
	VeterinarioResponsavel uint           `json:"veterinario_responsavel" gorm:"not null"`
	PresencaEmbriao        *bool          `json:"presenca_embriao"`
	NumeroEmbrioes         *int           `json:"numero_embrioes"`
	BatimentoCardiaco      *bool          `json:"batimento_cardiaco"`
	DesenvolvimentoNormal  *bool          `json:"desenvolvimento_normal"`
	TamanhoEmbriao         *float64       `json:"tamanho_embriao" gorm:"type:decimal(8,2)"`
	FrequenciaCardiaca     *int           `json:"frequencia_cardiaca"`
	Diagnostico            string         `json:"diagnostico" gorm:"type:text"`
	Observacoes            string         `json:"observacoes" gorm:"type:text"`
	ProximoExame           *time.Time     `json:"proximo_exame"`
	Documentos             JSONB          `json:"documentos" gorm:"type:jsonb"`
	CreatedAt              time.Time      `json:"created_at"`
	DeletedAt              gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Gestacao    *Gestacao `json:"gestacao,omitempty" gorm:"foreignKey:GestacaoID"`
	Veterinario *User     `json:"veterinario,omitempty" gorm:"foreignKey:VeterinarioResponsavel"`
}

// PerformanceMaterna representa a performance materna
type PerformanceMaterna struct {
	ID                       uint            `json:"id" gorm:"primaryKey"`
	MatrizEquinoid           string          `json:"matriz_equinoid" gorm:"size:25;not null"`
	GestacaoID               uint            `json:"gestacao_id" gorm:"not null"`
	PesoInicioGestacao       *float64        `json:"peso_inicio_gestacao" gorm:"type:decimal(8,2)"`
	PesoFimGestacao          *float64        `json:"peso_fim_gestacao" gorm:"type:decimal(8,2)"`
	GanhoPesoGestacao        *float64        `json:"ganho_peso_gestacao" gorm:"type:decimal(8,2)"`
	ProducaoLeiteDiaria      *float64        `json:"producao_leite_diaria" gorm:"type:decimal(8,2)"`
	QualidadeLeite           *QualidadeLeite `json:"qualidade_leite"`
	CuidadoMaterno           *CuidadoMaterno `json:"cuidado_materno"`
	TempoDesmame             *int            `json:"tempo_desmame"`
	PesoPotroDesmame         *float64        `json:"peso_potro_desmame" gorm:"type:decimal(8,2)"`
	TempoRecuperacaoPosParto *int            `json:"tempo_recuperacao_pos_parto"`
	IntervaloProximoParto    *int            `json:"intervalo_proximo_parto"`
	Observacoes              string          `json:"observacoes" gorm:"type:text"`
	CreatedAt                time.Time       `json:"created_at"`
	DeletedAt                gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Matriz   *Equino   `json:"matriz,omitempty" gorm:"foreignKey:MatrizEquinoid;references:Equinoid"`
	Gestacao *Gestacao `json:"gestacao,omitempty" gorm:"foreignKey:GestacaoID"`
}

// QualidadeLeite define a qualidade do leite
type QualidadeLeite string

const (
	QualidadeLeiteExcelente QualidadeLeite = "excelente"
	QualidadeLeiteBoa       QualidadeLeite = "boa"
	QualidadeLeiteRegular   QualidadeLeite = "regular"
	QualidadeLeiteRuim      QualidadeLeite = "ruim"
)

// CuidadoMaterno define o cuidado materno
type CuidadoMaterno string

const (
	CuidadoMaternoExcelente CuidadoMaterno = "excelente"
	CuidadoMaternoBom       CuidadoMaterno = "bom"
	CuidadoMaternoRegular   CuidadoMaterno = "regular"
	CuidadoMaternoRuim      CuidadoMaterno = "ruim"
)

// RankingReprodutivo representa um ranking reprodutivo
type RankingReprodutivo struct {
	ID                   uint                   `json:"id" gorm:"primaryKey"`
	Equinoid             string                 `json:"equinoid" gorm:"size:25;not null"`
	TipoRanking          TipoRankingReprodutivo `json:"tipo_ranking" gorm:"not null"`
	CategoriaRanking     string                 `json:"categoria_ranking" gorm:"size:100"`
	NumeroCrias          int                    `json:"numero_crias" gorm:"default:0"`
	CriasPremiadas       int                    `json:"crias_premiadas" gorm:"default:0"`
	TaxaSucesso          *float64               `json:"taxa_sucesso" gorm:"type:decimal(5,2)"`
	ValorTotalCrias      *float64               `json:"valor_total_crias" gorm:"type:decimal(15,2)"`
	ValorMedioCria       *float64               `json:"valor_medio_cria" gorm:"type:decimal(15,2)"`
	PosicaoRanking       *int                   `json:"posicao_ranking"`
	PontuacaoReprodutiva int                    `json:"pontuacao_reprodutiva" gorm:"default:0"`
	PeriodoReferencia    string                 `json:"periodo_referencia" gorm:"size:20"`
	DataRanking          time.Time              `json:"data_ranking" gorm:"not null"`
	CreatedAt            time.Time              `json:"created_at"`
	DeletedAt            gorm.DeletedAt         `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Equino *Equino `json:"equino,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
}

// TipoRankingReprodutivo define o tipo de ranking reprodutivo
type TipoRankingReprodutivo string

const (
	TipoRankingReprodutor TipoRankingReprodutivo = "reprodutor"
	TipoRankingMatriz     TipoRankingReprodutivo = "matriz"
)

// LaboratorioDNA representa um laboratório de DNA certificado
type LaboratorioDNA struct {
	ID                 uint           `json:"id" gorm:"primaryKey"`
	Nome               string         `json:"nome" gorm:"size:100;not null"`
	Codigo             string         `json:"codigo" gorm:"uniqueIndex;size:20;not null"`
	Pais               string         `json:"pais" gorm:"size:3;not null"`
	CertificacaoStatus string         `json:"certificacao_status" gorm:"size:20;default:'ativo'"`
	DataCertificacao   *time.Time     `json:"data_certificacao"`
	DataVencimento     *time.Time     `json:"data_vencimento"`
	ContatoEmail       string         `json:"contato_email" gorm:"size:100"`
	ContatoTelefone    string         `json:"contato_telefone" gorm:"size:20"`
	APIEndpoint        string         `json:"api_endpoint" gorm:"size:255"`
	APIKeyHash         string         `json:"api_key_hash" gorm:"size:255"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Coberturas      []Cobertura      `json:"coberturas,omitempty" gorm:"foreignKey:LaboratorioID"`
	AvaliacoesSemen []AvaliacaoSemen `json:"avaliacoes_semen,omitempty" gorm:"foreignKey:LaboratorioID"`
}
