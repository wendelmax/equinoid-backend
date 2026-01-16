package auth

import (
	"errors"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

// JWTService gerencia tokens JWT
type JWTService struct {
	secretKey []byte
	issuer    string
	expiry    time.Duration
}

// Claims representa as claims do JWT
type Claims struct {
	UserID   uint            `json:"user_id"`
	Email    string          `json:"email"`
	UserType models.UserType `json:"user_type"`
	jwt.RegisteredClaims
}

// TokenPair representa um par de tokens (access + refresh)
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// RefreshClaims representa as claims do refresh token
type RefreshClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidClaims    = errors.New("invalid token claims")
	ErrTokenNotFound    = errors.New("token not found")
	ErrInvalidTokenType = errors.New("invalid token type")
)

// NewJWTService cria um novo serviço JWT
func NewJWTService(secretKey string, issuer string, expiry time.Duration) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
		issuer:    issuer,
		expiry:    expiry,
	}
}

// GenerateTokenPair gera um par de tokens para um usuário
func (j *JWTService) GenerateTokenPair(user *models.User) (*TokenPair, error) {
	// Gerar access token
	accessToken, err := j.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// Gerar refresh token
	refreshToken, err := j.GenerateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(j.expiry.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// GenerateAccessToken gera um access token
func (j *JWTService) GenerateAccessToken(user *models.User) (string, error) {
	now := time.Now()

	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		UserType: user.UserType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Email,
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiry)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        generateJTI(user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateRefreshToken gera um refresh token
func (j *JWTService) GenerateRefreshToken(user *models.User) (string, error) {
	now := time.Now()
	refreshExpiry := time.Hour * 24 * 30 // 30 dias

	claims := &RefreshClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Email,
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        generateJTI(user.ID) + "_refresh",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateAccessToken valida um access token
func (j *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return j.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// ValidateRefreshToken valida um refresh token
func (j *JWTService) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return j.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// RefreshTokenPair renova um par de tokens usando refresh token
func (j *JWTService) RefreshTokenPair(refreshTokenString string, user *models.User) (*TokenPair, error) {
	// Validar refresh token
	refreshClaims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	// Verificar se o refresh token pertence ao usuário
	if refreshClaims.UserID != user.ID {
		return nil, ErrInvalidClaims
	}

	// Gerar novo par de tokens
	return j.GenerateTokenPair(user)
}

// ExtractTokenFromHeader extrai o token do header Authorization
func ExtractTokenFromHeader(authHeader string) string {
	const bearerPrefix = "Bearer "
	if len(authHeader) > len(bearerPrefix) && authHeader[:len(bearerPrefix)] == bearerPrefix {
		return authHeader[len(bearerPrefix):]
	}
	return ""
}

// generateJTI gera um JWT ID único
func generateJTI(userID uint) string {
	return jwt.NewNumericDate(time.Now()).String() + "_" + string(rune(userID))
}

// TokenBlacklist gerencia tokens revogados/blacklistados
type TokenBlacklist interface {
	AddToBlacklist(tokenID string, expiry time.Time) error
	IsBlacklisted(tokenID string) (bool, error)
	CleanupExpired() error
}

// MemoryTokenBlacklist implementação em memória da blacklist (não recomendado para produção)
type MemoryTokenBlacklist struct {
	tokens map[string]time.Time
}

// NewMemoryTokenBlacklist cria uma nova blacklist em memória
func NewMemoryTokenBlacklist() *MemoryTokenBlacklist {
	return &MemoryTokenBlacklist{
		tokens: make(map[string]time.Time),
	}
}

// AddToBlacklist adiciona um token à blacklist
func (m *MemoryTokenBlacklist) AddToBlacklist(tokenID string, expiry time.Time) error {
	m.tokens[tokenID] = expiry
	return nil
}

// IsBlacklisted verifica se um token está na blacklist
func (m *MemoryTokenBlacklist) IsBlacklisted(tokenID string) (bool, error) {
	expiry, exists := m.tokens[tokenID]
	if !exists {
		return false, nil
	}

	// Se o token expirou, remove da blacklist
	if time.Now().After(expiry) {
		delete(m.tokens, tokenID)
		return false, nil
	}

	return true, nil
}

// CleanupExpired remove tokens expirados da blacklist
func (m *MemoryTokenBlacklist) CleanupExpired() error {
	now := time.Now()
	for tokenID, expiry := range m.tokens {
		if now.After(expiry) {
			delete(m.tokens, tokenID)
		}
	}
	return nil
}

// PasswordHasher interface para hash de senhas
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
}

// BcryptHasher implementação com bcrypt
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher cria um novo hasher bcrypt
func NewBcryptHasher(cost int) *BcryptHasher {
	return &BcryptHasher{cost: cost}
}

// HashPassword gera hash da senha
func (b *BcryptHasher) HashPassword(password string) (string, error) {
	// Para compatibilidade, vamos usar a implementação do Go crypto
	return hashPassword(password, b.cost)
}

// CheckPassword verifica se a senha confere com o hash
func (b *BcryptHasher) CheckPassword(password, hash string) bool {
	return checkPassword(password, hash)
}

// AuthService combina JWT e password hashing
type AuthService struct {
	jwt       *JWTService
	hasher    PasswordHasher
	blacklist TokenBlacklist
}

// NewAuthService cria um novo serviço de autenticação
func NewAuthService(jwt *JWTService, hasher PasswordHasher, blacklist TokenBlacklist) *AuthService {
	return &AuthService{
		jwt:       jwt,
		hasher:    hasher,
		blacklist: blacklist,
	}
}

// GenerateTokenPair gera tokens para um usuário
func (a *AuthService) GenerateTokenPair(user *models.User) (*TokenPair, error) {
	return a.jwt.GenerateTokenPair(user)
}

// ValidateToken valida um access token
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	claims, err := a.jwt.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Verificar blacklist
	if a.blacklist != nil {
		blacklisted, err := a.blacklist.IsBlacklisted(claims.ID)
		if err != nil {
			return nil, err
		}
		if blacklisted {
			return nil, ErrInvalidToken
		}
	}

	return claims, nil
}

// RefreshToken renova tokens
func (a *AuthService) RefreshToken(refreshToken string, user *models.User) (*TokenPair, error) {
	return a.jwt.RefreshTokenPair(refreshToken, user)
}

// RevokeToken adiciona token à blacklist
func (a *AuthService) RevokeToken(claims *Claims) error {
	if a.blacklist == nil {
		return nil // Blacklist não configurada
	}

	return a.blacklist.AddToBlacklist(claims.ID, claims.ExpiresAt.Time)
}

// HashPassword gera hash de senha
func (a *AuthService) HashPassword(password string) (string, error) {
	return a.hasher.HashPassword(password)
}

// CheckPassword verifica senha
func (a *AuthService) CheckPassword(password, hash string) bool {
	return a.hasher.CheckPassword(password, hash)
}

// AuthenticateUser autentica um usuário
func (a *AuthService) AuthenticateUser(user *models.User, password string) (*TokenPair, error) {
	if !a.CheckPassword(password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	return a.GenerateTokenPair(user)
}

func GenerateTokenPair(userID uint, email string, userType models.UserType, secret string, expireHours int) (*models.TokenPair, error) {
	jwtService := NewJWTService(secret, "equinoid", time.Duration(expireHours)*time.Hour)
	
	user := &models.User{
		ID:       userID,
		Email:    email,
		UserType: userType,
	}
	
	pair, err := jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}
	
	return &models.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		TokenType:    pair.TokenType,
	}, nil
}

func ValidateRefreshToken(tokenString, secret string) (*RefreshClaims, error) {
	jwtService := NewJWTService(secret, "equinoid", 24*time.Hour)
	return jwtService.ValidateRefreshToken(tokenString)
}

func GenerateResetToken(userID uint, secret string) (string, error) {
	jwtService := NewJWTService(secret, "equinoid", 1*time.Hour)
	
	user := &models.User{ID: userID}
	return jwtService.GenerateAccessToken(user)
}

func ValidateResetToken(tokenString, secret string) (*Claims, error) {
	jwtService := NewJWTService(secret, "equinoid", 1*time.Hour)
	return jwtService.ValidateAccessToken(tokenString)
}
