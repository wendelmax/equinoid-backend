# COMO RESOLVER O ERRO DO REDIS

## Problema

```
redis connection failed: dial tcp [::1]:6379: connect: connection refused
```

## Soluções Disponíveis

### Opção 1: Rodar SEM Redis (Mais Rápido)

Se você não precisa de cache agora, desabilite o Redis:

```powershell
fly secrets set REDIS_ENABLED=false --app equinoid-backend
fly deploy
```

Veja [SEM-REDIS.md](SEM-REDIS.md) para mais detalhes.

### Opção 2: Configurar Redis (Recomendado para Produção)

Execute este comando no PowerShell:

```powershell
.\fix-redis.ps1
```

Este script vai:
1. Verificar se o Fly CLI está instalado (instalar se necessário)
2. Fazer login no Fly.io
3. Perguntar como você quer configurar o Redis
4. Configurar as variáveis de ambiente automaticamente
5. Fazer deploy (opcional)

## Passo a Passo Manual

Se preferir fazer manualmente:

### 1. Instalar Fly CLI

```powershell
iwr https://fly.io/install.ps1 -useb | iex
```

**IMPORTANTE:** Feche e abra o PowerShell depois da instalação!

### 2. Fazer Login

```powershell
fly auth login
```

### 3. Configurar Redis

Escolha UMA das opções:

**Opção A: Criar Redis no Fly.io**

```powershell
fly apps create equinoid-redis
fly volumes create redis_data --size 1 --region gru --app equinoid-redis
fly deploy -c fly-redis.toml
fly secrets set REDIS_HOST="equinoid-redis.internal" REDIS_PORT="6379" --app equinoid-backend
```

**Opção B: Usar Upstash (Grátis)**

1. Acesse https://upstash.com/
2. Crie conta e um Redis Database
3. Configure:

```powershell
fly secrets set REDIS_HOST="seu-endpoint.upstash.io" REDIS_PORT="6379" REDIS_PASSWORD="sua-senha" --app equinoid-backend
```

### 4. Deploy

```powershell
fly deploy
```

### 5. Verificar

```powershell
fly logs --app equinoid-backend
```

Você deve ver:
```
✅ Server started on :8080
✅ Connected to Redis
```

## Por Que Este Erro Aconteceu?

O código em `internal/config/config.go` usa defaults quando as variáveis de ambiente não estão configuradas:

```go
func buildRedisURL() string {
    host := getEnv("REDIS_HOST", "localhost")  // <- Default: localhost
    port := getEnv("REDIS_PORT", "6379")
    // ...
}
```

Como você não tinha `REDIS_HOST` configurado no Fly.io, a aplicação tentou conectar em `localhost`, que não existe no container.

## Arquivos de Ajuda

- `fix-redis.ps1` - Script automatizado
- `CONFIGURE_REDIS.md` - Guia detalhado
- `DEPLOY_GUIDE.md` - Guia completo de deploy
- `fly.toml` - Configuração do backend
- `fly-redis.toml` - Configuração do Redis

## Precisa de Ajuda?

Execute e me mostre a saída:

```powershell
fly apps list
fly secrets list --app equinoid-backend
fly logs --app equinoid-backend
```
