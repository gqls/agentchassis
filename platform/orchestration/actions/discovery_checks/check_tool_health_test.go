package discovery_checks

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/content"
)

// auditTool is a pure function over (template, rendered, build_status, isFork),
// so the whole of bugs_open/281's detectability claim can be pinned here without
// a database.

// portedWrapper is the shared ported-page component's html_template verbatim
// (sql_for_agents/208): a passthrough with no script, style or contract of its
// own. Every ported instance carries it as templateHTML.
const portedWrapper = `<section class="ported-page" data-component="ported-page">{{.body}}</section>`

// forkTemplate is a minimal well-formed real fork: doc header, style with a
// breakpoint and var() colours, a script.
var forkTemplate = content.ToolDocOpen + "\npurpose: test\n" + content.ToolDocClose + `
<style>.t{color:var(--color-text,#111)} @media (max-width:600px){.t{display:block}}</style>
<div class="t" id="root"></div>
<script>document.getElementById('root').textContent='ok';</script>`

// mindMapShaped mirrors what the owner's visual gate found on webdesign.co.uk's
// Mind Map Studio (bugs_open/281): the tool's OWN <style> block full of bare
// hex on colour properties, an interactive script, and no breakpoint. Before
// 281 this instance was never examined, because it is a ported instance.
const mindMapShaped = `<section class="ported-page">
<style>
  .mm-toolbar { background: #f7f7f9; color: #c9c9d1; border: 1px solid #ececf1; }
  .mm-node { background: #ffffff; color: #d0d0d8; }
  .mm-node.central { background: #f0f0f4; color: #b8b8c2; }
  .mm-btn { color: #cfcfd7; border-bottom: 2px solid #e5e5ea; }
</style>
<div class="mm-toolbar"><button class="mm-btn" id="newMap">+ New Map</button></div>
<div id="canvas"></div>
<script>
  const nodes=[{id:1,text:'Centrdgsdgsdgsdal Idea'}];
  document.getElementById('newMap').addEventListener('click',()=>nodes.push({id:nodes.length+1,text:'Idea'}));
</script>
</section>`

func hasCheck(issues []toolIssue, name string) bool {
	for _, i := range issues {
		if i.check == name {
			return true
		}
	}
	return false
}

func checkNames(issues []toolIssue) string {
	var names []string
	for _, i := range issues {
		names = append(names, i.check)
	}
	return strings.Join(names, ",")
}

// A real fork's contract checks are unchanged by the widening.
func TestAuditTool_ForkTemplateChecksStillFire(t *testing.T) {
	// Empty template on a fork is a blocker.
	issues := auditTool("", "<div>x</div>", "deployed", true)
	if !hasCheck(issues, "empty_template") {
		t.Fatalf("fork with empty template must raise empty_template; got %s", checkNames(issues))
	}

	// Fork template without a doc header draws the warning.
	noHeader := strings.Replace(forkTemplate, content.ToolDocOpen, "", 1)
	noHeader = strings.Replace(noHeader, content.ToolDocClose, "", 1)
	issues = auditTool(noHeader, noHeader, "deployed", true)
	if !hasCheck(issues, "no_doc_header") {
		t.Fatalf("fork without doc header must raise no_doc_header; got %s", checkNames(issues))
	}

	// Opener without closer is the malformed (error) shape.
	malformed := strings.Replace(forkTemplate, content.ToolDocClose, "", 1)
	issues = auditTool(malformed, malformed, "deployed", true)
	if !hasCheck(issues, "malformed_doc_header") {
		t.Fatalf("fork with unclosed doc header must raise malformed_doc_header; got %s", checkNames(issues))
	}

	// A well-formed fork is clean.
	issues = auditTool(forkTemplate, forkTemplate, "deployed", true)
	if len(issues) != 0 {
		t.Fatalf("well-formed fork should be clean; got %s", checkNames(issues))
	}
}

// A ported instance's templateHTML is the shared wrapper — nobody's contract —
// so the TEMPLATE-marked checks must not fire on it, or every one of the 63
// instances draws the same phantom findings.
func TestAuditTool_PortedInstanceSkipsTemplateContractChecks(t *testing.T) {
	clean := `<section class="ported-page"><style>.x{color:var(--c)} @media (max-width:600px){.x{}}</style><script>1</script></section>`
	issues := auditTool(portedWrapper, clean, "deployed", false)
	for _, phantom := range []string{"empty_template", "no_doc_header", "malformed_doc_header"} {
		if hasCheck(issues, phantom) {
			t.Fatalf("ported instance must not raise template-contract check %s; got %s", phantom, checkNames(issues))
		}
	}
	if len(issues) != 0 {
		t.Fatalf("clean ported instance should be clean; got %s", checkNames(issues))
	}

	// And an EMPTY wrapper must not masquerade as an empty fork.
	issues = auditTool("", clean, "deployed", false)
	if hasCheck(issues, "empty_template") {
		t.Fatalf("empty_template is a fork check; ported got %s", checkNames(issues))
	}
}

// The motivating case: the Mind Map Studio's pale-on-pale controls are bare
// hex in its own <style> block, and it has no breakpoint. Both must be
// DETECTABLE on a ported instance — that is the whole of what 281 asks for.
func TestAuditTool_MindMapShapedPortedInstanceIsDetectable(t *testing.T) {
	issues := auditTool(portedWrapper, mindMapShaped, "deployed", false)
	if !hasCheck(issues, "hardcoded_colors") {
		t.Fatalf("Mind-Map-shaped instance must raise hardcoded_colors; got %s", checkNames(issues))
	}
	if !hasCheck(issues, "no_responsive") {
		t.Fatalf("Mind-Map-shaped instance must raise no_responsive; got %s", checkNames(issues))
	}
	// It IS interactive and styled — those must not fire.
	if hasCheck(issues, "no_script") || hasCheck(issues, "no_style") {
		t.Fatalf("instance has script and style; got %s", checkNames(issues))
	}
}

// Blockers that are about the PAGE, not the template, still apply to ported
// instances — and a ported instance with no rendered_html is judged on nothing
// (no fallback to the wrapper, which would draw every content warning at once).
func TestAuditTool_PortedInstanceBlockersStillApply(t *testing.T) {
	issues := auditTool(portedWrapper, "", "deployed", false)
	if !hasCheck(issues, "no_rendered_html") {
		t.Fatalf("ported instance without rendered_html must raise no_rendered_html; got %s", checkNames(issues))
	}
	if hasCheck(issues, "no_script") || hasCheck(issues, "no_style") {
		t.Fatalf("no fallback to the wrapper for content checks; got %s", checkNames(issues))
	}

	issues = auditTool(portedWrapper, mindMapShaped, "pending", false)
	if !hasCheck(issues, "not_deployed") {
		t.Fatalf("undeployed ported instance must raise not_deployed; got %s", checkNames(issues))
	}
}
