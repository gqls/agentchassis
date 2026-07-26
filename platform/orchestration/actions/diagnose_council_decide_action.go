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
//  1. any hard-veto reviewer says veto                 → "rejected"
//  2. any reviewer says veto (advisory veto)           → "rejected"
//  3. any reviewer raises a HIGH-severity objection    → "revise"
//  4. otherwise (only low/medium objections, or none)  → "approved"
//
// Rule 3 changed 2026-07-22 (owner ruling). It used to be "any reviewer says
// object → revise", with severity ignored entirely — so a single low nit from
// one of ~16 seats blocked exactly as hard as a real flaw, and an approval was
// effectively unreachable (measured: ~4.5% approval rate over a week; 67% of
// revise rounds carried NO high-severity objection at all — they were blocked by
// low/medium nits). Now only a HIGH-severity objection gates; low/medium
// objections are ADVISORY — still recorded in the report and returned to the
// proposer, but they do not force a revise. The objections keep their value; a
// minor one just stops being a block. Deliberately conservative against a minor
// label hiding a real problem: a Degraded (truncated) object still gates — its
// high-severity objection may have been cut off before we saw it — and an object
// with an un-graded / unrecognised severity still gates. Only an EXPLICITLY
// low/medium objection is waved through. This corrects every council at once
// (fix-proposer, gate, experience, concept-register) because they share it.
//
// The report is persisted to diagnosis_artifacts (kind='council_report') on
// the SAME correlation_id as the diagnosis and the plan, and the decision is
// returned so the workflow can route on it.
//
// A malformed reviewer output used to FAIL the step, on the principle that a
// council which cannot read its reviewers must not wave a plan through. The
// principle is right and is kept; failing the step was the wrong place to
// enforce it. The council's whole value is that it is MANY INDEPENDENT
// OPINIONS, so one seat's bad output must cost one seat — never the round, and
// least of all at the final step, after every seat has been run and paid for.
// So every per-seat failure mode now converges on the same three-way handling
// (bugs_open/019 truncation, bugs_open/036 schema slip):
//
//	recover the opinion if it is there  → count it, marked Degraded
//	no opinion recoverable              → count the seat UNREADABLE
//	zero readable opinions in the round → THEN fail, because that is the one
//	                                      state that must not become a verdict
//
// The original principle is enforced where it actually bites: an unreadable
// seat blocks an approval (downgraded to revise) further down, so this can only
// ever make an outcome more conservative, never less.
//
// F2.3 adds the routing flags for the decision router: should_revise (revise
// with rounds left → repropose), should_reframe (first rejection with rounds
// left → ONE narrower replan before escalating). See applyCouncilCaps.
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
	Reviewer   string             `json:"reviewer"`
	Verdict    string             `json:"verdict"` // approve | object | veto
	Objections []councilObjection `json:"objections,omitempty"`
	Missing    []string           `json:"missing,omitempty"` // mechanisms the plan should cover but doesn't
	Notes      string             `json:"notes,omitempty"`
	// Degraded marks a review recovered from a TRUNCATED response by closing its
	// open brackets: the verdict and any objections before the cut are real, but
	// anything after it is missing. Surfaced in the report so a reader can tell a
	// complete opinion from a salvaged fragment.
	Degraded bool `json:"degraded,omitempty"`
}

// councilObjection is one thing a reviewer wants changed, and which edit it is
// about.
type councilObjection struct {
	Edit     objectionEdit `json:"edit"`
	Problem  string        `json:"problem"`
	Severity string        `json:"severity,omitempty"` // low | medium | high
}

// objectionEdit points at WHICH edit an objection concerns. The wire contract
// asks for a 1-based index into the plan's edits (0 = plan-wide), and it used to
// be a plain int — which meant exactly one of the three registers a model
// naturally answers "which edit?" in would parse. Live evidence (bugs_open/036,
// three voided rounds, all the same seat): `3`, `"3"`, and free text such as
// "plan-level (deploy verification)" or "risks/summary (item 5)". The last form
// is the interesting one — it is not a malformed index, it is a reviewer saying
// the objection is about the PLAN rather than any single edit, which the contract
// already spells 0. So the tolerant reading is not a leniency hack; it recovers
// the meaning the strict one discarded.
//
// Unparseable pointers land on 0 (plan-wide) rather than failing, because an
// objection whose target is unclear is still an objection, and the problem text
// is the part that carries the review's value. Raw keeps the reviewer's own token
// so the persisted report shows what was actually written rather than a laundered
// 0 — the report is read by humans deciding whether the council was fair.
type objectionEdit struct {
	Index int             // 1-based index into the plan's edits; 0 = plan-wide/unresolved
	Raw   json.RawMessage // exactly what the reviewer wrote
}

// UnmarshalJSON never returns an error, by design: this field must not be able to
// cost a review, let alone a round.
func (e *objectionEdit) UnmarshalJSON(b []byte) error {
	e.Raw = append(json.RawMessage(nil), b...)
	e.Index = 0

	// A bare number, including the 3.0 an LLM sometimes emits for 3.
	var num json.Number
	if err := json.Unmarshal(b, &num); err == nil {
		if f, ferr := num.Float64(); ferr == nil {
			e.Index = int(f)
		}
		return nil
	}
	// A number in string clothing ("3"), or free text — which resolves to
	// plan-wide.
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if n, aerr := strconv.Atoi(strings.TrimSpace(s)); aerr == nil {
			e.Index = n
		}
	}
	return nil
}

// MarshalJSON round-trips the reviewer's own token so the council report is a
// faithful record. Falls back to the index for a value built in Go (tests).
func (e objectionEdit) MarshalJSON() ([]byte, error) {
	if len(e.Raw) > 0 {
		return e.Raw, nil
	}
	return json.Marshal(e.Index)
}

var councilVerdicts = map[string]bool{"approve": true, "object": true, "veto": true}

// markerFieldFor maps a reviewer field path to its step's __truncated marker:
// "review_editquality.result" -> "review_editquality.__truncated". The marker is
// a SIBLING of the terminal segment (ExecuteLLMPromptAction stamps it on the
// step's result map), so a path with no parent has nowhere to look — return "".
func markerFieldFor(field string) string {
	dot := strings.LastIndex(field, ".")
	if dot <= 0 {
		return ""
	}
	return field[:dot] + ".__truncated"
}

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

// salvageMistypedReview recovers an opinion from a review that is VALID, COMPLETE
// JSON which simply does not fit the schema — a field emitted in the wrong
// register (bugs_open/036). This is the sibling of salvageTruncatedReview and
// deliberately NOT the same mechanism: there is nothing to repair here, the
// document is whole, so the recovery is to decode field by field and keep what
// decodes instead of losing the review to whichever field did not.
//
// Reviewer and verdict are read strictly because they are load-bearing; the rest
// is best-effort, since an objection list that will not decode costs the detail,
// not the opinion. Returns ok=false unless a RECOGNISED verdict comes back — a
// document with no usable verdict is not an opinion, and must be counted
// unreadable rather than quietly become one (same bar as the truncation path).
func salvageMistypedReview(rb []byte) (councilReview, bool) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(rb, &fields); err != nil {
		return councilReview{}, false
	}
	var rv councilReview
	var verdict string
	if err := json.Unmarshal(fields["verdict"], &verdict); err != nil {
		return rv, false
	}
	rv.Verdict = strings.ToLower(strings.TrimSpace(verdict))
	if !councilVerdicts[rv.Verdict] {
		return rv, false
	}
	// Best-effort from here: a field that will not decode is left at its zero
	// value rather than taking the review down with it. Note encoding/json
	// continues past a TYPE error (unlike a syntax error) and keeps what did
	// decode, so a mistyped `severity` costs that field and not the objection
	// around it — the ignored errors below are doing more work than they look.
	_ = json.Unmarshal(fields["reviewer"], &rv.Reviewer)
	// A seat emitting `objections` as ONE object rather than a list is the same
	// wrong-register slip one level up, and dropping it is not free: the objection
	// list is what goes back to the proposer in a revise round, so losing it costs
	// the round its content even though the verdict survives. Coerce it.
	// (Raised as an objection by the edit-quality seat on council round 80cdd428 —
	// the fix's own council caught a gap in the fix.)
	if err := json.Unmarshal(fields["objections"], &rv.Objections); err != nil {
		var one councilObjection
		if json.Unmarshal(fields["objections"], &one) == nil {
			rv.Objections = []councilObjection{one}
		}
	}
	_ = json.Unmarshal(fields["missing"], &rv.Missing)
	_ = json.Unmarshal(fields["notes"], &rv.Notes)
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
	// Seats damaged by a TRUNCATION specifically, accumulated for the durable
	// record written after the decision (bugs_open/076 residual). Separate from
	// `unreadable` because the two overlap without coinciding: a truncation can
	// be salvaged (readable, but degraded) or fatal (unreadable), and an
	// unreadable seat can have causes that have nothing to do with truncation.
	var truncationDamage []truncationDegradation
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
			// One seat's output being unserialisable is that seat's loss, not the
			// round's — same rule as every other per-seat failure below.
			unreadable = append(unreadable, field)
			logger.Warn("diagnose_council_decide: reviewer output is UNREADABLE (not serialisable) — recorded as unreadable; an approval cannot stand alongside it",
				zap.String("field", field), zap.Error(err))
			continue
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
					truncationDamage = append(truncationDamage, truncationDegradation{
						Field: field, Reviewer: salvaged.Reviewer, Verdict: salvaged.Verdict,
						Objections: len(salvaged.Objections), Branch: "salvaged_from_invalid_json",
					})
					reviews = append(reviews, salvaged)
					continue
				}
				unreadable = append(unreadable, field)
				logger.Warn("diagnose_council_decide: reviewer output is UNREADABLE (invalid JSON, likely truncated at max_tokens) and could not be salvaged — recorded as unreadable; an approval cannot stand alongside it",
					zap.String("field", field),
					zap.Int("bytes", len(rb)),
					zap.Error(err))
				truncationDamage = append(truncationDamage, truncationDegradation{
					Field: field, Branch: "unsalvageable_invalid_json",
				})
				continue
			}
			// SCHEMA SLIP (bugs_open/036). The JSON is complete and well-formed but
			// a field arrived in the wrong register, so 019's salvage cannot help —
			// there is nothing to repair. This used to return, discarding every other
			// seat's completed review at the LAST step of the round, after all of them
			// had been paid for. It is the same principle as the truncation branch and
			// gets the same answer: recover the opinion if it is there, else record
			// the seat unreadable. One seat's malformed output costs one seat.
			salvaged, ok := salvageMistypedReview(rb)
			if ok {
				salvaged.Degraded = true
				if strings.TrimSpace(salvaged.Reviewer) == "" {
					salvaged.Reviewer = field
				}
				logger.Warn("diagnose_council_decide: reviewer output did not match the schema — verdict recovered field-by-field, some detail may be dropped",
					zap.String("field", field),
					zap.String("recovered_verdict", salvaged.Verdict),
					zap.Int("objections_recovered", len(salvaged.Objections)),
					zap.Error(err))
				reviews = append(reviews, salvaged)
				continue
			}
			unreadable = append(unreadable, field)
			logger.Warn("diagnose_council_decide: reviewer output does not match the review schema and no verdict could be recovered — recorded as unreadable; an approval cannot stand alongside it",
				zap.String("field", field),
				zap.Int("bytes", len(rb)),
				zap.Error(err))
			continue
		}
		rv.Verdict = strings.ToLower(strings.TrimSpace(rv.Verdict))
		if !councilVerdicts[rv.Verdict] {
			// A well-formed review carrying a verdict outside the contract
			// ("approve-with-comments", "") is an opinion we cannot count. Deliberately
			// NOT normalised into the nearest legal verdict — guessing what a seat
			// meant is how a veto becomes an approval. Unreadable, which blocks
			// approval below without inventing a position for the seat.
			unreadable = append(unreadable, field)
			logger.Warn("diagnose_council_decide: reviewer returned an unrecognised verdict — recorded as unreadable; an approval cannot stand alongside it",
				zap.String("field", field),
				zap.String("reviewer", rv.Reviewer),
				zap.String("verdict", rv.Verdict))
			continue
		}
		if strings.TrimSpace(rv.Reviewer) == "" {
			rv.Reviewer = field
		}
		// A partial that PARSES is still a partial. A review cut mid-array can
		// close into valid JSON with a recognised verdict and silently missing
		// objections — so a clean parse is not proof of a complete opinion
		// (council round 2eed453a: three seats converged on exactly this gap).
		// The step that tolerated the truncation stamped __truncated beside
		// .result; consult it, so luck of the cut cannot outrank the record.
		if mf := markerFieldFor(field); mf != "" {
			if tm, ok := datahelpers.ExtractNestedField(params.CollectedData, mf).(bool); ok && tm {
				rv.Degraded = true
				logger.Warn("diagnose_council_decide: review parsed cleanly but the step recorded a TRUNCATION — marking degraded; trailing objections may be missing",
					zap.String("field", field),
					zap.String("verdict", rv.Verdict))
				truncationDamage = append(truncationDamage, truncationDegradation{
					Field: field, Reviewer: rv.Reviewer, Verdict: rv.Verdict,
					Objections: len(rv.Objections), Branch: "producer_marker",
				})
			}
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

	// Deliberately AFTER the report insert: the decision is already durable, so a
	// failure to record the damage cannot cost a decided council.
	recordTruncationDegradation(ctx, params, corr, decision, truncationDamage, logger)

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

// truncationDegradation is one seat whose opinion was damaged by a TRUNCATION —
// recovered from a partial, or lost outright. Collected during the review loop
// and written after the decision by recordTruncationDegradation.
type truncationDegradation struct {
	Field      string
	Reviewer   string
	Verdict    string
	Objections int
	// Branch names HOW the truncation presented, because the three cases carry
	// different amounts of loss and are worth telling apart in the data:
	//   producer_marker            — parsed cleanly; the producing step stamped
	//                                __truncated, so trailing objections may be gone
	//   salvaged_from_invalid_json — JSON was cut mid-structure and closed by hand
	//   unsalvageable_invalid_json — cut too early to recover; the seat is lost
	Branch string
}

// recordTruncationDegradation persists, per damaged seat, the fact that a
// consumer DEGRADED on a truncated response — to agent_error_log, where the
// immune-system sweep and the dashboards already look.
//
// WHY THIS EXISTS (bugs_open/076 residual). The tolerate-and-mark contract has
// two halves and only one of them was durable. The PRODUCER half is recorded
// permanently: ExecuteLLMPromptAction writes an llm_call_log row prefixed
// "TOLERATED (step continued on the partial):". The CONSUMER half — what this
// action then did about it — existed only as a zap.Warn, and a pod log dies with
// its pod (the same argument that put recordUnknownVerdict in
// complete_work_item_verification.go, council objection bug_historian
// 2026-07-18). So "has a consumer ever actually degraded?" was unanswerable from
// data, which is a poor position for a mechanism whose entire purpose is to make
// an invisible failure visible. The orchestration_states copy of the marker is
// not a substitute: it is pruned at 24h, so a zero there is not evidence of never.
//
// Severity is 'warning', not 'error': degrading is the CORRECT behaviour here and
// the round is sound. What needs a human eventually is the pattern — a seat that
// truncates repeatedly wants a higher max_tokens, not a nightly salvage.
//
// Scope is truncation only. The schema-slip salvage (bugs_closed/036) shares this
// loop but is a different defect with a different remedy, and folding it in would
// make the error_code mean two things.
//
// Best-effort by design, like every other durable-record helper in this package:
// a failure to record must never change a decision already made and already
// persisted. It warns and returns.
func recordTruncationDegradation(ctx context.Context, params ActionParams, corr, decision string, damage []truncationDegradation, logger *zap.Logger) {
	if len(damage) == 0 || params.DB == nil {
		return
	}

	for _, d := range damage {
		contextJSON, _ := json.Marshal(map[string]interface{}{
			"correlation_id":       corr,
			"review_field":         d.Field,
			"reviewer":             d.Reviewer,
			"recovered_verdict":    d.Verdict,
			"objections_recovered": d.Objections,
			"branch":               d.Branch,
			"council_decision":     decision,
		})
		if contextJSON == nil {
			contextJSON = []byte("{}")
		}

		if _, err := params.DB.ExecContext(ctx, `
			INSERT INTO agent_error_log (
				orchestration_id, agent_type, agent_id, pod_name,
				step_name, action, error_message, error_code, severity, context
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
		`,
			nullIfEmpty(params.ExecutionContext.OrchestrationID),
			params.ExecutionContext.Sender.AgentType,
			params.ExecutionContext.Sender.AgentID,
			params.ExecutionContext.Sender.PodName,
			params.ExecutionContext.StepName,
			"diagnose_council_decide",
			"council seat '"+d.Field+"' was damaged by a TRUNCATED response ("+d.Branch+") — the opinion counted here is partial or lost",
			"TRUNCATION_DEGRADED_REVIEW",
			"warning",
			string(contextJSON),
		); err != nil {
			logger.Warn("recordTruncationDegradation: could not persist to agent_error_log (the decision is unaffected)",
				zap.String("field", d.Field),
				zap.Error(err))
		}
	}
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

// severityGates reports whether an objection's severity is one that FORCES a
// revise. Owner ruling 2026-07-22: only "high" gates; "low"/"medium" are
// advisory. Anything that is NOT explicitly low or medium — unset, "", or an
// unrecognised value — gates: the change only wants to wave through an objection
// a reviewer EXPLICITLY marked minor, never one it forgot to grade.
func severityGates(sev string) bool {
	switch strings.ToLower(strings.TrimSpace(sev)) {
	case "low", "medium":
		return false
	default:
		return true // "high", unset, or anything unrecognised
	}
}

// objectionGates reports whether an `object` review should force a revise rather
// than be recorded as an advisory note. Only meaningful for verdict "object"
// (veto and approve are handled by their own rules above). Two conservative
// carve-outs so a minor label cannot hide a real problem:
//   - a Degraded review was recovered from a truncated response, so a
//     high-severity objection it raised may have been cut off before we ever saw
//     it — a Degraded object always gates;
//   - an object with no gradable objection at all is not "explicitly minor", so
//     it gates. This also preserves the pre-severity behaviour for a bare object.
func objectionGates(r councilReview) bool {
	if r.Verdict != "object" {
		return false
	}
	if r.Degraded || len(r.Objections) == 0 {
		return true
	}
	for _, o := range r.Objections {
		if severityGates(o.Severity) {
			return true
		}
	}
	return false
}

// decideCouncil applies the ordered rules. decidedBy names the rule that fired
// so the report is auditable without re-deriving. A low/medium-only object does
// not gate (rule 3, changed 2026-07-22) — see the file header for why.
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
	// Rule 3: only a HIGH-severity (or un-graded/degraded) objection gates. A
	// low/medium objection is advisory — it rides along in the persisted report's
	// `reviews` and reaches the proposer, but it does not force a revise.
	var firstGate string
	gating, advisory := 0, 0
	for _, r := range reviews {
		if r.Verdict != "object" {
			continue
		}
		if objectionGates(r) {
			gating++
			if firstGate == "" {
				firstGate = r.Reviewer
			}
		} else {
			advisory++
		}
	}
	if gating > 0 {
		return "revise", "gating objection from " + firstGate
	}
	if advisory > 0 {
		return "approved", fmt.Sprintf("approved with %d advisory objection(s) — none high-severity", advisory)
	}
	return "approved", "all reviewers approve"
}
