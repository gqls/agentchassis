// FILE: platform/orchestration/actions/feed_actions.go
// Actions for the content feed pipeline:
//   - FetchRSSAction: fetches and parses RSS/Atom feeds (sync)
//   - FetchLLMNewsAction: fetches news via xAI/Grok API (sync)
//   - WriteFeedItemsAction: normalises items from any source and writes to content_feed_items
//   - LoadDueSourcesAction: queries content_sources for sources due to be fetched
//   - UpdateSourceTimestampsAction: updates last_fetched_at and next_fetch_at after ingestion

package actions

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Input specs
// ---------------------------------------------------------------------------

var FetchRSSInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"source_config"},
	Optional:    []string{"source_id"},
}

var FetchLLMNewsInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"source_config"},
	Optional:    []string{"source_id"},
}

var WriteFeedItemsInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id", "source_id"},
	Optional: []string{"items", "source_type"},
}

var LoadDueSourcesInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"source_type"},
}

var UpdateSourceTimestampsInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"source_id"},
	Optional:    []string{"error_message"},
}

func init() {
	datahelpers.RegisterActionInputSpec("fetch_rss", FetchRSSInputSpec)
	datahelpers.RegisterActionInputSpec("fetch_llm_news", FetchLLMNewsInputSpec)
	datahelpers.RegisterActionInputSpec("write_feed_items", WriteFeedItemsInputSpec)
	datahelpers.RegisterActionInputSpec("load_due_sources", LoadDueSourcesInputSpec)
	datahelpers.RegisterActionInputSpec("update_source_timestamps", UpdateSourceTimestampsInputSpec)
}

// ---------------------------------------------------------------------------
// RSS/Atom XML structures
// ---------------------------------------------------------------------------

type rssDocument struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string   `xml:"title"`
	Link      atomLink `xml:"link"`
	Summary   string   `xml:"summary"`
	Content   string   `xml:"content"`
	Published string   `xml:"published"`
	Updated   string   `xml:"updated"`
	ID        string   `xml:"id"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

// ---------------------------------------------------------------------------
// FetchRSSAction — sync HTTP GET + XML parse
// ---------------------------------------------------------------------------
//
// Config:
//   - source_config: the content_sources.config JSONB (must contain feed_url)
//   - source_id: UUID of the content_source row
//
// Output:
//   - items: array of normalised feed items
//   - source_id: passed through
//   - feed_url: the URL fetched
//   - item_count: number of items found
//   - fetched_at: ISO timestamp

func FetchRSSAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "fetch_rss"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		FetchRSSInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	sourceConfig := inputs.GetMap("source_config")
	if sourceConfig == nil {
		return nil, fmt.Errorf("source_config is required and must be a map")
	}

	feedURL, _ := sourceConfig["feed_url"].(string)
	if feedURL == "" {
		return nil, fmt.Errorf("feed_url not found in source_config")
	}

	maxItems := 20
	if mi, ok := sourceConfig["max_items"].(float64); ok {
		maxItems = int(mi)
	}

	sourceID := inputs.Get("source_id")

	logger.Info("FetchRSSAction: fetching feed",
		zap.String("feed_url", feedURL),
		zap.String("source_id", sourceID),
		zap.Int("max_items", maxItems),
	)

	// HTTP GET with timeout
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", feedURL, err)
	}
	req.Header.Set("User-Agent", "PersonaeBot/1.0 (News Aggregator)")
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml")

	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("FetchRSSAction: HTTP request failed",
			zap.String("feed_url", feedURL),
			zap.Error(err),
		)
		return map[string]interface{}{
			"items":      []interface{}{},
			"source_id":  sourceID,
			"feed_url":   feedURL,
			"item_count": 0,
			"error":      err.Error(),
			"fetched_at": time.Now().UTC().Format(time.RFC3339),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{
			"items":      []interface{}{},
			"source_id":  sourceID,
			"feed_url":   feedURL,
			"item_count": 0,
			"error":      fmt.Sprintf("HTTP %d", resp.StatusCode),
			"fetched_at": time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // 5MB limit
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try RSS first, then Atom
	items := parseRSSBody(body, maxItems, logger)
	if len(items) == 0 {
		items = parseAtomBody(body, maxItems, logger)
	}

	logger.Info("FetchRSSAction: parsed feed",
		zap.String("feed_url", feedURL),
		zap.Int("items_found", len(items)),
	)

	return map[string]interface{}{
		"items":      items,
		"source_id":  sourceID,
		"feed_url":   feedURL,
		"item_count": len(items),
		"fetched_at": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func parseRSSBody(body []byte, maxItems int, logger *zap.Logger) []map[string]interface{} {
	var doc rssDocument
	if err := xml.Unmarshal(body, &doc); err != nil {
		logger.Info("FetchRSSAction: not RSS format, will try Atom", zap.Error(err))
		return nil
	}

	var items []map[string]interface{}
	for i, item := range doc.Channel.Items {
		if i >= maxItems {
			break
		}

		publishedAt := parseRSSDate(item.PubDate)
		externalID := item.GUID
		if externalID == "" {
			externalID = item.Link
		}

		items = append(items, map[string]interface{}{
			"title":        strings.TrimSpace(item.Title),
			"url":          strings.TrimSpace(item.Link),
			"summary":      stripHTML(item.Description),
			"external_id":  externalID,
			"published_at": publishedAt,
		})
	}
	return items
}

func parseAtomBody(body []byte, maxItems int, logger *zap.Logger) []map[string]interface{} {
	var feed atomFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		logger.Info("FetchRSSAction: not Atom format either", zap.Error(err))
		return nil
	}

	var items []map[string]interface{}
	for i, entry := range feed.Entries {
		if i >= maxItems {
			break
		}

		publishedAt := entry.Published
		if publishedAt == "" {
			publishedAt = entry.Updated
		}

		summary := entry.Summary
		if summary == "" {
			summary = entry.Content
		}

		items = append(items, map[string]interface{}{
			"title":        strings.TrimSpace(entry.Title),
			"url":          strings.TrimSpace(entry.Link.Href),
			"summary":      stripHTML(summary),
			"external_id":  entry.ID,
			"published_at": publishedAt,
		})
	}
	return items
}

// parseRSSDate handles common RSS date formats
func parseRSSDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC3339,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}

	for _, f := range formats {
		if t, err := time.Parse(f, strings.TrimSpace(dateStr)); err == nil {
			return t.UTC().Format(time.RFC3339)
		}
	}
	return dateStr // return raw if can't parse
}

// stripHTML removes HTML tags from a string (simple approach)
func stripHTML(s string) string {
	if s == "" {
		return ""
	}
	// Simple tag stripping — good enough for RSS summaries
	var result strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

// ---------------------------------------------------------------------------
// FetchLLMNewsAction — sync HTTP call to LLM API for news
// ---------------------------------------------------------------------------
//
// All providers now use real-time web search:
//
//   xAI (grok-4-1-fast): Responses API with web_search + x_search tools.
//     Searches the web and X (Twitter) in real-time.
//
//   OpenAI (gpt-4.1-mini): Responses API with web_search tool.
//     Searches the web in real-time via Bing.
//
//   Perplexity (sonar): Chat completions with built-in search.
//     Every Sonar request searches the web automatically — no tools config.
//     Returns citations in the response.
//
// Config (in content_sources.config):
//   Required: provider, model, prompt_template
//   Optional: hours_lookback (default 12), max_items (default 10),
//             search_tools (xAI only, default ["web_search"])
//
// Output: same shape as FetchRSSAction
//   {items, source_id, provider, item_count, fetched_at}

func FetchLLMNewsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "fetch_llm_news"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		FetchLLMNewsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	sourceConfig := inputs.GetMap("source_config")
	if sourceConfig == nil {
		return nil, fmt.Errorf("source_config is required and must be a map")
	}

	provider, _ := sourceConfig["provider"].(string)
	model, _ := sourceConfig["model"].(string)
	promptTemplate, _ := sourceConfig["prompt_template"].(string)

	if provider == "" || model == "" || promptTemplate == "" {
		return nil, fmt.Errorf("source_config must contain provider, model, and prompt_template")
	}

	hoursLookback := 12
	if hl, ok := sourceConfig["hours_lookback"].(float64); ok {
		hoursLookback = int(hl)
	}

	maxItems := 10
	if mi, ok := sourceConfig["max_items"].(float64); ok {
		maxItems = int(mi)
	}

	sourceID := inputs.Get("source_id")
	prompt := strings.ReplaceAll(promptTemplate, "{{.hours}}", fmt.Sprintf("%d", hoursLookback))

	logger.Info("FetchLLMNewsAction: calling LLM",
		zap.String("provider", provider),
		zap.String("model", model),
		zap.String("source_id", sourceID),
		zap.Int("hours_lookback", hoursLookback),
	)

	apiURL, apiKey, err := resolveLLMNewsProvider(provider)
	if err != nil {
		return nil, err
	}

	// Route to the right API format
	providerLower := strings.ToLower(provider)
	switch providerLower {
	case "xai", "grok", "openai":
		return fetchViaResponsesAPI(ctx, apiURL, apiKey, model, prompt, maxItems, sourceID, providerLower, sourceConfig, logger)
	case "perplexity":
		return fetchViaPerplexity(ctx, apiURL, apiKey, model, prompt, maxItems, sourceID, sourceConfig, logger)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// ---------------------------------------------------------------------------
// Responses API path — used by xAI and OpenAI
// ---------------------------------------------------------------------------
// Both use /v1/responses with tools array and input array.
// Response format: { output: [{ type: "message", content: [{ type: "output_text", text: "..." }] }] }

func fetchViaResponsesAPI(ctx context.Context, apiURL, apiKey, model, prompt string, maxItems int, sourceID, provider string, sourceConfig map[string]interface{}, logger *zap.Logger) (interface{}, error) {
	emptyResult := func(errMsg string) (interface{}, error) {
		return map[string]interface{}{
			"items":      []interface{}{},
			"source_id":  sourceID,
			"provider":   provider,
			"item_count": 0,
			"error":      errMsg,
			"fetched_at": time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	// Build tools array based on provider
	var tools []map[string]interface{}
	switch provider {
	case "xai", "grok":
		// Default: web_search. Optionally add x_search from config.
		tools = []map[string]interface{}{{"type": "web_search"}}
		if searchTools, ok := sourceConfig["search_tools"].([]interface{}); ok {
			tools = nil
			for _, t := range searchTools {
				if ts, ok := t.(string); ok {
					tools = append(tools, map[string]interface{}{"type": ts})
				}
			}
		}
	case "openai":
		tools = []map[string]interface{}{{"type": "web_search"}}
	}

	systemPrompt := "You are a news research assistant with access to web search. " +
		"Use your search tools to find REAL, CURRENT news articles. " +
		"Return ONLY a valid JSON array of news items you found. " +
		"Each object must have: title (string), summary (string, 2-3 sentences), " +
		"source_url (string, the ACTUAL URL you found), source_name (string, the publication name), " +
		"published_at (string, ISO 8601 format). " +
		"Only include items you found via search with real URLs. Do not fabricate any URLs or articles."

	reqBody := map[string]interface{}{
		"model": model,
		"input": []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"tools": tools,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Longer timeout — model does multiple searches server-side
	client := &http.Client{Timeout: 90 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(reqBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("FetchLLMNewsAction: API call failed",
			zap.String("provider", provider), zap.Error(err))
		return emptyResult(err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 100*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("%s API returned HTTP %d: %s", provider, resp.StatusCode, datahelpers.TruncateString(string(respBody), 500))
		logger.Warn("FetchLLMNewsAction: API error", zap.String("error", errMsg))
		return emptyResult(errMsg)
	}

	// Parse the Responses API response
	// Structure: { output: [{ type: "message", content: [{ type: "output_text", text: "..." }] }] }
	var apiResp struct {
		Output []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		logger.Warn("FetchLLMNewsAction: failed to decode response",
			zap.String("provider", provider),
			zap.String("body_preview", datahelpers.TruncateString(string(respBody), 500)),
			zap.Error(err))
		return emptyResult(fmt.Sprintf("response decode error: %s", err.Error()))
	}

	// Extract text content
	var content string
	for _, output := range apiResp.Output {
		if output.Type != "message" {
			continue
		}
		for _, c := range output.Content {
			if c.Type == "output_text" && c.Text != "" {
				content = c.Text
				break
			}
		}
		if content != "" {
			break
		}
	}

	if content == "" {
		logger.Warn("FetchLLMNewsAction: no text content in response",
			zap.String("provider", provider),
			zap.String("body_preview", datahelpers.TruncateString(string(respBody), 500)))
		return emptyResult("no text content in response")
	}

	items := parseLLMNewsItems(content, maxItems, sourceID, logger)

	logger.Info("FetchLLMNewsAction: search returned items",
		zap.String("provider", provider),
		zap.Int("items_found", len(items)),
		zap.String("source_id", sourceID))

	return map[string]interface{}{
		"items":      items,
		"source_id":  sourceID,
		"provider":   provider,
		"item_count": len(items),
		"fetched_at": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ---------------------------------------------------------------------------
// Perplexity path — chat completions with built-in search
// ---------------------------------------------------------------------------
// Sonar models always search the web — no tools config needed.
// Uses OpenAI-compatible chat completions format.

func fetchViaPerplexity(ctx context.Context, apiURL, apiKey, model, prompt string, maxItems int, sourceID string, sourceConfig map[string]interface{}, logger *zap.Logger) (interface{}, error) {
	emptyResult := func(errMsg string) (interface{}, error) {
		return map[string]interface{}{
			"items":      []interface{}{},
			"source_id":  sourceID,
			"provider":   "perplexity",
			"item_count": 0,
			"error":      errMsg,
			"fetched_at": time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	// The output budget is configuration, exactly like every other number in
	// source_config (hours_lookback, max_items). Until 2026-09-03 it was a
	// hardcoded 4096 — the bugs_open/257 class in a spelling no census had caught,
	// because this path talks to Perplexity over raw HTTP and never touches
	// platform/aiservice, so a grep for GenerateText could not see it. It was
	// found by the package-wide audit in llm_budget_call_sites_test.go.
	//
	// 4096 stays as the DEFAULT when nobody has chosen one. That is the honest
	// difference from what this replaced: a fallback an operator can override,
	// rather than a number no configuration can reach.
	maxTokens := 4096
	if mt, ok := sourceConfig["max_tokens"].(float64); ok && mt > 0 {
		maxTokens = int(mt)
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{
				"role": "system",
				"content": "You are a news research assistant. Search the web for current news. " +
					"Return ONLY a valid JSON array of news items you found. " +
					"Each object must have: title (string), summary (string, 2-3 sentences), " +
					"source_url (string, the actual URL), source_name (string, the publication), " +
					"published_at (string, ISO 8601 format). " +
					"Only include real articles with real URLs.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature":      0.3,
		"max_tokens":       maxTokens,
		"return_citations": true,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Perplexity request: %w", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(reqBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Perplexity request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("FetchLLMNewsAction: Perplexity API call failed", zap.Error(err))
		return emptyResult(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		errMsg := fmt.Sprintf("Perplexity API returned HTTP %d: %s", resp.StatusCode, string(respBody))
		logger.Warn("FetchLLMNewsAction: Perplexity API error", zap.String("error", errMsg))
		return emptyResult(errMsg)
	}

	var llmResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return nil, fmt.Errorf("failed to decode Perplexity response: %w", err)
	}

	if len(llmResp.Choices) == 0 {
		return emptyResult("no choices in Perplexity response")
	}

	content := strings.TrimSpace(llmResp.Choices[0].Message.Content)
	items := parseLLMNewsItems(content, maxItems, sourceID, logger)

	logger.Info("FetchLLMNewsAction: Perplexity returned items",
		zap.Int("items_found", len(items)),
		zap.String("source_id", sourceID))

	return map[string]interface{}{
		"items":      items,
		"source_id":  sourceID,
		"provider":   "perplexity",
		"item_count": len(items),
		"fetched_at": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ---------------------------------------------------------------------------
// Shared: parse JSON array of news items from LLM text output
// ---------------------------------------------------------------------------

func parseLLMNewsItems(content string, maxItems int, sourceID string, logger *zap.Logger) []map[string]interface{} {
	// Strip markdown code fences if present
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var rawItems []map[string]interface{}
	if err := json.Unmarshal([]byte(content), &rawItems); err != nil {
		logger.Warn("parseLLMNewsItems: failed to parse as JSON array",
			zap.String("content_preview", datahelpers.TruncateString(content, 300)),
			zap.Error(err))
		return nil
	}

	var items []map[string]interface{}
	for i, raw := range rawItems {
		if i >= maxItems {
			break
		}
		title, _ := raw["title"].(string)
		summary, _ := raw["summary"].(string)
		sourceURL, _ := raw["source_url"].(string)
		sourceName, _ := raw["source_name"].(string)
		publishedAt, _ := raw["published_at"].(string)

		externalID := sourceURL
		if externalID == "" {
			h := sha256.Sum256([]byte(title + publishedAt))
			externalID = fmt.Sprintf("llm-%x", h[:8])
		}

		if sourceName != "" && summary != "" {
			summary = fmt.Sprintf("[%s] %s", sourceName, summary)
		}

		items = append(items, map[string]interface{}{
			"title":        strings.TrimSpace(title),
			"url":          strings.TrimSpace(sourceURL),
			"summary":      strings.TrimSpace(summary),
			"external_id":  externalID,
			"published_at": publishedAt,
		})
	}
	return items
}

// resolveLLMNewsProvider returns API URL and key for the given provider.
//
// xAI:        Responses API — /v1/responses (web_search + x_search tools)
//
//	Recommended model: grok-4-1-fast
//
// OpenAI:     Responses API — /v1/responses (web_search tool)
//
//	Recommended model: gpt-4.1-mini
//
// Perplexity: Chat completions — /chat/completions (built-in search via Sonar)
//
//	Recommended model: sonar (fast) or sonar-pro (deeper)
func resolveLLMNewsProvider(provider string) (string, string, error) {
	switch strings.ToLower(provider) {
	case "xai", "grok":
		apiKey := os.Getenv("XAI_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("GROK_API_KEY")
		}
		if apiKey == "" {
			return "", "", fmt.Errorf("XAI_API_KEY or GROK_API_KEY environment variable not set")
		}
		return "https://api.x.ai/v1/responses", apiKey, nil
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return "", "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
		}
		return "https://api.openai.com/v1/responses", apiKey, nil
	case "perplexity":
		apiKey := os.Getenv("PERPLEXITY_API_KEY")
		if apiKey == "" {
			return "", "", fmt.Errorf("PERPLEXITY_API_KEY environment variable not set")
		}
		return "https://api.perplexity.ai/chat/completions", apiKey, nil
	default:
		return "", "", fmt.Errorf("unsupported LLM news provider: %s (supported: xai, openai, perplexity)", provider)
	}
}

// ---------------------------------------------------------------------------
// WriteFeedItemsAction — normalise and write items to content_feed_items
// ---------------------------------------------------------------------------
//
// This is the shared sink for all source types. It takes the normalised
// items array from FetchRSS, FetchLLMNews, or from web_search/scrape
// results post-processing, and writes them to the database with dedup.
//
// Config:
//   - site_id: UUID
//   - source_id: UUID
//   - items: array of {title, url, summary, external_id, published_at}
//   - source_type: string (for logging/metrics)
//
// Dedup: skips items whose source_url already exists in content_feed_items
// for this site (using the idx_cfi_dedup partial index).

func WriteFeedItemsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "write_feed_items"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		WriteFeedItemsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	sourceIDStr := inputs.Get("source_id")
	sourceType := inputs.Get("source_type")

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	var sourceID *uuid.UUID
	if sourceIDStr != "" {
		parsed, err := uuid.Parse(sourceIDStr)
		if err != nil {
			logger.Warn("WriteFeedItemsAction: invalid source_id, will be null",
				zap.String("source_id", sourceIDStr))
		} else {
			sourceID = &parsed
		}
	}

	// Get items from collected data
	itemsRaw := datahelpers.ExtractNestedField(params.CollectedData,
		datahelpers.GetStringField(params.StepConfig.Config, "items_field", "items"))
	if itemsRaw == nil {
		// Try fetched_items, rss_result.items, etc.
		for _, fallback := range []string{"fetched_items", "rss_result.items", "llm_result.items", "search_results.results"} {
			itemsRaw = datahelpers.ExtractNestedField(params.CollectedData, fallback)
			if itemsRaw != nil {
				break
			}
		}
	}

	itemsSlice, ok := itemsRaw.([]interface{})
	if !ok || len(itemsSlice) == 0 {
		logger.Info("WriteFeedItemsAction: no items to write",
			zap.String("site_id", siteIDStr),
			zap.String("source_type", sourceType),
		)
		return map[string]interface{}{
			"written":   0,
			"skipped":   0,
			"site_id":   siteIDStr,
			"source_id": sourceIDStr,
		}, nil
	}

	written := 0
	skipped := 0

	for _, itemRaw := range itemsSlice {
		item, ok := itemRaw.(map[string]interface{})
		if !ok {
			continue
		}

		title, _ := item["title"].(string)
		url, _ := item["url"].(string)
		summary, _ := item["summary"].(string)
		externalID, _ := item["external_id"].(string)
		publishedAtStr, _ := item["published_at"].(string)

		if title == "" {
			skipped++
			continue
		}

		// Parse published_at
		var publishedAt *time.Time
		if publishedAtStr != "" {
			if t, err := time.Parse(time.RFC3339, publishedAtStr); err == nil {
				publishedAt = &t
			}
		}

		// Validate published date — skip items with obviously wrong dates
		if publishedAt != nil {
			age := time.Since(*publishedAt)
			if age > 30*24*time.Hour {
				logger.Info("WriteFeedItemsAction: skipping item with old published date",
					zap.String("title", title),
					zap.Time("published_at", *publishedAt),
					zap.Duration("age", age))
				skipped++
				continue
			}
			if age < -24*time.Hour { // more than 1 day in the future
				logger.Info("WriteFeedItemsAction: skipping item with future published date",
					zap.String("title", title),
					zap.Time("published_at", *publishedAt))
				skipped++
				continue
			}
		}

		// Insert with dedup — ON CONFLICT on source_url (via partial unique index)
		// The idx_cfi_dedup index covers source_url WHERE status NOT IN (duplicate, expired, rejected)
		// We use a simple existence check instead of relying on UNIQUE constraint
		var exists bool
		if url != "" {
			err := params.DB.QueryRowContext(ctx, `
				SELECT EXISTS(
					SELECT 1 FROM content_feed_items
					WHERE site_id = $1
					  AND source_url = $2
					  AND status NOT IN ('duplicate', 'expired', 'rejected')
				)
			`, siteID, url).Scan(&exists)
			if err != nil {
				logger.Warn("WriteFeedItemsAction: dedup check failed",
					zap.String("url", url), zap.Error(err))
			}
		}

		if exists {
			skipped++
			continue
		}

		_, err := params.DB.ExecContext(ctx, `
			INSERT INTO content_feed_items (
				site_id, source_id, external_id, source_url,
				source_title, source_summary, source_published_at,
				status, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, 'ingested', NOW())
		`, siteID, sourceID, externalID, url, title, summary, publishedAt)

		if err != nil {
			logger.Warn("WriteFeedItemsAction: insert failed",
				zap.String("title", title),
				zap.String("url", url),
				zap.Error(err),
			)
			skipped++
			continue
		}

		written++
	}

	logger.Info("WriteFeedItemsAction: complete",
		zap.String("site_id", siteIDStr),
		zap.String("source_type", sourceType),
		zap.Int("written", written),
		zap.Int("skipped", skipped),
		zap.Int("total", len(itemsSlice)),
	)

	return map[string]interface{}{
		"written":   written,
		"skipped":   skipped,
		"total":     len(itemsSlice),
		"site_id":   siteIDStr,
		"source_id": sourceIDStr,
	}, nil
}

// ---------------------------------------------------------------------------
// LoadDueSourcesAction — query content_sources for sources needing fetch
// ---------------------------------------------------------------------------
//
// Returns all active sources for a site that are due within the half-cadence
// look-ahead (feedSourceDuePredicate — see feed_due_lookahead.go and
// bugs_open/410 for why a bare NOW() phase-locks cadence-length intervals).
//
// No live workflow step names this action as of 2026-08-26 (the
// content-feed-orchestrator dispatches via dispatch_feed_sources instead); it
// keeps the shared predicate so any future caller inherits the fixed due test.
//
// Config:
//   - site_id: UUID
//   - source_type: (optional) filter by type
//
// Output:
//   - sources: array of source objects with id, source_type, name, config
//   - source_count: number of due sources

func LoadDueSourcesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "load_due_sources"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		LoadDueSourcesInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	sourceTypeFilter := inputs.Get("source_type")

	query := `
		SELECT id, source_type, name, config, fetch_interval
		FROM content_sources
		WHERE site_id = $1
		  AND is_active = true
		  AND ` + feedSourceDuePredicate + `
	`
	args := []interface{}{siteID}

	if sourceTypeFilter != "" {
		query += " AND source_type = $2"
		args = append(args, sourceTypeFilter)
	}

	query += " ORDER BY next_fetch_at ASC NULLS FIRST"

	rows, err := params.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query due sources: %w", err)
	}
	defer rows.Close()

	var sources []map[string]interface{}
	for rows.Next() {
		var (
			id            string
			sourceType    string
			name          string
			configJSON    []byte
			fetchInterval string
		)
		if err := rows.Scan(&id, &sourceType, &name, &configJSON, &fetchInterval); err != nil {
			logger.Warn("LoadDueSourcesAction: scan failed", zap.Error(err))
			continue
		}

		var config map[string]interface{}
		if err := json.Unmarshal(configJSON, &config); err != nil {
			logger.Warn("LoadDueSourcesAction: config parse failed",
				zap.String("source_id", id), zap.Error(err))
			config = map[string]interface{}{}
		}

		sources = append(sources, map[string]interface{}{
			"id":             id,
			"source_type":    sourceType,
			"name":           name,
			"config":         config,
			"fetch_interval": fetchInterval,
		})
	}

	logger.Info("LoadDueSourcesAction: found due sources",
		zap.String("site_id", siteIDStr),
		zap.Int("source_count", len(sources)),
	)

	return map[string]interface{}{
		"sources":      sources,
		"source_count": len(sources),
		"site_id":      siteIDStr,
	}, nil
}

// ---------------------------------------------------------------------------
// UpdateSourceTimestampsAction — update last_fetched_at and next_fetch_at
// ---------------------------------------------------------------------------
//
// Called after a source has been fetched (success or failure).
// On success: updates last_fetched_at, resets error_count, sets next_fetch_at.
// On failure: increments error_count, records error, still sets next_fetch_at
// (with exponential backoff on repeated errors).

func UpdateSourceTimestampsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "update_source_timestamps"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		UpdateSourceTimestampsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	sourceIDStr := inputs.Get("source_id")
	errorMessage := inputs.Get("error_message")

	sourceID, err := uuid.Parse(sourceIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid source_id %q: %w", sourceIDStr, err)
	}

	if errorMessage != "" {
		// Failure path: increment error_count, set backoff
		_, err = params.DB.ExecContext(ctx, `
			UPDATE content_sources
			SET last_fetched_at = NOW(),
			    error_count = error_count + 1,
			    last_error = $2,
			    last_error_at = NOW(),
			    next_fetch_at = NOW() + (fetch_interval * LEAST(error_count + 1, 4)),
			    updated_at = NOW()
			WHERE id = $1
		`, sourceID, errorMessage)
	} else {
		// Success path: reset errors, schedule next fetch
		_, err = params.DB.ExecContext(ctx, `
			UPDATE content_sources
			SET last_fetched_at = NOW(),
			    error_count = 0,
			    last_error = NULL,
			    last_error_at = NULL,
			    next_fetch_at = NOW() + fetch_interval,
			    updated_at = NOW()
			WHERE id = $1
		`, sourceID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update source timestamps: %w", err)
	}

	logger.Info("UpdateSourceTimestampsAction: updated",
		zap.String("source_id", sourceIDStr),
		zap.Bool("had_error", errorMessage != ""),
	)

	return map[string]interface{}{
		"source_id": sourceIDStr,
		"updated":   true,
	}, nil
}
