package config

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis() error {

	redisURL := os.Getenv("REDIS_URL")

	if redisURL == "" {
		return fmt.Errorf("REDIS_URL is not set")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("invalid REDIS_URL: %v", err)
	}

	RedisClient = redis.NewClient(opt)

	ctx := context.Background()

	_, err = RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis connection failed: %v", err)
	}

	fmt.Println("Redis connected successfully!")

	return nil
}