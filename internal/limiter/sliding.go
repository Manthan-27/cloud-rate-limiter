package limiter

import (
	"strconv"
	"time"

	"github.com/Manthan-27/cloud-rate-limiter/internal/storage"
)

type Sliding struct {
	Limit     int
	WindowSec int
}

func NewSliding(limit, windowSec int) *Sliding {
	return &Sliding{Limit: limit, WindowSec: windowSec}
}

func (s *Sliding) Allow(key string) (Decision, error) {
	window := time.Duration(s.WindowSec) * time.Second
	nowMs := time.Now().UnixMilli()
	cutoff := nowMs - int64(window.Milliseconds())

	// 1) purge old
	if err := storage.ZRemRangeByScore(key, "-inf", strconv.FormatInt(cutoff, 10)); err != nil {
		return Decision{}, err
	}
	// 2) add current
	if err := storage.ZAddNow(key, nowMs); err != nil {
		return Decision{}, err
	}
	// 3) count
	n, err := storage.ZCard(key)
	if err != nil {
		return Decision{}, err
	}
	// 4) set expire (helps cleanup)
	_ = storage.Expire(key, window)

	allowed := n <= int64(s.Limit)
	dec := Decision{
		Allowed:   allowed,
		Remaining: max(0, s.Limit-int(n)),
		Limit:     s.Limit,
		WindowSec: s.WindowSec,
	}
	if !allowed {
		// find retry-after: when will earliest event roll out of window?
		earliest, err := storage.ZRangeFirstScore(key)
		if err == nil && earliest > 0 {
			msLeft := (earliest + int64(window.Milliseconds())) - nowMs
			if msLeft > 0 {
				dec.RetryAfterMs = int(msLeft)
			}
		}
	}
	return dec, nil
}
