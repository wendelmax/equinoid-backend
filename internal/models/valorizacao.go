package models

import (
	"time"

	"gorm.io/gorm"
)

// RegistroValorizacao representa um registro de valorização
type RegistroValorizacao struct {
	ID                       uint                 `json:"id" gorm:"primaryKey"`
	Equinoid                 string               `json:"equinoid" gorm:"size:25;not null"`
	Categoria                CategoriaValorizacao `json:"categoria" gorm:"size:50;not null"`
	TipoRegistro             string               `json:"tipo_registro" gorm:"size:100;not null"`
	Titulo                   string               `json:"titulo" gorm:"size:200;not null"`
	Descricao                string               `json:"descricao" gorm:"type:text"`
	DataRegistro             time.Time            `json:"data_registro" gorm:"not null"`
	DataValidade             *time.Time           `json:"data_validade"`
	LocalEvento              string               `json:"local_evento" gorm:"size:200"`
	Pais                     string               `json:"pais" gorm:"size:3"`
	Estado                   string               `json:"estado" gorm:"size:100"`
	Cidade                   string               `json:"cidade" gorm:"size:100"`
	Organizacao              string               `json:"organizacao" gorm:"size:200"`
	InstituicaoCertificadora string               `json:"instituicao_certificadora" gorm:"size:200"`
	NumeroCertificado        string               `json:"numero_certificado" gorm:"size:100"`
	ValorMonetario           *float64             `json:"valor_monetario" gorm:"type:decimal(15,2)"`
	PontosValorizacao        int                  `json:"pontos_valorizacao" gorm:"default:0"`
	NivelImportancia         NivelImportancia     `json:"nivel_importancia" gorm:"default:'medio'"`
	Documentos               JSONB                `json:"documentos" gorm:"type:jsonb"`
	EvidenciaFotografica     JSONB                `json:"evidencia_fotografica" gorm:"type:jsonb"`
	EvidenciaVideo           JSONB                `json:"evidencia_video" gorm:"type:jsonb"`
	StatusValidacao          StatusValidacao      `json:"status_validacao" gorm:"default:'pendente'"`
	ValidadoPor              *uint                `json:"validado_por"`
	DataValidacao            *time.Time           `json:"data_validacao"`
	ObservacoesValidacao     string               `json:"observacoes_validacao" gorm:"type:text"`
	HashDocumento            string               `json:"hash_documento" gorm:"size:255"`
	BlockchainTxHash         string               `json:"blockchain_tx_hash" gorm:"size:255"`
	CriadoPor                uint                 `json:"criado_por" gorm:"not null"`
	CreatedAt                time.Time            `json:"created_at"`
	UpdatedAt                time.Time            `json:"updated_at"`
	DeletedAt                gorm.DeletedAt       `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Equino             *Equino            `json:"equino,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
	Validador          *User              `json:"validador,omitempty" gorm:"foreignKey:ValidadoPor"`
	Criador            *User              `json:"criador,omitempty" gorm:"foreignKey:CriadoPor"`
	Competicoes        []Competicao       `json:"competicoes,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
	Leiloes            []LeilaoValorizacao `json:"leiloes,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
	RegistrosMidia     []RegistroMidia    `json:"registros_midia,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
	RegistrosSaude     []RegistroSaude    `json:"registros_saude,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
	RegistrosEducacao  []RegistroEducacao `json:"registros_educacao,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
	RegistrosViagens   []RegistroViagem   `json:"registros_viagens,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
	RegistrosParcerias []RegistroParceria `json:"registros_parcerias,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
	RegistrosAnalise   []RegistroAnalise  `json:"registros_analise,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
}

// CategoriaValorizacao define as categorias de valorização
type CategoriaValorizacao string

const (
	CategoriaCompeticao      CategoriaValorizacao = "competicao"
	CategoriaReproducao      CategoriaValorizacao = "reproducao"
	CategoriaSaude           CategoriaValorizacao = "saude"
	CategoriaTreinamento     CategoriaValorizacao = "treinamento"
	CategoriaComercial       CategoriaValorizacao = "comercial"
	CategoriaMidia           CategoriaValorizacao = "midia"
	CategoriaEducacao        CategoriaValorizacao = "educacao"
	CategoriaViagens         CategoriaValorizacao = "viagens"
	CategoriaReconhecimentos CategoriaValorizacao = "reconhecimentos"
	CategoriaParcerias       CategoriaValorizacao = "parcerias"
	CategoriaAnalise         CategoriaValorizacao = "analise"
)

// NivelImportancia define o nível de importância
type NivelImportancia string

const (
	NivelBaixo       NivelImportancia = "baixo"
	NivelMedio       NivelImportancia = "medio"
	NivelAlto        NivelImportancia = "alto"
	NivelMuitoAlto   NivelImportancia = "muito_alto"
	NivelCritico     NivelImportancia = "critico"
	NivelExcepcional NivelImportancia = "excepcional"
)

// StatusValidacao define o status de validação
type StatusValidacao string

const (
	StatusPendente  StatusValidacao = "pendente"
	StatusValidado  StatusValidacao = "validado"
	StatusAprovado  StatusValidacao = "aprovado"
	StatusRejeitado StatusValidacao = "rejeitado"
)

// RankingValorizacao representa um ranking de valorização
type RankingValorizacao struct {
	ID                uint             `json:"id" gorm:"primaryKey"`
	Equinoid          string           `json:"equinoid" gorm:"size:25;not null"`
	TipoRanking       TipoRanking      `json:"tipo_ranking" gorm:"size:50;not null"`
	CategoriaRanking  string           `json:"categoria_ranking" gorm:"size:100"`
	Posicao           int              `json:"posicao" gorm:"not null"`
	PontuacaoTotal    int              `json:"pontuacao_total" gorm:"not null"`
	NivelValorizacao  NivelValorizacao `json:"nivel_valorizacao" gorm:"not null"`
	DataRanking       time.Time        `json:"data_ranking" gorm:"not null"`
	PeriodoReferencia string           `json:"periodo_referencia" gorm:"size:20"`
	CreatedAt         time.Time        `json:"created_at"`
	DeletedAt         gorm.DeletedAt   `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Equino *Equino `json:"equino,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
}

// TipoRanking define os tipos de ranking
type TipoRanking string

const (
	TipoRankingNacional      TipoRanking = "nacional"
	TipoRankingInternacional TipoRanking = "internacional"
	TipoRankingRaca          TipoRanking = "raca"
	TipoRankingCategoria     TipoRanking = "categoria"
)

// NivelValorizacao define os níveis de valorização
type NivelValorizacao string

const (
	NivelEstrela  NivelValorizacao = "estrela"
	NivelBronze   NivelValorizacao = "bronze"
	NivelPrata    NivelValorizacao = "prata"
	NivelOuro     NivelValorizacao = "ouro"
	NivelDiamante NivelValorizacao = "diamante"
)

// Competicao representa uma competição
type Competicao struct {
	ID                    uint            `json:"id" gorm:"primaryKey"`
	RegistroValorizacaoID uint            `json:"registro_valorizacao_id" gorm:"not null"`
	NomeCompeticao        string          `json:"nome_competicao" gorm:"size:200;not null"`
	CategoriaCompeticao   string          `json:"categoria_competicao" gorm:"size:100"`
	NivelCompeticao       NivelCompeticao `json:"nivel_competicao" gorm:"not null"`
	Modalidade            string          `json:"modalidade" gorm:"size:100"`
	Posicao               *int            `json:"posicao"`
	TotalParticipantes    *int            `json:"total_participantes"`
	Pontuacao             *float64        `json:"pontuacao" gorm:"type:decimal(8,2)"`
	TempoProva            *time.Duration  `json:"tempo_prova" swaggertype:"integer"`
	PremioMonetario       *float64        `json:"premio_monetario" gorm:"type:decimal(15,2)"`
	Trofeu                string          `json:"trofeu" gorm:"size:200"`
	Medalha               *TipoMedalha    `json:"medalha"`
	Certificado           string          `json:"certificado" gorm:"size:200"`
	JuizPrincipal         string          `json:"juiz_principal" gorm:"size:200"`
	JuizesAuxiliares      JSONB           `json:"juizes_auxiliares" gorm:"type:jsonb"`
	CreatedAt             time.Time       `json:"created_at"`
	DeletedAt             gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	RegistroValorizacao *RegistroValorizacao `json:"registro_valorizacao,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
}

// NivelCompeticao define o nível da competição
type NivelCompeticao string

const (
	NivelLocal         NivelCompeticao = "local"
	NivelRegional      NivelCompeticao = "regional"
	NivelNacional      NivelCompeticao = "nacional"
	NivelInternacional NivelCompeticao = "internacional"
)

// TipoMedalha define o tipo de medalha
type TipoMedalha string

const (
	MedalhaOuro   TipoMedalha = "ouro"
	MedalhaPrata  TipoMedalha = "prata"
	MedalhaBronze TipoMedalha = "bronze"
)

type LeilaoValorizacao struct {
	ID                    uint            `json:"id" gorm:"primaryKey"`
	RegistroValorizacaoID uint            `json:"registro_valorizacao_id" gorm:"not null"`
	NomeLeilao            string          `json:"nome_leilao" gorm:"size:200;not null"`
	TipoLeilao            TipoLeilao      `json:"tipo_leilao" gorm:"not null"`
	Especializacao        string          `json:"especializacao" gorm:"size:100"`
	CasaLeiloeira         string          `json:"casa_leiloeira" gorm:"size:200"`
	Organizador           string          `json:"organizador" gorm:"size:200"`
	LocalLeilao           string          `json:"local_leilao" gorm:"size:200"`
	Pais                  string          `json:"pais" gorm:"size:3"`
	Estado                string          `json:"estado" gorm:"size:100"`
	Cidade                string          `json:"cidade" gorm:"size:100"`
	DataLeilao            time.Time       `json:"data_leilao" gorm:"not null"`
	DataInicio            *time.Time      `json:"data_inicio"`
	DataFim               *time.Time      `json:"data_fim"`
	PrecoLance            *float64        `json:"preco_lance" gorm:"type:decimal(15,2)"`
	PrecoVenda            *float64        `json:"preco_venda" gorm:"type:decimal(15,2)"`
	PosicaoLeilao         *int            `json:"posicao_leilao"`
	TotalParticipantes    *int            `json:"total_participantes"`
	ValorizacaoPercentual *float64        `json:"valorizacao_percentual" gorm:"type:decimal(5,2)"`
	StatusLeilao          StatusLeilao    `json:"status_leilao" gorm:"default:'agendado'"`
	Resultado             ResultadoLeilao `json:"resultado" gorm:"default:'nao_vendido'"`
	CatalogoLeilao        string          `json:"catalogo_leilao" gorm:"size:255"`
	FotosLeilao           JSONB           `json:"fotos_leilao" gorm:"type:jsonb"`
	VideosLeilao          JSONB           `json:"videos_leilao" gorm:"type:jsonb"`
	CertificadoVenda      string          `json:"certificado_venda" gorm:"size:255"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	DeletedAt             gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	RegistroValorizacao *RegistroValorizacao `json:"registro_valorizacao,omitempty" gorm:"foreignKey:RegistroValorizacaoID"`
	LancesLeilao        []LanceLeilao        `json:"lances_leilao,omitempty" gorm:"foreignKey:LeilaoID"`
}

func (LeilaoValorizacao) TableName() string {
	return "leilaos"
}

// TipoLeilao define o tipo de leilão

// ResultadoLeilao define o resultado do leilão
type ResultadoLeilao string

const (
	ResultadoVendido    ResultadoLeilao = "vendido"
	ResultadoNaoVendido ResultadoLeilao = "nao_vendido"
	ResultadoRetirado   ResultadoLeilao = "retirado"
)

// LanceLeilao representa um lance de leilão
type LanceLeilao struct {
	ID               uint             `json:"id" gorm:"primaryKey"`
	LeilaoID         uint             `json:"leilao_id" gorm:"not null"`
	ValorLance       float64          `json:"valor_lance" gorm:"type:decimal(15,2);not null"`
	DataLance        time.Time        `json:"data_lance" gorm:"not null"`
	TipoLance        TipoLance        `json:"tipo_lance" gorm:"not null"`
	ParticipanteID   *uint            `json:"participante_id"`
	ParticipanteNome string           `json:"participante_nome" gorm:"size:200"`
	ParticipanteTipo TipoParticipante `json:"participante_tipo"`
	StatusLance      StatusLance      `json:"status_lance" gorm:"default:'ativo'"`
	CreatedAt        time.Time        `json:"created_at"`
	DeletedAt        gorm.DeletedAt   `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Leilao       *LeilaoValorizacao `json:"leilao,omitempty" gorm:"foreignKey:LeilaoID"`
	Participante *User   `json:"participante,omitempty" gorm:"foreignKey:ParticipanteID"`
}

// TipoLance define o tipo de lance
type TipoLance string

const (
	TipoLanceInicial    TipoLance = "inicial"
	TipoLanceIncremento TipoLance = "incremento"
	TipoLanceFinal      TipoLance = "final"
)

// TipoParticipante define o tipo de participante
type TipoParticipante string

const (
	TipoParticipantePessoaFisica   TipoParticipante = "pessoa_fisica"
	TipoParticipantePessoaJuridica TipoParticipante = "pessoa_juridica"
	TipoParticipanteRepresentante  TipoParticipante = "representante"
)

// StatusLance define o status do lance
type StatusLance string

const (
	StatusLanceAtivo     StatusLance = "ativo"
	StatusLanceSuperado  StatusLance = "superado"
	StatusLanceVencedor  StatusLance = "vencedor"
	StatusLanceCancelado StatusLance = "cancelado"
)

// Request DTOs para valorização
type RankingItem struct {
	Equinoid       string `json:"equinoid"`
	Nome           string `json:"nome"`
	TotalPontos    int    `json:"total_pontos"`
	TotalRegistros int    `json:"total_registros"`
}

type CreateValorizacaoRequest struct {
	Categoria                CategoriaValorizacao `json:"categoria" validate:"required"`
	TipoRegistro             string               `json:"tipo_registro" validate:"required"`
	Titulo                   string               `json:"titulo" validate:"required"`
	Descricao                string               `json:"descricao"`
	DataRegistro             time.Time            `json:"data_registro" validate:"required"`
	DataValidade             *time.Time           `json:"data_validade"`
	LocalEvento              string               `json:"local_evento"`
	Pais                     string               `json:"pais"`
	Estado                   string               `json:"estado"`
	Cidade                   string               `json:"cidade"`
	Organizacao              string               `json:"organizacao"`
	InstituicaoCertificadora string               `json:"instituicao_certificadora"`
	NumeroCertificado        string               `json:"numero_certificado"`
	ValorMonetario           *float64             `json:"valor_monetario"`
	NivelImportancia         NivelImportancia     `json:"nivel_importancia" validate:"required"`
	Documentos               []DocumentoEvento    `json:"documentos"`
	EvidenciaFotografica     []DocumentoEvento    `json:"evidencia_fotografica"`
	EvidenciaVideo           []DocumentoEvento    `json:"evidencia_video"`

	// Dados específicos para leilões
	DadosLeilao *DadosLeilaoRequest `json:"dados_leilao,omitempty"`
}

type DadosLeilaoRequest struct {
	NomeLeilao            string          `json:"nome_leilao" validate:"required"`
	TipoLeilao            TipoLeilao      `json:"tipo_leilao" validate:"required"`
	Especializacao        string          `json:"especializacao"`
	CasaLeiloeira         string          `json:"casa_leiloeira"`
	DataLeilao            time.Time       `json:"data_leilao" validate:"required"`
	PrecoLance            *float64        `json:"preco_lance"`
	PrecoVenda            *float64        `json:"preco_venda"`
	PosicaoLeilao         *int            `json:"posicao_leilao"`
	ValorizacaoPercentual *float64        `json:"valorizacao_percentual"`
	StatusLeilao          StatusLeilao    `json:"status_leilao"`
	Resultado             ResultadoLeilao `json:"resultado"`
}
