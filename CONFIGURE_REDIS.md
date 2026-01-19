# Configurar Redis - Solução Rápida

## Problema

Erro: `redis connection failed: dial tcp [::1]:6379: connect: connection refused`

**Causa**: As variáveis de ambiente REDIS_HOST e REDIS_PORT não estão configuradas no Fly.io.

## Solução

### Opção 1: Instalar Fly CLI e Configurar (Recomendado)

#### 1. Instalar Fly CLI no Windows

```powershell
iwr https://fly.io/install.ps1 -useb | iex
```

Depois, **FECHE E ABRA o terminal novamente**.

#### 2. Fazer Login

```powershell
fly auth login
```

#### 3. Verificar Secrets Atuais

```powershell
fly secrets list --app equinoid-backend
```

#### 4. Configurar Redis

**Opção A: Criar Redis no Fly.io (Recomendado)**

```powershell
fly apps create equinoid-redis

fly volumes create redis_data --size 1 --region gru --app equinoid-redis

fly deploy -c fly-redis.toml
```

Depois configure as variáveis:

```powershell
fly secrets set REDIS_HOST="equinoid-redis.internal" REDIS_PORT="6379" --app equinoid-backend
```

**Opção B: Usar Upstash (Redis Grátis)**

1. Acesse: https://upstash.com/
2. Crie conta grátis
3. Crie um Redis Database
4. Copie as credenciais
5. Configure:

```powershell
fly secrets set REDIS_HOST="seu-endpoint.upstash.io" REDIS_PORT="6379" REDIS_PASSWORD="sua-senha" --app equinoid-backend
```

#### 5. Redeploy

```powershell
fly deploy
```

#### 6. Verificar Logs

```powershell
fly logs --app equinoid-backend
```

---

### Opção 2: Configurar pelo Dashboard Fly.io (Sem CLI)

Se não quiser instalar o CLI:

#### 1. Acesse o Dashboard

https://fly.io/dashboard

#### 2. Selecione o App "equinoid-backend"

#### 3. Vá em "Secrets"

#### 4. Adicione as seguintes secrets:

```
REDIS_HOST = equinoid-redis.internal
REDIS_PORT = 6379
```

OU se usar Upstash:

```
REDIS_HOST = seu-endpoint.upstash.io
REDIS_PORT = 6379
REDIS_PASSWORD = sua-senha
```

#### 5. Redeploy o App

No dashboard, clique em "Deploy" ou faça push no Git.

---

### Opção 3: Usar Redis Temporário (Desenvolvimento)

Se quiser apenas testar sem Redis (NÃO RECOMENDADO para produção):

1. Modifique o código para fazer Redis opcional
2. Configure `REDIS_HOST=""` para desabilitar

---

## Verificação

Após configurar, os logs devem mostrar:

```
✅ Server started on :8080
✅ Connected to Redis
✅ Database connected
```

Em vez de:

```
❌ redis connection failed: dial tcp [::1]:6379: connect: connection refused
```

---

## Comandos Úteis

```powershell
fly status --app equinoid-backend

fly logs --app equinoid-backend

fly secrets list --app equinoid-backend

fly ssh console --app equinoid-backend
```

---

## Troubleshooting

### Fly CLI não reconhecido após instalação

Feche e abra o PowerShell/Terminal novamente.

### Erro ao criar Redis

Você já tem um Redis criado? Verifique:

```powershell
fly apps list
```

### Redis não conecta

Verifique se está na mesma região:

```powershell
fly status --app equinoid-redis
fly status --app equinoid-backend
```

Ambos devem estar em `gru` (São Paulo).

### Ainda com erro

Verifique as variáveis:

```powershell
fly ssh console --app equinoid-backend
env | grep REDIS
```

---

## Custo

- **Fly.io**: $5/mês de crédito grátis
- **Redis 512MB**: ~$2/mês
- **Upstash**: Grátis até 10k comandos/dia

---

## Próximos Passos

Depois de configurar:

1. ✅ Redis funcionando
2. ✅ API rodando
3. Configure domínio custom
4. Configure CI/CD
5. Configure backup do Redis
