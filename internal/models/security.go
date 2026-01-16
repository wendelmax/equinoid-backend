package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     *uuid.UUID     `json:"user_id"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	ResourceID *uuid.UUID     `json:"resource_id"`
	Details    JSONB          `json:"details" gorm:"type:jsonb"`
	IPAddress  string         `json:"ip_address"`
	UserAgent  string         `json:"user_agent"`
	Location   string         `json:"location"`
	Success    bool           `json:"success"`
	ErrorMsg   string         `json:"error_msg"`
	RiskLevel  string         `json:"risk_level"`
	Timestamp  time.Time      `json:"timestamp"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type BiometricData struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	UserID            uuid.UUID      `json:"user_id"`
	BiometricType     string         `json:"biometric_type"`
	BiometricTemplate []byte         `json:"biometric_template"`
	Quality           float64        `json:"quality"`
	EnrollmentDate    time.Time      `json:"enrollment_date"`
	IsActive          bool           `json:"is_active" gorm:"default:true"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type DigitalSignature struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uuid.UUID      `json:"user_id"`
	DocumentID    uuid.UUID      `json:"document_id"`
	CertificateID uint           `json:"certificate_id"`
	Signature     string         `json:"signature"`
	SignatureHash string         `json:"signature_hash"`
	DocumentHash  string         `json:"document_hash"`
	Algorithm     string         `json:"algorithm"`
	Timestamp     time.Time      `json:"timestamp"`
	Location      string         `json:"location"`
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type ComplianceRecord struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	UserID         uuid.UUID      `json:"user_id"`
	ComplianceType string         `json:"compliance_type"`
	RequestType    string         `json:"request_type"`
	Status         string         `json:"status"`
	RequestData    JSONB          `json:"request_data" gorm:"type:jsonb"`
	LegalBasis     string         `json:"legal_basis"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type MFADevice struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uuid.UUID      `json:"user_id"`
	DeviceType  string         `json:"device_type"`
	DeviceName  string         `json:"device_name"`
	Secret      string         `json:"secret"`
	BackupCodes JSONB          `json:"backup_codes" gorm:"type:jsonb"`
	IsActive    bool           `json:"is_active" gorm:"default:false"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type Session struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uuid.UUID      `json:"user_id"`
	Token     string         `json:"token" gorm:"uniqueIndex"`
	IP        string         `json:"ip"`
	UserAgent string         `json:"user_agent"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type BlockchainRecord struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	RecordType    string         `json:"record_type" gorm:"not null"` // certificate, signature, audit
	RecordID      uint           `json:"record_id" gorm:"not null"`
	DataHash      string         `json:"data_hash" gorm:"size:255;not null"`
	TxHash        string         `json:"tx_hash" gorm:"size:255;not null"`
	BlockNumber   uint64         `json:"block_number"`
	Network       string         `json:"network" gorm:"size:50;not null"`         // ethereum, hyperledger
	Status        string         `json:"status" gorm:"size:50;default:'pending'"` // pending, confirmed, failed
	GasUsed       *uint64        `json:"gas_used,omitempty"`
	Confirmations int            `json:"confirmations" gorm:"default:0"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
