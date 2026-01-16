package app

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/pkg/cache"
	"github.com/equinoid/backend/pkg/logging"
	"github.com/equinoid/backend/pkg/metrics"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db      *gorm.DB
	cache   cache.CacheInterface
	logger  *logging.Logger
	metrics *metrics.Metrics
}

type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Services  map[string]interface{} `json:"services"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	System    SystemInfo             `json:"system"`
}

type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	MemoryUsage  string `json:"memory_usage"`
	CPUCount     int    `json:"cpu_count"`
}

var startTime = time.Now()

func NewHealthHandler(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger) *HealthHandler {
	m := metrics.NewMetrics()
	m.SetApplicationStartTime()
	
	return &HealthHandler{
		db:      db,
		cache:   cache,
		logger:  logger,
		metrics: m,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]interface{})
	overallStatus := "healthy"

	dbStatus := h.checkDatabase(ctx)
	services["database"] = dbStatus
	if dbStatus != "healthy" {
		overallStatus = "unhealthy"
	}

	redisStatus := h.checkRedis(ctx)
	services["redis"] = redisStatus
	if redisStatus == "unhealthy" {
		overallStatus = "degraded"
	}

	services["external_apis"] = h.checkExternalAPIs(ctx)

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	systemInfo := SystemInfo{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		MemoryUsage:  formatBytes(mem.Alloc),
		CPUCount:     runtime.NumCPU(),
	}

	h.metrics.UpdateSystemMetrics(
		float64(mem.Alloc),
		runtime.NumGoroutine(),
		0,
	)

	response := HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  services,
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
		System:    systemInfo,
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

func (h *HealthHandler) checkDatabase(ctx context.Context) string {
	db, err := h.db.DB()
	if err != nil {
		return "unhealthy"
	}

	if err := db.PingContext(ctx); err != nil {
		return "unhealthy"
	}

	stats := db.Stats()
	h.metrics.DBConnectionsTotal.Set(float64(stats.MaxOpenConnections))
	h.metrics.DBConnectionsActive.Set(float64(stats.InUse))
	h.metrics.DBConnectionsIdle.Set(float64(stats.Idle))

	return "healthy"
}

func (h *HealthHandler) checkRedis(ctx context.Context) string {
	if h.cache == nil {
		return "disabled"
	}
	if err := h.cache.Ping(ctx).Err(); err != nil {
		return "unhealthy"
	}
	return "healthy"
}

func (h *HealthHandler) checkExternalAPIs(ctx context.Context) string {
	return "healthy"
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	ready := true
	details := make(map[string]interface{})

	if h.checkDatabase(ctx) != "healthy" {
		ready = false
		details["database"] = "not ready"
	} else {
		details["database"] = "ready"
	}

	if ready {
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"ready":   ready,
				"details": details,
			},
			Message:   "Service is ready",
			Timestamp: time.Now(),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Success:   false,
			Error:     "Service not ready",
			Details:   details,
			Timestamp: time.Now(),
		})
	}
}

func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"alive": true,
		},
		Message:   "Service is alive",
		Timestamp: time.Now(),
	})
}
