// FILE: platform/orchestration/actions/refresh_product_specs_action.go
//
// RefreshProductSpecsAction re-verifies product specifications from each
// product's own source_url and refreshes the `products` row — the reusable
// capability that was missing when the first gripper rows were researched by
// hand (empty_sections_loop_integrity workstream, Session 6).
//
// Flow per product (rows with content_data.source_url, active, given category):
//   1. Firecrawl-scrape source_url -> markdown (FIRECRAWL_API_KEY, same infra
//      as vet_med_price_scrape_action.go).
//   2. Ground-extract specs with an LLM (Ollama; strict "only what is literally
//      stated, never infer" prompt) -> JSON of the known spec fields.
//   3. Merge NON-EMPTY extracted fields into products.specifications and stamp
//      content_data.verified_date. Empty/absent fields are left as they were —
//      a blocked or thin scrape never wipes a previously-good spec, and a
//      field the page doesn't state is never invented.
//
// DELIBERATELY NOT a discovery agent. It refreshes KNOWN source_urls; it does
// not web-search for products. Picking the right product page from search
// results is where wrong-product fabrication creeps in (the first manual
// Robotiq fetch returned the Hand-E, not the 2F-85). Adding a new manufacturer
// = a human inserts a products stub (name + source_url) and this action fills
// the specs. Discovery stays a human judgement; extraction+write is automated.
//
// Config:
//   - site_id      (required) — path or literal
//   - category     (optional, default "gripper")
//   - limit        (optional, default 20; hard cap 50)
//   - delay_ms     (optional, default 1500 — respectful pacing between scrapes)
//   - llm_model    (optional, default "mistral-small3.1")
//
// Registration: registry.go "refresh_product_specs".

package actions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// productSpecFields is the closed set of spec keys the gripper-spec-sheet
// template renders. The extraction prompt and the merge both use exactly this
// list — nothing else is written, so the LLM cannot smuggle new keys in.
var productSpecFields = []string{
	"manufacturer", "stroke", "gripping_force", "payload",
	"weight", "ip_rating", "interface", "voltage",
}

// specRegionMaxChars caps how much page text reaches the model, and it is a
// throughput budget rather than a taste call. Measured 2026-07-16 against this
// cluster's ollama-adapter (24B Mistral, Q4, CPU-only — no GPU on any node):
// prompt eval runs at ~3 tok/s and markdown spec tables tokenize at ~2.6
// chars/token, so every 1000 chars of page text costs ~2 minutes of inference.
// 1500 chars ≈ 1030 prompt tokens ≈ 5.5 min/product, which fits the 600s
// per-call timeout and keeps a 5-product run near ~30 min — inside the reaper's
// 90-min ceiling. It also stays under Ollama's default 2048 num_ctx, which
// would otherwise silently drop the START of the prompt (our instructions).
// The old value was 6000, which needed ~19 min of prompt eval against a 90s
// timeout: unreachable by a factor of 12. Raising this trades directly against
// wall-clock — re-measure before you touch it. Same budget as the sibling
// vet_med_price_scrape_action.go, which uses 1500 against this same model.
const specRegionMaxChars = 1500

// specSignalRe marks text that looks like gripper spec data — a field name or a
// number with an engineering unit. Used only to RANK regions of a page, so
// occasional false hits are harmless; missing the spec table is not.
var specSignalRe = regexp.MustCompile(
	`(?i)(technical\s+data|specifications?|technical\s+specs|datasheet|data\s+sheet|` +
		`stroke|gripping\s+force|grip\s+force|payload|workpiece\s+weight|weight|mass|` +
		`protection\s+class|ip\s?\d{2}|interface|voltage|manufacturer)` +
		`|\d+([.,]\d+)?\s*(mm|kg|n\b|v\s?dc|vdc|volt|bar|ms\b)`)

// selectSpecRegion returns the densest run of spec-looking lines that fits in
// limit chars, instead of the page's first limit chars.
//
// This is not an optimisation — at 1500 chars the choice of WHICH 1500 decides
// whether the run works at all. Manufacturer pages front-load nav, cookie
// banners and marketing; their spec table often sits thousands of chars down,
// so a head-of-page slice reliably hands the model everything except the data
// it was asked for, and the model then correctly returns {}.
//
// If nothing scores (no spec signals anywhere) this degrades to the head of the
// page — the old behaviour — and the caller's empty-object warning fires with
// the evidence, which is the honest outcome for a page that has no specs on it.
func selectSpecRegion(md string, limit int) string {
	if len(md) <= limit {
		return md
	}
	lines := strings.Split(md, "\n")
	scores := make([]int, len(lines))
	for i, l := range lines {
		scores[i] = len(specSignalRe.FindAllString(l, -1))
	}

	bestStart, bestScore := 0, -1
	for start := range lines {
		sum, chars := 0, 0
		for j := start; j < len(lines); j++ {
			c := len(lines[j]) + 1
			if chars+c > limit {
				break
			}
			chars += c
			sum += scores[j]
		}
		if sum > bestScore {
			bestScore, bestStart = sum, start
		}
	}

	var b strings.Builder
	for j := bestStart; j < len(lines); j++ {
		if b.Len()+len(lines[j])+1 > limit {
			break
		}
		b.WriteString(lines[j])
		b.WriteString("\n")
	}

	// A single line can be longer than the whole budget — Firecrawl emits pages
	// whose body is one unbroken line — and the loop above would then select
	// nothing and send the model an empty page. Fall back to a hard slice of the
	// best line rather than silently asking about no text at all.
	if strings.TrimSpace(b.String()) == "" {
		tail := strings.Join(lines[bestStart:], "\n")
		if len(tail) > limit {
			tail = tail[:limit]
		}
		return strings.TrimSpace(tail)
	}
	return strings.TrimSpace(b.String())
}

// specValueNormalizeRe collapses runs of whitespace so that values differing
// only in spacing compare equal.
var specValueNormalizeRe = regexp.MustCompile(`\s+`)

// normalizeSpecValue lowercases, unifies the dash characters spec sheets mix
// freely (en/em dash vs hyphen), and collapses whitespace.
func normalizeSpecValue(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.NewReplacer("–", "-", "—", "-", "−", "-").Replace(s)
	return specValueNormalizeRe.ReplaceAllString(s, " ")
}

// specValueIsRestatement reports whether `extracted` says the same thing as
// `existing` but with less information — i.e. existing already contains it and
// carries extra qualifying detail ("6 mm per jaw" vs "6 mm").
//
// Deliberately conservative: it only suppresses writes that strictly LOSE
// detail. Anything the existing value does not already contain — a changed
// number, a new unit, an added equivalent — is treated as a real update and
// written.
func specValueIsRestatement(existing, extracted string) bool {
	e, x := normalizeSpecValue(existing), normalizeSpecValue(extracted)
	if e == x || x == "" {
		return false // identical values aren't "restatements"; caller no-ops anyway
	}
	return len(e) > len(x) && strings.Contains(e, x)
}

var RefreshProductSpecsInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"category", "limit", "delay_ms", "llm_model"},
}

func init() {
	datahelpers.RegisterActionInputSpec("refresh_product_specs", RefreshProductSpecsInputSpec)
}

type productToRefresh struct {
	ID            string
	Name          string
	SourceURL     string
	Specification map[string]interface{}
}

func RefreshProductSpecsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "refresh_product_specs"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, RefreshProductSpecsInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}
	category := inputs.Get("category")
	if category == "" {
		category = "gripper"
	}
	limit := inputs.GetInt("limit", 20)
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	delayMs := inputs.GetInt("delay_ms", 1500)
	if delayMs <= 0 {
		delayMs = 1500
	}
	model := inputs.Get("llm_model")
	if model == "" {
		model = "mistral-small3.1"
	}

	firecrawlKey := os.Getenv("FIRECRAWL_API_KEY")
	if firecrawlKey == "" {
		return nil, fmt.Errorf("FIRECRAWL_API_KEY not set")
	}
	firecrawlURL := os.Getenv("FIRECRAWL_API_URL")
	if firecrawlURL == "" {
		firecrawlURL = "https://api.firecrawl.dev/v2"
	}

	products, err := loadProductsToRefresh(ctx, params.DB, siteID, category, limit)
	if err != nil {
		return nil, fmt.Errorf("load products failed: %w", err)
	}
	if len(products) == 0 {
		logger.Info("refresh_product_specs: no products with source_url to refresh",
			zap.String("category", category))
		return map[string]interface{}{"status": "complete", "refreshed": 0, "failed": 0, "products": 0}, nil
	}

	// 600s, not 90s. The extraction LLM is a 24B model on CPU-only nodes
	// (no nvidia.com/gpu anywhere in the cluster), measured at ~3 tok/s prompt
	// eval, and Mistral-Small's chat template alone costs ~360 tokens (~2 min)
	// before a single character of page text. 90s could never return: the
	// 2026-07-16 zero-refresh run was 5/5 "Client.Timeout exceeded while
	// awaiting headers", not a content problem. Matches the same-shape sibling
	// vet_med_price_scrape_action.go, which uses 600s against this same model.
	httpClient := &http.Client{Timeout: 600 * time.Second}
	refreshed, failed := 0, 0
	var details []map[string]interface{}

	for _, p := range products {
		if ctx.Err() != nil {
			logger.Info("refresh_product_specs: context cancelled, stopping")
			break
		}

		markdown, scrapeErr := firecrawlScrape(ctx, httpClient, firecrawlKey, firecrawlURL, p.SourceURL)
		if scrapeErr != nil {
			logger.Warn("refresh_product_specs: scrape failed — leaving existing specs intact",
				zap.String("product", p.Name), zap.String("url", p.SourceURL), zap.Error(scrapeErr))
			failed++
			details = append(details, map[string]interface{}{"product": p.Name, "status": "scrape_failed", "error": scrapeErr.Error()})
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
			continue
		}

		extracted := llmExtractProductSpecs(ctx, httpClient, model, p.Name, markdown, logger)
		// Keep only known, non-empty fields — the LLM cannot add keys, and an
		// absent/empty value never overwrites an existing good one.
		merged := map[string]interface{}{}
		for k, v := range p.Specification {
			merged[k] = v
		}
		updatedFields := 0
		for _, k := range productSpecFields {
			val, ok := extracted[k]
			if !ok {
				continue
			}
			s, ok := val.(string)
			if !ok || strings.TrimSpace(s) == "" {
				continue
			}
			s = strings.TrimSpace(s)

			cur, had := merged[k]
			curStr := ""
			if had {
				curStr = strings.TrimSpace(fmt.Sprintf("%v", cur))
			}

			// Never trade a richer value for a barer restatement of itself.
			// Spec tables split meaning across label and value ("Stroke per jaw
			// | 6 mm"), so the model correctly extracts "6 mm" where a human had
			// recorded "6 mm per jaw" — and for a parallel gripper that silently
			// halves the stated stroke. Same doctrine as the empty-field rule
			// above: a refresh may enrich or genuinely correct, never degrade.
			// A real change ("30 N" -> "45 N") is not a restatement and still
			// lands; an enrichment ("11 kg" -> "11 kg (24.3 lb)") still lands.
			if had && specValueIsRestatement(curStr, s) {
				continue
			}

			if !had || curStr != s {
				updatedFields++
			}
			merged[k] = s
		}

		if updatedFields == 0 {
			logger.Info("refresh_product_specs: no fields extracted — specs unchanged, verified_date not bumped",
				zap.String("product", p.Name))
			failed++
			details = append(details, map[string]interface{}{"product": p.Name, "status": "no_fields_extracted"})
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
			continue
		}

		if err := writeRefreshedSpecs(ctx, params.DB, p.ID, merged); err != nil {
			logger.Warn("refresh_product_specs: write failed",
				zap.String("product", p.Name), zap.Error(err))
			failed++
			details = append(details, map[string]interface{}{"product": p.Name, "status": "write_failed", "error": err.Error()})
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
			continue
		}

		refreshed++
		details = append(details, map[string]interface{}{"product": p.Name, "status": "refreshed", "fields_updated": updatedFields})
		logger.Info("refresh_product_specs: refreshed",
			zap.String("product", p.Name), zap.Int("fields_updated", updatedFields))
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	return map[string]interface{}{
		"status":    "complete",
		"products":  len(products),
		"refreshed": refreshed,
		"failed":    failed,
		"details":   details,
	}, nil
}

func loadProductsToRefresh(ctx context.Context, db *sql.DB, siteID uuid.UUID, category string, limit int) ([]productToRefresh, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, name,
		       COALESCE(content_data->>'source_url', ''),
		       COALESCE(specifications, '{}'::jsonb)::text
		FROM products
		WHERE site_id = $1
		  AND status = 'active'
		  AND category = $2
		  AND COALESCE(content_data->>'source_url', '') <> ''
		ORDER BY name
		LIMIT $3
	`, siteID, category, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []productToRefresh
	for rows.Next() {
		var p productToRefresh
		var specJSON string
		if err := rows.Scan(&p.ID, &p.Name, &p.SourceURL, &specJSON); err != nil {
			return nil, err
		}
		p.Specification = map[string]interface{}{}
		_ = json.Unmarshal([]byte(specJSON), &p.Specification)
		out = append(out, p)
	}
	return out, rows.Err()
}

// firecrawlScrape fetches a URL through Firecrawl and returns the page
// markdown. Mirrors the request shape used by vet_med_price_scrape_action.go.
func firecrawlScrape(ctx context.Context, client *http.Client, apiKey, apiURL, targetURL string) (string, error) {
	payload := map[string]interface{}{
		"url":     targetURL,
		"formats": []string{"markdown"},
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(apiURL, "/")+"/scrape", bytes.NewReader(payloadBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("firecrawl HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	// Firecrawl v2: {success, data: {markdown, ...}}
	var parsed struct {
		Success bool `json:"success"`
		Data    struct {
			Markdown string `json:"markdown"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("firecrawl decode: %w", err)
	}
	if strings.TrimSpace(parsed.Data.Markdown) == "" {
		return "", fmt.Errorf("firecrawl returned empty markdown")
	}
	return parsed.Data.Markdown, nil
}

// llmExtractProductSpecs asks a local LLM to pull ONLY literally-stated spec
// values from the page markdown. Grounded prompt, low temperature, strict JSON.
// Returns whatever known keys it found; callers keep only non-empty ones.
func llmExtractProductSpecs(ctx context.Context, client *http.Client, model, productName, markdown string, logger *zap.Logger) map[string]interface{} {
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
	}

	md := selectSpecRegion(markdown, specRegionMaxChars)

	prompt := fmt.Sprintf(`You are extracting robot-gripper specifications for the product "%s" from the page text below.

STRICT RULES:
- Extract ONLY values that are LITERALLY stated in the text for THIS product.
- Never infer, estimate, convert, or fill from prior knowledge. If a field is not explicitly present, OMIT it entirely.
- Copy units exactly as written (e.g. "20-235 N", "85 mm", "IP67").
- If the text is not a spec page for this product, return {}.

Return ONLY a JSON object using these keys (include a key only if the value is explicitly stated):
{"manufacturer": "...", "stroke": "...", "gripping_force": "...", "payload": "...", "weight": "...", "ip_rating": "...", "interface": "...", "voltage": "..."}

Page text:
%s`, productName, md)

	reqBody := map[string]interface{}{
		"model":    model,
		"messages": []map[string]interface{}{{"role": "user", "content": prompt}},
		"stream":   false,
		"format":   "json",
		// num_predict 200: the answer is 8 short fields (~150 tokens at most).
		// Output decode also runs at ~3 tok/s, so an over-generous cap is pure
		// wall-clock risk against the 600s timeout, not extra safety.
		"options": map[string]interface{}{"temperature": 0.0, "num_predict": 200},
	}
	reqBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(ollamaURL, "/")+"/api/chat", bytes.NewReader(reqBytes))
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("refresh_product_specs: LLM call failed", zap.Error(err))
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("refresh_product_specs: LLM response read failed",
			zap.String("product", productName), zap.Error(err))
		return nil
	}

	var chat struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(body, &chat); err != nil {
		logger.Warn("refresh_product_specs: LLM envelope unparseable",
			zap.String("product", productName),
			zap.String("body", truncate(string(body), 300)), zap.Error(err))
		return nil
	}

	// Diagnostics: a silent nil here is indistinguishable from "model said {}",
	// which is exactly the kind of blind spot this workstream exists to remove.
	// Log the model's own words whenever we can't turn them into fields.
	out := map[string]interface{}{}
	content := strings.TrimSpace(chat.Message.Content)
	if json.Unmarshal([]byte(content), &out) != nil {
		// Some models wrap JSON in prose/fences — salvage the first {...} block.
		if i, j := strings.Index(content, "{"), strings.LastIndex(content, "}"); i >= 0 && j > i {
			if err := json.Unmarshal([]byte(content[i:j+1]), &out); err != nil {
				logger.Warn("refresh_product_specs: LLM content not JSON even after salvage",
					zap.String("product", productName),
					zap.String("content", truncate(content, 300)), zap.Error(err))
				return nil
			}
		} else {
			logger.Warn("refresh_product_specs: LLM content has no JSON object",
				zap.String("product", productName),
				zap.String("content", truncate(content, 300)))
			return nil
		}
	}
	// Parsed, but the model may legitimately have returned {} ("not a spec page
	// for this product" per the prompt). Say so out loud with the evidence —
	// otherwise "no_fields_extracted" is an unexplainable dead end, which is
	// what the 2026-07-16 zero-refresh run looked like.
	if len(out) == 0 {
		// Say which text the model actually judged, not just how much of it.
		// "Empty object" has two very different causes — the page genuinely
		// lacks specs, or selectSpecRegion handed over the wrong slice — and
		// only the chosen text tells them apart.
		logger.Warn("refresh_product_specs: LLM returned an empty object — the selected page region did not state this product's specs",
			zap.String("product", productName),
			zap.Int("page_chars_scraped", len(markdown)),
			zap.Int("markdown_chars_sent", len(md)),
			zap.Int("spec_signals_in_region", len(specSignalRe.FindAllString(md, -1))),
			zap.String("region_head", truncate(md, 200)))
	}
	return out
}

// writeRefreshedSpecs replaces the row's specifications with the merged map and
// stamps content_data.verified_date to today (UTC). source_url is preserved.
func writeRefreshedSpecs(ctx context.Context, db *sql.DB, productID string, specs map[string]interface{}) error {
	specsJSON, err := json.Marshal(specs)
	if err != nil {
		return err
	}
	today := time.Now().UTC().Format("2006-01-02")
	_, err = db.ExecContext(ctx, `
		UPDATE products
		SET specifications = $2::jsonb,
		    content_data = COALESCE(content_data, '{}'::jsonb) || jsonb_build_object('verified_date', $3::text),
		    updated_at = NOW()
		WHERE id = $1::uuid
	`, productID, string(specsJSON), today)
	return err
}
