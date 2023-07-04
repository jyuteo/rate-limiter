package ratelimiter_test

import (
	"fmt"
	ratelimiter "rate-limiter"
	"sync"
	"sync/atomic"
	"time"
)

func LeakyBucketRateLimiter() {
	// initialize rate limiter that can hold at most 2000 requests in queue and allow 1000 requests per second
	f := ratelimiter.Frequency{
		Unit:     int64(1),
		Interval: time.Millisecond,
	}
	rl := ratelimiter.NewLeakyBucketRateLimiter(f, 2000)
	defer rl.Stop()

	start := time.Now()
	allow := int64(0)
	rejected := int64(0)
	var wg sync.WaitGroup
	for i := 0; i < 300000; i++ {
		wg.Add(1)
		req := ratelimiter.LeakyBucketRateLimiterRequest{Id: int64(i)}
		go checkReq(rl, &req, &allow, &rejected, &wg, i)

	}
	wg.Wait()
	testResult(start, allow, rejected)

	// Output:
	// Run time: 2.04s
	// Allowed requests: 2040
	// Rejected requests: 97960
	// RPS: 999.90
}

func TokenBucketRateLimiter() {
	// initialize rate limiter that allows 1000 requests per second
	f := ratelimiter.Frequency{
		Unit:     int64(1),
		Interval: time.Millisecond,
	}
	rl := ratelimiter.NewTokenBucketRateLimiter(f)
	defer rl.Stop()

	start := time.Now()
	allow := int64(0)
	rejected := int64(0)
	var wg sync.WaitGroup
	for i := 0; i < 300000; i++ {
		wg.Add(1)
		req := ratelimiter.TokenBucketRateLimiterRequest{}
		go checkReq(rl, &req, &allow, &rejected, &wg, i)
	}
	wg.Wait()
	testResult(start, allow, rejected)

	// Output:
	// Run time: 0.0922s
	// Allowed requests: 92
	// Rejected requests: 299908
	// RPS: 997.76
}

func checkReq(l ratelimiter.RateLimiter, req ratelimiter.Request, allow *int64, rejected *int64, wg *sync.WaitGroup, i int) {
	if l.AllowRequest(req) {
		atomic.AddInt64(allow, 1)
	} else {
		atomic.AddInt64(rejected, 1)
	}
	wg.Done()
}

func testResult(start time.Time, allowedCount int64, rejectedCount int64) {
	duration := time.Since(start).Seconds()
	fmt.Printf("Run time: %.4fs\n", duration)
	fmt.Printf("Allowed requests: %d\n", allowedCount)
	fmt.Printf("Rejected requests: %d\n", rejectedCount)
	fmt.Printf("RPS: %.2f\n", float64(allowedCount)/duration)
}
