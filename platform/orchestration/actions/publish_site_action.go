// FILE: platform/orchestration/actions/publish_site_action.go
//
// publish_site: one reconciliation pass of the publish seam (platform/publish)
// for one site. Reads the site's opt-in (sites.publish_target — NULL means
// OFF, migration 412), computes the built tree's hash from the portfolio-sites
// listing, and publishes only on drift from sites.published_hash. Acceptance
// is served bytes fetched back from the hosted copy, never a provider API
// status; published_hash is written only after acceptance, so a failed or
// unaccepted publish leaves the drift standing and the next reconciler tick
// retries it.
//
// MUST run in a spawned storage-enabled pod (site-publisher is on the
// spawner's storage allow-list): the standing chassis deliberately carries no
// B2 credentials (owner ruling 2026-08-08), so this action constructs its own
// client bound to the portfolio-sites bucket rather than using
// params.StorageClient, which binds IMAGE_BUCKET and is nil on the chassis.
package actions

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	cfgpkg "github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/publish"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

var PublishSiteInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"domain"},
	// allow_bulk_unpublish lifts b2worker's bulk floor (a deletion sweep
	// removing >20 keys AND >50% of the destination refuses without it).
	// Default OFF; the scheduled reconciler dispatch never passes it, so a
	// mass restructure needs a hand dispatch that does — deliberately.
	Optional: []string{"site_id", "force", "source_bucket", "allow_bulk_unpublish"},
	Defaults: map[string]interface{}{
		"source_bucket": "portfolio-sites",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("publish_site", PublishSiteInputSpec)
}

// newPortfolioStore is swapped by tests; production constructs a second
// S3Client because storage.Client binds exactly one bucket per instance.
var newPortfolioStore = func(ctx context.Context, bucket string, logger *zap.Logger) (publish.ObjectStore, error) {
	return storage.NewS3Client(ctx, cfgpkg.ObjectStorageConfig{
		Provider:        "s3",
		Bucket:          bucket,
		AccessKeyEnvVar: "B2_APPLICATION_KEY_ID",
		SecretKeyEnvVar: "B2_APPLICATION_KEY",
	}, *logger)
}

// servedFetch is swapped by tests. The query string is a cache-buster: the
// worker ignores it when building the object key, but the CDN cache keys on
// it, so the fetch reaches origin instead of a stale edge copy.
var servedFetch = func(ctx context.Context, url string) ([]byte, int, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
	return body, resp.StatusCode, err
}

func PublishSiteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "publish_site"),
		zap.String("step_name", params.ExecutionContext.StepName))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("publish_site: database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(params.CollectedData, params.StepConfig.Config, PublishSiteInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("publish_site: extract inputs: %w", err)
	}
	domain := strings.TrimSpace(inputs.Get("domain"))
	if domain == "" {
		return map[string]interface{}{
			"published": false, "skipped": true, "reason": "no domain resolvable from inputs",
		}, nil
	}

	site, err := loadPublishSiteRow(ctx, params.DB, domain, inputs.Get("site_id"))
	if err != nil {
		return nil, err
	}
	if site == nil {
		return map[string]interface{}{
			"published": false, "skipped": true, "reason": "no sites row for domain " + domain,
		}, nil
	}
	if site.target == "" {
		// The seam is opt-in, default OFF (migration 412). Not an error: the
		// workflow completes and nothing retries.
		return map[string]interface{}{
			"published": false, "skipped": true,
			"reason": "publish_target not set for " + domain + " (seam is opt-in, default OFF)",
		}, nil
	}

	store, err := newPortfolioStore(ctx, inputs.Get("source_bucket"), logger)
	if err != nil {
		// No credentials means this ran in the wrong pod (the chassis carries
		// none by ruling) — a real failure, not a skip: it must be seen.
		return nil, fmt.Errorf("publish_site: portfolio store unavailable (is this a storage-enabled spawned pod?): %w", err)
	}

	source := publish.NewS3Source(store, domain)
	files, err := source.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("publish_site: %w", err)
	}
	if len(files) == 0 {
		reason := fmt.Sprintf("no built artefacts under %s/%s/ — nothing to publish", inputs.Get("source_bucket"), domain)
		if site.publishedHash != "" {
			// An empty origin under a site that HAS a hosted copy means the
			// mirror is now orphaned whole. Unpublishing it off the back of
			// an empty listing is a delete-all keyed on an absence — refused
			// implicitly here and explicitly by the backend; that verb is
			// bugs_open/304's decision to make (bugs_open/429 for the sweep).
			reason = fmt.Sprintf("origin tree under %s/%s/ is EMPTY but a hosted copy is still standing (published_hash %s) — refusing to unpublish implicitly; see bugs_open/304 and bugs_open/429", inputs.Get("source_bucket"), domain, site.publishedHash)
		}
		return map[string]interface{}{
			"published": false, "skipped": true,
			"reason": reason,
		}, nil
	}

	treeHash := publish.TreeHash(files)
	if treeHash == site.publishedHash && !inputs.GetBool("force", false) {
		return map[string]interface{}{
			"published": false, "skipped": true, "reason": "no drift", "tree_hash": treeHash,
		}, nil
	}

	backend, err := publish.For(site.target, publish.Deps{Store: store})
	if err != nil {
		return nil, fmt.Errorf("publish_site: %w", err)
	}

	logger.Info("publishing on drift",
		zap.String("domain", domain), zap.String("backend", backend.Name()),
		zap.String("tree_hash", treeHash), zap.String("published_hash", site.publishedHash),
		zap.Int("files", len(files)))

	res, err := backend.Publish(ctx, publish.Request{
		Domain: domain, Project: site.project, Files: files, Source: source,
		AllowBulkUnpublish: inputs.GetBool("allow_bulk_unpublish", false),
	})
	if err != nil {
		return nil, fmt.Errorf("publish_site: backend %s: %w", backend.Name(), err)
	}

	// Acceptance: fetch a page back from the hosted copy and compare bytes
	// against the origin store. robots.txt is excluded estate-wide (the edge
	// rewrites it — LANDMINES); index.html is the acceptance page. When the
	// backend swept orphans, acceptance is a PAIR: a swept key must 404 and a
	// kept page must still 200 (deletionAcceptance below).
	if accepted, reason := servedAcceptance(ctx, source, files, res, treeHash); !accepted {
		// Deliberately NOT written back: the drift stands and the next tick
		// retries. The result records what happened; the error would only
		// mask it behind a retry loop.
		return map[string]interface{}{
			"published": true, "accepted": false, "reason": reason,
			"tree_hash": treeHash, "files": len(files), "url": res.URL,
		}, nil
	}
	if accepted, reason := deletionAcceptance(ctx, files, res, treeHash); !accepted {
		// Same contract as above: hash not written, drift stands, next tick
		// retries — the sweep is idempotent, so the retry re-verifies rather
		// than re-deletes.
		return map[string]interface{}{
			"published": true, "accepted": false, "reason": reason,
			"tree_hash": treeHash, "files": len(files), "url": res.URL,
			"deleted": res.Deleted,
		}, nil
	}

	if _, err := params.DB.ExecContext(ctx,
		`UPDATE sites SET published_hash = $1, published_at = NOW(),
		        publish_project = COALESCE(NULLIF($2, ''), publish_project),
		        updated_at = NOW()
		  WHERE id = $3`,
		treeHash, res.Project, site.id); err != nil {
		return nil, fmt.Errorf("publish_site: record published_hash: %w", err)
	}

	// Cap the recorded key list: action results land in collected_data, and
	// an unbounded list from a bulk sweep would bloat orchestration rows.
	recordedKeys := res.DeletedKeys
	if len(recordedKeys) > 20 {
		recordedKeys = recordedKeys[:20]
	}
	return map[string]interface{}{
		"published": true, "accepted": true,
		"tree_hash": treeHash, "files": len(files), "url": res.URL, "backend": backend.Name(),
		"deleted": res.Deleted, "deleted_keys": recordedKeys,
	}, nil
}

type publishSiteRow struct {
	id            string
	target        string
	project       string
	publishedHash string
}

func loadPublishSiteRow(ctx context.Context, db *sql.DB, domain, siteID string) (*publishSiteRow, error) {
	row := &publishSiteRow{}
	query := `SELECT id::text, COALESCE(publish_target,''), COALESCE(publish_project,''), COALESCE(published_hash,'')
	            FROM sites WHERE domain = $1`
	arg := domain
	if siteID != "" {
		query = `SELECT id::text, COALESCE(publish_target,''), COALESCE(publish_project,''), COALESCE(published_hash,'')
		           FROM sites WHERE id = $1::uuid`
		arg = siteID
	}
	err := db.QueryRowContext(ctx, query, arg).Scan(&row.id, &row.target, &row.project, &row.publishedHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("publish_site: load sites row: %w", err)
	}
	return row, nil
}

// deletionAcceptance is the 404 half of served acceptance (bugs_open/429):
// when the publish swept orphans, one swept key must no longer serve
// (under-deletion) and one kept page must still serve (over-deletion) — the
// planted-control pair. Status codes only: byte-compares against a fetched
// body walk into the CDN-adds-bytes landmine; a status cannot. robots.txt is
// excluded from the probe set — the edge rewrites it to a 200 regardless of
// the object, so probing it would fail acceptance on every tick for ever.
// The swept key is probed in its literal form, never a directory form (the
// worker rewrites "/x/" to "/x/index.html", DGH-012).
func deletionAcceptance(ctx context.Context, files []publish.File, res publish.Result, treeHash string) (bool, string) {
	if res.Deleted == 0 || res.URL == "" {
		return true, ""
	}
	probe := ""
	for _, k := range res.DeletedKeys {
		if k == "robots.txt" {
			continue
		}
		if probe == "" {
			probe = k
		}
		if strings.HasSuffix(k, ".html") {
			probe = k
			break
		}
	}
	if probe == "" {
		return true, "" // only robots.txt was swept — nothing probe-worthy
	}

	cacheBust := treeHash
	if len(cacheBust) > 12 {
		cacheBust = cacheBust[len(cacheBust)-12:]
	}
	_, status, err := servedFetch(ctx, res.URL+probe+"?pub="+cacheBust)
	if err != nil {
		return false, "deletion acceptance: fetch swept key " + probe + ": " + err.Error()
	}
	if status != http.StatusNotFound {
		return false, fmt.Sprintf("deletion acceptance: swept key %s still serves %d, want 404 — the unpublish did not reach the hosted copy", probe, status)
	}

	// Kept-200 control: a swept-key 404 alone is blind to over-deletion.
	// When the tree has index.html, servedAcceptance already byte-verified a
	// kept page; this covers the tree that does not.
	kept := ""
	for _, f := range files {
		if f.Key == "index.html" {
			kept = ""
			break
		}
		if kept == "" && strings.HasSuffix(f.Key, ".html") {
			kept = f.Key
		}
	}
	if kept != "" {
		_, keptStatus, err := servedFetch(ctx, res.URL+kept+"?pub="+cacheBust)
		if err != nil {
			return false, "deletion acceptance: fetch kept key " + kept + ": " + err.Error()
		}
		if keptStatus != http.StatusOK {
			return false, fmt.Sprintf("deletion acceptance: kept key %s serves %d, want 200 — the sweep may have over-deleted", kept, keptStatus)
		}
	}
	return true, ""
}

// servedAcceptance compares the hosted copy's index.html bytes against the
// origin store's. A tree with no index.html accepts on the backend's own
// verification alone (recorded in the reason path only on failure).
func servedAcceptance(ctx context.Context, source publish.Source, files []publish.File, res publish.Result, treeHash string) (bool, string) {
	hasIndex := false
	for _, f := range files {
		if f.Key == "index.html" {
			hasIndex = true
			break
		}
	}
	if !hasIndex || res.URL == "" {
		return true, ""
	}

	rc, err := source.Open(ctx, "index.html")
	if err != nil {
		return false, "acceptance: origin open index.html: " + err.Error()
	}
	defer rc.Close()
	originBytes, err := io.ReadAll(rc)
	if err != nil {
		return false, "acceptance: origin read index.html: " + err.Error()
	}

	cacheBust := treeHash
	if len(cacheBust) > 12 {
		cacheBust = cacheBust[len(cacheBust)-12:]
	}
	servedBytes, status, err := servedFetch(ctx, res.URL+"index.html?pub="+cacheBust)
	if err != nil {
		return false, "acceptance: served fetch: " + err.Error()
	}
	if status != http.StatusOK {
		return false, fmt.Sprintf("acceptance: served index.html returned %d", status)
	}
	o, s := sha256.Sum256(originBytes), sha256.Sum256(servedBytes)
	if o != s {
		return false, fmt.Sprintf("acceptance: served index.html sha256 %s != origin %s",
			hex.EncodeToString(s[:8]), hex.EncodeToString(o[:8]))
	}
	return true, ""
}
