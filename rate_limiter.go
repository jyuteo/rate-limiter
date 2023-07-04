package ratelimiter

import "time"

// Request is an interface that defines the requests limted by RateLimter.
type Request interface {
}

// Frequency defines the rate of RateLimiter.
// RateLimiter wtih Frequency{Unit: 100, Interval: time.Second} allows 100 requests per second.
type Frequency struct {
	Unit     int64
	Interval time.Duration
}

// RateLimiter is used to rate-limit requests to some functions.
type RateLimiter interface {
	// AllowRequest() returns true when the request can be allowed or throttled otherwise.
	AllowRequest(r Request) bool
	// SetFrequency(f) sets RateLimiter to allow f.unit requests per f.interval.
	SetFrequency(f Frequency)
}
