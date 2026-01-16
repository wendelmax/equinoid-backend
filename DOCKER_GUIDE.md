# Guia de Execução com Docker

## Problema Identificado

O erro ocorre porque a aplicação precisa de **Redis** para funcionar, mas quando rodamos apenas o container da API, o Redis não está disponível.

## Solução: Docker Compose

Use o Docker Compose para rodar todos os serviços necessários juntos.

## Passo a Passo

### 1. Configurar Variáveis de Ambiente

Copie o template e configure as variáveis:

```bash
cp env.template .env
```

Certifique-se que o arquivo `.env` contém:

```env
DATABASE_URL=postgresql://postgres:M4EYIU4ne9j5JIId@db.rqaemzdqntwuomycewrn.supabase.co:5432/postgres
SUPABASE_URL=https://rqaemzdqntwuomycewrn.supabase.co
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
SUPABASE_JWT_SECRET=COPIAR_DO_DASHBOARD_PROJECT_SETTINGS_API
JWT_SECRET=equinoid-fallback-secret
PORT=8080
GIN_MODE=release
ENVIRONMENT=production

REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

### 2. Iniciar os Serviços

```bash
docker-compose up -d
```

Este comando irá:
- Baixar a imagem do Redis
- Construir a imagem da API
- Iniciar ambos os serviços
- Conectar automaticamente a API ao Redis

### 3. Verificar Status

```bash
docker-compose ps
```

Você deve ver:
```
NAME               STATUS          PORTS
equinoid-api       Up (healthy)    0.0.0.0:8080->8080/tcp
equinoid-redis     Up (healthy)    0.0.0.0:6379->6379/tcp
```

### 4. Ver Logs

```bash
docker-compose logs -f api
```

Para ver logs de todos os serviços:

```bash
docker-compose logs -f
```

### 5. Testar a API

```bash
curl http://localhost:8080/health
```

## Comandos Úteis

### Parar os serviços

```bash
docker-compose down
```

### Parar e remover volumes

```bash
docker-compose down -v
```

### Rebuild da API

```bash
docker-compose up -d --build api
```

### Reiniciar apenas a API

```bash
docker-compose restart api
```

### Ver logs em tempo real

```bash
docker-compose logs -f api
```

## Estrutura dos Serviços

### Redis
- **Porta**: 6379
- **Persistência**: Volume `redis-data`
- **Configuração**: AOF habilitado para persistência
- **Health check**: `redis-cli ping`

### API (equinoid-backend)
- **Porta**: 8080
- **Dependências**: Redis (aguarda ficar saudável)
- **Health check**: `GET /health`
- **Restart**: Automático em caso de falha

## Variáveis de Ambiente Importantes

### Obrigatórias
- `DATABASE_URL`: URL completa do PostgreSQL (Supabase)
- `JWT_SECRET`: Chave secreta para JWT
- `REDIS_HOST`: Host do Redis (definido como `redis` no docker-compose)

### Opcionais
- `GIN_MODE`: `debug` ou `release`
- `ENVIRONMENT`: `development` ou `production`
- `REDIS_PASSWORD`: Senha do Redis (se houver)

## Troubleshooting

### Erro de conexão com Redis

Se ainda houver erro de conexão:

1. Verifique se o Redis está rodando:
   ```bash
   docker-compose ps redis
   ```

2. Teste a conexão com o Redis:
   ```bash
   docker-compose exec redis redis-cli ping
   ```

3. Verifique as variáveis de ambiente da API:
   ```bash
   docker-compose exec api env | grep REDIS
   ```

### Erro de health check

Se o health check falhar:

1. Acesse o container:
   ```bash
   docker-compose exec api sh
   ```

2. Verifique os logs:
   ```bash
   docker-compose logs api
   ```

### Rebuild completo

Para fazer rebuild completo sem cache:

```bash
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## Usando Makefile

Você também pode adicionar comandos ao Makefile:

```bash
make docker-compose-up     # docker-compose up -d
make docker-compose-down   # docker-compose down
make docker-compose-logs   # docker-compose logs -f
```

## Produção

Para produção, considere:

1. Usar Redis com senha (configure `REDIS_PASSWORD`)
2. Usar volumes externos para backup
3. Configurar monitoramento (Prometheus/Grafana)
4. Usar secrets em vez de variáveis de ambiente
5. Configurar reverse proxy (Nginx/Traefik)

## Acesso aos Serviços

- **API**: http://localhost:8080
- **Swagger**: http://localhost:8080/swagger/index.html
- **Health**: http://localhost:8080/health
- **Metrics**: http://localhost:8080/metrics
- **Redis**: localhost:6379
