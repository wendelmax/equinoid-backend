# Script PowerShell para criar usuário admin no Equinoid Backend
# Uso: .\scripts\create-admin.ps1 [email] [senha] [nome] [cpf_cnpj]

param(
    [string]$Email = "admin@equinoid.com",
    [string]$Password = "admin123",
    [string]$Name = "Administrador",
    [string]$CPFCNPJ = "00000000000"
)

if (-not $env:DATABASE_URL) {
    Write-Host "ERRO: DATABASE_URL não configurada" -ForegroundColor Red
    Write-Host "Configure a variável de ambiente DATABASE_URL"
    exit 1
}

Write-Host "Criando usuário admin..." -ForegroundColor Yellow
Write-Host "Email: $Email"
Write-Host "Nome: $Name"

# Gerar hash da senha usando Go
$hashScript = @"
package main
import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)
func main() {
    hash, err := bcrypt.GenerateFromPassword([]byte("$Password"), bcrypt.DefaultCost)
    if err != nil {
        panic(err)
    }
    fmt.Print(string(hash))
}
"@

$hashScript | Out-File -FilePath "$env:TEMP\genhash.go" -Encoding UTF8
$hash = go run "$env:TEMP\genhash.go" 2>$null
Remove-Item "$env:TEMP\genhash.go" -ErrorAction SilentlyContinue

if (-not $hash) {
    Write-Host "AVISO: Não foi possível gerar hash. Usando hash padrão." -ForegroundColor Yellow
    $hash = '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYqJZ5x5K5K'
}

# Inserir no banco usando psql
$sql = @"
INSERT INTO users (
    email,
    password,
    name,
    user_type,
    cpf_cnpj,
    is_active,
    is_email_verified,
    role,
    created_at,
    updated_at
) VALUES (
    '$Email',
    '$hash',
    '$Name',
    'admin',
    '$CPFCNPJ',
    true,
    true,
    'admin',
    NOW(),
    NOW()
) ON CONFLICT (email) DO UPDATE SET
    password = EXCLUDED.password,
    name = EXCLUDED.name,
    user_type = 'admin',
    is_active = true,
    updated_at = NOW();
"@

$sql | psql $env:DATABASE_URL

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Usuário admin criado/atualizado com sucesso!" -ForegroundColor Green
    Write-Host "   Email: $Email"
    Write-Host "   Senha: $Password"
    Write-Host "`nFaça login em: POST /api/v1/auth/login" -ForegroundColor Cyan
} else {
    Write-Host "`nERRO ao criar usuário admin" -ForegroundColor Red
    exit 1
}
