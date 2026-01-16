package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/equinoid/backend/internal/models"
)

// EthereumService gerencia integração com blockchain Ethereum
type EthereumService struct {
	client       *ethclient.Client
	contractAddr common.Address
	privateKey   *ecdsa.PrivateKey
	chainID      *big.Int
	gasLimit     uint64
	gasPrice     *big.Int
}

// NewEthereumService cria um novo serviço Ethereum
func NewEthereumService(nodeURL, contractAddress, privateKeyHex string) (*EthereumService, error) {
	// Conectar ao nó Ethereum
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	// Parse da chave privada
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Obter chain ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Definir gas price padrão
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		gasPrice = big.NewInt(20000000000) // 20 gwei como fallback
	}

	return &EthereumService{
		client:       client,
		contractAddr: common.HexToAddress(contractAddress),
		privateKey:   privateKey,
		chainID:      chainID,
		gasLimit:     300000, // Limite padrão de gas
		gasPrice:     gasPrice,
	}, nil
}

// BlockchainRecord representa um registro na blockchain
type BlockchainRecord struct {
	RecordType  string                 `json:"record_type"`
	RecordID    string                 `json:"record_id"`
	DataHash    string                 `json:"data_hash"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   int64                  `json:"timestamp"`
	UserAddress string                 `json:"user_address"`
}

// TransactionResult representa o resultado de uma transação
type TransactionResult struct {
	TxHash        string    `json:"tx_hash"`
	BlockNumber   uint64    `json:"block_number"`
	GasUsed       uint64    `json:"gas_used"`
	Status        string    `json:"status"`
	Confirmations int       `json:"confirmations"`
	Timestamp     time.Time `json:"timestamp"`
}

// StoreRecord armazena um registro na blockchain
func (e *EthereumService) StoreRecord(record *BlockchainRecord) (*TransactionResult, error) {
	// Preparar dados para transação
	auth, err := e.createTransactor()
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	// Simular contrato (em produção, usar contrato real)
	data, err := e.encodeRecordData(record)
	if err != nil {
		return nil, fmt.Errorf("failed to encode record data: %w", err)
	}

	// Criar transação
	tx, err := e.sendTransaction(auth, data)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Aguardar confirmação
	receipt, err := e.waitForConfirmation(tx.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to confirm transaction: %w", err)
	}

	result := &TransactionResult{
		TxHash:      tx.Hash().Hex(),
		BlockNumber: receipt.BlockNumber.Uint64(),
		GasUsed:     receipt.GasUsed,
		Status:      e.getTransactionStatus(receipt),
		Timestamp:   time.Now(),
	}

	return result, nil
}

// StoreCertificate armazena informações de certificado na blockchain
func (e *EthereumService) StoreCertificate(certificate *models.Certificate) (*TransactionResult, error) {
	record := &BlockchainRecord{
		RecordType: "certificate",
		RecordID:   fmt.Sprintf("%d", certificate.ID),
		DataHash:   e.calculateCertificateHash(certificate),
		Metadata: map[string]interface{}{
			"serial_number": certificate.SerialNumber,
			"common_name":   certificate.CommonName,
			"issued_at":     certificate.IssuedAt,
			"expires_at":    certificate.ExpiresAt,
			"is_revoked":    certificate.IsRevoked,
		},
		Timestamp:   time.Now().Unix(),
		UserAddress: e.getPublicAddress(),
	}

	return e.StoreRecord(record)
}

// StoreSignature armazena assinatura digital na blockchain
func (e *EthereumService) StoreSignature(signature *models.DigitalSignature) (*TransactionResult, error) {
	record := &BlockchainRecord{
		RecordType: "signature",
		RecordID:   fmt.Sprintf("%d", signature.ID),
		DataHash:   signature.SignatureHash,
		Metadata: map[string]interface{}{
			"user_id":        signature.UserID.String(),
			"document_id":    signature.DocumentID.String(),
			"certificate_id": fmt.Sprintf("%d", signature.CertificateID),
			"document_hash":  signature.DocumentHash,
			"algorithm":      signature.Algorithm,
			"timestamp":      signature.Timestamp,
			"location":       signature.Location,
		},
		Timestamp:   signature.Timestamp.Unix(),
		UserAddress: e.getPublicAddress(),
	}

	return e.StoreRecord(record)
}

// StoreAuditLog armazena log de auditoria na blockchain
func (e *EthereumService) StoreAuditLog(auditLog *models.AuditLog) (*TransactionResult, error) {
	record := &BlockchainRecord{
		RecordType: "audit",
		RecordID:   fmt.Sprintf("%d", auditLog.ID),
		DataHash:   e.calculateAuditHash(auditLog),
		Metadata: map[string]interface{}{
			"user_id":     auditLog.UserID,
			"action":      auditLog.Action,
			"resource":    auditLog.Resource,
			"resource_id": auditLog.ResourceID,
			"success":     auditLog.Success,
			"risk_level":  auditLog.RiskLevel,
			"ip_address":  auditLog.IPAddress,
		},
		Timestamp:   auditLog.Timestamp.Unix(),
		UserAddress: e.getPublicAddress(),
	}

	return e.StoreRecord(record)
}

// VerifyRecord verifica a integridade de um registro na blockchain
func (e *EthereumService) VerifyRecord(txHash string, expectedHash string) (bool, error) {
	// Buscar transação
	tx, _, err := e.client.TransactionByHash(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return false, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Decodificar dados da transação
	record, err := e.decodeRecordData(tx.Data())
	if err != nil {
		return false, fmt.Errorf("failed to decode transaction data: %w", err)
	}

	// Verificar hash
	return record.DataHash == expectedHash, nil
}

// GetTransactionStatus obtém o status de uma transação
func (e *EthereumService) GetTransactionStatus(txHash string) (*TransactionResult, error) {
	hash := common.HexToHash(txHash)

	// Buscar recibo da transação
	receipt, err := e.client.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Buscar bloco para timestamp
	block, err := e.client.BlockByHash(context.Background(), receipt.BlockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	// Calcular confirmações
	currentBlock, err := e.client.BlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get current block number: %w", err)
	}

	confirmations := int(currentBlock - receipt.BlockNumber.Uint64())

	return &TransactionResult{
		TxHash:        txHash,
		BlockNumber:   receipt.BlockNumber.Uint64(),
		GasUsed:       receipt.GasUsed,
		Status:        e.getTransactionStatus(receipt),
		Confirmations: confirmations,
		Timestamp:     time.Unix(int64(block.Time()), 0),
	}, nil
}

// BatchStoreRecords armazena múltiplos registros em lote
func (e *EthereumService) BatchStoreRecords(records []*BlockchainRecord) ([]*TransactionResult, error) {
	results := make([]*TransactionResult, len(records))

	for i, record := range records {
		result, err := e.StoreRecord(record)
		if err != nil {
			return nil, fmt.Errorf("failed to store record %d: %w", i, err)
		}
		results[i] = result
	}

	return results, nil
}

// createTransactor cria um transactor para assinar transações
func (e *EthereumService) createTransactor() (*bind.TransactOpts, error) {
	publicKey := e.privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Obter nonce
	nonce, err := e.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(e.privateKey, e.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = e.gasLimit
	auth.GasPrice = e.gasPrice

	return auth, nil
}

// encodeRecordData codifica dados do registro para armazenamento
func (e *EthereumService) encodeRecordData(record *BlockchainRecord) ([]byte, error) {
	// Simular encoding (em produção, usar ABI encoding)
	data := fmt.Sprintf("%s:%s:%s:%d",
		record.RecordType,
		record.RecordID,
		record.DataHash,
		record.Timestamp,
	)
	return []byte(data), nil
}

// decodeRecordData decodifica dados do registro
func (e *EthereumService) decodeRecordData(data []byte) (*BlockchainRecord, error) {
	// Simular decoding (em produção, usar ABI decoding)
	return &BlockchainRecord{
		RecordType: "decoded",
		DataHash:   string(data),
	}, nil
}

// sendTransaction envia uma transação para a blockchain
func (e *EthereumService) sendTransaction(auth *bind.TransactOpts, data []byte) (*types.Transaction, error) {
	// Criar transação
	tx := types.NewTransaction(
		auth.Nonce.Uint64(),
		e.contractAddr,
		auth.Value,
		auth.GasLimit,
		auth.GasPrice,
		data,
	)

	// Assinar transação
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(e.chainID), e.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Enviar transação
	err = e.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx, nil
}

// waitForConfirmation aguarda confirmação da transação
func (e *EthereumService) waitForConfirmation(txHash common.Hash) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for confirmation")
		default:
			receipt, err := e.client.TransactionReceipt(context.Background(), txHash)
			if err != nil {
				if err == ethereum.NotFound {
					time.Sleep(2 * time.Second)
					continue
				}
				return nil, err
			}
			return receipt, nil
		}
	}
}

// getTransactionStatus determina o status da transação
func (e *EthereumService) getTransactionStatus(receipt *types.Receipt) string {
	if receipt.Status == 1 {
		return "success"
	}
	return "failed"
}

// getPublicAddress obtém o endereço público da chave privada
func (e *EthereumService) getPublicAddress() string {
	publicKey := e.privateKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	return crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
}

// calculateCertificateHash calcula hash de um certificado
func (e *EthereumService) calculateCertificateHash(cert *models.Certificate) string {
	data := fmt.Sprintf("%s:%s:%s:%s",
		cert.SerialNumber,
		cert.CommonName,
		cert.IssuedAt.String(),
		cert.ExpiresAt.String(),
	)
	return fmt.Sprintf("%x", crypto.Keccak256([]byte(data)))
}

// calculateAuditHash calcula hash de um log de auditoria
func (e *EthereumService) calculateAuditHash(audit *models.AuditLog) string {
	data := fmt.Sprintf("%s:%s:%s:%s:%t",
		audit.Action,
		audit.Resource,
		audit.IPAddress,
		audit.Timestamp.String(),
		audit.Success,
	)
	return fmt.Sprintf("%x", crypto.Keccak256([]byte(data)))
}

// GetGasPrice obtém preço atual do gas
func (e *EthereumService) GetGasPrice() (*big.Int, error) {
	return e.client.SuggestGasPrice(context.Background())
}

// EstimateGas estima gas necessário para uma transação
func (e *EthereumService) EstimateGas(data []byte) (uint64, error) {
	publicKey := e.privateKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	msg := ethereum.CallMsg{
		From: fromAddress,
		To:   &e.contractAddr,
		Data: data,
	}

	return e.client.EstimateGas(context.Background(), msg)
}

// Close fecha a conexão com o cliente Ethereum
func (e *EthereumService) Close() {
	e.client.Close()
}
