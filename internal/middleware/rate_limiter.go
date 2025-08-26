package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/Manthan-27/cloud-rate-limiter/internal/config"
	"github.com/Manthan-27/cloud-rate-limiter/internal/limiter"
)

func buildLimiter(cfg *config.Config) limiter.Limiter {
	switch cfg.Strategy {
	case "sliding":
		log.Printf("[RateLimiter] Using Sliding Window strategy: limit=%d, window=%ds", cfg.RateLimit, cfg.RateDuration)
		return limiter.NewSliding(cfg.RateLimit, cfg.RateDuration)

	case "token_bucket":
		log.Printf("[RateLimiter] Using Token Bucket strategy: burst=%d, refill=%d/sec", cfg.Burst, cfg.Refillpersec)
		return limiter.NewTokenBucket(cfg.Burst, cfg.Refillpersec)

	default: // "fixed"
		log.Printf("[RateLimiter] Using Fixed Window strategy: limit=%d, window=%ds", cfg.RateLimit, cfg.RateDuration)
		return limiter.NewFixed(cfg.RateLimit, cfg.RateDuration)
	}
}

func clientIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		return xf
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func RateLimiter(cfg *config.Config) func(http.Handler) http.Handler {
	l := buildLimiter(cfg)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			key := fmt.Sprintf("rl:%s", ip)

			dec, err := l.Allow(key)
			if err != nil {
				log.Printf("[RateLimiter] ERROR checking limit for client=%s: %v", ip, err)
				http.Error(w, "Internal Error", http.StatusInternalServerError)
				return
			}

			// log the decision
			if dec.Allowed {
				log.Printf("[RateLimiter] ✅ ALLOWED client=%s remaining=%d/%d", ip, dec.Remaining, dec.Limit)
			} else {
				log.Printf("[RateLimiter] ❌ BLOCKED client=%s (limit=%d, retry_after=%dms)", ip, dec.Limit, dec.RetryAfterMs)
			}

			// helpful headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprint(dec.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprint(dec.Remaining))
			if !dec.Allowed && dec.RetryAfterMs > 0 {
				w.Header().Set("Retry-After", fmt.Sprint((dec.RetryAfterMs+999)/1000))
			}

			if !dec.Allowed {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
