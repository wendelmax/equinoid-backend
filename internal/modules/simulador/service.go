package simulador

import (
	"context"
	"math"

	"github.com/equinoid/backend/internal/modules/equinos"
	"github.com/equinoid/backend/pkg/cache"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
)

type SimulacaoResult struct {
	Inbreeding           float64 `json:"inbreeding"`
	AptidaoEsportiva     int     `json:"aptidao_esportiva"`
	ValorizacaoEstimada  string  `json:"valorizacao_estimada"`
	Rating               string  `json:"rating"`
	Mensagem             string  `json:"mensagem"`
}

type Service interface {
	SimularCruzamento(ctx context.Context, paiEquinoid, maeEquinoid string) (*SimulacaoResult, error)
}

type service struct {
	equinoRepo equinos.Repository
	cache      cache.CacheInterface
	logger     *logging.Logger
}

func NewService(equinoRepo equinos.Repository, cache cache.CacheInterface, logger *logging.Logger) Service {
	return &service{
		equinoRepo: equinoRepo,
		cache:      cache,
		logger:     logger,
	}
}

func (s *service) SimularCruzamento(ctx context.Context, paiEquinoid, maeEquinoid string) (*SimulacaoResult, error) {
	pai, err := s.equinoRepo.FindByEquinoid(ctx, paiEquinoid)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "pai não encontrado", ID: paiEquinoid}
		}
		s.logger.LogError(err, "SimuladorService.SimularCruzamento", logging.Fields{"pai": paiEquinoid})
		return nil, err
	}

	mae, err := s.equinoRepo.FindByEquinoid(ctx, maeEquinoid)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "mãe não encontrada", ID: maeEquinoid}
		}
		s.logger.LogError(err, "SimuladorService.SimularCruzamento", logging.Fields{"mae": maeEquinoid})
		return nil, err
	}

	inbreeding := s.calcularConsanguinidade(ctx, pai.ID, mae.ID)
	
	aptidao := s.calcularAptidaoEsportiva(pai, mae, inbreeding)
	
	valorizacao, rating := s.calcularValorizacao(aptidao, inbreeding)
	
	mensagem := s.gerarMensagem(aptidao, inbreeding, rating)

	result := &SimulacaoResult{
		Inbreeding:          math.Round(inbreeding*100) / 100,
		AptidaoEsportiva:    aptidao,
		ValorizacaoEstimada: valorizacao,
		Rating:              rating,
		Mensagem:            mensagem,
	}

	s.logger.WithFields(logging.Fields{
		"pai":        paiEquinoid,
		"mae":        maeEquinoid,
		"inbreeding": result.Inbreeding,
		"aptidao":    result.AptidaoEsportiva,
	}).Info("Simulação de cruzamento realizada")

	return result, nil
}

func (s *service) calcularConsanguinidade(ctx context.Context, paiID, maeID uint) float64 {
	ancestraisComuns := s.buscarAncestaisComuns(ctx, paiID, maeID)
	
	if len(ancestraisComuns) == 0 {
		return 0.0
	}

	coeficiente := 0.0
	for _, ancestral := range ancestraisComuns {
		geracoesPai := ancestral.GeracoesPai
		geracoesMae := ancestral.GeracoesMae
		
		contribuicao := math.Pow(0.5, float64(geracoesPai+geracoesMae+1))
		coeficiente += contribuicao
	}

	return coeficiente * 100
}

type ancestralComum struct {
	ID          uint
	GeracoesPai int
	GeracoesMae int
}

func (s *service) buscarAncestaisComuns(ctx context.Context, paiID, maeID uint) []ancestralComum {
	ancestraisPai := s.buscarAncestaisRecursivo(ctx, paiID, 0, make(map[uint]int))
	ancestraisMae := s.buscarAncestaisRecursivo(ctx, maeID, 0, make(map[uint]int))

	var comuns []ancestralComum
	for id, geracoesPai := range ancestraisPai {
		if geracoesMae, existe := ancestraisMae[id]; existe {
			comuns = append(comuns, ancestralComum{
				ID:          id,
				GeracoesPai: geracoesPai,
				GeracoesMae: geracoesMae,
			})
		}
	}

	return comuns
}

func (s *service) buscarAncestaisRecursivo(ctx context.Context, equinoID uint, geracao int, visitados map[uint]int) map[uint]int {
	if geracao > 5 {
		return visitados
	}

	_, err := s.equinoRepo.FindByID(ctx, equinoID)
	if err != nil {
		return visitados
	}

	if existingGeracao, existe := visitados[equinoID]; !existe || geracao < existingGeracao {
		visitados[equinoID] = geracao
	}

	return visitados
}

func (s *service) calcularAptidaoEsportiva(pai, mae interface{}, inbreeding float64) int {
	aptidaoBase := 75.0
	
	variacao := (math.Sin(float64(pai.(interface{ GetID() uint }).GetID())*0.1) + 
		         math.Cos(float64(mae.(interface{ GetID() uint }).GetID())*0.1)) * 10
	
	penalizacaoConsanguinidade := inbreeding * 2
	
	aptidaoFinal := aptidaoBase + variacao - penalizacaoConsanguinidade
	
	if aptidaoFinal < 50 {
		aptidaoFinal = 50
	}
	if aptidaoFinal > 100 {
		aptidaoFinal = 100
	}

	return int(math.Round(aptidaoFinal))
}

func (s *service) calcularValorizacao(aptidao int, inbreeding float64) (string, string) {
	rating := "B"
	valorizacao := "Média"

	if inbreeding > 5 {
		rating = "C"
		valorizacao = "Baixa"
	} else if aptidao >= 95 {
		rating = "AAA+"
		valorizacao = "Excepcional"
	} else if aptidao >= 90 {
		rating = "AAA"
		valorizacao = "Muito Alta"
	} else if aptidao >= 85 {
		rating = "AA+"
		valorizacao = "Alta"
	} else if aptidao >= 80 {
		rating = "AA"
		valorizacao = "Alta"
	} else if aptidao >= 75 {
		rating = "A"
		valorizacao = "Média-Alta"
	} else if aptidao >= 70 {
		rating = "BBB"
		valorizacao = "Média"
	} else if aptidao >= 60 {
		rating = "BB"
		valorizacao = "Média-Baixa"
	} else {
		rating = "C"
		valorizacao = "Baixa"
	}

	return valorizacao, rating
}

func (s *service) gerarMensagem(aptidao int, inbreeding float64, rating string) string {
	if inbreeding > 5 {
		return "⚠️ ATENÇÃO: Nível de consanguinidade ALTO detectado. Cruzamento NÃO recomendado. Alto risco de problemas genéticos e características recessivas."
	}

	if inbreeding > 3 {
		return "⚠️ Nível de consanguinidade MODERADO. Recomenda-se avaliação veterinária especializada antes do cruzamento. Monitorar características recessivas."
	}

	if aptidao >= 95 {
		return "🏆 EXCELENTE! Combinação genética excepcional com altíssimo potencial esportivo. Cruzamento altamente recomendado para produção de elite."
	}

	if aptidao >= 90 {
		return "⭐ MUITO BOM! Excelente combinação genética com forte potencial para alta performance esportiva e valorização significativa."
	}

	if aptidao >= 85 {
		return "✅ BOM cruzamento. Combinação equilibrada com bom potencial esportivo e valorização esperada acima da média."
	}

	if aptidao >= 75 {
		return "✅ Cruzamento equilibrado com potencial esportivo satisfatório e baixo risco de consanguinidade."
	}

	if aptidao >= 65 {
		return "ℹ️ Cruzamento viável, porém com potencial esportivo moderado. Considerar outras opções para maior valorização."
	}

	return "ℹ️ Cruzamento com potencial limitado. Recomenda-se avaliar outras opções de reprodutores para melhor resultado genético."
}
