package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var ctx = context.Background()

func InitRedis(host, port string) {
	addr := fmt.Sprintf("%s:%s", host, port)
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password for dev
		DB:       0,
	})

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")
}

// increment counter with expiry
func Increment(key string, duration time.Duration) (int, error) {
	val, err := RedisClient.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if val == 1 {
		// first time, set expiry
		RedisClient.Expire(ctx, key, duration)
	}

	return int(val), nil
}
