package users

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/pkg/cache"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	GetByID(ctx context.Context, id uint) (*models.User, error)
	UpdateProfile(ctx context.Context, id uint, req *models.UpdateProfileRequest) (*models.User, error)
	Delete(ctx context.Context, id uint) error
	ChangePassword(ctx context.Context, id uint, currentPassword, newPassword string) error
	IsEmailAvailable(ctx context.Context, email string) (bool, error)
	List(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.User, int64, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, req *models.RegisterRequest) (*models.User, error)
	UpdateUserByAdmin(ctx context.Context, id uint, req *models.UpdateUserAdminRequest) (*models.User, error)
	DeleteUserByAdmin(ctx context.Context, id uint) error
	ActivateUser(ctx context.Context, id uint) (*models.User, error)
	DeactivateUser(ctx context.Context, id uint) (*models.User, error)
}

type service struct {
	repo   Repository
	cache  cache.CacheInterface
	logger *logging.Logger
}

func NewService(repo Repository, cache cache.CacheInterface, logger *logging.Logger) Service {
	return &service{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

func (s *service) GetByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "UserService.GetByID", logging.Fields{"user_id": id})
		}
		return nil, err
	}
	user.Password = ""
	return user, nil
}

func (s *service) UpdateProfile(ctx context.Context, id uint, req *models.UpdateProfileRequest) (*models.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "UserService.UpdateProfile", logging.Fields{"user_id": id})
		}
		return nil, err
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.CPFCNPJ != "" {
		user.CPFCNPJ = req.CPFCNPJ
	}

	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.LogError(err, "UserService.UpdateProfile", logging.Fields{"user_id": id})
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "UserService.Delete", logging.Fields{"user_id": id})
		}
		return err
	}

	s.logger.WithFields(logging.Fields{"user_id": id}).Info("Usuário deletado com sucesso")
	return nil
}

func (s *service) ChangePassword(ctx context.Context, id uint, currentPassword, newPassword string) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "UserService.ChangePassword", logging.Fields{"user_id": id})
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return &apperrors.AuthenticationError{Message: "credenciais inválidas", Reason: "senha atual incorreta"}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.LogError(err, "UserService.ChangePassword", logging.Fields{"user_id": id})
		return apperrors.NewDatabaseError("change_password", "erro ao gerar hash da senha", err)
	}

	user.Password = string(hashedPassword)
	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.LogError(err, "UserService.ChangePassword", logging.Fields{"user_id": id})
		return err
	}

	s.logger.WithFields(logging.Fields{"user_id": id}).Info("Senha alterada com sucesso")
	return nil
}

func (s *service) IsEmailAvailable(ctx context.Context, email string) (bool, error) {
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		s.logger.LogError(err, "UserService.IsEmailAvailable", logging.Fields{"email": email})
		return false, err
	}
	return !exists, nil
}

func (s *service) List(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.User, int64, error) {
	users, total, err := s.repo.List(ctx, page, limit, filters)
	if err != nil {
		s.logger.LogError(err, "UserService.List", logging.Fields{"filters": filters})
		return nil, 0, err
	}
	return users, total, nil
}

func (s *service) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "UserService.FindByEmail", logging.Fields{"email": email})
		}
		return nil, err
	}
	return user, nil
}

func (s *service) CreateUser(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		s.logger.LogError(err, "UserService.CreateUser", logging.Fields{"email": req.Email})
		return nil, err
	}
	if exists {
		return nil, &apperrors.ValidationError{Message: "Email já está em uso"}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.LogError(err, "UserService.CreateUser", logging.Fields{"email": req.Email})
		return nil, apperrors.NewDatabaseError("create_user", "erro ao gerar hash da senha", err)
	}

	user := &models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
		UserType: req.UserType,
		CPFCNPJ:  req.CPFCNPJ,
		IsActive: true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.LogError(err, "UserService.CreateUser", logging.Fields{"email": req.Email})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{"user_id": user.ID, "email": user.Email}).Info("Novo usuário criado com sucesso")
	user.Password = ""
	return user, nil
}

func (s *service) UpdateUserByAdmin(ctx context.Context, id uint, req *models.UpdateUserAdminRequest) (*models.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "UserService.UpdateUserByAdmin", logging.Fields{"user_id": id})
		}
		return nil, err
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil && *req.Email != user.Email {
		exists, err := s.repo.ExistsByEmail(ctx, *req.Email)
		if err != nil {
			s.logger.LogError(err, "UserService.UpdateUserByAdmin", logging.Fields{"user_id": id})
			return nil, err
		}
		if exists {
			return nil, &apperrors.ValidationError{Message: "Email já está em uso"}
		}
		user.Email = *req.Email
	}
	if req.UserType != nil {
		user.UserType = *req.UserType
	}
	if req.CPFCNPJ != nil {
		user.CPFCNPJ = *req.CPFCNPJ
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.LogError(err, "UserService.UpdateUserByAdmin", logging.Fields{"user_id": id})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{"user_id": id}).Info("Usuário atualizado pelo admin com sucesso")
	user.Password = ""
	return user, nil
}

func (s *service) DeleteUserByAdmin(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "UserService.DeleteUserByAdmin", logging.Fields{"user_id": id})
		}
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.LogError(err, "UserService.DeleteUserByAdmin", logging.Fields{"user_id": id})
		return err
	}

	s.logger.WithFields(logging.Fields{"user_id": id}).Info("Usuário deletado pelo admin com sucesso")
	return nil
}

func (s *service) ActivateUser(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "UserService.ActivateUser", logging.Fields{"user_id": id})
		}
		return nil, err
	}

	user.IsActive = true
	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.LogError(err, "UserService.ActivateUser", logging.Fields{"user_id": id})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{"user_id": id}).Info("Usuário ativado com sucesso")
	user.Password = ""
	return user, nil
}

func (s *service) DeactivateUser(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "UserService.DeactivateUser", logging.Fields{"user_id": id})
		}
		return nil, err
	}

	user.IsActive = false
	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.LogError(err, "UserService.DeactivateUser", logging.Fields{"user_id": id})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{"user_id": id}).Info("Usuário desativado com sucesso")
	user.Password = ""
	return user, nil
}
