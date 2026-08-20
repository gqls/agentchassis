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
// ⚠ The cross-iteration persistence of that counter is READ FROM the loop
// design, not yet proven in production. If it does not hold, every section sees
// a fresh budget and the gate degrades to "repair headlines, allow two per
// section" — which is weaker than intended but not wrong. There is a canary in
// the lane's RUNBOOK to settle it after the roll.

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
	plan := planNegationRepairs(content, supplied, copyGatePageBudget(config), pageHitsSoFar(params))
	recordPageHits(params, plan.pageHits)

	marker := map[string]interface{}{
		"hits_before":    plan.total,
		"exempt":         plan.exemptCount,
		"exempt_reasons": plan.exemptReasons,
		"within_budget":  plan.withinBudget,
		"targets":        len(plan.targets),
		"page_hits":      plan.pageHits,
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
		marker["status"] = "repair_unavailable"
		marker["error"] = callErr
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
	targets       []negationTarget
}

// planNegationRepairs decides, per hit, whether it is exempt, allowed by the page
// budget, or a target. Order matters and is fixed: exemptions first (they are
// never counted against the budget, because a brief-supplied phrase is not the
// writer's doing), then headline hits (always repaired), then the budget.
func planNegationRepairs(content map[string]interface{}, supplied []string, budget, alreadyUsed int) negationPlan {
	plan := negationPlan{exemptReasons: map[string]int{}, pageHits: alreadyUsed}
	used := alreadyUsed
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
			plan.pageHits++
			headline := datahelpers.IsHeadlineField(field.Path)
			key := field.Path + "\x00" + h.Sentence
			if seen[key] {
				continue
			}
			if !headline {
				used++
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

func recordPageHits(params ActionParams, hits int) {
	m, ok := params.CollectedData[copyGateMarkerKey].(map[string]interface{})
	if !ok {
		m = map[string]interface{}{}
		params.CollectedData[copyGateMarkerKey] = m
	}
	m["page_hits"] = hits
}

func stampCopyGate(params ActionParams, marker map[string]interface{}) {
	m, ok := params.CollectedData[copyGateMarkerKey].(map[string]interface{})
	if !ok {
		m = map[string]interface{}{}
		params.CollectedData[copyGateMarkerKey] = m
	}
	for k, v := range marker {
		if k == "page_hits" {
			continue // owned by recordPageHits, which runs per section
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
	b.WriteString(`Rewrite the sentences listed below.

Each one says what something is NOT in order to say what it is. Say what it IS,
or what it DOES, directly. Drop the contrasted alternative unless the sentence
is meaningless without it.

Keep, exactly as they are: every number, every name, every link or URL, and any
markup (tags like <p> or <a>) the sentence already contains. Keep any statement
of something we do not do, cannot promise, or cannot guarantee — those are there
on purpose. Keep the same voice and roughly the same length.

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
	options := map[string]interface{}{}
	if mt, ok := aiServiceConfig["max_tokens"].(float64); ok && mt > 0 {
		options["max_tokens"] = int(mt)
	} else {
		// The answer is a handful of sentences; a section-sized ceiling would only
		// buy room for the model to write an essay we would then reject.
		options["max_tokens"] = 2000
	}
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
		MaxTokens: options["max_tokens"].(int), Options: options,
	})
	if callErr != nil {
		params.Logger.Warn("rewrite_negations: the repair call failed — leaving the copy as written",
			zap.Error(callErr))
		return rewritten, rejected, callErr.Error()
	}
	if sent, ok := options["max_tokens"].(int); ok && sent > 0 && outTok >= sent {
		truncated = true
	}
	if truncated {
		params.Logger.Warn("rewrite_negations: the repair answer hit the output ceiling — splicing nothing",
			zap.Int("output_tokens", outTok), zap.Any("max_tokens", options["max_tokens"]))
		return rewritten, rejected, "repair answer truncated at the output ceiling"
	}

	parsed, _, perr := ParseLLMJSONWithProvenance(stripMarkdownFromResponse(raw))
	if perr != nil {
		params.Logger.Warn("rewrite_negations: the repair answer would not parse — leaving the copy as written",
			zap.Error(perr))
		return rewritten, rejected, "unparseable repair answer"
	}
	for _, r := range decodeReplacements(parsed) {
		t, found := matchTarget(plan.targets, r.Field, r.From)
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
		if ok, why := datahelpers.AcceptNegationRewrite(t.Sentence, r.To, t.MatchAt); !ok {
			rejected = append(rejected, map[string]interface{}{
				"field": t.Field, "reason": why, "shape": t.Shape,
				"from": datahelpers.TruncateString(t.Sentence, 160),
				"to":   datahelpers.TruncateString(r.To, 160)})
			continue
		}
		// Splice: replace the sentence, once, inside the field's own value. The
		// map this writes to is the one the renderer reads, so nothing has to be
		// copied back.
		updated := strings.Replace(t.text, t.Sentence, strings.TrimSpace(r.To), 1)
		if updated == t.text {
			rejected = append(rejected, map[string]interface{}{
				"field": t.Field, "reason": "splice_missed",
				"from": datahelpers.TruncateString(t.Sentence, 160)})
			continue
		}
		t.set(updated)
		rewritten = append(rewritten, map[string]interface{}{
			"field": t.Field, "shape": t.Shape, "headline": t.Headline,
			"from": datahelpers.TruncateString(t.Sentence, 160),
			"to":   datahelpers.TruncateString(strings.TrimSpace(r.To), 160)})
	}
	return rewritten, rejected, ""
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
