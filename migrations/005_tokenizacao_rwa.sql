-- Sprint 2 - Módulo de Tokenização RWA (Real World Assets)

-- Tabela de tokenizações
CREATE TABLE IF NOT EXISTS tokenizacoes (
    id SERIAL PRIMARY KEY,
    equino_id INTEGER NOT NULL UNIQUE REFERENCES equinos(id) ON DELETE CASCADE,
    total_tokens INTEGER NOT NULL CHECK (total_tokens >= 100),
    tokens_bloqueados_dono INTEGER NOT NULL,
    tokens_disponiveis_venda INTEGER NOT NULL,
    tokens_vendidos INTEGER NOT NULL DEFAULT 0,
    preco_inicial_token DECIMAL(15,2) NOT NULL CHECK (preco_inicial_token > 0),
    valor_total_tokenizado DECIMAL(15,2) NOT NULL,
    percentual_minimo_dono DECIMAL(5,2) NOT NULL CHECK (percentual_minimo_dono >= 51 AND percentual_minimo_dono <= 100),
    percentual_comercializavel_publicamente DECIMAL(5,2) NOT NULL CHECK (percentual_comercializavel_publicamente >= 0 AND percentual_comercializavel_publicamente <= 49),
    trava_controle_dono BOOLEAN NOT NULL DEFAULT TRUE,
    prioridade_recompra BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pendente' CHECK (status IN ('pendente', 'ativo', 'suspenso', 'encerrado')),
    custo_custodia_mensal DECIMAL(15,2) DEFAULT 0,
    tem_seguro BOOLEAN NOT NULL DEFAULT FALSE,
    valor_assegurado DECIMAL(15,2),
    apolice_seguro_url VARCHAR(500),
    garantias_biologicas JSONB,
    rating_risco VARCHAR(5) NOT NULL DEFAULT 'A' CHECK (rating_risco IN ('AAA+', 'AAA', 'AA+', 'AA', 'A+', 'A', 'BBB+', 'BBB', 'BB+', 'BB', 'B+', 'B', 'C')),
    data_inicio TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_encerramento TIMESTAMP,
    smart_contract_address VARCHAR(100),
    blockchain_network VARCHAR(50),
    observacoes_compliance TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_percentuais_validos CHECK (percentual_minimo_dono + percentual_comercializavel_publicamente <= 100),
    CONSTRAINT chk_tokens_consistency CHECK (tokens_bloqueados_dono + tokens_disponiveis_venda = total_tokens),
    CONSTRAINT chk_tokens_vendidos CHECK (tokens_vendidos <= tokens_disponiveis_venda),
    CONSTRAINT chk_seguro_valor CHECK (tem_seguro = FALSE OR valor_assegurado IS NOT NULL)
);

-- Índices para otimização
CREATE INDEX idx_tokenizacoes_equino_id ON tokenizacoes(equino_id);
CREATE INDEX idx_tokenizacoes_status ON tokenizacoes(status);
CREATE INDEX idx_tokenizacoes_rating ON tokenizacoes(rating_risco);
CREATE INDEX idx_tokenizacoes_deleted_at ON tokenizacoes(deleted_at);

-- Tabela de transações de tokens
CREATE TABLE IF NOT EXISTS transacoes_tokens (
    id SERIAL PRIMARY KEY,
    tokenizacao_id INTEGER NOT NULL REFERENCES tokenizacoes(id) ON DELETE CASCADE,
    vendedor_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    comprador_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    quantidade INTEGER NOT NULL CHECK (quantidade > 0),
    preco_unitario DECIMAL(15,2) NOT NULL CHECK (preco_unitario > 0),
    valor_total DECIMAL(15,2) NOT NULL,
    tipo_transacao VARCHAR(50) NOT NULL CHECK (tipo_transacao IN ('emissao', 'venda_direta', 'recompra', 'transferencia')),
    hash_blockchain VARCHAR(100) UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pendente' CHECK (status IN ('pendente', 'confirmado', 'cancelado', 'falha')),
    data_transacao TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_transacoes_tokenizacao_id (tokenizacao_id),
    INDEX idx_transacoes_vendedor_id (vendedor_id),
    INDEX idx_transacoes_comprador_id (comprador_id),
    INDEX idx_transacoes_data (data_transacao),
    INDEX idx_transacoes_hash (hash_blockchain)
);

-- Tabela de participações de investidores
CREATE TABLE IF NOT EXISTS participacoes_tokens (
    id SERIAL PRIMARY KEY,
    tokenizacao_id INTEGER NOT NULL REFERENCES tokenizacoes(id) ON DELETE CASCADE,
    investidor_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quantidade_tokens INTEGER NOT NULL CHECK (quantidade_tokens > 0),
    percentual_total DECIMAL(5,2) NOT NULL,
    valor_investido DECIMAL(15,2) NOT NULL,
    data_aquisicao TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(tokenizacao_id, investidor_id),
    INDEX idx_participacoes_tokenizacao_id (tokenizacao_id),
    INDEX idx_participacoes_investidor_id (investidor_id)
);

-- Tabela de ofertas de venda
CREATE TABLE IF NOT EXISTS ofertas_tokens (
    id SERIAL PRIMARY KEY,
    tokenizacao_id INTEGER NOT NULL REFERENCES tokenizacoes(id) ON DELETE CASCADE,
    vendedor_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quantidade_ofertada INTEGER NOT NULL CHECK (quantidade_ofertada > 0),
    preco_unitario DECIMAL(15,2) NOT NULL CHECK (preco_unitario > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'ativa' CHECK (status IN ('ativa', 'executada', 'cancelada', 'expirada')),
    data_criacao TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_expiracao TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_ofertas_tokenizacao_id (tokenizacao_id),
    INDEX idx_ofertas_vendedor_id (vendedor_id),
    INDEX idx_ofertas_status (status),
    INDEX idx_ofertas_expiracao (data_expiracao)
);

-- Tabela de ordens de compra
CREATE TABLE IF NOT EXISTS ordens_compra_tokens (
    id SERIAL PRIMARY KEY,
    tokenizacao_id INTEGER NOT NULL REFERENCES tokenizacoes(id) ON DELETE CASCADE,
    comprador_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quantidade_desejada INTEGER NOT NULL CHECK (quantidade_desejada > 0),
    preco_maximo DECIMAL(15,2) NOT NULL CHECK (preco_maximo > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'pendente' CHECK (status IN ('pendente', 'executada', 'cancelada', 'parcial')),
    data_criacao TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_ordens_tokenizacao_id (tokenizacao_id),
    INDEX idx_ordens_comprador_id (comprador_id),
    INDEX idx_ordens_status (status)
);

-- Comentários detalhados
COMMENT ON TABLE tokenizacoes IS 'Tokenização de equinos como RWA (Real World Assets) - ativos tokenizados com compliance regulatório';
COMMENT ON COLUMN tokenizacoes.percentual_minimo_dono IS 'Percentual mínimo obrigatório do proprietário original (compliance: mínimo 51%)';
COMMENT ON COLUMN tokenizacoes.trava_controle_dono IS 'Impede venda que reduza participação do dono abaixo do mínimo regulatório';
COMMENT ON COLUMN tokenizacoes.prioridade_recompra IS 'Dono tem prioridade na recompra de tokens em ofertas';
COMMENT ON COLUMN tokenizacoes.garantias_biologicas IS 'Garantias biológicas (sêmen congelado, embriões, etc) em formato JSON';
COMMENT ON COLUMN tokenizacoes.rating_risco IS 'Rating de risco do ativo: AAA+ (melhor) até C (maior risco)';
COMMENT ON COLUMN tokenizacoes.smart_contract_address IS 'Endereço do smart contract na blockchain (quando implementado)';

COMMENT ON TABLE transacoes_tokens IS 'Histórico completo de transações de tokens com rastreabilidade blockchain';
COMMENT ON COLUMN transacoes_tokens.hash_blockchain IS 'Hash único da transação para auditoria e rastreabilidade';
COMMENT ON COLUMN transacoes_tokens.tipo_transacao IS 'Tipos: emissao (criação inicial), venda_direta, recompra, transferencia';

COMMENT ON TABLE participacoes_tokens IS 'Participação atual de cada investidor em cada tokenização';
COMMENT ON COLUMN participacoes_tokens.percentual_total IS 'Percentual de propriedade do investidor nesta tokenização';

COMMENT ON TABLE ofertas_tokens IS 'Ofertas de venda de tokens por investidores atuais';
COMMENT ON TABLE ordens_compra_tokens IS 'Ordens de compra pendentes de novos investidores';

-- Trigger para atualizar updated_at automaticamente
CREATE OR REPLACE FUNCTION update_tokenizacao_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_tokenizacoes_updated_at
    BEFORE UPDATE ON tokenizacoes
    FOR EACH ROW
    EXECUTE FUNCTION update_tokenizacao_updated_at();

CREATE TRIGGER trigger_participacoes_tokens_updated_at
    BEFORE UPDATE ON participacoes_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_tokenizacao_updated_at();

CREATE TRIGGER trigger_ofertas_tokens_updated_at
    BEFORE UPDATE ON ofertas_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_tokenizacao_updated_at();

CREATE TRIGGER trigger_ordens_compra_tokens_updated_at
    BEFORE UPDATE ON ordens_compra_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_tokenizacao_updated_at();
