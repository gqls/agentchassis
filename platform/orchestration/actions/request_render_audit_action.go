// FILE: platform/orchestration/actions/request_render_audit_action.go
//
// request_render_audit — asks the browser-runner adapter to RENDER a site's
// deployed pages and measure what a visitor actually sees: text contrast against
// the effective background, images that failed to load, horizontal overflow.
//
// Sibling of `request_browser_run` in shape — same envelope, same
// `AwaitResponse=true` so the engine registers the awaited request and resumes
// when the adapter replies — but deliberately on a DIFFERENT topic. See
// `renderAuditTopic` below for why a big audit gets its own pod.
//
// WHY THIS EXISTS AS AN ACTION AT ALL. The measurement lives in the
// browser-runner adapter because that is the only pod with Chromium; the chassis
// has none. Until this action existed the measurement was reachable only by a
// human typing `scripts/render_audit.py`, which is exactly how an unreadable
// chart reached a live site and was found by the owner in a screenshot rather
// than by the platform. `check_palette_contrast` runs on every build and cannot
// see that class by construction — it reads the composed palette, and the defect
// was a component hard-coding an ink over a themed fill.
//
// WHAT IT SENDS. Every ACTIVE, DEPLOYED page of the site, not a sample. A
// contrast defect is usually per-component, so auditing one page and declaring
// the site clean is the false-green shape this whole line of work exists to
// stop. `max_pages` caps it, and when the cap bites the action says so in its
// result rather than silently truncating — a truncated sweep that reports
// "clean" is worse than no sweep.
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// renderAuditTopic is the DEDICATED lane for whole-site audits.
//
// Deliberately not browserRunnerTopic. A full-site audit is tens of sequential
// Chromium navigations holding a pod for minutes; on the shared browser-runner
// it starves tool acceptance, buries that pod's logs, and takes acceptance down
// with it when a browser wedges. Owner ruling 2026-07-28: a big audit gets its
// own pod, its own logs and its own failure state. The pod is a second
// Deployment of the SAME browser-runner image with REQUESTS_TOPIC overridden
// (deployments/kustomize/services/render-audit-adapter), so this costs a topic,
// not a binary.
//
// `bugs_open/096` is the same lesson one layer up: a long job on a shared lane
// head-of-line blocks everything until it gets a lane of its own.
const renderAuditTopic = "system.adapter.render-audit.requests"

// RequestRenderAuditInputSpec documents the step config.
var RequestRenderAuditInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{"site_id_field", "domain_field", "max_pages", "page_names", "topic", "capture_renders", "rotate_coverage"},
	Defaults: map[string]interface{}{
		"site_id_field":   "site_record.site_id",
		"domain_field":    "site_record.domain",
		"max_pages":       25,
		"topic":           renderAuditTopic,
		"capture_renders": false,
		// rotate_coverage OPTS IN to the coverage cursor (bugs_open/394).
		// Default FALSE, and the default is the UNSAFE-side-off shape the owner
		// ruled for new authority on a shared seam (2026-08-02 §2): a caller that
		// has not asked for rotation keeps today's deterministic prefix exactly.
		// Two live carriers as of 2026-08-26 — render-audit-agent (cap 60, opted
		// in by migration 646) and design-critique-agent (cap 8, deliberately NOT
		// opted in: it is a manual sampler with no cadence, and its 8 pages are
		// plausibly meant as the most important 8 rather than any 8).
		"rotate_coverage": false,
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("request_render_audit", RequestRenderAuditInputSpec)
}

// RequestRenderAuditAction resolves the site's deployed URLs and dispatches one
// render_audit request covering all of them.
func RequestRenderAuditAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "request_render_audit"))
	config := params.StepConfig.Config

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "site_id_field", "site_record.site_id"))
	domain := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "domain_field", "site_record.domain"))
	if siteID == "" || domain == "" {
		return nil, fmt.Errorf("request_render_audit: need both site_id and domain (got %q / %q)", siteID, domain)
	}
	if params.DB == nil {
		return nil, fmt.Errorf("request_render_audit: no database handle")
	}

	maxPages := datahelpers.GetIntField(config, "max_pages", 25)
	if maxPages <= 0 {
		maxPages = 25
	}
	captureRenders, _ := config["capture_renders"].(bool)
	rotateCoverage, _ := config["rotate_coverage"].(bool)

	// SHIPPED pages only. An unshipped page has no live URL to render, and
	// asking for one produces a navigation failure that reads like a defect.
	//
	// bugs_open/185 tranche 2 (2026-08-03): this read `build_status = 'deployed'`,
	// which is not "has a live URL" — a needs_rebuild page that once deployed is
	// still serving its previous artefact, so it is exactly what a RENDER audit
	// should photograph. Measured before converging: 36 pages across 8 sites were
	// live and invisible to this audit. The intent in the comment above was always
	// "has a live URL"; the predicate now says what the comment meant.
	//
	// The ordering columns are SELECTed as well as ordered by, because the
	// coverage cursor (bugs_open/394) is a keyset position in this exact
	// ordering. Reading them here is what makes it impossible for the cursor and
	// the ORDER BY to disagree — the alternative, re-deriving the tuple
	// elsewhere, is two spellings of one ordering and the drift is invisible
	// until a window is silently skipped. nav_order is COALESCED in the SELECT
	// too, so the stored tuple is the sorted value and never the raw one.
	rows, err := params.DB.QueryContext(ctx, `
		SELECT COALESCE(url, ''), COALESCE(nav_order, 999), name
		FROM pages
		WHERE site_id = $1::uuid
		  AND `+datahelpers.PageWantedLivePredicateFor("")+`
		  AND `+datahelpers.PageHasShippedPredicateFor("")+`
		  AND COALESCE(url, '') <> ''
		ORDER BY COALESCE(nav_order, 999), name
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("request_render_audit: page lookup failed: %w", err)
	}
	defer rows.Close()

	// Every live page is materialised, not just the first max_pages. The old loop
	// read them all anyway (it kept counting so the truncation stayed reportable);
	// this keeps the rows so a window can be cut from anywhere in the ordering
	// rather than only from the front.
	var all []auditPageRow
	for rows.Next() {
		var r auditPageRow
		if err := rows.Scan(&r.Path, &r.Ord, &r.Name); err != nil {
			return nil, fmt.Errorf("request_render_audit: scan: %w", err)
		}
		r.URL = absoluteAuditURL(domain, r.Path)
		all = append(all, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("request_render_audit: rows: %w", err)
	}
	total := len(all)

	// ── window selection ────────────────────────────────────────────────────
	//
	// PREFIX MODE (rotate_coverage off, or the site fits inside the cap) is
	// byte-identical to the behaviour this action has always had. That is
	// deliberate and it is what makes the pre-existing truncation tests a
	// COMPATIBILITY GUARD rather than tests that had to be rewritten: if they
	// still pass unchanged, the default path did not move.
	var (
		window       []auditPageRow
		nextCursor   *auditCursor
		pri          priorityResult
		rotating     bool
		rotationSize int
		agentType    string
	)
	// THE CURSOR KEY MUST BE THE RUNNING AGENT, NOT THE DISPATCHER.
	//
	// This read `params.ExecutionContext.Sender.AgentType` until 2026-08-26, and
	// the first LIVE run caught it: a hand dispatch over
	// `system.agent.generic.requests` stored the cursor under agent_type
	// "generic" while the SAME run's durable truncation row recorded
	// "render-audit-agent" — two identities for one run. `Sender` is whoever put
	// the message on the topic; the scheduled rotation uses a third topic again
	// (`system.agent.scheduled.requests`). Keyed on that, one logical caller
	// keeps a SEPARATE cursor per dispatch path, so a hand run's coverage is
	// invisible to the scheduled run and each restarts from the top.
	//
	// runningStepProvenance is the canonical resolver
	// (ExecutionContext.ResolvedAgentType() with a params.AgentType fallback) and
	// is what LogActionFindings already uses to stamp the truncation row. Sharing
	// it means the cursor key and the durable row cannot disagree — the estate's
	// "copy the predicate, never retype it" rule applied to an identity.
	//
	// ⚠ NO UNIT TEST COULD HAVE CAUGHT THIS: the fixture sets Sender.AgentType to
	// "render-audit-agent" by hand, so both readings agreed. It took the artefact.
	agentType, _ = runningStepProvenance(params)

	switch {
	case !rotateCoverage || total <= maxPages:
		// Nothing to rotate: either the caller did not opt in, or the whole site
		// fits in one run. In the second case the cursor is never read or
		// written, so a stale row left by a site that has since shrunk is inert.
	case agentType == "":
		// The cursor is keyed on the dispatching agent, so an unnamed sender has
		// nowhere to store one. Degrade to the prefix rather than guessing a key
		// — a cursor written under the wrong identity would be silently shared
		// between callers with different caps.
		logger.Warn("request_render_audit: rotate_coverage is set but the sender agent type is empty — falling back to the prefix window",
			zap.String("domain", domain))
	default:
		rotating = true
		cur, err := loadAuditCursor(ctx, params.DB, siteID, agentType)
		if err != nil {
			// Fail OPEN and loudly. A cursor we cannot read must cost coverage
			// progress, never the audit itself.
			logger.Warn("request_render_audit: coverage cursor unreadable — this run takes the prefix window and does not advance",
				zap.String("domain", domain), zap.String("agent_type", agentType), zap.Error(err))
			rotating = false
			break
		}
		// The priority set is assembled FIRST because it takes its slots off the
		// top: the rotation gets what is left, never the other way round.
		pri = selectPriorityRegradeSet(ctx, params.DB, siteID, all, cur, maxPages/2, logger)
		skip := make(map[string]bool, len(pri.taken))
		for _, r := range pri.taken {
			skip[r.Path] = true
		}
		rotationSize = maxPages - len(pri.taken)
		window, nextCursor = selectAuditWindow(all, cur, rotationSize, skip)
	}

	var urls []string
	if rotating {
		// Priority pages go FIRST. If the adapter abandons a run part-way, the
		// pages whose grading latency is the thing being protected are the ones
		// most likely to have been measured.
		for _, r := range pri.taken {
			urls = append(urls, r.URL)
		}
		for _, r := range window {
			urls = append(urls, r.URL)
		}
	} else {
		for i := 0; i < len(all) && i < maxPages; i++ {
			urls = append(urls, all[i].URL)
		}
	}
	if len(urls) == 0 {
		// No await, and NOT a failure: a site with nothing deployed has nothing
		// to measure. Declaring it clean would be the lie.
		logger.Info("request_render_audit: no deployed pages — skipping (nothing to render)",
			zap.String("domain", domain))
		return map[string]interface{}{
			"skipped": true,
			"reason":  "no_deployed_pages",
			"domain":  domain,
		}, nil
	}

	truncated := total > len(urls)
	if truncated {
		// Say it loudly: a capped sweep reporting "clean" is a false green.
		logger.Warn("request_render_audit: page list TRUNCATED by max_pages — a clean result covers only the audited pages",
			zap.String("domain", domain), zap.Int("audited", len(urls)), zap.Int("total", total),
			zap.Bool("rotating", rotating), zap.Int("priority_pages", len(pri.taken)))
		// And say it DURABLY, before the dispatch. This step awaits, and an
		// awaiting step's own result never survives the park
		// (persistAwaitingStateWithRetry loads fresh state and keeps only the
		// awaited-request entries — RFC_012 addendum 2, owner-ruled option B):
		// agent_error_log is the one sink that outlives the await, and the row
		// must land before the send so a failed dispatch cannot unrecord the
		// truncation (bugs_open/242). LogActionFindings is the named door for
		// exactly this class; it fills the join/provenance columns from params.
		attempted, recorded := LogActionFindings(ctx, params, siteID, domain,
			"request_render_audit", []agenterrors.Finding{{
				ErrorCode: "RENDER_AUDIT_TRUNCATED",
				Severity:  "warning",
				Message:   truncationMessage(rotating, len(urls), total, domain, pri, rotationSize),
				Context: truncationContext(rotating, len(urls), total, maxPages,
					auditedPaths(pri.taken, window), nextCursor, window, pri),
			}}, logger)
		if recorded < attempted {
			logger.Warn("request_render_audit: truncation row did not land — the pod log line above is the only record of this run's cap bite")
		}
	}

	// Overridable so a cluster that has not yet deployed the dedicated pod can
	// point this back at the shared browser-runner rather than publishing into a
	// topic nothing consumes. A producer aimed at an unconsumed topic piles
	// messages where nothing will run them, and it looks exactly like latency
	// (the 096 rollout order: consumer first, producer second).
	topic := datahelpers.GetStringField(config, "topic", renderAuditTopic)

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
			"action":                  "render_audit",
			"sender_agent_type":       params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":         params.ExecutionContext.OrchestrationID,
			"sender_pod_name":         params.ExecutionContext.Sender.PodName,
			"responses_topic":         myResponsesTopic,
			"parent_responses_topic":  myResponsesTopic,
			"reply_to_topic":          myResponsesTopic,
			"timestamp":               time.Now().UTC().Format(time.RFC3339),
		},
		"body": map[string]interface{}{
			"action":         "render_audit",
			"reply_to_topic": myResponsesTopic,
			"data": map[string]interface{}{
				"run_id":  params.ExecutionContext.OrchestrationID,
				"urls":    urls,
				"site_id": siteID,
				"domain":  domain,
				// The cap's bite travels IN THE REQUEST so the adapter can echo
				// it back in its summary — the reply envelope is the only part
				// of an awaited step that reaches the stored artefact (see the
				// truncation block above). Without these, `pages: 25` has no
				// total beside it and a capped sweep reads as a complete one.
				"pages_total": total,
				"truncated":   truncated,
				// Renders (desktop+mobile full-page screenshots) are opt-in per
				// step config; the adapter degrades to measurement-only when
				// object storage is absent. See RenderAuditRequest.CaptureRenders.
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
		return nil, fmt.Errorf("request_render_audit: marshal request: %w", err)
	}

	logger.Info("request_render_audit: sending to browser-runner adapter",
		zap.String("topic", topic),
		zap.String("request_id", newRequestID),
		zap.String("domain", domain),
		zap.Int("urls", len(urls)),
		zap.Bool("truncated", truncated))

	if err := params.Producer.ProduceWithValidation(ctx, topic, headers,
		[]byte(params.ExecutionContext.CorrelationID), messageBytes); err != nil {
		return nil, fmt.Errorf("request_render_audit: send to browser-runner adapter: %w", err)
	}

	// ── the cursor advances HERE, and the position is the ROTATION window's ──
	//
	// AFTER the produce, deliberately the OPPOSITE ordering from the truncation
	// row above. That row is a RECORD OF FACT and must land before the send so a
	// failed dispatch cannot unrecord it (bugs_open/242). The cursor is a
	// COMMITMENT ABOUT THE NEXT RUN: written before a produce that then fails, it
	// would skip a window nothing ever requested.
	//
	// nextCursor comes from selectAuditWindow, so it is the last page of the
	// ROTATION SLICE and never a priority page. That distinction is the whole
	// coverage guarantee: a priority page is carried out-of-band and may sit far
	// past the window, so letting one set the boundary would skip every page
	// between — in this run and in every run after it, because the boundary only
	// moves forward. TestPriorityPageBeyondTheWindowDoesNotMoveTheStoredCursor
	// pins it at this call site, where the union is actually assembled.
	//
	// A failed cursor write is NOT fatal — the audit is already in flight and
	// nothing about it is wrong — but it is LOUD, and it names the window that
	// will repeat, because a silently unrecorded advance looks exactly like a
	// working cursor while the same window is audited for ever.
	if rotating {
		var cerr error
		if nextCursor != nil {
			cerr = saveAuditCursor(ctx, params.DB, siteID, agentType, nextCursor)
		} else {
			// Cycle complete: removing the row is what makes the next capped run
			// start from the top. Leaving it would park the cursor past the end
			// for ever, which the past-the-end branch absorbs but only by
			// restarting — one wasted comparison per run, and a stored position
			// that no longer means anything.
			cerr = deleteAuditCursor(ctx, params.DB, siteID, agentType)
		}
		if cerr != nil {
			logger.Warn("request_render_audit: coverage cursor NOT advanced — this window will repeat on the next run",
				zap.String("domain", domain),
				zap.String("agent_type", agentType),
				zap.Bool("cycle_complete", nextCursor == nil),
				zap.String("window_first", firstWindowName(window)),
				zap.Error(cerr))
		}
	}

	return &RequestRepoAnalysisResult{ // same await-signal shape as the sibling requests
		Success:       true,
		RequestID:     newRequestID,
		TopicSentTo:   topic,
		AwaitResponse: true,
		Metadata: map[string]interface{}{
			"domain":          domain,
			"site_id":         siteID,
			"urls_audited":    len(urls),
			"pages_total":     total,
			"truncated":       truncated,
			"responses_topic": myResponsesTopic,
		},
	}, nil
}
