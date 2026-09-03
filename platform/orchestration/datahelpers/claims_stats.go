package datahelpers

// Stat-field claim extraction — bugs_open/043 fix candidates 2 and 4.
//
// WHY THIS EXISTS RATHER THAN REUSING THE HTML SCAN.
//
// The claims engine in claims.go scans rendered HTML, and it is structurally
// blind to exactly the shape bug 043 is about. Two independent reasons, both
// measured:
//
//  1. extractAssertions (claims.go:164) flushes a block boundary on entering
//     and leaving every element in assertionBlockElements (claims.go:147),
//     and `div` is in that map. A stat card renders as
//
//         <div class="stat-value">170<span class="stat-suffix"></span></div>
//         <div class="stat-label">Agent Definitions</div>
//
//     so the number and its label land in SEPARATE blocks. claimWindow for the
//     number is the bare string "170", businessClaimContextRe cannot match it,
//     and the candidate is dropped before numberSupported is ever consulted.
//     TestStatCardHTMLSplitsValueFromLabel pins this as a property of the
//     extractor, so a future widening of assertionBlockElements is a deliberate
//     act rather than an accident.
//
//  2. businessClaimContextRe is a PROSE lexical gate, and the right one there.
//     It is the wrong gate here: it does not match "Gripper Models Indexed",
//     "Actuation Technologies", "Takes Filed Today" or "PRD Accuracy Gap" —
//     roughly half the labels bug 043 documents. A field named stat1_value in
//     a stat component is a published quantitative claim BY CONSTRUCTION. Its
//     structural position in the schema replaces the lexical gate; it must not
//     be filtered by it. So ScanStatClaims deliberately does NOT apply it.
//
// And candidate 4 (the placeholder-unit lint) is only implementable here at
// all: rendered HTML gives you the single string "2,400+%", from which nothing
// can tell whether the "%" was authored or leaked out of the component
// schema's static `fallback`. In content_data they are two fields, and the
// whole of 043's update (b) is about the second one.
//
// Severity is deliberately NOT decided in this package — the action layer
// assigns it, exactly as it already does for ScanBannedClaims and
// ScanUnregisteredNumbers.

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// StatPairing records how confidently a value field was tied to its label.
type StatPairing string

const (
	// StatPairedExact: the label key is the value key's stem plus a known
	// label-role token (stat1_value -> stat1_label).
	StatPairedExact StatPairing = "exact"
	// StatPairedAnchor: no exact sibling, but exactly one candidate string
	// shares the value's index anchor (row2_spark_value -> row2_feature).
	StatPairedAnchor StatPairing = "anchor"
	// StatUnpaired: zero or several candidates. We never guess — an unpaired
	// claim is a DIFFERENT finding, not a weaker version of the same one.
	StatUnpaired StatPairing = "unpaired"
)

// StatClaim is one published figure lifted out of a component's content_data,
// reassembled with the sibling fields that give it meaning.
type StatClaim struct {
	Component string      `json:"component"`
	ValueKey  string      `json:"value_key"`
	Value     string      `json:"value"`
	LabelKey  string      `json:"label_key,omitempty"`
	Label     string      `json:"label,omitempty"`
	SuffixKey string      `json:"suffix_key,omitempty"`
	Suffix    string      `json:"suffix,omitempty"`
	Detail    string      `json:"detail,omitempty"`
	Pairing   StatPairing `json:"pairing"`
}

// Rendered returns what a visitor actually sees: the value immediately
// followed by its suffix, which is how every stat template emits the pair.
func (c StatClaim) Rendered() string { return c.Value + c.Suffix }

// Window is the synthetic claim window an HTML stat card can never produce,
// because its parts are separate block elements. It is what gets handed to
// numberSupported, so a fact's ContextTerms scoping keeps working unchanged.
func (c StatClaim) Window() string {
	parts := make([]string, 0, 3)
	if c.Label != "" {
		parts = append(parts, c.Label)
	}
	parts = append(parts, c.Rendered())
	if c.Detail != "" {
		parts = append(parts, c.Detail)
	}
	return strings.Join(parts, " ")
}

// Role tokens. A key's final underscore-separated token decides its role.
var (
	statValueTokens  = map[string]bool{"value": true, "number": true, "figure": true}
	statLabelTokens  = []string{"label", "name", "title", "caption", "heading", "metric", "key"}
	statSuffixTokens = []string{"suffix", "unit", "units"}
	statDetailTokens = []string{"description", "desc", "detail", "details", "note", "subtext"}
)

// statRoleToken reports whether a token names any known role — used to keep
// the anchor fallback from pairing a value with another value.
func statRoleToken(tok string) bool {
	if statValueTokens[tok] {
		return true
	}
	for _, group := range [][]string{statLabelTokens, statSuffixTokens, statDetailTokens} {
		for _, t := range group {
			if t == tok {
				return true
			}
		}
	}
	return false
}

// numberWords let an anchor be derived from stat_one_value as well as stat1_value.
var numberWords = map[string]bool{
	"one": true, "two": true, "three": true, "four": true, "five": true, "six": true,
	"seven": true, "eight": true, "nine": true, "ten": true, "eleven": true, "twelve": true,
}

var digitRunRe = regexp.MustCompile(`[0-9]+$`)

// statValueStem returns the key minus its trailing value-role token, and
// whether the key is a value field at all.
func statValueStem(key string) (string, bool) {
	k := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
	i := strings.LastIndex(k, "_")
	if i < 0 || i == len(k)-1 {
		return "", false
	}
	if !statValueTokens[k[i+1:]] {
		return "", false
	}
	return k[:i], true
}

// statAnchor returns the shortest prefix of a stem ending in an index — a
// digit run or a number word. "row2_spark" -> "row2"; "card3_stat" -> "card3";
// "stat_one" -> "stat_one". Empty when the stem carries no index.
func statAnchor(stem string) string {
	parts := strings.Split(stem, "_")
	for i, p := range parts {
		if numberWords[p] || digitRunRe.MatchString(p) {
			return strings.Join(parts[:i+1], "_")
		}
	}
	return ""
}

// contentString normalises a content_data value to a trimmed string. Numbers
// and booleans are stringified — a writer that returns 170 rather than "170"
// is still publishing a figure.
func contentString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case bool:
		return ""
	case []interface{}, map[string]interface{}:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

var statBlankSentinels = map[string]bool{
	"": true, "—": true, "–": true, "-": true, "n/a": true, "na": true, "tbc": true, "tbd": true,
}

// statDisplayOrdinalRe matches a bare zero-padded index: "01", "02", "03".
//
// A leading zero is never how a quantity is written — nothing counts "01
// customers" — so in a stat field this is a step or card marker for display.
// FOUND BY MEASUREMENT 2026-07-26 while building bugs_open/093's second call
// site: vonc.com's about page publishes 01/02/03 as step numbers in a process
// component, and each was extracted as a published figure and reported as
// unsupported. On a site with registered facts that is `error` severity in the
// build gate — a page made unbuildable by its own step numbering.
//
// Deliberately narrow, and deliberately HERE rather than in claims.go's shared
// isExcludedNumber: this is a property of a bare stat FIELD VALUE, whereas the
// exclusions in claims.go reason about a number's position inside a prose
// block. Widening those to cover this would change what the prose scan sees on
// every site, for a shape only the stat path produces. Note "0" alone is not
// matched — zero is a real count, and an honest one (bugs_closed/073).
var statDisplayOrdinalRe = regexp.MustCompile(`^0\d+$`)

// hasDigit reports whether a value could carry a numeric claim at all.
func hasDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

// ExtractStatClaims lifts every value/label/suffix/detail group out of ONE
// component's content_data. Pure: no DB, no evidence base, no severity.
// Values that carry no digits at all (a qualitative stat like "Vendor-Neutral")
// and blank sentinels are dropped — they are not quantitative claims.
func ExtractStatClaims(component string, contentData map[string]interface{}) []StatClaim {
	if len(contentData) == 0 {
		return nil
	}

	// Normalised view of the component's own keys, for sibling lookup.
	norm := make(map[string]string, len(contentData)) // normalised key -> raw key
	vals := make(map[string]string, len(contentData)) // normalised key -> string value
	for k, v := range contentData {
		nk := strings.ToLower(strings.ReplaceAll(k, "-", "_"))
		norm[nk] = k
		vals[nk] = contentString(v)
	}

	// Deterministic order — findings must not reshuffle between runs.
	keys := make([]string, 0, len(norm))
	for nk := range norm {
		keys = append(keys, nk)
	}
	sort.Strings(keys)

	var claims []StatClaim
	for _, nk := range keys {
		stem, ok := statValueStem(nk)
		if !ok {
			continue
		}
		value := vals[nk]
		if statBlankSentinels[strings.ToLower(value)] || !hasDigit(value) {
			continue
		}
		if statDisplayOrdinalRe.MatchString(value) {
			continue // "01"/"02"/"03" — a display index, not a quantity
		}

		c := StatClaim{Component: component, ValueKey: norm[nk], Value: value, Pairing: StatUnpaired}

		// Tier 1 — exact stem + a label-role token.
		for _, t := range statLabelTokens {
			if lv, ok := vals[stem+"_"+t]; ok && lv != "" {
				c.LabelKey, c.Label, c.Pairing = norm[stem+"_"+t], lv, StatPairedExact
				break
			}
		}

		// Tier 2 — anchor fallback, gated on uniqueness. Resolves the shapes
		// where the label token is domain-specific (row2_spark_value paired to
		// row2_feature only because exactly one candidate sits under row2).
		if c.Pairing == StatUnpaired {
			if anchor := statAnchor(stem); anchor != "" {
				var cands []string
				for _, ok2 := range keys {
					if ok2 == nk || !strings.HasPrefix(ok2, anchor+"_") {
						continue
					}
					tok := ok2[strings.LastIndex(ok2, "_")+1:]
					if statRoleToken(tok) {
						continue
					}
					cv := vals[ok2]
					if cv == "" || hasDigit(cv) {
						continue // a number is not a label
					}
					cands = append(cands, ok2)
				}
				// Tier 3 is the else: zero or several candidates means refuse.
				if len(cands) == 1 {
					c.LabelKey, c.Label, c.Pairing = norm[cands[0]], vals[cands[0]], StatPairedAnchor
				}
			}
		}

		// Siblings — tier 1 only. A wrongly-paired unit is worse than none.
		for _, t := range statSuffixTokens {
			if sv, ok := vals[stem+"_"+t]; ok && sv != "" {
				c.SuffixKey, c.Suffix = norm[stem+"_"+t], sv
				break
			}
		}
		for _, t := range statDetailTokens {
			if dv, ok := vals[stem+"_"+t]; ok && dv != "" {
				c.Detail = dv
				break
			}
		}

		claims = append(claims, c)
	}
	return claims
}

// ScanStatClaims flags published figures that no registered fact supports.
//
// Deliberately does NOT apply businessClaimContextRe — see the file header.
// The Check value distinguishes a confidently-paired claim from one whose
// label could not be determined, so the caller can grade them differently.
func (eb *EvidenceBase) ScanStatClaims(claims []StatClaim) []ClaimFinding {
	if eb == nil {
		return nil
	}
	var findings []ClaimFinding
	for _, c := range claims {
		loc := numberCandidateRe.FindStringIndex(c.Value)
		if loc == nil {
			continue
		}
		// Reuse the structural exclusions: years, currency, versions, dates,
		// ratios, measurements. "£29" and "2019" are not business counts.
		// eb.citationPrefixPattern(): RFC_060 Q5, this site's own declared
		// citation codes union onto the fleet default.
		if isExcludedNumber(c.Value, loc[0], loc[1], eb.citationPrefixPattern()) {
			continue
		}
		val, ok := parseClaimNumber(c.Value[loc[0]:loc[1]], c.Value[loc[1]:])
		if !ok {
			continue
		}
		window := c.Window()
		if eb.numberSupported(val, window) {
			continue
		}
		check := "unregistered_stat"
		reason := "a figure published in a stat field matches no evidence_base fact value"
		if c.Pairing == StatUnpaired {
			check = "unregistered_stat_unpaired"
			reason = "a figure is published in a stat field that no registered fact supports, and its label could not be determined from sibling fields — check it by hand"
		}
		findings = append(findings, ClaimFinding{
			Check:       check,
			Matched:     c.Rendered(),
			Pattern:     c.Component + "." + c.ValueKey,
			Reason:      reason,
			Snippet:     window,
			Occurrences: 1,
		})
	}
	return findings
}

// ── Candidate 4: the placeholder-unit lint ─────────────────────────────────
//
// bugs_open/043 update (b): the system-stats schema declared its four suffixes
// as source=static with fallbacks "%", "ms", "+", "x", the render resolved an
// empty suffix to the fallback and PERSISTED it, and four sites shipped
// "2,400+%", "140+ms", "1,000sms" and "14,203%" live. 043 calls this "the only
// candidate that would have caught this specific instance at deploy time".

// Magnitude markers that may legitimately end a value: 2,400+ / 14k / 2.4M /
// 1,000s. These are never dimensions.
var statMagnitudeTailRe = regexp.MustCompile(`(?i)[0-9]\s*(\+|k|m|b|s)$`)

// Dimensional suffix classes. Bare "m" is deliberately absent: as a suffix it
// reads as the magnitude "million" far more often than "minutes", and a lint
// that is wrong about that would be worse than no lint.
var statDimensionalSuffixes = map[string]string{
	"%":  "rate",
	"ms": "duration", "s": "duration", "sec": "duration", "secs": "duration",
	"h": "duration", "hr": "duration", "hrs": "duration",
	"min": "duration", "mins": "duration",
	"x": "multiplier", "×": "multiplier",
}

// A dimensional suffix is licensed by a word in the label or detail that makes
// that dimension meaningful.
var statUnitLicence = map[string]*regexp.Regexp{
	"rate":       regexp.MustCompile(`(?i)\b(percent|percentage|rate|share|accuracy|precision|uptime|availability|satisfaction|margin|growth|conversion|completion|coverage|utilisation|utilization|reduction|increase|improvement|discount|gap)\b`),
	"duration":   regexp.MustCompile(`(?i)\b(time|latency|response|speed|duration|delay|wait|load|setup|build|render|turnaround|per\s+(request|query|call|page))\b`),
	"multiplier": regexp.MustCompile(`(?i)\b(faster|slower|more|fewer|times|multiple|improvement|gain|reduction|speedup|throughput|roi|return)\b`),
}

// LintStatUnits reports unit suffixes that cannot be true of their own value or
// label, and stat blocks that were only half cleaned up.
//
// No evidence base: it compares a component against ITSELF, so it is safe to
// run on any site whether or not a register exists. That is the whole reason
// it is a free function rather than a method on *EvidenceBase.
func LintStatUnits(claims []StatClaim) []ClaimFinding {
	var findings []ClaimFinding

	for _, c := range claims {
		suffix := strings.TrimSpace(c.Suffix)
		if suffix == "" {
			continue
		}
		class, dimensional := statDimensionalSuffixes[strings.ToLower(suffix)]
		if !dimensional {
			continue // "+", "k", "M" — magnitude markers, never flagged
		}

		// (a) A magnitude marker immediately followed by a dimension cannot be
		//     a real unit under any reading. This is 043's original instance.
		if statMagnitudeTailRe.MatchString(c.Value) {
			findings = append(findings, ClaimFinding{
				Check:       "stat_unit_impossible",
				Matched:     c.Rendered(),
				Pattern:     c.Component + "." + c.ValueKey,
				Reason:      "the value ends in a magnitude marker and is then given a dimensional unit, so the rendered figure (" + c.Rendered() + ") cannot mean anything — almost certainly an unedited schema fallback",
				Snippet:     c.Window(),
				Occurrences: 1,
			})
			continue
		}

		// (b) A dimension with nothing in the label licensing it. Warning, not
		//     error: a site owner may genuinely intend "36.6" + "%".
		if lic, ok := statUnitLicence[class]; ok && !lic.MatchString(c.Label+" "+c.Detail) {
			findings = append(findings, ClaimFinding{
				Check:       "stat_unit_mismatch",
				Matched:     c.Rendered(),
				Pattern:     c.Component + "." + c.ValueKey,
				Reason:      "a " + class + " unit is appended to a figure whose label does not describe a " + class + " — check the unit is intended and not a leftover schema fallback",
				Snippet:     c.Window(),
				Occurrences: 1,
			})
		}
	}

	// NOT IMPLEMENTED HERE: 043's point (c), "a partial cleanup is worse than
	// none" — a block where the unsourceable stats were blanked to an em dash
	// and the rest left in place reads as CHECKED while carrying a surviving
	// invention. It cannot be detected from []StatClaim, because
	// ExtractStatClaims drops blank sentinels by design (they are not
	// quantitative claims). Detecting it needs the raw content_data, i.e. its
	// own function with its own signature. Left out deliberately rather than
	// shipped as a check that can never fire; recorded in bugs_open/043.

	return findings
}
