// FILE: platform/orchestration/actions/discovery_checks/check_tool_health_contract_rules_test.go
//
// Contract rules 16/17 (check_tool_health.go header §CONTRACT RULES). The
// positive fixture is VERBATIM excerpts from a real ported tool's retired bytes
// — page_components row f1d11768-7aa3-49a1-aca0-451ae0bf400e, webdesign.co.uk's
// tool-text-sanitizer tombstone — NOT from a live page. The live ported corpus
// is shrinking ~5 tools/day under the rebuild programme (41 of 63 gone as of
// 2026-08-25), so a fixture captured from a served page would stop reproducing
// within days; the tombstones keep the ported bytes verbatim for ever and the
// set only grows (the rebuild lane retires by status flip, never DELETE).
package discovery_checks

import (
	"strings"
	"testing"
)

// Verbatim from the tombstone (attribute line and dialog call, with their real
// surrounding context).
const portedTombstoneExcerpt = `
        <div>
            <div style="display: flex; justify-content: space-between; margin-bottom: 0.5rem;">
                <strong>Output (Clean Text)</strong>
                <button class="btn-outline" onclick="copyOutput()" style="font-size: 0.75rem; padding: 2px 8px;">Copy Clean Text</button>
            </div>
            <div id="outputText" class="output-box"></div>
        </div>
<script>
    function copyOutput() {
        navigator.clipboard.writeText(output.innerText);
        alert("Copied clean text!");
    }
    input.addEventListener('input', clean);
</script>`

// The rebuilt shape: real listeners, inline status messages, and the two known
// literal edges that must NOT match — a function whose name merely CONTAINS a
// dialog word, and prose that mentions "onclick" outside any tag.
const rebuiltCleanFixture = `
<div class="tool-container">
  <p>The ported version used an onclick attribute; this one binds real listeners.</p>
  <button type="button" id="copy-btn">Copy</button>
  <p id="copy-status" role="status" aria-live="polite"></p>
</div>
<script>
  function confirmSelection(value) { return value.length > 0; }
  copyBtn.addEventListener('click', function () {
    navigator.clipboard.writeText(text).then(showSuccess, showFailure);
  });
</script>`

func TestContractRulesFireOnRealPortedBytes(t *testing.T) {
	issues := auditContractRules("", portedTombstoneExcerpt, false)
	var checks []string
	for _, i := range issues {
		checks = append(checks, i.check)
		if i.severity != "warning" {
			t.Errorf("%s severity = %q, want warning — these are contract-quality findings, not breakage", i.check, i.severity)
		}
	}
	joined := strings.Join(checks, ",")
	if !strings.Contains(joined, "inline_event_handlers") {
		t.Errorf("rule 16 did not fire on real ported bytes carrying onclick=\"copyOutput()\": %v", checks)
	}
	if !strings.Contains(joined, "blocking_dialogs") {
		t.Errorf("rule 17 did not fire on real ported bytes carrying alert(\"Copied clean text!\"): %v", checks)
	}
}

func TestContractRulesStaySilentOnTheRebuiltShape(t *testing.T) {
	issues := auditContractRules("", rebuiltCleanFixture, false)
	if len(issues) != 0 {
		var got []string
		for _, i := range issues {
			got = append(got, i.check+": "+i.description)
		}
		t.Errorf("contract rules fired on the rebuilt shape — confirmSelection( and prose 'onclick' must not match: %v", got)
	}
}

func TestContractRulesJudgeTemplateForForksAndNothingWhenEmpty(t *testing.T) {
	// A fork with the defect only in its template (rendered empty) is judged on
	// the template — the fork's contract — mirroring auditTool's html selection.
	if issues := auditContractRules(portedTombstoneExcerpt, "", true); len(issues) != 2 {
		t.Errorf("fork with defective template judged wrong: got %d issues, want 2", len(issues))
	}
	// A ported instance with no rendered bytes has nothing to judge (blockers
	// for that state belong to auditTool, not here).
	if issues := auditContractRules(portedTombstoneExcerpt, "", false); len(issues) != 0 {
		t.Errorf("ported instance with empty rendered_html must yield no contract findings, got %d", len(issues))
	}
}
