// FILE: platform/orchestration/actions/discovery_checks/check_tool_acceptance.go
//
// Discovery check: tool_acceptance — Tier 2 of the verification ladder
// (RUNBOOK_travelling_docs §0 Stage 5). A static, browserless verifier of a
// tool's ACCEPTANCE CRITERIA: it loads the current PLAN's ```criteria fence
// (doc_plans, same extraction as load_doc_context) and asserts the
// statically-visible subset against the DEPLOYED page HTML.
//
// THE ANCHOR RULE (settled 2026-07-09 on tool-xp-curve-designer): validate a
// selector's ANCHOR — its leftmost id/class/tag token — never the whole path.
//   #tableWrap tr  → anchor #tableWrap exists in the page  ⇒ PASS
//                    (rows are JS-built; Tier 4 asserts them for real)
//   #xpTableBody tr → anchor exists nowhere                ⇒ FAIL
//                    (an invented selector; the check is unsatisfiable)
// Static checks CONFIRM, never refute: a runtime-built path passes on its
// anchor; only an absent anchor fails. Check ids ending "-EDIT" are skipped
// (placeholder selectors, filled later by supersede).
//
// Evaluated statically:
//   selector_exists / selector_count — anchor rule on .selector
//   interaction                      — anchor rule on every steps[].selector
//                                      and expect.selector
//   asset_loads                      — .path referenced in the deployed HTML
//   page_status_ok                   — the fetch itself (HTTP 200)
//   + built-in shell checks: tool-doc header not leaked to the public page;
//     no "<no value>" template residue.
// Everything else (no_console_errors, no_horizontal_overflow, ...) is
// SKIPPED — that is Tier 4's job (headless runner).
//
// Outcomes:
//   no current PLAN with criteria → needs_criteria doc_note (30-day cooldown
//     per subject), NEVER a fake pass;
//   failures → one improve_tool work item (the failing criteria embedded as
//     acceptance_test) + an acceptance-fail doc_note;
//   passes → a finding only (Tier-2 pass must never be read as "the tool
//     works" — that claim belongs to Tier 4).
//
// Cooldown: shares tool_health's 7-day improve_tool-per-component cooldown so
// the two checks never double-team tool-improver.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/content"
)

func init() { Register(&ToolAcceptanceCheck{}) }

type ToolAcceptanceCheck struct{}

func (c *ToolAcceptanceCheck) Name() string { return "tool_acceptance" }

// fetchDeployedPage is swappable in tests.
var fetchDeployedPage = func(url string) (int, string, error) {
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2MB cap
	if err != nil {
		return resp.StatusCode, "", err
	}
	return resp.StatusCode, string(body), nil
}

func (c *ToolAcceptanceCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	var domain string
	if err := dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT domain FROM sites WHERE id = $1`, dctx.SiteID).Scan(&domain); err != nil {
		return nil, fmt.Errorf("tool_acceptance: site lookup failed: %w", err)
	}

	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT
			cc.id::text   AS component_id,
			cc.function,
			COALESCE(cc.display_name, cc.function) AS display_name,
			p.id::text    AS page_id,
			p.name        AS page_name,
			COALESCE(p.url, '') AS page_url,
			COALESCE(p.build_status, '') AS build_status
		FROM content_components cc
		JOIN page_components pc ON pc.component_id = cc.id
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND cc.component_level = 'tool'
		  AND cc.is_active = true
		  AND p.status = 'active'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("tool_acceptance: tool query failed: %w", err)
	}
	defer rows.Close()

	type toolRow struct {
		ComponentID, Function, DisplayName, PageID, PageName, PageURL, BuildStatus string
	}
	var tools []toolRow
	for rows.Next() {
		var t toolRow
		if err := rows.Scan(&t.ComponentID, &t.Function, &t.DisplayName,
			&t.PageID, &t.PageName, &t.PageURL, &t.BuildStatus); err != nil {
			dctx.Logger.Warn("tool_acceptance: scan error", zap.Error(err))
			continue
		}
		tools = append(tools, t)
	}
	if len(tools) == 0 {
		return result, nil
	}

	// Shared cooldown with tool_health: components with a recent improve_tool
	// item. Cancelled items don't count — a cancelled item means the finding
	// was resolved another way (e.g. PLAN-side, migrations 143/144), and the
	// tool should be re-checkable immediately (recorded follow-up 2026-07-10).
	recentItems := map[string]bool{}
	itemRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT spec->>'component_id'
		FROM site_work_items
		WHERE site_id = $1
		  AND item_type = 'improve_tool'
		  AND status <> 'cancelled'
		  AND created_at > NOW() - INTERVAL '7 days'
	`, dctx.SiteID)
	if err == nil {
		defer itemRows.Close()
		for itemRows.Next() {
			var compID sql.NullString
			if itemRows.Scan(&compID) == nil && compID.Valid {
				recentItems[compID.String] = true
			}
		}
	}

	pageCache := map[string]fetchedPage{} // url → page, one fetch per URL per run

	for _, tool := range tools {
		if recentItems[tool.ComponentID] {
			continue
		}

		criteriaJSON, hasPlan := loadCurrentCriteria(dctx, tool.Function)
		if criteriaJSON == "" {
			// No PLAN with criteria — a needs_criteria note, never a fake pass.
			c.noteNeedsCriteria(dctx, tool.Function, hasPlan)
			result.Findings = append(result.Findings, map[string]interface{}{
				"check": "tool_acceptance", "tool": tool.Function, "outcome": "needs_criteria",
			})
			continue
		}

		var crit criteriaDoc
		if err := json.Unmarshal([]byte(criteriaJSON), &crit); err != nil {
			// Unparseable fence is itself a contract failure worth surfacing.
			c.noteNeedsCriteria(dctx, tool.Function, hasPlan)
			result.Findings = append(result.Findings, map[string]interface{}{
				"check": "tool_acceptance", "tool": tool.Function,
				"outcome": "criteria_unparseable", "error": err.Error(),
			})
			continue
		}

		if tool.BuildStatus != "deployed" {
			// tool_health owns the not-deployed failure; nothing to assert against.
			result.Findings = append(result.Findings, map[string]interface{}{
				"check": "tool_acceptance", "tool": tool.Function, "outcome": "not_deployed_skip",
			})
			continue
		}

		url := "https://" + domain + tool.PageURL
		page, ok := pageCache[url]
		if !ok {
			status, html, ferr := fetchDeployedPage(url)
			page = fetchedPage{status: status, html: html, err: ferr}
			pageCache[url] = page
		}
		if page.err != nil {
			// Cannot confirm anything — and static checks never refute.
			dctx.Logger.Warn("tool_acceptance: fetch failed",
				zap.String("url", url), zap.Error(page.err))
			result.Findings = append(result.Findings, map[string]interface{}{
				"check": "tool_acceptance", "tool": tool.Function,
				"outcome": "fetch_failed", "url": url,
			})
			continue
		}

		ev := evaluateStaticCriteria(crit, page.status, page.html)

		result.Findings = append(result.Findings, map[string]interface{}{
			"check": "tool_acceptance", "tool": tool.Function,
			"passed": len(ev.passed), "failed": len(ev.failed), "skipped": len(ev.skipped),
		})

		if len(ev.failed) == 0 {
			continue // Tier-2 pass: finding only. "Works" is Tier 4's claim.
		}

		// Failure → improve_tool item carrying the failing criteria + a note.
		var failDescs []string
		for _, f := range ev.failed {
			failDescs = append(failDescs, f.id+": "+f.detail)
		}
		issue := strings.Join(failDescs, "; ")

		spec := map[string]interface{}{
			"component_id":    tool.ComponentID,
			"check":           "tool_acceptance",
			"page_id":         tool.PageID,
			"page_name":       tool.PageName,
			"issue":           issue,
			"failing_checks":  failedIDs(ev.failed),
			"acceptance_test": json.RawMessage(criteriaJSON),
		}
		specJSON, _ := json.Marshal(spec)

		var pageUUID *uuid.UUID
		if pu, perr := uuid.Parse(tool.PageID); perr == nil {
			pageUUID = &pu
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			PageID:       pageUUID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "improve_tool",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Acceptance (Tier 2) failed for %s: %s", tool.DisplayName, firstOf(failDescs)),
			SpecJSON:     string(specJSON),
			Priority:     60,
			HandlerAgent: "tool-improver",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("tool_acceptance:%s:%s", tool.Function, dctx.SiteID),
			BatchID:      dctx.BatchID,
		})

		c.noteAcceptanceFail(dctx, tool.Function, url, ev, issue)

		dctx.Logger.Info("tool_acceptance: static criteria failed",
			zap.String("tool", tool.Function),
			zap.Int("failed", len(ev.failed)),
			zap.String("issues", issue))
	}

	return result, nil
}

type fetchedPage struct {
	status int
	html   string
	err    error
}

// ── criteria document (schema v0, PLAN_tool_acceptance_runner) ─────────────

type criteriaDoc struct {
	Profiles []string        `json:"profiles"`
	Checks   []criteriaCheck `json:"checks"`
}

type criteriaCheck struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Selector string          `json:"selector"`
	Path     string          `json:"path"`
	Steps    []criteriaStep  `json:"steps"`
	Expect   criteriaExpect  `json:"expect"`
}

type criteriaStep struct {
	Action   string `json:"action"`
	Selector string `json:"selector"`
	Value    string `json:"value"`
}

type criteriaExpect struct {
	Selector string `json:"selector"`
}

// ── static evaluation ───────────────────────────────────────────────────────

type checkOutcome struct {
	id     string
	detail string
}

type evaluation struct {
	passed  []checkOutcome
	failed  []checkOutcome
	skipped []checkOutcome
}

func failedIDs(fs []checkOutcome) []string {
	ids := make([]string, 0, len(fs))
	for _, f := range fs {
		ids = append(ids, f.id)
	}
	return ids
}

func firstOf(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	return ss[0]
}

// evaluateStaticCriteria applies the anchor rule to the statically-visible
// subset of the criteria against the deployed page HTML.
func evaluateStaticCriteria(crit criteriaDoc, httpStatus int, html string) evaluation {
	var ev evaluation

	pass := func(id, detail string) { ev.passed = append(ev.passed, checkOutcome{id, detail}) }
	fail := func(id, detail string) { ev.failed = append(ev.failed, checkOutcome{id, detail}) }
	skip := func(id, detail string) { ev.skipped = append(ev.skipped, checkOutcome{id, detail}) }

	for _, ch := range crit.Checks {
		if strings.HasSuffix(ch.ID, "-EDIT") {
			skip(ch.ID, "placeholder selector (-EDIT) — awaiting real selectors")
			continue
		}
		switch ch.Type {
		case "selector_exists", "selector_count":
			anchor := selectorAnchor(ch.Selector)
			if anchor == "" {
				skip(ch.ID, "no parseable anchor in selector "+ch.Selector)
			} else if anchorPresent(html, anchor) {
				pass(ch.ID, "anchor "+anchor+" present")
			} else {
				fail(ch.ID, "anchor "+anchor+" absent from deployed page (selector "+ch.Selector+" is unsatisfiable)")
			}
		case "interaction":
			// Confirm the anchors of everything the interaction touches.
			bad := ""
			for _, st := range ch.Steps {
				if a := selectorAnchor(st.Selector); a != "" && !anchorPresent(html, a) {
					bad = a
					break
				}
			}
			if bad == "" && ch.Expect.Selector != "" {
				if a := selectorAnchor(ch.Expect.Selector); a != "" && !anchorPresent(html, a) {
					bad = a
				}
			}
			if bad != "" {
				fail(ch.ID, "interaction anchor "+bad+" absent from deployed page")
			} else {
				// Anchors confirmed; the behaviour itself is Tier 4's claim.
				pass(ch.ID, "interaction anchors present (behaviour asserted at Tier 4)")
			}
		case "asset_loads":
			if ch.Path == "" {
				skip(ch.ID, "asset_loads with empty path")
			} else if strings.Contains(html, ch.Path) {
				pass(ch.ID, "asset path referenced: "+ch.Path)
			} else {
				fail(ch.ID, "asset path "+ch.Path+" not referenced in deployed page (js-not-extracted?)")
			}
		case "page_status_ok":
			if httpStatus == http.StatusOK {
				pass(ch.ID, "HTTP 200")
			} else {
				fail(ch.ID, fmt.Sprintf("HTTP %d", httpStatus))
			}
		default:
			skip(ch.ID, ch.Type+" is not statically checkable (Tier 4)")
		}
	}

	// Built-in shell checks — always run, independent of the criteria.
	if strings.Contains(html, content.ToolDocOpen) {
		ev.failed = append(ev.failed, checkOutcome{"shell-doc-header",
			"tool-doc header leaked to the public page (strip runs at deploy assembly — 019)"})
	}
	if strings.Contains(html, "<no value>") {
		ev.failed = append(ev.failed, checkOutcome{"shell-template-residue",
			"template residue '<no value>' present in deployed page"})
	}

	return ev
}

// selectorAnchor returns the leftmost simple selector token: "#id", ".class"
// or a bare tag name. "#tableWrap tr" → "#tableWrap"; ".a > li" → ".a".
var anchorRe = regexp.MustCompile(`^\s*([#.]?[A-Za-z][A-Za-z0-9_-]*)`)

func selectorAnchor(selector string) string {
	m := anchorRe.FindStringSubmatch(selector)
	if m == nil {
		return ""
	}
	return m[1]
}

// anchorPresent implements the confirm-only presence test against raw HTML.
var classAttrRe = regexp.MustCompile(`class=["']([^"']*)["']`)

func anchorPresent(html, anchor string) bool {
	switch {
	case strings.HasPrefix(anchor, "#"):
		id := anchor[1:]
		return strings.Contains(html, `id="`+id+`"`) || strings.Contains(html, `id='`+id+`'`)
	case strings.HasPrefix(anchor, "."):
		// CSS class tokens are whitespace-delimited — "tool-page" does NOT
		// contain class "tool" (regexp \b would wrongly split on the hyphen).
		want := anchor[1:]
		for _, m := range classAttrRe.FindAllStringSubmatch(html, -1) {
			for _, token := range strings.Fields(m[1]) {
				if token == want {
					return true
				}
			}
		}
		return false
	default:
		return strings.Contains(strings.ToLower(html), "<"+strings.ToLower(anchor))
	}
}

// ── criteria + notes plumbing ───────────────────────────────────────────────

// loadCurrentCriteria returns the raw contents of the ```criteria fence in the
// tool's CURRENT PLAN ("" when no plan or no fence) and whether a plan exists.
func loadCurrentCriteria(dctx DiscoveryCheckContext, function string) (string, bool) {
	var body string
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT body FROM doc_plans
		WHERE subject_type = 'tool' AND subject_key = $1 AND is_current
	`, function).Scan(&body)
	if err != nil {
		return "", false
	}
	return extractCriteriaFence(body), true
}

// extractCriteriaFence mirrors load_doc_context's extraction: the raw contents
// of the first ```criteria fenced block, "" when absent.
func extractCriteriaFence(body string) string {
	const fence = "```criteria"
	start := strings.Index(body, fence)
	if start < 0 {
		return ""
	}
	rest := body[start+len(fence):]
	end := strings.Index(rest, "```")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(rest[:end])
}

// noteNeedsCriteria appends a needs_criteria doc_note (30-day per-subject
// cooldown) — a tool without criteria gets a note, never a fake pass.
func (c *ToolAcceptanceCheck) noteNeedsCriteria(dctx DiscoveryCheckContext, function string, hasPlan bool) {
	var recent bool
	_ = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT EXISTS (
			SELECT 1 FROM doc_notes
			WHERE subject_type='tool' AND subject_key=$1
			  AND categories ? 'needs_criteria'
			  AND created_at > NOW() - INTERVAL '30 days')
	`, function).Scan(&recent)
	if recent {
		return
	}
	body := fmt.Sprintf(`## Tier-2: no acceptance criteria — %s
Observed: tool_acceptance sweep found no current PLAN criteria fence (has_plan=%t).
Root cause: not-applicable
Fix: none applied — write a PLAN with the five standard checks plus interaction checks built from REAL selectors (never invented).
Verified: n/a
Categories: needs_criteria`, function, hasPlan)
	c.insertNote(dctx, function, body, `["needs_criteria"]`)
}

func (c *ToolAcceptanceCheck) noteAcceptanceFail(dctx DiscoveryCheckContext, function, url string, ev evaluation, issue string) {
	body := fmt.Sprintf(`## Tier-2 static acceptance failed — %s
Observed: %d of %d evaluable checks failed against %s: %s
Root cause: not diagnosed at this tier (static contract-presence only; anchors confirm, never refute)
Fix: improve_tool item created carrying the criteria as acceptance_test
Verified: n/a — behaviour is asserted at Tier 4
Categories: acceptance-fail`,
		function, len(ev.failed), len(ev.failed)+len(ev.passed), url, issue)
	c.insertNote(dctx, function, body, `["acceptance-fail"]`)
}

func (c *ToolAcceptanceCheck) insertNote(dctx DiscoveryCheckContext, function, body, categories string) {
	_, err := dctx.DB.ExecContext(dctx.Ctx, `
		INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories,
		                       source, source_agent, created_by)
		VALUES ('tool', $1, $2, $3, $4::jsonb, 'tool-acceptance-tier2', NULLIF($5,''), $5)
	`, function, dctx.SiteID, body, categories, dctx.AgentType)
	if err != nil {
		dctx.Logger.Warn("tool_acceptance: note insert failed",
			zap.String("tool", function), zap.Error(err))
	}
}
