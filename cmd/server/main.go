package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Manthan-27/cloud-rate-limiter/internal/api"
	"github.com/Manthan-27/cloud-rate-limiter/internal/config"
	"github.com/Manthan-27/cloud-rate-limiter/internal/middleware"
	"github.com/Manthan-27/cloud-rate-limiter/internal/storage"
	"github.com/Manthan-27/cloud-rate-limiter/pkg/logger"
)

// Prometheus metrics
var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"endpoint", "method"},
	)

	rateLimitHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_hits_total",
			Help: "Total number of rate-limited requests",
		},
		[]string{"endpoint"},
	)
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(rateLimitHits)
}

// MetricsMiddleware increments Prometheus counters for each request
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestsTotal.WithLabelValues(r.URL.Path, r.Method).Inc()
		next.ServeHTTP(w, r)
	})
}

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
	router.Use(MetricsMiddleware)           // Prometheus metrics middleware
	router.Use(middleware.RateLimiter(cfg)) // Your existing rate limiter middleware

	// Routes
	router.HandleFunc("/health", api.HealthHandler).Methods("GET")
	router.Handle("/metrics", promhttp.Handler()) // Prometheus metrics endpoint

	// Start server
	port := fmt.Sprintf(":%s", cfg.ServerPort)
	fmt.Printf("🚀 Server running on %s\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}
