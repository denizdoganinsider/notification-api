package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

func RateLimiterMiddleware(redisClient *redis.Client, limit int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			key := fmt.Sprintf("rate_limit:%s", ip)

			ctx := context.Background()

			count, err := redisClient.Incr(ctx, key).Result()
			if err != nil {
				return next(c)
			}

			if count == 1 {
				redisClient.Expire(ctx, key, window)
			}

			if count > int64(limit) {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "rate limit exceeded",
				})
			}

			return next(c)
		}
	}
}
