// FILE: pkg/releaseset/predicates.go
//
// The predicates. Pure functions over a parsed Decl and a scanned overlay set,
// so every one of them can be mutation-proven as a table row without editing a
// makefile that forty sessions share (WRONG_CALLS.md, f016b07ec, 2026-08-22).
package releaseset

import (
	"fmt"
	"sort"
	"strings"
)

// Violation is one finding. It carries the remedy as well as the fault, which
// is the one habit worth keeping from the shell gate it replaces: a gate that
// says only "no" gets worked around.
type Violation struct {
	Kind    string // stable slug, for tests and for grep
	Service string
	Detail  string
	Remedy  string
	Source  string // the file or declaration the finding came from
}

func (v Violation) String() string {
	s := fmt.Sprintf("%s: %s — %s", v.Kind, v.Service, v.Detail)
	if v.Source != "" {
		s += fmt.Sprintf(" [%s]", v.Source)
	}
	return s
}

// Violation kinds.
const (
	// KindUnbuiltImage is bugs_closed/318's main gap: an overlay pins one of OUR
	// images and no release builds it, so it freezes at whatever tag its author
	// pushed and nothing ever notices.
	KindUnbuiltImage = "OUR IMAGE, NO RELEASE BUILDS IT"
	// KindNoReleasePath is bugs_open/237's original case, ported unchanged: the
	// release builds the image, but this overlay is in no retag path, so the
	// release moves the image and never moves this service.
	KindNoReleasePath = "NO RELEASE PATH"
	// KindDeployImageUnbuilt is the reverse direction BLD-022 §(iv) names:
	// deploy-agents would retag a service to $(IMAGE_TAG) for an image the
	// release does not build or push.
	KindDeployImageUnbuilt = "DEPLOY ENTRY POINTS AT AN UNBUILT IMAGE"
	// KindExemptionWithoutOverlay is an exemption or retag-exempt entry naming
	// a service that has no overlay — a mute with nothing under it, which ages
	// into a silently-wrong clearance.
	KindExemptionWithoutOverlay = "EXEMPTION NAMES A SERVICE WITH NO OVERLAY"
	// KindExemptionUnreasoned is an OWN_LINEAGE entry that does not name the
	// target that moves the service instead.
	KindExemptionUnreasoned = "EXEMPTION NAMES NO RETAG TARGET"
	// KindExemptionBudget is the ACCUMULATION guard: individually-reasoned
	// exemptions are fine, and a pile of them is a hiding place.
	KindExemptionBudget = "TOO MANY SERVICES EXCUSED FROM THE RELEASE"
)

// ExemptionBudget is how many OWN_LINEAGE entries may stand without a review.
//
// WHY A BUDGET AND NOT A PER-ENTRY RULE (owner decision, 2026-08-22). The owner
// ruled OUT the staleness build this bug originally proposed — "we can skip the
// 18 August staleness build" — and took this instead. The reasoning is worth
// keeping, because the two decisions are one decision:
//
// UncoveredOverlays makes an image of ours that no release builds a violation,
// and OWN_LINEAGE is the only way out. That closes the old hiding place ("not in
// RELEASE_IMAGES") and opens a new one ("excused"). Every entry is individually
// reasoned and reviewable — that is the point of the form — but nothing about
// reviewing entries one at a time notices that there are now nine of them. This
// bug's own history is the argument: EIGHT services fell into the previous
// hiding place, TWO of them within three days of the owner ruling meant to close
// it, each addition individually unremarkable.
//
// So this polices the ACCUMULATION, not the entry — the same shape as the
// optional-key budget (RFC_022, owner-set N=10, register WFA-013), which exists
// because "ten individually inert opt-in fields are a shared action nobody
// understands".
//
// N = 3 IS A JUDGEMENT AND THE OWNER MAY SET IT OTHERWISE — it is one line.
// It is lower than the optional-key budget's 10 because an exemption from the
// release is rarer and costlier: the failure it permits is a service running
// months-old code with nothing reporting it, which is this bug.
//
// ⚠ IT CANNOT FIRE TODAY — **1 entry as of 2026-08-22** (`admin-dashboard`). A
// guard with no live subject is normally this estate's own failure mode (see
// BLD-023 on why `assert_live_capability()` was deliberately NOT built: "a
// fail-closed helper with exactly ONE caller is a mechanism nobody exercises").
// What makes this one different is that it needs no caller: it runs on every
// `deploy-core` whether or not it fires, and the gate NAMES the standing
// exemptions in its green output, so the count is in front of a human on every
// release rather than waiting for a threshold.
const ExemptionBudget = 3

// ExemptionBudgetExceeded reports the accumulation guard.
func ExemptionBudgetExceeded(d Decl) []Violation {
	if len(d.OwnLineage) <= ExemptionBudget {
		return nil
	}
	names := make([]string, 0, len(d.OwnLineage))
	for _, e := range d.OwnLineage {
		names = append(names, e.Raw)
	}
	sort.Strings(names)
	return []Violation{{
		Kind:    KindExemptionBudget,
		Service: fmt.Sprintf("%d services", len(d.OwnLineage)),
		Detail: fmt.Sprintf(
			"OWN_LINEAGE now holds %d entries against a budget of %d: %s. Each one may be "+
				"individually correct; the pile is the problem — an exemption list is where the "+
				"next frozen service hides, and nothing that reviews entries one at a time notices "+
				"there are now %d of them",
			len(d.OwnLineage), ExemptionBudget, strings.Join(names, ", "), len(d.OwnLineage)),
		Remedy: "fold one back into RELEASE_IMAGES + AGENT_DEPLOY_SERVICES, or review the whole " +
			"accumulated set and raise ExemptionBudget in pkg/releaseset/predicates.go with the " +
			"review recorded — raising it silently is the failure this guard exists to prevent.",
		Source: "makefile OWN_LINEAGE",
	}}
}

// UncoveredOverlays is the birth-admission predicate (P1) and the ported
// no-release-path predicate (the old gate) in one walk, because they are two
// arms of one question: "can a release move this service?"
//
// ⚠ THE INVERSION IS THE FIX. The shell gate asked "is this overlay's image one
// the release builds?" and CONTINUED when the answer was no. That made
// membership of RELEASE_IMAGES the gate's own admission criterion, so a service
// omitted at birth was out of scope rather than in violation, and the gate
// printed OK about it — eight times, twice after the owner ruling meant to close
// it. Here, an image of OURS that no release builds is the violation, and the
// only way out is an explicit OWN_LINEAGE entry a reviewer can see.
//
// registry is the bare registry prefix, e.g. "docker.io/aqls". An overlay
// pinning anything outside it (ollama/ollama, postgres:16-alpine) is not ours
// to roll and is never a violation.
func UncoveredOverlays(d Decl, pins []Pin, registry string) []Violation {
	var out []Violation
	for _, p := range pins {
		bare := p.Bare(registry)
		if bare == "" {
			continue // not our image
		}
		if !d.HasReleaseImage(bare) {
			if e, ok := lookup(d.OwnLineage, p.Service); ok {
				if e.Qualifier == "" {
					out = append(out, Violation{
						Kind:    KindExemptionUnreasoned,
						Service: p.Service,
						Detail: fmt.Sprintf(
							"OWN_LINEAGE entry %q exempts it from the release but names no target that retags it",
							e.Raw),
						Remedy: fmt.Sprintf("write it as '%s:<the make target that retags it>'.", p.Service),
						Source: "makefile OWN_LINEAGE",
					})
				}
				continue // exempt, deliberately and visibly
			}
			out = append(out, Violation{
				Kind:    KindUnbuiltImage,
				Service: p.Service,
				Detail: fmt.Sprintf(
					"pins %s, which is one of OUR images and is in NO release path — no release builds, "+
						"pushes or retags it, so it freezes at %s for ever and nothing reports it",
					p.Image, orNone(p.Tag)),
				Remedy: fmt.Sprintf(
					"add '%s' to RELEASE_IMAGES and '%s' (or '%s:<image>') to AGENT_DEPLOY_SERVICES "+
						"in the commit that creates the service; or, if it genuinely has its own lineage, "+
						"declare '%s:<its retag target>' in OWN_LINEAGE.",
					bare, p.Service, p.Service, p.Service),
				Source: p.Path,
			})
			continue
		}
		// The image IS built by a release. The 237 question then applies.
		if !d.InAnyReleasePath(p.Service) {
			out = append(out, Violation{
				Kind:    KindNoReleasePath,
				Service: p.Service,
				Detail: fmt.Sprintf(
					"pins %s (a release-built image) but is in no release path — the release would move "+
						"the image and leave this service on %s",
					p.Image, orNone(p.Tag)),
				Remedy: fmt.Sprintf(
					"add '%s' (or '%s:<image>') to AGENT_DEPLOY_SERVICES, or exempt it as "+
						"'%s:<its-deploy-target>' in RETAG_EXEMPT.",
					p.Service, p.Service, p.Service),
				Source: p.Path,
			})
		}
	}
	return out
}

// InvalidDeployEntries is P3 — the direction the shell gate never covered, and
// which BLD-022 §(iv) records as unpoliced.
//
// It has two arms:
//
//  1. A deploy entry whose resolved image is not in RELEASE_IMAGES. deploy-agents
//     retags that overlay to $(IMAGE_TAG) unconditionally, so nothing builds or
//     pushes the image it now points at. The worked case is the pair of GitHub
//     runners: remove `github-actions-runner` from RELEASE_IMAGES while leaving
//     both runners in AGENT_DEPLOY_SERVICES and both ImagePullBackOff TOGETHER,
//     taking CI with them. Note this survived the 95757b6c2 derivation:
//     deleting a name from RELEASE_IMAGES consistently stops it being built AND
//     pushed, but the deploy loop still retags the orphaned entry.
//
//  2. A RETAG_EXEMPT or OWN_LINEAGE entry naming a service with no overlay. Both
//     lists CLEAR a service, so a stale entry is a mute with nothing under it —
//     and it will silently clear the next service that happens to take the name.
func InvalidDeployEntries(d Decl, pins []Pin) []Violation {
	haveOverlay := map[string]bool{}
	for _, p := range pins {
		haveOverlay[p.Service] = true
	}

	var out []Violation
	for _, e := range d.AgentDeploy {
		img := e.Image()
		if d.HasReleaseImage(img) {
			continue
		}
		out = append(out, Violation{
			Kind:    KindDeployImageUnbuilt,
			Service: e.Service,
			Detail: fmt.Sprintf(
				"AGENT_DEPLOY_SERVICES entry %q resolves to image %q, which is NOT in RELEASE_IMAGES — "+
					"deploy-agents would retag this overlay to $(IMAGE_TAG) for an image no release "+
					"builds or pushes, and every service sharing that image would ImagePullBackOff together",
				e.Raw, img),
			Remedy: fmt.Sprintf(
				"add '%s' to RELEASE_IMAGES (the build and the retag are ONE change), or remove '%s' "+
					"from AGENT_DEPLOY_SERVICES.", img, e.Raw),
			Source: "makefile AGENT_DEPLOY_SERVICES",
		})
	}

	for _, list := range []struct {
		name    string
		entries []Entry
	}{
		{"RETAG_EXEMPT", d.RetagExempt},
		{"OWN_LINEAGE", d.OwnLineage},
	} {
		for _, e := range list.entries {
			if haveOverlay[e.Service] {
				continue
			}
			out = append(out, Violation{
				Kind:    KindExemptionWithoutOverlay,
				Service: e.Service,
				Detail: fmt.Sprintf(
					"%s entry %q names a service with no production overlay — it clears nothing today "+
						"and will silently clear whatever service next takes that name",
					list.name, e.Raw),
				Remedy: fmt.Sprintf("remove %q from %s, or add the overlay it refers to.", e.Raw, list.name),
				Source: "makefile " + list.name,
			})
		}
	}
	return out
}

// Check runs every predicate and returns the findings, most structural first.
func Check(d Decl, pins []Pin, registry string) []Violation {
	out := append(UncoveredOverlays(d, pins, registry), InvalidDeployEntries(d, pins)...)
	out = append(out, ExemptionBudgetExceeded(d)...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Service < out[j].Service
	})
	return out
}

// OurPins is the subset of pins the gate actually judges, for the report's
// coverage line. A report that does not say HOW MANY things it looked at cannot
// be told apart from one that looked at nothing.
func OurPins(pins []Pin, registry string) []Pin {
	var out []Pin
	for _, p := range pins {
		if p.Bare(registry) != "" {
			out = append(out, p)
		}
	}
	return out
}

func orNone(tag string) string {
	if strings.TrimSpace(tag) == "" {
		return "its current tag (no newTag pinned)"
	}
	return tag
}
