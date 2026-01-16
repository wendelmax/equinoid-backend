package equinos

import (
	"context"
	"errors"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"gorm.io/gorm"
)

type Repository interface {
	FindByID(ctx context.Context, id uint) (*models.Equino, error)
	FindByEquinoid(ctx context.Context, equinoid string) (*models.Equino, error)
	FindByMicrochipID(ctx context.Context, microchipID string) (*models.Equino, error)
	FindByProprietarioID(ctx context.Context, proprietarioID uint) ([]*models.Equino, error)
	FindByGenitor(ctx context.Context, genitorEquinoid string) ([]*models.Equino, error)
	FindByGenitora(ctx context.Context, genitoraEquinoid string) ([]*models.Equino, error)
	Create(ctx context.Context, equino *models.Equino) error
	Update(ctx context.Context, equino *models.Equino) error
	Delete(ctx context.Context, equinoid string) error
	List(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.Equino, int64, error)
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
	ExistsByEquinoid(ctx context.Context, equinoid string) (bool, error)
	ExistsByMicrochipID(ctx context.Context, microchipID string) (bool, error)
	TransferOwnership(ctx context.Context, equinoid string, newOwnerID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByID(ctx context.Context, id uint) (*models.Equino, error) {
	var equino models.Equino
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&equino).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado", ID: id}
		}
		return nil, apperrors.NewDatabaseError("find_by_id", "erro ao buscar equino", err)
	}
	return &equino, nil
}

func (r *repository) FindByEquinoid(ctx context.Context, equinoid string) (*models.Equino, error) {
	var equino models.Equino
	if err := r.db.WithContext(ctx).Where("equinoid = ?", equinoid).First(&equino).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado", ID: equinoid}
		}
		return nil, apperrors.NewDatabaseError("find_by_equinoid", "erro ao buscar equino por EquinoId", err)
	}
	return &equino, nil
}

func (r *repository) FindByMicrochipID(ctx context.Context, microchipID string) (*models.Equino, error) {
	var equino models.Equino
	if err := r.db.WithContext(ctx).Where("microchip_id = ?", microchipID).First(&equino).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado", ID: microchipID}
		}
		return nil, apperrors.NewDatabaseError("find_by_microchip_id", "erro ao buscar equino por MicrochipID", err)
	}
	return &equino, nil
}

func (r *repository) FindByProprietarioID(ctx context.Context, proprietarioID uint) ([]*models.Equino, error) {
	var equinos []*models.Equino
	if err := r.db.WithContext(ctx).Where("proprietario_id = ?", proprietarioID).Find(&equinos).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_by_proprietario_id", "erro ao buscar equinos do proprietário", err)
	}
	return equinos, nil
}

func (r *repository) FindByGenitor(ctx context.Context, genitorEquinoid string) ([]*models.Equino, error) {
	var equinos []*models.Equino
	if err := r.db.WithContext(ctx).Where("genitor_equinoid = ?", genitorEquinoid).Find(&equinos).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_by_genitor", "erro ao buscar descendentes do genitor", err)
	}
	return equinos, nil
}

func (r *repository) FindByGenitora(ctx context.Context, genitoraEquinoid string) ([]*models.Equino, error) {
	var equinos []*models.Equino
	if err := r.db.WithContext(ctx).Where("genitora_equinoid = ?", genitoraEquinoid).Find(&equinos).Error; err != nil {
		return nil, apperrors.NewDatabaseError("find_by_genitora", "erro ao buscar descendentes da genitora", err)
	}
	return equinos, nil
}

func (r *repository) Create(ctx context.Context, equino *models.Equino) error {
	if err := r.db.WithContext(ctx).Create(equino).Error; err != nil {
		return apperrors.NewDatabaseError("create", "erro ao criar equino", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, equino *models.Equino) error {
	if err := r.db.WithContext(ctx).Save(equino).Error; err != nil {
		return apperrors.NewDatabaseError("update", "erro ao atualizar equino", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, equinoid string) error {
	result := r.db.WithContext(ctx).Where("equinoid = ?", equinoid).Delete(&models.Equino{})
	if result.Error != nil {
		return apperrors.NewDatabaseError("delete", "erro ao deletar equino", result.Error)
	}
	if result.RowsAffected == 0 {
		return &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado", ID: equinoid}
	}
	return nil
}

func (r *repository) List(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.Equino, int64, error) {
	var equinos []*models.Equino
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Equino{})

	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("nome ILIKE ? OR equinoid ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if raca, ok := filters["raca"].(string); ok && raca != "" {
		query = query.Where("raca ILIKE ?", "%"+raca+"%")
	}
	if proprietarioID, ok := filters["proprietario_id"].(uint); ok {
		query = query.Where("proprietario_id = ?", proprietarioID)
	}
	if veterinarioID, ok := filters["veterinario_id"].(uint); ok {
		query = query.Where("veterinario_id = ?", veterinarioID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.NewDatabaseError("list_count", "erro ao contar equinos", err)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&equinos).Error; err != nil {
		return nil, 0, apperrors.NewDatabaseError("list", "erro ao listar equinos", err)
	}

	return equinos, total, nil
}

func (r *repository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Equino{})

	if proprietarioID, ok := filters["proprietario_id"].(uint); ok {
		query = query.Where("proprietario_id = ?", proprietarioID)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, apperrors.NewDatabaseError("count", "erro ao contar equinos", err)
	}

	return count, nil
}

func (r *repository) ExistsByEquinoid(ctx context.Context, equinoid string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Equino{}).Where("equinoid = ?", equinoid).Count(&count).Error; err != nil {
		return false, apperrors.NewDatabaseError("exists_by_equinoid", "erro ao verificar existência de equinoid", err)
	}
	return count > 0, nil
}

func (r *repository) ExistsByMicrochipID(ctx context.Context, microchipID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Equino{}).Where("microchip_id = ?", microchipID).Count(&count).Error; err != nil {
		return false, apperrors.NewDatabaseError("exists_by_microchip", "erro ao verificar existência de microchip", err)
	}
	return count > 0, nil
}

func (r *repository) TransferOwnership(ctx context.Context, equinoid string, newOwnerID uint) error {
	result := r.db.WithContext(ctx).Model(&models.Equino{}).
		Where("equinoid = ?", equinoid).
		Update("proprietario_id", newOwnerID)

	if result.Error != nil {
		return apperrors.NewDatabaseError("transfer_ownership", "erro ao transferir propriedade", result.Error)
	}
	if result.RowsAffected == 0 {
		return &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado", ID: equinoid}
	}

	return nil
}
