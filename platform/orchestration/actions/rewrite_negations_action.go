// FILE: platform/orchestration/actions/rewrite_negations_action.go
//
// ACTION: rewrite_negations — the repair half of the copy gate (bugs_open/305).
//
// WHAT IT DOES. Between a writer generating a section and that section being
// rendered, it finds define-by-negation constructions in the generated content,
// leaves alone the ones the brief supplied or the regulator requires, and asks
// the model ONCE to rewrite the remaining sentences directly. Each proposed
// rewrite is accepted or rejected on its own merits; the accepted ones are
// spliced into the content map in place. It NEVER fails the step: style is
// softer than truth, and a page lost to a style check is a worse outcome than
// the tell it was chasing.
//
// ── WHY IT IS AN ACTION OF ITS OWN, NOT A KEY ON execute_llm_prompt ──────────
// That action has 66 carriers and no ActionInputSpec, so the RFC_022 optional-key
// budget cannot see keys added to it — a new field there is invisible to the one
// mechanism that counts accumulated authority on a shared seam. And a
// prose-quality rewrite is writer behaviour, not transport behaviour, in an
// action that also serves council seats, classifiers and researchers.
//
// ── WHY IT ASKS FOR SENTENCES, NOT A NEW SECTION ─────────────────────────────
// Measured on the live writer (2026-08-19): mean call 11,009 input / 2,126 output
// tokens, 0 cached prompts. A whole-section re-ask is ~5x the cost of a
// replacement patch, and it can truncate, drop a key, or change a field's type —
// each of which loses the section at RenderComponentAction's required-field
// refusal. Splicing sentences into the existing map cannot do any of that: the
// key set and the types are preserved by construction.
//
// ── WHY EVERY REWRITE IS JUDGED, AND HOW ─────────────────────────────────────
// The obvious design — ask again, keep whichever answer scores lower — ADOPTS
// DISPLACEMENT. In the same corpus "instead of" is present in 5.9% of sections,
// "isn't just/a" 6.4%, "more than just" 10.8%: a rewrite to "X instead of Y"
// scores zero on the five shapes and wins the comparison while being the same
// instinct. This estate has the finding from the other direction too — "a
// prohibition displaces a problem rather than solving it" (copy_quality_two_stage,
// 2026-08-12, where banning an opening moved the fault to the end of the
// sentence). So AcceptNegationRewrite fails CLOSED and every rejection is
// recorded with its reason. That log is the instrument: it is how we find out
// whether the repair is fixing the copy or teaching the model a new tic.
//
// ── WHAT IT DELIBERATELY LEAVES ALONE ────────────────────────────────────────
// A sentence the site's own brief supplied, and a regulatory or capability
// negation. The house voice's first sentence is "A site's own voice
// specification outranks these rules wherever the two disagree", and rewriting
// "these figures are an estimate, not financial advice" would be a compliance
// harm dressed as a style fix. Both are COUNTED and reported, so the brief that
// causes them is visible; fixing them means fixing the brief, which is the
// owning site lane's decision.
//
// ── THE BUDGET IS PER PAGE, AND THAT IS WHY IT IS IN CollectedData ───────────
// The standard is the house voice's own: "a matched contrasting pair is earned
// once or twice per page at most". A per-section threshold cannot express it —
// six sections carrying one construction each is six on the page, and every one
// passes. Loop iterations share one workflow state, so the running count lives
// at CollectedData["__copy_gate"], and the first two non-exempt hits on a page
// are left alone. A hit in a headline-class field is repaired regardless of
// budget: it is the first thing a reader meets, and it is what the owner quoted.
//
// ⚠ MEASURED 2026-08-21: the counter does NOT persist across iterations, so the
// budget IS per section — headline hits are still repaired regardless, which is
// the part that matters. `saveStepResultWithRetry` (coordinator.go) reloads a
// fresh state and copies only the current step's own stepName/output_field, so a
// bare CollectedData key does not survive. To make the budget truly per-page,
// carry the count in this step's OUTPUT and read `copy_gate_<N-1>`.
//
// ── AND THE SAME MECHANISM ATE THE REPAIR ITSELF, WHICH IS WHY `result` EXISTS ─
//
// This action used to rely SOLELY on mutating the writer's content map in place,
// on the reasoning that the map in CollectedData is the one the renderer reads.
// It is — in memory. It is not what survives to the next step: the same
// fresh-state copy that drops `__copy_gate` also drops an in-place edit to the
// PREVIOUS step's output, because only the CURRENT step's keys are copied
// forward. Measured at the artefact on 2026-08-21
// (remortgagecalculator.uk/mortgage-lenders): the gate reported
// `status: repaired, hits_after: 0`, and the stored `content_data.subheadline`
// was byte-identical to the pre-repair `generated_content.result.subheadline` —
// the renderer never saw the patch. An honest marker over a page that never
// changed: the silent no-op this design was supposed to be immune to.
//
// So the patched map is ALSO returned as this step's own `result`, which is the
// mechanism that demonstrably persists (`copy_gate_0`, `copy_gate_1` … all reach
// the durable row), and `render_section.content_from` points at `copy_gate.result`
// instead of `generated_content.result`. The in-place mutation is kept as well:
// it costs nothing and it keeps `hits_after` honest about the object this action
// actually holds.

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// copyGateMarkerKey is where the per-PAGE state lives in collected data. The
// "__" prefix keeps it out of the loop's output-propagation sweep, which
// deliberately skips internal keys.
const copyGateMarkerKey = "__copy_gate"

// defaultBriefFields are the paths whose text counts as SUPPLIED BY THE BRIEF.
//
// This list is the exemption's whole surface and it is deliberately short. It is
// NOT the rendered prompt: the literal string "rather than" appears in every
// rendered writer prompt (the house voice uses it six times, and STRICT RULE 19
// uses it), so a prompt-wide exemption silently exempts the broadest arm of the
// family — 43% of sections. These are the fields a site's own brief actually
// hands over, plus the two paths that carry copy the writer was TOLD to preserve.
var defaultBriefFields = []string{
	"site_specs.specs.content_direction.formatted",
	"site_specs.specs.identity.key_differentiators",
	"site_specs.specs.identity.target_audience",
	"render_context.tagline",
	"current_section.existing_content_html",
	"current_section.component.content_brief",
	"rewrite_guidance",
}

func init() {
	datahelpers.RegisterActionInputSpec("rewrite_negations", datahelpers.ActionInputSpec{
		Optional: []string{"content_from"},
		ConfigKeys: []string{
			"content_from", // where the writer's content map is (loop-suffixed by the engine)
			"brief_fields", // paths whose text counts as brief-supplied (exemption)
			"page_budget",  // non-exempt hits allowed per PAGE before repair starts
			"ai_service",   // model overlay for the one repair call
		},
		CheckConfig: true,
	})
}

// RewriteNegationsAction is the step handler. It returns a marker describing what
// it did and never an error, except when the step cannot be run at all.
func RewriteNegationsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	config := params.StepConfig.Config

	contentFrom := "generated_content.result"
	if cf, ok := config["content_from"].(string); ok && cf != "" {
		contentFrom = cf
	}
	content := extractContentWithFallbacks(params.CollectedData, contentFrom, params.Logger)
	if len(content) == 0 {
		params.Logger.Info("rewrite_negations: no content map at the configured path — nothing to do",
			zap.String("content_from", contentFrom))
		return map[string]interface{}{"status": "no_content", "content_from": contentFrom}, nil
	}

	supplied := collectBriefText(params, config)
	plan := planNegationRepairs(content, supplied, copyGatePageBudget(config), pageHitsSoFar(params), mildHitsSoFar(params))
	recordPageHits(params, plan.pageHits)
	recordMildHits(params, plan.mildHits)

	marker := map[string]interface{}{
		// The content this step hands ON. ALWAYS set when a content map was
		// found, patched or not, because render_section now reads it — a step
		// that returned it only when it changed something would drop the section
		// on every clean page.
		"result":         content,
		"hits_before":    plan.total,
		"exempt":         plan.exemptCount,
		"exempt_reasons": plan.exemptReasons,
		"within_budget":  plan.withinBudget,
		"targets":        len(plan.targets),
		"page_hits":      plan.pageHits,
		"mild_hits":      plan.mildHits,
	}
	if len(plan.targets) == 0 {
		marker["status"] = "clean"
		stampCopyGate(params, marker)
		return marker, nil
	}

	rewritten, rejected, callErr := runNegationRepair(ctx, params, config, plan)
	marker["rewritten"] = rewritten
	marker["rejected"] = rejected
	marker["hits_after"] = countNegationHits(content)
	marker["status"] = "repaired"
	if callErr != "" {
		// NOT the same thing as "nothing needed rewriting", and the distinction is
		// deliberate (council round 4, bug_historian): this action never fails the
		// STEP for a style outcome, but an infrastructure failure — no AI client,
		// a provider error, an answer that would not parse — is a different event
		// and must not read as a quiet pass. It is stamped with its own status and
		// logged at Error, so a census can find the runs where the gate was
		// PRESENT and BLIND rather than present and satisfied.
		marker["status"] = "repair_unavailable"
		marker["error"] = callErr
		params.Logger.Error("rewrite_negations: the repair could not run — the copy stands as written, and this is NOT a clean result",
			zap.String("reason", callErr),
			zap.Int("targets", len(plan.targets)))
	}
	stampCopyGate(params, marker)

	params.Logger.Info("rewrite_negations: done",
		zap.Int("hits_before", plan.total), zap.Int("targets", len(plan.targets)),
		zap.Int("rewritten", len(rewritten)), zap.Int("rejected", len(rejected)),
		zap.Any("hits_after", marker["hits_after"]))
	return marker, nil
}

// ── planning ────────────────────────────────────────────────────────────────

type negationTarget struct {
	Field    string
	Sentence string
	MatchAt  int // offset of the construction within Sentence
	Shape    string
	Headline bool
	set      func(string)
	text     string // the whole field value, for the splice
}

type negationPlan struct {
	total         int
	exemptCount   int
	exemptReasons map[string]int
	withinBudget  int
	pageHits      int
	mildHits      int // mild-shape hits so far this PAGE — what the budget counts
	targets       []negationTarget
}

// planNegationRepairs decides, per hit, whether it is exempt, allowed by the page
// budget, or a target. Order matters and is fixed: exemptions first (they are
// never counted against the budget, because a brief-supplied phrase is not the
// writer's doing), then headline hits (always repaired), then the budget.
func planNegationRepairs(content map[string]interface{}, supplied []string, budget, alreadyUsed, alreadyMild int) negationPlan {
	plan := negationPlan{exemptReasons: map[string]int{}, pageHits: alreadyUsed, mildHits: alreadyMild}
	used := alreadyMild
	seen := map[string]bool{} // one repair per (field, sentence), whatever shapes it carries

	for _, f := range datahelpers.WalkContentStrings(content) {
		field := f
		for _, h := range datahelpers.ScanDefineByNegation(field.Text) {
			plan.total++
			if ok, why := datahelpers.NegationExempt(h, supplied); ok {
				plan.exemptCount++
				plan.exemptReasons[why]++
				continue
			}
			// IDENTITY — a value the walker deliberately yields but no writer may
			// be asked to rewrite (a listing item's `name` is its page slug and the
			// stem of its own url). Recorded as an EXEMPTION, not filtered out, so
			// total = exempt + withinBudget + targets still reconciles with the
			// annotation's count over the same walk.
			//
			// ⚠ THIS MUST STAY ABOVE THE HEADLINE BRANCH, and the ordering is the
			// whole of the protection. `name` is headline-class as of 2026-09-04,
			// and a headline hit is never forgiven by the budget — so if this ran
			// after, an identity field would be FORCED to the model rather than
			// merely allowed there. A test pins this; nothing about the two field
			// regexes on their own would show it.
			if field.Identity != "" {
				plan.exemptCount++
				plan.exemptReasons[field.Identity]++
				continue
			}
			plan.pageHits++
			headline := datahelpers.IsHeadlineField(field.Path)
			key := field.Path + "\x00" + h.Sentence
			if seen[key] {
				continue
			}
			// THE BUDGET IS FORGIVENESS, AND ONLY A MILD SHAPE CAN SPEND IT.
			// OWNER RULING 2026-08-24 (D3): `rather than` is "a little bit of a
			// tic". Before it, the two survivors were whichever the scanner
			// walked past first — document order, nothing to do with severity —
			// so a page could keep both its `x_not_y` constructions and have the
			// gate rewrite two `rather than`s instead, which is the ruling
			// inverted. A sharp shape now never consumes the budget and is
			// always repaired; a mild one is tolerated up to `page_budget`.
			// `rather than` is still fully detected and still counts toward
			// `page_hits`, so the density signal is unchanged.
			if !headline && datahelpers.NegationShapeIsMild(h.Shape) {
				used++
				plan.mildHits++
				if used <= budget {
					plan.withinBudget++
					continue
				}
			}
			seen[key] = true
			plan.targets = append(plan.targets, negationTarget{
				Field: field.Path, Sentence: h.Sentence, MatchAt: h.MatchInSent,
				Shape: h.Shape, Headline: headline, set: field.Set, text: field.Text,
			})
		}
	}
	return plan
}

func copyGatePageBudget(config map[string]interface{}) int {
	if v, ok := config["page_budget"]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return 2
}

func pageHitsSoFar(params ActionParams) int {
	m, ok := params.CollectedData[copyGateMarkerKey].(map[string]interface{})
	if !ok {
		return 0
	}
	switch n := m["page_hits"].(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

// mildHitsSoFar / recordMildHits carry the MILD-shape count across the sections
// of one page, exactly as page_hits carries the total. Two counters are needed
// rather than one because `page_hits` is published for the density signal and
// counts every shape, while the budget must count only the shapes that may spend
// it — seeding the budget from the total would let a sharp hit in an earlier
// section consume a later section's forgiveness.
func mildHitsSoFar(params ActionParams) int {
	m, ok := params.CollectedData[copyGateMarkerKey].(map[string]interface{})
	if !ok {
		return 0
	}
	switch n := m["mild_hits"].(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

func recordMildHits(params ActionParams, hits int) {
	m, ok := params.CollectedData[copyGateMarkerKey].(map[string]interface{})
	if !ok {
		m = map[string]interface{}{}
		params.CollectedData[copyGateMarkerKey] = m
	}
	m["mild_hits"] = hits
}

func recordPageHits(params ActionParams, hits int) {
	m, ok := params.CollectedData[copyGateMarkerKey].(map[string]interface{})
	if !ok {
		m = map[string]interface{}{}
		params.CollectedData[copyGateMarkerKey] = m
	}
	m["page_hits"] = hits
}

// stampCopyGate merges this section's marker into the page-level one.
//
// NOT a shared marker utility, despite the generic name: 2 callers as of
// 2026-08-24, both in THIS file. A council seat reading the D3 submission
// reasonably assumed otherwise and objected that adding a skipped key might
// affect other pipelines — it cannot, and the absence of this comment is why the
// question had to be asked. `rewrite_negations` is dispatched by exactly ONE
// agent (`page-content-writer`); enumerate with the RECURSIVE walk, because the
// step is nested in a loop sub_workflow and a top-level `workflow.steps` query
// returns zero rows and reads as "no agent dispatches this":
//
//	FROM agent_definitions a,
//	     LATERAL jsonb_path_query(a.default_config,'$.**.steps') AS steps,
//	     LATERAL jsonb_each(steps) AS s(key,value)
//	WHERE s.value->>'action' = 'rewrite_negations'
//
// The `page_hits` / `mild_hits` keys are skipped here because they are owned
// per-section by recordPageHits / recordMildHits; letting a section's marker
// overwrite them would reset the per-PAGE budget on every section.
func stampCopyGate(params ActionParams, marker map[string]interface{}) {
	m, ok := params.CollectedData[copyGateMarkerKey].(map[string]interface{})
	if !ok {
		m = map[string]interface{}{}
		params.CollectedData[copyGateMarkerKey] = m
	}
	for k, v := range marker {
		if k == "page_hits" || k == "mild_hits" {
			continue // owned by recordPageHits/recordMildHits, which run per section
		}
		m[k] = v
	}
}

func countNegationHits(content map[string]interface{}) int {
	n := 0
	for _, f := range datahelpers.WalkContentStrings(content) {
		n += len(datahelpers.ScanDefineByNegation(f.Text))
	}
	return n
}

// collectBriefText flattens the brief-supplied fields to plain strings.
func collectBriefText(params ActionParams, config map[string]interface{}) []string {
	paths := defaultBriefFields
	if raw, ok := config["brief_fields"].([]interface{}); ok && len(raw) > 0 {
		paths = nil
		for _, p := range raw {
			if s, ok := p.(string); ok && s != "" {
				paths = append(paths, s)
			}
		}
	}
	var out []string
	for _, p := range paths {
		if v := datahelpers.ExtractNestedField(params.CollectedData, p); v != nil {
			if s := flattenToText(v); strings.TrimSpace(s) != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

func flattenToText(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []interface{}:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			parts = append(parts, flattenToText(e))
		}
		return strings.Join(parts, " \n")
	case map[string]interface{}:
		parts := make([]string, 0, len(t))
		for _, k := range sortedBriefKeys(t) {
			parts = append(parts, flattenToText(t[k]))
		}
		return strings.Join(parts, " \n")
	default:
		return ""
	}
}

// sortedBriefKeys: stable order so the exemption corpus is byte-identical run to
// run. (The package already has a sortedKeys for map[string]string.)
func sortedBriefKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

// ── the one repair call ─────────────────────────────────────────────────────

// negationRepairPrompt asks for the positive claim, and carries NO rule text and
// NO example of the construction it is removing.
//
// That is not fastidiousness. This estate ran the experiment: a rule was deleted
// from the voice prompt and its three worked examples left in place, and the
// behaviour continued unchanged — "the example is the instruction; the rule is
// commentary". A repair prompt that quotes the house-voice rule (which itself
// says "not what it is not" and demonstrates the shape) re-supplies the very
// form it is asking the model to drop.
func negationRepairPrompt(targets []negationTarget) string {
	var b strings.Builder
	// TRIAL (owner instruction, 2026-08-26, verbatim in substance): "whenever we
	// want to write the second half of one of these sentences, we should just
	// stop before the negative … and leave that part of the comparison out all
	// together. We don't need to sound competitive like this. There is no hidden
	// competition. We offer what we offer straight up." So the repair is
	// truncation-first, and SHORTER is the point — the old ask for "roughly the
	// same length" invited a rewritten comparison.
	b.WriteString(`Rewrite the sentences listed below.

Each one builds a comparison to say what something is. The repair is to END THE
SENTENCE BEFORE THE COMPARISON: keep the first half, the part that says what the
thing IS or DOES, tidy the punctuation so it stands as a complete sentence, and
leave the compared alternative out altogether. The result should usually be
SHORTER than the original; that is the point. Only when the first half cannot
stand on its own may you rewrite the whole sentence to say directly what it
says. There is no hidden competition to argue against: the site offers what it
offers, straight up.

For a heading or title, what remains may be very short. Two or three words that
name the thing are a complete heading; do not pad one back up to sentence length.

Keep, exactly as they are: every number, every name, every link or URL, and any
markup (tags like <p> or <a>) the sentence already contains. Keep any statement
of something we do not do, cannot promise, or cannot guarantee — those are there
on purpose; end each one where its own statement ends. Keep the same voice.

Say only what the original sentence already supports. Do not add a claim about
capability, coverage, speed, reliability, accuracy or completeness, and do not
reach for a superlative to fill the gap the removed contrast leaves: "the
definitive", "fully verified", "guaranteed", "always", "every" are all worse than
the sentence you were given. If what remains is a plain, modest statement, that
is the right answer.

Return ONLY this JSON object, nothing before or after it:

{"replacements": [{"field": "<field>", "from": "<the sentence, copied exactly>", "to": "<your rewrite>"}]}

If a sentence is already right, or you cannot rewrite it without losing meaning,
leave it out of the list. A shorter list is a fine answer.

Sentences:
`)
	for i, t := range targets {
		fmt.Fprintf(&b, "\n%d. field: %s%s\n   %s\n", i+1, t.Field,
			map[bool]string{true: "  (this is a heading or title)", false: ""}[t.Headline], t.Sentence)
	}
	return b.String()
}

// runNegationRepair makes the single call, judges each replacement, and splices
// the accepted ones into the content map in place.
func runNegationRepair(ctx context.Context, params ActionParams, config map[string]interface{}, plan negationPlan) ([]map[string]interface{}, []map[string]interface{}, string) {
	rewritten := []map[string]interface{}{}
	rejected := []map[string]interface{}{}
	// field path -> that field's text as it stands after the replacements
	// accepted so far. See the splice below for why this is not optional.
	spliced := map[string]string{}

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
		return rewritten, rejected, "no ai_service configuration resolvable"
	}
	client, err := createAIClient(ctx, aiServiceConfig)
	if err != nil {
		return rewritten, rejected, "ai client: " + err.Error()
	}

	prompt := negationRepairPrompt(plan.targets)
	// The budget is resolved by the ONE resolver this package has
	// (llm_options.go), never here.
	//
	// This block used to read `ai_service.max_tokens` by hand and fall back to a
	// hardcoded `2000`. Both halves were wrong, and the literal was the worse one:
	// an explicitly supplied option WINS at the wire (anthropic.go:307), so a call
	// that always sends a number can never inherit anything, and the literal
	// therefore DEFEATED bugs_open/257's client-side resolution rather than being
	// covered by it. Passing nothing is strictly safer than passing 2000.
	//
	// The small budget this step wants is not lost — it is migration 517's
	// `ai_service.max_tokens`, raised to 16000 by migration 569 when the 305 lane
	// measured a dense page repairing zero of ten targets. That is a lever an
	// operator can move; a Go literal is not, and for three days in August the two
	// were the same number, which made the fleet's own instrument unable to tell a
	// working config from a dropped one (bugs_open/257 §2026-09-03).
	options := llmOptionsFromConfig(config, aiServiceConfig, params.Logger, "rewrite_negations")
	if t, ok := aiServiceConfig["temperature"].(float64); ok {
		options["temperature"] = t
	}

	start := time.Now()
	raw, callErr := client.GenerateText(ctx, prompt, options)
	latency := int(time.Since(start).Milliseconds())

	// A TRUNCATED repair is not a smaller repair — it is an answer that was cut,
	// and a `to` string cut mid-sentence can still look like prose. The provider
	// surfaces this as a typed error; a provider that does not is caught by the
	// usage check below, because a call that stopped AT its ceiling is the shape
	// of a cut answer. Either way: log it and splice nothing (council round 1,
	// llm_reliability seat).
	truncated := false
	if te, isTrunc := aiservice.IsTruncated(callErr); isTrunc {
		truncated = true
		_ = te
	}

	// One call, one forensic row — the rule every LLM caller in this package
	// follows. The marker makes a census able to separate repair calls from the
	// writer's own; note it is present on SUCCESSFUL repairs too, so filter
	// failures on success=false, never on a non-empty error_message
	// (the bugs_open/119 precedent, which has already misled one census).
	logMsg := "RETRY (bugs_open/305: define-by-negation rewrite)"
	if callErr != nil {
		logMsg = "RETRY (bugs_open/305) FAILED: " + callErr.Error()
	}
	inTok, _ := options["__usage_input_tokens"].(int)
	outTok, _ := options["__usage_output_tokens"].(int)
	// ⚠ THE CEILING OF RECORD IS THE ONE THAT WAS APPLIED, NOT THE ONE WE ASKED
	// FOR. The provider writes `__sent_max_tokens` back into this same options
	// map during the call, and that is what the platform means by the ceiling:
	// `ai_actions.go` feeds `llm_call_log.max_tokens` from it, and gemini
	// deliberately reports the VISIBLE-text budget there so it stays
	// commensurable with `__usage_output_tokens` across providers
	// (bugs_open/110). Our own `max_tokens` is only the request: equal on
	// anthropic, and ABSENT when a caller lets the client choose its own cap —
	// where reading it would report UNKNOWN over a ceiling the provider knew.
	// Ollama records no sent value, hence the fallback.
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
		PromptTemplate: "rewrite_negations", PromptRendered: prompt, ResponseText: raw,
		InputTokens: inTok, OutputTokens: outTok, LatencyMs: latency,
		Success: callErr == nil, ErrorMessage: logMsg,
		// Same column, same meaning as every other caller: the ceiling the
		// provider APPLIED. Also no longer an unchecked type assertion — this
		// runs inside an action, where a panic kills the step, and it was one
		// edit to the options block above away from firing.
		MaxTokens: sentMax, Options: options,
	})
	if callErr != nil {
		params.Logger.Warn("rewrite_negations: the repair call failed — leaving the copy as written",
			zap.Error(callErr))
		return rewritten, rejected, callErr.Error()
	}
	// ⚠ THE SECOND ARM CANNOT BE TRUSTED TO FIRE, and saying so is the point
	// (council round 2, llm_reliability, HIGH). A provider that returns a cut
	// answer as a 200 may report NO usage at all, and `outTok >= sent` with
	// outTok defaulting to 0 is then FALSE — the arm silently never runs, which
	// is indistinguishable from "the answer was complete". So the three states
	// are kept distinct: known-and-at-the-ceiling (truncated), known-and-below
	// (fine), and UNKNOWN, which is recorded on the marker rather than assumed
	// safe. The load-bearing protection is not this arm anyway: it is the typed
	// truncation error above, plus the fact that a cut JSON object does not
	// parse and an unparseable answer splices nothing.
	usage := aiservice.ClassifyTruncation(outTok, sentMax)
	if usage.Truncated() {
		truncated = true
	}
	if truncated {
		params.Logger.Warn("rewrite_negations: the repair answer hit the output ceiling — splicing nothing",
			zap.Int("output_tokens", outTok), zap.Any("max_tokens", options["max_tokens"]))
		return rewritten, rejected, "repair answer truncated at the output ceiling"
	}
	if usage == aiservice.TruncationUnknown {
		params.Logger.Info("rewrite_negations: the provider reported no output-token usage, so the ceiling check could not run — relying on the parse",
			zap.String("step_name", stepName), zap.String("usage_state", usage.String()))
	}

	parsed, _, perr := ParseLLMJSONWithProvenance(stripMarkdownFromResponse(raw))
	if perr != nil {
		params.Logger.Warn("rewrite_negations: the repair answer would not parse — leaving the copy as written",
			zap.Error(perr))
		return rewritten, rejected, "unparseable repair answer"
	}
	// ⚠ THIS LOOP RUNS OVER WHAT THE MODEL RETURNED, NOT OVER WHAT WE ASKED
	// ABOUT — so a target the answer simply never mentions is visited by no
	// branch below and lands in NEITHER list. `answered` is what closes that,
	// after the loop. Measured 2026-08-23 before the fix: **15 of 49** markers
	// with `targets > 0` did not reconcile (`targets != rewritten + rejected`)
	// and **12 of those accounted for NONE of their targets**.
	//
	// `hits_after` stayed honest throughout — it is recomputed from the real
	// content — so this never misdescribed a PAGE. What it corrupted is the
	// INSTRUMENT: this action's whole displacement defence is "every rejection
	// is recorded with its reason, and that log is how we find out whether the
	// repair is fixing the copy or teaching the model a new tic". A silently
	// dropped target is an outcome that log cannot show, and `bugs_open/305`'s
	// D3 (is `rather than` a tic or ordinary English?) is explicitly to be
	// settled FROM that log — so an unreconciled 31% was about to inform an
	// owner decision.
	answered := make(map[string]bool, len(plan.targets))
	for _, r := range decodeReplacements(parsed) {
		t, found := matchTarget(plan.targets, r.Field, r.From)
		if found {
			answered[negationTargetKey(t)] = true
		}
		if !found {
			rejected = append(rejected, map[string]interface{}{
				"field": r.Field, "reason": "no_such_sentence",
				"from": datahelpers.TruncateString(r.From, 120)})
			continue
		}
		// CLAIM SAFETY, before structure (council round 1, compliance seat, HIGH).
		// AcceptNegationRewrite is purely structural — it cannot tell an honest
		// reframing from an overclaim. "Say what it IS" is exactly the pressure
		// that fills an affirmative slot with an invented superlative ("the
		// definitive source", "fully verified"), and nothing downstream inspects
		// a spliced sentence for claim content: the deploy-time claims gate reads
		// the page, and by then this sentence is the page. So the candidate goes
		// through the SAME banned-claims scan the meta-description gate uses,
		// including its fleet-wide arm, which applies to a site with no evidence
		// register at all.
		if reason := negationRewriteClaimSafe(ctx, params, r.To); reason != "" {
			rejected = append(rejected, map[string]interface{}{
				"field": t.Field, "reason": reason, "shape": t.Shape,
				"from": datahelpers.TruncateString(t.Sentence, 160),
				"to":   datahelpers.TruncateString(r.To, 160)})
			continue
		}
		// A HEADING IS JUDGED BY THE HEADING FLOOR (owner ruling 2026-09-03).
		// Only the word floor differs; every other guard is the same code. The
		// sentence floor of 5 was calibrated on body sentences and would refuse
		// 25 of the 36 live heading repairs bugs_open/420 exposed — making them
		// visible and then declining to fix them.
		judge := datahelpers.AcceptNegationRewrite
		if t.Headline {
			judge = datahelpers.AcceptNegationHeadingRewrite
		}
		if ok, why := judge(t.Sentence, r.To, t.MatchAt); !ok {
			rejected = append(rejected, map[string]interface{}{
				"field": t.Field, "reason": why, "shape": t.Shape,
				"from": datahelpers.TruncateString(t.Sentence, 160),
				"to":   datahelpers.TruncateString(r.To, 160)})
			continue
		}
		// Splice: replace the sentence, once, inside the field's CURRENT value.
		//
		// ⚠ CURRENT, not `t.text`. Every target of one field shares the same
		// captured original, so splicing each against `t.text` and writing the
		// whole field back makes each accepted rewrite OVERWRITE the previous
		// one — last writer wins, and the marker still reports them all as
		// rewritten. Measured in production 2026-08-21 (orch ce002822,
		// webdesign.co.uk/tool-social-card-guide): SIX accepted replacements,
		// all into one `content` field, and `hits_before 8 -> hits_after 7` —
		// one net repair out of six, with the marker claiming six. A lost repair
		// AND a false report, which is the worse half.
		//
		// `spliced` carries each field's text as it stands after earlier
		// replacements, so N targets in one field compose instead of racing.
		// The map this writes to is the one the renderer reads, so nothing has
		// to be copied back.
		base, seen := spliced[t.Field]
		if !seen {
			base = t.text
		}
		updated := strings.Replace(base, t.Sentence, strings.TrimSpace(r.To), 1)
		if updated == base {
			rejected = append(rejected, map[string]interface{}{
				"field": t.Field, "reason": "splice_missed",
				"from": datahelpers.TruncateString(t.Sentence, 160)})
			continue
		}
		spliced[t.Field] = updated
		t.set(updated)
		rewritten = append(rewritten, map[string]interface{}{
			"field": t.Field, "shape": t.Shape, "headline": t.Headline,
			"from": datahelpers.TruncateString(t.Sentence, 160),
			"to":   datahelpers.TruncateString(strings.TrimSpace(r.To), 160)})
	}

	// Every target we asked about and did not hear back on. Recorded as a
	// rejection with its own reason rather than silently dropped, so a census can
	// tell "the model declined this shape" from "the model never saw it". It is
	// deliberately NOT a failure: the copy stands as written, exactly as it does
	// for any other rejection, and `hits_after` already said so.
	//
	// ⚠ THE RECONCILIATION INVARIANT, STATED PRECISELY — an earlier version of
	// this comment claimed `targets == len(rewritten) + len(rejected)` holds "for
	// every marker", and that is FALSE in two directions. Both are now measured in
	// production, so a census written against the loose form flags healthy markers:
	//
	//	targets == len(rewritten) + len(rejected) - count(reason="no_such_sentence")
	//	                    ... and ONLY for a marker whose status is "repaired"
	//
	//  (1) UNDER-COUNT. The five early returns above (no ai_service, no client,
	//      call error, output ceiling, unparseable answer) all return BEFORE this
	//      line, so a `repair_unavailable` marker accounts for none of its targets
	//      by design. Its `error` field names which one fired. Segment by status.
	//  (2) OVER-COUNT. A replacement naming a sentence that is in no target is
	//      rejected as `no_such_sentence` — an entry with NO target behind it, so
	//      it pushes the sum ABOVE `targets`. Measured 2026-08-24: 1 marker in 122
	//      post-roll (`targets=5, rewritten=4, rejected=2`, the 2 being one
	//      `no_answer_for_target` and one `no_such_sentence`) — all 5 targets
	//      correctly accounted, plus one hallucinated sentence correctly logged.
	//
	// `TestReconciliationExcludesHallucinatedReplacements` pins (2); the whole
	// point is that a census must not be "fixed" until it reads zero, because
	// doing so hides the ceiling failures of (1). See bugs_open/305 §27a, §29.
	rejected = append(rejected, unansweredTargetRejections(plan.targets, answered)...)
	return rewritten, rejected, ""
}

// negationTargetKey identifies a target for answered-bookkeeping. The NUL
// separator is what stops a field name ending in the first characters of a
// sentence from colliding with a shorter field plus a longer sentence, and the
// sentence half is normalised so the key matches on the same terms
// `matchTarget` does — otherwise a target could be answered and still counted
// unanswered on a curly apostrophe.
func negationTargetKey(t negationTarget) string {
	return t.Field + "\x00" + normaliseSentenceKey(t.Sentence)
}

// unansweredTargetRejections records every target the repair asked about and
// heard nothing back on, so that for every marker
//
//	targets == len(rewritten) + len(rejected)
//
// It is separate from the loop above, and exported to the test, because the
// invariant is the point: the loop iterates the MODEL'S answer, so a target the
// answer never mentions is visited by no branch and would otherwise be recorded
// nowhere at all.
func unansweredTargetRejections(targets []negationTarget, answered map[string]bool) []map[string]interface{} {
	var out []map[string]interface{}
	for _, t := range targets {
		if answered[negationTargetKey(t)] {
			continue
		}
		out = append(out, map[string]interface{}{
			"field": t.Field, "reason": "no_answer_for_target", "shape": t.Shape,
			"from": datahelpers.TruncateString(t.Sentence, 160)})
	}
	return out
}

// negationRewriteClaimSafe returns a rejection reason when the candidate
// replacement introduces a banned or unevidenced claim. Empty string = safe.
//
// It reuses the estate's one banned-claims implementation rather than adding a
// second opinion about what an overclaim is. A site with no evidence register
// still gets the fleet-wide arm (RFC 003 / bugs_closed/104), which is the arm
// that matters here: the register is sparse, and this seam is where new text is
// born.
func negationRewriteClaimSafe(ctx context.Context, params ActionParams, candidate string) string {
	if params.DB == nil {
		return "" // no database, no register, and no way to ask — never a silent block
	}
	siteID := uuid.Nil
	for _, path := range []string{"site_record.site_id", "site_record.id", "input_data.site_id", "site_id"} {
		if s := datahelpers.ExtractNestedFieldString(params.CollectedData, path); s != "" {
			if id, err := uuid.Parse(s); err == nil {
				siteID = id
				break
			}
		}
	}
	if siteID == uuid.Nil {
		return ""
	}
	eb := loadEvidenceBase(ctx, params.DB, siteID, params.Logger)
	if issues := checkBannedClaims([]string{candidate}, eb, true, siteID.String(), params.Logger); len(issues) > 0 {
		return "banned_claim_" + issues[0].Category
	}
	return ""
}

type negationReplacement struct {
	Field string `json:"field"`
	From  string `json:"from"`
	To    string `json:"to"`
}

func decodeReplacements(parsed interface{}) []negationReplacement {
	b, err := json.Marshal(parsed)
	if err != nil {
		return nil
	}
	var envelope struct {
		Replacements []negationReplacement `json:"replacements"`
	}
	if err := json.Unmarshal(b, &envelope); err == nil && len(envelope.Replacements) > 0 {
		return envelope.Replacements
	}
	// A bare array is a common shape from a model that skipped the envelope, and
	// it costs three lines to accept rather than throwing the round away.
	var bare []negationReplacement
	if err := json.Unmarshal(b, &bare); err == nil {
		return bare
	}
	return nil
}

// matchTarget finds the planned target a replacement refers to.
//
// The field name is a HINT, not the key: models rename fields, and a rewrite
// matched to the wrong field would be spliced into the wrong copy. The sentence
// is the identity, compared on its normalised form so that a model which
// re-typed a curly quote still matches.
func matchTarget(targets []negationTarget, field, from string) (negationTarget, bool) {
	want := normaliseSentenceKey(from)
	if want == "" {
		return negationTarget{}, false
	}
	for _, t := range targets {
		if normaliseSentenceKey(t.Sentence) == want {
			return t, true
		}
	}
	for _, t := range targets {
		if t.Field == field && strings.Contains(normaliseSentenceKey(t.Sentence), want) {
			return t, true
		}
	}
	return negationTarget{}, false
}

func normaliseSentenceKey(s string) string {
	r := strings.NewReplacer("’", "'", "“", `"`, "”", `"`, "\n", " ", "\t", " ")
	return strings.ToLower(strings.Join(strings.Fields(r.Replace(s)), " "))
}
