// FILE: platform/orchestration/actions/retract_page_graph.go
//
// RETRACTION IS A GRAPH OPERATION, NOT A FILE DELETION (owner directive,
// 2026-08-03: "when we retract a page we should look through the whole site to
// change any links inward, nav links, and any places that it linked to but now
// noone links to").
//
// The directive is the same finding `bugs_open/098` reached from the other end.
// Its second surface records relojistas' `contacto`: the page was archived and
// its file deleted, so /contacto.html correctly 404'd — and every page on the
// site then advertised that 404 from its own footer, because the nav row was
// still `active`. 098 states the general form: "a fix that reaches only the
// deployed artefact will still leave the nav dead-linking". A retraction that
// only deletes a file trades one broken state for another.
//
// So a retraction has three obligations beyond removing the artefact, and they
// are deliberately NOT handled the same way, because they are not the same kind
// of thing:
//
//	INBOUND, editorial (a link in body copy or site chrome)
//	    -> REFUSE the retraction and name the referrers.
//	       Repairing prose is a content decision. The content-governance rule
//	       this platform already runs on is that auditors raise work items and
//	       never rewrite content (check_unverified_claims.go:40-45), and a
//	       retraction is an auditor. Refusing also makes the bad state
//	       unrepresentable rather than merely detected: you cannot retract a
//	       page that something still links to, so "dead link created by a
//	       retraction" stops being reachable.
//
//	INBOUND, structural (a site_nav_items row)
//	    -> DEACTIVATE it as part of the retraction, and report it.
//	       A nav row is a pointer, not prose: it carries no authored meaning
//	       that a human must re-word, and leaving it is precisely the
//	       relojistas defect. This is the mechanical half, so it is mechanised.
//	       Convention trap, from 098's own repair note: the four pre-existing
//	       non-live rows fleet-wide use 'inactive'. 'archived' is also excluded
//	       by the `= 'active'` filter every reader uses, but would be invisible
//	       to anyone querying by convention — so 'inactive' it is.
//
//	OUTBOUND, newly stranded (a page only the retracted page linked to)
//	    -> REPORT it. Not repaired here: what to do about a page nothing links
//	       to is a site-shaping decision (link it from somewhere, retract it
//	       too, or accept it), and the platform already has a standing detector
//	       that owns the question — discovery_checks/check_orphan_pages.go,
//	       which routes orphans to nav_drift / needs_internal_links. This
//	       reports them at retraction time so the operator sees the consequence
//	       in the same breath as the cause, rather than discovering it on the
//	       next sweep.
//
// THE INBOUND JUDGEMENT IS THE ORPHAN CHECK'S, DELIBERATELY. The three sources
// below — site_nav_items, site_components.rendered_html, other pages'
// page_components.rendered_html — are exactly the three
// check_orphan_pages.go:184-238 uses to decide "is this page reachable". Asking
// the same question with a different set of sources is how two checks come to
// disagree about the same site, so the source list is copied on purpose and
// this comment is the lockstep note: CHANGE ONE, READ THE OTHER.
//
// WHERE IT DIFFERS, AND WHY. The orphan check asks "does anything link here?"
// with a substring match (`LIKE '%' || p.url || '%'`). This asks the same
// question with a QUOTE-DELIMITED match (`href="<url>"`). The difference is not
// style: `/index.html` is a substring of `/learning-center/index.html`, so a
// substring match reports a referrer to the shorter url that does not exist.
// For the orphan check that over-match is the safe direction (it under-reports
// orphans). For a retraction it is not symmetric — it decides whether live copy
// is about to be left pointing at a deleted file — so the precise form is used,
// and its own failure direction is stated below.
package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// auditRetraction runs the whole graph audit for one retraction batch: what
// still points AT these pages, and what THEY point at that nothing else does.
// Callers act on the three lists differently — see the file header.
func auditRetraction(ctx context.Context, db *sql.DB, siteID uuid.UUID, pages []retractionCandidate, logger *zap.Logger) (*retractionGraph, error) {
	g := &retractionGraph{}
	if len(pages) == 0 {
		return g, nil
	}
	urls := make([]string, 0, len(pages))
	ids := make([]string, 0, len(pages))
	for _, p := range pages {
		if p.URL != "" {
			urls = append(urls, p.URL)
		}
		ids = append(ids, p.PageID)
	}

	editorial, nav, err := auditRetractionInbound(ctx, db, siteID, urls, ids, logger)
	if err != nil {
		return nil, err
	}
	g.Editorial, g.Nav = editorial, nav

	stranded, err := findStrandedTargets(ctx, db, siteID, ids, logger)
	if err != nil {
		return nil, err
	}
	g.Stranded = stranded
	return g, nil
}

// inboundRef is one live reference to a page that is about to be retracted.
type inboundRef struct {
	Source    string `json:"source"`             // "nav" | "chrome" | "page"
	OnPage    string `json:"on_page,omitempty"`  // page name, for source="page"
	Slot      string `json:"slot,omitempty"`     // slot/component, where known
	Label     string `json:"label,omitempty"`    // nav label, for source="nav"
	NavItem   string `json:"nav_item,omitempty"` // site_nav_items.id, for source="nav"
	TargetURL string `json:"target_url"`
}

// strandedTarget is a url the retracted page linked to whose last remaining
// referrer WAS the retracted page.
type strandedTarget struct {
	URL      string `json:"url"`
	PageName string `json:"page_name,omitempty"`
	Status   string `json:"status,omitempty"`
}

// retractionGraph is what the audit returns for one retraction batch.
type retractionGraph struct {
	Editorial []inboundRef     `json:"editorial_inbound"` // BLOCKING
	Nav       []inboundRef     `json:"nav_inbound"`       // deactivated
	Stranded  []strandedTarget `json:"stranded_targets"`  // reported
}

// auditRetractionInbound finds every live reference to the given urls, split
// into the editorial references that must block a retraction and the nav rows
// that the retraction will deactivate.
//
// PRECISION, and which way it fails. The match is `href="<url>"` — the form
// this platform's renderers actually emit (verified against the live
// robot-hands.com chrome and body components, 2026-08-03). It therefore MISSES
// a single-quoted href, a relative href ("about.html"), and an href assembled
// in JavaScript. That is a false-NEGATIVE direction: a missed referrer means a
// retraction proceeds and leaves one dead link, which is the pre-existing
// behaviour rather than a new harm. The alternative — a bare substring match —
// fails the other way and would refuse legitimate retractions for referrers
// that do not exist (`/index.html` is a substring of
// `/learning-center/index.html`). Neither is free; this one is stated.
func auditRetractionInbound(ctx context.Context, db *sql.DB, siteID uuid.UUID, urls []string, retractingPageIDs []string, logger *zap.Logger) ([]inboundRef, []inboundRef, error) {
	if len(urls) == 0 {
		return nil, nil, nil
	}

	var editorial, nav []inboundRef

	// 1. Body copy on OTHER pages. The retracted pages are excluded as
	//    referrers — a page linking to itself, or to a sibling in the same
	//    retraction batch, is not a reason to refuse.
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT p.name, COALESCE(pc.slot_name,''), u.url
		  FROM unnest($2::text[]) AS u(url)
		  JOIN page_components pc ON pc.rendered_html LIKE '%href="' || u.url || '"%'
		  JOIN pages p ON p.id = pc.page_id
		 WHERE p.site_id = $1
		   AND `+datahelpers.PageHasShippedPredicateFor("p")+`
		   AND COALESCE(p.status,'') = 'active'
		   AND NOT (p.id::text = ANY($3::text[]))
		 ORDER BY 1, 2`, siteID, datahelpers.PGTextArrayLiteral(urls), datahelpers.PGTextArrayLiteral(retractingPageIDs))
	if err != nil {
		return nil, nil, fmt.Errorf("inbound body-copy scan: %w", err)
	}
	for rows.Next() {
		var r inboundRef
		if err := rows.Scan(&r.OnPage, &r.Slot, &r.TargetURL); err != nil {
			rows.Close()
			return nil, nil, fmt.Errorf("scan inbound body ref: %w", err)
		}
		r.Source = "page"
		editorial = append(editorial, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// 2. Site chrome (header/footer). Editorial for the same reason: the
	//    chrome is authored content, and it is the surface that made the
	//    relojistas case visible on EVERY page at once.
	rows, err = db.QueryContext(ctx, `
		SELECT DISTINCT COALESCE(sc.slot_name, sc.id::text), u.url
		  FROM unnest($2::text[]) AS u(url)
		  JOIN site_components sc ON sc.rendered_html LIKE '%href="' || u.url || '"%'
		 WHERE sc.site_id = $1
		 ORDER BY 1`, siteID, datahelpers.PGTextArrayLiteral(urls))
	if err != nil {
		return nil, nil, fmt.Errorf("inbound chrome scan: %w", err)
	}
	for rows.Next() {
		var r inboundRef
		if err := rows.Scan(&r.Slot, &r.TargetURL); err != nil {
			rows.Close()
			return nil, nil, fmt.Errorf("scan inbound chrome ref: %w", err)
		}
		r.Source = "chrome"
		editorial = append(editorial, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// 3. Nav rows — by url OR by page_id. Both arms are needed: 098's own
	//    fleet-wide measurement found four live rows with page_id NULL that are
	//    authored by url alone, and a url-only check would miss an FK-authored
	//    row just as an FK-only check missed those four.
	rows, err = db.QueryContext(ctx, `
		SELECT DISTINCT i.id::text, COALESCE(i.label,''), COALESCE(i.url,'')
		  FROM site_nav_items i
		  JOIN site_nav_groups g ON g.id = i.group_id
		 WHERE g.site_id = $1
		   AND i.status = 'active'
		   AND (i.url = ANY($2::text[]) OR i.page_id::text = ANY($3::text[]))
		 ORDER BY 2`, siteID, datahelpers.PGTextArrayLiteral(urls), datahelpers.PGTextArrayLiteral(retractingPageIDs))
	if err != nil {
		return nil, nil, fmt.Errorf("inbound nav scan: %w", err)
	}
	for rows.Next() {
		var r inboundRef
		if err := rows.Scan(&r.NavItem, &r.Label, &r.TargetURL); err != nil {
			rows.Close()
			return nil, nil, fmt.Errorf("scan inbound nav ref: %w", err)
		}
		r.Source = "nav"
		nav = append(nav, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	logger.Info("retraction inbound audit",
		zap.Int("editorial_refs", len(editorial)),
		zap.Int("nav_refs", len(nav)),
		zap.Int("urls", len(urls)))
	return editorial, nav, nil
}

// findStrandedTargets returns pages that the retracted pages link to and that
// nothing else on the site will link to once those pages are gone.
//
// This is the orphan check's question asked one step early — "what BECOMES
// unreachable", rather than "what IS unreachable" — so it cannot simply call
// check_orphan_pages: that check reads the world as it stands, in which the
// page being retracted is still a referrer. Hence the same three sources with
// the retracted pages subtracted from the referrer set.
func findStrandedTargets(ctx context.Context, db *sql.DB, siteID uuid.UUID, retractingPageIDs []string, logger *zap.Logger) ([]strandedTarget, error) {
	if len(retractingPageIDs) == 0 {
		return nil, nil
	}

	rows, err := db.QueryContext(ctx, `
		WITH outbound AS (
		    -- every shipped page this site holds, that a retracting page links to
		    SELECT DISTINCT t.id, t.name, t.url, COALESCE(t.status,'') AS status
		      FROM pages t
		      JOIN page_components pc ON pc.rendered_html LIKE '%href="' || t.url || '"%'
		     WHERE t.site_id = $1
		       AND t.url IS NOT NULL AND t.url <> ''
		       AND pc.page_id::text = ANY($2::text[])
		       AND NOT (t.id::text = ANY($2::text[]))
		)
		SELECT o.name, o.url, o.status
		  FROM outbound o
		 WHERE NOT EXISTS (            -- no OTHER active page links to it
		           SELECT 1 FROM page_components pc2
		             JOIN pages p2 ON p2.id = pc2.page_id
		            WHERE p2.site_id = $1
		              AND COALESCE(p2.status,'') = 'active'
		              AND NOT (p2.id::text = ANY($2::text[]))
		              AND pc2.rendered_html LIKE '%href="' || o.url || '"%')
		   AND NOT EXISTS (            -- not in the chrome
		           SELECT 1 FROM site_components sc
		            WHERE sc.site_id = $1
		              AND sc.rendered_html LIKE '%href="' || o.url || '"%')
		   AND NOT EXISTS (            -- not in the nav
		           SELECT 1 FROM site_nav_items i
		             JOIN site_nav_groups g ON g.id = i.group_id
		            WHERE g.site_id = $1 AND i.status = 'active'
		              AND (i.url = o.url OR i.page_id = o.id))
		 ORDER BY o.url`, siteID, datahelpers.PGTextArrayLiteral(retractingPageIDs))
	if err != nil {
		return nil, fmt.Errorf("stranded-target scan: %w", err)
	}
	defer rows.Close()

	var out []strandedTarget
	for rows.Next() {
		var s strandedTarget
		if err := rows.Scan(&s.PageName, &s.URL, &s.Status); err != nil {
			return nil, fmt.Errorf("scan stranded target: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	logger.Info("retraction stranded-target audit", zap.Int("stranded", len(out)))
	return out, nil
}

// deactivateNavItems retires the nav rows that pointed at a retracted page.
//
// 'inactive', not 'archived': both are excluded by the `status = 'active'`
// filter every nav reader uses, but the four pre-existing non-live rows
// fleet-wide are 'inactive', and a second spelling is invisible to anyone
// querying by convention (098's repair note records exactly this trap).
func deactivateNavItems(ctx context.Context, db *sql.DB, refs []inboundRef, logger *zap.Logger) (int, error) {
	if len(refs) == 0 {
		return 0, nil
	}
	ids := make([]string, 0, len(refs))
	for _, r := range refs {
		ids = append(ids, r.NavItem)
	}
	res, err := db.ExecContext(ctx, `
		UPDATE site_nav_items
		   SET status = 'inactive', updated_at = NOW()
		 WHERE id::text = ANY($1::text[]) AND status = 'active'`,
		datahelpers.PGTextArrayLiteral(ids))
	if err != nil {
		return 0, fmt.Errorf("deactivate nav items: %w", err)
	}
	n, _ := res.RowsAffected()
	logger.Info("retraction deactivated nav rows",
		zap.Int64("rows", n), zap.Int("requested", len(ids)))
	return int(n), nil
}
