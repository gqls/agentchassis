package actions

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Coverage test for the per-slot floors: EVERY action that rewrites an existing
// page_components.rendered_html must either enforce the floors or be listed here
// with a reason.
//
// WHY A COVERAGE TEST AND NOT ANOTHER UNIT TEST. Council round 1 on b30ac52c
// gated on exactly this (bug_historian, high): both floors were wired into
// SavePageSectionsAction only, while nine call sites write that column, so
// "a flattening save through one of them will fail exactly as silently as the
// bug this plan fixes". Wiring the second call site fixes the instance. This
// fixes the CLASS: a tenth writer added later fails this test until its author
// decides, in writing, which side of the line it is on.
//
// It was written because the obvious alternative did not work. After wiring
// ApplySectionEditAction, deleting that wiring again broke NOTHING in the whole
// package — the unit tests exercise the decision functions
// (evaluateSectionShrink, evaluateComponentLoss) and are blind to whether anyone
// calls them. A guard nothing proves is reached is the same defect one level up.
//
// THE KNOWN WEAKNESS OF THIS SHAPE, stated rather than discovered later: it reads
// SOURCE, so it is fooled by a call that is present but unreachable (inside a
// branch that never runs), and it can be satisfied by naming the function in a
// comment. It proves WIRING EXISTS, not that wiring EXECUTES. That is strictly
// more than the zero we had, and the behavioural half belongs in each action's
// own test.

// writesRenderedHTML matches a statement that sets rendered_html on an EXISTING
// row. INSERTs are excluded by construction: a first write has no prior to
// compare against, so the floors are meaningless there.
var updatesRenderedHTML = regexp.MustCompile(`(?is)UPDATE\s+page_components\b[^;]{0,400}?\bSET\b[^;]{0,400}?\brendered_html\s*=`)

// enforcesAFloor: any of the three entry points counts.
var enforcesAFloor = regexp.MustCompile(`enforceSingleSlotFloors|enforceSectionComponentFloor|enforceSectionShrinkFloor`)

// exemptWriters — the audit of 2026-08-13, in code, with the reason each is out
// of scope. A file here is a DECISION, not an oversight. If you add to it, say
// why in the same breath.
var exemptWriters = map[string]string{
	// NOTE — section_editor_actions.go is deliberately NOT exempt. Its
	// content_edit branch enforces the floors, so the file passes on its own
	// merits; exempting it would mean unwiring that call went undetected, which
	// is precisely the hole this test exists to close (verified: removing the
	// call makes this test fail). Its component_swap branch is out of scope for
	// a different reason — component_id, slot_name and html change together, so
	// the markup is SUPPOSED to differ — and that judgement lives in
	// single_slot_floors.go's scope table, not here, because this test's
	// granularity is the file.

	// Writes the ORIGINAL adopted document. There is no prior prose to flatten
	// — this is what creates the prior everything else is measured against.
	"adopt_verbatim.go": "adoption writes the first content; no prior to compare",

	// Narrow, attribute-level rewrites: they change colour declarations, not
	// structure. NOT MEASURED — the reasoning is that they cannot strip a card
	// or a grid because they only touch colour attributes, and that is a code
	// reading, not an experiment. Residual exposure, stated in bugs_open/253.
	"fix_forced_text_colours_action.go": "attribute-level colour rewrite; structure untouched [UNMEASURED]",
	"fix_harcoded_colours_action.go":    "attribute-level colour rewrite; structure untouched [UNMEASURED]",

	// Regenerates a machine-built listing from its source rows. Its markup is
	// generated wholesale every time, so "how much of the previous markup
	// survived" is not a meaningful question about it.
	"rebuild_blog_listing_action.go": "machine-generated listing, regenerated wholesale from source rows",

	// Same shape as the listing above, and FOUND BY THIS TEST rather than by the
	// manual audit — which had it filed as create-only. It looks up its own
	// report component by (page_id, slot_name) and overwrites it with a freshly
	// rendered dossier ("the dossier render is machine-made", its own comment).
	// The row is never a decomposed prose block: it is created and owned by this
	// action, so there is no hand-authored layout for a floor to protect.
	"create_report_page_action.go": "machine-rendered dossier section it owns; regenerated wholesale, never a decomposed prose row",
}

func TestEveryRenderedHTMLRewriterEnforcesTheFloorsOrIsDeclaredExempt(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	var unguarded []string
	var sawAnyWriter bool

	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, readErr := os.ReadFile(f)
		if readErr != nil {
			t.Fatalf("read %s: %v", f, readErr)
		}
		body := string(src)
		if !updatesRenderedHTML.MatchString(body) {
			continue
		}
		sawAnyWriter = true
		if _, exempt := exemptWriters[f]; exempt {
			continue
		}
		if !enforcesAFloor.MatchString(body) {
			unguarded = append(unguarded, f)
		}
	}

	// The detector must have found SOMETHING, or a regex that silently stopped
	// matching would make this test pass by seeing no writers at all — the
	// empty-set trap that makes a green check meaningless.
	if !sawAnyWriter {
		t.Fatal("found NO page_components rendered_html writers — the detector is broken, " +
			"not the codebase; this test cannot pass vacuously")
	}

	if len(unguarded) > 0 {
		t.Fatalf("these actions rewrite page_components.rendered_html without enforcing the "+
			"per-slot floors, and are not declared exempt: %s\n"+
			"Either call enforceSingleSlotFloors (single-row) / enforceSectionComponentFloor "+
			"(whole-page), or add the file to exemptWriters WITH A REASON. "+
			"See bugs_open/253: both floors guarded 1 of 9 writers until 2026-08-13.",
			strings.Join(unguarded, ", "))
	}
}

// The exemption list must not rot into a place where files are parked. Every
// entry has to still be a writer; an entry naming a file that no longer touches
// rendered_html is a stale exemption that would silently cover a future one.
func TestExemptWritersAreAllStillWriters(t *testing.T) {
	for f, reason := range exemptWriters {
		if strings.TrimSpace(reason) == "" {
			t.Errorf("%s is exempt with no reason", f)
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Errorf("exempt file %s does not exist — remove the stale exemption", f)
			continue
		}
		if !updatesRenderedHTML.MatchString(string(src)) {
			t.Errorf("%s is on the exemption list but no longer rewrites rendered_html — "+
				"remove it, or the exemption will silently cover a future write", f)
		}
	}
}
