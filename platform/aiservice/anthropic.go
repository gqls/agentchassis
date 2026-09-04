// FILE: platform/aiservice/anthropic.go
package aiservice

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// AnthropicClient implements the AIService interface for Anthropic's Claude
type AnthropicClient struct {
	apiKey string
	model  string
	// maxTokens is the output budget applied when a caller passes none in the
	// options map. Resolved once at construction from the same `ai_service`
	// config `model` above comes from; DefaultMaxOutputTokens when unconfigured.
	// See max_tokens.go for why the client owns this rather than the caller
	// (bugs_open/257): a nil options map used to mean a silent 2048 no matter
	// what `ai_service.max_tokens` said.
	maxTokens  int
	httpClient *http.Client
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(ctx context.Context, config map[string]interface{}) (*AnthropicClient, error) {
	// Validate config exists
	if config == nil {
		return nil, fmt.Errorf("ai_service config is nil - ensure 'ai_service' block exists in step config")
	}

	// Safely extract api_key_env_var with helpful error message
	apiKeyEnvVarRaw, exists := config["api_key_env_var"]
	if !exists || apiKeyEnvVarRaw == nil {
		return nil, fmt.Errorf("ai_service.api_key_env_var not configured - add 'api_key_env_var: \"ANTHROPIC_API_KEY\"' to ai_service config")
	}
	apiKeyEnvVar, ok := apiKeyEnvVarRaw.(string)
	if !ok || apiKeyEnvVar == "" {
		return nil, fmt.Errorf("ai_service.api_key_env_var must be a non-empty string, got: %T", apiKeyEnvVarRaw)
	}

	// Get API key from environment
	apiKey := os.Getenv(apiKeyEnvVar)
	if apiKey == "" {
		return nil, fmt.Errorf("API key environment variable '%s' is not set or empty", apiKeyEnvVar)
	}

	// Safely extract model with default fallback
	model := "claude-sonnet-4-6" // sensible default
	// model := "claude-haiku-4-5-20251001"
	if modelRaw, exists := config["model"]; exists && modelRaw != nil {
		if modelStr, ok := modelRaw.(string); ok && modelStr != "" {
			model = modelStr
		}
	}

	// Resolve alias to actual model name
	model = ResolveModelAlias(model, nil) // or pass a logger if available

	/*if model == "" {
		return nil, fmt.Errorf("no AI model selected")
	}*/
	return &AnthropicClient{
		apiKey: apiKey,
		model:  model,
		// Opt-in by construction: absent `max_tokens` yields
		// DefaultMaxOutputTokens, i.e. the pre-257 constant, so a config that
		// never named the key produces a byte-identical request.
		maxTokens: resolvedMaxTokens(config),
		// Ceiling, not a per-call bound — callers still bound calls via ctx.
		// The chassis's action path carries NO deadline end-to-end, so without
		// this a stalled connection holds the consume lane until a pod roll
		// (bugs_open/130: one call waited 30 minutes and was freed only by
		// pod shutdown). 600s = 1.66x the slowest successful call in 44k
		// logged calls; a hang now surfaces in llm_call_log as
		// "Client.Timeout exceeded" at ~600,0xx ms.
		httpClient: &http.Client{Timeout: 600 * time.Second},
	}, nil
}

// GenerateText generates text using Claude
// GenerateText sends a plain text prompt. The shared request/response path is
// generate; GenerateWithImages rides the same path with block content, so the
// hard-won response handling below (truncation, refusal, usage write-back —
// bugs_open/008/019) exists exactly once.
func (c *AnthropicClient) GenerateText(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	return c.generate(ctx, prompt, options)
}

// GenerateWithImages implements VisionCapable: images as base64 source blocks
// before the text block, one user message, same response handling.
func (c *AnthropicClient) GenerateWithImages(ctx context.Context, prompt string, images []ImageInput, options map[string]interface{}) (string, error) {
	content := make([]interface{}, 0, len(images)+1)
	for _, img := range images {
		content = append(content, map[string]interface{}{
			"type": "image",
			"source": map[string]interface{}{
				"type":       "base64",
				"media_type": img.MediaType,
				"data":       base64.StdEncoding.EncodeToString(img.Data),
			},
		})
	}
	content = append(content, map[string]interface{}{"type": "text", "text": prompt})
	return c.generate(ctx, content, options)
}

// CacheBreakpointMarker splits a rendered prompt into a cacheable shared
// prefix and a per-call suffix. A prompt containing it is sent as two text
// blocks with cache_control on the FIRST; a prompt without it is sent exactly
// as before, byte for byte.
//
// OPT-IN BY CONSTRUCTION, and deliberately so (owner ruling 2026-08-02,
// RFC_010): this client is the single seam every agent in the fleet calls, so
// the unsafe default is OFF. A caller that has never heard of caching cannot
// be affected by it — absence of the marker is not "caching disabled", it is
// the identical code path that ran yesterday.
//
// WHY A MARKER AND NOT A SEPARATE PARAMETER: the prompt is rendered to a flat
// string by execute_llm_prompt from a DB-held template long before it reaches
// this client, and nothing in between knows which half is shared. Putting the
// boundary IN the template keeps the decision where the shared/varying split
// is actually authored and visible to a reviewer, instead of requiring a
// parallel parameter that could drift out of step with the text it describes.
//
// ⚠ Anthropic caching is a PREFIX match: everything before the marker must be
// byte-identical across calls or nothing is ever read from cache. Putting a
// timestamp, a UUID or a per-seat name above the marker silently costs the
// write premium and returns zero reads — see the concept-register entry.
const CacheBreakpointMarker = "<!--CACHE_BREAKPOINT-->"

// TTL IS THE ONE-HOUR EXTENDED CACHE. Owner ruling 2026-08-15.
//
// ⚠ READ THE HISTORY BEFORE CHANGING THIS LINE — it has been flipped once already,
// and the reason it was OFF was sound at the time it was written.
//
// WHAT IT WAS, AND WHY. This constant was `""` (send no ttl field, get the
// 5-minute default) because the council gate's edit-quality seat found
// (correlation b54f173e, medium) that the extended TTL was gated behind a beta
// header this client does not send — so the first caller to opt in would get a
// 400 rather than a cache hit, and because council-gate reviews every platform
// change, a 400 there takes out the review path for the whole estate. That
// asymmetry is the whole argument and it still holds: a too-short TTL is a worse
// saving; an unsupported ttl field is an outage.
//
// The comment that stood here set the bar for changing it: "confirm the current
// beta header for extended TTL, send it, and set ttl here. Do NOT reintroduce ttl
// without that confirmation." This is that confirmation — and the answer turned
// out to be that there is no header to send.
//
// THE EVIDENCE, measured 2026-08-15 against the live account from inside a chassis
// pod (so the fleet's own key, not a personal one), on BOTH models the fleet uses:
//
//	claude-sonnet-5   → HTTP 200, and the usage block credits the write to the
//	                    1-hour bucket, which is the part that proves the TTL was
//	                    honoured rather than merely tolerated:
//	                    "cache_creation": {"ephemeral_5m_input_tokens": 0,
//	                                       "ephemeral_1h_input_tokens": 6003}
//	claude-sonnet-4-6 → HTTP 200, no 400. Stated limit: that call returned a cache
//	                    READ (0 in both creation buckets), so it confirms the field
//	                    is ACCEPTED on this model but does not by itself re-prove
//	                    the 1h bucket. The acceptance is what the outage risk hung
//	                    on; the bucket is proven above.
//
// No beta header was sent in either probe. Anthropic's current reference lists the
// 5m and 1h TTLs as generally available.
//
// WHY 1h RATHER THAN THE DEFAULT — the arithmetic, so a future reader can re-judge
// it instead of inheriting it. A cache write costs 1.25x base input at 5m and 2x at
// 1h; a read costs 0.1x. Break-even hit rate is therefore ~22% at 5m and ~53% at 1h.
// Measured over three days on same-prefix call gaps:
//
//	council-gate         96% @5m / 98% @1h — already winning; 1h is marginally better
//	diagnose-agent       64% @5m / 93% @1h — better at 1h
//	content-gap-planner   1% @5m / 99.8% @1h — the whole reason for this change:
//	                     at 5m it would cost ~24% MORE than not caching at all
//
// IF YOU ARE REVERTING THIS: set the value back to "" — do NOT delete the field
// handling below, and do not assume a 400 you are seeing comes from the ttl. Check
// `cache_creation.ephemeral_1h_input_tokens` on a live call first; if the API had
// stopped accepting the field, every call in the fleet would be failing, not one.
const cacheTTL = "1h"

func (c *AnthropicClient) generate(ctx context.Context, content interface{}, options map[string]interface{}) (string, error) {
	// Build request. content is either a plain string (GenerateText) or a
	// block array (GenerateWithImages) — the Messages API accepts both shapes
	// in the same field.
	// A client built through NewAnthropicClient always carries a positive
	// budget. A client built as a STRUCT LITERAL does not — every fake in this
	// package's own tests is one, and a zero would be sent to the API as
	// `max_tokens: 0`, which is a 400. Measured while writing this: the whole
	// aiservice suite passed with the field unset, so the existing tests cannot
	// see it. Floor it here rather than trusting construction, so the bad state
	// is unrepresentable at the wire (bugs_open/257).
	budget := c.maxTokens
	if budget <= 0 {
		budget = DefaultMaxOutputTokens
	}

	requestBody := map[string]interface{}{
		"model": c.model,
		// The CONFIGURED budget, not a literal. Still overridden by
		// options["max_tokens"] below, so the canonical ExecuteLLMPromptAction
		// path is unaffected; this is what a direct caller inherits instead of
		// a hardcoded 2048 (bugs_open/257).
		"max_tokens": budget,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": content,
			},
		},
	}

	// Split a marked prompt into [cacheable shared prefix][per-call suffix].
	// Only a plain string is eligible: the image path builds its own block
	// array and its caching story is different (an image block is cacheable
	// too, but nothing asks for that yet, and guessing would put a breakpoint
	// somewhere no caller chose).
	if promptStr, ok := content.(string); ok {
		if idx := strings.Index(promptStr, CacheBreakpointMarker); idx >= 0 {
			shared := promptStr[:idx]
			rest := promptStr[idx+len(CacheBreakpointMarker):]
			// A marker at position 0 leaves nothing to cache. Sending an empty
			// text block is a 400, and a cache_control on nothing is a silent
			// write of zero tokens — so treat it as "no marker" rather than
			// inventing a breakpoint the caller did not ask for.
			//
			// BUT THE MARKER STILL HAS TO GO. The first draft fell back to
			// sending promptStr UNCHANGED here, which left the literal
			// "<!--CACHE_BREAKPOINT-->" in the text for the model to read as
			// content — caught by the council gate's edit-quality seat
			// (correlation b54f173e). The marker is plumbing, never content:
			// whatever path we take, it must not reach the model. Stripping
			// every occurrence also covers a template that accidentally
			// contains two.
			if strings.TrimSpace(shared) == "" {
				content = strings.ReplaceAll(promptStr, CacheBreakpointMarker, "")
				requestBody["messages"] = []map[string]interface{}{
					{"role": "user", "content": content},
				}
			} else {
				cacheControl := map[string]interface{}{"type": "ephemeral"}
				// Only send ttl when one is configured. An empty string is not
				// a valid ttl and would 400 — see the cacheTTL comment for why
				// the default (no field at all) is the deliberate choice.
				if cacheTTL != "" {
					cacheControl["ttl"] = cacheTTL
				}
				blocks := []map[string]interface{}{
					{
						"type":          "text",
						"text":          shared,
						"cache_control": cacheControl,
					},
				}
				// A marker at the very END is legitimate: it means "cache all
				// of it". Appending an empty second block would 400, so omit it.
				if rest != "" {
					blocks = append(blocks, map[string]interface{}{
						"type": "text",
						"text": rest,
					})
				}
				requestBody["messages"] = []map[string]interface{}{
					{"role": "user", "content": blocks},
				}
			}
		}
	}

	// Check if extended thinking is requested
	if options != nil {
		if budgetTokens, ok := options["budget_tokens"]; ok {
			switch bt := budgetTokens.(type) {
			case float64:
				if bt > 0 {
					requestBody["thinking"] = map[string]interface{}{
						"type":          "enabled",
						"budget_tokens": int(bt),
					}
				}
			case int:
				if bt > 0 {
					requestBody["thinking"] = map[string]interface{}{
						"type":          "enabled",
						"budget_tokens": bt,
					}
				}
			}
		}
	}

	// Temperature is intentionally NOT sent to Anthropic.
	// Claude Opus 4.7+ returns a 400 for any non-default temperature, and
	// extended thinking is incompatible with temperature on any model.
	// Temperature stays in the generic options contract for other providers
	// (e.g. ollama); the Anthropic client simply ignores it.
	if options != nil {
		// Override max_tokens from provided options
		if maxTokens, ok := options["max_tokens"]; ok {
			requestBody["max_tokens"] = maxTokens
		}
		// Record what was actually sent, for llm_call_log (mirrors the
		// __usage_* write-back below). No temperature is sent, so
		// __sent_temperature is deliberately left unset.
		switch mt := requestBody["max_tokens"].(type) {
		case int:
			options["__sent_max_tokens"] = mt
		case float64:
			options["__sent_max_tokens"] = int(mt)
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response — handle both standard and extended thinking formats
	// With thinking enabled, content array has {type:"thinking"} blocks
	// followed by {type:"text"} blocks. We want the text block.
	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens int `json:"input_tokens"`
			// ⚠ input_tokens is the UNCACHED REMAINDER only, not the prompt
			// size. Once a caller uses CacheBreakpointMarker, the true prompt
			// is input + cache_creation + cache_read, and reading
			// input_tokens alone makes a cached call look ~95% smaller than
			// it is. Every one of these three must reach llm_call_log or the
			// fleet's own cost measurements quietly start understating.
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
			OutputTokens             int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Write usage tokens back to options so the caller can log them
	if options != nil {
		options["__usage_input_tokens"] = response.Usage.InputTokens
		options["__usage_output_tokens"] = response.Usage.OutputTokens
		// Always written, including the zeros an uncached call returns. A key
		// that appears only on cache hits cannot distinguish "this call did
		// not use the cache" from "this build predates cache support" — which
		// is exactly the question anyone reading the numbers will have.
		options["__usage_cache_creation_input_tokens"] = response.Usage.CacheCreationInputTokens
		options["__usage_cache_read_input_tokens"] = response.Usage.CacheReadInputTokens
	}

	// Extract the text BEFORE deciding whether this was a truncation, so a cut
	// completion can carry its partial back to the caller. Returning "" here is
	// what made bugs_open/019 unrecoverable: the platform detected the cut and
	// then destroyed the only thing that could survive it.
	// Block TYPES, not just a count. A single non-text block is the tell that
	// separates a refusal from a thinking-only response, and the old terminal
	// error reported only "had 1 blocks" — a number that distinguishes neither
	// (bugs_open/008 §REAL-CASE).
	blockTypes := make([]string, 0, len(response.Content))
	for _, block := range response.Content {
		if block.Type == "" {
			blockTypes = append(blockTypes, "(untyped)")
			continue
		}
		blockTypes = append(blockTypes, block.Type)
	}

	partial := ""
	for _, block := range response.Content {
		if block.Type == "text" || block.Type == "" {
			partial = block.Text
			break
		}
		if partial == "" && block.Text != "" {
			partial = block.Text
		}
	}

	if response.StopReason == "max_tokens" {
		// Non-nil error, so every existing `if err != nil` caller behaves exactly
		// as before. Only a caller that asks via aiservice.IsTruncated sees the
		// partial — tolerating a cut stays an opt-in decision at the step.
		return partial, &TruncatedError{
			Partial:      partial,
			OutputTokens: response.Usage.OutputTokens,
			Reason:       "stop_reason=max_tokens",
			Provider:     "anthropic",
		}
	}

	// A refusal is a DECISION, not a failure to produce: the model understood the
	// request and declined it. Decoded here so it cannot fall through to the
	// terminal fallback below, where it read as a parse fault and sent the last
	// reader looking in the wrong layer entirely (bugs_open/008 item 5).
	if response.StopReason == "refusal" {
		return "", &RefusalError{
			Reason:   "stop_reason=refusal",
			Blocks:   blockTypes,
			Provider: "anthropic",
			Model:    c.model,
		}
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no content in response (stop_reason=%q)", response.StopReason)
	}

	if partial != "" {
		return partial, nil
	}

	// Fallback: return first block with any text
	for _, block := range response.Content {
		if block.Text != "" {
			return block.Text, nil
		}
	}

	// Name the cause. "had 1 blocks" was a count that distinguished a refusal, a
	// thinking-only response and a genuine parse fault from each other not at all.
	return "", fmt.Errorf("no text content in response (stop_reason=%q, %d block(s): %s)",
		response.StopReason, len(response.Content), strings.Join(blockTypes, ", "))
}

// GenerateEmbedding generates embeddings (not implemented for Anthropic)
func (c *AnthropicClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, fmt.Errorf("embedding generation not supported by Anthropic")
}

func (c *AnthropicClient) Model() string {
	return c.model
}

func (c *AnthropicClient) Provider() string {
	return "anthropic"
}
