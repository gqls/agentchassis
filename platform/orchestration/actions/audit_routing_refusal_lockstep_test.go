// FILE: platform/orchestration/actions/audit_routing_refusal_lockstep_test.go
//
// bugs_open/323: the router (write_audit_findings_action.go, classifyFinding)
// must never hand component-template-fixer a fix_type that handler refuses by
// design (fix_component_template_action.go, fixTypesRefusedByDesign).
//
// Why this is a build-time test and not a comment. From 2026-03-14 to
// 2026-08-19 classifyFinding routed every `cta` / `nav_restructure` audit
// finding at component-template-fixer as cta_improvement / nav_restructure, and
// the handler's dispatch switch returned {fixed:false, action:"needs_review"}
// for both on every run. Nothing read the flag, so 993 items across 22 sites
// closed 'complete' with 0 ever fixed. Two files, each internally consistent,
// disagreeing about what the other does — the same lockstep class as
// TestNoChangeRosterMatchesLiveRouting (routing ↔ roster) and
// handler_coverage_test.go (routing ↔ handler existence), one rung further:
// routing ↔ handler CAPABILITY.
//
// The test drives the category universe through classifyFinding exactly as
// TestClassifyFindingEmitsOnlyDeclaredItemTypes does, and for every result that
// names component-template-fixer resolves the fix_type the HANDLER would see —
// spec.fix_type first, then the handler's own category/item_type fallbacks, so
// a route that merely omits spec.fix_type cannot hide a refused type behind the
// ladder in FixComponentTemplateAction.
//
// Mutation-proven on 2026-08-19: re-adding "cta": "cta_improvement" to
// categoryToFixType plus the old Rule 3 fails TestAuditRoutingNeverTargetsAFixerRefusalArm.

package actions

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

// fixTypeTheHandlerWouldResolve mirrors FixComponentTemplateAction's resolution
// ladder (config literal → input_data.fix_type → input_data.spec.fix_type →
// spec.category → item_type) for a finding the dispatch loop would hand it.
// Kept beside the test so a change to the ladder is a change here too.
func fixTypeTheHandlerWouldResolve(c classifiedFinding) string {
	if ft, _ := c.Spec["fix_type"].(string); ft != "" {
		return ft
	}
	if cat, _ := c.Spec["category"].(string); cat != "" {
		if ft := inferFixTypeFromCategory(cat); ft != "" {
			return ft
		}
	}
	return inferFixTypeFromItemType(c.ItemType)
}

func TestAuditRoutingNeverTargetsAFixerRefusalArm(t *testing.T) {
	// A quiet test passes when the rule is gone: the refusal set must be the one
	// this test exists for, or the assertion below is vacuous.
	if len(fixTypesRefusedByDesign) == 0 {
		t.Fatal("fixTypesRefusedByDesign is empty — the refusal set this test guards has been removed; " +
			"either the fixer now does LLM-driven work (then delete this test on purpose) or the map was lost")
	}
	for _, must := range []string{"cta_improvement", "nav_restructure"} {
		if _, ok := fixTypesRefusedByDesign[must]; !ok {
			t.Errorf("fixTypesRefusedByDesign no longer lists %q — if the fixer really accepts it now, "+
				"route the category at it deliberately and update this test; do not drop the entry quietly", must)
		}
	}

	// The static half: the router's own category → fix_type table.
	for cat, ft := range categoryToFixType {
		if reason, refused := fixTypesRefusedByDesign[ft]; refused {
			t.Errorf("categoryToFixType[%q] = %q, which component-template-fixer refuses by design (%s) — "+
				"every item filed this way closes without work (bugs_open/323)", cat, ft, reason)
		}
	}

	// The dynamic half: every route that actually names the handler.
	siteID := uuid.New()
	pages := map[string]pageInfo{"index": {ID: uuid.New(), Name: "index"}}
	checked := 0
	for _, cat := range classifyCategoryUniverse() {
		for _, page := range []string{"index", "pricing", "site-wide", ""} {
			c := classifyFinding(auditFinding{
				Category: cat, Page: page, Severity: "medium", Description: "x",
			}, pages, siteID, "test-audit")
			if c.HandlerAgent != "component-template-fixer" {
				continue
			}
			checked++
			ft := fixTypeTheHandlerWouldResolve(c)
			if ft == "" {
				t.Errorf("category %q page %q routes to component-template-fixer with NO resolvable fix_type "+
					"(spec.fix_type empty, no category/item_type fallback) — the handler will refuse it as "+
					"'fix_type is required'", cat, page)
				continue
			}
			if reason, refused := fixTypesRefusedByDesign[ft]; refused {
				t.Errorf("category %q page %q routes item_type %q to component-template-fixer with fix_type %q, "+
					"which that handler refuses by design (%s). Route it at a handler that can do the work, or "+
					"file it as capability_gap (bugs_closed/077) — never at a refusal arm (bugs_open/323)",
					cat, page, c.ItemType, ft, reason)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no category routed to component-template-fixer at all — the dynamic half checked nothing; " +
			"either designRouting no longer names the fixer (update this test) or classifyCategoryUniverse is empty")
	}
}

// TestNoHandlerCategoriesFileACapabilityGap pins Rule 3's new shape: a `cta` or
// `nav_restructure` finding is a roadmap row, not a dispatch — and the finding's
// own detail survives on it for whoever builds the handler.
func TestNoHandlerCategoriesFileACapabilityGap(t *testing.T) {
	siteID := uuid.New()
	pages := map[string]pageInfo{"index": {ID: uuid.New(), Name: "index"}}

	for _, cat := range []string{"cta", "nav_restructure"} {
		// Both an existing page and a placeholder: the old Rule 3 keyed on page
		// existence for page_id only, and the new arm must not depend on it.
		for _, page := range []string{"index", "site-wide"} {
			c := classifyFinding(auditFinding{
				Category:       cat,
				Page:           page,
				Severity:       "high",
				Description:    "both CTA buttons link to the same URL",
				Suggestion:     "point the secondary button at the catalogue",
				AcceptanceTest: "the two CTA hrefs differ",
			}, pages, siteID, "site-review")

			if c.ItemType != "capability_gap" {
				t.Fatalf("%s/%s: item_type = %q, want capability_gap", cat, page, c.ItemType)
			}
			if c.HandlerAgent != "" || c.Status != "deferred" {
				t.Errorf("%s/%s: handler=%q status=%q — want empty handler + deferred, the undispatchable "+
					"shape (bugs_open/077); anything else dispatches work no handler can do", cat, page, c.HandlerAgent, c.Status)
			}
			if c.Priority != 200 || c.Severity != "low" {
				t.Errorf("%s/%s: priority/severity = %d/%q, want 200/low (CapabilityGapItem conventions)", cat, page, c.Priority, c.Severity)
			}
			if c.Spec["gap_kind"] != checks.GapHandlerMissing {
				t.Errorf("%s/%s: gap_kind = %v, want %q — the router HAS a rule; the estate lacks a HANDLER "+
					"(contrast the unknown-category fallback, which is rule_missing)", cat, page, c.Spec["gap_kind"], checks.GapHandlerMissing)
			}
			want := "capability_gap:no_handler_for_audit_category:" + cat
			if c.DedupKey != want {
				t.Errorf("%s/%s: dedup key = %q, want %q (one open row per site per category)", cat, page, c.DedupKey, want)
			}
			for key, wantVal := range map[string]string{
				"category":         cat,
				"finding_severity": "high",
				"suggestion":       "point the secondary button at the catalogue",
				"acceptance_test":  "the two CTA hrefs differ",
				"page_name":        page,
			} {
				if got, _ := c.Spec[key].(string); got != wantVal {
					t.Errorf("%s/%s: spec.%s = %q, want %q — the finding's detail must survive for whoever builds the handler", cat, page, key, got, wantVal)
				}
			}
			if bn, _ := c.Spec["builder_needed"].(string); !strings.Contains(bn, "bugs_open/323") {
				t.Errorf("%s/%s: spec.builder_needed should name bugs_open/323: %q", cat, page, bn)
			}
			if !strings.Contains(c.Summary, "no handler for audit category") {
				t.Errorf("%s/%s: summary should say what is missing: %q", cat, page, c.Summary)
			}
		}
	}
}
