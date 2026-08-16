// FILE: platform/orchestration/datahelpers/action_inputs.go
// Standardized input extraction for all actions
// Replaces repetitive boilerplate in each action

package datahelpers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// ActionInputSpec declares what inputs an action needs
// Used to standardize extraction across all actions
type ActionInputSpec struct {
	// Required fields - action fails if not found
	Required []string

	// Optional fields - extracted if present, no error if missing
	Optional []string

	// Deprecated field mappings: old config key -> new field name
	// e.g., "site_id_field" -> "site_id"
	// Will log deprecation warning when old pattern is used
	Deprecated map[string]string

	// DefaultValues for optional fields
	Defaults map[string]interface{}

	// ConfigKeys declares the step-config keys this action reads that are NOT
	// data-input fields — settings rather than references (e.g. "max_pages",
	// "upload_results", "scrape_config").
	//
	// Declaring this OPTS THE ACTION IN to unknown-config-key detection: a step
	// carrying a key that is in neither Required, Optional, ConfigKeys,
	// Deprecated nor the framework's own set is reported by UnknownConfigKeys.
	// An action that leaves this empty is not checked at all, and behaves
	// exactly as it did before this field existed.
	//
	// Why opt-in rather than fleet-wide (bugs_open/101, D2): there are 811
	// distinct (action, key) pairs across 228 live actions. A declaration that
	// wrong at that scale would reject working definitions, and an over-strict
	// validator is a considerably worse bug than the inert key it is chasing —
	// cf. this bug's own recorded trap, where a fleet-wide cleanup keyed on
	// "max_pages" would have stripped a live page cap off build-site-planner.
	// Adoption is driven by the coverage report (scripts/audit-config-keys.sh),
	// not by a flag day.
	ConfigKeys []string

	// CheckConfig opts this action into unknown-config-key detection WITHOUT
	// requiring a non-empty ConfigKeys. The recognised set is then
	// Required ∪ Optional ∪ ConfigKeys ∪ Deprecated, exactly as it always was.
	//
	// This exists because the opt-in signal was overloaded onto ConfigKeys, and
	// that is why adoption stalled at 1 action of 152 (measured 2026-07-28).
	// ConfigKeys means something specific — "settings rather than references" —
	// so for the large class of actions whose every key is an extractor field
	// there was NO honest way to opt in: you had to either duplicate keys into a
	// list they do not belong in, or misdescribe a reference as a setting. The
	// cheapest correct action was to do nothing, and 151 actions did.
	//
	// Prefer CheckConfig for an action that passes its own spec to
	// ExtractActionInputs. That extractor reads config[k] for every Required and
	// Optional key (Strategy 0, line ~339), so for such an action the spec is
	// already a verified statement of what it reads from config and opting in
	// asserts nothing new. Use ConfigKeys for genuine settings the extractor
	// does not handle.
	CheckConfig bool

	// StrictConfig turns an unrecognised config key into a hard validation
	// error for this action instead of a warning. Set it only once the
	// recognised set is known to be COMPLETE — verified against every live
	// definition using the action, not just the one in front of you.
	StrictConfig bool

	// ConditionalKeys maps a config key to a human-readable statement of the
	// condition under which it actually takes effect. Keys listed here must
	// ALSO appear in ConfigKeys — they are recognised, not unknown.
	//
	// This exists because declaring a key was, on its own, a way of hiding it.
	// bugs_closed/101 shipped with max_pages/follow_links declared on scrape_web,
	// which made them recognised — so the validator went quiet and
	// scripts/audit-config-keys.sh printed "UNKNOWN KEYS: none" while two live
	// steps still advertised a three-page crawl that a single-page /scrape
	// cannot perform. The detector had two states, unknown and recognised, and
	// the real world had three. Logged in WRONG_CALLS.md 2026-07-28: the fix
	// authored the filter that hid its own remaining case.
	//
	// The condition is prose because it is not generically evaluable — it
	// depends on the action's own dispatch (here, whether the step resolves to a
	// crawl). Enforcement stays where it can actually be decided, at execution;
	// this field is what stops the OFFLINE report claiming a clean bill.
	ConditionalKeys map[string]string

	// SingleOwner declares that this action's effect is scoped to something
	// WIDER than the run that calls it — typically "every row on the site in
	// state X" — so a second live agent carrying the same step is a defect
	// rather than a duplication.
	//
	// It changes NO runtime behaviour. Nothing in the engine reads it; the only
	// consumer is the offline detector `config-key-audit --single-owner-actions`
	// (scripts/audit-single-owner-actions.sh), which reports any declared
	// single-owner action appearing in more than one live definition. This is
	// deliberate: an action cannot tell at execution time whether a sibling
	// agent also carries it, so the check has to be offline and fleet-wide —
	// the same argument recorded on findUnregisteredActions.
	//
	// Why it exists (RFC 006, owner ruling 2026-08-02; bugs_closed/150 is the
	// instance). `triage_detected_items` promotes every `detected` row on a
	// site. It was a step in THREE live agents, so the first copy to run took
	// every row and each later copy honestly reported `promoted: 0` — which
	// improvement-loop's branch read as "nothing to do" and ended the run on
	// "site is clean" while 67 findings sat in the queue. Two properties make
	// that class invisible: the action is idempotent BY EMPTYING (a second run
	// is not an error, it is a success returning zero), and its result
	// describes what THAT CALL did while the branch wants to know what is now
	// TRUE. Those two readings coincide only while exactly one caller exists.
	//
	// So the declaration is the durable half of the fix. Removing the duplicate
	// steps once is a tidy-up that the next agent to gain one silently undoes;
	// a fleet-wide check that fails on the second copy is what closes it.
	//
	// Set this ONLY where a second caller would be wrong. An action that is
	// merely shared — most of them — must leave it false. The check is a
	// per-action assertion, never a fleet-wide default, for the same reason
	// ConfigKeys is opt-in: an over-strict detector is a worse bug than the
	// thing it chases.
	SingleOwner bool

	// DeprecatedConfigKeys maps an OLD LITERAL-SETTING config key to the
	// canonical key the action reads now: e.g. "check_domain" ->
	// "check_pipeline". Honoured by ResolveConfigSetting at the action's own
	// config-read site; see config_key_aliases.go.
	//
	// THIS IS NOT `Deprecated`, AND THE DIFFERENCE IS THE WHOLE POINT
	// (bugs_open/136). `Deprecated` is for extractor FIELDS: there the old
	// key's VALUE is a dot-path resolved against collected_data by
	// ExtractActionInputs Strategy 3. Declaring a literal setting there —
	// `check_domain: "content"` — would make the runtime look up a
	// collected_data key called "content", find nothing, and fall through to
	// the default while reading as though it were wired. Silently wrong, and
	// worse than the inert key it was meant to rescue.
	//
	// Why it exists rather than three hand-rolled shims. The `domain` ->
	// `pipeline` rename landed in Go and never in the data. `create_work_item`
	// carried it with three lines of back-compat Go (item_pipeline falling back
	// to item_domain); `run_discovery_checks` and `triage_detected_items` did
	// not, and their configs have been inert ever since — correct only because
	// each hardcoded default happened to equal what the config asked for. That
	// coincidence broke on 2026-08-04, when two checks that propagate
	// dctx.Pipeline joined completeness-discovery-agent and filed four rows
	// under `design` on an agent whose config says `content`. A shim written in
	// one action's body is invisible to the audit and to the next author; a
	// declaration is neither.
	//
	// Deliberately does NOT set checksConfig(): the alias must work, and be
	// reported, for an action that has not opted into unknown-key detection.
	// Coupling the two would make a declaration silently inert for every action
	// that has not adopted an unrelated mechanism — the "declared but never
	// consulted" shape recorded in WRONG_CALLS.md 2026-07-28, and the same
	// argument already written on SingleOwner above.
	DeprecatedConfigKeys map[string]string

	// RemovedConfigKeys maps a RETIRED config key to a message naming what
	// replaced it. A step carrying one is a HARD validation error — the whole
	// workflow is rejected, on every message — with an error that tells the
	// author the correct spelling instead of silently ignoring the key.
	//
	// This is the third and last state the detector needed. `Deprecated` /
	// `DeprecatedConfigKeys` mean "old name, still WORKS"; ConfigKeys means
	// "recognised"; unknown means "nothing is known". A key that was RETIRED —
	// an author wrote it, it reads as wired, and no code path has ever resolved
	// it — fitted none of those: declaring it recognised silences the detector
	// while the behaviour stays broken (bugs_closed/101), declaring it
	// deprecated makes the runtime resolve its value as a data path and take
	// the default while reading as wired (the LANDMINES `Deprecated`-alias
	// trap), and leaving it unknown produces a once-per-pod Warn that nobody
	// reads while the intended behaviour is silently dropped — which is how
	// improvement-loop's refresh_site_components flag was lost for months
	// (bugs_open/234, the case this field ships with).
	//
	// The map VALUE is prose for a human: name the replacement spelling and,
	// where useful, the bug that retired the key. It is printed verbatim in the
	// validation error.
	//
	// ORDERING IS ON THE DECLARER. The error fires the moment a binary carrying
	// the declaration validates a definition carrying the key — so migrate
	// every live carrier BEFORE committing the declaration (on this tree,
	// committing is shipping). The offline audit reports removed-keys-in-use
	// so the drift is visible before a roll.
	//
	// Opt-in with the unsafe default OFF (empty map — the owner ruling of
	// 2026-08-02 #2), and deliberately independent of checksConfig(), for the
	// same reason as DeprecatedConfigKeys above: a retired key must hard-fail
	// even on an action that has not adopted unknown-key detection.
	RemovedConfigKeys map[string]string
}

// frameworkStepConfigKeys are step-config keys the orchestrator itself reads or
// injects, on any step regardless of action. They are never "unknown".
//
// Derived by grepping the engine's own reads (platform/orchestration/*.go,
// platform/messaging/*.go) rather than written from memory — an under-populated
// list here would produce false positives on working steps, which is the failure
// mode most likely to get this whole check switched off.
var frameworkStepConfigKeys = map[string]bool{
	"action":            true, // dispatch/adapter action selector
	"agent_type":        true, // routing (processor.go, coordinator.go)
	"continue_on_error": true, // loop error handler
	"error_step":        true, // per-step error routing
	"input_fields":      true, // read by ExtractActionInputs for every action
	"loop_item_index":   true, // injected by loop expansion
	"loop_iteration":    true, // injected by loop expansion
	"loop_name":         true, // injected by loop expansion
	"loop_var_name":     true, // injected by loop expansion
	"output_mapping":    true, // coordinator result mapping
	"role":              true, // coordinator
	"target_action":     true, // coordinator
	"timeout_seconds":   true, // coordinator step timeout
	"total_iterations":  true, // loop error handler
}

// IsFrameworkStepConfigKey reports whether a key is read by the orchestrator
// itself rather than by any particular action.
func IsFrameworkStepConfigKey(key string) bool {
	return frameworkStepConfigKeys[key]
}

// UnknownConfigKeys returns the step-config keys that the named action does not
// recognise, sorted.
//
// The second return value, `checked`, is the load-bearing one: it distinguishes
// "this action declared its keys and none were unknown" (checked=true, empty
// result) from "this action has not opted in, so nothing was examined"
// (checked=false, empty result). Both return no keys. Only one of them is
// evidence of anything, and collapsing them is how a quiet result comes to read
// as a pass — the exact shape bugs_open/101 is about.
func UnknownConfigKeys(actionName string, config map[string]interface{}) (unknown []string, checked bool) {
	spec, ok := GetActionInputSpec(actionName)
	if !ok || !spec.checksConfig() {
		return nil, false
	}

	recognised := make(map[string]bool, len(spec.Required)+len(spec.Optional)+len(spec.ConfigKeys))
	for _, k := range spec.Required {
		recognised[k] = true
	}
	for _, k := range spec.Optional {
		recognised[k] = true
	}
	for _, k := range spec.ConfigKeys {
		recognised[k] = true
	}
	// Deprecated keys are recognised on purpose: they are wired, and warning
	// about them here would duplicate ExtractActionInputs' own deprecation
	// warning with a misleading "unknown" label.
	for k := range spec.Deprecated {
		recognised[k] = true
	}
	// Same for DeprecatedConfigKeys, and for the same reason: ResolveConfigSetting
	// honours them and warns at the read site. Calling one "unknown" here would be
	// FALSE — the action does read it. The audit reports these separately, as
	// DEPRECATED, so they stay visible as a migration list rather than becoming
	// quiet (bugs_open/136; the failure mode is bugs_closed/101's, where declaring
	// a key was itself a way of hiding it).
	for k := range spec.DeprecatedConfigKeys {
		recognised[k] = true
	}

	for k := range config {
		// A `!`-suffixed key is the STRICT spelling of the base field
		// (RFC_029 §9 D3) — recognised exactly when the base is. The original
		// spelling is what gets reported if it is not, so the author sees the
		// key they actually wrote.
		base := strings.TrimSuffix(k, "!")
		if base == "" {
			base = k
		}
		if recognised[base] || IsFrameworkStepConfigKey(base) {
			continue
		}
		// Removed keys are NOT unknown — they are known-dead, a distinct third
		// state with its own (harder) consequence: RemovedConfigKeysInUse
		// reports them and the validator rejects them outright. Listing one
		// here as well would double-report it under the softer label, and the
		// softer label is the one that gets read.
		if _, isRemoved := spec.RemovedConfigKeys[base]; isRemoved {
			continue
		}
		unknown = append(unknown, k)
	}
	sort.Strings(unknown)
	return unknown, true
}

// RemovedConfigKeysInUse returns the RETIRED keys present in this step's
// config, mapped to the replacement message each declares. Any entry is a
// definition error — the validator rejects the workflow — so this is the one
// key category where a non-empty result means "stop", not "note".
//
// `declared` draws the same load-bearing distinction as UnknownConfigKeys and
// DeprecatedConfigKeysInUse: "declares removed keys, none present" versus
// "declares none, nothing examined". Deliberately independent of
// checksConfig(), like DeprecatedConfigKeysInUse and for the same reason: a
// retired key must hard-fail even on an action that has not adopted
// unknown-key detection.
func RemovedConfigKeysInUse(actionName string, config map[string]interface{}) (inUse map[string]string, declared bool) {
	spec, ok := GetActionInputSpec(actionName)
	if !ok || len(spec.RemovedConfigKeys) == 0 {
		return nil, false
	}

	inUse = make(map[string]string)
	for removed, message := range spec.RemovedConfigKeys {
		if _, set := config[removed]; set {
			inUse[removed] = message
		}
	}
	return inUse, true
}

// ListRemovedConfigKeys returns every declared retired key, by action.
// Consumed by cmd/config-key-audit so the offline report can find a removed
// key in live definitions BEFORE a binary that rejects it rolls out.
func ListRemovedConfigKeys() map[string]map[string]string {
	out := make(map[string]map[string]string)
	for name, spec := range actionInputSpecs {
		if len(spec.RemovedConfigKeys) == 0 {
			continue
		}
		removed := make(map[string]string, len(spec.RemovedConfigKeys))
		for k, msg := range spec.RemovedConfigKeys {
			removed[k] = msg
		}
		out[name] = removed
	}
	return out
}

// ConditionalConfigKeys returns the conditionally-honoured keys PRESENT in this
// step's config, mapped to the condition under which each takes effect.
//
// These are recognised keys, so UnknownConfigKeys is silent about them by
// design. That silence is the whole reason this function exists: a key can be
// perfectly well declared and still describe behaviour a given step will not
// perform, and an audit that reports only unknown keys calls that clean.
func ConditionalConfigKeys(actionName string, config map[string]interface{}) (map[string]string, bool) {
	spec, ok := GetActionInputSpec(actionName)
	if !ok || len(spec.ConditionalKeys) == 0 {
		return nil, false
	}

	present := make(map[string]string)
	for key, condition := range spec.ConditionalKeys {
		if _, set := config[key]; set {
			present[key] = condition
		}
	}
	if len(present) == 0 {
		return nil, true
	}
	return present, true
}

// ListConditionalConfigKeys returns every declared conditional key, by action.
// Consumed by cmd/config-key-audit so the offline report can name them.
func ListConditionalConfigKeys() map[string]map[string]string {
	out := make(map[string]map[string]string)
	for name, spec := range actionInputSpecs {
		if len(spec.ConditionalKeys) == 0 {
			continue
		}
		conds := make(map[string]string, len(spec.ConditionalKeys))
		for k, v := range spec.ConditionalKeys {
			conds[k] = v
		}
		out[name] = conds
	}
	return out
}

// UndeclaredConditionalKeys returns conditional keys an action declares that are
// missing from its ConfigKeys list — a declaration error, since a conditional key
// is by definition a recognised one. Used by a test rather than at runtime; a
// spec that trips this would make the key report as BOTH unknown and conditional.
func UndeclaredConditionalKeys(actionName string) []string {
	spec, ok := GetActionInputSpec(actionName)
	if !ok {
		return nil
	}
	declared := make(map[string]bool, len(spec.ConfigKeys))
	for _, k := range spec.ConfigKeys {
		declared[k] = true
	}
	var missing []string
	for k := range spec.ConditionalKeys {
		if !declared[k] {
			missing = append(missing, k)
		}
	}
	sort.Strings(missing)
	return missing
}

// checksConfig reports whether this action has opted into unknown-config-key
// detection, by either route. ONE definition of "opted in", used by every gate —
// UnknownConfigKeys, IsStrictConfigAction and ListDeclaredConfigKeys each tested
// `len(ConfigKeys) == 0` separately, which is three chances for the runtime
// check and the offline audit to disagree about which actions are covered.
func (s ActionInputSpec) checksConfig() bool {
	return s.CheckConfig || len(s.ConfigKeys) > 0
}

// IsStrictConfigAction reports whether unrecognised keys should fail validation
// for this action rather than warn.
func IsStrictConfigAction(actionName string) bool {
	spec, ok := GetActionInputSpec(actionName)
	return ok && spec.StrictConfig && spec.checksConfig()
}

// ListDeclaredConfigKeys returns the full recognised key set for every action
// that has opted in. Used by the offline audit (scripts/audit-config-keys.sh) to
// compare the binary's declarations against what live agent_definitions actually
// carry.
func ListDeclaredConfigKeys() map[string][]string {
	out := make(map[string][]string, len(actionInputSpecs))
	for name, spec := range actionInputSpecs {
		if !spec.checksConfig() {
			continue
		}
		keys := make([]string, 0, len(spec.Required)+len(spec.Optional)+len(spec.ConfigKeys)+len(spec.Deprecated)+len(spec.DeprecatedConfigKeys))
		keys = append(keys, spec.Required...)
		keys = append(keys, spec.Optional...)
		keys = append(keys, spec.ConfigKeys...)
		for k := range spec.Deprecated {
			keys = append(keys, k)
		}
		// Alias keys are part of the recognised set at runtime (UnknownConfigKeys
		// above), so they must be part of it here too. Two copies of one gate that
		// can disagree is the drift class this whole tool exists to detect.
		for k := range spec.DeprecatedConfigKeys {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out[name] = keys
	}
	return out
}

// ListSingleOwnerActions returns the names of every action whose spec declares
// SingleOwner, sorted. Read by the offline detector, which cannot ask the
// registry directly — it joins these names against live agent_definitions.
//
// Deliberately NOT gated on checksConfig(): single-ownership is a property of
// the action's EFFECT, unrelated to whether it has opted into config-key
// checking. Coupling the two would make a declaration silently inert for every
// action that has not adopted an unrelated mechanism — the exact "declared but
// never consulted" shape recorded in WRONG_CALLS.md 2026-07-28.
func ListSingleOwnerActions() []string {
	out := make([]string, 0, len(actionInputSpecs))
	for name, spec := range actionInputSpecs {
		if spec.SingleOwner {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// ActionInputs is the result of extraction
type ActionInputs struct {
	// All extracted values (required + optional + defaults)
	Values map[string]interface{}

	// Which deprecated patterns were used (for logging)
	DeprecatedUsed []string

	// Which required fields were missing (before error)
	MissingRequired []string

	// Defaulted records the fields whose value came from spec.Defaults and was
	// never replaced by a caller-supplied one. Provenance only — no value in
	// Values changes because of this map.
	//
	// WHY IT EXISTS (bugs_open/248 finding (b)). Defaults are written into Values
	// BEFORE Strategy 1/2 run, and every later strategy skips a field that
	// already has a value. So a field WITH a default can only ever be set by a
	// Strategy 0 explicit dot-path that resolves; the recursive search that finds
	// a nested `spec.<field>` for every other field never gets to run for it. A
	// caller therefore cannot tell "the caller asked for hero" from "nobody said
	// anything and hero is the default" — and deploy_image_asset shipped every
	// asset as a hero for months because of exactly that.
	//
	// Consumers that must distinguish the two use WasDefaulted. Nothing is
	// obliged to: the map is additive and ignoring it preserves today's
	// behaviour precisely.
	Defaulted map[string]bool
}

// WasDefaulted reports whether key's value came from spec.Defaults rather than
// from the caller. Use it before treating a value as an instruction — see the
// Defaulted field's comment for why a defaulted field is indistinguishable from
// a supplied one without it.
func (ai *ActionInputs) WasDefaulted(key string) bool {
	return ai.Defaulted[key]
}

// Get retrieves a string value, returns empty string if not found
func (ai *ActionInputs) Get(key string) string {
	if v, ok := ai.Values[key].(string); ok {
		return v
	}
	return ""
}

// GetMap retrieves a map value, returns nil if not found
func (ai *ActionInputs) GetMap(key string) map[string]interface{} {
	if v, ok := ai.Values[key].(map[string]interface{}); ok {
		return v
	}
	return nil
}

// GetInt retrieves an int value, returns default if not found
func (ai *ActionInputs) GetInt(key string, defaultVal int) int {
	if v, ok := ai.Values[key].(float64); ok {
		return int(v)
	}
	if v, ok := ai.Values[key].(int); ok {
		return v
	}
	return defaultVal
}

// GetBool retrieves a bool value, returns default if not found
func (ai *ActionInputs) GetBool(key string, defaultVal bool) bool {
	if v, ok := ai.Values[key].(bool); ok {
		return v
	}
	return defaultVal
}

// Has checks if a key exists and is non-nil
func (ai *ActionInputs) Has(key string) bool {
	v, ok := ai.Values[key]
	return ok && v != nil
}

// GetRaw retrieves the raw interface{} value, returns nil if not found
func (ai *ActionInputs) GetRaw(key string) interface{} {
	return ai.Values[key]
}

// LiteralKind classifies a value for Strategy 6's kind guard: "string", "bool",
// "number", or "" for anything that is not a scalar literal (composites, nil).
//
// Numeric widths are deliberately ONE kind. A spec Default is written in Go, so
// `25` is an int; the same value arriving from a jsonb step config is a float64.
// Comparing concrete Go types would reject every numeric override in the fleet.
func LiteralKind(v interface{}) string {
	switch v.(type) {
	case string:
		return "string"
	case bool:
		return "bool"
	case float64, float32,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		json.Number:
		return "number"
	default:
		return ""
	}
}

// IsDottedPathReference reports whether a step-config string is a REFERENCE to
// data in collectedData rather than a literal value to hand the action.
//
// THIS IS THE RESOLVER'S DOT-DISCRIMINATOR, AND IT IS LOAD-BEARING. A string
// containing a dot is a path; it is never its own value. Every arm of
// ExtractActionInputs that has to tell a reference from a literal MUST call this
// rather than open-code `strings.Contains(s, ".")`, so that a future arm cannot
// re-derive the rule slightly differently. Owner ruling 2026-08-15, decision 2 of
// six, raised by RFC_028 and the council's editquality seat: before this existed
// the rule lived in prose at three call sites (Strategy 0, 4 and 6) and in a
// comment block, which is not a control on a tree this many sessions share.
//
// WHAT GETTING IT WRONG COSTS, once, in production: bugs_open/248 finding (a).
// config["asset_key"] was read as a filename instead of a path, and the pipeline
// published 150+ page-visible 404s named `input-data.asset-key.jpg`. That fix
// chose exactly this discriminator and deleted the literal rung rather than
// gating it. Strategy 6 (bugs_open/231) then needed the same rule to decide that
// a dotted string must stay dead against a spec Default.
//
// NOTE THE ASYMMETRY, because it is real and is NOT what this predicate decides:
// a DOTLESS string is still a reference to Strategy 4 (a single-segment
// collected_data key) and a literal to Strategy 6 (a value its author typed into
// a defaulted field). This predicate answers only "is this a multi-segment path",
// which is the half all three arms agree on. CTS-059 records the open question
// about the other half; do not fold it in here without answering that.
func IsDottedPathReference(s string) bool {
	return strings.Contains(s, ".")
}

// ExtractActionInputs extracts inputs according to spec
// Priority order:
//  1. input_fields from config (preferred pattern)
//  2. Deprecated *_field configs (with warning)
//  3. Direct values in input_data
//
// This centralizes the extraction logic that was previously duplicated
// in every action's boilerplate code.
//
// THE PRECEDENCE CHAIN IS A SHARED CONTRACT WITH A STATED OWNER (owner ruling
// 2026-08-15, decision 1 of six; RFC_028). Every action in every workflow resolves
// its inputs here. Two consequences for anyone editing it:
//
//   - THIS COMMENT BLOCK IS THE CONTRACT OF RECORD. The same guarantees used to be
//     stated in five places (here, the ActionInputSpec field docs, defaultshadow.go's
//     class list, concept-register CTS-059, and a LANDMINES entry). They agreed only
//     because a session made them agree. The other four now point HERE; when you
//     change what this function guarantees, change it here and let them point.
//   - THERE IS A BUDGET ON HOW MANY WAYS A VALUE CAN ARRIVE, and it is enforced by
//     a test, not by this comment: resolver_arm_budget_test.go counts the sites in
//     this function that write into result.Values and fails past the budget. Adding
//     one past it is an RFC, not a council round. That is RFC_022's argument one
//     level down — each arm is individually well-argued, and the ninth is still a
//     ninth.
//   - A `!`-SUFFIXED CONFIG KEY IS A STRICT MAPPING (RFC_029 §9 D3, owner-delegated
//     ruling 2026-08-15; opt-in with the unsafe default OFF per the 2026-08-02 §2
//     ruling). `"field!": "<reference>"` means the field resolves via that explicit
//     reference and NOTHING else: never ExtractFields' whole-tree search, never the
//     nested-object fallback, never a deprecated-alias bridge. If the reference
//     does not resolve, extraction FAILS with an error naming the field — for a
//     strict field, no value is better than a searched one (the owner's stated
//     ranking). An unmarked key behaves exactly as it always did.
func ExtractActionInputs(
	collectedData map[string]interface{},
	config map[string]interface{},
	spec ActionInputSpec,
	logger *zap.Logger,
) (*ActionInputs, error) {

	result := &ActionInputs{
		Values:         make(map[string]interface{}),
		DeprecatedUsed: []string{},
		Defaulted:      make(map[string]bool),
	}

	// The `!` STRICT marker (RFC_029 §9 D3): peel it off the config key and
	// remember which fields carry it. Strictness is declared where the mapping is
	// written — beside the path it protects — not in the spec: the author of the
	// step config is the one asserting "this reference is the only truth". A `!`
	// key for a field the spec does not declare is inert here (nothing extracts
	// undeclared fields) and is reported by UnknownConfigKeys where the action
	// has opted in. If a step carries BOTH spellings, the strict one wins: an
	// author who wrote the marker did not mean the unmarked entry to override it.
	strictFields := map[string]bool{}
	for k := range config {
		if base := strings.TrimSuffix(k, "!"); base != k && base != "" {
			strictFields[base] = true
		}
	}
	if len(strictFields) > 0 {
		normalised := make(map[string]interface{}, len(config))
		for k, v := range config {
			if base := strings.TrimSuffix(k, "!"); base != k && base != "" {
				normalised[base] = v
				continue
			}
			if strictFields[k] {
				continue // unmarked twin of a strict key: the marker wins
			}
			normalised[k] = v
		}
		config = normalised
	}

	// Apply defaults first
	for k, v := range spec.Defaults {
		result.Values[k] = v
		result.Defaulted[k] = true
	}

	// Combine all fields we need to look for
	allFields := append([]string{}, spec.Required...)
	allFields = append(allFields, spec.Optional...)

	// Strategy 0: Resolve config values as EXPLICIT path references FIRST.
	// This must run before ExtractFields (Strategy 1/2) because ExtractFields
	// uses aggressive recursive search that can find stale values from
	// previous loop iterations (e.g. claim_result.work_item_id from iter 0).
	// When the config explicitly says "work_item_id": "current_item.id",
	// that explicit path should win over any aggressive search result.
	for _, field := range allFields {
		pathStr, ok := config[field].(string)
		if !ok || pathStr == "" {
			continue
		}
		// Only resolve multi-segment dot-paths (these are unambiguously path references)
		if IsDottedPathReference(pathStr) {
			value := ExtractNestedField(collectedData, pathStr)
			if value != nil {
				result.Values[field] = value
				// Provenance must be cleared wherever a default is overwritten.
				// Three places do that now: here, the Strategy 3 bridge, and
				// Strategy 6 (bugs_open/231) — every OTHER strategy still skips a
				// field that already holds a value, and a default IS a value.
				// TestDefaultedIsClearedWhenACallerSuppliesTheValue pins it.
				delete(result.Defaulted, field)
				logger.Info("Strategy 0: Resolved config path before ExtractFields",
					zap.String("field", field),
					zap.String("path", pathStr),
				)
			}
		}
	}

	// A STRICT field never meets ExtractFields (RFC_029 §9 D3): it is excluded
	// from what Strategy 1/2 request AND from what they merge back, because
	// ExtractFields is the doorway to the whole-tree search — and "explicit
	// resolution only" is the marker's entire meaning.
	withoutStrict := func(fields []string) []string {
		if len(strictFields) == 0 {
			return fields
		}
		kept := make([]string, 0, len(fields))
		for _, f := range fields {
			if !strictFields[f] {
				kept = append(kept, f)
			}
		}
		return kept
	}

	// Strategy 1: Use input_fields if specified in config (preferred)
	if inputFields, ok := config["input_fields"].([]interface{}); ok {
		fieldNames := make([]string, 0, len(inputFields))
		for _, f := range inputFields {
			if name, _ := f.(string); name != "" {
				fieldNames = append(fieldNames, name)
			}
		}
		extracted := ExtractFields(collectedData, withoutStrict(fieldNames), logger)
		for k, v := range extracted {
			// Don't overwrite values already resolved by Strategy 0 (explicit config paths)
			if _, alreadyResolved := result.Values[k]; alreadyResolved {
				continue
			}
			if strictFields[k] {
				continue // ExtractFields' core-field recovery may still emit one; a strict field takes nothing from it
			}
			if v != nil {
				result.Values[k] = v
			}
		}
	} else {
		// Strategy 2: Try to extract all needed fields directly
		extracted := ExtractFields(collectedData, withoutStrict(allFields), logger)
		for k, v := range extracted {
			// Don't overwrite values already resolved by Strategy 0 (explicit config paths)
			if _, alreadyResolved := result.Values[k]; alreadyResolved {
				continue
			}
			if strictFields[k] {
				continue // ExtractFields' core-field recovery may still emit one; a strict field takes nothing from it
			}
			if v != nil {
				result.Values[k] = v
			}
		}
	}

	// OBSERVATION ONLY (RFC_029, the bugfix-213 lane's contribution of
	// 2026-08-15): a DOTLESS config reference — `"result": "handler_result"` — is
	// invisible to Strategy 0's dot-discriminator, so the whole-tree search
	// (Strategy 1/2, above) runs FIRST and a same-named key anywhere in the tree
	// beats the mapping the author actually wrote; Strategy 4 then skips the
	// field as already-resolved. Unique-or-nothing cannot catch that shape: a
	// single foreign key is "unique" and resolves with full confidence. This
	// block measures the class during the Phase 1 observation window and changes
	// NOTHING — same window, same WARN-grep as the conflict instrument. A field
	// marked `!` never meets the search, so it cannot appear here.
	for _, field := range allFields {
		if strictFields[field] || result.Defaulted[field] {
			continue
		}
		pathStr, ok := config[field].(string)
		if !ok || pathStr == "" || IsDottedPathReference(pathStr) {
			continue
		}
		mapped, exists := collectedData[pathStr]
		if !exists || mapped == nil {
			continue
		}
		got, has := result.Values[field]
		if !has {
			continue
		}
		if !reflect.DeepEqual(got, mapped) {
			logger.Warn("aggressive search: explicit single-segment mapping bypassed",
				zap.String("field", field),
				zap.String("reference", pathStr),
				zap.String("resolved_type", fmt.Sprintf("%T", got)),
			)
			// ...and PERSISTED, every occurrence (resolver_findings.go) — the
			// same reason as the conflict instrument: a 48h+ window cannot be
			// read from a ~90s pod log. No-op with no recorder registered.
			recordResolverFinding(logger, ResolverFinding{
				Code:    ResolverFindingMappingBypassed,
				Field:   field,
				Message: "aggressive search: explicit single-segment mapping bypassed",
				Context: map[string]interface{}{
					"field":         field,
					"reference":     pathStr,
					"resolved_type": fmt.Sprintf("%T", got),
				},
			})
		}
	}

	// bridgeSupplied records fields whose value came from a DEPRECATED alias
	// (Strategy 3) in place of a spec Default. Strategy 6 treats these as still
	// overridable, so a step carrying BOTH the deprecated alias and the canonical
	// key resolves to the canonical one — see the note at Strategy 6.
	bridgeSupplied := map[string]bool{}

	// Strategy 3: Check deprecated *_field patterns for any missing fields
	for oldKey, newField := range spec.Deprecated {
		// A STRICT field takes nothing from the bridge (RFC_029 §9 D3): its own
		// `!` mapping is the only reference allowed to decide it. A step carrying
		// both a strict canonical key and a deprecated alias for the same field
		// is asking two sources to be the single truth; the marker wins, and if
		// its reference fails the loud strict error below names the field rather
		// than letting the alias quietly stand in.
		if strictFields[newField] {
			continue
		}
		// Only use deprecated pattern if we don't already have the value.
		//
		// A value that is still only the spec DEFAULT does not count as "already
		// have" (bugs_open/231, owner ruling 2026-08-11 #2). The bridge resolves a
		// dot-path against collected_data — the same operation Strategy 0 performs,
		// reached through a deprecated spelling — so it must beat a Default for the
		// same reason Strategy 0 does. Leaving it to lose meant that renaming a key
		// and providing the documented alias silently reverted the field to its
		// Default: the exact shape DeprecatedConfigKeys' own doc comment warns
		// about, one layer up. Zero live definitions carry a Deprecated alias for a
		// defaulted field (measured 2026-08-13, 164 specs / 184 live agents), so
		// this arm is correctness for the next author, not a live behaviour change.
		if _, hasValue := result.Values[newField]; hasValue && !result.Defaulted[newField] {
			continue
		}

		if pathStr, ok := config[oldKey].(string); ok && pathStr != "" {
			value := ExtractNestedField(collectedData, pathStr)
			if value != nil {
				wasDefaulted := result.Defaulted[newField]
				result.Values[newField] = value
				delete(result.Defaulted, newField)
				// Record that this value came from a DEPRECATED alias, so Strategy 6
				// can let the CANONICAL key win over it (council REVISE round 1,
				// editquality seat). Only tracked when the bridge beat a Default:
				// that is the only case where Strategy 6 can still act, and it is
				// the case where a step carries BOTH spellings.
				if wasDefaulted {
					bridgeSupplied[newField] = true
				}
				result.DeprecatedUsed = append(result.DeprecatedUsed, oldKey)

				logger.Warn("Using deprecated config pattern",
					zap.String("deprecated_key", oldKey),
					zap.String("path", pathStr),
					zap.String("use_instead", fmt.Sprintf("input_fields: [\"%s\"]", newField)),
				)
			}
		}
	}

	// Check for nested object access (backward compat)
	// e.g., if we need "page_id" but got "current_page" object containing "page_id"
	for _, field := range allFields {
		if strictFields[field] {
			continue // implicit parent-object hunting is exactly what `!` forbids
		}
		if _, hasValue := result.Values[field]; hasValue {
			continue
		}

		// Check common nested patterns
		nestedSources := []struct {
			parent string
			child  string
		}{
			{"current_page", field},
			{"rerender_pages", field},
			{"site_record", field},
			{"input_data", field},
		}

		for _, ns := range nestedSources {
			if parent, ok := result.Values[ns.parent].(map[string]interface{}); ok {
				if childVal, exists := parent[ns.child]; exists && childVal != nil {
					result.Values[field] = childVal
					break
				}
			}
		}
	}

	// Strategy 4: Resolve remaining config value references.
	// Dot-path references (e.g. "current_item.id") were already handled
	// by Strategy 0 above. This handles single-segment references
	// (e.g. "spec_data": "site_plan") and any dot-paths that Strategy 0
	// couldn't resolve (data may have been populated by Strategies 1-3).
	for _, field := range allFields {
		if _, hasValue := result.Values[field]; hasValue {
			continue
		}

		pathStr, ok := config[field].(string)
		if !ok || pathStr == "" {
			continue
		}

		// Multi-segment path (has dot): resolve via ExtractNestedField
		if IsDottedPathReference(pathStr) {
			value := ExtractNestedField(collectedData, pathStr)
			if value != nil {
				result.Values[field] = value
				logger.Debug("Resolved config value as dot-path",
					zap.String("field", field),
					zap.String("path", pathStr),
				)
			}
			continue
		}

		// Single-segment: check if it matches a top-level key in collectedData
		// e.g. config has "spec_data": "site_plan" and collectedData has "site_plan": {...}
		if val, exists := collectedData[pathStr]; exists && val != nil {
			result.Values[field] = val
			logger.Debug("Resolved config value as collected_data key",
				zap.String("field", field),
				zap.String("key", pathStr),
			)
		}
	}

	// Strategy 5: numeric and boolean config values, taken as LITERALS.
	//
	// Every strategy above reads step config as config[field].(string) and
	// treats that string as a REFERENCE to resolve against collectedData. A
	// JSON number or boolean fails the type assertion outright, so it never
	// reaches the action and the call site's Go fallback wins instead —
	// silently, while the config reads as though it were live. See
	// bugs_open/042: render_news_json carried max_age_hours 72 and
	// RenderNewsSectionAction defaulted to 72, so config and behaviour agreed
	// and nothing looked wrong until the value was changed to 720 and the
	// render kept behaving like 72. max_items had never been read either.
	//
	// Deliberately restricted to NON-STRING scalars. References are always
	// strings, so this cannot change how any existing reference resolves — it
	// only fills fields that are currently dropped on the floor. A string
	// literal that fails to resolve is left alone on purpose: taking it as its
	// own value would turn a broken reference into a silent literal and mask
	// real wiring bugs. Composite values (objects, arrays) are left alone too,
	// since there is no evidence they were ever intended as literals here.
	for _, field := range allFields {
		if _, hasValue := result.Values[field]; hasValue {
			continue
		}
		raw, exists := config[field]
		if !exists || raw == nil {
			continue
		}
		switch raw.(type) {
		case bool,
			float64, float32,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			json.Number:
			result.Values[field] = raw
			logger.Debug("Strategy 5: took literal scalar config value",
				zap.String("field", field),
			)
		}
	}

	// Strategy 6: an explicit config VALUE beats a spec Default.
	//
	// WHY (bugs_open/231; owner ruling 2026-08-11 #2 — "an explicit config value
	// beats a default" is now the resolver's rule). Defaults are written into
	// Values before every other strategy runs, and each of those strategies opens
	// with `if _, hasValue := result.Values[field]; hasValue { continue }`. A
	// default IS a value and nothing ever deletes from Values, so for a DEFAULTED
	// field Strategies 1/2/3/4/5 and the nested-object block were all unreachable:
	// the only thing that could ever set it was a Strategy 0 dot-path that
	// resolved. Net effect, measured fleet-wide 2026-08-11 and re-measured
	// 2026-08-13: 99 live config entries bound to a defaulted field were dead —
	// 48 static strings and 51 non-string literals — of which 21 disagreed with
	// the default they lost to. `audit_source` was the visible face: four auditors
	// each set a distinctive label and every finding any of them ever wrote was
	// stamped "design-audit" (bugs_open/264 has since repaired that instance).
	//
	// That unreachability is also why this arm's blast radius is provably the dead
	// set and nothing more: a field this loop can touch is one no other strategy
	// could reach, so no behaviour that works today can change. 78 of the 99 carry
	// a value equal to their default (no-op by equality); the other 21 all read
	// their value from config DIRECTLY in the action body, so their live behaviour
	// was already what config said (each read line cited in bugs_open/231).
	//
	// A DOT MEANS REFERENCE, NEVER VALUE. A dotted string that reached here is one
	// Strategy 0 already tried and failed to resolve; taking it as a literal would
	// hand the action a path EXPRESSION as data. That is not hypothetical — it is
	// bugs_open/248 finding (a), where config["asset_key"] was read as a filename
	// and published 150+ page-visible 404s named `input-data.asset-key.jpg`. That
	// fix chose the same discriminator this one uses (a dot means the string is a
	// path, whatever produced it) and deleted the literal rung rather than gating
	// it. So: dotted stays dead, and the default stands.
	//
	// Nor is a resolving DOTLESS string treated as a collected_data reference here,
	// which is what Strategy 4 would do for a field with no default. For a
	// defaulted field that arm has never been reachable, so no live config can
	// depend on it, and every one of the 48 live dotless statics is plainly a value
	// its author typed: repo_name 'agentchassis', ref 'main', country 'GB',
	// severity 'high', and the *_field family naming a field ('repo_analysis'),
	// which wants the NAME and not the object it names. Reading them as literals is
	// what those authors meant; resolving them would replace a typed default with
	// an object of unknown shape. The asymmetry with a non-defaulted field is
	// deliberate and is the known wart: a field carrying a Default is a SETTING
	// with a sensible fallback, which is why authors write values into it.
	//
	// Composites are still left alone (no evidence they were ever meant as
	// literals; zero live instances), and the kind guard below refuses a literal
	// whose type differs from the Default's, so a config typo cannot hand an action
	// a value of a type its spec promised it would never see.
	// PRECEDENCE WITH THE DEPRECATED BRIDGE (council REVISE round 1, editquality
	// seat). Strategy 3 runs earlier in numeric order, so on a step carrying BOTH
	// the deprecated alias and the canonical key the bridge would claim the field
	// first and the canonical spelling would lose — the opposite of what a
	// migration is for, and inconsistent with the dotted case, where Strategy 0
	// already beats the bridge because it runs before it. So a bridge-supplied
	// value stays overridable here: canonical beats deprecated, whatever the shape.
	// The alias is still reported in DeprecatedUsed, because it IS present and IS
	// deprecated; it simply did not decide the value.
	for _, field := range allFields {
		if !result.Defaulted[field] && !bridgeSupplied[field] {
			continue // a caller supplied this field; there is no default to beat
		}
		raw, exists := config[field]
		if !exists || raw == nil {
			continue
		}

		if s, ok := raw.(string); ok && IsDottedPathReference(s) {
			// Reported OFFLINE, not here: whether a path CAN resolve is a runtime
			// fact (the step may legitimately not have run the branch that
			// populates it), so a Warn on every message would be noise on healthy
			// pipelines. `config-key-audit --default-shadowed-keys` can see the
			// whole fleet at once and is where this class is judged.
			logger.Debug("Strategy 6: config path did not resolve; spec default stands",
				zap.String("field", field),
				zap.String("path", s),
			)
			continue
		}

		defaultKind, literal := LiteralKind(spec.Defaults[field]), LiteralKind(raw)
		if defaultKind == "" || literal == "" || defaultKind != literal {
			logger.Warn("Strategy 6: config value's type differs from the spec default's; default stands",
				zap.String("field", field),
				zap.String("default_kind", defaultKind),
				zap.String("config_kind", literal),
				zap.String("bug", "bugs_open/231"),
			)
			continue
		}

		// A required field's Default is the only thing keeping it satisfiable, and
		// "" is not a meaning a required field can carry — so an explicit empty
		// string never turns one into a hard validation failure. No live spec puts
		// a field in both Required and Defaults (measured 2026-08-13, 164 specs),
		// so this guards a state that does not exist yet rather than one it found.
		if s, ok := raw.(string); ok && s == "" && slices.Contains(spec.Required, field) {
			logger.Warn("Strategy 6: refusing an empty config value for a REQUIRED field; default stands",
				zap.String("field", field),
			)
			continue
		}

		result.Values[field] = raw
		delete(result.Defaulted, field)
		logger.Info("Strategy 6: explicit config value beat the spec default",
			zap.String("field", field),
			zap.Any("value", raw),
		)
	}

	// STRICT enforcement (RFC_029 §9 D3): a `!` field that did not resolve via
	// its explicit reference is a HARD error, whatever the spec says about the
	// field being optional or defaulted. A still-standing Default counts as
	// unresolved here — the author demanded the mapping resolve, and a Default
	// silently standing in is the exact silent-substitution shape the marker
	// exists to refuse. Runs before required-field validation so the error names
	// the marker, not a generic missing-field.
	var strictMissing []string
	for field := range strictFields {
		if !slices.Contains(allFields, field) {
			continue // undeclared: nothing extracts it, UnknownConfigKeys reports it
		}
		val, exists := result.Values[field]
		if !exists || val == nil || result.Defaulted[field] {
			strictMissing = append(strictMissing, field)
		}
	}
	if len(strictMissing) > 0 {
		sort.Strings(strictMissing)
		for _, field := range strictMissing {
			ref, _ := config[field].(string)
			logger.Error("strict '!' field did not resolve via its explicit mapping",
				zap.String("field", field),
				zap.String("reference", ref),
			)
		}
		return result, fmt.Errorf(
			"strict '!' fields did not resolve via their explicit mapping: %v "+
				"(RFC_029 §9 D3: a strict field is never resolved by the whole-tree search — "+
				"fix the mapping, or remove the '!' to restore the search fallback)",
			strictMissing)
	}

	// Validate required fields
	var missing []string
	for _, req := range spec.Required {
		val, exists := result.Values[req]
		if !exists || val == nil || val == "" {
			missing = append(missing, req)
		}
	}

	if len(missing) > 0 {
		result.MissingRequired = missing
		return result, fmt.Errorf("missing required fields: %v", missing)
	}

	// Log deprecation summary if any
	if len(result.DeprecatedUsed) > 0 {
		logger.Info("Action used deprecated config patterns",
			zap.Strings("deprecated_keys", result.DeprecatedUsed),
			zap.String("recommendation", "migrate to input_fields pattern"),
		)
	}

	return result, nil
}

// QuickExtract is a convenience wrapper for simple cases
// Returns map directly, panics on missing required (use in tests or simple actions)
func QuickExtract(
	collectedData map[string]interface{},
	config map[string]interface{},
	required []string,
	optional []string,
	logger *zap.Logger,
) (map[string]interface{}, error) {
	spec := ActionInputSpec{
		Required: required,
		Optional: optional,
	}
	result, err := ExtractActionInputs(collectedData, config, spec, logger)
	if err != nil {
		return nil, err
	}
	return result.Values, nil
}

// BuildDeprecationMap creates a Deprecated map from a list of field names
// Assumes old pattern was fieldName + "_field" suffix
// e.g., ["site_id", "page_id"] -> {"site_id_field": "site_id", "page_id_field": "page_id"}
func BuildDeprecationMap(fields []string) map[string]string {
	m := make(map[string]string)
	for _, f := range fields {
		m[f+"_field"] = f
	}
	return m
}

// ---- Input specification registry ----
// Actions can register their specs for documentation and validation

var actionInputSpecs = make(map[string]ActionInputSpec)

// RegisterActionInputSpec registers an action's input specification
// Used for documentation generation and contract validation
func RegisterActionInputSpec(actionName string, spec ActionInputSpec) {
	actionInputSpecs[actionName] = spec
}

// GetActionInputSpec retrieves a registered spec
func GetActionInputSpec(actionName string) (ActionInputSpec, bool) {
	spec, ok := actionInputSpecs[actionName]
	return spec, ok
}

// ListActionInputSpecNames returns the names of every registered spec.
// Used to check spec/registry parity — an action that declares inputs but has
// no registry entry is invisible to the workflow validator, which then reads it
// as a remote action and rejects the workflow (see registry_parity_test.go).
func ListActionInputSpecNames() []string {
	names := make([]string, 0, len(actionInputSpecs))
	for name := range actionInputSpecs {
		names = append(names, name)
	}
	return names
}

// GenerateInputContract creates an input_contract from a spec
// Can be used to auto-generate contract JSON for agent_definitions
func GenerateInputContract(spec ActionInputSpec) map[string]interface{} {
	contract := map[string]interface{}{
		"required": spec.Required,
	}
	if len(spec.Optional) > 0 {
		contract["optional"] = spec.Optional
	}
	if len(spec.Deprecated) > 0 {
		deprecated := make([]string, 0, len(spec.Deprecated))
		for k := range spec.Deprecated {
			deprecated = append(deprecated, k)
		}
		contract["deprecated"] = deprecated
	}
	return contract
}

// ---- Helper to migrate actions ----

// MigrationHelper logs what an action is receiving and suggests migration
// Use temporarily when converting actions to new pattern
func MigrationHelper(actionName string, collectedData, config map[string]interface{}, logger *zap.Logger) {
	logger.Info("=== Migration Helper ===",
		zap.String("action", actionName),
	)

	// Log what's in input_data
	if inputData, ok := collectedData["input_data"].(map[string]interface{}); ok {
		keys := make([]string, 0, len(inputData))
		for k := range inputData {
			keys = append(keys, k)
		}
		logger.Info("Available in input_data", zap.Strings("keys", keys))
	}

	// Log what's in config
	configKeys := make([]string, 0, len(config))
	for k := range config {
		configKeys = append(configKeys, k)
	}
	logger.Info("Config keys", zap.Strings("keys", configKeys))

	// Check for *_field patterns
	var fieldPatterns []string
	for k := range config {
		if strings.HasSuffix(k, "_field") {
			fieldPatterns = append(fieldPatterns, k)
		}
	}
	if len(fieldPatterns) > 0 {
		logger.Warn("Found deprecated *_field patterns - migrate to input_fields",
			zap.Strings("patterns", fieldPatterns),
		)
	}
}
