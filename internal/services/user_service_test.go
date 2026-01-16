package services

import (
	"context"
	"testing"

	"github.com/equinoid/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupUserService(t *testing.T) (*UserService, *gorm.DB) {
	db := setupTestDB(t)
	mockCache := new(MockCache)
	logger := newTestLogger()

	service := NewUserService(db, mockCache, logger)
	return service, db
}

func TestUserService_GetByID(t *testing.T) {
	service, db := setupUserService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	user := &models.User{
		Email:    "getbyid@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
		UserType: "criador",
		IsActive: true,
	}
	db.Create(user)

	t.Run("Usuário encontrado", func(t *testing.T) {
		foundUser, err := service.GetByID(context.Background(), user.ID)

		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Empty(t, foundUser.Password)
	})

	t.Run("Usuário não encontrado", func(t *testing.T) {
		_, err := service.GetByID(context.Background(), 99999)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "usuário não encontrado")
	})

	t.Run("Usuário inativo não retornado", func(t *testing.T) {
		inactiveUser := &models.User{
			Email:    "inactive@example.com",
			Password: "hashedpassword",
			Name:     "Inactive User",
			UserType: "criador",
			IsActive: false,
		}
		db.Create(inactiveUser)

		_, err := service.GetByID(context.Background(), inactiveUser.ID)

		assert.Error(t, err)
	})
}

func TestUserService_UpdateProfile(t *testing.T) {
	service, db := setupUserService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	user := &models.User{
		Email:    "update@example.com",
		Password: "hashedpassword",
		Name:     "Original Name",
		UserType: "criador",
		CPFCNPJ:  "11111111111",
		IsActive: true,
	}
	db.Create(user)

	t.Run("Atualização com sucesso", func(t *testing.T) {
		req := &models.UpdateProfileRequest{
			Name:    "Updated Name",
			CPFCNPJ: "22222222222",
		}

		updatedUser, err := service.UpdateProfile(context.Background(), user.ID, req)

		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, "Updated Name", updatedUser.Name)
		assert.Equal(t, "22222222222", updatedUser.CPFCNPJ)
	})

	t.Run("Atualização parcial", func(t *testing.T) {
		req := &models.UpdateProfileRequest{
			Name: "Only Name Changed",
		}

		updatedUser, err := service.UpdateProfile(context.Background(), user.ID, req)

		assert.NoError(t, err)
		assert.Equal(t, "Only Name Changed", updatedUser.Name)
	})

	t.Run("Usuário não encontrado", func(t *testing.T) {
		req := &models.UpdateProfileRequest{
			Name: "Test",
		}

		_, err := service.UpdateProfile(context.Background(), 99999, req)

		assert.Error(t, err)
	})
}

func TestUserService_ChangePassword(t *testing.T) {
	service, db := setupUserService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("senhaantiga"), bcrypt.DefaultCost)
	user := &models.User{
		Email:    "changepass@example.com",
		Password: string(hashedPassword),
		Name:     "Test User",
		UserType: "criador",
		IsActive: true,
	}
	db.Create(user)

	t.Run("Troca de senha com sucesso", func(t *testing.T) {
		err := service.ChangePassword(context.Background(), user.ID, "senhaantiga", "senhanova")

		assert.NoError(t, err)

		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		err = bcrypt.CompareHashAndPassword([]byte(updatedUser.Password), []byte("senhanova"))
		assert.NoError(t, err)
	})

	t.Run("Senha atual incorreta", func(t *testing.T) {
		err := service.ChangePassword(context.Background(), user.ID, "senhaerrada", "novasenha")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "senha atual incorreta")
	})

	t.Run("Usuário não encontrado", func(t *testing.T) {
		err := service.ChangePassword(context.Background(), 99999, "senha", "novasenha")

		assert.Error(t, err)
	})
}

func TestUserService_IsEmailAvailable(t *testing.T) {
	service, db := setupUserService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	user := &models.User{
		Email:    "existing@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
		UserType: "criador",
		IsActive: true,
	}
	db.Create(user)

	t.Run("Email já em uso", func(t *testing.T) {
		available, err := service.IsEmailAvailable(context.Background(), "existing@example.com")

		assert.NoError(t, err)
		assert.False(t, available)
	})

	t.Run("Email disponível", func(t *testing.T) {
		available, err := service.IsEmailAvailable(context.Background(), "new@example.com")

		assert.NoError(t, err)
		assert.True(t, available)
	})
}

func TestUserService_Delete(t *testing.T) {
	service, db := setupUserService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	user := &models.User{
		Email:    "delete@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
		UserType: "criador",
		IsActive: true,
	}
	db.Create(user)

	t.Run("Deleção soft delete", func(t *testing.T) {
		err := service.Delete(context.Background(), user.ID)

		assert.NoError(t, err)

		var deletedUser models.User
		result := db.Unscoped().First(&deletedUser, user.ID)
		assert.NoError(t, result.Error)
		assert.NotNil(t, deletedUser.DeletedAt)
	})

	t.Run("Usuário não encontrado", func(t *testing.T) {
		err := service.Delete(context.Background(), 99999)

		assert.Error(t, err)
	})
}
