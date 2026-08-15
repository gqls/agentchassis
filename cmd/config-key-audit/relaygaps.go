// FILE: cmd/config-key-audit/relaygaps.go
//
// `config-key-audit --relay-gaps` — does a DISPATCHER still forward everything
// its handler declares it accepts?
//
// bugs_open/174. `diagnose-dispatch-loop` claims a `needs_diagnosis` work item
// and calls the handler named on it. The envelope it relays is defined by two
// hand-maintained lists on one path, and they drifted from the callee's declared
// contract in lockstep — agreeing with each other, and with nothing else:
//
//	1. `claim_item`'s SQL RETURNING clause, which projects spec keys into
//	   `claimed.*`. A key it does not project cannot be forwarded, whatever the
//	   input_mapping says. The 174 ticket's own fix candidate named only list 2,
//	   and would have failed silently against this one.
//	2. `call_handler`'s `input_mapping`, which is an ALLOW-LIST, not a
//	   passthrough — an unlisted key is dropped without a word
//	   (`input_contracts/input_mapping.go`: an optional field that does not
//	   resolve is logged at Info and skipped).
//
// Nothing checked either against `diagnose-orchestrator`'s `input_contract`,
// which named `seed_scope` and `runtime_page` the whole time. Three lanes' real
// diagnoses ran against a scope nobody chose.
//
// WHY THIS IS A REGISTRY AND NOT A FLEET-WIDE RULE. The obvious general check —
// "every call_agent must forward every key its callee declares" — was written
// and MEASURED first, and it is not sound. Over the live fleet on 2026-08-02 it
// produced 31 findings from 75 resolvable call sites, and the ones spot-checked
// were legitimate: `pageflow-builder.apply_site_design` omits `site_context`,
// and `webdesign-agent` has an explicit `else_step: load_site_context` to load it
// itself. Tightening to "the callee also USES the key" cut that to 3 and STILL
// could not tell "the caller dropped it" from "the caller never had it" — which
// is the whole question. Worse, it was blind to 174 itself: `call_handler`
// resolves its callee through `agent_type_field: claimed.handler_agent`, a
// runtime value, so a static resolver skips the one site the check exists for.
//
// A dispatcher is the case where the question IS answerable, because the
// caller's envelope is not "whatever happens to be in collected_data" — it is
// the work item spec, and the spec's shape is exactly what the handler's
// contract declares. So the assertion is made where it is sound and declared
// where it is made.
//
// AND THE HALF THAT KEEPS IT FROM GOING INERT: the tool also DISCOVERS
// dispatcher-shaped relays in the live config and reports any that the registry
// does not cover. A new dispatch loop cannot quietly appear unchecked, and
// deleting a registry entry to silence a finding shows up as an uncovered relay
// rather than as a clean report.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/validation"
)

// contractAgent is this mode's own view of the live export. It deliberately does
// NOT extend main.go's liveAgent: --relay-gaps is the only mode that needs
// input_contract, and on 2026-08-02 main.go was being edited concurrently by
// another session, so widening a shared struct would have made this file's commit
// carry theirs — or broken the build at HEAD, which every `make build-*` compiles
// from. Same JSON, decoded independently; the cost is one extra unmarshal of an
// export that is ~1MB and read once.
type contractAgent struct {
	Type     string              `json:"type"`
	Workflow models.WorkflowPlan `json:"workflow"`
	// Absent on roughly half the fleet (95 of 181 live definitions carried one on
	// 2026-08-02), so a consumer must treat "no contract" as "nothing to check",
	// never as "accepts nothing" — findRelayGaps reports that case as UNMATCHED
	// rather than as a clean pass.
	InputContract struct {
		Required []string `json:"required"`
		Optional []string `json:"optional"`
	} `json:"input_contract"`
}

// decodeContractAgents mirrors decodeLiveAgents' contract exactly: one
// undecodable definition costs that definition, not the whole report, and the
// count is returned so the caller can print it rather than swallow it. A
// truncated kubectl exec exits 0, so a short read arrives here looking like a
// small, healthy fleet.
func decodeContractAgents(raw []byte) (agents []contractAgent, failed int, err error) {
	var rows []json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, 0, fmt.Errorf("stdin is not a JSON array of agents: %w\n"+
			"A truncated kubectl exec exits 0, so a short read arrives here looking like a small fleet.", err)
	}
	for _, row := range rows {
		var agent contractAgent
		if jsonErr := json.Unmarshal(row, &agent); jsonErr != nil {
			failed++
			fmt.Fprintf(os.Stderr, "config-key-audit --relay-gaps: undecodable agent row: %v\n", jsonErr)
			continue
		}
		agents = append(agents, agent)
	}
	return agents, failed, nil
}

// relaySpec declares one dispatcher->handler hop we assert the envelope on.
//
// Callee is named EXPLICITLY rather than resolved from config because the live
// dispatchers resolve theirs at runtime from the claimed row (`agent_type_field:
// claimed.handler_agent`). Naming it is what makes the check possible at all;
// LoopInternal is the named exception list the 174 ticket asked for, so a key
// the loop legitimately invents (its own bookkeeping) is not read as a gap in
// the other direction.
type relaySpec struct {
	Caller       string   // the dispatching agent
	CallStep     string   // its call_agent step
	ClaimStep    string   // the query_database step whose RETURNING builds the envelope
	Envelope     string   // the output_field that step writes (the `claimed.` root)
	Callee       string   // the handler whose input_contract is the authority
	LoopInternal []string // keys the loop supplies itself; not expected in the spec
}

// declaredRelays is the registry. Adding a dispatch loop means adding a line
// here — and if you forget, --relay-gaps reports your loop as an UNCOVERED
// relay rather than passing quietly.
//
// Only diagnose-dispatch-loop is asserted today. The other two dispatcher-shaped
// relays discovered live on 2026-08-02 (`report-dispatch-loop.call_handler`,
// `build-pipeline-trigger.call_dispatch`) are deliberately NOT registered: their
// callees are resolved at runtime and their handlers' contracts have not been
// read, so registering them would be asserting something nobody has checked.
// They show up in the uncovered list, which is the honest state.
var declaredRelays = []relaySpec{
	{
		Caller:       "diagnose-dispatch-loop",
		CallStep:     "call_handler",
		ClaimStep:    "claim_item",
		Envelope:     "claimed",
		Callee:       "diagnose-orchestrator",
		LoopInternal: []string{"work_item_id"},
	},
}

type relayGapFinding struct {
	Caller string `json:"caller"`
	Step   string `json:"step"`
	Callee string `json:"callee"`
	// NotProjected: the callee declares it, but the claim query's RETURNING never
	// produces it — so no input_mapping entry could carry it. This is the list the
	// 174 ticket missed, and the one that makes a mapping-only fix silently inert.
	NotProjected []string `json:"not_projected"`
	// NotForwarded: projected (or projectable) but absent from the input_mapping.
	NotForwarded []string `json:"not_forwarded"`
	// MapsToNothing: an input_mapping entry sourced from the envelope at a key the
	// claim query does not project. It reads as wired and resolves to nothing.
	MapsToNothing []string `json:"maps_to_nothing"`
}

type uncoveredRelay struct {
	Caller string `json:"caller"`
	Step   string `json:"step"`
	// Why it looked like a dispatcher, so the reader can judge the heuristic
	// rather than take it on trust.
	Envelope string `json:"envelope"`
	Reason   string `json:"reason"`
}

type relayGapReport struct {
	Findings  []relayGapFinding `json:"findings"`
	Uncovered []uncoveredRelay  `json:"uncovered_relays"`
	// Registered relays whose caller/step/callee could not be found in the export
	// at all. A registry entry that matches nothing is a check that silently stopped
	// running — the failure mode this whole file exists to catch, one level up.
	Unmatched []string `json:"unmatched_registry_entries"`
}

// projectedAliases pulls the column aliases out of a RETURNING clause. It is a
// regex over SQL, which is a blunt instrument — so it is used ONLY to answer
// "does this alias appear", never to rewrite anything, and a miss produces a
// finding (a false positive a human reads) rather than a silent pass.
var returningAliasRe = regexp.MustCompile(`(?i)\bAS\s+([a-z_][a-z0-9_]*)`)

func projectedAliases(query string) map[string]bool {
	out := map[string]bool{}
	idx := strings.Index(strings.ToUpper(query), "RETURNING")
	if idx < 0 {
		return out
	}
	for _, m := range returningAliasRe.FindAllStringSubmatch(query[idx:], -1) {
		out[strings.ToLower(m[1])] = true
	}
	// A bare column in RETURNING (no AS) is its own alias. `handler_agent` in
	// diagnose-dispatch-loop is exactly this, and treating it as unprojected
	// would invent a gap.
	tail := query[idx+len("RETURNING"):]
	for _, part := range strings.Split(tail, ",") {
		f := strings.Fields(strings.TrimSpace(part))
		if len(f) == 1 && regexp.MustCompile(`^[a-z_][a-z0-9_]*$`).MatchString(f[0]) {
			out[f[0]] = true
		}
	}
	return out
}

// stepsOf indexes an agent's steps by path using the same traversal every other
// mode here uses (validation.WalkSteps, top-level and nested) — bugs_open/144's
// rule that a second hand-written descent goes blind in its own direction.
func stepsOf(agent contractAgent) map[string]models.Step {
	out := map[string]models.Step{}
	validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
		out[path] = step
	})
	return out
}

// stepName is the last segment of a WalkSteps path. WalkSteps qualifies every
// step ("steps.call_handler", and deeper for nested sub-workflows) while the
// registry names the step as an author would write it, so the two must be
// compared through one normalisation rather than by equality.
//
// This mismatch is not hypothetical: the first live run of --relay-gaps matched
// NOTHING and reported one unmatched registry entry. That is the only reason it
// was caught in a minute rather than becoming a detector that passes for ever —
// which is the same failure, one level up, as the two allow-lists this file
// exists to check.
func stepName(path string) string {
	if i := strings.LastIndex(path, "."); i >= 0 {
		return path[i+1:]
	}
	return path
}

// lookupStep finds a step by its authored name, whatever depth WalkSteps found
// it at. Returns the qualified path too, so callers can record where it matched.
func lookupStep(steps map[string]models.Step, name string) (models.Step, string, bool) {
	for path, st := range steps {
		if stepName(path) == name {
			return st, path, true
		}
	}
	return models.Step{}, "", false
}

func mappingOf(step models.Step) map[string]string {
	raw, ok := step.Config["input_mapping"].(map[string]interface{})
	if !ok {
		return nil
	}
	out := map[string]string{}
	for k, v := range raw {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

func queryOf(step models.Step) string {
	q, _ := step.Config["query"].(string)
	return q
}

// findRelayGaps is the pure check. Split from I/O so a fixture test can drive it,
// matching findUnregisteredActions/findSingleOwnerViolations.
func findRelayGaps(agents []contractAgent) relayGapReport {
	byType := map[string]contractAgent{}
	for _, a := range agents {
		byType[a.Type] = a
	}

	report := relayGapReport{
		Findings:  []relayGapFinding{},
		Uncovered: []uncoveredRelay{},
		Unmatched: []string{},
	}
	covered := map[string]bool{} // caller+"."+step

	for _, spec := range declaredRelays {
		id := spec.Caller + "." + spec.CallStep
		covered[id] = true

		caller, ok := byType[spec.Caller]
		if !ok {
			report.Unmatched = append(report.Unmatched, id+" (caller not in export)")
			continue
		}
		callee, ok := byType[spec.Callee]
		if !ok {
			report.Unmatched = append(report.Unmatched, id+" (callee "+spec.Callee+" not in export)")
			continue
		}
		steps := stepsOf(caller)
		callStep, callPath, ok := lookupStep(steps, spec.CallStep)
		if !ok {
			report.Unmatched = append(report.Unmatched, id+" (call step not found)")
			continue
		}
		covered[spec.Caller+"."+callPath] = true // the qualified path the discovery half sees
		claimStep, _, ok := lookupStep(steps, spec.ClaimStep)
		if !ok {
			report.Unmatched = append(report.Unmatched, id+" (claim step "+spec.ClaimStep+" not found)")
			continue
		}
		mapping := mappingOf(callStep)
		if mapping == nil {
			report.Unmatched = append(report.Unmatched, id+" (call step has no input_mapping)")
			continue
		}

		declared := append(nonNil(callee.InputContract.Required), callee.InputContract.Optional...)
		if len(declared) == 0 {
			report.Unmatched = append(report.Unmatched, id+" (callee "+spec.Callee+" declares no input_contract — nothing to check against)")
			continue
		}

		internal := map[string]bool{}
		for _, k := range spec.LoopInternal {
			internal[k] = true
		}
		projected := projectedAliases(queryOf(claimStep))

		// forwarded: the callee-facing key names the mapping supplies, with the
		// optional "?" and strict "!" (RFC_029 §9 D3) suffixes stripped — those
		// suffixes are about resolution, not about which field the child receives.
		forwarded := map[string]string{}
		for k, src := range mapping {
			forwarded[strings.TrimSuffix(strings.TrimSuffix(k, "!"), "?")] = src
		}

		f := relayGapFinding{Caller: spec.Caller, Step: spec.CallStep, Callee: spec.Callee}
		for _, key := range declared {
			if internal[key] {
				continue
			}
			if _, ok := forwarded[key]; !ok {
				f.NotForwarded = append(f.NotForwarded, key)
				if !projected[key] {
					f.NotProjected = append(f.NotProjected, key)
				}
				continue
			}
		}
		// The other direction: a mapping entry that reads the envelope at a key the
		// claim never projects. It looks wired and resolves to nothing — which is
		// what adding seed_scope to the input_mapping ALONE would have produced.
		for key, src := range forwarded {
			root, col, found := strings.Cut(src, ".")
			if !found || root != spec.Envelope {
				continue
			}
			if !projected[strings.ToLower(col)] {
				f.MapsToNothing = append(f.MapsToNothing, key+" -> "+src)
			}
		}

		sort.Strings(f.NotForwarded)
		sort.Strings(f.NotProjected)
		sort.Strings(f.MapsToNothing)
		// nonNil for the same reason as main.go's: a missing list must encode as []
		// rather than null, so a consumer can iterate without a nil check and cannot
		// read "this category is empty" as "this key is absent from the report".
		f.NotForwarded = nonNil(f.NotForwarded)
		f.NotProjected = nonNil(f.NotProjected)
		f.MapsToNothing = nonNil(f.MapsToNothing)
		if len(f.NotForwarded) > 0 || len(f.MapsToNothing) > 0 {
			report.Findings = append(report.Findings, f)
		}
	}

	report.Uncovered = discoverUncoveredRelays(agents, covered)
	sort.Slice(report.Findings, func(i, j int) bool { return report.Findings[i].Caller < report.Findings[j].Caller })
	sort.Strings(report.Unmatched)
	return report
}

// discoverUncoveredRelays finds dispatcher-SHAPED call sites the registry does
// not cover: a call_agent whose input_mapping is sourced, for the most part,
// from the output_field of a query_database step in the same workflow.
//
// This is a heuristic and is reported as "look at this", never as a defect —
// which is why it is a separate list from Findings. Its job is to stop the
// registry silently falling behind the fleet, the way the two allow-lists in 174
// silently fell behind the contract.
func discoverUncoveredRelays(agents []contractAgent, covered map[string]bool) []uncoveredRelay {
	out := []uncoveredRelay{}
	for _, agent := range agents {
		steps := stepsOf(agent)
		envelopes := map[string]bool{} // output_field of every query_database step
		for _, st := range steps {
			if st.Action == "query_database" && st.OutputField != "" {
				envelopes[st.OutputField] = true
			}
		}
		if len(envelopes) == 0 {
			continue
		}
		for path, st := range steps {
			if st.Action != "call_agent" {
				continue
			}
			mapping := mappingOf(st)
			if len(mapping) == 0 {
				continue
			}
			if covered[agent.Type+"."+path] {
				continue
			}
			roots := map[string]int{}
			for _, src := range mapping {
				root, _, _ := strings.Cut(src, ".")
				roots[root]++
			}
			for root, n := range roots {
				if !envelopes[root] {
					continue
				}
				if n*2 < len(mapping) {
					continue // a stray reference, not an envelope
				}
				out = append(out, uncoveredRelay{
					Caller:   agent.Type,
					Step:     path,
					Envelope: root,
					Reason: fmt.Sprintf("%d of %d input_mapping sources read '%s', the output of a query_database step — dispatcher-shaped, and not in declaredRelays",
						n, len(mapping), root),
				})
				break
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Caller != out[j].Caller {
			return out[i].Caller < out[j].Caller
		}
		return out[i].Step < out[j].Step
	})
	return out
}

// emitRelayGaps reads the same stdin export as the other modes, with one extra
// field: input_contract. Refuses to print over an empty agent set for the reason
// on emitUnregisteredActions — a truncated export exits 0 and arrives here
// looking like a small, healthy fleet.
func emitRelayGaps() {
	if len(declaredRelays) == 0 {
		fmt.Fprintln(os.Stderr,
			"config-key-audit --relay-gaps: declaredRelays is empty — refusing to print "+
				"a clean report that no export could ever fail.")
		os.Exit(2)
	}

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --relay-gaps: reading stdin: %v\n", err)
		os.Exit(2)
	}
	agents, failed, err := decodeContractAgents(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --relay-gaps: %v\n", err)
		os.Exit(2)
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --relay-gaps: 0 agents decoded (%d undecodable) — "+
				"refusing to print a clean report over an empty or broken export.\n", failed)
		os.Exit(2)
	}

	report := findRelayGaps(agents)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --relay-gaps: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr,
		"config-key-audit --relay-gaps: %d agents decoded (%d undecodable), %d relay(s) asserted, %d finding(s), %d uncovered dispatcher-shaped relay(s), %d unmatched registry entr(ies)\n",
		len(agents), failed, len(declaredRelays), len(report.Findings), len(report.Uncovered), len(report.Unmatched))
	// Unmatched is an ERROR, not a finding: it means an assertion stopped running.
	if len(report.Findings) > 0 || len(report.Unmatched) > 0 {
		os.Exit(1)
	}
}
