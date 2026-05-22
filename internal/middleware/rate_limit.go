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
		key := fmt.Sprintf("rate_limit:ip:%s", clientIP)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		count, err := global.Redis.Incr(ctx, key).Result()
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError, 
				response.FailData(code.RedisError, "限流服务异常"),
			)
			return
		}

		if count == 1 {
			_ = global.Redis.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)
		}

		if count > int64(maxRequests) {
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests, 
				response.FailData(code.RateLimited, "请求过于频繁，请稍后再试"),
			)
			return
		}

		c.Next()
	}
}