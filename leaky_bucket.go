package ratelimiter

import (
	"time"

	queue "github.com/enriquebris/goconcurrentqueue"
)

type LeakyBucketRateLimiter struct {
	q                  *queue.FixedFIFO
	frequency          Frequency
	capacity           int64
	isStarted          bool
	allowedRequestChan chan *LeakyBucketRateLimiterRequest
	stopChan           chan bool
}

type LeakyBucketRateLimiterRequest struct {
	Id int64
}

func NewLeakyBucketRateLimiter(f Frequency, capacity int64) *LeakyBucketRateLimiter {
	r := &LeakyBucketRateLimiter{
		capacity:           capacity,
		allowedRequestChan: make(chan *LeakyBucketRateLimiterRequest),
		stopChan:           make(chan bool),
	}
	r.q = queue.NewFixedFIFO(int(r.capacity))
	r.SetFrequency(f)
	go r.start()
	return r
}

// SetFrequency() updates the frequency of rate limiter.
// It will restart the background process that leaks request from queue with new frequency.
func (r *LeakyBucketRateLimiter) SetFrequency(f Frequency) {
	defer r.Stop()
	if f.Unit != 1 {
		panic(`Frequency.Unit for LeakyBucketRateLimiter must be 1.
To configure RateLimiter that allows 100 requests per second, use Frequency{Unit: int64(1), Interval: time.Millisecond}`)
	}
	r.frequency = f
	r.Stop()
	go r.start()
}

// AllowRequest() rejects request immediately when the requests in queue is at its capacity.
// If the requests in queue are less than capacity, it enqueues the new request and waits
// until it is being dequeued based on FIFO.
func (r *LeakyBucketRateLimiter) AllowRequest(req Request) bool {
	request, ok := req.(*LeakyBucketRateLimiterRequest)
	if !ok {
		panic(`Invalid request type for LeakyBucketRateLimiterRequest`)
	}
	if err := r.q.Enqueue(request); err != nil {
		return false
	}
	<-r.allowedRequestChan
	return true
}

// start() dequeues requests based on its frequency until t.Stop() is called.
// It will run in background with a goroutine.
func (r *LeakyBucketRateLimiter) start() {
	if r.isStarted {
		return
	}
	r.isStarted = true
	ticker := time.NewTicker(r.frequency.Interval)
	for {
		select {
		case <-ticker.C:
			item, err := r.q.DequeueOrWaitForNextElement()
			if err != nil {
				panic(err)
			}
			allowedRequest, _ := item.(*LeakyBucketRateLimiterRequest)
			r.allowedRequestChan <- allowedRequest
		case <-r.stopChan:
			r.q = queue.NewFixedFIFO(int(r.capacity))
			ticker.Stop()
			return
		}
	}
}

// Stop() stops the background process that dequeues requests.
func (r *LeakyBucketRateLimiter) Stop() {
	if !r.isStarted {
		return
	}
	r.isStarted = false
	r.stopChan <- true
}
