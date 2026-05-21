package api

import (
	"context"

	"go-test/pkg/response"
	"go-test/internal/global"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck 健康检查接口
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx := context.Background()

	// 检查数据库
	sqlDB, err := global.DB.DB()
	if err != nil {
		response.Fail(c, 50001, "数据库不可用")
		return
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		response.Fail(c, 50001, "数据库不可用")
		return
	}

	// 检查 Redis
	_, redisErr := global.Redis.Ping(ctx).Result()
	if redisErr != nil {
		response.Fail(c, 50002, "Redis 不可用")
		return
	}

	// 所有服务正常
	response.Success(c, gin.H{"status": "ok"})
}