package middleware

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Manthan-27/cloud-rate-limiter/internal/config"
	"github.com/Manthan-27/cloud-rate-limiter/internal/storage"
)

func RateLimiter(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "cannot parse IP", http.StatusInternalServerError)
				return
			}

			key := fmt.Sprintf("rl:%s", ip)
			count, err := storage.Increment(key, time.Duration(cfg.RateDuration)*time.Second)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if count > cfg.RateLimit {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
