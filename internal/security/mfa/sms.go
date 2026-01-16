package mfa

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// SMSService gerencia autenticação por SMS
type SMSService struct {
	codes map[string]CodeData // Em produção, usar Redis
}

// CodeData representa dados de um código de verificação
type CodeData struct {
	Code      string
	ExpiresAt time.Time
	Attempts  int
}

// NewSMSService cria um novo serviço SMS
func NewSMSService() *SMSService {
	return &SMSService{
		codes: make(map[string]CodeData),
	}
}

// SendVerificationCode envia um código de verificação por SMS
func (s *SMSService) SendVerificationCode(phoneNumber string) (string, error) {
	// Gerar código de 6 dígitos
	code, err := generateNumericCode(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate SMS code: %w", err)
	}

	// Armazenar código com expiração de 5 minutos
	s.codes[phoneNumber] = CodeData{
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Attempts:  0,
	}

	// Simular envio de SMS (em produção, usar serviço real como Twilio)
	err = s.sendSMS(phoneNumber, code)
	if err != nil {
		delete(s.codes, phoneNumber)
		return "", fmt.Errorf("failed to send SMS: %w", err)
	}

	return code, nil
}

// VerifyCode verifica um código SMS
func (s *SMSService) VerifyCode(phoneNumber, providedCode string) (bool, error) {
	codeData, exists := s.codes[phoneNumber]
	if !exists {
		return false, fmt.Errorf("no verification code found for this phone number")
	}

	// Verificar se código não expirou
	if time.Now().After(codeData.ExpiresAt) {
		delete(s.codes, phoneNumber)
		return false, fmt.Errorf("verification code has expired")
	}

	// Incrementar tentativas
	codeData.Attempts++
	s.codes[phoneNumber] = codeData

	// Verificar limite de tentativas
	if codeData.Attempts > 3 {
		delete(s.codes, phoneNumber)
		return false, fmt.Errorf("too many attempts, please request a new code")
	}

	// Verificar código
	if codeData.Code == providedCode {
		delete(s.codes, phoneNumber) // Código usado, remover
		return true, nil
	}

	return false, nil
}

// ResendCode reenvia o código de verificação
func (s *SMSService) ResendCode(phoneNumber string) (string, error) {
	// Verificar se existe código pendente
	if codeData, exists := s.codes[phoneNumber]; exists {
		// Verificar se ainda não passou 1 minuto desde o último envio
		if time.Since(codeData.ExpiresAt.Add(-5*time.Minute)) < time.Minute {
			return "", fmt.Errorf("please wait before requesting a new code")
		}
	}

	// Enviar novo código
	return s.SendVerificationCode(phoneNumber)
}

// sendSMS simula o envio de SMS (implementar com serviço real)
func (s *SMSService) sendSMS(phoneNumber, code string) error {
	// Simular delay de envio
	time.Sleep(100 * time.Millisecond)

	// Em produção, implementar com Twilio, AWS SNS, etc.
	fmt.Printf("SMS sent to %s: Your verification code is %s\n", phoneNumber, code)

	return nil
}

// generateNumericCode gera um código numérico aleatório
func generateNumericCode(length int) (string, error) {
	code := ""
	for i := 0; i < length; i++ {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += digit.String()
	}
	return code, nil
}

// GetCodeStatus retorna o status de um código (para debugging)
func (s *SMSService) GetCodeStatus(phoneNumber string) *CodeData {
	if codeData, exists := s.codes[phoneNumber]; exists {
		return &codeData
	}
	return nil
}

// CleanupExpiredCodes remove códigos expirados (executar periodicamente)
func (s *SMSService) CleanupExpiredCodes() {
	now := time.Now()
	for phoneNumber, codeData := range s.codes {
		if now.After(codeData.ExpiresAt) {
			delete(s.codes, phoneNumber)
		}
	}
}
