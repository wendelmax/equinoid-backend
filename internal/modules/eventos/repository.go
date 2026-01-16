package eventos

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"gorm.io/gorm"
)

type Repository interface {
	ListAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.EventoResponse, int64, error)
	FindByID(ctx context.Context, id uint) (*models.EventoResponse, error)
	FindByEquino(ctx context.Context, equinoid string) ([]*models.EventoResponse, error)
	Create(ctx context.Context, evento *models.Evento) error
	Update(ctx context.Context, evento *models.Evento) error
	Delete(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.EventoResponse, int64, error) {
	var eventos []*models.Evento
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Evento{})

	if categoria, ok := filters["categoria"].(string); ok && categoria != "" {
		query = query.Where("categoria = ?", categoria)
	}
	if tipoEvento, ok := filters["tipo_evento"].(string); ok && tipoEvento != "" {
		query = query.Where("tipo_evento = ?", tipoEvento)
	}
	if dataInicio, ok := filters["data_inicio"].(string); ok && dataInicio != "" {
		query = query.Where("data_evento >= ?", dataInicio)
	}
	if dataFim, ok := filters["data_fim"].(string); ok && dataFim != "" {
		query = query.Where("data_evento <= ?", dataFim)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.NewDatabaseError("list_eventos", "erro ao contar eventos", err)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("data_evento DESC").Find(&eventos).Error; err != nil {
		return nil, 0, apperrors.NewDatabaseError("list_eventos", "erro ao buscar eventos", err)
	}

	responses := make([]*models.EventoResponse, len(eventos))
	for i, evento := range eventos {
		responses[i] = evento.ToResponse()
	}

	return responses, total, nil
}

func (r *repository) FindByID(ctx context.Context, id uint) (*models.EventoResponse, error) {
	var evento models.Evento
	if err := r.db.WithContext(ctx).First(&evento, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &apperrors.NotFoundError{Resource: "evento", Message: "evento não encontrado"}
		}
		return nil, apperrors.NewDatabaseError("find_evento", "erro ao buscar evento", err)
	}
	return evento.ToResponse(), nil
}

func (r *repository) FindByEquino(ctx context.Context, equinoid string) ([]*models.EventoResponse, error) {
	var eventos []*models.Evento
	
	if err := r.db.WithContext(ctx).
		Joins("JOIN equinos ON equinos.id = eventos.equino_id").
		Where("equinos.equinoid = ?", equinoid).
		Order("eventos.data_evento DESC").
		Find(&eventos).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_eventos_by_equino", "erro ao buscar eventos do equino", err)
	}

	responses := make([]*models.EventoResponse, len(eventos))
	for i, evento := range eventos {
		responses[i] = evento.ToResponse()
	}

	return responses, nil
}

func (r *repository) Create(ctx context.Context, evento *models.Evento) error {
	if err := r.db.WithContext(ctx).Create(evento).Error; err != nil {
		return apperrors.NewDatabaseError("create_evento", "erro ao criar evento", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, evento *models.Evento) error {
	if err := r.db.WithContext(ctx).Model(&models.Evento{}).Where("id = ?", evento.ID).Updates(evento).Error; err != nil {
		return apperrors.NewDatabaseError("update_evento", "erro ao atualizar evento", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Evento{}, id)
	if result.Error != nil {
		return apperrors.NewDatabaseError("delete_evento", "erro ao deletar evento", result.Error)
	}
	if result.RowsAffected == 0 {
		return &apperrors.NotFoundError{Resource: "evento", Message: "evento não encontrado"}
	}
	return nil
}
