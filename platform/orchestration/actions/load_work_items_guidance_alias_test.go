// FILE: platform/orchestration/actions/load_work_items_guidance_alias_test.go
//
// bugs_open/271 — spec.content_guidance was written by four emitters and read
// by nothing, so every rewrite brief carried by that key was silently ignored.
//
// The channel that DOES work has a different spelling, and it is the fleet
// convention. Verified end to end against live agent_definitions 2026-08-15:
//
//	site_work_items.spec
//	  → LoadWorkItemsAction parses it to a map (this file's subject)
//	  → build-dispatch-loop / site-work-orchestrator: "spec": "current_item.spec"
//	  → page-build-handler:  "rewrite_guidance?": "input_data.spec.suggestion"
//	  → page-content-writer: {{if .rewrite_guidance}}## Rewrite Guidance (IMPORTANT…)
//
// Census the same day: 56 of 231 content_rewrite rows and 34 of 113
// needs_content_page rows carried a brief ONLY in the dead spelling; no row
// carried both keys; every other item type uses suggestion exclusively.
//
// aliasGuidanceIntoSuggestion closes that at the one point every dispatched
// item passes through. The tests below are each chosen to FAIL against a
// specific wrong implementation rather than merely pass against the right one —
// a test that would pass with the rule deleted proves nothing (the
// a-quiet-test-passes-when-the-rule-is-gone trap). The mutation that kills each
// one is named in its comment.
package actions

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// TestAliasGuidance_FillsWhenOnlyGuidancePresent is the fix itself: the 90-row
// population, and every future guidance-only producer.
//
// Killed by: deleting the alias body.
func TestAliasGuidance_FillsWhenOnlyGuidancePresent(t *testing.T) {
	spec := map[string]interface{}{
		"page_name":        "services",
		"content_guidance": "State the six service names in the FAQ answer.",
	}
	aliasGuidanceIntoSuggestion(spec)

	if got := spec["suggestion"]; got != "State the six service names in the FAQ answer." {
		t.Fatalf("suggestion = %#v, want the guidance text — the brief still reaches no prompt", got)
	}
	if got := spec["content_guidance"]; got != "State the six service names in the FAQ answer." {
		t.Errorf("content_guidance must be left intact for any existing reader: %#v", got)
	}
}

// TestAliasGuidance_NeverOverwritesAnAuthorsSuggestion pins precedence. No live
// row carries both keys today, so this is prospective armour — and it is the
// property that makes the alias safe to run over every item unconditionally.
//
// Killed by: an unconditional copy (writing suggestion before checking it).
func TestAliasGuidance_NeverOverwritesAnAuthorsSuggestion(t *testing.T) {
	spec := map[string]interface{}{
		"suggestion":       "the author's own brief",
		"content_guidance": "the other spelling",
	}
	aliasGuidanceIntoSuggestion(spec)

	if got := spec["suggestion"]; got != "the author's own brief" {
		t.Fatalf("suggestion = %#v — the alias overwrote a real value", got)
	}
}

// TestAliasGuidance_NeverMaterialisesEmpty locks the absent-vs-empty rule that
// the optional "rewrite_guidance?" mapping depends on: a MISSING path is
// skipped, a path that RESOLVES to empty is forwarded as "". Materialising ""
// would turn "not supplied" into "supplied as empty" for every item on the
// fleet — the same hazard setRoutingField's own note describes.
//
// Killed by: dropping the non-empty guidance check.
func TestAliasGuidance_NeverMaterialisesEmpty(t *testing.T) {
	cases := []struct {
		name string
		spec map[string]interface{}
	}{
		{"no keys at all", map[string]interface{}{"page_name": "x"}},
		{"guidance empty", map[string]interface{}{"content_guidance": ""}},
		{"suggestion empty, no guidance", map[string]interface{}{"suggestion": ""}},
		{"suggestion empty, guidance empty", map[string]interface{}{"suggestion": "", "content_guidance": ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, wasPresent := tc.spec["suggestion"]
			aliasGuidanceIntoSuggestion(tc.spec)
			got, isPresent := tc.spec["suggestion"]

			if !wasPresent && isPresent {
				t.Fatalf("suggestion appeared as %#v where it had been ABSENT — "+
					"an optional mapping now resolves to empty instead of being skipped", got)
			}
			if isPresent && got != "" {
				t.Fatalf("suggestion = %#v, want it left exactly as found", got)
			}
		})
	}
}

// TestAliasGuidance_WritesAtMostOneKey is the fence around the narrowed
// invariant. The header of load_work_item_actions.go now says the spec map is
// mutated for exactly one prose key; this asserts that claim mechanically, so a
// later "while we're here" normalisation cannot quietly widen it.
//
// Killed by: any additional write inside the alias.
func TestAliasGuidance_WritesAtMostOneKey(t *testing.T) {
	spec := map[string]interface{}{
		"page_name":        "services",
		"mode":             "edit_live",
		"source":           "content-gap-planner",
		"content_guidance": "brief",
	}
	before := make(map[string]interface{}, len(spec))
	for k, v := range spec {
		before[k] = v
	}

	aliasGuidanceIntoSuggestion(spec)

	if len(spec) != len(before)+1 {
		t.Fatalf("spec grew from %d to %d keys — the alias must add exactly one", len(before), len(spec))
	}
	for k, v := range before {
		if spec[k] != v {
			t.Errorf("pre-existing key %q changed from %#v to %#v", k, v, spec[k])
		}
	}
	if _, ok := spec["suggestion"]; !ok {
		t.Errorf("the one added key must be \"suggestion\": %#v", spec)
	}
}

// TestAliasGuidance_RefusesNonStrings. spec is caller-supplied and can carry
// anything; a non-string under either key is a shape this function will not
// judge, and a naive assertion would either panic or write a nonsense value
// into a prompt.
//
// Killed by: asserting the type without the comma-ok, or ignoring the type of
// an existing suggestion.
func TestAliasGuidance_RefusesNonStrings(t *testing.T) {
	spec := map[string]interface{}{"content_guidance": 42}
	aliasGuidanceIntoSuggestion(spec)
	if got, present := spec["suggestion"]; present {
		t.Errorf("non-string guidance produced suggestion = %#v; want no write", got)
	}

	structured := map[string]interface{}{
		"suggestion":       map[string]interface{}{"text": "structured"},
		"content_guidance": "flat text",
	}
	aliasGuidanceIntoSuggestion(structured)
	if _, isString := structured["suggestion"].(string); isString {
		t.Errorf("a non-string suggestion was replaced by the guidance: %#v", structured["suggestion"])
	}
}

// TestLoadWorkItems_AliasesGuidanceOnTheLoadPath is the end-to-end regression,
// and the only test here that fails if the CALL SITE is removed while the
// helper above stays perfect. A helper with no callers looks exactly like a
// finished refactor, so the unit tests cannot be the whole guard.
//
// Shape follows TestLoadWorkItems_ExposesRoutingColumns: the real SELECT, the
// real Scan, three rows covering the three populations.
func TestLoadWorkItems_AliasesGuidanceOnTheLoadPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	cols := []string{
		"id", "site_id", "source", "pipeline", "item_type",
		"severity", "summary", "spec", "page_id",
		"priority", "handler_agent", "status", "item_key",
		"batch_id", "attempt_count", "approval_mode",
		"component_id", "entity_id", "affected_url",
	}

	rows := sqlmock.NewRows(cols).
		// The bugs_open/271 population: a brief in the dead spelling only.
		AddRow(uuid.New(), siteID, "content-gap-planner", "build", "content_rewrite",
			"medium", "add content", []byte(`{"page_name":"faq","content_guidance":"State the six service names."}`), nil,
			35, "page-build-handler", "triaged", "gap_plan_add_faq",
			nil, 0, "auto",
			nil, nil, nil).
		// The majority shape: already on the live channel, must pass through.
		AddRow(uuid.New(), siteID, "design-audit", "build", "content_rewrite",
			"medium", "tone", []byte(`{"page_name":"index","suggestion":"Warm the opening paragraph."}`), nil,
			35, "page-build-handler", "triaged", "audit_tone_index",
			nil, 0, "auto",
			nil, nil, nil).
		// Neither key: suggestion must stay ABSENT, not become "".
		AddRow(uuid.New(), siteID, "generic", "build", "needs_rerender",
			"low", "rerender", []byte(`{"page_name":"about"}`), nil,
			60, "rerender-pages", "triaged", "rerender:about",
			nil, 0, "auto",
			nil, nil, nil)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{"site_id": siteID.String()},
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id": "input_data.site_id",
		}},
	}

	out, err := LoadWorkItemsAction(context.Background(), params)
	if err != nil {
		t.Fatalf("LoadWorkItemsAction: %v", err)
	}

	items := out.(map[string]interface{})["items"].([]interface{})
	if len(items) != 3 {
		t.Fatalf("loaded %d items, want 3", len(items))
	}

	specOf := func(i int) map[string]interface{} {
		item, ok := items[i].(map[string]interface{})
		if !ok {
			t.Fatalf("item %d is not a map: %#v", i, items[i])
		}
		spec, ok := item["spec"].(map[string]interface{})
		if !ok {
			t.Fatalf("item %d spec is not a map: %#v", i, item["spec"])
		}
		return spec
	}

	if got := specOf(0)["suggestion"]; got != "State the six service names." {
		t.Errorf("guidance-only item: spec.suggestion = %#v, want the guidance text — "+
			"this is the bugs_open/271 failure, and it means the call site is missing", got)
	}
	if got := specOf(1)["suggestion"]; got != "Warm the opening paragraph." {
		t.Errorf("suggestion-only item: spec.suggestion = %#v, want it untouched", got)
	}
	if got, present := specOf(2)["suggestion"]; present {
		t.Errorf("item with neither key gained spec.suggestion = %#v; it must stay absent", got)
	}
}

// deadGuidanceKeyWrite matches a spec being BUILT with the dead spelling, in
// the two forms a write takes: `"content_guidance":` in map-literal key
// position, and `spec["content_guidance"] =` as an index assignment.
//
// It deliberately does NOT match the legitimate READ at
// apply_gap_plan_action.go:178 (`addPlan["content_guidance"].(string)`), which
// takes the value off the gap PLAN — that is the content-gap-planner LLM's
// output contract and must keep working. The discriminator is what FOLLOWS the
// closing bracket: `=` is a write, `.` or `==` is a read.
var deadGuidanceKeyWrite = regexp.MustCompile(
	`("content_guidance"\s*:)|(\[\s*"content_guidance"\s*\]\s*=[^=])`)

// TestNoEmitterWritesTheDeadGuidanceKey is the write-side ratchet (shape from
// work_item_type_minting_ratchet_test.go, the bugs_open/279 precedent).
//
// COMMENTS ARE STRIPPED BEFORE MATCHING — several files now explain this bug in
// prose, and a source-scanning test that reads comments makes them load-bearing.
//
// WHAT IT CANNOT SEE, stated so the coverage is not overread: a spec built in
// workflow CONFIG (the generic create_work_item action takes its spec from step
// config) or by operator SQL. Those are exactly why the read-side alias, not
// this ratchet, is the fix that closes the door — the alias covers producers no
// source scan can reach.
func TestNoEmitterWritesTheDeadGuidanceKey(t *testing.T) {
	dirs := []string{".", "discovery_checks"}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			path := filepath.Join(dir, name)
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			for i, line := range strings.Split(string(src), "\n") {
				if idx := strings.Index(line, "//"); idx >= 0 {
					line = line[:idx]
				}
				if deadGuidanceKeyWrite.MatchString(line) {
					t.Errorf("%s:%d writes the dead guidance key into a spec:\n\t%s\n"+
						"Nothing reads spec.content_guidance (bugs_open/271). Write \"suggestion\", "+
						"which page-build-handler maps to rewrite_guidance and the writer prompt renders.",
						path, i+1, strings.TrimSpace(line))
				}
			}
		}
	}
}

// TestGuidanceRatchetPatternStillBites proves the pattern matches what it bans
// and spares what it must not. Without this, a rename could leave the ratchet
// passing forever on an empty match set — and, worse, a widened pattern could
// start convicting the gap-plan READ, whose removal would break the live
// content-gap contract.
func TestGuidanceRatchetPatternStillBites(t *testing.T) {
	for _, bad := range []string{
		`"content_guidance": contentGuidance,`,
		`"content_guidance": fmt.Sprintf("Write a guide about %s.", name),`,
		`spec["content_guidance"] = brief`,
		`spec[ "content_guidance" ] = brief`,
	} {
		if !deadGuidanceKeyWrite.MatchString(bad) {
			t.Errorf("ratchet pattern no longer matches %q — the guard is disarmed", bad)
		}
	}
	for _, good := range []string{
		`contentGuidance, _ := addPlan["content_guidance"].(string)`,
		`"suggestion": contentGuidance,`,
		`guidance, ok := spec["content_guidance"].(string)`,
		`if spec["content_guidance"] == "" {`,
	} {
		if deadGuidanceKeyWrite.MatchString(good) {
			t.Errorf("ratchet pattern false-positives on a legitimate READ %q — "+
				"convicting the gap-plan contract is how this guard would do damage", good)
		}
	}
}
