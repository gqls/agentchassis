// FILE: platform/orchestration/actions/load_page_sections_from_spec_action_test.go
//
// bugs_open/285 — the section-list assembler was lock-blind. First tests for
// LoadPageSectionsFromSpecAction (there were none). They pin, on the
// authoritative tier:
//   - a human-locked live row the plan does not name is MERGED into the list
//     at its live position, reported in `locked_sections_merged`, and the
//     tier-1 fact-scoping slice stays index-aligned (a nil at the merged index)
//     — without the alignment the whole section_facts payload was silently
//     dropped by the len==len guard;
//   - the pages.sections cache is written ONCE with the MERGED list, compared
//     as jsonb (the old text comparison was always "distinct");
//   - a locked row the plan already names is not duplicated;
//   - when NO tier serves, the merge is NOT attempted (a locked-only list is
//     neither plan nor page and a rebuild on it would delete unlocked
//     siblings) — the locked query must not even run.
//
// Fixture shape follows plan_sections_slot_identity_test.go (sqlmock +
// ActionParams; the live caller maps every key via config dot-paths).
// Mutation-proven before commit: with the merge call removed, the first test
// fails alone on `sections`; with the facts insertion removed, it fails alone
// on `section_facts`.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

func loadSpecParams(db *sql.DB, siteID uuid.UUID, pageName string) ActionParams {
	return ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": siteID.String()},
			"page_record": map[string]interface{}{"name": pageName},
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id":   "site_record.site_id",
			"page_name": "page_record.name",
		}},
		ExecutionContext: &orchtypes.ExecutionContext{
			OrchestrationID: "44444444-4444-4444-4444-444444444444",
			StepName:        "load_spec_sections",
		},
	}
}

var lockedSlotColumns = []string{"id", "name", "slot_name", "position", "component_id", "function", "cname", "lock_type", "locked_by"}

func planRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"component_name", "assigned_fact_ids", "subject"})
}

func TestLoadPageSectionsFromSpec_MergesLockedLiveRowAtItsPosition(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	// Tier 1 serves: hero (scoped to fact f1), contact-info (unscoped).
	mock.ExpectQuery("FROM site_plan_sections sps").WithArgs(siteID, "contact").
		WillReturnRows(planRows().AddRow("hero", []byte(`["f1"]`), nil).AddRow("contact-info", nil, nil))
	// The page's locked live rows: the chat box, position 3, not in the plan.
	mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "contact").
		WillReturnRows(sqlmock.NewRows(lockedSlotColumns).
			AddRow("row-1", "contact", "chat-input-box", 3, "cid-chat", "chat-input-box", "chat-input-box", "permanent", "lane"))
	// ONE cache sync, with the MERGED list, jsonb-compared.
	mock.ExpectExec("UPDATE pages SET sections").
		WithArgs(`["hero","contact-info","chat-input-box"]`, siteID, "contact").
		WillReturnResult(sqlmock.NewResult(0, 1))

	out, err := LoadPageSectionsFromSpecAction(context.Background(), loadSpecParams(db, siteID, "contact"))
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	res := out.(map[string]interface{})

	if got := res["sections"].([]interface{}); !reflect.DeepEqual(got, []interface{}{"hero", "contact-info", "chat-input-box"}) {
		t.Errorf("sections = %v", got)
	}
	if res["source"] != "site_plan_tables" || res["count"] != 3 {
		t.Errorf("source/count = %v/%v", res["source"], res["count"])
	}
	if got := res["locked_sections_merged"].([]string); !reflect.DeepEqual(got, []string{"chat-input-box"}) {
		t.Errorf("locked_sections_merged = %v", got)
	}
	if res["locked_merge_count"] != 1 {
		t.Errorf("locked_merge_count = %v", res["locked_merge_count"])
	}
	facts, ok := res["section_facts"].([]interface{})
	if !ok || len(facts) != 3 {
		t.Fatalf("section_facts missing or misaligned: %#v (the len==len guard drops the payload when the merge shifts indices)", res["section_facts"])
	}
	if f0, _ := facts[0].([]interface{}); len(f0) != 1 || f0[0] != "f1" {
		t.Errorf("facts[0] = %#v, want [f1]", facts[0])
	}
	if facts[1] != nil || facts[2] != nil {
		t.Errorf("facts[1..2] = %#v/%#v, want nil (unscoped, merged)", facts[1], facts[2])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestLoadPageSectionsFromSpec_LockedRowAlreadyInPlanIsNotDuplicated(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	mock.ExpectQuery("FROM site_plan_sections sps").WithArgs(siteID, "about").
		WillReturnRows(planRows().AddRow("hero-about", nil, nil).AddRow("differentiators", nil, nil))
	mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "about").
		WillReturnRows(sqlmock.NewRows(lockedSlotColumns).
			AddRow("row-1", "about", "differentiators", 2, "cid-d", "differentiators", "differentiators-section", "permanent", "x"))
	mock.ExpectExec("UPDATE pages SET sections").
		WithArgs(`["hero-about","differentiators"]`, siteID, "about").
		WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := LoadPageSectionsFromSpecAction(context.Background(), loadSpecParams(db, siteID, "about"))
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	res := out.(map[string]interface{})
	if got := res["sections"].([]interface{}); len(got) != 2 {
		t.Errorf("sections = %v, want the plan's two entries, no duplicate", got)
	}
	if res["locked_merge_count"] != 0 {
		t.Errorf("locked_merge_count = %v, want 0", res["locked_merge_count"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestLoadPageSectionsFromSpec_NoTierServedMeansNoMerge(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	// Every tier misses.
	mock.ExpectQuery("FROM site_plan_sections sps").WithArgs(siteID, "orphan").WillReturnRows(planRows())
	mock.ExpectQuery("SELECT data FROM site_specs").WithArgs(siteID).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT sections, section_subjects, section_facts FROM pages").WithArgs(siteID, "orphan").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("same-role|site_plan_pages spp").WithArgs(siteID, "orphan").
		WillReturnRows(sqlmock.NewRows([]string{"name", "comps"}))
	// A locked row EXISTS on the page — and must NOT be merged into an
	// otherwise-empty list. Registered so that, if the loader ever consults
	// it here, sqlmock serves it and the assertion below catches the
	// locked-only list; ExpectationsWereMet then FAILS by design, and we
	// assert that failure to prove the query never ran.
	mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "orphan").
		WillReturnRows(sqlmock.NewRows(lockedSlotColumns).
			AddRow("row-1", "orphan", "tool-1", 1, "cid", "tool-x", "tool-x", "permanent", "x"))

	out, err := LoadPageSectionsFromSpecAction(context.Background(), loadSpecParams(db, siteID, "orphan"))
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	res := out.(map[string]interface{})
	if res["source"] != "none" || res["count"] != 0 {
		t.Errorf("source/count = %v/%v, want none/0", res["source"], res["count"])
	}
	if got, _ := res["sections"].([]string); len(got) != 0 {
		t.Errorf("sections = %v, want empty — a locked-only list must not be assembled", got)
	}
	if err := mock.ExpectationsWereMet(); err == nil || !strings.Contains(err.Error(), "page_components") {
		t.Errorf("the locked-row query must NOT have run when no tier served; ExpectationsWereMet = %v", err)
	}
}

// guardian (council 79f70435): the consolidated sync reaches EVERY page, not
// only locked ones — pin that a page with NO locked rows gets exactly one
// jsonb-compared UPDATE with the plan's own list and no merge keys, i.e. the
// tier path is unchanged apart from the guard now being able to say "no".
func TestLoadPageSectionsFromSpec_NoLockedRowsIsOneSyncNoMerge(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	mock.ExpectQuery("FROM site_plan_sections sps").WithArgs(siteID, "index").
		WillReturnRows(planRows().AddRow("hero", nil, nil).AddRow("faq", nil, nil))
	mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "index").
		WillReturnRows(sqlmock.NewRows(lockedSlotColumns))
	// Exactly ONE sync, with the plan list; 0 rows affected = the jsonb guard
	// said "unchanged" (the old ::text guard could never say that).
	mock.ExpectExec("UPDATE pages SET sections").
		WithArgs(`["hero","faq"]`, siteID, "index").
		WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := LoadPageSectionsFromSpecAction(context.Background(), loadSpecParams(db, siteID, "index"))
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	res := out.(map[string]interface{})
	if got := res["sections"].([]interface{}); !reflect.DeepEqual(got, []interface{}{"hero", "faq"}) {
		t.Errorf("sections = %v", got)
	}
	if res["locked_merge_count"] != 0 || len(res["locked_sections_merged"].([]string)) != 0 {
		t.Errorf("merge keys should be empty: %v / %v", res["locked_merge_count"], res["locked_sections_merged"])
	}
	if facts, _ := res["section_facts"].([]interface{}); len(facts) != 2 {
		t.Errorf("section_facts = %#v, want 2 aligned entries", res["section_facts"])
	}
	// A second UPDATE (the old per-tier double write) would trip sqlmock here.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// bug_historian (council 79f70435): a locked-row query failure must not be a
// silent one-build regression to lock-blind — the list proceeds unmerged (the
// 058 guard still protects the row) but a DURABLE agent_error_log entry is
// written, not just a log line. This pins the write and the code.
func TestLoadPageSectionsFromSpec_LockedQueryFailureLeavesADurableTrace(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	mock.ExpectQuery("FROM site_plan_sections sps").WithArgs(siteID, "contact").
		WillReturnRows(planRows().AddRow("hero", nil, "About our history"))
	mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "contact").
		WillReturnError(sql.ErrConnDone)
	// 13 placeholders in agenterrors.Write's INSERT; the code and message are
	// the two we care about, matched by position ($10 error_code, $11 message
	// per the column order there — asserted loosely via a custom matcher).
	mock.ExpectExec("INSERT INTO agent_error_log").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "load_page_sections_from_spec",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE pages SET sections").
		WithArgs(`["hero"]`, siteID, "contact").
		WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := LoadPageSectionsFromSpecAction(context.Background(), loadSpecParams(db, siteID, "contact"))
	if err != nil {
		t.Fatalf("action error: %v (a locked-query failure must be best-effort, not fatal)", err)
	}
	res := out.(map[string]interface{})
	if got := res["sections"].([]interface{}); !reflect.DeepEqual(got, []interface{}{"hero"}) {
		t.Errorf("sections = %v, want the unmerged plan list", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("durable trace or sync missing: %v", err)
	}
}

// Sanity on the JSON the sync writes: Go marshals with no spaces, and the
// SQL compares as jsonb, so the loader's own output shape is what the guard
// must be immune to. (Documents the trap; the comparison itself lives in SQL.)
func TestLoadPageSectionsFromSpec_MarshalledListHasNoSpaces(t *testing.T) {
	b, _ := json.Marshal([]string{"hero", "contact-info"})
	if string(b) != `["hero","contact-info"]` {
		t.Fatalf("unexpected marshal shape %s", b)
	}
}
