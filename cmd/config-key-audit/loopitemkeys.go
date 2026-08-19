// FILE: cmd/config-key-audit/loopitemkeys.go
//
// --loop-sitewide-item-keys (bugs_open/321): which create_work_item step, filed
// PER ITEM inside a loop, still builds its item_key PER SITE — so every
// iteration after the first collides on idx_swi_dedup and is silently dropped?
//
// create_work_item builds item_key as '<item_key_prefix>_<domain>'
// (create_work_item_action.go:225-234). That key is one-per-site, which is the
// INTENDED dedupe for a once-per-site step — most carriers of item_key_prefix
// are exactly that, and they are not findings. The defect exists only where the
// step executes once per loop item: the second iteration's insert hits the
// unique index, ON CONFLICT swallows it, the loop reports success, and the
// finding is gone. Measured on tool-suggester before migration 493: 40
// suggestions -> 11 work items, ~72% silently lost. The remedy is the action's
// own item_key_suffix_field config key (tool-auditor's two loop steps are the
// proven idiom); this mode reports the loop-nested steps that lack it.
//
// Every clause of the rule mirrors an EXECUTED read, so the check convicts what
// the runtime would actually do rather than what the config looks like:
//
//   - Action == "create_work_item", NESTED, and the CONTAINING step's action is
//     "loop". Top-level steps are excluded structurally (site-wide is the
//     intended dedupe there). A sub_workflow under a non-loop step never
//     executes, so flagging it would report behaviour that does not happen.
//   - item_key_prefix under the action's exact read (a bare `.(string)`,
//     :225): no prefix -> item_key NULL -> outside idx_swi_dedup's partial
//     index -> no collision possible.
//   - the suffix judged by the action's exact honouring test (:251,
//     `f, ok := ...(string); ok && f != ""`): a suffix declared as "" or as a
//     non-string is silently ignored at execution and the site-wide key is
//     used anyway — reported with suffix_declared_but_unhonoured=true, because
//     the author who wrote the key believes the door is closed.
//   - loop-over-SITES guard: when the step's site_id config path is rooted at
//     the loop variable (resolved with LoopAction's own fallback order
//     loop_var -> item_variable -> "loop_item", loop_actions.go:44-52), each
//     iteration writes a DISTINCT site_id and (site_id, item_key) cannot
//     collide — not a finding.
//
// THE DESCENT IS validation.WalkSteps, NOT ITS OWN — the same single walk every
// mode in this binary uses, for bugs_open/144's reason (a second hand-written
// descent goes blind in its own direction; sharedoutputs.go's header records
// this package's own instance of exactly that). The walk populates one
// path->step map; containment is recovered from the qualified path (the parent
// step's path is the path minus its last two segments), never re-derived.
//
// DOCUMENTED LIMITATION: a suffix that resolves but is LOOP-INVARIANT (the same
// value every iteration) passes this check — the config alone cannot know what
// a dotted path resolves to at runtime. The runtime Warn in
// create_work_item_action.go (same lane) is the net under that: it fires when
// an insert deduped away inside a loop.
//
// Wired to scripts/audit-loop-sitewide-item-keys.sh and the
// loop-sitewide-item-key-check CronJob (--report writes one doc_notes row per
// run, clean runs included — a MISSING row means the job did not run and must
// not read as "nothing is wrong").
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

// loopItemKeyFinding is one loop-nested create_work_item step whose item_key is
// still site-wide. LoopVariable is the fix hint: the suffix path almost always
// starts with it (tool-suggester's fix was `<loop_variable>.function`).
type loopItemKeyFinding struct {
	Agent                       string `json:"agent"`
	Path                        string `json:"path"`
	LoopPath                    string `json:"loop_path"`
	LoopVariable                string `json:"loop_variable"`
	ItemType                    string `json:"item_type"`
	ItemKeyPrefix               string `json:"item_key_prefix"`
	SuffixDeclaredButUnhonoured bool   `json:"suffix_declared_but_unhonoured"`
}

// parentStepPath recovers the containing step's qualified path from a nested
// step's path: "steps.create_items_loop.sub_workflow.create_novel_item" ->
// "steps.create_items_loop". Returns "" for a top-level step ("steps.<name>"),
// which is how top-level steps are excluded without a second traversal. Works
// at any depth because WalkSteps qualifies every level the same way
// (<parent>.<container>.<name>, container in {sub_workflow, substeps}).
func parentStepPath(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) < 4 {
		return ""
	}
	return strings.Join(parts[:len(parts)-2], ".")
}

// loopVariableOf resolves a loop step's iteration variable with the SAME
// fallback order the executor uses (loop_actions.go:44-52): loop_var, then
// item_variable, then the default "loop_item". Mirroring the executed order
// matters: judging the site_id guard against a key the runtime would not read
// would convict or acquit the wrong steps.
func loopVariableOf(loopStep models.Step) string {
	if v, ok := loopStep.Config["loop_var"].(string); ok && v != "" {
		return v
	}
	if v, ok := loopStep.Config["item_variable"].(string); ok && v != "" {
		return v
	}
	return "loop_item"
}

// findLoopSitewideItemKeys is the pure check (I/O split off, per this binary's
// convention). One walk per agent; containment recovered from paths.
func findLoopSitewideItemKeys(agents []liveAgent) []loopItemKeyFinding {
	findings := []loopItemKeyFinding{}

	for _, agent := range agents {
		steps := map[string]models.Step{}
		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			steps[path] = step
		})

		for path, step := range steps {
			if step.Action != "create_work_item" {
				continue
			}
			loopPath := parentStepPath(path)
			if loopPath == "" {
				continue // top-level: site-wide is the intended dedupe
			}
			parent, ok := steps[loopPath]
			if !ok || parent.Action != "loop" {
				continue // a sub_workflow under a non-loop step never executes
			}

			// The action's exact prefix read (:225).
			prefix, _ := step.Config["item_key_prefix"].(string)
			if prefix == "" {
				continue // item_key NULL -> outside idx_swi_dedup -> no collision
			}

			// The action's exact honouring test (:251). Honoured -> not a
			// finding (the loop-invariant case is the documented limitation).
			if f, ok := step.Config["item_key_suffix_field"].(string); ok && f != "" {
				continue
			}
			_, declared := step.Config["item_key_suffix_field"]

			// Loop-over-sites guard: a site_id rooted at the loop variable
			// makes (site_id, item_key) distinct per iteration.
			loopVar := loopVariableOf(parent)
			if sid, ok := step.Config["site_id"].(string); ok {
				if strings.SplitN(sid, ".", 2)[0] == loopVar {
					continue
				}
			}

			itemType, _ := step.Config["item_type"].(string)
			findings = append(findings, loopItemKeyFinding{
				Agent:                       agent.Type,
				Path:                        path,
				LoopPath:                    loopPath,
				LoopVariable:                loopVar,
				ItemType:                    itemType,
				ItemKeyPrefix:               prefix,
				SuffixDeclaredButUnhonoured: declared,
			})
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Agent != findings[j].Agent {
			return findings[i].Agent < findings[j].Agent
		}
		return findings[i].Path < findings[j].Path
	})
	return findings
}

// emitLoopSitewideItemKeys is the I/O half: DB route when PG_CLIENTS_HOST is
// set (the CronJob), stdin otherwise (the wrapper script). Same refusals as
// every mode here: a DB error when the env asked for the DB is fatal, and a
// 0-agent fleet is never reported as clean.
func emitLoopSitewideItemKeys() {
	report := false
	for _, a := range os.Args[2:] {
		switch a {
		case "--report":
			report = true
		default:
			fmt.Fprintf(os.Stderr, "config-key-audit --loop-sitewide-item-keys: unknown argument %q\n", a)
			os.Exit(2)
		}
	}

	var agents []liveAgent
	var failed int
	var err error
	if db, derr := dbConn(); derr != nil || db != nil {
		if derr != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --loop-sitewide-item-keys: %v\n", derr)
			os.Exit(2)
		}
		defer db.Close()
		agents, failed, err = loadLiveAgentsFromDB(db, "--loop-sitewide-item-keys")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	} else {
		raw, rerr := io.ReadAll(os.Stdin)
		if rerr != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --loop-sitewide-item-keys: reading stdin: %v\n", rerr)
			os.Exit(2)
		}
		agents, failed, err = decodeLiveAgents(raw, "--loop-sitewide-item-keys")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --loop-sitewide-item-keys: 0 live agents decoded — refusing to print a clean report over an empty fleet.\n")
		os.Exit(2)
	}

	findings := findLoopSitewideItemKeys(agents)

	out := map[string]interface{}{
		"agents_scanned":   len(agents),
		"agents_undecoded": failed,
		"findings":         findings,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)

	if report {
		writeDocNote("loop-sitewide-item-keys",
			loopItemKeyRunSummary(len(agents), failed, findings),
			"config-integrity", "loop_sitewide_item_key_check")
	}

	if len(findings) > 0 {
		fmt.Fprintf(os.Stderr,
			"%d loop-nested create_work_item step(s) still key per SITE: every iteration after the first is silently dropped (bugs_open/321).\n",
			len(findings))
		os.Exit(1)
	}
}

// loopItemKeyRunSummary is the doc_notes body — prose stating the SCOPE as well
// as the result, because "0 findings over 3 agents" and "0 findings over 177"
// have opposite meanings (see sharedOutputRunSummary, same convention).
func loopItemKeyRunSummary(scanned, undecoded int, findings []loopItemKeyFinding) string {
	var b strings.Builder
	if len(findings) == 0 {
		fmt.Fprintf(&b, "loop-sitewide-item-keys check CLEAN: every loop-nested create_work_item across %d live agents keys per item.", scanned)
	} else {
		fmt.Fprintf(&b, "loop-sitewide-item-keys: %d loop-nested step(s) across %d live agents still key per SITE — iterations 2..N of each batch are silently dropped (bugs_open/321): ", len(findings), scanned)
		for i, f := range findings {
			if i > 0 {
				b.WriteString("; ")
			}
			fmt.Fprintf(&b, "%s %s (prefix %s, loop var %s)", f.Agent, f.Path, f.ItemKeyPrefix, f.LoopVariable)
		}
		b.WriteString(".")
	}
	if undecoded > 0 {
		fmt.Fprintf(&b, " %d agent row(s) failed to decode and were not scanned.", undecoded)
	}
	return b.String()
}
