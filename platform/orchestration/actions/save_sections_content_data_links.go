// FILE: platform/orchestration/actions/save_sections_content_data_links.go
//
// The content_data half of dead-internal-link handling (bugs_open/097), at the
// same persistence chokepoint as the markup half (save_sections_link_repair.go).
//
// WHY A SECOND PASS AND NOT A WIDER FIRST ONE. The two operate on different
// objects and cannot be one function. RepairPageLinks parses anchors out of a
// rendered STRING; this resolves typed FIELDS in a structured map. They share
// the one thing that must never diverge — the page index, the scope classifier
// and the normaliser (datahelpers/links.go) — so the gate, the markup repair,
// the post-deploy audit and this pass can never disagree about what a real page
// is. The index is loaded ONCE and handed to both, for the reason
// validate_page_content.go:272-277 gives for doing the same between check 4 and
// its repair: two loads are two chances to differ.
//
// WHY IT MATTERS THAT content_data IS COVERED AT ALL. The markup repair fixes
// the copy that ships. content_data is the copy the platform RE-RENDERS FROM:
// rerender_page_sections rebuilds each section from it with no writer pass, so
// a dead href left in content_data is regenerated on every re-render, repaired
// again on the way out, and never once reported as the authoring defect it is.
// bugs_open/097's own words for that state: "the repair now deletes a phantom
// link silently, and the authoring defect that wrote it goes unreported."
// component_link_repair.go names the same limit and routes it here.
//
// WHAT THIS DOES NOT CLAIM. It covers the six agent types that persist body
// sections through SavePageSectionsAction. It is not fleet-wide coverage of
// page_components.content_data: the single-component writers listed in
// bugs_open/136 set rendered_html and, where they set content_data, do not pass
// through here. Extending to them is the same call shape component_link_repair.go
// already established and is deliberately left as a separate change rather than
// ridden in on this one.
//
// THE MAP IS NOT A COPY, and a reader should know it rather than assume it.
// extractSectionsFromMetadata assigns SectionData.ContentData by type assertion
// straight off the sections_metadata entry (save_page_sections_action.go:905-907),
// so the map this pass rewrites is the SAME object still sitting in
// params.CollectedData. That is the direction you want — a correction made here is
// what gets persisted at :650 AND what any later step in the same run reads, so
// the two cannot disagree — but it does mean the rewrite arm is visible outside
// this function. It is stated here because "the audit works on a copy" is the
// natural assumption and it is wrong.

package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// contentDataLinkErrorCode is a THIRD code beside CONTENT_LINK_REPAIR_DETAIL
// and CONTENT_LINK_REPAIR_SKIPPED, for the reason linkRepairErrorCode gives for
// being distinct from validationDetailErrorCode: it answers a different
// question. The repair rows say what the gate changed in the markup that
// shipped; these say what the page's SOURCE still holds and who authored it.
// Folding them together would mean every existing query on
// CONTENT_LINK_REPAIR_DETAIL silently started counting a different population.
const contentDataLinkErrorCode = "CONTENT_DATA_LINK_AUDIT"

// contentDataLinkFinding is one finding with the component it was authored in.
// datahelpers reports the path within a content_data map; only the caller knows
// which component that map belongs to, and "which component authors phantoms"
// is the question the record exists to answer.
type contentDataLinkFinding struct {
	Component string `json:"component"`
	Path      string `json:"path"`
	Href      string `json:"href"`
	NewHref   string `json:"new_href,omitempty"`
	Action    string `json:"action"`
}

// auditSectionContentDataLinks resolves the URL-bearing fields of every
// section's content_data against the real page set, rewriting the
// extension-only misses in place and returning every finding.
//
// The pure seam: no database, no logging, no config — the same split
// repairSectionLinks uses, and for the same reason. Its degrade is identical
// too: an untrustworthy index leaves every section untouched, because rewriting
// against a half-loaded page set would be rewriting against a partial truth and
// reporting against one would call a site's real pages phantoms.
//
// A runtime-fill section is exempted, per section, exactly as the markup pass
// exempts it. Two reasons, and the second is the load-bearing one: its stored
// values are placeholders the client replaces, so judging them judges something
// no visitor sees; and skipping in lockstep with the markup pass is what keeps
// the two representations of one section from diverging — a rewrite here with
// no matching rewrite there would leave content_data and rendered_html
// disagreeing about the same link until the next re-render.
func auditSectionContentDataLinks(sections []SectionData, index datahelpers.PageURLIndex, indexOK bool) []contentDataLinkFinding {
	if !indexOK || len(index) == 0 {
		return nil
	}
	var out []contentDataLinkFinding
	for i := range sections {
		if datahelpers.HasRuntimeFillMarker(sections[i].HTML) {
			continue
		}
		for _, f := range datahelpers.RepairContentDataLinks(sections[i].ContentData, index) {
			out = append(out, contentDataLinkFinding{
				Component: sections[i].ComponentName,
				Path:      f.Path,
				Href:      f.Href,
				NewHref:   f.NewHref,
				Action:    f.Action,
			})
		}
	}
	return out
}

// countContentDataLinkFindings splits the findings by arm for the log line and
// the durable record.
func countContentDataLinkFindings(findings []contentDataLinkFinding) (rewritten, phantom int) {
	for _, f := range findings {
		switch f.Action {
		case datahelpers.LinkRepairRewrite:
			rewritten++
		case datahelpers.ContentDataLinkPhantom:
			phantom++
		}
	}
	return rewritten, phantom
}

// writeContentDataLinkLog persists the audit, on the success path.
//
// A work RECORD, not a work ITEM, on exactly the precedent writeLinkRepairLog
// set and for the reasons it states: bugs_open/083 measured phantom_internal_link
// detected 22 times and fixed zero times ever, because the only thing that
// promotes 'detected' rows lives in the disabled improvement-sweep, and
// bugs_open/077 warns against filing items whose handler has no remit. Nothing
// the platform has today would drain a content_data-authoring queue. What is
// owed and previously missing is an auditable account naming the component, the
// field and the href — recoverable long after collected_data is pruned at ~24h.
//
// Best-effort: a logging failure must never fail a save whose content is
// otherwise correct.
func writeContentDataLinkLog(
	ctx context.Context,
	params ActionParams,
	siteIDStr string,
	domain string,
	origin linkRepairOrigin,
	findings []contentDataLinkFinding,
	rewritten, phantom int,
	logger *zap.Logger,
) {
	if params.DB == nil || len(findings) == 0 {
		return
	}

	contextJSON, err := json.Marshal(map[string]interface{}{
		"rewritten": rewritten,
		"phantom":   phantom,
		"findings":  findings,
		"page_name": origin.PageName,
		"page_url":  origin.PageURL,
	})
	if err != nil {
		logger.Warn("content_data link audit: failed to marshal context", zap.Error(err))
		return
	}

	var siteIDArg interface{}
	if siteIDStr != "" {
		if id, err := uuid.Parse(siteIDStr); err == nil {
			siteIDArg = id
		}
	}

	if _, err := params.DB.ExecContext(ctx, `
		INSERT INTO agent_error_log
		    (site_id, domain, agent_type, step_name, action,
		     error_message, error_code, severity, context)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, 'warning', $8::jsonb)`,
		siteIDArg, domain, origin.AgentType, origin.StepName, origin.ActionName,
		fmt.Sprintf("content_data holds %d internal link(s) that resolve to no page: "+
			"%d rewritten to the real stored url, %d phantom and left for authoring review; see context.findings",
			len(findings), rewritten, phantom),
		contentDataLinkErrorCode,
		string(contextJSON),
	); err != nil {
		logger.Warn("content_data link audit: failed to write record", zap.Error(err))
		return
	}

	logger.Info("SavePageSectionsAction: audited content_data internal links",
		zap.Int("rewritten", rewritten),
		zap.Int("phantom", phantom))
}
