# 📚 EquinoId API - Documentação Completa

## 🚀 Acesso à Documentação

### Swagger UI (Interativo)
```
http://localhost:8080/swagger/index.html
```

### OpenAPI JSON
```
http://localhost:8080/swagger/doc.json
```

---

## 📋 Informações Gerais

- **Base URL**: `http://localhost:8080`
- **Versão**: v1.0.0
- **Protocolo**: HTTP/HTTPS
- **Formato**: JSON
- **Autenticação**: JWT Bearer Token

---

## 🔐 Autenticação

### JWT Bearer Token

Todas as rotas protegidas requerem header:
```http
Authorization: Bearer <seu_token_jwt>
```

### Fluxo de Autenticação

1. **Registro** → Obter credenciais
2. **Login** → Receber access_token + refresh_token
3. **Usar Token** → Incluir em requisições
4. **Renovar** → Usar refresh_token quando access_token expirar

---

## 📍 Endpoints Implementados

### 1. Autenticação

#### POST /api/v1/auth/register
Registrar novo usuário

**Request Body**:
```json
{
  "email": "usuario@example.com",
  "password": "senha123",
  "name": "Nome do Usuário",
  "user_type": "criador",
  "cpfcnpj": "12345678900"
}
```

**Response 201**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "usuario@example.com",
    "name": "Nome do Usuário",
    "user_type": "criador",
    "is_active": true,
    "created_at": "2025-10-08T15:00:00Z"
  },
  "message": "Usuário registrado com sucesso",
  "timestamp": "2025-10-08T15:00:00Z"
}
```

**Erros**:
- `400` - Dados inválidos
- `409` - Email já existe

---

#### POST /api/v1/auth/login
Fazer login

**Request Body**:
```json
{
  "email": "usuario@example.com",
  "password": "senha123"
}
```

**Response 200**:
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 86400,
    "token_type": "Bearer",
    "user": {
      "id": 1,
      "email": "usuario@example.com",
      "name": "Nome do Usuário",
      "user_type": "criador"
    }
  },
  "timestamp": "2025-10-08T15:00:00Z"
}
```

**Erros**:
- `400` - Dados inválidos
- `401` - Credenciais incorretas

---

#### POST /api/v1/auth/refresh
Renovar token

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response 200**:
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 86400,
    "token_type": "Bearer"
  }
}
```

---

### 2. Usuários

#### GET /api/v1/users/profile
Obter perfil do usuário logado

**Headers**:
```
Authorization: Bearer <token>
```

**Response 200**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "usuario@example.com",
    "name": "Nome do Usuário",
    "user_type": "criador",
    "cpfcnpj": "12345678900",
    "is_active": true,
    "created_at": "2025-10-08T15:00:00Z",
    "updated_at": "2025-10-08T15:00:00Z"
  }
}
```

---

#### PUT /api/v1/users/profile
Atualizar perfil

**Headers**:
```
Authorization: Bearer <token>
```

**Request Body**:
```json
{
  "name": "Novo Nome",
  "cpfcnpj": "98765432100"
}
```

**Response 200**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "usuario@example.com",
    "name": "Novo Nome",
    "cpfcnpj": "98765432100"
  },
  "message": "Perfil atualizado com sucesso"
}
```

---

#### DELETE /api/v1/users/profile
Deletar conta (soft delete)

**Headers**:
```
Authorization: Bearer <token>
```

**Response 200**:
```json
{
  "success": true,
  "message": "Conta deletada com sucesso"
}
```

---

### 3. Equinos

#### GET /api/v1/equinos
Listar equinos com paginação e filtros

**Headers**:
```
Authorization: Bearer <token>
```

**Query Params**:
- `page` (int) - Número da página (default: 1)
- `limit` (int) - Itens por página (default: 20, max: 100)
- `nome` (string) - Filtrar por nome (busca parcial)
- `raca` (string) - Filtrar por raça
- `sexo` (string) - Filtrar por sexo (macho/femea)

**Exemplo**:
```
GET /api/v1/equinos?page=1&limit=10&raca=Mangalarga&sexo=macho
```

**Response 200**:
```json
{
  "success": true,
  "data": {
    "equinos": [
      {
        "id": 1,
        "equinoid_id": "BRA-2020-00000001",
        "microchip_id": "CHIP001",
        "nome": "Cavalo Exemplo",
        "sexo": "macho",
        "raca": "Mangalarga",
        "cor": "Alazão",
        "data_nascimento": "2020-05-15T00:00:00Z",
        "status": "ativo",
        "proprietario_id": 1,
        "created_at": "2025-10-08T15:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 50,
      "pages": 5
    }
  }
}
```

---

#### POST /api/v1/equinos
Criar novo equino

**Headers**:
```
Authorization: Bearer <token>
```

**Request Body**:
```json
{
  "equinoid_id": "BRA-2020-00000001",
  "microchip_id": "CHIP001",
  "nome": "Cavalo Exemplo",
  "sexo": "macho",
  "raca": "Mangalarga",
  "cor": "Alazão",
  "data_nascimento": "2020-05-15T00:00:00Z",
  "marca_dagua": "Marca identificadora",
  "genitora_equinoid": "BRA-2015-00000100",
  "genitor_equinoid": "BRA-2014-00000200"
}
```

**Response 201**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "equinoid_id": "BRA-2020-00000001",
    "microchip_id": "CHIP001",
    "nome": "Cavalo Exemplo",
    "status": "ativo",
    "created_at": "2025-10-08T15:00:00Z"
  },
  "message": "Equino criado com sucesso"
}
```

**Erros**:
- `400` - Dados inválidos
- `409` - EquinoID ou MicrochipID já existe

---

#### GET /api/v1/equinos/:equinoid
Obter detalhes de um equino

**Headers**:
```
Authorization: Bearer <token>
```

**Params**:
- `equinoid` - ID do equino (ex: BRA-2020-00000001)

**Response 200**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "equinoid_id": "BRA-2020-00000001",
    "microchip_id": "CHIP001",
    "nome": "Cavalo Exemplo",
    "sexo": "macho",
    "raca": "Mangalarga",
    "cor": "Alazão",
    "data_nascimento": "2020-05-15T00:00:00Z",
    "genitora_equinoid": "BRA-2015-00000100",
    "genitor_equinoid": "BRA-2014-00000200",
    "proprietario": {
      "id": 1,
      "name": "Proprietário",
      "email": "proprietario@example.com"
    },
    "status": "ativo",
    "created_at": "2025-10-08T15:00:00Z",
    "updated_at": "2025-10-08T15:00:00Z"
  }
}
```

---

#### PUT /api/v1/equinos/:equinoid
Atualizar equino

**Headers**:
```
Authorization: Bearer <token>
```

**Request Body**:
```json
{
  "nome": "Novo Nome",
  "cor": "Tordilho"
}
```

**Response 200**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "equinoid_id": "BRA-2020-00000001",
    "nome": "Novo Nome",
    "cor": "Tordilho"
  },
  "message": "Equino atualizado com sucesso"
}
```

---

#### DELETE /api/v1/equinos/:equinoid
Deletar equino (soft delete)

**Headers**:
```
Authorization: Bearer <token>
```

**Response 200**:
```json
{
  "success": true,
  "message": "Equino deletado com sucesso"
}
```

---

### 4. Health & Monitoring

#### GET /health
Verificar saúde da API

**Response 200**:
```json
{
  "status": "healthy",
  "timestamp": "2025-10-08T18:10:19Z",
  "services": {
    "database": "healthy",
    "redis": "healthy",
    "external_apis": "healthy"
  },
  "version": "1.0.0",
  "uptime": "1h2m37s",
  "system": {
    "go_version": "go1.25.1",
    "num_goroutine": 9,
    "memory_usage": "10.7 MB",
    "cpu_count": 8
  }
}
```

---

#### GET /metrics
Métricas Prometheus

**Response 200** (text/plain):
```
# HELP go_goroutines Number of goroutines
go_goroutines 9
# HELP api_requests_total Total API requests
api_requests_total{method="GET",path="/health",status="200"} 100
...
```

---

## 🔒 Autenticação e Autorização

### JWT Token Structure

**Access Token Claims**:
```json
{
  "user_id": 1,
  "email": "usuario@example.com",
  "user_type": "criador",
  "exp": 1696809600,
  "iat": 1696723200
}
```

**Refresh Token Claims**:
```json
{
  "user_id": 1,
  "exp": 1697328000,
  "iat": 1696723200
}
```

### Tipos de Usuário

- `criador` - Criador de equinos
- `veterinario` - Veterinário
- `admin` - Administrador do sistema
- `laboratorio` - Laboratório de análises

---

## 📊 Códigos de Status HTTP

| Código | Significado | Quando usar |
|--------|-------------|-------------|
| 200 | OK | Sucesso geral |
| 201 | Created | Recurso criado |
| 204 | No Content | Sucesso sem corpo |
| 400 | Bad Request | Dados inválidos |
| 401 | Unauthorized | Token ausente/inválido |
| 403 | Forbidden | Sem permissão |
| 404 | Not Found | Recurso não encontrado |
| 409 | Conflict | Conflito (ex: email duplicado) |
| 422 | Unprocessable Entity | Validação falhou |
| 429 | Too Many Requests | Rate limit excedido |
| 500 | Internal Server Error | Erro do servidor |

---

## ⚠️ Tratamento de Erros

### Formato Padrão de Erro

```json
{
  "success": false,
  "error": "Mensagem de erro amigável",
  "details": "Detalhes técnicos do erro",
  "timestamp": "2025-10-08T15:00:00Z",
  "path": "/api/v1/equinos",
  "request_id": "uuid-1234-5678"
}
```

### Exemplos de Erros

**400 - Bad Request**:
```json
{
  "success": false,
  "error": "Invalid request data",
  "details": "email: required field missing",
  "timestamp": "2025-10-08T15:00:00Z"
}
```

**401 - Unauthorized**:
```json
{
  "success": false,
  "error": "Unauthorized",
  "details": "Token expired or invalid",
  "timestamp": "2025-10-08T15:00:00Z"
}
```

**409 - Conflict**:
```json
{
  "success": false,
  "error": "EquinoID já existe",
  "details": "BRA-2020-00000001 is already registered",
  "timestamp": "2025-10-08T15:00:00Z"
}
```

---

## 🔄 Rate Limiting

- **Limite**: 100 requisições por minuto por IP
- **Header Response**: `X-RateLimit-Remaining: 95`
- **Erro**: 429 Too Many Requests

---

## 📝 Exemplos de Uso

### cURL

**Login**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "usuario@example.com",
    "password": "senha123"
  }'
```

**Listar Equinos**:
```bash
curl -X GET "http://localhost:8080/api/v1/equinos?page=1&limit=10" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

**Criar Equino**:
```bash
curl -X POST http://localhost:8080/api/v1/equinos \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "equinoid_id": "BRA-2020-00000001",
    "microchip_id": "CHIP001",
    "nome": "Cavalo Teste",
    "sexo": "macho",
    "raca": "Mangalarga"
  }'
```

### JavaScript/Fetch

```javascript
// Login
const login = async () => {
  const response = await fetch('http://localhost:8080/api/v1/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      email: 'usuario@example.com',
      password: 'senha123'
    })
  });
  
  const data = await response.json();
  localStorage.setItem('token', data.data.access_token);
};

// Listar Equinos
const getEquinos = async () => {
  const token = localStorage.getItem('token');
  const response = await fetch('http://localhost:8080/api/v1/equinos?page=1&limit=10', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  const data = await response.json();
  return data.data.equinos;
};
```

### Python

```python
import requests

# Login
response = requests.post(
    'http://localhost:8080/api/v1/auth/login',
    json={
        'email': 'usuario@example.com',
        'password': 'senha123'
    }
)
token = response.json()['data']['access_token']

# Listar Equinos
headers = {'Authorization': f'Bearer {token}'}
response = requests.get(
    'http://localhost:8080/api/v1/equinos',
    headers=headers,
    params={'page': 1, 'limit': 10}
)
equinos = response.json()['data']['equinos']
```

---

## 🚀 Primeiros Passos

### 1. Iniciar Backend
```bash
docker-compose up -d postgres redis
cd services/api
./bin/api.exe
```

### 2. Acessar Swagger
```
http://localhost:8080/swagger/index.html
```

### 3. Criar Conta
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "seu@email.com",
    "password": "senha123",
    "name": "Seu Nome",
    "user_type": "criador"
  }'
```

### 4. Fazer Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "seu@email.com",
    "password": "senha123"
  }'
```

### 5. Usar Token
Copie o `access_token` da resposta e use em requisições:
```bash
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer SEU_TOKEN_AQUI"
```

---

## 📖 Recursos Adicionais

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health
- **Metrics**: http://localhost:8080/metrics
- **Repositório**: https://github.com/wendelmax/equinoId

---

## 🆘 Suporte

- **Email**: suporte@equinoid.org
- **Issues**: GitHub Issues
- **Docs**: https://docs.equinoid.org

---

**Última Atualização**: 08/10/2025
**Versão**: 1.0.0

