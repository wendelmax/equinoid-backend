package services

import (
	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/pkg/cache"
	"github.com/equinoid/backend/pkg/logging"
	"gorm.io/gorm"
)

// LeilaoService gerencia operações relacionadas a leilões
type LeilaoService struct {
	db     *gorm.DB
	cache  cache.CacheInterface
	logger *logging.Logger
}

// NewLeilaoService cria uma nova instância do serviço de leilões
func NewLeilaoService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger) *LeilaoService {
	return &LeilaoService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// List retorna lista de leilões, opcionalmente filtrada por leiloeiro
func (s *LeilaoService) List(leiloeiroID *uint) ([]models.Leilao, error) {
	var leiloes []models.Leilao
	query := s.db.Preload("Leiloeiro")

	if leiloeiroID != nil {
		query = query.Where("leiloeiro_id = ?", *leiloeiroID)
	}

	if err := query.Order("data_inicio DESC").Find(&leiloes).Error; err != nil {
		s.logger.Errorf("Erro ao listar leilões: %v", err)
		return nil, err
	}

	return leiloes, nil
}

// Get busca um leilão por ID
func (s *LeilaoService) Get(id uint) (*models.Leilao, error) {
	var leilao models.Leilao
	if err := s.db.Preload("Leiloeiro").First(&leilao, id).Error; err != nil {
		s.logger.Errorf("Erro ao buscar leilão %d: %v", id, err)
		return nil, err
	}
	return &leilao, nil
}

// Create cria um novo leilão
func (s *LeilaoService) Create(leilao *models.Leilao) error {
	if err := s.db.Create(leilao).Error; err != nil {
		s.logger.Errorf("Erro ao criar leilão: %v", err)
		return err
	}
	s.logger.Infof("Leilão criado: %s (ID: %d)", leilao.Nome, leilao.ID)
	return nil
}

// Finalizar encerra um leilão e calcula totais
func (s *LeilaoService) Finalizar(id uint) (*models.Leilao, error) {
	var leilao models.Leilao
	if err := s.db.First(&leilao, id).Error; err != nil {
		s.logger.Errorf("Erro ao buscar leilão %d: %v", id, err)
		return nil, err
	}

	// Calcular totais das participações vendidas
	var participacoes []models.ParticipacaoLeilao
	s.db.Where("leilao_id = ? AND status = ?", id, "vendido").Find(&participacoes)

	var totalArrecadado, totalComissoes float64
	for _, p := range participacoes {
		if p.ValorVendido != nil {
			totalArrecadado += *p.ValorVendido
		}
		if p.ComissaoLeiloeiro != nil {
			totalComissoes += *p.ComissaoLeiloeiro
		}
	}

	// Atualizar leilão
	leilao.Status = "encerrado"
	leilao.TotalArrecadado = &totalArrecadado
	leilao.TotalComissoes = &totalComissoes

	if err := s.db.Save(&leilao).Error; err != nil {
		s.logger.Errorf("Erro ao finalizar leilão %d: %v", id, err)
		return nil, err
	}

	s.logger.Infof("Leilão %d finalizado. Total arrecadado: R$ %.2f, Comissões: R$ %.2f",
		id, totalArrecadado, totalComissoes)

	return &leilao, nil
}

// ListParticipacoes retorna participações de um leilão
func (s *LeilaoService) ListParticipacoes(leilaoID uint) ([]models.ParticipacaoLeilao, error) {
	var participacoes []models.ParticipacaoLeilao
	if err := s.db.Where("leilao_id = ?", leilaoID).
		Preload("Equino").
		Preload("Criador").
		Preload("Comprador").
		Order("created_at DESC").
		Find(&participacoes).Error; err != nil {
		s.logger.Errorf("Erro ao listar participações do leilão %d: %v", leilaoID, err)
		return nil, err
	}
	return participacoes, nil
}

// CreateParticipacao cria uma nova participação em leilão
func (s *LeilaoService) CreateParticipacao(participacao *models.ParticipacaoLeilao) error {
	if err := s.db.Create(participacao).Error; err != nil {
		s.logger.Errorf("Erro ao criar participação: %v", err)
		return err
	}
	s.logger.Infof("Participação criada: Equino ID %d no leilão %d", participacao.EquinoID, participacao.LeilaoID)
	return nil
}

// AprovarParticipacao aprova uma participação
func (s *LeilaoService) AprovarParticipacao(id uint) error {
	if err := s.db.Model(&models.ParticipacaoLeilao{}).
		Where("id = ?", id).
		Update("status", "aprovado").Error; err != nil {
		s.logger.Errorf("Erro ao aprovar participação %d: %v", id, err)
		return err
	}
	s.logger.Infof("Participação %d aprovada", id)
	return nil
}

// RegistrarVenda registra a venda de um equino no leilão
func (s *LeilaoService) RegistrarVenda(id uint, valorVendido float64, compradorID uint, leilao *models.Leilao) error {
	// Calcular comissão
	comissaoPercentual := (valorVendido * leilao.TaxaComissaoPercentual) / 100
	taxaFixa := 0.0
	if leilao.TaxaFixa != nil {
		taxaFixa = *leilao.TaxaFixa
	}
	comissaoTotal := comissaoPercentual + taxaFixa

	// Atualizar participação
	if err := s.db.Model(&models.ParticipacaoLeilao{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"valor_vendido":      valorVendido,
			"valor_final":        valorVendido,
			"comprador_id":       compradorID,
			"status":             "vendido",
			"comissao_leiloeiro": comissaoTotal,
		}).Error; err != nil {
		s.logger.Errorf("Erro ao registrar venda da participação %d: %v", id, err)
		return err
	}

	s.logger.Infof("Venda registrada: Participação %d vendida por R$ %.2f (comissão: R$ %.2f)",
		id, valorVendido, comissaoTotal)

	return nil
}

// MarcarAusencia marca ausência de um equino em leilão presencial
func (s *LeilaoService) MarcarAusencia(id uint) error {
	compareceu := false
	penalizacao := -50.0

	if err := s.db.Model(&models.ParticipacaoLeilao{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"compareceu":           &compareceu,
			"penalizacao_ausencia": &penalizacao,
		}).Error; err != nil {
		s.logger.Errorf("Erro ao marcar ausência da participação %d: %v", id, err)
		return err
	}

	s.logger.Infof("Ausência marcada para participação %d (penalização: %.0f pontos)", id, penalizacao)
	return nil
}

// MarcarPresenca marca presença de um equino em leilão presencial
func (s *LeilaoService) MarcarPresenca(id uint) error {
	compareceu := true

	if err := s.db.Model(&models.ParticipacaoLeilao{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"compareceu":           &compareceu,
			"penalizacao_ausencia": nil,
		}).Error; err != nil {
		s.logger.Errorf("Erro ao marcar presença da participação %d: %v", id, err)
		return err
	}

	s.logger.Infof("Presença confirmada para participação %d", id)
	return nil
}

// GetRelatorioGanhos retorna relatório de ganhos de leilões finalizados
func (s *LeilaoService) GetRelatorioGanhos(leiloeiroID *uint) ([]map[string]interface{}, error) {
	var leiloes []models.Leilao
	query := s.db.Where("status = ?", "encerrado")

	if leiloeiroID != nil {
		query = query.Where("leiloeiro_id = ?", *leiloeiroID)
	}

	if err := query.Order("data_fim DESC").Find(&leiloes).Error; err != nil {
		s.logger.Errorf("Erro ao buscar leilões finalizados: %v", err)
		return nil, err
	}

	var relatorios []map[string]interface{}
	for _, leilao := range leiloes {
		// Buscar participações vendidas
		var participacoes []models.ParticipacaoLeilao
		s.db.Where("leilao_id = ? AND status = ?", leilao.ID, "vendido").
			Preload("Equino").
			Preload("Criador").
			Preload("Comprador").
			Find(&participacoes)

		var totalAnimais int64
		s.db.Model(&models.ParticipacaoLeilao{}).Where("leilao_id = ?", leilao.ID).Count(&totalAnimais)

		totalTaxasFixas := 0.0
		if leilao.TaxaFixa != nil {
			totalTaxasFixas = *leilao.TaxaFixa * float64(len(participacoes))
		}

		relatorio := map[string]interface{}{
			"leilao_id":         leilao.ID,
			"leilao_nome":       leilao.Nome,
			"data_fim":          leilao.DataFim,
			"total_animais":     totalAnimais,
			"animais_vendidos":  len(participacoes),
			"total_arrecadado":  leilao.TotalArrecadado,
			"total_comissoes":   leilao.TotalComissoes,
			"total_taxas_fixas": totalTaxasFixas,
			"ganho_total":       leilao.TotalComissoes,
			"participacoes":     participacoes,
		}

		relatorios = append(relatorios, relatorio)
	}

	return relatorios, nil
}
