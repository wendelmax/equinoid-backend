package models

import (
	"time"

	"gorm.io/gorm"
)

// PerfilSocial representa um perfil social de equino
type PerfilSocial struct {
	ID                     uint                  `json:"id" gorm:"primaryKey"`
	Equinoid               string                `json:"equinoid" gorm:"size:25;not null"`
	NomePerfil             string                `json:"nome_perfil" gorm:"size:100"`
	Bio                    string                `json:"bio" gorm:"type:text"`
	Localizacao            string                `json:"localizacao" gorm:"size:200"`
	StatusDisponibilidade  StatusDisponibilidade `json:"status_disponibilidade" gorm:"default:'disponivel'"`
	TipoPerfil             TipoPerfil            `json:"tipo_perfil" gorm:"default:'publico'"`
	TotalSeguidores        int                   `json:"total_seguidores" gorm:"default:0"`
	TotalSeguindo          int                   `json:"total_seguindo" gorm:"default:0"`
	TotalPosts             int                   `json:"total_posts" gorm:"default:0"`
	TotalCurtidas          int                   `json:"total_curtidas" gorm:"default:0"`
	TotalComentarios       int                   `json:"total_comentarios" gorm:"default:0"`
	TotalCompartilhamentos int                   `json:"total_compartilhamentos" gorm:"default:0"`
	MostrarLocalizacao     bool                  `json:"mostrar_localizacao" gorm:"default:true"`
	PermitirOfertas        bool                  `json:"permitir_ofertas" gorm:"default:true"`
	PermitirContato        bool                  `json:"permitir_contato" gorm:"default:true"`
	PermitirSeguir         bool                  `json:"permitir_seguir" gorm:"default:true"`
	CriadoPor              uint                  `json:"criado_por" gorm:"not null"`
	CreatedAt              time.Time             `json:"created_at"`
	UpdatedAt              time.Time             `json:"updated_at"`
	DeletedAt              gorm.DeletedAt        `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Equino              *Equino              `json:"equino,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
	Criador             *User                `json:"criador,omitempty" gorm:"foreignKey:CriadoPor"`
	Posts               []PostSocial         `json:"posts,omitempty" gorm:"foreignKey:PerfilSocialID"`
	IntegracaoInstagram *IntegracaoInstagram `json:"integracao_instagram,omitempty" gorm:"foreignKey:PerfilSocialID"`
}

// StatusDisponibilidade define o status de disponibilidade
type StatusDisponibilidade string

const (
	StatusDisponivel   StatusDisponibilidade = "disponivel"
	StatusDispVendido  StatusDisponibilidade = "vendido"
	StatusIndisponivel StatusDisponibilidade = "indisponivel"
)

// TipoPerfil define o tipo de perfil
type TipoPerfil string

const (
	TipoPerfilPublico   TipoPerfil = "publico"
	TipoPerfilPrivado   TipoPerfil = "privado"
	TipoPerfilComercial TipoPerfil = "comercial"
)

// PostSocial representa um post social
type PostSocial struct {
	ID                       uint           `json:"id" gorm:"primaryKey"`
	Equinoid                 string         `json:"equinoid" gorm:"size:25;not null"`
	PerfilSocialID           uint           `json:"perfil_social_id" gorm:"not null"`
	TipoConteudo             TipoConteudo   `json:"tipo_conteudo" gorm:"not null"`
	Legenda                  string         `json:"legenda" gorm:"type:text"`
	LocalizacaoPost          string         `json:"localizacao_post" gorm:"size:200"`
	ArquivosMidia            JSONB          `json:"arquivos_midia" gorm:"type:jsonb"`
	ThumbnailURL             string         `json:"thumbnail_url" gorm:"size:500"`
	DuracaoVideo             *int           `json:"duracao_video"`
	DataPostagem             time.Time      `json:"data_postagem" gorm:"default:CURRENT_TIMESTAMP"`
	DataExpiracao            *time.Time     `json:"data_expiracao"`
	StatusPost               StatusPost     `json:"status_post" gorm:"default:'ativo'"`
	TotalCurtidas            int            `json:"total_curtidas" gorm:"default:0"`
	TotalComentarios         int            `json:"total_comentarios" gorm:"default:0"`
	TotalCompartilhamentos   int            `json:"total_compartilhamentos" gorm:"default:0"`
	TotalVisualizacoes       int            `json:"total_visualizacoes" gorm:"default:0"`
	PermitirComentarios      bool           `json:"permitir_comentarios" gorm:"default:true"`
	PermitirCompartilhamento bool           `json:"permitir_compartilhamento" gorm:"default:true"`
	CriadoPor                uint           `json:"criado_por" gorm:"not null"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Equino       *Equino            `json:"equino,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
	PerfilSocial *PerfilSocial      `json:"perfil_social,omitempty" gorm:"foreignKey:PerfilSocialID"`
	Criador      *User              `json:"criador,omitempty" gorm:"foreignKey:CriadoPor"`
	Interacoes   []InteracaoSocial  `json:"interacoes,omitempty" gorm:"foreignKey:PostID"`
	Comentarios  []ComentarioSocial `json:"comentarios,omitempty" gorm:"foreignKey:PostID"`
}

// TipoConteudo define o tipo de conteúdo do post
type TipoConteudo string

const (
	TipoConteudoFoto      TipoConteudo = "foto"
	TipoConteudoVideo     TipoConteudo = "video"
	TipoConteudoCarrossel TipoConteudo = "carrossel"
	TipoConteudoStory     TipoConteudo = "story"
	TipoConteudoReel      TipoConteudo = "reel"
)

// StatusPost define o status do post
type StatusPost string

const (
	StatusPostAtivo     StatusPost = "ativo"
	StatusPostArquivado StatusPost = "arquivado"
	StatusPostRemovido  StatusPost = "removido"
)

// InteracaoSocial representa uma interação social
type InteracaoSocial struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	PostID        uint           `json:"post_id" gorm:"not null"`
	UserID        uint           `json:"user_id" gorm:"not null"`
	TipoInteracao TipoInteracao  `json:"tipo_interacao" gorm:"not null"`
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Post *PostSocial `json:"post,omitempty" gorm:"foreignKey:PostID"`
	User *User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TipoInteracao define o tipo de interação
type TipoInteracao string

const (
	TipoInteracaoCurtida          TipoInteracao = "curtida"
	TipoInteracaoCompartilhamento TipoInteracao = "compartilhamento"
	TipoInteracaoSalvar           TipoInteracao = "salvar"
	TipoInteracaoInteresse        TipoInteracao = "interesse"
	TipoInteracaoAmor             TipoInteracao = "amor"
	TipoInteracaoRisada           TipoInteracao = "risada"
	TipoInteracaoSurpresa         TipoInteracao = "surpresa"
)

// ComentarioSocial representa um comentário social
type ComentarioSocial struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	PostID    uint           `json:"post_id" gorm:"not null"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	ParentID  *uint          `json:"parent_id"`
	Conteudo  string         `json:"conteudo" gorm:"type:text;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Post      *PostSocial        `json:"post,omitempty" gorm:"foreignKey:PostID"`
	User      *User              `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Parent    *ComentarioSocial  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Respostas []ComentarioSocial `json:"respostas,omitempty" gorm:"foreignKey:ParentID"`
}

// IntegracaoInstagram representa a integração com Instagram
type IntegracaoInstagram struct {
	ID                   uint             `json:"id" gorm:"primaryKey"`
	Equinoid             string           `json:"equinoid" gorm:"size:25;not null"`
	PerfilSocialID       uint             `json:"perfil_social_id" gorm:"not null"`
	InstagramUsername    string           `json:"instagram_username" gorm:"size:100"`
	InstagramUserID      string           `json:"instagram_user_id" gorm:"size:50"`
	AccessToken          string           `json:"access_token" gorm:"type:text"`
	RefreshToken         string           `json:"refresh_token" gorm:"type:text"`
	TokenExpiresAt       *time.Time       `json:"token_expires_at"`
	AutoPost             bool             `json:"auto_post" gorm:"default:true"`
	AutoStories          bool             `json:"auto_stories" gorm:"default:true"`
	AutoReels            bool             `json:"auto_reels" gorm:"default:true"`
	HorarioPostagem      time.Time        `json:"horario_postagem" gorm:"type:time;default:'18:00:00'"`
	FusoHorario          string           `json:"fuso_horario" gorm:"size:50;default:'America/Sao_Paulo'"`
	IncluirHashtags      bool             `json:"incluir_hashtags" gorm:"default:true"`
	IncluirLegenda       bool             `json:"incluir_legenda" gorm:"default:true"`
	IncluirLocalizacao   bool             `json:"incluir_localizacao" gorm:"default:true"`
	IdiomaLegenda        IdiomaLegenda    `json:"idioma_legenda" gorm:"default:'pt'"`
	TotalPosts           int              `json:"total_posts" gorm:"default:0"`
	TotalStories         int              `json:"total_stories" gorm:"default:0"`
	TotalReels           int              `json:"total_reels" gorm:"default:0"`
	TotalAlcance         int              `json:"total_alcance" gorm:"default:0"`
	TotalEngajamento     int              `json:"total_engajamento" gorm:"default:0"`
	StatusIntegracao     StatusIntegracao `json:"status_integracao" gorm:"default:'ativa'"`
	UltimaSincronizacao  *time.Time       `json:"ultima_sincronizacao"`
	ProximaSincronizacao *time.Time       `json:"proxima_sincronizacao"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
	DeletedAt            gorm.DeletedAt   `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Equino       *Equino       `json:"equino,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
	PerfilSocial *PerfilSocial `json:"perfil_social,omitempty" gorm:"foreignKey:PerfilSocialID"`
}

// IdiomaLegenda define o idioma da legenda
type IdiomaLegenda string

const (
	IdiomaPortugues       IdiomaLegenda = "pt"
	IdiomaIngles          IdiomaLegenda = "en"
	IdiomaPortuguesIngles IdiomaLegenda = "pt-en"
)

// StatusIntegracao define o status da integração
type StatusIntegracao string

const (
	StatusIntegracaoAtiva    StatusIntegracao = "ativa"
	StatusIntegracaoPausada  StatusIntegracao = "pausada"
	StatusIntegracaoErro     StatusIntegracao = "erro"
	StatusIntegracaoExpirada StatusIntegracao = "expirada"
)

// SeguirEquino representa o relacionamento de seguir entre usuários e equinos
type SeguirEquino struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Equinoid  string         `json:"equinoid" gorm:"size:25;not null"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	User   *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Equino *Equino `json:"equino,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
}

// Oferta representa uma oferta feita por um equino
type Oferta struct {
	ID                   uint           `json:"id" gorm:"primaryKey"`
	Equinoid             string         `json:"equinoid" gorm:"size:25;not null"`
	OfertantePorID       uint           `json:"ofertante_por_id" gorm:"not null"`
	TipoOferta           TipoOferta     `json:"tipo_oferta" gorm:"not null"`
	ValorOferta          float64        `json:"valor_oferta" gorm:"type:decimal(15,2);not null"`
	Moeda                string         `json:"moeda" gorm:"size:3;default:'BRL'"`
	CondicoesOferta      string         `json:"condicoes_oferta" gorm:"type:text"`
	PrazoOferta          *time.Time     `json:"prazo_oferta"`
	StatusOferta         StatusOferta   `json:"status_oferta" gorm:"default:'pendente'"`
	RespostaProprietario string         `json:"resposta_proprietario" gorm:"type:text"`
	DataResposta         *time.Time     `json:"data_resposta"`
	ObservacoesInternas  string         `json:"observacoes_internas" gorm:"type:text"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string"`

	// Relacionamentos
	Equino    *Equino `json:"equino,omitempty" gorm:"foreignKey:Equinoid;references:Equinoid"`
	Ofertante *User   `json:"ofertante,omitempty" gorm:"foreignKey:OfertantePorID"`
}

// TipoOferta define o tipo de oferta
type TipoOferta string

const (
	TipoOfertaCompra       TipoOferta = "compra"
	TipoOfertaCobertura    TipoOferta = "cobertura"
	TipoOfertaParticipacao TipoOferta = "participacao"
	TipoOfertaSociedade    TipoOferta = "sociedade"
	TipoOfertaAluguel      TipoOferta = "aluguel"
)

// StatusOferta define o status da oferta
type StatusOferta string

const (
	StatusOfertaPendente  StatusOferta = "pendente"
	StatusOfertaAceita    StatusOferta = "aceita"
	StatusOfertaRecusada  StatusOferta = "recusada"
	StatusOfertaExpirada  StatusOferta = "expirada"
	StatusOfertaCancelada StatusOferta = "cancelada"
)

// Request DTOs para sistema social
type CreatePerfilSocialRequest struct {
	NomePerfil            string                `json:"nome_perfil"`
	Bio                   string                `json:"bio"`
	Localizacao           string                `json:"localizacao"`
	StatusDisponibilidade StatusDisponibilidade `json:"status_disponibilidade"`
	TipoPerfil            TipoPerfil            `json:"tipo_perfil"`
	MostrarLocalizacao    bool                  `json:"mostrar_localizacao"`
	PermitirOfertas       bool                  `json:"permitir_ofertas"`
	PermitirContato       bool                  `json:"permitir_contato"`
	PermitirSeguir        bool                  `json:"permitir_seguir"`
}

type CreatePostRequest struct {
	TipoConteudo             TipoConteudo      `json:"tipo_conteudo" validate:"required"`
	Legenda                  string            `json:"legenda"`
	LocalizacaoPost          string            `json:"localizacao_post"`
	ArquivosMidia            []DocumentoEvento `json:"arquivos_midia"`
	DataExpiracao            *time.Time        `json:"data_expiracao"`
	PermitirComentarios      bool              `json:"permitir_comentarios"`
	PermitirCompartilhamento bool              `json:"permitir_compartilhamento"`
}

type CreateInteracaoRequest struct {
	TipoInteracao TipoInteracao `json:"tipo_interacao" validate:"required"`
}

type CreateOfertaRequest struct {
	TipoOferta      TipoOferta `json:"tipo_oferta" validate:"required"`
	ValorOferta     float64    `json:"valor_oferta" validate:"required,gt=0"`
	Moeda           string     `json:"moeda"`
	CondicoesOferta string     `json:"condicoes_oferta"`
	PrazoOferta     *time.Time `json:"prazo_oferta"`
}
