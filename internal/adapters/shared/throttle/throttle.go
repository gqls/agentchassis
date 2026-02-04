// FILE: adapters/shared/throttle/throttle.go
// Simple request throttling for adapters.
// Reads REQUEST_THROTTLE_MS from environment (default: 0 = no throttle).
// Each adapter calls throttle.Wait() after processing a request.
//
// For the business intelligence pipeline:
//   - webscrape adapter: REQUEST_THROTTLE_MS=5000 (5s between scrapes)
//   - web search adapter: REQUEST_THROTTLE_MS=5000 (5s between searches)
//   - For non-BI workloads: leave at 0 or don't set the env var
//
// Per-domain throttling is a future enhancement.
// For now, single-pod sequential processing with a global delay is sufficient.

package throttle

import (
	"os"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Throttle controls the minimum delay between external requests.
type Throttle struct {
	mu          sync.Mutex
	minDelay    time.Duration
	lastRequest time.Time
	logger      *zap.Logger
}

// New creates a throttle from the REQUEST_THROTTLE_MS environment variable.
// If the env var is not set or is 0, no throttling occurs.
func New(logger *zap.Logger) *Throttle {
	delayMs := 0
	if val := os.Getenv("REQUEST_THROTTLE_MS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			delayMs = parsed
		}
	}

	t := &Throttle{
		minDelay: time.Duration(delayMs) * time.Millisecond,
		logger:   logger,
	}

	if t.minDelay > 0 {
		logger.Info("Request throttle enabled",
			zap.Duration("min_delay", t.minDelay),
		)
	}

	return t
}

// NewWithDelay creates a throttle with a specific delay (for testing or explicit config).
func NewWithDelay(delay time.Duration, logger *zap.Logger) *Throttle {
	return &Throttle{
		minDelay: delay,
		logger:   logger,
	}
}

// Wait blocks until enough time has passed since the last request.
// Call this AFTER processing each message (after the response has been sent).
// Returns the actual time waited (0 if no wait was needed).
func (t *Throttle) Wait() time.Duration {
	if t.minDelay == 0 {
		return 0
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	elapsed := time.Since(t.lastRequest)
	if elapsed < t.minDelay {
		waitTime := t.minDelay - elapsed
		t.logger.Debug("Throttle: waiting between requests",
			zap.Duration("wait_time", waitTime),
			zap.Duration("elapsed", elapsed),
		)
		time.Sleep(waitTime)
		t.lastRequest = time.Now()
		return waitTime
	}

	t.lastRequest = time.Now()
	return 0
}

// MinDelay returns the configured minimum delay.
func (t *Throttle) MinDelay() time.Duration {
	return t.minDelay
}
