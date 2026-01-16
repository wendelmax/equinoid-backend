package biometric

import (
	stdcrypto "crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"time"

	"github.com/equinoid/backend/internal/models"
	cryptoservice "github.com/equinoid/backend/internal/security/crypto"
	"github.com/google/uuid"
)

// SignatureService gerencia assinatura digital biométrica
type SignatureService struct {
	faceIDService     *FaceIDService
	encryptionService *cryptoservice.EncryptionService
}

// NewSignatureService cria um novo serviço de assinatura
func NewSignatureService(faceIDService *FaceIDService, encryptionService *cryptoservice.EncryptionService) *SignatureService {
	return &SignatureService{
		faceIDService:     faceIDService,
		encryptionService: encryptionService,
	}
}

// SignatureRequest representa uma solicitação de assinatura
type SignatureRequest struct {
	UserID        uuid.UUID `json:"user_id"`
	DocumentID    uuid.UUID `json:"document_id"`
	DocumentHash  string    `json:"document_hash"`
	BiometricData []byte    `json:"biometric_data"` // Imagem facial
	Location      string    `json:"location"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
}

// SignatureResponse representa o resultado da assinatura
type SignatureResponse struct {
	Success           bool                `json:"success"`
	SignatureID       uint                `json:"signature_id"`
	Signature         []byte              `json:"signature"`
	BiometricVerified bool                `json:"biometric_verified"`
	BiometricScore    float64             `json:"biometric_score"`
	Timestamp         time.Time           `json:"timestamp"`
	Certificate       *models.Certificate `json:"certificate"`
	BlockchainTxHash  string              `json:"blockchain_tx_hash,omitempty"`
	Message           string              `json:"message"`
}

// DocumentSigningData representa dados para assinatura de documento
type DocumentSigningData struct {
	UserID        uuid.UUID `json:"user_id"`
	DocumentHash  string    `json:"document_hash"`
	Timestamp     time.Time `json:"timestamp"`
	Location      string    `json:"location"`
	IPAddress     string    `json:"ip_address"`
	BiometricHash string    `json:"biometric_hash"`
}

// SignDocument assina um documento com verificação biométrica
func (s *SignatureService) SignDocument(req *SignatureRequest, certificate *models.Certificate, storedBiometric []byte) (*SignatureResponse, error) {
	timestamp := time.Now()

	// 1. Verificar biometria
	verificationReq := &VerificationRequest{
		UserID:    req.UserID,
		ImageData: req.BiometricData,
	}

	biometricResult, err := s.faceIDService.VerifyFace(verificationReq, storedBiometric)
	if err != nil {
		return &SignatureResponse{
			Success: false,
			Message: fmt.Sprintf("Biometric verification failed: %v", err),
		}, nil
	}

	if !biometricResult.Success {
		return &SignatureResponse{
			Success:           false,
			BiometricVerified: false,
			BiometricScore:    biometricResult.Score,
			Message:           "Biometric verification failed: " + biometricResult.Message,
		}, nil
	}

	// 2. Preparar dados para assinatura
	signingData := DocumentSigningData{
		UserID:        req.UserID,
		DocumentHash:  req.DocumentHash,
		Timestamp:     timestamp,
		Location:      req.Location,
		IPAddress:     req.IPAddress,
		BiometricHash: s.hashBiometricData(req.BiometricData),
	}

	signingDataJSON, err := json.Marshal(signingData)
	if err != nil {
		return &SignatureResponse{
			Success: false,
			Message: "Failed to prepare signing data",
		}, nil
	}

	// 3. Criar assinatura digital
	decryptedPrivateKey, err := s.encryptionService.Decrypt(certificate.PrivateKeyPEM)
	if err != nil {
		return &SignatureResponse{
			Success: false,
			Message: "Failed to decrypt certificate private key",
		}, nil
	}

	signature, err := s.createDigitalSignature(signingDataJSON, []byte(decryptedPrivateKey))
	if err != nil {
		return &SignatureResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create digital signature: %v", err),
		}, nil
	}

	// 4. Criar registro de assinatura
	signatureRecord := &models.DigitalSignature{
		UserID:        req.UserID,
		DocumentID:    req.DocumentID,
		CertificateID: certificate.ID,
		SignatureHash: s.hashSignature(signature),
		DocumentHash:  req.DocumentHash,
		Signature:     string(signature),
		Algorithm:     "RSA-SHA256",
		Timestamp:     timestamp,
		Location:      req.Location,
		CreatedAt:     timestamp,
	}

	// 5. Preparar resposta
	response := &SignatureResponse{
		Success:           true,
		SignatureID:       signatureRecord.ID,
		Signature:         signature,
		BiometricVerified: true,
		BiometricScore:    biometricResult.Score,
		Timestamp:         timestamp,
		Certificate:       certificate,
		Message:           "Document signed successfully with biometric verification",
	}

	return response, nil
}

// VerifySignature verifica uma assinatura digital
func (s *SignatureService) VerifySignature(signatureData *models.DigitalSignature, certificate *models.Certificate) (bool, error) {
	// 1. Verificar se certificado não está revogado
	if certificate.IsRevoked {
		return false, fmt.Errorf("certificate is revoked")
	}

	// 2. Verificar se certificado não expirou
	if time.Now().After(certificate.ExpiresAt) {
		return false, fmt.Errorf("certificate has expired")
	}

	// 3. Reconstruir dados de assinatura
	signingData := DocumentSigningData{
		UserID:       signatureData.UserID,
		DocumentHash: signatureData.DocumentHash,
		Timestamp:    signatureData.Timestamp,
		Location:     signatureData.Location,
	}

	signingDataJSON, err := json.Marshal(signingData)
	if err != nil {
		return false, fmt.Errorf("failed to reconstruct signing data: %w", err)
	}

	// 4. Verificar assinatura digital
	valid, err := s.verifyDigitalSignature(signingDataJSON, []byte(signatureData.Signature), []byte(certificate.PublicKeyPEM))
	if err != nil {
		return false, fmt.Errorf("signature verification failed: %w", err)
	}

	return valid, nil
}

// createDigitalSignature cria uma assinatura digital RSA
func (s *SignatureService) createDigitalSignature(data, privateKeyPEM []byte) ([]byte, error) {
	// Parse da chave privada
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyPEM)
	if err != nil {
		// Tentar PKCS8
		pkcs8Key, err2 := x509.ParsePKCS8PrivateKey(privateKeyPEM)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse private key: %v, %v", err, err2)
		}
		var ok bool
		privateKey, ok = pkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key is not RSA private key")
		}
	}

	// Hash dos dados
	hash := sha256.Sum256(data)

	// Criar assinatura
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, stdcrypto.SHA256, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return signature, nil
}

// verifyDigitalSignature verifica uma assinatura digital RSA
func (s *SignatureService) verifyDigitalSignature(data, signature, publicKeyPEM []byte) (bool, error) {
	// Parse da chave pública
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyPEM)
	if err != nil {
		return false, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("key is not RSA public key")
	}

	// Hash dos dados
	hash := sha256.Sum256(data)

	// Verificar assinatura
	err = rsa.VerifyPKCS1v15(rsaPublicKey, stdcrypto.SHA256, hash[:], signature)
	return err == nil, nil
}

// createBiometricProof cria uma prova biométrica criptografada
func (s *SignatureService) createBiometricProof(biometricData []byte, result *VerificationResponse) []byte {
	proof := map[string]interface{}{
		"hash":       s.hashBiometricData(biometricData),
		"score":      result.Score,
		"quality":    result.Quality,
		"is_live":    result.IsLive,
		"session_id": result.SessionID,
		"timestamp":  time.Now().Unix(),
	}

	proofJSON, _ := json.Marshal(proof)

	// Criptografar proof com chave AES
	encryptedProof, err := s.encryptionService.EncryptBytes(proofJSON)
	if err != nil {
		return proofJSON // Fallback para não criptografado se falhar
	}
	return encryptedProof
}

// extractBiometricHash extrai hash biométrico da prova
func (s *SignatureService) extractBiometricHash(biometricProof []byte) string {
	// Descriptografar proof
	decryptedProof, err := s.encryptionService.DecryptBytes(biometricProof)
	if err != nil {
		decryptedProof = biometricProof // Fallback para tentar ler se não estiver criptografado
	}

	var proof map[string]interface{}
	if err := json.Unmarshal(decryptedProof, &proof); err != nil {
		return ""
	}

	if hash, ok := proof["hash"].(string); ok {
		return hash
	}
	return ""
}

// hashBiometricData cria hash dos dados biométricos
func (s *SignatureService) hashBiometricData(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// hashSignature cria hash da assinatura
func (s *SignatureService) hashSignature(signature []byte) string {
	hash := sha256.Sum256(signature)
	return fmt.Sprintf("%x", hash)
}

// BatchSignDocuments assina múltiplos documentos em lote
func (s *SignatureService) BatchSignDocuments(requests []*SignatureRequest, certificate *models.Certificate, storedBiometric []byte) ([]*SignatureResponse, error) {
	if len(requests) == 0 {
		return nil, fmt.Errorf("no documents to sign")
	}

	// Verificar biometria uma vez para todo o lote
	firstReq := requests[0]
	verificationReq := &VerificationRequest{
		UserID:    firstReq.UserID,
		ImageData: firstReq.BiometricData,
	}

	biometricResult, err := s.faceIDService.VerifyFace(verificationReq, storedBiometric)
	if err != nil || !biometricResult.Success {
		return nil, fmt.Errorf("biometric verification failed for batch signing")
	}

	responses := make([]*SignatureResponse, len(requests))

	for i, req := range requests {
		// Usar a mesma verificação biométrica para todos
		response, err := s.signWithVerifiedBiometric(req, certificate, biometricResult)
		if err != nil {
			response = &SignatureResponse{
				Success: false,
				Message: err.Error(),
			}
		}
		responses[i] = response
	}

	return responses, nil
}

// signWithVerifiedBiometric assina documento com biometria já verificada
func (s *SignatureService) signWithVerifiedBiometric(req *SignatureRequest, certificate *models.Certificate, biometricResult *VerificationResponse) (*SignatureResponse, error) {
	timestamp := time.Now()

	// Preparar dados para assinatura
	signingData := DocumentSigningData{
		UserID:        req.UserID,
		DocumentHash:  req.DocumentHash,
		Timestamp:     timestamp,
		Location:      req.Location,
		IPAddress:     req.IPAddress,
		BiometricHash: s.hashBiometricData(req.BiometricData),
	}

	signingDataJSON, err := json.Marshal(signingData)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare signing data")
	}

	// Criar assinatura digital
	decryptedPrivateKey, err := s.encryptionService.Decrypt(certificate.PrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt certificate private key")
	}

	signature, err := s.createDigitalSignature(signingDataJSON, []byte(decryptedPrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create digital signature: %w", err)
	}

	// Criar registro de assinatura
	signatureRecord := &models.DigitalSignature{
		UserID:        req.UserID,
		DocumentID:    req.DocumentID,
		CertificateID: certificate.ID,
		SignatureHash: s.hashSignature(signature),
		DocumentHash:  req.DocumentHash,
		Signature:     string(signature),
		Algorithm:     "RSA-SHA256",
		Timestamp:     timestamp,
		Location:      req.Location,
		CreatedAt:     timestamp,
	}

	return &SignatureResponse{
		Success:           true,
		SignatureID:       signatureRecord.ID,
		Signature:         signature,
		BiometricVerified: true,
		BiometricScore:    biometricResult.Score,
		Timestamp:         timestamp,
		Certificate:       certificate,
		Message:           "Document signed successfully",
	}, nil
}

// GetSignatureMetadata retorna metadados de uma assinatura
func (s *SignatureService) GetSignatureMetadata(signatureData *models.DigitalSignature) map[string]interface{} {
	return map[string]interface{}{
		"signature_id":   signatureData.ID,
		"user_id":        signatureData.UserID,
		"document_id":    signatureData.DocumentID,
		"certificate_id": signatureData.CertificateID,
		"algorithm":      signatureData.Algorithm,
		"timestamp":      signatureData.Timestamp,
		"location":       signatureData.Location,
	}
}
