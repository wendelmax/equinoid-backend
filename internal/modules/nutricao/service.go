package nutricao

import (
	"context"
	"fmt"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/internal/modules/equinos"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
)

type Service interface {
	GetPlanoByEquinoid(ctx context.Context, equinoid string) (*models.PlanoNutricionalResponse, error)
	CreatePlano(ctx context.Context, userID uint, req *models.CreatePlanoNutricionalRequest) (*models.PlanoNutricionalResponse, error)
	GetSugestaoIA(ctx context.Context, req *models.SugestaoIARequest) (*models.SugestaoIAResponse, error)
	CreateRefeicao(ctx context.Context, userID uint, req *models.CreateRefeicaoRequest) (*models.Refeicao, error)
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

func (s *service) GetPlanoByEquinoid(ctx context.Context, equinoid string) (*models.PlanoNutricionalResponse, error) {
	plano, err := s.repo.FindByEquinoid(ctx, equinoid)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "NutricaoService.GetPlanoByEquinoid", logging.Fields{"equinoid": equinoid})
		}
		return nil, err
	}
	return plano.ToResponse(), nil
}

func (s *service) CreatePlano(ctx context.Context, userID uint, req *models.CreatePlanoNutricionalRequest) (*models.PlanoNutricionalResponse, error) {
	equino, err := s.equinoRepo.FindByEquinoid(ctx, req.Equinoid)
	if err != nil {
		return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado"}
	}

	plano := &models.PlanoNutricional{
		EquinoID:            equino.ID,
		TipoPlano:           req.TipoPlano,
		Objetivo:            req.Objetivo,
		PesoAtual:           req.PesoAtual,
		PesoIdeal:           req.PesoIdeal,
		FrequenciaRefeicoes: req.FrequenciaRefeicoes,
		ObservacoesGerais:   req.ObservacoesGerais,
		GeradoPorIA:         req.GerarComIA,
		Status:              models.StatusPlanoAtivo,
		DataInicio:          time.Now(),
	}

	if req.GerarComIA {
		sugestaoIA := s.gerarPlanoComIA(req)
		plano.CaloriasObjetivo = sugestaoIA.PlanoSugerido.CaloriasObjetivo
		plano.ProteinasGramas = sugestaoIA.PlanoSugerido.Macronutrientes.Proteinas
		plano.CarboidratosGramas = sugestaoIA.PlanoSugerido.Macronutrientes.Carboidratos
		plano.GordurasGramas = sugestaoIA.PlanoSugerido.Macronutrientes.Gorduras
		plano.FibrasGramas = sugestaoIA.PlanoSugerido.Macronutrientes.Fibras
		plano.PromptIA = fmt.Sprintf("Tipo: %s, Objetivo: %s, Peso: %.2f kg", req.TipoPlano, req.Objetivo, req.PesoAtual)
		plano.RespostaIA = sugestaoIA.Sugestao
	} else {
		plano.CaloriasObjetivo = s.calcularCaloriasBase(req.PesoAtual, req.TipoPlano)
		plano.ProteinasGramas = float64(plano.CaloriasObjetivo) * 0.25 / 4
		plano.CarboidratosGramas = float64(plano.CaloriasObjetivo) * 0.45 / 4
		plano.GordurasGramas = float64(plano.CaloriasObjetivo) * 0.30 / 9
		plano.FibrasGramas = req.PesoAtual * 1.5
	}

	if len(req.Suplementos) > 0 {
		suplementos := make(models.JSONB)
		itens := make([]interface{}, len(req.Suplementos))
		for i, s := range req.Suplementos {
			itens[i] = s
		}
		suplementos["itens"] = itens
		plano.Suplementos = suplementos
	}

	if len(req.Restricoes) > 0 {
		restricoes := make(models.JSONB)
		itens := make([]interface{}, len(req.Restricoes))
		for i, r := range req.Restricoes {
			itens[i] = r
		}
		restricoes["itens"] = itens
		plano.Restricoes = restricoes
	}

	if err := s.repo.Create(ctx, plano); err != nil {
		s.logger.LogError(err, "NutricaoService.CreatePlano", logging.Fields{"equinoid": req.Equinoid})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{
		"plano_id":  plano.ID,
		"equinoid":  req.Equinoid,
		"gerado_ia": req.GerarComIA,
	}).Info("Plano nutricional criado")

	return s.GetPlanoByEquinoid(ctx, req.Equinoid)
}

func (s *service) GetSugestaoIA(ctx context.Context, req *models.SugestaoIARequest) (*models.SugestaoIAResponse, error) {
	_, err := s.equinoRepo.FindByEquinoid(ctx, req.Equinoid)
	if err != nil {
		return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado"}
	}

	sugestao := s.gerarPlanoComIA(&models.CreatePlanoNutricionalRequest{
		Equinoid:  req.Equinoid,
		TipoPlano: req.TipoPlano,
		Objetivo:  req.Objetivo,
		PesoAtual: req.PesoAtual,
	})

	return sugestao, nil
}

func (s *service) CreateRefeicao(ctx context.Context, userID uint, req *models.CreateRefeicaoRequest) (*models.Refeicao, error) {
	equino, err := s.equinoRepo.FindByEquinoid(ctx, req.Equinoid)
	if err != nil {
		return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado"}
	}

	plano, _ := s.repo.FindByEquinoid(ctx, req.Equinoid)
	var planoID uint
	if plano != nil {
		planoID = plano.ID
	}

	dataRefeicao := time.Now()
	if req.DataRefeicao != nil {
		dataRefeicao = *req.DataRefeicao
	}

	alimentosJSON := make(models.JSONB)
	itens := make([]interface{}, len(req.Alimentos))
	totalCalorias := 0
	quantidadeTotal := 0.0

	for i, a := range req.Alimentos {
		itens[i] = map[string]interface{}{
			"nome":       a.Nome,
			"quantidade": a.Quantidade,
			"unidade":    a.Unidade,
			"calorias":   a.Calorias,
		}
		totalCalorias += a.Calorias
		quantidadeTotal += a.Quantidade
	}
	alimentosJSON["itens"] = itens

	refeicao := &models.Refeicao{
		PlanoNutricionalID: planoID,
		EquinoID:           equino.ID,
		DataRefeicao:       dataRefeicao,
		TipoRefeicao:       req.TipoRefeicao,
		Alimentos:          alimentosJSON,
		QuantidadeTotal:    quantidadeTotal,
		CaloriasConsumidas: totalCalorias,
		Observacoes:        req.Observacoes,
		RegistradoPor:      userID,
	}

	if err := s.repo.CreateRefeicao(ctx, refeicao); err != nil {
		s.logger.LogError(err, "NutricaoService.CreateRefeicao", logging.Fields{"equinoid": req.Equinoid})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{
		"refeicao_id": refeicao.ID,
		"equinoid":    req.Equinoid,
		"calorias":    totalCalorias,
	}).Info("Refeição registrada")

	return refeicao, nil
}

func (s *service) calcularCaloriasBase(peso float64, tipo models.TipoPlano) int {
	caloriasPorKg := 25.0
	
	switch tipo {
	case models.TipoPlanoGanhoMassa:
		caloriasPorKg = 35.0
	case models.TipoPlanoPerdaPeso:
		caloriasPorKg = 20.0
	case models.TipoPlanoAltoRendimento:
		caloriasPorKg = 40.0
	case models.TipoPlanoGestacao:
		caloriasPorKg = 32.0
	case models.TipoPlanoLactacao:
		caloriasPorKg = 38.0
	}

	return int(peso * caloriasPorKg)
}

func (s *service) gerarPlanoComIA(req *models.CreatePlanoNutricionalRequest) *models.SugestaoIAResponse {
	calorias := s.calcularCaloriasBase(req.PesoAtual, req.TipoPlano)

	sugestao := fmt.Sprintf(
		"Plano Nutricional %s para equino de %.2f kg com objetivo de %s. "+
			"Recomendação: %d kcal/dia com foco em alimentos naturais de alta qualidade.",
		req.TipoPlano, req.PesoAtual, req.Objetivo, calorias,
	)

	recomendacoes := []string{
		"Fornecer água fresca e limpa à vontade",
		"Dividir alimentação em múltiplas refeições ao dia",
		"Incluir forragem de qualidade (feno, capim)",
		"Monitorar peso semanalmente",
		"Ajustar quantidades conforme resposta do animal",
	}

	if req.TipoPlano == models.TipoPlanoGanhoMassa {
		recomendacoes = append(recomendacoes, "Adicionar ração concentrada rica em proteínas", "Suplementar com vitaminas e minerais")
	}

	return &models.SugestaoIAResponse{
		Sugestao: sugestao,
		PlanoSugerido: models.PlanoNutricionalIA{
			CaloriasObjetivo: calorias,
			Macronutrientes: models.MacronutrientesIA{
				Proteinas:    float64(calorias) * 0.25 / 4,
				Carboidratos: float64(calorias) * 0.45 / 4,
				Gorduras:     float64(calorias) * 0.30 / 9,
				Fibras:       req.PesoAtual * 1.5,
			},
			RefeicoesDetalhadas: []models.RefeicaoIA{
				{
					Horario: "07:00",
					Alimentos: []models.AlimentoRefeicao{
						{Nome: "Feno de alfafa", Quantidade: 3.0, Unidade: "kg", Calorias: 600},
						{Nome: "Ração concentrada", Quantidade: 1.5, Unidade: "kg", Calorias: 450},
					},
					Calorias: 1050,
				},
				{
					Horario: "12:00",
					Alimentos: []models.AlimentoRefeicao{
						{Nome: "Capim fresco", Quantidade: 5.0, Unidade: "kg", Calorias: 500},
					},
					Calorias: 500,
				},
				{
					Horario: "18:00",
					Alimentos: []models.AlimentoRefeicao{
						{Nome: "Feno", Quantidade: 4.0, Unidade: "kg", Calorias: 700},
						{Nome: "Aveia", Quantidade: 1.0, Unidade: "kg", Calorias: 380},
					},
					Calorias: 1080,
				},
			},
			Suplementos: []string{"Sal mineral", "Vitamina E", "Selênio"},
		},
		RecomendacoesExtras: recomendacoes,
		GeradoEm:            time.Now(),
	}
}
