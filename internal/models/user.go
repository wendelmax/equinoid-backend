package models

import (
	"time"

	"gorm.io/gorm"
)

// User representa um usuário no sistema
type User struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	SupabaseID      string         `json:"supabase_id,omitempty" gorm:"uniqueIndex;size:36"`
	KeycloakSub     string         `json:"keycloak_sub,omitempty" gorm:"uniqueIndex;size:64"`
	Email           string         `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	Password        string         `json:"-" gorm:"" validate:""`
	Name            string         `json:"name" gorm:"not null" validate:"required"`
	UserType        UserType       `json:"user_type" gorm:"not null" validate:"required"`
	CPFCNPJ         string         `json:"cpf_cnpj" gorm:"index" validate:"required"`
	Role            string         `json:"role" gorm:"size:50;default:'usuario'"`
	IsEmailVerified bool           `json:"is_email_verified" gorm:"default:false"`
	IsActive        bool           `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relacionamentos
	Certificate *Certificate `json:"certificate,omitempty" gorm:"foreignKey:UserID"`
	Equinos     []Equino     `json:"equinos,omitempty" gorm:"foreignKey:ProprietarioID"`
	Eventos     []Evento     `json:"eventos,omitempty" gorm:"foreignKey:VeterinarioID"`
}

// UserType define os tipos de usuário
type UserType string

const (
	UserTypeCriador     UserType = "criador"
	UserTypeVeterinario UserType = "veterinario"
	UserTypeAdmin       UserType = "admin"
	UserTypeLaboratorio UserType = "laboratorio"
	UserTypeParceiro    UserType = "parceiro"
	UserTypeLeiloeiro   UserType = "leiloeiro"
)

// UserResponse representa a resposta do usuário (sem dados sensíveis)
type UserResponse struct {
	ID          uint                 `json:"id"`
	Email       string               `json:"email"`
	Name        string               `json:"name"`
	UserType    UserType             `json:"user_type"`
	CPFCNPJ     string               `json:"cpf_cnpj"`
	IsActive    bool                 `json:"is_active"`
	Certificate *CertificateResponse `json:"certificate,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
}

// ToResponse converte User para UserResponse
func (u *User) ToResponse() *UserResponse {
	response := &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		UserType:  u.UserType,
		CPFCNPJ:   u.CPFCNPJ,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
	}

	if u.Certificate != nil {
		response.Certificate = u.Certificate.ToResponse()
	}

	return response
}

// TokenPair representa um par de tokens JWT
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// LoginRequest representa a requisição de login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest representa a requisição de registro
type RegisterRequest struct {
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,min=8"`
	Name     string   `json:"name" validate:"required"`
	UserType UserType `json:"user_type" validate:"required"`
	CPFCNPJ  string   `json:"cpf_cnpj" validate:"required"`
}

// UpdateProfileRequest representa a requisição de atualização de perfil
type UpdateProfileRequest struct {
	Name    string `json:"name" validate:"required"`
	CPFCNPJ string `json:"cpf_cnpj" validate:"required"`
}

// UpdateUserAdminRequest representa a requisição de atualização de usuário pelo admin
type UpdateUserAdminRequest struct {
	Name     *string   `json:"name" validate:"omitempty,min=3"`
	Email    *string   `json:"email" validate:"omitempty,email"`
	UserType *UserType `json:"user_type" validate:"omitempty"`
	CPFCNPJ  *string   `json:"cpf_cnpj" validate:"omitempty"`
	IsActive *bool     `json:"is_active" validate:"omitempty"`
}

// TokenResponse representa a resposta de autenticação
type TokenResponse struct {
	Token     string        `json:"token"`
	ExpiresIn int64         `json:"expires_in"`
	User      *UserResponse `json:"user"`
}

// BeforeCreate é executado antes de criar um usuário
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.IsActive = true
	return nil
}

// IsVeterinario verifica se o usuário é veterinário
func (u *User) IsVeterinario() bool {
	return u.UserType == UserTypeVeterinario
}

// IsCriador verifica se o usuário é criador
func (u *User) IsCriador() bool {
	return u.UserType == UserTypeCriador
}

// IsAdmin verifica se o usuário é administrador
func (u *User) IsAdmin() bool {
	return u.UserType == UserTypeAdmin
}
