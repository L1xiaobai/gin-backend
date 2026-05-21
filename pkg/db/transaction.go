package db

import (
	"context"
	"time"

	"go-test/internal/global"
	"go-test/pkg/xcontext"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WithTransaction 执行事务并记录耗时日志
func WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	start := time.Now()
	requestID := xcontext.GetRequestID(ctx)

	err := global.DB.WithContext(ctx).Transaction(fn)

	elapsed := time.Since(start)
	if err != nil {
		global.Logger.Error("transaction failed",
			zap.String("request_id", requestID),
			zap.Duration("elapsed", elapsed),
			zap.Error(err),
		)
	} else {
		global.Logger.Info("transaction commit",
			zap.String("request_id", requestID),
			zap.Duration("elapsed", elapsed),
		)
	}

	return err
}

// WithTransactionAndCache 执行事务后删除 Redis 缓存
func WithTransactionAndCache(ctx context.Context, fn func(tx *gorm.DB) error, cacheKeys ...string) error {
	err := WithTransaction(ctx, fn)
	if err != nil {
		return err
	}

	for _, key := range cacheKeys {
		if key == "" {
			continue
		}
		if global.Redis != nil {
			if delErr := global.Redis.Del(ctx, key).Err(); delErr != nil && delErr != redis.Nil {
				global.Logger.Warn("redis delete failed after transaction",
					zap.String("request_id", xcontext.GetRequestID(ctx)),
					zap.String("key", key),
					zap.Error(delErr),
				)
			}
		}
	}
	return nil
}