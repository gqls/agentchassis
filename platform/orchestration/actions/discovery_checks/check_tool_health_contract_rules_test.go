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
	"os"
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

// TestToolPopulationQueriesExcludeTombstones holds the tombstone slot filter
// where it now lives — IN the shared predicate (centralised per council
// 21540c8e round 2's reuse advisory: three ad-hoc copies at three call sites
// is the drift a shared predicate exists to prevent, and a future fourth
// caller inherits the filter instead of repeating the exposure) — and holds
// every page_components-joining caller to actually USING that predicate, so
// the inheritance is real rather than assumed. Round 1's measurement had found
// both acceptance checks exposed (the bugs_closed/360 shape). Comment lines
// are skipped; zero matches of an anchor is a loud failure.
func TestToolPopulationQueriesExcludeTombstones(t *testing.T) {
	// The filter must be inside the shared predicate's const.
	src, err := osReadFileForScan("tool_eligibility.go")
	if err != nil {
		t.Fatalf("cannot read tool_eligibility.go: %v", err)
	}
	constLine, filterLine := -1, -1
	for i, l := range strings.Split(src, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "//") {
			continue
		}
		if constLine == -1 && strings.Contains(l, "const toolEligibilityWhere") {
			constLine = i + 1
		}
		if filterLine == -1 && strings.Contains(l, "pc.build_status IS DISTINCT FROM 'removed'") {
			filterLine = i + 1
		}
	}
	if constLine == -1 {
		t.Fatal("tool_eligibility.go: toolEligibilityWhere const not found — the scan is anchored on nothing")
	}
	if filterLine == -1 || filterLine < constLine {
		t.Errorf("tool_eligibility.go: the shared predicate does not carry pc.build_status IS DISTINCT FROM 'removed' — every caller loses the tombstone exclusion at once (bugs_closed/360 shape; council 21540c8e), or has regressed to a NULL-unsafe spelling (see TestAssemblerAndEligibilityShareTheTombstonePredicate)")
	}

	// Every caller that joins page_components must route through the shared
	// predicate — inheritance only protects a query that actually appends it.
	for _, file := range []string{
		"check_tool_health.go",
		"check_tool_acceptance.go",
		"check_tool_acceptance_due.go",
	} {
		src, err := osReadFileForScan(file)
		if err != nil {
			t.Fatalf("cannot read %s: %v", file, err)
		}
		joinLine, whereLine := -1, -1
		for i, l := range strings.Split(src, "\n") {
			if strings.HasPrefix(strings.TrimSpace(l), "//") {
				continue
			}
			if joinLine == -1 && strings.Contains(l, "JOIN page_components pc ON pc.component_id") {
				joinLine = i + 1
			}
			if whereLine == -1 && strings.Contains(l, "toolEligibilityWhere") {
				whereLine = i + 1
			}
		}
		if joinLine == -1 {
			t.Errorf("%s: population join not found — the scan is anchored on nothing and asserts nothing", file)
			continue
		}
		if whereLine == -1 {
			t.Errorf("%s: joins page_components but does NOT append toolEligibilityWhere — it no longer inherits the tombstone filter (or any eligibility rule); a retire-without-replace slot's unreachable markup would be audited and filed against", file)
		}
	}
}

func osReadFileForScan(name string) (string, error) {
	b, err := os.ReadFile(name)
	return string(b), err
}

// TestAssemblerAndEligibilityShareTheTombstonePredicate is the demand control
// the council's debug_historian seat asked for on 21540c8e round 2: the claim
// "markup in a removed row cannot reach a visitor, so excluding it from audits
// is safe" is only true while the ASSEMBLER's exclusion and this package's
// eligibility filter agree — and that was "a logical argument, not a measured
// one". Writing this test measured it and found them DISAGREEING: the
// assembler read `build_status IS DISTINCT FROM 'removed'` (NULL-safe — a
// NULL-status row is SERVED) while the eligibility predicate read a bare
// `<> 'removed'` (a NULL-status row vanishes from every tool audit). Served
// -but-unaudited is the inversion of the tombstone defect, latent only because
// zero NULL rows existed fleet-wide when measured (2026-08-26; the column is
// nullable, so nothing keeps it that way). Both sides must use the NULL-safe
// spelling, and neither may drift alone. Comment lines are skipped on both
// files — a source-scan test must never let its own prose satisfy it.
func TestAssemblerAndEligibilityShareTheTombstonePredicate(t *testing.T) {
	const nullSafe = "build_status IS DISTINCT FROM 'removed'"

	findInCode := func(src, needle string) int {
		for i, l := range strings.Split(src, "\n") {
			if strings.HasPrefix(strings.TrimSpace(l), "//") {
				continue
			}
			if strings.Contains(l, needle) {
				return i + 1
			}
		}
		return -1
	}

	assembler, err := osReadFileForScan("../rerender_single_page_action.go")
	if err != nil {
		t.Fatalf("cannot read the assembler (rerender_single_page_action.go): %v", err)
	}
	if findInCode(assembler, nullSafe) == -1 {
		t.Errorf("assembler no longer excludes tombstones with %q — either the serving contract moved (update BOTH halves and this test) or the exclusion is gone and removed rows can be served again (bugs_closed/360)", nullSafe)
	}

	eligibility, err := osReadFileForScan("tool_eligibility.go")
	if err != nil {
		t.Fatalf("cannot read tool_eligibility.go: %v", err)
	}
	if findInCode(eligibility, "pc."+nullSafe) == -1 {
		t.Errorf("toolEligibilityWhere does not carry the assembler's NULL-safe spelling (pc.%s) — a NULL build_status row would be served by the assembler and invisible to every tool audit", nullSafe)
	}
}
