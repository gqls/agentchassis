// FILE: platform/orchestration/actions/write_site_plan_action.go
//
// WriteSitePlanAction is the plan-builder's terminal step. It takes the
// LLM's single-call output (pages + design_direction + content_strategy
// + imagery in one JSON shape) and writes:
//
//   - one site_plans row (the version anchor)
//   - N site_plan_pages rows (per page, canonicalised + role-validated)
//   - M site_plan_sections rows (per section within each page)
//   - K site_plan_directives rows (flattened from nested design and
//     content JSON; site / page / section scopes; ordering preserved
//     for multi-cardinality subjects)
//   - L site_plan_imagery rows (Phase 2G — structured imagery requests
//     at site / page / section scope; kind + prompt + optional
//     style_hints + constraints)
//
// All in one transaction. Marks the previous current plan superseded.
//
// Lock transfer: any directive on the previous current plan with
// locked_at IS NOT NULL is matched against the new plan by composite
// key (scope, scope_ref, category, subject, ordering). When matched,
// locked_at and locked_by are copied onto the new directive, and the
// HITL-approved directive text wins over the LLM's new text. See
// doc 030 §"Lock transfer across plan rebuilds". Imagery rows mirror
// this — match key (scope, scope_ref, key); locked prompt wins.
//
// The action does NOT emit work items. That responsibility belongs to
// the reconciler (doc 030 step 5). Plan-builder's job ends with
// "the new plan is durably current"; the reconciler decides what to
// build from it.
//
// CollectedData keys read:
//   - page_plan / site_plan  (extracted by existing extractPagesFromPlan)
//   - design_direction / design_intent  (optional — flattened to directives)
//   - content_strategy / content_direction  (optional — flattened to directives)
//   - imagery  (optional — Phase 2G; site/page/section imagery requests)
//   - target_site_id  (declared in WriteSitePlanInputSpec; deliberately
//     not 'site_id' to avoid nested-lookup collision risk)
//
// Returns:
//
//	{
//	  plan_id: <uuid>,
//	  superseded_plan_id: <uuid string, empty if first plan for site>,
//	  pages_written: <int>,
//	  sections_written: <int>,
//	  directives_written: <int>,
//	  locks_transferred: <int>,
//	  imagery_written: <int>,
//	  imagery_locks_transferred: <int>,
//	  role_corrections: <int>,
//	  duplicate_pages_merged: <int>,   // pages that canonicalised onto a
//	                                   // name already taken (bugs_open/215)
//	  imagery_refs_canonicalised: <int>, // imagery scope_refs rewritten onto
//	                                     // the page name the plan actually
//	                                     // holds (bugs_open/214)
//	  imagery_refs_unresolved: <int>,    // scope_refs naming no page in this
//	                                     // plan — written verbatim, each with
//	                                     // an IMAGERY_SCOPE_REF_UNRESOLVED row
//	  imagery_duplicates_merged: <int>,  // imagery rows that collapsed onto one
//	                                     // identity after canonicalisation
//	}
//
// Imagery scope_ref resolution (bugs_open/214) lives in
// write_site_plan_imagery_scope.go. In short: the planner keys imagery by page
// name, this action renames pages via CanonicalisePage, and until 2026-08-10
// nothing carried the reference across — so an imagery row could name a page
// its own plan did not contain, and no build would ever reference it.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// WriteSitePlanInputSpec declares the inputs this action expects.
// Free-form JSON partials (design_direction, content_strategy) are
// pass-through blobs and not declared here — ActionInputSpec is for
// typed scalars.
//
// Field naming: `target_site_id` rather than `site_id` to avoid
// collision with the nested-lookup behaviour in ExtractActionInputs.
// Many callers carry `site_record.site_id` in CollectedData; an input
// named `site_id` could be silently shadowed by that nested value when
// the top-level isn't supplied. `target_site_id` is unambiguous and
// not a column in any of sites / pages / site_record. The caller's
// input_mapping should look like:
//
//	"input_mapping": { "target_site_id": "site_record.site_id" }
var WriteSitePlanInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"target_site_id"},
	Optional:    []string{},
	Defaults:    map[string]interface{}{},
	Deprecated:  map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("write_site_plan", WriteSitePlanInputSpec)
}

// directiveSpec is one row in the LLM-output-to-directive mapping table.
// Each spec declares: which sibling field under the located directive
// tree carries the value, what (category, subject) the directive
// belongs to, and whether it produces a single row or multiple rows
// (one per item in a list).
//
// The table is the contract. Adding a new directive subject means
// adding one row here, not changing flatten logic.
type directiveSpec struct {
	llmField string // sibling key under the directive tree (e.g. "colour_mood")
	scope    string // "site" — page/section directives come from per-page LLM fields, handled separately
	category string // "design" | "content"
	subject  string // canonical subject name on the directive row
	multi    bool   // false: scalar string. true: list of strings; one row per item.
}

// directiveTreeSource describes where to find a directive tree in the
// LLM output. The tree is a map of sibling fields (e.g.
// {colour_mood: "...", typography_mood: "...", avoid: [...]}).
//
// The action looks for trees by these names — both legacy (design_intent,
// content_direction) and new (design_direction, content_strategy) —
// and uses the first one found per category. That makes the action
// tolerant to the planner prompt being migrated incrementally.
type directiveTreeSource struct {
	treeKey  string
	category string // "design" | "content"
}

var directiveTreeSources = []directiveTreeSource{
	{treeKey: "design_direction", category: "design"},
	{treeKey: "design_intent", category: "design"},
	{treeKey: "content_strategy", category: "content"},
	{treeKey: "content_direction", category: "content"},
}

// siteScopeDesignDirectiveSpecs declares fields under a design-category
// directive tree.
var siteScopeDesignDirectiveSpecs = []directiveSpec{
	{llmField: "style_direction", scope: "site", category: "design", subject: "style_direction", multi: false},
	{llmField: "colour_mood", scope: "site", category: "design", subject: "colour_mood", multi: false},
	{llmField: "typography_mood", scope: "site", category: "design", subject: "typography_mood", multi: false},
	{llmField: "imagery_direction", scope: "site", category: "design", subject: "imagery_direction", multi: false},
	{llmField: "layout_preference", scope: "site", category: "design", subject: "layout_preference", multi: false},
	{llmField: "avoid", scope: "site", category: "design", subject: "avoid", multi: true},
}

// siteScopeContentDirectiveSpecs declares fields under a content-category
// directive tree.
var siteScopeContentDirectiveSpecs = []directiveSpec{
	{llmField: "voice", scope: "site", category: "content", subject: "voice", multi: false},
	{llmField: "emphasis", scope: "site", category: "content", subject: "emphasis", multi: false},
	{llmField: "avoid_phrases", scope: "site", category: "content", subject: "avoid_phrases", multi: true},
	{llmField: "social_proof_style", scope: "site", category: "content", subject: "social_proof_style", multi: false},
	{llmField: "blog_strategy", scope: "site", category: "content", subject: "blog_strategy", multi: false},
}

// directiveRow is the in-memory shape we accumulate before writing.
// It maps 1:1 to a row in site_plan_directives.
type directiveRow struct {
	Scope     string
	ScopeRef  *string // nil for site scope
	Category  string
	Subject   string
	Directive string
	Ordering  int
	Source    string
}

// imageryRow is the in-memory shape for site_plan_imagery rows.
// Maps 1:1 to a row in site_plan_imagery (Phase 2G).
//
// StyleHints and Constraints carry the serialised JSONB bytes (nil
// when absent). Caller passes them via the nullableJSONB helper below
// to coerce empty/nil to a SQL NULL.
type imageryRow struct {
	Scope    string
	ScopeRef *string // nil for site scope
	// RawScopeRef is the planner's own spelling, before canonicalisation.
	// It is NOT written to any column — it exists so a collapse log can name
	// both entries that collided, and so the dedupe guard can prefer the entry
	// that was already aimed at the plan's real page name. Mirrors
	// planPageRow.RawName and its reasoning (bugs_open/214).
	RawScopeRef string
	Key         string
	Kind        string
	Prompt      string
	StyleHints  []byte // serialised JSONB; nil for absent
	Constraints []byte // serialised JSONB; nil for absent
	Ordering    int
	Source      string
}

// validImageryKinds is the in-process mirror of the chk_kind constraint
// on site_plan_imagery. Adding a new kind requires both a code change
// here and a migration to extend the constraint. Keep in sync.
var validImageryKinds = map[string]bool{
	"logo":         true,
	"hero":         true,
	"illustration": true,
	"icon":         true,
	"infographic":  true,
	"sprite_sheet": true, // Phase I2 (SQL_2026-07-12_add_sprite_sheet_kind.sql)
}

// nullableJSONB coerces a byte slice into the value to pass to
// tx.ExecContext for a JSONB column. Empty/nil → SQL NULL; otherwise
// the bytes. Inlined here rather than added to datahelpers to keep the
// package surface tight; lift later if used elsewhere.
func nullableJSONB(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return b
}

// planPageRow is the in-memory shape of one site_plan_pages row, plus the
// page's section list (written to site_plan_sections under the same name).
//
// RawName is the LLM's own spelling before canonicalisation. It is NOT
// written to any column — it exists so that a merge log line can name the
// two emitted entries that collided, which is the only thing that makes a
// collision diagnosable after the fact (see dedupePlanPageRows).
type planPageRow struct {
	RawName string
	// ReconciledFrom is the spelling the PLANNER used for a page the reconciler
	// then snapped onto a realised identity (bugs_open/215). Like RawName it is
	// written to no column: it exists so the imagery scope map can still resolve
	// a block keyed to the planner's spelling after the name changed under it
	// (bugs_open/214's name-space mismatch).
	ReconciledFrom  string
	Name            string
	Role            string
	Slug            string
	URL             string
	ParentSection   string
	Title           string
	MetaDescription string
	NavLabel        string
	InHeader        bool
	InFooter        bool
	NavOrder        int
	Sections        []sectionEntry // component names + fact assignments, declared order
}

// dedupePlanPageRows collapses rows that canonicalise to the same name,
// returning the surviving rows in first-appearance order and the number of
// merges performed.
//
// WHY THIS EXISTS (bugs_open/215). CanonicalisePage is deliberately
// idempotent and collapsing: it maps several distinct LLM spellings onto one
// identity. Three families do this today —
//
//	prefix collapse:  slug "llm-cost-calculator" and slug
//	                  "tool-llm-cost-calculator", both role "tool", both
//	                  become "tool-llm-cost-calculator" (same for guide-,
//	                  game-)
//	homepage collapse: role "index", and slug "home"/"index" under an
//	                  empty/content/landing role, all become "index"
//	section-index:    slug "guides" and slug "guides-index" under any
//	                  section-index role both become "guides-index"
//
// so a planner that emits one page under two spellings in a single plan —
// emission variance, not a prompt defect — produces two rows carrying one
// name. Nothing deduped after canonicalisation, and BOTH destination tables
// are uniquely indexed on that name: site_plan_pages on (plan_id, name) and
// site_plan_sections on (plan_id, page_name, ordering). The insert therefore
// aborted the whole transaction and the entire replan was lost, with an error
// naming neither the canonicaliser nor the LLM (SQLSTATE 23505 on
// idx_site_plan_pages_name, live incident 2026-08-08).
//
// Merge rule: the entry with MORE sections wins, because a composed page and
// a zero-section stub is the observed shape and the stub carries nothing; on
// a tie the first wins, so the rule is total and order-stable. Whenever the
// LOSER had sections of its own the merge is logged at Warn with both lists,
// because that is the only case where authored content is actually discarded
// — but the log only reports that choice, it does not make it. Empty optional
// metadata on the winner is backfilled from the loser: it can fill a blank,
// never overwrite an authored value.
//
// This makes the collision unrepresentable at the only door that writes these
// tables, whatever the LLM emits.
//
// The third return lists every merge where the LOSER was itself composed —
// authored content actually discarded. The Warn below reports the same event
// but chassis logs rotate sub-second, so the caller must persist these
// durably; the lossy branch is otherwise unobservable after the fact (owner
// ruling 2026-08-10 on bugs_open/215: richer-wins stands, on condition the
// loss is recorded somewhere that survives).
func dedupePlanPageRows(rows []planPageRow, logger *zap.Logger) ([]planPageRow, int, []lossyPageMerge) {
	out := make([]planPageRow, 0, len(rows))
	at := make(map[string]int, len(rows)) // canonical name -> index in out
	merges := 0
	var lossy []lossyPageMerge

	for _, r := range rows {
		i, seen := at[r.Name]
		if !seen {
			at[r.Name] = len(out)
			out = append(out, r)
			continue
		}
		merges++

		winner, loser := out[i], r
		if len(r.Sections) > len(out[i].Sections) {
			winner, loser = r, out[i]
		}

		// Both composed: a real section list is being dropped. Loud, and
		// name both lists so the discarded content is recoverable from logs.
		// This branch REPORTS the loss; it does not decide it — the
		// section-count rule above already picked the entry that discards
		// least, and overriding it here would drop the richer page.
		if len(loser.Sections) > 0 {
			logger.Warn("WriteSitePlanAction: two composed pages canonicalise to one name; keeping the richer, DISCARDING the other's sections",
				zap.String("canonical_name", r.Name),
				zap.String("kept_raw_name", winner.RawName),
				zap.String("dropped_raw_name", loser.RawName),
				zap.Strings("kept_sections", sectionNames(winner.Sections)),
				zap.Strings("dropped_sections", sectionNames(loser.Sections)),
			)
			lossy = append(lossy, lossyPageMerge{
				CanonicalName:   r.Name,
				KeptRawName:     winner.RawName,
				DroppedRawName:  loser.RawName,
				KeptSections:    sectionNames(winner.Sections),
				DroppedSections: sectionNames(loser.Sections),
			})
		} else {
			logger.Info("WriteSitePlanAction: duplicate page collapsed after canonicalisation",
				zap.String("canonical_name", r.Name),
				zap.String("kept_raw_name", winner.RawName),
				zap.String("dropped_raw_name", loser.RawName),
				zap.Int("kept_sections", len(winner.Sections)),
				zap.Int("dropped_sections", len(loser.Sections)),
			)
		}

		// Backfill only-blank optional metadata from the discarded entry.
		if winner.Title == "" {
			winner.Title = loser.Title
		}
		if winner.MetaDescription == "" {
			winner.MetaDescription = loser.MetaDescription
		}
		if winner.NavLabel == "" {
			winner.NavLabel = loser.NavLabel
		}
		out[i] = winner
	}

	return out, merges, lossy
}

// lossyPageMerge is one merge that discarded a composed page's section list —
// the shape dedupePlanPageRows's Warn reports, carried out of the pure helper
// so the caller can persist it where a sub-second log rotation cannot eat it.
type lossyPageMerge struct {
	CanonicalName   string
	KeptRawName     string
	DroppedRawName  string
	KeptSections    []string
	DroppedSections []string
}

// recordLossyPageMerges persists one durable row per merge in which authored
// sections were discarded. agent_error_log for the standing reason: the plan
// this write produces is identical whether or not content was lost here, so
// the row is the only artefact that can answer "did richer-wins ever actually
// drop anything" — the measurement the 2026-08-10 ratification of
// bugs_open/215's merge rule is conditional on. Best-effort: a failed write
// must not fail the plan write it describes.
func recordLossyPageMerges(ctx context.Context, params ActionParams, siteID string, merges []lossyPageMerge) {
	if len(merges) == 0 {
		return
	}
	findings := make([]agenterrors.Finding, 0, len(merges))
	for _, m := range merges {
		findings = append(findings, agenterrors.Finding{
			ErrorCode: "PLAN_PAGE_MERGE_LOSSY",
			Severity:  "warning",
			Message: fmt.Sprintf("pages %q and %q canonicalise to %q; kept the richer, discarded %d authored section(s)",
				m.KeptRawName, m.DroppedRawName, m.CanonicalName, len(m.DroppedSections)),
			Context: map[string]interface{}{
				"canonical_name":   m.CanonicalName,
				"kept_raw_name":    m.KeptRawName,
				"dropped_raw_name": m.DroppedRawName,
				"kept_sections":    m.KeptSections,
				"dropped_sections": m.DroppedSections,
				"remedy":           "the discarded composition is recoverable from this row; if both variants were wanted as separate pages, re-plan with distinct roles/slugs that do not canonicalise together (bugs_open/215)",
			},
		})
	}
	LogActionFindings(ctx, params, siteID, "", "write_site_plan", findings, params.Logger)
}

// sectionNames projects section entries to their component names, for logging.
func sectionNames(entries []sectionEntry) []string {
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name)
	}
	return names
}

// WriteSitePlanAction handler.
func WriteSitePlanAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "write_site_plan"))
	logger.Info("WriteSitePlanAction: Starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		WriteSitePlanInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteIDStr := inputs.Get("target_site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid target_site_id %q: %w", siteIDStr, err)
	}

	// ── 1. Extract LLM page list, validate roles, canonicalise ─────────
	rawPages := extractPagesFromPlan(params.CollectedData, logger)
	if len(rawPages) == 0 {
		return nil, fmt.Errorf("no pages found in page_plan / site_plan")
	}

	llmPages := make([]datahelpers.LLMPlannedPage, 0, len(rawPages))
	for _, p := range rawPages {
		llmPages = append(llmPages, datahelpers.LLMPlannedPage{
			Name:          datahelpers.GetStringField(p, "name", ""),
			Role:          firstNonEmptyField(p, "page_type", "type", "role"),
			Slug:          datahelpers.GetStringField(p, "slug", ""),
			URL:           datahelpers.GetStringField(p, "url", ""),
			ParentSection: datahelpers.GetStringField(p, "parent_section", ""),
			// Whether the reconciler already settled this entry's identity. Read
			// through the ONE shared helper, and set identically in
			// SyncPagesToDBAction, for the same reason the URL shape and the
			// identity policy are: these two surfaces must not disagree about a
			// page's identity (bugs_open/241).
			RealisedIdentity: datahelpers.PlanPageCarriesRealisedIdentity(p),
		})
	}
	validated := datahelpers.ValidateRoles(llmPages)

	corrections := 0
	for _, v := range validated {
		if v.CorrectedFromRole != "" {
			corrections++
			logger.Info("WriteSitePlanAction: role corrected by validator",
				zap.String("name", v.Name),
				zap.String("from", v.CorrectedFromRole),
				zap.String("to", v.Role))
		}
	}

	// Site-level URL shape, read through the ONE shared helper — this and
	// SyncPagesToDBAction must agree on it or site_plan_pages and pages carry
	// different URLs for the same page (bugs_open/241).
	flatURLs := siteUsesFlatURLs(ctx, params.DB, siteID, logger)
	// Same rule, same reason: read through the ONE shared helper so this and
	// SyncPagesToDBAction cannot disagree about who owns a page's identity
	// (bugs_open/215).
	identityPolicy := siteIdentityPolicyFor(ctx, params.DB, siteID, logger)

	planRows := make([]planPageRow, 0, len(validated))

	for i, v := range validated {
		raw := rawPages[i]

		canonicalName, canonicalURL, canonicalPageType := datahelpers.CanonicalisePage(datahelpers.PageDescriptor{
			Role:          v.Role,
			Slug:          firstNonEmpty(v.Slug, v.Name),
			ParentSection: v.ParentSection,
			FlatURLs:      flatURLs,
		})
		// A page the reconciler derived from a realised page keeps the identity
		// that page is ACTUALLY serving under, instead of the one its role would
		// synthesise (bugs_open/215). Opt-in per site; without it a legacy-shaped
		// live page is re-derived here and re-minted as a twin.
		if identityPolicy.HonourRealisedIdentity {
			if rName, rURL, rType, ok := realisedIdentityOf(raw); ok {
				if rName != canonicalName || rURL != canonicalURL {
					logger.Info("WriteSitePlanAction: kept realised page identity over canonicalisation",
						zap.String("realised_name", rName), zap.String("realised_url", rURL),
						zap.String("would_have_been_name", canonicalName),
						zap.String("would_have_been_url", canonicalURL))
				}
				canonicalName, canonicalURL, canonicalPageType = rName, rURL, rType
			}
		}
		if canonicalName == "" || canonicalURL == "" {
			logger.Warn("WriteSitePlanAction: page failed canonicalisation, skipping",
				zap.String("raw_name", v.Name),
				zap.String("role", v.Role))
			continue
		}
		// Log when input role and canonical page_type differ — useful
		// for spotting LLM mislabels and the empty-role-defaulted-to-
		// content path. CanonicalisePage normalises dashes to
		// underscores and collapses the section-index family, so a
		// difference here is informative not alarming.
		if v.Role != canonicalPageType {
			logger.Info("WriteSitePlanAction: canonicaliser changed role",
				zap.String("name", canonicalName),
				zap.String("from", v.Role),
				zap.String("to", canonicalPageType))
		}

		planRows = append(planRows, planPageRow{
			RawName:         v.Name,
			ReconciledFrom:  datahelpers.GetStringField(raw, "reconciled_from", ""),
			Name:            canonicalName,
			Role:            canonicalPageType, // canonicaliser's output is authoritative
			Slug:            firstNonEmpty(v.Slug, v.Name),
			URL:             canonicalURL,
			ParentSection:   v.ParentSection,
			Title:           datahelpers.GetStringField(raw, "title", ""),
			MetaDescription: datahelpers.GetStringField(raw, "meta_description", ""),
			NavLabel:        datahelpers.GetStringField(raw, "nav_label", ""),
			InHeader:        datahelpers.GetBoolField(raw, "in_header", true),
			InFooter:        datahelpers.GetBoolField(raw, "in_footer", true),
			NavOrder:        datahelpers.GetIntField(raw, "nav_order", 100),
			Sections:        extractSectionEntries(raw),
		})
	}

	if len(planRows) == 0 {
		return nil, fmt.Errorf("no pages survived validation/canonicalisation")
	}

	// ── 1b. Collapse post-canonicalisation duplicates (bugs_open/215) ──
	//        Must run before ANY insert: both site_plan_pages and
	//        site_plan_sections are uniquely indexed on this name, so a
	//        survivor pair aborts the whole transactional plan write.
	planRows, duplicatesMerged, lossyMerges := dedupePlanPageRows(planRows, logger)
	if duplicatesMerged > 0 {
		logger.Warn("WriteSitePlanAction: plan contained pages that canonicalise to a shared name",
			zap.Int("duplicates_merged", duplicatesMerged),
			zap.Int("pages_after_merge", len(planRows)),
		)
	}
	recordLossyPageMerges(ctx, params, siteIDStr, lossyMerges)

	// Council 4bd35ed8 round-2 advisory (bug_historian, MEDIUM): rule 17's
	// "subject REQUIRED when a component repeats" was prompt-only — the
	// decorative-decision pattern (016b §9). This is the structural half:
	// observe-only, never blocks a plan write.
	if f := subjectlessRepeatFindings(planRows); len(f) > 0 {
		LogActionFindings(ctx, params, siteIDStr, "", "write_site_plan", f, params.Logger)
	}

	// ── 2. Flatten LLM design / content JSON to directive rows ─────────
	directives := flattenSiteScopeDirectives(params.CollectedData, logger)

	// ── 2b. Flatten imagery block to imagery rows (Phase 2G). Empty
	//        until the planner prompt extension ships in step 3; this
	//        is dormant in step 2.
	imageryRows := flattenImageryBlock(params.CollectedData, logger)

	// ── 2c. Resolve imagery scope_refs onto the page names this plan
	//        actually contains (bugs_open/214). The planner keys imagery by
	//        page name; canonicalisation above may have renamed that page
	//        (about -> about-index), and nothing carried the reference
	//        across — so the row named a page its own plan did not have and
	//        no build ever picked it up. planRows is already canonicalised
	//        AND deduped here, so it is the authority. Unresolvable refs are
	//        kept verbatim (today's exact behaviour) and recorded after
	//        commit; the dedupe guard protects idx_site_plan_imagery_unique
	//        from the collapse this resolution makes possible (bugs_open/215
	//        on the neighbouring table).
	canonByRaw := buildCanonicalPageNameMap(planRows, logger)
	imageryRows, imageryMisses, imageryRefsCanonicalised :=
		canonicaliseImageryScopeRefs(imageryRows, canonByRaw, sectionCountByPage(planRows), logger)
	imageryRows, imageryDuplicatesMerged := dedupeImageryRows(imageryRows, logger)

	// ── 3. Single transaction: supersede prior plan, write new plan ────
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 3a. Find previous current plan (for lock transfer) and supersede.
	var prevPlanID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		UPDATE site_plans
		   SET is_current    = false,
		       superseded_at = NOW()
		 WHERE site_id    = $1
		   AND is_current = true
		RETURNING id
	`, siteID).Scan(&prevPlanID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("supersede previous plan: %w", err)
	}

	// 3b. Insert the new plan row.
	var planID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO site_plans (site_id, source_agent, created_by)
		VALUES ($1, $2, $3)
		RETURNING id
	`, siteID, "build-site-planner", "write_site_plan").Scan(&planID)
	if err != nil {
		return nil, fmt.Errorf("insert site_plans: %w", err)
	}

	// 3c. Insert site_plan_pages rows.
	for _, r := range planRows {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO site_plan_pages
			    (plan_id, name, role, slug, url, parent_section,
			     title, meta_description, nav_label,
			     in_header, in_footer, nav_order)
			VALUES ($1, $2, $3, $4, $5, $6,
			        $7, $8, $9,
			        $10, $11, $12)
		`,
			planID, r.Name, r.Role, r.Slug, r.URL,
			datahelpers.NullableString(r.ParentSection),
			datahelpers.NullableString(r.Title),
			datahelpers.NullableString(r.MetaDescription),
			datahelpers.NullableString(r.NavLabel),
			r.InHeader, r.InFooter, r.NavOrder)
		if err != nil {
			return nil, fmt.Errorf("insert site_plan_pages for %q: %w", r.Name, err)
		}
	}

	// 3d. Insert site_plan_sections rows. ordering = position in the
	//     page's section list. component_name = the section component
	//     name from the LLM; resolved IDs (component_version_id,
	//     palette_id, layout_id, typography_set_id) are NULL here and
	//     filled by a downstream resolver. assigned_fact_ids is NULL for
	//     an unscoped section (nil Facts) and a JSON array otherwise —
	//     the NULL/[] distinction is load-bearing (see migration 327).
	//     subject is NULL when the planner assigned none (migration 638).
	sectionsWritten := 0
	for _, r := range planRows {
		for ord, section := range r.Sections {
			var assignedFacts interface{}
			if section.Facts != nil {
				b, merr := json.Marshal(section.Facts)
				if merr != nil {
					return nil, fmt.Errorf("marshal assigned_fact_ids for %q[%d]: %w", r.Name, ord, merr)
				}
				assignedFacts = b
			}
			_, err = tx.ExecContext(ctx, `
				INSERT INTO site_plan_sections
				    (plan_id, page_name, ordering, component_name, assigned_fact_ids, subject)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, planID, r.Name, ord, section.Name, assignedFacts,
				datahelpers.NullableString(section.Subject))
			if err != nil {
				return nil, fmt.Errorf("insert site_plan_sections for %q[%d]: %w", r.Name, ord, err)
			}
			sectionsWritten++
		}
	}

	// 3e. Insert site_plan_directives rows.
	for _, d := range directives {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO site_plan_directives
			    (plan_id, scope, scope_ref, category, subject,
			     directive, ordering, source, source_agent, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`,
			planID, d.Scope, d.ScopeRef, d.Category, d.Subject,
			d.Directive, d.Ordering, d.Source,
			"build-site-planner", "write_site_plan")
		if err != nil {
			return nil, fmt.Errorf("insert site_plan_directives (%s/%s/%s): %w",
				d.Scope, d.Category, d.Subject, err)
		}
	}

	// 3f. Insert site_plan_imagery rows (Phase 2G). Empty list during
	//     transition; populated once the planner prompt extension ships.
	//     Note source/source_agent columns don't exist on site_plan_imagery
	//     — provenance is just the `source` text column on the row itself.
	imageryWritten := 0
	for _, r := range imageryRows {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO site_plan_imagery
			    (plan_id, scope, scope_ref, key, kind, prompt,
			     style_hints, constraints, ordering, source)
			VALUES ($1, $2, $3, $4, $5, $6,
			        $7::jsonb, $8::jsonb, $9, $10)
		`,
			planID, r.Scope, r.ScopeRef, r.Key, r.Kind, r.Prompt,
			nullableJSONB(r.StyleHints),
			nullableJSONB(r.Constraints),
			r.Ordering, r.Source)
		if err != nil {
			scopeRefStr := ""
			if r.ScopeRef != nil {
				scopeRefStr = *r.ScopeRef
			}
			return nil, fmt.Errorf("insert site_plan_imagery (%s/%s/%s): %w",
				r.Scope, scopeRefStr, r.Key, err)
		}
		imageryWritten++
	}

	// 3g. Lock transfer for directives. For each locked directive on the
	//     previous plan, find the matching new directive by composite key
	//     and copy locked_at + locked_by. If the locked text differs from
	//     the LLM's new text, the locked text wins.
	locksTransferred := 0
	if prevPlanID != uuid.Nil {
		locksTransferred, err = transferDirectiveLocks(ctx, tx, prevPlanID, planID, logger)
		if err != nil {
			return nil, fmt.Errorf("lock transfer: %w", err)
		}
	}

	// 3h. Lock transfer for imagery (Phase 2G). Mirrors 3g exactly. Match
	//     key for imagery is (scope, scope_ref, key) under the same site.
	//     Locked prompt wins over LLM rewrite. No-op when prevPlanID is
	//     Nil (first plan for site).
	imageryLocksTransferred := 0
	if prevPlanID != uuid.Nil {
		imageryLocksTransferred, err = transferImageryLocks(ctx, tx, prevPlanID, planID, canonByRaw, logger)
		if err != nil {
			return nil, fmt.Errorf("imagery lock transfer: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// Record anomalous imagery refs AFTER the commit: agenterrors writes on
	// *sql.DB and cannot join the tx anyway, and a rolled-back plan write must
	// not leave error rows describing imagery that was never persisted.
	recordImageryRefMisses(ctx, params, siteID, planID, imageryMisses, logger)

	prevPlanIDStr := ""
	if prevPlanID != uuid.Nil {
		prevPlanIDStr = prevPlanID.String()
	}

	logger.Info("WriteSitePlanAction: Complete",
		zap.String("site_id", siteID.String()),
		zap.String("plan_id", planID.String()),
		zap.String("superseded_plan_id", prevPlanIDStr),
		zap.Int("pages_written", len(planRows)),
		zap.Int("sections_written", sectionsWritten),
		zap.Int("directives_written", len(directives)),
		zap.Int("locks_transferred", locksTransferred),
		zap.Int("imagery_written", imageryWritten),
		zap.Int("imagery_locks_transferred", imageryLocksTransferred),
		zap.Int("role_corrections", corrections),
		zap.Int("duplicate_pages_merged", duplicatesMerged),
		zap.Int("imagery_refs_canonicalised", imageryRefsCanonicalised),
		zap.Int("imagery_refs_unresolved", countMissReason(imageryMisses, "page_not_in_plan")),
		zap.Int("imagery_duplicates_merged", imageryDuplicatesMerged),
	)

	return map[string]interface{}{
		"plan_id":                    planID.String(),
		"superseded_plan_id":         prevPlanIDStr,
		"pages_written":              len(planRows),
		"sections_written":           sectionsWritten,
		"directives_written":         len(directives),
		"locks_transferred":          locksTransferred,
		"imagery_written":            imageryWritten,
		"imagery_locks_transferred":  imageryLocksTransferred,
		"role_corrections":           corrections,
		"duplicate_pages_merged":     duplicatesMerged,
		"imagery_refs_canonicalised": imageryRefsCanonicalised,
		"imagery_refs_unresolved":    countMissReason(imageryMisses, "page_not_in_plan"),
		"imagery_duplicates_merged":  imageryDuplicatesMerged,
	}, nil
}

// flattenSiteScopeDirectives produces directive rows from CollectedData.
// It first locates a design-category tree and a content-category tree
// from the planner's output (tolerating either legacy or new key names
// and wrapper shapes), then walks the per-category spec table to flatten
// each tree's leaf fields into directive rows.
//
// Trees are located via findDirectiveTree, which mirrors the
// "try multiple wrappers" pattern used by extractPagesFromPlan.
// Once a tree is found for a category, the alternate-name tree for the
// same category is ignored — we don't double-write.
func flattenSiteScopeDirectives(data map[string]interface{}, logger *zap.Logger) []directiveRow {
	out := make([]directiveRow, 0, 16)

	for _, category := range []string{"design", "content"} {
		var tree map[string]interface{}
		var foundKey string
		for _, src := range directiveTreeSources {
			if src.category != category {
				continue
			}
			if t := findDirectiveTree(data, src.treeKey); t != nil {
				tree = t
				foundKey = src.treeKey
				break
			}
		}
		if tree == nil {
			logger.Info("WriteSitePlanAction: no directive tree found for category, skipping",
				zap.String("category", category))
			continue
		}
		logger.Info("WriteSitePlanAction: located directive tree",
			zap.String("category", category),
			zap.String("source_key", foundKey))

		var specs []directiveSpec
		switch category {
		case "design":
			specs = siteScopeDesignDirectiveSpecs
		case "content":
			specs = siteScopeContentDirectiveSpecs
		}

		for _, spec := range specs {
			raw, ok := tree[spec.llmField]
			if !ok || raw == nil {
				continue
			}

			if spec.multi {
				items, ok := raw.([]interface{})
				if !ok {
					logger.Info("WriteSitePlanAction: expected list for directive subject, got non-list",
						zap.String("subject", spec.subject))
					continue
				}
				for i, item := range items {
					text, ok := item.(string)
					if !ok || text == "" {
						continue
					}
					out = append(out, directiveRow{
						Scope:     spec.scope,
						ScopeRef:  nil,
						Category:  spec.category,
						Subject:   spec.subject,
						Directive: text,
						Ordering:  i,
						Source:    "llm",
					})
				}
			} else {
				text, ok := raw.(string)
				if !ok || text == "" {
					continue
				}
				out = append(out, directiveRow{
					Scope:     spec.scope,
					ScopeRef:  nil,
					Category:  spec.category,
					Subject:   spec.subject,
					Directive: text,
					Ordering:  0,
					Source:    "llm",
				})
			}
		}
	}
	return out
}

// findDirectiveTree locates a named subtree in CollectedData, tolerating
// the wrapper shapes the planner pipeline produces. Mirrors the
// strategy used by extractPagesFromPlan for the pages list.
//
// Search order:
//  1. data[key]                                         (top level)
//  2. data["site_plan"][key], data["site_plan"]["plan_data"][key],
//     data["site_plan"]["response"][key]               (after validate)
//  3. data["llm_plan"][key], data["llm_plan"]["result"][key],
//     data["llm_plan"]["response"][key]                (raw LLM output)
//  4. data["page_plan"][key], data["page_plan"]["plan_data"][key],
//     data["page_plan"]["response"][key]               (legacy)
//
// Returns nil if no map-shaped value is found at any of these locations.
func findDirectiveTree(data map[string]interface{}, key string) map[string]interface{} {
	// Top-level
	if v, ok := data[key].(map[string]interface{}); ok {
		return v
	}
	// Common wrapper keys
	for _, wrapper := range []string{"site_plan", "llm_plan", "page_plan"} {
		w, ok := data[wrapper].(map[string]interface{})
		if !ok {
			continue
		}
		// Direct under wrapper
		if v, ok := w[key].(map[string]interface{}); ok {
			return v
		}
		// Under wrapper.result / wrapper.response / wrapper.plan_data
		for _, inner := range []string{"result", "response", "plan_data"} {
			if i, ok := w[inner].(map[string]interface{}); ok {
				if v, ok := i[key].(map[string]interface{}); ok {
					return v
				}
			}
		}
	}
	return nil
}

// transferDirectiveLocks copies locked_at / locked_by from previous-plan
// directives onto matching new-plan directives. Match key:
// (scope, scope_ref, category, subject, ordering). When the previous
// directive's text differs from the new one's text, the previous text
// wins (HITL approval > LLM rewrite).
//
// Returns the number of locks transferred.
func transferDirectiveLocks(ctx context.Context, tx *sql.Tx, prevPlanID, newPlanID uuid.UUID, logger *zap.Logger) (int, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT scope, scope_ref, category, subject, ordering, directive, locked_at, locked_by
		FROM site_plan_directives
		WHERE plan_id = $1 AND locked_at IS NOT NULL
	`, prevPlanID)
	if err != nil {
		return 0, fmt.Errorf("query previous locks: %w", err)
	}
	defer rows.Close()

	transferred := 0
	for rows.Next() {
		var (
			scope, category, subject, lockedBy, prevText string
			scopeRef                                     sql.NullString
			ordering                                     int
			lockedAt                                     sql.NullTime
		)
		if err := rows.Scan(&scope, &scopeRef, &category, &subject, &ordering, &prevText, &lockedAt, &lockedBy); err != nil {
			return transferred, fmt.Errorf("scan previous lock: %w", err)
		}

		// Apply the lock — and prevText — onto the matching new-plan row.
		// scope_ref is nullable; query handles both cases.
		var result sql.Result
		if scopeRef.Valid {
			result, err = tx.ExecContext(ctx, `
				UPDATE site_plan_directives
				   SET locked_at = $1,
				       locked_by = $2,
				       directive = $3
				 WHERE plan_id   = $4
				   AND scope     = $5
				   AND scope_ref = $6
				   AND category  = $7
				   AND subject   = $8
				   AND ordering  = $9
			`, lockedAt, lockedBy, prevText, newPlanID, scope, scopeRef.String, category, subject, ordering)
		} else {
			result, err = tx.ExecContext(ctx, `
				UPDATE site_plan_directives
				   SET locked_at = $1,
				       locked_by = $2,
				       directive = $3
				 WHERE plan_id   = $4
				   AND scope     = $5
				   AND scope_ref IS NULL
				   AND category  = $6
				   AND subject   = $7
				   AND ordering  = $8
			`, lockedAt, lockedBy, prevText, newPlanID, scope, category, subject, ordering)
		}
		if err != nil {
			return transferred, fmt.Errorf("apply lock to new directive: %w", err)
		}
		n, _ := result.RowsAffected()
		if n > 0 {
			transferred++
		} else {
			logger.Info("WriteSitePlanAction: locked directive on previous plan has no match on new plan, lock dropped",
				zap.String("scope", scope),
				zap.String("category", category),
				zap.String("subject", subject),
				zap.Int("ordering", ordering))
		}
	}
	if err := rows.Err(); err != nil {
		return transferred, fmt.Errorf("iterate previous locks: %w", err)
	}
	return transferred, nil
}

// sectionEntry is one planned section: its component name and, optionally,
// the evidence-base fact IDs assigned to it at plan time (bugs_open/151
// candidate 1). Facts == nil means unscoped (the pre-existing behaviour: the
// section gets the whole-site writer block); a non-nil empty slice means the
// planner deliberately assigned NO facts to this section.
type sectionEntry struct {
	Name  string
	Facts []string
	// Subject is the planner's one-line statement of what THIS section
	// specifically covers — RFC_016 §5.1's "next structured field"
	// (per-section subjects build, 2026-08-26). Empty = unassigned, the
	// pre-existing behaviour where same-named sections share one brief.
	Subject string
}

// subjectlessRepeatFindings is the structural signal behind seed 640's rule 17:
// a page holding the SAME component name twice where any of those entries has
// no subject is exactly the shape that produced apis.uk's six one-topic
// sections, and prompt compliance is the only other line of defence (an LLM
// re-guesses a rule every run — 016b §9's decorative-decision pattern).
//
// GATED on the plan carrying at least ONE subject anywhere: a pre-rule-17 plan
// (bare strings everywhere — every plan before seed 640 applies) produces ZERO
// findings, so this cannot spam the fleet retroactively; it fires only when
// the planner is demonstrably playing the subject game and missed a repeat.
// Observe-only: the caller records and the write proceeds unchanged.
func subjectlessRepeatFindings(planRows []planPageRow) []agenterrors.Finding {
	anySubject := false
	for _, r := range planRows {
		for _, s := range r.Sections {
			if s.Subject != "" {
				anySubject = true
				break
			}
		}
		if anySubject {
			break
		}
	}
	if !anySubject {
		return nil
	}
	var findings []agenterrors.Finding
	for _, r := range planRows {
		repeats := map[string]int{}
		subjectless := map[string]int{}
		for _, s := range r.Sections {
			repeats[s.Name]++
			if s.Subject == "" {
				subjectless[s.Name]++
			}
		}
		for name, n := range repeats {
			if n < 2 || subjectless[name] == 0 {
				continue
			}
			findings = append(findings, agenterrors.Finding{
				ErrorCode: "SUBJECT_MISSING_ON_REPEATED_COMPONENT",
				Severity:  "warning",
				Message: fmt.Sprintf("page %q holds %d %q sections and %d of them carry no subject — they will receive identical briefs (rule 17, seed 640)",
					r.Name, n, name, subjectless[name]),
				Context: map[string]interface{}{
					"page":            r.Name,
					"component":       name,
					"repeats":         n,
					"without_subject": subjectless[name],
					"remedy":          "replan, or write distinct site_plan_sections.subject values for the repeated rows; the writer cannot differentiate what the plan does not",
				},
			})
		}
	}
	return findings
}

// extractSectionEntries pulls the per-page sections list out of the raw
// LLM page map. Tolerates both shapes the planner has historically
// emitted: a list of strings ("hero", "features") or a list of objects
// ({"name": "hero", ...}). Returns entries in declared order.
//
// Fact assignments are read from two places, in precedence order: the
// section object's own "facts" key, else the page-level "section_facts"
// array written by ValidateSitePlanAction's normalise pass. section_facts
// is indexed by RAW position in the sections array — the lookup happens
// before any malformed entry is skipped, so a skip can never shift an
// assignment onto the wrong section.
func extractSectionEntries(raw map[string]interface{}) []sectionEntry {
	v, ok := raw["sections"].([]interface{})
	if !ok {
		return nil
	}
	pageFacts, _ := raw["section_facts"].([]interface{})
	pageSubjects, _ := raw["section_subjects"].([]interface{})
	// subjectAt mirrors factsAt: RAW-index lookup, before malformed entries
	// are skipped, so a skip can never shift a subject onto the wrong section.
	subjectAt := func(i int) string {
		if i >= len(pageSubjects) {
			return ""
		}
		s, _ := pageSubjects[i].(string)
		return strings.TrimSpace(s)
	}
	factsAt := func(i int) []string {
		if i >= len(pageFacts) {
			return nil
		}
		entry, ok := pageFacts[i].([]interface{})
		if !ok {
			return nil // null / wrong shape -> unscoped
		}
		ids := make([]string, 0, len(entry))
		for _, e := range entry {
			if s, ok := e.(string); ok && s != "" {
				ids = append(ids, s)
			}
		}
		return ids
	}
	out := make([]sectionEntry, 0, len(v))
	for i, item := range v {
		switch x := item.(type) {
		case string:
			if x != "" {
				out = append(out, sectionEntry{Name: x, Facts: factsAt(i), Subject: subjectAt(i)})
			}
		case map[string]interface{}:
			for _, key := range []string{"name", "type", "component", "component_name"} {
				if name, ok := x[key].(string); ok && name != "" {
					facts := factsAt(i)
					if objFacts, ok := x["facts"].([]interface{}); ok {
						ids := make([]string, 0, len(objFacts))
						for _, e := range objFacts {
							if s, ok := e.(string); ok && s != "" {
								ids = append(ids, s)
							}
						}
						facts = ids
					}
					subject := subjectAt(i)
					if s, ok := x["subject"].(string); ok && strings.TrimSpace(s) != "" {
						subject = strings.TrimSpace(s)
					}
					out = append(out, sectionEntry{Name: name, Facts: facts, Subject: subject})
					break
				}
			}
		}
	}
	return out
}

// firstNonEmpty returns the first non-empty string from its arguments.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// firstNonEmptyField reads keys from a map in order, returning the first
// non-empty string value. Used to handle LLM emissions that vary in
// which key carries the role ("page_type" vs "type" vs "role").
func firstNonEmptyField(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v := datahelpers.GetStringField(m, k, ""); v != "" {
			return v
		}
	}
	return ""
}

// ============================================================================
// Imagery helpers (Phase 2G)
// ============================================================================

// flattenImageryBlock walks the "imagery" key in CollectedData and
// produces imageryRow values ready to insert. The expected shape (all
// keys optional):
//
//	imagery: {
//	  site: [{key, kind, prompt, style_hints?, constraints?}, ...],
//	  pages: {page_name: [{...}, ...], ...},
//	  sections: {"page_name:ordering": [{...}, ...], ...}
//	}
//
// Imagery is located via findDirectiveTree, which tolerates the wrapper
// shapes the planner pipeline produces — top-level, site_plan,
// llm_plan.result, etc. If imagery is absent the function returns nil
// and logs a warning (the planner is expected to emit it post-Phase 2G
// step 3).
//
// Rows with malformed fields (missing key/kind/prompt) or unknown kind
// values are SKIPPED with a warning log rather than failing the plan.
// Partial imagery is better than a failed plan write.
func flattenImageryBlock(data map[string]interface{}, logger *zap.Logger) []imageryRow {
	imagery := findDirectiveTree(data, "imagery")
	if imagery == nil || len(imagery) == 0 {
		logger.Warn("flattenImageryBlock: no imagery block found at any expected path; skipping")
		return nil
	}

	var rows []imageryRow

	// Site scope
	if siteEntries, ok := imagery["site"].([]interface{}); ok {
		for ord, raw := range siteEntries {
			if r, ok := buildImageryRow("site", nil, raw, ord, logger); ok {
				rows = append(rows, r)
			}
		}
	}

	// Page scope. Keys are walked in sorted order, not map order: the
	// canonicalisation pass downstream can collapse two spellings onto one
	// identity, and its dedupe guard resolves that by first appearance. Go
	// randomises map iteration per run, so without this the survivor of a
	// collision would vary between two runs over identical input.
	if pageMap, ok := imagery["pages"].(map[string]interface{}); ok {
		for _, pageName := range sortedAnyKeys(pageMap) {
			entries, ok := pageMap[pageName].([]interface{})
			if !ok {
				logger.Info("flattenImageryBlock: page entry not an array, skipping",
					zap.String("page", pageName))
				continue
			}
			pn := pageName
			for ord, e := range entries {
				if r, ok := buildImageryRow("page", &pn, e, ord, logger); ok {
					rows = append(rows, r)
				}
			}
		}
	}

	// Section scope — scope_ref is "page_name:ordering"
	if sectionMap, ok := imagery["sections"].(map[string]interface{}); ok {
		for _, sectionRef := range sortedAnyKeys(sectionMap) {
			entries, ok := sectionMap[sectionRef].([]interface{})
			if !ok {
				logger.Info("flattenImageryBlock: section entry not an array, skipping",
					zap.String("section_ref", sectionRef))
				continue
			}
			sr := sectionRef
			for ord, e := range entries {
				if r, ok := buildImageryRow("section", &sr, e, ord, logger); ok {
					rows = append(rows, r)
				}
			}
		}
	}

	if len(rows) > 0 {
		logger.Info("flattenImageryBlock: rows flattened",
			zap.Int("count", len(rows)))
	}
	return rows
}

// buildImageryRow validates one imagery entry from the LLM output.
// Required fields: key, kind, prompt. kind must be in validImageryKinds.
// Optional fields: style_hints, constraints (JSONB; serialised if present).
//
// Returns (zero, false) for invalid entries — the caller logs and skips.
// We don't fail the whole plan write for a bad entry; the rest of the
// plan is more valuable than perfect imagery.
func buildImageryRow(scope string, scopeRef *string, raw interface{}, ordering int, logger *zap.Logger) (imageryRow, bool) {
	entry, ok := raw.(map[string]interface{})
	if !ok {
		logger.Info("buildImageryRow: entry not a map, skipping",
			zap.String("scope", scope),
			zap.Int("ordering", ordering))
		return imageryRow{}, false
	}

	key := datahelpers.GetStringField(entry, "key", "")
	kind := datahelpers.GetStringField(entry, "kind", "")
	prompt := datahelpers.GetStringField(entry, "prompt", "")

	if key == "" || kind == "" || prompt == "" {
		logger.Info("buildImageryRow: missing required field, skipping",
			zap.String("scope", scope),
			zap.String("key", key),
			zap.String("kind", kind),
			zap.Bool("prompt_empty", prompt == ""),
			zap.Int("ordering", ordering),
		)
		return imageryRow{}, false
	}

	if !validImageryKinds[kind] {
		logger.Info("buildImageryRow: unknown kind, skipping",
			zap.String("scope", scope),
			zap.String("key", key),
			zap.String("kind", kind),
		)
		return imageryRow{}, false
	}

	// style_hints and constraints are optional JSONB. Serialise them
	// directly if present.
	var styleHints, constraints []byte
	if v, ok := entry["style_hints"]; ok && v != nil {
		if b, err := json.Marshal(v); err == nil {
			styleHints = b
		} else {
			logger.Info("buildImageryRow: style_hints not serialisable, dropping",
				zap.Error(err), zap.String("key", key))
		}
	}
	if v, ok := entry["constraints"]; ok && v != nil {
		if b, err := json.Marshal(v); err == nil {
			constraints = b
		} else {
			logger.Info("buildImageryRow: constraints not serialisable, dropping",
				zap.Error(err), zap.String("key", key))
		}
	}

	rawScopeRef := ""
	if scopeRef != nil {
		rawScopeRef = *scopeRef
	}

	return imageryRow{
		Scope:       scope,
		ScopeRef:    scopeRef,
		RawScopeRef: rawScopeRef,
		Key:         key,
		Kind:        kind,
		Prompt:      prompt,
		StyleHints:  styleHints,
		Constraints: constraints,
		Ordering:    ordering,
		Source:      "llm",
	}, true
}

// transferImageryLocks copies locked_at / locked_by (and the locked
// prompt) from previous-plan imagery rows onto matching new-plan rows.
// Match key: (scope, scope_ref, key) under the same plan_id.
// Mirrors transferDirectiveLocks. HITL-approved prompt wins over LLM
// rewrite.
//
// When the new plan has no row matching a previous locked row's key
// (the LLM didn't produce a row at this scope/scope_ref/key in the new
// plan), the lock is dropped and an Info log records it. The
// previous-plan rows persist with is_current=false on their site_plans
// row, so audit retrieval is possible.
//
// TRANSITION (bugs_open/214). Since imagery scope_refs are now canonicalised at
// write time, the first plan written after that change has a new-plan row
// spelt "about-index:2" where the previous plan's locked row still says
// "about:2". The exact match then affects zero rows and a human-approved lock
// is silently dropped. canonByRaw supplies the same resolution the new rows
// received, and is consulted ONLY after the exact match has already found
// nothing — so it cannot alter any transfer that succeeds today. It retires
// itself: from the second post-change replan both sides are canonical.
//
// Returns the number of locks successfully transferred.
func transferImageryLocks(ctx context.Context, tx *sql.Tx, prevPlanID, newPlanID uuid.UUID, canonByRaw map[string]string, logger *zap.Logger) (int, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT scope, scope_ref, key, prompt, locked_at, locked_by
		FROM site_plan_imagery
		WHERE plan_id = $1 AND locked_at IS NOT NULL
	`, prevPlanID)
	if err != nil {
		return 0, fmt.Errorf("query previous imagery locks: %w", err)
	}
	defer rows.Close()

	transferred := 0
	for rows.Next() {
		var (
			scope, key, lockedBy, prevPrompt string
			scopeRef                         sql.NullString
			lockedAt                         sql.NullTime
		)
		if err := rows.Scan(&scope, &scopeRef, &key, &prevPrompt, &lockedAt, &lockedBy); err != nil {
			return transferred, fmt.Errorf("scan previous imagery lock: %w", err)
		}

		// Apply the lock and the locked prompt onto the matching new-plan
		// row. scope_ref is nullable; branch the SQL to handle both cases
		// — same pattern as transferDirectiveLocks.
		var result sql.Result
		if scopeRef.Valid {
			result, err = tx.ExecContext(ctx, `
				UPDATE site_plan_imagery
				   SET locked_at = $1,
				       locked_by = $2,
				       prompt    = $3
				 WHERE plan_id   = $4
				   AND scope     = $5
				   AND scope_ref = $6
				   AND key       = $7
			`, lockedAt, lockedBy, prevPrompt, newPlanID, scope, scopeRef.String, key)
		} else {
			result, err = tx.ExecContext(ctx, `
				UPDATE site_plan_imagery
				   SET locked_at = $1,
				       locked_by = $2,
				       prompt    = $3
				 WHERE plan_id   = $4
				   AND scope     = $5
				   AND scope_ref IS NULL
				   AND key       = $6
			`, lockedAt, lockedBy, prevPrompt, newPlanID, scope, key)
		}
		if err != nil {
			return transferred, fmt.Errorf("apply lock to new imagery row: %w", err)
		}
		n, _ := result.RowsAffected()

		// Transition fallback: the previous plan may hold a pre-canonicalisation
		// ref. Only reached when the exact match found nothing, so no working
		// transfer can be changed by it.
		if n == 0 && scopeRef.Valid && canonByRaw != nil {
			pagePart, rest := splitScopeRef(scopeRef.String)
			if canonical, ok := canonByRaw[normalisePageKey(pagePart)]; ok && canonical != pagePart {
				canonicalRef := canonical + rest
				result, err = tx.ExecContext(ctx, `
					UPDATE site_plan_imagery
					   SET locked_at = $1,
					       locked_by = $2,
					       prompt    = $3
					 WHERE plan_id   = $4
					   AND scope     = $5
					   AND scope_ref = $6
					   AND key       = $7
				`, lockedAt, lockedBy, prevPrompt, newPlanID, scope, canonicalRef, key)
				if err != nil {
					return transferred, fmt.Errorf("apply lock to new imagery row (canonical fallback): %w", err)
				}
				n, _ = result.RowsAffected()
				if n > 0 {
					logger.Info("WriteSitePlanAction: imagery lock matched on the canonicalised scope_ref",
						zap.String("previous_scope_ref", scopeRef.String),
						zap.String("canonical_scope_ref", canonicalRef),
						zap.String("key", key))
				}
			}
		}

		if n > 0 {
			transferred++
		} else {
			logger.Info("WriteSitePlanAction: locked imagery row on previous plan has no match on new plan, lock dropped",
				zap.String("scope", scope),
				zap.String("scope_ref", scopeRef.String),
				zap.String("key", key),
				zap.String("locked_by", lockedBy))
		}
	}
	if err := rows.Err(); err != nil {
		return transferred, fmt.Errorf("iterate previous imagery locks: %w", err)
	}
	return transferred, nil
}
