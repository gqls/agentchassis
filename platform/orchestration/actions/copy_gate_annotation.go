// FILE: platform/orchestration/actions/copy_gate_annotation.go
//
// Define-by-negation, COUNTED on every rendered LLM section and every compiled
// page (bugs_open/305). Annotation only: it adds keys to a result map and can
// change nothing else. There is no gate here, no refusal, and no LLM call — the
// repair is a separate, opt-in action.
//
// ⚠ OPT-IN PER STEP, DEFAULT OFF — AND THAT IS A CHANGE MADE UNDER A VETO.
//
// It was default-ON fleet-wide, for a reason I still think is good: the repair
// runs on one agent, so counting everywhere is what stops "the copy improved"
// being indistinguishable from "the check was not wired here" — the rule the
// meta-description gate states in its own header, that "a gate a workflow author
// can forget to wire is a comment".
//
// The council's guardian seat VETOED it (correlation c48b7612, round 3) and was
// right on the point that mattered: render_component and compile_page_sections
// are two of the most-invoked actions in the platform, consumed by every
// pipeline that renders or compiles a page, and a default-ON scanner on them is
// a change to a shared contract — which is not licensed by having filed an RFC
// about it. Verbatim: *"Routing a scope objection to architecture review does
// not license deploying the disputed change"*, and *"'we wrote it down and
// routed it' is not the same as 'it was contained'"*.
//
// So the scanner runs only where a step asks for it (`copy_gate_annotate: true`),
// which today is page-content-writer's own render and compile steps, set by
// migration 509 — the same "default OFF in code, the migration is the entire
// enablement surface" shape as migration 474's strip_literal_markdown.
//
// WHAT THAT COSTS, stated so nobody has to rediscover it: the fleet-wide count is
// gone. The number these keys produce is now a number about the wired agents, and
// **`RFC_044` is the open question of whether it should go back to default-ON**.
// Anyone reading `copy_gate_findings` fleet-wide should read that RFC first.
//
// WHY IT IS A WRAPPER RATHER THAN TWO HUNKS INSIDE THE ACTIONS.
//
//	The honest reason first: v3_site_actions.go is a 6,000-line file that other
//	sessions are editing continuously, and on 2026-08-20 it carried another
//	session's uncommitted work referencing symbols that do not exist at HEAD
//	(applyWorkItemFailureLadder, in an untracked file). A pathspec commit takes
//	same-file passengers — that is a documented landmine on this tree — so
//	committing my two hunks would have shipped their half-finished change and
//	broken the compile at HEAD for every session, since `make build-*` builds
//	from HEAD. Moving my edit out of the file was the only version that touches
//	nobody else's work.
//
//	It is also the better structure, which is why it stays after that reason
//	expires. Both annotations are the same idea applied at two altitudes, they
//	are pure functions of a result map, and keeping them together means the
//	section count and the page count cannot drift apart. Nothing is lost by
//	reading the result instead of the action's locals: `content_data` on a render
//	result IS the map that was rendered, and `sections_metadata` on a compile
//	result carries each section's, so the wrapper sees exactly what an inline
//	hunk would have seen.
//
//	The cost is indirection: a reader of RenderComponentAction does not see this.
//	That is paid for at the registry, which is where a reader looks up what an
//	action does, and both entries carry a pointer to this file.
//
// The wrappers are registered in registry.go. Nothing else in the repo calls
// RenderComponentAction or CompilePageSectionsAction directly (verified
// 2026-08-20 by grep over the whole tree), so registry registration is complete
// coverage rather than a partial hook.

package actions

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// copyGateAnnotateEnabled is the opt-in. Absent or false = this wrapper is a
// pass-through and the action behaves exactly as it did before it existed, which
// is what "contained" has to mean for a shared action: not "it only adds keys",
// but "no caller that did not ask for it can tell the difference".
func copyGateAnnotateEnabled(params ActionParams) bool {
	on, _ := params.StepConfig.Config["copy_gate_annotate"].(bool)
	return on
}

// annotateSectionNegation wraps a component-render handler: on success it counts
// the define-by-negation constructions in the content the section was rendered
// from, and attaches them as `copy_gate_findings`.
//
// nil when clean, so the key's presence means "examined, and found something".
// Its ABSENCE means "clean, OR this binary predates the check" — which is why the
// verification for this bug reads the repair's own marker and the artefact, never
// an absent key.
func annotateSectionNegation(inner ActionFunc) ActionFunc {
	return func(ctx context.Context, params ActionParams) (interface{}, error) {
		out, err := inner(ctx, params)
		if err != nil || !copyGateAnnotateEnabled(params) {
			return out, err
		}
		result, ok := out.(map[string]interface{})
		if !ok {
			return out, nil
		}
		content, ok := result["content_data"].(map[string]interface{})
		if !ok || len(content) == 0 {
			return out, nil
		}
		if findings := datahelpers.ScanContentDataForNegation(content); len(findings) > 0 {
			result["copy_gate_findings"] = findings
			params.Logger.Info("copy gate: define-by-negation present in rendered LLM content",
				zap.Int("finding_count", len(findings)),
				zap.Any("component_function", result["component_function"]))
		}
		return out, nil
	}
}

// annotatePageNegation wraps the page-compile handler and reports the PAGE total.
//
// The standard the owner's complaint is measured against is per PAGE — the house
// voice's "a matched contrasting pair is earned once or twice per page at most" —
// and no per-section check can see it: six sections carrying one construction
// each is six on the page, and every one passes a per-section threshold.
//
// Counted by re-scanning each section's content_data rather than by summing the
// per-section annotations: this is the first place that holds the whole page, a
// template-rendered section that never carried an annotation is still counted,
// and a sum would silently read zero the day the metadata extractor stops
// forwarding a key.
func annotatePageNegation(inner ActionFunc) ActionFunc {
	return func(ctx context.Context, params ActionParams) (interface{}, error) {
		out, err := inner(ctx, params)
		if err != nil || !copyGateAnnotateEnabled(params) {
			return out, err
		}
		result, ok := out.(map[string]interface{})
		if !ok {
			return out, nil
		}
		metas, ok := result["sections_metadata"].([]map[string]interface{})
		if !ok {
			return out, nil
		}
		hits := 0
		fields := []string{}
		for i, meta := range metas {
			cd, ok := meta["content_data"].(map[string]interface{})
			if !ok {
				continue
			}
			for _, f := range datahelpers.ScanContentDataForNegation(cd) {
				hits++
				if len(fields) < 24 {
					fields = append(fields, fmt.Sprintf("%d:%v:%v", i, meta["stored_slot_name"], f["field"]))
				}
			}
		}
		if hits > 0 {
			result["copy_gate_page_hits"] = hits
			result["copy_gate_page_fields"] = fields
			params.Logger.Info("copy gate: define-by-negation on the compiled page",
				zap.Int("copy_gate_page_hits", hits),
				zap.Any("page_name", result["page_name"]))
		}
		return out, nil
	}
}
