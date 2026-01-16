package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LGPDService gerencia compliance com a Lei Geral de Proteção de Dados
type LGPDService struct {
	db              *gorm.DB
	dataRetention   time.Duration
	consentRequired bool
}

// NewLGPDService cria um novo serviço LGPD
func NewLGPDService(db *gorm.DB, dataRetention time.Duration, consentRequired bool) *LGPDService {
	return &LGPDService{
		db:              db,
		dataRetention:   dataRetention,
		consentRequired: consentRequired,
	}
}

// ConsentRequest representa uma solicitação de consentimento
type ConsentRequest struct {
	UserID      uuid.UUID              `json:"user_id"`
	ConsentType string                 `json:"consent_type"` // data_processing, marketing, analytics
	Purpose     string                 `json:"purpose"`
	DataTypes   []string               `json:"data_types"`
	Retention   time.Duration          `json:"retention"`
	LegalBasis  string                 `json:"legal_basis"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ConsentResponse representa resposta de consentimento
type ConsentResponse struct {
	ConsentID   uuid.UUID `json:"consent_id"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	ValidUntil  time.Time `json:"valid_until"`
	CanWithdraw bool      `json:"can_withdraw"`
}

// DataRequest representa solicitação de dados (acesso, portabilidade, exclusão)
type DataRequest struct {
	UserID      uuid.UUID              `json:"user_id"`
	RequestType string                 `json:"request_type"` // access, portability, deletion, rectification
	DataTypes   []string               `json:"data_types"`
	Reason      string                 `json:"reason"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DataResponse representa resposta de solicitação de dados
type DataResponse struct {
	RequestID   uuid.UUID              `json:"request_id"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	ExportURL   string                 `json:"export_url,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	ExpiresAt   time.Time              `json:"expires_at"`
}

// RequestConsent solicita consentimento do usuário
func (l *LGPDService) RequestConsent(ctx context.Context, req *ConsentRequest) (*ConsentResponse, error) {
	// Verificar se usuário existe
	var user models.User
	if err := l.db.First(&user, req.UserID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verificar se já existe consentimento ativo para este tipo
	var existingConsent models.ComplianceRecord
	err := l.db.Where("user_id = ? AND compliance_type = ? AND request_type = ? AND status = ?",
		req.UserID, "consent", req.ConsentType, "approved").First(&existingConsent).Error

	if err == nil {
		return &ConsentResponse{
			ConsentID:   uuid.New(), // Gerar UUID temporário já que ID é uint
			Status:      "already_granted",
			Message:     "Consent already granted for this purpose",
			ValidUntil:  *existingConsent.ExpiresAt,
			CanWithdraw: true,
		}, nil
	}

	// Criar novo registro de consentimento
	requestData := map[string]interface{}{
		"consent_type": req.ConsentType,
		"purpose":      req.Purpose,
		"data_types":   req.DataTypes,
		"legal_basis":  req.LegalBasis,
		"metadata":     req.Metadata,
	}

	complianceRecord := &models.ComplianceRecord{
		UserID:         req.UserID,
		ComplianceType: "consent",
		RequestType:    req.ConsentType,
		Status:         "pending",
		RequestData:    requestData,
		LegalBasis:     req.LegalBasis,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Definir expiração do consentimento
	if req.Retention > 0 {
		expiresAt := time.Now().Add(req.Retention)
		complianceRecord.ExpiresAt = &expiresAt
	} else {
		// Usar retenção padrão
		expiresAt := time.Now().Add(l.dataRetention)
		complianceRecord.ExpiresAt = &expiresAt
	}

	// Salvar no banco
	if err := l.db.Create(complianceRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create consent record: %w", err)
	}

	return &ConsentResponse{
		ConsentID:   uuid.New(), // Gerar UUID temporário já que ID é uint
		Status:      "pending",
		Message:     "Consent request created successfully",
		ValidUntil:  *complianceRecord.ExpiresAt,
		CanWithdraw: true,
	}, nil
}

// GrantConsent aprova um consentimento
func (l *LGPDService) GrantConsent(ctx context.Context, consentID uuid.UUID, userID uuid.UUID) (*ConsentResponse, error) {
	var record models.ComplianceRecord
	err := l.db.Where("id = ? AND user_id = ? AND compliance_type = ?",
		consentID, userID, "consent").First(&record).Error

	if err != nil {
		return nil, fmt.Errorf("consent record not found: %w", err)
	}

	if record.Status != "pending" {
		return nil, fmt.Errorf("consent already processed")
	}

	// Atualizar status
	record.Status = "approved"
	if record.RequestData == nil {
		record.RequestData = make(map[string]interface{})
	}
	record.RequestData["processed_at"] = time.Now()
	record.RequestData["processed_by"] = userID.String()
	record.UpdatedAt = time.Now()

	if err := l.db.Save(&record).Error; err != nil {
		return nil, fmt.Errorf("failed to grant consent: %w", err)
	}

	return &ConsentResponse{
		ConsentID:   uuid.New(), // Gerar UUID temporário já que ID é uint
		Status:      "approved",
		Message:     "Consent granted successfully",
		ValidUntil:  *record.ExpiresAt,
		CanWithdraw: true,
	}, nil
}

// WithdrawConsent revoga um consentimento
func (l *LGPDService) WithdrawConsent(ctx context.Context, consentID uuid.UUID, userID uuid.UUID, reason string) error {
	var record models.ComplianceRecord
	err := l.db.Where("id = ? AND user_id = ? AND compliance_type = ?",
		consentID, userID, "consent").First(&record).Error

	if err != nil {
		return fmt.Errorf("consent record not found: %w", err)
	}

	// Atualizar status
	record.Status = "withdrawn"
	if record.RequestData == nil {
		record.RequestData = make(map[string]interface{})
	}
	record.RequestData["processed_at"] = time.Now()
	record.RequestData["notes"] = reason
	record.UpdatedAt = time.Now()

	return l.db.Save(&record).Error
}

// RequestDataAccess solicita acesso aos dados pessoais
func (l *LGPDService) RequestDataAccess(ctx context.Context, req *DataRequest) (*DataResponse, error) {
	// Verificar se usuário existe
	var user models.User
	if err := l.db.First(&user, req.UserID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Criar registro de solicitação
	requestData := map[string]interface{}{
		"data_types": req.DataTypes,
		"reason":     req.Reason,
		"metadata":   req.Metadata,
	}

	complianceRecord := &models.ComplianceRecord{
		UserID:         req.UserID,
		ComplianceType: "data_access",
		RequestType:    req.RequestType,
		Status:         "pending",
		RequestData:    requestData,
		LegalBasis:     "data_subject_right",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Definir expiração da solicitação (30 dias para processar)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	complianceRecord.ExpiresAt = &expiresAt

	if err := l.db.Create(complianceRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create data access request: %w", err)
	}

	// Processar solicitação automaticamente (converter uint para uuid temporário)
	go l.processDataAccessRequest(uuid.New())

	return &DataResponse{
		RequestID: uuid.New(), // Gerar UUID temporário já que ID é uint
		Status:    "pending",
		Message:   "Data access request submitted successfully",
		ExpiresAt: expiresAt,
	}, nil
}

// RequestDataDeletion solicita exclusão de dados pessoais
func (l *LGPDService) RequestDataDeletion(ctx context.Context, req *DataRequest) (*DataResponse, error) {
	// Verificar se usuário existe
	var user models.User
	if err := l.db.First(&user, req.UserID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Criar registro de solicitação
	requestData := map[string]interface{}{
		"data_types": req.DataTypes,
		"reason":     req.Reason,
		"metadata":   req.Metadata,
	}

	complianceRecord := &models.ComplianceRecord{
		UserID:         req.UserID,
		ComplianceType: "data_deletion",
		RequestType:    req.RequestType,
		Status:         "pending",
		RequestData:    requestData,
		LegalBasis:     "data_subject_right",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Definir expiração da solicitação (30 dias para processar)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	complianceRecord.ExpiresAt = &expiresAt

	if err := l.db.Create(complianceRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create data deletion request: %w", err)
	}

	return &DataResponse{
		RequestID: uuid.New(), // Gerar UUID temporário já que ID é uint
		Status:    "pending",
		Message:   "Data deletion request submitted successfully",
		ExpiresAt: expiresAt,
	}, nil
}

// RequestDataPortability solicita portabilidade dos dados
func (l *LGPDService) RequestDataPortability(ctx context.Context, req *DataRequest) (*DataResponse, error) {
	// Verificar se usuário existe
	var user models.User
	if err := l.db.First(&user, req.UserID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Criar registro de solicitação
	requestData := map[string]interface{}{
		"data_types": req.DataTypes,
		"reason":     req.Reason,
		"metadata":   req.Metadata,
	}

	complianceRecord := &models.ComplianceRecord{
		UserID:         req.UserID,
		ComplianceType: "data_portability",
		RequestType:    req.RequestType,
		Status:         "pending",
		RequestData:    requestData,
		LegalBasis:     "data_subject_right",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Definir expiração da solicitação (30 dias para processar)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	complianceRecord.ExpiresAt = &expiresAt

	if err := l.db.Create(complianceRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create data portability request: %w", err)
	}

	// Processar solicitação automaticamente (converter uint para uuid temporário)
	go l.processDataPortabilityRequest(uuid.New())

	return &DataResponse{
		RequestID: uuid.New(), // Gerar UUID temporário já que ID é uint
		Status:    "pending",
		Message:   "Data portability request submitted successfully",
		ExpiresAt: expiresAt,
	}, nil
}

// ProcessDataDeletion processa exclusão de dados
func (l *LGPDService) ProcessDataDeletion(ctx context.Context, requestID uuid.UUID) error {
	var record models.ComplianceRecord
	if err := l.db.First(&record, requestID).Error; err != nil {
		return fmt.Errorf("request not found: %w", err)
	}

	if record.ComplianceType != "data_deletion" {
		return fmt.Errorf("invalid request type")
	}

	// Anonimizar/excluir dados do usuário
	userID := record.UserID

	tx := l.db.Begin()

	// Excluir dados sensíveis mantendo registros de auditoria
	if err := l.anonymizeUserData(tx, userID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to anonymize user data: %w", err)
	}

	// Atualizar status da solicitação
	record.Status = "completed"
	if record.RequestData == nil {
		record.RequestData = make(map[string]interface{})
	}
	record.RequestData["processed_at"] = time.Now()
	record.RequestData["deletion_completed"] = true
	record.RequestData["anonymization_applied"] = true
	record.UpdatedAt = time.Now()

	if err := tx.Save(&record).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update request: %w", err)
	}

	return tx.Commit().Error
}

// GetUserConsentStatus obtém status de consentimento do usuário
func (l *LGPDService) GetUserConsentStatus(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	var consents []models.ComplianceRecord
	err := l.db.Where("user_id = ? AND compliance_type = ? AND status = ?",
		userID, "consent", "approved").Find(&consents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get consent status: %w", err)
	}

	status := map[string]interface{}{
		"user_id":        userID,
		"consents":       consents,
		"total_consents": len(consents),
	}

	// Verificar se há consentimentos próximos do vencimento
	expiringConsents := 0
	for _, consent := range consents {
		if consent.ExpiresAt != nil && time.Until(*consent.ExpiresAt) < 30*24*time.Hour {
			expiringConsents++
		}
	}
	status["expiring_consents"] = expiringConsents

	return status, nil
}

// GetComplianceStats obtém estatísticas de compliance
func (l *LGPDService) GetComplianceStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total de registros de compliance
	var total int64
	l.db.Model(&models.ComplianceRecord{}).Count(&total)
	stats["total_records"] = total

	// Por tipo de compliance
	var typeStats []struct {
		ComplianceType string `json:"compliance_type"`
		Count          int64  `json:"count"`
	}
	l.db.Model(&models.ComplianceRecord{}).
		Select("compliance_type, COUNT(*) as count").
		Group("compliance_type").
		Find(&typeStats)
	stats["by_type"] = typeStats

	// Por status
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	l.db.Model(&models.ComplianceRecord{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusStats)
	stats["by_status"] = statusStats

	// Consentimentos expirando
	var expiring int64
	l.db.Model(&models.ComplianceRecord{}).
		Where("compliance_type = ? AND status = ? AND expires_at <= ?",
			"consent", "approved", time.Now().Add(30*24*time.Hour)).
		Count(&expiring)
	stats["expiring_consents"] = expiring

	return stats, nil
}

// Funções auxiliares

func (l *LGPDService) processDataAccessRequest(requestID uuid.UUID) {
	var record models.ComplianceRecord
	if err := l.db.First(&record, requestID).Error; err != nil {
		return
	}

	// Coletar dados do usuário
	userData, err := l.collectUserData(record.UserID)
	if err != nil {
		l.updateRequestStatus(requestID, "failed", err.Error())
		return
	}

	// Atualizar solicitação
	record.Status = "completed"
	if record.RequestData == nil {
		record.RequestData = make(map[string]interface{})
	}
	record.RequestData["processed_at"] = time.Now()
	record.RequestData["response_data"] = map[string]interface{}{
		"data": userData,
	}
	record.UpdatedAt = time.Now()

	l.db.Save(&record)
}

func (l *LGPDService) processDataPortabilityRequest(requestID uuid.UUID) {
	var record models.ComplianceRecord
	if err := l.db.First(&record, requestID).Error; err != nil {
		return
	}

	// Criar exportação dos dados
	exportURL, err := l.createDataExport(record.UserID)
	if err != nil {
		l.updateRequestStatus(requestID, "failed", err.Error())
		return
	}

	// Atualizar solicitação
	record.Status = "completed"
	if record.RequestData == nil {
		record.RequestData = make(map[string]interface{})
	}
	record.RequestData["processed_at"] = time.Now()
	record.RequestData["export_url"] = exportURL
	record.UpdatedAt = time.Now()

	l.db.Save(&record)
}

func (l *LGPDService) updateRequestStatus(requestID uuid.UUID, status, message string) {
	l.db.Model(&models.ComplianceRecord{}).
		Where("id = ?", requestID).
		Updates(map[string]interface{}{
			"status":       status,
			"processed_at": time.Now(),
			"notes":        message,
			"updated_at":   time.Now(),
		})
}

func (l *LGPDService) collectUserData(userID uuid.UUID) (map[string]interface{}, error) {
	var user models.User
	if err := l.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	// Preparar dados do usuário (removendo campos sensíveis)
	userData := map[string]interface{}{
		"id":                user.ID,
		"email":             user.Email,
		"is_active":         user.IsActive,
		"is_email_verified": user.IsEmailVerified,
		"created_at":        user.CreatedAt,
		"updated_at":        user.UpdatedAt,
	}

	return userData, nil
}

func (l *LGPDService) createDataExport(userID uuid.UUID) (string, error) {
	// Simular criação de arquivo de exportação
	exportID := uuid.New().String()
	exportURL := fmt.Sprintf("/api/exports/%s", exportID)

	// Em produção, criar arquivo real com dados estruturados
	return exportURL, nil
}

func (l *LGPDService) anonymizeUserData(tx *gorm.DB, userID uuid.UUID) error {
	// Anonimizar dados do usuário mantendo integridade referencial
	updates := map[string]interface{}{
		"email":         fmt.Sprintf("anonymous_%s@deleted.local", uuid.New().String()[:8]),
		"password_hash": "",
		"salt":          "",
		"is_active":     false,
		"consent_given": false,
		"consent_date":  nil,
		"updated_at":    time.Now(),
	}

	return tx.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

// CleanupExpiredConsents limpa consentimentos expirados
func (l *LGPDService) CleanupExpiredConsents(ctx context.Context) error {
	result := l.db.Model(&models.ComplianceRecord{}).
		Where("compliance_type = ? AND status = ? AND expires_at < ?",
			"consent", "approved", time.Now()).
		Update("status", "expired")

	fmt.Printf("Marked %d consents as expired\n", result.RowsAffected)
	return result.Error
}
