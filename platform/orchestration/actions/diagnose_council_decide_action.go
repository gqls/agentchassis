// FILE: platform/orchestration/actions/diagnose_council_decide_action.go
//
// F2.1 of the diagnosis→fix loop: the council's DECISION layer. Reviewers are
// LLM steps (parallel opinions, per Q-D); the aggregation is deliberately
// DETERMINISTIC Go — a decision you can read off the rules beats a third
// model opinion about two model opinions. Q-D (owner, 2026-07-07): all
// opinions are advisory by default; a hard_veto reviewer's negative verdict
// is a BLOCK. v1 places the hard-veto flag in step config
// (hard_veto_from: [reviewer names]); the definition-column vs council-config
// placement question stays open for F2.2.
//
// Decision rules, in order:
//  1. any hard-veto reviewer says veto            → "rejected"
//  2. any reviewer says veto (advisory veto)      → "rejected"
//  3. any reviewer says object                    → "revise"
//  4. all approve                                 → "approved"
//
// The report is persisted to diagnosis_artifacts (kind='council_report') on
// the SAME correlation_id as the diagnosis and the plan, and the decision is
// returned so the workflow can route on it. Like the plan-persist action and
// unlike the bundle write-through, a malformed reviewer verdict FAILS the
// step: a council that cannot read its reviewers must not wave a plan through.
//
// F2.3 adds the routing flags for the decision router: should_revise (revise
// with rounds left → repropose), should_reframe (first rejection with rounds
// left → ONE narrower replan before escalating). See applyCouncilCaps.
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var DiagnoseCouncilDecideInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"fix_correlation_id"},
	Optional:   []string{"review_fields", "hard_veto_from", "max_rounds"},
	Defaults:   map[string]interface{}{"max_rounds": 2},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_council_decide", DiagnoseCouncilDecideInputSpec)
}

// councilReview is one reviewer's structured opinion (the verdict-wire-style
// contract the F2 design asked for: verdict + objections + suggestions).
type councilReview struct {
	Reviewer   string `json:"reviewer"`
	Verdict    string `json:"verdict"` // approve | object | veto
	Objections []struct {
		Edit     int    `json:"edit"` // 1-based index into the plan's edits; 0 = plan-wide
		Problem  string `json:"problem"`
		Severity string `json:"severity,omitempty"` // low | medium | high
	} `json:"objections,omitempty"`
	Missing []string `json:"missing,omitempty"` // mechanisms the plan should cover but doesn't
	Notes   string   `json:"notes,omitempty"`
	// Degraded marks a review recovered from a TRUNCATED response by closing its
	// open brackets: the verdict and any objections before the cut are real, but
	// anything after it is missing. Surfaced in the report so a reader can tell a
	// complete opinion from a salvaged fragment.
	Degraded bool `json:"degraded,omitempty"`
}

var councilVerdicts = map[string]bool{"approve": true, "object": true, "veto": true}

// salvageTruncatedReview tries to recover a usable opinion from a review cut off
// at max_tokens, reusing repairTruncatedJSON (apply_adoption_plan_action.go) —
// the same helper the truncation family has needed in three other places.
//
// This works as often as it does because of field ORDER: councilReview puts
// `reviewer` and `verdict` first, and models emit them first, so the load-bearing
// part of a review is usually intact and only trailing prose is lost. That is
// also the limit of it — a recovered review may be missing objections the seat
// had not written yet, which is why the caller marks it Degraded rather than
// treating it as complete.
//
// Returns ok=false unless a review with a RECOGNISED verdict comes back: a
// fragment that repairs into valid JSON with no usable verdict is not an
// opinion, and must be counted unreadable rather than quietly become one.
func salvageTruncatedReview(rb []byte) (councilReview, bool) {
	var rv councilReview
	repaired := repairTruncatedJSON(string(rb))
	if repaired == "" {
		return rv, false
	}
	if err := json.Unmarshal([]byte(repaired), &rv); err != nil {
		return rv, false
	}
	rv.Verdict = strings.ToLower(strings.TrimSpace(rv.Verdict))
	if !councilVerdicts[rv.Verdict] {
		return rv, false
	}
	return rv, true
}

// DiagnoseCouncilDecideAction aggregates the reviewer steps' verdicts.
func DiagnoseCouncilDecideAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "diagnose_council_decide"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		DiagnoseCouncilDecideInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	corr := strings.TrimSpace(inputs.Get("fix_correlation_id"))
	if corr == "" {
		return nil, fmt.Errorf("fix_correlation_id is empty")
	}

	reviewFields := configStringSlice(params.StepConfig.Config, "review_fields", nil)
	if len(reviewFields) == 0 {
		return nil, fmt.Errorf("no review_fields configured — a council with no reviewers decides nothing")
	}
	hardVeto := map[string]bool{}
	for _, r := range configStringSlice(params.StepConfig.Config, "hard_veto_from", nil) {
		hardVeto[strings.ToLower(strings.TrimSpace(r))] = true
	}

	var reviews []councilReview
	abstained := 0
	// Seats that RAN and produced something unreadable — kept separate from
	// `abstained` on purpose. An abstention is a seat the relevance filter skipped,
	// which is information ("not applicable"); an unreadable seat is an opinion we
	// were owed and lost, which is the absence of information. Conflating them
	// would let a lost opinion read as a considered non-objection.
	var unreadable []string
	for _, field := range reviewFields {
		raw := datahelpers.ExtractNestedField(params.CollectedData, field)
		if raw == nil {
			// A configured reviewer produced no output. Historically a hard
			// error — a council that cannot read its reviewers must not wave a
			// plan through. But the stage-3 relevance filter
			// (select_review_panel + per-seat conditionals) deliberately SKIPS
			// seats not relevant to a given fix, leaving their field absent. A
			// skipped seat is an ABSTENTION, not a failure: it did not object,
			// so it does not gate. Tolerate it, and count it so the decision
			// stays auditable. (edit-quality and guardian are always-on and
			// never skipped, so a council can never end up with zero opinions.)
			abstained++
			logger.Debug("diagnose_council_decide: reviewer abstained (field absent — skipped by relevance filter)",
				zap.String("field", field))
			continue
		}
		rb, err := planBytes(raw) // same map-or-string defence as the plan itself
		if err != nil {
			return nil, fmt.Errorf("reviewer output at %q not serialisable: %w", field, err)
		}
		var rv councilReview
		if err := json.Unmarshal(rb, &rv); err != nil {
			if !json.Valid(rb) {
				// TRUNCATED (bugs_open/019). Historically a hard error, which
				// discarded every other seat's completed review and returned no
				// verdict at all. Two things are true at once and the fix has to
				// honour both: a council that cannot read a reviewer must not wave
				// a plan through, AND one seat overrunning its cap must not destroy
				// the round. So: try to recover the opinion; if that fails, record
				// the seat as UNREADABLE and carry it into the decision, where it
				// blocks an approval further down. Never a silent drop.
				salvaged, ok := salvageTruncatedReview(rb)
				if ok {
					salvaged.Degraded = true
					if strings.TrimSpace(salvaged.Reviewer) == "" {
						salvaged.Reviewer = field
					}
					logger.Warn("diagnose_council_decide: reviewer output was TRUNCATED — verdict recovered from the partial, later objections may be missing",
						zap.String("field", field),
						zap.String("recovered_verdict", salvaged.Verdict),
						zap.Int("objections_recovered", len(salvaged.Objections)))
					reviews = append(reviews, salvaged)
					continue
				}
				unreadable = append(unreadable, field)
				logger.Warn("diagnose_council_decide: reviewer output is UNREADABLE (invalid JSON, likely truncated at max_tokens) and could not be salvaged — recorded as unreadable; an approval cannot stand alongside it",
					zap.String("field", field),
					zap.Int("bytes", len(rb)),
					zap.Error(err))
				continue
			}
			return nil, fmt.Errorf("reviewer output at %q does not match the review schema: %w", field, err)
		}
		rv.Verdict = strings.ToLower(strings.TrimSpace(rv.Verdict))
		if !councilVerdicts[rv.Verdict] {
			return nil, fmt.Errorf("reviewer %q returned unknown verdict %q (want approve|object|veto)", rv.Reviewer, rv.Verdict)
		}
		if strings.TrimSpace(rv.Reviewer) == "" {
			rv.Reviewer = field
		}
		reviews = append(reviews, rv)
	}

	if len(reviews) == 0 {
		// A council with zero opinions must never read as "all approve". Fail
		// closed. No longer merely defensive: with truncation tolerated upstream,
		// a round where every seat that ran was unreadable reaches here, and that
		// is precisely the state that must not become a verdict.
		return nil, fmt.Errorf("no reviewer produced a readable opinion (%d abstained, %d unreadable: %s) — a council with no opinions cannot decide",
			abstained, len(unreadable), strings.Join(unreadable, ", "))
	}

	decision, decidedBy := decideCouncil(reviews, hardVeto)

	// An unreadable seat must never be the difference between revise and approve.
	// The hard error this replaces was protecting a real property — "a council
	// that cannot read its reviewers must not wave a plan through" — and it is
	// kept here, at the only point where it actually matters, instead of by
	// voiding the round. Note the asymmetry is deliberate: an objection or veto
	// from a seat that WAS read stays decisive and is not softened, so this can
	// only ever make the outcome more conservative, never less.
	if decision == "approved" && len(unreadable) > 0 {
		logger.Warn("diagnose_council_decide: downgrading approve->revise because a seat could not be read",
			zap.Strings("unreadable", unreadable),
			zap.Int("readable_reviews", len(reviews)))
		decision = "revise"
		decidedBy = fmt.Sprintf("unreadable reviewer(s): %s", strings.Join(unreadable, ", "))
	}

	report, _ := json.Marshal(map[string]interface{}{
		"decision":   decision,
		"decided_by": decidedBy,
		"reviews":    reviews,
		"abstained":  abstained,
		"unreadable": unreadable,
	})
	metadata, _ := json.Marshal(map[string]interface{}{
		"decision":   decision,
		"reviewers":  len(reviews),
		"abstained":  abstained,
		"unreadable": len(unreadable),
	})
	if _, err := params.DB.ExecContext(ctx, `
		INSERT INTO diagnosis_artifacts (
			correlation_id, orchestration_id, iteration, kind, body,
			metadata, source_agent, created_by
		) VALUES ($1, $2, 0, 'council_report', $3, $4::jsonb, $5, 'diagnose_council_decide')
	`, corr,
		nullIfEmpty(params.ExecutionContext.OrchestrationID),
		string(report),
		string(metadata),
		nullIfEmpty(params.AgentType),
	); err != nil {
		return nil, fmt.Errorf("persist council report: %w", err)
	}

	// The revise loop's counter IS the council_report count for THIS PROPOSER
	// RUN — one per round, just written, so this reflects the current round.
	// Sourcing it from the durable artifact avoids threading loop state through
	// the workflow. Scoped by orchestration_id, NOT correlation alone: the
	// correlation belongs to the DIAGNOSIS and accumulates reports across
	// proposer re-runs (e08c5b01 had a report from a prior run, which would have
	// started this run at round 2 and exhausted it before any repropose). A
	// count failure must not strand a decided council: fall back to round 1,
	// which caps the loop at once more.
	round := 1
	if err := params.DB.QueryRowContext(ctx,
		`SELECT count(*) FROM diagnosis_artifacts
		 WHERE correlation_id = $1 AND kind = 'council_report' AND orchestration_id = $2`,
		corr, nullIfEmpty(params.ExecutionContext.OrchestrationID)).Scan(&round); err != nil {
		logger.Warn("diagnose_council_decide: round count failed; treating as round 1", zap.Error(err))
		round = 1
	}

	// F2.3 reframe-once bookkeeping: how many of THIS RUN's reports (the one
	// just written included) were REJECTED? The persisted metadata carries the
	// raw decision, so 'rejected' is queryable directly. Fail CLOSED on a count
	// error: a failure must not grant extra LLM rounds — the safe terminal is
	// escalation (a human sees the package), not another reframe.
	rejectedCount := 0
	if decision == "rejected" {
		if err := params.DB.QueryRowContext(ctx,
			`SELECT count(*) FROM diagnosis_artifacts
			 WHERE correlation_id = $1 AND kind = 'council_report'
			   AND orchestration_id = $2 AND metadata->>'decision' = 'rejected'`,
			corr, nullIfEmpty(params.ExecutionContext.OrchestrationID)).Scan(&rejectedCount); err != nil {
			logger.Warn("diagnose_council_decide: rejected count failed; treating reframe as spent", zap.Error(err))
			rejectedCount = 2
		}
	}

	maxRounds := datahelpers.GetIntField(params.StepConfig.Config, "max_rounds", 2)
	decision, decidedBy, shouldRevise, shouldReframe := applyCouncilCaps(decision, decidedBy, round, maxRounds, rejectedCount)

	logger.Info("diagnose_council_decide: decided",
		zap.String("correlation_id", corr),
		zap.String("decision", decision),
		zap.String("decided_by", decidedBy),
		zap.Int("reviewers", len(reviews)),
		zap.Int("abstained", abstained),
		zap.Strings("unreadable", unreadable),
		zap.Int("round", round),
		zap.Bool("should_revise", shouldRevise),
		zap.Bool("should_reframe", shouldReframe))

	return map[string]interface{}{
		"decision":       decision,
		"decided_by":     decidedBy,
		"reviewers":      len(reviews),
		"abstained":      abstained,
		"unreadable":     len(unreadable),
		"unreadable_at":  unreadable,
		"round":          round,
		"should_revise":  shouldRevise,
		"should_reframe": shouldReframe,
	}, nil
}

// applyCouncilCaps maps the raw council decision onto the loop's routing flags.
// Extracted as a pure function so the tests exercise the real mapping.
//   - revise with rounds left  → should_revise (repropose loop)
//   - revise out of rounds     → 'exhausted' — terminal, named so a human sees
//     the loop gave up rather than silently approved
//   - rejected, FIRST time, rounds left → should_reframe: ONE narrower replan
//     (F2.3 — benchmark run 8c770fd5's guardian veto was correct, and reproposing
//     the same shape would be vetoed again; a reframe changes the shape)
//   - rejected again, or out of rounds  → terminal (workflow escalates)
//
// max_rounds bounds TOTAL review cycles: a reframe consumes a round like any
// other, so revise + reframe together can never exceed the cap.
func applyCouncilCaps(decision, decidedBy string, round, maxRounds, rejectedCount int) (string, string, bool, bool) {
	shouldRevise := decision == "revise" && round < maxRounds
	if decision == "revise" && !shouldRevise {
		decision = "exhausted"
		decidedBy = fmt.Sprintf("%s — revise cap reached (%d rounds)", decidedBy, maxRounds)
	}
	shouldReframe := decision == "rejected" && rejectedCount <= 1 && round < maxRounds
	return decision, decidedBy, shouldRevise, shouldReframe
}

// decideCouncil applies the ordered rules. decidedBy names the rule that fired
// so the report is auditable without re-deriving.
func decideCouncil(reviews []councilReview, hardVeto map[string]bool) (decision, decidedBy string) {
	for _, r := range reviews {
		if r.Verdict == "veto" && hardVeto[strings.ToLower(r.Reviewer)] {
			return "rejected", "hard veto from " + r.Reviewer
		}
	}
	for _, r := range reviews {
		if r.Verdict == "veto" {
			return "rejected", "veto from " + r.Reviewer
		}
	}
	for _, r := range reviews {
		if r.Verdict == "object" {
			return "revise", "objection from " + r.Reviewer
		}
	}
	return "approved", "all reviewers approve"
}
