package leiloes

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/internal/modules/equinos"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
)

type Service interface {
	ListAll(ctx context.Context, leiloeiroID *uint) ([]*models.Leilao, error)
	GetByID(ctx context.Context, id uint) (*models.Leilao, error)
	
	ListParticipacoes(ctx context.Context, leilaoID uint) ([]*models.ParticipacaoLeilaoResponse, error)
	CriarParticipacao(ctx context.Context, leilaoID, criadorID uint, req *models.CreateParticipacaoLeilaoRequest) (*models.ParticipacaoLeilaoResponse, error)
	AprovarParticipacao(ctx context.Context, participacaoID uint) (*models.ParticipacaoLeilaoResponse, error)
	RegistrarVenda(ctx context.Context, participacaoID uint, req *models.RegistrarVendaRequest) (*models.ParticipacaoLeilaoResponse, error)
	MarcarAusencia(ctx context.Context, participacaoID uint) (*models.ParticipacaoLeilaoResponse, error)
	MarcarPresenca(ctx context.Context, participacaoID uint) (*models.ParticipacaoLeilaoResponse, error)
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

func (s *service) ListAll(ctx context.Context, leiloeiroID *uint) ([]*models.Leilao, error) {
	leiloes, err := s.repo.FindAll(ctx, leiloeiroID)
	if err != nil {
		s.logger.LogError(err, "LeilaoService.ListAll", logging.Fields{"leiloeiro_id": leiloeiroID})
		return nil, err
	}
	return leiloes, nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*models.Leilao, error) {
	leilao, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "LeilaoService.GetByID", logging.Fields{"leilao_id": id})
		}
		return nil, err
	}
	return leilao, nil
}

func (s *service) ListParticipacoes(ctx context.Context, leilaoID uint) ([]*models.ParticipacaoLeilaoResponse, error) {
	participacoes, err := s.repo.FindParticipacoesByLeilaoID(ctx, leilaoID)
	if err != nil {
		s.logger.LogError(err, "LeilaoService.ListParticipacoes", logging.Fields{"leilao_id": leilaoID})
		return nil, err
	}

	responses := make([]*models.ParticipacaoLeilaoResponse, len(participacoes))
	for i, p := range participacoes {
		responses[i] = p.ToResponse()
	}
	return responses, nil
}

func (s *service) CriarParticipacao(ctx context.Context, leilaoID, criadorID uint, req *models.CreateParticipacaoLeilaoRequest) (*models.ParticipacaoLeilaoResponse, error) {
	leilao, err := s.repo.FindByID(ctx, leilaoID)
	if err != nil {
		return nil, err
	}

	if leilao.Status == models.StatusLeilaoEncerrado || leilao.Status == models.StatusLeilaoCancelado {
		return nil, &apperrors.ValidationError{Message: "leilão não aceita mais inscrições"}
	}

	var equinoID uint
	equinoResp, err := s.equinoRepo.FindByEquinoid(ctx, req.Equinoid)
	if err != nil {
		return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado"}
	}
	equinoID = equinoResp.ID

	participacao := &models.ParticipacaoLeilao{
		LeilaoID:         leilaoID,
		EquinoID:         equinoID,
		CriadorID:        criadorID,
		ValorInicial:     req.ValorInicial,
		ValorReserva:     req.ValorReserva,
		Particularidades: req.Particularidades,
		Status:           models.StatusParticipacaoInscrito,
	}

	if err := s.repo.CreateParticipacao(ctx, participacao); err != nil {
		s.logger.LogError(err, "LeilaoService.CriarParticipacao", logging.Fields{
			"leilao_id":  leilaoID,
			"criador_id": criadorID,
		})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{
		"participacao_id": participacao.ID,
		"leilao_id":       leilaoID,
		"equino_id":       equinoID,
	}).Info("Participação em leilão criada com sucesso")

	return s.getParticipacaoResponse(ctx, participacao.ID)
}

func (s *service) AprovarParticipacao(ctx context.Context, participacaoID uint) (*models.ParticipacaoLeilaoResponse, error) {
	participacao, err := s.repo.FindParticipacaoByID(ctx, participacaoID)
	if err != nil {
		return nil, err
	}

	if participacao.Status != models.StatusParticipacaoInscrito {
		return nil, &apperrors.ValidationError{Message: "apenas participações inscritas podem ser aprovadas"}
	}

	participacao.Status = models.StatusParticipacaoAprovado

	if err := s.repo.UpdateParticipacao(ctx, participacao); err != nil {
		s.logger.LogError(err, "LeilaoService.AprovarParticipacao", logging.Fields{"participacao_id": participacaoID})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{"participacao_id": participacaoID}).Info("Participação aprovada")
	return s.getParticipacaoResponse(ctx, participacaoID)
}

func (s *service) RegistrarVenda(ctx context.Context, participacaoID uint, req *models.RegistrarVendaRequest) (*models.ParticipacaoLeilaoResponse, error) {
	participacao, err := s.repo.FindParticipacaoByID(ctx, participacaoID)
	if err != nil {
		return nil, err
	}

	if participacao.Status != models.StatusParticipacaoAprovado {
		return nil, &apperrors.ValidationError{Message: "apenas participações aprovadas podem ser vendidas"}
	}

	leilao := participacao.Leilao
	if leilao == nil {
		leilao, err = s.repo.FindByID(ctx, participacao.LeilaoID)
		if err != nil {
			return nil, err
		}
	}

	participacao.ValorVendido = &req.ValorVendido
	participacao.ValorFinal = &req.ValorVendido
	participacao.CompradorID = &req.CompradorID
	participacao.Status = models.StatusParticipacaoVendido

	comissaoPercentual := req.ValorVendido * (leilao.TaxaComissaoPercentual / 100)
	comissaoTotal := comissaoPercentual
	if leilao.TaxaFixa != nil {
		comissaoTotal += *leilao.TaxaFixa
	}
	participacao.ComissaoLeiloeiro = &comissaoTotal

	if err := s.repo.UpdateParticipacao(ctx, participacao); err != nil {
		s.logger.LogError(err, "LeilaoService.RegistrarVenda", logging.Fields{"participacao_id": participacaoID})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{
		"participacao_id": participacaoID,
		"valor_vendido":   req.ValorVendido,
		"comissao":        comissaoTotal,
	}).Info("Venda registrada com sucesso")

	return s.getParticipacaoResponse(ctx, participacaoID)
}

func (s *service) MarcarAusencia(ctx context.Context, participacaoID uint) (*models.ParticipacaoLeilaoResponse, error) {
	participacao, err := s.repo.FindParticipacaoByID(ctx, participacaoID)
	if err != nil {
		return nil, err
	}

	leilao := participacao.Leilao
	if leilao == nil {
		leilao, err = s.repo.FindByID(ctx, participacao.LeilaoID)
		if err != nil {
			return nil, err
		}
	}

	if leilao.TipoLeilao != models.TipoLeilaoPresencial {
		return nil, &apperrors.ValidationError{Message: "controle de presença só é válido para leilões presenciais"}
	}

	compareceu := false
	penalizacao := -50
	participacao.Compareceu = &compareceu
	participacao.PenalizacaoAusencia = &penalizacao

	if err := s.repo.UpdateParticipacao(ctx, participacao); err != nil {
		s.logger.LogError(err, "LeilaoService.MarcarAusencia", logging.Fields{"participacao_id": participacaoID})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{
		"participacao_id": participacaoID,
		"penalizacao":     penalizacao,
	}).Info("Ausência marcada com penalização")

	return s.getParticipacaoResponse(ctx, participacaoID)
}

func (s *service) MarcarPresenca(ctx context.Context, participacaoID uint) (*models.ParticipacaoLeilaoResponse, error) {
	participacao, err := s.repo.FindParticipacaoByID(ctx, participacaoID)
	if err != nil {
		return nil, err
	}

	compareceu := true
	participacao.Compareceu = &compareceu
	participacao.PenalizacaoAusencia = nil

	if err := s.repo.UpdateParticipacao(ctx, participacao); err != nil {
		s.logger.LogError(err, "LeilaoService.MarcarPresenca", logging.Fields{"participacao_id": participacaoID})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{"participacao_id": participacaoID}).Info("Presença marcada")
	return s.getParticipacaoResponse(ctx, participacaoID)
}

func (s *service) getParticipacaoResponse(ctx context.Context, id uint) (*models.ParticipacaoLeilaoResponse, error) {
	participacao, err := s.repo.FindParticipacaoByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return participacao.ToResponse(), nil
}
