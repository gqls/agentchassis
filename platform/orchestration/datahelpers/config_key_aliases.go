// FILE: platform/orchestration/datahelpers/config_key_aliases.go
//
// Deprecated aliases for LITERAL step-config settings — the half of the
// deprecation story ActionInputSpec did not have (bugs_open/136).
//
// The estate already had one alias mechanism, ActionInputSpec.Deprecated, and it
// covers exactly one shape: a config key whose VALUE is a dot-path into
// collected_data ("site_id_field": "site_record.site_id"). ExtractActionInputs
// Strategy 3 resolves it and warns. That is a reference alias.
//
// It does not cover a SETTING — a key whose value IS the value, read straight
// from params.StepConfig.Config, which is what ConfigKeys is documented to mean.
// Putting one in Deprecated is not merely ineffective, it is actively wrong:
// `check_domain: "content"` would be resolved as a collected_data lookup for a
// key named "content", find nothing, and fall through to the hardcoded default
// while the declaration reads as though the key were wired.
//
// So an author who renamed a setting had exactly two options: hand-roll a
// fallback in the action body, or do nothing. Of the three actions that carried
// the `domain` -> `pipeline` rename, one hand-rolled it and two did nothing —
// and the two that did nothing spent nine months correct only by luck, because
// each hardcoded default happened to equal what its config asked for. On
// 2026-08-04 the luck ran out: two checks that propagate dctx.Pipeline joined
// completeness-discovery-agent, whose config asks for `check_domain: "content"`,
// and four work items were filed under `design`.
//
// This file gives the third option: DECLARE the alias. The declaration is read
// by the runtime (here), by the unknown-key detector (UnknownConfigKeys, which
// must not call a key it honours "unknown"), and by the offline audit (which
// reports it as DEPRECATED — a standing migration list rather than silence).

package datahelpers

import (
	"sort"

	"go.uber.org/zap"
)

// ResolveConfigSetting reads a literal string setting from a step's config,
// falling back through the spec's declared DeprecatedConfigKeys aliases and then
// to defaultVal.
//
// Precedence — canonical key, then declared old names in sorted order, then the
// default. This is byte-for-byte the truth table of the hand-rolled shim it
// replaces (create_work_item_action.go, "backwards compat"): the new name wins
// when both are set, the old name works when it is alone, the default applies
// when neither is.
//
// A NON-STRING value under any name is treated as absent. Every call site this
// replaces read `config[k].(string)` and dropped anything else on the floor;
// coercing here would be a behaviour change smuggled in under a refactor. A
// numeric setting that needs an alias should get its own typed helper, written
// when something actually needs one.
//
// The second return value names the deprecated key that supplied the value, or
// "" when the canonical key or the default did. Callers rarely need it; it
// exists so a test can assert WHICH path was taken rather than only that the
// answer was right, and those are different questions when the alias and the
// canonical key hold the same string.
//
// The warning fires here, at the moment an alias actually supplies a value —
// the symmetric twin of ExtractActionInputs' own Strategy 3 warn. It is
// therefore scoped to steps that really execute; the offline audit is what
// reports the ones that merely exist.
func ResolveConfigSetting(
	config map[string]interface{},
	spec ActionInputSpec,
	key string,
	defaultVal string,
	logger *zap.Logger,
) (value string, viaDeprecatedKey string) {

	if v, ok := config[key].(string); ok && v != "" {
		return v, ""
	}

	// Sorted, so a spec that (wrongly) declares two aliases for one canonical
	// key resolves deterministically rather than by map iteration order. The
	// spec-parity test rejects that shape, but a deterministic wrong answer is
	// debuggable and a random one is not.
	olds := make([]string, 0, len(spec.DeprecatedConfigKeys))
	for old, canonical := range spec.DeprecatedConfigKeys {
		if canonical == key {
			olds = append(olds, old)
		}
	}
	sort.Strings(olds)

	for _, old := range olds {
		v, ok := config[old].(string)
		if !ok || v == "" {
			continue
		}
		if logger != nil {
			logger.Warn("config setting arrived via a deprecated alias",
				zap.String("deprecated_key", old),
				zap.String("canonical_key", key),
				zap.String("value", v),
				zap.String("action_required", "rename the key in the agent definition; the alias is honoured but reported"),
			)
		}
		return v, old
	}

	return defaultVal, ""
}

// DeprecatedConfigKeysInUse returns the alias keys PRESENT in this step's
// config, mapped to the canonical key each supplies.
//
// `declared` distinguishes "this action declares aliases and none are present"
// from "this action declares none, so nothing was examined" — the same
// load-bearing distinction UnknownConfigKeys draws, and for the same reason: two
// empty results that mean opposite things are how a quiet report comes to read
// as a pass.
//
// Deliberately independent of checksConfig(). create_work_item has not opted
// into unknown-key detection and carries the alias on nine live steps; gating
// this on an unrelated opt-in would make those nine invisible.
func DeprecatedConfigKeysInUse(actionName string, config map[string]interface{}) (inUse map[string]string, declared bool) {
	spec, ok := GetActionInputSpec(actionName)
	if !ok || len(spec.DeprecatedConfigKeys) == 0 {
		return nil, false
	}

	inUse = make(map[string]string)
	for old, canonical := range spec.DeprecatedConfigKeys {
		if v, ok := config[old].(string); ok && v != "" {
			inUse[old] = canonical
		}
	}
	return inUse, true
}

// ListDeprecatedConfigKeys returns every declared literal-setting alias, by
// action. Consumed by cmd/config-key-audit so the offline report can name them
// in their own section — recognised, honoured, and still worth renaming.
func ListDeprecatedConfigKeys() map[string]map[string]string {
	out := make(map[string]map[string]string)
	for name, spec := range actionInputSpecs {
		if len(spec.DeprecatedConfigKeys) == 0 {
			continue
		}
		aliases := make(map[string]string, len(spec.DeprecatedConfigKeys))
		for old, canonical := range spec.DeprecatedConfigKeys {
			aliases[old] = canonical
		}
		out[name] = aliases
	}
	return out
}
