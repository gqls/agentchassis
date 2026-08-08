# NOTES — bug 205 (append-only, newest at the bottom)

## 2026-08-06 evening — session start, ownership, validity

- `who-owns.py 205` said OWNED by the 183 lane — but the bug file itself records the
  183 lane deliberately not fixing it, and the live transcript with 36 mechanism hits
  (`1c06862f`) is the 183 lane raising `classify_and_extract`'s cap, not touching 205.
  Taken on here at the owner's direction (session named "bugfix 205").
- Validity re-confirmed: failures every hour through 11:00 UTC, all
  `stop_reason=max_tokens` at 2048, zero successes.

## 2026-08-06 evening — the [INFERRED] part 3 read in code; the mechanism is the reaper

- All 08-06 truncation failures share ONE task (`ea489aed-…`), each under a FRESH
  `vet-batch-processor` parent (~50 min apart) — so re-dispatch, not in-orchestration
  retry. Query: join `llm_call_log` → `orchestration_states` on orchestration_id.
- Whole-repo grep: only three Go writers of `collection_tasks.status` — claim
  (`pending`→`in_progress`), and two `completed` variants. Nothing writes `pending`.
- The resetter is CONFIG, not code: `stale-orchestration-reaper.pre_query`'s
  `reset_tasks` CTE resets `in_progress` >20 min → `pending` unconditionally.
  Found via `SELECT … FROM scheduled_tasks WHERE pre_query ILIKE '%collection_tasks%'`.
- **The bug file's inferred mechanism ("batch selects by not-yet-verified") was
  WRONG in detail** — the batch honestly selects `pending` only; it is the reaper
  that manufactures the `pending`. Same loop shape, different door to close.
  Corrected in the bug file rather than silently.
- Class size [MEASURED 20:30 UTC]: 1,576 verifier runs / 24 h, 1,575 FAILED,
  33 distinct tasks. 32 of 33 fail at `scrape_website` (external API/URL errors)
  BEFORE any LLM call — invisible to the token-pressure check; the truncation task
  was just the one that cost money.
- The second poisoned prompt from the bug file (`33749de2…`) no longer appears in
  the log window — only `105ca46f…` loops now. Not chased: whichever exit it took
  (manual cancel among the 574, or a successful run), the mechanism is unchanged.
- `retry_count` column exists on the table, default 0, and NOTHING increments it —
  confirmed by grep and by the poisoned task's live row still reading 0 after ~50
  dispatches.
- Prior art: 016b §9 "staleness reaper keyed on ROW AGE" (2026-07-25) already
  prescribes attempt-counting for reapers. The concept register's
  `scheduler-and-tasks.md` claim that the reaper "fails 24h-stale orchestrations" is
  STALE (live thresholds: 30/90 min, 4 h, 20 min tasks) — to be corrected visibly.
- Timezone false alarm (cost ~10 min): `collection_tasks.started_at` is naive and
  looked 65 min stale while un-reaped; DB `timezone` is UTC and local wall-clock is
  BST, so the gap was 7 min, not 65. Check `SELECT now(), current_setting('timezone')`
  before reading naive timestamps against wall-clock.

## 2026-08-07 ~01:30 UTC — config applied, loopers seeded

- Council submission fired: `SUBMISSION_CORR=2db88f8f-11ea-47ed-b37d-35a6096d5597`
  (budget ~30 min; find the run by payload, not the printed id).
- Go + docs committed `d1eb3a6b5` with `Council-Submitted:` trailer. Commit-scope
  report showed one same-file passenger in LANDMINES.md (another session's
  window-bounded-census correction) — named in the commit message, taken as found.
- Live row backed up to `BACKUP_2026-08-07_reaper_pre_query.sql` (60 lines,
  reset_tasks present), then `SQL_2026-08-07_reaper_parking.sql` applied clean
  (both DO/RAISE verifies passed). Live row re-read: `tasks_parked` counter and
  `bugs_open/205` marker both present, `updated_at 01:26:49Z`.
- Seed: exactly **33 rows** updated to `retry_count=4` — row identity returned and
  it matches the measured looper set, poisoned task `ea489aed…` included. All 33
  were `in_progress` at seed time, so each parks when its current claim passes
  20 min stale. The parked message will say "5 stale-claim resets", counting from
  the seed — their true history (~50 dispatches each) is recorded here and in the
  bug file.
- Council round 1 went `complete_invalid` in 7 min: edit 3's `file` field described
  the DB row in prose and the fix_plan validator requires a repo-relative path
  ("file path must be repo-relative with no traversal or whitespace"). Re-pointed
  edit 3 at the repo-recorded apply script and resubmitted with
  `RESUBMIT_CORR=2db88f8f…` — same correlation, run `a45cee3f…`. Lesson for the
  next config-change submission: name the apply SCRIPT in `file`, put the live-row
  target in the rationale.

## 2026-08-07 ~01:45 UTC — parking PROVEN; council round 2 (REVISE) answered and resubmitted

- **Parking behaviourally proven**: all 33 seeded loopers parked in ONE reaper pass
  (watcher: parked=0 at 01:37:46, parked=33 at 01:40:48). Table now
  failed=33 / completed=2527 / cancelled=574, **pending=0, in_progress=0**.
- **Quiet-because-parked, not dead**: both scheduled tasks' stamps advance past the
  apply; 0 llm calls on the step since; the only verifier runs since 01:41 all
  belong to parent `c3c8aaea…` (created 01:15:39 — the last pre-parking batch
  draining its in-memory item list; parked tasks cannot be RE-claimed but an
  already-running batch keeps walking its list until done).
- **Council REVISE (round counting ours: refused-schema, then REVISE), gating seat
  `debug_historian`.** All objections answered with queries, not prose:
  - task_type census: the whole table is `initial_verification`/`veterinary`
    (3,134 rows) — guardian's blast-radius concern has an empty population today;
    the structural point (a future task_type inherits park-at-5 silently) is
    recorded as an owner note in the bug file.
  - pre_query census: exactly 2 rows mention the table; only the reaper writes it.
  - `vet-sweep-continue` enabled=f BY QUERY; `ensure_collection_tasks` has exactly
    one consumer by definitions census.
  - `ROLLBACK_2026-08-07_reaper_pre_query.sql` generated mechanically FROM the
    backup (asserts the parking branch is gone after rollback).
- **The best objection was a mirror**: two seats cited "a landmine contradicting
  the premise that nothing writes retry_count" — that landmine is THIS lane's own
  entry, written after the fix, synced to doc_notes at ~01:20Z, read by the council
  at 01:36Z as pre-existing prior art. **Lesson: a landmine you sync mid-task
  becomes prior art against your own submission — date-stamp premise claims
  ("before this fix, as of <time>") in both the landmine and the submission, or
  sync after the round.** Recorded in the resubmission verbatim.
- Round 2 resubmitted same corr, run `8310b61e…`.

## 2026-08-07 ~01:55 UTC — council APPROVED (round 2)

- Verdict read from `diagnosis_artifacts` (latest council_report on the
  correlation): **approved**, run `8310b61e…`, completed 01:52:56Z. The censuses
  did the arguing: empty blast-radius population, sole-writer proof, enabled=f by
  query, backup+rollback+RETURNING discipline shown.
- Earlier commits (`d1eb3a6b5` Go, `8ab75a9fb` docs) carry `Council-Submitted:` and
  are credited automatically by the 098 report now the verdict is approved; this
  closing commit carries `Council-Reviewed:` on a verdict actually read.
- Remaining to close the bug (per the "fixed AND live" bar — and note the owner's
  08-06 ruling: a finished bug STAYS in `bugs_open/`):
  1. Next chassis roll carries the Go halves — pod-prove with
     `strings /app/agent-chassis | grep -c "max_tokens not configured at any level"`
     (positive) and a negative control, every replica.
  2. Owner call on candidate 2 (the step's own cap, ~8000 fits the fleet mode) —
     the 33 parked tasks are the standing prompt; un-park runbook in RUNBOOK.
  3. The WARN's first live firing will name any of the 8 uncapped steps that runs.

## 2026-08-07 ~08:20 UTC — Go halves POD-PROVEN on v1.0.1262; loop still dead 6h on

- Fresh roll: both replicas on `v1.0.1262`. Pod-grep, each replica: WARN string
  `max_tokens not configured at any level` = 1; the backfill's SQL literal
  `'in_progress', 'failed'` = 1; nonsense negative control = 0. **Both Go halves
  live** — cite as "live on v1.0.1262 as at 2026-08-07".
- Six hours of behaviour: parked=33 holding, pending=0, in_progress=0,
  **0 verifier runs since 02:00Z** (the single llm call in the 01:41–02:00 window
  was the pre-parking batch draining, already recorded above). WARN unfired in 6h
  of logs — correct: the only ACTIVE uncapped step is the parked verifier; it will
  fire the first time any of the 8 uncapped steps runs.
- Bug 205 is now **fixed AND live in full** (config proven 01:40Z, Go proven
  08:17Z on v1.0.1262). Stays in `bugs_open/` per the owner's 08-06 ruling.
  What remains is OWNER DECISIONS (see README_where_we_are).

## 2026-08-08 ~15:00 UTC — all four owner decisions executed and proven

- **Decision 1 (cap 8000):** 067-sweep first — `extract_and_reconcile` is the
  agent's ONLY LLM step. Backup of the ai_service block taken
  (`BACKUP_2026-08-08_verifier_ai_service_block.txt`), then `jsonb_set` at the
  nested path, guarded (row count 1, `jsonb_typeof='number'` asserted — the
  resolver type-asserts float64). Un-parked `ea489aed…`; batch claimed it 14:48;
  **LLM call succeeded: max_tokens=8000, output_tokens=3135** — the document
  needed ~3,100 tokens, so the 2048 fallback could NEVER have passed it. Task
  `completed`, business `926410bc…` now **`verified`**, first attempt at the
  right cap after ~64 failures.
- **Decision 2 (cancel the 32):** cancelled in the same guarded transaction as
  the un-park, RETURNING all 32 ids, `error_message` appended with the owner
  decision and date. Precedent: the 574 earlier cancels.
- **Decisions 3+4 (per-type ceiling + shared mechanism):** migration
  `sql_for_agents/335` applied — `reaper_policies` table (a task type declares
  its ceiling by INSERT; undeclared → 5/20m/20m defaults),
  `business_intel.reap_stale_collection_tasks()` carries the accounting once,
  and the reaper's `reset_tasks` CTE is now a one-line call to it. **Induced
  test (transaction, rolled back, zero residue): undeclared type parked on the
  5th reset at defaults; declared `park_after=2` policy honoured with backoff
  stamped.** RFC_018 defines the contract + the deliberate stopping point (no
  dynamic-SQL generalisation until a second queue adopts — a mechanism with no
  live caller rots). Register SCH-024; LANDMINES 205 entry corrected in place
  (the check now includes `\sf` the function + reading reaper_policies).
- Post-migration the reaper runs the function-backed pre_query on schedule.
  Rollback ladder: 335_ROLLBACK (back to the 08-07 inline parking CTE) →
  ROLLBACK_2026-08-07 (emergency, back to unconditional resets).

## 2026-08-08 ~15:30 UTC — cold-start re-verification; docs closed out; lane ends

- Fresh session picked up the 08-07 HANDOFF, found the 08-08 execution already
  committed (`b6e70cd70`), and re-verified live rather than re-proving:
  - `collection_tasks`: cancelled=606 (574+32), completed=2528 (2527+1),
    **zero** pending/in_progress/failed — decisions 1+2 hold.
  - `public.reaper_policies` exists with the declared
    `initial_verification` policy (5/20m);
    `business_intel.reap_stale_collection_tasks()` exists; the live
    `stale-orchestration-reaper.pre_query` contains the function call (grep=1);
    reaper stamp advancing post-migration (15:13:53Z).
- **Misstep, paid this session:** probed the table as
  `business_intel.reaper_policies` (reasoning: the function and the tasks table
  live there) → "relation does not exist", which reads as *migration never
  applied*. Migration 335 creates the table UNQUALIFIED, so it landed in
  `public`. The LANDMINES 205 check already spells the query unqualified
  (`SELECT * FROM reaper_policies;`) and is correct as written — the trap only
  fires if you "improve" it with a schema prefix. Recorded in the HANDOFF
  banner; LANDMINES left unedited (entry is right, and the file carries another
  session's 30 uncommitted lines — not taking a same-file passenger for a nil
  correction).
- Closed the three doc gaps the 08-08 commit left: bug 205 header status line
  (still said "OPEN — FIX IN FLIGHT / live and burning"), README_where_we_are
  closing entry (owner's log ended at "four decisions for you"), HANDOFF
  superseded-banner. No new SUMMARY: against the five headings it would repeat
  SUMMARY_2026-08-08 with only the tense changed ("executing" → "executed"),
  and the rarity rule says that is not a milestone.
