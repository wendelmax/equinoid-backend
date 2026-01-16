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

func setupLinhagemService(t *testing.T) (*LinhagemService, *gorm.DB) {
	db := setupTestDB(t)
	mockCache := new(MockCache)
	logger := newTestLogger()

	service := NewLinhagemService(db, mockCache, logger)
	return service, db
}

func TestLinhagemService_GetArvoreGenealogica(t *testing.T) {
	service, db := setupLinhagemService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	dataNascimento := time.Now().AddDate(-10, 0, 0)

	avo := &models.Equino{
		Equinoid:       "BRA-2010-00000001",
		MicrochipID:    "CHIP_AVO",
		Nome:           "Avô",
		Sexo:           models.SexoMacho,
		Raca:           "Mangalarga",
		Pelagem:        "Alazão",
		PaisOrigem:     "BRA",
		DataNascimento: &dataNascimento,
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(avo)

	pai := &models.Equino{
		Equinoid:       "BRA-2015-00000001",
		MicrochipID:    "CHIP_PAI",
		Nome:           "Pai",
		Sexo:           models.SexoMacho,
		Raca:           "Mangalarga",
		Pelagem:        "Preto",
		PaisOrigem:     "BRA",
		DataNascimento: &dataNascimento,
		Genitor:        "BRA-2010-00000001",
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(pai)

	filho := &models.Equino{
		Equinoid:       "BRA-2020-00000001",
		MicrochipID:    "CHIP_FILHO",
		Nome:           "Filho",
		Sexo:           models.SexoMacho,
		Raca:           "Mangalarga",
		Pelagem:        "Tordilho",
		PaisOrigem:     "BRA",
		DataNascimento: &dataNascimento,
		Genitor:        "BRA-2015-00000001",
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(filho)

	t.Run("Obter árvore genealógica", func(t *testing.T) {
		arvore, err := service.GetArvoreGenealogica(context.Background(), "BRA-2020-00000001", 2)

		assert.NoError(t, err)
		assert.NotNil(t, arvore)
		assert.Equal(t, "BRA-2020-00000001", arvore.Equinoid)
		assert.Equal(t, 2, arvore.Geracoes)
	})

	t.Run("Equino não encontrado", func(t *testing.T) {
		_, err := service.GetArvoreGenealogica(context.Background(), "BRA-9999-99999999", 3)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "equino não encontrado")
	})

	t.Run("Limitar gerações a 10", func(t *testing.T) {
		arvore, err := service.GetArvoreGenealogica(context.Background(), "BRA-2020-00000001", 20)

		assert.NoError(t, err)
		assert.Equal(t, 10, arvore.Geracoes)
	})
}

func TestLinhagemService_ValidarParentesco(t *testing.T) {
	service, db := setupLinhagemService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	dataNascimento := time.Now().AddDate(-10, 0, 0)

	pai := &models.Equino{
		Equinoid:       "BRA-2015-00000010",
		MicrochipID:    "CHIP_PAI_2",
		Nome:           "Pai Comum",
		Sexo:           models.SexoMacho,
		Raca:           "Mangalarga",
		Pelagem:        "Baio",
		PaisOrigem:     "BRA",
		DataNascimento: &dataNascimento,
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(pai)

	filho1 := &models.Equino{
		Equinoid:       "BRA-2020-00000010",
		MicrochipID:    "CHIP_FILHO1",
		Nome:           "Filho 1",
		Sexo:           models.SexoMacho,
		Raca:           "Mangalarga",
		Pelagem:        "Alazão",
		PaisOrigem:     "BRA",
		DataNascimento: &dataNascimento,
		Genitor:        "BRA-2015-00000010",
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(filho1)

	filho2 := &models.Equino{
		Equinoid:       "BRA-2020-00000011",
		MicrochipID:    "CHIP_FILHO2",
		Nome:           "Filho 2",
		Sexo:           models.SexoFemea,
		Raca:           "Mangalarga",
		Pelagem:        "Branco",
		PaisOrigem:     "BRA",
		DataNascimento: &dataNascimento,
		Genitor:        "BRA-2015-00000010",
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(filho2)

	t.Run("Detectar irmãos", func(t *testing.T) {
		resultado, err := service.ValidarParentesco(context.Background(), "BRA-2020-00000010", "BRA-2020-00000011")

		assert.NoError(t, err)
		assert.True(t, resultado.SaoParentes)
		assert.Contains(t, resultado.GrauParentesco, "irmão")
	})

	t.Run("Não parentes", func(t *testing.T) {
		naoParente := &models.Equino{
			Equinoid:       "BRA-2020-00000020",
			MicrochipID:    "CHIP_NAO_PARENTE",
			Nome:           "Não Parente",
			Sexo:           models.SexoMacho,
			Raca:           "Mangalarga",
			Pelagem:        "Tordilho",
			PaisOrigem:     "BRA",
			DataNascimento: &dataNascimento,
			Status:         models.StatusAtivo,
			ProprietarioID: 1,
		}
		db.Create(naoParente)

		resultado, err := service.ValidarParentesco(context.Background(), "BRA-2020-00000010", "BRA-2020-00000020")

		assert.NoError(t, err)
		assert.False(t, resultado.SaoParentes)
		assert.Equal(t, "não relacionados", resultado.GrauParentesco)
		assert.Equal(t, 0.0, resultado.CoeficienteConsanguinidade)
	})
}

func TestLinhagemService_GetDescendentes(t *testing.T) {
	service, db := setupLinhagemService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	dataNascimento := time.Now().AddDate(-10, 0, 0)

	reprodutor := &models.Equino{
		Equinoid:       "BRA-2015-00000020",
		MicrochipID:    "CHIP_REPRODUTOR",
		Nome:           "Reprodutor",
		Sexo:           models.SexoMacho,
		Raca:           "Mangalarga",
		Pelagem:        "Preto",
		PaisOrigem:     "BRA",
		DataNascimento: &dataNascimento,
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(reprodutor)

	for i := 1; i <= 5; i++ {
		filho := &models.Equino{
			Equinoid:       fmt.Sprintf("BRA-2020-0000%04d", i+50),
			MicrochipID:    fmt.Sprintf("CHIP_DESC_%d", i),
			Nome:           fmt.Sprintf("Descendente %d", i),
			Sexo:           models.SexoMacho,
			Raca:           "Mangalarga",
			Pelagem:        "Alazão",
			PaisOrigem:     "BRA",
			DataNascimento: &dataNascimento,
			Genitor:        "BRA-2015-00000020",
			Status:         models.StatusAtivo,
			ProprietarioID: 1,
		}
		db.Create(filho)
	}

	t.Run("Listar descendentes", func(t *testing.T) {
		descendentes, err := service.GetDescendentes(context.Background(), "BRA-2015-00000020")

		assert.NoError(t, err)
		assert.Len(t, descendentes, 5)
	})

	t.Run("Sem descendentes", func(t *testing.T) {
		semDescendentes := &models.Equino{
			Equinoid:       "BRA-2020-00000100",
			MicrochipID:    "CHIP_SEM_DESC",
			Nome:           "Sem Descendentes",
			Sexo:           models.SexoMacho,
			Raca:           "Mangalarga",
			Pelagem:        "Branco",
			PaisOrigem:     "BRA",
			DataNascimento: &dataNascimento,
			Status:         models.StatusAtivo,
			ProprietarioID: 1,
		}
		db.Create(semDescendentes)

		descendentes, err := service.GetDescendentes(context.Background(), "BRA-2020-00000100")

		assert.NoError(t, err)
		assert.Len(t, descendentes, 0)
	})
}
