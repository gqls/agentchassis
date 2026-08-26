package main

// orders.go — the committed-brief store: the box half of the order-intake
// connection (owner GO 2026-08-26, PLAN_2026-07-31_p4_order_intake §8).
//
// A visitor who approves their brief in chat no longer copies it into an
// email: the model calls submit_brief (chat.go), the brief lands here with a
// minted order reference, and the CLUSTER collects it later over the internal
// endpoints (orders_http.go) — the box never dials in, per the standing trust
// boundary. The reference is the join key the owner ruled for (2026-08-26):
// the customer quotes it at payment, billing_orders carries it as
// external_reference, and the collector releases a brief to build_queue only
// when a PAID billing order names it. The brief itself deliberately does NOT
// travel into billing — briefs change; the reference does not.
//
// Same persistence pattern as store.go: one JSON file, atomic tmp+rename
// rewrite, coarse lock. Volumes are human-scale (a handful of briefs a day at
// the very best), so a database would be pure liability on a box that holds
// no cluster credential.

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// BriefOrder is one committed brief. CollectedAt is the collection marker the
// cluster's acknowledge sets — nil means "not yet collected", which is exactly
// what the list endpoint serves. Nothing here is ever deleted; a collected
// order is history, and history on this box is cheap.
type BriefOrder struct {
	Reference      string     `json:"reference"`
	ConversationID string     `json:"conversation_id"`
	ClientIP       string     `json:"client_ip"`
	ContactEmail   string     `json:"contact_email"`
	ContactName    string     `json:"contact_name,omitempty"`
	Domain         string     `json:"domain,omitempty"`
	Brief          string     `json:"brief"`
	CreatedAt      time.Time  `json:"created_at"`
	CollectedAt    *time.Time `json:"collected_at,omitempty"`
}

// Reference alphabet: unambiguous (no 0/O/1/I/L), matching the voucher-code
// alphabet the estate already uses. Prefix "BR-" (brief), deliberately NOT
// "WD-": vouchers are WD-…, and a customer reading two codes that look alike
// will quote the wrong one at exactly the moment it matters.
const refAlphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"

// maxBriefsPerConversation bounds what one chat session can commit. Three
// allows a genuine "actually, change it and resubmit" without letting a
// runaway conversation fill the store; the IP limiter and turn cap bound the
// rest.
const maxBriefsPerConversation = 3

// Validation bounds. The brief floor exists because a one-line "brief" is a
// submission the build cannot use and the visitor was almost certainly
// mis-served by the model; the tool_result error sends it back to the
// conversation where it can be fixed, which an accepted-then-unusable
// submission cannot be.
const (
	minBriefLen = 40
	maxBriefLen = 10000
	maxNameLen  = 200
)

var errTooManySubmissions = errors.New("this conversation has already submitted the maximum number of briefs")

type OrderStore struct {
	mu     sync.Mutex
	path   string
	orders map[string]*BriefOrder // by reference
}

func NewOrderStore(path string) (*OrderStore, error) {
	s := &OrderStore{path: path, orders: map[string]*BriefOrder{}}
	if b, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(b, &s.orders); err != nil {
			return nil, fmt.Errorf("orders store %s unreadable: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// ValidateSubmission checks the fields a submit_brief call carries, returning
// a message written for the MODEL to act on (it becomes the tool_result
// error), or "" when the submission is acceptable. Domain is optional and
// loosely checked: the visitor may not have decided, and a wrong domain is a
// conversation to have, not a build to block.
func ValidateSubmission(email, name, domain, brief string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return "contact_email is missing. Ask the visitor for an email address to reach them on before submitting."
	}
	if len(email) > 254 {
		return "contact_email is too long to be a real address. Ask the visitor to give it again."
	}
	if a, err := mail.ParseAddress(email); err != nil || a.Address != email {
		return "contact_email does not look like a plain email address. Ask the visitor to give just the address, like name@example.com."
	}
	if len(name) > maxNameLen {
		return "contact_name is too long. Use the short name they actually gave."
	}
	if d := strings.TrimSpace(domain); d != "" {
		if len(d) > 253 || strings.ContainsAny(d, " \t\n/@:") || !strings.Contains(d, ".") {
			return "domain does not look like a domain name. Give it as just the name, like example.co.uk, or leave it out if they have not decided."
		}
	}
	b := strings.TrimSpace(brief)
	if len(b) < minBriefLen {
		return "the brief is too short to build from. Write back the full brief the visitor approved, not a summary of it."
	}
	if len(b) > maxBriefLen {
		return "the brief is too long to submit. Keep the visitor's specifics but trim it to the agreed shape."
	}
	return ""
}

// Submit stores a validated brief and returns it with its minted reference.
// Callers run ValidateSubmission first; Submit re-checks nothing but the
// per-conversation cap, which only the store can see.
func (s *OrderStore) Submit(conversationID, clientIP, email, name, domain, brief string) (BriefOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for _, o := range s.orders {
		if o.ConversationID == conversationID {
			count++
		}
	}
	if count >= maxBriefsPerConversation {
		return BriefOrder{}, errTooManySubmissions
	}

	ref, err := s.mintReferenceLocked()
	if err != nil {
		return BriefOrder{}, err
	}
	o := &BriefOrder{
		Reference:      ref,
		ConversationID: conversationID,
		ClientIP:       clientIP,
		ContactEmail:   strings.TrimSpace(email),
		ContactName:    strings.TrimSpace(name),
		Domain:         strings.ToLower(strings.TrimSpace(domain)),
		Brief:          strings.TrimSpace(brief),
		CreatedAt:      time.Now().UTC(),
	}
	s.orders[ref] = o
	if err := s.saveLocked(); err != nil {
		delete(s.orders, ref) // an unpersisted order must not be quotable
		return BriefOrder{}, err
	}
	return *o, nil
}

// ListUncollected returns every order the cluster has not yet acknowledged,
// oldest first — the collector processes in arrival order.
func (s *OrderStore) ListUncollected() []BriefOrder {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []BriefOrder
	for _, o := range s.orders {
		if o.CollectedAt == nil {
			out = append(out, *o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

// Ack marks the named references collected and returns how many actually
// changed state. Idempotent by construction: an already-collected reference
// keeps its ORIGINAL CollectedAt (a lost ack response retried later must not
// rewrite history), and an unknown reference is skipped, not an error — the
// collector may legitimately re-ack after a partial failure.
func (s *OrderStore) Ack(references []string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	changed := 0
	now := time.Now().UTC()
	for _, ref := range references {
		if o, ok := s.orders[ref]; ok && o.CollectedAt == nil {
			t := now
			o.CollectedAt = &t
			changed++
		}
	}
	if changed > 0 {
		if err := s.saveLocked(); err != nil {
			return changed, err
		}
	}
	return changed, nil
}

func (s *OrderStore) mintReferenceLocked() (string, error) {
	for attempt := 0; attempt < 10; attempt++ {
		b := make([]byte, 6)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		out := make([]byte, 0, 9)
		out = append(out, 'B', 'R', '-')
		for _, c := range b {
			out = append(out, refAlphabet[int(c)%len(refAlphabet)])
		}
		ref := string(out)
		if _, taken := s.orders[ref]; !taken {
			return ref, nil
		}
	}
	return "", errors.New("could not mint a unique reference in 10 attempts")
}

func (s *OrderStore) saveLocked() error {
	b, err := json.MarshalIndent(s.orders, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
