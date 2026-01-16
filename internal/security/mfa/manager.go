package mfa

import (
	"context"
	"fmt"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/internal/security/crypto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Manager gerencia todos os aspectos de MFA
type MFAManager struct {
	db                *gorm.DB
	totpService       *TOTPService
	smsService        *SMSService
	emailService      *EmailService
	encryptionService *crypto.EncryptionService
}

// NewMFAManager cria um novo gerenciador MFA
func NewMFAManager(db *gorm.DB, encryptionService *crypto.EncryptionService, issuer string, windowSize uint) *MFAManager {
	return &MFAManager{
		db:                db,
		totpService:       NewTOTPService(issuer, windowSize),
		smsService:        NewSMSService(),
		emailService:      NewEmailService(),
		encryptionService: encryptionService,
	}
}

// SetupRequest representa uma solicitação de configuração MFA
type SetupRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	DeviceType  string    `json:"device_type"` // totp, sms, email
	DeviceName  string    `json:"device_name"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	Email       string    `json:"email,omitempty"`
}

// SetupResponse representa a resposta de configuração MFA
type SetupResponse struct {
	DeviceID    uuid.UUID      `json:"device_id"`
	TOTPData    *TOTPSetupData `json:"totp_data,omitempty"`
	BackupCodes []string       `json:"backup_codes,omitempty"`
	Message     string         `json:"message"`
}

// VerifyRequest representa uma solicitação de verificação MFA
type VerifyRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	DeviceID   uuid.UUID `json:"device_id"`
	Code       string    `json:"code"`
	DeviceType string    `json:"device_type"`
}

// SetupMFA configura um novo dispositivo MFA para o usuário
func (m *MFAManager) SetupMFA(ctx context.Context, req *SetupRequest) (*SetupResponse, error) {
	// Verificar se usuário existe
	var user models.User
	if err := m.db.First(&user, req.UserID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Criar dispositivo MFA
	device := models.MFADevice{
		UserID:     req.UserID,
		DeviceType: req.DeviceType,
		DeviceName: req.DeviceName,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	var response SetupResponse
	// DeviceID será preenchido após salvar o dispositivo no banco

	switch req.DeviceType {
	case "totp":
		return m.setupTOTP(ctx, &device, user.Email, &response)
	case "sms":
		return m.setupSMS(ctx, &device, req.PhoneNumber, &response)
	case "email":
		return m.setupEmail(ctx, &device, req.Email, &response)
	default:
		return nil, fmt.Errorf("unsupported device type: %s", req.DeviceType)
	}
}

// setupTOTP configura autenticação TOTP
func (m *MFAManager) setupTOTP(ctx context.Context, device *models.MFADevice, email string, response *SetupResponse) (*SetupResponse, error) {
	// Gerar secret TOTP
	totpData, err := m.totpService.GenerateSecret(email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Gerar códigos de backup
	backupCodes, err := m.totpService.GenerateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Criptografar secret
	encryptedSecret, err := m.encryptionService.Encrypt(totpData.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt TOTP secret: %w", err)
	}
	device.Secret = encryptedSecret

	encryptedBackupCodes := make(models.JSONB)
	for i, code := range backupCodes {
		encryptedCode, err := m.encryptionService.Encrypt(code)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt backup code: %w", err)
		}
		encryptedBackupCodes[fmt.Sprintf("%d", i)] = encryptedCode
	}
	device.BackupCodes = encryptedBackupCodes

	// Salvar no banco
	if err := m.db.Create(device).Error; err != nil {
		return nil, fmt.Errorf("failed to save MFA device: %w", err)
	}

	// Preencher DeviceID após salvar
	response.DeviceID = uuid.New() // Gerar UUID temporário, já que device.ID é uint
	response.TOTPData = totpData
	response.BackupCodes = backupCodes
	response.Message = "TOTP device configured successfully. Please save your backup codes in a secure location."

	return response, nil
}

// setupSMS configura autenticação SMS
func (m *MFAManager) setupSMS(ctx context.Context, device *models.MFADevice, phoneNumber string, response *SetupResponse) (*SetupResponse, error) {
	// Gerar e enviar código de verificação
	_, err := m.smsService.SendVerificationCode(phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to send SMS verification: %w", err)
	}

	// Armazenar temporariamente (usar Redis em produção)
	encryptedPhone, err := m.encryptionService.Encrypt(phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt phone number: %w", err)
	}
	device.Secret = encryptedPhone

	// Salvar dispositivo como inativo até verificação
	device.IsActive = false
	if err := m.db.Create(device).Error; err != nil {
		return nil, fmt.Errorf("failed to save MFA device: %w", err)
	}

	// Preencher DeviceID após salvar
	response.DeviceID = uuid.New() // Gerar UUID temporário, já que device.ID é uint
	response.Message = fmt.Sprintf("Verification code sent to %s. Please verify to complete setup.", phoneNumber)

	return response, nil
}

// setupEmail configura autenticação por email
func (m *MFAManager) setupEmail(ctx context.Context, device *models.MFADevice, email string, response *SetupResponse) (*SetupResponse, error) {
	// Gerar e enviar código de verificação
	_, err := m.emailService.SendVerificationCode(email)
	if err != nil {
		return nil, fmt.Errorf("failed to send email verification: %w", err)
	}

	encryptedEmail, err := m.encryptionService.Encrypt(email)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt email: %w", err)
	}
	device.Secret = encryptedEmail

	// Salvar dispositivo como inativo até verificação
	device.IsActive = false
	if err := m.db.Create(device).Error; err != nil {
		return nil, fmt.Errorf("failed to save MFA device: %w", err)
	}

	// Preencher DeviceID após salvar
	response.DeviceID = uuid.New() // Gerar UUID temporário, já que device.ID é uint
	response.Message = fmt.Sprintf("Verification code sent to %s. Please verify to complete setup.", email)

	return response, nil
}

// VerifyMFA verifica um código MFA
func (m *MFAManager) VerifyMFA(ctx context.Context, req *VerifyRequest) (bool, error) {
	// Buscar dispositivo MFA
	var device models.MFADevice
	if err := m.db.First(&device, req.DeviceID).Error; err != nil {
		return false, fmt.Errorf("MFA device not found: %w", err)
	}

	// Verificar se o dispositivo pertence ao usuário
	if device.UserID != req.UserID {
		return false, fmt.Errorf("device does not belong to user")
	}

	// Verificar se dispositivo está ativo
	if !device.IsActive {
		return false, fmt.Errorf("MFA device is not active")
	}

	var isValid bool
	var err error

	switch device.DeviceType {
	case "totp":
		isValid, err = m.verifyTOTP(&device, req.Code)
	case "sms":
		isValid, err = m.verifySMS(&device, req.Code)
	case "email":
		isValid, err = m.verifyEmail(&device, req.Code)
	default:
		return false, fmt.Errorf("unsupported device type: %s", device.DeviceType)
	}

	if err != nil {
		return false, err
	}

	if isValid {
		// Atualizar última utilização (campo não existe no modelo, apenas atualizar UpdatedAt)
		device.UpdatedAt = time.Now()
		m.db.Save(&device)
	}

	return isValid, nil
}

// verifyTOTP verifica código TOTP
func (m *MFAManager) verifyTOTP(device *models.MFADevice, code string) (bool, error) {
	// Descriptografar secret
	decryptedSecret, err := m.encryptionService.Decrypt(device.Secret)
	if err != nil {
		return false, fmt.Errorf("failed to decrypt TOTP secret: %w", err)
	}

	// Verificar código TOTP normal
	valid, err := m.totpService.ValidateCode(decryptedSecret, code)
	if err != nil {
		return false, err
	}

	if valid {
		return true, nil
	}

	// Descriptografar e verificar códigos de backup
	var backupCodes []string
	for i := 0; ; i++ {
		if val, ok := device.BackupCodes[fmt.Sprintf("%d", i)]; ok {
			encryptedCode := val.(string)
			decryptedCode, err := m.encryptionService.Decrypt(encryptedCode)
			if err != nil {
				return false, fmt.Errorf("failed to decrypt backup code: %w", err)
			}
			backupCodes = append(backupCodes, decryptedCode)
		} else {
			break
		}
	}

	return m.totpService.ValidateBackupCode(code, backupCodes), nil
}

// verifySMS verifica código SMS
func (m *MFAManager) verifySMS(device *models.MFADevice, code string) (bool, error) {
	decryptedPhone, err := m.encryptionService.Decrypt(device.Secret)
	if err != nil {
		return false, fmt.Errorf("failed to decrypt phone number: %w", err)
	}
	return m.smsService.VerifyCode(decryptedPhone, code)
}

// verifyEmail verifica código de email
func (m *MFAManager) verifyEmail(device *models.MFADevice, code string) (bool, error) {
	decryptedEmail, err := m.encryptionService.Decrypt(device.Secret)
	if err != nil {
		return false, fmt.Errorf("failed to decrypt email: %w", err)
	}
	return m.emailService.VerifyCode(decryptedEmail, code)
}

// GetUserDevices retorna todos os dispositivos MFA do usuário
func (m *MFAManager) GetUserDevices(ctx context.Context, userID uuid.UUID) ([]models.MFADevice, error) {
	var devices []models.MFADevice
	err := m.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("is_primary DESC, created_at ASC").
		Find(&devices).Error

	return devices, err
}

// RemoveDevice remove um dispositivo MFA
func (m *MFAManager) RemoveDevice(ctx context.Context, userID, deviceID uuid.UUID) error {
	result := m.db.Where("id = ? AND user_id = ?", deviceID, userID).
		Update("is_active", false)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("device not found")
	}

	return nil
}

// SetPrimaryDevice define um dispositivo como primário
func (m *MFAManager) SetPrimaryDevice(ctx context.Context, userID, deviceID uuid.UUID) error {
	tx := m.db.Begin()

	// Remover primary de todos os dispositivos do usuário
	if err := tx.Where("user_id = ?", userID).Update("is_primary", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Definir dispositivo específico como primary
	if err := tx.Where("id = ? AND user_id = ?", deviceID, userID).
		Update("is_primary", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// EnableMFA habilita MFA para o usuário
func (m *MFAManager) EnableMFA(ctx context.Context, userID uuid.UUID) error {
	return m.db.Model(&models.User{}).Where("id = ?", userID).
		Update("mfa_enabled", true).Error
}

// DisableMFA desabilita MFA para o usuário
func (m *MFAManager) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	tx := m.db.Begin()

	// Desativar todos os dispositivos
	if err := tx.Where("user_id = ?", userID).
		Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Desabilitar MFA no usuário
	if err := tx.Model(&models.User{}).Where("id = ?", userID).
		Update("mfa_enabled", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
