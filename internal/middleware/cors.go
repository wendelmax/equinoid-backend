package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS configura middleware de CORS
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowedOrigins := getAllowedOrigins()

		// Verificar se a origem está permitida
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}

		// Se não estiver na lista, permitir localhost para desenvolvimento
		if !isAllowed && (origin == "" ||
			(len(origin) > 16 && origin[:16] == "http://localhost") ||
			(len(origin) > 17 && origin[:17] == "https://localhost")) {
			isAllowed = true
		}

		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 horas

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CORSConfig configuração customizável para CORS
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// CORSWithConfig middleware CORS com configuração customizada
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Verificar origem permitida
		isAllowed := false
		for _, allowedOrigin := range config.AllowedOrigins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				isAllowed = true
				break
			}
		}

		if isAllowed || len(config.AllowedOrigins) == 0 {
			if len(config.AllowedOrigins) > 0 && config.AllowedOrigins[0] == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else if isAllowed {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		// Methods
		if len(config.AllowedMethods) > 0 {
			methods := ""
			for i, method := range config.AllowedMethods {
				if i > 0 {
					methods += ", "
				}
				methods += method
			}
			c.Header("Access-Control-Allow-Methods", methods)
		}

		// Headers
		if len(config.AllowedHeaders) > 0 {
			headers := ""
			for i, header := range config.AllowedHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += header
			}
			c.Header("Access-Control-Allow-Headers", headers)
		}

		// Credentials
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Max Age
		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
		}

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// DefaultCORSConfig retorna uma configuração padrão para CORS
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: getAllowedOrigins(),
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"accept",
			"origin",
			"Cache-Control",
			"X-Requested-With",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 horas
	}
}

// getAllowedOrigins retorna a lista de origens permitidas
// Pode ser configurada via variável de ambiente CORS_ALLOWED_ORIGINS (separada por vírgula)
// ou usa a lista padrão
func getAllowedOrigins() []string {
	envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if envOrigins != "" {
		origins := strings.Split(envOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		return origins
	}

	return []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:5173",
		"https://equinoid.com",
		"https://www.equinoid.com",
		"https://app.equinoid.com",
		"https://app.equinoid.org",
		"https://equinoid.org",
	}
}
