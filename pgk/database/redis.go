package database

import (
	"RAG/config"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis(cfg *config.Config) {
	addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,

		//	Config for scale system
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status := RedisClient.Ping(ctx)

	if err := status.Err(); err != nil {
		panic("Can't connect to redis" + err.Error())
	}
	fmt.Println("Connected to redis")
}
