# 🐴 Equinoid Backend - Sistema Perfeito

> API REST completa com Tokenização RWA (Real World Assets) para gestão de equinos

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)]()
[![Status](https://img.shields.io/badge/Status-Production%20Ready-success)]()
[![Endpoints](https://img.shields.io/badge/Endpoints-78-orange)]()
[![Módulos](https://img.shields.io/badge/Módulos-16-blue)]()

---

## 🚀 Quick Start

```bash
# Instalar dependências
go mod download

# Configurar ambiente
cp .env.example .env

# Executar migrations
psql $DATABASE_URL < migrations/004_sprint1_core.sql
psql $DATABASE_URL < migrations/005_tokenizacao_rwa.sql

# Iniciar servidor
go run cmd/server/main.go

# ✅ Server running on http://localhost:8080
```

---

## 📊 O Que Está Implementado

### 16 Módulos Completos

| Módulo | Endpoints | Destaque |
|--------|-----------|----------|
| auth | 6 | JWT + Refresh tokens |
| users | 12 | Perfil + Admin CRUD |
| equinos | 6 | Gestão completa |
| simulador | 1 | Genético com IA |
| participacoes | 7 | Eventos competitivos |
| gestacao | 3 | Reprodução + ultrassom |
| eventos | 5 | Listagem geral 🆕 |
| **tokenizacao** | **7** | **RWA completo** 🆕 ⭐ |
| leiloes | 7 | Participações 🆕 |
| exames | 8 | Workflow laboratorial 🆕 |
| rankings | 2 | Gamificação 🆕 |
| relatorios | 1 | Dashboard stats 🆕 |
| financeiro | 5 | Gestão financeira 🆕 |
| nutricao | 4 | Planos + IA 🆕 |
| treinamento | 4 | Performance tracking 🆕 |
| social | proxy | Perfis (legado) |

**Total**: 78 endpoints implementados

---

## 🔥 Tokenização RWA - Destaque

Sistema completo de tokenização de ativos do mundo real:

- ✅ Compliance regulatório (51% mínimo do dono)
- ✅ Trading system (ofertas + ordens)
- ✅ Rating de risco (AAA+ até C)
- ✅ Blockchain-ready (hash SHA-256)
- ✅ Garantias biológicas
- ✅ ROI automático

**Único no mercado para equinos!**

---

## 🏗️ Arquitetura

### Modular Monolith

```
internal/modules/
├── [modulo]/
│   ├── repository.go    # Data access
│   ├── service.go       # Business logic
│   ├── handler.go       # HTTP handlers
│   └── routes.go        # Route definitions
```

**Padrões**: Repository Pattern | DI | Clean Architecture | SOLID

---

## 📚 Documentação

- 📖 [API Documentation](API_DOCUMENTATION.md) - Endpoints detalhados
- 📖 [Why SQLite Tests](WHY_SQLITE_TESTS.md) - Estratégia de testes
- 📖 [Sistema Perfeito](../SISTEMA_PERFEITO_COMPLETO.md) - Overview completo
- 📖 [Guia de Deploy](../GUIA_DEPLOY_PRODUCAO.md) - Deploy em produção
- 📖 Swagger UI: `/swagger/index.html`

---

## 🧪 Testes

```bash
# Rodar testes
go test ./...

# Com coverage
go test -cover ./...
```

---

## 🔐 Variáveis de Ambiente

```env
# Database
DATABASE_URL=postgresql://user:pass@host:5432/equinoid

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-secret-key-256-bits
JWT_EXPIRATION_HOURS=24

# Server
PORT=8080
GO_ENV=development
```

---

## 📊 Status

- ✅ 78 endpoints implementados (86%)
- ✅ 16 módulos completos
- ✅ 100% das páginas frontend funcionais
- ✅ Compilação sem erros
- ✅ Production ready

---

## 🚀 Deploy

Ver [GUIA_DEPLOY_PRODUCAO.md](../GUIA_DEPLOY_PRODUCAO.md) para instruções completas.

---

**Versão**: 2.0  
**Status**: 🟢 Production Ready  
**Licença**: Proprietária
