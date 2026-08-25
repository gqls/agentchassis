package browserrunner

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestCascadeAttributionIsComposedIntoTheProbe pins the fragments the
// attribution's soundness rests on.
//
// ⚠ WHY THIS EXISTS SEPARATELY, when TestAuditJSComposition already guards the
// probe: that test asserts ONE CONTIGUOUS DIVERGENCE from a pre-2026-08-22
// golden, and its own comment says a second edit elsewhere would merge into the
// same region and fail. It did not fail for this change — the attribution edits
// land BETWEEN the existing common prefix and suffix, so they were absorbed into
// the region 352 had already opened. That is the design of a
// one-region heuristic ageing, not a defect introduced here, and it is recorded
// rather than left for someone to discover: after two changes, that test can no
// longer tell a reviewed edit from an unreviewed one inside the region. These
// assertions are per-fragment and do not decay that way.
func TestCascadeAttributionIsComposedIntoTheProbe(t *testing.T) {
	if !strings.Contains(auditJS, cascadeAttributionJS) {
		t.Fatal("auditJS no longer embeds cascadeAttributionJS byte-for-byte — winningDecl " +
			"would be undefined in the page and every attribution would throw")
	}
	// The shared kernel must be untouched: it is byte-identical to the string
	// the contrast_ratio check also runs, and changing it here would change that
	// check's behaviour silently.
	if strings.Contains(contrastMathsJS, "winningDecl") {
		t.Fatal("attribution has leaked into the SHARED contrast maths kernel; it must stay probe-local")
	}

	// Each fragment is load-bearing, and each is named with what breaks without it.
	fragments := map[string]string{
		"removal test (the proof that the candidate really wins)":             "style.removeProperty(prop)",
		"restore after removal (a mutated page must not be measured further)": "style.setProperty(prop, old, pri)",
		"restore is CHECKED, not assumed":                                     "restored = (getComputedStyle(el).getPropertyValue(prop) === before)",
		"dirty-page latch (stop attributing once a restore failed)":           "window.__cascadeDirty = true",
		"cross-origin sheets are counted, not fatal":                          "catch (e) { opaque++; continue; }",
		"media conditions are re-evaluated rather than assumed to apply":      "window.matchMedia",
		"@supports conditions are re-evaluated":                               "CSS.supports",
		"selector lists are split so the MATCHING part is what is reported":   "String(r.selectorText).split(',')",
		"importance is read per-property, not per-rule":                       "getPropertyPriority(prop)",
		"the element's own style attribute is a candidate":                    "el.style.getPropertyValue(prop)",
		"document order against the theme link is recorded, not assumed":      "compareDocumentPosition",
		"the theme link is found by its deployed path, not by being first":    "/assets/css/styles.css",
	}
	for why, frag := range fragments {
		if !strings.Contains(cascadeAttributionJS, frag) {
			t.Errorf("missing %q — %s", frag, why)
		}
	}

	// The probe must be wired into the finding, budgeted, and its accounting
	// returned; each of these is a way the whole thing goes silently inert.
	for _, frag := range []string{
		"fgWinner:fgw",          // attached to the finding
		"cascBudget",            // per-page cap exists
		"out.cascadeCapped++",   // and what it dropped is counted
		"out.cascadeUnverified", // unproven attributions are counted
		"out.cascadeDirty",      // an abandoned page is reported
	} {
		if !strings.Contains(auditJS, frag) {
			t.Errorf("auditJS is missing %q — the attribution would run and report nothing", frag)
		}
	}
}

// TestCascadeSchemeIsAConstantNotAComputedString. The consumer keys its entire
// routing decision on this string's presence, and a computed one could differ
// between a clean run and a finding-bearing one — which is the exact failure
// selector_scheme was introduced to end.
func TestCascadeSchemeIsAConstantNotAComputedString(t *testing.T) {
	if cascadeSchemeV1 != "cascade/v1" {
		t.Fatalf("cascadeSchemeV1 = %q; a consumer keying on equality would stop trusting good replies", cascadeSchemeV1)
	}
}

// TestCascadeWinnerJSONTags is the adapter half of the mirror contract. The
// chassis decodes these tags by hand in another module; renaming one here
// silently stops that field arriving and reads as an older adapter.
func TestCascadeWinnerJSONTags(t *testing.T) {
	raw, err := json.Marshal(CascadeWinner{
		Property: "color", Surface: surfaceStyleBlock, Selector: ".s .e",
		SheetHref: "/x.css", Important: true, ThemeAfterWinner: true,
		Decl: "var(--x)", VarName: "--x", Verified: true, OpaqueSheets: 1, Candidates: 2,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, tag := range []string{
		"property", "surface", "selector", "sheet_href", "important",
		"theme_after_winner", "decl", "var_name", "verified", "opaque_sheets", "candidates",
	} {
		if _, ok := m[tag]; !ok {
			t.Errorf("JSON tag %q absent — the chassis mirror decodes this name and would "+
				"silently receive a zero value", tag)
		}
	}
	// `verified` must NOT be omitempty: false is the meaningful, load-bearing
	// value, and omitting it would make "could not prove it" indistinguishable
	// from "this adapter does not attribute".
	if !strings.Contains(string(raw), `"verified"`) {
		t.Error(`"verified" is omitted when false; the consumer must be able to see a failed proof`)
	}
}
