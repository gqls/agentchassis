// FILE: platform/orchestration/actions/validate_page_content_stats_test.go
//
// Severity and plumbing contract for the stat audit (validate_page_content
// check 9). The extraction and scan logic itself is covered by
// datahelpers/claims_stats_test.go; what is pinned here is how findings are
// GRADED and how the gate behaves for callers that produce no sections.
//
// The grading rule that matters: severity `error` must mean "a machine checked
// this and it failed". It must never mean "we could not check this" — which is
// why a site with an evidence_base row but no facts[] gets warnings, and why an
// unpaired figure never blocks.

package actions

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

const statsGateTestEB = `{
  "facts": [
    {"id": "agents", "claim": "agent definitions", "value": 170, "kind": "count",
     "tolerance": "exact", "context_terms": ["agent definition"],
     "source": {"sql": "SELECT 1"}, "verified_at": "2026-07-24"}
  ],
  "banned_claims": []
}`

func statsTestEB(t *testing.T) *datahelpers.EvidenceBase {
	t.Helper()
	eb, err := datahelpers.ParseEvidenceBase([]byte(statsGateTestEB))
	if err != nil || eb == nil {
		t.Fatalf("ParseEvidenceBase: %v", err)
	}
	return eb
}

func TestStatClaimsGateSeverities(t *testing.T) {
	eb := statsTestEB(t)

	// A fabricated figure on a site with real facts registered: error, so the
	// page does not deploy and the item parks at needs_human_review.
	fabricated := datahelpers.ExtractStatClaims("system-stats", map[string]interface{}{
		"stat1_value": "2,400+", "stat1_label": "Gripper Models Indexed",
	})
	got := checkStatClaims(fabricated, eb, true)
	if len(got) != 1 {
		t.Fatalf("expected 1 issue, got %d: %+v", len(got), got)
	}
	if got[0].Severity != "error" || got[0].Category != "stat_claims" {
		t.Errorf("want error/stat_claims, got %s/%s", got[0].Severity, got[0].Category)
	}
	if got[0].Location != "system-stats.stat1_value" {
		t.Errorf("location = %q, want system-stats.stat1_value", got[0].Location)
	}

	// The SAME finding on a site whose register holds no facts: warning, and
	// the description must say why nothing could be verified. This is the
	// writer_block-only shape that silently disabled the existing claims gate
	// on the three sites bug 043 had actually caught fabricating.
	got = checkStatClaims(fabricated, eb, false)
	if len(got) != 1 || got[0].Severity != "warning" {
		t.Fatalf("a site with no registered facts must warn, not error: %+v", got)
	}
	if !strings.Contains(got[0].Description, "no machine-readable facts") {
		t.Errorf("the warning must name the gap, got %q", got[0].Description)
	}

	// A registered figure passes.
	if got := checkStatClaims(datahelpers.ExtractStatClaims("system-stats", map[string]interface{}{
		"stat1_value": "170", "stat1_label": "Agent Definitions",
	}), eb, true); len(got) != 0 {
		t.Fatalf("a registered figure was flagged: %+v", got)
	}
}

func TestUnpairedStatValueIsWarningNotError(t *testing.T) {
	eb := statsTestEB(t)
	unpaired := datahelpers.ExtractStatClaims("mystery", map[string]interface{}{
		"hero_value": "9,001",
	})
	if len(unpaired) != 1 || unpaired[0].Pairing != datahelpers.StatUnpaired {
		t.Fatalf("test premise: expected one unpaired claim, got %+v", unpaired)
	}
	got := checkStatClaims(unpaired, eb, true)
	if len(got) != 1 {
		t.Fatalf("expected 1 issue, got %+v", got)
	}
	if got[0].Severity != "warning" {
		t.Errorf("an unpaired figure must never block — severity = %q", got[0].Severity)
	}
}

func TestStatUnitsGateSeverities(t *testing.T) {
	impossible := datahelpers.ExtractStatClaims("system-stats", map[string]interface{}{
		"stat1_value": "2,400+", "stat1_suffix": "%", "stat1_label": "Gripper Models Indexed",
	})
	got := checkStatUnits(impossible)
	if len(got) != 1 || got[0].Severity != "error" || got[0].Type != "stat_unit_impossible" {
		t.Fatalf("an impossible unit must be an error: %+v", got)
	}
	if got[0].Category != "stat_units" {
		t.Errorf("category = %q, want stat_units", got[0].Category)
	}

	mismatch := datahelpers.ExtractStatClaims("system-stats", map[string]interface{}{
		"stat1_value": "14,203", "stat1_suffix": "%", "stat1_label": "Takes Filed Today",
	})
	got = checkStatUnits(mismatch)
	if len(got) != 1 || got[0].Severity != "warning" || got[0].Type != "stat_unit_mismatch" {
		t.Fatalf("a mismatched unit must be a warning (the owner may intend it): %+v", got)
	}
}

func TestCollectStatClaimsFromSectionsMetadata(t *testing.T) {
	logger := zap.NewNop()
	collected := map[string]interface{}{
		"page_content": map[string]interface{}{
			"response": map[string]interface{}{
				"sections_metadata": []interface{}{
					map[string]interface{}{
						"component_function": "system-stats",
						"content_data": map[string]interface{}{
							"stat1_value": "170", "stat1_label": "Agent Definitions",
						},
					},
					map[string]interface{}{
						"component_function": "case-studies-grid",
						"content_data": map[string]interface{}{
							"card1_stat_value": "4 days", "card1_stat_label": "brief to production",
						},
					},
				},
			},
		},
	}

	claims, _ := collectStatClaims(collected, defaultSectionsMetadataField, logger)
	if len(claims) != 2 {
		t.Fatalf("expected 2 claims across the two sections, got %d: %+v", len(claims), claims)
	}
	seen := map[string]bool{}
	for _, c := range claims {
		seen[c.Component] = true
	}
	if !seen["system-stats"] || !seen["case-studies-grid"] {
		t.Errorf("component was not carried through: %+v", claims)
	}
}

// TestStatChecksNoOpWithoutSectionsMetadata is the compatibility proof for the
// gate's other callers — tool-recreation, report-builder and content-reviewer
// run validate_page_content over an HTML blob and produce no sections_metadata.
// They must see no findings and no error.
func TestStatChecksNoOpWithoutSectionsMetadata(t *testing.T) {
	logger := zap.NewNop()

	for name, collected := range map[string]map[string]interface{}{
		"absent":     {"page_content": map[string]interface{}{"response": map[string]interface{}{"page_html": "<p>hi</p>"}}},
		"empty map":  {},
		"wrong type": {"page_content": map[string]interface{}{"response": map[string]interface{}{"sections_metadata": "not-a-list"}}},
	} {
		t.Run(name, func(t *testing.T) {
			if got, _ := collectStatClaims(collected, defaultSectionsMetadataField, logger); len(got) != 0 {
				t.Fatalf("expected no claims, got %+v", got)
			}
			if got := runStatChecks(nil, nil, collected, nil, "", true, true, logger); len(got) != 0 {
				t.Fatalf("expected no issues, got %+v", got)
			}
		})
	}
}

func TestStatChecksRespectToggles(t *testing.T) {
	logger := zap.NewNop()
	collected := map[string]interface{}{
		"page_content": map[string]interface{}{
			"response": map[string]interface{}{
				"sections_metadata": []interface{}{
					map[string]interface{}{
						"component_function": "system-stats",
						"content_data": map[string]interface{}{
							"stat1_value": "2,400+", "stat1_suffix": "%", "stat1_label": "Gripper Models Indexed",
						},
					},
				},
			},
		},
	}

	// Units on, claims off, no DB: the unit lint still fires, because it needs
	// no evidence base at all.
	got := runStatChecks(nil, nil, collected, nil, "", false, true, logger)
	if len(got) != 1 || got[0].Type != "stat_unit_impossible" {
		t.Fatalf("the unit lint must run without a DB or a register: %+v", got)
	}

	// Both off: silent.
	if got := runStatChecks(nil, nil, collected, nil, "", false, false, logger); len(got) != 0 {
		t.Fatalf("both toggles off must produce nothing, got %+v", got)
	}
}

func TestStatChecksHonourConfiguredField(t *testing.T) {
	logger := zap.NewNop()
	collected := map[string]interface{}{
		"elsewhere": []interface{}{
			map[string]interface{}{
				"component_function": "system-stats",
				"content_data":       map[string]interface{}{"stat1_value": "2,400+", "stat1_suffix": "%", "stat1_label": "X"},
			},
		},
	}
	config := map[string]interface{}{"sections_metadata_field": "elsewhere"}
	if got := runStatChecks(nil, nil, collected, config, "", false, true, logger); len(got) != 1 {
		t.Fatalf("expected the configured field to be read, got %+v", got)
	}
}

// TestStatAuditUnavailableIsNotSilence answers the council's medium objection
// (2026-07-26, bug_historian): a bare nil made "checked, nothing to audit"
// indistinguishable from "could not check". On a step that DECLARES it builds
// sections, an absent sections_metadata must surface as a warning that says so.
func TestStatAuditUnavailableIsNotSilence(t *testing.T) {
	logger := zap.NewNop()
	// A page-build-shaped payload that lost its sections metadata.
	collected := map[string]interface{}{
		"page_content": map[string]interface{}{
			"response": map[string]interface{}{"page_html": "<p>hi</p>"},
		},
	}
	cfg := map[string]interface{}{"require_sections_metadata": true}

	got := runStatChecks(nil, nil, collected, cfg, "", true, true, logger)
	if len(got) != 1 {
		t.Fatalf("expected one unavailable warning, got %+v", got)
	}
	if got[0].Type != "stat_audit_unavailable" || got[0].Severity != "warning" {
		t.Errorf("want stat_audit_unavailable/warning, got %s/%s", got[0].Type, got[0].Severity)
	}
	if !strings.Contains(got[0].Description, "NOT checked") {
		t.Errorf("the warning must say the figures were not checked, got %q", got[0].Description)
	}
	// It must warn, never block — a missing input is not a content defect.
	if got[0].Severity == "error" || got[0].Severity == "blocker" {
		t.Error("an unavailable audit must not stop a deploy")
	}

	// And the other three callers, which do not declare it, stay silent.
	if got := runStatChecks(nil, nil, collected, nil, "", true, true, logger); len(got) != 0 {
		t.Fatalf("a step that does not declare sections must stay silent, got %+v", got)
	}
}
