# Como Rodar Sem Redis

## Solução Rápida

Defina a variável de ambiente `REDIS_ENABLED=false` e a aplicação rodará SEM Redis.

## Fly.io

Configure a variável:

```powershell
fly secrets set REDIS_ENABLED=false --app equinoid-backend
```

Depois faça deploy:

```powershell
fly deploy
```

## Railway

No dashboard Railway:
1. Vá em "Variables"
2. Adicione: `REDIS_ENABLED = false`
3. O deploy será feito automaticamente

## Render

No dashboard Render:
1. Vá em "Environment"
2. Adicione: `REDIS_ENABLED = false`
3. Clique em "Save Changes"

## Docker Local

Edite o arquivo `.env`:

```env
REDIS_ENABLED=false
```

E rode apenas a API (sem Redis):

```powershell
docker build -t equinoid-api .
docker run -p 8080:8080 --env-file .env equinoid-api
```

## Desenvolvimento Local

Execute com a variável:

```powershell
$env:REDIS_ENABLED="false"
go run cmd/server/main.go
```

Ou no Linux/Mac:

```bash
REDIS_ENABLED=false go run cmd/server/main.go
```

## Verificação

Após iniciar a aplicação, acesse:

```
http://localhost:8080/health
```

Você verá:

```json
{
  "status": "healthy",
  "services": {
    "database": "healthy",
    "redis": "disabled"
  }
}
```

## O Que Acontece Sem Redis?

Quando Redis está desabilitado:

- ✅ Aplicação inicia normalmente
- ✅ Todas as funcionalidades funcionam
- ❌ Cache de dados não estará disponível
- ❌ Performance pode ser menor em endpoints com cache
- ❌ Rate limiting pode não funcionar corretamente

## Impacto na Performance

Sem cache Redis, algumas operações podem ser mais lentas:

- Listagens de equinos (vai consultar o banco toda vez)
- Busca de usuários
- Simulador genético (cálculos não são cacheados)
- Rankings e estatísticas

## Quando Usar Sem Redis?

**Use sem Redis para:**
- Desenvolvimento local rápido
- Testes simples
- Ambientes de demonstração
- POCs (Proof of Concept)

**Use COM Redis para:**
- Produção
- Staging
- Testes de carga
- Quando performance é crítica

## Como Adicionar Redis Depois

Se quiser adicionar Redis depois:

1. **Opção 1: Upstash (Grátis)**
   - Acesse https://upstash.com
   - Crie um Redis Database
   - Configure:

```powershell
fly secrets set REDIS_ENABLED=true --app equinoid-backend
fly secrets set REDIS_HOST="seu-endpoint.upstash.io" --app equinoid-backend
fly secrets set REDIS_PASSWORD="sua-senha" --app equinoid-backend
```

2. **Opção 2: Fly.io Redis**

```powershell
fly apps create equinoid-redis
fly volumes create redis_data --size 1 --region gru --app equinoid-redis
fly deploy -c fly-redis.toml

fly secrets set REDIS_ENABLED=true --app equinoid-backend
fly secrets set REDIS_HOST="equinoid-redis.internal" --app equinoid-backend
```

## Troubleshooting

### Erro mesmo com REDIS_ENABLED=false

Verifique se a variável está realmente configurada:

```powershell
fly secrets list --app equinoid-backend
```

### Como verificar se Redis está desabilitado

Nos logs, você verá:

```
INFO Redis is disabled. Running without cache.
```

Em vez de:

```
INFO Redis connected successfully
```

## Resumo

Para rodar **SEM Redis** agora:

```powershell
fly secrets set REDIS_ENABLED=false --app equinoid-backend
fly deploy
```

Pronto! A aplicação rodará normalmente sem precisar de Redis.
