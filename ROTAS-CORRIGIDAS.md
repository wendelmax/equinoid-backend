# Conflito de Rotas Resolvido

## Problema

```
panic: ':equino_id' in new path '/api/v1/equinos/:equino_id/participacoes-eventos' 
conflicts with existing wildcard ':equinoid' in existing prefix '/api/v1/equinos/:equinoid'
```

## Causa

O Gin (framework HTTP) não permite diferentes nomes de parâmetros na mesma posição da rota:

**Antes:**
- `/api/v1/equinos/:equinoid` (equinos module)
- `/api/v1/equinos/:equino_id/participacoes-eventos` (participacoes module)

Gin vê isso como conflito porque ambos usam `/api/v1/equinos/:PARAM`, mas com nomes diferentes.

## Solução

Padronizar TODOS os parâmetros de ID de equino para `:equinoid`.

### Arquivos Modificados

#### 1. `internal/modules/participacoes/routes.go`

**Antes:**
```go
equinos.GET("/:equino_id/participacoes-eventos", handler.ListByEquino)
```

**Depois:**
```go
equinos.GET("/:equinoid/participacoes-eventos", handler.ListByEquino)
```

#### 2. `internal/modules/participacoes/handler.go`

**Antes:**
```go
equinoID, err := strconv.ParseUint(c.Param("equino_id"), 10, 32)
```

**Depois:**
```go
equinoID, err := strconv.ParseUint(c.Param("equinoid"), 10, 32)
```

## Rotas Padronizadas

Todas as rotas de equinos agora usam `:equinoid` consistentemente:

```
GET    /api/v1/equinos/:equinoid
PUT    /api/v1/equinos/:equinoid
DELETE /api/v1/equinos/:equinoid
POST   /api/v1/equinos/:equinoid/transferir
GET    /api/v1/equinos/:equinoid/participacoes-eventos
GET    /api/v1/equinos/:equinoid/eventos
GET    /api/v1/equinos/:equinoid/rankings
POST   /api/v1/equinos/:equinoid/performance-materna
GET    /api/v1/nutricao/equino/:equinoid
POST   /api/v1/nutricao/equino/:equinoid/ai-suggestion
GET    /api/v1/tokenizacao/equinos/:equinoid
```

## Verificação

Compilação bem-sucedida:

```powershell
go build -o bin/equinoid-api.exe cmd/server/main.go
```

```
Exit code: 0
```

## Deploy

Agora você pode fazer deploy:

```powershell
fly deploy
```

Ou com Redis desabilitado:

```powershell
fly secrets set REDIS_ENABLED=false --app equinoid-backend
fly deploy
```

## Logs Esperados

Após o deploy, você verá:

```
✅ Redis is disabled. Running without cache.
[GIN-debug] GET    /api/v1/equinos/:equinoid --> ...
[GIN-debug] GET    /api/v1/equinos/:equinoid/participacoes-eventos --> ...
✅ Server started on :8080
```

Sem mais panic de conflito de rotas!

## Resumo das Mudanças

1. ✅ Redis agora é opcional (REDIS_ENABLED=false)
2. ✅ Conflito de rotas resolvido (`:equino_id` → `:equinoid`)
3. ✅ Compilação sem erros
4. ✅ Pronto para deploy

## Próximo Passo

```powershell
fly deploy
```

Aguarde alguns segundos e verifique os logs:

```powershell
fly logs --app equinoid-backend
```

Você deve ver:
```
✅ Server started on :8080
```
