package services

import (
	"context"

	"github.com/equinoid/backend/internal/models"
)

func (d *D4SignService) RegisterDocument(ctx context.Context, equinoid string, docType string, filePath string) (string, error) {
	req := models.CreateD4SignDocumentRequest{
		Base64File:   filePath,
		Name:         docType,
		DocumentType: docType,
		Signers:      []models.D4SignSigner{},
	}
	
	doc, err := d.CreateDocument(ctx, equinoid, req, 1)
	if err != nil {
		return "", err
	}
	
	return doc.DocumentUUID, nil
}
