package redis

import (
	"context"
	"fmt"
	"time"

	"go-test/internal/global"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() error {
	addr := fmt.Sprintf("%s:%d",
		viper.GetString("redis.host"),
		viper.GetInt("redis.port"),
	)

	rdb := redis.NewClient(&redis.Options{
		Addr:     	  addr,
		Password: 	  viper.GetString("redis.password"),
		DB:       	  viper.GetInt("redis.db"),
		PoolSize:     viper.GetInt("redis.pool_size"),
		MinIdleConns: viper.GetInt("redis.min_idle_conns"),
		DialTimeout:  time.Duration(viper.GetInt("redis.dial_timeout")) * time.Second,
		ReadTimeout:  time.Duration(viper.GetInt("redis.read_timeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("redis.write_timeout")) * time.Second,
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)

	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil{
		return err
	}

	global.Redis = rdb
	
	return nil
}
