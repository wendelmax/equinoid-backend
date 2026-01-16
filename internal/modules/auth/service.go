package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/equinoid/backend/internal/config"
	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/internal/modules/users"
	"github.com/equinoid/backend/pkg/auth"
	"github.com/equinoid/backend/pkg/cache"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Login(ctx context.Context, email, password string) (*models.TokenPair, *models.User, error)
	Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*models.TokenPair, *models.User, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	RevokeToken(ctx context.Context, token interface{}) error
}

type service struct {
	userRepo users.Repository
	cache    cache.CacheInterface
	logger   *logging.Logger
	config   *config.Config
}

func NewService(userRepo users.Repository, cache cache.CacheInterface, logger *logging.Logger, config *config.Config) Service {
	return &service{
		userRepo: userRepo,
		cache:    cache,
		logger:   logger,
		config:   config,
	}
}

func (s *service) Login(ctx context.Context, email, password string) (*models.TokenPair, *models.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, nil, &apperrors.AuthenticationError{Message: "credenciais inválidas"}
		}
		s.logger.LogError(err, "AuthService.Login", logging.Fields{"email": email})
		return nil, nil, err
	}

	if !user.IsActive {
		return nil, nil, &apperrors.ValidationError{Field: "is_active", Message: "usuário inativo"}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, nil, &apperrors.AuthenticationError{Message: "credenciais inválidas"}
	}

	tokenPair, err := auth.GenerateTokenPair(user.ID, user.Email, user.UserType, s.config.JWTSecret, int(s.config.JWTExpireHours))
	if err != nil {
		s.logger.LogError(err, "AuthService.Login", logging.Fields{"user_id": user.ID})
		return nil, nil, apperrors.NewBusinessError("TOKEN_GENERATION_FAILED", "erro ao gerar token", nil)
	}

	s.logger.WithFields(logging.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("Login realizado com sucesso")

	user.Password = ""
	return tokenPair, user, nil
}

func (s *service) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		s.logger.LogError(err, "AuthService.Register", logging.Fields{"email": req.Email})
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrUserEmailExists.WithValue(req.Email)
	}

	if req.CPFCNPJ != "" {
		exists, err := s.userRepo.ExistsByCPFCNPJ(ctx, req.CPFCNPJ)
		if err != nil {
			s.logger.LogError(err, "AuthService.Register", logging.Fields{"cpf_cnpj": req.CPFCNPJ})
			return nil, err
		}
		if exists {
			return nil, &apperrors.ConflictError{Resource: "cpf_cnpj", Message: "CPF/CNPJ já está em uso", Value: req.CPFCNPJ}
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.LogError(err, "AuthService.Register", logging.Fields{"email": req.Email})
		return nil, apperrors.NewDatabaseError("register", "erro ao gerar hash da senha", err)
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		CPFCNPJ:  req.CPFCNPJ,
		UserType: models.UserTypeCriador,
		IsActive: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.LogError(err, "AuthService.Register", logging.Fields{"email": req.Email})
		return nil, err
	}

	s.logger.WithFields(logging.Fields{
		"user_id":   user.ID,
		"email":     user.Email,
		"user_type": user.UserType,
	}).Info("Usuário registrado com sucesso")

	user.Password = ""
	return user, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenPair, *models.User, error) {
	claims, err := auth.ValidateRefreshToken(refreshToken, s.config.JWTSecret)
	if err != nil {
		return nil, nil, &apperrors.AuthenticationError{Message: "token inválido ou expirado"}
	}

	var isRevoked string
	err = s.cache.Get(ctx, fmt.Sprintf("revoked:refresh:%s", refreshToken), &isRevoked)
	if err == nil && isRevoked == "true" {
		return nil, nil, &apperrors.AuthenticationError{Message: "token revogado"}
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, nil, &apperrors.AuthenticationError{Message: "usuário não encontrado"}
		}
		s.logger.LogError(err, "AuthService.RefreshToken", logging.Fields{"user_id": claims.UserID})
		return nil, nil, err
	}

	if !user.IsActive {
		return nil, nil, &apperrors.ValidationError{Field: "is_active", Message: "usuário inativo"}
	}

	tokenPair, err := auth.GenerateTokenPair(user.ID, user.Email, user.UserType, s.config.JWTSecret, int(s.config.JWTExpireHours))
	if err != nil {
		s.logger.LogError(err, "AuthService.RefreshToken", logging.Fields{"user_id": user.ID})
		return nil, nil, apperrors.NewBusinessError("TOKEN_GENERATION_FAILED", "erro ao gerar token", nil)
	}

	if err := s.cache.Set(ctx, fmt.Sprintf("revoked:refresh:%s", refreshToken), "true", 7*24*time.Hour); err != nil {
		s.logger.LogError(err, "AuthService.RefreshToken", logging.Fields{"user_id": user.ID})
	}

	user.Password = ""
	return tokenPair, user, nil
}

func (s *service) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil
		}
		s.logger.LogError(err, "AuthService.ForgotPassword", logging.Fields{"email": email})
		return err
	}

	resetToken, err := auth.GenerateResetToken(user.ID, s.config.JWTSecret)
	if err != nil {
		s.logger.LogError(err, "AuthService.ForgotPassword", logging.Fields{"user_id": user.ID})
		return apperrors.NewBusinessError("TOKEN_GENERATION_FAILED", "erro ao gerar token de reset", nil)
	}

	if err := s.cache.Set(ctx, fmt.Sprintf("reset:%d", user.ID), resetToken, 1*time.Hour); err != nil {
		s.logger.LogError(err, "AuthService.ForgotPassword", logging.Fields{"user_id": user.ID})
		return apperrors.NewDatabaseError("forgot_password", "erro ao salvar token de reset", err)
	}

	s.logger.WithFields(logging.Fields{
		"user_id": user.ID,
		"email":   email,
	}).Info("Token de reset de senha gerado")

	return nil
}

func (s *service) ResetPassword(ctx context.Context, token, newPassword string) error {
	claims, err := auth.ValidateResetToken(token, s.config.JWTSecret)
	if err != nil {
		return &apperrors.AuthenticationError{Message: "token inválido ou expirado"}
	}

	var savedToken string
	err = s.cache.Get(ctx, fmt.Sprintf("reset:%d", claims.UserID), &savedToken)
	if err != nil || savedToken != token {
		return &apperrors.AuthenticationError{Message: "token inválido ou expirado"}
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return &apperrors.AuthenticationError{Message: "usuário não encontrado"}
		}
		s.logger.LogError(err, "AuthService.ResetPassword", logging.Fields{"user_id": claims.UserID})
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.LogError(err, "AuthService.ResetPassword", logging.Fields{"user_id": user.ID})
		return apperrors.NewDatabaseError("reset_password", "erro ao gerar hash da senha", err)
	}

	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.LogError(err, "AuthService.ResetPassword", logging.Fields{"user_id": user.ID})
		return err
	}

	if err := s.cache.Delete(ctx, fmt.Sprintf("reset:%d", user.ID)); err != nil {
		s.logger.LogError(err, "AuthService.ResetPassword", logging.Fields{"user_id": user.ID})
	}

	s.logger.WithFields(logging.Fields{"user_id": user.ID}).Info("Senha resetada com sucesso")

	return nil
}

func (s *service) RevokeToken(ctx context.Context, token interface{}) error {
	tokenStr, ok := token.(string)
	if !ok {
		return &apperrors.ValidationError{Field: "token", Message: "token inválido"}
	}

	if err := s.cache.Set(ctx, fmt.Sprintf("revoked:access:%s", tokenStr), "true", 24*time.Hour); err != nil {
		s.logger.LogError(err, "AuthService.RevokeToken", logging.Fields{"token": tokenStr})
		return apperrors.NewDatabaseError("revoke_token", "erro ao revogar token", err)
	}

	s.logger.WithFields(logging.Fields{"token": tokenStr}).Info("Token revogado com sucesso")

	return nil
}
