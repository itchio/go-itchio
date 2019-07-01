package itchio

import "golang.org/x/time/rate"

var defaultRateLimiter *rate.Limiter

// DefaultRateLimiter returns a rate.Limiter suitable
// for consuming the itch.io API. It is shared across all
// instances of Client, unless a custom limiter is set.
func DefaultRateLimiter() *rate.Limiter {
	if defaultRateLimiter == nil {
		limit := rate.Limit(8.0)
		burst := 20
		defaultRateLimiter = rate.NewLimiter(limit, burst)
	}
	return defaultRateLimiter
}
