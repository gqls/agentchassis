package actions

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// bugs_open/044 — plan_sections carried an empty-schema component to `deferred`
// (which downstream means CARRIED unchanged, its template fix discarded) by
// PATTERN-MATCHING the component's Function name. A self-contained TOOL whose
// name happened to contain "content"/"body"/"article"/"text"/"blog" would be
// mis-deferred silently — the identical end-state as bugs_open/024, reached one
// function away by a different route.
//
// The fix keys the exemption on the SAME explicit component_level='tool' marker
// the rerender escalation guard uses (isSelfContainedSection), so the two call
// sites of the "is this emptiness legitimate?" judgement share ONE predicate
// and cannot drift. These tests exercise planSection directly, at the level the
// bug's verification section names.
//
// planSection never dereferences the resolver in the empty-schema branch (the
// resolver only drives the field-resolution loop for components that DECLARE
// fields), so a DB-less resolver is sufficient here.
func planEmptySchema(t *testing.T, function, level string) sectionPlanItem {
	t.Helper()
	comp := componentInfo{
		ID:          uuid.New().String(),
		Function:    function,
		InputSchema: map[string]interface{}{}, // empty → no declared fields
		Raw:         map[string]interface{}{"component_level": level},
	}
	resolver := newSourceResolver(uuid.New(), nil, zap.NewNop(), "")
	return planSection(context.Background(), function, comp, resolver, zap.NewNop())
}

// The exact case the bug describes and cannot reproduce live today: a future
// tool whose name trips the content name-heuristic. Before the fix it returned
// "deferred"; the explicit tool marker must now win over the name match.
func TestPlanSection_EmptySchemaTool_TrippingNameHeuristic_IsReady(t *testing.T) {
	for _, fn := range []string{
		"tool-content-planner",
		"tool-blog-outliner",
		"tool-body-copy-scorer",
		"tool-article-summariser",
		"tool-text-fitter",
	} {
		t.Run(fn, func(t *testing.T) {
			item := planEmptySchema(t, fn, "tool")
			if item.Status != "ready" {
				t.Fatalf("tool %q with empty schema = %q (reason %q), want ready — the name heuristic must not carry a tool to deferred (bugs_open/044)",
					fn, item.Status, item.Reason)
			}
		})
	}
}

// Regression guard for the benchmark tool and its 13 siblings, none of which
// trip the heuristic: their behaviour must be UNCHANGED (still ready). They now
// reach ready via the explicit-marker exemption rather than by falling through
// the name check, but the observable status is the same.
func TestPlanSection_EmptySchemaTool_NonTripping_StillReady(t *testing.T) {
	item := planEmptySchema(t, "tool-loot-table-balancer", "tool")
	if item.Status != "ready" {
		t.Fatalf("benchmark tool = %q, want ready", item.Status)
	}
}

// The protection the fix must NOT break (bugs_closed/004,005 — blanked article
// bodies): a genuine SECTION-level component with a content-class name and an
// empty schema is a broken component-creator artefact, and must STILL defer for
// regeneration. isSelfContainedSection returns false for non-tool levels, so
// the name heuristic still fires here.
func TestPlanSection_EmptySchemaSection_ContentName_StillDeferred(t *testing.T) {
	for _, fn := range []string{"article-body", "content-block", "text-body"} {
		t.Run(fn, func(t *testing.T) {
			item := planEmptySchema(t, fn, "section")
			if item.Status != "deferred" {
				t.Fatalf("section %q with empty schema = %q, want deferred — the blanked-article protection must survive",
					fn, item.Status)
			}
		})
	}
}

// A schemaless component with a decorative (non-content) name at section level
// stays LLM-generated (ready) — the pre-existing backward-compat path, retained.
func TestPlanSection_EmptySchemaSection_DecorativeName_IsReady(t *testing.T) {
	item := planEmptySchema(t, "divider-rule", "section")
	if item.Status != "ready" {
		t.Fatalf("decorative section = %q, want ready", item.Status)
	}
}
