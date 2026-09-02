// FILE: platform/orchestration/actions/load_page_sections_fallback_subjects_test.go
//
// bugs_open/443 — per-section subjects (and facts) for pages the plan tables do
// not serve. The change is a set of GUARDS, so each test pins one guard arm and
// each was mutation-proven before commit (the mutation named beside it):
//   - tier 3, row-served list: aligned arrays attach, normalised to tier-1
//     shape (trimmed subjects, ""→nil; facts arrays or nil).
//   - tier 3, misaligned array: ignored, independently per array — weaken the
//     length check and the misaligned test fails alone.
//   - tier 3, collected_data-served list: arrays attach ONLY when the row's
//     stored sections CONTENT-equal the served list — replace the equality
//     check with a length check and the same-length-different-content test
//     fails.
//   - tier 2: same-object sibling arrays, RAW-index aligned across skipped
//     non-string entries (a skipped entry drops its subject WITH it).
//   - tier 4 (sibling synthesis): arrays are never emitted.
//   - LOCK-008 merge on tier 3: nil-inserted at merged indices, and the STORED
//     array is re-aligned in the same run (the origin-vs-destination lesson:
//     without the realign the next build reads a stale-length array and drops
//     it silently).
// Plus the pure half of the build-side detector (repeatedComponentSubjectGaps):
// negatives are asserted on the pure function, not on mock silence.
//
// Fixture shape follows load_page_sections_from_spec_action_test.go.

package actions

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// tier3Row serves the pages-row query (sections + the two scoping arrays).
func tier3Row(sections, subjects, facts interface{}) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"sections", "section_subjects", "section_facts"})
	toB := func(v interface{}) interface{} {
		if v == nil {
			return nil
		}
		return []byte(v.(string))
	}
	return r.AddRow(toB(sections), toB(subjects), toB(facts))
}

func missTier1And2(mock sqlmock.Sqlmock, siteID uuid.UUID, page string) {
	mock.ExpectQuery("FROM site_plan_sections sps").WithArgs(siteID, page).WillReturnRows(planRows())
	mock.ExpectQuery("SELECT data FROM site_specs").WithArgs(siteID).WillReturnError(sql.ErrNoRows)
}

func TestLoadPageSections_Tier3RowServesAlignedArrays(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	missTier1And2(mock, siteID, "playground")
	mock.ExpectQuery("SELECT sections, section_subjects, section_facts FROM pages").
		WithArgs(siteID, "playground").
		WillReturnRows(tier3Row(
			`["hero","generic-text-block","generic-text-block"]`,
			`["  The hour itself  ","What you actually do",""]`,
			`[null,["f1"],null]`))
	mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "playground").
		WillReturnRows(sqlmock.NewRows(lockedSlotColumns))
	mock.ExpectExec("UPDATE pages SET sections").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := LoadPageSectionsFromSpecAction(context.Background(), loadSpecParams(db, siteID, "playground"))
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	res := out.(map[string]interface{})
	if res["source"] != "pages_table" {
		t.Fatalf("source = %v", res["source"])
	}
	subj, ok := res["section_subjects"].([]interface{})
	if !ok {
		t.Fatalf("section_subjects missing: %#v", res)
	}
	// Normalised to tier-1 shape: trimmed, "" -> nil.
	if !reflect.DeepEqual(subj, []interface{}{"The hour itself", "What you actually do", nil}) {
		t.Errorf("section_subjects = %#v", subj)
	}
	facts, ok := res["section_facts"].([]interface{})
	if !ok || len(facts) != 3 {
		t.Fatalf("section_facts missing/misaligned: %#v", res["section_facts"])
	}
	if facts[0] != nil || facts[2] != nil {
		t.Errorf("facts[0]/[2] = %#v/%#v, want nil (unscoped)", facts[0], facts[2])
	}
	if f1, _ := facts[1].([]interface{}); len(f1) != 1 || f1[0] != "f1" {
		t.Errorf("facts[1] = %#v, want [f1]", facts[1])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestLoadPageSections_Tier3MisalignedArrayIgnoredIndependently(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	// subjects: 2 entries for 3 sections (a writer replaced sections without
	// re-aligning) -> ignored. facts: aligned -> still attach. The guards must
	// be independent or one stale array silently disarms the other.
	missTier1And2(mock, siteID, "playground")
	mock.ExpectQuery("SELECT sections, section_subjects, section_facts FROM pages").
		WithArgs(siteID, "playground").
		WillReturnRows(tier3Row(
			`["hero","generic-text-block","generic-text-block"]`,
			`["stale","stale"]`,
			`[null,null,["f9"]]`))
	mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "playground").
		WillReturnRows(sqlmock.NewRows(lockedSlotColumns))
	mock.ExpectExec("UPDATE pages SET sections").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := LoadPageSectionsFromSpecAction(context.Background(), loadSpecParams(db, siteID, "playground"))
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	res := out.(map[string]interface{})
	if _, present := res["section_subjects"]; present {
		t.Errorf("misaligned section_subjects must be IGNORED, got %#v", res["section_subjects"])
	}
	if facts, ok := res["section_facts"].([]interface{}); !ok || len(facts) != 3 {
		t.Errorf("aligned section_facts must still attach: %#v", res["section_facts"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func loadSpecParamsWithFallback(db *sql.DB, siteID uuid.UUID, pageName string, sections []interface{}) ActionParams {
	p := loadSpecParams(db, siteID, pageName)
	p.CollectedData["page_record"].(map[string]interface{})["sections"] = sections
	p.StepConfig.Config["page_sections_fallback"] = "page_record.sections"
	return p
}

func TestLoadPageSections_Tier3CollectedListMustContentEqualRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	// collected_data serves ["hero","faq"]; the row says ["hero","cta"] — SAME
	// LENGTH, different content. The row's subjects are aligned to the ROW's
	// list, so applying them to the served list would put "Why trust us" on
	// the faq. Equality, not length, is the test — this test FAILS if
	// stringSlicesEqual is weakened to a length comparison.
	missTier1And2(mock, siteID, "about")
	mock.ExpectQuery("SELECT sections, section_subjects, section_facts FROM pages").
		WithArgs(siteID, "about").
		WillReturnRows(tier3Row(`["hero","cta"]`, `["Who we are","Why trust us"]`, nil))
	mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "about").
		WillReturnRows(sqlmock.NewRows(lockedSlotColumns))
	mock.ExpectExec("UPDATE pages SET sections").WillReturnResult(sqlmock.NewResult(0, 1))

	out, err := LoadPageSectionsFromSpecAction(context.Background(),
		loadSpecParamsWithFallback(db, siteID, "about", []interface{}{"hero", "faq"}))
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	res := out.(map[string]interface{})
	if res["source"] != "pages_table" {
		t.Fatalf("source = %v", res["source"])
	}
	if _, present := res["section_subjects"]; present {
		t.Errorf("subjects aligned to a DIFFERENT list must not attach: %#v", res["section_subjects"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestLoadPageSections_Tier3CollectedListEqualRowAttaches(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	missTier1And2(mock, siteID, "about")
	mock.ExpectQuery("SELECT sections, section_subjects, section_facts FROM pages").
		WithArgs(siteID, "about").
		WillReturnRows(tier3Row(`["hero","faq"]`, `["Who we are","The questions we hear"]`, nil))
	mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "about").
		WillReturnRows(sqlmock.NewRows(lockedSlotColumns))
	mock.ExpectExec("UPDATE pages SET sections").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := LoadPageSectionsFromSpecAction(context.Background(),
		loadSpecParamsWithFallback(db, siteID, "about", []interface{}{"hero", "faq"}))
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	res := out.(map[string]interface{})
	subj, ok := res["section_subjects"].([]interface{})
	if !ok || !reflect.DeepEqual(subj, []interface{}{"Who we are", "The questions we hear"}) {
		t.Errorf("section_subjects = %#v", res["section_subjects"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestLoadPageSections_Tier2AspectArraysAlignByRawIndexAcrossSkips(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	// The aspect page object carries a non-string sections entry (index 1).
	// The walk skips it, and its subject must be dropped WITH it — RAW-index
	// lookups before skips, appended in the same branch as the name.
	aspect := `{"pages":[{"name":"index",
		"sections":["hero",7,"generic-text-block","generic-text-block"],
		"section_subjects":["The opening",null,"One thing","Another thing"],
		"section_facts":null}]}`
	mock.ExpectQuery("FROM site_plan_sections sps").WithArgs(siteID, "index").WillReturnRows(planRows())
	mock.ExpectQuery("SELECT data FROM site_specs").WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(aspect)))
	mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "index").
		WillReturnRows(sqlmock.NewRows(lockedSlotColumns))
	mock.ExpectExec("UPDATE pages SET sections").WillReturnResult(sqlmock.NewResult(0, 1))

	out, err := LoadPageSectionsFromSpecAction(context.Background(), loadSpecParams(db, siteID, "index"))
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	res := out.(map[string]interface{})
	if res["source"] != "site_specs" {
		t.Fatalf("source = %v", res["source"])
	}
	if got := res["sections"].([]interface{}); !reflect.DeepEqual(got, []interface{}{"hero", "generic-text-block", "generic-text-block"}) {
		t.Fatalf("sections = %#v", got)
	}
	subj, ok := res["section_subjects"].([]interface{})
	if !ok || !reflect.DeepEqual(subj, []interface{}{"The opening", "One thing", "Another thing"}) {
		t.Errorf("section_subjects = %#v — a skipped entry must drop its subject with it, never shift", res["section_subjects"])
	}
	if _, present := res["section_facts"]; present {
		t.Errorf("null section_facts must stay absent, got %#v", res["section_facts"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestLoadPageSections_Tier4SiblingNeverEmitsArrays(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	// Tiers 1-3 miss; a same-role sibling lends its layout, repeats and all.
	// A borrowed skeleton's subjects would be ANOTHER page's — the arrays must
	// never be emitted here, whatever the sibling's plan rows carry.
	missTier1And2(mock, siteID, "new-guide")
	mock.ExpectQuery("SELECT sections, section_subjects, section_facts FROM pages").
		WithArgs(siteID, "new-guide").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("site_plan_pages spp").WithArgs(siteID, "new-guide").
		WillReturnRows(sqlmock.NewRows([]string{"name", "comps"}).
			AddRow("old-guide", []byte(`["hero","generic-text-block","generic-text-block"]`)))
	mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "new-guide").
		WillReturnRows(sqlmock.NewRows(lockedSlotColumns))
	mock.ExpectExec("UPDATE pages SET sections").WillReturnResult(sqlmock.NewResult(0, 1))

	out, err := LoadPageSectionsFromSpecAction(context.Background(), loadSpecParams(db, siteID, "new-guide"))
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	res := out.(map[string]interface{})
	if res["source"] != "same_role_sibling" {
		t.Fatalf("source = %v", res["source"])
	}
	if _, present := res["section_subjects"]; present {
		t.Errorf("tier 4 must never emit section_subjects")
	}
	if _, present := res["section_facts"]; present {
		t.Errorf("tier 4 must never emit section_facts")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestLoadPageSections_Tier3LockedMergeNilInsertsAndRealignsStorage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	missTier1And2(mock, siteID, "contact")
	mock.ExpectQuery("SELECT sections, section_subjects, section_facts FROM pages").
		WithArgs(siteID, "contact").
		WillReturnRows(tier3Row(`["hero","generic-text-block"]`, `["Opening","The details"]`, nil))
	// A human-locked live row the list does not name, position 3.
	mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "contact").
		WillReturnRows(sqlmock.NewRows(lockedSlotColumns).
			AddRow("row-1", "contact", "chat-input-box", 3, "cid-chat", "chat-input-box", "chat-input-box", "permanent", "lane"))
	// ONE cache sync with the MERGED list…
	mock.ExpectExec("UPDATE pages SET sections").
		WithArgs(`["hero","generic-text-block","chat-input-box"]`, siteID, "contact").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// …then the STORED subjects array is re-aligned (nil at the merged index).
	// Without this the next build reads a 2-entry array against a 3-entry list
	// and silently drops it — the fix must be measured at the destination.
	mock.ExpectExec("UPDATE pages SET section_subjects").
		WithArgs(`["Opening","The details",null]`, siteID, "contact").
		WillReturnResult(sqlmock.NewResult(0, 1))

	out, err := LoadPageSectionsFromSpecAction(context.Background(), loadSpecParams(db, siteID, "contact"))
	if err != nil {
		t.Fatalf("action error: %v", err)
	}
	res := out.(map[string]interface{})
	subj, ok := res["section_subjects"].([]interface{})
	if !ok || !reflect.DeepEqual(subj, []interface{}{"Opening", "The details", nil}) {
		t.Errorf("section_subjects = %#v — merge must nil-insert at the merged index", res["section_subjects"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations (including the storage realign): %v", err)
	}
}

// ── the pure half of the build-side detector ────────────────────────────────

func TestRepeatedComponentSubjectGaps(t *testing.T) {
	cases := []struct {
		name     string
		sections []string
		subjects []string
		want     []repeatSubjectGap
	}{
		{"repeats with no subjects fire",
			[]string{"hero", "generic-text-block", "generic-text-block", "generic-text-block"},
			nil,
			[]repeatSubjectGap{{"generic-text-block", 3, 3}}},
		{"non-adjacent repeats fire (443 our-position-on-ai)",
			[]string{"generic-text-block", "features", "generic-text-block"},
			nil,
			[]repeatSubjectGap{{"generic-text-block", 2, 2}}},
		{"one subjectless instance among subjected repeats still fires",
			[]string{"g", "g", "g"},
			[]string{"a", "", "c"},
			[]repeatSubjectGap{{"g", 3, 1}}},
		{"subjects covering every repeat stay quiet",
			[]string{"g", "g"},
			[]string{"one thing", "another"},
			nil},
		{"singletons never fire however subjectless",
			[]string{"hero", "faq", "cta"},
			nil,
			nil},
		{"whitespace-only subject counts as missing",
			[]string{"g", "g"},
			[]string{"real", "   "},
			[]repeatSubjectGap{{"g", 2, 1}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := repeatedComponentSubjectGaps(tc.sections, tc.subjects)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("gaps = %#v, want %#v", got, tc.want)
			}
		})
	}
}

// The pure function above could pass every table arm while sitting uncalled (a
// helper with no callers is not a detector) — this arm proves the ACTION files
// the durable row, before any planning work, when a repeated type has no
// subject.
func TestPlanSections_RepeatedSubjectlessComponentFilesDurableTrace(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()
	heroID, gtbID := uuid.New().String(), uuid.New().String()

	// The detector's durable trace comes FIRST — before component loading —
	// so a failure later in the build cannot unrecord the finding.
	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRow(heroID, "hero", "hero", "<section>h</section>", "section").
			AddRow(gtbID, "generic-text-block", "generic-text-block", "", nil, nil, nil, "<section>g</section>", planSectionsSchema, "template", nil, "section"))
	mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "playground").
		WillReturnRows(slotRows())
	mock.ExpectQuery("FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"aspect", "data"}))
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"section_name", "summary"}))
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	params := planParams(db, siteID, "playground",
		[]string{"hero", "generic-text-block", "generic-text-block"})

	if _, err := PlanSectionsAction(context.Background(), params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations (including the agent_error_log insert): %v", err)
	}
}
