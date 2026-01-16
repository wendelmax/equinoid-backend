package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	EquinoIdVersion = "1"
	EquinoIdPrefix  = "EQ"
	EquinoIdLength  = 24

	EquinoIdNamespace = "a3f5d8e7-1c4b-4a9e-8f2d-6b3c5e7a9d1f"
)

// GenerateEquinoId gera um EquinoId único e compacto
// Formato: EQ + VER(1) + PAIS(3) + UUID4(6) + UUID5(8) + CRC16(4) = 24 chars
// Exemplo: EQ1BRA3F8A2C5D7E9B1F2A4C (24 caracteres)
func GenerateEquinoId(paisOrigem, microchipID, nome string, dataNascimento time.Time, sexo, pelagem, raca string) (string, error) {
	if len(paisOrigem) != 3 {
		return "", errors.New("paisOrigem deve ter exatamente 3 caracteres")
	}
	if microchipID == "" {
		return "", errors.New("microchipID é obrigatório")
	}
	if nome == "" {
		return "", errors.New("nome é obrigatório")
	}

	paisOrigem = strings.ToUpper(paisOrigem)

	uuidv4 := uuid.New().String()
	uuidv4Part := strings.ToUpper(strings.ReplaceAll(uuidv4, "-", "")[:6])

	namespace := uuid.MustParse(EquinoIdNamespace)
	dataForUUIDv5 := fmt.Sprintf("%s|%s|%s|%s|%s",
		microchipID,
		dataNascimento.Format("2006-01-02"),
		sexo,
		raca,
		nome,
	)
	uuidv5 := uuid.NewSHA1(namespace, []byte(dataForUUIDv5)).String()
	uuidv5Part := strings.ToUpper(strings.ReplaceAll(uuidv5, "-", "")[:8])

	baseId := fmt.Sprintf("%s%s%s%s%s", EquinoIdPrefix, EquinoIdVersion, paisOrigem, uuidv4Part, uuidv5Part)

	checksum := calculateCRC16Checksum(baseId)

	equinoId := fmt.Sprintf("%s%s", baseId, checksum)

	return equinoId, nil
}

// calculateCRC16Checksum calcula um checksum CRC-16 (4 caracteres hex)
func calculateCRC16Checksum(data string) string {
	crc := crc32.ChecksumIEEE([]byte(data))
	return strings.ToUpper(fmt.Sprintf("%04X", crc&0xFFFF))
}

// calculateSHA256Checksum alternativa com SHA256 (4 chars)
func calculateSHA256Checksum(data string) string {
	hash := sha256.Sum256([]byte(data))
	return strings.ToUpper(hex.EncodeToString(hash[:])[:4])
}

// ValidateEquinoIdFormat valida se um EquinoId está no formato correto
func ValidateEquinoIdFormat(equinoId string) error {
	if len(equinoId) != EquinoIdLength {
		return fmt.Errorf("EquinoId deve ter %d caracteres, tem %d", EquinoIdLength, len(equinoId))
	}

	if !strings.HasPrefix(equinoId, EquinoIdPrefix) {
		return fmt.Errorf("EquinoId deve começar com %s", EquinoIdPrefix)
	}

	version := string(equinoId[2])
	if version != EquinoIdVersion {
		return fmt.Errorf("versão inválida: %s (esperado: %s)", version, EquinoIdVersion)
	}

	hexPart := equinoId[6:]
	if _, err := hex.DecodeString(hexPart); err != nil {
		return errors.New("EquinoId contém caracteres inválidos")
	}

	return nil
}

// ValidateEquinoIdChecksum valida o checksum de um EquinoId
func ValidateEquinoIdChecksum(equinoId string) error {
	if err := ValidateEquinoIdFormat(equinoId); err != nil {
		return err
	}

	baseId := equinoId[:20]
	providedChecksum := equinoId[20:]

	calculatedChecksum := calculateCRC16Checksum(baseId)

	if providedChecksum != calculatedChecksum {
		return errors.New("checksum inválido")
	}

	return nil
}

// ParseEquinoId extrai informações do EquinoId
func ParseEquinoId(equinoId string) (map[string]string, error) {
	if err := ValidateEquinoIdFormat(equinoId); err != nil {
		return nil, err
	}

	return map[string]string{
		"prefix":   equinoId[:2],
		"version":  equinoId[2:3],
		"pais":     equinoId[3:6],
		"uuid4":    equinoId[6:12],
		"uuid5":    equinoId[12:20],
		"checksum": equinoId[20:24],
		"full":     equinoId,
	}, nil
}
