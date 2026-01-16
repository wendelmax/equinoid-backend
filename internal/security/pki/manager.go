package pki

import (
	"context"
	"fmt"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PKIManager gerencia todos os aspectos de PKI
type PKIManager struct {
	db        *gorm.DB
	caService *CAService
}

// NewPKIManager cria um novo gerenciador PKI
func NewPKIManager(db *gorm.DB, caService *CAService) *PKIManager {
	return &PKIManager{
		db:        db,
		caService: caService,
	}
}

// RequestCertificate solicita um novo certificado para um usuário
func (p *PKIManager) RequestCertificate(ctx context.Context, req *CertificateRequest) (*models.Certificate, error) {
	// Verificar se usuário existe
	var user models.User
	if err := p.db.First(&user, req.UserID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verificar se usuário já tem certificado válido
	var existingCert models.Certificate
	err := p.db.Where("user_id = ? AND is_revoked = ? AND expires_at > ?",
		req.UserID, false, time.Now()).First(&existingCert).Error

	if err == nil {
		return nil, fmt.Errorf("user already has a valid certificate")
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing certificates: %w", err)
	}

	// Emitir novo certificado
	certResponse, err := p.caService.IssueCertificate(req)
	if err != nil {
		return nil, fmt.Errorf("failed to issue certificate: %w", err)
	}

	// Salvar no banco de dados
	if err := p.db.Create(certResponse.Certificate).Error; err != nil {
		return nil, fmt.Errorf("failed to save certificate: %w", err)
	}

	return certResponse.Certificate, nil
}

// GetUserCertificates retorna todos os certificados de um usuário
func (p *PKIManager) GetUserCertificates(ctx context.Context, userID uuid.UUID) ([]models.Certificate, error) {
	var certificates []models.Certificate
	err := p.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&certificates).Error

	return certificates, err
}

// GetActiveCertificate retorna o certificado ativo de um usuário
func (p *PKIManager) GetActiveCertificate(ctx context.Context, userID uuid.UUID) (*models.Certificate, error) {
	var certificate models.Certificate
	err := p.db.Where("user_id = ? AND is_revoked = ? AND expires_at > ?",
		userID, false, time.Now()).
		Order("expires_at DESC").
		First(&certificate).Error

	if err != nil {
		return nil, fmt.Errorf("no active certificate found: %w", err)
	}

	return &certificate, nil
}

// RevokeCertificate revoga um certificado
func (p *PKIManager) RevokeCertificate(ctx context.Context, userID, certificateID uuid.UUID, reason string) error {
	var certificate models.Certificate
	err := p.db.Where("id = ? AND user_id = ?", certificateID, userID).
		First(&certificate).Error

	if err != nil {
		return fmt.Errorf("certificate not found: %w", err)
	}

	if certificate.IsRevoked {
		return fmt.Errorf("certificate already revoked")
	}

	// Revogar usando CA service
	if err := p.caService.RevokeCertificate(&certificate, reason); err != nil {
		return fmt.Errorf("failed to revoke certificate: %w", err)
	}

	// Atualizar no banco
	if err := p.db.Save(&certificate).Error; err != nil {
		return fmt.Errorf("failed to update certificate: %w", err)
	}

	return nil
}

// VerifyCertificate verifica a validade de um certificado
func (p *PKIManager) VerifyCertificate(ctx context.Context, certificateID uuid.UUID) (*CertificateInfo, error) {
	var certificate models.Certificate
	if err := p.db.First(&certificate, certificateID).Error; err != nil {
		return nil, fmt.Errorf("certificate not found: %w", err)
	}

	// Verificar usando CA service
	info, err := p.caService.VerifyCertificate([]byte(certificate.CertificatePEM))
	if err != nil {
		return nil, fmt.Errorf("certificate verification failed: %w", err)
	}

	// Verificar se não foi revogado
	if certificate.IsRevoked {
		info.IsValid = false
		return info, fmt.Errorf("certificate has been revoked")
	}

	return info, nil
}

// RenewCertificate renova um certificado próximo do vencimento
func (p *PKIManager) RenewCertificate(ctx context.Context, userID, certificateID uuid.UUID) (*models.Certificate, error) {
	var oldCertificate models.Certificate
	err := p.db.Where("id = ? AND user_id = ?", certificateID, userID).
		First(&oldCertificate).Error

	if err != nil {
		return nil, fmt.Errorf("certificate not found: %w", err)
	}

	// Verificar se está próximo do vencimento (30 dias)
	if time.Until(oldCertificate.ExpiresAt) > 30*24*time.Hour {
		return nil, fmt.Errorf("certificate not yet eligible for renewal")
	}

	// Buscar dados do usuário
	var user models.User
	if err := p.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	req := &CertificateRequest{
		UserID:       user.ID,
		CommonName:   oldCertificate.CommonName,
		Organization: "Agent4 Security",
		Country:      "BR",
		Province:     "SP",
		Locality:     "São Paulo",
		EmailAddress: user.Email,
		KeySize:      2048,
		ValidityDays: 365,
		KeyUsage:     []string{"digital_signature", "key_encipherment"},
	}

	// Emitir novo certificado
	certResponse, err := p.caService.IssueCertificate(req)
	if err != nil {
		return nil, fmt.Errorf("failed to renew certificate: %w", err)
	}

	// Salvar novo certificado
	if err := p.db.Create(certResponse.Certificate).Error; err != nil {
		return nil, fmt.Errorf("failed to save renewed certificate: %w", err)
	}

	// Revogar certificado antigo
	p.caService.RevokeCertificate(&oldCertificate, "superseded")
	p.db.Save(&oldCertificate)

	return certResponse.Certificate, nil
}

// ListExpiringCertificates lista certificados que vencerão em X dias
func (p *PKIManager) ListExpiringCertificates(ctx context.Context, days int) ([]models.Certificate, error) {
	expireDate := time.Now().Add(time.Duration(days) * 24 * time.Hour)

	var certificates []models.Certificate
	err := p.db.Where("is_revoked = ? AND expires_at <= ? AND expires_at > ?",
		false, expireDate, time.Now()).
		Preload("User").
		Find(&certificates).Error

	return certificates, err
}

// GetCertificateStats retorna estatísticas dos certificados
func (p *PKIManager) GetCertificateStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total de certificados
	var total int64
	p.db.Model(&models.Certificate{}).Count(&total)
	stats["total"] = total

	// Certificados ativos
	var active int64
	p.db.Model(&models.Certificate{}).
		Where("is_revoked = ? AND expires_at > ?", false, time.Now()).
		Count(&active)
	stats["active"] = active

	// Certificados revogados
	var revoked int64
	p.db.Model(&models.Certificate{}).Where("is_revoked = ?", true).Count(&revoked)
	stats["revoked"] = revoked

	// Certificados expirados
	var expired int64
	p.db.Model(&models.Certificate{}).
		Where("is_revoked = ? AND expires_at <= ?", false, time.Now()).
		Count(&expired)
	stats["expired"] = expired

	// Certificados expirando em 30 dias
	var expiringSoon int64
	expireDate := time.Now().Add(30 * 24 * time.Hour)
	p.db.Model(&models.Certificate{}).
		Where("is_revoked = ? AND expires_at <= ? AND expires_at > ?",
			false, expireDate, time.Now()).
		Count(&expiringSoon)
	stats["expiring_soon"] = expiringSoon

	return stats, nil
}

// ExportCertificate exporta um certificado para diferentes formatos
func (p *PKIManager) ExportCertificate(ctx context.Context, certificateID uuid.UUID, format string) ([]byte, error) {
	var certificate models.Certificate
	if err := p.db.First(&certificate, certificateID).Error; err != nil {
		return nil, fmt.Errorf("certificate not found: %w", err)
	}

	switch format {
	case "pem":
		return []byte(certificate.CertificatePEM), nil
	case "der":
		// Converter PEM para DER (implementação simplificada)
		return []byte(certificate.CertificatePEM), nil // TODO: implementar conversão real
	case "p12":
		// Exportar como PKCS#12 (implementação simplificada)
		return []byte(certificate.CertificatePEM), nil // TODO: implementar P12
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// ValidateCertificateChain valida a cadeia de um certificado
func (p *PKIManager) ValidateCertificateChain(ctx context.Context, certificateID uuid.UUID) error {
	var certificate models.Certificate
	if err := p.db.First(&certificate, certificateID).Error; err != nil {
		return fmt.Errorf("certificate not found: %w", err)
	}

	return p.caService.ValidateCertificateChain([]byte(certificate.CertificatePEM))
}

// ImportCertificate importa um certificado externo
func (p *PKIManager) ImportCertificate(ctx context.Context, userID uuid.UUID, certPEM []byte) (*models.Certificate, error) {
	// Verificar o certificado
	info, err := p.caService.VerifyCertificate(certPEM)
	if err != nil {
		return nil, fmt.Errorf("invalid certificate: %w", err)
	}

	// Buscar User.ID (uint) do usuário
	var user models.User
	if err := p.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	// Criar registro do certificado importado
	certificate := &models.Certificate{
		UserID:         user.ID,
		SerialNumber:   info.SerialNumber,
		CommonName:     info.Subject,
		CertificatePEM: string(certPEM),
		// PrivateKeyPEM não disponível para certificados importados
		IssuedAt:  info.NotBefore,
		ExpiresAt: info.NotAfter,
		IsRevoked: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Salvar no banco
	if err := p.db.Create(certificate).Error; err != nil {
		return nil, fmt.Errorf("failed to import certificate: %w", err)
	}

	return certificate, nil
}

// BackupCertificates faz backup de todos os certificados
func (p *PKIManager) BackupCertificates(ctx context.Context) ([]byte, error) {
	var certificates []models.Certificate
	if err := p.db.Find(&certificates).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve certificates: %w", err)
	}

	// Criar backup em formato JSON (simplificado)
	// Em produção, usar formato mais seguro e criptografado
	return nil, fmt.Errorf("backup functionality not implemented")
}

// RestoreCertificates restaura certificados de um backup
func (p *PKIManager) RestoreCertificates(ctx context.Context, backupData []byte) error {
	// Implementar restauração de backup
	return fmt.Errorf("restore functionality not implemented")
}
