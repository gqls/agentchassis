// FILE: cmd/brief-negation-check/briefcheck_test.go
//
// The fixtures are real brief text: the tagline block from
// ai-agent-orchestration.com's content_direction (the site that drew the
// owner's complaint), a differentiator list, an instructional contrast, and a
// possessive that an earlier tool read as a quotation.
//
// Mutation checks (by hand, recorded in the lane NOTES):
//   - drop the possessive guard from quotedRe        -> TestPossessiveIsNotAQuotation fails
//   - count instructional contrasts as supplied      -> TestInstructionalIsCountedNotFiled fails
//   - keep aspect-only paths in visibleSurface       -> TestVisibleSurfaceDropsAspectOnlyGuards fails

package main

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func TestVisibleSurfaceIsDerivedFromTheLiveConfig(t *testing.T) {
	cfg := `"prompt_template": "## Content Direction\n{{if .site_specs.specs.content_direction}}{{if .site_specs.specs.content_direction.formatted}}{{.site_specs.specs.content_direction.formatted}}{{end}}{{end}}\n{{if .site_specs.specs.identity.target_audience}}Target Audience: {{.site_specs.specs.identity.target_audience}}{{end}}\n{{if .site_specs.specs.identity.key_differentiators}}Key Differentiators: {{.site_specs.specs.identity.key_differentiators}}{{end}}\n{{if .site_specs.specs.evidence_base.writer_block}}{{.site_specs.specs.evidence_base.writer_block}}{{end}}"`
	got := visibleSurface(cfg)
	want := []string{"content_direction.formatted", "evidence_base.writer_block",
		"identity.key_differentiators", "identity.target_audience"}
	if len(got) != len(want) {
		t.Fatalf("surface: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("surface[%d]: got %q want %q", i, got[i], want[i])
		}
	}
}

// An {{if .site_specs.specs.content_direction}} guard names the ASPECT. Keeping
// it would pull the whole 15,760-character document back into the measurement
// through the back door — which is the error this check exists not to repeat.
func TestVisibleSurfaceDropsAspectOnlyGuards(t *testing.T) {
	for _, s := range visibleSurface(`{{if .site_specs.specs.content_direction}}{{.site_specs.specs.content_direction.formatted}}{{end}}`) {
		if s == "content_direction" {
			t.Error("the aspect-only guard must not become a measured field")
		}
	}
	// An aspect with NO deeper path is a real field reference and must survive.
	got := visibleSurface(`{{.site_specs.specs.voice}}`)
	if len(got) != 1 || got[0] != "voice" {
		t.Errorf("an aspect referenced with no deeper path is a real field: %v", got)
	}
}

func TestSuppliedPhraseRoutesAndMandate(t *testing.T) {
	specs := []siteSpecs{{
		Domain: "ai-agent-orchestration.com", SiteID: "00000000-0000-0000-0000-000000000001",
		Aspect: "content_direction",
		Data: map[string]interface{}{
			"formatted": "Emphasis: The canonical tagline \"Multi-agent systems deployed to production in days, not months\" must appear in the homepage hero, services page hero, site footer, and meta descriptions.\n\n" +
				"Voice: Use concrete stack references (Kubernetes, Kafka, Postgres) naturally, not as buzzwords.",
		},
	}, {
		Domain: "ai-agent-orchestration.com", SiteID: "00000000-0000-0000-0000-000000000001",
		Aspect: "identity",
		Data: map[string]interface{}{
			"key_differentiators": []interface{}{
				"Verified, not hallucinated: agents check facts against official sources.",
				"We run the platform on our own sites first.",
			},
		},
	}}
	got := assess(specs, []string{"content_direction.formatted", "identity.key_differentiators"})
	if len(got) != 1 {
		t.Fatalf("expected one site, got %d", len(got))
	}
	a := got[0]
	if len(a.Supplied) != 2 {
		t.Fatalf("expected two supplied phrases (one quoted, one list item), got %d: %+v", len(a.Supplied), a.Supplied)
	}
	var quoted, list *suppliedPhrase
	for i := range a.Supplied {
		switch a.Supplied[i].Route {
		case "quoted":
			quoted = &a.Supplied[i]
		case "list_item":
			list = &a.Supplied[i]
		}
	}
	if quoted == nil || list == nil {
		t.Fatalf("expected both routes, got %+v", a.Supplied)
	}
	if !quoted.Mandated {
		t.Error("a block ordering the tagline onto pages must be marked MANDATED — it is the strongest evidence a brief can carry")
	}
	if a.Mandated() != 1 {
		t.Errorf("expected exactly one mandated phrase, got %d", a.Mandated())
	}
	if !a.Finding() {
		t.Error("a site supplying phrases must produce a finding")
	}
}

// "not as buzzwords" is the brief INSTRUCTING. Whether that transfers into
// output is unmeasured, so it is counted and never filed. Adding the two classes
// together is what made an earlier fleet census wrong by roughly 2x.
func TestInstructionalIsCountedNotFiled(t *testing.T) {
	specs := []siteSpecs{{
		Domain: "example.uk", SiteID: "00000000-0000-0000-0000-000000000002", Aspect: "content_direction",
		Data: map[string]interface{}{
			"formatted": "Voice: Use concrete stack references naturally, not as buzzwords. Frame the next step as an engineering discussion, not a purchase decision.",
		},
	}}
	a := assess(specs, []string{"content_direction.formatted"})[0]
	if len(a.Supplied) != 0 {
		t.Errorf("instructional contrast must not be filed as supplied: %+v", a.Supplied)
	}
	if a.Instructional < 2 {
		t.Errorf("instructional contrasts must still be COUNTED, got %d", a.Instructional)
	}
	if a.Finding() {
		t.Error("a brief that only instructs must not produce a finding")
	}
}

func TestPossessiveIsNotAQuotation(t *testing.T) {
	got := quotedSpans("Write in the client's own voice, not a long form and not a questionnaire.")
	if len(got) != 0 {
		t.Errorf("a possessive apostrophe must not open a quoted span: %q", got)
	}
	if got := quotedSpans("Use 'shipped in days, not months' verbatim."); len(got) != 1 || got[0] != "shipped in days, not months" {
		t.Errorf("a real single-quoted span must still be caught: %q", got)
	}
	if got := quotedSpans(`Use "shipped in days, not months" verbatim.`); len(got) != 1 {
		t.Errorf("a double-quoted span must be caught: %q", got)
	}
}

// A clean brief must produce no finding — and the run must still be able to
// report, so the check can distinguish "examined and clean" from "did not run".
func TestCleanBriefProducesNoFinding(t *testing.T) {
	specs := []siteSpecs{{
		Domain: "clean.uk", SiteID: "00000000-0000-0000-0000-000000000003", Aspect: "content_direction",
		Data: map[string]interface{}{"formatted": "Voice: plain, direct, British English. Say what the tool does and who it is for."},
	}}
	a := assess(specs, []string{"content_direction.formatted"})[0]
	if a.Finding() || len(a.Supplied) != 0 || a.Instructional != 0 {
		t.Errorf("clean brief flagged: %+v", a)
	}
	if r := render([]siteAssessment{a}, []string{"content_direction.formatted"}, nil, nil, true, false); r == "" {
		t.Error("a clean run must still render a report — a silent clean run is indistinguishable from a job that did not run")
	}
}

// The visible/document split is the whole discipline of this check.
func TestVisibleCharsIsNotTheDocument(t *testing.T) {
	specs := []siteSpecs{{
		Domain: "big.uk", SiteID: "00000000-0000-0000-0000-000000000004", Aspect: "content_direction",
		Data: map[string]interface{}{
			"formatted":     "Voice: plain and direct.",
			"cta_style":     map[string]interface{}{"approach": "CTAs should initiate a technical conversation, not a sales process."},
			"example_lines": []interface{}{"Agents fail in isolation, not in cascades."},
		},
	}}
	a := assess(specs, []string{"content_direction.formatted"})[0]
	if a.VisibleChars >= a.DocChars {
		t.Errorf("visible (%d) must be smaller than the document (%d) — the writer reads named fields, not the spec",
			a.VisibleChars, a.DocChars)
	}
	if len(a.Supplied) != 0 || a.Instructional != 0 {
		t.Errorf("text outside the visible surface must not be measured at all: %+v", a)
	}
}

// MANDATED must mean "the block handing this phrase over also orders it onto
// pages" — not "somewhere in this 15,000-character brief there is a mandate
// verb". The first live run reported a block about sentence rhythm as mandating
// its own example, which is how this test came to exist.
func TestMandateIsScopedToTheBlock(t *testing.T) {
	specs := []siteSpecs{{
		Domain: "two-blocks.uk", SiteID: "00000000-0000-0000-0000-000000000005", Aspect: "content_direction",
		Data: map[string]interface{}{
			"formatted": "Emphasis: the canonical tagline \"Built in days, not months\" must appear in the homepage hero.\n\n" +
				"Structure patterns: many key sentences are structured as \"X is Y, not Z\".",
		},
	}}
	a := assess(specs, []string{"content_direction.formatted"})[0]
	if len(a.Supplied) != 2 {
		t.Fatalf("expected both quoted phrases, got %+v", a.Supplied)
	}
	byPhrase := map[string]bool{}
	for _, s := range a.Supplied {
		byPhrase[s.Phrase] = s.Mandated
	}
	if !byPhrase["Built in days, not months"] {
		t.Error("the phrase in the block carrying 'must appear' is mandated")
	}
	if byPhrase["X is Y, not Z"] {
		t.Error("a phrase in a DIFFERENT block must not inherit that block's mandate")
	}
	if a.Mandated() != 1 {
		t.Errorf("exactly one mandated phrase, got %d", a.Mandated())
	}
}

// A required disclosure is not a house mannerism. The writer-seam gate refuses
// to rewrite these; this check must refuse to file them, by the SAME rule — or
// the two halves send a human in opposite directions, and this one's direction
// is "edit a compliance statement".
func TestRegulatorySuppliedIsCountedNotFiled(t *testing.T) {
	specs := []siteSpecs{{
		Domain: "lender.uk", SiteID: "00000000-0000-0000-0000-000000000006", Aspect: "content_direction",
		Data: map[string]interface{}{
			"formatted": "Compliance: every calculator page must carry \"The figures are an estimate, not financial advice.\"\n\n" +
				"Emphasis: use the tagline \"Answers in minutes, not days\" on the homepage hero.",
		},
	}}
	a := assess(specs, []string{"content_direction.formatted"})[0]
	if a.Regulatory != 1 {
		t.Errorf("the disclosure must be counted as regulatory, got %d", a.Regulatory)
	}
	if len(a.Supplied) != 1 || a.Supplied[0].Phrase != "Answers in minutes, not days" {
		t.Errorf("only the house tagline should be filed, got %+v", a.Supplied)
	}
}

// A key coarser than its finding drops every finding after the first: the second,
// DIFFERENT finding hits ON CONFLICT DO NOTHING and is gone while the open row
// goes on describing the first thing it saw. A brief is edited by config at any
// time, so this item's CONTENT changes under a stable site id — the key has to
// carry the phrase set. (Council round 2, guidelines seat.)
func TestItemKeyIsAsGranularAsTheFinding(t *testing.T) {
	site := "00000000-0000-0000-0000-000000000009"
	a := []suppliedPhrase{{Phrase: "in days, not months"}}
	b := []suppliedPhrase{{Phrase: "in days, not months"}, {Phrase: "verified, not hallucinated"}}

	if itemKey(site, a) == itemKey(site, b) {
		t.Error("two different phrase sets must not share an item_key — the second finding would be dropped")
	}
	// Order must not matter: the same set found in a different walk order is the
	// same finding, and a key that churned on order would file a new item daily.
	rev := []suppliedPhrase{{Phrase: "verified, not hallucinated"}, {Phrase: "in days, not months"}}
	if itemKey(site, b) != itemKey(site, rev) {
		t.Error("the same phrase set in a different order must produce the same key")
	}
	// And it is still per-site.
	if itemKey(site, a) == itemKey("00000000-0000-0000-0000-00000000000a", a) {
		t.Error("two sites with the same phrase must not share a key")
	}
}

// ---------------------------------------------------------------------------
// The word arm (2026-09-02). BANNED_REGISTER_v1's banned_words had NO reader in
// this binary — six scan sites, all shapes — so `plainly` and `honest*`, the two
// words the owner named FIRST, were invisible to the nightly run while every
// shape was fenced. Same disciplines as the shapes: a handover files, an
// instruction is counted, and one span yields one finding with the shape taking
// precedence.
// ---------------------------------------------------------------------------

func TestWordArmQuotedHandoverFiles(t *testing.T) {
	specs := []siteSpecs{{
		Domain: "example.uk", SiteID: "00000000-0000-0000-0000-000000000010", Aspect: "content_direction",
		Data: map[string]interface{}{
			"formatted": "Emphasis: use this exact phrase \"We explain your cover plainly and without jargon\" across the site.",
		},
	}}
	a := assess(specs, []string{"content_direction.formatted"})[0]
	if len(a.Supplied) != 1 {
		t.Fatalf("a quoted span carrying a banned WORD is a handover and must file: %+v", a.Supplied)
	}
	p := a.Supplied[0]
	if !strings.HasPrefix(p.Shape, "word:") {
		t.Errorf("word-arm finding must be labelled word:<name>, got %q", p.Shape)
	}
	if !p.Mandated {
		t.Error("'use this exact phrase' is a mandate for the word arm exactly as for shapes")
	}
	if !a.Finding() {
		t.Error("a word handover must produce a finding")
	}
}

func TestWordArmInstructionalIsCountedNotFiled(t *testing.T) {
	specs := []siteSpecs{{
		Domain: "example.uk", SiteID: "00000000-0000-0000-0000-000000000011", Aspect: "content_direction",
		Data: map[string]interface{}{
			"formatted": "Voice: write plainly. Be honest about limitations without ceremony.",
		},
	}}
	a := assess(specs, []string{"content_direction.formatted"})[0]
	if len(a.Supplied) != 0 {
		t.Errorf("instructional word use must not file: %+v", a.Supplied)
	}
	if a.WordInstructional < 2 {
		t.Errorf("instructional word hits must be COUNTED, got %d", a.WordInstructional)
	}
	if a.Finding() {
		t.Error("word instruction alone must not produce a finding")
	}
}

// One span, both violations: the shape files, the word does not double-count.
// A double-filed span would perturb the phrase-set item key and churn open items.
func TestWordArmShapeTakesPrecedenceInOneSpan(t *testing.T) {
	specs := []siteSpecs{{
		Domain: "example.uk", SiteID: "00000000-0000-0000-0000-000000000012", Aspect: "content_direction",
		Data: map[string]interface{}{
			"formatted": "Emphasis: the tagline \"We say it plainly, not in marketing language\" must appear on every page.",
		},
	}}
	a := assess(specs, []string{"content_direction.formatted"})[0]
	if len(a.Supplied) != 1 {
		t.Fatalf("one span must yield exactly one finding, got %d: %+v", len(a.Supplied), a.Supplied)
	}
	if strings.HasPrefix(a.Supplied[0].Shape, "word:") {
		t.Errorf("shape must take precedence over the word arm on one span, got %q", a.Supplied[0].Shape)
	}
}

// Control: the word patterns must actually fire on this binary's inputs — a
// vocabulary drift in the datahelpers mirror would make every word test above
// pass vacuously if this one did not pin the arm itself.
func TestWordArmControlFiresOnTheNamedWords(t *testing.T) {
	for _, s := range []string{"we explain plainly", "an honest assessment"} {
		if len(datahelpers.ScanBannedRegisterWords(s)) == 0 {
			t.Errorf("word arm silent on %q — the register mirror lost the owner's named words", s)
		}
	}
}
