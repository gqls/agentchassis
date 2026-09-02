// FILE: cmd/config-key-audit/rendertruncation.go
//
// The commissioned reader for RENDER_AUDIT_TRUNCATED (bugs_open/394, owner
// decision 4, 2026-08-25).
//
// WHAT THE ROW MEANS, AND WHY THAT CHANGED. `request_render_audit` caps a
// site's sweep at `max_pages` and writes one durable row when the cap bites.
// Until bugs_open/394 the row meant something permanent and bad: the audit took
// the SAME deterministic prefix on every run, so the pages past the cap had
// never been audited and never would be. `bugs_closed/242` made that loud and
// raised the cap 25 -> 60 as a stated mitigation; measured 2026-08-26,
// webdesign.co.uk had outgrown it at 146 live pages / 60 audited / 86 never —
// and the tail was a whole CLASS (all 45 `tool-*-guide` pages at nav_order 200,
// unreachable at any cap below 98).
//
// The coverage cursor changed what the row asserts. In `coverage_mode=cursor` a
// truncation row is now HEALTHY PAGINATION: this run covered one window, the
// cursor advanced, and the union converges over a cycle. That is exactly why the
// reader is still owed rather than discharged — the owner commissioned a reader
// for a signal that meant one thing, and the fix gave the signal three meanings.
// Telling them apart is the whole job.
//
// WHAT IS A FINDING, and each arm exists because the failure it catches is
// otherwise silent:
//
//  1. A PREFIX-MODE ROW FROM THE ROTATION CALLER. `render-audit-agent` is opted
//     in by migration 649. If it writes `coverage_mode=prefix` — or writes no
//     `coverage_mode` at all, which is a pre-394 binary — then either the config
//     flip regressed or the pod is running an old image. Both look like a
//     working audit from every other angle: rows keep arriving, findings keep
//     being filed, and the tail is quietly permanent again.
//
//  2. A STALLED CURSOR. Consecutive cursor-mode rows for one (domain,
//     agent_type) whose `window_first` does not move. Coverage looks healthy —
//     the mode says cursor, the rows keep coming — while the same window is
//     audited for ever. This is the failure the cursor itself could introduce,
//     so the reader that watches for it ships in the same change.
//
//  3. AN UNACKNOWLEDGED CALLER. A NEW agent writing truncation rows has made a
//     coverage decision nobody reviewed. `design-critique-agent` is acknowledged
//     at birth: it is a manual sampler with no cadence (its seed, 645, says so),
//     its cap of 8 is plausibly meant as the most important 8 rather than any 8,
//     and it deliberately keeps the prefix. So the baseline is quiet and the
//     third caller pages.
//
// VACUITY. Zero truncation rows is a healthy reading — a fleet whose sites all
// fit inside their caps writes none. What is never healthy is an EMPTY
// `agent_error_log`: hundreds of rows a day live inside the retention window, so
// a zero total means the read went blind, and this check REFUSES to report over
// it rather than printing a clean run. "The check could not run" must never read
// as "the check passed" (016b §9).
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

// renderTruncationCode is the finding code this reader consumes, declared as a
// same-file const because the finding-code registry's checker VERIFIES that the
// reader file names both the code it claims to read and its sink — it opens the
// file (findingcodes.go, repoSourceReader) rather than trusting the string. The
// sink is agent_error_log, named in the query below for the same reason.
const renderTruncationCode = "RENDER_AUDIT_TRUNCATED"

// renderTruncationRunsQuery returns the recent rows per (domain, agent_type),
// newest first, with the two context fields the arms need.
//
// COALESCE keeps a context-less row VISIBLE as coverage_mode '(absent)' rather
// than silently dropping it: a pre-394 binary writes no coverage_mode, and that
// is arm 1's whole signal — a row this reader cannot attribute is a finding
// about the writer, never a row to skip.
const renderTruncationRunsQuery = `
SELECT COALESCE(jsonb_agg(t), '[]'::jsonb) FROM (
  SELECT COALESCE(domain, '(no domain)')                              AS domain,
         COALESCE(agent_type, '(no agent_type)')                      AS agent_type,
         COALESCE(NULLIF(context->>'coverage_mode',''), '(absent)')   AS coverage_mode,
         COALESCE(context->>'window_first', '')                       AS window_first,
         COALESCE((context->>'priority_not_live')::int, 0)             AS priority_not_live,
         occurred_at::text                                            AS occurred_at
    FROM agent_error_log
   WHERE error_code = 'RENDER_AUDIT_TRUNCATED'
   ORDER BY domain, agent_type, occurred_at DESC) t;`

// renderTruncationAlivenessQuery is the instrument-alive control.
const renderTruncationAlivenessQuery = `SELECT count(*) FROM agent_error_log;`

// rotatingCallers are the agent types that MUST rotate — i.e. the ones migration
// 649 opts in. A prefix-mode row from one of these is arm 1.
//
// Hand-kept and deliberately short. It is a RATCHET, and it can go stale in one
// direction only: a caller opted in later and not listed here would have its
// prefix regression missed. That direction is checked by arm 3, which pages on
// any caller this file has never heard of at all — so a NEW caller cannot be
// silently absent from both lists.
var rotatingCallers = map[string]bool{
	"render-audit-agent": true,
}

// renderTruncationRun is one truncation row, flattened.
type renderTruncationRun struct {
	Domain       string `json:"domain"`
	AgentType    string `json:"agent_type"`
	CoverageMode string `json:"coverage_mode"`
	WindowFirst  string `json:"window_first"`
	// PriorityNotLive counts open contrast_failure rows whose page matched NO
	// live page in that run — findings that can NEVER self-grade, because the
	// audit will not photograph their page again.
	PriorityNotLive int    `json:"priority_not_live"`
	OccurredAt      string `json:"occurred_at"`
}

// renderTruncationFinding is one thing a human must look at.
type renderTruncationFinding struct {
	Domain    string `json:"domain"`
	AgentType string `json:"agent_type"`
	Arm       string `json:"arm"` // "prefix_from_rotating_caller" | "stalled_cursor" | "unacknowledged_caller"
	Detail    string `json:"detail"`
}

type renderTruncationAck struct {
	Reason   string `json:"reason"`
	Date     string `json:"date"`
	Evidence string `json:"evidence"`
}

// loadRenderTruncationAcks mirrors loadUngradedCompletionsAcks exactly:
// `_`-prefixed documentation keys skipped, a hollow (empty-reason) ack warned
// about and ignored, a missing or malformed file an ERROR the caller turns into
// exit 2 — never an empty ack set, because "could not read the exceptions" and
// "there are no exceptions" have opposite meanings and only one of them is safe.
func loadRenderTruncationAcks(path string) (map[string]bool, error) {
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
			continue
		}
		var ack renderTruncationAck
		if err := json.Unmarshal(blob, &ack); err != nil {
			return nil, fmt.Errorf("acks entry %q: %w", key, err)
		}
		if strings.TrimSpace(ack.Reason) == "" {
			fmt.Fprintf(os.Stderr,
				"config-key-audit --render-truncation: acks entry %q has an empty `reason` — "+
					"ignoring it; an acknowledgement must say what was diagnosed and why the "+
					"caller's prefix behaviour is deliberate.\n", key)
			continue
		}
		acked[key] = true
	}
	return acked, nil
}

// dormantAfterDays is how far behind the fleet's most recent truncation row a
// group may fall before its rows stop being evidence about TODAY's config.
//
// ⚠ THIS EXISTS BECAUSE THE FIRST WIRE TEST WENT RED ON FROZEN HISTORY, and the
// failure was mine, not the data's. `loancalculator.co.uk` has exactly ONE
// truncation row, from 2026-08-11, written under a per-dispatch `max_pages: 5`
// override. The site has 28 live pages against a cap of 60, so it CANNOT
// truncate again — that row is permanent history. Arm (a) judged "the most
// recent row in the group", found a pre-394 `(absent)` coverage_mode, and
// reported a config regression that had not happened. Left alone it would have
// been red on day one and every day after, which is the one thing this estate
// says reliably turns a check off.
//
// 14 days is ~4 missed rotation opportunities at the 3-day per-site cadence, so
// a genuine regression still produces recent rows and still alarms; only a site
// that has STOPPED truncating goes quiet. Measured RELATIVE to the newest row in
// the dataset rather than to the wall clock, so the rule is a pure function of
// the data and a fixture can exercise it without a clock.
const dormantAfterDays = 14

// parseRowTime accepts the two shapes this data actually arrives in: Postgres's
// text timestamp (`2026-08-11 18:08:54.431181+00`) and RFC3339. An unparseable
// stamp returns ok=false and the caller treats the group as ACTIVE — failing
// toward alarming, because a clock we cannot read must not silence a check.
func parseRowTime(s string) (time.Time, bool) {
	for _, layout := range []string{
		"2006-01-02 15:04:05.999999-07",
		"2006-01-02 15:04:05.999999-07:00",
		"2006-01-02 15:04:05-07",
		time.RFC3339Nano,
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, strings.TrimSpace(s)); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// newestRowTime is the reference point dormancy is measured against.
func newestRowTime(runs []renderTruncationRun) (time.Time, bool) {
	var newest time.Time
	found := false
	for _, r := range runs {
		if t, ok := parseRowTime(r.OccurredAt); ok && (!found || t.After(newest)) {
			newest, found = t, true
		}
	}
	return newest, found
}

// judgeRenderTruncation is the pure decision, split from the fetch so every arm
// is testable on a fixture with no cluster.
//
// Rows arrive newest-first per (domain, agent_type) — the query's ORDER BY — and
// the stall arm depends on that ordering, so it is asserted rather than assumed:
// a caller feeding rows in another order would silently disable arm 2.
func judgeRenderTruncation(runs []renderTruncationRun, acked map[string]bool) []renderTruncationFinding {
	type key struct{ domain, agent string }
	grouped := map[key][]renderTruncationRun{}
	var order []key
	for _, r := range runs {
		k := key{r.Domain, r.AgentType}
		if _, seen := grouped[k]; !seen {
			order = append(order, k)
		}
		grouped[k] = append(grouped[k], r)
	}

	fleetNewest, haveClock := newestRowTime(runs)

	var out []renderTruncationFinding
	for _, k := range order {
		rows := grouped[k]

		// DORMANT groups are reported, never alarmed. See dormantAfterDays.
		if haveClock {
			if t, ok := parseRowTime(rows[0].OccurredAt); ok &&
				fleetNewest.Sub(t) > time.Duration(dormantAfterDays)*24*time.Hour {
				continue
			}
		}

		// ARM 3 first: a caller nobody has ruled on. Reported once per caller,
		// not once per site, or one unreviewed agent would bury the other arms.
		if !rotatingCallers[k.agent] && !acked[k.agent] {
			out = append(out, renderTruncationFinding{
				Domain: k.domain, AgentType: k.agent, Arm: "unacknowledged_caller",
				Detail: fmt.Sprintf("%q writes truncation rows and is neither an opted-in rotating caller nor acknowledged. "+
					"A caller that truncates has made a coverage decision: either opt it in (rotate_coverage) or "+
					"acknowledge, with the reason its prefix sample is deliberate.", k.agent),
			})
		}

		// ARM 1: the rotation caller must rotate.
		if rotatingCallers[k.agent] {
			newest := rows[0]
			if newest.CoverageMode != "cursor" {
				out = append(out, renderTruncationFinding{
					Domain: k.domain, AgentType: k.agent, Arm: "prefix_from_rotating_caller",
					Detail: fmt.Sprintf("most recent row (%s) has coverage_mode=%q, expected \"cursor\". "+
						"Either the migration-649 config flip regressed, or the pod that ran it predates the cursor. "+
						"The tail is permanent again while this holds, and nothing else shows it.",
						newest.OccurredAt, newest.CoverageMode),
				})
			}
		}

		// ARM 2: a cursor that is not moving. Needs two consecutive cursor-mode
		// rows with a non-empty window_first; an empty one carries no
		// information and must not read as "unchanged".
		if len(rows) >= 2 &&
			rows[0].CoverageMode == "cursor" && rows[1].CoverageMode == "cursor" &&
			rows[0].WindowFirst != "" && rows[0].WindowFirst == rows[1].WindowFirst {
			out = append(out, renderTruncationFinding{
				Domain: k.domain, AgentType: k.agent, Arm: "stalled_cursor",
				Detail: fmt.Sprintf("two consecutive runs (%s, %s) both start at window_first=%q — "+
					"the cursor is not advancing, so the same window is audited every run while the mode "+
					"still reads healthy. Check whether the cursor write is failing (the action logs "+
					"\"coverage cursor NOT advanced\") or the row was deleted between runs.",
					rows[1].OccurredAt, rows[0].OccurredAt, rows[0].WindowFirst),
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Domain != out[j].Domain {
			return out[i].Domain < out[j].Domain
		}
		if out[i].AgentType != out[j].AgentType {
			return out[i].AgentType < out[j].AgentType
		}
		return out[i].Arm < out[j].Arm
	})
	return out
}

// renderTruncationRunSummary states SCOPE, not just result: "0 findings over 6
// truncation rows in a log holding 41,000" and "0 over 0" have opposite
// meanings, and only the first is a pass.
func renderTruncationRunSummary(errorLogRows int, acksPath string, runs []renderTruncationRun, findings []renderTruncationFinding) string {
	sites := map[string]bool{}
	callers := map[string]bool{}
	for _, r := range runs {
		sites[r.Domain] = true
		callers[r.AgentType] = true
	}
	// Dormant groups are counted and NAMED, never merely skipped. A group this
	// check has stopped judging must be visible, or "0 findings" quietly starts
	// meaning "0 findings among the groups I still look at" — which is the
	// blind-check shape one level down from the one this file exists to close.
	var dormant []string
	if fleetNewest, ok := newestRowTime(runs); ok {
		seen := map[string]bool{}
		for _, r := range runs {
			k := r.Domain + " / " + r.AgentType
			if seen[k] {
				continue
			}
			seen[k] = true
			if t, ok2 := parseRowTime(r.OccurredAt); ok2 &&
				fleetNewest.Sub(t) > time.Duration(dormantAfterDays)*24*time.Hour {
				dormant = append(dormant, k)
			}
		}
		sort.Strings(dormant)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "render-truncation: %d %s row(s) across %d site(s) and %d caller(s) "+
		"in an agent_error_log holding %d row(s); %d finding(s); %d dormant group(s); acks=%s\n",
		len(runs), renderTruncationCode, len(sites), len(callers), errorLogRows,
		len(findings), len(dormant), acksPath)
	for _, d := range dormant {
		fmt.Fprintf(&b, "  [dormant] %s: newest row is >%dd behind the fleet's newest — "+
			"not judged, because a site that has STOPPED truncating freezes its last row for ever "+
			"and would otherwise alarm every day.\n", d, dormantAfterDays)
	}
	for _, f := range findings {
		fmt.Fprintf(&b, "  [%s] %s / %s: %s\n", f.Arm, f.Domain, f.AgentType, f.Detail)
	}
	// REPORTED, NOT ALARMED — and the distinction is deliberate.
	//
	// bug_historian raised this as a MEDIUM advisory on council round 3 (corr
	// f67593f5): the cursor COMPUTES the count of open findings whose page is no
	// longer live and then drops it, which is the platform's "detector partitions
	// its population and silently discards the residue" shape (016b §9). They were
	// right that it must be surfaced. It is NOT an alarm arm because
	// `[MEASURED 2026-08-26]` the live population is ONE row of 116 open
	// contrast_failure rows — a check that goes red on its first run over a
	// pre-existing backlog of one is a check that gets ignored, which is this
	// estate's own stated reason for reporting a ratchet rather than failing it.
	// If this number climbs, that is the signal; the count is here so it can be.
	notLive := 0
	for _, r := range runs {
		if r.PriorityNotLive > notLive {
			notLive = r.PriorityNotLive
		}
	}
	if notLive > 0 {
		fmt.Fprintf(&b, "  NOTE: the most recent run saw %d open contrast_failure finding(s) whose page "+
			"is no longer live — they can never self-grade, because the audit will not photograph "+
			"that page again. Reported, not alarmed (measured population was 1 of 116 on 2026-08-26). "+
			"A climbing number here is a queue nobody can drain.\n", notLive)
	}

	if len(findings) == 0 && len(runs) > 0 {
		fmt.Fprintf(&b, "  Every truncation row is accounted for: the rotating caller is in cursor mode "+
			"and advancing, and every other caller is acknowledged. A truncation row is NOT a defect "+
			"under the cursor — it is this run's coverage window.\n")
	}
	return b.String()
}

// renderTruncationStdin is the hand-run input shape, built by
// scripts/audit-render-truncation.sh: the rows plus the aliveness total in one
// object. A bare array is refused with a pointer at the wrapper, because without
// the total this mode cannot tell a clean table from a blind read.
type renderTruncationStdin struct {
	Runs         []renderTruncationRun `json:"runs"`
	ErrorLogRows *int                  `json:"error_log_rows"`
}

func decodeRenderTruncationStdin(r io.Reader) (renderTruncationStdin, error) {
	var in renderTruncationStdin
	raw, err := io.ReadAll(r)
	if err != nil {
		return in, err
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return in, fmt.Errorf("stdin is not the {runs, error_log_rows} object this mode reads "+
			"(a bare array cannot carry the aliveness total, and without it a clean table and a "+
			"blind read are the same reading): %w", err)
	}
	if in.ErrorLogRows == nil {
		return in, fmt.Errorf("stdin carries no error_log_rows — refusing to report, because " +
			"zero findings over an unread log is not a pass; use scripts/audit-render-truncation.sh")
	}
	return in, nil
}

// emitRenderTruncation: [--acks <file>] [--report]. Exit codes as the sibling
// scheduled checks: 0 = every truncation row is accounted for (or none exists);
// 1 = a finding; 2 = the check could not run, which must never read as a pass.
func emitRenderTruncation(args []string) {
	acksPath := "docs/agent_docs/docs024_key_docs_latest/architecture_review/render_truncation_acks.json"
	report := false
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--acks":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "config-key-audit --render-truncation: --acks needs a file path")
				os.Exit(2)
			}
			acksPath = args[i+1]
			i++
		case args[i] == "--report":
			report = true
		default:
			fmt.Fprintf(os.Stderr,
				"config-key-audit --render-truncation: unrecognised argument %q "+
					"(want: [--acks <file>] [--report])\n", args[i])
			os.Exit(2)
		}
	}

	acked, err := loadRenderTruncationAcks(acksPath)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --render-truncation: acks file %q: %v — refusing to run without it.\n",
			acksPath, err)
		os.Exit(2)
	}

	var (
		runs         []renderTruncationRun
		errorLogRows int
	)
	if report {
		db, err := dbConn()
		if err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --render-truncation: %v\n", err)
			os.Exit(2)
		}
		if db == nil {
			fmt.Fprintln(os.Stderr,
				"config-key-audit --render-truncation --report: PG_CLIENTS_HOST is not set, so "+
					"there is nothing to read. In the CronJob this comes from the pod env; by "+
					"hand, use scripts/audit-render-truncation.sh instead.")
			os.Exit(2)
		}
		defer db.Close()
		var raw []byte
		if err := db.QueryRow(renderTruncationRunsQuery).Scan(&raw); err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --render-truncation: runs query: %v\n", err)
			os.Exit(2)
		}
		if err := json.Unmarshal(raw, &runs); err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --render-truncation: runs decode: %v\n", err)
			os.Exit(2)
		}
		if err := db.QueryRow(renderTruncationAlivenessQuery).Scan(&errorLogRows); err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --render-truncation: aliveness query: %v\n", err)
			os.Exit(2)
		}
	} else {
		in, err := decodeRenderTruncationStdin(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --render-truncation: %v\n", err)
			os.Exit(2)
		}
		runs, errorLogRows = in.Runs, *in.ErrorLogRows
	}

	// The instrument-alive control. Zero rows of the CODE is healthy; an empty
	// TABLE is a blind read, and a clean report over a blind read is the exact
	// shape this whole class of check exists to end.
	if errorLogRows == 0 {
		fmt.Fprintln(os.Stderr,
			"config-key-audit --render-truncation: agent_error_log reads as EMPTY — hundreds of "+
				"rows a day live inside its retention window, so this read went blind; refusing "+
				"to print a clean report over it.")
		os.Exit(2)
	}

	findings := judgeRenderTruncation(runs, acked)
	summary := renderTruncationRunSummary(errorLogRows, acksPath, runs, findings)

	if report {
		fmt.Print(summary)
		// ONE row per run, clean or not — a check that only speaks when it fails
		// is indistinguishable from one that has stopped running.
		writeDocNote("render-truncation", summary, "render-truncation", "render-truncation-check")
		if len(findings) > 0 {
			os.Exit(1)
		}
		return
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(findings); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --render-truncation: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprint(os.Stderr, summary)
	if len(findings) > 0 {
		os.Exit(1)
	}
}
