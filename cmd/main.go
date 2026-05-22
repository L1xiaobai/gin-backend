package main

import (
	"fmt"
	"time"
	"os"
	"os/signal"
	"context"
	"net/http"
	"log"

	"go-test/internal/model"
	"go-test/internal/global"
	"go-test/internal/router"
	"go-test/pkg/config"
	"go-test/pkg/db"
	"go-test/pkg/logger"
	redisPkg "go-test/pkg/redis"
	"go-test/pkg/metrics"

	"github.com/spf13/viper"
)

func main() {
	if err := config.InitConfig(); err != nil {
		panic(err)
	}
	if err := logger.InitLogger(); err != nil {
		panic(err)
	}
	defer global.Logger.Sync()

	metrics.InitMetrics()

	if err := db.InitMySQL(); err != nil {
		panic(err)
	}

	if err := redisPkg.InitRedis(); err != nil {
		panic(err)
	}
	
	metrics.CollectDependencyMetrics()

	if err := global.DB.AutoMigrate(&model.User{}); err != nil {
		panic(err)
	}

	r := router.InitRouter()

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", viper.GetInt("server.port")),
		Handler:        r,
		ReadTimeout:    time.Duration(viper.GetInt("server.read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(viper.GetInt("server.write_timeout")) * time.Second,
		IdleTimeout:    time.Duration(viper.GetInt("server.idle_timeout")) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed{
			panic(err)
		}
	}()
	log.Println("Server started on :8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt) // 捕获 Ctrl+C
	<-quit
	log.Println("Shutting down server...")

	// 创建超时 Context，用于优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭 HTTP 服务
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// 关闭 Redis
	if global.Redis != nil {
		if err := global.Redis.Close(); err != nil {
			log.Println("Redis close error:", err)
		}
	}

	// 关闭数据库连接（如果 gorm.DB 需要关闭底层连接池）
	sqlDB, err := global.DB.DB()
	if err == nil {
		sqlDB.Close()
	}

	log.Println("Server exited gracefully")
}
