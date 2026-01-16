package handlers

import (
	"net/http"
	"time"

	"github.com/equinoid/backend/internal/middleware"
	"github.com/equinoid/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// GenerateCertificate gera um novo certificado para um equino
func (h *Handlers) GenerateCertificate(c *gin.Context) {
	var req struct {
		Equinoid        string `json:"equinoid" binding:"required"`
		TipoCertificado string `json:"tipo_certificado" binding:"required"`
		ValidDays       int    `json:"valid_days"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos",
			Timestamp: time.Now(),
		})
		return
	}

	userID, _ := middleware.GetUserIDFromContext(c)
	if req.ValidDays <= 0 {
		req.ValidDays = 365
	}

	cert, err := h.CertificateService.GenerateCertificate(c.Request.Context(), userID, req.Equinoid, req.TipoCertificado, req.ValidDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao gerar certificado",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Data:      cert.ToResponse(),
		Message:   "Certificado gerado com sucesso",
		Timestamp: time.Now(),
	})
}

// ListCertificates lista os certificados do usuário autenticado
func (h *Handlers) ListCertificates(c *gin.Context) {
	userID, _ := middleware.GetUserIDFromContext(c)

	certs, err := h.CertificateService.ListCertificates(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao listar certificados",
			Timestamp: time.Now(),
		})
		return
	}

	var response []models.CertificateResponse
	for _, cert := range certs {
		response = append(response, *cert.ToResponse())
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      response,
		Timestamp: time.Now(),
	})
}

// ValidateCertificate verifica a validade de um certificado
func (h *Handlers) ValidateCertificate(c *gin.Context) {
	serial := c.Param("serial")

	cert, err := h.CertificateService.ValidateCertificate(c.Request.Context(), serial)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success:   false,
			Error:     "Certificado não encontrado",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      cert.ToResponse(),
		Timestamp: time.Now(),
	})
}

// RevokeCertificate revoga um certificado existente
func (h *Handlers) RevokeCertificate(c *gin.Context) {
	var req struct {
		SerialNumber string `json:"serial_number" binding:"required"`
		Motivo       string `json:"motivo" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos",
			Timestamp: time.Now(),
		})
		return
	}

	err := h.CertificateService.RevokeCertificate(c.Request.Context(), req.SerialNumber, req.Motivo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao revogar certificado",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Certificado revogado com sucesso",
		Timestamp: time.Now(),
	})
}
