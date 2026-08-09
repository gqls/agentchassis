package main

// claude.go — the Anthropic call, adapted from idea.uk/golang_files/engine.go's
// callClaudeOpts (raw net/http, stdlib only — no SDK dependency, matching the
// "stdlib-first like site-engine" brief). Trimmed to what an intake chat
// needs: no tools, no thinking (Haiku 4.5 is intake, not the product — PLAN
// §4 — and doesn't support adaptive thinking or a manual budget without one
// explicitly requested; omitting the field entirely is correct and simplest).

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

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResult struct {
	Text         string
	InputTokens  int
	OutputTokens int
	StopReason   string
}

// callClaude sends the system prompt + conversation history and returns the
// assistant's reply. It does not persist anything — chat.go owns logging and
// spend accounting, so this function's only job is the wire call.
func callClaude(system string, messages []claudeMessage) (claudeResult, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return claudeResult{}, fmt.Errorf("ANTHROPIC_API_KEY not set")
	}
	wireMessages := make([]map[string]any, len(messages))
	for i, m := range messages {
		wireMessages[i] = map[string]any{"role": m.Role, "content": m.Content}
	}
	body := map[string]any{
		"model":      claudeModel,
		"max_tokens": 1024, // intake replies are short; see RUNBOOK if this needs raising
		"system":     system,
		"messages":   wireMessages,
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
			Type string `json:"type"`
			Text string `json:"text"`
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
		if c.Type == "text" {
			result.Text += c.Text
		}
	}
	if result.Text == "" {
		return result, fmt.Errorf("anthropic: empty response text (stop_reason=%s)", out.StopReason)
	}
	return result, nil
}
