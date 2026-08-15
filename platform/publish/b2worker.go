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

// Publish copies every file to "<project>/<key>" and then verifies the copy
// at the destination listing: per-key ETag equality against the source. Both
// sides are single-part uploads in the same bucket, so equal bytes mean equal
// ETags — a mismatch is a real failure, not hash-scheme noise. Served-bytes
// acceptance stays with the caller; this verification is origin-side only.
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

	return Result{
		Project:   project,
		Published: len(req.Files),
		URL:       "https://" + project + "/",
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
