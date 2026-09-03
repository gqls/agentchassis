// FILE: platform/orchestration/actions/repair_ordering_register_action.go
//
// ACTION: repair_ordering_register — the PRODUCER-SIDE half of the copy gate.
//
// WHAT IT GUARDS. `rewrite_negations` catches the writer's prose on its way to a
// page. This catches the ANALYST's prose on its way into `site_specs` — the
// ranked `lead_with[]` points the offer-analyser mints, which are about to
// become writer input under the owner's Decision C.
//
// ── WHY A SECOND GATE, WHEN THE COPY GATE ALREADY EXISTS ────────────────────
// Because the copy gate is downstream of a corpus that is 23% dirty at birth,
// and washing that corpus cannot win.
//
// `[MEASURED 2026-08-31]` migration 667 repaired 41 banned-register points
// across the estate at 10:34Z. By 16:25Z the producer had minted 75 fresh points
// and 18 of them (24%) carried the same constructions — 15 live across 8 sites,
// including rank 1 on `finetuning.uk` and `mortgagecalculator.co.uk`. Rank 1 is
// what a hero writer reads first. The wash moved the live corpus from 27% to
// 13.5% and the mint was refilling it within the hour.
//
// A previous session of this lane told the copy lane "the corpus must be clean
// on both axes before fleet-wide wiring". That condition is true at an instant
// and false by the next regeneration; it was retracted. THE GATE IS NOT A CORPUS
// STATE, IT IS A PRODUCER PROPERTY — which is this action.
//
// ── WHY IT REPAIRS RATHER THAN REFUSES (owner ruling, 2026-08-31) ────────────
// Three modes were costed. FAIL leaves the site's PREVIOUS ordering current, and
// that one is dirty too, so nothing gets cleaner and one phrasing tic costs an
// entire re-analysis. DROP — what the neighbouring `verify_cited_cardinals` step
// does for unsourced numbers — throws away about one benefit in four, some of
// them ranked first; there the removed thing is a FALSE CLAIM, here it is a real
// benefit wearing a tic, and the two do not deserve the same treatment. REPAIR,
// judged, keeps the benefit and costs one model call on roughly a quarter of
// runs (~16 runs/day fleet-wide as of 2026-08-31). The owner ruled repair.
//
// ── ⚠ THE JUDGE NEEDED A STRICTER FLOOR HERE THAN IT USES ON PAGE COPY ──────
// `AcceptNegationRewrite` rejects a rewrite shorter than 40% of the original
// ("gutted"). On THIS artefact that floor is measurably too low. The 667 wash's
// repairs averaged −28.7% on `differentiated: true` points — comfortably inside
// the 40% floor — and ten of the fifty-one still had to be excluded BY HAND
// because the truncation had removed the differentiating clause. The reason is
// structural: in an `X, not Y` construction the differentiation lives in the Y,
// so ruling 7's truncate-before-the-comparison systematically strips exactly the
// half that made the point worth ranking (HANDOFF_2026-08-26b §H1b).
//
// So a `differentiated: true` point carries an ADDITIONAL floor, defaulting to
// 60% — this lane's own measured rule: "a repair removing >=40% of a
// differentiated point has removed the differentiating clause, not a flourish".
// It is applied HERE and not in AcceptNegationRewrite deliberately: that helper
// is shared with the page copy gate, where the 40% floor is correct and where a
// silent tightening would change repair behaviour fleet-wide.
//
// ── ⚠ AND THE JUDGE IS BLIND TO THE WORD ARM, WHICH IS WHY THE RE-SCAN IS HERE ─
// AcceptNegationRewrite re-scans its candidate with ScanDefineByNegation and
// rejects "still_<shape>". It knows nothing about banned WORDS — that arm had no
// Go reader at all before datahelpers/registerwords.go. So a rewrite that
// removes "X, not Y" and reaches for "we say so plainly" passes every check the
// shared judge makes. This action re-scans with the FULL register instead, which
// is the only reason the two arms cannot displace into each other.
//
// ── IT NEVER DROPS A POINT AND NEVER FAILS THE STEP ON A VIOLATION ──────────
// An unrepairable point keeps its original text and is RECORDED. That is the
// "fail loud, never filter silently" commitment this lane made to the copy lane,
// honoured in the arm that matters: what makes the producer's error rate
// measurable is the record, not the refusal. A silent filter would make the rate
// unmeasurable, and that rate is the only evidence any of this works — the
// baseline it has to move is 23%.
//
// Config:
//   - object_field         (required) path to the object holding the items array
//   - items_key            (required) key within it holding the array
//   - text_key             (optional, default "point")            key holding the prose
//   - differentiated_key   (optional, default "differentiated")   bool driving the stricter floor
//   - record_key           (optional, default "register_repairs") where the record lands
//   - differentiated_floor (optional, default 60) percent of original length a repaired
//     differentiated point must retain
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// RepairOrderingRegisterInputSpec declares every step-config key this action
// reads under ConfigKeys, opting it into unknown-config-key detection. Same
// reasoning as VerifyCitedCardinalsInputSpec: these are settings resolved by
// ExtractNestedField, not inputs extracted by ExtractActionInputs, so they are
// NOT in Optional and do not inflate the RFC_022 optional-key budget, which
// counts accumulated authority rather than settings.
var RepairOrderingRegisterInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
	ConfigKeys: []string{
		"object_field", "items_key", "text_key",
		"differentiated_key", "record_key", "differentiated_floor",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("repair_ordering_register", RepairOrderingRegisterInputSpec)
}

// defaultDifferentiatedFloorPct is this lane's measured rule, as a percentage of
// the original point's length that a repair must retain when the point is marked
// differentiated. See the header for the measurement behind the number.
const defaultDifferentiatedFloorPct = 60

// registerTarget is one point that tripped the register and is being sent for
// repair.
type registerTarget struct {
	Index          int
	Text           string
	Rank           int
	Differentiated bool
	Hits           []datahelpers.RegisterViolation
	// ProtectFrom is the byte offset of the EARLIEST construction in the point.
	// Facts before it are the claim and must survive the rewrite; the contrasted
	// alternative after it is what the repair is allowed to drop. This is why
	// ScanBannedRegister sorts by offset.
	ProtectFrom int
}

// registerRepairRecord is one row of the durable record. It is written into the
// persisted artefact, so its field names are part of what a census reads.
type registerRepairRecord struct {
	Index          int      `json:"index"`
	Rank           int      `json:"rank,omitempty"`
	Differentiated bool     `json:"differentiated"`
	Shapes         []string `json:"violations"`
	Outcome        string   `json:"outcome"` // repaired | kept
	Reason         string   `json:"reason,omitempty"`
	From           string   `json:"from"`
	To             string   `json:"to,omitempty"`
}

// registerSummary is written into the persisted artefact beside the per-item
// record, and it exists because of a council objection (bug_historian, 4054f4d9
// round 1) that is a real presentation defect rather than a mechanism one:
//
//	"the gate's presence could read as 'this content is now guarded' when a
//	 meaningful fraction of violations will still ship dirty."
//
// That is exactly right, and it is this estate's own *a PASS from a BLIND check
// outlives the blindness* shape pointed at a gate that is not blind but is
// DELIBERATELY INCOMPLETE. Layers 2-4 fail closed, so a refused repair keeps the
// original violating text — by design, because the alternative is accepting a
// rewrite that guts the point. The number of items that still ship dirty is
// therefore an EXPECTED, RECURRING output of a working gate, not an error state.
//
// Deriving it (counting outcome='kept' across the record) is possible but is an
// inference a reader has to think to make. `StillViolating` states it, so a
// census, a dashboard or the next session reads it without reconstructing the
// rule. ⚠ It is written on the CLEAN path too, all zeros — same deep-merge
// reason as the record itself.
type registerSummary struct {
	Checked        int    `json:"checked"`
	Violations     int    `json:"violations"`
	Repaired       int    `json:"repaired"`
	StillViolating int    `json:"still_violating"`
	Register       string `json:"register"`
	RegisterVer    int    `json:"register_version"`
}

func newRegisterSummary(checked, violations, repaired int) registerSummary {
	return registerSummary{
		Checked: checked, Violations: violations, Repaired: repaired,
		StillViolating: violations - repaired,
		Register:       datahelpers.BannedRegisterPath,
		RegisterVer:    datahelpers.BannedRegisterVersion,
	}
}

type registerReplacement struct {
	Index int    `json:"index"`
	To    string `json:"to"`
}

// RepairOrderingRegisterAction is the handler.
func RepairOrderingRegisterAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "repair_ordering_register"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	objectField, _ := config["object_field"].(string)
	itemsKey, _ := config["items_key"].(string)
	if objectField == "" || itemsKey == "" {
		return nil, fmt.Errorf("repair_ordering_register requires object_field and items_key in step config")
	}
	textKey := stringOrDefault(config["text_key"], "point")
	diffKey := stringOrDefault(config["differentiated_key"], "differentiated")
	recordKey := stringOrDefault(config["record_key"], "register_repairs")
	floorPct := defaultDifferentiatedFloorPct
	if f, ok := config["differentiated_floor"].(float64); ok && f > 0 && f <= 100 {
		floorPct = int(f)
	}

	objRaw := datahelpers.ExtractNestedField(params.CollectedData, objectField)
	obj, ok := objRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("object_field %q did not resolve to an object (got %T)", objectField, objRaw)
	}
	itemsRaw, present := obj[itemsKey]
	if !present {
		return nil, fmt.Errorf("object at %q has no key %q", objectField, itemsKey)
	}
	items, ok := itemsRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("%s.%s is not an array (got %T)", objectField, itemsKey, itemsRaw)
	}

	// Work on a copy of each item so a rejected repair cannot leave a half-edited
	// map behind: the artefact this step returns is either the original text or
	// an accepted rewrite, never something in between.
	repaired := make([]interface{}, len(items))
	copy(repaired, items)

	var targets []registerTarget
	for i, raw := range items {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		text, _ := item[textKey].(string)
		if strings.TrimSpace(text) == "" {
			continue
		}
		hits := datahelpers.ScanBannedRegister(text)
		if len(hits) == 0 {
			continue
		}
		diff, _ := item[diffKey].(bool)
		rank := 0
		switch r := item["rank"].(type) {
		case float64:
			rank = int(r)
		case int:
			rank = r
		}
		targets = append(targets, registerTarget{
			Index: i, Text: text, Rank: rank, Differentiated: diff,
			Hits: hits, ProtectFrom: hits[0].At,
		})
	}

	// ⚠ THE CLEAN PATH MUST STILL WRITE record_key, AS AN EMPTY ARRAY.
	//
	// write_site_spec DEEP-MERGES the returned object over the stored one, so a
	// key this document omits keeps the PREVIOUS run's value and reads as
	// current. A run that repaired two points, followed by a clean run, would
	// otherwise leave the earlier run's repair record standing next to an
	// ordering that no longer needed one — an audit record accusing a clean
	// artefact, for ever. This is the same trap verify_cited_cardinals documents
	// for dropped_unsourced (bugs_open/327), and the reason `len(register_repairs)
	// > 0` is a sound census predicate for a reader who cannot know how many runs
	// came before.
	if len(targets) == 0 {
		logger.Info("repair_ordering_register: every point is clean against the banned register",
			zap.Int("points_checked", len(items)),
			zap.String("register", datahelpers.BannedRegisterPath),
			zap.Int("register_version", datahelpers.BannedRegisterVersion))
		return map[string]interface{}{
			"clean": true, "checked": len(items), "violations": 0,
			"repaired": 0, "unrepaired": 0,
			"object": withKey(
				withKey(obj, recordKey, []registerRepairRecord{}),
				recordKey+"_summary", newRegisterSummary(len(items), 0, 0)),
		}, nil
	}

	// LOUD, before any repair is attempted: this line is the producer's error
	// rate as measured at the mint, and it is emitted whether or not the repair
	// then succeeds. A census that only counted successful repairs would report
	// the gate working exactly when it was failing.
	logger.Warn("repair_ordering_register: the producer minted banned-register points",
		zap.Int("points_checked", len(items)),
		zap.Int("violating_points", len(targets)),
		zap.String("detail", describeTargets(targets)))

	records, repairedCount, callErr := runRegisterRepair(ctx, params, config, targets, repaired, textKey, floorPct)

	unrepaired := len(targets) - repairedCount
	if callErr != "" {
		logger.Warn("repair_ordering_register: the repair call did not complete — every point keeps its original text",
			zap.String("error", callErr), zap.Int("unrepaired", unrepaired))
	}
	// ⚠ `still_violating` is the honest headline, not `repaired`. A gate that
	// repaired 3 of 8 has left 5 dirty points in a live artefact, and a log line
	// leading with the 3 is how "the gate is working" outlives the 5.
	logger.Warn("repair_ordering_register: register repair finished — points STILL VIOLATING ship as written",
		zap.Int("violating_points", len(targets)),
		zap.Int("repaired", repairedCount),
		zap.Int("still_violating", unrepaired))

	out := map[string]interface{}{
		"clean": false, "checked": len(items), "violations": len(targets),
		"repaired": repairedCount, "unrepaired": unrepaired,
		"object": withKey(
			withKey(withKey(obj, itemsKey, repaired), recordKey, records),
			recordKey+"_summary", newRegisterSummary(len(items), len(targets), repairedCount)),
	}
	if callErr != "" {
		out["repair_error"] = callErr
	}
	return out, nil
}

// withKey returns a shallow copy of obj with one key set. The stored object is
// never mutated in place: it belongs to the previous step's output, and this
// estate has been bitten by an in-place edit that the coordinator's fresh-state
// reload then discarded (rewrite_negations_action.go's `result` note).
func withKey(obj map[string]interface{}, key string, val interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(obj)+1)
	for k, v := range obj {
		out[k] = v
	}
	out[key] = val
	return out
}

func describeTargets(targets []registerTarget) string {
	parts := make([]string, 0, len(targets))
	for _, t := range targets {
		parts = append(parts, fmt.Sprintf("rank %d: %s", t.Rank, datahelpers.DescribeRegisterViolations(t.Hits)))
	}
	return strings.Join(parts, " | ")
}

func violationNames(hits []datahelpers.RegisterViolation) []string {
	out := make([]string, 0, len(hits))
	for _, h := range hits {
		out = append(out, h.Kind+":"+h.Name)
	}
	return out
}

// registerRepairPrompt asks for one restatement per point.
//
// ⚠ IT CARRIES NO RULE TEXT AND NO EXAMPLES OF THE BANNED SHAPES, and that is
// load-bearing rather than terse. Prompt text demonstrating a form teaches it
// and scores as it — this estate's `prompt-text-poisons-its-own-detector`
// lesson, and the register's own usage rule says the same: "structured input
// ONLY. Never paste these patterns or their examples as prose into a prompt."
// So the model is shown its own sentence and told which fragment to remove, and
// never a catalogue of the constructions to avoid.
func registerRepairPrompt(targets []registerTarget) string {
	var b strings.Builder
	b.WriteString(`Each numbered item below is one benefit statement, followed by the exact fragment that must not appear in it.

Restate each one so the fragment is gone.

Rules:
- Keep the specific, distinguishing content. If the statement distinguishes this business from others, the distinguishing detail must survive the restatement — that detail is the whole value of the statement.
- Keep every figure, name and link exactly as written.
- Do not add any claim, superlative or name that is not already in the original.
- Say what the thing IS. Do not replace the fragment with a different comparison.
- If you cannot restate an item without losing its meaning, leave it out of your answer. A shorter list is a fine answer.

Answer with JSON only:
{"replacements":[{"index":<the item number>,"to":"<the restated statement>"}]}

Items:
`)
	for _, t := range targets {
		fmt.Fprintf(&b, "\n%d. statement: %s\n   remove this fragment: %q\n", t.Index, t.Text, t.Hits[0].Matched)
		if t.Differentiated {
			fmt.Fprintf(&b, "   (this statement is marked as distinguishing — the detail that distinguishes must survive)\n")
		}
	}
	return b.String()
}

// runRegisterRepair makes the single call and judges each replacement. It
// returns the record for EVERY target — repaired or kept — plus the count
// repaired, plus a call-level error string if the round produced nothing.
func runRegisterRepair(
	ctx context.Context, params ActionParams, config map[string]interface{},
	targets []registerTarget, items []interface{}, textKey string, floorPct int,
) ([]registerRepairRecord, int, string) {

	// keptRecord builds the record for a target that keeps its original text.
	keptRecord := func(t registerTarget, reason string, to string) registerRepairRecord {
		r := registerRepairRecord{
			Index: t.Index, Rank: t.Rank, Differentiated: t.Differentiated,
			Shapes: violationNames(t.Hits), Outcome: "kept", Reason: reason,
			From: datahelpers.TruncateString(t.Text, 240),
		}
		if to != "" {
			r.To = datahelpers.TruncateString(to, 240)
		}
		return r
	}
	allKept := func(reason string) []registerRepairRecord {
		out := make([]registerRepairRecord, 0, len(targets))
		for _, t := range targets {
			out = append(out, keptRecord(t, reason, ""))
		}
		return out
	}

	agentConfig, _ := params.CollectedData["agent_config"].(map[string]interface{})
	if agentConfig == nil && params.AgentType != "" && params.DB != nil {
		if def, err := loadAgentDefinitionForAction(ctx, params.DB, params.AgentType); err == nil {
			agentConfig = def.DefaultConfig
		}
	}
	stepName := ""
	if params.ExecutionContext != nil {
		stepName = params.ExecutionContext.StepName
	}
	aiServiceConfig, _ := resolveAIServiceConfig(agentConfig, config, stepName, params.Logger)
	if len(aiServiceConfig) == 0 {
		return allKept("no ai_service configuration resolvable"), 0, "no ai_service configuration resolvable"
	}
	client, err := createAIClient(ctx, aiServiceConfig)
	if err != nil {
		return allKept("ai client: " + err.Error()), 0, "ai client: " + err.Error()
	}

	prompt := registerRepairPrompt(targets)
	// The budget is resolved by the ONE resolver this package has
	// (llm_options.go), never here. See rewrite_negations_action.go for the full
	// account; the short version is that the hardcoded `2000` this replaced was
	// numerically EQUAL to `offer-analyser`'s configured 2000, so
	// `llm_call_log.max_tokens` read the same whether the configuration was
	// honoured or dropped, and no query could tell the two apart.
	options := llmOptionsFromConfig(config, aiServiceConfig, params.Logger, "repair_ordering_register")
	if t, ok := aiServiceConfig["temperature"].(float64); ok {
		options["temperature"] = t
	}

	start := time.Now()
	raw, callErr := client.GenerateText(ctx, prompt, options)
	latency := int(time.Since(start).Milliseconds())

	truncated := false
	if _, isTrunc := aiservice.IsTruncated(callErr); isTrunc {
		truncated = true
	}

	// One call, one forensic row. The marker is present on SUCCESSFUL repairs
	// too, so a census must filter on success=false and never on a non-empty
	// error_message (the bugs_open/119 precedent, which has already misled one
	// census).
	logMsg := "RETRY (producer register gate: offer_ordering lead_with)"
	if callErr != nil {
		logMsg = "RETRY (producer register gate) FAILED: " + callErr.Error()
	}
	inTok, _ := options["__usage_input_tokens"].(int)
	outTok, _ := options["__usage_output_tokens"].(int)
	// The ceiling of record is the one the provider APPLIED, not the one we
	// asked for — see rewrite_negations_action.go for why this is read from
	// __sent_max_tokens with our own request only as a fallback.
	sentMax, _ := options["__sent_max_tokens"].(int)
	if sentMax <= 0 {
		sentMax, _ = options["max_tokens"].(int)
	}
	model, _ := aiServiceConfig["model"].(string)
	provider, _ := aiServiceConfig["provider"].(string)
	orchID, corrID, agentID := "", "", ""
	if params.ExecutionContext != nil {
		orchID, corrID = params.ExecutionContext.OrchestrationID, params.ExecutionContext.CorrelationID
	}
	if params.Headers != nil {
		agentID = params.Headers["agent_id"]
	}
	LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
		AgentType: params.AgentType, AgentID: agentID, StepName: stepName,
		OrchestrationID: orchID, CorrelationID: corrID,
		Model: model, Provider: provider,
		PromptTemplate: "repair_ordering_register", PromptRendered: prompt, ResponseText: raw,
		InputTokens: inTok, OutputTokens: outTok, LatencyMs: latency,
		Success: callErr == nil, ErrorMessage: logMsg,
		MaxTokens: sentMax, Options: options,
	})

	if callErr != nil {
		return allKept("repair call failed: " + callErr.Error()), 0, callErr.Error()
	}
	// A truncated answer is not a smaller answer, it is a cut one, and a `to`
	// string cut mid-sentence still looks like prose. Three states are kept
	// distinct: at-the-ceiling, below it, and UNKNOWN — a provider that returns
	// a cut answer as a 200 may report no usage at all, in which case the load
	// bearing protection is the parse, since cut JSON does not parse.
	if usage := aiservice.ClassifyTruncation(outTok, sentMax); usage.Truncated() {
		truncated = true
	} else if usage == aiservice.TruncationUnknown {
		params.Logger.Info("repair_ordering_register: the provider reported no output-token usage, so the ceiling check could not run — relying on the parse",
			zap.String("step_name", stepName))
	}
	if truncated {
		return allKept("repair answer truncated at the output ceiling"), 0, "repair answer truncated at the output ceiling"
	}

	parsed, _, perr := ParseLLMJSONWithProvenance(stripMarkdownFromResponse(raw))
	if perr != nil {
		return allKept("unparseable repair answer"), 0, "unparseable repair answer"
	}

	byIndex := make(map[int]registerTarget, len(targets))
	for _, t := range targets {
		byIndex[t.Index] = t
	}

	records := make([]registerRepairRecord, 0, len(targets))
	repairedCount := 0
	// ⚠ THIS LOOP RUNS OVER WHAT THE MODEL RETURNED, NOT OVER WHAT WE ASKED
	// ABOUT. A target the answer never mentions is visited by no branch here and
	// would land in NEITHER list — measured at 31% unreconciled on the page gate
	// before it was fixed, which corrupted the very log its displacement defence
	// depends on. `answered` plus the reconciliation sweep below is what closes
	// it: every target gets exactly one record.
	answered := make(map[int]bool, len(targets))
	for _, r := range decodeRegisterReplacements(parsed) {
		t, found := byIndex[r.Index]
		if !found || answered[r.Index] {
			continue
		}
		answered[r.Index] = true

		to := strings.TrimSpace(r.To)
		if ok, why := judgeRegisterRewrite(t, to, floorPct); !ok {
			records = append(records, keptRecord(t, why, to))
			continue
		}

		item, ok := items[t.Index].(map[string]interface{})
		if !ok {
			records = append(records, keptRecord(t, "item_not_an_object", to))
			continue
		}
		patched := make(map[string]interface{}, len(item))
		for k, v := range item {
			patched[k] = v
		}
		patched[textKey] = to
		items[t.Index] = patched

		records = append(records, registerRepairRecord{
			Index: t.Index, Rank: t.Rank, Differentiated: t.Differentiated,
			Shapes: violationNames(t.Hits), Outcome: "repaired",
			From: datahelpers.TruncateString(t.Text, 240),
			To:   datahelpers.TruncateString(to, 240),
		})
		repairedCount++
	}

	// Reconciliation: every target the model never answered gets its own record,
	// so len(records) == len(targets) always and the rate stays countable.
	for _, t := range targets {
		if !answered[t.Index] {
			records = append(records, keptRecord(t, "not_addressed_by_the_model", ""))
		}
	}
	return records, repairedCount, ""
}

// judgeRegisterRewrite decides whether a proposed restatement may stand in for
// the original point. Pure, and separate from the call so it is testable without
// a model: this is the half that decides what reaches the artefact.
//
// It fails CLOSED in three layers, in this order:
//
//  1. AcceptNegationRewrite — the shared structural judge. Rejects an empty,
//     unchanged, gutted or ballooned candidate, one that still carries a shape,
//     one that DISPLACED into a neighbouring construction, one that dropped a
//     figure or link protected by ProtectFrom, and one that invented a figure,
//     superlative or name. Shared with the page copy gate deliberately: two
//     judges would drift.
//
//  2. The FULL register re-scan. ⚠ Layer 1 is blind to the word arm — it
//     re-scans with ScanDefineByNegation only — so a rewrite that drops
//     "X, not Y" and reaches for "we say so plainly" passes it. Without this
//     layer the two arms displace into each other and the gate teaches the
//     word.
//
//  3. The differentiated floor. AcceptNegationRewrite's "gutted" floor is 40%
//     of the original length. `[MEASURED 2026-08-31]` on this artefact the 667
//     wash's repairs averaged −28.7% on differentiated points — inside that
//     floor — and ten of fifty-one still had the distinguishing clause removed,
//     because in an `X, not Y` construction the differentiation lives in the Y.
//     So a differentiated point must retain floorPct (default 60%). Applied here
//     rather than in the shared helper, where 40% is correct for page copy.
func judgeRegisterRewrite(t registerTarget, to string, floorPct int) (bool, string) {
	if ok, why := datahelpers.AcceptNegationRewrite(t.Text, to, t.ProtectFrom); !ok {
		return false, why
	}
	if hits := datahelpers.ScanBannedRegister(to); len(hits) > 0 {
		return false, "still_" + hits[0].Kind + "_" + hits[0].Name
	}
	if t.Differentiated && len(to)*100 < len(t.Text)*floorPct {
		return false, fmt.Sprintf("differentiation_stripped (kept %d%% of a differentiated point, floor %d%%)",
			len(to)*100/max1(len(t.Text)), floorPct)
	}
	// LAYER 4: a differentiated point may not be repaired by TRUNCATION.
	//
	// ⚠ FOUND BY MUTATION, and it is why layer 3 alone is not enough. The
	// motivating case — 667's repair of leopardessconsulting rank 2, "…in days,
	// not months." -> "…in days." — retains **84%** of the original, so it sails
	// past both the shared 40% floor AND the 60% one above. The differentiation
	// was lost in TWELVE BYTES. A length rule cannot see that, and writing this
	// test against the real repair is what exposed it; the first version of this
	// judge passed the motivating case and would have shipped believing it caught
	// it.
	//
	// The rule is exact rather than heuristic. On a differentiated point the
	// violating construction is a COMPARISON, and the differentiation is the
	// thing compared against — the Y in "X, not Y". Ruling 7's repair is
	// "truncate before the comparison", which on such a point removes the Y by
	// construction, EVERY time, whatever the length arithmetic says. So a
	// candidate that is merely a PREFIX of the original has not restated the
	// point, it has cut it, and for a differentiated point that is always the
	// wrong repair. It must express the distinction positively instead.
	//
	// Undifferentiated points are deliberately exempt: there truncation is the
	// sanctioned repair and it demonstrably works — "…rather than standardising
	// on one, because…" -> "…, because…", with nothing lost.
	if t.Differentiated && isTruncationOf(t.Text, to) {
		return false, "truncation_only (a differentiated point loses its distinction when the comparison is cut; restate it positively)"
	}
	return true, ""
}

// isTruncationOf reports whether `to` is the opening of `from` with the tail
// dropped — the shape ruling 7's truncate-before-the-comparison produces.
// Trailing punctuation and space are ignored on the candidate, because a
// truncation necessarily re-terminates the sentence it cut.
func isTruncationOf(from, to string) bool {
	trimmed := strings.TrimRight(strings.TrimSpace(to), ".,;:!? \t—-")
	if trimmed == "" {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(from), trimmed)
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

// decodeRegisterReplacements accepts the envelope and a bare array, the two
// shapes a model actually returns.
func decodeRegisterReplacements(parsed interface{}) []registerReplacement {
	b, err := json.Marshal(parsed)
	if err != nil {
		return nil
	}
	var envelope struct {
		Replacements []registerReplacement `json:"replacements"`
	}
	if err := json.Unmarshal(b, &envelope); err == nil && len(envelope.Replacements) > 0 {
		return envelope.Replacements
	}
	var bare []registerReplacement
	if err := json.Unmarshal(b, &bare); err == nil {
		return bare
	}
	return nil
}
