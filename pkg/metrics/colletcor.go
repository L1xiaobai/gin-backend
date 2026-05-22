package metrics

import (
	"context"
	"time"

	"go-test/internal/global"
)

func CollectDependencyMetrics() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			collectMySQLMetrics()
			collectRedisMetrics()
			<-ticker.C
		}
	}()
}

func collectMySQLMetrics() {
	if global.DB == nil {
		MySQLUp.Set(0)
		return
	}

	sqlDB, err := global.DB.DB()
	if err != nil {
		MySQLUp.Set(0)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		MySQLUp.Set(0)
	} else {
		MySQLUp.Set(1)
	}

	stats := sqlDB.Stats()
	MySQLOpenConnections.Set(float64(stats.OpenConnections))
	MySQLInUseConnections.Set(float64(stats.InUse))
	MySQLIdleConnections.Set(float64(stats.Idle))
}

func collectRedisMetrics() {
	if global.Redis == nil {
		RedisUp.Set(0)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := global.Redis.Ping(ctx).Err(); err != nil {
		RedisUp.Set(0)
		return
	}

	RedisUp.Set(1)
}