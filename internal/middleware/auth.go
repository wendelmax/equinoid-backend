package middleware

import (
	"context"
	"net/http"

	"github.com/equinoid/backend/pkg/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware middleware de autenticação JWT
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	jwtService := auth.NewJWTService(jwtSecret, "EquinoId", 0) // expiry será definido nas configurações

	return func(c *gin.Context) {
		// Extrair token do header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
			})
			c.Abort()
			return
		}

		token := auth.ExtractTokenFromHeader(authHeader)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Bearer token required",
			})
			c.Abort()
			return
		}

		// Validar token
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			status := http.StatusUnauthorized
			message := "Invalid token"

			if err == auth.ErrExpiredToken {
				status = http.StatusUnauthorized
				message = "Token has expired"
			}

			c.JSON(status, gin.H{
				"success": false,
				"error":   message,
			})
			c.Abort()
			return
		}

		// Adicionar claims ao contexto
		ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "user_type", claims.UserType)
		ctx = context.WithValue(ctx, "jwt_claims", claims)

		c.Request = c.Request.WithContext(ctx)

		// Adicionar claims ao contexto do Gin para fácil acesso
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_type", claims.UserType)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// OptionalAuthMiddleware middleware de autenticação opcional (não falha se não houver token)
func OptionalAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	jwtService := auth.NewJWTService(jwtSecret, "EquinoId", 0)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		token := auth.ExtractTokenFromHeader(authHeader)
		if token == "" {
			c.Next()
			return
		}

		// Tentar validar token
		claims, err := jwtService.ValidateAccessToken(token)
		if err == nil {
			// Token válido - adicionar ao contexto
			ctx := context.WithValue(c.Request.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "user_email", claims.Email)
			ctx = context.WithValue(ctx, "user_type", claims.UserType)
			ctx = context.WithValue(ctx, "jwt_claims", claims)

			c.Request = c.Request.WithContext(ctx)

			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_type", claims.UserType)
			c.Set("jwt_claims", claims)
		}

		c.Next()
	}
}

// RequireRoleMiddleware middleware que requer role específico
func RequireRoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authentication required",
			})
			c.Abort()
			return
		}

		userTypeStr := string(userType.(string))

		// Verificar se o tipo de usuário está nas roles permitidas
		allowed := false
		for _, role := range allowedRoles {
			if userTypeStr == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOwnershipMiddleware middleware que requer propriedade do recurso
func RequireOwnershipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Esta implementação será específica para cada endpoint
		// Por enquanto, apenas passar adiante
		c.Next()
	}
}

// GetUserIDFromContext obtém o ID do usuário do contexto
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uint); ok {
			return uid, true
		}
	}
	return 0, false
}

// GetUserTypeFromContext obtém o tipo do usuário do contexto
func GetUserTypeFromContext(c *gin.Context) (string, bool) {
	if userType, exists := c.Get("user_type"); exists {
		if ut, ok := userType.(string); ok {
			return ut, true
		}
	}
	return "", false
}

// GetJWTClaimsFromContext obtém as claims JWT do contexto
func GetJWTClaimsFromContext(c *gin.Context) (*auth.Claims, bool) {
	if claims, exists := c.Get("jwt_claims"); exists {
		if jwtClaims, ok := claims.(*auth.Claims); ok {
			return jwtClaims, true
		}
	}
	return nil, false
}

// IsAuthenticated verifica se o usuário está autenticado
func IsAuthenticated(c *gin.Context) bool {
	_, exists := GetUserIDFromContext(c)
	return exists
}

// IsVeterinario verifica se o usuário é veterinário
func IsVeterinario(c *gin.Context) bool {
	userType, exists := GetUserTypeFromContext(c)
	return exists && userType == "veterinario"
}

// IsCriador verifica se o usuário é criador
func IsCriador(c *gin.Context) bool {
	userType, exists := GetUserTypeFromContext(c)
	return exists && userType == "criador"
}

// IsAdmin verifica se o usuário é administrador
func IsAdmin(c *gin.Context) bool {
	userType, exists := GetUserTypeFromContext(c)
	return exists && userType == "admin"
}

// CanAccessEquino verifica se o usuário pode acessar um equino específico
func CanAccessEquino(c *gin.Context, equinoOwnerID uint) bool {
	userID, authenticated := GetUserIDFromContext(c)
	if !authenticated {
		return false
	}

	// Administradores e veterinários podem acessar qualquer equino
	if IsAdmin(c) || IsVeterinario(c) {
		return true
	}

	// Criadores só podem acessar seus próprios equinos
	if IsCriador(c) {
		return userID == equinoOwnerID
	}

	return false
}

// RequireVeterinarioMiddleware middleware que requer usuário veterinário
func RequireVeterinarioMiddleware() gin.HandlerFunc {
	return RequireRoleMiddleware("veterinario")
}

// RequireCriadorMiddleware middleware que requer usuário criador
func RequireCriadorMiddleware() gin.HandlerFunc {
	return RequireRoleMiddleware("criador")
}

// RequireAdminMiddleware middleware que requer usuário administrador
func RequireAdminMiddleware() gin.HandlerFunc {
	return RequireRoleMiddleware("admin")
}

// RequireVeterinarioOrAdminMiddleware middleware que requer veterinário ou admin
func RequireVeterinarioOrAdminMiddleware() gin.HandlerFunc {
	return RequireRoleMiddleware("veterinario", "admin")
}

// RequireCriadorOrAdminMiddleware middleware que requer criador ou admin
func RequireCriadorOrAdminMiddleware() gin.HandlerFunc {
	return RequireRoleMiddleware("criador", "admin")
}
