-- Script SQL para criar usuário admin inicial
-- Execute este script diretamente no banco de dados PostgreSQL

-- IMPORTANTE: Altere a senha antes de executar!
-- Use: SELECT crypt('SUA_SENHA_AQUI', gen_salt('bf', 12));
-- E substitua o hash abaixo

INSERT INTO users (
    email,
    password,
    name,
    user_type,
    cpf_cnpj,
    is_active,
    is_email_verified,
    role,
    created_at,
    updated_at
) VALUES (
    'admin@equinoid.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYqJZ5x5K5K', -- senha: admin123 (ALTERE!)
    'Administrador',
    'admin',
    '00000000000',
    true,
    true,
    'admin',
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Para gerar um novo hash de senha, execute no psql:
-- SELECT crypt('sua_senha_aqui', gen_salt('bf', 12));
