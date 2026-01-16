package exames

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context, filters map[string]interface{}) ([]*models.ExameLaboratorial, error)
	FindByID(ctx context.Context, id uint) (*models.ExameLaboratorial, error)
	Create(ctx context.Context, exame *models.ExameLaboratorial) error
	Update(ctx context.Context, exame *models.ExameLaboratorial) error
	Delete(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, filters map[string]interface{}) ([]*models.ExameLaboratorial, error) {
	var exames []*models.ExameLaboratorial
	query := r.db.WithContext(ctx).
		Preload("Equino").
		Preload("VeterinarioSolicitante").
		Preload("Laboratorio")

	if equinoid, ok := filters["equinoid"].(string); ok && equinoid != "" {
		query = query.Joins("JOIN equinos ON equinos.id = exames_laboratoriais.equino_id").
			Where("equinos.equinoid = ?", equinoid)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if tipoExame, ok := filters["tipo_exame"].(string); ok && tipoExame != "" {
		query = query.Where("tipo_exame = ?", tipoExame)
	}
	if veterinarioID, ok := filters["veterinario_id"].(uint); ok && veterinarioID > 0 {
		query = query.Where("veterinario_solicitante_id = ?", veterinarioID)
	}
	if laboratorioID, ok := filters["laboratorio_id"].(uint); ok && laboratorioID > 0 {
		query = query.Where("laboratorio_id = ?", laboratorioID)
	}

	if err := query.Order("data_solicitacao DESC").Find(&exames).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_exames", "erro ao buscar exames", err)
	}
	return exames, nil
}

func (r *repository) FindByID(ctx context.Context, id uint) (*models.ExameLaboratorial, error) {
	var exame models.ExameLaboratorial
	if err := r.db.WithContext(ctx).
		Preload("Equino").
		Preload("VeterinarioSolicitante").
		Preload("Laboratorio").
		First(&exame, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &apperrors.NotFoundError{Resource: "exame", Message: "exame não encontrado"}
		}
		return nil, apperrors.NewDatabaseError("find_exame", "erro ao buscar exame", err)
	}
	return &exame, nil
}

func (r *repository) Create(ctx context.Context, exame *models.ExameLaboratorial) error {
	if err := r.db.WithContext(ctx).Create(exame).Error; err != nil {
		return apperrors.NewDatabaseError("create_exame", "erro ao criar exame", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, exame *models.ExameLaboratorial) error {
	if err := r.db.WithContext(ctx).Save(exame).Error; err != nil {
		return apperrors.NewDatabaseError("update_exame", "erro ao atualizar exame", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.ExameLaboratorial{}, id)
	if result.Error != nil {
		return apperrors.NewDatabaseError("delete_exame", "erro ao deletar exame", result.Error)
	}
	if result.RowsAffected == 0 {
		return &apperrors.NotFoundError{Resource: "exame", Message: "exame não encontrado"}
	}
	return nil
}
