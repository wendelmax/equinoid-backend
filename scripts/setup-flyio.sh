#!/bin/bash

set -e

echo "🚀 Equinoid Backend - Setup Fly.io"
echo "===================================="
echo ""

if ! command -v fly &> /dev/null; then
    echo "❌ Fly CLI não encontrado. Instale com:"
    echo "   curl -L https://fly.io/install.sh | sh"
    exit 1
fi

echo "📝 Fazendo login no Fly.io..."
fly auth login

echo ""
echo "🗄️  Configurando Redis..."
echo ""

read -p "Criar app Redis? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    fly apps create equinoid-redis || echo "App já existe"
    fly volumes create redis_data --size 1 --region gru --app equinoid-redis || echo "Volume já existe"
    fly deploy -c fly-redis.toml
fi

echo ""
echo "🐴 Configurando API Backend..."
echo ""

read -p "Criar app Backend? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    fly apps create equinoid-backend || echo "App já existe"
fi

echo ""
echo "🔐 Configurando secrets..."
echo ""

fly secrets set \
  DATABASE_URL="postgresql://postgres:M4EYIU4ne9j5JIId@db.rqaemzdqntwuomycewrn.supabase.co:5432/postgres" \
  SUPABASE_URL="https://rqaemzdqntwuomycewrn.supabase.co" \
  SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InJxYWVtemRxbnR3dW9teWNld3JuIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjAyMDM0ODgsImV4cCI6MjA3NTc3OTQ4OH0.wfXXKFZxVMnF4TZmX4_ZuAuaNNVA6MiZQyCNLy0rhzQ" \
  SUPABASE_SERVICE_ROLE_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InJxYWVtemRxbnR3dW9teWNld3JuIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc2MDIwMzQ4OCwiZXhwIjoyMDc1Nzc5NDg4fQ.6DYc_ZDGkShqPgaPKY5u5vPURn0LtrQgIt54aAoXBrE" \
  JWT_SECRET="equinoid-fallback-secret" \
  REDIS_HOST="equinoid-redis.internal" \
  REDIS_PORT="6379" \
  --app equinoid-backend

echo ""
echo "🚀 Fazendo deploy..."
echo ""

fly deploy

echo ""
echo "✅ Deploy concluído!"
echo ""
echo "📊 Status: fly status"
echo "📋 Logs: fly logs"
echo "🌐 URL: https://equinoid-backend.fly.dev"
echo ""
