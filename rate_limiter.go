package itchio

import "golang.org/x/time/rate"

var defaultRateLimiter *rate.Limiter

// DefaultRateLimiter returns a rate.Limiter suitable
// for consuming the itch.io API. It is shared across all
// instances of Client, unless a custom limiter is set.
func DefaultRateLimiter() *rate.Limiter {
	// Although - at the time of this writing - the server-side settings allow for
	// 8 reqs/s and a burst of 20, in practice, anything above a 15 burst fails
	// with more than 2 concurrent workers (which is not unusual).
	// If you know more than me about this (ie. how Nginx actually enforces
	// rate limits), have a thread: https://twitter.com/fasterthanlime/status/1150751037352865797
	if defaultRateLimiter == nil {
		limit := rate.Limit(8.0)
		burst := 15
		defaultRateLimiter = rate.NewLimiter(limit, burst)
	}
	return defaultRateLimiter
}
