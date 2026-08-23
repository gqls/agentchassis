// FILE: cmd/config-key-audit/recurrencedeclared.go
//
// --undeclared-recurrence (bugs_open/326): which keyed create_work_item step has
// never said whether the item it files is an ACTION REQUEST or a DETECTED DEFECT?
//
// THE FINDING IS A MISSING DECLARATION, NOT A WRONG GUESS. This mode does not
// know which answer is right for any given step and does not try — it reports
// the steps where nobody has answered at all. Either explicit value is clean.
//
// WHY THE DISTINCTION MATTERS. writeWorkItem runs an anti-churn brake over
// (site_id, item_key): siblings that reached complete/failed inside 7 days slow
// the key down. That is correct for a DETECTED DEFECT, where a repeat means the
// fix is not working. It is wrong for an ACTION REQUEST — a pipeline stage
// handoff, a re-render request, a re-submission — where a `complete` predecessor
// means the previous request SUCCEEDED and asking again is the normal course of
// business. workItem.recurrenceExpected is the opt-out, and it waives only the
// heuristics: idx_swi_dedup still refuses a second OPEN item, so declaring it
// never weakens the concurrency guarantee.
//
// WHY AN AUDIT AND NOT A BETTER DEFAULT. This is the SECOND time the estate has
// paid for the same missing declaration. bugs_closed/024 established the rule in
// 2026-07 for a tool re-render request the brake killed; the remedy reached the
// Go call sites that lane touched and stopped there. Two years of config-driven
// steps later, a customer's domain re-submission died the same way
// (bugs_open/326) because the build chain reaches the same helper through
// create_work_item and nobody had classified it. 024 drew the right conclusion
// and it still did not propagate — because nothing counted adoption. At the
// commit that ships this mode: 19 of 21 keyed steps had never declared.
//
// Flipping the DEFAULT was considered and rejected: it changes what the shared
// mechanism guarantees for enumerated live consumers that genuinely need the
// counter (claims-auditor's revalidator-close loop writes `complete` into the
// two-strike window BY DESIGN), which is the owner's stated architecture trigger
// rather than something to do from inside a bug patch.
//
// THE RULE, every clause mirroring an EXECUTED read so the check convicts what
// the runtime would actually do:
//
//   - Action == "create_work_item". Only this action reads the key.
//   - item_key_prefix under the action's exact read (a bare `.(string)`): with
//     no prefix the item_key is NULL, the row sits outside idx_swi_dedup, AND
//     writeWorkItem's brake is gated on `item.itemKey != ""` — so the brake
//     cannot fire and there is nothing to declare.
//   - `recurrence_expected` ABSENT from config -> finding. Present with any
//     value -> clean, including `false`, and including a non-bool the action
//     would ignore: the latter is reported as declared_unhonoured=true, because
//     an author who wrote `"true"` as a string believes they have opted in and
//     the runtime read (`config[...].(bool)`) silently disagrees. That is the
//     same shape --loop-sitewide-item-keys reports for its suffix field.
//
// THE DESCENT IS validation.WalkSteps — the one walk every mode in this binary
// uses, for bugs_open/144's reason (a second hand-written descent goes blind on
// substeps, and real pairs live only inside loop sub-workflows). Nested steps
// are IN scope here, unlike --loop-sitewide-item-keys, whose defect is
// specifically about loop nesting: the brake fires wherever a keyed item is
// written, top-level or not.
//
// Wired to scripts/audit-undeclared-recurrence.sh.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/validation"
)

// undeclaredRecurrenceFinding is one keyed create_work_item step that has never
// said which kind of item it files. ItemType is the routing hint a reviewer
// needs: it is usually enough to tell a stage handoff from a detector's finding.
type undeclaredRecurrenceFinding struct {
	Agent             string `json:"agent"`
	Path              string `json:"path"`
	ItemType          string `json:"item_type"`
	ItemKeyPrefix     string `json:"item_key_prefix"`
	DeclaredUnhonoured bool  `json:"declared_unhonoured"`
}

// findUndeclaredRecurrence is the pure check (I/O split off, per this binary's
// convention).
func findUndeclaredRecurrence(agents []liveAgent) []undeclaredRecurrenceFinding {
	findings := []undeclaredRecurrenceFinding{}

	for _, agent := range agents {
		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			if step.Action != "create_work_item" {
				return
			}
			// The action's exact prefix read. No prefix -> no item_key -> the
			// brake's `item.itemKey != ""` gate never opens.
			prefix, _ := step.Config["item_key_prefix"].(string)
			if prefix == "" {
				return
			}

			raw, present := step.Config["recurrence_expected"]
			if present {
				// Declared. Clean — unless it is declared in a shape the action
				// cannot read, in which case the author's belief and the
				// runtime's behaviour have quietly diverged.
				if _, isBool := raw.(bool); isBool {
					return
				}
				itemType, _ := step.Config["item_type"].(string)
				findings = append(findings, undeclaredRecurrenceFinding{
					Agent: agent.Type, Path: path, ItemType: itemType,
					ItemKeyPrefix: prefix, DeclaredUnhonoured: true,
				})
				return
			}

			itemType, _ := step.Config["item_type"].(string)
			findings = append(findings, undeclaredRecurrenceFinding{
				Agent: agent.Type, Path: path, ItemType: itemType,
				ItemKeyPrefix: prefix,
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

// emitUndeclaredRecurrence is the I/O half: DB route when PG_CLIENTS_HOST is
// set, stdin otherwise. Same refusals as every mode here — a DB error when the
// env asked for the DB is fatal, and a 0-agent fleet is never reported clean.
func emitUndeclaredRecurrence() {
	report := false
	for _, a := range os.Args[2:] {
		switch a {
		case "--report":
			report = true
		default:
			fmt.Fprintf(os.Stderr, "config-key-audit --undeclared-recurrence: unknown argument %q\n", a)
			os.Exit(2)
		}
	}

	var agents []liveAgent
	var failed int
	var err error
	if db, derr := dbConn(); derr != nil || db != nil {
		if derr != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --undeclared-recurrence: %v\n", derr)
			os.Exit(2)
		}
		defer db.Close()
		agents, failed, err = loadLiveAgentsFromDB(db, "--undeclared-recurrence")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	} else {
		raw, rerr := io.ReadAll(os.Stdin)
		if rerr != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --undeclared-recurrence: reading stdin: %v\n", rerr)
			os.Exit(2)
		}
		agents, failed, err = decodeLiveAgents(raw, "--undeclared-recurrence")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --undeclared-recurrence: 0 live agents decoded — refusing to print a clean report over an empty fleet.\n")
		os.Exit(2)
	}

	findings := findUndeclaredRecurrence(agents)

	out := map[string]interface{}{
		"agents_scanned":   len(agents),
		"agents_undecoded": failed,
		"findings":         findings,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)

	if report {
		writeDocNote("undeclared-recurrence",
			undeclaredRecurrenceRunSummary(len(agents), failed, findings),
			"config-integrity", "undeclared_recurrence_check")
	}

	if len(findings) > 0 {
		fmt.Fprintf(os.Stderr,
			"%d keyed create_work_item step(s) have not declared recurrence_expected: the anti-churn brake will act on them and nobody has said whether it should (bugs_open/326).\n",
			len(findings))
		os.Exit(1)
	}
}

// undeclaredRecurrenceRunSummary is the doc_notes body — prose stating the SCOPE
// as well as the result, because "0 findings over 3 agents" and "0 findings over
// 177" have opposite meanings (the convention every mode here follows).
func undeclaredRecurrenceRunSummary(scanned, undecoded int, findings []undeclaredRecurrenceFinding) string {
	var b strings.Builder
	if len(findings) == 0 {
		fmt.Fprintf(&b, "undeclared-recurrence check CLEAN: every keyed create_work_item step across %d live agents declares whether its item is an action request or a detected defect.", scanned)
	} else {
		fmt.Fprintf(&b, "undeclared-recurrence: %d keyed create_work_item step(s) across %d live agents have not declared recurrence_expected — the anti-churn brake acts on them unreviewed (bugs_open/326): ", len(findings), scanned)
		for i, f := range findings {
			if i > 0 {
				b.WriteString("; ")
			}
			fmt.Fprintf(&b, "%s %s (item_type %s, prefix %s", f.Agent, f.Path, f.ItemType, f.ItemKeyPrefix)
			if f.DeclaredUnhonoured {
				b.WriteString(", DECLARED IN A SHAPE THE ACTION CANNOT READ")
			}
			b.WriteString(")")
		}
		b.WriteString(".")
	}
	if undecoded > 0 {
		fmt.Fprintf(&b, " %d agent row(s) failed to decode and were not scanned.", undecoded)
	}
	return b.String()
}
