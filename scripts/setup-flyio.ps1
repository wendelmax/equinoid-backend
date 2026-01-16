Write-Host "🚀 Equinoid Backend - Setup Fly.io" -ForegroundColor Green
Write-Host "====================================" -ForegroundColor Green
Write-Host ""

if (-not (Get-Command fly -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Fly CLI não encontrado. Instale com:" -ForegroundColor Red
    Write-Host "   iwr https://fly.io/install.ps1 -useb | iex" -ForegroundColor Yellow
    exit 1
}

Write-Host "📝 Fazendo login no Fly.io..." -ForegroundColor Cyan
fly auth login

Write-Host ""
Write-Host "🗄️  Configurando Redis..." -ForegroundColor Cyan
Write-Host ""

$createRedis = Read-Host "Criar app Redis? (y/n)"
if ($createRedis -eq 'y') {
    fly apps create equinoid-redis
    fly volumes create redis_data --size 1 --region gru --app equinoid-redis
    fly deploy -c fly-redis.toml
}

Write-Host ""
Write-Host "🐴 Configurando API Backend..." -ForegroundColor Cyan
Write-Host ""

$createBackend = Read-Host "Criar app Backend? (y/n)"
if ($createBackend -eq 'y') {
    fly apps create equinoid-backend
}

Write-Host ""
Write-Host "🔐 Configurando secrets..." -ForegroundColor Cyan
Write-Host ""

fly secrets set `
  DATABASE_URL="postgresql://postgres:M4EYIU4ne9j5JIId@db.rqaemzdqntwuomycewrn.supabase.co:5432/postgres" `
  SUPABASE_URL="https://rqaemzdqntwuomycewrn.supabase.co" `
  SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InJxYWVtemRxbnR3dW9teWNld3JuIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjAyMDM0ODgsImV4cCI6MjA3NTc3OTQ4OH0.wfXXKFZxVMnF4TZmX4_ZuAuaNNVA6MiZQyCNLy0rhzQ" `
  SUPABASE_SERVICE_ROLE_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InJxYWVtemRxbnR3dW9teWNld3JuIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc2MDIwMzQ4OCwiZXhwIjoyMDc1Nzc5NDg4fQ.6DYc_ZDGkShqPgaPKY5u5vPURn0LtrQgIt54aAoXBrE" `
  JWT_SECRET="equinoid-fallback-secret" `
  REDIS_HOST="equinoid-redis.internal" `
  REDIS_PORT="6379" `
  --app equinoid-backend

Write-Host ""
Write-Host "🚀 Fazendo deploy..." -ForegroundColor Cyan
Write-Host ""

fly deploy

Write-Host ""
Write-Host "✅ Deploy concluído!" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Status: fly status" -ForegroundColor Yellow
Write-Host "📋 Logs: fly logs" -ForegroundColor Yellow
Write-Host "🌐 URL: https://equinoid-backend.fly.dev" -ForegroundColor Yellow
Write-Host ""
