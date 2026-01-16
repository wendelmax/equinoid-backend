package models

import (
	"time"

	"gorm.io/gorm"
)

// D4SignDocument representa um documento na D4Sign armazenado localmente
type D4SignDocument struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	DocumentUUID      string         `json:"document_uuid" gorm:"uniqueIndex;not null"`
	SafeUUID          string         `json:"safe_uuid" gorm:"not null"`
	Name              string         `json:"name" gorm:"not null"`
	Status            string         `json:"status" gorm:"default:'pending'"` // pending, signed, cancelled, expired
	DocumentType      string         `json:"document_type" gorm:"not null"`   // transferencia, contrato, leilao, exportacao
	RelatedEntityID   *uint          `json:"related_entity_id"`               // ID do equino, contrato, etc
	RelatedEntityType string         `json:"related_entity_type"`             // equino, contrato, leilao
	CreatedBy         uint           `json:"created_by" gorm:"not null"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	SignedAt          *time.Time     `json:"signed_at,omitempty"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	Creator *User `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// D4SignSigner representa um signatário de um documento
type D4SignSigner struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"` // 1-Signer, 2-Approver, etc.
}

// CreateD4SignDocumentRequest representa a requisição para criar um documento
type CreateD4SignDocumentRequest struct {
	Base64File        string         `json:"base64_file" validate:"required"`
	Name              string         `json:"name" validate:"required"`
	DocumentType      string         `json:"document_type" validate:"required,oneof=transferencia contrato leilao exportacao"`
	RelatedEntityID   *uint          `json:"related_entity_id"`
	RelatedEntityType string         `json:"related_entity_type"`
	Signers           []D4SignSigner `json:"signers" validate:"required,min=1"`
	SafeUUID          string         `json:"safe_uuid"`
}

// D4SignWebhookPayload representa o payload do webhook da D4Sign
type D4SignWebhookPayload struct {
	Event     string                 `json:"event"` // document_signed, document_cancelled, etc
	Document  D4SignWebhookDocument  `json:"document"`
	Signer    D4SignWebhookSigner    `json:"signer,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// D4SignWebhookDocument representa o documento no webhook
type D4SignWebhookDocument struct {
	UUID   string `json:"uuid"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// D4SignWebhookSigner representa o signatário no webhook
type D4SignWebhookSigner struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// D4SignDocumentStatusResponse representa a resposta de status do documento
type D4SignDocumentStatusResponse struct {
	UUID      string               `json:"uuid"`
	Name      string               `json:"name"`
	Status    string               `json:"status"`
	SignedAt  *time.Time           `json:"signed_at,omitempty"`
	Signers   []D4SignSignerStatus `json:"signers"`
	CreatedAt time.Time            `json:"created_at"`
}

// D4SignSignerStatus representa o status de um signatário
type D4SignSignerStatus struct {
	Email    string     `json:"email"`
	Name     string     `json:"name"`
	Status   string     `json:"status"` // pending, signed, rejected
	SignedAt *time.Time `json:"signed_at,omitempty"`
}

// D4SignResponse representa uma resposta genérica da D4Sign
type D4SignResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
