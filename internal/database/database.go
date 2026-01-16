package database

import (
	"fmt"
	"log"
	"time"

	"github.com/equinoid/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewConnection cria uma nova conexão com o banco de dados
func NewConnection(databaseURL string) (*gorm.DB, error) {
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar com banco de dados: %w", err)
	}

	// Configurar pool de conexões
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("falha ao obter instância SQL DB: %w", err)
	}

	// Configurações do pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// RunMigrations executa todas as migrações
func RunMigrations(db *gorm.DB) error {
	log.Println("Executando migrações...")

	// Lista de modelos para migração
	models := []interface{}{
		// Modelos base
		&models.User{},
		&models.Certificate{},
		&models.Equino{},
		&models.Propriedade{},
		&models.EquinoVeterinario{},
		&models.Evento{},

		// Modelos de laboratório
		&models.LaboratorioDNA{},

		// Modelos de reprodução
		&models.Cobertura{},
		&models.AvaliacaoSemen{},
		&models.Gestacao{},
		&models.Ultrassonografia{},
		&models.PerformanceMaterna{},
		&models.RankingReprodutivo{},

		// Modelos de valorização
		&models.RegistroValorizacao{},
		&models.RankingValorizacao{},
		&models.Competicao{},
		&models.LeilaoValorizacao{},
		&models.LanceLeilao{},

		// Modelos de linhagem

		// Modelos do sistema social
		&models.PerfilSocial{},
		&models.PostSocial{},
		&models.InteracaoSocial{},
		&models.ComentarioSocial{},
		&models.IntegracaoInstagram{},
		&models.SeguirEquino{},
		&models.Oferta{},

		// Modelos adicionais
		&models.RegistroMidia{},
		&models.RegistroSaude{},
		&models.RegistroEducacao{},
		&models.RegistroViagem{},
		&models.RegistroParceria{},
		&models.RegistroAnalise{},
		&models.PerformanceReprodutiva{},
		&models.Webhook{},
		&models.ChatbotQuery{},
	}

	// Executar auto-migração para todos os modelos
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("falha ao migrar modelo %T: %w", model, err)
		}
	}

	// Criar índices customizados
	if err := createCustomIndexes(db); err != nil {
		return fmt.Errorf("falha ao criar índices customizados: %w", err)
	}

	log.Println("Migrações executadas com sucesso!")
	return nil
}

// createCustomIndexes cria índices customizados para otimização
func createCustomIndexes(db *gorm.DB) error {
	log.Println("Criando índices customizados...")

	// Índices para tabela equinos
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_equinos_proprietario_id ON equinos(proprietario_id)",
		"CREATE INDEX IF NOT EXISTS idx_equinos_propriedade_id ON equinos(propriedade_id)",
		"CREATE INDEX IF NOT EXISTS idx_equinos_genitora ON equinos(genitora)",
		"CREATE INDEX IF NOT EXISTS idx_equinos_genitor ON equinos(genitor)",
		"CREATE INDEX IF NOT EXISTS idx_equinos_raca ON equinos(raca)",
		"CREATE INDEX IF NOT EXISTS idx_equinos_status ON equinos(status)",
		"CREATE INDEX IF NOT EXISTS idx_equinos_data_nascimento ON equinos(data_nascimento)",

		// Índices para tabela eventos
		"CREATE INDEX IF NOT EXISTS idx_eventos_equino_id ON eventos(equino_id)",
		"CREATE INDEX IF NOT EXISTS idx_eventos_tipo_evento ON eventos(tipo_evento)",
		"CREATE INDEX IF NOT EXISTS idx_eventos_data_evento ON eventos(data_evento)",
		"CREATE INDEX IF NOT EXISTS idx_eventos_veterinario_id ON eventos(veterinario_id)",

		// Índices para tabela registros_valorizacao
		"CREATE INDEX IF NOT EXISTS idx_registros_valorizacao_equinoid_id ON registros_valorizacao(equinoid_id)",
		"CREATE INDEX IF NOT EXISTS idx_registros_valorizacao_categoria ON registros_valorizacao(categoria)",
		"CREATE INDEX IF NOT EXISTS idx_registros_valorizacao_data_registro ON registros_valorizacao(data_registro)",
		"CREATE INDEX IF NOT EXISTS idx_registros_valorizacao_status_validacao ON registros_valorizacao(status_validacao)",
		"CREATE INDEX IF NOT EXISTS idx_registros_valorizacao_criado_por ON registros_valorizacao(criado_por)",

		// Índices para tabela rankings_valorizacao
		"CREATE INDEX IF NOT EXISTS idx_rankings_valorizacao_equinoid_id ON rankings_valorizacao(equinoid_id)",
		"CREATE INDEX IF NOT EXISTS idx_rankings_valorizacao_tipo_ranking ON rankings_valorizacao(tipo_ranking)",
		"CREATE INDEX IF NOT EXISTS idx_rankings_valorizacao_posicao ON rankings_valorizacao(posicao)",
		"CREATE INDEX IF NOT EXISTS idx_rankings_valorizacao_data_ranking ON rankings_valorizacao(data_ranking)",

		// Índices para tabela coberturas
		"CREATE INDEX IF NOT EXISTS idx_coberturas_reprodutor_equinoid ON coberturas(reprodutor_equinoid)",
		"CREATE INDEX IF NOT EXISTS idx_coberturas_matriz_equinoid ON coberturas(matriz_equinoid)",
		"CREATE INDEX IF NOT EXISTS idx_coberturas_data_cobertura ON coberturas(data_cobertura)",
		"CREATE INDEX IF NOT EXISTS idx_coberturas_status_cobertura ON coberturas(status_cobertura)",

		// Índices para tabela gestacoes
		"CREATE INDEX IF NOT EXISTS idx_gestacoes_matriz_equinoid ON gestacoes(matriz_equinoid)",
		"CREATE INDEX IF NOT EXISTS idx_gestacoes_data_prevista_parto ON gestacoes(data_prevista_parto)",
		"CREATE INDEX IF NOT EXISTS idx_gestacoes_status_gestacao ON gestacoes(status_gestacao)",

		// Índices para tabela posts_sociais
		"CREATE INDEX IF NOT EXISTS idx_posts_sociais_equinoid_id ON posts_sociais(equinoid_id)",
		"CREATE INDEX IF NOT EXISTS idx_posts_sociais_data_postagem ON posts_sociais(data_postagem)",
		"CREATE INDEX IF NOT EXISTS idx_posts_sociais_status_post ON posts_sociais(status_post)",
		"CREATE INDEX IF NOT EXISTS idx_posts_sociais_criado_por ON posts_sociais(criado_por)",

		// Índices para tabela perfis_sociais
		"CREATE INDEX IF NOT EXISTS idx_perfis_sociais_equinoid_id ON perfis_sociais(equinoid_id)",
		"CREATE INDEX IF NOT EXISTS idx_perfis_sociais_status_disponibilidade ON perfis_sociais(status_disponibilidade)",
		"CREATE INDEX IF NOT EXISTS idx_perfis_sociais_total_seguidores ON perfis_sociais(total_seguidores)",

		// Índices para tabela validacoes_geneticas
		"CREATE INDEX IF NOT EXISTS idx_validacoes_geneticas_equinoid_id ON validacoes_geneticas(equinoid_id)",
		"CREATE INDEX IF NOT EXISTS idx_validacoes_geneticas_genitora_equinoid ON validacoes_geneticas(genitora_equinoid)",
		"CREATE INDEX IF NOT EXISTS idx_validacoes_geneticas_genitor_equinoid ON validacoes_geneticas(genitor_equinoid)",
		"CREATE INDEX IF NOT EXISTS idx_validacoes_geneticas_status ON validacoes_geneticas(status)",

		// Índices para tabela certificates
		"CREATE INDEX IF NOT EXISTS idx_certificates_user_id ON certificates(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_certificates_valid_to ON certificates(valid_to)",
		"CREATE INDEX IF NOT EXISTS idx_certificates_is_revoked ON certificates(is_revoked)",

		// Índices para tabela ofertas
		"CREATE INDEX IF NOT EXISTS idx_ofertas_equinoid_id ON ofertas(equinoid_id)",
		"CREATE INDEX IF NOT EXISTS idx_ofertas_ofertante_por_id ON ofertas(ofertante_por_id)",
		"CREATE INDEX IF NOT EXISTS idx_ofertas_status_oferta ON ofertas(status_oferta)",
		"CREATE INDEX IF NOT EXISTS idx_ofertas_prazo_oferta ON ofertas(prazo_oferta)",

		// Índices compostos para melhor performance
		"CREATE INDEX IF NOT EXISTS idx_equinos_proprietario_status ON equinos(proprietario_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_eventos_equino_tipo ON eventos(equino_id, tipo_evento)",
		"CREATE INDEX IF NOT EXISTS idx_registros_equinoid_categoria ON registros_valorizacao(equinoid_id, categoria)",
		"CREATE INDEX IF NOT EXISTS idx_posts_equinoid_status ON posts_sociais(equinoid_id, status_post)",
		"CREATE INDEX IF NOT EXISTS idx_coberturas_reprodutor_status ON coberturas(reprodutor_equinoid, status_cobertura)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("Aviso: falha ao criar índice: %v", err)
			// Não retorna erro para permitir que a aplicação continue
		}
	}

	log.Println("Índices customizados criados com sucesso!")
	return nil
}

// CreateInitialData cria dados iniciais se necessário
func CreateInitialData(db *gorm.DB) error {
	log.Println("Criando dados iniciais...")

	// Criar laboratórios de DNA padrão
	laboratories := []models.LaboratorioDNA{
		{
			Nome:               "Laboratório Genética Equina Brasil",
			Codigo:             "LGEB-001",
			Pais:               "076", // Brasil
			CertificacaoStatus: "ativo",
			ContatoEmail:       "contato@lgeb.com.br",
			ContatoTelefone:    "+55 11 99999-9999",
		},
		{
			Nome:               "International Equine Genetics Lab",
			Codigo:             "IEGL-001",
			Pais:               "840", // EUA
			CertificacaoStatus: "ativo",
			ContatoEmail:       "contact@iegl.com",
			ContatoTelefone:    "+1 555-123-4567",
		},
		{
			Nome:               "European Horse DNA Center",
			Codigo:             "EHDC-001",
			Pais:               "276", // Alemanha
			CertificacaoStatus: "ativo",
			ContatoEmail:       "info@ehdc.de",
			ContatoTelefone:    "+49 30 12345678",
		},
	}

	for _, lab := range laboratories {
		var existingLab models.LaboratorioDNA
		result := db.Where("codigo = ?", lab.Codigo).First(&existingLab)
		if result.Error != nil {
			if err := db.Create(&lab).Error; err != nil {
				log.Printf("Falha ao criar laboratório %s: %v", lab.Nome, err)
			} else {
				log.Printf("Laboratório criado: %s", lab.Nome)
			}
		}
	}

	log.Println("Dados iniciais criados com sucesso!")
	return nil
}
