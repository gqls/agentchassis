// FILE: platform/orchestration/actions/plan_section_subjects_test.go
//
// Per-section subjects (apis_uk lane, 2026-08-26) — the SECOND structured
// per-section field, extending RFC_016's ratified contract (§5.1): object-form
// entries live only between the planner LLM and validate_plan's normalise
// pass; downstream, every carrier keeps its historical shape plus a named,
// aligned, per-page sibling key (section_subjects, beside section_facts).
//
// The defect this exists for: pages.sections is []string, so N same-named
// slots get one identical brief and the writer produces N variations on one
// subject (apis.uk: one content_rewrite rewrote all six sections about the
// waggle dance). Each seam's silent failure is pinned the same way
// fact_scoping_151_test.go pins the facts field:
//
//   - extractSectionEntries: subject read from the entry's own "subject" key,
//     else the page-level section_subjects sibling at RAW index;
//   - ValidateSitePlanAction's normalise pass: subjects leave as an aligned
//     section_subjects array; an all-string plan gains no new key;
//   - carrySectionFactsOntoRealised: a subject rides the carry onto a
//     realised composition, and a subject-only entry never fabricates a
//     "facts" key (the NULL/[] distinction on facts is load-bearing);
//   - PlanSectionsAction: subject rides the ready item, aligned through the
//     site-level filter; "" leaves the item byte-identical to today's.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// ── extractSectionEntries ───────────────────────────────────────────────────

func TestExtractSectionEntries_SubjectFromObjectKeyWinsOverSibling(t *testing.T) {
	entries := extractSectionEntries(map[string]interface{}{
		"sections": []interface{}{
			map[string]interface{}{"name": "features", "facts": []interface{}{}, "subject": "  How honey is graded  "},
			"testimonials",
		},
		"section_subjects": []interface{}{"SIBLING-IGNORED-WHEN-OBJECT-HAS-ONE", "What beekeepers say"},
	})
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Subject != "How honey is graded" {
		t.Errorf("features: object subject (trimmed) must win, got %q", entries[0].Subject)
	}
	if entries[1].Subject != "What beekeepers say" {
		t.Errorf("testimonials: sibling subject must apply to a plain-string entry, got %q", entries[1].Subject)
	}
}

func TestExtractSectionEntries_SectionSubjectsAlignByRawIndexAcrossSkips(t *testing.T) {
	// Mirrors the facts test: the malformed second raw entry is skipped; an
	// OUTPUT-index lookup would hand "cta" the skipped slot's subject.
	entries := extractSectionEntries(map[string]interface{}{
		"sections":         []interface{}{"hero", 42, "cta"},
		"section_subjects": []interface{}{"S-hero", "S-skipped", "S-cta"},
	})
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Subject != "S-hero" || entries[1].Subject != "S-cta" {
		t.Errorf("raw-index lookup broken: got %q / %q", entries[0].Subject, entries[1].Subject)
	}
}

// ── ValidateSitePlanAction normalise pass ───────────────────────────────────

func TestValidateSitePlan_NormalisesSubjectsToAlignedSibling(t *testing.T) {
	plan := map[string]interface{}{
		"pages": []interface{}{
			map[string]interface{}{
				"name": "index", "page_type": "index",
				"sections": []interface{}{
					map[string]interface{}{"name": "hero", "facts": []interface{}{}, "subject": "A closer look at bees"},
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
	pm := out.(map[string]interface{})["pages"].([]interface{})[0].(map[string]interface{})

	sections, ok := pm["sections"].([]interface{})
	if !ok || len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %#v", pm["sections"])
	}
	for i, s := range sections {
		if _, isString := s.(string); !isString {
			t.Errorf("section %d: expected a plain string after normalise, got %T", i, s)
		}
	}
	subjects, ok := pm["section_subjects"].([]interface{})
	if !ok || len(subjects) != 3 {
		t.Fatalf("expected aligned section_subjects of length 3, got %#v", pm["section_subjects"])
	}
	if subjects[0] != "A closer look at bees" {
		t.Errorf("hero subject: got %#v", subjects[0])
	}
	if subjects[1] != nil || subjects[2] != nil {
		t.Errorf("no-subject entries must map to nil, got %#v / %#v", subjects[1], subjects[2])
	}
}

func TestValidateSitePlan_AllStringSectionsGainNoSectionSubjectsKey(t *testing.T) {
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
	if _, present := pm["section_subjects"]; present {
		t.Errorf("an all-string plan must be left byte-identical — section_subjects key found: %#v", pm["section_subjects"])
	}
}

// ── carrySectionFactsOntoRealised ───────────────────────────────────────────

func TestCarrySectionFacts_SubjectRidesTheCarry(t *testing.T) {
	realised := []interface{}{"features", "cta"}
	proposed := []interface{}{
		map[string]interface{}{"name": "features", "facts": []interface{}{"F1"}, "subject": "How honey is graded"},
		map[string]interface{}{"name": "cta", "subject": "Visit a local apiary"}, // facts key ABSENT (333 disobedience) but subject present
	}
	merged, carried, unmatched, absent := carrySectionFactsOntoRealised(realised, proposed)
	if carried != 2 {
		t.Fatalf("expected 2 carried, got %d (unmatched=%v)", carried, unmatched)
	}
	m0 := merged[0].(map[string]interface{})
	if m0["subject"] != "How honey is graded" {
		t.Errorf("features: subject must ride the carry, got %#v", m0["subject"])
	}
	if f, ok := m0["facts"].([]interface{}); !ok || len(f) != 1 || f[0] != "F1" {
		t.Errorf("features: facts must still carry beside the subject, got %#v", m0["facts"])
	}
	m1 := merged[1].(map[string]interface{})
	if m1["subject"] != "Visit a local apiary" {
		t.Errorf("cta: subject-only entry must still carry its subject, got %#v", m1["subject"])
	}
	if _, hasFacts := m1["facts"]; hasFacts {
		t.Errorf("cta: a facts key must NEVER be fabricated for a subject-only entry — the NULL/[] distinction is load-bearing, got %#v", m1["facts"])
	}
	// The 333 record is about facts and must still name the entry without one.
	if len(absent) != 1 || absent[0] != "cta" {
		t.Errorf("absent must still record the facts-less entry, got %v", absent)
	}
}

func TestCarrySectionFacts_SubjectOnlyEntryWithEmptySubjectStillSkipped(t *testing.T) {
	realised := []interface{}{"cta"}
	proposed := []interface{}{
		map[string]interface{}{"name": "cta", "subject": "   "}, // no facts, blank subject
	}
	merged, carried, _, absent := carrySectionFactsOntoRealised(realised, proposed)
	if carried != 0 {
		t.Fatalf("a blank-subject facts-less entry must carry nothing, got carried=%d", carried)
	}
	if _, isString := merged[0].(string); !isString {
		t.Errorf("realised entry must be returned unchanged, got %#v", merged[0])
	}
	if len(absent) != 1 {
		t.Errorf("the facts-less entry must still be recorded absent, got %v", absent)
	}
}

// ── PlanSectionsAction: subject rides the ready item, aligned through filter ─

func TestPlanSections_SubjectRidesTheReadyItemsThroughTheSiteLevelFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	heroID, featID, ctaID := uuid.New().String(), uuid.New().String(), uuid.New().String()

	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRow(heroID, "hero", "hero", "<section>h</section>", "section").
			AddRow(featID, "features", "features", "", nil, nil, nil, "<section>f</section>", planSectionsSchema, "template", nil, "section").
			AddRow(ctaID, "cta", "cta", "", nil, nil, nil, "<section>c</section>", planSectionsSchema, "template", nil, "section"))
	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "index").
		WillReturnRows(slotRows())
	mock.ExpectQuery("FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"aspect", "data"}))
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"section_name", "summary"}))
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	// "site-footer" is dropped by the site-level filter; its subject must be
	// dropped WITH it, or every later subject lands one slot early.
	params := planParams(db, siteID, "index", []string{"hero", "site-footer", "features", "cta"})
	params.CollectedData["input_data"].(map[string]interface{})["section_subjects"] = []interface{}{
		"The colony as one organism",
		"DROPPED-WITH-THE-CHROME-ENTRY",
		nil,
		"  Meet the solitary bees  ",
	}
	params.StepConfig.Config["section_subjects"] = "input_data.section_subjects"

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
	if got := byName["hero"].Subject; got != "The colony as one organism" {
		t.Errorf("hero subject: got %q", got)
	}
	if got := byName["features"].Subject; got != "" {
		t.Errorf("features: a nil subject must leave the item byte-identical (omitempty), got %q", got)
	}
	if got := byName["cta"].Subject; got != "Meet the solitary bees" {
		t.Errorf("cta subject (trimmed, aligned across the filtered entry): got %q", got)
	}
	if scoped := byName["hero"].FactsScoped; scoped {
		t.Errorf("no section_facts were wired — subject wiring must not mark items facts-scoped")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// ── round-2 additions (council 4bd35ed8, REVISE round 1) ────────────────────
//
// bug_historian (MEDIUM): three parallel index-aligned lists are the platform's
// most repeated silent-content-loss shape (bugs_closed/041, 095, 039) — an
// unequal-length section_subjects must degrade to UNASSIGNED, never shift a
// subject onto the wrong sibling. editquality/bug_historian (LOW): the carry's
// object-realised arm shipped untested; realised lists are plain strings today,
// but the arm exists and its clone must preserve foreign keys.

func TestPlanSections_ShortOrIllTypedSubjectsDegradeToUnassignedNeverShift(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	heroID, featID, ctaID := uuid.New().String(), uuid.New().String(), uuid.New().String()

	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRow(heroID, "hero", "hero", "<section>h</section>", "section").
			AddRow(featID, "features", "features", "", nil, nil, nil, "<section>f</section>", planSectionsSchema, "template", nil, "section").
			AddRow(ctaID, "cta", "cta", "", nil, nil, nil, "<section>c</section>", planSectionsSchema, "template", nil, "section"))
	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "index").
		WillReturnRows(slotRows())
	mock.ExpectQuery("FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"aspect", "data"}))
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"section_name", "summary"}))
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	// Three sections; the subjects array is SHORT (1 entry) and its would-be
	// second entry, if the walk ran off the end wrongly, is a non-string.
	// Correct behaviour: index 0 gets its subject, everything past the end or
	// ill-typed is "" — no shift, no error, no misassignment.
	params := planParams(db, siteID, "index", []string{"hero", "features", "cta"})
	params.CollectedData["input_data"].(map[string]interface{})["section_subjects"] = []interface{}{
		"Only the first slot has one",
	}
	params.StepConfig.Config["section_subjects"] = "input_data.section_subjects"

	out, err := PlanSectionsAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items := readyItems(t, out)
	if len(items) != 3 {
		t.Fatalf("expected 3 ready items, got %d", len(items))
	}
	byName := map[string]sectionPlanItem{}
	for _, it := range items {
		byName[it.Name] = it
	}
	if got := byName["hero"].Subject; got != "Only the first slot has one" {
		t.Errorf("hero subject: got %q", got)
	}
	if byName["features"].Subject != "" || byName["cta"].Subject != "" {
		t.Errorf("beyond-length entries must be UNASSIGNED, never shifted: features=%q cta=%q",
			byName["features"].Subject, byName["cta"].Subject)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestPlanSections_LongAndIllTypedSubjectsAreIgnoredBeyondSections(t *testing.T) {
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
		WillReturnRows(sqlmock.NewRows([]string{"aspect", "data"}))
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"section_name", "summary"}))
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	// One section; the subjects array is LONGER and holds a non-string at the
	// aligned index. Non-string => "" (unassigned); surplus entries ignored.
	params := planParams(db, siteID, "index", []string{"hero"})
	params.CollectedData["input_data"].(map[string]interface{})["section_subjects"] = []interface{}{
		42, "surplus-1", "surplus-2",
	}
	params.StepConfig.Config["section_subjects"] = "input_data.section_subjects"

	out, err := PlanSectionsAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items := readyItems(t, out)
	if len(items) != 1 {
		t.Fatalf("expected 1 ready item, got %d", len(items))
	}
	if items[0].Subject != "" {
		t.Errorf("a non-string aligned entry must read as unassigned, got %q", items[0].Subject)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestCarrySectionFacts_ObjectRealisedClonePreservesForeignKeys(t *testing.T) {
	realised := []interface{}{
		map[string]interface{}{"name": "features", "kept_by_someone_else": "survives"},
	}
	proposed := []interface{}{
		map[string]interface{}{"name": "features", "facts": []interface{}{"F1"}, "subject": "How honey is graded"},
	}
	merged, carried, unmatched, absent := carrySectionFactsOntoRealised(realised, proposed)
	if carried != 1 || len(unmatched) != 0 || len(absent) != 0 {
		t.Fatalf("carried=%d unmatched=%v absent=%v", carried, unmatched, absent)
	}
	m := merged[0].(map[string]interface{})
	if m["kept_by_someone_else"] != "survives" {
		t.Errorf("a realised object entry's foreign keys are not this function's to discard, got %#v", m)
	}
	if m["subject"] != "How honey is graded" {
		t.Errorf("object-arm subject set missing, got %#v", m["subject"])
	}
	if f, ok := m["facts"].([]interface{}); !ok || len(f) != 1 || f[0] != "F1" {
		t.Errorf("object-arm facts set missing, got %#v", m["facts"])
	}
}

func TestCarrySectionFacts_ObjectRealisedSubjectOnlyFabricatesNoFactsKey(t *testing.T) {
	realised := []interface{}{
		map[string]interface{}{"name": "cta", "kept": "v"},
	}
	proposed := []interface{}{
		map[string]interface{}{"name": "cta", "subject": "Visit a local apiary"},
	}
	merged, carried, _, absent := carrySectionFactsOntoRealised(realised, proposed)
	if carried != 1 {
		t.Fatalf("carried=%d", carried)
	}
	m := merged[0].(map[string]interface{})
	if m["subject"] != "Visit a local apiary" || m["kept"] != "v" {
		t.Errorf("subject or foreign key lost on the object arm: %#v", m)
	}
	if _, hasFacts := m["facts"]; hasFacts {
		t.Errorf("a facts key must never be fabricated on the object arm either — NULL/[] is load-bearing: %#v", m["facts"])
	}
	if len(absent) != 1 || absent[0] != "cta" {
		t.Errorf("seed-333 absence record must be unchanged, got %v", absent)
	}
}

// ── post-approval advisory (council 4bd35ed8 r2, bug_historian MEDIUM) ──────
// The structural half of rule 17: repeated components without subjects leave a
// durable record — but ONLY on plans that carry subjects at all, so every
// pre-seed-640 plan (bare strings) stays silent. The helper is pure; the
// recording wire reuses the PLAN_PAGE_MERGE_LOSSY channel one line above it.

func TestSubjectlessRepeatFindings_GateAndDetection(t *testing.T) {
	page := func(name string, entries ...sectionEntry) planPageRow {
		return planPageRow{Name: name, Sections: entries}
	}
	// (a) No subject anywhere: repeats alone must produce NOTHING — this is
	// every plan before seed 640 applies, and firing here would spam the
	// fleet retroactively.
	if f := subjectlessRepeatFindings([]planPageRow{
		page("index", sectionEntry{Name: "text"}, sectionEntry{Name: "text"}, sectionEntry{Name: "hero"}),
	}); len(f) != 0 {
		t.Fatalf("pre-rule plans must be silent, got %d findings", len(f))
	}
	// (b) Subjects present and every repeat has one: nothing.
	if f := subjectlessRepeatFindings([]planPageRow{
		page("index",
			sectionEntry{Name: "text", Subject: "A"},
			sectionEntry{Name: "text", Subject: "B"},
			sectionEntry{Name: "hero"}),
	}); len(f) != 0 {
		t.Fatalf("compliant repeats must be silent, got %d findings", len(f))
	}
	// (c) The planner plays the subject game but misses a repeat: one finding,
	// carrying page + component + counts.
	f := subjectlessRepeatFindings([]planPageRow{
		page("index",
			sectionEntry{Name: "text", Subject: "A"},
			sectionEntry{Name: "text"},
			sectionEntry{Name: "hero"}),
	})
	if len(f) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(f))
	}
	if f[0].ErrorCode != "SUBJECT_MISSING_ON_REPEATED_COMPONENT" || f[0].Severity != "warning" {
		t.Errorf("finding shape: %+v", f[0])
	}
	if f[0].Context["component"] != "text" || f[0].Context["repeats"] != 2 || f[0].Context["without_subject"] != 1 {
		t.Errorf("finding context: %#v", f[0].Context)
	}
	// (d) A subject on ANOTHER page arms the gate for the whole plan: page two
	// has bare repeats and must be flagged — post-640 the planner owes every
	// page object form.
	f = subjectlessRepeatFindings([]planPageRow{
		page("about", sectionEntry{Name: "hero", Subject: "Who we are"}),
		page("index", sectionEntry{Name: "text"}, sectionEntry{Name: "text"}),
	})
	if len(f) != 1 || f[0].Context["page"] != "index" {
		t.Fatalf("plan-level gate must cover sibling pages, got %#v", f)
	}
	// (e) A singleton without a subject is fine — only REPEATS owe one.
	if f := subjectlessRepeatFindings([]planPageRow{
		page("index", sectionEntry{Name: "text", Subject: "A"}, sectionEntry{Name: "cta"}),
	}); len(f) != 0 {
		t.Fatalf("singletons owe no subject, got %d findings", len(f))
	}
}
