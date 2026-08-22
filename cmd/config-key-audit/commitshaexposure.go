// FILE: cmd/config-key-audit/commitshaexposure.go
//
// The STANDING form of migration 537's apply-time guard (bugs_closed/334;
// staged_component_build handoff 2026-08-21 §1 item 4). Which live handler can
// produce a commit of its own but does not expose it at the standard path
// `response.commit_sha` via its complete step's `result_mapping`?
//
// WHY THIS EXISTS. 537 wired build-dispatch-loop's `mark_complete` to
// `"commit_sha?": "handler_result.response.commit_sha"` — the handler's own
// reply, or nothing; never the whole-tree search. Its guard proved, AT APPLY
// TIME, that every handler which could produce a commit already exposed one
// (it fired once, correctly, naming content-feed-orchestrator; migration 540
// cleared it). But the handler population is per-item and DYNAMIC: a NEW
// commit-producing handler appearing after the apply that does not expose
// `response.commit_sha` will simply never record `result.commit_sha` — no
// error, no log, no row, nothing to notice, because with the `?` wire absence
// IS the contract. The guard ran once; the population moves daily. This mode
// asks the guard's question on a schedule, exactly like the `?` adoption gate
// (--optional-explicit-wires) and RFC_022's budget check do for their own
// once-was-manual questions.
//
// THE QUESTION IS 537's, DELIBERATELY UNCHANGED. Three sets, verbatim from the
// guard (537_bdl_mark_complete_declares_commit_sha.sql):
//
//	producers    — agents whose OWN orchestrations carry a `commit_sha` in
//	               `collected_data`, last 30 days. A property of the HANDLER,
//	               not of what an item recorded — counting recorded items
//	               counts the very contamination the wire removed, which cost
//	               537's author a false positive (its header, "demand census"
//	               correction).
//	handlers     — agents named in `site_work_items.handler_agent`, last
//	               7 days: the empirical dispatch population.
//	exposed      — agents whose `complete_workflow` step carries `commit_sha`
//	               in `result_mapping`. The ONE deliberate difference from the
//	               guard's SQL: the guard walked top-level steps only; this
//	               walks with validation.WalkSteps (nested included), because a
//	               second hand-written traversal blind to nesting is
//	               bugs_open/144, and a handler whose complete step moves into
//	               a sub-workflow should not start paging as unexposed.
//
// A finding is (producers ∩ handlers) \ exposed. The failure direction matches
// the guard's: a handler wrongly flagged pages loudly and costs a look (cheap);
// a handler wrongly passed loses a field silently for ever (expensive).
//
// ACKS: a handler that genuinely should NOT expose a commit (none is known
// today) gets an entry in the acks file stating the reason — the same
// deliberate-exception shape as optional_explicit_wire_acks.json, keyed on the
// agent type. An entry with an empty `reason` does not count.
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
	"github.com/gqls/agentchassis/platform/validation"
)

// commitProducersQuery and liveHandlersQuery are character-for-character the
// CTEs of 537's guard (windows included). If the guard's definition of "can
// produce a commit" ever changes, change BOTH — a divergence means the apply
// gate and the standing gate answer different questions under one name.
const commitProducersQuery = `
SELECT COALESCE(jsonb_agg(DISTINCT owner_agent_type), '[]'::jsonb)
  FROM orchestration_states
 WHERE created_at > now() - interval '30 days'
   AND collected_data::text LIKE '%commit_sha%';`

const liveHandlersQuery = `
SELECT COALESCE(jsonb_agg(DISTINCT handler_agent), '[]'::jsonb)
  FROM site_work_items
 WHERE handler_agent IS NOT NULL AND handler_agent <> ''
   AND created_at > now() - interval '7 days';`

// commitShaExposureFinding is one member of (producers ∩ handlers): a handler
// that can produce a commit and is live in dispatch. Exposed handlers are
// reported too — "8 handlers, all exposed" and "0 handlers in the window" are
// both clean and a reader must be able to tell them apart.
type commitShaExposureFinding struct {
	Agent   string `json:"agent"`
	Exposed bool   `json:"exposed"`
	// Acknowledged is true when the acks file carries this agent with a
	// non-empty reason: a human decided this handler deliberately does not
	// expose a commit, and said why.
	Acknowledged bool `json:"acknowledged"`
}

type commitShaExposureAck struct {
	Reason string `json:"reason"`
	Date   string `json:"date"`
	Review string `json:"review"`
}

func loadCommitShaExposureAcks(path string) (map[string]bool, error) {
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
			continue // documentation keys, as in the sibling acks files
		}
		var ack commitShaExposureAck
		if err := json.Unmarshal(blob, &ack); err != nil {
			return nil, fmt.Errorf("acks entry %q: %w", key, err)
		}
		if strings.TrimSpace(ack.Reason) == "" {
			fmt.Fprintf(os.Stderr,
				"config-key-audit --commit-sha-exposure: acks entry %q has an empty "+
					"`reason` — ignoring it; an exception must say why the handler "+
					"deliberately exposes no commit.\n", key)
			continue
		}
		acked[key] = true
	}
	return acked, nil
}

// exposesCommitSha reports whether one agent's workflow carries a
// `complete_workflow` step (any depth) whose `result_mapping` maps
// `commit_sha`. Split out so the DECISION is testable on fixtures, per this
// lane's own trap 3 (a test that pins a guard's inputs does not pin the guard).
func exposesCommitSha(agent liveAgent) bool {
	exposed := false
	validation.WalkSteps(agent.Workflow, func(_ string, step models.Step, _ bool) {
		if step.Action != "complete_workflow" {
			return
		}
		rm, ok := step.Config["result_mapping"].(map[string]interface{})
		if !ok {
			return
		}
		if _, has := rm["commit_sha"]; has {
			exposed = true
		}
	})
	return exposed
}

// commitShaExposureFindings is the pure check: every member of
// (producers ∩ handlers), with its exposure and acknowledgement state.
func commitShaExposureFindings(agents []liveAgent, producers, handlers []string,
	acked map[string]bool) []commitShaExposureFinding {

	exposed := make(map[string]bool, len(agents))
	for _, agent := range agents {
		if exposesCommitSha(agent) {
			exposed[agent.Type] = true
		}
	}

	handlerSet := make(map[string]bool, len(handlers))
	for _, h := range handlers {
		handlerSet[h] = true
	}

	findings := []commitShaExposureFinding{}
	seen := map[string]bool{}
	for _, p := range producers {
		if !handlerSet[p] || seen[p] {
			continue
		}
		seen[p] = true
		findings = append(findings, commitShaExposureFinding{
			Agent:        p,
			Exposed:      exposed[p],
			Acknowledged: acked[p],
		})
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i].Agent < findings[j].Agent })
	return findings
}

// fleetExposedCount is the vacuity input for the exposed set: how many live
// agents anywhere in the fleet expose commit_sha via result_mapping.
func fleetExposedCount(agents []liveAgent) int {
	n := 0
	for _, agent := range agents {
		if exposesCommitSha(agent) {
			n++
		}
	}
	return n
}

// commitShaExposureRefusal returns the refusal message when this run cannot
// mean anything, and "" when it can. Every branch here is a state in which the
// clean report and the blind report are byte-identical, which is exactly what a
// scheduled check must never emit (the lane's standing lesson: as its fixes
// land, a check destroys its own instrument-alive control — so the controls
// must be structural, not observational):
//
//   - zero commit producers in 30 days: hundreds of items a day record a sha,
//     so an empty producers set means the orchestration_states query went
//     blind, not that the fleet stopped committing;
//   - zero dispatch handlers in 7 days: dispatch runs daily, so an empty set
//     means the site_work_items query went blind;
//   - zero exposing agents fleet-wide: 537's own second guard — the handler
//     standardisation (migrations 519–540) was rolled back or the walk went
//     blind, and reporting over that state is how a guard becomes decoration.
//
// The agents/walked-steps vacuity is vacuityRefusal (optionalexplicitwires.go),
// checked by the caller before this.
func commitShaExposureRefusal(producers, handlers []string, exposedCount int) string {
	if len(producers) == 0 {
		return "config-key-audit --commit-sha-exposure: ZERO commit-producing agents in the " +
			"30-day window — the producers query has gone blind (items record shas daily), " +
			"refusing to report.\n"
	}
	if len(handlers) == 0 {
		return "config-key-audit --commit-sha-exposure: ZERO dispatch handlers in the 7-day " +
			"window — the handlers query has gone blind (dispatch runs daily), refusing to " +
			"report.\n"
	}
	if exposedCount == 0 {
		return "config-key-audit --commit-sha-exposure: NO live agent exposes commit_sha via " +
			"result_mapping. Either the 519–540 handler standardisation was rolled back or " +
			"the workflow walk has gone blind — refusing to report over either.\n"
	}
	return ""
}

// commitShaExposureStdin is the hand-run input shape, built by
// scripts/audit-commit-sha-exposure.sh: the fleet export plus the two set
// queries, in one object. A bare array (the sibling modes' shape) is refused
// with a pointer at the wrapper, because without the sets this mode has no
// question to ask.
type commitShaExposureStdin struct {
	Agents    []json.RawMessage `json:"agents"`
	Producers []string          `json:"producers"`
	Handlers  []string          `json:"handlers"`
}

func unexposedUnackedCount(findings []commitShaExposureFinding) int {
	n := 0
	for _, f := range findings {
		if !f.Exposed && !f.Acknowledged {
			n++
		}
	}
	return n
}

// commitShaExposureRunSummary is the doc_notes body and the --report stdout.
func commitShaExposureRunSummary(agents, failed int, acksPath string,
	findings []commitShaExposureFinding) string {

	exposed, acked := 0, 0
	for _, f := range findings {
		if f.Exposed {
			exposed++
		} else if f.Acknowledged {
			acked++
		}
	}
	bad := unexposedUnackedCount(findings)
	var b strings.Builder
	fmt.Fprintf(&b, "commit-sha-exposure: %d commit-producing handler(s) live in dispatch "+
		"(30d producers ∩ 7d handlers), %d exposing response.commit_sha via result_mapping, "+
		"%d acknowledged exception(s), %d UNEXPOSED (%d agents, %d undecodable; acks=%s)\n",
		len(findings), exposed, acked, bad, agents, failed, acksPath)
	for _, f := range findings {
		state := "ok"
		if !f.Exposed {
			state = "UNEXPOSED"
			if f.Acknowledged {
				state = "ack"
			}
		}
		fmt.Fprintf(&b, "  [%s] %s\n", state, f.Agent)
	}
	if bad > 0 {
		fmt.Fprintf(&b, "  An UNEXPOSED handler's items will simply never record "+
			"result.commit_sha — no error, no row, nothing to notice (the ? wire's absence "+
			"contract). Convert its complete step to result_mapping exposure like migrations "+
			"519–540, or record a deliberate exception with its reason in %s.\n", acksPath)
	}
	return b.String()
}

// emitCommitShaExposure: [--acks <file>] [--report]. Exit codes as the sibling
// scheduled checks: 0 = every commit-producing live handler exposes or is
// acknowledged; 1 = at least one is unexposed and unacknowledged; 2 = the
// check did not run, which must never read as a pass.
func emitCommitShaExposure(args []string) {
	acksPath := "docs/agent_docs/docs024_key_docs_latest/architecture_review/commit_sha_exposure_acks.json"
	report := false
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--acks":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr,
					"config-key-audit --commit-sha-exposure: --acks needs a file path")
				os.Exit(2)
			}
			acksPath = args[i+1]
			i++
		case args[i] == "--report":
			report = true
		default:
			fmt.Fprintf(os.Stderr,
				"config-key-audit --commit-sha-exposure: unrecognised argument %q "+
					"(want: [--acks <file>] [--report])\n", args[i])
			os.Exit(2)
		}
	}

	acked, err := loadCommitShaExposureAcks(acksPath)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --commit-sha-exposure: acks file %q: %v — refusing to run "+
				"without it.\n", acksPath, err)
		os.Exit(2)
	}

	var (
		agents    []liveAgent
		failed    int
		producers []string
		handlers  []string
	)
	if report {
		db, err := dbConn()
		if err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --commit-sha-exposure: %v\n", err)
			os.Exit(2)
		}
		if db == nil {
			fmt.Fprintln(os.Stderr,
				"config-key-audit --commit-sha-exposure --report: PG_CLIENTS_HOST is not set, "+
					"so there is no fleet to read. In the CronJob this comes from the pod env; "+
					"by hand, use scripts/audit-commit-sha-exposure.sh instead.")
			os.Exit(2)
		}
		defer db.Close()
		agents, failed, err = loadLiveAgentsFromDB(db, "--commit-sha-exposure")
		if err == nil {
			producers, err = loadStringSet(db, commitProducersQuery, "producers")
		}
		if err == nil {
			handlers, err = loadStringSet(db, liveHandlersQuery, "handlers")
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --commit-sha-exposure: %v\n", err)
			os.Exit(2)
		}
	} else {
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --commit-sha-exposure: reading stdin: %v\n", err)
			os.Exit(2)
		}
		var in commitShaExposureStdin
		if err := json.Unmarshal(raw, &in); err != nil || in.Agents == nil {
			fmt.Fprintln(os.Stderr,
				"config-key-audit --commit-sha-exposure: stdin must be one object "+
					`{"agents": [...], "producers": [...], "handlers": [...]} — the fleet export `+
					"alone cannot answer this mode's question. Use "+
					"scripts/audit-commit-sha-exposure.sh, which builds it.")
			os.Exit(2)
		}
		agentsRaw, _ := json.Marshal(in.Agents)
		agents, failed, err = decodeLiveAgents(agentsRaw, "--commit-sha-exposure")
		if err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --commit-sha-exposure: %v\n", err)
			os.Exit(2)
		}
		producers, handlers = in.Producers, in.Handlers
	}

	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --commit-sha-exposure: 0 agents decoded (%d undecodable) — "+
				"refusing to print a clean report over an empty or broken export.\n", failed)
		os.Exit(2)
	}
	if msg := vacuityRefusal(agents); msg != "" {
		fmt.Fprint(os.Stderr, msg)
		os.Exit(2)
	}
	if msg := commitShaExposureRefusal(producers, handlers, fleetExposedCount(agents)); msg != "" {
		fmt.Fprint(os.Stderr, msg)
		os.Exit(2)
	}

	findings := commitShaExposureFindings(agents, producers, handlers, acked)

	if report {
		summary := commitShaExposureRunSummary(len(agents), failed, acksPath, findings)
		fmt.Print(summary)
		// ONE row per run, clean or not — a check that only speaks when it fails
		// is indistinguishable from one that has stopped running.
		writeDocNote("commit-sha-exposure", summary,
			"commit-sha-exposure", "commit-sha-exposure-check")
		if unexposedUnackedCount(findings) > 0 {
			os.Exit(1)
		}
		return
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(findings); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --commit-sha-exposure: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprint(os.Stderr, commitShaExposureRunSummary(len(agents), failed, acksPath, findings))
	if unexposedUnackedCount(findings) > 0 {
		os.Exit(1)
	}
}

// loadStringSet scans one jsonb-array-returning query into a []string.
func loadStringSet(db *sql.DB, query, what string) ([]string, error) {
	var raw []byte
	if err := db.QueryRow(query).Scan(&raw); err != nil {
		return nil, fmt.Errorf("%s query failed: %w", what, err)
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("%s query returned non-array: %w", what, err)
	}
	return out, nil
}
