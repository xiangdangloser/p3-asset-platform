package database

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // 无密码
		DB:       0,  // 默认DB
	})

	_, err := RDB.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Redis 连接失败: %v", err)
	}

	log.Println("Redis 初始化成功")
}
