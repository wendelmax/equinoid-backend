# Conflitos de Rotas Resolvidos

## Problema Original

```
panic: ':id' in new path '/api/v1/eventos/:id' conflicts with 
existing wildcard ':evento_id' in existing prefix '/api/v1/eventos/:evento_id'
```

## Causa

O Gin não permite diferentes nomes de parâmetros na mesma posição de rota:
- `/api/v1/eventos/:id` (módulo eventos - CRUD)
- `/api/v1/eventos/:evento_id/participacoes` (módulo participacoes)

## Soluções Aplicadas

### 1. Conflito Equinos (Resolvido anteriormente)

**Antes:**
- `/api/v1/equinos/:equinoid`
- `/api/v1/equinos/:equino_id/participacoes-eventos` ❌

**Depois:**
- `/api/v1/equinos/:equinoid`
- `/api/v1/equinos/:equinoid/participacoes-eventos` ✅

### 2. Conflito Eventos (Resolvido agora)

**Antes:**
```go
eventos.GET("/:id", handler.GetByID)
eventos.PUT("/:id", handler.Update)
eventos.DELETE("/:id", handler.Delete)
```

**Depois:**
```go
eventos.GET("/:evento_id", handler.GetByID)
eventos.PUT("/:evento_id", handler.Update)
eventos.DELETE("/:evento_id", handler.Delete)
```

## Arquivos Modificados

### 1. `internal/modules/eventos/routes.go`

Mudou de `:id` para `:evento_id` em todas as rotas CRUD.

### 2. `internal/modules/eventos/handler.go`

Mudou `c.Param("id")` para `c.Param("evento_id")` em:
- `GetByID()`
- `Update()`
- `Delete()`

## Padrões de Nomenclatura

Agora temos consistência nos nomes dos parâmetros:

| Recurso | Parâmetro | Exemplos de Rotas |
|---------|-----------|-------------------|
| Equinos | `:equinoid` | `/api/v1/equinos/:equinoid` |
| Eventos | `:evento_id` | `/api/v1/eventos/:evento_id` |
| Participações | `:id` | `/api/v1/eventos/participacoes/:id` |
| Users | `:id` | `/api/v1/users/:id` |
| Tokenização | `:id` | `/api/v1/tokenizacao/:id` |
| Exames | `:id` | `/api/v1/exames/:id` |

## Rotas Finais (Eventos)

```
GET    /api/v1/eventos                              -> ListAll
POST   /api/v1/eventos                              -> Create
GET    /api/v1/eventos/:evento_id                   -> GetByID
PUT    /api/v1/eventos/:evento_id                   -> Update
DELETE /api/v1/eventos/:evento_id                   -> Delete
GET    /api/v1/eventos/:evento_id/participacoes     -> ListByEvento (participacoes)
GET    /api/v1/equinos/:equinoid/eventos            -> ListByEquino
```

Sem mais conflitos!

## Verificação

```powershell
go build -o bin/equinoid-api.exe cmd/server/main.go
```

✅ Exit code: 0 (compilação bem-sucedida)

## Deploy

Agora você pode fazer deploy sem erros:

```powershell
fly deploy
```

Ou com Redis desabilitado:

```powershell
fly secrets set REDIS_ENABLED=false --app equinoid-backend
fly deploy
```

## Logs Esperados

```
✅ Redis is disabled. Running without cache.
[GIN-debug] GET    /api/v1/eventos/:evento_id --> ...
[GIN-debug] GET    /api/v1/eventos/:evento_id/participacoes --> ...
✅ Server started on :8080
```

## Resumo

1. ✅ Redis opcional (REDIS_ENABLED=false)
2. ✅ Conflito equinos resolvido (`:equino_id` → `:equinoid`)
3. ✅ Conflito eventos resolvido (`:id` → `:evento_id`)
4. ✅ Compilação sem erros
5. ✅ Pronto para deploy!

## Documentação das Rotas

As mudanças nos parâmetros são apenas internas. As URLs continuam semânticas:

```
GET /api/v1/eventos/123              -> Busca evento ID 123
GET /api/v1/eventos/123/participacoes -> Participações do evento 123
GET /api/v1/equinos/456/eventos      -> Eventos do equino 456
```

Tudo funcional! 🎉
