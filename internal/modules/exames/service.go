package exames

import (
	"context"
	"time"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
)

type Service interface {
	ListAll(ctx context.Context, filters map[string]interface{}) ([]*models.ExameLaboratorial, error)
	GetByID(ctx context.Context, id uint) (*models.ExameLaboratorial, error)
	Create(ctx context.Context, exame *models.CreateExameRequest) (*models.ExameLaboratorial, error)
	Update(ctx context.Context, id uint, req *models.UpdateExameRequest) (*models.ExameLaboratorial, error)
	Delete(ctx context.Context, id uint) error
	
	ReceberAmostra(ctx context.Context, id uint, dataRecebimento *string) (*models.ExameLaboratorial, error)
	IniciarAnalise(ctx context.Context, id uint) (*models.ExameLaboratorial, error)
	ConcluirExame(ctx context.Context, id uint, resultado models.ResultadoExame, valores map[string]interface{}, laudo *string) (*models.ExameLaboratorial, error)
}

type service struct {
	repo   Repository
	logger *logging.Logger
}

func NewService(repo Repository, logger *logging.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) ListAll(ctx context.Context, filters map[string]interface{}) ([]*models.ExameLaboratorial, error) {
	exames, err := s.repo.FindAll(ctx, filters)
	if err != nil {
		s.logger.LogError(err, "ExameService.ListAll", logging.Fields{"filters": filters})
		return nil, err
	}
	return exames, nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*models.ExameLaboratorial, error) {
	exame, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "ExameService.GetByID", logging.Fields{"id": id})
		}
		return nil, err
	}
	return exame, nil
}

func (s *service) Create(ctx context.Context, req *models.CreateExameRequest) (*models.ExameLaboratorial, error) {
	exame := &models.ExameLaboratorial{
		Equinoid:                  req.Equinoid,
		TipoExame:                 req.TipoExame,
		NomeExame:                 req.NomeExame,
		Descricao:                 req.Descricao,
		VeterinarioSolicitanteID:  req.VeterinarioSolicitanteID,
		LaboratorioID:             req.LaboratorioID,
		Status:                    "solicitado",
		DataSolicitacao:           time.Now(),
		Observacoes:               req.Observacoes,
	}

	if err := s.repo.Create(ctx, exame); err != nil {
		s.logger.LogError(err, "ExameService.Create", logging.Fields{"equinoid": req.Equinoid})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{"exame_id": exame.ID, "tipo": exame.TipoExame}).Info("Exame solicitado")
	return s.GetByID(ctx, exame.ID)
}

func (s *service) Update(ctx context.Context, id uint, req *models.UpdateExameRequest) (*models.ExameLaboratorial, error) {
	exame, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Status != nil {
		exame.Status = *req.Status
	}
	if req.DataColeta != nil {
		exame.DataColeta = req.DataColeta
	}
	if req.DataRecebimentoAmostra != nil {
		exame.DataRecebimentoAmostra = req.DataRecebimentoAmostra
	}
	if req.DataInicioAnalise != nil {
		exame.DataInicioAnalise = req.DataInicioAnalise
	}
	if req.DataConclusao != nil {
		exame.DataConclusao = req.DataConclusao
	}
	if req.Resultado != nil {
		exame.Resultado = req.Resultado
	}
	if req.Valores != nil {
		exame.Valores = req.Valores
	}
	if req.Laudo != nil {
		exame.Laudo = req.Laudo
	}
	if req.Observacoes != nil {
		exame.Observacoes = req.Observacoes
	}

	if err := s.repo.Update(ctx, exame); err != nil {
		s.logger.LogError(err, "ExameService.Update", logging.Fields{"id": id})
		return nil, err
	}

	return s.GetByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "ExameService.Delete", logging.Fields{"id": id})
		}
		return err
	}
	s.logger.WithFields(logging.Fields{"exame_id": id}).Info("Exame deletado")
	return nil
}

func (s *service) ReceberAmostra(ctx context.Context, id uint, dataRecebimento *string) (*models.ExameLaboratorial, error) {
	exame, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if exame.Status != "solicitado" {
		return nil, &apperrors.ValidationError{Message: "apenas exames solicitados podem receber amostra"}
	}

	status := "amostra_recebida"
	var dataRecebimentoTime *time.Time
	if dataRecebimento != nil {
		t, err := time.Parse("2006-01-02", *dataRecebimento)
		if err == nil {
			dataRecebimentoTime = &t
		}
	}
	if dataRecebimentoTime == nil {
		now := time.Now()
		dataRecebimentoTime = &now
	}

	updateReq := &models.UpdateExameRequest{
		Status:                  &status,
		DataRecebimentoAmostra: dataRecebimentoTime,
	}

	return s.Update(ctx, id, updateReq)
}

func (s *service) IniciarAnalise(ctx context.Context, id uint) (*models.ExameLaboratorial, error) {
	exame, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if exame.Status != "amostra_recebida" {
		return nil, &apperrors.ValidationError{Message: "apenas exames com amostra recebida podem iniciar análise"}
	}

	status := "em_analise"
	now := time.Now()
	updateReq := &models.UpdateExameRequest{
		Status:             &status,
		DataInicioAnalise: &now,
	}

	return s.Update(ctx, id, updateReq)
}

func (s *service) ConcluirExame(ctx context.Context, id uint, resultado models.ResultadoExame, valores map[string]interface{}, laudo *string) (*models.ExameLaboratorial, error) {
	exame, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if exame.Status != "em_analise" {
		return nil, &apperrors.ValidationError{Message: "apenas exames em análise podem ser concluídos"}
	}

	status := "concluido"
	now := time.Now()
	updateReq := &models.UpdateExameRequest{
		Status:         &status,
		DataConclusao:  &now,
		Resultado:      &resultado,
		Valores:        valores,
		Laudo:          laudo,
	}

	exameAtualizado, err := s.Update(ctx, id, updateReq)
	if err != nil {
		return nil, err
	}

	s.logger.WithFields(logging.Fields{
		"exame_id":  id,
		"resultado": resultado,
	}).Info("Exame concluído - certificado pode ser gerado")

	return exameAtualizado, nil
}
