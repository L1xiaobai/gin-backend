package db_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go-test/internal/global"
	"go-test/internal/model"
	"go-test/pkg/db"
	pkgRedis "go-test/pkg/redis"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/driver/mysql"
	"go.uber.org/zap"
)

// 初始化 Redis 和 DB，用于测试
func initTestEnv() {
	// 初始化 Redis
	global.Redis = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	if _, err := global.Redis.Ping(context.Background()).Result(); err != nil {
		panic("Redis 初始化失败: " + err.Error())
	}

	// 初始化数据库
	dsn := "gobiz:123456@tcp(127.0.0.1:3306)/go_business?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	global.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库初始化失败: " + err.Error())
	}
	global.DB.Where("username LIKE ?", "tx_user_%").Delete(&model.User{})
	if global.Logger == nil {
		logger, _ := zap.NewDevelopment() // 测试用
		global.Logger = logger
	}
}


func TestMain(m *testing.M) {
	initTestEnv()
	m.Run()
}

// Test WithTransaction: 正常提交
func TestWithTransaction_Success(t *testing.T) {
	initTestEnv()
	ctx := context.Background()

	err := db.WithTransaction(ctx, func(tx *gorm.DB) error {
		user := &model.User{Username: "tx_user1", Password: "123"}
		return tx.Create(user).Error
	})
	if err != nil {
		t.Fatalf("WithTransaction failed: %v", err)
	}

	// 验证数据是否写入
	var user model.User
	if err := global.DB.Where("username=?", "tx_user1").First(&user).Error; err != nil {
		t.Fatalf("Record not found after transaction commit: %v", err)
	}

	// 清理测试数据
	global.DB.Delete(&user)
}

// Test WithTransaction: 返回错误回滚
func TestWithTransaction_Rollback(t *testing.T) {
	initTestEnv()
	ctx := context.Background()

	txErr := db.WithTransaction(ctx, func(tx *gorm.DB) error {
		user := &model.User{Username: "tx_user2", Password: "123"}
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		// 返回错误强制回滚
		return fmt.Errorf("force rollback")
	})
	if txErr == nil {
		t.Fatalf("Expected transaction error")
	}

	// 验证数据未写入
	var user model.User
	if err := global.DB.Where("username=?", "tx_user2").First(&user).Error; err == nil {
		t.Fatalf("Record should not exist after rollback")
	}
}

// Test WithTransactionAndCache: 提交事务 + 删除缓存
func TestWithTransactionAndCache(t *testing.T) {
	initTestEnv()
	ctx := context.Background()
	cacheKey := "tx_cache_user"

	// 写入缓存
	pkgRedis.Set(ctx, cacheKey, "old_value", time.Minute*5)

	err := db.WithTransactionAndCache(ctx, func(tx *gorm.DB) error {
		user := &model.User{Username: "tx_user3", Password: "123"}
		return tx.Create(user).Error
	}, cacheKey)
	if err != nil {
		t.Fatalf("WithTransactionAndCache failed: %v", err)
	}

	// 验证缓存是否删除
	val, _ := pkgRedis.Get(ctx, cacheKey)
	if val != "" {
		t.Fatalf("Cache should be deleted after transaction commit")
	}

	// 清理测试数据
	global.DB.Where("username=?", "tx_user3").Delete(&model.User{})
}

// Test WithTransactionAndCache: 事务出错回滚 + 缓存不删除
func TestWithTransactionAndCache_Rollback(t *testing.T) {
	initTestEnv()
	ctx := context.Background()
	cacheKey := "tx_cache_user2"

	// 写入缓存
	pkgRedis.Set(ctx, cacheKey, "old_value", time.Minute*5)

	err := db.WithTransactionAndCache(ctx, func(tx *gorm.DB) error {
		user := &model.User{Username: "tx_user4", Password: "123"}
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		// 返回错误触发回滚
		return fmt.Errorf("force rollback")
	}, cacheKey)

	if err == nil {
		t.Fatalf("Expected transaction error")
	}

	// 缓存应该还存在
	val, _ := pkgRedis.Get(ctx, cacheKey)
	if val == "" {
		t.Fatalf("Cache should not be deleted if transaction failed")
	}

	// 清理测试数据
	global.DB.Where("username=?", "tx_user4").Delete(&model.User{})
	pkgRedis.Del(ctx, cacheKey)
}

// Test panic 捕获
func TestWithTransaction_Panic(t *testing.T) {
	initTestEnv()
	ctx := context.Background()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic to be propagated")
		}
	}()

	db.WithTransaction(ctx, func(tx *gorm.DB) error {
		panic("simulate panic")
	})
}