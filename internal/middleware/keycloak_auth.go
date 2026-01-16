package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/equinoid/backend/internal/models"
)

type KeycloakAuth struct {
	verifier *oidc.IDTokenVerifier
	provider *oidc.Provider
	db       *gorm.DB
	logger   *logrus.Logger
	mu       sync.RWMutex
}

func NewKeycloakAuth(keycloakURL, realm, clientID string, db *gorm.DB, logger *logrus.Logger) (*KeycloakAuth, error) {
	ctx := context.Background()

	issuer := keycloakURL + "/realms/" + realm
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{
		SkipClientIDCheck: true,
	})

	return &KeycloakAuth{
		verifier: verifier,
		provider: provider,
		db:       db,
		logger:   logger,
	}, nil
}

func (k *KeycloakAuth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format. Use: Bearer <token>"})
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		idToken, err := k.verifier.Verify(ctx, tokenString)
		if err != nil {
			k.logger.WithError(err).Warn("Failed to verify Keycloak token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		var claims struct {
			Sub               string `json:"sub"`
			Email             string `json:"email"`
			EmailVerified     bool   `json:"email_verified"`
			Name              string `json:"name"`
			PreferredUsername string `json:"preferred_username"`
			GivenName         string `json:"given_name"`
			FamilyName        string `json:"family_name"`
		}

		if err := idToken.Claims(&claims); err != nil {
			k.logger.WithError(err).Error("Failed to parse token claims")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token claims"})
			c.Abort()
			return
		}

		user, err := k.syncUser(ctx, claims.Sub, claims.Email, claims.Name)
		if err != nil {
			k.logger.WithError(err).Error("Failed to sync user from Keycloak")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync user"})
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("keycloak_sub", claims.Sub)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Set("email_verified", claims.EmailVerified)

		k.logger.WithFields(logrus.Fields{
			"user_id":    user.ID,
			"email":      claims.Email,
			"keycloak":   true,
			"request_id": c.GetString("request_id"),
		}).Debug("Keycloak authentication successful")

		c.Next()
	}
}

func (k *KeycloakAuth) syncUser(ctx context.Context, keycloakSub, email, name string) (*models.User, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	var user models.User

	err := k.db.WithContext(ctx).Where("keycloak_sub = ?", keycloakSub).First(&user).Error
	if err == nil {
		if user.Email != email || user.Name != name {
			user.Email = email
			user.Name = name
			if err := k.db.WithContext(ctx).Save(&user).Error; err != nil {
				k.logger.WithError(err).Warn("Failed to update user from Keycloak")
			}
		}
		return &user, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	err = k.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err == nil {
		user.KeycloakSub = keycloakSub
		if err := k.db.WithContext(ctx).Save(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	user = models.User{
		Email:           email,
		Name:            name,
		KeycloakSub:     keycloakSub,
		IsEmailVerified: true,
		IsActive:        true,
		Role:            "usuario",
		UserType:        "criador",
	}

	if err := k.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}

	k.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   email,
	}).Info("New user created from Keycloak")

	return &user, nil
}

func (k *KeycloakAuth) GetProvider() *oidc.Provider {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.provider
}
