package models

import "time"

type CreateCoberturaRequest struct {
	DataCobertura          time.Time     `json:"data_cobertura" validate:"required"`
	TipoCobertura          TipoCobertura `json:"tipo_cobertura" validate:"required"`
	MetodoCobertura        string        `json:"metodo_cobertura"`
	LaboratorioID          *uint         `json:"laboratorio_id"`
	ProbabilidadeConcepcao *float64      `json:"probabilidade_concepcao"`
	Observacoes            string        `json:"observacoes"`
}

type CreateAvaliacaoSemenRequest struct {
	CoberturaID                 *uint      `json:"cobertura_id"`
	DataColeta                  time.Time  `json:"data_coleta" validate:"required"`
	DataAnalise                 time.Time  `json:"data_analise" validate:"required"`
	VolumeSemen                 *float64   `json:"volume_semen"`
	ConcentracaoEspermatozoides *float64   `json:"concentracao_espermatozoides"`
	MotilidadeProgressiva       *float64   `json:"motilidade_progressiva"`
	MotilidadeTotal             *float64   `json:"motilidade_total"`
	Viabilidade                 *float64   `json:"viabilidade"`
	MorfologiaNormal            *float64   `json:"morfologia_normal"`
	DataValidade                *time.Time `json:"data_validade"`
	TemperaturaArmazenamento    *float64   `json:"temperatura_armazenamento"`
	TecnicoResponsavel          string     `json:"tecnico_responsavel"`
	Observacoes                 string     `json:"observacoes"`
}

type RegistrarPartoRequest struct {
	DataParto        time.Time  `json:"data_parto" validate:"required"`
	TipoParto        *TipoParto `json:"tipo_parto"`
	ResultadoParto   string     `json:"resultado_parto" validate:"required"`
	ObservacoesParto string     `json:"observacoes_parto"`
}

type CreatePerformanceMaternaRequest struct {
	PesoInicioGestacao       *float64        `json:"peso_inicio_gestacao"`
	PesoFimGestacao          *float64        `json:"peso_fim_gestacao"`
	GanhoPesoGestacao        *float64        `json:"ganho_peso_gestacao"`
	ProducaoLeiteDiaria      *float64        `json:"producao_leite_diaria"`
	QualidadeLeite           *QualidadeLeite `json:"qualidade_leite"`
	CuidadoMaterno           *CuidadoMaterno `json:"cuidado_materno"`
	TempoDesmame             *int            `json:"tempo_desmame"`
	PesoPotroDesmame         *float64        `json:"peso_potro_desmame"`
	TempoRecuperacaoPosParto *int            `json:"tempo_recuperacao_pos_parto"`
	IntervaloProximoParto    *int            `json:"intervalo_proximo_parto"`
	Observacoes              string          `json:"observacoes"`
}

type CreateUltrassonografiaRequest struct {
	DataExame             time.Time  `json:"data_exame" validate:"required"`
	IdadeGestacional      *int       `json:"idade_gestacional"`
	PresencaEmbriao       *bool      `json:"presenca_embriao"`
	NumeroEmbrioes        *int       `json:"numero_embrioes"`
	BatimentoCardiaco     *bool      `json:"batimento_cardiaco"`
	DesenvolvimentoNormal *bool      `json:"desenvolvimento_normal"`
	TamanhoEmbriao        *float64   `json:"tamanho_embriao"`
	FrequenciaCardiaca    *int       `json:"frequencia_cardiaca"`
	Diagnostico           string     `json:"diagnostico"`
	Observacoes           string     `json:"observacoes"`
	ProximoExame          *time.Time `json:"proximo_exame"`
}
