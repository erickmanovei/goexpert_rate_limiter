package test

import (
	"context"
	"testing"
	"time"

	"github.com/erickmanovei/goexpert_rate_limiter/internal/ratelimiter"
	"github.com/erickmanovei/goexpert_rate_limiter/internal/strategy"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func setupRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Certifique-se de que o Redis está rodando nesta configuração
	})
}

func cleanupRedis(rdb *redis.Client) {
	rdb.FlushAll(context.Background())
}

func TestIncrementCount(t *testing.T) {
	rdb := setupRedis()
	defer cleanupRedis(rdb)

	strategy := strategy.NewRedisStrategy(rdb)
	ctx := context.Background()
	key := "test-ip:1234"

	count, err := strategy.IncrementCount(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	count, err = strategy.IncrementCount(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestSetAndCheckBlockDuration(t *testing.T) {
	rdb := setupRedis()
	defer cleanupRedis(rdb)

	strategy := strategy.NewRedisStrategy(rdb)
	ctx := context.Background()
	key := "test-ip:1234"

	// Bloqueia o IP por 2 segundos
	err := strategy.SetBlockDuration(ctx, key, 2*time.Second)
	assert.NoError(t, err)

	// Verifica se o IP está bloqueado imediatamente
	blocked, err := strategy.IsBlocked(ctx, key)
	assert.NoError(t, err)
	assert.True(t, blocked)

	// Espera 3 segundos para garantir que o bloqueio expirou
	time.Sleep(3 * time.Second)
	blocked, err = strategy.IsBlocked(ctx, key)
	assert.NoError(t, err)
	assert.False(t, blocked)
}

func TestRateLimitingByIP(t *testing.T) {
	rdb := setupRedis()
	defer cleanupRedis(rdb)

	rateLimiter := ratelimiter.NewRateLimiter(rdb)
	rateLimiter.IpRateLimit = 5
	ctx := context.Background()
	key := "ip:test-ip:1234"

	// Envia 5 requisições dentro do limite
	for i := 0; i < 5; i++ {
		rateLimited, err := rateLimiter.IsRateLimited(ctx, key, rateLimiter.IpRateLimit)
		assert.NoError(t, err)
		assert.False(t, rateLimited)
	}

	// Sexta requisição deve ser limitada
	rateLimited, err := rateLimiter.IsRateLimited(ctx, key, rateLimiter.IpRateLimit)
	assert.NoError(t, err)
	assert.True(t, rateLimited)
}

func TestRateLimitingByToken(t *testing.T) {
	rdb := setupRedis()
	defer cleanupRedis(rdb)

	rateLimiter := ratelimiter.NewRateLimiter(rdb)
	rateLimiter.TokenRateLimit = 10
	ctx := context.Background()
	key := "token:test-token:abc123"

	// Envia 10 requisições dentro do limite
	for i := 0; i < 10; i++ {
		rateLimited, err := rateLimiter.IsRateLimited(ctx, key, rateLimiter.TokenRateLimit)
		assert.NoError(t, err)
		assert.False(t, rateLimited)
	}

	// Décima primeira requisição deve ser limitada
	rateLimited, err := rateLimiter.IsRateLimited(ctx, key, rateLimiter.TokenRateLimit)
	assert.NoError(t, err)
	assert.True(t, rateLimited)
}

func TestAutomaticUnblockAfterBlockDuration(t *testing.T) {
	rdb := setupRedis()
	defer cleanupRedis(rdb)

	rateLimiter := ratelimiter.NewRateLimiter(rdb)
	rateLimiter.IpRateLimit = 2
	rateLimiter.BlockDuration = 2 * time.Second
	ctx := context.Background()
	key := "ip:test-ip:5678"

	// Excede o limite
	for i := 0; i < 3; i++ {
		rateLimited, err := rateLimiter.IsRateLimited(ctx, key, rateLimiter.IpRateLimit)
		if i < 2 {
			assert.NoError(t, err)
			assert.False(t, rateLimited)
		} else {
			assert.NoError(t, err)
			assert.True(t, rateLimited)
		}
	}

	// Espera o tempo de bloqueio expirar
	time.Sleep(3 * time.Second)

	// Verifica se o bloqueio foi removido
	rateLimited, err := rateLimiter.IsRateLimited(ctx, key, rateLimiter.IpRateLimit)
	assert.NoError(t, err)
	assert.False(t, rateLimited)
}
