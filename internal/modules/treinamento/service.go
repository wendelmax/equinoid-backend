package treinamento

import (
	"context"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/internal/modules/equinos"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
)

type Service interface {
	GetSessoesByEquinoid(ctx context.Context, equinoid string) ([]*models.SessaoTreinamento, error)
	GetProgramasByEquinoid(ctx context.Context, equinoid string) ([]*models.ProgramaTreinamento, error)
	CreateSessao(ctx context.Context, treinadorID uint, req *models.CreateSessaoTreinamentoRequest) (*models.SessaoTreinamento, error)
	CreatePrograma(ctx context.Context, treinadorID uint, req *models.CreateProgramaTreinamentoRequest) (*models.ProgramaTreinamento, error)
}

type service struct {
	repo       Repository
	equinoRepo equinos.Repository
	logger     *logging.Logger
}

func NewService(repo Repository, equinoRepo equinos.Repository, logger *logging.Logger) Service {
	return &service{
		repo:       repo,
		equinoRepo: equinoRepo,
		logger:     logger,
	}
}

func (s *service) GetSessoesByEquinoid(ctx context.Context, equinoid string) ([]*models.SessaoTreinamento, error) {
	sessoes, err := s.repo.FindSessoesByEquinoid(ctx, equinoid)
	if err != nil {
		s.logger.LogError(err, "TreinamentoService.GetSessoesByEquinoid", logging.Fields{"equinoid": equinoid})
		return nil, err
	}
	return sessoes, nil
}

func (s *service) GetProgramasByEquinoid(ctx context.Context, equinoid string) ([]*models.ProgramaTreinamento, error) {
	programas, err := s.repo.FindProgramasByEquinoid(ctx, equinoid)
	if err != nil {
		s.logger.LogError(err, "TreinamentoService.GetProgramasByEquinoid", logging.Fields{"equinoid": equinoid})
		return nil, err
	}
	return programas, nil
}

func (s *service) CreateSessao(ctx context.Context, treinadorID uint, req *models.CreateSessaoTreinamentoRequest) (*models.SessaoTreinamento, error) {
	equino, err := s.equinoRepo.FindByEquinoid(ctx, req.Equinoid)
	if err != nil {
		return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado"}
	}

	dataSessao := time.Now()
	if req.DataSessao != nil {
		dataSessao = *req.DataSessao
	}

	var exerciciosJSON models.JSONB
	if len(req.ExerciciosRealizados) > 0 {
		exerciciosJSON = make(models.JSONB)
		itens := make([]interface{}, len(req.ExerciciosRealizados))
		for i, ex := range req.ExerciciosRealizados {
			itens[i] = map[string]interface{}{
				"nome":       ex.Nome,
				"series":     ex.Series,
				"repeticoes": ex.Repeticoes,
				"duracao":    ex.Duracao,
			}
		}
		exerciciosJSON["exercicios"] = itens
	}

	sessao := &models.SessaoTreinamento{
		ProgramaTreinamentoID:   req.ProgramaTreinamentoID,
		EquinoID:                equino.ID,
		TreinadorID:             treinadorID,
		DataSessao:              dataSessao,
		Modalidade:              req.Modalidade,
		DuracaoMinutos:          req.DuracaoMinutos,
		Intensidade:             req.Intensidade,
		Distancia:               req.Distancia,
		VelocidadeMedia:         req.VelocidadeMedia,
		FrequenciaCardiacaMedia: req.FrequenciaCardiacaMedia,
		CaloriasGastas:          req.CaloriasGastas,
		ExerciciosRealizados:    exerciciosJSON,
		DesempenhoGeral:         req.DesempenhoGeral,
		Observacoes:             req.Observacoes,
		CondicoesClimaticas:     req.CondicoesClimaticas,
		TemperaturaC:            req.TemperaturaC,
	}

	if err := s.repo.CreateSessao(ctx, sessao); err != nil {
		s.logger.LogError(err, "TreinamentoService.CreateSessao", logging.Fields{"equinoid": req.Equinoid})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{
		"sessao_id":   sessao.ID,
		"equinoid":    req.Equinoid,
		"modalidade":  req.Modalidade,
		"desempenho":  req.DesempenhoGeral,
	}).Info("Sessão de treinamento registrada")

	return sessao, nil
}

func (s *service) CreatePrograma(ctx context.Context, treinadorID uint, req *models.CreateProgramaTreinamentoRequest) (*models.ProgramaTreinamento, error) {
	equino, err := s.equinoRepo.FindByEquinoid(ctx, req.Equinoid)
	if err != nil {
		return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado"}
	}

	var modalidadesJSON models.JSONB
	if len(req.Modalidades) > 0 {
		modalidadesJSON = make(models.JSONB)
		itens := make([]interface{}, len(req.Modalidades))
		for i, m := range req.Modalidades {
			itens[i] = m
		}
		modalidadesJSON["modalidades"] = itens
	}

	dataFim := time.Now().AddDate(0, 0, req.DuracaoSemanas*7)

	programa := &models.ProgramaTreinamento{
		EquinoID:          equino.ID,
		TreinadorID:       treinadorID,
		NomePrograma:      req.NomePrograma,
		Objetivo:          req.Objetivo,
		TipoPrograma:      req.TipoPrograma,
		Intensidade:       req.Intensidade,
		DuracaoSemanas:    req.DuracaoSemanas,
		FrequenciaSemanal: req.FrequenciaSemanal,
		DuracaoSessaoMin:  req.DuracaoSessaoMin,
		Modalidades:       modalidadesJSON,
		Observacoes:       req.Observacoes,
		DataInicio:        time.Now(),
		DataFim:           &dataFim,
		Status:            models.StatusProgramaAtivo,
	}

	if err := s.repo.CreatePrograma(ctx, programa); err != nil {
		s.logger.LogError(err, "TreinamentoService.CreatePrograma", logging.Fields{"equinoid": req.Equinoid})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{
		"programa_id": programa.ID,
		"equinoid":    req.Equinoid,
		"nome":        req.NomePrograma,
		"duracao":     req.DuracaoSemanas,
	}).Info("Programa de treinamento criado")

	return programa, nil
}
