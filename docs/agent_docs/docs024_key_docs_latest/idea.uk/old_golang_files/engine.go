package old_golang_files

// engine.go — the ideation method (v2) as a staged, multi-model, web-verified
// pipeline. Go port of idea_method_runner.py. Calls Anthropic/OpenAI directly
// over net/http (no SDKs). Output is similar, not identical, to the by-hand runs;
// LLM output is non-deterministic.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Models. The cut runs on a DIFFERENT vendor if OPENAI_API_KEY is set
// (genuine cross-vendor critique), else a different Anthropic model.
// Strings verified against docs.claude.com (May 2026); override via env.
var (
	genModel      = env("GEN_MODEL", "claude-opus-4-7")
	critiqueModel = env("CRITIQUE_MODEL", "claude-sonnet-4-6")
	verifyModel   = env("VERIFY_MODEL", "claude-opus-4-7")
	scoreModel    = env("SCORE_MODEL", "claude-sonnet-4-6")
	openAIModel   = env("OPENAI_CRITIQUE_MODEL", "gpt-4o")
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

var httpClient = &http.Client{Timeout: 180 * time.Second}

// EngineFunc is the shape both front doors call. Swappable in tests.
type EngineFunc func(domain, audience, assets string) (string, error)

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
	ID            string `json:"id"`
	Title         string `json:"title"`
	Defensibility int    `json:"defensibility"`
	Willingness   int    `json:"willingness"`
	Buildability  int    `json:"buildability"`
	Reuse         int    `json:"reuse"`
	Durability    int    `json:"durability"`
	Sum           int    `json:"sum"`
	Advances      bool   `json:"advances"`
	ShortLived    bool   `json:"short_lived"`
	CheapestTest  string `json:"cheapest_test"`
	Flag          string `json:"flag"`
	candidate     `json:"-"`
}

// ── API calls ────────────────────────────────────────────────────────────────
func callClaude(model, system, user string, tools []map[string]any, maxTokens int) (string, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return "", errors.New("ANTHROPIC_API_KEY not set")
	}
	body := map[string]any{
		"model":      model,
		"max_tokens": maxTokens,
		"system":     system,
		"messages":   []map[string]any{{"role": "user", "content": user}},
	}
	if tools != nil {
		body["tools"] = tools
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
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, c := range out.Content { // concat text blocks; skip tool-use/result blocks
		if c.Type == "text" {
			sb.WriteString(c.Text)
		}
	}
	return sb.String(), nil
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
		return callOpenAI(systemBase, user)
	}
	return callClaude(critiqueModel, systemBase, user, nil, 4000)
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

// ── RunMethod: the pipeline ──────────────────────────────────────────────────
func RunMethod(domain, audience, assets string) (string, error) {
	// STEP 1 — audience framing + challenge
	s1, err := callClaude(genModel, systemBase,
		fill(audiencePrompt, "{domain}", domain, "{audience}", audience, "{assets}", assets), nil, 4096)
	if err != nil {
		return "", fmt.Errorf("audience step: %w", err)
	}
	var aud struct {
		CarriedAudience  string `json:"carried_audience"`
		WillingnessToPay string `json:"willingness_to_pay"`
	}
	if err := parseJSON(s1, &aud); err != nil {
		return "", fmt.Errorf("audience parse: %w", err)
	}

	// STEP 2 — generate (multi-lens)
	s2, err := callClaude(genModel, systemBase,
		fill(generatePrompt, "{domain}", domain, "{audience}", aud.CarriedAudience,
			"{wtp}", aud.WillingnessToPay, "{assets}", assets, "{capabilities}", capabilityMenu),
		nil, 6000)
	if err != nil {
		return "", fmt.Errorf("generate step: %w", err)
	}
	var gen struct {
		Candidates []candidate `json:"candidates"`
	}
	if err := parseJSON(s2, &gen); err != nil {
		return "", fmt.Errorf("generate parse: %w", err)
	}
	for i := range gen.Candidates {
		gen.Candidates[i].ID = fmt.Sprintf("c%d", i+1) // stable ids, threaded by id not title
	}

	// STEP 3 — cut (cross-vendor if configured)
	candJSON, _ := json.MarshalIndent(gen.Candidates, "", "  ")
	s3, err := critique(fill(cutPrompt, "{candidates_json}", string(candJSON)))
	if err != nil {
		return "", fmt.Errorf("cut step: %w", err)
	}
	var cut struct {
		Results []struct {
			ID      string `json:"id"`
			Verdict string `json:"verdict"`
		} `json:"results"`
	}
	if err := parseJSON(s3, &cut); err != nil {
		return "", fmt.Errorf("cut parse: %w", err)
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
		return render(domain, aud.CarriedAudience, aud.WillingnessToPay, nil, nil,
			"No candidate survived the cut."), nil
	}

	// STEP 4 — verify (web search)
	survJSON, _ := json.MarshalIndent(survivors, "", "  ")
	s4, err := callClaude(verifyModel, systemBase,
		fill(verifyPrompt, "{domain}", domain, "{survivors_json}", string(survJSON)),
		[]map[string]any{{"type": "web_search_20250305", "name": "web_search", "max_uses": 6}}, 6000)
	if err != nil {
		return "", fmt.Errorf("verify step: %w", err)
	}
	var ver struct {
		Results []struct {
			ID           string `json:"id"`
			Findings     string `json:"findings"`
			PremiseHolds bool   `json:"premise_holds"`
		} `json:"results"`
	}
	if err := parseJSON(s4, &ver); err != nil {
		return "", fmt.Errorf("verify parse: %w", err)
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
		return render(domain, aud.CarriedAudience, aud.WillingnessToPay, nil, nil,
			"No premise survived verification."), nil
	}

	// STEP 5 — score
	verJSON, _ := json.MarshalIndent(verified, "", "  ")
	s5, err := callClaude(scoreModel, systemBase,
		fill(scorePrompt, "{verified_json}", string(verJSON)), nil, 4000)
	if err != nil {
		return "", fmt.Errorf("score step: %w", err)
	}
	var sc struct {
		Scored []scored `json:"scored"`
	}
	if err := parseJSON(s5, &sc); err != nil {
		return "", fmt.Errorf("score parse: %w", err)
	}
	byID := map[string]candidate{}
	for _, c := range verified {
		byID[c.ID] = c
	}
	var advancing, dropped []scored
	for _, s := range sc.Scored {
		base, ok := byID[s.ID] // merge by id; skip unmatched (no blank rows)
		if !ok {
			continue
		}
		s.candidate = base
		if s.Advances {
			advancing = append(advancing, s)
		} else {
			dropped = append(dropped, s)
		}
	}
	// rank advancing by sum, descending
	for i := 0; i < len(advancing); i++ {
		for j := i + 1; j < len(advancing); j++ {
			if advancing[j].Sum > advancing[i].Sum {
				advancing[i], advancing[j] = advancing[j], advancing[i]
			}
		}
	}
	return render(domain, aud.CarriedAudience, aud.WillingnessToPay, advancing, dropped, ""), nil
}

func render(domain, audience, wtp string, advancing, dropped []scored, note string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Idea report — %s\n\n", domain)
	fmt.Fprintf(&b, "**Audience:** %s  \n**Willingness to pay:** %s\n\n", audience, wtp)
	if note != "" {
		fmt.Fprintf(&b, "> %s\n\n", note)
	}
	if len(advancing) == 0 {
		b.WriteString("**No candidate advanced the gate** (Defensibility ≥3 AND Willingness ≥3). " +
			"That is a real outcome — consider a different audience, a new asset, or a different monetisation.\n")
	} else {
		b.WriteString("## Advancing candidates (ranked)\n\n")
		for i, x := range advancing {
			sl := ""
			if x.ShortLived {
				sl = " — *short-lived (low durability)*"
			}
			fmt.Fprintf(&b, "### %d. %s  [%s]%s\n", i+1, x.Title, x.Flag, sl)
			fmt.Fprintf(&b, "- **Idea:** use *%s* on *%s* to %s\n", x.Capability, x.Asset, x.BeatsFreeBecause)
			fmt.Fprintf(&b, "- **Verification:** %s\n", x.Findings)
			fmt.Fprintf(&b, "- **Scores:** Defensibility %d/5 · Willingness %d/5 · Buildability %d/5 · Reuse %d/5 · Durability %d/5 (sum %d)\n",
				x.Defensibility, x.Willingness, x.Buildability, x.Reuse, x.Durability, x.Sum)
			fmt.Fprintf(&b, "- **Cheapest test:** %s\n\n", x.CheapestTest)
		}
	}
	if len(dropped) > 0 {
		b.WriteString("## Did not advance\n\n")
		for _, x := range dropped {
			fmt.Fprintf(&b, "- **%s** — Def %d / Will %d (failed the gate)\n", x.Title, x.Defensibility, x.Willingness)
		}
	}
	return b.String()
}
