// FILE: internal/adapters/websearch/search_options_test.go
//
// Regression tests for bugs_open/127: the adapter must pass the request's
// search options through to providers (they were previously unmarshalled,
// logged and dropped), and a provider declining a search type must fall
// through to the next provider rather than end the request.
package websearch

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/internal/adapters/websearch/providers"
	"go.uber.org/zap"
)

type fakeProvider struct {
	name     string
	declines bool
	calls    int
	gotOpts  providers.SearchOptions
}

func (f *fakeProvider) Search(_ context.Context, _ string, _ int, opts providers.SearchOptions) ([]providers.SearchResult, error) {
	f.calls++
	f.gotOpts = opts
	if f.declines {
		return nil, fmt.Errorf("%s: %w", f.name, providers.ErrUnsupportedSearchType)
	}
	return []providers.SearchResult{{Title: "ok", URL: "https://example.com", Source: f.name}}, nil
}

func (f *fakeProvider) Name() string      { return f.name }
func (f *fakeProvider) IsAvailable() bool { return true }

func newTestAdapter(ps ...providers.SearchProvider) *Adapter {
	return &Adapter{
		ctx:             context.Background(),
		logger:          zap.NewNop(),
		providers:       ps,
		primaryProvider: ps[0].Name(),
	}
}

func TestSearchOptionsReachTheProvider(t *testing.T) {
	fake := &fakeProvider{name: "fake"}
	a := newTestAdapter(fake)

	opts := providers.SearchOptions{SearchType: "news", TimeRange: "week"}
	_, providerUsed, _, err := a.performSearchWithFallback("q", 5, "", opts)
	if err != nil {
		t.Fatalf("performSearchWithFallback returned error: %v", err)
	}
	if providerUsed != "fake" {
		t.Fatalf("provider used = %q, want fake", providerUsed)
	}
	if fake.gotOpts.SearchType != opts.SearchType || fake.gotOpts.TimeRange != opts.TimeRange {
		t.Fatalf("provider received opts %+v, want %+v — the options were dropped in the adapter", fake.gotOpts, opts)
	}
}

func TestDeclinedSearchTypeFallsThroughWithoutRetry(t *testing.T) {
	declining := &fakeProvider{name: "declining", declines: true}
	serving := &fakeProvider{name: "serving"}
	a := newTestAdapter(declining, serving)

	results, providerUsed, fallbacks, err := a.performSearchWithFallback("q", 5, "", providers.SearchOptions{SearchType: "news"})
	if err != nil {
		t.Fatalf("performSearchWithFallback returned error: %v", err)
	}
	if providerUsed != "serving" || len(results) != 1 {
		t.Fatalf("provider used = %q with %d results, want serving with 1", providerUsed, len(results))
	}
	if declining.calls != 1 {
		t.Fatalf("declining provider called %d times, want exactly 1 — a decline must not be retried", declining.calls)
	}
	if len(fallbacks) != 1 || fallbacks[0] != "declining" {
		t.Fatalf("fallbacks = %v, want [declining]", fallbacks)
	}
}

func TestAllProvidersDecliningFailsLoudlyNamingTheSearchType(t *testing.T) {
	a := newTestAdapter(
		&fakeProvider{name: "one", declines: true},
		&fakeProvider{name: "two", declines: true},
	)

	_, _, _, err := a.performSearchWithFallback("q", 5, "", providers.SearchOptions{SearchType: "news"})
	if err == nil {
		t.Fatal("want an error when no provider can serve the search type — web results must not be served as news")
	}
	if !strings.Contains(err.Error(), `search_type="news"`) {
		t.Fatalf("error %q does not name the declined search type", err)
	}
}
