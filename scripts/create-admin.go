package main

import (
	"context"
	"fmt"
	"os"

	"github.com/equinoid/backend/internal/database"
	"github.com/equinoid/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Uso: go run scripts/create-admin.go <email> <senha> <nome> [cpf_cnpj]")
		fmt.Println("Exemplo: go run scripts/create-admin.go admin@equinoid.com senha123 Admin 12345678900")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]
	name := os.Args[3]
	cpfCnpj := "00000000000"
	if len(os.Args) >= 5 {
		cpfCnpj = os.Args[4]
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Println("ERRO: DATABASE_URL não configurada")
		fmt.Println("Configure a variável de ambiente DATABASE_URL")
		os.Exit(1)
	}

	db, err := database.NewConnection(dbURL)
	if err != nil {
		fmt.Printf("ERRO ao conectar ao banco: %v\n", err)
		os.Exit(1)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("ERRO ao gerar hash da senha: %v\n", err)
		os.Exit(1)
	}

	admin := &models.User{
		Email:           email,
		Password:        string(hashedPassword),
		Name:            name,
		UserType:        models.UserTypeAdmin,
		CPFCNPJ:         cpfCnpj,
		IsActive:        true,
		IsEmailVerified: true,
		Role:            "admin",
	}

	ctx := context.Background()
	result := db.WithContext(ctx).Create(admin)
	if result.Error != nil {
		fmt.Printf("ERRO ao criar admin: %v\n", result.Error)
		os.Exit(1)
	}

	fmt.Println("✅ Usuário admin criado com sucesso!")
	fmt.Printf("   ID: %d\n", admin.ID)
	fmt.Printf("   Email: %s\n", admin.Email)
	fmt.Printf("   Nome: %s\n", admin.Name)
	fmt.Printf("   Tipo: %s\n", admin.UserType)
	fmt.Println("\nVocê pode fazer login com essas credenciais em /api/v1/auth/login")
}
