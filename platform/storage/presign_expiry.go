// FILE: platform/storage/presign_expiry.go
//
// The one definition of how long a presigned URL may live, and the clamp that
// enforces it for every caller of this package.
//
// WHY THIS IS HERE AND NOT IN THE CALLER. A presigned URL is bounded by the SigV4
// signing protocol at 604,800 seconds (7 days). The AWS SDK does NOT enforce it —
// checked 2026-08-20 against `aws-sdk-go-v2 v1.25.1` (`aws/signer/v4`) and
// `service/s3 v1.51.0`: neither contains the ceiling, so `PresignGetObject` signs
// whatever duration it is handed and returns a well-formed URL. The object store
// refuses it only when someone clicks, and refuses it as
// **`SignatureDoesNotMatch`** — which reads as wrong credentials or a clock skew,
// sending whoever debugs it to the one part that is healthy.
//
// MEASURED against the live bucket, with a deliberately absent key so the HTTP
// status is the whole answer and the 404 arm is the control:
//
//	X-Amz-Expires=604800  -> HTTP 404 NoSuchKey            (signature accepted)
//	X-Amz-Expires=604801  -> HTTP 403 SignatureDoesNotMatch
//	X-Amz-Expires=3628800 -> HTTP 403 SignatureDoesNotMatch  (six weeks)
//
// Exact to the second.
//
// WHY THE CLAMP LIVES AT THE SHARED HELPER, which is the correction a council
// round earned (DGH-014 round 2, `reuse_agent` and `prior_art_librarian`): the
// first version of this put the ceiling in the ONE new package that wanted it,
// which protected the only call site that could not yet get it wrong and left the
// six that could. Enumerated 2026-08-20 — every presign call site in the estate,
// and NONE of them clamped:
//
//	platform/orchestration/actions/zip_deliverable_action.go:165   inputs.GetInt("expiry_minutes", 10080)
//	platform/orchestration/actions/ingest_staged_asset_action.go:205   7*24*60
//	platform/orchestration/actions/storage_actions.go:131, :640    60*24*7
//	internal/adapters/webscrape/adapter.go:1025                    10080
//	internal/adapters/imagegenerator/dynamic_adapter.go:635        10080
//	internal/adapters/browserrunner/screenshots.go:42              7*24*60 (const)
//
// Six sites, six hand-written spellings of the same magic number, no clamp. The
// first one is the live exposure: `expiry_minutes` is an ACTION INPUT, so a seed
// or a dispatch payload can set it to anything, and before this clamp that minted
// a link which reported success and then 403'd in a customer's browser.
//
// WHY CLAMPING SILENTLY IS RIGHT HERE, and it is not the usual answer. Normally a
// helper that quietly alters its caller's request is hiding a bug. This one cannot
// be: there is NO value above the ceiling that would have worked, so the choice is
// between a URL that functions for 7 days and a URL that functions for none. The
// clamp can only convert a guaranteed failure into a success. (`S3Client` carries
// no logger — only the constructor takes one — so warning would mean widening the
// struct and every construction site; `ClampPresignExpiryMinutes` is exported so a
// caller that wants to notice can compare before it calls.)
//
// NO LIVE CALLER CHANGES BEHAVIOUR. All six sites above pass exactly the ceiling,
// so the clamp is a no-op for every one of them today and bites only a future
// mistake or a config-supplied override. That is measured, not assumed — it is the
// list above.
package storage

// MaxPresignExpiryMinutes is the SigV4 ceiling on a presigned URL, in minutes
// (604,800 seconds). It is the ONE definition in the estate; do not re-derive it.
// If an object store ever raises its cap, this constant and the measurement in the
// file comment are what go stale together, which is why the measurement is dated.
const MaxPresignExpiryMinutes = 7 * 24 * 60

// ClampPresignExpiryMinutes returns the largest expiry the signing protocol will
// actually honour for the requested one. A non-positive request also returns the
// ceiling: a zero or negative expiry signs a URL that is already dead, which is
// the same class of silently-useless artefact this file exists to prevent.
func ClampPresignExpiryMinutes(requested int) int {
	if requested <= 0 || requested > MaxPresignExpiryMinutes {
		return MaxPresignExpiryMinutes
	}
	return requested
}
