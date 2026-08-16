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
