// FILE: pkg/releaseset/census_test.go
//
// Every case names, in its comment, what a different result would MEAN. The
// census is a DETECTOR, and this estate's standing lesson about detectors is
// that a clean first run on a healthy fleet could only have come out clean — so
// the discriminating power has to live here, in cases where it is shown able to
// speak.
package releaseset

import (
	"strings"
	"testing"
)

func wl(name, kind, image, tag string) Workload {
	return Workload{Name: name, Kind: kind, Image: image, Tag: tag}
}

// A fleet that is coherent: everything declared, everything on one tag.
func healthyFleet() []Workload {
	return []Workload{
		wl("agent-chassis", "Deployment", registry+"/agent-chassis", "v1.0.1323"),
		wl("browser-runner-adapter", "Deployment", registry+"/browser-runner-adapter", "v1.0.1323"),
		wl("render-audit-adapter", "Deployment", registry+"/browser-runner-adapter", "v1.0.1323"),
		wl("github-actions-runner", "Deployment", registry+"/github-actions-runner", "v1.0.1323"),
		wl("github-actions-runner-vmsites", "Deployment", registry+"/github-actions-runner", "v1.0.1323"),
		wl("auth-service", "Deployment", registry+"/auth-service", "v1.0.1323"),
		wl("core-manager", "Deployment", registry+"/core-manager", "v1.0.1323"),
		// Not ours: never a finding, in any direction.
		wl("bugs-open-staleness-sweep", "CronJob", "postgres:16-alpine", "16-alpine"),
		wl("ollama-adapter", "Deployment", "ollama/ollama", "latest"),
	}
}

func censusKinds(vs []Violation) []string {
	out := make([]string, 0, len(vs))
	for _, v := range vs {
		out = append(out, v.Kind+"/"+v.Service)
	}
	return out
}

// THE NEGATIVE CONTROL, and it is first on purpose: a coherent fleet must be
// silent, or the census becomes noise and noise is fatal to a detector.
//
//	Any finding here would mean the census reports on a fleet that is correct,
//	which is how a report stops being read.
func TestCensus_HealthyFleetIsSilent(t *testing.T) {
	res, err := Census(mustDecl(t, goodMakefile), healthyFleet(), registry)
	if err != nil {
		t.Fatalf("Census: %v", err)
	}
	if len(res.Findings) != 0 {
		t.Fatalf("a coherent fleet produced findings: %v", censusKinds(res.Findings))
	}
	if res.FleetTag != "v1.0.1323" {
		t.Fatalf("fleet tag misread: %q", res.FleetTag)
	}
	if res.Examined != 7 || res.Total != 9 {
		t.Fatalf("examined/total wrong: %d/%d — the report cannot be told from one that looked at nothing",
			res.Examined, res.Total)
	}
}

// C1 — THE STRAGGLER, which is the shape the original six sat in for months.
// `optional-explicit-wires-check` was live in exactly this state on 2026-08-22:
// running fine, three tags behind, nothing saying so.
//
//	Silence here would mean the census cannot see the case it was built for.
func TestCensus_Straggler(t *testing.T) {
	fleet := append(healthyFleet(),
		wl("github-actions-runner", "Deployment", registry+"/github-actions-runner", "v1.0.948"))
	// Replace, not add, so the name is not duplicated.
	fleet = fleet[:len(fleet)-1]
	for i := range fleet {
		if fleet[i].Name == "github-actions-runner" {
			fleet[i].Tag = "v1.0.948"
		}
	}
	res, err := Census(mustDecl(t, goodMakefile), fleet, registry)
	if err != nil {
		t.Fatalf("Census: %v", err)
	}
	if !hasKind(res.Findings, KindStraggler, "github-actions-runner") {
		t.Fatalf("a service four hundred tags behind was not reported: %v", censusKinds(res.Findings))
	}
	// And nothing else moved: the sibling on the SAME image at the fleet tag is fine.
	if hasKind(res.Findings, KindStraggler, "github-actions-runner-vmsites") {
		t.Fatalf("a current service sharing the straggler's image was wrongly reported")
	}
}

// C1 boundary — DURING A ROLLOUT the fleet is split, and the census must measure
// against the tag it is moving TO, not the one it is leaving. Otherwise every
// release produces a page of findings about services that are about to be fine.
//
//	Reporting the NEW half would mean the census is unusable on release day,
//	which is precisely the day someone reads it.
func TestCensus_RolloutMeasuresAgainstTheNewerTag(t *testing.T) {
	fleet := healthyFleet()
	// Even split: 4 on the new tag, 3 left on the old.
	newer := 0
	for i := range fleet {
		if bareImage(fleet[i].Image, registry) == "" {
			continue
		}
		if newer < 4 {
			fleet[i].Tag = "v1.0.1324"
			newer++
		}
	}
	res, err := Census(mustDecl(t, goodMakefile), fleet, registry)
	if err != nil {
		t.Fatalf("Census: %v", err)
	}
	if res.FleetTag != "v1.0.1324" {
		t.Fatalf("a split fleet must measure against the NEWER tag, got %q", res.FleetTag)
	}
	for _, f := range res.Findings {
		if f.Kind == KindStraggler && strings.Contains(f.Detail, "v1.0.1324 while") {
			t.Fatalf("a service already on the new tag was reported as a straggler: %s", f)
		}
	}
}

// C1, THE OTHER DIRECTION — AHEAD of the fleet, which the first live run got
// WRONG. On 2026-08-22 this census reported `commit-sha-exposure-check` and
// `content-loss-check` (both v1.0.1324, fleet on v1.0.1323) as "RUNNING AN OLD
// FLEET TAG". They were newer. A report that states the opposite of the truth is
// worse than one that stays quiet: a reader chasing a frozen service would have
// found one that was, if anything, too new, and concluded the instrument works.
//
//	Reporting this as a straggler would reproduce that false statement. Staying
//	silent would hide a hand deploy, which is how a tag comes to mean two
//	different images — the exact contamination that forced the v1.0.1325 bump.
func TestCensus_AheadOfFleetIsNotAStraggler(t *testing.T) {
	fleet := healthyFleet()
	fleet = append(fleet, wl("commit-sha-exposure-check", "CronJob",
		registry+"/commit-sha-exposure-check", "v1.0.1324"))
	withIt := strings.Replace(goodMakefile, "AGENT_DEPLOY_SERVICES := agent-chassis",
		"AGENT_DEPLOY_SERVICES := commit-sha-exposure-check agent-chassis", 1)
	withIt = strings.Replace(withIt, "RELEASE_IMAGES := auth-service",
		"RELEASE_IMAGES := commit-sha-exposure-check auth-service", 1)

	res, err := Census(mustDecl(t, withIt), fleet, registry)
	if err != nil {
		t.Fatalf("Census: %v", err)
	}
	if hasKind(res.Findings, KindStraggler, "commit-sha-exposure-check") {
		t.Fatalf("a service AHEAD of the fleet was reported as behind it — the 2026-08-22 false statement, reproduced")
	}
	if !hasKind(res.Findings, KindAheadOfFleet, "commit-sha-exposure-check") {
		t.Fatalf("a hand-deployed service ahead of the fleet was not reported at all: %v", censusKinds(res.Findings))
	}
	for _, f := range res.Findings {
		if f.Kind == KindAheadOfFleet && !strings.Contains(f.Remedy, "MUST NOT reuse this tag") {
			t.Fatalf("the ahead-of-fleet remedy must warn against reusing the tag: %q", f.Remedy)
		}
	}
}

// TAG ORDERING IS NUMERIC, NOT LEXICAL, and the boundary is real rather than
// theoretical: this estate is on v1.0.13xx, so it has already crossed
// v1.0.999 -> v1.0.1000, where a string comparison says 999 is newer.
//
//	A lexical answer here would invert every finding across that boundary — and
//	it would have done so silently, since both readings produce a confident
//	direction.
func TestCompareTags(t *testing.T) {
	cases := []struct {
		a, b string
		want int
		ok   bool
	}{
		{"v1.0.1323", "v1.0.1324", -1, true},
		{"v1.0.1324", "v1.0.1323", 1, true},
		{"v1.0.1323", "v1.0.1323", 0, true},
		{"v1.0.999", "v1.0.1000", -1, true}, // lexically WRONG the other way
		{"v1.0.1000", "v1.0.999", 1, true},  // the same boundary, other order
		{"v1.0.948", "v1.0.1126", -1, true}, // the real github-runner freeze
		{"latest", "v1.0.1323", 0, false},   // not orderable, and says so
		{"v1.0.1323", "2026-08-22", 0, false},
		{"v1.0", "v1.0.1", 0, false}, // different arity: refuse rather than pad
	}
	for _, c := range cases {
		got, ok := compareTags(c.a, c.b)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("compareTags(%q,%q) = (%d,%v), want (%d,%v)", c.a, c.b, got, ok, c.want, c.ok)
		}
	}
}

// An unorderable tag is reported as unorderable, never guessed at.
//
//	A straggler or ahead-of-fleet verdict here would be the report asserting a
//	direction it cannot know.
func TestCensus_UncomparableTagIsNotGuessed(t *testing.T) {
	fleet := healthyFleet()
	for i := range fleet {
		if fleet[i].Name == "auth-service" {
			fleet[i].Tag = "latest"
		}
	}
	res, err := Census(mustDecl(t, goodMakefile), fleet, registry)
	if err != nil {
		t.Fatalf("Census: %v", err)
	}
	if !hasKind(res.Findings, KindTagUncomparable, "auth-service") {
		t.Fatalf("an unorderable tag was not reported as such: %v", censusKinds(res.Findings))
	}
	if hasKind(res.Findings, KindStraggler, "auth-service") || hasKind(res.Findings, KindAheadOfFleet, "auth-service") {
		t.Fatalf("an unorderable tag was given a direction it cannot have")
	}
}

// C2 — DECLARED BUT NOT RUNNING. The worked case is real:
// `capped-schedule-ordering-check` had an overlay, a dockerfile, build/push/
// deploy targets and both release lists, and no CronJob at all, on 2026-08-22.
// The finding is not that it is missing — it is that NOTHING SAID SO, and the
// next release creates it without anyone deciding to.
//
//	Silence would mean the one shape no filesystem gate can see stays invisible
//	after building the thing that exists to see it.
func TestCensus_DeclaredButNotRunning(t *testing.T) {
	fleet := healthyFleet()
	var trimmed []Workload
	for _, w := range fleet {
		if w.Name != "render-audit-adapter" {
			trimmed = append(trimmed, w)
		}
	}
	res, err := Census(mustDecl(t, goodMakefile), trimmed, registry)
	if err != nil {
		t.Fatalf("Census: %v", err)
	}
	if !hasKind(res.Findings, KindDeclaredNotRunning, "render-audit-adapter") {
		t.Fatalf("a declared service with no workload was not reported: %v", censusKinds(res.Findings))
	}
	f := res.Findings[0]
	if !strings.Contains(f.Remedy, "AGENT_DEPLOY_SERVICES") {
		t.Fatalf("the remedy must name the list to edit, and which way: %q", f.Remedy)
	}
}

// C3 — RUNNING BUT NOT DECLARED, the shape with no paved road back: nothing on
// disk will ever move it, and the gate cannot see it because there is nothing on
// disk to see.
//
//	Silence would mean a service can run one of our images indefinitely with no
//	mechanism anywhere able to notice.
func TestCensus_RunningButNotDeclared(t *testing.T) {
	fleet := append(healthyFleet(),
		wl("mystery-adapter", "Deployment", registry+"/mystery-adapter", "v1.0.1323"))
	res, err := Census(mustDecl(t, goodMakefile), fleet, registry)
	if err != nil {
		t.Fatalf("Census: %v", err)
	}
	if !hasKind(res.Findings, KindRunningNotDeclared, "mystery-adapter") {
		t.Fatalf("an undeclared workload on our image was not reported: %v", censusKinds(res.Findings))
	}
	// The three clearing/declaring lists must all count as "declared", or the
	// census reports every legitimately-exempt service every single day and
	// stops being read. admin-dashboard-shaped case:
	fleet = append(fleet, wl("admin-dashboard", "Deployment", registry+"/admin-dashboard", "v1.0.1323"))
	withExempt := goodMakefile + "\nOWN_LINEAGE := admin-dashboard:deploy-dashboard\n"
	res, err = Census(mustDecl(t, withExempt), fleet, registry)
	if err != nil {
		t.Fatalf("Census: %v", err)
	}
	if hasKind(res.Findings, KindRunningNotDeclared, "admin-dashboard") {
		t.Fatalf("an OWN_LINEAGE service was reported as undeclared — the census would nag daily about a reviewed exemption")
	}
}

// THE DEMAND CONTROL. An empty read is the failure mode that matters most for a
// scheduled detector: wrong namespace, lost RBAC, or a dead fleet all produce
// zero findings, and zero findings is what "healthy" looks like.
//
//	A nil error here would mean this census can report a clean fleet having
//	measured nothing — the exact blind-pass shape the gate half refuses.
func TestCensus_EmptyReadIsAnError(t *testing.T) {
	if _, err := Census(mustDecl(t, goodMakefile), nil, registry); err == nil {
		t.Fatal("an empty workload list reported clean")
	}
	notOurs := []Workload{wl("x", "CronJob", "postgres:16-alpine", "16-alpine")}
	if _, err := Census(mustDecl(t, goodMakefile), notOurs, registry); err == nil {
		t.Fatal("a fleet with none of our images reported clean")
	}
	// And a fleet of ours that is entirely digest-pinned has no tag to measure
	// against — also an error, not a silent pass.
	digest := []Workload{wl("agent-chassis", "Deployment", registry+"/agent-chassis", "")}
	if _, err := Census(mustDecl(t, goodMakefile), digest, registry); err == nil {
		t.Fatal("an all-digest fleet reported clean with no fleet tag")
	}
}

// SplitImageTag has three shapes and the last two are the ones that go wrong
// silently: a digest pin has NO tag (fabricating one would make every
// digest-pinned service a straggler), and a registry host with a port is not a
// tag (splitting on the last colon is right, on the first is not).
func TestSplitImageTag(t *testing.T) {
	cases := []struct{ in, repo, tag string }{
		{registry + "/agent-chassis:v1.0.1323", registry + "/agent-chassis", "v1.0.1323"},
		{registry + "/agent-chassis@sha256:abc123", registry + "/agent-chassis", ""},
		{"registry.local:5000/team/svc", "registry.local:5000/team/svc", ""},
		{"registry.local:5000/team/svc:v2", "registry.local:5000/team/svc", "v2"},
		{"postgres", "postgres", ""},
	}
	for _, c := range cases {
		repo, tag := SplitImageTag(c.in)
		if repo != c.repo || tag != c.tag {
			t.Errorf("SplitImageTag(%q) = (%q, %q), want (%q, %q)", c.in, repo, tag, c.repo, c.tag)
		}
	}
}
