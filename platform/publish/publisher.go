// FILE: platform/publish/publisher.go
//
// The publish seam: mirrors a site's built artefact tree (the objects under
// b2://portfolio-sites/<domain>/) to a hosted copy. Provider-agnostic;
// sites.publish_target names the backend, and NULL means the seam is OFF for
// that site (migration 412, the 2026-08-02 opt-in ruling). The seam never
// decides WHEN to publish — the reconciler does that on tree-hash drift —
// and acceptance is served bytes, never a provider API 200.
package publish

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sort"

	"github.com/gqls/agentchassis/platform/storage"
)

// Registered publish_target values. sites.publish_target must hold one of
// these; anything else is a configuration error the action reports.
const (
	// TargetB2Worker copies the tree under a second hostname prefix in the
	// same portfolio-sites bucket, served by the existing *.ugg2.com worker.
	// publish_project is the serving hostname, e.g. "canary.ugg2.com".
	TargetB2Worker = "b2worker"
	// TargetCFPages deploys to Cloudflare Pages via Direct Upload.
	// Not yet armed — see cfpages.go.
	TargetCFPages = "cfpages"
)

// File is one object in the built tree, keyed RELATIVE to the site prefix
// ("index.html", not "example.com/index.html").
type File struct {
	Key  string
	ETag string
	Size int64
}

// ObjectStore is the slice of storage.Client the seam needs. *storage.S3Client
// satisfies it. Kept narrow so tests can fake it without S3. Delete is part of
// the contract, not an optional interface: a backend that can copy but not
// delete is exactly the mirror-cannot-unpublish defect (bugs_open/429), and an
// optional method would let a store degrade back into it silently.
type ObjectStore interface {
	ListObjects(ctx context.Context, prefix string) ([]storage.ObjectInfo, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Upload(ctx context.Context, key, contentType string, body io.Reader) (string, error)
	Delete(ctx context.Context, key string) error
}

// Source reads a site's built tree.
type Source interface {
	// List returns every file in the tree, keys relative to the site prefix.
	List(ctx context.Context) ([]File, error)
	// Open returns the bytes of one file by its relative key.
	Open(ctx context.Context, key string) (io.ReadCloser, error)
}

// Request is one publish of one site's current tree.
type Request struct {
	Domain string
	// Project is sites.publish_project — the provider-side identifier. Its
	// meaning is backend-specific (b2worker: the serving hostname; cfpages:
	// the Pages project name).
	Project string
	Files   []File
	Source  Source
	// AllowBulkUnpublish permits a deletion sweep that would remove a large
	// share of the destination in one pass (b2worker's bulk floor). Default
	// OFF per the 2026-08-02 ruling: new authority on a shared seam ships as
	// an opt-in field. The scheduled reconciler's dispatch never sets it —
	// deliberately — so a mass restructure needs a hand dispatch that does.
	AllowBulkUnpublish bool
}

// Result reports what a backend did. URL is the served base the caller's
// acceptance check fetches from — acceptance belongs to the caller, not the
// backend, so a backend must never report success it cannot serve. Deleted /
// DeletedKeys report the deletion sweep (bugs_open/429): the caller needs the
// keys for the served-404 half of acceptance.
type Result struct {
	Project     string
	Published   int
	Deleted     int
	DeletedKeys []string
	URL         string
}

// Publisher is one hosting backend behind the seam.
type Publisher interface {
	Name() string
	Publish(ctx context.Context, req Request) (Result, error)
}

// Deps carries what backends need. Only the fields a chosen backend uses are
// required; For validates per backend.
type Deps struct {
	Store ObjectStore
}

// For returns the backend named by a sites.publish_target value. The caller
// has already handled NULL (seam OFF) — an empty or unknown target here is a
// configuration error, not a skip.
func For(target string, d Deps) (Publisher, error) {
	switch target {
	case TargetB2Worker:
		if d.Store == nil {
			return nil, fmt.Errorf("publish: %s backend needs an object store", TargetB2Worker)
		}
		return NewB2Worker(d.Store), nil
	case TargetCFPages:
		return NewCFPages(), nil
	case "":
		return nil, fmt.Errorf("publish: empty publish_target reached the seam — NULL handling belongs to the caller")
	default:
		return nil, fmt.Errorf("publish: unknown publish_target %q (known: %s, %s)", target, TargetB2Worker, TargetCFPages)
	}
}

// TreeHash is the drift detector: a stable digest of the whole built tree,
// order-independent, sensitive to any file appearing, disappearing or
// changing. ETags from B2's S3 gateway are content-MD5 for single-part
// uploads, which everything in these trees is; size rides along as a belt.
// The prefix names the algorithm AND the publish semantics, so a change to
// either republishes once, explicably, instead of silently. th1→th2
// (2026-09-02, bugs_open/429) changed no algorithm: it marks the deletion
// half arriving — "published" now asserts the destination CONVERGES on the
// source, including removals, so every stored th1 hash owes exactly one
// converging republish (which is what sweeps the pre-fix orphans).
func TreeHash(files []File) string {
	lines := make([]string, 0, len(files))
	for _, f := range files {
		lines = append(lines, fmt.Sprintf("%s\x00%s\x00%d", f.Key, f.ETag, f.Size))
	}
	sort.Strings(lines)
	h := sha256.New()
	for _, l := range lines {
		h.Write([]byte(l))
		h.Write([]byte{'\n'})
	}
	return "th2:" + hex.EncodeToString(h.Sum(nil))
}

// S3Source reads a site tree from an ObjectStore under "<domain>/".
type S3Source struct {
	store  ObjectStore
	domain string
}

func NewS3Source(store ObjectStore, domain string) *S3Source {
	return &S3Source{store: store, domain: domain}
}

func (s *S3Source) List(ctx context.Context) ([]File, error) {
	prefix := s.domain + "/"
	objs, err := s.store.ListObjects(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("publish: list %q: %w", prefix, err)
	}
	files := make([]File, 0, len(objs))
	for _, o := range objs {
		if len(o.Key) <= len(prefix) {
			continue // the prefix marker object itself, if any
		}
		files = append(files, File{Key: o.Key[len(prefix):], ETag: o.ETag, Size: o.Size})
	}
	return files, nil
}

func (s *S3Source) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.store.Download(ctx, s.domain+"/"+key)
}
