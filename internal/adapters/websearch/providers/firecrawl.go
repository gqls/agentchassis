// FILE: internal/adapters/websearch/providers/firecrawl.go
//
//   1. API URL: v0/search -> v2/search
//   2. Request payload: flat {query, limit} instead of nested pageOptions/searchOptions
//   3. Response struct: data.web[] array instead of flat data[]
//

package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"
)

type FirecrawlProvider struct {
	apiKey     string
	httpClient *http.Client
	apiURL     string
	logger     *zap.Logger
}

func NewFirecrawlProvider(httpClient *http.Client, logger *zap.Logger) *FirecrawlProvider {
	return &FirecrawlProvider{
		apiKey:     os.Getenv("FIRECRAWL_API_KEY"),
		httpClient: httpClient,
		apiURL:     "https://api.firecrawl.dev/v2/search", // was v0/search
		logger:     logger.With(zap.String("provider", "firecrawl")),
	}
}

func (f *FirecrawlProvider) Name() string {
	return "firecrawl"
}

func (f *FirecrawlProvider) IsAvailable() bool {
	return f.apiKey != ""
}

func (f *FirecrawlProvider) Search(ctx context.Context, query string, numResults int, opts SearchOptions) ([]SearchResult, error) {
	if numResults == 0 {
		numResults = 10
	}

	// v2 uses flat payload with "limit" at top level
	// Was: nested "pageOptions" and "searchOptions"
	payload := map[string]interface{}{
		"query": query,
		"limit": numResults,
	}

	// v2 sources selects the result vertical; the default (absent) is web.
	// Images is a documented source but has no response parsing here, so
	// decline it rather than serve mislabelled results (bugs_open/127).
	switch opts.SearchType {
	case "", "web":
	case "news":
		payload["sources"] = []string{"news"}
	default:
		return nil, fmt.Errorf("firecrawl search_type %q not supported here: %w", opts.SearchType, ErrUnsupportedSearchType)
	}
	if opts.TimeRange != "" {
		if tbs, ok := googleTbsByTimeRange[opts.TimeRange]; ok {
			payload["tbs"] = tbs
		} else {
			f.logger.Warn("Unrecognised time_range, searching without a date filter",
				zap.String("time_range", opts.TimeRange))
		}
	}
	// v2's "country" param geo-targets both web and news sources (verified
	// against Firecrawl's own docs) and defaults to "US" when absent —
	// bugs_open/427 rider: this default is why a UK site's news came back
	// American. Region is opt-in per source (SeedContentSourcesAction sets
	// it for .uk/.co.uk sites), so absence here preserves the prior default.
	if opts.Region != "" {
		payload["country"] = strings.ToUpper(opts.Region)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	f.logger.Info("Executing search",
		zap.String("query", query),
		zap.Int("num_results", numResults),
		zap.String("search_type", opts.SearchType),
		zap.String("time_range", opts.TimeRange),
		zap.String("region", opts.Region))

	req, err := http.NewRequestWithContext(ctx, "POST", f.apiURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+f.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	// Response struct matches v2 format: each requested source comes back
	// under its own key. Web items carry description; news items carry
	// snippet and a human-readable relative date ("3 months ago").
	type fcResult struct {
		URL         string `json:"url"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Snippet     string `json:"snippet"`
		Date        string `json:"date,omitempty"`
		Position    int    `json:"position,omitempty"`
	}
	var apiResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message,omitempty"`
		Data    struct {
			Web  []fcResult `json:"web"`
			News []fcResult `json:"news"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResponse.Success {
		return nil, fmt.Errorf("search failed: %s", apiResponse.Message)
	}

	raw := append(apiResponse.Data.News, apiResponse.Data.Web...)
	now := time.Now()
	results := make([]SearchResult, 0, len(raw))
	for _, r := range raw {
		snippet := strings.TrimSpace(r.Description)
		if snippet == "" {
			snippet = strings.TrimSpace(r.Snippet)
		}
		snippet = truncateSnippet(snippet, snippetMaxBytes)

		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Snippet:     snippet,
			PublishedAt: normalisePublishedAt(r.Date, now),
			Source:      f.Name(),
		})
	}

	f.logger.Info("Search completed",
		zap.Int("results_count", len(results)))

	return results, nil
}

// snippetMaxBytes caps a stored result summary. Deliberately a constant and not
// a config key: raising it does not fix the defect truncateSnippet exists for
// (any budget can land mid-link), and `web_search` has no ActionInputSpec at
// all, so an optional key here would be added to a surface the RFC_022 audit
// already reports as "NOT COUNTED — unknowable" rather than under budget.
const snippetMaxBytes = 200

// truncateSnippet cuts s to at most max BYTES, on a rune boundary, without
// leaving a half-written markdown link behind. Both halves fix real damage
// measured 2026-09-03 over 30 days of content_feed_items (5,863 rows):
//
//   - the previous `snippet[:197]` was a byte slice, so it could split a
//     multi-byte rune and emit U+FFFD — 2 rows already carry one;
//   - 288 rows carry an unclosed `](url` tail. That is manufactured HERE: a cut
//     landing inside `[text](url)` leaves a fragment the downstream literal
//     markdown strip is structurally blind to, because its pattern needs the
//     closing paren. The raw URL then reaches visitors as article text
//     (bugs_open/332).
//
// Dropping the partial link is NOT the contested "strip markdown before the
// cut" transform: it removes only bytes that truncation has already made
// unusable, and it changes the truncation BOUNDARY rather than the content.
// Stripping markdown wholesale stays out on purpose — it writes a lossy
// transform to disk that the DISABLE_NEWS_MARKDOWN_STRIP kill switch cannot
// undo, on a path shared with web search, and it would make the reasoning-set
// corpus silently bimodal. Display-time stripping belongs in the shared
// queryresolve projection every reader of source_summary calls.
func truncateSnippet(s string, max int) string {
	if max < 4 || len(s) <= max {
		return s
	}
	cut := max - len("...")
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	out := s[:cut]

	// Only a genuine link opening is trimmed — `[` followed by `](` with no
	// closing paren. A bare bracket in prose (a footnote marker like "[1]")
	// carries no URL and is left alone.
	if i := strings.LastIndexByte(out, '['); i >= 0 {
		tail := out[i:]
		if j := strings.Index(tail, "]("); j >= 0 && !strings.Contains(tail[j:], ")") {
			out = strings.TrimRight(out[:i], " ")
		}
	}

	return out + "..."
}
