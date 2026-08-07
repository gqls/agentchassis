// FILE: platform/orchestration/actions/fact_scoping_151_test.go
//
// bugs_open/151 candidate 1 — facts assigned to sections at plan time. The
// mechanism spans four seams and each has a way to fail silently, so each is
// pinned here:
//
//   - extractSectionEntries: fact assignments are looked up by RAW index in
//     the page's section_facts array BEFORE malformed entries are skipped —
//     a skip that shifted alignment would put another section's facts on the
//     wrong section, which is worse than the duplication being fixed;
//   - ValidateSitePlanAction's normalise pass: object-form entries leave the
//     validator as plain string sections + an aligned section_facts array,
//     because sync_pages_to_db serialises the raw sections array into
//     pages.sections whose every reader expects strings;
//   - composeScopedWriterBlock: the per-section block carries ONLY assigned
//     facts (values current at compose time), an unknown ID is inert, and an
//     empty assignment composes an empty block rather than falling back to
//     the site-wide one;
//   - PlanSectionsAction: scoping rides the ready item (facts_scoped /
//     assigned_fact_ids / assigned_writer_block), nil facts leave the item
//     byte-identical to today's.
package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// ── extractSectionEntries ───────────────────────────────────────────────────

func TestExtractSectionEntries_StringsWithoutFactsStayUnscoped(t *testing.T) {
	entries := extractSectionEntries(map[string]interface{}{
		"sections": []interface{}{"hero", "features"},
	})
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	for i, e := range entries {
		if e.Facts != nil {
			t.Errorf("entry %d: expected nil Facts (unscoped), got %v", i, e.Facts)
		}
	}
}

func TestExtractSectionEntries_SectionFactsAlignByRawIndexAcrossSkips(t *testing.T) {
	// The second raw entry is malformed (a number) and is skipped by the name
	// walk. The third entry's facts live at RAW index 2 — if the lookup used
	// the OUTPUT index, "cta" would inherit ["F2"] meant for the skipped slot.
	entries := extractSectionEntries(map[string]interface{}{
		"sections": []interface{}{"hero", 42, "cta"},
		"section_facts": []interface{}{
			[]interface{}{"F1"},
			[]interface{}{"F2"},
			[]interface{}{"F3"},
		},
	})
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if len(entries[0].Facts) != 1 || entries[0].Facts[0] != "F1" {
		t.Errorf("hero: expected [F1], got %v", entries[0].Facts)
	}
	if len(entries[1].Facts) != 1 || entries[1].Facts[0] != "F3" {
		t.Errorf("cta: expected [F3] (raw-index lookup), got %v", entries[1].Facts)
	}
}

func TestExtractSectionEntries_NullVsEmptyAreDistinct(t *testing.T) {
	entries := extractSectionEntries(map[string]interface{}{
		"sections": []interface{}{"hero", "features"},
		"section_facts": []interface{}{
			nil,             // unscoped
			[]interface{}{}, // deliberately factless
		},
	})
	if entries[0].Facts != nil {
		t.Errorf("hero: null assignment must stay nil (unscoped), got %v", entries[0].Facts)
	}
	if entries[1].Facts == nil || len(entries[1].Facts) != 0 {
		t.Errorf("features: [] assignment must be non-nil empty (factless), got %#v", entries[1].Facts)
	}
}

func TestExtractSectionEntries_ObjectFormCarriesItsOwnFacts(t *testing.T) {
	entries := extractSectionEntries(map[string]interface{}{
		"sections": []interface{}{
			map[string]interface{}{"name": "features", "facts": []interface{}{"F2", "F3"}},
			map[string]interface{}{"component": "cta"},
		},
	})
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "features" || len(entries[0].Facts) != 2 {
		t.Errorf("expected features with 2 facts, got %+v", entries[0])
	}
	if entries[1].Name != "cta" || entries[1].Facts != nil {
		t.Errorf("expected cta unscoped via component key, got %+v", entries[1])
	}
}

// ── ValidateSitePlanAction normalise pass ───────────────────────────────────

// validateParams builds the minimal ActionParams ValidateSitePlanAction needs.
// DB is nil, which skips the chrome-strip and component-resolution blocks —
// the normalise pass under test runs unconditionally after them.
func validateParams(plan map[string]interface{}) ActionParams {
	return ActionParams{
		Context:       context.Background(),
		Logger:        zap.NewNop(),
		CollectedData: map[string]interface{}{"llm_plan": plan},
		StepConfig:    models.Step{Config: map[string]interface{}{"plan_field": "llm_plan"}},
		ExecutionContext: &orchtypes.ExecutionContext{
			OrchestrationID: "44444444-4444-4444-4444-444444444444",
			StepName:        "validate_plan",
		},
	}
}

func TestValidateSitePlan_NormalisesObjectSectionsToStringsPlusFacts(t *testing.T) {
	plan := map[string]interface{}{
		"pages": []interface{}{
			map[string]interface{}{
				"name": "index", "page_type": "index",
				"sections": []interface{}{
					map[string]interface{}{"name": "hero", "facts": []interface{}{"F1"}},
					"testimonials",
					map[string]interface{}{"name": "cta", "facts": []interface{}{}},
				},
			},
		},
	}
	out, err := ValidateSitePlanAction(context.Background(), validateParams(plan))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pages := out.(map[string]interface{})["pages"].([]interface{})
	pm := pages[0].(map[string]interface{})

	sections, ok := pm["sections"].([]interface{})
	if !ok || len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %#v", pm["sections"])
	}
	for i, s := range sections {
		if _, isString := s.(string); !isString {
			t.Errorf("section %d: expected a plain string after normalise, got %T — pages.sections readers expect strings", i, s)
		}
	}
	facts, ok := pm["section_facts"].([]interface{})
	if !ok || len(facts) != 3 {
		t.Fatalf("expected aligned section_facts of length 3, got %#v", pm["section_facts"])
	}
	if f0, ok := facts[0].([]interface{}); !ok || len(f0) != 1 || f0[0] != "F1" {
		t.Errorf("hero facts: expected [F1], got %#v", facts[0])
	}
	if facts[1] != nil {
		t.Errorf("testimonials (plain string) must map to nil facts, got %#v", facts[1])
	}
	if f2, ok := facts[2].([]interface{}); !ok || len(f2) != 0 {
		t.Errorf("cta facts: expected explicit empty array preserved, got %#v", facts[2])
	}
}

func TestValidateSitePlan_AllStringSectionsGainNoSectionFactsKey(t *testing.T) {
	plan := map[string]interface{}{
		"pages": []interface{}{
			map[string]interface{}{
				"name": "index", "page_type": "index",
				"sections": []interface{}{"hero", "features"},
			},
		},
	}
	out, err := ValidateSitePlanAction(context.Background(), validateParams(plan))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pm := out.(map[string]interface{})["pages"].([]interface{})[0].(map[string]interface{})
	if _, present := pm["section_facts"]; present {
		t.Errorf("an all-string plan must be left byte-identical — section_facts key found: %#v", pm["section_facts"])
	}
}

// ── composeScopedWriterBlock ────────────────────────────────────────────────

func factScopingEvidenceBase() map[string]interface{} {
	return map[string]interface{}{
		"facts": []interface{}{
			map[string]interface{}{
				"id": "F1", "claim": "Runs 12 live sites",
				"value": float64(12), "writer_line": "We run {value} live production sites.",
			},
			map[string]interface{}{
				"id": "F2", "claim": "Corrects errors when told",
				"writer_line": "We correct published errors when they are pointed out.",
			},
		},
	}
}

func TestComposeScopedWriterBlock_FiltersToAssignedFactsOnly(t *testing.T) {
	block := composeScopedWriterBlock(factScopingEvidenceBase(), []string{"F1"}, zap.NewNop(), "hero")
	if !strings.Contains(block, "We run 12 live production sites.") {
		t.Errorf("expected the assigned fact's writer_line with the value substituted, got: %q", block)
	}
	if strings.Contains(block, "correct published errors") {
		t.Errorf("unassigned fact leaked into the scoped block: %q", block)
	}
}

func TestComposeScopedWriterBlock_UnknownIDIsInertAndEmptyAssignmentComposesEmpty(t *testing.T) {
	if block := composeScopedWriterBlock(factScopingEvidenceBase(), []string{"F9-not-real"}, zap.NewNop(), "hero"); block != "" {
		t.Errorf("an assignment of only unknown IDs must compose an empty block, got: %q", block)
	}
	if block := composeScopedWriterBlock(factScopingEvidenceBase(), []string{}, zap.NewNop(), "cta"); block != "" {
		t.Errorf("an explicit empty assignment must compose an empty block, got: %q", block)
	}
	if block := composeScopedWriterBlock(nil, []string{"F1"}, zap.NewNop(), "hero"); block != "" {
		t.Errorf("no evidence base must compose an empty block, got: %q", block)
	}
}

// ── PlanSectionsAction: scoping rides the ready item ────────────────────────

func TestPlanSections_FactScopingRidesTheReadyItems(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	heroID, featID, ctaID := uuid.New().String(), uuid.New().String(), uuid.New().String()

	// loadComponentSchemas: all three resolve by name.
	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRow(heroID, "hero", "hero", "<section>h</section>", "section").
			AddRow(featID, "features", "features", "", nil, nil, nil, "<section>f</section>", planSectionsSchema, "template", nil, "section").
			AddRow(ctaID, "cta", "cta", "", nil, nil, nil, "<section>c</section>", planSectionsSchema, "template", nil, "section"))

	// Stored-identity read (Path 0): no rows — name resolution decides.
	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "index").
		WillReturnRows(slotRows())

	// resolver.ensureSpecs: the evidence base is in the spec cache — scoping
	// must not issue its own query for it.
	mock.ExpectQuery("FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"aspect", "data"}).
			AddRow("evidence_base", `{"facts":[
				{"id":"F1","claim":"Runs 12 live sites","value":12,"writer_line":"We run {value} live production sites."},
				{"id":"F2","claim":"Corrects errors","writer_line":"We correct published errors when they are pointed out."}]}`))
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"section_name", "summary"}))

	// Ready sections ⇒ persistSectionSkips merge.
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	params := planParams(db, siteID, "index", []string{"hero", "features", "cta"})
	params.CollectedData["input_data"].(map[string]interface{})["section_facts"] = []interface{}{
		[]interface{}{"F1"}, // hero: scoped to F1
		nil,                 // features: unscoped
		[]interface{}{},     // cta: deliberately factless
	}
	params.StepConfig.Config["section_facts"] = "input_data.section_facts"

	out, err := PlanSectionsAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items := readyItems(t, out)
	if len(items) != 3 {
		t.Fatalf("expected 3 ready items, got %d (%+v)", len(items), out)
	}
	byName := map[string]sectionPlanItem{}
	for _, it := range items {
		byName[it.Name] = it
	}

	hero := byName["hero"]
	if !hero.FactsScoped || len(hero.AssignedFactIDs) != 1 || hero.AssignedFactIDs[0] != "F1" {
		t.Errorf("hero: expected facts_scoped with [F1], got scoped=%v ids=%v", hero.FactsScoped, hero.AssignedFactIDs)
	}
	if !strings.Contains(hero.AssignedWriterBlock, "We run 12 live production sites.") {
		t.Errorf("hero: assigned block missing the fact (value substituted at compose time), got: %q", hero.AssignedWriterBlock)
	}
	if strings.Contains(hero.AssignedWriterBlock, "correct published errors") {
		t.Errorf("hero: unassigned fact leaked into the scoped block: %q", hero.AssignedWriterBlock)
	}

	features := byName["features"]
	if features.FactsScoped || features.AssignedFactIDs != nil || features.AssignedWriterBlock != "" {
		t.Errorf("features: a nil assignment must leave the item unscoped and byte-identical to today's, got %+v", features)
	}

	cta := byName["cta"]
	if !cta.FactsScoped {
		t.Errorf("cta: an explicit [] assignment must mark the section scoped (factless), got %+v", cta)
	}
	if cta.AssignedWriterBlock != "" {
		t.Errorf("cta: factless section must carry an empty block, got: %q", cta.AssignedWriterBlock)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_EmptyCompositionDegradesToUnscopedWithDurableRecord pins the
// council-flagged case (corr 902a8563, bug_historian): a NON-EMPTY assignment
// whose every ID is unknown composes an empty block. That is a composition
// anomaly, not a deliberately factless section — if it were marked scoped, the
// writer would render the "state no facts here" branch and a broken assignment
// would read exactly like a deliberate one. It must instead degrade to
// UNSCOPED (today's site-wide block) and leave a durable agent_error_log row,
// not just pod output.
func TestPlanSections_EmptyCompositionDegradesToUnscopedWithDurableRecord(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	heroID := uuid.New().String()

	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRow(heroID, "hero", "hero", "<section>h</section>", "section"))
	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "index").
		WillReturnRows(slotRows())
	mock.ExpectQuery("FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"aspect", "data"}).
			AddRow("evidence_base", `{"facts":[{"id":"F1","claim":"Runs 12 live sites","value":12,"writer_line":"We run {value} live production sites."}]}`))
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"section_name", "summary"}))

	// The durable record is the assertion: severity and error_code pinned.
	mock.ExpectExec("INSERT INTO agent_error_log").
		// Canonical 13-column bind list (RFC_012 option B): action is bind 9,
		// error_code bind 11, severity bind 12 — all three still pinned by value.
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"plan_sections", sqlmock.AnyArg(), "FACT_SCOPING_EMPTY_COMPOSITION", "warning", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	params := planParams(db, siteID, "index", []string{"hero"})
	params.CollectedData["input_data"].(map[string]interface{})["section_facts"] = []interface{}{
		[]interface{}{"F9-not-real", "F10-also-not-real"},
	}
	params.StepConfig.Config["section_facts"] = "input_data.section_facts"

	out, err := PlanSectionsAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items := readyItems(t, out)
	if len(items) != 1 {
		t.Fatalf("expected 1 ready item, got %d", len(items))
	}
	hero := items[0]
	if hero.FactsScoped {
		t.Errorf("an all-unknown assignment must degrade to UNSCOPED, not render the factless branch: %+v", hero)
	}
	if hero.AssignedWriterBlock != "" || hero.AssignedFactIDs != nil {
		t.Errorf("degraded item must carry no scoping fields, got %+v", hero)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations (the agent_error_log INSERT is the durable-record assertion): %v", err)
	}
}
