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
// ── THE COPY GATES, AND WHY THEY LIVE HERE (OWNER REQUIREMENT 2026-08-19) ────
//
// The owner authorised the fleet-wide backfill and WAIVED the read-first review
// pass ("I don't need a review"), with one condition given in the same exchange:
//
//	"please make sure the summaries go through the copy guidance and checks so
//	 they don't sound like AI"
//
// The checks REPLACE the review, so they have to be real and they have to run
// before the write. They are enforced HERE, in the action, rather than in the
// calling workflow, because a gate a workflow author can forget to wire is a
// comment (owner ruling 2026-08-02 §2, the same reasoning as overwrite_existing).
//
// ⚠ AND THE OBVIOUS GATE CANNOT BE USED. `check_voice_tells` scans
// `page_components.rendered_html`; there is a standing landmine saying it — and
// every rendered_html census — is BLIND to `pages.title` and
// `pages.meta_description`. So running that check would pass this text without
// ever looking at it. What is reused instead is the layer underneath, which does
// take plain text: `VoiceGate.ScanVoiceSingleValue(string)` and
// `checkBannedClaims([]string, …)`. Same rules, same site config, applied to the
// one string that is actually about to be published.
//
// ⚠ SingleValue, not ScanVoice (bugs_open/338). This field is ONE VALUE, and the
// gate's counts-per-page and shares-over-sentences are not measurements at that
// sample size — they refused a good 24-word description as "mean sentence length
// 24.0 words" and left the page blank on an hourly schedule. The per-hit
// patterns and the em-dash RATE still apply and are unchanged.
//
// Both are OPT-IN AT THE SITE, exactly as they are everywhere else: a site with
// no `voice` spec has no gate, and a site with no evidence register is still
// swept for the fleet-wide banned claims. A site that has not opted in is not
// silently held to rules it never accepted.
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
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
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

	// The copy gates. Both are advisory-by-absence (a site that has not opted in
	// has no gate) and BLOCKING when they fire: a description that reads as AI
	// boilerplate is exactly what the waived review pass would have caught, and
	// there is no later reader to catch it.
	if reason, detail := metaDescriptionFailsCopyGates(ctx, params, candidate, logger); reason != "" {
		logger.Warn("candidate refused by the copy gates — nothing written",
			zap.String("reason", reason), zap.String("detail", detail))
		return map[string]interface{}{
			"updated": false,
			"reason":  reason,
			"detail":  detail,
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

// metaDescriptionFailsCopyGates runs the site's own voice and claims rules over
// the candidate sentence. Returns ("", "") when it may be published.
//
// It reuses the text-level entry points rather than the page-level checks on
// purpose — see the file header: `check_voice_tells` reads rendered_html and
// cannot see this column at all, so calling it would produce a confident pass
// over text it never examined.
//
// A site that has not opted in to a voice gate is not held to one. The
// banned-claims sweep still runs, because its fleet-wide arm applies to every
// site with or without an evidence register (RFC 003 / bugs_closed/104).
func metaDescriptionFailsCopyGates(ctx context.Context, params ActionParams, candidate string, logger *zap.Logger) (reason, detail string) {
	siteID := resolveMetaDescriptionSiteID(params)
	if siteID == uuid.Nil {
		// No site in context: run nothing rather than pretend. Returning a pass
		// here is honest only because the caller cannot have a site-scoped rule
		// to apply; it is logged so a silent skip is visible.
		logger.Warn("copy gates skipped: no site_id resolvable from collected data")
		return "", ""
	}

	blocks := []string{candidate}

	if gate, err := checks.LoadVoiceGate(ctx, params.DB, siteID); err != nil {
		logger.Warn("voice gate could not be loaded — not treating that as a pass",
			zap.Error(err))
		return "voice_gate_unreadable", err.Error()
	} else if gate != nil {
		// ScanVoiceSingleValue, NOT ScanVoice (bugs_open/338).
		//
		// A meta description is ONE VALUE, not a corpus, and the gate carries
		// two kinds of rule. The per-hit patterns and the em-dash rate mean the
		// same thing here as on a page and still apply. The counts-per-page and
		// shares-over-sentences do not: over a sample of one, "mean sentence
		// length" is just this sentence's length, so a good 24-word description
		// was refused against a trip of 22 and the page stayed blank on an
		// hourly schedule. The classification lives next to the checks in
		// datahelpers/voicetells.go and is guarded by a test, so a check added
		// later cannot silently reach this field.
		//
		// This does NOT relax the site's own thresholds — that would disable the
		// rules for the site's PAGES too, which is where they work.
		if findings := gate.ScanVoiceSingleValue(candidate); len(findings) > 0 {
			f := findings[0]
			return "voice_tell", fmt.Sprintf("%s: %s (%q)", f.Check, f.Reason, f.Matched)
		}
	}

	eb := loadEvidenceBase(ctx, params.DB, siteID, logger) // nil = no register; fleet-wide arm still applies
	if issues := checkBannedClaims(blocks, eb, true, siteID.String(), logger); len(issues) > 0 {
		i := issues[0]
		return "banned_claim", fmt.Sprintf("%s: %s", i.Category, i.Description)
	}

	return "", ""
}

// resolveMetaDescriptionSiteID finds the site the candidate belongs to, trying
// the shapes a real loop actually emits before giving up.
func resolveMetaDescriptionSiteID(params ActionParams) uuid.UUID {
	for _, path := range []string{"site_record.site_id", "site_record.id", "input_data.site_id", "site_id"} {
		if s := datahelpers.ExtractNestedFieldString(params.CollectedData, path); s != "" {
			if id, err := uuid.Parse(s); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}
