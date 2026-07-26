package main

// engine.go — the ideation method (v2) as a staged, multi-model, web-verified
// pipeline. Go port of idea_method_runner.py. Calls Anthropic/OpenAI directly
// over net/http (no SDKs). Output is similar, not identical, to the by-hand runs;
// LLM output is non-deterministic.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Models. The cut runs on a DIFFERENT vendor if OPENAI_API_KEY is set
// (genuine cross-vendor critique), else a different Anthropic model.
// Strings verified against the current model catalogue (2026-07-26); override
// via env. Moved off Opus 4.8 / Sonnet 4.6 to the 5 family on that date — see
// usesManualThinkingBudget below, which had to change with them: the 5 family
// REJECTS the manual thinking budget the 4.6-and-older wire format uses, so a
// model swap on its own would have 400'd every call.
var (
	genModel      = env("GEN_MODEL", "claude-opus-5")
	critiqueModel = env("CRITIQUE_MODEL", "claude-sonnet-5")
	verifyModel   = env("VERIFY_MODEL", "claude-opus-5")
	scoreModel    = env("SCORE_MODEL", "claude-sonnet-5")
	// Cross-vendor cut only; dormant unless OPENAI_API_KEY is set. Deliberately
	// NOT updated alongside the Anthropic models — it is a different vendor's
	// catalogue and nothing here has verified a current id for it.
	openAIModel = env("OPENAI_CRITIQUE_MODEL", "gpt-4o")
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// envInt reads an int from the environment, falling back to def on unset or
// unparseable. Used for tunables like the web-search budget.
func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// 15-minute ceiling. The verify step (Opus 4.8 at xhigh effort + up to 6 web
// searches) is a long agentic call: the server holds the connection while it
// thinks, searches, thinks again — no response headers arrive until it's done,
// which blew the old 180s timeout ("context deadline exceeded while awaiting
// headers"). The other calls finish in seconds; this is just a safe ceiling.
// The more robust long-term fix for very long calls is streaming (Anthropic
// recommends it for large thinking budgets — incremental SSE events keep the
// connection alive and dodge intermediary proxy timeouts); that's a bigger
// change deferred for now.
var httpClient = &http.Client{Timeout: 900 * time.Second}

// EngineFunc is the shape both front doors call. Swappable in tests.
// It returns the report in two forms: Text (plain, for the email body and the
// JSON record) and HTML (styled, for the multipart HTML email).
type EngineFunc func(domain, audience, assets string) (renderedReport, error)

// ── JSON shapes between steps ────────────────────────────────────────────────

// source is one checkable reference the model actually relied on. Carried into
// the rendered report so "check it yourself" is true of what the customer
// receives, not just of our process.
type source struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// assessment is step 0 — the researched assessment of the thing the customer
// actually submitted. Field-for-field it is the coverage the /report.html copy
// promises (problem/evidence, who else, substitutes, defensible/exposed, next
// step); IsAssessable=false is the honest "this is too early to assess" outcome
// the copy also promises ("Not every idea warrants a report").
type assessment struct {
	IsAssessable     bool     `json:"is_assessable"`
	Reading          string   `json:"reading"`
	Problem          string   `json:"problem"`
	DemandEvidence   string   `json:"demand_evidence"`
	WhoElse          string   `json:"who_else"`
	SubstitutesToday string   `json:"substitutes_today"`
	Defensible       string   `json:"defensible"`
	Exposed          string   `json:"exposed"`
	NextStep         string   `json:"next_step"`
	Sources          []source `json:"sources"`
}

type candidate struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Lens             string   `json:"lens"`
	Asset            string   `json:"asset"`
	Capability       string   `json:"capability"`
	BeatsFreeBecause string   `json:"beats_free_because"`
	Findings         string   `json:"findings,omitempty"`
	Sources          []source `json:"sources,omitempty"`
}

type scored struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Defensibility      int    `json:"defensibility"`
	Willingness        int    `json:"willingness"`
	Buildability       int    `json:"buildability"`
	Reuse              int    `json:"reuse"`
	Durability         int    `json:"durability"`
	Risk               int    `json:"risk"` // 1–5, 5 = low operator-side risk (consequences of being wrong). Separate from Sum; gates and ranks via dedicated rules.
	Sum                int    `json:"sum"`  // Def+Will+Build+Reuse+Dur (Risk excluded — it's a hazard label, not a fitness contribution).
	Advances           bool   `json:"advances"`
	ShortLived         bool   `json:"short_lived"`
	NeedsLiabilityWork bool   `json:"-"` // derived in Go from Risk ≤ 2; not asked of the model.
	CheapestTest       string `json:"cheapest_test"`
	Flag               string `json:"flag"`
	candidate          `json:"-"`
}

// ── API calls ────────────────────────────────────────────────────────────────
// callOpts bundles the parameters for an Anthropic Messages API call. Most
// fields are optional. Defaults preserve the simple-string-system / no-thinking
// behaviour the engine had before extended thinking and prompt caching landed.
//
//   - Effort one of low|medium|high|xhigh|max. Every call site sets it.
//     Effort "" sends NO thinking field, which is no longer the same as "no
//     thinking": on Opus 4.8 an omitted field meant thinking off, but on the 5
//     family it means thinking runs adaptively by default. Leave it set.
//     The wire format DIFFERS BY MODEL (this bit us — see debugging guide §0):
//     Opus 4.7+ and the 5 family use adaptive thinking (thinking:{type:adaptive}
//   - output_config:{effort}); they 400 on manual budgets. Sonnet 4.6 and
//     older use manual extended thinking (thinking:{type:enabled,budget_tokens}).
//     callClaudeOpts picks via usesManualThinkingBudget, a deny-list, so an
//     unrecognised (newer) model gets the modern format rather than a 400.
//   - CacheSystem = true wraps the system prompt as a single content block with
//     cache_control: ephemeral, so the system text becomes a cache hit on
//     subsequent calls within the same TTL (5 min). One run makes 5 calls in
//     under 10 min, so steps 2–5 land cache reads on the (identical) system.
type callOpts struct {
	Model       string
	System      string
	User        string
	Tools       []map[string]any
	MaxTokens   int
	Effort      string // "" = no thinking; else low|medium|high|xhigh|max
	CacheSystem bool
}

// legacyManualBudgetModels are the model families that still take the OLD
// thinking wire format, thinking:{type:"enabled",budget_tokens:N}. Everything
// from Opus 4.7 onwards — including the whole 5 family — rejects it with a 400
// and takes thinking:{type:"adaptive"} + output_config.effort instead.
//
// Matching on substrings keeps this working across the dated/aliased forms of
// the same id ("claude-sonnet-4-6", "claude-sonnet-4-5-20250929", …).
var legacyManualBudgetModels = []string{
	"opus-4-6", "opus-4-5", "opus-4-1", "opus-4-0",
	"sonnet-4-6", "sonnet-4-5", "sonnet-4-0", "sonnet-3",
	"haiku",
}

// usesManualThinkingBudget reports whether a model needs the legacy manual
// budget. It is deliberately a DENY-LIST, and that inversion is the point.
//
// This function used to be an allow-list of models that take adaptive thinking
// (opus-4-7 / opus-4-8 / mythos), so any model it had never heard of fell
// through to the manual-budget branch. That is backwards: a newer model is
// exactly the case that rejects manual budgets, so the old shape turned every
// future upgrade into a 400 at runtime — which is precisely what the 2026-07-26
// move to the 5 family would have hit. Unknown models now get the modern
// format, and only the known-old list gets the legacy one.
func usesManualThinkingBudget(model string) bool {
	for _, legacy := range legacyManualBudgetModels {
		if strings.Contains(model, legacy) {
			return true
		}
	}
	return false
}

// effortToBudget maps an effort level to a manual thinking budget (tokens), for
// the older models that still use thinking:{type:enabled,budget_tokens}.
func effortToBudget(effort string) int {
	switch effort {
	case "low":
		return 2000
	case "medium":
		return 4000
	case "high":
		return 8000
	case "xhigh":
		return 12000
	case "max":
		return 16000
	default:
		return 0
	}
}

func callClaudeOpts(o callOpts) (string, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return "", errors.New("ANTHROPIC_API_KEY not set")
	}
	body := map[string]any{
		"model":      o.Model,
		"max_tokens": o.MaxTokens,
		"messages":   []map[string]any{{"role": "user", "content": o.User}},
	}
	// System: either a plain string, or a single cached content block.
	if o.CacheSystem {
		body["system"] = []map[string]any{{
			"type":          "text",
			"text":          o.System,
			"cache_control": map[string]any{"type": "ephemeral"},
		}}
	} else {
		body["system"] = o.System
	}
	if o.Tools != nil {
		body["tools"] = o.Tools
	}
	// Thinking — wire format depends on the model family.
	//
	// Effort is now set at EVERY call site on purpose. It used to be optional,
	// and an empty Effort meant "send no thinking field at all" — which on
	// Opus 4.8 meant no thinking. On the 5 family the same omission means
	// thinking runs ADAPTIVELY by default, and thinking shares the max_tokens
	// cap with the answer, so the silent effect of upgrading would have been
	// longer, more expensive calls that can truncate mid-answer. Being explicit
	// keeps the behaviour a decision rather than a default that moved under us.
	if o.Effort != "" {
		if usesManualThinkingBudget(o.Model) {
			// Sonnet 4.6 and older: manual budget. Must be >=1024 and < max_tokens,
			// leaving room for the actual output.
			b := effortToBudget(o.Effort)
			if b < 1024 {
				b = 1024
			}
			if b >= o.MaxTokens {
				b = o.MaxTokens - 1024
			}
			body["thinking"] = map[string]any{
				"type":          "enabled",
				"budget_tokens": b,
				"display":       "omitted", // we never surface thinking; skip the wire cost
			}
		} else {
			// Opus 4.7+ and the whole 5 family: adaptive thinking + effort.
			// A manual budget is a 400 here.
			body["thinking"] = map[string]any{"type": "adaptive"}
			body["output_config"] = map[string]any{"effort": o.Effort}
		}
	}
	raw, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(raw))
	req.Header.Set("x-api-key", key)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("anthropic %d: %s", resp.StatusCode, b)
	}
	var out struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens              int `json:"input_tokens"`
			OutputTokens             int `json:"output_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	// A CUT completion is not a short one — it is a fragment that looks like an
	// answer. HTTP is 200, the text parses, and nothing downstream can tell.
	// This matters more since the move to the 5 family: thinking is on by
	// default there and shares the max_tokens cap with the answer, so the same
	// prompt that fitted comfortably before can now run out of room. Fail loudly
	// instead of persisting half a report.
	if out.StopReason == "max_tokens" {
		return "", fmt.Errorf("anthropic %s: response CUT at max_tokens=%d (output=%d) — "+
			"raise MaxTokens for this step or lower its effort; the partial answer was discarded",
			o.Model, o.MaxTokens, out.Usage.OutputTokens)
	}
	// A refusal is a successful HTTP 200 with no usable content. Surfacing it as
	// an error keeps it out of a customer's report.
	if out.StopReason == "refusal" {
		return "", fmt.Errorf("anthropic %s: request declined by safety classifiers (stop_reason=refusal)", o.Model)
	}
	// Log the cache hit rate so the operator can see caching actually working
	// when it should. Cache reads come "for free" at 10% of the input rate.
	if out.Usage.CacheReadInputTokens > 0 || out.Usage.CacheCreationInputTokens > 0 {
		fmt.Fprintf(os.Stderr, "[cache] %s: created=%d read=%d input=%d output=%d\n",
			o.Model, out.Usage.CacheCreationInputTokens, out.Usage.CacheReadInputTokens,
			out.Usage.InputTokens, out.Usage.OutputTokens)
	}
	var sb strings.Builder
	for _, c := range out.Content { // concat text blocks; skip thinking/tool blocks
		if c.Type == "text" {
			sb.WriteString(c.Text)
		}
	}
	return sb.String(), nil
}

// callClaude is the simple form. It currently has NO CALLERS — the comment here
// used to say it "keeps every existing call site unchanged", and there are none
// left. It is kept as the convenience wrapper for a future caller.
//
// ⚠️ If you revive it, set Effort. It passes none, which sends no thinking field
// — and on the 5 family that means adaptive thinking runs by DEFAULT and shares
// the max_tokens cap with the answer, so a caller sized for a bare completion
// can truncate. That is the same trap the audience and generate steps hit in the
// 2026-07-26 model migration; they now set Effort explicitly. Prefer
// callClaudeOpts directly so the choice is visible at the call site.
func callClaude(model, system, user string, tools []map[string]any, maxTokens int) (string, error) {
	return callClaudeOpts(callOpts{
		Model: model, System: system, User: user, Tools: tools, MaxTokens: maxTokens,
	})
}

func callOpenAI(system, user string) (string, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return "", errors.New("OPENAI_API_KEY not set")
	}
	body := map[string]any{
		"model": openAIModel,
		"messages": []map[string]any{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
	}
	raw, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("content-type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("openai %d: %s", resp.StatusCode, b)
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", errors.New("openai: no choices")
	}
	return out.Choices[0].Message.Content, nil
}

// critique runs the cut on a different vendor if available, else a different
// Anthropic model — so the method isn't one vendor marking its own work.
func critique(user string) (string, error) {
	if os.Getenv("OPENAI_API_KEY") != "" {
		fmt.Fprintf(os.Stderr, "[cut] cross-vendor: OpenAI (%s)\n", openAIModel)
		return callOpenAI(systemBase, user)
	}
	fmt.Fprintf(os.Stderr, "[cut] same-vendor: Anthropic (%s) with extended thinking\n", critiqueModel)
	return callClaudeOpts(callOpts{
		Model:  critiqueModel,
		System: systemBase,
		User:   user,
		// Raised 12000→16000 with the move to Sonnet 5: its tokenizer produces
		// ~30% more tokens for the same text than Sonnet 4.6, so a cap sized
		// against 4.6 can cut equivalent output. max_tokens is a ceiling, not a
		// reservation — headroom costs nothing unless it is used.
		MaxTokens:   16000,
		Effort:      "high", // critique is reasoning-heavy; budget pays off here
		CacheSystem: true,
	})
}

// parseJSON pulls the first JSON object/array out of a model reply, tolerating
// ```json fences and surrounding prose.
func parseJSON(text string, v any) error {
	t := strings.TrimSpace(text)
	t = strings.TrimPrefix(t, "```json")
	t = strings.TrimPrefix(t, "```")
	t = strings.TrimSuffix(strings.TrimSpace(t), "```")
	t = strings.TrimSpace(t)
	start := -1
	for i, r := range t {
		if r == '{' || r == '[' {
			start = i
			break
		}
	}
	if start == -1 {
		return fmt.Errorf("no JSON found in: %.200s", t)
	}
	opener := t[start]
	closer := byte('}')
	if opener == '[' {
		closer = ']'
	}
	depth := 0
	for i := start; i < len(t); i++ {
		switch t[i] {
		case opener:
			depth++
		case closer:
			depth--
			if depth == 0 {
				return json.Unmarshal([]byte(t[start:i+1]), v)
			}
		}
	}
	return errors.New("unbalanced JSON in response")
}

func fill(tmpl string, kv ...string) string {
	return strings.NewReplacer(kv...).Replace(tmpl)
}

// audienceResult is the full parsed shape of step 1. RunMethod only needs the
// first two fields, but the public /audience-check endpoint surfaces the
// alternatives — that's the whole point of the free taster.
type audienceResult struct {
	CarriedAudience  string `json:"carried_audience"`
	WillingnessToPay string `json:"willingness_to_pay"`
	Alternatives     []struct {
		Audience string `json:"audience"`
		Why      string `json:"why"`
	} `json:"alternatives"`
}

// runAudience runs just step 1 of the method. Used both internally by RunMethod
// (the carried audience feeds step 2) and externally by the /audience-check
// public endpoint (the free taster on the page).
func runAudience(business, audience, assets string) (audienceResult, error) {
	var aud audienceResult
	out, err := callClaudeOpts(callOpts{
		Model:  genModel,
		System: systemBase,
		User:   fill(audiencePrompt, "{domain}", business, "{audience}", audience, "{assets}", assets),
		// Was 4096 with no Effort, i.e. no thinking at all. On the 5 family that
		// omission would silently mean adaptive thinking sharing this cap with
		// the answer — and this step returns JSON, so a truncated answer is a
		// parse error, not a short report. Explicit low effort keeps the step
		// cheap (it is also the free taster, ~£0.02 a call) and the raised cap
		// leaves room for thinking + the JSON.
		MaxTokens:   8000,
		Effort:      "low",
		CacheSystem: true, // first call in a run; subsequent steps will hit the cache
	})
	if err != nil {
		return aud, fmt.Errorf("audience step: %w", err)
	}
	if err := parseJSON(out, &aud); err != nil {
		return aud, fmt.Errorf("audience parse: %w", err)
	}
	return aud, nil
}

// runAssess runs step 0 — the web-verified assessment of the submitted idea
// itself. Same model + tool posture as the verify step (it is the same kind of
// work: multi-step inference over search results), and the promise-critical
// call, so it gets the search budget and xhigh effort.
func runAssess(submission, audience, notes string) (assessment, error) {
	var a assessment
	out, err := callClaudeOpts(callOpts{
		Model:  verifyModel,
		System: systemBase,
		User: fill(assessPrompt, "{submission}", submission, "{audience}", audience,
			"{notes}", notes),
		Tools: []map[string]any{
			{"type": "web_search_20260209", "name": "web_search", "max_uses": envInt("WEB_SEARCH_MAX_USES", 12)},
		},
		MaxTokens:   32000,
		Effort:      "xhigh",
		CacheSystem: true, // first call in a run; establishes the cache steps 1-5 read
	})
	if err != nil {
		return a, fmt.Errorf("assess step: %w", err)
	}
	if err := parseJSON(out, &a); err != nil {
		return a, fmt.Errorf("assess parse: %w", err)
	}
	return a, nil
}

// ── RunMethod: the pipeline ──────────────────────────────────────────────────
func RunMethod(domain, audience, assets string) (renderedReport, error) {
	// STEP 0 — assess the submitted idea itself (the headline promise of the
	// report page: problem/evidence, who else, substitutes, defensible/exposed,
	// a next step — with checkable sources). Hard-fail like every other step:
	// a report without its headline section should not be sent, and the
	// operator flow can re-run.
	assess, err := runAssess(domain, audience, assets)
	if err != nil {
		return renderedReport{}, err
	}

	// STEP 1 — audience framing + challenge
	aud, err := runAudience(domain, audience, assets)
	if err != nil {
		return renderedReport{}, err
	}

	// STEP 2 — generate (multi-lens). CacheSystem true; reads the warm cache
	// that runAudience just established (5-min TTL — fits inside a full run).
	s2, err := callClaudeOpts(callOpts{
		Model:  genModel,
		System: systemBase,
		User: fill(generatePrompt, "{domain}", domain, "{audience}", aud.CarriedAudience,
			"{wtp}", aud.WillingnessToPay, "{assets}", assets, "{capabilities}", capabilityMenu),
		// Same reasoning as the audience step: was 8000 with no Effort. Explicit
		// low effort + headroom now that thinking shares the cap.
		MaxTokens:   16000,
		Effort:      "low",
		CacheSystem: true,
	})
	if err != nil {
		return renderedReport{}, fmt.Errorf("generate step: %w", err)
	}
	var gen struct {
		Candidates []candidate `json:"candidates"`
	}
	if err := parseJSON(s2, &gen); err != nil {
		return renderedReport{}, fmt.Errorf("generate parse: %w", err)
	}
	for i := range gen.Candidates {
		gen.Candidates[i].ID = fmt.Sprintf("c%d", i+1) // stable ids, threaded by id not title
	}

	// STEP 3 — cut (cross-vendor if configured)
	candJSON, _ := json.MarshalIndent(gen.Candidates, "", "  ")
	s3, err := critique(fill(cutPrompt, "{candidates_json}", string(candJSON)))
	if err != nil {
		return renderedReport{}, fmt.Errorf("cut step: %w", err)
	}
	var cut struct {
		Results []struct {
			ID      string `json:"id"`
			Verdict string `json:"verdict"`
		} `json:"results"`
	}
	if err := parseJSON(s3, &cut); err != nil {
		return renderedReport{}, fmt.Errorf("cut parse: %w", err)
	}
	keep := map[string]bool{}
	for _, r := range cut.Results {
		if r.Verdict == "keep" && r.ID != "" {
			keep[r.ID] = true
		}
	}
	var survivors []candidate
	for _, c := range gen.Candidates {
		if keep[c.ID] {
			survivors = append(survivors, c)
		}
	}
	if len(survivors) == 0 {
		return renderReport(domain, aud.CarriedAudience, aud.WillingnessToPay, assess, nil, nil, nil,
			"No candidate survived the cut."), nil
	}

	// STEP 4 — verify (web_search v2, with dynamic filtering).
	// web_search_20260209 auto-injects a code_execution tool behind the scenes
	// to filter results before they hit the context window — Anthropic's own
	// benchmarks: ~11% accuracy lift on BrowseComp/DeepsearchQA, ~24% fewer
	// input tokens, and that filtering code execution is free. We must NOT
	// declare code_execution ourselves — doing so collides with the
	// auto-injected one ("tool names must be unique", a 400). Extended thinking
	// is on — verify is the most reasoning-heavy step (multi-step inference over
	// a stream of search hits).
	survJSON, _ := json.MarshalIndent(survivors, "", "  ")
	s4, err := callClaudeOpts(callOpts{
		Model:  verifyModel,
		System: systemBase,
		User:   fill(verifyPrompt, "{domain}", domain, "{survivors_json}", string(survJSON)),
		Tools: []map[string]any{
			{"type": "web_search_20260209", "name": "web_search", "max_uses": envInt("WEB_SEARCH_MAX_USES", 12)},
		},
		MaxTokens:   32000,   // room for adaptive thinking + search-result text + output
		Effort:      "xhigh", // Anthropic recommends xhigh for agentic work with repeated search
		CacheSystem: true,
	})
	if err != nil {
		return renderedReport{}, fmt.Errorf("verify step: %w", err)
	}
	var ver struct {
		Results []struct {
			ID           string   `json:"id"`
			Findings     string   `json:"findings"`
			PremiseHolds bool     `json:"premise_holds"`
			Sources      []source `json:"sources"`
		} `json:"results"`
	}
	if err := parseJSON(s4, &ver); err != nil {
		return renderedReport{}, fmt.Errorf("verify parse: %w", err)
	}
	type verdict struct {
		findings string
		sources  []source
	}
	holds := map[string]verdict{}
	for _, r := range ver.Results {
		if r.PremiseHolds && r.ID != "" {
			holds[r.ID] = verdict{r.Findings, r.Sources}
		}
	}
	var verified []candidate
	for _, c := range survivors {
		if v, ok := holds[c.ID]; ok {
			c.Findings = v.findings
			c.Sources = v.sources
			verified = append(verified, c)
		}
	}
	if len(verified) == 0 {
		return renderReport(domain, aud.CarriedAudience, aud.WillingnessToPay, assess, nil, nil, nil,
			"No premise survived verification."), nil
	}

	// STEP 5 — score. Small thinking budget supports careful rubric application
	// (the Risk column especially benefits from a moment of deliberation about
	// "consequence of being wrong"). Sonnet 4.6 stays as the model — different
	// from the generator, so the method isn't marking its own work.
	verJSON, _ := json.MarshalIndent(verified, "", "  ")
	s5, err := callClaudeOpts(callOpts{
		Model:       scoreModel,
		System:      systemBase,
		User:        fill(scorePrompt, "{verified_json}", string(verJSON)),
		MaxTokens:   12000, // +Sonnet 5 tokenizer headroom (see the critique step)
		Effort:      "medium",
		CacheSystem: true,
	})
	if err != nil {
		return renderedReport{}, fmt.Errorf("score step: %w", err)
	}
	var sc struct {
		Scored []scored `json:"scored"`
	}
	if err := parseJSON(s5, &sc); err != nil {
		return renderedReport{}, fmt.Errorf("score parse: %w", err)
	}
	byID := map[string]candidate{}
	for _, c := range verified {
		byID[c.ID] = c
	}
	var advancing, dropped, riskDropped []scored
	for _, s := range sc.Scored {
		base, ok := byID[s.ID] // merge by id; skip unmatched (no blank rows)
		if !ok {
			continue
		}
		s.candidate = base
		// Risk=1 = regulated profession territory. Don't recommend, even if it
		// would have advanced — surface separately so the operator sees what got
		// killed for risk vs what failed the Def/Will gate.
		if s.Risk == 1 {
			fmt.Fprintf(os.Stderr, "[score] dropping %s for risk=1 (regulated-profession territory)\n", s.Title)
			riskDropped = append(riskDropped, s)
			continue
		}
		// Risk ≤ 2 = "needs liability work before building" — still advances if it
		// passes the Def/Will gate, but flagged in the report.
		s.NeedsLiabilityWork = s.Risk <= 2
		if s.Advances {
			advancing = append(advancing, s)
		} else {
			dropped = append(dropped, s)
		}
	}
	// Rank advancing by sum desc; tiebreak by Risk desc (prefer safer builds
	// when fitness is equal — Risk is a tiebreaker, not a heavy weight).
	for i := 0; i < len(advancing); i++ {
		for j := i + 1; j < len(advancing); j++ {
			if advancing[j].Sum > advancing[i].Sum ||
				(advancing[j].Sum == advancing[i].Sum && advancing[j].Risk > advancing[i].Risk) {
				advancing[i], advancing[j] = advancing[j], advancing[i]
			}
		}
	}
	return renderReport(domain, aud.CarriedAudience, aud.WillingnessToPay, assess, advancing, dropped, riskDropped, ""), nil
}

type renderedReport struct {
	Text string // plain-text report (email body + the JSON record)
	HTML string // styled HTML report (the multipart HTML email)
}

// renderReport produces both renderings from the same structured result, so the
// HTML version is built from the data — not re-parsed from the text.
func renderReport(domain, audience, wtp string, assess assessment, advancing, dropped, riskDropped []scored, note string) renderedReport {
	return renderedReport{
		Text: render(domain, audience, wtp, assess, advancing, dropped, riskDropped, note),
		HTML: renderHTML(domain, audience, wtp, assess, advancing, dropped, riskDropped, note),
	}
}

// renderHTML is the styled HTML version of the report for the HTML email. Inline
// styles only (many mail clients drop <style> blocks). All model/user text is
// HTML-escaped.
// renderHTML is the styled HTML version of the report for the HTML email. Inline
// styles only (many mail clients drop <style> blocks). All model/user text is
// HTML-escaped. Its own professional palette + type — deliberately not the
// landing-page brand — so the report reads like a considered, paid-for document.
func renderHTML(domain, audience, wtp string, assess assessment, advancing, dropped, riskDropped []scored, note string) string {
	esc := html.EscapeString
	const (
		navy  = "#15243d" // headings, idea titles, labels
		slate = "#36424f" // body text
		muted = "#707b88" // secondary text
		gold  = "#b08a3e" // restrained accent: rule, badge, card edge
		line  = "#e2e6ec" // hairlines / borders
		card  = "#fbfbfc" // idea-card background
		serif = "Georgia,'Times New Roman',serif"
		sans  = "'Helvetica Neue',Helvetica,Arial,sans-serif"
	)
	var b strings.Builder
	fmt.Fprintf(&b, `<div style="background:#eceff3;padding:24px 12px;font-family:%s;font-size:16px;line-height:1.7;color:%s">`, sans, slate)
	fmt.Fprintf(&b, `<div style="max-width:640px;margin:0 auto;background:#ffffff;border:1px solid %s;border-radius:10px">`, line)
	// header band: wordmark + a short gold rule
	fmt.Fprintf(&b, `<div style="padding:22px 34px 0 34px"><div style="font-family:%s;font-size:20px;font-weight:bold;color:%s;letter-spacing:.2px">idea<span style="color:%s">.</span>uk</div><div style="height:3px;width:46px;background:%s;margin-top:10px"></div></div>`, serif, navy, gold, gold)
	b.WriteString(`<div style="padding:14px 34px 32px 34px">`)
	fmt.Fprintf(&b, `<h1 style="font-family:%s;font-size:27px;line-height:1.2;color:%s;margin:14px 0 8px;font-weight:bold">Your idea report</h1>`, serif, navy)
	fmt.Fprintf(&b, `<p style="margin:0 0 22px;color:%s">%s</p>`, slate, esc(reportIntro(domain)))
	sect := func(t string) {
		fmt.Fprintf(&b, `<div style="font-size:12px;font-weight:bold;letter-spacing:.12em;text-transform:uppercase;color:%s;border-top:1px solid %s;padding-top:18px;margin:26px 0 8px">%s</div>`, navy, line, esc(t))
	}
	// srcList renders a "Check it yourself" source list — the concrete delivery
	// of the report page's "we explain its source so you can check it yourself".
	srcList := func(srcs []source) {
		if len(srcs) == 0 {
			return
		}
		fmt.Fprintf(&b, `<p style="font-size:13px;color:%s;margin:8px 0 0"><span style="font-weight:bold;color:%s">Check it yourself:</span></p><ul style="font-size:13px;color:%s;margin:2px 0 0;padding-left:18px">`, muted, navy, muted)
		for _, s := range srcs {
			t := s.Title
			if t == "" {
				t = s.URL
			}
			fmt.Fprintf(&b, `<li style="margin:2px 0"><a href="%s" style="color:%s">%s</a></li>`, esc(s.URL), navy, esc(t))
		}
		b.WriteString(`</ul>`)
	}
	// ── Part 1: the submitted idea, assessed ────────────────────────────────
	arow := func(label, val string) {
		if val == "" {
			return
		}
		fmt.Fprintf(&b, `<p style="margin:0 0 10px;color:%s"><span style="color:%s;font-weight:bold">%s</span> %s</p>`, slate, navy, esc(label), esc(val))
	}
	sect("Your idea, assessed")
	if assess.Reading != "" {
		fmt.Fprintf(&b, `<p style="margin:0 0 12px;color:%s"><em>%s</em></p>`, muted, esc(assess.Reading))
	}
	if !assess.IsAssessable {
		fmt.Fprintf(&b, `<p style="background:#f6f1e6;border-left:3px solid %s;padding:12px 16px;margin:0 0 12px;color:%s">What you sent us is too early to assess honestly — it reads as an area of interest rather than a worked-out proposition, and padding it into a verdict would not serve you. The note above says what is missing; the free tools on the site are the right next step, and the rest of this report looks at directions worth considering around it.</p>`, gold, slate)
	} else {
		arow("The problem, and the evidence people have it:", assess.Problem)
		arow("Signs of real demand:", assess.DemandEvidence)
		arow("Who else is addressing it, and how:", assess.WhoElse)
		arow("What people would use instead today:", assess.SubstitutesToday)
		arow("Where it is defensible:", assess.Defensible)
		arow("Where it is exposed:", assess.Exposed)
		arow("A considered next step:", assess.NextStep)
	}
	srcList(assess.Sources)
	sect("Who it's for")
	fmt.Fprintf(&b, `<p style="margin:0;color:%s">%s</p>`, slate, esc(audience))
	sect("Why they'd pay")
	fmt.Fprintf(&b, `<p style="margin:0;color:%s">%s</p>`, slate, esc(wtp))
	if note != "" {
		fmt.Fprintf(&b, `<p style="background:#f6f1e6;border-left:3px solid %s;padding:12px 16px;margin:16px 0;color:%s">%s</p>`, gold, slate, esc(note))
	}
	if len(advancing) == 0 {
		sect("No further idea cleared the bar")
		fmt.Fprintf(&b, `<p style="margin:0;color:%s">Beyond the assessment above, none of the further ideas we generated passed both of our main tests — being hard for someone else to copy, and being something enough people would pay for (each needs at least 3 out of 5 on both). That is a real result, not a dead end: it usually means a different audience, a different starting asset, or a different way to charge.</p>`, slate)
	} else {
		sect("Further ideas worth pursuing (best first)")
		for i, x := range advancing {
			fmt.Fprintf(&b, `<div style="background:%s;border:1px solid %s;border-left:3px solid %s;border-radius:6px;padding:16px 18px;margin:0 0 14px">`, card, line, gold)
			fmt.Fprintf(&b, `<div style="font-family:%s;font-size:19px;font-weight:bold;color:%s;margin:0 0 4px">%d. %s</div>`, serif, navy, i+1, esc(x.Title))
			fmt.Fprintf(&b, `<div style="font-size:11px;font-weight:bold;letter-spacing:.08em;text-transform:uppercase;color:%s;margin:0 0 10px">%s</div>`, gold, esc(flagLabel(x.Flag)))
			if x.ShortLived {
				fmt.Fprintf(&b, `<p style="font-size:14px;color:%s;margin:0 0 6px"><em>Heads-up: this one may not last — as the general AI models improve, they may do it too.</em></p>`, muted)
			}
			if x.NeedsLiabilityWork {
				fmt.Fprintf(&b, `<p style="font-size:14px;color:%s;margin:0 0 6px"><em>Heads-up: needs legal and insurance groundwork before building (risk %d/5 — see below).</em></p>`, muted, x.Risk)
			}
			fmt.Fprintf(&b, `<p style="margin:0 0 10px;color:%s">%s</p>`, slate, esc(x.BeatsFreeBecause))
			row := func(label, val string) {
				fmt.Fprintf(&b, `<p style="margin:0 0 6px;color:%s"><span style="color:%s;font-weight:bold">%s</span> %s</p>`, slate, navy, esc(label), esc(val))
			}
			row("What it's built on:", sentence(x.Asset+", using "+midSentence(x.Capability)))
			row("What we found:", x.Findings)
			fmt.Fprintf(&b, `<p style="margin:0 0 6px;color:%s"><span style="color:%s;font-weight:bold">How it scored</span> (each out of 5): hard to copy %d &middot; people will pay %d &middot; easy to build %d &middot; reusable elsewhere %d &middot; built to last %d <span style="color:%s">(%d out of 25 overall)</span></p>`, slate, navy, x.Defensibility, x.Willingness, x.Buildability, x.Reuse, x.Durability, muted, x.Sum)
			fmt.Fprintf(&b, `<p style="margin:0 0 6px;color:%s"><span style="color:%s;font-weight:bold">Risk to you:</span> %d/5 %s</p>`, slate, navy, x.Risk, esc(riskNote(x.Risk)))
			row("A cheap first test:", x.CheapestTest)
			srcList(x.Sources)
			b.WriteString(`</div>`)
		}
	}
	if len(dropped) > 0 {
		sect("Didn't make the cut")
		fmt.Fprintf(&b, `<p style="margin:0 0 10px;color:%s">We came up with these too, but set them aside — either because they'd be too easy for someone else to copy, or because not enough people would pay for them.</p>`, slate)
		for _, x := range dropped {
			fmt.Fprintf(&b, `<div style="margin:0 0 12px"><div style="font-weight:bold;color:%s">%s</div>`, navy, esc(x.Title))
			if x.BeatsFreeBecause != "" {
				fmt.Fprintf(&b, `<p style="margin:2px 0 2px;color:%s">%s</p>`, slate, esc(x.BeatsFreeBecause))
			}
			fmt.Fprintf(&b, `<p style="margin:2px 0 0;color:%s;font-size:15px">%s</p></div>`, muted, esc(dropReason(x)))
		}
	}
	if len(riskDropped) > 0 {
		sect("Set aside on risk")
		fmt.Fprintf(&b, `<p style="margin:0 0 10px;color:%s">These could be real opportunities, but they sit in regulated or high-stakes territory — things like medical, legal, or financial advice. To offer them safely we'd need the right professional qualifications, insurance, and legal cover, so we're flagging them for your awareness rather than recommending them.</p>`, slate)
		for _, x := range riskDropped {
			fmt.Fprintf(&b, `<div style="margin:0 0 12px"><div style="font-weight:bold;color:%s">%s</div>`, navy, esc(x.Title))
			if x.BeatsFreeBecause != "" {
				fmt.Fprintf(&b, `<p style="margin:2px 0 2px;color:%s">%s</p>`, slate, esc(x.BeatsFreeBecause))
			}
			fmt.Fprintf(&b, `<p style="margin:2px 0 0;color:%s;font-size:15px">Set aside because it falls in regulated territory (risk %d/5).</p></div>`, muted, x.Risk)
		}
	}
	addr := reportContact()
	fmt.Fprintf(&b, `<div style="background:#f6f1e6;border:1px solid %s;border-radius:6px;padding:16px 18px;margin:26px 0 12px;color:%s"><span style="font-weight:bold;color:%s">Want to take one of these further?</span><br>If you'd like us to help turn any of these ideas into a working tool — or you have any questions about this report — just email us at <a href="mailto:%s" style="color:%s;font-weight:bold;text-decoration:none">%s</a>.</div>`, line, slate, navy, esc(addr), navy, esc(addr))
	fmt.Fprintf(&b, `<div style="border-top:1px solid %s;margin-top:8px;padding-top:16px;color:%s;font-size:14px">%s</div>`, line, muted, esc(reportFooter()))
	b.WriteString(`</div></div></div>`)
	return b.String()
}

const reportRule = "------------------------------------------------------------"

// reportIntro is the plain-English opener so a recipient with a full inbox knows
// what this email is and what it contains. domain is the business-or-idea text
// the customer submitted. It also carries the report's own AI disclosure — the
// page promises AI use is "clearly indicated", and the T&Cs alone are not the
// report the customer actually reads.

// sentence ends a fragment with exactly one full stop. Customer- and
// model-supplied text is spliced into our prose, and it may or may not arrive
// punctuated: the live run of 2026-07-26 rendered "…finding out at the
// counter.. First we assess", because the submitted description already ended
// in a full stop and we appended another.
func sentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasSuffix(s, ".") || strings.HasSuffix(s, "!") || strings.HasSuffix(s, "?") {
		return s
	}
	return s + "."
}

// midSentence lowercases a fragment's first letter so it reads correctly when
// spliced into the middle of one of our sentences. Model-generated fields
// arrive sentence-cased — the same live run produced "…, using A form the
// receptionist fills in about the pet". Acronyms and proper nouns are left
// alone: only a first word that is Capitalised-then-lowercase is touched.
func midSentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	r := []rune(s)
	if !unicode.IsUpper(r[0]) {
		return s
	}
	// "AI", "UK", "CMA" — leave anything whose second rune is also upper.
	if len(r) > 1 && unicode.IsUpper(r[1]) {
		return s
	}
	// A capitalised word that is a known proper noun would be wrong to lower,
	// but we cannot tell; the fields this is used on are descriptions ("A form
	// the receptionist fills in"), so the sentence-case reading is the safe one.
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
func reportIntro(domain string) string {
	return "This report is from idea.uk, about what you sent us: " + sentence(domain) + " " +
		"First we assess the idea you submitted — the problem it addresses, the evidence of demand, " +
		"who else is out there, where it is defensible and where it is exposed, and a specific next " +
		"step. Then we go looking for further ideas around it, check each against what already exists " +
		"and whether people would actually pay, and set out the ones worth pursuing — and the ones we " +
		"set aside, and why. We use AI to research and draft this report, with live web searches for " +
		"the checking; a person reviews it before it is sent. Where a finding rests on something we " +
		"read, the sources are listed under it so you can check them yourself."
}

// reportContact is the address the report tells the reader to write to.
//
// It is set ONCE at startup from config so OPERATOR_EMAIL is the single source
// of truth: previously this read CONTACT_EMAIL straight from os.Getenv with a
// hardcoded fallback, bypassing both the config and the ContactEmail →
// OperatorEmail fallback the rest of the service uses (contactEmail(),
// service.go). That meant the report could address a different mailbox from
// every other email the service sends, and the duplicate CONTACT_EMAIL line on
// the box (2026-07-26) made which one win a matter of file ordering.
var reportContactAddr = "idea-uk@leopardess.uk" // last-resort default; overwritten by SetReportContact

// SetReportContact wires the resolved address in at startup. Called from NewApp
// so both the report and the rest of the service agree.
func SetReportContact(addr string) {
	if addr != "" {
		reportContactAddr = addr
	}
}

func reportContact() string {
	return reportContactAddr
}

// reportCTA invites the reader to hire us to build one of the ideas.
func reportCTA() string {
	return "If you'd like us to help turn any of these ideas into a working tool — or you have any " +
		"questions about this report — just email us at " + reportContact() + "."
}

func reportFooter() string {
	return "idea.uk finds and tests AI product ideas for your business, so you can spend your time " +
		"building the ones most likely to pay off."
}

// flagLabel turns the internal flag into plain words for the reader.
func flagLabel(f string) string {
	switch f {
	case "test_now":
		return "worth testing now"
	case "consider":
		return "worth considering"
	default:
		return f
	}
}

// dropReason explains in plain words why an idea didn't make the cut, from its scores.
func dropReason(x scored) string {
	var why []string
	if x.Defensibility < 3 {
		why = append(why, fmt.Sprintf("it would be too easy for someone else to copy (hard-to-copy %d/5)", x.Defensibility))
	}
	if x.Willingness < 3 {
		why = append(why, fmt.Sprintf("not enough people would pay for it (people-will-pay %d/5)", x.Willingness))
	}
	if len(why) == 0 {
		return "We set it aside because it didn't clear our bar on the main tests."
	}
	return "We set it aside because " + strings.Join(why, ", and ") + "."
}

func render(domain, audience, wtp string, assess assessment, advancing, dropped, riskDropped []scored, note string) string {
	var b strings.Builder
	// srcLines mirrors the HTML "Check it yourself" list for the plain-text part.
	srcLines := func(prefix string, srcs []source) {
		if len(srcs) == 0 {
			return
		}
		fmt.Fprintf(&b, "%sCheck it yourself:\n", prefix)
		for _, s := range srcs {
			t := s.Title
			if t == "" {
				t = s.URL
			}
			fmt.Fprintf(&b, "%s - %s — %s\n", prefix, t, s.URL)
		}
	}
	fmt.Fprintf(&b, "IDEA REPORT — %s\n%s\n\n", domain, reportRule)
	fmt.Fprintf(&b, "%s\n\n", reportIntro(domain))
	b.WriteString("YOUR IDEA, ASSESSED\n")
	if assess.Reading != "" {
		fmt.Fprintf(&b, "(%s)\n\n", assess.Reading)
	}
	if !assess.IsAssessable {
		b.WriteString("What you sent us is too early to assess honestly — it reads as an area of\n" +
			"interest rather than a worked-out proposition, and padding it into a verdict\n" +
			"would not serve you. The note above says what is missing; the free tools on the\n" +
			"site are the right next step, and the rest of this report looks at directions\n" +
			"worth considering around it.\n")
	} else {
		arow := func(label, val string) {
			if val != "" {
				fmt.Fprintf(&b, "%s\n   %s\n", label, val)
			}
		}
		arow("The problem, and the evidence people have it:", assess.Problem)
		arow("Signs of real demand:", assess.DemandEvidence)
		arow("Who else is addressing it, and how:", assess.WhoElse)
		arow("What people would use instead today:", assess.SubstitutesToday)
		arow("Where it is defensible:", assess.Defensible)
		arow("Where it is exposed:", assess.Exposed)
		arow("A considered next step:", assess.NextStep)
	}
	srcLines("", assess.Sources)
	b.WriteString("\n")
	fmt.Fprintf(&b, "WHO IT'S FOR\n%s\n\n", audience)
	fmt.Fprintf(&b, "WHY THEY'D PAY\n%s\n\n", wtp)
	if note != "" {
		fmt.Fprintf(&b, "%s\n\n", note)
	}
	if len(advancing) == 0 {
		b.WriteString("NO FURTHER IDEA CLEARED THE BAR\n")
		b.WriteString("Beyond the assessment above, none of the further ideas we generated passed both\n" +
			"of our main tests: being hard for someone else to copy, and being something\n" +
			"enough people would pay for (each needs at least 3 out of 5 on both). That is a\n" +
			"real result, not a dead end — it usually means a different audience, a different\n" +
			"starting asset, or a different way to charge.\n\n")
	} else {
		b.WriteString("FURTHER IDEAS WORTH PURSUING (best first)\n\n")
		for i, x := range advancing {
			fmt.Fprintf(&b, "%d) %s  [%s]\n", i+1, x.Title, flagLabel(x.Flag))
			if x.ShortLived {
				b.WriteString("   Heads-up: this one may not last — as the general AI models improve, they may do it too.\n")
			}
			if x.NeedsLiabilityWork {
				fmt.Fprintf(&b, "   Heads-up: needs legal and insurance groundwork before building (risk %d/5 — see below).\n", x.Risk)
			}
			fmt.Fprintf(&b, "   %s\n", x.BeatsFreeBecause)
			fmt.Fprintf(&b, "   What it's built on:  %s\n", sentence(x.Asset+", using "+midSentence(x.Capability)))
			fmt.Fprintf(&b, "   What we found:       %s\n", x.Findings)
			fmt.Fprintf(&b, "   How it scored:       (each out of 5) hard to copy %d, people will pay %d, easy to build %d, reusable elsewhere %d, built to last %d (%d out of 25 overall).\n",
				x.Defensibility, x.Willingness, x.Buildability, x.Reuse, x.Durability, x.Sum)
			fmt.Fprintf(&b, "   Risk to you:         %d/5 %s\n", x.Risk, riskNote(x.Risk))
			fmt.Fprintf(&b, "   A cheap first test:  %s\n", x.CheapestTest)
			srcLines("   ", x.Sources)
			b.WriteString("\n")
		}
	}
	if len(dropped) > 0 {
		fmt.Fprintf(&b, "%s\nDIDN'T MAKE THE CUT\n", reportRule)
		b.WriteString("We came up with these too, but set them aside — either because they'd be too\n" +
			"easy for someone else to copy, or because not enough people would pay for them.\n\n")
		for _, x := range dropped {
			fmt.Fprintf(&b, "   %s\n", x.Title)
			if x.BeatsFreeBecause != "" {
				fmt.Fprintf(&b, "   %s\n", x.BeatsFreeBecause)
			}
			fmt.Fprintf(&b, "   %s\n\n", dropReason(x))
		}
	}
	if len(riskDropped) > 0 {
		fmt.Fprintf(&b, "%s\nSET ASIDE ON RISK\n", reportRule)
		b.WriteString("These could be real opportunities, but they sit in regulated or high-stakes\n" +
			"territory — things like medical, legal, or financial advice. To offer them safely\n" +
			"we'd need the right professional qualifications, insurance, and legal cover, so\n" +
			"we're flagging them for your awareness rather than recommending them.\n\n")
		for _, x := range riskDropped {
			fmt.Fprintf(&b, "   %s\n", x.Title)
			if x.BeatsFreeBecause != "" {
				fmt.Fprintf(&b, "   %s\n", x.BeatsFreeBecause)
			}
			fmt.Fprintf(&b, "   Set aside because it falls in regulated territory (risk %d/5).\n\n", x.Risk)
		}
	}
	fmt.Fprintf(&b, "%s\n%s\n\n%s\n", reportRule, reportCTA(), reportFooter())
	return b.String()
}

// riskNote returns the short label shown next to the Risk score in the report.
func riskNote(r int) string {
	switch r {
	case 5:
		return "(pure analysis; customer decides)"
	case 4:
		return "(low — a mistake would be minor, and a refund would put it right)"
	case 3:
		return "(moderate — show your sources; insurance for handling personal data is recommended)"
	case 2:
		return "(high — needs a person to check every report, insurance, and carefully checked terms before building)"
	case 1:
		return "(regulated territory — should not build without proper qualifications)"
	default:
		return ""
	}
}
