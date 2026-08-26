// FILE: platform/orchestration/datahelpers/active_page_file_paths.go
//
// The ACTIVE-PAGE FILE-PATH SET — one implementation, two consumers with
// opposite jobs.
//
// THE RULE. Two page urls can designate one file. "/foo/" and "/foo/index.html"
// both derive `foo/index.html`, and the site serves ONE artefact at that path.
// So if an ARCHIVED page's derived file path is also derived by an ACTIVE page
// on the same site, the artefact at that path belongs to the live page. Two
// things follow, and they are the two consumers:
//
//   - RETRACTION MUST REFUSE that page. Deleting the file would delete the live
//     page's artefact. That is guard 3 of retract_page_deployment_action.go,
//     shipped with bugs_closed/098.
//   - DETECTION MUST SKIP that page. A 200 there is the LIVE page answering, not
//     the retired one — so flagging it is a false positive, and worse, it raises
//     a work item whose only stated remedy (retraction) refuses to run. An item
//     nobody can act on is the bugs_open/077 shape.
//
// WHY THIS LIVES HERE RATHER THAN BEING COPIED. It acquired its second consumer
// when bugs_open/359's `archived_page_still_serving` check was written, and the
// two must agree EXACTLY: the moment the detector's skip-set and the retraction's
// refuse-set differ, the detector files findings the remedy declines. Nothing at
// either call site would say so — the two packages never mention each other, and
// `discovery_checks` cannot import `actions` at all (the dependency runs the
// other way). check_site_structural_validity.go:117-119 states the estate's rule
// for exactly this moment: "If a third consumer of this rule ever appears, that
// is the signal to hoist it into datahelpers." One consumer earlier is cheaper
// than one later.
//
// ⚠ NO BUILD-AXIS ARM, AND THAT IS DELIBERATE — do not "tighten" it.
// The predicate is the lifecycle axis alone (PageWantedLivePredicateFor). An
// active page's file path is protected BEFORE it has ever deployed: retraction
// must not delete the path a live page is about to publish to, and detection must
// skip the same set or its findings stop being retractable. Adding
// PageHasShippedPredicateFor here would silently narrow both, and the retraction
// side would start deleting artefacts out from under pages mid-build.
//
// ⚠ THE WEAKER TEST IS `url` EQUALITY, AND IT IS NOT THIS. A census that compares
// `pages.url` strings finds no collisions on this estate today and would still
// miss the "/foo/" vs "/foo/index.html" pair, which is the only shape that
// matters here. Always derive.
package datahelpers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// ActivePageFilePaths maps every file path an ACTIVE page on this site derives,
// to that page's name. The derivation is the shared one
// (PageFilePathFromURL), so this answers the question that matters — "would
// deleting this file remove a live page's artefact?" — rather than the weaker
// url-equality question.
//
// A page whose url does not designate a file of its own (a fragment, a query, an
// off-origin url) contributes nothing: PageFilePathFromURL declines it rather
// than sanitising it, so it can never claim a path it does not own.
func ActivePageFilePaths(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string]string, error) {
	// The lifecycle arm is the shared helper; for the `=` direction the
	// COALESCE form it replaces was NULL-identical (both reject a NULL
	// status). Only the `<>` COMPLEMENT in loadRetractionCandidates differs
	// on NULL and deliberately keeps its COALESCE spelling.
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(name,''), COALESCE(url,'')
		  FROM pages
		 WHERE site_id = $1 AND `+PageWantedLivePredicateFor("")+``, siteID)
	if err != nil {
		return nil, fmt.Errorf("load active page paths: %w", err)
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var name, url string
		if err := rows.Scan(&name, &url); err != nil {
			return nil, fmt.Errorf("scan active page: %w", err)
		}
		if fp, ok := PageFilePathFromURL(url); ok {
			out[fp] = name
		}
	}
	return out, rows.Err()
}
