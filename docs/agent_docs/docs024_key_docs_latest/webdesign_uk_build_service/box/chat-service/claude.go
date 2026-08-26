package main

// claude.go — the Anthropic call, adapted from idea.uk/golang_files/engine.go's
// callClaudeOpts (raw net/http, stdlib only — no SDK dependency, matching the
// "stdlib-first like site-engine" brief). No thinking (Haiku 4.5 is intake,
// not the product — PLAN §4 — and doesn't support adaptive thinking or a
// manual budget without one explicitly requested; omitting the field entirely
// is correct and simplest). Tools joined 2026-08-26 for submit_brief — the
// order-intake connection; chat.go owns the single permitted tool round.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const claudeModel = "claude-haiku-4-5"

// Pricing per Haiku 4.5, $1.00 / $5.00 per MTok (input/output). Used ONLY to
// compute the actual cost of a completed call for the spend ledger — never to
// estimate a call in advance. A ceiling check against an estimate would be
// checking the wrong number; see chat.go for why the check runs on the
// ALREADY-SPENT total instead.
const (
	priceInputPerMTok  = 1.00
	priceOutputPerMTok = 5.00
)

func costUSD(inputTokens, outputTokens int) float64 {
	return float64(inputTokens)/1_000_000*priceInputPerMTok +
		float64(outputTokens)/1_000_000*priceOutputPerMTok
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

// claudeMessage is one wire turn. Content carries an ordinary text turn;
// Blocks, when non-nil, wins and carries structured content (the assistant's
// tool_use turn and our tool_result answer to it) exactly as the Messages API
// expects it. Stored history stays plain text — chat.go flattens a tool round
// to its final text, so Blocks never persists and never needs replaying.
type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Blocks  []any  `json:"-"`
}

// claudeTool is one tool definition on the wire.
type claudeTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// claudeToolUse is one tool invocation the model asked for. Input is kept raw:
// the caller owns validation, and a decode error there must produce a
// tool_result the model can read, not a dropped reply.
type claudeToolUse struct {
	ID    string
	Name  string
	Input json.RawMessage
}

type claudeResult struct {
	Text         string
	ToolUses     []claudeToolUse
	InputTokens  int
	OutputTokens int
	StopReason   string
}

// callClaude sends the system prompt + conversation history (+ tool
// definitions, when any are offered) and returns the assistant's reply. It
// does not persist anything — chat.go owns logging and spend accounting, so
// this function's only job is the wire call.
func callClaude(system string, messages []claudeMessage, tools []claudeTool) (claudeResult, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return claudeResult{}, fmt.Errorf("ANTHROPIC_API_KEY not set")
	}
	wireMessages := make([]map[string]any, len(messages))
	for i, m := range messages {
		if m.Blocks != nil {
			wireMessages[i] = map[string]any{"role": m.Role, "content": m.Blocks}
		} else {
			wireMessages[i] = map[string]any{"role": m.Role, "content": m.Content}
		}
	}
	body := map[string]any{
		"model": claudeModel,
		// 2048, raised from 1024 on 2026-08-26: a brief that keeps a visitor's
		// pasted specifics verbatim (topic lists, feeds, links) can honestly
		// exceed what 1024 held. Worst-case cost of one reply at Haiku output
		// rates is ~$0.011 — see chat.go's ceiling-overshoot bound, updated in
		// the same change.
		"max_tokens": 2048,
		"system":     system,
		"messages":   wireMessages,
	}
	if len(tools) > 0 {
		body["tools"] = tools
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return claudeResult{}, err
	}
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(raw))
	if err != nil {
		return claudeResult{}, err
	}
	req.Header.Set("x-api-key", key)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return claudeResult{}, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return claudeResult{}, err
	}
	if resp.StatusCode != 200 {
		return claudeResult{}, fmt.Errorf("anthropic %d: %s", resp.StatusCode, b)
	}

	var out struct {
		Content []struct {
			Type  string          `json:"type"`
			Text  string          `json:"text"`
			ID    string          `json:"id"`
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return claudeResult{}, err
	}

	result := claudeResult{
		InputTokens:  out.Usage.InputTokens,
		OutputTokens: out.Usage.OutputTokens,
		StopReason:   out.StopReason,
	}

	// A CUT completion is a fragment that looks like an answer, not a short
	// one — HTTP is 200 and the text parses fine. Fail loudly rather than
	// hand a visitor half a sentence (idea.uk/engine.go's own hard-earned
	// rule, bugs_open-class: a truncation that looks complete).
	if out.StopReason == "max_tokens" {
		return result, fmt.Errorf("anthropic: response CUT at max_tokens (output=%d tokens) — raise max_tokens or shorten the system prompt", out.Usage.OutputTokens)
	}
	// A refusal is a successful HTTP 200 with no usable content.
	if out.StopReason == "refusal" {
		return result, fmt.Errorf("anthropic: request declined by safety classifiers (stop_reason=refusal)")
	}

	for _, c := range out.Content {
		switch c.Type {
		case "text":
			result.Text += c.Text
		case "tool_use":
			result.ToolUses = append(result.ToolUses, claudeToolUse{ID: c.ID, Name: c.Name, Input: c.Input})
		}
	}
	// A tool_use turn legitimately carries no text; only a reply with NEITHER
	// text NOR a tool call is the empty-response failure.
	if result.Text == "" && len(result.ToolUses) == 0 {
		return result, fmt.Errorf("anthropic: empty response text (stop_reason=%s)", out.StopReason)
	}
	return result, nil
}
