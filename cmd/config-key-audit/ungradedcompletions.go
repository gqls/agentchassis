// FILE: cmd/config-key-audit/ungradedcompletions.go
//
// The commissioned reader for NO_CHANGE_GATE_UNREADABLE_RESULT (bugs_open/393,
// owner decision 4, 2026-08-25).
//
// WHAT THE ROW MEANS. The no-change completion gate (gate 1b,
// complete_work_item_no_change.go) grades a completing work item by reading the
// handler's result against the counters its roster entry declares. When it
// cannot read that result AND the type's policy is abstain, the item completes
// UNGRADED and the gate writes one durable row saying so — the right failure
// direction, loudly. The refuse-policy arm blocks instead and never reaches
// this record, so every row of this code means exactly one thing: a rostered
// abstain-policy item_type completed ungraded because its handler's result
// shape has drifted from what the roster declares.
//
// WHY A READER. Nothing consumed those rows. The first instance
// (dark_section_audit, 11 rows, 2026-08-14→17) sat unread for 11 days and was
// found by a census accident; the rows then start dying on the retention clock.
// A handler whose shape drifts is PERMANENTLY exempt from grading, silently —
// so the ratchet here turns the NEXT drifting type into a finding the morning
// after it first appears: rows are grouped by item_type, and any type not on
// the acknowledged list fails the run.
//
// THE ACKS FILE is this check's own definition of a permitted exception, in the
// commit_sha_exposure_acks.json shape: keyed on item_type, an entry needs a
// non-empty `reason` (a hollow ack is warned about and ignored), and the file
// ships INSIDE the image from committed HEAD so an unreviewed acknowledgement
// is unrepresentable. dark_section_audit is acknowledged at birth: its 11 rows
// are fully diagnosed (7 were bugs_closed/287 spawn records, fixed and rolled
// 08-17; the rest were color-variable-fixer shapes retired by the 2026-08-19
// rerouting) and the type currently files under RFC_056 filing_mode=record —
// nothing can complete, so nothing can complete ungraded.
//
// VACUITY. Zero rows of this code is the HEALTHY state, not a blind one — the
// code is written only when an abstain-policy rostered type completes ungraded.
// What is never healthy is an EMPTY agent_error_log: hundreds of rows a day
// live inside the retention window, so a zero total means the query went blind,
// and this check refuses to report over it rather than printing a clean run.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// ungradedCompletionsCode is the finding code this reader consumes, declared as
// a same-file const because the finding-code registry's checker verifies that
// the READER FILE names the code it claims to read (findingcodes.go:452, and
// the LINK_CONTEXT_UNAVAILABLE precedent). The sink it reads from is
// agent_error_log, named in the query below for the same reason.
const ungradedCompletionsCode = "NO_CHANGE_GATE_UNREADABLE_RESULT"

// ungradedGroupsQuery groups the code's rows by the item_type the writer stamps
// into context (recordUnknownNoChangeShape, complete_work_item_no_change.go —
// context->>'item_type' is the clean key; the error_message carries the same
// value only prose-embedded). COALESCE keeps a context-less row VISIBLE as its
// own group rather than silently dropped: a row this reader cannot attribute is
// a finding about the writer, not a row to skip.
const ungradedGroupsQuery = `
SELECT COALESCE(jsonb_agg(t), '[]'::jsonb) FROM (
  SELECT COALESCE(NULLIF(context->>'item_type',''), '(no item_type in context)') AS item_type,
         count(*)                AS rows,
         min(occurred_at)::text  AS first_seen,
         max(occurred_at)::text  AS last_seen
    FROM agent_error_log
   WHERE error_code = 'NO_CHANGE_GATE_UNREADABLE_RESULT'
   GROUP BY 1
   ORDER BY 1) t;`

// ungradedAlivenessQuery is the instrument-alive control: agent_error_log is
// never empty inside its retention window in practice (findingcodes.go's own
// vacuity premise), so zero here means the read went blind, never that the
// fleet stopped erring.
const ungradedAlivenessQuery = `SELECT count(*) FROM agent_error_log;`

// ungradedGroup is one item_type's slice of the code's rows.
type ungradedGroup struct {
	ItemType  string `json:"item_type"`
	Rows      int    `json:"rows"`
	FirstSeen string `json:"first_seen"`
	LastSeen  string `json:"last_seen"`
	// Acknowledged is true when the acks file carries this item_type with a
	// non-empty reason: a human diagnosed this population and said why it is
	// not (or no longer) actionable.
	Acknowledged bool `json:"acknowledged"`
}

type ungradedCompletionsAck struct {
	Reason   string `json:"reason"`
	Date     string `json:"date"`
	Evidence string `json:"evidence"`
}

// loadUngradedCompletionsAcks mirrors loadCommitShaExposureAcks: `_`-prefixed
// documentation keys skipped, a hollow (empty-reason) ack warned about and
// ignored, a missing or malformed file an error the caller turns into exit 2 —
// never an empty ack set, because "could not read the exceptions" and "there
// are no exceptions" have opposite meanings.
func loadUngradedCompletionsAcks(path string) (map[string]bool, error) {
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
		var ack ungradedCompletionsAck
		if err := json.Unmarshal(blob, &ack); err != nil {
			return nil, fmt.Errorf("acks entry %q: %w", key, err)
		}
		if strings.TrimSpace(ack.Reason) == "" {
			fmt.Fprintf(os.Stderr,
				"config-key-audit --ungraded-completions: acks entry %q has an empty "+
					"`reason` — ignoring it; an acknowledgement must say what was diagnosed "+
					"and why the population is not actionable.\n", key)
			continue
		}
		acked[key] = true
	}
	return acked, nil
}

// markAcknowledged is the pure decision: stamp each group with its ack state.
// Split from the fetch so both directions are testable on one fixture
// (optionalexplicitwires_test.go's canonical shape).
func markAcknowledged(groups []ungradedGroup, acked map[string]bool) []ungradedGroup {
	out := make([]ungradedGroup, len(groups))
	for i, g := range groups {
		g.Acknowledged = acked[g.ItemType]
		out[i] = g
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ItemType < out[j].ItemType })
	return out
}

func unackedUngradedCount(groups []ungradedGroup) int {
	n := 0
	for _, g := range groups {
		if !g.Acknowledged {
			n++
		}
	}
	return n
}

// ungradedCompletionsRunSummary is the doc_notes body and the --report stdout.
// It states SCOPE, not just result — "0 drifting types over 41,000 error rows"
// and "0 over 0" have opposite meanings (the sibling checks' pinned rule).
func ungradedCompletionsRunSummary(errorLogRows int, acksPath string, groups []ungradedGroup) string {
	acked, fresh := 0, 0
	rows := 0
	for _, g := range groups {
		rows += g.Rows
		if g.Acknowledged {
			acked++
		} else {
			fresh++
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "ungraded-completions: %d %s row(s) across %d item_type(s) "+
		"(%d acknowledged, %d NEW) in an agent_error_log holding %d row(s); acks=%s\n",
		rows, ungradedCompletionsCode, len(groups), acked, fresh, errorLogRows, acksPath)
	for _, g := range groups {
		state := "NEW"
		if g.Acknowledged {
			state = "ack"
		}
		fmt.Fprintf(&b, "  [%s] %s: %d row(s), %s → %s\n",
			state, g.ItemType, g.Rows, g.FirstSeen, g.LastSeen)
	}
	if fresh > 0 {
		fmt.Fprintf(&b, "  A NEW type here means a rostered abstain-policy item_type is "+
			"completing UNGRADED: its handler's result shape has drifted from the counters "+
			"noChangeGates declares (complete_work_item_no_change.go). Diagnose the shape "+
			"(each row's error_message carries what arrived), fix the roster or the handler, "+
			"and only then — if the population is genuinely settled — acknowledge it with the "+
			"diagnosis in %s.\n", acksPath)
	}
	return b.String()
}

// ungradedCompletionsStdin is the hand-run input shape, built by
// scripts/audit-ungraded-completions.sh: the grouped rows plus the aliveness
// total in one object. A bare array is refused with a pointer at the wrapper,
// because without the total this mode cannot tell a clean table from a blind
// read.
type ungradedCompletionsStdin struct {
	Groups       []ungradedGroup `json:"groups"`
	ErrorLogRows *int            `json:"error_log_rows"`
}

// emitUngradedCompletions: [--acks <file>] [--report]. Exit codes as the
// sibling scheduled checks: 0 = every drifting type is acknowledged (or none
// exists); 1 = a NEW drifting type; 2 = the check could not run, which must
// never read as a pass.
func emitUngradedCompletions(args []string) {
	acksPath := "docs/agent_docs/docs024_key_docs_latest/architecture_review/no_change_unreadable_acks.json"
	report := false
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--acks":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr,
					"config-key-audit --ungraded-completions: --acks needs a file path")
				os.Exit(2)
			}
			acksPath = args[i+1]
			i++
		case args[i] == "--report":
			report = true
		default:
			fmt.Fprintf(os.Stderr,
				"config-key-audit --ungraded-completions: unrecognised argument %q "+
					"(want: [--acks <file>] [--report])\n", args[i])
			os.Exit(2)
		}
	}

	acked, err := loadUngradedCompletionsAcks(acksPath)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --ungraded-completions: acks file %q: %v — refusing to run "+
				"without it.\n", acksPath, err)
		os.Exit(2)
	}

	var (
		groups       []ungradedGroup
		errorLogRows int
	)
	if report {
		db, err := dbConn()
		if err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --ungraded-completions: %v\n", err)
			os.Exit(2)
		}
		if db == nil {
			fmt.Fprintln(os.Stderr,
				"config-key-audit --ungraded-completions --report: PG_CLIENTS_HOST is not set, "+
					"so there is nothing to read. In the CronJob this comes from the pod env; "+
					"by hand, use scripts/audit-ungraded-completions.sh instead.")
			os.Exit(2)
		}
		defer db.Close()
		var raw []byte
		if err := db.QueryRow(ungradedGroupsQuery).Scan(&raw); err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --ungraded-completions: groups query: %v\n", err)
			os.Exit(2)
		}
		if err := json.Unmarshal(raw, &groups); err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --ungraded-completions: groups decode: %v\n", err)
			os.Exit(2)
		}
		if err := db.QueryRow(ungradedAlivenessQuery).Scan(&errorLogRows); err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --ungraded-completions: aliveness query: %v\n", err)
			os.Exit(2)
		}
	} else {
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --ungraded-completions: reading stdin: %v\n", err)
			os.Exit(2)
		}
		var in ungradedCompletionsStdin
		if err := json.Unmarshal(raw, &in); err != nil || in.ErrorLogRows == nil {
			fmt.Fprintln(os.Stderr,
				"config-key-audit --ungraded-completions: stdin must be one object "+
					`{"groups": [...], "error_log_rows": N} — the groups alone cannot prove the `+
					"read was not blind. Use scripts/audit-ungraded-completions.sh, which builds it.")
			os.Exit(2)
		}
		groups, errorLogRows = in.Groups, *in.ErrorLogRows
	}

	// The instrument-alive control. Zero rows of the CODE is healthy; an empty
	// TABLE is a blind read, and a clean report over a blind read is the exact
	// shape this check exists to end one level down.
	if errorLogRows == 0 {
		fmt.Fprintln(os.Stderr,
			"config-key-audit --ungraded-completions: agent_error_log reads as EMPTY — "+
				"hundreds of rows a day live inside its retention window, so this read went "+
				"blind; refusing to print a clean report over it.")
		os.Exit(2)
	}

	graded := markAcknowledged(groups, acked)

	summary := ungradedCompletionsRunSummary(errorLogRows, acksPath, graded)
	if report {
		fmt.Print(summary)
		// ONE row per run, clean or not — a check that only speaks when it fails
		// is indistinguishable from one that has stopped running.
		writeDocNote("ungraded-completions", summary,
			"ungraded-completions", "ungraded-completions-check")
		if unackedUngradedCount(graded) > 0 {
			os.Exit(1)
		}
		return
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(graded); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --ungraded-completions: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprint(os.Stderr, summary)
	if unackedUngradedCount(graded) > 0 {
		os.Exit(1)
	}
}
