package main

// chat_test.go — proves the turn cap and spend ceiling actually STOP a call
// reaching Claude, not just that the code path exists. Both tests set
// ANTHROPIC_API_KEY to a value that will fail loudly if callClaude is ever
// actually invoked (no network in this test run, so any real HTTP attempt
// errors — but the assertion is on TURN COUNT / SPEND staying unchanged,
// which is the thing that would be wrong if the gate were a no-op).

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestChatServer(t *testing.T, maxTurns int, dailyCeiling float64) *chatServer {
	t.Helper()
	dir := t.TempDir()
	store, err := NewStore(
		filepath.Join(dir, "state.json"),
		filepath.Join(dir, "requests.jsonl"),
		filepath.Join(dir, "transcripts.jsonl"),
	)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return &chatServer{
		store:           store,
		ipLimiter:       newChatIPLimiter(),
		maxTurns:        maxTurns,
		dailyCeilingUSD: dailyCeiling,
		contactLine:     "contact us at test@example.com",
	}
}

func postChat(t *testing.T, cs *chatServer, convID, message string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(chatRequest{ConversationID: convID, Message: message})
	r := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(string(body)))
	r.Header.Set("CF-Connecting-IP", "198.51.100.9")
	w := httptest.NewRecorder()
	cs.handleChat(w, r)
	return w
}

// TestTurnCapStopsAtLimit is the mutation-proof version: a turn cap of 1
// means the FIRST call is the only one that can reach the gate as "allowed"
// (it will then fail on the network, which is fine — we're proving the gate,
// not the network call). The SECOND call on the same conversation must be
// intercepted by the cap and never increment the turn count further.
func TestTurnCapStopsAtLimit(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "test-key-no-network")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	cs := newTestChatServer(t, 1, 1000.00) // turn cap 1, spend ceiling irrelevant here
	convID := "conv-turncap-test"

	// First call: turn cap not yet hit (TurnCount starts at 0, cap is 1), so
	// the gate lets it through to callClaude — which fails on no network, so
	// the response is the contactLine via the callErr path, and TurnCount
	// becomes 1 via IncrementTurn.
	postChat(t, cs, convID, "hello")

	conv, err := cs.store.GetOrCreateConversation(convID, "198.51.100.9")
	if err != nil {
		t.Fatalf("GetOrCreateConversation: %v", err)
	}
	if conv.TurnCount != 1 {
		t.Fatalf("after 1st call: TurnCount = %d, want 1", conv.TurnCount)
	}

	// Second call: TurnCount(1) >= maxTurns(1) — gate 2 must intercept BEFORE
	// IncrementTurn runs again. If the gate were a no-op, TurnCount would
	// become 2 here.
	w := postChat(t, cs, convID, "are you there")
	conv, _ = cs.store.GetOrCreateConversation(convID, "198.51.100.9")
	if conv.TurnCount != 1 {
		t.Fatalf("after 2nd call: TurnCount = %d, want STILL 1 — the turn cap did not stop the call", conv.TurnCount)
	}
	var resp chatResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response not JSON: %v", err)
	}
	if resp.Reply != cs.contactLine {
		t.Fatalf("2nd call reply = %q, want the fail-closed contact line", resp.Reply)
	}
}

// TestSpendCeilingStopsNewCalls proves gate 3 in isolation: pre-load the
// ledger to (at) the ceiling, then a request must be intercepted BEFORE
// IncrementTurn runs — proving the ceiling check happens before any new
// turn is counted, not just before the Claude call.
func TestSpendCeilingStopsNewCalls(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "test-key-no-network")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	cs := newTestChatServer(t, 20, 0.01) // ceiling almost zero
	if _, err := cs.store.AddSpendUSD(0.01); err != nil {
		t.Fatalf("AddSpendUSD: %v", err)
	}

	convID := "conv-spendcap-test"
	w := postChat(t, cs, convID, "hello")

	conv, err := cs.store.GetOrCreateConversation(convID, "198.51.100.9")
	if err != nil {
		t.Fatalf("GetOrCreateConversation: %v", err)
	}
	if conv.TurnCount != 0 {
		t.Fatalf("TurnCount = %d, want 0 — the spend ceiling did not stop the call before it was counted as a turn", conv.TurnCount)
	}
	var resp chatResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response not JSON: %v", err)
	}
	if resp.Reply != cs.contactLine {
		t.Fatalf("reply = %q, want the fail-closed contact line", resp.Reply)
	}
}

// TestConversationHistoryThreadsAcrossTurns catches the bug found before
// Phase 5 shipped: handleChat used to call claudeCaller with ONLY the current
// message, never the conversation's prior turns, even though callClaude's own
// doc comment claims it sends "conversation history". A fake claudeCaller
// records exactly what it was sent on each call, proving turn 2 carries turn
// 1's exchange rather than starting fresh.
func TestConversationHistoryThreadsAcrossTurns(t *testing.T) {
	origCaller := claudeCaller
	defer func() { claudeCaller = origCaller }()

	var seenMessages [][]claudeMessage
	claudeCaller = func(system string, messages []claudeMessage) (claudeResult, error) {
		cp := append([]claudeMessage(nil), messages...)
		seenMessages = append(seenMessages, cp)
		return claudeResult{
			Text:        fmt.Sprintf("reply %d", len(seenMessages)),
			InputTokens: 1, OutputTokens: 1, StopReason: "end_turn",
		}, nil
	}

	cs := newTestChatServer(t, 20, 1000.00)
	convID := "conv-history-test"

	postChat(t, cs, convID, "I run a bakery")
	postChat(t, cs, convID, "what do you need from me")

	if len(seenMessages) != 2 {
		t.Fatalf("claudeCaller invoked %d times, want 2", len(seenMessages))
	}
	if len(seenMessages[0]) != 1 || seenMessages[0][0].Content != "I run a bakery" {
		t.Fatalf("first call messages = %+v, want a single message 'I run a bakery'", seenMessages[0])
	}

	// The bug this test exists to catch: the second call must include turn
	// 1's user message and assistant reply, not just turn 2's new message.
	second := seenMessages[1]
	if len(second) != 3 {
		t.Fatalf("second call sent %d messages, want 3 (turn1 user, turn1 assistant, turn2 user) — history was not threaded through", len(second))
	}
	if second[0].Role != "user" || second[0].Content != "I run a bakery" {
		t.Fatalf("second call message[0] = %+v, want turn 1's user message", second[0])
	}
	if second[1].Role != "assistant" || second[1].Content != "reply 1" {
		t.Fatalf("second call message[1] = %+v, want turn 1's assistant reply", second[1])
	}
	if second[2].Role != "user" || second[2].Content != "what do you need from me" {
		t.Fatalf("second call message[2] = %+v, want turn 2's user message", second[2])
	}
}

// TestNoCFConnectingIPIsRefused proves the clientIP()=="" path in the
// handler actually refuses, not just that clientIP() itself returns "".
func TestNoCFConnectingIPIsRefused(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "test-key-no-network")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	cs := newTestChatServer(t, 20, 1000.00)
	body, _ := json.Marshal(chatRequest{Message: "hello"})
	r := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(string(body)))
	// deliberately NOT setting CF-Connecting-IP
	w := httptest.NewRecorder()
	cs.handleChat(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (forbidden) for a request with no CF-Connecting-IP", w.Code, http.StatusForbidden)
	}
}
