package actions

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// Covers the delivery path that bugs_open/024 found dead end-to-end: a
// tool-improver fix reached content_components.html_template and was never
// rendered onto the page, so three cycles of re-verification tested an
// unchanged page and the benchmark could never go green.
//
// Four defects in series, one test group each:
//   - the re-render request carried no reason/component_id (spec composition)
//   - its item_key was site-wide, and the anti-churn rule branded it
//     'unresolved' after two SUCCESSFUL predecessors (key scoping + opt-in)
//   - a tool section escalated instead of rendering (isSelfContainedSection)
//   - the tool's own template read as truncated (toolTemplateValid)

// ---------------------------------------------------------------------------
// create_work_item: spec composition
// ---------------------------------------------------------------------------

// specArgMatcher records argument $7 (the spec JSONB) as the insert goes past.
// sqlmock has no way to hand back an argument after the fact, so the assertion
// target is captured through a matcher that accepts anything.
type specArgMatcher struct{ got *string }

func (m specArgMatcher) Match(v driver.Value) bool {
	if s, ok := v.(string); ok {
		*m.got = s
	}
	return true
}

// runCreateWorkItem drives CreateWorkItemAction against a mock that accepts the
// insert, and returns the spec JSON the action composed.
func runCreateWorkItem(t *testing.T, config, collected map[string]interface{}) (spec string, err error) {
	t.Helper()

	db, mock, mErr := sqlmock.New()
	if mErr != nil {
		t.Fatalf("sqlmock: %v", mErr)
	}
	defer db.Close()

	mock.MatchExpectationsInOrder(false)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			specArgMatcher{got: &spec}, // $7
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	stepConfig := map[string]interface{}{
		"item_type":     "needs_rerender",
		"handler_agent": "rerender-pages",
		"source":        "tool-improver",
	}
	for k, v := range config {
		stepConfig[k] = v
	}

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData:    collected,
		StepConfig:       models.Step{Config: stepConfig},
	}

	_, err = CreateWorkItemAction(context.Background(), params)
	return spec, err
}

func siteCollected(extra map[string]interface{}) map[string]interface{} {
	c := map[string]interface{}{
		"input_data": map[string]interface{}{"site_id": uuid.New().String()},
	}
	for k, v := range extra {
		c[k] = v
	}
	return c
}

// The defect: spec_data is a PATH, so no step could stamp a constant, and the
// re-render request went out with no reason at all. page-rerender's
// check_rerender_mode gates the section re-render on exactly that value, so a
// reason-less item always took else_step render_page — "Simple concatenation -
// no template re-rendering" — and shipped the stale stored HTML as a success.
func TestCreateWorkItem_SpecLiteral_StampsConstant(t *testing.T) {
	spec, err := runCreateWorkItem(t,
		map[string]interface{}{
			"site_id":      "input_data.site_id",
			"spec_literal": map[string]interface{}{"reason": "section_data_resolved"},
		},
		siteCollected(nil),
	)
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}

	var got map[string]interface{}
	if uErr := json.Unmarshal([]byte(spec), &got); uErr != nil {
		t.Fatalf("spec is not valid JSON (%q): %v", spec, uErr)
	}
	if got["reason"] != "section_data_resolved" {
		t.Fatalf("spec.reason = %v, want section_data_resolved — a reason-less item takes render_page and deploys stale HTML", got["reason"])
	}
}

// component_id must reach the spec too: create_rerender_items sets
// scoped = (reason == section_data_resolved || image_landed) && component_id != "",
// and only a scoped run stamps the reason onto the per-page items. reason
// WITHOUT component_id silently degrades to assemble-only.
func TestCreateWorkItem_SpecPaths_ResolvesFromCollectedData(t *testing.T) {
	componentID := uuid.New().String()

	spec, err := runCreateWorkItem(t,
		map[string]interface{}{
			"site_id":      "input_data.site_id",
			"spec_literal": map[string]interface{}{"reason": "section_data_resolved"},
			"spec_paths":   map[string]interface{}{"component_id": "update_result.component_id"},
		},
		siteCollected(map[string]interface{}{
			"update_result": map[string]interface{}{"component_id": componentID},
		}),
	)
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}

	var got map[string]interface{}
	if uErr := json.Unmarshal([]byte(spec), &got); uErr != nil {
		t.Fatalf("spec is not valid JSON (%q): %v", spec, uErr)
	}
	if got["component_id"] != componentID {
		t.Fatalf("spec.component_id = %v, want %s", got["component_id"], componentID)
	}
	if got["reason"] != "section_data_resolved" {
		t.Fatalf("spec.reason = %v, want section_data_resolved", got["reason"])
	}
}

// The loud-failure rule, and the whole reason spec_paths is not a best-effort
// lookup. A spec missing component_id makes scoped=false downstream, which
// degrades the re-render to assemble-from-stale AND STILL REPORTS SUCCESS —
// bugs_open/024's signature. Failing the step is the only outcome that
// produces a signal.
func TestCreateWorkItem_SpecPaths_UnresolvedIsHardError(t *testing.T) {
	_, err := runCreateWorkItem(t,
		map[string]interface{}{
			"site_id":    "input_data.site_id",
			"spec_paths": map[string]interface{}{"component_id": "update_result.component_id"},
		},
		siteCollected(nil), // update_result absent — the write never happened
	)
	if err == nil {
		t.Fatal("expected a hard error when a spec_paths path does not resolve; a silently incomplete spec is the bug")
	}
	if !strings.Contains(err.Error(), "component_id") {
		t.Fatalf("error should name the unresolved key, got: %v", err)
	}
}

// Layering is deterministic: spec_data, then spec_paths, then spec_literal.
func TestCreateWorkItem_SpecLayering_LiteralWins(t *testing.T) {
	spec, err := runCreateWorkItem(t,
		map[string]interface{}{
			"site_id":      "input_data.site_id",
			"spec_data":    "prior_spec",
			"spec_literal": map[string]interface{}{"reason": "section_data_resolved"},
		},
		siteCollected(map[string]interface{}{
			"prior_spec": map[string]interface{}{"reason": "stale", "kept": "yes"},
		}),
	)
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}

	var got map[string]interface{}
	if uErr := json.Unmarshal([]byte(spec), &got); uErr != nil {
		t.Fatalf("spec is not valid JSON (%q): %v", spec, uErr)
	}
	if got["reason"] != "section_data_resolved" {
		t.Fatalf("spec_literal must win over spec_data: reason = %v", got["reason"])
	}
	if got["kept"] != "yes" {
		t.Fatalf("spec_data keys not overridden must survive: %v", got)
	}
}

// ---------------------------------------------------------------------------
// isSelfContainedSection — the escalation-guard exemption
// ---------------------------------------------------------------------------

// The guard escalates any section with empty content_data, and check_escalated
// routes escalated==true to complete, bypassing save_sections — the ONLY writer
// of rendered_html. A tool has content_data={} by design, so the fix was
// computed and thrown away every cycle.
//
// Keyed on the EXPLICIT marker, never on field shape: bug_historian's round-1
// council objection was that a "no required LLM fields" predicate exempts a
// broader class than the evidence justifies, including components that declare
// OPTIONAL source:"llm" fields.
func TestIsSelfContainedSection(t *testing.T) {
	cases := []struct {
		name   string
		level  string
		schema map[string]interface{}
		want   bool
	}{
		{
			name:  "tool with no input_schema is self-contained",
			level: "tool",
			want:  true,
		},
		{
			name:   "tool that declares content fields is NOT exempt",
			level:  "tool",
			schema: map[string]interface{}{"headline": map[string]interface{}{"source": "llm"}},
			want:   false,
		},
		{
			name:  "schemaless SECTION is not exempt — it is the blanked-article case",
			level: "section",
			want:  false,
		},
		{
			name:  "schemaless SITE component is not exempt",
			level: "site",
			want:  false,
		},
		{
			name:  "missing marker defaults closed",
			level: "",
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			comp := componentInfo{
				InputSchema: tc.schema,
				Raw:         map[string]interface{}{"component_level": tc.level},
			}
			if got := isSelfContainedSection(comp); got != tc.want {
				t.Fatalf("isSelfContainedSection(level=%q, schemaFields=%d) = %v, want %v",
					tc.level, len(tc.schema), got, tc.want)
			}
		})
	}
}

// Raw is built by loadSectionComponents, which COALESCEs the column — but a
// component loaded by any other path may not carry the key at all. Absent must
// not panic and must not exempt.
func TestIsSelfContainedSection_MissingRawKey(t *testing.T) {
	comp := componentInfo{Raw: map[string]interface{}{}}
	if isSelfContainedSection(comp) {
		t.Fatal("a component with no component_level key must not be treated as a tool")
	}
	var nilRaw componentInfo
	if isSelfContainedSection(nilRaw) {
		t.Fatal("a zero-value componentInfo must not be treated as a tool")
	}
}

// ---------------------------------------------------------------------------
// toolTemplateValid — the truncation guard, per component level
// ---------------------------------------------------------------------------

// sectionTemplateValid keys on '</section>', which is the wrong signal for a
// tool in BOTH directions. Calibrated against all 27 active tool components on
// 2026-07-20.
func TestToolTemplateValid(t *testing.T) {
	pad := strings.Repeat("x", 200)

	cases := []struct {
		name string
		tpl  string
		want bool
	}{
		{
			// The benchmark tool's real shape: ends '</script>', contains NO
			// '</section>' at all. sectionTemplateValid rejects it, which drops
			// it from the schemas map — so the re-render carries stored HTML and
			// the durable fix never lands.
			name: "whole tool ending in </script> with no </section>",
			tpl:  `<div class="ltb">` + pad + `<script>renderRows();</script></div>`,
			want: true,
		},
		{
			// Real shape of 4 of the 8 damaged rows: truncated mid-JavaScript,
			// but '</section>' appears upstream of the cut, so the section guard
			// passes them.
			name: "truncated mid-JS but contains </section> upstream",
			tpl:  `<section>` + pad + `</section><script>const x = 'Epic`,
			want: false,
		},
		{
			name: "unterminated script with no closing tag",
			tpl:  `<div>` + pad + `<script>ctx.textAlign = 'left';`,
			want: false,
		},
		{
			name: "ends mid-CSS declaration",
			tpl:  `<div>` + pad + `<style>.a{font-weight: bold;`,
			want: false,
		},
		{
			name: "empty template is a stub, not a truncation",
			tpl:  "",
			want: true,
		},
		{
			name: "very short template is a stub, not a truncation",
			tpl:  `<div>stub</div>`,
			want: true,
		},
		{
			name: "case-folded closing tag still counts",
			tpl:  `<div>` + pad + `<SCRIPT>go();</SCRIPT></div>`,
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := toolTemplateValid(tc.tpl); got != tc.want {
				t.Fatalf("toolTemplateValid() = %v, want %v", got, tc.want)
			}
		})
	}
}

// The two guards must stay distinct: the tool shape that motivated this passes
// one and fails the other. If this ever stops being true, the split has become
// pointless and one of them changed underneath.
func TestToolAndSectionGuardsDisagreeOnToolShape(t *testing.T) {
	toolShape := `<div class="ltb">` + strings.Repeat("x", 200) + `<script>renderRows();</script></div>`

	if sectionTemplateValid(toolShape) {
		t.Fatal("precondition changed: sectionTemplateValid now accepts a tool with no </section>")
	}
	if !toolTemplateValid(toolShape) {
		t.Fatal("toolTemplateValid must accept a structurally whole tool template")
	}
}
