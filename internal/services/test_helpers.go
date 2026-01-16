package services

import (
	"context"
	"testing"
	"time"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/pkg/logging"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *MockCache) DeletePattern(ctx context.Context, pattern string) error {
	return nil
}

func (m *MockCache) Exists(ctx context.Context, key string) bool {
	return false
}

func (m *MockCache) Increment(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *MockCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return true, nil
}

func (m *MockCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}

func (m *MockCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, nil
}

func (m *MockCache) HSet(ctx context.Context, key string, field string, value interface{}) error {
	return nil
}

func (m *MockCache) HGet(ctx context.Context, key string, field string, dest interface{}) error {
	return nil
}

func (m *MockCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return nil, nil
}

func (m *MockCache) HDel(ctx context.Context, key string, fields ...string) error {
	return nil
}

func (m *MockCache) ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	return nil
}

func (m *MockCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return nil, nil
}

func (m *MockCache) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return nil, nil
}

func (m *MockCache) ZRank(ctx context.Context, key string, member string) (int64, error) {
	return 0, nil
}

func (m *MockCache) ZScore(ctx context.Context, key string, member string) (float64, error) {
	return 0, nil
}

func (m *MockCache) Close() error {
	return nil
}

func (m *MockCache) Ping(ctx context.Context) *redis.StatusCmd {
	return nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Skip("sqlite driver unavailable for tests")
	}

	db.AutoMigrate(&models.User{}, &models.Equino{})
	return db
}

func newTestLogger() *logging.Logger {
	return &logging.Logger{}
}
