package actions

// Tests for extractFieldValue and walkFieldPath (bugs_open/408).
//
// The defect being guarded against is NON-TERMINATION: the old shape held two
// mutually-recursive ".response." fallbacks (strip one occurrence / re-insert
// it) whose conditions were exact inverses, so any path that resolved on no
// form recursed until the goroutine stack passed 1 GB and the runtime killed
// the pod. A plain assert cannot fail on that shape — it crashes the test
// process (exit 2) or hangs before any t.Errorf runs.
//
// ⚠ Every table row runs under a 30s in-test watchdog (council 3918db52,
// guardian seat: the file must not depend on the invoker remembering
// -timeout). A HANG fails via the watchdog; a stack-overflow CRASH (the old
// code's actual mode) kills the process with exit 2, which fails the suite
// loudly on its own. Belt and braces, still run it capped:
//
//	go test -timeout 60s ./platform/orchestration/actions/ -run 'TestExtractFieldValue|TestWalkFieldPath|TestUpstreamDeclaredSkip'
//
// The termination guard is deliberately BEHAVIOURAL (the crash inputs, rows
// "the exact crash input…" and "second pathological shape…"), not a source
// scan for self-calls: this comment necessarily contains the string
// "extractFieldValue(", so a source scan would either match a comment
// vacuously or make comment phrasing load-bearing.

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

// extractWithWatchdog runs extractFieldValue with an in-test deadline so a
// reintroduced hang fails THIS test rather than stalling the whole invocation.
// The goroutine leaks on deadline — acceptable in a test that is already
// failing for the defect the leak evidences.
func extractWithWatchdog(t *testing.T, data map[string]interface{}, path string) string {
	t.Helper()
	done := make(chan string, 1)
	go func() { done <- extractFieldValue(data, path, zap.NewNop()) }()
	select {
	case got := <-done:
		return got
	case <-time.After(30 * time.Second):
		t.Fatalf("extractFieldValue(%q) did not return within 30s — non-termination reintroduced (bugs_open/408)", path)
		return ""
	}
}

func TestExtractFieldValue(t *testing.T) {
	cases := []struct {
		name string
		data map[string]interface{}
		path string
		want string
	}{
		{
			// bugs_open/408 §6 verbatim: the writer skipped, so
			// page_content_0.response.page_html is legitimately absent on
			// every candidate form. Under the old code this row NEVER RETURNS
			// (A strips ".response.", B re-inserts it, for ever) — the
			// -timeout on the suite is what turns that into a red test.
			name: "the exact crash input returns empty",
			data: map[string]interface{}{
				"page_content_0": map[string]interface{}{
					"response": map[string]interface{}{"skipped": true},
				},
			},
			path: "page_content_0.response.page_html",
			want: "",
		},
		{
			name: "direct resolution",
			data: map[string]interface{}{"a": map[string]interface{}{"b": "x"}},
			path: "a.b",
			want: "x",
		},
		{
			name: "resolves only by stripping .response.",
			data: map[string]interface{}{
				"page_content_0": map[string]interface{}{"page_html": "<html>ok</html>"},
			},
			path: "page_content_0.response.page_html",
			want: "<html>ok</html>",
		},
		{
			name: "resolves only by adding .response.",
			data: map[string]interface{}{
				"page_content_0": map[string]interface{}{
					"response": map[string]interface{}{"page_html": "<html>ok</html>"},
				},
			},
			path: "page_content_0.page_html",
			want: "<html>ok</html>",
		},
		{
			// The old recursion tried MORE than three forms on this shape:
			// a.b.response.c → strip → a.b.c → add after first segment →
			// a.response.b.c, and terminated if any resolved. A flat
			// three-candidate list drops this route; the loop builder keeps it.
			name: "deep .response. occurrence resolves via the re-add form",
			data: map[string]interface{}{
				"a": map[string]interface{}{
					"response": map[string]interface{}{
						"b": map[string]interface{}{"c": "deep"},
					},
				},
			},
			path: "a.b.response.c",
			want: "deep",
		},
		{
			name: "terminal map yields a content key",
			data: map[string]interface{}{"step": map[string]interface{}{"result": "the content"}},
			path: "step",
			want: "the content",
		},
		{
			// contentKeys order: "result" outranks "html".
			name: "terminal map content key priority",
			data: map[string]interface{}{
				"step": map[string]interface{}{"html": "second", "result": "first"},
			},
			path: "step",
			want: "first",
		},
		{
			// Candidate 1 RESOLVES (to a map with no content key), so the
			// tail's contentKeys miss must return "" without consulting the
			// add-form candidate — which here would resolve to "trap".
			// Fallbacks fire on walk failure only, never on a tail miss.
			name: "contentKeys miss does not consult further candidates",
			data: map[string]interface{}{
				"a": map[string]interface{}{
					"b":        map[string]interface{}{"status": "ok"},
					"response": map[string]interface{}{"b": map[string]interface{}{"result": "trap"}},
				},
			},
			path: "a.b",
			want: "",
		},
		{
			name: "single-part missing path",
			data: map[string]interface{}{"x": "y"},
			path: "zzz",
			want: "",
		},
		{
			// Preserved short-circuit: a non-map met mid-path ends the lookup
			// outright — no fallback candidate is tried after it, matching the
			// old default branch. If the walk treated this like a missing key,
			// the add-form would resolve to "val" and this row would fail.
			name: "non-map mid-path short-circuits without trying candidates",
			data: map[string]interface{}{
				"a": map[string]interface{}{
					"b":        "str",
					"response": map[string]interface{}{"b": map[string]interface{}{"c": "val"}},
				},
			},
			path: "a.b.c",
			want: "",
		},
		{
			name: "terminal non-string returns empty",
			data: map[string]interface{}{"a": map[string]interface{}{"b": 42}},
			path: "a.b",
			want: "",
		},
		{
			// Second pathological non-terminating shape: under the old code
			// "a.b.response.c" against {"a": {}} looped between a.b.c and
			// a.response.b.c for ever. Guards the multi-form route.
			name: "second pathological shape terminates empty",
			data: map[string]interface{}{"a": map[string]interface{}{}},
			path: "a.b.response.c",
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractWithWatchdog(t, tc.data, tc.path)
			if got != tc.want {
				t.Errorf("extractFieldValue(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestUpstreamDeclaredSkip(t *testing.T) {
	cases := []struct {
		name string
		data map[string]interface{}
		want bool
	}{
		{
			// The production shape from bugs_open/408 §4: the writer skipped
			// and said so — the legitimate quiet-skip case, no error row owed.
			name: "skip declared under response",
			data: map[string]interface{}{
				"page_content_0": map[string]interface{}{
					"response": map[string]interface{}{"skipped": true},
				},
			},
			want: true,
		},
		{
			name: "skip declared at top level",
			data: map[string]interface{}{
				"page_content_0": map[string]interface{}{"skipped": true},
			},
			want: true,
		},
		{
			// Upstream produced a result and declared nothing — an
			// unresolvable content_field here is the misconfiguration
			// signature and must be counted (ASSEMBLE_CONTENT_FIELD_UNRESOLVED).
			name: "no declaration means content was expected",
			data: map[string]interface{}{
				"page_content_0": map[string]interface{}{
					"response": map[string]interface{}{"page_html": "<html>x</html>"},
				},
			},
			want: false,
		},
		{
			name: "skipped false is not a declaration",
			data: map[string]interface{}{
				"page_content_0": map[string]interface{}{
					"response": map[string]interface{}{"skipped": false},
				},
			},
			want: false,
		},
		{
			name: "missing step root",
			data: map[string]interface{}{},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := upstreamDeclaredSkip(tc.data, "page_content_0.response.page_html")
			if got != tc.want {
				t.Errorf("upstreamDeclaredSkip = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestWalkFieldPath(t *testing.T) {
	data := map[string]interface{}{
		"a": map[string]interface{}{
			"b": "x",
			"n": 7,
		},
	}

	t.Run("resolved", func(t *testing.T) {
		value, stoppedAt, outcome := walkFieldPath(data, "a.b")
		if outcome != pathWalkResolved || stoppedAt != "" || value != "x" {
			t.Errorf("got (%v, %q, %v), want (x, \"\", resolved)", value, stoppedAt, outcome)
		}
	})

	t.Run("key missing names the segment", func(t *testing.T) {
		_, stoppedAt, outcome := walkFieldPath(data, "a.missing")
		if outcome != pathWalkKeyMissing || stoppedAt != "missing" {
			t.Errorf("got (%q, %v), want (missing, keyMissing)", stoppedAt, outcome)
		}
	})

	t.Run("non-map mid-path names the segment", func(t *testing.T) {
		_, stoppedAt, outcome := walkFieldPath(data, "a.n.deeper")
		if outcome != pathWalkNotTraversable || stoppedAt != "deeper" {
			t.Errorf("got (%q, %v), want (deeper, notTraversable)", stoppedAt, outcome)
		}
	})
}
