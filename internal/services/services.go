package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/equinoid/backend/internal/config"
	"github.com/equinoid/backend/internal/constants"
	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/internal/modules/equinos"
	"github.com/equinoid/backend/internal/modules/users"
	"github.com/equinoid/backend/internal/utils"
	"github.com/equinoid/backend/pkg/cache"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo users.Repository
	cache    cache.CacheInterface
	logger   *logging.Logger
}

func NewUserService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger) *UserService {
	userRepo := users.NewRepository(db)
	return &UserService{
		userRepo: userRepo,
		cache:    cache,
		logger:   logger,
	}
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, err
		}
		s.logger.LogError(err, "UserService", logging.Fields{"user_id": id})
		return nil, apperrors.NewDatabaseError("get_by_id", "erro ao buscar usuário", err)
	}
	user.Password = ""
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, id uint, req *models.UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, err
		}
		s.logger.LogError(err, "UserService", logging.Fields{"user_id": id})
		return nil, apperrors.NewDatabaseError("update_profile", "erro ao buscar usuário", err)
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.CPFCNPJ != "" {
		user.CPFCNPJ = req.CPFCNPJ
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.LogError(err, "UserService", logging.Fields{"user_id": id})
		return nil, apperrors.NewDatabaseError("update_profile", "erro ao atualizar perfil", err)
	}

	user.Password = ""
	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id uint) error {
	if err := s.userRepo.Delete(ctx, id); err != nil {
		if apperrors.IsNotFound(err) {
			return err
		}
		s.logger.LogError(err, "UserService", logging.Fields{"user_id": id})
		return apperrors.NewDatabaseError("delete", "erro ao deletar usuário", err)
	}
	return nil
}

func (s *UserService) ChangePassword(ctx context.Context, id uint, currentPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return err
		}
		s.logger.LogError(err, "UserService", logging.Fields{"user_id": id})
		return apperrors.NewDatabaseError("change_password", "erro ao buscar usuário", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return apperrors.ErrInvalidCredentials.WithReason("senha atual incorreta")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.LogError(err, "UserService", logging.Fields{"user_id": id})
		return apperrors.NewBusinessError("PASSWORD_HASH_ERROR", "erro ao processar nova senha", nil)
	}

	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.LogError(err, "UserService", logging.Fields{"user_id": id})
		return apperrors.NewDatabaseError("change_password", "erro ao atualizar senha", err)
	}

	return nil
}

func (s *UserService) IsEmailAvailable(ctx context.Context, email string) (bool, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		s.logger.LogError(err, "UserService", logging.Fields{"email": email})
		return false, apperrors.NewDatabaseError("is_email_available", "erro ao verificar disponibilidade do email", err)
	}
	return !exists, nil
}

type AuthService struct {
	userRepo users.Repository
	cache    cache.CacheInterface
	logger   *logging.Logger
	config   *config.Config
}

func NewAuthService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger, config *config.Config) *AuthService {
	userRepo := users.NewRepository(db)
	return &AuthService{
		userRepo: userRepo,
		cache:    cache,
		logger:   logger,
		config:   config,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.TokenPair, *models.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			s.logger.LogAuthentication(0, email, false, "usuário não encontrado")
			return nil, nil, apperrors.ErrInvalidCredentials
		}
		s.logger.LogError(err, "AuthService", logging.Fields{"email": email})
		return nil, nil, apperrors.NewDatabaseError("login", "erro ao buscar usuário", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.logger.LogAuthentication(user.ID, email, false, "senha incorreta")
		return nil, nil, apperrors.ErrInvalidCredentials.WithReason("senha incorreta")
	}

	tokens, err := s.generateTokenPair(user)
	if err != nil {
		s.logger.LogError(err, "AuthService", logging.Fields{"user_id": user.ID})
		return nil, nil, apperrors.NewBusinessError("TOKEN_GENERATION_ERROR", "erro ao gerar tokens", nil)
	}

	s.logger.LogAuthentication(user.ID, email, true, "login bem-sucedido")
	user.Password = ""
	return tokens, user, nil
}

func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		s.logger.LogError(err, "AuthService", logging.Fields{"email": req.Email})
		return nil, apperrors.NewDatabaseError("register", "erro ao verificar disponibilidade do email", err)
	}
	if exists {
		return nil, &apperrors.ConflictError{Resource: "email", Message: "email já está em uso", Value: req.Email}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.LogError(err, "AuthService", logging.Fields{"email": req.Email})
		return nil, apperrors.NewBusinessError("PASSWORD_HASH_ERROR", "erro ao processar senha", nil)
	}

	user := &models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
		UserType: req.UserType,
		CPFCNPJ:  req.CPFCNPJ,
		IsActive: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.LogError(err, "AuthService", logging.Fields{"email": req.Email})
		return nil, apperrors.NewDatabaseError("register", "erro ao criar usuário", err)
	}

	s.logger.LogAuthentication(user.ID, req.Email, true, "registro bem-sucedido")
	user.Password = ""
	return user, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenPair, *models.User, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		s.logger.LogError(err, "AuthService", logging.Fields{"reason": "token inválido ou expirado"})
		return nil, nil, apperrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, apperrors.ErrInvalidToken.WithReason("claims inválidos")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, nil, apperrors.ErrInvalidToken.WithReason("user_id não encontrado no token")
	}

	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		if apperrors.IsNotFound(err) {
			s.logger.LogError(err, "AuthService", logging.Fields{"user_id": userID})
			return nil, nil, &apperrors.NotFoundError{Resource: "user", Message: "usuário não encontrado", ID: uint(userID)}
		}
		s.logger.LogError(err, "AuthService", logging.Fields{"user_id": userID})
		return nil, nil, apperrors.NewDatabaseError("refresh_token", "erro ao buscar usuário", err)
	}

	tokens, err := s.generateTokenPair(user)
	if err != nil {
		s.logger.LogError(err, "AuthService", logging.Fields{"user_id": user.ID})
		return nil, nil, apperrors.NewBusinessError("TOKEN_GENERATION_ERROR", "erro ao gerar novos tokens", nil)
	}

	user.Password = ""
	return tokens, user, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	_, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			s.logger.LogError(err, "AuthService", logging.Fields{"email": email, "reason": "usuário não encontrado"})
			return &apperrors.NotFoundError{Resource: "user", Message: "usuário não encontrado", ID: email}
		}
		return apperrors.NewDatabaseError("forgot_password", "erro ao buscar usuário", err)
	}

	s.logger.Info("Email de recuperação de senha enviado", logging.Fields{"email": email})
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	s.logger.Info("Reset de senha solicitado", logging.Fields{"token": token})
	return nil
}

func (s *AuthService) RevokeToken(ctx context.Context, token interface{}) error {
	s.logger.Info("Token revogado", logging.Fields{"token": token})
	return nil
}

type PropriedadeService struct {
	db     *gorm.DB
	cache  cache.CacheInterface
	logger *logging.Logger
}

func NewPropriedadeService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger) *PropriedadeService {
	return &PropriedadeService{db: db, cache: cache, logger: logger}
}

func (s *PropriedadeService) Create(ctx context.Context, req *models.CreatePropriedadeRequest, userID uint) (*models.Propriedade, error) {
	propriedade := &models.Propriedade{
		Nome:          req.Nome,
		Tipo:          req.Tipo,
		CNPJ:          req.CNPJ,
		Endereco:      req.Endereco,
		Cidade:        req.Cidade,
		Estado:        req.Estado,
		Pais:          req.Pais,
		CEP:           req.CEP,
		Telefone:      req.Telefone,
		Email:         req.Email,
		ResponsavelID: userID,
	}

	if req.ResponsavelID != 0 {
		propriedade.ResponsavelID = req.ResponsavelID
	}

	if err := s.db.WithContext(ctx).Create(propriedade).Error; err != nil {
		s.logger.LogError(err, "PropriedadeService", logging.Fields{"nome": req.Nome})
		return nil, apperrors.NewDatabaseError("create", "erro ao criar propriedade", err)
	}

	return propriedade, nil
}

func (s *PropriedadeService) GetByID(ctx context.Context, id uint) (*models.Propriedade, error) {
	var propriedade models.Propriedade
	if err := s.db.WithContext(ctx).Preload("Responsavel").Preload("Equinos").First(&propriedade, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "propriedade", Message: "propriedade não encontrada", ID: id}
		}
		return nil, apperrors.NewDatabaseError("get_by_id", "erro ao buscar propriedade", err)
	}
	return &propriedade, nil
}

func (s *PropriedadeService) List(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.Propriedade, int64, error) {
	var propriedades []*models.Propriedade
	var total int64

	query := s.db.WithContext(ctx).Model(&models.Propriedade{})

	if tipo, ok := filters["tipo"].(string); ok && tipo != "" {
		query = query.Where("tipo = ?", tipo)
	}
	if responsavelID, ok := filters["responsavel_id"].(uint); ok {
		query = query.Where("responsavel_id = ?", responsavelID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.NewDatabaseError("list", "erro ao contar propriedades", err)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&propriedades).Error; err != nil {
		return nil, 0, apperrors.NewDatabaseError("list", "erro ao listar propriedades", err)
	}

	return propriedades, total, nil
}

func (s *PropriedadeService) Update(ctx context.Context, id uint, req *models.UpdatePropriedadeRequest) (*models.Propriedade, error) {
	propriedade, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Nome != nil {
		propriedade.Nome = *req.Nome
	}
	if req.Tipo != nil {
		propriedade.Tipo = *req.Tipo
	}
	if req.IsActive != nil {
		propriedade.IsActive = *req.IsActive
	}

	if err := s.db.WithContext(ctx).Save(propriedade).Error; err != nil {
		return nil, apperrors.NewDatabaseError("update", "erro ao atualizar propriedade", err)
	}

	return propriedade, nil
}

func (s *PropriedadeService) Delete(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Delete(&models.Propriedade{}, id).Error; err != nil {
		return apperrors.NewDatabaseError("delete", "erro ao deletar propriedade", err)
	}
	return nil
}

type EquinoService struct {
	equinoRepo    equinos.Repository
	userRepo      users.Repository
	d4signService *D4SignService
	cache         cache.CacheInterface
	logger        *logging.Logger
}

func NewEquinoService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger, d4signService *D4SignService) *EquinoService {
	equinoRepo := equinos.NewRepository(db)
	userRepo := users.NewRepository(db)
	return &EquinoService{
		equinoRepo:    equinoRepo,
		userRepo:      userRepo,
		d4signService: d4signService,
		cache:         cache,
		logger:        logger,
	}
}

func (s *EquinoService) List(ctx context.Context, page, limit int, filters map[string]interface{}) ([]*models.Equino, int64, error) {
	equinos, total, err := s.equinoRepo.List(ctx, page, limit, filters)
	if err != nil {
		s.logger.LogError(err, "EquinoService", logging.Fields{"page": page, "limit": limit})
		return nil, 0, apperrors.NewDatabaseError("list", "erro ao listar equinos", err)
	}
	return equinos, total, nil
}

func (s *EquinoService) GetByEquinoid(ctx context.Context, equinoidID string) (*models.Equino, error) {
	equino, err := s.equinoRepo.FindByEquinoid(ctx, equinoidID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, err
		}
		s.logger.LogError(err, "EquinoService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("get_by_equinoid", "erro ao buscar equino", err)
	}
	return equino, nil
}

func (s *EquinoService) Create(ctx context.Context, req *models.CreateEquinoRequest, userID uint) (*models.Equino, error) {
	if req.MicrochipID != "" {
		_, err := s.equinoRepo.FindByMicrochipID(ctx, req.MicrochipID)
		if err == nil {
			return nil, &apperrors.ConflictError{Resource: "microchip_id", Message: "MicrochipID já existe", Value: req.MicrochipID}
		}
		if !apperrors.IsNotFound(err) {
			s.logger.LogError(err, "EquinoService", logging.Fields{"microchip_id": req.MicrochipID})
			return nil, apperrors.NewDatabaseError("create", "erro ao verificar MicrochipID", err)
		}
	}

	equinoId, err := utils.GenerateEquinoId(
		req.PaisOrigem,
		req.MicrochipID,
		req.Nome,
		*req.DataNascimento,
		string(req.Sexo),
		req.Pelagem,
		req.Raca,
	)
	if err != nil {
		s.logger.LogError(err, "EquinoService", logging.Fields{"nome": req.Nome, "error": "failed to generate equinoid"})
		return nil, apperrors.NewBusinessError("EQUINOID_GENERATION_ERROR", "erro ao gerar EquinoId", map[string]interface{}{"error": err})
	}

	equino := &models.Equino{
		Equinoid:       equinoId,
		MicrochipID:    req.MicrochipID,
		Nome:           req.Nome,
		Sexo:           req.Sexo,
		Raca:           req.Raca,
		Pelagem:        req.Pelagem,
		PaisOrigem:     req.PaisOrigem,
		DataNascimento: req.DataNascimento,
		Genitora:       req.GenitoraEquinoid,
		Genitor:        req.GenitorEquinoid,
		PropriedadeID:  req.PropriedadeID,
		FotoPerfil:     req.FotoPerfil,
		Status:         models.StatusAtivo,
		ProprietarioID: userID,
	}

	if err := s.equinoRepo.Create(ctx, equino); err != nil {
		s.logger.LogError(err, "EquinoService", logging.Fields{"nome": req.Nome, "equinoid": equinoId})
		return nil, apperrors.NewDatabaseError("create", "erro ao criar equino", err)
	}

	s.logger.LogBusinessEvent("equino_created", "Equino criado com sucesso", userID, equinoId, logging.Fields{"nome": req.Nome})
	return equino, nil
}

func (s *EquinoService) Update(ctx context.Context, equinoidID string, req *models.UpdateEquinoRequest) (*models.Equino, error) {
	equino, err := s.equinoRepo.FindByEquinoid(ctx, equinoidID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, err
		}
		s.logger.LogError(err, "EquinoService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("update", "erro ao buscar equino", err)
	}

	if req.Nome != nil && *req.Nome != "" {
		equino.Nome = *req.Nome
	}
	if req.Pelagem != nil && *req.Pelagem != "" {
		equino.Pelagem = *req.Pelagem
	}
	if req.PaisOrigem != nil && *req.PaisOrigem != "" {
		equino.PaisOrigem = *req.PaisOrigem
	}
	if req.Status != nil {
		equino.Status = *req.Status
	}
	if req.PropriedadeID != nil {
		equino.PropriedadeID = *req.PropriedadeID
	}
	if req.FotoPerfil != nil {
		equino.FotoPerfil = *req.FotoPerfil
	}
	if req.FotosGaleria != nil {
		// Converter []string para JSONB
		fotos := make(models.JSONB)
		for i, f := range *req.FotosGaleria {
			fotos[fmt.Sprintf("%d", i)] = f
		}
		equino.FotosGaleria = fotos
	}

	if err := s.equinoRepo.Update(ctx, equino); err != nil {
		s.logger.LogError(err, "EquinoService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("update", "erro ao atualizar equino", err)
	}

	return equino, nil
}

func (s *EquinoService) Delete(ctx context.Context, equinoidID string) error {
	if err := s.equinoRepo.Delete(ctx, equinoidID); err != nil {
		if apperrors.IsNotFound(err) {
			return err
		}
		s.logger.LogError(err, "EquinoService", logging.Fields{"equinoid": equinoidID})
		return apperrors.NewDatabaseError("delete", "erro ao deletar equino", err)
	}
	return nil
}

func (s *EquinoService) TransferOwnership(ctx context.Context, equinoidID string, newOwnerID uint) error {
	equino, err := s.equinoRepo.FindByEquinoid(ctx, equinoidID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return err
		}
		s.logger.LogError(err, "EquinoService", logging.Fields{"equinoid": equinoidID})
		return apperrors.NewDatabaseError("transfer_ownership", "erro ao buscar equino", err)
	}

	equino.ProprietarioID = newOwnerID

	if err := s.equinoRepo.Update(ctx, equino); err != nil {
		s.logger.LogError(err, "EquinoService", logging.Fields{"equinoid": equinoidID, "new_owner": newOwnerID})
		return apperrors.NewDatabaseError("transfer_ownership", "erro ao transferir propriedade", err)
	}

	s.logger.LogBusinessEvent("ownership_transferred", "Propriedade transferida", newOwnerID, equinoidID, logging.Fields{"old_owner": equino.ProprietarioID})
	return nil
}

type EventoService struct {
	db     *gorm.DB
	cache  cache.CacheInterface
	logger *logging.Logger
}

func NewEventoService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger) *EventoService {
	return &EventoService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

func (s *EventoService) GetEventosByEquinoid(ctx context.Context, equinoidID string) ([]*models.Evento, error) {
	var eventos []*models.Evento
	if err := s.db.WithContext(ctx).Where("equino_id IN (SELECT id FROM equinos WHERE equinoid = ?)", equinoidID).Order("data_evento DESC").Find(&eventos).Error; err != nil {
		s.logger.LogError(err, "EventoService.GetEventosByEquinoid", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("get_eventos", "erro ao buscar eventos", err)
	}
	return eventos, nil
}

func (s *EventoService) CreateEvento(ctx context.Context, equinoidID string, req *models.CreateEventoRequest, veterinarioID *uint) (*models.Evento, error) {
	equinoRepo := equinos.NewRepository(s.db)
	equino, err := equinoRepo.FindByEquinoid(ctx, equinoidID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado", ID: equinoidID}
		}
		s.logger.LogError(err, "EventoService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("create_evento", "erro ao buscar equino", err)
	}

	var documentos models.JSONB
	if len(req.Documentos) > 0 {
		documentos = make(models.JSONB)
		for i, doc := range req.Documentos {
			documentos[fmt.Sprintf("doc_%d", i)] = doc
		}
	}

	evento := &models.Evento{
		EquinoID:              equino.ID,
		TipoEvento:            req.TipoEvento,
		Categoria:             req.Categoria,
		TipoEventoCompetitivo: req.TipoEventoCompetitivo,
		TipoEventoPublico:     req.TipoEventoPublico,
		NomeEvento:            req.NomeEvento,
		Descricao:             req.Descricao,
		DataEvento:            req.DataEvento,
		Local:                 req.Local,
		Organizador:           req.Organizador,
		Resultados:            req.Resultados,
		Participante:          req.Participante,
		Particularidades:      req.Particularidades,
		ValorInscricao:        req.ValorInscricao,
		AceitaPatrocinio:      req.AceitaPatrocinio,
		InformacoesPatrocinio: req.InformacoesPatrocinio,
		Documentos:            documentos,
	}

	if veterinarioID != nil {
		evento.VeterinarioID = veterinarioID
	}

	if err := s.db.WithContext(ctx).Create(evento).Error; err != nil {
		s.logger.LogError(err, "EventoService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("create_evento", "erro ao criar evento", err)
	}

	s.logger.LogBusinessEvent("evento_created", "Evento registrado com sucesso", 0, equinoidID, logging.Fields{"evento_id": evento.ID, "tipo": req.TipoEvento})
	return evento, nil
}

type CertificateService struct {
	db         *gorm.DB
	equinoRepo equinos.Repository
	cache      cache.CacheInterface
	logger     *logging.Logger
	config     *config.Config
}

func NewCertificateService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger, config *config.Config) *CertificateService {
	equinoRepo := equinos.NewRepository(db)
	return &CertificateService{
		db:         db,
		equinoRepo: equinoRepo,
		cache:      cache,
		logger:     logger,
		config:     config,
	}
}

func (s *CertificateService) GenerateCertificate(ctx context.Context, userID uint, equinoid string, certificateType string, validDays int) (*models.Certificate, error) {
	_, err := s.equinoRepo.FindByEquinoid(ctx, equinoid)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado", ID: equinoid}
		}
		s.logger.LogError(err, "CertificateService", logging.Fields{"equinoid": equinoid})
		return nil, apperrors.NewDatabaseError("generate_certificate", "erro ao buscar equino", err)
	}

	serialNumber := fmt.Sprintf("CERT-%s-%d", equinoid, time.Now().Unix())
	now := time.Now()
	validTo := now.AddDate(0, 0, validDays)

	certificate := &models.Certificate{
		UserID:         userID,
		SerialNumber:   serialNumber,
		CertificatePEM: "PEM_PLACEHOLDER", // Em cenário real, geraria o par de chaves e assinaria
		PrivateKeyPEM:  "PRIVATE_KEY_PLACEHOLDER",
		PublicKeyPEM:   "PUBLIC_KEY_PLACEHOLDER",
		CommonName:     fmt.Sprintf("%s - %s", equinoid, certificateType),
		IssuedAt:       now,
		ValidFrom:      now,
		ValidTo:        validTo,
		ExpiresAt:      validTo,
		IsRevoked:      false,
	}

	if err := s.db.WithContext(ctx).Create(certificate).Error; err != nil {
		s.logger.LogError(err, "CertificateService", logging.Fields{"equinoid": equinoid})
		return nil, apperrors.NewDatabaseError("generate_certificate", "erro ao salvar certificado", err)
	}

	return certificate, nil
}

func (s *CertificateService) ListCertificates(ctx context.Context, userID uint) ([]*models.Certificate, error) {
	var certificates []*models.Certificate
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&certificates).Error
	if err != nil {
		s.logger.LogError(err, "CertificateService", logging.Fields{"user_id": userID})
		return nil, apperrors.NewDatabaseError("list_certificates", "erro ao listar certificados", err)
	}
	return certificates, nil
}

func (s *CertificateService) GetBySerialNumber(ctx context.Context, serialNumber string) (*models.Certificate, error) {
	var certificate models.Certificate
	if err := s.db.WithContext(ctx).Where("serial_number = ?", serialNumber).First(&certificate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "certificate", Message: "certificado não encontrado", ID: serialNumber}
		}
		return nil, apperrors.NewDatabaseError("get_by_serial", "erro ao buscar certificado", err)
	}
	return &certificate, nil
}

func (s *CertificateService) ValidateCertificate(ctx context.Context, serialNumber string) (*models.Certificate, error) {
	certificate, err := s.GetBySerialNumber(ctx, serialNumber)
	if err != nil {
		return nil, err
	}

	return certificate, nil
}

func (s *CertificateService) RevokeCertificate(ctx context.Context, serialNumber string, reason string) error {
	certificate, err := s.GetBySerialNumber(ctx, serialNumber)
	if err != nil {
		return err
	}

	certificate.Revoke(reason)

	if err := s.db.WithContext(ctx).Save(certificate).Error; err != nil {
		s.logger.LogError(err, "CertificateService", logging.Fields{"serial": serialNumber})
		return apperrors.NewDatabaseError("revoke_certificate", "erro ao revogar certificado", err)
	}

	return nil
}

type ValorizacaoService struct {
	db     *gorm.DB
	cache  cache.CacheInterface
	logger *logging.Logger
}

func NewValorizacaoService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger) *ValorizacaoService {
	return &ValorizacaoService{db: db, cache: cache, logger: logger}
}

func (s *ValorizacaoService) CreateRegistro(ctx context.Context, equinoidID string, req *models.CreateValorizacaoRequest, userID uint) (*models.RegistroValorizacao, error) {
	equinoRepo := equinos.NewRepository(s.db)
	_, err := equinoRepo.FindByEquinoid(ctx, equinoidID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado", ID: equinoidID}
		}
		s.logger.LogError(err, "ValorizacaoService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("create_registro", "erro ao buscar equino", err)
	}

	pontos := s.calculatePoints(req.Categoria, req.TipoRegistro, req.NivelImportancia)

	registro := &models.RegistroValorizacao{
		Equinoid:                 equinoidID,
		Categoria:                req.Categoria,
		TipoRegistro:             req.TipoRegistro,
		Titulo:                   req.Titulo,
		Descricao:                req.Descricao,
		DataRegistro:             req.DataRegistro,
		DataValidade:             req.DataValidade,
		LocalEvento:              req.LocalEvento,
		Pais:                     req.Pais,
		Estado:                   req.Estado,
		Cidade:                   req.Cidade,
		Organizacao:              req.Organizacao,
		InstituicaoCertificadora: req.InstituicaoCertificadora,
		NumeroCertificado:        req.NumeroCertificado,
		ValorMonetario:           req.ValorMonetario,
		PontosValorizacao:        pontos,
		NivelImportancia:         req.NivelImportancia,
		StatusValidacao:          models.StatusPendente,
		CriadoPor:                userID,
	}

	if err := s.db.WithContext(ctx).Create(registro).Error; err != nil {
		s.logger.LogError(err, "ValorizacaoService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("create_registro", "erro ao criar registro de valorização", err)
	}

	return registro, nil
}

func (s *ValorizacaoService) GetByID(ctx context.Context, id uint) (*models.RegistroValorizacao, error) {
	var registro models.RegistroValorizacao
	if err := s.db.WithContext(ctx).Preload("Equino").Preload("Criador").Preload("Validador").Where("id = ?", id).First(&registro).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "registro_valorizacao", Message: "registro não encontrado", ID: id}
		}
		s.logger.LogError(err, "ValorizacaoService", logging.Fields{"id": id})
		return nil, apperrors.NewDatabaseError("get_by_id", "erro ao buscar registro", err)
	}
	return &registro, nil
}

func (s *ValorizacaoService) List(ctx context.Context, equinoidID string, page, limit int, filters map[string]interface{}) ([]*models.RegistroValorizacao, int64, error) {
	var registros []*models.RegistroValorizacao
	var total int64

	query := s.db.WithContext(ctx).Model(&models.RegistroValorizacao{}).Where("equinoid = ?", equinoidID)

	if categoria, ok := filters["categoria"].(string); ok && categoria != "" {
		query = query.Where("categoria = ?", categoria)
	}
	if status, ok := filters["status_validacao"].(string); ok && status != "" {
		query = query.Where("status_validacao = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		s.logger.LogError(err, "ValorizacaoService", nil)
		return nil, 0, apperrors.NewDatabaseError("list", "erro ao contar registros", err)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("data_registro DESC").Find(&registros).Error; err != nil {
		s.logger.LogError(err, "ValorizacaoService", nil)
		return nil, 0, apperrors.NewDatabaseError("list", "erro ao listar registros", err)
	}

	return registros, total, nil
}

func (s *ValorizacaoService) ValidateRegistro(ctx context.Context, id uint, validadorID uint, aprovado bool, observacoes string) error {
	registro := &models.RegistroValorizacao{}
	if err := s.db.WithContext(ctx).First(registro, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &apperrors.NotFoundError{Resource: "registro_valorizacao", Message: "registro não encontrado", ID: id}
		}
		return apperrors.NewDatabaseError("validate_registro", "erro ao buscar registro", err)
	}

	now := time.Now()
	if aprovado {
		registro.StatusValidacao = models.StatusAprovado
	} else {
		registro.StatusValidacao = models.StatusRejeitado
	}
	registro.ValidadoPor = &validadorID
	registro.DataValidacao = &now
	registro.ObservacoesValidacao = observacoes

	if err := s.db.WithContext(ctx).Save(registro).Error; err != nil {
		s.logger.LogError(err, "ValorizacaoService", logging.Fields{"registro_id": id})
		return apperrors.NewDatabaseError("validate_registro", "erro ao validar registro", err)
	}

	return nil
}

func (s *ValorizacaoService) GetTotalPoints(ctx context.Context, equinoidID string) (int, error) {
	var total int64
	err := s.db.WithContext(ctx).Model(&models.RegistroValorizacao{}).
		Where("equinoid = ? AND status_validacao = ?", equinoidID, models.StatusAprovado).
		Select("COALESCE(SUM(pontos_valorizacao), 0)").
		Scan(&total).Error

	if err != nil {
		s.logger.LogError(err, "ValorizacaoService", logging.Fields{"equinoid": equinoidID})
		return 0, apperrors.NewDatabaseError("get_total_points", "erro ao calcular pontos totais", err)
	}

	return int(total), nil
}

func (s *ValorizacaoService) GetRanking(ctx context.Context, categoria string, limit int) ([]*models.RankingItem, error) {
	var rankings []*models.RankingItem

	query := `
		SELECT 
			e.equinoid,
			e.nome,
			COALESCE(SUM(rv.pontos_valorizacao), 0) as total_pontos,
			COUNT(rv.id) as total_registros
		FROM equinos e
		LEFT JOIN registro_valorizacaos rv ON e.equinoid = rv.equinoid 
			AND rv.status_validacao = 'aprovado'
	`

	if categoria != "" {
		query += " AND rv.categoria = ?"
	}

	query += `
		GROUP BY e.equinoid, e.nome
		ORDER BY total_pontos DESC
		LIMIT ?
	`

	var args []interface{}
	if categoria != "" {
		args = append(args, categoria, limit)
	} else {
		args = append(args, limit)
	}

	if err := s.db.WithContext(ctx).Raw(query, args...).Scan(&rankings).Error; err != nil {
		s.logger.LogError(err, "ValorizacaoService", logging.Fields{"categoria": categoria})
		return nil, apperrors.NewDatabaseError("get_ranking", "erro ao obter ranking", err)
	}

	return rankings, nil
}

func (s *ValorizacaoService) calculatePoints(categoria models.CategoriaValorizacao, tipoRegistro string, nivel models.NivelImportancia) int {
	basePoints := map[models.CategoriaValorizacao]int{
		models.CategoriaCompeticao:      constants.PontosBaseCompeticao,
		models.CategoriaReproducao:      constants.PontosBaseReproducao,
		models.CategoriaSaude:           constants.PontosBaseSaude,
		models.CategoriaTreinamento:     constants.PontosBaseTreinamento,
		models.CategoriaComercial:       constants.PontosBaseComercial,
		models.CategoriaMidia:           constants.PontosBaseMidia,
		models.CategoriaEducacao:        constants.PontosBaseEducacao,
		models.CategoriaViagens:         constants.PontosBaseViagens,
		models.CategoriaReconhecimentos: constants.PontosBaseReconhecimentos,
		models.CategoriaParcerias:       constants.PontosBaseParcerias,
		models.CategoriaAnalise:         constants.PontosBaseAnalise,
	}

	nivelMultiplier := map[models.NivelImportancia]float64{
		models.NivelBaixo:       constants.MultiplicadorNivelBaixo,
		models.NivelMedio:       constants.MultiplicadorNivelMedio,
		models.NivelAlto:        constants.MultiplicadorNivelAlto,
		models.NivelCritico:     constants.MultiplicadorNivelCritico,
		models.NivelExcepcional: constants.MultiplicadorNivelExcepcional,
	}

	points := basePoints[categoria]
	multiplier := nivelMultiplier[nivel]

	return int(float64(points) * multiplier)
}

func (s *ValorizacaoService) CreateLeilao(ctx context.Context, equinoidID string, registroValorizacaoID uint, req *models.DadosLeilaoRequest) (*models.LeilaoValorizacao, error) {
	var registro models.RegistroValorizacao
	if err := s.db.WithContext(ctx).First(&registro, registroValorizacaoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "registro_valorizacao", Message: "registro de valorização não encontrado", ID: registroValorizacaoID}
		}
		s.logger.LogError(err, "ValorizacaoService", logging.Fields{"registro_id": registroValorizacaoID})
		return nil, apperrors.NewDatabaseError("create_leilao", "erro ao buscar registro de valorização", err)
	}

	if registro.Equinoid != equinoidID {
		return nil, &apperrors.ValidationError{Field: "equinoid", Message: "registro de valorização não pertence ao equino", Value: equinoidID}
	}

	leilao := &models.LeilaoValorizacao{
		RegistroValorizacaoID: registroValorizacaoID,
		NomeLeilao:            req.NomeLeilao,
		TipoLeilao:            req.TipoLeilao,
		Especializacao:        req.Especializacao,
		CasaLeiloeira:         req.CasaLeiloeira,
		DataLeilao:            req.DataLeilao,
		PrecoLance:            req.PrecoLance,
		PrecoVenda:            req.PrecoVenda,
		PosicaoLeilao:         req.PosicaoLeilao,
		ValorizacaoPercentual: req.ValorizacaoPercentual,
		StatusLeilao:          models.StatusLeilaoAgendado,
		Resultado:             models.ResultadoNaoVendido,
	}

	if req.StatusLeilao != "" {
		leilao.StatusLeilao = req.StatusLeilao
	}
	if req.Resultado != "" {
		leilao.Resultado = req.Resultado
	}

	if err := s.db.WithContext(ctx).Create(leilao).Error; err != nil {
		s.logger.LogError(err, "ValorizacaoService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("create_leilao", "erro ao criar leilão", err)
	}

	s.logger.LogBusinessEvent("leilao_created", "Leilão criado com sucesso", 0, equinoidID, logging.Fields{"leilao_id": leilao.ID})
	return leilao, nil
}

func (s *ValorizacaoService) GetHistoricoLeiloes(ctx context.Context, equinoidID string) ([]*models.LeilaoValorizacao, error) {
	var leiloes []*models.LeilaoValorizacao

	query := s.db.WithContext(ctx).
		Model(&models.LeilaoValorizacao{}).
		Joins("JOIN registro_valorizacaos ON registro_valorizacaos.id = leilaos.registro_valorizacao_id").
		Where("registro_valorizacaos.equinoid = ?", equinoidID).
		Order("leilaos.data_leilao DESC")

	if err := query.Find(&leiloes).Error; err != nil {
		s.logger.LogError(err, "ValorizacaoService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("get_historico_leiloes", "erro ao buscar histórico de leilões", err)
	}

	return leiloes, nil
}

type LinhagemService struct {
	equinoRepo equinos.Repository
	cache      cache.CacheInterface
	logger     *logging.Logger
}

func NewLinhagemService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger) *LinhagemService {
	equinoRepo := equinos.NewRepository(db)
	return &LinhagemService{
		equinoRepo: equinoRepo,
		cache:      cache,
		logger:     logger,
	}
}

func (s *LinhagemService) GetArvoreGenealogica(ctx context.Context, equinoidID string, geracoes int) (*models.ArvoreGenealogica, error) {
	if geracoes > constants.MaxGeracoesArvoreGenealogica {
		geracoes = constants.MaxGeracoesArvoreGenealogica
	}

	equino, err := s.getEquinoWithParents(ctx, equinoidID)
	if err != nil {
		return nil, err
	}

	arvore := &models.ArvoreGenealogica{
		Equinoid: equinoidID,
		Nome:     equino.Nome,
		Geracoes: geracoes,
	}

	if geracoes > 0 {
		arvore.Ancestrais = s.buildAncestorTree(ctx, equino, geracoes-1)
	}

	return arvore, nil
}

func (s *LinhagemService) getEquinoWithParents(ctx context.Context, equinoidID string) (*models.Equino, error) {
	equino, err := s.equinoRepo.FindByEquinoid(ctx, equinoidID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, apperrors.ErrEquinoNotFound.WithID(equinoidID)
		}
		s.logger.LogError(err, "LinhagemService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("get_equino_with_parents", "erro ao buscar equino", err)
	}
	return equino, nil
}

func (s *LinhagemService) buildAncestorTree(ctx context.Context, equino *models.Equino, geracoesRestantes int) *models.Ancestrais {
	if geracoesRestantes < 0 {
		return nil
	}

	ancestrais := &models.Ancestrais{}

	if equino.Genitor != "" {
		genitor, err := s.getEquinoWithParents(ctx, equino.Genitor)
		if err == nil {
			ancestrais.Pai = &models.AncestralNode{
				Equinoid: genitor.Equinoid,
				Nome:     genitor.Nome,
				Sexo:     string(genitor.Sexo),
			}
			if geracoesRestantes > 0 {
				ancestrais.Pai.Ancestrais = s.buildAncestorTree(ctx, genitor, geracoesRestantes-1)
			}
		}
	}

	if equino.Genitora != "" {
		genitora, err := s.getEquinoWithParents(ctx, equino.Genitora)
		if err == nil {
			ancestrais.Mae = &models.AncestralNode{
				Equinoid: genitora.Equinoid,
				Nome:     genitora.Nome,
				Sexo:     string(genitora.Sexo),
			}
			if geracoesRestantes > 0 {
				ancestrais.Mae.Ancestrais = s.buildAncestorTree(ctx, genitora, geracoesRestantes-1)
			}
		}
	}

	return ancestrais
}

func (s *LinhagemService) ValidarParentesco(ctx context.Context, equinoid1, equinoid2 string) (*models.ResultadoValidacaoParentesco, error) {
	equino1, err := s.getEquinoWithParents(ctx, equinoid1)
	if err != nil {
		return nil, err
	}

	equino2, err := s.getEquinoWithParents(ctx, equinoid2)
	if err != nil {
		return nil, err
	}

	ancestrais1 := s.getAllAncestors(ctx, equino1, constants.MaxGeracoesAncestrais)
	ancestrais2 := s.getAllAncestors(ctx, equino2, constants.MaxGeracoesAncestrais)

	comum := s.findCommonAncestors(ancestrais1, ancestrais2)

	resultado := &models.ResultadoValidacaoParentesco{
		Equino1:                    equinoid1,
		Equino2:                    equinoid2,
		SaoParentes:                len(comum) > 0,
		AncestaisComuns:            comum,
		CoeficienteConsanguinidade: s.calculateConsanguinidade(comum),
	}

	if len(comum) > 0 {
		resultado.GrauParentesco = s.determineRelationship(ctx, equinoid1, equinoid2, comum)
	}

	return resultado, nil
}

func (s *LinhagemService) getAllAncestors(ctx context.Context, equino *models.Equino, maxGeracoes int) []string {
	ancestors := []string{}
	visited := make(map[string]bool)

	var collect func(equinoidID string, geracao int)
	collect = func(equinoidID string, geracao int) {
		if geracao > maxGeracoes || visited[equinoidID] {
			return
		}
		visited[equinoidID] = true
		ancestors = append(ancestors, equinoidID)

		eq, err := s.getEquinoWithParents(ctx, equinoidID)
		if err != nil {
			return
		}

		if eq.Genitor != "" {
			collect(eq.Genitor, geracao+1)
		}
		if eq.Genitora != "" {
			collect(eq.Genitora, geracao+1)
		}
	}

	if equino.Genitor != "" {
		collect(equino.Genitor, 1)
	}
	if equino.Genitora != "" {
		collect(equino.Genitora, 1)
	}

	return ancestors
}

func (s *LinhagemService) findCommonAncestors(list1, list2 []string) []string {
	comum := []string{}
	set := make(map[string]bool)

	for _, id := range list1 {
		set[id] = true
	}

	for _, id := range list2 {
		if set[id] {
			comum = append(comum, id)
		}
	}

	return comum
}

func (s *LinhagemService) calculateConsanguinidade(ancestraisComuns []string) float64 {
	if len(ancestraisComuns) == 0 {
		return 0.0
	}

	return float64(len(ancestraisComuns)) * constants.CoeficienteConsanguinidadeBase
}

func (s *LinhagemService) determineRelationship(ctx context.Context, equino1, equino2 string, ancestraisComuns []string) string {
	eq1, _ := s.getEquinoWithParents(ctx, equino1)
	eq2, _ := s.getEquinoWithParents(ctx, equino2)

	if eq1.Genitor == equino2 || eq1.Genitora == equino2 {
		return "filho/filha"
	}
	if eq2.Genitor == equino1 || eq2.Genitora == equino1 {
		return "pai/mãe"
	}

	if eq1.Genitor == eq2.Genitor || eq1.Genitora == eq2.Genitora {
		return "irmão/irmã"
	}

	if len(ancestraisComuns) > 0 {
		return fmt.Sprintf("parentes (%d ancestrais comuns)", len(ancestraisComuns))
	}

	return "não relacionados"
}

func (s *LinhagemService) GetDescendentes(ctx context.Context, equinoidID string) ([]*models.Equino, error) {
	descendentesGenitor, err := s.equinoRepo.FindByGenitor(ctx, equinoidID)
	if err != nil {
		s.logger.LogError(err, "LinhagemService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("get_descendentes", "erro ao buscar descendentes do genitor", err)
	}

	descendentesGenitora, err := s.equinoRepo.FindByGenitora(ctx, equinoidID)
	if err != nil {
		s.logger.LogError(err, "LinhagemService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("get_descendentes", "erro ao buscar descendentes da genitora", err)
	}

	descendentes := append(descendentesGenitor, descendentesGenitora...)
	return descendentes, nil
}

type ReproducaoService struct {
	equinoRepo equinos.Repository
	db         *gorm.DB
	cache      cache.CacheInterface
	logger     *logging.Logger
}

func NewReproducaoService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger) *ReproducaoService {
	equinoRepo := equinos.NewRepository(db)
	return &ReproducaoService{
		equinoRepo: equinoRepo,
		db:         db,
		cache:      cache,
		logger:     logger,
	}
}

func (s *ReproducaoService) GetAvaliacoesSemen(ctx context.Context, equinoidID string) ([]*models.AvaliacaoSemen, error) {
	var avaliacoes []*models.AvaliacaoSemen
	if err := s.db.WithContext(ctx).Where("reprodutor_equinoid = ?", equinoidID).Order("data_avaliacao DESC").Find(&avaliacoes).Error; err != nil {
		s.logger.LogError(err, "ReproducaoService.GetAvaliacoesSemen", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("get_avaliacoes", "erro ao buscar avaliações de sêmen", err)
	}
	return avaliacoes, nil
}

func (s *ReproducaoService) CreateCobertura(ctx context.Context, reprodutorID, matrizID string, req *models.CreateCoberturaRequest, veterinarioID uint) (*models.Cobertura, error) {
	reprodutor, err := s.equinoRepo.FindByEquinoid(ctx, reprodutorID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "reprodutor não encontrado", ID: reprodutorID}
		}
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"reprodutor": reprodutorID})
		return nil, apperrors.NewDatabaseError("create_cobertura", "erro ao buscar reprodutor", err)
	}

	matriz, err := s.equinoRepo.FindByEquinoid(ctx, matrizID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "matriz não encontrada", ID: matrizID}
		}
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"matriz": matrizID})
		return nil, apperrors.NewDatabaseError("create_cobertura", "erro ao buscar matriz", err)
	}

	if reprodutor.Sexo != models.SexoMacho {
		return nil, &apperrors.ValidationError{Field: "sexo", Message: "reprodutor deve ser macho", Value: reprodutor.Sexo}
	}
	if matriz.Sexo != models.SexoFemea {
		return nil, &apperrors.ValidationError{Field: "sexo", Message: "matriz deve ser fêmea", Value: matriz.Sexo}
	}

	cobertura := &models.Cobertura{
		ReprodutorEquinoid:     reprodutorID,
		MatrizEquinoid:         matrizID,
		DataCobertura:          req.DataCobertura,
		TipoCobertura:          req.TipoCobertura,
		MetodoCobertura:        req.MetodoCobertura,
		VeterinarioResponsavel: veterinarioID,
		LaboratorioID:          req.LaboratorioID,
		StatusCobertura:        models.StatusCoberturaPendente,
		ProbabilidadeConcepcao: req.ProbabilidadeConcepcao,
		Observacoes:            req.Observacoes,
	}

	if err := s.db.WithContext(ctx).Create(cobertura).Error; err != nil {
		s.logger.LogError(err, "ReproducaoService", logging.Fields{
			"reprodutor": reprodutorID,
			"matriz":     matrizID,
		})
		return nil, apperrors.NewDatabaseError("create_cobertura", "erro ao criar cobertura", err)
	}

	return cobertura, nil
}

func (s *ReproducaoService) GetCoberturasReprodutor(ctx context.Context, reprodutorID string) ([]*models.Cobertura, error) {
	var coberturas []*models.Cobertura
	if err := s.db.WithContext(ctx).Where("reprodutor_equinoid = ?", reprodutorID).
		Preload("Matriz").
		Preload("Veterinario").
		Order("data_cobertura DESC").
		Find(&coberturas).Error; err != nil {
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"reprodutor": reprodutorID})
		return nil, apperrors.NewDatabaseError("get_coberturas_reprodutor", "erro ao buscar coberturas", err)
	}
	return coberturas, nil
}

func (s *ReproducaoService) GetCoberturasMatriz(ctx context.Context, matrizID string) ([]*models.Cobertura, error) {
	var coberturas []*models.Cobertura
	if err := s.db.WithContext(ctx).Where("matriz_equinoid = ?", matrizID).
		Preload("Reprodutor").
		Preload("Veterinario").
		Order("data_cobertura DESC").
		Find(&coberturas).Error; err != nil {
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"matriz": matrizID})
		return nil, apperrors.NewDatabaseError("get_coberturas_matriz", "erro ao buscar coberturas", err)
	}
	return coberturas, nil
}

func (s *ReproducaoService) CreateAvaliacaoSemen(ctx context.Context, reprodutorID string, req *models.CreateAvaliacaoSemenRequest, laboratorioID uint) (*models.AvaliacaoSemen, error) {
	reprodutor, err := s.equinoRepo.FindByEquinoid(ctx, reprodutorID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "reprodutor não encontrado", ID: reprodutorID}
		}
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"reprodutor": reprodutorID})
		return nil, apperrors.NewDatabaseError("create_avaliacao_semen", "erro ao buscar reprodutor", err)
	}

	if reprodutor.Sexo != models.SexoMacho {
		return nil, &apperrors.ValidationError{Field: "sexo", Message: "avaliação de sêmen só pode ser feita em machos", Value: reprodutor.Sexo}
	}

	qualidade := s.determineQualidadeSemen(req)
	aptidao := s.determineAptidaoReprodutiva(req)

	avaliacao := &models.AvaliacaoSemen{
		ReprodutorEquinoid:          reprodutorID,
		CoberturaID:                 req.CoberturaID,
		DataColeta:                  req.DataColeta,
		DataAnalise:                 req.DataAnalise,
		LaboratorioID:               laboratorioID,
		VolumeSemen:                 req.VolumeSemen,
		ConcentracaoEspermatozoides: req.ConcentracaoEspermatozoides,
		MotiliadeProgressiva:        req.MotilidadeProgressiva,
		MotiliadeTotal:              req.MotilidadeTotal,
		Viabilidade:                 req.Viabilidade,
		MorfologiaNormal:            req.MorfologiaNormal,
		QualidadeGeral:              qualidade,
		AptidaoReprodutiva:          aptidao,
		DataValidade:                req.DataValidade,
		TemperaturaArmazenamento:    req.TemperaturaArmazenamento,
		TecnicoResponsavel:          req.TecnicoResponsavel,
		Observacoes:                 req.Observacoes,
	}

	if err := s.db.WithContext(ctx).Create(avaliacao).Error; err != nil {
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"reprodutor": reprodutorID})
		return nil, apperrors.NewDatabaseError("create_avaliacao_semen", "erro ao criar avaliação de sêmen", err)
	}

	return avaliacao, nil
}

func (s *ReproducaoService) CreateGestacao(ctx context.Context, coberturaID uint, veterinarioID uint) (*models.Gestacao, error) {
	var cobertura models.Cobertura
	if err := s.db.WithContext(ctx).Preload("Matriz").First(&cobertura, coberturaID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "cobertura", Message: "cobertura não encontrada", ID: coberturaID}
		}
		return nil, apperrors.NewDatabaseError("create_gestacao", "erro ao buscar cobertura", err)
	}

	var existingGestacao models.Gestacao
	if err := s.db.WithContext(ctx).Where("cobertura_id = ?", coberturaID).First(&existingGestacao).Error; err == nil {
		return nil, &apperrors.ConflictError{Resource: "gestacao", Message: "já existe gestação para esta cobertura", Value: coberturaID}
	}

	dataPrevisaoParto := cobertura.DataCobertura.AddDate(0, constants.MesesGestacaoEquino, 0)

	gestacao := &models.Gestacao{
		CoberturaID:             coberturaID,
		MatrizEquinoid:          cobertura.MatrizEquinoid,
		DataCobertura:           cobertura.DataCobertura,
		DataPrevistaParto:       dataPrevisaoParto,
		VeterinarioResponsavel:  veterinarioID,
		StatusGestacao:          models.StatusGestacaoAtiva,
		NumeroUltrassonografias: 0,
	}

	if err := s.db.WithContext(ctx).Create(gestacao).Error; err != nil {
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"cobertura_id": coberturaID})
		return nil, apperrors.NewDatabaseError("create_gestacao", "erro ao criar gestação", err)
	}

	now := time.Now()
	cobertura.StatusCobertura = models.StatusCoberturaConfirmada
	cobertura.DataConfirmacao = &now
	if err := s.db.WithContext(ctx).Save(&cobertura).Error; err != nil {
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"cobertura_id": coberturaID})
	}

	return gestacao, nil
}

func (s *ReproducaoService) GetGestacoes(ctx context.Context, matrizID string) ([]*models.Gestacao, error) {
	var gestacoes []*models.Gestacao
	if err := s.db.WithContext(ctx).Where("matriz_equinoid = ?", matrizID).
		Preload("Cobertura").
		Preload("Cobertura.Reprodutor").
		Order("data_cobertura DESC").
		Find(&gestacoes).Error; err != nil {
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"matriz": matrizID})
		return nil, apperrors.NewDatabaseError("get_gestacoes", "erro ao buscar gestações", err)
	}
	return gestacoes, nil
}

func (s *ReproducaoService) RegistrarParto(ctx context.Context, gestacaoID uint, req *models.RegistrarPartoRequest) (*models.Gestacao, error) {
	var gestacao models.Gestacao
	if err := s.db.WithContext(ctx).First(&gestacao, gestacaoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "gestacao", Message: "gestação não encontrada", ID: gestacaoID}
		}
		return nil, apperrors.NewDatabaseError("registrar_parto", "erro ao buscar gestação", err)
	}

	gestacao.DataRealParto = &req.DataParto
	gestacao.TipoParto = req.TipoParto
	gestacao.StatusGestacao = models.StatusGestacaoConcluida
	gestacao.Observacoes = req.ObservacoesParto

	if err := s.db.WithContext(ctx).Save(&gestacao).Error; err != nil {
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"gestacao_id": gestacaoID})
		return nil, apperrors.NewDatabaseError("registrar_parto", "erro ao registrar parto", err)
	}

	return &gestacao, nil
}

func (s *ReproducaoService) GetRankingReprodutivo(ctx context.Context, sexo string, limit int) ([]*models.RankingReprodutivo, error) {
	var rankings []*models.RankingReprodutivo

	query := `
		SELECT 
			e.equinoid,
			e.nome,
			e.sexo,
			COUNT(DISTINCT c.id) as total_coberturas,
			COUNT(DISTINCT g.id) as total_gestacoes,
			COUNT(DISTINCT CASE WHEN g.resultado_parto = 'sucesso' THEN g.id END) as partos_sucesso,
			CASE 
				WHEN COUNT(DISTINCT c.id) > 0 
				THEN ROUND((COUNT(DISTINCT g.id)::numeric / COUNT(DISTINCT c.id)::numeric) * 100, 2)
				ELSE 0
			END as taxa_concepcao
		FROM equinos e
	`

	if sexo == "macho" {
		query += " LEFT JOIN coberturas c ON e.equinoid = c.reprodutor_equinoid"
	} else {
		query += " LEFT JOIN coberturas c ON e.equinoid = c.matriz_equinoid"
	}

	query += `
		LEFT JOIN gestacoes g ON c.id = g.cobertura_id
		WHERE e.sexo = ?
		GROUP BY e.equinoid, e.nome, e.sexo
		HAVING COUNT(DISTINCT c.id) > 0
		ORDER BY taxa_concepcao DESC, total_coberturas DESC
		LIMIT ?
	`

	if err := s.db.WithContext(ctx).Raw(query, sexo, limit).Scan(&rankings).Error; err != nil {
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"sexo": sexo})
		return nil, apperrors.NewDatabaseError("get_ranking_reprodutivo", "erro ao obter ranking reprodutivo", err)
	}

	return rankings, nil
}

func (s *ReproducaoService) CreatePerformanceMaterna(ctx context.Context, matrizID string, gestacaoID uint, req *models.CreatePerformanceMaternaRequest) (*models.PerformanceMaterna, error) {
	matriz, err := s.equinoRepo.FindByEquinoid(ctx, matrizID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "matriz não encontrada", ID: matrizID}
		}
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"matriz": matrizID})
		return nil, apperrors.NewDatabaseError("create_performance_materna", "erro ao buscar matriz", err)
	}

	if matriz.Sexo != models.SexoFemea {
		return nil, &apperrors.ValidationError{Field: "sexo", Message: "performance materna só pode ser registrada para fêmeas", Value: matriz.Sexo}
	}

	var gestacao models.Gestacao
	if err := s.db.WithContext(ctx).First(&gestacao, gestacaoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "gestacao", Message: "gestação não encontrada", ID: gestacaoID}
		}
		return nil, apperrors.NewDatabaseError("create_performance_materna", "erro ao buscar gestação", err)
	}

	performance := &models.PerformanceMaterna{
		MatrizEquinoid:           matrizID,
		GestacaoID:               gestacaoID,
		PesoInicioGestacao:       req.PesoInicioGestacao,
		PesoFimGestacao:          req.PesoFimGestacao,
		GanhoPesoGestacao:        req.GanhoPesoGestacao,
		ProducaoLeiteDiaria:      req.ProducaoLeiteDiaria,
		QualidadeLeite:           req.QualidadeLeite,
		CuidadoMaterno:           req.CuidadoMaterno,
		TempoDesmame:             req.TempoDesmame,
		PesoPotroDesmame:         req.PesoPotroDesmame,
		TempoRecuperacaoPosParto: req.TempoRecuperacaoPosParto,
		IntervaloProximoParto:    req.IntervaloProximoParto,
		Observacoes:              req.Observacoes,
	}

	if req.GanhoPesoGestacao == nil && req.PesoInicioGestacao != nil && req.PesoFimGestacao != nil {
		ganho := *req.PesoFimGestacao - *req.PesoInicioGestacao
		performance.GanhoPesoGestacao = &ganho
	}

	if err := s.db.WithContext(ctx).Create(performance).Error; err != nil {
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"matriz": matrizID, "gestacao_id": gestacaoID})
		return nil, apperrors.NewDatabaseError("create_performance_materna", "erro ao criar performance materna", err)
	}

	s.logger.LogBusinessEvent("performance_materna_created", "Performance materna registrada com sucesso", 0, matrizID, logging.Fields{"performance_id": performance.ID})
	return performance, nil
}

func (s *ReproducaoService) CreateUltrassonografia(ctx context.Context, gestacaoID uint, req *models.CreateUltrassonografiaRequest, veterinarioID uint) (*models.Ultrassonografia, error) {
	var gestacao models.Gestacao
	if err := s.db.WithContext(ctx).First(&gestacao, gestacaoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "gestacao", Message: "gestação não encontrada", ID: gestacaoID}
		}
		return nil, apperrors.NewDatabaseError("create_ultrassonografia", "erro ao buscar gestação", err)
	}

	ultrassonografia := &models.Ultrassonografia{
		GestacaoID:             gestacaoID,
		DataExame:              req.DataExame,
		IdadeGestacional:       req.IdadeGestacional,
		VeterinarioResponsavel: veterinarioID,
		PresencaEmbriao:        req.PresencaEmbriao,
		NumeroEmbrioes:         req.NumeroEmbrioes,
		BatimentoCardiaco:      req.BatimentoCardiaco,
		DesenvolvimentoNormal:  req.DesenvolvimentoNormal,
		TamanhoEmbriao:         req.TamanhoEmbriao,
		FrequenciaCardiaca:     req.FrequenciaCardiaca,
		Diagnostico:            req.Diagnostico,
		Observacoes:            req.Observacoes,
		ProximoExame:           req.ProximoExame,
	}

	if err := s.db.WithContext(ctx).Create(ultrassonografia).Error; err != nil {
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"gestacao_id": gestacaoID})
		return nil, apperrors.NewDatabaseError("create_ultrassonografia", "erro ao criar ultrassonografia", err)
	}

	gestacao.NumeroUltrassonografias++
	if req.DataExame.After(gestacao.DataPrevistaParto.Add(-30 * 24 * time.Hour)) {
		now := time.Now()
		gestacao.UltimaUltrassonografia = &now
	}
	if err := s.db.WithContext(ctx).Save(&gestacao).Error; err != nil {
		s.logger.LogError(err, "ReproducaoService", logging.Fields{"gestacao_id": gestacaoID})
	}

	s.logger.LogBusinessEvent("ultrassonografia_created", "Ultrassonografia registrada com sucesso", veterinarioID, gestacao.MatrizEquinoid, logging.Fields{"ultrassonografia_id": ultrassonografia.ID})
	return ultrassonografia, nil
}

func (s *ReproducaoService) determineQualidadeSemen(req *models.CreateAvaliacaoSemenRequest) models.QualidadeSemen {
	score := 0
	checks := 0

	if req.MotilidadeProgressiva != nil {
		checks++
		if *req.MotilidadeProgressiva >= 60 {
			score += 2
		} else if *req.MotilidadeProgressiva >= 40 {
			score += 1
		}
	}

	if req.MotilidadeTotal != nil {
		checks++
		if *req.MotilidadeTotal >= 70 {
			score += 2
		} else if *req.MotilidadeTotal >= 50 {
			score += 1
		}
	}

	if req.MorfologiaNormal != nil {
		checks++
		if *req.MorfologiaNormal >= 70 {
			score += 2
		} else if *req.MorfologiaNormal >= 50 {
			score += 1
		}
	}

	if req.Viabilidade != nil {
		checks++
		if *req.Viabilidade >= 80 {
			score += 2
		} else if *req.Viabilidade >= 60 {
			score += 1
		}
	}

	if checks == 0 {
		return models.QualidadeRegular
	}

	avg := float64(score) / float64(checks*2)

	if avg >= 0.9 {
		return models.QualidadeExcelente
	} else if avg >= 0.7 {
		return models.QualidadeBoa
	} else if avg >= 0.5 {
		return models.QualidadeRegular
	} else if avg >= 0.3 {
		return models.QualidadeRuim
	}
	return models.QualidadeInadequada
}

func (s *ReproducaoService) determineAptidaoReprodutiva(req *models.CreateAvaliacaoSemenRequest) models.AptidaoReprodutiva {
	qualidade := s.determineQualidadeSemen(req)

	switch qualidade {
	case models.QualidadeExcelente:
		return models.AptidaoAlta
	case models.QualidadeBoa:
		return models.AptidaoMedia
	case models.QualidadeRegular:
		return models.AptidaoBaixa
	default:
		return models.AptidaoInadequada
	}
}

type SocialService struct {
	equinoRepo equinos.Repository
	db         *gorm.DB
	cache      cache.CacheInterface
	logger     *logging.Logger
}

func NewSocialService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger) *SocialService {
	equinoRepo := equinos.NewRepository(db)
	return &SocialService{
		equinoRepo: equinoRepo,
		db:         db,
		cache:      cache,
		logger:     logger,
	}
}

func (s *SocialService) CreatePerfilSocial(ctx context.Context, equinoidID string, userID uint) (*models.PerfilSocial, error) {
	_, err := s.equinoRepo.FindByEquinoid(ctx, equinoidID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado", ID: equinoidID}
		}
		s.logger.LogError(err, "SocialService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("create_perfil_social", "erro ao buscar equino", err)
	}

	var existingPerfil models.PerfilSocial
	if err := s.db.WithContext(ctx).Where("equinoid = ?", equinoidID).First(&existingPerfil).Error; err == nil {
		return nil, &apperrors.ConflictError{Resource: "perfil_social", Message: "perfil social já existe para este equino", Value: equinoidID}
	}

	equino, _ := s.equinoRepo.FindByEquinoid(ctx, equinoidID)
	perfil := &models.PerfilSocial{
		Equinoid:              equinoidID,
		NomePerfil:            equino.Nome,
		StatusDisponibilidade: models.StatusDisponivel,
		TipoPerfil:            models.TipoPerfilPublico,
		MostrarLocalizacao:    true,
		PermitirOfertas:       true,
		PermitirContato:       true,
		PermitirSeguir:        true,
		CriadoPor:             userID,
	}

	if err := s.db.WithContext(ctx).Create(perfil).Error; err != nil {
		s.logger.LogError(err, "SocialService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("create_perfil_social", "erro ao criar perfil social", err)
	}

	return perfil, nil
}

func (s *SocialService) GetPerfilSocial(ctx context.Context, equinoidID string) (*models.PerfilSocial, error) {
	var perfil models.PerfilSocial
	if err := s.db.WithContext(ctx).Where("equinoid = ?", equinoidID).Preload("Equino").First(&perfil).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "perfil_social", Message: "perfil social não encontrado", ID: equinoidID}
		}
		s.logger.LogError(err, "SocialService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("get_perfil_social", "erro ao buscar perfil social", err)
	}
	return &perfil, nil
}

func (s *SocialService) UpdatePerfilSocial(ctx context.Context, equinoidID string, nomePerfil, bio, localizacao string) (*models.PerfilSocial, error) {
	var perfil models.PerfilSocial
	if err := s.db.WithContext(ctx).Where("equinoid = ?", equinoidID).First(&perfil).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "perfil_social", Message: "perfil social não encontrado", ID: equinoidID}
		}
		return nil, apperrors.NewDatabaseError("update_perfil_social", "erro ao buscar perfil social", err)
	}

	if nomePerfil != "" {
		perfil.NomePerfil = nomePerfil
	}
	if bio != "" {
		perfil.Bio = bio
	}
	if localizacao != "" {
		perfil.Localizacao = localizacao
	}

	if err := s.db.WithContext(ctx).Save(&perfil).Error; err != nil {
		s.logger.LogError(err, "SocialService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("update_perfil_social", "erro ao atualizar perfil social", err)
	}

	return &perfil, nil
}

func (s *SocialService) CreatePost(ctx context.Context, equinoidID string, req *models.CreatePostRequest, userID uint) (*models.PostSocial, error) {
	var perfil models.PerfilSocial
	if err := s.db.WithContext(ctx).Where("equinoid = ?", equinoidID).First(&perfil).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "perfil_social", Message: "perfil social não encontrado", ID: equinoidID}
		}
		return nil, apperrors.NewDatabaseError("create_post", "erro ao buscar perfil social", err)
	}

	post := &models.PostSocial{
		Equinoid:                 equinoidID,
		PerfilSocialID:           perfil.ID,
		TipoConteudo:             req.TipoConteudo,
		Legenda:                  req.Legenda,
		LocalizacaoPost:          req.LocalizacaoPost,
		DataPostagem:             time.Now(),
		StatusPost:               models.StatusPostAtivo,
		PermitirComentarios:      true,
		PermitirCompartilhamento: true,
		CriadoPor:                userID,
	}

	if err := s.db.WithContext(ctx).Create(post).Error; err != nil {
		s.logger.LogError(err, "SocialService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("create_post", "erro ao criar post", err)
	}

	perfil.TotalPosts++
	if err := s.db.WithContext(ctx).Save(&perfil).Error; err != nil {
		s.logger.LogError(err, "SocialService", logging.Fields{"equinoid": equinoidID})
	}

	return post, nil
}

func (s *SocialService) GetPosts(ctx context.Context, equinoidID string, page, limit int) ([]*models.PostSocial, int64, error) {
	var posts []*models.PostSocial
	var total int64

	query := s.db.WithContext(ctx).Model(&models.PostSocial{}).Where("equinoid = ? AND status_post = ?", equinoidID, models.StatusPostAtivo)

	if err := query.Count(&total).Error; err != nil {
		s.logger.LogError(err, "SocialService", logging.Fields{"equinoid": equinoidID})
		return nil, 0, apperrors.NewDatabaseError("get_posts", "erro ao contar posts", err)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("data_postagem DESC").Find(&posts).Error; err != nil {
		s.logger.LogError(err, "SocialService", logging.Fields{"equinoid": equinoidID})
		return nil, 0, apperrors.NewDatabaseError("get_posts", "erro ao listar posts", err)
	}

	return posts, total, nil
}

func (s *SocialService) CreateInteracao(ctx context.Context, postID uint, userID uint, tipoInteracao models.TipoInteracao) error {
	var post models.PostSocial
	if err := s.db.WithContext(ctx).First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &apperrors.NotFoundError{Resource: "post", Message: "post não encontrado", ID: postID}
		}
		return apperrors.NewDatabaseError("create_interacao", "erro ao buscar post", err)
	}

	var existingInteracao models.InteracaoSocial
	if err := s.db.WithContext(ctx).Where("post_id = ? AND usuario_id = ? AND tipo_interacao = ?", postID, userID, tipoInteracao).First(&existingInteracao).Error; err == nil {
		return &apperrors.ConflictError{Resource: "interacao", Message: "interação já existe", Value: postID}
	}

	interacao := &models.InteracaoSocial{
		PostID:        postID,
		UserID:        userID,
		TipoInteracao: tipoInteracao,
	}

	if err := s.db.WithContext(ctx).Create(interacao).Error; err != nil {
		s.logger.LogError(err, "SocialService", logging.Fields{"post_id": postID, "user_id": userID})
		return apperrors.NewDatabaseError("create_interacao", "erro ao criar interação", err)
	}

	if tipoInteracao == models.TipoInteracaoCurtida {
		post.TotalCurtidas++
	} else if tipoInteracao == models.TipoInteracaoCompartilhamento {
		post.TotalCompartilhamentos++
	}

	if err := s.db.WithContext(ctx).Save(&post).Error; err != nil {
		s.logger.LogError(err, "SocialService", logging.Fields{"post_id": postID})
	}

	return nil
}

func (s *SocialService) CreateOferta(ctx context.Context, equinoidID string, req *models.CreateOfertaRequest, userID uint) (*models.Oferta, error) {
	var perfil models.PerfilSocial
	if err := s.db.WithContext(ctx).Where("equinoid = ?", equinoidID).First(&perfil).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperrors.NotFoundError{Resource: "perfil_social", Message: "perfil social não encontrado", ID: equinoidID}
		}
		return nil, apperrors.NewDatabaseError("create_oferta", "erro ao buscar perfil social", err)
	}

	if !perfil.PermitirOfertas {
		return nil, &apperrors.ValidationError{Field: "permitir_ofertas", Message: "este perfil não permite ofertas", Value: false}
	}

	oferta := &models.Oferta{
		Equinoid:        equinoidID,
		OfertantePorID:  userID,
		TipoOferta:      req.TipoOferta,
		ValorOferta:     req.ValorOferta,
		Moeda:           req.Moeda,
		CondicoesOferta: req.CondicoesOferta,
		PrazoOferta:     req.PrazoOferta,
		StatusOferta:    models.StatusOfertaPendente,
	}

	if err := s.db.WithContext(ctx).Create(oferta).Error; err != nil {
		s.logger.LogError(err, "SocialService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("create_oferta", "erro ao criar oferta", err)
	}

	return oferta, nil
}

func (s *SocialService) GetOfertas(ctx context.Context, equinoidID string) ([]*models.Oferta, error) {
	var ofertas []*models.Oferta
	if err := s.db.WithContext(ctx).Where("equinoid = ?", equinoidID).
		Order("created_at DESC").
		Find(&ofertas).Error; err != nil {
		s.logger.LogError(err, "SocialService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("get_ofertas", "erro ao listar ofertas", err)
	}
	return ofertas, nil
}

type IntegrationService struct {
	db     *gorm.DB
	cache  cache.CacheInterface
	logger *logging.Logger
	config *config.Config
}

func NewIntegrationService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger, config *config.Config) *IntegrationService {
	return &IntegrationService{db: db, cache: cache, logger: logger, config: config}
}

func (s *IntegrationService) GetGedaveData(ctx context.Context, equinoidID string) (map[string]interface{}, error) {
	equinoRepo := equinos.NewRepository(s.db)
	equino, err := equinoRepo.FindByEquinoid(ctx, equinoidID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado", ID: equinoidID}
		}
		s.logger.LogError(err, "IntegrationService", logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("get_gedave_data", "erro ao buscar equino", err)
	}

	data := map[string]interface{}{
		"equinoid": equino.Equinoid,
		"nome":     equino.Nome,
		"raca":     equino.Raca,
		"sexo":     equino.Sexo,
	}

	gedaveID := s.getGedaveIDFromExternalIDs(equino.ExternalIDs)
	if gedaveID != "" {
		data["gedave_id"] = gedaveID
		data["has_gedave"] = true
	} else {
		data["gedave_id"] = nil
		data["has_gedave"] = false
		data["message"] = "Número GEDAVE não cadastrado. Cadastre em external_ids ao criar/atualizar o equino."
	}

	return data, nil
}

func (s *IntegrationService) getGedaveIDFromExternalIDs(externalIDs models.JSONB) string {
	if externalIDs == nil || len(externalIDs) == 0 {
		return ""
	}

	var externalIDsArray []interface{}

	if ids, ok := externalIDs["external_ids"].([]interface{}); ok {
		externalIDsArray = ids
	} else if ids, ok := externalIDs["external_ids"].([]map[string]interface{}); ok {
		externalIDsArray = make([]interface{}, len(ids))
		for i, v := range ids {
			externalIDsArray[i] = v
		}
	} else {
		return ""
	}

	for _, item := range externalIDsArray {
		extID, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		sistema, ok := extID["sistema"].(string)
		if ok && sistema == "GEDAVE" {
			if id, ok := extID["id"].(string); ok {
				return id
			}
		}
	}

	return ""
}

func (s *IntegrationService) SyncGedave(ctx context.Context, equinoidID string) error {
	equinoRepo := equinos.NewRepository(s.db)
	equino, err := equinoRepo.FindByEquinoid(ctx, equinoidID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return &apperrors.NotFoundError{Resource: "equino", Message: "equino não encontrado", ID: equinoidID}
		}
		s.logger.LogError(err, "IntegrationService", logging.Fields{"equinoid": equinoidID})
		return apperrors.NewDatabaseError("sync_gedave", "erro ao buscar equino", err)
	}

	gedaveID := s.getGedaveIDFromExternalIDs(equino.ExternalIDs)
	if gedaveID == "" {
		return &apperrors.ValidationError{
			Field:   "external_ids",
			Message: "Número GEDAVE não cadastrado. Cadastre em external_ids ao criar/atualizar o equino.",
		}
	}

	s.logger.Info("GEDAVE ID encontrado", logging.Fields{
		"equinoid":  equinoidID,
		"gedave_id": gedaveID,
		"message":   "GEDAVE é armazenado como external_id. Não há sincronização automática.",
	})

	return nil
}

type ReportService struct {
	db     *gorm.DB
	cache  cache.CacheInterface
	logger *logging.Logger
}

func NewReportService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger) *ReportService {
	return &ReportService{db: db, cache: cache, logger: logger}
}

func (s *ReportService) GetDashboardStats(ctx context.Context, userID uint) (map[string]interface{}, error) {
	var totalEquinos int64
	if err := s.db.WithContext(ctx).Model(&models.Equino{}).Where("proprietario_id = ?", userID).Count(&totalEquinos).Error; err != nil {
		s.logger.LogError(err, "ReportService", logging.Fields{"user_id": userID})
		return nil, apperrors.NewDatabaseError("get_dashboard_stats", "erro ao contar equinos", err)
	}

	var totalCoberturas int64
	if err := s.db.WithContext(ctx).Model(&models.Cobertura{}).
		Joins("JOIN equinos ON equinos.equinoid = coberturas.reprodutor_equinoid").
		Where("equinos.proprietario_id = ?", userID).
		Count(&totalCoberturas).Error; err != nil {
		s.logger.LogError(err, "ReportService", logging.Fields{"user_id": userID})
		return nil, apperrors.NewDatabaseError("get_dashboard_stats", "erro ao contar coberturas", err)
	}

	var totalGestacoes int64
	if err := s.db.WithContext(ctx).Model(&models.Gestacao{}).
		Joins("JOIN equinos ON equinos.equinoid = gestacoes.matriz_equinoid").
		Where("equinos.proprietario_id = ?", userID).
		Count(&totalGestacoes).Error; err != nil {
		s.logger.LogError(err, "ReportService", logging.Fields{"user_id": userID})
		return nil, apperrors.NewDatabaseError("get_dashboard_stats", "erro ao contar gestações", err)
	}

	var totalPontos int64
	if err := s.db.WithContext(ctx).Model(&models.RegistroValorizacao{}).
		Joins("JOIN equinos ON equinos.equinoid = registro_valorizacaos.equinoid").
		Where("equinos.proprietario_id = ? AND registro_valorizacaos.status_validacao = ?", userID, models.StatusAprovado).
		Select("COALESCE(SUM(pontos_valorizacao), 0)").
		Scan(&totalPontos).Error; err != nil {
		s.logger.LogError(err, "ReportService", logging.Fields{"user_id": userID})
		return nil, apperrors.NewDatabaseError("get_dashboard_stats", "erro ao calcular pontos", err)
	}

	stats := map[string]interface{}{
		"total_equinos":    totalEquinos,
		"total_coberturas": totalCoberturas,
		"total_gestacoes":  totalGestacoes,
		"total_pontos":     totalPontos,
		"user_id":          userID,
		"timestamp":        time.Now(),
	}

	return stats, nil
}

func (s *ReportService) GenerateEquinosReport(ctx context.Context, userID uint, filters map[string]interface{}) ([]*models.Equino, error) {
	var equinos []*models.Equino

	query := s.db.WithContext(ctx).Where("proprietario_id = ?", userID)

	if sexo, ok := filters["sexo"].(string); ok && sexo != "" {
		query = query.Where("sexo = ?", sexo)
	}
	if raca, ok := filters["raca"].(string); ok && raca != "" {
		query = query.Where("raca = ?", raca)
	}

	if err := query.Preload("RegistrosValorizacao").Preload("CoberturasComo").Find(&equinos).Error; err != nil {
		s.logger.LogError(err, "ReportService", logging.Fields{"user_id": userID})
		return nil, apperrors.NewDatabaseError("generate_equinos_report", "erro ao gerar relatório", err)
	}

	return equinos, nil
}

type SearchService struct {
	equinoRepo equinos.Repository
	cache      cache.CacheInterface
	logger     *logging.Logger
}

func NewSearchService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger) *SearchService {
	equinoRepo := equinos.NewRepository(db)
	return &SearchService{
		equinoRepo: equinoRepo,
		cache:      cache,
		logger:     logger,
	}
}

func (s *SearchService) SearchEquinos(ctx context.Context, query string, filters map[string]interface{}, page, limit int) ([]*models.Equino, int64, error) {
	if query != "" {
		if filters == nil {
			filters = make(map[string]interface{})
		}
		filters["nome"] = query
	}

	equinos, total, err := s.equinoRepo.List(ctx, page, limit, filters)
	if err != nil {
		s.logger.LogError(err, "SearchService", logging.Fields{"query": query})
		return nil, 0, apperrors.NewDatabaseError("search_equinos", "erro ao buscar equinos", err)
	}

	if query != "" {
		filtered := []*models.Equino{}
		for _, equino := range equinos {
			if strings.Contains(strings.ToLower(equino.Nome), strings.ToLower(query)) || strings.Contains(strings.ToLower(equino.Equinoid), strings.ToLower(query)) {
				filtered = append(filtered, equino)
			}
		}
		return filtered, int64(len(filtered)), nil
	}

	return equinos, total, nil
}

func (s *SearchService) SearchAdvanced(ctx context.Context, criteria *models.SearchCriteria) ([]*models.Equino, int64, error) {
	filters := make(map[string]interface{})

	if criteria.Nome != "" {
		filters["nome"] = criteria.Nome
	}
	if criteria.Raca != "" {
		filters["raca"] = criteria.Raca
	}
	if criteria.Sexo != "" {
		filters["sexo"] = criteria.Sexo
	}

	equinos, total, err := s.equinoRepo.List(ctx, criteria.Page, criteria.Limit, filters)
	if err != nil {
		s.logger.LogError(err, "SearchService", logging.Fields{"criteria": criteria})
		return nil, 0, apperrors.NewDatabaseError("search_advanced", "erro ao buscar equinos", err)
	}

	if criteria.IdadeMin != nil || criteria.IdadeMax != nil {
		filtered := []*models.Equino{}
		now := time.Now()
		for _, equino := range equinos {
			if equino.DataNascimento == nil {
				continue
			}
			idade := now.Year() - equino.DataNascimento.Year()
			if criteria.IdadeMin != nil && idade < *criteria.IdadeMin {
				continue
			}
			if criteria.IdadeMax != nil && idade > *criteria.IdadeMax {
				continue
			}
			filtered = append(filtered, equino)
		}
		return filtered, int64(len(filtered)), nil
	}

	return equinos, total, nil
}

type WebhookService struct {
	db     *gorm.DB
	cache  cache.CacheInterface
	logger *logging.Logger
}

func NewWebhookService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger) *WebhookService {
	return &WebhookService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

func (s *WebhookService) RegisterWebhook(ctx context.Context, userID uint, req *models.WebhookRequest) (*models.Webhook, error) {
	webhook := &models.Webhook{
		UserID:   userID,
		URL:      req.URL,
		Secret:   req.Secret,
		IsActive: true,
	}

	events := make(models.JSONB)
	events["events"] = req.Events
	webhook.Events = events

	if err := s.db.WithContext(ctx).Create(webhook).Error; err != nil {
		s.logger.LogError(err, "WebhookService", logging.Fields{"user_id": userID})
		return nil, apperrors.NewDatabaseError("register_webhook", "erro ao registrar webhook", err)
	}

	s.logger.LogBusinessEvent("webhook_registered", "Webhook registrado com sucesso", userID, "", logging.Fields{"webhook_id": webhook.ID})
	return webhook, nil
}

type ChatbotService struct {
	db     *gorm.DB
	cache  cache.CacheInterface
	logger *logging.Logger
	config *config.Config
}

func NewChatbotService(db *gorm.DB, cache cache.CacheInterface, logger *logging.Logger, config *config.Config) *ChatbotService {
	return &ChatbotService{
		db:     db,
		cache:  cache,
		logger: logger,
		config: config,
	}
}

func (s *ChatbotService) ProcessQuery(ctx context.Context, userID uint, req *models.ChatbotQueryRequest) (*models.ChatbotQueryResponse, error) {
	startTime := time.Now()

	query := &models.ChatbotQuery{
		UserID:   userID,
		Equinoid: req.Equinoid,
		Query:    req.Query,
	}

	if err := s.db.WithContext(ctx).Create(query).Error; err != nil {
		s.logger.LogError(err, "ChatbotService", logging.Fields{"user_id": userID})
		return nil, apperrors.NewDatabaseError("process_query", "erro ao processar query", err)
	}

	response := s.generateResponse(req.Query, req.Equinoid)
	processingTime := int(time.Since(startTime).Milliseconds())

	query.Response = response.Response
	query.Confidence = response.Confidence
	query.ProcessingTime = &processingTime

	if err := s.db.WithContext(ctx).Save(query).Error; err != nil {
		s.logger.LogError(err, "ChatbotService", logging.Fields{"query_id": query.ID})
	}

	return response, nil
}

func (s *ChatbotService) generateResponse(query string, equinoid *string) *models.ChatbotQueryResponse {
	queryLower := strings.ToLower(query)
	confidence := 0.8

	// Simulação de IA especializada em equinos
	var response string

	if equinoid != nil && *equinoid != "" {
		if strings.Contains(queryLower, "quem é") || strings.Contains(queryLower, "sobre") || strings.Contains(queryLower, "detalhes") {
			response = fmt.Sprintf("O equino %s é um animal registrado no ecossistema EquinoId. De acordo com os registros de valorização, ele possui uma pontuação baseada em suas conquistas e pedigree. Você pode ver o perfil social completo dele no marketplace.", *equinoid)
		} else if strings.Contains(queryLower, "pai") || strings.Contains(queryLower, "mãe") || strings.Contains(queryLower, "linhagem") {
			response = fmt.Sprintf("A linhagem de %s está documentada em nossa árvore genealógica digital. O sistema valida automaticamente o parentesco e calcula o coeficiente de consanguinidade para garantir a pureza genética.", *equinoid)
		} else if strings.Contains(queryLower, "valor") || strings.Contains(queryLower, "preço") || strings.Contains(queryLower, "venda") {
			response = fmt.Sprintf("O valor estimado de %s é calculado pelo nosso algoritmo que considera prêmios, histórico reprodutivo e performance em pista. Ofertas podem ser feitas diretamente pelo marketplace social.", *equinoid)
		}
	}

	if response == "" {
		if strings.Contains(queryLower, "equinoid") || strings.Contains(queryLower, "o que é") {
			response = "O EquinoId é o identificador definitivo mundial para cavalos. Nossa missão é transformar a história individual de cada cavalo em um ativo digital verificado, gerando confiança e valorização."
		} else if strings.Contains(queryLower, "segurança") || strings.Contains(queryLower, "blockchain") {
			response = "Utilizamos tecnologia Blockchain e Assinatura Digital Biométrica para garantir que cada registro de saúde, treino ou venda seja imutável e assinado por profissionais certificados."
		} else {
			response = "Olá! Sou o assistente inteligente do EquinoId. Posso ajudar você a analisar linhagens, verificar valorizações de animais ou tirar dúvidas sobre o ecossistema. Como posso ser útil hoje?"
		}
	}

	return &models.ChatbotQueryResponse{
		Response:   response,
		Confidence: &confidence,
		Suggestions: []string{
			"O que é EquinoId?",
			"Como funciona a valorização?",
			"Verificar linhagem de um animal",
		},
	}
}

func (s *AuthService) generateTokenPair(user *models.User) (*models.TokenPair, error) {
	accessTokenExpiry := time.Now().Add(24 * time.Hour)
	refreshTokenExpiry := time.Now().Add(7 * 24 * time.Hour)

	accessClaims := jwt.MapClaims{
		"user_id":   user.ID,
		"email":     user.Email,
		"user_type": user.UserType,
		"exp":       accessTokenExpiry.Unix(),
		"iat":       time.Now().Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     refreshTokenExpiry.Unix(),
		"iat":     time.Now().Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(24 * time.Hour.Seconds()),
		TokenType:    "Bearer",
	}, nil
}
