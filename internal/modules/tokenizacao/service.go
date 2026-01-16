package tokenizacao

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/internal/modules/equinos"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
)

type Service interface {
	ListAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.TokenizacaoResponse, int64, error)
	GetByID(ctx context.Context, id uint) (*models.TokenizacaoResponse, error)
	GetByEquinoid(ctx context.Context, equinoid string) (*models.TokenizacaoResponse, error)
	Create(ctx context.Context, userID uint, req *models.CreateTokenizacaoRequest) (*models.TokenizacaoResponse, error)
	
	ListTransacoes(ctx context.Context, tokenizacaoID uint) ([]*models.TransacaoTokenResponse, error)
	ExecutarOrdem(ctx context.Context, userID uint, req *models.OrdemCompraTokenRequest) (*models.TransacaoTokenResponse, error)
	CriarOferta(ctx context.Context, userID uint, req *models.OfertaTokenRequest) error
	
	CalcularRatingRisco(ctx context.Context, equinoID uint) (models.RatingRisco, error)
}

type service struct {
	repo        Repository
	equinoRepo  equinos.Repository
	logger      *logging.Logger
}

func NewService(repo Repository, equinoRepo equinos.Repository, logger *logging.Logger) Service {
	return &service{
		repo:       repo,
		equinoRepo: equinoRepo,
		logger:     logger,
	}
}

func (s *service) ListAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.TokenizacaoResponse, int64, error) {
	tokenizacoes, total, err := s.repo.FindAll(ctx, page, limit, filters)
	if err != nil {
		s.logger.LogError(err, "TokenizacaoService.ListAll", logging.Fields{"filters": filters})
		return nil, 0, err
	}

	responses := make([]*models.TokenizacaoResponse, len(tokenizacoes))
	for i, t := range tokenizacoes {
		responses[i] = t.ToResponse()
		
		participacoes, _ := s.repo.FindParticipacoesByTokenizacaoID(ctx, t.ID)
		responses[i].NumeroInvestidores = len(participacoes)
	}

	return responses, total, nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*models.TokenizacaoResponse, error) {
	tokenizacao, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "TokenizacaoService.GetByID", logging.Fields{"id": id})
		}
		return nil, err
	}

	response := tokenizacao.ToResponse()
	
	participacoes, _ := s.repo.FindParticipacoesByTokenizacaoID(ctx, id)
	response.NumeroInvestidores = len(participacoes)

	return response, nil
}

func (s *service) GetByEquinoid(ctx context.Context, equinoid string) (*models.TokenizacaoResponse, error) {
	tokenizacao, err := s.repo.FindByEquinoid(ctx, equinoid)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "TokenizacaoService.GetByEquinoid", logging.Fields{"equinoid": equinoid})
		}
		return nil, err
	}

	response := tokenizacao.ToResponse()
	
	participacoes, _ := s.repo.FindParticipacoesByTokenizacaoID(ctx, tokenizacao.ID)
	response.NumeroInvestidores = len(participacoes)

	return response, nil
}

func (s *service) Create(ctx context.Context, userID uint, req *models.CreateTokenizacaoRequest) (*models.TokenizacaoResponse, error) {
	_, err := s.equinoRepo.FindByID(ctx, req.EquinoID)
	if err != nil {
		return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado"}
	}

	existing, _ := s.repo.FindByEquinoID(ctx, req.EquinoID)
	if existing != nil {
		return nil, &apperrors.ValidationError{Message: "equino já está tokenizado"}
	}

	if req.PercentualMinimoDono < 51 {
		return nil, &apperrors.ValidationError{
			Message: "percentual mínimo do dono deve ser no mínimo 51% (compliance regulatório)",
		}
	}

	if req.PercentualMinimoDono+req.PercentualComercializavelPublicamente > 100 {
		return nil, &apperrors.ValidationError{
			Message: "soma dos percentuais não pode exceder 100%",
		}
	}

	if req.TemSeguro && req.ValorAssegurado == nil {
		return nil, &apperrors.ValidationError{
			Message: "valor assegurado é obrigatório quando tem_seguro = true",
		}
	}

	rating, err := s.CalcularRatingRisco(ctx, req.EquinoID)
	if err != nil {
		rating = models.RatingA
	}

	tokensBloqueados := int(float64(req.TotalTokens) * req.PercentualMinimoDono / 100)
	tokensDisponiveis := req.TotalTokens - tokensBloqueados

	tokenizacao := &models.Tokenizacao{
		EquinoID:                              req.EquinoID,
		TotalTokens:                           req.TotalTokens,
		TokensBloqueadosDono:                  tokensBloqueados,
		TokensDisponiveisVenda:                tokensDisponiveis,
		TokensVendidos:                        0,
		PrecoInicialToken:                     req.PrecoInicialToken,
		ValorTotalTokenizado:                  float64(req.TotalTokens) * req.PrecoInicialToken,
		PercentualMinimoDono:                  req.PercentualMinimoDono,
		PercentualComercializavelPublicamente: req.PercentualComercializavelPublicamente,
		TravaControleDono:                     req.TravaControleDono,
		PrioridadeRecompra:                    req.PrioridadeRecompra,
		Status:                                models.StatusTokenPendente,
		CustoCustodiaMensal:                   *req.CustoCustodiaMensal,
		TemSeguro:                             req.TemSeguro,
		ValorAssegurado:                       req.ValorAssegurado,
		ApoliceSeguroURL:                      req.ApoliceSeguroURL,
		RatingRisco:                           rating,
		DataInicio:                            time.Now(),
		ObservacoesCompliance:                 req.ObservacoesCompliance,
	}

	if len(req.GarantiasBiologicas) > 0 {
		garantias := make(models.JSONB)
		garantiasArray := make([]interface{}, len(req.GarantiasBiologicas))
		for i, g := range req.GarantiasBiologicas {
			garantiasArray[i] = g
		}
		garantias["itens"] = garantiasArray
		tokenizacao.GarantiasBiologicas = garantias
	}

	if err := s.repo.Create(ctx, tokenizacao); err != nil {
		s.logger.LogError(err, "TokenizacaoService.Create", logging.Fields{
			"equino_id": req.EquinoID,
			"user_id":   userID,
		})
		return nil, err
	}

	transacaoEmissao := &models.TransacaoToken{
		TokenizacaoID:  tokenizacao.ID,
		VendedorID:     nil,
		CompradorID:    &userID,
		Quantidade:     tokensBloqueados,
		PrecoUnitario:  req.PrecoInicialToken,
		ValorTotal:     float64(tokensBloqueados) * req.PrecoInicialToken,
		TipoTransacao:  models.TipoTransacaoEmissao,
		HashBlockchain: s.gerarHashBlockchain(tokenizacao.ID, userID, tokensBloqueados),
		Status:         "confirmado",
		DataTransacao:  time.Now(),
	}

	if err := s.repo.CreateTransacao(ctx, transacaoEmissao); err != nil {
		s.logger.LogError(err, "TokenizacaoService.Create.TransacaoEmissao", logging.Fields{
			"tokenizacao_id": tokenizacao.ID,
		})
	}

	participacao := &models.ParticipacaoToken{
		TokenizacaoID:   tokenizacao.ID,
		InvestidorID:    userID,
		QuantidadeTokens: tokensBloqueados,
		PercentualTotal: req.PercentualMinimoDono,
		ValorInvestido:  tokenizacao.ValorTotalTokenizado * (req.PercentualMinimoDono / 100),
		DataAquisicao:   time.Now(),
	}

	if err := s.repo.UpsertParticipacao(ctx, participacao); err != nil {
		s.logger.LogError(err, "TokenizacaoService.Create.Participacao", logging.Fields{
			"tokenizacao_id": tokenizacao.ID,
		})
	}

	s.logger.WithFields(logging.Fields{
		"tokenizacao_id": tokenizacao.ID,
		"equino_id":      tokenizacao.EquinoID,
		"total_tokens":   tokenizacao.TotalTokens,
		"valor_total":    tokenizacao.ValorTotalTokenizado,
	}).Info("Tokenização criada com sucesso")

	return s.GetByID(ctx, tokenizacao.ID)
}

func (s *service) ListTransacoes(ctx context.Context, tokenizacaoID uint) ([]*models.TransacaoTokenResponse, error) {
	transacoes, err := s.repo.FindTransacoesByTokenizacaoID(ctx, tokenizacaoID)
	if err != nil {
		s.logger.LogError(err, "TokenizacaoService.ListTransacoes", logging.Fields{"tokenizacao_id": tokenizacaoID})
		return nil, err
	}

	responses := make([]*models.TransacaoTokenResponse, len(transacoes))
	for i, t := range transacoes {
		responses[i] = t.ToResponse()
	}

	return responses, nil
}

func (s *service) ExecutarOrdem(ctx context.Context, userID uint, req *models.OrdemCompraTokenRequest) (*models.TransacaoTokenResponse, error) {
	tokenizacao, err := s.repo.FindByID(ctx, req.TokenizacaoID)
	if err != nil {
		return nil, err
	}

	if tokenizacao.Status != models.StatusTokenAtivo {
		return nil, &apperrors.ValidationError{Message: "tokenização não está ativa para negociação"}
	}

	if req.QuantidadeDesejada > tokenizacao.TokensDisponiveisVenda {
		return nil, &apperrors.ValidationError{
			Message: fmt.Sprintf("tokens insuficientes (disponível: %d, solicitado: %d)",
				tokenizacao.TokensDisponiveisVenda, req.QuantidadeDesejada),
		}
	}

	precoAtual := tokenizacao.PrecoInicialToken
	if tokenizacao.TokensVendidos > 0 {
		precoAtual = tokenizacao.PrecoInicialToken * 1.1
	}

	if req.PrecoMaximo < precoAtual {
		return nil, &apperrors.ValidationError{
			Message: fmt.Sprintf("preço máximo (%.2f) é menor que o preço atual (%.2f)",
				req.PrecoMaximo, precoAtual),
		}
	}

	transacao := &models.TransacaoToken{
		TokenizacaoID:  req.TokenizacaoID,
		VendedorID:     nil,
		CompradorID:    &userID,
		Quantidade:     req.QuantidadeDesejada,
		PrecoUnitario:  precoAtual,
		TipoTransacao:  models.TipoTransacaoVendaDireta,
		HashBlockchain: s.gerarHashBlockchain(req.TokenizacaoID, userID, req.QuantidadeDesejada),
		DataTransacao:  time.Now(),
	}

	if err := s.repo.CreateTransacao(ctx, transacao); err != nil {
		s.logger.LogError(err, "TokenizacaoService.ExecutarOrdem", logging.Fields{
			"tokenizacao_id": req.TokenizacaoID,
			"comprador_id":   userID,
		})
		return nil, err
	}

	participacao := &models.ParticipacaoToken{
		TokenizacaoID:   req.TokenizacaoID,
		InvestidorID:    userID,
		QuantidadeTokens: req.QuantidadeDesejada,
		PercentualTotal: float64(req.QuantidadeDesejada) / float64(tokenizacao.TotalTokens) * 100,
		ValorInvestido:  transacao.ValorTotal,
		DataAquisicao:   time.Now(),
	}

	if err := s.repo.UpsertParticipacao(ctx, participacao); err != nil {
		s.logger.LogError(err, "TokenizacaoService.ExecutarOrdem.Participacao", logging.Fields{
			"tokenizacao_id": req.TokenizacaoID,
		})
	}

	s.logger.WithFields(logging.Fields{
		"transacao_id":   transacao.ID,
		"tokenizacao_id": req.TokenizacaoID,
		"comprador_id":   userID,
		"quantidade":     req.QuantidadeDesejada,
		"valor_total":    transacao.ValorTotal,
	}).Info("Ordem executada com sucesso")

	return transacao.ToResponse(), nil
}

func (s *service) CriarOferta(ctx context.Context, userID uint, req *models.OfertaTokenRequest) error {
	tokenizacao, err := s.repo.FindByID(ctx, req.TokenizacaoID)
	if err != nil {
		return err
	}

	if tokenizacao.Status != models.StatusTokenAtivo {
		return &apperrors.ValidationError{Message: "tokenização não está ativa"}
	}

	participacoes, err := s.repo.FindParticipacoesByTokenizacaoID(ctx, req.TokenizacaoID)
	if err != nil {
		return err
	}

	var participacaoVendedor *models.ParticipacaoToken
	for _, p := range participacoes {
		if p.InvestidorID == userID {
			participacaoVendedor = p
			break
		}
	}

	if participacaoVendedor == nil {
		return &apperrors.ValidationError{Message: "você não possui tokens desta tokenização"}
	}

	if participacaoVendedor.QuantidadeTokens < req.QuantidadeOfertada {
		return &apperrors.ValidationError{
			Message: fmt.Sprintf("tokens insuficientes (você possui: %d, ofertado: %d)",
				participacaoVendedor.QuantidadeTokens, req.QuantidadeOfertada),
		}
	}

	if tokenizacao.TravaControleDono && participacaoVendedor.PercentualTotal >= tokenizacao.PercentualMinimoDono {
		percentualAposVenda := participacaoVendedor.PercentualTotal - 
			(float64(req.QuantidadeOfertada) / float64(tokenizacao.TotalTokens) * 100)
		
		if percentualAposVenda < tokenizacao.PercentualMinimoDono {
			return &apperrors.ValidationError{
				Message: fmt.Sprintf("venda violaria trava de controle do dono (mínimo: %.2f%%)", 
					tokenizacao.PercentualMinimoDono),
			}
		}
	}

	oferta := &models.OfertaToken{
		TokenizacaoID:     req.TokenizacaoID,
		VendedorID:        userID,
		QuantidadeOfertada: req.QuantidadeOfertada,
		PrecoUnitario:     req.PrecoUnitario,
		Status:            "ativa",
		DataCriacao:       time.Now(),
		DataExpiracao:     time.Now().AddDate(0, 0, req.DiasValidade),
	}

	if err := s.repo.CreateOferta(ctx, oferta); err != nil {
		s.logger.LogError(err, "TokenizacaoService.CriarOferta", logging.Fields{
			"tokenizacao_id": req.TokenizacaoID,
			"vendedor_id":    userID,
		})
		return err
	}

	s.logger.WithFields(logging.Fields{
		"oferta_id":      oferta.ID,
		"tokenizacao_id": req.TokenizacaoID,
		"vendedor_id":    userID,
		"quantidade":     req.QuantidadeOfertada,
		"preco":          req.PrecoUnitario,
	}).Info("Oferta de venda criada com sucesso")

	return nil
}

func (s *service) CalcularRatingRisco(ctx context.Context, equinoID uint) (models.RatingRisco, error) {
	equino, err := s.equinoRepo.FindByID(ctx, equinoID)
	if err != nil {
		return models.RatingA, err
	}

	pontuacao := 0

	if equino.Status == "ativo" {
		pontuacao += 20
	}

	if equino.DataNascimento != nil {
		idade := time.Since(*equino.DataNascimento).Hours() / 24 / 365
		if idade >= 3 && idade <= 12 {
			pontuacao += 20
		} else if idade < 3 || idade > 15 {
			pontuacao += 10
		}
	}

	if equino.Sexo == "macho" {
		pontuacao += 10
	} else if equino.Sexo == "femea" {
		pontuacao += 15
	}

	if pontuacao >= 90 {
		return models.RatingAAAPlus, nil
	} else if pontuacao >= 85 {
		return models.RatingAAA, nil
	} else if pontuacao >= 80 {
		return models.RatingAAPlus, nil
	} else if pontuacao >= 75 {
		return models.RatingAA, nil
	} else if pontuacao >= 70 {
		return models.RatingAPlus, nil
	} else if pontuacao >= 60 {
		return models.RatingA, nil
	} else if pontuacao >= 50 {
		return models.RatingBBB, nil
	} else if pontuacao >= 40 {
		return models.RatingBB, nil
	} else if pontuacao >= 30 {
		return models.RatingB, nil
	}
	return models.RatingC, nil
}

func (s *service) gerarHashBlockchain(tokenizacaoID uint, userID uint, quantidade int) string {
	data := fmt.Sprintf("%d-%d-%d-%d", tokenizacaoID, userID, quantidade, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return "0x" + hex.EncodeToString(hash[:])
}
