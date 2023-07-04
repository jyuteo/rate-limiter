package ratelimiter

import (
	"sync/atomic"
	"time"
)

type TokenBucketRateLimiter struct {
	frequency    Frequency
	currentToken int64
	isStarted    bool
	stopChan     chan bool
}

type TokenBucketRateLimiterRequest struct{}

func NewTokenBucketRateLimiter(f Frequency) *TokenBucketRateLimiter {
	r := &TokenBucketRateLimiter{
		frequency: f,
		stopChan:  make(chan bool),
	}
	go r.start()
	return r
}

// SetFrequency() updates the frequency of rate limiter.
// It will restart the background process that refills bucket tokens with new frequency.
func (r *TokenBucketRateLimiter) SetFrequency(f Frequency) {
	r.frequency = f
	r.Stop()
	go r.start()
}

// AllowRequest() allows request when there are tokens available in the bucket.
func (r *TokenBucketRateLimiter) AllowRequest(req Request) bool {
	_, ok := req.(*TokenBucketRateLimiterRequest)
	if !ok {
		panic(`Invalid request type for TokenBucketRateLimiterRequest`)
	}
	currentToken := atomic.LoadInt64(&r.currentToken)
	if currentToken > 0 {
		atomic.AddInt64(&r.currentToken, int64(-1))
		return true
	}
	return false
}

// CurrentToken() returns current token in the bucket.
func (r *TokenBucketRateLimiter) CurrentToken() int64 {
	return atomic.LoadInt64(&r.currentToken)
}

// start() refills bucket tokens based on its frequency until t.Stop() is called.
// It will run in background with a goroutine.
func (r *TokenBucketRateLimiter) start() {
	if r.isStarted {
		return
	}
	r.isStarted = true
	ticker := time.NewTicker(r.frequency.Interval)
	for {
		select {
		case <-ticker.C:
			currentToken := atomic.LoadInt64(&r.currentToken)
			maxTokenToAdd := atomic.LoadInt64(&r.frequency.Unit)
			if currentToken > maxTokenToAdd {
				atomic.AddInt64(&r.currentToken, maxTokenToAdd)
			} else {
				atomic.AddInt64(&r.currentToken, maxTokenToAdd-currentToken)
			}
		case <-r.stopChan:
			ticker.Stop()
			return
		}
	}
}

// Stop() stops the background process that refills bucket tokens.
func (r *TokenBucketRateLimiter) Stop() {
	if !r.isStarted {
		return
	}
	r.isStarted = false
	r.stopChan <- true
}
