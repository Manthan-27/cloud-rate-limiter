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

func IncrWithTTL(key string, ttl time.Duration) (int64, time.Duration, error) {
	pipe := RedisClient.TxPipeline()
	incr := pipe.Incr(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, 0, err
	}
	count, _ := incr.Result()
	// if first time, set expire
	if count == 1 {
		RedisClient.Expire(ctx, key, ttl)
		ttlCmd = RedisClient.TTL(ctx, key)
	}
	left, _ := ttlCmd.Result()
	return count, left, nil
}

// Sliding window helpers (ZSET)
func ZRemRangeByScore(key string, min string, max string) error {
	return RedisClient.ZRemRangeByScore(ctx, key, min, max).Err()
}

func ZAddNow(key string, scoreMs int64) error {
	m := redis.Z{Score: float64(scoreMs), Member: scoreMs}
	return RedisClient.ZAdd(ctx, key, m).Err()
}

func ZCard(key string) (int64, error) {
	return RedisClient.ZCard(ctx, key).Result()
}

func ZRangeFirstScore(key string) (int64, error) {
	// get the lowest score (earliest event)
	xs, err := RedisClient.ZRangeWithScores(ctx, key, 0, 0).Result()
	if err != nil || len(xs) == 0 {
		return 0, err
	}
	return int64(xs[0].Score), nil
}

func Expire(key string, ttl time.Duration) error {
	return RedisClient.Expire(ctx, key, ttl).Err()
}

// --- Token bucket (Lua) ---
type TBResult struct {
	Allowed      bool
	Remaining    int
	RetryAfterMs int
}

var tokenBucketLua = redis.NewScript(`
-- KEYS[1] = key
-- ARGV[1] = capacity
-- ARGV[2] = refill_per_sec
-- ARGV[3] = now_ms
-- state: HSET key tokens, last_ms ; use PX TTL to clean up eventually
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local d = redis.call('HMGET', key, 'tokens', 'last_ms')
local tokens = tonumber(d[1])
local last = tonumber(d[2])

if tokens == nil then
  tokens = capacity
  last = now
else
  if now > last then
    local elapsed = (now - last) / 1000.0
    local add = math.floor(elapsed * refill)
    tokens = math.min(capacity, tokens + add)
    if add > 0 then
      last = now
    end
  end
end

local allowed = 0
local retry_ms = 0

if tokens > 0 then
  tokens = tokens - 1
  allowed = 1
else
  -- compute time until next token
  local missing = 1 - tokens -- tokens is 0 or negative (should be 0)
  retry_ms = math.floor((missing / refill) * 1000)
end

redis.call('HMSET', key, 'tokens', tokens, 'last_ms', last)
-- set a TTL to auto-clean stale buckets (e.g., 2 minutes)
redis.call('PEXPIRE', key, 120000)

return {allowed, tokens, retry_ms}
`)

func TokenBucketAllow(key string, capacity, refillPerSec int) (TBResult, error) {
	now := time.Now().UnixMilli()
	res, err := tokenBucketLua.Run(ctx, RedisClient, []string{key}, capacity, refillPerSec, now).Result()
	if err != nil {
		return TBResult{}, err
	}
	arr := res.([]interface{})
	allowed := arr[0].(int64) == 1
	remaining := int(arr[1].(int64))
	retryMs := int(arr[2].(int64))
	return TBResult{Allowed: allowed, Remaining: remaining, RetryAfterMs: retryMs}, nil
}
