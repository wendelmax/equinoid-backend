# Guia de Deploy - Equinoid Backend

## Fly.io

### 1. Instalar CLI

```bash
# Windows (PowerShell)
iwr https://fly.io/install.ps1 -useb | iex

# macOS/Linux
curl -L https://fly.io/install.sh | sh
```

### 2. Login

```bash
fly auth login
```

### 3. Deploy Redis

```bash
fly apps create equinoid-redis

fly volumes create redis_data --size 1 --region gru --app equinoid-redis

fly deploy -c fly-redis.toml
```

### 4. Configurar Secrets da API

```bash
fly apps create equinoid-backend

fly secrets set \
  DATABASE_URL="postgresql://postgres:M4EYIU4ne9j5JIId@db.rqaemzdqntwuomycewrn.supabase.co:5432/postgres" \
  SUPABASE_URL="https://rqaemzdqntwuomycewrn.supabase.co" \
  SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InJxYWVtemRxbnR3dW9teWNld3JuIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjAyMDM0ODgsImV4cCI6MjA3NTc3OTQ4OH0.wfXXKFZxVMnF4TZmX4_ZuAuaNNVA6MiZQyCNLy0rhzQ" \
  SUPABASE_SERVICE_ROLE_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InJxYWVtemRxbnR3dW9teWNld3JuIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc2MDIwMzQ4OCwiZXhwIjoyMDc1Nzc5NDg4fQ.6DYc_ZDGkShqPgaPKY5u5vPURn0LtrQgIt54aAoXBrE" \
  JWT_SECRET="equinoid-fallback-secret" \
  REDIS_HOST="equinoid-redis.internal" \
  REDIS_PORT="6379" \
  --app equinoid-backend
```

### 5. Deploy API

```bash
fly deploy
```

### 6. Verificar Status

```bash
fly status
fly logs
```

---

## Railway

### 1. Instalar CLI

```bash
npm i -g @railway/cli
```

### 2. Login

```bash
railway login
```

### 3. Criar Projeto

```bash
railway init
```

### 4. Adicionar Redis

No dashboard Railway:
1. Clique em "New Service"
2. Selecione "Redis"
3. Copie a URL de conexão

### 5. Configurar Variáveis

No dashboard Railway ou via CLI:

```bash
railway variables set DATABASE_URL="postgresql://..."
railway variables set SUPABASE_URL="https://..."
railway variables set SUPABASE_ANON_KEY="..."
railway variables set SUPABASE_SERVICE_ROLE_KEY="..."
railway variables set JWT_SECRET="equinoid-fallback-secret"
railway variables set REDIS_HOST="redis.railway.internal"
railway variables set REDIS_PORT="6379"
```

### 6. Deploy

```bash
railway up
```

---

## Render

### 1. Criar Web Service

1. Acesse render.com
2. New > Web Service
3. Conecte seu repositório
4. Configure:
   - **Build Command**: `go build -o bin/equinoid-api cmd/server/main.go`
   - **Start Command**: `./bin/equinoid-api`
   - **Environment**: Go

### 2. Adicionar Redis

1. New > Redis
2. Copie o Internal Redis URL

### 3. Configurar Environment Variables

No dashboard do Web Service:

```
DATABASE_URL = postgresql://...
SUPABASE_URL = https://...
SUPABASE_ANON_KEY = ...
SUPABASE_SERVICE_ROLE_KEY = ...
JWT_SECRET = equinoid-fallback-secret
REDIS_URL = redis://...
PORT = 8080
GIN_MODE = release
ENVIRONMENT = production
```

### 4. Deploy

Render fará deploy automaticamente ao detectar mudanças no Git.

---

## Troubleshooting

### Erro: "redis connection failed"

**Causa**: Redis não está configurado ou variáveis incorretas

**Solução**:
1. Verifique se o Redis está rodando:
   ```bash
   fly status --app equinoid-redis
   ```

2. Verifique as variáveis de ambiente:
   ```bash
   fly secrets list
   ```

3. Configure `REDIS_HOST` corretamente:
   - Fly.io: `equinoid-redis.internal`
   - Railway: Usar REDIS_URL fornecida
   - Render: Usar Internal Redis URL

### Erro: "Arquivo .env não encontrado"

**Causa**: Normal em produção, o app usa variáveis de ambiente do sistema

**Solução**: Configure as secrets/variáveis na plataforma, não use arquivo .env

### Health Check Failing

**Causa**: App não está respondendo na porta correta

**Solução**:
1. Verifique se `PORT=8080` está configurado
2. Verifique se o Dockerfile expõe a porta 8080
3. Verifique os logs: `fly logs` ou `railway logs`

---

## Comandos Úteis

### Fly.io

```bash
fly status                    # Status da aplicação
fly logs                      # Ver logs em tempo real
fly ssh console               # Acessar container
fly scale memory 512          # Ajustar memória
fly secrets list              # Listar secrets
fly secrets set KEY=VALUE     # Adicionar secret
fly apps list                 # Listar apps
```

### Railway

```bash
railway status                # Status
railway logs                  # Logs
railway variables             # Ver variáveis
railway open                  # Abrir no browser
```

### Render

Usar o dashboard web: render.com/dashboard

---

## Estrutura de Custos

### Fly.io
- Free tier: $5/mês de crédito
- Redis: ~$2/mês (512MB)
- API: ~$3/mês (1GB RAM)

### Railway
- Free tier: $5/mês de crédito
- Após: $5/mês por GB de RAM

### Render
- Free tier disponível
- Starter: $7/mês

---

## Monitoramento

### Health Check

Todas as plataformas verificam automaticamente:
- URL: `/health`
- Porta: `8080`
- Intervalo: 30s

### Logs

```bash
# Fly.io
fly logs --app equinoid-backend

# Railway
railway logs

# Render
Ver no dashboard
```

### Metrics

Acesse o endpoint:
```
https://seu-app.fly.dev/metrics
```

---

## Checklist de Deploy

- [ ] Redis configurado e rodando
- [ ] Variáveis de ambiente configuradas
- [ ] DATABASE_URL apontando para Supabase
- [ ] JWT_SECRET definido
- [ ] REDIS_HOST correto para a plataforma
- [ ] Porta 8080 configurada
- [ ] Health check respondendo
- [ ] Logs sem erros
- [ ] API acessível publicamente

---

## Próximos Passos

1. **Domain Custom**: Configure domínio personalizado
2. **SSL/TLS**: Configurado automaticamente
3. **CI/CD**: Configure GitHub Actions para deploy automático
4. **Backup**: Configure backup do Redis
5. **Monitoring**: Configure Sentry ou similar
6. **Scaling**: Configure auto-scaling conforme demanda
