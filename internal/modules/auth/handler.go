package auth

import (
	"net/http"
	"time"

	"github.com/equinoid/backend/internal/models"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
	logger  *logging.Logger
}

func NewHandler(service Service, logger *logging.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Login godoc
// @Summary Login de usuário
// @Description Autentica um usuário e retorna tokens JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param login body models.LoginRequest true "Credenciais de login"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	tokenPair, user, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if apperrors.IsAuthentication(err) {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success:   false,
				Error:     "Credenciais inválidas",
				Timestamp: time.Now(),
			})
			return
		}

		if apperrors.IsValidation(err) {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao realizar login",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Login realizado com sucesso",
		Timestamp: time.Now(),
		Data: gin.H{
			"user":          user,
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"expires_in":    tokenPair.ExpiresIn,
		},
	})
}

// Register godoc
// @Summary Registro de novo usuário
// @Description Cria uma nova conta de usuário no sistema
// @Tags auth
// @Accept json
// @Produce json
// @Param register body models.RegisterRequest true "Dados de registro"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	user, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		if apperrors.IsConflict(err) {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao registrar usuário",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Message:   "Usuário registrado com sucesso",
		Timestamp: time.Now(),
		Data:      user,
	})
}

// RefreshToken godoc
// @Summary Atualizar token de acesso
// @Description Gera um novo par de tokens usando o refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body object{refresh_token=string} true "Refresh token"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	tokenPair, user, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if apperrors.IsAuthentication(err) || apperrors.IsValidation(err) {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao atualizar token",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Token atualizado com sucesso",
		Timestamp: time.Now(),
		Data: gin.H{
			"user":          user,
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"expires_in":    tokenPair.ExpiresIn,
		},
	})
}

// ForgotPassword godoc
// @Summary Esqueci minha senha
// @Description Envia email para recuperação de senha
// @Tags auth
// @Accept json
// @Produce json
// @Param forgot body object{email=string} true "Email do usuário"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	if err := h.service.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao processar solicitação",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Se o email existir, você receberá instruções para redefinir sua senha",
		Timestamp: time.Now(),
		Data:      nil,
	})
}

// ResetPassword godoc
// @Summary Redefinir senha
// @Description Redefine a senha do usuário usando token de recuperação
// @Tags auth
// @Accept json
// @Produce json
// @Param reset body object{token=string,new_password=string} true "Token e nova senha"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		if apperrors.IsAuthentication(err) {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success:   false,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao redefinir senha",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Senha redefinida com sucesso",
		Timestamp: time.Now(),
		Data:      nil,
	})
}

// Logout godoc
// @Summary Logout de usuário
// @Description Invalida o token atual do usuário
// @Tags auth
// @Produce json
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/logout [post]
// @Security BearerAuth
func (h *Handler) Logout(c *gin.Context) {
	token, exists := c.Get("token")
	if !exists {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Token não encontrado",
			Timestamp: time.Now(),
		})
		return
	}

	if err := h.service.RevokeToken(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao fazer logout",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Logout realizado com sucesso",
		Timestamp: time.Now(),
		Data:      nil,
	})
}
