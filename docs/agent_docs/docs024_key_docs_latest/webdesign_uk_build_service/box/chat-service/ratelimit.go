package main

// ratelimit.go — per-IP request limiting. Sliding-window counter, adapted from
// idea.uk/golang_files/audience_check.go's rateLimiter. In-memory only: state
// is lost on restart, which is fine for this risk profile (the daily SPEND
// ceiling in spend.go is the control that must survive a restart — see there).

import (
	"sync"
	"time"
)

type rateBand struct {
	window time.Duration
	max    int
}

type rateLimiter struct {
	mu    sync.Mutex
	hits  map[string][]time.Time
	bands []rateBand
}

// newChatIPLimiter bounds new-conversation starts per visitor. Two bands so a
// single burst can't be follow by a slow drip that never trips an hourly cap:
// 5/hour stops a script hammering the endpoint; 15/day stops a slow trickle
// from the same address across a working day.
func newChatIPLimiter() *rateLimiter {
	return &rateLimiter{
		hits: map[string][]time.Time{},
		bands: []rateBand{
			{window: time.Hour, max: 5},
			{window: 24 * time.Hour, max: 15},
		},
	}
}

// allow returns (ok, retryAfter). retryAfter is the earliest moment the key
// could try again if currently throttled; zero when allowed.
func (rl *rateLimiter) allow(key string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	longest := rl.bands[0].window
	for _, b := range rl.bands {
		if b.window > longest {
			longest = b.window
		}
	}
	cutoff := now.Add(-longest)
	hits := rl.hits[key]
	pruned := hits[:0]
	for _, t := range hits {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}
	for _, b := range rl.bands {
		bandCutoff := now.Add(-b.window)
		count := 0
		var oldestInBand time.Time
		for _, t := range pruned {
			if t.After(bandCutoff) {
				if count == 0 || t.Before(oldestInBand) {
					oldestInBand = t
				}
				count++
			}
		}
		if count >= b.max {
			retryAfter := oldestInBand.Add(b.window).Sub(now)
			if retryAfter < 0 {
				retryAfter = 0
			}
			rl.hits[key] = pruned
			return false, retryAfter
		}
	}
	pruned = append(pruned, now)
	rl.hits[key] = pruned
	return true, 0
}
