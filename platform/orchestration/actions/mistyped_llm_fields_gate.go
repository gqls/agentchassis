// FILE: platform/orchestration/actions/mistyped_llm_fields_gate.go
//
// The opt-in arming for bugs_open/260's PRE-RENDER type gate, kept in its own
// file for the same reason dead_url_guard.go is: the decision to refuse is a
// policy with a history, and burying it in a render function is how the history
// gets lost.
//
// WHY OPT-IN WHEN THE PRESENCE GATE BESIDE IT IS UNCONDITIONAL. The presence
// gate (missingRequiredLLMFields) fires on content that would render EMPTY —
// the section vanishes either way, so refusing changes nothing that works. The
// type check is different: it keys on the SCHEMA, not on the template, so a
// mistyped field that this component's template never references renders
// perfectly well today. An unconditional refusal would therefore be new
// authority over content that currently ships. Owner ruling 2026-08-02 §2:
// new authority on a shared seam arrives as a field whose UNSAFE side is the
// default, so the decision is visible to a reviewer of the CALLER.
//
// The seam's hard error is the complete detector and is NOT gated by this key.
// What the key buys is EARLIER and BETTER-NAMED: the build fails before the
// render with "steps[2].branches: declared array (items: object), got string"
// rather than after it with a template line number.
//
// Fails OPEN on a mis-typed config value, exactly as shouldRefuseDeadURLControls
// does: a config value that is not a bool is a mistake, and a mistake must not
// switch a fleet-wide refusal on by accident.
package actions

// mistypedLLMFieldsConfigKey is the step-config field that arms the refusal.
// Unset or false ⇒ pre-existing behaviour, byte for byte.
const mistypedLLMFieldsConfigKey = "refuse_mistyped_llm_fields"

// refuseMistypedLLMFields reports whether this step is armed to refuse a render
// whose content contradicts the component's declared field types.
func refuseMistypedLLMFields(config map[string]interface{}) bool {
	armed, _ := config[mistypedLLMFieldsConfigKey].(bool)
	return armed
}

// absentRequiredRecordConfigKey arms the RENDER-TIME record of schema-required
// fields that rendered empty (bugs_open/342). A separate key from the refusals
// above because it is a different authority: this one declines to ship nothing
// and only files a note, where those decline to render.
//
// Default OFF for the reason three seats gave the dead-URL record arm on council
// 98852baa: an unconditional new DB write on a shared render path is new
// authority whatever its size. Unset means today's behaviour, byte for byte.
const absentRequiredRecordConfigKey = "record_absent_required_fields"

// recordAbsentRequiredFields reports whether this step should file a
// required_fields_missing item for fields the seam found absent at render time.
// Same fail-OPEN-on-a-mistyped-value semantics as its siblings: a config value
// that is not a bool is a mistake, and a mistake must not switch on a fleet-wide
// DB write by accident.
func recordAbsentRequiredFields(config map[string]interface{}) bool {
	armed, _ := config[absentRequiredRecordConfigKey].(bool)
	return armed
}

// absentRequiredRefuseConfigKey arms the REFUSAL half of bugs_open/342 on the
// section-editor persist switch: an edit whose render left a schema-required
// source:"llm" field EMPTY is declined, and the live section keeps what it had.
//
// A third key rather than a reuse of either sibling, because it is a third
// authority: refuse_mistyped_llm_fields declines to render on a TYPE
// contradiction, record_absent_required_fields files a note, and this one
// declines to PERSIST onto an already-live page. The editor routes are the two
// call sites that write rendered_html with no validate_content between them and
// the reader — the pre-render gates already refuse on the build and rerender
// paths, so this closes the last door on which "no refusal was added anywhere"
// (the bug file's own residual) was still true.
//
// Default OFF (owner ruling 2026-08-02 §2): an edit that leaves a required
// field empty SUCCEEDS today — the section ships blank and page assembly drops
// it — so refusing is new authority over content that currently persists, and
// it arrives as a field whose unsafe side is the default, visible to a reviewer
// of the STEP. Armed by migration on section-editor's apply_edit step ONLY
// after a chassis image carrying this code has rolled (the 502 ordering rule:
// a key naming behaviour the binary does not have is a no-op that LOOKS
// applied).
//
// ⚠ Interaction, stated rather than hidden: section-editor's apply_edit step
// declares no error_step, so until bugs_open/344 (completion tramples a
// failure flag) lands, the DRIVING work item of a refused edit may still read
// `complete`. The refusal's protection does not depend on that — the live page
// is untouched either way, and the required_fields_missing item is filed
// BEFORE the refusal fires, so the defect is on the queue regardless.
const absentRequiredRefuseConfigKey = "refuse_absent_required_fields"

// refuseAbsentRequiredFields reports whether this step is armed to refuse
// persisting an edit whose render left schema-required fields empty. Fail-OPEN
// on a mistyped value, exactly as both siblings: a config mistake must not
// switch on a refusal by accident.
func refuseAbsentRequiredFields(config map[string]interface{}) bool {
	armed, _ := config[absentRequiredRefuseConfigKey].(bool)
	return armed
}

// refusePersistForAbsentRequired is the DECIDING ARM of the bugs_open/342
// refusal, factored so a test exercises exactly what the action runs: armed on
// this step, AND the render actually left schema-required fields empty. Both
// halves matter — unarmed must mean today's behaviour byte for byte, and an
// armed step with a clean render must persist normally or the arm has merely
// stopped completing edits.
func refusePersistForAbsentRequired(config map[string]interface{}, absent []string) bool {
	return refuseAbsentRequiredFields(config) && len(absent) > 0
}
