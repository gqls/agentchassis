// FILE: internal/adapters/browserrunner/screenshots.go
//
// P3 — screenshots on failure. When any check fails on a (url, profile) run,
// the runner captures a full-page screenshot and uploads it to object storage
// (the imagegenerator's S3 client), returning:
//
//   uri      — the durable s3://bucket/key pointer. This is what goes into a
//              travelling-doc NOTE: notes are loaded into LLM prompt contexts
//              by load_doc_context, so they must never carry presigned-URL
//              signatures (hundreds of chars, expiring) or inline image bytes.
//   view_url — a presigned GET a human can click from the work item's spec.
//              Expires (7 days, the S3 presign maximum); the uri outlives it
//              and can be re-presigned on demand.
//
// Screenshots are EVIDENCE, never a dependency: with no object storage
// configured the runner behaves exactly as before P3, and a failed capture or
// upload degrades to a log line — it can never fail, slow-fail, or skew the
// acceptance verdict itself (the "docs never fail the work" rule).

package browserrunner

import (
	"bytes"
	"context"
	"fmt"
	"regexp"

	"go.uber.org/zap"

	platformconfig "github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/storage"
)

// screenshotStore persists one captured screenshot and returns its durable
// object URI plus a time-limited view URL. The test seam for S3.
type screenshotStore interface {
	Save(ctx context.Context, key string, png []byte) (uri, viewURL string, err error)
}

// viewURLExpiryMinutes is 7 days — the S3 presign maximum and the
// imagegenerator's convention for generated-image links.
const viewURLExpiryMinutes = 7 * 24 * 60

type s3ScreenshotStore struct {
	client storage.Client
}

func (s *s3ScreenshotStore) Save(ctx context.Context, key string, png []byte) (string, string, error) {
	uri, err := s.client.Upload(ctx, key, "image/png", bytes.NewReader(png))
	if err != nil {
		return "", "", err
	}
	viewURL, err := s.client.GetPresignedURL(ctx, key, viewURLExpiryMinutes)
	if err != nil {
		// The upload stood: a missing view link degrades the convenience,
		// it does not discard the evidence.
		return uri, "", nil
	}
	return uri, viewURL, nil
}

// newScreenshotStore builds the S3-backed store from the service config, or
// returns nil when object storage is not configured / not constructible.
// nil means "screenshots disabled" — the runner's P0–P2 behaviour is
// untouched either way.
func newScreenshotStore(ctx context.Context, cfg *platformconfig.ServiceConfig, logger *zap.Logger) screenshotStore {
	oc := cfg.Infrastructure.ObjectStorage
	if oc.AccessKeyEnvVar == "" && oc.Bucket == "" {
		logger.Info("failure screenshots disabled: no infrastructure.object_storage in config")
		return nil
	}
	client, err := storage.NewS3Client(ctx, oc, *logger)
	if err != nil {
		logger.Error("failure screenshots disabled: object storage configured but the client would not build — checks still run, evidence is skipped",
			zap.Error(err))
		return nil
	}
	logger.Info("failure screenshots enabled", zap.String("bucket", oc.Bucket))
	return &s3ScreenshotStore{client: client}
}

// keyUnsafe matches anything we won't put in an object key segment. Dots are
// excluded too — no legitimate segment (uuid, slug, profile) contains one, and
// that keeps ".." out of keys by construction.
var keyUnsafe = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func keySegment(s, fallback string) string {
	s = keyUnsafe.ReplaceAllString(s, "-")
	if s == "" || s == "-" {
		return fallback
	}
	return s
}

// screenshotKey names one captured screenshot:
//
//	acceptance-evidence/<site_id>/<function>/<run_id>_<profile>[_<urlIdx>].png
//
// The prefix keeps evidence apart from the bucket's asset images; run_id is
// the orchestration id, so one Tier-4 run's evidence groups together.
func screenshotKey(siteID, function, runID, profile string, urlIdx int) string {
	suffix := ""
	if urlIdx > 0 {
		suffix = fmt.Sprintf("_%d", urlIdx)
	}
	return fmt.Sprintf("acceptance-evidence/%s/%s/%s_%s%s.png",
		keySegment(siteID, "unknown-site"),
		keySegment(function, "unknown-tool"),
		keySegment(runID, "unknown-run"),
		keySegment(profile, "profile"),
		suffix)
}
