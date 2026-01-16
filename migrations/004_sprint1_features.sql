-- Sprint 1 Features: Participação em Eventos e Ultrassonografias

-- Tabela de participações em eventos
CREATE TABLE IF NOT EXISTS participacoes_eventos (
    id SERIAL PRIMARY KEY,
    evento_id INTEGER NOT NULL REFERENCES eventos(id) ON DELETE CASCADE,
    equino_id INTEGER NOT NULL REFERENCES equinos(id) ON DELETE CASCADE,
    participante_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    particularidades TEXT,
    resultado VARCHAR(100),
    classificacao INTEGER,
    compareceu BOOLEAN,
    penalizacao_ausencia INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    INDEX idx_participacoes_evento_id (evento_id),
    INDEX idx_participacoes_equino_id (equino_id),
    INDEX idx_participacoes_participante_id (participante_id),
    INDEX idx_participacoes_deleted_at (deleted_at)
);

-- Tabela de ultrassonografias (se não existir)
CREATE TABLE IF NOT EXISTS ultrassonografias (
    id SERIAL PRIMARY KEY,
    gestacao_id INTEGER NOT NULL REFERENCES gestacoes(id) ON DELETE CASCADE,
    data_exame TIMESTAMP NOT NULL,
    idade_gestacional INTEGER,
    veterinario_responsavel INTEGER NOT NULL REFERENCES users(id),
    presenca_embriao BOOLEAN,
    numero_embrioes INTEGER,
    batimento_cardiaco BOOLEAN,
    desenvolvimento_normal BOOLEAN,
    tamanho_embriao DECIMAL(8,2),
    frequencia_cardiaca INTEGER,
    diagnostico TEXT,
    observacoes TEXT,
    proximo_exame TIMESTAMP,
    documentos JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    INDEX idx_ultrassonografias_gestacao_id (gestacao_id),
    INDEX idx_ultrassonografias_data_exame (data_exame),
    INDEX idx_ultrassonografias_deleted_at (deleted_at)
);

-- Comentários
COMMENT ON TABLE participacoes_eventos IS 'Participações de equinos em eventos competitivos';
COMMENT ON COLUMN participacoes_eventos.penalizacao_ausencia IS 'Pontos de penalização por ausência no evento (-50 padrão)';
COMMENT ON COLUMN participacoes_eventos.compareceu IS 'Indica se o equino compareceu ao evento';

COMMENT ON TABLE ultrassonografias IS 'Exames de ultrassonografia durante gestação';
COMMENT ON COLUMN ultrassonografias.idade_gestacional IS 'Idade gestacional em dias';
COMMENT ON COLUMN ultrassonografias.tamanho_embriao IS 'Tamanho do embrião/feto em centímetros';
