# Como Criar o Primeiro Usuário Admin

O sistema **não possui** um usuário master pré-configurado. Você precisa criar o primeiro administrador manualmente.

## ⚠️ Situação Atual

- O endpoint `/auth/register` sempre cria usuários como **"criador"** (não admin)
- O endpoint `/users` (POST) requer autenticação admin para criar usuários
- **Não há seed automático** de usuário admin

## 🔧 Métodos para Criar o Primeiro Admin

### Método 1: Script Go (Recomendado)

```bash
# 1. Configure a DATABASE_URL
export DATABASE_URL="postgresql://user:pass@host:5432/equinoid"

# 2. Execute o script
go run scripts/create-admin.go admin@equinoid.com senha123 "Administrador" 12345678900
```

**Parâmetros:**
- Email: email do admin
- Senha: senha do admin
- Nome: nome completo
- CPF/CNPJ: (opcional, padrão: 00000000000)

### Método 2: Script SQL Direto

```bash
# 1. Conecte ao banco
psql $DATABASE_URL

# 2. Execute o SQL (ALTERE A SENHA!)
# Primeiro gere o hash da senha:
SELECT crypt('sua_senha_aqui', gen_salt('bf', 12));

# 3. Use o hash gerado no INSERT:
INSERT INTO users (
    email, password, name, user_type, cpf_cnpj,
    is_active, is_email_verified, role, created_at, updated_at
) VALUES (
    'admin@equinoid.com',
    '$2a$12$SEU_HASH_AQUI', -- Cole o hash gerado acima
    'Administrador',
    'admin',
    '00000000000',
    true,
    true,
    'admin',
    NOW(),
    NOW()
);
```

### Método 3: Via API (após criar primeiro admin)

Depois de criar o primeiro admin, você pode criar outros admins via API:

```bash
# 1. Faça login como admin
curl -X POST https://equinoid.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@equinoid.com",
    "password": "senha123"
  }'

# 2. Use o access_token para criar novos usuários
curl -X POST https://equinoid.com/api/v1/users \
  -H "Authorization: Bearer SEU_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "novo-admin@equinoid.com",
    "password": "senha123",
    "name": "Novo Admin",
    "user_type": "admin",
    "cpf_cnpj": "12345678900"
  }'
```

### Método 4: Script PowerShell (Windows)

```powershell
# Configure DATABASE_URL
$env:DATABASE_URL = "postgresql://user:pass@host:5432/equinoid"

# Execute o script
.\scripts\create-admin.ps1 -Email "admin@equinoid.com" -Password "senha123" -Name "Administrador"
```

## 🔐 Credenciais Padrão Sugeridas

**⚠️ IMPORTANTE: Altere essas credenciais em produção!**

- **Email**: `admin@equinoid.com`
- **Senha**: `admin123` (ou outra de sua escolha)
- **Nome**: `Administrador`
- **CPF/CNPJ**: `00000000000` (ou um CPF/CNPJ válido)

## ✅ Verificação

Após criar o admin, teste o login:

```bash
curl -X POST https://equinoid.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@equinoid.com",
    "password": "sua_senha"
  }'
```

Você deve receber um `access_token` e `refresh_token`.

## 📋 Permissões do Admin

Usuários com `user_type = "admin"` podem:

- ✅ Criar novos usuários (qualquer tipo)
- ✅ Listar todos os usuários
- ✅ Atualizar qualquer usuário
- ✅ Deletar usuários
- ✅ Ativar/Desativar usuários
- ✅ Acessar todos os endpoints protegidos

## 🚨 Segurança

1. **Altere a senha padrão** imediatamente após criar
2. **Use senhas fortes** (mínimo 8 caracteres, recomendado 12+)
3. **Não compartilhe** credenciais de admin
4. **Use HTTPS** em produção
5. **Configure JWT_SECRET** forte em produção

## 🔄 Criar Admin via Swagger

1. Acesse: https://equinoid.com/swagger/index.html
2. Faça login como admin existente
3. Use o endpoint `POST /users` para criar novos admins
