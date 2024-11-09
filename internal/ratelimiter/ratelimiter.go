package ratelimiter

import (
	"context"
	"time"

	"github.com/erickmanovei/goexpert_rate_limiter/internal/utils"
	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	Rdb            *redis.Client
	IpRateLimit    int
	TokenRateLimit int
	BlockDuration  time.Duration
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	ipRateLimit := utils.GetEnvAsInt("IP_RATE_LIMIT", 5)
	tokenRateLimit := utils.GetEnvAsInt("TOKEN_RATE_LIMIT", 100)
	blockDuration := time.Duration(utils.GetEnvAsInt("BLOCK_DURATION", 300)) * time.Second

	return &RateLimiter{
		Rdb:            rdb,
		IpRateLimit:    ipRateLimit,
		TokenRateLimit: tokenRateLimit,
		BlockDuration:  blockDuration,
	}
}

func (rl *RateLimiter) IsRateLimited(ctx context.Context, key string, rateLimit int) (bool, error) {
	current, err := rl.Rdb.Get(ctx, key).Int()
	if err == redis.Nil {
		rl.Rdb.Set(ctx, key, 1, time.Second).Err()
		return false, nil
	} else if err != nil {
		return false, err
	}

	if current >= rateLimit {
		rl.Rdb.Expire(ctx, key, rl.BlockDuration)
		return true, nil
	}

	rl.Rdb.Incr(ctx, key)
	return false, nil
}
