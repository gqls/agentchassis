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
// did retire nineteen hand-copied INSERTs into it. It is no longer true. **FIVE**
// paths insert rows, as of 2026-08-22 (re-census with
// `grep -rn "INSERT INTO agent_error_log"` across every language; a bare count
// goes stale by ADDITION and reads as current for ever — owner ruling 2026-08-22):
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
// those five writers WHILE LOOKING COMPLETE, which is 358's own defect reproducing
// itself one level up. `SELECT DISTINCT error_code` is blind to none of them:
// it sees every writer regardless of language, seam, or whether the code is a
// literal, a constant, a positional argument or a value from config. It parses
// nothing, so no comment can become load-bearing.
//
// ITS ONE BLIND SPOT, STATED. A code that has never fired inside the 30-day
// retention window is invisible here. ~~That blind spot is harmless BY
// CONSTRUCTION — an unfired code produces no unread findings and costs nothing —
// so the authoritative half is blind to exactly the set of cases that do not
// matter.~~ A conservative Go source scan is kept as an EARLY WARNING at commit
// time (findingcodes_scan_test.go in the actions package), and it is explicitly
// not the guarantee: anything it misses is caught here within a day of first
// firing.
//
// > **CORRECTED 2026-08-24, and BOTH halves of that paragraph were wrong.**
// >
// > 1. **The scan test DID NOT EXIST.** This comment named
// >    `findingcodes_scan_test.go` from the day it shipped; the file was written
// >    on 2026-08-24, two days later. The claim passed a council round and was
// >    quoted in the concept register, and nobody opened the path it named. What
// >    existed instead was a hand-written list of eleven constants, which could
// >    only catch a code somebody remembered to add to it. It is real now, it
// >    DISCOVERS codes rather than listing them, and its own limits are written
// >    at the top of it.
// > 2. **"Harmless by construction" was too strong.** An unfired code costs
// >    nothing *today*, which is true and is not the same claim. Measured
// >    2026-08-24: 13 codes are written by the actions package and declared
// >    nowhere, all with zero rows in the window — so the population this half is
// >    blind to is not empty, it is thirteen, and each becomes a live undeclared
// >    finding the moment it first fires. That is precisely how
// >    LINK_CONTEXT_UNAVAILABLE arrived: written 2026-08-24, first row two hours
// >    later, red CronJob the same afternoon. The blind spot is BOUNDED and
// >    SHORT-LIVED — a day at most — which is the honest claim; it is not empty.
// >    The 13 are recorded in the registry's `_scan_baseline` and that list may
// >    only shrink.
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

	"github.com/gqls/agentchassis/pkg/buildinfo"
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
	Disposition string `json:"disposition"`
	Writer      string `json:"writer,omitempty"`
	Reader      string `json:"reader,omitempty"`

	// ReaderSink names the TABLE the reader actually selects the code from, and
	// it exists because `consumed` was silently ambiguous between two very
	// different states (batch 1, 2026-08-22).
	//
	// component_validation_rejected has a genuine automated reader — migration
	// 563 branches the component-creator's prompt on the failure class, so a
	// rejected component is regenerated with the right remedy. But that reader
	// reads `site_work_items.retry_feedback`, a dedicated single-writer column
	// (store_generated_component_action.go:1587), wired to the prompt by 564.
	// The agent_error_log row is a SECOND, parallel record that nothing reads.
	//
	// Without this field that entry would have been marked `consumed`, passed
	// the reader check (563 does contain the string), and read as healthy for
	// ever while its row stayed unread — bugs_open/358's own defect wearing a
	// green badge. Note this is a DIFFERENT failure from the one the registry
	// already warns about on STRUCTURAL_KEY_CARRY_MISS: that is the WRITER
	// being blind to part of its population; this is the READER reading
	// somewhere else entirely.
	ReaderSink  string   `json:"reader_sink,omitempty"`
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

	// ForeignSinks lists `consumed` codes whose reader consumes them from a
	// table other than agent_error_log. REPORT, never a finding: a parallel
	// record is not automatically a defect. It is here so the registry can
	// never again read as a closed loop over this table when it is not one.
	ForeignSinks []string `json:"consumed_from_another_sink"`

	// UnruledCap is the ratchet's state, emitted ALWAYS so "at the cap",
	// "below it" and "no cap at all" are distinguishable in one read.
	UnruledCap string `json:"unruled_cap"`

	// RetentionParity says whether the registry was compared against the live
	// sweep's short-retention list, and is emitted ALWAYS — "not checked" is a
	// state a reader must be able to see. Without it a stdin run (no DB, so no
	// parity) and a clean parity run produce the same empty findings list, and
	// the reader cannot tell which they are holding.
	RetentionParity string `json:"retention_parity"`

	// SourceArms says whether the two checks that open a `consumed` entry's
	// reader FILE actually ran, and is emitted ALWAYS for exactly the reason
	// RetentionParity is: a run that skipped them and a run that passed them
	// print the same empty findings list otherwise. See the sourceArms* consts.
	SourceArms string `json:"source_arms"`

	// DeclaredCodes is how many codes the registry this run graded against
	// declares. It is here because the registry TRAVELS IN THE IMAGE: a cluster
	// row and a local run disagreeing on this number is the visible tell that
	// the image is behind the repo, and without it that lag is invisible.
	DeclaredCodes int `json:"declared_codes"`
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
func loadFindingCodeRegistry(path string) (map[string]findingCodeEntry, int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	var entries map[string]json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, 0, fmt.Errorf("registry is not a JSON object: %w", err)
	}
	// The unruled CAP (owner ruling 2026-08-23). Absent is -1, which
	// auditUnruledCap reports as a finding: a cap you can delete is not a cap.
	unruledCap := -1
	if v, ok := entries["_unruled_cap"]; ok {
		if err := json.Unmarshal(v, &unruledCap); err != nil {
			return nil, 0, fmt.Errorf("_unruled_cap is not a number: %w", err)
		}
	}
	reg := make(map[string]findingCodeEntry, len(entries))
	for code, v := range entries {
		if strings.HasPrefix(code, "_") {
			continue // "_doc" — the file explains itself
		}
		var e findingCodeEntry
		if err := json.Unmarshal(v, &e); err != nil {
			return nil, 0, fmt.Errorf("registry entry %q: %w", code, err)
		}
		reg[code] = e
	}
	if len(reg) == 0 {
		return nil, 0, fmt.Errorf("registry %s declares no codes", path)
	}
	return reg, unruledCap, nil
}

// ─── THE UNRULED CAP, AS A RATCHET (owner ruling 2026-08-23) ────────────────
//
// The owner's instruction was "cap it". A plain cap is the obvious reading and
// it is the wrong mechanism here, for a reason this file already states about
// itself: `unruled` entries are listed at exit 0 because "a check that fails
// from day one over a pre-existing backlog is a check that gets ignored".
// Setting a target of, say, 10 against a live backlog of 32 makes the check red
// immediately and permanently, and a permanently red check is a disabled one.
//
// So the cap is a RATCHET, which gives what a cap gives and more:
//
//   - the count may never EXCEED the recorded cap  -> finding. A new code parked
//     as `unruled` is exactly what this stops: the backlog cannot grow, which is
//     the whole point of capping it.
//   - the count BELOW the cap  -> reported, with the number to lower it to.
//     Lowering is a one-word edit in the same commit as the ruling, and the
//     report is what stops the ground gained being quietly given back.
//   - the cap ABSENT -> finding. A cap that can be deleted is not a cap, and
//     deleting it would restore the unbounded backlog silently.
//
// The estate precedent is 102_coverage_ratchet.txt, which works the same way and
// for the same reason.
func auditUnruledCap(unruledCount, cap int) ([]findingCodeFinding, string) {
	switch {
	case cap < 0:
		return []findingCodeFinding{{
			Kind: "unruled-cap-missing", Code: "_unruled_cap",
			Detail: "the registry declares no `_unruled_cap`, so the undecided backlog is " +
				"unbounded again. It is a finding rather than a default because a cap that can be " +
				"removed without anything noticing is not a cap.",
		}}, "ABSENT — the backlog is unbounded"
	case unruledCount > cap:
		return []findingCodeFinding{{
			Kind: "unruled-over-cap", Code: "_unruled_cap",
			Detail: fmt.Sprintf("%d codes are unruled against a cap of %d. A code was declared and "+
				"parked rather than decided. Rule one (consumed / instrumented / human-evidence / "+
				"operational), or raise the cap deliberately and say why — but raising it is the "+
				"thing this ratchet exists to make visible.", unruledCount, cap),
		}}, fmt.Sprintf("OVER — %d unruled against a cap of %d", unruledCount, cap)
	case unruledCount < cap:
		return nil, fmt.Sprintf("%d unruled against a cap of %d — LOWER THE CAP TO %d in the same "+
			"commit as the ruling, or the ground gained is given back silently",
			unruledCount, cap, unruledCount)
	default:
		return nil, fmt.Sprintf("%d unruled, exactly at the cap — the backlog cannot grow", unruledCount)
	}
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

// ─── RUNNING WITHOUT A SOURCE TREE (bugs_open/358 phase 2) ──────────────────
//
// A NIL sourceReader means "this binary has no repo beside it", which is the
// CronJob's situation and nothing else's. It is passed explicitly, from the
// `--no-source` flag, and never inferred from a failed read: an unreadable file
// where source IS available stays a finding, because that is a registry pointing
// at a file that does not exist.
//
// WHY THE ARMS MOVE RATHER THAN THE SOURCE. Two of this mode's checks open the
// Go file a `consumed` entry names. No check image in this estate ships a source
// tree, so in the container all five `consumed` entries would raise
// `reader-unreadable` and the job would be RED EVERY DAY for a non-defect — and
// a permanently red check is a disabled one. Shipping ~17MB of Go source to
// satisfy them buys nothing: those arms compare the registry against source, and
// BOTH change only by commit, so they cannot come out differently in the cluster
// than they did at build time. Their home is commit time, and they now have a
// runner there (scripts/check-finding-code-registry.sh, wired into
// .githooks/pre-commit, which runs this package's tests including
// TestShippedRegistryIsSelfConsistent). Before that script they had none: the
// existing hook runs only `-run 'BudgetCron'`.
//
// WHAT IS STILL CHECKED WITHOUT SOURCE — everything a clock or the live table can
// move, which is the whole reason for scheduling this at all: the undeclared
// ratchet, the unruled cap, retention parity, `review_by` EXPIRY (the one arm
// only the passage of time can trip), the human-evidence window, bad
// dispositions, prefix collisions, and both `consumed` field-presence checks.
// The foreign-sink REPORT survives too — it compares two registry fields and
// never needed a file.
const (
	sourceArmsChecked = "checked — every `consumed` entry's reader file was opened and must name " +
		"both the code and its sink"
	sourceArmsSkipped = "NOT run (--no-source): this binary has no repo beside it. These two arms " +
		"grade the registry against Go SOURCE, and both halves change only by commit — so they " +
		"cannot come out differently here than they did at build time. They run at commit time " +
		"instead: scripts/check-finding-code-registry.sh (pre-commit) → `go test " +
		"./cmd/config-key-audit/`. Everything a clock or the live table can move DID run here, " +
		"including the undeclared ratchet and review_by expiry"
)

// auditFindingCodes is the pure check (same pure/impure split as
// findSingleOwnerViolations and censusOptionalKeys, for the same testability
// reason). `live` is the observed code list; `now` is passed rather than read so
// a review_by expiry test cannot depend on the wall clock. A nil `src` is the
// no-source-tree case above — stated in the report, never silent.
func auditFindingCodes(live []string, reg map[string]findingCodeEntry, src sourceReader, now time.Time) findingCodeReport {
	rep := findingCodeReport{Findings: []findingCodeFinding{}, Unruled: []string{},
		NotObserved: []string{}, ForeignSinks: []string{}}
	rep.SourceArms = sourceArmsChecked
	if src == nil {
		rep.SourceArms = sourceArmsSkipped
	}
	rep.DeclaredCodes = len(reg)

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
			// The two source-side arms. Skipped, and SAID so in rep.SourceArms,
			// when this binary ships without a repo — never skipped silently.
			var body string
			haveSource := src != nil
			if haveSource {
				b, err := src(e.Reader)
				if err != nil {
					rep.Findings = append(rep.Findings, findingCodeFinding{
						Kind: "reader-unreadable", Code: code,
						Detail: fmt.Sprintf("reader %s cannot be read: %v", e.Reader, err),
					})
					break
				}
				body = b
				// A plain substring check, and it is worth being exact about what
				// that accepts: the LITERAL must appear in the reader file. A
				// reader that reaches the code through a constant declared IN THE
				// SAME FILE passes, because the declaration line carries the
				// literal (358 §3.2 — page_build_failure_guard.go:65 is that
				// case). A reader that uses a constant declared in ANOTHER file
				// does NOT pass, and will draw reader-does-not-name-code; there is
				// no constant resolution here (review finding, Fable 2026-08-24 —
				// the previous comment implied there was). All five live readers
				// carry the literal, so this is a stated limit, not a defect.
				if !strings.Contains(body, code) {
					rep.Findings = append(rep.Findings, findingCodeFinding{
						Kind: "reader-does-not-name-code", Code: code,
						Detail: fmt.Sprintf("reader %s does not mention %s — a reader reference that "+
							"names the wrong file is worse than none, because it reads as a closed loop",
							e.Reader, code),
					})
				}
			}

			// WHICH SINK the reader reads. See the ReaderSink doc comment: a
			// reader that genuinely consumes the code from somewhere ELSE
			// leaves this table's row unread while the entry reads as closed.
			switch {
			case strings.TrimSpace(e.ReaderSink) == "":
				rep.Findings = append(rep.Findings, findingCodeFinding{
					Kind: "consumed-without-reader-sink", Code: code,
					Detail: "disposition 'consumed' requires reader_sink — the table the reader " +
						"actually selects this code from. 'Something reads it' and 'this table's " +
						"row is read' are different claims and the first was hiding the second",
				})
			case haveSource && !strings.Contains(body, e.ReaderSink):
				// Same strength as the reader check above, and the same limit:
				// it proves the reader MENTIONS the sink, not that it selects
				// this code from it. A file naming several tables can satisfy
				// this wrongly — stated rather than papered over. It is still
				// the check that would have caught the motivating case, where
				// the named reader never mentions agent_error_log at all.
				rep.Findings = append(rep.Findings, findingCodeFinding{
					Kind: "reader-sink-not-in-reader", Code: code,
					Detail: fmt.Sprintf("reader %s never mentions %s, so it cannot be selecting "+
						"this code from there — either the sink is wrong or the reader is",
						e.Reader, e.ReaderSink),
				})
			case e.ReaderSink != agentErrorLogSink:
				// REPORT, not a finding. A parallel record is not automatically
				// a defect — an append-only history beside an overwritten
				// column is often the point. What must not happen is it reading
				// as a closed loop over THIS table when it is not one.
				rep.ForeignSinks = append(rep.ForeignSinks,
					fmt.Sprintf("%s — reader %s consumes it from %s, so its %s row is still unread",
						code, e.Reader, e.ReaderSink, agentErrorLogSink))
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
	// The registry TRAVELS IN THE IMAGE, so its declared-code count and the
	// commit this binary was built from are what let a reader of an old row tell
	// which declarations it was graded against. Without them a stale image's
	// clean report is indistinguishable from a current one's.
	fmt.Fprintf(&b, "registry: %s — %d code(s) declared (bugs_open/358)\n",
		registryPath, rep.DeclaredCodes)
	fmt.Fprintf(&b, "built from: %s\n", buildinfo.GitCommit)
	fmt.Fprintf(&b, "retention parity: %s\n", rep.RetentionParity)
	fmt.Fprintf(&b, "unruled cap:      %s\n", rep.UnruledCap)
	fmt.Fprintf(&b, "source-side arms: %s\n", rep.SourceArms)
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
	if len(rep.ForeignSinks) > 0 {
		fmt.Fprintf(&b, "\nconsumed, but NOT from this table (%d): %s\n",
			len(rep.ForeignSinks), strings.Join(rep.ForeignSinks, "; "))
		b.WriteString("  Report only. The code has a real automated reader, but this table's row " +
			"is still unread - a parallel record, which is often deliberate. What it must not do " +
			"is read as a closed loop over agent_error_log.\n")
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
	sweepFile := ""
	// OPT-IN, with the unsafe side OFF (owner ruling 2026-08-02: new authority on
	// a shared seam ships as a field whose default is the safe one). Absent, the
	// source-side arms run exactly as they always have. Its ONE live consumer is
	// the finding-code-registry-check image's CMD, which has no repo beside it.
	noSource := false
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
		case "--no-source":
			noSource = true
		case "--sweep-file":
			// The hand-run wrapper has no DB handle of its own — it reads the
			// table through kubectl and pipes. Without this flag the parity check
			// would run ONLY in --report mode, i.e. only in the CronJob that does
			// not exist yet, which is the "built but never exercised" shape this
			// estate keeps paying for. The wrapper fetches database-cleanup's
			// pre_query the same way it fetches the codes and passes it here.
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "config-key-audit --finding-codes: --sweep-file needs a file path")
				os.Exit(2)
			}
			sweepFile = args[i+1]
			i++
		default:
			fmt.Fprintf(os.Stderr, "config-key-audit --finding-codes: unrecognised argument %q "+
				"(want: [--registry <file>] [--root <dir>] [--report] [--no-source] "+
				"[--sweep-file <file>])\n", args[i])
			os.Exit(2)
		}
	}

	reg, unruledCap, err := loadFindingCodeRegistry(registryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --finding-codes: registry %q: %v — refusing to "+
			"run without it, because every observed code would then read as undeclared and the "+
			"report would be noise rather than a finding.\n", registryPath, err)
		os.Exit(2)
	}

	var live []string
	var parityFindings []findingCodeFinding
	parityState := "not checked — no database connection (stdin mode). Retention parity needs " +
		"the live sweep; run with --report to check it."
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

		// Parity against the live retention sweep. Loaded here, while the
		// connection is open; the findings are merged after the audit below.
		arm, armFindings := loadRetentionArmFromDB(db)
		switch {
		case len(armFindings) > 0:
			parityFindings = armFindings
			parityState = "NOT checked — the sweep itself is missing or unreadable (see findings)"
		default:
			parityFindings = auditRetentionParity(arm, reg)
			parityState = fmt.Sprintf("checked against %s's agent_error_log arm — %d disagreement(s)",
				retentionSweepTask, len(parityFindings))
		}
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
		if sweepFile != "" {
			parityFindings, parityState = parityFromSweepFile(sweepFile, reg)
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

	// nil is the no-source-tree case, and it is passed EXPLICITLY from the flag
	// rather than inferred from a failed read: where source is available an
	// unreadable reader file stays a finding, because that is a registry
	// pointing at a file that does not exist.
	src := repoSourceReader(root)
	if noSource {
		src = nil
	}
	rep := auditFindingCodes(live, reg, src, time.Now())
	rep.Findings = append(rep.Findings, parityFindings...)
	rep.RetentionParity = parityState
	capFindings, capState := auditUnruledCap(len(rep.Unruled), unruledCap)
	rep.Findings = append(rep.Findings, capFindings...)
	rep.UnruledCap = capState

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

// ─── RETENTION PARITY (bugs_open/358, owner ruling 2026-08-22) ───────────────
//
// Migration 567 gave deliberate findings a 365-day clock and left ordinary
// failure plumbing at 30 days. It does that with a LIST of codes inside
// `database-cleanup`'s pre_query — the short-retention list — because nothing
// else in an agent_error_log row separates the two kinds. `severity` was the
// obvious discriminator and was measured and rejected: findings are written as
// error, warning AND info, plumbing as error, fatal AND warning, and three codes
// emit mixed severities.
//
// A hand-written list in a live pre_query is exactly the drift shape this lane
// retired twice over, so it does not get to sit unchecked. This is the parity
// half: the registry says what each code IS, the sweep says how long it LIVES,
// and the two must agree. Modelled on optional_budget_cron_parity_test.go, which
// exists because two actions entered the registry counted as ZERO and were
// invisible to their check for four days.
//
// THE ASYMMETRY IS DELIBERATE AND MATCHES THE MIGRATION'S OWN SAFETY PROPERTY.
// 567's default is KEEP: a code absent from the list is retained for 365 days.
// So the two directions carry different weight:
//
//   - `operational` MISSING from the list  -> finding. The table grows without
//     anyone deciding it, silently, and nothing else would ever say so.
//   - a FINDING code PRESENT in the list   -> finding, and the worse of the two:
//     it means deliberate evidence is being deleted at 30 days again, which is
//     the whole defect 358 exists about.
//   - `instrumented` on either side        -> NOT asserted, reported. Whether a
//     time-boxed measurement wants history or only frequency is a per-code
//     judgement its `review_by` governs; RFC_029's two resolver codes want
//     frequency and are deliberately short-retention. Asserting either way here
//     would make this check the owner of a decision that is not its to make.
//
// agentErrorLogSink is the table this whole registry is about. A `consumed`
// entry naming any other sink is reported rather than assumed closed.
const agentErrorLogSink = "agent_error_log"

const retentionSweepTask = "database-cleanup"

const retentionSweepQuery = `SELECT pre_query FROM scheduled_tasks WHERE name = $1`

// errorLogArmBounds delimits the ONE arm of the sweep that deletes from
// agent_error_log. Scoping to it matters: a bare Contains over the whole
// pre_query would also match a code name that some future arm mentions for an
// unrelated reason, and would then report parity that is not there.
func extractErrorLogArm(preQuery string) (string, bool) {
	start := strings.Index(preQuery, "DELETE FROM agent_error_log")
	if start < 0 {
		return "", false
	}
	rest := preQuery[start:]
	end := strings.Index(rest, "RETURNING")
	if end < 0 {
		return "", false
	}
	return rest[:end], true
}

// namedInShortRetention asks whether the arm names this code as a quoted SQL
// literal. Exact by construction: measured 2026-08-22, none of the sixteen
// short-retention names occurs anywhere else in the sweep.
func namedInShortRetention(arm, code string) bool {
	return strings.Contains(arm, "'"+code+"'")
}

func auditRetentionParity(arm string, reg map[string]findingCodeEntry) []findingCodeFinding {
	var out []findingCodeFinding
	codes := make([]string, 0, len(reg))
	for c := range reg {
		codes = append(codes, c)
	}
	sort.Strings(codes)

	for _, code := range codes {
		named := namedInShortRetention(arm, code)
		switch reg[code].Disposition {
		case "operational":
			if !named {
				out = append(out, findingCodeFinding{
					Kind: "retention_parity_missing",
					Code: code,
					Detail: "declared `operational` but NOT named in database-cleanup's short-retention " +
						"list, so it now lives 365 days instead of 30. Plumbing is the high-volume half " +
						"of this table — add it to migration 567's list, or change its disposition.",
				})
			}
		case "consumed", "human-evidence", "unruled":
			if named {
				out = append(out, findingCodeFinding{
					Kind: "retention_parity_finding_expires_early",
					Code: code,
					Detail: "declared `" + reg[code].Disposition + "` — a deliberate finding — yet named in " +
						"database-cleanup's short-retention list, so its rows are deleted at 30 days " +
						"unread. That is bugs_open/358 itself. Remove it from the list, or rule it " +
						"`operational` if that is what it really is.",
				})
			}
		case "instrumented":
			// deliberately unasserted — see the block comment above.
		}
	}
	return out
}

// loadRetentionArmFromDB returns the agent_error_log arm of the live sweep.
// A MISSING OR RENAMED TASK IS A FINDING, NOT A PASS: if this returned "" and
// the caller skipped the parity check, a deleted database-cleanup row would read
// as perfect agreement — the exact shape ("absence reads as health") that this
// whole mode exists to refuse.
func loadRetentionArmFromDB(db *sql.DB) (string, []findingCodeFinding) {
	var pre sql.NullString
	err := db.QueryRow(retentionSweepQuery, retentionSweepTask).Scan(&pre)
	if err == sql.ErrNoRows {
		return "", []findingCodeFinding{{
			Kind: "retention_sweep_absent",
			Code: retentionSweepTask,
			Detail: "no scheduled_tasks row named `" + retentionSweepTask + "` — the sweep that " +
				"enforces retention is GONE or renamed. Reported as a finding rather than skipped: " +
				"a skipped parity check and a perfect one are the same clean report.",
		}}
	}
	if err != nil {
		return "", []findingCodeFinding{{
			Kind:   "retention_sweep_unreadable",
			Code:   retentionSweepTask,
			Detail: "could not read its pre_query: " + err.Error() + " — parity was NOT checked.",
		}}
	}
	arm, ok := extractErrorLogArm(pre.String)
	if !ok {
		return "", []findingCodeFinding{{
			Kind: "retention_sweep_no_error_log_arm",
			Code: retentionSweepTask,
			Detail: "its pre_query no longer contains a `DELETE FROM agent_error_log … RETURNING` " +
				"arm. Either retention moved somewhere this check cannot see, or the arm was lost — " +
				"and a lost arm means NOTHING is ever deleted, which looks like health from here.",
		}}
	}
	return arm, nil
}

// parityFromSweepFile runs the retention parity check over a pre_query the
// caller fetched itself. Same three refusals as the DB path, for the same
// reason: an unreadable or empty sweep must be a FINDING, never a skip — a
// skipped parity check and a clean one print the same thing, and this mode's
// whole subject is checks that cannot fail.
func parityFromSweepFile(path string, reg map[string]findingCodeEntry) ([]findingCodeFinding, string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return []findingCodeFinding{{
			Kind:   "retention_sweep_unreadable",
			Code:   retentionSweepTask,
			Detail: "--sweep-file " + path + ": " + err.Error() + " — parity was NOT checked.",
		}}, "NOT checked — the sweep file could not be read (see findings)"
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return []findingCodeFinding{{
			Kind: "retention_sweep_absent",
			Code: retentionSweepTask,
			Detail: "--sweep-file " + path + " is empty. The fetch of `" + retentionSweepTask +
				"`'s pre_query returned nothing, which means the task is gone, renamed, or the " +
				"query failed — not that retention agrees with the registry.",
		}}, "NOT checked — the sweep file is empty (see findings)"
	}
	arm, ok := extractErrorLogArm(string(raw))
	if !ok {
		return []findingCodeFinding{{
			Kind: "retention_sweep_no_error_log_arm",
			Code: retentionSweepTask,
			Detail: "--sweep-file " + path + " contains no `DELETE FROM agent_error_log … RETURNING` " +
				"arm. A lost arm means nothing is ever deleted, which looks like health from here.",
		}}, "NOT checked — no agent_error_log arm in the sweep (see findings)"
	}
	f := auditRetentionParity(arm, reg)
	return f, fmt.Sprintf("checked against %s's agent_error_log arm (via --sweep-file) — %d disagreement(s)",
		retentionSweepTask, len(f))
}
