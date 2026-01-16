-- Migration: 003_extended_modules
-- Description: Add tables for Finance, RWA, Nutrition, Training, and Marketplace
-- Author: Antigravity
-- Date: 2025-12-21

-- Create schemas if not already present
CREATE SCHEMA IF NOT EXISTS financeiro;
CREATE SCHEMA IF NOT EXISTS tokenizacao;
CREATE SCHEMA IF NOT EXISTS nutricao;
CREATE SCHEMA IF NOT EXISTS treinamento;
CREATE SCHEMA IF NOT EXISTS marketplace;

-- ============================================================================
-- SCHEMA: financeiro
-- ============================================================================

CREATE TABLE IF NOT EXISTS financeiro.transacoes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users.users(id),
    tipo VARCHAR(10) NOT NULL, -- 'receita', 'despesa'
    categoria VARCHAR(100) NOT NULL,
    titulo VARCHAR(200) NOT NULL,
    descricao TEXT,
    data TIMESTAMP NOT NULL DEFAULT NOW(),
    valor DECIMAL(15, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'confirmado',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_finance_user ON financeiro.transacoes(user_id);
CREATE INDEX idx_finance_tipo ON financeiro.transacoes(tipo);
CREATE INDEX idx_finance_date ON financeiro.transacoes(data);

-- ============================================================================
-- SCHEMA: tokenizacao (RWA)
-- ============================================================================

CREATE TABLE IF NOT EXISTS tokenizacao.ativos (
    id SERIAL PRIMARY KEY,
    equinoid VARCHAR(25) NOT NULL REFERENCES equinos.equinos(equinoid_id),
    total_tokens INTEGER NOT NULL,
    tokens_bloqueados_dono INTEGER NOT NULL,
    tokens_disponiveis_venda INTEGER NOT NULL,
    tokens_vendidos INTEGER DEFAULT 0,
    preco_inicial_token DECIMAL(15, 2) NOT NULL,
    valor_total_tokenizado DECIMAL(15, 2) NOT NULL,
    percentual_minimo_dono DECIMAL(5, 2) DEFAULT 51.0,
    status VARCHAR(20) DEFAULT 'ativo',
    custo_custodia_mensal DECIMAL(15, 2) DEFAULT 0.0,
    tem_seguro BOOLEAN DEFAULT FALSE,
    valor_assegurado DECIMAL(15, 2),
    garantias_biologicas JSONB, -- Array of strings
    rating_risco VARCHAR(10),
    data_inicio TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tokenizacao.transacoes_tokens (
    id SERIAL PRIMARY KEY,
    tokenizacao_id INTEGER NOT NULL REFERENCES tokenizacao.ativos(id),
    comprador_id INTEGER REFERENCES users.users(id),
    vendedor_id INTEGER REFERENCES users.users(id),
    quantidade INTEGER NOT NULL,
    preco_unitario DECIMAL(15, 2) NOT NULL,
    tipo_transacao VARCHAR(50) NOT NULL, -- 'venda_direta', 'oferta_publica'
    hash_blockchain VARCHAR(255),
    data_transacao TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_token_equino ON tokenizacao.ativos(equinoid);
CREATE INDEX idx_token_tx_token ON tokenizacao.transacoes_tokens(tokenizacao_id);

-- ============================================================================
-- SCHEMA: nutricao
-- ============================================================================

CREATE TABLE IF NOT EXISTS nutricao.planos_alimentares (
    id SERIAL PRIMARY KEY,
    equinoid VARCHAR(25) NOT NULL REFERENCES equinos.equinos(equinoid_id),
    titulo VARCHAR(200) NOT NULL,
    descricao TEXT,
    data_inicio DATE NOT NULL,
    data_fim DATE,
    status VARCHAR(20) DEFAULT 'ativo',
    criado_por INTEGER REFERENCES users.users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nutricao.itens_dieta (
    id SERIAL PRIMARY KEY,
    plano_id INTEGER NOT NULL REFERENCES nutricao.planos_alimentares(id) ON DELETE CASCADE,
    alimento VARCHAR(200) NOT NULL,
    quantidade VARCHAR(50) NOT NULL,
    frequencia VARCHAR(100) NOT NULL,
    observacoes TEXT
);

-- ============================================================================
-- SCHEMA: treinamento
-- ============================================================================

CREATE TABLE IF NOT EXISTS treinamento.sessoes (
    id SERIAL PRIMARY KEY,
    equinoid VARCHAR(25) NOT NULL REFERENCES equinos.equinos(equinoid_id),
    tipo VARCHAR(100) NOT NULL,
    data TIMESTAMP NOT NULL DEFAULT NOW(),
    duracao_minutos INTEGER,
    intensidade VARCHAR(20), -- 'baixa', 'media', 'alta'
    treinador VARCHAR(200),
    observacoes TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ============================================================================
-- SCHEMA: marketplace
-- ============================================================================

CREATE TABLE IF NOT EXISTS marketplace.anuncios (
    id SERIAL PRIMARY KEY,
    usuario_id INTEGER NOT NULL REFERENCES users.users(id),
    tipo VARCHAR(50) NOT NULL, -- 'animal', 'semen', 'embrio', 'equipamento'
    titulo VARCHAR(200) NOT NULL,
    descricao TEXT,
    preco DECIMAL(15, 2) NOT NULL,
    equinoid VARCHAR(25) REFERENCES equinos.equinos(equinoid_id),
    fotos JSONB,
    status VARCHAR(20) DEFAULT 'ativo', -- 'ativo', 'vendido', 'pausado'
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Add updated_at triggers
CREATE TRIGGER update_finance_transacoes_updated_at BEFORE UPDATE ON financeiro.transacoes
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_token_ativos_updated_at BEFORE UPDATE ON tokenizacao.ativos
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_nutricao_planos_updated_at BEFORE UPDATE ON nutricao.planos_alimentares
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_marketplace_anuncios_updated_at BEFORE UPDATE ON marketplace.anuncios
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

-- Grants
GRANT USAGE ON SCHEMA financeiro TO equinoid_app;
GRANT USAGE ON SCHEMA tokenizacao TO equinoid_app;
GRANT USAGE ON SCHEMA nutricao TO equinoid_app;
GRANT USAGE ON SCHEMA treinamento TO equinoid_app;
GRANT USAGE ON SCHEMA marketplace TO equinoid_app;

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA financeiro TO equinoid_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA tokenizacao TO equinoid_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA nutricao TO equinoid_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA treinamento TO equinoid_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA marketplace TO equinoid_app;
