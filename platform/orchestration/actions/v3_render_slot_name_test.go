// FILE: platform/orchestration/actions/v3_render_slot_name_test.go
//
// bugs_open/189, the BUILD path's half (§blast-radius extension 2026-08-06).
//
// RenderComponentAction only ever named the COMPONENT — component_name and
// component_function are both the component's own identities. On a decomposed
// page the section's name is positional ("prose-0", "tool-2"), lives on the plan
// item at current_section.name, and was copied into NEITHER; extractSectionFromMap
// then forwarded only the component keys. So once bugs_open/204 gave the build
// path component_id-first resolution, a build run would come back with BOTH prose
// slots named "ported-prose" — the positional naming destroyed with no field
// anywhere still holding it, and the locked-row guard defeated by the same rename
// that bit the re-render path.
//
// slot_name_from is the opt-in that closes it: a config path to the section's own
// name, emitted as stored_slot_name when it resolves and silent when it does not,
// so a caller with no structured slot identity is unaffected.
package actions

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// renderSlotParams mirrors the live page-content-writer loop: `current_section`
// is the sectionPlanItem the loop variable holds, and `name` on it is the
// pages.sections entry — which for a decomposed page IS the stored slot name.
func renderSlotParams(componentID string, config map[string]interface{}) ActionParams {
	return ActionParams{
		DB:               nil, // replaced by the caller
		Logger:           zap.NewNop(),
		AgentType:        "page-content-writer",
		ExecutionContext: &types.ExecutionContext{Action: "execute", StepName: "render_section"},
		StepConfig:       models.Step{Config: config},
		CollectedData: map[string]interface{}{
			"render_context": map[string]interface{}{"company_name": "Loan Calculator"},
			"current_section": map[string]interface{}{
				"name":         "prose-0",
				"function":     "ported-prose",
				"component_id": componentID,
			},
			"empty_section": map[string]interface{}{"name": ""},
		},
	}
}

// renderSlotConfig is the live render_section config, minus the keys this test
// does not exercise.
func renderSlotConfig(componentID string, extra map[string]interface{}) map[string]interface{} {
	c := map[string]interface{}{
		"component_id": componentID,
		"context_from": "render_context",
	}
	for k, v := range extra {
		c[k] = v
	}
	return c
}

// runRenderSlot wires the one component read RenderComponentAction performs and
// returns its result map.
func runRenderSlot(t *testing.T, config map[string]interface{}, componentID string) map[string]interface{} {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM content_components").
		WithArgs(componentID).
		WillReturnRows(componentRowFor(componentID, "ported-prose", "<section>{{.company_name}}</section>", `{}`))

	params := renderSlotParams(componentID, config)
	params.DB = db

	out, err := RenderComponentAction(context.Background(), params)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("expected a result map, got %T", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
	return m
}

// MUTATION THAT MUST BREAK IT: delete the slot_name_from block from
// RenderComponentAction. The result then carries only the component's identities
// and the save has nothing to preserve the positional name with.
func TestRenderComponentAction_EmitsStoredSlotNameWhenSlotNameFromResolves(t *testing.T) {
	componentID := uuid.NewString()
	m := runRenderSlot(t, renderSlotConfig(componentID, map[string]interface{}{
		"slot_name_from": "current_section.name",
	}), componentID)

	if got, _ := m["stored_slot_name"].(string); got != "prose-0" {
		t.Fatalf("stored_slot_name = %q, want prose-0 (result: %v)", got, m)
	}
	// The component's own identities are unchanged beside it — the point of the
	// change is that the two facts stop sharing a field, not that one replaces
	// the other.
	if fn, _ := m["component_function"].(string); fn != "ported-prose" {
		t.Errorf("component_function = %q, want ported-prose", fn)
	}
}

// The silent cases, which are the whole of the backward-compatibility promise:
// a caller that names no path, and a caller whose path is populated with
// nothing. Emitting "" would be worse than emitting nothing — the save's
// fallthrough is keyed on the empty string, but a key that is always present
// makes "this producer knows its slot name" unaskable.
//
// MUTATION THAT MUST BREAK IT: set result["stored_slot_name"] unconditionally.
func TestRenderComponentAction_OmitsStoredSlotNameWhenUnsetOrUnresolved(t *testing.T) {
	cases := []struct {
		name  string
		extra map[string]interface{}
	}{
		{name: "config key absent — every caller today", extra: nil},
		{name: "path resolves to an empty string", extra: map[string]interface{}{"slot_name_from": "empty_section.name"}},
		{name: "path resolves to nothing at all", extra: map[string]interface{}{"slot_name_from": "no_such.path"}},
		{name: "config key present but empty", extra: map[string]interface{}{"slot_name_from": ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			componentID := uuid.NewString()
			m := runRenderSlot(t, renderSlotConfig(componentID, tc.extra), componentID)
			if v, present := m["stored_slot_name"]; present {
				t.Errorf("stored_slot_name must be absent, got %#v", v)
			}
		})
	}
}

// extractSectionFromMap is the seam between the render step and
// sections_metadata; a key it drops is a key the save never sees. Both shapes
// matter because the live content-writer loop produces the NESTED one
// (section_output), which is exactly why that fallback exists at all.
//
// MUTATION THAT MUST BREAK IT: remove either stored_slot_name forward. The
// nested half also breaks if the field is left out of the "have I got
// everything" condition — with all four legacy keys present at top level the
// loop would break before it ever looked.
func TestExtractSectionFromMap_ForwardsStoredSlotName(t *testing.T) {
	logger := zap.NewNop()

	t.Run("top-level direct render output", func(t *testing.T) {
		_, meta := extractSectionFromMap(map[string]interface{}{
			"rendered_html":      "<section>x</section>",
			"component_id":       "c-1",
			"component_name":     "Ported Prose",
			"component_function": "ported-prose",
			"content_data":       map[string]interface{}{"body": "x"},
			"stored_slot_name":   "prose-0",
		}, logger)
		if got, _ := meta["stored_slot_name"].(string); got != "prose-0" {
			t.Fatalf("stored_slot_name = %q, want prose-0 (meta: %v)", got, meta)
		}
	})

	t.Run("nested section_output, the live loop shape", func(t *testing.T) {
		_, meta := extractSectionFromMap(map[string]interface{}{
			"rendered_html": "<section>x</section>",
			"section_output": map[string]interface{}{
				"component_id":       "c-1",
				"component_name":     "Ported Prose",
				"component_function": "ported-prose",
				"content_data":       map[string]interface{}{"body": "x"},
				"stored_slot_name":   "prose-1",
			},
		}, logger)
		if got, _ := meta["stored_slot_name"].(string); got != "prose-1" {
			t.Fatalf("stored_slot_name = %q, want prose-1 (meta: %v)", got, meta)
		}
	})

	t.Run("nested, with every legacy key already at top level", func(t *testing.T) {
		// The break-condition trap: if stored_slot_name is not part of the
		// completeness test, the loop exits before reading the nested map and
		// this case silently loses the field.
		_, meta := extractSectionFromMap(map[string]interface{}{
			"rendered_html":      "<section>x</section>",
			"component_id":       "c-1",
			"component_name":     "Ported Prose",
			"component_function": "ported-prose",
			"content_data":       map[string]interface{}{"body": "x"},
			"render_section": map[string]interface{}{
				"stored_slot_name": "prose-2",
			},
		}, logger)
		if got, _ := meta["stored_slot_name"].(string); got != "prose-2" {
			t.Fatalf("stored_slot_name = %q, want prose-2 (meta: %v)", got, meta)
		}
	})

	t.Run("absent in both shapes — no empty key manufactured", func(t *testing.T) {
		_, meta := extractSectionFromMap(map[string]interface{}{
			"rendered_html":      "<section>x</section>",
			"component_function": "ported-prose",
		}, logger)
		if v, present := meta["stored_slot_name"]; present {
			t.Fatalf("stored_slot_name must not be manufactured, got %#v", v)
		}
		if fn, _ := meta["component_function"].(string); fn != "ported-prose" {
			t.Errorf("the existing keys must still be forwarded, got %v", meta)
		}
	})
}

// The lockstep the behaviour tests above cannot see: the seed and the action
// have to agree on the exact spelling of the config key, and when they do not
// NOTHING fails — the config is simply inert and the build path goes back to
// renaming positional slots. That is the "a dead config key looks like a live
// one" shape, so it gets a test rather than a comment.
//
// Reads the seed the way doc_subjects_common_test.go reads its migrations: from
// the checkout, relative to the package dir.
//
// MUTATION THAT MUST BREAK IT: rename the key on either side.
func TestRenderSlotNameConfigKeyMatchesTheSeededWorkflow(t *testing.T) {
	const key = "slot_name_from"

	// The action half: the key reaches the emit.
	componentID := uuid.NewString()
	m := runRenderSlot(t, renderSlotConfig(componentID, map[string]interface{}{
		key: "current_section.name",
	}), componentID)
	if _, present := m["stored_slot_name"]; !present {
		t.Fatalf("config key %q no longer reaches the action", key)
	}

	// The seed half: both render steps set it, at the path the loop populates.
	seed := filepath.Join("..", "..", "..", "docs", "agent_docs", "sql_for_agents",
		"023_page_content_writer_agent.sql")
	raw, err := os.ReadFile(seed)
	if err != nil {
		t.Fatalf("cannot read %s (test runs from the package dir; the checkout must include docs/): %v", seed, err)
	}
	const setting = `"` + key + `": "current_section.name"`
	if n := strings.Count(string(raw), setting); n < 2 {
		t.Fatalf("expected %s on BOTH render_section and render_from_template in %s, found %d occurrence(s)",
			setting, seed, n)
	}
}
