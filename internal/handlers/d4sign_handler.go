package handlers

import (
	"net/http"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// CreateD4SignDocument cria um novo documento na D4Sign
func (h *Handlers) CreateD4SignDocument(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success:   false,
			Error:     "Usuário não autenticado",
			Timestamp: time.Now(),
		})
		return
	}

	safeUUID := c.Param("safe_uuid")
	var req models.CreateD4SignDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos",
			Timestamp: time.Now(),
		})
		return
	}

	doc, err := h.D4SignService.CreateDocument(c.Request.Context(), safeUUID, req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao criar documento na D4Sign",
			Timestamp: time.Now(),
		})
		return
	}

	// Adicionar signatários
	if len(req.Signers) > 0 {
		if err := h.D4SignService.AddSigners(c.Request.Context(), doc.DocumentUUID, req.Signers); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success:   false,
				Error:     "Erro ao adicionar signatários",
				Timestamp: time.Now(),
			})
			return
		}
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Data:      map[string]interface{}{"uuid": doc.DocumentUUID, "id": doc.ID},
		Message:   "Documento criado com sucesso",
		Timestamp: time.Now(),
	})
}

// GetD4SignDocumentStatus busca o status de um documento
func (h *Handlers) GetD4SignDocumentStatus(c *gin.Context) {
	docUUID := c.Param("doc_uuid")
	status, err := h.D4SignService.GetDocumentStatus(c.Request.Context(), docUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao buscar status do documento: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      status,
		Timestamp: time.Now(),
	})
}

// GetD4SignEmbedURL gera URL para assinatura embutida
func (h *Handlers) GetD4SignEmbedURL(c *gin.Context) {
	docUUID := c.Param("doc_uuid")
	email := c.Query("email")

	url, err := h.D4SignService.GetEmbedURL(c.Request.Context(), docUUID, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao gerar URL de assinatura",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      map[string]string{"embed_url": url},
		Timestamp: time.Now(),
	})
}
