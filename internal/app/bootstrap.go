package app

import (
	"github.com/equinoid/backend/internal/config"
	"github.com/equinoid/backend/internal/modules/auth"
	"github.com/equinoid/backend/internal/modules/equinos"
	"github.com/equinoid/backend/internal/modules/eventos"
	"github.com/equinoid/backend/internal/modules/exames"
	"github.com/equinoid/backend/internal/modules/financeiro"
	"github.com/equinoid/backend/internal/modules/gestacao"
	"github.com/equinoid/backend/internal/modules/leiloes"
	"github.com/equinoid/backend/internal/modules/nutricao"
	"github.com/equinoid/backend/internal/modules/participacoes"
	"github.com/equinoid/backend/internal/modules/rankings"
	"github.com/equinoid/backend/internal/modules/relatorios"
	"github.com/equinoid/backend/internal/modules/simulador"
	"github.com/equinoid/backend/internal/modules/tokenizacao"
	"github.com/equinoid/backend/internal/modules/treinamento"
	"github.com/equinoid/backend/internal/modules/users"
	"github.com/equinoid/backend/internal/services"
	"github.com/equinoid/backend/pkg/cache"
	"github.com/equinoid/backend/pkg/logging"
	"gorm.io/gorm"
)

type ModuleContainer struct {
	EquinosHandler       *equinos.Handler
	UsersHandler         *users.Handler
	AuthHandler          *auth.Handler
	HealthHandler        *HealthHandler
	SimuladorHandler     *simulador.Handler
	ParticipacoesHandler *participacoes.Handler
	GestacaoHandler      *gestacao.Handler
	EventosHandler       *eventos.Handler
	TokenizacaoHandler   *tokenizacao.Handler
	LeiloesHandler       *leiloes.Handler
	ExamesHandler        *exames.Handler
	RankingsHandler      *rankings.Handler
	RelatoriosHandler    *relatorios.Handler
	FinanceiroHandler    *financeiro.Handler
	NutricaoHandler      *nutricao.Handler
	TreinamentoHandler   *treinamento.Handler
	
	LegacyHandlers *LegacyHandlers
}

type LegacyHandlers struct {
	ValorizacaoService        *services.ValorizacaoService
	LinhagemService           *services.LinhagemService
	ReproducaoService         *services.ReproducaoService
	SocialService             *services.SocialService
	EventoService             *services.EventoService
	CertificateService        *services.CertificateService
	IntegrationService        *services.IntegrationService
	ReportService             *services.ReportService
	SearchService             *services.SearchService
	WebhookService            *services.WebhookService
	ChatbotService            *services.ChatbotService
	PropriedadeService        *services.PropriedadeService
	D4SignService             *services.D4SignService
	LeilaoService             *services.LeilaoService
	ExameLaboratorialService  *services.ExameLaboratorialService
}

func InitializeModules(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger, cfg *config.Config) *ModuleContainer {
	usersRepo := users.NewRepository(db)
	usersService := users.NewService(usersRepo, cache, logger)
	usersHandler := users.NewHandler(usersService, logger)

	authService := auth.NewService(usersRepo, cache, logger, cfg)
	authHandler := auth.NewHandler(authService, logger)

	d4signService := services.NewD4SignService(db, logger, cfg)
	
	equinosRepo := equinos.NewRepository(db)
	equinosService := equinos.NewService(equinosRepo, cache, logger, d4signService)
	equinosHandler := equinos.NewHandler(equinosService, logger)

	legacyHandlers := &LegacyHandlers{
		ValorizacaoService:       services.NewValorizacaoService(db, cache, logger),
		LinhagemService:          services.NewLinhagemService(db, cache, logger),
		ReproducaoService:        services.NewReproducaoService(db, cache, logger),
		SocialService:            services.NewSocialService(db, cache, logger),
		EventoService:            services.NewEventoService(db, cache, logger),
		CertificateService:       services.NewCertificateService(db, cache, logger, cfg),
		IntegrationService:       services.NewIntegrationService(db, cache, logger, cfg),
		ReportService:            services.NewReportService(db, cache, logger),
		SearchService:            services.NewSearchService(db, cache, logger),
		WebhookService:           services.NewWebhookService(db, cache, logger),
		ChatbotService:           services.NewChatbotService(db, cache, logger, cfg),
		PropriedadeService:       services.NewPropriedadeService(db, cache, logger),
		D4SignService:            d4signService,
		LeilaoService:            services.NewLeilaoService(db, cache, logger),
		ExameLaboratorialService: services.NewExameLaboratorialService(db, cache, logger),
	}

	healthHandler := NewHealthHandler(db, cache, logger)

	simuladorService := simulador.NewService(equinosRepo, cache, logger)
	simuladorHandler := simulador.NewHandler(simuladorService, logger)

	participacoesRepo := participacoes.NewRepository(db)
	participacoesService := participacoes.NewService(participacoesRepo, cache, logger)
	participacoesHandler := participacoes.NewHandler(participacoesService, logger)

	gestacaoRepo := gestacao.NewRepository(db)
	gestacaoService := gestacao.NewService(gestacaoRepo, cache, logger)
	gestacaoHandler := gestacao.NewHandler(gestacaoService, logger)

	eventosRepo := eventos.NewRepository(db)
	eventosService := eventos.NewService(eventosRepo, logger)
	eventosHandler := eventos.NewHandler(eventosService, logger)

	tokenizacaoRepo := tokenizacao.NewRepository(db)
	tokenizacaoService := tokenizacao.NewService(tokenizacaoRepo, equinosRepo, logger)
	tokenizacaoHandler := tokenizacao.NewHandler(tokenizacaoService, logger)

	leiloesRepo := leiloes.NewRepository(db)
	leiloesService := leiloes.NewService(leiloesRepo, equinosRepo, logger)
	leiloesHandler := leiloes.NewHandler(leiloesService, logger)

	examesRepo := exames.NewRepository(db)
	examesService := exames.NewService(examesRepo, logger)
	examesHandler := exames.NewHandler(examesService, logger)

	rankingsRepo := rankings.NewRepository(db)
	rankingsService := rankings.NewService(rankingsRepo, logger)
	rankingsHandler := rankings.NewHandler(rankingsService, logger)

	relatoriosRepo := relatorios.NewRepository(db)
	relatoriosService := relatorios.NewService(relatoriosRepo, logger)
	relatoriosHandler := relatorios.NewHandler(relatoriosService, logger)

	financeiroRepo := financeiro.NewRepository(db)
	financeiroHandler := financeiro.NewHandler(financeiroRepo, logger)

	nutricaoRepo := nutricao.NewRepository(db)
	nutricaoService := nutricao.NewService(nutricaoRepo, equinosRepo, logger)
	nutricaoHandler := nutricao.NewHandler(nutricaoService, logger)

	treinamentoRepo := treinamento.NewRepository(db)
	treinamentoService := treinamento.NewService(treinamentoRepo, equinosRepo, logger)
	treinamentoHandler := treinamento.NewHandler(treinamentoService, logger)

	return &ModuleContainer{
		EquinosHandler:       equinosHandler,
		UsersHandler:         usersHandler,
		AuthHandler:          authHandler,
		HealthHandler:        healthHandler,
		SimuladorHandler:     simuladorHandler,
		ParticipacoesHandler: participacoesHandler,
		GestacaoHandler:      gestacaoHandler,
		EventosHandler:       eventosHandler,
		TokenizacaoHandler:   tokenizacaoHandler,
		LeiloesHandler:       leiloesHandler,
		ExamesHandler:        examesHandler,
		RankingsHandler:      rankingsHandler,
		RelatoriosHandler:    relatoriosHandler,
		FinanceiroHandler:    financeiroHandler,
		NutricaoHandler:      nutricaoHandler,
		TreinamentoHandler:   treinamentoHandler,
		LegacyHandlers:       legacyHandlers,
	}
}
