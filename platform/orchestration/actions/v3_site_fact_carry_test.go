// FILE: platform/orchestration/actions/v3_site_fact_carry_test.go
//
// bugs_open/151 candidate 1b (RFC_016 §3b) — plan-time fact assignments must
// SURVIVE the re-plan preservation guard.
//
// The defect these pin: an assignment travels INSIDE its section entry, and
// Pass B/B2 restore a built page's realised composition by replacing the LLM's
// entries wholesale. Measured on fundamentallyai 2026-08-08 (corr 1cb17b11):
// the raw plan carried 6 object entries with facts on the index page, the
// validated plan carried the 6 realised STRING sections and no section_facts at
// all. Candidate 1 could therefore only ever reach pages built AFTER their plan
// first carried assignments — never the already-built pages that motivated it.
//
// Two failure modes are pinned deliberately, because they produce identical
// plans and only the second is obvious:
//
//   - the carry does nothing (facts silently absent);
//   - the carry matches NOTHING and reports success (a name-match against a
//     composition that does not contain those names). A miss must be counted
//     and recorded, never inferred from the absence of a hit.
package actions

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// ── fixtures ────────────────────────────────────────────────────────────────

// objSection builds the planner's object-form entry: the shape seed 333's
// rule 17 makes mandatory per page.
func objSection(name string, facts ...string) map[string]interface{} {
	f := make([]interface{}, len(facts))
	for i, v := range facts {
		f[i] = v
	}
	return map[string]interface{}{"name": name, "facts": f}
}

// llmPageEntries is llmPage for a page whose entries are not all bare strings.
func llmPageEntries(name, url string, entries ...interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name": name, "url": url, "page_type": "content",
		"title": name, "sections": entries,
	}
}

// planSectionsOf returns a page's raw section entries — unlike sectionsOf it
// tolerates object form, which is the whole point here.
func planSectionsOf(t *testing.T, pages []interface{}, page string) []interface{} {
	t.Helper()
	for _, p := range pages {
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		if n, _ := pm["name"].(string); n != page {
			continue
		}
		raw, _ := pm["sections"].([]interface{})
		return raw
	}
	t.Fatalf("page %q not present in plan", page)
	return nil
}

// factsOn returns the facts assigned to the section at index i, and whether the
// entry carries an assignment at all (absent != empty — the tri-state).
func factsOn(t *testing.T, entries []interface{}, i int) ([]string, bool) {
	t.Helper()
	if i >= len(entries) {
		t.Fatalf("section index %d out of range (%d entries)", i, len(entries))
	}
	obj, ok := entries[i].(map[string]interface{})
	if !ok {
		return nil, false
	}
	raw, ok := obj["facts"].([]interface{})
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(raw))
	for _, f := range raw {
		s, _ := f.(string)
		out = append(out, s)
	}
	return out, true
}

func sectionNamesOf(t *testing.T, entries []interface{}) []string {
	t.Helper()
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		n, ok := sectionEntryName(e)
		if !ok {
			t.Fatalf("entry %#v has no resolvable section name", e)
		}
		out = append(out, n)
	}
	return out
}

// ── Pass B2: the motivating case ────────────────────────────────────────────

// The built page keeps its realised composition AND the planner's assignments
// reach it. Before candidate 1b the assignments went with the discarded entries.
func TestPassB2_CarriesFactAssignmentsOntoRestoredSections(t *testing.T) {
	existing := []interface{}{
		realised("about", "/about.html", "deployed",
			`["hero-about","info-card-grid","call-to-action"]`, false),
	}
	// The planner re-composes the page (a real composition change, so Pass B2
	// restores) and scopes facts to two sections that DO exist on the built page.
	llm := []interface{}{
		llmPageEntries("about", "/about.html",
			objSection("hero-about", "F1-live-sites"),
			objSection("generic-text-block", "F9-not-on-this-page"),
			objSection("call-to-action"),
		),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	entries := planSectionsOf(t, got, "about")
	if names := sectionNamesOf(t, entries); !equalStrings(names,
		[]string{"hero-about", "info-card-grid", "call-to-action"}) {
		t.Fatalf("realised composition was not preserved: %v", names)
	}
	if counts.SnappedSections != 1 {
		t.Errorf("snapped_sections = %d, want 1", counts.SnappedSections)
	}

	if f, ok := factsOn(t, entries, 0); !ok || !equalStrings(f, []string{"F1-live-sites"}) {
		t.Errorf("hero-about: facts = %v (present=%v), want [F1-live-sites] carried onto the restored entry", f, ok)
	}
	if _, ok := factsOn(t, entries, 1); ok {
		t.Errorf("info-card-grid was assigned nothing and must stay UNSCOPED, not empty-scoped — null and [] mean different things to the writer")
	}
	if f, ok := factsOn(t, entries, 2); !ok || len(f) != 0 {
		t.Errorf("call-to-action: want a deliberately factless assignment ([] preserved), got %v (present=%v)", f, ok)
	}
	if counts.SectionFactsCarried != 2 {
		t.Errorf("section_facts_carried = %d, want 2", counts.SectionFactsCarried)
	}

	// The half that is invisible in the plan: the assignment that matched nothing.
	if len(counts.FactCarryMisses) != 1 {
		t.Fatalf("expected 1 page with unmatched assignments, got %#v", counts.FactCarryMisses)
	}
	miss := counts.FactCarryMisses[0]
	if miss.Page != "about" || !equalStrings(miss.Sections, []string{"generic-text-block"}) {
		t.Errorf("miss = %+v, want about/[generic-text-block] — an assignment that matched nothing must be NAMED, not merely absent", miss)
	}
}

// A plan with no assignments at all — every plan before seed 333, and every
// page the planner emits as bare strings — must come out exactly as it did
// before candidate 1b: plain strings, no object entries, nothing counted.
func TestPassB2_NoAssignments_RestoreStaysPlainStrings(t *testing.T) {
	existing := []interface{}{
		realised("about", "/about.html", "deployed",
			`["hero-about","info-card-grid"]`, false),
	}
	llm := []interface{}{llmPage("about", "/about.html", "hero", "cta")}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	for i, e := range planSectionsOf(t, got, "about") {
		if _, isString := e.(string); !isString {
			t.Errorf("section %d: %T — a plan carrying no assignments must be shaped exactly as before", i, e)
		}
	}
	if counts.SectionFactsCarried != 0 || len(counts.FactCarryMisses) != 0 {
		t.Errorf("carried=%d misses=%#v, want a silent no-op",
			counts.SectionFactsCarried, counts.FactCarryMisses)
	}
}

// The cheapest carry is no carry: when the planner re-emits the realised names
// (candidate 1b (i)'s whole purpose) the composition is UNCHANGED, so Pass B2
// must leave the entries — and their assignments — alone.
//
// This also pins the counter correction. Comparing whole entries with %v, as
// this did until 1b, read object-vs-string as "changed" for every composed page
// the moment seed 333 shipped: snapped_sections silently became a count of
// shape differences rather than composition changes.
func TestPassB2_SameNamesInObjectFormIsNotACompositionChange(t *testing.T) {
	existing := []interface{}{
		realised("index", "/index.html", "deployed", `["hero","stat-band"]`, false),
	}
	llm := []interface{}{
		llmPageEntries("index", "/index.html",
			objSection("hero"),
			objSection("stat-band", "F1-live-sites", "F2-council-seats"),
		),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	if counts.SnappedSections != 0 {
		t.Errorf("snapped_sections = %d, want 0 — the same components in the same order is not a composition change", counts.SnappedSections)
	}
	entries := planSectionsOf(t, got, "index")
	if f, ok := factsOn(t, entries, 1); !ok ||
		!equalStrings(f, []string{"F1-live-sites", "F2-council-seats"}) {
		t.Errorf("stat-band: facts = %v (present=%v), want both assignments untouched", f, ok)
	}
}

// A page may hold two entries of the same component. Assignments queue by name
// in emission order rather than all landing on the first slot.
func TestPassB2_RepeatedComponentTakesAssignmentsInOrder(t *testing.T) {
	existing := []interface{}{
		realised("services", "/services.html", "deployed",
			`["hero","generic-text-block","generic-text-block"]`, false),
	}
	llm := []interface{}{
		llmPageEntries("services", "/services.html",
			objSection("generic-text-block", "F3-first"),
			objSection("generic-text-block", "F4-second"),
		),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	entries := planSectionsOf(t, got, "services")
	if f, _ := factsOn(t, entries, 1); !equalStrings(f, []string{"F3-first"}) {
		t.Errorf("first text block: %v, want [F3-first]", f)
	}
	if f, _ := factsOn(t, entries, 2); !equalStrings(f, []string{"F4-second"}) {
		t.Errorf("second text block: %v, want [F4-second] — assignments must queue, not collapse onto slot one", f)
	}
	if counts.SectionFactsCarried != 2 {
		t.Errorf("section_facts_carried = %d, want 2", counts.SectionFactsCarried)
	}
}

// More assignments than slots: the surplus is a miss, and must be reported once
// per dropped assignment rather than once per name.
func TestPassB2_SurplusAssignmentForARepeatedComponentIsReported(t *testing.T) {
	existing := []interface{}{
		realised("services", "/services.html", "deployed",
			`["hero","generic-text-block"]`, false),
	}
	llm := []interface{}{
		llmPageEntries("services", "/services.html",
			objSection("generic-text-block", "F3-first"),
			objSection("generic-text-block", "F4-second"),
		),
	}

	_, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	if len(counts.FactCarryMisses) != 1 ||
		!equalStrings(counts.FactCarryMisses[0].Sections, []string{"generic-text-block"}) {
		t.Fatalf("misses = %#v, want the one surplus assignment named", counts.FactCarryMisses)
	}
	if counts.SectionFactsCarried != 1 {
		t.Errorf("section_facts_carried = %d, want 1", counts.SectionFactsCarried)
	}
}

// ── Pass B: the same loss one pass earlier ──────────────────────────────────

// A renamed page snaps back to the realised identity and its realised
// composition. The planner's assignments must ride across the rename.
func TestPassB_RenameSnapBackCarriesFactAssignments(t *testing.T) {
	existing := []interface{}{
		realised("guide-pricing", "/guide/pricing.html", "deployed",
			`["hero-guide","pricing-table"]`, false),
	}
	llm := []interface{}{
		llmPageEntries("pricing-guide", "/guide/pricing.html",
			objSection("pricing-table", "F5-unit-cost"),
			objSection("hero-guide"),
		),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	if counts.SnappedRename != 1 {
		t.Fatalf("snapped_rename = %d, want 1", counts.SnappedRename)
	}
	entries := planSectionsOf(t, got, "guide-pricing")
	if names := sectionNamesOf(t, entries); !equalStrings(names, []string{"hero-guide", "pricing-table"}) {
		t.Fatalf("realised composition not preserved across the rename: %v", names)
	}
	if f, ok := factsOn(t, entries, 1); !ok || !equalStrings(f, []string{"F5-unit-cost"}) {
		t.Errorf("pricing-table: facts = %v (present=%v) — assignments must survive the rename snap-back too", f, ok)
	}
	if counts.SectionFactsCarried != 2 {
		t.Errorf("section_facts_carried = %d, want 2", counts.SectionFactsCarried)
	}
}

// A NOT-deployed realised page with no sections keeps the LLM's own entries, so
// its assignments ride along whole — no carry involved, and nothing counted.
func TestPassB_CataloguedPageKeepsProposedEntriesWithTheirFacts(t *testing.T) {
	// locked == on the site's first plan, so the page is in the preservation set
	// while still being un-shipped and uncomposed (bugs_open/050 / 051).
	existing := []interface{}{
		realised("tool-audience-check", "/tools/audience-check.html", "planned", `[]`, true),
	}
	llm := []interface{}{
		llmPageEntries("audience-check-tool", "/tools/audience-check.html",
			objSection("tool-hero", "F6-zero-fabricated"),
		),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	entries := planSectionsOf(t, got, "tool-audience-check")
	if f, ok := factsOn(t, entries, 0); !ok || !equalStrings(f, []string{"F6-zero-fabricated"}) {
		t.Errorf("tool-hero: facts = %v (present=%v), want the proposal kept whole", f, ok)
	}
	if counts.SectionFactsCarried != 0 || len(counts.FactCarryMisses) != 0 {
		t.Errorf("nothing was restored over, so nothing should be counted: carried=%d misses=%#v",
			counts.SectionFactsCarried, counts.FactCarryMisses)
	}
}

// ── the branch where an assignment cannot survive at all ────────────────────

// A DEPLOYED sectionless page renders through another subsystem (bugs_open/050),
// so its proposed sections are forced back to empty. Any assignment on them is
// genuinely lost — which is exactly why it must be RECORDED rather than left to
// be inferred from an absence.
func TestPassB2_DeployedSectionlessPageRecordsDiscardedAssignments(t *testing.T) {
	existing := []interface{}{
		realised("blog", "/blog.html", "deployed", `[]`, false),
	}
	llm := []interface{}{
		llmPageEntries("blog", "/blog.html",
			objSection("hero", "F7-idea-stripe"),
			"generic-text-block",
		),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	if s := planSectionsOf(t, got, "blog"); len(s) != 0 {
		t.Fatalf("deployed sectionless page must stay empty, got %#v", s)
	}
	if len(counts.FactCarryMisses) != 1 ||
		!equalStrings(counts.FactCarryMisses[0].Sections, []string{"hero"}) {
		t.Errorf("misses = %#v, want the discarded hero assignment named — silent loss here reads exactly like no assignment ever being made", counts.FactCarryMisses)
	}
}

// ── end to end: nothing downstream ever sees object form ────────────────────

// The claim candidate 1b rests on: entries carried in object form are split by
// ValidateSitePlanAction's normalise pass into plain strings plus an aligned
// section_facts array, because that pass runs LATER IN THE SAME FUNCTION than
// the reconcile call. If the order were the other way round this test fails and
// the 15+ readers of pages["sections"] would be in the blast radius.
func TestValidateSitePlan_CarriedFactsLeaveAsStringsPlusSectionFacts(t *testing.T) {
	plan := map[string]interface{}{
		"pages": []interface{}{
			llmPageEntries("about", "/about.html",
				objSection("hero-about", "F1-live-sites"),
				objSection("generic-text-block", "F2-council-seats"),
			),
		},
	}
	params := ActionParams{
		Context: context.Background(),
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"llm_plan": plan,
			"existing_pages": []interface{}{
				realised("about", "/about.html", "deployed",
					`["hero-about","info-card-grid"]`, false),
			},
		},
		StepConfig: models.Step{Config: map[string]interface{}{"plan_field": "llm_plan"}},
		ExecutionContext: &orchtypes.ExecutionContext{
			OrchestrationID: "44444444-4444-4444-4444-444444444444",
			StepName:        "validate_plan",
		},
	}

	out, err := ValidateSitePlanAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pm := out.(map[string]interface{})["pages"].([]interface{})[0].(map[string]interface{})

	sections, ok := pm["sections"].([]interface{})
	if !ok || len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %#v", pm["sections"])
	}
	for i, s := range sections {
		if _, isString := s.(string); !isString {
			t.Errorf("section %d is %T — object form must not escape validate_plan; sync_pages_to_db serialises this array straight into pages.sections", i, s)
		}
	}
	if names := sectionNamesOf(t, sections); !equalStrings(names, []string{"hero-about", "info-card-grid"}) {
		t.Fatalf("realised composition not preserved: %v", names)
	}

	facts, ok := pm["section_facts"].([]interface{})
	if !ok || len(facts) != 2 {
		t.Fatalf("expected an aligned section_facts of length 2, got %#v", pm["section_facts"])
	}
	f0, ok := facts[0].([]interface{})
	if !ok || len(f0) != 1 || f0[0] != "F1-live-sites" {
		t.Errorf("hero-about facts = %#v, want [F1-live-sites] — the carry did not reach the persisted shape", facts[0])
	}
	if facts[1] != nil {
		t.Errorf("info-card-grid was never assigned anything; want nil (unscoped), got %#v", facts[1])
	}
}

// ── The absent-facts hole (council round a06ff850, objection §3.5) ──────────

// An object entry whose `facts` key is absent or malformed used to be skipped
// before `pending`, so it never reached `unmatched` and never produced a
// durable row — indistinguishable from a page correctly assigned no facts,
// which is exactly the disobedience seed 333's measurement depends on
// catching. It must land in FactAssignmentAbsent, and NOT in FactCarryMisses:
// the two report different planner faults and the buckets must not blur (this
// test contains one of each, so a transposed return would fail it).
func TestPassB2_AbsentFactsEntryIsRecordedDistinctly(t *testing.T) {
	existing := []interface{}{
		realised("about", "/about.html", "deployed",
			`["hero-about","info-card-grid"]`, false),
	}
	llm := []interface{}{
		llmPageEntries("about", "/about.html",
			map[string]interface{}{"name": "hero-about"},                    // facts key ABSENT
			map[string]interface{}{"name": "info-card-grid", "facts": "F2"}, // facts MALFORMED (string)
			objSection("generic-text-block", "F9-not-on-this-page"),         // well-formed but unmatched
		),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	entries := planSectionsOf(t, got, "about")
	if names := sectionNamesOf(t, entries); !equalStrings(names,
		[]string{"hero-about", "info-card-grid"}) {
		t.Fatalf("realised composition was not preserved: %v", names)
	}
	if counts.SectionFactsCarried != 0 {
		t.Errorf("section_facts_carried = %d, want 0 — no usable assignment existed", counts.SectionFactsCarried)
	}

	if len(counts.FactAssignmentAbsent) != 1 {
		t.Fatalf("expected 1 page with absent-facts entries, got %#v", counts.FactAssignmentAbsent)
	}
	absent := counts.FactAssignmentAbsent[0]
	if absent.Page != "about" || !equalStrings(absent.Sections, []string{"hero-about", "info-card-grid"}) {
		t.Errorf("absent = %+v, want about/[hero-about info-card-grid] — an entry with no usable facts value must be NAMED, not skipped", absent)
	}

	if len(counts.FactCarryMisses) != 1 {
		t.Fatalf("expected 1 page with unmatched assignments, got %#v", counts.FactCarryMisses)
	}
	if miss := counts.FactCarryMisses[0]; miss.Page != "about" ||
		!equalStrings(miss.Sections, []string{"generic-text-block"}) {
		t.Errorf("miss = %+v, want about/[generic-text-block] — absent and unmatched must not blur into one bucket", miss)
	}
}

// Bare strings carry no facts key and must NOT read as disobedience: they are
// the pre-scoping emission and the no-op guarantee covers them.
func TestPassB2_BareStringRecompositionProducesNoAbsentRecords(t *testing.T) {
	existing := []interface{}{
		realised("about", "/about.html", "deployed",
			`["hero-about","info-card-grid"]`, false),
	}
	llm := []interface{}{
		llmPageEntries("about", "/about.html", "hero-about", "generic-text-block"),
	}

	_, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	if len(counts.FactAssignmentAbsent) != 0 {
		t.Errorf("bare-string entries must not be counted absent, got %#v", counts.FactAssignmentAbsent)
	}
	if len(counts.FactCarryMisses) != 0 {
		t.Errorf("bare-string entries carry no assignments to miss, got %#v", counts.FactCarryMisses)
	}
}

// ── recompose_pages outcome classification (owner ruling 2026-08-10, D3) ────

// Only the two silent shapes come back: a released page proposed verbatim (the
// seed-362 no-op gap) and a released page absent from the plan (the sanctioned
// drop that must be loud). A genuinely recomposed page produces no outcome.
func TestRecomposeOutcomes_SilentShapesReturned_RecomposedIsNot(t *testing.T) {
	recomposeRealised := map[string][]interface{}{
		"index":   {"hero", "features"},
		"about":   {"hero-about"},
		"contact": {"contact-form"},
	}
	pages := []interface{}{
		map[string]interface{}{"name": "index",
			"sections": []interface{}{"hero", "features"}}, // verbatim -> no-op
		map[string]interface{}{"name": "about",
			"sections": []interface{}{"hero-about", "team-grid"}}, // recomposed -> silent
		// contact absent -> dropped
	}

	got := recomposeOutcomes(pages, recomposeRealised)

	byPage := map[string]recomposeOutcome{}
	for _, o := range got {
		byPage[o.Page] = o
	}
	if len(got) != 2 {
		t.Fatalf("want exactly 2 outcomes (verbatim + absent), got %#v", got)
	}
	if o := byPage["index"]; o.Outcome != "proposed_verbatim" || o.RealisedSections != 2 {
		t.Errorf("index = %+v, want proposed_verbatim with 2 realised sections", o)
	}
	if o := byPage["contact"]; o.Outcome != "absent_from_plan" || o.RealisedSections != 1 {
		t.Errorf("contact = %+v, want absent_from_plan with 1 realised section", o)
	}
	if _, present := byPage["about"]; present {
		t.Errorf("a genuinely recomposed page must produce NO outcome, got %+v", byPage["about"])
	}
}
