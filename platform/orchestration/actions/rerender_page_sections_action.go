// FILE: platform/orchestration/actions/rerender_page_sections_action.go
//
// RerenderPageSectionsAction re-renders ALL of a page's sections from their
// STORED content_data plus FRESHLY re-resolved dynamic fields, WITHOUT invoking
// the content writer (no LLM). It is the lightweight path for "a resolved field
// changed, re-render the page" — specifically an image asset landing
// (hero/section image) or deferred section data becoming query-resolvable —
// where the page's copy is unchanged and only resolved fields (asset URLs,
// query-backed lists) need refreshing.
//
// WHY (and why not the writer): routing these through page-build-handler ->
// page-content-writer regenerates copy via the LLM (cost, latency) and exposes
// an asset swap to the content-regression guard. The writer already proved
// content_data is a complete render source: RenderComponentAction persists
// content_data = LLM copy overlaid with resolved_data. So we re-render each
// section by feeding STORED content_data as the content and FRESH resolved_data
// on top, through the same RenderTemplate path render_component uses.
//
// MECHANISM (all reuse, in-package):
//   - newSourceResolver + planSection (plan_sections_action.go) rebuild each
//     section's resolved_data (queryresolve for query.*, the page-aware
//     ensureAssets for the hero). planSection is SIDE-EFFECT-FREE — the
//     needs_new_component / deferred-item writes live in PlanSectionsAction's
//     caller loop, not in planSection itself.
//   - RenderTemplate renders html_template against a RenderContext built from a
//     minimal site base + stored content_data + fresh resolved_data
//     (resolved_data merged last so it wins, matching RenderComponentAction).
//   - Emits sections_metadata in the exact shape CompilePageSectionsAction
//     produces ({rendered_html, component_id, component_name,
//     component_function, content_data}), so save_page_sections ingests it
//     unchanged — no compile step needed.
//
// NULL content_data (older pages that predate the writer's content_data
// capture): re-render-all needs EVERY section to have stored content. If ANY
// section's content_data is missing, this escalates the WHOLE page to the
// content generator (emits needs_page -> page-build-handler), which regenerates
// AND backfills content_data; the light path then works on the next trigger.
//
// OUTPUT (output_field, e.g. "rerender_sections"):
//   { sections_metadata, page_id, site_id, domain, page_name,
//     escalated (bool), skipped (bool), section_count, rerendered, carried }
// page_id/site_id/domain are surfaced so the downstream render_page
// (rerender_single_page) finds them via ExtractFields' recursive search — the
// work item itself only carries page_name.
//
// WIRING (page-rerender workflow, as a pre-pass gated by spec.reason):
//   check_rerender_mode (conditional: reason==image_landed OR
//     reason==section_data_resolved) -> rerender_sections -> check_escalated
//     -> save_sections -> render_page -> (check_skipped -> deploy -> status)
//   else_step (no/other reason) -> render_page   (unchanged assemble-only path)
//
// REGISTRATION (registry.go):
//   "rerender_page_sections": {
//       Handler:     RerenderPageSectionsAction,
//       Category:    "site",
//       Description: "Re-render a page's sections from stored content_data + fresh resolved fields (no LLM)",
//       IsLocal:     true,
//   }

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/livespec"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RerenderPageSectionsInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	// target_site_id (NOT site_id) per 001 §Field name collisions: site_id is a
	// key on the nested-source objects (site_record.site_id, input_data.site_id),
	// so a bare site_id can be silently bound from the wrong source. The wiring
	// maps it explicitly: "target_site_id": "input_data.site_id". Same precedent
	// as reconcile_site_plan.
	// page_name moved Required -> Optional (bugs_open/094). Callers that know the
	// page by ID had no way in: the required-field check rejected the envelope
	// before the action ran, with "missing required fields: [page_name]", even
	// though the action's very next act is a DB lookup that could have resolved
	// it. 049b_deploy_single_page.sh publishes page_id/site_id/domain and could
	// not use its own documented section_data_resolved branch at all.
	//
	// EITHER page_name OR page_id is now sufficient and the action derives the
	// other. That is what makes the bad envelope unrepresentable rather than
	// merely documented — every existing caller is fixed at once, including any
	// not yet found.
	Required: []string{"target_site_id"},
	Optional: []string{"page_name", "page_id", "reason"},
	// strip_literal_markdown is a SETTING, not a data reference (bugs_open/184):
	// when true, stored content_data is passed through
	// datahelpers.StripLiteralMarkdownFromContentData before it feeds both the
	// render context and the persisted mergedContent — which is what makes a
	// plain rerender the mechanical repair for literal_markdown items. Default
	// OFF; enabled on page-rerender's rerender_sections step by migration 473.
	//
	// record_dead_url_controls is bugs_open/238's RECORD-ONLY half (PBP-040),
	// DECLARED here 2026-08-20 rather than added: this action has read it since
	// the guard shipped, but through recordDeadURLControls(params.StepConfig.Config)
	// rather than a literal config["..."] in the function body — the same reason
	// refuse_dead_url_controls went undeclared on RenderComponentInputSpec until
	// 2026-08-19, and a grep over the function body cannot see either.
	//
	// Declaring it costs nothing and buys the unknown-config-key report's
	// honesty: without this line, migration 497 arming the key on
	// page-rerender's rerender_sections step makes a LIVE, WORKING setting read
	// as "keys this action does not read — silently ignored at execution"
	// (platform/validation/workflow.go). That is the report's one dishonesty
	// surface, and it points the next reader at deleting the key.
	//
	// It is NOT a StrictConfig spec, so the undeclared state warns rather than
	// failing — verified at the deciding arm before arming, because the same
	// mistake on a StrictConfig spec took the fleet's page-publishing path down
	// for 33 minutes on 2026-08-19 (WRONG_CALLS, migration 494).
	//
	// No RFC_022 budget impact: the optional-key counter reads spec.Optional and
	// skips ConfigKeys on purpose (cmd/config-key-audit/optionalbudget.go) —
	// settings are not the accumulated authority it was built to notice, and
	// both dead-URL flags are settings.
	ConfigKeys: []string{"strip_literal_markdown", "record_dead_url_controls"},
	Defaults:   map[string]interface{}{},
}

func init() {
	datahelpers.RegisterActionInputSpec("rerender_page_sections", RerenderPageSectionsInputSpec)
}

// shouldStripLiteralMarkdown is the rerender strip's DOUBLE GATE (bugs_open/184;
// council 060bcc0a r2 guardian): the step config flag arms the capability, the
// dispatch's own spec.reason scopes it to the literal_markdown repair. Either
// alone must NOT strip — the flag alone would strip on EVERY sections-branch
// rerender (image_landed/template_changed/cta_links_stale, the fleet's busiest
// pipeline), and the reason alone means the operator has not enabled the
// capability. Extracted so the containment property is asserted by a direct
// unit test (rerender_strip_gate_test.go), not inferred from code review.
func shouldStripLiteralMarkdown(stepConfig map[string]interface{}, reason string) bool {
	on, _ := stepConfig["strip_literal_markdown"].(bool)
	// bugs_open/404: name the vocabulary value, do not re-spell it. Retiring or
	// renaming it must break the build here, not silently disarm this gate.
	return on && reason == livespec.ReasonLiteralMarkdown
}

// storedSection is one page_components row as loaded for re-render.
type storedSection struct {
	// id and parentInstanceID are the composition pair (features_open/035 D1).
	// parentInstanceID is "" for an ordinary top-level section, which is every
	// row on the estate today — 0 of 1,903 rows carry one as of 2026-08-26 — so
	// reading them changes nothing until a composed page exists.
	id               string
	parentInstanceID string

	componentID  string
	slotName     string
	contentData  map[string]interface{}
	renderedHTML string
	position     int
	// componentVersionID is the provenance stamp this row already carries
	// (RFC_046). A carry does not change the bytes, so it must not change what
	// produced them — without this the save re-inserts the row with no stamp and
	// a fact the estate already owned is lost on every rerender.
	componentVersionID string
}

func RerenderPageSectionsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "rerender_page_sections"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		RerenderPageSectionsInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("target_site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid target_site_id %q: %w", inputs.Get("target_site_id"), err)
	}
	pageName := inputs.Get("page_name")
	pageIDIn := inputs.Get("page_id")
	if pageName == "" && pageIDIn == "" {
		return nil, fmt.Errorf("need page_name or page_id (bugs_open/094): got neither")
	}
	reason := inputs.Get("reason")

	// ── Resolve page_id + domain (also surfaced for the downstream render_page,
	//    which finds them via ExtractFields' recursive search); page url feeds
	//    the CTA recompute's self-link test ────────────────────────────────────
	//
	// bugs_open/094: resolve by whichever key the caller supplied. Both branches
	// are scoped to the site, so a page_id belonging to a DIFFERENT site is a
	// not-found rather than a cross-site re-render — target_site_id stays
	// authoritative and page_id cannot be used to reach past it.
	var pageID uuid.UUID
	var domain, pageURL string
	if pageName != "" {
		err = params.DB.QueryRowContext(ctx, `
			SELECT p.id, s.domain, COALESCE(p.url, ''), p.name
			FROM pages p
			JOIN sites s ON s.id = p.site_id
			WHERE p.site_id = $1 AND p.name = $2
		`, siteID, pageName).Scan(&pageID, &domain, &pageURL, &pageName)
	} else {
		parsed, perr := uuid.Parse(pageIDIn)
		if perr != nil {
			return nil, fmt.Errorf("invalid page_id %q: %w", pageIDIn, perr)
		}
		err = params.DB.QueryRowContext(ctx, `
			SELECT p.id, s.domain, COALESCE(p.url, ''), p.name
			FROM pages p
			JOIN sites s ON s.id = p.site_id
			WHERE p.site_id = $1 AND p.id = $2
		`, siteID, parsed).Scan(&pageID, &domain, &pageURL, &pageName)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			// Name the key that was actually used, so "not found" is actionable
			// rather than ambiguous about which lookup failed.
			if pageIDIn != "" && inputs.Get("page_name") == "" {
				return nil, fmt.Errorf("page_id %s not found for site %s", pageIDIn, siteID)
			}
			return nil, fmt.Errorf("page %q not found for site %s", pageName, siteID)
		}
		return nil, fmt.Errorf("resolve page: %w", err)
	}

	// bugs_open/094, council objection (bug_historian, medium): page_id is not
	// mapped by any step config, so it arrives via ExtractActionInputs Strategy 2,
	// whose own comment warns it "uses aggressive recursive search that can find
	// stale values". Before this change a missing page_name failed LOUDLY at the
	// input-spec gate; now a stale page_id could in principle resolve a DIFFERENT
	// page of the same site and re-render it silently.
	//
	// The site scoping stops it crossing sites; it does not stop it picking the
	// wrong page WITHIN a site. So record which key was used and what it resolved
	// to. This does not prevent the wrong resolution — it makes it attributable
	// instead of invisible, which is the difference between a bug you can find
	// and one you cannot.
	resolvedBy := "page_name"
	if inputs.Get("page_name") == "" {
		resolvedBy = "page_id (recursive search — no step config maps it)"
	}
	logger.Info("rerender_page_sections: page resolved",
		zap.String("resolved_by", resolvedBy),
		zap.String("page_name", pageName),
		zap.String("page_id", pageID.String()),
		zap.String("site_id", siteID.String()),
		zap.String("url", pageURL))

	out := map[string]interface{}{
		"page_id":   pageID.String(),
		"site_id":   siteID.String(),
		"domain":    domain,
		"page_name": pageName,
		"escalated": false,
		"skipped":   false,
	}

	// ── Load stored sections ────────────────────────────────────────────────
	stored, err := loadStoredSections(ctx, params.DB, pageID, logger)
	if err != nil {
		return nil, fmt.Errorf("load stored sections: %w", err)
	}

	// bugs_open/184: with strip_literal_markdown enabled (migration 473 sets it
	// on page-rerender's rerender_sections step) AND this dispatch's own
	// spec.reason naming the repair, a rerender IS the mechanical repair for
	// literal markdown: each section's STORED content_data is stripped here,
	// before it feeds both the render context (the render loop) and the
	// persisted mergedContent — so one no-LLM rerender heals both surfaces.
	//
	// DOUBLE-GATED ON PURPOSE (council round 2, guardian): the step config flag
	// alone would strip on EVERY sections-branch rerender — image_landed,
	// template_changed, cta_links_stale, the fleet's highest-volume pipeline —
	// which is a blanket behavioural change dressed as a scoped repair. The
	// reason gate bounds the blast radius to exactly the dispatched repair;
	// other reasons carry markdown through untouched, where the discovery check
	// files a literal_markdown item and THIS path repairs it, scoped.
	//
	// Strip-only, letter-guarded, identical patterns to the discovery check
	// (single-sourced in datahelpers), so the completion verifier finds nothing
	// left by construction. Runs before the content pre-check below on purpose:
	// stripping never removes a field, so the required-field contract is
	// unaffected. The changed paths are surfaced on the action OUTPUT (below),
	// not just logged — pod logs rotate in minutes; collected_data is the
	// durable record (council round 2, bug_historian) — and the pre-strip state
	// is archived by save_page_sections into page_component_history before the
	// replace, so a disputed strip is recoverable, not lost.
	var strippedMarkdownFields []string
	if shouldStripLiteralMarkdown(params.StepConfig.Config, reason) {
		for _, s := range stored {
			if len(s.contentData) == 0 {
				continue
			}
			if changed := datahelpers.StripLiteralMarkdownFromContentData(s.contentData); len(changed) > 0 {
				for _, f := range changed {
					strippedMarkdownFields = append(strippedMarkdownFields, s.slotName+":"+f)
				}
				logger.Info("rerender_page_sections: stripped literal markdown from stored content_data",
					zap.String("slot", s.slotName),
					zap.Strings("fields", changed))
			}
		}
	}
	// (out["stripped_markdown_fields"] is set after the section loop, which can
	// append resolved-data strips — see the plan.ResolvedData block below.)
	if len(stored) == 0 {
		logger.Info("rerender_page_sections: page has no stored components, nothing to re-render",
			zap.String("page", pageName))
		out["skipped"] = true
		out["section_count"] = 0
		return out, nil
	}

	// One component-schema load for all sections (loadComponentSchemas keys by
	// both name and function — slot_name matches either). Loaded before the
	// pre-check so the check can consult each section's required-field contract.
	names := make([]string, 0, len(stored))
	for _, s := range stored {
		names = append(names, s.slotName)
	}
	schemas := loadComponentSchemas(ctx, params.DB, names, logger)

	// bugs_open/182: slot_name is only ever a naming convention. On a site
	// whose slots are positional ("prose-0", "tool-2") it is not, and never
	// will be, any component's name/function — so `schemas` above can never
	// resolve those sections, every one of them is silently carried, and a
	// re-render in which nothing was re-rendered completes successfully.
	// page_components.component_id is the row's own identity and does not
	// depend on slot naming at all; resolve by it FIRST, falling back to the
	// name/function map for sites where slot_name legitimately IS a component
	// identity.
	componentIDs := make([]string, 0, len(stored))
	seenComponentIDs := make(map[string]bool, len(stored))
	for _, s := range stored {
		if s.componentID != "" && !seenComponentIDs[s.componentID] {
			seenComponentIDs[s.componentID] = true
			componentIDs = append(componentIDs, s.componentID)
		}
	}
	byID, byIDDrops := loadComponentSchemasByID(ctx, params.DB, componentIDs, logger)

	// resolveComponent is the single lookup every branch below now goes
	// through. invalidTemplate distinguishes "a row exists but the template
	// guard rejected it" (bugs_open/024's second route into an empty schemas
	// map) from "nothing resolved at all" — the two get different fatal-list
	// entries in the diagnostic below because their remediations differ.
	resolveComponent := func(s storedSection) (comp componentInfo, invalidTemplate bool, haveComp bool) {
		if s.componentID != "" {
			if ci, ok := byID[s.componentID]; ok {
				// Observe-only (mirrors the CTA ownership-conflict log below):
				// when BOTH routes resolve and disagree, the id now wins where
				// the name map used to. Measured 2026-08-03: 13 live sections
				// fleet-wide hit this, each to a substantively different
				// template — logged so that flip stays visible, not silently
				// exchanged for the id's answer.
				if nameComp, nameOK := schemas[s.slotName]; nameOK && nameComp.ID != ci.ID {
					logger.Info("rerender_page_sections: component_id and slot_name resolve to different components (observe-only, id wins)",
						zap.String("section", s.slotName),
						zap.String("id_resolved_component", ci.ID),
						zap.String("id_resolved_name", ci.Name),
						zap.String("name_resolved_component", nameComp.ID),
						zap.String("name_resolved_name", nameComp.Name))
				}
				return ci, false, true
			}
			if _, dropped := byIDDrops[s.componentID]; dropped {
				// Deliberately NOT falling through to the name map here. The
				// page names this exact component as its own; a template guard
				// rejecting THAT row is a defect in the row the page is pinned
				// to, and silently rendering a different, coincidentally
				// name-matched component instead would be the same silent
				// substitution 182 is about, just moved one level down.
				return componentInfo{}, true, false
			}
		}
		if ci, ok := schemas[s.slotName]; ok {
			return ci, false, true
		}
		return componentInfo{}, false, false
	}

	// ── Content pre-check: re-render-all renders each section from its STORED
	//    content_data, so every section must actually carry its content. Two
	//    ways it can fail, both of which would render an empty section and
	//    OVERWRITE good HTML with a blank shell (the exact defect that silently
	//    blanked live article bodies):
	//      (a) content_data is entirely absent (older pages predating capture);
	//      (b) content_data is present but MISSING a schema-required source:"llm"
	//          field — e.g. the stored {type,result} envelope that was never
	//          unwrapped, so it has no `content` key the article-body template
	//          needs.
	//    In either case escalate the WHOLE page to the writer (regenerate +
	//    backfill) and do NOT re-render here, leaving the existing HTML intact. ─
	for _, s := range stored {
		// A self-contained TOOL section legitimately has no content_data: a tool
		// is complete HTML with no LLM-authored fields, so content_data={} is
		// its correct shape, not the missing-content defect this pre-check
		// exists to catch. Escalating it bypasses save_sections — the ONLY
		// writer of rendered_html — so the re-render is computed and thrown
		// away, and a durable template fix never reaches the page
		// (bugs_open/024).
		//
		// Keyed on the EXPLICIT component_level='tool' marker plus an empty
		// input_schema, never on a heuristic about field shape: predicating on
		// "has no required LLM fields" would also exempt components declaring
		// OPTIONAL source:"llm" fields, a broader class than is justified.
		if comp, _, ok := resolveComponent(s); ok && isSelfContainedSection(comp) {
			logger.Info("rerender_page_sections: self-contained tool section, no content_data expected — rendering from template",
				zap.String("page", pageName),
				zap.String("section", s.slotName))
			continue
		}

		reason := ""
		if len(s.contentData) == 0 {
			reason = "no stored content_data"
		} else if comp, _, ok := resolveComponent(s); ok && len(comp.InputSchema) > 0 {
			// The rerender path reaches the gate WITHOUT a plan_sections pass, so
			// this is where a re-rendered legacy-dialect component would otherwise
			// be enforced silently — fire the fail-loud tripwire here.
			datahelpers.WarnIfLegacyDialect(comp.InputSchema, logger, "render-gate", comp.Function)
			if missing := missingRequiredLLMFields(comp.InputSchema, s.contentData); len(missing) > 0 {
				reason = fmt.Sprintf("stored content_data missing required field(s) %v", missing)
			}
		}
		if reason != "" {
			logger.Warn("rerender_page_sections: section content incomplete — escalating page to writer instead of blanking it",
				zap.String("page", pageName),
				zap.String("section", s.slotName),
				zap.String("reason", reason))
			disposition, err := escalateRerenderToWriter(ctx, params.DB, siteID, pageName,
				"a section had no stored content_data", logger)
			if err != nil {
				return nil, fmt.Errorf("escalate to writer: %w", err)
			}
			// Either way this page is not re-rendered here — rendering a section
			// with no content would blank it. Both facts are named in the output:
			// `escalated` says an item was raised, `skipped` says nothing was
			// re-rendered, and `escalation` says which of the two happened
			// (bugs_open/187 added the skip; bugs_open/182 is why it is named).
			out["escalation"] = disposition
			out["escalated"] = disposition == "raised"
			out["skipped"] = disposition != "raised"
			out["section_count"] = len(stored)
			return out, nil
		}
	}

	// ── Re-resolve + re-render each section (no LLM) ────────────────────────
	//
	// The stored slots, in position order, are the list this path iterates and
	// therefore the list per-section imagery binding counts occurrences over.
	// It binds only where this agrees with the plan's own order; a page whose
	// stored composition has drifted from its plan keeps page-wide resolution,
	// which is what this path did before binding existed.
	storedSlotNames := make([]string, 0, len(stored))
	for _, ss := range stored {
		storedSlotNames = append(storedSlotNames, ss.slotName)
	}
	resolver := newSourceResolver(siteID, params.DB, logger, pageName).
		withLiveSectionNames(storedSlotNames)

	// Minimal render-context base from sites.content_data (company/contact/etc).
	// Section templates take colours from CSS vars and copy from content_data,
	// so this base is small; it only matters for a section that reads an ambient
	// field. Built once and re-merged per section.
	baseData := buildRerenderBaseData(ctx, params.DB, siteID, domain, pageName, logger)

	// features_open/035 P1: the per-section pass is a function now — MOVED, not
	// rewritten — so a composition child can take the identical path a section does.
	outcome := rerenderFlatSections(ctx, params, stored, resolveComponent, resolver,
		baseData, siteID, pageID, pageName, domain, pageURL, reason, strippedMarkdownFields, logger)
	sectionsMetadata := outcome.sectionsMetadata
	reRendered := outcome.reRendered
	carried := outcome.carried
	resolution := outcome.resolution
	strippedMarkdownFields = outcome.strippedMarkdownFields

	// Instance-scope check over the page as assembled from THIS run's sections.
	// Recorded unconditionally, refused only when armed — see enforceInstanceScope
	// for why the default cannot be "refuse" (pages already collide today, and
	// arming by default would turn a latent defect into an outage on them).
	//
	// Concatenating the sections is the right unit even though it is not the
	// final page: an id collision is a property of what shares a document, and
	// every section here does. Chrome is excluded, so a collision BETWEEN a
	// section and the header/footer is not visible to this check — stated
	// because a clean result here is not a clean page.
	{
		var assembled strings.Builder
		for _, entry := range sectionsMetadata {
			if h, ok := entry["rendered_html"].(string); ok {
				assembled.WriteString(h)
				assembled.WriteString("\n")
			}
		}
		if collisions := DetectInstanceCollisions(assembled.String()); !collisions.Clean() {
			out["instance_collisions"] = collisions.Summary()
			out["instance_collision_ids"] = collisions.DuplicateElementIDs
			logger.Warn("rerender_page_sections: page is not safe to carry repeated components",
				zap.String("page", pageName),
				zap.String("detail", collisions.Summary()),
				zap.Bool("armed", enforceInstanceScope(params.StepConfig.Config)))
			if enforceInstanceScope(params.StepConfig.Config) {
				return nil, fmt.Errorf("instance scope: %s (page %s)",
					collisions.Summary(), pageName)
			}
		}
	}

	out["sections_metadata"] = sectionsMetadata
	out["section_count"] = len(sectionsMetadata)
	out["rerendered"] = reRendered
	out["carried"] = carried
	if len(strippedMarkdownFields) > 0 {
		// Durable record of every strip this run performed — stored AND
		// resolved surfaces (bugs_open/184; council 060bcc0a r2).
		out["stripped_markdown_fields"] = strippedMarkdownFields
	}
	// Per-reason breakdown for the legitimate (non-fatal) carries, so an
	// operator or a discovery check can tell "the run did nothing because
	// every section is a deliberate fallback" from the fatal case below
	// without re-deriving it from the logs.
	if len(resolution.NotReadySlots) > 0 {
		out["carried_not_ready"] = resolution.NotReadySlots
	}
	if len(resolution.EmptyTemplateSlots) > 0 {
		out["carried_empty_template"] = resolution.EmptyTemplateSlots
	}
	if len(resolution.RenderFailedSlots) > 0 {
		out["carried_render_failed"] = resolution.RenderFailedSlots

		// ESCALATE, don't just name it (council a44d9eb8 round 1 — three seats
		// made this point: bug_historian, guardian and render_guardian all read
		// the first version as "reports complete but degraded", which is the
		// shape bugs_closed/028 and /040 were about, and render_guardian named
		// the asymmetry exactly: a mistyped field that breaks the render is the
		// same class as a MISSING required field, which this action already
		// escalates to the writer in its pre-check. So it takes the same route,
		// through the same helper, keyed the same way.
		//
		// Two differences from the pre-check, both deliberate. It does NOT
		// return early: the pre-check fires before anything is rendered, while
		// here the other sections have re-rendered correctly and their bytes
		// are worth saving — a failed section keeps its stored HTML either way.
		// And a FAILED escalation does not fail the action: a record that gates
		// the repair it describes is worse than a record that is missing, which
		// is the same reasoning emitSectionDeadControlItem states for its own
		// write.
		//
		// This is not new authority over content that works today: it fires
		// only where a render ACTUALLY failed, a state that could not exist
		// before this change because the fallback swallowed it.
		disposition, escErr := escalateRerenderToWriter(ctx, params.DB, siteID, pageName,
			"a section's template could not execute against its stored content (bugs_open/260)", logger)
		if escErr != nil {
			logger.Error("rerender_page_sections: could not escalate a render failure to the writer — the carried sections still save",
				zap.String("page", pageName),
				zap.Strings("sections", resolution.RenderFailedSlots),
				zap.Error(escErr))
		}
		out["escalation"] = disposition
		out["escalated"] = disposition == "raised"
	}

	logger.Info("rerender_page_sections: done",
		zap.String("page", pageName),
		zap.String("reason", reason),
		zap.Int("sections", len(sectionsMetadata)),
		zap.Int("rerendered", reRendered),
		zap.Int("carried", carried))

	// bugs_open/182: a section carried because its component could not be
	// resolved at all (name AND id both missed, or the id's row failed the
	// template guard) is not a degraded render — the action failed to do the
	// one thing it exists to do, and the run must not report success. The
	// predicate is "any slot in these two lists", NOT rerendered==0: five of
	// the six sites this was measured against are PARTIAL carries, so an
	// all-or-nothing test would clear exactly the cases hardest to spot by eye.
	//
	// The two legitimate-carry lists (not-ready, empty-template) are
	// deliberately excluded — bugs_closed/095's second application
	// (section_editor_actions.go) had its first, blanket version of this
	// exact check rejected by the council as "asserted rather than evidenced";
	// this mirrors the corrected, narrower predicate.
	if resolution.fatal() {
		return nil, fmt.Errorf(
			"page %q: %d of %d section(s) could not resolve a component and were carried unrendered instead — %s (bugs_open/182)",
			pageName, len(resolution.UnresolvedSlots)+len(resolution.InvalidTemplateSlots), len(stored), resolution.describe())
	}

	// Every carry was a legitimate, individually-logged fallback — not fatal,
	// but worth one line an operator or discovery check can key on without
	// wading through per-section logs (weakest of the bug file's fix
	// candidates on its own; here it is additive to the two above, not a
	// substitute).
	if reRendered == 0 && carried > 0 {
		logger.Warn("rerender_page_sections: every section was carried (all legitimate fallbacks) — nothing changed on this run",
			zap.String("page", pageName),
			zap.Int("carried", carried))
	}

	return out, nil
}

// rerenderResolution accumulates, across a page's sections, the reasons a
// section did not re-render — named, not merely counted, mirroring
// bugs_closed/095's pageAssembly one layer up. UnresolvedSlots and
// InvalidTemplateSlots are FATAL: bugs_open/182 is exactly that nothing else
// distinguished "the re-render did nothing" from "the re-render worked" once
// every section silently took one of these two routes. NotReadySlots and
// EmptyTemplateSlots are legitimate, evidenced fallbacks and stay non-fatal.
type rerenderResolution struct {
	UnresolvedSlots      []string
	InvalidTemplateSlots []string
	NotReadySlots        []string
	EmptyTemplateSlots   []string
	// RenderFailedSlots names sections whose component template FAILED TO
	// EXECUTE (bugs_open/260) and were therefore carried with their stored
	// HTML. Non-fatal on purpose — see the carry at the render call — but it is
	// a real defect in the content, not a legitimate fallback like the two
	// above, so it is reported separately rather than folded into either.
	RenderFailedSlots []string
	// DeadURLSlots names sections that RENDERED, but with a URL attribute left
	// empty (bugs_open/238). Deliberately NOT part of fatal(): this path merges
	// stored ⊕ fresh and so cannot lose a key — it can only re-ship damage that
	// is already live, and it is the vehicle by which a repaired row reaches the
	// artefact. Named in the output because a re-render that quietly re-shipped
	// five empty <img src=""> is exactly the "reported complete, changed
	// nothing" reading bugs_open/182 exists because of.
	DeadURLSlots []string
}

// fatal reports whether this page's re-render must fail the step rather than
// complete. Deliberately NOT keyed on rerendered==0 — see the call site.
func (r rerenderResolution) fatal() bool {
	return len(r.UnresolvedSlots) > 0 || len(r.InvalidTemplateSlots) > 0
}

// describe names every slot in every list, so the operator reading the error
// (or the immune-system failure sweep) sees exactly which sections and why,
// the same shape as pageAssembly.describe() one layer up.
func (r rerenderResolution) describe() string {
	return fmt.Sprintf("unresolved component %v; invalid template %v; not ready (legitimate) %v; empty template (legitimate) %v; dead URL controls (recorded, non-fatal) %v; render failed, stored HTML carried (bugs_open/260) %v",
		r.UnresolvedSlots, r.InvalidTemplateSlots, r.NotReadySlots, r.EmptyTemplateSlots, r.DeadURLSlots, r.RenderFailedSlots)
}

// slotLabel names a slot together with its position, so a duplicate slot_name
// on one page (11 legitimate cases fleet-wide) still names a SPECIFIC section
// rather than an ambiguous repeated string.
func slotLabel(s storedSection) string {
	return fmt.Sprintf("%s (pos %d)", s.slotName, s.position)
}

// rerenderCTAState holds the per-page CTA targets for a cta_links_stale
// recompute — computed once, shared by every CTA section on the page.
type rerenderCTAState struct {
	primary, secondary contentHub
	validPages         datahelpers.PageURLSet
	candidates         []datahelpers.LabelMatchCandidate
	pageName, pageURL  string // this page's own identity — never its own CTA target
}

// loadRerenderCTAState reuses the internal-link-resolver's loaders and ranking
// (interactive pages first, then hubs). Loader failures degrade to empty
// candidate lists: applyCTARecompute then leaves fields untouched rather than
// aborting the rerender.
func loadRerenderCTAState(ctx context.Context, params ActionParams, siteID uuid.UUID, pageName, pageURL string, logger *zap.Logger) *rerenderCTAState {
	hubs, err := loadContentHubs(ctx, params, siteID, logger)
	if err != nil {
		logger.Warn("rerender_page_sections: loadContentHubs failed for CTA recompute", zap.Error(err))
	}
	interactive, err := loadInteractivePages(ctx, params, siteID, logger)
	if err != nil {
		logger.Warn("rerender_page_sections: loadInteractivePages failed for CTA recompute", zap.Error(err))
	}
	primary, secondary := chooseCTATargets("", pageName, interactive, hubs)
	// LABEL-MATCH supply is the SHARED universe (bugs_open/308 Phase B): the
	// misdirected_cta check that FILES this rerender and this recompute must
	// answer "which pages may this label name?" identically, or the repair
	// cannot reach the destination the check suggested — which is the whole of
	// bug 308 (99 findings sat on work items marked `complete`, 2026-08-23).
	// The POSITIONAL pick above still reads the hub/tool loaders and still
	// refuses every utility area.
	//
	// A failure degrades to an empty universe, matching the two loaders above:
	// applyCTARecompute then leaves fields untouched rather than aborting a
	// rerender that has other work to do. That asymmetry with the build path
	// (which returns the error) is deliberate — a rerender is a repair pass over
	// an already-serving page, and half-repairing it beats failing it.
	candidates, err := datahelpers.LoadCTALabelUniverse(ctx, params.DB, siteID)
	if err != nil {
		logger.Warn("rerender_page_sections: CTA label universe failed for CTA recompute", zap.Error(err))
	}
	return &rerenderCTAState{
		pageName:   pageName,
		pageURL:    pageURL,
		primary:    primary,
		secondary:  secondary,
		validPages: loadResolverPageSet(ctx, params, siteID, logger),
		candidates: candidates,
	}
}

// applyCTARecompute writes the recomputed CTA target into resolved (which the
// caller merges LAST, beating stale stored content_data).
//
// bugs_open/203 follow-on: this is the check_misdirected_cta.go repair path
// (triggered by reason=cta_links_stale, which that check itself sets on the
// work item it files) — so it must be able to fix exactly the defect that
// triggered it: a label naming one real page while the href points at a
// different one. Before this change it could not — "authored link to a real,
// sensible destination, keep it" accepted ANY valid, non-excluded, non-self
// URL, including a misdirected one, so a detected mismatch on a stored (as
// opposed to phantom) target was silently kept forever and re-detected every
// pass. Now: if the slot's existing label names a real candidate whose URL
// disagrees with what's currently stored, the label-matched candidate wins —
// this is the actual repair. Only when the label matches nothing (generic
// text, or no candidate at all) does the function fall back to today's
// keep-if-valid-else-positional-target behaviour, unchanged.
//
// bugs_open/248 (cta_recompute_clobbers_authored_contact_links): the keep
// branch no longer refuses a utility-area destination. It used to, which meant
// an authored /contact.html could take NO branch — generic contact copy
// ("Get in Touch" reduces to [touch]) names no page, so the label match
// declines too — and fell through to the positional pick that overwrote it.
// The label match stays AHEAD of the keep on purpose: a FABRICATED contact url
// whose label names a real page is 203's defect and must still be repaired.
func applyCTARecompute(resolved, stored map[string]interface{}, field string, target contentHub,
	validPages datahelpers.PageURLSet, pageURL string, existingLabel string, candidates []datahelpers.LabelMatchCandidate,
	pageName string) {

	if field == "" {
		return // single-URL component — no field in this slot
	}

	// Carry the section's existing mint record forward before deciding anything
	// (bugs_open/308 Phase A) — the persist below merges SHALLOWLY and the record
	// is a nested map, so stamping only this field would replace the stored
	// record and drop the sibling slot's stamp, freezing it. Inside the function,
	// not in the caller's loop: a loop-level call can be deleted with every test
	// still green (proven by mutation). SeedCTAMinted fills only absent entries,
	// so once per field is idempotent.
	datahelpers.SeedCTAMinted(resolved, stored)

	current, hasCurrent := stored[field].(string)

	if existingLabel != "" {
		// Two ways this declines, both meaning "no opinion, keep what is there":
		// AMBIGUOUS copy (two pages tied on everything but alphabetical order —
		// 137 of the 428 writes the widening would otherwise perform), and copy
		// that names THE PAGE IT SITS ON (35 of 291 after the refusal — the
		// 2026-08-23 hand audit). A repair that rewrites a live button to a page
		// chosen by the alphabet, or to the page the reader is already on, is
		// not a repair.
		if match, ok, _ := datahelpers.BestLabelMatchForPage(existingLabel, candidates, pageName, pageURL); ok && validPages.Contains(match.URL) &&
			(!hasCurrent || datahelpers.NormalizePagePath(match.URL) != datahelpers.NormalizePagePath(current)) {
			resolved[field] = match.URL
			datahelpers.SetCTAMinted(resolved, field, match.URL)
			if match.Title != "" {
				resolved[ctaTargetTitleField(field)] = match.Title
			}
			return
		}
	}

	// KEEP #1 — an AUTHORED utility-area destination (bugs_open/248, slug
	// cta_recompute_clobbers_authored_contact_links). A stored, VALID
	// /contact.html cannot have been produced by this resolver (see
	// storedCTADestinationIsAuthored), so it was authored. Refusing it here was
	// the clobber: generic contact copy names no page, so the label match
	// declines too, and control fell through to the positional pick — which
	// replaced a working contact button with an unrelated tool page.
	//
	// This one is WRITTEN rather than merely left alone. Keep #2 below can
	// safely return bare because resolved merges OVER the stored content_data,
	// so an untouched field flows through unchanged. That is true today for
	// this branch too — but the field's schema fallback is the one thing that
	// could beat it, and a utility-area fallback is exactly the shape someone
	// adds back by accident (site-header still carries one). Writing makes the
	// keep independent of anything else that populated ResolvedData.
	//
	// bugs_open/308 Phase A: the predicate now reads a RECORD instead of
	// inferring from the resolver's constraints. Unchanged in behaviour until
	// the candidate set widens — and, as in setCTAField, this branch writes no
	// mint stamp on purpose: reaching it proves the stamp does not cover
	// `current`, which is exactly what "a person authored this" means here.
	if hasCurrent && storedCTADestinationIsAuthored(stored, field, validPages) &&
		datahelpers.NormalizePagePath(current) != datahelpers.NormalizePagePath(pageURL) {
		resolved[field] = current
		if title, _ := stored[ctaTargetTitleField(field)].(string); title != "" {
			resolved[ctaTargetTitleField(field)] = title
		}
		return
	}

	// KEEP #2 — any ordinary valid destination. Kept by leaving the field
	// untouched, so the stored value flows through the merge (unchanged since
	// 203).
	//
	// ⚠ THE UTILITY-AREA EXCLUSION WAS REMOVED HERE (bugs_open/308 Phase B) and
	// removing it is REQUIRED BY the widening, not incidental to it. Until Phase
	// B the resolver could not mint a utility destination, so a stored one was
	// authored, KEEP #1 caught it, and this branch never saw one. Now the label
	// match CAN write /contact.html and stamp it — and KEEP #1's predicate
	// deliberately refuses a stamped value (that is what the stamp MEANS). With
	// the old exclusion still here, a minted /contact.html whose copy later went
	// generic ("Get Started" names no page, so the label match declines) took no
	// keep at all and fell to the POSITIONAL PICK, which replaced a working
	// contact button with an unrelated tool page. That is bugs_open/248's exact
	// damage, re-created by 308's own fix.
	//
	// The invariant that replaces it, and it is the symmetric half of the one
	// rank() enforces: THE POSITIONAL PICK MAY NEITHER CHOOSE NOR DISPLACE A
	// UTILITY DESTINATION. Only a confident label match moves one.
	//
	// ⚠ `validPages.Contains` STAYS — it is what makes KEEP #3 reachable, since
	// every non-page href (tel:/mailto:/external) fails it by construction. The
	// LANDMINES entry warns against "tidying" this branch by dropping that test;
	// this edit drops the AREA test, which is a different condition.
	if hasCurrent && current != "" &&
		validPages.Contains(current) &&
		datahelpers.NormalizePagePath(current) != datahelpers.NormalizePagePath(pageURL) {
		return // a real, sensible destination — keep it
	}

	// KEEP #3 — an AUTHORED NON-PAGE destination (bugs_open/299): tel:/
	// mailto:/external/named-fragment. Fails validPages.Contains by
	// construction, so keeps #1 and #2 can never see one and it fell to the
	// positional pick — this is the LANDMINES.md "cta_links_stale repair
	// CANNOT tell a genuine 'Get in Touch' from …" trap in its second form:
	// webdesign.uk's faq and how-it-works carry genuine "Call us on …" tel:
	// buttons that a rerender would have replaced with a tool link. DISJOINT
	// from keep #1 (which REQUIRES validPages membership).
	//
	// WRITTEN, for keep #1's schema-fallback reason — and the write is the
	// URI repair: NormalizeTelHref fixes the unambiguous malformed tel:
	// forms (spaces/parens) on the next rerender; the collapsed-trunk
	// "+440…" it refuses stays RAW for a human (check_cta_nonpage files it).
	// The title is COMPUTED so the content writer can write copy FOR the
	// destination instead of inventing one (bugs_open/299's mechanism).
	//
	// ⚠ THE ORDER IS LOAD-BEARING: this branch is reachable only because
	// keep #2's predicate requires validPages.Contains, which every non-page
	// url fails. "Tidying" keep #2 into a broader keep (dropping the
	// validPages test) would silently swallow these cases — kept, but never
	// normalised and never titled. The tripwire is the WRITE expectation in
	// resolve_internal_links_nonpage_test.go's applyCTARecompute cases: a
	// swallowed tel: leaves resolved empty and they fail. If one fails after
	// an edit to keep #2, the fix is keep #2, not the test. Sibling seam:
	// LNK-033's asymmetry pin (TestFreshPickRefusesUtilityWhileStoredUtilityIsKept).
	if hasCurrent && datahelpers.IsAuthoredNonPageCTADestination(current) {
		kept := current
		if norm, ok := datahelpers.NormalizeTelHref(current); ok {
			kept = norm
		}
		resolved[field] = kept
		if title := datahelpers.DescribeCTADestination(current); title != "" {
			resolved[ctaTargetTitleField(field)] = title
		}
		return
	}

	if target.URL == "" || !validPages.Contains(target.URL) {
		// Nothing valid to write — leave the field as stored. No stamp carry is
		// needed on this path (unlike setCTAField's sibling fallthrough): the
		// rerender MERGES resolved over the stored content_data rather than
		// replacing it, so an untouched field keeps its stored value AND its
		// stored stamp together.
		return
	}
	resolved[field] = target.URL
	datahelpers.SetCTAMinted(resolved, field, target.URL)
	if title := targetTitle(target); title != "" {
		resolved[ctaTargetTitleField(field)] = title
	}
}

// loadStoredSections reads the page's current page_components rows in order.
//
// build_status='removed' is EXCLUDED, and that exclusion is load-bearing: this
// path re-renders from stored content_data and hands the result to
// save_page_sections, which replaces the page's rows wholesale. Without the
// filter a section that was deliberately removed comes back — the row is still
// present (removal marks status and empties rendered_html; it does not clear
// content_data, which is exactly what this path renders from), so it is
// re-rendered and re-saved as 'deployed', and the removal is undone with a
// successful-looking rerender. That happened on idea.uk/index on 2026-08-10:
// a section the owner had had removed returned via an unrelated
// section_data_resolved rerender, and the only reason it was noticed is that a
// decision guard (RFC_015) asserted its absence and filed
// decision_regression:...:D-002. The current plan (site_plan_sections for the
// is_current plan) and pages.sections both still excluded it — page_components
// was the only store that disagreed.
//
// 'removed' is NOT a local convention: v3_site_actions.go:4366 and
// internal/core-manager/admin/page_admin_handlers.go:59 already filter it out.
// This reader was the one out of step.
//
// IS DISTINCT FROM, not !=, deliberately: build_status is NULLABLE (default
// 'pending', no NOT NULL constraint). There are no NULLs today, so both forms
// behave identically now — but != would silently drop every NULL-status row
// from every rerender fleet-wide the moment one appeared, which is a far worse
// failure than the one being fixed here. The two sibling readers above carry
// that latent flaw; this one does not.
func loadStoredSections(ctx context.Context, db *sql.DB, pageID uuid.UUID, logger *zap.Logger) ([]storedSection, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id::text,
		       COALESCE(parent_instance_id::text, ''),
		       COALESCE(component_id::text, ''),
		       COALESCE(slot_name, ''),
		       content_data,
		       COALESCE(rendered_html, ''),
		       position,
		       COALESCE(component_version_id::text, '')
		FROM page_components
		WHERE page_id = $1
		  AND `+datahelpers.NotRemoved("")+`
		-- (position, id), not position alone: this walk is what assigns
		-- per-instance element-id occurrences (InstanceCounter), and the
		-- single-section paths now count predecessors with the same
		-- (position, id) comparison (component_instance_occurrence.go). A tie
		-- ordered arbitrarily by Postgres would let the two derivations
		-- disagree on the same page. 2 live ties fleet-wide as of 2026-08-24,
		-- both cross-function, so no instance token changes value today.
		ORDER BY position ASC, id ASC
	`, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []storedSection
	// offered counts rows the CURSOR yielded, against len(out) kept
	// (bugs_open/410's pinned count). Not "rows in page_components for this
	// page": the WHERE above legitimately drops 'removed' tombstones in SQL, so
	// that comparison would fire on every page carrying a removed component — a
	// large, healthy population — and a guard that fires constantly on correct
	// input gets loosened within a week.
	offered := 0
	for rows.Next() {
		offered++
		var s storedSection
		var cdJSON []byte
		if err := rows.Scan(&s.id, &s.parentInstanceID, &s.componentID, &s.slotName, &cdJSON, &s.renderedHTML, &s.position, &s.componentVersionID); err != nil {
			// scan-loss:accepted: counted — ScanShortfall below refuses the
			// partial result. ⚠ This continue is safe ONLY while that trailing
			// check survives: delete the ScanShortfall return and this branch
			// reverts to the exact defect it was (renders less, reports
			// complete — bugs_open/410). The mutation check in
			// rerender_page_sections_scan_completeness_test.go goes red on
			// that deletion; if it doesn't, the guard is not what is producing
			// the refusal. (bug_historian advisory, round c8385154.)
			// The per-row Warn is kept deliberately: on a mixed
			// failure it records EVERY failing row's cause, where returning here
			// would record only the first, and the first is rarely the
			// informative one when a projection has drifted.
			logger.Warn("rerender_page_sections: row scan failed", zap.Error(err))
			continue
		}
		if len(cdJSON) > 0 {
			// Closed residual (bugs_open/410): this was `_ = json.Unmarshal(...)`,
			// which KEPT the row and EMPTIED its content on a parse failure —
			// offered == kept, invisible to the count guard below, and
			// save_page_sections would then replace the stored row wholesale
			// with the emptied one. Same destruction argument as the scan
			// branch above, same mechanism: drop the row so ScanShortfall
			// refuses the whole load. The column is jsonb, so the only
			// reachable failure is a non-object value — 0 of 2,751 rows
			// fleet-wide as of 2026-08-31, so this refusal fires on no page
			// today and exists for the first writer that changes that.
			// Deliberately NOT stricter than the parse: SQL NULL (55 loadable
			// rows as of 2026-08-31) never enters this branch, and jsonb
			// `null` unmarshals to the same nil map without error. Both stay
			// loadable — a nil-content section is a live, legitimate
			// population; only content that FAILS to parse is refused.
			if err := json.Unmarshal(cdJSON, &s.contentData); err != nil {
				logger.Warn("rerender_page_sections: content_data does not parse into a section object; dropping the row so the shortfall guard refuses",
					zap.String("page_component_id", s.id), zap.Error(err))
				continue
			}
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// ANY loss is an error here — deliberately STRICTER than scanBlogArticles'
	// graded policy in rebuild_blog_listing_action.go, and the asymmetry is the
	// point. That reader feeds a PROJECTION: a listing aggregates posts that
	// still exist, so one malformed row degrades a nav surface and must not
	// blank a live listing. This reader feeds a WHOLESALE REPLACE —
	// save_page_sections deletes the page's rows and rewrites them
	// (save_page_sections_action.go's `DELETE FROM page_components WHERE
	// page_id = $1`), so a section missing from THIS slice is not merely
	// unrendered, it is DELETED, and the page ships with a hole under a fresh
	// deploy stamp with the work item reported complete. Degradation is not on
	// the menu: the choice is a loud failure or a silent destruction.
	//
	// The refusal lands before anything is written (this loader runs at the top
	// of the action, ahead of strip/render/save), so a page keeps serving its
	// last good render while the step fails and retries.
	return out, datahelpers.ScanShortfall(offered, len(out), "rerender_page_sections: stored page_components")
}

// carryStoredSection builds a sections_metadata entry that re-emits a section's
// stored render unchanged (used when a section can't be safely re-rendered).
// Shape matches CompilePageSectionsAction so save_page_sections ingests it.
func carryStoredSection(s storedSection) map[string]interface{} {
	m := map[string]interface{}{
		"rendered_html":  s.renderedHTML,
		"component_name": s.slotName,
		// The stored row's own slot identity, stated explicitly so the save
		// preserves it verbatim rather than deriving a name (bugs_open/189). A
		// no-op on this path today — a carry sets no component_function, so the
		// save already fell through to component_name — but stating it keeps the
		// two producers in this file saying the same thing, which is what stops
		// the next edit re-opening the gap.
		"stored_slot_name": s.slotName,
	}
	if s.componentID != "" {
		m["component_id"] = s.componentID
	}
	// The stamp travels with the bytes it describes (RFC_046). Carrying the HTML
	// while dropping its provenance would make every rerender quietly downgrade a
	// known row to an unknown one.
	if s.componentVersionID != "" {
		m["component_version_id"] = s.componentVersionID
	}
	if len(s.contentData) > 0 {
		m["content_data"] = s.contentData
	}
	return m
}

// buildRerenderBaseData assembles the minimal ambient render-context data from
// sites.content_data (top-level + reviewed_brief), plus domain and year. The
// section templates examined take colours from CSS vars and copy from
// content_data, so this base is rarely read — it only covers a section that
// references an ambient field (company name, contact, etc).
func buildRerenderBaseData(ctx context.Context, db *sql.DB, siteID uuid.UUID, domain string, pageName string, logger *zap.Logger) map[string]interface{} {
	base := map[string]interface{}{
		"domain": domain,
		"year":   fmt.Sprintf("%d", time.Now().Year()),
	}

	// The page's own identity. bugs_open/085 was fixed on the page-BUILD path
	// (BuildRenderContextAction); this is the same defect on the scoped
	// section-re-render path, and it was found by firing a re-render on the
	// fixed binary and watching a page still render three charts assigned to
	// two different pages. The page name was already in scope here — it is
	// passed to newSourceResolver on the line above the call — so the identity
	// was available and simply never reached the render base.
	//
	// mergeIntoRenderContext restores this into RenderContext.CurrentPage, so
	// setting the map key is all that is needed. Trimmed to match
	// buildHeaderConfig, which is the form every other producer uses.
	//
	// Under the STEP-BOUNDARY name (current_page_name), not the template name:
	// this map is the restore contract's input, and that contract renamed the
	// page-name string so it can never share a key with the page RECORD in a
	// collected_data tree (renderContextStepContractRenames, RFC_029 §10.13
	// step 4). Templates still see it as {{.current_page}} — the struct field
	// is projected under its tag on the render side.
	if pageName != "" {
		base[renderContextStepContractKey("current_page")] = strings.TrimSuffix(pageName, ".html")
	} else {
		logger.Warn("rerender_page_sections: no page name for the render base — every section will see an empty current_page and cannot vary per page (bugs_open/085)")
	}

	// Read content_data AND the canonical sites.email COLUMN together. The
	// full-writer render path (loadSiteDataFull) sources ctx.Email from the
	// sites.email column; this light path historically read only
	// content_data.email, which is empty or stale on most sites (idea.uk held a
	// stale idea-uk@leopardess.uk while its column carried the current address).
	// So a section re-render could not convert a dead contact form to a mailto
	// even where the site had a real address. We now prefer the column, applied
	// AFTER the content_data merge so it wins — making both render paths agree.
	// See bugs_open/006 §B.
	var cdJSON []byte
	var siteEmail string
	if err := db.QueryRowContext(ctx, `SELECT content_data, COALESCE(email, '') FROM sites WHERE id = $1`, siteID).Scan(&cdJSON, &siteEmail); err != nil {
		if err != sql.ErrNoRows {
			logger.Warn("rerender_page_sections: load sites row failed", zap.Error(err))
		}
	} else if len(cdJSON) > 0 {
		var cd map[string]interface{}
		if err := json.Unmarshal(cdJSON, &cd); err != nil {
			logger.Warn("rerender_page_sections: parse sites.content_data failed", zap.Error(err))
		} else {
			// reviewed_brief first so its keys are present, then top-level wins on overlap.
			if rb, ok := cd["reviewed_brief"].(map[string]interface{}); ok {
				for k, v := range rb {
					if _, exists := base[k]; !exists {
						base[k] = v
					}
				}
			}
			for k, v := range cd {
				if k == "reviewed_brief" {
					continue
				}
				base[k] = v
			}
		}
	}

	// The sites.email column is canonical (matches loadSiteDataFull) — applied
	// last so it overrides any stale/empty content_data.email merged above.
	// The sites.email column is canonical (matches loadSiteDataFull) — applied
	// last so it overrides any stale/empty content_data.email merged above.
	if siteEmail != "" {
		base["email"] = siteEmail
		base["contact_email"] = siteEmail
	}

	return base
}

// isSelfContainedSection reports whether a section's component renders entirely
// from its own template, with no LLM-authored content_data to supply.
//
// Both signals are required and both are explicit:
//   - component_level == "tool", the marker set at component creation. It is
//     already SELECTed by loadSectionComponents (COALESCE(component_level,
//     'section')) and carried on componentInfo.Raw, so reading it here costs no
//     extra query and no struct change.
//   - an empty input_schema, i.e. the component declares no content fields at
//     all.
//
// Deliberately NOT a heuristic over field shape (e.g. "has no REQUIRED llm
// fields"), which would also exempt components declaring optional source:"llm"
// fields — a broader class than the evidence justifies. As of 2026-07-20 this
// matches 12 of 122 active components.
func isSelfContainedSection(comp componentInfo) bool {
	if len(comp.InputSchema) > 0 {
		return false
	}
	level, _ := comp.Raw["component_level"].(string)
	return level == "tool"
}

// escalateRerenderToWriter emits a needs_page work item so page-build-handler
// rebuilds the page through the writer (regenerate + backfill content_data).
// Keyed needs_page:<page> so it co-dedups with reconcile_site_plan's items.
//
// GUARDED (bugs_open/187). The trigger for this escalation is a section with no
// stored content_data, and that is not always a defect: a widget slot rendering
// from something other than content_data legitimately carries none, so on a
// page that declares no sections the escalation was a FALSE ALARM asking the
// writer to rebuild from a section plan that does not exist. Four such items
// parked in needs_human_review and nothing drained them. If the handler would
// resolve no sections and the page is in no current plan, no item is emitted.
//
// Returns the disposition so the caller can put it in the action's output —
// "raised", "skipped_sectionless_page", "skipped_owned_page", or
// "escalate_failed" alongside a non-nil error. A no-op that only the absence
// of a row records is a silent no-op (bugs_open/182).
// `cause` is the human half of the summary only. The spec's machine reason
// stays content_data_backfill for BOTH callers because the remedy is the same
// one — regenerate the page's content through the writer and backfill
// content_data — and the item_key is unchanged, so the two causes co-dedup on
// one page exactly as they should.
func escalateRerenderToWriter(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, cause string, logger *zap.Logger) (string, error) {
	satisfiable, _, sectionSource := pageSectionsSatisfiable(ctx, db, logger, siteID, pageName)
	if !satisfiable {
		logger.Info("rerender_page_sections: skipped_sectionless_page — the page resolves no sections and is in no current plan, so the writer has no plan to rebuild from; see bugs_open/187",
			zap.String("site_id", siteID.String()),
			zap.String("page", pageName),
			zap.String("sections_source", sectionSource))
		return "skipped_sectionless_page", nil
	}

	// GUARDED, second alarm (bugs_open/333, owner ruling 2026-08-25, remedy (i)).
	// An owned page's widget slots legitimately hold no content_data — that is
	// what rebuild_policy='owned' MEANS — so this escalation's trigger is its
	// normal false alarm on a tool page, not a defect: 13 needs_page items were
	// minted this way in the writeWorkItem door's first 14 hours, and every one
	// was refused wont_fix by page-build-handler's ownership guard. Those rows
	// also never met that door, because this emit carries no page_id and is born
	// 'triaged', which no promoter re-examines. So the false alarm stops HERE,
	// at source — and inside this helper rather than in a caller, because a
	// precondition parked in a caller is one port away from gone
	// (owned_page_guard.go's own lesson).
	//
	// Fail-open on every lookup problem, matching pageIsOwnedForGuard's posture:
	// a page that does not resolve yet is the legitimate build-request case, and
	// an unreadable policy must not suppress a real escalation.
	if guardPageID, _, lookupErr := saveSectionsLookupPageID(ctx, db, siteID, pageName); lookupErr == nil {
		if owned, checked := pageIsOwnedForGuard(ctx, db, guardPageID, logger); checked && owned {
			logger.Info("rerender_page_sections: skipped_owned_page — the page is rebuild_policy=owned, so the writer this would escalate to is forbidden to touch it; owned slots legitimately carry no content_data (bugs_open/333)",
				zap.String("site_id", siteID.String()),
				zap.String("page", pageName),
				zap.String("cause", cause))
			// The skip leaves a DURABLE record, not only a log line (council
			// round 70a1e557, bug_historian): if an owned page ever genuinely
			// lost its content, this escalation was the alarm that would have
			// fired, and a suppression only a pod log records is invisible to
			// any later audit. owned_page_review is the platform's established
			// per-page trail for exactly this refusal class — one row per page
			// (ON CONFLICT), same shape reconcile and the composition guards
			// already write, so it cannot flood and a reader finds every
			// owned-page refusal in one place. Errors are logged and swallowed
			// inside the emitter: the skip protects the page either way.
			emitOwnedPageReviewItem(ctx, db, siteID, pageName, "page-rerender",
				ownedPageSkipReasonPrefix+": rerender escalation skipped — "+cause+
					" is the normal state of an owned page's widget slots; if this page's tool genuinely lost content, rebuild via the tool pipeline",
				logger)
			return "skipped_owned_page", nil
		}
	} else if !errors.Is(lookupErr, sql.ErrNoRows) {
		logger.Warn("rerender_page_sections: ownership lookup failed — escalation guard standing down",
			zap.String("page", pageName), zap.Error(lookupErr))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "escalate_failed", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	batchID := uuid.New()
	spec := fmt.Sprintf(`{"reason":"content_data_backfill","page_name":%q}`, pageName)
	if _, err := insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "page-rerender",
		pipeline:     "build",
		itemType:     "needs_page",
		severity:     "medium",
		summary:      fmt.Sprintf("Full rebuild of %s — %s", pageName, cause),
		spec:         spec,
		priority:     90,
		handlerAgent: "page-build-handler",
		status:       "triaged",
		createdBy:    "page-rerender",
		itemKey:      fmt.Sprintf("needs_page:%s", pageName),
		batchID:      batchID,
	}, logger); err != nil {
		return "escalate_failed", fmt.Errorf("emit needs_page: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "escalate_failed", fmt.Errorf("commit: %w", err)
	}
	return "raised", nil
}

// rerenderSectionsOutcome is what one pass over a page's stored sections
// produces.
type rerenderSectionsOutcome struct {
	sectionsMetadata       []map[string]interface{}
	reRendered             int
	carried                int
	resolution             rerenderResolution
	strippedMarkdownFields []string
}

// rerenderFlatSections is the ORIGINAL per-section pass, MOVED and not rewritten.
//
// This is a pure extraction: every branch, every resolution bucket and every
// diagnostic is the code that was inline, and the existing suite is the proof.
// It is separated on its own, ahead of any composition branch, precisely so the
// council's four-seat objection to restructuring the fleet's busiest pipeline can
// be answered by a diff that shows a MOVE (correlation 53d71504: guardian HIGH,
// echoed by architecture, render_guardian and debug_historian). When the
// composition branch lands beside it, this function is what a flag falls back to,
// so it must stay the code the fleet runs today.
//
// resolveComponent arrives as a PARAMETER because it is a closure over the
// caller's component maps. Passing it keeps the moved body byte-identical rather
// than dragging its captures along — the alternative was rewriting the block,
// which is the one thing this extraction must not do.
func rerenderFlatSections(
	ctx context.Context,
	params ActionParams,
	stored []storedSection,
	resolveComponent func(storedSection) (componentInfo, bool, bool),
	resolver *sourceResolver,
	baseData map[string]interface{},
	siteID, pageID uuid.UUID,
	pageName, domain, pageURL, reason string,
	strippedMarkdownFields []string,
	logger *zap.Logger,
) *rerenderSectionsOutcome {
	sectionsMetadata := make([]map[string]interface{}, 0, len(stored))
	reRendered := 0
	carried := 0
	var resolution rerenderResolution

	// CTA targets for reason=cta_links_stale, loaded once on first CTA section.
	// Lives on the pass struct now, so a composition child reaches the SAME lazily
	// loaded targets rather than triggering a second load of its own.

	// One counter for the whole page, advanced in position order (loadStoredSections
	// orders by position) — the canonical per-instance token derivation.
	instances := NewInstanceCounter()
	// Seeded with what the pre-loop strip already recorded, and read back into the
	// outcome below — otherwise every resolved-data strip the render half performs
	// would be dropped on the floor, which is a DURABLE RECORD lost silently
	// (bugs_open/184: the strip record is the evidence the strip happened).
	pass := &rerenderPass{resolution: &resolution, instances: instances,
		strippedMarkdownFields: strippedMarkdownFields}
	// A SECOND counter, not a second RULE: `instances` is consumed as it walks
	// (Next returns a token and advances), so imagery binding cannot read its
	// state without stealing its position. Both call the same NextOccurrence on
	// the same list in the same order, which is what has to be true.
	sectionOccurrences := NewInstanceCounter()

	// Which occurrence of its own slot name each stored section is, in position
	// order — the section's identity for per-section imagery binding (see
	// sectionRef in plan_sections_action.go). It must be counted here as well
	// as on the build path: this path merges stored content_data with FRESHLY
	// resolved fields, resolved last, so a re-render that resolved every
	// section's figure page-wide would overwrite the per-section figures the
	// build had just got right. The two paths agree because both count
	// occurrences of a slot name in page order — never a position integer,
	// whose base differs between the two tables.

	for _, s := range stored {
		// Counted before any early `continue`: a section that carries its
		// stored HTML still occupies its place in the page, so skipping the
		// count here would shift every later section of the same name onto the
		// wrong figure — the quiet half of an off-by-one, on the path that
		// runs most often.
		thisSection := newSectionRef(s.slotName, sectionOccurrences.NextOccurrence(s.slotName))

		cls := classifyStoredSection(ctx, s, thisSection, resolveComponent, resolver, logger)
		if cls.carryKind != "" {
			switch cls.carryKind {
			case carryInvalidTemplate:
				resolution.InvalidTemplateSlots = append(resolution.InvalidTemplateSlots, slotLabel(s))
			case carryNotFound:
				resolution.UnresolvedSlots = append(resolution.UnresolvedSlots, slotLabel(s))
			case carryNotReady:
				resolution.NotReadySlots = append(resolution.NotReadySlots, slotLabel(s))
			case carryEmptyTemplate:
				resolution.EmptyTemplateSlots = append(resolution.EmptyTemplateSlots, slotLabel(s))
			}
			sectionsMetadata = append(sectionsMetadata, carryStoredSection(s))
			carried++
			continue
		}
		entry, rendered := renderPlannedSection(ctx, params, s, thisSection, cls, pass,
			resolver, baseData, siteID, pageID, pageName, domain, pageURL, reason, logger)
		if !rendered {
			sectionsMetadata = append(sectionsMetadata, carryStoredSection(s))
			carried++
			continue
		}
		sectionsMetadata = append(sectionsMetadata, entry)
		reRendered++
	}

	return &rerenderSectionsOutcome{
		sectionsMetadata:       sectionsMetadata,
		reRendered:             reRendered,
		carried:                carried,
		resolution:             resolution,
		strippedMarkdownFields: pass.strippedMarkdownFields,
	}
}

// Carry kinds. Each names one of the four legitimate reasons a section keeps its
// stored HTML instead of re-rendering, and each maps to its own resolution
// bucket — the buckets exist because a run in which EVERY section carried is
// otherwise indistinguishable from one that worked (bugs_open/182).
const (
	carryInvalidTemplate = "invalid_template"
	carryNotFound        = "not_found"
	carryNotReady        = "not_ready"
	carryEmptyTemplate   = "empty_template"
)

// sectionClassification is the decision the carry guards reach about one row,
// with no loop state mutated. carryKind empty means the row renders.
type sectionClassification struct {
	comp         componentInfo
	plan         sectionPlanItem
	htmlTemplate string
	carryKind    string
}

// classifyStoredSection runs the carry guards for one row and returns what it
// decided. Extracted from rerenderFlatSections so a COMPOSITION CHILD gets the
// identical treatment a section does — a child is an ordinary page_components
// row (features_open/035 D1), so anything that carries a section must carry a
// child for the same reason and into the same bucket.
//
// Every log line and every message below is the code that was inline. What
// changed is only that the branches RETURN a kind rather than appending to the
// loop's slices, because the caller owns those.
//
// It allocates NO instance token, deliberately. The live rule is that a CARRIED
// section contributes no token, so classification has to complete before any
// token is taken — see the counter note in rerenderFlatSections.
func classifyStoredSection(
	ctx context.Context,
	s storedSection,
	thisSection sectionRef,
	resolveComponent func(storedSection) (componentInfo, bool, bool),
	resolver *sourceResolver,
	logger *zap.Logger,
) sectionClassification {
	var c sectionClassification
	comp, invalidTemplate, haveComp := resolveComponent(s)

	// Can't load the component → carry the stored HTML untouched. Named in
	// the diagnostic (not just logged) because a re-render in which every
	// section takes this branch is otherwise indistinguishable from one
	// that worked — bugs_open/182.
	if !haveComp {
		if invalidTemplate {
			logger.Warn("rerender_page_sections: component template invalid, carrying stored HTML",
				zap.String("section", s.slotName), zap.Int("position", s.position))
			c.carryKind = carryInvalidTemplate
		} else {
			logger.Warn("rerender_page_sections: component not found, carrying stored HTML",
				zap.String("section", s.slotName), zap.Int("position", s.position))
			c.carryKind = carryNotFound
		}
		return c
	}

	// Reuse planSection to rebuild resolved_data (side-effect-free).
	plan := planSection(ctx, s.slotName, thisSection, comp, resolver, logger)
	if plan.Status != "ready" {
		// A required non-LLM field can't resolve now — carry the stored HTML
		// rather than render a broken/empty section. Legitimate, evidenced
		// fallback (bugs_closed/095's council history): named, not fatal.
		logger.Info("rerender_page_sections: section not ready, carrying stored HTML",
			zap.String("section", s.slotName),
			zap.String("status", plan.Status))
		c.carryKind = carryNotReady
		return c
	}

	htmlTemplate, _ := comp.Raw["html_template"].(string)
	if htmlTemplate == "" {
		// Also legitimate/non-fatal — an intentionally empty template stub.
		logger.Warn("rerender_page_sections: empty html_template, carrying stored HTML",
			zap.String("section", s.slotName))
		c.carryKind = carryEmptyTemplate
		return c
	}
	c.comp = comp
	c.htmlTemplate = htmlTemplate
	return c
}

// rerenderPass carries the per-PAGE mutable state one section's render may touch.
// It exists because the render body reaches 7 pieces of loop state — the lazily
// loaded CTA targets, the resolution buckets, the stripped-markdown record and the
// instance counter — and a composition CHILD must reach exactly the same ones, or
// the two paths would keep separate books about one page.
type rerenderPass struct {
	cta                    *rerenderCTAState
	resolution             *rerenderResolution
	strippedMarkdownFields []string
	instances              *InstanceCounter
}

// renderPlannedSection renders ONE classified row and returns its sections_metadata
// entry. ok=false means the row could not be rendered and its stored HTML must be
// carried — the caller owns that, because only the caller knows whether this is a
// page section or a composition child.
//
// Extracted from rerenderFlatSections verbatim apart from that hoisting, so a
// composition child takes the identical path a section does (features_open/035 D1:
// a child IS an ordinary page_components row).
//
// ⚠ THE TOKEN IS TAKEN HERE, NOT IN classifyStoredSection, and that ordering is
// load-bearing: the live rule is that a CARRIED section contributes no instance
// token, so classification must complete first. A walk that allocated on the way
// down would shift every later element id on any page that carries anything.
func renderPlannedSection(
	ctx context.Context,
	params ActionParams,
	s storedSection,
	thisSection sectionRef,
	cls sectionClassification,
	st *rerenderPass,
	resolver *sourceResolver,
	baseData map[string]interface{},
	siteID, pageID uuid.UUID,
	pageName, domain, pageURL, reason string,
	logger *zap.Logger,
) (map[string]interface{}, bool) {
	comp, plan, htmlTemplate := cls.comp, cls.plan, cls.htmlTemplate

	// Derived once, reused by both the CTA recompute below and the
	// observe-only ownership-conflict log — previously computed twice.
	var derivedCTAFields []datahelpers.CTAField
	if schema := datahelpers.ParseInputSchemaValue(comp.Raw["input_schema"]); schema != nil {
		derivedCTAFields = datahelpers.DeriveCTAURLFields(schema)
	}

	// CTA recompute — ONLY for reason=cta_links_stale, so image_landed /
	// section_data_resolved rerenders behave byte-identically to before.
	// After migrations 091/098 the schema no longer sources CTA urls, so a
	// stale url survives in stored content_data; writing the recomputed
	// target into plan.ResolvedData wins the merge below (resolved_data last).
	if reason == livespec.ReasonCTALinksStale { // bugs_open/404: named, not re-spelled
		fn := comp.Function
		if fn == "" {
			fn = s.slotName
		}
		if fields, isCTA := ctaFieldNames[fn]; isCTA {
			if st.cta == nil {
				st.cta = loadRerenderCTAState(ctx, params, siteID, pageName, pageURL, logger)
			}
			if plan.ResolvedData == nil {
				plan.ResolvedData = map[string]interface{}{}
			}
			labelFieldOf := make(map[string]string, len(derivedCTAFields))
			for _, cf := range derivedCTAFields {
				labelFieldOf[cf.URLField] = cf.LabelField
			}
			applyCTARecompute(plan.ResolvedData, s.contentData, fields[0], st.cta.primary, st.cta.validPages, pageURL,
				existingLabelFor(s.contentData, labelFieldOf[fields[0]]), st.cta.candidates, st.cta.pageName)
			applyCTARecompute(plan.ResolvedData, s.contentData, fields[1], st.cta.secondary, st.cta.validPages, pageURL,
				existingLabelFor(s.contentData, labelFieldOf[fields[1]]), st.cta.candidates, st.cta.pageName)
		}
	}

	// OBSERVE-ONLY (council trail 2525f980): this merge is where a
	// resolver-written CTA destination is actually lost — stored
	// content_data (the resolver's last write) merges FIRST, fresh
	// plan.ResolvedData merges LAST and wins. Log each derived CTA field
	// where the fresh value would replace a differing stored one,
	// carrying the rerender reason so deliberate cta_links_stale
	// recomputes are distinguishable from silent clobbers. No behaviour
	// change; the precedence flip returns to the council gate with this
	// log as its evidence. (An earlier sketch placed this log inside
	// planSection, where resolvedData is a fresh local map and the
	// condition could never fire — doc_notes correction, b6e374fc2.)
	for _, cf := range derivedCTAFields {
		stored, hasStored := s.contentData[cf.URLField]
		fresh, hasFresh := plan.ResolvedData[cf.URLField]
		if hasStored && hasFresh && stored != fresh {
			logger.Info("rerender_page_sections: cta ownership conflict (observe-only)",
				zap.String("section", s.slotName),
				zap.String("field", cf.URLField),
				zap.String("source", cf.Source),
				zap.String("reason", reason))
		}
	}

	// bugs_open/184 (canary finding, 2026-08-19): stripping stored
	// content_data alone DOES NOT CONVERGE for query-resolved fields —
	// plan.ResolvedData merges LAST and wins, so a dirty resolver source
	// (content_feed_items.source_summary carries markdown in ~700 rows)
	// re-poisons the very field the strip just cleaned, in the same run.
	// Proven live: dartsonline news-index, items[18].summary stripped then
	// re-imposed, verifier refused. So the SAME double-gated strip runs on
	// the fresh resolved data too — then both the render context below and
	// the persisted mergedContent compose from clean parts. URL-typed
	// resolved fields are safe by pattern construction (a bare URL matches
	// nothing; only [text](url) composites match, which a URL field never
	// carries). The news resolver additionally strips at source
	// (queryresolve/news_items.go) so unflagged callers are covered too.
	//
	// ALIASING, stated (council 060bcc0a r5, editquality/guardian): the
	// strip is in place. plan.ResolvedData is a map planSection allocates
	// fresh per call (plan_sections_action.go, `resolvedData := make(...)`
	// — the doc_notes correction b6e374fc2 is about exactly this: a fresh
	// local map per section), and `plan` is this iteration's local. Its
	// only readers after this line are the render-context merge and
	// mergedContent below, both in this iteration. Nested values MAY alias
	// the resolver's per-invocation caches (sourceResolver.specs /
	// storedContent), whose only other readers are later sections of this
	// same run — and StripLiteralMarkdown is a fixpoint, so a value seen
	// twice is stripped once. Nothing outside this action holds a
	// reference. Callers: rerender_page_sections is run by ONE live step,
	// page-rerender's rerender_sections (measured 2026-08-19, every step,
	// any depth); the reason gate reads the dispatching item's spec, which
	// only the literal_markdown route writes as "literal_markdown".
	//
	// STRIP-TO-EMPTY cannot make a new blank (render_guardian r5): every
	// strip pattern keeps at least one letter/digit of visible text, and
	// the heading strip removes only the `#… ` prefix — so the only input
	// that strips to "" had no letter or digit to begin with
	// (datahelpers/literal_markdown_test.go pins the property). A bare
	// image token `![alt](url)` strips to `!alt`, not to nothing. And the
	// stored-content strip above runs BEFORE the required-field pre-check,
	// whose test is isEmptyContentValue, so an emptied required LLM field
	// would escalate, not render blank.
	if shouldStripLiteralMarkdown(params.StepConfig.Config, reason) && plan.ResolvedData != nil {
		if changed := datahelpers.StripLiteralMarkdownFromContentData(plan.ResolvedData); len(changed) > 0 {
			for _, f := range changed {
				st.strippedMarkdownFields = append(st.strippedMarkdownFields, s.slotName+":resolved:"+f)
			}
			logger.Info("rerender_page_sections: stripped literal markdown from fresh resolved_data",
				zap.String("slot", s.slotName),
				zap.Strings("fields", changed))
		}
	}

	// Render context: base ⊕ stored content_data ⊕ fresh resolved_data
	// (resolved_data merged last so it overrides stale values — matching
	// RenderComponentAction's content_from-then-merge_with ordering).
	rc := &RenderContext{Year: fmt.Sprintf("%d", time.Now().Year())}
	mergeIntoRenderContext(rc, baseData)
	mergeIntoRenderContext(rc, s.contentData)
	if plan.ResolvedData != nil {
		mergeIntoRenderContext(rc, plan.ResolvedData)
	}
	if rc.ContentData == nil {
		rc.ContentData = make(map[string]interface{})
	}
	// {{.ComponentID}} is RETIRED here too (RFC_032 §8, 2026-08-22): it bound
	// the component ROW id, identical for every instance of the same
	// component on a page. All templates spelling it were converted
	// 2026-08-23; the census is zero. See v3_site_actions.go's note for why
	// the binding is deleted rather than left bound-but-unused.
	//
	// The counter advances only for sections that RESOLVED a component: a
	// carried section (component missing or template invalid) keeps its
	// stored HTML and contributes no new ids, so counting it would shift
	// every later token for no reason. The consequence, stated rather than
	// hidden: if a previously-unresolvable component starts resolving, the
	// tokens after it move, and that re-render is not byte-identical.
	BindInstanceToken(rc, st.instances.Next(comp.Function))

	// bugs_open/342 — this path's pre-check already applies the same rule to
	// STORED content; setting it here covers the merged stored ⊕ resolved
	// data the template actually sees.
	if comp.Raw != nil {
		rc.InputSchema = datahelpers.ParseInputSchemaValue(comp.Raw["input_schema"])
	}
	rendered, _, deadURLFields, renderErr := RenderTemplate(htmlTemplate, rc, logger)

	// bugs_open/260: the seam no longer substitutes a regex render for a
	// failed one, so an execution failure arrives here for the first time.
	// CARRY the stored HTML — this path's own existing answer to "this
	// section cannot be safely re-rendered" (four sibling branches above,
	// carryStoredSection). Never replace good stored bytes with a failed
	// render, and never fail the whole page: this action IS the repair
	// vehicle, and a re-render that refuses on the state it was dispatched
	// to fix would deadlock its own remedy.
	//
	// Named in the diagnostic, not merely logged, for the bugs_open/182
	// reason the other carries cite: a run in which every section took a
	// carry branch is otherwise indistinguishable from one that worked.
	//
	// The type report is an ENRICHER — unconditional because the render has
	// ALREADY failed, so it can refuse nothing that works. There is
	// deliberately NO opt-in pre-render refusal on this path (unlike the
	// build path): the checker keys on the schema rather than the template,
	// so arming one here would carry a page that renders perfectly well,
	// which on the repair vehicle is a regression, not a guard.
	if renderErr != nil {
		var schema map[string]interface{}
		if comp.Raw != nil {
			schema = datahelpers.ParseInputSchemaValue(comp.Raw["input_schema"])
		}
		diagnosis := datahelpers.DescribeTypeViolations(
			datahelpers.ContentTypeViolations(schema, rc.ContentData))
		logger.Error("rerender_page_sections: template execution failed — carrying stored HTML, the live section is unchanged",
			zap.String("section", s.slotName),
			zap.String("component_function", comp.Function),
			zap.Error(renderErr),
			zap.String("type_violations", diagnosis),
		)
		st.resolution.RenderFailedSlots = append(st.resolution.RenderFailedSlots, slotLabel(s))
		return nil, false
	}

	// RECORD-ONLY here, deliberately, where the build path refuses
	// (dead_url_guard.go). Two reasons, and neither is squeamishness. First,
	// this path MERGES stored ⊕ fresh below, so it cannot LOSE a key — the
	// worst it can do is re-ship damage that is already live, which refusing
	// would not undo. Second, this is the repair vehicle: a no-LLM re-render
	// is how a fixed row reaches the artefact, and a re-render that refuses
	// on the state it was dispatched to fix would deadlock its own remedy.
	//
	// ⚠ AND IT IS OPT-IN, added in round 2 after THREE seats (guardian,
	// architecture, render_guardian) independently made the same point about
	// the first version: recording was unconditional while the refusal was
	// gated, so this shared repair path would have gained a new DB write on
	// every invocation, on every page, with no default-OFF protection — the
	// exact thing the 2026-08-02 owner ruling asks a shared seam not to do,
	// and inconsistent with my own safety framing one file over. The write
	// is small, but "small and unconditional on a shared path" is how the
	// volume questions this council could not size get answered by
	// production instead of by measurement.
	if recordDeadURLControls(params.StepConfig.Config) &&
		len(deadURLFields) > 0 && !datahelpers.HasRuntimeFillMarker(cls.htmlTemplate) {
		st.resolution.DeadURLSlots = append(st.resolution.DeadURLSlots, slotLabel(s))
		emitSectionDeadControlItem(ctx, params.DB, siteID, nil,
			pageName, s.slotName, comp.Function, deadURLFields, false, logger)
		logger.Warn("rerender_page_sections: URL attribute(s) rendered empty — recorded, not refused",
			zap.String("section", s.slotName),
			zap.Strings("dead_url_fields", deadURLFields))
	}

	// Persisted content_data = stored ⊕ fresh resolved_data, mirroring
	// RenderComponentAction so the row remains a complete render source.
	mergedContent := make(map[string]interface{}, len(s.contentData)+len(plan.ResolvedData))
	for k, v := range s.contentData {
		mergedContent[k] = v
	}
	for k, v := range plan.ResolvedData {
		mergedContent[k] = v
	}

	// stored_slot_name carries the page_components row's OWN identity through
	// to the save, which prefers it verbatim over the component's function
	// (bugs_open/189). This action holds the stored row in hand, so it is the
	// one producer that can never be wrong about it — and without the field
	// the save renamed a positional slot to comp.Function here, defeating the
	// locked-row guard that matches on that name.
	entry := map[string]interface{}{
		"rendered_html":      rendered,
		"component_name":     s.slotName,
		"component_function": comp.Function,
		"stored_slot_name":   s.slotName,
		"content_data":       mergedContent,
	}
	// Provenance for a FRESH render (RFC_046). This path called RenderTemplate
	// above, so the seam has already told us which template text produced these
	// bytes — a fact only it knows, and one this producer held and threw away
	// until now. Carrying the stamp on the carry path while dropping it on the
	// re-render path made the fleet's own repair vehicle write NULL: a page
	// mended through a rerender came out less well-provenanced than one left
	// alone. Empty stays absent — unknown must reach the database as NULL.
	if rc.RenderedTemplateSHA != "" {
		entry["rendered_template_sha"] = rc.RenderedTemplateSHA
	}
	if comp.ID != "" {
		entry["component_id"] = comp.ID
	} else if s.componentID != "" {
		entry["component_id"] = s.componentID
	}

	return entry, true
}
