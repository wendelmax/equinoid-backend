-- Migration: Add Keycloak support to users table
-- Date: 2025-10-12
-- Description: Adds keycloak_sub, role, and email_verified fields for Keycloak integration

-- Add keycloak_sub column (unique identifier from Keycloak)
ALTER TABLE users ADD COLUMN IF NOT EXISTS keycloak_sub VARCHAR(64);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_keycloak_sub ON users(keycloak_sub) WHERE keycloak_sub IS NOT NULL;

-- Add role column (for authorization)
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'usuario';

-- Add email_verified column (from OAuth providers)
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT false;

-- Update existing users to have role if null
UPDATE users SET role = 'usuario' WHERE role IS NULL;

-- Validation
SELECT 
    'Keycloak fields added:' as status,
    COUNT(*) FILTER (WHERE keycloak_sub IS NOT NULL) as users_with_keycloak,
    COUNT(*) FILTER (WHERE role IS NOT NULL) as users_with_role,
    COUNT(*) FILTER (WHERE email_verified IS NOT NULL) as users_with_email_verified
FROM users;

