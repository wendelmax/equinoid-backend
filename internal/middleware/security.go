package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/equinoid/backend/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID middleware adiciona ID único para cada requisição
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar se já existe um request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Adicionar ao header de resposta
		c.Header("X-Request-ID", requestID)

		// Adicionar ao contexto
		c.Set("request_id", requestID)

		c.Next()
	}
}

// Logger middleware de logging estruturado
func Logger(logger *logging.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log estruturado
		logger.WithFields(logging.Fields{
			"method":      param.Method,
			"path":        param.Path,
			"status":      param.StatusCode,
			"duration_ms": param.Latency.Milliseconds(),
			"ip":          param.ClientIP,
			"user_agent":  param.Request.UserAgent(),
			"request_id":  param.Keys["request_id"],
		}).Info("HTTP Request")

		return ""
	})
}

// Recovery middleware de recuperação de panics
func Recovery(logger *logging.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Log do panic
		logger.WithFields(logging.Fields{
			"panic":      recovered,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"request_id": c.GetString("request_id"),
		}).Error("Panic recovered")

		// Resposta de erro
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error":      "Internal server error",
			"request_id": c.GetString("request_id"),
		})
	})
}

// RateLimit implementação simples de rate limiting
type RateLimiter struct {
	visitors map[string]*visitor
	mutex    sync.RWMutex
	rate     int
	window   time.Duration
}

type visitor struct {
	requests  int
	resetTime time.Time
}

// NewRateLimiter cria um novo rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	// Cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup remove entradas expiradas do rate limiter
func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mutex.Lock()
		for ip, v := range rl.visitors {
			if time.Now().After(v.resetTime) {
				delete(rl.visitors, ip)
			}
		}
		rl.mutex.Unlock()
	}
}

// isAllowed verifica se o IP pode fazer mais requisições
func (rl *RateLimiter) isAllowed(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			requests:  1,
			resetTime: time.Now().Add(rl.window),
		}
		return true
	}

	if time.Now().After(v.resetTime) {
		v.requests = 1
		v.resetTime = time.Now().Add(rl.window)
		return true
	}

	if v.requests >= rl.rate {
		return false
	}

	v.requests++
	return true
}

// RateLimit middleware de rate limiting
func RateLimit(rate int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, time.Minute)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.isAllowed(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":     false,
				"error":       "Too many requests",
				"retry_after": "60",
			})
			c.Header("Retry-After", "60")
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityHeaders adiciona headers de segurança
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// XSS Protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict Transport Security (HTTPS only)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self'"
		c.Header("Content-Security-Policy", csp)

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Feature Policy / Permissions Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// IPWhitelist middleware de whitelist de IPs
func IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	whitelist := make(map[string]bool)
	for _, ip := range allowedIPs {
		whitelist[ip] = true
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if len(whitelist) > 0 && !whitelist[clientIP] {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied from this IP",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserAgent middleware que bloqueia user agents suspeitos
func UserAgent() gin.HandlerFunc {
	// Lista de user agents suspeitos ou bloqueados
	blockedUserAgents := []string{
		"",
		"curl",
		"wget",
		"python-requests",
		"bot",
		"crawler",
		"scraper",
	}

	return func(c *gin.Context) {
		userAgent := c.Request.UserAgent()

		// Bloquear requisições sem user agent
		if userAgent == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "User agent required",
			})
			c.Abort()
			return
		}

		// Verificar user agents bloqueados
		for _, blocked := range blockedUserAgents {
			if blocked != "" && contains(userAgent, blocked) {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "Access denied",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// contains verifica se uma string contém outra (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					findInString(s, substr)))
}

// findInString busca substring ignorando case
func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Timeout middleware de timeout para requisições
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Criar contexto com timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Substituir contexto da requisição
		c.Request = c.Request.WithContext(ctx)

		// Channel para capturar o fim da requisição
		finished := make(chan struct{})

		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// Requisição terminou normalmente
		case <-ctx.Done():
			// Timeout atingido
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"error":   "Request timeout",
			})
			c.Abort()
		}
	}
}

// ValidateContentType valida o content type das requisições
func ValidateContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Apenas validar para métodos que enviam dados
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")

			// Lista de content types permitidos
			allowedTypes := []string{
				"application/json",
				"application/x-www-form-urlencoded",
				"multipart/form-data",
			}

			allowed := false
			for _, allowedType := range allowedTypes {
				if contains(contentType, allowedType) {
					allowed = true
					break
				}
			}

			if !allowed {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"success": false,
					"error":   "Unsupported content type",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// MaxRequestSize limita o tamanho máximo das requisições
func MaxRequestSize(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Request too large. Maximum size: %d bytes", maxSize),
			})
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}
