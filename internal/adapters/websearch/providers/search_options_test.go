// FILE: internal/adapters/websearch/providers/search_options_test.go
//
// Regression tests for bugs_open/127: search_type/time_range must reach each
// provider's API, and a provider that cannot honour a search type must
// decline rather than serve mislabelled web results.
package providers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNormalisePublishedAt(t *testing.T) {
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"rfc3339 passthrough", "2026-07-27T09:00:00Z", "2026-07-27T09:00:00Z"},
		{"iso with millis passthrough", "2023-07-11T09:23:18.283Z", "2023-07-11T09:23:18.283Z"},
		{"bare date", "2026-07-20", "2026-07-20T00:00:00Z"},
		{"relative hours", "3 hours ago", "2026-07-28T09:00:00Z"},
		{"relative days", "2 days ago", "2026-07-26T12:00:00Z"},
		{"relative months", "3 months ago", "2026-04-29T12:00:00Z"},
		{"relative singular", "1 hour ago", "2026-07-28T11:00:00Z"},
		{"yesterday", "yesterday", "2026-07-27T12:00:00Z"},
		{"unrecognised returned unchanged", "sometime last century", "sometime last century"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalisePublishedAt(tc.in, now); got != tc.want {
				t.Fatalf("normalisePublishedAt(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFirecrawlNewsSearchSendsSourcesAndTbs(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Errorf("request body is not JSON: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"success":true,"data":{"news":[
			{"title":"Firm acquired","url":"https://example.com/a","snippet":"deal closed","date":"2 days ago","position":1}
		]}}`)
	}))
	defer srv.Close()

	p := &FirecrawlProvider{apiKey: "test", httpClient: srv.Client(), apiURL: srv.URL, logger: zap.NewNop()}
	results, err := p.Search(context.Background(), "acquisitions", 5, SearchOptions{SearchType: "news", TimeRange: "week"})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	sources, ok := gotBody["sources"].([]interface{})
	if !ok || len(sources) != 1 || sources[0] != "news" {
		t.Fatalf("request sources = %v, want [news]", gotBody["sources"])
	}
	if gotBody["tbs"] != "qdr:w" {
		t.Fatalf("request tbs = %v, want qdr:w", gotBody["tbs"])
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Snippet != "deal closed" {
		t.Fatalf("snippet = %q, want the news item's snippet field", results[0].Snippet)
	}
	if _, err := time.Parse(time.RFC3339, results[0].PublishedAt); err != nil {
		t.Fatalf("PublishedAt %q is not RFC3339 — the feed writer would store NULL", results[0].PublishedAt)
	}
}

func TestFirecrawlWebSearchOmitsSourcesAndTbs(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"success":true,"data":{"web":[{"title":"T","url":"https://example.com","description":"D"}]}}`)
	}))
	defer srv.Close()

	p := &FirecrawlProvider{apiKey: "test", httpClient: srv.Client(), apiURL: srv.URL, logger: zap.NewNop()}
	if _, err := p.Search(context.Background(), "anything", 5, SearchOptions{}); err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if _, present := gotBody["sources"]; present {
		t.Fatalf("web search must not send sources, got %v", gotBody["sources"])
	}
	if _, present := gotBody["tbs"]; present {
		t.Fatalf("web search without time_range must not send tbs, got %v", gotBody["tbs"])
	}
}

func TestFirecrawlDeclinesUnparseableSearchType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("provider must decline before making an HTTP request")
	}))
	defer srv.Close()

	p := &FirecrawlProvider{apiKey: "test", httpClient: srv.Client(), apiURL: srv.URL, logger: zap.NewNop()}
	_, err := p.Search(context.Background(), "q", 5, SearchOptions{SearchType: "images"})
	if !errors.Is(err, ErrUnsupportedSearchType) {
		t.Fatalf("err = %v, want ErrUnsupportedSearchType", err)
	}
}

func TestScrapingBeeNewsSearchSendsSearchTypeAndTbs(t *testing.T) {
	var gotQuery map[string][]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"news_results":[
			{"title":"Rate cut announced","link":"https://example.com/n","snippet":"today","date":"2026-07-27T09:23:18.283Z","source":"Example"}
		]}`)
	}))
	defer srv.Close()

	p := &ScrapingBeeProvider{apiKey: "test", httpClient: srv.Client(), apiURL: srv.URL, logger: zap.NewNop()}
	results, err := p.Search(context.Background(), "rates", 5, SearchOptions{SearchType: "news", TimeRange: "day"})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if got := gotQuery["search_type"]; len(got) != 1 || got[0] != "news" {
		t.Fatalf("search_type param = %v, want [news]", got)
	}
	if got := gotQuery["tbs"]; len(got) != 1 || got[0] != "qdr:d" {
		t.Fatalf("tbs param = %v, want [qdr:d]", got)
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1 (news_results must be parsed)", len(results))
	}
	if _, err := time.Parse(time.RFC3339, results[0].PublishedAt); err != nil {
		t.Fatalf("PublishedAt %q is not RFC3339", results[0].PublishedAt)
	}
}

func TestScrapingBeeWebSearchOmitsSearchType(t *testing.T) {
	var gotQuery map[string][]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"organic_results":[{"title":"T","link":"https://example.com","snippet":"S"}]}`)
	}))
	defer srv.Close()

	p := &ScrapingBeeProvider{apiKey: "test", httpClient: srv.Client(), apiURL: srv.URL, logger: zap.NewNop()}
	if _, err := p.Search(context.Background(), "anything", 5, SearchOptions{}); err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if got, present := gotQuery["search_type"]; present {
		t.Fatalf("web search must not send search_type, got %v", got)
	}
}

func TestDuckDuckGoDeclinesNews(t *testing.T) {
	p := NewDuckDuckGoProvider(&http.Client{}, zap.NewNop())
	_, err := p.Search(context.Background(), "q", 5, SearchOptions{SearchType: "news"})
	if !errors.Is(err, ErrUnsupportedSearchType) {
		t.Fatalf("err = %v, want ErrUnsupportedSearchType — the html endpoint has no news vertical", err)
	}
}

func TestDuckDuckGoTimeRangeMapsToDf(t *testing.T) {
	withRange := buildDDGSearchURL("q", SearchOptions{TimeRange: "week"})
	if !strings.Contains(withRange, "df=w") {
		t.Fatalf("URL %q does not carry df=w", withRange)
	}
	without := buildDDGSearchURL("q", SearchOptions{})
	if strings.Contains(without, "df=") {
		t.Fatalf("URL %q must not carry a df filter when no time_range is set", without)
	}
}
