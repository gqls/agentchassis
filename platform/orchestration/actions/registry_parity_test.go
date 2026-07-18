// FILE: platform/orchestration/actions/registry_parity_test.go
//
// Guards the drift class that produced bugs_open/017: an action can be fully
// written — handler, ActionInputSpec, the lot — and still never be added to
// GlobalActionRegistry. Nothing complains at build time. The workflow validator
// asks the registry whether the action is local (validation/workflow.go:69),
// gets "no", concludes it must be a remote action, and rejects every workflow
// naming it with the misleading "requires a topic".
//
// fix_forced_text_colors sat like that and failed every color-variable-fixer
// dispatch for months; 49 work items were stamped 'complete' carrying the
// WORKFLOW_INVALID that proves nothing ran.
//
// Registering an ActionInputSpec is the closest thing to a declaration of
// intent an action file makes, so spec-without-registry-entry is the signal.

package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// dormantActions are written but deliberately NOT registered: no seeded
// workflow dispatches them, and registering an untested action would make a
// future seed silently run it rather than fail loudly. Each entry is a
// decision, not an oversight — that is the point of listing them here.
// Verified unseeded against agent_definitions on 2026-07-18.
var dormantActions = map[string]string{
	"cleanup_stale_topics":       "Kafka topic GC; no seeded workflow dispatches it",
	"load_pending_verifications": "business-verification loop; no seeded workflow dispatches it",
}

func TestEveryActionInputSpecHasARegistryEntry(t *testing.T) {
	for _, name := range datahelpers.ListActionInputSpecNames() {
		if _, registered := GlobalActionRegistry[name]; registered {
			if why, dormant := dormantActions[name]; dormant {
				t.Errorf("action %q is registered but still listed as dormant (%s) — remove it from dormantActions", name, why)
			}
			continue
		}
		if _, dormant := dormantActions[name]; dormant {
			continue
		}
		t.Errorf("action %q registers an ActionInputSpec but has no GlobalActionRegistry entry.\n"+
			"The workflow validator will treat it as remote and reject every workflow using it with\n"+
			"\"step '<step>' with action '%s' requires a topic\". Add an entry to registry.go,\n"+
			"or add it to dormantActions with the reason it is deliberately unreachable.", name, name)
	}
}

// The bug that started it: fix_forced_text_colors must be local and reachable.
func TestFixForcedTextColorsIsRegisteredLocal(t *testing.T) {
	def, ok := GlobalActionRegistry["fix_forced_text_colors"]
	if !ok {
		t.Fatal("fix_forced_text_colors is not registered — color-variable-fixer workflows will fail WORKFLOW_INVALID")
	}
	if !def.IsLocal {
		t.Error("fix_forced_text_colors must be IsLocal: it runs in-process and its workflow step carries no topic")
	}
	if def.Handler == nil {
		t.Error("fix_forced_text_colors has no handler")
	}
}
