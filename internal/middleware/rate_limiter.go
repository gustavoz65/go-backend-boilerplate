package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/repository"
)

var rateLimitScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return count
`)

func RateLimit(rdb *redis.Client, config repository.RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if rdb == nil {
				return next(c)
			}
			ctx := c.Request().Context()

			key := ""
			if config.KeyFunc != nil {
				key = config.KeyFunc(c)
			}
			if key == "" {
				key = fmt.Sprintf("ip:%s:path:%s", c.RealIP(), c.Path())
			}
			redisKey := fmt.Sprintf("ratelimit:%s", key)

			// Incrementa o contador e define a expiração se for a primeira vez
			durationSecs := int64(config.Duration.Seconds())
			count, err := rateLimitScript.Run(ctx, rdb, []string{redisKey}, durationSecs).Int()
			if err != nil {
				if config.SkipOnerror {
					return next(c)
				}
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"error":   "Error checking rate limit",
				})
			}

			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Max))

			if count > config.Max {
				c.Response().Header().Set("Retry-After", fmt.Sprintf("%.0f", config.Duration.Seconds()))
				c.Response().Header().Set("X-RateLimit-Remaining", "0")

				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"success": false,
					"error":   "Rate limit exceeded. Please try again later.",
				})
			}

			remaining := config.Max - count
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			return next(c)
		}
	}
}

func DefaultRateLimit(rdb *redis.Client) echo.MiddlewareFunc {
	return RateLimit(rdb, repository.RateLimiter{
		Max:      300,
		Duration: time.Minute,
		KeyFunc: func(c echo.Context) string {
			return fmt.Sprintf("ip:%s:path:%s", c.RealIP(), c.Path())
		},
		Endpoint:    "global",
		SkipOnerror: true,
	})
}

func UserRateLimit(rdb *redis.Client) echo.MiddlewareFunc {
	return RateLimit(rdb, repository.RateLimiter{
		Max:      1000,
		Duration: time.Minute,
		KeyFunc: func(c echo.Context) string {
			if userID := c.Get("user_id"); userID != nil {
				if uid, ok := userID.(uuid.UUID); ok {
					return fmt.Sprintf("user:%s", uid.String())
				}
			}
			return fmt.Sprintf("ip:%s", c.RealIP())
		},
		Endpoint:    "user",
		SkipOnerror: true,
	})
}

func AuthRateLimit(rdb *redis.Client) echo.MiddlewareFunc {
	return RateLimit(rdb, repository.RateLimiter{
		Max:      5,
		Duration: 15 * time.Minute,
		KeyFunc: func(c echo.Context) string {
			return fmt.Sprintf("auth:%s", c.RealIP())
		},
		Endpoint:    "auth",
		SkipOnerror: true,
	})
}

// MutationRateLimit aplica rate limiting para operações de escrita (POST/PUT/PATCH/DELETE)
func MutationRateLimit(rdb *redis.Client) echo.MiddlewareFunc {
	return RateLimit(rdb, repository.RateLimiter{
		Max:      100, // 100 mutations por minuto
		Duration: time.Minute,
		KeyFunc: func(c echo.Context) string {
			// Usar user_id se autenticado, senão IP
			if userID := c.Get("user_id"); userID != nil {
				if uid, ok := userID.(uuid.UUID); ok {
					return fmt.Sprintf("mutations:user:%s", uid.String())
				}
			}
			return fmt.Sprintf("mutations:ip:%s", c.RealIP())
		},
		Endpoint:    "mutations",
		SkipOnerror: true,
	})
}

// ReadRateLimit aplica rate limiting para operações de leitura (GET)
func ReadRateLimit(rdb *redis.Client) echo.MiddlewareFunc {
	return RateLimit(rdb, repository.RateLimiter{
		Max:      500, // 500 reads por minuto
		Duration: time.Minute,
		KeyFunc: func(c echo.Context) string {
			// Usar user_id se autenticado, senão IP
			if userID := c.Get("user_id"); userID != nil {
				if uid, ok := userID.(uuid.UUID); ok {
					return fmt.Sprintf("reads:user:%s", uid.String())
				}
			}
			return fmt.Sprintf("reads:ip:%s", c.RealIP())
		},
		Endpoint:    "reads",
		SkipOnerror: true,
	})
}

func RefreshRateLimit(rdb *redis.Client) echo.MiddlewareFunc {
	return RateLimit(rdb, repository.RateLimiter{
		Max:      30,
		Duration: time.Minute,
		KeyFunc: func(c echo.Context) string {
			return fmt.Sprintf("refresh:%s", c.RealIP())
		},
		Endpoint:    "refresh",
		SkipOnerror: true,
	})
}

func UploadRateLimit(rdb *redis.Client) echo.MiddlewareFunc {
	return RateLimit(rdb, repository.RateLimiter{
		Max:      10,
		Duration: time.Minute,
		KeyFunc: func(c echo.Context) string {
			if userID := c.Get("user_id"); userID != nil {
				if uid, ok := userID.(uuid.UUID); ok {
					return fmt.Sprintf("upload:user:%s", uid.String())
				}
			}
			return fmt.Sprintf("upload:ip:%s", c.RealIP())
		},
		Endpoint:    "upload",
		SkipOnerror: true,
	})
}
