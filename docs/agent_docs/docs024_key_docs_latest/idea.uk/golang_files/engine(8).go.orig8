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
)

// Models. The cut runs on a DIFFERENT vendor if OPENAI_API_KEY is set
// (genuine cross-vendor critique), else a different Anthropic model.
// Strings verified against docs.claude.com (May 2026); override via env.
var (
	genModel      = env("GEN_MODEL", "claude-opus-4-8")
	critiqueModel = env("CRITIQUE_MODEL", "claude-sonnet-4-6")
	verifyModel   = env("VERIFY_MODEL", "claude-opus-4-8")
	scoreModel    = env("SCORE_MODEL", "claude-sonnet-4-6")
	openAIModel   = env("OPENAI_CRITIQUE_MODEL", "gpt-4o")
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
type candidate struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Lens             string `json:"lens"`
	Asset            string `json:"asset"`
	Capability       string `json:"capability"`
	BeatsFreeBecause string `json:"beats_free_because"`
	Findings         string `json:"findings,omitempty"`
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
//   - Effort "" disables thinking. Otherwise one of low|medium|high|xhigh|max.
//     The wire format DIFFERS BY MODEL (this bit us — see debugging guide §0):
//     Opus 4.7/4.8 and Mythos use adaptive thinking (thinking:{type:adaptive}
//   - output_config:{effort}); they 400 on manual budgets. Sonnet 4.6 and
//     older use manual extended thinking (thinking:{type:enabled,budget_tokens}).
//     callClaudeOpts picks the right one based on the model string.
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

// usesAdaptiveThinking reports whether a model takes adaptive thinking +
// output_config.effort (true) rather than manual thinking budgets (false).
// Opus 4.7, Opus 4.8, and Mythos dropped manual budgets; Sonnet 4.6 and older
// still take them. Matching on substrings keeps this working across the
// dated/aliased forms of the same model id.
func usesAdaptiveThinking(model string) bool {
	return strings.Contains(model, "opus-4-7") ||
		strings.Contains(model, "opus-4-8") ||
		strings.Contains(model, "mythos")
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
	if o.Effort != "" {
		if usesAdaptiveThinking(o.Model) {
			// Opus 4.7/4.8/Mythos: adaptive thinking + effort. Manual budgets 400.
			body["thinking"] = map[string]any{"type": "adaptive"}
			body["output_config"] = map[string]any{"effort": o.Effort}
		} else {
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
		Usage struct {
			InputTokens              int `json:"input_tokens"`
			OutputTokens             int `json:"output_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
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

// callClaude is the simple form. New code should use callClaudeOpts directly
// when it needs thinking or caching; leaving this in keeps every existing call
// site unchanged.
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
		Model:       critiqueModel,
		System:      systemBase,
		User:        user,
		MaxTokens:   12000,
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
		Model:       genModel,
		System:      systemBase,
		User:        fill(audiencePrompt, "{domain}", business, "{audience}", audience, "{assets}", assets),
		MaxTokens:   4096,
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

// ── RunMethod: the pipeline ──────────────────────────────────────────────────
func RunMethod(domain, audience, assets string) (renderedReport, error) {
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
		MaxTokens:   8000,
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
		return renderReport(domain, aud.CarriedAudience, aud.WillingnessToPay, nil, nil, nil,
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
			ID           string `json:"id"`
			Findings     string `json:"findings"`
			PremiseHolds bool   `json:"premise_holds"`
		} `json:"results"`
	}
	if err := parseJSON(s4, &ver); err != nil {
		return renderedReport{}, fmt.Errorf("verify parse: %w", err)
	}
	holds := map[string]string{}
	for _, r := range ver.Results {
		if r.PremiseHolds && r.ID != "" {
			holds[r.ID] = r.Findings
		}
	}
	var verified []candidate
	for _, c := range survivors {
		if f, ok := holds[c.ID]; ok {
			c.Findings = f
			verified = append(verified, c)
		}
	}
	if len(verified) == 0 {
		return renderReport(domain, aud.CarriedAudience, aud.WillingnessToPay, nil, nil, nil,
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
		MaxTokens:   10000,
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
	return renderReport(domain, aud.CarriedAudience, aud.WillingnessToPay, advancing, dropped, riskDropped, ""), nil
}

type renderedReport struct {
	Text string // plain-text report (email body + the JSON record)
	HTML string // styled HTML report (the multipart HTML email)
}

// renderReport produces both renderings from the same structured result, so the
// HTML version is built from the data — not re-parsed from the text.
func renderReport(domain, audience, wtp string, advancing, dropped, riskDropped []scored, note string) renderedReport {
	return renderedReport{
		Text: render(domain, audience, wtp, advancing, dropped, riskDropped, note),
		HTML: renderHTML(domain, audience, wtp, advancing, dropped, riskDropped, note),
	}
}

// renderHTML is the styled HTML version of the report for the HTML email. Inline
// styles only (many mail clients drop <style> blocks). All model/user text is
// HTML-escaped.
// renderHTML is the styled HTML version of the report for the HTML email. Inline
// styles only (many mail clients drop <style> blocks). All model/user text is
// HTML-escaped. Its own professional palette + type — deliberately not the
// landing-page brand — so the report reads like a considered, paid-for document.
func renderHTML(domain, audience, wtp string, advancing, dropped, riskDropped []scored, note string) string {
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
	sect("Who it's for")
	fmt.Fprintf(&b, `<p style="margin:0;color:%s">%s</p>`, slate, esc(audience))
	sect("Why they'd pay")
	fmt.Fprintf(&b, `<p style="margin:0;color:%s">%s</p>`, slate, esc(wtp))
	if note != "" {
		fmt.Fprintf(&b, `<p style="background:#f6f1e6;border-left:3px solid %s;padding:12px 16px;margin:16px 0;color:%s">%s</p>`, gold, slate, esc(note))
	}
	if len(advancing) == 0 {
		sect("No idea cleared the bar")
		fmt.Fprintf(&b, `<p style="margin:0;color:%s">None of the ideas passed both of our main tests — being hard for someone else to copy, and being something enough people would pay for (each needs at least 3 out of 5 on both). That is a real result, not a dead end: it usually means a different audience, a different starting asset, or a different way to charge.</p>`, slate)
	} else {
		sect("Ideas worth pursuing (best first)")
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
			row("What it's built on:", x.Asset+", using "+x.Capability+".")
			row("What we found:", x.Findings)
			fmt.Fprintf(&b, `<p style="margin:0 0 6px;color:%s"><span style="color:%s;font-weight:bold">How it scored:</span> out of 5 — hard to copy %d &middot; people will pay %d &middot; easy to build %d &middot; reusable elsewhere %d &middot; built to last %d <span style="color:%s">(%d out of 25 overall)</span></p>`, slate, navy, x.Defensibility, x.Willingness, x.Buildability, x.Reuse, x.Durability, muted, x.Sum)
			fmt.Fprintf(&b, `<p style="margin:0 0 6px;color:%s"><span style="color:%s;font-weight:bold">Risk to you:</span> %d/5 %s</p>`, slate, navy, x.Risk, esc(riskNote(x.Risk)))
			row("A cheap first test:", x.CheapestTest)
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
	fmt.Fprintf(&b, `<div style="border-top:1px solid %s;margin-top:24px;padding-top:16px;color:%s;font-size:14px">%s</div>`, line, muted, esc(reportFooter()))
	b.WriteString(`</div></div></div>`)
	return b.String()
}

const reportRule = "------------------------------------------------------------"

// reportIntro is the plain-English opener so a recipient with a full inbox knows
// what this email is and what it contains. domain is the business descriptor.
func reportIntro(domain string) string {
	return "This report is from idea.uk. You asked us to find AI product ideas for " + domain + ". " +
		"We came up with a wide range of ideas, then checked each one against what already exists and " +
		"whether people would actually pay for it. Below are the ideas worth pursuing — with what we " +
		"found and a cheap way to test each one — followed by the ideas we looked at and set aside, and why."
}

func reportFooter() string {
	return "idea.uk finds and tests AI product ideas for your business, so you can spend your time " +
		"building the ones most likely to pay off. Questions about this report? Just reply to this email."
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

func render(domain, audience, wtp string, advancing, dropped, riskDropped []scored, note string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "IDEA REPORT — %s\n%s\n\n", domain, reportRule)
	fmt.Fprintf(&b, "%s\n\n", reportIntro(domain))
	fmt.Fprintf(&b, "WHO IT'S FOR\n%s\n\n", audience)
	fmt.Fprintf(&b, "WHY THEY'D PAY\n%s\n\n", wtp)
	if note != "" {
		fmt.Fprintf(&b, "%s\n\n", note)
	}
	if len(advancing) == 0 {
		b.WriteString("NO IDEA CLEARED THE BAR\n")
		b.WriteString("None of the ideas passed both of our main tests: being hard for someone else\n" +
			"to copy, and being something enough people would pay for (each needs at least 3\n" +
			"out of 5 on both). That is a real result, not a dead end — it usually means a\n" +
			"different audience, a different starting asset, or a different way to charge.\n\n")
	} else {
		b.WriteString("IDEAS WORTH PURSUING (best first)\n\n")
		for i, x := range advancing {
			fmt.Fprintf(&b, "%d) %s  [%s]\n", i+1, x.Title, flagLabel(x.Flag))
			if x.ShortLived {
				b.WriteString("   Heads-up: this one may not last — as the general AI models improve, they may do it too.\n")
			}
			if x.NeedsLiabilityWork {
				fmt.Fprintf(&b, "   Heads-up: needs legal and insurance groundwork before building (risk %d/5 — see below).\n", x.Risk)
			}
			fmt.Fprintf(&b, "   %s\n", x.BeatsFreeBecause)
			fmt.Fprintf(&b, "   What it's built on:  %s, using %s.\n", x.Asset, x.Capability)
			fmt.Fprintf(&b, "   What we found:       %s\n", x.Findings)
			fmt.Fprintf(&b, "   How it scored:       out of 5 — hard to copy %d, people will pay %d, easy to build %d, reusable elsewhere %d, built to last %d (%d out of 25 overall).\n",
				x.Defensibility, x.Willingness, x.Buildability, x.Reuse, x.Durability, x.Sum)
			fmt.Fprintf(&b, "   Risk to you:         %d/5 %s\n", x.Risk, riskNote(x.Risk))
			fmt.Fprintf(&b, "   A cheap first test:  %s\n\n", x.CheapestTest)
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
	fmt.Fprintf(&b, "%s\n%s\n", reportRule, reportFooter())
	return b.String()
}

// riskNote returns the short label shown next to the Risk score in the report.
func riskNote(r int) string {
	switch r {
	case 5:
		return "(pure analysis; customer decides)"
	case 4:
		return "(low — refunds make customers whole)"
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
