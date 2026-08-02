// FILE: platform/orchestration/actions/save_sections_link_repair.go
//
// Dead-internal-link repair at the PERSISTENCE point (bugs_open/079, REOPENED
// 2026-07-28).
//
// WHY THIS FILE EXISTS. The build gate already repaired dead in-body links —
// ValidatePageContentAction applies RepairPageLinks to cleanHTML and returns it
// as `clean_html` (validate_page_content.go:356-358). But SavePageSectionsAction
// takes the structured path first: when `sections_metadata_field` resolves to a
// non-empty array it persists those per-section strings and never reads
// `html_field` (save_page_sections_action.go:166-186; the html_field fallback at
// :190 runs only when metadata yields zero sections). The live page-build plan
// sets both `sections_metadata_field` and `require_sections_metadata: true`, so
// metadata is ALWAYS present and the repaired string is structurally discarded
// on every page build. Not a race — a dead branch. Measured on three sites:
// fundamentallyai.com/capabilities.html wrote CONTENT_LINK_REPAIR_DETAIL naming
// 10 repairs at 10:45:01.347Z and saved the unrepaired components 400ms later,
// all nine targets serving 404 from the deployed page hours afterwards.
//
// WHY HERE AND NOT IN THE GATE. Repairing `sections_metadata` inside validate
// (candidate 2) would fix the build path and nothing else: page-rerender's
// structured save (034_page_rerender_agent.sql:197-202,
// `sections_metadata_field: rerender_sections.sections_metadata`) has NO
// validate step at all, so a gate-side repair can never cover it. Persistence
// is the one chokepoint every body-section writer passes through —
// page-build-handler, page-rerender, page-rebuild, site-work-orchestrator,
// pageflow-builder, tool-recreation-handler. Putting the repair here makes an
// unrepaired section unsaveable whatever the workflow config says, which is the
// property the config-shaped defect above actually needs.
//
// The outbound rerender seams (rerender_link_repair.go) stay exactly as they
// are. They repair the ASSEMBLED string on its way to deploy; this repairs the
// STORED sections. Both remain necessary — 117/118 record that stored chrome is
// never rebuilt by a page re-render, and the two representations of one page
// have already diverged once (016b §9, "fix applied to one, other left stale").
// After this change the outbound pass is belt-and-braces on the body sections,
// which is the correct relationship between a gate and a persistence guard.

package actions

import (
	"context"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// repairSectionLinks applies the dead-internal-link repair to every section's
// HTML in place, returning an account of every change.
//
// This is the pure seam: no database, no logging, no config. It is where the
// behaviour is tested (save_sections_link_repair_test.go), the same way
// rerender_link_repair_test.go tests its own. The rewrite/unlink SEMANTICS are
// not retested here — datahelpers/link_repair_test.go owns those; what is
// pinned here is that the save path applies them per section and that its
// degrade ships every section untouched.
//
// Fail-open on !indexOK is the load-bearing branch. An untrustworthy page set
// means "we could not read the pages", not "this site has no pages": repairing
// against a half-loaded index would unlink some of the page's real links, and
// against an empty one would unlink all of them. A save must never be degraded
// by a failed metadata read — unrepaired is exactly what shipped before this
// file existed.
//
// Per-section application narrows RepairPageLinks' `data-runtime-fill`
// exemption from whole-document to whole-section. That is deliberate and
// correct: the marker is emitted per section in practice, so a runtime-filled
// section still exempts itself while its statically-linked neighbours no longer
// ride on its exemption.
func repairSectionLinks(sections []SectionData, index datahelpers.PageURLIndex, indexOK bool) (rewritten, unlinked int, repairs []datahelpers.LinkRepair) {
	if !indexOK || len(index) == 0 {
		return 0, 0, nil
	}
	for i := range sections {
		repairedHTML, sectionRepairs := datahelpers.RepairPageLinks(sections[i].HTML, index)
		if len(sectionRepairs) == 0 {
			continue // byte-identical — leave the section alone
		}
		sections[i].HTML = repairedHTML
		repairs = append(repairs, sectionRepairs...)
	}
	rewritten, unlinked = countLinkRepairs(repairs)
	return rewritten, unlinked, repairs
}

// repairSectionsBeforePersist loads the site's real page set and runs the seam
// above over the sections about to be written, persisting a durable account of
// what it changed (or of why it declined to act).
//
// Called from SavePageSectionsAction AFTER the interactive-tool preservation
// block and BEFORE the content-regression guards, so the guards measure the
// bytes that will actually be persisted. Unlinking keeps the anchor text, so
// the stripped-text totals those guards compare are unaffected by this pass.
func repairSectionsBeforePersist(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	domain, pageName, pageURL string,
	sections []SectionData,
	logger *zap.Logger,
) {
	if params.DB == nil || len(sections) == 0 {
		return
	}

	// The reversal lever, same name and same default as the gate's
	// (validate_page_content.go:212-218). It defaults ON because an
	// off-by-default repair would be inert and 079 would still be live; DB
	// config is live-immediately, so the behaviour can still be withdrawn
	// fleet-wide without waiting for an image roll.
	if !configBoolOrDefault(params.StepConfig.Config, "repair_internal_links", true) {
		logger.Info("SavePageSectionsAction: internal-link repair disabled by step config",
			zap.String("page_name", pageName))
		return
	}

	stepName := saveSectionsStepName(params)

	index, indexOK := loadValidPagePaths(ctx, params.DB, siteID, logger)
	if !indexOK {
		logger.Warn("SavePageSectionsAction: link repair SKIPPED — page index unavailable; persisting unrepaired",
			zap.String("page_name", pageName))
		writeLinkRepairSkipLog(ctx, params, siteID.String(), domain,
			linkRepairOrigin{
				AgentType:  componentRepairAgentType(params),
				StepName:   stepName,
				ActionName: "save_page_sections",
				PageName:   pageName,
				PageURL:    pageURL,
			},
			"Section link repair SKIPPED — page index unavailable; sections persisted unrepaired",
			logger)
		return
	}

	rewritten, unlinked, repairs := repairSectionLinks(sections, index, indexOK)
	if len(repairs) == 0 {
		return
	}

	// Distinctive compiled string: this is the pod-grep marker for the deploy.
	logger.Info("SavePageSectionsAction: repaired dead internal links before persist",
		zap.String("page_name", pageName),
		zap.String("domain", domain),
		zap.Int("rewritten", rewritten),
		zap.Int("unlinked", unlinked),
		zap.Int("sections", len(sections)))

	// Same error code as the gate and the rerender seam — the origin fields are
	// what discriminate the three paths, exactly as rerender_link_repair.go:77-85
	// established. A new code would break every query already written against
	// CONTENT_LINK_REPAIR_DETAIL and would make "which path stopped repairing"
	// unanswerable, which is the question 097 exists to keep answerable.
	writeLinkRepairLog(ctx, params, siteID.String(), domain,
		linkRepairOrigin{
			AgentType:  params.AgentType,
			StepName:   stepName,
			ActionName: "save_page_sections",
			PageName:   pageName,
			PageURL:    pageURL,
		},
		repairs, rewritten, unlinked, logger)
}

// The skip record this path writes now lives in component_link_repair.go as the
// ONE writeLinkRepairSkipLog. It used to be a deliberate near-twin of the insert
// in rerender_link_repair.go, disclosed and deferred in 079's council round on
// the ground that merging them would edit a live, proven rerender path for a
// tidiness gain. bugs_open/136 ticketed that extraction, and 136's own fix is
// what forces it: a third writer would have meant a third copy. The row this
// path writes is unchanged — same agent_type, same 'save_page_sections' action,
// same message, same context.

// saveSectionsStepName names the workflow step for the durable record,
// degrading to a stable label when the execution context is absent (unit tests,
// odd adoptions). Distinct from skipStepName, whose fallback label is
// "rerender" and whose chain has no CurrentStep rung.
func saveSectionsStepName(params ActionParams) string {
	return componentRepairStepName(params, "save_sections")
}
