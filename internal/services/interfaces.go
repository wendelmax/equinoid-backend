package services

import (
	"context"

	"github.com/equinoid/backend/internal/models"
)

type UserServiceInterface interface {
	GetByID(ctx context.Context, id uint) (*models.User, error)
	UpdateProfile(ctx context.Context, id uint, req *models.UpdateProfileRequest) (*models.User, error)
	Delete(ctx context.Context, id uint) error
	ChangePassword(ctx context.Context, id uint, currentPassword, newPassword string) error
	IsEmailAvailable(ctx context.Context, email string) (bool, error)
}

type AuthServiceInterface interface {
	Login(ctx context.Context, email, password string) (*models.TokenPair, *models.User, error)
	Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*models.TokenPair, *models.User, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	RevokeToken(ctx context.Context, token interface{}) error
}

type EquinoServiceInterface interface {
	List(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.Equino, int64, error)
	GetByEquinoid(ctx context.Context, equinoidID string) (*models.Equino, error)
	Create(ctx context.Context, req *models.CreateEquinoRequest, userID uint) (*models.Equino, error)
	Update(ctx context.Context, equinoidID string, req *models.UpdateEquinoRequest) (*models.Equino, error)
	Delete(ctx context.Context, equinoidID string) error
	TransferOwnership(ctx context.Context, equinoidID string, newOwnerID uint) error
}

type LinhagemServiceInterface interface {
	GetArvoreGenealogica(ctx context.Context, equinoidID string, geracoes int) (*models.ArvoreGenealogica, error)
	ValidarParentesco(ctx context.Context, equinoid1, equinoid2 string) (*models.ResultadoValidacaoParentesco, error)
	GetDescendentes(ctx context.Context, equinoidID string) ([]*models.Equino, error)
}

type ValorizacaoServiceInterface interface {
	CreateRegistro(ctx context.Context, equinoidID string, req *models.CreateValorizacaoRequest, userID uint) (*models.RegistroValorizacao, error)
	GetByID(ctx context.Context, id uint) (*models.RegistroValorizacao, error)
	List(ctx context.Context, equinoidID string, page, limit int, filters map[string]interface{}) ([]*models.RegistroValorizacao, int64, error)
	ValidateRegistro(ctx context.Context, id uint, validadorID uint, aprovado bool, observacoes string) error
	GetTotalPoints(ctx context.Context, equinoidID string) (int, error)
	GetRanking(ctx context.Context, categoria string, limit int) ([]*models.RankingItem, error)
}

type ReproducaoServiceInterface interface {
	CreateCobertura(ctx context.Context, reprodutorID, matrizID string, req *models.CreateCoberturaRequest, veterinarioID uint) (*models.Cobertura, error)
	GetCoberturasReprodutor(ctx context.Context, reprodutorID string) ([]*models.Cobertura, error)
	GetCoberturasMatriz(ctx context.Context, matrizID string) ([]*models.Cobertura, error)
	CreateAvaliacaoSemen(ctx context.Context, reprodutorID string, req *models.CreateAvaliacaoSemenRequest, laboratorioID uint) (*models.AvaliacaoSemen, error)
	CreateGestacao(ctx context.Context, coberturaID uint, veterinarioID uint) (*models.Gestacao, error)
	GetGestacoes(ctx context.Context, matrizID string) ([]*models.Gestacao, error)
	RegistrarParto(ctx context.Context, gestacaoID uint, req *models.RegistrarPartoRequest) (*models.Gestacao, error)
	GetRankingReprodutivo(ctx context.Context, sexo string, limit int) ([]*models.RankingReprodutivo, error)
}

type SocialServiceInterface interface {
	CreatePerfilSocial(ctx context.Context, equinoidID string, userID uint) (*models.PerfilSocial, error)
	GetPerfilSocial(ctx context.Context, equinoidID string) (*models.PerfilSocial, error)
	UpdatePerfilSocial(ctx context.Context, equinoidID string, nomePerfil, bio, localizacao string) (*models.PerfilSocial, error)
	CreatePost(ctx context.Context, equinoidID string, req *models.CreatePostRequest, userID uint) (*models.PostSocial, error)
	GetPosts(ctx context.Context, equinoidID string, page, limit int) ([]*models.PostSocial, int64, error)
	CreateInteracao(ctx context.Context, postID uint, userID uint, tipoInteracao models.TipoInteracao) error
	CreateOferta(ctx context.Context, equinoidID string, req *models.CreateOfertaRequest, userID uint) (*models.Oferta, error)
	GetOfertas(ctx context.Context, equinoidID string) ([]*models.Oferta, error)
}

type CertificateServiceInterface interface {
	GenerateCertificate(ctx context.Context, equinoidID string, certificateType string) (string, error)
	ValidateCertificate(ctx context.Context, serialNumber string) (bool, error)
	RevokeCertificate(ctx context.Context, serialNumber string) error
}

type SearchServiceInterface interface {
	SearchEquinos(ctx context.Context, query string, filters map[string]interface{}, page, limit int) ([]*models.Equino, int64, error)
	SearchAdvanced(ctx context.Context, criteria *models.SearchCriteria) ([]*models.Equino, int64, error)
}

type ReportServiceInterface interface {
	GetDashboardStats(ctx context.Context, userID uint) (map[string]interface{}, error)
	GenerateEquinosReport(ctx context.Context, userID uint, filters map[string]interface{}) ([]*models.Equino, error)
}

type IntegrationServiceInterface interface {
	GetGedaveData(ctx context.Context, equinoidID string) (map[string]interface{}, error)
	SyncGedave(ctx context.Context, equinoidID string) error
}
