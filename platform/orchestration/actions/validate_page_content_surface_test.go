// FILE: platform/orchestration/actions/validate_page_content_surface_test.go
//
// The build gate's half of bugs_open/102: it must actually FIND the page type,
// or the fix is inert on the only path that runs today (the post-deploy audit
// is bugs_open/083).
//
// resolvePageType reads collected data, and collected data is shaped by whatever
// workflow calls the action — so the resolution order is the load-bearing part,
// not the policy. page-build-handler runs load_page_record (which selects
// page_type) into output_field "page_record" before validate_content; the other
// paths are the same fallbacks load_page_record itself walks.

package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func TestResolvePageTypeFindsTheLivePath(t *testing.T) {
	logger := zap.NewNop()

	cases := []struct {
		name      string
		config    map[string]interface{}
		collected map[string]interface{}
		want      string
	}{
		{
			// THE LIVE SHAPE: page-build-handler's load_page_record output.
			name:   "page_record from load_page_record",
			config: map[string]interface{}{},
			collected: map[string]interface{}{
				"page_record": map[string]interface{}{
					"found": true, "name": "learn-operations-scaling", "page_type": "guide",
				},
			},
			want: "guide",
		},
		{
			name:   "current_page, which page-build-handler maps from page_record",
			config: map[string]interface{}{},
			collected: map[string]interface{}{
				"current_page": map[string]interface{}{"page_type": "blog-post"},
			},
			want: "blog-post",
		},
		{
			name:      "input_data.spec, the work-item shape",
			config:    map[string]interface{}{},
			collected: map[string]interface{}{"input_data": map[string]interface{}{"spec": map[string]interface{}{"page_type": "tool"}}},
			want:      "tool",
		},
		{
			name:      "a config dot-path wins over the fallbacks",
			config:    map[string]interface{}{"page_type": "site_page.kind"},
			collected: map[string]interface{}{"site_page": map[string]interface{}{"kind": "landing"}, "page_record": map[string]interface{}{"page_type": "guide"}},
			want:      "landing",
		},
		{
			name:      "a config literal is taken as written",
			config:    map[string]interface{}{"page_type": "guide"},
			collected: map[string]interface{}{},
			want:      "guide",
		},
		{
			// Site chrome, a component reviewer, a page that does not exist yet.
			name:      "nothing to find is UNKNOWN, not an error",
			config:    map[string]interface{}{},
			collected: map[string]interface{}{"page_content": map[string]interface{}{"response": map[string]interface{}{}}},
			want:      "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolvePageType(tc.config, tc.collected, logger); got != tc.want {
				t.Errorf("resolvePageType = %q, want %q", got, tc.want)
			}
		})
	}
}

// The gate's severity contract is unchanged on a business surface and the
// number check goes silent on an editorial one — while the banned claim, which
// is what BLOCKS a deploy, is untouched on both.
func TestClaimsGateHonoursTheSurface(t *testing.T) {
	eb, err := datahelpers.ParseEvidenceBase([]byte(claimsGateTestEB))
	if err != nil || eb == nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}
	blocks := datahelpers.ExtractAssertionText(
		`<p>We span eight departments and serve 45 clients.</p>`)

	guide := datahelpers.ClaimSurface{PageType: "guide"}
	if n := checkUnregisteredNumbers(blocks, eb, guide); len(n) != 0 {
		t.Errorf("a guide must raise no unregistered-number errors, got %+v", n)
	}
	if b := checkBannedClaims(blocks, eb, true, "test-site", zap.NewNop()); len(b) != 1 || b[0].Severity != "blocker" {
		t.Errorf("the banned claim must still block on a guide, got %+v", b)
	}

	content := datahelpers.ClaimSurface{PageType: "content"}
	n := checkUnregisteredNumbers(blocks, eb, content)
	if len(n) != 1 || n[0].Severity != "error" || n[0].Value != "45" {
		t.Errorf("a business page must still raise the unregistered number as an error, got %+v", n)
	}
}

// ============================================================================
// COMPONENT GRAIN (RFC_053 / bugs_open/364 Phase 2)
// ============================================================================

func sectionsMetadataFixture(sections []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"page_content": map[string]interface{}{
			"response": map[string]interface{}{
				"sections_metadata": sections,
			},
		},
	}
}

// The happy path: every section resolves, so the gate may judge per component.
func TestCollectPageSectionsResolvesEverySection(t *testing.T) {
	collected := sectionsMetadataFixture([]interface{}{
		map[string]interface{}{"stored_slot_name": "hero", "rendered_html": "<section><p>We hold 45,000 client records.</p></section>"},
		map[string]interface{}{"stored_slot_name": "adoption-tracker-listing", "rendered_html": "<section><p>Over 80% of Fortune 500.</p></section>"},
	})
	got, ok := collectPageSections(collected, defaultSectionsMetadataField, zap.NewNop())
	if !ok {
		t.Fatal("expected component grain to be available when every section resolves")
	}
	if len(got) != 2 {
		t.Fatalf("want 2 sections, got %d", len(got))
	}
	if got[0].ComponentFunction != "hero" || got[1].ComponentFunction != "adoption-tracker-listing" {
		t.Errorf("slot identity not carried: %+v", got)
	}
}

// THE ONE THAT MATTERS — the council's bug_historian objection, round 1
// (correlation 3ed2b792, medium).
//
// A section whose HTML the canonical reader cannot resolve must NOT be silently
// dropped from the scan while the rest of the page is judged and reported as
// scanned. That is a partial, invisible loss of coverage on the only gate that
// refuses a page. The guard is a COUNT: fewer sections back than the metadata
// held means something was dropped, whatever the reason, and the whole page
// falls back to the whole-page scan — which still reads those bytes.
//
// Mutation check for whoever changes this: delete the len(sections) != len(items)
// guard and this test must fail. If it still passes, the guard is not what is
// producing the refusal and the coverage hole is back.
func TestAnUnreadableSectionRefusesComponentGrainForTheWholePage(t *testing.T) {
	for _, tc := range []struct {
		name    string
		section interface{}
	}{
		{"html under a key the reader does not know", map[string]interface{}{
			"stored_slot_name": "mystery", "html_output": "<section><p>We hold 45,000 client records.</p></section>"}},
		{"not a map at all", "a bare string the producer should never have emitted"},
		{"empty html", map[string]interface{}{"stored_slot_name": "empty", "rendered_html": ""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			collected := sectionsMetadataFixture([]interface{}{
				map[string]interface{}{"stored_slot_name": "hero", "rendered_html": "<section><p>ok</p></section>"},
				tc.section,
			})
			if _, ok := collectPageSections(collected, defaultSectionsMetadataField, zap.NewNop()); ok {
				t.Error("component grain was claimed while a section could not be resolved — " +
					"that page would be scanned in part and reported as scanned in full")
			}
		})
	}
}

// Absent metadata is the common case (3 of the gate's 4 callers) and must read as
// "cannot answer at component grain", never as "there are no sections".
func TestAbsentSectionsMetadataFallsBackRatherThanScanningNothing(t *testing.T) {
	for name, collected := range map[string]map[string]interface{}{
		"no field at all": {},
		"not a list":      sectionsMetadataFixture(nil),
		"empty list":      sectionsMetadataFixture([]interface{}{}),
	} {
		if _, ok := collectPageSections(collected, defaultSectionsMetadataField, zap.NewNop()); ok {
			t.Errorf("%s: claimed component grain with nothing to judge", name)
		}
	}
}
