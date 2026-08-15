// FILE: platform/orchestration/actions/work_item_type_minting_ratchet_test.go
//
// OWNER DIRECTIVE 2026-08-15 (bugs_open/279 follow-up): "add some sort of check to
// stop agents writing triage items with labels that don't exist. It's happened
// before."
//
// THE FAILURE MODE. A work item's item_type is the join between who files it and
// what can act on it. The vocabulary exists only in Go source (routing sets,
// registries, agent handler names) — the database cannot tell a real label from an
// invented one, so an item filed under a constructed type inserts cleanly, reports
// success, and then nothing ever claims it. write_audit_findings did exactly this
// ("audit_finding_" + category) and the rows died in 'detected' for weeks
// (bugs_open/115, fixed in bugs_open/279 by filing capability_gap instead).
//
// WHAT THIS RATCHET DOES. It scans this package and discovery_checks for source
// lines that BUILD an item type dynamically — string concatenation or fmt.Sprintf
// flowing into an ItemType/itemType assignment — and fails on any hit. Measured at
// the commit that introduced it: zero remain (the write_audit_findings minting line
// was the only one in the tree, all spellings grepped). A literal item type is
// covered by the other guards (the discovery_checks source sensor, the
// classifyEmittableItemTypes closed set, verifier_coverage's ratchet); a
// CONSTRUCTED one bypasses them all, which is why construction itself is the thing
// this test bans.
//
// WHAT IT CANNOT SEE, stated so nobody mistakes the coverage: item types minted
// from WORKFLOW CONFIG (the generic create_work_item action takes item_type from
// step config — `section_edit` reaches site_work_items with no Go literal
// anywhere). That path's backstop is claim-time (ClaimWorkItemAction blocks items
// whose handler_agent resolves to no live agent_definitions row); a write-time
// vocabulary check there needs a registry that does not exist yet and is recorded
// in bugs_open/279 as the residual.
//
// COMMENTS ARE STRIPPED BEFORE MATCHING — a prose mention of the old minting line
// must not fail the build (the a-source-scanning-test-makes-comments-load-bearing
// trap, memorised fleet-wide). The stripper is line-based (`//` to end of line),
// which would also eat a `//` inside a string literal on the same line; for the
// patterns matched here that costs nothing, and the error direction is a missed
// exotic case, never a false failure.

package actions

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// itemTypeConstruction matches an ItemType/itemType assignment whose value is
// built on the same line by concatenation or Sprintf. The two spellings cover
// the exported struct field and the local-variable/lowercase-field idiom
// (workItem.itemType, itemType :=).
var itemTypeConstruction = regexp.MustCompile(
	`\b[Ii]temType\s*[:=][^,\n]*("\s*\+|\+\s*"|fmt\.Sprintf)`)

func TestNoDynamicallyConstructedItemTypes(t *testing.T) {
	dirs := []string{".", "discovery_checks"}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			path := filepath.Join(dir, name)
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			for i, line := range strings.Split(string(src), "\n") {
				if idx := strings.Index(line, "//"); idx >= 0 {
					line = line[:idx]
				}
				if itemTypeConstruction.MatchString(line) {
					t.Errorf("%s:%d constructs an item_type dynamically:\n\t%s\n"+
						"An invented label inserts cleanly and is then claimable by nothing — "+
						"the rows die open (bugs_open/115, bugs_open/279). Use a literal from "+
						"the routing vocabulary, or file the finding as capability_gap "+
						"(the platform's \"work I have no handler for\" shape) if no label fits.",
						path, i+1, strings.TrimSpace(line))
				}
			}
		}
	}
}

// TestRatchetPatternStillBites proves the regex actually matches the shape it
// bans — a ratchet whose pattern rots (e.g. after a rename) would pass forever
// on an empty match set, which is the a-quiet-test-passes-when-the-rule-is-gone
// trap. The examples are the real minting line this ratchet exists because of,
// in both spellings and both construction styles.
func TestRatchetPatternStillBites(t *testing.T) {
	for _, bad := range []string{
		`ItemType:     "audit_finding_" + category,`,
		`itemType := "audit_finding_" + category`,
		`ItemType: fmt.Sprintf("audit_%s", category),`,
		`itemType = prefix + "_finding"`,
	} {
		if !itemTypeConstruction.MatchString(bad) {
			t.Errorf("ratchet pattern no longer matches %q — the guard is disarmed", bad)
		}
	}
	for _, good := range []string{
		`ItemType:     "capability_gap",`,
		`ItemType: itemType,`,
		`itemType := "cta_improvement"`,
		`ItemType: designItemTypes[category],`,
		`DedupKey: fmt.Sprintf("capability_gap:%s", category),`,
	} {
		if itemTypeConstruction.MatchString(good) {
			t.Errorf("ratchet pattern false-positives on legitimate line %q", good)
		}
	}
}
