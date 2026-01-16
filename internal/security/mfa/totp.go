package mfa

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

// TOTPService gerencia autenticação TOTP
type TOTPService struct {
	issuer     string
	windowSize uint
}

// NewTOTPService cria um novo serviço TOTP
func NewTOTPService(issuer string, windowSize uint) *TOTPService {
	return &TOTPService{
		issuer:     issuer,
		windowSize: windowSize,
	}
}

// TOTPSetupData contém os dados necessários para configurar TOTP
type TOTPSetupData struct {
	Secret    string `json:"secret"`
	QRCode    []byte `json:"qr_code"`
	URL       string `json:"url"`
	ManualKey string `json:"manual_key"`
}

// GenerateSecret gera um novo secret TOTP para o usuário
func (t *TOTPService) GenerateSecret(accountName string) (*TOTPSetupData, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      t.issuer,
		AccountName: accountName,
		SecretSize:  32,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Gerar QR Code
	qrCode, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	return &TOTPSetupData{
		Secret:    key.Secret(),
		QRCode:    qrCode,
		URL:       key.URL(),
		ManualKey: formatSecret(key.Secret()),
	}, nil
}

// ValidateCode valida um código TOTP
func (t *TOTPService) ValidateCode(secret, code string) (bool, error) {
	return totp.ValidateCustom(
		code,
		secret,
		time.Now().UTC(),
		totp.ValidateOpts{
			Period:    30,
			Skew:      t.windowSize,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)
}

// GenerateBackupCodes gera códigos de backup para o usuário
func (t *TOTPService) GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)

	for i := 0; i < count; i++ {
		code, err := generateBackupCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code %d: %w", i+1, err)
		}
		codes[i] = code
	}

	return codes, nil
}

// ValidateBackupCode valida um código de backup
func (t *TOTPService) ValidateBackupCode(providedCode string, validCodes []string) bool {
	for _, code := range validCodes {
		if code == providedCode {
			return true
		}
	}
	return false
}

// formatSecret formata o secret em grupos de 4 caracteres para facilitar entrada manual
func formatSecret(secret string) string {
	formatted := ""
	for i, char := range secret {
		if i > 0 && i%4 == 0 {
			formatted += " "
		}
		formatted += string(char)
	}
	return formatted
}

// generateBackupCode gera um código de backup de 8 caracteres
func generateBackupCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567" // Base32 sem 0, 1, 8, 9
	const length = 8

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	// Formatar como XXXX-XXXX para facilitar leitura
	code := string(bytes)
	return code[:4] + "-" + code[4:], nil
}

// GetCurrentCode gera o código TOTP atual (para testes)
func (t *TOTPService) GetCurrentCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now().UTC())
}

// GetTimeRemaining retorna o tempo restante até o próximo código
func (t *TOTPService) GetTimeRemaining() int {
	now := time.Now().UTC()
	period := 30 // TOTP period in seconds
	return period - (int(now.Unix()) % period)
}

// IsTimeValid verifica se ainda há tempo suficiente para usar o código
func (t *TOTPService) IsTimeValid(minSecondsRemaining int) bool {
	return t.GetTimeRemaining() >= minSecondsRemaining
}
