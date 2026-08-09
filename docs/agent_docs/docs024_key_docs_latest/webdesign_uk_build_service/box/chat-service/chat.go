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
	"fmt"
	"log"
	"net/http"
	"time"
)

// systemPromptFacts are the ONLY numbers and named commitments this bot may
// state — copied verbatim from evidence_base at the time of writing (NOTES
// 2026-08-09). This is the SAME split that cost three wasted rewrite rounds
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
- The customer sees the finished site on a private preview link, has 14 days from receiving that link to accept it, ask for changes, or take a full refund.
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
	ipLimiter       *rateLimiter
	maxTurns        int
	dailyCeilingUSD float64
	contactLine     string // pre-built fail-closed message, checked non-empty at startup
}

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

	var req chatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8*1024)).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(req.Message) == 0 || len(req.Message) > 2000 {
		http.Error(w, "message must be 1-2000 characters", http.StatusBadRequest)
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
	// estimate would itself be a guess. The bound on overshoot is one call's
	// worth (max_tokens=1024 output, ≤$0.005 at Haiku rates) — negligible
	// against the ceiling, and simpler and more honest than estimating.
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

	start := time.Now()
	result, callErr := callClaude(systemPromptFacts, []claudeMessage{
		{Role: "user", Content: req.Message},
	})
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
		log.Printf("claude call failed (conversation=%s turn=%d): %v", convID, turnCount, callErr)
		writeChatJSON(w, convID, cs.contactLine)
		return
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
	if err := cs.store.LogTranscript(TranscriptEntry{
		Timestamp: time.Now().UTC(), ConversationID: convID, ClientIP: ip,
		Role: "assistant", Content: result.Text,
	}); err != nil {
		log.Printf("transcript log error: %v", err)
	}

	writeChatJSON(w, convID, result.Text)
}

func writeChatJSON(w http.ResponseWriter, convID, reply string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(chatResponse{ConversationID: convID, Reply: reply})
}
