// FILE: platform/orchestration/actions/component_link_repair.go
//
// Dead-internal-link repair for the writers that persist ONE component's
// rendered_html without going through SavePageSectionsAction
// (bugs_open/136 — the section-editor slug, filed by the council's
// bug_historian seat against bugs_open/079's fix).
//
// WHY THIS FILE EXISTS. 079 put the repair at the persistence point of the
// full-page section save, which is where LLM-authored body prose normally
// enters. The seat's objection was that this was never shown to be the only
// writer of page_components.rendered_html — only the only one with a
// 'save_page_sections' STEP NAME — and the check it asked for found three more.
// The family is the one 016b §9 and CLAUDE.md both name: a mechanism made
// generic, then guarded at one call site while the siblings stay open.
//
// WHY A WRAPPER AND NOT A SECOND IMPLEMENTATION. The repair SEMANTICS live in
// repairSectionLinks (save_sections_link_repair.go) and are tested there. A
// single-component caller is that function with one element, which is exactly
// what the bug file proposed — so this file adds a CALL SHAPE, not a second set
// of rules. Two hand-maintained copies of one repair is the drift class this
// platform reviews for; the fail-open branch in particular must never be
// reasoned about twice.
//
// WHAT IT DELIBERATELY DOES NOT COVER.
//   - RebuildBlogListingAction. Not a member of the class: its article hrefs
//     come from pages.url (blogPostsQuery -> articles[].url), the same table
//     the repair index is built from, under a STRICTER predicate — and the live
//     content-listing template carries exactly one anchor, href="{{.url}}"
//     (measured 2026-08-02). A repair pass there could only ever no-op, and a
//     no-op call still costs a pages query per rebuild.
//   - The tool-markup writers (create_tool_component_action.go,
//     deploy_tool_action.go). These ARE members and are the ranked next
//     candidate — the 2026-08-02 census puts 8 of 30 live unresolved internal
//     links in tool slots. Left out on collision grounds, not merit: those files
//     are in play across the tool lanes, and a pathspec commit still takes a
//     same-file passenger.
//   - content_data. Same limit as 079's fix: this repairs the STORED HTML, and
//     a re-render from content_data can reintroduce the href in storage. The
//     deployed artefact stays covered by the outbound rerender seam
//     (repairOutboundPageLinks, bugs_open/097).

package actions

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// repairComponentHTMLBeforePersist repairs the dead internal links in ONE
// component's HTML and returns what should be written to the database.
//
// Call it immediately before the UPDATE/INSERT that persists the column, not at
// the point the HTML is rendered: a writer with two persist sites and a repair
// at one of them is this bug's own shape one level in.
//
// Fail-open, in three places and for the same reason each time — an unrepaired
// component is exactly what shipped before this file existed, and no repair may
// cost a caller its write:
//   - no DB or empty HTML: nothing to do;
//   - repair_internal_links=false in step config: the reversal lever, live
//     immediately in DB config, same name and same ON default as the gate's;
//   - an untrustworthy page index: repairing against a half-loaded index would
//     unlink some of the page's real links and against an empty one would
//     unlink all of them, so the HTML is returned untouched and the skip is
//     recorded durably (a pod log line is not a record — 071 gap 3).
//
// actionName is what discriminates this path in the durable record. It is the
// caller's own action name ("apply_section_edit", "create_report_page"), never a
// shared label: a row that cannot say which path acted cannot be used to spot
// the path that STOPPED acting, which is the drift bugs_open/097 exists to keep
// answerable.
func repairComponentHTMLBeforePersist(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	domain, pageName, pageURL, actionName, html string,
	logger *zap.Logger,
) string {
	if params.DB == nil || html == "" {
		return html
	}

	if !configBoolOrDefault(params.StepConfig.Config, "repair_internal_links", true) {
		logger.Info("component link repair disabled by step config",
			zap.String("action", actionName),
			zap.String("page_name", pageName))
		return html
	}

	origin := linkRepairOrigin{
		AgentType:  componentRepairAgentType(params),
		StepName:   componentRepairStepName(params, actionName),
		ActionName: actionName,
		PageName:   pageName,
		PageURL:    pageURL,
	}

	index, indexOK := loadValidPagePaths(ctx, params.DB, siteID, logger)
	if !indexOK {
		logger.Warn("component link repair SKIPPED — page index unavailable; persisting unrepaired",
			zap.String("action", actionName),
			zap.String("page_name", pageName),
			zap.String("domain", domain))
		writeLinkRepairSkipLog(ctx, params, siteID.String(), domain, origin,
			"Component link repair SKIPPED — page index unavailable; component persisted unrepaired",
			logger)
		return html
	}

	// One element, so the semantics stay defined in exactly one place. Note
	// this also inherits repairSectionLinks' narrowing of the data-runtime-fill
	// exemption to the section being written, which is the correct scope here:
	// the unit being persisted IS one component.
	sections := []SectionData{{HTML: html}}
	rewritten, unlinked, repairs := repairSectionLinks(sections, index, indexOK)
	if len(repairs) == 0 {
		return html // byte-identical — a clean component is never perturbed
	}

	// Distinctive compiled string: this is the pod-grep marker for the deploy.
	logger.Info("repaired dead internal links before persisting a single component",
		zap.String("action", actionName),
		zap.String("page_name", pageName),
		zap.String("domain", domain),
		zap.Int("rewritten", rewritten),
		zap.Int("unlinked", unlinked))

	writeLinkRepairLog(ctx, params, siteID.String(), domain, origin,
		repairs, rewritten, unlinked, logger)

	return sections[0].HTML
}

// componentRepairAgentType names WHO repaired, degrading to a stable label
// rather than writing an empty agent_type — the column is what a later query
// groups by, and "" and "unknown" would split one path across two buckets.
func componentRepairAgentType(params ActionParams) string {
	if params.AgentType != "" {
		return params.AgentType
	}
	return "unknown"
}

// componentRepairStepName names the workflow step for the durable record,
// degrading to the caller's action name when the execution context is absent
// (unit tests, odd adoptions).
//
// saveSectionsStepName delegates here — its chain is identical with the
// fallback "save_sections". skipStepName deliberately does NOT: its chain has
// no CurrentStep rung, and adding one would change the step_name written by a
// live, proven rerender path for a tidiness gain. Folding those two together is
// a change to make on its own merits, not as a passenger here.
func componentRepairStepName(params ActionParams, fallback string) string {
	if params.ExecutionContext != nil && params.ExecutionContext.StepName != "" {
		return params.ExecutionContext.StepName
	}
	if params.CurrentStep != "" {
		return params.CurrentStep
	}
	return fallback
}

// writeLinkRepairSkipLog records that a component or page was persisted or
// deployed WITHOUT repair because the page index could not be trusted.
//
// THE ONE COPY, extracted here (bugs_open/136's own tail ticket). It existed
// twice — save_sections_link_repair.go and rerender_link_repair.go:52-63 — and
// the reuse_agent seat objected to the duplication at medium severity in 079's
// round, accepting the deferral on condition it was ticketed. This change would
// have written the THIRD copy, which is when a deferral stops being defensible.
//
// Every caller keeps its own agent_type, action and message, so the rows are
// byte-identical to the ones already written and every query matching on
// CONTENT_LINK_REPAIR_SKIPPED keeps returning what it returned before.
//
// Best-effort, like the repair log it mirrors: a logging failure must never fail
// the write it describes.
func writeLinkRepairSkipLog(
	ctx context.Context,
	params ActionParams,
	siteIDStr string,
	domain string,
	origin linkRepairOrigin,
	message string,
	logger *zap.Logger,
) {
	if params.DB == nil {
		return
	}

	skipCtx, err := json.Marshal(map[string]string{
		"page_name": origin.PageName,
		"page_url":  origin.PageURL,
	})
	if err != nil {
		logger.Warn("link repair: failed to marshal skip context", zap.Error(err))
		return
	}

	if _, err := params.DB.ExecContext(ctx, `
		INSERT INTO agent_error_log
		    (site_id, domain, agent_type, step_name, action,
		     error_message, error_code, severity, context)
		VALUES (NULLIF($1,'')::uuid, NULLIF($2, ''), $3, $4, $5,
		        $6, 'CONTENT_LINK_REPAIR_SKIPPED', 'warning', $7::jsonb)`,
		siteIDStr, domain, origin.AgentType, origin.StepName, origin.ActionName,
		message, string(skipCtx),
	); err != nil {
		logger.Warn("link repair: failed to write skip record", zap.Error(err))
	}
}
