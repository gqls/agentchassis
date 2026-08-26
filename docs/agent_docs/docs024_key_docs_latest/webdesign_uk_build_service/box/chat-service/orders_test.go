package main

// orders_test.go — the order-intake connection, proved at the three seams
// that matter: the store's invariants (unique references, idempotent ack,
// the per-conversation cap), the collection endpoints' refusals (unset token,
// wrong token, through-the-tunnel traffic), and the tool round end to end
// (a submit_brief call stores an order, the model gets the reference in its
// tool_result, and a broken follow-up still hands the visitor their
// reference). Same mutation-test bar as the rest of this service: each
// guard here was proven to FAIL when its guard was removed, not just to pass.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func newTestOrderStore(t *testing.T) (*OrderStore, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "orders.json")
	s, err := NewOrderStore(path)
	if err != nil {
		t.Fatalf("NewOrderStore: %v", err)
	}
	return s, path
}

const validBrief = "A site for a five-a-side league in Leeds: fixtures, tables, and how to join. Plain and friendly."

func TestSubmitMintsUniqueReferencesAndSurvivesReload(t *testing.T) {
	s, path := newTestOrderStore(t)
	a, err := s.Submit("conv-1", "198.51.100.9", "a@example.com", "A", "", validBrief)
	if err != nil {
		t.Fatalf("submit a: %v", err)
	}
	b, err := s.Submit("conv-2", "198.51.100.9", "b@example.com", "B", "leeds5s.uk", validBrief)
	if err != nil {
		t.Fatalf("submit b: %v", err)
	}
	for _, o := range []BriefOrder{a, b} {
		if !strings.HasPrefix(o.Reference, "BR-") || len(o.Reference) != 9 {
			t.Errorf("reference %q: want BR- prefix and 9 chars", o.Reference)
		}
	}
	if a.Reference == b.Reference {
		t.Fatalf("two submissions minted the same reference %q", a.Reference)
	}

	// A reference the visitor was told to keep must survive a process restart.
	reloaded, err := NewOrderStore(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	got := reloaded.ListUncollected()
	if len(got) != 2 {
		t.Fatalf("after reload: %d uncollected, want 2", len(got))
	}
	// Oldest first — the collector processes in arrival order.
	if got[0].Reference != a.Reference {
		t.Errorf("list order: got %q first, want the earlier submission %q", got[0].Reference, a.Reference)
	}
	if got[1].Domain != "leeds5s.uk" || got[1].ContactEmail != "b@example.com" {
		t.Errorf("reloaded fields wrong: %+v", got[1])
	}
}

func TestAckIsIdempotentAndKeepsTheOriginalTimestamp(t *testing.T) {
	s, _ := newTestOrderStore(t)
	o, err := s.Submit("conv-1", "ip", "a@example.com", "", "", validBrief)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := s.Ack([]string{o.Reference, "BR-UNKNOWN"})
	if err != nil || changed != 1 {
		t.Fatalf("first ack: changed=%d err=%v, want 1, nil (unknown refs skipped, not errors)", changed, err)
	}
	if left := s.ListUncollected(); len(left) != 0 {
		t.Fatalf("still uncollected after ack: %d", len(left))
	}
	firstAt := *s.orders[o.Reference].CollectedAt

	// The retry after a lost response: nothing changes, history keeps its date.
	changed, err = s.Ack([]string{o.Reference})
	if err != nil || changed != 0 {
		t.Fatalf("second ack: changed=%d err=%v, want 0, nil", changed, err)
	}
	if !s.orders[o.Reference].CollectedAt.Equal(firstAt) {
		t.Error("re-ack rewrote CollectedAt — a retried ack must not rewrite history")
	}
}

func TestPerConversationSubmissionCap(t *testing.T) {
	s, _ := newTestOrderStore(t)
	for i := 0; i < maxBriefsPerConversation; i++ {
		if _, err := s.Submit("conv-1", "ip", "a@example.com", "", "", validBrief); err != nil {
			t.Fatalf("submit %d: %v", i+1, err)
		}
	}
	if _, err := s.Submit("conv-1", "ip", "a@example.com", "", "", validBrief); err == nil {
		t.Fatal("submission past the cap was accepted")
	}
	// A DIFFERENT conversation is not caught by conv-1's cap.
	if _, err := s.Submit("conv-2", "ip", "b@example.com", "", "", validBrief); err != nil {
		t.Fatalf("another conversation blocked by the wrong cap: %v", err)
	}
}

func TestValidateSubmissionRejectsWhatTheBuildCannotUse(t *testing.T) {
	long := strings.Repeat("x", maxBriefLen+1)
	if msg := ValidateSubmission("a@example.com", "", "", validBrief); msg != "" {
		t.Errorf("valid minimal rejected: %s", msg)
	}
	if msg := ValidateSubmission("a@example.com", "Leeds 5s", "leeds5s.uk", validBrief); msg != "" {
		t.Errorf("valid full rejected: %s", msg)
	}
	for name, args := range map[string][4]string{
		"missing email":           {"", "", "", validBrief},
		"junk email":              {"not-an-email", "", "", validBrief},
		"email with display name": {"A B <a@example.com>", "", "", validBrief},
		"junk domain":             {"a@example.com", "", "not a domain", validBrief},
		"brief too short":         {"a@example.com", "", "", "make me a site"},
		"brief too long":          {"a@example.com", "", "", long},
	} {
		if msg := ValidateSubmission(args[0], args[1], args[2], args[3]); msg == "" {
			t.Errorf("%s: accepted, want a rejection message", name)
		}
	}
}

// --- collection endpoints ---

func newTestOrdersAPI(t *testing.T, token string) (*ordersAPI, *OrderStore) {
	t.Helper()
	s, _ := newTestOrderStore(t)
	return &ordersAPI{store: s, token: token}, s
}

func internalReq(t *testing.T, api *ordersAPI, method, path, bearer, cfIP, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if bearer != "" {
		r.Header.Set("Authorization", "Bearer "+bearer)
	}
	if cfIP != "" {
		r.Header.Set("CF-Connecting-IP", cfIP)
	}
	w := httptest.NewRecorder()
	switch path {
	case "/internal/orders":
		api.handleList(w, r)
	case "/internal/orders/ack":
		api.handleAck(w, r)
	default:
		t.Fatalf("unrouted test path %s", path)
	}
	return w
}

func TestCollectionEndpointsRefuseCorrectly(t *testing.T) {
	// No token configured: OFF means 503, valid-looking bearer or not.
	off, _ := newTestOrdersAPI(t, "")
	if w := internalReq(t, off, http.MethodGet, "/internal/orders", "anything", "", ""); w.Code != http.StatusServiceUnavailable {
		t.Errorf("unset token: status %d, want 503", w.Code)
	}

	api, store := newTestOrdersAPI(t, "right-token")
	if _, err := store.Submit("conv-1", "ip", "a@example.com", "", "", validBrief); err != nil {
		t.Fatal(err)
	}

	if w := internalReq(t, api, http.MethodGet, "/internal/orders", "wrong-token", "", ""); w.Code != http.StatusUnauthorized {
		t.Errorf("wrong token: status %d, want 401", w.Code)
	}
	if w := internalReq(t, api, http.MethodGet, "/internal/orders", "", "", ""); w.Code != http.StatusUnauthorized {
		t.Errorf("no auth header: status %d, want 401", w.Code)
	}
	// The tunnel-header refusal: a member of the public who found a route in
	// is refused even with the right token (a token on a public route is a
	// token to rotate, not honour).
	if w := internalReq(t, api, http.MethodGet, "/internal/orders", "right-token", "203.0.113.5", ""); w.Code != http.StatusForbidden {
		t.Errorf("tunnel traffic: status %d, want 403", w.Code)
	}

	w := internalReq(t, api, http.MethodGet, "/internal/orders", "right-token", "", "")
	if w.Code != http.StatusOK {
		t.Fatalf("authorized list: status %d, want 200", w.Code)
	}
	var listResp struct {
		Orders []BriefOrder `json:"orders"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("list response not JSON: %v", err)
	}
	if len(listResp.Orders) != 1 || listResp.Orders[0].ContactEmail != "a@example.com" {
		t.Fatalf("list content wrong: %+v", listResp.Orders)
	}

	ackBody := `{"references":["` + listResp.Orders[0].Reference + `"]}`
	w = internalReq(t, api, http.MethodPost, "/internal/orders/ack", "right-token", "", ackBody)
	if w.Code != http.StatusOK {
		t.Fatalf("ack: status %d, want 200", w.Code)
	}
	var ackResp struct {
		Collected int `json:"collected"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &ackResp); err != nil || ackResp.Collected != 1 {
		t.Fatalf("ack response wrong: %s (err %v)", w.Body.String(), err)
	}
	if left := store.ListUncollected(); len(left) != 0 {
		t.Fatalf("order still uncollected after acknowledged ack")
	}
}

// --- the tool round through handleChat ---

func newToolTestChatServer(t *testing.T) (*chatServer, *OrderStore) {
	t.Helper()
	cs := newTestChatServer(t, 20, 1000.00)
	orders, _ := newTestOrderStore(t)
	cs.orders = orders
	return cs, orders
}

func submitBriefToolUse(id string, input map[string]any) claudeToolUse {
	raw, _ := json.Marshal(input)
	return claudeToolUse{ID: id, Name: "submit_brief", Input: raw}
}

func TestToolRoundStoresTheOrderAndRelaysTheReference(t *testing.T) {
	origCaller := claudeCaller
	defer func() { claudeCaller = origCaller }()

	var calls [][]claudeMessage
	claudeCaller = func(system string, messages []claudeMessage, tools []claudeTool) (claudeResult, error) {
		cp := append([]claudeMessage(nil), messages...)
		calls = append(calls, cp)
		if len(calls) == 1 {
			if len(tools) == 0 {
				t.Error("first call carried no tools — the model can never submit")
			}
			return claudeResult{
				Text: "Submitting that for you now.",
				ToolUses: []claudeToolUse{submitBriefToolUse("tu_1", map[string]any{
					"contact_email": "vis@example.com",
					"contact_name":  "Leeds 5s",
					"domain":        "leeds5s.uk",
					"brief":         validBrief,
				})},
				InputTokens: 10, OutputTokens: 10, StopReason: "tool_use",
			}, nil
		}
		// Follow-up: the tool_result must be on the wire and must carry the
		// reference the model is being asked to relay.
		last := messages[len(messages)-1]
		raw, _ := json.Marshal(last.Blocks)
		if last.Role != "user" || !strings.Contains(string(raw), "tool_result") || !strings.Contains(string(raw), "BR-") {
			t.Errorf("follow-up's last message is not a tool_result carrying the reference: %s", raw)
		}
		return claudeResult{Text: "All done. Your reference is in the result.", InputTokens: 10, OutputTokens: 10, StopReason: "end_turn"}, nil
	}

	cs, orders := newToolTestChatServer(t)
	w := postChat(t, cs, "conv-tool", "yes, submit it")

	if len(calls) != 2 {
		t.Fatalf("claudeCaller invoked %d times, want 2 (the tool round)", len(calls))
	}
	stored := orders.ListUncollected()
	if len(stored) != 1 {
		t.Fatalf("stored %d orders, want 1", len(stored))
	}
	o := stored[0]
	if o.ContactEmail != "vis@example.com" || o.Domain != "leeds5s.uk" || o.ConversationID != "conv-tool" || o.Brief != validBrief {
		t.Errorf("stored order fields wrong: %+v", o)
	}
	var resp chatResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Reply != "All done. Your reference is in the result." {
		t.Errorf("reply = %q, want the follow-up call's text", resp.Reply)
	}
	// Both calls hit the spend ledger — the ceiling must see the whole round.
	if got := cs.store.TodaySpendUSD(); got < 2*costUSD(10, 10)-1e-12 {
		t.Errorf("ledger %v, want both calls' cost (%v)", got, 2*costUSD(10, 10))
	}
}

func TestToolRoundValidationErrorStoresNothingAndFeedsTheModel(t *testing.T) {
	origCaller := claudeCaller
	defer func() { claudeCaller = origCaller }()

	callN := 0
	claudeCaller = func(system string, messages []claudeMessage, tools []claudeTool) (claudeResult, error) {
		callN++
		if callN == 1 {
			return claudeResult{
				ToolUses: []claudeToolUse{submitBriefToolUse("tu_1", map[string]any{
					"contact_email": "not-an-email",
					"brief":         validBrief,
				})},
				InputTokens: 1, OutputTokens: 1, StopReason: "tool_use",
			}, nil
		}
		raw, _ := json.Marshal(messages[len(messages)-1].Blocks)
		if !strings.Contains(string(raw), `"is_error":true`) {
			t.Errorf("validation failure did not reach the model as is_error: %s", raw)
		}
		return claudeResult{Text: "Could I take your email address?", InputTokens: 1, OutputTokens: 1, StopReason: "end_turn"}, nil
	}

	cs, orders := newToolTestChatServer(t)
	w := postChat(t, cs, "conv-badmail", "submit it")
	if got := len(orders.ListUncollected()); got != 0 {
		t.Fatalf("a rejected submission stored %d orders, want 0", got)
	}
	var resp chatResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Reply != "Could I take your email address?" {
		t.Errorf("reply = %q, want the model's corrective question", resp.Reply)
	}
}

// THE UNRECOVERABLE-OUTCOME GUARD: once a submission is stored, its reference
// must reach the visitor even when the follow-up call dies or misbehaves —
// a minted reference that never reaches them is a paid-order join that can
// never happen.
func TestReferenceSurvivesAFailedFollowUp(t *testing.T) {
	origCaller := claudeCaller
	defer func() { claudeCaller = origCaller }()

	callN := 0
	claudeCaller = func(system string, messages []claudeMessage, tools []claudeTool) (claudeResult, error) {
		callN++
		if callN == 1 {
			return claudeResult{
				ToolUses: []claudeToolUse{submitBriefToolUse("tu_1", map[string]any{
					"contact_email": "vis@example.com",
					"brief":         validBrief,
				})},
				InputTokens: 1, OutputTokens: 1, StopReason: "tool_use",
			}, nil
		}
		// The follow-up tries to call tools AGAIN — the round must terminate
		// with the server-built confirmation, not loop and not execute it.
		return claudeResult{
			ToolUses: []claudeToolUse{submitBriefToolUse("tu_2", map[string]any{
				"contact_email": "vis@example.com",
				"brief":         validBrief,
			})},
			InputTokens: 1, OutputTokens: 1, StopReason: "tool_use",
		}, nil
	}

	cs, orders := newToolTestChatServer(t)
	w := postChat(t, cs, "conv-fallback", "submit it")

	stored := orders.ListUncollected()
	if len(stored) != 1 {
		t.Fatalf("stored %d orders, want exactly 1 — the second call's tool use must NOT execute", len(stored))
	}
	var resp chatResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if !strings.Contains(resp.Reply, stored[0].Reference) {
		t.Errorf("reply %q does not carry the reference %s — the visitor lost their join key", resp.Reply, stored[0].Reference)
	}
}
