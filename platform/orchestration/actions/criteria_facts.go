// FILE: platform/orchestration/actions/criteria_facts.go
//
// The fence-level `facts` declaration — Piece 2 of
// docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/
// PLAN_2026-08-09_facts_into_tool_acceptance.md (bugs_closed/225's class fix):
// a tool's criteria document names WHICH evidence-register facts the tool
// encodes, so the daily evidence sweep can tell the tool when one of them
// moves (Piece 3, refresh_evidence_fact_drift.go).
//
//	```criteria
//	{ "profiles": [...], "no_auto_fix": true,
//	  "facts": ["sdlt-ftb-relief-cap", "sdlt-additional-surcharge-floor"],
//	  "checks": [...] }
//	```
//
// Three properties, each deliberate:
//
//   - FENCE-LEVEL, not per-check. A tool encodes a fact; a check does not.
//     Precedent: `no_auto_fix` (tool_acceptance_actions.go, TL-040). A
//     per-check `facts` stays REFUSED by the validator's P7 inert-field rule,
//     because neither checker reads it there.
//   - IDS ONLY, never values. PBP-037's rule applied to tools: the declaration
//     pins WHICH facts a tool encodes; the values are always resolved from the
//     CURRENT register of the site whose page is being driven (doc_plans has no
//     site_id — landmine §5.1 of the plan). Pinning a value here would recreate
//     F3 (a golden that defends a wrong number) one rung higher.
//   - READ BY ONE CONSUMER. Tier 2 (discovery_checks/check_tool_acceptance.go)
//     and Tier 4 (internal/adapters/browserrunner/run_checks_action.go) ignore
//     unknown fence keys, so `facts` asserts NOTHING at acceptance time; only
//     refresh_evidence_base reads it. Do not expect a green fence to mean the
//     figures were compared — that is Piece 4 (an RFC), not this key.
//
// parseCriteriaFacts is FAIL-OPEN on absent/unparseable criteria, exactly like
// parseNoAutoFix beside it, and for the same reason: a fence that never said
// anything about facts must not become an error somewhere else. A fence that
// DID say something, malformed, is reported as issues so a human can see the
// declaration was ignored, not silently dropped.

package actions

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// criteriaFactsDoc is the fan-out's view of the fence — the one key it reads.
// Deliberately NOT folded into acceptanceFenceFlags: that struct is the JUDGE's
// view, and a malformed `facts` must not make parseNoAutoFix fail-open on a
// fence that set no_auto_fix correctly.
type criteriaFactsDoc struct {
	Facts json.RawMessage `json:"facts"`
}

// criteriaFactsFromValue validates the decoded `facts` value: a JSON array of
// non-empty strings, trimmed, de-duplicated (order preserved). Anything else is
// an issue naming what was wrong. Shared by the P11 validator rule and the
// fan-out parser so the two cannot disagree on what a well-formed declaration is.
func criteriaFactsFromValue(raw interface{}) (ids []string, issues []string) {
	if raw == nil {
		return nil, nil
	}
	list, ok := raw.([]interface{})
	if !ok {
		return nil, []string{"facts must be a JSON array of fact ids (strings)"}
	}
	seen := map[string]bool{}
	for i, item := range list {
		s, ok := item.(string)
		if !ok {
			issues = append(issues, fmt.Sprintf("facts[%d] is not a string", i))
			continue
		}
		s = strings.TrimSpace(s)
		if s == "" {
			issues = append(issues, fmt.Sprintf("facts[%d] is empty", i))
			continue
		}
		if seen[s] {
			issues = append(issues, fmt.Sprintf("facts[%d] %q is declared twice", i, s))
			continue
		}
		seen[s] = true
		ids = append(ids, s)
	}
	return ids, issues
}

// factsKeyRe matches `facts` in JSON KEY POSITION — the quoted name followed by
// a colon. It cannot be a map lookup, because the whole point is to classify
// text that FAILED to parse as JSON; and it deliberately is not a bare
// `strings.Contains(criteria, "\"facts\"")`, which the council's guardian seat
// objected to (medium) as a soft match guarding a hard refusal: prose, a nested
// string value, or a check id quoting the word would all trip it.
//
// RELATION TO THE FAN-OUT'S SQL PREFILTER, stated precisely because the first
// version of this comment overclaimed it and the editquality seat caught that.
// The prefilter is `dp.body LIKE '%"facts"%'` over the WHOLE document; this runs
// over the EXTRACTED fence and requires key position. So this predicate is a
// strict SUBSET of the SQL's, which is the safe direction: the SQL over-selects
// rows and Go decides, so nothing the write gate refuses can be a document the
// sweep would silently ignore. It is NOT exact parity and must not be described
// as such.
//
// Residual, unfixed and worth knowing: `extractCriteriaBlock` has its own
// documented landmine (prose naming the criteria fence in backticks hijacks
// fence extraction, load_doc_context_action.go). A mis-extraction upstream can
// still hand this function the wrong text. The guardian seat's point that two
// soft-matching layers feed one refusal stands; tightening THIS layer to key
// position removes the one of the two that was in scope here.
var factsKeyRe = regexp.MustCompile(`"facts"\s*:`)

// factsKeyMentioned reports whether a raw criteria fence declares a `facts` key.
func factsKeyMentioned(criteria string) bool {
	return factsKeyRe.MatchString(criteria)
}

// parseCriteriaFacts reads the fence-level `facts` list off a raw criteria
// string. Returns (nil, nil) for empty or unparseable criteria (fail-open, see
// the file header), and (ids, issues) otherwise — a present-but-malformed
// declaration yields the well-formed subset plus one issue per defect.
func parseCriteriaFacts(criteria string) (ids []string, issues []string) {
	if strings.TrimSpace(criteria) == "" {
		return nil, nil
	}
	var doc criteriaFactsDoc
	if err := json.Unmarshal([]byte(criteria), &doc); err != nil {
		// ⚠ FAIL-OPEN ONLY FOR A FENCE THAT NEVER MENTIONED FACTS.
		//
		// This used to return (nil, nil) unconditionally — no ids AND no issues
		// — which contradicted this file's own header contract ("a fence that
		// DID say something, malformed, is reported as issues") and did real
		// damage two rungs down: planSiteFactDrift's zero-rows warning, added at
		// the council's bug_historian seat's request, is gated on issues or
		// unresolved being non-empty. So one trailing comma in a fence edit
		// disarmed the declaration AND the warning that exists to say a
		// declaration was disarmed, leaving no signal anywhere. Measured
		// 2026-08-24: 1 of 132 current tool PLANs declares anything, so
		// "silently declared nothing" is indistinguishable from the fleet norm.
		// bugs_open/288 defect B.
		if factsKeyMentioned(criteria) {
			return nil, []string{
				"criteria fence is not valid JSON, so its facts declaration is being IGNORED: " + err.Error(),
			}
		}
		return nil, nil
	}
	if len(doc.Facts) == 0 || string(doc.Facts) == "null" {
		return nil, nil
	}
	var raw interface{}
	if err := json.Unmarshal(doc.Facts, &raw); err != nil {
		return nil, []string{"facts is not valid JSON: " + err.Error()}
	}
	return criteriaFactsFromValue(raw)
}
