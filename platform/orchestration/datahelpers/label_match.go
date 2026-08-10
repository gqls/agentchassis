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
// NewLabelMatchCandidate — tokens is deliberately unexported so a caller
// cannot construct a candidate whose tokens disagree with its own name/title.
type LabelMatchCandidate struct {
	ID          string
	Name        string
	Title       string
	URL         string
	Interactive bool // page_type tool/game — preferred destination on a tie
	tokens      map[string]bool
}

// NewLabelMatchCandidate builds a candidate from whichever text sources name
// it (typically name, title, nav_label). ok is false if none of them produce
// a distinctive token — such a page can never be matched and callers should
// not carry it forward.
func NewLabelMatchCandidate(id, name, title, url string, interactive bool, tokenSources ...string) (candidate LabelMatchCandidate, ok bool) {
	tokens := map[string]bool{}
	for _, src := range tokenSources {
		for _, t := range LabelTokens(src) {
			tokens[t] = true
		}
	}
	if len(tokens) == 0 {
		return LabelMatchCandidate{}, false
	}
	return LabelMatchCandidate{
		ID: id, Name: name, Title: title, URL: url,
		Interactive: interactive, tokens: tokens,
	}, true
}

// BestLabelMatch returns the candidate whose tokens best overlap label's, and
// whether one was found at all. Ranking: higher token overlap wins outright;
// interactive (tool/game) candidates beat non-interactive ones only when
// overlap is tied (this is the "equal-strength" case TestBestLabelMatch's own
// comment describes — overlap must be compared FIRST, or a barely-related
// tool page with 1 overlapping token beats a clearly-on-topic hub page with 2,
// which is exactly what shipped in bugs_open/203's follow-on and was caught
// live on robot-hands.com: "Browse the Gripper Catalog" [gripper,catalog]
// losing to a cycle-time-estimator tool page over the gripper-catalog-index
// hub, on "gripper" alone). Name is the final deterministic tie-break.
// Requires >= 1 overlapping token — a label with no distinctive tokens, or
// one that matches no candidate, reports !ok rather than guessing.
func BestLabelMatch(label string, candidates []LabelMatchCandidate) (best LabelMatchCandidate, ok bool) {
	tokens := LabelTokens(label)
	if len(tokens) == 0 {
		return LabelMatchCandidate{}, false
	}
	var bestPtr *LabelMatchCandidate
	bestOverlap := 0
	for i := range candidates {
		c := &candidates[i]
		overlap := 0
		for _, t := range tokens {
			if c.tokens[t] {
				overlap++
			}
		}
		if overlap == 0 {
			continue
		}
		if bestPtr == nil ||
			overlap > bestOverlap ||
			(overlap == bestOverlap && c.Interactive && !bestPtr.Interactive) ||
			(overlap == bestOverlap && c.Interactive == bestPtr.Interactive && c.Name < bestPtr.Name) {
			bestPtr = c
			bestOverlap = overlap
		}
	}
	if bestPtr == nil {
		return LabelMatchCandidate{}, false
	}
	return *bestPtr, true
}
