package models

import (
	"time"

	"gorm.io/gorm"
)

// Tokenizacao representa a tokenização de um equino (RWA - Real World Asset)
type Tokenizacao struct {
	ID                                    uint           `json:"id" gorm:"primaryKey"`
	EquinoID                              uint           `json:"equino_id" gorm:"not null;uniqueIndex"`
	TotalTokens                           int            `json:"total_tokens" gorm:"not null"`
	TokensBloqueadosDono                  int            `json:"tokens_bloqueados_dono" gorm:"not null"`
	TokensDisponiveisVenda                int            `json:"tokens_disponiveis_venda" gorm:"not null"`
	TokensVendidos                        int            `json:"tokens_vendidos" gorm:"default:0"`
	PrecoInicialToken                     float64        `json:"preco_inicial_token" gorm:"type:decimal(15,2);not null"`
	ValorTotalTokenizado                  float64        `json:"valor_total_tokenizado" gorm:"type:decimal(15,2);not null"`
	PercentualMinimoDono                  float64        `json:"percentual_minimo_dono" gorm:"type:decimal(5,2);not null"`
	PercentualComercializavelPublicamente float64        `json:"percentual_comercializavel_publicamente" gorm:"type:decimal(5,2);not null"`
	TravaControleDono                     bool           `json:"trava_controle_dono" gorm:"default:true"`
	PrioridadeRecompra                    bool           `json:"prioridade_recompra" gorm:"default:true"`
	Status                                StatusToken    `json:"status" gorm:"size:20;not null;default:'pendente'"`
	CustoCustodiaMensal                   float64        `json:"custo_custodia_mensal" gorm:"type:decimal(15,2)"`
	TemSeguro                             bool           `json:"tem_seguro" gorm:"default:false"`
	ValorAssegurado                       *float64       `json:"valor_assegurado" gorm:"type:decimal(15,2)"`
	ApoliceSeguroURL                      string         `json:"apolice_seguro_url" gorm:"size:500"`
	GarantiasBiologicas                   JSONB          `json:"garantias_biologicas" gorm:"type:jsonb"`
	RatingRisco                           RatingRisco    `json:"rating_risco" gorm:"size:5;default:'A'"`
	DataInicio                            time.Time      `json:"data_inicio"`
	DataEncerramento                      *time.Time     `json:"data_encerramento"`
	SmartContractAddress                  string         `json:"smart_contract_address" gorm:"size:100"`
	BlockchainNetwork                     string         `json:"blockchain_network" gorm:"size:50"`
	ObservacoesCompliance                 string         `json:"observacoes_compliance" gorm:"type:text"`
	CreatedAt                             time.Time      `json:"created_at"`
	UpdatedAt                             time.Time      `json:"updated_at"`
	DeletedAt                             gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	Equino       *Equino                `json:"equino,omitempty" gorm:"foreignKey:EquinoID"`
	Transacoes   []TransacaoToken       `json:"transacoes,omitempty" gorm:"foreignKey:TokenizacaoID"`
	Participacoes []ParticipacaoToken   `json:"participacoes,omitempty" gorm:"foreignKey:TokenizacaoID"`
}

// StatusToken define os status possíveis de uma tokenização
type StatusToken string

const (
	StatusTokenPendente  StatusToken = "pendente"
	StatusTokenAtivo     StatusToken = "ativo"
	StatusTokenSuspenso  StatusToken = "suspenso"
	StatusTokenEncerrado StatusToken = "encerrado"
)

// RatingRisco define os ratings de risco
type RatingRisco string

const (
	RatingAAAPlus RatingRisco = "AAA+"
	RatingAAA     RatingRisco = "AAA"
	RatingAAPlus  RatingRisco = "AA+"
	RatingAA      RatingRisco = "AA"
	RatingAPlus   RatingRisco = "A+"
	RatingA       RatingRisco = "A"
	RatingBBBPlus RatingRisco = "BBB+"
	RatingBBB     RatingRisco = "BBB"
	RatingBBPlus  RatingRisco = "BB+"
	RatingBB      RatingRisco = "BB"
	RatingBPlus   RatingRisco = "B+"
	RatingB       RatingRisco = "B"
	RatingC       RatingRisco = "C"
)

// TransacaoToken representa uma transação de tokens
type TransacaoToken struct {
	ID              uint              `json:"id" gorm:"primaryKey"`
	TokenizacaoID   uint              `json:"tokenizacao_id" gorm:"not null;index"`
	VendedorID      *uint             `json:"vendedor_id" gorm:"index"`
	CompradorID     *uint             `json:"comprador_id" gorm:"index"`
	Quantidade      int               `json:"quantidade" gorm:"not null"`
	PrecoUnitario   float64           `json:"preco_unitario" gorm:"type:decimal(15,2);not null"`
	ValorTotal      float64           `json:"valor_total" gorm:"type:decimal(15,2);not null"`
	TipoTransacao   TipoTransacaoToken `json:"tipo_transacao" gorm:"size:50;not null"`
	HashBlockchain  string            `json:"hash_blockchain" gorm:"size:100;uniqueIndex"`
	Status          string            `json:"status" gorm:"size:20;default:'pendente'"`
	DataTransacao   time.Time         `json:"data_transacao"`
	CreatedAt       time.Time         `json:"created_at"`

	Tokenizacao *Tokenizacao `json:"tokenizacao,omitempty" gorm:"foreignKey:TokenizacaoID"`
	Vendedor    *User        `json:"vendedor,omitempty" gorm:"foreignKey:VendedorID"`
	Comprador   *User        `json:"comprador,omitempty" gorm:"foreignKey:CompradorID"`
}

// TipoTransacaoToken define os tipos de transação
type TipoTransacaoToken string

const (
	TipoTransacaoEmissao      TipoTransacaoToken = "emissao"
	TipoTransacaoVendaDireta  TipoTransacaoToken = "venda_direta"
	TipoTransacaoRecompra     TipoTransacaoToken = "recompra"
	TipoTransacaoTransferencia TipoTransacaoToken = "transferencia"
)

// ParticipacaoToken representa a participação de um investidor em uma tokenização
type ParticipacaoToken struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	TokenizacaoID   uint      `json:"tokenizacao_id" gorm:"not null;index"`
	InvestidorID    uint      `json:"investidor_id" gorm:"not null;index"`
	QuantidadeTokens int      `json:"quantidade_tokens" gorm:"not null"`
	PercentualTotal float64   `json:"percentual_total" gorm:"type:decimal(5,2)"`
	ValorInvestido  float64   `json:"valor_investido" gorm:"type:decimal(15,2)"`
	DataAquisicao   time.Time `json:"data_aquisicao"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	Tokenizacao *Tokenizacao `json:"tokenizacao,omitempty" gorm:"foreignKey:TokenizacaoID"`
	Investidor  *User        `json:"investidor,omitempty" gorm:"foreignKey:InvestidorID"`
}

// OfertaToken representa uma oferta de venda de tokens
type OfertaToken struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	TokenizacaoID    uint      `json:"tokenizacao_id" gorm:"not null;index"`
	VendedorID       uint      `json:"vendedor_id" gorm:"not null;index"`
	QuantidadeOfertada int     `json:"quantidade_ofertada" gorm:"not null"`
	PrecoUnitario    float64   `json:"preco_unitario" gorm:"type:decimal(15,2);not null"`
	Status           string    `json:"status" gorm:"size:20;default:'ativa'"`
	DataCriacao      time.Time `json:"data_criacao"`
	DataExpiracao    time.Time `json:"data_expiracao"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	Tokenizacao *Tokenizacao `json:"tokenizacao,omitempty" gorm:"foreignKey:TokenizacaoID"`
	Vendedor    *User        `json:"vendedor,omitempty" gorm:"foreignKey:VendedorID"`
}

// OrdemCompraToken representa uma ordem de compra de tokens
type OrdemCompraToken struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	TokenizacaoID uint      `json:"tokenizacao_id" gorm:"not null;index"`
	CompradorID   uint      `json:"comprador_id" gorm:"not null;index"`
	QuantidadeDesejada int  `json:"quantidade_desejada" gorm:"not null"`
	PrecoMaximo   float64   `json:"preco_maximo" gorm:"type:decimal(15,2);not null"`
	Status        string    `json:"status" gorm:"size:20;default:'pendente'"`
	DataCriacao   time.Time `json:"data_criacao"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Tokenizacao *Tokenizacao `json:"tokenizacao,omitempty" gorm:"foreignKey:TokenizacaoID"`
	Comprador   *User        `json:"comprador,omitempty" gorm:"foreignKey:CompradorID"`
}

// --- REQUEST/RESPONSE MODELS ---

// CreateTokenizacaoRequest representa a requisição para criar uma tokenização
type CreateTokenizacaoRequest struct {
	EquinoID                              uint       `json:"equinoid" validate:"required"`
	TotalTokens                           int        `json:"total_tokens" validate:"required,min=100"`
	PrecoInicialToken                     float64    `json:"preco_inicial_token" validate:"required,gt=0"`
	PercentualMinimoDono                  float64    `json:"percentual_minimo_dono" validate:"required,min=51,max=100"`
	PercentualComercializavelPublicamente float64    `json:"percentual_comercializavel_publicamente" validate:"required,min=0,max=49"`
	TravaControleDono                     bool       `json:"trava_controle_dono"`
	PrioridadeRecompra                    bool       `json:"prioridade_recompra"`
	CustoCustodiaMensal                   *float64   `json:"custo_custodia_mensal" validate:"omitempty,gt=0"`
	TemSeguro                             bool       `json:"tem_seguro"`
	ValorAssegurado                       *float64   `json:"valor_assegurado" validate:"omitempty,gt=0"`
	ApoliceSeguroURL                      string     `json:"apolice_seguro_url" validate:"omitempty,url"`
	GarantiasBiologicas                   []string   `json:"garantias_biologicas"`
	ObservacoesCompliance                 string     `json:"observacoes_compliance"`
}

// TokenizacaoResponse representa a resposta de tokenização
type TokenizacaoResponse struct {
	ID                                    uint        `json:"id"`
	Equinoid                              string      `json:"equinoid"`
	EquinoNome                            string      `json:"equino_nome"`
	TotalTokens                           int         `json:"total_tokens"`
	TokensBloqueadosDono                  int         `json:"tokens_bloqueados_dono"`
	TokensDisponiveisVenda                int         `json:"tokens_disponiveis_venda"`
	TokensVendidos                        int         `json:"tokens_vendidos"`
	PrecoInicialToken                     float64     `json:"preco_inicial_token"`
	PrecoAtualToken                       float64     `json:"preco_atual_token"`
	ValorTotalTokenizado                  float64     `json:"valor_total_tokenizado"`
	ValorMercadoAtual                     float64     `json:"valor_mercado_atual"`
	PercentualMinimoDono                  float64     `json:"percentual_minimo_dono"`
	PercentualComercializavelPublicamente float64     `json:"percentual_comercializavel_publicamente"`
	PercentualVendido                     float64     `json:"percentual_vendido"`
	TravaControleDono                     bool        `json:"trava_controle_dono"`
	PrioridadeRecompra                    bool        `json:"prioridade_recompra"`
	Status                                StatusToken `json:"status"`
	CustoCustodiaMensal                   float64     `json:"custo_custodia_mensal"`
	TemSeguro                             bool        `json:"tem_seguro"`
	ValorAssegurado                       *float64    `json:"valor_assegurado"`
	ApoliceSeguroURL                      string      `json:"apolice_seguro_url,omitempty"`
	GarantiasBiologicas                   []string    `json:"garantias_biologicas"`
	RatingRisco                           RatingRisco `json:"rating_risco"`
	NumeroInvestidores                    int         `json:"numero_investidores"`
	ROI                                   float64     `json:"roi"`
	DataInicio                            time.Time   `json:"data_inicio"`
	DataEncerramento                      *time.Time  `json:"data_encerramento,omitempty"`
	SmartContractAddress                  string      `json:"smart_contract_address,omitempty"`
	BlockchainNetwork                     string      `json:"blockchain_network,omitempty"`
	CreatedAt                             time.Time   `json:"created_at"`
	UpdatedAt                             time.Time   `json:"updated_at"`
}

// TransacaoTokenResponse representa a resposta de transação
type TransacaoTokenResponse struct {
	ID             uint               `json:"id"`
	TokenizacaoID  uint               `json:"tokenizacao_id"`
	VendedorID     *uint              `json:"vendedor_id"`
	VendedorNome   string             `json:"vendedor_nome,omitempty"`
	CompradorID    *uint              `json:"comprador_id"`
	CompradorNome  string             `json:"comprador_nome,omitempty"`
	Quantidade     int                `json:"quantidade"`
	PrecoUnitario  float64            `json:"preco_unitario"`
	ValorTotal     float64            `json:"valor_total"`
	TipoTransacao  TipoTransacaoToken `json:"tipo_transacao"`
	HashBlockchain string             `json:"hash_blockchain"`
	Status         string             `json:"status"`
	DataTransacao  time.Time          `json:"data_transacao"`
}

// OfertaTokenRequest representa a requisição para criar oferta
type OfertaTokenRequest struct {
	TokenizacaoID     uint    `json:"tokenizacao_id" validate:"required"`
	QuantidadeOfertada int    `json:"quantidade_ofertada" validate:"required,min=1"`
	PrecoUnitario     float64 `json:"preco_unitario" validate:"required,gt=0"`
	DiasValidade      int     `json:"dias_validade" validate:"required,min=1,max=30"`
}

// OrdemCompraTokenRequest representa a requisição para executar ordem de compra
type OrdemCompraTokenRequest struct {
	TokenizacaoID     uint    `json:"tokenizacao_id" validate:"required"`
	QuantidadeDesejada int    `json:"quantidade" validate:"required,min=1"`
	PrecoMaximo       float64 `json:"preco_maximo" validate:"required,gt=0"`
}

// ToResponse converte Tokenizacao para TokenizacaoResponse
func (t *Tokenizacao) ToResponse() *TokenizacaoResponse {
	response := &TokenizacaoResponse{
		ID:                                    t.ID,
		TotalTokens:                           t.TotalTokens,
		TokensBloqueadosDono:                  t.TokensBloqueadosDono,
		TokensDisponiveisVenda:                t.TokensDisponiveisVenda,
		TokensVendidos:                        t.TokensVendidos,
		PrecoInicialToken:                     t.PrecoInicialToken,
		ValorTotalTokenizado:                  t.ValorTotalTokenizado,
		PercentualMinimoDono:                  t.PercentualMinimoDono,
		PercentualComercializavelPublicamente: t.PercentualComercializavelPublicamente,
		TravaControleDono:                     t.TravaControleDono,
		PrioridadeRecompra:                    t.PrioridadeRecompra,
		Status:                                t.Status,
		CustoCustodiaMensal:                   t.CustoCustodiaMensal,
		TemSeguro:                             t.TemSeguro,
		ValorAssegurado:                       t.ValorAssegurado,
		ApoliceSeguroURL:                      t.ApoliceSeguroURL,
		RatingRisco:                           t.RatingRisco,
		DataInicio:                            t.DataInicio,
		DataEncerramento:                      t.DataEncerramento,
		SmartContractAddress:                  t.SmartContractAddress,
		BlockchainNetwork:                     t.BlockchainNetwork,
		CreatedAt:                             t.CreatedAt,
		UpdatedAt:                             t.UpdatedAt,
	}

	if t.Equino != nil {
		response.Equinoid = t.Equino.Equinoid
		response.EquinoNome = t.Equino.Nome
	}

	if t.GarantiasBiologicas != nil {
		if itens, ok := t.GarantiasBiologicas["itens"].([]interface{}); ok {
			response.GarantiasBiologicas = make([]string, len(itens))
			for i, g := range itens {
				if str, ok := g.(string); ok {
					response.GarantiasBiologicas[i] = str
				}
			}
		}
	}

	response.PercentualVendido = float64(t.TokensVendidos) / float64(t.TotalTokens) * 100
	response.PrecoAtualToken = t.PrecoInicialToken
	if t.TokensVendidos > 0 {
		response.PrecoAtualToken = t.PrecoInicialToken * 1.1
	}
	response.ValorMercadoAtual = response.PrecoAtualToken * float64(t.TotalTokens)
	response.ROI = ((response.PrecoAtualToken - t.PrecoInicialToken) / t.PrecoInicialToken) * 100

	return response
}

// ToResponse converte TransacaoToken para TransacaoTokenResponse
func (t *TransacaoToken) ToResponse() *TransacaoTokenResponse {
	response := &TransacaoTokenResponse{
		ID:             t.ID,
		TokenizacaoID:  t.TokenizacaoID,
		VendedorID:     t.VendedorID,
		CompradorID:    t.CompradorID,
		Quantidade:     t.Quantidade,
		PrecoUnitario:  t.PrecoUnitario,
		ValorTotal:     t.ValorTotal,
		TipoTransacao:  t.TipoTransacao,
		HashBlockchain: t.HashBlockchain,
		Status:         t.Status,
		DataTransacao:  t.DataTransacao,
	}

	if t.Vendedor != nil {
		response.VendedorNome = t.Vendedor.Name
	}
	if t.Comprador != nil {
		response.CompradorNome = t.Comprador.Name
	}

	return response
}

// TableName especifica o nome da tabela
func (Tokenizacao) TableName() string {
	return "tokenizacoes"
}

func (TransacaoToken) TableName() string {
	return "transacoes_tokens"
}

func (ParticipacaoToken) TableName() string {
	return "participacoes_tokens"
}

func (OfertaToken) TableName() string {
	return "ofertas_tokens"
}

func (OrdemCompraToken) TableName() string {
	return "ordens_compra_tokens"
}
