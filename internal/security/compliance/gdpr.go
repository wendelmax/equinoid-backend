package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GDPRService gerencia compliance com o Regulamento Geral sobre a Proteção de Dados
type GDPRService struct {
	db               *gorm.DB
	dataRetention    time.Duration
	dpoContact       string // Data Protection Officer
	organizationName string
}

// NewGDPRService cria um novo serviço GDPR
func NewGDPRService(db *gorm.DB, dataRetention time.Duration, dpoContact, organizationName string) *GDPRService {
	return &GDPRService{
		db:               db,
		dataRetention:    dataRetention,
		dpoContact:       dpoContact,
		organizationName: organizationName,
	}
}

// LegalBasis representa bases legais do GDPR
type LegalBasis string

const (
	LegalBasisConsent             LegalBasis = "consent"
	LegalBasisContract            LegalBasis = "contract"
	LegalBasisLegalObligation     LegalBasis = "legal_obligation"
	LegalBasisVitalInterests      LegalBasis = "vital_interests"
	LegalBasisPublicTask          LegalBasis = "public_task"
	LegalBasisLegitimateInterests LegalBasis = "legitimate_interests"
)

// ProcessingPurpose representa finalidades de processamento
type ProcessingPurpose string

const (
	PurposeAuthentication     ProcessingPurpose = "authentication"
	PurposeDigitalSignature   ProcessingPurpose = "digital_signature"
	PurposeAuditLogging       ProcessingPurpose = "audit_logging"
	PurposeSecurityMonitoring ProcessingPurpose = "security_monitoring"
	PurposeLegalCompliance    ProcessingPurpose = "legal_compliance"
	PurposeServiceDelivery    ProcessingPurpose = "service_delivery"
)

// DataCategory representa categorias de dados pessoais
type DataCategory string

const (
	CategoryBasicData      DataCategory = "basic_data"     // nome, email
	CategoryBiometric      DataCategory = "biometric"      // dados biométricos
	CategoryAuthentication DataCategory = "authentication" // MFA, senhas
	CategoryLocation       DataCategory = "location"       // localização
	CategoryBehavioral     DataCategory = "behavioral"     // logs de acesso
	CategoryTechnical      DataCategory = "technical"      // IP, user agent
)

// GDPRRequest representa solicitação GDPR
type GDPRRequest struct {
	UserID         uuid.UUID              `json:"user_id"`
	RequestType    string                 `json:"request_type"` // access, rectification, erasure, portability, restriction, objection
	LegalBasis     LegalBasis             `json:"legal_basis"`
	Purpose        ProcessingPurpose      `json:"purpose"`
	DataCategories []DataCategory         `json:"data_categories"`
	Justification  string                 `json:"justification"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// GDPRResponse representa resposta GDPR
type GDPRResponse struct {
	RequestID    uuid.UUID              `json:"request_id"`
	Status       string                 `json:"status"`
	Message      string                 `json:"message"`
	Data         map[string]interface{} `json:"data,omitempty"`
	ExportURL    string                 `json:"export_url,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	ResponseTime time.Duration          `json:"response_time"`
	DPOContact   string                 `json:"dpo_contact"`
}

// ProcessingActivity representa atividade de processamento
type ProcessingActivity struct {
	ID             uuid.UUID         `json:"id"`
	Name           string            `json:"name"`
	Purpose        ProcessingPurpose `json:"purpose"`
	LegalBasis     LegalBasis        `json:"legal_basis"`
	DataCategories []DataCategory    `json:"data_categories"`
	Retention      time.Duration     `json:"retention"`
	Recipients     []string          `json:"recipients"`
	Transfers      []string          `json:"transfers"` // transferências internacionais
	Safeguards     []string          `json:"safeguards"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// RequestAccess processa solicitação de acesso (Art. 15)
func (g *GDPRService) RequestAccess(ctx context.Context, req *GDPRRequest) (*GDPRResponse, error) {
	startTime := time.Now()

	// Verificar se usuário existe
	var user models.User
	if err := g.db.First(&user, req.UserID).Error; err != nil {
		return nil, fmt.Errorf("data subject not found: %w", err)
	}

	// Criar registro de solicitação
	complianceRecord := &models.ComplianceRecord{
		UserID:         req.UserID,
		ComplianceType: "gdpr_access",
		RequestType:    "data_access",
		Status:         "processing",
		RequestData: map[string]interface{}{
			"legal_basis":     req.LegalBasis,
			"purpose":         req.Purpose,
			"data_categories": req.DataCategories,
			"justification":   req.Justification,
			"metadata":        req.Metadata,
		},
		LegalBasis: string(req.LegalBasis),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// GDPR Art. 12 - resposta em até 30 dias
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	complianceRecord.ExpiresAt = &expiresAt

	if err := g.db.Create(complianceRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create GDPR access request: %w", err)
	}

	// Processar solicitação
	data, err := g.collectPersonalData(req.UserID, req.DataCategories)
	if err != nil {
		return g.createErrorResponse(complianceRecord.ID, "failed to collect data", startTime), nil
	}

	// Adicionar informações sobre processamento
	data["processing_information"] = g.getProcessingInformation()
	data["data_sources"] = g.getDataSources()
	data["retention_periods"] = g.getRetentionPeriods()
	data["recipients"] = g.getRecipients()
	data["transfers"] = g.getInternationalTransfers()

	// Atualizar registro
	complianceRecord.Status = "completed"
	// Armazenar dados de resposta no RequestData
	if complianceRecord.RequestData == nil {
		complianceRecord.RequestData = make(map[string]interface{})
	}
	complianceRecord.RequestData["response_data"] = data
	complianceRecord.RequestData["processed_at"] = time.Now()
	complianceRecord.UpdatedAt = time.Now()
	g.db.Save(complianceRecord)

	processedAt := time.Now()
	return &GDPRResponse{
		RequestID:    uuid.New(), // Gerar UUID temporário para resposta
		Status:       "completed",
		Message:      "Personal data access request completed successfully",
		Data:         data,
		CompletedAt:  &processedAt,
		ResponseTime: time.Since(startTime),
		DPOContact:   g.dpoContact,
	}, nil
}

// RequestRectification processa solicitação de retificação (Art. 16)
func (g *GDPRService) RequestRectification(ctx context.Context, req *GDPRRequest) (*GDPRResponse, error) {
	startTime := time.Now()

	// Verificar dados a serem corrigidos
	corrections := req.Metadata["corrections"]
	if corrections == nil {
		return g.createErrorResponse(uuid.Nil, "corrections data required", startTime), nil
	}

	// Criar registro de solicitação
	complianceRecord := &models.ComplianceRecord{
		UserID:         req.UserID,
		ComplianceType: "gdpr_rectification",
		RequestType:    "data_rectification",
		Status:         "processing",
		RequestData: map[string]interface{}{
			"corrections":     corrections,
			"justification":   req.Justification,
			"data_categories": req.DataCategories,
		},
		LegalBasis: string(req.LegalBasis),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	complianceRecord.ExpiresAt = &expiresAt

	if err := g.db.Create(complianceRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create rectification request: %w", err)
	}

	// Aplicar correções
	if err := g.applyDataCorrections(req.UserID, corrections); err != nil {
		return g.createErrorResponse(complianceRecord.ID, "failed to apply corrections", startTime), nil
	}

	// Atualizar registro
	complianceRecord.Status = "completed"
	if complianceRecord.RequestData == nil {
		complianceRecord.RequestData = make(map[string]interface{})
	}
	complianceRecord.RequestData["corrections_applied"] = corrections
	complianceRecord.RequestData["processed_at"] = time.Now()
	complianceRecord.UpdatedAt = time.Now()
	g.db.Save(complianceRecord)

	processedAt := time.Now()
	return &GDPRResponse{
		RequestID:    uuid.New(),
		Status:       "completed",
		Message:      "Data rectification completed successfully",
		CompletedAt:  &processedAt,
		ResponseTime: time.Since(startTime),
		DPOContact:   g.dpoContact,
	}, nil
}

// RequestErasure processa solicitação de apagamento (Art. 17 - Right to be Forgotten)
func (g *GDPRService) RequestErasure(ctx context.Context, req *GDPRRequest) (*GDPRResponse, error) {
	startTime := time.Now()

	// Verificar se apagamento é aplicável
	if !g.isErasureApplicable(req) {
		return &GDPRResponse{
			RequestID:    uuid.New(),
			Status:       "rejected",
			Message:      "Erasure not applicable due to legal obligations or legitimate interests",
			ResponseTime: time.Since(startTime),
			DPOContact:   g.dpoContact,
		}, nil
	}

	// Criar registro de solicitação
	complianceRecord := &models.ComplianceRecord{
		UserID:         req.UserID,
		ComplianceType: "gdpr_erasure",
		RequestType:    "data_erasure",
		Status:         "processing",
		RequestData: map[string]interface{}{
			"justification":   req.Justification,
			"data_categories": req.DataCategories,
			"erasure_grounds": req.Metadata["erasure_grounds"],
		},
		LegalBasis: string(req.LegalBasis),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	complianceRecord.ExpiresAt = &expiresAt

	if err := g.db.Create(complianceRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create erasure request: %w", err)
	}

	// Processar apagamento
	erasureReport, err := g.processDataErasure(req.UserID, req.DataCategories)
	if err != nil {
		return g.createErrorResponse(complianceRecord.ID, "failed to process erasure", startTime), nil
	}

	// Atualizar registro
	complianceRecord.Status = "completed"
	if complianceRecord.RequestData == nil {
		complianceRecord.RequestData = make(map[string]interface{})
	}
	complianceRecord.RequestData["erasure_report"] = erasureReport
	complianceRecord.RequestData["processed_at"] = time.Now()
	complianceRecord.UpdatedAt = time.Now()
	g.db.Save(complianceRecord)

	processedAt := time.Now()
	return &GDPRResponse{
		RequestID:    uuid.New(),
		Status:       "completed",
		Message:      "Data erasure completed successfully",
		Data:         erasureReport,
		CompletedAt:  &processedAt,
		ResponseTime: time.Since(startTime),
		DPOContact:   g.dpoContact,
	}, nil
}

// RequestPortability processa solicitação de portabilidade (Art. 20)
func (g *GDPRService) RequestPortability(ctx context.Context, req *GDPRRequest) (*GDPRResponse, error) {
	startTime := time.Now()

	// Verificar se portabilidade é aplicável (apenas dados fornecidos pelo usuário)
	if !g.isPortabilityApplicable(req.LegalBasis) {
		return &GDPRResponse{
			RequestID:    uuid.New(),
			Status:       "rejected",
			Message:      "Portability only applies to data processed based on consent or contract",
			ResponseTime: time.Since(startTime),
			DPOContact:   g.dpoContact,
		}, nil
	}

	// Criar registro de solicitação
	complianceRecord := &models.ComplianceRecord{
		UserID:         req.UserID,
		ComplianceType: "gdpr_portability",
		RequestType:    "data_portability",
		Status:         "processing",
		RequestData: map[string]interface{}{
			"data_categories": req.DataCategories,
			"format":          req.Metadata["format"],   // JSON, CSV, XML
			"delivery":        req.Metadata["delivery"], // download, email, api
		},
		LegalBasis: string(req.LegalBasis),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	complianceRecord.ExpiresAt = &expiresAt

	if err := g.db.Create(complianceRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create portability request: %w", err)
	}

	// Criar exportação em formato estruturado
	exportURL, err := g.createStructuredExport(req.UserID, req.DataCategories, req.Metadata)
	if err != nil {
		return g.createErrorResponse(complianceRecord.ID, "failed to create export", startTime), nil
	}

	// Atualizar registro
	complianceRecord.Status = "completed"
	if complianceRecord.RequestData == nil {
		complianceRecord.RequestData = make(map[string]interface{})
	}
	complianceRecord.RequestData["export_url"] = exportURL
	complianceRecord.RequestData["format"] = req.Metadata["format"]
	complianceRecord.RequestData["data_exported"] = req.DataCategories
	complianceRecord.RequestData["processed_at"] = time.Now()
	complianceRecord.UpdatedAt = time.Now()
	g.db.Save(complianceRecord)

	processedAt := time.Now()
	return &GDPRResponse{
		RequestID:    uuid.New(),
		Status:       "completed",
		Message:      "Data portability export completed successfully",
		ExportURL:    exportURL,
		CompletedAt:  &processedAt,
		ResponseTime: time.Since(startTime),
		DPOContact:   g.dpoContact,
	}, nil
}

// CreateDataBreachNotification cria notificação de violação de dados (Art. 33/34)
func (g *GDPRService) CreateDataBreachNotification(ctx context.Context, breach *DataBreach) error {
	// GDPR Art. 33 - notificar autoridade em até 72 horas
	notificationDeadline := breach.DetectedAt.Add(72 * time.Hour)

	notification := &models.ComplianceRecord{
		ComplianceType: "gdpr_breach",
		RequestType:    "data_breach_notification",
		Status:         "pending",
		RequestData: map[string]interface{}{
			"breach_type":           breach.Type,
			"affected_records":      breach.AffectedRecords,
			"data_categories":       breach.DataCategories,
			"detected_at":           breach.DetectedAt,
			"contained_at":          breach.ContainedAt,
			"notification_deadline": notificationDeadline,
			"severity":              breach.Severity,
			"description":           breach.Description,
			"measures_taken":        breach.MeasuresTaken,
			"likely_consequences":   breach.LikelyConsequences,
		},
		ExpiresAt: &notificationDeadline,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := g.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create breach notification: %w", err)
	}

	// Se alto risco, notificar indivíduos afetados (Art. 34)
	if breach.Severity == "high" {
		go g.notifyAffectedIndividuals(breach)
	}

	return nil
}

// DataBreach representa uma violação de dados
type DataBreach struct {
	Type               string         `json:"type"` // confidentiality, integrity, availability
	AffectedRecords    int            `json:"affected_records"`
	DataCategories     []DataCategory `json:"data_categories"`
	DetectedAt         time.Time      `json:"detected_at"`
	ContainedAt        *time.Time     `json:"contained_at"`
	Severity           string         `json:"severity"` // low, medium, high
	Description        string         `json:"description"`
	MeasuresTaken      []string       `json:"measures_taken"`
	LikelyConsequences []string       `json:"likely_consequences"`
}

// Funções auxiliares

func (g *GDPRService) collectPersonalData(userID uuid.UUID, categories []DataCategory) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Dados básicos do usuário
	var user models.User
	if err := g.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	for _, category := range categories {
		switch category {
		case CategoryBasicData:
			data["basic_data"] = map[string]interface{}{
				"id":                user.ID,
				"email":             user.Email,
				"is_active":         user.IsActive,
				"is_email_verified": user.IsEmailVerified,
				"created_at":        user.CreatedAt,
				"updated_at":        user.UpdatedAt,
			}

		case CategoryBiometric:
			var biometricData []models.BiometricData
			g.db.Where("user_id = ?", userID).Find(&biometricData)
			data["biometric_data"] = len(biometricData) // Não expor dados biométricos reais

		case CategoryAuthentication:
			var mfaDevices []models.MFADevice
			g.db.Where("user_id = ?", userID).Find(&mfaDevices)
			data["mfa_devices"] = len(mfaDevices)

		case CategoryBehavioral:
			var auditLogs []models.AuditLog
			g.db.Where("user_id = ?", userID).Limit(100).Find(&auditLogs) // Limitar para evitar excesso
			data["recent_activities"] = len(auditLogs)

		case CategoryTechnical:
			var sessions []models.Session
			g.db.Where("user_id = ?", userID).Find(&sessions)
			data["sessions"] = len(sessions)
		}
	}

	return data, nil
}

func (g *GDPRService) getProcessingInformation() map[string]interface{} {
	return map[string]interface{}{
		"controller":  g.organizationName,
		"dpo_contact": g.dpoContact,
		"purposes": []string{
			"Authentication and access control",
			"Digital signature services",
			"Security monitoring and audit",
			"Legal compliance",
		},
		"legal_bases": []string{
			"Consent (Art. 6(1)(a))",
			"Contract (Art. 6(1)(b))",
			"Legal obligation (Art. 6(1)(c))",
			"Legitimate interests (Art. 6(1)(f))",
		},
	}
}

func (g *GDPRService) getDataSources() []string {
	return []string{
		"User registration",
		"Biometric enrollment",
		"System logs",
		"Third-party identity providers",
	}
}

func (g *GDPRService) getRetentionPeriods() map[string]interface{} {
	return map[string]interface{}{
		"user_data":      "Duration of service + 7 years for legal compliance",
		"biometric_data": "Duration of service",
		"audit_logs":     "7 years for legal and regulatory compliance",
		"session_data":   "90 days after session expiry",
	}
}

func (g *GDPRService) getRecipients() []string {
	return []string{
		"Internal security team",
		"Legal and compliance team",
		"Authorized service providers (under DPA)",
	}
}

func (g *GDPRService) getInternationalTransfers() map[string]interface{} {
	return map[string]interface{}{
		"transfers": []string{
			"EU/EEA cloud providers (adequacy decision)",
		},
		"safeguards": []string{
			"Standard Contractual Clauses",
			"Encryption in transit and at rest",
		},
	}
}

func (g *GDPRService) isErasureApplicable(req *GDPRRequest) bool {
	// Verificar se há obrigações legais que impedem o apagamento
	legalObligations := []string{"audit_logs", "legal_compliance", "financial_records"}

	for _, category := range req.DataCategories {
		for _, obligation := range legalObligations {
			if string(category) == obligation {
				return false // Não pode apagar devido a obrigação legal
			}
		}
	}

	return true
}

func (g *GDPRService) isPortabilityApplicable(legalBasis LegalBasis) bool {
	// Portabilidade só se aplica a dados processados com base em consentimento ou contrato
	return legalBasis == LegalBasisConsent || legalBasis == LegalBasisContract
}

func (g *GDPRService) processDataErasure(userID uuid.UUID, categories []DataCategory) (map[string]interface{}, error) {
	report := make(map[string]interface{})

	for _, category := range categories {
		switch category {
		case CategoryBasicData:
			// Pseudonimizar dados básicos
			err := g.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
				"email":      fmt.Sprintf("erased_%s@deleted.local", uuid.New().String()[:8]),
				"is_active":  false,
				"updated_at": time.Now(),
			}).Error
			report[string(category)] = map[string]interface{}{
				"status": "pseudonymized",
				"error":  err,
			}

		case CategoryBiometric:
			// Apagar dados biométricos
			result := g.db.Where("user_id = ?", userID).Delete(&models.BiometricData{})
			report[string(category)] = map[string]interface{}{
				"status":  "deleted",
				"records": result.RowsAffected,
				"error":   result.Error,
			}

		case CategoryAuthentication:
			// Apagar dispositivos MFA
			result := g.db.Where("user_id = ?", userID).Delete(&models.MFADevice{})
			report[string(category)] = map[string]interface{}{
				"status":  "deleted",
				"records": result.RowsAffected,
				"error":   result.Error,
			}
		}
	}

	return report, nil
}

func (g *GDPRService) applyDataCorrections(userID uuid.UUID, corrections interface{}) error {
	// Aplicar correções solicitadas pelo usuário
	if corrMap, ok := corrections.(map[string]interface{}); ok {
		return g.db.Model(&models.User{}).Where("id = ?", userID).Updates(corrMap).Error
	}
	return fmt.Errorf("invalid corrections format")
}

func (g *GDPRService) createStructuredExport(userID uuid.UUID, categories []DataCategory, metadata map[string]interface{}) (string, error) {
	// Criar exportação estruturada dos dados
	exportID := uuid.New().String()
	exportURL := fmt.Sprintf("/api/gdpr/exports/%s", exportID)

	// Em produção, gerar arquivo real em formato estruturado
	return exportURL, nil
}

func (g *GDPRService) notifyAffectedIndividuals(breach *DataBreach) {
	// Implementar notificação aos indivíduos afetados
	fmt.Printf("Notifying affected individuals about data breach: %s\n", breach.Description)
}

func (g *GDPRService) createErrorResponse(requestID interface{}, message string, startTime time.Time) *GDPRResponse {
	var requestUUID uuid.UUID
	switch v := requestID.(type) {
	case uuid.UUID:
		requestUUID = v
	case uint:
		// Converter uint para UUID temporário (não ideal, mas necessário para compatibilidade)
		requestUUID = uuid.New()
	default:
		requestUUID = uuid.Nil
	}
	return &GDPRResponse{
		RequestID:    requestUUID,
		Status:       "error",
		Message:      message,
		ResponseTime: time.Since(startTime),
		DPOContact:   g.dpoContact,
	}
}
