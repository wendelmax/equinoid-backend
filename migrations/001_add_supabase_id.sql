-- Migration: Add Supabase ID support
-- Description: Adiciona campo supabase_id para integração com Supabase Auth
-- Date: 2025-10-11

-- Adicionar coluna supabase_id na tabela users
ALTER TABLE users ADD COLUMN IF NOT EXISTS supabase_id VARCHAR(36) UNIQUE;

-- Criar índice para performance
CREATE INDEX IF NOT EXISTS idx_users_supabase_id ON users(supabase_id);

-- Permitir password_hash NULL (usuários Supabase não têm senha local)
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- Comentários
COMMENT ON COLUMN users.supabase_id IS 'Supabase user UUID for OAuth/SSO authentication';
COMMENT ON INDEX idx_users_supabase_id IS 'Index for fast Supabase user lookups';

-- Rollback script (comentado)
-- ALTER TABLE users DROP COLUMN IF EXISTS supabase_id;
-- DROP INDEX IF EXISTS idx_users_supabase_id;
-- ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;

