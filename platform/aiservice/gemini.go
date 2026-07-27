// FILE: platform/aiservice/gemini.go
//
// Gemini provider (Google's generativelanguage.googleapis.com API).
// Implements the same AIService interface as AnthropicClient/OllamaClient.
//
// Agent definitions use it like:
//
//	"ai_service": {
//	    "provider": "gemini",
//	    "model": "gemini-pro-latest",
//	    "api_key_env_var": "GEMINI_API_KEY"
//	}
//
// Optional keys, all defaulted (see the constructor):
//
//	"thinking_reserve_tokens": 8192          // output budget reserved for thinking
//	"thinking_level": "low"                  // cost lever ) at most one — Gemini
//	"thinking_budget_tokens": 512            // cost lever ) refuses both together
//	"embedding_model": "text-embedding-004"
//
// A thinking knob only makes a call cheaper; the reserve is what makes it work.
// Neither knob caps thinking (measured 2026-07-27), so do not treat one as a
// substitute for the reserve.
package aiservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const geminiAPIBase = "https://generativelanguage.googleapis.com/v1beta/models"

// defaultGeminiThinkingReserve is the output budget requested ON TOP OF the
// caller's visible-text budget whenever the model thinks by default.
//
// Gemini's maxOutputTokens is a TOTAL output ceiling, and thinking tokens are
// spent from it before any visible text is emitted. Every max_tokens value in
// this platform was sized against Anthropic, where — with extended thinking off,
// which is how every agent here runs — the whole cap is visible text. Passing
// that number straight through therefore starves the answer rather than capping
// it. Measured against this platform's own key on 2026-07-24 with
// gemini-pro-latest (then resolving to gemini-3.1-pro-preview): the 100-token
// tier produced ZERO visible text, and the 500-token tier ~85 characters, both
// with finishReason=MAX_TOKENS. It was read as a model-quality problem and the
// provider was reverted to Anthropic (commit 4dd5d6378).
//
// The reserve is a CEILING, not a purchase — Gemini bills tokens actually
// produced — so provisioning it generously costs nothing when the model thinks
// briefly, and is the difference between an answer and an empty string when it
// does not.
//
// 8192 is measured, not guessed (2026-07-27, gemini-pro-latest → gemini-3.1-pro-
// preview, against the REAL 12,570-char page-content-writer prompt): thinking
// came to 2,764 and 2,878 tokens on two runs, so the default carries ~3x
// headroom. On a trivial prompt it is 786–1,145. Two figures worth keeping:
//   - Thinking EXPANDS to fill a small ceiling. At maxOutputTokens=100 it spent
//     92 and left 4 tokens of text; at 500 it spent 477 and left 19. That is the
//     2026-07-24 failure, reproduced exactly.
//   - Neither thinking knob caps it (see the struct fields), so the reserve is
//     not made redundant by configuring one.
//
// Re-measure with scripts/gemini-probe.sh if the prompts grow materially: this
// is a property of prompt complexity, not a constant.
const defaultGeminiThinkingReserve = 8192

// geminiNonThinkingModels are the model families known NOT to spend output
// budget on thinking (flash-lite verified against this key on 2026-07-24: clean
// output at every budget tested, no thinking overhead).
//
// This is deliberately a DENY-list: anything unrecognised is assumed to think.
// An unfamiliar Gemini model name is almost always a NEWER one, and every
// generation since 2.5 thinks by default, so an allow-list here would silently
// under-provision each new model exactly the way the 2024-era caps did — and
// under-provisioning presents as bad copy, not as an error.
var geminiNonThinkingModels = []string{
	"flash-lite",
	"embedding",
}

// geminiRetiredPins are pinned snapshots Google has closed to newly-provisioned
// API keys: they answer 404 "no longer available to new users" rather than
// carrying a deprecation warning, so a config naming one fails at generate time,
// mid-orchestration, on every call. Verified against this platform's key on
// 2026-07-24 (commit 4dd5d6378). Construction fails here instead, naming a
// replacement, because a dead model pin is a config error and not a runtime one.
//
// Floating pointers (`gemini-pro-latest`, `gemini-flash-lite-latest`) are the
// safer thing to configure precisely because of this failure mode: Google gates
// pinned GENERATIONS by key provenance, so a snapshot that works today can stop
// working for a key issued tomorrow.
var geminiRetiredPins = map[string]string{
	"gemini-2.5-pro":   "gemini-pro-latest",
	"gemini-2.5-flash": "gemini-flash-lite-latest",
}

// GeminiClient implements the AIService interface for Google's Gemini
type GeminiClient struct {
	apiKey     string
	model      string
	httpClient *http.Client

	// thinkingReserve is added to the caller's max_tokens when the model thinks.
	thinkingReserve int
	// thinkingLevel / thinkingBudget carry an explicit operator override for how
	// much the model may think. Google accepts either but refuses both together
	// ("You can only set only one of thinking budget and thinking level"),
	// so at most one is set.
	//
	// Both default to empty, and NEITHER removes the need for the reserve:
	// measured on gemini-pro-latest 2026-07-27, thinkingBudget is a soft target
	// the model overshoots freely — 128 requested produced 483 thinking tokens,
	// 32768 requested produced 783. It reduces thinking substantially (2,764 →
	// ~940 on the real page-writer prompt) but bounds nothing. thinkingLevel
	// behaves the same way (2,764 → ~1,080 at "low").
	//
	// So these are a COST lever, not a correctness one. The reserve is what makes
	// the call work; a knob only makes it cheaper.
	thinkingLevel  string
	thinkingBudget *int

	embeddingModel string
}

// geminiModelThinks reports whether the model should be assumed to spend output
// budget on thinking before emitting visible text.
func geminiModelThinks(model string) bool {
	lower := strings.ToLower(model)
	for _, family := range geminiNonThinkingModels {
		if strings.Contains(lower, family) {
			return false
		}
	}
	return true
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(ctx context.Context, config map[string]interface{}) (*GeminiClient, error) {
	if config == nil {
		return nil, fmt.Errorf("ai_service config is nil - ensure 'ai_service' block exists in step config")
	}

	apiKeyEnvVarRaw, exists := config["api_key_env_var"]
	if !exists || apiKeyEnvVarRaw == nil {
		return nil, fmt.Errorf("ai_service.api_key_env_var not configured - add 'api_key_env_var: \"GEMINI_API_KEY\"' to ai_service config")
	}
	apiKeyEnvVar, ok := apiKeyEnvVarRaw.(string)
	if !ok || apiKeyEnvVar == "" {
		return nil, fmt.Errorf("ai_service.api_key_env_var must be a non-empty string, got: %T", apiKeyEnvVarRaw)
	}

	apiKey := os.Getenv(apiKeyEnvVar)
	if apiKey == "" {
		return nil, fmt.Errorf("API key environment variable '%s' is not set or empty", apiKeyEnvVar)
	}

	// No default model. The previous default was a pinned 2.5 snapshot that has
	// since been closed to this key — a default that quietly rots into a 404 on
	// every call is worse than a config error, because it reads as an outage.
	modelRaw, exists := config["model"]
	if !exists || modelRaw == nil {
		return nil, fmt.Errorf("ai_service.model not configured - add e.g. 'model: \"gemini-pro-latest\"'; there is deliberately no default, because Google closes pinned model generations to newly-issued keys and a stale default fails only at generate time")
	}
	model, ok := modelRaw.(string)
	if !ok || model == "" {
		return nil, fmt.Errorf("ai_service.model must be a non-empty string, got: %T", modelRaw)
	}
	if replacement, retired := geminiRetiredPins[model]; retired {
		return nil, fmt.Errorf("ai_service.model %q is closed to newly-issued API keys - Google answers 404 %q for it (verified against this platform's key 2026-07-24). Use %q, or run scripts/gemini-probe.sh to list what this key can actually reach", model, "not available to new users", replacement)
	}

	client := &GeminiClient{
		apiKey:          apiKey,
		model:           model,
		httpClient:      &http.Client{},
		thinkingReserve: defaultGeminiThinkingReserve,
		embeddingModel:  "text-embedding-004",
	}

	if raw, exists := config["thinking_reserve_tokens"]; exists && raw != nil {
		reserve, err := geminiConfigInt(raw)
		if err != nil {
			return nil, fmt.Errorf("ai_service.thinking_reserve_tokens: %w", err)
		}
		if reserve < 0 {
			return nil, fmt.Errorf("ai_service.thinking_reserve_tokens must not be negative, got %d", reserve)
		}
		client.thinkingReserve = reserve
	}

	levelRaw, hasLevel := config["thinking_level"]
	budgetRaw, hasBudget := config["thinking_budget_tokens"]
	if hasLevel && levelRaw != nil && hasBudget && budgetRaw != nil {
		// Google's own words, verified 2026-07-27: "You can only set only one of
		// thinking budget and thinking level." Caught here so it is a startup
		// error rather than a 400 on every generation.
		return nil, fmt.Errorf("ai_service sets both thinking_level and thinking_budget_tokens - Gemini accepts either but refuses both together (\"You can only set only one of thinking budget and thinking level\"); set one")
	}
	if hasLevel && levelRaw != nil {
		level, ok := levelRaw.(string)
		if !ok || level == "" {
			return nil, fmt.Errorf("ai_service.thinking_level must be a non-empty string, got: %T", levelRaw)
		}
		client.thinkingLevel = level
	}
	if hasBudget && budgetRaw != nil {
		budget, err := geminiConfigInt(budgetRaw)
		if err != nil {
			return nil, fmt.Errorf("ai_service.thinking_budget_tokens: %w", err)
		}
		if budget < 0 {
			return nil, fmt.Errorf("ai_service.thinking_budget_tokens must not be negative, got %d", budget)
		}
		client.thinkingBudget = &budget
	}

	if raw, exists := config["embedding_model"]; exists && raw != nil {
		if embeddingModel, ok := raw.(string); ok && embeddingModel != "" {
			client.embeddingModel = embeddingModel
		}
	}

	return client, nil
}

// geminiConfigInt accepts the number shapes an ai_service block can arrive in:
// JSON config decodes to float64, Go-literal config to int.
func geminiConfigInt(raw interface{}) (int, error) {
	switch v := raw.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("must be a number, got %T", raw)
	}
}

// thinks reports whether this client should reserve budget for thinking. An
// operator who has explicitly set thinking_budget_tokens: 0 has asked for
// thinking off and gets no reserve.
func (c *GeminiClient) thinks() bool {
	if c.thinkingBudget != nil && *c.thinkingBudget == 0 {
		return false
	}
	return geminiModelThinks(c.model)
}

// GenerateText generates text using Gemini
func (c *GeminiClient) GenerateText(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	generationConfig := map[string]interface{}{}

	// visibleBudget is what the CALLER asked for: tokens of answer. Gemini's
	// maxOutputTokens is a total that thinking is drawn from first, so for a
	// thinking model the two are not the same number (see
	// defaultGeminiThinkingReserve).
	visibleBudget := 2048
	if options != nil {
		switch mt := options["max_tokens"].(type) {
		case int:
			visibleBudget = mt
		case float64:
			visibleBudget = int(mt)
		}
		if temperature, ok := options["temperature"].(float64); ok {
			generationConfig["temperature"] = temperature
			options["__sent_temperature"] = temperature
		}
	}

	totalBudget := visibleBudget
	if c.thinks() {
		totalBudget = visibleBudget + c.thinkingReserve
	}
	generationConfig["maxOutputTokens"] = totalBudget

	// Only ever sent when an operator configured it: the two generations take
	// incompatible knobs, so the default is to send no thinkingConfig at all.
	if c.thinkingLevel != "" {
		generationConfig["thinkingConfig"] = map[string]interface{}{"thinkingLevel": c.thinkingLevel}
	} else if c.thinkingBudget != nil {
		generationConfig["thinkingConfig"] = map[string]interface{}{"thinkingBudget": *c.thinkingBudget}
	}

	if options != nil {
		// __sent_max_tokens records what was actually PUT ON THE WIRE, matching
		// anthropic.go, so llm_call_log stays comparable across providers.
		// __sent_visible_budget_tokens records the caller's ask.
		//
		// Consequence worth knowing: llm_call_log's "output_tokens ==
		// max_tokens means the completion was CUT" heuristic goes quiet for a
		// thinking model, because visible output is compared against a total
		// that includes the thinking reserve. That heuristic was always a proxy
		// for the finish reason, and this client returns a *TruncatedError on
		// finishReason=MAX_TOKENS directly, which is the authoritative signal.
		options["__sent_max_tokens"] = totalBudget
		options["__sent_visible_budget_tokens"] = visibleBudget
		options["__sent_thinking_reserve_tokens"] = totalBudget - visibleBudget
	}

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role":  "user",
				"parts": []map[string]string{{"text": prompt}},
			},
		},
		"generationConfig": generationConfig,
	}

	if options != nil {
		if systemPrompt, ok := options["system"].(string); ok && systemPrompt != "" {
			requestBody["systemInstruction"] = map[string]interface{}{
				"parts": []map[string]string{{"text": systemPrompt}},
			}
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent", geminiAPIBase, c.model)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d for model %q: %s", resp.StatusCode, c.model, string(body))
	}

	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
					// Thought marks a part as the model's reasoning rather than
					// its answer. Gemini returns both in the SAME parts array.
					Thought bool `json:"thought"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		PromptFeedback struct {
			BlockReason string `json:"blockReason"`
		} `json:"promptFeedback"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			ThoughtsTokenCount   int `json:"thoughtsTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if options != nil {
		options["__usage_input_tokens"] = response.UsageMetadata.PromptTokenCount
		// __usage_output_tokens stays VISIBLE tokens, matching the other
		// providers. Thinking tokens are billed as output but are not answer,
		// so they are reported separately rather than folded in silently.
		options["__usage_output_tokens"] = response.UsageMetadata.CandidatesTokenCount
		options["__usage_thinking_tokens"] = response.UsageMetadata.ThoughtsTokenCount
		options["__usage_total_tokens"] = response.UsageMetadata.TotalTokenCount
	}

	// The prompt itself can be blocked before any candidate is produced —
	// there is no candidate to inspect at all in that case, unlike a
	// per-candidate finishReason below.
	if response.PromptFeedback.BlockReason != "" {
		return "", &RefusalError{
			Reason:   fmt.Sprintf("promptFeedback.blockReason=%s", response.PromptFeedback.BlockReason),
			Blocks:   nil,
			Provider: "gemini",
			Model:    c.model,
		}
	}

	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	candidate := response.Candidates[0]
	var textBuilder strings.Builder
	thoughtParts := 0
	for _, part := range candidate.Content.Parts {
		if part.Thought {
			// Reasoning is not answer. Concatenating every part would splice the
			// model's thinking into published page copy, and nothing above this
			// layer inspects the text closely enough to notice. Defensive: with
			// no thinking summary requested these should not appear at all.
			thoughtParts++
			continue
		}
		textBuilder.WriteString(part.Text)
	}
	partial := textBuilder.String()

	thinkingNote := ""
	if response.UsageMetadata.ThoughtsTokenCount > 0 || thoughtParts > 0 {
		thinkingNote = fmt.Sprintf(", thinking=%d tokens/%d parts", response.UsageMetadata.ThoughtsTokenCount, thoughtParts)
	}

	// Mirrors anthropic.go: extract the partial BEFORE deciding whether this
	// was a truncation or refusal, so a cut or declined completion still
	// carries back whatever text the model produced (bugs_open/019).
	if candidate.FinishReason == "MAX_TOKENS" {
		reason := fmt.Sprintf("finishReason=MAX_TOKENS (visible=%d%s of maxOutputTokens=%d sent)",
			response.UsageMetadata.CandidatesTokenCount, thinkingNote, totalBudget)
		if response.UsageMetadata.CandidatesTokenCount == 0 && response.UsageMetadata.ThoughtsTokenCount > 0 {
			// The failure that got the provider reverted on 2026-07-24. Naming
			// it here is the difference between a five-minute config change and
			// reading it as the model being incapable.
			reason += " - thinking consumed the entire output ceiling before any visible text; raise ai_service.thinking_reserve_tokens or set ai_service.thinking_level"
		}
		return partial, &TruncatedError{
			Partial:      partial,
			OutputTokens: response.UsageMetadata.CandidatesTokenCount,
			Reason:       reason,
			Provider:     "gemini",
		}
	}

	switch candidate.FinishReason {
	case "SAFETY", "RECITATION", "BLOCKLIST", "PROHIBITED_CONTENT", "SPII":
		return "", &RefusalError{
			Reason:   fmt.Sprintf("finishReason=%s", candidate.FinishReason),
			Blocks:   []string{candidate.FinishReason},
			Provider: "gemini",
			Model:    c.model,
		}
	}

	if partial == "" {
		return "", fmt.Errorf("no text content in response (finishReason=%q, model=%q%s)", candidate.FinishReason, c.model, thinkingNote)
	}

	return partial, nil
}

// GenerateEmbedding generates embeddings using Gemini's embedding model.
//
// The model is configurable (ai_service.embedding_model) for the same reason the
// generation model has no default: embedding snapshots retire too, and a
// hardcoded one turns into a 404 nobody can fix from config.
func (c *GeminiClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	requestBody := map[string]interface{}{
		"model": "models/" + c.embeddingModel,
		"content": map[string]interface{}{
			"parts": []map[string]string{{"text": text}},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:embedContent", geminiAPIBase, c.embeddingModel)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedding response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API returned status %d for model %q: %s", resp.StatusCode, c.embeddingModel, string(body))
	}

	var response struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if len(response.Embedding.Values) == 0 {
		return nil, fmt.Errorf("empty embedding from gemini")
	}

	return response.Embedding.Values, nil
}

func (c *GeminiClient) Model() string {
	return c.model
}

func (c *GeminiClient) Provider() string {
	return "gemini"
}
