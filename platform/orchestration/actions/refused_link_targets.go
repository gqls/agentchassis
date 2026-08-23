// FILE: platform/orchestration/actions/refused_link_targets.go
//
// The site-level half of outbound link suppression — bugs_open/328.
//
// datahelpers/link_suppress.go knows how to remove an anchor. This file decides
// WHICH targets are refused, and it is deliberately a sibling of
// chrome_link_policy.go rather than a new invention: same shape, same two
// escapes, same fail directions, for the same reasons written out there. 328 is
// the page-BODY spelling of the question bugs_open/191 answered for chrome.
//
// THE ONE DIFFERENCE FROM CHROME, and it is why the escapes can be copied but
// the conclusion cannot. Chrome renders once behind an idempotence gate and has
// no repair pass, so a chrome link dropped in error may never come back. Body
// suppression runs on the ASSEMBLED STRING on its way to deploy, recomputed on
// every render, and writes nothing: content_data keeps the authored href. So a
// target wrongly refused today is linked again the moment the referrer next
// renders, and a target that ships tomorrow needs no cascade to be re-linked.
// That asymmetry is what lets the predicate carry a time threshold at all.
//
// OPT-IN, DEFAULT OFF (owner ruling 2026-08-02, RFC_010 §2). Suppression is a
// writer action, so the unsafe side is the one that acts, and it is off unless a
// step config says `suppress_unshipped_links: true`. The ruling's point is that
// the decision must be visible to a reviewer of the CALLER: it is a field in the
// step config, not a constant in this package, precisely so that the migration
// enabling it is the thing a reviewer reads.
package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// suppressUnshippedLinksKey is the opt-in step-config field. Named once here so
// the migration, the tests and the reader are looking at one literal.
const suppressUnshippedLinksKey = "suppress_unshipped_links"

// linkSuppressionErrorCode is a FOURTH code beside CONTENT_LINK_REPAIR_DETAIL,
// CONTENT_LINK_REPAIR_SKIPPED and CONTENT_DATA_LINK_AUDIT, for the reason
// save_sections_content_data_links.go gives for being distinct from the first
// two: it answers a different question. The repair rows say "this href pointed
// at nothing"; these say "this href pointed at a page of this site that has
// never been served and is not arriving" — a different defect with a different
// owner and a different fix, and folding them together would make every existing
// query on the repair code start returning a population it was not written for.
const linkSuppressionErrorCode = "CONTENT_LINK_SUPPRESSED_UNSHIPPED"

// loadRefusedLinkTargets returns the page URLs an outbound seam must not link
// to, and whether the answer is TRUSTWORTHY.
//
// The bool is load-bearing in exactly the way loadValidPagePaths' is, but in the
// opposite direction: there, a failed load would strip every link on the page;
// here a failed load must strip none. Both failure modes and both escapes return
// (nil, false), and a nil set makes SuppressRefusedPageLinks a no-op, so the
// page ships precisely what it holds today.
//
// ESCAPE 1 — a failed lookup. An infrastructure blip must not amputate a page's
// internal links.
//
// ESCAPE 2 — the site has never shipped a page. "Never deployed" is then true of
// everything and carries no signal: this is a first build or a pre-adoption
// site, not a site full of dead links. Measured 2026-08-23, this escape is not
// theoretical — one live domain in the fleet serves a parked-registrar redirect
// on every path while holding a full set of never-deployed page rows, and
// without this escape it would be the site suppression hit hardest.
//
// ⚠ THE ESCAPE KEYS ON THE SHIPPED-PAGE COUNT, NOT ON THE SIZE OF THE REFUSED
// SET. An empty refused set means "nothing is refused", which is the normal,
// healthy state of most sites; it says nothing about whether the site has
// shipped. Reading emptiness as the first-build signal is the mistake
// chrome_link_policy.go records having been caught by a test.
func loadRefusedLinkTargets(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (datahelpers.PageURLSet, bool) {
	if db == nil {
		return nil, false
	}

	rows, err := db.QueryContext(ctx, `
		SELECT p.url,
		       `+datahelpers.PageLinkRefusedPredicateFor("p")+` AS refused,
		       `+datahelpers.PageHasShippedPredicateFor("p")+` AS shipped
		FROM pages p
		WHERE p.site_id = $1 AND p.`+linkablePageStatusPredicate, siteID)
	if err != nil {
		logger.Warn("link suppression: refused-target lookup failed; outbound links go UNSUPPRESSED rather than risk stripping a page's links (bugs_open/328)",
			zap.String("site_id", siteID.String()), zap.Error(err))
		return nil, false
	}
	defer rows.Close()

	var refusedURLs []string
	shippedPages := 0
	for rows.Next() {
		var url sql.NullString
		var refused, shipped bool
		if err := rows.Scan(&url, &refused, &shipped); err != nil {
			logger.Warn("link suppression: scan failed; outbound links go UNSUPPRESSED (bugs_open/328)",
				zap.String("site_id", siteID.String()), zap.Error(err))
			return nil, false
		}
		if shipped {
			shippedPages++
		}
		// A NULL url is not a link target and not evidence of truncation.
		if refused && url.Valid && url.String != "" {
			refusedURLs = append(refusedURLs, url.String)
		}
	}
	if err := rows.Err(); err != nil {
		// A mid-iteration failure TRUNCATES the set. For this function a short
		// set under-suppresses rather than over-suppresses, which is the safe
		// direction — but it is still an untrustworthy answer, and saying so
		// costs nothing.
		logger.Warn("link suppression: page list truncated by a row error; outbound links go UNSUPPRESSED (bugs_open/328)",
			zap.String("site_id", siteID.String()), zap.Error(err))
		return nil, false
	}

	if shippedPages == 0 {
		logger.Warn("link suppression: site has NO shipped pages — outbound links go UNSUPPRESSED (expected during a first build, anomalous on an established site) (bugs_open/328)",
			zap.String("site_id", siteID.String()), zap.Int("refused_targets_ignored", len(refusedURLs)))
		return nil, false
	}

	return datahelpers.NewPageURLSet(refusedURLs), true
}

// suppressUnshippedOutboundLinks is the whole seam, in one call, so an outbound
// path adds one line rather than five. It returns the html unchanged whenever it
// is not enabled, cannot decide, or finds nothing.
//
// Fail-safe direction is the same as every pass around it: an unusable read
// suppresses NOTHING and says so durably. A writer acting on a failed read is
// how a bad artefact gets manufactured rather than repaired.
func suppressUnshippedOutboundLinks(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	domain, pageName, pageURL, html string,
	logger *zap.Logger,
) string {
	if html == "" || siteID == uuid.Nil {
		return html
	}
	if !configBoolOrDefault(params.StepConfig.Config, suppressUnshippedLinksKey, false) {
		return html
	}

	origin := linkRepairOrigin{
		AgentType:  params.AgentType,
		StepName:   suppressionStepName(params),
		ActionName: params.StepConfig.Action,
		PageName:   pageName,
		PageURL:    pageURL,
	}

	refused, ok := loadRefusedLinkTargets(ctx, params.DB, siteID, logger)
	if !ok {
		// A pod log line is not a record — the same argument repairOutboundPageLinks
		// makes for its own skip row. A page that shipped UNSUPPRESSED because the
		// lookup failed must be findable a day later.
		writeLinkRepairSkipLog(ctx, params, siteID.String(), domain, origin,
			"Outbound link suppression SKIPPED — refused-target set unavailable or site has not shipped; page deployed unsuppressed (bugs_open/328)",
			logger)
		return html
	}

	suppressed, changes := datahelpers.SuppressRefusedPageLinks(html, refused)
	if len(changes) == 0 {
		return html
	}

	unlinked, controlsDropped := countLinkSuppressions(changes)
	logger.Warn("outbound: suppressed link(s) to a page of this site that has never shipped (bugs_open/328)",
		zap.String("page", pageName),
		zap.String("domain", domain),
		zap.Int("unlinked", unlinked),
		zap.Int("controls_dropped", controlsDropped))
	writeLinkSuppressionLog(ctx, params, siteID.String(), domain, origin, changes, unlinked, controlsDropped, logger)
	return suppressed
}

// suppressionStepName is skipStepName's sibling, and it exists because that one
// falls back to the literal "rerender". This seam also runs on the ASSEMBLE
// path, where a row filed under "rerender" would send whoever reads it to the
// wrong workflow. Falls back to the running step, then to the action.
func suppressionStepName(params ActionParams) string {
	if params.ExecutionContext != nil && params.ExecutionContext.StepName != "" {
		return params.ExecutionContext.StepName
	}
	if params.CurrentStep != "" {
		return params.CurrentStep
	}
	return params.StepConfig.Action
}

// countLinkSuppressions splits the account by arm, mirroring countLinkRepairs.
func countLinkSuppressions(changes []datahelpers.LinkRepair) (unlinked, controlsDropped int) {
	for _, c := range changes {
		switch c.Action {
		case datahelpers.LinkRepairSuppress:
			unlinked++
		case datahelpers.LinkRepairDropControl:
			controlsDropped++
		}
	}
	return unlinked, controlsDropped
}

// writeLinkSuppressionLog persists what was suppressed, with the hrefs.
//
// This is the record that keeps the change honest. Suppression removes the
// anchor the post-deploy audit reads on the wire, so without a row of its own
// the only trace would be a pod log line that scrolls. Best-effort, like every
// recorder in this family: a logging failure must never fail a deploy whose HTML
// is already correct.
func writeLinkSuppressionLog(
	ctx context.Context,
	params ActionParams,
	siteIDStr, domain string,
	origin linkRepairOrigin,
	changes []datahelpers.LinkRepair,
	unlinked, controlsDropped int,
	logger *zap.Logger,
) {
	if params.DB == nil || len(changes) == 0 {
		return
	}

	entries := make([]map[string]string, 0, len(changes))
	for _, c := range changes {
		entries = append(entries, map[string]string{"href": c.Href, "action": c.Action})
	}

	LogActionEntry(ctx, params, agenterrors.Entry{
		SiteID:    siteIDStr,
		Domain:    domain,
		AgentType: origin.AgentType,
		StepName:  origin.StepName,
		Action:    origin.ActionName,
		ErrorMessage: fmt.Sprintf(
			"Suppressed %d outbound link(s) to never-shipped page(s) of this site: %d unlinked, %d control(s) dropped; the authored href is unchanged in content_data and returns when the target ships; see context.suppressed",
			len(changes), unlinked, controlsDropped),
		ErrorCode: linkSuppressionErrorCode,
		Severity:  "warning",
		Context: map[string]interface{}{
			"page":             origin.PageName,
			"page_url":         origin.PageURL,
			"suppressed":       entries,
			"unlinked":         unlinked,
			"controls_dropped": controlsDropped,
			"bug":              "bugs_open/328",
		},
	}, logger)
}
