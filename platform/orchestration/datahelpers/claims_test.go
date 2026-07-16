// FILE: platform/orchestration/datahelpers/claims_test.go
//
// Benchmark corpus for the claims-verification layer (SPEC_claims_verification
// §8). Every B-numbered fixture is a fabrication that PREVIOUSLY SHIPPED on
// leopardessconsulting.co.uk and passed generation, validation, and deploy.
// The corpus is the regression suite: history graded as benchmark.
//
// The evidence base here mirrors the live site_specs 'evidence_base' row for
// leopardess (V0, written 2026-07-16) — same banned patterns, same fact shapes.

package datahelpers

import (
	"strings"
	"testing"
)

const testEvidenceBaseJSON = `{
  "audit_doc": "docs/leopardessconsulting/AUDIT_verified_facts.md",
  "facts": [
    {"id": "C1-records-verified", "claim": "2,767 business records verified against Companies House",
     "value": 2767, "kind": "metric",
     "source": {"sql": "SELECT count(*) FROM business_intel.businesses WHERE verification_status = 'verified'"},
     "verified_at": "2026-07-16", "tolerance": "exact"},
    {"id": "C1-records-enriched", "claim": "937 records enriched with filed accounts data",
     "value": 937, "kind": "metric",
     "source": {"sql": "SELECT count(*) FROM business_intel.companies_house_data"},
     "verified_at": "2026-07-16", "tolerance": "exact"},
    {"id": "C4-agent-definitions-catalogue", "claim": "agent definitions in the catalogue (supports 'over 150')",
     "value": 157, "kind": "metric",
     "source": {"sql": "SELECT count(*) FROM agent_definitions WHERE deleted_at IS NULL"},
     "verified_at": "2026-07-16", "tolerance": "gte", "context_terms": ["definition", "agent"]},
    {"id": "C4-orchestration-state-records", "claim": "orchestration state records",
     "value": 90790, "kind": "metric",
     "source": {"sql": "SELECT count(*) FROM orchestration_states"},
     "verified_at": "2026-07-16", "tolerance": "gte", "context_terms": ["orchestration", "state record"]},
    {"id": "A-worldsoccernews-uniques", "claim": "worldsoccernews.com ~12 million unique users a month at peak",
     "value": 12000000, "kind": "attestation",
     "source": {"attested_by": "owner, 2026-07-10"},
     "verified_at": "2026-07-10", "tolerance": "approx_pct:10",
     "context_terms": ["unique", "user", "worldsoccernews"]}
  ],
  "banned_claims": [
    {"pattern": "eight (functional |business )?(departments|functions)", "reason": "L0 audit U1: invented departmental taxonomy"},
    {"pattern": "\\b(70|seventy)[ -]?(\\+|plus)?[ -]?(running |specialised |specialized )?agents\\b", "reason": "L0 audit U2: UNSUPPORTED fleet claim"},
    {"pattern": "awards won", "reason": "turn-15: fabricated stat label"},
    {"pattern": "peter grenfell", "reason": "L0 audit U5: invented person"},
    {"pattern": "veterinary data aggregator", "reason": "L0 audit U3: invented client"},
    {"pattern": "playwright", "reason": "L0 audit U6: fabricated scraping tech"}
  ],
  "allowed_entities": ["Companies House", "New Media Age", "worldsoccernews.com"]
}`

func mustParseTestEB(t *testing.T) *EvidenceBase {
	t.Helper()
	eb, err := ParseEvidenceBase([]byte(testEvidenceBaseJSON))
	if err != nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}
	if eb == nil {
		t.Fatal("ParseEvidenceBase returned nil for populated evidence base")
	}
	return eb
}

func scanAll(eb *EvidenceBase, html string) (banned, numbers []ClaimFinding) {
	blocks := ExtractAssertionText(html)
	return eb.ScanBannedClaims(blocks), eb.ScanUnregisteredNumbers(blocks)
}

// ============================================================================
// The benchmark corpus (spec §8)
// ============================================================================

// B1 — "eight departments": banned claim, must be caught. This one shipped,
// was audited out, and came back weeks later on an orphan page.
func TestCorpusB1_EightDepartments(t *testing.T) {
	eb := mustParseTestEB(t)
	for _, html := range []string{
		`<section><p>Over 70 specialised AI agents in eight departments.</p></section>`,
		// The two forms found LIVE on 2026-07-16 (both post-dating the sweep):
		`<p>At Leopardess, the deployed platform runs 70+ agents across eight functional departments.</p>`,
		`<p>The agent platform running across Leopardess's own operations — 70+ agents covering eight business functions — is the same infrastructure we build for clients.</p>`,
	} {
		banned, _ := scanAll(eb, html)
		if len(banned) == 0 {
			t.Errorf("B1: banned claim not caught in %q", html)
		}
	}
}

// B2 — "more than 70 agents": banned claim (audit: UNSUPPORTED fleet claim).
func TestCorpusB2_SeventyAgents(t *testing.T) {
	eb := mustParseTestEB(t)
	for _, html := range []string{
		`<p>The platform runs more than 70 agents in production.</p>`,
		`<p>70+ agents work your problem around the clock.</p>`,
		`<h2>Seventy-plus agents organised for you</h2>`,
		`<p><strong>70+</strong> agents — split across inline markup.</p>`, // inline-split claim
	} {
		banned, _ := scanAll(eb, html)
		found := false
		for _, f := range banned {
			if strings.Contains(f.Reason, "fleet claim") {
				found = true
			}
		}
		if !found {
			t.Errorf("B2: fleet claim not caught in %q (findings: %+v)", html, banned)
		}
	}
}

// B3 — "2,767 Awards Won": a TRUE number under a FALSE label. The spec graded
// this a V3 hard case; the banned label pattern catches it deterministically
// at V1. The number lane correctly does NOT flag 2,767 (it is registered).
func TestCorpusB3_TrueNumberFalseLabel(t *testing.T) {
	eb := mustParseTestEB(t)
	html := `<div class="stat"><span class="stat-value">2,767</span><span class="stat-label">Awards Won</span></div>`
	banned, numbers := scanAll(eb, html)
	if len(banned) != 1 || !strings.Contains(banned[0].Reason, "fabricated stat label") {
		t.Errorf("B3: expected the banned 'awards won' label to catch, got %+v", banned)
	}
	if len(numbers) != 0 {
		t.Errorf("B3: 2,767 is a registered fact value and must not be flagged as a number, got %+v", numbers)
	}
}

// B4 — invented client case studies. The NAMED invented clients are banned
// patterns and are caught. Unnamed prose case studies ("Revenue Operations at
// a Growth-Stage SaaS Company… days to minutes") carry no number and no
// banned pattern — that is V3's lane (LLM claims auditor), DOCUMENTED MISS
// for V1 by design.
func TestCorpusB4_FakeCaseStudies(t *testing.T) {
	eb := mustParseTestEB(t)

	named := `<h3>Case study: Veterinary Data Aggregator</h3>`
	banned, _ := scanAll(eb, named)
	if len(banned) != 1 {
		t.Errorf("B4: named invented client must be caught, got %+v", banned)
	}

	prose := `<p>Revenue Operations at a Growth-Stage SaaS Company: we cut reporting from days to minutes.</p>`
	banned, numbers := scanAll(eb, prose)
	if len(banned) != 0 || len(numbers) != 0 {
		t.Errorf("B4: unnamed prose case study is V3's lane; V1 must stay silent, got banned=%+v numbers=%+v", banned, numbers)
	}
}

// B5 — jane@company.com in body text: still extracted as an asserted email
// after the text-node refactor (the deterministic email check then compares
// it against the site contact).
func TestCorpusB5_EmailInBodyText(t *testing.T) {
	emails := ExtractAssertionEmails(`<p>Contact us at jane@company.com today.</p>`)
	if len(emails) != 1 || emails[0] != "jane@company.com" {
		t.Errorf("B5: body-text email must be extracted, got %v", emails)
	}
}

// B6 — jane@… in a placeholder= attribute: an HTML attribute example, NOT a
// contact assertion. Must NOT be extracted (landmine 1 — this false positive
// once blocked every build of every page using the shared contact-block).
func TestCorpusB6_EmailInPlaceholderAttribute(t *testing.T) {
	emails := ExtractAssertionEmails(
		`<form><input type="email" placeholder="jane@company.com"><button>Send</button></form>`)
	if len(emails) != 0 {
		t.Errorf("B6: placeholder-attribute email must NOT be extracted, got %v", emails)
	}
}

// B7 — "over 150 agent definitions" (157 live): registered, live-verifiable,
// gte tolerance. Must pass V1 with no findings. (V4 freshness later flags it
// if the live count drops below 150.)
func TestCorpusB7_RegisteredLiveVerifiableNumber(t *testing.T) {
	eb := mustParseTestEB(t)
	banned, numbers := scanAll(eb, `<p>The catalogue holds over 150 agent definitions today.</p>`)
	if len(banned) != 0 || len(numbers) != 0 {
		t.Errorf("B7: registered gte number must pass, got banned=%+v numbers=%+v", banned, numbers)
	}
}

// ============================================================================
// Assertion-context extraction (landmine 1: DOM position matters)
// ============================================================================

func TestAssertionContexts(t *testing.T) {
	eb := mustParseTestEB(t)

	// Banned text inside script/code/attribute contexts is not an assertion.
	for name, html := range map[string]string{
		"script":    `<script>var msg = "70+ agents across eight departments";</script>`,
		"code":      `<code>banned: "eight departments"</code>`,
		"attribute": `<img src="x.jpg" alt="eight departments diagram">`,
		"comment":   `<!-- removed: eight departments -->`,
	} {
		banned, _ := scanAll(eb, html)
		if len(banned) != 0 {
			t.Errorf("non-assertion context %q must not be scanned, got %+v", name, banned)
		}
	}
	// NOTE the "attribute" case is a recorded V1 boundary: alt/title text is
	// user-visible but lives in attributes; scanning it is a V3 decision.

	// mailto: hrefs ARE assertions (published contact), code samples are not.
	emails := ExtractAssertionEmails(
		`<a href="mailto:wrong@other.com">Email us</a><code>user@example.com</code>`)
	if len(emails) != 1 || emails[0] != "wrong@other.com" {
		t.Errorf("mailto must be extracted, code sample must not; got %v", emails)
	}
}

// ============================================================================
// Number-scan precision (landmine 2: false-positive classes)
// ============================================================================

// Real copy shapes from the live site that must produce ZERO findings.
func TestNumberScanExclusions(t *testing.T) {
	eb := mustParseTestEB(t)
	for name, html := range map[string]string{
		"phone":        `<p>Reach us at leopardess@contactforsales.com or by phone on +44 (0) 7934 524 911.</p>`,
		"year":         `<p>Founded in 2019 and still going in 2026.</p>`,
		"date":         `<p>Verified on 2026-07-16 against the live database.</p>`,
		"hypothetical": `<p>A 500-token system prompt repeated across 100,000 daily calls costs real money.</p>`,
		"formula":      `<p>ROI = (Annual Labour Recovered − Build Cost) / Build Cost × 100, calculated from your inputs.</p>`,
		"band_label":   `<p>Band 3: Hierarchical Multi-Agent System involves high decision complexity.</p>`,
		"list_ordinal": `<li>2. Current Cost of Human Time on That Work</li>`,
		"time_ratio":   `<p>Support that runs 24/7 without a rota.</p>`,
		"version":      `<p>Deployed as v1.0.1124 last week.</p>`,
		"currency":     `<p>That plan costs £49 per month for the team.</p>`,
		"reading_time": `<p>8 min read</p>`,
	} {
		_, numbers := scanAll(eb, html)
		if len(numbers) != 0 {
			t.Errorf("exclusion class %q must produce no findings, got %+v", name, numbers)
		}
	}
}

// Registered live-copy numbers must pass; the same sentences with drifted
// numbers must flag.
func TestNumberScanRegistration(t *testing.T) {
	eb := mustParseTestEB(t)

	clean := []string{
		`<p>So far it has verified 2,767 records and enriched 937 of them with filed accounts.</p>`,
		`<p>There are 75,061 of those state records from our own systems.</p>`,        // gte: 75061 <= 90790, "state record" context
		`<p>worldsoccernews.com reached 12 million unique users a month at peak.</p>`, // approx_pct:10
		// Both numbers sit in a window containing "agent definitions", so the
		// catalogue gte fact (157) supports 143 AND 56. A recorded precision
		// boundary: window-scoped gte matching over-supports smaller numbers
		// that share the fact's context. Claim-wording precision is V3's lane.
		`<p>Results: 143 agent definitions, 56 of them active.</p>`,
	}
	for _, html := range clean {
		_, numbers := scanAll(eb, html)
		if len(numbers) != 0 {
			t.Errorf("registered numbers must pass in %q, got %+v", html, numbers)
		}
	}

	// Drift beyond a gte fact's value must flag: 900 > 157.
	_, numbers := scanAll(eb, `<p>We ship 900 agent definitions out of the box.</p>`)
	if len(numbers) != 1 || numbers[0].Matched != "900" {
		t.Errorf("number exceeding gte fact value must flag, got %+v", numbers)
	}

	// The core catch case: a fabricated business count.
	_, numbers = scanAll(eb, `<p>We have delivered projects for 45 clients this year.</p>`)
	if len(numbers) != 1 || numbers[0].Matched != "45" {
		t.Errorf("fabricated '45 clients' must flag, got %+v", numbers)
	}
}

// context_terms enforcement: a large gte fact must not blanket-support
// unrelated small numbers ("12 clients" is not supported by the
// orchestration-records count just because 12 <= 90790).
func TestNumberScanContextTermScoping(t *testing.T) {
	eb := mustParseTestEB(t)
	_, numbers := scanAll(eb, `<p>We support 12 clients across the UK.</p>`)
	if len(numbers) != 1 || numbers[0].Matched != "12" {
		t.Errorf("gte facts must be scoped by context_terms; '12 clients' must flag, got %+v", numbers)
	}
}

// ============================================================================
// Parse behaviour
// ============================================================================

func TestParseEvidenceBase(t *testing.T) {
	// Empty / not opted in.
	for _, in := range [][]byte{nil, []byte(`{}`), []byte(`{"facts":[],"banned_claims":[]}`)} {
		eb, err := ParseEvidenceBase(in)
		if err != nil || eb != nil {
			t.Errorf("empty evidence base must parse to nil, got %v / %v", eb, err)
		}
	}

	// An invalid regex degrades to a literal substring — a typo must never
	// silently drop a banned pattern.
	eb, err := ParseEvidenceBase([]byte(
		`{"banned_claims":[{"pattern":"eight departments(", "reason":"typo'd regex"}]}`))
	if err != nil || eb == nil {
		t.Fatalf("evidence base with bad regex must still parse, got %v", err)
	}
	banned := eb.ScanBannedClaims([]string{"we have eight departments( in the org"})
	if len(banned) != 1 {
		t.Errorf("literal fallback must still match, got %+v", banned)
	}
}
