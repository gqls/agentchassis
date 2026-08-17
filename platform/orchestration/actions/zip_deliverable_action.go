// FILE: platform/orchestration/actions/zip_deliverable_action.go
//
// zip_deliverable: cut a downloadable ZIP of one site's built artefact tree
// (the objects under portfolio-sites/<domain>/), upload it under
// deliverables/<domain>/, and return a presigned download URL. The ZIP is the
// ownership artefact of the delivery architecture (PLAN 2026-08-17): it is
// what the customer downloads, what a Netlify deploy uploads, and what "take
// it elsewhere" hands over — one build, three doors.
//
// MUST run in a spawned storage-enabled pod (the standing chassis carries no
// B2 credentials, owner ruling 2026-08-08), so it constructs its own client
// via the newPortfolioStore idiom shared with publish_site_action.go.
//
// The archive is composed to a temp file, NOT buffered in memory and NOT
// streamed bare: B2's S3 gateway 411s a non-seekable body
// (MissingContentLength — the first live canary, fixed in b4981634d), and a
// whole-site ZIP is too big to buffer the way b2worker buffers small site
// files. A seekable *os.File lets the SDK compute Content-Length.
//
// A truncated ZIP is a silent contractual failure (the PLAN's ranked risk 3),
// so the action verifies at the artefact before reporting success: the
// archive is re-opened and its entry count must equal the tree listing, the
// index.html entry's bytes must equal a fresh origin read, and the uploaded
// object's remote size must equal the local archive's. Oversized trees ALERT
// and continue — never truncate.
package actions

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/publish"
	"go.uber.org/zap"
)

var ZipDeliverableInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"domain"},
	Optional:    []string{"source_bucket", "expiry_minutes", "size_alert_bytes"},
	Defaults: map[string]interface{}{
		"source_bucket": "portfolio-sites",
		// 7 days: the delivery email's link must survive a weekend unread.
		"expiry_minutes": 10080,
		// Alert threshold, not a cap: past this the result carries
		// size_alert=true and the log warns, but the cut always completes.
		// Configurable so an induced oversize can prove the alert path fires
		// (the acceptance's demand control).
		"size_alert_bytes": 536870912,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("zip_deliverable", ZipDeliverableInputSpec)
}

// presigner is the slice of the storage client this action needs beyond
// publish.ObjectStore. *storage.S3Client satisfies it; expiry is in minutes.
type presigner interface {
	GetPresignedURL(ctx context.Context, key string, expiryMinutes int) (string, error)
}

func ZipDeliverableAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "zip_deliverable"),
		zap.String("step_name", params.ExecutionContext.StepName))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(params.CollectedData, params.StepConfig.Config, ZipDeliverableInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("zip_deliverable: extract inputs: %w", err)
	}
	domain := strings.TrimSpace(inputs.Get("domain"))
	if domain == "" {
		return map[string]interface{}{
			"zipped": false, "skipped": true, "reason": "no domain resolvable from inputs",
		}, nil
	}

	store, err := newPortfolioStore(ctx, inputs.Get("source_bucket"), logger)
	if err != nil {
		return nil, fmt.Errorf("zip_deliverable: portfolio store unavailable (is this a storage-enabled spawned pod?): %w", err)
	}

	source := publish.NewS3Source(store, domain)
	files, err := source.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("zip_deliverable: %w", err)
	}
	if len(files) == 0 {
		return map[string]interface{}{
			"zipped": false, "skipped": true,
			"reason": fmt.Sprintf("no built artefacts under %s/%s/ — nothing to deliver", inputs.Get("source_bucket"), domain),
		}, nil
	}

	// The tree hash names this cut: same tree, same key, so a re-cut
	// overwrites its equivalent instead of accumulating copies.
	treeHash := publish.TreeHash(files)
	sizeAlertBytes := inputs.GetInt("size_alert_bytes", 536870912)
	var totalSourceBytes int64
	for _, f := range files {
		totalSourceBytes += f.Size
	}
	sizeAlert := sizeAlertBytes > 0 && totalSourceBytes > int64(sizeAlertBytes)
	if sizeAlert {
		logger.Warn("zip_deliverable: tree exceeds size alert threshold — cutting anyway, never truncating",
			zap.String("domain", domain),
			zap.Int64("total_source_bytes", totalSourceBytes),
			zap.Int("size_alert_bytes", sizeAlertBytes))
	}

	zipPath, zipSize, err := composeZip(ctx, source, files)
	if zipPath != "" {
		defer os.Remove(zipPath)
	}
	if err != nil {
		return nil, fmt.Errorf("zip_deliverable: %w", err)
	}

	// Verify the archive at the artefact before it leaves the pod: entry
	// count against the listing, and index.html bytes against a fresh origin
	// read (proves the ZIP round-trips, not just that writes returned nil).
	if err := verifyArchive(ctx, zipPath, source, files); err != nil {
		return nil, fmt.Errorf("zip_deliverable: %w", err)
	}

	hashTail := treeHash
	if i := strings.IndexByte(hashTail, ':'); i >= 0 {
		hashTail = hashTail[i+1:]
	}
	if len(hashTail) > 12 {
		hashTail = hashTail[:12]
	}
	key := fmt.Sprintf("deliverables/%s/%s-%s.zip", domain, domain, hashTail)

	zf, err := os.Open(zipPath)
	if err != nil {
		return nil, fmt.Errorf("zip_deliverable: reopen archive for upload: %w", err)
	}
	defer zf.Close()
	// *os.File is seekable, so the SDK sends Content-Length (B2 requires it).
	if _, err := store.Upload(ctx, key, "application/zip", zf); err != nil {
		return nil, fmt.Errorf("zip_deliverable: upload %s: %w", key, err)
	}

	// Truncation check at the remote artefact: the stored object's size must
	// equal the local archive's, read back from the listing rather than
	// trusted from the upload return.
	if err := verifyRemoteSize(ctx, store, key, zipSize); err != nil {
		return nil, fmt.Errorf("zip_deliverable: %w", err)
	}

	expiryMinutes := inputs.GetInt("expiry_minutes", 10080)
	ps, ok := store.(presigner)
	if !ok {
		return nil, fmt.Errorf("zip_deliverable: store %T cannot presign download URLs", store)
	}
	url, err := ps.GetPresignedURL(ctx, key, expiryMinutes)
	if err != nil {
		return nil, fmt.Errorf("zip_deliverable: presign %s: %w", key, err)
	}

	logger.Info("zip deliverable cut",
		zap.String("domain", domain), zap.String("zip_key", key),
		zap.Int64("zip_size_bytes", zipSize), zap.Int("files", len(files)),
		zap.String("tree_hash", treeHash), zap.Bool("size_alert", sizeAlert))

	return map[string]interface{}{
		"zipped":             true,
		"zip_key":            key,
		"zip_size_bytes":     zipSize,
		"files":              len(files),
		"tree_hash":          treeHash,
		"total_source_bytes": totalSourceBytes,
		"size_alert":         sizeAlert,
		"presigned_url":      url,
		"expiry_minutes":     expiryMinutes,
	}, nil
}

// composeZip streams every file of the tree into a temp-file archive and
// returns its path and size. The path is returned even on error so the
// caller can remove the partial file.
func composeZip(ctx context.Context, source publish.Source, files []publish.File) (string, int64, error) {
	tmp, err := os.CreateTemp("", "zip-deliverable-*.zip")
	if err != nil {
		return "", 0, fmt.Errorf("create temp archive: %w", err)
	}
	path := tmp.Name()

	sorted := make([]publish.File, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Key < sorted[j].Key })

	now := time.Now().UTC()
	zw := zip.NewWriter(tmp)
	for _, f := range sorted {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: f.Key, Method: zip.Deflate, Modified: now})
		if err != nil {
			tmp.Close()
			return path, 0, fmt.Errorf("archive entry %s: %w", f.Key, err)
		}
		rc, err := source.Open(ctx, f.Key)
		if err != nil {
			tmp.Close()
			return path, 0, fmt.Errorf("open %s: %w", f.Key, err)
		}
		n, err := io.Copy(w, rc)
		rc.Close()
		if err != nil {
			tmp.Close()
			return path, 0, fmt.Errorf("copy %s into archive: %w", f.Key, err)
		}
		if n != f.Size {
			// The tree changed between listing and read — this cut no longer
			// matches its hash; fail and let the caller re-dispatch.
			tmp.Close()
			return path, 0, fmt.Errorf("copy %s: read %d bytes, listing said %d — tree changed mid-cut", f.Key, n, f.Size)
		}
	}
	if err := zw.Close(); err != nil {
		tmp.Close()
		return path, 0, fmt.Errorf("finalise archive: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return path, 0, fmt.Errorf("close temp archive: %w", err)
	}
	st, err := os.Stat(path)
	if err != nil {
		return path, 0, fmt.Errorf("stat archive: %w", err)
	}
	return path, st.Size(), nil
}

// verifyArchive re-opens the finished archive and proves it holds the tree:
// entry count equals the listing, and (when present) the index.html entry's
// bytes hash-match a fresh origin read.
func verifyArchive(ctx context.Context, zipPath string, source publish.Source, files []publish.File) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("verify: reopen archive: %w", err)
	}
	defer zr.Close()

	if len(zr.File) != len(files) {
		return fmt.Errorf("verify: archive holds %d entries, tree listing has %d files", len(zr.File), len(files))
	}

	hasIndex := false
	for _, f := range files {
		if f.Key == "index.html" {
			hasIndex = true
			break
		}
	}
	if !hasIndex {
		return nil
	}

	var zipIndex []byte
	for _, zf := range zr.File {
		if zf.Name == "index.html" {
			rc, err := zf.Open()
			if err != nil {
				return fmt.Errorf("verify: open archived index.html: %w", err)
			}
			zipIndex, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return fmt.Errorf("verify: read archived index.html: %w", err)
			}
			break
		}
	}
	if zipIndex == nil {
		return fmt.Errorf("verify: index.html in listing but not in archive")
	}

	rc, err := source.Open(ctx, "index.html")
	if err != nil {
		return fmt.Errorf("verify: origin open index.html: %w", err)
	}
	defer rc.Close()
	origin, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("verify: origin read index.html: %w", err)
	}
	if !bytes.Equal(zipIndex, origin) {
		o, z := sha256.Sum256(origin), sha256.Sum256(zipIndex)
		return fmt.Errorf("verify: archived index.html sha256 %x != origin %x", z[:8], o[:8])
	}
	return nil
}

// verifyRemoteSize confirms the uploaded object's stored size equals the
// local archive's, from the destination listing (the upload return value is
// not evidence — a truncated ZIP is a silent contractual failure).
func verifyRemoteSize(ctx context.Context, store publish.ObjectStore, key string, want int64) error {
	objs, err := store.ListObjects(ctx, key)
	if err != nil {
		return fmt.Errorf("verify upload: list %q: %w", key, err)
	}
	for _, o := range objs {
		if o.Key == key {
			if o.Size != want {
				return fmt.Errorf("verify upload: remote size %d != local archive %d — truncated upload", o.Size, want)
			}
			return nil
		}
	}
	return fmt.Errorf("verify upload: %q absent from destination listing", key)
}
