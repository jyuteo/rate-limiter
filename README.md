# Rate limiter

Implements token bucket rate limiter and leaky bucket rate limiter

## Usage

### Token Bucket Rate Limiter

```go
import (
	ratelimiter "rate-limiter"
	"sync"
	"time"
)

func main() {
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
```

### Leaky Bucket Rate Limiter

```go
import (
	ratelimiter "rate-limiter"
	"sync"
	"time"
)

func main() {
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
```
