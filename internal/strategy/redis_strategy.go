package strategy

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStrategy define a estrutura de uma estratégia de persistência usando Redis.
type RedisStrategy struct {
	client *redis.Client
}

// NewRedisStrategy cria uma nova instância de RedisStrategy.
func NewRedisStrategy(client *redis.Client) *RedisStrategy {
	return &RedisStrategy{client: client}
}

// IncrementCount incrementa o contador de requisições para a chave especificada e define uma expiração de 1 segundo.
func (r *RedisStrategy) IncrementCount(ctx context.Context, key string) (int, error) {
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Definir a expiração da chave para garantir a renovação da janela de tempo de 1 segundo
	if count == 1 {
		r.client.Expire(ctx, key, time.Second)
	}
	return int(count), nil
}

// SetBlockDuration define o tempo de bloqueio para uma chave específica (IP ou Token).
func (r *RedisStrategy) SetBlockDuration(ctx context.Context, key string, duration time.Duration) error {
	return r.client.Expire(ctx, key, duration).Err()
}

// IsBlocked verifica se uma chave específica está bloqueada.
func (r *RedisStrategy) IsBlocked(ctx context.Context, key string) (bool, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return ttl > 0, nil
}

// ResetCount zera o contador de requisições para a chave especificada.
func (r *RedisStrategy) ResetCount(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
