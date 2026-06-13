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
