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
	if b := checkBannedClaims(blocks, eb, true); len(b) != 1 || b[0].Severity != "blocker" {
		t.Errorf("the banned claim must still block on a guide, got %+v", b)
	}

	content := datahelpers.ClaimSurface{PageType: "content"}
	n := checkUnregisteredNumbers(blocks, eb, content)
	if len(n) != 1 || n[0].Severity != "error" || n[0].Value != "45" {
		t.Errorf("a business page must still raise the unregistered number as an error, got %+v", n)
	}
}
