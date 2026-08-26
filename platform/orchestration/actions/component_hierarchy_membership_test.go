// Tests for the parent/child membership helpers (features_open/035 P1).
//
// The cycle and depth cases INDUCE their failure rather than asserting a happy
// path: the FK cannot forbid a row pointing at itself or at a descendant, so
// these guards are the only thing between a malformed chain and an infinite
// loop, and a guard proven only by a quiet test is not proven.
package actions

import (
	"context"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func hierarchyChildRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "parent_instance_id", "position", "slot_name"})
}

func TestHierarchyChildrenOfReturnsChildrenInRenderOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	parent := uuid.New()
	a, b := uuid.New().String(), uuid.New().String()
	mock.ExpectQuery("FROM page_components").
		WithArgs(parent).
		WillReturnRows(hierarchyChildRows().
			AddRow(a, parent.String(), 1, "insight-article.lead").
			AddRow(b, parent.String(), 2, "insight-article.quote"))

	kids, err := hierarchyChildrenOf(context.Background(), db, parent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(kids) != 2 {
		t.Fatalf("expected 2 children, got %d", len(kids))
	}
	if kids[0].SlotName != "insight-article.lead" || kids[1].SlotName != "insight-article.quote" {
		t.Errorf("children out of order: %+v", kids)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// A child dropped here becomes a slot the parent renders EMPTY while reporting
// success — the defect this feature exists to prevent, and the same shape
// bugs_open/410 found in loadStoredSections' own scan loop. So a scan failure
// must propagate, never `continue`.
func TestHierarchyChildrenOfPropagatesAScanFailureRatherThanDroppingTheChild(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	parent := uuid.New()
	mock.ExpectQuery("FROM page_components").
		WithArgs(parent).
		// position is NOT NULL in the schema, so a NULL here is the poison that
		// makes the scan fail. If someone ever makes it nullable, this test goes
		// GREEN FOR THE WRONG REASON — pick another non-nullable destination
		// rather than deleting it.
		WillReturnRows(hierarchyChildRows().
			AddRow(uuid.New().String(), parent.String(), nil, "insight-article.lead"))

	kids, err := hierarchyChildrenOf(context.Background(), db, parent)
	if err == nil {
		t.Fatalf("a failed child scan must be an error, not a silently shorter list (got %d children)", len(kids))
	}
	if !strings.Contains(err.Error(), "child row scan failed") {
		t.Errorf("error should name what happened, got: %v", err)
	}
}

func TestHierarchyAncestorChainIsNearestFirstAndStopsAtTopLevel(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	child, mid, root := uuid.New(), uuid.New(), uuid.New()
	col := []string{"parent_instance_id"}
	mock.ExpectQuery("SELECT parent_instance_id").WithArgs(child).
		WillReturnRows(sqlmock.NewRows(col).AddRow(mid.String()))
	mock.ExpectQuery("SELECT parent_instance_id").WithArgs(mid).
		WillReturnRows(sqlmock.NewRows(col).AddRow(root.String()))
	mock.ExpectQuery("SELECT parent_instance_id").WithArgs(root).
		WillReturnRows(sqlmock.NewRows(col).AddRow(nil)) // top-level: chain ends

	chain, err := hierarchyAncestorChain(context.Background(), db, child)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// NEAREST FIRST is load-bearing: a caller iterating this recomposes bottom-up,
	// so each level embeds an already-current child.
	if len(chain) != 2 || chain[0] != mid || chain[1] != root {
		t.Fatalf("chain = %v, want [mid root] nearest-first", chain)
	}
}

func TestHierarchyAncestorChainRefusesACycle(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	a, b := uuid.New(), uuid.New()
	col := []string{"parent_instance_id"}
	mock.ExpectQuery("SELECT parent_instance_id").WithArgs(a).
		WillReturnRows(sqlmock.NewRows(col).AddRow(b.String()))
	mock.ExpectQuery("SELECT parent_instance_id").WithArgs(b).
		WillReturnRows(sqlmock.NewRows(col).AddRow(a.String())) // back to a

	_, err = hierarchyAncestorChain(context.Background(), db, a)
	if err == nil {
		t.Fatal("a cycling parent chain must be refused — the FK cannot forbid one, so this guard " +
			"is all that stands between a malformed chain and an infinite loop")
	}
	if !strings.Contains(err.Error(), "cycles at") {
		t.Errorf("error should name the cycle, got: %v", err)
	}
}

func TestHierarchyAncestorChainRefusesAChainDeeperThanTheCap(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	ids := make([]uuid.UUID, hierarchyMaxDepth+2)
	for i := range ids {
		ids[i] = uuid.New()
	}
	col := []string{"parent_instance_id"}
	for i := 0; i < hierarchyMaxDepth; i++ {
		mock.ExpectQuery("SELECT parent_instance_id").WithArgs(ids[i]).
			WillReturnRows(sqlmock.NewRows(col).AddRow(ids[i+1].String()))
	}

	_, err = hierarchyAncestorChain(context.Background(), db, ids[0])
	if err == nil {
		t.Fatalf("a chain deeper than the cap of %d must be refused", hierarchyMaxDepth)
	}
	if !strings.Contains(err.Error(), "deeper than the cap") {
		t.Errorf("error should name the cap, got: %v", err)
	}
}

func TestHierarchyAncestorChainOfATopLevelRowIsEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	top := uuid.New()
	mock.ExpectQuery("SELECT parent_instance_id").WithArgs(top).
		WillReturnRows(sqlmock.NewRows([]string{"parent_instance_id"}).AddRow(""))

	chain, err := hierarchyAncestorChain(context.Background(), db, top)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain) != 0 {
		t.Errorf("a top-level row has no ancestors, got %v", chain)
	}
}
