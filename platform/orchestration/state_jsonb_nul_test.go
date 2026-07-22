package orchestration

import (
	"bytes"
	"encoding/json"
	"testing"
)

// bugs_open/056: Postgres jsonb rejects exactly one Unicode escape — \u0000 —
// so a single NUL byte anywhere in an orchestration's state failed the whole
// UPDATE (22P05) and silently killed the run. These pin the boundary guard.
func TestSanitiseJSONBNulEscapes(t *testing.T) {
	t.Run("genuine escape is replaced and stays valid JSON", func(t *testing.T) {
		in := []byte(`{"k":"a\u0000b"}`)
		out := sanitiseJSONBNulEscapes(in)
		want := []byte(`{"k":"a\ufffdb"}`)
		if !bytes.Equal(out, want) {
			t.Fatalf("got %s, want %s", out, want)
		}
		if !json.Valid(out) {
			t.Fatalf("output is not valid JSON: %s", out)
		}
	})

	// A diagnosis quoting this very escape from a doc or a source file
	// produces literal backslash-u0000 TEXT (marshalled as an escaped
	// backslash followed by "u0000"); replacing that would corrupt content
	// that was never a NUL.
	t.Run("literal backslash-u0000 text is preserved", func(t *testing.T) {
		in := []byte(`{"k":"a\\u0000b"}`)
		out := sanitiseJSONBNulEscapes(in)
		if !bytes.Equal(out, in) {
			t.Fatalf("literal text was mangled: got %s, want %s", out, in)
		}
	})

	t.Run("only even-backslash occurrences are escapes", func(t *testing.T) {
		in := []byte(`{"a\u0000b":"\\u0000 then \u0000 and \\\u0000"}`)
		want := []byte(`{"a\ufffdb":"\\u0000 then \ufffd and \\\ufffd"}`)
		out := sanitiseJSONBNulEscapes(in)
		if !bytes.Equal(out, want) {
			t.Fatalf("got %s, want %s", out, want)
		}
		if !json.Valid(out) {
			t.Fatalf("output is not valid JSON: %s", out)
		}
	})

	t.Run("no match returns the input slice itself", func(t *testing.T) {
		in := []byte(`{"k":"clean"}`)
		out := sanitiseJSONBNulEscapes(in)
		if &in[0] != &out[0] {
			t.Fatal("clean input should pass through without copying")
		}
	})

	// The exact 056 shape: a map keyed with a NUL delimiter (as the old
	// CodeRequestKey built) must come out jsonb-safe even if some future
	// code reintroduces a NUL source.
	t.Run("a NUL-delimited map key survives the boundary", func(t *testing.T) {
		m, err := json.Marshal(map[string]bool{"symbol\x00generatetext": true})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if !bytes.Contains(m, []byte(`\u0000`)) {
			t.Fatalf("precondition failed: marshal should emit the escape, got %s", m)
		}
		out := sanitiseJSONBNulEscapes(m)
		if bytes.Contains(out, []byte(`\u0000`)) {
			t.Fatalf("jsonb-fatal escape survived the boundary: %s", out)
		}
		if !json.Valid(out) {
			t.Fatalf("output is not valid JSON: %s", out)
		}
	})
}
