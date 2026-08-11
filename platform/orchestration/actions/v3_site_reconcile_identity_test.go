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

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

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

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

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

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

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

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

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

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{StemTwinSnap: true}, zap.NewNop())

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

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{StemTwinSnap: true}, zap.NewNop())

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

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{StemTwinSnap: true}, zap.NewNop())

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

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{StemTwinSnap: true}, zap.NewNop())

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

	_, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{StemTwinSnap: true}, zap.NewNop())

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

	_, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{StemTwinSnap: true}, zap.NewNop())

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

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

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

	got, _ := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

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

	got, _ := reconcilePlanWithRealised(llm, existing, reconcileOptions{StemTwinSnap: true}, zap.NewNop())

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

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

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
