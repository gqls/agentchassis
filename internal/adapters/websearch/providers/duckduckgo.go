// FILE: internal/adapters/websearch/providers/duckduckgo.go
package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// DuckDuckGoProvider searches via DuckDuckGo's HTML interface.
// No API key required. Uses the lite HTML endpoint which is
// simpler to parse and lighter on resources.
type DuckDuckGoProvider struct {
	httpClient    *http.Client
	logger        *zap.Logger
	lastRequestAt time.Time
	minRequestGap time.Duration
	enabled       bool
}

func NewDuckDuckGoProvider(httpClient *http.Client, logger *zap.Logger) *DuckDuckGoProvider {
	return &DuckDuckGoProvider{
		httpClient:    httpClient,
		logger:        logger.With(zap.String("provider", "duckduckgo")),
		minRequestGap: 2 * time.Second, // Be polite — at least 2s between requests
		enabled:       true,            // Always available, no API key needed
	}
}

func (d *DuckDuckGoProvider) Name() string {
	return "duckduckgo"
}

func (d *DuckDuckGoProvider) IsAvailable() bool {
	return d.enabled
}

func (d *DuckDuckGoProvider) Search(ctx context.Context, query string, numResults int, opts SearchOptions) ([]SearchResult, error) {
	if numResults == 0 {
		numResults = 10
	}

	// The html.duckduckgo.com endpoint has no news or image vertical, so a
	// non-web search must be declined rather than answered with web results
	// wearing the requested label (bugs_open/127).
	if opts.SearchType != "" && opts.SearchType != "web" {
		return nil, fmt.Errorf("duckduckgo html endpoint has no %q vertical: %w", opts.SearchType, ErrUnsupportedSearchType)
	}
	if opts.TimeRange != "" {
		if _, ok := ddgTimeFilters[opts.TimeRange]; !ok {
			d.logger.Warn("Unrecognised time_range, searching without a date filter",
				zap.String("time_range", opts.TimeRange))
		}
	}

	// Rate limiting — wait if we're calling too quickly
	if !d.lastRequestAt.IsZero() {
		elapsed := time.Since(d.lastRequestAt)
		if elapsed < d.minRequestGap {
			wait := d.minRequestGap - elapsed
			d.logger.Debug("Rate limiting: waiting before next request",
				zap.Duration("wait", wait))
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	d.lastRequestAt = time.Now()

	searchURL := buildDDGSearchURL(query, opts)

	d.logger.Info("Executing DuckDuckGo search",
		zap.String("query", query),
		zap.Int("num_results", numResults))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to look like a normal browser request
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AgentChassis/1.0; +https://github.com/gqls/agentchassis)")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.9")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == 403 {
		return nil, fmt.Errorf("DuckDuckGo rate limited (status %d), try again later", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DuckDuckGo returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	htmlStr := string(body)

	// Parse results from the HTML
	results := d.parseHTMLResults(htmlStr, numResults)

	d.logger.Info("DuckDuckGo search completed",
		zap.Int("results_count", len(results)),
		zap.Int("requested", numResults))

	return results, nil
}

// ddgTimeFilters maps SearchOptions.TimeRange onto the html endpoint's df
// date-filter parameter.
var ddgTimeFilters = map[string]string{
	"day":   "d",
	"week":  "w",
	"month": "m",
	"year":  "y",
}

func buildDDGSearchURL(query string, opts SearchOptions) string {
	params := url.Values{}
	params.Add("q", query)
	// kl = region. "uk-en" for UK English results
	params.Add("kl", "uk-en")
	if df, ok := ddgTimeFilters[opts.TimeRange]; ok {
		params.Add("df", df)
	}
	return fmt.Sprintf("https://html.duckduckgo.com/html/?%s", params.Encode())
}

// parseHTMLResults extracts search results from DuckDuckGo's HTML response.
//
// The HTML structure of html.duckduckgo.com/html/ contains result blocks like:
//
//	<div class="result results_links results_links_deep web-result">
//	  <div class="links_main links_deep result__body">
//	    <h2 class="result__title">
//	      <a rel="nofollow" class="result__a" href="https://example.com">Title Here</a>
//	    </h2>
//	    <a class="result__snippet" href="...">Snippet text here...</a>
//	    <a class="result__url" href="...">example.com</a>
//	  </div>
//	</div>
//
// We use regex rather than an HTML parser to avoid adding a dependency.
// The structure is simple enough that regex works reliably here.
func (d *DuckDuckGoProvider) parseHTMLResults(html string, maxResults int) []SearchResult {
	var results []SearchResult

	// Pattern to match each result block
	// DuckDuckGo HTML results have class="result__a" for the main link
	resultLinkRe := regexp.MustCompile(`class="result__a"[^>]*href="([^"]*)"[^>]*>([^<]*)`)

	// Pattern for snippets
	snippetRe := regexp.MustCompile(`class="result__snippet"[^>]*>([^<]*)`)

	// Split HTML into result blocks for easier parsing
	// Each result is in a div with class containing "web-result"
	blocks := splitResultBlocks(html)

	for _, block := range blocks {
		if len(results) >= maxResults {
			break
		}

		// Extract link and title
		linkMatch := resultLinkRe.FindStringSubmatch(block)
		if linkMatch == nil || len(linkMatch) < 3 {
			continue
		}

		rawURL := linkMatch[1]
		title := strings.TrimSpace(decodeHTMLEntities(linkMatch[2]))

		// DuckDuckGo wraps URLs in a redirect. Extract the actual URL.
		actualURL := extractDDGURL(rawURL)
		if actualURL == "" {
			continue
		}

		// Skip DuckDuckGo's own pages
		if strings.Contains(actualURL, "duckduckgo.com") {
			continue
		}

		// Extract snippet
		snippet := ""
		snippetMatch := snippetRe.FindStringSubmatch(block)
		if snippetMatch != nil && len(snippetMatch) >= 2 {
			snippet = strings.TrimSpace(decodeHTMLEntities(snippetMatch[1]))
		}

		if title == "" {
			continue
		}

		results = append(results, SearchResult{
			Title:   title,
			URL:     actualURL,
			Snippet: snippet,
			Source:  "duckduckgo",
		})
	}

	return results
}

// splitResultBlocks splits the HTML into individual result blocks
func splitResultBlocks(html string) []string {
	// Split on the result container class
	parts := strings.Split(html, `class="result results_links`)
	if len(parts) <= 1 {
		// Fallback: try splitting on result__body
		parts = strings.Split(html, `class="result__body"`)
	}
	// Skip the first part (before any results)
	if len(parts) > 1 {
		return parts[1:]
	}
	return nil
}

// extractDDGURL extracts the actual URL from DuckDuckGo's redirect wrapper.
// DDG links look like: //duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com&rut=...
// or sometimes they're direct URLs.
func extractDDGURL(rawURL string) string {
	// If it's already a direct URL, return it
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		// But check if it's a DDG redirect
		if !strings.Contains(rawURL, "duckduckgo.com/l/") {
			return rawURL
		}
	}

	// Handle //duckduckgo.com/l/?uddg=ENCODED_URL format
	fullURL := rawURL
	if strings.HasPrefix(rawURL, "//") {
		fullURL = "https:" + rawURL
	}

	parsed, err := url.Parse(fullURL)
	if err != nil {
		return ""
	}

	// Extract the uddg parameter which contains the actual URL
	uddg := parsed.Query().Get("uddg")
	if uddg != "" {
		return uddg
	}

	// If no uddg param and it's not a DDG URL, use as-is
	if !strings.Contains(fullURL, "duckduckgo.com") {
		return fullURL
	}

	return ""
}

// decodeHTMLEntities handles common HTML entities in result text
func decodeHTMLEntities(s string) string {
	replacer := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&#39;", "'",
		"&apos;", "'",
		"&#x27;", "'",
		"&ndash;", "–",
		"&mdash;", "—",
		"&hellip;", "…",
		"&nbsp;", " ",
	)
	return replacer.Replace(s)
}
