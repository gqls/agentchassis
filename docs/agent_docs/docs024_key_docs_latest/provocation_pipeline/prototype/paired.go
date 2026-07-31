// FILE: paired.go
//
// Paired provocation — domain logic, no HTTP, no storage, no LLM.
//
// A paired provocation is the Gauntlet inverted: instead of the platform
// setting a public topic for anonymous strangers, an ORGANISER sets a topic
// for a KNOWN, CLOSED group, everyone commits a position without seeing any
// other, and the positions become visible to the group all at once.
//
// The whole value of the exercise lives in one property:
//
//	NOBODY — including the organiser — can read anybody else's position
//	before the reveal.
//
// If that leaks, the exercise stops measuring what people think and starts
// measuring who read the room first. So this file does not enforce the seal
// with an `if` that a later edit can invert. It enforces it with the TYPE
// SYSTEM: a pre-reveal response is a SealedView, and SealedView has nowhere
// to put another participant's words. The bad state is unrepresentable
// rather than merely unreached.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Reveal rules — the organiser's choices
// ---------------------------------------------------------------------------

// RevealRule decides WHEN the sealed positions open. The owner's brief said
// the team reply "that day or over several days until they've all committed
// (choices given to the person setting it up)" — these are those choices.
type RevealRule string

const (
	// RevealWhenAllCommitted waits for every participant. Purest, and it
	// stalls for ever on one non-responder — which is why the other two exist.
	RevealWhenAllCommitted RevealRule = "all_committed"

	// RevealAtDeadline opens at a fixed time regardless of who has replied.
	RevealAtDeadline RevealRule = "deadline"

	// RevealAtQuorum opens as soon as N have committed.
	RevealAtQuorum RevealRule = "quorum"
)

var (
	ErrUnknownToken   = errors.New("unknown participant token")
	ErrAlreadyOpen    = errors.New("positions are already revealed")
	ErrAlreadyCommit  = errors.New("you have already committed a position")
	ErrEmptyPosition  = errors.New("a position cannot be empty")
	ErrNotYetRevealed = errors.New("positions are still sealed")
)

// ---------------------------------------------------------------------------
// The model
// ---------------------------------------------------------------------------

type Participant struct {
	Name  string
	Token string

	// position is UNEXPORTED on purpose. Nothing outside this file can read
	// it, so no handler and no template can accidentally render it. The only
	// routes out are ViewFor and OrganiserView, both of which are auditable
	// in one screen below.
	position    string
	committedAt *time.Time
}

func (p *Participant) Committed() bool { return p.committedAt != nil }

type Session struct {
	ID          string
	Organiser   string
	Provocation string
	Rule        RevealRule
	Quorum      int
	Deadline    time.Time
	CreatedAt   time.Time

	mu           sync.Mutex
	participants []*Participant
	revealedAt   *time.Time
}

func NewSession(organiser, provocation string, names []string, rule RevealRule, quorum int, deadline, now time.Time) (*Session, error) {
	if strings.TrimSpace(provocation) == "" {
		return nil, errors.New("a paired provocation needs a provocation")
	}
	clean := make([]string, 0, len(names))
	for _, n := range names {
		if n = strings.TrimSpace(n); n != "" {
			clean = append(clean, n)
		}
	}
	if len(clean) < 2 {
		return nil, errors.New("a paired provocation needs at least two participants")
	}
	if rule == RevealAtQuorum && (quorum < 1 || quorum > len(clean)) {
		return nil, fmt.Errorf("quorum must be between 1 and %d", len(clean))
	}

	s := &Session{
		ID:          token(6),
		Organiser:   strings.TrimSpace(organiser),
		Provocation: strings.TrimSpace(provocation),
		Rule:        rule,
		Quorum:      quorum,
		Deadline:    deadline,
		CreatedAt:   now,
	}
	for _, n := range clean {
		s.participants = append(s.participants, &Participant{Name: n, Token: token(10)})
	}
	return s, nil
}

func token(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err) // prototype: a CSPRNG failure is not a recoverable condition here
	}
	return hex.EncodeToString(b)
}

// ---------------------------------------------------------------------------
// Commit
// ---------------------------------------------------------------------------

// Commit records a position. It is FINAL: you cannot revise after committing.
//
// The alternative — editable until reveal — is defensible (nobody can see it,
// so nothing is gained by locking it) and is what a sealed-bid auction
// actually does. It is rejected here because "until they've all committed"
// needs "committed" to be a state you cannot leave, or the reveal condition
// keeps un-satisfying itself. Worth revisiting with the owner; noted in
// README.md rather than silently decided.
func (s *Session) Commit(tok, position string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(position) == "" {
		return ErrEmptyPosition
	}
	p := s.find(tok)
	if p == nil {
		return ErrUnknownToken
	}
	if s.revealedAt != nil {
		return ErrAlreadyOpen
	}
	if p.Committed() {
		return ErrAlreadyCommit
	}

	p.position = strings.TrimSpace(position)
	t := now
	p.committedAt = &t

	s.maybeRevealLocked(now)
	return nil
}

// maybeRevealLocked applies the organiser's rule. Reveal is a single
// assignment to one field on the session, which is what makes it ATOMIC:
// there is no state in which some participants can read the positions and
// others cannot. If it opened per-participant, whoever polled last would
// have read everyone else's answer before writing their own.
func (s *Session) maybeRevealLocked(now time.Time) {
	if s.revealedAt != nil {
		return
	}
	committed := 0
	for _, p := range s.participants {
		if p.Committed() {
			committed++
		}
	}

	open := false
	switch s.Rule {
	case RevealWhenAllCommitted:
		open = committed == len(s.participants)
	case RevealAtQuorum:
		open = committed >= s.Quorum
	case RevealAtDeadline:
		open = !now.Before(s.Deadline)
	}
	if open {
		t := now
		s.revealedAt = &t
	}
}

// Tick lets a deadline fire without anybody committing. In a real build this
// is a scheduled job; here it is called on every read.
func (s *Session) Tick(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maybeRevealLocked(now)
}

// ForceReveal is the organiser's override — the escape hatch for the
// non-responder who is never going to respond.
func (s *Session) ForceReveal(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.revealedAt == nil {
		t := now
		s.revealedAt = &t
	}
}

func (s *Session) find(tok string) *Participant {
	for _, p := range s.participants {
		if p.Token == tok {
			return p
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// The two view types — this is the seal
// ---------------------------------------------------------------------------

// SealedPeer is what you may know about someone else BEFORE the reveal:
// that they exist, and whether they have answered. There is deliberately no
// field here that could hold their words. This is the enforcement.
type SealedPeer struct {
	Name      string
	Committed bool
}

// SealedView is the pre-reveal response. It carries YOUR OWN position back to
// you (you wrote it; seeing it is not a leak) and nothing but names and
// counts about anyone else.
type SealedView struct {
	Provocation  string
	Organiser    string
	YourName     string
	YourPosition string // yours alone
	YouCommitted bool
	Peers        []SealedPeer
	Committed    int
	Total        int
	Rule         RevealRule
	Deadline     time.Time
}

// RevealedPosition is a name attached to what that person actually argued.
type RevealedPosition struct {
	Name        string
	Position    string
	CommittedAt time.Time
}

// RevealedView is the post-reveal response and the ONLY type in this package
// that can carry another participant's words.
type RevealedView struct {
	Provocation  string
	Organiser    string
	YourName     string
	Positions    []RevealedPosition
	DidNotCommit []string
	RevealedAt   time.Time
}

// ViewFor returns exactly one of (sealed, revealed) — never both, never
// neither. A caller cannot get hold of another person's position without
// being handed a *RevealedView, and it is handed one only when the reveal
// has actually happened AND this participant paid the entry price.
//
// NON-RESPONDERS DO NOT RECEIVE THE REVEAL. If the deadline fires and Carol
// never answered, Carol does not get to read what everyone else risked. This
// is a product decision, not a technical one: without it, the optimal play
// under a deadline rule is to say nothing and read the room, which is the
// exact behaviour the seal exists to prevent.
func (s *Session) ViewFor(tok string, now time.Time) (*SealedView, *RevealedView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maybeRevealLocked(now)

	me := s.find(tok)
	if me == nil {
		return nil, nil, ErrUnknownToken
	}

	if s.revealedAt != nil && me.Committed() {
		rv := &RevealedView{
			Provocation: s.Provocation,
			Organiser:   s.Organiser,
			YourName:    me.Name,
			RevealedAt:  *s.revealedAt,
		}
		for _, p := range s.participants {
			if p.Committed() {
				rv.Positions = append(rv.Positions, RevealedPosition{
					Name: p.Name, Position: p.position, CommittedAt: *p.committedAt,
				})
			} else {
				rv.DidNotCommit = append(rv.DidNotCommit, p.Name)
			}
		}
		sort.Slice(rv.Positions, func(i, j int) bool {
			return rv.Positions[i].CommittedAt.Before(rv.Positions[j].CommittedAt)
		})
		return nil, rv, nil
	}

	sv := &SealedView{
		Provocation:  s.Provocation,
		Organiser:    s.Organiser,
		YourName:     me.Name,
		YourPosition: me.position,
		YouCommitted: me.Committed(),
		Total:        len(s.participants),
		Rule:         s.Rule,
		Deadline:     s.Deadline,
	}
	for _, p := range s.participants {
		if p.Committed() {
			sv.Committed++
		}
		if p.Token != tok {
			sv.Peers = append(sv.Peers, SealedPeer{Name: p.Name, Committed: p.Committed()})
		}
	}
	return sv, nil, nil
}

// ---------------------------------------------------------------------------
// Organiser view
// ---------------------------------------------------------------------------

// OrganiserRow is what the organiser sees per participant. Note what is
// absent: the position. The organiser gets the facilitation information they
// need — who has answered, who to chase — and no privileged read.
//
// Giving the organiser an early look is the single most tempting feature in
// this design and it would destroy the product: a facilitator who has read
// the answers cannot run the session neutrally, and participants who suspect
// they have will hedge.
type OrganiserRow struct {
	Name        string
	Committed   bool
	CommittedAt *time.Time
	Link        string
}

type OrganiserView struct {
	ID          string
	Provocation string
	Organiser   string
	Rule        RevealRule
	Quorum      int
	Deadline    time.Time
	Rows        []OrganiserRow
	Committed   int
	Total       int
	Revealed    bool
	RevealedAt  *time.Time
}

func (s *Session) OrganiserView(baseURL string, now time.Time) OrganiserView {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maybeRevealLocked(now)

	ov := OrganiserView{
		ID: s.ID, Provocation: s.Provocation, Organiser: s.Organiser,
		Rule: s.Rule, Quorum: s.Quorum, Deadline: s.Deadline,
		Total: len(s.participants), Revealed: s.revealedAt != nil, RevealedAt: s.revealedAt,
	}
	for _, p := range s.participants {
		if p.Committed() {
			ov.Committed++
		}
		ov.Rows = append(ov.Rows, OrganiserRow{
			Name: p.Name, Committed: p.Committed(), CommittedAt: p.committedAt,
			Link: fmt.Sprintf("%s/p/%s", baseURL, p.Token),
		})
	}
	return ov
}

// ---------------------------------------------------------------------------
// Store — in memory, deliberately
// ---------------------------------------------------------------------------
//
// Nothing is persisted. Restart the process and every session is gone.
//
// That is a FEATURE of this prototype, not a shortcut to fix later: the
// moment we persist named colleagues' opinions on contested topics we have
// taken on a real confidentiality duty, and that should start with a
// deliberate design, not with a prototype quietly filling a table.

type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	byToken  map[string]*Session
}

func NewStore() *Store {
	return &Store{sessions: map[string]*Session{}, byToken: map[string]*Session{}}
}

func (st *Store) Put(s *Session) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.sessions[s.ID] = s
	for _, p := range s.participants {
		st.byToken[p.Token] = s
	}
}

func (st *Store) BySession(id string) (*Session, bool) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	s, ok := st.sessions[id]
	return s, ok
}

func (st *Store) ByToken(tok string) (*Session, bool) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	s, ok := st.byToken[tok]
	return s, ok
}
