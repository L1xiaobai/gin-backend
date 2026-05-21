package redis_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	pkgRedis "go-test/pkg/redis"
	"go-test/internal/global"
	appErrors "go-test/pkg/errors"
	"go-test/pkg/code"

	"github.com/redis/go-redis/v9"
)

func initRedis() {
    global.Redis = redis.NewClient(&redis.Options{
        Addr:     "127.0.0.1:6379", // 本地 Redis 地址
        Password: "",               // 没有密码
        DB:       0,
    })
}

func TestMain(m *testing.M) {
    initRedis()
    m.Run()
}	

// 测试基础 Set/Get/Del
func TestRedisSetGetDel(t *testing.T) {
	ctx := context.Background()
	key := "test:key"

	// Set
	if err := pkgRedis.Set(ctx, key, "value123", 5*time.Minute); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get
	val, err := pkgRedis.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value123" {
		t.Fatalf("Expected value123, got %s", val)
	}

	// Del
	if err := pkgRedis.Del(ctx, key); err != nil {
		t.Fatalf("Del failed: %v", err)
	}

	// Get after Del
	val, err = pkgRedis.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get after Del failed: %v", err)
	}
	if val != "" {
		t.Fatalf("Expected empty string after Del, got %s", val)
	}
}

// 测试结构体缓存
func TestRedisSetGetStruct(t *testing.T) {
	ctx := context.Background()
	type User struct {
		ID       uint
		Username string
	}

	user := User{ID: 1, Username: "admin"}

	// SetStruct
	if err := pkgRedis.SetStruct(ctx, "user:1", &user, 5*time.Minute); err != nil {
		t.Fatalf("SetStruct failed: %v", err)
	}

	// GetStruct
	var cached User
	if err := pkgRedis.GetStruct(ctx, "user:1", &cached); err != nil {
		t.Fatalf("GetStruct failed: %v", err)
	}
	if cached.Username != "admin" {
		t.Fatalf("Expected admin, got %s", cached.Username)
	}

	// GetStruct cache miss
	var empty User
	if err := pkgRedis.GetStruct(ctx, "user:2", &empty); err != nil {
		t.Fatalf("GetStruct miss failed: %v", err)
	}
	if empty.ID != 0 {
		t.Fatalf("Expected ID 0 for miss, got %d", empty.ID)
	}
}

// 测试 GetOrSet 功能
func TestRedisGetOrSet(t *testing.T) {
	ctx := context.Background()
	key := "cache:or:set"

	val, err := pkgRedis.GetOrSet(ctx, key, 5*time.Minute, func() (interface{}, error) {
		return "generated-value", nil
	})
	if err != nil {
		t.Fatalf("GetOrSet failed: %v", err)
	}
	if val != "generated-value" {
		t.Fatalf("Expected generated-value, got %v", val)
	}

	// 再次命中缓存
	val, err = pkgRedis.GetOrSet(ctx, key, 5*time.Minute, func() (interface{}, error) {
		return "new-value", nil
	})
	if err != nil {
		t.Fatalf("GetOrSet second call failed: %v", err)
	}
	if val != "generated-value" {
		t.Fatalf("Expected cache hit generated-value, got %v", val)
	}
}

// 测试 TTL 过期
func TestRedisTTL(t *testing.T) {
	ctx := context.Background()
	key := "ttl:key"

	if err := pkgRedis.Set(ctx, key, "val", 2*time.Second); err != nil {
		t.Fatalf("Set with TTL failed: %v", err)
	}

	time.Sleep(3 * time.Second)

	val, err := pkgRedis.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get after TTL failed: %v", err)
	}
	if val != "" {
		t.Fatalf("Expected empty string after TTL, got %s", val)
	}
}

// 测试 Redis 错误封装
func TestRedisError(t *testing.T) {
	ctx := context.Background()

	// 模拟错误：假设连接关闭或 key 类型错误
	err := pkgRedis.Set(ctx, "", "val", time.Minute) // 空 key 触发错误
	if err == nil {
		t.Fatalf("Expected error for empty key")
	}

	if appErr, ok := err.(*appErrors.AppError); ok {
		if appErr.Code != code.RedisError {
			t.Fatalf("Expected RedisError code, got %d", appErr.Code)
		} else {
			fmt.Println("RedisError correctly wrapped:", appErr)
		}
	} else {
		t.Fatalf("Error is not AppError")
	}
}