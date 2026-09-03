// FILE: platform/orchestration/actions/tool_acceptance_actions.go
//
// The actions behind tool-acceptance-agent — the orchestrator that makes
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
//     concern, never a fake browser pass. Resolves the page by NAME:
//     `pages.name IN (function, 'tool-'||function)` — correct for a tool,
//     which is placed on exactly one page named after itself.
//
//   request_component_browser_run — the same request, for a section
//     component (DOC-068 / staged_component_build P2). A component's
//     placement is many-to-many (`page_components`), so there is no name to
//     resolve by; the target page is given explicitly (page_id_field) and
//     CHECKED against `page_components`/`content_components.function` rather
//     than derived. A sibling action rather than a branch on
//     request_browser_run so the tool path's existing guarantee needs no
//     re-proving — see PLAN_2026-07-30_staged_component_build.md D9. Shares
//     dispatchBrowserRun (envelope build + send) with the tool action; only
//     page resolution differs.
//
//   judge_acceptance_results — turns the adapter's reply into the loop's
//     artifacts: all pass → one acceptance-run doc_note; any fail → an
//     acceptance-fail doc_note + ONE improve_tool work item carrying the
//     criteria as acceptance_test and the failing check ids (the fixer loads
//     PLAN+NOTES first, per Task 4). Reads the reply through a fallback chain
//     of paths (003 action-level defense) because awaited-response shapes
//     vary across the codebase (.response.data vs flattened). Serves both
//     request actions unchanged — it keys off `function`, never off how the
//     page was resolved. A fence carrying top-level `no_auto_fix: true` never
//     reaches tool-improver at all: a failing verdict goes straight to the
//     acceptance_stuck human-review path instead (bugs_open/126 candidate 2 —
//     an automated rewriter aimed at a consent gate can only pass by deleting
//     it). Default-OFF; a fence that says nothing behaves exactly as before.

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

var RequestComponentBrowserRunInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"function_field", "criteria_field", "site_id_field", "domain_field",
		"page_id_field", "url_field", "profiles",
	},
	Defaults: map[string]interface{}{
		"function_field": "input_data.spec.function",
		"criteria_field": "doc_context.criteria_json",
		"site_id_field":  "site_record.site_id",
		"domain_field":   "site_record.domain",
		"page_id_field":  "input_data.spec.page_id",
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
	datahelpers.RegisterActionInputSpec("request_component_browser_run", RequestComponentBrowserRunInputSpec)
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
	// table. Generator tools: page name == function. PORTED tools (eligible
	// since tool_eligibility.go widened the ladder): the subject key is the
	// page name MINUS its 'tool-' prefix, so the page is 'tool-'||function —
	// without the second candidate every acceptance run the widened due-sweep
	// emits for a ported tool would hard-error right here, on the ladder's
	// first attempt to look at the population it was widened to see. The exact
	// name wins the tie so a generator tool named like a ported page cannot be
	// shadowed.
	pageURL := ""
	if uf := datahelpers.GetStringField(config, "url_field", ""); uf != "" {
		pageURL = datahelpers.ExtractNestedFieldString(params.CollectedData, uf)
	}
	if pageURL == "" && params.DB != nil && siteID != "" {
		// $2 is cast explicitly at every use: mixing a bare $2 with
		// 'tool-' || $2 makes Postgres deduce inconsistent types for the one
		// parameter (SQLSTATE 42P08). The first live exercise of this lookup —
		// the smart-contrast pilot — failed on exactly that, which go build
		// could never have caught.
		err := params.DB.QueryRowContext(ctx, `
			SELECT COALESCE(url, '') FROM pages
			WHERE site_id = $1::uuid AND status = 'active'
			  AND name IN ($2::text, 'tool-' || $2::text)
			ORDER BY (name = $2::text) DESC
			LIMIT 1
		`, siteID, function).Scan(&pageURL)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("request_browser_run: page lookup failed: %w", err)
		}
	}
	if pageURL == "" {
		return nil, fmt.Errorf("request_browser_run: no deployed page URL for function %q on site %s", function, siteID)
	}

	return dispatchBrowserRun(ctx, params, logger, config, function, siteID, domain, pageURL, criteria)
}

// dispatchBrowserRun builds the run_checks envelope and sends it to the
// browser-runner adapter — the part of a browser-run request that does not
// depend on how the target page was resolved. Shared by
// RequestBrowserRunAction (tool: page resolved by name) and
// RequestComponentBrowserRunAction (component: page resolved by explicit
// placement) rather than duplicated, the same reasoning envelopePaths above
// is already built on: two copies of an envelope builder can only drift.
func dispatchBrowserRun(ctx context.Context, params ActionParams, logger *zap.Logger, config map[string]interface{}, function, siteID, domain, pageURL, criteria string) (interface{}, error) {
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

	// capture_renders asks the adapter to photograph a run that PASSES, not only
	// one that fails (TL-035). Opt-in per step, default false = the behaviour
	// every existing config already has.
	//
	// It is here because the adapter half shipped switched off: the camera exists
	// and nothing asks it to fire on a clean page, so the defect class that
	// reaches a human — text flush against a border, links off their baseline,
	// chart labels overprinting — still lands on a page where every check passes
	// and therefore leaves no picture behind. A render is a look, never a verdict:
	// it lands in a separate list the judge reads for the note only.
	captureRenders := datahelpers.GetBoolField(config, "capture_renders", false)

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
				"run_id":          params.ExecutionContext.OrchestrationID,
				"urls":            []string{fullURL},
				"profiles":        profiles,
				"criteria_json":   criteria,
				"function":        function,
				"site_id":         siteID,
				"capture_renders": captureRenders,
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
		return nil, fmt.Errorf("dispatchBrowserRun: marshal request: %w", err)
	}

	logger.Info("dispatchBrowserRun: sending to browser-runner adapter",
		zap.String("topic", browserRunnerTopic),
		zap.String("request_id", newRequestID),
		zap.String("url", fullURL),
		zap.String("function", function),
		zap.Strings("profiles", profiles),
		zap.Bool("capture_renders", captureRenders))

	if err := params.Producer.ProduceWithValidation(ctx, browserRunnerTopic, headers,
		[]byte(params.ExecutionContext.CorrelationID), messageBytes); err != nil {
		return nil, fmt.Errorf("dispatchBrowserRun: send to browser-runner adapter: %w", err)
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

// ── request_component_browser_run ───────────────────────────────────────────

// RequestComponentBrowserRunAction is request_browser_run's sibling for a
// section component (DOC-068). A tool's page names itself after the tool
// (pages.name IN (function, 'tool-'||function)); a component's placement is
// many-to-many (page_components), so there is no name to resolve by — the
// target page is given explicitly (page_id_field) and CHECKED against the
// placement, never derived from function alone. See
// staged_component_build/PLAN_2026-07-30... D9 for why this is a sibling
// action rather than a branch on RequestBrowserRunAction.
func RequestComponentBrowserRunAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "request_component_browser_run"))
	config := params.StepConfig.Config

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	function := resolveWithFallbacks(params.CollectedData,
		datahelpers.GetStringField(config, "function_field", "input_data.spec.function"),
		"input_data.function")
	if function == "" {
		return nil, fmt.Errorf("request_component_browser_run: no function (input_data.spec.function / input_data.function)")
	}

	criteria := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "criteria_field", "doc_context.criteria_json"))
	if strings.TrimSpace(criteria) == "" {
		// Same "no fake pass" rule as the tool action: an undocumented
		// component is Tier-2's needs_criteria concern, never a fake browser
		// pass.
		logger.Info("request_component_browser_run: no criteria in the current PLAN — skipping (no fake pass)",
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
	pageID := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "page_id_field", "input_data.spec.page_id"))
	if pageID == "" {
		return nil, fmt.Errorf("request_component_browser_run: no page_id (page_id_field) — a component can be placed on more than one page, so the target page cannot be inferred from function alone")
	}

	// Resolve the deployed URL: explicit config path first, else the
	// page_components/content_components join. The given page_id is
	// ASSERTED by the caller and then CHECKED against the real placement row
	// (a stale or wrong page_id fails closed with an error, not a silent
	// pass against the wrong page).
	pageURL := ""
	if uf := datahelpers.GetStringField(config, "url_field", ""); uf != "" {
		pageURL = datahelpers.ExtractNestedFieldString(params.CollectedData, uf)
	}
	if pageURL == "" && params.DB != nil {
		err := params.DB.QueryRowContext(ctx, `
			SELECT COALESCE(p.url, '') FROM pages p
			JOIN page_components pc ON pc.page_id = p.id
			JOIN content_components cc ON cc.id = pc.component_id
			WHERE pc.page_id = $1::uuid AND cc.function = $2::text
			  AND p.status = 'active'
			LIMIT 1
		`, pageID, function).Scan(&pageURL)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("request_component_browser_run: placement lookup failed: %w", err)
		}
	}
	if pageURL == "" {
		return nil, fmt.Errorf("request_component_browser_run: component %q is not placed on page %s (or that page is inactive)", function, pageID)
	}

	return dispatchBrowserRun(ctx, params, logger, config, function, siteID, domain, pageURL, criteria)
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
	// Renders: one full-page screenshot per PASSING (url, profile), present only
	// when the step set capture_renders (TL-035). Never evidence and never part
	// of a verdict — a picture of a page that passed, so a human or a vision
	// check can look at it without a failure having to justify the look.
	//
	// A run can produce BOTH lists: two profiles where desktop fails and mobile
	// passes files the desktop shot under Shots and the mobile one here. So this
	// is not "the pass list" — it is the per-run-that-passed list, which is why
	// the failure path below reports it too rather than assuming it is empty.
	Renders []screenshotRef
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
	Profile  string
	URL      string
	URI      string
	ViewURL  string
	Viewport string   // layout viewport + device scale, e.g. "390x844@3x" (empty on pre-viewport adapters)
	Stage    string   // "landing" = photographed before checks drove the page; empty = driven state (all failure evidence, and renders from older adapters)
	Failing  []string // id@profile instances the screenshot evidences
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

	// Which envelope the reply arrived in is discovered ONCE, from results, and
	// then reused for the shot lists — see envelopePaths.
	raw, envIdx := extractAtFirstMatch(collected, envelopePaths(field, "results"))
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

	v.Shots = extractShotList(collected, field, "screenshots", envIdx)
	// Renders arrive in their OWN key and are parsed by the same code — the two
	// lists differ in what they mean, never in their shape. Empty unless the
	// request opted in, and empty on any adapter built before TL-035, so the
	// absence of the key is indistinguishable from "not asked for", which is the
	// correct reading either way.
	v.Renders = extractShotList(collected, field, "renders", envIdx)
	return v
}

// envelopePaths gives the reply shapes this codebase produces, most specific
// first, for one key. ONE list, used for results and for both shot lists.
//
// It is a function rather than three literal copies because of a council
// objection (submission 2c895dd1, bug_historian, medium) worth restating: a
// reader that recognises four dialects and none other FAILS OPEN on the fifth —
// it returns an empty list, which for renders is indistinguishable from "not
// requested". Three separate copies of the list made that a matter of keeping
// three things in step, which is the drift this platform keeps paying for.
func envelopePaths(field, key string) []string {
	return []string{
		field + ".response.data." + key,
		field + ".response." + key,
		field + ".data." + key,
		field + "." + key,
	}
}

// extractAtFirstMatch returns the first non-nil value among paths, and WHICH
// path matched (-1 if none did).
func extractAtFirstMatch(collected map[string]interface{}, paths []string) (interface{}, int) {
	for i, p := range paths {
		if v := datahelpers.ExtractNestedField(collected, p); v != nil {
			return v, i
		}
	}
	return nil, -1
}

// extractShotList parses one adapter screenshot list (screenshots | renders)
// out of the reply. Shared rather than duplicated: a render and a failure shot
// are the same ScreenshotRef on the wire, so a second copy of this parser could
// only ever drift from the first.
//
// envIdx is the envelope `results` was found in, and it is tried FIRST. That is
// what answers the fail-open objection at its root rather than by logging: the
// shot lists can no longer be read from a different envelope than the results
// they belong to, so an adapter shape this function does not recognise cannot
// silently drop renders WITHOUT ALSO hiding results — and empty results is a
// hard error one caller up ("no results at %q"). A silent drop is therefore
// unrepresentable rather than merely reported. The remaining paths are still
// tried afterwards, so this can only ever find more than the old order did.
func extractShotList(collected map[string]interface{}, field, key string, envIdx int) []screenshotRef {
	paths := envelopePaths(field, key)
	if envIdx >= 0 && envIdx < len(paths) {
		paths = append([]string{paths[envIdx]}, paths...)
	}
	raw, _ := extractAtFirstMatch(collected, paths)
	items, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	var out []screenshotRef
	for _, it := range items {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		ref := screenshotRef{}
		ref.Profile, _ = m["profile"].(string)
		ref.URL, _ = m["url"].(string)
		ref.URI, _ = m["uri"].(string)
		ref.ViewURL, _ = m["view_url"].(string)
		ref.Viewport, _ = m["viewport"].(string)
		ref.Stage, _ = m["stage"].(string)
		if fc, ok := m["failing_checks"].([]interface{}); ok {
			for _, f := range fc {
				if s, ok := f.(string); ok {
					ref.Failing = append(ref.Failing, s)
				}
			}
		}
		if ref.URI != "" {
			out = append(out, ref)
		}
	}
	return out
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
		parts = append(parts, s.URI+profileTag(s))
	}
	return "\nEvidence: full-page screenshot(s) at failure: " + strings.Join(parts, "; ")
}

// renderLine renders the TL-035 renders for a note body. Deliberately NOT
// evidenceLine with a different string: the word "Evidence" is a claim about a
// failure, and every render here is a photograph of a run that passed. Durable
// s3:// URIs only, for the same reason as evidenceLine — a note body is loaded
// into LLM prompt contexts, where a presigned URL is hundreds of characters of
// expiring signature that will be stale by the time anyone reads it.
func renderLine(renders []screenshotRef) string {
	if len(renders) == 0 {
		return ""
	}
	parts := make([]string, 0, len(renders))
	for _, s := range renders {
		parts = append(parts, s.URI+profileTag(s))
	}
	return "\nRendered: full-page screenshot(s) of the page as it passed: " + strings.Join(parts, "; ") +
		"\nNote: a render is a look, not a verdict — nothing here asserts the page is free of defects no check covers."
}

// profileTag labels one shot for a note line: " (desktop 1366x900@1x, landing
// state)", or " (desktop)" from an adapter that predates the newer fields. The
// viewport matters because the PNG's pixel width is viewport × device scale —
// a 22,491px-tall, 1170px-wide render is a phone at 3x, and nothing else on
// the line says so. The stage matters because a render captured after the
// checks drove the page shows a state no visitor sees (a post-Clear empty
// panel read as a false bug) — per-ref, so a reader never needs deploy dates.
func profileTag(s screenshotRef) string {
	tag := s.Profile
	if s.Viewport != "" {
		if tag != "" {
			tag += " "
		}
		tag += s.Viewport
	}
	if s.Stage == "landing" {
		if tag != "" {
			tag += ", "
		}
		tag += "landing state"
	}
	if tag == "" {
		return ""
	}
	return " (" + tag + ")"
}

// shotsForSpec renders screenshot refs for a work item's spec (uri + view_url).
// profile filters to one profile's evidence; "" keeps everything.
func shotsForSpec(shots []screenshotRef, profile string) []map[string]interface{} {
	var out []map[string]interface{}
	for _, s := range shots {
		if profile != "" && s.Profile != "" && s.Profile != profile {
			continue
		}
		entry := map[string]interface{}{
			"profile": s.Profile, "uri": s.URI, "view_url": s.ViewURL,
		}
		if s.Viewport != "" {
			entry["viewport"] = s.Viewport
		}
		out = append(out, entry)
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
		// WHAT DID THIS PASS ACTUALLY COVER? (bugs_open/449.)
		//
		// A verdict is only as strong as its fence, and the two families of check
		// in this estate make opposite promises: the generated ones assert that
		// the tool is ALIVE, the operator-written ones that its numbers are RIGHT,
		// and measured 2026-09-03 not one tool anywhere was covered by both. 115
		// of `tool-generator`'s 186 current fences assert no expected value of any
		// kind, so "Tier-4 PASSED" on one of them means the page loaded and
		// something appeared when we clicked — and it is read as "the calculator
		// works". Nothing in the record distinguished the two, which is the whole
		// damage: the weak fence is only the cause.
		//
		// So the verdict now states its own scope. This changes no outcome — a
		// pass still passes — it removes the overclaim, which is the part that
		// can be fixed for all 186 at once, today, with no author cooperation and
		// no backfill. It cannot go stale either: the grade is derived from the
		// fence at run time, so a fence that gets weaker gets a weaker verdict
		// with nobody remembering to do anything.
		passCriteria := datahelpers.ExtractNestedFieldString(params.CollectedData,
			datahelpers.GetStringField(config, "criteria_field", "doc_context.criteria_json"))
		assertions := summariseCriteriaValueAssertions(passCriteria)

		// The TOOL passed. The page may still carry a site-chrome defect — say so
		// plainly rather than let "PASSED" imply the page is clean.
		rootCause := "not-applicable"
		fix := "none required"
		if len(v.Chrome) > 0 {
			rootCause = "site chrome, not this tool: " + chromeSummary(v.Chrome)
			fix = fmt.Sprintf("%d responsive_fix item(s) raised for the site template (handler component-template-fixer); the tool itself needs no change", chromeRouted)
		}
		// Evidence exists on a pass ONLY when the page failed for site-chrome
		// reasons — the screenshot shows the chrome defect, not the tool. A
		// RENDER, by contrast, is the ordinary case on a clean pass whenever the
		// step opted in: it is what the page looked like when nothing failed, and
		// it is the only artefact of this run a human can actually look at.
		body := fmt.Sprintf(`## Tier-4 acceptance PASSED — %s
Observed: all %d of the tool's own checks passed in headless Chromium%s (%d skipped: %s).%s%s
Scope of this verdict: %s
Root cause: %s
Fix: %s
Verified: browser-runner-adapter run; checks (id@profile): %s
Categories: acceptance-run`,
			function, len(v.Passed), profilesPhrase(v.Profiles),
			len(v.SkipList), strings.Join(orNone(v.SkipList), ", "),
			evidenceLine(v.Shots), renderLine(v.Renders),
			criteriaAssertionPhrase(assertions),
			rootCause, fix,
			strings.Join(v.Passed, ", "))
		if _, err := insertDocNote(ctx, params.DB, "tool", function, siteID, body,
			`["acceptance-run"]`, "tool-acceptance", sourceAgent, "", "tool-acceptance-agent"); err != nil {
			logger.Warn("judge: acceptance-run note insert failed", zap.Error(err))
		}
		logger.Info("judge: acceptance PASSED",
			zap.String("function", function), zap.Int("passed", len(v.Passed)),
			zap.Int("site_chrome_items", chromeRouted),
			zap.String("assertion_grade", assertions.Grade()),
			zap.Int("value_assertions", assertions.Total()))
		out := map[string]interface{}{
			"all_passed": true, "passed": len(v.Passed), "failed": 0, "skipped_checks": v.SkipList,
			"site_chrome_failures": len(v.Chrome), "site_chrome_items_created": chromeRouted,
			// bugs_open/449. `assertion_grade` is none | pattern | exact and is
			// the field to read before quoting a PASS: `none` means the tool
			// responded and nothing was checked about what it computed.
			"assertion_grade":        assertions.Grade(),
			"value_assertions":       assertions.Total(),
			"exact_value_assertions": assertions.Exact,
		}
		// Present ONLY in the state worth acting on, so a consumer can branch on
		// the key's presence rather than parse a grade string — and so a fence
		// that DOES assert values carries no extra keys at all.
		if assertions.AssertsNoValue() {
			out["verdict_scope"] = "liveness_only"
		}
		return out, nil
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

	// bugs_open/126 candidate 2: a fence may declare itself OUT of the auto-fix
	// loop. Some markup is load-bearing for reasons no acceptance run can see —
	// a consent gate or a disclaimer whose wording and placement are what a
	// negligent-misstatement analysis rests on — and the only way an automated
	// rewriter can turn a fence green over such an element is to weaken or
	// delete it. So when the criteria carry `no_auto_fix: true`, a FAILING
	// verdict goes to the human-review path below instead of ever dispatching
	// tool-improver, no matter how few cycles have been spent.
	//
	// Opt-in and default-OFF: a fence that says nothing behaves exactly as it
	// did before this existed (owner ruling 2026-08-02 — new authority on a
	// shared seam ships as a field with the unsafe default off, not as a rule in
	// a doc comment).
	noAutoFix, noAutoFixReason := parseNoAutoFix(criteria)

	itemCreated := false
	itemDeduped := false
	escalated := false
	portedRouted := false
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

		if componentID != "" && (stuck || noAutoFix) {
			// Two reasons converge on the SAME escalation, deliberately (owner
			// ruling 2026-08-02: converging producers onto one item_type is fine
			// provided the producer set and the item_key shape are written down):
			//
			//   stuck     — two cycles have already failed at these criteria, so
			//               a third would be the same fix again;
			//   noAutoFix — the fence forbids an automated rewrite at all
			//               (bugs_open/126), whatever the cycle count.
			//
			// Either way the auto-loop stops for this criterion and a human gets
			// the item. The why_escalated text below says WHICH of the two fired,
			// because the two ask a very different question of the reader.
			// idx_swi_dedup holds one open escalation per (site, key):
			// needs_human_review is a non-terminal status, so a later verdict
			// CONFLICTS with the standing item and DO UPDATE refreshes its count
			// and reason in place (a re-escalation should say "5 cycles", not the
			// stale "2" a DO NOTHING would leave) without minting a duplicate.
			// The spec is MERGED (site_work_items.spec || EXCLUDED.spec), not
			// replaced: a human triaging this item may have written an owner,
			// a note or an override into the same jsonb bag, and a full replace
			// would silently discard it on the next cycle. Our keys win on the
			// right of ||; any human-added keys survive.
			// Why this item exists, in the reader's terms — a human triaging it
			// needs to know whether the loop gave up (try something different) or
			// was never allowed to start (decide what, if anything, may change).
			whyEscalated := fmt.Sprintf(
				"%d improve_tool cycle(s) since the last passing Tier-4 verdict left %s still failing; the one-shot fixer is not converging on this defect",
				attempts, strings.Join(v.FailedIDs, ", "))
			summary := fmt.Sprintf("Tier-4 acceptance not converging for %s after %d fix cycle(s): %s",
				function, attempts, first(v.Details))
			if noAutoFix {
				whyEscalated = fmt.Sprintf(
					"this fence declares no_auto_fix: %s — %s failed and must NOT be handed to tool-improver, because an automated rewriter can only turn a fence like this green by weakening the markup the fence exists to protect (bugs_open/126); a human decides what may change",
					reasonOrUnstated(noAutoFixReason), strings.Join(v.FailedIDs, ", "))
				summary = fmt.Sprintf("Tier-4 acceptance failed for %s and the fence forbids auto-fix (no_auto_fix): %s",
					function, first(v.Details))
			}
			spec := map[string]interface{}{
				"component_id":      componentID,
				"check":             "tool_acceptance_tier4",
				"issue":             issue,
				"failing_checks":    v.FailedIDs,
				"failing_instances": v.Failed,
				"fix_cycles_spent":  attempts,
				"why_escalated":     whyEscalated,
			}
			if noAutoFix {
				// Carried in the spec as well as in the prose so a downstream
				// reader (or a query over the queue) can select these without
				// parsing why_escalated.
				spec["no_auto_fix"] = true
				if noAutoFixReason != "" {
					spec["no_auto_fix_reason"] = noAutoFixReason
				}
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
			// The arbiter predicate is the canonical dedup one (insertWorkItem's,
			// via the shared terminal-status list), so the ON CONFLICT target
			// matches idx_swi_dedup's partial index rather than risking 42P10.
			_, err := params.DB.ExecContext(ctx, fmt.Sprintf(`
				INSERT INTO site_work_items (
					site_id, source, pipeline, item_type, severity, summary,
					priority, handler_agent, status, created_by, spec, item_key, batch_id
				) VALUES ($1::uuid, 'acceptance', 'build', 'acceptance_stuck',
				          'medium', $2, 20, 'human-review', 'needs_human_review',
				          'tool-acceptance-agent', $3::jsonb, $4, $5::uuid)
				ON CONFLICT (site_id, item_key)
					WHERE item_key IS NOT NULL AND status NOT IN (%s)
				DO UPDATE SET spec = site_work_items.spec || EXCLUDED.spec,
				              summary = EXCLUDED.summary, updated_at = now()`,
				sqlInList(workItemTerminalStatuses)),
				siteID,
				summary,
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
			res, err := params.DB.ExecContext(ctx, `
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
			// ON CONFLICT DO NOTHING returns NO error when it inserts NOTHING:
			// idx_swi_dedup already holds an open item under this key, so the
			// verdict queued no work. Deriving "created" from err==nil is the
			// "trust the status, not the artefact" failure the note below is
			// written to avoid — the flag must come from rows affected.
			switch n, raErr := rowsAffected(res, err); {
			case err != nil:
				logger.Warn("judge: improve_tool insert failed", zap.Error(err))
			case raErr != nil:
				logger.Warn("judge: improve_tool insert rows-affected unavailable — recording as not created",
					zap.Error(raErr))
			case n > 0:
				itemCreated = true
			default:
				itemDeduped = true
				logger.Info("judge: improve_tool NOT created — an item for this key is already open (previous cycle unfinished)",
					zap.String("function", function))
			}
		} else {
			// No content_components row under this FUNCTION. For a PORTED
			// instance that is not a data problem — it is what a ported tool IS:
			// its subject key is the page name stem and its component is the
			// shared ported-page wrapper, so the fork lookup above can never hit
			// (bugs_open/146; bugs_closed/281 Finding B). This arm used to stop
			// at the log below, and the note's "route this manually" routed to
			// nobody: pasteboard and vibe-equalizer each FAILED Tier 4 twice
			// (2026-08-05 and 08-14) and filed nothing. The run item's own spec
			// names the instance (check_tool_acceptance_due writes component_id
			// and page_id), so when that component resolves to a NON-tool level
			// the verdict is filed as ported_tool_fix — the third producer of
			// the vocabulary check_tool_health and check_tool_acceptance already
			// emit, handler-less at needs_human_review for the same reason they
			// are: tool-improver's writeback targets the wrapper SHARED by every
			// ported page on the site. Anything else — no spec component, a fork
			// whose function moved, a lookup error — keeps this branch's old
			// behaviour byte-for-byte.
			portedRouted = routePortedAcceptanceFailure(ctx, params, logger, v, function, siteID, criteria, issue)
			if !portedRouted {
				logger.Info("judge: no content_components row for function — improve_tool item not created (recreated/adopted tool; route manually)",
					zap.String("function", function))
			}
		}
	}

	// The note is written LAST and from the outcome, not from the intent: it is
	// the loop's own durable record, and a note claiming a fix was queued when
	// the insert missed is the "trust the status, not the artefact" failure this
	// codebase keeps paying for. Both branches can miss — a recreated or adopted
	// tool has no content_components row at all.
	var fixLine string
	switch {
	case escalated && noAutoFix:
		// The no_auto_fix reason comes FIRST wherever both hold: "the fence
		// forbids a rewrite" is the fact that decides what a human may do next,
		// and the cycle count is only ever context beside it.
		fixLine = fmt.Sprintf(
			"NOT auto-fixed — this fence declares no_auto_fix (%s), so %s is escalated to human review (acceptance_stuck) and NO improve_tool item was created; an automated rewrite here could only pass by weakening the protected markup",
			reasonOrUnstated(noAutoFixReason), strings.Join(v.FailedIDs, ", "))
	case escalated:
		fixLine = fmt.Sprintf(
			"NOT auto-fixed — %d previous improve_tool cycle(s) failed to turn %s green, so this is escalated to human review (acceptance_stuck) instead of a %d%s identical attempt",
			attempts, strings.Join(v.FailedIDs, ", "), attempts+1, ordinalSuffix(attempts+1))
	case itemCreated:
		fixLine = "improve_tool item created carrying the criteria as acceptance_test"
	case itemDeduped:
		// Not a failure: the previous cycle's item is still open, so this verdict
		// is a repeat of work already queued. Said plainly because the alternative
		// — the "none could be created" default — reads as a defect and sent a
		// previous reader looking for a broken insert (bugs_open/010).
		fixLine = fmt.Sprintf(
			"no new improve_tool item — one for %s is ALREADY OPEN under this key (the previous fix cycle has not finished); this verdict re-confirms work already queued",
			strings.Join(v.FailedIDs, ", "))
	case portedRouted:
		// A ported instance's verdict now has a destination. When the fence ALSO
		// says no_auto_fix the destination is identical (human review) — the
		// ported reason is the one recorded, because it is the one that decides
		// what a human may edit: the instance's markup, never the shared wrapper.
		fixLine = fmt.Sprintf(
			"no automated fixer may act — this is a PORTED instance whose component is the shared ported-page wrapper (bugs_closed/281): filed ported_tool_fix at needs_human_review carrying %s for a human to route",
			strings.Join(v.FailedIDs, ", "))
	case noAutoFix:
		// Same precedence as above, and the same known gap: no component row (or
		// a failed insert) means no escalation item, but crucially also no
		// improve_tool item — nothing is aimed at the protected markup either way.
		fixLine = fmt.Sprintf(
			"NOT auto-fixed — this fence declares no_auto_fix (%s), and the escalation item could NOT be raised (no active content_components row for this function, or the insert failed); NO improve_tool item was created, route this manually",
			reasonOrUnstated(noAutoFixReason))
	case stuck:
		fixLine = fmt.Sprintf(
			"NOT auto-fixed — %d previous improve_tool cycle(s) failed to turn %s green, and the escalation item could NOT be raised (no active content_components row for this function, or the insert failed); route this manually",
			attempts, strings.Join(v.FailedIDs, ", "))
	default:
		fixLine = "none — no improve_tool item could be created (no active content_components row for this function, or the insert failed); route this manually"
	}
	// renderLine is here too, and it is not a copy-paste slip: a two-profile run
	// where desktop fails and mobile passes files a shot AND a render, so a
	// FAILED verdict can still carry a picture of the profile that was fine.
	// Dropping it would lose exactly the comparison a human wants.
	body := fmt.Sprintf(`## Tier-4 acceptance FAILED — %s
Observed: %d of %d evaluated checks failed in headless Chromium%s: %s%s%s%s
Root cause: not diagnosed at this tier (behavioural run; the fixer loads PLAN+NOTES first)
Fix: %s
Verified: n/a — failing run recorded
Categories: acceptance-fail`,
		function, len(v.Failed), len(v.Results), profilesPhrase(v.Profiles), issue, chromeLine,
		evidenceLine(v.Shots), renderLine(v.Renders), fixLine)
	if _, err := insertDocNote(ctx, params.DB, "tool", function, siteID, body,
		`["acceptance-fail"]`, "tool-acceptance", sourceAgent, "", "tool-acceptance-agent"); err != nil {
		logger.Warn("judge: acceptance-fail note insert failed", zap.Error(err))
	}

	logger.Info("judge: acceptance FAILED",
		zap.String("function", function),
		zap.Strings("failed", v.Failed),
		zap.Bool("improve_tool_created", itemCreated),
		zap.Bool("escalated", escalated),
		zap.Bool("ported_tool_fix_filed", portedRouted),
		zap.Bool("no_auto_fix", noAutoFix),
		zap.Int("fix_cycles_spent", attempts))
	out := map[string]interface{}{
		"all_passed": false, "passed": len(v.Passed), "failed": len(v.Failed),
		"failing_checks": v.FailedIDs, "failing_instances": v.Failed,
		"improve_tool_created": itemCreated,
		"escalated":            escalated,
		"fix_cycles_spent":     attempts,
	}
	if portedRouted {
		// Present ONLY when the ported route fired, mirroring no_auto_fix below:
		// an ordinary verdict's result map keeps exactly the shape it has always
		// had, so a workflow branching on key presence keeps working.
		out["ported_tool_fix_filed"] = true
	}
	if noAutoFix {
		// Present ONLY when the fence opted in, so an ordinary verdict's result
		// map is exactly the shape it has always been — a workflow branching on
		// key presence keeps working, and nothing downstream sees a new key it
		// was never written against.
		out["no_auto_fix"] = true
		if noAutoFixReason != "" {
			out["no_auto_fix_reason"] = noAutoFixReason
		}
	}
	return out, nil
}

// routePortedAcceptanceFailure files a failing Tier-4 verdict on a PORTED
// instance as a ported_tool_fix work item (bugs_open/146; bugs_closed/281
// Finding B). It fires only on positive evidence, in this order:
//
//  1. the run item's spec names its component (input_data.spec.component_id —
//     check_tool_acceptance_due writes it on every item it files); absent
//     means an older item or a bespoke dispatch, and nothing new happens;
//  2. that component exists, is active, and its component_level is NOT
//     'tool' — a real fork whose function moved, a deleted component, or any
//     lookup error all fall through to the caller's old behaviour.
//
// The item joins the contract its two sibling producers established
// (check_tool_health.go, check_tool_acceptance.go): item_type ported_tool_fix,
// status needs_human_review, NO handler — tool-improver's writeback targets
// content_components.html_template, which for a ported instance is the wrapper
// SHARED by every ported page on the site (clobbered fleet-wide 2026-08-05 and
// 08-14), and no automated fixer edits page_components.rendered_html from a
// finding today. The key's check segment is its own (tool_acceptance_tier4),
// so a Tier-2 static finding and a Tier-4 behavioural one hold separate open
// decisions, exactly as tool_health's do.
//
// The insert is the acceptance_stuck idiom from this file: the arbiter
// predicate matches idx_swi_dedup, and a re-verdict REFRESHES the standing
// item (spec MERGE keeps any human-added keys) rather than duplicating it.
func routePortedAcceptanceFailure(ctx context.Context, params ActionParams, logger *zap.Logger,
	v acceptanceVerdict, function, siteID, criteria, issue string) bool {

	specComponentID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.component_id")
	if specComponentID == "" {
		return false
	}
	var level string
	if err := params.DB.QueryRowContext(ctx, `
		SELECT COALESCE(component_level, '')
		FROM content_components
		WHERE id = $1::uuid AND is_active`, specComponentID).Scan(&level); err != nil {
		logger.Info("judge: ported-route component lookup did not resolve — keeping the manual-routing behaviour",
			zap.String("component_id", specComponentID), zap.Error(err))
		return false
	}
	if level == "tool" {
		// A fork: the caller's function-keyed lookup missing it is a different
		// defect (a renamed function), and filing a ported item for it would
		// mislabel a tool that HAS an automated fixer. Not this route's case.
		return false
	}

	spec := map[string]interface{}{
		"component_id":      specComponentID,
		"check":             "tool_acceptance_tier4",
		"subject_key":       function,
		"issue":             issue,
		"failing_checks":    v.FailedIDs,
		"failing_instances": v.Failed,
		"acceptance_test":   json.RawMessage(criteriaOrNull(criteria)),
	}
	if pageID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.page_id"); pageID != "" {
		spec["page_id"] = pageID
	}
	if pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.page_name"); pageName != "" {
		spec["page_name"] = pageName
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
	specJSON, _ := json.Marshal(spec)

	res, err := params.DB.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			priority, handler_agent, status, created_by, spec, item_key, batch_id
		) VALUES ($1::uuid, 'acceptance', 'build', 'ported_tool_fix',
		          'medium', $2, 60, '', 'needs_human_review',
		          'tool-acceptance-agent', $3::jsonb, $4, $5::uuid)
		ON CONFLICT (site_id, item_key)
			WHERE item_key IS NOT NULL AND status NOT IN (%s)
		DO UPDATE SET spec = site_work_items.spec || EXCLUDED.spec,
		              summary = EXCLUDED.summary, updated_at = now()`,
		sqlInList(workItemTerminalStatuses)),
		siteID,
		fmt.Sprintf("Tier-4 acceptance failed for ported tool %s: %s", function, first(v.Details)),
		string(specJSON),
		fmt.Sprintf("ported_tool_fix:tool_acceptance_tier4:%s:%s", function, siteID),
		uuid.NewString(),
	)
	if err != nil {
		logger.Warn("judge: ported_tool_fix insert failed", zap.Error(err))
		return false
	}
	if n, raErr := rowsAffected(res, err); raErr != nil || n == 0 {
		// DO UPDATE means a conflicting open item is refreshed, so zero rows is
		// unexpected — treat it as not-filed and let the note say routing is
		// manual rather than claim a filing that did not happen.
		logger.Warn("judge: ported_tool_fix wrote no row — recording as not filed",
			zap.Error(raErr))
		return false
	}
	logger.Info("judge: ported instance FAILED Tier-4 — ported_tool_fix filed for human review",
		zap.String("subject_key", function),
		zap.Strings("failing_checks", v.FailedIDs))
	return true
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

// rowsAffected reports how many rows a write actually touched, tolerating the
// nil Result that accompanies a failed Exec. It exists because every keyed
// insert here is ON CONFLICT DO NOTHING, which succeeds while inserting zero
// rows when idx_swi_dedup already holds an open item for the key — so err==nil
// says the statement ran, NOT that work was queued. Callers that report their
// outcome in a durable note must branch on this, not on the error alone.
func rowsAffected(res sql.Result, err error) (int64, error) {
	if err != nil || res == nil {
		return 0, err
	}
	return res.RowsAffected()
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

		res, err := params.DB.ExecContext(ctx, `
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
		// Count rows, not absent errors — this number is quoted verbatim in the
		// acceptance-fail note ("routed separately as N responsive_fix item(s)"),
		// and DO NOTHING inserts nothing without erroring (bugs_open/010).
		if n, raErr := rowsAffected(res, err); raErr != nil || n == 0 {
			logger.Info("judge: responsive_fix NOT created — an item for this chrome component is already open",
				zap.String("component", component), zap.String("item_key", itemKey))
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

// acceptanceFenceFlags is the JUDGE's view of the criteria document — only the
// fence-level keys the judge itself acts on. The runner owns the rest of the
// schema (profiles / checks / container, in
// internal/adapters/browserrunner/run_checks_action.go's criteriaDoc); mirroring
// those here would be a second copy to keep in step for no gain, and json
// ignores every key we do not name.
//
//	{"profiles":[...], "checks":[...], "no_auto_fix": true,
//	 "no_auto_fix_reason": "section B consent gate — owner-approved wording"}
type acceptanceFenceFlags struct {
	// NoAutoFix says a FAILING run of this fence must go to a human rather than
	// to tool-improver. Zero value false, so it is opt-in per fence: every
	// existing criteria document behaves exactly as it did before this key
	// existed (bugs_open/126 candidate 2).
	NoAutoFix bool `json:"no_auto_fix"`
	// NoAutoFixReason is free text shown to whoever picks the escalation up —
	// what the fence is protecting and why a rewriter must not touch it.
	NoAutoFixReason string `json:"no_auto_fix_reason"`
}

// parseNoAutoFix reads the fence-level opt-out out of the raw criteria string.
//
// FAIL-OPEN BY CONSTRUCTION: empty, absent, or unparseable criteria all return
// false — today's behaviour, unchanged. A criteria document that does not parse
// is a fence that never said anything about auto-fixing, and inventing a
// protection out of a JSON error would silently stop the fix loop across every
// tool whose PLAN happens to hold malformed criteria. The error is deliberately
// dropped rather than logged here: the runner already refuses a criteria
// document it cannot parse ("run_checks: criteria_json does not parse"), so a
// verdict reaching this judge with garbage criteria has a louder complaint
// upstream than anything this helper could add.
func parseNoAutoFix(criteria string) (bool, string) {
	if strings.TrimSpace(criteria) == "" {
		return false, ""
	}
	var flags acceptanceFenceFlags
	if err := json.Unmarshal([]byte(criteria), &flags); err != nil {
		return false, ""
	}
	return flags.NoAutoFix, strings.TrimSpace(flags.NoAutoFixReason)
}

// reasonOrUnstated keeps the escalation prose readable when an author set the
// flag but wrote no reason: "no_auto_fix ()" reads as a bug in the message.
func reasonOrUnstated(reason string) string {
	if strings.TrimSpace(reason) == "" {
		return "no reason stated on the fence"
	}
	return reason
}
