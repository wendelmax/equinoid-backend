package app

import (
	"github.com/equinoid/backend/internal/config"
	"github.com/equinoid/backend/internal/handlers"
	"github.com/equinoid/backend/internal/middleware"
	"github.com/equinoid/backend/internal/modules/auth"
	"github.com/equinoid/backend/internal/modules/equinos"
	"github.com/equinoid/backend/internal/modules/eventos"
	"github.com/equinoid/backend/internal/modules/gestacao"
	"github.com/equinoid/backend/internal/modules/participacoes"
	"github.com/equinoid/backend/internal/modules/simulador"
	"github.com/equinoid/backend/internal/modules/tokenizacao"
	"github.com/equinoid/backend/internal/modules/users"
	"github.com/equinoid/backend/internal/modules/leiloes"
	"github.com/equinoid/backend/internal/modules/exames"
	"github.com/equinoid/backend/internal/modules/rankings"
	"github.com/equinoid/backend/internal/modules/relatorios"
	"github.com/equinoid/backend/internal/modules/financeiro"
	"github.com/equinoid/backend/internal/modules/nutricao"
	"github.com/equinoid/backend/internal/modules/treinamento"
	"github.com/equinoid/backend/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func BuildRouter(modules *ModuleContainer, cfg *config.Config, logger *logging.Logger, keycloakAuth *middleware.KeycloakAuth, useKeycloak bool, db *gorm.DB) *gin.Engine {
	router := gin.New()

	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.RateLimit(cfg.RateLimitPerMinute))

	if cfg.MetricsEnabled {
		router.GET(cfg.MetricsPath, gin.WrapH(promhttp.Handler()))
	}

	legacyHandlers := createLegacyHandlersAdapter(modules.LegacyHandlers, db, logger, cfg)
	
	router.GET("/health", modules.HealthHandler.HealthCheck)
	router.GET("/health/ready", modules.HealthHandler.ReadinessCheck)
	router.GET("/health/live", modules.HealthHandler.LivenessCheck)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")

	var authMiddleware gin.HandlerFunc
	if useKeycloak && keycloakAuth != nil {
		authMiddleware = keycloakAuth.AuthMiddleware()
	} else {
		authMiddleware = middleware.AuthMiddleware(cfg.JWTSecret)
	}

	auth.RegisterRoutes(v1, modules.AuthHandler, authMiddleware)
	users.RegisterRoutes(v1, modules.UsersHandler, authMiddleware)
	equinos.RegisterRoutes(v1, modules.EquinosHandler, authMiddleware)
	simulador.RegisterRoutes(v1, modules.SimuladorHandler, authMiddleware)
	participacoes.RegisterRoutes(v1, modules.ParticipacoesHandler, authMiddleware)
	gestacao.RegisterRoutes(v1, modules.GestacaoHandler, authMiddleware)
	eventos.RegisterRoutes(v1, modules.EventosHandler, authMiddleware)
	tokenizacao.RegisterRoutes(v1, modules.TokenizacaoHandler, authMiddleware)
	leiloes.RegisterRoutes(v1, modules.LeiloesHandler, authMiddleware)
	exames.RegisterRoutes(v1, modules.ExamesHandler, authMiddleware)
	rankings.RegisterRoutes(v1, modules.RankingsHandler, authMiddleware)
	relatorios.RegisterRoutes(v1, modules.RelatoriosHandler, authMiddleware)
	financeiro.RegisterRoutes(v1, modules.FinanceiroHandler, authMiddleware)
	nutricao.RegisterRoutes(v1, modules.NutricaoHandler, authMiddleware)
	treinamento.RegisterRoutes(v1, modules.TreinamentoHandler, authMiddleware)

	protected := v1.Group("")
	protected.Use(authMiddleware)
	{
		registerLegacyRoutes(protected, legacyHandlers)
	}

	return router
}

func createLegacyHandlersAdapter(legacy *LegacyHandlers, db *gorm.DB, logger *logging.Logger, cfg *config.Config) *handlers.Handlers {
	return &handlers.Handlers{
		DB:                       db,
		Logger:                   logger,
		ValorizacaoService:       legacy.ValorizacaoService,
		LinhagemService:          legacy.LinhagemService,
		ReproducaoService:        legacy.ReproducaoService,
		SocialService:            legacy.SocialService,
		EventoService:            legacy.EventoService,
		CertificateService:       legacy.CertificateService,
		IntegrationService:       legacy.IntegrationService,
		ReportService:            legacy.ReportService,
		SearchService:            legacy.SearchService,
		WebhookService:           legacy.WebhookService,
		ChatbotService:           legacy.ChatbotService,
		PropriedadeService:       legacy.PropriedadeService,
		D4SignService:            legacy.D4SignService,
		LeilaoService:            legacy.LeilaoService,
		ExameLaboratorialService: legacy.ExameLaboratorialService,
	}
}

func registerLegacyRoutes(rg *gin.RouterGroup, h *handlers.Handlers) {
	propriedades := rg.Group("/propriedades")
	{
		propriedades.GET("", h.ListPropriedades)
		propriedades.POST("", h.CreatePropriedade)
		propriedades.GET("/:id", h.GetPropriedade)
		propriedades.PUT("/:id", h.UpdatePropriedade)
		propriedades.DELETE("/:id", h.DeletePropriedade)
	}

	certificates := rg.Group("/certificates")
	{
		certificates.GET("", h.ListCertificates)
		certificates.POST("/generate", h.GenerateCertificate)
		certificates.GET("/validate/:serial", h.ValidateCertificate)
		certificates.POST("/revoke", h.RevokeCertificate)
	}


}
