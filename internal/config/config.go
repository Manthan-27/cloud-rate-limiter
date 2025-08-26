package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort   string
	RedisHost    string
	RedisPort    string
	RateLimit    int
	RateDuration int // seconds

	// NEW
	Strategy     string // fixed | sliding | token_bucket
	Burst        int    // token bucket only
	Refillpersec int    // token bucket only
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	rateLimit := atoi(getEnv("RATE_LIMIT", "10"))
	rateDuration := atoi(getEnv("RATE_DURATION", "60"))
	burst := atoi(getEnv("BURST", "10"))
	refillPerSec := atoi(getEnv("REFILL_PER_SEC", "5"))

	return &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		RedisHost:    getEnv("REDIS_HOST", "localhost"),
		RedisPort:    getEnv("REDIS_PORT", "6379"),
		RateLimit:    rateLimit,
		RateDuration: rateDuration,

		Strategy:     getEnv("RATE_LIMIT_STRATEGY", "fixed"),
		Burst:        burst,
		Refillpersec: refillPerSec,
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func atoi(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("invalid int for config: %s", s)
	}
	return v
}
