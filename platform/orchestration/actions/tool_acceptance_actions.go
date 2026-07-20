// FILE: platform/orchestration/actions/tool_acceptance_actions.go
//
// The two actions behind tool-acceptance-agent — the orchestrator that makes
// Tier 4 self-driving (RUNBOOK_travelling_docs §0 Stage 6; flow pinned in
// PLAN_tool_acceptance_runner):
//
//   request_browser_run — sends a run_checks request to the browser-runner
//     adapter (system.adapter.browser-runner.requests) and returns
//     AwaitResponse=true, mirroring request_repo_analysis / webscrape: the
//     engine registers an awaited request and resumes when the adapter
//     replies; the reply lands under this step's output_field. The action
//     resolves the tool's deployed URL from pages itself (complexity in Go,
//     workflow stays flat) and NO-OP SKIPS — without awaiting — when the
//     criteria are empty: an undocumented tool is Tier-2's needs_criteria
//     concern, never a fake browser pass.
//
//   judge_acceptance_results — turns the adapter's reply into the loop's
//     artifacts: all pass → one acceptance-run doc_note; any fail → an
//     acceptance-fail doc_note + ONE improve_tool work item carrying the
//     criteria as acceptance_test and the failing check ids (the fixer loads
//     PLAN+NOTES first, per Task 4). Reads the reply through a fallback chain
//     of paths (003 action-level defense) because awaited-response shapes
//     vary across the codebase (.response.data vs flattened).

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

const browserRunnerTopic = "system.adapter.browser-runner.requests"

var RequestBrowserRunInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"function_field", "criteria_field", "site_id_field", "domain_field",
		"url_field", "profiles",
	},
	Defaults: map[string]interface{}{
		"function_field": "input_data.spec.function",
		"criteria_field": "doc_context.criteria_json",
		"site_id_field":  "site_record.site_id",
		"domain_field":   "site_record.domain",
	},
}

var JudgeAcceptanceResultsInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"results_field", "function_field", "criteria_field", "site_id_field",
	},
	Defaults: map[string]interface{}{
		"results_field":  "browser_run",
		"function_field": "input_data.spec.function",
		"criteria_field": "doc_context.criteria_json",
		"site_id_field":  "site_record.site_id",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("request_browser_run", RequestBrowserRunInputSpec)
	datahelpers.RegisterActionInputSpec("judge_acceptance_results", JudgeAcceptanceResultsInputSpec)
}

// resolveWithFallbacks tries each dot-path in order, returning the first
// non-empty string (003 action-level defense).
func resolveWithFallbacks(collected map[string]interface{}, paths ...string) string {
	for _, p := range paths {
		if v := datahelpers.ExtractNestedFieldString(collected, p); v != "" {
			return v
		}
	}
	return ""
}

// ── request_browser_run ─────────────────────────────────────────────────────

func RequestBrowserRunAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "request_browser_run"))
	config := params.StepConfig.Config

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	function := resolveWithFallbacks(params.CollectedData,
		datahelpers.GetStringField(config, "function_field", "input_data.spec.function"),
		"input_data.function")
	if function == "" {
		return nil, fmt.Errorf("request_browser_run: no function (input_data.spec.function / input_data.function)")
	}

	criteria := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "criteria_field", "doc_context.criteria_json"))
	if strings.TrimSpace(criteria) == "" {
		// Tier-2's sweeps own the needs_criteria note; here we simply decline
		// to fake a run. No await — judge sees skipped=true and passes through.
		logger.Info("request_browser_run: no criteria in the current PLAN — skipping (no fake pass)",
			zap.String("function", function))
		return map[string]interface{}{
			"skipped": true,
			"reason":  "needs_criteria",
		}, nil
	}

	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "site_id_field", "site_record.site_id"))
	domain := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "domain_field", "site_record.domain"))

	// Resolve the deployed URL: explicit config path first, else the pages
	// table (generator tools: page name == function).
	pageURL := ""
	if uf := datahelpers.GetStringField(config, "url_field", ""); uf != "" {
		pageURL = datahelpers.ExtractNestedFieldString(params.CollectedData, uf)
	}
	if pageURL == "" && params.DB != nil && siteID != "" {
		err := params.DB.QueryRowContext(ctx, `
			SELECT COALESCE(url, '') FROM pages
			WHERE site_id = $1::uuid AND name = $2 AND status = 'active'
		`, siteID, function).Scan(&pageURL)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("request_browser_run: page lookup failed: %w", err)
		}
	}
	if pageURL == "" {
		return nil, fmt.Errorf("request_browser_run: no deployed page URL for function %q on site %s", function, siteID)
	}
	fullURL := pageURL
	if !strings.HasPrefix(fullURL, "http") {
		fullURL = "https://" + domain + "/" + strings.TrimPrefix(pageURL, "/")
	}

	profiles := []string{"desktop"}
	if raw, ok := config["profiles"].([]interface{}); ok && len(raw) > 0 {
		profiles = profiles[:0]
		for _, p := range raw {
			if s, ok := p.(string); ok {
				profiles = append(profiles, s)
			}
		}
	}

	newRequestID := uuid.NewString()

	myResponsesTopic := params.ExecutionContext.ResponsesTopic
	if myResponsesTopic == "" {
		myResponsesTopic = os.Getenv("RESPONSES_TOPIC")
	}
	if myResponsesTopic == "" {
		myResponsesTopic = fmt.Sprintf("system.agent.%s.responses", params.ExecutionContext.Sender.AgentType)
	}
	clientID := params.ExecutionContext.ClientID
	if clientID == "" {
		clientID = "default"
	}

	// Envelope exactly as the adapter parses it (035 §1.2): body.action +
	// body.data + reply_to_topic in body AND headers (belt-and-braces).
	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":          params.ExecutionContext.CorrelationID,
			"orchestration_id":        params.ExecutionContext.OrchestrationID,
			"orchestration_name":      params.ExecutionContext.OrchestrationName,
			"parent_orchestration_id": params.ExecutionContext.ParentOrchestrationID,
			"client_id":               clientID,
			"step_name":               params.ExecutionContext.StepName,
			"step_id":                 params.ExecutionContext.StepID,
			"request_id":              newRequestID,
			"message_type":            "request",
			"action":                  "run_checks",
			"sender_agent_type":       params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":         params.ExecutionContext.OrchestrationID,
			"sender_pod_name":         params.ExecutionContext.Sender.PodName,
			"responses_topic":         myResponsesTopic,
			"parent_responses_topic":  myResponsesTopic,
			"reply_to_topic":          myResponsesTopic,
			"timestamp":               time.Now().UTC().Format(time.RFC3339),
		},
		"body": map[string]interface{}{
			"action":         "run_checks",
			"reply_to_topic": myResponsesTopic,
			"data": map[string]interface{}{
				"run_id":        params.ExecutionContext.OrchestrationID,
				"urls":          []string{fullURL},
				"profiles":      profiles,
				"criteria_json": criteria,
				"function":      function,
				"site_id":       siteID,
			},
		},
	}

	rawHeaders := adapterRequest["headers"].(map[string]interface{})
	headers := make(map[string]string, len(rawHeaders))
	for k, v := range rawHeaders {
		if s, ok := v.(string); ok {
			headers[k] = s
		} else {
			headers[k] = fmt.Sprintf("%v", v)
		}
	}

	messageBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return nil, fmt.Errorf("request_browser_run: marshal request: %w", err)
	}

	logger.Info("request_browser_run: sending to browser-runner adapter",
		zap.String("topic", browserRunnerTopic),
		zap.String("request_id", newRequestID),
		zap.String("url", fullURL),
		zap.String("function", function),
		zap.Strings("profiles", profiles))

	if err := params.Producer.ProduceWithValidation(ctx, browserRunnerTopic, headers,
		[]byte(params.ExecutionContext.CorrelationID), messageBytes); err != nil {
		return nil, fmt.Errorf("request_browser_run: send to browser-runner adapter: %w", err)
	}

	return &RequestRepoAnalysisResult{ // same await-signal shape as the analyser/webscrape requests
		Success:       true,
		RequestID:     newRequestID,
		TopicSentTo:   browserRunnerTopic,
		AwaitResponse: true,
		Metadata: map[string]interface{}{
			"function":        function,
			"url":             fullURL,
			"profiles":        profiles,
			"responses_topic": myResponsesTopic,
		},
	}, nil
}

// ── judge_acceptance_results ────────────────────────────────────────────────

// acceptanceVerdict is what extractRunResults distils from the adapter reply.
//
// A check runs once PER PROFILE, so a bare check id is NOT a unique result:
// mobile-fit passes on mobile and is (correctly) skipped on desktop. Passed /
// Failed / SkipList therefore hold "id@profile" LABELS — a note that said
// "skipped: mobile-fit" while mobile-fit passed on mobile read as "mobile was
// never checked", the opposite of the truth. FailedIDs keeps the bare, deduped
// ids for the improve_tool spec, whose contract is criteria ids (as Tier 2's).
type acceptanceVerdict struct {
	Skipped   bool
	Results   []map[string]interface{}
	Profiles  []string
	Passed    []string
	Failed    []string // tool-scoped failures (incl. unattributed) — the tool's problem
	FailedIDs []string
	Details   []string // "id@profile: detail" for each tool-scoped failure
	SkipList  []string
	Chrome    []chromeFailure // failures the adapter attributed to SITE CHROME
	Shots     []screenshotRef // P3 evidence: one full-page screenshot per failing (url, profile)
	// Drill-down attribution for a tool-scoped overflow: the element that
	// actually forces the width, and why (bugs_open/010). First non-empty seen —
	// a run overflows once. Empty when the widest offender is itself the cause.
	ForcedBy     string
	ForcedReason string
}

// screenshotRef is the P3 evidence the adapter attached to a failing run.
// URI is the durable s3:// pointer — the ONLY form that may enter a doc_note
// (notes are loaded into LLM prompt contexts; presigned URLs are hundreds of
// chars of expiring signature). ViewURL is a presigned GET for the work item's
// spec, where a human triaging the ticket can click it (7-day expiry).
type screenshotRef struct {
	Profile string
	URL     string
	URI     string
	ViewURL string
	Failing []string // id@profile instances the screenshot evidences
}

// chromeFailure is a document-level failure the adapter proved lies OUTSIDE the
// tool's container (e.g. an overflowing site footer). It is real and must be
// fixed — but by the template fixer, not by editing a tool that cannot reach it.
type chromeFailure struct {
	CheckID   string
	Profile   string
	Component string // e.g. "site-footer" — what a fixer would edit
	Culprit   string // human: "div.footer-legal (506px)"
	Selector  string // machine: "div.footer-legal" — the fixer patches THIS
	Slot      string // header | footer | head — which site_components row owns it
	Detail    string
	URL       string
	// The deepest descendant that forces the width, and why (bugs_open/010).
	ForcedBy     string
	ForcedReason string
}

// checkLabel names one result instance: "curve-switch@mobile" (bare id when the
// adapter reports no profile, e.g. a pre-P1 runner).
func checkLabel(id, profile string) string {
	if profile == "" {
		return id
	}
	return id + "@" + profile
}

// appendUnique keeps first-seen order (ids repeat once per profile).
func appendUnique(ss []string, s string) []string {
	for _, existing := range ss {
		if existing == s {
			return ss
		}
	}
	return append(ss, s)
}

// extractRunResults reads the awaited adapter reply through a fallback chain
// of paths — response shapes vary across the codebase (.response.data.results
// vs flattened) — and recomputes the verdict from the results themselves.
func extractRunResults(collected map[string]interface{}, field string) acceptanceVerdict {
	var v acceptanceVerdict

	if skipped, ok := datahelpers.ExtractNestedField(collected, field+".skipped").(bool); ok && skipped {
		v.Skipped = true
		return v
	}

	var raw interface{}
	for _, p := range []string{
		field + ".response.data.results",
		field + ".response.results",
		field + ".data.results",
		field + ".results",
	} {
		if raw = datahelpers.ExtractNestedField(collected, p); raw != nil {
			break
		}
	}
	items, _ := raw.([]interface{})
	for _, it := range items {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		v.Results = append(v.Results, m)
		id, _ := m["check_id"].(string)
		profile, _ := m["profile"].(string)
		pass, _ := m["pass"].(bool)
		detail, _ := m["detail"].(string)
		label := checkLabel(id, profile)
		if profile != "" {
			v.Profiles = appendUnique(v.Profiles, profile)
		}
		if pass {
			v.Passed = append(v.Passed, label)
			continue
		}
		// A failure the adapter attributed to site chrome is NOT the tool's bug:
		// route it, never blame the tool for a footer it cannot reach. Anything
		// else (tool-scoped or unattributed) stays the tool's problem.
		forcedBy, _ := m["forced_by"].(string)
		forcedReason, _ := m["forced_reason"].(string)
		if scope, _ := m["scope"].(string); scope == "chrome" {
			component, _ := m["component"].(string)
			culprit, _ := m["culprit"].(string)
			selector, _ := m["culprit_selector"].(string)
			slot, _ := m["slot"].(string)
			url, _ := m["url"].(string)
			v.Chrome = append(v.Chrome, chromeFailure{
				CheckID: id, Profile: profile, Component: component,
				Culprit: culprit, Selector: selector, Slot: slot,
				Detail: detail, URL: url,
				ForcedBy: forcedBy, ForcedReason: forcedReason,
			})
			continue
		}
		v.Failed = append(v.Failed, label)
		v.FailedIDs = appendUnique(v.FailedIDs, id)
		v.Details = append(v.Details, label+": "+detail)
		// Keep the first drill-down attribution seen for the tool's own overflow
		// (a run overflows once) — it becomes a structured hint on the fix ticket.
		if v.ForcedBy == "" && forcedBy != "" {
			v.ForcedBy = forcedBy
		}
		if v.ForcedReason == "" && forcedReason != "" {
			v.ForcedReason = forcedReason
		}
	}

	var rawSkip interface{}
	for _, p := range []string{
		field + ".response.data.skipped",
		field + ".response.skipped",
		field + ".data.skipped",
		field + ".skipped",
	} {
		if rawSkip = datahelpers.ExtractNestedField(collected, p); rawSkip != nil {
			break
		}
	}
	if skipItems, ok := rawSkip.([]interface{}); ok {
		for _, it := range skipItems {
			if m, ok := it.(map[string]interface{}); ok {
				if id, _ := m["check_id"].(string); id != "" {
					profile, _ := m["profile"].(string)
					v.SkipList = append(v.SkipList, checkLabel(id, profile))
				}
			}
		}
	}

	var rawShots interface{}
	for _, p := range []string{
		field + ".response.data.screenshots",
		field + ".response.screenshots",
		field + ".data.screenshots",
		field + ".screenshots",
	} {
		if rawShots = datahelpers.ExtractNestedField(collected, p); rawShots != nil {
			break
		}
	}
	if shotItems, ok := rawShots.([]interface{}); ok {
		for _, it := range shotItems {
			m, ok := it.(map[string]interface{})
			if !ok {
				continue
			}
			ref := screenshotRef{}
			ref.Profile, _ = m["profile"].(string)
			ref.URL, _ = m["url"].(string)
			ref.URI, _ = m["uri"].(string)
			ref.ViewURL, _ = m["view_url"].(string)
			if fc, ok := m["failing_checks"].([]interface{}); ok {
				for _, f := range fc {
					if s, ok := f.(string); ok {
						ref.Failing = append(ref.Failing, s)
					}
				}
			}
			if ref.URI != "" {
				v.Shots = append(v.Shots, ref)
			}
		}
	}
	return v
}

// evidenceLine renders the P3 screenshots for a note body — durable URIs only,
// never the presigned ViewURL (that lives in the item spec). Empty when the
// run produced no evidence (clean pass, or screenshots not configured).
func evidenceLine(shots []screenshotRef) string {
	if len(shots) == 0 {
		return ""
	}
	parts := make([]string, 0, len(shots))
	for _, s := range shots {
		p := s.URI
		if s.Profile != "" {
			p += " (" + s.Profile + ")"
		}
		parts = append(parts, p)
	}
	return "\nEvidence: full-page screenshot(s) at failure: " + strings.Join(parts, "; ")
}

// shotsForSpec renders screenshot refs for a work item's spec (uri + view_url).
// profile filters to one profile's evidence; "" keeps everything.
func shotsForSpec(shots []screenshotRef, profile string) []map[string]interface{} {
	var out []map[string]interface{}
	for _, s := range shots {
		if profile != "" && s.Profile != "" && s.Profile != profile {
			continue
		}
		out = append(out, map[string]interface{}{
			"profile": s.Profile, "uri": s.URI, "view_url": s.ViewURL,
		})
	}
	return out
}

func JudgeAcceptanceResultsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "judge_acceptance_results"))
	config := params.StepConfig.Config

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	function := resolveWithFallbacks(params.CollectedData,
		datahelpers.GetStringField(config, "function_field", "input_data.spec.function"),
		"input_data.function")
	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "site_id_field", "site_record.site_id"))
	resultsField := datahelpers.GetStringField(config, "results_field", "browser_run")

	v := extractRunResults(params.CollectedData, resultsField)

	if v.Skipped {
		// No criteria — nothing was run; Tier 2 owns the needs_criteria note.
		return map[string]interface{}{
			"all_passed": false, "skipped": true, "reason": "needs_criteria",
		}, nil
	}
	if len(v.Results) == 0 {
		return nil, fmt.Errorf("judge_acceptance_results: no results at %q (or its response fallbacks)", resultsField)
	}

	sourceAgent := ""
	if params.ExecutionContext != nil {
		sourceAgent = params.Headers["agent_type"]
	}

	// Site-chrome failures are routed FIRST and independently of the tool's own
	// verdict: they are real, user-visible defects that no tool edit can fix.
	chromeRouted := routeChromeFailures(ctx, params, logger, v.Chrome, v.Shots, function, siteID, sourceAgent)

	allPassed := len(v.Failed) == 0
	if allPassed {
		// The TOOL passed. The page may still carry a site-chrome defect — say so
		// plainly rather than let "PASSED" imply the page is clean.
		rootCause := "not-applicable"
		fix := "none required"
		if len(v.Chrome) > 0 {
			rootCause = "site chrome, not this tool: " + chromeSummary(v.Chrome)
			fix = fmt.Sprintf("%d responsive_fix item(s) raised for the site template (handler component-template-fixer); the tool itself needs no change", chromeRouted)
		}
		// Evidence exists on a pass ONLY when the page failed for site-chrome
		// reasons — the screenshot shows the chrome defect, not the tool.
		body := fmt.Sprintf(`## Tier-4 acceptance PASSED — %s
Observed: all %d of the tool's own checks passed in headless Chromium%s (%d skipped: %s).%s
Root cause: %s
Fix: %s
Verified: browser-runner-adapter run; checks (id@profile): %s
Categories: acceptance-run`,
			function, len(v.Passed), profilesPhrase(v.Profiles),
			len(v.SkipList), strings.Join(orNone(v.SkipList), ", "),
			evidenceLine(v.Shots),
			rootCause, fix,
			strings.Join(v.Passed, ", "))
		if _, err := insertDocNote(ctx, params.DB, "tool", function, siteID, body,
			`["acceptance-run"]`, "tool-acceptance", sourceAgent, "", "tool-acceptance-agent"); err != nil {
			logger.Warn("judge: acceptance-run note insert failed", zap.Error(err))
		}
		logger.Info("judge: acceptance PASSED",
			zap.String("function", function), zap.Int("passed", len(v.Passed)),
			zap.Int("site_chrome_items", chromeRouted))
		return map[string]interface{}{
			"all_passed": true, "passed": len(v.Passed), "failed": 0, "skipped_checks": v.SkipList,
			"site_chrome_failures": len(v.Chrome), "site_chrome_items_created": chromeRouted,
		}, nil
	}

	// Failures: one acceptance-fail note + ONE improve_tool item carrying the
	// criteria as acceptance_test (findings pattern). The fixer's per-item
	// max_fix_attempts convention never engages ACROSS cycles — each verdict
	// raises a fresh item once the last went terminal — so the convergence
	// guard below is what bounds the loop (bugs_open/010 candidate b).
	issue := strings.Join(v.Details, "; ")
	chromeLine := ""
	if len(v.Chrome) > 0 {
		chromeLine = fmt.Sprintf("\nSite chrome (NOT this tool, routed separately as %d responsive_fix item(s)): %s",
			chromeRouted, chromeSummary(v.Chrome))
	}
	// Convergence guard (bugs_open/010 b): how many cycles have already failed at
	// THESE criteria? Counted before the note is written so the note records what
	// actually happened. Fail-open — a counting error must not cost a fix cycle.
	maxCycles := datahelpers.GetIntField(config, "max_fix_cycles", 2)
	attempts := 0
	if params.DB != nil && siteID != "" {
		n, err := convergenceAttempts(ctx, params.DB, function, siteID, v.FailedIDs)
		if err != nil {
			logger.Warn("judge: convergence attempt count failed — raising improve_tool as usual",
				zap.Error(err))
		} else {
			attempts = n
		}
	}
	stuck := attempts >= maxCycles

	criteria := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "criteria_field", "doc_context.criteria_json"))

	itemCreated := false
	escalated := false
	if params.DB != nil && siteID != "" {
		// The improve_tool spec needs the component; recreated/adopted tools
		// have no content_components row — record the miss honestly instead.
		var componentID, pageID string
		_ = params.DB.QueryRowContext(ctx, `
			SELECT cc.id::text, COALESCE(p.id::text, '')
			FROM content_components cc
			LEFT JOIN page_components pc ON pc.component_id = cc.id
			LEFT JOIN pages p ON p.id = pc.page_id AND p.site_id = $2::uuid
			WHERE cc.function = $1 AND cc.is_active
			LIMIT 1`, function, siteID).Scan(&componentID, &pageID)

		if componentID != "" && stuck {
			// Two cycles have already failed at these criteria. A third would be
			// the same fix again — the failure mode this guard exists to stop —
			// so hand it to a human and stop the auto-loop for this criterion.
			// Not deduped away by idx_swi_dedup: needs_human_review is an OPEN
			// status, so the weekly re-verdict finds this item still standing and
			// ON CONFLICT DO NOTHING keeps it as raised rather than duplicating it.
			spec := map[string]interface{}{
				"component_id":      componentID,
				"check":             "tool_acceptance_tier4",
				"issue":             issue,
				"failing_checks":    v.FailedIDs,
				"failing_instances": v.Failed,
				"fix_cycles_spent":  attempts,
				"why_escalated": fmt.Sprintf(
					"%d improve_tool cycle(s) since the last passing Tier-4 verdict left %s still failing; the one-shot fixer is not converging on this defect",
					attempts, strings.Join(v.FailedIDs, ", ")),
			}
			if v.ForcedBy != "" {
				spec["overflow_forced_by"] = v.ForcedBy
			}
			if v.ForcedReason != "" {
				spec["overflow_fix_hint"] = v.ForcedReason
			}
			if shots := shotsForSpec(v.Shots, ""); len(shots) > 0 {
				spec["screenshots"] = shots
			}
			if pageID != "" {
				spec["page_id"] = pageID
			}
			specJSON, _ := json.Marshal(spec)
			_, err := params.DB.ExecContext(ctx, `
				INSERT INTO site_work_items (
					site_id, source, pipeline, item_type, severity, summary,
					priority, handler_agent, status, created_by, spec, item_key, batch_id
				) VALUES ($1::uuid, 'acceptance', 'build', 'acceptance_stuck',
				          'medium', $2, 20, 'human-review', 'needs_human_review',
				          'tool-acceptance-agent', $3::jsonb, $4, $5::uuid)
				ON CONFLICT DO NOTHING`,
				siteID,
				fmt.Sprintf("Tier-4 acceptance not converging for %s after %d fix cycle(s): %s",
					function, attempts, first(v.Details)),
				string(specJSON),
				fmt.Sprintf("acceptance_stuck:%s:%s", function, siteID),
				uuid.NewString(),
			)
			if err != nil {
				logger.Warn("judge: acceptance_stuck insert failed", zap.Error(err))
			} else {
				escalated = true
			}
		} else if componentID != "" {
			spec := map[string]interface{}{
				"component_id": componentID,
				"check":        "tool_acceptance_tier4",
				"issue":        issue,
				// Bare criteria ids — the fixer matches these against the PLAN's
				// checks (same shape Tier 2 emits). failing_instances keeps the
				// profile detail: a check that fails ONLY on mobile is a mobile bug.
				"failing_checks":    v.FailedIDs,
				"failing_instances": v.Failed,
				"acceptance_test":   json.RawMessage(criteriaOrNull(criteria)),
			}
			// Drill-down attribution: point the fixer at the element that forces
			// the width and why, not the ancestor that inherited it (bugs_open/010).
			if v.ForcedBy != "" {
				spec["overflow_forced_by"] = v.ForcedBy
			}
			if v.ForcedReason != "" {
				spec["overflow_fix_hint"] = v.ForcedReason
			}
			if shots := shotsForSpec(v.Shots, ""); len(shots) > 0 {
				spec["screenshots"] = shots // P3 evidence: what the page looked like when it failed
			}
			if pageID != "" {
				spec["page_id"] = pageID
			}
			specJSON, _ := json.Marshal(spec)
			_, err := params.DB.ExecContext(ctx, `
				INSERT INTO site_work_items (
					site_id, source, pipeline, item_type, severity, summary,
					priority, handler_agent, status, created_by, spec, item_key, batch_id
				) VALUES ($1::uuid, 'acceptance', 'build', 'improve_tool',
				          'medium', $2, 60, 'tool-improver', 'detected',
				          'tool-acceptance-agent', $3::jsonb, $4, $5::uuid)
				ON CONFLICT DO NOTHING`,
				siteID,
				fmt.Sprintf("Tier-4 acceptance failed for %s: %s", function, first(v.Details)),
				string(specJSON),
				fmt.Sprintf("acceptance_fail:%s:%s", function, siteID),
				uuid.NewString(),
			)
			if err != nil {
				logger.Warn("judge: improve_tool insert failed", zap.Error(err))
			} else {
				itemCreated = true
			}
		} else {
			logger.Info("judge: no content_components row for function — improve_tool item not created (recreated/adopted tool; route manually)",
				zap.String("function", function))
		}
	}

	// The note is written LAST and from the outcome, not from the intent: it is
	// the loop's own durable record, and a note claiming a fix was queued when
	// the insert missed is the "trust the status, not the artefact" failure this
	// codebase keeps paying for. Both branches can miss — a recreated or adopted
	// tool has no content_components row at all.
	var fixLine string
	switch {
	case escalated:
		fixLine = fmt.Sprintf(
			"NOT auto-fixed — %d previous improve_tool cycle(s) failed to turn %s green, so this is escalated to human review (acceptance_stuck) instead of a %d%s identical attempt",
			attempts, strings.Join(v.FailedIDs, ", "), attempts+1, ordinalSuffix(attempts+1))
	case itemCreated:
		fixLine = "improve_tool item created carrying the criteria as acceptance_test"
	case stuck:
		fixLine = fmt.Sprintf(
			"NOT auto-fixed — %d previous improve_tool cycle(s) failed to turn %s green, and the escalation item could NOT be raised (no active content_components row for this function, or the insert failed); route this manually",
			attempts, strings.Join(v.FailedIDs, ", "))
	default:
		fixLine = "none — no improve_tool item could be created (no active content_components row for this function, or the insert failed); route this manually"
	}
	body := fmt.Sprintf(`## Tier-4 acceptance FAILED — %s
Observed: %d of %d evaluated checks failed in headless Chromium%s: %s%s%s
Root cause: not diagnosed at this tier (behavioural run; the fixer loads PLAN+NOTES first)
Fix: %s
Verified: n/a — failing run recorded
Categories: acceptance-fail`,
		function, len(v.Failed), len(v.Results), profilesPhrase(v.Profiles), issue, chromeLine,
		evidenceLine(v.Shots), fixLine)
	if _, err := insertDocNote(ctx, params.DB, "tool", function, siteID, body,
		`["acceptance-fail"]`, "tool-acceptance", sourceAgent, "", "tool-acceptance-agent"); err != nil {
		logger.Warn("judge: acceptance-fail note insert failed", zap.Error(err))
	}

	logger.Info("judge: acceptance FAILED",
		zap.String("function", function),
		zap.Strings("failed", v.Failed),
		zap.Bool("improve_tool_created", itemCreated),
		zap.Bool("escalated", escalated),
		zap.Int("fix_cycles_spent", attempts))
	return map[string]interface{}{
		"all_passed": false, "passed": len(v.Passed), "failed": len(v.Failed),
		"failing_checks": v.FailedIDs, "failing_instances": v.Failed,
		"improve_tool_created": itemCreated,
		"escalated":            escalated,
		"fix_cycles_spent":     attempts,
	}, nil
}

// convergenceAttempts counts the improve_tool cycles that have already tried,
// and failed, to turn THESE criteria green — the count the loop was missing
// (bugs_open/010 b). Each acceptance verdict raises a FRESH item, so the
// fixer's per-item max_fix_attempts never engages across cycles; without this
// count nothing distinguishes a first attempt from a fourth identical one.
//
// Bounded three ways, so it measures non-convergence rather than history:
//
//   - terminal attempts only — an item still open is the CURRENT cycle, not a
//     past failure, and counting it would escalate a loop mid-flight;
//   - only attempts raised since the tool last PASSED Tier 4 — a green verdict
//     resets the count, so a criterion that regresses weeks later is treated as
//     a new defect rather than inheriting an old tally;
//   - only attempts overlapping the criteria failing NOW — a fixer that fixed X
//     and left Y unfixed has not yet failed at Y twice.
//
// The item_key matches the one the judge writes below, so the count sees the
// judge's own history and nothing else. jsonb_typeof guards the array cast:
// older improve_tool rows raised by other paths carry no failing_checks.
func convergenceAttempts(ctx context.Context, db *sql.DB, function, siteID string, failedIDs []string) (int, error) {
	if len(failedIDs) == 0 {
		return 0, nil
	}
	var n int
	err := db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM site_work_items w
		WHERE w.site_id = $1::uuid
		  AND w.item_type = 'improve_tool'
		  AND w.item_key = $2
		  AND w.status IN ('complete', 'failed')
		  AND w.created_at > COALESCE((
		        SELECT max(n.created_at) FROM doc_notes n
		        WHERE n.subject_type = 'tool' AND n.subject_key = $3
		          AND n.source = 'tool-acceptance'
		          AND n.categories @> '["acceptance-run"]'::jsonb
		      ), '-infinity'::timestamptz)
		  AND jsonb_typeof(w.spec->'failing_checks') = 'array'
		  AND EXISTS (
		        SELECT 1 FROM jsonb_array_elements_text(w.spec->'failing_checks') e
		        WHERE e = ANY($4::text[]))`,
		siteID,
		fmt.Sprintf("acceptance_fail:%s:%s", function, siteID),
		function,
		toPGTextArrayLiteral(failedIDs),
	).Scan(&n)
	return n, err
}

// ordinalSuffix renders 1st/2nd/3rd/4th for a note that has to read like prose.
func ordinalSuffix(n int) string {
	if n%100 >= 11 && n%100 <= 13 {
		return "th"
	}
	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	}
	return "th"
}

// chromeForcedHint appends the drill-down attribution to a fix suggestion when
// the adapter located an element deeper than the widest offender that forces the
// width (bugs_open/010). Empty when the widest offender is itself the cause.
func chromeForcedHint(c chromeFailure) string {
	if c.ForcedBy == "" && c.ForcedReason == "" {
		return ""
	}
	hint := " The width is actually forced by " + c.ForcedBy
	if c.ForcedReason != "" {
		hint += " — " + c.ForcedReason
	}
	return hint + "; fix THAT element, not just its container."
}

// chromeSummary renders site-chrome failures for a note line.
func chromeSummary(cf []chromeFailure) string {
	parts := make([]string, 0, len(cf))
	for _, c := range cf {
		part := checkLabel(c.CheckID, c.Profile) + " — " + c.Culprit
		if c.Component != "" {
			part += " in " + c.Component
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, "; ")
}

// routeChromeFailures raises ONE responsive_fix item per (component, profile)
// for defects the adapter proved lie outside the tool — the site template's
// problem, handled by component-template-fixer (the established route for
// responsive defects), NOT by tool-improver.
//
// The dedup key deliberately excludes the tool: an overflowing site footer is
// ONE site defect, not one per tool that happens to sit on the page. idx_swi_dedup
// (site_id, item_key) collapses repeat reports while the item is open, and lets
// it be raised afresh if the defect returns after a fix.
func routeChromeFailures(ctx context.Context, params ActionParams, logger *zap.Logger,
	failures []chromeFailure, shots []screenshotRef, function, siteID, sourceAgent string) int {

	if len(failures) == 0 || params.DB == nil || siteID == "" {
		return 0
	}
	created := 0
	for _, cf := range failures {
		component := cf.Component
		if component == "" {
			component = "site-chrome"
		}
		itemKey := fmt.Sprintf("chrome_overflow:%s:%s", component, cf.Profile)

		// fix_type MUST be chrome_overflow_fix, not the legacy responsive_fix:
		// that path defaults to the header slot and injects canned header-nav CSS,
		// so it "fixes" the wrong thing and reports success (observed 2026-07-14 —
		// it patched vonc's HEADER for a FOOTER defect and returned fixed=true).
		// slot_name and overflow_selector are what make the fix targeted; the
		// fixer refuses to guess without them.
		spec := map[string]interface{}{
			"category":     "responsive",
			"fix_type":     "chrome_overflow_fix",
			"slot_name":    cf.Slot,
			"audit_source": "tool-acceptance-tier4",
			"description": fmt.Sprintf(
				"The page overflows horizontally on %s. The widest offending element is %s, inside %s — OUTSIDE the tool's container, so this is a site-template defect, not a tool defect. Found while running Tier-4 acceptance for %s (%s), but it affects every page that renders this chrome.",
				cf.Profile, cf.Culprit, component, function, cf.URL),
			"suggestion": fmt.Sprintf(
				"Constrain %s to the viewport at mobile widths: let it wrap or shrink (flex-wrap / max-width:100%%) so no descendant exceeds the viewport at 390px.%s",
				cf.Selector, chromeForcedHint(cf)),
			"overflow_selector":  cf.Selector,
			"current_value":      cf.Culprit,
			"acceptance_test":    fmt.Sprintf("At 390px viewport width, %s document.scrollWidth <= document.clientWidth (no horizontal overflow)", cf.URL),
			"affected_component": component,
			"page_name":          "global",
			"max_fix_attempts":   2,
			"original_pipeline":  "build",
			"original_domain":    "build",
			"found_via":          map[string]interface{}{"tool": function, "check": cf.CheckID, "profile": cf.Profile},
		}
		// Drill-down attribution (bugs_open/010): the element deeper than the
		// widest offender that actually forces the width, when the adapter found one.
		if cf.ForcedBy != "" {
			spec["overflow_forced_by"] = cf.ForcedBy
		}
		if cf.ForcedReason != "" {
			spec["overflow_fix_hint"] = cf.ForcedReason
		}
		// P3 evidence: the failing profile's full-page screenshot shows the
		// chrome defect (the whole page was photographed, footer included).
		if evidence := shotsForSpec(shots, cf.Profile); len(evidence) > 0 {
			spec["screenshots"] = evidence
		}
		specJSON, _ := json.Marshal(spec)

		_, err := params.DB.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				priority, handler_agent, status, created_by, spec, item_key, batch_id
			) VALUES ($1::uuid, 'acceptance', 'build', 'responsive_fix',
			          'medium', $2, 40, 'component-template-fixer', 'detected',
			          'tool-acceptance-agent', $3::jsonb, $4, $5::uuid)
			ON CONFLICT DO NOTHING`,
			siteID,
			fmt.Sprintf("Horizontal overflow on %s from site chrome: %s in %s", cf.Profile, cf.Culprit, component),
			string(specJSON),
			itemKey,
			uuid.NewString(),
		)
		if err != nil {
			logger.Warn("judge: responsive_fix (site chrome) insert failed",
				zap.String("component", component), zap.Error(err))
			continue
		}
		created++
		logger.Info("judge: site-chrome defect routed to component-template-fixer (NOT the tool)",
			zap.String("component", component), zap.String("culprit", cf.Culprit),
			zap.String("profile", cf.Profile), zap.String("item_key", itemKey))
	}
	return created
}

// profilesPhrase renders " across profiles: desktop, mobile" (empty when the
// adapter reported none — a pre-P1 runner).
func profilesPhrase(profiles []string) string {
	if len(profiles) == 0 {
		return ""
	}
	return " across profiles: " + strings.Join(profiles, ", ")
}

func orNone(ss []string) []string {
	if len(ss) == 0 {
		return []string{"none"}
	}
	return ss
}

func first(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	return ss[0]
}

func criteriaOrNull(criteria string) string {
	if strings.TrimSpace(criteria) == "" {
		return "null"
	}
	return criteria
}
