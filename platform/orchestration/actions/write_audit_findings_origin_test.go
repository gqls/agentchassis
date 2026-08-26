// FILE: platform/orchestration/actions/write_audit_findings_origin_test.go
//
// Pins the provenance stamp (bugs_open/405, candidate 1) at both ends of its
// seam: the Go writer and migration 629's SQL door must carry the SAME literal,
// and every classification arm must carry the stamp — because the door's whole
// premise is that the measurement-vs-judgement axis exists on the row ONLY if
// this action writes it (the 391 lane measured it cannot be derived from any
// existing column).

package actions

import (
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestOriginStamp_EveryClassificationArmCarriesIt(t *testing.T) {
	siteID := uuid.New()
	pageID := uuid.New()
	pages := map[string]pageInfo{"about": {ID: pageID, Name: "about"}}
	for name, f := range map[string]auditFinding{
		"routed content arm (content_rewrite)": {Category: "content", Page: "about", Description: "x", Severity: "medium"},
		"design arm (needs_design_review)":     {Category: "colour", Page: "about", Description: "x", Severity: "low"},
		"no-handler arm (capability_gap)":      {Category: "cta", Page: "about", Description: "x", Severity: "low"},
		"unknown-category fallback":            {Category: "zzz_never_a_category", Page: "about", Description: "x", Severity: "low"},
	} {
		c := classifyFinding(f, pages, siteID, "site-review")
		if got := c.Spec["origin"]; got != workItemOriginModelOpinion {
			t.Errorf("%s: spec.origin = %v, want %q — an arm without the stamp is a row the promoter's origin door cannot see", name, got, workItemOriginModelOpinion)
		}
	}
}

func TestOriginStamp_SurvivesRecordMode(t *testing.T) {
	pageID := uuid.New()
	c := classifyFinding(auditFinding{Category: "content", Page: "about", Description: "x", Severity: "medium"},
		map[string]pageInfo{"about": {ID: pageID, Name: "about"}}, uuid.New(), "offer-analysis")
	out := recordOnlyFinding(c, "offer-analysis")
	if out.Spec["origin"] != workItemOriginModelOpinion {
		t.Fatalf("record mode dropped the origin stamp: %v", out.Spec)
	}
}

// TestOriginDoorLockstep reads migration 629 and fails if the SQL door and the
// Go stamp disagree — the Go↔SQL drift 405 §7 names as where candidate 1 breaks.
// The migration file is committed for ever, so a missing file is a FAILURE
// (vacuity guard), not a skip.
func TestOriginDoorLockstep(t *testing.T) {
	const migration = "../../../docs/agent_docs/sql_for_agents/629_promoter_origin_door_holds_model_opinions.sql"
	raw, err := os.ReadFile(migration)
	if err != nil {
		t.Fatalf("lockstep half missing: %v — if the migration was renamed, this test's path is the other copy of that fact", err)
	}
	sql := string(raw)
	// The full predicate, built FROM the Go constant so a drifted stamp cannot
	// match. It must appear exactly twice: once in the edit's replacement text
	// and once in the verify block's verbatim assertion — the two sites that
	// must agree with each other and with the Go.
	door := "(COALESCE(wi.spec->>'origin', '') <> '" + workItemOriginModelOpinion + "')"
	if n := strings.Count(sql, door); n != 2 {
		t.Fatalf("the door predicate %q appears %d times in the migration (want exactly 2: the edit and its verify) — the Go stamp and the SQL door have drifted", door, n)
	}
	if !strings.Contains(sql, "origin_ok") {
		t.Fatalf("migration lost its origin_ok column — the door is not wired to the candidates/held sets")
	}
	// The stamp key the door reads must be the key the Go writes.
	if !strings.Contains(sql, "spec->>'origin'") {
		t.Fatalf("the door no longer reads spec->>'origin' — the seam moved without this test moving")
	}
}
