// FILE: platform/orchestration/actions/discovery_checks/exported_static_criteria.go
//
// An exported door onto the Tier 2 static evaluator, so callers outside this
// package run THE SAME code the tool-acceptance check runs rather than a copy
// of it.
//
// WHY A WRAPPER RATHER THAN A COPY. The experience register needs to evaluate a
// bound entry's criteria and decide whether the fork is verified. Reimplementing
// the anchor rule there would create a second definition of "this check passed",
// and the two would drift — at which point `verified` in the register and
// `passed` in tool acceptance would mean different things while sharing a
// vocabulary. That is the split-contract class (bugs_closed/064) applied to a
// judgement instead of a value, and it is worse, because a divergent judgement
// produces confident disagreement rather than an error.
//
// Nothing here re-implements anything: it decodes the same criteria document
// shape and returns the same outcomes, with the internal types projected onto an
// exported struct.
package discovery_checks

import "encoding/json"

// StaticCheckOutcome is one check's result.
type StaticCheckOutcome struct {
	ID     string `json:"id"`
	Detail string `json:"detail"`
}

// StaticEvaluation is the outcome of applying the Tier 2 anchor rule to a
// criteria document.
//
// SKIPPED IS NOT PASSED. It is a separate list on purpose: a skipped check
// asserts nothing, and the failure this whole area keeps producing is a skip
// being read as a pass. Callers deciding "did this verify?" must require at
// least one PASSED and zero FAILED — never merely "no failures", which is also
// true of a document in which every check was skipped.
type StaticEvaluation struct {
	Passed  []StaticCheckOutcome `json:"passed"`
	Failed  []StaticCheckOutcome `json:"failed"`
	Skipped []StaticCheckOutcome `json:"skipped"`
}

// EvaluateStaticCriteriaJSON applies the Tier 2 static rules to a criteria
// document supplied as JSON, against the deployed page's status and HTML.
//
// It is the same evaluator `ToolAcceptanceCheck` uses; the only difference is
// that the document arrives as JSON rather than already decoded.
func EvaluateStaticCriteriaJSON(criteriaJSON []byte, httpStatus int, html string) (StaticEvaluation, error) {
	var doc criteriaDoc
	if err := json.Unmarshal(criteriaJSON, &doc); err != nil {
		return StaticEvaluation{}, err
	}
	return projectEvaluation(evaluateStaticCriteria(doc, httpStatus, html)), nil
}

func projectEvaluation(ev evaluation) StaticEvaluation {
	out := StaticEvaluation{
		Passed:  make([]StaticCheckOutcome, 0, len(ev.passed)),
		Failed:  make([]StaticCheckOutcome, 0, len(ev.failed)),
		Skipped: make([]StaticCheckOutcome, 0, len(ev.skipped)),
	}
	for _, o := range ev.passed {
		out.Passed = append(out.Passed, StaticCheckOutcome{ID: o.id, Detail: o.detail})
	}
	for _, o := range ev.failed {
		out.Failed = append(out.Failed, StaticCheckOutcome{ID: o.id, Detail: o.detail})
	}
	for _, o := range ev.skipped {
		out.Skipped = append(out.Skipped, StaticCheckOutcome{ID: o.id, Detail: o.detail})
	}
	return out
}
