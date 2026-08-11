// FILE: cmd/config-key-audit/defaultshadow.go
//
// --default-shadowed-keys: which live step-config entries can never take effect
// because the action's ActionInputSpec carries a Default for the same field?
//
// bugs_open/231 is why this exists. ExtractActionInputs applies spec.Defaults
// FIRST (action_inputs.go, "Apply defaults first"), and every later strategy
// except one skips a field that already has a value. The single exception is
// Strategy 0, which only reads a config value that is a MULTI-SEGMENT dotted
// string and only wins if that path actually resolves. So against a defaulted
// field, a static config value is dead, a numeric or boolean literal is dead
// (Strategy 5 skips populated fields), the deprecated *_field bridge is dead
// (Strategy 3 skips populated fields), and a defaulted field that is not in
// Required or Optional cannot be touched by ANY config shape, because
// strategies 0, 4 and 5 iterate Required+Optional only.
//
// The proven damage: pageflow-builder and site-work-orchestrator each carried
// `"purpose": "logo"` on deploy_image_asset, whose spec defaults purpose to
// "hero" — the logo step's effective purpose was "hero" for months (repaired by
// migration 348, but only for those two definitions; this check is the durable
// half, because the next author to write a static for a defaulted field would
// otherwise re-create the bug with nothing reporting it).
//
// Finding classes, by the exact resolver arm that kills the key:
//
//	unextractable_field — the field has a Default but is in neither Required nor
//	                      Optional: no strategy iterates it, so NO config shape
//	                      (dotted or not) can ever touch it. The Default is a
//	                      constant and the config entry is documentation.
//	static_string       — a string without ".": invisible to Strategy 0, and
//	                      every strategy that could read it skips the populated
//	                      field. (For a NON-defaulted field the same string
//	                      would be a live single-segment reference — Strategy 4
//	                      — which is why this check is gated on Defaults.)
//	non_string_literal  — a bool or number: Strategy 5 takes literals only for
//	                      fields that are still empty, and a defaulted field
//	                      never is.
//	composite_literal   — an object or array: no strategy reads composites at
//	                      all (recorded on Strategy 5), so this is dead with or
//	                      without the Default; reported here because the Default
//	                      is what guarantees the field silently gets a value.
//	deprecated_bridge   — spec.Deprecated maps this config key onto a defaulted
//	                      field: Strategy 3 checks "already has a value" before
//	                      reading the path, so the bridge can never fire
//	                      (pinned by TestPurposeFieldBridge_DeadForDefaultedField
//	                      in the actions package).
//	dotted_conditional  — a dotted path bound to a defaulted field. NOT dead:
//	                      Strategy 0 resolves it and overwrites the Default —
//	                      but only when the path resolves against the dispatch
//	                      shape that actually arrives. When it resolves to
//	                      nothing the Default wins SILENTLY (bugs_open/231's
//	                      second face: asset-deployer's `input_data.purpose`
//	                      resolved nothing on the undeployed_asset dispatch
//	                      shape, so every logo deployed as a hero). Reported so
//	                      the census can enumerate the exposure; never exit 1,
//	                      because resolvability is a runtime fact this offline
//	                      check cannot decide.
//
// matches_default: for the literal classes, whether the dead config value
// happens to EQUAL the default it is shadowed by. A match means behaviour
// coincides with the author's literal intent today (deploy_hero_image's static
// "hero" against default "hero" — invisible, but the first edit to that value
// changes nothing and says nothing). Only a MISMATCHED dead key is behaviour
// silently differing from what the config says, so only those drive exit 1.
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
}

// dead reports whether this finding's config entry can never take effect on
// the inputs path. ONE definition, shared by the exit rule and the summary —
// the class list and its consumers disagreeing about which classes are dead is
// exactly the drift this binary exists to catch in others.
func (f defaultShadowFinding) dead() bool {
	return f.Class != "dotted_conditional"
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

				var class string
				switch {
				case !extractable[field]:
					class = "unextractable_field"
				default:
					switch v := raw.(type) {
					case string:
						if strings.Contains(v, ".") {
							class = "dotted_conditional"
						} else {
							class = "static_string"
						}
					case bool, float64, float32,
						int, int8, int16, int32, int64,
						uint, uint8, uint16, uint32, uint64,
						json.Number:
						class = "non_string_literal"
					default:
						class = "composite_literal"
					}
				}

				findings = append(findings, defaultShadowFinding{
					Agent: agent.Type, Path: path, Action: step.Action,
					Field: field, Key: field, Class: class,
					ConfigValue: raw, DefaultValue: defaultValue,
					MatchesDefault: literalMatchesDefault(raw, defaultValue),
					Nested:         nested,
				})
			}

			// The deprecated bridge is a separate config key, checked over the
			// alias map rather than the field name. Only a non-empty string
			// would ever have been read by Strategy 3, so only that shape is a
			// "would have worked but for the Default" finding.
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
					Field: field, Key: oldKey, Class: "deprecated_bridge",
					ConfigValue: pathStr, DefaultValue: spec.Defaults[field],
					MatchesDefault: literalMatchesDefault(pathStr, spec.Defaults[field]),
					Nested:         nested,
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
// dead keys and dotted_conditional bindings are reported for the census but do
// not fail the run.
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

	deadMismatched, deadMatched, conditional := 0, 0, 0
	for _, f := range findings {
		switch {
		case !f.dead():
			conditional++
		case f.MatchesDefault:
			deadMatched++
		default:
			deadMismatched++
		}
	}
	fmt.Fprintf(os.Stderr,
		"config-key-audit --default-shadowed-keys: %d agents decoded (%d undecodable), "+
			"%d specs with Defaults; %d dead mismatched, %d dead matching, %d conditional dotted\n",
		len(agents), failed, withDefaults, deadMismatched, deadMatched, conditional)
	if deadMismatched > 0 {
		os.Exit(1)
	}
}
