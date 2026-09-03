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
	// RetainHistory opts this fact into remembering its superseded readings, and
	// governs BOTH halves of the mechanism: the refresh only records history for
	// an armed fact, and numberSupported only consults it for one. Default false,
	// because the unsafe side — widening what the scan accepts — must be the side
	// a human turns on per fact (bugs_open/386; the 2026-08-02 ruling on new
	// authority on a shared seam). A History array with this flag absent is inert
	// by design: seeding one must not widen support on its own.
	RetainHistory bool `json:"retain_history,omitempty"`
	// History carries this fact's superseded readings — the values the register
	// actually held, each with the date it was last verified at. It is NOT a
	// series (see claims_series.go): a series is one assertion measured
	// repeatedly and is the page's subject, whereas this is the same single
	// assertion's own past, kept so a page rendered while a value was current
	// is not convicted of inventing it. Deliberately a distinct field, because
	// IsSeries() keys on len(Observations) and reusing that slot would silently
	// turn every armed fact into a series.
	History []FactHistoryEntry `json:"history,omitempty"`
}

// FactHistoryEntry is one superseded reading of a fact.
//
// It carries no Source of its own, unlike Observation. That is not an oversight
// and it is the reason this type is separate: an observation is evidence a human
// or a query asserted about a point in time, and must prove itself. A history
// entry is a record of what THIS platform's register previously held — its
// provenance is the parent fact's source plus the date, and inventing a
// per-entry source would imply an independent attestation nobody made.
type FactHistoryEntry struct {
	Value      float64 `json:"value"`
	VerifiedAt string  `json:"verified_at"`
}

// FactHistoryMaxEntries caps a fact's retained history.
//
// A cap is required rather than tidy: every retained reading is a number the
// scan will accept near this fact's context terms, so an unbounded history
// converts a precise check into a wide one over time. 90 entries is about three
// months at the daily refresh cadence measured 2026-08-25 — long enough to cover
// any page still carrying a stale render, short enough that the accepted set
// stays small and reviewable.
const FactHistoryMaxEntries = 90

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

	// Regulated, when present AND complete, exempts this site from the
	// fleet-wide regulated-identity family (claims_regulated.go) — the one
	// legitimate way for a site to describe itself as an authorised firm.
	// Absent or incomplete means NOT exempt; see RegulatedAttestation.Attested.
	Regulated *RegulatedAttestation `json:"regulated,omitempty"`

	// OperatingHistory, when present AND complete, exempts this site from the
	// fleet-wide practice-claims family (claims_practice.go) — the one
	// legitimate way for a site to state that it tests, buys, measures or is
	// sent products. Absent or incomplete means NOT exempt (bugs_open/380).
	OperatingHistory *OperatingHistoryAttestation `json:"operating_history,omitempty"`

	// CitationCodes (RFC_060 Q5, owner-ruled 2026-09-03) names ad hoc
	// regulatory rulebook codes this site cites, ADDITIVE over the
	// fleet-wide default (regulatoryRulebookCodesBase) — never a
	// replacement, so an absent field regresses nobody. Same purpose and
	// same shape as BannedClaims/AllowedEntities: site-declared data, no Go
	// change to onboard a new one. A code of two characters or fewer is
	// silently dropped at compile time (compileCitationCodeRegexes) — even
	// site-declared, it is "too collidable" per fad209b92's own reasoning
	// for excluding bare two-letter codes fleet-wide.
	CitationCodes []string `json:"citation_codes,omitempty"`

	// CitationCodePresets opts this site into a NAMED sector's rulebook
	// codes (KnownCitationCodePresets lists the valid names) instead of
	// typing them out via CitationCodes. RFC_060 Q5: the owner named
	// veterinary and legal as imminent second/third consumers, which is
	// what turned the sector-preset half of the design from "hold it back"
	// to "build it now, with those two in mind".
	CitationCodePresets []string `json:"citation_code_presets,omitempty"`

	// citationPrefixRe and citationContextRe are this site's OWN compiled
	// citation-recognition regexes — regulatoryRulebookCodesBase unioned
	// with CitationCodePresets (expanded) and CitationCodes — compiled once
	// at parse time (compileCitationCodeRegexes) exactly as BannedClaims[i].re
	// is. A site with neither field set compiles to patterns BYTE IDENTICAL
	// to fad209b92's fleet-wide default.
	citationPrefixRe  *regexp.Regexp
	citationContextRe *regexp.Regexp

	// MalformedFacts names the facts in this register that did not decode,
	// and is the reason facts are decoded ONE AT A TIME below.
	//
	// It is not populated from the register's JSON — it is produced by the
	// parse, so it carries no struct tag and never round-trips into a
	// register a consumer writes back.
	MalformedFacts []MalformedFact `json:"-"`
}

// MalformedFact is one fact this register carries that could not be decoded.
//
// WHY THIS TYPE EXISTS, because a reader will otherwise assume the parse is
// simply lenient now. Until 2026-09-03 ParseEvidenceBase decoded `facts` as
// one array, so a SINGLE undecodable fact returned an error for the WHOLE
// base — and every caller treats that error as "this site has no register".
// The three claims gates (validate_page_content.go, its stat audit, and
// check_unverified_claims.go) then skip the site entirely, INCLUDING its
// banned_claims, which do not depend on facts at all. Measured 2026-09-03,
// that had silently disarmed two live registers and 10 banned claims:
// finetuning.uk (3 bans, since 08-24) and noted.co.uk (7 bans, since 08-25),
// whose ban list forbids exactly the unearnable security absolutes a notes
// product must not claim. Both were found by an audit, not by any signal.
//
// The cause was not a typo. EvidenceFact.Value is a *float64 and authors were
// legitimately registering text-valued facts ("MIT", "Apache 2.0",
// "Llama 3.3 Community License"), which is a shape the register has no way to
// hold. So the failure was reasonable authorship meeting a missing capability,
// and the blast radius was the entire site's guard list.
//
// A malformed fact now costs you THAT FACT. Everything else in the register —
// every ban, both attestations, the citation codes — parses and stays armed.
type MalformedFact struct {
	// Index is the fact's position in the register's `facts` array. It is what
	// a human needs to find the row, and it is available even when the fact is
	// too broken to have a readable id.
	Index int `json:"index"`
	// ID is the fact's `id` if it could be read as a string, else "". Read
	// separately from the failed decode precisely because the decode failed:
	// one bad field must not cost us the name of the fact carrying it.
	ID string `json:"id,omitempty"`
	// Err is the decode error verbatim, which names the offending field and
	// its type (e.g. "cannot unmarshal string into ... .value of type
	// float64"). Verbatim because paraphrasing it would lose the one detail
	// that makes the row findable.
	Err string `json:"error"`
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
	// Facts are decoded ONE AT A TIME, and that is the whole point — see
	// MalformedFact's own doc comment for the two live registers this was
	// found on. `facts` is shadowed here as raw JSON (the outer field wins
	// over the embedded struct's, being at the shallower depth) so that one
	// undecodable fact costs that fact and nothing else. Every other field,
	// banned_claims above all, decodes exactly as it always did.
	//
	// This can only ever make a register parse where it previously did not.
	// There is no input for which it returns an error that the old code
	// accepted: the top-level decode below is the same call on the same
	// bytes, minus the facts array.
	type evidenceBaseWire struct {
		EvidenceBase
		Facts []json.RawMessage `json:"facts"`
	}
	var wire evidenceBaseWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("evidence_base unmarshal: %w", err)
	}
	eb := wire.EvidenceBase
	eb.Facts = nil
	for i, raw := range wire.Facts {
		var f EvidenceFact
		if err := json.Unmarshal(raw, &f); err != nil {
			eb.MalformedFacts = append(eb.MalformedFacts, MalformedFact{
				Index: i,
				ID:    rawFactID(raw),
				Err:   err.Error(),
			})
			continue
		}
		eb.Facts = append(eb.Facts, f)
	}
	// A site carrying ONLY a regulated attestation must still parse to a
	// non-nil base, or its attestation would be silently discarded here and the
	// exemption would never fire — the caller would see "site not opted in" and
	// apply the regulated family anyway. Failing safe is right; failing safe
	// while the operator believes they have recorded an attestation is not.
	// The same holds for an operating-history attestation (bugs_open/380): an
	// attestation-only base must parse non-nil or the practice-claims exemption
	// silently never fires. Neither attestation makes the base SCANNABLE — see
	// HasScannableRegister, which is what the register-comparison scans key on.
	// len(eb.MalformedFacts) is part of this test for the same reason the two
	// attestations are: returning nil here means "site not opted in", and a
	// register whose ONLY content is facts that failed to decode is a site
	// that opted in and got nothing. Without this clause the single signal
	// that something is wrong would be discarded exactly where it is most
	// needed — a register with no bans to save it.
	if len(eb.Facts) == 0 && len(eb.BannedClaims) == 0 && eb.Regulated == nil &&
		eb.OperatingHistory == nil && len(eb.MalformedFacts) == 0 {
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
	eb.compileCitationCodeRegexes()
	return &eb, nil
}

// rawFactID reads just the `id` of a fact that failed to decode as a whole.
//
// It decodes into a map rather than a struct with one string field, because
// the fact failed on a TYPE mismatch and a struct would fail again on the
// same field. Any failure here yields "", which the caller reports as an
// unnamed fact at a known index — never a wrong name.
func rawFactID(raw []byte) string {
	var probe map[string]interface{}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return ""
	}
	id, _ := probe["id"].(string)
	return id
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

// NegationGuard is the clause-local negation test behind negatedClaimMatch,
// factored out (bugs_open/222) so a second domain can reuse the ALGORITHM
// without inheriting this package's CUE VOCABULARY.
//
// What is shared: a bounded backwards window, trimmed to the current clause,
// tested against a cue regex — including the multibyte-rune handling in
// NegatedAt, which cost a real bug to get right (a dash's trailing bytes
// surviving the trim). What is deliberately NOT shared: negationCueRe itself.
// It excludes bare "no"/"without" on purpose (see the doctrine block above) —
// intensifiers in marketing prose, not negators — and that exclusion is
// pinned by TestBareNoIsAKnownResidualOfTheSharedGuard. A different domain
// (check_tool_fabrication_action.go's code-comment denials) needs those cues
// and builds its OWN NegationGuard with its own vocabulary; it must never
// reach into this package's negationCueRe, and this package's guard must
// never be widened to serve it. Two vocabularies, one algorithm — the same
// anti-drift shape CLM-004 states for scan implementations generally.
type NegationGuard struct {
	Cue      *regexp.Regexp // matches a cue that negates a following token
	Boundary string         // clause-boundary characters that end the backwards scan
	Window   int            // maximum bytes scanned back from the token
}

// NegatedAt reports whether the token starting at byte offset pos in text is
// negated by a cue earlier in the same clause.
func (g NegationGuard) NegatedAt(text string, pos int) bool {
	window := text[maxInt(0, pos-g.Window):pos]

	// Trim to the current clause. LastIndexAny returns the byte index of the
	// START of a multibyte boundary rune, so step over the whole rune rather
	// than one byte — otherwise the window keeps a dash's trailing bytes.
	if i := strings.LastIndexAny(window, g.Boundary); i >= 0 {
		_, size := utf8.DecodeRuneInString(window[i:])
		window = window[i+size:]
	}
	return g.Cue.MatchString(window)
}

// claimNegationGuard binds the shared algorithm to this package's vocabulary.
var claimNegationGuard = NegationGuard{Cue: negationCueRe, Boundary: negationClauseBoundary, Window: negationWindowBytes}

// negatedClaimMatch reports whether the banned phrase starting at `start` is
// negated by a cue earlier in the same clause.
func negatedClaimMatch(block string, start int) bool {
	return claimNegationGuard.NegatedAt(block, start)
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
//
// `orchestration` GAINED ITS PLURAL 2026-08-24 (bugs_open/364), and the reason
// is worth more than the character: it was the ONLY countable noun in this list
// without one (audited all 35 alternatives — the rest are already `s?`/`(y|ies)`
// grouped, or are mass nouns (`staff`, `uptime`), verbs (`verified`, `collected`)
// or fixed phrases). Because this gate uses `\b…\b` while numberSupported's
// ContextTerms use strings.Contains, a registered fact whose term is
// "orchestration" HAPPILY VOUCHES for a number next to "orchestrations" that
// this gate never let through in the first place — so the register looked like
// it was doing the work and the gate was simply blind. Live copy on
// ai-agent-orchestration.com reads "We run over 1,600 orchestrations a day
// across 13 live production systems": a first-person quantified claim that had
// NEVER been scanned.
//
// MEASURED before changing it, both directions: fleet-wide, exactly ONE site's
// components contain "orchestrations" at all (11 components, all
// ai-agent-orchestration.com [MEASURED 2026-08-24]) and adding the plural changes
// the finding count by ZERO — the values there are already supported by that
// site's registered `aao-orchestrations` gte fact.
//
// ⚠ "AT NO MEASURED COST" IS WHAT THIS ORIGINALLY SAID, AND IT WAS TRUE ONLY OF
// THE DAY IT WAS MEASURED. That fact is a ROLLING-WINDOW count, not a total: 35
// recorded snapshots range **1,494 to 7,281**, it has FALLEN 17 times across 34
// transitions, and it was **below 1,600 on 3 of the 35** (`bugs_open/386`'s
// census, 2026-08-25). The live copy it vouches for reads *"We run over 1,600
// orchestrations a day"*. Demonstrated by replaying the site's own historic low
// through this scan: identical copy scores **0 findings at 7,281 and a
// finding at 1,494** — at `error` severity, which refuses the page build.
//
// So the honest statement is: the plural fix costs nothing *while the counter is
// high*, and on a low day it surfaces claims that were always unsupported and
// merely un-scanned. Fleet-shaped: 4 such findings appear on that site at the
// historic low, 3 of them on the tracker pages Phase 2 newly scans.
//
// ⚠ AND THE FIX IS NOT WHERE I FIRST SAID IT WAS — corrected 2026-08-25 by the
// `bugs_open/386` lane, whose fix I had described wrongly. I wrote that their
// history retention "would keep vouching for 1,600 because 7,281 IS a value the
// register held". That is `gte` applied to HISTORY. `historySupports` is
// EXACT-match only, so it would not — and it SHOULD not: supporting anything at
// or below the all-time maximum means one busy day vouches for "over 7,000 a
// day" for ever on every quiet day after, which is this comment's own
// accidental-support hazard with time as the amplifier. Exact-only is precisely
// what makes history safe.
//
// THE REAL REMEDY NEEDS NO CODE, and the register already contains it:
// `aao-orchestrations`' own `writer_line` reads *"over a thousand orchestrations
// a day ({value} in the 24 hours to 2026-07-26)"*. **A floor of 1,000 sits BELOW
// the historic low of 1,494, so that instruction is safe on every day in the
// record.** The convicted copy says 1,600 / 1,699 / 1,834 — all ABOVE that low,
// so it deviates from the register's own instruction or predates it. So this is
// not an evidence-layer defect at all: it is the owner's "state a floor" ruling
// applied with the floor set too high. **The rule that generalises: a published
// floor must sit below the LOWEST recorded value, not below today's.**
//
// Recorded rather than papered over: a `[MEASURED]` claim about a moving value
// expires, and this one did — twice, because my first correction of it named the
// wrong fix.
//
// ⚠ That fact's VALUE is deliberately not quoted here any more. It read 4068 when
// this was written on 2026-08-24 and 7281 the next day — it is a live `count(*)`
// that climbs, so a bare number in a comment is stale within a day and reads as
// current for ever (CLAUDE.md: a count of things must carry the date it was
// counted; caught by the bugs_open/386 lane, which is the bug about exactly this
// drift). The load-bearing fact is not the number: it is that ONE gte fact with
// the single broad context term "orchestration" silently vouches for EVERY
// smaller figure near that word, and the ceiling rises on its own.
//
// ⚠ THE GENERAL POINT, which is the durable half: this list is an ALLOW-LIST OF
// NOUNS with the same unbounded-miss property as unitSuffixRe's allow-list of
// units. A plural, synonym or hyphenation it has not met is a SILENT MISS
// anywhere on the fleet — and unlike a false positive, nothing reports it.
// bugs_open/364 §5b.
var businessClaimContextRe = regexp.MustCompile(`(?i)\b(clients?|customers?|records?|businesses|compan(y|ies)|agents?|sites?|users?|subscribers?|departments?|awards?|employees?|staff|engagements?|projects?|deployments?|case\s+stud(y|ies)|definitions?|orchestrations?|integrations?|providers?|items?|uptime|verified|enrich(ed|ment)|scored|collected|processed|deployed|delivered|years\s+of\s+experience|uniques?)\b`)

// Phone-context exclusion — phone numbers are validated separately and their
// digit groups must not reach the number scan.
var phoneContextRe = regexp.MustCompile(`(?i)(\bphone\b|\bcall\b|\btel\b|\+\d{1,3}\s?\(0\)|\+44|\+1\s?\()`)

// Label-prefixed numbers ("Band 3", "Tier 2", "Step 1") are ordinals, not claims.
var labelPrefixRe = regexp.MustCompile(`(?i)\b(band|tier|step|phase|part|stage|level|section|chapter|question|rule|point|option|version|v)\s*$`)

// REGULATORY CITATIONS (bugs_closed/414 follow-up, 2026-09-02). On a
// finance site a rule reference is not a quantity, and there are two shapes.
//
// THE CODE LIST IS CASE-SENSITIVE AND THAT IS THE POINT. A real citation is
// written in capitals — "CONC 5A", never "conc 5a" — so requiring uppercase is
// what lets short codes like SUP, MAR and DISP be in the list at all. With
// `(?i)` they would swallow ordinary prose ("...on the map 5 miles..."), which
// is how a narrow exclusion turns into a silent hole in the scan.
//
// FCA Handbook sourcebooks plus the prudential ones. Two-letter codes (LR, PR,
// TC, EG) are DELIBERATELY ABSENT: even uppercase they are too collidable, and
// the measurement did not need them.
//
// ⚠ VERIFIED, not assumed — three council seats asked the same question
// independently (corr 1dd3d298, all medium) and they were right to: this file
// carries a landmine that "every banned_claims pattern is compiled with a forced
// (?i)", which would make the case-sensitivity above compile-time true and
// runtime false. Checked 2026-09-02, three ways:
//
//  1. the forced (?i) lives ONLY in the four BANNED-CLAIM compilers
//     (claims.go:~346 per-site, claims_global.go, claims_regulated.go,
//     claims_practice.go) — each wraps `bans[i].Pattern`. These two regexes are
//     not banned claims and never reach that path; they are compiled here, at
//     package level, with a bare regexp.MustCompile and no (?i). Contrast
//     labelPrefixRe directly above, which carries `(?i)` in its own literal —
//     the difference is visible in one screenful, deliberately.
//  2. nothing case-folds the scanned text before the exclusion tests run. The
//     ToLower calls in this file are fact-Kind normalisation, a mailto href
//     check, allowed-entity normalisation, the surface key, the finding DEDUP
//     key (applied after the tests) and the magnitude-word suffix (applied to
//     text after the number). None touches the block or the window.
//  3. behaviourally, which is the one that would survive a refactor of 1 and 2:
//     TestRegulatoryExclusionDoesNotLaunderBusinessNumbers asserts that
//     "We processed 12 conc records last year." is still CAUGHT. If a forced
//     (?i) ever reaches these patterns, that fixture starts passing silently
//     and that test starts failing loudly.
//
// regulatoryRulebookCodesBase is the fleet-wide default vocabulary —
// unchanged from fad209b92, still authored here in Go, still the codes
// EVERY site gets whether or not it declares anything of its own.
var regulatoryRulebookCodesBase = []string{
	"CONC", "MCOB", "ICOBS", "BCOBS", "COBS", "DISP", "SYSC", "PRIN", "CASS",
	"PERG", "CREDS", "COLL", "DEPP", "GENPRU", "MIPRU", "BIPRU", "IPRU",
	"FEES", "SUP", "MAR", "DTR",
}

// regulatoryRulebookCodes is regulatoryRulebookCodesBase as the exact regex
// alternation string fad209b92 shipped — kept as a derived value, not
// retyped, so there is no way for the fleet-wide default to drift from the
// per-site-capable version below. A `var` rather than the original `const`
// only because building it from the slice needs a function call; every
// existing use (string concatenation into a regexp) works identically
// either way.
var regulatoryRulebookCodes = "(?:" + strings.Join(regulatoryRulebookCodesBase, "|") + ")"

// citationCodePresets maps a named sector to the rulebook codes a site in
// that sector may cite. RFC_060 Q5 (owner-ruled 2026-09-03): "I will be
// extending to vet and legal quite soon so let's fix it with those in
// mind" — the second and third consumers this estate's own "don't design
// ahead of one" instinct had been withholding presets for. An unrecognised
// preset name contributes nothing silently (ParseEvidenceBase is a pure
// parse with no logger to warn through); KnownCitationCodePresets exists so
// a future validator can flag a typo without this file needing one.
var citationCodePresets = map[string][]string{
	"veterinary": {"RCVS", "VMD"},
	"legal":      {"SRA"},
	"medical":    {"GMC", "MHRA", "CQC"},
}

// KnownCitationCodePresets lists the preset names citation_code_presets may
// name. Exported so a discovery check or admin-door validator can flag an
// unrecognised preset without this file growing one itself.
func KnownCitationCodePresets() []string {
	names := make([]string, 0, len(citationCodePresets))
	for name := range citationCodePresets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// compileCitationCodeRegexes builds THIS site's own citation-recognition
// regexes: regulatoryRulebookCodesBase unioned with eb.CitationCodePresets
// (expanded) and eb.CitationCodes (literal, ad hoc) — additive only, never
// a replacement, so a site declaring nothing compiles to patterns BYTE
// IDENTICAL to the fleet default. RFC_060 §3f constraint 1: the matching
// RULE does not change — case-sensitive (no `(?i)`, load-bearing per the
// comment on the base patterns below), code immediately followed by a
// digit, two-letter codes excluded (silently dropped here, not stored,
// because even a site-declared two-letter code is "too collidable" per
// fad209b92's own measurement — FCA said the same about bare `FCA`/`PRA`).
func (eb *EvidenceBase) compileCitationCodeRegexes() {
	codes := append([]string{}, regulatoryRulebookCodesBase...)
	for _, preset := range eb.CitationCodePresets {
		codes = append(codes, citationCodePresets[preset]...)
	}
	for _, c := range eb.CitationCodes {
		if len(c) <= 2 {
			continue
		}
		codes = append(codes, c)
	}
	pattern := "(?:" + strings.Join(codes, "|") + ")"
	eb.citationPrefixRe = regexp.MustCompile(`\b` + pattern + `\s*$`)
	eb.citationContextRe = regexp.MustCompile(`\b` + pattern + `\s*\d`)
}

// citationContextPattern and citationPrefixPattern return this site's own
// compiled regexes, falling back to the fleet-wide default when eb was
// constructed directly (a struct literal, as every existing test in this
// package does) rather than through ParseEvidenceBase — the only place
// compileCitationCodeRegexes runs. The fallback is exactly the pattern a
// direct construction already got before this file existed, so no existing
// test's expectations change.
func (eb *EvidenceBase) citationContextPattern() *regexp.Regexp {
	if eb != nil && eb.citationContextRe != nil {
		return eb.citationContextRe
	}
	return regulatoryCitationContextRe
}

func (eb *EvidenceBase) citationPrefixPattern() *regexp.Regexp {
	if eb != nil && eb.citationPrefixRe != nil {
		return eb.citationPrefixRe
	}
	return rulebookCitationPrefixRe
}

// Shape 1 — the number IS the citation: "CONC 5A", "MCOB 4.1". The digits are
// part of the rule's name, exactly as "Tier 2" is handled by labelPrefixRe.
// The FLEET-WIDE default, used only as isExcludedNumber's fallback for a nil
// eb; every parsed EvidenceBase carries its own site-specific version on
// eb.citationPrefixRe (compileCitationCodeRegexes), which a site with no
// CitationCodes/CitationCodePresets compiles byte-identical to this.
var rulebookCitationPrefixRe = regexp.MustCompile(`\b` + regulatoryRulebookCodes + `\s*$`)

// Shape 2 — a REGULATORY FIGURE quoted WITH its rule: "0.8% per day under
// CONC 5A". The figure is the regulator's, not a claim about this business, and
// quoting it beside its rule is the behaviour these sites' own briefs REQUIRE
// ("every regulatory figure must be quoted together with the named rule it
// comes from"). So the site doing the right thing was what tripped the scan.
//
// ⚠ THE TEST IS FOR A CITATION, NOT FOR THE REGULATOR'S NAME. An earlier draft
// included bare `FCA`/`PRA`/`FOS`, which on a consumer-credit site appears in
// nearly every paragraph and would have switched the numeric scan off for the
// whole sector — the opposite of what this change is for. A code followed by a
// digit is a citation; "the FCA" is a subject.
//
// FLEET-WIDE DEFAULT ONLY — see rulebookCitationPrefixRe's comment above;
// ScanUnregisteredNumbers uses eb.citationContextRe instead.
var regulatoryCitationContextRe = regexp.MustCompile(`\b` + regulatoryRulebookCodes + `\s*\d`)

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
// `am`/`pm` are here for the same reason as `min read`: a CLOCK TIME is not a
// claim about the business, but it sits happily inside business prose and the
// lexical gate cannot tell them apart. Measured on ai-agent-orchestration.com,
// 2026-08-22: a generated pricing page was refused because "debug a failing
// agent chain at 2am" put a `2` next to the word "agent". The `\b` matters —
// without it "5 amazing" would be excluded too.
var unitSuffixRe = regexp.MustCompile(`(?i)^\s*(px|rem|em|vh|vw|ms|sec|seconds?|min(ute)?s?\s+read|kb|mb|gb|tb|fps|am\b|pm\b|st\b|nd\b|rd\b|th\b|[-–]\s*(hour|day|week|month|year|minute|second|token|character|step|person|page)\b)`)

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

	// ComponentFunction is content_components.function for the component this
	// text renders in — carried on the row as page_components.slot_name, and as
	// component_function in sections_metadata. EMPTY means unknown, which scans:
	// a caller holding only whole-page HTML, or site chrome belonging to no
	// component, must not silently inherit a component-grain exemption.
	//
	// It is framework-authored data, never anything parsed out of the rendered
	// HTML — see thirdPartyDataComponents for why that distinction is the whole
	// security of this field.
	ComponentFunction string
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
//
// THE TRACKER/DIRECTORY THREE WERE HERE FROM 2026-08-22..24 AND ARE NOW GONE
// (bugs_open/364, RFC_053). They were the interim: measured, real, and knowingly
// in breach of the second half of the bar above, because those pages carry a
// marketing hero and a call-to-action that a PAGE-grain gate silences along with
// the listing. Phase 2 moved the decision to the COMPONENT
// (thirdPartyDataComponents below), so the listing is skipped and the hero and
// CTA are scanned again — which is what this map's own 'report' note asks for
// when it says excluding a page type "would fix those by coincidence, not by
// mechanism". The interim's measurements are preserved on that map, since they
// are what justified the membership; do not re-add the page types here.

var editorialPageTypes = map[string]bool{
	"guide":      true,
	"blog-post":  true,
	"news-index": true,
	"tool":       true,
	"game":       true,
}

// thirdPartyDataComponents are the COMPONENT functions whose body renders records
// about OTHER organisations — an aggregated listing, a tracker, a directory —
// rather than anything about the site that publishes it. RFC_053 / bugs_open/364
// Phase 2.
//
// WHY THIS EXISTS AT COMPONENT GRAIN. The same page mixes both voices: a tracker
// page carries a marketing `hero` and a `call-to-action` in the site's own
// first-person voice, ABOVE AND BELOW a table of third parties' figures. Page
// type cannot express that, so gating there silenced the site's own claims to
// buy silence on the listing — which is what the interim did from 2026-08-22 to
// 2026-08-25, knowingly and in breach of editorialPageTypes' own membership bar.
//
// THE KEY IS THE COMPONENT'S REGISTERED FUNCTION, WHICH IS TRUSTED DATA. It comes
// from `content_components.function` (carried on the row as `page_components
// .slot_name`, and as `component_function` in sections_metadata) — the framework's
// own record of what it built. It is deliberately NOT read from the rendered HTML:
// that HTML is LLM-generated, so a marker embedded in it could be emitted by the
// very thing being policed. A declaration a writer can forge is not a control.
//
// THE BAR FOR MEMBERSHIP, same shape as editorialPageTypes' and same discipline:
// a MEASURED false positive on live copy, AND a body that renders third-party
// records BY CONSTRUCTION rather than by editorial choice. Do not add from
// intuition, and do not add a component merely because it is a list — measured
// 2026-08-24, `case-studies-list` on the same site is a list of the site's OWN
// work ("orchestrates 30+ specialised agents", "under 4 hours") and carries three
// genuine claims. A `-list`/`-listing` suffix is a naming convention, not a
// content guarantee — the same trap `editorialPageTypes` records for `-index`.
//
// Each member, with what earns it (measured 2026-08-24, cmd/claimscan against
// ai-agent-orchestration.com's own live register over live rendered_html, export
// asserted row-for-row against the DB — 20 findings across these three, precision
// ZERO): every finding is a third party's figure — "rollout_scope Over 80% of
// Fortune 500…", "200,000 onboarded users" (someone else's), "JSON-RPC 2.0" (a
// version string), and the digit `2` inside the acronym A2A. Two of the eleven
// tokens are not statistics at all, which is the sharpest evidence that a lexical
// gate was never answering the question it was asked.
// ⚠ VERIFY A NEW KEY AGAINST LIVE `slot_name`, NOT AGAINST
// `content_components.function` — they are not the same column and they do not
// always agree. Measured 2026-08-25: **106 of 2,033** live rows have
// `page_components.slot_name` different from their component's `function`
// (`prose-0` vs `ported-prose`, `call_to_action` vs `call-to-action`,
// `FAQ Section` vs `faq`). The surface is keyed on the SLOT, because that is what
// every call site actually holds. A key that matches neither simply never fires
// and the component is scanned — the safe direction, and also a silent no-op, so
// check rather than assume (council round 1, correlation 3ed2b792, edit-quality
// seat, medium — the objection was right):
//
//	SELECT pc.slot_name, count(*) FROM page_components pc
//	WHERE pc.slot_name = '<candidate>' AND pc.locked_at IS NULL GROUP BY 1;
//
// All three below were confirmed present as live `slot_name` values on
// 2026-08-25, and the corpus mutation check is the stronger proof that they fire:
// emptying this map returns exactly the original 36 findings with 20 on tracker
// pages, so the exemption is doing the work and is not a no-op.
var thirdPartyDataComponents = map[string]bool{
	"adoption-tracker-listing": true,
	"protocol-tracker-listing": true,
	"model-directory-listing":  true,
}

// ProseNumbersAreClaims reports whether the heuristic number scan applies to
// prose on this surface. It governs ScanUnregisteredNumbers ONLY — and that
// claim is a grep, not an assertion (council round 1, correlation 3ed2b792,
// prior-art seat: "asserted, not shown checked"). Re-run it before trusting it,
// because a sixth construction point or a second reader would silently leave a
// surface un-widened:
//
//	grep -rn 'editorialPageTypes\|thirdPartyDataComponents\|ProseNumbersAreClaims' --include=*.go .
//	grep -rn 'ClaimSurface{' --include=*.go . | grep -v _test
//
// On 2026-08-25 that returned: both maps read ONLY by ProseNumbersAreClaims,
// which has ONE caller (the guard below); and 5 production ClaimSurface literals
// (validate_page_content, check_unverified_claims ×2, save_sections_claims_guard,
// cmd/claimscan), all of which pass ComponentFunction.
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
	// Component grain FIRST: it is the more specific signal, and on a mixed page
	// it is the only one that can be right. An UNRECOGNISED or ABSENT component
	// function falls through to the page-type question rather than deciding
	// anything — so a new template, site chrome, or a caller with no component in
	// hand is scanned exactly as before. That direction is the same one the zero
	// value takes and for the same reason: a scanner that has gone quiet and one
	// that is broken look identical from outside.
	if thirdPartyDataComponents[normaliseSurfaceKey(s.ComponentFunction)] {
		return false
	}
	return !editorialPageTypes[normaliseSurfaceKey(s.PageType)]
}

// normaliseSurfaceKey is the single normalisation both lookups use, so the two
// maps cannot drift on case or padding. Pinned by
// TestPageTypeMatchingIsCaseAndSpaceInsensitive.
func normaliseSurfaceKey(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
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
// HasScannableRegister reports whether this base carries CONTENT the
// deterministic register-comparison scans can work against — facts or banned
// claims. A base that parses non-nil only because it carries an attestation
// (regulated, operating_history) is NOT scannable: arming the unregistered-
// number scan against an empty fact list reports every number on the site at
// error severity and refuses the build (the bugs_open/364 class). The Regulated
// widening carried that latent hazard; measured 2026-08-24, zero live
// attestation-only registers, so this changes no live site (bugs_open/380).
func (eb *EvidenceBase) HasScannableRegister() bool {
	return eb != nil && (len(eb.Facts) > 0 || len(eb.BannedClaims) > 0)
}

// Callers gate this on HasScannableRegister() (the gate, the discovery check
// and claimscan do); the function itself keeps its nil-only contract so a test
// or tool may still scan against a deliberately empty register.
func (eb *EvidenceBase) ScanUnregisteredNumbers(blocks []string, surface ClaimSurface) []ClaimFinding {
	if eb == nil || !surface.ProseNumbersAreClaims() {
		return nil
	}
	found := make(map[string]*ClaimFinding)
	var order []string

	for _, block := range blocks {
		for _, loc := range numberCandidateRe.FindAllStringIndex(block, -1) {
			token := block[loc[0]:loc[1]]
			if isExcludedNumber(block, loc[0], loc[1], eb.citationPrefixPattern()) {
				continue
			}
			window := claimWindow(block, loc[0], loc[1])
			if !businessClaimContextRe.MatchString(window) {
				continue
			}
			if phoneContextRe.MatchString(window) {
				continue
			}
			// A figure quoted beside its rule is the REGULATOR'S number, not a
			// claim about this business — and quoting it that way is what the
			// finance sites' own briefs require of them. eb's OWN pattern (RFC_060
			// Q5): a site's declared CitationCodes/CitationCodePresets union onto
			// the fleet default here, not just at isExcludedNumber's prefix check.
			if eb.citationContextPattern().MatchString(window) {
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
// prefixRe is the rulebook-citation-prefix pattern to test against — the
// caller's eb.citationPrefixPattern() (RFC_060 Q5: a site's own declared
// codes union onto the fleet default) or the package-level default directly
// when no EvidenceBase is in scope (claims_stats.go's fleet-wide caller).
func isExcludedNumber(block string, start, end int, prefixRe *regexp.Regexp) bool {
	token := block[start:end]

	// A digit with a LETTER on BOTH sides is inside an identifier, not a
	// quantity: A2A (Agent-to-Agent), W3C, B2B, H2O. bugs_open/364 Phase 2
	// measured this as a live build-refusal: restoring the scan to a tracker
	// page's hero (correct in itself) made "MCP, A2A and half a dozen other
	// proposals" raise `unregistered_number "2"` at ERROR severity, which
	// refuses the page — over a digit inside an acronym.
	//
	// Both sides must be letters, which is what keeps it bounded. A real
	// quantity always has a word boundary in front of it ("we serve 45,000
	// clients", "over 1,600 orchestrations"); nothing this platform publishes
	// writes one flanked by letters. One-sided forms stay scanned on purpose —
	// "3D" and "MP3" are not excluded by this rule, only by whatever else
	// applies to them.
	if start > 0 && end < len(block) {
		before, after := block[start-1], block[end]
		if isASCIILetter(before) && isASCIILetter(after) {
			return true
		}
	}

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

	// Rulebook citation: the digits of "CONC 5A" are the rule's name.
	if prefixRe.MatchString(block[:start]) {
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

// isASCIILetter reports whether b is an unaccented ASCII letter. Deliberately
// ASCII-only: this guards an IDENTIFIER shape (A2A, W3C), and identifiers on
// this estate are ASCII. A multibyte letter falls through and the number stays
// scanned, which is the noisy direction.
func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
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
		// The fact's CURRENT value did not support this number. A value the
		// register PREVIOUSLY held does, if this fact is armed: a page rendered
		// while that reading was current published a true figure, and the
		// register moving underneath it afterwards is not the page inventing
		// anything (bugs_open/386). Deliberately INSIDE the context-term gate
		// above and exact-only — a former reading is one specific number this
		// fact actually held, never a range, so unlike a tolerance it cannot
		// vouch for a number the register never carried.
		if f.historySupports(val) {
			return true
		}
	}
	return false
}

// historySupports reports whether one of this fact's superseded readings is
// exactly the published number.
//
// Ordering is not relied upon for correctness — every retained entry is equally
// valid evidence of what the register once held — but the writer appends, so the
// newest are last, and that is the half the cap keeps.
func (f *EvidenceFact) historySupports(val float64) bool {
	if f == nil || !f.RetainHistory || len(f.History) == 0 {
		return false
	}
	// The cap is enforced at the READER as well as at the writer that trims.
	// It bounds what the scan will ACCEPT, so a fact whose stored history has
	// grown past it — by a hand seed, a migration backfill, or any writer that
	// forgot to trim — must not be allowed to accept more than an armed fact is
	// permitted to. A guard that lives only in the writer is not a bound.
	h := f.History
	if len(h) > FactHistoryMaxEntries {
		h = h[len(h)-FactHistoryMaxEntries:]
	}
	for i := range h {
		if math.Abs(val-h[i].Value) < 1e-6 {
			return true
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
