// FILE: platform/orchestration/actions/dead_url_guard.go
//
// bugs_open/238's detection half: stop a SECTION render shipping a URL
// attribute that resolved to nothing.
//
// The report this consumes is not new. `missingBareFields` already parses the
// template, walks root-scope actions only, and returns the fields sitting inside
// an href=/src= that rendered empty — and it computed exactly
// [card1_image_url … card5_image_url] at the render that shipped bugs_open/238.
// `RenderTemplateReportingMissing` logs it at Error. `RenderTemplate` throws it
// away (`out, _, _ :=`), and RenderComponentAction calls RenderTemplate. The
// finding has been available, by name, on the failing path, for the whole life
// of the defect. The site-chrome renderer is the only consumer today
// (DropDeadURLControls + emitChromeDeadControlItem).
//
// WHY REFUSE HERE RATHER THAN DROP, as chrome does. Chrome drops a nav link or
// header CTA — a self-contained control whose absence is clean, and
// DropDeadURLControls was written for chrome markup. A section's dead control is
// typically an <img> inside a card grid: dropping it ships a structurally
// degraded card, and DropDeadURLControls has never been exercised on arbitrary
// section markup. Refusing composes correctly instead — the step fails, so
// save_page_sections never runs, the stored row and the live page are left
// exactly as they were. That is the same shape, at the same altitude, as the
// missingRequiredLLMFields gate ~50 lines above the render this guards, which
// already refuses "to render an empty section … leaving existing content
// untouched".
//
// WHY IT IS OPT-IN AND DEFAULTS OFF. Owner ruling 2026-08-02 (RFC_010/RFC_022):
// new authority on a shared seam ships as a field with the unsafe default OFF,
// so the decision is visible where a reviewer of the CALLER can see it. With the
// flag unset this file changes nothing — the Error log that already existed
// still fires and the render proceeds byte-identically. Measured 2026-08-10,
// exactly ONE live agent has a render_component step (page-content-writer), so
// flipping one config value is full live coverage and is trivially reversible.
//
// WHAT IT IS NOT: it is not a carry, and it does not overlap PBP-039. The carry
// re-supplies a value the page ALREADY HAD; this fires precisely where there is
// nothing to carry — a first build whose source never resolved, or a rebuild of
// a row that is already damaged. The two are complements, and after PBP-039 this
// guard's population is the residue the carry cannot reach.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// deadURLGuardConfigKey is the step-config field that arms the refusal. Unset or
// false ⇒ pre-existing behaviour, byte for byte.
const deadURLGuardConfigKey = "refuse_dead_url_controls"

// shouldRefuseDeadURLControls decides whether a rendered section must be refused
// for shipping empty URL attributes. Pure, so the decision is testable without a
// database or a render.
//
// The data-runtime-fill exemption mirrors the chrome renderer's exactly
// (render_site_components_action.go, render_guardian council note 2026-07-22):
// those shells hydrate their own hrefs client-side, so an empty URL attribute
// there is intentional rather than dead. Two call sites of one judgement, kept
// identical on purpose — the drift class this repo keeps closing.
func shouldRefuseDeadURLControls(config map[string]interface{}, deadURLFields []string, rendered string) bool {
	if len(deadURLFields) == 0 {
		return false
	}
	armed, _ := config[deadURLGuardConfigKey].(bool)
	if !armed {
		return false
	}
	return !strings.Contains(rendered, "data-runtime-fill")
}

// emitSectionDeadControlItem files the human-visible record of a dead section
// control. Mirrors emitChromeDeadControlItem: the shared insertWorkItem helper
// (so it inherits the idx_swi_dedup-matched ON CONFLICT and the two-strike
// anti-churn label), needs_human_review with NO handler — nothing automated can
// invent a missing image or destination — and a failure that is logged, never
// returned, because the refusal must not depend on the record being written.
//
// THE ITEM KEY CARRIES PAGE AND SLOT, and that is load-bearing rather than
// cosmetic. Its nearest sibling, image_url_404's empty-src finding, uses the
// site-wide key `image_url_404:empty-src`; finetuning.uk has held one of those
// at status `blocked` since 2026-08-03, and `blocked` is not in
// workItemTerminalStatuses, so the dedup index has kept that single slot
// occupied ever since. New damage on a different page of that site could not
// mint an item at all — the detector was live, correct, and structurally unable
// to report. A key without page+slot reproduces that.
func emitSectionDeadControlItem(ctx context.Context, db *sql.DB, siteID uuid.UUID, componentID *uuid.UUID,
	pageName, slot, componentFunction string, deadURLFields []string, refused bool, logger *zap.Logger) {

	if db == nil || siteID == uuid.Nil {
		// Identity we cannot establish is not a reason to skip the refusal —
		// only a reason to skip the record. Say so rather than failing quietly.
		logger.Warn("dead URL control: no site identity available, work item not filed",
			zap.String("slot", slot),
			zap.Strings("dead_url_fields", deadURLFields))
		return
	}

	fieldList := strings.Join(deadURLFields, ", ")
	disposition := "recorded (the section still rendered)"
	if refused {
		disposition = "REFUSED (the section was not rendered and nothing was saved)"
	}

	spec := map[string]interface{}{
		"surface":            "page_section",
		"page_name":          pageName,
		"slot_name":          slot,
		"component_function": componentFunction,
		"dead_url_fields":    deadURLFields,
		"refused":            refused,
		"source":             "render_component",
		"fix": "A section control (image, card link, CTA) had no resolvable " +
			"destination, so it would render as an empty src=\"\"/href=\"\" — a " +
			"broken image or a control that silently disappears if the template " +
			"gates it (bugs_open/238). Decide: populate the field's declared " +
			"source (the site_specs aspect, or the site_plan_imagery row behind " +
			"site_assets.*), generate the missing asset, or gate the template on " +
			"the field. Never point a link at /contact.html to fill the hole (the " +
			"LNK-007 fossil).",
	}
	specJSON, _ := json.Marshal(spec)

	summary := fmt.Sprintf("Dead URL control on %s/%s: no destination for %s — %s",
		pageName, slot, fieldList, disposition)
	if len(summary) > 250 {
		summary = summary[:247] + "..."
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("dead URL control: begin tx failed", zap.String("slot", slot), zap.Error(err))
		return
	}
	inserted, err := insertWorkItem(ctx, tx, workItem{
		siteID:      siteID,
		componentID: componentID,
		source:      "render-component",
		pipeline:    "build",
		itemType:    "dead_url_control",
		severity:    "high",
		summary:     summary,
		spec:        string(specJSON),
		priority:    40,
		status:      "needs_human_review",
		createdBy:   "render_component",
		itemKey:     deadURLControlItemKey(pageName, slot, deadURLFields),
	}, logger)
	if err != nil {
		_ = tx.Rollback()
		logger.Warn("dead URL control: insert work item failed", zap.String("slot", slot), zap.Error(err))
		return
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("dead URL control: commit failed", zap.String("slot", slot), zap.Error(err))
		return
	}
	if inserted {
		logger.Info("dead URL control: item filed for review",
			zap.String("page", pageName),
			zap.String("slot", slot),
			zap.Bool("refused", refused),
			zap.Strings("dead_url_fields", deadURLFields))
	}
}

// deadURLControlItemKey builds the dedup key. Page and slot are both present so
// one page's unresolved item cannot hold the slot against every other page on
// the site — see emitSectionDeadControlItem's note on image_url_404:empty-src.
// deadFields arrives pre-sorted from missingBareFields, so the key is stable
// across runs of the same defect.
func deadURLControlItemKey(pageName, slot string, deadFields []string) string {
	return fmt.Sprintf("dead_url_control:%s:%s:%s", pageName, slot, strings.Join(deadFields, ","))
}
