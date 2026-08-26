package main

// chat.go — POST /api/chat. The five PLAN §5.1 controls, in the order they
// apply on every request:
//
//  1. per-IP limit (clientIP + chatIPLimiter)      — ratelimit.go
//  2. hard turn cap per conversation                — checkTurnCap below
//  3. per-day spend ceiling, fails closed to contact — checkSpendCeiling below
//  4. request log with tokens + cost per call        — store.LogRequest
//  5. transcripts stored as structured rows           — store.LogTranscript
//
// Order matters: cheapest/least-trusting checks first. A request that fails
// the IP limiter or the turn cap never reaches the spend check, and nothing
// reaches Claude until all three gates pass.

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

// chatTools is offered on every call. The model is told (promptConduct) to
// use submit_brief only after the visitor has approved the brief and given an
// email address; the tool_result is written for the model and carries the
// minted order reference for it to relay. One tool round per HTTP request —
// completeToolRound never executes a second call's tool uses.
var chatTools = []claudeTool{{
	Name: "submit_brief",
	Description: "Submit the visitor's approved website brief to webdesign.uk. " +
		"Use only after the visitor has seen the final brief text and clearly agreed to submit it, " +
		"and has given an email address to be reached on. " +
		"The result carries the order reference to hand back to the visitor.",
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"contact_email": map[string]any{"type": "string", "description": "The visitor's email address, exactly as they gave it."},
			"contact_name":  map[string]any{"type": "string", "description": "Their name, or the business or project name, if they gave one."},
			"domain":        map[string]any{"type": "string", "description": "The domain they want the site on, if decided. Omit otherwise."},
			"brief":         map[string]any{"type": "string", "description": "The full brief text the visitor approved, verbatim."},
		},
		"required": []string{"contact_email", "brief"},
	},
}}

// systemPromptFacts are the ONLY numbers and named commitments this bot may
// state — copied verbatim from evidence_base at the time of writing (NOTES
// 2026-08-09, synced 2026-08-10 for the £75 deposit). This is the SAME split
// that cost three wasted rewrite rounds
// on the site's own copy (LANDMINES: "evidence_base.facts[] is bookkeeping,
// writer_block is the wire") — there is no code link between this string and
// the database evidence_base row, so if the owner ever changes the price,
// the terms, or the contact details there, THIS FILE MUST BE UPDATED BY HAND
// or the bot will state stale facts. Whoever changes evidence_base owns
// checking this file too.
const systemPromptFacts = `You are the intake assistant for webdesign.uk, a service that builds complete websites for small and medium UK businesses.

Facts you may state, and the ONLY facts you may state as numbers or commitments — never invent, round, or approximate anything beyond these:
- Price: £1,200 total, paid once. The owner is not VAT registered, so no VAT is added.
- Typical turnaround: three to four days from having what we need from the customer.
- The customer sees the finished site on a private preview link, has 14 days from receiving that link to accept it, ask for changes, or decline it for a refund of the price paid, minus a £75 non-refundable deposit.
- Two rounds of revisions are included in the price; further rounds are charged as work.
- Once accepted (or once the 14 days run out, whichever happens first), the site is theirs and further changes are charged as work.
- Contact for anything you cannot handle: webdesign@contactforsales.com or +44 (0) 7934 524 911.

Your job: have a short, plain conversation. Ask what business the visitor runs and what domain they'd want the site on. Do not ask for anything else unless they offer it. Do not invent services, features, or numbers beyond the facts above. Do not promise anything about timing, price, or process that isn't stated above. If asked something you don't know, say so plainly and point at the contact details. Write in plain, direct British English — short sentences, no agency-marketing language, no em dashes. This is a first conversation, not a sales pitch: restraint reads as confidence here.`

type chatRequest struct {
	ConversationID string `json:"conversation_id"`
	Message        string `json:"message"`
}

type chatResponse struct {
	ConversationID string `json:"conversation_id"`
	Reply          string `json:"reply"`
}

type chatServer struct {
	store           *Store
	orders          *OrderStore // nil only in tests that never reach a tool round
	ipLimiter       *rateLimiter
	maxTurns        int
	dailyCeilingUSD float64
	contactLine     string // pre-built fail-closed message, checked non-empty at startup

	// systemPrompt supplies the prompt per call. Defaults to the compiled-in
	// systemPromptFacts (legacy mode, FACTS_URL unset); main.go swaps in a
	// factsProvider's live-rendered prompt when the operator opts in. A
	// function rather than a string so a background facts refresh reaches the
	// NEXT call without any coordination here.
	systemPrompt func() string
}

// claudeCaller indirects the Anthropic call so tests can substitute a fake
// and prove multi-turn history is actually threaded through — mirrors the
// mutation-test rigor already applied to the turn cap / spend ceiling gates.
var claudeCaller = callClaude

func newChatID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (cs *chatServer) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ip := clientIP(r)
	if ip == "" {
		// No CF-Connecting-IP means this request did not come through the
		// tunnel — refuse rather than fall back to a spoofable key.
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Gate 1: per-IP limit.
	if ok, retryAfter := cs.ipLimiter.allow(ip); !ok {
		w.Header().Set("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
		http.Error(w, "too many requests, try again later", http.StatusTooManyRequests)
		return
	}

	// 16KB body / 5000-char message, raised from 8KB/2000 on 2026-08-26: the
	// conduct now invites a visitor to PASTE a prepared description and have
	// it taken as the brief, and a real one does not fit 2000 characters.
	var req chatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(req.Message) == 0 || len(req.Message) > 5000 {
		http.Error(w, "message must be 1-5000 characters", http.StatusBadRequest)
		return
	}

	convID := req.ConversationID
	if convID == "" {
		convID = newChatID()
	}
	conv, err := cs.store.GetOrCreateConversation(convID, ip)
	if err != nil {
		log.Printf("store error (get conversation): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Gate 2: hard turn cap per conversation.
	if conv.TurnCount >= cs.maxTurns {
		writeChatJSON(w, convID, cs.contactLine)
		return
	}

	// Gate 3: daily spend ceiling, fails closed to contact details. Checked
	// against the ALREADY-SPENT total, not an estimate of this call's cost —
	// output tokens aren't known until the call completes, so any pre-call
	// estimate would itself be a guess. The bound on overshoot is one REQUEST's
	// worth — up to two calls when a tool round runs (max_tokens=2048 each,
	// ≤$0.02 total at Haiku rates) — negligible against the ceiling, and
	// simpler and more honest than estimating.
	if cs.store.TodaySpendUSD() >= cs.dailyCeilingUSD {
		writeChatJSON(w, convID, cs.contactLine)
		return
	}

	turnCount, err := cs.store.IncrementTurn(convID)
	if err != nil {
		log.Printf("store error (increment turn): %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := cs.store.LogTranscript(TranscriptEntry{
		Timestamp: time.Now().UTC(), ConversationID: convID, ClientIP: ip,
		Role: "user", Content: req.Message,
	}); err != nil {
		log.Printf("transcript log error: %v", err)
	}

	// Build the wire history from what this conversation has said so far, then
	// this turn's message — callClaude/the Messages API is stateless per call,
	// so without this every reply would be generated with no memory of
	// anything the visitor already said (caught before shipping Phase 5).
	wireMessages := make([]claudeMessage, 0, len(conv.Messages)+1)
	for _, m := range conv.Messages {
		wireMessages = append(wireMessages, claudeMessage{Role: m.Role, Content: m.Content})
	}
	wireMessages = append(wireMessages, claudeMessage{Role: "user", Content: req.Message})

	// cs.systemPrompt is set at construction: the compiled-in facts in legacy
	// mode, a live-rendered prompt when FACTS_URL is configured. Never nil —
	// main.go and tests both set it; a nil here is a construction bug and a
	// panic is the honest report of one.
	result, callErr := cs.callAndAccount(convID, ip, wireMessages)
	if callErr != nil {
		// Keep the user's side of this turn in history even though it failed —
		// a later successful turn should still know what was asked. Do NOT
		// record the canned contactLine as an "assistant" turn: it isn't
		// something Claude said, and feeding it back would make the model
		// think it already sent that exact line.
		if err := cs.store.AppendMessages(convID, StoredMessage{Role: "user", Content: req.Message}); err != nil {
			log.Printf("store error (append messages): %v", err)
		}
		log.Printf("claude call failed (conversation=%s turn=%d): %v", convID, turnCount, callErr)
		writeChatJSON(w, convID, cs.contactLine)
		return
	}

	replyText := result.Text
	if len(result.ToolUses) > 0 {
		// The one permitted tool round: execute submit_brief, answer every
		// tool_use block, ask the model to relay the outcome. Always returns
		// something safe to show the visitor — after a stored submission the
		// reference must reach them even if the follow-up call fails.
		replyText = cs.completeToolRound(convID, ip, wireMessages, result)
	}

	if err := cs.store.LogTranscript(TranscriptEntry{
		Timestamp: time.Now().UTC(), ConversationID: convID, ClientIP: ip,
		Role: "assistant", Content: replyText,
	}); err != nil {
		log.Printf("transcript log error: %v", err)
	}
	// History stores the FLATTENED turn (user text in, final assistant text
	// out). The tool exchange itself is not replayed on later turns; the
	// final text carries the reference, which is what the conversation needs.
	if err := cs.store.AppendMessages(convID,
		StoredMessage{Role: "user", Content: req.Message},
		StoredMessage{Role: "assistant", Content: replyText},
	); err != nil {
		log.Printf("store error (append messages): %v", err)
	}

	writeChatJSON(w, convID, replyText)
}

// callAndAccount is one wire call plus its bookkeeping: request log row,
// spend ledger entry. Both calls of a tool round go through here, so the
// ledger sees the true cost of the request, not just its first half.
func (cs *chatServer) callAndAccount(convID, ip string, messages []claudeMessage) (claudeResult, error) {
	start := time.Now()
	result, callErr := claudeCaller(cs.systemPrompt(), messages, chatTools)
	latency := time.Since(start)

	logEntry := RequestLogEntry{
		Timestamp: time.Now().UTC(), ConversationID: convID, ClientIP: ip,
		Model: claudeModel, LatencyMS: latency.Milliseconds(),
	}
	if callErr != nil {
		logEntry.Error = callErr.Error()
		if err := cs.store.LogRequest(logEntry); err != nil {
			log.Printf("request log error: %v", err)
		}
		return result, callErr
	}
	cost := costUSD(result.InputTokens, result.OutputTokens)
	logEntry.InputTokens = result.InputTokens
	logEntry.OutputTokens = result.OutputTokens
	logEntry.CostUSD = cost
	logEntry.StopReason = result.StopReason
	if err := cs.store.LogRequest(logEntry); err != nil {
		log.Printf("request log error: %v", err)
	}
	if _, err := cs.store.AddSpendUSD(cost); err != nil {
		log.Printf("spend ledger error: %v", err)
	}
	return result, nil
}

// completeToolRound executes the first submit_brief call from `first`,
// answers EVERY tool_use block (the API requires a tool_result per id), and
// makes one follow-up call so the model can relay the outcome in its own
// words. Exactly one round: a follow-up that tries to call tools again, or
// fails, gets the server-built fallback instead — which, after a stored
// submission, must carry the reference, because losing a minted reference is
// the one unrecoverable outcome here.
func (cs *chatServer) completeToolRound(convID, ip string, wire []claudeMessage, first claudeResult) string {
	var submitted *BriefOrder
	toolResults := make([]any, 0, len(first.ToolUses))
	for i, tu := range first.ToolUses {
		var content string
		var isErr bool
		switch {
		case tu.Name != "submit_brief":
			content, isErr = "unknown tool "+tu.Name, true
		case i > 0:
			content, isErr = "only the first submission in a turn is handled; this one was not stored", true
		default:
			content, isErr, submitted = cs.execSubmitBrief(convID, ip, tu.Input)
		}
		toolResults = append(toolResults, map[string]any{
			"type": "tool_result", "tool_use_id": tu.ID, "content": content, "is_error": isErr,
		})
	}

	assistantBlocks := make([]any, 0, len(first.ToolUses)+1)
	if first.Text != "" {
		assistantBlocks = append(assistantBlocks, map[string]any{"type": "text", "text": first.Text})
	}
	for _, tu := range first.ToolUses {
		assistantBlocks = append(assistantBlocks, map[string]any{
			"type": "tool_use", "id": tu.ID, "name": tu.Name, "input": tu.Input,
		})
	}
	followup := append(append([]claudeMessage(nil), wire...),
		claudeMessage{Role: "assistant", Blocks: assistantBlocks},
		claudeMessage{Role: "user", Blocks: toolResults},
	)

	second, err := cs.callAndAccount(convID, ip, followup)
	if err != nil || second.Text == "" || len(second.ToolUses) > 0 {
		if err != nil {
			log.Printf("tool-round follow-up failed (conversation=%s): %v", convID, err)
		}
		if submitted != nil {
			return "Your brief is submitted. Your order reference is " + submitted.Reference +
				". Please keep it: quote it when you pay, and in any message to us, so everything matches up to your brief."
		}
		return cs.contactLine
	}
	return second.Text
}

// execSubmitBrief validates and stores one submission. The returned string is
// the tool_result content and is written FOR THE MODEL: on error it says what
// to ask the visitor for; on success it hands over the reference and says how
// to relay it.
func (cs *chatServer) execSubmitBrief(convID, ip string, input json.RawMessage) (string, bool, *BriefOrder) {
	var in struct {
		ContactEmail string `json:"contact_email"`
		ContactName  string `json:"contact_name"`
		Domain       string `json:"domain"`
		Brief        string `json:"brief"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return "the submission fields could not be read; call submit_brief again with contact_email and brief", true, nil
	}
	if msg := ValidateSubmission(in.ContactEmail, in.ContactName, in.Domain, in.Brief); msg != "" {
		return msg, true, nil
	}
	if cs.orders == nil {
		return "brief submission is not available right now; apologise and give the visitor the contact details from the facts", true, nil
	}
	order, err := cs.orders.Submit(convID, ip, in.ContactEmail, in.ContactName, in.Domain, in.Brief)
	if err != nil {
		if errors.Is(err, errTooManySubmissions) {
			return "this conversation has already submitted its maximum number of briefs; tell the visitor to use the contact details in the facts for further changes", true, nil
		}
		log.Printf("brief submit failed (conversation=%s): %v", convID, err)
		return "the submission could not be stored just now; apologise and give the visitor the contact details from the facts", true, nil
	}
	log.Printf("brief submitted: reference=%s conversation=%s", order.Reference, convID)
	return "Submitted successfully. Order reference: " + order.Reference +
		". Tell the visitor their brief is in, give them this reference exactly as written, and tell them to keep it " +
		"and quote it when they pay and in any message to us. Do not invent any other next steps.", false, &order
}

func writeChatJSON(w http.ResponseWriter, convID, reply string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(chatResponse{ConversationID: convID, Reply: reply})
}
