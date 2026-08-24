// FILE: platform/orchestration/actions/create_work_item_action.go
//
// Inserts a single work item into site_work_items. Used by handler agents
// to chain to the next pipeline stage after completing their work.
//
// Reuses the existing insertWorkItem() private helper for dedup behavior.
//
// Workflow config example (domain-research-classifier creating next item):
//
//   "create_next_item": {
//       "action": "create_work_item",
//       "config": {
//           "site_id":       "input_data.site_id",
//           "item_type":     "needs_briefing",
//           "handler_agent": "briefing-agent",
//           "item_pipeline":   "build",
//           "severity":      "high",
//           "source":        "domain-research-classifier",
//           "summary":       "Briefing needed",
//           "item_key_prefix": "briefing",
//           "spec_data":     "classification_result",
//           "parent_item_id": "input_data.work_item_id",
//           "priority":      10
//       },
//       "output_field": "next_item_created",
//       "next_step": "complete"
//   }
//
// Three optional config keys exist for callers whose item is an ACTION REQUEST
// rather than a detected defect (bugs_open/024 — a tool re-render request that
// was born dead three ways at once):
//
//   "spec_literal":          {"reason": "section_data_resolved"}
//       Constants written into the spec verbatim. spec_data is a PATH into
//       collected data and can never express a literal, so without this no
//       workflow step can stamp a value that a downstream gate reads.
//
//   "spec_paths":            {"component_id": "update_result.component_id"}
//       Per-key paths resolved individually. A configured path that does not
//       resolve is a HARD ERROR — an incomplete spec silently degrades the
//       downstream re-render to assemble-from-stale and still reports success.
//
//   "item_key_suffix_field": "update_result.component_id"
//       Appends a resolved value to the item_key. The default key is
//       '<prefix>_<domain>', which is SITE-wide: without a suffix, two
//       components fixed close together collide on idx_swi_dedup and one
//       request is lost.
//
//   "recurrence_expected":   true
//       Sets workItem.recurrenceExpected, which skips insertWorkItem's
//       anti-churn heuristics (dedup is NOT waived). For an action request a
//       'complete' predecessor is a SUCCESS, not a strike.
//
// Layering: spec_data, then spec_paths, then spec_literal — later wins.
//
// Registration (add to registry.go):
//   "create_work_item": { Handler: CreateWorkItemAction, IsLocal: true },

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var CreateWorkItemInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"spec_data", "parent_item_id", "page_id", "component_id", "summary"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
	// The `domain` -> `pipeline` rename, declared rather than hand-rolled
	// (bugs_open/136). This action carried the rename in three lines of
	// back-compat Go for months; the declaration does the same job and adds the
	// two things the Go could not: the offline audit can now see that nine live
	// steps still spell it the old way, and the next author renaming a setting
	// has a mechanism to reach for instead of a precedent to copy.
	//
	DeprecatedConfigKeys: map[string]string{"item_domain": "item_pipeline"},

	// The literal-setting contract: every key this action reads straight from
	// params.StepConfig.Config rather than through ExtractActionInputs. The
	// data-input keys are above (Required/Optional); the framework's own keys
	// (input_fields, error_step, timeout_seconds, …) are recognised centrally
	// by datahelpers.IsFrameworkStepConfigKey and must NOT be repeated here —
	// listing one would misstate whose contract it belongs to.
	//
	// Taken from the action body, key name by key name, NOT by grepping for
	// `config["`: `priority` arrives through datahelpers.GetIntField and
	// `item_pipeline` through ResolveConfigSetting, so an access-pattern grep
	// misses exactly the two least obvious entries. That grep is the recorded
	// mistake in WRONG_CALLS.md 2026-08-08 — the §3 read-list it produced was
	// wrong about `priority`, which is read.
	ConfigKeys: []string{
		"item_type",             // :124
		"handler_agent",         // :128
		"item_pipeline",         // :135, via ResolveConfigSetting
		"severity",              // :137
		"source",                // :141
		"status",                // :152
		"priority",              // :156, via GetIntField
		"item_key_prefix",       // :159
		"item_key_suffix_field", // :184
		"recurrence_expected",   // :198
		"spec_paths",            // :219
		"spec_literal",          // :236
	},

	// Opted in (bugs_open/136 item D). The three keys that blocked this are
	// adjudicated and gone from live config: `summary_template` (migration 343,
	// owner decision — a static literal, template rendering declined),
	// `spec_fields` and `domain` (migration 350 — both dead, established by
	// reading every strategy in ExtractActionInputs rather than by grep: each
	// one iterates Required ∪ Optional, or config["input_fields"], or
	// spec.Deprecated, so a key in none of the three is resolved by nothing).
	CheckConfig: true,

	// `spec` is RETIRED, not unknown (bugs_open/234). Three live steps carried
	// it for months; this action has never read a key by that name — it builds
	// the item spec from spec_data / spec_paths / spec_literal (below) — so
	// every item they filed carried spec = '{}', and improvement-loop's
	// `refresh_site_components` flag never reached the rerender gate (16/16
	// rows empty, measured 2026-08-09). Migration 364 translated all three
	// carriers onto the real spellings; this declaration is what stops a
	// fourth appearing: a step carrying `spec` now fails validation with the
	// replacement named, instead of filing empty specs that read as success.
	RemovedConfigKeys: map[string]string{
		"spec": "never read by any version of this action; use spec_data (a path to a map), " +
			"spec_paths ({key: path}), or spec_literal ({key: constant}) — bugs_open/234",
	},

	// Strict as of bugs_open/234 (owner decision 2026-08-10). The ConfigKeys
	// doc comment's own precondition is met: the recognised set was checked
	// against EVERY live step — a recursive all-depths walk over every live
	// definition, not the top-level-only census that undercounted 356's
	// carriers — and after migration 364 the unknown-key count is zero. From
	// here an unrecognised key on this action is a definition error caught at
	// seed time, not a silently-empty spec found by archaeology months later:
	// `spec`, `spec_fields`, `domain` and `summary_template` were all exactly
	// that, and warn-only surfaced none of them.
	StrictConfig: true,
}

func init() {
	datahelpers.RegisterActionInputSpec("create_work_item", CreateWorkItemInputSpec)
}

func CreateWorkItemAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "create_work_item"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		CreateWorkItemInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	// Config literals
	config := params.StepConfig.Config
	itemType, _ := config["item_type"].(string)
	if itemType == "" {
		return nil, fmt.Errorf("item_type config is required")
	}
	// An empty handler_agent is required to be an ERROR only when the item is
	// born dispatchable (an omitted status defaults to 'triaged' below) — a
	// handler-less dispatchable row could only ever become claim's `blocked`,
	// and since migration 443 the INSERT itself would be refused by CHECK
	// swi_no_handlerless_promotable. A PARKED item may omit the handler on
	// purpose: `status: "needs_human_review"` + no handler is the platform's
	// HITL idiom (migration 217; bugs_open/291 — tool-auditor's review items
	// bled to `blocked` for months because this validation forced every config
	// to name SOME handler, and the name it chose had never existed).
	handlerAgent, _ := config["handler_agent"].(string)
	status, _ := config["status"].(string)
	if handlerAgent == "" && (status == "" || workItemStatusRequiresRegisteredHandler(status)) {
		return nil, fmt.Errorf("handler_agent config is required when the item is born dispatchable (status %q; an omitted status defaults to \"triaged\") — to park an item for a human, set status \"needs_human_review\" and omit the handler (migration 217 idiom; bugs_open/291)", status)
	}
	// Same truth table as the hand-rolled shim this replaces — new name wins,
	// old name works, "build" when neither is set — now driven by the spec's
	// DeprecatedConfigKeys so the audit and the next author can both see it.
	itemPipeline, _ := datahelpers.ResolveConfigSetting(
		config, CreateWorkItemInputSpec, "item_pipeline", "build", logger)
	severity, _ := config["severity"].(string)
	if severity == "" {
		severity = "high"
	}
	source, _ := config["source"].(string)
	if source == "" {
		source = params.AgentType
	}
	summary := inputs.Get("summary")
	if summary == "" {
		summary, _ = config["summary"].(string)
	}
	if summary == "" {
		summary = itemType
	}
	// status was read beside handler_agent above; the default is the dispatch
	// queue's entry status.
	if status == "" {
		status = "triaged"
	}
	priority := datahelpers.GetIntField(config, "priority", 100)

	// Build item_key from prefix + domain
	itemKeyPrefix, _ := config["item_key_prefix"].(string)
	var itemKey string
	if itemKeyPrefix != "" {
		// Try to get domain for the key
		domain := inputs.Get("domain")
		if domain == "" {
			domain = siteID.String()[:8]
		}
		itemKey = fmt.Sprintf("%s_%s", itemKeyPrefix, domain)

		// Optional scoping suffix. '<prefix>_<domain>' is SITE-wide: two
		// components fixed close together on one site collide on
		// idx_swi_dedup (site_id, item_key) and one request is simply lost.
		// A step that knows what it is acting on names it here.
		//
		// Unset, the key is exactly as it was, so no existing caller changes.
		// CONFIGURED BUT UNRESOLVED IS A HARD ERROR, matching spec_paths below.
		// The earlier version logged a Warn and fell back to the site-wide key,
		// on the reasoning that the fallback is non-regressive. Two council
		// seats (editquality r5, bug_historian r6) independently called that
		// inconsistent with spec_paths in this same function, and they are
		// right: the site-wide key IS the collision defect, so "fall back to
		// it" means "silently reinstate the bug this field exists to fix" —
		// the log-only shape that is the root cause of every prior occurrence
		// in this family.
		if f, ok := config["item_key_suffix_field"].(string); ok && f != "" {
			sfx := datahelpers.ExtractNestedFieldString(params.CollectedData, f)
			if sfx == "" {
				return nil, fmt.Errorf(
					"item_key_suffix_field: path %q did not resolve; refusing to fall back to the site-wide key %q, which is the collision this field prevents",
					f, itemKey)
			}
			itemKey = fmt.Sprintf("%s_%s", itemKey, sfx)
		}
	}

	// recurrence_expected marks an item whose RE-REQUEST is normal rather than
	// evidence that previous handlers failed — see workItem.recurrenceExpected.
	// Default false, so existing callers keep today's behaviour exactly.
	recurrenceExpected, _ := config["recurrence_expected"].(bool)

	// Build the spec JSONB in three layers, later winning over earlier:
	//   spec_data    — a PATH to a map already in collected data (as before)
	//   spec_paths   — {key: path}, each resolved individually
	//   spec_literal — {key: value}, verbatim
	// spec_data alone could never express a constant, which is why no workflow
	// step could stamp e.g. reason='section_data_resolved' — the value
	// create_rerender_items gates the section re-render on (bugs_open/024).
	specMap := map[string]interface{}{}
	if specData := inputs.GetMap("spec_data"); specData != nil {
		for k, v := range specData {
			specMap[k] = v
		}
	}

	// spec_paths: a configured path that does not resolve is a HARD ERROR.
	// The step author asked for this field; silently omitting it is exactly how
	// bugs_open/024 stayed invisible for three cycles — a spec missing
	// component_id makes create_rerender_items' scoped=false, which degrades
	// the re-render to assemble-from-stale and still reports success.
	if sp, ok := config["spec_paths"].(map[string]interface{}); ok {
		for key, rawPath := range sp {
			path, isStr := rawPath.(string)
			if !isStr || path == "" {
				return nil, fmt.Errorf("spec_paths[%q] must be a non-empty string path", key)
			}
			val := datahelpers.ExtractNestedField(params.CollectedData, path)
			if val == nil || val == "" {
				return nil, fmt.Errorf(
					"spec_paths[%q]: path %q did not resolve; refusing to create a work item with an incomplete spec",
					key, path)
			}
			specMap[key] = val
		}
	}

	// spec_literal: constants, no resolution attempted.
	if sl, ok := config["spec_literal"].(map[string]interface{}); ok {
		for k, v := range sl {
			specMap[k] = v
		}
	}

	specJSON := "{}"
	if len(specMap) > 0 {
		b, err := json.Marshal(specMap)
		if err != nil {
			return nil, fmt.Errorf("marshal spec: %w", err)
		}
		specJSON = string(b)
	}

	// Optional parent_item_id
	var dependsOn []uuid.UUID
	if parentIDStr := inputs.Get("parent_item_id"); parentIDStr != "" {
		if parentID, err := uuid.Parse(parentIDStr); err == nil {
			dependsOn = append(dependsOn, parentID)
		}
	}

	// Optional page_id
	var pageID *uuid.UUID
	if pageIDStr := inputs.Get("page_id"); pageIDStr != "" {
		if parsed, err := uuid.Parse(pageIDStr); err == nil {
			pageID = &parsed
		}
	}

	// Optional component_id
	var componentID *uuid.UUID
	if componentIDStr := inputs.Get("component_id"); componentIDStr != "" {
		if parsed, err := uuid.Parse(componentIDStr); err == nil {
			componentID = &parsed
		}
	}

	// Insert via transaction (insertWorkItem expects *sql.Tx)
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	w, err := writeWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       source,
		pipeline:     itemPipeline,
		itemType:     itemType,
		severity:     severity,
		summary:      summary,
		spec:         specJSON,
		pageID:       pageID,
		componentID:  componentID,
		priority:     priority,
		handlerAgent: handlerAgent,
		status:       status,
		createdBy:    source,
		itemKey:      itemKey,
		dependsOn:    dependsOn,

		recurrenceExpected: recurrenceExpected,
	}, dropOnConflict, logger)
	if err != nil {
		return nil, fmt.Errorf("insert work item: %w", err)
	}
	inserted := w.Inserted

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("CreateWorkItemAction: complete",
		zap.String("item_type", itemType),
		zap.String("handler_agent", handlerAgent),
		zap.Bool("inserted", inserted),
		zap.Bool("born_blocked", w.BornBlocked),
		zap.Bool("owned_page_parked", w.OwnedPageParked),
		zap.String("item_key", itemKey),
	)

	// A dedup INSIDE a loop iteration is the runtime shadow of bugs_open/321:
	// with a key coarser than the loop item, iterations 2..N of a batch vanish
	// exactly here, and the loop reports success either way. The offline
	// detector (config-key-audit --loop-sitewide-item-keys) catches the
	// missing-suffix shape at definition level, but it cannot see a suffix that
	// RESOLVES yet is loop-invariant — this Warn is the net under that blind
	// spot. It also fires on a legitimate cross-run dedup (an earlier run's item
	// still open), so it is observability, not an error: the reader separates
	// the two by whether the deduped keys differ WITHIN one orchestration.
	if fire, loopVar, suffixConfigured := shouldWarnLoopDedup(config, inserted); fire {
		logger.Warn("CreateWorkItemAction: insert deduped away inside a loop iteration — "+
			"if other iterations of this batch share the key, their findings are being dropped (bugs_open/321)",
			zap.String("item_key", itemKey),
			zap.String("item_type", itemType),
			zap.String("loop_var", loopVar),
			zap.Bool("suffix_configured", suffixConfigured),
		)
	}

	return map[string]interface{}{
		"inserted":      inserted,
		"item_type":     itemType,
		"handler_agent": handlerAgent,
		"site_id":       siteID.String(),
		"item_key":      itemKey,
		"deduped":       !inserted,
		// born_blocked: the row exists but was demoted to 'blocked' at the door
		// because handler_agent names no registered agent (bugs_open/291).
		// Additive key — nothing consumed item_created.* before it existed.
		"born_blocked": w.BornBlocked,
		// owned_page_parked: the row exists but was parked at 'deferred' with no
		// handler, because its page is rebuild_policy='owned' and the configured
		// handler declares it refuses owned pages (bugs_open/333).
		//
		// ⚠ `handler_agent` above still reports what the CONFIG asked for, not
		// what the row carries — that is the contract every existing consumer
		// reads, and changing it would rewrite history for the 291 case too.
		// `row_status` is the row's ACTUAL status, so a producer counting its own
		// successes can tell "filed and dispatchable" from "filed and parked"
		// without re-querying. Both additive.
		"owned_page_parked": w.OwnedPageParked,

		// deferred / retry_after / prior_attempts: the row EXISTS and WILL
		// dispatch, but not yet — the anti-churn brake's within-cycle arm held it
		// back (bugs_open/326, option D). Additive, on the born_blocked precedent:
		// no live definition branched on this action's output when these were
		// added (0 conditional steps reference deduped/inserted, 2026-08-24).
		// Before them, `deduped: true` had two meanings — "an open item covers
		// this" and "the brake ate your request" — and the second was the bug.
		"deferred":       w.Deferred,
		"retry_after":    deferredUntilString(w),
		"prior_attempts": w.PriorAttempts,
		"row_status":     effectiveRowStatus(status, w),
	}, nil
}

// effectiveRowStatus reports the status the row actually carries after the write
// door has had its say, rather than the status the caller asked for.
//
// It exists because bugs_open/177's lesson is that a producer whose completion
// metric counts a FILING as a success will report 100% while filing items that
// are unsatisfiable at birth. Both demotions are exactly that shape, and a
// producer cannot see either one from `inserted` alone.
func effectiveRowStatus(requested string, w workItemWrite) string {
	switch {
	case w.BornBlocked:
		return "blocked"
	case w.OwnedPageParked:
		return ownedPageParkedStatus
	default:
		return requested
	}
}

// shouldWarnLoopDedup decides whether a non-inserted (deduped) work item
// deserves the bugs_open/321 loop-dedup Warn: only when the step is executing
// as a loop iteration. Loop context is read from the framework-injected
// loop_iteration / loop_var_name keys (loop_expansion_handler.go:155; both in
// frameworkStepConfigKeys, which is why neither appears in ConfigKeys above —
// an author never writes them, so declaring them would invite exactly that).
//
// Extracted so the gate has a direct test (the shouldStripLiteralMarkdown
// precedent): an inserted item, or a top-level dedup, must never fire — a Warn
// that fires on every dedup fleet-wide would be noise nobody reads, which is
// how an observability line dies.
func shouldWarnLoopDedup(config map[string]interface{}, inserted bool) (fire bool, loopVar string, suffixConfigured bool) {
	if inserted {
		return false, "", false
	}
	if _, inLoop := config["loop_iteration"]; !inLoop {
		return false, "", false
	}
	loopVar, _ = config["loop_var_name"].(string)
	suffixField, _ := config["item_key_suffix_field"].(string)
	return true, loopVar, suffixField != ""
}

// deferredUntilString renders the deferral boundary for the result map, and
// empty when there is none. A zero time.Time formats as a real-looking RFC3339
// timestamp in the year 1, which is exactly the kind of value a reader trusts.
func deferredUntilString(w workItemWrite) string {
	if !w.Deferred || w.DeferredUntil.IsZero() {
		return ""
	}
	return w.DeferredUntil.UTC().Format(time.RFC3339)
}
