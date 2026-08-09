package main

// store.go — persistence. Stdlib only, no DB driver: this box holds no
// cluster credential and dials nothing in (owner ruling, PLAN §2a "the trust
// boundary" — the box never dials in, never holds a cluster credential). Two
// kinds of file:
//
//  1. state.json — small, rewritten atomically on every change. Holds the
//     ONE thing that must survive a process restart: today's running spend,
//     because the daily ceiling (chat.go) is the fail-closed control and a
//     restart must not reset an attacker's budget to zero. Per-conversation
//     turn counts live here too, for the same reason, at negligible cost.
//  2. requests.jsonl / transcripts.jsonl — append-only. One line per API
//     call and one line per message respectively. This is deliberately NOT
//     one file: a request log entry (tokens, cost, latency) and a transcript
//     entry (what was actually said) answer different questions, and mixing
//     them would force every reader to filter. Transcripts are the demand
//     signal P1 exists to collect (PLAN §4 point 5) — keep them as their own
//     stream so a later ingestion job can tail just that file.
//
// Coarse locking throughout — fine at this volume (idea.uk/store.go's own
// justification, and it holds here too).

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConversationState is the turn-cap bookkeeping for one visitor session.
type ConversationState struct {
	ID           string    `json:"id"`
	ClientIP     string    `json:"client_ip"`
	TurnCount    int       `json:"turn_count"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

type persistedState struct {
	Conversations map[string]*ConversationState `json:"conversations"`
	DailySpendUSD map[string]float64            `json:"daily_spend_usd"` // key: "2026-08-09" (UTC date)
}

// RequestLogEntry is one row per Claude API call — the cost record.
type RequestLogEntry struct {
	Timestamp      time.Time `json:"timestamp"`
	ConversationID string    `json:"conversation_id"`
	ClientIP       string    `json:"client_ip"`
	Model          string    `json:"model"`
	InputTokens    int       `json:"input_tokens"`
	OutputTokens   int       `json:"output_tokens"`
	CostUSD        float64   `json:"cost_usd"`
	LatencyMS      int64     `json:"latency_ms"`
	StopReason     string    `json:"stop_reason,omitempty"`
	Error          string    `json:"error,omitempty"`
}

// TranscriptEntry is one row per message — the demand signal.
type TranscriptEntry struct {
	Timestamp      time.Time `json:"timestamp"`
	ConversationID string    `json:"conversation_id"`
	ClientIP       string    `json:"client_ip"`
	Role           string    `json:"role"` // "user" | "assistant"
	Content        string    `json:"content"`
}

type Store struct {
	mu             sync.Mutex
	statePath      string
	requestLogPath string
	transcriptPath string
	state          *persistedState
}

func NewStore(statePath, requestLogPath, transcriptPath string) (*Store, error) {
	s := &Store{
		statePath:      statePath,
		requestLogPath: requestLogPath,
		transcriptPath: transcriptPath,
		state: &persistedState{
			Conversations: map[string]*ConversationState{},
			DailySpendUSD: map[string]float64{},
		},
	}
	if b, err := os.ReadFile(statePath); err == nil {
		if err := json.Unmarshal(b, s.state); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if s.state.Conversations == nil {
		s.state.Conversations = map[string]*ConversationState{}
	}
	if s.state.DailySpendUSD == nil {
		s.state.DailySpendUSD = map[string]float64{}
	}
	return s, nil
}

// saveLocked writes state atomically (tmp file + rename) so a crash mid-write
// cannot corrupt the file the daily spend ceiling depends on. Caller must
// hold s.mu.
func (s *Store) saveLocked() error {
	b, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.statePath + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.statePath)
}

func today() string {
	return time.Now().UTC().Format("2006-01-02") // Go's reference date, not literally 2026
}

// GetOrCreateConversation returns the conversation's turn-cap state, creating
// it on first sight of an ID.
func (s *Store) GetOrCreateConversation(id, clientIP string) (*ConversationState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.state.Conversations[id]
	if !ok {
		c = &ConversationState{ID: id, ClientIP: clientIP, CreatedAt: time.Now().UTC()}
		s.state.Conversations[id] = c
		if err := s.saveLocked(); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// IncrementTurn records one more turn on the conversation and persists
// immediately — the turn cap must survive a restart mid-conversation just
// like the spend ceiling does.
func (s *Store) IncrementTurn(id string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.state.Conversations[id]
	if !ok {
		return 0, os.ErrNotExist
	}
	c.TurnCount++
	c.LastActivity = time.Now().UTC()
	if err := s.saveLocked(); err != nil {
		return 0, err
	}
	return c.TurnCount, nil
}

// TodaySpendUSD returns the running total for the current UTC date.
func (s *Store) TodaySpendUSD() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state.DailySpendUSD[today()]
}

// AddSpendUSD adds to today's ledger and returns the new total. Called AFTER
// a real API call completes, with its ACTUAL cost — never an estimate — so
// the ledger never drifts from what was actually spent.
func (s *Store) AddSpendUSD(amount float64) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := today()
	s.state.DailySpendUSD[k] += amount
	if err := s.saveLocked(); err != nil {
		return 0, err
	}
	return s.state.DailySpendUSD[k], nil
}

func appendJSONLine(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	if _, err := w.WriteString("\n"); err != nil {
		return err
	}
	return w.Flush()
}

func (s *Store) LogRequest(e RequestLogEntry) error {
	return appendJSONLine(s.requestLogPath, e)
}

func (s *Store) LogTranscript(e TranscriptEntry) error {
	return appendJSONLine(s.transcriptPath, e)
}
