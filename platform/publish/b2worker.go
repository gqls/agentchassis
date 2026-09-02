// FILE: platform/publish/b2worker.go
//
// The in-estate backend: copies the built tree to a second hostname prefix in
// the same bucket, so the existing *.ugg2.com Cloudflare worker serves it
// (object key = hostname + path — scripts/cloudflare/worker.js). Zero new
// infrastructure and zero new credentials: the worker already resolves any
// hostname prefix it is routed for.
package publish

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"sort"
	"strings"
)

// woff2 and a few friends are missing from Go's built-in mime table, and the
// debian-slim images carry no /etc/mime.types to fill the gap.
var extraContentTypes = map[string]string{
	".woff2": "font/woff2",
	".woff":  "font/woff",
	".ico":   "image/x-icon",
	".txt":   "text/plain; charset=utf-8",
}

func contentTypeFor(key string) string {
	ext := strings.ToLower(filepath.Ext(key))
	if ct, ok := extraContentTypes[ext]; ok {
		return ct
	}
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

type B2Worker struct {
	store ObjectStore
}

func NewB2Worker(store ObjectStore) *B2Worker {
	return &B2Worker{store: store}
}

func (w *B2Worker) Name() string { return TargetB2Worker }

// Publish copies every file to "<project>/<key>", verifies the copy at the
// destination listing (per-key ETag equality against the source — both sides
// are single-part uploads in the same bucket, so equal bytes mean equal
// ETags and a mismatch is a real failure, not hash-scheme noise), and then
// CONVERGES: destination keys whose source key is gone are deleted and
// verified gone (bugs_open/429 — a mirror that only copies can never
// unpublish, so a retracted page's object serves at the slug for ever).
// Served-bytes acceptance stays with the caller; verification here is
// origin-side only.
func (w *B2Worker) Publish(ctx context.Context, req Request) (Result, error) {
	project := strings.TrimSpace(req.Project)
	if project == "" {
		return Result{}, fmt.Errorf("b2worker: publish_project is empty — set it to the serving hostname (e.g. \"canary.ugg2.com\") before opting the site in")
	}
	if strings.Contains(project, "/") {
		return Result{}, fmt.Errorf("b2worker: publish_project %q must be a bare hostname, not a path", project)
	}
	if project == req.Domain {
		return Result{}, fmt.Errorf("b2worker: publish_project equals the site domain %q — that would copy the tree onto itself", project)
	}
	if len(req.Files) == 0 {
		// Per-backend defence in depth (publish_site already skips empty
		// trees before reaching any backend; cfpages does not inherit this):
		// with the deletion sweep below, an empty source would read as
		// "delete the whole mirror" — a destructive verb keyed on an absence
		// (bugs_open/304), which must be its own explicit decision, never a
		// side effect of a listing that came back empty.
		return Result{}, fmt.Errorf("b2worker: refusing an empty file set for %q — an empty source cannot license sweeping the mirror (bugs_open/304, bugs_open/429)", req.Domain)
	}

	for _, f := range req.Files {
		if err := w.copyOne(ctx, req, project, f); err != nil {
			return Result{}, err
		}
	}

	// Verify at the destination listing, not at the upload return values.
	destObjs, err := w.store.ListObjects(ctx, project+"/")
	if err != nil {
		return Result{}, fmt.Errorf("b2worker: verify list %q: %w", project+"/", err)
	}
	destETags := make(map[string]string, len(destObjs))
	for _, o := range destObjs {
		if len(o.Key) > len(project)+1 {
			destETags[o.Key[len(project)+1:]] = o.ETag
		}
	}
	var bad []string
	for _, f := range req.Files {
		if got, ok := destETags[f.Key]; !ok {
			bad = append(bad, f.Key+" (missing)")
		} else if got != f.ETag {
			bad = append(bad, f.Key+" (etag mismatch)")
		}
		if len(bad) >= 5 {
			break
		}
	}
	if len(bad) > 0 {
		return Result{}, fmt.Errorf("b2worker: copy verification failed for %d+ file(s): %s", len(bad), strings.Join(bad, ", "))
	}

	// The deletion sweep: any destination key whose source key is gone is an
	// orphan and gets removed, reusing the verify listing above. A concurrent
	// publish of the same site could in principle have this sweep remove the
	// other run's fresh upload — that run's own verification then fails, its
	// drift stands, and the next tick converges; overlap needs a manual force
	// anyway (the reconciler runs one site per tick, max_concurrent=1).
	sourceKeys := make(map[string]bool, len(req.Files))
	for _, f := range req.Files {
		sourceKeys[f.Key] = true
	}
	var orphans []string
	for _, o := range destObjs {
		if len(o.Key) <= len(project)+1 {
			continue
		}
		if rel := o.Key[len(project)+1:]; !sourceKeys[rel] {
			orphans = append(orphans, rel)
		}
	}
	sort.Strings(orphans)

	// Bulk floor: a sweep removing most of the destination in one pass is far
	// more likely a lying source listing or a whole-site teardown than a
	// retraction (routine sweeps are a few keys), so it needs the explicit
	// opt-in. If a truncated listing DID slip under the floor — which takes
	// an SDK-level fault, since ListObjects paginates to exhaustion and a
	// mid-page error aborts the whole listing — the source is authoritative
	// and untouched: the truncated tree's hash IS the drift, and the next
	// full-listing tick republishes the wrongly-swept copies. Bounded
	// staleness, never data loss. Missed deletes retry the same way.
	if len(orphans) > 20 && len(orphans)*2 > len(destObjs) && !req.AllowBulkUnpublish {
		return Result{}, fmt.Errorf("b2worker: deletion sweep would remove %d of %d destination keys under %q — refusing without allow_bulk_unpublish (a routine retraction sweeps a few keys; this shape is a partial source listing or a mass teardown)", len(orphans), len(destObjs), project+"/")
	}

	for _, rel := range orphans {
		if err := w.store.Delete(ctx, project+"/"+rel); err != nil {
			return Result{}, fmt.Errorf("b2worker: delete orphan %s/%s: %w", project, rel, err)
		}
	}

	// Verify the sweep at a fresh destination listing, mirroring the copy
	// half — a Delete return value is not evidence of absence.
	if len(orphans) > 0 {
		after, err := w.store.ListObjects(ctx, project+"/")
		if err != nil {
			return Result{}, fmt.Errorf("b2worker: post-sweep list %q: %w", project+"/", err)
		}
		remaining := make(map[string]bool, len(after))
		for _, o := range after {
			remaining[o.Key] = true
		}
		var still []string
		for _, rel := range orphans {
			if remaining[project+"/"+rel] {
				still = append(still, rel)
				if len(still) >= 5 {
					break
				}
			}
		}
		if len(still) > 0 {
			return Result{}, fmt.Errorf("b2worker: sweep verification failed — %d+ orphan(s) still present after delete: %s", len(still), strings.Join(still, ", "))
		}
	}

	return Result{
		Project:     project,
		Published:   len(req.Files),
		Deleted:     len(orphans),
		DeletedKeys: orphans,
		URL:         "https://" + project + "/",
	}, nil
}

func (w *B2Worker) copyOne(ctx context.Context, req Request, project string, f File) error {
	rc, err := req.Source.Open(ctx, f.Key)
	if err != nil {
		return fmt.Errorf("b2worker: open %s/%s: %w", req.Domain, f.Key, err)
	}
	defer rc.Close()
	// Buffer before upload: B2's S3 gateway REQUIRES Content-Length (HTTP 411
	// MissingContentLength on a bare stream — hit live on the first canary,
	// 2026-08-15), and the SDK can only compute it from a seekable body.
	// Site files are small; the ZIP deliverable (Phase 3) must NOT copy this
	// pattern for its own output.
	body, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("b2worker: read %s/%s: %w", req.Domain, f.Key, err)
	}
	if _, err := w.store.Upload(ctx, project+"/"+f.Key, contentTypeFor(f.Key), bytes.NewReader(body)); err != nil {
		return fmt.Errorf("b2worker: upload %s/%s: %w", project, f.Key, err)
	}
	return nil
}
