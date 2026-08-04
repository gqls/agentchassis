package discovery_checks

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func TestScanPlainTextMarkdown_Positives(t *testing.T) {
	cases := []struct {
		name string
		text string
		want []string // expected pattern names, in any order
	}{
		{
			name: "bold from the bug's own founding row",
			text: "Banks evaluate your application using a **Decision Engine** (an automated algorithm that grades your financial history).",
			want: []string{"bold"},
		},
		{
			name: "bold trailing colon",
			text: "**Recommended next steps:**",
			want: []string{"bold"},
		},
		{
			name: "bold AND code span in the same value — a **-only fix must not miss this",
			text: "**the `animation`**",
			want: []string{"bold", "code_span"},
		},
		{
			name: "heading",
			text: "Intro\n## Why choose us\nBody",
			want: []string{"heading"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := scanPlainTextMarkdown(tc.text, "slot", "field", "content_data", true)
			gotPatterns := map[string]bool{}
			for _, f := range got {
				gotPatterns[f.Pattern] = true
			}
			for _, w := range tc.want {
				if !gotPatterns[w] {
					t.Errorf("expected pattern %q to fire on %q, got %+v", w, tc.text, got)
				}
			}
			if len(gotPatterns) != len(tc.want) {
				t.Errorf("expected exactly %v, got %+v", tc.want, got)
			}
		})
	}
}

func TestScanPlainTextMarkdown_GuardNegatives(t *testing.T) {
	// None of these are markdown syntax and must produce zero findings —
	// the letter-guard discipline the bug filer's own query already used.
	cases := []string{
		"3 * 4 = 12",
		"2**10 is 1024",
		"a ** b",
		"color: #fff",
		"#1 rated service in Leeds",
		"issue #12 closed",
		"`${url}` interpolation",
		"`/api` path",
	}
	for _, text := range cases {
		t.Run(text, func(t *testing.T) {
			got := scanPlainTextMarkdown(text, "slot", "field", "content_data", true)
			if len(got) != 0 {
				t.Errorf("expected zero findings on %q, got %+v", text, got)
			}
		})
	}
}

func TestScanPlainTextMarkdown_CodeSpanSuppressedOnMarkup(t *testing.T) {
	// A value carrying HTML markup is not a text-typed field; backticks there
	// are code, not prose, so the code_span pattern must be suppressed —
	// but bold still fires if present.
	text := "<p>Use the **animation** `helper` function</p>"
	got := scanPlainTextMarkdown(text, "slot", "field", "content_data", !htmlMarkupRe.MatchString(text))
	patterns := map[string]bool{}
	for _, f := range got {
		patterns[f.Pattern] = true
	}
	if patterns["code_span"] {
		t.Errorf("code_span must be suppressed on a markup-bearing value, got %+v", got)
	}
	if !patterns["bold"] {
		t.Errorf("bold must still fire on a markup-bearing value, got %+v", got)
	}
}

func TestExtractAssertionText_ScriptBackticksInvisible(t *testing.T) {
	clean := `<html><body><script>const t = ` + "`x`" + `;</script><p>Clean prose with no markup issues.</p></body></html>`
	for _, block := range datahelpers.ExtractAssertionText(clean) {
		if got := scanPlainTextMarkdown(block, "slot", "", "rendered_html", true); len(got) != 0 {
			t.Errorf("script backticks must not surface via ExtractAssertionText, got %+v from block %q", got, block)
		}
	}

	dirty := `<html><body><script>const t = ` + "`x`" + `;</script><p>Prose with **bold** in it.</p></body></html>`
	var found []literalMarkdownFinding
	for _, block := range datahelpers.ExtractAssertionText(dirty) {
		found = append(found, scanPlainTextMarkdown(block, "slot", "", "rendered_html", true)...)
	}
	if len(found) != 1 || found[0].Pattern != "bold" {
		t.Errorf("expected exactly one bold finding from the <p> text, got %+v", found)
	}
}

func TestWalkContentDataStrings(t *testing.T) {
	data := map[string]interface{}{
		"headline":  "**bold**",
		"_built_at": "2026-08-03T00:00:00Z", // platform metadata — must be skipped
		"items": []interface{}{
			map[string]interface{}{"label": "one"},
			map[string]interface{}{"label": "**two**"},
		},
	}
	var visited []string
	walkContentDataStrings("", data, func(path, s string) {
		visited = append(visited, path+"="+s)
	})

	sawHeadline, sawBuiltAt, sawItemTwo := false, false, false
	for _, v := range visited {
		if v == "headline=**bold**" {
			sawHeadline = true
		}
		if v == "_built_at=2026-08-03T00:00:00Z" {
			sawBuiltAt = true
		}
		if v == "items[1].label=**two**" {
			sawItemTwo = true
		}
	}
	if !sawHeadline {
		t.Errorf("expected to visit headline, got %v", visited)
	}
	if sawBuiltAt {
		t.Errorf("expected _built_at (platform metadata) to be skipped, got %v", visited)
	}
	if !sawItemTwo {
		t.Errorf("expected to visit nested array item field, got %v", visited)
	}
}
