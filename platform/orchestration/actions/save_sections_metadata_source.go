// FILE: platform/orchestration/actions/save_sections_metadata_source.go
//
// Where SavePageSectionsAction finds its structured section data, and what it
// says when it cannot (bugs_open/194).
//
// THE DEFECT THIS CLOSES. SavePageSectionsAction persists two representations of
// a page section: `rendered_html`, which is what a visitor sees, and
// `content_data`, which is the ONLY thing the platform can re-render from
// (rerender_page_sections_action.go rebuilds each section from it with no writer
// pass). The structured half reaches the action through the config key
// `sections_metadata_field`, and a caller that does not set it falls through to
// the regex HTML-parse path, whose sections carry no content_data — so the INSERT
// writes SQL NULL, the page serves perfectly, and the save reports success.
// Measured 2026-08-04: four of the six live callers had never set the key. That is
// not four callers being careless; it is a saver that depends on being told where
// its own input lives, so forgetting is always available. `page-build-handler`'s
// copy of that key dates from 2026-02-18 and has been hand-copied three times
// since (seeds 034, 310, 312).
//
// WHAT A NULL COSTS, measured rather than asserted. The rerender path refuses to
// render a section with no stored content_data — rendering from nothing would
// blank a live page — and escalates the WHOLE page to a full LLM rebuild instead
// (rerender_page_sections_action.go:326). site_work_items carries 44 such
// escalations across 8 sites since 2026-07-12 ("a section had no stored
// content_data"), of which 13 FAILED outright on 2026-08-03. That is exposure for
// the class, not damage attributed to these callers — some of those pages predate
// content_data capture entirely — but it is what the state costs whoever wrote it.
//
// WHY THE DEFAULT LIVES HERE AND NOT IN SIX CONFIGS. save_sections_link_repair.go
// already argued this for its own defect, in this same action, and the argument
// transfers verbatim: "Persistence is the one chokepoint every body-section writer
// passes through … Putting the repair here makes an unrepaired section unsaveable
// whatever the workflow config says, which is the property the config-shaped
// defect above actually needs." bugs_closed/087 reached the same shape one layer
// up on 2026-08-04 — the writer plans its own sections so "no caller, present or
// future, can get this wrong". This is that rule applied to the saver's input.
//
// THE DEFAULT IS THE GATE'S OWN CONSTANT, not a second copy of the same string.
// defaultSectionsMetadataField is declared in validate_page_content_stats.go:75
// and is referenced here rather than redeclared, so the gate and the save cannot
// drift about where a page's structured content lives. Two hand-maintained copies
// of one path is the drift class this whole bug is an instance of.
//
// A SINGLE DEFAULT, DELIBERATELY NOT A PROBE. An earlier draft tried both known
// paths in order (page_content…, then rerender_sections…). Rejected: page-rerender
// names its field explicitly and always will, so probing its path from every save
// adds a second way to resolve for zero current benefit — and a save that finds a
// *different* run's metadata under a path nobody configured is a worse failure
// than the NULL it was meant to prevent. One default, consulted only when the
// caller has said nothing at all.
//
// THREE DECLARED STATES, mirroring the design rule validate_page_content_stats.go
// states for the same key ("the expectation is DECLARED per step rather than
// guessed from the payload shape"):
//
//	sections_metadata_field set          -> that path, origin "configured"
//	nothing set                          -> the default,  origin "default"
//	expects_no_sections_metadata: true   -> no path,      origin "declared_absent"
//	require_sections_metadata: true      -> absence is a REFUSAL, not a fallback
//
// `expects_no_sections_metadata` exists because one live caller legitimately has
// no structured content: tool-recreation-handler recreates a whole-page tool as a
// single HTML blob (steps recreate_tool -> validate_tool -> save from
// validation_result.clean_html; no writer call anywhere on that path), and the
// rerender side already agrees with that reading — it exempts a self-contained
// tool section from the missing-content escalation by design. Its NULL is correct.
// The declaration is a FIELD rather than a comment on purpose: RFC_010's ruling of
// 2026-08-02 is that authority on a shared seam ships as caller-visible config,
// because "a comment is not a control on a tree this many sessions share".
package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Config keys read by the resolution seam. Named constants so a test and the
// action cannot disagree about the spelling — the failure mode this file exists
// to remove is precisely a string that one side sets and the other never reads.
const (
	sectionsMetadataFieldKey     = "sections_metadata_field"
	expectsNoSectionsMetadataKey = "expects_no_sections_metadata"
	requireSectionsMetadataKey   = "require_sections_metadata"
)

// Origins of the metadata path, reported in the action's result map so an
// operator can tell WHICH of the three states a given save was in without
// re-reading the agent definition.
const (
	metadataOriginConfigured     = "configured"
	metadataOriginDefault        = "default"
	metadataOriginDeclaredAbsent = "declared_absent"
)

// Which representation the persisted sections actually came from.
const (
	sectionsSourceMetadata  = "metadata"
	sectionsSourceHTMLParse = "html_parse"
)

// contentDataRegressionErrorCode is a FOURTH code beside the three link-repair
// ones, for the reason contentDataLinkErrorCode gives for being distinct: it
// answers a different question. Those say what a page's links are; this says that
// a page which HAD structured content has just been saved without any — the
// silence bugs_open/194 is really about, since the save itself succeeds and the
// served page looks perfect.
const contentDataRegressionErrorCode = "CONTENT_DATA_REGRESSION"

// resolveSectionsMetadataField decides which collected_data path holds this
// save's structured sections, and how that decision was reached.
//
// An empty field with origin declared_absent means "this caller has no structured
// content and that is correct"; it is not the same as an empty field because
// nothing was configured, which is the bug. The two are distinguished so the
// report below can stay quiet for the first and fire for the second.
func resolveSectionsMetadataField(config map[string]interface{}) (field string, origin string) {
	if configBoolOrDefault(config, expectsNoSectionsMetadataKey, false) {
		return "", metadataOriginDeclaredAbsent
	}
	if f, ok := config[sectionsMetadataFieldKey].(string); ok && f != "" {
		return f, metadataOriginConfigured
	}
	return defaultSectionsMetadataField, metadataOriginDefault
}

// countSectionsWithContentData reports how many of the section set carry
// structured content. Zero is the 194 signature: the HTML-parse path cannot
// produce content_data at all, so every section arrives without it.
func countSectionsWithContentData(sections []SectionData) int {
	n := 0
	for _, s := range sections {
		if len(s.ContentData) > 0 {
			n++
		}
	}
	return n
}

// shouldReportContentDataLoss decides whether this save is destroying structured
// content the page already had.
//
// The predicate is deliberately the unambiguous one — the page HAD structured
// content on at least one row, and this save carries none at all — because that
// is exactly what falling through to the HTML-parse path looks like, and because
// a partial-loss predicate would fire on legitimate compositions (a page whose
// new plan drops one section keeps content on the rest). Partial loss is
// therefore NOT covered here, and that boundary is stated rather than left for a
// reader to discover: this reports the class 194 is about, not every possible way
// a row can end up with less than it had.
//
// A declared-absent caller is exempt: it never had structured content to lose,
// and firing on it would train an operator to ignore the code.
func shouldReportContentDataLoss(origin string, existingRowsWithContentData int, sections []SectionData) bool {
	if origin == metadataOriginDeclaredAbsent {
		return false
	}
	if existingRowsWithContentData <= 0 {
		return false
	}
	return countSectionsWithContentData(sections) == 0
}

// countExistingRowsWithContentData counts the agent-writable deployed rows this
// save is about to replace that currently hold structured content.
//
// Guarded with pageComponentAgentWritableSQL for the same reason the save's own
// DELETE is: a human-locked row is not this save's to replace, so it is not
// evidence of anything this save is destroying. Best-effort — a counting failure
// returns 0, which suppresses the report rather than inventing one.
func countExistingRowsWithContentData(ctx context.Context, params ActionParams, pageID uuid.UUID) int {
	if params.DB == nil {
		return 0
	}
	var n int
	if err := params.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM page_components
		WHERE page_id = $1
		  AND build_status = 'deployed'
		  AND content_data IS NOT NULL
		  AND `+pageComponentAgentWritableSQL(""), pageID).Scan(&n); err != nil {
		params.Logger.Warn("SavePageSectionsAction: could not count existing content_data rows (194 report)",
			zap.Error(err))
		return 0
	}
	return n
}

// writeContentDataRegressionLog persists the finding, on the SUCCESS path.
//
// A work RECORD, not a work ITEM, and not a refusal — on the precedent
// writeContentDataLinkLog states, and for one more reason specific to this
// action. It already carries five refusing guards, and LANDMINES records the
// consequence: "save_page_sections can now REFUSE a save, so a green orchestration
// status no longer means the sections were written". A sixth unconditional
// refusal on the fleet's highest-traffic save path (page-rerender, 2,878 runs in
// nine days) would be authority taken on a prediction rather than a measurement.
// The record is what makes the measurement possible; opting a caller in to
// require_sections_metadata is the decision it enables, taken later and per
// caller.
//
// Best-effort: a logging failure must never fail a save whose content is
// otherwise correct.
func writeContentDataRegressionLog(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	domain string,
	pageName string,
	metadataField string,
	origin string,
	existingRowsWithContentData int,
	sections []SectionData,
	logger *zap.Logger,
) {
	if params.DB == nil {
		return
	}

	contextJSON, err := json.Marshal(map[string]interface{}{
		"page_name":               pageName,
		"metadata_field":          metadataField,
		"metadata_field_origin":   origin,
		"existing_rows_with_data": existingRowsWithContentData,
		"incoming_sections":       len(sections),
		"bug":                     "bugs_open/194",
	})
	if err != nil {
		logger.Warn("content_data regression: failed to marshal context", zap.Error(err))
		return
	}

	var siteIDArg interface{}
	if siteID != uuid.Nil {
		siteIDArg = siteID
	}

	if _, err := params.DB.ExecContext(ctx, `
		INSERT INTO agent_error_log
		    (site_id, domain, agent_type, step_name, action,
		     error_message, error_code, severity, context)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, 'warning', $8::jsonb)`,
		siteIDArg, domain, componentRepairAgentType(params), saveSectionsStepName(params),
		"save_page_sections",
		fmt.Sprintf("page %q had %d component(s) holding structured content_data and this save carries "+
			"none for any of its %d section(s): the page keeps its HTML but loses the only thing the "+
			"rerender path can regenerate it from. Structured sections were sought at %q (%s). "+
			"If this caller has no structured content by design, declare %s on the step; "+
			"otherwise map %s to the writer's reply (bugs_open/194)",
			pageName, existingRowsWithContentData, len(sections),
			metadataField, origin, expectsNoSectionsMetadataKey, sectionsMetadataFieldKey),
		contentDataRegressionErrorCode,
		string(contextJSON),
	); err != nil {
		logger.Warn("content_data regression: failed to write record", zap.Error(err))
		return
	}

	logger.Warn("SavePageSectionsAction: saving a page that had structured content_data with none (bugs_open/194)",
		zap.String("page_name", pageName),
		zap.String("metadata_field", metadataField),
		zap.String("metadata_field_origin", origin),
		zap.Int("existing_rows_with_content_data", existingRowsWithContentData),
		zap.Int("incoming_sections", len(sections)))
}
