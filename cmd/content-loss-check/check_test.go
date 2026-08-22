// FILE: cmd/content-loss-check/check_test.go
//
// The pure predicates, mutation-proven at authoring time (mutations and their
// failing tests named in the commit):
//
//	M1 isNonLLMLoss: drop the llm-prefix exclusion → TestLossDefinition fails
//	   on the "llm prose rewrite" case
//	M2 isNonLLMLoss: drop the blank-before guard → TestLossDefinition fails on
//	   "nothing held, nothing lost"
//	M3 healVerdict: return "healed" whenever deployedHoldsKey OR rows==0 →
//	   TestHealVerdict fails on the row-gone case (row_gone is NOT healed —
//	   the two write different resolved_by stamps and mean different things)
//	M4 carryMissVerdict: resolve when ANY field healed → TestCarryMissVerdict
//	   fails on the partially-healed case
//
// The SQL side is NOT tested here — it is instrumented at RUNTIME instead:
// the pinned demand control must re-find the known pre-fix funnel losses on
// every run, and a zero exits 2 (refusal), never 0. A unit test against a mock
// cannot prove a query against a live schema; the control can, and does so on
// the same run whose cleanliness it licenses.
package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestLossDefinition(t *testing.T) {
	cases := []struct {
		name                  string
		source, before, after string
		want                  bool
	}{
		{"resolver key lost", "site_specs.identity.email", "a@b.c", "", true},
		{"renderer key lost (the 268/72-losses class)", "renderer.cta_url", "/contact", "", true},
		{"static key blanked to whitespace", "static.icon", "star", "   ", true},
		{"llm prose rewrite is the writer's right", "llm", "old prose", "", false},
		{"llm.summary variant excluded too", "llm.summary", "x", "", false},
		{"undeclared source is not this class", "", "v", "", false},
		{"nothing held, nothing lost", "static.icon", "", "", false},
		{"whitespace-only before is not a held value", "static.icon", "  ", "", false},
		{"a changed value is not a loss", "static.icon", "v", "v2", false},
	}
	for _, c := range cases {
		if got := isNonLLMLoss(c.source, c.before, c.after); got != c.want {
			t.Errorf("%s: isNonLLMLoss(%q,%q,%q) = %v, want %v", c.name, c.source, c.before, c.after, got, c.want)
		}
	}
}

func TestHealVerdict(t *testing.T) {
	cases := []struct {
		name           string
		rows, deployed int
		holds          bool
		want           string
	}{
		{"deployed row holds the key again", 2, 1, true, "healed"},
		{"no row at the slot at all", 0, 0, false, "row_gone"},
		{"row exists, key still blank", 1, 1, false, "open"},
		{"row exists but mid-flux (not deployed), key absent", 1, 0, false, "open"},
	}
	for _, c := range cases {
		if got := healVerdict(c.rows, c.deployed, c.holds); got != c.want {
			t.Errorf("%s: healVerdict(%d,%d,%v) = %q, want %q", c.name, c.rows, c.deployed, c.holds, got, c.want)
		}
	}
}

func TestCarryMissVerdict(t *testing.T) {
	// Partially healed keeps the WHOLE finding open and names the stragglers.
	resolved, verdict, open := carryMissVerdict(map[string]string{
		"cta_url": "healed", "image_url": "open", "alt_text": "open",
	})
	if resolved || verdict != "open" {
		t.Fatalf("partially-healed finding must stay open, got resolved=%v verdict=%q", resolved, verdict)
	}
	if !reflect.DeepEqual(open, []string{"alt_text", "image_url"}) {
		t.Fatalf("open fields = %v, want sorted [alt_text image_url]", open)
	}

	// All healed resolves as healed.
	resolved, verdict, _ = carryMissVerdict(map[string]string{"a": "healed", "b": "row_gone"})
	if !resolved || verdict != "healed" {
		t.Fatalf("all-accounted-for with one healed = (%v,%q), want (true,healed)", resolved, verdict)
	}

	// All gone resolves as row_gone — a different fact than healed.
	resolved, verdict, _ = carryMissVerdict(map[string]string{"a": "row_gone"})
	if !resolved || verdict != "row_gone" {
		t.Fatalf("all-gone = (%v,%q), want (true,row_gone)", resolved, verdict)
	}
}

func TestLossQueryCarriesItsPredicates(t *testing.T) {
	// A tripwire, not a proof (the runtime demand control is the proof): the
	// assembled SQL must carry each load-bearing predicate atom, so a
	// refactor cannot silently drop one and leave a query that still parses.
	q := buildLossQuery(21, false)
	for _, atom := range []string{
		"NOT LIKE 'llm%'", // the definition's exclusion
		"IS NOT NULL",     // declared-source requirement
		"btrim(COALESCE(s.before_data->>k.key,'')) <> ''", // held a value
		"btrim(COALESCE(s.after_data->>k.key,'')) = ''",   // and lost it
		"IS NOT DISTINCT FROM p.slot_name",                // the slot fallback join
	} {
		if !strings.Contains(q, atom) {
			t.Errorf("loss query lost its predicate atom %q", atom)
		}
	}
	if strings.Contains(q, "%!") {
		t.Errorf("fmt verb damage in assembled query")
	}
	// The control variant must pin the clock, not slide with now().
	if c := buildLossQuery(3650, true); !strings.Contains(c, "'2026-08-15'") {
		t.Errorf("demand control lost its pinned upper bound — a sliding control drains to zero and refuses for ever")
	}
}

func TestLossFindingKeyIdentity(t *testing.T) {
	a := lossRow{HistoryID: "h1", Key: "cta_url"}
	b := lossRow{HistoryID: "h1", Key: "image_url"}
	if lossFindingKey(a) == lossFindingKey(b) {
		t.Fatal("two keys lost from one archived generation must be two findings")
	}
}
