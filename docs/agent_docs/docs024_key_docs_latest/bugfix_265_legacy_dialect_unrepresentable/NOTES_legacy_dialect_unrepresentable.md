# NOTES — bug 265, legacy `input_schema` dialect made unrepresentable (append-only, newest at the bottom)

## 2026-08-16 — session 1

**Bug selection.** Session was named for 213 and 274; both closed at HEAD (`git ls-tree HEAD --
bugs_closed/` shows both; 213 by owner ruling 08-15, 274 live+proven). Census of `bugs_open/`:
last-commit date per file, edits by live transcripts (grep `"file_path":…bugs_open/NNN` in
`~/.claude/projects/…/*.jsonl` modified <36 h), header status keywords, `who-owns.py`. 265 was
the one open, unowned (no workstream dir; one live *Read* of the helper by the 260 lane, no
edits), platform-level candidate. 207/217 also surfaced as fixed-live-proven but still in
`bugs_open/` — closure housekeeping for their lane, not taken here.

**Re-verification** — see the bug file's 2026-08-16 section for the table. Two things moved:
the count is 3 not 4 (LMC converted loans-consolidation on 08-15), and the producer is hand
SQL, not the component-creator. The second is the decisive one for the design.

**Misstep avoided (recorded because it was close):** my first instinct was to write the
birth-path Go check as *the* fix — that is what the bug file's candidate 3 pointed at. One
`GROUP BY created_from, source_agent_type` showed the component-creator has never emitted the
dialect and every legacy row was `manual`. Had I not run it, the "fix" would have guarded a door
none of the four rows came through, and the next hand seed would have walked straight in. The
check that caught it: **read the provenance columns before choosing the seam.**

**Design decisions and reasons** — in the PLAN. Short form: CHECK constraint (the seam every
producer crosses) + behaviour-preserving conversion of the 3 rows in the same migration + Go
birth check for legibility + comment/tripwire rewritten to cite the constraint. Not: a CronJob
(constraint doesn't drift), deleting the reader's projection (fail-open again), rewriting
applied seeds.

**Probe run of 437** — COMMIT swapped for `SELECT …; ROLLBACK;` via sed into the scratchpad;
`psql -v ON_ERROR_STOP=1`. Output: guards passed, `UPDATE 3`, `NOTICE: 437: 3 legacy rows
converted…`, ALTER, COMMENT, then rollback. Live table read back: legacy 3, constraint absent,
backup absent. Converted rows captured → Go fixture. First attempt of the probe printed 11
lines (psql command tags + 3 rows); filtered `^{` for the fixture.

**Tests** — `TestLegacyDialectConversionMatchesProjection` (3 rows) and
`TestIsLegacyInputSchemaDialect` (6 shapes) green. Mutations: fixture minus report-dossier's
`required` → FAIL "projection mismatch"; predicate `return false` → FAIL on `legacy` and
`fields and properties`. Restored; `git diff` on the helper checked to be only the intended
hunks (the mutation restore was a first-occurrence replace and could have hit the wrong
function — it did not, verified by reading the func body).

**Same-file passenger.** `store_generated_component_action.go` was already ` M` at session
start: one hunk at ~:890 (`WHERE status = 'active'` → `PageWantedLivePredicateFor("")`), the
185 lane's, uncommitted. The helper exists at HEAD (`git grep … HEAD` hit
`datahelpers/links.go:361`), so it compiles. Named in the commit message per the pathspec
passenger practice rather than reverting another session's WIP in the tree.

**Council** — submitted `aba82416-de79-4452-8730-3e35ca0a15bb`, 6 edits, 11 grounded_in.
Verdict pending at time of writing; commit carries `Council-Submitted:`.

**`go vet ./platform/orchestration/actions/`** reports one pre-existing "unreachable code" in
`load_component_library_actions.go:207` — not mine, not touched.

**Two content questions surfaced for other lanes, not fixed:** report-dossier `body` marked
`source: llm` (its seed says never LLM-authored) — gripper-dossier lane's call; `site-header`
carries v2 field defs with no `fields` wrapper — a third shape the reader cannot see, harmless
today. Both in the bug file.
