package services

import (
	"context"
	"testing"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupReproducaoService(t *testing.T) (*ReproducaoService, *gorm.DB) {
	db := setupTestDB(t)
	mockCache := new(MockCache)
	logger := newTestLogger()

	service := NewReproducaoService(db, mockCache, logger)
	return service, db
}

func TestReproducaoService_CreateCobertura(t *testing.T) {
	service, db := setupReproducaoService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	dataNascimento := time.Now().AddDate(-5, 0, 0)

	reprodutor := &models.Equino{
		Equinoid:       "BRA-2015-00000001",
		MicrochipID:    "CHIP_REP",
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

	matriz := &models.Equino{
		Equinoid:       "BRA-2015-00000002",
		MicrochipID:    "CHIP_MAT",
		Nome:           "Matriz",
		Sexo:           models.SexoFemea,
		Raca:           "Mangalarga",
		Pelagem:        "Alazão",
		PaisOrigem:     "BRA",
		DataNascimento: &dataNascimento,
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(matriz)

	t.Run("Criar cobertura com sucesso", func(t *testing.T) {
		req := &models.CreateCoberturaRequest{
			DataCobertura:   time.Now(),
			TipoCobertura:   models.TipoCoberturaNatural,
			MetodoCobertura: "Monta natural",
		}

		cobertura, err := service.CreateCobertura(context.Background(), "BRA-2015-00000001", "BRA-2015-00000002", req, 1)

		assert.NoError(t, err)
		assert.NotNil(t, cobertura)
		assert.Equal(t, "BRA-2015-00000001", cobertura.ReprodutorEquinoid)
		assert.Equal(t, "BRA-2015-00000002", cobertura.MatrizEquinoid)
		assert.Equal(t, models.StatusCoberturaPendente, cobertura.StatusCobertura)
	})

	t.Run("Reprodutor deve ser macho", func(t *testing.T) {
		req := &models.CreateCoberturaRequest{
			DataCobertura: time.Now(),
			TipoCobertura: models.TipoCoberturaNatural,
		}

		_, err := service.CreateCobertura(context.Background(), "BRA-2015-00000002", "BRA-2015-00000001", req, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reprodutor deve ser macho")
	})

	t.Run("Matriz deve ser fêmea", func(t *testing.T) {
		req := &models.CreateCoberturaRequest{
			DataCobertura: time.Now(),
			TipoCobertura: models.TipoCoberturaNatural,
		}

		_, err := service.CreateCobertura(context.Background(), "BRA-2015-00000001", "BRA-2015-00000001", req, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "matriz deve ser fêmea")
	})
}

func TestReproducaoService_CreateAvaliacaoSemen(t *testing.T) {
	service, db := setupReproducaoService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	dataNascimento := time.Now().AddDate(-5, 0, 0)

	reprodutor := &models.Equino{
		Equinoid:       "BRA-2015-00000010",
		MicrochipID:    "CHIP_REP_SEMEN",
		Nome:           "Reprodutor Sêmen",
		Sexo:           models.SexoMacho,
		Raca:           "Mangalarga",
		Pelagem:        "Preto",
		PaisOrigem:     "BRA",
		DataNascimento: &dataNascimento,
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(reprodutor)

	motProgressiva := 70.0
	motTotal := 80.0
	morfologia := 75.0
	viabilidade := 85.0

	t.Run("Avaliação de sêmen excelente", func(t *testing.T) {
		req := &models.CreateAvaliacaoSemenRequest{
			DataColeta:            time.Now(),
			DataAnalise:           time.Now(),
			MotilidadeProgressiva: &motProgressiva,
			MotilidadeTotal:       &motTotal,
			MorfologiaNormal:      &morfologia,
			Viabilidade:           &viabilidade,
		}

		avaliacao, err := service.CreateAvaliacaoSemen(context.Background(), "BRA-2015-00000010", req, 1)

		assert.NoError(t, err)
		assert.NotNil(t, avaliacao)
		assert.Equal(t, models.QualidadeExcelente, avaliacao.QualidadeGeral)
		assert.Equal(t, models.AptidaoAlta, avaliacao.AptidaoReprodutiva)
	})

	t.Run("Apenas machos podem ter avaliação de sêmen", func(t *testing.T) {
		femea := &models.Equino{
			Equinoid:       "BRA-2015-00000011",
			MicrochipID:    "CHIP_FEM",
			Nome:           "Fêmea",
			Sexo:           models.SexoFemea,
			Raca:           "Mangalarga",
			Pelagem:        "Baio",
			PaisOrigem:     "BRA",
			DataNascimento: &dataNascimento,
			Status:         models.StatusAtivo,
			ProprietarioID: 1,
		}
		db.Create(femea)

		req := &models.CreateAvaliacaoSemenRequest{
			DataColeta:  time.Now(),
			DataAnalise: time.Now(),
		}

		_, err := service.CreateAvaliacaoSemen(context.Background(), "BRA-2015-00000011", req, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "só pode ser feita em machos")
	})
}

func TestReproducaoService_DetermineQualidadeSemen(t *testing.T) {
	service, db := setupReproducaoService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	tests := []struct {
		name        string
		motProg     float64
		motTotal    float64
		morfologia  float64
		viabilidade float64
		expected    models.QualidadeSemen
	}{
		{"Excelente", 70, 80, 75, 85, models.QualidadeExcelente},
		{"Boa", 50, 60, 60, 70, models.QualidadeBoa},
		{"Regular", 40, 50, 50, 60, models.QualidadeRegular},
		{"Ruim", 30, 40, 40, 50, models.QualidadeRuim},
		{"Inadequada", 20, 30, 30, 40, models.QualidadeInadequada},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &models.CreateAvaliacaoSemenRequest{
				MotilidadeProgressiva: &tt.motProg,
				MotilidadeTotal:       &tt.motTotal,
				MorfologiaNormal:      &tt.morfologia,
				Viabilidade:           &tt.viabilidade,
			}

			qualidade := service.determineQualidadeSemen(req)
			assert.Equal(t, tt.expected, qualidade)
		})
	}
}

func TestReproducaoService_CreateGestacao(t *testing.T) {
	service, db := setupReproducaoService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	dataNascimento := time.Now().AddDate(-5, 0, 0)

	reprodutor := &models.Equino{
		Equinoid:       "BRA-2015-00000020",
		MicrochipID:    "CHIP_REP_GEST",
		Nome:           "Reprodutor Gestação",
		Sexo:           models.SexoMacho,
		Raca:           "Mangalarga",
		Pelagem:        "Castanho",
		PaisOrigem:     "BRA",
		DataNascimento: &dataNascimento,
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(reprodutor)

	matriz := &models.Equino{
		Equinoid:       "BRA-2015-00000021",
		MicrochipID:    "CHIP_MAT_GEST",
		Nome:           "Matriz Gestação",
		Sexo:           models.SexoFemea,
		Raca:           "Mangalarga",
		Pelagem:        "Rosilho",
		PaisOrigem:     "BRA",
		DataNascimento: &dataNascimento,
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(matriz)

	cobertura := &models.Cobertura{
		ReprodutorEquinoid:     "BRA-2015-00000020",
		MatrizEquinoid:         "BRA-2015-00000021",
		DataCobertura:          time.Now(),
		TipoCobertura:          models.TipoCoberturaNatural,
		VeterinarioResponsavel: 1,
		StatusCobertura:        models.StatusCoberturaPendente,
	}
	db.Create(cobertura)

	t.Run("Criar gestação com sucesso", func(t *testing.T) {
		gestacao, err := service.CreateGestacao(context.Background(), cobertura.ID, 1)

		assert.NoError(t, err)
		assert.NotNil(t, gestacao)
		assert.Equal(t, cobertura.ID, gestacao.CoberturaID)
		assert.Equal(t, "BRA-2015-00000021", gestacao.MatrizEquinoid)
		assert.Equal(t, models.StatusGestacaoAtiva, gestacao.StatusGestacao)
	})

	t.Run("Não permitir gestação duplicada", func(t *testing.T) {
		_, err := service.CreateGestacao(context.Background(), cobertura.ID, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "já existe gestação")
	})

	t.Run("Cobertura não encontrada", func(t *testing.T) {
		_, err := service.CreateGestacao(context.Background(), 99999, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cobertura não encontrada")
	})
}
