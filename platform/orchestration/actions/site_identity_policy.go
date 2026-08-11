package actions

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// siteIdentityPolicy carries the site-level switches that govern who decides a
// page's identity during a re-plan (bugs_open/215, the quiet mode).
//
// ONE FACT, ONE READER — the same rule site_url_shape.go states for the flat-URL
// flag, and for the same reason. HonourRealisedIdentity is read by BOTH
// canonicalisation surfaces (WriteSitePlanAction writes site_plan_pages,
// SyncPagesToDBAction writes pages); if they ever read it from two places and
// disagree, the plan and the pages table carry different names for one page —
// which is worse than the bug being fixed, because it breaks the
// site_plan_pages.name = pages.name join that check_sectionless_pages and the
// reconciler both depend on. TestIdentityPolicyReachesBothCanonicalisationSurfaces
// pins that they both go through this helper.
//
// SITE-LEVEL RATHER THAN STEP CONFIG, deliberately. A step config would be a
// fleet-wide switch flipped in one migration, and there is no honest way to
// pilot one: the first site to exercise it would be whichever site replanned
// next. Per site, the switch can be turned on where the population has actually
// been measured, which is the discipline the stem layer in particular needs.
//
// Both default FALSE — the unsafe default is off, per the owner ruling of
// 2026-08-02 (new authority on a shared seam ships as a field with the unsafe
// default off; a comment is not a control on a tree this many sessions share).
// Absent spec, absent key, nil db, or a read error all mean false, so a site
// that has never heard of these keys behaves exactly as it does today.
type siteIdentityPolicy struct {
	// HonourRealisedIdentity: when a plan page was derived from a realised page
	// (the reconciler stamped identity_authority="realised"), keep that page's
	// STORED name/url/page_type instead of re-deriving them from its role.
	//
	// Why it must be switchable at all rather than simply correct: for the 71
	// live shipped rows measured fleet-wide on 2026-08-11 whose stored identity
	// is not a fixed point of CanonicalisePage, turning this on CHANGES the name
	// and URL written for that page — from the canonicaliser's answer to the
	// one the site is actually serving. That is the point, and it is also a
	// visible change to live artefacts, so it is opted into per site.
	HonourRealisedIdentity bool

	// StemTwinSnap: enable the reconciler's weakest twin-identity layer, which
	// matches a bare plan page against a prefixed realised one (tools ->
	// tool-tools). Off, the layer still records what it would have done.
	//
	// Separately switchable from HonourRealisedIdentity because they fail
	// differently: this one can pair two pages that are not the same page, while
	// the other only ever preserves an identity that already exists.
	StemTwinSnap bool
}

// siteIdentityPolicyFor reads the policy from the site's current structure
// spec — the same row and aspect siteUsesFlatURLs reads, because these are the
// same kind of fact about a site (how its page identities are shaped) and
// splitting them across two specs would be two things to keep in step.
//
// Nothing writes these keys automatically; they are seeded per site into the
// structure aspect (supersede-then-insert, see any SEED_*.sql).
func siteIdentityPolicyFor(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) siteIdentityPolicy {
	var policy siteIdentityPolicy
	if db == nil {
		return policy
	}
	var honour, stem bool
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE((data->>'honour_realised_identity')::boolean, false),
		       COALESCE((data->>'stem_twin_snap')::boolean, false)
		FROM site_specs
		WHERE site_id = $1 AND aspect = 'structure' AND is_current = true
	`, siteID).Scan(&honour, &stem)
	switch {
	case err == sql.ErrNoRows:
		return policy
	case err != nil:
		// A malformed value (anything ::boolean rejects) lands here too, and the
		// answer is the same as absent: the site keeps today's behaviour. Warn
		// rather than fail — a plan write must not be lost over a spec typo.
		logger.Warn("siteIdentityPolicyFor: structure spec read failed; keeping today's identity behaviour",
			zap.String("site_id", siteID.String()),
			zap.Error(err))
		return policy
	}
	policy.HonourRealisedIdentity = honour
	policy.StemTwinSnap = stem
	return policy
}

// realisedIdentityOf returns the stored identity a plan page carries when the
// reconciler derived it from a realised page, and ok=false for every ordinary
// LLM-proposed page.
//
// This is the guard that makes the reconciler's decision survive the write path
// (bugs_open/215). Both canonicalisation surfaces re-run CanonicalisePage over
// every page they are given, and CanonicalisePage cannot EXPRESS a legacy
// identity — a tool-typed page always comes back "tool-<bare>" at the role's
// default hub. So without this, a reconciler that correctly recognises
// "llm-cost-calculator" as an existing page still hands the writer a page that
// is re-derived to "tool-llm-cost-calculator", conflicts with nothing on
// (site_id, name), and is INSERTed as a second row: the phantom, re-minted by
// the very pass that spotted it.
//
// Fails safe in both directions. The marker is minted only by
// reconcilePlanWithRealised (which strips any the LLM emitted, so a plan cannot
// forge one), and a marked page missing any of the three identity fields falls
// through to canonicalisation rather than writing a blank.
func realisedIdentityOf(page map[string]interface{}) (name, url, pageType string, ok bool) {
	if authority, _ := page["identity_authority"].(string); authority != "realised" {
		return "", "", "", false
	}
	name, _ = page["name"].(string)
	url, _ = page["url"].(string)
	pageType, _ = page["page_type"].(string)
	if name == "" || url == "" || pageType == "" {
		return "", "", "", false
	}
	return name, url, pageType, true
}
