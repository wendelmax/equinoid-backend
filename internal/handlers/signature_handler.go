package handlers

import (
	"net/http"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type CreateSignatureRequest struct {
	DocumentType      string                 `json:"document_type" binding:"required"`
	DocumentHash      string                 `json:"document_hash" binding:"required"`
	DocumentData      map[string]interface{} `json:"document_data"`
	RelatedEntityID   *uint                  `json:"related_entity_id"`
	RelatedEntityType string                 `json:"related_entity_type"`
	Signers           []models.D4SignSigner  `json:"signers,omitempty"`
	Base64File        string                 `json:"base64_file,omitempty"`
	BiometricData     []byte                 `json:"biometric_data,omitempty"`
	CertificateID     *uint                  `json:"certificate_id,omitempty"`
}

type SignatureResponse struct {
	Success       bool   `json:"success"`
	Method        string `json:"method"`
	DocumentUUID  string `json:"document_uuid,omitempty"`
	SignatureHash string `json:"signature_hash,omitempty"`
	Status        string `json:"status"`
	Message       string `json:"message"`
	EmbedURL      string `json:"embed_url,omitempty"`
}

func (h *Handlers) CreateSignature(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success:   false,
			Error:     "Usuário não autenticado",
			Timestamp: time.Now(),
		})
		return
	}

	var req CreateSignatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Dados inválidos: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	router := services.NewSignatureRouter(h.Logger)
	shouldUseD4Sign := router.ShouldUseD4Sign(req.DocumentType)

	if shouldUseD4Sign {
		h.createD4SignSignature(c, req, userID.(uint))
	} else {
		h.createPKISignature(c, req, userID.(uint))
	}
}

func (h *Handlers) createD4SignSignature(c *gin.Context, req CreateSignatureRequest, userID uint) {
	if req.Base64File == "" || len(req.Signers) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Para documentos D4Sign, base64_file e signers são obrigatórios",
			Timestamp: time.Now(),
		})
		return
	}

	d4SignReq := models.CreateD4SignDocumentRequest{
		Base64File:        req.Base64File,
		Name:              req.DocumentType + "_" + time.Now().Format("20060102_150405"),
		DocumentType:      req.DocumentType,
		RelatedEntityID:   req.RelatedEntityID,
		RelatedEntityType: req.RelatedEntityType,
		Signers:           req.Signers,
	}

	doc, err := h.D4SignService.CreateDocument(c.Request.Context(), "", d4SignReq, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao criar documento na D4Sign: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	if err := h.D4SignService.AddSigners(c.Request.Context(), doc.DocumentUUID, req.Signers); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao adicionar signatários: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	if err := h.D4SignService.SendToSign(c.Request.Context(), doc.DocumentUUID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao enviar documento para assinatura: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	embedURL, _ := h.D4SignService.GetEmbedURL(c.Request.Context(), doc.DocumentUUID, req.Signers[0].Email)

	c.JSON(http.StatusCreated, SignatureResponse{
		Success:      true,
		Method:       "d4sign",
		DocumentUUID: doc.DocumentUUID,
		Status:       "pending",
		Message:      "Documento criado e enviado para assinatura via D4Sign",
		EmbedURL:     embedURL,
	})
}

func (h *Handlers) createPKISignature(c *gin.Context, req CreateSignatureRequest, userID uint) {
	if req.BiometricData == nil || req.CertificateID == nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Para assinatura PKI, biometric_data e certificate_id são obrigatórios",
			Timestamp: time.Now(),
		})
		return
	}

	var cert models.Certificate
	if err := h.DB.First(&cert, *req.CertificateID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success:   false,
			Error:     "Certificado não encontrado",
			Timestamp: time.Now(),
		})
		return
	}

	if cert.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Success:   false,
			Error:     "Certificado não pertence ao usuário",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Success:   false,
		Error:     "Assinatura PKI interna será implementada em breve. Use o endpoint específico de assinatura biométrica.",
		Timestamp: time.Now(),
	})
}

func (h *Handlers) HandleD4SignWebhook(c *gin.Context) {
	var payload models.D4SignWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success:   false,
			Error:     "Payload inválido: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	_, err := h.D4SignService.GetDocumentByUUID(c.Request.Context(), payload.Document.UUID)
	if err != nil {
		h.Logger.WithContext(c.Request.Context()).Errorf("Documento não encontrado no webhook: %s", payload.Document.UUID)
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success:   false,
			Error:     "Documento não encontrado",
			Timestamp: time.Now(),
		})
		return
	}

	var signedAt *time.Time
	if payload.Event == "document_signed" {
		now := time.Now()
		signedAt = &now
	}

	status := "pending"
	switch payload.Event {
	case "document_signed":
		status = "signed"
	case "document_cancelled":
		status = "cancelled"
	case "document_expired":
		status = "expired"
	}

	if err := h.D4SignService.UpdateDocumentStatus(c.Request.Context(), payload.Document.UUID, status, signedAt); err != nil {
		h.Logger.WithContext(c.Request.Context()).Errorf("Erro ao atualizar status do documento: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success:   false,
			Error:     "Erro ao processar webhook",
			Timestamp: time.Now(),
		})
		return
	}

	h.Logger.WithContext(c.Request.Context()).Infof("Webhook processado: documento %s, evento %s, status %s", payload.Document.UUID, payload.Event, status)

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Message:   "Webhook processado com sucesso",
		Timestamp: time.Now(),
	})
}
