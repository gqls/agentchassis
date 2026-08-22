// FILE: pkg/releaseset/census.go
//
// The CLUSTER half of bugs_open/318, and it asks a different question from the
// rest of this package.
//
// `UncoveredOverlays` and `InvalidDeployEntries` read the filesystem and the
// makefile: *can a release reach this service?* They are preventive — they run
// before a release and refuse it. They are also, by construction, blind to two
// shapes, and BOTH were found by hand while measuring on 2026-08-22:
//
//   - `capped-schedule-ordering-check` had an overlay, a dockerfile,
//     build/push/deploy targets and membership of both release lists, and **no
//     CronJob in the cluster at all**. Declared everywhere, running nowhere.
//   - `site-discovery-staleness-check` and `site-locale-unset-check` were
//     running as CronJobs with **no `overlays/` tree on disk**. Running with
//     nothing on disk describing them.
//
// Neither is visible to a filesystem gate, in either direction, ever. The
// filesystem and the cluster are two enumerations and **neither is a superset of
// the other** — that sentence is the whole reason this file exists.
//
// SO THIS DETECTS RATHER THAN PREVENTS, and that is a real limit rather than a
// stage on the way to something better: a cluster-side absence has no commit to
// gate. The estate's own lesson applies — *"detection works; schedule and
// dispatch do not"* — so whatever drives this must be verified at the artefact,
// not assumed from the fact that it was built.
//
// ⚠ IT MEASURES TAGS, AND A TAG IS NOT THE CODE. A service sitting on the fleet
// tag reads as CURRENT here and can still be running a stale binary: a same-tag
// rebuild serves the node's cached image, so `v1.0.1323` can mean two different
// binaries on two different nodes. This census's own AHEAD OF THE FLEET remedy
// says as much ("the next release MUST NOT reuse this tag") — the limit is that
// it can only see the tag that caused the problem, never the problem.
//
// Raised by the council's `debug_historian` seat (low, corr b0883c17) against
// the estate's standing rule: verify at the artefact, never at git or a tag.
// Stated rather than fixed, deliberately. The artefact-side answer already
// exists and belongs to other entries — BLD-019 stamps the commit into every
// image and binary, BLD-020 makes one release one revision, BLD-023 lets a
// running pod publish what it can do — and duplicating it here would give one
// question two answers that can disagree. **So a clean census means "every
// service is on the tag it should be on", NOT "every service is running the
// code it should be running."** Those are different claims and only the first
// is made.
//
// THE PREDICATES TAKE A []Workload, NOT A KUBERNETES CLIENT, on purpose. Every
// judgement here is a pure function over data, so the whole census is table-
// testable with no cluster, and `cmd/releasecheck` owns the one part that must
// talk to the API. A predicate that could only be exercised against a live
// cluster is a predicate nobody exercises.
package releaseset

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Workload is one thing running in the cluster that carries an image: a
// Deployment, a CronJob, a DaemonSet. Deliberately not the k8s type — this
// package must not import client-go, or the predicates stop being testable
// without one.
type Workload struct {
	Name  string // metadata.name
	Kind  string // Deployment | CronJob | DaemonSet, for the report
	Image string // fully qualified, as the pod spec states it
	Tag   string // the tag portion, "" when the image is digest-pinned
}

// Census kinds. Separate from the gate's Violation kinds because these are
// REPORTS, not refusals — nothing here stops a release, and conflating the two
// vocabularies would let a detector's finding read as a gate's verdict.
const (
	// KindStraggler is a service running an image of ours at a tag the rest of
	// the fleet has moved PAST. The original six frozen services were all this
	// shape, for months, with nothing saying so.
	KindStraggler = "BEHIND THE FLEET TAG"
	// KindAheadOfFleet is the OPPOSITE, and it is a different fault with a
	// different cause: a service on a tag the fleet has not reached was put
	// there BY HAND, outside a release.
	//
	// It earned its own kind on this census's first live run (2026-08-22), which
	// reported `commit-sha-exposure-check` and `content-loss-check` — both on
	// v1.0.1324 against a fleet on v1.0.1323 — as "RUNNING AN OLD FLEET TAG".
	// That was a FALSE STATEMENT in a report, which is worse than a missing one:
	// a reader chasing a frozen service would have found a service that was, if
	// anything, too new. Reported separately because the remedy is opposite —
	// a straggler needs a release, a hand-deploy needs the tag not to be reused.
	KindAheadOfFleet = "AHEAD OF THE FLEET TAG (hand-deployed)"
	// KindTagUncomparable is the honest third answer: two tags that differ and
	// cannot be ordered. Never guessed at — a fabricated ordering is how a
	// report starts asserting things it does not know.
	KindTagUncomparable = "TAG DIFFERS FROM THE FLEET AND CANNOT BE ORDERED"
	// KindDeclaredNotRunning is a service the release would retag and apply, with
	// nothing in the cluster to apply it to.
	KindDeclaredNotRunning = "DECLARED BUT NOT RUNNING"
	// KindRunningNotDeclared is a workload on one of our images that no
	// declaration accounts for — nothing on disk will ever move it.
	KindRunningNotDeclared = "RUNNING BUT NOT DECLARED"
)

// CensusResult carries the findings AND what was measured, because a report
// that does not say how much it looked at cannot be told apart from one that
// looked at nothing. Every consumer must print Examined.
type CensusResult struct {
	Findings []Violation
	// Examined is how many cluster workloads carried one of our images.
	Examined int
	// Total is how many workloads were seen at all.
	Total int
	// FleetTag is the tag the majority of our workloads are on, which is what
	// "old" is measured against.
	FleetTag string
}

// Census compares what the cluster is running against what the makefile
// declares.
//
// registry is our image prefix; anything else (ollama, postgres, busybox) is
// not ours to roll and is never a finding.
//
// ⚠ IT RETURNS AN ERROR WHEN THERE IS NOTHING TO MEASURE, rather than an empty
// clean result. Zero of our workloads means the client read the wrong namespace,
// or lost its permissions, or the fleet is down — none of which is "nothing is
// wrong". This is the same refusal the gate makes for an empty overlay scan, for
// the same reason: a check that reports clean about what it failed to measure is
// this estate's own blind-pass landmine.
func Census(d Decl, workloads []Workload, registry string) (CensusResult, error) {
	res := CensusResult{Total: len(workloads)}

	ours := make([]Workload, 0, len(workloads))
	for _, w := range workloads {
		if bareImage(w.Image, registry) != "" {
			ours = append(ours, w)
		}
	}
	res.Examined = len(ours)
	if len(ours) == 0 {
		return res, fmt.Errorf(
			"no cluster workload runs a %s/ image (%d workloads seen) — that is a broken read, "+
				"not a clean fleet: wrong namespace, lost permissions, or nothing running. "+
				"Refusing to report a healthy census from an empty measurement", registry, len(workloads))
	}

	res.FleetTag = modalTag(ours)
	if res.FleetTag == "" {
		return res, fmt.Errorf(
			"none of the %d %s/ workloads carries a readable tag — every one is digest-pinned or "+
				"malformed, so there is no fleet tag to measure staleness against", len(ours), registry)
	}

	res.Findings = append(res.Findings, stragglers(ours, res.FleetTag)...)
	res.Findings = append(res.Findings, declaredNotRunning(d, workloads)...)
	res.Findings = append(res.Findings, runningNotDeclared(d, ours, registry)...)

	sort.SliceStable(res.Findings, func(i, j int) bool {
		if res.Findings[i].Kind != res.Findings[j].Kind {
			return res.Findings[i].Kind < res.Findings[j].Kind
		}
		return res.Findings[i].Service < res.Findings[j].Service
	})
	return res, nil
}

// C1 — a service whose tag differs from the fleet's, in whichever direction.
//
// Measured against the MODAL tag rather than the makefile's `IMAGE_TAG`, and the
// difference matters: `IMAGE_TAG` is what the NEXT release will use, so between
// a bump and the release every service would read as stale against it. The modal
// tag is what the fleet is actually on, which is the question.
//
// ⚠ THE DIRECTION IS PART OF THE FINDING, and getting it wrong is not cosmetic.
// The first live run of this census (2026-08-22) called two services on
// `v1.0.1324` "RUNNING AN OLD FLEET TAG" against a fleet on `v1.0.1323`. They
// were newer, not older. A report that states the opposite of the truth is worse
// than one that stays quiet, because a reader chasing a frozen service finds one
// that is too new and concludes the instrument works.
func stragglers(ours []Workload, fleetTag string) []Violation {
	var out []Violation
	for _, w := range ours {
		if w.Tag == "" || w.Tag == fleetTag {
			continue
		}
		switch cmp, ok := compareTags(w.Tag, fleetTag); {
		case !ok:
			out = append(out, Violation{
				Kind:    KindTagUncomparable,
				Service: w.Name,
				Detail: fmt.Sprintf(
					"%s runs %s at %s while the fleet is on %s, and the two tags cannot be ordered — "+
						"which of them is newer is not something this check will guess at",
					w.Kind, w.Image, w.Tag, fleetTag),
				Remedy: "read the image's provenance stamp (BLD-019: the OCI `revision` label, or the " +
					"binary's buildinfo.GitCommit) — that is an ancestry question, not a string one.",
				Source: "cluster",
			})
		case cmp < 0:
			out = append(out, Violation{
				Kind:    KindStraggler,
				Service: w.Name,
				Detail: fmt.Sprintf(
					"%s runs %s at %s while the fleet is on %s — nothing about a healthy pod says so, "+
						"because the pod IS healthy; it is just old",
					w.Kind, w.Image, w.Tag, fleetTag),
				Remedy: "if it is in a release path, the next release moves it — check that it did. If it " +
					"is not, that is the gate's question, not this one: run `make check-release-coverage`.",
				Source: "cluster",
			})
		default:
			out = append(out, Violation{
				Kind:    KindAheadOfFleet,
				Service: w.Name,
				Detail: fmt.Sprintf(
					"%s runs %s at %s while the fleet is on %s — nothing but a hand deploy puts a service "+
						"ahead of the fleet, so that tag now means two different things depending on when "+
						"you pulled it",
					w.Kind, w.Image, w.Tag, fleetTag),
				Remedy: "harmless in itself, but the next release MUST NOT reuse this tag — a same-tag " +
					"re-push serves the node's cached image, so the release would ship the hand-built one. " +
					"Bump IMAGE_TAG past it.",
				Source: "cluster",
			})
		}
	}
	return out
}

// compareTags orders two `vMAJOR.MINOR.PATCH` tags NUMERICALLY, and says so when
// it cannot.
//
// Numerically, not lexically, and the boundary is real rather than theoretical:
// this estate is on v1.0.13xx, so it has already crossed `v1.0.999` →
// `v1.0.1000`, where a string comparison says 999 is the newer. It returns ok
// false for anything that is not that shape — `latest`, a date tag, a digest —
// because a fabricated ordering is how a report starts asserting what it does
// not know.
func compareTags(a, b string) (int, bool) {
	pa, oka := parseVersionTag(a)
	pb, okb := parseVersionTag(b)
	if !oka || !okb || len(pa) != len(pb) {
		return 0, false
	}
	for i := range pa {
		if pa[i] != pb[i] {
			if pa[i] < pb[i] {
				return -1, true
			}
			return 1, true
		}
	}
	return 0, true
}

// parseVersionTag turns "v1.0.1323" into [1 0 1323]. Any non-numeric component,
// or no leading v, is not this shape.
func parseVersionTag(tag string) ([]int, bool) {
	if !strings.HasPrefix(tag, "v") {
		return nil, false
	}
	parts := strings.Split(strings.TrimPrefix(tag, "v"), ".")
	if len(parts) < 2 {
		return nil, false
	}
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, false
		}
		out = append(out, n)
	}
	return out, true
}

// C2 — declared, and nothing in the cluster to apply it to.
//
// Worked case: `capped-schedule-ordering-check` on 2026-08-22 — overlay,
// dockerfile, build/push/deploy targets, both release lists, and no CronJob.
// Scaffolded and never applied. Harmless in itself; the finding is that NOTHING
// SAID SO, and the next release will create it without anyone deciding to.
func declaredNotRunning(d Decl, workloads []Workload) []Violation {
	running := map[string]bool{}
	for _, w := range workloads {
		running[w.Name] = true
	}
	var out []Violation
	for _, e := range d.AgentDeploy {
		if running[e.Service] {
			continue
		}
		out = append(out, Violation{
			Kind:    KindDeclaredNotRunning,
			Service: e.Service,
			Detail: fmt.Sprintf(
				"AGENT_DEPLOY_SERVICES entry %q has no workload in the cluster — the release will "+
					"retag and apply its overlay, which CREATES it, so a service arrives without "+
					"anyone deciding to switch it on",
				e.Raw),
			Remedy: fmt.Sprintf(
				"if it is meant to run, nothing to do — the next release starts it. If it is not ready, "+
					"take %q out of AGENT_DEPLOY_SERVICES (leaving it in RELEASE_IMAGES is harmless: "+
					"the image gets built and pushed and nothing applies it).", e.Service),
			Source: "makefile AGENT_DEPLOY_SERVICES vs cluster",
		})
	}
	return out
}

// C3 — running on one of our images and accounted for by no declaration.
//
// This is the shape with no paved road back: nothing on disk will ever move it,
// and the gate cannot see it because the gate enumerates the filesystem.
func runningNotDeclared(d Decl, ours []Workload, registry string) []Violation {
	declared := map[string]bool{}
	for _, e := range d.AgentDeploy {
		declared[e.Service] = true
	}
	for _, e := range d.RetagExempt {
		declared[e.Service] = true
	}
	for _, e := range d.OwnLineage {
		declared[e.Service] = true
	}
	var out []Violation
	for _, w := range ours {
		if declared[w.Name] {
			continue
		}
		out = append(out, Violation{
			Kind:    KindRunningNotDeclared,
			Service: w.Name,
			Detail: fmt.Sprintf(
				"%s runs %s and appears in no release declaration — nothing on disk will ever move it, "+
					"and the filesystem gate cannot see it because there is nothing on disk to see",
				w.Kind, w.Image),
			Remedy: fmt.Sprintf(
				"add %q to AGENT_DEPLOY_SERVICES (with `:<image>` if it runs another service's binary), "+
					"or delete the workload if it should not be there.", w.Name),
			Source: "cluster vs makefile",
		})
	}
	return out
}

// modalTag is the tag most of our workloads are on.
//
// TIES BREAK TOWARD THE NEWER TAG, which for `v1.0.NNNN` means a fleet split
// evenly across a rollout measures staleness against the tag it is moving TO,
// not the one it is leaving. Reporting the older half as stale during a rollout
// is noise; reporting the un-rolled half is the finding.
//
// ⚠ "NEWER" IS `compareTags`, NOT `>`. This function compared tags with `tag >
// best` in its first cut — a lexical comparison, in the same file as, and one
// screen below, the numeric comparator written to fix exactly that. The
// council's editquality seat caught it (HIGH, corr b0883c17) and it was the
// gating objection: a tie between five workloads on `v1.0.999` and five on
// `v1.0.1000` picks `v1.0.999` as the fleet tag — lexically higher, numerically
// older — and that inverts EVERY straggler and ahead-of-fleet classification
// downstream, because they are all measured against this one value.
//
// That is the same defect three times in one file: once in the shipped report
// (WRONG_CALLS.md, the live run that called two hand-deploys "old"), once in the
// first repair for it, and once here. **A helper that does the comparison
// correctly does not protect the call sites that do not use it.**
//
// When two tied tags cannot be ordered at all (`latest` against a date, say),
// the fall-back is lexical and DETERMINISTIC rather than arbitrary — a stable
// choice keeps the report reproducible, and every workload whose tag differs
// from it is then reported as KindTagUncomparable, which is the honest outcome.
func modalTag(ours []Workload) string {
	counts := map[string]int{}
	for _, w := range ours {
		if w.Tag != "" {
			counts[w.Tag]++
		}
	}
	best, bestN := "", 0
	for tag, n := range counts {
		if n > bestN {
			best, bestN = tag, n
			continue
		}
		if n != bestN {
			continue
		}
		if cmp, ok := compareTags(tag, best); ok {
			if cmp > 0 {
				best = tag
			}
			continue
		}
		if tag > best { // unorderable: stable, not arbitrary
			best = tag
		}
	}
	return best
}

func bareImage(image, registry string) string {
	prefix := strings.TrimSuffix(registry, "/") + "/"
	if !strings.HasPrefix(image, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(image, prefix)
	if i := strings.LastIndex(rest, ":"); i >= 0 {
		rest = rest[:i]
	}
	return rest
}

// SplitImageTag separates a pod spec's image string into repository and tag.
// A digest-pinned image (`repo@sha256:…`) has no tag, and says so with an empty
// string rather than a fabricated one.
func SplitImageTag(image string) (repo, tag string) {
	if i := strings.Index(image, "@"); i >= 0 {
		return image[:i], ""
	}
	i := strings.LastIndex(image, ":")
	if i < 0 {
		return image, ""
	}
	// A port in a registry host (`host:5000/x`) is not a tag.
	if strings.Contains(image[i+1:], "/") {
		return image, ""
	}
	return image[:i], image[i+1:]
}
