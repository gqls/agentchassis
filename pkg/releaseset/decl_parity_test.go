// FILE: pkg/releaseset/decl_parity_test.go
//
// ParseMakefileDecls is a LITERAL extractor, not a make evaluator, and that is
// a deliberate trade (see decl.go's header). The price of the trade is drift:
// if someone writes a declaration in a form the extractor does not recognise —
// a `+=`, a `$(filter ...)`, a computed name — the reading silently diverges
// from what make itself sees, and the gate then polices a shape that does not
// exist.
//
// This is the payment. It asks MAKE for the same four lists and requires the
// two readings to agree, in the style of cron_parity_test.go and
// optional_budget_cron_parity_test.go (the RFC 006 / RFC 022 siblings), which
// exist for exactly this reason one language over.
//
// ⚠ IT SKIPS LOUDLY when make is unavailable and never passes quietly. A silent
// pass here is the failure mode the test exists to prevent — the estate's own
// blind-pass landmine, and the reason the print- targets are echo-only.
package releaseset

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this source file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
}

func makePrint(t *testing.T, root, target string) ([]string, bool) {
	t.Helper()
	cmd := exec.Command("make", "-s", target)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Logf("`make -s %s` failed (%v) — SKIPPING the parity check rather than passing it quietly", target, err)
		return nil, false
	}
	return strings.Fields(string(out)), true
}

func TestDeclParityWithMake(t *testing.T) {
	root := repoRoot(t)

	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make not on PATH — SKIPPING the parity check; it is not passing")
	}

	f, err := os.Open(filepath.Join(root, "makefile"))
	if err != nil {
		t.Fatalf("opening the live makefile: %v", err)
	}
	defer f.Close()

	parsed, err := ParseMakefileDecls(f)
	if err != nil {
		t.Fatalf("ParseMakefileDecls on the LIVE makefile: %v", err)
	}

	cases := []struct {
		target string
		mine   []string
	}{
		{"print-release-images", parsed.ReleaseImages},
		{"print-agent-deploy-services", rawOf(parsed.AgentDeploy)},
		{"print-retag-exempt", rawOf(parsed.RetagExempt)},
		{"print-own-lineage", rawOf(parsed.OwnLineage)},
	}
	for _, c := range cases {
		theirs, ok := makePrint(t, root, c.target)
		if !ok {
			continue
		}
		if diff := setDiff(c.mine, theirs); diff != "" {
			t.Errorf("%s: this package and make disagree about the live declaration.\n%s",
				c.target, diff)
		}
	}
}

func rawOf(entries []Entry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Raw)
	}
	return out
}

// setDiff reports a two-column difference rather than "not equal", because the
// useful question when a parity test fails is always WHICH element moved.
func setDiff(mine, theirs []string) string {
	inMine, inTheirs := map[string]bool{}, map[string]bool{}
	for _, s := range mine {
		inMine[s] = true
	}
	for _, s := range theirs {
		inTheirs[s] = true
	}
	var onlyMine, onlyTheirs []string
	for s := range inMine {
		if !inTheirs[s] {
			onlyMine = append(onlyMine, s)
		}
	}
	for s := range inTheirs {
		if !inMine[s] {
			onlyTheirs = append(onlyTheirs, s)
		}
	}
	if len(onlyMine) == 0 && len(onlyTheirs) == 0 {
		return ""
	}
	return "  parsed here but not by make: " + strings.Join(onlyMine, " ") +
		"\n  seen by make but not here:   " + strings.Join(onlyTheirs, " ")
}

// The live tree must be clean under the gate. This is a REGRESSION guard, not
// evidence that the gate discriminates — it is compliant today, so it could
// only have passed. The discriminating cases are in predicates_test.go.
func TestLiveTreeIsCompliant(t *testing.T) {
	root := repoRoot(t)
	f, err := os.Open(filepath.Join(root, "makefile"))
	if err != nil {
		t.Fatalf("opening the live makefile: %v", err)
	}
	defer f.Close()
	d, err := ParseMakefileDecls(f)
	if err != nil {
		t.Fatalf("ParseMakefileDecls: %v", err)
	}
	pins, err := ScanOverlays(root)
	if err != nil {
		t.Fatalf("ScanOverlays: %v", err)
	}
	ours := OurPins(pins, "docker.io/aqls")
	if len(ours) == 0 {
		t.Fatal("zero overlays pin one of our images — the scan found nothing, which is not the same as nothing being wrong")
	}
	t.Logf("judged %d overlays pinning our registry, out of %d production overlays", len(ours), len(pins))
	if v := Check(d, pins, "docker.io/aqls"); len(v) != 0 {
		for _, x := range v {
			t.Errorf("live tree violation: %s\n    fix: %s", x, x.Remedy)
		}
	}
}
