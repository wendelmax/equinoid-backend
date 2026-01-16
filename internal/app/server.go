package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/equinoid/backend/internal/config"
	"github.com/equinoid/backend/internal/database"
	"github.com/equinoid/backend/internal/middleware"
	"github.com/equinoid/backend/pkg/cache"
	"github.com/equinoid/backend/pkg/logging"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Config *config.Config
	Logger *logging.Logger
}

func Start(deps Dependencies) error {
	cfg := deps.Config
	logger := deps.Logger

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	redisClient := cache.NewRedisClient(cfg.RedisURL)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}

	var keycloakAuth *middleware.KeycloakAuth
	useKeycloak := os.Getenv("USE_KEYCLOAK") == "true"
	keycloakURL := os.Getenv("KEYCLOAK_URL")
	keycloakRealm := os.Getenv("KEYCLOAK_REALM")
	keycloakClientID := os.Getenv("KEYCLOAK_CLIENT_ID")

	if useKeycloak && keycloakURL != "" {
		keycloakAuth, err = middleware.NewKeycloakAuth(keycloakURL, keycloakRealm, keycloakClientID, db, logger.Logger)
		if err != nil {
			useKeycloak = false
		}
	}

	modules := InitializeModules(db, redisClient, logger, cfg)
	router := BuildRouter(modules, cfg, logger, keycloakAuth, useKeycloak, db)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	go func() {
		_ = srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}

