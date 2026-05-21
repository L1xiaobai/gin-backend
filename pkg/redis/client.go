package redis

import (
	"context"
	"encoding/json"
	"time"

	"go-test/internal/global"
	appErrors "go-test/pkg/errors"
	"go-test/pkg/code"

	"github.com/redis/go-redis/v9"
)

// Set 设置字符串缓存
func Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    if key == "" {
        return appErrors.New(code.RedisError, "Redis key 不能为空")
    }

	if err := global.Redis.Set(ctx, key, value, ttl).Err(); err != nil {
		return appErrors.Wrap(code.RedisError, "Redis操作失败", err)
	}
	return nil
}

// Get 获取字符串缓存
func Get(ctx context.Context, key string) (string, error) {
	val, err := global.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", appErrors.Wrap(code.RedisError, "Redis操作失败", err)
	}
	return val, nil
}

// Del 删除缓存
func Del(ctx context.Context, key string) error {
	if err := global.Redis.Del(ctx, key).Err(); err != nil {
		return appErrors.Wrap(code.RedisError, "Redis操作失败", err)
	}
	return nil
}

// SetStruct 设置结构体缓存（自动序列化）
func SetStruct(ctx context.Context, key string, val interface{}, ttl time.Duration) error {
    if key == "" {
        return appErrors.New(code.RedisError, "Redis key 不能为空")
    }

	b, err := json.Marshal(val)
	if err != nil {
		return appErrors.New(code.RedisError, "序列化失败")
	}
	return Set(ctx, key, b, ttl)
}

// GetStruct 获取结构体缓存（自动反序列化）
func GetStruct(ctx context.Context, key string, out interface{}) error {
	b, err := global.Redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return appErrors.Wrap(code.RedisError, "Redis操作失败", err)
	}
	return json.Unmarshal(b, out)
}

// GetOrSet 获取或设置缓存
func GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	val, err := Get(ctx, key)
	if val != "" && err == nil {
		return val, nil
	}
	v, err := fn()
	if err != nil {
		return nil, err
	}
	if err := Set(ctx, key, v, ttl); err != nil {
		return nil, err
	}
	return v, nil
}