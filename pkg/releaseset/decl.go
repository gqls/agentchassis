// FILE: pkg/releaseset/decl.go
//
// Package releaseset answers one question the makefile could only ask about
// itself: is every service of ours actually reachable by a release?
//
// WHY IT IS A PACKAGE AND NOT MORE SHELL (bugs_open/318). The gate this
// replaces, `check-release-coverage` (register BLD-022), skipped any overlay
// whose pinned image was not already in RELEASE_IMAGES. Membership of that list
// was therefore the gate's OWN admission criterion, so a service left out at
// birth was not "uncovered" — it was out of scope, and the gate printed
// "Release coverage OK" about the exact omission it existed to catch. Eight
// services fell into that hole; two of them on 2026-08-21 and 2026-08-22, by
// sessions that had the closing owner ruling in front of them. The makefile's
// remedy was a comment in capitals, and CLAUDE.md's owner ruling of 2026-08-02
// §2 already settles what that is worth: "a comment is not a control on a tree
// this many sessions share."
//
// The second reason is that the old gate could not be proven safely. On
// 2026-08-22 a session mutated the shared makefile IN PLACE to show the gate
// discriminates, and another session committed the file inside the window
// (WRONG_CALLS.md, f016b07ec). Every predicate here is a pure function over
// parsed data, so the mutation proofs are table rows and testdata fixtures and
// no shared file is ever edited to run them.
//
// THIS IS NOT A MAKE EVALUATOR, deliberately. It extracts four literal
// `NAME := a b \` continuation blocks — the cron_parity_test.go idiom, one
// language over — and decl_parity_test.go pins the reading against
// `make -s print-release-images` so the two cannot drift. A block that cannot
// be found is an ERROR, never an empty set: a gate that passes what it failed
// to measure is this estate's own blind-pass landmine.
package releaseset

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Entry is one `<service>[:<image>]` element of AGENT_DEPLOY_SERVICES, or one
// `<service>:<make-target>` element of RETAG_EXEMPT / OWN_LINEAGE. The estate
// writes a service that runs ANOTHER service's binary in this form — visibly,
// in one place — rather than as a buried special case, which is what let
// `render-audit-adapter:browser-runner-adapter` and the two `agent-chassis`
// consumers stop being latent (bugs_open/237).
type Entry struct {
	Service string
	// Qualifier is whatever followed the colon: an image name for
	// AGENT_DEPLOY_SERVICES, a make target for RETAG_EXEMPT and OWN_LINEAGE.
	// Empty when the entry was bare.
	Qualifier string
	// Raw is the element exactly as declared, for error messages that a reader
	// can grep the makefile for.
	Raw string
}

// Image resolves an AGENT_DEPLOY_SERVICES entry's image: the qualifier when
// given, else the service name. Meaningless for the other two lists.
func (e Entry) Image() string {
	if e.Qualifier != "" {
		return e.Qualifier
	}
	return e.Service
}

// Decl is the release's declared shape, read out of the makefile.
type Decl struct {
	// ReleaseImages are the images build-backend/push-backend produce at
	// $(IMAGE_TAG). Since 95757b6c2 build-backend is DERIVED from this list, so
	// "the release builds exactly what it ships" is true by construction and is
	// not re-asserted here.
	ReleaseImages []string
	// AgentDeploy is what deploy-agents retags and applies.
	AgentDeploy []Entry
	// RetagExempt are overlays that pin a release-built image but are retagged
	// by their OWN deploy path, each naming the target that does it.
	RetagExempt []Entry
	// OwnLineage is the opt-in exemption from the birth-admission rule
	// (UncoveredOverlays): a service whose image the release deliberately does
	// NOT build, naming the target that moves it instead.
	//
	// ⚠ THE UNSAFE SIDE IS THE DEFAULT OFF, on purpose. An absent entry means
	// "this service must be in the release", which is the safe reading; opting
	// out is an explicit, greppable declaration a reviewer of the OVERLAY can
	// see. That is CLAUDE.md's owner ruling of 2026-08-02 §2 applied literally:
	// new authority on a shared seam ships as an opt-in field with the unsafe
	// default OFF, because a comment is not a control here. It is EMPTY today,
	// and it must stay a list rather than becoming a predicate — a rule that
	// guesses which services are legitimately outside the release is a rule
	// nobody can review.
	OwnLineage []Entry
}

// HasReleaseImage reports whether img (bare name, no registry prefix) is built
// and pushed by a release.
func (d Decl) HasReleaseImage(img string) bool {
	for _, r := range d.ReleaseImages {
		if r == img {
			return true
		}
	}
	return false
}

// lookup finds a service in one of the entry lists.
func lookup(entries []Entry, service string) (Entry, bool) {
	for _, e := range entries {
		if e.Service == service {
			return e, true
		}
	}
	return Entry{}, false
}

// InAnyReleasePath reports whether the service is retagged by deploy-agents or
// by its own named target.
func (d Decl) InAnyReleasePath(service string) bool {
	if _, ok := lookup(d.AgentDeploy, service); ok {
		return true
	}
	_, ok := lookup(d.RetagExempt, service)
	return ok
}

// declBlocks are the four variables read, in the form they are declared.
var declBlocks = []string{
	"RELEASE_IMAGES",
	"AGENT_DEPLOY_SERVICES",
	"RETAG_EXEMPT",
	"OWN_LINEAGE",
}

// ParseMakefileDecls reads the four declarations out of a makefile.
//
// RELEASE_IMAGES, AGENT_DEPLOY_SERVICES and RETAG_EXEMPT are REQUIRED: a
// missing one is an error, not an empty set. That asymmetry is the whole point
// — an empty RELEASE_IMAGES would make UncoveredOverlays flag every overlay on
// disk (loud, self-correcting), but an empty AGENT_DEPLOY_SERVICES would make
// InAnyReleasePath answer false for everything, and an empty read of a list the
// gate uses to CLEAR a service is how a check passes a tree it never measured.
// OWN_LINEAGE is optional precisely because it is the clearing list and it is
// legitimately empty today; absent and empty must mean the same thing there, or
// the first exemption anyone adds changes the meaning of every prior run.
func ParseMakefileDecls(r io.Reader) (Decl, error) {
	raw := map[string][]string{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var current string
	for sc.Scan() {
		line := sc.Text()
		if current == "" {
			name, rest, ok := declStart(line)
			if !ok {
				continue
			}
			current = name
			raw[current] = append(raw[current], fields(rest)...)
			if !strings.HasSuffix(strings.TrimRight(line, " \t"), `\`) {
				current = ""
			}
			continue
		}
		// Inside a continuation. A comment line cannot appear here — make would
		// not accept one — so every line is either more values or the last.
		raw[current] = append(raw[current], fields(line)...)
		if !strings.HasSuffix(strings.TrimRight(line, " \t"), `\`) {
			current = ""
		}
	}
	if err := sc.Err(); err != nil {
		return Decl{}, fmt.Errorf("reading makefile: %w", err)
	}

	var d Decl
	for _, name := range declBlocks[:3] {
		if _, ok := raw[name]; !ok {
			return Decl{}, fmt.Errorf(
				"%s is not declared in the makefile — refusing to report on a release shape "+
					"that was never read (a check that passes what it failed to measure is worse "+
					"than no check; bugs_open/318)", name)
		}
	}
	d.ReleaseImages = raw["RELEASE_IMAGES"]
	d.AgentDeploy = toEntries(raw["AGENT_DEPLOY_SERVICES"])
	d.RetagExempt = toEntries(raw["RETAG_EXEMPT"])
	d.OwnLineage = toEntries(raw["OWN_LINEAGE"]) // optional; nil when absent

	sort.Strings(d.ReleaseImages)
	return d, nil
}

// declStart recognises `NAME := ...` / `NAME = ...` / `NAME ?= ...` for one of
// the four blocks and returns the value part.
func declStart(line string) (name, rest string, ok bool) {
	for _, n := range declBlocks {
		if !strings.HasPrefix(line, n) {
			continue
		}
		after := strings.TrimLeft(line[len(n):], " \t")
		for _, op := range []string{":=", "?=", "="} {
			if strings.HasPrefix(after, op) {
				return n, after[len(op):], true
			}
		}
	}
	return "", "", false
}

// fields splits a declaration line into values, dropping the trailing
// backslash. A `#` starts a comment in make, so anything from it is not a value.
func fields(s string) []string {
	if i := strings.Index(s, "#"); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimRight(strings.TrimRight(s, " \t"), `\`)
	var out []string
	for _, f := range strings.Fields(s) {
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

func toEntries(vals []string) []Entry {
	var out []Entry
	for _, v := range vals {
		svc, qual, _ := strings.Cut(v, ":")
		out = append(out, Entry{Service: svc, Qualifier: qual, Raw: v})
	}
	return out
}
