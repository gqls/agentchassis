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
		{
			// JSON-LD is code-shaped: its keys collide with the bracket
			// patterns exactly like JS. headProseBlocks must not read it.
			"json-ld in head is not prose",
			`<html><head><title>Fine title</title><script type="application/ld+json">{"@type":"FAQPage","mainEntity":[{"name":"What is APR?","acceptedAnswer":{"text":"The yearly cost."}}]}</script></head><body><p>Fine prose.</p></body></html>`,
		},
		{
			// A meta tag that is NOT one of the prose names must not be
			// scanned even if its content carries a pattern-shaped string.
			"non-prose meta content ignored",
			`<html><head><meta name="generator" content="builder [name] v2"></head><body><p>Fine prose.</p></body></html>`,
		},
		{
			// Second-person B2B prose. Bare "your company" was removed from
			// the pattern list 2026-08-24: all 46 of its recorded convictions
			// were sentences like these (finetuning.uk was serially blocked on
			// its own ratified proposition). Killed by: re-adding the bare
			// pattern.
			"second-person company prose is not a placeholder",
			`<html><body><h1>Your company's voice, in a model you own</h1><p>Your company data stays private. Fill in the form with your company details.</p></body></html>`,
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
			// The unbracketed residue form that replaced bare "your company".
			"unambiguous company-name prompt still convicts",
			`<html><body><h1>Your Company Name Here</h1><p>Real content.</p></body></html>`,
			"your company name here",
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
		{
			// ExtractAssertionText skips <head>, but the old scan covered it
			// and title text is visitor-visible prose — headProseBlocks keeps it.
			"placeholder in the title element",
			`<html><head><title>[Your Company] Mortgage Tools</title></head><body><p>Fine prose.</p></body></html>`,
			"[your ",
		},
		{
			"placeholder in a meta description, content attribute first",
			`<html><head><meta content="Tools from [Company] for UK buyers" name="description"></head><body><p>Fine prose.</p></body></html>`,
			"[company",
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

// ── Numeric stand-in placeholders (bugs_open/387) ───────────────────────────
//
// "NNN+" reached the public on 2026-08-25: a writer_block quoted the exemplar
// 'Phrase it as "NNN+ AI agents"' and the writer copied it into served copy 14
// times in 137 instructed calls, while no detector had the shape. The positive
// cases pin the live defect's own sentence; the negative cases pin the two
// shapes the fleet census PROVED convict honest copy (bare roman-numeral "XX",
// a quoted "[number]" fill-in template) plus the code idioms ("N+1") and
// lowercase words the case-sensitive match must keep ignoring.

func TestPlaceholderScanCatchesNumericStandIns(t *testing.T) {
	cases := []struct {
		name string
		html string
		want string // exact matched text that must be reported
	}{
		{
			// The live defect, verbatim (model-directory hero, 2026-08-25).
			"the shipped NNN+ hero sentence",
			`<html><body><p>We track providers, context windows and pricing tiers against the NNN+ agent types already running in production.</p></body></html>`,
			"NNN+",
		},
		{
			"two-letter stand-in with plus",
			`<html><body><p>Serving NN+ clients across the UK.</p></body></html>`,
			"NN+",
		},
		{
			"bare NNN as its own word",
			`<html><body><p>This registry tracks NNN agents.</p></body></html>`,
			"NNN",
		},
		{
			"comma-grouped stand-in",
			`<html><body><p>Over N,NNN records verified.</p></body></html>`,
			"N,NNN",
		},
		{
			"stand-in in a meta description is still prose",
			`<html><head><meta name="description" content="Tracks NNN+ agent types in production."></head><body><p>Fine.</p></body></html>`,
			"NNN+",
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
			t.Fatalf("numeric stand-in %q not caught; got %v", tc.want, got)
		})
	}
}

func TestPlaceholderScanIgnoresNumericLookalikes(t *testing.T) {
	cases := []struct {
		name string
		html string
	}{
		{
			// relojistas.com glosario-cronografo — live honest prose the census found.
			"roman numeral century (siglo XX)",
			`<html><body><p>Zenith desarrolló en el siglo XX un cronógrafo de alta frecuencia.</p></body></html>`,
		},
		{
			// idea.uk guide-testing-it — a deliberately quoted fill-in template.
			"quoted fill-in template with [number]",
			`<html><body><p><em>"We'll test it by [this] for [this long]. If fewer than [number] do it, the idea as stated is wrong."</em></p></body></html>`,
		},
		{
			"N+1 is a code idiom, not a stand-in",
			`<html><body><p>Avoiding the classic N+1 queries problem in ORMs.</p></body></html>`,
		},
		{
			"XXL is a size, not a stand-in",
			`<html><body><p>Available in sizes up to XXL.</p></body></html>`,
		},
		{
			// The X-family was cut in council review (round 6cfaa8f0): zero
			// measured stand-in occurrences fleet-wide, and repeated capital
			// letters with '+' have honest uses this census cannot enumerate.
			// This case PINS the non-coverage — re-adding an X shape must
			// bring a fresh census and delete this case knowingly.
			"XXX+ is outside the measured N-family, deliberately",
			`<html><body><h2>XXX+ projects delivered</h2></body></html>`,
		},
		{
			"CNN+ is a brand, not a stand-in",
			`<html><body><p>The streaming service CNN+ closed within a month.</p></body></html>`,
		},
		{
			"lowercase letters never match",
			`<html><body><p>The word cannonball contains nnn... no it does not, but banned and running do carry doubled letters.</p></body></html>`,
		},
		{
			"stand-in inside a script is not prose",
			`<html><body><script>const mask = "NNN+"; render(mask);</script><p>Real copy.</p></body></html>`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := placeholderValues(checkPlaceholderPatterns(tc.html)); len(got) != 0 {
				t.Fatalf("honest content convicted: %v", got)
			}
		})
	}
}
