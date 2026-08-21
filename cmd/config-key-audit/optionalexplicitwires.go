// FILE: cmd/config-key-audit/optionalexplicitwires.go
//
// RFC_029 §10.15, council REVISE round 1 (bug_historian, GATING). Which live
// step wires a field with the `?` OPTIONAL-EXPLICIT marker, and has each one
// been acknowledged as SAFE WHEN ABSENT?
//
// WHY THIS EXISTS, in the seat's own terms. `?` means "resolve this declared
// path or leave the field ABSENT" — deliberately with no error, no log and no
// row, because absence IS the contract. That silence is the point of the marker
// AND the estate's most repeatedly-rediscovered failure class: something is
// dropped, nothing says so, and the empty value is rendered or persisted before
// anyone notices. The marker cannot distinguish "absent because the author
// meant optional" from "absent because the wire is a typo" — only the adopter
// can, and only at adoption time.
//
// So the guard is offline and fleet-wide, exactly like --single-owner-actions
// and for the same reason: a run cannot see whether ITS absence was intended,
// and by the time a page renders empty the decision is long past. This mode
// makes adoption COUNTABLE from config (the RFC says do not count it in logs —
// there are none to count) and makes an UNACKNOWLEDGED adoption loud.
//
// The acks file follows RFC_022's optional_key_budget_acks.json exactly: an
// entry records that a human read the DOWNSTREAM code and confirmed it treats
// an absent field as a safe no-op rather than an unguarded empty string or
// zero. `downstream` is not optional in spirit — an ack with no statement of
// what was checked is the thing this guard exists to prevent.
//
// SURFACE NOTE, and it is the whole reason the finding is precise: a `?` key
// inside a step's `input_mapping` is the OTHER surface (ResolveInputMapping,
// 77 live keys, years old, hard-fails nothing) and is NOT reported here. This
// mode reports only a `?` key sitting directly on a step's `config`, which is
// the ExtractActionInputs surface the marker was added to on 2026-08-20.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/validation"
)

// optionalExplicitWire is one live `?` wire on the ExtractActionInputs surface.
// Reference is carried because the ack has to be judged against the path the
// author actually wrote — "is absence safe here" is unanswerable without it.
type optionalExplicitWire struct {
	Agent     string `json:"agent"`
	Path      string `json:"path"`
	Action    string `json:"action"`
	Field     string `json:"field"`
	Reference string `json:"reference"`
	// Acknowledged is true when the acks file carries this (action, field) with
	// a non-empty downstream statement. An ack keyed on the action+field rather
	// than the agent is deliberate: the question "does the DOWNSTREAM code cope
	// with this field being absent" is a property of the action that consumes
	// it, not of which workflow happens to wire it.
	Acknowledged bool `json:"acknowledged"`
}

// ackedOptionalExplicit is the acks file: {"<action>.<field>": {...}}.
// Only the presence of a non-empty `downstream` makes an entry count — a key
// with an empty statement reads as acknowledged to a grep and is exactly the
// hollow ack the seat objected to.
type optionalExplicitAck struct {
	Downstream string `json:"downstream"`
	Date       string `json:"date"`
	Review     string `json:"review"`
}

func loadOptionalExplicitAcks(path string) (map[string]bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var all map[string]json.RawMessage
	if err := json.Unmarshal(raw, &all); err != nil {
		return nil, fmt.Errorf("acks file is not a JSON object: %w", err)
	}
	acked := make(map[string]bool, len(all))
	for key, blob := range all {
		if strings.HasPrefix(key, "_") {
			continue // documentation keys, as in optional_key_budget_acks.json
		}
		var ack optionalExplicitAck
		if err := json.Unmarshal(blob, &ack); err != nil {
			return nil, fmt.Errorf("acks entry %q: %w", key, err)
		}
		if strings.TrimSpace(ack.Downstream) == "" {
			// Counted as NOT acknowledged, loudly: an entry that says nothing
			// about the downstream is the shape the gating objection named.
			fmt.Fprintf(os.Stderr,
				"config-key-audit --optional-explicit-wires: acks entry %q has an empty "+
					"`downstream` — ignoring it; an ack must say what was checked.\n", key)
			continue
		}
		acked[key] = true
	}
	return acked, nil
}

// findOptionalExplicitWires is the pure check. Same validation.WalkSteps
// traversal as every other mode here — top-level AND nested, both container
// spellings — for the bugs_open/144 reason two hand-written descents go blind
// in different directions. That matters more than usual here: this mode is
// also the code-shaped re-check of the submission's "zero live `?` keys"
// census, and a SQL walk that knew only `sub_workflow` would have missed every
// `substeps` population (which is what execution actually prefers).
func findOptionalExplicitWires(agents []liveAgent, acked map[string]bool) []optionalExplicitWire {
	findings := []optionalExplicitWire{}

	for _, agent := range agents {
		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			for key, val := range step.Config {
				base, strict, optional := datahelpers.MarkedConfigKey(key)
				if strict || !optional || base == "" {
					continue
				}
				ref, _ := val.(string)
				findings = append(findings, optionalExplicitWire{
					Agent:        agent.Type,
					Path:         path,
					Action:       step.Action,
					Field:        base,
					Reference:    ref,
					Acknowledged: acked[step.Action+"."+base],
				})
			}
		})
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Agent != findings[j].Agent {
			return findings[i].Agent < findings[j].Agent
		}
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		return findings[i].Field < findings[j].Field
	})
	return findings
}

// walkedStepCount is the VACUITY GUARD's input: how many workflow steps the
// walker actually visited across the whole fleet.
//
// WHY IT EXISTS (reported 2026-08-21 by the bugs_open/286 session while
// confirming an ack, and it is a defect in this tool of exactly the class this
// tool was built to prevent): the fleet export must carry each agent's workflow
// at the TOP level as `workflow`. Feed it the shape one reaches for first —
// `json_build_object('type', type, 'default_config', default_config)` — and
// every agent decodes cleanly, `liveAgent.Workflow` is the zero value for all of
// them, the walker visits NOTHING, and the report says "0 live `?` wires, 0
// unacknowledged, exit 0". That is indistinguishable from a genuinely clean
// fleet: same wire count, same exit code, same absence of any error. The only
// tell was the number this guard now makes load-bearing.
//
// So: agents decoded but zero steps walked is "the check did not run", which
// must NEVER be reported as a pass. Deliberately not phrased as "workflow is
// empty for every agent" — this counts what the walker actually reached, so it
// also catches a future export that carries `workflow` but loses `steps`, or a
// walker change that stops descending.
// vacuityRefusal returns the refusal message when this run cannot mean
// anything, and "" when it can. Split out of the caller so the DECISION is
// testable: os.Exit is not, and a guard whose only test asserts its INPUTS
// (walkedStepCount) leaves the branch itself unpinned — which is what a
// mutation proof caught here, on a comment claiming otherwise.
func vacuityRefusal(agents []liveAgent) string {
	if len(agents) == 0 || walkedStepCount(agents) > 0 {
		return ""
	}
	return fmt.Sprintf(
		"config-key-audit --optional-explicit-wires: %d agents decoded but ZERO workflow steps "+
			"walked — refusing to report. This is almost always the export SHAPE: each element "+
			"must carry the workflow at the top level as `workflow`, i.e. "+
			"jsonb_build_object('type', type, 'workflow', default_config->'workflow'). "+
			"A `default_config`-shaped feed decodes cleanly and walks nothing, which would "+
			"otherwise print as a clean fleet.\n", len(agents))
}

func walkedStepCount(agents []liveAgent) int {
	n := 0
	for _, agent := range agents {
		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			n++
		})
	}
	return n
}

// emitOptionalExplicitWires reads the same stdin shape as the other modes.
//
// EXIT CODES CARRY THE FINDING: 0 = every live `?` wire is acknowledged (zero
// wires is a legitimate 0 — the marker shipped with no adopters, and unlike
// --single-owner-actions there is no declaration set whose emptiness would make
// a clean report vacuous, because the population being audited is the CONFIG,
// not the code). 1 = at least one wire is live and unacknowledged. 2 = the
// export or the acks file could not be read, i.e. the check did not run — which
// must never be reported as a pass.
func emitOptionalExplicitWires(args []string) {
	acksPath := "docs/agent_docs/docs024_key_docs_latest/architecture_review/optional_explicit_wire_acks.json"
	report := false
	for i := 0; i < len(args); i++ {
		if args[i] == "--acks" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr,
					"config-key-audit --optional-explicit-wires: --acks needs a file path")
				os.Exit(2)
			}
			acksPath = args[i+1]
			i++
			continue
		}
		if args[i] == "--report" {
			report = true
			continue
		}
		fmt.Fprintf(os.Stderr,
			"config-key-audit --optional-explicit-wires: unrecognised argument %q "+
				"(want: [--acks <file>] [--report])\n", args[i])
		os.Exit(2)
	}

	acked, err := loadOptionalExplicitAcks(acksPath)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --optional-explicit-wires: acks file %q: %v — refusing to run "+
				"without it, because every wire would then read as unacknowledged and the "+
				"report would be noise rather than a finding.\n", acksPath, err)
		os.Exit(2)
	}

	var (
		agents []liveAgent
		failed int
	)
	if report {
		// Straight from Postgres: this image contains no kubectl and the service
		// account has no pods/exec RBAC (see fleetdb.go).
		var db *sql.DB
		db, err = dbConn()
		if err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --optional-explicit-wires: %v\n", err)
			os.Exit(2)
		}
		// dbConn returns (nil, nil) when PG_CLIENTS_HOST is unset — "no DB
		// configured" is deliberately not an error there. Every caller must
		// therefore check for nil, and the sibling --report modes currently do
		// not: they go straight to `defer db.Close()` and panic. A panic still
		// exits non-zero so it cannot read as a pass, which is why this is a
		// papercut rather than a defect — but it only ever fires for a person
		// running the check by hand, i.e. exactly when a legible message is worth
		// most. Not fixing the siblings from here: that is three other modes'
		// behaviour and belongs in its own change.
		if db == nil {
			fmt.Fprintln(os.Stderr,
				"config-key-audit --optional-explicit-wires --report: PG_CLIENTS_HOST is not set, "+
					"so there is no fleet to read. In the CronJob this comes from the pod env; by "+
					"hand, either export it or drop --report and pipe the fleet export on stdin.")
			os.Exit(2)
		}
		defer db.Close()
		agents, failed, err = loadLiveAgentsFromDB(db, "--optional-explicit-wires")
	} else {
		var raw []byte
		raw, err = io.ReadAll(os.Stdin)
		if err == nil {
			agents, failed, err = decodeLiveAgents(raw, "--optional-explicit-wires")
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --optional-explicit-wires: %v\n", err)
		os.Exit(2)
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --optional-explicit-wires: 0 agents decoded (%d undecodable) — "+
				"refusing to print a clean report over an empty or broken export.\n", failed)
		os.Exit(2)
	}

	// VACUITY GUARD — before any finding is reported, prove the walker saw
	// something. A fleet of agents with no steps yields the same clean report as
	// a fleet with no `?` wires.
	if msg := vacuityRefusal(agents); msg != "" {
		fmt.Fprint(os.Stderr, msg)
		os.Exit(2)
	}

	findings := findOptionalExplicitWires(agents, acked)

	if report {
		summary := optionalExplicitRunSummary(len(agents), failed, acksPath, findings)
		fmt.Print(summary)
		// ONE row per run, clean or not — the convention every scheduled check
		// here follows, and the reason is that a check which only speaks when it
		// fails is indistinguishable from one that has stopped running. A MISSING
		// row must mean "the job did not run", never "nothing is wrong".
		writeDocNote("optional-explicit-wires", summary,
			"optional-explicit-wires", "optional-explicit-wires-check")
		if unacknowledgedCount(findings) > 0 {
			os.Exit(1)
		}
		return
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(findings); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --optional-explicit-wires: %v\n", err)
		os.Exit(1)
	}

	unacked := unacknowledgedCount(findings)
	fmt.Fprintf(os.Stderr,
		"config-key-audit --optional-explicit-wires: %d agents decoded (%d undecodable), "+
			"%d live `?` wire(s) on the ExtractActionInputs surface, %d unacknowledged\n",
		len(agents), failed, len(findings), unacked)
	if unacked > 0 {
		fmt.Fprintf(os.Stderr,
			"  Each needs an entry in %s stating what was checked DOWNSTREAM: that the "+
				"consuming code treats an absent field as a safe no-op, not as an empty "+
				"string or zero it will render or persist.\n", acksPath)
		os.Exit(1)
	}
}

// unacknowledgedCount is the exit-code decision in one place: the report half
// and the stdin half must agree on what counts as a finding, or a green cron
// and a red hand-run would disagree about the same fleet.
func unacknowledgedCount(findings []optionalExplicitWire) int {
	n := 0
	for _, f := range findings {
		if !f.Acknowledged {
			n++
		}
	}
	return n
}

// optionalExplicitRunSummary is the doc_notes body. It states the ACKNOWLEDGED
// count as well as the unacknowledged one, because "0 wires" and "3 wires, all
// acknowledged" are both clean and a reader needs to tell them apart — the
// first can also mean the marker is unused, and reading that as "the gate is
// working" is the failure this whole check exists to prevent.
func optionalExplicitRunSummary(agents, failed int, acksPath string, findings []optionalExplicitWire) string {
	unacked := unacknowledgedCount(findings)
	var b strings.Builder
	fmt.Fprintf(&b, "optional-explicit-wires: %d live `?` wire(s) on the ExtractActionInputs surface, "+
		"%d acknowledged, %d UNACKNOWLEDGED (%d agents, %d undecodable; acks=%s)\n",
		len(findings), len(findings)-unacked, unacked, agents, failed, acksPath)
	if len(findings) == 0 {
		b.WriteString("  No adopters. The marker is opt-in, so this is the expected state until a " +
			"migration wires one; it is NOT evidence that anything was checked and passed.\n")
	}
	for _, f := range findings {
		state := "ack"
		if !f.Acknowledged {
			state = "UNACKNOWLEDGED"
		}
		fmt.Fprintf(&b, "  [%s] %s %s -> %s.%s = %s\n",
			state, f.Agent, f.Path, f.Action, f.Field, f.Reference)
	}
	if unacked > 0 {
		fmt.Fprintf(&b, "  Each unacknowledged wire needs an entry in %s stating what was checked "+
			"DOWNSTREAM: that the consuming code treats an absent field as a safe no-op, not as an "+
			"empty string or zero it will render or persist.\n", acksPath)
	}
	return b.String()
}
