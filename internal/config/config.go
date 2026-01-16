package config

import (
	"os"
	"strconv"
	"time"
)

// Config representa a configuração da aplicação
type Config struct {
	// Servidor
	Port        string
	Environment string
	GinMode     string

	// Banco de dados
	DatabaseURL string

	// Redis
	RedisURL     string
	RedisEnabled bool

	// JWT
	JWTSecret      string
	JWTExpireHours time.Duration

	// Segurança
	BcryptCost         int
	RateLimitPerMinute int

	// Upload de arquivos
	UploadMaxSize int64
	UploadPath    string

	// AWS S3
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSS3Bucket        string

	// Logs
	LogLevel string
	LogFile  string

	// Certificados digitais
	CACertPath string
	CAKeyPath  string

	// Monitoramento
	MetricsEnabled bool
	MetricsPath    string

	// Email
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromEmail    string

	// Integração externa
	GedaveAPIURL string
	GedaveAPIKey string

	// Blockchain
	BlockchainEnabled  bool
	EthereumRPCURL     string
	EthereumPrivateKey string

	// D4Sign
	D4SignAPIURL   string
	D4SignTokenAPI string
	D4SignCryptKey string
	D4SignSafeUUID string
}

// Load carrega as configurações a partir das variáveis de ambiente
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		GinMode:     getEnv("GIN_MODE", "debug"),

		DatabaseURL:  buildDatabaseURL(),
		RedisURL:     buildRedisURL(),
		RedisEnabled: getEnvAsBool("REDIS_ENABLED", true),

		JWTSecret:      getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		JWTExpireHours: time.Duration(getEnvAsInt("JWT_EXPIRE_HOURS", 24)) * time.Hour,

		BcryptCost:         getEnvAsInt("BCRYPT_COST", 12),
		RateLimitPerMinute: getEnvAsInt("RATE_LIMIT_PER_MINUTE", 100),

		UploadMaxSize: int64(getEnvAsInt("UPLOAD_MAX_SIZE", 10485760)), // 10MB
		UploadPath:    getEnv("UPLOAD_PATH", "./uploads"),

		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSS3Bucket:        getEnv("AWS_S3_BUCKET", "equinoid-documents"),

		LogLevel: getEnv("LOG_LEVEL", "info"),
		LogFile:  getEnv("LOG_FILE", "./logs/equinoid.log"),

		CACertPath: getEnv("CA_CERT_PATH", "./certs/ca-cert.pem"),
		CAKeyPath:  getEnv("CA_KEY_PATH", "./certs/ca-key.pem"),

		MetricsEnabled: getEnvAsBool("METRICS_ENABLED", true),
		MetricsPath:    getEnv("METRICS_PATH", "/metrics"),

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@equinoid.org"),

		GedaveAPIURL: getEnv("GEDAVE_API_URL", "https://gedave.gov.br/api"),
		GedaveAPIKey: getEnv("GEDAVE_API_KEY", ""),

		BlockchainEnabled:  getEnvAsBool("BLOCKCHAIN_ENABLED", false),
		EthereumRPCURL:     getEnv("ETHEREUM_RPC_URL", ""),
		EthereumPrivateKey: getEnv("ETHEREUM_PRIVATE_KEY", ""),

		D4SignAPIURL:   getEnv("D4SIGN_API_URL", "https://sandbox.d4sign.com.br/api/v1"),
		D4SignTokenAPI: getEnv("D4SIGN_TOKEN_API", ""),
		D4SignCryptKey: getEnv("D4SIGN_CRYPT_KEY", ""),
		D4SignSafeUUID: getEnv("D4SIGN_SAFE_UUID", ""),
	}
}

// getEnv obtém variável de ambiente ou retorna valor padrão
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt obtém variável de ambiente como inteiro ou retorna valor padrão
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsBool obtém variável de ambiente como boolean ou retorna valor padrão
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// buildDatabaseURL constrói a URL do banco de dados
func buildDatabaseURL() string {
	// Primeiro tenta usar DATABASE_URL diretamente (Supabase, Heroku, etc)
	if databaseURL := getEnv("DATABASE_URL", ""); databaseURL != "" {
		return databaseURL
	}

	// Se não tiver DATABASE_URL, constrói manualmente
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "equinoid")
	password := getEnv("DB_PASSWORD", "equinoid123")
	dbname := getEnv("DB_NAME", "equinoid")
	sslmode := getEnv("DB_SSL_MODE", "disable")

	return "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname + "?sslmode=" + sslmode
}

// buildRedisURL constrói a URL do Redis
func buildRedisURL() string {
	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")
	password := getEnv("REDIS_PASSWORD", "")
	db := getEnv("REDIS_DB", "0")

	if password != "" {
		return "redis://:" + password + "@" + host + ":" + port + "/" + db
	}
	return "redis://" + host + ":" + port + "/" + db
}
