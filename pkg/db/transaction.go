package db

import (
    "context"
    "go-test/internal/global"

    "gorm.io/gorm"
)

// WithTransaction 封装事务执行，支持 ctx
func WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
    return global.DB.WithContext(ctx).Transaction(fn)
}