package actions

import (
	"testing"

	"go.uber.org/zap"
)

// Pass C's URL-form contract (bugs_open/463).
//
// Pass C exists to drop a page that COLLIDES with a realised section index — a
// flat /articles.html re-proposed beside a live /articles/index.html. It used to
// compare the FIRST PATH SEGMENT of each side, under which a legitimate CHILD of
// that index (/articles/x.html) is the same string as the collider, so every
// newly planned child was deleted. Nothing recorded it: the count was logged and
// the orchestration completed green.
//
// These tests are written in PAIRS on purpose. A suite of KEEP cases alone would
// be vacuous — it would pass against a Pass C deleted outright — so every hub
// shape below asserts a DROP that must survive as well as a KEEP that must
// start working. The DROP cases are bugs_closed/141's ratified behaviour.
//
// The fixtures are shaped from the real gamedesign.uk regression of 2026-09-03
// (orchestration 9fe9660e): a deployed articles-index at /articles/index.html
// and five new blog-post children under /articles/, of which nine planned pages
// became four.

// sectionHub builds a realised section-index row in the shape the
// load_existing_pages query returns. build_status "deployed" is what puts it in
// the preservation set, without which reconcilePlanWithRealised returns early
// and no pass runs at all.
func sectionHub(name, url, pageType string) map[string]interface{} {
	return map[string]interface{}{
		"name": name, "url": url, "page_type": pageType,
		"title": name, "build_status": "deployed",
		"adoption_locked": false,
		"sections":        `["hero","content-listing"]`,
	}
}

func proposedPage(name, url, pageType string) map[string]interface{} {
	return map[string]interface{}{
		"name": name, "url": url, "page_type": pageType,
		"title": name, "sections": []interface{}{"hero", "rich-text"},
	}
}

func TestPassC_ChildOfSectionIndexSurvives_ColliderStillDropped(t *testing.T) {
	cases := []struct {
		name                        string
		hubName, hubURL, hubType    string
		planName, planURL, planType string
		wantKept                    bool
		why                         string
	}{
		// ── /articles/ hub at the canonical URL — the filed case ───────────
		{
			name:    "flat page claiming the hub's own path is dropped",
			hubName: "articles-index", hubURL: "/articles/index.html", hubType: "section-index",
			planName: "articles", planURL: "/articles.html", planType: "content",
			wantKept: false,
			why:      "this is the collision Pass C exists for; /articles.html and /articles/index.html both claim /articles",
		},
		{
			name:    "child at the flat child URL survives",
			hubName: "articles-index", hubURL: "/articles/index.html", hubType: "section-index",
			planName: "article-sign-off-problem", planURL: "/articles/the-sign-off-problem.html", planType: "blog-post",
			wantKept: true,
			why:      "bugs_open/463: the child claims /articles/the-sign-off-problem, a LONGER path than the hub's",
		},
		{
			name:    "child at the directory child URL survives",
			hubName: "guides-index", hubURL: "/guides/index.html", hubType: "section-index",
			planName: "guide-buy-to-let", planURL: "/guides/buy-to-let/index.html", planType: "guide",
			wantKept: true,
			why:      "slugOf stripped the trailing /index BEFORE taking the first segment, so this form collapsed to the hub too",
		},
		// ── flat hub (/news.html) — bugs_closed/141's shape ────────────────
		{
			name:    "flat collider beside a FLAT news hub is still dropped",
			hubName: "news", hubURL: "/news.html", hubType: "news-index",
			planName: "news-updates", planURL: "/news.html", planType: "content",
			wantKept: false,
			why:      "bugs_closed/141 ratified exactly this drop; both sides claim /news",
		},
		{
			name:    "child under a FLAT news hub survives",
			hubName: "news", hubURL: "/news.html", hubType: "news-index",
			planName: "news-item", planURL: "/news/season-opener.html", planType: "blog-post",
			wantKept: true,
			why:      "same defect as the canonical hub, in the flat-hub shape",
		},

		// ── nested hub — the old rule ate a sibling section entirely ───────
		{
			name:    "collider beside a NESTED hub is dropped",
			hubName: "guides-index", hubURL: "/resources/guides/index.html", hubType: "section-index",
			planName: "resources-guides", planURL: "/resources/guides.html", planType: "content",
			wantKept: false,
			why:      "claims /resources/guides, the hub's own path",
		},
		{
			name:    "an unrelated page in the PARENT directory of a nested hub survives",
			hubName: "guides-index", hubURL: "/resources/guides/index.html", hubType: "section-index",
			planName: "resources-downloads", planURL: "/resources/downloads.html", planType: "content",
			wantKept: true,
			why:      "the old first-segment rule registered the stem 'resources' and dropped every page under it, including siblings of the hub",
		},

		// ── the URL-less fallback: unchanged, name-only ────────────────────
		{
			name:    "a page with NO url colliding by name is still dropped",
			hubName: "articles-index", hubURL: "/articles/index.html", hubType: "section-index",
			planName: "articles", planURL: "", planType: "content",
			wantKept: false,
			why:      "no URL to compare, so the name-stem comparison still governs — deliberately unchanged",
		},
		{
			name:    "a page with no url and an unrelated name survives",
			hubName: "articles-index", hubURL: "/articles/index.html", hubType: "section-index",
			planName: "about", planURL: "", planType: "content",
			wantKept: true,
			why:      "control for the fallback arm",
		},

		// ── URL shapes PagePathKey passes through but Pass C used to accept ─
		{
			name:    "a collider url with no leading slash is still dropped",
			hubName: "articles-index", hubURL: "/articles/index.html", hubType: "section-index",
			planName: "articles-flat", planURL: "articles.html", planType: "content",
			wantKept: false,
			why:      "sectionPathKey normalises the missing leading slash; the old first-segment rule accepted this form too",
		},
		{
			name:    "a child url with no leading slash survives",
			hubName: "articles-index", hubURL: "/articles/index.html", hubType: "section-index",
			planName: "article-two", planURL: "articles/two.html", planType: "blog-post",
			wantKept: true,
			why:      "same normalisation, child side",
		},
		{
			name:    "a collider against a hub url with a TRAILING SLASH is dropped",
			hubName: "insights-index", hubURL: "/insights/", hubType: "section-index",
			planName: "insights", planURL: "/insights.html", planType: "content",
			wantKept: false,
			why:      "PagePathKey leaves /insights/ alone, so sectionPathKey trims it; without that this drop would silently stop working",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			existing := []interface{}{sectionHub(tc.hubName, tc.hubURL, tc.hubType)}
			llm := []interface{}{proposedPage(tc.planName, tc.planURL, tc.planType)}

			got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

			kept := hasPage(got, tc.planName)
			if kept != tc.wantKept {
				verb := "was DROPPED"
				if kept {
					verb = "was KEPT"
				}
				t.Errorf("page %q (%s) beside hub %q (%s) %s; want kept=%v\n  why: %s",
					tc.planName, tc.planURL, tc.hubName, tc.hubURL, verb, tc.wantKept, tc.why)
			}
			// The drop must also be RECORDED, not merely counted — a bare count
			// is what made bugs_open/463 invisible for three months.
			if !tc.wantKept {
				if len(counts.DroppedPages) != 1 {
					t.Fatalf("dropped_pages = %d, want 1 (the drop must name the page)", len(counts.DroppedPages))
				}
				d := counts.DroppedPages[0]
				if d.Name != tc.planName || d.Pass != "C" {
					t.Errorf("dropped record = %+v, want name=%q pass=C", d, tc.planName)
				}
			} else if len(counts.DroppedPages) != 0 {
				t.Errorf("dropped_pages = %+v, want none", counts.DroppedPages)
			}
		})
	}
}

// A re-proposed section index at the hub's OWN url is exempt from Pass C by
// type — and then Pass B snaps it onto the realised identity, so it survives
// under the STORED name rather than the planner's.
//
// Recorded as its own test because I got this wrong first: I expected the page
// to survive under its proposed name, and the table above failed. It is not a
// Pass C drop at all, and reading DroppedCollision is what distinguishes the two
// — a page absent under the name you looked for is not the same as a page
// deleted.
func TestPassC_SectionIndexAtTheHubURLIsSnappedNotDropped(t *testing.T) {
	existing := []interface{}{sectionHub("articles-index", "/articles/index.html", "section-index")}
	llm := []interface{}{proposedPage("articles-hub", "/articles/index.html", "section-index")}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	if counts.DroppedCollision != 0 {
		t.Errorf("dropped_collision = %d, want 0 — a section-index type is exempt from Pass C", counts.DroppedCollision)
	}
	if counts.SnappedRename != 1 {
		t.Errorf("snapped_rename = %d, want 1 (Pass B owns this case)", counts.SnappedRename)
	}
	if !hasPage(got, "articles-index") {
		t.Error("the hub is absent from the plan entirely")
	}
}

// The filed regression end to end: the whole plan survives validation.
//
// This is the assertion bugs_open/463 §7 names as the pass condition —
// proposed == survived at the plan/validate step boundary — expressed at the
// function that decides it. Before the fix this returned 4 of 9.
func TestPassC_GamedesignArticlesPlanSurvivesIntact(t *testing.T) {
	existing := []interface{}{
		sectionHub("articles-index", "/articles/index.html", "section-index"),
		sectionHub("index", "/index.html", "landing"),
	}
	llm := []interface{}{
		proposedPage("index", "/index.html", "landing"),
		proposedPage("about", "/about.html", "content"),
		proposedPage("contact", "/contact.html", "content"),
		proposedPage("articles-index", "/articles/index.html", "section-index"),
		proposedPage("article-sign-off-problem", "/articles/the-sign-off-problem.html", "blog-post"),
		proposedPage("article-gdd-abandonment", "/articles/gdd-abandonment.html", "blog-post"),
		proposedPage("article-design-engineering-handoff", "/articles/design-engineering-handoff.html", "blog-post"),
		proposedPage("article-narrative-design-pipeline", "/articles/narrative-design-pipeline.html", "blog-post"),
		proposedPage("article-principal-transition", "/articles/principal-transition.html", "blog-post"),
	}

	got, counts := reconcilePlanWithRealised(llm, existing, reconcileOptions{}, zap.NewNop())

	if counts.DroppedCollision != 0 {
		t.Errorf("dropped_collision = %d, want 0; dropped: %+v", counts.DroppedCollision, counts.DroppedPages)
	}
	for _, want := range []string{
		"article-sign-off-problem", "article-gdd-abandonment",
		"article-design-engineering-handoff", "article-narrative-design-pipeline",
		"article-principal-transition",
	} {
		if !hasPage(got, want) {
			t.Errorf("child page %q was dropped; this is the filed regression", want)
		}
	}
	if len(got) < len(llm) {
		t.Errorf("plan shrank from %d to %d pages", len(llm), len(got))
	}
}

// sectionPathKey's own contract, including the two shapes PagePathKey
// deliberately passes through unchanged and this wrapper must normalise.
func TestSectionPathKey(t *testing.T) {
	cases := []struct{ url, want string }{
		{"/articles/index.html", "/articles"},
		{"/articles.html", "/articles"},
		{"/articles/x.html", "/articles/x"},
		{"/articles/x/index.html", "/articles/x"},
		{"/articles/", "/articles"},
		{"articles.html", "/articles"},
		{"articles/x.html", "/articles/x"},
		{"/resources/guides/index.html", "/resources/guides"},
		{"/index.html", "/"},
		{"/", "/"},
		{"", ""},
	}
	for _, c := range cases {
		if got := sectionPathKey(c.url); got != c.want {
			t.Errorf("sectionPathKey(%q) = %q, want %q", c.url, got, c.want)
		}
	}
}

// A child page and its hub must never produce the same key — that equality IS
// the bug. Stated as its own assertion so it cannot be lost in a table edit.
func TestSectionPathKey_ChildNeverEqualsItsHub(t *testing.T) {
	hub := sectionPathKey("/articles/index.html")
	for _, child := range []string{
		"/articles/one.html", "/articles/two/index.html", "/articles/deep/nested/three.html",
	} {
		if sectionPathKey(child) == hub {
			t.Errorf("child %q collapses onto its hub key %q — Pass C would delete it", child, hub)
		}
	}
	// ...and the collider must, or Pass C stops doing its job (bugs_closed/141).
	if sectionPathKey("/articles.html") != hub {
		t.Errorf("collider /articles.html no longer claims the hub path %q", hub)
	}
}
