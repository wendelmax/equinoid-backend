# Makefile para EquinoId Backend

# Variáveis
APP_NAME = equinoid-api
GO_VERSION = 1.24
DOCKER_REGISTRY = 
DOCKER_IMAGE = $(DOCKER_REGISTRY)equinoid/backend
VERSION = $(shell git describe --tags --always --dirty)
BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)"

# Cores para output
RED = \033[0;31m
GREEN = \033[0;32m
YELLOW = \033[1;33m
BLUE = \033[0;34m
NC = \033[0m # No Color

.PHONY: help build run clean test lint fmt vet deps docker-build docker-run migrate tools

# Help
help: ## Mostra esta ajuda
	@echo "$(GREEN)EquinoId Backend - Comandos disponíveis:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BLUE)%-20s$(NC) %s\n", $$1, $$2}'

# Build
build: ## Compila a aplicação
	@echo "$(YELLOW)Building $(APP_NAME)...$(NC)"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME) cmd/server/main.go
	@echo "$(GREEN)Build completed!$(NC)"

build-dev: ## Compila a aplicação para desenvolvimento
	@echo "$(YELLOW)Building $(APP_NAME) for development...$(NC)"
	@go build $(LDFLAGS) -o bin/$(APP_NAME) cmd/server/main.go
	@echo "$(GREEN)Development build completed!$(NC)"

# Run
run: ## Executa a aplicação
	@echo "$(YELLOW)Running $(APP_NAME)...$(NC)"
	@go run cmd/server/main.go

run-dev: build-dev ## Compila e executa em modo desenvolvimento
	@echo "$(YELLOW)Running $(APP_NAME) in development mode...$(NC)"
	@./bin/$(APP_NAME)

# Clean
clean: ## Remove arquivos de build
	@echo "$(YELLOW)Cleaning up...$(NC)"
	@rm -rf bin/
	@rm -rf tmp/
	@go clean
	@echo "$(GREEN)Cleanup completed!$(NC)"

# Test
test: ## Executa todos os testes
	@echo "$(YELLOW)Running tests...$(NC)"
	@go test -v ./...

test-coverage: ## Executa testes com coverage
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

test-race: ## Executa testes com race detector
	@echo "$(YELLOW)Running tests with race detector...$(NC)"
	@go test -race -v ./...

# Lint and Format
lint: ## Executa linter
	@echo "$(YELLOW)Running golangci-lint...$(NC)"
	@golangci-lint run

fmt: ## Formata o código
	@echo "$(YELLOW)Formatting code...$(NC)"
	@go fmt ./...

vet: ## Executa go vet
	@echo "$(YELLOW)Running go vet...$(NC)"
	@go vet ./...

# Dependencies
deps: ## Instala dependências
	@echo "$(YELLOW)Installing dependencies...$(NC)"
	@go mod tidy
	@go mod download

deps-upgrade: ## Atualiza dependências
	@echo "$(YELLOW)Upgrading dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy

# Docker
docker-build: ## Build da imagem Docker
	@echo "$(YELLOW)Building Docker image...$(NC)"
	@docker build -t $(DOCKER_IMAGE):$(VERSION) .
	@docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(VERSION)$(NC)"

docker-run: ## Executa container Docker
	@echo "$(YELLOW)Running Docker container...$(NC)"
	@docker run --rm -p 8080:8080 --name $(APP_NAME) $(DOCKER_IMAGE):latest

docker-push: docker-build ## Faz push da imagem para registry
	@echo "$(YELLOW)Pushing Docker image...$(NC)"
	@docker push $(DOCKER_IMAGE):$(VERSION)
	@docker push $(DOCKER_IMAGE):latest
	@echo "$(GREEN)Docker image pushed!$(NC)"

# Docker Compose
docker-compose-up: ## Inicia todos os serviços com docker-compose
	@echo "$(YELLOW)Starting services with docker-compose...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)Services started! Check http://localhost:8080/health$(NC)"

docker-compose-down: ## Para todos os serviços
	@echo "$(YELLOW)Stopping services...$(NC)"
	@docker-compose down
	@echo "$(GREEN)Services stopped!$(NC)"

docker-compose-logs: ## Mostra logs dos serviços
	@docker-compose logs -f

docker-compose-ps: ## Lista status dos serviços
	@docker-compose ps

docker-compose-rebuild: ## Rebuild completo dos serviços
	@echo "$(YELLOW)Rebuilding services...$(NC)"
	@docker-compose down
	@docker-compose build --no-cache
	@docker-compose up -d
	@echo "$(GREEN)Services rebuilt and started!$(NC)"

dev-up: ## Inicia ambiente de desenvolvimento
	@echo "$(YELLOW)Starting development environment...$(NC)"
	@docker-compose -f ../docker-compose.dev.yml up -d
	@echo "$(GREEN)Development environment started!$(NC)"

dev-down: ## Para ambiente de desenvolvimento
	@echo "$(YELLOW)Stopping development environment...$(NC)"
	@docker-compose -f ../docker-compose.dev.yml down
	@echo "$(GREEN)Development environment stopped!$(NC)"

dev-logs: ## Mostra logs do ambiente de desenvolvimento
	@docker-compose -f ../docker-compose.dev.yml logs -f

prod-up: ## Inicia ambiente de produção
	@echo "$(YELLOW)Starting production environment...$(NC)"
	@docker-compose -f ../docker-compose.yml up -d
	@echo "$(GREEN)Production environment started!$(NC)"

prod-down: ## Para ambiente de produção
	@echo "$(YELLOW)Stopping production environment...$(NC)"
	@docker-compose -f ../docker-compose.yml down
	@echo "$(GREEN)Production environment stopped!$(NC)"

# Database
migrate-up: ## Executa migrações do banco de dados
	@echo "$(YELLOW)Running database migrations...$(NC)"
	@go run cmd/migrate/main.go up

migrate-down: ## Reverte migrações do banco de dados
	@echo "$(YELLOW)Reverting database migrations...$(NC)"
	@go run cmd/migrate/main.go down

migrate-create: ## Cria nova migração
	@echo "$(YELLOW)Creating new migration...$(NC)"
	@read -p "Migration name: " name; \
	go run cmd/migrate/main.go create $$name

seed: ## Executa seeds do banco de dados
	@echo "$(YELLOW)Running database seeds...$(NC)"
	@go run cmd/seed/main.go

# Tools
tools: ## Instala ferramentas de desenvolvimento
	@echo "$(YELLOW)Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "$(GREEN)Development tools installed!$(NC)"

swagger: ## Gera documentação Swagger
	@echo "$(YELLOW)Generating Swagger documentation...$(NC)"
	@swag init -g cmd/server/main.go -o ./docs
	@echo "$(GREEN)Swagger documentation generated!$(NC)"

# Quality
check: lint vet test ## Executa todas as verificações de qualidade

ci: deps check test-coverage ## Pipeline de CI

# Security
sec-scan: ## Executa scanner de segurança
	@echo "$(YELLOW)Running security scanner...$(NC)"
	@gosec ./...

# Monitoring
metrics: ## Mostra métricas da aplicação
	@echo "$(YELLOW)Application metrics:$(NC)"
	@curl -s http://localhost:8080/metrics | grep -E "^# HELP|^equinoid_"

health: ## Verifica saúde da aplicação
	@echo "$(YELLOW)Checking application health...$(NC)"
	@curl -s http://localhost:8080/health | jq '.'

# Desenvolvimento
hot-reload: ## Executa com hot reload (requires air)
	@echo "$(YELLOW)Starting hot reload...$(NC)"
	@air

install-air: ## Instala ferramenta de hot reload
	@echo "$(YELLOW)Installing air for hot reload...$(NC)"
	@go install github.com/cosmtrek/air@latest

# Release
release: ## Cria release
	@echo "$(YELLOW)Creating release...$(NC)"
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "$(GREEN)Release $(VERSION) created!$(NC)"

# Default target
.DEFAULT_GOAL := help