package actions

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// The axis these tests pin is the one that let a live article body be replaced
// by a stylesheet (bugs_closed/285). They are written against the SHAPE of the
// real pair, with the real measured numbers in the assertions, so a future
// "simplification" of visibleTextLength that stops excluding <style>/<script>
// content fails here rather than on a served page.

// portedArticleFixture reproduces the shape of the row that was overwritten:
// a small inline stylesheet plus a real article body.
func portedArticleFixture(proseChars int) string {
	css := strings.Repeat(".article-content h3{color:var(--primary);margin-top:2.5rem;}", 8)
	prose := strings.Repeat("word ", proseChars/5)
	return `<section class="ported-page" data-component="ported-page"><style>` + css +
		`</style><h1>The Content-First Strategy</h1><div class="article-content"><p class="lead">` + prose + `</p></div></section>`
}

// poisonFixture reproduces the shape that replaced it: a LARGER stylesheet, an
// EMPTY article, a comment, and a script that would build the visible content
// client-side. Reader-visible static text is almost nothing.
func poisonFixture(cssChars int) string {
	css := strings.Repeat(".ported-page-section .asset-row{display:flex;align-items:center;}", cssChars/60)
	js := strings.Repeat("var seededRows=[{title:'Checklist',meta:'PDF',href:'/a.pdf'}];", 20)
	return `<style>` + css + `</style><section class="ported-page-section"><div class="ported-page-container">` +
		`<article class="ported-page-content"></article>` +
		`<div class="asset-list" id="portedPageAssetList"><h2>Related Downloads</h2>` +
		`<!-- .asset-row anchors are seeded via JS below so the selector always resolves --></div>` +
		`</div></section><script>` + js + `</script>`
}

func TestVisibleTextLengthExcludesStylesheetScriptAndComments(t *testing.T) {
	cases := []struct {
		name string
		html string
		want int
	}{
		{"plain text", `<p>hello there</p>`, len("hellothere")},
		{"stylesheet content is not text", `<style>.a{color:#fff;background:#000;padding:2rem}</style><p>hi</p>`, 2},
		{"script content is not text", `<script>var x = "a long javascript string literal";</script><p>hi</p>`, 2},
		{"comment content is not text", `<!-- a long explanatory comment about seeding --><p>hi</p>`, 2},
		{"entities are dropped, both of them (&amp; is an entity too)", `<p>a&nbsp;&amp;b</p>`, 2},
		{"case and attributes on the style tag", `<STYLE type="text/css">.a{color:red}</STYLE><p>hi</p>`, 2},
		{"an empty article is empty", `<article class="ported-page-content"></article>`, 0},
	}
	for _, c := range cases {
		if got := visibleTextLength(c.html); got != c.want {
			t.Errorf("%s: visibleTextLength = %d, want %d", c.name, got, c.want)
		}
	}
}

// The regression the axis change exists for: the OLD axis (tag-stripped, still
// used by the whole-page path) reports GROWTH on the poisoning pair, so it
// cannot refuse it. Both halves are asserted — if either stops holding, the
// calibration in the header no longer describes the code.
func TestTagStrippedAxisGrowsWhileVisibleTextCollapses(t *testing.T) {
	existing := portedArticleFixture(2750)
	incoming := poisonFixture(7000)

	oldExisting := len(strings.TrimSpace(shrinkGuardTagStripper.ReplaceAllString(existing, "")))
	oldIncoming := len(strings.TrimSpace(shrinkGuardTagStripper.ReplaceAllString(incoming, "")))
	if oldIncoming <= oldExisting {
		t.Fatalf("fixture does not reproduce the trap: the tag-stripped axis must GROW (%d → %d)", oldExisting, oldIncoming)
	}
	if ratio := float64(oldIncoming) / float64(oldExisting); ratio < defaultSectionShrinkFloor {
		t.Fatalf("tag-stripped axis would have refused (%.2f < %.2f) — fixture wrong, the real pair read 2.62", ratio, defaultSectionShrinkFloor)
	}

	newExisting := visibleTextLength(existing)
	newIncoming := visibleTextLength(incoming)
	if newExisting < minShrinkGuardChars {
		t.Fatalf("fixture's existing side (%d visible) is below evaluateSectionShrink's own minimum %d — it would be exempt, not refused",
			newExisting, minShrinkGuardChars)
	}
	ratio := float64(newIncoming) / float64(newExisting)
	if ratio >= defaultSectionShrinkFloor {
		t.Errorf("visible-text axis kept %.0f%% (floor %.0f%%) — this axis must REFUSE the write that emptied the article",
			ratio*100, defaultSectionShrinkFloor*100)
	}
}

// The other direction, and the reason this is a correction rather than a second
// floor: putting the article BACK reads 38% on the tag-stripped axis (a refusal
// of a repair) and passes on the visible axis. Measured on the real pair,
// 2026-08-15 18:18Z, seed 431.
func TestRepairingAHollowedSlotIsRefusedByTheOldAxisAndAllowedByThisOne(t *testing.T) {
	hollow := poisonFixture(7000)          // what the slot held
	repaired := portedArticleFixture(2750) // what the restore put back

	oldRatio := float64(len(strings.TrimSpace(shrinkGuardTagStripper.ReplaceAllString(repaired, "")))) /
		float64(len(strings.TrimSpace(shrinkGuardTagStripper.ReplaceAllString(hollow, ""))))
	if oldRatio >= defaultSectionShrinkFloor {
		t.Fatalf("fixture does not reproduce the false positive: the old axis kept %.2f, expected below %.2f", oldRatio, defaultSectionShrinkFloor)
	}

	// On the visible axis the repair grows, so the floor cannot fire...
	if visibleTextLength(repaired) <= visibleTextLength(hollow) {
		t.Errorf("visible text must GROW on a repair: %d → %d", visibleTextLength(hollow), visibleTextLength(repaired))
	}
	// ...and it is ALSO out of scope by evaluateSectionShrink's own 500-char
	// minimum, which is the belt to that braces: a hollowed slot must stay
	// writable however the ratio falls.
	if visibleTextLength(hollow) >= minShrinkGuardChars {
		t.Errorf("a hollowed slot (%d visible chars) should sit below minShrinkGuardChars %d so a repair is never gated by a ratio",
			visibleTextLength(hollow), minShrinkGuardChars)
	}
}

// The governing minimum is evaluateSectionShrink's own minShrinkGuardChars, and
// this pins what it must and must not exempt on the VISIBLE axis: short captions
// out, real article bodies in. If someone raises it, the population this floor
// protects (180 of 198 live at-risk rows have ≥500 visible chars) shrinks.
func TestTheGoverningMinimumStillProtectsRealBodiesAndExemptsCaptions(t *testing.T) {
	if visibleTextLength(`<div class="cta"><p>Read more about our approach to content</p></div>`) >= minShrinkGuardChars {
		t.Error("a short caption must be out of scope, or every button-caption edit is ratio-gated")
	}
	if v := visibleTextLength(portedArticleFixture(2750)); v < minShrinkGuardChars {
		t.Errorf("a real article body (%d visible) must be IN scope or the floor protects nothing", v)
	}
}

// Guards the fixtures themselves: a test whose fixture stops reproducing the
// shape it names silently stops testing anything.
func TestFixturesReproduceTheMeasuredShape(t *testing.T) {
	existing, incoming := portedArticleFixture(2750), poisonFixture(7000)
	for name, h := range map[string]string{"existing": existing, "incoming": incoming} {
		if !strings.Contains(h, "<style>") && !strings.Contains(h, "<STYLE") {
			t.Errorf("%s fixture must carry a stylesheet — that is the string the old axis counted as text", name)
		}
	}
	if !strings.Contains(incoming, `<article class="ported-page-content"></article>`) {
		t.Error("incoming fixture must carry the EMPTY article: an empty body with a big stylesheet is the defect's signature")
	}
	if got := visibleTextLength(incoming); got > 40 {
		t.Errorf("incoming fixture has %d visible chars; the real poison had 68 in a 8,855-char row — keep it small: %s",
			got, fmt.Sprintf("%.60s…", incoming))
	}
}

// THE WIRING, not just the arithmetic. The tests above all pass with the axis
// reverted at the call site — measured 2026-08-17 by doing exactly that — which
// is the "a guard nothing proves is reached" hole one level up. This one drives
// enforceSingleSlotFloors itself, so unwiring visibleTextLength fails the build.
//
// params.DB is nil on purpose: emitPruneRefusalWorkItem returns early on a nil
// handle (prune_floor.go:353), so the decision is exercised without a database.
// Both fixtures carry fewer than minComponentGuardClasses (10) class attributes,
// so the COMPONENT floor is out of scope by its own rule and the verdict here is
// the text axis speaking alone.
func TestEnforceSingleSlotFloors_RefusesTheStylesheetForProseSwap(t *testing.T) {
	params := ActionParams{
		Logger:           zap.NewNop(),
		DB:               nil,
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
	}
	article, poison := portedArticleFixture(2750), poisonFixture(7000)

	if n := countComponentClasses(article); n >= minComponentGuardClasses {
		t.Fatalf("fixture carries %d class attributes — at or above the component floor's minimum %d, so this test would not isolate the text axis",
			n, minComponentGuardClasses)
	}

	err := enforceSingleSlotFloors(context.Background(), params, uuid.Nil, uuid.Nil,
		"learn-ai-builders-content-first", "ported-page", article, poison)
	if err == nil {
		t.Fatal("enforceSingleSlotFloors ALLOWED the write that replaced a live article body with a stylesheet " +
			"(bugs_closed/285). The tag-stripped axis reads 262% retained here; only the visible-text axis refuses it.")
	}
	if !strings.Contains(err.Error(), "VISIBLE text") {
		t.Errorf("refusal should name the axis so the operator knows what was measured; got: %v", err)
	}

	// And the repair direction must still be permitted, or the guard blocks the
	// only thing that fixes the damage.
	if err := enforceSingleSlotFloors(context.Background(), params, uuid.Nil, uuid.Nil,
		"learn-ai-builders-content-first", "ported-page", poison, article); err != nil {
		t.Errorf("enforceSingleSlotFloors REFUSED the repair that put the article back (seed 431's write): %v", err)
	}
}

// The absolute minimum, isolated. Without this test the minimum is a clause
// nothing exercises: removing it kept every other test green (measured
// 2026-08-17, MUT-2), because on the repair pair the visible text GROWS and the
// ratio never speaks. So drive the case the minimum exists for — a short caption
// legitimately shrinking well past 50% — and assert it is ALLOWED.
func TestEnforceSingleSlotFloors_ShortCaptionEditIsNotRatioGated(t *testing.T) {
	params := ActionParams{
		Logger:           zap.NewNop(),
		DB:               nil,
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
	}
	existing := `<div class="cta"><p>Read more about our approach to content</p></div>`
	incoming := `<div class="cta"><p>Read more</p></div>`

	if v := visibleTextLength(existing); v >= minShrinkGuardChars {
		t.Fatalf("fixture is %d visible chars — at or above minShrinkGuardChars %d, so it would not isolate the exemption",
			v, minShrinkGuardChars)
	}
	ratio := float64(visibleTextLength(incoming)) / float64(visibleTextLength(existing))
	if ratio >= defaultSectionShrinkFloor {
		t.Fatalf("fixture does not shrink past the floor (%.2f) — it would pass anyway and prove nothing", ratio)
	}

	if err := enforceSingleSlotFloors(context.Background(), params, uuid.Nil, uuid.Nil,
		"home", "cta-0", existing, incoming); err != nil {
		t.Errorf("a %d→%d visible-char caption edit must NOT be refused: it is under minShrinkGuardChars, "+
			"where a ratio is noise. Got: %v",
			visibleTextLength(existing), visibleTextLength(incoming), err)
	}
}
