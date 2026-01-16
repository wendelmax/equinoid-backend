package handlers

import (
	"github.com/equinoid/backend/internal/services"
	"github.com/equinoid/backend/pkg/logging"
	"gorm.io/gorm"
)

type Handlers struct {
	DB     *gorm.DB
	Logger *logging.Logger
	
	EventoService            *services.EventoService
	CertificateService       *services.CertificateService
	ValorizacaoService       *services.ValorizacaoService
	LinhagemService          *services.LinhagemService
	ReproducaoService        *services.ReproducaoService
	SocialService            *services.SocialService
	IntegrationService       *services.IntegrationService
	ReportService            *services.ReportService
	SearchService            *services.SearchService
	WebhookService           *services.WebhookService
	ChatbotService           *services.ChatbotService
	PropriedadeService       *services.PropriedadeService
	D4SignService            *services.D4SignService
	LeilaoService            *services.LeilaoService
	ExameLaboratorialService *services.ExameLaboratorialService
}
