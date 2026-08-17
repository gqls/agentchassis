// FILE: platform/orchestration/actions/retract_asset_files_action.go
//
// RETRACT ASSET FILES — the asset-level sibling of `retract_page_deployment`.
//
// THE GAP THIS CLOSES. The platform can commit an asset file into a site repo
// (deploy_image_asset, derive_brand_head_assets, emit_sprite_css) but has no
// implemented way to remove one. That asymmetry produced a real population:
// bugs_closed/248's placeholder defect committed files under names no reader
// derives, and its fix stopped NEW instances while leaving the old files in
// place — an orphan file has NO assets row, so every row-keyed mechanism
// (page-retraction, the admin soft-delete, the drain) is structurally unable
// to name it. The admin API's DELETE handler is an `assets.status='deleted'`
// UPDATE that touches no file and returns `{"deleted": true}` — success with
// the artefact still serving, which is this estate's canonical trap. Live
// instance at introduction: gaswholesalers.com `/assets/images/logo.jpg`,
// owner-directed removal, no supported path (checked three times, three ways,
// 2026-08-16/17).
//
// Removal is effective for the same reason retraction is: gqls/sites'
// deploy-to-b2 workflow runs `b2 sync --delete` + a Cloudflare purge on every
// push, so deleting the repo file deletes it from the edge.
//
// ────────────────────────────────────────────────────────────────────────────
// THE GUARDS ARE THE DESIGN — and one of them is an inversion the reader
// must not miss. `retract_page_deployment`'s guard 2 says a caller names
// PAGES, never files, and no caller-supplied string can reach the adapter as
// a deletion. This action exists precisely for files the database cannot
// derive (an orphan has no row), so the caller DOES name paths. That widens
// the blast radius from "files a publish could have written" to "files the
// caller can spell", and everything below is what narrows it back:
//
//  1. SCOPE. Every path must sit under `/assets/` after cleaning, with no
//     `..` and no scheme. A page, a data feed, site chrome, `.github/` — all
//     structurally out of reach, whatever the caller spells.
//  2. NEVER a path the platform still believes in. A path that equals the
//     shared derivation (storage.DeployedAssetPath) of ANY non-deleted asset
//     row on the site is REFUSED — that file is somebody's artefact, and the
//     row, not the caller, is the authority. Note `superseded`/`retired` rows
//     refuse too: bugs_closed/248's drain measured that redeploying (and by
//     symmetry, deleting) a replaced asset's path is how a current artefact
//     gets clobbered.
//  3. NEVER a brand-head published path (favicon, og-card). Those files are
//     owned by derive_brand_head_assets under fixed names
//     (storage.BrandHeadAssetPaths) and may legitimately exist with no row.
//  4. NEVER a referenced file. Any deployed component on the site whose
//     rendered_html mentions the path refuses it — a 200 that something links
//     to is live content whatever the assets table thinks. This is the check
//     that saved the bucket-A drain twice (bugs_closed/248, the two logo
//     skips).
//  5. DRY RUN IS THE DEFAULT. The deleting branch is opt-in behind an
//     explicit config literal `"dry_run": false` — the owner ruling of
//     2026-08-02 §2: when a seam's widest branch is dangerous, the unsafe
//     side defaults OFF. A caller that says nothing gets the audit, not the
//     deletion. The literal is read from step config ONLY, never resolved
//     from collected_data, so no upstream step output can flip it.
//
// Paths are read from config — a literal `paths` array or a `paths_field`
// dotted reference — and deliberately NEVER through ExtractActionInputs'
// recursive search: a stray "paths" key in some step result must not become
// a deletion (the same argument recorded on the deploy_path refusal, and
// RFC_029's whole subject).
//
// Refusals are RETURNED and RECORDED (agent_error_log via agenterrors), never
// swallowed: pod logs retain ~90s, and a refusal that survives nowhere is a
// decision nobody can audit — the lesson the RFC_029 REVISE round taught this
// same estate about its own WARNs.
package actions

import (
	"context"
	"database/sql"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// RetractAssetFilesInputSpec — every key is a setting or an explicit
// reference; nothing here is hunted for by the recursive search (see the
// header's paths note; CheckConfig opts the action into unknown-key
// detection).
var RetractAssetFilesInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Optional:    []string{"site_id", "site_id_field", "domain_field", "repo_name"},
	ConfigKeys:  []string{"paths", "paths_field", "dry_run"},
}

func init() {
	datahelpers.RegisterActionInputSpec("retract_asset_files", RetractAssetFilesInputSpec)
}

// assetRetractionRefusal is one requested path this action declined to
// delete, with the reason. Returned in the result AND written to
// agent_error_log — same double-record shape as the page sibling's
// retractionCandidate.
type assetRetractionRefusal struct {
	Path    string `json:"path"`
	Refused string `json:"refused"`
}

// RetractAssetFilesAction deletes named orphan asset files from a site's
// repo via git-adapter delete_file, after the guard chain in the file header.
func RetractAssetFilesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "retract_asset_files"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("retract_asset_files requires a database handle")
	}

	siteID, err := resolveRetractionSiteID(params, config)
	if err != nil {
		return nil, err
	}
	var domain string
	if err := params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain); err != nil {
		return nil, fmt.Errorf("resolve domain for site %s: %w", siteID, err)
	}
	if strings.TrimSpace(domain) == "" {
		return nil, fmt.Errorf("site %s has no domain; an asset retraction cannot be scoped", siteID)
	}

	requested, err := requestedAssetPaths(config, params.CollectedData)
	if err != nil {
		return nil, err
	}
	if len(requested) == 0 {
		return nil, fmt.Errorf("retract_asset_files: no paths requested — this action never selects targets itself; name them via config 'paths' or 'paths_field'")
	}

	// Guard 2's population: the derived path of every non-deleted asset row.
	owned, err := loadOwnedAssetWebPaths(ctx, params.DB, siteID)
	if err != nil {
		return nil, fmt.Errorf("load owned asset paths: %w", err)
	}
	brandHead := map[string]bool{}
	for _, published := range storage.BrandHeadAssetPaths {
		brandHead[published] = true
	}

	var refusals []assetRetractionRefusal
	var paths []string
	for _, web := range requested {
		if reason := assetPathScopeViolation(web); reason != "" {
			refusals = append(refusals, assetRetractionRefusal{Path: web, Refused: reason})
			continue
		}
		if owner, ok := owned[web]; ok {
			refusals = append(refusals, assetRetractionRefusal{Path: web,
				Refused: fmt.Sprintf("the platform still owns this path: non-deleted asset row %s derives it — retire the row first if this is intended", owner)})
			continue
		}
		if brandHead[web] {
			refusals = append(refusals, assetRetractionRefusal{Path: web,
				Refused: "brand-head published path — owned by derive_brand_head_assets under a fixed name; re-derive, never delete"})
			continue
		}
		referrers, err := deployedReferrersOf(ctx, params.DB, siteID, web)
		if err != nil {
			return nil, fmt.Errorf("scan deployed references of %s: %w", web, err)
		}
		if len(referrers) > 0 {
			refusals = append(refusals, assetRetractionRefusal{Path: web,
				Refused: fmt.Sprintf("referenced by deployed content (%s) — a linked 200 is live whatever the assets table says", strings.Join(referrers, ", "))})
			continue
		}
		// The adapter wants repo-relative paths (it prefixes the domain
		// itself, same as the commit verb); the web path is site-absolute.
		paths = append(paths, strings.TrimPrefix(web, "/"))
	}
	sort.Strings(paths)

	// dry_run: STEP CONFIG ONLY, and absence means TRUE. The zero value of a
	// missing bool is the dangerous direction here, so the read is explicit:
	// only a literal false in the step's own config arms the deletion.
	dryRun := true
	if v, ok := config["dry_run"].(bool); ok {
		dryRun = v
	}

	result := map[string]interface{}{
		"site_id":   siteID.String(),
		"domain":    domain,
		"requested": len(requested),
		"refusals":  refusals,
		"paths":     paths,
		"retracted": len(paths),
		"dry_run":   dryRun,
	}

	if dryRun {
		result["await_response"] = false
		logger.Info("retract_asset_files: dry run (the default — set config dry_run:false to delete)",
			zap.String("domain", domain), zap.Strings("paths", paths),
			zap.Int("refused", len(refusals)))
		return result, nil
	}

	// Durable record BEFORE the dispatch, refusals included even when the
	// surviving batch is empty — a run that deletes nothing because every
	// path was refused is the loudest of the silent cases.
	recordAssetRetraction(ctx, params, siteID, domain, requested, paths, refusals, logger)

	if len(paths) == 0 {
		result["await_response"] = false
		result["reason"] = "every requested path was refused; nothing dispatched"
		return result, nil
	}

	requestID, err := sendGitDeleteFileRequest(ctx, params, config, domain, paths,
		fmt.Sprintf("Retract %d orphan asset file(s) from %s (retract_asset_files)", len(paths), domain), logger)
	if err != nil {
		return nil, err
	}
	result["request_id"] = requestID
	result["await_response"] = true
	return result, nil
}

// requestedAssetPaths reads the caller's path list from step config only: a
// literal []interface{} under "paths", or a dotted "paths_field" reference.
// No recursive search, no single-string convenience form — a deletion list
// is the one input worth making the caller spell in full.
func requestedAssetPaths(config, collected map[string]interface{}) ([]string, error) {
	var raw []interface{}
	if lit, ok := config["paths"].([]interface{}); ok {
		raw = lit
	} else if ref, ok := config["paths_field"].(string); ok && ref != "" {
		if v, ok := datahelpers.ExtractNestedField(collected, ref).([]interface{}); ok {
			raw = v
		}
	}
	out := make([]string, 0, len(raw))
	for _, r := range raw {
		s, ok := r.(string)
		if !ok || strings.TrimSpace(s) == "" {
			return nil, fmt.Errorf("retract_asset_files: paths must be non-empty strings, got %T %q", r, r)
		}
		out = append(out, strings.TrimSpace(s))
	}
	return out, nil
}

// assetPathScopeViolation returns the refusal reason for a path outside this
// action's scope, or "" if the path is admissible. path.Clean collapses any
// `..` and duplicate slashes FIRST, so the prefix check runs on what the
// filesystem would actually resolve — checking the raw string instead is the
// classic traversal mistake.
func assetPathScopeViolation(web string) string {
	if !strings.HasPrefix(web, "/") {
		return "not a site-absolute path — spell it as served, e.g. /assets/images/x.jpg"
	}
	if strings.Contains(web, "://") || strings.Contains(web, "?") || strings.Contains(web, "#") {
		return "not a plain file path"
	}
	cleaned := path.Clean(web)
	if cleaned != web {
		return fmt.Sprintf("path does not survive cleaning (%q -> %q) — refusing rather than sanitising, the sibling's rule", web, cleaned)
	}
	if !strings.HasPrefix(cleaned, "/assets/") || cleaned == "/assets" || strings.HasSuffix(cleaned, "/") {
		return "outside /assets/ — pages, feeds and chrome are page-retraction's or nobody's, never this action's"
	}
	return ""
}

// loadOwnedAssetWebPaths derives the served path of every non-deleted asset
// row on the site, through the one shared derivation every reader uses. The
// map value is the row's identity, so a refusal can name its owner.
func loadOwnedAssetWebPaths(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, COALESCE(asset_key, ''), COALESCE(purpose, '')
		FROM assets
		WHERE site_id = $1 AND COALESCE(status, '') <> 'deleted'
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var id, assetKey, purpose string
		if err := rows.Scan(&id, &assetKey, &purpose); err != nil {
			return nil, err
		}
		if purpose == "" {
			continue
		}
		out[storage.DeployedWebPath(assetKey, purpose)] = fmt.Sprintf("%s (purpose %s, status<>deleted)", id, purpose)
	}
	return out, rows.Err()
}

// deployedReferrersOf names the deployed components on this site whose
// rendered_html mentions the path. A LIKE scan, not a parse: an href, a CSS
// url(), an inline style and a srcset all contain the literal path, and a
// false positive here costs a refusal the operator can read — the cheap
// direction to be wrong in.
func deployedReferrersOf(ctx context.Context, db *sql.DB, siteID uuid.UUID, web string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT p.name
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND pc.build_status = 'deployed'
		  AND pc.rendered_html LIKE '%' || $2 || '%'
		ORDER BY p.name
		LIMIT 8
	`, siteID, web)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// recordAssetRetraction writes the durable audit: one warning row per
// refusal, one info row for the run. Best-effort by design (agenterrors
// swallows its own failures into a bool) — a lost audit row must not fail a
// retraction that is otherwise sound, but the attempt is never skipped.
func recordAssetRetraction(ctx context.Context, params ActionParams, siteID uuid.UUID, domain string, requested, paths []string, refusals []assetRetractionRefusal, logger *zap.Logger) {
	for _, r := range refusals {
		agenterrors.Write(ctx, params.DB, logger, agenterrors.Entry{
			SiteID:       siteID.String(),
			Domain:       domain,
			AgentType:    params.ExecutionContext.Sender.AgentType,
			StepName:     params.ExecutionContext.StepName,
			Action:       "retract_asset_files",
			ErrorCode:    "ASSET_RETRACTION_REFUSED",
			Severity:     "warning",
			ErrorMessage: fmt.Sprintf("refused to delete %s: %s", r.Path, r.Refused),
			Context:      map[string]interface{}{"path": r.Path, "refused": r.Refused},
		})
	}
	agenterrors.Write(ctx, params.DB, logger, agenterrors.Entry{
		SiteID:       siteID.String(),
		Domain:       domain,
		AgentType:    params.ExecutionContext.Sender.AgentType,
		StepName:     params.ExecutionContext.StepName,
		Action:       "retract_asset_files",
		ErrorCode:    "ASSET_RETRACTION_AUDIT",
		Severity:     "info",
		ErrorMessage: fmt.Sprintf("asset retraction on %s: %d requested, %d refused, %d dispatched", domain, len(requested), len(refusals), len(paths)),
		Context: map[string]interface{}{
			"requested": requested,
			"paths":     paths,
			"refused":   len(refusals),
			"at":        time.Now().UTC().Format(time.RFC3339),
		},
	})
}
