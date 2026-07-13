package main

// store.go — persistence for site-engine intent events. Forked from idea.uk's store.go:
// same JSON-file model (stdlib only, coarse lock, atomic-ish rename on persist),
// so it runs standalone with no DB driver and the on-box file can be checkpointed
// to B2 later. What changed from idea.uk:
//   - Order            → IntentEvent   (we capture stated intent, not paid orders)
//   - Orders/Events/Subs maps → Events (host → []*IntentEvent) + Visits (host → n)
// Everything is keyed by canonical host (domains.go::canonicalHost) so one binary
// serving many vhosts keeps each domain's data cleanly separated.
//
// Privacy (UK GDPR/PECR, low risk appetite): we do NOT store IP addresses. Referer
// is reduced to its host before it reaches here; Country is only ever a coarse
// CDN-supplied code or empty. Free-text Value is treated as potentially personal —
// retention is enforced by the periodic checkpoint/prune step (later), not here.

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// IntentEvent is one action a visitor took on a probe page.
type IntentEvent struct {
	ID        string    `json:"id"`
	Host      string    `json:"host"`       // canonical host the event belongs to
	Kind      string    `json:"kind"`       // ProbeSearch | ProbeCategory | ProbeFreeText
	Value     string    `json:"value"`      // search term / chosen category / free text
	RefHost   string    `json:"ref_host"`   // referer reduced to host only ("" if none/same)
	Country   string    `json:"country"`    // coarse CDN country code, or ""
	CreatedAt time.Time `json:"created_at"`
}

// flushInterval is how often the background flusher persists visit-count
// changes. Events are persisted immediately; only visit increments ride the
// flusher, so a hard crash loses at most this window of COUNTS, never events.
const flushInterval = 5 * time.Second

type Store struct {
	mu     sync.Mutex
	path   string
	dirty  bool // unpersisted visit increments pending (NEW field)
	Events map[string][]*IntentEvent `json:"events"` // host → events
	Visits map[string]int            `json:"visits"` // host → page-load count
}

func NewStore(path string) (*Store, error) {
	s := &Store{
		path:   path,
		Events: map[string][]*IntentEvent{},
		Visits: map[string]int{},
	}
	if path == "" {
		return s, nil // in-memory only (tests)
	}
	b, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(b, s)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	// Unmarshal of an older/empty file can leave nil maps; guarantee non-nil.
	if s.Events == nil {
		s.Events = map[string][]*IntentEvent{}
	}
	if s.Visits == nil {
		s.Visits = map[string]int{}
	}
	go s.flushLoop() // background persist of dirty visit counts (NEW)
	return s, nil
}

// persist must be called with the lock held. Compact Marshal (not Indent):
// the file is machine-read; smaller writes matter at volume.
func (s *Store) persist() {
	if s.path == "" {
		return
	}
	b, _ := json.Marshal(s)
	tmp := s.path + ".tmp"
	if os.WriteFile(tmp, b, 0o600) == nil {
		_ = os.Rename(tmp, s.path) // atomic-ish replace
	}
	s.dirty = false
}

// flushLoop persists pending visit increments at most once per flushInterval.
// This replaces the old persist-per-visit behaviour, which rewrote the whole
// (ever-growing) file on EVERY beacon hit — the known scaling cliff.
func (s *Store) flushLoop() {
	for range time.Tick(flushInterval) {
		s.mu.Lock()
		if s.dirty {
			s.persist()
		}
		s.mu.Unlock()
	}
}

// Flush persists any pending changes now. Called on shutdown (SIGTERM/SIGINT).
func (s *Store) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dirty {
		s.persist()
	}
}

// AddVisit counts one page load for a host (denominator for intent-per-1k).
// BEHAVIOUR CHANGE (deliberate): no longer persists per call — marks dirty for
// the flusher instead.
func (s *Store) AddVisit(host string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Visits[host]++
	s.dirty = true
}

// AddEvent records one intent event under its host. Persists immediately —
// captured intent is the product; it never waits on the flusher.
func (s *Store) AddEvent(ev *IntentEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *ev
	s.Events[ev.Host] = append(s.Events[ev.Host], &cp)
	s.persist()
}

// HostStat is the per-domain summary used by /stats and ranking.
type HostStat struct {
	Host             string  `json:"host"`
	Visits           int     `json:"visits"`
	Events           int     `json:"events"`
	EventsPer1kVisit float64 `json:"events_per_1k_visits"`
}

// Stats returns a per-host summary (no event bodies), safe to expose internally.
func (s *Store) Stats() []HostStat {
	s.mu.Lock()
	defer s.mu.Unlock()
	hosts := map[string]bool{}
	for h := range s.Visits {
		hosts[h] = true
	}
	for h := range s.Events {
		hosts[h] = true
	}
	out := make([]HostStat, 0, len(hosts))
	for h := range hosts {
		v := s.Visits[h]
		e := len(s.Events[h])
		var per float64
		if v > 0 {
			per = float64(e) * 1000 / float64(v)
		}
		out = append(out, HostStat{Host: h, Visits: v, Events: e, EventsPer1kVisit: per})
	}
	return out
}

// Snapshot returns a deep-ish copy of the whole store for off-box checkpointing
// (the B2 upload step, later). Taken under lock so it is internally consistent.
func (s *Store) Snapshot() *Store {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := &Store{
		Events: make(map[string][]*IntentEvent, len(s.Events)),
		Visits: make(map[string]int, len(s.Visits)),
	}
	for h, n := range s.Visits {
		cp.Visits[h] = n
	}
	for h, evs := range s.Events {
		c := make([]*IntentEvent, len(evs))
		for i, e := range evs {
			ev := *e
			c[i] = &ev
		}
		cp.Events[h] = c
	}
	return cp
}
