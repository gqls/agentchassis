// FILE: platform/orchestration/actions/discovery_checks/static_attribute_checks.go
//
// Tier 2 attribute assertion: `attribute_absent` and `attribute_matches`.
//
// WHY THIS EXISTS. The experience register's single largest harness gap is the
// inability to say "this element does / does not carry this attribute". Measured
// 2026-07-28 across the nine harvested entries: 13 of the 38 deferred clauses
// are blocked on it, and there is no entry it does not affect. It is the
// `no-inert-control` invariant — *a control that cannot do anything must not be
// presented as one* — which arrived in six independent implementations across
// two sites and five authors, and which is the reason the register exists at
// all. Until now the register could record that rule and never check it.
//
// THE SHAPE IS THE ENTRIES', NOT MINE. Eight checks were authored across six
// register entries before any of this code existed, and they had already agreed
// on a shape:
//
//	{"type":"attribute_absent",  "selector":"…", "attributes":["href","tabindex"]}
//	{"type":"attribute_matches", "selector":"…", "attribute":"href",
//	                             "not_matches":"^(#|\\s*)$"}
//
// That is what is implemented. This matters because of a landmine this
// workstream has already stepped on once: a shape invented alongside its own
// tests passes its own tests and disagrees with every real caller. The real
// callers were written first here, so they are the specification.
//
// TWO RULES, BOTH LOAD-BEARING.
//
//  1. ZERO MATCHING ELEMENTS IS A SKIP — never a pass. "No element carries a
//     dead href" is vacuously true of a page with no such elements, and for a
//     list whose rows are cloned client-side that is the NORMAL Tier 2 state.
//     A vacuous pass is the precise failure this register exists to end, so the
//     check declines instead: skipped is reported, and a skip can never satisfy
//     `experienceVerdict`, which requires at least one PASS and zero FAILS.
//     It is also not a FAIL, because failing there would call a correct page
//     broken — the defect found on 2026-07-28c, when a check we had already
//     recorded as impossible was executed anyway and produced the register's
//     first verdict: "broken", about a page that was fine.
//
//  2. REFUTATION IS CONFINED TO ELEMENTS ACTUALLY IN THE SERVED HTML. Tier 2's
//     governing rule is "static checks CONFIRM, never refute" — a runtime-built
//     path passes on its anchor. An attribute check does refute, but only about
//     elements it can see; anything JS builds is invisible to it and therefore
//     skipped rather than judged. The precedent is this file's neighbour: the
//     built-in `shell-dead-controls` sweep already fails a page for a no-op
//     href, with the runtime-fill shells exempted.
//
// A REAL DOM, AND WHY THAT IS NOT A SECOND SELECTOR SEMANTICS. The rest of the
// Tier 2 evaluator uses the anchor rule — the leftmost id/class/tag token,
// matched against raw HTML — because presence is all it needs. An attribute
// assertion needs element IDENTITY: which elements matched, and what each of
// them carries. That cannot be done on a string, so these two checks parse the
// page with goquery. The two semantics coexist safely because the DOM selector
// is strictly STRICTER than the anchor rule: anything it matches, the anchor
// rule would also have matched. It can therefore never pass something the
// anchor rule would fail, which is the direction that would matter.
package discovery_checks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
)

// attributeVerdict is one of the three outcomes the caller projects onto the
// evaluation's passed/failed/skipped lists.
type attributeVerdict int

const (
	attrPass attributeVerdict = iota
	attrFail
	attrSkip
)

type attributeOutcome struct {
	verdict attributeVerdict
	detail  string
}

func attrSkipf(format string, args ...interface{}) attributeOutcome {
	return attributeOutcome{attrSkip, fmt.Sprintf(format, args...)}
}

func attrFailf(format string, args ...interface{}) attributeOutcome {
	return attributeOutcome{attrFail, fmt.Sprintf(format, args...)}
}

func attrPassf(format string, args ...interface{}) attributeOutcome {
	return attributeOutcome{attrPass, fmt.Sprintf(format, args...)}
}

// evaluateAttributeCheck applies one attribute check to the parsed page.
//
// Every detail string carries the ELEMENT COUNT. A count is the only thing that
// distinguishes a real assertion from a vacuous one after the fact, and a
// reader of a stored result has nothing else to go on.
func evaluateAttributeCheck(doc *goquery.Document, ch criteriaCheck) attributeOutcome {
	if strings.TrimSpace(ch.Selector) == "" {
		return attrSkipf("%s with no selector — nothing to assert on", ch.Type)
	}

	// Compile the selector explicitly rather than letting goquery.Find swallow
	// the error: Find returns an empty selection for a MALFORMED selector, which
	// is indistinguishable from a selector that legitimately matched nothing.
	// Those two need different answers — one is a criteria defect, the other is
	// the normal client-side-rendering case — so they must not collapse.
	sel, err := cascadia.Compile(ch.Selector)
	if err != nil {
		return attrSkipf("criteria defect, not a page defect: selector %q does not compile (%v) — nothing was asserted", ch.Selector, err)
	}

	nodes := doc.FindMatcher(sel)
	count := nodes.Length()
	if count == 0 {
		return attrSkipf("selector %q matched no elements in the served HTML — nothing was asserted (elements built client-side are Tier 4's claim, not a pass here)", ch.Selector)
	}

	switch ch.Type {
	case "attribute_absent":
		return evaluateAttributeAbsent(nodes, ch, count)
	case "attribute_matches":
		return evaluateAttributeMatches(nodes, ch, count)
	default:
		return attrSkipf("%s is not an attribute check", ch.Type)
	}
}

// evaluateAttributeAbsent fails if ANY matched element carries ANY of the named
// attributes. It reports the first offender with its value, because "which one
// and what does it say" is the whole content of the finding.
func evaluateAttributeAbsent(nodes *goquery.Selection, ch criteriaCheck, count int) attributeOutcome {
	wanted := trimmedNonEmpty(ch.Attributes)
	if len(wanted) == 0 {
		return attrSkipf("attribute_absent with no attributes listed — the check asserts nothing")
	}

	var offences []string
	nodes.Each(func(i int, s *goquery.Selection) {
		for _, attr := range wanted {
			if v, ok := s.Attr(attr); ok {
				offences = append(offences, fmt.Sprintf("element %d carries %s=%q", i+1, attr, v))
			}
		}
	})
	if len(offences) > 0 {
		return attrFailf("%d of %d element(s) matching %q carry a forbidden attribute (%s): %s",
			len(offences), count, ch.Selector, strings.Join(wanted, ", "), strings.Join(offences, "; "))
	}
	return attrPassf("none of %d element(s) matching %q carry %s",
		count, ch.Selector, strings.Join(wanted, ", "))
}

// evaluateAttributeMatches asserts one attribute's VALUE across every matched
// element, against `matches` (must match) and/or `not_matches` (must not).
//
// A MISSING ATTRIBUTE IS A FAILURE, for both forms. The conditionality these
// entries actually want — "the link, if there is one, is real" — is carried by
// the SELECTOR: a component that renders no link matches no elements and the
// check skips at the guard above. Once an element has matched, the contract
// says it carries this attribute, and an absent one is the strongest possible
// way of not matching.
func evaluateAttributeMatches(nodes *goquery.Selection, ch criteriaCheck, count int) attributeOutcome {
	attr := strings.TrimSpace(ch.Attribute)
	if attr == "" {
		return attrSkipf("attribute_matches with no attribute named — the check asserts nothing")
	}
	if strings.TrimSpace(ch.Matches) == "" && strings.TrimSpace(ch.NotMatches) == "" {
		return attrSkipf("attribute_matches on %q with neither matches nor not_matches — the check asserts nothing", attr)
	}

	var must, mustNot *regexp.Regexp
	var err error
	if p := strings.TrimSpace(ch.Matches); p != "" {
		if must, err = regexp.Compile(p); err != nil {
			return attrSkipf("criteria defect, not a page defect: matches %q does not compile (%v) — nothing was asserted", p, err)
		}
	}
	if p := strings.TrimSpace(ch.NotMatches); p != "" {
		if mustNot, err = regexp.Compile(p); err != nil {
			return attrSkipf("criteria defect, not a page defect: not_matches %q does not compile (%v) — nothing was asserted", p, err)
		}
	}

	var offences []string
	nodes.Each(func(i int, s *goquery.Selection) {
		v, ok := s.Attr(attr)
		if !ok {
			offences = append(offences, fmt.Sprintf("element %d has no %s attribute", i+1, attr))
			return
		}
		if must != nil && !must.MatchString(v) {
			offences = append(offences, fmt.Sprintf("element %d has %s=%q which does not match %s", i+1, attr, v, ch.Matches))
		}
		if mustNot != nil && mustNot.MatchString(v) {
			offences = append(offences, fmt.Sprintf("element %d has %s=%q which matches the forbidden %s", i+1, attr, v, ch.NotMatches))
		}
	})
	if len(offences) > 0 {
		return attrFailf("%d of %d element(s) matching %q fail the %s assertion: %s",
			len(offences), count, ch.Selector, attr, strings.Join(offences, "; "))
	}
	return attrPassf("all %d element(s) matching %q carry an acceptable %s (%s)",
		count, ch.Selector, attr, describeAttributeAssertion(ch))
}

func describeAttributeAssertion(ch criteriaCheck) string {
	var parts []string
	if p := strings.TrimSpace(ch.Matches); p != "" {
		parts = append(parts, "matches "+p)
	}
	if p := strings.TrimSpace(ch.NotMatches); p != "" {
		parts = append(parts, "not_matches "+p)
	}
	return strings.Join(parts, " and ")
}

func trimmedNonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}
