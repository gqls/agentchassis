// FILE: platform/orchestration/actions/ndjson_feed_test.go
//
// scanInternalNDJSONFeed now serves a LIVE path (intent-collector) as well as
// the new island pull, so its two failure branches are what matter: a
// transport/status failure must abort the caller's site attempt, while a
// stream that dies mid-body must be reported as PARTIAL — the lines already
// persisted stand and the caller resumes from its own checkpoint. Conflating
// the two either loses data or re-runs work.

package actions

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func collectLines(t *testing.T, srv *httptest.Server, key string) ([]string, error) {
	t.Helper()
	var got []string
	err := scanInternalNDJSONFeed(context.Background(), srv.URL, key, 5*time.Second,
		func(line []byte) { got = append(got, string(line)) })
	return got, err
}

func TestScanFeedDeliversLinesAndSendsKey(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Internal-Key")
		w.Write([]byte("{\"id\":\"a\"}\n{\"id\":\"b\"}\n{\"_meta\":{\"n\":2}}\n"))
	}))
	defer srv.Close()

	got, err := collectLines(t, srv, "secret-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotKey != "secret-key" {
		t.Errorf("shared secret not sent: got %q", gotKey)
	}
	// The _meta trailer is delivered too — filtering it is the caller's job,
	// and both callers do it by shape, not by position.
	if len(got) != 3 || got[0] != `{"id":"a"}` {
		t.Errorf("lines not delivered verbatim: %v", got)
	}
}

func TestScanFeedNon200IsFatalNotPartial(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusUnauthorized)
	}))
	defer srv.Close()

	got, err := collectLines(t, srv, "wrong-key")
	if err == nil {
		t.Fatal("401 did not produce an error")
	}
	var partial *partialStreamError
	if errors.As(err, &partial) {
		t.Error("401 must be fatal, not partial — nothing was delivered")
	}
	if len(got) != 0 {
		t.Errorf("lines delivered despite 401: %v", got)
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should name the status: %v", err)
	}
}

func TestScanFeedOversizedLineIsPartialNotSilentTruncation(t *testing.T) {
	// A line past the buffer ceiling must surface as a partial stream, never
	// as a clean short read that looks like a complete feed.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"id\":\"a\"}\n"))
		w.Write([]byte(strings.Repeat("x", ndjsonScanBufMax+1) + "\n"))
	}))
	defer srv.Close()

	got, err := collectLines(t, srv, "k")
	if err == nil {
		t.Fatal("oversized line reported success — a truncated feed looked complete")
	}
	var partial *partialStreamError
	if !errors.As(err, &partial) {
		t.Errorf("oversized line should be a partialStreamError, got %T: %v", err, err)
	}
	if len(got) != 1 {
		t.Errorf("lines before the bad one should still be delivered: %v", got)
	}
}
