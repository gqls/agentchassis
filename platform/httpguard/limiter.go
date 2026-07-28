package httpguard

import (
	"sync"
	"time"
)

// Band is one rate ceiling: at most Max events per Window.
type Band struct {
	Window time.Duration
	Max    int
}

// Limiter is an in-memory, per-key sliding-window limiter over one or more bands.
//
// Chosen over the token bucket it replaces for two reasons that matter on a
// public endpoint. It returns a RETRY-AFTER, so a throttled caller can be told
// when to come back instead of guessing; and it supports several bands at once,
// which is how you say "a few per hour AND a sane daily ceiling" — a single
// bucket can express only one of those.
//
// In-memory and per-process: a restart or a second replica loses state. That is
// the accepted profile for the public tools estate, where this is a backstop
// against a flood from one address rather than a billing control. If it ever
// needs to be exact across replicas it belongs in Postgres, and that is a
// different design, not a tweak to this one.
type Limiter struct {
	mu      sync.Mutex
	hits    map[string][]time.Time
	bands   []Band
	longest time.Duration
	now     func() time.Time // injectable so tests need no sleeps
}

// NewLimiter builds a limiter over the given bands. Panics on no bands, because
// a limiter that permits everything is never what the caller meant and failing
// at construction beats failing silently in production.
func NewLimiter(bands ...Band) *Limiter {
	if len(bands) == 0 {
		panic("httpguard: NewLimiter needs at least one band")
	}
	l := &Limiter{hits: map[string][]time.Time{}, bands: bands, now: time.Now}
	for _, b := range bands {
		if b.Window > l.longest {
			l.longest = b.Window
		}
	}
	return l
}

// Allow records an event for key and reports whether it is permitted. When it is
// not, retryAfter is how long until the oldest event in the offending band ages
// out. A refused event is NOT recorded, so being throttled cannot extend the
// throttle.
func (l *Limiter) Allow(key string) (ok bool, retryAfter time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	pruned := l.hits[key][:0]
	cutoff := now.Add(-l.longest)
	for _, t := range l.hits[key] {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}

	for _, b := range l.bands {
		bandCutoff := now.Add(-b.Window)
		count := 0
		var oldest time.Time
		for _, t := range pruned {
			if t.After(bandCutoff) {
				if count == 0 || t.Before(oldest) {
					oldest = t
				}
				count++
			}
		}
		if count >= b.Max {
			l.hits[key] = pruned
			return false, oldest.Add(b.Window).Sub(now)
		}
	}

	l.hits[key] = append(pruned, now)
	return true, 0
}

// Forget drops a key's history. For tests and for an operator clearing a
// false positive.
func (l *Limiter) Forget(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.hits, key)
}

// Sweep drops keys with no events inside the longest band. Nothing calls this
// automatically: the map is bounded by distinct callers per window, which is
// fine for this estate, and a caller that wants it can run it on a ticker.
func (l *Limiter) Sweep() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := l.now().Add(-l.longest)
	dropped := 0
	for k, hits := range l.hits {
		live := false
		for _, t := range hits {
			if t.After(cutoff) {
				live = true
				break
			}
		}
		if !live {
			delete(l.hits, k)
			dropped++
		}
	}
	return dropped
}
