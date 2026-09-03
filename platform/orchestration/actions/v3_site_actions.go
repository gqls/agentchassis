// FILE: platform/orchestration/actions/v3_site_actions.go
// Additional actions needed for the v3 multipage website builder component-based architecture.
// These complement existing actions in site_db_actions.go and component_library.go.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// SelectStyleCollectionAction selects a style collection for a site
// Either by explicit ID, by site_id lookup, or by domain/industry matching
// Config:
//   - style_collection_id: explicit UUID (optional)
//   - site_id_field: path to site_id in collected_data (optional)
//   - domain_field: path to domain for industry matching (optional)
func SelectStyleCollectionAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("SelectStyleCollectionAction: Starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	if params.DB == nil {
		params.Logger.Warn("SelectStyleCollectionAction: No database, returning default style")
		return getDefaultStyleCollection(), nil
	}

	// Resolve site_id early — needed for persist step at the end
	var siteID uuid.UUID
	if siteIDField, ok := config["site_id_field"].(string); ok && siteIDField != "" {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
		if siteIDStr != "" {
			if parsed, err := uuid.Parse(siteIDStr); err == nil {
				siteID = parsed
			}
		}
	}

	// Helper: persist and return
	persistAndReturn := func(coll *StyleCollection, source string) (interface{}, error) {
		params.Logger.Info("SelectStyleCollectionAction: Resolved style",
			zap.String("name", coll.Name),
			zap.String("source", source),
			zap.String("id", coll.ID.String()))

		// Persist style_collection_id to sites table so downstream agents
		// (webdesign-agent, maintenance agents) can find it via DB lookup
		if siteID != uuid.Nil {
			query := `UPDATE sites SET style_collection_id = $2, updated_at = NOW() WHERE id = $1`
			if err := execDB(ctx, params.DB, query, siteID, coll.ID); err != nil {
				params.Logger.Warn("SelectStyleCollectionAction: Failed to persist style_collection_id",
					zap.Error(err))
				// Non-fatal — continue with the result
			} else {
				params.Logger.Info("SelectStyleCollectionAction: Persisted style_collection_id to sites table",
					zap.String("site_id", siteID.String()),
					zap.String("style_collection_id", coll.ID.String()))
			}
		}

		return styleCollectionToResult(coll), nil
	}

	// Priority 1: Explicit style_collection_id in config
	if scID, ok := config["style_collection_id"].(string); ok && scID != "" {
		scUUID, err := uuid.Parse(scID)
		if err == nil {
			coll, err := getStyleCollectionByID(ctx, params.DB, scUUID, params.Logger)
			if err == nil {
				return persistAndReturn(coll, "explicit_id")
			}
		}
	}

	// Priority 2 (NEW): Planner's style choice via style_from config
	// Config example: "style_from": "site_plan.style_collection"
	// The planner writes a style name (e.g. "professional-dark") to that path
	if styleFromField, ok := config["style_from"].(string); ok && styleFromField != "" {
		styleName := datahelpers.ExtractNestedFieldString(params.CollectedData, styleFromField)
		if styleName != "" {
			params.Logger.Info("SelectStyleCollectionAction: Trying planner style choice",
				zap.String("style_from", styleFromField),
				zap.String("style_name", styleName))
			coll, err := GetStyleCollectionByName(ctx, params.DB, styleName, params.Logger)
			if err == nil && coll != nil {
				return persistAndReturn(coll, "planner_style_from")
			}
			params.Logger.Warn("SelectStyleCollectionAction: Planner style not found in DB",
				zap.String("style_name", styleName))
		}
	}

	// Priority 3: Look up by site_id (existing style_collection_id on sites table)
	if siteID != uuid.Nil {
		coll, err := GetStyleCollectionForSite(ctx, params.DB, siteID, params.Logger)
		if err == nil && coll != nil {
			// Already persisted — just return
			return styleCollectionToResult(coll), nil
		}
	}

	// Priority 4: Match by domain keywords
	domainField := "input_data.domain"
	if df, ok := config["domain_field"].(string); ok && df != "" {
		domainField = df
	}
	domain := datahelpers.ExtractNestedFieldString(params.CollectedData, domainField)
	if domain != "" {
		coll, err := SelectStyleCollectionByDomain(ctx, params.DB, domain, params.Logger)
		if err == nil && coll != nil {
			return persistAndReturn(coll, "domain_keywords")
		}
	}

	// Fallback: Return default (not persisted — it's a synthetic default)
	params.Logger.Info("SelectStyleCollectionAction: No matching style found, using default")
	return getDefaultStyleCollection(), nil
}

func styleCollectionToResult(coll *StyleCollection) map[string]interface{} {
	result := map[string]interface{}{
		"style_collection_id": coll.ID.String(),
		"name":                coll.Name,
		"display_name":        coll.DisplayName,
		"category":            coll.Category,
		"color_palette":       coll.ColorPalette,
		"typography":          coll.Typography,
	}
	if coll.HeaderComponentID != nil {
		result["header_component_id"] = coll.HeaderComponentID.String()
	}
	if coll.FooterComponentID != nil {
		result["footer_component_id"] = coll.FooterComponentID.String()
	}
	if coll.CSSThemeID != nil {
		result["css_theme_id"] = coll.CSSThemeID.String()
	}
	return result
}

func getDefaultStyleCollection() map[string]interface{} {
	return map[string]interface{}{
		"style_collection_id": "",
		"name":                "default",
		"display_name":        "Default Style",
		"category":            "general",
		"color_palette": map[string]string{
			"primary":    "#1a1a2e",
			"secondary":  "#16213e",
			"accent":     "#0f3460",
			"text":       "#333333",
			"background": "#ffffff",
		},
		"typography": map[string]string{
			"heading": "system-ui, sans-serif",
			"body":    "system-ui, sans-serif",
		},
	}
}

// ============================================================================
// ACTION: update_site_content
// ============================================================================

// UpdateSiteContentAction updates the sites.content_data JSONB column
// Used to store the site plan, brand DNA, or other structured content
// Config:
//   - site_id_field: path to site_id in collected_data
//   - content_field: path to content data to store
//   - merge: boolean - if true, merges with existing content_data
func UpdateSiteContentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateSiteContentAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Get content to store
	contentField := "page_plan"
	if f, ok := config["content_field"].(string); ok && f != "" {
		contentField = f
	}
	contentValue := datahelpers.ExtractNestedField(params.CollectedData, contentField)
	if contentValue == nil {
		return nil, fmt.Errorf("content not found at %s", contentField)
	}

	contentJSON, err := json.Marshal(contentValue)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteContentAction: No database, skipping update")
		return map[string]interface{}{
			"updated":     false,
			"site_id":     siteIDStr,
			"reason":      "no database connection",
			"content_key": contentField,
		}, nil
	}

	// Determine if merge or replace
	merge, _ := config["merge"].(bool)

	var query string
	if merge {
		query = `
			UPDATE sites 
			SET content_data = COALESCE(content_data, '{}'::jsonb) || $2::jsonb,
			    updated_at = NOW()
			WHERE id = $1
		`
	} else {
		query = `
			UPDATE sites 
			SET content_data = $2::jsonb,
			    updated_at = NOW()
			WHERE id = $1
		`
	}

	if err := execDB(ctx, params.DB, query, siteID, string(contentJSON)); err != nil {
		return nil, fmt.Errorf("failed to update site content: %w", err)
	}

	// --- Sync key columns to sites table ---
	// When storing brief/identity data, also populate the sites columns
	// so loadSiteDataFull and RenderSiteComponentsAction can find them.
	syncColumns, _ := config["sync_columns"].(bool)
	if syncColumns {
		if contentMap, ok := contentValue.(map[string]interface{}); ok {
			companyName := getFirstNonEmpty(contentMap, "company_name")
			tagline := getFirstNonEmpty(contentMap, "tagline")
			email := getFirstNonEmpty(contentMap, "contact_email", "email")
			phone := getFirstNonEmpty(contentMap, "contact_phone", "phone")

			if companyName != "" || tagline != "" || email != "" || phone != "" {
				syncErr := execDB(ctx, params.DB, `
					UPDATE sites SET
						company_name = CASE WHEN COALESCE(company_name, '') IN ('', domain) AND $2 != '' THEN $2 ELSE company_name END,
						tagline      = CASE WHEN COALESCE(tagline, '')      = '' AND $3 != '' THEN $3 ELSE tagline END,
						email        = CASE WHEN COALESCE(email, '')        = '' AND $4 != '' THEN $4 ELSE email END,
						phone        = CASE WHEN COALESCE(phone, '')        = '' AND $5 != '' THEN $5 ELSE phone END,
						updated_at   = now()
					WHERE id = $1
				`, siteID, companyName, tagline, email, phone)
				if syncErr != nil {
					params.Logger.Warn("UpdateSiteContentAction: column sync failed", zap.Error(syncErr))
				} else {
					params.Logger.Info("UpdateSiteContentAction: synced columns",
						zap.String("company_name", companyName),
						zap.String("email", email),
					)
				}
			}
		}
	}

	params.Logger.Info("UpdateSiteContentAction: Site content updated",
		zap.String("site_id", siteIDStr),
		zap.String("content_field", contentField),
		zap.Bool("merged", merge),
	)

	return map[string]interface{}{
		"updated":      true,
		"site_id":      siteIDStr,
		"content_key":  contentField,
		"content_size": len(contentJSON),
		"merged":       merge,
	}, nil
}

// getFirstNonEmpty returns the first non-empty string value found at any of
// the given keys in the map.
func getFirstNonEmpty(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// ============================================================================
// ACTION: update_site_status
// ============================================================================

// UpdateSiteStatusAction updates the sites.status column
// Config:
//   - site_id_field: path to site_id
//   - status: new status value (draft, building, review, published, archived)
func UpdateSiteStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateSiteStatusAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Get new status
	newStatus, ok := config["status"].(string)
	if !ok || newStatus == "" {
		return nil, fmt.Errorf("status is required in config")
	}

	// Validate status
	validStatuses := map[string]bool{
		"draft": true, "building": true, "review": true,
		"published": true, "deployed": true, "archived": true, "error": true,
	}
	if !validStatuses[newStatus] {
		return nil, fmt.Errorf("invalid status: %s (valid: draft, building, review, published, deployed, archived, error)", newStatus)
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteStatusAction: No database")
		return map[string]interface{}{"updated": false, "status": newStatus}, nil
	}

	// Check if deployed_at should be set
	var query string
	deployedAt, hasDeployedAt := config["deployed_at"].(string)
	if hasDeployedAt && (deployedAt == "now" || deployedAt == "NOW()") && newStatus == "deployed" {
		query = `UPDATE sites SET status = $2, last_deployed_at = NOW(), updated_at = NOW() WHERE id = $1`
	} else {
		query = `UPDATE sites SET status = $2, updated_at = NOW() WHERE id = $1`
	}

	if err := execDB(ctx, params.DB, query, siteID, newStatus); err != nil {
		return nil, fmt.Errorf("failed to update site status: %w", err)
	}

	params.Logger.Info("UpdateSiteStatusAction: Status updated",
		zap.String("site_id", siteIDStr),
		zap.String("status", newStatus),
	)

	return map[string]interface{}{
		"updated":   true,
		"site_id":   siteIDStr,
		"status":    newStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ============================================================================
// ACTION: update_site_defaults
// ============================================================================

// UpdateSiteDefaultsAction updates the sites.default_components JSONB column
// Used to store default header/footer component IDs
// Config:
//   - site_id_field: path to site_id in collected_data
//   - header_component_id: UUID of header component (optional)
//   - footer_component_id: UUID of footer component (optional)
//   - css_theme_id: UUID of CSS theme (optional)
//   - defaults_field: path to a map containing all defaults (alternative to individual fields)
func UpdateSiteDefaultsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateSiteDefaultsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Build defaults map
	defaults := make(map[string]interface{})

	// Try getting from defaults_field first
	if defaultsField, ok := config["defaults_field"].(string); ok && defaultsField != "" {
		if defaultsData := datahelpers.ExtractNestedField(params.CollectedData, defaultsField); defaultsData != nil {
			if m, ok := defaultsData.(map[string]interface{}); ok {
				defaults = m
			}
		}
	}

	// Override/add from explicit config fields
	if headerID, ok := config["header_component_id"].(string); ok && headerID != "" {
		defaults["header_component_id"] = headerID
	}
	if footerID, ok := config["footer_component_id"].(string); ok && footerID != "" {
		defaults["footer_component_id"] = footerID
	}
	if cssThemeID, ok := config["css_theme_id"].(string); ok && cssThemeID != "" {
		defaults["css_theme_id"] = cssThemeID
	}

	// Also check collected data for style_collection results
	if sc := datahelpers.ExtractNestedField(params.CollectedData, "style_collection"); sc != nil {
		if scMap, ok := sc.(map[string]interface{}); ok {
			if hID, ok := scMap["header_component_id"].(string); ok && hID != "" {
				if defaults["header_component_id"] == nil {
					defaults["header_component_id"] = hID
				}
			}
			if fID, ok := scMap["footer_component_id"].(string); ok && fID != "" {
				if defaults["footer_component_id"] == nil {
					defaults["footer_component_id"] = fID
				}
			}
			if cID, ok := scMap["css_theme_id"].(string); ok && cID != "" {
				if defaults["css_theme_id"] == nil {
					defaults["css_theme_id"] = cID
				}
			}
		}
	}

	if len(defaults) == 0 {
		params.Logger.Warn("UpdateSiteDefaultsAction: No defaults to set")
		return map[string]interface{}{
			"updated": false,
			"site_id": siteIDStr,
			"reason":  "no defaults provided",
		}, nil
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteDefaultsAction: No database")
		return map[string]interface{}{
			"updated":  false,
			"site_id":  siteIDStr,
			"defaults": defaults,
			"reason":   "no database connection",
		}, nil
	}

	defaultsJSON, err := json.Marshal(defaults)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal defaults: %w", err)
	}

	query := `
		UPDATE sites 
		SET default_components = COALESCE(default_components, '{}'::jsonb) || $2::jsonb,
		    updated_at = NOW()
		WHERE id = $1
	`

	if err := execDB(ctx, params.DB, query, siteID, string(defaultsJSON)); err != nil {
		return nil, fmt.Errorf("failed to update site defaults: %w", err)
	}

	params.Logger.Info("UpdateSiteDefaultsAction: Defaults updated",
		zap.String("site_id", siteIDStr),
		zap.Any("defaults", defaults),
	)

	return map[string]interface{}{
		"updated":  true,
		"site_id":  siteIDStr,
		"defaults": defaults,
	}, nil
}

// ============================================================================
// ACTION: update_page_status
// ============================================================================

// UpdatePageStatusInputSpec declares this action's step-config contract and opts
// it into unknown-config-key detection (bugs_open/101 machinery,
// datahelpers.UnknownConfigKeys).
//
// There are no Required/Optional entries because this action never calls
// ExtractActionInputs: it reads every value straight from params.StepConfig.Config
// (:543) and resolves the *_field ones itself with ExtractNestedFieldString. So
// the whole contract is ConfigKeys, and CheckConfig is what turns the check on.
//
// Derived by reading the handler body end to end, key by key — NOT by grepping
// for `config["`. That grep is sound HERE (this action uses no
// ResolveConfigSetting / GetStringField / GetIntField indirection at all, which
// was verified rather than assumed), but it is the recorded mistake in
// WRONG_CALLS.md 2026-08-08 and the reading is what establishes the list.
//
// The framework's own keys (input_fields, next_step, output_field, error_step,
// timeout_seconds, the loop_* injections, …) are recognised centrally by
// datahelpers.IsFrameworkStepConfigKey and must NOT be repeated here — listing
// one would misstate whose contract it belongs to.
var UpdatePageStatusInputSpec = datahelpers.ActionInputSpec{
	ConfigKeys: []string{
		"status",                  // :550 — required; the new pages.build_status
		"page_id_field",           // :558 — path to the page id in collected data
		"site_id_field",           // :584 — with page_name_field, the lookup route
		"page_name_field",         // :585 — with site_id_field, the lookup route
		"page_component_id_field", // :799 — mirrors the deploy mark onto one page_component
		// bugs_open/315 / RFC_038: names the DEPLOY step's output field, so this
		// action can read what the deploy actually did before claiming it
		// happened. OPT-IN with the unsafe side as the default: absent means
		// today's behaviour byte for byte. It cannot be a literal — the 19 live
		// git_commit steps carry NINE distinct output_field names (deploy_result
		// on only three; section-editor uses git_result), so a hard-coded name
		// would be blind on 16 of 19 and would fail open on all of them in
		// silence.
		//
		// DECLARED HERE AS OF bugs_open/336, AND THIS IS THE WHOLE OF THAT BUG.
		// It was first declared on RenderComponentInputSpec — the other spec
		// this file's init() registers, for an action that never reads the key —
		// while THIS spec sets StrictConfig below. So the moment migration 494
		// armed the key on live update_page_status steps (2026-08-20 06:49:49Z)
		// the strict branch of validation/workflow.go:checkStepConfigKeys
		// hard-failed every workflow that stamps a page: 8 items across 4 item
		// types, 123 page_rerender queued fleet-wide and none draining, until
		// the config was rolled back.
		//
		// Nothing caught it earlier because every instrument agreed: the literal
		// is in the binary from the reader and its zap.String calls whichever
		// spec lists it, so a /proc/1/exe grep says PRESENT; and
		// `git log -S'"deploy_result_field",'` matches the zap call, so it names
		// the commit that shipped the READER and reads like the declaration's.
		// Only reading the LIST inside the named spec settles it — which is what
		// update_page_status_config_contract_test.go now does on every run.
		"deploy_result_field",
	},

	// DO NOT WRITE A CENSUS COUNT IN THIS COMMENT. An earlier version of this
	// block stated a carrier count as of a date; ~30 sessions share this tree
	// and it was false within the hour (356's data half was applied by its own
	// lane meanwhile). A number here cannot be re-checked by the reader and
	// gets quoted as authority — the concept-register stale-status landmine,
	// wearing a code comment. Ask the fleet instead:
	//   scripts/audit-config-keys.sh   (or the removed-config-keys-check CronJob)
	CheckConfig: true,

	// Three RETIRED keys. This action reads exactly the ConfigKeys above,
	// established by reading the handler end to end (it indexes
	// params.StepConfig.Config directly — no ResolveConfigSetting/GetStringField
	// indirection, and no ExtractActionInputs call, so there is no second access
	// pattern for a key to hide behind).
	//
	//   notes_field / validation_issues_field — retired under RFC_021 Q3
	//     (owner ruling 2026-08-10). Migration 370 removed the one live carrier,
	//     content-reviewer.mark_page_needs_attention.
	//   commit_from — retired 2026-08-11 at the owner's direction. Six live
	//     steps carried it for months; migration 356 (bugfix_136 lane) removed
	//     them. It looked live because coordinator.go's dataRefKeys carried a
	//     comment saying "Used by update_page_status", which was FALSE — that
	//     list only names keys whose VALUE gets rewritten under loop expansion,
	//     never what an action consumes. Three separate readers took it as a
	//     statement of consumption; the entry is gone, and this declaration is
	//     what stops the key drifting back in on that same false authority.
	//
	// All three encoded an author's intent this action has never had, and NONE
	// of the three intents is implemented today. Deleting the keys does not
	// erase them — they are recorded in migrations 356/370's headers and in the
	// messages below.
	//
	// CORRECTED 2026-08-11, by the council's prior_art_librarian seat (round 3,
	// corr 3eb0d1f1): an earlier version of this comment and of `commit_from`'s
	// message claimed the intent was "now genuinely SERVED elsewhere —
	// bugs_open/153's build-provenance stamp, BLD-019". THAT WAS FALSE, and
	// wrong about which provenance. BLD-019 stamps the CHASSIS BINARY with the
	// commit it was BUILT from. `commit_from`'s value was
	// `page_deployed.commit_sha` — the commit a PAGE's content was DEPLOYED in,
	// from the git_commit step beside it. Two unrelated facts about two
	// different artefacts. `pages` still has no column for the second (only
	// `built_from_plan_version`, which is a plan version, not a git sha), so
	// this intent remains exactly as unimplemented as the other two.
	// A retirement message that points at the wrong replacement is worse than
	// one that admits there is none: it sends the next author to build on
	// something that cannot carry them.
	//
	// Each adoption carried its own all-depths census at commit time, per the
	// RFC_021 Q1 protocol. Do not re-derive those numbers from this comment;
	// run the audit.
	RemovedConfigKeys: map[string]string{
		"notes_field": "never read by any version of this action; the intent (recording why " +
			"the page was flagged) is unimplemented — pages has no such column. See migration " +
			"370's header; implement it as a feature if wanted, do not re-add the key",
		"validation_issues_field": "never read by any version of this action; the intent " +
			"(recording the validation issues that flagged the page) is unimplemented — see " +
			"migration 370's header; implement it as a feature if wanted, do not re-add the key",
		"commit_from": "never read by any version of this action — it wrote no column, and " +
			"coordinator.go's dataRefKeys comment claiming otherwise was false (migration 356). " +
			"Its intent (recording which git commit a page's content was DEPLOYED in, from the " +
			"git_commit step's output) is unimplemented — pages has no such column. Do NOT confuse " +
			"it with the build-provenance stamp (bugs_open/153, BLD-019), which is a different " +
			"fact about a different artefact: the commit the CHASSIS BINARY was built from. " +
			"Implement it as a feature if wanted, do not re-add the key",
	},

	// Strict as of bugs_open/234's follow-on (owner decision 2026-08-12), under
	// the RFC_021 Q1 protocol: a fresh all-depths census at adoption, recorded
	// in the adopting commit, plus the standing daily removed-config-keys-check
	// as the ongoing guard — no per-adoption council round or producer
	// inventory required (that question was settled once, by RFC_021 itself).
	// Census at adoption: 9 live steps carry this action (default_config only —
	// task_workflow/orchestrator_workflow/orchestration_workflow carry none),
	// and every key they set was one of the ConfigKeys as they then stood (the
	// list has since gained deploy_result_field — bugs_open/336 — so do not read
	// this sentence as a live count; run the audit). Zero
	// unrecognised. From here an unrecognised key on this action is a
	// definition error caught at validation, not a silent no-op found by
	// archaeology months later — the exact shape that cost bugs_open/234 and
	// this action's own three retired keys.
	StrictConfig: true,
}

// RenderComponentInputSpec is this action's first registered spec (bugs_open/184,
// council 060bcc0a round 3: two seats objected that landing a new capability flag
// on the fleet's most-shared render action with NO spec left it invisible to the
// RFC_022 optional-key budget counter — they were right, so the spec exists now).
//
// The declared set is the census of keys RenderComponentAction actually READS
// (grep config["..."] over the function, 2026-08-18): ten keys, all settings or
// by-name references resolved from step config, none through ExtractActionInputs
// — hence ConfigKeys, not Optional. Declaring ConfigKeys opts the action into
// unknown-config-key REPORTING (warning-only; no StrictConfig — this is a first
// census on a 40+-carrier action, and an over-strict validator is a worse bug
// than an inert key, per action_inputs.go's own header).
//
// DELIBERATELY NOT DECLARED: `output_html` — carried by live step configs and
// read by NOTHING in this binary (grep '"output_html"' = zero hits). Declaring
// it would hide exactly that (the bugs_closed/101 trap: declaring a key is a way
// of hiding it). Leave it undeclared so the coverage report surfaces it; retire
// or implement it as its own task.
var RenderComponentInputSpec = datahelpers.ActionInputSpec{
	ConfigKeys: []string{
		"component_from", "component_function", "component_id",
		"content_field", "content_from",
		"context_field", "context_from",
		"merge_with", "slot_name_from",
		// bugs_open/184: gates the literal-markdown strip on LLM content at
		// birth. Default OFF; enabled per step by migration 474.
		"strip_literal_markdown",
		// bugs_open/260: arms the PRE-render declared-type refusal
		// (mistyped_llm_fields_gate.go). Opt-in, unsafe default OFF, zero live
		// steps set it as this ships; the seam's hard error is unconditional
		// and is NOT gated by this key.
		"refuse_mistyped_llm_fields",
		// bugs_open/238's guard, DECLARED here 2026-08-19 rather than added:
		// this action has read it since the guard shipped, but through
		// shouldRefuseDeadURLControls(config, ...) rather than a literal
		// config["..."] in the function body — so the 2026-08-18 census, which
		// was a grep over the function, could not see it.
		//
		// What that costs is the unknown-config-key REPORT (a live step setting
		// it read as an unknown key) and this list's own honesty, NOT the
		// RFC_022 budget: that counter reads spec.Optional only and skips
		// ConfigKeys on purpose — settings are not the accumulated authority it
		// was built to notice (cmd/config-key-audit/optionalbudget.go:14-21).
		// Both flags here are settings, so ConfigKeys is where they belong and
		// neither moves the budget.
		"refuse_dead_url_controls",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("update_page_status", UpdatePageStatusInputSpec)
	datahelpers.RegisterActionInputSpec("render_component", RenderComponentInputSpec)
}

// UpdatePageStatusAction updates a single page's build_status
// Config:
//   - page_id_field: path to page_id OR
//   - site_id_field + page_name_field: to look up page
//   - status: new build_status value (e.g., "deployed", "failed", "building")
func UpdatePageStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdatePageStatusAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		params.Logger.Warn("UpdatePageStatusAction: No database connection available")
		return map[string]interface{}{"updated": false, "reason": "no database"}, nil
	}

	newStatus, ok := config["status"].(string)
	if !ok || newStatus == "" {
		return nil, fmt.Errorf("status is required")
	}

	var pageID uuid.UUID

	// Try direct page_id from config field
	if pageIDField, ok := config["page_id_field"].(string); ok && pageIDField != "" {
		pageIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, pageIDField)
		if pageIDStr != "" {
			var err error
			pageID, err = uuid.Parse(pageIDStr)
			if err != nil {
				params.Logger.Warn("UpdatePageStatusAction: Invalid page_id format",
					zap.String("page_id_field", pageIDField),
					zap.String("value", pageIDStr),
					zap.Error(err))
				return nil, fmt.Errorf("invalid page_id: %w", err)
			}
		}
	}

	// Alternative: try current_page.id (common in loop iterations)
	if pageID == uuid.Nil {
		if pageIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "current_page.id"); pageIDStr != "" {
			if parsed, err := uuid.Parse(pageIDStr); err == nil {
				pageID = parsed
			}
		}
	}

	// Alternative: look up by site_id + page_name
	if pageID == uuid.Nil {
		siteIDField, _ := config["site_id_field"].(string)
		pageNameField, _ := config["page_name_field"].(string)

		if siteIDField != "" && pageNameField != "" {
			siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
			pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, pageNameField)

			if siteIDStr != "" && pageName != "" {
				siteUUID, _ := uuid.Parse(siteIDStr)
				var err error
				pageID, err = lookupPageID(ctx, params.DB, siteUUID, pageName, params.Logger)
				if err != nil {
					return nil, fmt.Errorf("page lookup failed: %w", err)
				}
			}
		}
	}

	// Last resort: try current_page.name with site_record.site_id
	if pageID == uuid.Nil {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
		pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, "current_page.name")

		if siteIDStr != "" && pageName != "" {
			siteUUID, _ := uuid.Parse(siteIDStr)
			var err error
			pageID, err = lookupPageID(ctx, params.DB, siteUUID, pageName, params.Logger)
			if err != nil {
				params.Logger.Warn("UpdatePageStatusAction: Page lookup by name failed",
					zap.String("site_id", siteIDStr),
					zap.String("page_name", pageName),
					zap.Error(err))
			}
		}
	}

	if pageID == uuid.Nil {
		params.Logger.Error("UpdatePageStatusAction: Could not determine page_id",
			zap.Any("config", config))
		return nil, fmt.Errorf("could not determine page_id")
	}

	// Option B guard: never mark a page "deployed" without rendered components.
	// build_status='deployed' is what the reconciler trusts to skip a page, so a
	// 0-component page marked deployed becomes permanently fileless (gamesdesign
	// homepage, 2026-06-04: its needs_content_page was auto-completed on a lost
	// response with zero components, then a deploy path stamped it deployed, so
	// the reconciler never rebuilt it). Refuse the deploy mark and flip to
	// needs_rebuild (clearing the stamp) so the reconciler re-emits a build.
	// Fail-open on a check error: a transient check failure must not halt
	// legitimate deploys; Option A (the claimed-item-timeout evidence check) is
	// the other layer of protection.
	//
	// bugs_open/040-partial-build: the same reasoning applies one step up, to a
	// build that wrote SOME but not ALL of its planned sections. dartsonline
	// index (2026-07-20) reached this deploy mark with 5 of 6 planned sections
	// (testimonials never written) and was stamped deployed + built_from_plan_version
	// = current, so decideEmit returns skip_built and the reconciler will never
	// revisit the missing section — a permanent five-sixths page that no longer
	// asks to be built. A partial build must be treated exactly like a 0-component
	// one: refuse the mark, flip to needs_rebuild, clear the stamp. This runs
	// AFTER save_page_sections has written the components (page-build-handler,
	// page-rerender, section-editor and tool-recreation-handler all write their
	// components before this step), so a shortfall here is a real one, not a race.
	// Ownership skip (bugs_open/208): the page was deliberately left alone by the
	// owned-page guard, so nothing was assembled and nothing was committed.
	// Stamping it 'deployed' would be a claim about work that did not happen, and
	// worse than cosmetic: the stamp also writes built_from_plan_version = the
	// current plan, which makes ReconcileSitePlanAction's decideEmit return
	// skip_built and permanently suppress the owned_page_review item that is this
	// design's own visibility channel (reconcile_site_plan_action.go:210-214,
	// 238-271). Refusing the stamp leaves the page at needs_rebuild with an open
	// review item — reconcile's existing protocol for "requested, parked for
	// owner-aware handling".
	//
	// Originally keyed to THIS guard's skip only; bugs_open/210 widened it to any
	// assembly skip, with the retry bound that widening was filed to demand (see
	// page_build_failure_guard.go). The OWNED branch keeps its 208 semantics —
	// no status flip, reconcile's owned_page_review protocol owns that state —
	// while every other skip refuses via refuseDeployStampOnSkip: needs_rebuild +
	// cleared plan stamp, an agent_error_log refusal row, and a park after the
	// third consecutive failure so the retry loop is bounded rather than silent.
	if newStatus == "deployed" {
		if skipped, skipReason := upstreamAssemblySkipped(params.CollectedData); skipped {
			if strings.Contains(skipReason, ownedPageSkipReasonPrefix) {
				params.Logger.Warn("UpdatePageStatusAction: refusing to stamp deployed — page was skipped by the owned-page guard",
					zap.String("page_id", pageID.String()),
					zap.String("skip_reason", skipReason))
				return map[string]interface{}{
					"updated": false,
					"page_id": pageID.String(),
					"skipped": true,
					"reason":  "refused deploy stamp: " + skipReason,
				}, nil
			}
			return refuseDeployStampOnSkip(ctx, params, pageID, skipReason), nil
		}
	}

	// Refuse to STAMP an archived page deployed (bugs_open/266).
	//
	// The commit-seam guard (git_deployer_actions.go) stops the file reaching the
	// site; this stops the row claiming it did. Both are needed: they are
	// different damage. Without this, a refused commit still leaves
	// `build_status='deployed'` + a fresh `deployed_at`, which is the exact state
	// 266 was filed about and is what every downstream selector reads.
	//
	// Asked of the DATABASE rather than of an upstream skip marker, because this
	// guard's refusal travels through git_commit's output field rather than the
	// assembly field `upstreamAssemblySkipped` reads — and because the page's own
	// status is the authority here, whatever shape the collected data happens to
	// have taken on the way in.
	//
	// NO STATUS FLIP, matching the OWNED branch above rather than the
	// shortfall guards below: `needs_rebuild` would queue the page to be built
	// again, which for an archived page is precisely the loop being closed.
	if newStatus == "deployed" {
		archived, checked := pageIsArchivedForGuard(ctx, params.DB, pageID, params.Logger)
		if !checked {
			LogActionError(ctx, params,
				datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id"),
				datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain"),
				"update_page_status", "ARCHIVED_PAGE_GUARD_UNCHECKED", "high",
				fmt.Sprintf("status for page %s could not be read; deploy stamp proceeded without an archived check", pageID),
				map[string]interface{}{"page_id": pageID.String()},
				params.Logger)
		}
		if archived {
			reason := fmt.Sprintf("%s: refused deploy stamp — page is status=archived", archivedPageSkipReasonPrefix)
			params.Logger.Warn("UpdatePageStatusAction: ARCHIVED PAGE — refusing to stamp deployed",
				zap.String("page_id", pageID.String()))
			LogActionError(ctx, params,
				datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id"),
				datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain"),
				"update_page_status", "ARCHIVED_PAGE_DEPLOY_REFUSED", "warning",
				reason,
				map[string]interface{}{"page_id": pageID.String()},
				params.Logger)
			return map[string]interface{}{
				"updated": false,
				"page_id": pageID.String(),
				"skipped": true,
				"reason":  reason,
			}, nil
		}
	}

	if newStatus == "deployed" {
		hasComponents, checkErr := pageHasComponents(ctx, params.DB, pageID)
		switch {
		case checkErr != nil:
			params.Logger.Warn("UpdatePageStatusAction: component check failed; proceeding with deploy",
				zap.String("page_id", pageID.String()),
				zap.Error(checkErr))
		case !hasComponents:
			params.Logger.Warn("UpdatePageStatusAction: refusing to mark page deployed with no rendered components; setting needs_rebuild",
				zap.String("page_id", pageID.String()))
			const rebuildQuery = `UPDATE pages SET build_status = 'needs_rebuild', built_from_plan_version = NULL, updated_at = NOW() WHERE id = $1`
			if rbErr := execDB(ctx, params.DB, rebuildQuery, pageID); rbErr != nil {
				params.Logger.Error("UpdatePageStatusAction: failed to set needs_rebuild after refusing deploy",
					zap.String("page_id", pageID.String()),
					zap.Error(rbErr))
				return nil, fmt.Errorf("failed to set needs_rebuild for 0-component page: %w", rbErr)
			}
			return map[string]interface{}{
				"updated":      false,
				"page_id":      pageID.String(),
				"build_status": "needs_rebuild",
				"reason":       "refused deploy: page has no rendered components",
			}, nil
		default:
			// Page has >= 1 component but may still be short of its plan.
			// Fail-open on a check error, same as the 0-component guard above.
			// suppressed_sections (subtracted inside pageSectionShortfall) is
			// maintained by plan_sections' persistSectionSkips: an
			// on_missing=skip_section name is added there and removed again the
			// build it plans ready — so a legitimately data-gated section does
			// not count as a shortfall here (bugs_open/040 skip-not-recorded).
			planned, rendered, shErr := pageSectionShortfall(ctx, params.DB, pageID)
			if shErr != nil {
				params.Logger.Warn("UpdatePageStatusAction: section-shortfall check failed; proceeding with deploy",
					zap.String("page_id", pageID.String()),
					zap.Error(shErr))
			} else if rendered < planned {
				params.Logger.Warn("UpdatePageStatusAction: refusing to mark page deployed; build is short of its plan; setting needs_rebuild",
					zap.String("page_id", pageID.String()),
					zap.Int("planned_sections", planned),
					zap.Int("rendered_components", rendered))
				const rebuildQuery = `UPDATE pages SET build_status = 'needs_rebuild', built_from_plan_version = NULL, updated_at = NOW() WHERE id = $1`
				if rbErr := execDB(ctx, params.DB, rebuildQuery, pageID); rbErr != nil {
					params.Logger.Error("UpdatePageStatusAction: failed to set needs_rebuild after refusing partial deploy",
						zap.String("page_id", pageID.String()),
						zap.Error(rbErr))
					return nil, fmt.Errorf("failed to set needs_rebuild for partial-build page: %w", rbErr)
				}
				return map[string]interface{}{
					"updated":      false,
					"page_id":      pageID.String(),
					"build_status": "needs_rebuild",
					"reason":       fmt.Sprintf("refused deploy: only %d of %d planned sections rendered", rendered, planned),
				}, nil
			}
		}
	}

	// DEPLOY EVIDENCE (bugs_open/315 / RFC_038). Everything above this point
	// asks the DATABASE whether the page deserves a stamp. This asks the DEPLOY
	// STEP whether anything was actually sent — the question nothing has ever
	// asked, and the reason a `git_commit` that returned
	// {"status":"skipped","skip_reason":"no files to commit"} was followed one
	// step later by a fresh deployed_at.
	//
	// OPT-IN, unsafe default OFF: with no deploy_result_field configured this
	// block does nothing at all and behaviour is unchanged.
	//
	// FAIL-OPEN, DELIBERATELY, AND COUNTABLE. If the field is named but no
	// evidence resolves we stamp anyway and write an agent_error_log row. Open,
	// because a config typo — or simply a git-adapter image older than RFC_038,
	// which is the normal state during a partial roll, the chassis and the
	// adapter being separate images — must not freeze deploy stamps fleet-wide.
	// Countable, because a SILENT fail-open is precisely how this bug survived
	// four completed rerenders with every layer reporting success.
	var deployContentHash, deployCommitSHA string
	deployGuardRan := false
	if newStatus == "deployed" {
		if field, _ := config["deploy_result_field"].(string); strings.TrimSpace(field) != "" {
			deployGuardRan = true
			ev, ok := resolveDeployEvidence(params.CollectedData, field, params.Logger)
			switch {
			case !ok:
				params.Logger.Warn("UpdatePageStatusAction: deploy_result_field set but no deploy evidence resolved; stamping anyway",
					zap.String("page_id", pageID.String()),
					zap.String("deploy_result_field", field))
				LogActionError(ctx, params,
					datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id"),
					datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain"),
					"update_page_status", "DEPLOY_EVIDENCE_UNREADABLE", "warning",
					fmt.Sprintf("deploy_result_field %q resolved no commit evidence for page %s; the deploy stamp proceeded unverified", field, pageID),
					map[string]interface{}{"page_id": pageID.String(), "deploy_result_field": field},
					params.Logger)

			case ev.Skipped:
				// Dispatch on the same reason prefixes the guards above use, so
				// a skip means the same thing wherever it is noticed.
				params.Logger.Warn("UpdatePageStatusAction: refusing to stamp deployed — the deploy step reported it skipped",
					zap.String("page_id", pageID.String()),
					zap.String("skip_reason", ev.SkipReason))
				if strings.Contains(ev.SkipReason, ownedPageSkipReasonPrefix) ||
					strings.Contains(ev.SkipReason, archivedPageSkipReasonPrefix) {
					// No status flip: an owned page is parked by reconcile's own
					// protocol, and needs_rebuild for an ARCHIVED page would
					// re-queue the very build being closed off.
					return map[string]interface{}{
						"updated": false,
						"page_id": pageID.String(),
						"skipped": true,
						"reason":  "refused deploy stamp: deploy skipped — " + ev.SkipReason,
					}, nil
				}
				return refuseDeployStampOnSkip(ctx, params, pageID, "deploy skipped — "+ev.SkipReason), nil

			default:
				deployCommitSHA = ev.CommitSHA
				deployContentHash = hashForPageFile(ev.FilesSHA256, pageURLForHash(ctx, params.DB, pageID))
			}
		}
	}

	// Build the query - use build_status column (not status)
	var query string
	args := []interface{}{pageID, newStatus}
	if newStatus == "deployed" {
		// Also set deployed_at, and stamp built_from_plan_version with the site's
		// current plan id. This is the build-time drift stamp the reconciler
		// compares against (029/030 design; the deferred item in HANDOFF_2026-05-07
		// #5). COALESCE keeps any existing value when no current plan exists yet —
		// e.g. tool-recreation deploys before build-site-planner has written the
		// plan — and SyncPagesToDBAction then fills it on its first pass. With this
		// stamp in place the reconciler detects genuine drift (built_from_plan_version
		// != current) rather than relying on the blunt deployed->needs_rebuild flip.
		// content_hash is the sha256 of the bytes this page was actually
		// committed as (bugs_open/315; owner ruling 2026-08-19 "wire up the page
		// fingerprint"). It turns "is the origin serving what we sent?" from a
		// four-step archaeology — pull rendered_html, cut a needle, fetch the
		// page, grep — into one comparison.
		//
		// THE COLUMN IS TOUCHED ONLY WHEN THE GUARD RAN, and then it is ASSIGNED
		// rather than COALESCEd. Both halves matter and an earlier version of
		// this code got the second one wrong:
		//
		//   * guard OFF (no deploy_result_field — every live step today, and
		//     every git_commit step with no stamp after it): the clause is not
		//     in the statement at all, so an unarmed path cannot disturb a hash
		//     some other path wrote.
		//   * guard ON: this stamp means NEW BYTES WENT OUT. Any previous
		//     fingerprint therefore describes an OLDER deploy and is stale by
		//     definition. COALESCE would preserve it, and the divergence check
		//     would then compare live bytes against a superseded intent and
		//     report a healthy page as diverged. NULL means "we do not know what
		//     we sent", which is what the check's `content_hash IS NOT NULL`
		//     predicate is for. An honest unknown beats a confident stale value.
		//
		// page_components.deploy_commit is deliberately NOT written — owner
		// ruling, same day: a section is not a file, so a section-level commit
		// reference cannot answer the publish question. See RFC_038 §8.
		query, args = buildPageDeployStampQuery(args, deployGuardRan, deployContentHash)
	} else {
		query = `UPDATE pages SET build_status = $2, updated_at = NOW() WHERE id = $1`
	}

	result, err := params.DB.ExecContext(ctx, query, args...)
	if err != nil {
		params.Logger.Error("UpdatePageStatusAction: Failed to update page",
			zap.String("page_id", pageID.String()),
			zap.String("build_status", newStatus),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update page build_status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	params.Logger.Info("UpdatePageStatusAction: Updated page",
		zap.String("page_id", pageID.String()),
		zap.String("build_status", newStatus),
		// The commit this page's bytes went out in. There is no column for it —
		// pages.deploy_commit was dropped deliberately as "belongs in
		// page_components", and the owner ruled 2026-08-19 not to wire the
		// section-level one. So it is logged, and it also lives in the
		// orchestration's own collected_data.
		zap.String("deploy_commit_sha", deployCommitSHA),
		zap.String("content_hash", deployContentHash),
		zap.Int64("rows_affected", rowsAffected))

	if newStatus == "deployed" {
		// A successful deploy resolves any parked build-failure condition for
		// this page (bugs_open/210); an open park would keep the page's
		// needs_page:<name> work-item slot blocked after the condition ended.
		closePageBuildFailureItems(ctx, params.DB, pageID, params.Logger)
	}

	// Mirror the deploy mark onto one page_component when the caller names it via
	// config page_component_id_field. Every discovery check matches
	// page_components.build_status = 'deployed', and apply_section_edit leaves its
	// row at 'approved', so without this an edited section silently disappears from
	// the whole audit surface (check_empty_sections, check_image_url_404,
	// check_undeployed_assets, check_placeholder_image_in_use, check_component_standards).
	// Non-fatal: the page row is already committed, and failing here would re-run the
	// step and re-deploy.
	componentUpdated := false
	if newStatus == "deployed" {
		if field, ok := config["page_component_id_field"].(string); ok && field != "" {
			pcIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, field)
			switch {
			case pcIDStr == "":
				params.Logger.Warn("UpdatePageStatusAction: page_component_id_field configured but empty",
					zap.String("page_component_id_field", field))
			default:
				pcID, parseErr := uuid.Parse(pcIDStr)
				if parseErr != nil {
					params.Logger.Warn("UpdatePageStatusAction: invalid page_component_id",
						zap.String("value", pcIDStr),
						zap.Error(parseErr))
					break
				}
				const pcQuery = `UPDATE page_components SET build_status = $2, updated_at = NOW() WHERE id = $1`
				if pcErr := execDB(ctx, params.DB, pcQuery, pcID, newStatus); pcErr != nil {
					params.Logger.Error("UpdatePageStatusAction: failed to mark page_component deployed",
						zap.String("page_component_id", pcID.String()),
						zap.Error(pcErr))
					break
				}
				componentUpdated = true
				params.Logger.Info("UpdatePageStatusAction: Marked page_component deployed",
					zap.String("page_component_id", pcID.String()))
			}
		}
	}

	return map[string]interface{}{
		"updated":                true,
		"page_id":                pageID.String(),
		"build_status":           newStatus,
		"rows_affected":          rowsAffected,
		"page_component_updated": componentUpdated,
	}, nil
}

// pageHasComponents reports whether a page has at least one real rendered
// component (non-null component_id, non-empty rendered_html). This is the
// "positive evidence" check from FOCUS_page_build_handler_silent_completion.md,
// used by Option B to stop a 0-component page being marked deployed. Mirrors the
// db type switch used by lookupPageID/execDB so it works with both *sql.DB and
// *pgxpool.Pool.
func pageHasComponents(ctx context.Context, db interface{}, pageID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM page_components
			WHERE page_id = $1
			  AND component_id IS NOT NULL
			  AND rendered_html IS NOT NULL
			  AND rendered_html <> ''
		)`
	var exists bool
	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, pageID).Scan(&exists)
		return exists, err
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, pageID).Scan(&exists)
		return exists, err
	default:
		return false, fmt.Errorf("unsupported database type: %T", db)
	}
}

// pageSectionShortfall reports how many sections the page's plan promised
// (planned) versus how many page_components rows exist for it (rendered). It is
// the 040-partial-build companion to pageHasComponents: a build that reaches the
// deploy mark having written a row for only 5 of 6 planned sections must NOT be
// stamped deployed + built_from_plan_version, or decideEmit
// (reconcile_site_plan_action.go) returns skip_built for it forever and the
// reconciler never revisits the shortfall (dartsonline index, 2026-07-20:
// testimonials dropped by a build that reported complete; caller flips it to
// needs_rebuild and clears the stamp instead).
//
// It compares ROW COUNTS, deliberately NOT section names. pages.sections names
// do NOT reliably equal page_components.slot_name / content_components.function
// across templates — gaswholesalers services planned ["services-hero", …,
// "call_to_action"] against live slots ["hero-services", …, "call-to-action"]
// (word order and _/- both differ) — so per-name matching produces false
// positives that would refuse a healthy page and drive it into a rebuild loop.
// The row count is the signal the bugs_open/040 fleet sweep validated; a genuine
// missing section (no row at all) is what it catches, while a present-but-hollow
// section (a row that exists but rendered nothing) is bugs_open/039's concern and
// still counts as rendered here. suppressed_sections are excluded from the
// planned count so a deliberately-dropped section is never read as a shortfall.
// Mirrors the db type switch used by pageHasComponents/execDB.
func pageSectionShortfall(ctx context.Context, db interface{}, pageID uuid.UUID) (planned int, rendered int, err error) {
	const query = `
		SELECT
			(SELECT count(*)
			   FROM jsonb_array_elements_text(COALESCE(p.sections, '[]'::jsonb)) AS sec
			  WHERE sec NOT IN (
			      SELECT jsonb_array_elements_text(COALESCE(p.suppressed_sections, '[]'::jsonb)))),
			(SELECT count(*) FROM page_components pc WHERE pc.page_id = p.id)
		FROM pages p
		WHERE p.id = $1`
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, pageID).Scan(&planned, &rendered)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, pageID).Scan(&planned, &rendered)
	default:
		err = fmt.Errorf("unsupported database type: %T", db)
	}
	return planned, rendered, err
}

// BuildRenderContextAction assembles a RenderContext from multiple sources
// Used before rendering components with templates
func BuildRenderContextAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("BuildRenderContextAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Start with empty context
	renderCtx := &RenderContext{
		Year: fmt.Sprintf("%d", time.Now().Year()),
	}

	sourcesMerged := 0

	// Handle sources config - can be array or map format
	// Array format: ["input_data", "site_record", ...]
	// Map format: {"page": "input_data.current_page", "site": "input_data.site_record", ...}

	if sourcesMap, ok := config["sources"].(map[string]interface{}); ok {
		// Map format: keys are logical names, values are paths to data
		params.Logger.Info("Using map-format sources config",
			zap.Int("source_count", len(sourcesMap)))

		for logicalName, pathVal := range sourcesMap {
			path, ok := pathVal.(string)
			if !ok {
				continue
			}

			sourceData := datahelpers.ExtractNestedField(params.CollectedData, path)
			if sourceData == nil {
				params.Logger.Debug("Source not found",
					zap.String("name", logicalName),
					zap.String("path", path))
				continue
			}

			if m, ok := sourceData.(map[string]interface{}); ok {
				params.Logger.Debug("Merging source",
					zap.String("name", logicalName),
					zap.String("path", path))
				mergeIntoRenderContextEnhanced(renderCtx, m, logicalName, params.Logger)
				sourcesMerged++
			}
		}
	} else if sourcesArray, ok := config["sources"].([]interface{}); ok {
		// Array format: direct paths
		for _, src := range sourcesArray {
			if srcStr, ok := src.(string); ok {
				sourceData := datahelpers.ExtractNestedField(params.CollectedData, srcStr)
				if sourceData == nil {
					continue
				}
				if m, ok := sourceData.(map[string]interface{}); ok {
					mergeIntoRenderContextEnhanced(renderCtx, m, srcStr, params.Logger)
					sourcesMerged++
				}
			}
		}
	} else {
		// Default sources
		defaultSources := []string{"input_data", "site_record", "style_collection", "reviewed_brief", "page_plan"}
		for _, source := range defaultSources {
			sourceData := datahelpers.ExtractNestedField(params.CollectedData, source)
			if sourceData == nil {
				// Try with input_data prefix
				sourceData = datahelpers.ExtractNestedField(params.CollectedData, "input_data."+source)
			}
			if sourceData == nil {
				continue
			}
			if m, ok := sourceData.(map[string]interface{}); ok {
				mergeIntoRenderContextEnhanced(renderCtx, m, source, params.Logger)
				sourcesMerged++
			}
		}
	}

	// Page identity. mergeIntoRenderContextEnhanced extracts a fixed allowlist of
	// branding fields (domain, colours, contact, a few image URLs) and drops
	// everything else, so the page record's own name never reached the context:
	// the step config passes "page": "input_data.current_page", and it was
	// thrown away. Every section component therefore saw an empty current_page
	// and could not vary per page, while the template data map advertised the
	// field as available. bugs_open/085.
	if renderCtx.CurrentPage == "" {
		renderCtx.CurrentPage = resolveCurrentPageName(params.CollectedData, config, params.Logger)
	}

	// Try to load navigation from DB if we have site_id
	if siteIDField, ok := config["site_id_field"].(string); ok && siteIDField != "" {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
		if siteIDStr != "" {
			if siteUUID, err := uuid.Parse(siteIDStr); err == nil && params.DB != nil {
				renderCtx.SiteID = siteUUID
				/*nav, _ := getNavigationFromDB(ctx, params.DB, siteUUID, "header", params.Logger)*/
				// NavFetchableOnly: this nav goes into renderCtx and is rendered into
				// page HTML, so it was a second live instance of bugs_open/049's
				// mechanism 2 alongside the chrome renderer.
				headerNav := GetNavItems(ctx, params.DB, siteUUID, []string{NavGroupPrimary}, NavFetchableOnly, 0, params.Logger)
				if len(headerNav) > 0 {
					renderCtx.NavItems = headerNav
				}
			}
		}
	}

	// Extract image URLs from deploy_image_asset output
	// (adds logo_deployed block between hero_deployed and fallback)
	// =========================================================================
	// A miss here is no longer silent: deployedImageURL reports a present-but-
	// unusable deploy result to agent_error_log, and stays quiet when no image was
	// deployed at all. bugs_open/236, commission item 2 — see
	// deployed_image_read_audit.go for why the durable row and not only a log line.
	if imageURL := deployedImageURL(ctx, params, "hero_deployed", "image_url", "hero_url", "build_render_context"); imageURL != "" {
		if renderCtx.ContentData == nil {
			renderCtx.ContentData = make(map[string]interface{})
		}
		renderCtx.ContentData["hero_url"] = imageURL
		params.Logger.Info("Set hero_url from hero_deployed.image_url",
			zap.String("url", imageURL))
	}

	if imageURL := deployedImageURL(ctx, params, "logo_deployed", "image_url", "logo_url", "build_render_context"); imageURL != "" {
		if renderCtx.ContentData == nil {
			renderCtx.ContentData = make(map[string]interface{})
		}
		renderCtx.ContentData["logo_url"] = imageURL
		renderCtx.LogoURL = imageURL
		params.Logger.Info("Set logo_url from logo_deployed.image_url",
			zap.String("url", imageURL))
	}

	// Also check for direct hero_url in collected_data (fallback)
	for _, field := range []string{"hero_url", "hero_home_url", "logo_url"} {
		if url := datahelpers.ExtractNestedFieldString(params.CollectedData, field); url != "" {
			if renderCtx.ContentData == nil {
				renderCtx.ContentData = make(map[string]interface{})
			}
			renderCtx.ContentData[field] = url
			if field == "logo_url" {
				renderCtx.LogoURL = url
			}
		}
	}

	params.Logger.Info("BuildRenderContextAction: Context built",
		zap.String("domain", renderCtx.Domain),
		zap.String("company_name", renderCtx.CompanyName),
		zap.String("logo_text", renderCtx.LogoText),
		zap.Int("nav_items", len(renderCtx.NavItems)),
		zap.Int("sources_merged", sourcesMerged),
	)

	// Return the context directly, not wrapped in "render_context" key
	// The workflow output_field already specifies where to store it
	// Adding metadata fields at same level
	result := renderCtxToMap(renderCtx)
	result["_sources_merged"] = sourcesMerged
	result["_built_at"] = time.Now().Format(time.RFC3339)

	return result, nil
}

// currentPageNameKeys are the keys under which a page record carries its own
// name, in the order they are trusted. Both are live: the page-content-writer's
// current_page (the only caller of build_render_context today) uses "name",
// while the rerender and page-build envelopes use "page_name" — and one
// observed shape carries both. Taken from the live payloads on 2026-07-27
// rather than from the struct, because the envelope is assembled by config.
var currentPageNameKeys = []string{"name", "page_name"}

// resolveCurrentPageName recovers the page's own name from whichever source the
// step config designates as the page, returning it without the ".html" suffix
// (the form the nav's active-page comparison and every existing CurrentPage
// producer use — see buildHeaderConfig).
//
// It reads the configured path rather than a hard-coded one so the fix follows
// the workflow rather than a single agent's spelling of it, and falls back to
// the conventional location when sources are absent or in array form.
//
// Every way of failing here degrades to an empty CurrentPage — which is today's
// behaviour, so it fails closed — but it does NOT fail quietly. Silence is the
// defect this function exists to fix; a second silent drop inside the fix would
// be the same bug wearing the fix's clothes. Each degradation logs the shape it
// actually saw, so the next unknown envelope costs one log line rather than
// another round of measuring rendered pages.
func resolveCurrentPageName(collectedData map[string]interface{}, config map[string]interface{}, logger *zap.Logger) string {
	const conventionalPath = "input_data.current_page"

	path := conventionalPath
	switch sources := config["sources"].(type) {
	case map[string]interface{}:
		if p, ok := sources["page"].(string); ok && p != "" {
			path = p
		} else {
			logger.Warn("resolveCurrentPageName: step config has a sources map with no usable \"page\" entry — falling back to the conventional path; current_page will be empty if the page is not there (bugs_open/085)",
				zap.String("fallback_path", conventionalPath),
				zap.Strings("source_names", datahelpers.GetMapKeys(sources)))
		}
	default:
		// The array form and an absent `sources` are both legitimate configs, so
		// this is not a warning — but it must not be silent either. Without it,
		// the one branch that takes the conventional path WITHOUT anyone having
		// declared it says nothing at all, which is the shape this whole function
		// exists to stop reproducing.
		logger.Info("resolveCurrentPageName: step config declares no sources map (absent, or the array form) — using the conventional page path",
			zap.String("path", conventionalPath),
			zap.String("sources_type", fmt.Sprintf("%T", config["sources"])))
	}

	page := datahelpers.ExtractNestedFieldMap(collectedData, path)
	if page == nil {
		logger.Warn("resolveCurrentPageName: no page object at the configured source path — every section component will see an empty current_page and cannot vary per page (bugs_open/085)",
			zap.String("path", path))
		return ""
	}
	for _, key := range currentPageNameKeys {
		if v, ok := page[key].(string); ok && v != "" {
			return strings.TrimSuffix(v, ".html")
		}
	}

	// A page object that carries none of the known name keys is a THIRD envelope
	// shape. Log the keys it does carry: that is the whole diagnosis for whoever
	// adds it to currentPageNameKeys, and it is the difference between this and
	// the silent allowlist drop that caused the bug.
	logger.Warn("resolveCurrentPageName: page object carries none of the known name keys — current_page will be empty; add the key it does use to currentPageNameKeys (bugs_open/085)",
		zap.String("path", path),
		zap.Strings("known_keys", currentPageNameKeys),
		zap.Strings("keys_present", datahelpers.GetMapKeys(page)))
	return ""
}

// mergeIntoRenderContextEnhanced extracts data from various source formats
func mergeIntoRenderContextEnhanced(ctx *RenderContext, data map[string]interface{}, sourceName string, logger *zap.Logger) {
	// =========================================================================
	// STEP 1: Unwrap .response wrapper if present (common in agent responses)
	// reviewed_brief has: {"response": {...actual data...}, "response_status": "complete"}
	// This recursively processes the unwrapped data first
	// =========================================================================
	if response, ok := data["response"].(map[string]interface{}); ok {
		// Check if this looks like a wrapped response (has response_status sibling)
		if _, hasStatus := data["response_status"]; hasStatus {
			logger.Debug("Unwrapping .response wrapper for source",
				zap.String("source", sourceName))
			// Recursively merge the unwrapped response data FIRST
			mergeIntoRenderContextEnhanced(ctx, response, sourceName+".response", logger)
			// Then continue to process the outer data for any additional fields
		}
	}

	// =========================================================================
	// STEP 2: Direct field extraction from current data level — DERIVED from
	// the struct's json tags (bugs_open/109). Every step-contract scalar is
	// accepted from data[tag] when non-empty; the exceptions live in
	// renderContextUnserialised / renderContextControlFields with reasons.
	// Before this, the hand-list here was the build-side allowlist that
	// silently dropped any field nobody remembered to add (bugs_open/085's
	// first drop point).
	// =========================================================================
	setRenderContextScalarsFromData(ctx, data)

	// Aliased inputs the tag derivation cannot know: reviewed_brief writes
	// contact_email / contact_phone, and they win over plain email/phone.
	if v, ok := data["contact_email"].(string); ok && v != "" {
		ctx.Email = v
	}
	if v, ok := data["contact_phone"].(string); ok && v != "" {
		ctx.Phone = v
	}

	// Company name doubles as the logo-text fallback.
	if ctx.LogoText == "" && ctx.CompanyName != "" {
		ctx.LogoText = ctx.CompanyName
	}

	// Extract image URLs from content_data or collected_data
	imageURLFields := []string{
		"hero_url",      // pageflow-builder uses this
		"hero_home_url", // multipage-website-builder uses this
		"hero_about_url",
		"hero_services_url",
		"logo_url",
	}

	for _, field := range imageURLFields {
		if v, ok := data[field].(string); ok && v != "" {
			if ctx.ContentData == nil {
				ctx.ContentData = make(map[string]interface{})
			}
			ctx.ContentData[field] = v
			logger.Debug("Extracted image URL",
				zap.String("field", field),
				zap.String("url", v))
		}
	}

	// =========================================================================
	// STEP 3: Check nested color_palette (from style_collection)
	// =========================================================================
	if palette, ok := data["color_palette"].(map[string]interface{}); ok {
		if v, ok := palette["primary"].(string); ok && v != "" {
			ctx.PrimaryColor = v
		}
		if v, ok := palette["secondary"].(string); ok && v != "" {
			ctx.SecondaryColor = v
		}
		if v, ok := palette["accent"].(string); ok && v != "" {
			ctx.AccentColor = v
		}
		if v, ok := palette["background"].(string); ok && v != "" {
			ctx.BackgroundColor = v
		}
		if v, ok := palette["text"].(string); ok && v != "" {
			ctx.TextColor = v
		}
	}

	// =========================================================================
	// STEP 4: Extract from nested structures (business_context, contact_info, brand)
	// =========================================================================

	// business_context (alternative structure)
	if brief, ok := data["business_context"].(map[string]interface{}); ok {
		if v, ok := brief["company_name"].(string); ok && v != "" {
			ctx.CompanyName = v
			if ctx.LogoText == "" {
				ctx.LogoText = v
			}
		}
		if v, ok := brief["tagline"].(string); ok && v != "" {
			ctx.Tagline = v
		}
		if v, ok := brief["industry"].(string); ok && v != "" {
			ctx.Industry = v
		}
	}

	// contact_info (nested contact structure)
	if contact, ok := data["contact_info"].(map[string]interface{}); ok {
		if v, ok := contact["email"].(string); ok && v != "" {
			ctx.Email = v
		}
		if v, ok := contact["phone"].(string); ok && v != "" {
			ctx.Phone = v
		}
	}

	// brand (nested brand/visual settings)
	if brand, ok := data["brand"].(map[string]interface{}); ok {
		if v, ok := brand["primary_color"].(string); ok && v != "" {
			ctx.PrimaryColor = v
		}
		if v, ok := brand["secondary_color"].(string); ok && v != "" {
			ctx.SecondaryColor = v
		}
		if v, ok := brand["tagline"].(string); ok && v != "" {
			ctx.Tagline = v
		}
	}

	// =========================================================================
	// STEP 5: Content generation context — covered by the derived extraction
	// in STEP 2 (tone, target_audience, industry are step-contract scalars).
	// =========================================================================

	// =========================================================================
	// STEP 6: Site/page identifiers
	// =========================================================================
	if v, ok := data["site_id"].(string); ok && v != "" {
		if siteUUID, err := uuid.Parse(v); err == nil {
			ctx.SiteID = siteUUID
		}
	}

	// =========================================================================
	// STEP 7: CTA settings — covered by the derived extraction in STEP 2.
	// =========================================================================

	// =========================================================================
	// STEP 8: Extract services array (for footer and services sections)
	// Services appear in reviewed_brief.response.services as []interface{}
	// Each service is {"name": "...", "description": "..."}
	//
	// ctx.Services is []string (just names)
	// ctx.ContentData["services"] is []interface{} (full objects for {{range .services}})
	// =========================================================================
	if services, ok := data["services"].([]interface{}); ok && len(services) > 0 && len(ctx.Services) == 0 {
		// Store full services in ContentData for template access via {{range .services}}
		// Only if not already populated (avoid duplicates from brief + brief.response)
		if ctx.ContentData == nil {
			ctx.ContentData = make(map[string]interface{})
		}
		// ctx.ContentData["services"] = services
		ctx.ContentData["services"] = normaliseToNameDescArray(services)

		// Also extract just the names to ctx.Services ([]string)
		for _, svc := range services {
			if svcMap, ok := svc.(map[string]interface{}); ok {
				if name, ok := svcMap["name"].(string); ok && name != "" {
					ctx.Services = append(ctx.Services, name)
				}
			}
		}

		logger.Info("Extracted services array",
			zap.String("source", sourceName),
			zap.Int("full_count", len(services)),
			zap.Int("names_count", len(ctx.Services)))
	}

	// =========================================================================
	// STEP 9: Extract navigation from db_sync source
	// db_sync contains: {"navigation": {"items": [{"label": "Home", "url": "/index.html"}, ...]}}
	// =========================================================================
	if sourceName == "db_sync" || sourceName == "db_sync.response" {
		if navigation, ok := data["navigation"].(map[string]interface{}); ok {
			if items, ok := navigation["items"].([]interface{}); ok {
				for _, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						label, _ := itemMap["label"].(string)
						url, _ := itemMap["url"].(string)
						if label != "" && url != "" {
							ctx.NavItems = append(ctx.NavItems, NavItem{
								Label: label,
								URL:   url,
							})
						}
					}
				}
				if len(ctx.NavItems) > 0 {
					logger.Info("Extracted navigation items from db_sync",
						zap.Int("count", len(ctx.NavItems)))
				}
			}
		}
	}

	// Handle site_record.content_data - recursively merge nested content_data
	if sourceName == "site_record" || sourceName == "site" {
		if contentData, ok := data["content_data"].(map[string]interface{}); ok {
			logger.Debug("Processing site_record.content_data",
				zap.Int("fields", len(contentData)))
			mergeIntoRenderContextEnhanced(ctx, contentData, "site_record.content_data", logger)
		}
	}

	// Handle services array (for services_html generation)
	if len(ctx.Services) == 0 {
		if services, ok := data["services"].([]interface{}); ok && len(services) > 0 {
			for _, svc := range services {
				if name, ok := svc.(string); ok && name != "" {
					ctx.Services = append(ctx.Services, name)
				} else if svcMap, ok := svc.(map[string]interface{}); ok {
					if name, ok := svcMap["name"].(string); ok && name != "" {
						ctx.Services = append(ctx.Services, name)
					}
				}
			}
		}
	}

	// =========================================================================
	// STEP 10: Log final state for debugging
	// =========================================================================
	logger.Info("Merged source into render context",
		zap.String("source", sourceName),
		zap.String("company_name", ctx.CompanyName),
		zap.String("domain", ctx.Domain),
		zap.String("tagline", ctx.Tagline),
		zap.String("email", ctx.Email),
		zap.Int("nav_items", len(ctx.NavItems)),
		zap.Int("services", len(ctx.Services)))
}

// renderCtxToMap converts RenderContext to map for template substitution
// renderContextUnserialised names the RenderContext string fields that
// deliberately do NOT cross the step boundary, each with the reason.
//
// bugs_open/109. This map is the entire point of the change that introduced it.
// Before, renderCtxToMap held a hand-written literal of ~18 keys, and a field
// that was missing from it was indistinguishable from a field nobody had
// thought about: "empty" is a legal value for every one of these, so a dropped
// field has no error surface at all. That is how `current_page` was advertised
// to every component author while arriving empty forever (bugs_open/085).
//
// Now the serialised set is DERIVED from the struct's json tags, so the default
// for a new field is "serialised", and omitting one is something you have to do
// on purpose, here, in writing. An entry with no reason is a bug in this map.
//
// Shrinking this map is the goal; each entry needs its producer decided first,
// which is a behaviour change and does not belong in the same edit as the
// mechanism fix.
var renderContextUnserialised = map[string]string{
	"logo_url": "latent, not live: both BuildRenderContextAction and " +
		"mergeIntoRenderContextEnhanced write ContentData[\"logo_url\"] alongside the " +
		"struct field, and ContentData is merged at the end of this function, so the " +
		"value arrives by that route. Serialising the struct field too would be " +
		"harmless but redundant.",
	"theme_css": "genuinely dropped, tracked in bugs_open/109. Written by " +
		"assemble_from_library.go:121 on the assembly path, which renders directly " +
		"rather than crossing a step boundary. Serialising it would carry one page's " +
		"stylesheet into the next step's context; the producer has to be decided first.",
	"title": "genuinely dropped, tracked in bugs_open/109. Written per-page by " +
		"rerender_pages_actions.go:191 and multipage_actions.go:94. Serialising it " +
		"from a SITE-level context would bleed one title onto every page — the " +
		"opposite of the per-page behaviour current_page needed — so this needs its " +
		"producer decided, not just a slot.",
	"description": "genuinely dropped, tracked in bugs_open/109. Same shape and " +
		"same hazard as title: written per-page, would bleed if serialised from a " +
		"site-level context.",
}

// renderContextControlFields are RenderContext fields that steer the render
// machinery rather than carry template data. They are excluded from every
// derived map: never advertised to templates, never settable from source or
// collected data — a control switch that arbitrary content could flip is a
// different bug than a dropped field.
// ⚠ DELIBERATELY EMPTY since 2026-08-21, and that is a statement, not an
// oversight. Its only entry was `schema_mode`, described here as a
// "validation-strictness switch read by RenderComponent" — by then read by
// NOTHING: its consumer, RenderTemplateWithValidation, went with the regex
// fallback in bugs_closed/260, and the field itself is deleted.
//
// The control field that replaced it, RenderContext.InputSchema
// (bugs_open/342), is NOT listed here and must not be: this map excludes
// STRING-typed scalars from the step contract, and InputSchema is a map, so
// reflection over string fields never sees it. It is excluded STRUCTURALLY,
// which is stronger than an entry — and render_context_derivation_test.go's
// TestInputSchemaNeverReachesTemplatesOrStruct is what fails if that changes.
//
// A future STRING-typed control field does belong here, with its reason.
var renderContextControlFields = map[string]string{}

// renderContextStepContractExcluded says whether a scalar key stays out of the
// step-boundary contract — the set that renderCtxToMap emits and
// setRenderContextScalarsFromData accepts. Omissions live in
// renderContextUnserialised (template fields that must not cross the boundary,
// with reasons) and renderContextControlFields (machinery fields).
func renderContextStepContractExcluded(key string) bool {
	if _, control := renderContextControlFields[key]; control {
		return true
	}
	_, unserialised := renderContextUnserialised[key]
	return unserialised
}

// renderContextStepContractRenames maps a TEMPLATE key (the json tag) to the
// DIFFERENT name it carries across the STEP BOUNDARY — the key renderCtxToMap
// writes into collected_data and setRenderContextScalarsFromData reads back.
// Every other scalar crosses under its tag; an entry here is a deliberate
// split between the two namespaces, and needs its reason.
//
// RFC_029 §10.13 step 4 (staged_component_build lane). The tag is the single
// declaration of a field's template name (bugs_open/109), which made it the
// step-boundary name too — and for ONE field those two namespaces disagree
// about what the key means:
//
//   - In collected_data, `current_page` is the page RECORD by fleet convention
//     (input_data.current_page; five live agents request it as an input field;
//     page-content-writer's generate_content prompt reads
//     {{.current_page.title}}).
//   - In component template data, `current_page` is the page's NAME STRING
//     (the nav/active-page field; the `structuralFields` list in this file;
//     bugs_open/085's fix).
//
// build_render_context's output is filed into collected_data, so its string
// `current_page` sat one level below the page object of the same name — and
// the resolver's whole-tree search, asked for `current_page`, collected both
// and recorded a conflict on EVERY page-content-writer run (23 rows in the
// four hours after v1.0.1315; 100% of the surviving class). The winner was
// always the object, pinned by bugs_open/306's declared tie-break, so no page
// was wrong — but a permanent false-positive population blocks RFC_029 §10.13
// step 5 (conflicts → refusal) for ever: refusing would strip the page object
// from generate_content in every run.
//
// The fix is one name for the object and another for the string, at the
// boundary where they met. The template contract is NOT renamed: there the key
// is unambiguous (one reader fleet-wide, `evidence-chart`), it is advertised to
// component authors, and renaming it would have to be coordinated with a live
// template edit across a roll. The step-boundary name is read by exactly one
// Go site outside this contract (render_content_envelope_guard.go's identity
// chain, updated with it) and by no live agent definition (0 of the active
// rows mention render_context.current_page, measured 2026-08-19).
//
// What this does NOT change (council round 1, bug_historian): the resolver's
// whole-tree search stays generic — depth, then declared rank, then
// reflect.DeepEqual across candidates — so the NEXT producer pair that files two
// types under one key will reproduce this exact shape, recorded to
// agent_error_log and never blocking, until someone reads the candidate-set
// query (staged_component_build RUNBOOK, "step 4's done-condition") and repeats
// this one-rename-at-the-producer move. That is the owner-ruled sequencing
// (RFC_029 §10.13 step 4 → step 5 flips conflicts to refusal, which is what
// makes the next collision LOUD). A second entry here needs its producer read
// first; the map is not a place to park collisions nobody has traced.
var renderContextStepContractRenames = map[string]string{
	"current_page": "current_page_name",
}

// renderContextStepContractKey returns the key under which a template-tag
// scalar crosses the step boundary: its rename if it has one, else the tag.
func renderContextStepContractKey(templateKey string) string {
	if renamed, ok := renderContextStepContractRenames[templateKey]; ok {
		return renamed
	}
	return templateKey
}

// setRenderContextScalarsFromData is the write-side twin of
// renderContextScalarFields (bugs_open/109): every tagged string field in the
// step contract is set from data[tag] when data carries a non-empty string.
// Build and restore both use it, so what the serialiser emits is exactly what
// they accept — one contract, three call sites, no hand-list to forget a field
// in. A non-string value under a contract key (e.g. current_page as an object
// on some envelopes) fails the type assertion and is left to the caller's
// structured handling.
//
// A renamed key (renderContextStepContractRenames) is read ONLY under its
// step-boundary name. The read-side tolerance that also accepted the TEMPLATE
// name as a string fallback (for trees written before the rename rolled) was
// retired on 2026-08-21, on grounds measured that day — NOT the retention
// argument the original comment here carried, which was unsound on both
// counts (orchestration_states holds rows back to 2026-07-19, and stored
// page_components content_data never expires at all):
//   - zero non-terminal orchestrations created before the rename roll
//     (v1.0.1317, pods up 2026-08-19 22:26Z) remained, checked against the
//     full status vocabulary — so no restorable tree written by the old
//     binary exists;
//   - the rerender base map writes the step name fresh from its pageName
//     argument (buildRerenderBaseData), and every live page_components row
//     whose stored content_data still carries the old key as a string agreed
//     exactly with its page's own name (18 of 18 across 11 sites, measured
//     at retirement) — the fallback's only remaining input supplied a value
//     the base already had.
//
// Retiring it also closes the door it held open: a stale stored string can no
// longer clobber the fresh page identity in the struct field between the base
// merge and the resolved-data merge (bugs_closed/085's shape). The old spelling
// is simply not part of the read contract — a string under it is ignored the
// same way the page RECORD always was.
func setRenderContextScalarsFromData(ctx *RenderContext, data map[string]interface{}) {
	v := reflect.ValueOf(ctx).Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type.Kind() != reflect.String {
			continue
		}
		key := strings.Split(f.Tag.Get("json"), ",")[0]
		if key == "" || key == "-" || renderContextStepContractExcluded(key) {
			continue
		}
		if s, ok := data[renderContextStepContractKey(key)].(string); ok && s != "" {
			v.Field(i).SetString(s)
		}
	}
}

// renderContextScalarFields returns the string-valued fields of RenderContext
// keyed by their json tag — the tags being the single declaration of what a
// field is called when it reaches a template (bugs_open/109).
//
// Deriving beats listing here because the failure this replaces was silent: a
// field left out of a hand-written map produces an empty string at the
// template, which is a legal value, so nothing anywhere reports a problem. With
// derivation the omission cannot be expressed by accident — a new field is
// serialised unless someone adds it to renderContextUnserialised with a reason.
//
// Only string fields are derived. Composites (NavItems, Services, ContentData,
// SiteID) need shaping rather than copying and are handled explicitly below.
func renderContextScalarFields(ctx *RenderContext) map[string]string {
	out := make(map[string]string, 24)
	v := reflect.ValueOf(*ctx)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type.Kind() != reflect.String {
			continue
		}
		key := strings.Split(f.Tag.Get("json"), ",")[0]
		if key == "" || key == "-" {
			continue
		}
		out[key] = v.Field(i).String()
	}
	return out
}

func renderCtxToMap(ctx *RenderContext) map[string]interface{} {
	// The scalar half of the contract is derived from the struct declaration,
	// not listed by hand — see renderContextScalarFields and
	// renderContextUnserialised (bugs_open/109).
	result := make(map[string]interface{}, 32)
	for key, value := range renderContextScalarFields(ctx) {
		if renderContextStepContractExcluded(key) {
			continue
		}
		// Under the step-boundary name, which differs from the template name
		// for the entries of renderContextStepContractRenames (current_page →
		// current_page_name: the page object and the page's name must not share
		// a key in one collected_data tree).
		result[renderContextStepContractKey(key)] = value
	}

	if ctx.SiteID != uuid.Nil {
		result["site_id"] = ctx.SiteID.String()
	}

	// =========================================================================
	// Generate nav_items and nav_items_html from NavItems
	// Templates may use either:
	//   {{range .nav_items}} for iteration
	//   {{.nav_items_html}} for pre-rendered HTML
	// =========================================================================
	if len(ctx.NavItems) > 0 {
		// nav_items as array (for {{range .nav_items}})
		navItems := make([]map[string]interface{}, len(ctx.NavItems))
		for i, item := range ctx.NavItems {
			navItems[i] = map[string]interface{}{
				"label":     item.Label,
				"url":       item.URL,
				"is_active": item.IsActive,
			}
		}
		result["nav_items"] = navItems

		// nav_items_html as pre-rendered string (for {{.nav_items_html}})
		// Format: <li><a href="/page.html">Label</a></li>
		var htmlParts []string
		for _, item := range ctx.NavItems {
			activeClass := ""
			if item.IsActive {
				activeClass = ` class="active"`
			}
			htmlParts = append(htmlParts, fmt.Sprintf(
				`<li><a href="%s"%s>%s</a></li>`,
				item.URL, activeClass, item.Label,
			))
		}
		result["nav_items_html"] = strings.Join(htmlParts, "\n                ")
	} else {
		// Ensure empty values don't render as "<no value>"
		result["nav_items"] = []map[string]interface{}{}
		result["nav_items_html"] = ""
	}

	// =========================================================================
	// Generate services list HTML for footer
	// Templates use {{.services_html}} for footer services list
	// ctx.Services is []string (service names extracted from reviewed_brief)
	// =========================================================================
	if len(ctx.Services) > 0 {
		result["services"] = ctx.Services

		// services_html as pre-rendered string
		var servicesParts []string
		for _, serviceName := range ctx.Services {
			if serviceName != "" {
				servicesParts = append(servicesParts, fmt.Sprintf(
					`<li><a href="/services.html">%s</a></li>`,
					serviceName,
				))
			}
		}
		result["services_html"] = strings.Join(servicesParts, "\n                ")
	} else {
		result["services"] = []string{}
		result["services_html"] = ""
	}

	// Add image URLs if present in ContentData
	if ctx.ContentData != nil {
		for _, field := range []string{"hero_url", "hero_home_url", "hero_about_url", "logo_url"} {
			if url, ok := ctx.ContentData[field].(string); ok && url != "" {
				result[field] = url
			}
		}
	}

	// =========================================================================
	// Merge ContentData fields for additional template access
	// This includes full service objects for {{range .services}} iteration
	// =========================================================================
	if ctx.ContentData != nil {
		for key, value := range ctx.ContentData {
			// Don't overwrite explicit fields
			if _, exists := result[key]; !exists {
				// A renamed template name never crosses the boundary under its
				// old spelling, whatever ContentData happens to carry — the
				// invariant the resolver relies on is "the step output has no
				// `current_page`", not "the struct field is renamed".
				if _, renamed := renderContextStepContractRenames[key]; renamed {
					continue
				}
				result[key] = value
			}
		}
	}

	return result
}

func mergeIntoRenderContext(ctx *RenderContext, data map[string]interface{}) {
	// Scalar restore is DERIVED from the struct's json tags — the same step
	// contract the serialiser emits and the build map accepts (bugs_open/109).
	// The struct restore matters because of the two render paths: the
	// ContentData catch-all below is enough for html/template
	// (contextToInterfaceMap merges ContentData over the base map) but NOT for
	// the regex fallback (contextToMap skips any key the base map already
	// holds, and the base map reads the struct). Before derivation only ~10
	// fields were restored here, so a fallback render saw default colours and
	// empty cta/industry/year where the main path saw real values —
	// bugs_open/085's exact shape, times twelve.
	setRenderContextScalarsFromData(ctx, data)
	if ctx.LogoText == "" && ctx.CompanyName != "" {
		ctx.LogoText = ctx.CompanyName
	}

	// Check nested color_palette
	if palette, ok := data["color_palette"].(map[string]interface{}); ok {
		if v, ok := palette["primary"].(string); ok {
			ctx.PrimaryColor = v
		}
		if v, ok := palette["secondary"].(string); ok {
			ctx.SecondaryColor = v
		}
		if v, ok := palette["accent"].(string); ok {
			ctx.AccentColor = v
		}
	}

	// Capture ALL fields into ContentData for template access
	if ctx.ContentData == nil {
		ctx.ContentData = make(map[string]interface{})
	}
	for key, value := range data {
		ctx.ContentData[key] = value
	}
}

// DEPRECATED
func convertNavigationItems(items []NavigationItem) []NavItem {
	result := make([]NavItem, len(items))
	for i, item := range items {
		result[i] = NavItem{
			Label: item.Label,
			URL:   item.URL,
		}
	}
	return result
}

// RenderComponentAction renders a single component template with context
// Config:
//   - component_function: function name to look up (e.g., "hero-banner")
//   - component_id: explicit component UUID (alternative to function)
//   - component_from: path to object containing function/id (e.g., "current_section")
//   - context_field: path to render context in collected_data
//   - content_field: path to additional content data
//   - content_from: alias for content_field (for consistency)
//   - slot_name_from: OPTIONAL path to the section's own slot name (e.g.
//     "current_section.name"). See the emit site below (bugs_open/189).
func RenderComponentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RenderComponentAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required for component rendering")
	}

	// Get component - now with support for component_from indirection
	var comp *Component
	var err error
	var componentFunction string
	var componentID string

	// Priority 1: Direct component_id in config
	if compID, ok := config["component_id"].(string); ok && compID != "" {
		componentID = compID
	}

	// Priority 2: Direct component_function in config
	if compFunc, ok := config["component_function"].(string); ok && compFunc != "" {
		componentFunction = compFunc
	}

	// Priority 3: Extract from component_from field (indirection)
	if componentID == "" && componentFunction == "" {
		if componentFrom, ok := config["component_from"].(string); ok && componentFrom != "" {
			componentData := datahelpers.ExtractNestedField(params.CollectedData, componentFrom)
			if componentData != nil {
				// Case 1: component_from points to a map (e.g., "current_section")
				if compMap, ok := componentData.(map[string]interface{}); ok {
					// Try to get function first (most common)
					if fn, ok := compMap["function"].(string); ok && fn != "" {
						componentFunction = fn
						params.Logger.Debug("RenderComponentAction: Extracted function from component_from",
							zap.String("component_from", componentFrom),
							zap.String("function", componentFunction),
						)
					}
					// Also check for component_function key
					if fn, ok := compMap["component_function"].(string); ok && fn != "" {
						componentFunction = fn
					}
					// Check for id/component_id
					if id, ok := compMap["id"].(string); ok && id != "" {
						componentID = id
					}
					if id, ok := compMap["component_id"].(string); ok && id != "" {
						componentID = id
					}
					// Check for name as fallback (some components use name as function)
					if componentFunction == "" {
						if name, ok := compMap["name"].(string); ok && name != "" {
							componentFunction = name
							params.Logger.Debug("RenderComponentAction: Using name as function fallback",
								zap.String("name", name),
							)
						}
					}
				} else if compStr, ok := componentData.(string); ok && compStr != "" {
					// Case 2: component_from points directly to a string value
					// e.g., "current_section.function" resolves to "hero"
					componentFunction = compStr
					params.Logger.Debug("RenderComponentAction: Extracted function string directly from component_from",
						zap.String("component_from", componentFrom),
						zap.String("function", componentFunction),
					)
				}
			} else {
				params.Logger.Warn("RenderComponentAction: component_from field not found",
					zap.String("component_from", componentFrom),
				)
			}
		}
	}

	// Enforce naming contract: normalize to kebab-case before lookup
	if componentFunction != "" {
		normalized := NormalizeComponentFunction(componentFunction)
		if normalized != componentFunction {
			params.Logger.Info("RenderComponentAction: Normalized component function to kebab-case",
				zap.String("original", componentFunction),
				zap.String("normalized", normalized),
			)
			componentFunction = normalized
		}
	}

	// Now resolve the component
	if componentID != "" {
		compUUID, parseErr := uuid.Parse(componentID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid component_id: %w", parseErr)
		}
		comp, err = GetComponentByID(ctx, params.DB, compUUID, params.Logger)
	} else if componentFunction != "" {
		comp, err = GetComponentWithFallback(ctx, params.DB, componentFunction, params.Logger)
	} else {
		// Log available info for debugging
		params.Logger.Error("RenderComponentAction: No component identifier found",
			zap.Any("config_keys", datahelpers.GetMapKeys(config)),
			zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
		)
		return nil, fmt.Errorf("component_function, component_id, or component_from required")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get component '%s': %w", componentFunction, err)
	}

	// Get render context
	contextField := "render_context"
	if cf, ok := config["context_field"].(string); ok && cf != "" {
		contextField = cf
	}
	// Also support context_from as an alias
	if cf, ok := config["context_from"].(string); ok && cf != "" {
		contextField = cf
	}

	renderCtxData := datahelpers.ExtractNestedField(params.CollectedData, contextField)
	renderCtx := &RenderContext{}
	if m, ok := renderCtxData.(map[string]interface{}); ok {
		mergeIntoRenderContext(renderCtx, m)
	}

	// Merge additional content if specified (support both content_field and content_from)
	contentField := ""
	if cf, ok := config["content_field"].(string); ok && cf != "" {
		contentField = cf
	}
	if cf, ok := config["content_from"].(string); ok && cf != "" {
		contentField = cf
	}

	var sectionContentData map[string]interface{}
	var strippedMarkdownFields []string

	if contentField != "" {
		// Try to extract content with fallback paths
		// LLM responses sometimes have .result wrapper, sometimes not
		contentData := extractContentWithFallbacks(params.CollectedData, contentField, params.Logger)
		if contentData != nil {
			// The resolver returns whatever map it found, and when the LLM step
			// fell to the text path that map is the transport envelope itself —
			// `result` is a STRING, so the ".result" path misses and the bare
			// step key matches instead. Decode it when that is provably
			// lossless, refuse the render when it is not: the same payload, the
			// same policy and the same normaliser as the storage seam, one step
			// earlier so the section renders its real content rather than a
			// blank the storage guard can only repair after the fact. A no-op
			// for every non-envelope map. See render_content_envelope_guard.go
			// (bugs_open/199) — including why this sits at the caller rather
			// than inside extractContentWithFallbacks, which leaks by two
			// separate branches.
			contentData, err = normalizeRenderContentEnvelope(ctx, params, comp, contentField, contentData)
			if err != nil {
				return nil, err
			}

			// Safety net: reconcile any array-item keys the LLM invented
			// (e.g. title/body) against the keys the component template reads
			// (e.g. name/description), before the content reaches the template
			// or is persisted. Expected keys are sourced from the component's
			// own input_schema (reloaded fresh above from the component store),
			// which is the authoritative contract the html_template is built
			// against — so reconciliation does not depend on section-plan
			// freshness or on the prompt. Scoped to source:"llm" array fields so
			// reach matches the writer loop; a no-op on the template-only path
			// (content is render_context, whose keys don't match the schema's
			// llm array fields).
			if comp != nil && len(comp.InputSchema) > 0 {
				reconcileGeneratedItemKeys(contentData, expectedItemFieldsFromComponentSchema(comp.InputSchema), componentFunction, params.Logger)
			}

			// bugs_open/184: strip literal markdown (**bold**, `code spans`,
			// # headings, [links](url)) from plain-text LLM fields at birth, so
			// BOTH surfaces — the render context and the persisted content_data
			// captured below — are built from clean values. text/template
			// escapes nothing, so the syntax would reach the visitor verbatim.
			// Strip-only (never inserts markup into the unescaping pipe);
			// markup-bearing values keep their backticks/brackets (the check's
			// own suppression rule); only freshly-generated values are touched
			// here — nothing pre-existing is altered on this seam. Opt-in per
			// step, default OFF: enabled by migration 474 on
			// page-content-writer's render_section. Strips are logged AND
			// surfaced on the result (pod logs rotate in minutes;
			// collected_data is the durable record — council 060bcc0a round 2).
			if on, _ := config["strip_literal_markdown"].(bool); on {
				if changed := datahelpers.StripLiteralMarkdownFromContentData(contentData); len(changed) > 0 {
					strippedMarkdownFields = changed
					params.Logger.Info("RenderComponentAction: stripped literal markdown from LLM content",
						zap.Strings("fields", changed),
						zap.String("component", componentFunction))
				}
			}
			sectionContentData = contentData // ← capture before merge
			params.Logger.Info("RenderComponentAction: Merging content data",
				zap.String("content_field", contentField),
				zap.Int("field_count", len(contentData)),
				zap.Any("keys", datahelpers.GetMapKeys(contentData)))
			mergeIntoRenderContext(renderCtx, contentData)
		} else {
			params.Logger.Warn("RenderComponentAction: No content data found at any path",
				zap.String("content_field", contentField),
				zap.Any("available_top_keys", datahelpers.GetMapKeys(params.CollectedData)))
		}
	}

	// Step 3: optional merge_with — overlay pre-resolved data on top of the
	// LLM/content output. Used by page-content-writer's loop with
	// `merge_with: current_section.resolved_data` so query-resolved items,
	// static fallback values, and other authoritative data land in both the
	// rendered HTML AND the persisted content_data. The merge happens AFTER
	// the content_from block so resolved_data wins on conflicts — by design,
	// because it's database-derived and authoritative; the LLM should never
	// be writing items/urls/labels that the resolver already produced.
	if mw, ok := config["merge_with"].(string); ok && mw != "" {
		mergeData := datahelpers.ExtractNestedField(params.CollectedData, mw)
		if mergeMap, ok := mergeData.(map[string]interface{}); ok && len(mergeMap) > 0 {
			params.Logger.Info("RenderComponentAction: Merging resolved data",
				zap.String("merge_with", mw),
				zap.Int("merge_field_count", len(mergeMap)),
				zap.Any("merge_keys", datahelpers.GetMapKeys(mergeMap)))
			if sectionContentData == nil {
				sectionContentData = make(map[string]interface{})
			}
			// bugs_open/184 (canary finding, 2026-08-19): the merge overlay
			// wins over the LLM content — so stripping the LLM map above is
			// not enough when the RESOLVED source is dirty (measured:
			// content_feed_items.source_summary carries markdown in ~700
			// rows, re-poisoning news-listing items on every render). Same
			// strip on the overlay, before it lands in both surfaces.
			//
			// FLAG-ONLY HERE, BY DESIGN — not the rerender path's double gate
			// (council 060bcc0a r5, editquality HIGH: the r5 rationale said
			// "same double gate" and was wrong; the sketch was right). The
			// reason gate (shouldStripLiteralMarkdown's spec.reason ==
			// "literal_markdown") scopes a REPAIR to the work item that
			// dispatched it. This is the WRITER seam: a section being born,
			// no work item, no spec, no reason to gate on — so the step flag
			// is the whole gate, exactly as it is for the LLM-map strip a few
			// lines above, which rounds 1-4 approved. Who reaches this branch
			// with merge_with, measured live 2026-08-19 (every step, any
			// depth): page-content-writer's render_section (flag ON, migration
			// 474) and render_from_template (flag UNSET → this branch is a
			// no-op there; the news resolver's own producer-side strip is what
			// covers that template-only path). No other live agent runs
			// render_component with merge_with.
			//
			// In-place on the CollectedData map, deliberately — the cleaned
			// values become canonical for later steps, the same aliasing
			// contract save_sections_content_data_links.go documents.
			if on, _ := config["strip_literal_markdown"].(bool); on {
				if changed := datahelpers.StripLiteralMarkdownFromContentData(mergeMap); len(changed) > 0 {
					strippedMarkdownFields = append(strippedMarkdownFields, changed...)
					params.Logger.Info("RenderComponentAction: stripped literal markdown from merge_with overlay",
						zap.Strings("fields", changed),
						zap.String("component", componentFunction))
				}
			}
			// Overlay merge data onto section content data so it lands in both
			// the render context AND the persisted content_data output.
			// Last write wins → resolved_data overrides LLM duplicates.
			for k, v := range mergeMap {
				sectionContentData[k] = v
			}
			mergeIntoRenderContext(renderCtx, mergeMap)
		} else if mergeData != nil {
			params.Logger.Warn("RenderComponentAction: merge_with did not resolve to a map",
				zap.String("merge_with", mw),
				zap.String("type", fmt.Sprintf("%T", mergeData)))
		}
		// If mergeData is nil, that's fine — the path simply wasn't populated
		// for this section (e.g. an all-LLM section with no resolved_data).
	}

	// Per-instance element-id token. This path renders ONE section, but it is
	// always a loop iteration, and loop expansion puts this step's index and the
	// pass's items where we can read them — so the occurrence is counted from the
	// sections already rendered in this pass rather than assumed to be 0.
	// Derivation, fallbacks and the honest blind spots:
	// component_instance_occurrence.go (bugs_closed/283, RFC_032 step 3).
	// Outside a loop it binds occurrence 0, exactly as this call site did before.
	//
	// The retired `renderCtx.ContentData["ComponentID"]` binding stood here until
	// 2026-08-24. It was the last of RFC_032 §8's three; a component-wide value
	// cannot namespace a per-instance id, which is what {{.InstanceID}} replaced
	// it with. Census before deleting: 0 active AND 0 inactive templates spell
	// {{.ComponentID}}, against a control of 140 spelling {{.InstanceID}}.
	DeriveAndBindInstanceToken(ctx, params.DB, renderCtx, comp.Function,
		PlacementFromLoopStep(config, params.CollectedData), params.Logger)

	// Fail loud rather than ship a silently-empty section. If the component's
	// schema marks a content field required (source:"llm") and it never arrived
	// — the LLM was truncated, or its response was unparseable and fell back to
	// a raw-text envelope that carries no such field — then missingkey=zero
	// would render {{.field}} as empty, page assembly would drop the visually
	// empty section, and the article would silently vanish. That is exactly the
	// mechanism that blanked 9 live article bodies. Refuse the render instead so
	// the step fails and the good content is left in place (the content
	// regression guard in save_page_sections blocks the overwrite).
	if len(comp.InputSchema) > 0 {
		datahelpers.WarnIfLegacyDialect(comp.InputSchema, params.Logger, "render-gate", comp.Function)
		if missing := missingRequiredLLMFields(comp.InputSchema, renderCtx.ContentData); len(missing) > 0 {
			params.Logger.Error("RenderComponentAction: required content field(s) missing — refusing to render an empty section",
				zap.String("component_function", comp.Function),
				zap.String("component_name", comp.Name),
				zap.Strings("missing_fields", missing),
			)
			return nil, fmt.Errorf(
				"component %q is missing required content field(s) %v — refusing to render an empty section "+
					"(likely LLM truncation or an unparseable response); leaving existing content untouched",
				comp.Function, missing)
		}

		// The TYPE half of the same gate (bugs_open/260). The presence check
		// above catches a field that never arrived; this one catches a field
		// that arrived as the wrong SHAPE — a sentence where the schema
		// declares a list of objects — which is what every one of the 26
		// recorded render failures actually was.
		//
		// OPT-IN, unsafe default OFF (mistyped_llm_fields_gate.go): unlike the
		// presence check, a mistyped field the template never references
		// renders fine today, so refusing unconditionally would be new
		// authority over content that currently ships. The seam's hard error
		// remains the complete detector; this is the early, named one.
		if refuseMistypedLLMFields(config) {
			if violations := datahelpers.ContentTypeViolations(comp.InputSchema, renderCtx.ContentData); len(violations) > 0 {
				params.Logger.Error("RenderComponentAction: content field(s) contradict the component's declared types — refusing to render",
					zap.String("component_function", comp.Function),
					zap.String("component_name", comp.Name),
					zap.String("type_violations", datahelpers.DescribeTypeViolations(violations)),
				)
				return nil, fmt.Errorf(
					"component %q: content does not match the declared field type(s) — %s; refusing to render (bugs_open/260)",
					comp.Function, datahelpers.DescribeTypeViolations(violations))
			}
		}
	}

	// Render template.
	//
	// The reporting form, not RenderTemplate: the two extra return values name
	// which bare placeholders rendered empty and which of those sat inside an
	// href=/src=. RenderTemplate discards both (`out, _, _ :=`), which is how
	// bugs_open/238 shipped five <img src=""> to a live homepage while this very
	// call had the field names in hand. See dead_url_guard.go for why the guard
	// refuses rather than dropping, and why it is opt-in with the unsafe default.
	// bugs_open/342: hand the seam the contract so it can name an ABSENT
	// required field for every caller, not just the two that pre-check.
	renderCtx.InputSchema = comp.InputSchema

	// ── Composite gate (features_open/035 P1, direction 1) ──────────────────
	// A component whose schema declares `slots` is a COMPOSITE: its template
	// references {{.slots.*}}, which are filled by the hierarchy walk from child
	// rows. This action has no children — it renders a component DEFINITION plus a
	// content map and RETURNS the html — so under missingkey=zero every slot would
	// resolve to the empty string and the caller would persist a parent-shaped
	// section with nothing in it, reporting success.
	//
	// NOTE THE PREDICATE, because it is not the one used on the edit path and the
	// difference is the point. apply_section_edit holds a stored ROW, so it asks
	// "does this row have children" (hierarchyChildrenOf). This path holds no row
	// at all, so the only honest question is "does this COMPONENT declare slots" —
	// and that is the stronger guard, because it fires before any child row exists
	// to be counted. My council submission conflated the two; they are different
	// questions about different objects.
	//
	// Refuse rather than render: unlike the section paths there is no stored HTML
	// here to carry, so the only alternatives are an empty section or an error, and
	// bugs_open/260's rule is to fail rather than stitch. Unreachable today —
	// 0 of 386 content_components declare a slots block (2026-08-31).
	if len(hierarchySlotsFromSchema(comp.InputSchema)) > 0 {
		params.Logger.Error("RenderComponentAction: refusing to render a COMPOSITE component with no children (035 P1)",
			zap.String("component_function", comp.Function),
			zap.String("component_name", comp.Name),
		)
		return nil, fmt.Errorf(
			"component %q declares slots and is composite: this path renders a definition without child rows, so every {{.slots.*}} would resolve empty; render the page instead (features_open/035 P1)",
			comp.Function)
	}

	rendered, _, deadURLFields, renderErr := RenderTemplate(comp.HTMLTemplate, renderCtx, params.Logger)

	// bugs_open/260: the seam no longer invents output it could not execute, so
	// a render error arrives here for the first time. Fail the step, in the same
	// shape as the two refusals either side of this line — the parked work item
	// then carries the CAUSE ("steps[2].branches: declared array (items:
	// object), got string") instead of 20 capped symptoms from a downstream
	// regex, which is what the 24 items sitting at needs_human_review carry.
	//
	// The type report is an ENRICHER here, not a gate: it runs unconditionally
	// because the render has ALREADY failed, so it can refuse nothing that
	// works. Only the PRE-render refusal above is opt-in.
	if renderErr != nil {
		violations := datahelpers.ContentTypeViolations(comp.InputSchema, renderCtx.ContentData)
		diagnosis := datahelpers.DescribeTypeViolations(violations)
		params.Logger.Error("RenderComponentAction: component template failed to execute — refusing to ship output that was not rendered",
			zap.String("component_function", comp.Function),
			zap.String("component_name", comp.Name),
			zap.Error(renderErr),
			zap.String("type_violations", diagnosis),
		)
		if diagnosis != "" {
			return nil, fmt.Errorf("component %q failed to render: %w — %s (bugs_open/260)",
				comp.Function, renderErr, diagnosis)
		}
		return nil, fmt.Errorf("component %q failed to render: %w (bugs_open/260)", comp.Function, renderErr)
	}

	if shouldRefuseDeadURLControls(config, deadURLFields, comp.HTMLTemplate) {
		// File the human record BEFORE refusing: the refusal fails the step, and
		// a signal that only exists inside a failed step is a signal nobody reads.
		// Identity is best-effort — see the emit's own guard — because a missing
		// site_id must not turn a refusal into a silent success.
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
		siteID, _ := uuid.Parse(siteIDStr)
		pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, "page_record.name")
		slot := comp.Function
		if slotFrom, ok := config["slot_name_from"].(string); ok && slotFrom != "" {
			if s := datahelpers.ExtractNestedFieldString(params.CollectedData, slotFrom); s != "" {
				slot = s
			}
		}
		var componentID *uuid.UUID
		if cid, err := uuid.Parse(comp.ID); err == nil {
			componentID = &cid
		}
		emitSectionDeadControlItem(ctx, params.DB, siteID, componentID,
			pageName, slot, comp.Function, deadURLFields, true, params.Logger)

		params.Logger.Error("RenderComponentAction: URL attribute(s) would render empty — refusing to ship a dead control",
			zap.String("component_function", comp.Function),
			zap.String("page", pageName),
			zap.String("slot", slot),
			zap.Strings("dead_url_fields", deadURLFields),
		)
		return nil, fmt.Errorf(
			"component %q would ship dead URL control(s) %v — src=/href= attributes that render empty "+
				"(bugs_open/238); refusing so the stored section and the live page stay intact",
			comp.Function, deadURLFields)
	}

	// Dark section contract validation (warning only, non-blocking)
	// Uses is_dark_section from DB when available, falls back to CSS auto-detection.
	// See validate_dark_section.go and 014_section_context_contract.md.
	if missing := ValidateDarkSectionContract(rendered, comp.IsDarkSection, params.Logger); len(missing) > 0 {
		params.Logger.Warn("RenderComponentAction: Dark section missing --section-* variables",
			zap.String("component_function", comp.Function),
			zap.Bool("is_dark_section", comp.IsDarkSection),
			zap.Strings("missing_vars", missing),
		)
	}

	params.Logger.Info("RenderComponentAction: Component rendered",
		zap.String("component", comp.Name),
		zap.String("function", comp.Function),
		zap.Int("output_length", len(rendered)),
	)

	result := map[string]interface{}{
		"rendered_html":      rendered,
		"component_id":       comp.ID,
		"component_name":     comp.Name,
		"component_function": comp.Function,
	}

	// Provenance: WHICH template text produced these bytes (RFC_046, ruled
	// 2026-08-22). component_id above says which component was asked; this says
	// which version of it actually ran, and the two stop agreeing the moment a
	// template is edited. The save resolves it to a component_versions row.
	//
	// Emitted only when the seam set it — empty means unknown, and a caller
	// downstream must write NULL rather than infer. That distinction is the whole
	// point of the field: bugs_open/357 is what happens when "I do not know what
	// produced this" is quietly upgraded into a confident answer.
	if renderCtx.RenderedTemplateSHA != "" {
		result["rendered_template_sha"] = renderCtx.RenderedTemplateSHA
	}

	// ── The section's OWN slot identity, when the caller knows it (bugs_open/189)
	// Everything above names the COMPONENT. On a decomposed page the section's
	// name is positional ("prose-0", "tool-2") and belongs to the page, not to
	// the component that renders it — so nothing in this result held it, and the
	// save derived a slot name from component_function instead: a silent rename
	// that also defeats the locked-row guard, which matches on that name.
	// The build loop already has it on `current_section.name`; the config path
	// makes that available without this action having to know the loop's shape.
	// Opt-in and silent when unset — a caller with no structured slot identity
	// (tool recreation) must emit nothing, so the save keeps deriving as before.
	if slotFrom, ok := config["slot_name_from"].(string); ok && slotFrom != "" {
		if slot := datahelpers.ExtractNestedFieldString(params.CollectedData, slotFrom); slot != "" {
			result["stored_slot_name"] = slot
		} else {
			params.Logger.Debug("RenderComponentAction: slot_name_from resolved to nothing — the save will derive the slot name",
				zap.String("slot_name_from", slotFrom))
		}
	}

	if sectionContentData != nil {
		result["content_data"] = sectionContentData
	}
	if len(strippedMarkdownFields) > 0 {
		result["stripped_markdown_fields"] = strippedMarkdownFields
	}
	return result, nil
}

// ============================================================================
// ACTION: compile_page_sections
// ============================================================================

// CompilePageSectionsAction combines multiple rendered sections into a page
// Config:
//   - sections_from OR sections_field: path to array of section results
//   - page_from: path to page data for name/title
//   - page_name: explicit name for the page (fallback)
//   - inject_header: boolean
//   - inject_footer: boolean
func CompilePageSectionsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CompilePageSectionsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Accept both "sections_from" and "sections_field" for compatibility
	sectionsField := "rendered_sections"
	if sf, ok := config["sections_from"].(string); ok && sf != "" {
		sectionsField = sf
	} else if sf, ok := config["sections_field"].(string); ok && sf != "" {
		sectionsField = sf
	}

	params.Logger.Info("CompilePageSectionsAction: Looking for sections",
		zap.String("sections_field", sectionsField))

	sectionsData := datahelpers.ExtractNestedField(params.CollectedData, sectionsField)
	if sectionsData == nil {
		// Try with .results suffix (loop action output format)
		sectionsData = datahelpers.ExtractNestedField(params.CollectedData, sectionsField+".results")
		if sectionsData != nil {
			params.Logger.Info("CompilePageSectionsAction: Found sections at .results path")
		}
	}

	if sectionsData == nil {
		params.Logger.Error("CompilePageSectionsAction: Sections not found",
			zap.String("tried_path", sectionsField),
			zap.String("also_tried", sectionsField+".results"),
			zap.Strings("available_keys", datahelpers.GetMapKeys(params.CollectedData)))
		return nil, fmt.Errorf("sections not found at %s", sectionsField)
	}

	var sections []string
	var sectionsMetadata []map[string]interface{}

	switch v := sectionsData.(type) {
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				sections = append(sections, s)
				// String-only item: no metadata available
				sectionsMetadata = append(sectionsMetadata, map[string]interface{}{
					"rendered_html": s,
				})
			} else if m, ok := item.(map[string]interface{}); ok {
				html, meta := extractSectionFromMap(m, params.Logger)
				if html != "" {
					sections = append(sections, html)
					sectionsMetadata = append(sectionsMetadata, meta)
				}
			}
		}
	case map[string]interface{}:
		// Check if this is a loop output with "results" array
		if results, ok := v["results"].([]interface{}); ok {
			for _, item := range results {
				if m, ok := item.(map[string]interface{}); ok {
					html, meta := extractSectionFromMap(m, params.Logger)
					if html != "" {
						sections = append(sections, html)
						sectionsMetadata = append(sectionsMetadata, meta)
					}
				}
			}
		} else {
			// Ordered by keys (section_0, section_1, etc.)
			for i := 0; i < len(v); i++ {
				key := fmt.Sprintf("section_%d", i)
				if section, ok := v[key]; ok {
					if s, ok := section.(string); ok {
						sections = append(sections, s)
						sectionsMetadata = append(sectionsMetadata, map[string]interface{}{
							"rendered_html": s,
						})
					} else if m, ok := section.(map[string]interface{}); ok {
						html, meta := extractSectionFromMap(m, params.Logger)
						if html != "" {
							sections = append(sections, html)
							sectionsMetadata = append(sectionsMetadata, meta)
						}
					}
				}
			}
		}
	}

	params.Logger.Info("CompilePageSectionsAction: Extracted sections",
		zap.Int("count", len(sections)))

	if len(sections) == 0 {
		params.Logger.Warn("CompilePageSectionsAction: No sections to compile, returning placeholder")

		// Get page info for context
		pageName := "page"
		if pageFrom, ok := config["page_from"].(string); ok && pageFrom != "" {
			if pageData := datahelpers.ExtractNestedField(params.CollectedData, pageFrom); pageData != nil {
				if pm, ok := pageData.(map[string]interface{}); ok {
					if name, ok := pm["name"].(string); ok && name != "" {
						pageName = name
					}
				}
			}
		}

		return map[string]interface{}{
			"page_body":     "",
			"page_name":     pageName,
			"section_count": 0,
			"skipped":       true,
			"reason":        "no sections defined for page",
		}, nil
	}

	// Build page body
	pageBody := strings.Join(sections, "\n\n")

	// Get page name - try page_from first, then page_name config
	pageName := "index"
	if pageFrom, ok := config["page_from"].(string); ok && pageFrom != "" {
		if pageData := datahelpers.ExtractNestedField(params.CollectedData, pageFrom); pageData != nil {
			if pm, ok := pageData.(map[string]interface{}); ok {
				if name, ok := pm["name"].(string); ok && name != "" {
					pageName = name
				}
			}
		}
	}
	if pn, ok := config["page_name"].(string); ok && pn != "" {
		pageName = pn
	}

	// Build full HTML page
	pageHTML := buildPageHTML(pageName, pageBody)

	// Optionally inject head/header/footer from component library.
	// Head is injected here (same time as header/footer) rather than deferred
	// to a later assemble_page step — deferring caused <head> to end up inside
	// <body> when cleanHTMLStructure's dedup logic picked the wrong block.
	injectHead, _ := config["inject_head"].(bool)
	injectHeader, _ := config["inject_header"].(bool)
	injectFooter, _ := config["inject_footer"].(bool)

	if params.DB != nil && (injectHead || injectHeader || injectFooter) {
		// Get site_id for component lookup
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
		siteUUID := uuid.Nil
		if siteIDStr != "" {
			siteUUID, _ = uuid.Parse(siteIDStr)
		}

		renderCtx := &RenderContext{}
		if rc := datahelpers.ExtractNestedField(params.CollectedData, "render_context"); rc != nil {
			if m, ok := rc.(map[string]interface{}); ok {
				mergeIntoRenderContext(renderCtx, m)
			}
		}

		// Ensure page-specific title/description are set for head component.
		// The head template uses {{.title}} and {{.description}} — these must
		// reflect the current page, not just the site-level defaults.
		if pageFrom, ok := config["page_from"].(string); ok && pageFrom != "" {
			if pageData := datahelpers.ExtractNestedField(params.CollectedData, pageFrom); pageData != nil {
				if pm, ok := pageData.(map[string]interface{}); ok {
					if t, ok := pm["title"].(string); ok && t != "" {
						renderCtx.Title = t
					} else if n, ok := pm["name"].(string); ok && n != "" && renderCtx.Title == "" {
						renderCtx.Title = strings.Title(strings.ReplaceAll(n, "-", " "))
					}
					if d, ok := pm["meta_description"].(string); ok && d != "" {
						renderCtx.Description = d
					}
				}
			}
		}

		if injectHead {
			pageHTML = InjectHead(ctx, params.DB, pageHTML, siteUUID, renderCtx, params.Logger)
		}
		if injectHeader {
			pageHTML = InjectHeader(ctx, params.DB, pageHTML, siteUUID, renderCtx, params.Logger)
		}
		if injectFooter {
			pageHTML = InjectFooter(ctx, params.DB, pageHTML, siteUUID, renderCtx, params.Logger)
		}
	}

	params.Logger.Info("CompilePageSectionsAction: Page compiled",
		zap.String("page_name", pageName),
		zap.Int("section_count", len(sections)),
		zap.Int("html_length", len(pageHTML)),
	)

	return map[string]interface{}{
		"page_html":         pageHTML,
		"page_name":         pageName,
		"section_count":     len(sections),
		"sections_metadata": sectionsMetadata,
	}, nil
}

// extractSectionFromMap reads HTML and component metadata from a section result map.
// It handles two shapes:
//
//  1. Direct RenderComponentAction output: metadata lives at top level alongside
//     rendered_html. This is what CompilePageSectionsAction used to assume.
//
//  2. Loop-wrapped output: LoopAction's completion step (Strategy 1,
//     substep_output_fields) promotes HTML to top-level rendered_html/page_html,
//     but component_id/component_name/component_function stay nested inside the
//     substep output key (section_output, render_section, render_from_template).
//     This is what content-writer's process_sections_loop produces in practice.
//
// Lookup order: top-level first, then nested substep keys (earliest-wins). This
// preserves the historical behaviour for callers with the flat shape while
// recovering metadata from the nested shape — without which save_page_sections
// ends up with ComponentName = "section" (the default from extractSectionsFromMetadata)
// and enrichSectionsWithComponentIDs skips every section, leaving page_components.component_id NULL.
//
// Returned meta always contains rendered_html. When available, also:
// component_id, component_name, component_function, content_data,
// stored_slot_name.
//
// stored_slot_name is forwarded in BOTH shapes for the same reason the others
// are (bugs_open/189): it is the section's own positional identity, the save
// prefers it verbatim over a name derived from the component, and a metadata
// entry that drops it here leaves the save with nothing but the component's
// function — which is the rename that defeats the locked-row guard.
// Returns ("", nil) if no HTML could be found in any known position.
func extractSectionFromMap(m map[string]interface{}, logger *zap.Logger) (string, map[string]interface{}) {
	// Extract HTML — check top-level first, then common substep keys.
	html := extractHTMLFromSectionMap(m)
	if html == "" {
		return "", nil
	}

	meta := map[string]interface{}{
		"rendered_html": html,
	}

	// Collect component metadata from top level first.
	//
	// ONE declared list, not a hand-written block per key (RFC_046;
	// bugs_open/357). Every key a producer sets is either in
	// sectionMetadataCarryKeys or in sectionMetadataDeniedKeys with its reason,
	// and section_metadata_parity_test.go fails on a key in neither. Before that
	// contract, this function rebuilt the map from a literal six-key list and
	// silently dropped rendered_template_sha, which severed the identity stamp
	// for a day of production without a single error or failing test — the same
	// way it dropped stored_slot_name in bugs_open/189.
	for _, key := range sectionMetadataCarryKeys {
		carrySectionMetaKey(meta, m, key)
	}

	// Remember whether top-level already had the name, so we only log recovery
	// when the nested fallback actually contributed it.
	_, hadTopName := m["component_name"].(string)

	if !sectionMetaComplete(meta) {

		for _, subKey := range []string{"section_output", "render_section", "render_from_template"} {
			nested, ok := m[subKey].(map[string]interface{})
			if !ok {
				continue
			}
			// Same declared list as the top-level pass. carrySectionMetaKey never
			// overwrites, so a value already recovered from an earlier substep — or
			// present at the top level — still wins.
			for _, key := range sectionMetadataCarryKeys {
				carrySectionMetaKey(meta, nested, key)
			}
			if sectionMetaComplete(meta) {
				break
			}
		}

		// Signal that the nested-lookup fallback fired — the primary diagnostic
		// signal that this fix is taking effect in production logs.
		if !hadTopName {
			if n, ok := meta["component_name"].(string); ok && n != "" {
				logger.Info("CompilePageSectionsAction: recovered component_name from nested substep output",
					zap.String("component_name", n))
			}
		}
	}

	return html, meta
}

// extractHTMLFromSectionMap pulls the rendered HTML string from a section-result
// map, checking top-level keys first, then common substep output keys where
// LoopAction may have nested the RenderComponentAction result.
func extractHTMLFromSectionMap(m map[string]interface{}) string {
	if h, ok := m["rendered_html"].(string); ok && h != "" {
		return h
	}
	if h, ok := m["page_html"].(string); ok && h != "" {
		return h
	}
	if h, ok := m["html"].(string); ok && h != "" {
		return h
	}
	// Also try nested substep keys (loop-wrapped shape where top-level HTML
	// promotion didn't happen for some reason).
	for _, subKey := range []string{"section_output", "render_section", "render_from_template"} {
		if nested, ok := m[subKey].(map[string]interface{}); ok {
			if h, ok := nested["rendered_html"].(string); ok && h != "" {
				return h
			}
			if h, ok := nested["page_html"].(string); ok && h != "" {
				return h
			}
		}
	}
	return ""
}

func buildPageHTML(pageName, body string) string {
	title := strings.Title(strings.ReplaceAll(pageName, "-", " "))
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
</head>
<body>
%s
</body>
</html>`, title, body)
}

// ============================================================================
// ACTION: insert_research_result
// ============================================================================

// InsertResearchResultAction stores research findings in the database
// Config:
//   - table: target table name (default: "research_results")
//   - fields: map of column_name -> data_path for dynamic field mapping
//   - site_id_field: path to site_id (fallback if not in fields)
//   - result_type: type of research (fallback if not in fields)
func InsertResearchResultAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("InsertResearchResultAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		params.Logger.Warn("InsertResearchResultAction: No database")
		return map[string]interface{}{"inserted": false, "reason": "no database"}, nil
	}

	// Get table name (default to research_results)
	tableName := "research_results"
	if tn, ok := config["table"].(string); ok && tn != "" {
		tableName = tn
	}

	// Get field mappings from config
	fieldMappings, hasFields := config["fields"].(map[string]interface{})

	// Generate new ID
	resultID := uuid.New()

	// Build dynamic INSERT based on field mappings
	if hasFields && len(fieldMappings) > 0 {
		// Dynamic field-based insert
		columns := []string{"id"}
		placeholders := []string{"$1"}
		values := []interface{}{resultID}
		paramIdx := 2

		for column, dataPath := range fieldMappings {
			dataPathStr, ok := dataPath.(string)
			if !ok {
				continue
			}

			// Extract the value from collected data
			value := datahelpers.ExtractNestedField(params.CollectedData, dataPathStr)

			// Special handling for site_id (needs UUID conversion)
			if column == "site_id" {
				if siteIDStr, ok := value.(string); ok && siteIDStr != "" {
					if siteUUID, err := uuid.Parse(siteIDStr); err == nil {
						value = siteUUID
					} else {
						value = nil
					}
				} else {
					value = nil
				}
			}

			// Skip nil values for optional fields, but include empty strings
			if value == nil {
				continue
			}

			// For complex types, marshal to JSON
			/*switch v := value.(type) {
			case map[string]interface{}, []interface{}:
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					params.Logger.Warn("InsertResearchResultAction: Failed to marshal field",
						zap.String("column", column),
						zap.Error(err))
					continue
				}
				columns = append(columns, column)
				placeholders = append(placeholders, fmt.Sprintf("$%d::jsonb", paramIdx))
				values = append(values, string(jsonBytes))
			default:
				columns = append(columns, column)
				placeholders = append(placeholders, fmt.Sprintf("$%d", paramIdx))
				values = append(values, value)
			}*/

			// make it all json
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				params.Logger.Warn("InsertResearchResultAction: Failed to marshal field",
					zap.String("column", column),
					zap.String("value_type", fmt.Sprintf("%T", value)),
					zap.Error(err))
				continue
			}
			columns = append(columns, column)
			placeholders = append(placeholders, fmt.Sprintf("$%d::jsonb", paramIdx))
			values = append(values, string(jsonBytes))
			paramIdx++
		}

		// Add created_at
		columns = append(columns, "created_at")
		placeholders = append(placeholders, "NOW()")

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
		)

		params.Logger.Info("InsertResearchResultAction: Executing dynamic insert",
			zap.String("table", tableName),
			zap.Strings("columns", columns),
			zap.Int("value_count", len(values)))

		if err := execDB(ctx, params.DB, query, values...); err != nil {
			params.Logger.Warn("InsertResearchResultAction: Insert failed",
				zap.String("query", query),
				zap.Error(err))
			return map[string]interface{}{
				"inserted":    false,
				"result_type": "general",
				"error":       err.Error(),
			}, nil
		}

		return map[string]interface{}{
			"inserted":  true,
			"id":        resultID.String(),
			"result_id": resultID.String(),
			"table":     tableName,
			"columns":   columns,
		}, nil
	}

	// Fallback: Legacy mode with hardcoded columns
	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	siteID := uuid.Nil
	if siteIDStr != "" {
		siteID, _ = uuid.Parse(siteIDStr)
	}

	// Get result type
	resultType := "general"
	if rt, ok := config["result_type"].(string); ok && rt != "" {
		resultType = rt
	}

	// Get research data - try multiple field paths
	dataField := "research_result"
	if df, ok := config["data_field"].(string); ok && df != "" {
		dataField = df
	}
	researchData := datahelpers.ExtractNestedField(params.CollectedData, dataField)

	// If no data found, try to build from synthesis
	if researchData == nil {
		researchData = map[string]interface{}{
			"summary":  datahelpers.ExtractNestedField(params.CollectedData, "synthesis.summary"),
			"findings": datahelpers.ExtractNestedField(params.CollectedData, "synthesis"),
			"query":    datahelpers.ExtractNestedField(params.CollectedData, "search_query.result"),
			"topic":    datahelpers.ExtractNestedField(params.CollectedData, "extracted.topic"),
		}
	}

	dataJSON, err := json.Marshal(researchData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal research data: %w", err)
	}

	// Try insert with 'findings' column first (new schema), fall back to 'data' (old schema)
	var siteIDArg interface{} = nil
	if siteID != uuid.Nil {
		siteIDArg = siteID
	}

	// Try new schema first
	query := `
		INSERT INTO research_results (id, site_id, result_type, findings, created_at)
		VALUES ($1, $2, $3, $4::jsonb, NOW())
	`

	if err := execDB(ctx, params.DB, query, resultID, siteIDArg, resultType, string(dataJSON)); err != nil {
		// Try with 'data' column (old schema)
		query = `
			INSERT INTO research_results (id, site_id, result_type, data, created_at)
			VALUES ($1, $2, $3, $4::jsonb, NOW())
		`
		if err2 := execDB(ctx, params.DB, query, resultID, siteIDArg, resultType, string(dataJSON)); err2 != nil {
			params.Logger.Warn("InsertResearchResultAction: Insert failed with both schemas",
				zap.Error(err),
				zap.Error(err2))
			return map[string]interface{}{
				"inserted":    false,
				"result_type": resultType,
				"error":       err2.Error(),
			}, nil
		}
	}

	return map[string]interface{}{
		"inserted":    true,
		"result_id":   resultID.String(),
		"result_type": resultType,
		"data_size":   len(dataJSON),
	}, nil
}

// ============================================================================
// ACTION: store_asset
// ============================================================================

// StoreAssetAction stores an asset (image, font, file) in the assets table
// This is a LOCAL action - it does NOT require a topic
// Config:
//   - asset_type: type of asset (image, font, css, js, etc.)
//   - site_id_field: path to site_id in collected_data
//   - data_field: path to asset data (URL, base64, or content)
//   - name_field: path to asset name
//   - metadata_field: optional path to additional metadata
//   - update_site_brand_assets: optional bool, default true. When false, the
//     site-wide sites.content_data.<purpose>_url is left untouched. Only ever
//     consulted for an asset that IS the site-wide one for its purpose
//     (asset_key == purpose); a page-scoped asset never writes it regardless.
func StoreAssetAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("StoreAssetAction: Starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get asset type
	assetType := "image"
	if at, ok := config["asset_type"].(string); ok && at != "" {
		assetType = at
	}

	// Get purpose from config — supports literal OR field lookup.
	// Phase 2H: purpose_field added so the new needs_imagery branch can pass
	// spec.purpose through dynamically (logo, hero, illustration, icon,
	// infographic) without hardcoding in workflow step config. Mirrors the
	// asset_key / asset_key_field pattern below.
	//
	// Resolution priority:
	//   1. config["purpose"]        — literal string (e.g. "hero", "logo")
	//   2. config["purpose_field"]  — JSONPath into collected_data
	//   3. ""                       — empty; downstream asset_key resolution
	//                                 may still backfill via asset_key_field
	purpose := ""
	if p, ok := config["purpose"].(string); ok && p != "" {
		purpose = p
	}
	if purpose == "" {
		if pf, ok := config["purpose_field"].(string); ok && pf != "" {
			purpose = datahelpers.ExtractNestedFieldString(params.CollectedData, pf)
		}
	}

	// Phase 2C: extract asset_key from config, defaulting to purpose.
	// Phase 2E: also support asset_key_field for JSONPath lookup so
	// per-item variants can be passed through the workflow without
	// hardcoded literals.
	//
	// Resolution priority:
	//   1. config["asset_key"]        — literal string (e.g. "logo")
	//   2. config["asset_key_field"]  — JSONPath into collected_data
	//   3. purpose (default — Phase 2C backward-compat)
	assetKey := ""
	if k, ok := config["asset_key"].(string); ok && k != "" {
		assetKey = k
	}
	if assetKey == "" {
		if kf, ok := config["asset_key_field"].(string); ok && kf != "" {
			assetKey = datahelpers.ExtractNestedFieldString(params.CollectedData, kf)
		}
	}
	if assetKey == "" {
		assetKey = purpose
	}

	// Get site_id (optional - assets can be global)
	var siteID *uuid.UUID
	if siteIDField, ok := config["site_id_field"].(string); ok && siteIDField != "" {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
		if siteIDStr != "" {
			if parsed, err := uuid.Parse(siteIDStr); err == nil {
				siteID = &parsed
			}
		}
	}

	// Get asset data
	dataField := "asset_data"
	if df, ok := config["data_field"].(string); ok && df != "" {
		dataField = df
	}
	assetData := datahelpers.ExtractNestedField(params.CollectedData, dataField)

	// Get asset name
	nameField := "asset_name"
	if nf, ok := config["name_field"].(string); ok && nf != "" {
		nameField = nf
	}
	assetName := datahelpers.ExtractNestedFieldString(params.CollectedData, nameField)
	if assetName == "" {
		assetName = fmt.Sprintf("%s_%s", assetType, uuid.New().String()[:8])
	}

	// Extract URL or content from asset data
	var assetURL string
	switch v := assetData.(type) {
	case string:
		if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") || strings.HasPrefix(v, "s3://") {
			assetURL = v
		}
	case map[string]interface{}:
		if url, ok := v["url"].(string); ok {
			assetURL = url
		}
		if url, ok := v["image_url"].(string); ok {
			assetURL = url
		}
	}

	if assetURL == "" {
		params.Logger.Warn("StoreAssetAction: No asset URL found")
		return map[string]interface{}{
			"stored":     false,
			"asset_name": assetName,
			"asset_type": assetType,
			"reason":     "no asset URL found",
		}, nil
	}

	// If no DB, return success without persisting
	if params.DB == nil {
		params.Logger.Warn("StoreAssetAction: No database, returning without persistence")
		return map[string]interface{}{
			"stored":     true,
			"persisted":  false,
			"asset_id":   uuid.New().String(),
			"asset_name": assetName,
			"asset_type": assetType,
			"asset_url":  assetURL,
		}, nil
	}

	// Determine origin_type based on URL
	originType := "uploaded"
	if strings.HasPrefix(assetURL, "s3://") || strings.Contains(assetURL, "backblazeb2.com") {
		originType = "generated"
	}

	// Phase 0.2: extract origin_prompt and origin_model from step config.
	// Bug fix — origin_prompt_field has been silently passed by workflows
	// without this action ever reading it. After this lands, new generations
	// populate the column. New — origin_model accepts a literal (used today,
	// e.g. "sdxl") or a path (origin_model_field) for future provider routing.
	// Literal wins if both are set.
	originPrompt := ""
	if pf, ok := config["origin_prompt_field"].(string); ok && pf != "" {
		originPrompt = datahelpers.ExtractNestedFieldString(params.CollectedData, pf)
	}
	originModel := ""
	if m, ok := config["origin_model"].(string); ok && m != "" {
		originModel = m
	} else if mf, ok := config["origin_model_field"].(string); ok && mf != "" {
		originModel = datahelpers.ExtractNestedFieldString(params.CollectedData, mf)
	}

	// Resolve the durable storage URI (s3://bucket/key) BEFORE the insert, so
	// the row records its own source object from birth (bugs_open/152 + /155:
	// this value used to be written only to the site-wide, purpose-keyed
	// content_data cache — last-write-wins across every same-purpose asset —
	// while the per-asset row kept just an expiring presigned url).
	storageURI := ""
	if assetDataMap, ok := assetData.(map[string]interface{}); ok {
		if uri, ok := assetDataMap["image_uri"].(string); ok {
			storageURI = uri
		}
	}
	// Also check collected_data for {purpose}_result.image_uri pattern
	if storageURI == "" {
		uriField := strings.TrimSuffix(dataField, ".image_url") + ".image_uri"
		if uri := datahelpers.ExtractNestedFieldString(params.CollectedData, uriField); uri != "" {
			storageURI = uri
		}
	}
	// Use assetURL if it's a storage URI
	if storageURI == "" && storage.IsS3URI(assetURL) {
		storageURI = assetURL
	}
	// For the COLUMN only: a presigned assetURL still names the object — keep
	// the durable s3:// form of it even when nothing upstream supplied a URI.
	// storageURI itself keeps its old value so the result fields and the
	// collected_data key carry exactly what they used to.
	storagePathValue := storageURI
	if storagePathValue == "" {
		storagePathValue = storage.PresignedURLToS3URI(assetURL)
	}
	storageProvider := ""
	if storagePathValue != "" {
		storageProvider = "backblaze"
	}

	// Insert into assets table - matches actual schema
	assetID := uuid.New()

	query := `
		INSERT INTO assets (id, site_id, name, asset_type, purpose, asset_key, url, origin_type,
		                    origin_prompt, origin_model, storage_path, storage_provider, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (site_id, asset_key) WHERE asset_key IS NOT NULL AND status = 'active' DO UPDATE SET
			purpose = EXCLUDED.purpose,
			url = EXCLUDED.url,
			name = EXCLUDED.name,
			origin_type = EXCLUDED.origin_type,
			origin_prompt = COALESCE(EXCLUDED.origin_prompt, assets.origin_prompt),
			origin_model = COALESCE(EXCLUDED.origin_model, assets.origin_model),
			storage_path = COALESCE(EXCLUDED.storage_path, assets.storage_path),
			storage_provider = COALESCE(EXCLUDED.storage_provider, assets.storage_provider),
			updated_at = NOW()
		WHERE ` + assetAgentWritableSQL("assets.") + `
		RETURNING id
	`

	var returnedID uuid.UUID
	err := queryRowScanUUID(ctx, params.DB, query, &returnedID,
		assetID, siteID, assetName, assetType, nullString(purpose),
		nullString(assetKey), assetURL, originType,
		nullString(originPrompt), nullString(originModel),
		nullString(storagePathValue), nullString(storageProvider))

	// Phase I1 (D5, logo permanence): the conflict target matched a LOCKED
	// asset — the DO UPDATE ... WHERE suppressed the write and RETURNING
	// produced no row. Approved assets (locked_at set, e.g. an
	// approve-and-locked logo) must never be silently replaced by a fresh
	// generation. Report it as a refusal, not an error, so callers complete.
	if err != nil && strings.Contains(err.Error(), "no rows") {
		params.Logger.Warn("StoreAssetAction: target asset is LOCKED — refusing to overwrite",
			zap.String("asset_key", assetKey),
			zap.String("purpose", purpose))
		return map[string]interface{}{
			"stored":     false,
			"locked":     true,
			"asset_key":  assetKey,
			"asset_name": assetName,
			"reason":     "asset is locked (locked_at set) — approved assets are never overwritten",
		}, nil
	}

	if err != nil {
		// Try simpler insert without upsert if constraint doesn't exist
		params.Logger.Warn("StoreAssetAction: Upsert failed, trying simple insert",
			zap.Error(err))

		simpleQuery := `
			INSERT INTO assets (id, site_id, name, asset_type, purpose, asset_key, url, origin_type,
			                    origin_prompt, origin_model, storage_path, storage_provider, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
			RETURNING id
		`
		err = queryRowScanUUID(ctx, params.DB, simpleQuery, &returnedID,
			assetID, siteID, assetName, assetType, nullString(purpose),
			nullString(assetKey), assetURL, originType,
			nullString(originPrompt), nullString(originModel),
			nullString(storagePathValue), nullString(storageProvider))

		if err != nil {
			params.Logger.Warn("StoreAssetAction: Insert failed",
				zap.Error(err))
			return map[string]interface{}{
				"stored":     true,
				"persisted":  false,
				"asset_id":   assetID.String(),
				"asset_name": assetName,
				"asset_type": assetType,
				"asset_url":  assetURL,
				"error":      err.Error(),
			}, nil
		}
	}

	// If purpose is set and we have a site_id, update sites.content_data with
	// the template-facing relative URL. The {purpose}_uri DB write that used
	// to sit here is GONE (bugs_open/155): it was a site-wide, last-write-wins
	// cache of a per-asset fact, and its only reader — deploy_image_asset's
	// old Priority-1 — deployed the wrong asset's bytes on any site with 2+
	// same-purpose assets. The source now travels on the asset row itself
	// (storage_path, written above). The in-run collected_data copy is gone
	// too (209 Phase 3, 2026-08-22): its stated reader — findStorageURI's
	// {purpose}_uri lookup in the legacy pageflow deploy step — was deleted on
	// 2026-08-09 (91dda3243, "resolved by identity or not at all"), and the
	// copy was measured readerless across Go, live agent_definitions,
	// workflow_templates and active component templates before removal.
	if purpose != "" && siteID != nil {
		// bugs_open/114: sites.content_data.<purpose>_url is SITE-WIDE brand
		// state, and a page-scoped asset is not a site-wide fact. Writing it on
		// every store made each page-scoped generation overwrite the site default
		// with a path derived from the purpose — measured 2026-08-22 as 18 sites
		// carrying an identical hero_url, 6 carrying content_hero.jpg (404 on all
		// six, and a filename the deployer cannot produce), and one site's
		// hand-repair silently undone by the next generation.
		//
		// Two conditions, both required:
		//   1. the asset IS the site-wide one for its purpose (asset_key == purpose);
		//   2. the caller has not declared update_site_brand_assets: false.
		// The second existed in workflow config from the start — image-build-handler's
		// imagery store step has passed false since it was written — and no Go code
		// read it, so the workflow's own instruction was silently discarded.
		deployedURL, writeSiteWide := storeAssetContentDataUpdate(assetKey, purpose, config)

		if writeSiteWide {
			// The one remaining way a page-scoped key can reach site-wide state,
			// made visible rather than left silent (council round 2 objection on
			// corr 3c0560f3, which was right to raise it). The declaration is
			// per-STEP config, so it cannot itself distinguish "this really is the
			// site's hero" from "this is one page's hero filed through the brand
			// branch". Reachability IS per-invocation — the only step combining a
			// dynamic asset_key with a true declaration
			// (store_imagery_brand_asset) is behind `input_data.spec.brand_update
			// == true`, a per-item decision by the producing check — and today
			// every such item is hero_home or logo, both legitimately
			// site-representative. But nothing in code constrains a future
			// producer from filing hero_about the same way, and that would
			// re-point the site default exactly as bugs_open/114 describes.
			// So it is greppable: if this WARN ever names a page-scoped key,
			// the producer that filed it is the bug.
			if assetKey != purpose {
				params.Logger.Warn("StoreAssetAction: site-wide brand state written from a PAGE-SCOPED asset key — the caller declared update_site_brand_assets, so this is permitted, but it re-points every page that falls back to this purpose",
					zap.String("purpose", purpose),
					zap.String("asset_key", assetKey),
					zap.String("relative_url", deployedURL),
					zap.String("bug", "bugs_open/114"))
			}
			updateContentDataField(ctx, params.DB, *siteID, purpose+"_url", deployedURL, params.Logger)
			params.Logger.Info("StoreAssetAction: Updated content_data for purpose",
				zap.String("purpose", purpose),
				zap.String("asset_key", assetKey),
				zap.String("storage_uri", storageURI),
				zap.String("relative_url", deployedURL))
		} else {
			params.Logger.Info("StoreAssetAction: Left site-wide content_data untouched",
				zap.String("purpose", purpose),
				zap.String("asset_key", assetKey),
				zap.String("relative_url", deployedURL))
		}

		// The in-run copy is per-asset, not site-wide, so it is written either
		// way — and it now names this asset's own deployed path.
		params.CollectedData[purpose+"_url"] = deployedURL
	}

	params.Logger.Info("StoreAssetAction: Asset stored",
		zap.String("asset_id", returnedID.String()),
		zap.String("asset_name", assetName),
		zap.String("asset_type", assetType),
	)

	result := map[string]interface{}{
		"stored":     true,
		"persisted":  true,
		"asset_id":   returnedID.String(),
		"asset_name": assetName,
		"asset_type": assetType,
		"asset_url":  assetURL,
	}

	// Add purpose-specific fields if set. Same derivation as the collected_data
	// copy above: this asset's own deployed path, not the purpose's generic one.
	if purpose != "" {
		result["purpose"] = purpose
		resultURL, _ := storeAssetContentDataUpdate(assetKey, purpose, config)
		result[purpose+"_url"] = resultURL
	}

	// Add storage URI to result for downstream deploy step
	if storageURI != "" {
		result["image_uri"] = storageURI
		result["s3_uri"] = storageURI
	}

	return result, nil
}

// writesSiteBrandState reports whether a store_asset call may write the
// site-wide sites.content_data.<purpose>_url.
//
// Two conditions, both required (bugs_open/114):
//
//  1. the asset IS the site-wide one for its purpose. asset_key defaults to
//     purpose, so a canonical brand asset satisfies this without asking for it,
//     while a page-scoped variant (hero_about, content_hero_tool_repayment)
//     never can. Site-wide brand state is not a per-page fact, and writing it
//     per page is last-write-wins across pages that have nothing to do with
//     each other.
//  2. the caller has not switched it off. update_site_brand_assets has been in
//     workflow config since image-build-handler was written and no Go code read
//     it, so a step that explicitly asked not to touch brand state touched it
//     anyway. Absent means true, which is what every existing caller relies on.
//
// It returns the URL as well as the decision, and the two travel together on
// purpose. The first cut of this fix returned only the boolean and left the
// derivation inline at the call site; a mutation reverting that line to the old
// purpose-derived BuildAssetPaths then PASSED the whole suite, because the
// accompanying test exercised storage.DeployedWebPath directly and never the
// action's use of it. A helper the action must call for both answers is what
// makes the derivation mutation-provable — see TestStoreAssetContentDataUpdate.
func storeAssetContentDataUpdate(assetKey, purpose string, config map[string]interface{}) (url string, writeSiteWide bool) {
	// The deployer's own derivation (deploy_image_asset → storage.DeployedAssetPath),
	// so the value recorded here and the file committed to the repo cannot disagree.
	url = storage.DeployedWebPath(assetKey, purpose)

	// An explicit declaration wins in both directions. The key was already on
	// every live step when this fix was written — six say true, two say false —
	// and the two saying false are exactly the page-scoped ones. Honouring it is
	// most of the fix.
	//
	// It has to win in the TRUE direction too, and that is not a formality: the
	// brand-update branch files items with purpose=hero and asset_key=hero_home
	// (10 such items on the live fleet, measured 2026-08-22). Those are a
	// deliberate "this asset is the site's hero", so they must still write — and
	// now they write /assets/images/hero-home.jpg, the file that actually exists,
	// instead of the hero.jpg that did not. That is the repair one site was given
	// by hand on 2026-07-29 and then silently lost, expressed as a rule.
	if declared, ok := config["update_site_brand_assets"].(bool); ok {
		return url, declared
	}

	// Undeclared: write only for the site-wide asset of its purpose. asset_key
	// defaults to purpose, so a canonical brand store satisfies this without
	// asking, while a page-scoped variant (hero_about, content_hero_tool_repayment)
	// never can. This is the arm that stops a future caller re-opening the hole
	// by simply not mentioning the key.
	return url, assetKey == purpose
}

// updateContentDataField updates a single field in sites.content_data
func updateContentDataField(ctx context.Context, db interface{}, siteID uuid.UUID, field, value string, logger *zap.Logger) {
	query := `
        UPDATE sites 
        SET content_data = jsonb_set(
            COALESCE(content_data, '{}'::jsonb),
            $2::text[],
            to_jsonb($3::text),
            true
        ),
        updated_at = NOW()
        WHERE id = $1
    `
	jsonPath := fmt.Sprintf("{%s}", field)

	if err := execDB(ctx, db, query, siteID, jsonPath, value); err != nil {
		logger.Warn("Failed to update content_data field",
			zap.String("field", field),
			zap.Error(err))
	} else {
		logger.Debug("Updated content_data field",
			zap.String("field", field),
			zap.String("value", value))
	}
}

func ValidateSitePlanAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ValidateSitePlanAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	planField := "llm_plan"
	if pf, ok := config["plan_field"].(string); ok && pf != "" {
		planField = pf
	}

	// Extract the plan data
	planData := datahelpers.ExtractNestedField(params.CollectedData, planField)
	if planData == nil {
		return nil, fmt.Errorf("plan not found at '%s'", planField)
	}

	planData = datahelpers.UnwrapDeep(planData, params.Logger)

	params.Logger.Info("ValidateSitePlanAction: After UnwrapDeep",
		zap.String("planData_type", fmt.Sprintf("%T", planData)))

	var plan map[string]interface{}
	switch v := planData.(type) {
	case map[string]interface{}:
		plan = v
	case string:
		if v == "" {
			return nil, fmt.Errorf("plan is empty string - LLM may have returned no content. Check template rendering logs for <no value> placeholders")
		}
		cleaned := strings.TrimSpace(v)
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned == "" {
			return nil, fmt.Errorf("plan is empty after cleaning markdown")
		}
		if err := json.Unmarshal([]byte(cleaned), &plan); err != nil {
			// Include content preview in error for debugging
			preview := cleaned
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			return nil, fmt.Errorf("failed to parse plan JSON: %w (preview: %s)", err, preview)
		}
	default:
		return nil, fmt.Errorf("plan must be object or JSON string, got %T", planData)
	}

	pagesRaw, ok := plan["pages"]
	if !ok {
		// Log available keys for debugging
		keys := make([]string, 0, len(plan))
		for k := range plan {
			keys = append(keys, k)
		}
		params.Logger.Error("ValidateSitePlanAction: plan missing 'pages'",
			zap.Strings("available_keys", keys))
		return nil, fmt.Errorf("plan must have 'pages' array (available keys: %v)", keys)
	}

	pages, ok := pagesRaw.([]interface{})
	if !ok || len(pages) == 0 {
		return nil, fmt.Errorf("pages must be non-empty array")
	}

	// The planner's OWN output, snapshotted before any pass below can touch it
	// (bugs_open/428, recommended_type_reconciliation.go). Taken here rather
	// than derived later because every pass from this point mutates the page
	// maps in place, so there is no way to recover what the planner proposed
	// once they have run — which is exactly why a type the planner planned and
	// this action then deleted has been indistinguishable from one the planner
	// never proposed. Measured on gamedesign.uk 2026-09-03: 9 pages in, 4 out,
	// five blog-posts gone, nothing durable written.
	proposedPageViews := planPageViewsOf(pages)

	// ── Deterministic convergence with realised pages ───────────────────────
	// existing_pages is loaded by the load_existing_pages workflow step and
	// carries site_has_no_current_plan and build_status per page. reconcilePlanWithRealised
	// force-preserves every page on the site's first plan OR built, so a re-plan
	// can no longer silently redesign or drop a built page (bugs_open/001). It is
	// a no-op only for a genuinely from-scratch build. See
	// FOCUS_adoption_faithfulness_via_locks.md.
	existingField := "existing_pages"
	if ef, ok := config["existing_pages_field"].(string); ok && ef != "" {
		existingField = ef
	}
	var existingPages []interface{}
	if ev := datahelpers.ExtractNestedField(params.CollectedData, existingField); ev != nil {
		switch vv := ev.(type) {
		case []interface{}:
			existingPages = vv
		case []map[string]interface{}:
			// query_database (output_format=array) returns []map[string]interface{},
			// which does NOT satisfy a []interface{} assertion in Go. Convert it so
			// the convergence actually sees the realised pages. Without this the
			// assertion silently fails, existingPages stays empty, and
			// reconcilePlanWithRealised no-ops for every site (adopted pages never
			// preserved, planner siblings never dropped).
			existingPages = make([]interface{}, len(vv))
			for i := range vv {
				existingPages[i] = vv[i]
			}
		}
	}
	// Explicit redesign intent (bugs_open/037 fix step 4 / features_open/012).
	// Pages named in the trigger spec's optional `recompose_pages` list are
	// RELEASED from the preserve guard for THIS re-plan only, so the LLM's proposed
	// composition governs them (a page may be recomposed or, if the LLM omits it,
	// dropped). This is the sanctioned way to deliberately redesign a preserved
	// (deployed / needs_rebuild) page — without an entry here the guard preserves
	// it. Filtering the realised set HERE, before both reconcilePlanWithRealised
	// and the truncation must-keep read `existingPages`, makes a recompose page
	// uniformly from-scratch. Ordinary re-plans carry no such field and are
	// unaffected.
	var recomposeRealised map[string][]interface{}
	if recompose := recomposePagesFromSpec(params.CollectedData, params.Logger); len(recompose) > 0 {
		// Captured before the filter: afterwards the realised composition of a
		// released page exists nowhere else, and it is the baseline the
		// verbatim-no-op check below compares against.
		recomposeRealised = realisedSectionsByName(existingPages, recompose)
		existingPages = filterOutRecomposePages(existingPages, recompose, params.Logger)
	}

	// Surface the convergence input size so an empty set is never silent again.
	params.Logger.Info("ValidateSitePlanAction: existing pages loaded for convergence",
		zap.Int("existing_pages", len(existingPages)),
		zap.String("existing_pages_field", existingField))
	var counts reconcileCounts
	// The twin-identity layers are opt-in per SITE, unsafe default OFF
	// (bugs_open/215), read through the same shared helper the two write surfaces
	// use so a site cannot be half opted-in. Off, the layers still measure what
	// they would have done — read TwinIdentityObserved / StemTwinObserved and the
	// durable PLAN_PAGE_*_OBSERVED rows before turning either on.
	reconcileOpts := reconcileOptions{}
	// Held beyond the block below because the same-name identity record needs to
	// say which side of honour_realised_identity this run is on: the same set of
	// pages is a no-op record with the flag on and a list of twins about to be
	// written with it off. A failed site-id parse leaves it false, which is the
	// same conservative reading every other consumer of this policy takes.
	var policy siteIdentityPolicy
	if params.DB != nil {
		if sid, err := uuid.Parse(datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")); err == nil {
			policy = siteIdentityPolicyFor(ctx, params.DB, sid, params.Logger)
			reconcileOpts.TwinIdentitySnap = policy.TwinIdentitySnap
			reconcileOpts.StemTwinSnap = policy.StemTwinSnap
		}
	}
	pages, counts = reconcilePlanWithRealised(pages, existingPages, reconcileOpts, params.Logger)
	plan["pages"] = pages
	params.Logger.Info("ValidateSitePlanAction: reconciled with realised pages",
		zap.Int("unioned_in", counts.Unioned),
		zap.Int("dropped_collision", counts.DroppedCollision),
		zap.Strings("dropped_pages", droppedPlanPageNames(counts.DroppedPages)),
		zap.Int("snapped_rename", counts.SnappedRename),
		zap.Int("snapped_sections", counts.SnappedSections),
		zap.Int("section_facts_carried", counts.SectionFactsCarried),
		zap.Int("fact_carry_miss_pages", len(counts.FactCarryMisses)),
		zap.Int("fact_assignment_absent_pages", len(counts.FactAssignmentAbsent)),
		zap.Int("snapped_identity_path_key", counts.SnappedIdentityPathKey),
		zap.Int("snapped_identity_canon_name", counts.SnappedIdentityCanonName),
		zap.Int("snapped_stem_twin", counts.SnappedStemTwin),
		zap.Int("twin_identity_observed", counts.TwinIdentityObserved),
		zap.Int("stem_twin_observed", counts.StemTwinObserved),
		zap.Int("stem_twin_ambiguous", counts.StemTwinAmbiguous),
		zap.Int("same_name_stamped", len(counts.SameNameStamps)),
		zap.Int("same_name_type_conflicts", len(counts.SameNameTypeConflicts)),
		zap.Int("pages_after", len(pages)))
	recordFactCarryMisses(ctx, params, counts.FactCarryMisses)
	recordFactAssignmentAbsent(ctx, params, counts.FactAssignmentAbsent)
	recordIdentitySnaps(ctx, params, counts.IdentitySnaps)
	recordSameNameIdentityOutcomes(ctx, params, counts, policy.HonourRealisedIdentity)
	recordRecomposeOutcomes(ctx, params, recomposeOutcomes(pages, recomposeRealised))

	// ── Truncate, preserving first-plan AND built pages ─────────────────────
	maxPages := 20
	if mp, ok := config["max_pages"].(float64); ok {
		maxPages = int(mp)
	}
	if len(pages) > maxPages {
		// Must-keep mirrors reconcilePlanWithRealised's preservation set: a
		// built page must survive truncation for exactly the reason it must
		// survive the plan — the LLM re-proposing 80 pages must not be able to
		// evict a page that is live on the site (bugs_open/001, fix step 3).
		var mustKeep []interface{}
		for _, rp := range existingPages {
			if rm, ok := rp.(map[string]interface{}); ok {
				if noCurrentPlanFlag(rm) || realisedPageCompositionIsPreserved(rm) {
					mustKeep = append(mustKeep, rp)
				}
			}
		}
		pages = truncatePreservingRealised(pages, mustKeep, maxPages, params.Logger)
		plan["pages"] = pages
	}

	pageNames := make(map[string]bool)
	for _, p := range pages {
		if pm, ok := p.(map[string]interface{}); ok {
			if name, ok := pm["name"].(string); ok {
				pageNames[name] = true
			}
		}
	}

	if ensurePages, ok := config["ensure_pages"].([]interface{}); ok {
		for _, req := range ensurePages {
			if reqName, ok := req.(string); ok && !pageNames[reqName] {
				pages = append(pages, map[string]interface{}{
					"name": reqName, "title": strings.Title(reqName),
					"nav_label": strings.Title(reqName), "nav_order": len(pages) + 1,
					"in_header": true, "in_footer": true, "sections": []interface{}{},
				})
			}
		}
		plan["pages"] = pages
	}

	if _, ok := plan["style_collection"].(string); !ok {
		if ds, ok := config["default_style"].(string); ok {
			plan["style_collection"] = ds
		}
	}

	if plan["needs_logo"] == nil {
		plan["needs_logo"] = false
	}
	if plan["needs_images"] == nil {
		plan["needs_images"] = false
	}
	if plan["image_prompts"] == nil {
		plan["image_prompts"] = map[string]interface{}{}
	}

	// ── Strip site-chrome components from page sections ──────────────────
	// The LLM sometimes includes header/footer components in page sections
	// arrays (e.g. "header-bold-gradient", "footer-standard"). These are
	// site-level components injected during assembly — not page content.
	// If left in, plan_sections creates bogus HITL items for them.
	if params.DB != nil {
		siteChrome := loadSiteChromeNames(ctx, params.DB, params.Logger)
		if len(siteChrome) > 0 {
			for _, p := range pages {
				pm, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				sectionsRaw, ok := pm["sections"].([]interface{})
				if !ok {
					continue
				}
				var filtered []interface{}
				for _, s := range sectionsRaw {
					name, ok := sectionEntryName(s)
					if !ok {
						filtered = append(filtered, s)
						continue
					}
					if siteChrome[name] {
						params.Logger.Info("ValidateSitePlanAction: stripped site-chrome component from page sections",
							zap.Any("page", pm["name"]),
							zap.String("component", name))
					} else {
						filtered = append(filtered, s)
					}
				}
				pm["sections"] = filtered
			}
		}
	}

	// ── Resolve section names to canonical component functions ───────────
	// Implements config flag `validate_components`. Each section name must
	// map to a real content_components.function. Display names ("FAQ
	// Section"), wrong case, and underscore variants are resolved;
	// unresolvable names are dropped + logged. This does NOT deduplicate or
	// make content-intent decisions — it only guarantees every surviving
	// section name is a valid component function.
	validateComponents := false
	if vc, ok := config["validate_components"].(bool); ok {
		validateComponents = vc
	}
	if validateComponents && params.DB != nil {
		resolver := loadComponentNameResolver(ctx, params.DB, params.Logger)
		// Accept what the planner was OFFERED. `menu_field` names the step whose
		// output the planner's prompt rendered as its component menu; those rows
		// are added to the valid set, so a widening of the menu (migration 407's
		// tool-level components, per-site opt-in) can never be silently eaten by
		// this surface again. Absent key = today's behaviour, unchanged.
		// bugs_open/282; the arm itself is component_name_resolver_menu.go.
		menuFieldConfigured, _ := config["menu_field"].(string)
		if menuFieldConfigured != "" {
			if rows, ok := menuRowsFrom(params.CollectedData, menuFieldConfigured); ok {
				added := resolver.addMenu(rows)
				params.Logger.Info("ValidateSitePlanAction: accepting the planner's own component menu",
					zap.String("menu_field", menuFieldConfigured),
					zap.Int("menu_rows", len(rows)),
					zap.Int("added_beyond_base", added))
			} else {
				params.Logger.Warn("ValidateSitePlanAction: menu_field configured but no component menu at that path — falling back to the section/element base",
					zap.String("menu_field", menuFieldConfigured))
			}
		}
		// Every drop is recorded durably, whether or not a menu was configured
		// — a Warn is not a record on a service whose logs rotate sub-second,
		// and a silently lost section is the shape this lane exists to remove
		// (bugs_open/282, council round 1). See recordDroppedSectionNames.
		var dropped []droppedSectionName
		// bugs_open/204: the resolver's key space is the component CATALOGUE, and
		// a decomposed page's sections are POSITIONAL slot names (prose-0, tool-1)
		// that are no component's name or function under any spelling. Consulted
		// only after the resolver AND its menu union have both missed, this asks
		// the page's own page_components rows whether it already carries a slot
		// under that name — see stored_slot_rescue.go for why this is not the
		// resolver widening LANDMINES forbids.
		rescue := storedSlotRescueFor(params.DB,
			datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id"),
			params.Logger)
		if len(resolver.validFunctions) > 0 { // only act if components actually loaded
			for _, p := range pages {
				pm, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				sectionsRaw, ok := pm["sections"].([]interface{})
				if !ok {
					continue
				}
				resolved := make([]interface{}, 0, len(sectionsRaw))
				for _, s := range sectionsRaw {
					name, ok := sectionEntryName(s)
					if !ok {
						resolved = append(resolved, s) // nameless entries pass through
						continue
					}
					fn, ok := resolver.resolve(name)
					if !ok {
						pageName, _ := pm["name"].(string)
						switch rescue.verdict(ctx, pageName, name) {
						case slotStored:
							// The page already serves a slot under this name. It is
							// the page's own record of its composition, not junk —
							// keep the entry VERBATIM (object shape and its RFC_016
							// facts intact; rewriting it to the component's function
							// would collapse prose-0/prose-1 onto one name, which is
							// exactly what the positional naming exists to prevent).
							params.Logger.Info("ValidateSitePlanAction: kept section name by stored slot identity",
								zap.Any("page", pm["name"]),
								zap.String("section", name))
							resolved = append(resolved, s)
							continue
						case slotUnknown:
							// Could not read the stored rows. Keep rather than drop:
							// a transient failure must not be able to empty a
							// decomposed page. Logged DISTINCTLY from a real rescue
							// (council f73f4eeb, bug_historian, medium): "kept
							// because recognised" and "kept because unreadable" must
							// never look alike, at any altitude — the read failure
							// also files its own durable row.
							params.Logger.Warn("ValidateSitePlanAction: kept section name WITHOUT checking — stored slot read failed",
								zap.Any("page", pm["name"]),
								zap.String("section", name))
							resolved = append(resolved, s)
							continue
						}
						dropped = append(dropped, droppedSectionName{Page: pageName, Name: name})
						params.Logger.Warn("ValidateSitePlanAction: dropped unresolvable section name",
							zap.Any("page", pm["name"]),
							zap.String("section", name))
						continue
					}
					if fn != name {
						params.Logger.Info("ValidateSitePlanAction: resolved section name to function",
							zap.Any("page", pm["name"]),
							zap.String("from", name),
							zap.String("to", fn))
					}
					if resolver.resolvedViaMenu(fn) {
						// The tell that the menu arm did work. Without it, "the
						// widening reached the plan" and "the planner proposed
						// nothing from the widened class" look identical.
						params.Logger.Info("ValidateSitePlanAction: section resolved via the planner's menu, not the section/element base",
							zap.Any("page", pm["name"]),
							zap.String("section", fn))
					}
					// Preserve the entry's shape: a string stays a string, an
					// object keeps its other keys (facts) with the name resolved.
					if obj, isObj := s.(map[string]interface{}); isObj {
						obj["name"] = fn
						resolved = append(resolved, obj)
					} else {
						resolved = append(resolved, fn)
					}
				}
				pm["sections"] = resolved
			}
			_ = recordDroppedSectionNames(ctx, params, dropped, menuFieldConfigured)
			// The positive tell. Without a durable record of what the rescue KEPT,
			// "the fix works" and "the planner happened to propose only catalogue
			// names this run" produce identical evidence — the resolvedViaMenu
			// lesson from bugs_open/282's council round.
			if f := rescue.keptFinding(); len(f) > 0 {
				siteID := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
				attempted, recorded := LogActionFindings(ctx, params, siteID, "", "validate_plan", f, params.Logger)
				warnUnrecordedDrops(attempted, recorded, params.Logger)
			}
			params.Logger.Info("ValidateSitePlanAction: section name resolution complete",
				zap.Int("dropped", len(dropped)),
				zap.Int("kept_by_stored_slot", rescue.keptCount()),
				zap.Bool("stored_slot_read_failed", rescue.readFailed()))
		} else {
			params.Logger.Warn("ValidateSitePlanAction: validate_components set but no components loaded — skipping name resolution")
		}
	}

	// ── Normalise section entries: strings + per-page section_facts ──────
	// The planner may emit object-form sections ({"name": …, "facts": […]})
	// so plan-time fact assignments (bugs_open/151 candidate 1) survive this
	// action's strips and drops with alignment intact — the assignment
	// travels INSIDE the entry, never as a positionally-keyed sibling.
	// Everything downstream of validate_plan expects plain string sections
	// (sync_pages_to_db serialises the raw array into pages.sections), so
	// the split happens HERE, after the last transformation that can remove
	// an entry: sections becomes strings, and the fact assignments move to a
	// page-level section_facts array aligned by index (null = unscoped), and
	// the subjects to a section_subjects array under the same rule (null = no
	// subject; RFC_016 §5.1's next structured field). A page with no
	// object-form entries is left byte-identical.
	for _, p := range pages {
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		sectionsRaw, ok := pm["sections"].([]interface{})
		if !ok {
			continue
		}
		sawObject := false
		names := make([]interface{}, 0, len(sectionsRaw))
		facts := make([]interface{}, 0, len(sectionsRaw))
		subjects := make([]interface{}, 0, len(sectionsRaw))
		for _, s := range sectionsRaw {
			switch x := s.(type) {
			case string:
				names = append(names, x)
				facts = append(facts, nil)
				subjects = append(subjects, nil)
			case map[string]interface{}:
				name, ok := sectionEntryName(x)
				if !ok {
					continue // nameless object: not a section, drop it
				}
				sawObject = true
				names = append(names, name)
				if f, ok := x["facts"].([]interface{}); ok {
					facts = append(facts, f)
				} else {
					facts = append(facts, nil)
				}
				if subj, ok := x["subject"].(string); ok && strings.TrimSpace(subj) != "" {
					subjects = append(subjects, strings.TrimSpace(subj))
				} else {
					subjects = append(subjects, nil)
				}
			default:
				// unrecognised entry shape: nothing downstream can use it
			}
		}
		if sawObject {
			pm["sections"] = names
			pm["section_facts"] = facts
			pm["section_subjects"] = subjects
			params.Logger.Info("ValidateSitePlanAction: normalised object-form sections",
				zap.Any("page", pm["name"]),
				zap.Int("sections", len(names)))
		}
	}

	// ── Tool pages must have a tool that could fill them (bugs_open/450) ────
	// Opt-in via step config; default false = exactly the prior behaviour. A
	// SIBLING key of enforce_listing_sources below, not a widening of it — the
	// 444 session asked for that in terms, so a tool-arm misfire is switchable
	// without losing the live listing gate.
	//
	// ⚠ RUNS BEFORE THE LISTING GATE, AND THE ORDER IS LOAD-BEARING. Held tool
	// children make a /tools/ hub resolve zero children, so the listing gate
	// below then holds the hub too and no phantom /tools/ URL is planned at all.
	// The reverse order plans an empty hub — a 444-class page — from a plan whose
	// tool pages were just removed. Pinned by TestToolGateRunsBeforeListingGate,
	// with TestListingGateFirstWouldKeepTheEmptyHub as the control that stops that
	// test passing for any order. (Both were written AFTER this comment first
	// claimed them — the council's guardian seat caught the claim, corr 4e7497ed.)
	// Snapshot 2 of 3 (bugs_open/428): the page set as it stands after this
	// action's own identity, truncation and component passes and BEFORE either
	// source gate. Diffing it against the planner's proposal isolates what THIS
	// action removed; diffing it against the final set isolates what the gates
	// held. Without the split, a gate's deliberate, already-recorded hold and a
	// silent identity-collision drop are the same evidence.
	preGatePageViews := planPageViewsOf(pages)

	if enforce, _ := config["enforce_tool_sources"].(bool); enforce {
		pages = enforceToolItemSources(ctx, params, pages, existingPages)
		plan["pages"] = pages
	}

	// ── Listing pages must have a resolvable item source (bugs_open/444) ────
	// Opt-in via step config; default false = exactly the prior behaviour.
	// A listing-family page (news-index, entity-directory, section-index,
	// blog-index, or any page carrying a listing component) whose item source
	// resolves to nothing for THIS site is held out of the plan and filed as a
	// capability_gap naming the missing producer — never built as meta-prose.
	if enforce, _ := config["enforce_listing_sources"].(bool); enforce {
		pages = enforceListingItemSources(ctx, params, pages, existingPages)
		plan["pages"] = pages
	}

	// ── Every strategy-recommended page_type is accounted for (bugs_open/428) ─
	// Record-only and fail-open: it changes no page, adds none, drops none, and
	// cannot fail this step. Runs LAST so `pages` is what the plan writer will
	// receive. Unconditional by design — no tenth optional config key on an
	// action that has no ActionInputSpec — with a fleet-wide env kill switch;
	// see the file header for why that is the arming choice here.
	reconcileRecommendedPageTypes(ctx, params, plan, proposedPageViews, preGatePageViews, planPageViewsOf(pages))

	params.Logger.Info("ValidateSitePlanAction: Complete", zap.Int("pages", len(pages)))
	return plan, nil
}

// recordFactCarryMisses persists one durable row per page whose plan-time fact
// assignments could not be carried onto its realised composition.
//
// Durable, not a log line, and the distinction is the whole point: a name match
// that matches nothing produces exactly the plan a working carry produces, so
// pod output is the only thing that could ever tell them apart — and it does not
// survive long enough to ask afterwards (chassis logs rotate fast; the
// orchestration row that holds collected_data is pruned at ~24h). agent_error_log
// is the channel plan_sections already uses for FACT_SCOPING_EMPTY_COMPOSITION,
// the sibling anomaly one step later in this same feature; querying both by
// error_code is how "did assignment reach the writer?" gets an answer.
//
// Best-effort by construction: a failed write must not change the plan the
// validator has already decided on.
func recordFactCarryMisses(ctx context.Context, params ActionParams, misses []factCarryMiss) {
	if len(misses) == 0 {
		return
	}
	// Best-effort identity: the row is worth writing without it (orchestration_id
	// is filled from the running step either way), so a missing site_id lands as
	// NULL rather than suppressing the record.
	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
	for _, m := range misses {
		LogActionError(ctx, params, siteID, "", "validate_plan",
			"FACT_CARRY_UNMATCHED_SECTION", "warning",
			fmt.Sprintf("page %q: %d plan-time fact assignment(s) named a section the realised composition does not contain — discarded",
				m.Page, len(m.Sections)),
			map[string]interface{}{
				"page":               m.Page,
				"unmatched_sections": m.Sections,
				"remedy":             "the planner scoped facts to section names this built page does not have, usually because it re-composed the page instead of re-emitting the realised section list; check the planner prompt's realised-sections block for this page (RFC_016 candidate 1b)",
			},
			params.Logger)
	}
}

// recordFactAssignmentAbsent persists one durable row per page whose proposed
// object-form section entries resolved a name but carried no usable `facts`
// value. Distinct code from FACT_CARRY_UNMATCHED_SECTION on the same channel:
// unmatched means the planner scoped facts to a section the page does not
// have; absent means it emitted a section with no facts key at all, which
// seed 333 forbids — and which, unrecorded, reads exactly like a page
// correctly assigned no facts (council round a06ff850, objection §3.5).
// Best-effort for the same reason as recordFactCarryMisses.
func recordFactAssignmentAbsent(ctx context.Context, params ActionParams, misses []factCarryMiss) {
	if len(misses) == 0 {
		return
	}
	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
	findings := make([]agenterrors.Finding, 0, len(misses))
	for _, m := range misses {
		findings = append(findings, agenterrors.Finding{
			ErrorCode: "FACT_ASSIGNMENT_ABSENT",
			Severity:  "warning",
			Message: fmt.Sprintf("page %q: %d object-form section entrie(s) carry no usable `facts` value — seed 333 makes the key mandatory, so this is planner disobedience, not a factless page",
				m.Page, len(m.Sections)),
			Context: map[string]interface{}{
				"page":            m.Page,
				"absent_sections": m.Sections,
				"remedy":          "the planner emitted an object-form section entry without a `facts` array; [] is the correct emission for a section with no assigned facts. Check the planner prompt's fact-assignment rules (RFC_016, seed 333)",
			},
		})
	}
	LogActionFindings(ctx, params, siteID, "", "validate_plan", findings, params.Logger)
}

// recordIdentitySnaps persists one durable row per twin-identity event: a plan
// page recognised as denoting an already-realised page under another spelling
// (bugs_open/215 quiet mode), and every refusal and dark-launch observation.
//
// Why durable and not a log line: an active chassis pod retains under one second
// of log (bugs_open/136 §11), so the Info lines these events also emit are not a
// record of anything. Both raw spellings and both URLs are carried, because the
// orchestration row that would otherwise hold the context expires in ~24h — the
// lesson this bug recorded against its own first verification step, when the
// pairing behind its founding incident became permanently unverifiable.
//
// Three codes, because they answer different questions and a single code would
// make the interesting one unfindable:
//
//   - PLAN_PAGE_IDENTITY_SNAPPED: a snap happened. Informational — this is the
//     fix working, and it names which layer fired.
//   - PLAN_PAGE_STEM_TWIN_OBSERVED: the stem layer is OFF and would have fired.
//     This is the dark-launch measurement; a non-zero count is the evidence for
//     (or against) enabling it, and each row is also a phantom about to be
//     minted, so it is warning-severity.
//   - PLAN_PAGE_STEM_TWIN_REFUSED: a stem twin was found and deliberately not
//     acted on. Recorded because silence here would be indistinguishable from
//     "no twin existed", which is the difference between a guard working and a
//     guard being dead.
//
// Best-effort identity, for the same reason as recordFactCarryMisses: the row is
// worth writing without a site_id.
func recordIdentitySnaps(ctx context.Context, params ActionParams, snaps []identitySnap) {
	if len(snaps) == 0 {
		return
	}
	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
	findings := make([]agenterrors.Finding, 0, len(snaps))
	for _, s := range snaps {
		code, severity, message := "PLAN_PAGE_IDENTITY_SNAPPED", "info",
			fmt.Sprintf("plan page %q denotes realised page %q (matched on %s); the realised identity was kept and the plan entry snapped onto it",
				s.PlanName, s.RealisedName, s.Layer)
		remedy := "no action needed — this is the twin-identity reconciliation that stops a second page row being minted for one page (bugs_open/215)"
		switch s.Layer {
		case "path_key_observed", "canonical_name_observed":
			code, severity = "PLAN_PAGE_IDENTITY_TWIN_OBSERVED", "warning"
			message = fmt.Sprintf("plan page %q denotes realised page %q (matched on %s), and twin_identity_snap is OFF — this plan will carry both identities for one page",
				s.PlanName, s.RealisedName, strings.TrimSuffix(s.Layer, "_observed"))
			remedy = "dark-launch signal: each of these is a second page identity about to be written. Read the population, then set twin_identity_snap on the site's structure spec (bugs_open/215)"
		case "stem_twin_observed":
			code, severity = "PLAN_PAGE_STEM_TWIN_OBSERVED", "warning"
			message = fmt.Sprintf("plan page %q is a stem twin of realised page %q, and the stem layer is OFF — this plan will carry both identities for one page",
				s.PlanName, s.RealisedName)
			remedy = "dark-launch signal: each of these is a phantom page row about to be created. Enable stem_twin_snap on the validate step once the observed population has been read (bugs_open/215)"
		case "stem_twin_refused":
			code, severity = "PLAN_PAGE_STEM_TWIN_REFUSED", "warning"
			message = fmt.Sprintf("plan page %q looks like a stem twin of realised page %q and was deliberately NOT snapped: %s",
				s.PlanName, s.RealisedName, s.Reason)
			remedy = "a refusal, not a failure. If the two really are one page, the duplicate needs a remediation decision (which URL survives, what redirects) — see the runbook in bugs_open/215"
		}
		findings = append(findings, agenterrors.Finding{
			ErrorCode: code,
			Severity:  severity,
			Message:   message,
			Context: map[string]interface{}{
				"layer":         s.Layer,
				"plan_name":     s.PlanName,
				"plan_url":      s.PlanURL,
				"realised_name": s.RealisedName,
				"realised_url":  s.RealisedURL,
				"reason":        s.Reason,
				"remedy":        remedy,
			},
		})
	}
	LogActionFindings(ctx, params, siteID, "", "validate_plan", findings, params.Logger)
}

// buildSameNameIdentityFindings turns the same-name stamps and refusals into the
// durable rows a later reader actually needs (bugs_open/215's same-name hole).
// Pure, so the shape of what gets recorded is unit-testable without a database.
//
// ONE SUMMARY ROW, NOT ONE PER PAGE. The stamp is the fix working normally and
// fires on every same-name pairing — measured populations are ~17 pages per
// re-plan on loanandmortgagecalculator.co.uk and ~31 on webdesign.co.uk. A row
// each would bury the events that mean something under the ones that do not, so
// only the DIVERGING stamps (those where the canonicaliser's answer differs from
// the stored name — i.e. where a second identity was actually prevented) are
// summarised into a single row per run, carrying every pair in its context.
//
// TWO CODES, BECAUSE THE SAME OBSERVATION MEANS OPPOSITE THINGS EITHER SIDE OF
// THE FLAG. With honour_realised_identity ON the stored identities will be kept
// and the row is an info-level record that the mechanism engaged. With it OFF the
// same pairs are a warning: both write surfaces will re-derive each of these, and
// each re-derivation is a twin. Reading a count from either code without joining
// would_derive_name back against the pages table conflates a twin about to be
// MINTED with one that already exists and is merely being re-detected — the trap
// this lane already recorded once for the dark-launch counters.
func buildSameNameIdentityFindings(counts reconcileCounts, honour bool) []agenterrors.Finding {
	var findings []agenterrors.Finding

	var diverging []map[string]interface{}
	for _, s := range counts.SameNameStamps {
		if s.WouldDeriveName == "" {
			continue
		}
		diverging = append(diverging, map[string]interface{}{
			"plan_name":         s.PlanName,
			"stored_url":        s.StoredURL,
			"would_derive_name": s.WouldDeriveName,
		})
	}

	if len(diverging) > 0 {
		code, severity := "PLAN_PAGE_SAME_NAME_IDENTITY_HELD", "info"
		message := fmt.Sprintf("%d plan page(s) named exactly as they are stored carry identities the canonicaliser would re-derive; honour_realised_identity is ON, so the stored identities were kept and no twin will be written",
			len(diverging))
		remedy := "no action needed — this is the same-name identity stamp doing its job (bugs_open/215). To confirm at the artefact, check that no new pages row appears for any would_derive_name after this run."
		if !honour {
			code, severity = "PLAN_PAGE_SAME_NAME_TWIN_PENDING", "warning"
			message = fmt.Sprintf("%d plan page(s) named exactly as they are stored carry identities the canonicaliser will re-derive, and honour_realised_identity is OFF — both write surfaces will write each would_derive_name as a SECOND identity for a page that already exists",
				len(diverging))
			remedy = "dark signal, and it is per-page damage, not a summary of it. Join each would_derive_name against pages for this site BEFORE reading a count: a row that already exists means this run re-detects an existing twin, an absent row means this run mints a fresh phantom. Setting honour_realised_identity on the site's structure spec is what stops it (bugs_open/215)."
		}
		findings = append(findings, agenterrors.Finding{
			ErrorCode: code,
			Severity:  severity,
			Message:   message,
			Context: map[string]interface{}{
				"pages":                    diverging,
				"diverging_pages":          len(diverging),
				"same_name_pages_total":    len(counts.SameNameStamps),
				"honour_realised_identity": honour,
				"remedy":                   remedy,
			},
		})
	}

	for _, c := range counts.SameNameTypeConflicts {
		findings = append(findings, agenterrors.Finding{
			ErrorCode: "PLAN_PAGE_IDENTITY_TYPE_CONFLICT",
			Severity:  "warning",
			Message: fmt.Sprintf("plan page %q names a realised page exactly but calls it a %q where the stored page is a %q; its identity was NOT stamped and the write path will re-derive it",
				c.PlanName, c.PlanType, c.RealisedType),
			Context: map[string]interface{}{
				"plan_name":     c.PlanName,
				"plan_type":     c.PlanType,
				"realised_type": c.RealisedType,
				"realised_url":  c.RealisedURL,
				"remedy":        "a refusal, not a failure: honouring would silently retype a live page. If the two really are one page, one of the two types is wrong — fix the stored page_type or the planner's role, and this page keeps today's re-derivation behaviour until then (bugs_open/215).",
			},
		})
	}

	return findings
}

// recordSameNameIdentityOutcomes persists what buildSameNameIdentityFindings
// produced. Takes the honour flag rather than reading it, so the recorder stays
// a function of the run's own facts — the reconciler itself stamps
// unconditionally, and only the writers gate on the flag.
func recordSameNameIdentityOutcomes(ctx context.Context, params ActionParams, counts reconcileCounts, honour bool) {
	findings := buildSameNameIdentityFindings(counts, honour)
	if len(findings) == 0 {
		return
	}
	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
	LogActionFindings(ctx, params, siteID, "", "validate_plan", findings, params.Logger)
}

// recomposeOutcome classifies what actually happened to a page the caller
// explicitly asked to redesign via recompose_pages (features_open/012).
type recomposeOutcome struct {
	Page string
	// Outcome: "proposed_verbatim" — the planner re-emitted the realised
	// composition unchanged, so releasing the page from the preserve guard
	// redesigned nothing (the seed-362 gap: the planner is instructed to
	// re-emit realised sections and is not told which pages are on the
	// recompose list). "absent_from_plan" — the planner omitted the page (or
	// renamed it, which this classifier cannot distinguish); the release
	// makes that a sanctioned drop, but it must be visible, not silent.
	Outcome          string
	RealisedSections int
}

// recomposeOutcomes compares each recompose-requested page's realised
// composition (captured BEFORE filterOutRecomposePages removed it from the
// convergence input) against the reconciled plan. A page that was genuinely
// recomposed produces no outcome — only the two silent shapes are returned.
// Pure so it is unit-testable without a database; the caller records each
// outcome durably.
func recomposeOutcomes(pages []interface{}, recomposeRealised map[string][]interface{}) []recomposeOutcome {
	if len(recomposeRealised) == 0 {
		return nil
	}
	byName := make(map[string]map[string]interface{}, len(pages))
	for _, p := range pages {
		if pm, ok := p.(map[string]interface{}); ok {
			if n, _ := pm["name"].(string); n != "" {
				byName[n] = pm
			}
		}
	}
	var out []recomposeOutcome
	for name, rs := range recomposeRealised {
		pm, present := byName[name]
		switch {
		case !present:
			out = append(out, recomposeOutcome{Page: name, Outcome: "absent_from_plan", RealisedSections: len(rs)})
		case sameSectionList(pm["sections"], rs):
			out = append(out, recomposeOutcome{Page: name, Outcome: "proposed_verbatim", RealisedSections: len(rs)})
		}
	}
	return out
}

// realisedSectionsByName captures the realised composition of the named pages,
// for comparison after the recompose filter has removed them from the
// convergence input (there is nowhere else to read it from afterwards).
func realisedSectionsByName(existingPages []interface{}, names map[string]bool) map[string][]interface{} {
	out := make(map[string][]interface{}, len(names))
	for _, rp := range existingPages {
		rm, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := rm["name"].(string)
		if name == "" || !names[name] {
			continue
		}
		out[name] = realisedSectionsOf(rm)
	}
	return out
}

// recordRecomposeOutcomes persists one durable row per recompose-requested
// page whose explicit redesign intent did not visibly happen. Durable rather
// than a log line for the standing reason (chassis logs rotate sub-second);
// warning severity because both shapes are legal outcomes of the mechanism —
// what they must not be is invisible (owner ruling 2026-08-10, decision 3;
// features_open/012's "loud-signal on a recompose drop" follow-up).
func recordRecomposeOutcomes(ctx context.Context, params ActionParams, outcomes []recomposeOutcome) {
	if len(outcomes) == 0 {
		return
	}
	siteID := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
	findings := make([]agenterrors.Finding, 0, len(outcomes))
	for _, o := range outcomes {
		msg := fmt.Sprintf("page %q was released for redesign via recompose_pages but the plan proposes its realised composition unchanged — the redesign silently no-opped", o.Page)
		remedy := "the planner re-emits realised sections (seed 362) and is not told which pages are on recompose_pages; state the redesign in the briefing the planner sees, or wait for the field-based fix (features_open/012)"
		if o.Outcome == "absent_from_plan" {
			msg = fmt.Sprintf("page %q was released for redesign via recompose_pages and is absent from the reconciled plan — dropped (or renamed) rather than recomposed", o.Page)
			remedy = "a released page the planner omits is dropped by design; if a drop was not intended, re-plan with the page named in the briefing"
		}
		findings = append(findings, agenterrors.Finding{
			ErrorCode: "RECOMPOSE_INTENT_NOT_REALISED",
			Severity:  "warning",
			Message:   msg,
			Context: map[string]interface{}{
				"page":              o.Page,
				"outcome":           o.Outcome,
				"realised_sections": o.RealisedSections,
				"remedy":            remedy,
			},
		})
	}
	LogActionFindings(ctx, params, siteID, "", "validate_plan", findings, params.Logger)
}

// sectionEntryName reads the component name from a planned-section entry in
// either shape the planner emits: a bare string, or an object holding the
// name under one of the historical keys. Key precedence mirrors
// extractSectionEntries in write_site_plan_action.go — the two must agree or
// a name validated here is not the name persisted there.
func sectionEntryName(entry interface{}) (string, bool) {
	switch x := entry.(type) {
	case string:
		if x != "" {
			return x, true
		}
	case map[string]interface{}:
		for _, key := range []string{"name", "type", "component", "component_name"} {
			if name, ok := x[key].(string); ok && name != "" {
				return name, true
			}
		}
	}
	return "", false
}

// nullString returns nil for empty strings, otherwise the string pointer
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// loadSiteChromeNames returns a set of component names that are site-level
// chrome (headers, footers, head) — not page content sections.
func loadSiteChromeNames(ctx context.Context, db *sql.DB, logger *zap.Logger) map[string]bool {
	result := make(map[string]bool)
	rows, err := db.QueryContext(ctx,
		`SELECT name FROM content_components WHERE component_level = 'site' AND is_active = true`)
	if err != nil {
		logger.Warn("loadSiteChromeNames: query failed", zap.Error(err))
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if rows.Scan(&name) == nil {
			result[name] = true
		}
	}
	return result
}

// componentNameResolver resolves plan section names to canonical
// content_components.function values. Used to implement the
// validate_components flag and to normalise gap-planner section names.
type componentNameResolver struct {
	validFunctions map[string]bool   // function -> true
	displayToFunc  map[string]string // lower(display_name) -> function
	nameToFunc     map[string]string // lower(name) -> function
	// menuOnly marks functions that owe their validity to a planner's own
	// component menu rather than to the section/element base below — the ACCEPT
	// half of a menu widening. See component_name_resolver_menu.go (bugs_open/282).
	menuOnly map[string]bool
}

// loadComponentNameResolver loads section/element component identity so
// plan section names can be resolved to a canonical function. Returns an
// empty (non-nil) resolver on error so callers can no-op safely.
func loadComponentNameResolver(ctx context.Context, db *sql.DB, logger *zap.Logger) *componentNameResolver {
	r := &componentNameResolver{
		validFunctions: make(map[string]bool),
		displayToFunc:  make(map[string]string),
		nameToFunc:     make(map[string]string),
	}
	if db == nil {
		return r
	}
	rows, err := db.QueryContext(ctx,
		`SELECT "function", name, COALESCE(display_name, '')
		   FROM content_components
		  WHERE component_level IN ('section','element')
		    AND is_active = true
		    AND "function" <> ''`)
	if err != nil {
		logger.Warn("loadComponentNameResolver: query failed", zap.Error(err))
		return r
	}
	defer rows.Close()
	for rows.Next() {
		var fn, name, display string
		if err := rows.Scan(&fn, &name, &display); err != nil {
			continue
		}
		r.validFunctions[fn] = true
		if name != "" {
			r.nameToFunc[strings.ToLower(name)] = fn
		}
		if display != "" {
			r.displayToFunc[strings.ToLower(display)] = fn
		}
	}
	return r
}

// resolve maps a raw section name to a canonical component function.
// Returns (function, true) if resolved, ("", false) if not. It does NOT
// deduplicate or make content-intent decisions — only name resolution.
func (r *componentNameResolver) resolve(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	// 1. Already a valid function.
	if r.validFunctions[raw] {
		return raw, true
	}
	// 2. Normalise (underscore->hyphen, camelCase->kebab) and re-check.
	norm := NormalizeComponentFunction(raw)
	if norm != raw && r.validFunctions[norm] {
		return norm, true
	}
	// 3. Display-name lookup (handles "FAQ Section" -> "faq").
	if fn, ok := r.displayToFunc[strings.ToLower(raw)]; ok {
		return fn, true
	}
	// 4. Component name lookup (row name differing from function).
	if fn, ok := r.nameToFunc[strings.ToLower(raw)]; ok {
		return fn, true
	}
	// 5. Display lookup on the normalised form.
	if fn, ok := r.displayToFunc[strings.ToLower(norm)]; ok {
		return fn, true
	}
	return "", false
}

// ============================================================================
// ACTION: db_sync
// ============================================================================

// DBSyncAction is a general-purpose database sync action
// Can be used to insert/update records in any table
// Config:
//   - operation: insert, update, upsert
//   - table: table name
//   - data_field: path to data to sync
//   - key_fields: array of fields that form the primary/unique key
func DBSyncAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("DBSyncAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		params.Logger.Warn("DBSyncAction: No database connection")
		return map[string]interface{}{"synced": false, "reason": "no database"}, nil
	}

	operation, _ := config["operation"].(string)
	if operation == "" {
		operation = "upsert"
	}

	table, ok := config["table"].(string)
	if !ok || table == "" {
		return nil, fmt.Errorf("table is required")
	}

	// Get data to sync
	dataField := "sync_data"
	if df, ok := config["data_field"].(string); ok && df != "" {
		dataField = df
	}
	syncData := datahelpers.ExtractNestedField(params.CollectedData, dataField)
	if syncData == nil {
		return nil, fmt.Errorf("no data found at %s", dataField)
	}

	dataMap, ok := syncData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("sync data must be a map")
	}

	// Build and execute query based on operation
	// This is a simplified implementation - in production you'd want more robust SQL building

	params.Logger.Info("DBSyncAction: Syncing data",
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Int("field_count", len(dataMap)),
	)

	// For now, just marshal and log - actual implementation would build SQL
	dataJSON, _ := json.Marshal(dataMap)

	return map[string]interface{}{
		"synced":      true,
		"operation":   operation,
		"table":       table,
		"field_count": len(dataMap),
		"data_size":   len(dataJSON),
	}, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func execDB(ctx context.Context, db interface{}, query string, args ...interface{}) error {
	switch d := db.(type) {
	case *sql.DB:
		_, err := d.ExecContext(ctx, query, args...)
		return err
	case *pgxpool.Pool:
		_, err := d.Exec(ctx, query, args...)
		return err
	default:
		return fmt.Errorf("unsupported database type: %T", db)
	}
}

func lookupPageID(ctx context.Context, db interface{}, siteID uuid.UUID, pageName string, logger *zap.Logger) (uuid.UUID, error) {
	query := `SELECT id FROM pages WHERE site_id = $1 AND name = $2`
	var pageID uuid.UUID

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, siteID, pageName).Scan(&pageID)
		return pageID, err
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, siteID, pageName).Scan(&pageID)
		return pageID, err
	default:
		return uuid.Nil, fmt.Errorf("unsupported database type: %T", db)
	}
}

// queryRowScanUUID executes a query and scans the result into a UUID
func queryRowScanUUID(ctx context.Context, db interface{}, query string, dest *uuid.UUID, args ...interface{}) error {
	switch d := db.(type) {
	case *sql.DB:
		return d.QueryRowContext(ctx, query, args...).Scan(dest)
	case *pgxpool.Pool:
		return d.QueryRow(ctx, query, args...).Scan(dest)
	default:
		return fmt.Errorf("unsupported database type: %T", db)
	}
}

// getStyleCollectionByID looks up a style collection by UUID
// This is a local helper since component_library.go doesn't have this function yet
func getStyleCollectionByID(ctx context.Context, db interface{}, id uuid.UUID, logger *zap.Logger) (*StyleCollection, error) {
	query := `
		SELECT id, name, display_name, category, color_palette, typography,
		       header_component_id, footer_component_id, css_theme_id
		FROM style_collections
		WHERE id = $1 AND is_active = true
	`

	var coll StyleCollection
	var colorPaletteJSON, typographyJSON []byte
	var headerID, footerID, cssThemeID *uuid.UUID

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, id).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName, &coll.Category,
			&colorPaletteJSON, &typographyJSON,
			&headerID, &footerID, &cssThemeID,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, id).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName, &coll.Category,
			&colorPaletteJSON, &typographyJSON,
			&headerID, &footerID, &cssThemeID,
		)
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if colorPaletteJSON != nil {
		json.Unmarshal(colorPaletteJSON, &coll.ColorPalette)
	}
	if typographyJSON != nil {
		json.Unmarshal(typographyJSON, &coll.Typography)
	}
	coll.HeaderComponentID = headerID
	coll.FooterComponentID = footerID
	coll.CSSThemeID = cssThemeID

	return &coll, nil
}

func BuildReviewResultAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("BuildReviewResultAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	result := map[string]interface{}{
		"reviewed_at": time.Now().UTC().Format(time.RFC3339),
		"review_mode": "unknown",
		"approved":    false,
		"issues":      []interface{}{},
		"edits":       map[string]interface{}{},
	}

	if approved, ok := config["approved"].(bool); ok {
		result["approved"] = approved
	} else if field, ok := config["approved_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			if b, ok := val.(bool); ok {
				result["approved"] = b
			}
		}
	}

	if reviewer, ok := config["reviewer"].(string); ok && reviewer != "" {
		result["reviewed_by"] = reviewer
	} else if field, ok := config["reviewer_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedFieldString(params.CollectedData, field); val != "" {
			result["reviewed_by"] = val
		}
	}
	if result["reviewed_by"] == nil {
		result["reviewed_by"] = "system"
	}

	if mode, ok := config["review_mode"].(string); ok {
		result["review_mode"] = mode
	}

	if field, ok := config["eval_score"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			result["eval_score"] = val
		}
	}

	if field, ok := config["edits_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			result["edits"] = val
		}
	}

	if field, ok := config["auto_eval_issues"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			result["auto_eval_issues"] = val
			if issues, ok := val.([]interface{}); ok {
				result["issues"] = issues
			}
		}
	}

	if content := datahelpers.ExtractNestedField(params.CollectedData, "page_content"); content != nil {
		result["content"] = content
	}

	params.Logger.Info("BuildReviewResultAction: Complete",
		zap.Bool("approved", result["approved"].(bool)),
		zap.String("mode", result["review_mode"].(string)))
	return result, nil
}

// ============================================================================
// ACTION: prepare_review_data
// ============================================================================

func PrepareReviewDataAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("PrepareReviewDataAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	reviewData := map[string]interface{}{
		"prepared_at": time.Now().UTC().Format(time.RFC3339),
		"fields":      map[string]interface{}{},
	}

	if includeFields, ok := config["include_fields"].([]interface{}); ok {
		fieldsMap := reviewData["fields"].(map[string]interface{})
		for _, field := range includeFields {
			if fieldName, ok := field.(string); ok {
				if value := datahelpers.ExtractNestedField(params.CollectedData, fieldName); value != nil {
					fieldsMap[fieldName] = value
				}
			}
		}
	}

	if formatForDisplay, ok := config["format_for_display"].(bool); ok && formatForDisplay {
		if page := datahelpers.ExtractNestedField(params.CollectedData, "current_page"); page != nil {
			if pm, ok := page.(map[string]interface{}); ok {
				reviewData["page_name"] = pm["name"]
				reviewData["page_title"] = pm["title"]
			}
		}
		if content := datahelpers.ExtractNestedField(params.CollectedData, "page_content"); content != nil {
			reviewData["content"] = content
		}
		if brief := datahelpers.ExtractNestedField(params.CollectedData, "reviewed_brief"); brief != nil {
			if bm, ok := brief.(map[string]interface{}); ok {
				reviewData["company_name"] = bm["company_name"]
				reviewData["tone"] = bm["tone"]
			}
		}
	}

	params.Logger.Info("PrepareReviewDataAction: Complete")
	return reviewData, nil
}

// ============================================================================
// ACTION: update_page_components_status
// ============================================================================

func UpdatePageComponentsStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdatePageComponentsStatusAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	newStatus := "approved"
	if status, ok := config["status"].(string); ok && status != "" {
		newStatus = status
	}

	pageField := "current_page"
	if pf, ok := config["page_from"].(string); ok && pf != "" {
		pageField = pf
	}

	pageData := datahelpers.ExtractNestedField(params.CollectedData, pageField)
	if pageData == nil {
		return map[string]interface{}{"updated": false, "reason": "no page data", "status_set": newStatus}, nil
	}

	page, ok := pageData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("page must be object")
	}

	var pageID uuid.UUID
	if idStr, ok := page["id"].(string); ok && idStr != "" {
		pageID, _ = uuid.Parse(idStr)
	}

	reviewedAt := time.Now().UTC()
	if field, ok := config["reviewed_at_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedFieldString(params.CollectedData, field); val != "" {
			if t, err := time.Parse(time.RFC3339, val); err == nil {
				reviewedAt = t
			}
		}
	}

	reviewedBy := "system"
	if field, ok := config["reviewed_by_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedFieldString(params.CollectedData, field); val != "" {
			reviewedBy = val
		}
	}

	if params.DB != nil && pageID != uuid.Nil {
		query := `UPDATE page_components SET build_status = $1, reviewed_at = $2, reviewed_by = $3, updated_at = NOW() WHERE page_id = $4`
		result, err := params.DB.ExecContext(ctx, query, newStatus, reviewedAt, reviewedBy, pageID)
		if err != nil {
			return nil, fmt.Errorf("failed to update: %w", err)
		}
		rows, _ := result.RowsAffected()
		return map[string]interface{}{
			"updated": true, "rows_affected": rows, "page_id": pageID.String(),
			"status_set": newStatus, "reviewed_at": reviewedAt.Format(time.RFC3339), "reviewed_by": reviewedBy,
		}, nil
	}

	return map[string]interface{}{
		"updated": false, "reason": "no db or page_id", "status_set": newStatus,
		"reviewed_at": reviewedAt.Format(time.RFC3339), "reviewed_by": reviewedBy,
	}, nil
}

// ============================================================================
// SHARED SECTION-COMPONENT LOADER
// ============================================================================
//
// loadSectionComponents is the canonical loader for component rows used by
// section-level callers. Extracted from LoadPageSectionComponentsAction so
// plan_sections can reuse the same logic without a second SQL path. Both
// callers get the same component-row shape; differences in behaviour are
// expressed by what each caller does with the returned data, not by what
// each caller queries.
//
// Behaviour matches the previous in-action implementation:
//   - Match by name first, fall back to function (DISTINCT ON, newest first)
//   - Stubs for sections with no matching component
//   - Order preserved relative to sectionNames input
//   - When pageID != "", content_brief is attached per slot
//   - When activeOnly is true, only is_active=true rows are returned
//
// activeOnly preserves the historical behaviour difference between the two
// callers: plan_sections used to filter `is_active = true` inline;
// LoadPageSectionComponentsAction did not. Passing the flag explicitly keeps
// both callers' behaviour intact while sharing one query path.
//
// The returned per-component map carries: component_id (when from DB),
// name, function, display_name, category, semantic_tags (when set),
// description (when set), html_template (when set), input_schema (when
// set, as raw JSON string), render_mode, agent_type (when set),
// component_level, needs_llm, and content_brief (when found).
func loadSectionComponents(
	ctx context.Context,
	db *sql.DB,
	sectionNames []string,
	pageID string,
	activeOnly bool,
	logger *zap.Logger,
) []map[string]interface{} {
	if len(sectionNames) == 0 {
		return []map[string]interface{}{}
	}
	if db == nil {
		// No DB available — return name-stubs so callers can still proceed.
		return buildStubSectionComponents(sectionNames)
	}

	// Match each requested section against BOTH its raw name and its kebab-
	// normalised form (bugs_open/041): the library stores kebab-case, but plans
	// may emit snake_case/CamelCase ("call_to_action"). This value set is a
	// strict superset of the raw names, so nothing that resolved before stops
	// resolving — including the few components whose *name* is itself snake_case.
	lookupValues := sectionLookupValueSet(sectionNames)
	placeholders := make([]string, len(lookupValues))
	args := make([]interface{}, len(lookupValues))
	for i, v := range lookupValues {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = v
	}

	var components []map[string]interface{}

	// Pass 1: lookup by name
	activeFilter := ""
	if activeOnly {
		activeFilter = " AND is_active = true"
	}
	nameQuery := fmt.Sprintf(`
		SELECT
			id,
			name,
			COALESCE(display_name, name) AS display_name,
			function,
			COALESCE(category, '') AS category,
			semantic_tags,
			description,
			html_template,
			input_schema,
			COALESCE(render_mode, 'template') AS render_mode,
			agent_type,
			COALESCE(component_level, 'section') AS component_level
		FROM content_components
		WHERE name IN (%s)%s`, strings.Join(placeholders, ", "), activeFilter)

	rows, err := db.QueryContext(ctx, nameQuery, args...)
	if err != nil {
		logger.Error("loadSectionComponents: name query failed",
			zap.Error(err),
			zap.Strings("sections", sectionNames))
	} else {
		for rows.Next() {
			comp, scanErr := scanSectionComponentRow(rows)
			if scanErr != nil {
				logger.Error("loadSectionComponents: name row scan failed",
					zap.Error(scanErr))
				continue
			}
			components = append(components, comp)
		}
		rows.Close()
	}

	// Track which inputs are already satisfied by name or function
	foundNames := make(map[string]bool)
	for _, comp := range components {
		if n, ok := comp["name"].(string); ok {
			foundNames[n] = true
		}
		if fn, ok := comp["function"].(string); ok {
			foundNames[fn] = true
		}
	}

	var missing []string
	for _, name := range sectionNames {
		if !sectionResolvedByFound(foundNames, name) {
			missing = append(missing, name)
		}
	}

	// Pass 2: lookup by function for anything still missing
	if len(missing) > 0 {
		logger.Info("loadSectionComponents: trying function lookup for missing",
			zap.Strings("missing", missing))

		// Same raw+normalised superset as Pass 1 (bugs_open/041).
		funcValues := sectionLookupValueSet(missing)
		funcValueSet := make(map[string]bool, len(funcValues))
		funcPlaceholders := make([]string, len(funcValues))
		funcArgs := make([]interface{}, len(funcValues))
		for i, v := range funcValues {
			funcValueSet[v] = true
			funcPlaceholders[i] = fmt.Sprintf("$%d", i+1)
			funcArgs[i] = v
		}

		funcQuery := fmt.Sprintf(`
			SELECT DISTINCT ON (function)
				id,
				name,
				COALESCE(display_name, name) AS display_name,
				function,
				COALESCE(category, '') AS category,
				semantic_tags,
				description,
				html_template,
				input_schema,
				COALESCE(render_mode, 'template') AS render_mode,
				agent_type,
				COALESCE(component_level, 'section') AS component_level
			FROM content_components
			WHERE function IN (%s)%s
			ORDER BY function, created_at DESC
		`, strings.Join(funcPlaceholders, ", "), activeFilter)

		funcRows, ferr := db.QueryContext(ctx, funcQuery, funcArgs...)
		if ferr != nil {
			logger.Warn("loadSectionComponents: function lookup failed",
				zap.Error(ferr))
		} else {
			for funcRows.Next() {
				comp, scanErr := scanSectionComponentRow(funcRows)
				if scanErr != nil {
					continue
				}
				function, _ := comp["function"].(string)
				if !funcValueSet[function] {
					continue
				}
				components = append(components, comp)
				foundNames[function] = true
				logger.Info("loadSectionComponents: found component by function",
					zap.String("function", function),
					zap.String("name", comp["name"].(string)))
			}
			funcRows.Close()
		}
	}

	// Stubs for anything still not found
	var stillMissing []string
	for _, name := range sectionNames {
		if !sectionResolvedByFound(foundNames, name) {
			stillMissing = append(stillMissing, name)
		}
	}
	if len(stillMissing) > 0 {
		logger.Warn("loadSectionComponents: stubs for unresolved sections",
			zap.Strings("missing", stillMissing))
		for _, name := range stillMissing {
			components = append(components, map[string]interface{}{
				"name":         name,
				"display_name": name,
				"function":     name,
				"category":     "",
				"needs_llm":    true,
				"description":  "",
			})
		}
	}

	// Reorder to match sectionNames input order. Match a component to a requested
	// section under either the raw or normalised form (bugs_open/041), mirroring
	// the lookup above so a resolved "call_to_action" lands in its slot.
	ordered := make([]map[string]interface{}, 0, len(components))
	for _, sectionName := range sectionNames {
		keys := sectionLookupKeys(sectionName)
		for _, comp := range components {
			name, _ := comp["name"].(string)
			function, _ := comp["function"].(string)
			if containsString(keys, name) || containsString(keys, function) {
				ordered = append(ordered, comp)
				break
			}
		}
	}

	// Optional: content_brief enrichment from page_components
	if pageID != "" {
		enrichSectionComponentsWithBriefs(ctx, db, pageID, ordered, logger)
	}

	return ordered
}

// scanSectionComponentRow turns one SQL row into the per-component map shape.
// Centralised so the by-name and by-function passes produce identical shapes.
func scanSectionComponentRow(rows *sql.Rows) (map[string]interface{}, error) {
	var id, name, function string
	var displayName, category sql.NullString
	var semanticTags, description, htmlTemplate, inputSchema sql.NullString
	var renderMode, agentType, componentLevel sql.NullString

	if err := rows.Scan(
		&id, &name, &displayName, &function, &category,
		&semanticTags, &description, &htmlTemplate, &inputSchema,
		&renderMode, &agentType, &componentLevel,
	); err != nil {
		return nil, err
	}

	comp := map[string]interface{}{
		"component_id": id,
		"name":         name,
		"function":     function,
	}
	if displayName.Valid {
		comp["display_name"] = displayName.String
	} else {
		comp["display_name"] = name
	}
	if category.Valid {
		comp["category"] = category.String
	} else {
		comp["category"] = ""
	}
	if semanticTags.Valid {
		comp["semantic_tags"] = semanticTags.String
	}
	if description.Valid && description.String != "" {
		comp["description"] = description.String
	}
	if htmlTemplate.Valid && htmlTemplate.String != "" {
		comp["html_template"] = htmlTemplate.String
	}
	if inputSchema.Valid && inputSchema.String != "" {
		comp["input_schema"] = inputSchema.String
	}
	if renderMode.Valid && renderMode.String != "" {
		comp["render_mode"] = renderMode.String
	} else {
		comp["render_mode"] = "template"
	}
	if agentType.Valid && agentType.String != "" {
		comp["agent_type"] = agentType.String
	}
	if componentLevel.Valid && componentLevel.String != "" {
		comp["component_level"] = componentLevel.String
	} else {
		comp["component_level"] = "section"
	}
	comp["needs_llm"] = detectNeedsLLMContent(htmlTemplate.String, inputSchema.String)
	return comp, nil
}

// loadContentComponentsByID loads components directly by id, the identity a
// page_components row actually carries (component_id) rather than the
// name/function loadSectionComponents matches on. Built for bugs_open/182:
// when a site's slot_name is positional ("prose-0", "tool-2") it is not a
// component identity at all, so no name/function lookup can ever resolve it —
// only the id the row already has can. Reuses scanSectionComponentRow so the
// row shape is identical to the name/function path (both feed
// componentInfoFromRaw).
//
// Only is_active rows are returned — an id pointing at a retired component is
// treated the same as "not found" rather than resurrecting retired markup;
// the caller's fallback (name/function lookup) still gets a chance.
func loadContentComponentsByID(ctx context.Context, db *sql.DB, componentIDs []string, logger *zap.Logger) []map[string]interface{} {
	if len(componentIDs) == 0 || db == nil {
		return nil
	}

	query := `
		SELECT
			id,
			name,
			COALESCE(display_name, name) AS display_name,
			function,
			COALESCE(category, '') AS category,
			semantic_tags,
			description,
			html_template,
			input_schema,
			COALESCE(render_mode, 'template') AS render_mode,
			agent_type,
			COALESCE(component_level, 'section') AS component_level
		FROM content_components
		WHERE id = ANY($1::uuid[]) AND is_active = true`

	rows, err := db.QueryContext(ctx, query, toPGTextArrayLiteral(componentIDs))
	if err != nil {
		logger.Error("loadContentComponentsByID: query failed",
			zap.Error(err), zap.Int("requested", len(componentIDs)))
		return nil
	}
	defer rows.Close()

	var out []map[string]interface{}
	for rows.Next() {
		comp, scanErr := scanSectionComponentRow(rows)
		if scanErr != nil {
			logger.Error("loadContentComponentsByID: row scan failed", zap.Error(scanErr))
			continue
		}
		out = append(out, comp)
	}
	if err := rows.Err(); err != nil {
		logger.Error("loadContentComponentsByID: row iteration failed", zap.Error(err))
	}
	return out
}

// buildStubSectionComponents returns minimal stubs for the no-DB code path.
func buildStubSectionComponents(sectionNames []string) []map[string]interface{} {
	stubs := make([]map[string]interface{}, len(sectionNames))
	for i, name := range sectionNames {
		stubs[i] = map[string]interface{}{
			"name":         name,
			"function":     name,
			"display_name": name,
			"description":  "",
			"needs_llm":    true,
		}
	}
	return stubs
}

// enrichSectionComponentsWithBriefs attaches per-section admin content briefs
// (from page_components.content_brief) onto the components in-place.
func enrichSectionComponentsWithBriefs(
	ctx context.Context,
	db *sql.DB,
	pageID string,
	components []map[string]interface{},
	logger *zap.Logger,
) {
	briefRows, briefErr := db.QueryContext(ctx, `
		SELECT COALESCE(slot_name, ''), content_brief
		FROM page_components
		WHERE page_id = $1
		  AND content_brief IS NOT NULL
		  AND `+datahelpers.NotRemoved("")+`
	`, pageID)
	if briefErr != nil {
		logger.Warn("enrichSectionComponentsWithBriefs: query failed",
			zap.Error(briefErr))
		return
	}
	defer briefRows.Close()

	briefMap := make(map[string]interface{})
	for briefRows.Next() {
		var slotName string
		var briefJSON []byte
		if err := briefRows.Scan(&slotName, &briefJSON); err != nil {
			continue
		}
		if len(briefJSON) > 0 && slotName != "" {
			var brief interface{}
			if err := json.Unmarshal(briefJSON, &brief); err == nil {
				briefMap[slotName] = brief
			}
		}
	}
	if len(briefMap) == 0 {
		return
	}

	for _, comp := range components {
		name, _ := comp["name"].(string)
		function, _ := comp["function"].(string)
		if brief, ok := briefMap[name]; ok {
			comp["content_brief"] = brief
		} else if brief, ok := briefMap[function]; ok {
			comp["content_brief"] = brief
		}
	}
	logger.Info("enrichSectionComponentsWithBriefs: attached briefs",
		zap.Int("briefs_found", len(briefMap)))
}

// ============================================================================
// ACTION: load_page_section_components
// ============================================================================
//
// LoadPageSectionComponentsAction loads component definitions for a page's
// sections. Thin workflow wrapper around loadSectionComponents.
//
// Config:
//   - page_from: collected_data path to the page record (default "current_page")
//   - include_templates, include_input_schema: kept for back-compat; the shared
//     loader always returns both, so these flags are not consulted. See
//     doc 019_tool_library.md for the rationale.
//
// When call_agent passes input_fields, they arrive under input_data.*
// extractWithInputDataFallback handles both root and input_data locations.
func LoadPageSectionComponentsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("LoadPageSectionComponentsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	pageField := "current_page"
	if pf, ok := config["page_from"].(string); ok && pf != "" {
		pageField = pf
	}

	pageData := extractWithInputDataFallback(params.CollectedData, pageField, params.Logger)
	if pageData == nil {
		keys := make([]string, 0, len(params.CollectedData))
		for k := range params.CollectedData {
			keys = append(keys, k)
		}
		params.Logger.Error("Page not found",
			zap.String("page_field", pageField),
			zap.Strings("available_keys", keys))
		return nil, fmt.Errorf("page not found at '%s'", pageField)
	}

	page, ok := pageData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("page must be object")
	}

	params.Logger.Info("Found page data",
		zap.String("page_field", pageField),
		zap.Any("page_name", page["name"]),
		zap.Any("page_title", page["title"]))

	sectionsRaw := page["sections"]
	if sectionsRaw == nil {
		sectionsRaw = extractWithInputDataFallback(params.CollectedData, "sections", params.Logger)
	}

	sectionNames := datahelpers.ExtractSectionNamesHelper(sectionsRaw)
	if len(sectionNames) == 0 {
		params.Logger.Warn("No sections found for page")
		return map[string]interface{}{
			"components":    []interface{}{},
			"count":         0,
			"from_database": false,
		}, nil
	}

	// Enforce naming contract: LLM site plans may output "social_proof" or "SocialProof"
	NormalizeSectionNames(sectionNames, params.Logger)

	params.Logger.Info("Loading components for sections",
		zap.Strings("sections", sectionNames))

	pageID, _ := page["id"].(string)
	// activeOnly=false: this action historically returned components regardless
	// of is_active. Preserved here so existing callers (page-content-writer's
	// load_page_components step, audit flows, admin tools) behave identically.
	components := loadSectionComponents(ctx, params.DB, sectionNames, pageID, false, params.Logger)

	// Detect whether any row carries a real component_id (i.e. came from DB)
	// to preserve the from_database signal callers rely on.
	fromDB := false
	for _, c := range components {
		if _, has := c["component_id"]; has {
			fromDB = true
			break
		}
	}

	return map[string]interface{}{
		"components":    components,
		"count":         len(components),
		"from_database": fromDB,
		"requested":     sectionNames,
	}, nil
}

// Order: direct path -> input_data.{path} -> FindByPath helper
// extractWithInputDataFallback tries to extract a field, falling back to input_data prefix
// This handles the common case where workflows specify paths like "current_section.name"
// but the data is actually at "input_data.current_section.name"
func extractWithInputDataFallback(data map[string]interface{}, path string, logger *zap.Logger) interface{} {
	// Try direct path first
	if value := datahelpers.ExtractNestedField(data, path); value != nil {
		logger.Debug("extractWithInputDataFallback: Found at direct path",
			zap.String("path", path),
		)
		return value
	}

	// If path doesn't already start with input_data, try with prefix
	if !strings.HasPrefix(path, "input_data.") {
		prefixedPath := "input_data." + path
		if value := datahelpers.ExtractNestedField(data, prefixedPath); value != nil {
			logger.Debug("extractWithInputDataFallback: Found via input_data prefix",
				zap.String("original_path", path),
				zap.String("actual_path", prefixedPath),
			)
			return value
		}
	}

	// Try in __raw_message__.body.input_data (deeply nested case from child agents)
	if !strings.HasPrefix(path, "__raw_message__") {
		rawMsgPath := "__raw_message__.body.input_data." + path
		if value := datahelpers.ExtractNestedField(data, rawMsgPath); value != nil {
			logger.Debug("extractWithInputDataFallback: Found via __raw_message__.body.input_data",
				zap.String("original_path", path),
			)
			return value
		}
	}

	// Also try agent_config location (for workflow config data)
	if !strings.HasPrefix(path, "agent_config") {
		agentConfigPath := "agent_config." + path
		if value := datahelpers.ExtractNestedField(data, agentConfigPath); value != nil {
			logger.Debug("extractWithInputDataFallback: Found via agent_config",
				zap.String("original_path", path),
			)
			return value
		}
	}

	logger.Debug("extractWithInputDataFallback: Not found anywhere",
		zap.String("path", path),
	)
	return nil
}

// ============================================================================
// ACTION: filter_search_results
// ============================================================================

// FilterSearchResultsAction filters search results based on criteria
// Handles various response formats: direct array, {results: []}, {data: {results: []}}, etc.
func FilterSearchResultsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("FilterSearchResultsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	resultsField := "search_results"
	if rf, ok := config["results_field"].(string); ok && rf != "" {
		resultsField = rf
	}

	// Try to find results array - handles various response formats
	results := datahelpers.FindResultsArray(params.CollectedData, resultsField, params.Logger)
	if results == nil {
		params.Logger.Warn("FilterSearchResultsAction: No results found",
			zap.String("results_field", resultsField),
			zap.Strings("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)))
		return map[string]interface{}{"filtered_results": []interface{}{}, "count": 0, "original_count": 0}, nil
	}

	params.Logger.Info("FilterSearchResultsAction: Found results",
		zap.Int("count", len(results)))

	// Support both max_results and max_sources config keys
	maxResults := 10
	if mr, ok := config["max_results"].(float64); ok {
		maxResults = int(mr)
	} else if ms, ok := config["max_sources"].(float64); ok {
		maxResults = int(ms)
	}

	excludePatterns := datahelpers.ExtractStringListHelper(config["exclude_patterns"])
	requiredKeywords := datahelpers.ExtractStringListHelper(config["required_keywords"])
	preferDomains := datahelpers.ExtractStringListHelper(config["prefer_domains"])

	var filtered []interface{}
	var preferred []interface{}

	for _, r := range results {
		result, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		title, _ := result["title"].(string)
		content, _ := result["content"].(string)
		snippet, _ := result["snippet"].(string)
		url, _ := result["url"].(string)
		searchText := strings.ToLower(title + " " + content + " " + snippet + " " + url)

		// Check exclusions
		excluded := false
		for _, pattern := range excludePatterns {
			if strings.Contains(searchText, strings.ToLower(pattern)) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Check required keywords
		if len(requiredKeywords) > 0 {
			hasKeyword := false
			for _, kw := range requiredKeywords {
				if strings.Contains(searchText, strings.ToLower(kw)) {
					hasKeyword = true
					break
				}
			}
			if !hasKeyword {
				continue
			}
		}

		// Check if from preferred domain
		isPreferred := false
		for _, domain := range preferDomains {
			if strings.Contains(strings.ToLower(url), strings.ToLower(domain)) {
				isPreferred = true
				break
			}
		}

		if isPreferred {
			preferred = append(preferred, result)
		} else {
			filtered = append(filtered, result)
		}
	}

	// Combine: preferred first, then others, up to maxResults
	combined := append(preferred, filtered...)
	if len(combined) > maxResults {
		combined = combined[:maxResults]
	}

	params.Logger.Info("FilterSearchResultsAction: Complete",
		zap.Int("original", len(results)),
		zap.Int("preferred", len(preferred)),
		zap.Int("filtered", len(combined)))

	return map[string]interface{}{
		"filtered_results": combined,
		"count":            len(combined),
		"original_count":   len(results),
		"preferred_count":  len(preferred),
	}, nil
}

// ============================================================================
// ACTION: extract_fields
// ============================================================================

func ExtractFieldsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ExtractFieldsAction: Starting",
		zap.Any("config_keys", getConfigKeys(params.StepConfig.Config)))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	result := make(map[string]interface{})

	// Handle fields array format: ["field1", "field2"]
	if fields, ok := config["fields"].([]interface{}); ok {
		params.Logger.Info("ExtractFieldsAction: Processing fields as array")
		for _, f := range fields {
			switch field := f.(type) {
			case string:
				if value := datahelpers.ExtractNestedField(params.CollectedData, field); value != nil {
					parts := strings.Split(field, ".")
					result[parts[len(parts)-1]] = value
				}
			case map[string]interface{}:
				source, _ := field["source"].(string)
				target, _ := field["target"].(string)
				if source == "" {
					continue
				}
				if target == "" {
					parts := strings.Split(source, ".")
					target = parts[len(parts)-1]
				}
				if value := datahelpers.ExtractNestedField(params.CollectedData, source); value != nil {
					result[target] = value
				}
			}
		}
	}

	// Handle fields as map-of-arrays format (fallback paths)
	// Example: {"topic": ["path1", "path2"], "company": ["path3"]}
	if fieldsMap, ok := config["fields"].(map[string]interface{}); ok {
		params.Logger.Info("ExtractFieldsAction: Processing fields as map-of-arrays",
			zap.Int("field_count", len(fieldsMap)))

		for targetField, pathsRaw := range fieldsMap {
			var found bool

			// Handle array of paths
			if paths, ok := pathsRaw.([]interface{}); ok {
				for _, pathRaw := range paths {
					if path, ok := pathRaw.(string); ok {
						// Try direct path first
						if value := datahelpers.ExtractNestedField(params.CollectedData, path); value != nil {
							result[targetField] = value
							found = true
							params.Logger.Info("ExtractFieldsAction: Found via direct path",
								zap.String("target", targetField),
								zap.String("path", path))
							break
						}

						// Try with input_data prefix
						if !strings.HasPrefix(path, "input_data.") {
							prefixedPath := "input_data." + path
							if value := datahelpers.ExtractNestedField(params.CollectedData, prefixedPath); value != nil {
								result[targetField] = value
								found = true
								params.Logger.Info("ExtractFieldsAction: Found via input_data prefix",
									zap.String("target", targetField),
									zap.String("original_path", path),
									zap.String("prefixed_path", prefixedPath))
								break
							}
						}
					}
				}
			}

			// Handle single string path (not array)
			if singlePath, ok := pathsRaw.(string); ok && !found {
				if value := datahelpers.ExtractNestedField(params.CollectedData, singlePath); value != nil {
					result[targetField] = value
					found = true
				} else if !strings.HasPrefix(singlePath, "input_data.") {
					prefixedPath := "input_data." + singlePath
					if value := datahelpers.ExtractNestedField(params.CollectedData, prefixedPath); value != nil {
						result[targetField] = value
						found = true
					}
				}
			}

			if !found {
				params.Logger.Warn("ExtractFieldsAction: Field not found in any path",
					zap.String("target", targetField),
					zap.Any("tried_paths", pathsRaw))
			}
		}
	}

	// Handle field_map format: {"target": "source"}
	if fieldMap, ok := config["field_map"].(map[string]interface{}); ok {
		for target, source := range fieldMap {
			if sourceStr, ok := source.(string); ok {
				if value := datahelpers.ExtractNestedField(params.CollectedData, sourceStr); value != nil {
					result[target] = value
				}
			}
		}
	}

	// Apply defaults
	if defaults, ok := config["defaults"].(map[string]interface{}); ok {
		for key, val := range defaults {
			if result[key] == nil {
				result[key] = val
			}
		}
	}

	// Opt-in strictness (bugs_open/192). By default this action reports success
	// having silently omitted any target whose configured paths all missed —
	// so the failure surfaces later, somewhere else, named after the symptom.
	// That is how a shape change upstream cost the fleet every page build for
	// hours: select_sections wrote an empty sections_for_render and the NEXT
	// step failed with "key 'sections_ready' not found", which points at the
	// wrong file.
	//
	// "required": ["field", ...] makes that state fail HERE, naming the cause,
	// the paths tried, and what was actually in scope to try them against.
	//
	// Default OFF: an absent "required" key preserves the historical lenient
	// behaviour exactly, so no existing step changes meaning. Per the OWNER
	// RULING of 2026-08-02 §2, new authority on a shared seam ships as an
	// opt-in FIELD with the unsafe default off, not as a doc comment policed by
	// review — a caller's own reviewer can see this one. Checked AFTER defaults,
	// so a supplied default legitimately satisfies a required field.
	if requiredRaw, ok := config["required"].([]interface{}); ok {
		// A name in "required" that matches no CONFIGURED target is a config
		// typo, and it fails in the most confusing possible way: the field can
		// never be produced, so the step fails every single run with a message
		// about a field the reader cannot find in "fields". Say what it really
		// is instead. Council advisory, round 2 of bugs_open/192.
		configuredTargets := map[string]interface{}{}
		for _, source := range []string{"fields", "field_map", "defaults"} {
			if m, ok := config[source].(map[string]interface{}); ok {
				for target := range m {
					configuredTargets[target] = true
				}
			}
		}

		var missing []string
		for _, r := range requiredRaw {
			name, _ := r.(string)
			if name == "" {
				continue
			}
			if _, configured := configuredTargets[name]; len(configuredTargets) > 0 && !configured {
				return nil, fmt.Errorf(
					"extract_fields: required field %q names no configured target "+
						"(configured targets: %v) — this is a step-config error, not a data problem",
					name, datahelpers.GetMapKeys(configuredTargets))
			}
			if value, present := result[name]; !present || value == nil {
				missing = append(missing, name)
			}
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf(
				"extract_fields: required field(s) %v resolved via no configured path "+
					"(configured fields: %v; collected_data top-level keys: %v)",
				missing, config["fields"], datahelpers.GetMapKeys(params.CollectedData))
		}
	}

	params.Logger.Info("ExtractFieldsAction: Complete",
		zap.Int("fields_extracted", len(result)),
		zap.Strings("result_keys", datahelpers.GetMapKeys(result)))
	return result, nil
}

// Priority 5: Try to build query from section context (for research-agent)
// Check both root level and inside input_data
func getSearchQueryFromSectionContext(params ActionParams) string {
	var currentSection map[string]interface{}

	// Try root level first
	if cs, ok := params.CollectedData["current_section"].(map[string]interface{}); ok {
		currentSection = cs
	}
	// Try input_data.current_section
	if currentSection == nil {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			if cs, ok := inputData["current_section"].(map[string]interface{}); ok {
				currentSection = cs
			}
		}
	}

	if currentSection == nil {
		return ""
	}

	// Get section name/function
	sectionName := ""
	if fn, ok := currentSection["function"].(string); ok && fn != "" {
		sectionName = fn
	} else if name, ok := currentSection["name"].(string); ok && name != "" {
		sectionName = name
	}

	if sectionName == "" {
		return ""
	}

	// Get domain
	domain := ""
	if d, ok := params.CollectedData["domain"].(string); ok {
		domain = d
	} else if siteRecord, ok := params.CollectedData["site_record"].(map[string]interface{}); ok {
		if d, ok := siteRecord["domain"].(string); ok {
			domain = d
		}
	} else if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if siteRecord, ok := inputData["site_record"].(map[string]interface{}); ok {
			if d, ok := siteRecord["domain"].(string); ok {
				domain = d
			}
		}
	}

	// Build query
	query := sectionName
	if domain != "" {
		query = fmt.Sprintf("%s %s", sectionName, domain)
	}

	return query
}

// containsString checks if a string slice contains a specific string
// extractContentWithFallbacks tries multiple paths to find content data
// This handles different output formats from execute_llm_prompt (with/without .result wrapper)
func extractContentWithFallbacks(data map[string]interface{}, contentField string, logger *zap.Logger) map[string]interface{} {
	// Build list of paths to try
	pathsToTry := []string{contentField}

	// If path ends with ".result", also try without it
	if strings.HasSuffix(contentField, ".result") {
		base := strings.TrimSuffix(contentField, ".result")
		pathsToTry = append(pathsToTry, base)
		pathsToTry = append(pathsToTry, base+".response")
		pathsToTry = append(pathsToTry, base+".content")
	} else {
		// If path doesn't end with .result, also try with it
		pathsToTry = append(pathsToTry, contentField+".result")
		pathsToTry = append(pathsToTry, contentField+".response")
		pathsToTry = append(pathsToTry, contentField+".content")
	}

	// Try each path
	for _, path := range pathsToTry {
		if extracted := datahelpers.ExtractNestedField(data, path); extracted != nil {
			if m, ok := extracted.(map[string]interface{}); ok && len(m) > 0 {
				logger.Debug("extractContentWithFallbacks: Found content",
					zap.String("path", path),
					zap.Int("field_count", len(m)))
				return m
			}
		}
	}

	// Last resort: check if the base field exists and contains the content directly
	// Sometimes LLM output is stored as the field value itself
	parts := strings.Split(contentField, ".")
	if len(parts) > 0 {
		baseField := parts[0]
		if baseData := datahelpers.ExtractNestedField(data, baseField); baseData != nil {
			if m, ok := baseData.(map[string]interface{}); ok && len(m) > 0 {
				// Check if this map contains content-like fields
				if hasContentFields(m) {
					logger.Debug("extractContentWithFallbacks: Found content at base field",
						zap.String("base_field", baseField),
						zap.Int("field_count", len(m)))
					return m
				}
			}
		}
	}

	logger.Debug("extractContentWithFallbacks: No content found",
		zap.String("original_path", contentField),
		zap.Strings("tried_paths", pathsToTry))

	return nil
}

// hasContentFields checks if a map contains typical content field names
func hasContentFields(m map[string]interface{}) bool {
	contentFieldNames := []string{
		"headline", "subheadline", "body", "content", "heading",
		"title", "description", "text", "features", "items",
		"primary_cta", "cta_text", "button_text",
	}
	for _, fieldName := range contentFieldNames {
		if _, exists := m[fieldName]; exists {
			return true
		}
	}
	return false
}

// detectNeedsLLMContent determines if a component template needs LLM-generated content
// Returns true if template has content placeholders that need dynamic content
// Returns false if template only has structural placeholders (domain, logo_text, etc.)
func detectNeedsLLMContent(htmlTemplate, inputSchema string) bool {
	// No template = needs LLM to generate something
	if htmlTemplate == "" {
		return true
	}

	// If has input_schema, it defines what content is needed
	if inputSchema != "" && inputSchema != "{}" && inputSchema != "null" {
		return true
	}

	// Content placeholders that indicate LLM content is needed
	contentPlaceholders := []string{
		"{{.headline}}", "{{.subheadline}}", "{{.body}}", "{{.content}}",
		"{{.description}}", "{{.text}}", "{{.paragraph}}",
		"{{.primary_cta}}", "{{.secondary_cta}}", "{{.cta_text}}",
		"{{.features}}", "{{.services}}", "{{.benefits}}", "{{.items}}",
		"{{.testimonial}}", "{{.quote}}", "{{.author}}",
		"{{.heading}}", "{{.subtitle}}",
		"{{ .headline}}", "{{ .body}}", "{{ .content}}",
		"{{range .features}}", "{{range .services}}", "{{range .items}}",
		"{{range .benefits}}", "{{range .testimonials}}",
	}

	// Check for content placeholders
	for _, placeholder := range contentPlaceholders {
		if strings.Contains(htmlTemplate, placeholder) {
			return true
		}
	}

	// Structural fields that DON'T require LLM (filled from render_context)
	structuralFields := map[string]bool{
		"domain": true, "logo_text": true, "company_name": true,
		"tagline": true, "current_page": true, "year": true,
		"primary_color": true, "secondary_color": true, "accent_color": true,
		"background_color": true, "text_color": true,
		"email": true, "phone": true, "address": true,
		"nav_items_html": true, "theme_css": true, "title": true,
		"cta_url": true, "industry": true, "tone": true,
	}

	// Check for any {{.X}} where X is not a known structural field
	re := regexp.MustCompile(`\{\{\s*\.(\w+)\s*\}\}`)
	matches := re.FindAllStringSubmatch(htmlTemplate, -1)
	for _, match := range matches {
		if len(match) > 1 {
			fieldName := match[1]
			if !structuralFields[fieldName] {
				return true
			}
		}
	}

	// Only structural placeholders - doesn't need LLM
	return false
}

// UpdateWorkItemStatusAction updates a site_work_items row's status.
// Config:
//   - work_item_id_field: path to work_item_id in collected_data
//     (default: "input_data.work_item_id")
//   - status:             new status — default "complete"
//     (valid: complete, failed, claimed, executing,
//     detected, wont_fix, needs_human_review, unresolved)
//   - skip_if_missing:    bool — when true (default), gracefully no-op if
//     work_item_id absent. When false, error.
//   - error_message:      optional literal recorded in the error column so
//     triage can see why a handler parked the item.
//     When omitted and the status is not 'complete',
//     the routed step error (__step_error) is
//     recorded instead — see below.
//   - result_fields:      optional map of extra fields to merge into the
//     row's result JSONB. Values are literals; the
//     action always adds orchestration_id and step
//     metadata automatically.
//   - owned_page_refusal_status: optional, OFF when absent. The status to
//     record INSTEAD of `status` when the routed step
//     error came from the owned-page guard — i.e. the
//     handler was refused permission to touch this
//     page, rather than having tried and failed. Same
//     status vocabulary. See the block in the body for
//     what reads the difference and why it is a field.
func UpdateWorkItemStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateWorkItemStatusAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get work_item_id
	workItemIDField := "input_data.work_item_id"
	if f, ok := config["work_item_id_field"].(string); ok && f != "" {
		workItemIDField = f
	}
	workItemIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, workItemIDField)

	// Skip gracefully if missing — supports manual triggers without work_item_id
	skipIfMissing := true
	if v, ok := config["skip_if_missing"].(bool); ok {
		skipIfMissing = v
	}

	if workItemIDStr == "" {
		if skipIfMissing {
			params.Logger.Info("UpdateWorkItemStatusAction: no work_item_id in collected_data; skipping",
				zap.String("looked_at", workItemIDField))
			return map[string]interface{}{
				"updated": false,
				"skipped": true,
				"reason":  "work_item_id not present",
			}, nil
		}
		return nil, fmt.Errorf("work_item_id not found at %s and skip_if_missing=false", workItemIDField)
	}

	workItemID, err := uuid.Parse(workItemIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid work_item_id %q: %w", workItemIDStr, err)
	}

	// Status (default complete)
	newStatus := "complete"
	if s, ok := config["status"].(string); ok && s != "" {
		newStatus = s
	}
	validStatuses := map[string]bool{
		"complete":  true,
		"failed":    true,
		"claimed":   true,
		"executing": true,
		"detected":  true,
		"wont_fix":  true,
		// Handler no-op flags: a handler that cannot address its work item
		// (no ready sections, writer skipped) parks the item visibly here.
		// Without a flagged status the dispatch loop would stamp the item
		// complete on saga success — the false-completion bug caught on
		// robot-hands' gripper-detail (2026-07-10).
		"needs_human_review": true,
		"unresolved":         true,
	}
	if !validStatuses[newStatus] {
		return nil, fmt.Errorf("invalid work item status: %s (valid: complete, failed, claimed, executing, detected, wont_fix, needs_human_review, unresolved)", newStatus)
	}

	// --- Opt-in: an ownership REFUSAL is not a handler FAILURE ---
	//
	// OWNER DECISION 1, 2026-08-18 ("do not switch the handler off for this —
	// write something other than `failed`"). bugs_open/301 §3, bugs_open/083.
	//
	// WHAT THE PROBLEM IS, in plain terms. A page with rebuild_policy='owned'
	// belongs to a tool or widget, so the owned-page guard refuses a generic
	// section save over it. That refusal is correct. But the refused work item
	// then dies `failed`, and `failed` is what a fleet-wide gate counts.
	//
	// WHAT READS IT. The detected-item-promoter's floor (scheduled task
	// `detected-item-promoter`, migration 444 corrected by 454/465) holds an
	// (item_type, handler_agent) pair whose lifetime successes fall below 25% of
	// complete+verified+failed. Read from the live pre_query 2026-08-18: it
	// counts `complete`/`verified` and `failed`, and names no other status. So a
	// refusal recorded as `failed` enters the DENOMINATOR of a competence
	// measure it says nothing about — and when the pair crosses the floor the
	// promoter stops dispatching that item type ENTIRELY, including the findings
	// on generic pages that were succeeding. phantom_internal_link is the live
	// case: 69% on generic pages, 0 for 14 on owned ones, 47% overall.
	//
	// WHAT THIS DOES. When the routed step error carries the ownership guard's
	// marker, record the configured refusal status instead. With `wont_fix` the
	// pair reads NEVER TESTED HERE, which is the truth: excluded from numerator
	// and denominator alike, because the floor names neither.
	//
	// WHY A FIELD AND NOT A RULE IN CODE (owner ruling 2026-08-02 §2): new
	// authority on a shared seam ships opt-in with the unsafe default OFF, so
	// the decision is visible to a reviewer of the CALLER. Absent — which is
	// every caller today — this block is inert and the status is exactly what
	// the workflow configured. It is scoped to the ownership marker on purpose:
	// a genuine save failure (shrink guard, banned claim) still records `failed`
	// and still counts, or the floor would go blind to real incompetence.
	//
	// THE ONE BEHAVIOUR DIFFERENCE, measured rather than assumed. Of the
	// consumers that read `wont_fix` positively: silentCoverageClause
	// (diagnose_silent_check_action.go) treats failed and wont_fix alike;
	// crossLinkFailedStatuses lists both; check_page_canonical_collision's
	// suppression is scoped to its own item_type. Only retraction differs —
	// workItemClosedStatuses holds wont_fix and not failed, so the row will not
	// later be retracted. That is correct for a row that is already closed. The
	// dedup key is released either way (idx_swi_dedup excludes both), so the
	// finding re-raises when it is detected again.
	ownedPageRefusal := false
	configuredStatus := newStatus
	if refusalStatus, ok := config["owned_page_refusal_status"].(string); ok && refusalStatus != "" {
		if !validStatuses[refusalStatus] {
			return nil, fmt.Errorf("invalid owned_page_refusal_status: %s (same vocabulary as status)", refusalStatus)
		}
		// The marker travels in the routed error only — routeToErrorStep copies
		// the action's error message verbatim to __step_error.message. A step
		// reached by next_step rather than error_step has no __step_error and so
		// is never downgraded, which is what keeps this to the refusal path.
		routed := datahelpers.ExtractNestedFieldString(params.CollectedData, "__step_error.message")
		if strings.Contains(routed, ownedPageSkipReasonPrefix) {
			params.Logger.Warn("UpdateWorkItemStatusAction: ownership refusal — recording the refusal status, not the configured one",
				zap.String("work_item_id", workItemIDStr),
				zap.String("configured_status", newStatus),
				zap.String("refusal_status", refusalStatus),
				zap.String("marker", ownedPageSkipReasonPrefix))
			newStatus = refusalStatus
			ownedPageRefusal = true
		}
	}

	// Optional error message — recorded in the error column so triage can see
	// why a handler parked the item.
	errorMessage, _ := config["error_message"].(string)

	// Fall back to the REAL step error when the workflow supplied no literal.
	//
	// Why (bugs_closed/040-partial-build, candidate 2): a literal only fits a
	// static reason ("writer skipped this page"). The genuinely failing path —
	// `mark_item_failed`, reached via error_step — has a *dynamic* reason and so
	// carries no literal, which left `site_work_items.error` EMPTY on every such
	// item while the coordinator had already written the real message to
	// agent_error_log 1s earlier from the very same routeToErrorStep call
	// (coordinator.go: __step_error and logAgentError are set together). 21 of 75
	// failed items fleet-wide carried a blank error on 2026-07-25; 20 of those 21
	// had exactly one agent_error_log row waiting to be joined. Triage looks at
	// the item, not the log, so the reason was effectively invisible.
	//
	// Never for 'complete': __step_error is never cleared once set, so a workflow
	// that recovers from a routed error and then stamps the item complete would
	// otherwise be given a stale failure. Fleet census 2026-07-25 — the only
	// literal-less update_work_item_status steps are 2×'failed'
	// (page-build-handler, image-build-handler) and 1×'complete'
	// (image-build-handler); the 2×'needs_human_review' steps carry literals and
	// are unaffected because a configured literal always wins.
	if errorMessage == "" && newStatus != "complete" {
		if stepErr := datahelpers.ExtractNestedFieldString(params.CollectedData, "__step_error.message"); stepErr != "" {
			errorMessage = stepErr
			// Name the step unless the message already does. The routed message
			// has two shapes: "step X failed: …" (action errors) and a bare
			// "Request <id> timed out after 3 retries" (awaited-request
			// timeouts), which alone does not say WHAT timed out. Converge on
			// the prefix the column already uses rather than inventing a
			// second format.
			if failedStep := datahelpers.ExtractNestedFieldString(params.CollectedData, "__step_error.failed_step"); failedStep != "" &&
				!strings.HasPrefix(errorMessage, "step ") {
				errorMessage = "step " + failedStep + " failed: " + errorMessage
			}
			params.Logger.Info("UpdateWorkItemStatusAction: no error_message literal — recording the routed step error",
				zap.String("work_item_id", workItemIDStr),
				zap.String("status", newStatus),
				zap.String("error", errorMessage))
		}
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateWorkItemStatusAction: No database")
		return map[string]interface{}{"updated": false, "status": newStatus}, nil
	}

	// Build result payload — always includes orchestration tracking; merges
	// any caller-supplied extras under result_fields.
	resultPayload := map[string]interface{}{
		"completed_by_orchestration_id": params.ExecutionContext.OrchestrationID,
		"completed_by_step":             params.ExecutionContext.StepName,
		"completed_at_iso":              time.Now().UTC().Format(time.RFC3339),
	}
	// Stamp the substitution on the row, not only in the log. Without this the
	// only trace that a status was chosen rather than configured is a scrolling
	// pod line, and the rows are indistinguishable from a human's wont_fix —
	// which matters because the whole point is that a census can separate
	// "refused on ownership grounds" from "the handler could not do this job".
	// Queryable as: result ? 'owned_page_refusal'.
	if ownedPageRefusal {
		resultPayload["owned_page_refusal"] = true
		resultPayload["owned_page_refusal_replaced_status"] = configuredStatus
		resultPayload["owned_page_refusal_marker"] = ownedPageSkipReasonPrefix
	}
	if extras, ok := config["result_fields"].(map[string]interface{}); ok {
		for k, v := range extras {
			resultPayload[k] = v
		}
	}
	// ── COMPLETION GATE 2, opt-in (bugs_open/375).
	//
	// This action is the platform's SECOND writer of `complete`, and until now the
	// only one that never asked the item type's verifier whether the defect was
	// actually gone. See update_work_item_status_verification.go for what a
	// verifier is, the measured blast radius, and why the consult is opt-in per
	// step rather than automatic (short version: making it automatic would
	// fail-close CQ-023's `converted` arm the day somebody registers the verifier
	// its own backlog invites them to write).
	//
	// Absent `verify_before_complete` — which is every live step as of 2026-08-24
	// — this call reaches `GetVerifier`, finds nothing registered for any of the
	// five live types, and returns (nil, true). The arm below is then byte-identical
	// to what it was. When a verifier IS registered and the step is unarmed, the
	// completion still proceeds and the bypass is recorded at result._verification
	// rather than passing in silence.
	verifyArmed, _ := config[updateStatusVerifyConfigKey].(bool)
	if newStatus == "complete" {
		verification, mayComplete := verifyBeforeUpdateStatusComplete(ctx, params.DB, workItemID, verifyArmed, params.Logger)
		if verification != nil {
			resultPayload["_verification"] = verification
		}
		if !mayComplete {
			blockedJSON, mErr := json.Marshal(resultPayload)
			if mErr != nil {
				return nil, fmt.Errorf("failed to marshal result payload: %w", mErr)
			}
			agentType := "unknown"
			if params.ExecutionContext.Sender.AgentType != "" {
				agentType = params.ExecutionContext.Sender.AgentType
			}
			// The SAME refusal path the guarded writer takes: attempt_count+1, claim
			// released, 'triaged' for retry or 'failed' once the budget is spent. A
			// second refusal path here would be a second definition of what a blocked
			// completion means, which is the drift bugs_closed/284 exists to stop.
			msg, reason := blockedCompletionReason(verification)
			return failUnverifiedCompletion(ctx, params.DB, workItemID, agentType, string(blockedJSON),
				msg, reason, params.Logger)
		}
	}

	resultJSON, err := json.Marshal(resultPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result payload: %w", err)
	}

	// ── A FAILURE goes through the shared failure-write contract, not through
	// this statement (bugs_open/307).
	//
	// WHAT WAS WRONG HERE, and it is the larger half of that bug. This UPDATE
	// incremented attempt_count and wrote the configured status under
	// `WHERE id = $1` — no CASE ladder and no terminal-status guard. So a step
	// configured `status: 'failed'` was TERMINAL ON ITS FIRST FAILURE however
	// many max_attempts the row declared, in fair weather as much as in an
	// outage. Five live agents carry such a step (page-build-handler,
	// image-build-handler, image-source-unsatisfiable-handler,
	// image-url-404-handler, required-fields-missing-handler) and the bleed is
	// measurable: 141 of 270 failed rows in the 14 days to 2026-08-19 died
	// before exhausting their budget, 139 of them at attempt_count=1 of 3.
	//
	// Scoped to `failed` on purpose. An ownership REFUSAL (WII-019's
	// owned_page_refusal_status, which lands `wont_fix` here) is a DECISION, not
	// a failure, and must not be re-triaged; nor must the six
	// needs_human_review steps or the six `complete` ones. The condition below
	// therefore excludes the refusal substitution explicitly rather than trusting
	// that `failed` implies "not a decision".
	if newStatus == "failed" && !ownedPageRefusal {
		outcome, lerr := applyWorkItemFailureLadder(ctx, params.DB, params.Logger,
			workItemID, errorMessage, "", resultJSON,
			// bugs_open/345 candidate 2's opt-in, read from THIS step's config.
			// No live update_work_item_status step opts in today, so this arm is
			// byte-identical until one does.
			params.StepConfig.Config)
		if lerr != nil {
			return nil, lerr
		}
		if outcome.Skipped {
			params.Logger.Info("UpdateWorkItemStatusAction: skipped — a deliberate status is already recorded, not overwriting",
				zap.String("work_item_id", workItemIDStr),
				zap.String("reason", outcome.SkipReason))
			return map[string]interface{}{
				"updated": false,
				"skipped": true,
				"reason":  outcome.SkipReason,
			}, nil
		}
		params.Logger.Info("UpdateWorkItemStatusAction: failure recorded through the work-item failure ladder",
			zap.String("work_item_id", workItemIDStr),
			zap.String("status", outcome.NewStatus),
			zap.Int("attempts_left", outcome.AttemptsLeft),
			zap.Bool("released_without_counting", outcome.Released),
			zap.Int("backoff_minutes", outcome.BackoffMins))
		return map[string]interface{}{
			"updated":         true,
			"status":          outcome.NewStatus,
			"work_item_id":    workItemIDStr,
			"attempts_left":   outcome.AttemptsLeft,
			"released":        outcome.Released,
			"release_reason":  outcome.ReleaseReason,
			"backoff_minutes": outcome.BackoffMins,
		}, nil
	}

	// Build query. completed_at only set when transitioning to complete; for
	// other statuses leave it alone and just update status.
	//
	// The `complete` arm carries the terminal-decision guard CompleteWorkItemAction
	// has (WII-003, load_work_item_actions.go) — this action is a third writer of
	// `complete` and had no guard at all, so a handler that deliberately flagged
	// its item could be re-stamped complete from here. Same defect, same remedy,
	// one writer over.
	//
	// ⚠ IT USES workItemCompletionGuardStatuses, **NOT** the failure path's
	// workItemDecisionStatuses, and the two are not interchangeable. The failure
	// list deliberately omits `failed` and `unresolved` so the retry ladder can
	// move a row through them; on THIS path both must be protected, or a
	// `complete` write silently overwrites a row that already failed or was given
	// up — which is the very defect class this change exists to close. The first
	// version of this edit reused the wrong list; the council's editquality seat
	// gated on it (corr 4cdec68b, round 1) and was right.
	var query string
	if newStatus == "complete" {
		query = `UPDATE site_work_items
		            SET status = $2,
		                completed_at = NOW(),
		                updated_at = NOW(),
		                attempt_count = attempt_count + 1,
		                result = COALESCE(result, '{}'::jsonb) || $3::jsonb,
		                error = COALESCE(NULLIF($4, ''), error)
		          WHERE id = $1
		            AND status NOT IN (` + sqlInList(workItemCompletionGuardStatuses) + `)`
	} else {
		query = `UPDATE site_work_items
		            SET status = $2,
		                updated_at = NOW(),
		                attempt_count = attempt_count + 1,
		                result = COALESCE(result, '{}'::jsonb) || $3::jsonb,
		                error = COALESCE(NULLIF($4, ''), error)
		          WHERE id = $1`
	}

	if err := execDB(ctx, params.DB, query, workItemID, newStatus, resultJSON, errorMessage); err != nil {
		return nil, fmt.Errorf("failed to update work item status: %w", err)
	}

	params.Logger.Info("UpdateWorkItemStatusAction: status updated",
		zap.String("work_item_id", workItemIDStr),
		zap.String("status", newStatus),
	)

	return map[string]interface{}{
		"updated":      true,
		"work_item_id": workItemIDStr,
		"status":       newStatus,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ============================================================================
// Adoption-faithfulness convergence (doc 029 Phase 1 / FOCUS_adoption_faithfulness_via_locks.md)
//
// These helpers make ValidateSitePlanAction deterministically preserve the
// pages that are currently under a live adoption lock, so the planner LLM
// cannot drop, rename, or duplicate adopted pages during the faithful-first-
// pass window. They become no-ops once the adoption lock has expired (or for
// from-scratch builds), letting the site develop normally thereafter.
// ============================================================================

// isSectionIndexType reports whether a page_type is a directory/section index.
//
// news-index belongs here for the same reason blog-index does: it is the typed
// PARENT of a section, not one of its children. Its absence meant a news-index
// page at the canonical /news/index.html URL matched classifyPagesForNav's
// child-prefix skip and could never enter the nav — while the same page at
// /news.html (the non-canonical shape bugs_open/080 exists to eliminate)
// navigated fine. bugs_open/141.
func isSectionIndexType(pageType string) bool {
	switch pageType {
	case "blog-index", "entity-directory", "section-index", "news-index":
		return true
	}
	return false
}

// sectionStemOf returns the section stem for a realised section-index page, or
// "" if it isn't a section index. e.g. ("games-index", "/games/index.html",
// "entity-directory") -> "games". Prefers the URL path segment; falls back to
// the name minus the "-index" suffix.
func sectionStemOf(name, url, pageType string) string {
	isIndex := isSectionIndexType(pageType)
	if !isIndex {
		// Also treat any non-root URL ending in /index.html as a section index.
		if strings.HasSuffix(url, "/index.html") && url != "/index.html" {
			isIndex = true
		}
	}
	if !isIndex {
		return ""
	}
	trimmed := strings.Trim(url, "/")
	if i := strings.Index(trimmed, "/"); i > 0 {
		return trimmed[:i]
	}
	return strings.TrimSuffix(name, "-index")
}

// slugOf derives a comparison slug from an LLM page's name/url. A flat page URL
// like /games.html yields "games"; falls back to the name.
func slugOf(name, url string) string {
	if url != "" {
		t := strings.Trim(url, "/")
		t = strings.TrimSuffix(t, ".html")
		t = strings.TrimSuffix(t, "/index")
		if t != "" {
			if i := strings.Index(t, "/"); i > 0 {
				return t[:i]
			}
			return t
		}
	}
	return name
}

// sectionPathKey is the site path a page CLAIMS, used by Pass C to tell a page
// that collides with a section index from one that lives UNDER it.
//
// datahelpers.PagePathKey is the estate's collision key and does the real work:
// "/news.html" and "/news/index.html" both claim "/news", while "/news/x.html"
// claims "/news/x". This wrapper only normalises the two URL shapes that key
// deliberately passes through unchanged but Pass C's first-segment predecessor
// accepted anyway — a URL with no leading slash, and a directory URL's trailing
// one. Returns "" for a page with no URL, which is the one case Pass C must
// still settle by name.
func sectionPathKey(url string) string {
	u := strings.TrimSpace(url)
	if u == "" {
		return ""
	}
	if !strings.HasPrefix(u, "/") {
		u = "/" + u
	}
	if u != "/" {
		u = strings.TrimSuffix(u, "/")
	}
	return datahelpers.PagePathKey(u)
}

// itemStemOf returns the topic stem of an item page name by stripping the
// role prefixes that CanonicalisePage adds (tool-, guide-, game-): e.g.
// "guide-economy-basics" -> "economy-basics", "economy-basics" ->
// "economy-basics". Returns the input unchanged when no role prefix is
// present, so two adopted pages on the same topic share a stem and a
// re-proposed bare sibling collides with them. Unlike sectionStemOf, this is
// name-based rather than URL/hub-based.
//
// Delegates to datahelpers (bugs_open/215): the prefix list used to be
// maintained here by hand against CanonicalisePage in another package, under a
// comment asking the reader to keep them in sync. It now has one definition,
// beside the canonicaliser whose behaviour it mirrors.
func itemStemOf(name string) string { return datahelpers.PageItemStem(name) }

// normaliseRealisedToPlanPage converts a realised pages-table row (as returned
// by the load_existing_pages query) into the plan-page shape the downstream
// write_site_plan / ValidateRoles expects. Carries a from_realised marker so
// logging/debugging can distinguish these from LLM-proposed pages.
func normaliseRealisedToPlanPage(rm map[string]interface{}) map[string]interface{} {
	name, _ := rm["name"].(string)
	pageType, _ := rm["page_type"].(string)
	// Carry the realised page's sections so a unioned adopted page keeps them.
	// load_existing_pages runs via query_database, which stringifies jsonb
	// columns, so rm["sections"] arrives as a JSON string (e.g. ["hero",
	// "guide-list"]); tolerate a native []interface{} too. Empty/missing -> [].
	// Without carrying these, the union emits empty values and the page sync's
	// "<col> = EXCLUDED.<col>" clobbers the adopted page's real sections,
	// meta_description, and nav_order. (nav_label is COALESCE-preserved by the
	// upsert, so it is safe without carrying.)
	sections := []interface{}{}
	switch v := rm["sections"].(type) {
	case []interface{}:
		sections = v
	case string:
		if v != "" {
			var parsed []interface{}
			if err := json.Unmarshal([]byte(v), &parsed); err == nil {
				sections = parsed
			}
		}
	}
	url, _ := rm["url"].(string)
	return map[string]interface{}{
		"name":             name,
		"page_type":        pageType,
		"url":              rm["url"],
		"title":            rm["title"],
		"nav_label":        rm["nav_label"],
		"in_header":        rm["in_header"],
		"in_footer":        rm["in_footer"],
		"meta_description": rm["meta_description"],
		"nav_order":        rm["nav_order"],
		"sections":         sections,
		"from_realised":    true,
		// The stored identity is authoritative for a page derived from a realised
		// row, and saying so is what stops the write path re-deriving it
		// (bugs_open/215). Both canonicalisation surfaces re-run CanonicalisePage
		// over every plan page, and CanonicalisePage cannot EXPRESS a legacy
		// identity: a tool-typed page always comes back "tool-<bare>" at the
		// nested URL. Measured 2026-08-11: 71 live shipped rows fleet-wide are not
		// fixed points of the canonicaliser, so for those pages a snap or a union
		// here is silently undone downstream and the twin identity is re-minted.
		//
		// Inert unless a step config sets honour_realised_identity (see
		// realisedIdentityOf) — the field is opt-in with the unsafe default off.
		// parent_section is carried for the same reason and helps even when the
		// flag is off: without it CanonicalisePage re-derives a blog-post's URL
		// under /blog/, which MOVES a live page that is serving from /guides/
		// (the bugs_open/241 URL-move hazard, measured on fundamentallyai's
		// guide twins).
		"identity_authority": "realised",
		"parent_section":     parentSectionFromURL(url),
	}
}

// parentSectionFromURL recovers the directory a realised page actually lives in,
// so a page derived from it canonicalises back to where it is SERVING rather
// than to the role's default hub. "/guides/x.html" and "/guides/x/index.html"
// both give "guides"; a root-level "/x.html" gives "" (no parent), which is
// also what a URL we cannot read gives — in both cases the canonicaliser's own
// default applies, which is the behaviour that existed before this carry.
//
// Delegates to datahelpers (bugs_open/463): ValidateRoles now needs the same
// derivation for a PROPOSED page, and a second copy in another package is the
// drift class this estate keeps paying for. One definition, beside the
// canonicaliser whose default it exists to override.
func parentSectionFromURL(url string) string { return datahelpers.ParentSectionFromURL(url) }

// factCarryMiss records plan-time fact assignments that could NOT be carried
// onto a restored realised composition: the planner assigned facts to section
// names the built page does not contain, so those assignments are gone.
//
// It exists because the failure is invisible otherwise — a name match that
// matches nothing produces exactly the same plan as a carry that worked. The
// caller turns each of these into a durable row (see recordFactCarryMisses).
type factCarryMiss struct {
	// Page is the identity the page KEEPS in the plan (for Pass B that is the
	// realised name, not the one the planner proposed).
	Page string
	// Sections holds one entry per dropped assignment, in the planner's own
	// emission order; a name appears twice if it was assigned twice and the
	// realised composition has only one such section.
	Sections []string
}

// carrySectionFactsOntoRealised carries the planner's per-section fact
// assignments (RFC_016's object-form entries) off the LLM's proposed section
// list and onto the realised list that Pass B/B2 restores over it.
//
// WHY (RFC_016 §3b, bugs_open/151 candidate 1b): an assignment travels INSIDE
// its section entry, so restoring the realised composition wholesale discarded
// every assignment on a deployed page — candidate 1 could reach only pages
// built AFTER their plan first carried assignments, i.e. never the pages that
// motivated it. Re-attachment is by component NAME because that is the only
// identity the two lists share: the realised list is plain strings, ordering
// and length differ, and position is not an address that survives (§1a).
//
// The result is in the planner's OWN object form, so nothing new has to
// understand it. ValidateSitePlanAction's normalise pass splits object entries
// into plain strings plus an aligned per-page section_facts array, and it runs
// LATER IN THE SAME FUNCTION BODY than the reconcile call that gets here —
// order verified by reading the caller, not assumed. So no reader of
// pages["sections"] downstream of validate_plan ever sees object form, and the
// 15+ consumers of that key are not in this change's blast radius.
//
// Since 2026-08-26 the same carry moves an entry's "subject" (RFC_016 §5.1's
// second structured field): a subject-only object entry — facts key absent —
// is still counted in the fourth return (the 333 disobedience record is about
// facts) but its subject carries, and a "facts" key is never fabricated for
// it. The unmatched list stays facts-worded; a subject on an unmatched entry
// is dropped with it, which is the same fate facts have always had.
//
// Returns the merged list, how many entries received an assignment, the
// assignments that matched nothing, and the object entries that resolved a name
// but carried NO usable `facts` value (key absent, null, or not an array).
// Seed 333 makes the key mandatory on every section the planner emits, so the
// fourth return is planner disobedience — kept separate from unmatched because
// without it the omission is invisible: the entry skips both `pending` and the
// unmatched sweep, indistinguishable from a page correctly assigned no facts
// (council round a06ff850, objection §3.5). When nothing carries, the realised
// slice is returned UNCHANGED — so a plan with no object entries (every plan
// before seed 333, and every page the planner emits as bare strings) is
// byte-identical to what this function produced before candidate 1b, and
// produces no absent entries either.
func carrySectionFactsOntoRealised(realised []interface{}, proposed interface{}) ([]interface{}, int, []string, []string) {
	pl, _ := proposed.([]interface{})
	if len(pl) == 0 || len(realised) == 0 {
		return realised, 0, nil, nil
	}

	// Assignments in emission order, plus a name index into them. A page may
	// legitimately hold two entries of the same component (two text blocks), so
	// a name maps to a QUEUE: the Nth realised entry of a name takes the Nth
	// assignment made for it. Entries the planner left as bare strings carry no
	// assignment and never enter this list — they must not blank an assignment
	// an object entry of the same name made.
	type pendingAssignment struct {
		name string
		// facts is meaningful only when factsPresent; a subject-only entry
		// (facts key absent — recorded in the fourth return either way) still
		// enqueues so its subject carries.
		facts        []interface{}
		factsPresent bool
		subject      string
		used         bool
	}
	var pending []pendingAssignment
	var absent []string
	byName := make(map[string][]int)
	for _, entry := range pl {
		obj, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := sectionEntryName(obj)
		if !ok {
			continue
		}
		// [] == deliberately factless and carries; an object entry with the key
		// absent, null, or non-array carries nothing — but under seed 333 the
		// key is mandatory, so that shape is disobedience, not a pre-scoping
		// plan (those emit bare strings, which never reach this cast). Record
		// it rather than skipping silently: skipped, it is indistinguishable
		// from a page correctly assigned no facts.
		subject, _ := obj["subject"].(string)
		subject = strings.TrimSpace(subject)
		facts, factsPresent := obj["facts"].([]interface{})
		if !factsPresent {
			absent = append(absent, name)
			if subject == "" {
				continue
			}
		}
		byName[name] = append(byName[name], len(pending))
		pending = append(pending, pendingAssignment{name: name, facts: facts, factsPresent: factsPresent, subject: subject})
	}
	if len(pending) == 0 {
		return realised, 0, nil, absent
	}

	merged := make([]interface{}, len(realised))
	carried := 0
	for i, entry := range realised {
		merged[i] = entry
		name, ok := sectionEntryName(entry)
		if !ok {
			continue
		}
		idx := -1
		for _, candidate := range byName[name] {
			if !pending[candidate].used {
				idx = candidate
				break
			}
		}
		if idx < 0 {
			continue
		}
		pending[idx].used = true
		// Clone rather than replace: a realised entry is a plain string today,
		// but if one ever arrives as an object its other keys are not this
		// function's to discard.
		if src, isObj := entry.(map[string]interface{}); isObj {
			clone := make(map[string]interface{}, len(src)+2)
			for k, v := range src {
				clone[k] = v
			}
			if pending[idx].factsPresent {
				clone["facts"] = pending[idx].facts
			}
			if pending[idx].subject != "" {
				clone["subject"] = pending[idx].subject
			}
			merged[i] = clone
		} else {
			m := map[string]interface{}{"name": name}
			if pending[idx].factsPresent {
				m["facts"] = pending[idx].facts
			}
			if pending[idx].subject != "" {
				m["subject"] = pending[idx].subject
			}
			merged[i] = m
		}
		carried++
	}

	var unmatched []string
	for _, p := range pending {
		if !p.used {
			unmatched = append(unmatched, p.name)
		}
	}
	if carried == 0 {
		return realised, 0, unmatched, absent
	}
	return merged, carried, unmatched, absent
}

// snapPlanPageOntoRealised replaces an LLM plan entry with the realised page it
// denotes, and is the ONE arm every identity match goes through — the exact-URL
// rename (Pass B) and all three twin-identity layers. Extracted rather than
// copied because its section rules are load bearing in two directions at once
// (bugs_open/050's empty-page routing, and the plan-time fact assignments the
// 151 lane measures), and a second copy of them would drift.
//
// The realised identity wins wholesale: name, url, page_type, nav and meta come
// from the realised row. The LLM entry contributes only its sections, and only
// where the realised page has no composition to protect:
//
//   - realised NOT shipped + sections empty -> take the LLM's proposal. A
//     catalogued page that has never been composed is finally composed, and the
//     LLM's entries are kept WHOLE so their fact assignments ride along.
//   - otherwise -> the realised composition wins, and the fact assignments on
//     the LLM's discarded entries are carried onto the surviving names.
//
// The entry is snapped, never dropped. Dropping would be simpler and is what
// Pass C2 does, but it discards the plan-time fact assignments with the entry,
// and Pass A's union cannot carry them back — it has no proposal to read.
func snapPlanPageOntoRealised(
	lm map[string]interface{},
	rp map[string]interface{},
	layer string,
	counts *reconcileCounts,
	logger *zap.Logger,
) map[string]interface{} {
	lname, _ := lm["name"].(string)
	lurl, _ := lm["url"].(string)
	rname, _ := rp["name"].(string)
	rurl, _ := rp["url"].(string)

	snapped := normaliseRealisedToPlanPage(rp)
	// The spelling the planner used, kept so the imagery scope map can still
	// resolve a block keyed to it (bugs_open/214's name-space mismatch: imagery
	// keys off the RAW page name, and a snap changes the name out from under it).
	snapped["reconciled_from"] = lname

	carried := 0
	if !realisedPageHasShipped(rp) && len(realisedSectionsOf(rp)) == 0 {
		if ls, ok := lm["sections"].([]interface{}); ok && len(ls) > 0 {
			snapped["sections"] = ls
		}
	} else if rs, ok := snapped["sections"].([]interface{}); ok && len(rs) > 0 {
		restored, n, unmatched, absent := carrySectionFactsOntoRealised(rs, lm["sections"])
		snapped["sections"] = restored
		carried = n
		counts.SectionFactsCarried += n
		if len(unmatched) > 0 {
			counts.FactCarryMisses = append(counts.FactCarryMisses,
				factCarryMiss{Page: rname, Sections: unmatched})
		}
		if len(absent) > 0 {
			counts.FactAssignmentAbsent = append(counts.FactAssignmentAbsent,
				factCarryMiss{Page: rname, Sections: absent})
		}
	}

	if layer == "url_exact" {
		// Unchanged wording: this line predates the twin layers and is what the
		// existing pod-greps and tests look for.
		logger.Info("validate: snapped renamed page back to realised identity",
			zap.String("llm_name", lname), zap.String("realised_name", rname),
			zap.Int("section_facts_carried", carried))
	} else {
		logger.Info("validate: snapped plan page onto realised identity",
			zap.String("layer", layer),
			zap.String("llm_name", lname), zap.String("llm_url", lurl),
			zap.String("realised_name", rname), zap.String("realised_url", rurl),
			zap.Int("section_facts_carried", carried))
		counts.IdentitySnaps = append(counts.IdentitySnaps, identitySnap{
			Layer: layer, PlanName: lname, PlanURL: lurl,
			RealisedName: rname, RealisedURL: rurl,
		})
	}
	return snapped
}

// stampSameNameRealisedIdentity records that a plan entry which names a realised
// page by its EXACT stored name is derived from that page, so the write path
// stops re-deriving its identity (bugs_open/215, the same-name hole).
//
// WHY THIS EXISTS AT ALL, and why it is the case nobody built for. The three
// twin layers above all answer "is this plan entry a page we have realised under
// a DIFFERENT spelling?", and their shared eligible closure refuses a same-name
// candidate — correctly, because such a candidate is not a twin, it is the page
// itself. It does so through EITHER of two clauses, redundantly (see the closure:
// rname == lname, and planNames[rname]), which is why relaxing one of them cannot
// reach this case. Pass B2 below then pairs exactly that case and
// reconciles only the page's COMPOSITION. So until this function existed, the
// well-behaved case — the planner reproducing a page's stored name verbatim, which
// is what we ask it to do — was reachable by no marker-minting path at all, and
// honour_realised_identity could not fire for it however the site was configured.
//
// MEASURED, and this is the incident that found it: loanandmortgagecalculator.co.uk
// seeded honour_realised_identity='true', fired one canary replan on 2026-08-17
// (corr 6fe6ee93-67b9-4831-bf17-2ca473e1d30c, chassis v1.0.1305), and had 19
// phantom page rows INSERTed anyway — 17 of them the predicted tool-<name> twins
// of pages the planner had named CORRECTLY. reconcile_result recorded
// pages_restamped: 0. The flag was not broken; it was unreachable.
//
// WHAT IS STAMPED, AND WHY EACH FIELD IS SAFE WHEN THE FLAG IS OFF. This function
// runs unconditionally — like Pass B's snap and Pass A's union, which also stamp
// without consulting a gate — because the reconciler decides identity and the two
// writers decide whether to honour it. That division is only safe if every field
// written here is inert to a writer that is NOT honouring:
//   - identity_authority: read by nothing except realisedIdentityOf, whose two
//     call sites are both inside `if identityPolicy.HonourRealisedIdentity`.
//   - url: both write surfaces derive and overwrite the URL themselves when they
//     are not honouring, so a stamped URL cannot reach an artefact through them.
//     It is stamped because realisedIdentityOf refuses an incomplete triple.
//   - page_type: NOT inert — it feeds the writers' Role, hence CanonicalisePage.
//     Which is why the type-equality precondition below is a precondition and not
//     a nicety: we only ever write back the value the writer would have derived
//     anyway, so the flag-off write is a no-op by construction.
//
// WHAT IS DELIBERATELY NOT STAMPED. parent_section and slug both feed the writers'
// canonicalisation directly, so writing them would change flag-OFF behaviour and
// this fix's entire safety argument with it. (normaliseRealisedToPlanPage does
// stamp parent_section, for a union or a snap — there the whole entry is already
// realised-derived and the bugs_open/241 URL-move hazard makes it necessary. Here
// the entry keeps the planner's own copy, so it is not that kind of entry.)
// from_realised is likewise not set: a B2-paired entry is NOT wholly derived from
// the realised row — it keeps the LLM's title, meta and nav, which is Pass B2's
// standing contract and the reason this stamp does not simply route through
// snapPlanPageOntoRealised.
//
// LIMITATION, stated because it is invisible from here: realisedByName is built
// over the PRESERVED set only, so a realised row that is neither deployed nor
// needs_rebuild (on a site that already has a current plan) never reaches this
// function and keeps today's re-derivation behaviour.
func stampSameNameRealisedIdentity(
	lm map[string]interface{},
	rp map[string]interface{},
	counts *reconcileCounts,
	logger *zap.Logger,
) {
	lname, _ := lm["name"].(string)
	rname, _ := rp["name"].(string)
	rurl, _ := rp["url"].(string)
	rtype, _ := rp["page_type"].(string)

	// Mirror realisedIdentityOf's own refusal. Claiming a stamp the reader would
	// reject would put a lie in the counters — the shape this lane has already
	// been caught by once, where a guard that never fires is indistinguishable
	// from one that is not there.
	if rname == "" || rurl == "" || rtype == "" {
		return
	}

	// The type the WRITERS will use, derived exactly as they derive it
	// (firstNonEmptyField over page_type, type, role — write_site_plan_action.go
	// and site_db_actions.go both). Comparing lm["page_type"] alone would miss an
	// entry whose type arrived under one of the other two keys and would then
	// overwrite a Role the writer was about to use.
	planType := firstNonEmptyField(lm, "page_type", "type", "role")
	if !strings.EqualFold(strings.TrimSpace(planType), strings.TrimSpace(rtype)) {
		// Two pages with one name and two roles is not a reconciliation this
		// function may perform: honouring would silently retype a live page, and
		// re-deriving keeps today's behaviour. Refuse, and record it — a chassis
		// pod retains under a second of log (bugs_open/136 §11), so a Warn here
		// would be no record at all.
		counts.SameNameTypeConflicts = append(counts.SameNameTypeConflicts, sameNameTypeConflict{
			PlanName:     lname,
			PlanType:     planType,
			RealisedType: rtype,
			RealisedURL:  rurl,
		})
		logger.Warn("validate: same-name plan page disagrees with the realised page's type; identity not stamped",
			zap.String("page", lname),
			zap.String("plan_type", planType),
			zap.String("realised_type", rtype))
		return
	}

	lm["identity_authority"] = "realised"
	lm["url"] = rurl
	lm["page_type"] = rtype

	// Which of these stamps actually CHANGES an outcome: the ones whose stored
	// name the canonicaliser would not have returned. Name-only, derived the way
	// the write path derives it (slug then name) — this is a classification for
	// the durable record, never a gate on the stamp itself, because predicting
	// the writers' full answer would need parent_section and the site's URL shape,
	// which this function does not have and must not read.
	wouldDerive := datahelpers.PageCanonicalNameForRow(
		firstNonEmpty(datahelpers.GetStringField(lm, "slug", ""), lname), rtype)
	if wouldDerive == rname {
		wouldDerive = ""
	}
	counts.SameNameStamps = append(counts.SameNameStamps, sameNameStamp{
		PlanName:        lname,
		StoredURL:       rurl,
		WouldDeriveName: wouldDerive,
	})
}

// matchTwinIdentity asks whether an LLM plan entry denotes a page that is
// already realised under a different spelling, and returns the snapped entry if
// so (bugs_open/215 quiet mode).
//
// Three keys, strongest first. All three refuse a key two realised pages claim,
// and none may match a realised page whose name the plan ALREADY carries:
//
//   - path_key: the two URLs claim one path. This is the flat-vs-nested family
//     — the legacy tool-deploy arm's /tools/x.html against the canonicaliser's
//     /tools/x/index.html — and it is deterministic: both sides are stored URLs,
//     no heuristic about names.
//   - canonical_name: the canonicaliser says the plan entry's own identity IS
//     the realised row's canonical name. This front-runs a collapse the estate
//     already performs one step later (write_site_plan canonicalises, then
//     dedupePlanPageRows merges) — doing it here just means the survivor is the
//     realised page rather than whichever spelling sorted first.
//   - stem_twin: the two names share a topic stem, one bare and one prefixed.
//     The weakest key, separately gated, and dark-launched — see below.
//
// The stem layer is OFF unless opts.StemTwinSnap, and while off it still counts
// what it WOULD have done. That is deliberate: the estate has declined to trust
// stem matching on re-plans since 2026-07-20 on a stated risk (a new
// "tool-pricing" beside a built "guide-pricing"), and the honest way to revisit
// a refusal like that is to measure the population first, not to argue about it.
// Note the guard makes that exact pair unmatchable in any case: both names carry
// a prefix, and this layer requires one side to be the bare stem.
func matchTwinIdentity(
	lm map[string]interface{},
	lname, lurl, ltype string,
	byPathKey, byCanonName, byStem, byName map[string]map[string]interface{},
	planNames map[string]bool,
	opts reconcileOptions,
	counts *reconcileCounts,
	logger *zap.Logger,
) (map[string]interface{}, string, bool) {
	if lname == "" {
		return nil, "", false
	}

	// eligible rejects a candidate that must not be snapped onto, whatever key
	// found it. Shared by all three layers so a guard cannot be true of one and
	// forgotten in another.
	// TWO OF THESE CLAUSES ARE REDUNDANT IN SERIES, and knowing which matters —
	// [MEASURED by mutation 2026-08-19] removing EITHER one alone changes no
	// behaviour and fails no test; removing BOTH lets the canonical layer snap a
	// page onto itself. The reason is arithmetic: a realised candidate whose name
	// equals the plan entry's name is, by construction, a name the plan carries,
	// so planNames[rname] is true wherever rname == lname is. That refutes the
	// remedy the 2026-08-19 investigation proposed for bugs_open/215's same-name
	// hole ("drop rname == lname from the canonical layer's eligibility") — that
	// edit is INERT, and the hole is closed at Pass B2 instead. Keep both: each
	// states a different intent, and neither costs anything.
	eligible := func(rp map[string]interface{}) (string, bool) {
		rname, _ := rp["name"].(string)
		// The candidate IS the entry under consideration, not a twin of it. Its
		// identity is settled by Pass B2's same-name stamp, which keeps the
		// planner's copy; snapping here would replace the whole entry.
		if rname == "" || rname == lname {
			return rname, false
		}
		// The plan already carries the realised spelling as its own entry: the
		// two identities are both in play and collapsing them here would hand
		// the writer a duplicate to resolve by dropping one.
		if planNames[rname] {
			return rname, false
		}
		return rname, true
	}

	// observeOrSnap applies the deterministic layers' gate. When the site has not
	// opted in, the layer still records what it WOULD have done — each such row
	// is a second page identity about to be written, so the dark-launch count is
	// both the evidence for enabling and a live phantom warning.
	observeOrSnap := func(rp map[string]interface{}, layer string) (map[string]interface{}, string, bool) {
		if opts.TwinIdentitySnap {
			return snapPlanPageOntoRealised(lm, rp, layer, counts, logger), layer, true
		}
		rname, _ := rp["name"].(string)
		rurl, _ := rp["url"].(string)
		counts.TwinIdentityObserved++
		logger.Info("validate: twin identity observed, layer disabled",
			zap.String("layer", layer),
			zap.String("llm_name", lname), zap.String("realised_name", rname))
		counts.IdentitySnaps = append(counts.IdentitySnaps, identitySnap{
			Layer: layer + "_observed", PlanName: lname, PlanURL: lurl,
			RealisedName: rname, RealisedURL: rurl,
			Reason: "twin_identity_snap is off; this plan entry denotes a page already realised under another spelling and will otherwise be written as a second identity for it",
		})
		return nil, "", false
	}

	if rp, ok := byPathKey[datahelpers.PagePathKey(lurl)]; ok && lurl != "" {
		if _, good := eligible(rp); good {
			return observeOrSnap(rp, "path_key")
		}
	}

	// The plan entry's own canonical identity, derived exactly as the write path
	// will derive it, so this layer predicts that collapse rather than inventing
	// a new one. That means SLUG-then-name, not name alone: both write surfaces
	// canonicalise firstNonEmpty(slug, name), and an entry whose slug disagrees
	// with its name collapses to the slug's answer. Measured on fundamentallyai
	// 2026-08-11 (PLAN_PAGE_MERGE_LOSSY rows at 10:21:47): an entry NAMED
	// "tool-model-approach-selector-guide" canonicalised to the bare
	// "model-approach-selector-guide" because its slug said so. Predicting from
	// the name alone would have modelled a collapse the writer does not perform.
	canonSource := firstNonEmpty(datahelpers.GetStringField(lm, "slug", ""), lname)
	if canon := datahelpers.PageCanonicalNameForRow(canonSource, ltype); canon != "" {
		if rp, ok := byCanonName[canon]; ok {
			if _, good := eligible(rp); good {
				return observeOrSnap(rp, "canonical_name")
			}
		}
		// The realised row may hold the canonical name directly while the plan
		// entry is the bare twin (llm-cost-calculator -> tool-llm-cost-calculator).
		if rp, ok := byName[canon]; ok && canon != lname {
			if _, good := eligible(rp); good {
				return observeOrSnap(rp, "canonical_name")
			}
		}
	}

	// Stem layer. Requires EXACTLY ONE side to carry a role prefix — the guard
	// that makes the on-record false positive (two differently-prefixed pages on
	// one topic, "tool-pricing" beside "guide-pricing") structurally unmatchable
	// rather than merely unlikely.
	//
	// BOTH DIRECTIONS, because the direction flips in real data and a fixed one
	// would miss half the population. Measured on fundamentallyai 2026-08-11: the
	// plan carried prefixed "tool-tools" against bare realised "tools", AND bare
	// "ai-readiness-checker-guide" against prefixed realised
	// "tool-ai-readiness-checker-guide". The invariant is two identities for one
	// page, not a fixed prefix.
	stem := datahelpers.PageItemStem(lname)
	if stem == "" {
		return nil, "", false
	}
	planIsBare := stem == strings.ToLower(strings.TrimSpace(lname))
	var rp map[string]interface{}
	var ok bool
	if planIsBare {
		// Bare plan name -> a PREFIXED realised page on the same stem.
		rp, ok = byStem[stem]
	} else {
		// Prefixed plan name -> the BARE realised page whose name IS the stem.
		// Looked up by name rather than by the stem index, which deliberately
		// holds prefixed names only; requiring the hit to be the bare form keeps
		// the exactly-one-prefix rule intact.
		rp, ok = byName[stem]
	}
	if !ok {
		return nil, "", false
	}
	rname, good := eligible(rp)
	if !good {
		if rname != "" {
			counts.StemTwinAmbiguous++
			counts.IdentitySnaps = append(counts.IdentitySnaps, identitySnap{
				Layer: "stem_twin_refused", PlanName: lname, PlanURL: lurl,
				RealisedName: rname,
				Reason:       "the plan already carries the realised spelling as its own entry, or the candidate is the entry itself — collapsing them is a remediation decision, not a reconciliation",
			})
		}
		return nil, "", false
	}
	// A realised twin that has never shipped is not evidence of identity: an
	// unshipped catalogued sibling may simply be a different page nobody has
	// built yet. The phantom this layer exists to prevent is always the twin of
	// a page that IS serving.
	if !realisedPageHasShipped(rp) {
		counts.StemTwinAmbiguous++
		counts.IdentitySnaps = append(counts.IdentitySnaps, identitySnap{
			Layer: "stem_twin_refused", PlanName: lname, PlanURL: lurl,
			RealisedName: rname,
			Reason:       "the realised stem twin has never shipped, so a shared stem is not evidence the two are one page",
		})
		return nil, "", false
	}
	if !opts.StemTwinSnap {
		counts.StemTwinObserved++
		rurl, _ := rp["url"].(string)
		logger.Info("validate: stem twin observed, layer disabled",
			zap.String("llm_name", lname), zap.String("realised_name", rname))
		counts.IdentitySnaps = append(counts.IdentitySnaps, identitySnap{
			Layer: "stem_twin_observed", PlanName: lname, PlanURL: lurl,
			RealisedName: rname, RealisedURL: rurl,
			Reason: "stem_twin_snap is off; this plan entry would have been snapped onto the realised page and will otherwise be written as a second identity for it",
		})
		return nil, "", false
	}
	return snapPlanPageOntoRealised(lm, rp, "stem_twin", counts, logger), "stem_twin", true
}

// reconcileCounts is what reconcilePlanWithRealised observed, for the caller's
// log and its durable record.
//
// Named fields rather than the five positional ints this used to return: the
// fact carry adds two more, and a counter whose meaning silently changes is
// exactly how a later measurement goes wrong — the hazard documented on
// sameSectionList, which was live and unnoticed the moment seed 333 shipped.
type reconcileCounts struct {
	// Unioned: preserved realised pages the LLM omitted, added back (Pass A).
	Unioned int
	// DroppedCollision: LLM pages dropped by Pass C / C2.
	DroppedCollision int
	// DroppedPages names them. A counter says how many pages this function
	// deleted; it cannot say WHICH, and "which" is the only thing that tells a
	// reader whether a plan lost something it needed (bugs_open/463 — five
	// article pages vanished between plan_site and validate_plan with nothing
	// durable written anywhere, and the count alone was indistinguishable from a
	// planner that never proposed them).
	DroppedPages []droppedPlanPage
	// SnappedRename: pages snapped back to a realised identity (Pass B).
	SnappedRename int
	// SnappedSections: pages whose proposed COMPOSITION was overridden by the
	// realised one, or forced back to empty (Pass B2). Composition changes
	// only — a page whose proposed entries name the same components in the same
	// order is not counted however differently it was shaped.
	SnappedSections int
	// SectionFactsCarried: section entries that received a plan-time fact
	// assignment carried off a discarded LLM entry. An assignment that survived
	// because the planner's names already matched the realised ones is NOT
	// counted — nothing had to carry it, so this is a count of RESCUES, not of
	// scoped sections. Read site_plan_sections.assigned_fact_ids for the latter.
	SectionFactsCarried int
	// FactCarryMisses: per page, the assignments that matched no restored
	// section name. Non-empty means the planner scoped facts to a section the
	// built page does not have and they were discarded.
	FactCarryMisses []factCarryMiss
	// FactAssignmentAbsent: per page, object-form section entries that resolved
	// a name but carried no usable `facts` value (key absent, null, or not an
	// array). Seed 333 makes the key mandatory on every emitted section, so
	// these are planner disobedience — recorded under a distinct code because a
	// silently skipped entry is indistinguishable from a page correctly
	// assigned no facts (council round a06ff850, objection §3.5).
	FactAssignmentAbsent []factCarryMiss
	// SnappedIdentityPathKey / SnappedIdentityCanonName / SnappedStemTwin: pages
	// snapped onto a realised identity by one of the three twin-identity layers
	// (bugs_open/215 quiet mode). Counted separately from SnappedRename, which
	// is the exact-URL match that has always been there: which LAYER fired is
	// the thing a later measurement needs, because they carry different
	// confidence and the weakest is separately gated.
	SnappedIdentityPathKey   int
	SnappedIdentityCanonName int
	SnappedStemTwin          int
	// TwinIdentityObserved: twins the two DETERMINISTIC layers found while they
	// were disabled — the same dark-launch signal as StemTwinObserved, for the
	// layers that shipped default-ON in the first draft until the council's
	// guardian and architecture seats pointed out that an argument is not a
	// measurement.
	TwinIdentityObserved int
	// StemTwinObserved: stem twins seen while the stem layer was DISABLED. This
	// is the dark-launch signal — it measures how often the layer WOULD fire in
	// production before anyone turns it on, which is the only honest way to
	// answer "is the false-positive risk real here" for a heuristic the estate
	// has declined to trust fleet-wide since 2026-07-20.
	StemTwinObserved int
	// StemTwinAmbiguous: stem twins refused because more than one realised page
	// claimed the stem, or because both sides had shipped. A refusal is a
	// decision not to guess, and it is recorded for the same reason a merge is:
	// silence here would be indistinguishable from "no twin existed".
	StemTwinAmbiguous int
	// IdentitySnaps: the detail behind the three counters above, one per event
	// (snaps and refusals both). Persisted durably by the caller — a chassis pod
	// retains under a second of log, so the Info lines these carry are not a
	// record of anything (bugs_open/136 §11).
	IdentitySnaps []identitySnap
	// SameNameStamps: plan entries paired with a realised page by EXACT name and
	// stamped with its stored identity (bugs_open/215, the same-name hole). Every
	// pairing is recorded, but only those whose WouldDeriveName is non-empty
	// changed an outcome — the rest are pages the canonicaliser would have
	// reproduced anyway. The caller records the diverging subset durably; a row
	// per stamp would be ~17 rows per re-plan on one site and ~31 on another, for
	// an event that is the fix working normally.
	SameNameStamps []sameNameStamp
	// SameNameTypeConflicts: same-name pairs where the plan's role and the
	// realised page's page_type disagree, so no identity was stamped. Recorded
	// per event, because unlike a stamp this is a state nobody intended: two
	// pages, one name, two roles.
	SameNameTypeConflicts []sameNameTypeConflict
}

// droppedPlanPage is one page this function DELETED from the plan, with enough
// identity to be traceable after the orchestration row expires (~24h).
//
// Carries the PASS deliberately: Pass C ("this page's path is a section index's
// path") and Pass C2 ("this page re-proposes an adopted item topic") are
// different judgements with different failure modes, and a shared counter made
// them one event. bugs_open/463 is what a Pass C drop looks like when nobody can
// see which pass fired — five planned article pages disappeared with a green
// status and an unremarkable count.
//
// This is the PRODUCER-side record and is deliberately narrow. The durable,
// strategy-level account of what a plan lost is the 428 lane's
// recommended_type_reconciliation.go, which classifies by STAGE rather than by
// pass — a vocabulary that survives these passes being renumbered. The two are
// complementary, not a duplication: that one answers "was a recommended page
// type lost, and where", this one answers "which page, and by which rule".
type droppedPlanPage struct {
	Name   string
	URL    string
	Pass   string // "C" | "C2"
	Reason string
}

// identitySnap is one twin-identity event: a plan page recognised as denoting
// an already-realised page under a different spelling, and what was done about
// it. Carries BOTH spellings and BOTH urls so the event stays diagnosable after
// the orchestration row expires (~24h) — the lesson bugs_open/215 recorded
// against its own first verification step.
type identitySnap struct {
	Layer        string // "path_key" | "canonical_name" | "stem_twin" | "stem_twin_observed" | "stem_twin_refused"
	PlanName     string
	PlanURL      string
	RealisedName string
	RealisedURL  string
	// Reason is set on the refusal/observation rows only: why nothing was done.
	Reason string
}

// sameNameStamp is one exact-name pairing that received the realised identity
// marker (bugs_open/215). WouldDeriveName is the name CanonicalisePage would have
// produced for this entry, and is empty when that equals the stored name — i.e. it
// is set only for the pages where the stamp actually prevented a second identity.
// Carries the stored URL so the event stays diagnosable after the orchestration
// row expires.
type sameNameStamp struct {
	PlanName        string
	StoredURL       string
	WouldDeriveName string
}

// sameNameTypeConflict is an exact-name pair the stamp REFUSED because the two
// sides disagree about the page's role. Carries both types so the reader can tell
// planner disobedience from a deliberate retype without re-running the plan.
type sameNameTypeConflict struct {
	PlanName     string
	PlanType     string
	RealisedType string
	RealisedURL  string
}

// reconcileOptions carries the parts of reconcilePlanWithRealised's behaviour a
// caller may switch on. Zero value is the conservative one: everything that
// changes what a plan MEANS beyond the existing passes stays off until a step
// config asks for it.
//
// StemTwinSnap enables the stem layer (bare-vs-prefixed twins). It is an OPT-IN
// FIELD with the unsafe default OFF, per the owner ruling of 2026-08-02: new
// authority on a shared seam ships as a field, not as a documented contract,
// because a comment is not a control on a tree this many sessions share. The
// stem key is the one layer that can pair two genuinely different pages, and
// the reconciler has declined to run it on re-plans since 2026-07-20 (Pass C2's
// header). Off, the layer still MEASURES itself — see StemTwinObserved.
type reconcileOptions struct {
	// TwinIdentitySnap gates the two deterministic layers (path key, canonical
	// identity). See siteIdentityPolicy for why these are gated rather than
	// default-on: the council objected that changing matching behaviour for every
	// existing caller is architecture-scope however sound the argument, and that
	// the new collapse population deserved the same dark-launch measurement the
	// stem layer was given. Off, both layers still count what they would do.
	TwinIdentitySnap bool
	StemTwinSnap     bool
}

// reconcilePlanWithRealised enforces preservation of and convergence on the
// realised pages a re-plan must not silently redesign or drop.
//
// PRESERVATION SET (widened 2026-07-19, bugs_open/001). Formerly this was the
// first-plan subset alone, which made the whole function a no-op on every
// re-plan. NOTE what that flag actually is (renamed 2026-07-22 from the
// misleading "adoption_locked", bugs_open/051): NOT a per-page or 90-day lock —
// there never was one. The live load_existing_pages query surfaces it as
// site_has_no_current_plan, derived per SITE, so it is true for every page on a
// site's FIRST plan and false for every page on every re-plan after that. The two-branch design in 053 §054 (branch (b): a live
// timed per-page preserve-directive) is absent from the live query and has zero
// rows behind it fleet-wide, so no per-page lock has ever existed. The realised
// composition of a BUILT page was then carried by nothing: a page the LLM
// re-proposed under the same name was silently re-composed to whatever the LLM
// proposed that run, and a page the LLM omitted was dropped from the plan
// outright. Proven on idea.uk 2026-07-14 (plan 32be2797 -> ff03bdef): four
// built pages regressed, two of which re-rendered and re-deployed the regressed
// artefact.
//
// A built page deserves preservation whether or not it is on the first plan, so
// the set is now:
//
//	site_has_no_current_plan == true  OR  build_status IN ("deployed", "needs_rebuild")
//
// needs_rebuild joined the set 2026-07-21 (bugs_open/037). A needs_rebuild page
// still holds its intended composition in pages.sections, and EVERY writer of
// that status keeps those sections and means "re-render this page as planned",
// never "recompose it from scratch": a refused 0-component or partial deploy
// (UpdatePageStatusAction — clears built_from_plan_version but keeps sections),
// an image/maintenance rebuild (flagPagesForRebuild), or a now-available
// component the sections already name (markPagesForRebuild,
// check_unresolved_sections — those two would be actively DEFEATED by
// recomposition, since the sections name the very components the rebuild exists
// to pick up). So letting a re-plan take the LLM's composition for a
// needs_rebuild page was silent loss, not an honoured redesign request. This
// widens only MEMBERSHIP of the preserved set (realisedPageCompositionIsPreserved);
// the empty-sections classification in Pass B/B2 still keys on realisedPageHasShipped
// (== deployed), because a needs_rebuild page with empty sections may be either
// rendered-elsewhere OR genuinely awaiting composition, which Pass B2's non-empty
// gate already routes correctly (see bugs_open/050 for the deployed-empty case).
//
// All flags come from the load_existing_pages query. build_status is only
// surfaced by that query as of migration 173 — if it is absent both status terms
// are empty and behaviour falls back to the first-plan set, so the Go change
// and the query change are safe to land in either order.
//
// Passes over the preservation set:
//
//	Pass C  — section-collision dedup: drop an LLM page whose slug equals the
//	          stem of a realised section index ("games" vs "games-index").
//	Pass C2 — item-topic dedup. Deliberately still scoped to the FIRST-PLAN
//	          subset, not the widened set: it is a name-stem heuristic, and a
//	          false positive suppresses a legitimately new page (a new
//	          "tool-pricing" beside a built "guide-pricing" shares the stem
//	          "pricing"). Made permanent for every built page that risk is not
//	          acceptable, and it is not needed for this bug — invented pages carry
//	          new topics and so collide with nothing. See bugs_open/001 "pages
//	          invented", which this does not claim to fix.
//	          CORRECTED 2026-07-20 (bugs_open/051): this used to read "bounded to
//	          the 90-day window that risk is acceptable". There is no 90-day
//	          window — see the preservation-set note above. Because noCurrentPlanPages
//	          is empty whenever the site has a current plan, Pass C2 can fire ONLY on
//	          a site's first plan and never on a re-plan. The scoping decision
//	          stands; the reason given for it did not exist.
//	Pass B  — rename snap-back: same URL as a realised page, different name ->
//	          replace with the realised identity. Its sections are carried too,
//	          EXCEPT when the realised page is NOT deployed and its sections are
//	          empty: that is a catalogued page that has never been composed, so
//	          keep the realised identity but take the LLM's proposed sections so
//	          the re-plan can finally compose it (bugs_open/050).
//	Pass B2 — composition snap-back: same NAME as a realised page. Reconciles the
//	          LLM's sections against the realised composition, keyed on the
//	          realised sections and deployed-ness (bugs_open/050):
//	            NON-EMPTY            -> restore those sections over the LLM's; a
//	                                    built page must not be re-composed.
//	            EMPTY + deployed     -> force the LLM's proposal back to empty; the
//	                                    page renders through another subsystem (a
//	                                    tool or blog-index page) and must not
//	                                    receive an injected generic layout.
//	            EMPTY + not-deployed -> keep the LLM's proposal; a catalogued page
//	                                    is finally composed.
//	          Carrying emptiness forward unconditionally is what made "re-plan to
//	          compose the missing pages" structurally impossible (bugs_open/001,
//	          second defect); composing onto a deployed sectionless page is the
//	          injection risk bugs_open/050 closes. For a deployed page, empty
//	          sections is a positive statement ("not section-composed here"), not
//	          an absence awaiting composition.
//	Pass A  — union: append every preserved realised page not already present.
//
// Returns the reconciled page slice plus counts for logging.
func reconcilePlanWithRealised(
	llmPages []interface{},
	existingPages []interface{},
	opts reconcileOptions,
	logger *zap.Logger,
) ([]interface{}, reconcileCounts) {
	var counts reconcileCounts
	// Force-preserve first-plan (no-current-plan) OR built pages (see header).
	var preserved []interface{}
	var noCurrentPlanPages []interface{}
	hasShippedColumnSeen := false
	for _, rp := range existingPages {
		rm, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		// The has_shipped column is a property of the QUERY, not the row, so its
		// absence is reported once per run — the honest surface for the fallback
		// realisedPageHasShipped otherwise takes silently (bugs_open/185 tranche 3,
		// bug_historian's advisory: "nothing distinguishes 'old semantics
		// intended' from 'migration 302 reverted and the gate has quietly gone
		// back to the buggy predicate'"). Warn, not Debug: on the live query the
		// column has been present since 2026-08-03, so its absence now is a
		// regression, and one Warn per re-plan is not a hot-path cost.
		if !hasShippedColumnSeen {
			hasShippedColumnSeen = true
			if _, present := rm["has_shipped"]; !present {
				logger.Warn("load_existing_pages rows carry no has_shipped column; realisedPageHasShipped is degrading to the narrow build_status-only test, which misreads a shipped needs_rebuild page as uncomposed — is migration 302 absent or reverted, or is this caller wired to a different loader?",
					zap.Int("existing_pages", len(existingPages)))
			}
		}
		noCurrentPlan := noCurrentPlanFlag(rm)
		if noCurrentPlan {
			noCurrentPlanPages = append(noCurrentPlanPages, rp)
		}
		if noCurrentPlan || realisedPageCompositionIsPreserved(rm) {
			preserved = append(preserved, rp)
		}
	}
	// The IDENTITY index, built over EVERY realised row rather than the preserved
	// subset (bugs_open/340). The preservation set answers "whose COMPOSITION must
	// a re-plan not redesign?"; who owns a page's NAME is a different question and
	// has no reason to inherit the narrower answer. A row that is neither deployed
	// nor needs_rebuild, on a site that already has a current plan, is invisible to
	// every map below — so before this index existed, the write path re-derived its
	// name, missed on (site_id, name), and INSERTed the twin the same-name stamp
	// exists to prevent.
	//
	// Used by the same-name stamp ONLY, and that restriction is the whole safety
	// argument. The stamp writes identity fields and never touches sections, so
	// widening what it can see cannot move a composition. The twin layers snap the
	// WHOLE entry through snapPlanPageOntoRealised, which carries the realised
	// composition with it — widening THEIR input would let a plan entry inherit the
	// empty composition of a page nobody has built yet, which is a different change
	// needing its own measurement. bugs_open/340 listed widening all four maps as
	// its candidate 1; this is its candidate 2, chosen for exactly that coupling.
	realisedByNameAll := make(map[string]map[string]interface{})
	for _, rp := range existingPages {
		if rm, ok := rp.(map[string]interface{}); ok {
			if n, _ := rm["name"].(string); n != "" {
				realisedByNameAll[n] = rm
			}
		}
	}

	if len(preserved) == 0 {
		// Nothing realised worth converging on: a genuinely from-scratch build.
		// Leave the LLM plan untouched.
		//
		// LIMITATION, stated rather than silently closed: the identity index above
		// is abandoned here too, so a site on which NO page is deployed or
		// needs_rebuild keeps the pre-340 behaviour. That is the narrow residue of
		// bugs_open/340 — every page on such a site is unbuilt, so the twin is a
		// twin of nothing anyone has served. Returning early is what makes a
		// from-scratch build cheap, and changing it would alter every one of them.
		return llmPages, counts
	}
	existingPages = preserved

	realisedByURL := make(map[string]map[string]interface{})
	realisedByName := make(map[string]map[string]interface{})
	// Two indexes of the realised section hubs, because Pass C asks its question
	// two ways. sectionIndexPaths is keyed on the path the hub actually CLAIMS
	// and is the one that runs; sectionStems is keyed on its first path segment
	// and now survives only as the fallback for a plan page that carries no URL
	// at all (bugs_open/463 — a first-segment comparison cannot tell a child
	// from a collider, because both reduce to the section name).
	sectionStems := make(map[string]string)      // stem      -> realised index name
	sectionIndexPaths := make(map[string]string) // path key  -> realised index name
	for _, rp := range existingPages {
		rm, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := rm["name"].(string)
		url, _ := rm["url"].(string)
		pageType, _ := rm["page_type"].(string)
		if url != "" {
			realisedByURL[url] = rm
		}
		if name != "" {
			realisedByName[name] = rm
		}
		if stem := sectionStemOf(name, url, pageType); stem != "" {
			sectionStems[stem] = name
			if k := sectionPathKey(url); k != "" {
				sectionIndexPaths[k] = name
			}
		}
	}

	// ── Twin-identity keys (bugs_open/215, quiet mode) ──────────────────────
	//
	// The three ways a plan page and a realised page can denote ONE page while
	// matching on neither exact URL (Pass B) nor exact name (Pass B2). Built
	// over the same preservation set Pass B uses, so a snap can only ever land
	// on a page this function already considers worth converging on.
	//
	// EVERY map REFUSES AMBIGUITY: if two realised rows claim one key, the key
	// is deleted and nothing matches it. That is the convention
	// buildCanonicalPageNameMap set for exactly this hazard — when two pages
	// want one alias, guessing is worse than missing, because a wrong snap
	// suppresses a real page while a miss only leaves the twin unreconciled.
	realisedByPathKey := make(map[string]map[string]interface{})
	realisedByCanonName := make(map[string]map[string]interface{})
	realisedByStem := make(map[string]map[string]interface{})
	ambiguousPathKey := make(map[string]bool)
	ambiguousCanonName := make(map[string]bool)
	ambiguousStem := make(map[string]bool)
	claim := func(m map[string]map[string]interface{}, ambiguous map[string]bool, key string, rm map[string]interface{}) {
		if key == "" || ambiguous[key] {
			return
		}
		if prev, taken := m[key]; taken {
			if prevName, _ := prev["name"].(string); prevName != "" {
				if thisName, _ := rm["name"].(string); thisName == prevName {
					return // same row seen twice, not a contest
				}
			}
			delete(m, key)
			ambiguous[key] = true
			return
		}
		m[key] = rm
	}
	for _, rp := range existingPages {
		rm, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := rm["name"].(string)
		url, _ := rm["url"].(string)
		pageType, _ := rm["page_type"].(string)
		if name == "" {
			continue
		}
		claim(realisedByPathKey, ambiguousPathKey, datahelpers.PagePathKey(url), rm)
		// The canonical-name key only carries a signal when it DIFFERS from the
		// row's own name — otherwise it duplicates Pass B2's exact-name match and
		// would shadow it with a weaker one.
		if canon := datahelpers.PageCanonicalNameForRow(name, pageType); canon != "" && canon != name {
			claim(realisedByCanonName, ambiguousCanonName, canon, rm)
		}
		// The stem key is keyed on PREFIXED realised names only. A bare realised
		// name has stem == name, which Pass B2 already matches exactly; indexing
		// it here would let a bare plan page match a bare realised page it is
		// simply equal to, and would let two bare siblings contest a key for no
		// reason.
		if stem := datahelpers.PageItemStem(name); stem != "" && stem != strings.ToLower(strings.TrimSpace(name)) {
			claim(realisedByStem, ambiguousStem, stem, rm)
		}
	}

	// Item-topic stems: the role-prefix-stripped name stem of each realised
	// page (guide-economy-basics -> economy-basics). Keyed to a SET of realised
	// names so a topic legitimately covered by two adopted pages (e.g. a tool
	// and a guide) does not false-positive on either of them. Lets Pass C2 drop
	// an LLM page that re-proposes an adopted item under a different
	// prefix/role/URL.
	//
	// Built deliberately from noCurrentPlanPages, NOT the widened preservation
	// set — see the Pass C2 note in the header for why this one heuristic stays
	// narrow. In practice that makes it first-plan-only: noCurrentPlanPages is
	// empty whenever the site has a current plan (bugs_open/051).
	itemStemSets := make(map[string]map[string]bool)
	for _, rp := range noCurrentPlanPages {
		rm, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := rm["name"].(string)
		if name == "" {
			continue
		}
		stem := itemStemOf(name)
		if stem == "" {
			continue
		}
		if itemStemSets[stem] == nil {
			itemStemSets[stem] = make(map[string]bool)
		}
		itemStemSets[stem][name] = true
	}

	// Every name the LLM proposed, for the twin layers' both-sides-in-plan
	// refusal. A plan that already carries BOTH spellings needs no snap: snapping
	// one onto the other would hand two entries the same name, and the writer's
	// richer-wins dedup would then resolve the pair by dropping one — which for a
	// pair of DEPLOYED twins means silently evicting a live page from plan
	// governance. Measured on robot-hands 2026-08-11: three such pairs, both
	// sides deployed, both sides in the current plan. That is a remediation
	// decision (which page survives, what redirects), not a reconciliation.
	planNames := make(map[string]bool, len(llmPages))
	for _, lp := range llmPages {
		if lm, ok := lp.(map[string]interface{}); ok {
			if n, _ := lm["name"].(string); n != "" {
				planNames[n] = true
			}
		}
	}

	// Pass C (collision) + Pass B (rename) + Pass B2 (composition) over the
	// LLM pages.
	var kept []interface{}
	for _, lp := range llmPages {
		lm, ok := lp.(map[string]interface{})
		if !ok {
			kept = append(kept, lp)
			continue
		}
		// This function is the ONLY minter of the realised-identity marker: it is
		// what tells the write path to stop re-deriving a page's name and URL, so
		// an LLM (or a replayed plan) that emits it must not be believed. Stripped
		// before any pass can read it, and re-stamped only by a snap, a union, or
		// Pass B2's same-name stamp (bugs_open/215 — the exact-name case, which no
		// snap can reach because the twin layers refuse a candidate equal to the
		// plan's own name).
		delete(lm, "identity_authority")
		delete(lm, "from_realised")
		lname, _ := lm["name"].(string)
		lurl, _ := lm["url"].(string)
		ltype, _ := lm["page_type"].(string)
		lslug := slugOf(lname, lurl)

		// Pass C: a page whose OWN PATH is a realised section index's path — a
		// flat /articles.html beside a realised /articles/index.html, which both
		// claim "/articles".
		//
		// A CHILD of that index claims a LONGER path ("/articles/x") and is
		// KEPT. That distinction is the whole of bugs_open/463: this comparison
		// used to run on the FIRST PATH SEGMENT of each side, under which a
		// child and a collider are the same string, so every newly planned child
		// of a section index was deleted here. It went unnoticed for three
		// months because Pass A's union restores REALISED pages immediately
		// afterwards — so on an established site the drop is invisible, and only
		// a hub that is empty TODAY can never be filled. [MEASURED 2026-09-03]
		// 53 of 78 section-index hubs fleet-wide, across 21 sites, had zero
		// children under their prefix.
		//
		// The path key is also what makes this pass and bugs_open/444's listing
		// gate agree BY CONSTRUCTION rather than by coincidence. The gate, ~350
		// lines below, counts a hub's children with
		// strings.HasPrefix(url, sectionPrefixOf(hub)); "claims the hub's path"
		// and "lives under the hub" are mutually exclusive and together partition
		// the old first-segment test. Before this, Pass C deleted the children
		// and the gate then held the childless hub, each guard's evidence reading
		// as a reason for the other.
		//
		// bugs_closed/141's behaviour is preserved exactly: a flat /news.html
		// re-proposed beside a realised /news/index.html still drops, because
		// both claim "/news". [MEASURED 2026-09-03] all 83 realised hubs
		// fleet-wide have a name-derived stem matching the one their URL yields,
		// so on today's estate this rule drops a strict SUBSET of what the old
		// one dropped — it can stop a drop, never start one. It could start one
		// only for a hub whose stored name disagrees with its served path, where
		// it would be right to: it compares where the hub actually serves.
		idxName, isCollision := "", false
		if k := sectionPathKey(lurl); k != "" {
			idxName, isCollision = sectionIndexPaths[k]
		} else {
			// No URL to compare. The name is the only signal left, which is what
			// this pass has always used — unchanged, deliberately.
			idxName, isCollision = sectionStems[lslug]
		}
		if isCollision && !isSectionIndexType(ltype) && lname != idxName {
			logger.Info("validate: dropped flat page colliding with realised section index",
				zap.String("dropped", lname), zap.String("dropped_url", lurl),
				zap.String("kept_index", idxName))
			counts.DroppedCollision++
			counts.DroppedPages = append(counts.DroppedPages, droppedPlanPage{
				Name: lname, URL: lurl, Pass: "C", Reason: "path collides with realised section index " + idxName,
			})
			continue
		}

		// Pass C2: item-topic collision - the LLM re-proposes an adopted item
		// under a different name/prefix/role (e.g. "economy-basics" beside the
		// adopted "guide-economy-basics"; different URL, so Pass B misses it).
		// Drop it; the adopted page already covers the topic. Skips when the LLM
		// name IS one of the realised names for that stem (a preserved page).
		if names, isStem := itemStemSets[itemStemOf(lname)]; isStem && !names[lname] {
			logger.Info("validate: dropped page duplicating an adopted item topic",
				zap.String("dropped", lname),
				zap.String("stem", itemStemOf(lname)))
			counts.DroppedCollision++
			counts.DroppedPages = append(counts.DroppedPages, droppedPlanPage{
				Name: lname, URL: lurl, Pass: "C2", Reason: "duplicates adopted item topic " + itemStemOf(lname),
			})
			continue
		}

		// Pass B: same URL as a realised page, different name -> snap back to the
		// realised identity, carrying its sections. Exception (bugs_open/050): a
		// NOT-deployed realised page with empty sections is a catalogued page that
		// has never been composed, so keep the realised identity but take the LLM's
		// proposed sections. A DEPLOYED empty page renders through another subsystem
		// and its emptiness is authoritative — carry it (as normalise already does).
		if rp, ok := realisedByURL[lurl]; ok {
			if rname, _ := rp["name"].(string); rname != "" && rname != lname {
				kept = append(kept, snapPlanPageOntoRealised(lm, rp, "url_exact", &counts, logger))
				counts.SnappedRename++
				continue
			}
		}

		// ── Twin-identity layers (bugs_open/215, quiet mode) ────────────────
		//
		// Same question Pass B just asked — "is this plan entry actually a page
		// we have already realised?" — with comparators that can see the two
		// ways the answer is yes while the URLs differ. Ordered strongest key
		// first; first hit wins and the arm is the same one Pass B uses, so a
		// page reached by any layer is treated identically thereafter.
		//
		// Why this is not Pass C2 widened: C2 DROPS the plan entry, and has
		// stayed first-plan-only since 2026-07-20 because a false positive
		// there destroys a legitimately new page. These layers SNAP instead —
		// the entry survives under the realised identity, carrying its fact
		// assignments — so the cost of being wrong is bounded by the same
		// preservation rules a rename already obeys.
		if snapped, layer, ok := matchTwinIdentity(lm, lname, lurl, ltype,
			realisedByPathKey, realisedByCanonName, realisedByStem,
			realisedByName, planNames, opts, &counts, logger); ok {
			kept = append(kept, snapped)
			switch layer {
			case "path_key":
				counts.SnappedIdentityPathKey++
			case "canonical_name":
				counts.SnappedIdentityCanonName++
			case "stem_twin":
				counts.SnappedStemTwin++
			}
			continue
		}

		// Pass B2: same NAME as a preserved realised page -> stamp the realised
		// IDENTITY (bugs_open/215's same-name hole; see
		// stampSameNameRealisedIdentity), then reconcile the LLM's sections
		// against the realised composition (bugs_open/050). Beyond the identity
		// fields, only "sections" is touched — title/meta/nav stay the LLM's, so a
		// re-plan can still refresh copy and navigation without touching the
		// layout. That is why this case is NOT routed through
		// snapPlanPageOntoRealised: a snap replaces the whole entry and would
		// throw the refreshed copy away for every page the planner named
		// correctly.
		//   - realised NON-EMPTY: restore those sections over the LLM's; a page
		//     built through the section composer must not be re-composed.
		//   - realised EMPTY + deployed: force the LLM's proposal back to empty; the
		//     page renders through another subsystem (a tool or blog-index page)
		//     and must not receive an injected generic layout.
		//   - realised EMPTY + not-deployed: keep the LLM's proposal (fall through);
		//     a catalogued page is finally composed.
		if rp, ok := realisedByName[lname]; !ok {
			// Not in the preservation set, but the site may still hold a row under
			// this exact name — an unbuilt or failed one (bugs_open/340). Its
			// IDENTITY is still authoritative: the (site_id, name) upsert will
			// collide with it whatever its build_status says, so re-deriving the
			// name here is what mints the twin. Stamp identity and nothing else —
			// there is deliberately no composition reconciliation against a row
			// the preservation rules excluded.
			if wider, widerOK := realisedByNameAll[lname]; widerOK {
				stampSameNameRealisedIdentity(lm, wider, &counts, logger)
			}
		} else {
			// Identity first, and independently of composition: the planner naming
			// a page exactly as it is stored is the case no twin layer can reach
			// (their shared eligible closure refuses rname == lname, rightly — such
			// a candidate is the page itself, not a twin of it). Runs on every
			// branch below, including the fall-through, because whether a page's
			// composition is restored, forced empty or left alone says nothing
			// about who owns its name.
			stampSameNameRealisedIdentity(lm, rp, &counts, logger)
			if rs := realisedSectionsOf(rp); len(rs) > 0 {
				// Same names in the same order: the composition is unchanged, so
				// there is nothing to restore and the LLM's entries stay as they
				// are — which is how a fact assignment survives on a built page
				// for free once the planner re-emits the realised list (candidate
				// 1b (i)). Only a genuine composition change gets here.
				if !sameSectionList(lm["sections"], rs) {
					restored, carried, unmatched, absent := carrySectionFactsOntoRealised(rs, lm["sections"])
					logger.Info("validate: snapped built page composition back to realised sections",
						zap.String("page", lname),
						zap.Int("realised_sections", len(rs)),
						zap.Int("section_facts_carried", carried))
					lm["sections"] = restored
					counts.SnappedSections++
					counts.SectionFactsCarried += carried
					if len(unmatched) > 0 {
						counts.FactCarryMisses = append(counts.FactCarryMisses,
							factCarryMiss{Page: lname, Sections: unmatched})
					}
					if len(absent) > 0 {
						counts.FactAssignmentAbsent = append(counts.FactAssignmentAbsent,
							factCarryMiss{Page: lname, Sections: absent})
					}
				}
			} else if realisedPageHasShipped(rp) {
				if ls, ok := lm["sections"].([]interface{}); ok && len(ls) > 0 {
					// Deployed and sectionless is authoritative: the page renders
					// through another subsystem. There is no section to scope a
					// fact to, so any assignment here is discarded WITH its
					// entry — recorded, because a planner scoping facts to a page
					// that has no sections is a prompt fault worth seeing.
					var dropped []string
					for _, e := range ls {
						if obj, isObj := e.(map[string]interface{}); isObj {
							if _, hasFacts := obj["facts"].([]interface{}); hasFacts {
								if n, ok := sectionEntryName(obj); ok {
									dropped = append(dropped, n)
								}
							}
						}
					}
					logger.Info("validate: forced deployed sectionless page back to empty",
						zap.String("page", lname),
						zap.Int("llm_sections", len(ls)),
						zap.Int("section_facts_dropped", len(dropped)))
					lm["sections"] = []interface{}{}
					counts.SnappedSections++
					if len(dropped) > 0 {
						counts.FactCarryMisses = append(counts.FactCarryMisses,
							factCarryMiss{Page: lname, Sections: dropped})
					}
				}
			}
		}
		kept = append(kept, lm)
	}

	// Pass A: union — add preserved realised pages not present by name.
	presentName := make(map[string]bool)
	for _, p := range kept {
		if pm, ok := p.(map[string]interface{}); ok {
			if n, _ := pm["name"].(string); n != "" {
				presentName[n] = true
			}
		}
	}
	for _, rp := range existingPages {
		rm, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := rm["name"].(string)
		if name == "" || presentName[name] {
			continue
		}
		kept = append(kept, normaliseRealisedToPlanPage(rm))
		presentName[name] = true
		counts.Unioned++
	}

	return kept, counts
}

// realisedPageHasShipped reports whether a realised pages-table row (as
// returned by the load_existing_pages query) has ever been SERVED — which is
// the question both callers actually ask: "is this page's emptiness
// authoritative?" A page that shipped sectionless renders through another
// subsystem (a tool, a blog-index) and must not receive an injected generic
// layout; a page that has never shipped is merely uncomposed.
//
// It reads `has_shipped`, surfaced by migration 302's load_existing_pages query
// as NOT(datahelpers.NeverDeployedPagePredicate) — the estate's ONE definition
// of "has been served" (bugs_open/185 tranche 3). It falls back to the old
// build_status test when the column is absent, so the Go change and migration
// 302 can land in either order (the same degradation contract migration 173
// established for build_status itself). That fallback is NOT silent any more:
// reconcilePlanWithRealised warns ONCE per run when the loaded rows carry no
// has_shipped key (2026-08-15, closing bug_historian's tranche-3 advisory) —
// once, because the column's absence is a property of the query, not the row.
// This helper stays pure so it can be called per page without a log per page.
//
// > **CORRECTED 2026-08-03 — this function (then `realisedPageIsBuilt`) claimed
// > it "mirrors decideEmit's skip_built test … keep the two in step". That
// > coupling was WRONG, and it is removed rather than obeyed.** The two share a
// > spelling, not a question. decideEmit asks "does this page need a BUILD item
// > emitted?" — a needs_rebuild page must answer not_built there, so decideEmit
// > is CORRECT to test build_status != 'deployed' and must never widen. This
// > function asks "has this page been served?", where build_status='deployed'
// > misses a shipped needs_rebuild page still serving its previous artefact
// > (bugs_closed/037's own population: 35 of 46 carry a deployed_at). One
// > spelling, two questions — the same trap as bugs_open/185's detectors.
//
// It is deliberately DISTINCT from realisedPageCompositionIsPreserved below:
// that is preservation MEMBERSHIP (deployed + needs_rebuild by status, because
// their sections are their intended composition); this is the empty-page GATE.
// TestReconcile_NeedsRebuildEmptyPageIsStillComposable pins the difference:
// dartsonline's brands-index (needs_rebuild, never shipped, 0 sections) must
// stay composable — and does, because its deployed_at is NULL, so has_shipped
// is false. Only a needs_rebuild page that actually SERVED empty is gated.
// THE FALLBACK IS NOT A NO-OP PATH, and the council's editquality seat was right
// to ask: if `query_database` stringified its scanned columns, the bool assertion
// would silently fail and this fix would ship INERT — the worst outcome, because
// nothing errors. Checked, not assumed: QueryDatabaseAction
// (database_actions.go:100-107) scans into interface{} and converts ONLY
// []byte to string, passing every other driver value through untouched. lib/pq
// returns a Postgres boolean as a Go bool, so `has_shipped` arrives as a bool and
// the assertion holds. The existing `site_has_no_current_plan` column proves the
// same path in production — noCurrentPlanFlag reads it with the identical
// `.(bool)` assertion and has worked since migration 173's era.
func realisedPageHasShipped(rm map[string]interface{}) bool {
	if v, ok := rm["has_shipped"].(bool); ok {
		return v
	}
	status, _ := rm["build_status"].(string)
	return status == "deployed"
}

// noCurrentPlanFlag reports the load_existing_pages "site_has_no_current_plan"
// flag for a realised page: true when the page's site has no current plan yet —
// uniquely the site's FIRST plan after adoption (bugs_open/051). It force-preserves
// adopted pages through that one plan and is empty on every re-plan thereafter.
//
// Renamed 2026-07-22 from the misleading "adoption_locked" (there was never a
// per-page or 90-day lock — bugs_open/051). The live query now emits ONLY
// site_has_no_current_plan: migration 193 added it beside the old alias, the
// renamed chassis (v1.0.1151) went fleet-live reading it, then migration 194
// dropped the adoption_locked alias. The adoption_locked read below is KEPT as a
// defensive compat path — it is dead against the current query, costs nothing,
// resolves a snapshot rollback of 194, and is what the reconcile tests exercise
// (their fixtures set the old key). An absent flag degrades to false, matching
// realisedPageHasShipped's treatment of a missing column.
func noCurrentPlanFlag(rm map[string]interface{}) bool {
	if v, _ := rm["site_has_no_current_plan"].(bool); v {
		return true
	}
	v, _ := rm["adoption_locked"].(bool) // legacy alias, kept as a defensive compat read
	return v
}

// realisedPageCompositionIsPreserved reports whether a realised pages-table row
// carries a composition a re-plan must not silently discard. Two build states
// qualify, for the same reason — the page already holds an intended composition
// in pages.sections that machinery expects to survive a re-plan:
//
//	"deployed"      -- built and live.
//	"needs_rebuild" -- awaiting a re-render, but its sections ARE its intended
//	    composition. Every writer of needs_rebuild keeps pages.sections and means
//	    "re-render as planned", never "recompose from scratch" (bugs_open/037; see
//	    the reconcilePlanWithRealised header for the enumeration of writers).
//
// This is the preservation-MEMBERSHIP predicate, used by reconcilePlanWithRealised
// and the truncation guard. It is deliberately DISTINCT from realisedPageHasShipped
// (== deployed), which the empty-sections gate in Pass B/B2 keeps using: an empty
// needs_rebuild page may be genuinely awaiting composition rather than rendered
// elsewhere, so it must not be force-emptied. Pass B2's non-empty gate routes both
// kinds correctly — a needs_rebuild page with real sections is snapped back; one
// with empty sections falls through to the LLM's proposal, exactly as before this
// change (bugs_open/050 owns the deployed-empty classification).
//
// Returns false when build_status is absent, so the preservation set degrades
// safely to the first-plan set on a chassis whose load_existing_pages query
// has not been updated to surface the column.
func realisedPageCompositionIsPreserved(rm map[string]interface{}) bool {
	switch status, _ := rm["build_status"].(string); status {
	case "deployed", "needs_rebuild":
		return true
	default:
		return false
	}
}

// recomposePagesFromSpec reads the OPTIONAL `recompose_pages` list from the
// needs_site_plan trigger spec, at input_data.spec.recompose_pages (the work
// item's spec travels there unchanged — see features_open/012). It names realised
// pages the caller has DELIBERATELY asked to redesign, the explicit-intent escape
// hatch for bugs_open/037: `needs_rebuild`/`deployed` pages are otherwise
// preserved, so this is the only way a re-plan may recompose one. Returns nil when
// the field is absent (every ordinary re-plan), so behaviour is unchanged unless a
// caller opts in. Names match realised page names (pages.name), case-sensitive as
// elsewhere in the planner.
func recomposePagesFromSpec(collectedData map[string]interface{}, logger *zap.Logger) map[string]bool {
	raw := datahelpers.ExtractNestedField(collectedData, "input_data.spec.recompose_pages")
	list, ok := raw.([]interface{})
	if !ok || len(list) == 0 {
		return nil
	}
	set := make(map[string]bool, len(list))
	for _, v := range list {
		if name, ok := v.(string); ok && name != "" {
			set[name] = true
		}
	}
	if len(set) == 0 {
		return nil
	}
	names := make([]string, 0, len(set))
	for n := range set {
		names = append(names, n)
	}
	logger.Info("ValidateSitePlanAction: recompose_pages requested (explicit redesign intent)",
		zap.Strings("pages", names))
	return set
}

// filterOutRecomposePages drops every realised page named in the recompose set
// from the convergence input, so reconcilePlanWithRealised — and the truncation
// must-keep, which reads the same slice — treat those pages as from-scratch: the
// LLM's proposed composition governs, and the page may be redesigned or dropped
// per the LLM's plan. A named page matching no realised page is a harmless no-op.
func filterOutRecomposePages(existingPages []interface{}, recompose map[string]bool, logger *zap.Logger) []interface{} {
	if len(recompose) == 0 {
		return existingPages
	}
	kept := make([]interface{}, 0, len(existingPages))
	var released []string
	for _, rp := range existingPages {
		if rm, ok := rp.(map[string]interface{}); ok {
			if name, _ := rm["name"].(string); recompose[name] {
				released = append(released, name)
				continue
			}
		}
		kept = append(kept, rp)
	}
	if len(released) > 0 {
		logger.Info("ValidateSitePlanAction: recompose — realised pages released from the preserve guard for this re-plan",
			zap.Strings("pages", released))
	}
	return kept
}

// realisedSectionsOf extracts a realised page's section list. The
// load_existing_pages query runs via query_database, which stringifies jsonb, so
// "sections" normally arrives as a JSON string; a native []interface{} is
// tolerated too. Returns nil for missing, empty, or unparseable values — callers
// treat nil/empty as "no realised composition to preserve".
func realisedSectionsOf(rm map[string]interface{}) []interface{} {
	switch v := rm["sections"].(type) {
	case []interface{}:
		return v
	case string:
		if v == "" {
			return nil
		}
		var parsed []interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return nil
		}
		return parsed
	}
	return nil
}

// sameSectionList reports whether an LLM-proposed sections value already names
// the same components, in the same order, as the realised one — so Pass B2 only
// logs and counts a snap-back that actually changed the COMPOSITION.
//
// Compares by section NAME, not by rendered value, and that is a correction
// rather than a refinement. This used to compare fmt.Sprintf("%v", …) of whole
// entries. Under RFC_016's object form the planner's entries are maps carrying
// facts while the realised list is plain strings, so from the moment seed 333
// went live every composed page compared "changed": snappedSections silently
// stopped counting composition changes and started counting shape differences,
// and each one forced a pointless restore over an identical list. A name is the
// identity the two shapes share.
//
// A proposed entry of neither shape (a number, nil) yields no name and so reads
// as "changed", which sends the page to the restore branch and keeps the built
// composition. That is the safe direction, and it is the direction the previous
// implementation took too.
func sameSectionList(proposed interface{}, realised []interface{}) bool {
	pl, ok := proposed.([]interface{})
	if !ok || len(pl) != len(realised) {
		return false
	}
	for i := range pl {
		pn, pok := sectionEntryName(pl[i])
		rn, rok := sectionEntryName(realised[i])
		if !pok || !rok || pn != rn {
			return false
		}
	}
	return true
}

// truncatePreservingRealised caps the plan at maxPages but never drops a
// must-keep page — on the site's first plan or built (see the caller). Must-keep
// pages are kept first; net-new proposed pages fill the remaining budget in order.
func truncatePreservingRealised(
	pages, mustKeep []interface{},
	maxPages int,
	logger *zap.Logger,
) []interface{} {
	keepNames := make(map[string]bool)
	for _, rp := range mustKeep {
		if rm, ok := rp.(map[string]interface{}); ok {
			if n, _ := rm["name"].(string); n != "" {
				keepNames[n] = true
			}
		}
	}
	var keep, netNew []interface{}
	for _, p := range pages {
		name := ""
		if pm, ok := p.(map[string]interface{}); ok {
			name, _ = pm["name"].(string)
		}
		if keepNames[name] {
			keep = append(keep, p)
		} else {
			netNew = append(netNew, p)
		}
	}
	if len(keep) >= maxPages {
		logger.Warn("validate: preserved pages exceed max_pages; keeping all preserved, dropping all net-new",
			zap.Int("preserved", len(keep)), zap.Int("max_pages", maxPages))
		return keep
	}
	budget := maxPages - len(keep)
	if budget > len(netNew) {
		budget = len(netNew)
	}
	return append(keep, netNew[:budget]...)
}

// ----------------------------------------------------------------------------
// LLM array-item key reconciliation (page-content-writer safety net)
//
// Array components declare the per-element field names in their input_schema
// (items / item_schema); plan_sections carries those onto each llm_field_spec
// as ItemFields, and the prompt now asks the LLM for exactly those keys. This
// is the belt-and-braces second line: if the model still emits a different
// spelling (title/body where the template reads name/description), repair it
// before it reaches the template or is persisted. See plan_sections_action.go.
// ----------------------------------------------------------------------------

// itemKeySynonyms groups field names that mean the same thing across component
// templates. Matching is case- and separator-insensitive, so "Title" and
// "card_title" both match "title". Keep common spellings first; extend a group
// rather than scattering new pairs elsewhere.
var itemKeySynonyms = [][]string{
	{"title", "name", "heading", "header", "label", "headline"},
	{"description", "body", "text", "content", "desc", "detail", "details", "summary", "copy", "caption"},
	{"icon_svg", "icon", "image", "img", "svg"},
	{"url", "href", "link"},
	{"cta_text", "cta", "button_text", "button_label", "action_text"},
}

func synonymsFor(key string) []string {
	for _, group := range itemKeySynonyms {
		for _, k := range group {
			if k == key {
				return group
			}
		}
	}
	return nil
}

func normaliseKeyForMatch(s string) string {
	return strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(s))
}

// expectedItemFieldsFromComponentSchema reads the component's own input_schema
// and returns, per source:"llm" array field that declares an item shape, the
// field names each element must contain. The schema is the authoritative
// contract the html_template is built against and is reloaded fresh on every
// render, so reconciliation no longer depends on the section plan carrying
// item_fields or on the prompt. Scoped to source:"llm" so the reconciler's
// reach matches the writer loop; query-resolved/static arrays (already keyed
// correctly by the system) are left untouched. Reuses extractArrayItemFields
// (plan_sections_action.go) so item-field extraction stays identical to how the
// plan and prompt derive it. Empty — reconcile becomes a no-op — when the
// schema has no fields map or no llm array fields (e.g. render_component called
// outside the writer loop, or on a non-array component).
func expectedItemFieldsFromComponentSchema(inputSchema map[string]interface{}) map[string][]string {
	out := map[string][]string{}
	// Read through datahelpers.SchemaContentFields rather than inputSchema["fields"]
	// directly, so a legacy JSON-Schema component is reconciled too. Reading the
	// key directly made this a silent no-op on exactly the dialect whose items the
	// writer mis-keys (bugs_open/240) — expected came back empty, reconcile
	// returned at its len(expected)==0 guard, and the "unrecoverable" ERROR that
	// exists to make a missing item field visible never fired.
	fields, ok, _ := datahelpers.SchemaContentFields(inputSchema)
	if !ok {
		return out
	}
	for name, raw := range fields {
		fieldDef, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if src, _ := fieldDef["source"].(string); src != "llm" {
			continue
		}
		if itemFields := extractArrayItemFields(fieldDef); len(itemFields) > 0 {
			out[name] = itemFields
		}
	}
	return out
}

// reconcileGeneratedItemKeys repairs LLM array output whose per-item keys don't
// match the keys the component template reads. Per element: an expected key
// that is missing but present under a case/separator variant or a known synonym
// is moved onto the expected key (WARN, so disobedience is visible); an expected
// key still missing afterwards is logged at ERROR with the element's actual
// keys, so a silent empty card cannot pass unseen. Modifies content in place.
// Non-fatal by design.
func reconcileGeneratedItemKeys(content map[string]interface{}, expected map[string][]string, componentFn string, logger *zap.Logger) {
	if len(content) == 0 || len(expected) == 0 {
		return
	}
	for fieldName, wantFields := range expected {
		raw, present := content[fieldName]
		if !present {
			continue
		}
		items, ok := raw.([]interface{})
		if !ok {
			logger.Warn("reconcileGeneratedItemKeys: array field is not a list — skipping",
				zap.String("component", componentFn), zap.String("field", fieldName),
				zap.String("got_type", fmt.Sprintf("%T", raw)))
			continue
		}
		wantSet := make(map[string]bool, len(wantFields))
		for _, w := range wantFields {
			wantSet[w] = true
		}
		for idx, itemRaw := range items {
			item, ok := itemRaw.(map[string]interface{})
			if !ok {
				logger.Warn("reconcileGeneratedItemKeys: array element is not an object — skipping",
					zap.String("component", componentFn), zap.String("field", fieldName),
					zap.Int("index", idx), zap.String("got_type", fmt.Sprintf("%T", itemRaw)))
				continue
			}
			norm := make(map[string]string, len(item))
			for k := range item {
				norm[normaliseKeyForMatch(k)] = k
			}
			for _, want := range wantFields {
				if _, has := item[want]; has {
					continue
				}
				wantNorm := normaliseKeyForMatch(want)
				if actual, ok := norm[wantNorm]; ok && actual != want {
					item[want] = item[actual]
					delete(item, actual)
					norm[wantNorm] = want
					logger.Warn("reconcileGeneratedItemKeys: normalised LLM item key",
						zap.String("component", componentFn), zap.String("field", fieldName),
						zap.Int("index", idx), zap.String("from", actual), zap.String("to", want))
					continue
				}
				remapped := false
				for _, syn := range synonymsFor(want) {
					if syn == want || wantSet[syn] {
						continue
					}
					if actual, ok := norm[normaliseKeyForMatch(syn)]; ok {
						item[want] = item[actual]
						delete(item, actual)
						norm[wantNorm] = want
						logger.Warn("reconcileGeneratedItemKeys: remapped LLM item key",
							zap.String("component", componentFn), zap.String("field", fieldName),
							zap.Int("index", idx), zap.String("from", actual), zap.String("to", want))
						remapped = true
						break
					}
				}
				if !remapped {
					logger.Error("reconcileGeneratedItemKeys: expected item field missing and unrecoverable",
						zap.String("component", componentFn), zap.String("field", fieldName),
						zap.Int("index", idx), zap.String("expected_key", want),
						zap.Any("item_keys", datahelpers.GetMapKeys(item)))
				}
			}
		}
	}
}

// droppedPlanPageNames renders the dropped set for the action's summary log:
// "<name> [<url>] (pass <n>)" per page. A log line that says only how many pages
// were deleted cannot be acted on; one that names them can.
func droppedPlanPageNames(dropped []droppedPlanPage) []string {
	if len(dropped) == 0 {
		return nil
	}
	out := make([]string, 0, len(dropped))
	for _, d := range dropped {
		out = append(out, fmt.Sprintf("%s [%s] (pass %s)", d.Name, d.URL, d.Pass))
	}
	return out
}
