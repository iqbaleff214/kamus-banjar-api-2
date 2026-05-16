package ratelimit

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/httperr"
	"github.com/redis/go-redis/v9"
)

// Counter increments a named key and returns the new count.
// Implementations must set the key TTL to window on first increment.
type Counter interface {
	Increment(ctx context.Context, key string, window time.Duration) (int64, error)
}

// RedisCounter implements Counter using Redis INCR + EXPIRE.
type RedisCounter struct {
	client *redis.Client
}

func NewRedisCounter(client *redis.Client) *RedisCounter {
	return &RedisCounter{client: client}
}

func (r *RedisCounter) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
	pipe := r.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// Limiter returns a Fiber middleware that enforces a sliding counter rate limit.
// Fails open (passes through) if the counter returns an error (e.g. Redis down).
func Limiter(counter Counter, keyFn func(*fiber.Ctx) string, limit int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		count, err := counter.Increment(c.Context(), keyFn(c), window)
		if err != nil {
			return c.Next()
		}
		if count > int64(limit) {
			return httperr.Send(c, fiber.StatusTooManyRequests, httperr.RateLimited("rate limit exceeded"))
		}
		return c.Next()
	}
}
