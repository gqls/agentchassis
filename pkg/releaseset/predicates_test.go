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
	for _, missing := range []string{"RELEASE_IMAGES", "AGENT_DEPLOY_SERVICES", "RETAG_EXEMPT"} {
		src := goodMakefile
		src = strings.Replace(src, missing+" :=", "SOMETHING_ELSE"+" :=", 1)
		if _, err := ParseMakefileDecls(strings.NewReader(src)); err == nil {
			t.Fatalf("T8: a makefile with no %s parsed clean — the gate would report on a shape it never read", missing)
		}
	}
	// OWN_LINEAGE is the one that may legitimately be absent: it is the list
	// that CLEARS services, it is empty today, and absent must mean the same as
	// empty or the first exemption anyone adds changes the meaning of every
	// prior run.
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

// newName must win over name — kustomize's semantics, and the shell gate read
// the wrong one. Latent on this estate (one placeholder overlay uses newName),
// so this pins a correctness fix rather than recording a live defect.
func TestFirstImage_NewNameWins(t *testing.T) {
	body := "images:\n  - name: PLACEHOLDER\n    newName: " + registry + "/real-thing\n    newTag: v1\n"
	img, tag, ok := firstImage(body)
	if !ok || img != registry+"/real-thing" || tag != "v1" {
		t.Fatalf("newName did not win: img=%q tag=%q ok=%v", img, tag, ok)
	}
	// And a second element must not overwrite the first.
	two := "images:\n  - name: " + registry + "/first\n    newTag: v1\n  - name: " + registry + "/second\n    newTag: v2\n"
	img, tag, _ = firstImage(two)
	if img != registry+"/first" || tag != "v1" {
		t.Fatalf("second images element leaked into the read: img=%q tag=%q", img, tag)
	}
	// A following top-level block must not leak in either.
	trailing := imagesBlock(registry+"/x", "v9") + "\nreplicas:\n  - name: x\n    count: 2\n"
	img, tag, _ = firstImage(trailing)
	if img != registry+"/x" || tag != "v9" {
		t.Fatalf("a later top-level block leaked into the read: img=%q tag=%q", img, tag)
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
