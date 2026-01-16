package models

import (
	"time"

	"gorm.io/gorm"
)

// Certificate representa um certificado digital
type Certificate struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	UserID           uint           `json:"user_id" gorm:"not null"`
	CertificatePEM   string         `json:"certificate_pem" gorm:"type:text;not null"`
	PrivateKeyPEM    string         `json:"private_key_pem" gorm:"type:text;not null"`
	PublicKeyPEM     string         `json:"public_key_pem" gorm:"type:text;not null"`
	CommonName       string         `json:"common_name" gorm:"size:100"`
	SerialNumber     string         `json:"serial_number" gorm:"uniqueIndex;size:100;not null"`
	IssuedAt         time.Time      `json:"issued_at" gorm:"not null"`
	ValidFrom        time.Time      `json:"valid_from" gorm:"not null"`
	ValidTo          time.Time      `json:"valid_to" gorm:"not null"`
	ExpiresAt        time.Time      `json:"expires_at" gorm:"not null"`
	IsRevoked        bool           `json:"is_revoked" gorm:"default:false"`
	RevokedAt        *time.Time     `json:"revoked_at,omitempty"`
	RevocationReason string         `json:"revocation_reason,omitempty" gorm:"type:text"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// CertificateResponse representa a resposta de certificado
type CertificateResponse struct {
	ID           uint       `json:"id"`
	SerialNumber string     `json:"serial_number"`
	ValidFrom    time.Time  `json:"valid_from"`
	ValidTo      time.Time  `json:"valid_to"`
	IsValid      bool       `json:"is_valid"`
	IsRevoked    bool       `json:"is_revoked"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// GenerateCertificateRequest representa a requisição de geração de certificado
type GenerateCertificateRequest struct {
	UserType     UserType `json:"user_type" validate:"required"`
	ValidityDays int      `json:"validity_days" validate:"required,min=1,max=3650"`
}

// GenerateCertificateResponse representa a resposta de geração de certificado
type GenerateCertificateResponse struct {
	Certificate    *CertificateResponse `json:"certificate"`
	CertificatePEM string               `json:"certificate_pem"`
	PrivateKey     string               `json:"private_key"`
}

// RevokeCertificateRequest representa a requisição de revogação de certificado
type RevokeCertificateRequest struct {
	SerialNumber string `json:"serial_number" validate:"required"`
	Reason       string `json:"reason" validate:"required"`
}

// ValidateCertificateResponse representa a resposta de validação de certificado
type ValidateCertificateResponse struct {
	IsValid     bool                 `json:"is_valid"`
	Certificate *CertificateResponse `json:"certificate,omitempty"`
}

// ToResponse converte Certificate para CertificateResponse
func (c *Certificate) ToResponse() *CertificateResponse {
	return &CertificateResponse{
		ID:           c.ID,
		SerialNumber: c.SerialNumber,
		ValidFrom:    c.ValidFrom,
		ValidTo:      c.ValidTo,
		IsValid:      c.IsValid(),
		IsRevoked:    c.IsRevoked,
		RevokedAt:    c.RevokedAt,
		CreatedAt:    c.CreatedAt,
	}
}

// IsValid verifica se o certificado é válido
func (c *Certificate) IsValid() bool {
	now := time.Now()
	return !c.IsRevoked && now.After(c.ValidFrom) && now.Before(c.ValidTo)
}

// IsExpired verifica se o certificado está expirado
func (c *Certificate) IsExpired() bool {
	return time.Now().After(c.ValidTo)
}

// Revoke revoga o certificado
func (c *Certificate) Revoke(reason string) {
	now := time.Now()
	c.IsRevoked = true
	c.RevokedAt = &now
	c.RevocationReason = reason
}

// BeforeCreate é executado antes de criar um certificado
func (c *Certificate) BeforeCreate(tx *gorm.DB) error {
	c.IsRevoked = false
	return nil
}
