// FILE: platform/orchestration/actions/diagnose_assemble_deadend_tail_test.go
//
// bugs_open/273 (bugs_closed/261 §8 follow-up 2) — for a file that can never
// render whole, the sibling section's "+N more" marker said "Name symbols
// individually" while withholding the names. The elided tail was unreachable by
// ANY advice the bundle gave: retrieval had already failed to surface those
// symbols (or they would be in scope), the bare-path re-read is refuted by the
// file's own size, and the handles were never shown. 261's worked case
// (coordinator.go, 169,139 chars against a 60,000 budget, 91 functions, ~10
// listed) lost the three functions its run needed behind exactly this marker.
//
// The fix appends the elided functions' CANONICAL handles to that one marker.
// Every test here is written to FAIL against the pre-fix code except the two
// byte-identity pins, which are the negative controls: the could-fit and
// unknown-size branches, and any file whose signatures all fit, must not move.
//
// Fixture arithmetic is asserted before every assertion that depends on it (a
// cap test that never trips the cap asserts nothing — 016b §9), and method
// fixtures carry receivers so CanonicalSymbolName emits the producer's real
// "(*Recv).Name" spelling — asserting the spelling the function already
// implements is how 261 survived its own test file (LANDMINES, "two grammars").

package actions

import (
	"fmt"
	"strings"
	"testing"
)

func TestDeadEndTail_ElidedHandlesAreListedCanonically(t *testing.T) {
	var fns []fnSpec
	for i := 0; i < 30; i++ {
		fns = append(fns, fnSpec{name: fmt.Sprintf("methodNumber%02d", i), recv: "*Big", lines: 3})
	}
	root, out := overCapFixture(t, "big.go", fns)
	fileChars := wholeFileSize(t, root, "big.go")

	const sigCap = 600 // small enough that the per-file share elides most of the 30
	got := siblingSignatures(out, []string{"big.go:(*Big).methodNumber00"}, sigCap,
		bodyCapView{repoRoot: root, budget: fileChars - 1})

	if !strings.Contains(got, "more in this file") {
		t.Fatalf("fixture is wrong: no +N-more marker, nothing under test.\n--- got ---\n%s", got)
	}
	if strings.Contains(got, "put the bare file path in next_scope to see it whole") {
		t.Fatalf("dead-end file offered whole — the fixture missed the dead-end branch.\n--- got ---\n%s", got)
	}
	// THE BUG: every sibling must now be visible SOMEWHERE — a head line
	// (`path:canon` with a signature) or a tail handle (`canon` bare). The canon
	// substring appears in both forms, so Contains is "shown somewhere" and the
	// pre-fix code fails this for every function past the per-file share.
	for i := 0; i < 30; i++ {
		canon := fmt.Sprintf("(*Big).methodNumber%02d", i)
		if i == 0 {
			// in scope — must appear in NEITHER (it is not its own sibling, 269 §2a)
			if strings.Count(got, canon) != 0 {
				t.Errorf("in-scope %s leaked into its own sibling section.\n--- got ---\n%s", canon, got)
			}
			continue
		}
		if !strings.Contains(got, canon) {
			t.Errorf("sibling %s is invisible: not in the head lines and not in the tail — the 261 §8.2 defect.\n--- got ---\n%s", canon, got)
		}
	}
	// The instruction must name the exact next_scope syntax with the real path.
	if !strings.Contains(got, "`big.go:<handle>`") {
		t.Errorf("the tail does not say how to turn a handle into a scope entry.\n--- got ---\n%s", got)
	}
	// The census phrases must not leak into the sibling section (LANDMINES: the
	// 267 §4b trend queries discriminate on these exact strings in the body).
	for _, phrase := range []string{"did not fit", "could not be read", "read it whole", "NO next_scope can render this path"} {
		if strings.Contains(got, phrase) {
			t.Errorf("censused marker phrase %q leaked into the sibling section.\n--- got ---\n%s", phrase, got)
		}
	}
}

// The tail has its own cap, and past it the marker must say what was withheld
// and name a remedy that can enumerate the remainder — never trail off silently
// (that would be this bug again, one layer down).
func TestDeadEndTail_OverflowIsCountedAndGivenARemedy(t *testing.T) {
	// Enough long-named methods that the tail alone exceeds siblingDeadEndTailCap:
	// ~55 chars per handle incl. separators × 120 ≈ 6.6KB > 4000.
	var fns []fnSpec
	for i := 0; i < 120; i++ {
		fns = append(fns, fnSpec{name: fmt.Sprintf("aVeryLongDescriptiveMethodNameForOverflow%03d", i), recv: "*Big", lines: 1})
	}
	root, out := overCapFixture(t, "huge.go", fns)
	fileChars := wholeFileSize(t, root, "huge.go")

	const sigCap = 600
	got := siblingSignatures(out, []string{"huge.go:(*Big).aVeryLongDescriptiveMethodNameForOverflow000"}, sigCap,
		bodyCapView{repoRoot: root, budget: fileChars - 1})

	if !strings.Contains(got, "past even this list's cap") {
		t.Fatalf("fixture is wrong or the tail cap is unbounded: 120 long handles did not overflow it.\n--- got ---\n%s", got)
	}
	if !strings.Contains(got, `code_request of kind "symbol", query "huge.go"`) {
		t.Errorf("the overflow names no satisfiable remedy for the remainder.\n--- got ---\n%s", got)
	}
	// The residual count and the listed handles must add up to every sibling:
	// nothing silently dropped between the list and the count.
	listed := strings.Count(got, "(*Big).aVeryLongDescriptiveMethodNameForOverflow")
	var residual int
	if _, err := fmt.Sscanf(got[strings.Index(got, "…and "):], "…and %d more", &residual); err != nil {
		t.Fatalf("cannot read the residual count: %v\n--- got ---\n%s", err, got)
	}
	// 119 siblings (one in scope); head lines also use the canonical spelling, so
	// `listed` counts head + tail occurrences, which is exactly "shown somewhere".
	if listed+residual != 119 {
		t.Errorf("shown (%d) + withheld (%d) = %d, want 119 — handles are being lost without being counted.\n--- got ---\n%s",
			listed, residual, listed+residual, got)
	}
}

// Byte-identity pins: the could-fit and unknown-size branches, and the global
// guard's behaviour for ordinary files, must not move. These are the negative
// controls — they pass before AND after the fix, and fail only if the fix
// widens beyond the dead-end branch.
func TestDeadEndTail_CouldFitAndUnknownBranchesAreByteIdentical(t *testing.T) {
	var fns []fnSpec
	for i := 0; i < 40; i++ {
		fns = append(fns, fnSpec{name: fmt.Sprintf("helperNumber%02d", i), lines: 2})
	}
	root, out := overCapFixture(t, "many.go", fns)
	fileChars := wholeFileSize(t, root, "many.go")
	const sigCap = 600

	fits := siblingSignatures(out, []string{"many.go:helperNumber00"}, sigCap,
		bodyCapView{repoRoot: root, budget: fileChars + 1})
	if !strings.Contains(fits, "put the bare file path in next_scope to see it whole") {
		t.Fatalf("the could-fit branch lost its advice — the fix widened past the dead end.\n--- got ---\n%s", fits)
	}
	if strings.Contains(fits, "The elided handles:") {
		t.Errorf("a tail was appended for a file the budget CAN render — the invitation there is satisfiable and cheaper.\n--- got ---\n%s", fits)
	}
	unknown := siblingSignatures(out, []string{"many.go:helperNumber00"}, sigCap, bodyCapView{})
	if unknown != fits {
		t.Errorf("unknown size must degrade to the could-fit wording byte for byte.\n--- unknown ---\n%s\n--- fits ---\n%s", unknown, fits)
	}
}

// The guard exemption: a single scoped dead-end file whose head consumes the
// whole per-file share must still render its section, tail and all. Counting
// the tail against the global guard would evict the section on exactly the
// motivating case (one over-budget file in scope), replacing the model's only
// map of the file with "further files omitted".
func TestDeadEndTail_DoesNotEvictItsOwnSection(t *testing.T) {
	var fns []fnSpec
	for i := 0; i < 80; i++ {
		fns = append(fns, fnSpec{name: fmt.Sprintf("methodWithARealisticLengthName%02d", i), recv: "*Big", lines: 2})
	}
	root, out := overCapFixture(t, "giant.go", fns)
	fileChars := wholeFileSize(t, root, "giant.go")

	// One scoped file: perFile == capChars, so the head fills the share and the
	// tail rides on top — the section exceeds capChars*5/4 iff the tail counts.
	const sigCap = 2000
	got := siblingSignatures(out, []string{"giant.go:(*Big).methodWithARealisticLengthName00"}, sigCap,
		bodyCapView{repoRoot: root, budget: fileChars - 1})

	if strings.Contains(got, "further files omitted") {
		t.Fatalf("the dead-end tail evicted its own section: the guard is counting tail bytes.\n--- got ---\n%s", got)
	}
	if !strings.Contains(got, "The elided handles:") {
		t.Fatalf("no tail rendered at all — the fixture missed the dead-end branch.\n--- got ---\n%s", got)
	}
	// And a SECOND file after the giant must still get its section: the guard
	// exemption must not let the tail spend the later files' shares either.
	fns2 := []fnSpec{{name: "tinyHelper", lines: 2}, {name: "otherHelper", lines: 2}}
	root2, out2 := overCapFixture(t, "small.go", fns2)
	_ = root2
	combined := out
	combined.Files = append(combined.Files, out2.Files[0])
	got2 := siblingSignatures(combined, []string{
		"giant.go:(*Big).methodWithARealisticLengthName00", "small.go:tinyHelper"}, sigCap,
		bodyCapView{repoRoot: root, budget: fileChars - 1})
	if !strings.Contains(got2, "**small.go**") {
		t.Errorf("the file after the dead-end giant lost its section.\n--- got ---\n%s", got2)
	}
	if !strings.Contains(got2, "`small.go:otherHelper`") {
		t.Errorf("the second file's sibling line is missing.\n--- got ---\n%s", got2)
	}
}
