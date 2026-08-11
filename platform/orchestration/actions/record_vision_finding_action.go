// FILE: platform/orchestration/actions/record_vision_finding_action.go
//
// RecordVisionFindingAction closes the void bugs_open/243 candidate 3 named:
// the tool-acceptance `look` step (TL-035 e) writes its critique to a
// `render-critique` doc_note that NO code reads — measured 2026-08-11, the
// day the vision half ran for the first time ever, reported a real defect
// (contrast 1.06:1 on dartsonline's setup-builder chips), and the run was
// stamped PASSED in the same second with nothing raised. The eyes were
// restored and wired to nothing.
//
// This action runs AFTER record_look. It reads the critique's machine line
// (the look prompt now ends `FINDINGS: none` / `FINDINGS: reported`) and, on
// a finding, files ONE deduped `vision_finding` work item routed to a human
// (handler `human-review`, status `needs_human_review`) — the acceptance_stuck
// shape, deliberately: same arbiter predicate, same spec-MERGE on conflict, so
// a nightly re-found defect refreshes one standing item instead of minting
// thirty.
//
// What it must NOT do, per TL-035's design line ("best-effort, never a
// verdict") and the Tier-4 guarantee (confirm, never refute — the RFC_002
// trigger): it cannot touch the acceptance verdict, cannot raise improve_tool
// (an auto-rewriter aimed by an eye's judgement call), and its own failure
// must not change the run's outcome (wire error_step to the same terminal
// step as next_step). A missing marker line files WITH verdict_line
// "unparsed" rather than staying silent — the failure mode of this mechanism
// must be a human seeing too much, never the void again.
//
// Registration:
//   "record_vision_finding": {
//       Handler:     RecordVisionFindingAction,
//       Category:    "tools",
//       Description: "File a vision critique that reports defects as one deduped vision_finding work item for human review",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RecordVisionFindingInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{},
	Optional:    []string{"critique_field", "function_field", "site_id_field", "images_field"},
	Defaults: map[string]interface{}{
		"critique_field": "vision_look.result",
		"function_field": "input_data.spec.function",
		"site_id_field":  "site_record.site_id",
		"images_field":   "browser_run",
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("record_vision_finding", RecordVisionFindingInputSpec)
}

// parseVisionVerdictLine reads the critique's LAST non-empty line for the
// machine marker the look prompt asks for. Three outcomes, and the zero-value
// direction is deliberate: an absent or mangled marker is "unparsed", which
// FILES (a human sees the ambiguity) — only an explicit "none" stays quiet.
func parseVisionVerdictLine(critique string) string {
	lines := strings.Split(critique, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		rest, ok := cutCaseInsensitivePrefix(line, "FINDINGS:")
		if !ok {
			return "unparsed"
		}
		if strings.Contains(strings.ToLower(rest), "none") {
			return "none"
		}
		return "reported"
	}
	return "unparsed"
}

func cutCaseInsensitivePrefix(s, prefix string) (string, bool) {
	if len(s) < len(prefix) || !strings.EqualFold(s[:len(prefix)], prefix) {
		return s, false
	}
	return s[len(prefix):], true
}

// visionFindingSummary compresses the critique's first sentence for the work
// item list; the full text travels in the spec.
func visionFindingSummary(function, critique string) string {
	gist := strings.TrimSpace(critique)
	if idx := strings.IndexAny(gist, ".\n"); idx > 0 {
		gist = gist[:idx]
	}
	if len(gist) > 160 {
		gist = gist[:157] + "…"
	}
	return fmt.Sprintf("Vision pass reports a visitor-visible defect on %s (checks all green): %s", function, gist)
}

func RecordVisionFindingAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "record_vision_finding"))
	config := params.StepConfig.Config

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	critiqueField := datahelpers.GetStringField(config, "critique_field", "vision_look.result")
	critique := datahelpers.ExtractNestedFieldString(params.CollectedData, critiqueField)
	if critique == "" {
		// look succeeded but its result is not text here — record_look would
		// already have failed on the same absence, so this is defensive. Filing
		// an item with no content would waste the reader it exists to reach.
		logger.Warn("record_vision_finding: no critique text found", zap.String("field", critiqueField))
		return map[string]interface{}{"filed": false, "verdict_line": "absent", "reason": "no critique text"}, nil
	}

	verdict := parseVisionVerdictLine(critique)
	if verdict == "none" {
		return map[string]interface{}{"filed": false, "verdict_line": verdict}, nil
	}

	function := resolveWithFallbacks(params.CollectedData,
		datahelpers.GetStringField(config, "function_field", "input_data.spec.function"),
		"input_data.function")
	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData,
		datahelpers.GetStringField(config, "site_id_field", "site_record.site_id"))
	if params.DB == nil || siteID == "" || function == "" {
		logger.Warn("record_vision_finding: cannot file — missing db/site/function",
			zap.Bool("db", params.DB != nil), zap.String("site_id", siteID), zap.String("function", function))
		return map[string]interface{}{"filed": false, "verdict_line": verdict, "reason": "missing db/site/function"}, nil
	}

	// The page the finding is about, judge's own lookup — best-effort.
	var pageID string
	_ = params.DB.QueryRowContext(ctx, `
		SELECT COALESCE(p.id::text, '')
		FROM content_components cc
		LEFT JOIN page_components pc ON pc.component_id = cc.id
		LEFT JOIN pages p ON p.id = pc.page_id AND p.site_id = $2::uuid
		WHERE cc.function = $1 AND cc.is_active
		LIMIT 1`, function, siteID).Scan(&pageID)

	spec := map[string]interface{}{
		"check":        "tool_acceptance_vision",
		"critique":     critique,
		"verdict_line": verdict,
	}
	if pageID != "" {
		spec["page_id"] = pageID
	}
	if params.ExecutionContext != nil && params.ExecutionContext.CorrelationID != "" {
		spec["run_correlation"] = params.ExecutionContext.CorrelationID
	}
	imagesField := datahelpers.GetStringField(config, "images_field", "browser_run")
	if refs, err := resolveVisionImageRefs(params.CollectedData, imagesField); err == nil && len(refs) > 0 {
		shots := make([]map[string]interface{}, 0, len(refs))
		for _, r := range refs {
			shots = append(shots, map[string]interface{}{
				"profile": r.Profile, "uri": r.URI, "page_url": r.PageURL,
			})
		}
		spec["screenshots"] = shots
	}
	specJSON, _ := json.Marshal(spec)

	// acceptance_stuck's arbiter, verbatim: the partial-index predicate must
	// match idx_swi_dedup (work_items_common.go lockstep), and the spec MERGES
	// so a human's triage keys survive the nightly re-file.
	_, err := params.DB.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			priority, handler_agent, status, created_by, spec, item_key, batch_id
		) VALUES ($1::uuid, 'acceptance', 'build', 'vision_finding',
		          'medium', $2, 40, 'human-review', 'needs_human_review',
		          'tool-acceptance-agent', $3::jsonb, $4, $5::uuid)
		ON CONFLICT (site_id, item_key)
			WHERE item_key IS NOT NULL AND status NOT IN (%s)
		DO UPDATE SET spec = site_work_items.spec || EXCLUDED.spec,
		              summary = EXCLUDED.summary, updated_at = now()`,
		sqlInList(workItemTerminalStatuses)),
		siteID,
		visionFindingSummary(function, critique),
		string(specJSON),
		fmt.Sprintf("vision_finding:%s:%s", function, siteID),
		uuid.NewString(),
	)
	if err != nil {
		// A filing failure must not change the run's outcome (TL-035: best-effort,
		// never a verdict) — but it must not be silent either.
		logger.Warn("record_vision_finding: vision_finding insert failed", zap.Error(err))
		return map[string]interface{}{"filed": false, "verdict_line": verdict, "reason": "insert failed: " + err.Error()}, nil
	}

	logger.Info("record_vision_finding: filed",
		zap.String("function", function), zap.String("site_id", siteID), zap.String("verdict_line", verdict))
	return map[string]interface{}{
		"filed":        true,
		"verdict_line": verdict,
		"item_key":     fmt.Sprintf("vision_finding:%s:%s", function, siteID),
	}, nil
}
