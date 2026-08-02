// FILE: platform/storage/uri_helpers.go
// Centralized URI parsing and asset path utilities

package storage

import (
	"fmt"
	"strings"
)

// S3URIPrefix is the standard S3 URI scheme
const S3URIPrefix = "s3://"

// DefaultAssetBasePath is the base path for web assets
const DefaultAssetBasePath = "assets/images"

// ParseS3URI extracts bucket and key from an S3 URI
// Input: s3://bucket-name/path/to/object
// Returns: bucket, key, error
func ParseS3URI(uri string) (bucket, key string, err error) {
	if !strings.HasPrefix(uri, S3URIPrefix) {
		return "", "", fmt.Errorf("invalid S3 URI: must start with %s", S3URIPrefix)
	}

	// Remove prefix: bucket-name/path/to/object
	path := uri[len(S3URIPrefix):]

	// Split into bucket and key
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 1 || parts[0] == "" {
		return "", "", fmt.Errorf("invalid S3 URI: missing bucket name")
	}

	bucket = parts[0]
	if len(parts) == 2 {
		key = parts[1]
	}

	return bucket, key, nil
}

// ExtractKeyFromS3URI extracts just the key (path) from an S3 URI
// This is a convenience function when you don't need the bucket
func ExtractKeyFromS3URI(uri string) string {
	if !strings.HasPrefix(uri, S3URIPrefix) {
		return uri // Return as-is if not an S3 URI
	}

	path := uri[len(S3URIPrefix):]
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// BuildS3URI constructs an S3 URI from bucket and key
func BuildS3URI(bucket, key string) string {
	return fmt.Sprintf("%s%s/%s", S3URIPrefix, bucket, key)
}

// IsS3URI checks if a string is an S3 URI
func IsS3URI(uri string) bool {
	return strings.HasPrefix(uri, S3URIPrefix)
}

// PresignedURLToS3URI converts a presigned S3/B2 HTTPS URL to an s3:// URI.
//
// Input:  https://s3.us-east-005.backblazeb2.com/bucket-name/path/to/file.png?X-Amz-Algorithm=...
// Output: s3://bucket-name/path/to/file.png
//
// Returns "" if the URL doesn't match the expected pattern.
// The query string (signature, expiry, etc.) is stripped.
func PresignedURLToS3URI(presignedURL string) string {
	if presignedURL == "" {
		return ""
	}

	// Strip query parameters
	base := presignedURL
	if idx := strings.Index(base, "?"); idx >= 0 {
		base = base[:idx]
	}

	// Find the path portion after the host
	// Pattern: https://s3.region.backblazeb2.com/bucket/key
	//          https://bucket.s3.region.amazonaws.com/key  (virtual-hosted)
	protocolEnd := strings.Index(base, "://")
	if protocolEnd < 0 {
		return ""
	}
	afterProtocol := base[protocolEnd+3:]

	// Find first slash after host
	slashIdx := strings.Index(afterProtocol, "/")
	if slashIdx < 0 {
		return ""
	}

	// Path is /bucket/key — trim leading slash, split into bucket + key
	path := strings.TrimPrefix(afterProtocol[slashIdx:], "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return ""
	}

	return BuildS3URI(parts[0], parts[1])
}

// AssetPaths holds the different path formats for an asset
type AssetPaths struct {
	// RelativeURL is the web-accessible path (e.g., /assets/images/hero.jpg)
	RelativeURL string

	// FilePath is the file system path without leading slash (e.g., assets/images/hero.jpg)
	FilePath string

	// Filename is just the filename (e.g., hero.jpg)
	Filename string
}

// BuildAssetPaths generates consistent paths for an asset based on purpose
func BuildAssetPaths(purpose string, extension string) AssetPaths {
	if extension == "" {
		extension = "jpg"
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	filename := purpose + extension
	filePath := fmt.Sprintf("%s/%s", DefaultAssetBasePath, filename)
	relativeURL := "/" + filePath

	return AssetPaths{
		RelativeURL: relativeURL,
		FilePath:    filePath,
		Filename:    filename,
	}
}

// AssetKeyFilename returns the committed git filename for an asset variant:
// the asset_key with underscores rendered as hyphens, plus the given
// extension. This is the single source of truth for the variant filename
// convention shared by the deployer (deploy_image_asset) and the render-time
// resolver (plan_sections) so the path a file is committed to and the path a
// page references cannot drift.
//
//	AssetKeyFilename("hero_home", ".jpg") == "hero-home.jpg"
//	AssetKeyFilename("icon_cycle_time", "jpg") == "icon-cycle-time.jpg"
func AssetKeyFilename(assetKey, extension string) string {
	if extension == "" {
		extension = ".jpg"
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	return strings.ReplaceAll(assetKey, "_", "-") + extension
}

// assetPathsForFilename builds the three path forms for a file published into
// the shared asset directory. One spelling of the directory join, so the
// relative URL and the repo file path cannot disagree about the leading slash.
func assetPathsForFilename(filename string) AssetPaths {
	filePath := fmt.Sprintf("%s/%s", DefaultAssetBasePath, filename)
	return AssetPaths{
		RelativeURL: "/" + filePath,
		FilePath:    filePath,
		Filename:    filename,
	}
}

// brandHeadAssetPathsFor expands a brand-head purpose's one declared published
// path (BrandHeadAssetPaths) into the three forms. Derived rather than declared
// a second time — the map stays the single spelling of the path.
//
// It takes the map's value as the WHOLE PATH and splits it, rather than lifting
// the filename out and re-joining it under DefaultAssetBasePath. The difference
// only shows up on an entry that is not served from the shared asset directory —
// and a favicon at the site root (`/favicon.ico`) is a common enough convention
// that the re-joining version would have been silently wrong the first time
// somebody added one, with no test catching it. Caught by the council's
// edit-quality seat on round 1 (`abd9b119`); pinned by
// TestBrandHeadPathsAreTakenWholeNotReconstructed.
func brandHeadAssetPathsFor(purpose string) (AssetPaths, bool) {
	published, ok := BrandHeadAssetPaths[purpose]
	if !ok || !strings.HasPrefix(published, "/") {
		// A map value that is not an absolute site path cannot be split into
		// the three forms coherently; refusing is better than inventing one.
		return AssetPaths{}, false
	}
	return AssetPaths{
		RelativeURL: published,
		FilePath:    strings.TrimPrefix(published, "/"),
		Filename:    published[strings.LastIndex(published, "/")+1:],
	}, true
}

// DeployedAssetPath is THE derivation of where a generated asset is committed
// and served from. It is the one function both halves of the contract resolve
// through — the writer (deploy_image_asset_action) and every reader — so the
// path a file is written to and the path a page references cannot drift.
//
// WHY IT IS ONE FUNCTION AND NOT TWO THAT MATCH (bugs_open/168). Until
// 2026-08-02 the deployer implemented this derivation itself and DeployedWebPath
// implemented it again, kept in step by a doc comment claiming to "mirror" it.
// They did agree — but agreement maintained by comment is the drift class this
// repo keeps paying for, and the shape had already produced one near-miss: the
// 128 lane nearly shipped a check reporting a 404 for the og card and favicon of
// every site in the fleet.
//
// TWO WRITERS, and that is the fact the old signature could not express.
//
//   - deploy_image_asset_action publishes most assets. purpose fixes the
//     directory and extension; asset_key (when it differs from purpose) fixes
//     the filename, with underscores rendered as hyphens by AssetKeyFilename.
//   - derive_brand_head_assets_action publishes favicon.png and og-card.png
//     directly, under its own fixed names — `og_card` publishes as `og-card.png`,
//     with a hyphen the purpose-derived spelling does not have.
//
// A caller holding (asset_key, purpose) cannot tell which writer published the
// row, so before this the brand-head case had to be learned separately at every
// call site — 016b §9 case 7, one call site guarded and the root mechanism left
// generic. It is answered here instead, once.
//
// NOT CLOSED, and saying so beats implying otherwise: deploy_image_asset's
// `deploy_path` input overrides everything below and is invisible from
// (asset_key, purpose). No Go code sets it and it appears in zero orchestrations
// in history (audited 2026-07-31, check_image_url_404.go), so it is an unused
// passthrough — but a caller that starts setting it has left this contract.
//
//	DeployedAssetPath("hero_home", "hero").RelativeURL == "/assets/images/hero-home.jpg"
//	DeployedAssetPath("logo", "logo").RelativeURL      == "/assets/images/logo.png"
//	DeployedAssetPath("og_card", "og_card").RelativeURL == "/assets/images/og-card.png"
func DeployedAssetPath(assetKey, purpose string) AssetPaths {
	if paths, ok := brandHeadAssetPathsFor(purpose); ok {
		return paths
	}
	_, _, _, ext := GetImageConfig(purpose)
	base := BuildAssetPaths(purpose, ext)
	if assetKey == "" || assetKey == purpose {
		return base
	}
	// Keep base's directory and extension; swap the purpose-based filename for
	// the asset_key-based one.
	dotExt := base.Filename[strings.LastIndex(base.Filename, "."):]
	return assetPathsForFilename(AssetKeyFilename(assetKey, dotExt))
}

// DeployedWebPath returns the site-relative path a generated asset is served
// from (e.g. /assets/images/hero-home.jpg). Use this — NOT assets.url, which
// for most rows holds an expiring presigned S3 URL — whenever a template or a
// check needs to reference a deployed generated asset.
//
// It is the RelativeURL of DeployedAssetPath; read that function's comment for
// the contract, including the brand-head case it now answers for you (callers
// no longer need to branch on IsBrandHeadPurpose to ask "where is this served
// from?" — that predicate remains correct for the different question of "which
// table holds the evidence that it is deployed?").
//
//	DeployedWebPath("hero_home", "hero") == "/assets/images/hero-home.jpg"
//	DeployedWebPath("logo", "logo")      == "/assets/images/logo.png"
func DeployedWebPath(assetKey, purpose string) string {
	return DeployedAssetPath(assetKey, purpose).RelativeURL
}

// ImagePurposes defines valid image purposes and their configurations
var ImagePurposes = map[string]struct {
	Width     uint
	Height    uint
	Quality   int
	Extension string
}{
	"hero":          {1600, 900, 85, "jpg"},
	"hero_home":     {1600, 900, 85, "jpg"},
	"hero_about":    {1600, 900, 85, "jpg"},
	"hero_services": {1600, 900, 85, "jpg"},
	// Content hero (Phase I3.1, D14): the per-article editorial hero — same
	// geometry as hero (it renders as the article-page header and the card
	// derivation cover-crops it to 800×450), only the kind/routing differ.
	"content_hero": {1600, 900, 85, "jpg"},
	"logo":         {400, 400, 90, "png"},
	"icon":         {240, 240, 85, "jpg"},
	// Brand head assets derived from the logo (Phase I1): favicon is a small
	// square PNG; og_card is the 1200×630 social card. The derivation action
	// (derive_brand_head_assets) writes fixed filenames favicon.png /
	// og-card.png directly — see BrandHeadAssetPaths below, which is the one
	// declaration of those filenames — so these entries exist for
	// GetImageConfig completeness and any future re-optimisation pass.
	"favicon": {64, 64, 90, "png"},
	"og_card": {1200, 630, 85, "png"},
	// Sprite sheet (Phase I2): fixed 768×768 so a 3×3 grid gives known
	// 256px cells — CSS background-position slicing needs exact geometry.
	// JPG (revised from PNG 2026-07-13): a lossless PNG of a detailed glyph
	// grid (a) exceeds the Kafka git-commit message-size limit — the deploy
	// leg failed "Message Size Too Large" — and (b) blows the ≤80KB sprite
	// budget (G7). At bullet display size the jpg-muddies-lines concern is
	// imperceptible; quality 88 on a flat dark ground keeps the lines clean.
	"sprite_sheet": {768, 768, 88, "jpg"},
	// Content card (Phase I3, Lane B): the listing-card crop of a content
	// entity's image, derived from the entity's hero — never a sibling
	// generation. Same 16:9 aspect as heroes so the usual derivation is a
	// pure downscale. JPG q78: the first live run at q82 produced 64,097B
	// from a busy photographic hero — just OVER the ≤60KB card budget (D8);
	// q78 keeps dense sources under it. (D11: WebP deferred to Phase I7 —
	// no encoder in the codebase.)
	"card":    {800, 450, 78, "jpg"},
	"default": {1200, 800, 85, "jpg"},
}

// BrandHeadAssetPaths maps each brand-head asset purpose to the site-relative
// path the artefact is published at.
//
// WHY THIS EXISTS (bugs_open/142). These two artefacts are unlike every other
// image the platform generates: they are referenced from the site HEAD, which
// `injectBrandHeadTags` writes into `site_components`, and NEVER from a page
// component. Anything asking "is this site's og card deployed?" therefore has
// to know the exact published filename, and the filename is not derivable from
// the purpose — `og_card` publishes as `og-card.png`, with a hyphen. Before
// this map, that pair of strings was spelled as a literal in four places in
// derive_brand_head_assets_action.go and twice more in injectBrandHeadTags, and
// the undeployed-asset discovery check reconstructed a THIRD spelling from the
// purpose and got it wrong. Two hand-maintained copies of one string that must
// stay identical is the drift class this repo keeps paying for.
//
// The deriver still spells its own literals; TestBrandHeadAssetPathsMatchTheDeriver
// (discovery_checks package) scans that file and fails the build if it publishes
// a brand-head path this map does not carry. Adopting the map inside the deriver
// is left to that file's owning lane — it was being actively edited for
// bugs_open/143 when this went in, and two sessions in one file is the one
// collision no hook can prevent.
//
// CORRECTED 2026-08-02 (bugs_open/168) — this comment used to end by explaining
// WHY NOT DeployedWebPath, on the ground that the generic helper "cannot express
// these paths". That was true and is no longer: DeployedAssetPath now consults
// this map first, so the helper answers the brand-head case correctly and the
// two are no longer rival answers to one question.
//
// The measurement that justified the split is kept, because it is what the
// helper is now protected against re-deriving. Before the fix:
//
//	DeployedWebPath("og_card", "og_card") == "/assets/images/og_card.png"   underscore, file is og-card.png
//	DeployedWebPath("",        "og_card") == "/assets/images/og_card.png"   same
//	DeployedWebPath("og_card", "")        == "/assets/images/og-card.jpg"   right name, wrong extension
//
// No argument pair yielded "/assets/images/og-card.png". The hyphen-swap in
// AssetKeyFilename fires only when assetKey differs from purpose, and for these
// two artefacts it does not; `favicon` came out right purely because it has no
// underscore to disagree about.
//
// So this map is still the one DECLARATION of these filenames — it has to be,
// because they are not derivable from the purpose — but it is now an INPUT to
// the shared derivation rather than a parallel one.
//
// BOTH WRITERS ARE PINNED, by two tests with different origins — spelled out
// because the council's edit-quality seat read this list as citing a guard the
// change had never built (`abd9b119`, round 1), which was a fair reading of the
// earlier wording:
//
//   - the DEPLOYER is pinned by construction — it calls DeployedAssetPath rather
//     than deriving its own path — plus TestDeployImageAssetResolvesThroughThe
//     SharedDerivation (platform/orchestration/actions), ADDED with this change,
//     which fails if it ever goes back to re-implementing the convention.
//   - the DERIVER is pinned by TestBrandHeadAssetPathsMatchTheDeriver
//     (discovery_checks package, check_undeployed_assets_test.go) — PRE-EXISTING,
//     written by the bugs_open/142 lane, NOT added here. It scans
//     derive_brand_head_assets_action.go's recordDerivedAsset call sites and fails
//     the build if it publishes a brand-head path this map does not carry.
//
// This map's own values are pinned against the derivation by
// TestDeployedAssetPathAgreesWithTheMapLiteral and
// TestBrandHeadPathsAreTakenWholeNotReconstructed (this package).
//
// Keys are `assets.purpose` values; values are site-relative, leading slash
// included, exactly as they appear in the rendered head.
var BrandHeadAssetPaths = map[string]string{
	"favicon": "/" + DefaultAssetBasePath + "/favicon.png",
	"og_card": "/" + DefaultAssetBasePath + "/og-card.png",
}

// IsBrandHeadPurpose reports whether an asset purpose is a brand-head artefact —
// i.e. one whose deployed reference lives in the site head rather than in any
// page component. Callers reasoning about "is this deployed?" must branch on
// this: the evidence lives in a different table for these two.
func IsBrandHeadPurpose(purpose string) bool {
	_, ok := BrandHeadAssetPaths[purpose]
	return ok
}

// GetImageConfig returns the configuration for an image purpose
func GetImageConfig(purpose string) (width, height uint, quality int, extension string) {
	cfg, ok := ImagePurposes[purpose]
	if !ok {
		cfg = ImagePurposes["default"]
	}
	return cfg.Width, cfg.Height, cfg.Quality, cfg.Extension
}
