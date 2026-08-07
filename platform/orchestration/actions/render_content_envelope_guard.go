// FILE: platform/orchestration/actions/render_content_envelope_guard.go
//
// Stops the LLM transport envelope reaching RENDER context, where the storage
// guard cannot see it (bugs_open/199). Same payload, same policy, same
// normaliser as the storage seam — decode when provably lossless, refuse when
// not — applied one step earlier, at the point a step's output is resolved into
// a component's content.
//
// WHY A SECOND SEAM AT ALL. bugs_open/190 closed the two WRITE seams
// (content_data_envelope_guard.go, concept register PBP-032): nothing can now
// persist {"type":"text","result":"<string>"} into page_components.content_data.
// That guard cannot speak for what a component RENDERS, and the render gate that
// can — missingRequiredLLMFields at v3_site_actions.go:1843 — only fires for a
// component whose input_schema declares a required source:"llm" field. A
// component with an empty, unrecognised, or optional-only schema declares
// nothing, so there is nothing for that gate to find missing, and the envelope
// renders through missingkey=zero as a blank section. That is the same hole
// shape as the missingkey=zero family in json_envelope.go's header.
//
// THE CENSUS, measured live 2026-08-05, numerator and denominator together:
//
//	315 / 1212 live page_components rows (26%) sit on components the render gate
//	    is structurally blind to — 44 with no/empty input_schema, 1 unrecognised
//	    dialect, 270 v2-dialect with no required source:"llm" field.
//	855 / 1212 sit on components it can speak for.
//
// It is not theoretical: the ONE envelope still live in content_data
// (gaswholesalers.com, page how-pricing-works, slot `pricing`) sits on component
// `pricing`, which is v2-dialect with no required llm field — i.e. squarely in
// the blind population. Its payload recovers only via prose_around, so it is a
// REFUSE case at this seam too, and both seams agree about it.
//
// TRIGGER RATE IS CURRENTLY ZERO, and that is stated rather than hidden: across
// the ~25h orchestration_states retention window, 62 runs carried a
// `generated_content` step output and none was envelope-shaped. The measurement
// is disconfirmable — the same query returns 111 hits for `compose_note` in the
// same window — so this is "the door is open and unused", not "the door is shut".
// The agent_error_log record below is what will tell the next reader which.
//
// WHICH BRANCH ACTUALLY LEAKS — and bugs_open/199's own file gets this wrong, so
// the correction lives here where the fix is. The bug file says the "last resort"
// branch of extractContentWithFallbacks returns the envelope. For the live config
// it is the FIRST loop (v3_site_actions.go:4494-4503). page-content-writer's
// render_section passes content_from: "generated_content.result", so pathsToTry is
// [generated_content.result, generated_content, generated_content.response,
// generated_content.content]; path[0] resolves to the envelope's `result` STRING,
// the map assertion fails and the loop continues; path[1] resolves to the envelope
// MAP, passes the bare `len(m) > 0` test, and is returned. hasContentFields never
// gets a say — that check guards only the last-resort branch.
//
// The last-resort branch is a SECOND real leak, not a dead one: a superset
// envelope such as finetuning's {content,result,type} passes hasContentFields
// precisely because `content` is in its list (v3_site_actions.go:4531-4536), so a
// config whose earlier paths all miss reaches storage by the other door. This is
// why the guard sits at the CALLER and not inside extractContentWithFallbacks:
// one call covers both branches identically, and a future reader who "fixes" the
// first loop cannot silently reopen the second.
//
// WHY DECODE HERE AND NOT ONLY REFUSE. The render seam's output flows on to the
// save — RenderComponentAction returns result["content_data"] = the resolved map
// (v3_site_actions.go:1884-1886), which becomes the sections_metadata the save
// action stores. So this seam is UPSTREAM of the storage seam on the only live
// path. If it refused everything, a payload the council approved the storage seam
// to decode would never reach the branch approved to decode it, and the two seams
// would answer differently about one payload from one constructor. Decoding here
// produces the identical map the save would have produced — ProvenanceClean and
// ProvenanceRepaired discard zero bytes by construction — with one strict gain:
// the SECTION RENDERS ITS REAL CONTENT, where today the template renders the
// envelope through missingkey=zero and ships a blank section whose content_data
// the storage guard then quietly repairs. The storage guard cannot reach back and
// fix rendered_html; only a fix at this seam can.
//
// The decode cannot smuggle anything past the existing gate, because the gate
// runs AFTER it: normalise -> reconcile -> merge -> missingRequiredLLMFields ->
// render. A decoded map that is still missing a required field is refused by the
// gate exactly as before.
//
// BLAST RADIUS OF A REFUSAL, traced rather than assumed. The live
// page-content-writer sets continue_on_error nowhere — not on
// process_sections_loop and not on any substep — and shouldContinueLoopOnError
// defaults false (loop_error_handler.go:70-90), so a refused render fails the
// whole page-content-writer orchestration rather than skipping one section. That
// is proportionate and not new:
//
//  1. It is already the disposition for the 855 rows the gate CAN speak for — the
//     gate fails the same step the same way, because an envelope carries none of
//     the required fields. This makes the disposition uniform instead of
//     schema-dependent.
//  2. With 190 live, a REFUSE-tier payload fails the run anyway, later and more
//     expensively: it rides content_data into the save, where
//     sanitizeSectionsContentData refuses the whole save. Refusing here is the
//     same terminal outcome moved earlier — it saves the remaining iterations'
//     LLM spend and names the component and step, instead of surfacing as a save
//     failure in a different agent's workflow. Nothing is persisted either way,
//     so existing page HTML stays untouched.
//  3. The sibling disposition is also page-scoped: rerender_page_sections
//     escalates the WHOLE page to the writer rather than blanking one section
//     (rerender_page_sections_action.go:325-355). Skipping the section instead
//     would ship a page silently missing a section — the same class, one level
//     up. If section-skip is ever wanted it is a one-key workflow-config change,
//     live immediately, and this file should not pre-empt it.
//
// FALSE POSITIVES AT THIS SEAM, measured, not argued: of 114 live render_context
// maps in orchestration_states, not one carries a `type` key at all — so
// render_from_template, whose content_from IS render_context, cannot trip the
// predicate. Every envelope-shaped object anywhere in live collected_data has
// keys exactly {result,type} and is an LLM step output. The predicate itself is
// unchanged from the approved one: type == "text" AND a STRING result, signature
// not arity (see content_data_envelope_guard.go for why the key count is the
// wrong test and which live row it would have missed).
//
// NOT GUARDED, deliberately, each with its backstop rather than a hope:
//
//	merge_with (current_section.resolved_data)  resolver/DB-derived; the envelope
//	                                            has exactly one constructor
//	                                            (ai_actions.go:864-870) which never
//	                                            writes resolved_data. Measured zero
//	                                            matches. Backstopped: merge_with
//	                                            keys land in sectionContentData and
//	                                            so reach the storage guard.
//	context_field (render_context)              Go-constructed by
//	                                            build_render_context; 114/114 live
//	                                            maps carry no `type` key. And the
//	                                            render_from_template step's
//	                                            content_from: render_context passes
//	                                            through THIS guard regardless,
//	                                            because it goes through
//	                                            extractContentWithFallbacks.
//
// NO OPT-IN FIELD, and the RFC_010 argument is inherited rather than re-made: that
// ruling (owner, 2026-08-02) fires when a seam's widest branch is licensed by
// "callers must all be X" — authority conditioned on CALLER IDENTITY, which only a
// comment can enforce. This refusal is licensed by the PAYLOAD: a shape that is
// never legitimate content, from one named constructor, whose arrival at a render
// is always a defect. Nothing at this seam re-conditions that on who is rendering.
// The same argument carried PBP-032 through council 09bc4b3d APPROVED, and a
// default-OFF switch here would rot unexercised at a trigger rate of zero — the
// failure mode the 2026-07-29 ruling names.
//
// WHAT THIS DOES NOT CLOSE, stated rather than deferred silently: only the
// ENVELOPE class. A schema-less component whose step output is some OTHER kind of
// garbage still renders unchecked. That is bugs_open/199's candidate 2 — make the
// render gate speak for schema-less components at all — which is a much larger
// decision and stays open on its own merits.

package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// normalizeRenderContentEnvelope is the render seam. It returns the content that
// should actually be rendered for a component.
//
// Returns contentData unchanged when it is not an envelope — the dominant path,
// which must leave legitimate content byte-identical. Returns the decoded map
// when the envelope was losslessly recoverable. Returns an error when it was not,
// and that error must fail the render: the alternative is a blank section on a
// live page, which is the defect this guard exists to stop.
func normalizeRenderContentEnvelope(
	ctx context.Context,
	params ActionParams,
	comp *Component,
	contentField string,
	contentData map[string]interface{},
) (map[string]interface{}, error) {
	if !isLLMTransportEnvelope(contentData) {
		return contentData, nil
	}

	componentFunction := ""
	if comp != nil {
		componentFunction = comp.Function
	}

	after, changed, err := normalizeContentDataEnvelope(contentData)
	if err != nil {
		writeRenderEnvelopeLog(ctx, params, componentFunction, contentField, "refused", err.Error())
		params.Logger.Error("RenderComponentAction: step output is an LLM transport envelope, not content — refusing to render (bugs_open/199)",
			zap.String("component_function", componentFunction),
			zap.String("content_field", contentField),
			zap.Strings("keys", sortedContentDataKeys(contentData)),
			zap.Error(err))
		return nil, fmt.Errorf(
			"component %q: the step output at %q is the LLM text-path transport envelope rather than content — "+
				"refusing to render an empty section; leaving existing content untouched (bugs_open/199): %w",
			componentFunction, contentField, err)
	}
	if !changed {
		return contentData, nil
	}

	writeRenderEnvelopeLog(ctx, params, componentFunction, contentField, "decoded", "")
	params.Logger.Warn("RenderComponentAction: decoded an LLM transport envelope out of the step output before rendering (bugs_open/199)",
		zap.String("component_function", componentFunction),
		zap.String("content_field", contentField),
		zap.Strings("keys_before", sortedContentDataKeys(contentData)),
		zap.Strings("keys_after", sortedContentDataKeys(after)),
		zap.Int("envelope_result_bytes", len(fmt.Sprintf("%v", contentData["result"]))),
		zap.Int("decoded_keys", len(after)))
	return after, nil
}

// renderEnvelopeIdentity resolves the site, page and domain for the durable
// record.
//
// The chains are MEASURED, not copied from the save seam — and that distinction
// is the whole reason this function exists. Across every stored
// page-content-writer run (n=110, live 2026-08-05), the paths
// writeContentDataEnvelopeLog uses at the save seam resolve to NOTHING here:
//
//	site_record.site_id            0/110      input_data.site_id          110/110
//	current_page.name              0/110      render_context.site_id       71/110
//	                                          input_data.current_page.name 72/110
//	                                          input_data.page_name         38/110
//	                                          render_context.current_page 109/110
//
// The unions below cover 110/110 on both. They are chains rather than single
// paths because 110 runs is one retention window, not the shape of every caller.
// A miss degrades to uuid.Nil / "" , which writeContentDataEnvelopeLog's sibling
// already tolerates — an unattributed row still counts the firing.
func renderEnvelopeIdentity(params ActionParams) (uuid.UUID, string, string) {
	siteID := uuid.Nil
	for _, path := range []string{"input_data.site_id", "render_context.site_id", "site_record.site_id"} {
		if raw := datahelpers.ExtractNestedFieldString(params.CollectedData, path); raw != "" {
			if parsed, err := uuid.Parse(raw); err == nil {
				siteID = parsed
				break
			}
		}
	}

	pageName := ""
	for _, path := range []string{"input_data.current_page.name", "input_data.page_name", "render_context.current_page"} {
		if v := datahelpers.ExtractNestedFieldString(params.CollectedData, path); v != "" {
			pageName = v
			break
		}
	}

	domain := ""
	for _, path := range []string{"render_context.domain", "input_data.domain", "site_record.domain"} {
		if v := datahelpers.ExtractNestedFieldString(params.CollectedData, path); v != "" {
			domain = v
			break
		}
	}

	return siteID, pageName, domain
}

// writeRenderEnvelopeLog records a decode or a refusal at the render seam, so
// this guard's firings are countable in SQL rather than visible only in a pod
// log. Best-effort: a logging failure must never change what the render does.
//
// Deliberately the SAME error_code as the storage seam. "How often does this
// envelope reach anywhere it should not" then stays one query, and `action`
// is the column that says which seam caught it:
//
//	SELECT action, context->>'outcome', count(*) FROM agent_error_log
//	WHERE error_code = 'CONTENT_DATA_ENVELOPE' GROUP BY 1, 2;
func writeRenderEnvelopeLog(
	ctx context.Context,
	params ActionParams,
	componentFunction string,
	contentField string,
	outcome string,
	detail string,
) {
	if params.DB == nil {
		return
	}

	siteID, pageName, domain := renderEnvelopeIdentity(params)

	severity := "warning"
	message := fmt.Sprintf(
		"component %q resolved its content from %q and found an LLM transport envelope; it was decoded "+
			"losslessly before rendering (bugs_open/199)", componentFunction, contentField)
	if outcome == "refused" {
		severity = "error"
		message = fmt.Sprintf("component %q, content from %q: %s", componentFunction, contentField, detail)
	}

	var siteIDStr string
	if siteID != uuid.Nil {
		siteIDStr = siteID.String()
	}

	// Files under the render seam's own provenance, not the running step's.
	LogActionEntry(ctx, params, agenterrors.Entry{
		SiteID:       siteIDStr,
		Domain:       domain,
		AgentType:    componentRepairAgentType(params),
		StepName:     componentRepairStepName(params, "render_component"),
		Action:       "render_component",
		ErrorMessage: message,
		ErrorCode:    contentDataEnvelopeErrorCode,
		Severity:     severity,
		Context: map[string]interface{}{
			"page_name":          pageName,
			"component_function": componentFunction,
			"content_field":      contentField,
			"outcome":            outcome,
			"seam":               "render",
			"bug":                "bugs_open/199",
		},
	}, params.Logger)
}
