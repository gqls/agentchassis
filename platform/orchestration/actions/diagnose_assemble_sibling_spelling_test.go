// FILE: platform/orchestration/actions/diagnose_assemble_sibling_spelling_test.go
//
// bugs_open/269 — siblingSignatures rendered methods with a BARE name into a
// section whose own heading invites the model to put those handles in next_scope.
// For a method that handle is ambiguous: analysis.spanOf takes the FIRST match in
// fi.Functions order, so in a file where two types share a method name the bundle
// offered one handle for two different bodies. It does not error. It returns the
// WRONG function's source, labelled as the right one.
//
// THE FIXTURE MUST CONTAIN A COLLISION OR THESE TESTS ASSERT NOTHING (269 §5).
// Every spelling resolves when only one candidate exists — which is exactly why
// the defect survived from bugs_closed/261's fix until now: it is invisible on
// ordinary input. So the fixture is built around two types sharing one method
// name, the collision is asserted before anything else, and the handles are
// checked by RESOLVING them rather than by matching strings.

package actions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/internal/analysis"
)

// collisionSource is analysed by the real analyser rather than hand-described, so
// the receivers and line spans come from the same parser production uses. Hand-
// counted spans are how a fixture ends up agreeing with itself and nothing else.
const collisionSource = `package fixture

type Alpha struct{}

type Beta struct{}

// Handle on Alpha — a VALUE receiver.
func (a Alpha) Handle() string {
	return "alpha handle, and this line is unique to Alpha"
}

// Handle on Beta — a POINTER receiver, same method name. THE COLLISION.
func (b *Beta) Handle() string {
	return "beta handle, and this line is unique to Beta"
}

func (a Alpha) OnlyOnAlpha() string {
	return "only on alpha"
}

func PlainFunction() string {
	return "not a method at all"
}
`

func collisionFixture(t *testing.T) (root string, out analysis.Output) {
	t.Helper()
	root = t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "collide.go"), []byte(collisionSource), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := analysis.Analyse(root)
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	fi := analysis.FindFile(out, "collide.go")
	if fi == nil {
		t.Fatal("fixture: collide.go not in analysis")
	}
	// THE FIXTURE PRECONDITION. Two functions must share the bare name "Handle"
	// with different receivers, or the ambiguity this file exists for is absent.
	var handles int
	for _, fn := range fi.Functions {
		if fn.Name == "Handle" && fn.Receiver != nil {
			handles++
		}
	}
	if handles != 2 {
		t.Fatalf("fixture is wrong: need exactly 2 methods named Handle on different receivers, got %d — without the collision every assertion below is vacuous", handles)
	}
	return root, out
}

// THE BUG. Nothing in scope for this file except a plain function, so both
// colliding methods are siblings — and each must be offered under a handle that
// names one body, not two.
func TestSiblingSpelling_CollidingMethodsAreOfferedUnambiguously(t *testing.T) {
	root, out := collisionFixture(t)

	got := siblingSignatures(out, []string{"collide.go:PlainFunction"}, 6000, bodyCapView{})
	if got == "" {
		t.Fatal("no sibling section at all")
	}

	// Pre-fix this section contained "- `collide.go:Handle` — …" TWICE: one handle,
	// two bodies, and the model told to name it.
	if strings.Contains(got, "`collide.go:Handle`") {
		t.Fatalf("a BARE method handle is still offered — analysis.spanOf resolves it to whichever of the two the analyser listed first.\n--- got ---\n%s", got)
	}
	for _, want := range []string{"`collide.go:(Alpha).Handle`", "`collide.go:(*Beta).Handle`"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s — both colliding methods must be individually addressable.\n--- got ---\n%s", want, got)
		}
	}

	// THE ASSERTION A STRING MATCH CANNOT MAKE. Resolve both offered handles and
	// require DIFFERENT bodies. This is what proves the handles disambiguate, and
	// it is the assertion the bare spelling could never have satisfied.
	alpha, err := analysis.ReadSymbolBody(root, out, "collide.go:(Alpha).Handle")
	if err != nil {
		t.Fatalf("the bundle offered a handle ReadSymbolBody cannot resolve: %v", err)
	}
	beta, err := analysis.ReadSymbolBody(root, out, "collide.go:(*Beta).Handle")
	if err != nil {
		t.Fatalf("the bundle offered a handle ReadSymbolBody cannot resolve: %v", err)
	}
	if alpha == beta {
		t.Fatal("both offered handles resolve to the SAME body — the receiver is not disambiguating")
	}
	if !strings.Contains(alpha, "unique to Alpha") || !strings.Contains(beta, "unique to Beta") {
		t.Errorf("the handles resolve to each other's bodies.\nalpha:\n%s\nbeta:\n%s", alpha, beta)
	}

	// A plain function must keep its bare handle — a receiver appearing on a
	// non-method would be the fix over-reaching.
	if !strings.Contains(got, "`collide.go:OnlyOnAlpha`") && !strings.Contains(got, "`collide.go:(Alpha).OnlyOnAlpha`") {
		t.Errorf("OnlyOnAlpha vanished entirely.\n--- got ---\n%s", got)
	}
}

// bugs_open/269 §2a. The de-duplication was keyed on fn.Name, so a method already
// in scope under its CANONICAL handle — which is the common case, since
// scopeFromCodeResults concatenates the code_symbols spelling straight in — was
// not suppressed, and appeared as a sibling of itself.
func TestSiblingSpelling_MethodInScopeCanonicallyIsNotItsOwnSibling(t *testing.T) {
	_, out := collisionFixture(t)

	got := siblingSignatures(out, []string{"collide.go:(Alpha).Handle"}, 6000, bodyCapView{})
	if strings.Contains(got, "`collide.go:(Alpha).Handle`") {
		t.Fatalf("(Alpha).Handle is IN SCOPE and is listed as its own sibling — budget spent on what the model has already been shown.\n--- got ---\n%s", got)
	}
	// The OTHER Handle is a genuine sibling and must survive: over-suppressing on
	// the shared bare name would hide the one method the model has not seen, which
	// inverts this section's purpose.
	if !strings.Contains(got, "`collide.go:(*Beta).Handle`") {
		t.Fatalf("(*Beta).Handle was suppressed too — a different type's method is not the scoped one.\n--- got ---\n%s", got)
	}
}

// The exactness edge, and the reason the fix tracks first-wins rather than
// suppressing every same-named method. A BARE scope entry resolves the way
// analysis.spanOf does — first match in fi.Functions order — so exactly one of
// the two Handles is in scope, and the other is still a sibling worth listing.
func TestSiblingSpelling_BareScopeEntrySuppressesOnlyTheOneItResolvesTo(t *testing.T) {
	root, out := collisionFixture(t)

	// Which one does a bare "Handle" actually resolve to? Ask the resolver rather
	// than assuming declaration order, so this test cannot drift from spanOf.
	bare, err := analysis.ReadSymbolBody(root, out, "collide.go:Handle")
	if err != nil {
		t.Fatalf("fixture: bare Handle should still resolve (first wins): %v", err)
	}
	resolvedToAlpha := strings.Contains(bare, "unique to Alpha")

	got := siblingSignatures(out, []string{"collide.go:Handle"}, 6000, bodyCapView{})
	suppressed, survivor := "`collide.go:(Alpha).Handle`", "`collide.go:(*Beta).Handle`"
	if !resolvedToAlpha {
		suppressed, survivor = survivor, suppressed
	}
	if strings.Contains(got, suppressed) {
		t.Errorf("%s is what the bare scope entry resolves to, so it is IN SCOPE and must not be listed as a sibling.\n--- got ---\n%s", suppressed, got)
	}
	if !strings.Contains(got, survivor) {
		t.Errorf("%s is a DIFFERENT type's method and was suppressed by the shared bare name — that is over-suppression, not de-duplication.\n--- got ---\n%s", survivor, got)
	}
}
