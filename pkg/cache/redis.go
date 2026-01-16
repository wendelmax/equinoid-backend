package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient representa um cliente Redis
type RedisClient struct {
	client *redis.Client
}

// CacheInterface define a interface para operações de cache
type CacheInterface interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) bool
	Increment(ctx context.Context, key string) (int64, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	HSet(ctx context.Context, key string, field string, value interface{}) error
	HGet(ctx context.Context, key string, field string, dest interface{}) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) error
	ZAdd(ctx context.Context, key string, score float64, member interface{}) error
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRank(ctx context.Context, key string, member string) (int64, error)
	ZScore(ctx context.Context, key string, member string) (float64, error)
	Close() error
	Ping(ctx context.Context) *redis.StatusCmd
}

// NewRedisClient cria um novo cliente Redis
func NewRedisClient(redisURL string) *RedisClient {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		// Fallback para configuração padrão se URL for inválida
		opts = &redis.Options{
			Addr: "localhost:6379",
			DB:   0,
		}
	}

	// Configurações adicionais
	opts.MaxRetries = 3
	opts.PoolSize = 10
	opts.MinIdleConns = 5
	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second
	opts.PoolTimeout = 4 * time.Second

	rdb := redis.NewClient(opts)

	return &RedisClient{client: rdb}
}

// Set armazena um valor no cache com expiração
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, jsonData, expiration).Err()
}

// Get obtém um valor do cache
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// Delete remove uma chave do cache
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// DeletePattern remove todas as chaves que correspondem ao padrão
func (r *RedisClient) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return r.client.Del(ctx, keys...).Err()
}

// Exists verifica se uma chave existe
func (r *RedisClient) Exists(ctx context.Context, key string) bool {
	result, err := r.client.Exists(ctx, key).Result()
	return err == nil && result == 1
}

// Increment incrementa o valor de uma chave numérica
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// SetNX define uma chave apenas se ela não existir
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	return r.client.SetNX(ctx, key, jsonData, expiration).Result()
}

// Expire define o tempo de expiração para uma chave
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL obtém o tempo restante de vida de uma chave
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// HSet define um campo em um hash
func (r *RedisClient) HSet(ctx context.Context, key string, field string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.HSet(ctx, key, field, jsonData).Err()
}

// HGet obtém um campo de um hash
func (r *RedisClient) HGet(ctx context.Context, key string, field string, dest interface{}) error {
	val, err := r.client.HGet(ctx, key, field).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// HGetAll obtém todos os campos de um hash
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel remove campos de um hash
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// ZAdd adiciona um membro a um sorted set
func (r *RedisClient) ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	jsonData, err := json.Marshal(member)
	if err != nil {
		return err
	}

	return r.client.ZAdd(ctx, key, &redis.Z{
		Score:  score,
		Member: string(jsonData),
	}).Err()
}

// ZRange obtém membros de um sorted set por range de índice
func (r *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRevRange obtém membros de um sorted set por range de índice (ordem reversa)
func (r *RedisClient) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRevRange(ctx, key, start, stop).Result()
}

// ZRank obtém o rank de um membro em um sorted set
func (r *RedisClient) ZRank(ctx context.Context, key string, member string) (int64, error) {
	return r.client.ZRank(ctx, key, member).Result()
}

// ZScore obtém o score de um membro em um sorted set
func (r *RedisClient) ZScore(ctx context.Context, key string, member string) (float64, error) {
	return r.client.ZScore(ctx, key, member).Result()
}

// Close fecha a conexão com o Redis
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Ping testa a conexão com o Redis
func (r *RedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	return r.client.Ping(ctx)
}

// ErrCacheMiss é retornado quando uma chave não é encontrada no cache
var ErrCacheMiss = errors.New("cache miss")

// CacheKeys define as chaves utilizadas no cache
type CacheKeys struct {
	// Usuários
	UserByID    string
	UserByEmail string

	// Equinos
	EquinoByID        string
	EquinosByOwner    string
	EquinoSearch      string
	LinhagemEquino    string
	ValorizacaoEquino string

	// Rankings
	RankingNacional      string
	RankingInternacional string
	RankingPorCategoria  string

	// Sessões
	UserSession  string
	JWTBlacklist string

	// Rate limiting
	RateLimit string

	// Sistema social
	PerfilSocial     string
	PostsSociais     string
	SeguidoresEquino string
	OfertasEquino    string

	// Sistema reprodutivo
	CoberturasReprodutor string
	GestacaosMatriz      string
	RankingReprodutivo   string

	// Cache de consultas
	DashboardStats string
	SearchResults  string
}

// NewCacheKeys retorna uma instância das chaves de cache
func NewCacheKeys() *CacheKeys {
	return &CacheKeys{
		// Usuários
		UserByID:    "user:id:%d",
		UserByEmail: "user:email:%s",

		// Equinos
		EquinoByID:        "equino:id:%d",
		EquinosByOwner:    "equinos:owner:%d",
		EquinoSearch:      "equino:search:%s",
		LinhagemEquino:    "equino:linhagem:%s",
		ValorizacaoEquino: "equino:valorizacao:%s",

		// Rankings
		RankingNacional:      "ranking:nacional:%s",
		RankingInternacional: "ranking:internacional:%s",
		RankingPorCategoria:  "ranking:categoria:%s",

		// Sessões
		UserSession:  "session:user:%d",
		JWTBlacklist: "jwt:blacklist:%s",

		// Rate limiting
		RateLimit: "rate:limit:%s",

		// Sistema social
		PerfilSocial:     "social:perfil:%s",
		PostsSociais:     "social:posts:%s",
		SeguidoresEquino: "social:seguidores:%s",
		OfertasEquino:    "social:ofertas:%s",

		// Sistema reprodutivo
		CoberturasReprodutor: "reproducao:coberturas:%s",
		GestacaosMatriz:      "reproducao:gestacoes:%s",
		RankingReprodutivo:   "reproducao:ranking:%s",

		// Cache de consultas
		DashboardStats: "stats:dashboard:%d",
		SearchResults:  "search:results:%s",
	}
}

// CacheService provê operações de cache de alto nível
type CacheService struct {
	client CacheInterface
	keys   *CacheKeys
}

// NewCacheService cria um novo serviço de cache
func NewCacheService(client CacheInterface) *CacheService {
	return &CacheService{
		client: client,
		keys:   NewCacheKeys(),
	}
}

// CacheUser armazena um usuário no cache
func (s *CacheService) CacheUser(ctx context.Context, user interface{}) error {
	// Implementação específica será adicionada conforme necessário
	return nil
}

// GetUser obtém um usuário do cache
func (s *CacheService) GetUser(ctx context.Context, userID uint) (interface{}, error) {
	// Implementação específica será adicionada conforme necessário
	return nil, ErrCacheMiss
}

// InvalidateUserCache invalida o cache de um usuário
func (s *CacheService) InvalidateUserCache(ctx context.Context, userID uint) error {
	// Implementação específica será adicionada conforme necessário
	return nil
}

// CacheEquino armazena um equino no cache
func (s *CacheService) CacheEquino(ctx context.Context, equino interface{}) error {
	// Implementação específica será adicionada conforme necessário
	return nil
}

// GetEquino obtém um equino do cache
func (s *CacheService) GetEquino(ctx context.Context, equinoID uint) (interface{}, error) {
	// Implementação específica será adicionada conforme necessário
	return nil, ErrCacheMiss
}

// InvalidateEquinoCache invalida o cache de um equino
func (s *CacheService) InvalidateEquinoCache(ctx context.Context, equinoID uint) error {
	// Implementação específica será adicionada conforme necessário
	return nil
}
