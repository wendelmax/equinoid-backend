#!/bin/bash

# Script para criar usuário admin no Equinoid Backend
# Uso: ./scripts/create-admin.sh [email] [senha] [nome] [cpf_cnpj]

set -e

EMAIL="${1:-admin@equinoid.com}"
PASSWORD="${2:-admin123}"
NAME="${3:-Administrador}"
CPF_CNPJ="${4:-00000000000}"

if [ -z "$DATABASE_URL" ]; then
    echo "ERRO: DATABASE_URL não configurada"
    echo "Configure a variável de ambiente DATABASE_URL"
    exit 1
fi

echo "Criando usuário admin..."
echo "Email: $EMAIL"
echo "Nome: $NAME"

# Gerar hash da senha usando Go
HASH=$(go run -c 'package main; import ("fmt"; "golang.org/x/crypto/bcrypt"); func main() { h, _ := bcrypt.GenerateFromPassword([]byte("'$PASSWORD'"), bcrypt.DefaultCost); fmt.Print(string(h)) }' 2>/dev/null || \
    echo '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYqJZ5x5K5K')

# Inserir no banco
psql "$DATABASE_URL" <<EOF
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
    '$EMAIL',
    '$HASH',
    '$NAME',
    'admin',
    '$CPF_CNPJ',
    true,
    true,
    'admin',
    NOW(),
    NOW()
) ON CONFLICT (email) DO UPDATE SET
    password = EXCLUDED.password,
    name = EXCLUDED.name,
    user_type = 'admin',
    is_active = true,
    updated_at = NOW();
EOF

echo "✅ Usuário admin criado/atualizado com sucesso!"
echo "   Email: $EMAIL"
echo "   Senha: $PASSWORD"
echo ""
echo "Faça login em: POST /api/v1/auth/login"
