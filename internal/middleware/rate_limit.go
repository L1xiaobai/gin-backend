package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-test/internal/global"
	"go-test/pkg/code"
	"go-test/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]

local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local ttl = tonumber(ARGV[5])

local bucket = redis.call("HMGET", key, "tokens", "timestamp")
local tokens = tonumber(bucket[1])
local timestamp = tonumber(bucket[2])

if tokens == nil then
    tokens = capacity
    timestamp = now
end

local elapsed = now - timestamp
if elapsed < 0 then
    elapsed = 0
end

local refill = elapsed * refill_rate
tokens = math.min(capacity, tokens + refill)

local allowed = 0
local remaining = tokens

if tokens >= requested then
    allowed = 1
    tokens = tokens - requested
    remaining = tokens
end

redis.call("HMSET", key, "tokens", tokens, "timestamp", now)
redis.call("EXPIRE", key, ttl)

return {allowed, remaining}
`)

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !viper.GetBool("rate_limit.enabled") {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}		

		if isSkipPath(path) {
			c.Next()
			return
		}
		capacity, rate := getRateLimitRule(path)

		key := fmt.Sprintf("rate_limit:token_bucket:ip:%s:path:%s", clientIP, path)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		now := time.Now().UnixMilli()
		requested := 1
		refillRatePerMillisecond := rate / 1000.0

		// key 的过期时间可以设置为“桶完全恢复所需时间的 2 倍”
		ttlSeconds := int(float64(capacity)/rate*2) + 1
		if ttlSeconds < 60 {
			ttlSeconds = 60
		}

		result, err := tokenBucketScript.Run(
			ctx,
			global.Redis,
			[]string{key},
			capacity,
			refillRatePerMillisecond,
			now,
			requested,
			ttlSeconds,
		).Result()

		if err != nil {
			if viper.GetBool("rate_limit.fail_open") {
				global.Logger.Error("rate limit redis error", zap.Error(err))
				c.Next()
				return
			}
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				response.FailData(code.RedisError, "限流服务异常"),
			)
			return
		}

		values, ok := result.([]interface{})
		if !ok || len(values) < 2 {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				response.FailData(code.InternalError, "限流结果解析失败"),
			)
			return
		}

		allowed, err := strconv.ParseInt(fmt.Sprint(values[0]), 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				response.FailData(code.InternalError, "限流结果解析失败"),
			)
			return
		}

		remainingFloat, err := strconv.ParseFloat(fmt.Sprint(values[1]), 64)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				response.FailData(code.InternalError, "限流结果解析失败"),
			)
			return
		}

		remaining := int(remainingFloat)
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", capacity))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if allowed == 0 {
			missingTokens := float64(requested) - remainingFloat
			retryAfter := int(missingTokens/rate) + 1
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				response.FailData(code.RateLimited, "请求过于频繁，请稍后再试"),
			)
			return
		}

		c.Next()
	}
}


func getRateLimitRule(path string) (int, float64) {
	capacity := viper.GetInt("rate_limit.rules." + path + ".capacity")
	rate := viper.GetFloat64("rate_limit.rules." + path + ".rate")

	if capacity <= 0 {
		capacity = viper.GetInt("rate_limit.default.capacity")
	}
	if rate <= 0 {
		rate = viper.GetFloat64("rate_limit.default.rate")
	}

	if capacity <= 0 {
		capacity = 10
	}
	if rate <= 0 {
		rate = 1
	}

	return capacity, rate
}

func isSkipPath(path string) bool {
	for _, p := range viper.GetStringSlice("rate_limit.skip_paths") {
		if p == path {
			return true
		}
	}
	return false
}


