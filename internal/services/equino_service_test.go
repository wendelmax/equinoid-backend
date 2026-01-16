package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupEquinoService(t *testing.T) (*EquinoService, *gorm.DB) {
	db := setupTestDB(t)
	mockCache := new(MockCache)
	logger := newTestLogger()

	// D4SignService não é necessário para os testes, passar nil
	service := NewEquinoService(db, mockCache, logger, nil)
	return service, db
}

func TestEquinoService_Create(t *testing.T) {
	service, db := setupEquinoService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	user := &models.User{
		Email:    "owner@example.com",
		Password: "hashedpassword",
		Name:     "Owner",
		UserType: "criador",
		IsActive: true,
	}
	db.Create(user)

	dataNascimento := time.Now().AddDate(-5, 0, 0)

	t.Run("Criação com sucesso", func(t *testing.T) {
		req := &models.CreateEquinoRequest{
			MicrochipID:    "CHIP001",
			Nome:           "Cavalo Teste",
			Sexo:           models.SexoMacho,
			Raca:           "Mangalarga",
			Pelagem:        "Alazão",
			DataNascimento: &dataNascimento,
			PaisOrigem:     "BRA",
		}

		equino, err := service.Create(context.Background(), req, user.ID)

		assert.NoError(t, err)
		assert.NotNil(t, equino)
		assert.NotEmpty(t, equino.Equinoid)
		assert.Equal(t, req.Nome, equino.Nome)
		assert.Equal(t, models.StatusAtivo, equino.Status)
	})

	t.Run("EquinoID duplicado", func(t *testing.T) {
		req := &models.CreateEquinoRequest{
			MicrochipID:    "CHIP002",
			Nome:           "Cavalo 1",
			Sexo:           models.SexoMacho,
			Raca:           "Mangalarga",
			Pelagem:        "Preto",
			DataNascimento: &dataNascimento,
			PaisOrigem:     "BRA",
		}

		_, err := service.Create(context.Background(), req, user.ID)
		assert.NoError(t, err)

		req2 := &models.CreateEquinoRequest{
			MicrochipID:    "CHIP003",
			Nome:           "Cavalo 2",
			Sexo:           models.SexoFemea,
			Raca:           "Mangalarga",
			Pelagem:        "Branco",
			DataNascimento: &dataNascimento,
			PaisOrigem:     "BRA",
		}

		_, err = service.Create(context.Background(), req2, user.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EquinoID já existe")
	})

	t.Run("MicrochipID duplicado", func(t *testing.T) {
		req := &models.CreateEquinoRequest{
			MicrochipID:    "CHIP_DUP",
			Nome:           "Cavalo 3",
			Sexo:           models.SexoMacho,
			Raca:           "Mangalarga",
			Pelagem:        "Tordilho",
			DataNascimento: &dataNascimento,
			PaisOrigem:     "BRA",
		}

		_, err := service.Create(context.Background(), req, user.ID)
		assert.NoError(t, err)

		req2 := &models.CreateEquinoRequest{
			MicrochipID:    "CHIP_DUP",
			Nome:           "Cavalo 4",
			Sexo:           models.SexoFemea,
			Raca:           "Mangalarga",
			Pelagem:        "Baio",
			DataNascimento: &dataNascimento,
			PaisOrigem:     "BRA",
		}

		_, err = service.Create(context.Background(), req2, user.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MicrochipID já existe")
	})
}

func TestEquinoService_GetByEquinoidID(t *testing.T) {
	service, db := setupEquinoService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	dataNascimento := time.Now().AddDate(-5, 0, 0)
	equino := &models.Equino{
		Equinoid:       "BRA-2020-00000010",
		MicrochipID:    "CHIP010",
		Nome:           "Cavalo Get",
		Sexo:           models.SexoMacho,
		Raca:           "Mangalarga",
		Pelagem:        "Preto",
		PaisOrigem:     "BRA",
		DataNascimento: &dataNascimento,
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(equino)

	t.Run("Equino encontrado", func(t *testing.T) {
		foundEquino, err := service.GetByEquinoid(context.Background(), "BRA-2020-00000010")

		assert.NoError(t, err)
		assert.NotNil(t, foundEquino)
		assert.Equal(t, equino.Nome, foundEquino.Nome)
	})

	t.Run("Equino não encontrado", func(t *testing.T) {
		_, err := service.GetByEquinoid(context.Background(), "BRA-9999-99999999")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "equino não encontrado")
	})
}

func TestEquinoService_List(t *testing.T) {
	service, db := setupEquinoService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	dataNascimento := time.Now().AddDate(-5, 0, 0)

	for i := 1; i <= 15; i++ {
		equino := &models.Equino{
			Equinoid:       fmt.Sprintf("BRA-2020-%08d", i+100),
			MicrochipID:    fmt.Sprintf("CHIP%03d", i+100),
			Nome:           fmt.Sprintf("Cavalo %d", i),
			Sexo:           models.SexoMacho,
			Raca:           "Mangalarga",
			Pelagem:        "Alazão",
			PaisOrigem:     "BRA",
			DataNascimento: &dataNascimento,
			Status:         models.StatusAtivo,
			ProprietarioID: 1,
		}
		db.Create(equino)
	}

	t.Run("Listagem paginada", func(t *testing.T) {
		equinos, total, err := service.List(context.Background(), 1, 10, map[string]interface{}{})

		assert.NoError(t, err)
		assert.Len(t, equinos, 10)
		assert.Equal(t, int64(15), total)
	})

	t.Run("Filtro por nome", func(t *testing.T) {
		result, total, err := service.List(context.Background(), 1, 10, map[string]interface{}{
			"nome": "Cavalo 1",
		})

		assert.NoError(t, err)
		assert.Greater(t, len(result), 0)
		assert.Greater(t, total, int64(0))
	})

	t.Run("Filtro por raça", func(t *testing.T) {
		_, total, err := service.List(context.Background(), 1, 10, map[string]interface{}{
			"raca": "Mangalarga",
		})

		assert.NoError(t, err)
		assert.Equal(t, int64(15), total)
	})

	t.Run("Filtro por sexo", func(t *testing.T) {
		_, total, err := service.List(context.Background(), 1, 10, map[string]interface{}{
			"sexo": "macho",
		})

		assert.NoError(t, err)
		assert.Greater(t, total, int64(0))
	})
}
