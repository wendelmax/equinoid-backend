package main

import (
	"log"

	_ "github.com/equinoid/backend/docs"
	"github.com/equinoid/backend/internal/app"
	"github.com/equinoid/backend/internal/config"
	"github.com/equinoid/backend/pkg/logging"
	"github.com/joho/godotenv"
)

// @title Equinoid API - Sistema Perfeito
// @version 2.0
// @description API completa para gestão de equinos com Tokenização RWA (Real World Assets). Inclui: autenticação, gestão de equinos, tokenização, leilões, exames laboratoriais, nutrição com IA, treinamento, rankings, financeiro e muito mais.
// @termsOfService https://equinoid.com/terms

// @contact.name Equinoid Support
// @contact.url https://equinoid.com/support
// @contact.email suporte@equinoid.com

// @license.name Proprietária
// @license.url https://equinoid.com/license

// @host seahorse-app-28du8.ondigitalocean.app
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token. Formato: "Bearer {token}"

// @tag.name Auth
// @tag.description Endpoints de autenticação e autorização (JWT)

// @tag.name Users
// @tag.description Gerenciamento de usuários (perfil + admin)

// @tag.name Equinos
// @tag.description CRUD completo de equinos

// @tag.name Tokenização
// @tag.description Sistema RWA - Tokenização de equinos (CORE BUSINESS)

// @tag.name Leilões
// @tag.description Participações em leilões e vendas

// @tag.name Exames
// @tag.description Workflow de exames laboratoriais

// @tag.name Financeiro
// @tag.description Gestão financeira e dashboard

// @tag.name Nutrição
// @tag.description Planos alimentares com IA

// @tag.name Treinamento
// @tag.description Programas e sessões de treinamento

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	cfg := config.Load()

	logger := logging.NewLogger(cfg.LogLevel)

	if err := app.Start(app.Dependencies{Config: cfg, Logger: logger}); err != nil {
		logger.Fatalf("Falha ao iniciar servidor: %v", err)
	}
}
