package discovery_checks

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// The SQL in check_missing_prose_links.go is deliberately untested here — it is
// instrumented at runtime by the in-run canary and the per-type census instead.
// What IS tested is every rule that decides whether a page is convicted, and the
// instrument that feeds it.
//
// EACH RULE HAS A MUTATION THAT KILLS IT, and every one below was RUN, not
// asserted (2026-08-25). A rule test that passes against its own deletion is
// decoration, so each mutation was applied to a copy, compiled, tested, and
// reverted by restoring the original text — never by `git restore`, which on an
// uncommitted file reverts to HEAD and takes the whole change with it.
//
//	M1 drop the proseLinkPageTypes gate      -> TestOutOfRemitPageTypesAreNotConvicted + canary
//	M2 drop the `internalPageLinks > 0` arm  -> TestAPageWithLinksIsClean + canary
//	M3 drop the minProseCharsForLinks arm    -> TestAStubIsTooThinToJudge + canary
//	M4 drop the minSiteTargetsForLinks arm   -> TestASiteWithNothingToLinkToIsClean + canary
//	M5a count LinkScopeExternal not ...Page  -> TestProseAnchorsAreCounted, OnlyInternal, Nested, canary
//	M5b count every href, ignoring scope     -> TestOnlyInternalPageLinksCount
//	M6 stop the walker descending containers -> TestNestedProseAnchorsAreFound
//	M7 return 0 rather than the -1 sentinel  -> TestUnparseableContentDataIsNotZero
//
// ⚠ The FIRST attempt at M5 replaced the whole scope comparison with `false`,
// which left `href` unused and DID NOT COMPILE. `go test` exited non-zero and it
// read as "mutation killed" — it was not; the compiler stopped it before a test
// could. A mutation must compile or it measures nothing, so each run now builds
// first and reports a non-compiling mutation as invalid rather than as a pass.

func TestALinklessProsePageIsConvicted(t *testing.T) {
	for _, pt := range []string{"blog-post", "guide", "content"} {
		if !pageMissingProseLinks(pt, 0, 5000, 10) {
			t.Errorf("page_type %q with no links, ample prose and a linkable site should be convicted", pt)
		}
	}
}

// M2. This is the positive control for the whole check: if it stops failing when
// the zero-link arm is deleted, the check convicts healthy pages.
func TestAPageWithLinksIsClean(t *testing.T) {
	if pageMissingProseLinks("guide", 1, 5000, 10) {
		t.Fatal("a page carrying one internal prose link must not be convicted")
	}
}

// M1. Tool pages are 159/200 link-less by design; convicting them would be the
// wrong instrument, and this is the assertion that says so.
func TestOutOfRemitPageTypesAreNotConvicted(t *testing.T) {
	for _, pt := range []string{"tool", "landing", "entity-page", "section-index", "news-index", ""} {
		if pageMissingProseLinks(pt, 0, 5000, 10) {
			t.Errorf("page_type %q is out of remit and must not be convicted", pt)
		}
	}
}

// M3.
func TestAStubIsTooThinToJudge(t *testing.T) {
	if pageMissingProseLinks("content", 0, minProseCharsForLinks-1, 10) {
		t.Fatal("a page under the prose floor must not be convicted")
	}
}

// M4.
func TestASiteWithNothingToLinkToIsClean(t *testing.T) {
	if pageMissingProseLinks("content", 0, 5000, minSiteTargetsForLinks-1) {
		t.Fatal("a page on a site with too few link targets must not be convicted")
	}
}

// M5. The instrument, both polarities.
func TestProseAnchorsAreCounted(t *testing.T) {
	cd := `{"content":"<p>See our <a href=\"/about.html\">about page</a> and <a href=\"/pricing.html\">pricing</a>.</p>"}`
	if n := countProseLinks(cd); n != 2 {
		t.Fatalf("expected 2 prose anchors, got %d", n)
	}
	if n := countProseLinks(`{"content":"<p>Plain prose, no anchors at all.</p>"}`); n != 0 {
		t.Fatalf("expected 0 prose anchors, got %d", n)
	}
}

// The stated scope of the instrument, asserted rather than assumed. Hero and CTA
// links live in content_data as STRUCTURED fields and are chosen by a different
// mechanism; counting them would make a link-less page read as healthy. A later
// maintainer "fixing" this exclusion breaks the check, so it is a test.
func TestStructuredLinkFieldsAreNotProseAnchors(t *testing.T) {
	cd := `{"cta_url":"/contact.html","cta_text":"Talk to us","hero_url":"/assets/images/h.jpg","link_url":"/x.html"}`
	if n := countProseLinks(cd); n != 0 {
		t.Fatalf("structured link fields must not count as prose anchors, got %d", n)
	}
}

// External, mailto, asset and pure-fragment hrefs are not internal page links.
// This is ClassifyLinkScope's contract; the test pins that we asked it, rather
// than counting every href we find.
func TestOnlyInternalPageLinksCount(t *testing.T) {
	cd := `{"content":"<p>` +
		`<a href=\"https://example.com/x\">ext</a>` +
		`<a href=\"mailto:a@b.c\">mail</a>` +
		`<a href=\"/assets/images/x.png\">asset</a>` +
		`<a href=\"#section\">frag</a>` +
		`</p>"}`
	if n := countProseLinks(cd); n != 0 {
		t.Fatalf("only internal PAGE links should count, got %d", n)
	}
}

// Unparseable content_data must be "cannot tell", never "no links" — otherwise a
// JSON defect silently manufactures a repair item for a page that may be fine.
func TestUnparseableContentDataIsNotZero(t *testing.T) {
	if n := countProseLinks(`{"content": broken`); n != -1 {
		t.Fatalf("unparseable content_data must return the -1 cannot-tell sentinel, got %d", n)
	}
	if n := countProseLinks(""); n != 0 {
		t.Fatalf("empty content_data is genuinely zero links, got %d", n)
	}
}

// Links nested inside arrays and sub-objects must be found — component content
// is rarely flat (cards[], sections[]), and a walker that only reads top-level
// values would under-count and convict healthy pages.
func TestNestedProseAnchorsAreFound(t *testing.T) {
	cd := `{"cards":[{"body":"<p><a href=\"/a.html\">a</a></p>"},{"body":"no link"}],` +
		`"nested":{"deep":{"body":"<p><a href=\"/b.html\">b</a></p>"}}}`
	if n := countProseLinks(cd); n != 2 {
		t.Fatalf("expected 2 nested prose anchors, got %d", n)
	}
}

// The canary is what makes a drifted predicate refuse instead of reporting a
// clean site, so it must pass today and must be sensitive to the rules above.
func TestCanaryPassesOnTheShippedPredicate(t *testing.T) {
	if err := canaryMissingProseLinks(); err != nil {
		t.Fatalf("in-run canary must pass against the shipped predicate: %v", err)
	}
}

// The key is the dedup slot. idx_swi_dedup is UNIQUE on (site_id, item_key) with
// no item_type column, so a prefix collision would let this check and another
// detector silently absorb each other's findings on the same page.
func TestItemKeyIsDistinctAndSelfDescribing(t *testing.T) {
	site := uuid.MustParse("0a538b4a-803c-4f82-b298-d916f893fe8e")
	got := missingProseLinksItemKey("seaweed-and-the-carbon-question", site)

	if !strings.HasPrefix(got, "no_outbound_links:") {
		t.Fatalf("item key must carry the agreed prefix, got %q", got)
	}
	for _, foreign := range []string{"internal_link:", "needs_links:", "tool_crosslink:", "page_rerender"} {
		if strings.HasPrefix(got, foreign) {
			t.Fatalf("item key %q collides with another producer's namespace %q", got, foreign)
		}
	}
	if !strings.Contains(got, site.String()) {
		t.Fatalf("item key must be site-scoped, got %q", got)
	}
	if missingProseLinksItemKey("a", site) == missingProseLinksItemKey("b", site) {
		t.Fatal("item key must distinguish pages")
	}
}

// Both axes, per bugs_open/356. The build arm alone enumerates an ARCHIVED page
// that shipped before it was retired, and this check's remedy REWRITES the page
// — so acting on one would republish content the platform withdrew. A source
// assertion is weak evidence, but the failure it guards is silent and expensive.
func TestPageQueryTakesBothLifecycleAndBuildArms(t *testing.T) {
	if !strings.Contains(missingProseLinksPagesSQL, "status = 'active'") {
		t.Error("page query must carry the LIFECYCLE arm (PageWantedLivePredicateFor)")
	}
	if !strings.Contains(missingProseLinksPagesSQL, "deployed_at IS NULL") {
		t.Error("page query must carry the BUILD arm (PageHasShippedPredicateFor)")
	}
}

// The registry checker requires a `consumed` code's reader FILE to contain the
// code literal. That is satisfied by the const, and this pins it so a later
// refactor to an import does not silently break the DBG-075 join.
func TestFindingCodeLiteralIsDeclaredInThisPackage(t *testing.T) {
	if linkContextUnavailableCode != "LINK_CONTEXT_UNAVAILABLE" {
		t.Fatalf("the finding code literal must not drift: got %q", linkContextUnavailableCode)
	}
}
