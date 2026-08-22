// FILE: pkg/releaseset/predicates_test.go
//
// Every case below names, in its comment, WHAT A DIFFERENT RESULT WOULD MEAN.
// That is the point of the table: a green gate on a compliant tree proves
// nothing — it could only have passed — so the discriminating power of
// bugs_open/318's fix has to live here, in cases where the predicate is shown
// able to FAIL.
package releaseset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const registry = "docker.io/aqls"

// A compliant declaration, close to the live shape.
const goodMakefile = `
RELEASE_IMAGES := auth-service core-manager agent-chassis \
	browser-runner-adapter \
	github-actions-runner

AGENT_DEPLOY_SERVICES := agent-chassis \
	browser-runner-adapter render-audit-adapter:browser-runner-adapter \
	github-actions-runner github-actions-runner-vmsites:github-actions-runner

RETAG_EXEMPT := auth-service:deploy-auth-service core-manager:deploy-core-manager
`

func mustDecl(t *testing.T, src string) Decl {
	t.Helper()
	d, err := ParseMakefileDecls(strings.NewReader(src))
	if err != nil {
		t.Fatalf("ParseMakefileDecls: %v", err)
	}
	return d
}

// writeOverlay builds a throwaway deployment tree. Fixtures on disk rather than
// edits to the real makefile: on 2026-08-22 a session mutated the shared
// makefile in place to prove this very gate discriminates, and another session
// committed the file inside the window (WRONG_CALLS.md, f016b07ec).
func writeOverlay(t *testing.T, root, svc, relPath, body string) {
	t.Helper()
	full := filepath.Join(root, "deployments", "kustomize", "services", svc, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func imagesBlock(name, tag string) string {
	return "apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\n\nimages:\n  - name: " +
		name + "\n    newTag: " + tag + "\n\npatches:\n  - path: patch-deployment.yaml\n"
}

// baseline writes the overlays that goodMakefile's RETAG_EXEMPT entries refer
// to. Without them every fixture tree would (correctly) draw
// KindExemptionWithoutOverlay findings, which would drown the case under test —
// and, worse, would make a fixture that is unrealistic in exactly the way the
// predicate is designed to notice. Realistic fixtures are not decoration here.
func baseline(t *testing.T, root string) {
	t.Helper()
	writeOverlay(t, root, "auth-service", "overlays/production/uk_001/kustomization.yaml",
		imagesBlock(registry+"/auth-service", "v1.0.1323"))
	writeOverlay(t, root, "core-manager", "overlays/production/uk_001/kustomization.yaml",
		imagesBlock(registry+"/core-manager", "v1.0.1323"))
}

func kinds(vs []Violation) []string {
	out := make([]string, 0, len(vs))
	for _, v := range vs {
		out = append(out, v.Kind+"/"+v.Service)
	}
	return out
}

func hasKind(vs []Violation, kind, svc string) bool {
	for _, v := range vs {
		if v.Kind == kind && v.Service == svc {
			return true
		}
	}
	return false
}

// T1 / T2 — THE MOTIVATING CASE and its negative control.
//
// T1 reproduces the exact 2026-08-21 / 2026-08-22 shape: a new check service
// with a correct overlay, a correct dockerfile and correct build/push/deploy
// targets, born outside both lists. The shell gate this replaces printed
// "Release coverage OK" for it, because non-membership of RELEASE_IMAGES was
// its ADMISSION test rather than a violation.
//
//	A pass here would mean the rewrite carries the self-referential blindness
//	forward — i.e. the whole fix is cosmetic.
//
// T2 is the same tree with the service listed.
//
//	A violation here would mean a gate that fails a compliant tree, and this
//	estate has its own record of what happens to those: pattern-check.py's
//	header records an invariant that decayed to a comment after it annoyed
//	people, then was violated by 84% of the anchors it governed.
func TestUncoveredOverlays_BirthOmission(t *testing.T) {
	root := t.TempDir()
	baseline(t, root)
	writeOverlay(t, root, "new-check", "overlays/production/uk_001/kustomization.yaml",
		imagesBlock(registry+"/new-check", "v1.0.1300"))
	writeOverlay(t, root, "agent-chassis", "overlays/production/uk_001/kustomization.yaml",
		imagesBlock(registry+"/agent-chassis", "v1.0.1323"))

	pins, err := ScanOverlays(root)
	if err != nil {
		t.Fatalf("ScanOverlays: %v", err)
	}

	// T1: absent from both lists.
	got := Check(mustDecl(t, goodMakefile), pins, registry)
	if !hasKind(got, KindUnbuiltImage, "new-check") {
		t.Fatalf("T1: birth omission NOT caught — this is the exact shape that fell through eight times; got %v", kinds(got))
	}
	if hasKind(got, KindUnbuiltImage, "agent-chassis") {
		t.Fatalf("T1: a listed service was flagged; got %v", kinds(got))
	}

	// T2: negative control — same tree, service listed in both places.
	listed := strings.Replace(goodMakefile,
		"RELEASE_IMAGES := auth-service", "RELEASE_IMAGES := new-check auth-service", 1)
	listed = strings.Replace(listed,
		"AGENT_DEPLOY_SERVICES := agent-chassis", "AGENT_DEPLOY_SERVICES := new-check agent-chassis", 1)
	if got := Check(mustDecl(t, listed), pins, registry); len(got) != 0 {
		t.Fatalf("T2: compliant tree produced findings %v", kinds(got))
	}
}

// T3 — THE PORTED DIRECTION. render-audit-adapter reconstructed: it pins
// browser-runner-adapter's image (a release-built one) and is in no retag path.
// This is bugs_open/237's original case, which the shell gate DID catch and
// which was mutation-proven when it shipped.
//
//	A pass would mean the port silently dropped a direction the predecessor
//	already had, which is how a replacement becomes a regression.
func TestUncoveredOverlays_PortedNoReleasePath(t *testing.T) {
	root := t.TempDir()
	baseline(t, root)
	writeOverlay(t, root, "render-audit-adapter", "overlays/production/uk_001/kustomization.yaml",
		imagesBlock(registry+"/browser-runner-adapter", "v1.0.1194"))

	pins, _ := ScanOverlays(root)
	stripped := strings.Replace(goodMakefile,
		"browser-runner-adapter render-audit-adapter:browser-runner-adapter \\\n\t", "browser-runner-adapter \\\n\t", 1)
	got := Check(mustDecl(t, stripped), pins, registry)
	if !hasKind(got, KindNoReleasePath, "render-audit-adapter") {
		t.Fatalf("T3: the 237 direction was lost in the port; got %v", kinds(got))
	}
	// And with the entry restored it must go quiet — the fix that shipped.
	if got := Check(mustDecl(t, goodMakefile), pins, registry); len(got) != 0 {
		t.Fatalf("T3 control: the shipped 237 fix now reads as a violation: %v", kinds(got))
	}
}

// T4 — THE RUNNER HAZARD (P3), the direction BLD-022 §(iv) records as policed
// by nothing. Removing github-actions-runner from RELEASE_IMAGES while both
// runners stay in AGENT_DEPLOY_SERVICES points them at an image nobody builds
// or pushes, and they ImagePullBackOff TOGETHER, taking CI with them.
//
// Note this survived the 95757b6c2 build-backend derivation: deleting the name
// stops it being built AND pushed, consistently — the deploy loop is what still
// retags the orphan.
//
//	A pass would mean §(iv) is still open after a change advertised as closing it.
func TestInvalidDeployEntries_RunnerImageDropped(t *testing.T) {
	root := t.TempDir()
	baseline(t, root)
	writeOverlay(t, root, "github-actions-runner", "overlays/production/uk_001/kustomization.yaml",
		imagesBlock(registry+"/github-actions-runner", "v1.0.1323"))
	writeOverlay(t, root, "github-actions-runner-vmsites", "overlays/production/uk_001/kustomization.yaml",
		imagesBlock(registry+"/github-actions-runner", "v1.0.1323"))
	pins, _ := ScanOverlays(root)

	dropped := strings.Replace(goodMakefile, "\tgithub-actions-runner\n", "\n", 1)
	got := Check(mustDecl(t, dropped), pins, registry)
	for _, svc := range []string{"github-actions-runner", "github-actions-runner-vmsites"} {
		if !hasKind(got, KindDeployImageUnbuilt, svc) {
			t.Fatalf("T4: %s not flagged when its image left RELEASE_IMAGES; got %v", svc, kinds(got))
		}
	}
	if got := Check(mustDecl(t, goodMakefile), pins, registry); len(got) != 0 {
		t.Fatalf("T4 control: the compliant runner pair reads as a violation: %v", kinds(got))
	}
}

// T5a / T5b — THE EXEMPTION, both directions.
//
// T5a: an OWN_LINEAGE entry must actually clear the service, or nobody will use
// it and the gate gets worked around instead.
// T5b: an entry that names no retag target is a blanket mute rather than a
// declaration — the whole reason the unsafe side is opt-in is that a reviewer
// of the OVERLAY can see WHY and WHAT MOVES IT.
//
//	T5a failing would mean the escape hatch does not work; T5b passing would
//	mean the escape hatch is a silence with nothing behind it.
func TestOwnLineageExemption(t *testing.T) {
	root := t.TempDir()
	baseline(t, root)
	writeOverlay(t, root, "legacy-thing", "overlays/production/uk_001/kustomization.yaml",
		imagesBlock(registry+"/legacy-thing", "v1.0.900"))
	pins, _ := ScanOverlays(root)

	// T5a — reasoned exemption clears it.
	reasoned := goodMakefile + "\nOWN_LINEAGE := legacy-thing:release-legacy-thing\n"
	if got := Check(mustDecl(t, reasoned), pins, registry); len(got) != 0 {
		t.Fatalf("T5a: a reasoned OWN_LINEAGE entry did not clear the service: %v", kinds(got))
	}

	// T5b — bare exemption is itself a violation.
	bare := goodMakefile + "\nOWN_LINEAGE := legacy-thing\n"
	got := Check(mustDecl(t, bare), pins, registry)
	if !hasKind(got, KindExemptionUnreasoned, "legacy-thing") {
		t.Fatalf("T5b: an unreasoned exemption was accepted as a clearance: %v", kinds(got))
	}

	// T5b(ii) — an exemption naming a service with no overlay at all.
	orphan := goodMakefile + "\nOWN_LINEAGE := ghost:release-ghost\n"
	got = Check(mustDecl(t, orphan), pins, registry)
	if !hasKind(got, KindExemptionWithoutOverlay, "ghost") {
		t.Fatalf("T5b(ii): a mute with nothing under it was accepted: %v", kinds(got))
	}
}

// T6 — DEPTH. tools-api's real production overlay lives at
// overlays/production/kustomization.yaml with NO region directory, which the
// shell gate's fixed `overlays/production/uk_001/...` glob can never see.
// Harmless today (placeholder image, no workload) — but a real service at that
// depth would have been invisible for exactly the reason this bug exists.
//
//	A pass would mean the new enumeration reproduced the old glob's blind spot.
func TestScanOverlays_AnyDepth(t *testing.T) {
	root := t.TempDir()
	baseline(t, root)
	writeOverlay(t, root, "shallow-svc", "overlays/production/kustomization.yaml",
		imagesBlock(registry+"/shallow-svc", "v1.0.1000"))
	pins, err := ScanOverlays(root)
	if err != nil {
		t.Fatalf("ScanOverlays: %v", err)
	}
	seen := false
	for _, p := range pins {
		if p.Service == "shallow-svc" && p.Image == registry+"/shallow-svc" {
			seen = true
		}
	}
	if !seen {
		t.Fatalf("T6: region-less overlay not seen; pins=%+v", pins)
	}
	if got := Check(mustDecl(t, goodMakefile), pins, registry); !hasKind(got, KindUnbuiltImage, "shallow-svc") {
		t.Fatalf("T6: region-less overlay not judged; got %v", kinds(got))
	}
}

// T7 — REGISTRY PREFIX. ollama and postgres are not ours to roll. A gate that
// polices upstream images is noise, and noise is fatal here.
//
//	A violation would mean every release now fails on images nobody in this
//	estate builds.
func TestScanOverlays_UpstreamImagesIgnored(t *testing.T) {
	root := t.TempDir()
	baseline(t, root)
	writeOverlay(t, root, "ollama-adapter", "overlays/production/uk_001/kustomization.yaml",
		imagesBlock("ollama/ollama", "latest"))
	pins, _ := ScanOverlays(root)
	if got := Check(mustDecl(t, goodMakefile), pins, registry); len(got) != 0 {
		t.Fatalf("T7: upstream image produced findings %v", kinds(got))
	}
	for _, p := range OurPins(pins, registry) {
		if p.Service == "ollama-adapter" {
			t.Fatalf("T7: an upstream pin (%s) was counted as one of ours", p.Image)
		}
	}
}

// T8 — PARSER LOUDNESS. A makefile with no RELEASE_IMAGES block must be an
// ERROR, never an empty set that reads as "nothing to report".
//
// This is the estate's own blind-pass landmine one rung up: four days ago the
// bugs_open/131 lane shipped a gate that passed any page it FAILED TO MEASURE,
// inside the check written to end exactly that.
//
//	A nil error here would mean this gate can report clean on a makefile it
//	never understood.
func TestParseMakefileDecls_MissingBlockIsAnError(t *testing.T) {
	// Only the JUDGING lists. The two CLEARING lists are optional by design —
	// see TestParseMakefileDecls_ClearingListsAreOptional and the council's
	// editquality objection (medium, corr 83442a5a).
	for _, missing := range []string{"RELEASE_IMAGES", "AGENT_DEPLOY_SERVICES"} {
		src := goodMakefile
		src = strings.Replace(src, missing+" :=", "SOMETHING_ELSE"+" :=", 1)
		if _, err := ParseMakefileDecls(strings.NewReader(src)); err == nil {
			t.Fatalf("T8: a makefile with no %s parsed clean — the gate would report on a shape it never read", missing)
		}
	}
	// OWN_LINEAGE may legitimately be absent: it CLEARS services, so its absence
	// can only ever produce MORE findings, and absent must mean the same as empty
	// or the first exemption anyone adds changes the meaning of every prior run.
	d, err := ParseMakefileDecls(strings.NewReader(goodMakefile))
	if err != nil {
		t.Fatalf("T8: absent OWN_LINEAGE must be tolerated: %v", err)
	}
	if len(d.OwnLineage) != 0 {
		t.Fatalf("T8: absent OWN_LINEAGE produced entries: %+v", d.OwnLineage)
	}
}

// Parsing detail with teeth: the live declarations are TAB-indented
// continuation lines. Splitting them on spaces alone leaves a leading tab on
// six entries, which on 2026-08-22 turned a true answer of three into a
// reported nine (WRONG_CALLS.md, same day).
func TestParseMakefileDecls_TabContinuations(t *testing.T) {
	d := mustDecl(t, goodMakefile)
	for _, img := range d.ReleaseImages {
		if strings.TrimSpace(img) != img || img == "" {
			t.Fatalf("whitespace survived into a parsed image name: %q", img)
		}
	}
	if !d.HasReleaseImage("github-actions-runner") {
		t.Fatalf("a tab-indented continuation entry was lost: %v", d.ReleaseImages)
	}
	e, ok := lookup(d.AgentDeploy, "render-audit-adapter")
	if !ok || e.Image() != "browser-runner-adapter" {
		t.Fatalf("<service>:<image> form not resolved: %+v", e)
	}
	if e, ok := lookup(d.AgentDeploy, "agent-chassis"); !ok || e.Image() != "agent-chassis" {
		t.Fatalf("bare entry did not default its image: %+v", e)
	}
}

// T9 — TWO IMAGES IN ONE BLOCK, AND OURS IS THE SECOND.
//
// The council's bug_historian seat raised this (medium, corr 83442a5a) against
// the first cut, which stopped at the first element and inherited the shell
// gate's `awk … exit`: an overlay pinning app + sidecar, with OUR uncovered
// image second, would produce no Pin at all and the gate could never flag it —
// this fix reproducing, in miniature, the very shape it exists to close. The cap
// is gone rather than merely warned about.
//
// [MEASURED 2026-08-22: no kustomization anywhere under deployments/ has more
// than one element, so the gap was LATENT. A cap that is safe only because of
// today's data is the wrong kind of safe.]
//
//	Returning one Pin here would mean an uncovered image of ours is invisible
//	whenever it is not written first in the file.
func TestScanOverlays_SecondImageInBlockIsSeen(t *testing.T) {
	root := t.TempDir()
	baseline(t, root)
	body := "apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\n\nimages:\n" +
		"  - name: busybox\n    newTag: 1.36\n" +
		"  - name: " + registry + "/sidecar-check\n    newTag: v1.0.900\n\npatches:\n  - path: p.yaml\n"
	writeOverlay(t, root, "two-image-svc", "overlays/production/uk_001/kustomization.yaml", body)

	pins, err := ScanOverlays(root)
	if err != nil {
		t.Fatalf("ScanOverlays: %v", err)
	}
	var second bool
	for _, p := range pins {
		if p.Image == registry+"/sidecar-check" && p.Tag == "v1.0.900" {
			second = true
		}
	}
	if !second {
		t.Fatalf("T9: the SECOND images element was dropped; pins=%+v", pins)
	}
	if got := Check(mustDecl(t, goodMakefile), pins, registry); !hasKind(got, KindUnbuiltImage, "two-image-svc") {
		t.Fatalf("T9: an uncovered image written second was never judged; got %v", kinds(got))
	}
	// And the upstream first element must still not be judged.
	for _, v := range Check(mustDecl(t, goodMakefile), pins, registry) {
		if strings.Contains(v.Detail, "busybox") {
			t.Fatalf("T9: an upstream image in the same block was judged: %s", v)
		}
	}
}

// Both clearing lists may be absent, and absent must equal empty — otherwise the
// first entry anyone adds changes the meaning of every prior run. Raised by the
// council's editquality seat against requiring RETAG_EXEMPT.
func TestParseMakefileDecls_ClearingListsAreOptional(t *testing.T) {
	src := strings.Replace(goodMakefile, "RETAG_EXEMPT :=", "NOT_RETAG_EXEMPT :=", 1)
	d, err := ParseMakefileDecls(strings.NewReader(src))
	if err != nil {
		t.Fatalf("absent RETAG_EXEMPT must be tolerated, not a could-not-run: %v", err)
	}
	if len(d.RetagExempt) != 0 || len(d.OwnLineage) != 0 {
		t.Fatalf("absent clearing lists produced entries: %+v %+v", d.RetagExempt, d.OwnLineage)
	}
	// ...and its absence produces MORE findings, never fewer: the two services it
	// used to clear now surface. That direction is what makes optional safe.
	root := t.TempDir()
	baseline(t, root)
	pins, _ := ScanOverlays(root)
	got := Check(d, pins, registry)
	if !hasKind(got, KindNoReleasePath, "auth-service") {
		t.Fatalf("dropping RETAG_EXEMPT should EXPOSE what it cleared; got %v", kinds(got))
	}
}

// T10 — THE ACCUMULATION GUARD (owner decision 2026-08-22, taken INSTEAD of the
// staleness build). Individually-reasoned exemptions are fine; a pile of them is
// the next hiding place, and nothing that reviews entries one at a time notices
// there are now more than a handful.
//
//	Silence above the budget would mean the exemption list can grow without
//	limit — rebuilding the hole this whole change closed, with better paperwork.
//	A finding AT the budget would mean the guard fires on a state the owner
//	explicitly permitted, which is how a gate gets disabled.
func TestExemptionBudget(t *testing.T) {
	root := t.TempDir()
	baseline(t, root)
	var names []string
	for i := 0; i < ExemptionBudget+1; i++ {
		svc := "own-lineage-svc-" + string(rune('a'+i))
		names = append(names, svc+":release-"+svc)
		writeOverlay(t, root, svc, "overlays/production/uk_001/kustomization.yaml",
			imagesBlock(registry+"/"+svc, "v1.0.900"))
	}
	pins, _ := ScanOverlays(root)

	// AT the budget: silent. Each entry reasoned, none over the line.
	at := goodMakefile + "\nOWN_LINEAGE := " + strings.Join(names[:ExemptionBudget], " ") + "\n"
	atPins := pins[:0:0]
	for _, p := range pins {
		if p.Service != "own-lineage-svc-"+string(rune('a'+ExemptionBudget)) {
			atPins = append(atPins, p)
		}
	}
	if got := Check(mustDecl(t, at), atPins, registry); len(got) != 0 {
		t.Fatalf("T10: %d exemptions is AT the budget and must be silent; got %v", ExemptionBudget, kinds(got))
	}

	// ONE over: the guard speaks, and names every entry so the pile is legible.
	over := goodMakefile + "\nOWN_LINEAGE := " + strings.Join(names, " ") + "\n"
	got := Check(mustDecl(t, over), pins, registry)
	var budget *Violation
	for i := range got {
		if got[i].Kind == KindExemptionBudget {
			budget = &got[i]
		}
	}
	if budget == nil {
		t.Fatalf("T10: %d exemptions is OVER the budget of %d and was not reported; got %v",
			len(names), ExemptionBudget, kinds(got))
	}
	for _, n := range names {
		if !strings.Contains(budget.Detail, n) {
			t.Fatalf("T10: the finding must NAME the accumulated set — %q missing from %q", n, budget.Detail)
		}
	}
	if !strings.Contains(budget.Remedy, "review") {
		t.Fatalf("T10: the remedy must offer a reviewed raise, not just a refusal: %q", budget.Remedy)
	}
}

// newName must win over name — kustomize's semantics, and the shell gate read
// the wrong one. Latent on this estate (one placeholder overlay uses newName),
// so this pins a correctness fix rather than recording a live defect.
func TestFirstImage_NewNameWins(t *testing.T) {
	body := "images:\n  - name: PLACEHOLDER\n    newName: " + registry + "/real-thing\n    newTag: v1\n"
	got := blockImages(body)
	if len(got) != 1 || got[0].image != registry+"/real-thing" || got[0].tag != "v1" {
		t.Fatalf("newName did not win: %+v", got)
	}
	// Two elements: BOTH are returned, each with its own tag, in order.
	two := "images:\n  - name: " + registry + "/first\n    newTag: v1\n  - name: " + registry + "/second\n    newTag: v2\n"
	got = blockImages(two)
	if len(got) != 2 || got[0].image != registry+"/first" || got[0].tag != "v1" ||
		got[1].image != registry+"/second" || got[1].tag != "v2" {
		t.Fatalf("both images elements must be returned with their own tags: %+v", got)
	}
	// A following top-level block must not leak in.
	trailing := imagesBlock(registry+"/x", "v9") + "\nreplicas:\n  - name: x\n    count: 2\n"
	got = blockImages(trailing)
	if len(got) != 1 || got[0].image != registry+"/x" || got[0].tag != "v9" {
		t.Fatalf("a later top-level block leaked into the read: %+v", got)
	}
}

// ScanOverlays must ERROR on a missing tree rather than return zero pins: zero
// findings over zero overlays is the shape of a check that never opened the
// tree it claims to have measured.
func TestScanOverlays_MissingTreeIsAnError(t *testing.T) {
	if _, err := ScanOverlays(t.TempDir()); err == nil {
		t.Fatal("ScanOverlays returned no error for a tree with no services directory")
	}
}
