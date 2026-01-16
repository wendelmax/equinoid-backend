package services

import (
	"context"
	"fmt"

	"github.com/equinoid/backend/pkg/logging"
)

type SignatureTier string

const (
	Tier1Legal       SignatureTier = "tier1_legal"       // D4Sign
	Tier2Operational SignatureTier = "tier2_operational" // PKI Interna
	Tier3Basic       SignatureTier = "tier3_basic"       // PKI Interna Simples
)

type SignatureRouter struct {
	logger *logging.Logger
}

func NewSignatureRouter(logger *logging.Logger) *SignatureRouter {
	return &SignatureRouter{
		logger: logger,
	}
}

func (r *SignatureRouter) DetermineTier(documentType string) SignatureTier {
	switch documentType {
	case "transferencia", "contrato", "leilao", "exportacao":
		return Tier1Legal
	case "prontuario_veterinario", "certificacao_veterinaria", "validacao_genetica", "certificado_pedigree", "diario_treinamento":
		return Tier2Operational
	case "post_social", "oferta_venda", "configuracao_perfil", "comentario", "avaliacao", "like":
		return Tier3Basic
	default:
		r.logger.Warnf("Tipo de documento desconhecido: %s, usando Tier 2 por padrão", documentType)
		return Tier2Operational
	}
}

func (r *SignatureRouter) ShouldUseD4Sign(documentType string) bool {
	return r.DetermineTier(documentType) == Tier1Legal
}

func (r *SignatureRouter) GetSignatureMethod(ctx context.Context, documentType string) (string, error) {
	tier := r.DetermineTier(documentType)

	switch tier {
	case Tier1Legal:
		return "d4sign", nil
	case Tier2Operational, Tier3Basic:
		return "pki_internal", nil
	default:
		return "", fmt.Errorf("tier desconhecido: %s", tier)
	}
}

type SignatureRequest struct {
	DocumentType      string
	DocumentHash      string
	DocumentData      interface{}
	SignerID          uint
	SignerEmail       string
	SignerName        string
	RelatedEntityID   *uint
	RelatedEntityType string
	BiometricData     []byte
	CertificateID     *uint
}

type SignatureResponse struct {
	Success       bool
	Method        string
	DocumentUUID  string
	SignatureHash string
	Status        string
	Message       string
}
