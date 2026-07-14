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
	Failed    []string
	FailedIDs []string
	Details   []string // "id@profile: detail" for each failed check
	SkipList  []string
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
		} else {
			v.Failed = append(v.Failed, label)
			v.FailedIDs = appendUnique(v.FailedIDs, id)
			v.Details = append(v.Details, label+": "+detail)
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
	return v
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

	allPassed := len(v.Failed) == 0
	if allPassed {
		body := fmt.Sprintf(`## Tier-4 acceptance PASSED — %s
Observed: all %d evaluated checks passed in headless Chromium%s (%d skipped: %s).
Root cause: not-applicable
Fix: none required
Verified: browser-runner-adapter run; checks (id@profile): %s
Categories: acceptance-run`,
			function, len(v.Passed), profilesPhrase(v.Profiles),
			len(v.SkipList), strings.Join(orNone(v.SkipList), ", "),
			strings.Join(v.Passed, ", "))
		if _, err := insertDocNote(ctx, params.DB, "tool", function, siteID, body,
			`["acceptance-run"]`, "tool-acceptance", sourceAgent, "", "tool-acceptance-agent"); err != nil {
			logger.Warn("judge: acceptance-run note insert failed", zap.Error(err))
		}
		logger.Info("judge: acceptance PASSED",
			zap.String("function", function), zap.Int("passed", len(v.Passed)))
		return map[string]interface{}{
			"all_passed": true, "passed": len(v.Passed), "failed": 0, "skipped_checks": v.SkipList,
		}, nil
	}

	// Failures: one acceptance-fail note + ONE improve_tool item carrying the
	// criteria as acceptance_test (findings pattern; bounded by the fixer's
	// max_fix_attempts convention).
	issue := strings.Join(v.Details, "; ")
	body := fmt.Sprintf(`## Tier-4 acceptance FAILED — %s
Observed: %d of %d evaluated checks failed in headless Chromium%s: %s
Root cause: not diagnosed at this tier (behavioural run; the fixer loads PLAN+NOTES first)
Fix: improve_tool item created carrying the criteria as acceptance_test
Verified: n/a — failing run recorded
Categories: acceptance-fail`,
		function, len(v.Failed), len(v.Results), profilesPhrase(v.Profiles), issue)
	if _, err := insertDocNote(ctx, params.DB, "tool", function, siteID, body,
		`["acceptance-fail"]`, "tool-acceptance", sourceAgent, "", "tool-acceptance-agent"); err != nil {
		logger.Warn("judge: acceptance-fail note insert failed", zap.Error(err))
	}

	criteria := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "criteria_field", "doc_context.criteria_json"))

	itemCreated := false
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

		if componentID != "" {
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

	logger.Info("judge: acceptance FAILED",
		zap.String("function", function),
		zap.Strings("failed", v.Failed),
		zap.Bool("improve_tool_created", itemCreated))
	return map[string]interface{}{
		"all_passed": false, "passed": len(v.Passed), "failed": len(v.Failed),
		"failing_checks": v.FailedIDs, "failing_instances": v.Failed,
		"improve_tool_created": itemCreated,
	}, nil
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
