package services

import (
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/pkg/cache"
	"github.com/equinoid/backend/pkg/logging"
	"gorm.io/gorm"
)

// ExameLaboratorialService gerencia operações relacionadas a exames laboratoriais
type ExameLaboratorialService struct {
	db     *gorm.DB
	cache  cache.CacheInterface
	logger *logging.Logger
}

// NewExameLaboratorialService cria uma nova instância do serviço de exames
func NewExameLaboratorialService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger) *ExameLaboratorialService {
	return &ExameLaboratorialService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// List retorna lista de exames, opcionalmente filtrada por laboratório ou solicitante
func (s *ExameLaboratorialService) List(laboratorioID *uint, solicitanteID *uint) ([]models.ExameLaboratorial, error) {
	var exames []models.ExameLaboratorial
	query := s.db.Preload("Equino").Preload("Solicitante").Preload("Laboratorio")

	if laboratorioID != nil {
		query = query.Where("laboratorio_id = ?", *laboratorioID)
	}

	if solicitanteID != nil {
		query = query.Where("solicitante_id = ?", *solicitanteID)
	}

	if err := query.Order("data_solicitacao DESC").Find(&exames).Error; err != nil {
		s.logger.Errorf("Erro ao listar exames: %v", err)
		return nil, err
	}

	return exames, nil
}

// Get busca um exame por ID
func (s *ExameLaboratorialService) Get(id uint) (*models.ExameLaboratorial, error) {
	var exame models.ExameLaboratorial
	if err := s.db.Preload("Equino").Preload("Solicitante").Preload("Laboratorio").
		First(&exame, id).Error; err != nil {
		s.logger.Errorf("Erro ao buscar exame %d: %v", id, err)
		return nil, err
	}
	return &exame, nil
}

// Create cria um novo exame
func (s *ExameLaboratorialService) Create(exame *models.ExameLaboratorial) error {
	if err := s.db.Create(exame).Error; err != nil {
		s.logger.Errorf("Erro ao criar exame: %v", err)
		return err
	}
	s.logger.Infof("Exame criado: %s para equino %s (ID: %d)", exame.TipoExame, exame.Equinoid, exame.ID)
	return nil
}

// Update atualiza um exame
func (s *ExameLaboratorialService) Update(id uint, updates map[string]interface{}) error {
	if err := s.db.Model(&models.ExameLaboratorial{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		s.logger.Errorf("Erro ao atualizar exame %d: %v", id, err)
		return err
	}
	s.logger.Infof("Exame %d atualizado", id)
	return nil
}

// AdicionarResultado adiciona resultado a um exame
func (s *ExameLaboratorialService) AdicionarResultado(id uint, resultado, observacoes, documentoURL string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"resultado":      resultado,
		"observacoes":    observacoes,
		"documento_url":  documentoURL,
		"status":         "concluido",
		"data_resultado": &now,
	}

	if err := s.Update(id, updates); err != nil {
		return err
	}

	s.logger.Infof("Resultado adicionado ao exame %d", id)
	return nil
}

// AtribuirLaboratorio atribui um laboratório a um exame
func (s *ExameLaboratorialService) AtribuirLaboratorio(id uint, laboratorioID uint) error {
	updates := map[string]interface{}{
		"laboratorio_id": laboratorioID,
		"status":         "em_analise",
	}

	if err := s.Update(id, updates); err != nil {
		return err
	}

	s.logger.Infof("Laboratório %d atribuído ao exame %d", laboratorioID, id)
	return nil
}

// RegistrarColeta registra a coleta de material para o exame
func (s *ExameLaboratorialService) RegistrarColeta(id uint) error {
	now := time.Now()
	updates := map[string]interface{}{
		"data_coleta": &now,
		"status":      "coletado",
	}

	if err := s.Update(id, updates); err != nil {
		return err
	}

	s.logger.Infof("Coleta registrada para exame %d", id)
	return nil
}

// Cancelar cancela um exame
func (s *ExameLaboratorialService) Cancelar(id uint, motivo string) error {
	updates := map[string]interface{}{
		"status":      "cancelado",
		"observacoes": motivo,
	}

	if err := s.Update(id, updates); err != nil {
		return err
	}

	s.logger.Infof("Exame %d cancelado: %s", id, motivo)
	return nil
}
