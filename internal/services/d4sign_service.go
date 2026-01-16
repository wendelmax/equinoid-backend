package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/equinoid/backend/internal/config"
	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/pkg/logging"
	"gorm.io/gorm"
)

type D4SignService struct {
	db     *gorm.DB
	logger *logging.Logger
	config *config.Config
}

func NewD4SignService(db *gorm.DB, logger *logging.Logger, config *config.Config) *D4SignService {
	return &D4SignService{
		db:     db,
		logger: logger,
		config: config,
	}
}

func (s *D4SignService) createRequest(method, endpoint string, body interface{}) (*http.Request, error) {
	u, err := url.Parse(s.config.D4SignAPIURL + endpoint)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("tokenAPI", s.config.D4SignTokenAPI)
	if s.config.D4SignCryptKey != "" {
		q.Set("cryptKey", s.config.D4SignCryptKey)
	}
	u.RawQuery = q.Encode()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (s *D4SignService) doRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	req, err := s.createRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	return client.Do(req.WithContext(ctx))
}

func (s *D4SignService) CreateDocument(ctx context.Context, safeUUID string, req models.CreateD4SignDocumentRequest, createdBy uint) (*models.D4SignDocument, error) {
	if safeUUID == "" {
		safeUUID = s.config.D4SignSafeUUID
	}

	endpoint := fmt.Sprintf("/documents/%s/upload", safeUUID)

	s.logger.WithContext(ctx).Infof("Enviando documento '%s' para D4Sign no cofre %s", req.Name, safeUUID)

	resp, err := s.doRequest(ctx, "POST", endpoint, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call D4Sign API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("D4Sign API returned status %d", resp.StatusCode)
	}

	var result struct {
		UUID string `json:"uuid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode D4Sign response: %w", err)
	}

	doc := &models.D4SignDocument{
		DocumentUUID:      result.UUID,
		SafeUUID:          safeUUID,
		Name:              req.Name,
		Status:            "pending",
		DocumentType:      req.DocumentType,
		RelatedEntityID:   req.RelatedEntityID,
		RelatedEntityType: req.RelatedEntityType,
		CreatedBy:         createdBy,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.SaveDocument(ctx, doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *D4SignService) AddSigners(ctx context.Context, documentUUID string, signers []models.D4SignSigner) error {
	endpoint := fmt.Sprintf("/documents/%s/createsigner", documentUUID)

	s.logger.WithContext(ctx).Infof("Adicionando %d signatários ao documento %s", len(signers), documentUUID)

	payload := map[string]interface{}{
		"signers": signers,
	}

	resp, err := s.doRequest(ctx, "POST", endpoint, payload)
	if err != nil {
		return fmt.Errorf("failed to add signers to D4Sign: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("D4Sign API returned status %d when adding signers", resp.StatusCode)
	}

	return nil
}

func (s *D4SignService) SendToSign(ctx context.Context, documentUUID string) error {
	endpoint := fmt.Sprintf("/documents/%s/sendtosigner", documentUUID)

	s.logger.WithContext(ctx).Infof("Enviando documento %s para assinatura", documentUUID)

	payload := map[string]interface{}{
		"message":  "Assinatura de documento EquinoId",
		"workflow": "0", // Assinatura em qualquer ordem
	}

	resp, err := s.doRequest(ctx, "POST", endpoint, payload)
	if err != nil {
		return fmt.Errorf("failed to send to sign: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("D4Sign API returned status %d when sending to sign", resp.StatusCode)
	}

	return nil
}

func (s *D4SignService) GetDocumentStatus(ctx context.Context, documentUUID string) (*models.D4SignDocumentStatusResponse, error) {
	endpoint := fmt.Sprintf("/documents/%s", documentUUID)

	s.logger.WithContext(ctx).Infof("Buscando status do documento %s na D4Sign", documentUUID)

	resp, err := s.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call D4Sign API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("D4Sign API returned status %d", resp.StatusCode)
	}

	var result models.D4SignDocumentStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode D4Sign response: %w", err)
	}

	return &result, nil
}

func (s *D4SignService) SaveDocument(ctx context.Context, doc *models.D4SignDocument) error {
	if err := s.db.WithContext(ctx).Create(doc).Error; err != nil {
		return fmt.Errorf("failed to save D4Sign document: %w", err)
	}
	return nil
}

func (s *D4SignService) UpdateDocumentStatus(ctx context.Context, documentUUID string, status string, signedAt *time.Time) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if signedAt != nil {
		updates["signed_at"] = signedAt
	}

	if err := s.db.WithContext(ctx).
		Model(&models.D4SignDocument{}).
		Where("document_uuid = ?", documentUUID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}

	s.logger.WithContext(ctx).Infof("Status do documento %s atualizado para %s", documentUUID, status)
	return nil
}

func (s *D4SignService) GetDocumentByUUID(ctx context.Context, documentUUID string) (*models.D4SignDocument, error) {
	var doc models.D4SignDocument
	if err := s.db.WithContext(ctx).
		Where("document_uuid = ?", documentUUID).
		First(&doc).Error; err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}
	return &doc, nil
}

func (s *D4SignService) GetEmbedURL(ctx context.Context, documentUUID string, signerEmail string) (string, error) {
	// Endpoint para gerar link de assinatura embutida
	s.logger.WithContext(ctx).Infof("Gerando URL de embed para %s assinar o documento %s", signerEmail, documentUUID)

	return fmt.Sprintf("https://secure.d4sign.com.br/embed/%s?email=%s", documentUUID, signerEmail), nil
}
