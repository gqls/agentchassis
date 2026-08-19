// FILE: platform/orchestration/actions/save_page_meta_description_action.go
//
// SavePageMetaDescriptionAction persists an LLM-written meta description onto a
// page. It is the one piece the framework was missing.
//
// ── WHY THIS EXISTS (bugs_open/320) ──────────────────────────────────────────
//
// `pages.meta_description` had seven writers and every one of them was a
// create-or-upsert path: the column was populated at page birth or not at all.
// Measured 2026-08-19, that left **407 of 731 active pages (55.7%) across 26 of
// 27 sites** with no description — served to search engines with nothing under
// the title, and rendered as an empty excerpt wherever the column is read as
// display content (`rebuild_blog_listing_action.go:99` is the case that surfaced
// it, via bugs_open/309).
//
// Nothing could repair one. There was no UPDATE path in Go, none in any live
// agent config (`default_config ~* 'update[[:space:]]+pages'` returned 0 rows),
// and it is not a field a writer agent can reach: `llm_fields`/`llm_field_specs`
// are per-SECTION (`plan_sections_action.go:864-897`) and this is a `pages`
// column. That mismatch is why `content_rewrite` items filed against missing
// descriptions **completed without writing anything** — two measured, both
// targets still empty, one page demonstrably touched by the handler on the way
// through. A work item that reports success and changes nothing is worse than no
// mechanism at all, because it leaves a green record saying the opposite.
//
// ── WHY IT IS ONLY THE PERSIST HALF ──────────────────────────────────────────
//
// Finding the pages is `query_database`. Writing the sentence is
// `execute_llm_prompt`. Both already exist and are exercised constantly, so this
// action deliberately does neither: a workflow composes the three. That also
// keeps the authorship where the owner's 2026-08-06 ruling puts it — the
// FRAMEWORK writes the copy, from the page's own title and content, not a
// session and not a Go format string. (`composedToolMetaDescription` remains the
// right shape for tool pages, where the sentence really is mechanical; it is the
// wrong shape for arbitrary content pages, which is most of the 407.)
//
// ── THE OPT-IN FIELD, AND WHY ITS DEFAULT IS THE SAFE SIDE ───────────────────
//
// Owner ruling 2026-08-02 §2: new authority on a shared seam ships as an opt-in
// field with the UNSAFE default OFF, because "callers must all be careful" is a
// comment, not a control, on a tree this many sessions share.
//
// The unsafe authority here is REPLACING copy that already exists. So
// `overwrite_existing` defaults to **false**: by default this action fills a
// blank and refuses a page that already has a description, which is the whole of
// what bugs_open/320 needs. A caller that genuinely means to replace published
// copy has to say so in its step config, where a reviewer of the CALLER can see
// it.
package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// maxSavedMetaDescription bounds what may be written. Google truncates around
// 155 characters and real copy is written to that; `datahelpers`' own internal/
// public split uses 320 as the point above which a string reads as a build brief
// rather than a description (bugs_closed/103). The same number is used here on
// purpose — two different ceilings for one column is the drift this area keeps
// producing.
const maxSavedMetaDescription = 320

// SavePageMetaDescriptionInputSpec declares the step config keys this action
// reads. Registered so `cmd/config-key-audit` counts them against the RFC_022
// optional-key budget (owner-ruled N=10) rather than seeing this action as zero.
var SavePageMetaDescriptionInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"page_id_field",
		"site_id_field",
		"page_name_field",
		"description_field",
		// The opt-in authority. Default false = refuse to replace existing copy.
		"overwrite_existing",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("save_page_meta_description", SavePageMetaDescriptionInputSpec)
}

// SavePageMetaDescriptionAction writes one page's meta description.
//
// Config:
//   - description_field:  path in collected data holding the text (default
//     "meta_description"; an LLM step's output field is the usual source)
//   - page_id_field:      path to a page id, OR
//   - site_id_field + page_name_field: to look one up
//   - overwrite_existing: bool, DEFAULT FALSE — see the file header
//
// The result map always distinguishes "wrote it" from "declined, and why", so a
// caller cannot read a refusal as a write. That is the specific failure this
// action was built to end.
func SavePageMetaDescriptionAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "save_page_meta_description"))

	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		// Deliberately an ERROR, not a soft {"updated": false}. This action's whole
		// reason for existing is that a no-op which reports success is how the
		// previous route failed; repeating that shape here would be the same bug.
		return nil, fmt.Errorf("save_page_meta_description: no database connection")
	}

	config := params.StepConfig.Config
	if config == nil {
		config = map[string]interface{}{}
	}

	descField := datahelpers.GetStringField(config, "description_field", "meta_description")
	candidate := strings.TrimSpace(datahelpers.ExtractNestedFieldString(params.CollectedData, descField))

	if candidate == "" {
		logger.Warn("no description text at the configured field — nothing written",
			zap.String("description_field", descField))
		return map[string]interface{}{
			"updated": false,
			"reason":  "empty_candidate",
			"field":   descField,
		}, nil
	}

	// Reuse the bugs_closed/103 guard rather than re-deriving the rule. Its own
	// doc comment says it is exported precisely so "a discovery check or a
	// backfill" can ask the same question — this is that backfill. A second copy
	// of the rule is how the two would drift.
	if datahelpers.MetaDescriptionLooksInternal(candidate) {
		logger.Warn("candidate reads as a build brief, not public copy — refused",
			zap.Int("length", len([]rune(candidate))))
		return map[string]interface{}{
			"updated": false,
			"reason":  "candidate_looks_internal",
			"length":  len([]rune(candidate)),
		}, nil
	}
	if len([]rune(candidate)) > maxSavedMetaDescription {
		return map[string]interface{}{
			"updated": false,
			"reason":  "candidate_too_long",
			"length":  len([]rune(candidate)),
			"max":     maxSavedMetaDescription,
		}, nil
	}

	pageID, err := resolveMetaDescriptionPageID(ctx, params, config, logger)
	if err != nil {
		return nil, err
	}
	if pageID == uuid.Nil {
		return nil, fmt.Errorf("save_page_meta_description: could not resolve a page id from config or collected data")
	}

	overwrite := datahelpers.GetBoolField(config, "overwrite_existing", false)

	// One statement, so the decision cannot race a concurrent writer: the WHERE
	// clause carries the overwrite policy rather than a read-then-write in Go.
	// $3 = overwrite. When false the row is only touched if it is currently blank.
	const q = `
		UPDATE pages
		SET meta_description = $2, updated_at = NOW()
		WHERE id = $1
		  AND ($3::bool OR COALESCE(meta_description, '') = '')
		RETURNING id`

	var updatedID uuid.UUID
	err = params.DB.QueryRowContext(ctx, q, pageID, candidate, overwrite).Scan(&updatedID)
	switch {
	case err == sql.ErrNoRows:
		// Either the page vanished or — far more likely — it already has a
		// description and overwrite_existing is false. Report it as a refusal, not
		// as a write and not as an error.
		logger.Info("page already has a description and overwrite_existing is false — left alone",
			zap.String("page_id", pageID.String()))
		return map[string]interface{}{
			"updated": false,
			"reason":  "already_has_description",
			"page_id": pageID.String(),
		}, nil
	case err != nil:
		return nil, fmt.Errorf("save_page_meta_description: update failed: %w", err)
	}

	logger.Info("meta description written",
		zap.String("page_id", updatedID.String()),
		zap.Int("length", len([]rune(candidate))),
		zap.Bool("overwrite_existing", overwrite))

	return map[string]interface{}{
		"updated": true,
		"page_id": updatedID.String(),
		"length":  len([]rune(candidate)),
	}, nil
}

// resolveMetaDescriptionPageID mirrors UpdatePageStatusAction's resolution order
// so a workflow author does not have to learn a second convention for naming the
// page a step acts on.
func resolveMetaDescriptionPageID(ctx context.Context, params ActionParams, config map[string]interface{}, logger *zap.Logger) (uuid.UUID, error) {
	if f := datahelpers.GetStringField(config, "page_id_field", ""); f != "" {
		if s := datahelpers.ExtractNestedFieldString(params.CollectedData, f); s != "" {
			id, err := uuid.Parse(s)
			if err != nil {
				return uuid.Nil, fmt.Errorf("save_page_meta_description: invalid page_id at %q: %w", f, err)
			}
			return id, nil
		}
	}

	// Common in loop iterations.
	if s := datahelpers.ExtractNestedFieldString(params.CollectedData, "current_page.id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			return id, nil
		}
	}

	siteField := datahelpers.GetStringField(config, "site_id_field", "")
	nameField := datahelpers.GetStringField(config, "page_name_field", "")
	if siteField != "" && nameField != "" {
		siteStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteField)
		pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, nameField)
		if siteStr != "" && pageName != "" {
			siteUUID, err := uuid.Parse(siteStr)
			if err != nil {
				return uuid.Nil, fmt.Errorf("save_page_meta_description: invalid site_id at %q: %w", siteField, err)
			}
			return lookupPageID(ctx, params.DB, siteUUID, pageName, logger)
		}
	}

	return uuid.Nil, nil
}
