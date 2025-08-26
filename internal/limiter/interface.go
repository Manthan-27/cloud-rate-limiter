package limiter

type Decision struct {
	Allowed      bool
	Remaining    int
	RetryAfterMs int // for blocked responses; 0 if allowed
	Limit        int
	WindowSec    int
}

type Limiter interface {
	// key is caller identity: ip/apiKey etc.
	Allow(key string) (Decision, error)
}
