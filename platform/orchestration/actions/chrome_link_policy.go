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
	"regexp"

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

// chromeStoredHrefs pulls every href out of a stored chrome artefact. Deliberately
// a plain scan over the rendered bytes rather than a read of some structured copy:
// the CTA lives ONLY in site_components.rendered_html — there is no `cta_url` key
// anywhere in the schema, and bugs_open/191's own diagnosis run was recorded
// UNVERIFIABLE precisely because it queried `content_data->>'cta_url'` and got a
// 0-row answer that could never have been anything else.
var chromeStoredHrefs = regexp.MustCompile(`href="([^"]*)"`)

// staleChromeLinkSlots reports which of the given slots ALREADY hold chrome that
// links somewhere this policy refuses.
//
// It exists because the idempotence exit in renderAndStoreSiteComponent asks
// "does this slot hold HTML?", never "does it hold HTML that still satisfies
// policy?" — so without this, a corrected predicate reaches only sites whose
// chrome happens to be re-rendered for some other reason, and every header
// already serving a dead CTA keeps it indefinitely (bugs_open/191). Same shape,
// and the same answer, as bugs_open/166's repointRetiredChromeSlot: detect
// BEFORE the exit and let the slot through the exit's own force channel.
//
// Fail-safe direction is deliberately "do not re-render": an unreadable row, a
// failed query or an unfiltered policy all return no slots. Re-rendering chrome
// is a write, and a write on the strength of a failed read is how a bad artefact
// gets manufactured rather than repaired.
func staleChromeLinkSlots(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	slots []string,
	policy ChromeLinkPolicy,
	logger *zap.Logger,
) map[string]bool {
	stale := map[string]bool{}
	// An unfiltered policy allows everything, so nothing can be stale against it.
	// Checking here also means a first build issues no query at all.
	if policy.Unfiltered() || len(slots) == 0 {
		return stale
	}

	// The slots being rendered are filtered in Go rather than passed as a SQL
	// array: a site holds a handful of chrome slots, and this package
	// deliberately carries no dependency on lib/pq's pq.Array (see the note on
	// toPGTextArrayLiteral in resolve_composition_helpers.go).
	wanted := make(map[string]bool, len(slots))
	for _, s := range slots {
		wanted[s] = true
	}

	rows, err := db.QueryContext(ctx, `
		SELECT slot_name, COALESCE(rendered_html, '')
		FROM site_components
		WHERE site_id = $1
		  AND rendered_html IS NOT NULL AND rendered_html <> ''
	`, siteID)
	if err != nil {
		logger.Warn("staleChromeLinkSlots: stored-chrome lookup failed; leaving the idempotence exit alone",
			zap.String("site_id", siteID.String()), zap.Error(err))
		return stale
	}
	defer rows.Close()

	for rows.Next() {
		var slot, html string
		if err := rows.Scan(&slot, &html); err != nil {
			continue
		}
		if !wanted[slot] {
			continue
		}
		for _, m := range chromeStoredHrefs.FindAllStringSubmatch(html, -1) {
			if policy.Allows(m[1]) {
				continue
			}
			// Name the href. "Chrome was re-rendered" with no reason attached is
			// indistinguishable from a spurious rebuild loop to whoever reads
			// the log next.
			logger.Warn("staleChromeLinkSlots: stored chrome links to a page the policy now refuses — forcing a re-render of this slot (bugs_open/191)",
				zap.String("site_id", siteID.String()),
				zap.String("slot", slot),
				zap.String("href", m[1]),
			)
			stale[slot] = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		logger.Warn("staleChromeLinkSlots: stored-chrome iteration failed; the slots found so far still stand",
			zap.String("site_id", siteID.String()), zap.Error(err))
	}
	return stale
}
