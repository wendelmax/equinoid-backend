package blockchain

import (
	"context"
	"fmt"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BlockchainManager gerencia integrações com múltiplas blockchains
type BlockchainManager struct {
	db                 *gorm.DB
	ethereumService    *EthereumService
	hyperledgerService *HyperledgerService
	defaultNetwork     string
	enableEthereum     bool
	enableHyperledger  bool
}

// NewBlockchainManager cria um novo gerenciador blockchain
func NewBlockchainManager(
	db *gorm.DB,
	ethereumService *EthereumService,
	hyperledgerService *HyperledgerService,
	defaultNetwork string,
) *BlockchainManager {
	return &BlockchainManager{
		db:                 db,
		ethereumService:    ethereumService,
		hyperledgerService: hyperledgerService,
		defaultNetwork:     defaultNetwork,
		enableEthereum:     ethereumService != nil,
		enableHyperledger:  hyperledgerService != nil,
	}
}

// StorageResult representa resultado de armazenamento na blockchain
type StorageResult struct {
	Network     string    `json:"network"`
	TxHash      string    `json:"tx_hash"`
	BlockNumber uint64    `json:"block_number"`
	Status      string    `json:"status"`
	GasUsed     uint64    `json:"gas_used,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Error       string    `json:"error,omitempty"`
}

// StoreCertificateOnBlockchain armazena certificado na blockchain
func (bm *BlockchainManager) StoreCertificateOnBlockchain(ctx context.Context, certificate *models.Certificate) (*StorageResult, error) {
	network := bm.getNetworkForRecord("certificate")

	switch network {
	case "ethereum":
		return bm.storeCertificateEthereum(certificate)
	case "hyperledger":
		return bm.storeCertificateHyperledger(certificate)
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
}

// StoreSignatureOnBlockchain armazena assinatura na blockchain
func (bm *BlockchainManager) StoreSignatureOnBlockchain(ctx context.Context, signature *models.DigitalSignature) (*StorageResult, error) {
	network := bm.getNetworkForRecord("signature")

	switch network {
	case "ethereum":
		return bm.storeSignatureEthereum(signature)
	case "hyperledger":
		return bm.storeSignatureHyperledger(signature)
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
}

// StoreAuditLogOnBlockchain armazena log de auditoria na blockchain
func (bm *BlockchainManager) StoreAuditLogOnBlockchain(ctx context.Context, auditLog *models.AuditLog) (*StorageResult, error) {
	network := bm.getNetworkForRecord("audit")

	switch network {
	case "ethereum":
		return bm.storeAuditEthereum(auditLog)
	case "hyperledger":
		return bm.storeAuditHyperledger(auditLog)
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
}

// BatchStoreOnBlockchain armazena múltiplos registros em lote
func (bm *BlockchainManager) BatchStoreOnBlockchain(ctx context.Context, items []interface{}) ([]*StorageResult, error) {
	results := make([]*StorageResult, len(items))

	for i, item := range items {
		var result *StorageResult
		var err error

		switch v := item.(type) {
		case *models.Certificate:
			result, err = bm.StoreCertificateOnBlockchain(ctx, v)
		case *models.DigitalSignature:
			result, err = bm.StoreSignatureOnBlockchain(ctx, v)
		case *models.AuditLog:
			result, err = bm.StoreAuditLogOnBlockchain(ctx, v)
		default:
			err = fmt.Errorf("unsupported item type: %T", v)
		}

		if err != nil {
			result = &StorageResult{
				Status: "failed",
				Error:  err.Error(),
			}
		}

		results[i] = result
	}

	return results, nil
}

// VerifyRecordOnBlockchain verifica integridade de um registro
func (bm *BlockchainManager) VerifyRecordOnBlockchain(ctx context.Context, recordID uuid.UUID, network string) (bool, error) {
	// Buscar registro na blockchain diretamente
	switch network {
	case "ethereum":
		if !bm.enableEthereum {
			return false, fmt.Errorf("ethereum not enabled")
		}
		// Para Ethereum, precisamos do txHash que não temos aqui
		// Retornar erro indicando que precisa do txHash
		return false, fmt.Errorf("ethereum verification requires txHash, use VerifyRecord directly")

	case "hyperledger":
		if !bm.enableHyperledger {
			return false, fmt.Errorf("hyperledger not enabled")
		}
		return bm.hyperledgerService.VerifyRecord(recordID.String())

	default:
		return false, fmt.Errorf("unsupported network: %s", network)
	}
}

// GetTransactionStatus obtém status de uma transação
func (bm *BlockchainManager) GetTransactionStatus(ctx context.Context, txHash string, network string) (*StorageResult, error) {
	switch network {
	case "ethereum":
		if !bm.enableEthereum {
			return nil, fmt.Errorf("ethereum not enabled")
		}

		txResult, err := bm.ethereumService.GetTransactionStatus(txHash)
		if err != nil {
			return nil, err
		}

		return &StorageResult{
			Network:     "ethereum",
			TxHash:      txResult.TxHash,
			BlockNumber: txResult.BlockNumber,
			Status:      txResult.Status,
			GasUsed:     txResult.GasUsed,
			Timestamp:   txResult.Timestamp,
		}, nil

	case "hyperledger":
		if !bm.enableHyperledger {
			return nil, fmt.Errorf("hyperledger not enabled")
		}

		// Para Hyperledger, usar txHash como recordID
		response, err := bm.hyperledgerService.QueryChaincode("GetTransactionStatus", []string{txHash})
		if err != nil {
			return nil, err
		}

		return &StorageResult{
			Network:     "hyperledger",
			TxHash:      response.TxID,
			BlockNumber: response.BlockNumber,
			Status:      response.Status,
			Timestamp:   response.Timestamp,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
}

// GetBlockchainStats obtém estatísticas das blockchains
func (bm *BlockchainManager) GetBlockchainStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Estatísticas gerais do banco
	var totalRecords int64
	bm.db.Model(&models.BlockchainRecord{}).Count(&totalRecords)

	// Estatísticas por rede
	var ethereumRecords, hyperledgerRecords int64
	bm.db.Model(&models.BlockchainRecord{}).Where("network = ?", "ethereum").Count(&ethereumRecords)
	bm.db.Model(&models.BlockchainRecord{}).Where("network = ?", "hyperledger").Count(&hyperledgerRecords)

	// Estatísticas por tipo
	var certRecords, sigRecords, auditRecords int64
	bm.db.Model(&models.BlockchainRecord{}).Where("record_type = ?", "certificate").Count(&certRecords)
	bm.db.Model(&models.BlockchainRecord{}).Where("record_type = ?", "signature").Count(&sigRecords)
	bm.db.Model(&models.BlockchainRecord{}).Where("record_type = ?", "audit").Count(&auditRecords)

	// Estatísticas por status
	var confirmedRecords, pendingRecords, failedRecords int64
	bm.db.Model(&models.BlockchainRecord{}).Where("status = ?", "confirmed").Count(&confirmedRecords)
	bm.db.Model(&models.BlockchainRecord{}).Where("status = ?", "pending").Count(&pendingRecords)
	bm.db.Model(&models.BlockchainRecord{}).Where("status = ?", "failed").Count(&failedRecords)

	stats["total_records"] = totalRecords
	stats["networks"] = map[string]interface{}{
		"ethereum":    ethereumRecords,
		"hyperledger": hyperledgerRecords,
	}
	stats["types"] = map[string]interface{}{
		"certificates": certRecords,
		"signatures":   sigRecords,
		"audits":       auditRecords,
	}
	stats["status"] = map[string]interface{}{
		"confirmed": confirmedRecords,
		"pending":   pendingRecords,
		"failed":    failedRecords,
	}

	return stats, nil
}

// Funções internas para cada rede

func (bm *BlockchainManager) storeCertificateEthereum(certificate *models.Certificate) (*StorageResult, error) {
	if !bm.enableEthereum {
		return nil, fmt.Errorf("ethereum not enabled")
	}

	txResult, err := bm.ethereumService.StoreCertificate(certificate)
	if err != nil {
		return &StorageResult{
			Network: "ethereum",
			Status:  "failed",
			Error:   err.Error(),
		}, nil
	}

	// Salvar registro no banco
	blockchainRecord := &models.BlockchainRecord{
		RecordType:    "certificate",
		RecordID:      certificate.ID,
		DataHash:      bm.ethereumService.calculateCertificateHash(certificate),
		TxHash:        txResult.TxHash,
		BlockNumber:   txResult.BlockNumber,
		Network:       "ethereum",
		Status:        txResult.Status,
		GasUsed:       &txResult.GasUsed,
		Confirmations: txResult.Confirmations,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	bm.db.Create(blockchainRecord)

	return &StorageResult{
		Network:     "ethereum",
		TxHash:      txResult.TxHash,
		BlockNumber: txResult.BlockNumber,
		Status:      txResult.Status,
		GasUsed:     txResult.GasUsed,
		Timestamp:   txResult.Timestamp,
	}, nil
}

func (bm *BlockchainManager) storeCertificateHyperledger(certificate *models.Certificate) (*StorageResult, error) {
	if !bm.enableHyperledger {
		return nil, fmt.Errorf("hyperledger not enabled")
	}

	response, err := bm.hyperledgerService.StoreCertificate(certificate)
	if err != nil {
		return &StorageResult{
			Network: "hyperledger",
			Status:  "failed",
			Error:   err.Error(),
		}, nil
	}

	// Salvar registro no banco
	blockchainRecord := &models.BlockchainRecord{
		RecordType:  "certificate",
		RecordID:    certificate.ID,
		DataHash:    bm.hyperledgerService.calculateHash(certificate.SerialNumber),
		TxHash:      response.TxID,
		BlockNumber: response.BlockNumber,
		Network:     "hyperledger",
		Status:      response.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	bm.db.Create(blockchainRecord)

	return &StorageResult{
		Network:     "hyperledger",
		TxHash:      response.TxID,
		BlockNumber: response.BlockNumber,
		Status:      response.Status,
		Timestamp:   response.Timestamp,
	}, nil
}

func (bm *BlockchainManager) storeSignatureEthereum(signature *models.DigitalSignature) (*StorageResult, error) {
	if !bm.enableEthereum {
		return nil, fmt.Errorf("ethereum not enabled")
	}

	txResult, err := bm.ethereumService.StoreSignature(signature)
	if err != nil {
		return &StorageResult{
			Network: "ethereum",
			Status:  "failed",
			Error:   err.Error(),
		}, nil
	}

	blockchainRecord := &models.BlockchainRecord{
		RecordType:    "signature",
		RecordID:      signature.ID,
		DataHash:      signature.SignatureHash,
		TxHash:        txResult.TxHash,
		BlockNumber:   txResult.BlockNumber,
		Network:       "ethereum",
		Status:        txResult.Status,
		GasUsed:       &txResult.GasUsed,
		Confirmations: txResult.Confirmations,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	bm.db.Create(blockchainRecord)

	return &StorageResult{
		Network:     "ethereum",
		TxHash:      txResult.TxHash,
		BlockNumber: txResult.BlockNumber,
		Status:      txResult.Status,
		GasUsed:     txResult.GasUsed,
		Timestamp:   txResult.Timestamp,
	}, nil
}

func (bm *BlockchainManager) storeSignatureHyperledger(signature *models.DigitalSignature) (*StorageResult, error) {
	if !bm.enableHyperledger {
		return nil, fmt.Errorf("hyperledger not enabled")
	}

	response, err := bm.hyperledgerService.StoreSignature(signature)
	if err != nil {
		return &StorageResult{
			Network: "hyperledger",
			Status:  "failed",
			Error:   err.Error(),
		}, nil
	}

	blockchainRecord := &models.BlockchainRecord{
		RecordType:  "signature",
		RecordID:    signature.ID,
		DataHash:    signature.SignatureHash,
		TxHash:      response.TxID,
		BlockNumber: response.BlockNumber,
		Network:     "hyperledger",
		Status:      response.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	bm.db.Create(blockchainRecord)

	return &StorageResult{
		Network:     "hyperledger",
		TxHash:      response.TxID,
		BlockNumber: response.BlockNumber,
		Status:      response.Status,
		Timestamp:   response.Timestamp,
	}, nil
}

func (bm *BlockchainManager) storeAuditEthereum(auditLog *models.AuditLog) (*StorageResult, error) {
	if !bm.enableEthereum {
		return nil, fmt.Errorf("ethereum not enabled")
	}

	txResult, err := bm.ethereumService.StoreAuditLog(auditLog)
	if err != nil {
		return &StorageResult{
			Network: "ethereum",
			Status:  "failed",
			Error:   err.Error(),
		}, nil
	}

	blockchainRecord := &models.BlockchainRecord{
		RecordType:    "audit",
		RecordID:      auditLog.ID,
		DataHash:      bm.ethereumService.calculateAuditHash(auditLog),
		TxHash:        txResult.TxHash,
		BlockNumber:   txResult.BlockNumber,
		Network:       "ethereum",
		Status:        txResult.Status,
		GasUsed:       &txResult.GasUsed,
		Confirmations: txResult.Confirmations,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	bm.db.Create(blockchainRecord)

	return &StorageResult{
		Network:     "ethereum",
		TxHash:      txResult.TxHash,
		BlockNumber: txResult.BlockNumber,
		Status:      txResult.Status,
		GasUsed:     txResult.GasUsed,
		Timestamp:   txResult.Timestamp,
	}, nil
}

func (bm *BlockchainManager) storeAuditHyperledger(auditLog *models.AuditLog) (*StorageResult, error) {
	if !bm.enableHyperledger {
		return nil, fmt.Errorf("hyperledger not enabled")
	}

	response, err := bm.hyperledgerService.StoreAuditLog(auditLog)
	if err != nil {
		return &StorageResult{
			Network: "hyperledger",
			Status:  "failed",
			Error:   err.Error(),
		}, nil
	}

	blockchainRecord := &models.BlockchainRecord{
		RecordType:  "audit",
		RecordID:    auditLog.ID,
		DataHash:    bm.hyperledgerService.calculateAuditHash(auditLog),
		TxHash:      response.TxID,
		BlockNumber: response.BlockNumber,
		Network:     "hyperledger",
		Status:      response.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	bm.db.Create(blockchainRecord)

	return &StorageResult{
		Network:     "hyperledger",
		TxHash:      response.TxID,
		BlockNumber: response.BlockNumber,
		Status:      response.Status,
		Timestamp:   response.Timestamp,
	}, nil
}

// getNetworkForRecord determina qual rede usar para um tipo de registro
func (bm *BlockchainManager) getNetworkForRecord(recordType string) string {
	// Lógica de seleção de rede baseada no tipo
	switch recordType {
	case "certificate", "signature":
		// Certificados e assinaturas em Ethereum para imutabilidade pública
		if bm.enableEthereum {
			return "ethereum"
		}
		if bm.enableHyperledger {
			return "hyperledger"
		}
	case "audit":
		// Logs de auditoria em Hyperledger para privacidade
		if bm.enableHyperledger {
			return "hyperledger"
		}
		if bm.enableEthereum {
			return "ethereum"
		}
	}

	return bm.defaultNetwork
}
