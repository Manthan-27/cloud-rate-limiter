package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Manthan-27/cloud-rate-limiter/internal/api"
	"github.com/Manthan-27/cloud-rate-limiter/internal/config"
	"github.com/Manthan-27/cloud-rate-limiter/internal/middleware"
	"github.com/Manthan-27/cloud-rate-limiter/internal/storage"
	"github.com/Manthan-27/cloud-rate-limiter/pkg/logger"
)

func main() {
	// Init logger
	logger.Init()
	logger.InfoLogger.Println("Starting Cloud Rate Limiter...")

	// Load config
	cfg := config.LoadConfig()

	// Init Redis
	storage.InitRedis(cfg.RedisHost, cfg.RedisPort)

	// Router
	router := mux.NewRouter()

	// Middleware
	router.Use(middleware.RateLimiter(cfg))

	// Routes
	router.HandleFunc("/health", api.HealthHandler).Methods("GET")

	// Start server
	port := fmt.Sprintf(":%s", cfg.ServerPort)
	fmt.Printf("🚀 Server running on %s\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}
