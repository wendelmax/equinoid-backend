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

func setupValorizacaoService(t *testing.T) (*ValorizacaoService, *gorm.DB) {
	db := setupTestDB(t)
	mockCache := new(MockCache)
	logger := newTestLogger()

	service := NewValorizacaoService(db, mockCache, logger)
	return service, db
}

func TestValorizacaoService_CreateRegistro(t *testing.T) {
	service, db := setupValorizacaoService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	dataNascimento := time.Now().AddDate(-5, 0, 0)
	equino := &models.Equino{
		Equinoid:       "BRA-2020-00000001",
		MicrochipID:    "CHIP001",
		Nome:           "Cavalo Teste",
		Sexo:           models.SexoMacho,
		Raca:           "Mangalarga",
		DataNascimento: &dataNascimento,
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(equino)

	user := &models.User{
		Email:    "user@example.com",
		Password: "hash",
		Name:     "User",
		UserType: "criador",
		IsActive: true,
	}
	db.Create(user)

	t.Run("Criar registro com sucesso", func(t *testing.T) {
		req := &models.CreateValorizacaoRequest{
			Categoria:        models.CategoriaCompeticao,
			TipoRegistro:     "Campeonato Nacional",
			Titulo:           "1º Lugar",
			Descricao:        "Campeão nacional",
			DataRegistro:     time.Now(),
			NivelImportancia: models.NivelAlto,
		}

		registro, err := service.CreateRegistro(context.Background(), "BRA-2020-00000001", req, user.ID)

		assert.NoError(t, err)
		assert.NotNil(t, registro)
		assert.Equal(t, "BRA-2020-00000001", registro.Equinoid)
		assert.Equal(t, models.StatusPendente, registro.StatusValidacao)
		assert.Greater(t, registro.PontosValorizacao, 0)
	})

	t.Run("Equino não encontrado", func(t *testing.T) {
		req := &models.CreateValorizacaoRequest{
			Categoria:        models.CategoriaCompeticao,
			TipoRegistro:     "Teste",
			Titulo:           "Teste",
			DataRegistro:     time.Now(),
			NivelImportancia: models.NivelMedio,
		}

		_, err := service.CreateRegistro(context.Background(), "BRA-9999-99999999", req, user.ID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "equino não encontrado")
	})
}

func TestValorizacaoService_CalculatePoints(t *testing.T) {
	service, db := setupValorizacaoService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	tests := []struct {
		categoria models.CategoriaValorizacao
		nivel     models.NivelImportancia
		expected  int
	}{
		{models.CategoriaCompeticao, models.NivelMedio, 100},
		{models.CategoriaCompeticao, models.NivelAlto, 200},
		{models.CategoriaCompeticao, models.NivelExcepcional, 500},
		{models.CategoriaReproducao, models.NivelAlto, 160},
		{models.CategoriaSaude, models.NivelBaixo, 25},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s-%s", tt.categoria, tt.nivel), func(t *testing.T) {
			pontos := service.calculatePoints(tt.categoria, "Teste", tt.nivel)
			assert.Equal(t, tt.expected, pontos)
		})
	}
}

func TestValorizacaoService_GetTotalPoints(t *testing.T) {
	service, db := setupValorizacaoService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	dataNascimento := time.Now().AddDate(-5, 0, 0)
	equino := &models.Equino{
		Equinoid:       "BRA-2020-00000002",
		MicrochipID:    "CHIP002",
		Nome:           "Cavalo Pontos",
		Sexo:           models.SexoMacho,
		Raca:           "Mangalarga",
		DataNascimento: &dataNascimento,
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(equino)

	registro1 := &models.RegistroValorizacao{
		Equinoid:          "BRA-2020-00000002",
		Categoria:         models.CategoriaCompeticao,
		TipoRegistro:      "Teste 1",
		Titulo:            "Titulo 1",
		DataRegistro:      time.Now(),
		PontosValorizacao: 100,
		NivelImportancia:  models.NivelMedio,
		StatusValidacao:   models.StatusAprovado,
		CriadoPor:         1,
	}
	db.Create(registro1)

	registro2 := &models.RegistroValorizacao{
		Equinoid:          "BRA-2020-00000002",
		Categoria:         models.CategoriaCompeticao,
		TipoRegistro:      "Teste 2",
		Titulo:            "Titulo 2",
		DataRegistro:      time.Now(),
		PontosValorizacao: 200,
		NivelImportancia:  models.NivelAlto,
		StatusValidacao:   models.StatusAprovado,
		CriadoPor:         1,
	}
	db.Create(registro2)

	registro3 := &models.RegistroValorizacao{
		Equinoid:          "BRA-2020-00000002",
		Categoria:         models.CategoriaCompeticao,
		TipoRegistro:      "Teste 3",
		Titulo:            "Titulo 3",
		DataRegistro:      time.Now(),
		PontosValorizacao: 50,
		NivelImportancia:  models.NivelBaixo,
		StatusValidacao:   models.StatusPendente,
		CriadoPor:         1,
	}
	db.Create(registro3)

	t.Run("Calcular pontos totais (apenas aprovados)", func(t *testing.T) {
		total, err := service.GetTotalPoints(context.Background(), "BRA-2020-00000002")

		assert.NoError(t, err)
		assert.Equal(t, 300, total)
	})
}

func TestValorizacaoService_ValidateRegistro(t *testing.T) {
	service, db := setupValorizacaoService(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	dataNascimento := time.Now().AddDate(-5, 0, 0)
	equino := &models.Equino{
		Equinoid:       "BRA-2020-00000003",
		MicrochipID:    "CHIP003",
		Nome:           "Cavalo Validação",
		Sexo:           models.SexoMacho,
		Raca:           "Mangalarga",
		DataNascimento: &dataNascimento,
		Status:         models.StatusAtivo,
		ProprietarioID: 1,
	}
	db.Create(equino)

	registro := &models.RegistroValorizacao{
		Equinoid:          "BRA-2020-00000003",
		Categoria:         models.CategoriaCompeticao,
		TipoRegistro:      "Teste",
		Titulo:            "Teste",
		DataRegistro:      time.Now(),
		PontosValorizacao: 100,
		NivelImportancia:  models.NivelMedio,
		StatusValidacao:   models.StatusPendente,
		CriadoPor:         1,
	}
	db.Create(registro)

	t.Run("Aprovar registro", func(t *testing.T) {
		err := service.ValidateRegistro(context.Background(), registro.ID, 2, true, "Aprovado")

		assert.NoError(t, err)

		var updated models.RegistroValorizacao
		db.First(&updated, registro.ID)
		assert.Equal(t, models.StatusAprovado, updated.StatusValidacao)
		assert.NotNil(t, updated.ValidadoPor)
		assert.NotNil(t, updated.DataValidacao)
	})

	t.Run("Rejeitar registro", func(t *testing.T) {
		registro2 := &models.RegistroValorizacao{
			Equinoid:          "BRA-2020-00000003",
			Categoria:         models.CategoriaCompeticao,
			TipoRegistro:      "Teste 2",
			Titulo:            "Teste 2",
			DataRegistro:      time.Now(),
			PontosValorizacao: 100,
			NivelImportancia:  models.NivelMedio,
			StatusValidacao:   models.StatusPendente,
			CriadoPor:         1,
		}
		db.Create(registro2)

		err := service.ValidateRegistro(context.Background(), registro2.ID, 2, false, "Documentação insuficiente")

		assert.NoError(t, err)

		var updated models.RegistroValorizacao
		db.First(&updated, registro2.ID)
		assert.Equal(t, models.StatusRejeitado, updated.StatusValidacao)
		assert.Equal(t, "Documentação insuficiente", updated.ObservacoesValidacao)
	})
}
