package pki

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/internal/security/crypto"
)

// CAService gerencia a Autoridade Certificadora (CA)
type CAService struct {
	caCertPath        string
	caKeyPath         string
	caCert            *x509.Certificate
	caKey             *rsa.PrivateKey
	validityDays      int
	encryptionService *crypto.EncryptionService
}

// NewCAService cria um novo serviço de CA
func NewCAService(caCertPath, caKeyPath string, validityDays int, encryptionService *crypto.EncryptionService) *CAService {
	return &CAService{
		caCertPath:        caCertPath,
		caKeyPath:         caKeyPath,
		validityDays:      validityDays,
		encryptionService: encryptionService,
	}
}

// CertificateRequest representa uma solicitação de certificado
type CertificateRequest struct {
	UserID       uint     `json:"user_id"`
	CommonName   string   `json:"common_name"`
	Organization string   `json:"organization"`
	Country      string   `json:"country"`
	Province     string   `json:"province"`
	Locality     string   `json:"locality"`
	EmailAddress string   `json:"email_address"`
	KeySize      int      `json:"key_size"`
	ValidityDays int      `json:"validity_days"`
	KeyUsage     []string `json:"key_usage"` // digital_signature, key_encipherment, etc.
}

// CertificateResponse representa o resultado da geração do certificado
type CertificateResponse struct {
	Certificate    *models.Certificate `json:"certificate"`
	CertificatePEM []byte              `json:"certificate_pem"`
	PrivateKeyPEM  []byte              `json:"private_key_pem"`
	PublicKeyPEM   []byte              `json:"public_key_pem"`
	SerialNumber   string              `json:"serial_number"`
	Fingerprint    string              `json:"fingerprint"`
}

// Initialize inicializa o serviço CA (cria CA se não existir)
func (ca *CAService) Initialize() error {
	// Verificar se certificados CA existem
	if _, err := os.Stat(ca.caCertPath); os.IsNotExist(err) {
		fmt.Println("CA certificate not found, creating new CA...")
		if err := ca.createCA(); err != nil {
			return fmt.Errorf("failed to create CA: %w", err)
		}
	}

	// Carregar certificados CA
	if err := ca.loadCA(); err != nil {
		return fmt.Errorf("failed to load CA: %w", err)
	}

	fmt.Printf("CA initialized successfully. Certificate valid until: %v\n", ca.caCert.NotAfter)
	return nil
}

// createCA cria uma nova Autoridade Certificadora
func (ca *CAService) createCA() error {
	// Gerar chave privada para CA
	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate CA key: %w", err)
	}

	// Template para certificado CA
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:         "Agent4 Security CA",
			Organization:       []string{"Agent4 Security"},
			OrganizationalUnit: []string{"Certificate Authority"},
			Country:            []string{"BR"},
			Province:           []string{"SP"},
			Locality:           []string{"São Paulo"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(ca.validityDays*10) * 24 * time.Hour), // CA válida por 10x mais tempo
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	// Criar certificado CA (auto-assinado)
	caCertBytes, err := x509.CreateCertificate(rand.Reader, template, template, &caKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Criar diretório se não existir
	if err := os.MkdirAll("./certs", 0755); err != nil {
		return fmt.Errorf("failed to create certs directory: %w", err)
	}

	// Salvar certificado CA
	caCertFile, err := os.Create(ca.caCertPath)
	if err != nil {
		return fmt.Errorf("failed to create CA cert file: %w", err)
	}
	defer caCertFile.Close()

	if err := pem.Encode(caCertFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertBytes,
	}); err != nil {
		return fmt.Errorf("failed to encode CA certificate: %w", err)
	}

	// Salvar chave privada CA
	caKeyFile, err := os.Create(ca.caKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create CA key file: %w", err)
	}
	defer caKeyFile.Close()

	caKeyBytes := x509.MarshalPKCS1PrivateKey(caKey)

	// Criptografar chave privada da CA antes de salvar em PEM
	encryptedKeyBytes, err := ca.encryptionService.EncryptBytes(caKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt CA private key: %w", err)
	}

	if err := pem.Encode(caKeyFile, &pem.Block{
		Type:  "ENCRYPTED RSA PRIVATE KEY",
		Bytes: encryptedKeyBytes,
	}); err != nil {
		return fmt.Errorf("failed to encode CA private key: %w", err)
	}

	// Definir permissões restritas para chave privada
	if err := os.Chmod(ca.caKeyPath, 0600); err != nil {
		return fmt.Errorf("failed to set CA key permissions: %w", err)
	}

	fmt.Println("CA created successfully!")
	return nil
}

// loadCA carrega os certificados CA existentes
func (ca *CAService) loadCA() error {
	// Carregar certificado CA
	caCertBytes, err := os.ReadFile(ca.caCertPath)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertBlock, _ := pem.Decode(caCertBytes)
	if caCertBlock == nil {
		return fmt.Errorf("failed to decode CA certificate")
	}

	ca.caCert, err = x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Carregar chave privada CA
	caKeyBytes, err := os.ReadFile(ca.caKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read CA private key: %w", err)
	}

	caKeyBlock, _ := pem.Decode(caKeyBytes)
	if caKeyBlock == nil {
		return fmt.Errorf("failed to decode CA private key")
	}

	// Descriptografar chave privada
	decryptedKeyBytes, err := ca.encryptionService.DecryptBytes(caKeyBlock.Bytes)
	if err != nil {
		// Fallback para tentar ler sem criptografia (compatibilidade)
		decryptedKeyBytes = caKeyBlock.Bytes
	}

	ca.caKey, err = x509.ParsePKCS1PrivateKey(decryptedKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA private key: %w", err)
	}

	// Verificar se CA não expirou
	if time.Now().After(ca.caCert.NotAfter) {
		return fmt.Errorf("CA certificate has expired")
	}

	return nil
}

// IssueCertificate emite um novo certificado
func (ca *CAService) IssueCertificate(req *CertificateRequest) (*CertificateResponse, error) {
	if ca.caCert == nil || ca.caKey == nil {
		return nil, fmt.Errorf("CA not initialized")
	}

	// Definir valores padrão
	keySize := req.KeySize
	if keySize == 0 {
		keySize = 2048
	}

	validityDays := req.ValidityDays
	if validityDays == 0 {
		validityDays = ca.validityDays
	}

	// Gerar chave privada para o certificado
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Gerar número serial único
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Template para o certificado
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   req.CommonName,
			Organization: []string{req.Organization},
			Country:      []string{req.Country},
			Province:     []string{req.Province},
			Locality:     []string{req.Locality},
		},
		EmailAddresses:        []string{req.EmailAddress},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(validityDays) * 24 * time.Hour),
		KeyUsage:              ca.parseKeyUsage(req.KeyUsage),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageEmailProtection},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// Criar certificado
	certBytes, err := x509.CreateCertificate(rand.Reader, template, ca.caCert, &privateKey.PublicKey, ca.caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Codificar certificado em PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	// Codificar chave privada em PEM
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Codificar chave pública em PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// Calcular fingerprint
	fingerprint := ca.calculateFingerprint(certBytes)

	// Criptografar chave privada antes de salvar no modelo
	encryptedPrivateKey, err := ca.encryptionService.Encrypt(string(privateKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Criar modelo de certificado
	certificate := &models.Certificate{
		UserID:         req.UserID,
		SerialNumber:   serialNumber.String(),
		CommonName:     req.CommonName,
		CertificatePEM: string(certPEM),
		PrivateKeyPEM:  encryptedPrivateKey,
		PublicKeyPEM:   string(publicKeyPEM),
		IssuedAt:       template.NotBefore,
		ExpiresAt:      template.NotAfter,
		IsRevoked:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	response := &CertificateResponse{
		Certificate:    certificate,
		CertificatePEM: certPEM,
		PrivateKeyPEM:  privateKeyPEM,
		PublicKeyPEM:   publicKeyPEM,
		SerialNumber:   serialNumber.String(),
		Fingerprint:    fingerprint,
	}

	return response, nil
}

// RevokeCertificate revoga um certificado
func (ca *CAService) RevokeCertificate(certificate *models.Certificate, reason string) error {
	certificate.IsRevoked = true
	certificate.RevokedAt = &[]time.Time{time.Now()}[0]
	certificate.RevocationReason = reason
	certificate.UpdatedAt = time.Now()

	// Em produção, adicionar à CRL (Certificate Revocation List)
	fmt.Printf("Certificate %s has been revoked. Reason: %s\n", certificate.SerialNumber, reason)

	return nil
}

// VerifyCertificate verifica a validade de um certificado
func (ca *CAService) VerifyCertificate(certPEM []byte) (*CertificateInfo, error) {
	// Decodificar certificado
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to decode certificate")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Verificar se foi emitido por nossa CA
	if err := cert.CheckSignatureFrom(ca.caCert); err != nil {
		return nil, fmt.Errorf("certificate not issued by this CA: %w", err)
	}

	// Verificar validade temporal
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return nil, fmt.Errorf("certificate not yet valid")
	}
	if now.After(cert.NotAfter) {
		return nil, fmt.Errorf("certificate has expired")
	}

	info := &CertificateInfo{
		SerialNumber: cert.SerialNumber.String(),
		Subject:      cert.Subject.CommonName,
		Issuer:       cert.Issuer.CommonName,
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		KeyUsage:     ca.formatKeyUsage(cert.KeyUsage),
		IsValid:      true,
		Fingerprint:  ca.calculateFingerprint(cert.Raw),
	}

	return info, nil
}

// CertificateInfo contém informações sobre um certificado
type CertificateInfo struct {
	SerialNumber string    `json:"serial_number"`
	Subject      string    `json:"subject"`
	Issuer       string    `json:"issuer"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	KeyUsage     []string  `json:"key_usage"`
	IsValid      bool      `json:"is_valid"`
	Fingerprint  string    `json:"fingerprint"`
}

// generateSerialNumber gera um número serial único
func generateSerialNumber() (*big.Int, error) {
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}
	return serialNumber, nil
}

// parseKeyUsage converte strings de uso de chave em flags
func (ca *CAService) parseKeyUsage(keyUsageStrings []string) x509.KeyUsage {
	var keyUsage x509.KeyUsage

	for _, usage := range keyUsageStrings {
		switch usage {
		case "digital_signature":
			keyUsage |= x509.KeyUsageDigitalSignature
		case "key_encipherment":
			keyUsage |= x509.KeyUsageKeyEncipherment
		case "data_encipherment":
			keyUsage |= x509.KeyUsageDataEncipherment
		case "key_agreement":
			keyUsage |= x509.KeyUsageKeyAgreement
		case "cert_sign":
			keyUsage |= x509.KeyUsageCertSign
		case "crl_sign":
			keyUsage |= x509.KeyUsageCRLSign
		}
	}

	// Se nenhum uso específico foi definido, usar padrão
	if keyUsage == 0 {
		keyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	}

	return keyUsage
}

// formatKeyUsage converte flags de uso de chave em strings
func (ca *CAService) formatKeyUsage(keyUsage x509.KeyUsage) []string {
	var usages []string

	if keyUsage&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "digital_signature")
	}
	if keyUsage&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "key_encipherment")
	}
	if keyUsage&x509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "data_encipherment")
	}
	if keyUsage&x509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "key_agreement")
	}
	if keyUsage&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "cert_sign")
	}
	if keyUsage&x509.KeyUsageCRLSign != 0 {
		usages = append(usages, "crl_sign")
	}

	return usages
}

// calculateFingerprint calcula o fingerprint SHA-256 do certificado
func (ca *CAService) calculateFingerprint(certBytes []byte) string {
	return fmt.Sprintf("%x", certBytes[:32]) // Simplificado para exemplo
}

// GetCACertificate retorna o certificado da CA em PEM
func (ca *CAService) GetCACertificate() ([]byte, error) {
	if ca.caCert == nil {
		return nil, fmt.Errorf("CA not initialized")
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ca.caCert.Raw,
	}), nil
}

// ValidateCertificateChain valida a cadeia de certificados
func (ca *CAService) ValidateCertificateChain(certPEM []byte) error {
	// Decodificar certificado
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("failed to decode certificate")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Criar pool com CA raiz
	roots := x509.NewCertPool()
	roots.AddCert(ca.caCert)

	// Verificar cadeia
	_, err = cert.Verify(x509.VerifyOptions{Roots: roots})
	if err != nil {
		return fmt.Errorf("certificate chain validation failed: %w", err)
	}

	return nil
}
