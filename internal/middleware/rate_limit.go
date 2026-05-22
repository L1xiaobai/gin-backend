package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go-test/internal/global"
	"go-test/pkg/response"
	"go-test/pkg/code"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !viper.GetBool("rate_limit.enabled") {
			c.Next()
			return
		}

		windowSeconds := viper.GetInt("rate_limit.window_seconds")
		maxRequests := viper.GetInt("rate_limit.max_requests")

		if windowSeconds <= 0 {
			windowSeconds = 60
		}
		if maxRequests <= 0 {
			maxRequests = 100
		}

		clientIP := c.ClientIP()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		key := fmt.Sprintf("rate_limit:ip:%s:path:%s", clientIP, path)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		var RateLimitScript = redis.NewScript(`
		local current = redis.call("INCR", KEYS[1])
		if current == 1 then
			redis.call("EXPIRE", KEYS[1], ARGV[1])
		end
		return current
		`)
		count, err := rateLimitScript.Run(
			ctx,
			global.Redis,
			[]string{key},
			windowSeconds,
		).Int64()

		ttl := global.Redis.TTL(ctx, key).Val()

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", maxRequests-int(count)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", ttl/time.Second))

		if count > int64(maxRequests) {
			c.Header("Retry-After", fmt.Sprintf("%d", ttl/time.Second))
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				response.FailData(code.RateLimited, "请求过于频繁，请稍后再试"),
			)
			return
		}

		c.Next()
	}
}