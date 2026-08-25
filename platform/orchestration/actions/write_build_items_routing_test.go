// FILE: platform/orchestration/actions/write_build_items_routing_test.go
//
// bugs_open/206 closure hardening — the SECOND producer door.
//
// reconcile_site_plan_routing_test.go pins that reconcile asks
// builderForPageType. This file pins that WriteBuildItemsAction does too,
// which until 2026-08-25 it did NOT: it carried its own closure-local copy of
// the same two maps, and the copy is what the 08-24 fix had to be kept
// byte-identical to (holding `section-index` out of the shared map for a day).
//
// [MEASURED 2026-08-25, live + archive, all history] this door had ZERO direct
// test coverage of any kind before this file — the only mentions of
// WriteBuildItemsAction in any _test.go were prose in two comments. Its
// per-page arm has also minted nothing since 2026-04-18, which is precisely
// why a silent routing regression here would go unnoticed: there is no traffic
// to notice it with. That makes the test the only instrument.
//
// Each test asserts the INSERT's own arguments, never the absence of a call —
// the package landmine about vacuously-passing negative assertions applies
// here exactly as it does at the reconcile door.
//
// > CORRECTED 2026-08-25 (same day, an adversarial review of this very file).
// > This header claimed "Every expectation below is mutation-proven: flip the
// > map entry in builder_routing.go and these fail." That was OVERBROAD, and
// > two mutations survived the whole suite:
// >   - flipping only the *itemType* half of a map entry failed NOTHING here
// >     (the constants wbiArgItemType/wbiArgHandlerAgent were declared and then
// >     never used in any pin — the assertion was intended and dropped);
// >   - changing the gap row's status from 'deferred' to 'triaged' failed
// >     nothing, which is the dangerous one: because this change made
// >     handler_agent EMPTY, writeWorkItem's registration probe (gated on
// >     handlerAgent != "", :1925) no longer runs, so `deferred` is now the ONLY
// >     thing keeping an empty-handler row out of both claim gates. Council
// >     round 2's argument was "deferred parks it, twice over"; this change
// >     removed the second wall and left the first unpinned.
// > Both are pinned below now, and both re-verified failing. The lesson for the
// > next editor: a claim that a test file is mutation-proven must name WHICH
// > mutations were run, or it reads as coverage nobody has.

package actions

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// writeBuildItemsFixture drives WriteBuildItemsAction over exactly ONE page.
// The trailing site-level items (composition/design/rerender) and the shared
// door probes inside writeWorkItem are registered permissively: they are noise
// for this question, and the assertion is the pinned argument on the per-page
// INSERT.
type writeBuildItemsFixture struct {
	pageName string
	pageType string
}

// The shared 16-arg work-item INSERT, by position (1-indexed in SQL):
// $1 site_id, $2 source, $3 pipeline, $4 item_type, $5 severity, $6 summary,
// $7 spec, $8 page_id, $9 component_id, $10 priority, $11 handler_agent,
// $12 status, $13 created_by, $14 item_key, $15 batch_id, $16 depends_on.
// driver.Value indexes are 0-based, so handler_agent is 10 and item_key 13.
const (
	wbiArgItemType     = 3
	wbiArgHandlerAgent = 10
	wbiArgStatus       = 11
	wbiArgItemKey      = 13
	wbiArgCount        = 16
)

func wbiInsertArgs(pin map[int]driver.Value) []driver.Value {
	args := make([]driver.Value, wbiArgCount)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	for idx, v := range pin {
		args[idx] = v
	}
	return args
}

// run drives the action and returns the handler_agent that the page item was
// actually inserted with, plus whether it was inserted at all.
//
// WHY THE ASSERTION IS SHAPED THIS WAY, and it matters — the obvious version of
// this test PASSES UNDER MUTATION and I wrote it first:
//
//   - WriteBuildItemsAction SWALLOWS a per-page insert error (`logger.Warn(...)
//     ; continue`), so a wrongly-routed page cannot be caught by asserting on
//     the action's returned error. It returns success either way.
//   - and a permissively-registered `ExpectExec("INSERT INTO site_work_items")`
//     will happily absorb the wrongly-routed INSERT, leaving the pinned
//     expectation merely unused — which sqlmock does not complain about unless
//     ExpectationsWereMet is called, and calling it here is not workable
//     because the shared door probes inside writeWorkItem fire a
//     case-dependent number of times.
//
// So: every INSERT expectation is pinned by item_key (nothing can absorb a call
// meant for another row), and the assertion is that the page's item_key was
// inserted WITH the expected handler — read from writeWorkItem's own
// "Work item inserted" log line, which it emits only after RowsAffected > 0.
// A mis-route therefore matches no expectation, errors, is swallowed, logs
// nothing, and the test fails on the absence. Mutation-proven in both
// directions.
func (f writeBuildItemsFixture) run(t *testing.T, wantItemKey string, pins map[int]driver.Value) (handler string, inserted bool) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	core, logs := observer.New(zapcore.InfoLevel)

	siteID := uuid.New()
	pageID := uuid.New()
	mock.MatchExpectationsInOrder(false)

	// 1. site lock check
	mock.ExpectQuery(regexp.QuoteMeta("SELECT locked_at IS NOT NULL FROM sites")).
		WillReturnRows(sqlmock.NewRows([]string{"locked"}).AddRow(false))

	// 2. queryPagesForBuild — one page, in the state a fresh plan leaves it.
	mock.ExpectQuery(regexp.QuoteMeta("FROM pages")).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "site_id", "name", "url", "title", "page_type", "status",
			"build_status", "sections", "nav_label", "nav_order", "in_header",
			"in_footer", "version", "meta_description", "content_direction",
		}).AddRow(
			pageID.String(), siteID.String(), f.pageName, "/"+f.pageName, f.pageName,
			// nav_order must be a real int, not NULL: scanPageRowsForBuild
			// scans it into an int and a NULL there fails the row, which
			// empties the page list and makes the per-page loop a no-op —
			// i.e. the fixture silently tests nothing. That is how the first
			// version of this file passed while asserting nothing at all.
			f.pageType, "active", "planned", []byte("[]"), nil, 0, false,
			false, 1, nil, nil,
		))

	mock.ExpectBegin()

	// 3. Every INSERT is pinned by its OWN item_key, so no expectation can
	// stand in for another. The page's is the one under test; the three
	// site-level items always follow and are registered so they cannot absorb
	// a page INSERT that went to the wrong handler.
	pageArgs := map[int]driver.Value{wbiArgItemKey: wantItemKey}
	for k, v := range pins {
		pageArgs[k] = v
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO site_work_items")).
		WithArgs(wbiInsertArgs(pageArgs)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	for _, key := range []string{"needs_composition", "needs_design", "needs_rerender"} {
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO site_work_items")).
			WithArgs(wbiInsertArgs(map[int]driver.Value{wbiArgItemKey: key})...).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	// 4. The shared door probes inside writeWorkItem, registered generously.
	// Over-registration is harmless precisely because ExpectationsWereMet is
	// deliberately NOT asserted — see the comment above.
	for i := 0; i < 8; i++ {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
			WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
		mock.ExpectQuery("SELECT EXISTS").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM site_work_items")).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	}
	mock.ExpectCommit()

	if _, err = WriteBuildItemsAction(context.Background(), ActionParams{
		DB:     db,
		Logger: zap.New(core),
		CollectedData: map[string]interface{}{
			"site_id": siteID.String(),
		},
		StepConfig:       models.Step{},
		ExecutionContext: &orchtypes.ExecutionContext{Action: "execute"},
	}); err != nil {
		t.Fatalf("WriteBuildItemsAction returned an error: %v", err)
	}

	for _, e := range logs.FilterMessage("Work item inserted").All() {
		fields := e.ContextMap()
		if k, _ := fields["item_key"].(string); k == wantItemKey {
			h, _ := fields["handler"].(string)
			return h, true
		}
	}
	return "", false
}

// A section-index page must reach directory-build-handler at THIS door too.
// Before 2026-08-25 it reached page-build-handler here, which has no
// layout-ensuring step and no-ops on a page whose layout is missing from every
// source ("no sections ready to build") — bugs_open/206's whole signature.
func TestWriteBuildItemsRoutesSectionIndexToTheLayoutEnsuringBuilder(t *testing.T) {
	f := writeBuildItemsFixture{pageName: "guides-index", pageType: "section-index"}
	handler, inserted := f.run(t, "needs_page:guides-index", map[int]driver.Value{
		wbiArgItemType:     "needs_directory",
		wbiArgHandlerAgent: "directory-build-handler",
		wbiArgStatus:       "triaged",
	})
	if !inserted {
		t.Fatal("no work item was inserted for the section-index page — the INSERT did not " +
			"match the pinned expectation, which means it carried a handler_agent other than " +
			"directory-build-handler")
	}
	if handler != "directory-build-handler" {
		t.Errorf("section-index routed to %q, want directory-build-handler — a page with no "+
			"layout in any source will no-op on the generic builder", handler)
	}
}

// The pre-existing entity-directory route must survive the extraction. This is
// the 2026-08-08 fix; it lived in the inline map that this change deleted.
func TestWriteBuildItemsRoutesEntityDirectory(t *testing.T) {
	f := writeBuildItemsFixture{pageName: "brand-directory-index", pageType: "entity-directory"}
	handler, inserted := f.run(t, "needs_page:brand-directory-index", map[int]driver.Value{
		wbiArgItemType:     "needs_directory",
		wbiArgHandlerAgent: "directory-build-handler",
		wbiArgStatus:       "triaged",
	})
	if !inserted {
		t.Fatal("no work item was inserted for the entity-directory page — the 2026-08-08 route " +
			"did not survive the extraction of the inline map")
	}
	if handler != "directory-build-handler" {
		t.Errorf("entity-directory routed to %q, want directory-build-handler", handler)
	}
}

// An ordinary content page must be unaffected. This is the no-regression arm:
// the extraction moved five routes, and this is the one carrying almost all
// real traffic.
func TestWriteBuildItemsLeavesContentPagesOnTheGenericBuilder(t *testing.T) {
	f := writeBuildItemsFixture{pageName: "about", pageType: "content"}
	handler, inserted := f.run(t, "needs_page:about", map[int]driver.Value{
		wbiArgItemType:     "needs_content_page",
		wbiArgHandlerAgent: "page-build-handler",
		wbiArgStatus:       "triaged",
	})
	if !inserted {
		t.Fatal("no work item was inserted for an ordinary content page — the extraction " +
			"changed a route it should not have")
	}
	if handler != "page-build-handler" {
		t.Errorf("content routed to %q, want page-build-handler", handler)
	}
}

// A type with a KNOWN-MISSING builder must file a deferred capability_gap with
// an EMPTY handler_agent, and must NOT mint a dispatch item.
//
// The empty handler is the part worth pinning. This arm used to write the
// needed builder's name — an agent that by construction does not exist — into
// handler_agent, which is bugs_closed/078's shape (an unroutable handler_agent
// livelocking the dispatcher) and which council round 2 ruled out at the
// sibling door on 2026-08-24. The two doors now agree. Note what the empty
// value also buys mechanically: writeWorkItem's registration probe is gated on
// handlerAgent != "", so a gap row cannot be demoted to `blocked` by a probe
// for an agent nobody ever intended to exist.
func TestWriteBuildItemsFilesCapabilityGapWithNoHandlerForUnbuildableType(t *testing.T) {
	f := writeBuildItemsFixture{pageName: "brand-profile", pageType: "entity-page"}
	handler, inserted := f.run(t, "capability_gap:entity-page:brand-profile", map[int]driver.Value{
		wbiArgItemType:     "capability_gap",
		wbiArgHandlerAgent: "",
		// status is load-bearing: see the header correction. With
		// handler_agent empty the registration probe never runs, so
		// `deferred` is the ONLY wall between this row and both claim gates.
		wbiArgStatus: "deferred",
	})
	if !inserted {
		t.Fatal("no capability_gap was filed for an entity-page — either it minted a dispatch " +
			"item instead, or the gap row carried a non-empty handler_agent")
	}
	if handler != "" {
		t.Errorf("capability_gap carried handler_agent %q, want empty — a deferred row naming an "+
			"agent that does not exist is bugs_closed/078's shape", handler)
	}
}

// The gap arm must ALSO not mint a needs_page dispatch item for the same page.
// Asserted separately from the row above because "filed a gap" and "did not
// also dispatch" are two different failures, and the second is the one that
// burns an attempt and parks the row in needs_human_review — dartsonline's
// brand-detail sat that way for 35 days.
func TestWriteBuildItemsFilesNoDispatchItemForUnbuildableType(t *testing.T) {
	f := writeBuildItemsFixture{pageName: "brand-profile", pageType: "entity-page"}
	if _, inserted := f.run(t, "needs_page:brand-profile", nil); inserted {
		t.Error("an entity-page minted a needs_page dispatch item — it must file only a " +
			"deferred capability_gap, or the item burns an attempt on a builder that " +
			"cannot build it and parks in needs_human_review")
	}
}
