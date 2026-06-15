package config

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		DB:   0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Redis connected")

	return client
}
