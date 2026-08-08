// FILE: platform/orchestration/actions/discovery_checks/check_phantom_internal_links_verifier_test.go
//
// Tests for VerifyUnbuiltInternalLinkResolved — bugs_open/220's completion guard.
//
// The load-bearing case is the REFUSAL: the bug is a dispatch defect that
// rebuilds the page CONTAINING the link and reports success, so a verifier only
// earns its place if it says NO when the target page is still unbuilt — that is
// the exact live shape of items 8abc9f8d/3f066b90 (container deployed, target
// still deployed_at NULL, item stamped 'complete' at attempt_count 0).
//
// The container/target split runs through every case: spec.page_id is the page
// CONTAINING the link; target.PageID (the work item's page_id column) is the
// TARGET the href points at. A test that conflated them would pass against an
// implementation with this bug's own confusion in it.

package discovery_checks

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func unbuiltLinkVerifyTarget() VerifyTarget {
	targetPageID := uuid.New()
	return VerifyTarget{
		ItemID:   uuid.New(),
		SiteID:   uuid.New(),
		PageID:   &targetPageID, // the TARGET (never-deployed) page — the column the check files
		ItemType: "unbuilt_internal_link",
		Spec: map[string]interface{}{
			"surface":   "page_component",
			"page_id":   uuid.New().String(), // the CONTAINER page
			"page_name": "index",
			"href":      "/directory/index.html",
		},
	}
}

// THE LOAD-BEARING CASE: href still rendered on the container, target still
// never-deployed — completion must be refused. Under the pre-fix routing this is
// what every "successful" dispatch of this item type produced.
func TestVerifyUnbuiltInternalLinkResolved_RefusesWhileTargetUnbuilt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mock.ExpectQuery("FROM page_components").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"never_deployed"}).AddRow(true))

	res, err := VerifyUnbuiltInternalLinkResolved(context.Background(), db, unbuiltLinkVerifyTarget(), zap.NewNop())
	if err != nil {
		t.Fatalf("verifier returned error: %v", err)
	}
	if res.Resolved {
		t.Fatalf("Resolved=true while the target has never been deployed and the link is still live — "+
			"this is exactly the green-loop stamp bugs_open/220 exists to prevent. Detail: %q", res.Detail)
	}
	if !strings.Contains(res.Detail, "never been deployed") {
		t.Errorf("Detail should state the target is still unbuilt so a human can act; got %q", res.Detail)
	}
}

// Target shipped, link still rendered: the defect the item describes is gone.
func TestVerifyUnbuiltInternalLinkResolved_PassesWhenTargetShipped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mock.ExpectQuery("FROM page_components").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"never_deployed"}).AddRow(false))

	res, err := VerifyUnbuiltInternalLinkResolved(context.Background(), db, unbuiltLinkVerifyTarget(), zap.NewNop())
	if err != nil {
		t.Fatalf("verifier returned error: %v", err)
	}
	if !res.Resolved {
		t.Fatalf("Resolved=false after the target deployed — a correctly-repaired item would burn its "+
			"attempts and strand in 'failed'. Detail: %q", res.Detail)
	}
}

// Link removed from the container: resolved by the item's own alternative remedy,
// whatever the target's state — the dead_fragment_link both-disjuncts precedent.
func TestVerifyUnbuiltInternalLinkResolved_PassesWhenLinkRemoved(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mock.ExpectQuery("FROM page_components").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	res, err := VerifyUnbuiltInternalLinkResolved(context.Background(), db, unbuiltLinkVerifyTarget(), zap.NewNop())
	if err != nil {
		t.Fatalf("verifier returned error: %v", err)
	}
	if !res.Resolved {
		t.Fatalf("Resolved=false with the href gone from the container — the item's own fix text names "+
			"link-removal as a remedy, and refusing it strands a fixed item. Detail: %q", res.Detail)
	}
}

// Target row deleted/archived while the link still renders: the href dangles with
// nothing left to build. Resolved:false and NOT an error — this is a judged
// defect (the phantom arm re-classifies it next pass), not a failed verification.
func TestVerifyUnbuiltInternalLinkResolved_RefusesWhenTargetRowGone(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mock.ExpectQuery("FROM page_components").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"never_deployed"})) // zero rows

	res, err := VerifyUnbuiltInternalLinkResolved(context.Background(), db, unbuiltLinkVerifyTarget(), zap.NewNop())
	if err != nil {
		t.Fatalf("must not error on a missing target row — since RFC_017 an error BLOCKS completion and "+
			"retries, but this state is a settled judgement, not an inability to check. Got: %v", err)
	}
	if res.Resolved {
		t.Fatalf("Resolved=true for a live link whose target row is gone — that certifies a dangling 404 "+
			"as repaired. Detail: %q", res.Detail)
	}
}

// A site_component item consults the site chrome, scoped by SiteID — it carries
// no container page.
func TestVerifyUnbuiltInternalLinkResolved_SiteSurfaceUsesChrome(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	target := unbuiltLinkVerifyTarget()
	target.Spec["surface"] = "site_component"
	target.Spec["page_id"] = "" // no container page for chrome findings

	mock.ExpectQuery("FROM site_components").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	res, err := VerifyUnbuiltInternalLinkResolved(context.Background(), db, target, zap.NewNop())
	if err != nil {
		t.Fatalf("verifier returned error: %v", err)
	}
	if !res.Resolved {
		t.Fatalf("Resolved=false with the href gone from every site component. Detail: %q", res.Detail)
	}
}

// No target page_id: the one thing this item type exists to name is missing, so
// the verifier must refuse to run (error → fail-closed under RFC_017) rather
// than certify or scan the wrong page.
func TestVerifyUnbuiltInternalLinkResolved_ErrorsWithoutTargetPageID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mock.ExpectQuery("FROM page_components").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	target := unbuiltLinkVerifyTarget()
	target.PageID = nil

	if _, err := VerifyUnbuiltInternalLinkResolved(context.Background(), db, target, zap.NewNop()); err == nil {
		t.Fatal("expected an error when the item carries no target page_id")
	}
}
