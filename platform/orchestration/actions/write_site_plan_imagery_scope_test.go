// FILE: platform/orchestration/actions/write_site_plan_imagery_scope_test.go
//
// bugs_open/214 — the planner keys imagery by page name, WriteSitePlanAction
// renames the page via CanonicalisePage, and nothing carried the reference
// across. The imagery row then named a page its own plan did not contain, and
// no build ever referenced it: the asset was planned, generated, deployed, paid
// for, and pointed at by nothing.
//
// Three layers, deliberately:
//
//  1. TestCanonicalisePage_RenamesTheNamesPlannersUse pins the PREMISE — that
//     the very spellings a planner writes really are renamed underneath it. If
//     the canonicaliser stops collapsing these, the resolution below becomes
//     dead code and this test says so rather than leaving it quietly inert.
//     (Same reasoning as write_site_plan_dedup_test.go's premise test.)
//  2. The buildCanonicalPageNameMap / canonicaliseImageryScopeRefs /
//     dedupeImageryRows tests cover the FIX's logic in isolation.
//  3. TestFlattenImageryBlock_ScopeRefsAreCanonicalisedEndToEnd covers the
//     flatten-to-resolve JOIN over a real planner-shaped payload.
//
// NOTE, and it is the point of the sibling file: EVERY test here calls the
// helpers directly, so none of them can see whether WriteSitePlanAction still
// CALLS them. Measured by mutation on 2026-08-10 — delete the resolution block
// from the action and all fifteen still pass, including the "end to end" one,
// which is end-to-end across the helpers and not across the action. The wiring
// guard therefore lives in write_site_plan_imagery_wiring_test.go, which drives
// the real action under sqlmock and asserts the value that reaches the INSERT.
// (This file first claimed the guard was here. It was not, and that claim was
// the same error WRONG_CALLS logged on this tree the same day.)
//
// Fixtures use the live 2026-08-10 shapes: gamesdesign's about -> about-index
// with its four icons at about:2, and fundamentallyai's news -> news-index.
package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// TestCanonicalisePage_RenamesTheNamesPlannersUse asserts the premise: the
// page names planners actually key imagery by are renamed by canonicalisation.
func TestCanonicalisePage_RenamesTheNamesPlannersUse(t *testing.T) {
	cases := []struct {
		what string
		in   datahelpers.PageDescriptor
		want string
	}{
		{"about (gamesdesign, live)", datahelpers.PageDescriptor{Role: "section-index", Slug: "about"}, "about-index"},
		{"contact (gamesdesign, live)", datahelpers.PageDescriptor{Role: "section-index", Slug: "contact"}, "contact-index"},
		{"news (fundamentallyai, live)", datahelpers.PageDescriptor{Role: "news-index", Slug: "news"}, "news-index"},
		{"home -> index (brand-update routing depends on it)", datahelpers.PageDescriptor{Role: "content", Slug: "home"}, "index"},
	}
	for _, tc := range cases {
		t.Run(tc.what, func(t *testing.T) {
			got, _, _ := datahelpers.CanonicalisePage(tc.in)
			if got != tc.want {
				t.Fatalf("expected %q to canonicalise to %q, got %q\n"+
					"if this changed, the rename this fix compensates for changed too — re-check the fix is still needed",
					tc.in.Slug, tc.want, got)
			}
		})
	}
}

// TestNormalisePageKey_MatchesTheNormalisationValidateRolesApplies pins the
// coupling normalisePageKey documents. planPageRow.RawName is whatever
// ValidateRoles produced; normalisePageKey routes through CanonicalisePage's
// content branch to reach the same cleanup without copying it. If those two ever
// diverge, imagery refs would be looked up under a key the map was never built
// with — a silent miss — so this asserts they agree, through exported API only.
func TestNormalisePageKey_MatchesTheNormalisationValidateRolesApplies(t *testing.T) {
	spellings := []string{
		"about", "Who We Help", "who-we-help.html", "  contact  ",
		"/pages/services", "Case-Studies", "tools",
	}
	for _, s := range spellings {
		validated := datahelpers.ValidateRoles([]datahelpers.LLMPlannedPage{{Name: s, Role: "content"}})
		if len(validated) != 1 {
			t.Fatalf("ValidateRoles(%q) returned %d entries", s, len(validated))
		}
		if got, want := normalisePageKey(s), validated[0].Name; got != want {
			t.Errorf("normalisePageKey(%q) = %q but ValidateRoles produced RawName %q\n"+
				"these must agree or an imagery ref is looked up under a key the map was never built with",
				s, got, want)
		}
	}

	// The one deliberate divergence, asserted so it stays deliberate: the
	// homepage collapse runs before the default branch, so "home" resolves to
	// "index" — the name the plan actually writes for the homepage.
	if got := normalisePageKey("home"); got != "index" {
		t.Errorf("normalisePageKey(\"home\") = %q, want \"index\" — the homepage collapse is what lets a planner's \"home\" imagery find the real page", got)
	}
}

// imgRow builds a page/section imagery row the way flattenImageryBlock does.
func imgRow(scope, ref, key string) imageryRow {
	r := imageryRow{Scope: scope, Key: key, Kind: "hero", Prompt: "p", Source: "llm"}
	if ref != "" {
		v := ref
		r.ScopeRef = &v
		r.RawScopeRef = ref
	}
	return r
}

func refOf(t *testing.T, r imageryRow) string {
	t.Helper()
	if r.ScopeRef == nil {
		return "<nil>"
	}
	return *r.ScopeRef
}

// ---------------------------------------------------------------------------
// buildCanonicalPageNameMap
// ---------------------------------------------------------------------------

func TestBuildCanonicalPageNameMap_AllSpellingsResolve(t *testing.T) {
	pages := []planPageRow{
		{RawName: "about", Slug: "about", Name: "about-index", Sections: sec("hero", "content-block-about", "differentiators")},
		{RawName: "home", Slug: "home", Name: "index", Sections: sec("hero")},
	}
	m := buildCanonicalPageNameMap(pages, zap.NewNop())

	for _, tc := range []struct{ in, want string }{
		{"about-index", "about-index"}, // identity: an already-canonical ref must resolve
		{"about", "about-index"},       // the raw alias — the whole point
		{"index", "index"},
		{"home", "index"},
	} {
		if got := m[normalisePageKey(tc.in)]; got != tc.want {
			t.Errorf("map[%q] = %q, want %q", tc.in, got, tc.want)
		}
	}
	if _, ok := m["team"]; ok {
		t.Error("a page the plan does not contain must not resolve — otherwise an invented ref is silently 'fixed' onto a real page")
	}
}

func TestBuildCanonicalPageNameMap_NormalisesTheKeyTheSameWayValidateRolesDid(t *testing.T) {
	// RawName is normaliseSlug'd by ValidateRoles, so a planner key carrying
	// caps, spaces or a .html tail must still find it. If normalisePageKey ever
	// diverges from what ValidateRoles applies, this is what catches it.
	pages := []planPageRow{{RawName: "who-we-help", Slug: "who-we-help", Name: "who-we-help"}}
	m := buildCanonicalPageNameMap(pages, zap.NewNop())

	for _, spelling := range []string{"Who We Help", "who-we-help.html", "  who-we-help  ", "/pages/who-we-help"} {
		if got := m[normalisePageKey(spelling)]; got != "who-we-help" {
			t.Errorf("planner spelling %q resolved to %q, want %q", spelling, got, "who-we-help")
		}
	}
}

func TestBuildCanonicalPageNameMap_AmbiguousAliasIsRefusedNotGuessed(t *testing.T) {
	// Two pages whose raw spellings both claim "about". Resolving it either way
	// would move imagery onto a page nobody asked for. Refusing sends the ref
	// down the miss path, which keeps today's behaviour and leaves a record.
	pages := []planPageRow{
		{RawName: "about", Name: "about-index"},
		{RawName: "about", Name: "about-us"},
	}
	m := buildCanonicalPageNameMap(pages, zap.NewNop())

	if got, ok := m["about"]; ok {
		t.Fatalf("ambiguous alias resolved to %q; a confident wrong rewrite is worse than none", got)
	}
	if m["about-index"] != "about-index" || m["about-us"] != "about-us" {
		t.Error("the ambiguity must not damage the identity entries — a canonical ref still has to resolve")
	}
}

// ---------------------------------------------------------------------------
// canonicaliseImageryScopeRefs
// ---------------------------------------------------------------------------

func TestCanonicaliseImageryScopeRefs_PageAndSectionRefsRewritten(t *testing.T) {
	pages := []planPageRow{{RawName: "about", Name: "about-index", Sections: sec("hero", "content-block-about", "differentiators")}}
	m := buildCanonicalPageNameMap(pages, zap.NewNop())

	rows := []imageryRow{
		imgRow("page", "about", "hero_about"),
		imgRow("section", "about:2", "icon_no_ads"),
	}
	rows, misses, rewritten := canonicaliseImageryScopeRefs(rows, m, sectionCountByPage(pages), zap.NewNop())

	if got := refOf(t, rows[0]); got != "about-index" {
		t.Errorf("page ref = %q, want %q", got, "about-index")
	}
	// The ordinal must survive byte-identical — it was right all along on the
	// live gamesdesign rows; only the page name had been renamed.
	if got := refOf(t, rows[1]); got != "about-index:2" {
		t.Errorf("section ref = %q, want %q", got, "about-index:2")
	}
	if rewritten != 2 {
		t.Errorf("rewritten = %d, want 2", rewritten)
	}
	if len(misses) != 0 {
		t.Errorf("expected no misses, got %+v", misses)
	}
}

func TestCanonicaliseImageryScopeRefs_ResolvedRefsUntouched(t *testing.T) {
	// The no-regression property: a ref already spelt the way the plan holds it
	// comes through byte-identical, with nothing recorded. This is what makes
	// "rows that work today cannot be broken by this change" a test rather than
	// an assurance.
	pages := []planPageRow{{RawName: "about-index", Name: "about-index", Sections: sec("hero", "body")}}
	m := buildCanonicalPageNameMap(pages, zap.NewNop())

	rows := []imageryRow{
		imgRow("page", "about-index", "hero_about"),
		imgRow("section", "about-index:1", "icon_x"),
	}
	rows, misses, rewritten := canonicaliseImageryScopeRefs(rows, m, sectionCountByPage(pages), zap.NewNop())

	if refOf(t, rows[0]) != "about-index" || refOf(t, rows[1]) != "about-index:1" {
		t.Errorf("already-canonical refs were altered: %q, %q", refOf(t, rows[0]), refOf(t, rows[1]))
	}
	if rewritten != 0 || len(misses) != 0 {
		t.Errorf("rewritten=%d misses=%+v; want 0 and none", rewritten, misses)
	}
}

func TestCanonicaliseImageryScopeRefs_SiteScopeIsNeverTouched(t *testing.T) {
	rows := []imageryRow{{Scope: "site", Key: "logo", Kind: "logo", Prompt: "p"}}
	rows, misses, rewritten := canonicaliseImageryScopeRefs(rows, map[string]string{}, map[string]int{}, zap.NewNop())

	if rows[0].ScopeRef != nil {
		t.Errorf("site scope must keep a NULL scope_ref (chk_scope_ref_consistency), got %q", refOf(t, rows[0]))
	}
	if rewritten != 0 || len(misses) != 0 {
		t.Errorf("site scope produced rewritten=%d misses=%+v", rewritten, misses)
	}
}

func TestCanonicaliseImageryScopeRefs_UnresolvableRefPreservedAndRecorded(t *testing.T) {
	// The planner named a page it never planned. We cannot resolve it without
	// inventing a target, and dropping it would delete imagery somebody asked
	// for — so it is written verbatim and recorded.
	pages := []planPageRow{{RawName: "about", Name: "about-index", Sections: sec("hero")}}
	m := buildCanonicalPageNameMap(pages, zap.NewNop())

	rows := []imageryRow{imgRow("page", "team", "hero_team")}
	rows, misses, rewritten := canonicaliseImageryScopeRefs(rows, m, sectionCountByPage(pages), zap.NewNop())

	if len(rows) != 1 {
		t.Fatalf("an unresolvable ref must NOT be dropped; got %d rows", len(rows))
	}
	if got := refOf(t, rows[0]); got != "team" {
		t.Errorf("unresolvable ref = %q, want it preserved verbatim as %q", got, "team")
	}
	if rewritten != 0 {
		t.Errorf("rewritten = %d, want 0", rewritten)
	}
	if len(misses) != 1 || misses[0].Reason != "page_not_in_plan" {
		t.Fatalf("want exactly one page_not_in_plan miss, got %+v", misses)
	}
	if misses[0].WrittenRef != misses[0].RawRef {
		t.Errorf("for an unresolved ref the written and raw spellings must agree, got %q / %q",
			misses[0].WrittenRef, misses[0].RawRef)
	}
}

func TestCanonicaliseImageryScopeRefs_OrdinalOutOfRangeIsLoggedNotRewritten(t *testing.T) {
	// Pins the deliberate decision (see the file header in
	// write_site_plan_imagery_scope.go): no consumer parses the ordinal, the
	// correct value is unknowable at this seam, and rewriting it risks the
	// unique index. So it is recorded, not repaired. If someone later makes it
	// rewrite, this test forces that to be a decision rather than a drift.
	pages := []planPageRow{{RawName: "about", Name: "about-index", Sections: sec("hero", "body", "cta")}}
	m := buildCanonicalPageNameMap(pages, zap.NewNop())

	rows := []imageryRow{imgRow("section", "about:7", "illustration_x")}
	rows, misses, _ := canonicaliseImageryScopeRefs(rows, m, sectionCountByPage(pages), zap.NewNop())

	if got := refOf(t, rows[0]); got != "about-index:7" {
		t.Errorf("page part must still be resolved and the ordinal left alone; got %q", got)
	}
	if len(misses) != 1 || misses[0].Reason != "ordinal_out_of_range" {
		t.Fatalf("want one ordinal_out_of_range miss, got %+v", misses)
	}
}

func TestCanonicaliseImageryScopeRefs_MalformedOrdinalSurvivesAndIsRecorded(t *testing.T) {
	pages := []planPageRow{{RawName: "about", Name: "about-index", Sections: sec("hero")}}
	m := buildCanonicalPageNameMap(pages, zap.NewNop())

	rows := []imageryRow{imgRow("section", "about:hero", "icon_x")}
	rows, misses, _ := canonicaliseImageryScopeRefs(rows, m, sectionCountByPage(pages), zap.NewNop())

	// Splitting on the FIRST colon and reattaching the remainder verbatim means
	// text we do not understand still round-trips, and the row keeps the colon
	// chk_scope_ref_consistency demands of a section ref.
	if got := refOf(t, rows[0]); got != "about-index:hero" {
		t.Errorf("ref = %q, want %q", got, "about-index:hero")
	}
	if len(misses) != 1 || misses[0].Reason != "ordinal_malformed" {
		t.Fatalf("want one ordinal_malformed miss, got %+v", misses)
	}
}

func TestSplitScopeRef_SplitsOnTheFirstColonOnly(t *testing.T) {
	// Every consumer splits on the first colon. If this helper ever split on
	// the last, a ref like "a:b:1" would be reassembled differently from the
	// way plan_sections' LIKE join reads it.
	for _, tc := range []struct{ in, page, rest string }{
		{"about:2", "about", ":2"},
		{"a:b:1", "a", ":b:1"},
		{"about", "about", ""},
	} {
		p, r := splitScopeRef(tc.in)
		if p != tc.page || r != tc.rest {
			t.Errorf("splitScopeRef(%q) = (%q,%q), want (%q,%q)", tc.in, p, r, tc.page, tc.rest)
		}
	}
}

// ---------------------------------------------------------------------------
// dedupeImageryRows — the guard on idx_site_plan_imagery_unique
// ---------------------------------------------------------------------------

func TestDedupeImageryRows_NoDuplicateIdentitiesSurvive(t *testing.T) {
	// The property the unique index requires. Canonicalisation is a collapsing
	// map, so this input is reachable the moment a planner keys one page under
	// two spellings with a shared imagery key — and without the guard the
	// insert aborts the WHOLE plan write (bugs_open/215 on the sibling table).
	a := imgRow("page", "about-index", "hero_about")
	a.RawScopeRef = "about" // was rewritten
	b := imgRow("page", "about-index", "hero_about")
	b.RawScopeRef = "about-index" // was already aimed correctly

	out, merges := dedupeImageryRows([]imageryRow{a, b}, zap.NewNop())

	if len(out) != 1 || merges != 1 {
		t.Fatalf("got %d rows / %d merges, want 1 and 1 — this is exactly the state the unique index rejects, and it would abort the entire plan write", len(out), merges)
	}
	if out[0].RawScopeRef != "about-index" {
		t.Errorf("survivor raw ref = %q; the entry the planner already aimed at the plan's real name should win", out[0].RawScopeRef)
	}
}

func TestDedupeImageryRows_DistinctRowsUntouched(t *testing.T) {
	rows := []imageryRow{
		imgRow("page", "about-index", "hero_about"),
		imgRow("page", "contact-index", "hero_contact"),
		imgRow("section", "about-index:2", "icon_no_ads"),
		{Scope: "site", Key: "logo", Kind: "logo", Prompt: "p"},
	}
	out, merges := dedupeImageryRows(rows, zap.NewNop())

	if merges != 0 || len(out) != 4 {
		t.Fatalf("distinct identities must all survive; got %d rows / %d merges", len(out), merges)
	}
}

func TestDedupeImageryRows_SameRefDifferentKeysBothSurvive(t *testing.T) {
	// The live gamesdesign shape: four icons on ONE section ref. The unique
	// index includes key, so these are four distinct identities and collapsing
	// them would silently delete three planned images.
	rows := []imageryRow{
		imgRow("section", "about-index:2", "icon_no_ads"),
		imgRow("section", "about-index:2", "icon_math_first"),
		imgRow("section", "about-index:2", "icon_browser_based"),
		imgRow("section", "about-index:2", "icon_practitioner"),
	}
	out, merges := dedupeImageryRows(rows, zap.NewNop())

	if merges != 0 || len(out) != 4 {
		t.Fatalf("four icons on one section ref are four identities; got %d rows / %d merges", len(out), merges)
	}
}

// ---------------------------------------------------------------------------
// The seam
// ---------------------------------------------------------------------------

// TestFlattenImageryBlock_ScopeRefsAreCanonicalisedEndToEnd is the unwired
// detector. It runs the real flatten over a real planner-shaped payload and
// then the real resolution over the real planRows, so it fails if the
// resolution is removed from WriteSitePlanAction's 2c block, if flatten stops
// populating RawScopeRef, or if the map is built from the wrong side.
//
// It uses the live 2026-08-10 gamesdesign shape end to end.
func TestFlattenImageryBlock_ScopeRefsAreCanonicalisedEndToEnd(t *testing.T) {
	collected := map[string]interface{}{
		"imagery": map[string]interface{}{
			"pages": map[string]interface{}{
				"about": []interface{}{
					map[string]interface{}{"key": "hero_about", "kind": "hero", "prompt": "an about hero"},
				},
			},
			"sections": map[string]interface{}{
				"about:2": []interface{}{
					map[string]interface{}{"key": "icon_no_ads", "kind": "icon", "prompt": "no ads"},
				},
			},
		},
	}

	// What WriteSitePlanAction holds by the time it flattens imagery: pages
	// already canonicalised, so "about" is now "about-index".
	planRows := []planPageRow{{
		RawName:  "about",
		Slug:     "about",
		Name:     "about-index",
		Sections: sec("hero", "content-block-about", "differentiators"),
	}}

	rows := flattenImageryBlock(collected, zap.NewNop())
	if len(rows) != 2 {
		t.Fatalf("expected 2 imagery rows from the fixture, got %d", len(rows))
	}

	rows, misses, rewritten := canonicaliseImageryScopeRefs(
		rows,
		buildCanonicalPageNameMap(planRows, zap.NewNop()),
		sectionCountByPage(planRows),
		zap.NewNop(),
	)
	rows, _ = dedupeImageryRows(rows, zap.NewNop())

	got := map[string]string{}
	for _, r := range rows {
		got[r.Key] = refOf(t, r)
	}
	if got["hero_about"] != "about-index" {
		t.Errorf("page imagery scope_ref = %q, want %q\n"+
			"if this reads \"about\", the canonicalisation is not running — which is the bug this file exists for",
			got["hero_about"], "about-index")
	}
	if got["icon_no_ads"] != "about-index:2" {
		t.Errorf("section imagery scope_ref = %q, want %q", got["icon_no_ads"], "about-index:2")
	}
	if rewritten != 2 {
		t.Errorf("rewritten = %d, want 2", rewritten)
	}
	if len(misses) != 0 {
		t.Errorf("the live gamesdesign shape has a valid ordinal (about-index has 3 sections) — unexpected misses %+v", misses)
	}
}

// TestFlattenImageryBlock_KeysAreWalkedInAStableOrder pins the determinism the
// dedupe guard's first-appearance rule depends on. Go randomises map iteration,
// so without sorting the survivor of a collision would vary between runs.
func TestFlattenImageryBlock_KeysAreWalkedInAStableOrder(t *testing.T) {
	entry := []interface{}{map[string]interface{}{"key": "k", "kind": "hero", "prompt": "p"}}
	collected := map[string]interface{}{
		"imagery": map[string]interface{}{
			"pages": map[string]interface{}{
				"zeta": entry, "alpha": entry, "mid": entry,
			},
		},
	}

	var first []string
	for i := 0; i < 20; i++ {
		var order []string
		for _, r := range flattenImageryBlock(collected, zap.NewNop()) {
			order = append(order, refOf(t, r))
		}
		if i == 0 {
			first = order
			continue
		}
		for j := range order {
			if order[j] != first[j] {
				t.Fatalf("iteration order varied between runs: %v vs %v", first, order)
			}
		}
	}
	if len(first) != 3 || first[0] != "alpha" || first[2] != "zeta" {
		t.Errorf("expected sorted page order, got %v", first)
	}
}
