package actions

import (
	"testing"

	"go.uber.org/zap"
)

// Twin-identity layer tests (bugs_open/215, quiet mode).
//
// Every fixture here is a shape MEASURED on the live fleet on 2026-08-11, not
// one invented to suit the code. That matters: the fix's whole claim is that
// these pairs occur, so a fixture set of plausible-looking inventions would
// prove the code self-consistent and nothing else. Provenance per fixture is in
// its comment.
//
// The population, for the record (current plans joined to realised pages, name
// differing): 3 pairs matched on normalised path, 11 on stem, 0 that a human
// read as genuinely different pages.

// realisedPage builds a realised row in the shape load_existing_pages returns.
func twinRealisedPage(name, url, pageType, buildStatus string, sections ...interface{}) map[string]interface{} {
	if sections == nil {
		sections = []interface{}{}
	}
	return map[string]interface{}{
		"name": name, "url": url, "page_type": pageType,
		"build_status": buildStatus,
		"sections":     sections,
		"title":        name + " title",
	}
}

// llmPage builds a planner-proposed page.
func twinLLMPage(name, url, pageType string, sections ...interface{}) map[string]interface{} {
	if sections == nil {
		sections = []interface{}{}
	}
	return map[string]interface{}{
		"name": name, "url": url, "page_type": pageType, "sections": sections,
	}
}

func twinNamesOf(pages []interface{}) []string {
	out := make([]string, 0, len(pages))
	for _, p := range pages {
		if m, ok := p.(map[string]interface{}); ok {
			n, _ := m["name"].(string)
			out = append(out, n)
		}
	}
	return out
}

func twinFindPage(pages []interface{}, name string) map[string]interface{} {
	for _, p := range pages {
		if m, ok := p.(map[string]interface{}); ok {
			if n, _ := m["name"].(string); n == name {
				return m
			}
		}
	}
	return nil
}

// TestReconcile_FlatNestedURLTwinSnapsToRealised — the path-key layer.
//
// MEASURED on fundamentallyai.com 2026-08-11: the current plan carries
// "tool-llm-cost-calculator" at /tools/llm-cost-calculator/index.html while the
// live page is "llm-cost-calculator" at /tools/llm-cost-calculator.html. Two
// names, two URLs, one page. Pass B misses it (URLs differ); Pass B2 misses it
// (names differ). Both URLs claim the path /tools/llm-cost-calculator.
func TestReconcile_FlatNestedURLTwinSnapsToRealised(t *testing.T) {
	llm := []interface{}{
		twinLLMPage("tool-llm-cost-calculator", "/tools/llm-cost-calculator/index.html", "tool", "hero", "calculator"),
	}
	existing := []interface{}{
		twinRealisedPage("llm-cost-calculator", "/tools/llm-cost-calculator.html", "tool", "deployed", "hero-tool", "tool-cta"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true}, zap.NewNop())

	if len(got) != 1 {
		t.Fatalf("expected exactly one page, got %v", twinNamesOf(got))
	}
	page := got[0].(map[string]interface{})
	if name, _ := page["name"].(string); name != "llm-cost-calculator" {
		t.Errorf("name = %q, want the realised spelling %q — the plan would otherwise mint a second identity",
			name, "llm-cost-calculator")
	}
	if url, _ := page["url"].(string); url != "/tools/llm-cost-calculator.html" {
		t.Errorf("url = %q, want the URL the page is actually serving from", url)
	}
	if counts.SnappedIdentityPathKey != 1 {
		t.Errorf("SnappedIdentityPathKey = %d, want 1", counts.SnappedIdentityPathKey)
	}
	// The realised composition of a deployed page wins — a built page must not
	// be re-composed (bugs_open/050 lineage).
	sections, _ := page["sections"].([]interface{})
	if len(sections) != 2 || sections[0] != "hero-tool" {
		t.Errorf("sections = %v, want the realised composition", sections)
	}
	if page["identity_authority"] != "realised" {
		t.Error("snapped page must carry the realised-identity marker, or the write path re-derives it")
	}
	if page["reconciled_from"] != "tool-llm-cost-calculator" {
		t.Errorf("reconciled_from = %v, want the planner's spelling so imagery keyed to it still resolves", page["reconciled_from"])
	}
	if len(counts.IdentitySnaps) != 1 || counts.IdentitySnaps[0].Layer != "path_key" {
		t.Errorf("expected one durable path_key snap record, got %+v", counts.IdentitySnaps)
	}
}

// TestReconcile_CanonicalSpellingTwinSnapsToLegacyRealised — the canonical layer,
// in the direction where the planner emits the BARE name and the realised page
// holds it too but under a different URL shape than the canonicaliser would pick.
//
// This is the direction that would otherwise be "self-healing" at the write path
// only by accident: canonicalisation maps the plan entry onto the realised name,
// but the URL still moves unless the identity is honoured.
func TestReconcile_CanonicalSpellingTwinSnapsToLegacyRealised(t *testing.T) {
	llm := []interface{}{
		twinLLMPage("llm-cost-calculator", "/llm-cost-calculator.html", "tool", "hero"),
	}
	existing := []interface{}{
		twinRealisedPage("tool-llm-cost-calculator", "/tools/llm-cost-calculator/index.html", "tool", "deployed", "hero-tool"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true}, zap.NewNop())

	page := twinFindPage(got, "tool-llm-cost-calculator")
	if page == nil {
		t.Fatalf("expected the plan entry to snap onto the realised identity, got %v", twinNamesOf(got))
	}
	if counts.SnappedIdentityCanonName != 1 {
		t.Errorf("SnappedIdentityCanonName = %d, want 1 (counts: %+v)", counts.SnappedIdentityCanonName, counts)
	}
}

// TestReconcile_CanonicalLayerUsesSlugLikeTheWritePath — fidelity to the collapse
// this layer claims to front-run.
//
// MEASURED on fundamentallyai.com 2026-08-11 (the PLAN_PAGE_MERGE_LOSSY rows at
// 10:21:47): a plan entry NAMED "tool-model-approach-selector-guide" canonicalised
// to the BARE "model-approach-selector-guide", because both write surfaces
// canonicalise firstNonEmpty(slug, name) and its slug said so. A layer that
// predicted from the name alone would model a collapse the writer does not
// perform — right answer by luck or wrong answer silently, depending on the entry.
func TestReconcile_CanonicalLayerUsesSlugLikeTheWritePath(t *testing.T) {
	page := twinLLMPage("tool-model-approach-selector-guide", "/guides/tool-model-approach-selector-guide.html", "blog-post", "article-body")
	page["slug"] = "model-approach-selector-guide"
	llm := []interface{}{page}
	existing := []interface{}{
		twinRealisedPage("model-approach-selector-guide", "/blog/model-approach-selector-guide.html", "blog-post", "deployed", "article-body"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true}, zap.NewNop())

	if counts.SnappedIdentityCanonName != 1 {
		t.Fatalf("SnappedIdentityCanonName = %d, want 1 — the layer must derive the canonical "+
			"name the way the write path does (slug then name), or it predicts the wrong collapse. counts=%+v",
			counts.SnappedIdentityCanonName, counts)
	}
	if twinFindPage(got, "model-approach-selector-guide") == nil {
		t.Errorf("expected the realised identity to win, got %v", twinNamesOf(got))
	}
}

// TestReconcile_StemTwin_DarkByDefault — the stem layer is OFF unless asked for,
// and while off it MEASURES rather than acts.
//
// MEASURED on fundamentallyai.com 2026-08-11, and note the DIRECTION: the plan
// carried the PREFIXED "tool-tools" (at /tools/tools/index.html) against the
// bare live "tools" (at /tools.html). Neither the path key nor the canonical
// name matches — the live page is content-typed, so it canonicalises to itself
// — and only the stem sees it.
//
// This fixture caught a real defect while being written: the layer originally
// handled only the bare-plan direction and was silently inert on this, the shape
// that actually cost fundamentallyai a phantom page.
func TestReconcile_StemTwin_DarkByDefault(t *testing.T) {
	llm := []interface{}{twinLLMPage("tool-tools", "/tools/tools/index.html", "tool", "hero")}
	existing := []interface{}{
		twinRealisedPage("tools", "/tools.html", "content", "deployed", "hero", "tool-grid"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true}, zap.NewNop())

	if counts.SnappedStemTwin != 0 {
		t.Errorf("stem layer fired while disabled: SnappedStemTwin = %d", counts.SnappedStemTwin)
	}
	if counts.StemTwinObserved != 1 {
		t.Fatalf("StemTwinObserved = %d, want 1 — a disabled layer must still measure itself, "+
			"or there is no evidence on which to decide whether to enable it", counts.StemTwinObserved)
	}
	if page := twinFindPage(got, "tool-tools"); page == nil {
		t.Error("the plan entry must be left untouched while the layer is off")
	}
	if len(counts.IdentitySnaps) != 1 || counts.IdentitySnaps[0].Layer != "stem_twin_observed" {
		t.Errorf("expected one durable observation record, got %+v", counts.IdentitySnaps)
	}
}

// TestReconcile_StemTwin_EnabledSnapsPrefixedPlanOntoBareRealised — the same
// fixture, layer enabled: the live bare page wins and the plan stops carrying a
// second identity for it.
func TestReconcile_StemTwin_EnabledSnapsPrefixedPlanOntoBareRealised(t *testing.T) {
	llm := []interface{}{twinLLMPage("tool-tools", "/tools/tools/index.html", "tool", "hero")}
	existing := []interface{}{
		twinRealisedPage("tools", "/tools.html", "content", "deployed", "hero", "tool-grid"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true, StemTwinSnap: true}, zap.NewNop())

	if counts.SnappedStemTwin != 1 {
		t.Fatalf("SnappedStemTwin = %d, want 1", counts.SnappedStemTwin)
	}
	page := twinFindPage(got, "tools")
	if page == nil {
		t.Fatalf("expected the realised identity to win, got %v", twinNamesOf(got))
	}
	if url, _ := page["url"].(string); url != "/tools.html" {
		t.Errorf("url = %q, want the URL the live page serves from", url)
	}
}

// TestReconcile_StemTwin_EnabledSnapsBareplanOntoPrefixedRealised — the OTHER
// direction, which also occurs in the same site's data on the same day: bare
// plan "ai-readiness-checker-guide" against prefixed realised
// "tool-ai-readiness-checker-guide". Both directions must work, or the layer
// closes half the population and reads as fixed.
func TestReconcile_StemTwin_EnabledSnapsBareplanOntoPrefixedRealised(t *testing.T) {
	llm := []interface{}{
		twinLLMPage("ai-readiness-checker-guide", "/blog/ai-readiness-checker-guide.html", "blog-post", "article-body"),
	}
	existing := []interface{}{
		twinRealisedPage("tool-ai-readiness-checker-guide", "/guides/tool-ai-readiness-checker-guide.html", "blog-post", "deployed", "article-body"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true, StemTwinSnap: true}, zap.NewNop())

	if counts.SnappedStemTwin != 1 {
		t.Fatalf("SnappedStemTwin = %d, want 1 — the bare-plan direction regressed", counts.SnappedStemTwin)
	}
	if twinFindPage(got, "tool-ai-readiness-checker-guide") == nil {
		t.Errorf("expected the realised identity to win, got %v", twinNamesOf(got))
	}
}

// TestReconcile_StemTwin_PrefixVsPrefixNeverFires — THE false-positive guard.
//
// This is the case the reconciler has refused to risk since 2026-07-20 and the
// reason Pass C2 stayed first-plan-only: a genuinely NEW "tool-pricing" beside a
// built "guide-pricing" shares the stem "pricing" while being a different page.
// The bare-vs-prefixed rule makes it structurally unmatchable rather than merely
// unlikely — both names carry a prefix, and the layer requires one side to be
// the bare stem.
//
// Mutation-checked: removing that rule makes this test fail.
func TestReconcile_StemTwin_PrefixVsPrefixNeverFires(t *testing.T) {
	llm := []interface{}{twinLLMPage("tool-pricing", "/tools/pricing/index.html", "tool", "hero")}
	existing := []interface{}{
		twinRealisedPage("guide-pricing", "/guides/pricing/index.html", "guide", "deployed", "article-body"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true, StemTwinSnap: true}, zap.NewNop())

	if counts.SnappedStemTwin != 0 || counts.StemTwinObserved != 0 {
		t.Fatalf("the stem layer matched two differently-prefixed pages — this suppresses a legitimately new page.\n"+
			"snapped=%d observed=%d", counts.SnappedStemTwin, counts.StemTwinObserved)
	}
	if twinFindPage(got, "tool-pricing") == nil {
		t.Errorf("the new page must survive untouched, got %v", twinNamesOf(got))
	}
}

// TestReconcile_StemTwin_RefusesWhenBothSpellingsAreInThePlan — the robot-hands
// hazard.
//
// MEASURED on robot-hands.com 2026-08-11: three pairs where BOTH spellings are
// deployed AND both are in the current plan. Snapping one onto the other would
// hand the writer two entries with one name; dedupePlanPageRows would then
// resolve the pair richer-wins, silently evicting a DEPLOYED page from plan
// governance. Which page survives is a remediation decision, not a
// reconciliation — so the layer must refuse, and say so durably.
func TestReconcile_StemTwin_RefusesWhenBothSpellingsAreInThePlan(t *testing.T) {
	llm := []interface{}{
		twinLLMPage("gripper-payload-calculator", "/gripper-payload-calculator.html", "content", "hero", "a", "b"),
		twinLLMPage("tool-gripper-payload-calculator", "/tools/gripper-payload-calculator/index.html", "tool", "hero-tool"),
	}
	existing := []interface{}{
		twinRealisedPage("tool-gripper-payload-calculator", "/tools/gripper-payload-calculator/index.html", "tool", "deployed", "hero-tool"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true, StemTwinSnap: true}, zap.NewNop())

	if counts.SnappedStemTwin != 0 {
		t.Fatalf("snapped a twin whose realised spelling the plan already carries: "+
			"the writer's dedup would then drop one of two DEPLOYED pages. counts=%+v", counts)
	}
	if counts.StemTwinAmbiguous != 1 {
		t.Errorf("StemTwinAmbiguous = %d, want 1 — a refusal must be recorded, or it is "+
			"indistinguishable from no twin having existed", counts.StemTwinAmbiguous)
	}
	if twinFindPage(got, "gripper-payload-calculator") == nil || twinFindPage(got, "tool-gripper-payload-calculator") == nil {
		t.Errorf("both entries must survive for the remediation decision, got %v", twinNamesOf(got))
	}
}

// TestReconcile_StemTwin_RefusesUnshippedTwin — a stem sibling that has never
// shipped is not evidence of identity. The phantom this layer prevents is always
// the twin of a page that IS serving; an unbuilt sibling may simply be a
// different page nobody has built yet.
func TestReconcile_StemTwin_RefusesUnshippedTwin(t *testing.T) {
	llm := []interface{}{twinLLMPage("report", "/report.html", "content", "hero")}
	existing := []interface{}{
		// needs_rebuild keeps it in the preservation set while leaving it unshipped.
		twinRealisedPage("tool-report", "/tools/report/index.html", "tool", "needs_rebuild", "hero-tool"),
	}

	_, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true, StemTwinSnap: true}, zap.NewNop())

	if counts.SnappedStemTwin != 0 {
		t.Fatalf("snapped onto a never-shipped stem sibling: counts=%+v", counts)
	}
	if counts.StemTwinAmbiguous != 1 {
		t.Errorf("StemTwinAmbiguous = %d, want 1", counts.StemTwinAmbiguous)
	}
}

// TestReconcile_TwinLayersRefuseAmbiguousKeys — two realised pages claiming one
// key means the key answers nothing. Guessing would suppress a real page; a miss
// only leaves a twin unreconciled, which is the failure this bug already has.
func TestReconcile_TwinLayersRefuseAmbiguousKeys(t *testing.T) {
	llm := []interface{}{twinLLMPage("widget", "/widget.html", "content", "hero")}
	existing := []interface{}{
		twinRealisedPage("tool-widget", "/tools/widget/index.html", "tool", "deployed", "hero-tool"),
		twinRealisedPage("guide-widget", "/guides/widget/index.html", "guide", "deployed", "article-body"),
	}

	_, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true, StemTwinSnap: true}, zap.NewNop())

	if counts.SnappedStemTwin != 0 {
		t.Fatalf("guessed between two realised pages claiming one stem: counts=%+v", counts)
	}
}

// TestReconcile_SnapCarriesFactAssignments — the 151 lane's measurement flows
// through the arm every layer now shares. A snap discards the planner's section
// entries in favour of the realised composition, so the plan-time fact
// assignments on them must be carried onto the surviving names, exactly as the
// exact-URL rename has always done.
func TestReconcile_SnapCarriesFactAssignments(t *testing.T) {
	llm := []interface{}{
		twinLLMPage("tool-llm-cost-calculator", "/tools/llm-cost-calculator/index.html", "tool",
			map[string]interface{}{"name": "hero-tool", "facts": []interface{}{"fact-1", "fact-2"}},
		),
	}
	existing := []interface{}{
		twinRealisedPage("llm-cost-calculator", "/tools/llm-cost-calculator.html", "tool", "deployed", "hero-tool"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true}, zap.NewNop())

	if counts.SectionFactsCarried != 1 {
		t.Fatalf("SectionFactsCarried = %d, want 1 — a snap must not silently drop plan-time "+
			"fact assignments (bugs_open/151)", counts.SectionFactsCarried)
	}
	page := twinFindPage(got, "llm-cost-calculator")
	if page == nil {
		t.Fatal("page missing after snap")
	}
	sections, _ := page["sections"].([]interface{})
	if len(sections) != 1 {
		t.Fatalf("sections = %v", sections)
	}
	entry, ok := sections[0].(map[string]interface{})
	if !ok || entry["name"] != "hero-tool" {
		t.Fatalf("expected the fact assignment carried onto the realised section name, got %#v", sections[0])
	}
	if facts, _ := entry["facts"].([]interface{}); len(facts) != 2 {
		t.Errorf("facts = %v, want both carried", entry["facts"])
	}
}

// TestReconcile_StripsForgedIdentityMarker — this function is the only minter of
// the marker that tells the write path to stop canonicalising. A plan that
// arrives carrying one (a replay, or an LLM echoing the schema) must not be
// believed, or any page could pin an arbitrary name and URL.
func TestReconcile_StripsForgedIdentityMarker(t *testing.T) {
	forged := twinLLMPage("attacker-chosen", "/somewhere-else.html", "content", "hero")
	forged["identity_authority"] = "realised"
	llm := []interface{}{forged}
	existing := []interface{}{
		twinRealisedPage("home-page", "/index.html", "landing", "deployed", "hero"),
	}

	got, _ := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true}, zap.NewNop())

	page := twinFindPage(got, "attacker-chosen")
	if page == nil {
		t.Fatalf("fixture page missing: %v", twinNamesOf(got))
	}
	if _, present := page["identity_authority"]; present {
		t.Error("an LLM-supplied identity_authority marker survived — the write path would honour a forged identity")
	}
}

// TestReconcile_ParentSectionCarriedFromRealisedURL — without this, snapping is
// worse than not snapping.
//
// MEASURED on fundamentallyai.com 2026-08-11: "tool-ai-readiness-checker-guide"
// is a blog-post-typed page SERVING from /guides/. CanonicalisePage puts
// blog-posts under /blog/ unless told otherwise, so a snapped entry that does
// not carry its parent section is re-derived to /blog/... and upsertPage moves
// the live URL — turning a duplicate-page bug into a URL-move bug
// (bugs_open/241's damage shape).
func TestReconcile_ParentSectionCarriedFromRealisedURL(t *testing.T) {
	llm := []interface{}{
		twinLLMPage("ai-readiness-checker-guide", "/blog/ai-readiness-checker-guide.html", "blog-post", "article-body"),
	}
	existing := []interface{}{
		twinRealisedPage("tool-ai-readiness-checker-guide", "/guides/tool-ai-readiness-checker-guide.html", "blog-post", "deployed", "article-body"),
	}

	got, _ := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true, StemTwinSnap: true}, zap.NewNop())

	page := twinFindPage(got, "tool-ai-readiness-checker-guide")
	if page == nil {
		t.Fatalf("expected a snap onto the realised identity, got %v", twinNamesOf(got))
	}
	if parent, _ := page["parent_section"].(string); parent != "guides" {
		t.Errorf("parent_section = %q, want %q — otherwise the canonicaliser re-derives this "+
			"page's URL under /blog/ and the live page moves", parent, "guides")
	}
}

// TestReconcile_UnionedRealisedPageCarriesItsIdentity — Pass A's union has the
// same exposure as a snap: a preserved realised page the planner omitted is
// re-added, then re-canonicalised at the write path, which can rename or move it.
// The marker is minted in normaliseRealisedToPlanPage precisely so the union is
// covered too, not only the snap.
func TestReconcile_UnionedRealisedPageCarriesItsIdentity(t *testing.T) {
	llm := []interface{}{twinLLMPage("about", "/about.html", "content", "hero")}
	existing := []interface{}{
		twinRealisedPage("llm-cost-calculator", "/tools/llm-cost-calculator.html", "tool", "deployed", "hero-tool"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true}, zap.NewNop())

	if counts.Unioned != 1 {
		t.Fatalf("Unioned = %d, want 1", counts.Unioned)
	}
	page := twinFindPage(got, "llm-cost-calculator")
	if page == nil {
		t.Fatalf("unioned page missing: %v", twinNamesOf(got))
	}
	if page["identity_authority"] != "realised" {
		t.Error("a unioned realised page must carry its stored identity, or the write path " +
			"re-derives it and mints a twin of a page it was added to preserve")
	}
}

// TestReconcile_DeterministicLayersAreGatedAndCountWhenOff — the council's
// guardian and architecture seats both objected that these two layers shipped
// default-ON, changing matching behaviour for every existing caller fleet-wide
// on deploy, while the weaker stem layer got a dark launch. They were right: an
// argument that "Pass B already asks this question" is not a measurement of the
// new collapse population, and a behaviour change for existing callers is
// architecture-scope however sound the reasoning.
//
// So: off by default, and counting while off. The fixture is the same real
// fundamentallyai pair as the path-key test above — the point is only that the
// gate decides whether it acts.
func TestReconcile_DeterministicLayersAreGatedAndCountWhenOff(t *testing.T) {
	llm := []interface{}{
		twinLLMPage("tool-llm-cost-calculator", "/tools/llm-cost-calculator/index.html", "tool", "hero"),
	}
	existing := []interface{}{
		twinRealisedPage("llm-cost-calculator", "/tools/llm-cost-calculator.html", "tool", "deployed", "hero-tool"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	if counts.SnappedIdentityPathKey != 0 || counts.SnappedIdentityCanonName != 0 {
		t.Fatalf("a deterministic layer fired with twin_identity_snap off: %+v", counts)
	}
	if counts.TwinIdentityObserved != 1 {
		t.Fatalf("TwinIdentityObserved = %d, want 1 — a gated layer must still measure what it "+
			"would have done, or there is no evidence on which to decide whether to enable it",
			counts.TwinIdentityObserved)
	}
	if twinFindPage(got, "tool-llm-cost-calculator") == nil {
		t.Errorf("the plan entry must be left untouched while the layer is off, got %v", twinNamesOf(got))
	}
	if len(counts.IdentitySnaps) != 1 || counts.IdentitySnaps[0].Layer != "path_key_observed" {
		t.Errorf("expected one durable observation record naming the layer, got %+v", counts.IdentitySnaps)
	}
}

// TestReconcile_MarkerSurvivesTheStepBoundary answers the council's editquality
// objection directly, and it is the objection that mattered most: if the
// realised-identity marker cannot travel from the reconciler (validate_plan) to
// the two write surfaces, the writer guard silently never fires — "the exact
// silent guard indistinguishable from a dead one" the seat quoted back at me.
//
// It does travel, and NOT through site_plan_pages (which has no column for it).
// The live build-site-planner workflow wires validate_plan's output_field to
// "site_plan"; both write surfaces then read that key out of collected_data via
// extractPagesFromPlan, which appends each page map WHOLE rather than rebuilding
// it from known fields. This test drives the real extractor over a
// validate-shaped payload, so it fails if anyone makes that extraction
// field-selective.
// EXTENDED 2026-08-19 (bugs_open/215, the same-name hole): the second page here is
// one the reconciler paired by exact NAME rather than by a snap. Before the B2
// stamp existed it travelled this same boundary carrying no marker at all, so this
// test passed while honour_realised_identity was unreachable for it — the test
// proved transport, which was never the thing that was broken. Both routes are
// now driven, so stamping a snap but not a same-name pairing (or vice versa) fails
// here rather than in production.
func TestReconcile_MarkerSurvivesTheStepBoundary(t *testing.T) {
	llm := []interface{}{
		twinLLMPage("tool-llm-cost-calculator", "/tools/llm-cost-calculator/index.html", "tool", "hero"),
		// Named exactly as stored, and with no url — validate_plan's ordinary shape.
		map[string]interface{}{
			"name": "mortgages-stamp-duty", "page_type": "tool", "sections": []interface{}{},
		},
	}
	existing := []interface{}{
		twinRealisedPage("llm-cost-calculator", "/tools/llm-cost-calculator.html", "tool", "deployed", "hero-tool"),
		twinRealisedPage("mortgages-stamp-duty", "/mortgages/mortgages-stamp-duty/index.html", "tool", "deployed", "hero"),
	}
	reconciled, _ := reconcilePlanWithRealised(llm, existing, reconcileOptions{TwinIdentitySnap: true}, zap.NewNop())

	// Exactly the shape validate_plan hands to collected_data under "site_plan".
	collected := map[string]interface{}{
		"site_plan": map[string]interface{}{"pages": reconciled},
	}
	pages := extractPagesFromPlan(collected, zap.NewNop())
	if len(pages) != 2 {
		t.Fatalf("extractPagesFromPlan returned %d pages, want 2", len(pages))
	}

	want := map[string]string{
		"llm-cost-calculator":  "/tools/llm-cost-calculator.html",
		"mortgages-stamp-duty": "/mortgages/mortgages-stamp-duty/index.html",
	}
	seen := map[string]bool{}
	for _, p := range pages {
		name, url, pageType, ok := realisedIdentityOf(p)
		if !ok {
			pn, _ := p["name"].(string)
			t.Fatalf("the realised-identity marker did not survive the step boundary for %q — the writer "+
				"guard would never fire and the twin would be re-minted. page keys: %v", pn,
				func() []string {
					out := []string{}
					for k := range p {
						out = append(out, k)
					}
					return out
				}())
		}
		wantURL, known := want[name]
		if !known {
			t.Fatalf("unexpected identity %q survived the boundary", name)
		}
		if url != wantURL || pageType != "tool" {
			t.Errorf("identity %q survived but altered: %q %q", name, url, pageType)
		}
		seen[name] = true
	}
	if len(seen) != 2 {
		t.Errorf("both the snapped and the same-name-paired identity must survive, saw %v", seen)
	}
}

// ── The same-name identity stamp (bugs_open/215, the hole found 2026-08-19) ──
//
// PROVENANCE, and it is a live incident rather than a shape anyone imagined:
// loanandmortgagecalculator.co.uk seeded honour_realised_identity='true' and fired
// one canary re-plan on 2026-08-17 (corr 6fe6ee93-67b9-4831-bf17-2ca473e1d30c,
// chassis v1.0.1305). 19 phantom page rows were INSERTed anyway, 17 of them the
// predicted tool-<name> twins of pages the planner had named CORRECTLY — verbatim
// as stored. The plan entries carried no url, no parent_section and sections: [],
// which is the ordinary shape of validate_plan's output.
//
// The three twin layers could not help: their shared eligible closure refuses a
// realised candidate whose name equals the plan name, and it is right to (such a
// candidate is the page itself, not a twin). Pass B2 paired them and stamped
// nothing. So realisedIdentityOf returned ok=false, the flag never fired, and
// CanonicalisePage re-derived each page at its role's default hub.
//
// Every fixture below uses the real stored names from that site.

// TestReconcile_SameNamePagePlannedVerbatimIsStampedWithStoredIdentity is the
// defect in one page. All layers OFF, because the stamp must not need a gate —
// the same reason Pass A's union does not have one.
func TestReconcile_SameNamePagePlannedVerbatimIsStampedWithStoredIdentity(t *testing.T) {
	llm := []interface{}{
		map[string]interface{}{
			"name": "mortgages-stamp-duty", "page_type": "tool",
			"title": "Stamp Duty Calculator — refreshed copy",
			// No url: the shape validate_plan actually emits.
			"sections": []interface{}{},
		},
	}
	existing := []interface{}{
		twinRealisedPage("mortgages-stamp-duty", "/mortgages/mortgages-stamp-duty/index.html",
			"tool", "deployed", "hero", "calculator"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	page := twinFindPage(got, "mortgages-stamp-duty")
	if page == nil {
		t.Fatalf("the page vanished from the plan: %v", twinNamesOf(got))
	}
	name, url, pageType, ok := realisedIdentityOf(page)
	if !ok {
		t.Fatalf("no realised identity stamped — honour_realised_identity cannot fire for this page, " +
			"and the write path will mint tool-mortgages-stamp-duty as a second identity (the 08-17 incident)")
	}
	if name != "mortgages-stamp-duty" || url != "/mortgages/mortgages-stamp-duty/index.html" || pageType != "tool" {
		t.Errorf("stamped identity is not the stored one: %q %q %q", name, url, pageType)
	}
	// Pass B2's own contract must survive the stamp: composition restored...
	if secs, _ := page["sections"].([]interface{}); len(secs) != 2 {
		t.Errorf("realised composition not restored, got %v", page["sections"])
	}
	// ...and the planner's copy still wins, which is why this is not a snap.
	if title, _ := page["title"].(string); title != "Stamp Duty Calculator — refreshed copy" {
		t.Errorf("the LLM's title was overwritten (%q) — a re-plan must still be able to refresh copy", title)
	}
	if len(counts.SameNameStamps) != 1 {
		t.Fatalf("expected one same-name stamp recorded, got %+v", counts.SameNameStamps)
	}
	if counts.SameNameStamps[0].WouldDeriveName != "tool-mortgages-stamp-duty" {
		t.Errorf("the prevented twin name is not recorded: %+v", counts.SameNameStamps[0])
	}
	if counts.SameNameStamps[0].StoredURL != "/mortgages/mortgages-stamp-duty/index.html" {
		t.Errorf("stored URL not carried for post-expiry diagnosis: %+v", counts.SameNameStamps[0])
	}
}

// TestReconcile_SameNameStamp_FixedPointIsStampedButNotDiverging: a page whose
// stored name IS the canonicaliser's answer is stamped like any other (the stamp
// is a no-op at the writers), but must NOT be reported as a prevented twin — the
// durable record would otherwise claim damage it did not prevent.
func TestReconcile_SameNameStamp_FixedPointIsStampedButNotDiverging(t *testing.T) {
	llm := []interface{}{
		map[string]interface{}{"name": "tool-repayment", "page_type": "tool", "sections": []interface{}{}},
	}
	existing := []interface{}{
		twinRealisedPage("tool-repayment", "/tools/repayment/index.html", "tool", "deployed", "hero"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	if _, _, _, ok := realisedIdentityOf(twinFindPage(got, "tool-repayment")); !ok {
		t.Fatalf("a fixed-point page should still be stamped (harmless, and one rule is better than two)")
	}
	if len(counts.SameNameStamps) != 1 || counts.SameNameStamps[0].WouldDeriveName != "" {
		t.Errorf("a fixed-point stamp must not be recorded as diverging: %+v", counts.SameNameStamps)
	}
	if findings := buildSameNameIdentityFindings(counts, false); len(findings) != 0 {
		t.Errorf("a run of fixed-point stamps must write no durable rows, got %d", len(findings))
	}
}

// TestReconcile_SameNameStamp_RefusesTypeConflict: one name, two roles. Honouring
// would silently retype a live page; re-deriving keeps today's behaviour. Refuse,
// and leave a durable record — the type is the one stamped field that is NOT
// inert when the flag is off, because it feeds the writers' Role.
func TestReconcile_SameNameStamp_RefusesTypeConflict(t *testing.T) {
	llm := []interface{}{
		map[string]interface{}{"name": "mortgages-simple", "page_type": "content", "sections": []interface{}{}},
	}
	existing := []interface{}{
		twinRealisedPage("mortgages-simple", "/mortgages/mortgages-simple/index.html", "tool", "deployed", "hero"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	page := twinFindPage(got, "mortgages-simple")
	if _, ok := page["identity_authority"]; ok {
		t.Errorf("a type conflict must not be stamped — honouring it would retype a live page")
	}
	if pt, _ := page["page_type"].(string); pt != "content" {
		t.Errorf("the plan's page_type was overwritten on a refusal: %q", pt)
	}
	if _, ok := page["url"]; ok {
		t.Errorf("url stamped despite the refusal: %v", page["url"])
	}
	if len(counts.SameNameTypeConflicts) != 1 {
		t.Fatalf("the refusal was silent — expected one durable conflict record, got %+v", counts.SameNameTypeConflicts)
	}
	c := counts.SameNameTypeConflicts[0]
	if c.PlanType != "content" || c.RealisedType != "tool" {
		t.Errorf("both types must be recorded to tell disobedience from a retype: %+v", c)
	}
	if len(counts.SameNameStamps) != 0 {
		t.Errorf("a refused pair must not be counted as a stamp: %+v", counts.SameNameStamps)
	}
}

// TestReconcile_SameNameStamp_TypeUnderTheRoleKeyIsNotAConflict: the writers
// derive Role via firstNonEmptyField(page_type, type, role), so an entry carrying
// its type under "role" agrees with the realised page and must be stamped. Testing
// page_type alone would have called this a conflict and left the page unprotected.
func TestReconcile_SameNameStamp_TypeUnderTheRoleKeyIsNotAConflict(t *testing.T) {
	llm := []interface{}{
		map[string]interface{}{"name": "mortgages-repayment", "role": "tool", "sections": []interface{}{}},
	}
	existing := []interface{}{
		twinRealisedPage("mortgages-repayment", "/mortgages/mortgages-repayment/index.html", "tool", "deployed", "hero"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	if len(counts.SameNameTypeConflicts) != 0 {
		t.Fatalf("a type supplied under \"role\" is the type the writers will use, not a conflict: %+v",
			counts.SameNameTypeConflicts)
	}
	_, _, pageType, ok := realisedIdentityOf(twinFindPage(got, "mortgages-repayment"))
	if !ok || pageType != "tool" {
		t.Errorf("expected the stored identity stamped with page_type tool, got ok=%v type=%q", ok, pageType)
	}
}

// TestReconcile_SameNameStamp_RefusesIncompleteStoredIdentity mirrors
// realisedIdentityOf's own refusal. Claiming a stamp the reader rejects would put
// a lie in the counters — a guard that never fires reads exactly like one that is
// not there.
func TestReconcile_SameNameStamp_RefusesIncompleteStoredIdentity(t *testing.T) {
	llm := []interface{}{
		map[string]interface{}{"name": "mortgages-stamp-duty", "page_type": "tool", "sections": []interface{}{}},
	}
	existing := []interface{}{
		twinRealisedPage("mortgages-stamp-duty", "", "tool", "deployed", "hero"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	if _, ok := twinFindPage(got, "mortgages-stamp-duty")["identity_authority"]; ok {
		t.Errorf("stamped an identity realisedIdentityOf would reject (empty url)")
	}
	if len(counts.SameNameStamps) != 0 {
		t.Errorf("counted a stamp that did not happen: %+v", counts.SameNameStamps)
	}
}

// TestReconcile_SameNameStamp_CoversTheB2FallThrough: identity and composition are
// separate questions. A catalogued page that has never shipped keeps the planner's
// proposed sections (bugs_open/050) — and must still be stamped, because who owns
// its NAME does not depend on whether it has been built.
func TestReconcile_SameNameStamp_CoversTheB2FallThrough(t *testing.T) {
	llm := []interface{}{
		map[string]interface{}{
			"name": "mortgages-overpayment", "page_type": "tool",
			"sections": []interface{}{"hero", "calculator"},
		},
	}
	existing := []interface{}{
		twinRealisedPage("mortgages-overpayment", "/mortgages/mortgages-overpayment/index.html",
			"tool", "needs_rebuild"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	page := twinFindPage(got, "mortgages-overpayment")
	if secs, _ := page["sections"].([]interface{}); len(secs) != 2 {
		t.Errorf("a never-composed page must keep the planner's proposal: %v", page["sections"])
	}
	if _, _, _, ok := realisedIdentityOf(page); !ok {
		t.Fatalf("the fall-through branch was not stamped — the stamp must not live inside one branch")
	}
	if len(counts.SameNameStamps) != 1 {
		t.Errorf("expected the stamp recorded, got %+v", counts.SameNameStamps)
	}
}

// TestReconcile_TwinLayersRefuseTheEntryItself pins that no twin layer snaps a
// page onto ITSELF — a behaviour nothing pinned before, and the one the 2026-08-19
// investigation proposed relaxing.
//
// WHAT IT PINS, STATED EXACTLY, because I first wrote that it pinned the eligible
// closure's rname == lname clause and the mutation run proved otherwise: this test
// pins the REFUSAL, not either clause. Removing rname == lname alone leaves it
// green (planNames[rname] refuses the same case, since a candidate equal to the
// plan's own name is necessarily a name the plan carries); removing planNames
// alone likewise; removing BOTH turns it red. The two clauses are redundant in
// series — which is also why the 08-19 remedy "drop rname == lname from the
// canonical layer" would have changed nothing at all.
//
// The refusal is correct: a realised candidate whose name equals the plan name is
// the page itself, not a twin of it. Routing it through the layers would snap the
// entry onto normaliseRealisedToPlanPage's whole-map replacement and throw away
// the planner's refreshed copy for every correctly-named page on the site. The
// same pairing is handled by Pass B2's stamp instead, which keeps the copy.
func TestReconcile_TwinLayersRefuseTheEntryItself(t *testing.T) {
	llm := []interface{}{
		map[string]interface{}{
			"name": "mortgages-stamp-duty", "page_type": "tool",
			"title": "refreshed by this re-plan", "sections": []interface{}{},
		},
	}
	existing := []interface{}{
		twinRealisedPage("mortgages-stamp-duty", "/mortgages/mortgages-stamp-duty/index.html",
			"tool", "deployed", "hero"),
	}

	// Both deterministic layers ON: the canonical layer's key
	// (tool-mortgages-stamp-duty) is claimed by this very row, so only the
	// rname == lname clause stops it snapping.
	got, counts := reconcilePlanWithRealised(llm, existing,
		reconcileOptions{TwinIdentitySnap: true, StemTwinSnap: true}, zap.NewNop())

	if counts.SnappedIdentityCanonName != 0 || counts.SnappedIdentityPathKey != 0 || counts.SnappedStemTwin != 0 {
		t.Errorf("a twin layer snapped the page onto itself: canon=%d path=%d stem=%d",
			counts.SnappedIdentityCanonName, counts.SnappedIdentityPathKey, counts.SnappedStemTwin)
	}
	if counts.TwinIdentityObserved != 0 {
		t.Errorf("the entry itself was recorded as an observed twin (%d) — that would report the "+
			"good case as damage on every re-plan", counts.TwinIdentityObserved)
	}
	page := twinFindPage(got, "mortgages-stamp-duty")
	if title, _ := page["title"].(string); title != "refreshed by this re-plan" {
		t.Errorf("the planner's copy was replaced by the realised row (%q) — this is what "+
			"routing same-name pairs through a snap would cost", title)
	}
	if _, _, _, ok := realisedIdentityOf(page); !ok {
		t.Errorf("the page still needs its identity — via the B2 stamp, not via a snap")
	}
}

// TestReconcile_SameNameStamp_ForgedMarkerIsReplacedByStoredValues: the reconciler
// is the only minter. A plan that arrives carrying its own marker and a chosen URL
// must not be believed — strip first, then stamp from the DB row.
func TestReconcile_SameNameStamp_ForgedMarkerIsReplacedByStoredValues(t *testing.T) {
	llm := []interface{}{
		map[string]interface{}{
			"name": "mortgages-stamp-duty", "page_type": "tool",
			"identity_authority": "realised",
			"url":                "/somewhere-the-planner-chose.html",
			"sections":           []interface{}{},
		},
	}
	existing := []interface{}{
		twinRealisedPage("mortgages-stamp-duty", "/mortgages/mortgages-stamp-duty/index.html",
			"tool", "deployed", "hero"),
	}

	got, _ := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	_, url, _, ok := realisedIdentityOf(twinFindPage(got, "mortgages-stamp-duty"))
	if !ok {
		t.Fatalf("expected the stored identity to be stamped after the forged one was stripped")
	}
	if url != "/mortgages/mortgages-stamp-duty/index.html" {
		t.Errorf("the planner's chosen URL survived: %q", url)
	}
}

// TestBuildSameNameIdentityFindings pins the durable record's shape — the only
// thing a later reader will have, since a chassis pod retains under a second of
// log. The two codes are the same observation on opposite sides of the flag.
func TestBuildSameNameIdentityFindings(t *testing.T) {
	counts := reconcileCounts{
		SameNameStamps: []sameNameStamp{
			{PlanName: "mortgages-stamp-duty", StoredURL: "/mortgages/mortgages-stamp-duty/index.html",
				WouldDeriveName: "tool-mortgages-stamp-duty"},
			{PlanName: "tool-repayment", StoredURL: "/tools/repayment/index.html"}, // fixed point
		},
	}

	held := buildSameNameIdentityFindings(counts, true)
	if len(held) != 1 || held[0].ErrorCode != "PLAN_PAGE_SAME_NAME_IDENTITY_HELD" || held[0].Severity != "info" {
		t.Fatalf("flag ON should record one info-level HELD row, got %+v", held)
	}
	if n, _ := held[0].Context["diverging_pages"].(int); n != 1 {
		t.Errorf("only the diverging stamp counts as prevented damage: %+v", held[0].Context)
	}
	pairs, _ := held[0].Context["pages"].([]map[string]interface{})
	if len(pairs) != 1 || pairs[0]["would_derive_name"] != "tool-mortgages-stamp-duty" {
		t.Errorf("the row must carry the pairs so the reader can join them against pages: %+v", held[0].Context)
	}

	pending := buildSameNameIdentityFindings(counts, false)
	if len(pending) != 1 || pending[0].ErrorCode != "PLAN_PAGE_SAME_NAME_TWIN_PENDING" || pending[0].Severity != "warning" {
		t.Fatalf("flag OFF is a warning about twins about to be written, got %+v", pending)
	}

	conflicts := buildSameNameIdentityFindings(reconcileCounts{
		SameNameTypeConflicts: []sameNameTypeConflict{
			{PlanName: "mortgages-simple", PlanType: "content", RealisedType: "tool"},
		},
	}, true)
	if len(conflicts) != 1 || conflicts[0].ErrorCode != "PLAN_PAGE_IDENTITY_TYPE_CONFLICT" ||
		conflicts[0].Severity != "warning" {
		t.Fatalf("a refused pair needs its own durable row, got %+v", conflicts)
	}

	if got := buildSameNameIdentityFindings(reconcileCounts{}, true); len(got) != 0 {
		t.Errorf("nothing observed must write nothing, got %+v", got)
	}
}

// TestReconcile_LMCCanaryShape_NoTwinIsMintable reproduces the 2026-08-17 incident
// at the unit level: the exact plan shape validate_plan emitted (names verbatim,
// no url, no parent_section, sections: []), the exact realised rows, every layer
// off — and drives it through the REAL extractor to the reader the write surfaces
// use. Before the stamp this test fails on every non-fixed-point page; the 19
// phantom rows are what that failure looked like in production.
func TestReconcile_LMCCanaryShape_NoTwinIsMintable(t *testing.T) {
	planned := []string{"mortgages-stamp-duty", "mortgages-simple", "mortgages-repayment", "tool-repayment"}
	llm := make([]interface{}, 0, len(planned))
	for _, n := range planned {
		llm = append(llm, map[string]interface{}{
			"name": n, "page_type": "tool", "sections": []interface{}{},
		})
	}
	existing := []interface{}{
		twinRealisedPage("mortgages-stamp-duty", "/mortgages/mortgages-stamp-duty/index.html", "tool", "deployed", "hero"),
		twinRealisedPage("mortgages-simple", "/mortgages/mortgages-simple/index.html", "tool", "deployed", "hero"),
		twinRealisedPage("mortgages-repayment", "/mortgages/mortgages-repayment/index.html", "tool", "deployed", "hero"),
		twinRealisedPage("tool-repayment", "/tools/repayment/index.html", "tool", "deployed", "hero"),
	}

	reconciled, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	// Through the step boundary exactly as the workflow wires it.
	pages := extractPagesFromPlan(map[string]interface{}{
		"site_plan": map[string]interface{}{"pages": reconciled},
	}, zap.NewNop())
	if len(pages) != len(planned) {
		t.Fatalf("extractPagesFromPlan returned %d pages, want %d", len(pages), len(planned))
	}

	wantURL := map[string]string{
		"mortgages-stamp-duty": "/mortgages/mortgages-stamp-duty/index.html",
		"mortgages-simple":     "/mortgages/mortgages-simple/index.html",
		"mortgages-repayment":  "/mortgages/mortgages-repayment/index.html",
		"tool-repayment":       "/tools/repayment/index.html",
	}
	for _, p := range pages {
		n, _, _, ok := realisedIdentityOf(p)
		if !ok {
			pn, _ := p["name"].(string)
			t.Fatalf("page %q reached the write surfaces with no realised identity: honour_realised_identity "+
				"cannot fire and CanonicalisePage will mint its twin (the 08-17 shape, 19 rows)", pn)
		}
		_, url, _, _ := realisedIdentityOf(p)
		if url != wantURL[n] {
			t.Errorf("page %q honours %q, want the stored %q", n, url, wantURL[n])
		}
	}

	// Three of the four are not fixed points; the fourth already is.
	var diverging []string
	for _, s := range counts.SameNameStamps {
		if s.WouldDeriveName != "" {
			diverging = append(diverging, s.WouldDeriveName)
		}
	}
	if len(diverging) != 3 {
		t.Errorf("expected 3 prevented twins (tool-<name> for the three bare pages), got %v", diverging)
	}
	findings := buildSameNameIdentityFindings(counts, true)
	if len(findings) != 1 || findings[0].ErrorCode != "PLAN_PAGE_SAME_NAME_IDENTITY_HELD" {
		t.Errorf("expected one durable HELD row for the run, got %+v", findings)
	}
}
