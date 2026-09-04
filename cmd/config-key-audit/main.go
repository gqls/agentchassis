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
//	go run ./cmd/config-key-audit
//	  {"declared":    {"<action>": ["<key>", ...], ...},
//	   "conditional": {"<action>": {"<key>": "<the condition>", ...}, ...}}
//
//	go run ./cmd/config-key-audit --specs
//	  {"<action>": {"required": [...], "optional": [...], "config_keys": [...],
//	                "deprecated": [...], "opted_in": bool}, ...}
//
//	go run ./cmd/config-key-audit --live-declaration-drift [--report]
//	  Does every LIVE database object still contain what platform/livespec
//	  declares? Reads the objects directly (scheduled_tasks.pre_query,
//	  pg_get_functiondef, pg_trigger counts) and compares them to the declaration
//	  the Go guards assert against. Exit 1 on drift; exit 2 if it could not LOOK
//	  (no DB, unreadable probe, NULL or missing row, empty registry) — never a
//	  clean report. --report writes ONE doc_notes row per run INCLUDING clean
//	  ones, carrying the scope, so a MISSING row means the job did not run.
//	  (bugs_open/363 phase 2. Phase 1 tied the guards to the declaration; this
//	  ties the declaration to production. It cannot be a unit test — go test has
//	  no cluster — nor a pre-commit hook, because at commit time the migration is
//	  unapplied, which is the RFC_006 ruling.)
//
//	go run ./cmd/config-key-audit --budget-placement [--report] < live-workflows-with-agent-config.json
//	  {"agents_scanned": N, "steps_scanned": N, "declarations": N,
//	   "findings": [{"agent": "...", "path": "...", "kind": "shadowed|non_canonical|unconfigured",
//	     "effective": 16000, "from": "root.ai_service", "declarations": [...], "detail": "..."}]}
//	  Where does each live step's output-token budget actually come from? Calls
//	  production's own ladder (actions.ResolveStepBudget) rather than carrying a
//	  copy of the precedence rule, so it cannot drift from what the pods do.
//	  UNCONFIGURED = declares an ai_service block and no budget at any level, so it
//	  runs at the 2048 provider floor (bugs_open/205). SHADOWED = two levels declare
//	  DIFFERENT numbers, so an operator's number is inert (bugs_open/257 round 3).
//	  NON_CANONICAL = declared outside an ai_service block: honoured, advisory, and
//	  never fails the run — it is where both round-3 failures started. Needs the
//	  `agent_config` projection and REFUSES an export without it. Exit 1 on
//	  shadowed/unconfigured only; exit 2 if it could not LOOK.
//
//	go run ./cmd/config-key-audit --single-owner-actions   < live-workflows.json
//	  [{"action": "...", "owners": ["<agent>", ...], "paths": [...]}, ...]
//	  Which action declared SingleOwner is carried by more than one live agent?
//	  (RFC 006 / bugs_closed/150 — a site-wide promoter with three callers made
//	  the improvement loop call a busy site clean. Exit 1 on findings.)
//
//	go run ./cmd/config-key-audit --optional-key-budget [N] [--acks <file>] < live-workflows.json
//	  {"budget": N|null, "actions": [{"action": "...", "optional_keys": K,
//	    "optional": [...], "consumers": M, "agents": [...],
//	    "acknowledged": A, "stale_ack": bool, "over_budget": bool}, ...]}
//	  Which SHARED action's optional-key set has accumulated past the budget?
//	  (RFC 022, budget RULED 2026-08-14: N=10 — the trigger moved from "any new
//	  opt-in field" to the COUNT: the accumulation is what gets reviewed, never
//	  the reuse, which is estate design.) Shared = >= 2 distinct live agents
//	  carry it. --acks names the reviewed-baselines file: a flagged action owes
//	  ONE review of its surface, then its acknowledged level is the baseline and
//	  only growth PAST it pages. Without N: report-only census, exit 0.
//	  Exit 1 on findings.
//
//	go run ./cmd/config-key-audit --unarmed-verified-completers < live-workflows.json
//	  [{"kind": "undeclared|stale", "agent": "...", "step": "...", "path": "...",
//	    "item_type": "...", "detail": "..."}, ...]
//	  Does livespec.UnarmedVerifiedCompleters still match production? An
//	  update_work_item_status step that stamps `complete` without setting
//	  verify_before_complete skips the item type's verifier entirely (bugs_open/375,
//	  WII-030). The build-time lockstep can only check the DECLARED arms; this is
//	  what notices a new one. ⚠ `status` DEFAULTS to `complete`, so a step omitting
//	  it IS a complete arm. Exit 1 on findings; exit 2 if it could not LOOK.
//
//	go run ./cmd/config-key-audit --suspicious-keys        < live-workflows.json
//	  [{"agent": "...", "path": "...", "action": "...", "key": "...", "nested": bool}, ...]
//	  Which live step-config key NAME carries schema-documentation punctuation
//	  ("?", "*", ":" or a space)? bugs_open/134: `category?` leaked out of a seed's
//	  own comment into two real key names, which then resolved to nothing and said
//	  nothing. Asked of EVERY action, opted in or not. Exit 1 on findings.
//
//	go run ./cmd/config-key-audit --default-shadowed-keys  < live-workflows.json
//	  [{"agent": "...", "path": "...", "action": "...", "field": "...", "key": "...",
//	    "class": "...", "config_value": ..., "default_value": ..., "matches_default": bool, ...}, ...]
//	  Which live step-config entry can never take effect because the action's
//	  spec carries a Default for that field? bugs_open/231: only a Strategy-0
//	  dotted path that RESOLVES can beat a Default; statics, literals and the
//	  deprecated bridge are all silently dead against one, and a defaulted field
//	  outside Required+Optional is dead to every config shape. Dotted paths on
//	  defaulted fields are reported as "dotted_conditional" (they fall back to
//	  the Default silently when unresolvable) but never fail the run. Exit 1
//	  only on a dead entry whose value MISMATCHES the default it is shadowed by.
//	  See defaultshadow.go for the class definitions and the direct-config-read
//	  caveat.
//
//	go run ./cmd/config-key-audit --relay-gaps             < live-workflows.json
//	  {"findings": [...], "uncovered_relays": [...], "unmatched_registry_entries": [...]}
//	  Can each DECLARED dispatcher relay still carry every key its handler's
//	  input_contract names? Checks BOTH allow-lists — the claim query's RETURNING
//	  projection and the call_agent input_mapping — because a dispatcher has two
//	  and fixing only the visible one leaves the key dropped (bugs_open/174).
//	  Needs input_contract in the export. Exit 1 on findings OR on an unmatched
//	  registry entry, which means an assertion stopped running.
//
//	go run ./cmd/config-key-audit --undeclared-recurrence [--report] < live-workflows.json
//	  {"agents_scanned": N, "findings": [{"agent": "...", "path": "...",
//	    "item_type": "...", "item_key_prefix": "...", "declared_unhonoured": false}]}
//	  Which keyed create_work_item step has never declared whether the item it
//	  files is an ACTION REQUEST or a DETECTED DEFECT? The anti-churn brake acts
//	  on the answer; a MISSING declaration is the finding, either explicit value
//	  is clean (bugs_open/326).
//
//	go run ./cmd/config-key-audit --template-input-fields [--report] < live-workflows-with-prompts.json
//	  {"templates_checked": N, "parse_failures": [...], "findings": [{"agent": "...",
//	    "path": "...", "action": "...", "template_tier": "step|agent",
//	    "kind": "unreachable_root|conditional_root|declared_unread", "roots": [...]}]}
//	  Which live prompt template names a variable its step's input_fields can
//	  never supply — so the block renders EMPTY, or "<no value>", with no error
//	  and no verdict (bugs_open/453)? Parses each template with production's own
//	  parser and func map, so a {{range}} body's rebound dot is not mistaken for a
//	  root. Exit 1 ONLY on unreachable_root (input_data absent, so config alone
//	  decides it); conditional_root and declared_unread are advisory. Needs the
//	  agent_prompt_template projection and REFUSES an export without it.
//
//	go run ./cmd/config-key-audit --loop-sitewide-item-keys [--report] < live-workflows.json
//	  {"agents_scanned": N, "findings": [{"agent": "...", "path": "...",
//	    "loop_path": "...", "loop_variable": "...", "item_key_prefix": "...", ...}]}
//	  Which create_work_item step, filed PER ITEM inside a loop, still keys its
//	  item_key PER SITE — so iterations 2..N collide on idx_swi_dedup and are
//	  silently dropped (bugs_open/321: tool-suggester lost ~72% of its
//	  suggestions this way)? Once-per-site steps and loops over SITES are not
//	  findings — the site-wide key is the intended dedupe there. --report writes
//	  one doc_notes row per run, clean runs included (the CronJob route). Exit 1
//	  on findings.
//
// --specs answers the ADOPTION question rather than the coverage one: not "who has
// opted in?" but "who is one line away?". An action whose spec already enumerates
// every key its live steps carry can opt in with `CheckConfig: true` and assert
// nothing new about behaviour; one whose spec is missing keys, or which has no spec
// at all, needs reading first. Declaring a key an action does NOT read is worse than
// declaring nothing — it silences the detector for a dead key (WRONG_CALLS.md
// 2026-07-28) — so that distinction is the whole cost model of the ratchet.
//
// It is a FLAG rather than a second binary because the council gate's editquality
// and reuse_agent seats both objected to the second binary independently
// (2026-07-28, corr 07cf67c6): same registry API, same struct, same init-side-effect
// import, overlapping question. They were right — "two code paths independently
// solving overlapping problems that nobody unifies once both exist" is a named
// failure pattern in this estate. The DEFAULT output is unchanged byte-for-byte, so
// scripts/audit-config-keys.sh is unaffected.
//
// TWO maps, not one merged list. "declared" is every key an action recognises;
// "conditional" is the subset honoured only under a stated condition. Merging
// them would recreate exactly the blindness the second map was added to fix —
// a key can be perfectly well declared and still describe behaviour a given step
// never performs (bugs_closed/101; WRONG_CALLS.md 2026-07-28).
//
// The only consumer of this output is scripts/audit-config-keys.sh — checked, not
// assumed, when the shape changed: `grep -rn "config-key-audit"` over the tree
// returns that script and this file's own comments, nothing else.
//
// Deliberately imports the actions package for its registration side effects and
// nothing else. If that import is ever dropped the output becomes an empty
// object, which would read as "nothing is declared" rather than failing — so the
// tool refuses to print an empty result instead.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/pkg/models"

	// Imported for its init() side effects: every action registers its
	// ActionInputSpec here, AND registers the actioncheck.IsLocalAction checker
	// that --unregistered-actions depends on (registry.go's init calls
	// actioncheck.RegisterLocalActionChecker). Drop this import and both modes
	// go quiet in different ways: ConfigKeys reports empty, IsLocalAction reports
	// "not local" for every action — which is why both refuse to print a
	// clean-looking result over zero registrations.
	"github.com/gqls/agentchassis/platform/orchestration/actioncheck"
	// Blank here because THIS file only needs the init() side effects; Go
	// scopes imports per file, and componentsourcevocabulary.go imports the
	// same package by name to call the birth gate's own
	// SourceVocabularyFindings rather than a second copy of the rule
	// (bugs_open/309, CLC-018). Do not "tidy" this into one named import —
	// main.go would then fail to compile as imported-and-not-used.
	_ "github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/validation"
)

// liveAgent is one row of the export both --live-pairs and --unregistered-actions
// read on stdin: [{"type": "<agent>", "workflow": {...}}, ...]. Shared so the two
// questions ("what (action,key) pairs exist" and "what action is unrunnable") are
// asked of the exact same decode, not two copies that could drift.
type liveAgent struct {
	Type     string              `json:"type"`
	Workflow models.WorkflowPlan `json:"workflow"`
	// AgentPromptTemplate is default_config.prompt_template — tier 2 of
	// getPromptWithPriority, the template a step with no prompt_template of its
	// own actually renders. Read only by --template-input-fields; every other
	// mode ignores it, and every other mode's export omits it entirely.
	//
	// json.RawMessage rather than *string ON PURPOSE: it is the only shape that
	// distinguishes "the export did not project this key" (nil) from "the agent
	// has no prompt_template" (the four bytes `null`), because encoding/json
	// calls RawMessage.UnmarshalJSON even for a JSON null while it would simply
	// leave a *string nil in both cases. --template-input-fields refuses an
	// export where no row carries the key, which is a check it could not make
	// against a pointer. See requireAgentPromptProjection.
	AgentPromptTemplate json.RawMessage `json:"agent_prompt_template"`
	// AgentConfig is default_config MINUS the workflow: the agent-level half of
	// the token-budget ladder (`max_tokens` and `ai_service.max_tokens` at the
	// root). Read only by --budget-placement; every other mode ignores it and
	// every other mode's export omits it entirely.
	//
	// json.RawMessage for the same reason AgentPromptTemplate is: it is the only
	// shape that distinguishes "the export did not project this key" (nil) from
	// "the agent has no root config" (the four bytes `null`). --budget-placement
	// REFUSES an export where no row carries it, because without the agent level
	// the ladder is missing its two lowest rungs and every verdict would be wrong
	// in the same direction — a confidently wrong report, worse than none.
	AgentConfig json.RawMessage `json:"agent_config"`
}

// agentPromptTemplate decodes the tier-2 template, returning "" for an absent
// key, a JSON null, or a non-string value.
func (a liveAgent) agentPromptTemplate() string {
	if len(a.AgentPromptTemplate) == 0 {
		return ""
	}
	var s *string
	if err := json.Unmarshal(a.AgentPromptTemplate, &s); err != nil || s == nil {
		return ""
	}
	return *s
}

// decodeLiveAgents parses the stdin export. One undecodable definition costs that
// definition, not the whole report — see emitLivePairs' original comment on why
// failures are counted and printed rather than swallowed.
func decodeLiveAgents(raw []byte, caller string) (agents []liveAgent, failed int, err error) {
	var rows []json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, 0, fmt.Errorf("stdin is not a JSON array of agents: %w\n"+
			"A truncated kubectl exec exits 0, so a short read arrives here looking like a small fleet.", err)
	}
	for _, row := range rows {
		var agent liveAgent
		if jsonErr := json.Unmarshal(row, &agent); jsonErr != nil {
			failed++
			fmt.Fprintf(os.Stderr, "config-key-audit %s: undecodable agent row: %v\n", caller, jsonErr)
			continue
		}
		agents = append(agents, agent)
	}
	return agents, failed, nil
}

// specDump is the --specs shape: an action's FULL declared contract, not just the
// key set the coverage report joins on.
type specDump struct {
	Required   []string `json:"required"`
	Optional   []string `json:"optional"`
	ConfigKeys []string `json:"config_keys"`
	Deprecated []string `json:"deprecated"`
	// DeprecatedConfigKeys is the LITERAL-setting alias map (old -> canonical),
	// distinct from Deprecated's path-reference aliases. Emitted so --specs keeps
	// answering "who is one line away" honestly: an action carrying an alias has
	// already stated part of its contract (bugs_open/136).
	DeprecatedConfigKeys map[string]string `json:"deprecated_config_keys,omitempty"`
	// RemovedConfigKeys maps a RETIRED key to the message naming its
	// replacement. A live step carrying one fails validation outright once the
	// declaring binary rolls — which is why the offline report must see these
	// (bugs_open/234): the join against live definitions is the only thing that
	// can catch a carrier BEFORE the roll turns it into a dead agent.
	RemovedConfigKeys map[string]string `json:"removed_config_keys,omitempty"`
	// Defaults is the spec's field→default map. Emitted because a Default is
	// load-bearing for what a step config can express at all: against a
	// defaulted field only a resolving dotted path takes effect (bugs_open/231),
	// so a reader sizing a config change needs the defaults in the same dump as
	// the key sets. bugs_open/223's var-blindness means a code search for the
	// declaration can miss it; this asks the binary instead.
	Defaults map[string]interface{} `json:"defaults,omitempty"`
	OptedIn  bool                   `json:"opted_in"`
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--specs" {
		emitSpecs()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--live-pairs" {
		emitLivePairs()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--unregistered-actions" {
		emitUnregisteredActions()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--budget-placement" {
		emitBudgetPlacement(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--optional-key-budget" {
		emitOptionalKeyBudget(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--unarmed-verified-completers" {
		emitUnarmedCompleterDrift()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--single-owner-actions" {
		emitSingleOwnerViolations()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--optional-explicit-wires" {
		emitOptionalExplicitWires(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--finding-codes" {
		emitFindingCodes(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--live-declaration-drift" {
		emitLiveDeclarationDrift(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--commit-sha-exposure" {
		emitCommitShaExposure(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--ungraded-completions" {
		emitUngradedCompletions(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--render-truncation" {
		emitRenderTruncation(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--template-input-fields" {
		emitTemplateInputFields(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--removed-keys-in-use" {
		emitRemovedKeyCarriers()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--suspicious-keys" {
		emitSuspiciousKeys()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--relay-gaps" {
		emitRelayGaps()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--default-shadowed-keys" {
		emitDefaultShadowedKeys()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--shared-output-fields" {
		emitSharedOutputFields()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--array-producer-conditions" {
		emitArrayProducerConditions()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--loop-sitewide-item-keys" {
		emitLoopSitewideItemKeys()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--undeclared-recurrence" {
		emitUndeclaredRecurrence()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--component-source-vocabulary" {
		emitComponentSourceVocabulary(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--capped-schedule-ordering" {
		emitCappedScheduleOrdering(os.Args[2:])
		return
	}
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

	// Three maps, not one: "every key this action recognises", "of those, the
	// ones honoured only under a condition", and "of those, the ones that are an
	// OLD NAME still being honoured". Merging any of them would recreate exactly
	// the blindness each was added to fix — a recognised key can be inert
	// (conditional) or wired-under-a-name-nobody-should-write (deprecated), and a
	// report with one bucket calls all three clean.
	//
	// `deprecated` is deliberately NOT gated on opt-in: create_work_item carries
	// an alias on nine live steps and has never declared ConfigKeys, so gating
	// would hide the largest real user of the mechanism.
	// `removed` is the fourth bucket, and the only one whose live hit means
	// "stop": a retired key still present in a definition will hard-fail
	// validation when a binary carrying the declaration rolls (bugs_open/234).
	// Not gated on opt-in, same as `deprecated` and for the same reason.
	out := map[string]interface{}{
		"declared":    declared,
		"conditional": conditional,
		"deprecated":  datahelpers.ListDeprecatedConfigKeys(),
		"removed":     datahelpers.ListRemovedConfigKeys(),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit: %v\n", err)
		os.Exit(1)
	}
}

// emitSpecs dumps every registered action's full ActionInputSpec.
//
// Same refusal-on-empty as the default path, and for the same reason: a dropped
// blank import would make this print "{}" — read as "no action declares anything",
// which is a coverage gap of exactly the wrong sign — rather than failing.
func emitSpecs() {
	names := datahelpers.ListActionInputSpecNames()
	if len(names) == 0 {
		fmt.Fprintln(os.Stderr,
			"config-key-audit --specs: no action registered an ActionInputSpec.\n"+
				"That is either true or the actions package is no longer linked in, in "+
				"which case this would report an empty gap rather than failing. Check "+
				"the blank import before believing it.")
		os.Exit(2)
	}

	// opted_in is taken from datahelpers' OWN gate rather than recomputed here.
	// ListDeclaredConfigKeys already returns exactly the opted-in actions, so
	// membership of it IS the answer. Re-deriving `CheckConfig || len(ConfigKeys)>0`
	// in this file would create a fourth copy of a rule that until today had three,
	// and those three silently disagreeing is the defect class this whole tool
	// exists to surface — writing a fresh one here would be the tool committing it.
	optedIn := datahelpers.ListDeclaredConfigKeys()

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
		var depConfig map[string]string
		if len(spec.DeprecatedConfigKeys) > 0 {
			depConfig = make(map[string]string, len(spec.DeprecatedConfigKeys))
			for old, canonical := range spec.DeprecatedConfigKeys {
				depConfig[old] = canonical
			}
		}
		var removed map[string]string
		if len(spec.RemovedConfigKeys) > 0 {
			removed = make(map[string]string, len(spec.RemovedConfigKeys))
			for k, msg := range spec.RemovedConfigKeys {
				removed[k] = msg
			}
		}
		var defaults map[string]interface{}
		if len(spec.Defaults) > 0 {
			defaults = make(map[string]interface{}, len(spec.Defaults))
			for k, v := range spec.Defaults {
				defaults[k] = v
			}
		}
		_, isOptedIn := optedIn[name]
		out[name] = specDump{
			Required:             nonNil(spec.Required),
			Optional:             nonNil(spec.Optional),
			ConfigKeys:           nonNil(spec.ConfigKeys),
			Deprecated:           dep,
			DeprecatedConfigKeys: depConfig,
			RemovedConfigKeys:    removed,
			Defaults:             defaults,
			OptedIn:              isOptedIn,
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --specs: %v\n", err)
		os.Exit(1)
	}
}

// emitLivePairs reads an export of live agent workflows on stdin and prints every
// (action, config-key) pair they carry, as TSV: action, key, "nested"|"top".
//
// WHY THIS IS A GO MODE RATHER THAN THE SQL IT REPLACES (bugs_open/144). The audit's
// query walked `default_config->'workflow'->'steps'` — the top level only, exactly
// like the runtime validator did. Two halves blind in the same direction agree with
// each other, and consistent blindness reads exactly like correctness: 25 (action,
// key) pairs existed ONLY inside loop sub-workflows and were invisible to both
// (measured 2026-07-29). Rewriting the SQL to descend would have fixed today's
// numbers and left two hand-written traversals to drift apart again. This walks the
// definition with validation.WalkSteps — the SAME traversal the validator enforces
// against — so "what the audit sees" and "what is validated" cannot diverge without
// a compile error.
//
// A flag rather than a second binary, on the precedent recorded in this file's
// header: the council gate objected to a second binary over the same registry and
// the same question, and was right.
//
// Input shape (see scripts/audit-config-keys.sh):
//
//	[{"type": "<agent>", "workflow": {"start_step": ..., "steps": {...}}}, ...]
func emitLivePairs() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --live-pairs: reading stdin: %v\n", err)
		os.Exit(2)
	}

	agents, failed, err := decodeLiveAgents(raw, "--live-pairs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --live-pairs: %v\n", err)
		os.Exit(2)
	}

	// Walked per agent: one undecodable definition (counted above, in failed) must
	// cost that definition, not the whole report. A report that quietly covers
	// less than it claims is the defect this tool exists to expose.
	pairs := map[string]bool{}
	decoded := len(agents)
	for _, agent := range agents {
		validation.WalkSteps(agent.Workflow, func(_ string, step models.Step, nested bool) {
			if step.Action == "" {
				return
			}
			where := "top"
			if nested {
				where = "nested"
			}
			for key := range step.Config {
				pairs[step.Action+"\t"+key+"\t"+where] = true
			}
		})
	}

	if len(pairs) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --live-pairs: no (action, key) pairs found in %d decoded agents (%d undecodable).\n"+
				"That is either an empty fleet or a broken export — refusing rather than printing "+
				"an empty report that reads as 'nothing to fix'.\n", decoded, failed)
		os.Exit(2)
	}

	out := make([]string, 0, len(pairs))
	for p := range pairs {
		out = append(out, p)
	}
	sort.Strings(out)
	for _, p := range out {
		fmt.Println(p)
	}
	fmt.Fprintf(os.Stderr, "config-key-audit --live-pairs: %d agents decoded (%d undecodable), %d distinct pairs\n",
		decoded, failed, len(pairs))
}

// unregisteredActionFinding names one step whose action neither the local
// registry nor a topic can dispatch — the exact condition
// platform/validation/workflow.go's validateStep and subworkflow.go's
// validateSubWorkflowStep reject a MESSAGE on ("requires a topic"). This makes the
// same check offline (bugs_open/148): the runtime validator only speaks once a
// message arrives at a structurally unrunnable step, so a definition can carry
// the defect for weeks with nothing reporting it.
type unregisteredActionFinding struct {
	Agent  string `json:"agent"`
	Path   string `json:"path"`
	Action string `json:"action"`
	Nested bool   `json:"nested"`
}

// findUnregisteredActions is the pure check, separated from I/O so it can be
// exercised by a fixture test without stdin/stdout plumbing. It walks with
// validation.WalkSteps — the SAME traversal the runtime validator enforces
// against, top-level and nested — for the reason recorded on that function:
// bugs_open/144 was two hand-written traversals blind in the same direction,
// agreeing with each other. A step with a topic is legitimately remote and is
// NOT a finding; only the unregistered-and-untopic'd pair is (bugs_open/148 §6).
func findUnregisteredActions(agents []liveAgent) []unregisteredActionFinding {
	findings := []unregisteredActionFinding{}
	for _, agent := range agents {
		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			if step.Action == "" {
				return
			}
			// Mirrors platform/validation/workflow.go's validateStep and
			// subworkflow.go's validateSubWorkflowStep EXACTLY — including their
			// choice not to exempt "fan_out": if a live fan_out step is neither
			// locally registered nor topic-routed, the real validator rejects it
			// too, and this tool exists to say so before a message finds out.
			if actioncheck.IsLocalAction(step.Action) || step.Topic != "" {
				return
			}
			findings = append(findings, unregisteredActionFinding{
				Agent: agent.Type, Path: path, Action: step.Action, Nested: nested,
			})
		})
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Agent != findings[j].Agent {
			return findings[i].Agent < findings[j].Agent
		}
		return findings[i].Path < findings[j].Path
	})
	return findings
}

// emitUnregisteredActions reads the same stdin shape as --live-pairs and prints
// every step that findUnregisteredActions flags, as JSON. Zero findings is a
// legitimate, meaningful result (unlike an empty decoded-agent set, which is
// refused below) — it says the fleet currently has none, not that the check
// didn't run.
func emitUnregisteredActions() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --unregistered-actions: reading stdin: %v\n", err)
		os.Exit(2)
	}

	agents, failed, err := decodeLiveAgents(raw, "--unregistered-actions")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --unregistered-actions: %v\n", err)
		os.Exit(2)
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --unregistered-actions: 0 agents decoded (%d undecodable) — "+
				"refusing to print a clean report over an empty or broken export.\n", failed)
		os.Exit(2)
	}

	findings := findUnregisteredActions(agents)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(findings); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --unregistered-actions: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "config-key-audit --unregistered-actions: %d agents decoded (%d undecodable), %d finding(s)\n",
		len(agents), failed, len(findings))
	if len(findings) > 0 {
		os.Exit(1)
	}
}

// suspiciousKeyFinding names one live step-config key whose NAME carries
// schema-documentation punctuation. Unlike an unknown key, this is not a
// judgement about what an action reads — a key called "category?" cannot be read
// by anything, whatever the action's contract says.
type suspiciousKeyFinding struct {
	Agent  string `json:"agent"`
	Path   string `json:"path"`
	Action string `json:"action"`
	Key    string `json:"key"`
	Nested bool   `json:"nested"`
}

// suspiciousKeyChars is the punctuation that schema DOCUMENTATION uses and a key
// an action reads never does: "?" for optional ("category?"), "*" for a glob or
// a wildcard, ":" for a "key: value" line lifted out of a contract sketch, and a
// space for prose that was never a key at all.
//
// bugs_open/134 is why this exists. Seed 156 described the agent's input
// contract in a comment — "{site_id, category?}", where the trailing "?" is the
// ordinary convention for "optional" and entirely correct there — and forty-five
// lines later the same notation was written inside the real JSON, as part of two
// key names. `refresh_product_specs` reads "category" and "limit", so neither
// key ever resolved; an unrecognised config key is silently ignored at
// execution, so nothing complained and the step went on describing behaviour it
// did not perform.
//
// The set is deliberately NARROW. Every character here is one that cannot
// plausibly appear in a key some action reads, so a finding is always signal and
// never a style opinion — no underscore, no dot, no hyphen, no case rule. A
// broader set would turn this into a linter people learn to ignore, which is the
// state that lets the next real one through.
const suspiciousKeyChars = "?* :"

// findSuspiciousKeys is the pure check, separated from I/O so a fixture test can
// exercise it (see findUnregisteredActions). It walks with validation.WalkSteps
// — the SAME traversal the runtime validator enforces against, top-level and
// nested — for the bugs_open/144 reason recorded there: two hand-written
// descents go blind in different directions and then agree with each other.
//
// It flags the key REGARDLESS of whether the action has opted into
// ActionInputSpec checking, and that independence is the whole point. The
// unknown-key report can only speak about actions someone has already declared,
// so it stayed silent on bug 134 for as long as `refresh_product_specs` had no
// declaration — while the defect was visible in the key name itself the entire
// time. This check needs to know nothing about the action to be certain.
func findSuspiciousKeys(agents []liveAgent) []suspiciousKeyFinding {
	findings := []suspiciousKeyFinding{}
	for _, agent := range agents {
		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			for key := range step.Config {
				if !strings.ContainsAny(key, suspiciousKeyChars) {
					continue
				}
				findings = append(findings, suspiciousKeyFinding{
					Agent: agent.Type, Path: path, Action: step.Action, Key: key, Nested: nested,
				})
			}
		})
	}
	// Map iteration over step.Config is random, so the sort is what makes the
	// report diffable between runs — a check whose output reorders itself cannot
	// be committed as a baseline.
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

// emitSuspiciousKeys reads the same stdin shape as --live-pairs and
// --unregistered-actions and prints every key findSuspiciousKeys flags, as JSON.
//
// Zero findings is a legitimate, meaningful result — it says no live definition
// currently carries doc notation in a key name. An empty decoded-agent set is
// not: it would produce that same clean report over a truncated or broken
// export, so it is refused, exactly as on emitUnregisteredActions.
func emitSuspiciousKeys() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --suspicious-keys: reading stdin: %v\n", err)
		os.Exit(2)
	}

	agents, failed, err := decodeLiveAgents(raw, "--suspicious-keys")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --suspicious-keys: %v\n", err)
		os.Exit(2)
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --suspicious-keys: 0 agents decoded (%d undecodable) — "+
				"refusing to print a clean report over an empty or broken export.\n", failed)
		os.Exit(2)
	}

	findings := findSuspiciousKeys(agents)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(findings); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --suspicious-keys: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "config-key-audit --suspicious-keys: %d agents decoded (%d undecodable), %d finding(s)\n",
		len(agents), failed, len(findings))
	if len(findings) > 0 {
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
