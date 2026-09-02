// FILE: platform/orchestration/actions/dead_url_guard.go
//
// bugs_open/238's detection half: stop a SECTION render shipping a URL
// attribute that resolved to nothing.
//
// The report this consumes is not new. `missingBareFields` already parses the
// template, walks root-scope actions only, and returns the fields sitting inside
// an href=/src= that rendered empty — and it computed exactly
// [card1_image_url … card5_image_url] at the render that shipped bugs_open/238.
// `RenderTemplate` logs it at Error and returns it. ⚠ HISTORICAL NOTE, kept
// because it is the reason this guard exists: until 2026-08-21 there were TWO
// spellings, and the short one — a one-line wrapper — threw the report away
// (`out, _ :=`) while RenderComponentAction called precisely that one. The
// finding was available, by name, on the failing path, for the whole life of the
// defect, and a wrapper hid it. There is now ONE spelling and a test that fails
// the build if a second appears (render_seam_one_spelling_test.go, owner ruling
// 2026-08-21): a caller that discards the report must write `out, _, _, err :=`,
// where a reviewer of the CALL SITE can see it. The site-chrome renderer is the only consumer today
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
// still fires and the render proceeds byte-identically.
//
// ⚠ COVERAGE, CORRECTED 2026-08-11 — my first measurement was wrong and the
// council's debug_historian seat caught the METHOD before anyone caught the
// number. I had counted with `default_config::text LIKE '%render_component%'`,
// which counts AGENTS and returned 1, and in which `_` is a SQL wildcard anyway.
// Re-measured with a jsonb path over `$.**.steps`, the answer is TWO steps, both
// inside page-content-writer: `render_section` AND `render_from_template`. So
// "one config key is full live coverage" was FALSE — arming one key would have
// left the second render path unguarded while the coverage report said armed.
// The HOLD migration arms both and asserts the count, so a future third step
// fails loudly instead of being silently unguarded.
//
// WHAT IT IS NOT: it is not a carry, and it does not overlap PBP-039. The carry
// re-supplies a value the page ALREADY HAD; this fires precisely where there is
// nothing to carry — a first build whose source never resolved, or a rebuild of
// a row that is already damaged. The two are complements, and after PBP-039 this
// guard's population is the residue the carry cannot reach.
//
// ⚠ WHAT REMAINS UNGUARDED, measured 2026-08-11 rather than argued — the council's
// bug_historian seat gated the submission on precisely this (corr 98852baa), and
// it was right to: the "one call site guarded, siblings still exposed" shape is
// bugs_closed/021 and bugs_open/093, and this estate keeps re-closing it.
//
// `RenderTemplate` (component_library.go) discards the report by construction.
// Every remaining caller therefore has the identical silent-empty-URL exposure:
//
//	assemble_from_library.go:288   assembleComponents      — library assembly path
//	section_editor_actions.go:805  applyContentEdit        — human/agent section edit
//	section_editor_actions.go:895  applyComponentSwap      — component swap
//	component_library.go:1943      RenderHeader            — chrome, but NOT the
//	component_library.go:2012      RenderFooter              guarded render_site_
//	component_library.go:2285      RenderHead                components path
//	rerender_pages_actions.go:532  (head template)         — legacy whole-page
//	rerender_pages_actions.go:721  RenderTemplateWithMap   — contact-info block
//
// Guarded today: RenderComponentAction (refuses, opt-in — below),
// rerender_page_sections (records), render_site_components (drops + files).
// So the score is 3 of 11, and the 8 above are stated so a human can size the
// remainder rather than reading one patched step as "the class is closed".
//
// This was NOT widened to all eight in the same change, deliberately: the two
// section-editor paths edit a row a human is holding rather than regenerating
// one, and turning `RenderTemplate` itself into the reporting form fleet-wide is
// a change to the primitive every render in the platform flows through — the
// RFC-shaped move, not a rider on a bug fix.
//
// ⚠ ONE EXEMPTION I CLAIMED IS FALSE, and the council's bug_historian seat asked
// the right question rather than accepting it. I wrote that the chrome renderers
// "already have a dead-control response one layer up (DropDeadURLControls)".
// That holds for `render_site_components_action.go`, which does guard. It does
// NOT hold for `rerender_pages_actions.go:532`, which renders the head template
// with a bare `RenderTemplate` call of its own and never routes through that
// guard — so the legacy whole-page path builds head HTML with exactly the
// exposure the exemption claimed it was covered against. Stated here rather than
// quietly dropped: that path is unguarded, it is on the audit list above, and
// the "chrome is covered" reasoning applies to one of the two chrome routes.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// deadURLGuardConfigKey is the step-config field that arms the refusal. Unset or
// false ⇒ pre-existing behaviour, byte for byte.
const deadURLGuardConfigKey = "refuse_dead_url_controls"

// deadURLRecordConfigKey arms the RECORD-ONLY emit on the re-render path. A
// separate key from the refusal above, because they are different authorities on
// different paths: one declines to ship, the other only files a note. Both
// default OFF.
//
// Added in round 2 of council 98852baa. The first version recorded
// unconditionally, on the reasoning that filing a work item is harmless — and
// three seats (guardian, architecture, render_guardian) independently made the
// same objection: an unconditional new DB write on a shared repair path is new
// authority on a shared seam whatever its size, and the 2026-08-02 owner ruling
// says that ships opt-in with the unsafe default off. They were right, and the
// inconsistency was visible inside my own patch — the refusal was already gated.
const deadURLRecordConfigKey = "record_dead_url_controls"

// recordDeadURLControls reports whether the re-render path should file a
// dead-control item. Same fail-OPEN-on-a-mis-typed-value semantics as
// shouldRefuseDeadURLControls: a config value that is not a bool is a mistake,
// and a mistake must not switch a fleet-wide behaviour on by accident.
func recordDeadURLControls(config map[string]interface{}) bool {
	armed, _ := config[deadURLRecordConfigKey].(bool)
	return armed
}

// shouldRefuseDeadURLControls decides whether a rendered section must be refused
// for shipping empty URL attributes. Pure, so the decision is testable without a
// database or a render.
//
// The data-runtime-fill exemption mirrors the chrome renderer's exactly
// (render_site_components_action.go, render_guardian council note 2026-07-22):
// those shells hydrate their own hrefs client-side, so an empty URL attribute
// there is intentional rather than dead. Two call sites of one judgement, kept
// identical on purpose — the drift class this repo keeps closing.
//
// ⚠ THE EXEMPTION TESTS THE TEMPLATE, NOT THE RENDERED OUTPUT, and the argument
// is that the trigger and the excuse must read the SAME artefact. deadURLFields
// is a fact about the TEMPLATE — bare root-scope {{.Field}} placeholders sitting
// after href=/src= in the template source. Testing the excuse against rendered
// bytes was a category slip: it asked a question about one artefact and took its
// answer from another.
//
// It also makes this structurally safe under features_open/035 composition. A
// composed parent's rendered output EMBEDS its children, so a marker in one child
// would have exempted the parent's own dead controls and its siblings' — bugs_closed/137's
// upward leak, one grain down. Children arrive as DATA ({{.slots.x}} holding
// pre-rendered HTML), never as template text, so a child's marker cannot appear in
// the parent's template and cannot exempt it. No scoping logic needed.
//
// Named predicate rather than a raw string test, per check_runtime_fill_marker in
// scripts/pattern-check.py — the residue of bugs_closed/137, which is CLOSED
// (fixed on v1.0.1223, 2026-07-31). NOTE: that gate's advisory text still cites it
// as "bugs_open/137"; the bug moved and the message did not.
func shouldRefuseDeadURLControls(config map[string]interface{}, deadURLFields []string, htmlTemplate string) bool {
	if len(deadURLFields) == 0 {
		return false
	}
	armed, _ := config[deadURLGuardConfigKey].(bool)
	if !armed {
		return false
	}
	return !datahelpers.HasRuntimeFillMarker(htmlTemplate)
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
