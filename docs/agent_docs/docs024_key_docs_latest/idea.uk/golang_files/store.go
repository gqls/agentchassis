package main

// store.go — persistence. A JSON-file store (stdlib only) so the service runs
// standalone with no DB driver. Production should swap in the chassis Postgres
// behind the same small method set. Coarse locking is fine at early-access volume.

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Order lifecycle: requested → (declined | awaiting_payment) → paid → running
//
//	→ (awaiting_review | delivered | failed)
//
// Terminal statuses that release a capacity slot: delivered, declined, failed,
// and `expired`. `expired` is deliberately DISTINCT from `declined`: declined
// means the operator looked and said no; expired means the order went cold and
// was released automatically. Keeping them apart keeps the record honest — an
// abandoned order must not be filed as a judgement we made.
type Order struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	Domain            string    `json:"domain"`
	Audience          string    `json:"audience"`
	Assets            string    `json:"assets"`
	Status            string    `json:"status"`
	Report            string    `json:"report"`
	ReportHTML        string    `json:"report_html,omitempty"`
	ProviderSessionID string    `json:"provider_session_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	// Captured at intake so a future IP block list can be seeded from real orders.
	// Empty on rows created before this was added (2026-07-16).
	IP        string `json:"ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

type Store struct {
	mu     sync.Mutex
	path   string
	Orders map[string]*Order `json:"orders"`
	Events map[string]bool   `json:"events"` // processed webhook event ids
	Subs   map[string]bool   `json:"subs"`   // subscriber emails
}

func NewStore(path string) (*Store, error) {
	s := &Store{
		path:   path,
		Orders: map[string]*Order{},
		Events: map[string]bool{},
		Subs:   map[string]bool{},
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
	return s, nil
}

// persist must be called with the lock held.
func (s *Store) persist() {
	if s.path == "" {
		return
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	tmp := s.path + ".tmp"
	if os.WriteFile(tmp, b, 0o600) == nil {
		_ = os.Rename(tmp, s.path) // atomic-ish replace
	}
}

func (s *Store) Save(o *Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *o
	s.Orders[o.ID] = &cp
	s.persist()
}

func (s *Store) Get(id string) (*Order, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	o, ok := s.Orders[id]
	if !ok {
		return nil, false
	}
	cp := *o // hand back a copy; callers mutate via Update
	return &cp, true
}

// Update applies mut to the stored order under lock and persists.
func (s *Store) Update(id string, mut func(*Order)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	o, ok := s.Orders[id]
	if !ok {
		return false
	}
	mut(o)
	o.UpdatedAt = time.Now().UTC()
	s.persist()
	return true
}

// ActiveCount = orders occupying a fulfilment slot (committed, not finished).
func (s *Store) ActiveCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, o := range s.Orders {
		switch o.Status {
		case "awaiting_payment", "paid", "running", "awaiting_review":
			n++
		}
	}
	return n
}

// MarkEventSeen returns true if the event was already processed (idempotency).
// ExpireStale releases capacity slots held by orders that have gone cold, and
// returns what it released (id + previous status) for logging.
//
// Why this exists: ActiveCount counts awaiting_review/awaiting_payment against
// MaxActive, and NOTHING ever aged those out — an order the operator never got
// to, or a customer who never paid, held a slot permanently. With MaxActive at
// 5 that silently closed the service to new work, and the only remedy was
// hand-editing this file, which is itself a trap (the store is read once at
// startup and rewritten wholesale from memory, so an edit under a running
// service is both invisible and doomed to be clobbered).
//
// The two ages are separate on purpose: an unreviewed order is waiting on US,
// so it gets longer; an unpaid one is waiting on the customer, and a stale
// checkout is abandoned much sooner in practice.
//
// Idempotent, safe to call on a ticker, and never touches a terminal order or
// one that is mid-run.
func (s *Store) ExpireStale(reviewAge, paymentAge time.Duration, now time.Time) []ExpiredOrder {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []ExpiredOrder
	for _, o := range s.Orders {
		var age time.Duration
		switch o.Status {
		case "awaiting_review":
			age = reviewAge
		case "awaiting_payment":
			age = paymentAge
		default:
			continue // requested/paid/running/terminal — not ours to expire
		}
		// UpdatedAt is when it entered this state; fall back to CreatedAt for
		// rows written before UpdatedAt was maintained.
		ref := o.UpdatedAt
		if ref.IsZero() {
			ref = o.CreatedAt
		}
		if ref.IsZero() || now.Sub(ref) < age {
			continue
		}
		out = append(out, ExpiredOrder{ID: o.ID, WasStatus: o.Status, Age: now.Sub(ref)})
		o.Status = "expired"
		o.UpdatedAt = now
	}
	if len(out) > 0 {
		s.persist()
	}
	return out
}

// ExpiredOrder is one released slot, for the operator log.
type ExpiredOrder struct {
	ID        string
	WasStatus string
	Age       time.Duration
}

func (s *Store) MarkEventSeen(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Events[id] {
		return true
	}
	s.Events[id] = true
	s.persist()
	return false
}

func (s *Store) AddSubscriber(email string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Subs[email] = true
	s.persist()
}
