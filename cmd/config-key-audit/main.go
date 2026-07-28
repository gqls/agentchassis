// FILE: cmd/config-key-audit/main.go
//
// Dumps the config-key contract every action has DECLARED, as JSON.
//
// This exists because of bugs_open/101: an unknown step-config key is silently
// ignored at execution, so a stale or aspirational key is indistinguishable by
// inspection from a live one. The runtime validator reports unknown keys for
// actions that have opted in (ActionInputSpec.ConfigKeys), but that only covers
// steps that actually RUN, and only for actions someone has already declared.
//
// The declarations are compiled into the binary — they are Go, registered by
// init() — so the only honest way to compare them against what live
// agent_definitions carry is to ask the binary. That is all this does; the join
// against the database is scripts/audit-config-keys.sh.
//
// Usage:
//
//	go run ./cmd/config-key-audit            # {"action": ["key", ...], ...}
//
// Deliberately imports the actions package for its registration side effects and
// nothing else. If that import is ever dropped the output becomes an empty
// object, which would read as "nothing is declared" rather than failing — so the
// tool refuses to print an empty result instead.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	// Imported for its init() side effects: every action registers its
	// ActionInputSpec here.
	_ "github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func main() {
	declared := datahelpers.ListDeclaredConfigKeys()
	conditional := datahelpers.ListConditionalConfigKeys()

	if len(declared) == 0 {
		fmt.Fprintln(os.Stderr,
			"config-key-audit: no action declares ConfigKeys.\n"+
				"That is either true (nothing has opted in yet) or the actions package "+
				"is no longer linked in, in which case this tool is silently reporting "+
				"nothing rather than failing. Check the blank import above before "+
				"believing an empty result.")
		os.Exit(2)
	}

	// Two maps, not one: "every key this action recognises" and "of those, the
	// ones honoured only under a condition". Merging them would recreate exactly
	// the blindness this second map was added to fix.
	out := map[string]interface{}{
		"declared":    declared,
		"conditional": conditional,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit: %v\n", err)
		os.Exit(1)
	}
}
