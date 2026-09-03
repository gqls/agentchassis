// Command reasoningset transforms the fix-loop's persisted reasoning into JSONL
// training/eval records.
//
// It reads the tagged JSON lines produced by extract.sql on stdin and writes one
// record per reasoning STEP, carrying provenance, three families of labels, and
// a guard block pairing the model's RAW verdict against the loop's COERCED one.
//
//	kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
//	  psql -U clients_user -d clients_db -At -f - < cmd/reasoningset/extract.sql \
//	  | go run ./cmd/reasoningset --labels labels.json --out reasoning_v1.jsonl
//
// Deliberately runs OUTSIDE the cluster and touches no DB directly — psql
// extracts, this transforms (the cmd/claimscan idiom). Running it as a pod would
// re-introduce the ephemeral-file problem training_data_export.go:3-8 retired.
//
// Bad rows are FLAGGED, never dropped: a file that silently omits the 19 poisoned
// repropose rows reads as "we have no repropose data". Consumers filter on
// input_complete.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/aiservice"
)

// ── the record we emit ──────────────────────────────────────────────────────

type Record struct {
	TrajectoryID string `json:"trajectory_id"`
	RunID        string `json:"run_id"`
	StepIndex    int    `json:"step_index"`
	Task         string `json:"task"`
	StepName     string `json:"step_name"`

	InputState   string          `json:"input_state"`
	Reasoning    json.RawMessage `json:"reasoning,omitempty"`
	ReasoningRaw string          `json:"reasoning_raw,omitempty"`
	Decision     string          `json:"decision"`

	InputComplete bool   `json:"input_complete"`
	ExcludeReason string `json:"exclude_reason,omitempty"`

	Guard      *Guard     `json:"guard,omitempty"`
	Labels     Labels     `json:"labels"`
	Provenance Provenance `json:"provenance"`
}

// Guard pairs the model's raw verdict against the one the loop actually used.
// pkg/diagnose/step.go:67-101 degrades a verdict that lacks citations or an
// evidence family and PREPENDS a diagnostic sentence to NeededEvidence — a
// natural-language label for why this reasoning was invalid, generated on real
// model output. It is the most distinctive signal in the corpus.
//
// ALIGNMENT IS NOT FREE. llm_call_log has no iteration column, so a verdict call
// can only be matched to its trail entry by order — and that is only sound when
// the counts agree. They frequently do not: runs exist with 5 verdict calls and
// a trail of length 1 (retries and failed calls leave no trail entry). Pairing
// by index regardless produced a raw UNVERIFIABLE against a coerced CONFIRMED,
// which the coercion logic cannot emit — it only ever degrades. So when the
// counts disagree we record Alignment and assert nothing.
type Guard struct {
	RawDecision     string `json:"raw_decision"`
	CoercedDecision string `json:"coerced_decision,omitempty"`
	Tripped         *bool  `json:"tripped,omitempty"`
	Diagnostic      string `json:"diagnostic,omitempty"`
	StopReason      string `json:"stop_reason,omitempty"`
	Alignment       string `json:"alignment,omitempty"`
}

type Labels struct {
	SelfOutcome    string `json:"self_outcome,omitempty"`
	BenchmarkGrade string `json:"benchmark_grade,omitempty"`
	// SubsystemGrade grades a MECHANISM a run exercised (the code-lookup tier,
	// the council) rather than the diagnosis. Kept separate from
	// BenchmarkGrade so a subsystem PASS can never be read as reasoning quality.
	SubsystemGrade string `json:"subsystem_grade,omitempty"`
	Terminal       string `json:"terminal,omitempty"`

	// feed-triage lane. ScoreOutcomeDivergence marks a judgement the triage
	// scored highly that was rejected anyway — the interesting subset, because
	// something OTHER than the score decided it and we do not yet know what.
	ScoreOutcomeDivergence bool `json:"score_outcome_divergence,omitempty"`

	// council lane, derived — no hand-labelling required.
	// RoundDecision is what the council concluded; Dissent is whether THIS seat
	// went against it. Dissent is the load-bearing signal: where seats split,
	// the reasoning had to do work.
	RoundDecision string `json:"round_decision,omitempty"`
	// RoundDecisionRule names which council decision rule produced RoundDecision:
	// "any_objection_gates" before the 2026-07-22 severity change, "high_severity_gates"
	// after. `round_decision` MUST NOT be pooled across the two — a pre-change
	// "revise" can be a low nit that would be an "approve" under the new rule.
	RoundDecisionRule string `json:"round_decision_rule,omitempty"`
	// RoundIndex is this round's ordinal within its correlation (resubmission
	// chains share one), so a consumer can follow a plan's trail across rounds.
	RoundIndex int `json:"round_index,omitempty"`
	// PlanSource names where input_state's plan text was recovered from:
	// "fix_plan" (the loop's own artifact) or "submission_payload" (a 097
	// submission's orchestration row — only recoverable inside the ~24h prune
	// window). Empty means no plan was recoverable; such rows are flagged
	// plan_missing.
	PlanSource string `json:"plan_source,omitempty"`
	Dissent    *bool  `json:"dissent,omitempty"`
	// Contested marks a round where seats actually SPLIT (some approved, some
	// did not). An uncontested round carries far less signal whichever way it
	// went — everyone agreeing is cheap.
	Contested bool `json:"contested,omitempty"`
	// ScopeApproveRisk marks a seat that has NEVER objected in the whole corpus.
	// Sampled, those seats are substantive — they write scope determinations
	// ("this change is outside my footprint") rather than merit judgements. Their
	// approve therefore does not mean what edit-quality's approve means, and
	// pooling the two would be wrong.
	ScopeApproveRisk bool `json:"scope_approve_risk,omitempty"`
}

type Provenance struct {
	AgentType     string `json:"agent_type,omitempty"`
	Model         string `json:"model,omitempty"`
	ModelResolved string `json:"model_resolved,omitempty"`
	MaxTokens     *int   `json:"max_tokens,omitempty"`
	InputTokens   *int   `json:"input_tokens,omitempty"`
	OutputTokens  *int   `json:"output_tokens,omitempty"`
	Success       *bool  `json:"success,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	PreFixCorpus  bool   `json:"pre_fix_corpus"`
}

// ── what extract.sql gives us ───────────────────────────────────────────────

type inRow struct {
	T            string          `json:"_t"`
	TrajectoryID string          `json:"trajectory_id"`
	RunID        string          `json:"run_id"`
	StepName     string          `json:"step_name"`
	CreatedAt    string          `json:"created_at"`
	InputState   string          `json:"input_state"`
	ReasoningRaw string          `json:"reasoning_raw"`
	Provenance   Provenance      `json:"provenance"`
	Trail        []trailStep     `json:"trail"`
	VerdictRaw   json.RawMessage `json:"verdict_raw"`
	Diagnosis    json.RawMessage `json:"diagnosis"`
	Iteration    int             `json:"iteration"`
	Body         string          `json:"body"`

	// council lane
	ReportID         string          `json:"report_id"`
	RoundIndex       int             `json:"round_index"`
	RoundDecision    string          `json:"round_decision"`
	Seat             string          `json:"seat"`
	Rationale        string          `json:"rationale"` // subplan rows only
	Summary          string          `json:"summary"`   // plan rows only
	SeatVerdict      string          `json:"seat_verdict"`
	Objections       json.RawMessage `json:"objections"`
	Notes            string          `json:"notes"`
	SeatVotes        int             `json:"seat_votes"`
	SeatNeverObjects bool            `json:"seat_never_objects"`

	// feed-triage lane
	ItemID            string          `json:"item_id"`
	Status            string          `json:"status"`
	RelevanceScore    *float64        `json:"relevance_score"`
	Credibility       string          `json:"credibility"`
	CredibilityReason string          `json:"credibility_reason"`
	SourceAttribution json.RawMessage `json:"source_attribution"`
	Topics            json.RawMessage `json:"topics"`
	SourceTitle       string          `json:"source_title"`
	SourceSummary     string          `json:"source_summary"`
}

// feedVerdict is one item's judgement inside a batched score_relevance response.
type feedVerdict struct {
	ID                string          `json:"id"`
	Score             *float64        `json:"score"`
	Credibility       string          `json:"credibility"`
	CredibilityReason string          `json:"credibility_reason"`
	Reason            string          `json:"reason"`
	Topics            json.RawMessage `json:"topics"`
	Flagged           *bool           `json:"flagged"`
	SourceAttribution json.RawMessage `json:"source_attribution"`
}

// trailStep mirrors pkg/diagnose.Step, which carries NO json tags — so the
// wire form uses Go field names and integer enums. Do not "tidy" these.
type trailStep struct {
	Iteration int    `json:"Iteration"`
	GuardStop string `json:"GuardStop"`
	Verdict   struct {
		Outcome        int    `json:"Outcome"` // 0=Unverifiable 1=Confirmed 2=Refuted
		NeededEvidence string `json:"NeededEvidence"`
		Citations      []struct {
			Tier int `json:"Tier"` // 0=static 1=state 2=runtime
		} `json:"Citations"`
	} `json:"Verdict"`
}

var outcomeName = map[int]string{0: "UNVERIFIABLE", 1: "CONFIRMED", 2: "REFUTED"}

// blindedMarkers name the fixloop documentation that must never be evaluated
// against, because it contains the benchmark's own answers.
var blindedMarkers = []string{"fixloop_eg_dartsonline", "RUBRIC_", "NOTES_running"}

// theFixLanded is the moment bugs_open/016's render fix hit the live
// fix-proposer row. Every repropose call BEFORE it reasoned against blank
// reviewer objections, so the corpus is bimodal across this instant and the two
// halves must not be pooled. Graded on the RUN start, never the step timestamp:
// a run that straddles the fix carries pre-fix config in its later steps.
var theFixLanded = time.Date(2026, 7, 18, 13, 15, 11, 0, time.UTC)

// councilSeverityGate is the instant the council's decision RULE changed on the
// live chassis: v1.0.1149, pod-verified 2026-07-22T13:56:14Z (symbol
// `severityGates` present, 2 hits). Before it, ANY seat objection at ANY severity
// gated a round to "revise" (commits 9e91999a4 + 872c830a8); after it, only a
// HIGH-severity objection gates — low/medium are advisory and no longer block. So
// `round_decision` means two different things across this instant and must NOT be
// pooled. The per-seat `dissent`/`contested` labels are computed from raw votes,
// independent of how the round decision was reached, so they are UNAFFECTED.
//
// Unlike theFixLanded (a run straddles many steps, so grade on the run start), a
// council_report's created_at IS its decision time — one action, not a
// trajectory — so we compare the row itself. The `≤` in the docs: the two commits
// landed 09:06Z/11:02Z on 2026-07-22 and an earlier roll that day may already
// have carried them, so the pod start is the LATEST instant it could have gone
// live; rows before it are conservatively old-rule, rows on/after it confirmed new.
var councilSeverityGate = time.Date(2026, 7, 22, 13, 56, 14, 0, time.UTC)

// councilRule names which decision rule produced a council round_decision, so a
// consumer never redoes the boundary arithmetic. Empty on an unparseable stamp.
func councilRule(ts string) string {
	t, err := parseTS(ts)
	if err != nil {
		return ""
	}
	if t.Before(councilSeverityGate) {
		return "any_objection_gates"
	}
	return "high_severity_gates"
}

func main() {
	labelsPath := flag.String("labels", "", "path to LABELS_benchmark.json (hand-curated gold grades)")
	outPath := flag.String("out", "", "output JSONL path (default stdout)")
	flag.Parse()

	labels, err := loadLabels(*labelsPath)
	if err != nil {
		fatal("labels: %v", err)
	}

	steps, trails, bundles, councils, plans, feedjudges, feeditems, err := readInput(os.Stdin)
	if err != nil {
		fatal("read: %v", err)
	}
	if len(steps) == 0 {
		fatal("no step rows on stdin — did extract.sql run? (psql -At -f -)")
	}

	records := build(steps, trails, bundles, labels)
	records = append(records, buildCouncil(councils, plans, steps, labels)...)
	records = append(records, buildFeed(feedjudges, feeditems)...)

	out := os.Stdout
	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			fatal("create %s: %v", *outPath, err)
		}
		defer f.Close()
		out = f
	}
	w := bufio.NewWriter(out)
	defer w.Flush()
	enc := json.NewEncoder(w)
	for _, r := range records {
		if err := enc.Encode(r); err != nil {
			fatal("encode: %v", err)
		}
	}
	report(records, len(trails), len(bundles))
}

func readInput(f *os.File) (steps []inRow, trails map[string]inRow, bundles map[string]string, councils []inRow,
	plans map[string][]inRow, feedjudges []inRow, feeditems map[string]inRow, err error) {
	trails = map[string]inRow{}
	bundles = map[string]string{}
	plans = map[string][]inRow{}
	feeditems = map[string]inRow{}
	sc := bufio.NewScanner(f)
	// Bundles run to tens of KB and prompts larger still.
	sc.Buffer(make([]byte, 0, 1<<20), 64<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || !strings.HasPrefix(line, "{") {
			continue // psql banners, blank separators between the three queries
		}
		var r inRow
		if e := json.Unmarshal([]byte(line), &r); e != nil {
			// A malformed line is a real signal (a body with an embedded
			// newline would split). Report rather than skip silently.
			fmt.Fprintf(os.Stderr, "WARN: unparseable line skipped (%d bytes): %v\n", len(line), e)
			continue
		}
		switch r.T {
		case "step":
			steps = append(steps, r)
		case "trail":
			trails[r.RunID] = r
		case "bundle":
			bundles[bundleKey(r.TrajectoryID, r.Iteration)] = r.Body
		case "council":
			councils = append(councils, r)
		case "plan", "subplan":
			plans[r.TrajectoryID] = append(plans[r.TrajectoryID], r)
		case "feedjudge":
			feedjudges = append(feedjudges, r)
		case "feeditem":
			feeditems[r.ItemID] = r
		}
	}
	return steps, trails, bundles, councils, plans, feedjudges, feeditems, sc.Err()
}

func bundleKey(traj string, iter int) string { return traj + "#" + fmt.Sprint(iter) }

func build(steps []inRow, trails map[string]inRow, bundles map[string]string, labels map[string]Labels) []Record {
	// Verdict calls within a run map onto trail entries in order — but only
	// when the counts agree (see Guard). Count them up front to know that.
	verdictSeen := map[string]int{}
	verdictCalls := map[string]int{}
	for _, s := range steps {
		if classify(s.StepName) == "diagnosis_verdict" {
			verdictCalls[s.RunID]++
		}
	}

	sort.SliceStable(steps, func(i, j int) bool {
		if steps[i].TrajectoryID != steps[j].TrajectoryID {
			return steps[i].TrajectoryID < steps[j].TrajectoryID
		}
		return steps[i].CreatedAt < steps[j].CreatedAt
	})

	var out []Record
	for _, s := range steps {
		task := classify(s.StepName)
		rec := Record{
			TrajectoryID: s.TrajectoryID,
			RunID:        s.RunID,
			Task:         task,
			StepName:     s.StepName,
			InputState:   s.InputState,
			ReasoningRaw: s.ReasoningRaw,
			Provenance:   s.Provenance,
		}

		// reasoning: structured when the model returned JSON, raw otherwise.
		if j := extractJSON(s.ReasoningRaw); j != nil {
			rec.Reasoning = j
			rec.ReasoningRaw = ""
		}
		rec.Decision = decisionOf(task, s.ReasoningRaw, rec.Reasoning)

		// input_complete — the bugs_open/016 filter, plus truncation.
		rec.InputComplete, rec.ExcludeReason = judgeInput(s)

		// provenance: which side of the 016 fix this run STARTED on.
		if t, ok := trails[s.RunID]; ok {
			rec.Provenance.PreFixCorpus = runStartedBeforeFix(t.CreatedAt)
		} else {
			rec.Provenance.PreFixCorpus = runStartedBeforeFix(s.CreatedAt)
		}

		// guard + step_index, verdict steps only.
		if task == "diagnosis_verdict" {
			idx := verdictSeen[s.RunID]
			verdictSeen[s.RunID]++
			rec.StepIndex = idx
			t, haveTrail := trails[s.RunID]
			raw := rec.Decision

			switch {
			case !haveTrail:
				rec.Guard = &Guard{RawDecision: raw, Alignment: "no trail for this run"}
			case len(t.Trail) != verdictCalls[s.RunID]:
				// Counts disagree — index pairing would be a guess. Say so.
				rec.Guard = &Guard{
					RawDecision: raw,
					Alignment: fmt.Sprintf("ambiguous: %d verdict calls vs %d trail entries",
						verdictCalls[s.RunID], len(t.Trail)),
				}
			case idx < len(t.Trail):
				ts := t.Trail[idx]
				if ts.Iteration > 0 {
					rec.StepIndex = ts.Iteration
				}
				coerced := outcomeName[ts.Verdict.Outcome]
				if raw == "" {
					raw = coerced
				}
				tripped := raw != coerced
				rec.Guard = &Guard{
					RawDecision:     raw,
					CoercedDecision: coerced,
					Tripped:         &tripped,
					Diagnostic:      ts.Verdict.NeededEvidence,
					StopReason:      ts.GuardStop,
					Alignment:       "by-index (counts agree)",
				}
				rec.Labels.SelfOutcome = coerced
				// The bundle is the real input_state the verdict saw; the
				// rendered prompt wraps it. Prefer the bundle when we have it.
				if b, ok := bundles[bundleKey(s.TrajectoryID, rec.StepIndex)]; ok && b != "" {
					rec.InputState = b
				}
			}
			if rec.Labels.SelfOutcome == "" {
				rec.Labels.SelfOutcome = rec.Decision
			}
		}

		// benchmark grade / terminal — hand-curated, keyed on an 8-char prefix.
		if l, ok := labels[shortID(s.TrajectoryID)]; ok {
			rec.Labels.BenchmarkGrade = l.BenchmarkGrade
			rec.Labels.SubsystemGrade = l.SubsystemGrade
			rec.Labels.Terminal = l.Terminal
		}

		out = append(out, rec)
	}
	return out
}

// buildCouncil turns each (report x seat) row into a record. The decision is
// the seat's own verdict; the round's decision and this seat's dissent from it
// are labels, derived rather than hand-curated — which is what makes this lane
// usable at 5.6x the diagnosis lane's volume without a labelling campaign.
// seatStepName maps a report's reviewer name to its llm_call_log step name.
// One alias exists in the live data; everything else is 'review_' + seat.
func seatStepName(seat string) string {
	if seat == "prior_art_librarian" {
		return "review_prior_art"
	}
	return "review_" + seat
}

func buildCouncil(rows []inRow, plans map[string][]inRow, steps []inRow, labels map[string]Labels) []Record {
	// Dissent must be measured against the seats in the SAME round, not against
	// the round's decision. The council returns "revise" if ANY seat objects, so
	// comparing a seat to the decision labels every approver in a revise round a
	// dissenter — which measured 520 of 836 and is obviously wrong: those seats
	// are the majority. Real dissent is going against your peers on one plan.
	//
	// And a round is a REPORT, not an orchestration (corrected 2026-07-24): the
	// reviser loops plan→review inside one run, so an orchestration holds up to
	// 12 report rows. The first cut keyed this tally by RunID and pooled seats
	// across those rounds — on the published 836-row corpus that inflated
	// dissent 170→201 and contested 799→822. Keyed by report id since.
	type tally struct{ approve, other int }
	roundKey := func(c inRow) string { return c.ReportID + "|" + c.RunID + "|" + c.CreatedAt }
	rounds := map[string]*tally{}
	for _, c := range rows {
		k := roundKey(c)
		t := rounds[k]
		if t == nil {
			t = &tally{}
			rounds[k] = t
		}
		if c.SeatVerdict == "approve" {
			t.approve++
		} else {
			t.other++
		}
	}

	// Per-seat model provenance joins the spine's review_* calls (already on
	// stdin) by ORCHESTRATION, not correlation: a 097 submission council runs
	// under its own orchestration correlation while its artifacts carry the
	// submission correlation, so the correlation key matches only 15% of seats
	// and the orchestration key 100% (both measured 2026-07-24). Within a key,
	// rounds and retries stack; the call a report actually consumed is the
	// newest SUCCESSFUL one at or before the report's own timestamp — a failed
	// final retry after an earlier success would otherwise mislabel a row that
	// has real content as call_failed.
	calls := map[string][]inRow{}
	for _, s := range steps {
		if strings.HasPrefix(s.StepName, "review_") && s.RunID != "" {
			k := s.RunID + "#" + s.StepName
			calls[k] = append(calls[k], s)
		}
	}
	for k := range calls {
		sort.SliceStable(calls[k], func(i, j int) bool { return calls[k][i].CreatedAt < calls[k][j].CreatedAt })
	}
	for traj := range plans {
		sort.SliceStable(plans[traj], func(i, j int) bool { return plans[traj][i].CreatedAt < plans[traj][j].CreatedAt })
	}

	// THE PLAN/PROVENANCE JOIN IS NOT IMPLEMENTED — scaffolding only, and this
	// comment is where its counters used to be.
	//
	// b82b3d8b4 ("v1.0.1188 prior to merge in main") added the whole frame for
	// joining plan text onto council records — the `PlanSource` field and its
	// doc comment, the "plan"/"subplan" branch in readInput, the `plans` map and
	// its sort above — plus three counters, `planJoined, planMissing,
	// provJoined`. It did not add the join itself, so the counters were declared
	// and never used and the package HAS NOT COMPILED SINCE 2026-07-28. Removing
	// the three declarations is the whole of the fix (2026-07-31); nothing else
	// referenced them, verified by grep.
	//
	// WHAT IS STILL MISSING, so whoever resumes does not have to re-derive it:
	// `PlanSource` is never SET and `plans` is never READ. The intent, from the
	// field's own doc comment, is that each record records where its plan text
	// was recovered from — "fix_plan" (the loop's own artifact) or
	// "submission_payload" (a 097 submission's orchestration row, recoverable
	// only inside the ~24h prune window) — with rows that recover neither
	// flagged plan_missing. The counters were to tally those outcomes by source.
	// The join key is the decision that was never made and is NOT guessed here:
	// `plans` is keyed by TrajectoryID while a council round is identified by
	// `roundKey` (ReportID|RunID|CreatedAt), and which plan row belongs to which
	// round — newest at or before the report, as the `calls` join above does, or
	// something else — is the author's call, not this fix's.
	out := make([]Record, 0, len(rows))
	for _, c := range rows {
		t := rounds[roundKey(c)]
		seatApproved := c.SeatVerdict == "approve"
		// The modal verdict among this round's seats. A tie is not dissent for
		// anyone — the round was genuinely split down the middle.
		var dissent bool
		if t.approve != t.other {
			majorityApproves := t.approve > t.other
			dissent = seatApproved != majorityApproves
		}
		contested := t.approve > 0 && t.other > 0

		rec := Record{
			TrajectoryID:  c.TrajectoryID,
			RunID:         c.RunID,
			Task:          "council_seat_review",
			StepName:      c.Seat,
			InputState:    "", // the plan under review; join via trajectory_id
			ReasoningRaw:  c.Notes,
			Decision:      strings.ToUpper(c.SeatVerdict),
			InputComplete: true,
			Labels: Labels{
				RoundDecision:     c.RoundDecision,
				RoundDecisionRule: councilRule(c.CreatedAt),
				Dissent:           &dissent,
				Contested:         contested,
				ScopeApproveRisk:  c.SeatNeverObjects && seatApproved,
			},
			Provenance: Provenance{
				AgentType:    "council",
				CreatedAt:    c.CreatedAt,
				PreFixCorpus: runStartedBeforeFix(c.CreatedAt),
			},
		}
		if len(c.Objections) > 0 && string(c.Objections) != "[]" {
			rec.Reasoning = c.Objections
		}
		if ok, why := commonExclusions(c.Notes+string(c.Objections), rec.Provenance); !ok {
			rec.InputComplete, rec.ExcludeReason = false, why
		}
		if c.Notes == "" && rec.Reasoning == nil {
			rec.InputComplete = false
			rec.ExcludeReason = "empty_review"
		}
		if l, ok := labels[shortID(c.TrajectoryID)]; ok {
			rec.Labels.BenchmarkGrade = l.BenchmarkGrade
			rec.Labels.SubsystemGrade = l.SubsystemGrade
			rec.Labels.Terminal = l.Terminal
		}
		out = append(out, rec)
	}
	return out
}

// buildFeed explodes each batched score_relevance response into one record per
// judged item, joined to that item's persisted outcome.
//
// This is the only lane where the judgement and its outcome sit on the SAME row,
// and the outcome is not a restatement of the judgement: `rejected` spans scores
// 0-90, and 494 items scoring >=60 were rejected anyway — with duplicate_of and
// expiry explaining NONE of them. Something downstream decides and we do not
// know what, so those rows are FLAGGED (ScoreOutcomeDivergence) rather than
// explained.
func buildFeed(judgeRows []inRow, items map[string]inRow) []Record {
	var out []Record
	unparseable := 0
	for _, j := range judgeRows {
		var verdicts []feedVerdict
		raw := extractJSON(j.ReasoningRaw)
		if raw == nil || json.Unmarshal(raw, &verdicts) != nil {
			// 23 successful calls carry malformed JSON that the
			// output_tokens>=max_tokens rule does not catch. Count, do not drop
			// silently — a silent skip here would understate the lane.
			unparseable++
			continue
		}
		for _, v := range verdicts {
			item, haveItem := items[v.ID]
			rec := Record{
				TrajectoryID:  j.TrajectoryID,
				RunID:         j.RunID,
				Task:          "feed_relevance_judgement",
				StepName:      "score_relevance",
				InputComplete: true,
				Provenance:    j.Provenance,
			}
			rec.Provenance.PreFixCorpus = runStartedBeforeFix(j.CreatedAt)

			// What the model was judging.
			if haveItem {
				rec.InputState = strings.TrimSpace(item.SourceTitle + "\n\n" + item.SourceSummary)
			}
			// Its reasoning, as emitted.
			reasoning, _ := json.Marshal(map[string]any{
				"reason":             v.Reason,
				"credibility":        v.Credibility,
				"credibility_reason": v.CredibilityReason,
				"topics":             v.Topics,
				"source_attribution": v.SourceAttribution,
				"flagged":            v.Flagged,
			})
			rec.Reasoning = reasoning
			if v.Score != nil {
				rec.Decision = fmt.Sprintf("score=%.0f", *v.Score)
			}

			if ok, why := commonExclusions(rec.InputState+string(reasoning), rec.Provenance); !ok {
				rec.InputComplete, rec.ExcludeReason = false, why
			}
			if haveItem {
				rec.Labels.Terminal = item.Status
				if v.Score != nil && *v.Score >= 60 && item.Status == "rejected" {
					rec.Labels.ScoreOutcomeDivergence = true
				}
			} else {
				// Judged, but the item row is gone or was never persisted.
				rec.InputComplete = false
				rec.ExcludeReason = "item_row_missing"
			}
			out = append(out, rec)
		}
	}
	if unparseable > 0 {
		fmt.Fprintf(os.Stderr,
			"WARN: %d score_relevance responses were unparseable JSON (truncated mid-array; "+
				"the output_tokens>=max_tokens rule does NOT catch these)\n", unparseable)
	}
	return out
}

func classify(step string) string {
	switch {
	case step == "verdict":
		return "diagnosis_verdict"
	case strings.HasPrefix(step, "review_"):
		return "council_review"
	case step == "propose" || step == "repropose" || step == "reframe":
		return step
	default:
		return step
	}
}

// commonExclusions holds the rules that apply to EVERY lane regardless of where
// the record came from. It exists because the lanes drifted: build() ran these
// via judgeInput while buildCouncil and buildFeed set InputComplete directly and
// skipped them, so 3 council rows citing the blinded fixloop docs would have
// reached an eval set. pattern-check's untouched-twin rule caught that drift on
// the commit that introduced it — a new lane must call this, not reimplement it.
func commonExclusions(text string, p Provenance) (bool, string) {
	for _, marker := range blindedMarkers {
		if strings.Contains(text, marker) {
			return false, "blinded_docs"
		}
	}
	if ok, why := judgeTruncation(p); !ok {
		return false, why
	}
	if p.Success != nil && !*p.Success {
		if strings.HasPrefix(p.ErrorMessage, "response truncated") {
			return false, "truncated"
		}
		return false, "call_failed"
	}
	return true, ""
}

// judgeTruncation decides what the usage counts say about a completion, through
// aiservice.ClassifyTruncation rather than a hand-rolled comparison (MDL-043 —
// and this file's own rule, one line above commonExclusions: a new lane must call
// the shared thing, not reimplement it).
//
// THE DEFECT THIS CLOSES (bugs_open/366). The old form was
// `p.OutputTokens != nil && p.MaxTokens != nil && *p.MaxTokens > 0 && *p.OutputTokens >= *p.MaxTokens`.
// The nil-guards, not the arithmetic, were the hole: when usage was never reported
// the check was SKIPPED, so "the provider told us nothing about this call" was
// silently handled as "this call finished normally". Those are different claims.
//
// ⚠ TruncationUnknown HAS TWO CAUSES AND ONLY ONE IS SUSPICIOUS — this is the
// judgement bugs_open/366 reserved, and it is taken here on measured evidence
// rather than by picking a side:
//
//   - a ceiling WAS recorded and the provider still reported no output usage.
//     Anomalous: the ceiling proves the call went through the configured path, so
//     usage should have come back. Excluded. [MEASURED 2026-09-03: ZERO such rows
//     in the corpus — this is a latent hazard being closed, not live damage, which
//     is the outcome 366 §4 said would be "a fine outcome to record".]
//   - NO ceiling was recorded at all. The question was never answerable and there
//     is no truncation signal either way. NOT excluded. [MEASURED 2026-09-03: 161
//     such rows, 8% of the corpus's 2,013 successful rows — every one a
//     score_relevance call from the Mar–May logging regime that did not record
//     max_tokens, averaging 3,092 output tokens with no empty responses.]
//
// Collapsing those two into "exclude everything Unknown" would have deleted 161
// substantial rows to guard a population of zero. That is why the ceiling is
// tested separately rather than leaning on ClassifyTruncation's verdict alone:
// the shared classifier answers "was it cut?", and it is right that it cannot
// distinguish these, because the distinction is about LOGGING provenance, not
// about the completion.
func judgeTruncation(p Provenance) (bool, string) {
	out, max := 0, 0
	if p.OutputTokens != nil {
		out = *p.OutputTokens
	}
	if p.MaxTokens != nil {
		max = *p.MaxTokens
	}
	switch aiservice.ClassifyTruncation(out, max) {
	case aiservice.TruncationAtCeiling:
		return false, "truncated"
	case aiservice.TruncationUnknown:
		if max > 0 {
			return false, "usage_unreported"
		}
		return true, ""
	}
	return true, ""
}

// judgeInput implements the two exclusion rules that matter. Rows are FLAGGED,
// not dropped.
func judgeInput(s inRow) (bool, string) {
	// bugs_open/016: a step whose prompt rendered <no value> reasoned without
	// seeing its inputs. 100% of repropose calls in the historical corpus.
	if strings.Contains(s.InputState, "<no value>") {
		return false, "no_value_injection"
	}
	// EVERYTHING ELSE IS THE SHARED RULE SET (bugs_open/366). This function used to
	// carry its own copy of the truncation guard, the blinded-marker scan and the
	// failure check — three rules, hand-maintained twice, in a file whose comment on
	// commonExclusions says "a new lane must call this, not reimplement it" and
	// records that the lanes had already drifted once. The copy here is why the
	// nil-check defect had to be fixed in two places to be fixed at all.
	//
	// The blinded-marker scan runs over InputState, which is what the old copy
	// scanned, so the same rows trip it. Ordering note, since it is the one visible
	// difference: blinded_docs is now tested before truncation, so a row that is BOTH
	// reports "blinded_docs" where it used to report "truncated". The exclusion
	// decision is identical — rows are flagged, not dropped — and only the reason
	// label moves. Pinned by TestJudgeInputDelegatesToCommonExclusions.
	return commonExclusions(s.InputState, s.Provenance)
}

func runStartedBeforeFix(ts string) bool {
	t, err := parseTS(ts)
	if err != nil {
		return false
	}
	return t.Before(theFixLanded)
}

func parseTS(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.999999Z07:00", "2006-01-02 15:04:05.999999-07"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparseable timestamp %q", s)
}

// extractJSON pulls the first balanced JSON object or array out of a completion,
// which often carries prose or a ```json fence around it.
func extractJSON(s string) json.RawMessage {
	if s == "" {
		return nil
	}
	start := strings.IndexAny(s, "{[")
	if start < 0 {
		return nil
	}
	open := s[start]
	close := byte('}')
	if open == '[' {
		close = ']'
	}
	depth, inStr, esc := 0, false, false
	for i := start; i < len(s); i++ {
		c := s[i]
		switch {
		case esc:
			esc = false
		case c == '\\' && inStr:
			esc = true
		case c == '"':
			inStr = !inStr
		case inStr:
		case c == open:
			depth++
		case c == close:
			depth--
			if depth == 0 {
				cand := s[start : i+1]
				if json.Valid([]byte(cand)) {
					return json.RawMessage(cand)
				}
				return nil
			}
		}
	}
	return nil
}

func decisionOf(task, raw string, j json.RawMessage) string {
	if j != nil {
		var m map[string]any
		if json.Unmarshal(j, &m) == nil {
			for _, k := range []string{"outcome", "verdict", "decision"} {
				if v, ok := m[k].(string); ok && v != "" {
					return strings.ToUpper(v)
				}
			}
		}
	}
	if task == "diagnosis_verdict" {
		for _, o := range []string{"CONFIRMED", "REFUTED", "UNVERIFIABLE"} {
			if strings.Contains(raw, o) {
				return o
			}
		}
	}
	return ""
}

func shortID(s string) string {
	if len(s) >= 8 {
		return s[:8]
	}
	return s
}

// loadLabels reads the hand-curated gold grades. Deliberately hand-maintained:
// the benchmark grades live in ~21 lines of a 141KB notes file in four
// incompatible prose shapes, and several uses of "FAILED" mean an API 529 rather
// than a grade — a parser would confidently mislabel them.
func loadLabels(path string) (map[string]Labels, error) {
	if path == "" {
		return map[string]Labels{}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Decode leniently: the file carries a "_README" array and per-entry "_note"
	// strings for the humans who maintain it. Underscore-prefixed keys are
	// documentation, not labels.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	out := make(map[string]Labels, len(raw))
	for k, v := range raw {
		if strings.HasPrefix(k, "_") {
			continue
		}
		var l Labels
		if err := json.Unmarshal(v, &l); err != nil {
			return nil, fmt.Errorf("label %q: %w", k, err)
		}
		out[shortID(k)] = l
	}
	return out, nil
}

// report prints the corpus census to stderr. It names what was EXCLUDED and why
// — a silent cap reads as "we covered everything".
func report(recs []Record, trails, bundles int) {
	byTask := map[string]int{}
	byModel := map[string]int{}
	byExclude := map[string]int{}
	byOutcome := map[string]int{}
	byGrade := map[string]int{}
	byTerminal := map[string]int{}
	guardTripped, guardUnaligned, graded, complete := 0, 0, 0, 0
	dissenting, scopeApprove, contested, divergent := 0, 0, 0, 0
	trajectories := map[string]bool{}

	for _, r := range recs {
		byTask[r.Task]++
		byModel[r.Provenance.Model]++
		trajectories[r.TrajectoryID] = true
		if r.InputComplete {
			complete++
		} else {
			byExclude[r.ExcludeReason]++
		}
		if r.Guard != nil {
			if r.Guard.Tripped != nil && *r.Guard.Tripped {
				guardTripped++
			}
			if r.Guard.CoercedDecision == "" {
				guardUnaligned++
			}
		}
		if r.Labels.BenchmarkGrade != "" {
			graded++
			byGrade[r.Labels.BenchmarkGrade]++
		}
		if r.Labels.Terminal != "" {
			byTerminal[r.Labels.Terminal]++
		}
		if r.Task == "council_seat_review" {
			if r.Labels.Dissent != nil && *r.Labels.Dissent {
				dissenting++
			}
			if r.Labels.Contested {
				contested++
			}
			if r.Labels.ScopeApproveRisk {
				scopeApprove++
			}
		}
		if r.Labels.ScoreOutcomeDivergence {
			divergent++
		}
		if r.Labels.SelfOutcome != "" {
			byOutcome[r.Labels.SelfOutcome]++
		}
	}

	e := func(f string, a ...any) { fmt.Fprintf(os.Stderr, f, a...) }
	e("\n── reasoningset census ──────────────────────────────\n")
	e("records            %d across %d trajectories\n", len(recs), len(trajectories))
	e("trails / bundles   %d / %d\n", trails, bundles)
	e("input_complete     %d  (excluded %d)\n", complete, len(recs)-complete)
	for _, k := range sortedKeys(byExclude) {
		e("  ! %-22s %d   ← flagged, NOT dropped\n", k, byExclude[k])
	}
	e("guard tripped      %d  (raw verdict != coerced verdict)\n", guardTripped)
	e("guard UNALIGNED    %d  (verdict-call count != trail length; not asserted)\n", guardUnaligned)
	e("benchmark-graded   %d\n", graded)
	e("\nby task\n")
	for _, k := range sortedKeys(byTask) {
		e("  %-22s %d\n", k, byTask[k])
	}
	e("\nby model (do NOT pool generations)\n")
	for _, k := range sortedKeys(byModel) {
		e("  %-22s %d\n", k, byModel[k])
	}
	if len(byOutcome) > 0 {
		e("\nself_outcome\n")
		for _, k := range sortedKeys(byOutcome) {
			e("  %-22s %d\n", k, byOutcome[k])
		}
	}
	if dissenting+scopeApprove > 0 {
		e("\ncouncil lane (labels derived, not hand-curated)\n")
		e("  %-22s %d\n", "seat dissents", dissenting)
		e("  %-22s %d   ← 'approve' may mean OUT OF SCOPE, not merit\n", "scope-approve risk", scopeApprove)
	}
	if divergent > 0 {
		e("\nfeed lane\n")
		e("  %-22s %d   ← scored >=60 and rejected anyway; the score did NOT decide\n",
			"score/outcome divergence", divergent)
	}
	if len(byGrade) > 0 {
		e("\nbenchmark_grade (gold, pre-registered rubric)\n")
		for _, k := range sortedKeys(byGrade) {
			e("  %-22s %d\n", k, byGrade[k])
		}
	}
	if len(byTerminal) > 0 {
		e("\nterminal\n")
		for _, k := range sortedKeys(byTerminal) {
			e("  %-22s %d\n", k, byTerminal[k])
		}
	}
	e("─────────────────────────────────────────────────────\n")
}

func sortedKeys(m map[string]int) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func fatal(f string, a ...any) {
	fmt.Fprintf(os.Stderr, "reasoningset: "+f+"\n", a...)
	os.Exit(1)
}
