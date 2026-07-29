// FILE: platform/orchestration/datahelpers/claims.go
//
// Shared claims-vs-evidence scanning for the claims-verification layer
// (SPEC_claims_verification, V1). Consumed by BOTH the build-time gate
// (validate_page_content) and the post-deploy audit (check_unverified_claims),
// so the two agree by one literal implementation on what counts as an
// asserted claim — the same sharing pattern as ExtractHrefs/PageURLSet.
//
// Three pieces live here:
//
//  1. The EvidenceBase type — the structured form of the site_specs
//     'evidence_base' aspect: verified facts (with optional numeric values),
//     banned claim patterns (the site's own audited-out fabrications), and
//     allowed entities (consumed by the V3 auditor, carried here for parity).
//
//  2. Assertion-text extraction. Claim scans MUST parse text nodes, not raw
//     HTML: an email or number in a placeholder= attribute, a <script> body,
//     or a <code> sample is not an assertion about the business. (The email
//     validator once blocked every build over placeholder="jane@company.com"
//     — an HTML attribute example, not a contact claim.) mailto: hrefs are
//     the one attribute surface that IS an assertion (published contact), so
//     email extraction includes them explicitly.
//
//  3. The two deterministic scans:
//     - ScanBannedClaims: regex scan of assertion text against banned_claims.
//       These are KNOWN falsehoods for this site — findings are definitive.
//     - ScanUnregisteredNumbers: high-precision extraction of number-bearing
//       business claims, flagged when no registered fact supports the number.
//       Extraction has false positives by design tolerance — callers must
//       route findings to human review, never block outright on them. It takes
//       a ClaimSurface (bugs_open/102): its lexical gate cannot tell an
//       explainer's worked example from a sales claim, so the page's structural
//       type decides whether prose numbers are claims at all. Banned claims and
//       stat fields are scanned on every surface regardless.
//
// Truth decisions are human: nothing in this file rewrites content. The
// scans produce findings; humans rule on them.

package datahelpers

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// ============================================================================
// Evidence base types
// ============================================================================

// EvidenceSource records where a fact's proof lives. Exactly one field is set:
//   - SQL: live-verifiable — a query whose result is the fact's value (V4
//     freshness re-runs these).
//   - Artifact: code path or URL evidence — checked for presence in the
//     register, not re-proved.
//   - AttestedBy: human word (e.g. "owner, 2026-07-10") — the honest standing
//     of a claim only a person can vouch for.
type EvidenceSource struct {
	SQL        string `json:"sql,omitempty"`
	Artifact   string `json:"artifact,omitempty"`
	AttestedBy string `json:"attested_by,omitempty"`
}

// EvidenceFact is one verified claim. A fact row supports a SPECIFIC claim
// wording, not a topic (the audit's caveat semantics: "verified against
// Companies House" is TRUE while "handles dissolved companies" would be FALSE).
type EvidenceFact struct {
	ID         string         `json:"id"`
	Claim      string         `json:"claim"`
	Value      *float64       `json:"value,omitempty"` // set for number-bearing facts
	Kind       string         `json:"kind"`            // metric | capability | entity | attestation
	Source     EvidenceSource `json:"source"`
	VerifiedAt string         `json:"verified_at"`
	// Tolerance for numeric matching: "exact" (default), "gte" (published
	// numbers up to Value are supported — "over 150 X" phrasing), or
	// "approx_pct:N" (within N percent).
	Tolerance string `json:"tolerance,omitempty"`
	// ContextTerms scope which claim windows this fact may support. REQUIRED
	// in effect for non-exact tolerances: a gte fact with a large value would
	// otherwise blanket-support every smaller number on the site ("12 clients"
	// must not pass because 12 <= the orchestration-records count). A non-exact
	// fact without context terms is matched as exact.
	ContextTerms []string `json:"context_terms,omitempty"`
	// Observations carries the dated points of a Kind=="series" fact. A series
	// fact leaves Value nil: it has no single value, which is the whole point.
	// See claims_series.go — every observation carries its OWN source and is
	// never allowed to inherit this fact's.
	Observations []Observation `json:"observations,omitempty"`
}

// ============================================================================
// Fact kinds (bugs_open/105)
// ============================================================================
//
// EvidenceFact.Kind was declared, documented in the spec, written by nine
// registers — and read by nothing. A field that looks like a contract and
// governs nothing eventually gets written by someone expecting behaviour from
// it, and in the meantime it is the slot a whole class of claim needs: a
// `capability` fact whose source is the mechanism that keeps it is how a
// PROMISE ("we correct errors when told") becomes mechanically checkable rather
// than prose. The design was already present; only the reader was missing.
//
// TWO THINGS SHAPED THIS, both measured rather than assumed:
//
//  1. The live vocabulary is NOT the documented one. Across the nine current
//     registers on 2026-07-27: metric 46, count 18, entity 11, capability 9,
//     attestation 4. `count` is used by four sites and appears in no
//     documentation. So the bug file's candidate 1 as written — "reject unknown
//     kinds" — would have failed the registers of four live sites closed. It is
//     an alias of metric, and is treated as one.
//
//  2. Kind is NOT normalised in place, deliberately. EvidenceBase is marshalled
//     BACK to site_specs by refresh_evidence_base_action.go:677 and
//     evidence_citations.go:350, so rewriting the field at parse time would
//     silently mutate 18 stored facts from "count" to "metric" through a write
//     path that never intended to touch them. Callers read CanonicalKind()
//     instead; the stored value stays exactly as its author wrote it.
//
// This mirrors the rule the banned-claim parser above already follows — a typo
// must never silently drop a claim — rather than inventing a stricter one for
// this field alone.
const (
	FactKindMetric      = "metric"      // a number backed by a query or artifact
	FactKindCapability  = "capability"  // something the platform DOES, incl. a promise
	FactKindEntity      = "entity"      // a named thing that exists
	FactKindAttestation = "attestation" // a human's word, not re-provable
)

// factKindAliases maps live-but-undocumented spellings onto the canonical
// vocabulary. `count` is the only one in the fleet today and is semantically a
// metric — both are "a number a query can re-derive".
var factKindAliases = map[string]string{
	"count":   FactKindMetric,
	"metrics": FactKindMetric,
	"counts":  FactKindMetric,
}

var canonicalFactKinds = map[string]bool{
	FactKindMetric: true, FactKindCapability: true,
	FactKindEntity: true, FactKindAttestation: true,
}

// CanonicalKind maps this fact's Kind onto the documented vocabulary without
// touching the stored value. An absent kind defaults to metric, which is what
// every kind-less fact in the fleet is. An unrecognised kind is returned as
// metric too — the safe default — and is separately reportable via
// KindIsRecognised so the anomaly is visible rather than silently absorbed.
func (f EvidenceFact) CanonicalKind() string {
	k := strings.ToLower(strings.TrimSpace(f.Kind))
	if k == "" {
		return FactKindMetric
	}
	if canonicalFactKinds[k] {
		return k
	}
	if alias, ok := factKindAliases[k]; ok {
		return alias
	}
	return FactKindMetric
}

// KindIsRecognised reports whether the stored Kind is one this code understands
// — canonical, a known alias, or deliberately absent. It is the discriminator a
// caller needs to tell "this fact is a metric" from "nobody knows what this fact
// is and it is being treated as a metric".
func (f EvidenceFact) KindIsRecognised() bool {
	k := strings.ToLower(strings.TrimSpace(f.Kind))
	if k == "" {
		return true
	}
	_, aliased := factKindAliases[k]
	return canonicalFactKinds[k] || aliased
}

// NOTE — IsLiveVerifiable was here and has been REMOVED (council, 2026-07-27).
//
// It encoded the spec's real distinction (a sql-sourced fact is re-provable and
// goes stale; an attestation is a human's word, checked for presence, never
// re-run) and had no production caller: only tests. The editquality seat pointed
// out that this reproduces, for a second symbol, the exact defect this change
// exists to fix — declared, documented, read by nothing. It was right, and the
// objection is not answerable by arguing the accessor is nice to have.
//
// It belongs with its consumer, which is the bug file's candidate 2 (distinct V4
// treatment for capability/attestation). V4 iterates raw map[string]interface{}
// facts rather than typed EvidenceFact, so it needs a map-side sibling anyway;
// that cost is part of candidate 2 and is not paid here.

// AliasedKinds returns the distinct stored Kind values that this code silently
// resolves to something else, so a caller can report them.
//
// The council's guardian seat objected that mapping `count` to `metric` is an
// interpretive judgement — it may have been an intentional distinct kind for the
// 4 sites and 18 facts that use it — and that because CanonicalKind resolves it
// silently, "no signal will ever surface if the guess is wrong". That is exactly
// right, and it is cheap to fix: the alias now announces itself wherever the
// register is loaded, so a human with authorship context on those sites sees it
// rather than having to go looking.
func (eb *EvidenceBase) AliasedKinds() []string {
	if eb == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, f := range eb.Facts {
		k := strings.ToLower(strings.TrimSpace(f.Kind))
		if k == "" || canonicalFactKinds[k] {
			continue
		}
		if target, ok := factKindAliases[k]; ok && !seen[k] {
			seen[k] = true
			out = append(out, strings.TrimSpace(f.Kind)+" -> "+target)
		}
	}
	sort.Strings(out)
	return out
}

// UnrecognisedKinds returns the distinct stored Kind values in this register
// that no rule understands, so a caller can report them. Empty for every live
// register on 2026-07-27; it exists so the NEXT typo is visible on the day it is
// written rather than after someone wonders why a fact behaves like a metric.
func (eb *EvidenceBase) UnrecognisedKinds() []string {
	if eb == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, f := range eb.Facts {
		if f.KindIsRecognised() {
			continue
		}
		k := strings.TrimSpace(f.Kind)
		if !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// BannedClaim is one audited-out fabrication — a per-site regression pattern.
type BannedClaim struct {
	Pattern string `json:"pattern"` // case-insensitive regex; invalid regex falls back to literal substring
	Reason  string `json:"reason"`

	re *regexp.Regexp // compiled on parse
}

// EvidenceBase is the structured site_specs 'evidence_base' aspect.
type EvidenceBase struct {
	AuditDoc        string         `json:"audit_doc"`
	GoverningRule   string         `json:"governing_rule,omitempty"`
	Facts           []EvidenceFact `json:"facts"`
	BannedClaims    []BannedClaim  `json:"banned_claims"`
	AllowedEntities []string       `json:"allowed_entities,omitempty"`
}

// ParseEvidenceBase decodes the site_specs data JSONB into an EvidenceBase and
// compiles the banned patterns. Returns nil (no error) for an evidence base
// with nothing scannable — callers treat nil as "site not opted in".
// A banned pattern that fails to compile as a regex is kept as a literal
// case-insensitive substring — a typo must never silently drop a banned claim.
func ParseEvidenceBase(data []byte) (*EvidenceBase, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var eb EvidenceBase
	if err := json.Unmarshal(data, &eb); err != nil {
		return nil, fmt.Errorf("evidence_base unmarshal: %w", err)
	}
	if len(eb.Facts) == 0 && len(eb.BannedClaims) == 0 {
		return nil, nil
	}
	for i := range eb.BannedClaims {
		p := eb.BannedClaims[i].Pattern
		re, err := regexp.Compile("(?i)" + p)
		if err != nil {
			re = regexp.MustCompile("(?i)" + regexp.QuoteMeta(p))
		}
		eb.BannedClaims[i].re = re
	}
	return &eb, nil
}

// ============================================================================
// Assertion-text extraction
// ============================================================================

// Elements whose entire subtree is a non-assertion context: machine content,
// samples, and UI internals. Text here is never a claim about the business.
var nonAssertionElements = map[string]bool{
	"script": true, "style": true, "noscript": true, "template": true,
	"code": true, "pre": true, "svg": true, "iframe": true,
	"textarea": true, "select": true, "option": true, "head": true,
}

// Elements that delimit assertion blocks. Text inside inline elements
// (<strong>, <a>, <em>, <span>…) is concatenated with its surroundings so a
// claim split across inline markup ("<strong>70+</strong> agents") still
// reads as one assertion.
var assertionBlockElements = map[string]bool{
	"p": true, "div": true, "section": true, "article": true, "main": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"li": true, "ul": true, "ol": true, "dl": true, "dt": true, "dd": true,
	"table": true, "tr": true, "td": true, "th": true, "caption": true,
	"blockquote": true, "figure": true, "figcaption": true,
	"header": true, "footer": true, "nav": true, "aside": true,
	"form": true, "fieldset": true, "label": true, "button": true,
	"details": true, "summary": true, "address": true,
	"br": true, "hr": true,
}

var wsCollapseRe = regexp.MustCompile(`\s+`)

// extractAssertions walks the parsed DOM once, returning the assertion text
// blocks and the mailto: addresses found in href attributes (the one
// attribute surface that asserts contact information).
func extractAssertions(htmlStr string) (blocks []string, mailtos []string) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		// html.Parse is lenient and effectively never errors on string input;
		// if it somehow does, fall back to a crude tag strip so scanning
		// degrades rather than silently passing everything.
		stripped := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(htmlStr, " ")
		return []string{wsCollapseRe.ReplaceAllString(strings.TrimSpace(stripped), " ")}, nil
	}

	var buf strings.Builder
	flush := func() {
		s := strings.TrimSpace(wsCollapseRe.ReplaceAllString(buf.String(), " "))
		if s != "" {
			blocks = append(blocks, s)
		}
		buf.Reset()
	}

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		switch n.Type {
		case html.ElementNode:
			if nonAssertionElements[n.Data] {
				return
			}
			if n.Data == "a" {
				for _, attr := range n.Attr {
					if attr.Key == "href" && strings.HasPrefix(strings.ToLower(strings.TrimSpace(attr.Val)), "mailto:") {
						addr := strings.TrimSpace(attr.Val[len("mailto:"):])
						if i := strings.IndexAny(addr, "?&"); i >= 0 {
							addr = addr[:i]
						}
						if addr != "" {
							mailtos = append(mailtos, addr)
						}
					}
				}
			}
			isBlock := assertionBlockElements[n.Data]
			if isBlock {
				flush()
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
			if isBlock {
				flush()
			}
		case html.TextNode:
			buf.WriteString(n.Data)
		case html.DocumentNode:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
		// Comments and doctypes are not assertions — skipped.
	}
	walk(doc)
	flush()
	return blocks, mailtos
}

// ExtractAssertionText returns the human-readable assertion blocks of an HTML
// fragment: text-node content outside non-assertion contexts, concatenated
// across inline elements and split at block boundaries.
func ExtractAssertionText(htmlStr string) []string {
	blocks, _ := extractAssertions(htmlStr)
	return blocks
}

var assertionEmailRe = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)

// ExtractAssertionEmails returns email addresses asserted by the page: those
// appearing in assertion text plus mailto: link targets. Emails that exist
// only in non-assertion contexts (placeholder= attributes, <code> samples,
// script bodies) are NOT returned — they are examples, not contact claims.
func ExtractAssertionEmails(htmlStr string) []string {
	blocks, mailtos := extractAssertions(htmlStr)
	seen := make(map[string]bool)
	var out []string
	add := func(e string) {
		e = strings.ToLower(strings.TrimSpace(e))
		if e == "" || seen[e] {
			return
		}
		seen[e] = true
		out = append(out, e)
	}
	for _, b := range blocks {
		for _, e := range assertionEmailRe.FindAllString(b, -1) {
			add(e)
		}
	}
	for _, m := range mailtos {
		if assertionEmailRe.MatchString(m) {
			add(m)
		}
	}
	return out
}

// ============================================================================
// Scans
// ============================================================================

// ClaimFinding is one violation found by a claims scan.
type ClaimFinding struct {
	Check       string `json:"check"`   // "banned_claim" | "unregistered_number"
	Matched     string `json:"matched"` // the offending text / number literal
	Pattern     string `json:"pattern,omitempty"`
	Reason      string `json:"reason"`
	Snippet     string `json:"snippet"`
	Occurrences int    `json:"occurrences"`
}

// ScanBannedClaims scans assertion blocks against the site's banned claim
// patterns. Every match is a KNOWN falsehood for this site (each pattern was
// audited out by a human) — callers treat findings as blockers.
//
// Matches whose claim is NEGATED in the same clause are dropped — see
// negatedClaimMatch, and ScanBannedClaimsIgnoringNegation to see what that
// suppressed.
func (eb *EvidenceBase) ScanBannedClaims(blocks []string) []ClaimFinding {
	return eb.scanBannedClaims(blocks, true)
}

// ScanBannedClaimsIgnoringNegation is the same scan with the negation guard
// DISABLED, i.e. the pre-2026-07-29 behaviour.
//
// It exists so the guard cannot suppress silently. A guard that quietly drops
// matches is indistinguishable from a gate that has stopped working, and this
// estate has already been bitten by exactly that shape (a declared allow-list
// that hid the case it was written to catch). cmd/claimscan diffs the two so an
// operator can see every match the guard removed; nothing enforces this result.
func (eb *EvidenceBase) ScanBannedClaimsIgnoringNegation(blocks []string) []ClaimFinding {
	return eb.scanBannedClaims(blocks, false)
}

func (eb *EvidenceBase) scanBannedClaims(blocks []string, guardNegation bool) []ClaimFinding {
	if eb == nil || len(eb.BannedClaims) == 0 {
		return nil
	}
	type agg struct {
		finding ClaimFinding
	}
	found := make(map[string]*agg)
	var order []string

	for _, block := range blocks {
		for i := range eb.BannedClaims {
			bc := &eb.BannedClaims[i]
			if bc.re == nil {
				continue
			}
			locs := bc.re.FindAllStringIndex(block, -1)
			if len(locs) == 0 {
				continue
			}
			// Keep only the matches that are actually asserted. A block whose
			// every match is negated yields no finding at all — the same
			// outcome as the pattern not matching, because the sentence is not
			// making the claim.
			live := locs
			if guardNegation {
				live = make([][]int, 0, len(locs))
				for _, loc := range locs {
					if negatedClaimMatch(block, loc[0]) {
						continue
					}
					live = append(live, loc)
				}
				if len(live) == 0 {
					continue
				}
			}
			if a, ok := found[bc.Pattern]; ok {
				a.finding.Occurrences += len(live)
				continue
			}
			loc := live[0]
			found[bc.Pattern] = &agg{finding: ClaimFinding{
				Check:       "banned_claim",
				Matched:     block[loc[0]:loc[1]],
				Pattern:     bc.Pattern,
				Reason:      bc.Reason,
				Snippet:     claimSnippet(block, loc[0], loc[1]),
				Occurrences: len(live),
			}}
			order = append(order, bc.Pattern)
		}
	}

	findings := make([]ClaimFinding, 0, len(order))
	for _, p := range order {
		findings = append(findings, found[p].finding)
	}
	return findings
}

// ============================================================================
// The negation guard (bugs_closed/104 follow-up, 2026-07-29)
// ============================================================================
//
// WHY THIS EXISTS
//   A banned-claim pattern matches a PHRASE, but what is banned is an
//   ASSERTION. "Every figure is independently verified" is a fabrication;
//   "where a figure has not been independently verified, that is stated" is the
//   hedged disclosure this whole layer exists to encourage — and it contains the
//   banned phrase verbatim. At severity blocker, the second one fails a page
//   build for being honest.
//
//   That is not hypothetical. The 2026-07-28 fleet-wide dry run over 908 live
//   components produced 7 findings and 4 were this exact class, all on negated
//   sentences, all from ONE pattern — which was therefore left OUT of the
//   fleet-wide set with a note saying it could not return until a guard existed
//   in code, because Go's RE2 has no lookbehind. This is that guard, and the
//   pattern is back (claims_global.go).
//
// WHAT IT DOES NOT DO
//   It is a clause-local cue test, not a parser. It cannot read "we would never
//   claim our analysis is anything other than accurate". It is deliberately
//   NARROW: a guard that is too eager silently disarms the gate, which is worse
//   than the false positive it set out to fix, and harder to notice. Every cue
//   below is one that negates a following predicate in English; the near-misses
//   that look like cues and are not are listed with their counter-examples, and
//   the whole suppression is observable via ScanBannedClaimsIgnoringNegation.

// negationCueRe matches a cue that negates a claim following it in the same
// clause.
//
// DELIBERATELY ABSENT, each because it disarms a real overclaim:
//
//	"without" — "Without exception, every claim is verified." (intensifier)
//	bare "no"  — "There are no exceptions: every claim is verified." (ditto)
//	"rarely", "seldom", "hardly" — hedges; "our data is rarely wrong" is still
//	                              an accuracy claim and should still be caught.
//
// Their absence is a KNOWN residual: "no figure here is independently verified"
// would still be a false positive. No sentence of that shape exists in the live
// corpus (checked 2026-07-29 over all 908 components), and the cost of getting
// it wrong in the other direction is a silently weaker gate.
// The contraction arm is written as "letter + n't" rather than a list of stems,
// because the stems are irregular ("can't" is ca+n't, "won't" is wo+n't) and a
// stem list gets one wrong silently. Curly apostrophes are included: site copy
// is typographic, and `’` is what a renderer actually emits.
var negationCueRe = regexp.MustCompile(
	`(?i)\b(?:not|never|nor|cannot|unable to|fails? to|failed to|refuses? to)\b` +
		`|[a-z]n['’‘]t\b` +
		`|\b(?:cant|dont|doesnt|didnt|isnt|arent|wasnt|werent|hasnt|havent|hadnt|wont|couldnt|shouldnt|wouldnt)\b`,
)

// negationClauseBoundary ends the backwards search. The cue must be in the same
// clause as the match: a comma or stronger stops the scan, so
// "we do not use AI, and our analysis is always accurate" is still caught. HTML
// angle brackets are included because assertion blocks are not always fully
// tag-stripped, and a tag is a boundary in any sensible reading.
const negationClauseBoundary = ".!?;:,<>\n\r\t–—"

// negationWindowBytes bounds how far back the cue may sit. 64 bytes clears the
// longest real form measured ("has not been " is 13) with room to spare, while
// keeping the scan cheap: this runs per match, per pattern, per block.
const negationWindowBytes = 64

// negatedClaimMatch reports whether the banned phrase starting at `start` is
// negated by a cue earlier in the same clause.
func negatedClaimMatch(block string, start int) bool {
	window := block[maxInt(0, start-negationWindowBytes):start]

	// Trim to the current clause. LastIndexAny returns the byte index of the
	// START of a multibyte boundary rune, so step over the whole rune rather
	// than one byte — otherwise the window keeps a dash's trailing bytes.
	if i := strings.LastIndexAny(window, negationClauseBoundary); i >= 0 {
		_, size := utf8.DecodeRuneInString(window[i:])
		window = window[i+size:]
	}
	return negationCueRe.MatchString(window)
}

// Number-bearing claim candidates: integers with thousands separators or
// plain numbers (decimals included).
var numberCandidateRe = regexp.MustCompile(`\d{1,3}(?:,\d{3})+(?:\.\d+)?|\d+(?:\.\d+)?`)

// A number is only a business claim when its surrounding window talks about
// the business. This gate is the main precision control: hypotheticals in
// guides ("100,000 daily calls"), formulas ("× 100"), and UI text never pass
// it. Widen only from measurement (spec landmine 2).
// Note "businesses" is plural-only: count-claims about businesses use the
// plural, while singular "business" is descriptive ("business hours",
// "business functions") and false-positived on calculator help text
// ("22 for business hours") in the first live run.
var businessClaimContextRe = regexp.MustCompile(`(?i)\b(clients?|customers?|records?|businesses|compan(y|ies)|agents?|sites?|users?|subscribers?|departments?|awards?|employees?|staff|engagements?|projects?|deployments?|case\s+stud(y|ies)|definitions?|orchestration|integrations?|providers?|items?|uptime|verified|enrich(ed|ment)|scored|collected|processed|deployed|delivered|years\s+of\s+experience|uniques?)\b`)

// Phone-context exclusion — phone numbers are validated separately and their
// digit groups must not reach the number scan.
var phoneContextRe = regexp.MustCompile(`(?i)(\bphone\b|\bcall\b|\btel\b|\+\d{1,3}\s?\(0\)|\+44|\+1\s?\()`)

// Label-prefixed numbers ("Band 3", "Tier 2", "Step 1") are ordinals, not claims.
var labelPrefixRe = regexp.MustCompile(`(?i)\b(band|tier|step|phase|part|stage|level|section|chapter|question|rule|point|option|version|v)\s*$`)

// A day number in a written-out date: "28 July 2026", "1 January", "3rd March".
// The composite-token test above catches ISO dates (2026-07-28) but not dates
// written the way British English actually writes them, which is the platform's
// stated convention — so without this every naturally-worded date on every site
// raises an unregistered-number finding. Noise is not harmless in a checker: a
// scanner that always reports something is one people stop reading.
var writtenDateRe = regexp.MustCompile(`(?i)^\s*(st|nd|rd|th)?\s+(january|february|march|april|may|june|july|august|september|october|november|december)\b`)

// The same date with the month first: "July 28, 2026".
var monthBeforeRe = regexp.MustCompile(`(?i)(january|february|march|april|may|june|july|august|september|october|november|december)\s*$`)

// Unit/measurement suffixes that mark a number as not-a-business-count.
var unitSuffixRe = regexp.MustCompile(`(?i)^\s*(px|rem|em|vh|vw|ms|sec|seconds?|min(ute)?s?\s+read|kb|mb|gb|tb|fps|st\b|nd\b|rd\b|th\b|[-–]\s*(hour|day|week|month|year|minute|second|token|character|step|person|page)\b)`)

// ============================================================================
// Claim surface — the STRUCTURAL gate on the prose number scan (bugs_open/102)
// ============================================================================
//
// businessClaimContextRe above is a LEXICAL gate: it asks whether the words
// around a number sound like business. On a marketing page that is the right
// question, because nearly every page asserts something about the business. On a
// page whose body is teaching content it is the wrong question, because an
// explainer's worked example is lexically identical to a sales claim — "10,000
// active players farming that item" reads exactly like "10,000 customers".
//
// The signal that separates them is not in the words, it is in the schema:
// pages.page_type says what kind of page the number is on. This is the same
// argument claims_stats.go makes one level down — a figure in a stat*_value
// field is a claim BY CONSTRUCTION, so structural position replaces the lexical
// gate there. Page type is structural position one level UP, and until
// 2026-07-28 the layer read it nowhere.
//
// MEASURED before this was built (2026-07-28, cmd/claimscan against each
// opted-in site's own live register over live rendered_html): 124
// unregistered-number findings across the nine sites with an evidence_base row.
// 47 of them sit on the page types below and ALL 47 are false positives —
// probability worked examples on gamesdesign blog posts, "an endpoint that
// returns 200" on an ai-agent-orchestration post, "Set to 0 to disable" in tool
// help text, a quoted third-party market share on a news index. The findings on
// content/landing pages are the real ones the layer exists for. A finding class
// with 0% measured precision is not a weaker finding, it is not a finding: the
// gate files these at `error` severity, which BLOCKS a rebuild, and the audit
// files them into a human queue that has no working surface (bugs_open/033).

// ClaimSurface describes WHERE scanned text came from, so the scans can apply
// structural knowledge the text itself does not carry.
//
// The zero value means UNKNOWN — site chrome (which belongs to no page), or a
// caller with no page record in hand — and unknown is scanned exactly as before.
// That direction is deliberate: a scanner that has gone quiet and one that is
// broken look identical from the outside, so an unrecognised or absent page type
// must stay noisy rather than silently stop checking.
type ClaimSurface struct {
	// PageType is pages.page_type for the page this text renders on.
	PageType string
}

// editorialPageTypes are the page types whose BODY PROSE is not a first-person
// claim about the business — instruction, commentary, aggregated third-party
// listings, or an interactive instrument's own help text.
//
// THE BAR FOR MEMBERSHIP: a measured false positive on live copy, AND a body
// that is never marketing. Do not widen this from intuition. The literals are
// the live `pages.page_type` vocabulary, checked against the column on
// 2026-07-28 (content 130, tool 110, blog-post 67, guide 52, section-index 20,
// entity-page 20, landing 17, game 5, news-index 4, blog-index 3, report 2,
// entity-directory 2, adoption-tracker 1, protocol-tracker 1, model-directory 1)
// — a literal that does not match a stored value is a silent no-op, so re-run
// that query before adding one.
//
// Each member, with what earns it (measured 2026-07-28, cmd/claimscan against
// each opted-in site's own register over live rendered_html):
//   - blog-post (46 false positives) — worked examples in explainers
//   - tool (7) — an instrument's own help text: "Set to 0 to disable"
//   - game (4) — the same, on interactive pages
//   - guide (1 live, plus the 15 on webdesign.co.uk that motivated bugs_open/102)
//   - news-index (1) — a quoted third-party market figure in a listing
//
// THREE PAGE TYPES ARE DELIBERATELY ABSENT, each for a different reason:
//
//   - 'blog-index': never measured. It exists on three sites and has raised zero
//     findings even scanned against an EMPTY register, so there is no evidence
//     either way. Adding it by analogy to blog-post would be exactly the
//     unmeasured extrapolation this comment warns against (council round 1,
//     correlation de4a19f5, edit-quality seat — the objection was right).
//   - 'section-index': measured, and REJECTED on the second half of the bar.
//     Two of its twenty pages are `about-index` and `contact-index` on
//     gamesdesign.co.uk — marketing surfaces filed under an index name. Its two
//     false positives are one quoted market-share sentence; a blind spot over
//     an about page is the worse trade.
//   - 'report': its 14 findings on robot-hands are a different class entirely —
//     model numbers inside product names ("Schunk EGP 40-N-S-B — manufacturer
//     specification") tripping on `verified` in the context window. Excluding it
//     here would fix those by coincidence, not by mechanism.
var editorialPageTypes = map[string]bool{
	"guide":      true,
	"blog-post":  true,
	"news-index": true,
	"tool":       true,
	"game":       true,
}

// ProseNumbersAreClaims reports whether the heuristic number scan applies to
// prose on this surface. It governs ScanUnregisteredNumbers ONLY:
//
//   - ScanBannedClaims runs on every surface. A banned pattern is a human-authored
//     record of a KNOWN falsehood, not a heuristic, so it has no false-positive
//     problem to protect against — and the case that motivated the whole check
//     was a banned claim found on a GUIDE (check_unverified_claims.go, first live
//     run 2026-07-16: "70+ agents across eight functional departments"). Skipping
//     editorial pages wholesale would have regressed exactly that catch.
//   - ScanStatClaims runs on every surface. A stat card on a guide is still a
//     published figure in a claim-shaped field (claims_stats.go).
func (s ClaimSurface) ProseNumbersAreClaims() bool {
	return !editorialPageTypes[strings.ToLower(strings.TrimSpace(s.PageType))]
}

// ScanUnregisteredNumbers extracts number-bearing business claims from
// assertion blocks and flags numbers no registered fact supports. False
// positives are possible by design (severity error → human review, never a
// blocker).
//
// surface is REQUIRED rather than optional (no ...Surface variant, no second
// entry point) so the compiler visits every call site. An "unknown means the old
// behaviour" default parameter would let a new caller silently inherit the
// page-type-blind scan, which is the shape of bugs_open/093 — one guarded call
// site and one nobody remembered.
func (eb *EvidenceBase) ScanUnregisteredNumbers(blocks []string, surface ClaimSurface) []ClaimFinding {
	if eb == nil || !surface.ProseNumbersAreClaims() {
		return nil
	}
	found := make(map[string]*ClaimFinding)
	var order []string

	for _, block := range blocks {
		for _, loc := range numberCandidateRe.FindAllStringIndex(block, -1) {
			token := block[loc[0]:loc[1]]
			if isExcludedNumber(block, loc[0], loc[1]) {
				continue
			}
			window := claimWindow(block, loc[0], loc[1])
			if !businessClaimContextRe.MatchString(window) {
				continue
			}
			if phoneContextRe.MatchString(window) {
				continue
			}
			val, ok := parseClaimNumber(token, block[loc[1]:])
			if !ok {
				continue
			}
			if eb.numberSupported(val, window) {
				continue
			}
			key := strings.ToLower(token)
			if f, ok := found[key]; ok {
				f.Occurrences++
				continue
			}
			found[key] = &ClaimFinding{
				Check:       "unregistered_number",
				Matched:     token,
				Reason:      "number asserted as a fact about the business matches no evidence_base fact value",
				Snippet:     claimSnippet(block, loc[0], loc[1]),
				Occurrences: 1,
			}
			order = append(order, key)
		}
	}

	findings := make([]ClaimFinding, 0, len(order))
	for _, k := range order {
		findings = append(findings, *found[k])
	}
	return findings
}

// isExcludedNumber applies the structural false-positive exclusions from the
// spec's landmine list: years, dates/times/versions/phone fragments (composite
// digit tokens), list ordinals, measurements, and currency amounts.
func isExcludedNumber(block string, start, end int) bool {
	token := block[start:end]

	// Composite tokens: a digit, '-', '/', ':' or '.' immediately adjacent
	// means this is part of a date, time, version, ratio, or phone group
	// (e.g. 2026-07-16, 24/7, 14:30, v1.0.1124).
	if start > 0 {
		prev := block[start-1]
		if prev == '-' || prev == '/' || prev == ':' || prev == '.' || (prev >= '0' && prev <= '9') {
			return true
		}
		if prev == 'v' || prev == 'V' {
			return true // version numbers
		}
		if prev == '$' || prev == 163 || prev == 128 { // '£' and '€' are multibyte; see below
			return true
		}
		// Multibyte currency symbols, and multibyte range dashes: check the
		// preceding rune. The dash cases are here rather than in the byte test
		// above for the same reason as the currency ones — '–' and '—' are
		// three bytes, so `prev == '-'` cannot see them.
		r := []rune(block[maxInt(0, start-3):start])
		if len(r) > 0 {
			switch r[len(r)-1] {
			case '£', '€', '$':
				return true
			case '–', '—':
				return true // the tail of a range: "8–12"
			}
		}
	}
	if end < len(block) {
		next := block[end]
		if next == '-' || next == '/' || next == ':' {
			// Only composite when followed by another digit (so "6-hour" is
			// handled by the unit rule, not swallowed here).
			if end+1 < len(block) && block[end+1] >= '0' && block[end+1] <= '9' {
				return true
			}
		}
		// Same rule for the typographic dashes.
		//
		// FOUND BY MEASUREMENT 2026-07-26, building bugs_open/093's second call
		// site: a live sweep of stored content_data flagged fundamentallyai's
		// "Read time: 8–12 minutes" as an unregistered business figure, on a
		// site with 15 registered facts — i.e. at `error` severity in the build
		// gate, which would have made a deployed page unbuildable for carrying
		// a reading-time estimate. That is bugs_closed/073's shape on a new
		// trigger. The hyphen form "8-12 minutes" was already excluded here;
		// only the en-dash escaped, and unitSuffixRe's own `[-–]` alternative
		// shows typographic dashes were always meant to be in scope.
		//
		// The trade this makes explicit: a range is excluded ENTIRELY, so
		// "2–3 million users" is no longer examined. That is not new blindness
		// — "2-3 million users" has always been treated that way — but it is a
		// real limit, and the honest place to note it is here.
		if rn, size := utf8.DecodeRuneInString(block[end:]); rn == '–' || rn == '—' {
			if i := end + size; i < len(block) && block[i] >= '0' && block[i] <= '9' {
				return true
			}
		}
	}

	// Standalone year.
	if len(token) == 4 && !strings.Contains(token, ",") {
		if v, err := strconv.Atoi(token); err == nil && v >= 1900 && v <= 2099 {
			return true
		}
	}

	// List ordinal at block start: "1. Volume of Repeatable Work Items".
	if start == 0 && end < len(block) && (block[end] == '.' || block[end] == ')') {
		return true
	}

	// Label-prefixed ordinal: "Band 3", "Tier 2".
	if labelPrefixRe.MatchString(block[:start]) {
		return true
	}

	// Measurement / ordinal / duration suffixes.
	if unitSuffixRe.MatchString(block[end:]) {
		return true
	}

	// A day number in a written-out date. Bounded to plausible day numbers so a
	// real business figure sitting next to a month name ("we placed 450 March
	// orders") is still scanned.
	if v, err := strconv.Atoi(token); err == nil && v >= 1 && v <= 31 {
		if writtenDateRe.MatchString(block[end:]) {
			return true
		}
		if monthBeforeRe.MatchString(block[:start]) {
			return true
		}
	}

	return false
}

// parseClaimNumber converts a matched token (plus any magnitude word that
// follows it) to a float: "2,767" → 2767, "12 million" → 12e6. Returns
// ok=false for unparseable tokens.
func parseClaimNumber(token, after string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.ReplaceAll(token, ",", ""), 64)
	if err != nil {
		return 0, false
	}
	rest := strings.ToLower(strings.TrimLeft(after, " \t\u00a0"))
	switch {
	case strings.HasPrefix(rest, "million"):
		v *= 1e6
	case strings.HasPrefix(rest, "billion"):
		v *= 1e9
	case strings.HasPrefix(rest, "thousand"):
		v *= 1e3
	}
	return v, true
}

// numberSupported reports whether any registered fact supports the published
// number in this claim window. Non-exact tolerances require the window to
// match one of the fact's context terms; a non-exact fact with no context
// terms degrades to exact matching (never to blanket support).
func (eb *EvidenceBase) numberSupported(val float64, window string) bool {
	lower := strings.ToLower(window)
	for i := range eb.Facts {
		f := &eb.Facts[i]
		// A series fact has no single Value; its observations are the registered
		// numbers. Without this branch every plotted point on a chart would be
		// reported as an unregistered number, and the honesty layer would be
		// fighting the rendering layer instead of backing it.
		if f.Value == nil && f.IsSeries() {
			if f.seriesSupports(val, lower) {
				return true
			}
			continue
		}
		if f.Value == nil {
			continue
		}
		if len(f.ContextTerms) > 0 {
			matched := false
			for _, t := range f.ContextTerms {
				if strings.Contains(lower, strings.ToLower(t)) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		tol := f.Tolerance
		if tol != "exact" && tol != "" && len(f.ContextTerms) == 0 {
			tol = "exact"
		}
		switch {
		case tol == "gte":
			if val <= *f.Value+1e-6 {
				return true
			}
		case strings.HasPrefix(tol, "approx_pct:"):
			pctStr := strings.TrimPrefix(tol, "approx_pct:")
			pct, err := strconv.ParseFloat(pctStr, 64)
			if err == nil && *f.Value != 0 &&
				math.Abs(val-*f.Value)/math.Abs(*f.Value)*100 <= pct {
				return true
			}
		default: // exact
			if math.Abs(val-*f.Value) < 1e-6 {
				return true
			}
		}
	}
	return false
}

// claimWindow returns the surrounding context used for gating decisions.
func claimWindow(block string, start, end int) string {
	ws := maxInt(0, start-70)
	we := minInt(len(block), end+70)
	return block[ws:we]
}

// claimSnippet returns a trimmed human-readable excerpt around a match.
func claimSnippet(block string, start, end int) string {
	ws := maxInt(0, start-60)
	we := minInt(len(block), end+60)
	return strings.TrimSpace(block[ws:we])
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
