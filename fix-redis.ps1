Write-Host "=================================" -ForegroundColor Cyan
Write-Host "Equinoid - Fix Redis Connection" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Verificando Fly CLI..." -ForegroundColor Yellow

if (-not (Get-Command fly -ErrorAction SilentlyContinue)) {
    Write-Host ""
    Write-Host "Fly CLI nao encontrado!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Instalando Fly CLI..." -ForegroundColor Yellow
    
    try {
        iwr https://fly.io/install.ps1 -useb | iex
        Write-Host ""
        Write-Host "Instalacao concluida!" -ForegroundColor Green
        Write-Host ""
        Write-Host "IMPORTANTE: Feche e abra o PowerShell novamente, depois execute:" -ForegroundColor Yellow
        Write-Host "  .\fix-redis.ps1" -ForegroundColor White
        Write-Host ""
        exit 0
    } catch {
        Write-Host "Erro ao instalar Fly CLI: $_" -ForegroundColor Red
        Write-Host ""
        Write-Host "Instale manualmente:" -ForegroundColor Yellow
        Write-Host "  iwr https://fly.io/install.ps1 -useb | iex" -ForegroundColor White
        exit 1
    }
}

Write-Host "Fly CLI encontrado!" -ForegroundColor Green
Write-Host ""

Write-Host "Fazendo login no Fly.io..." -ForegroundColor Yellow
fly auth login

Write-Host ""
Write-Host "Verificando app backend..." -ForegroundColor Yellow

$appExists = fly apps list 2>&1 | Select-String "equinoid-backend"

if (-not $appExists) {
    Write-Host "App 'equinoid-backend' nao encontrado!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Criando app..." -ForegroundColor Yellow
    fly apps create equinoid-backend
}

Write-Host ""
Write-Host "Verificando secrets atuais..." -ForegroundColor Yellow
Write-Host ""

fly secrets list --app equinoid-backend

Write-Host ""
Write-Host "=================================" -ForegroundColor Cyan
Write-Host "Escolha uma opcao para Redis:" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. Criar Redis no Fly.io (recomendado, ~$2/mes)" -ForegroundColor White
Write-Host "2. Usar Upstash (gratis, requer cadastro)" -ForegroundColor White
Write-Host "3. Apenas configurar variaveis (Redis ja existe)" -ForegroundColor White
Write-Host ""

$opcao = Read-Host "Digite o numero da opcao"

switch ($opcao) {
    "1" {
        Write-Host ""
        Write-Host "Criando Redis no Fly.io..." -ForegroundColor Yellow
        
        $redisExists = fly apps list 2>&1 | Select-String "equinoid-redis"
        
        if (-not $redisExists) {
            fly apps create equinoid-redis
            fly volumes create redis_data --size 1 --region gru --app equinoid-redis
            Write-Host "Fazendo deploy do Redis..." -ForegroundColor Yellow
            fly deploy -c fly-redis.toml
        } else {
            Write-Host "Redis ja existe!" -ForegroundColor Green
        }
        
        Write-Host ""
        Write-Host "Configurando variaveis de ambiente..." -ForegroundColor Yellow
        
        fly secrets set REDIS_HOST="equinoid-redis.internal" REDIS_PORT="6379" --app equinoid-backend
    }
    
    "2" {
        Write-Host ""
        Write-Host "Configure o Upstash:" -ForegroundColor Yellow
        Write-Host "1. Acesse: https://upstash.com/" -ForegroundColor White
        Write-Host "2. Crie conta gratis" -ForegroundColor White
        Write-Host "3. Crie um Redis Database" -ForegroundColor White
        Write-Host "4. Copie as credenciais" -ForegroundColor White
        Write-Host ""
        
        $redisHost = Read-Host "Digite o REDIS_HOST (ex: abc-12345.upstash.io)"
        $redisPort = Read-Host "Digite o REDIS_PORT (geralmente 6379)"
        $redisPass = Read-Host "Digite o REDIS_PASSWORD"
        
        Write-Host ""
        Write-Host "Configurando variaveis..." -ForegroundColor Yellow
        
        fly secrets set REDIS_HOST="$redisHost" REDIS_PORT="$redisPort" REDIS_PASSWORD="$redisPass" --app equinoid-backend
    }
    
    "3" {
        Write-Host ""
        $redisHost = Read-Host "Digite o REDIS_HOST"
        $redisPort = Read-Host "Digite o REDIS_PORT (geralmente 6379)"
        
        Write-Host ""
        Write-Host "Configurando variaveis..." -ForegroundColor Yellow
        
        fly secrets set REDIS_HOST="$redisHost" REDIS_PORT="$redisPort" --app equinoid-backend
    }
    
    default {
        Write-Host ""
        Write-Host "Opcao invalida!" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "=================================" -ForegroundColor Cyan
Write-Host "Configuracao concluida!" -ForegroundColor Green
Write-Host "=================================" -ForegroundColor Cyan
Write-Host ""

$deploy = Read-Host "Fazer deploy agora? (y/n)"

if ($deploy -eq "y") {
    Write-Host ""
    Write-Host "Fazendo deploy..." -ForegroundColor Yellow
    fly deploy
    
    Write-Host ""
    Write-Host "Aguardando app iniciar..." -ForegroundColor Yellow
    Start-Sleep -Seconds 10
    
    Write-Host ""
    Write-Host "Logs do app:" -ForegroundColor Yellow
    Write-Host ""
    fly logs --app equinoid-backend
} else {
    Write-Host ""
    Write-Host "Para fazer deploy depois, execute:" -ForegroundColor Yellow
    Write-Host "  fly deploy" -ForegroundColor White
    Write-Host ""
    Write-Host "Para ver logs:" -ForegroundColor Yellow
    Write-Host "  fly logs --app equinoid-backend" -ForegroundColor White
}

Write-Host ""
Write-Host "Comandos uteis:" -ForegroundColor Yellow
Write-Host "  fly status --app equinoid-backend" -ForegroundColor White
Write-Host "  fly logs --app equinoid-backend" -ForegroundColor White
Write-Host "  fly secrets list --app equinoid-backend" -ForegroundColor White
Write-Host ""
