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
//	}

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
	Required:   []string{"target_site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
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
	Scope       string
	ScopeRef    *string // nil for site scope
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

	type planPageRow struct {
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
		Sections        []string // component names in declared order
	}
	planRows := make([]planPageRow, 0, len(validated))

	for i, v := range validated {
		canonicalName, canonicalURL, canonicalPageType := datahelpers.CanonicalisePage(datahelpers.PageDescriptor{
			Role:          v.Role,
			Slug:          firstNonEmpty(v.Slug, v.Name),
			ParentSection: v.ParentSection,
		})
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

		raw := rawPages[i]
		planRows = append(planRows, planPageRow{
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
			Sections:        extractSectionNames(raw),
		})
	}

	if len(planRows) == 0 {
		return nil, fmt.Errorf("no pages survived validation/canonicalisation")
	}

	// ── 2. Flatten LLM design / content JSON to directive rows ─────────
	directives := flattenSiteScopeDirectives(params.CollectedData, logger)

	// ── 2b. Flatten imagery block to imagery rows (Phase 2G). Empty
	//        until the planner prompt extension ships in step 3; this
	//        is dormant in step 2.
	imageryRows := flattenImageryBlock(params.CollectedData, logger)

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
	//     filled by a downstream resolver.
	sectionsWritten := 0
	for _, r := range planRows {
		for ord, componentName := range r.Sections {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO site_plan_sections
				    (plan_id, page_name, ordering, component_name)
				VALUES ($1, $2, $3, $4)
			`, planID, r.Name, ord, componentName)
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
		imageryLocksTransferred, err = transferImageryLocks(ctx, tx, prevPlanID, planID, logger)
		if err != nil {
			return nil, fmt.Errorf("imagery lock transfer: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

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
	)

	return map[string]interface{}{
		"plan_id":                   planID.String(),
		"superseded_plan_id":        prevPlanIDStr,
		"pages_written":             len(planRows),
		"sections_written":          sectionsWritten,
		"directives_written":        len(directives),
		"locks_transferred":         locksTransferred,
		"imagery_written":           imageryWritten,
		"imagery_locks_transferred": imageryLocksTransferred,
		"role_corrections":          corrections,
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

// extractSectionNames pulls the per-page sections list out of the raw
// LLM page map. Tolerates both shapes the planner has historically
// emitted: a list of strings ("hero", "features") or a list of objects
// ({"name": "hero", ...}). Returns names in declared order.
func extractSectionNames(raw map[string]interface{}) []string {
	v, ok := raw["sections"].([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(v))
	for _, item := range v {
		switch x := item.(type) {
		case string:
			if x != "" {
				out = append(out, x)
			}
		case map[string]interface{}:
			for _, key := range []string{"name", "type", "component", "component_name"} {
				if name, ok := x[key].(string); ok && name != "" {
					out = append(out, name)
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

	// Page scope
	if pageMap, ok := imagery["pages"].(map[string]interface{}); ok {
		for pageName, raw := range pageMap {
			entries, ok := raw.([]interface{})
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
		for sectionRef, raw := range sectionMap {
			entries, ok := raw.([]interface{})
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

	return imageryRow{
		Scope:       scope,
		ScopeRef:    scopeRef,
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
// Returns the number of locks successfully transferred.
func transferImageryLocks(ctx context.Context, tx *sql.Tx, prevPlanID, newPlanID uuid.UUID, logger *zap.Logger) (int, error) {
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
