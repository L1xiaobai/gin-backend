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
	appConfig "go-test/pkg/config"
	"go-test/pkg/xcontext"
	"go-test/pkg/metrics"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]

local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
if refill_rate <= 0 then
    refill_rate = 0.001
end
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
local retry_after_ms = 0

if tokens >= requested then
    allowed = 1
    tokens = tokens - requested
    remaining = tokens
else
    local missing = requested - tokens
    retry_after_ms = math.ceil(missing / refill_rate)
end

redis.call("HMSET", key, "tokens", tokens, "timestamp", now)
redis.call("EXPIRE", key, ttl)

return {allowed, remaining, retry_after_ms}
`)

func RateLimit(cfg appConfig.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		if cfg.IsSkipPath(path) {
			c.Next()
			return
		}

		rule := cfg.GetRule(path)
		capacity := rule.Capacity
		rate := rule.Rate

		key := fmt.Sprintf("rate_limit:token_bucket:ip:%s:path:%s", clientIP, path)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		now := time.Now().UnixMilli()
		requested := 1
		refillRatePerMillisecond := rate / 1000.0

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
			metrics.RateLimitRedisErrorTotal.WithLabelValues(path).Inc()

			if cfg.FailOpen {
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
		if !ok || len(values) < 3 {
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

		retryAfterMS, err := strconv.ParseInt(fmt.Sprint(values[2]), 10, 64)
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
			metrics.RateLimitBlockedTotal.WithLabelValues(path).Inc()
			retryAfter := int((retryAfterMS + 999) / 1000)
			if retryAfter < 1 {
				retryAfter = 1
			}

			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))

			global.Logger.Warn("api rate limit exceeded",
				zap.String("request_id", xcontext.GetRequestID(c.Request.Context())),
				zap.String("client_ip", clientIP),
				zap.String("path", path),
				zap.String("method", c.Request.Method),
				zap.Int("capacity", capacity),
				zap.Float64("rate", rate),
				zap.Float64("remaining_tokens", remainingFloat),
				zap.Int("requested_tokens", requested),
				zap.Int64("retry_after_ms", retryAfterMS), 
				zap.Int("retry_after_seconds", retryAfter),
				zap.String("user_agent", c.Request.UserAgent()),
			)

			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				response.FailData(code.RateLimited, "请求过于频繁，请稍后再试"),
			)
			return
		}
		
		metrics.RateLimitAllowedTotal.WithLabelValues(path).Inc()
		c.Next()
	}
}