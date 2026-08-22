// FILE: cmd/releasecheck/main.go
//
// The release-coverage gate, invoked by the makefile's check-release-coverage
// target (which deploy-core requires, which `make release` runs).
//
// THIS FILE IS DELIBERATELY THIN, and the layering is load-bearing rather than
// stylistic. `scripts/council-scope.sh` scopes review to platform/, internal/,
// pkg/ and appliable migrations — `cmd/` is OUTSIDE it. A helper written wholly
// here would draw no council round at all, which is precisely the trap
// bugs_open/318 warned about in its own fix-candidate list ("decide the shape
// before assuming a review is available"). Every decision therefore lives in
// pkg/releaseset; main only parses flags and prints. Same pattern as
// cmd/regcheck: the operator CLI runs the SAME code as the gate, never a
// re-implementation that can drift.
//
// EXIT CODES — and the 1-vs-2 distinction is also printed IN WORDS, because the
// makefile invokes this through `go run`, which collapses a non-zero child exit
// into its own 1. [MEASURED 2026-08-22, not assumed: a throwaway program doing
// os.Exit(2) gives `go run .` → 1 and its compiled binary → 2.] The makefile
// treats any non-zero as failure, so the release stops correctly either way;
// what is lost is an operator's ability to tell "the release is wrong" from
// "the check is broken", which is why the message says which:
//
//	0  every overlay pinning one of our images is in a release path, or is
//	   explicitly exempt in OWN_LINEAGE
//	1  at least one is not — the release must not proceed
//	2  the check COULD NOT RUN (unparseable makefile, unreadable deployment
//	   tree). Never conflated with 0: a gate that passes what it failed to
//	   measure is this estate's own blind-pass landmine.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gqls/agentchassis/pkg/releaseset"
)

const (
	exitOK        = 0
	exitViolation = 1
	exitCannotRun = 2
)

// ANSI colours, matched to the makefile's own palette so the gate's output
// looks like the rest of a release.
const (
	red    = "\033[0;31m"
	green  = "\033[0;32m"
	yellow = "\033[1;33m"
	dim    = "\033[2m"
	reset  = "\033[0m"
)

func main() {
	var (
		root     = flag.String("root", ".", "repository root to read the makefile and deployment tree from")
		registry = flag.String("registry", "docker.io/aqls", "our image registry prefix; overlays pinning anything else are not ours to roll")
		quiet    = flag.Bool("quiet", false, "print only violations")
	)
	flag.Parse()

	code, err := run(*root, *registry, *quiet)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sRELEASE COVERAGE: THE CHECK COULD NOT RUN%s — %v\n", red, reset, err)
		fmt.Fprintf(os.Stderr, "%s  This is exit %d, NOT a pass. Nothing was measured, so nothing was cleared.%s\n",
			yellow, exitCannotRun, reset)
		os.Exit(exitCannotRun)
	}
	os.Exit(code)
}

// printExemptions names the standing OWN_LINEAGE entries on EVERY green run.
//
// This is the half of the accumulation guard that does the work. The budget in
// pkg/releaseset only speaks when it is exceeded, and a threshold that is silent
// until it trips is a threshold nobody is watching — the estate's own lesson
// about detectors with no reader. Printing the list every time puts the count in
// front of whoever runs the release, so the FOURTH exemption is noticed as it
// arrives rather than when a number is crossed.
//
// Deliberately not a warning and not coloured as one: an exemption is a
// legitimate, reviewed state. It is reported, not complained about.
func printExemptions(d releaseset.Decl) {
	if len(d.OwnLineage) == 0 {
		fmt.Printf("%s  No service is excused from the release (OWN_LINEAGE is empty).%s\n", dim, reset)
		return
	}
	verb := "are"
	if len(d.OwnLineage) == 1 {
		verb = "is"
	}
	fmt.Printf("%s  %d of those %s EXCUSED from the release (OWN_LINEAGE, budget %d) — each moved by its own target:%s\n",
		dim, len(d.OwnLineage), verb, releaseset.ExemptionBudget, reset)
	for _, e := range d.OwnLineage {
		target := e.Qualifier
		if target == "" {
			target = "(none named — this is a violation and will be reported above)"
		}
		fmt.Printf("%s      %s  ->  %s%s\n", dim, e.Service, target, reset)
	}
}

func run(root, registry string, quiet bool) (int, error) {
	mf, err := os.Open(root + "/makefile")
	if err != nil {
		return exitCannotRun, fmt.Errorf("opening %s/makefile: %w", root, err)
	}
	defer mf.Close()

	decl, err := releaseset.ParseMakefileDecls(mf)
	if err != nil {
		return exitCannotRun, err
	}
	pins, err := releaseset.ScanOverlays(root)
	if err != nil {
		return exitCannotRun, err
	}
	ours := releaseset.OurPins(pins, registry)
	if len(ours) == 0 {
		return exitCannotRun, fmt.Errorf(
			"no production overlay pins a %s/ image — either the tree moved or the registry "+
				"prefix is wrong; refusing to report a clean release from an empty measurement", registry)
	}

	violations := releaseset.Check(decl, pins, registry)
	if len(violations) == 0 {
		if !quiet {
			fmt.Printf("%sRelease coverage OK%s: %d of %d production overlays pin a %s/ image, "+
				"and every one is in a release path or explicitly exempt.\n",
				green, reset, len(ours), len(pins), registry)
			printExemptions(decl)
			fmt.Printf("%s  Not covered by this check, by construction: services with no overlays/ tree "+
				"(applied by hand), overlays never applied to the cluster, and whether the tag an image "+
				"carries matches its contents (that is BLD-019/BLD-020).%s\n", dim, reset)
		}
		return exitOK, nil
	}

	fmt.Fprintf(os.Stderr, "%sRELEASE COVERAGE: %d problem(s) (bugs_open/318, bugs_open/237)%s\n",
		red, len(violations), reset)
	for _, v := range violations {
		fmt.Fprintf(os.Stderr, "%s  %s%s\n", red, v.String(), reset)
		fmt.Fprintf(os.Stderr, "%s      fix: %s%s\n", yellow, v.Remedy, reset)
	}
	fmt.Fprintf(os.Stderr, "%s  (%d of %d production overlays pin a %s/ image and were judged.)%s\n",
		dim, len(ours), len(pins), registry, reset)
	return exitViolation, nil
}
