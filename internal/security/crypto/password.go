package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// PasswordService gerencia hash e verificação de senhas
type PasswordService struct {
	defaultAlgorithm string
	bcryptCost       int
	argon2Config     *Argon2Config
}

// Argon2Config configuração para Argon2
type Argon2Config struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

// NewPasswordService cria um novo serviço de senhas
func NewPasswordService(defaultAlgorithm string, bcryptCost int) *PasswordService {
	return &PasswordService{
		defaultAlgorithm: defaultAlgorithm,
		bcryptCost:       bcryptCost,
		argon2Config: &Argon2Config{
			Time:    3,
			Memory:  64 * 1024, // 64MB
			Threads: 4,
			KeyLen:  32,
			SaltLen: 16,
		},
	}
}

// HashedPassword representa uma senha com hash
type HashedPassword struct {
	Hash      string `json:"hash"`
	Salt      string `json:"salt,omitempty"`
	Algorithm string `json:"algorithm"`
	Params    string `json:"params,omitempty"`
}

// HashPassword cria hash de uma senha
func (p *PasswordService) HashPassword(password string) (*HashedPassword, error) {
	switch p.defaultAlgorithm {
	case "bcrypt":
		return p.hashWithBcrypt(password)
	case "argon2":
		return p.hashWithArgon2(password)
	default:
		return p.hashWithBcrypt(password) // Default para bcrypt
	}
}

// VerifyPassword verifica se uma senha corresponde ao hash
func (p *PasswordService) VerifyPassword(password string, hashed *HashedPassword) (bool, error) {
	switch hashed.Algorithm {
	case "bcrypt":
		return p.verifyBcrypt(password, hashed.Hash)
	case "argon2":
		return p.verifyArgon2(password, hashed)
	default:
		return false, fmt.Errorf("unsupported algorithm: %s", hashed.Algorithm)
	}
}

// hashWithBcrypt cria hash usando bcrypt
func (p *PasswordService) hashWithBcrypt(password string) (*HashedPassword, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), p.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password with bcrypt: %w", err)
	}

	return &HashedPassword{
		Hash:      string(hash),
		Algorithm: "bcrypt",
		Params:    fmt.Sprintf("cost=%d", p.bcryptCost),
	}, nil
}

// verifyBcrypt verifica hash bcrypt
func (p *PasswordService) verifyBcrypt(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, fmt.Errorf("bcrypt verification error: %w", err)
	}
	return true, nil
}

// hashWithArgon2 cria hash usando Argon2
func (p *PasswordService) hashWithArgon2(password string) (*HashedPassword, error) {
	// Gerar salt aleatório
	salt := make([]byte, p.argon2Config.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Gerar hash
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		p.argon2Config.Time,
		p.argon2Config.Memory,
		p.argon2Config.Threads,
		p.argon2Config.KeyLen,
	)

	// Codificar salt e hash em base64
	saltEncoded := encodeToString(salt)
	hashEncoded := encodeToString(hash)

	params := fmt.Sprintf("t=%d,m=%d,p=%d,keylen=%d",
		p.argon2Config.Time,
		p.argon2Config.Memory,
		p.argon2Config.Threads,
		p.argon2Config.KeyLen,
	)

	return &HashedPassword{
		Hash:      hashEncoded,
		Salt:      saltEncoded,
		Algorithm: "argon2",
		Params:    params,
	}, nil
}

// verifyArgon2 verifica hash Argon2
func (p *PasswordService) verifyArgon2(password string, hashed *HashedPassword) (bool, error) {
	// Decodificar salt
	salt, err := decodeFromString(hashed.Salt)
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	// Parse dos parâmetros
	config, err := p.parseArgon2Params(hashed.Params)
	if err != nil {
		return false, fmt.Errorf("failed to parse Argon2 params: %w", err)
	}

	// Gerar hash com a mesma configuração
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Time,
		config.Memory,
		config.Threads,
		config.KeyLen,
	)

	// Decodificar hash armazenado
	storedHash, err := decodeFromString(hashed.Hash)
	if err != nil {
		return false, fmt.Errorf("failed to decode stored hash: %w", err)
	}

	// Comparar de forma segura
	return subtle.ConstantTimeCompare(computedHash, storedHash) == 1, nil
}

// parseArgon2Params faz parse dos parâmetros Argon2
func (p *PasswordService) parseArgon2Params(params string) (*Argon2Config, error) {
	config := &Argon2Config{}

	parts := strings.Split(params, ",")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "t":
			val, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid time parameter: %w", err)
			}
			config.Time = uint32(val)
		case "m":
			val, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid memory parameter: %w", err)
			}
			config.Memory = uint32(val)
		case "p":
			val, err := strconv.ParseUint(value, 10, 8)
			if err != nil {
				return nil, fmt.Errorf("invalid threads parameter: %w", err)
			}
			config.Threads = uint8(val)
		case "keylen":
			val, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid keylen parameter: %w", err)
			}
			config.KeyLen = uint32(val)
		}
	}

	return config, nil
}

// NeedsRehash verifica se uma senha precisa de rehash
func (p *PasswordService) NeedsRehash(hashed *HashedPassword) bool {
	// Se o algoritmo mudou
	if hashed.Algorithm != p.defaultAlgorithm {
		return true
	}

	switch hashed.Algorithm {
	case "bcrypt":
		// Verificar se o custo mudou
		return p.needsBcryptRehash(hashed)
	case "argon2":
		// Verificar se os parâmetros mudaram
		return p.needsArgon2Rehash(hashed)
	}

	return false
}

// needsBcryptRehash verifica se bcrypt precisa rehash
func (p *PasswordService) needsBcryptRehash(hashed *HashedPassword) bool {
	// Extrair custo atual do hash
	currentCost, err := bcrypt.Cost([]byte(hashed.Hash))
	if err != nil {
		return true // Se não conseguir extrair, rehash
	}

	return currentCost != p.bcryptCost
}

// needsArgon2Rehash verifica se Argon2 precisa rehash
func (p *PasswordService) needsArgon2Rehash(hashed *HashedPassword) bool {
	config, err := p.parseArgon2Params(hashed.Params)
	if err != nil {
		return true
	}

	return config.Time != p.argon2Config.Time ||
		config.Memory != p.argon2Config.Memory ||
		config.Threads != p.argon2Config.Threads ||
		config.KeyLen != p.argon2Config.KeyLen
}

// GenerateRandomPassword gera uma senha aleatória
func (p *PasswordService) GenerateRandomPassword(length int, includeSpecial bool) (string, error) {
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numbers   = "0123456789"
		special   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	charset := lowercase + uppercase + numbers
	if includeSpecial {
		charset += special
	}

	password := make([]byte, length)
	for i := range password {
		randomIndex := make([]byte, 1)
		if _, err := rand.Read(randomIndex); err != nil {
			return "", fmt.Errorf("failed to generate random password: %w", err)
		}
		password[i] = charset[randomIndex[0]%byte(len(charset))]
	}

	return string(password), nil
}

// CheckPasswordStrength verifica a força de uma senha
func (p *PasswordService) CheckPasswordStrength(password string) *PasswordStrength {
	strength := &PasswordStrength{
		Password:    password,
		Length:      len(password),
		Score:       0,
		Level:       "Very Weak",
		Issues:      []string{},
		Suggestions: []string{},
	}

	// Verificar comprimento
	if strength.Length < 8 {
		strength.Issues = append(strength.Issues, "Password is too short")
		strength.Suggestions = append(strength.Suggestions, "Use at least 8 characters")
	} else if strength.Length >= 8 {
		strength.Score += 1
	}

	if strength.Length >= 12 {
		strength.Score += 1
	}

	// Verificar variedade de caracteres
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasNumber := strings.ContainsAny(password, "0123456789")
	hasSpecial := strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?")

	if hasLower {
		strength.Score += 1
	} else {
		strength.Issues = append(strength.Issues, "Missing lowercase letters")
		strength.Suggestions = append(strength.Suggestions, "Add lowercase letters (a-z)")
	}

	if hasUpper {
		strength.Score += 1
	} else {
		strength.Issues = append(strength.Issues, "Missing uppercase letters")
		strength.Suggestions = append(strength.Suggestions, "Add uppercase letters (A-Z)")
	}

	if hasNumber {
		strength.Score += 1
	} else {
		strength.Issues = append(strength.Issues, "Missing numbers")
		strength.Suggestions = append(strength.Suggestions, "Add numbers (0-9)")
	}

	if hasSpecial {
		strength.Score += 1
	} else {
		strength.Issues = append(strength.Issues, "Missing special characters")
		strength.Suggestions = append(strength.Suggestions, "Add special characters (!@#$%^&*)")
	}

	// Verificar padrões comuns
	if p.isCommonPassword(password) {
		strength.Score = 0
		strength.Issues = append(strength.Issues, "Password is too common")
		strength.Suggestions = append(strength.Suggestions, "Avoid common passwords")
	}

	// Determinar nível baseado no score
	switch {
	case strength.Score >= 6:
		strength.Level = "Very Strong"
	case strength.Score >= 5:
		strength.Level = "Strong"
	case strength.Score >= 4:
		strength.Level = "Moderate"
	case strength.Score >= 2:
		strength.Level = "Weak"
	default:
		strength.Level = "Very Weak"
	}

	return strength
}

// PasswordStrength representa a análise de força de uma senha
type PasswordStrength struct {
	Password    string   `json:"password"`
	Length      int      `json:"length"`
	Score       int      `json:"score"`
	Level       string   `json:"level"`
	Issues      []string `json:"issues"`
	Suggestions []string `json:"suggestions"`
}

// isCommonPassword verifica se é uma senha comum
func (p *PasswordService) isCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "123456", "password123", "admin", "qwerty",
		"letmein", "welcome", "123456789", "password1", "abc123",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}

	return false
}

// Funções auxiliares para encoding
func encodeToString(data []byte) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	encoded := make([]byte, 0, ((len(data)+2)/3)*4)

	for i := 0; i < len(data); i += 3 {
		chunk := make([]byte, 4)

		chunk[0] = charset[data[i]>>2]

		if i+1 < len(data) {
			chunk[1] = charset[((data[i]&0x03)<<4)|(data[i+1]>>4)]

			if i+2 < len(data) {
				chunk[2] = charset[((data[i+1]&0x0F)<<2)|(data[i+2]>>6)]
				chunk[3] = charset[data[i+2]&0x3F]
			} else {
				chunk[2] = charset[(data[i+1]&0x0F)<<2]
				chunk[3] = '='
			}
		} else {
			chunk[1] = charset[(data[i]&0x03)<<4]
			chunk[2] = '='
			chunk[3] = '='
		}

		encoded = append(encoded, chunk...)
	}

	return string(encoded)
}

func decodeFromString(encoded string) ([]byte, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	// Simular decode (implementação básica)
	decoded := make([]byte, len(encoded)/4*3)
	return decoded, nil // TODO: implementar decode completo
}
