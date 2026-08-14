// FILE: cmd/config-key-audit/defaultshadow.go
//
// --default-shadowed-keys: which live step-config entries can never take effect
// because the action's ActionInputSpec carries a Default for the same field?
//
// bugs_open/231 is why this exists, and THIS FILE WAS RE-SPECIFIED 2026-08-13
// when the resolver changed under it. Read the history, because the two most
// populous classes inverted their meaning:
//
// BEFORE: ExtractActionInputs applied spec.Defaults FIRST and every later
// strategy skipped a field that already held a value, so against a defaulted
// field only a Strategy-0 dotted path that resolved could ever win. A static
// string was dead, a numeric or boolean literal was dead, and the deprecated
// *_field bridge was dead. The proven damage: pageflow-builder and
// site-work-orchestrator each carried `"purpose": "logo"` on deploy_image_asset,
// whose spec defaults purpose to "hero", so the logo step's effective purpose
// was "hero" for months.
//
// AFTER (candidate 2, owner ruling 2026-08-11 #2 — "an explicit config value
// beats a default" is now the resolver's rule): Strategy 6 takes a dotless
// config scalar as a LITERAL for a field still holding only its Default, and the
// Strategy 3 bridge may beat a Default too. So `static_string` and
// `non_string_literal` — 99 of the 195 live findings this mode reported on
// 2026-08-11 — are no longer dead keys at all. They are working config, and
// reporting them as damage would be this checker lying about the resolver.
//
// THE CHECKER FOLLOWS THE RESOLVER, NEVER THE REVERSE. It shares the resolver's
// own datahelpers.LiteralKind rather than restating the kind rule, because a
// mirrored rule is the drift this binary exists to catch in others.
//
// Finding classes, by the exact resolver arm that decides the key:
//
//	live_override       — a dotless scalar whose kind matches the Default's:
//	                      Strategy 6 applies it and clears provenance. NOT a
//	                      defect. Reported so the census can enumerate which
//	                      entries the resolver is now honouring — the set that
//	                      changed meaning on this commit, and the set a future
//	                      session re-checks to prove it still does.
//	unextractable_field — the field has a Default but is in neither Required nor
//	                      Optional: no strategy iterates it, Strategy 6 included
//	                      (they all walk Required+Optional), so NO config shape
//	                      can ever touch it. The Default is a constant and the
//	                      config entry is documentation. STILL DEAD.
//	type_mismatch       — a scalar whose kind differs from the Default's kind
//	                      (`max_pages: "60"` against default 25, or any scalar
//	                      against a composite Default). Strategy 6's kind guard
//	                      refuses it and the Default stands, so a config typo
//	                      cannot hand an action a type its spec promised it would
//	                      never see. STILL DEAD, and now the main dead class.
//	required_empty_string — an explicit "" for a field that is BOTH Required and
//	                      Defaulted. Strategy 6 refuses it: a required field's
//	                      Default is the only thing keeping it satisfiable. No
//	                      live spec has that overlap (measured 2026-08-13, 164
//	                      specs), so this class exists to keep the checker honest
//	                      if one ever does. STILL DEAD.
//	composite_literal   — an object or array: Strategy 6 takes scalars only (see
//	                      LiteralKind), so this is dead with or without the
//	                      Default; reported here because the Default is what
//	                      guarantees the field silently gets a value instead.
//	                      STILL DEAD.
//	deprecated_bridge   — spec.Deprecated maps this config key onto a defaulted
//	                      field. NO LONGER DEAD: Strategy 3's has-value skip now
//	                      ignores a value that is still only the Default, so the
//	                      bridge resolves its path and wins. Conditional on that
//	                      path resolving, exactly like dotted_conditional, and
//	                      counted with it. Pinned by
//	                      TestPurposeFieldBridge_BeatsTheDefault in the actions
//	                      package.
//	dotted_conditional  — a dotted path bound to a defaulted field. NOT dead:
//	                      Strategy 0 resolves it and overwrites the Default —
//	                      but only when the path resolves against the dispatch
//	                      shape that actually arrives. When it resolves to
//	                      nothing the Default wins SILENTLY (bugs_open/231's
//	                      second face: asset-deployer's `input_data.purpose`
//	                      resolved nothing on the undeployed_asset dispatch
//	                      shape, so every logo deployed as a hero). A dotted
//	                      string is still never read as a literal — bugs_open/248
//	                      finding (a) is what taking a path expression for a value
//	                      costs. Reported so the census can enumerate the
//	                      exposure; never exit 1, because resolvability is a
//	                      runtime fact this offline check cannot decide.
//
// matches_default: whether the config value happens to EQUAL the default. For a
// dead class a match means behaviour coincides with the author's literal intent
// today, and only a MISMATCHED dead key is behaviour silently differing from what
// the config says — so only those drive exit 1. For live_override a match means
// the entry is simply redundant.
//
// CAVEAT the consumer must hold: "dead" here means dead ON THE INPUTS PATH. An
// action that additionally reads step.Config directly in its own Run body can
// still honour the key through that private read. bugs_open/235's static
// purpose was exactly such a case — honoured, and wrong. Check the flagged
// action for direct config reads before asserting live damage.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/validation"
)

// defaultShadowFinding names one live step-config entry that a spec Default
// shadows (or, for dotted_conditional, can silently shadow at runtime).
type defaultShadowFinding struct {
	Agent  string `json:"agent"`
	Path   string `json:"path"`
	Action string `json:"action"`
	// Field is the spec field whose Default wins; Key is the config key
	// carrying the dead value. They differ only for deprecated_bridge, where
	// Key is the old alias.
	Field          string      `json:"field"`
	Key            string      `json:"key"`
	Class          string      `json:"class"`
	ConfigValue    interface{} `json:"config_value"`
	DefaultValue   interface{} `json:"default_value"`
	MatchesDefault bool        `json:"matches_default"`
	Nested         bool        `json:"nested"`
	// Verdict is the emitted form of dead()/conditional: "dead", "conditional"
	// or "live". It exists so CONSUMERS NEVER RE-DERIVE THE RULE FROM THE CLASS
	// NAME. scripts/audit-default-shadowed-keys.sh used to compute dead as
	// `class != "dotted_conditional"`, a second copy of the rule below that this
	// re-spec would have silently falsified — it would have printed the 99 newly
	// live overrides as dead keys. Same defect class this binary reports on.
	Verdict string `json:"verdict"`
}

// Classes emitted by findDefaultShadowedKeys.
const (
	classLiveOverride      = "live_override"
	classUnextractable     = "unextractable_field"
	classTypeMismatch      = "type_mismatch"
	classRequiredEmpty     = "required_empty_string"
	classComposite         = "composite_literal"
	classDeprecatedBridge  = "deprecated_bridge"
	classDottedConditional = "dotted_conditional"
)

// shadowClassVerdict is the ONE definition of what each class means for the exit
// rule, the summary and the printed report:
//
//	"dead"        — the config entry can never take effect; the Default always wins.
//	"conditional" — it wins if and only if its path resolves at runtime, which an
//	                offline check cannot decide.
//	"live"        — the resolver honours it.
//
// A class missing from this map is a classifier change that forgot the exit rule;
// dead() treats it as dead so it is reported rather than hidden, and
// TestEveryClassHasAVerdict makes that unreachable.
var shadowClassVerdict = map[string]string{
	classLiveOverride:      "live",
	classUnextractable:     "dead",
	classTypeMismatch:      "dead",
	classRequiredEmpty:     "dead",
	classComposite:         "dead",
	classDeprecatedBridge:  "conditional",
	classDottedConditional: "conditional",
}

func verdictFor(class string) string {
	if v, ok := shadowClassVerdict[class]; ok {
		return v
	}
	return "dead"
}

// dead reports whether this finding's config entry can never take effect on the
// inputs path.
func (f defaultShadowFinding) dead() bool {
	return verdictFor(f.Class) == "dead"
}

// registeredSpecs snapshots the registry into a plain map so the pure check is
// testable with synthetic specs while the emit path asks the real binary.
func registeredSpecs() map[string]datahelpers.ActionInputSpec {
	names := datahelpers.ListActionInputSpecNames()
	out := make(map[string]datahelpers.ActionInputSpec, len(names))
	for _, name := range names {
		if spec, ok := datahelpers.GetActionInputSpec(name); ok {
			out[name] = spec
		}
	}
	return out
}

// literalMatchesDefault compares a config literal against the spec default it
// is shadowed by. JSON numbers decode as float64 while Go defaults are often
// typed ints, so numeric values are compared as floats; everything else by
// deep equality.
func literalMatchesDefault(configValue, defaultValue interface{}) bool {
	if cf, ok := toFloat64(configValue); ok {
		df, ok2 := toFloat64(defaultValue)
		return ok2 && cf == df
	}
	return reflect.DeepEqual(configValue, defaultValue)
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}

// findDefaultShadowedKeys is the pure check, separated from I/O so fixture
// tests can exercise it (see findUnregisteredActions). It walks with
// validation.WalkSteps — the SAME traversal the runtime validator enforces
// against, top-level and nested — for the bugs_open/144 reason recorded there.
//
// The classification MUST mirror ExtractActionInputs' arms exactly — the
// resolver is production and this checker follows it, never the reverse. The
// load-bearing details it copies: Strategy 0's dot test is
// strings.Contains(value, "."), strategies 0/4/5 iterate Required+Optional
// only, and every skip is "already present in Values", which a Default
// guarantees.
func findDefaultShadowedKeys(agents []liveAgent, specs map[string]datahelpers.ActionInputSpec) []defaultShadowFinding {
	findings := []defaultShadowFinding{}
	for _, agent := range agents {
		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			spec, ok := specs[step.Action]
			if !ok || len(spec.Defaults) == 0 {
				return
			}

			required := make(map[string]bool, len(spec.Required))
			for _, f := range spec.Required {
				required[f] = true
			}
			extractable := make(map[string]bool, len(spec.Required)+len(spec.Optional))
			for _, f := range spec.Required {
				extractable[f] = true
			}
			for _, f := range spec.Optional {
				extractable[f] = true
			}

			for field, defaultValue := range spec.Defaults {
				raw, present := step.Config[field]
				if !present || raw == nil {
					continue
				}

				// The arms below are in the SAME ORDER as Strategy 6's, and that
				// ordering is load-bearing: unextractable is tested before the
				// value's shape because no strategy iterates the field at all, and
				// the dot test comes before the kind guard because a dotted string
				// is a reference the guard never sees.
				str, isString := raw.(string)
				kind := datahelpers.LiteralKind(raw)
				var class string
				switch {
				case !extractable[field]:
					class = classUnextractable
				case isString && strings.Contains(str, "."):
					class = classDottedConditional
				case kind == "":
					class = classComposite
				case datahelpers.LiteralKind(defaultValue) != kind:
					class = classTypeMismatch
				case isString && str == "" && required[field]:
					class = classRequiredEmpty
				default:
					class = classLiveOverride
				}

				findings = append(findings, defaultShadowFinding{
					Agent: agent.Type, Path: path, Action: step.Action,
					Field: field, Key: field, Class: class,
					ConfigValue: raw, DefaultValue: defaultValue,
					MatchesDefault: literalMatchesDefault(raw, defaultValue),
					Nested:         nested,
					Verdict:        verdictFor(class),
				})
			}

			// The deprecated bridge is a separate config key, checked over the
			// alias map rather than the field name. Only a non-empty string is
			// ever read by Strategy 3, so only that shape is a finding — and since
			// 2026-08-13 it is a CONDITIONAL one: the bridge resolves its path
			// against collected_data and now beats the Default when it does.
			for oldKey, field := range spec.Deprecated {
				if _, hasDefault := spec.Defaults[field]; !hasDefault {
					continue
				}
				pathStr, isString := step.Config[oldKey].(string)
				if !isString || pathStr == "" {
					continue
				}
				findings = append(findings, defaultShadowFinding{
					Agent: agent.Type, Path: path, Action: step.Action,
					Field: field, Key: oldKey, Class: classDeprecatedBridge,
					ConfigValue: pathStr, DefaultValue: spec.Defaults[field],
					MatchesDefault: literalMatchesDefault(pathStr, spec.Defaults[field]),
					Nested:         nested,
					Verdict:        verdictFor(classDeprecatedBridge),
				})
			}
		})
	}
	// Map iteration over Defaults and step.Config is random, so the sort is
	// what makes the report diffable between runs.
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Agent != findings[j].Agent {
			return findings[i].Agent < findings[j].Agent
		}
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		return findings[i].Key < findings[j].Key
	})
	return findings
}

// emitDefaultShadowedKeys reads the same stdin shape as --live-pairs and
// prints every finding as JSON. Zero findings is a legitimate, meaningful
// result; an empty decoded-agent set is not and is refused, exactly as on
// emitUnregisteredActions. An empty spec registry is refused too — a dropped
// blank import must fail loudly, not report a clean fleet.
//
// Exit 1 only when a DEAD finding's value MISMATCHES its default: that is the
// set where live behaviour silently differs from what the config says. Matched
// dead keys, conditional bindings and live overrides are reported for the census
// but do not fail the run.
func emitDefaultShadowedKeys() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --default-shadowed-keys: reading stdin: %v\n", err)
		os.Exit(2)
	}

	agents, failed, err := decodeLiveAgents(raw, "--default-shadowed-keys")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --default-shadowed-keys: %v\n", err)
		os.Exit(2)
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --default-shadowed-keys: 0 agents decoded (%d undecodable) — "+
				"refusing to print a clean report over an empty or broken export.\n", failed)
		os.Exit(2)
	}

	specs := registeredSpecs()
	withDefaults := 0
	for _, spec := range specs {
		if len(spec.Defaults) > 0 {
			withDefaults++
		}
	}
	if withDefaults == 0 {
		fmt.Fprintln(os.Stderr,
			"config-key-audit --default-shadowed-keys: no registered spec carries Defaults.\n"+
				"That is either true or the actions package is no longer linked in, in which "+
				"case this would report a clean fleet rather than failing. Check the blank "+
				"import in main.go before believing it.")
		os.Exit(2)
	}

	findings := findDefaultShadowedKeys(agents, specs)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(findings); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --default-shadowed-keys: %v\n", err)
		os.Exit(1)
	}

	deadMismatched, deadMatched, conditional, live := 0, 0, 0, 0
	for _, f := range findings {
		switch f.Verdict {
		case "live":
			live++
		case "conditional":
			conditional++
		default:
			if f.MatchesDefault {
				deadMatched++
			} else {
				deadMismatched++
			}
		}
	}
	fmt.Fprintf(os.Stderr,
		"config-key-audit --default-shadowed-keys: %d agents decoded (%d undecodable), "+
			"%d specs with Defaults; %d dead mismatched, %d dead matching, "+
			"%d conditional (dotted paths + deprecated bridges), %d live overrides\n",
		len(agents), failed, withDefaults, deadMismatched, deadMatched, conditional, live)
	if deadMismatched > 0 {
		os.Exit(1)
	}
}
