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

// ---------------------------------------------------------------------------
// Round-6 council findings
// ---------------------------------------------------------------------------

// bug_historian (r6, MEDIUM) predicted a SECOND call site of the same
// template-validity judgement, from this platform's documented history of the
// identical filter existing twice. It was right: loadSingleComponentSchema was
// still rejecting self-contained tool templates on the '</section>' marker
// while loadComponentSchemas had been fixed. Both now share this predicate, and
// this test is what stops them drifting apart again.
func TestComponentTemplateValid_RoutesByLevel(t *testing.T) {
	pad := strings.Repeat("x", 200)
	toolShape := `<div class="ltb">` + pad + `<script>renderRows();</script></div>`
	sectionShape := `<section>` + pad + `</section>`

	cases := []struct {
		name  string
		tpl   string
		level string
		want  bool
	}{
		{"tool template judged as a tool", toolShape, "tool", true},
		{"the SAME tool template judged as a section is rejected", toolShape, "section", false},
		{"section template judged as a section", sectionShape, "section", true},
		{"truncated tool rejected even with </section> upstream",
			`<section>` + pad + `</section><script>const x = 'Epic`, "tool", false},
		{"empty component_level falls back to the section rule", toolShape, "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := componentTemplateValid(tc.tpl, tc.level); got != tc.want {
				t.Fatalf("componentTemplateValid(level=%q) = %v, want %v", tc.level, got, tc.want)
			}
		})
	}
}

// The asymmetry both editquality (r5) and bug_historian (r6) named: a
// configured-but-unresolved item_key_suffix_field used to Warn and fall back to
// the site-wide key — which IS the collision the field exists to prevent — while
// spec_paths in the same function hard-failed. Now both hard-fail.
func TestCreateWorkItem_ItemKeySuffix_UnresolvedIsHardError(t *testing.T) {
	_, err := runCreateWorkItem(t,
		map[string]interface{}{
			"site_id":               "input_data.site_id",
			"item_key_prefix":       "rerender_tool_fix",
			"item_key_suffix_field": "update_result.component_id",
		},
		siteCollected(nil), // update_result absent
	)
	if err == nil {
		t.Fatal("expected a hard error: falling back to the site-wide key silently reinstates the dedup collision")
	}
	if !strings.Contains(err.Error(), "item_key_suffix_field") {
		t.Fatalf("error should name the field, got: %v", err)
	}
}

// ...and the resolving case still scopes the key rather than leaving it site-wide.
func TestCreateWorkItem_ItemKeySuffix_ScopesTheKey(t *testing.T) {
	componentID := uuid.New().String()
	spec, err := runCreateWorkItem(t,
		map[string]interface{}{
			"site_id":               "input_data.site_id",
			"item_key_prefix":       "rerender_tool_fix",
			"item_key_suffix_field": "update_result.component_id",
			"spec_literal":          map[string]interface{}{"reason": "section_data_resolved"},
		},
		siteCollected(map[string]interface{}{
			"update_result": map[string]interface{}{"component_id": componentID},
		}),
	)
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	if spec == "" {
		t.Fatal("no insert captured")
	}
}

// ---------------------------------------------------------------------------
// create_rerender_items: the per-page item_key must discriminate render MODE
// (bugs_open/024 defect 6). A reason-less assemble-only request and a
// reason-bearing section-render request for the same page must NOT dedup
// against each other: keyed on page_rerender_<page>_<site> alone, a stale
// reason-less item in the backlog suppressed the correct reason-bearing one
// (INSERT ... ON CONFLICT DO NOTHING → 0 items), then re-deployed stale HTML.
// ---------------------------------------------------------------------------

func TestPageRerenderItemKey_DiscriminatesRenderMode(t *testing.T) {
	site := uuid.New()

	assemble := pageRerenderItemKey("about", site, "")
	section := pageRerenderItemKey("about", site, "section_data_resolved")
	image := pageRerenderItemKey("about", site, "image_landed")
	cta := pageRerenderItemKey("about", site, "cta_links_stale")

	// A reason-less request keys on the explicit "assemble" token — never a
	// bare page_rerender_<page>_<site> that a reason-bearing request could also
	// land on.
	if !strings.HasSuffix(assemble, "_assemble") {
		t.Fatalf("reason-less key should end _assemble, got %q", assemble)
	}

	// Every mode must be distinct — the whole point is that they cannot collide
	// on idx_swi_dedup for the same page.
	seen := map[string]string{}
	for name, k := range map[string]string{
		"assemble": assemble, "section": section, "image": image, "cta": cta,
	} {
		if prev, dup := seen[k]; dup {
			t.Fatalf("modes %s and %s produced the SAME key %q — they would suppress each other", name, prev, k)
		}
		seen[k] = name
	}

	// Same mode + same page + same site MUST still collapse to one key, so two
	// concurrent site-wide refreshes still dedup (the behaviour we keep).
	if pageRerenderItemKey("about", site, "") != assemble {
		t.Fatal("same mode/page/site must produce a stable key")
	}
	// Different pages never share a key.
	if pageRerenderItemKey("contact", site, "section_data_resolved") == section {
		t.Fatal("different pages must not share a key")
	}
}

// runCreateRerenderItems drives CreateRerenderItemsAction against a single page
// and returns the item_key it composed for the page_rerender insert ($5).
func runCreateRerenderItems(t *testing.T, site, pageID, reason, componentID string) (itemKey string) {
	t.Helper()

	db, mock, mErr := sqlmock.New()
	if mErr != nil {
		t.Fatalf("sqlmock: %v", mErr)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	// Scoped (component-triggered) runs look up the changed component's
	// dependent pages; hand back our page so it is not filtered out.
	if reason == "section_data_resolved" || reason == "image_landed" {
		mock.ExpectQuery("SELECT pc.page_id").
			WillReturnRows(sqlmock.NewRows([]string{"page_id"}).AddRow(pageID))
	}
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			specArgMatcher{got: &itemKey}, // $5 = item_key
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	collected := map[string]interface{}{
		"site_id":      site,
		"domain":       "example.com",
		"reason":       reason,
		"component_id": componentID,
		"rerender_pages": map[string]interface{}{
			"pages": []interface{}{
				map[string]interface{}{
					"page_id":  pageID,
					"name":     "about",
					"filename": "about.html",
				},
			},
		},
	}
	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData:    collected,
		StepConfig:       models.Step{Config: map[string]interface{}{}},
	}
	if _, err := CreateRerenderItemsAction(context.Background(), params); err != nil {
		t.Fatalf("action failed: %v", err)
	}
	return itemKey
}

func TestCreateRerenderItems_ItemKeyScopedByRenderMode(t *testing.T) {
	// Same page, same site — so the ONLY thing that can differ between the two
	// keys is the render-mode discriminator. That is the regression guard.
	site := uuid.New().String()
	pageID := uuid.New().String()

	assemble := runCreateRerenderItems(t, site, pageID, "", "")
	section := runCreateRerenderItems(t, site, pageID, "section_data_resolved", uuid.New().String())

	if !strings.HasSuffix(assemble, "_assemble") {
		t.Fatalf("reason-less run should emit an _assemble key, got %q", assemble)
	}
	if !strings.HasSuffix(section, "_section_data_resolved") {
		t.Fatalf("section-render run should emit a _section_data_resolved key, got %q", section)
	}
	if assemble == section {
		t.Fatalf("same page/site produced identical keys across modes (%q) — the defect-6 collision is not fixed", assemble)
	}
}
