package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLogger gerencia logs de auditoria
type AuditLogger struct {
	db            *gorm.DB
	enabledEvents map[string]bool
	riskLevels    map[string]string
	anonymize     bool
	retention     time.Duration
}

// NewAuditLogger cria um novo logger de auditoria
func NewAuditLogger(db *gorm.DB, retention time.Duration) *AuditLogger {
	logger := &AuditLogger{
		db:            db,
		enabledEvents: make(map[string]bool),
		riskLevels:    make(map[string]string),
		anonymize:     true,
		retention:     retention,
	}

	// Configurar eventos padrão
	logger.setupDefaultEvents()
	logger.setupRiskLevels()

	return logger
}

// AuditEvent representa um evento de auditoria
type AuditEvent struct {
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID *uuid.UUID             `json:"resource_id,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Location   string                 `json:"location,omitempty"`
	Success    bool                   `json:"success"`
	ErrorMsg   string                 `json:"error_msg,omitempty"`
	RiskLevel  string                 `json:"risk_level,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// LogEvent registra um evento de auditoria
func (a *AuditLogger) LogEvent(ctx context.Context, event *AuditEvent) error {
	// Verificar se evento está habilitado
	if !a.isEventEnabled(event.Action) {
		return nil // Evento desabilitado, não registrar
	}

	// Determinar nível de risco se não fornecido
	if event.RiskLevel == "" {
		event.RiskLevel = a.getRiskLevel(event.Action)
	}

	// Anonimizar dados se necessário
	if a.anonymize {
		event = a.anonymizeEvent(event)
	}

	// Enriquecer evento com contexto
	event = a.enrichEvent(ctx, event)

	// Criar registro de auditoria
	auditLog := &models.AuditLog{
		UserID:     event.UserID,
		Action:     event.Action,
		Resource:   event.Resource,
		ResourceID: event.ResourceID,
		Details:    event.Details,
		IPAddress:  event.IPAddress,
		UserAgent:  event.UserAgent,
		Location:   event.Location,
		Success:    event.Success,
		ErrorMsg:   event.ErrorMsg,
		RiskLevel:  event.RiskLevel,
		Timestamp:  event.Timestamp,
		CreatedAt:  time.Now(),
	}

	// Salvar no banco de dados
	if err := a.db.Create(auditLog).Error; err != nil {
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	return nil
}

// LogAuth registra evento de autenticação
func (a *AuditLogger) LogAuth(ctx context.Context, userID *uuid.UUID, action string, success bool, details map[string]interface{}) error {
	event := &AuditEvent{
		UserID:    userID,
		Action:    action,
		Resource:  "authentication",
		Details:   details,
		Success:   success,
		RiskLevel: a.getAuthRiskLevel(action, success),
		Timestamp: time.Now(),
	}

	if !success && details != nil {
		if errorMsg, ok := details["error"].(string); ok {
			event.ErrorMsg = errorMsg
		}
	}

	return a.LogEvent(ctx, event)
}

// LogMFA registra evento de MFA
func (a *AuditLogger) LogMFA(ctx context.Context, userID uuid.UUID, deviceType string, success bool, details map[string]interface{}) error {
	event := &AuditEvent{
		UserID:    &userID,
		Action:    "mfa_verification",
		Resource:  "mfa_device",
		Details:   details,
		Success:   success,
		RiskLevel: a.getMFARiskLevel(success),
		Timestamp: time.Now(),
	}

	if details == nil {
		event.Details = make(map[string]interface{})
	}
	event.Details["device_type"] = deviceType

	return a.LogEvent(ctx, event)
}

// LogBiometric registra evento biométrico
func (a *AuditLogger) LogBiometric(ctx context.Context, userID uuid.UUID, biometricType string, success bool, score float64) error {
	event := &AuditEvent{
		UserID:   &userID,
		Action:   "biometric_verification",
		Resource: "biometric_data",
		Details: map[string]interface{}{
			"biometric_type": biometricType,
			"score":          score,
		},
		Success:   success,
		RiskLevel: a.getBiometricRiskLevel(success, score),
		Timestamp: time.Now(),
	}

	return a.LogEvent(ctx, event)
}

// LogCertificate registra evento de certificado
func (a *AuditLogger) LogCertificate(ctx context.Context, userID uuid.UUID, action string, certificateID uuid.UUID, success bool) error {
	event := &AuditEvent{
		UserID:     &userID,
		Action:     action,
		Resource:   "certificate",
		ResourceID: &certificateID,
		Success:    success,
		RiskLevel:  a.getCertificateRiskLevel(action),
		Timestamp:  time.Now(),
	}

	return a.LogEvent(ctx, event)
}

// LogSignature registra evento de assinatura digital
func (a *AuditLogger) LogSignature(ctx context.Context, userID, documentID uuid.UUID, success bool, biometricScore float64) error {
	event := &AuditEvent{
		UserID:     &userID,
		Action:     "digital_signature",
		Resource:   "document",
		ResourceID: &documentID,
		Details: map[string]interface{}{
			"biometric_score": biometricScore,
		},
		Success:   success,
		RiskLevel: a.getSignatureRiskLevel(success, biometricScore),
		Timestamp: time.Now(),
	}

	return a.LogEvent(ctx, event)
}

// LogDataAccess registra acesso a dados (LGPD/GDPR)
func (a *AuditLogger) LogDataAccess(ctx context.Context, userID *uuid.UUID, dataType, purpose string, success bool) error {
	event := &AuditEvent{
		UserID:   userID,
		Action:   "data_access",
		Resource: dataType,
		Details: map[string]interface{}{
			"purpose": purpose,
		},
		Success:   success,
		RiskLevel: "medium",
		Timestamp: time.Now(),
	}

	return a.LogEvent(ctx, event)
}

// LogDataDeletion registra exclusão de dados (LGPD/GDPR)
func (a *AuditLogger) LogDataDeletion(ctx context.Context, userID uuid.UUID, dataType string, reason string, success bool) error {
	event := &AuditEvent{
		UserID:     &userID,
		Action:     "data_deletion",
		Resource:   dataType,
		ResourceID: &userID,
		Details: map[string]interface{}{
			"reason": reason,
		},
		Success:   success,
		RiskLevel: "high",
		Timestamp: time.Now(),
	}

	return a.LogEvent(ctx, event)
}

// LogComplianceRequest registra solicitação de compliance
func (a *AuditLogger) LogComplianceRequest(ctx context.Context, userID uuid.UUID, requestType string, status string) error {
	event := &AuditEvent{
		UserID:     &userID,
		Action:     "compliance_request",
		Resource:   "user_data",
		ResourceID: &userID,
		Details: map[string]interface{}{
			"request_type": requestType,
			"status":       status,
		},
		Success:   status == "completed",
		RiskLevel: "medium",
		Timestamp: time.Now(),
	}

	return a.LogEvent(ctx, event)
}

// LogSecurityEvent registra evento de segurança
func (a *AuditLogger) LogSecurityEvent(ctx context.Context, eventType, severity string, details map[string]interface{}) error {
	event := &AuditEvent{
		Action:    eventType,
		Resource:  "security_system",
		Details:   details,
		Success:   true,
		RiskLevel: severity,
		Timestamp: time.Now(),
	}

	return a.LogEvent(ctx, event)
}

// QueryLogs consulta logs de auditoria
func (a *AuditLogger) QueryLogs(ctx context.Context, filter *AuditFilter) ([]models.AuditLog, error) {
	query := a.db.Model(&models.AuditLog{})

	// Aplicar filtros
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}

	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}

	if filter.Resource != "" {
		query = query.Where("resource = ?", filter.Resource)
	}

	if filter.RiskLevel != "" {
		query = query.Where("risk_level = ?", filter.RiskLevel)
	}

	if !filter.StartDate.IsZero() {
		query = query.Where("timestamp >= ?", filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		query = query.Where("timestamp <= ?", filter.EndDate)
	}

	if filter.Success != nil {
		query = query.Where("success = ?", *filter.Success)
	}

	// Aplicar ordenação e limitação
	query = query.Order("timestamp DESC")
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	var logs []models.AuditLog
	err := query.Preload("User").Find(&logs).Error
	return logs, err
}

// AuditFilter representa filtros para consulta de logs
type AuditFilter struct {
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Action    string     `json:"action,omitempty"`
	Resource  string     `json:"resource,omitempty"`
	RiskLevel string     `json:"risk_level,omitempty"`
	StartDate time.Time  `json:"start_date,omitempty"`
	EndDate   time.Time  `json:"end_date,omitempty"`
	Success   *bool      `json:"success,omitempty"`
	Limit     int        `json:"limit,omitempty"`
}

// GetAuditStats obtém estatísticas de auditoria
func (a *AuditLogger) GetAuditStats(ctx context.Context, period time.Duration) (map[string]interface{}, error) {
	since := time.Now().Add(-period)
	stats := make(map[string]interface{})

	// Total de eventos
	var total int64
	a.db.Model(&models.AuditLog{}).Where("timestamp >= ?", since).Count(&total)
	stats["total_events"] = total

	// Eventos por ação
	var actionStats []struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	a.db.Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("timestamp >= ?", since).
		Group("action").
		Find(&actionStats)
	stats["by_action"] = actionStats

	// Eventos por nível de risco
	var riskStats []struct {
		RiskLevel string `json:"risk_level"`
		Count     int64  `json:"count"`
	}
	a.db.Model(&models.AuditLog{}).
		Select("risk_level, COUNT(*) as count").
		Where("timestamp >= ?", since).
		Group("risk_level").
		Find(&riskStats)
	stats["by_risk_level"] = riskStats

	// Eventos falharam
	var failed int64
	a.db.Model(&models.AuditLog{}).
		Where("timestamp >= ? AND success = ?", since, false).
		Count(&failed)
	stats["failed_events"] = failed

	// Top IPs
	var ipStats []struct {
		IPAddress string `json:"ip_address"`
		Count     int64  `json:"count"`
	}
	a.db.Model(&models.AuditLog{}).
		Select("ip_address, COUNT(*) as count").
		Where("timestamp >= ? AND ip_address != ''", since).
		Group("ip_address").
		Order("count DESC").
		Limit(10).
		Find(&ipStats)
	stats["top_ips"] = ipStats

	return stats, nil
}

// CleanupOldLogs limpa logs antigos baseado na retenção
func (a *AuditLogger) CleanupOldLogs(ctx context.Context) error {
	cutoff := time.Now().Add(-a.retention)

	result := a.db.Where("timestamp < ?", cutoff).Delete(&models.AuditLog{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old logs: %w", result.Error)
	}

	fmt.Printf("Cleaned up %d old audit logs\n", result.RowsAffected)
	return nil
}

// Funções auxiliares

func (a *AuditLogger) setupDefaultEvents() {
	events := []string{
		"login", "logout", "login_failed", "password_reset",
		"mfa_setup", "mfa_verification", "mfa_failed",
		"biometric_enrollment", "biometric_verification",
		"certificate_issued", "certificate_revoked",
		"digital_signature", "data_access", "data_deletion",
		"compliance_request", "security_event",
	}

	for _, event := range events {
		a.enabledEvents[event] = true
	}
}

func (a *AuditLogger) setupRiskLevels() {
	a.riskLevels["login"] = "low"
	a.riskLevels["logout"] = "low"
	a.riskLevels["login_failed"] = "medium"
	a.riskLevels["password_reset"] = "high"
	a.riskLevels["mfa_setup"] = "medium"
	a.riskLevels["mfa_verification"] = "low"
	a.riskLevels["mfa_failed"] = "high"
	a.riskLevels["biometric_enrollment"] = "medium"
	a.riskLevels["biometric_verification"] = "low"
	a.riskLevels["certificate_issued"] = "medium"
	a.riskLevels["certificate_revoked"] = "high"
	a.riskLevels["digital_signature"] = "medium"
	a.riskLevels["data_access"] = "low"
	a.riskLevels["data_deletion"] = "critical"
	a.riskLevels["compliance_request"] = "medium"
}

func (a *AuditLogger) isEventEnabled(action string) bool {
	return a.enabledEvents[action]
}

func (a *AuditLogger) getRiskLevel(action string) string {
	if level, exists := a.riskLevels[action]; exists {
		return level
	}
	return "medium" // Padrão
}

func (a *AuditLogger) getAuthRiskLevel(action string, success bool) string {
	if !success {
		return "high"
	}
	return a.getRiskLevel(action)
}

func (a *AuditLogger) getMFARiskLevel(success bool) string {
	if !success {
		return "critical"
	}
	return "low"
}

func (a *AuditLogger) getBiometricRiskLevel(success bool, score float64) string {
	if !success {
		return "high"
	}
	if score < 0.8 {
		return "medium"
	}
	return "low"
}

func (a *AuditLogger) getCertificateRiskLevel(action string) string {
	if strings.Contains(action, "revoked") {
		return "high"
	}
	return "medium"
}

func (a *AuditLogger) getSignatureRiskLevel(success bool, biometricScore float64) string {
	if !success {
		return "high"
	}
	if biometricScore < 0.8 {
		return "medium"
	}
	return "low"
}

func (a *AuditLogger) enrichEvent(ctx context.Context, event *AuditEvent) *AuditEvent {
	// Extrair IP e User-Agent do contexto HTTP se disponível
	if req, ok := ctx.Value("http_request").(*http.Request); ok {
		if event.IPAddress == "" {
			event.IPAddress = a.extractIP(req)
		}
		if event.UserAgent == "" {
			event.UserAgent = req.UserAgent()
		}
	}

	// Definir timestamp se não fornecido
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	return event
}

func (a *AuditLogger) extractIP(req *http.Request) string {
	// Verificar headers de proxy
	ip := req.Header.Get("X-Forwarded-For")
	if ip != "" {
		return strings.Split(ip, ",")[0]
	}

	ip = req.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Usar IP direto
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}

func (a *AuditLogger) anonymizeEvent(event *AuditEvent) *AuditEvent {
	// Anonimizar IP (manter apenas primeiros 3 octetos)
	if event.IPAddress != "" {
		parts := strings.Split(event.IPAddress, ".")
		if len(parts) == 4 {
			event.IPAddress = fmt.Sprintf("%s.%s.%s.xxx", parts[0], parts[1], parts[2])
		}
	}

	// Anonimizar User-Agent (remover versões específicas)
	if event.UserAgent != "" {
		event.UserAgent = a.anonymizeUserAgent(event.UserAgent)
	}

	return event
}

func (a *AuditLogger) anonymizeUserAgent(userAgent string) string {
	// Implementação básica - remover números de versão específicos
	patterns := []string{
		`\d+\.\d+\.\d+\.\d+`, // Versões com 4 números
		`\d+\.\d+\.\d+`,      // Versões com 3 números
		`\d+\.\d+`,           // Versões com 2 números
	}

	anonymized := userAgent
	for _, pattern := range patterns {
		// Em produção, usar regex real
		anonymized = strings.ReplaceAll(anonymized, pattern, "x.x")
	}

	return anonymized
}

// ExportLogs exporta logs para compliance
func (a *AuditLogger) ExportLogs(ctx context.Context, filter *AuditFilter, format string) ([]byte, error) {
	logs, err := a.QueryLogs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}

	switch format {
	case "json":
		return json.Marshal(logs)
	case "csv":
		return a.exportLogsCSV(logs), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (a *AuditLogger) exportLogsCSV(logs []models.AuditLog) []byte {
	csv := "Timestamp,User ID,Action,Resource,Success,Risk Level,IP Address\n"

	for _, log := range logs {
		userID := ""
		if log.UserID != nil {
			userID = log.UserID.String()
		}

		line := fmt.Sprintf("%s,%s,%s,%s,%t,%s,%s\n",
			log.Timestamp.Format(time.RFC3339),
			userID,
			log.Action,
			log.Resource,
			log.Success,
			log.RiskLevel,
			log.IPAddress,
		)
		csv += line
	}

	return []byte(csv)
}
