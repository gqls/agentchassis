// FILE: platform/orchestration/actions/validate_page_content_placeholder_test.go
//
// The placeholder scan reads PROSE, not code. Its bracket patterns ('[name',
// '[client', '[company') are substrings of idiomatic JavaScript and CSS
// attribute selectors, and every hit is a blocker — so when the scan read the
// whole artefact, a correct self-contained tool failed validation outright.
// The three "convicted JS" cases below are the exact matched snippets from the
// live agent_error_log rows (2026-08-05: idea.uk once, mortgagecalculator
// twice); each cost a tool recreation that completed without saving anything
// (bugs_open/218).
//
// Prose scope comes from datahelpers.ExtractAssertionText — the claims checks'
// existing mechanism in this same file (council REVISE a9ffed15: reuse it, do
// not add a second stripper). One deliberate exception: "<no value>" is a Go
// template artifact that PARSES AWAY as markup, so it must scan the raw
// document or it goes silently inert.
//
// The "still caught" cases are the other half of the contract: scoping to
// prose must not take the rule with it. A genuine template leftover in
// visible HTML stays a blocker.

package actions

import (
	"testing"
)

func placeholderValues(issues []ValidationIssue) []string {
	vals := make([]string, 0, len(issues))
	for _, i := range issues {
		if i.Type == "placeholder_text" {
			vals = append(vals, i.Value)
		}
	}
	return vals
}

func TestPlaceholderScanIgnoresNonProseContexts(t *testing.T) {
	cases := []struct {
		name string
		html string
	}{
		{
			// idea.uk, 2026-08-05 01:00 UTC — CSS attribute selector.
			"attribute selector in querySelector",
			`<html><body><h1>Quiz</h1><script>
			 const el = document.querySelector('input[name="' + KEYS[i] + '"]:checked'); if (!el) return null; a[KEYS[i]] = el.value;
			 </script></body></html>`,
		},
		{
			// mortgagecalculator tool-overpayment, 2026-08-05 12:21 UTC — object indexing.
			"bracket indexing with a variable named name",
			`<html><body><h1>Overpayment</h1><script>
			 const msg = err || ""; if (fields[name]) fields[name].classList.toggle("invalid", !!msg);
			 </script></body></html>`,
		},
		{
			// mortgagecalculator game-fact-finder, 2026-08-05 12:59 UTC — destructuring.
			"array destructuring in a map callback",
			`<html><body><h1>Fact finder</h1><script>
			 el.innerHTML = rows.map(([name, val]) => { const max = CAT_MAX[name]; const pct = clamp((val / max) * 100, 0, 100); return row(name, pct); }).join("");
			 </script></body></html>`,
		},
		{
			"todo comment inside script",
			`<html><body><p>Fine prose.</p><script>// TODO: tune the rounding
			 let x = 1;</script></body></html>`,
		},
		{
			"attribute selector inside style body",
			`<html><head><style>input[name="term"] { width: 4rem; }</style></head><body><p>Fine prose.</p></body></html>`,
		},
		{
			// New coverage the ExtractAssertionText reuse buys over a regex
			// stripper: code samples are examples, not page prose.
			"bracket pattern inside a code sample",
			`<html><body><p>Use the selector like this:</p><pre><code>document.querySelector('input[name="email"]')</code></pre></body></html>`,
		},
		{
			"multiple script blocks all excluded",
			`<html><body><script>a[name] = 1;</script><p>Real content.</p><script>b[client] = 2;</script></body></html>`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := placeholderValues(checkPlaceholderPatterns(tc.html)); len(got) != 0 {
				t.Fatalf("legitimate code convicted as placeholder text: %v", got)
			}
		})
	}
}

func TestPlaceholderScanStillCatchesProse(t *testing.T) {
	cases := []struct {
		name string
		html string
		want string // pattern that must be reported
	}{
		{
			"bracket placeholder in visible text",
			`<html><body><h1>Welcome to [Name]'s mortgage hub</h1><script>let a = 1;</script></body></html>`,
			"[name",
		},
		{
			"your-company placeholder survives adjacent scripts",
			`<html><body><script>let a = 1;</script><p>Contact [your company] today.</p><script>let b = 2;</script></body></html>`,
			"[your ",
		},
		{
			"lorem ipsum in prose",
			`<html><body><p>Lorem ipsum filler that never got replaced.</p></body></html>`,
			"lorem ipsum",
		},
		{
			// A claim split across inline markup still reads as one block, so
			// a placeholder wrapped in <strong> is still caught.
			"placeholder spanning inline markup",
			`<html><body><p>Brought to you by <strong>[Company</strong> Ltd].</p></body></html>`,
			"[company",
		},
		{
			// "<no value>" is markup-shaped and parses away — it must be
			// caught on the raw document, or the pattern is silently inert.
			"unrendered no-value artifact in raw html",
			`<html><body><p>Rates from <no value> percent.</p></body></html>`,
			"<no value>",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := placeholderValues(checkPlaceholderPatterns(tc.html))
			for _, v := range got {
				if v == tc.want {
					return
				}
			}
			t.Fatalf("real placeholder %q not caught; got %v", tc.want, got)
		})
	}
}
