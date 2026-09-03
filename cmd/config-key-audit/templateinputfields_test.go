// FILE: cmd/config-key-audit/templateinputfields_test.go
//
// bugs_open/453. The positive fixture is the REAL live shape of
// page-content-writer's loop body (step name, action, input_fields and the
// research block verbatim from agent_definitions, 2026-09-03) — the exact
// configuration that discards the researcher's output in production — so the
// check is proven by FIRING against the state it was written to catch, not by
// passing over an invented one (the WFA-007 convention, as loopitemkeys_test.go
// states it).
//
// The negative fixtures are the shapes the rule MUST acquit, and they are the
// interesting half, because every one of them is a false positive a cruder
// implementation actually produces:
//
//   - a {{range}} body's rebound dot (a regex convicts it; on the real writer
//     template that is dozens of false findings);
//   - a root the ACTION injects, which differs per action — the false positive
//     this lane's own first sizing pass shipped for both vision steps;
//   - a root ensureCoreFields always supplies whatever input_fields says;
//   - a step whose action renders no template at all.
package main

import (
	"strings"
	"testing"

	_ "github.com/gqls/agentchassis/platform/orchestration/actions"
)

// ---------------------------------------------------------------------------
// templateRootsReferenced — the parser half
// ---------------------------------------------------------------------------

func TestTemplateRootsReferencedRespectsDotScope(t *testing.T) {
	cases := []struct {
		name     string
		template string
		want     []string
	}{
		{"plain field", `{{.alpha}}`, []string{"alpha"}},
		{"dotted path reports only the root", `{{.alpha.beta.gamma}}`, []string{"alpha"}},
		{
			// The one that matters most. A regex over {{\.(\w+)}} reports
			// "name" here, and "name" is a field of the ITEM, not a root.
			"range body rebinds the dot",
			`{{range .items}}{{.name}} {{.value}}{{end}}`,
			[]string{"items"},
		},
		{"with body rebinds the dot", `{{with .cfg}}{{.enabled}}{{end}}`, []string{"cfg"}},
		{"if does NOT rebind the dot", `{{if .flag}}{{.body}}{{end}}`, []string{"flag", "body"}},
		{"range else is outer scope", `{{range .items}}{{.x}}{{else}}{{.fallback}}{{end}}`, []string{"items", "fallback"}},
		{"nested ranges stay rebound", `{{range .outer}}{{range .inner}}{{.leaf}}{{end}}{{end}}`, []string{"outer"}},
		{"range with declared variables", `{{range $i, $v := .items}}{{$i}} {{$v.name}}{{end}}`, []string{"items"}},
		{"variable assignment reads its source", `{{$x := .source}}{{$x.field}}`, []string{"source"}},
		{"function argument", `{{toJSON .payload}}`, []string{"payload"}},
		{"pipeline", `{{.raw | toJSON}}`, []string{"raw"}},
		{"multi-argument function", `{{if and .left .right}}x{{end}}`, []string{"left", "right"}},
		{"parenthesised pipeline", `{{if or (.a) (.b)}}x{{end}}`, []string{"a", "b"}},
		{"no variables at all", `just prose, no actions`, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := templateRootsReferenced(tc.template)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("roots = %v, want %v", keysOf(got), tc.want)
			}
			for _, w := range tc.want {
				if !got[w] {
					t.Errorf("missing root %q; got %v", w, keysOf(got))
				}
			}
		})
	}
}

// The func map has to be production's or every template using a helper reports
// a parse failure — a fleet-wide false alarm that reads like corrupted config.
func TestTemplateRootsReferencedUsesProductionFuncMap(t *testing.T) {
	if _, err := templateRootsReferenced(`{{placeholder "x"}}{{rangeStart}}{{rangeEnd}}{{toJSON .d}}`); err != nil {
		t.Fatalf("production helpers must parse: %v", err)
	}
	if _, err := templateRootsReferenced(`{{noSuchHelperAnywhere .d}}`); err == nil {
		t.Fatal("an unknown function must be reported as a parse failure, not silently skipped")
	}
}

func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// ---------------------------------------------------------------------------
// The positive fixture: the live writer, verbatim
// ---------------------------------------------------------------------------

// liveWriterLoopBody is page-content-writer's real loop body as of 2026-09-03.
// call_researcher writes research_result and hands to generate_content;
// generate_content's input_fields does not name it, so the whole research block
// — including its {{range}} over sources — renders nothing, on every row.
const liveWriterLoopBody = `[
 {"type": "page-content-writer", "agent_prompt_template": null, "workflow": {"steps": {
  "process_sections_loop": {"action": "loop", "config": {
    "items_field": "sections_for_render.sections_ready",
    "item_variable": "current_section",
    "sub_workflow": {"start_step": "check_needs_research", "steps": {
      "check_needs_research": {"action": "conditional", "config": {
        "condition": "current_section.component.needs_research == true",
        "then_step": "call_researcher", "else_step": "generate_content"}},
      "call_researcher": {"action": "call_agent", "next_step": "generate_content",
        "output_field": "research_result", "config": {
        "agent_type": "research-agent", "target_role": "researcher", "timeout_seconds": 90}},
      "generate_content": {"action": "execute_llm_prompt", "config": {
        "input_fields": ["current_section","render_context","reviewed_brief","current_page",
                         "link_context","site_plan","site_specs","existing_content",
                         "build_mode","rewrite_guidance"],
        "prompt_template": "Write content for the {{.current_section.name}} section of {{.current_page.title}}.\n{{.voice_style}}\nCompany: {{.render_context.company_name}}\nServices: {{.reviewed_brief.services}}\n{{if .link_context.link_constraint_text}}{{.link_context.link_constraint_text}}{{end}}\n{{if .site_specs.specs.evidence_base}}x{{end}}\n{{if .existing_content}}{{.existing_content.raw_markdown}}{{end}}\n{{if .rewrite_guidance}}{{.rewrite_guidance}}{{end}}\n{{range .current_section.llm_field_specs}}- {{.name}} ({{.type}}){{end}}\n{{if .research_result}}\n## Research Findings\n{{.research_result.response.summary}}\n{{range $index, $src := .research_result.response.sources}}- [{{$index}}] {{$src.title}}\n{{end}}{{end}}\n"}}
    }}}}
 }}}
]`

func TestLiveWriterResearchResultIsUnreachable(t *testing.T) {
	rep := auditFixture(t, liveWriterLoopBody)

	if rep.TemplatesChecked != 1 {
		t.Fatalf("templates_checked = %d, want 1 (only the execute_llm_prompt step renders one)", rep.TemplatesChecked)
	}
	var unreachable []templateInputFieldFinding
	for _, f := range rep.Findings {
		if f.Kind == kindUnreachable {
			unreachable = append(unreachable, f)
		}
	}
	if len(unreachable) != 1 {
		t.Fatalf("unreachable findings = %d, want exactly 1: %+v", len(unreachable), rep.Findings)
	}
	f := unreachable[0]
	if f.Path != "steps.process_sections_loop.sub_workflow.generate_content" {
		t.Errorf("path = %q", f.Path)
	}
	if len(f.Roots) != 1 || f.Roots[0] != "research_result" {
		t.Fatalf("roots = %v, want [research_result]", f.Roots)
	}
	if rep.UnreachableFindings != 1 {
		t.Errorf("UnreachableFindings = %d, want 1 (this is what exits 1)", rep.UnreachableFindings)
	}

	// The acquittals matter as much as the conviction. Every root below is one a
	// cruder check reports: rebound-dot fields inside the two {{range}} bodies
	// (.name, .type, $src.title), a platform-injected root (.voice_style), and
	// the sub-fields of declared roots.
	for _, f := range rep.Findings {
		for _, r := range f.Roots {
			switch r {
			case "name", "type", "title", "index", "src", "voice_style", "summary", "response":
				t.Errorf("false positive: reported %q on %s (kind %s)", r, f.Path, f.Kind)
			}
		}
	}
}

// MUTATION PROOF for the dot-scope tracking. The guard under test is the
// `walkNode(v.List, false)` arm for RangeNode; nothing else in the check would
// fail if it were wrong, and the live fixture would still report its one true
// finding — so a passing run does NOT prove the arm works. This asserts the
// counterfactual directly: with the dot treated as root-scoped everywhere (what
// a regex does), the SAME fixture yields findings for the range-body fields.
func TestRangeBodyFieldsWouldConvictWithoutDotScoping(t *testing.T) {
	naive, err := naiveRootsIgnoringDotScope(liveWriterLoopBody)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"name", "type"} {
		if !naive[want] {
			t.Fatalf("the mutation is vacuous: a dot-scope-blind walk of this fixture does not "+
				"even see %q, so the real walk acquitting it proves nothing", want)
		}
	}
	real, err := templateRootsReferenced(extractWriterTemplate(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, mustNotBeRoot := range []string{"name", "type", "title"} {
		if real[mustNotBeRoot] {
			t.Errorf("%q is a range-body field, not a root", mustNotBeRoot)
		}
	}
}

// ---------------------------------------------------------------------------
// Classification, and the shapes that must be acquitted
// ---------------------------------------------------------------------------

func TestClassification(t *testing.T) {
	cases := []struct {
		name      string
		agent     string
		wantKinds map[string][]string // kind -> roots
	}{
		{
			name: "input_data promoted makes it conditional, never a conviction",
			agent: agentJSON("execute_llm_prompt",
				`["input_data","current_page"]`, `{{.mystery_root}}`),
			// current_page is declared and never read, so the reverse arm fires
			// too — both are true of this step at once.
			wantKinds: map[string][]string{
				kindConditional:    {"mystery_root"},
				kindDeclaredUnread: {"current_page"},
			},
		},
		{
			name: "no input_fields at all defaults to input_data, so also conditional",
			agent: agentJSON("execute_llm_prompt",
				``, `{{.mystery_root}}`),
			wantKinds: map[string][]string{kindConditional: {"mystery_root"}},
		},
		{
			name: "always-ensured roots are never findings",
			agent: agentJSON("execute_llm_prompt",
				`["current_page"]`, `{{.domain}} {{.objective}} {{.model}} {{.current_page.title}}`),
			wantKinds: map[string][]string{},
		},
		{
			name: "the platform blocks execute_llm_prompt injects are not findings",
			agent: agentJSON("execute_llm_prompt",
				`["current_page"]`, `{{.voice_style}}{{.build_standard}}{{.current_page.x}}`),
			wantKinds: map[string][]string{},
		},
		{
			// The false positive this lane's first sizing pass actually shipped.
			name: "execute_vision_prompt injects its manifest",
			agent: agentJSON("execute_vision_prompt",
				`["current_page"]`, `{{range .vision_image_manifest}}{{.page_url}}{{end}}{{.current_page.x}}`),
			wantKinds: map[string][]string{},
		},
		{
			// ... and the mirror, which proves the per-action set is really per
			// action rather than a union: vision does NOT call injectPlatformBlocks.
			name: "voice_style is unreachable from a vision prompt",
			agent: agentJSON("execute_vision_prompt",
				`["current_page"]`, `{{.voice_style}}{{.current_page.x}}`),
			wantKinds: map[string][]string{kindUnreachable: {"voice_style"}},
		},
		{
			// The dotted-field trap: ExtractFields stores under the LAST segment.
			name: "a dotted input_field supplies its LEAF, not its head",
			agent: agentJSON("execute_llm_prompt",
				`["sections_for_render.sections_ready"]`, `{{range .sections_for_render.sections_ready}}x{{end}}`),
			// Both arms fire, and together they say the whole thing: the template
			// names a root nothing supplies, AND the declaration that was meant
			// to supply it goes unread because it lands under `sections_ready`.
			wantKinds: map[string][]string{
				kindUnreachable:    {"sections_for_render"},
				kindDeclaredUnread: {"sections_for_render.sections_ready"},
			},
		},
		{
			name: "the same template written against the leaf is clean",
			agent: agentJSON("execute_llm_prompt",
				`["sections_for_render.sections_ready"]`, `{{range .sections_ready}}x{{end}}`),
			wantKinds: map[string][]string{},
		},
		{
			name: "an input_fields entry no template reads",
			agent: agentJSON("execute_llm_prompt",
				`["current_page","nobody_reads_me"]`, `{{.current_page.title}}`),
			wantKinds: map[string][]string{kindDeclaredUnread: {"nobody_reads_me"}},
		},
		{
			name: "input_data is never reported as unread — it is the promotion switch",
			agent: agentJSON("execute_llm_prompt",
				`["input_data"]`, `{{.domain}}`),
			wantKinds: map[string][]string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rep := auditFixture(t, tc.agent)
			if rep.TemplatesChecked != 1 {
				t.Fatalf("templates_checked = %d, want 1", rep.TemplatesChecked)
			}
			got := map[string][]string{}
			for _, f := range rep.Findings {
				got[f.Kind] = append(got[f.Kind], f.Roots...)
			}
			if len(got) != len(tc.wantKinds) {
				t.Fatalf("kinds = %v, want %v", got, tc.wantKinds)
			}
			for kind, wantRoots := range tc.wantKinds {
				if strings.Join(got[kind], ",") != strings.Join(wantRoots, ",") {
					t.Errorf("kind %s roots = %v, want %v", kind, got[kind], wantRoots)
				}
			}
		})
	}
}

// A prompt_template under an action that renders none is inert config, not a
// broken template. Scope is the ACTION, which is what actions.RendersPromptTemplate
// answers.
func TestActionThatRendersNoTemplateIsNotChecked(t *testing.T) {
	rep := auditFixture(t, agentJSON("create_work_item", `["current_page"]`, `{{.never_rendered}}`))
	if rep.TemplatesChecked != 0 {
		t.Fatalf("templates_checked = %d, want 0", rep.TemplatesChecked)
	}
	if len(rep.Findings) != 0 {
		t.Fatalf("findings = %+v, want none", rep.Findings)
	}
}

// Tier 2: a step with no prompt_template of its own renders the AGENT's.
func TestAgentTierTemplateIsChecked(t *testing.T) {
	const fixture = `[
	 {"type": "a", "agent_prompt_template": "{{.from_the_agent_row}}", "workflow": {"steps": {
	   "s": {"action": "execute_llm_prompt", "config": {"input_fields": ["current_page"]}}}}}]`
	rep := auditFixture(t, fixture)
	if rep.TemplatesChecked != 1 || rep.TemplatesAgentTier != 1 {
		t.Fatalf("checked=%d agent_tier=%d, want 1/1", rep.TemplatesChecked, rep.TemplatesAgentTier)
	}
	if len(rep.Findings) != 2 { // unreachable root + current_page declared-unread
		t.Fatalf("findings = %+v", rep.Findings)
	}
	if rep.Findings[0].Kind != kindUnreachable || rep.Findings[0].Tier != "agent" {
		t.Errorf("first finding = %+v, want an agent-tier unreachable root", rep.Findings[0])
	}
}

// A step-level template WINS over the agent's, matching getPromptWithPriority's
// order once tier 1 is absent.
func TestStepTemplateOutranksAgentTemplate(t *testing.T) {
	const fixture = `[
	 {"type": "a", "agent_prompt_template": "{{.from_the_agent_row}}", "workflow": {"steps": {
	   "s": {"action": "execute_llm_prompt", "config": {
	     "input_fields": ["current_page"], "prompt_template": "{{.current_page.x}}"}}}}}]`
	rep := auditFixture(t, fixture)
	if rep.TemplatesAgentTier != 0 {
		t.Fatalf("agent_tier = %d, want 0", rep.TemplatesAgentTier)
	}
	if len(rep.Findings) != 0 {
		t.Fatalf("findings = %+v, want none — the agent template must not be analysed", rep.Findings)
	}
}

// ---------------------------------------------------------------------------
// The blindness refusal
// ---------------------------------------------------------------------------

func TestRefusesAnExportWithoutTheAgentPromptProjection(t *testing.T) {
	// Another mode's export shape: no agent_prompt_template key at all.
	agents, _, err := decodeLiveAgents([]byte(`[{"type":"a","workflow":{"steps":{}}}]`), "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := requireAgentPromptProjection(agents); err == nil {
		t.Fatal("an export missing the projection must be REFUSED — running blind to tier 2 " +
			"produces fewer findings and reads as clean")
	}

	// The correct export, for an agent that simply has no prompt_template:
	// jsonb_build_object emits the key with a JSON null, which IS presence.
	agents, _, err = decodeLiveAgents([]byte(`[{"type":"a","agent_prompt_template":null,"workflow":{"steps":{}}}]`), "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := requireAgentPromptProjection(agents); err != nil {
		t.Fatalf("a projected JSON null is presence, not absence: %v", err)
	}
	if got := agents[0].agentPromptTemplate(); got != "" {
		t.Errorf("agentPromptTemplate() = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func auditFixture(t *testing.T, fixture string) templateInputFieldReport {
	t.Helper()
	agents, failed, err := decodeLiveAgents([]byte(fixture), "test")
	if err != nil {
		t.Fatalf("fixture does not decode: %v", err)
	}
	if failed != 0 {
		t.Fatalf("%d agent row(s) failed to decode", failed)
	}
	return auditTemplateInputFields(agents, failed)
}

// agentJSON builds a one-step agent. inputFields is a JSON array literal, or ""
// to omit the key entirely.
func agentJSON(action, inputFields, tmpl string) string {
	cfg := `"prompt_template": ` + jsonString(tmpl)
	if inputFields != "" {
		cfg = `"input_fields": ` + inputFields + `, ` + cfg
	}
	return `[{"type":"a","agent_prompt_template":null,"workflow":{"steps":{` +
		`"s":{"action":` + jsonString(action) + `,"config":{` + cfg + `}}}}}]`
}

func jsonString(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\t", `\t`)
	return `"` + r.Replace(s) + `"`
}

// extractWriterTemplate pulls the writer fixture's prompt_template back out, so
// the mutation proof works on the same bytes the audit did.
func extractWriterTemplate(t *testing.T) string {
	t.Helper()
	agents, _, err := decodeLiveAgents([]byte(liveWriterLoopBody), "test")
	if err != nil {
		t.Fatal(err)
	}
	sub := agents[0].Workflow.Steps["process_sections_loop"].Config["sub_workflow"].(map[string]interface{})
	steps := sub["steps"].(map[string]interface{})
	gen := steps["generate_content"].(map[string]interface{})
	return gen["config"].(map[string]interface{})["prompt_template"].(string)
}

// naiveRootsIgnoringDotScope is the MUTANT: the same walk with dot-scope
// tracking removed, i.e. what a regex or a scope-blind walker sees. Used only to
// prove the real walk's acquittals are load-bearing rather than vacuous.
func naiveRootsIgnoringDotScope(fixture string) (map[string]bool, error) {
	var out map[string]bool
	agents, _, err := decodeLiveAgents([]byte(fixture), "naive")
	if err != nil {
		return nil, err
	}
	sub := agents[0].Workflow.Steps["process_sections_loop"].Config["sub_workflow"].(map[string]interface{})
	steps := sub["steps"].(map[string]interface{})
	gen := steps["generate_content"].(map[string]interface{})
	tmpl := gen["config"].(map[string]interface{})["prompt_template"].(string)

	out = map[string]bool{}
	// Deliberately crude, exactly as the alternative implementation would be.
	for _, chunk := range strings.Split(tmpl, "{{") {
		chunk = strings.TrimLeft(chunk, " ")
		if !strings.HasPrefix(chunk, ".") {
			continue
		}
		chunk = chunk[1:]
		end := strings.IndexAny(chunk, " .}|")
		if end <= 0 {
			continue
		}
		out[chunk[:end]] = true
	}
	return out, nil
}
