package main

import (
	"fmt"
	"time"
	"net/http"

	"go-test/internal/model"
	"go-test/internal/global"
	"go-test/internal/router"
	"go-test/pkg/config"
	"go-test/pkg/db"
	"go-test/pkg/logger"
	redisPkg "go-test/pkg/redis"

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

	if err := db.InitMySQL(); err != nil {
		panic(err)
	}

	if err := redisPkg.InitRedis(); err != nil {
		panic(err)
	}

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

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}

}
