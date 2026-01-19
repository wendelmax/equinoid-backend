# Configuração Frontend → Backend

## Backend rodando em: `172.66.0.96`

## Variáveis de Ambiente do Frontend

### Arquivo `.env` ou `.env.local`

```env
# URL da API Backend
NEXT_PUBLIC_API_URL=http://172.66.0.96:8080
NEXT_PUBLIC_API_BASE_URL=http://172.66.0.96:8080/api/v1

# Se o backend estiver com HTTPS
# NEXT_PUBLIC_API_URL=https://172.66.0.96:8080
# NEXT_PUBLIC_API_BASE_URL=https://172.66.0.96:8080/api/v1

# Supabase (se usar direto no frontend)
NEXT_PUBLIC_SUPABASE_URL=https://rqaemzdqntwuomycewrn.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InJxYWVtemRxbnR3dW9teWNld3JuIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjAyMDM0ODgsImV4cCI6MjA3NTc3OTQ4OH0.wfXXKFZxVMnF4TZmX4_ZuAuaNNVA6MiZQyCNLy0rhzQ
```

### Para React/Vite

```env
VITE_API_URL=http://172.66.0.96:8080
VITE_API_BASE_URL=http://172.66.0.96:8080/api/v1
```

### Para Create React App

```env
REACT_APP_API_URL=http://172.66.0.96:8080
REACT_APP_API_BASE_URL=http://172.66.0.96:8080/api/v1
```

## Exemplo de Uso no Código

### Next.js / React

```typescript
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://172.66.0.96:8080';
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://172.66.0.96:8080/api/v1';

const response = await fetch(`${API_BASE_URL}/auth/login`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'senha123'
  })
});
```

### Axios Config

```typescript
import axios from 'axios';

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://172.66.0.96:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export default api;
```

## Endpoints Disponíveis

Com base no IP `172.66.0.96:8080`:

### Autenticação
```
POST   http://172.66.0.96:8080/api/v1/auth/login
POST   http://172.66.0.96:8080/api/v1/auth/register
POST   http://172.66.0.96:8080/api/v1/auth/refresh
POST   http://172.66.0.96:8080/api/v1/auth/logout
```

### Equinos
```
GET    http://172.66.0.96:8080/api/v1/equinos
POST   http://172.66.0.96:8080/api/v1/equinos
GET    http://172.66.0.96:8080/api/v1/equinos/:equinoid
PUT    http://172.66.0.96:8080/api/v1/equinos/:equinoid
DELETE http://172.66.0.96:8080/api/v1/equinos/:equinoid
```

### Eventos
```
GET    http://172.66.0.96:8080/api/v1/eventos
POST   http://172.66.0.96:8080/api/v1/eventos
GET    http://172.66.0.96:8080/api/v1/eventos/:evento_id
PUT    http://172.66.0.96:8080/api/v1/eventos/:evento_id
DELETE http://172.66.0.96:8080/api/v1/eventos/:evento_id
```

### Health Check
```
GET    http://172.66.0.96:8080/health
GET    http://172.66.0.96:8080/health/ready
GET    http://172.66.0.96:8080/health/live
```

### Swagger (Documentação)
```
http://172.66.0.96:8080/swagger/index.html
```

## CORS - Importante!

Se você está rodando o frontend localmente (ex: `localhost:3000`) e conectando ao backend em `172.66.0.96`, você precisa garantir que o backend está configurado para aceitar requisições do seu domínio.

### Verificar CORS no Backend

O backend já deve ter CORS configurado em `internal/middleware/cors.go`, mas verifique se permite sua origem:

```go
AllowOrigins: []string{
    "http://localhost:3000",
    "http://localhost:5173",
    "http://172.66.0.96:3000",
    "*", // ou permitir todos (não recomendado em produção)
}
```

## Teste Rápido

### 1. Testar Backend

```bash
curl http://172.66.0.96:8080/health
```

Deve retornar:
```json
{
  "status": "healthy",
  "services": {
    "database": "healthy",
    "redis": "disabled"
  }
}
```

### 2. Testar Login

```bash
curl -X POST http://172.66.0.96:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@equinoid.com",
    "password": "senha123"
  }'
```

## Estrutura de Arquivos Frontend

```
frontend/
├── .env.local              # Variáveis locais (gitignore)
├── .env.example            # Exemplo para o time
├── src/
│   ├── config/
│   │   └── api.ts         # Configuração da API
│   ├── services/
│   │   ├── auth.ts        # Serviços de autenticação
│   │   ├── equinos.ts     # Serviços de equinos
│   │   └── eventos.ts     # Serviços de eventos
│   └── ...
```

### `src/config/api.ts`

```typescript
export const API_CONFIG = {
  baseURL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://172.66.0.96:8080/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
};

export const API_ENDPOINTS = {
  auth: {
    login: '/auth/login',
    register: '/auth/register',
    logout: '/auth/logout',
    refresh: '/auth/refresh',
  },
  equinos: {
    list: '/equinos',
    create: '/equinos',
    get: (id: string) => `/equinos/${id}`,
    update: (id: string) => `/equinos/${id}`,
    delete: (id: string) => `/equinos/${id}`,
  },
  eventos: {
    list: '/eventos',
    create: '/eventos',
    get: (id: number) => `/eventos/${id}`,
    update: (id: number) => `/eventos/${id}`,
    delete: (id: number) => `/eventos/${id}`,
  },
};
```

## Troubleshooting

### Erro: Network Error / Failed to Fetch

**Causa:** Backend não está acessível ou CORS bloqueado

**Solução:**
1. Verifique se o backend está rodando:
   ```bash
   curl http://172.66.0.96:8080/health
   ```

2. Verifique se você consegue acessar do navegador:
   ```
   http://172.66.0.96:8080/health
   ```

3. Verifique CORS no backend

### Erro: 401 Unauthorized

**Causa:** Token inválido ou expirado

**Solução:**
1. Faça login novamente
2. Verifique se o token está sendo enviado no header:
   ```typescript
   Authorization: `Bearer ${token}`
   ```

### Erro: 404 Not Found

**Causa:** Endpoint errado

**Solução:**
1. Verifique se a URL está correta (deve ter `/api/v1`)
2. Veja a documentação em `http://172.66.0.96:8080/swagger/index.html`

## Deploy do Frontend

### Variáveis em Produção

Se o backend estiver em um domínio próprio:

```env
NEXT_PUBLIC_API_URL=https://api.equinoid.com
NEXT_PUBLIC_API_BASE_URL=https://api.equinoid.com/api/v1
```

Se estiver no Fly.io:

```env
NEXT_PUBLIC_API_URL=https://equinoid-backend.fly.dev
NEXT_PUBLIC_API_BASE_URL=https://equinoid-backend.fly.dev/api/v1
```

## Exemplo Completo de Login

```typescript
import { useState } from 'react';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://172.66.0.96:8080/api/v1';

export function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const response = await fetch(`${API_BASE_URL}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        throw new Error('Login failed');
      }

      const data = await response.json();
      
      localStorage.setItem('token', data.data.token);
      localStorage.setItem('user', JSON.stringify(data.data.user));
      
      window.location.href = '/dashboard';
    } catch (error) {
      console.error('Login error:', error);
      alert('Erro ao fazer login');
    }
  };

  return (
    <form onSubmit={handleLogin}>
      <input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Email"
        required
      />
      <input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="Senha"
        required
      />
      <button type="submit">Entrar</button>
    </form>
  );
}
```

## Resumo

Para conectar o frontend ao backend em `172.66.0.96`:

1. Crie arquivo `.env.local` no frontend
2. Adicione:
   ```env
   NEXT_PUBLIC_API_BASE_URL=http://172.66.0.96:8080/api/v1
   ```
3. Use nas requisições:
   ```typescript
   fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/auth/login`, ...)
   ```
4. Teste com:
   ```
   http://172.66.0.96:8080/health
   ```

Pronto! 🎉
