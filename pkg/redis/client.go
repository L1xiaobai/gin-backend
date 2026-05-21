package redis

import (
    "context"
    "time"

    "go-test/internal/global"
)

// Set 带超时写入
func Set(ctx context.Context, key string, value any, ttl time.Duration) error {
    return global.Redis.Set(ctx, key, value, ttl).Err()
}

// Get 获取
func Get(ctx context.Context, key string) (string, error) {
    return global.Redis.Get(ctx, key).Result()
}

// Delete 删除
func Delete(ctx context.Context, key string) error {
    return global.Redis.Del(ctx, key).Err()
}