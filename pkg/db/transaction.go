package db

import (
	"context"
	"time"

	"go-test/internal/global"
	"go-test/pkg/xcontext"
	appErrors "go-test/pkg/errors"
	"go-test/pkg/code"
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
    tx := global.DB.Begin()
    if tx.Error != nil {
        return appErrors.Wrap(code.DatabaseError, "事务开始失败", tx.Error)
    }

    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            global.Logger.Error("事务 panic", zap.Any("recover", r))
            panic(r)
        }
    }()

    if err := fn(tx); err != nil {
        tx.Rollback()
        return err
    }

    if err := tx.Commit().Error; err != nil {
        return appErrors.Wrap(code.DatabaseError, "事务提交失败", err)
    }

    for _, key := range cacheKeys {
        if key != "" {
            _ = global.Redis.Del(ctx, key)
        }
    }

    return nil
}