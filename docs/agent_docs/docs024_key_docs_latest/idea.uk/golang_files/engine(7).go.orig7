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
func renderHTML(domain, audience, wtp string, advancing, dropped, riskDropped []scored, note string) string {
	esc := html.EscapeString
	var b strings.Builder
	b.WriteString(`<div style="max-width:640px;margin:0 auto;padding:24px;background:#EFE7D6;color:#1A1816;font-family:-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;font-size:16px;line-height:1.6">`)
	b.WriteString(`<h1 style="font-family:Georgia,'Times New Roman',serif;font-size:26px;line-height:1.2;margin:0 0 4px">Idea report</h1>`)
	fmt.Fprintf(&b, `<p style="margin:0 0 20px;color:#837C72;font-size:15px">%s</p>`, esc(domain))
	sect := func(t string) {
		fmt.Fprintf(&b, `<h2 style="font-size:13px;letter-spacing:.08em;text-transform:uppercase;color:#A8391A;margin:24px 0 6px">%s</h2>`, esc(t))
	}
	sect("Who it's for")
	fmt.Fprintf(&b, `<p style="margin:0">%s</p>`, esc(audience))
	sect("Why they'd pay")
	fmt.Fprintf(&b, `<p style="margin:0">%s</p>`, esc(wtp))
	if note != "" {
		fmt.Fprintf(&b, `<p style="background:#E8DFCC;border-left:3px solid #A8391A;padding:12px 16px;margin:16px 0">%s</p>`, esc(note))
	}
	if len(advancing) == 0 {
		sect("No idea cleared the bar")
		b.WriteString(`<p style="margin:0">Nothing passed the test of being both hard to copy and something people will pay for (each idea needs at least 3 out of 5 on both). That is a real result, not a dead end — it usually points to a different audience, a different asset, or a different way to charge.</p>`)
	} else {
		sect("Advancing ideas (best first)")
		for i, x := range advancing {
			b.WriteString(`<div style="background:#fff;border:1px solid #D9CFB8;border-radius:6px;padding:16px 18px;margin:0 0 14px">`)
			fmt.Fprintf(&b, `<div style="font-family:Georgia,serif;font-size:18px;font-weight:bold;margin:0 0 8px">%d. %s <span style="font-size:12px;font-weight:normal;color:#A8391A;text-transform:uppercase;letter-spacing:.06em">[%s]</span></div>`, i+1, esc(x.Title), esc(x.Flag))
			if x.ShortLived {
				b.WriteString(`<div style="font-size:13px;color:#7D2A12;margin:0 0 6px">Short-lived — base-model progress may erode this.</div>`)
			}
			if x.NeedsLiabilityWork {
				fmt.Fprintf(&b, `<div style="font-size:13px;color:#7D2A12;margin:0 0 6px">Needs liability work before building (risk %d/5).</div>`, x.Risk)
			}
			row := func(label, val string) {
				fmt.Fprintf(&b, `<div style="margin:0 0 4px"><span style="color:#837C72">%s</span> %s</div>`, esc(label), esc(val))
			}
			row("Idea", x.BeatsFreeBecause)
			row("Built on", x.Asset+", using "+x.Capability+".")
			row("Checks out", x.Findings)
			fmt.Fprintf(&b, `<div style="margin:0 0 4px"><span style="color:#837C72">Scores</span> Defensibility %d/5, Willingness %d/5, Buildability %d/5, Reuse %d/5, Durability %d/5 (total %d/25).</div>`,
				x.Defensibility, x.Willingness, x.Buildability, x.Reuse, x.Durability, x.Sum)
			fmt.Fprintf(&b, `<div style="margin:0 0 4px"><span style="color:#837C72">Risk</span> %d/5 %s</div>`, x.Risk, esc(riskNote(x.Risk)))
			row("First test", x.CheapestTest)
			b.WriteString(`</div>`)
		}
	}
	if len(dropped) > 0 {
		sect("Didn't make the cut")
		b.WriteString(`<p style="margin:0 0 6px;color:#837C72;font-size:14px">Not hard enough to copy, or too little willingness to pay.</p><ul style="margin:0;padding-left:20px">`)
		for _, x := range dropped {
			fmt.Fprintf(&b, `<li style="margin:0 0 4px">%s <span style="color:#837C72">(Defensibility %d/5, Willingness %d/5)</span></li>`, esc(x.Title), x.Defensibility, x.Willingness)
		}
		b.WriteString(`</ul>`)
	}
	if len(riskDropped) > 0 {
		sect("Set aside on risk")
		b.WriteString(`<p style="margin:0 0 6px;color:#837C72;font-size:14px">Regulated-profession territory or similar. These may be real opportunities, but we can't build them safely without the right qualifications and cover, so they're here for awareness, not as advice.</p><ul style="margin:0;padding-left:20px">`)
		for _, x := range riskDropped {
			fmt.Fprintf(&b, `<li style="margin:0 0 4px">%s <span style="color:#837C72">(Risk 1/5; Defensibility %d/5, Willingness %d/5, Buildability %d/5)</span></li>`, esc(x.Title), x.Defensibility, x.Willingness, x.Buildability)
		}
		b.WriteString(`</ul>`)
	}
	b.WriteString(`<p style="margin:24px 0 0;padding-top:14px;border-top:1px solid #D9CFB8;color:#837C72;font-size:13px">idea.uk &middot; by leopardess.uk</p>`)
	b.WriteString(`</div>`)
	return b.String()
}

const reportRule = "------------------------------------------------------------"

func render(domain, audience, wtp string, advancing, dropped, riskDropped []scored, note string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "IDEA REPORT — %s\n%s\n\n", domain, reportRule)
	fmt.Fprintf(&b, "WHO IT'S FOR\n%s\n\n", audience)
	fmt.Fprintf(&b, "WHY THEY'D PAY\n%s\n\n", wtp)
	if note != "" {
		fmt.Fprintf(&b, "%s\n\n", note)
	}
	if len(advancing) == 0 {
		b.WriteString("NO IDEA CLEARED THE BAR\n")
		b.WriteString("Nothing passed the test of being both hard to copy and something people will\n" +
			"pay for (each idea needs at least 3 out of 5 on both). That is a real result,\n" +
			"not a dead end — it usually points to a different audience, a different asset,\n" +
			"or a different way to charge.\n\n")
	} else {
		b.WriteString("ADVANCING IDEAS (best first)\n\n")
		for i, x := range advancing {
			fmt.Fprintf(&b, "%d) %s  [%s]\n", i+1, x.Title, x.Flag)
			if x.ShortLived {
				b.WriteString("   Note: short-lived — base-model progress may erode this.\n")
			}
			if x.NeedsLiabilityWork {
				fmt.Fprintf(&b, "   Note: needs liability work before building (risk %d/5).\n", x.Risk)
			}
			fmt.Fprintf(&b, "   Idea:       %s\n", x.BeatsFreeBecause)
			fmt.Fprintf(&b, "   Built on:   %s, using %s.\n", x.Asset, x.Capability)
			fmt.Fprintf(&b, "   Checks out: %s\n", x.Findings)
			fmt.Fprintf(&b, "   Scores:     Defensibility %d/5, Willingness %d/5, Buildability %d/5, Reuse %d/5, Durability %d/5 (total %d/25).\n",
				x.Defensibility, x.Willingness, x.Buildability, x.Reuse, x.Durability, x.Sum)
			fmt.Fprintf(&b, "   Risk:       %d/5 %s\n", x.Risk, riskNote(x.Risk))
			fmt.Fprintf(&b, "   First test: %s\n\n", x.CheapestTest)
		}
	}
	if len(dropped) > 0 {
		fmt.Fprintf(&b, "%s\nDIDN'T MAKE THE CUT (not hard enough to copy, or too little willingness to pay)\n", reportRule)
		for _, x := range dropped {
			fmt.Fprintf(&b, "   - %s (Defensibility %d/5, Willingness %d/5)\n", x.Title, x.Defensibility, x.Willingness)
		}
		b.WriteString("\n")
	}
	if len(riskDropped) > 0 {
		fmt.Fprintf(&b, "%s\nSET ASIDE ON RISK (regulated-profession territory or similar)\n", reportRule)
		b.WriteString("These may be real opportunities, but we can't build them safely without the\n" +
			"right qualifications and cover, so they're here for awareness, not as advice.\n")
		for _, x := range riskDropped {
			fmt.Fprintf(&b, "   - %s (Risk 1/5; Defensibility %d/5, Willingness %d/5, Buildability %d/5)\n",
				x.Title, x.Defensibility, x.Willingness, x.Buildability)
		}
	}
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
		return "(moderate — cite sources; PII recommended)"
	case 2:
		return "(high — needs review, insurance, tight T&Cs before building)"
	case 1:
		return "(regulated territory — should not build without proper qualifications)"
	default:
		return ""
	}
}
