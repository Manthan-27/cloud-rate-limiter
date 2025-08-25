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
	RateDuration int // in seconds

	Strategy     string
	Burst        int
	Refillpersec int
}

func LoadConfig() *Config {
	// load .env file if exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	rateLimit, err := strconv.Atoi(getEnv("RATE_LIMIT", "5"))
	if err != nil {
		log.Fatalf("Invalid RATE_LIMIT: %v", err)
	}
	rateDuration, err := strconv.Atoi(getEnv("RATE_DURATION", "60"))
	if err != nil {
		log.Fatalf("Invalid RATE_DURATION: %v", err)
	}

	return &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		RedisHost:    getEnv("REDIS_HOST", "localhost"),
		RedisPort:    getEnv("REDIS_PORT", "6379"),
		RateLimit:    rateLimit,
		RateDuration: rateDuration,
	}
}

// helper to read env variable or fallback default
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}
