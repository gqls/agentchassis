package aiservice

import (
	"strings"
	"testing"
)

// The fingerprint replaced a capped text excerpt, so these tests assert the two
// properties that justified the swap: it must answer the diagnostic questions
// the excerpt answered, AND it must not reproduce content.

func TestFingerprint_EmitsNoContent(t *testing.T) {
	// A response that quotes the visitor back — the exposure the excerpt had.
	body := `Looking at your argument that AI art is expensive plagiarism, the ` +
		`stronger reading is that the tool changed and the intent did not. ` +
		`{"verdict":"user wins","reasons":"the defence held"}`

	got := Fingerprint(body)

	for _, leaked := range []string{"plagiarism", "argument", "verdict", "user wins", "defence"} {
		if strings.Contains(got, leaked) {
			t.Errorf("fingerprint leaked content %q\ngot: %s", leaked, got)
		}
	}
}

// The load-bearing case. bugs_closed/088: a complete object, then commentary,
// then a SECOND complete object. The old 300-char excerpt could not see object
// two — it starts ~1,500 chars in. This must find it at any distance.
func TestFingerprint_FindsTheSecondObjectBeyondAnyExcerptCap(t *testing.T) {
	first := `{"verdict":"user wins","reasons":"` + strings.Repeat("a genuine and lengthy reason. ", 60) + `"}`
	body := first + "\n\nActually, on reflection I should correct that:\n\n" +
		`{"verdict":"opponent wins","reasons":"revised"}`

	if len(first) < 1000 {
		t.Fatalf("test setup too short to be discriminating: first object is %d chars", len(first))
	}

	got := Fingerprint(body)
	if !strings.Contains(got, "objects=2") {
		t.Errorf("the double-JSON case must be detected regardless of length\ngot: %s", got)
	}
}

func TestFingerprint_DistinguishesTheFailureShapes(t *testing.T) {
	cases := []struct {
		name, body string
		want       []string
	}{
		{"clean object", `{"verdict":"x","reasons":"y"}`,
			[]string{"first={", "fence=no", "objects=1", "parses=true", "keys=[reasons,verdict]"}},
		{"prose wrapper", `Sure! Here you go: {"verdict":"x"}`,
			[]string{"first=S", "objects=1", "parses=false"}},
		{"markdown fence", "```json\n{\"verdict\":\"x\"}\n```",
			[]string{"fence=yes", "parses=false"}},
		{"empty completion", "",
			[]string{"chars=0", "first=none", "objects=0", "parses=false"}},
		{"whitespace only", "   \n\t ",
			[]string{"first=none", "objects=0"}},
	}

	for _, c := range cases {
		got := Fingerprint(c.body)
		for _, want := range c.want {
			if !strings.Contains(got, want) {
				t.Errorf("%s: missing %q\ngot: %s", c.name, want, got)
			}
		}
	}
}

// Without string-awareness a brace or an escaped quote inside prose miscounts,
// and the count is the entire value of this function.
func TestTopLevelJSONObjects_IgnoresBracesInsideStrings(t *testing.T) {
	cases := []struct {
		name, body string
		want       int
	}{
		{"brace inside a string value", `{"reasons":"they wrote { and } at me"}`, 1},
		{"escaped quote then brace", `{"reasons":"he said \"{\" loudly"}`, 1},
		{"nested objects count once", `{"a":{"b":{"c":1}}}`, 1},
		{"two separate objects", `{"a":1} {"b":2}`, 2},
		{"stray closing brace in prose", `oh no } {"a":1}`, 1},
		{"no objects at all", `I refuse to answer that.`, 0},
		{"unterminated object", `{"a":1`, 0},
	}

	for _, c := range cases {
		if got := TopLevelJSONObjects(c.body); got != c.want {
			t.Errorf("%s: got %d objects, want %d", c.name, got, c.want)
		}
	}
}

func TestFingerprint_CapsAnAbsurdKey(t *testing.T) {
	long := strings.Repeat("k", 200)
	got := Fingerprint(`{"` + long + `":1}`)

	if strings.Contains(got, long) {
		t.Errorf("an oversized key must be capped, not emitted whole\ngot: %s", got)
	}
	if !strings.Contains(got, "~") {
		t.Errorf("a capped key should be marked as truncated\ngot: %s", got)
	}
}
