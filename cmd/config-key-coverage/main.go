// FILE: cmd/config-key-coverage/main.go
//
// Dumps the FULL ActionInputSpec of every registered action as JSON — not just
// the ConfigKeys that cmd/config-key-audit reports.
//
// Why both exist. config-key-audit answers "what has opted in?", because
// UnknownConfigKeys gates on `len(spec.ConfigKeys) == 0` and an action with an
// empty ConfigKeys is not checked at all. This tool answers the different and
// more useful adoption question: "which actions ALREADY enumerate their keys in
// Required/Optional and are therefore one line away from opting in?"
//
// That distinction is the whole cost model of the coverage ratchet. Declaring a
// key an action does not read is worse than declaring nothing — it silences the
// detector for a dead key (WRONG_CALLS.md 2026-07-28) — so the expensive part of
// adoption is proving what each action reads. For an action whose spec already
// lists its keys, that proof was done when the spec was written, and opting in
// costs one line with no new claim about behaviour.
//
// Usage:
//
//	go run ./cmd/config-key-coverage
//	  {"<action>": {"required": [...], "optional": [...],
//	                "config_keys": [...], "deprecated": [...],
//	                "opted_in": bool}, ...}
//
// Join it against live agent_definitions with scripts/audit-config-keys.sh
// --json, which is where the database half lives. Same reason as its sibling:
// the declarations are Go, registered by init(), so the binary is the only
// honest source for them.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	// Imported for its init() side effects: every action registers its
	// ActionInputSpec here. If this import is dropped the output silently
	// becomes an empty object, so the tool refuses to print one.
	_ "github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

type specDump struct {
	Required   []string `json:"required"`
	Optional   []string `json:"optional"`
	ConfigKeys []string `json:"config_keys"`
	Deprecated []string `json:"deprecated"`
	OptedIn    bool     `json:"opted_in"`
}

func main() {
	names := datahelpers.ListActionInputSpecNames()
	if len(names) == 0 {
		fmt.Fprintln(os.Stderr,
			"config-key-coverage: no action registered an ActionInputSpec.\n"+
				"That is either true or the actions package is no longer linked in, "+
				"in which case this tool would report an empty gap rather than failing. "+
				"Check the blank import before believing this.")
		os.Exit(2)
	}

	out := make(map[string]specDump, len(names))
	for _, name := range names {
		spec, ok := datahelpers.GetActionInputSpec(name)
		if !ok {
			continue
		}
		dep := make([]string, 0, len(spec.Deprecated))
		for k := range spec.Deprecated {
			dep = append(dep, k)
		}
		sort.Strings(dep)
		out[name] = specDump{
			Required:   nonNil(spec.Required),
			Optional:   nonNil(spec.Optional),
			ConfigKeys: nonNil(spec.ConfigKeys),
			Deprecated: dep,
			OptedIn:    len(spec.ConfigKeys) > 0,
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-coverage: %v\n", err)
		os.Exit(1)
	}
}

// nonNil keeps the JSON shape stable: a missing list encodes as [] rather than
// null, so a consumer can iterate without a nil check and cannot mistake
// "declares nothing" for "key absent from the dump".
func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
