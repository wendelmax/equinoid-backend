package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/equinoid/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SupabaseAuth struct {
	projectURL string
	publicKeys map[string]*rsa.PublicKey
	db         *gorm.DB
	logger     *logrus.Logger
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func NewSupabaseAuth(projectURL string, db *gorm.DB, logger *logrus.Logger) (*SupabaseAuth, error) {
	sa := &SupabaseAuth{
		projectURL: projectURL,
		publicKeys: make(map[string]*rsa.PublicKey),
		db:         db,
		logger:     logger,
	}

	if err := sa.fetchJWKS(); err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	return sa, nil
}

func (sa *SupabaseAuth) fetchJWKS() error {
	jwksURL := fmt.Sprintf("%s/auth/v1/jwks", sa.projectURL)

	resp, err := http.Get(jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	for _, jwk := range jwks.Keys {
		publicKey, err := sa.jwkToRSAPublicKey(jwk)
		if err != nil {
			sa.logger.Warnf("Failed to parse JWK: %v", err)
			continue
		}
		sa.publicKeys[jwk.Kid] = publicKey
	}

	if len(sa.publicKeys) == 0 {
		return errors.New("no valid public keys found in JWKS")
	}

	return nil
}

func (sa *SupabaseAuth) jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, err
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

func (sa *SupabaseAuth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, errors.New("missing kid in token header")
			}

			publicKey, exists := sa.publicKeys[kid]
			if !exists {
				if err := sa.fetchJWKS(); err != nil {
					return nil, err
				}

				publicKey, exists = sa.publicKeys[kid]
				if !exists {
					return nil, fmt.Errorf("public key not found for kid: %s", kid)
				}
			}

			return publicKey, nil
		})

		if err != nil || !token.Valid {
			sa.logger.Warnf("Invalid Supabase JWT: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		iss, ok := claims["iss"].(string)
		if !ok || !strings.HasPrefix(iss, sa.projectURL) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token issuer"})
			c.Abort()
			return
		}

		supabaseUserID, ok := claims["sub"].(string)
		if !ok || supabaseUserID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		email, _ := claims["email"].(string)

		user, err := sa.syncUser(supabaseUserID, email, claims)
		if err != nil {
			sa.logger.Errorf("Failed to sync user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync user"})
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("supabase_id", supabaseUserID)
		c.Set("email", email)
		c.Set("user", user)

		sa.logger.Infof("User authenticated via Supabase: %s (local ID: %d)", email, user.ID)

		c.Next()
	}
}

func (sa *SupabaseAuth) syncUser(supabaseUserID, email string, claims jwt.MapClaims) (*models.User, error) {
	var user models.User

	result := sa.db.Where("supabase_id = ?", supabaseUserID).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			user = models.User{
				SupabaseID: supabaseUserID,
				Email:      email,
				Name:       email,
				UserType:   models.UserTypeCriador,
				IsActive:   true,
			}

			if userMetadata, ok := claims["user_metadata"].(map[string]interface{}); ok {
				if nome, ok := userMetadata["full_name"].(string); ok {
					user.Name = nome
				}
			}

			if err := sa.db.Create(&user).Error; err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}

			sa.logger.Infof("Created new user from Supabase: %s (ID: %d)", email, user.ID)
		} else {
			return nil, result.Error
		}
	} else {
		if user.Email != email {
			user.Email = email
			if err := sa.db.Save(&user).Error; err != nil {
				sa.logger.Warnf("Failed to update user email: %v", err)
			}
		}
	}

	return &user, nil
}

func (sa *SupabaseAuth) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		sa.AuthMiddleware()(c)
	}
}
