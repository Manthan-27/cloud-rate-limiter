package limiter

import (
	"time"

	"github.com/Manthan-27/cloud-rate-limiter/internal/storage"
)

type Fixed struct {
	Limit     int
	WindowSec int
}

func NewFixed(limit, windowSec int) *Fixed {
	return &Fixed{Limit: limit, WindowSec: windowSec}
}

func (f *Fixed) Allow(key string) (Decision, error) {
	ttl := time.Duration(f.WindowSec) * time.Second
	count, ttlLeft, err := storage.IncrWithTTL(key, ttl)
	if err != nil {
		return Decision{}, err
	}
	dec := Decision{
		Allowed:   count <= int64(f.Limit),
		Remaining: max(0, f.Limit-int(count)),
		Limit:     f.Limit,
		WindowSec: f.WindowSec,
	}
	if !dec.Allowed {
		dec.RetryAfterMs = int(ttlLeft.Milliseconds())
	}
	return dec, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
