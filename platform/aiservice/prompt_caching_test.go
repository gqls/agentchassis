// FILE: platform/aiservice/prompt_caching_test.go
//
// Wire-shape tests for the opt-in cache breakpoint (CacheBreakpointMarker).
// Every assertion here is against the request BODY, for the reason vision_test
// gives: a client that silently dropped cache_control would pass any
// response-only test while quietly paying full price on every call — which is
// exactly the failure this seam exists to prevent, and it is invisible from
// the outside because the answers keep coming back correct.
package aiservice

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// cacheUsageBody carries the two cache counters a real cached response
// returns, so the write-back assertions below have something to read.
const cacheUsageBody = `{
	"content": [{"type": "text", "text": "a verdict"}],
	"stop_reason": "end_turn",
	"usage": {
		"input_tokens": 11,
		"cache_creation_input_tokens": 64700,
		"cache_read_input_tokens": 0,
		"output_tokens": 5
	}
}`

func cachingClient(body string) (*AnthropicClient, *capturingTransport) {
	tr := &capturingTransport{body: body}
	return &AnthropicClient{
		apiKey:     "test-key",
		model:      "claude-test",
		httpClient: &http.Client{Transport: tr},
	}, tr
}

// userContent digs the user message's content out of a captured request body.
// Returns the raw value because its TYPE is the thing under test: a plain
// string means "unchanged legacy path", an array means "split into blocks".
func userContent(t *testing.T, captured []byte) interface{} {
	t.Helper()
	var req struct {
		Messages []struct {
			Role    string      `json:"role"`
			Content interface{} `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(captured, &req); err != nil {
		t.Fatalf("captured body is not valid JSON: %v", err)
	}
	if len(req.Messages) != 1 {
		t.Fatalf("expected exactly 1 message, got %d", len(req.Messages))
	}
	return req.Messages[0].Content
}

// THE NEGATIVE CONTROL, and the most important test in this file. Every other
// agent in the fleet goes down this path unchanged. If this ever fails, the
// blast radius is not "caching broke", it is "every prompt shape changed".
func TestNoMarkerSendsPlainStringUnchanged(t *testing.T) {
	c, tr := cachingClient(cacheUsageBody)

	const prompt = "a perfectly ordinary prompt with no marker in it"
	if _, err := c.GenerateText(context.Background(), prompt, map[string]interface{}{}); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}

	got := userContent(t, tr.captured)
	s, ok := got.(string)
	if !ok {
		t.Fatalf("unmarked prompt must stay a plain string, got %T — this changes the wire shape for EVERY agent", got)
	}
	if s != prompt {
		t.Errorf("prompt altered: want %q, got %q", prompt, s)
	}
	if strings.Contains(string(tr.captured), "cache_control") {
		t.Error("cache_control present on an unmarked prompt — the opt-in default is not OFF")
	}
}

func TestMarkerSplitsIntoTwoBlocksWithCacheControlOnFirstOnly(t *testing.T) {
	c, tr := cachingClient(cacheUsageBody)

	prompt := "SHARED EVIDENCE" + CacheBreakpointMarker + "SEAT INSTRUCTION"
	if _, err := c.GenerateText(context.Background(), prompt, map[string]interface{}{}); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}

	blocks, ok := userContent(t, tr.captured).([]interface{})
	if !ok {
		t.Fatalf("marked prompt must become a block array, got %T", userContent(t, tr.captured))
	}
	if len(blocks) != 2 {
		t.Fatalf("want 2 blocks (shared, per-call), got %d", len(blocks))
	}

	first, _ := blocks[0].(map[string]interface{})
	second, _ := blocks[1].(map[string]interface{})

	if first["text"] != "SHARED EVIDENCE" {
		t.Errorf("first block text: want %q, got %q", "SHARED EVIDENCE", first["text"])
	}
	if second["text"] != "SEAT INSTRUCTION" {
		t.Errorf("second block text: want %q, got %q", "SEAT INSTRUCTION", second["text"])
	}
	// The marker itself must not survive into either block — it is a
	// boundary, not content, and a model that saw it would be reading
	// plumbing as instruction.
	for i, b := range []map[string]interface{}{first, second} {
		if txt, _ := b["text"].(string); strings.Contains(txt, CacheBreakpointMarker) {
			t.Errorf("block %d still contains the marker: %q", i, txt)
		}
	}

	cc, ok := first["cache_control"].(map[string]interface{})
	if !ok {
		t.Fatal("first block has no cache_control — nothing would ever be cached")
	}
	if cc["type"] != "ephemeral" {
		t.Errorf("cache_control.type: want ephemeral, got %v", cc["type"])
	}
	if cc["ttl"] != cacheTTL {
		t.Errorf("cache_control.ttl: want %q, got %v", cacheTTL, cc["ttl"])
	}

	// THE POINT OF THE WHOLE CHANGE. A breakpoint on the varying block
	// writes a distinct cache entry per call and reads nothing back — it
	// costs MORE than no caching at all, and looks identical in every log.
	if _, present := second["cache_control"]; present {
		t.Error("cache_control on the per-call block — every call would write its own entry and read none")
	}
}

// A marker at position 0 has no prefix to cache. Sending an empty text block
// is a 400 from the API, so the client must fall back to the legacy path
// rather than construct something the API rejects.
func TestMarkerAtStartFallsBackToPlainString(t *testing.T) {
	c, tr := cachingClient(cacheUsageBody)

	prompt := CacheBreakpointMarker + "everything is per-call"
	if _, err := c.GenerateText(context.Background(), prompt, map[string]interface{}{}); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}

	if _, ok := userContent(t, tr.captured).(string); !ok {
		t.Error("marker with an empty prefix must fall back to the plain-string path, not emit an empty cached block")
	}
}

// A marker at the very end means "cache all of it" — legitimate, and the
// second block must be omitted rather than sent empty.
func TestMarkerAtEndEmitsSingleCachedBlock(t *testing.T) {
	c, tr := cachingClient(cacheUsageBody)

	prompt := "cache every last byte of this" + CacheBreakpointMarker
	if _, err := c.GenerateText(context.Background(), prompt, map[string]interface{}{}); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}

	blocks, ok := userContent(t, tr.captured).([]interface{})
	if !ok {
		t.Fatalf("want a block array, got %T", userContent(t, tr.captured))
	}
	if len(blocks) != 1 {
		t.Fatalf("trailing marker must emit exactly 1 block, got %d (an empty text block is a 400)", len(blocks))
	}
	first, _ := blocks[0].(map[string]interface{})
	if _, ok := first["cache_control"]; !ok {
		t.Error("the sole block must carry cache_control")
	}
}

// Without these counters the saving is unmeasurable: input_tokens alone is the
// UNCACHED REMAINDER, so a cached call reports ~5k where the real prompt was
// ~100k, and any cost report built on it silently understates by ~95%.
func TestCacheUsageCountersReachOptions(t *testing.T) {
	c, _ := cachingClient(cacheUsageBody)

	options := map[string]interface{}{}
	if _, err := c.GenerateText(context.Background(), "p"+CacheBreakpointMarker+"q", options); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}

	if got := options["__usage_cache_creation_input_tokens"]; got != 64700 {
		t.Errorf("cache_creation write-back: want 64700, got %v", got)
	}
	if got, present := options["__usage_cache_read_input_tokens"]; !present {
		t.Error("cache_read key missing entirely — a key that appears only on hits cannot distinguish 'no cache used' from 'build predates caching'")
	} else if got != 0 {
		t.Errorf("cache_read write-back: want 0, got %v", got)
	}
}

// The zeros must be written on the ordinary path too, for the reason above:
// their ABSENCE is what tells an operator the binary is too old, and that
// signal only exists if a modern binary always writes them.
func TestCacheCountersWrittenEvenWhenUncached(t *testing.T) {
	c, _ := cachingClient(anthropicOKBody) // no cache fields in this response at all

	options := map[string]interface{}{}
	if _, err := c.GenerateText(context.Background(), "no marker here", options); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}

	for _, k := range []string{"__usage_cache_creation_input_tokens", "__usage_cache_read_input_tokens"} {
		v, present := options[k]
		if !present {
			t.Errorf("%s not written on an uncached call", k)
			continue
		}
		if v != 0 {
			t.Errorf("%s: want 0 on an uncached call, got %v", k, v)
		}
	}
}
