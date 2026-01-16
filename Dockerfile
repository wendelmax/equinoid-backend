# Multi-stage build para otimização
FROM golang:1.24-alpine AS builder

# Instalar dependências necessárias
RUN apk add --no-cache git ca-certificates tzdata

# Criar usuário não-root
RUN adduser -D -g '' appuser

# Definir diretório de trabalho
WORKDIR /build

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Download das dependências
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação com otimizações
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o equinoid-api \
    cmd/server/main.go

# Estágio final - imagem mínima
FROM scratch

# Importar certificados e timezone da imagem builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copiar binário da aplicação
COPY --from=builder /build/equinoid-api /equinoid-api

# Copiar migrations se existirem
COPY --from=builder /build/migrations /migrations

# Usar usuário não-root
USER appuser

# Expor porta
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/equinoid-api", "--health-check"]

# Comando de execução
ENTRYPOINT ["/equinoid-api"]