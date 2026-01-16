package users

import (
	"context"
	"errors"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"gorm.io/gorm"
)

type Repository interface {
	FindByID(ctx context.Context, id uint) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByCPFCNPJ(ctx context.Context, cpfCnpj string) (*models.User, error)
	FindBySupabaseID(ctx context.Context, supabaseID string) (*models.User, error)
	FindByKeycloakID(ctx context.Context, keycloakID string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByCPFCNPJ(ctx context.Context, cpfCnpj string) (bool, error)
	List(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.User, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "user", Message: "usuário não encontrado", ID: id}
		}
		return nil, apperrors.NewDatabaseError("find_by_id", "erro ao buscar usuário", err)
	}
	return &user, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "user", Message: "usuário não encontrado", ID: email}
		}
		return nil, apperrors.NewDatabaseError("find_by_email", "erro ao buscar usuário por email", err)
	}
	return &user, nil
}

func (r *repository) FindByCPFCNPJ(ctx context.Context, cpfCnpj string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("cpf_cnpj = ?", cpfCnpj).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "user", Message: "usuário não encontrado", ID: cpfCnpj}
		}
		return nil, apperrors.NewDatabaseError("find_by_cpf_cnpj", "erro ao buscar usuário por CPF/CNPJ", err)
	}
	return &user, nil
}

func (r *repository) FindBySupabaseID(ctx context.Context, supabaseID string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("supabase_id = ?", supabaseID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "user", Message: "usuário não encontrado", ID: supabaseID}
		}
		return nil, apperrors.NewDatabaseError("find_by_supabase_id", "erro ao buscar usuário por Supabase ID", err)
	}
	return &user, nil
}

func (r *repository) FindByKeycloakID(ctx context.Context, keycloakID string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("keycloak_id = ?", keycloakID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "user", Message: "usuário não encontrado", ID: keycloakID}
		}
		return nil, apperrors.NewDatabaseError("find_by_keycloak_id", "erro ao buscar usuário por Keycloak ID", err)
	}
	return &user, nil
}

func (r *repository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return apperrors.NewDatabaseError("create", "erro ao criar usuário", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return apperrors.NewDatabaseError("update", "erro ao atualizar usuário", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.User{}, id)
	if result.Error != nil {
		return apperrors.NewDatabaseError("delete", "erro ao deletar usuário", result.Error)
	}
	if result.RowsAffected == 0 {
		return &apperrors.NotFoundError{Resource: "user", Message: "usuário não encontrado", ID: id}
	}
	return nil
}

func (r *repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, apperrors.NewDatabaseError("exists_by_email", "erro ao verificar existência de email", err)
	}
	return count > 0, nil
}

func (r *repository) ExistsByCPFCNPJ(ctx context.Context, cpfCnpj string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("cpf_cnpj = ?", cpfCnpj).Count(&count).Error; err != nil {
		return false, apperrors.NewDatabaseError("exists_by_cpf_cnpj", "erro ao verificar existência de CPF/CNPJ", err)
	}
	return count > 0, nil
}

func (r *repository) List(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	query := r.db.WithContext(ctx).Model(&models.User{})

	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if userType, ok := filters["tipo_usuario"].(string); ok && userType != "" {
		query = query.Where("user_type = ?", userType)
	}
	if isActive, ok := filters["is_active"].(bool); ok {
		query = query.Where("is_active = ?", isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.NewDatabaseError("list_count", "erro ao contar usuários", err)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, apperrors.NewDatabaseError("list", "erro ao listar usuários", err)
	}

	for _, user := range users {
		user.Password = ""
	}

	return users, total, nil
}
