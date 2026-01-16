package mfa

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// EmailService gerencia autenticação por email
type EmailService struct {
	codes map[string]EmailCodeData // Em produção, usar Redis
}

// EmailCodeData representa dados de um código de verificação por email
type EmailCodeData struct {
	Code      string
	ExpiresAt time.Time
	Attempts  int
}

// NewEmailService cria um novo serviço de email
func NewEmailService() *EmailService {
	return &EmailService{
		codes: make(map[string]EmailCodeData),
	}
}

// SendVerificationCode envia um código de verificação por email
func (e *EmailService) SendVerificationCode(email string) (string, error) {
	// Gerar código alfanumérico de 8 caracteres
	code, err := generateAlphanumericCode(8)
	if err != nil {
		return "", fmt.Errorf("failed to generate email code: %w", err)
	}

	// Armazenar código com expiração de 10 minutos
	e.codes[email] = EmailCodeData{
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Attempts:  0,
	}

	// Enviar email
	err = e.sendEmail(email, code)
	if err != nil {
		delete(e.codes, email)
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	return code, nil
}

// VerifyCode verifica um código de email
func (e *EmailService) VerifyCode(email, providedCode string) (bool, error) {
	codeData, exists := e.codes[email]
	if !exists {
		return false, fmt.Errorf("no verification code found for this email")
	}

	// Verificar se código não expirou
	if time.Now().After(codeData.ExpiresAt) {
		delete(e.codes, email)
		return false, fmt.Errorf("verification code has expired")
	}

	// Incrementar tentativas
	codeData.Attempts++
	e.codes[email] = codeData

	// Verificar limite de tentativas
	if codeData.Attempts > 5 {
		delete(e.codes, email)
		return false, fmt.Errorf("too many attempts, please request a new code")
	}

	// Verificar código (case insensitive)
	if codeData.Code == providedCode {
		delete(e.codes, email) // Código usado, remover
		return true, nil
	}

	return false, nil
}

// ResendCode reenvia o código de verificação
func (e *EmailService) ResendCode(email string) (string, error) {
	// Verificar se existe código pendente
	if codeData, exists := e.codes[email]; exists {
		// Verificar se ainda não passou 2 minutos desde o último envio
		if time.Since(codeData.ExpiresAt.Add(-10*time.Minute)) < 2*time.Minute {
			return "", fmt.Errorf("please wait before requesting a new code")
		}
	}

	// Enviar novo código
	return e.SendVerificationCode(email)
}

// sendEmail simula o envio de email (implementar com serviço real)
func (e *EmailService) sendEmail(email, code string) error {
	// Simular delay de envio
	time.Sleep(200 * time.Millisecond)

	// Em produção, implementar com SendGrid, AWS SES, etc.
	emailContent := fmt.Sprintf(`
Subject: Agent4 Security - Verification Code

Dear User,

Your verification code for Agent4 Security is: %s

This code will expire in 10 minutes.

If you did not request this code, please ignore this email.

Best regards,
Agent4 Security Team
	`, code)

	fmt.Printf("Email sent to %s:\n%s\n", email, emailContent)

	return nil
}

// generateAlphanumericCode gera um código alfanumérico aleatório
func generateAlphanumericCode(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := ""

	for i := 0; i < length; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code += string(charset[index.Int64()])
	}

	return code, nil
}

// GetCodeStatus retorna o status de um código (para debugging)
func (e *EmailService) GetCodeStatus(email string) *EmailCodeData {
	if codeData, exists := e.codes[email]; exists {
		return &codeData
	}
	return nil
}

// CleanupExpiredCodes remove códigos expirados (executar periodicamente)
func (e *EmailService) CleanupExpiredCodes() {
	now := time.Now()
	for email, codeData := range e.codes {
		if now.After(codeData.ExpiresAt) {
			delete(e.codes, email)
		}
	}
}

// SendMFANotification envia notificação de nova autenticação MFA
func (e *EmailService) SendMFANotification(email, location, device string) error {
	emailContent := fmt.Sprintf(`
Subject: Agent4 Security - MFA Authentication Alert

Dear User,

A new multi-factor authentication was performed on your account:

Time: %s
Location: %s
Device: %s

If this was not you, please secure your account immediately.

Best regards,
Agent4 Security Team
	`, time.Now().Format("2006-01-02 15:04:05"), location, device)

	fmt.Printf("MFA notification sent to %s:\n%s\n", email, emailContent)

	return nil
}
