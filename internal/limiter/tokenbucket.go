package limiter

import (
	"github.com/Manthan-27/cloud-rate-limiter/internal/storage"
)

type TokenBucket struct {
	Bucket    int // capacity (burst)
	RefillPer int // tokens per second
}

func NewTokenBucket(burst, refillPerSec int) *TokenBucket {
	return &TokenBucket{Bucket: burst, RefillPer: refillPerSec}
}

func (t *TokenBucket) Allow(key string) (Decision, error) {
	// Use a Lua script that:
	// - reads tokens + last_refill
	// - refills based on elapsed time
	// - if tokens>0: decrement and allow
	// - else: deny and compute retry-after
	res, err := storage.TokenBucketAllow(key, t.Bucket, t.RefillPer)
	if err != nil {
		return Decision{}, err
	}
	// res.Allowed, res.Remaining, res.RetryAfterMs
	dec := Decision{
		Allowed:      res.Allowed,
		Remaining:    res.Remaining,
		RetryAfterMs: res.RetryAfterMs,
		Limit:        t.Bucket,
		WindowSec:    1, // semantic only; bucket is continuous
	}
	return dec, nil
}
