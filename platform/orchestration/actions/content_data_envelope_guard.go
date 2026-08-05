// FILE: platform/orchestration/actions/content_data_envelope_guard.go
//
// Stops the LLM transport envelope being persisted as page_components.content_data
// (bugs_open/190). Decodes it when that is provably lossless; refuses the save
// when it is not.
//
// THE DEFECT. When ExecuteLLMPromptAction cannot parse a model's reply as JSON it
// falls back to the text path (ai_actions.go), which produces the transport
// envelope {"type":"text","result":"<the raw reply>"}. That envelope used to flow
// into content_data verbatim — json_envelope.go's header is the full diagnosis and
// the parse side has since been fixed in three tiers. What was never closed is the
// STORAGE side: nothing between the envelope and the column ever looked at the
// shape, so a poisoned row survives every rebuild.
//
// content_data is the source of truth every reasoned rerender regenerates from, so
// a stored envelope is an armed regression: the render above it can be perfectly
// good — both live instances serve real HTML today — while the next rerender finds
// `type` and `result` where the component's declared fields should be.
//
// MEASURED 2026-08-04, live, numerator and denominator in one query:
//
//	page_components:   2 envelope rows / 1054 non-null content_data / 1207 total
//	site_components:   0 / 54 / 54
//
// The two survivors are finetuning.uk tool-ai-data-risk-checker-guide and
// gaswholesalers.com how-pricing-works.
//
// IT IS PROPAGATION, NOT MINTING — and that is why the guard belongs at the write
// seam. page_component_history holds 65 envelope-shaped rows across 25 pages and 6
// sites, all source='save_page_sections_overwrite'. That is NOT a count of
// envelopes written: the history INSERT archives the state being REPLACED, so it
// counts rebuilds of pages that already carried one. Exactly ONE of the 65
// post-dates 2026-07-18, when the parse fix landed — and on that occasion the save
// archived an envelope and 84ms later created a new row carrying the same envelope
// forward. So no new envelope has been MINTED since mid-July; what continues is a
// stored one surviving every rebuild, via the rerender recycle path and this
// action's own Layer 2 interactive carry-forward.
//
// THE DISCRIMINATOR, and why it is not the key count. type == "text" AND a STRING
// `result`. That is the producer's exact fingerprint: the text path emits
// {"result": <string>, "type": "text"}, where the JSON path emits a
// result that is an OBJECT — and the object form never survives to storage, because
// extractContentWithFallbacks unwraps `.result` when it is a map.
//
// bugs_open/190's own verification recipe says the guard "must key on the exact
// two-key shape". That is wrong, and would have missed one of the two live rows:
// finetuning's is {content, result, type} — THREE keys, an envelope that accreted a
// sibling. Supersets happen by construction, because the rerender merge overlays
// resolved_data onto whatever is stored, and because the modern text path adds
// __json_contract_unmet / __truncated markers. Key on the signature, never the
// arity. The correction is recorded in the bug file.
//
// FALSE POSITIVES. Both conditions must hold with the producer's exact value types,
// so legitimate content passes: a map with a `type` field but no string `result`
// (e.g. {"type":"text","headline":…}) passes; a map with a string `result` but no
// type=="text" (a calculator's display value) passes. Measured: 0 legitimate
// matches in the 1054-row live corpus. The residual risk is deliberately
// asymmetric — an unforeseen legitimate match is REFUSED loudly and nothing is
// written, rather than silently rewritten; silent rewriting additionally requires
// its `result` string to parse as a JSON object, which no observed row does.
//
// DECODE VS REFUSE — the parser's provenance IS the line.
//
//	clean / repaired  -> decode and store. These are the only tiers that discard
//	                     zero bytes; `repaired` rewrites escaping only.
//	everything else   -> REFUSE the save. fenced_block, prose_around and
//	                     reemitted_value all discard bytes, and the discarded bytes
//	                     can be the real content.
//
// gaswholesalers is the living proof of the second rule and the reason it is not
// merely cautious: its payload recovers via prose_around to 7 tier_* keys totalling
// 131 chars, while the actual section copy sits in the markdown tail the recovery
// correctly refuses to guess at. Decoding it would silently replace a page's
// content with 131 characters. So this policy is also the mechanical guarantee that
// no automated repair ever fires at that row: every save carrying it is refused,
// the page keeps serving its current HTML, and its existing needs_human_review item
// stays the owner.
//
// THE SUPERSET CASE is the subtle one. When envelope keys coexist with real sibling
// content, a lossless decode keeps every sibling, drops `type`/`result`/`__`-markers,
// and — where a key is present on BOTH sides — stores it only if the two values are
// deep-equal. Two candidate values for one field is a choice, and this platform does
// not guess at content: differing values refuse.
//
// NO OPT-IN FIELD, and the argument, because RFC_010 (owner ruling 2026-08-02) is
// the first thing a reviewer will raise. That ruling fires when a seam's widest
// branch is licensed by "callers must all be X" — authority conditioned on CALLER
// IDENTITY, which only a comment can enforce. This refusal is licensed by the
// PAYLOAD: a shape that is never legitimate content, from one named constructor,
// whose storage is always a defect. There is no caller-conditional licence to turn
// into a field. The neighbouring guards in save_page_sections_action.go — the
// content-regression guard, the interactivity guard, the owned-page guard, the stub
// skip — all refuse unconditionally for the same reason. The 194 lane's
// refuse_save_without_sections_metadata is opt-in for the OPPOSITE reason: absent metadata is
// common and often legitimate, so unconditional refusal there would have converted
// thousands of good saves into failures; here it converts approximately zero. An
// opt-OUT is rejected too: it would make the bad state representable again by
// config, and the 2026-07-29 ruling records that default-OFF switches rot
// unexercised.
//
// WHICH WRITERS ADOPT THIS, and why the rest cannot produce the shape (census of
// every content_data writer, each read rather than grepped):
//
//	save_page_sections_action.go INSERT        ADOPTS — proven live ingress
//	section_editor_actions.go persist switch   ADOPTS — applyContentEdit seeds from
//	                                           the existing row, so a field_updates
//	                                           merge carries type/result forward
//	rebuild_blog_listing / create_report_page  no — Go map literal, shape fixed
//	deploy_tool / create_tool_component        no — writes '{}'::jsonb
//	adopt_verbatim                             no — Go map literal
//	refresh_product_specs                      no — merges one key, cannot introduce
//	save_page_sections history INSERT          no — archiving the replaced state is
//	                                           its job; the evidence must be kept
//
// site_components needs no guard: it has NO automated content_data writer at all
// (every actions-package writer touches rendered_html / build_status / component_id
// only), which is why it measures 0/54 clean — structural, not luck.
//
// The predicate is pure and would move to datahelpers unchanged the day a caller
// outside this package needs it; it lives here because the normaliser needs
// ParseLLMJSONWithProvenance, which is in this package, and splitting one rule
// across two packages is the drift class this codebase keeps being bitten by.

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// contentDataEnvelopeErrorCode is the agent_error_log code for a refused save, so
// refusals are countable in SQL rather than only visible in pod logs.
const contentDataEnvelopeErrorCode = "CONTENT_DATA_ENVELOPE"

// contentDataEnvelopeSQLPredicate is the SQL twin of isLLMTransportEnvelope, kept
// beside it so census and verification queries cannot drift from the Go. Assumes a
// column named content_data.
const contentDataEnvelopeSQLPredicate = `content_data->>'type' = 'text' AND jsonb_typeof(content_data->'result') = 'string'`

// maxEnvelopeUnwrapDepth bounds the double-wrap defence. bugs_closed/054 records a
// double-wrapped envelope reaching storage, so one level of nesting is real; three
// is far past anything observed and stops a crafted payload spinning here.
const maxEnvelopeUnwrapDepth = 3

// isLLMTransportEnvelope reports whether m is the text-path transport envelope
// rather than component content.
//
// Signature, not arity — see the file header for why the key count is the wrong
// test and which live row it would have missed.
func isLLMTransportEnvelope(m map[string]interface{}) bool {
	if len(m) == 0 {
		return false
	}
	if t, _ := m["type"].(string); t != "text" {
		return false
	}
	_, resultIsString := m["result"].(string)
	return resultIsString
}

// isEnvelopeMarkerKey reports whether k is a platform marker the text path adds
// alongside the envelope (json_contract_unmet, truncation markers). They describe
// the transport, not the content, so a decode drops them with type and result.
func isEnvelopeMarkerKey(k string) bool {
	return strings.HasPrefix(k, "__")
}

// normalizeContentDataEnvelope returns the content_data that should actually be
// stored for m.
//
// Returns (m, false, nil) unchanged when m is not an envelope — the no-op path,
// which must leave legitimate content byte-identical. Returns the decoded map and
// true when the envelope was losslessly recoverable. Returns an error when it was
// not, and that error must fail the save: the alternative is persisting a payload
// whose real content nobody can rebuild.
func normalizeContentDataEnvelope(m map[string]interface{}) (map[string]interface{}, bool, error) {
	if !isLLMTransportEnvelope(m) {
		return m, false, nil
	}

	current := m
	changed := false

	for depth := 0; depth < maxEnvelopeUnwrapDepth; depth++ {
		if !isLLMTransportEnvelope(current) {
			return current, changed, nil
		}

		raw, _ := current["result"].(string)
		decoded, provenance, err := ParseLLMJSONWithProvenance(raw)
		if err != nil {
			return nil, false, fmt.Errorf(
				"content_data holds an LLM transport envelope whose payload is not recoverable JSON "+
					"(%w) — refusing to persist it (bugs_open/190)", err)
		}

		// Only the two lossless tiers may be stored. The rest discard bytes, and
		// the discarded bytes can be the whole of the real content.
		if provenance != ProvenanceClean && provenance != ProvenanceRepaired {
			return nil, false, fmt.Errorf(
				"content_data holds an LLM transport envelope recoverable only via %q, which discards "+
					"the text around the value it found — refusing to guess at the content; this needs a "+
					"human (bugs_open/190)", provenance)
		}

		decodedMap, ok := decoded.(map[string]interface{})
		if !ok {
			return nil, false, fmt.Errorf(
				"content_data holds an LLM transport envelope whose payload decoded to %T rather than a "+
					"JSON object — refusing to persist it (bugs_open/190)", decoded)
		}

		merged, err := mergeEnvelopeSiblings(current, decodedMap)
		if err != nil {
			return nil, false, err
		}

		current = merged
		changed = true
	}

	if isLLMTransportEnvelope(current) {
		return nil, false, fmt.Errorf(
			"content_data holds an LLM transport envelope nested more than %d deep — refusing to "+
				"persist it (bugs_open/190)", maxEnvelopeUnwrapDepth)
	}
	return current, changed, nil
}

// mergeEnvelopeSiblings combines the decoded payload with any real sibling keys the
// envelope had accreted, refusing when the two disagree about a key.
func mergeEnvelopeSiblings(envelope, decoded map[string]interface{}) (map[string]interface{}, error) {
	merged := make(map[string]interface{}, len(envelope)+len(decoded))

	for k, v := range envelope {
		if k == "type" || k == "result" || isEnvelopeMarkerKey(k) {
			continue
		}
		merged[k] = v
	}

	for k, v := range decoded {
		existing, hasSibling := merged[k]
		if hasSibling && !reflect.DeepEqual(existing, v) {
			return nil, fmt.Errorf(
				"content_data holds an LLM transport envelope whose payload and its sibling key %q "+
					"disagree — two candidate values for one field is a choice, and this needs a human "+
					"rather than a guess (bugs_open/190)", k)
		}
		merged[k] = v
	}

	return merged, nil
}

// sanitizeSectionsContentData is the save seam. It normalises every section's
// content_data in place and returns an error if any of them must be refused.
//
// Placement matters and is asserted by the caller's position, not by this comment:
// it runs after the last point at which content_data can enter the section set
// (the Layer 2 interactive carry-forward) and before the history snapshot and the
// DELETE, so a refused save writes NOTHING — the same rule the completeness floor
// and the 194 refusal state in the same action.
func sanitizeSectionsContentData(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	pageName string,
	sections []SectionData,
) error {
	for i := range sections {
		before := sections[i].ContentData
		if !isLLMTransportEnvelope(before) {
			continue
		}

		slot := sections[i].ComponentName
		after, changed, err := normalizeContentDataEnvelope(before)
		if err != nil {
			writeContentDataEnvelopeLog(ctx, params, siteID, pageName, slot, "refused", err.Error())
			return fmt.Errorf("save_page_sections: section %q: %w", slot, err)
		}
		if !changed {
			continue
		}

		sections[i].ContentData = after
		writeContentDataEnvelopeLog(ctx, params, siteID, pageName, slot, "decoded", "")
		params.Logger.Warn("SavePageSectionsAction: decoded an LLM transport envelope out of content_data (bugs_open/190)",
			zap.String("page_name", pageName),
			zap.String("slot_name", slot),
			zap.Strings("keys_before", sortedContentDataKeys(before)),
			zap.Strings("keys_after", sortedContentDataKeys(after)),
			zap.Int("envelope_result_bytes", len(fmt.Sprintf("%v", before["result"]))),
			zap.Int("decoded_keys", len(after)))
	}
	return nil
}

// sortedKeys renders a map's key set deterministically for logging. A recovery
// that is not measured is how this class stayed invisible for four months, so the
// before/after key sets are logged rather than a bare "repaired" flag.
func sortedContentDataKeys(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// writeContentDataEnvelopeLog records a decode or a refusal in agent_error_log, so
// the guard's firings are countable in SQL. Best-effort: a logging failure must
// never change what the save does.
func writeContentDataEnvelopeLog(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	pageName string,
	slotName string,
	outcome string,
	detail string,
) {
	if params.DB == nil {
		return
	}

	severity := "warning"
	message := fmt.Sprintf(
		"page %q section %q carried an LLM transport envelope as content_data and it was decoded "+
			"losslessly before persisting (bugs_open/190)", pageName, slotName)
	if outcome == "refused" {
		severity = "error"
		message = fmt.Sprintf("page %q section %q: %s", pageName, slotName, detail)
	}

	contextJSON, err := json.Marshal(map[string]interface{}{
		"page_name": pageName,
		"slot_name": slotName,
		"outcome":   outcome,
		"bug":       "bugs_open/190",
	})
	if err != nil {
		params.Logger.Warn("content_data envelope: failed to marshal context", zap.Error(err))
		return
	}

	var siteIDArg interface{}
	if siteID != uuid.Nil {
		siteIDArg = siteID
	}

	if _, err := params.DB.ExecContext(ctx, `
		INSERT INTO agent_error_log
		    (site_id, domain, agent_type, step_name, action,
		     error_message, error_code, severity, context)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, $8, $9::jsonb)`,
		siteIDArg,
		datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain"),
		componentRepairAgentType(params), saveSectionsStepName(params),
		"save_page_sections",
		message, contentDataEnvelopeErrorCode, severity, string(contextJSON),
	); err != nil {
		params.Logger.Warn("content_data envelope: failed to write record", zap.Error(err))
	}
}
