// FILE: cmd/config-key-audit/findingcodes.go
//
// bugs_open/358. Which agent_error_log finding code is firing with no declared
// disposition?
//
// WHAT THE DEFECT ACTUALLY IS. 358 measured that most deliberately-written
// finding codes have no automated reader and are deleted unread at 30 days
// (migration 466's `database-cleanup` pre_query, live hourly). But "readerless"
// is a symptom, and for some codes a legitimate state — operational plumbing is
// correctly consumed by the generic newest-N diagnostic readers, and RFC_029's
// two resolver codes are deliberate time-boxed instrumentation with an owner.
// The defect one rung up is what this mode checks: A FINDING CODE CAN ENTER THE
// ESTATE WITH NO DECLARED DISPOSITION, AND NOTHING NOTICES. Every property 358
// measured follows from that — the count grows (one more code arrived the same
// morning the bug file was written), the reader is never added afterwards, and
// retention erases the backlog before anyone counts it.
//
// WHY THE AUTHORITY IS THE LIVE TABLE AND NOT A SOURCE SCAN. This is the whole
// design, and it is a correction to 358's own fix candidate B2.
// platform/orchestration/agenterrors/agenterrors.go:3 declares itself "The ONE
// writer against agent_error_log", and RFC_012 (owner ruling 2026-08-06) really
// did retire nineteen hand-copied INSERTs into it. It is no longer true. Five
// paths insert rows:
//
//	agenterrors.go:89                                  — the seam
//	store_generated_component_action.go:1439           — own INSERT, kept DELIBERATELY
//	                                                     (a council round's edit-quality and
//	                                                     guardian seats objected to
//	                                                     consolidating it; reason at :1428)
//	internal/agents/contentcreator/claims_guard.go:184 — CANNOT use the seam: the agent holds a
//	                                                     *pgxpool.Pool (agent.go:92) against
//	                                                     Write's *sql.DB. A TYPE-LEVEL barrier
//	sql_for_agents/214_build_dispatch_watchdog.sql:108 — SQL, inside a scheduled pre_query
//	cmd/content-loss-check/main.go:292                 — a standalone binary
//
// So a check placed at the seam — the obvious home — would be blind to four of
// the five writers WHILE LOOKING COMPLETE, which is 358's own defect reproducing
// itself one level up. `SELECT DISTINCT error_code` is blind to none of them:
// it sees every writer regardless of language, seam, or whether the code is a
// literal, a constant, a positional argument or a value from config. It parses
// nothing, so no comment can become load-bearing.
//
// ITS ONE BLIND SPOT, STATED. A code that has never fired inside the 30-day
// retention window is invisible here. That blind spot is harmless BY
// CONSTRUCTION — an unfired code produces no unread findings and costs nothing —
// so the authoritative half is blind to exactly the set of cases that do not
// matter. A conservative Go source scan is kept as an EARLY WARNING at commit
// time (findingcodes_scan_test.go in the actions package), and it is explicitly
// not the guarantee: anything it misses is caught here within a day of first
// firing.
//
// THE TWO DIRECTIONS ARE NOT SYMMETRIC.
//   - observed but not registered -> FINDING, exit 1. This is the ratchet.
//   - registered but not observed -> REPORT ONLY. Retention is 30 days, so an
//     absence proves 30 days and never "never" (358 §8). Same discipline as the
//     StaleAck precedent at optionalbudget.go:73-77: bookkeeping to tidy, not a
//     defect to page on. It is still worth reading — BUILD_DISPATCH_STALLED sits
//     here because migration 214 was never applied, so a loop 358 lists as
//     closed does not exist live, and its zero rows read as "quiet".
//   - `unruled` entries are COUNTED AND LISTED at exit 0. A check that fails
//     from day one over a pre-existing backlog is a check that gets ignored; the
//     crisp signal is "someone shipped a code nobody declared", and the backlog
//     is a number that should go down.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const findingCodeRegistryPath = "docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json"

// liveCodesQuery is the authority. DISTINCT over the whole retained window:
// this mode asks which codes EXIST, never how many rows each has, so no
// occurred_at bound belongs here — the retention job already provides the only
// window there is.
const liveCodesQuery = `
	SELECT DISTINCT error_code FROM agent_error_log
	WHERE error_code IS NOT NULL AND error_code <> ''`

// findingCodeEntry is one declaration. The required field varies by
// disposition, and each is chosen so it cannot be satisfied by typing — the
// lesson optional_explicit_wire_acks.json paid for (RFC_029 §10.15: "an ack
// satisfiable by typing the key is no ack"). `reader` is verified against the
// file it names; `review_by` expires on its own; `why` must name the window it
// accepts.
type findingCodeEntry struct {
	Disposition string   `json:"disposition"`
	Writer      string   `json:"writer,omitempty"`
	Reader      string   `json:"reader,omitempty"`
	Owner       string   `json:"owner,omitempty"`
	ReviewBy    string   `json:"review_by,omitempty"`
	Why         string   `json:"why,omitempty"`
	RawVariants []string `json:"raw_variants,omitempty"`
	Note        string   `json:"note,omitempty"`
	OwnerLane   string   `json:"owner_lane,omitempty"`
}

// findingCodeFinding is one thing wrong. Kind discriminates so a consumer of
// the JSON can tell the ratchet's catch (undeclared) from a bad declaration,
// which are different jobs for different people.
type findingCodeFinding struct {
	Kind   string `json:"kind"`
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

// findingCodeReport is the whole answer, findings and census together, for the
// reason optionalKeyCensusRow gives: a consumer that sees only the findings
// cannot tell a shrinking backlog from a broken check.
type findingCodeReport struct {
	Findings      []findingCodeFinding `json:"findings"`
	Unruled       []string             `json:"unruled"`
	NotObserved   []string             `json:"registered_not_observed"`
	ObservedCount int                  `json:"observed_codes"`
}

var validDispositions = map[string]bool{
	"consumed": true, "instrumented": true,
	"human-evidence": true, "operational": true, "unruled": true,
}

// normaliseFindingCode is the normalisation decision 358 §8 says must be made
// explicitly or a family will double-count as compliance: the key is the code
// up to the first colon. One writer (create_tool_cross_link_items.go) emits
// `tool_crosslink_not_emitted:<reason>`, and a live `LIKE
// 'tool_crosslink_not_emitted%'` query already reads those variants as one
// population — so collapsing them here matches how the estate already queries
// them, rather than inventing a second grouping.
func normaliseFindingCode(code string) string {
	return strings.SplitN(strings.TrimSpace(code), ":", 2)[0]
}

// loadFindingCodeRegistry reads the declaration file. A missing or unreadable
// registry is an ERROR, never an empty map: with no declarations every observed
// code reads as undeclared, and the report becomes noise rather than a finding —
// the same reasoning loadOptionalExplicitAcks gives for refusing to run.
func loadFindingCodeRegistry(path string) (map[string]findingCodeEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries map[string]json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, fmt.Errorf("registry is not a JSON object: %w", err)
	}
	reg := make(map[string]findingCodeEntry, len(entries))
	for code, v := range entries {
		if strings.HasPrefix(code, "_") {
			continue // "_doc" — the file explains itself
		}
		var e findingCodeEntry
		if err := json.Unmarshal(v, &e); err != nil {
			return nil, fmt.Errorf("registry entry %q: %w", code, err)
		}
		reg[code] = e
	}
	if len(reg) == 0 {
		return nil, fmt.Errorf("registry %s declares no codes", path)
	}
	return reg, nil
}

// sourceReader is how the check verifies a `consumed` entry's reader. Injected
// so the pure half is testable without a tree — and so the verification is a
// real read of the named file rather than a trusted string.
type sourceReader func(fileLine string) (string, error)

// repoSourceReader reads the file part of a "path/to/file.go:123" reference.
// The line number is not used to slice: a reader can be a query several lines
// below the reference, and pinning the exact line would make the registry rot
// on every unrelated edit above it. What must be true is that the named FILE
// mentions the code — enough to catch a reference pointed at the wrong file,
// which is the failure this guard exists for, without inventing a maintenance
// burden that would make people stop updating the registry.
func repoSourceReader(root string) sourceReader {
	return func(fileLine string) (string, error) {
		path := fileLine
		if i := strings.LastIndex(path, ":"); i > 0 {
			if _, err := fmt.Sscanf(path[i+1:], "%d", new(int)); err == nil {
				path = path[:i]
			}
		}
		b, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}

// auditFindingCodes is the pure check (same pure/impure split as
// findSingleOwnerViolations and censusOptionalKeys, for the same testability
// reason). `live` is the observed code list; `now` is passed rather than read so
// a review_by expiry test cannot depend on the wall clock.
func auditFindingCodes(live []string, reg map[string]findingCodeEntry, src sourceReader, now time.Time) findingCodeReport {
	rep := findingCodeReport{Findings: []findingCodeFinding{}, Unruled: []string{}, NotObserved: []string{}}

	observed := map[string]bool{}
	for _, raw := range live {
		code := normaliseFindingCode(raw)
		if code == "" {
			continue
		}
		observed[code] = true
	}
	rep.ObservedCount = len(observed)

	// DIRECTION 1 — the ratchet. An observed code with no declaration.
	for code := range observed {
		if _, ok := reg[code]; !ok {
			rep.Findings = append(rep.Findings, findingCodeFinding{
				Kind: "undeclared",
				Code: code,
				Detail: "written to agent_error_log but absent from the registry — declare it " +
					"(consumed / instrumented / human-evidence / operational), or unruled if the " +
					"decision is genuinely open",
			})
		}
	}

	// DIRECTION 2 — the declarations themselves must hold up.
	for code, e := range reg {
		if !validDispositions[e.Disposition] {
			rep.Findings = append(rep.Findings, findingCodeFinding{
				Kind: "bad-disposition", Code: code,
				Detail: fmt.Sprintf("disposition %q is not one of consumed / instrumented / "+
					"human-evidence / operational / unruled", e.Disposition),
			})
			continue
		}

		switch e.Disposition {
		case "consumed":
			// The reader claim is VERIFIED, not trusted. This is what stops the
			// field being satisfiable by typing.
			if strings.TrimSpace(e.Reader) == "" {
				rep.Findings = append(rep.Findings, findingCodeFinding{
					Kind: "consumed-without-reader", Code: code,
					Detail: "disposition 'consumed' requires a reader file:line",
				})
				break
			}
			body, err := src(e.Reader)
			if err != nil {
				rep.Findings = append(rep.Findings, findingCodeFinding{
					Kind: "reader-unreadable", Code: code,
					Detail: fmt.Sprintf("reader %s cannot be read: %v", e.Reader, err),
				})
				break
			}
			// The code may be reached through a Go constant rather than a
			// literal (358 §3.2 — the trap that made the original census miss
			// page_build_failure_guard's reader), so accept either the code
			// itself or a constant declared to it in the same file.
			if !strings.Contains(body, code) {
				rep.Findings = append(rep.Findings, findingCodeFinding{
					Kind: "reader-does-not-name-code", Code: code,
					Detail: fmt.Sprintf("reader %s does not mention %s — a reader reference that "+
						"names the wrong file is worse than none, because it reads as a closed loop",
						e.Reader, code),
				})
			}
		case "instrumented":
			if strings.TrimSpace(e.Owner) == "" {
				rep.Findings = append(rep.Findings, findingCodeFinding{
					Kind: "instrumented-without-owner", Code: code,
					Detail: "disposition 'instrumented' requires the owning doc — a measurement " +
						"nobody owns is an unread finding wearing a better word",
				})
			}
			due, err := time.Parse("2006-01-02", strings.TrimSpace(e.ReviewBy))
			if err != nil {
				rep.Findings = append(rep.Findings, findingCodeFinding{
					Kind: "instrumented-without-review-date", Code: code,
					Detail: "disposition 'instrumented' requires review_by as YYYY-MM-DD — the date " +
						"is what makes this a TIME-BOXED state rather than a permanent exemption",
				})
			} else if now.After(due) {
				rep.Findings = append(rep.Findings, findingCodeFinding{
					Kind: "instrumentation-expired", Code: code,
					Detail: fmt.Sprintf("review_by %s has passed — re-rule it (consume / demote / "+
						"keep) or extend the date with a reason in the owning doc", e.ReviewBy),
				})
			}
		case "human-evidence":
			// The reason must name the retention window it is accepting.
			// Anything less is "we decided not to decide" with a nicer label.
			if !strings.Contains(e.Why, "30") {
				rep.Findings = append(rep.Findings, findingCodeFinding{
					Kind: "human-evidence-without-window", Code: code,
					Detail: "disposition 'human-evidence' requires `why` to name the retention " +
						"window it accepts (30 days unresolved, 14 resolved — migration 466); a " +
						"reason that does not mention it has not accepted anything",
				})
			}
		case "unruled":
			rep.Unruled = append(rep.Unruled, code)
		}

		if !observed[code] {
			rep.NotObserved = append(rep.NotObserved, code)
		}
	}

	// DIRECTION 3 — registry-wide properties, checked once for every code
	// instead of once per new code against a frozen snapshot of the others.
	// This subsumes the two hand-maintained `taken` lists in
	// discovery_checks_error_log_test.go and
	// save_sections_content_data_links_test.go, both of which had gone stale.
	// Prefix-disjointness is a REAL property, not a style rule: the estate has
	// live LIKE queries on `tool_crosslink_not_emitted%` and
	// `component_validation_%`, so a code sharing a prefix with another is
	// silently swept into someone else's population.
	codes := make([]string, 0, len(reg))
	for c := range reg {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	for i, a := range codes {
		for _, b := range codes[i+1:] {
			if strings.HasPrefix(b, a) {
				rep.Findings = append(rep.Findings, findingCodeFinding{
					Kind: "prefix-collision", Code: a,
					Detail: fmt.Sprintf("%q is a prefix of %q — a LIKE query on either catches both", a, b),
				})
			}
		}
	}

	sort.Slice(rep.Findings, func(i, j int) bool {
		if rep.Findings[i].Kind != rep.Findings[j].Kind {
			return rep.Findings[i].Kind < rep.Findings[j].Kind
		}
		return rep.Findings[i].Code < rep.Findings[j].Code
	})
	sort.Strings(rep.Unruled)
	sort.Strings(rep.NotObserved)
	return rep
}

// loadLiveCodesFromDB is the impure half. Separate from the pure check for the
// reason every mode here splits them, and separate from loadLiveAgentsFromDB
// because this mode asks the error table a question, not the fleet.
func loadLiveCodesFromDB(db *sql.DB) ([]string, error) {
	rows, err := db.Query(liveCodesQuery)
	if err != nil {
		return nil, fmt.Errorf("reading agent_error_log: %w", err)
	}
	defer rows.Close()
	var codes []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

func findingCodeRunSummary(rep findingCodeReport, registryPath string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "finding-code registry check — %d distinct error_code(s) observed live, "+
		"%d finding(s), %d unruled, %d registered-but-unobserved\n",
		rep.ObservedCount, len(rep.Findings), len(rep.Unruled), len(rep.NotObserved))
	fmt.Fprintf(&b, "registry: %s (bugs_open/358)\n", registryPath)
	if len(rep.Findings) == 0 {
		b.WriteString("\nEvery code firing today is declared. The unruled count below is the " +
			"backlog, not a defect — it should go down.\n")
	}
	for _, f := range rep.Findings {
		fmt.Fprintf(&b, "\n  [%s] %s\n      %s\n", f.Kind, f.Code, f.Detail)
	}
	if len(rep.Unruled) > 0 {
		fmt.Fprintf(&b, "\nunruled (%d): %s\n", len(rep.Unruled), strings.Join(rep.Unruled, ", "))
	}
	if len(rep.NotObserved) > 0 {
		fmt.Fprintf(&b, "\nregistered but not observed in the retained window (%d): %s\n",
			len(rep.NotObserved), strings.Join(rep.NotObserved, ", "))
		b.WriteString("  Report only. Retention is 30 days, so this proves 30 days of silence " +
			"and never 'never' — a quiet code and a dead mechanism look identical here.\n")
	}
	return b.String()
}

func emitFindingCodes(args []string) {
	registryPath := findingCodeRegistryPath
	root := "."
	report := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--registry":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "config-key-audit --finding-codes: --registry needs a file path")
				os.Exit(2)
			}
			registryPath = args[i+1]
			i++
		case "--root":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "config-key-audit --finding-codes: --root needs a directory")
				os.Exit(2)
			}
			root = args[i+1]
			i++
		case "--report":
			report = true
		default:
			fmt.Fprintf(os.Stderr, "config-key-audit --finding-codes: unrecognised argument %q "+
				"(want: [--registry <file>] [--root <dir>] [--report])\n", args[i])
			os.Exit(2)
		}
	}

	reg, err := loadFindingCodeRegistry(registryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --finding-codes: registry %q: %v — refusing to "+
			"run without it, because every observed code would then read as undeclared and the "+
			"report would be noise rather than a finding.\n", registryPath, err)
		os.Exit(2)
	}

	var live []string
	if report {
		// Straight from Postgres: this image contains no kubectl and the service
		// account has no pods/exec RBAC (see fleetdb.go).
		db, dbErr := dbConn()
		if dbErr != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --finding-codes: %v\n", dbErr)
			os.Exit(2)
		}
		if db == nil {
			fmt.Fprintln(os.Stderr,
				"config-key-audit --finding-codes --report: PG_CLIENTS_HOST is not set, so there "+
					"is no table to read. In the CronJob this comes from the pod env; by hand, "+
					"either export it or drop --report and pipe the code list on stdin.")
			os.Exit(2)
		}
		defer db.Close()
		live, err = loadLiveCodesFromDB(db)
	} else {
		var raw []byte
		raw, err = io.ReadAll(os.Stdin)
		if err == nil {
			for _, line := range strings.Split(string(raw), "\n") {
				if s := strings.TrimSpace(line); s != "" {
					live = append(live, s)
				}
			}
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --finding-codes: %v\n", err)
		os.Exit(2)
	}

	// VACUITY GUARD. An empty code list produces a clean report indistinguishable
	// from a healthy estate, and this check's own subject is checks that cannot
	// fail — so an empty read must refuse, never pass. The live table has carried
	// tens of thousands of rows continuously since 2026-07-23; zero distinct
	// codes means the query did not run, not that nothing is wrong.
	if len(live) == 0 {
		fmt.Fprintln(os.Stderr,
			"config-key-audit --finding-codes: 0 distinct error_code values read — refusing to "+
				"print a clean report over an empty or failed read. agent_error_log is never empty "+
				"in practice, so this is an instrument failure, not a healthy estate.")
		os.Exit(2)
	}

	rep := auditFindingCodes(live, reg, repoSourceReader(root), time.Now())

	if report {
		summary := findingCodeRunSummary(rep, registryPath)
		fmt.Print(summary)
		// ONE row per run, clean or not — the convention every scheduled check
		// here follows. A check that only speaks when it fails is
		// indistinguishable from one that has stopped running, so a MISSING row
		// must mean "the job did not run", never "nothing is wrong".
		writeDocNote("finding-code-registry", summary,
			"finding-code-registry", "finding-code-registry-check")
		if len(rep.Findings) > 0 {
			os.Exit(1)
		}
		return
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --finding-codes: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprint(os.Stderr, findingCodeRunSummary(rep, registryPath))
	if len(rep.Findings) > 0 {
		os.Exit(1)
	}
}
