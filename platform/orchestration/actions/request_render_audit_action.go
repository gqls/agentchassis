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
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

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
	Optional: []string{"site_id_field", "domain_field", "max_pages", "page_names", "topic"},
	Defaults: map[string]interface{}{
		"site_id_field": "site_record.site_id",
		"domain_field":  "site_record.domain",
		"max_pages":     25,
		"topic":         renderAuditTopic,
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

	// Deployed pages only. An undeployed page has no live URL to render, and
	// asking for one produces a navigation failure that reads like a defect.
	rows, err := params.DB.QueryContext(ctx, `
		SELECT COALESCE(url, '')
		FROM pages
		WHERE site_id = $1::uuid
		  AND status = 'active'
		  AND build_status = 'deployed'
		  AND COALESCE(url, '') <> ''
		ORDER BY COALESCE(nav_order, 999), name
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("request_render_audit: page lookup failed: %w", err)
	}
	defer rows.Close()

	var urls []string
	total := 0
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, fmt.Errorf("request_render_audit: scan: %w", err)
		}
		total++
		if len(urls) >= maxPages {
			continue // keep counting so the truncation is reportable
		}
		if !strings.HasPrefix(u, "http") {
			u = "https://" + domain + "/" + strings.TrimPrefix(u, "/")
		}
		urls = append(urls, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("request_render_audit: rows: %w", err)
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
			zap.String("domain", domain), zap.Int("audited", len(urls)), zap.Int("total", total))
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
