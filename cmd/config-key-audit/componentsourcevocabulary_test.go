// FILE: cmd/config-key-audit/componentsourcevocabulary_test.go
//
// bugs_open/309. These tests exist to pin the two properties that make the
// at-rest audit safe rather than merely present:
//
//  1. THE BASELINE CANNOT GROW, and cannot silence anything but the exact
//     finding it was written for. A baseline that is keyed too broadly, or that
//     a session can append to, converts a live debt into a false all-clear —
//     which is the failure this estate has already recorded against
//     COMPONENT_WRITE_ALLOWED in scripts/pattern-check.py.
//  2. THE RULE IS THE BIRTH GATE'S RULE. Not a copy pinned by a parity test —
//     the same function. TestAuditRunsTheBirthGatesOwnRule is what makes that
//     assertable rather than merely intended.
//
// The exit-code paths (0/1/2) are exercised end to end by the controls recorded
// in the lane NOTES, run against the real live library; these tests pin the
// decisions those exit codes are computed from.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/actions"
)

const repoBaselinePath = "../../docs/agent_docs/docs024_key_docs_latest/" +
	"bugfix_309_unclickable_index_cards/component_source_baseline.json"

// schema builds a house-dialect input_schema with the given field->source pairs.
func schema(t *testing.T, fields map[string]string) string {
	t.Helper()
	f := map[string]map[string]string{}
	for name, src := range fields {
		f[name] = map[string]string{"type": "string", "source": src}
	}
	raw, err := json.Marshal(map[string]interface{}{"fields": f})
	if err != nil {
		t.Fatalf("marshalling schema: %v", err)
	}
	return string(raw)
}

func liveAspects() map[string]bool {
	return map[string]bool{"identity": true, "branding": true, "content": true}
}

func baselineOf(entries ...componentSourceBaselineEntry) map[string]componentSourceBaselineEntry {
	out := map[string]componentSourceBaselineEntry{}
	for _, e := range entries {
		out[baselineKey(e.ComponentID, e.Field, e.Source, e.Class)] = e
	}
	return out
}

// ─────────────────────────────────────────────────────────────────────────────
// The one-rule property
// ─────────────────────────────────────────────────────────────────────────────

// TestAuditRunsTheBirthGatesOwnRule is the whole architectural claim of this
// mode, made assertable. CLC-018 asked for an audit that CALLS
// SourceVocabularyIssues rather than re-implementing it, because a second
// predicate can only ever be pinned against the first, never made identical to
// it. If someone later "optimises" the audit by inlining its own classifier,
// this test is what notices.
func TestAuditRunsTheBirthGatesOwnRule(t *testing.T) {
	s := schema(t, map[string]string{
		"a": "site_specs.blog.post1_url", // phantom aspect
		"b": "query.featured_post",       // unregistered query
		"c": "config",                    // no dot at all
		"d": "llm",                       // fine
		"e": "query.blog_posts",          // fine
	})

	gate := actions.SourceVocabularyIssues(s, liveAspects())
	findings := componentSourceFindings(
		[]componentRow{{ID: "id", Name: "c", InputSchema: s}}, liveAspects(), nil)

	if len(gate) != len(findings) {
		t.Fatalf("audit and birth gate disagree on how many issues exist: gate %d, audit %d",
			len(gate), len(findings))
	}
	gateSet := map[string]bool{}
	for _, msg := range gate {
		gateSet[msg] = true
	}
	for _, f := range findings {
		if !gateSet[f.Message] {
			t.Errorf("audit reported an issue the birth gate does not: %q", f.Message)
		}
	}
}

// TestEveryClassIsReachable — a classifier with an unreachable arm is a
// classifier that cannot report that class, and the baseline is grouped BY
// class. Both arms of every branch must fire.
func TestEveryClassIsReachable(t *testing.T) {
	s := schema(t, map[string]string{
		"phantom":     "site_specs.blog.x",
		"unknownq":    "query.featured_post",
		"junk_prefix": "wibble.x",
		"nodot":       "config",
	})
	seen := map[string]int{}
	for _, f := range componentSourceFindings(
		[]componentRow{{ID: "i", Name: "n", InputSchema: s}}, liveAspects(), nil) {
		seen[f.Class]++
	}
	for _, class := range []string{
		actions.SourceIssuePhantomAspect,
		actions.SourceIssueUnregisteredQuery,
		actions.SourceIssuePrefixOutsideVocabulary,
	} {
		if seen[class] == 0 {
			t.Errorf("class %q never fired — it cannot be reported or baselined", class)
		}
	}
	if seen[actions.SourceIssuePrefixOutsideVocabulary] != 2 {
		t.Errorf("expected BOTH junk-prefix and no-dot to classify as %q, got %d",
			actions.SourceIssuePrefixOutsideVocabulary, seen[actions.SourceIssuePrefixOutsideVocabulary])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// The baseline cannot silence what it was not written for
// ─────────────────────────────────────────────────────────────────────────────

// TestBaselineKeyIsNarrow is the allow-list-silences-your-own-detector proof,
// and it is the reason the key is a 4-tuple rather than a component name.
//
// The mutation this defends against is a one-word edit — dropping `field` from
// baselineKey — so the test presents a component that is ALREADY baselined for
// one field and gives it a SECOND bad field. Under the correct key that is a
// new finding; under a component-keyed baseline it is silently grandfathered,
// on 32 live page instances.
func TestBaselineKeyIsNarrow(t *testing.T) {
	const id = "comp-1"
	s := schema(t, map[string]string{
		"already_known": "site_specs.blog.x",
		"newly_added":   "site_specs.pricing.x",
	})
	base := baselineOf(componentSourceBaselineEntry{
		ComponentID: id, Component: "info-card-grid", Field: "already_known",
		Source: "site_specs.blog.x", Class: actions.SourceIssuePhantomAspect,
		LiveInstancesAtBaseline: 32, Baselined: componentSourceBaselineClosedOn,
	})

	findings := componentSourceFindings(
		[]componentRow{{ID: id, Name: "info-card-grid", InputSchema: s, LiveInstances: 32}},
		liveAspects(), base)

	byField := map[string]componentSourceFinding{}
	for _, f := range findings {
		byField[f.Field] = f
	}
	if !byField["already_known"].Grandfathered {
		t.Error("the baselined field is not grandfathered — the key is too NARROW to match itself")
	}
	if byField["newly_added"].Grandfathered {
		t.Fatal("a NEW bad field on an already-baselined component was grandfathered — " +
			"the baseline key is too broad and has become an all-clear for the component")
	}
	if ung, _, _ := componentSourceReds(findings, nil); ung != 1 {
		t.Errorf("expected exactly 1 red, got %d", ung)
	}
}

// TestChangedSourceIsANewFinding — same component, same field, different dead
// source. The old entry goes stale and the new source is ungrandfathered, so
// the job is red twice over. A key omitting `source` would call this clean.
func TestChangedSourceIsANewFinding(t *testing.T) {
	const id = "comp-1"
	base := baselineOf(componentSourceBaselineEntry{
		ComponentID: id, Component: "c", Field: "f", Source: "site_specs.blog.x",
		Class: actions.SourceIssuePhantomAspect, Baselined: componentSourceBaselineClosedOn,
	})
	findings := componentSourceFindings([]componentRow{{
		ID: id, Name: "c", InputSchema: schema(t, map[string]string{"f": "site_specs.pricing.x"}),
	}}, liveAspects(), base)

	if len(findings) != 1 || findings[0].Grandfathered {
		t.Fatalf("a changed dead source was grandfathered: %+v", findings)
	}
	if stale := staleBaselineEntries(base, findings); len(stale) != 1 {
		t.Errorf("the superseded entry should read as stale, got %d", len(stale))
	}
}

// TestDormantWakingIsRed — grandfathering the dormant components is conditional
// on their staying dormant. Deploying one is a new page acquiring a known silent
// field-drop, which is new damage and not covered by the original decision.
func TestDormantWakingIsRed(t *testing.T) {
	const id = "dormant"
	base := baselineOf(componentSourceBaselineEntry{
		ComponentID: id, Component: "Pricing Tiers", Field: "f", Source: "site_specs.pricing.x",
		Class: actions.SourceIssuePhantomAspect, LiveInstancesAtBaseline: 0,
		Baselined: componentSourceBaselineClosedOn,
	})
	row := componentRow{ID: id, Name: "Pricing Tiers",
		InputSchema: schema(t, map[string]string{"f": "site_specs.pricing.x"})}

	// Still dormant: clean.
	if _, woke, _ := componentSourceReds(
		componentSourceFindings([]componentRow{row}, liveAspects(), base), nil); woke != 0 {
		t.Error("a still-dormant baselined component was reported as woken")
	}
	// Deployed: red.
	row.LiveInstances = 1
	if _, woke, _ := componentSourceReds(
		componentSourceFindings([]componentRow{row}, liveAspects(), base), nil); woke != 1 {
		t.Error("a dormant baselined component that gained a live instance was NOT red — " +
			"conditional grandfathering has become unconditional")
	}
}

// TestRepairedEntryIsStale pins the ratchet's pawl: the file must shrink as
// repairs land, or it accumulates dead entries that could later mask a
// re-offence on the same tuple.
func TestRepairedEntryIsStale(t *testing.T) {
	base := baselineOf(componentSourceBaselineEntry{
		ComponentID: "gone", Component: "c", Field: "f", Source: "site_specs.blog.x",
		Class: actions.SourceIssuePhantomAspect, Baselined: componentSourceBaselineClosedOn,
	})
	stale := staleBaselineEntries(base, nil)
	if len(stale) != 1 || stale[0].Field != "f" {
		t.Fatalf("a baseline entry matching nothing live was not reported stale: %+v", stale)
	}
	if _, _, n := componentSourceReds(nil, stale); n != 1 {
		t.Error("a stale entry did not count as a red")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// The baseline cannot grow
// ─────────────────────────────────────────────────────────────────────────────

// TestBaselineIsClosed — the loader must REFUSE an entry dated anything but the
// closure date. A warning would not do: a warning on stderr in a CronJob is a
// warning nobody reads, and the entire value of this file is that it can only
// shrink.
func TestBaselineIsClosed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "b.json")
	write := func(date string) {
		body := componentSourceBaselineFile{Entries: []componentSourceBaselineEntry{{
			ComponentID: "i", Component: "c", Field: "f", Source: "s",
			Class: actions.SourceIssuePhantomAspect, Baselined: date,
		}}}
		raw, _ := json.Marshal(body)
		if err := os.WriteFile(path, raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	write(componentSourceBaselineClosedOn)
	if _, err := loadComponentSourceBaseline(path); err != nil {
		t.Fatalf("a correctly dated baseline was refused: %v", err)
	}

	write("2026-08-23") // a session appending tomorrow
	_, err := loadComponentSourceBaseline(path)
	if err == nil {
		t.Fatal("an entry dated AFTER the closure was accepted — the baseline can grow, " +
			"and a new finding can be silenced by adding a line to it")
	}
	if !strings.Contains(err.Error(), "CLOSED") {
		t.Errorf("the refusal should say the baseline is closed, got: %v", err)
	}
}

// TestRepoBaselineMatchesItsRecordedCensus pins the shipped file itself. The
// figures are the ones measured on 2026-08-22 and written into the lane NOTES
// and bugs_open/309; if the file is edited, this is what makes the edit visible
// rather than merely committed.
//
// It is a RATCHET assertion, not an equality one on the totals: the file may
// shrink as repairs land (that is the design), so a smaller file passes and a
// LARGER one cannot — which is the direction that matters.
func TestRepoBaselineMatchesItsRecordedCensus(t *testing.T) {
	base, err := loadComponentSourceBaseline(repoBaselinePath)
	if err != nil {
		t.Fatalf("the shipped baseline does not load: %v", err)
	}
	const censusedOnBaselineDay = 69
	if len(base) > censusedOnBaselineDay {
		t.Fatalf("the baseline has GROWN: %d entries against %d censused on the closure day. "+
			"This file may only ever shrink", len(base), censusedOnBaselineDay)
	}

	classes := map[string]int{}
	components := map[string]bool{}
	for _, e := range base {
		classes[e.Class]++
		components[e.ComponentID] = true
		if e.Route == "" {
			t.Errorf("entry %s.%s has no route — every grandfathered finding must name the "+
				"file that owns its repair, or it is an excuse rather than a deferral",
				e.Component, e.Field)
		}
	}
	for class := range classes {
		switch class {
		case actions.SourceIssuePhantomAspect, actions.SourceIssueUnregisteredQuery,
			actions.SourceIssuePrefixOutsideVocabulary:
		default:
			t.Errorf("baseline carries class %q, which the rule cannot produce — "+
				"an entry keyed on a class that never fires can never go stale, so it "+
				"would sit in this file for ever", class)
		}
	}
	t.Logf("shipped baseline: %d entries, %d components, classes %v",
		len(base), len(components), classes)
}
