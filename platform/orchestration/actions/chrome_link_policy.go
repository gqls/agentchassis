// FILE: platform/orchestration/actions/chrome_link_policy.go
//
// ONE answer to "which page may a piece of CHROME link to?" — bugs_open/191.
//
// Chrome (nav items, the header CTA, footer links) is not like page content.
// It ships on EVERY page, it is written once behind an idempotence gate
// (render_site_components_action.go, the EXISTS probe around :640), and it has
// no repair pass. So a chrome link to a page that has never deployed is a
// site-wide 404 that nothing later corrects.
//
// Before this file, that question had two answers rendered into ONE header:
// nav items went through loadFetchablePageSet (status floor AND
// datahelpers.NeverDeployedPagePredicate), while the header CTA was validated
// against loadResolverPageSet — the page-CONTENT set, which has no deployment
// predicate at all. mortgagecalculator.co.uk shipped a header whose nav had
// been filtered down to its single deployed page and whose CTA button, beside
// it, pointed at /tools/stamp-duty/index.html: build_status 'planned',
// deployed_at NULL, HTTP 404 on the wire.
//
// The reason the CTA could not simply have called the nav's loader is the
// structural half of the defect: the two escapes below (lookup error, zero
// deployed pages) lived INLINE inside applyNavVisibility, so the policy they
// encode was unreachable from any other caller. Extracting them is what makes
// "the two elements in one component agree" true by construction rather than
// by two authors happening to choose the same helper.
//
// Same construction, and the same reason, as ResolveChromeComponent (bug 118):
// one predicate, one resolver, and a lockstep scan in chrome_link_policy_test.go
// so a fourth caller cannot quietly hand-roll a loose set for a chrome link.

package actions

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// ChromeLinkPolicy decides whether a URL may appear as an <a href> in chrome.
//
// LANDMINE: Allows reports TRUE for everything when the policy is unfiltered
// (first build, or a failed lookup). It is a "may chrome ship this link?"
// policy, NOT a proof that a page has deployed. Anything that needs "has this
// page shipped" must use the predicate family in datahelpers/links.go
// (NeverDeployedPagePredicate / NeverDeployedPagePredicateFor), never this type.
type ChromeLinkPolicy struct {
	fetchable  datahelpers.PageURLSet
	unfiltered bool
}

// LoadChromeLinkPolicy builds the policy for one site. It never returns an
// error: every failure mode degrades to unfiltered, because an infrastructure
// blip during the one gated chrome render must not amputate a site's navigation
// or its CTA for ever.
func LoadChromeLinkPolicy(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) ChromeLinkPolicy {
	fetchable, deployedPages, err := loadFetchablePageSet(ctx, db, siteID)
	switch {
	case err != nil:
		logger.Warn("ChromeLinkPolicy: fetchable-page lookup failed; chrome links go UNFILTERED rather than risk empty chrome",
			zap.String("site_id", siteID.String()),
			zap.Error(err),
		)
		return ChromeLinkPolicy{unfiltered: true}
	case deployedPages == 0:
		// The site has not deployed a single page, so "never deployed" is true
		// of everything and carries no signal — this is a first build, not a
		// site full of dead links. Filtering here would freeze near-empty
		// chrome in place (the idempotence gate means it may never be
		// re-rendered). Note this cannot be detected from the surviving-item
		// count: loadFetchablePageSet always injects the site root, so a "Home"
		// item survives even when nothing is deployed, and the rest would be
		// silently dropped.
		logger.Warn("ChromeLinkPolicy: site has NO deployed pages — chrome links go UNFILTERED (expected during a first build, anomalous on an established site)",
			zap.String("site_id", siteID.String()),
		)
		return ChromeLinkPolicy{unfiltered: true}
	}
	return ChromeLinkPolicy{fetchable: fetchable}
}

// Unfiltered reports whether the policy is in its degraded, allow-everything
// mode. Callers that filter a SET (rather than validating one URL) need this to
// distinguish "nothing was dropped" from "dropping was disabled".
func (p ChromeLinkPolicy) Unfiltered() bool { return p.unfiltered }

// Allows reports whether chrome may link to this href.
func (p ChromeLinkPolicy) Allows(href string) bool {
	if p.unfiltered {
		return true
	}
	// Only page links can 404 against the pages table. External links, mailto
	// and in-page anchors are out of scope.
	if datahelpers.ClassifyLinkScope(href) != datahelpers.LinkScopePage {
		return true
	}
	return p.fetchable.Contains(href)
}
