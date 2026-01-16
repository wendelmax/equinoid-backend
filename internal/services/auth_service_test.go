package services

import (
	"context"
	"testing"

	"github.com/equinoid/backend/internal/config"
	"github.com/equinoid/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupAuthService(t *testing.T) (*AuthService, *gorm.DB) {
	db := setupTestDB(t)
	mockCache := new(MockCache)
	logger := newTestLogger()
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	service := NewAuthService(db, mockCache, logger, cfg)
	return service, db
}

func TestAuthService_Register(t *testing.T) {
	service, db := setupAuthService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	t.Run("Registro com sucesso", func(t *testing.T) {
		req := &models.RegisterRequest{
			Email:    "test@example.com",
			Password: "senha123",
			Name:     "Test User",
			UserType: "criador",
			CPFCNPJ:  "12345678900",
		}

		user, err := service.Register(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Email, user.Email)
		assert.Equal(t, req.Name, user.Name)
		assert.Empty(t, user.Password)
		assert.True(t, user.IsActive)
	})

	t.Run("Email duplicado", func(t *testing.T) {
		req := &models.RegisterRequest{
			Email:    "duplicate@example.com",
			Password: "senha123",
			Name:     "Test User",
			UserType: "criador",
		}

		_, err := service.Register(context.Background(), req)
		assert.NoError(t, err)

		_, err = service.Register(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email já está em uso")
	})
}

func TestAuthService_Login(t *testing.T) {
	service, db := setupAuthService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("senha123"), bcrypt.DefaultCost)
	user := &models.User{
		Email:    "login@example.com",
		Password: string(hashedPassword),
		Name:     "Test User",
		UserType: "criador",
		IsActive: true,
	}
	db.Create(user)

	t.Run("Login com sucesso", func(t *testing.T) {
		tokens, returnedUser, err := service.Login(context.Background(), "login@example.com", "senha123")

		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.NotNil(t, returnedUser)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
		assert.Equal(t, "Bearer", tokens.TokenType)
		assert.Empty(t, returnedUser.Password)
	})

	t.Run("Email inválido", func(t *testing.T) {
		_, _, err := service.Login(context.Background(), "naoexiste@example.com", "senha123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credenciais inválidas")
	})

	t.Run("Senha incorreta", func(t *testing.T) {
		_, _, err := service.Login(context.Background(), "login@example.com", "senhaerrada")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credenciais inválidas")
	})

	t.Run("Usuário inativo", func(t *testing.T) {
		inactiveUser := &models.User{
			Email:    "inactive@example.com",
			Password: string(hashedPassword),
			Name:     "Inactive User",
			UserType: "criador",
			IsActive: false,
		}
		db.Create(inactiveUser)

		_, _, err := service.Login(context.Background(), "inactive@example.com", "senha123")

		assert.Error(t, err)
	})
}

func TestAuthService_GenerateTokenPair(t *testing.T) {
	service, db := setupAuthService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	user := &models.User{
		ID:       1,
		Email:    "token@example.com",
		Name:     "Token User",
		UserType: "criador",
	}

	tokens, err := service.generateTokenPair(user)

	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Equal(t, "Bearer", tokens.TokenType)
	assert.Greater(t, tokens.ExpiresIn, int64(0))
}
