// FILE: platform/orchestration/datahelpers/label_match.go
//
// Deterministic label→page matching: given a button/anchor's own text (e.g.
// "Run the Risk Checker"), find the real page it most plausibly names, or
// report that nothing does. Extracted from
// discovery_checks/check_misdirected_cta.go (behaviour-preserving — that
// file's own tests exercise this code unchanged), so the platform has ONE
// definition instead of two: the audit-time detector and any write-time
// consumer must agree on what "this label names that page" means, or a
// detector and a repair path built on separate copies can disagree with each
// other (exactly the failure this extraction closes — see bugs_open/203,
// where the repair path reused a different, label-blind rule and could
// "recompute" straight into a fresh mismatch).
//
// No LLM call and no embeddings: token-overlap against each candidate's
// name/title/nav_label, stopword-filtered so generic copy ("Learn More",
// "Get Started") never falsely claims a page. This is deliberately the same
// bar the discovery check already applies to LIVE, deployed anchors — reusing
// it at resolution time is verified-mechanism reuse, not a new algorithm.
// name/title tokens are the candidate's IDENTITY and outrank nav_label tokens,
// which are DESCRIPTION only — see BestLabelMatch's own comment (bugs_open/253).

package datahelpers

import "strings"

// LabelStopwords are words that carry no destination meaning in button/link
// text — grammatical stopwords plus generic CTA vocabulary. Text reducing to
// zero non-stopword tokens is "generic" and matches no candidate.
var LabelStopwords = map[string]bool{
	"a": true, "an": true, "and": true, "as": true, "at": true, "be": true,
	"by": true, "for": true, "from": true, "in": true, "into": true,
	"is": true, "it": true, "its": true, "of": true, "on": true, "or": true,
	"our": true, "s": true, "the": true, "their": true, "this": true,
	"to": true, "with": true, "your": true, "you": true, "yours": true,
	"we": true, "us": true,
	// interrogatives — grammatical, not destination-naming, but common enough
	// in ordinary page titles ("What It Costs...", "How It Works") to produce
	// a false-positive single-token overlap if left in. Found by calibration
	// against the live fleet before this fix shipped: "See What We Build"
	// (tokens without this entry: "what","build") token-matched a pricing
	// page whose title happens to start with "What", on "what" alone.
	"what": true, "how": true, "why": true, "when": true, "where": true,
	"who": true, "which": true,
	// generic CTA vocabulary
	"all": true, "click": true, "discover": true, "enter": true,
	"explore": true, "get": true, "go": true, "here": true, "learn": true,
	"meet": true, "more": true, "now": true, "read": true, "see": true,
	"start": true, "started": true, "take": true, "today": true,
	"todays": true, "try": true, "view": true, "visit": true, "join": true,
}

// LabelTokens reduces text to its distinctive lowercase alphanumeric tokens,
// dropping anything in LabelStopwords. Returns nil for fully generic text.
func LabelTokens(text string) []string {
	var tokens []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			t := cur.String()
			if !LabelStopwords[t] {
				tokens = append(tokens, t)
			}
			cur.Reset()
		}
	}
	for _, r := range strings.ToLower(text) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			cur.WriteRune(r)
		default:
			flush()
		}
	}
	flush()
	return tokens
}

// LabelMatchCandidate is one real page as a match candidate. Build with
// NewLabelMatchCandidate — tokens and identityTokens are deliberately
// unexported so a caller cannot construct a candidate whose tokens disagree
// with its own name/title.
type LabelMatchCandidate struct {
	ID          string
	Name        string
	Title       string
	URL         string
	Interactive bool // page_type tool/game — preferred destination on a tie
	// IneligibleAsCTATarget is `NOT pages.eligible_as_cta_target` — true only
	// when the row EXPLICITLY opts out of being a CTA destination
	// (bugs_open/436). INVERTED from the column on purpose: this struct's zero
	// value must mean "eligible", because candidates are built literally in
	// tests and by callers that predate the flag, and a zero value meaning
	// "ineligible" would silently refuse every label match fleet-wide.
	// BestLabelMatchForPage REFUSES (ok=false) when the best match carries
	// this — the page deliberately STAYS in the pool, because removing it lets
	// a weak-token runner-up win instead (measured for the self-link rule,
	// cta_label_universe.go: 10 of 35 wrote somewhere else, most wrong).
	IneligibleAsCTATarget bool
	// identityTokens come structurally from the page's own name/title — its
	// identity. tokens is the wider union of identityTokens plus nav_label
	// (description). BestLabelMatch ranks identityTokens overlap first for
	// exactly this reason: a page's own name/title says what it IS, while its
	// nav_label is marketing copy that can incidentally absorb another page's
	// distinctive words (bugs_open/253).
	identityTokens map[string]bool
	tokens         map[string]bool
}

// NewLabelMatchCandidate builds a candidate from its name, title and
// nav_label. identityTokens is derived from name+title alone (the page's own
// identity); tokens is the wider union with navLabel folded in (description).
// ok is false if none of the three produce a distinctive token — such a page
// can never be matched and callers should not carry it forward.
func NewLabelMatchCandidate(id, name, title, url string, interactive bool, navLabel string) (candidate LabelMatchCandidate, ok bool) {
	identityTokens := map[string]bool{}
	for _, t := range LabelTokens(name) {
		identityTokens[t] = true
	}
	for _, t := range LabelTokens(title) {
		identityTokens[t] = true
	}
	tokens := map[string]bool{}
	for t := range identityTokens {
		tokens[t] = true
	}
	for _, t := range LabelTokens(navLabel) {
		tokens[t] = true
	}
	if len(tokens) == 0 {
		return LabelMatchCandidate{}, false
	}
	return LabelMatchCandidate{
		ID: id, Name: name, Title: title, URL: url,
		Interactive: interactive, identityTokens: identityTokens, tokens: tokens,
	}, true
}

// BestLabelMatch returns the candidate that best names label, and whether one
// was found at all. Ranking, first difference wins:
//
//  1. higher identityOverlap (label tokens present in the candidate's own
//     name/title) — this is checked BEFORE total overlap because name/title
//     are the page's identity while nav_label is description copy. A long
//     marketing nav_label can incidentally contain another page's distinctive
//     words and must not tie with the page the label actually names: on
//     robot-hands.com (bugs_open/253, 2026-08-11) the label "Gripper Safety
//     Factor Calculator" resolved to the payload calculator instead, because
//     its nav_label read "…Validate Capacity with Safety Factor…" — a
//     total-overlap tie that the alphabetical tie-break then broke the wrong
//     way.
//  2. higher totalOverlap (label tokens present anywhere in the candidate's
//     tokens — name/title/nav_label combined). This is the "equal-strength"
//     case TestBestLabelMatch's own comment describes — a barely-related tool
//     page with 1 overlapping token must not beat a clearly-on-topic hub page
//     with 2, which is exactly what shipped in bugs_open/203's follow-on and
//     was caught live on robot-hands.com: "Browse the Gripper Catalog"
//     [gripper,catalog] losing to a cycle-time-estimator tool page over the
//     gripper-catalog-index hub, on "gripper" alone.
//  3. interactive (tool/game) candidates beat non-interactive ones.
//  4. Name ascending — final deterministic tie-break, unchanged from before
//     this fix. A candidate-token-set-size tie-break (smaller wins) was
//     tried here during 2026-08-11 calibration and DROPPED before shipping:
//     on live fleet data it was decided almost entirely by tokenisation
//     artefacts and site-wide generic words carrying no real signal — a
//     stray hyphen in one candidate's own copy ("Break-Even" splitting into
//     two tokens the sibling candidate's "breakeven" didn't) flipped 9
//     already-correct, live gaswholesalers.com CTAs onto the wrong tool
//     purely because the loser's title happened to be one token longer, and
//     the same pattern recurred wherever a tie was driven by a single
//     domain-wide word ("cma" on vetcomparison.uk) or a generic imperative
//     verb sitting in a page's own title ("Run MatchMatrix" absorbing
//     "Run Payload Calculation" on robot-hands.com). None of that is a
//     property of which candidate the label actually names, so it made
//     ties WORSE than plain alphabetical, not better. See
//     CALIBRATION_2026-08-11_label_match_identity_report.txt for the full
//     evidence. identityOverlap (key 1) is what fixes bugs_open/253 itself —
//     that case resolves outright at key 1 and never reaches this tie-break.
//
// Requires totalOverlap >= 1 — a page matchable only via nav_label stays
// matchable; a label with no distinctive tokens, or one that matches no
// candidate at all, reports !ok rather than guessing.
//
// ── AMBIGUITY (bugs_open/308, Phase B, 2026-08-23) ─────────────────────────
//
// Key 4 above is ALPHABETICAL ORDER. When it is what decides the winner, the
// result carries no information about which page the label names — and the
// 2026-08-11 calibration had already established that no substitute key fixes
// that (a token-set-size key was tried, measured and dropped; this lane
// measured a name-tier key and a path-depth key and dropped both for the same
// reason: each repaired some cases and broke others on the same mechanism —
// see CALIBRATION_2026-08-23_phase_b_widening_report.md §3).
//
// So a winner separated from a DIFFERENT page by nothing but its name's
// spelling is reported as `ambiguous`, with ok=false. Measured on the frozen
// fleet dump of 2026-08-23: 263 of 1,146 matches (23%) against the widened CTA
// candidate universe, of which 137 would have OVERWRITTEN a live CTA
// destination. Two families in that set are known-wrong and were about to be
// executed fleet-wide by 308's widening: finetuning.uk's "how we work"
// (stored /how-we-work.html, the page its copy names) losing to /about.html
// because the About page's TITLE reads "…Who We Are and How We Work"; and
// dartsonline.com's "Read the guides" (stored /guides/index.html) losing to an
// About page whose title ends "Spec-First Darts Guides".
//
// ok=false on ambiguity is the SAFE READING and is deliberate: a caller that
// ignores the third value gets "this label names nothing", which for both CTA
// writers means their keep branches hold the stored value, and for the
// detector means no finding is filed. The third return exists so a caller that
// wants to TELL the two apart (copy names nothing / copy names two things
// equally) can, without any caller being able to reach the ambiguous winner by
// accident — the same signature-is-the-enforcement reasoning as Phase A's
// storedCTADestinationIsAuthored.
//
// What it does NOT catch, stated because the gap is real: a CONFIDENT false
// match. dartsonline's "See how each brand differs, spec by spec" resolves to
// /about.html on an identity-token win ("spec"), not a tie. No tie rule sees
// that, and no stopword list can (`spec` is distinctive on that site).
//
// ⚠ THE CONSUMER SET IS CLOSED, NOT OPEN, and the refusal above is tuned for
// it (council 00732119, architecture seat). FOUR non-test call sites as of
// 2026-08-26, all CTA, verified by `grep -rn "BestLabelMatch(\|BestLabelMatchForPage(\|JudgeCTALabel("
// --include=*.go platform/ internal/ pkg/ cmd/`:
//
//	resolve_internal_links_action.go   setCTAField        (build writer)
//	rerender_page_sections_action.go   applyCTARecompute  (repair writer)
//	cta_label_agreement.go             JudgeCTALabel      (the shared question),
//	                                   through which BOTH read-side consumers now
//	                                   go — discovery_checks/check_misdirected_cta.go's
//	                                   ctaClassifyAnchor (itself reused by
//	                                   check_cta_nonpage.go) and, since
//	                                   bugs_open/399, actions/cta_label_audit.go's
//	                                   write-time pass.
//
// The count is FOUR and not five on purpose: ctaClassifyAnchor stopped being a
// direct caller when bugs_open/399 extracted the question it was asking. That
// extraction is what keeps the detector and the write-time audit from drifting
// (RFC_047 §9), and it means a change to the refusal semantics below now reaches
// both of them together rather than one of them silently.
//
// A FOURTH consumer — a breadcrumb or sitemap label matcher, say — would
// silently inherit "refuse rather than guess", which is the right default when
// the consequence is REWRITING A LIVE BUTTON and may be the wrong one when the
// consequence is merely a weaker suggestion. Decide it deliberately at the new
// call site, and add it to the list above; do not assume the semantics were
// chosen with your case in mind.
func BestLabelMatch(label string, candidates []LabelMatchCandidate) (best LabelMatchCandidate, ok bool, ambiguous bool) {
	tokens := LabelTokens(label)
	if len(tokens) == 0 {
		return LabelMatchCandidate{}, false, false
	}

	// scored pairs a candidate with its overlap counts against label, so
	// better (below) can compare two candidates without recomputing overlap.
	type scored struct {
		c        *LabelMatchCandidate
		identity int
		total    int
	}

	// better reports whether a should replace b as the current best, under
	// the ordering documented above: identity overlap, then total overlap,
	// then interactivity, then name.
	better := func(a, b scored) bool {
		if a.identity != b.identity {
			return a.identity > b.identity
		}
		if a.total != b.total {
			return a.total > b.total
		}
		if a.c.Interactive != b.c.Interactive {
			return a.c.Interactive
		}
		return a.c.Name < b.c.Name
	}

	overlap := func(c *LabelMatchCandidate) scored {
		identityOverlap, totalOverlap := 0, 0
		for _, t := range tokens {
			if c.tokens[t] {
				totalOverlap++
			}
			if c.identityTokens[t] {
				identityOverlap++
			}
		}
		return scored{c: c, identity: identityOverlap, total: totalOverlap}
	}

	var bestScored scored
	found := false
	for i := range candidates {
		cur := overlap(&candidates[i])
		if cur.total == 0 {
			continue
		}
		if !found || better(cur, bestScored) {
			bestScored = cur
			found = true
		}
	}
	if !found {
		return LabelMatchCandidate{}, false, false
	}

	// Was the winner separated from a DIFFERENT page only by key 4? Compared on
	// the NORMALISED path, so two candidate rows for one page (/contact/ and
	// /contact/index.html) are the same destination and not a disagreement —
	// the ambiguity being tested is "which page", never "which row".
	winner := NormalizePagePath(bestScored.c.URL)
	for i := range candidates {
		c := &candidates[i]
		if NormalizePagePath(c.URL) == winner {
			continue
		}
		rival := overlap(c)
		if rival.total == 0 {
			continue
		}
		if rival.identity == bestScored.identity && rival.total == bestScored.total &&
			c.Interactive == bestScored.c.Interactive {
			return LabelMatchCandidate{}, false, true
		}
	}
	return *bestScored.c, true, false
}
