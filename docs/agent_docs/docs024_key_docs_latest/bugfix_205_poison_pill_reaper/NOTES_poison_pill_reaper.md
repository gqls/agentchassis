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

## 2026-08-08 ~17:10 UTC — post-roll re-proof on v1.0.1266; migration 335 ledger gap closed

- Fresh chassis roll (`v1.0.1266`, both replicas started ~16:00Z; vet-intel +
  business-intel on the same image). Per the fleet practice (a roll is not
  evidence your fix still ships), re-proved both Go halves at the pod, each
  replica, same exec: WARN string `max_tokens not configured at any level` = 1,
  backfill guard literal `'in_progress', 'failed'` = 1, nonsense negative
  control = 0. **205's Go halves survive the roll — cite as live on v1.0.1266.**
- WARN watch-item: 0 firings on all four chassis-image pods — scoped claim:
  logs reach back only to pod start (~16:00Z), so this proves the last hour,
  not "never". Standing watch unchanged.
- Parked-state and reaper both hold post-roll: 606 cancelled / 2528 completed,
  zero pending/in_progress/failed; both scheduled tasks' stamps advance past
  the roll (17:02–17:03Z vs now() 17:04Z).
- **Migration dry-run (per session + after every roll) caught a real gap: 335
  was never RECORDED.** The 08-08 session applied it by hand but skipped the
  ledger; its `IF NOT EXISTS` / `CREATE OR REPLACE` shapes re-run clean, so the
  dry-run listed it as innocently pending — exactly the "applied by hand and
  never recorded" case the runner's NOTE warns about, and a future scoped
  `--apply` would have re-run it silently. Recorded via `--record-only` with
  the verification note (artifacts confirmed live this session, above).
- Cross-thread observation, not ours to fix: 305's probe refusal ("4-entry
  exclusion list, found 0") is NOT the 205 reaper rewrite — 305 targets the
  `claimed-item-timeout` sweep on site_work_items, and its live exclusion list
  is now an 8-entry SUPERSET already containing both types 305 adds
  (plus `dead_fragment_link`, `literal_markdown`). Applied-in-substance by
  other means; the 080 lane owns the record-or-supersede call.

## 2026-08-08 ~22:40 UTC — the WARN HAS fired; the log-based watch could never have told us

- Asked directly ("has the uncapped-step WARN fired?"), swept all **125**
  chassis-image pods in parallel: 0 WARN lines everywhere. But every pod had
  restarted ~22:00–22:25Z (another fleet roll — the second today), so that
  sweep's honest reach was under an hour. The logs were the wrong instrument.
- **The DB answered: the WARN's condition occurred 7 times on 08-07,
  15:14–21:23Z** — `med-price-collector/scrape_prices`, model
  `mistral-small3.1` (Ollama), rows with `max_tokens IS NULL`. Config census on
  the live row: no `max_tokens` at the step's `ai_service` path, the wrong
  path, or agent level — genuinely uncapped, and the 183 lane's clock-blind-spot
  work already knew this step as "no cap recorded". Fleet was on v1.0.1262
  (WARN live) from ~08:00Z that day [INFERRED from whole-fleet release
  practice; that pod no longer exists to grep].
- **Shape pin, both transports**: uncapped ANTHROPIC call → `max_tokens=2048`
  logged (transport fallback; 112 pre-fix verifier rows 08-05→08-07 all read
  2048). Uncapped OLLAMA call → `max_tokens NULL` (all 243 mistral rows since
  07-25 are NULL). So the **durable watch query** is:
  `SELECT agent_type, step_name, count(*), max(created_at) FROM llm_call_log
   WHERE created_at > '<since>' AND (max_tokens = 2048 OR max_tokens IS NULL)
   GROUP BY 1,2;` — cross-check a 2048 hit against config before calling it
  uncapped (2048 could someday be a chosen value).
- Impact: nil so far — all 7 calls `success=t`, local model (no paid spend),
  and the Anthropic 2048 cliff does not apply to the Ollama transport (its own
  defaults govern). The step still needs a deliberately chosen cap — med-price
  lane / owner call, not taken here.
- **Misstep recorded in WRONG_CALLS (2026-08-08)**: "WARN unfired" was asserted
  three times across two sessions (each time correctly scoped to pod age, each
  time conveying "has not fired") while the firing sat 7 hours after the first
  claim, in logs whose delivery window was hours. A watch-item defined as
  "grep pod stdout" on a fleet that restarts daily reads as a clean pass
  precisely when it has lost the evidence. Signals meant to outlive a pod
  belong in a table; the WARN worked, the watch definition did not.
- Since 08-08 00:00Z: zero uncapped calls on either shape. 6 uncapped steps
  remain unheard-from.

## 2026-08-09 — caps decided and applied (mig 347); yesterday's "WARN fired" claim RE-CORRECTED

- Owner asked for help deciding token caps. Measured rather than guessed, and
  the measurement immediately falsified yesterday's entry: **the uncapped-step
  census does not contain `scrape_prices`.** Its action calls Ollama directly
  with a hardcoded `num_predict: 500` (`vet_med_price_scrape_action.go:1197`)
  and never passes `ai_actions.go` — **the WARN has never fired**; the NULL
  `max_tokens`/`output_tokens` rows are that action's logging omission. Its
  outputs: p50 6 chars (usually `[]`), p95 222, max 607 ≈ 150–200 tokens —
  the code cap has ~3× headroom, needs no change. Correction recorded in
  WRONG_CALLS 2026-08-09 ("membership before mechanism"), the bug file and the
  HANDOFF banner.
- **Decision inputs** (all llm_call_log / agent_definitions, 08-09): all 7
  uncapped step-rows have ZERO calls since 07-26 (dormant); design-class steps
  measured p95 14,456–17,352, max 20,189 output tokens; fleet cap norms per
  model pulled from live definitions (sonnet-5 mode = 8000 across 40 steps);
  principle: a cap is an insurance ceiling, not a budget — ~2× biggest
  plausible output, rounded to a fleet value. Sonnet-5 steps never get small
  caps (thinking lands in output_tokens — the 138 cap-120 lesson).
- **Owner chose the recommended set** (AskUserQuestion): 32000
  site-architect/design · 16000 chief-strategist/generate_build_plan +
  content-creator/create_content (both rows) · 8000 brand-designer/
  analyze_brand, domain-analyst/analyze, provocation-gate-calibration/gate ·
  scrape_prices stays at its code 500.
- **Applied as `sql_for_agents/347`** (+ value-matched ROLLBACK), by hand +
  `--record-only` in one motion (the 335 lesson). Every UPDATE scoped
  `max_tokens IS NULL` at the path — the pre-existing chief-strategist 8192
  (older of TWO active rows, 2025-11-15) survives, guard-asserted. UPDATE
  counts exactly as measured (1,1,2,1,1,1); in-transaction guards + an
  external re-run both put the fleet uncapped census at **0**. Write-path
  mechanism (config.ai_service.max_tokens honoured) rests on the 08-08
  behavioural proof (verifier ran at 8000, output 3135); all six steps are
  dormant so no live-fire proof until one runs — the first run of each is the
  remaining watch.
- `snapshot_agent` captured ONE pre-image per type (its own selection rule);
  the second content-creator row's pre-state is reconstructible only via the
  value-matched ROLLBACK (original = key absent). Acceptable, noted.
- **Side-finding, not this lane's:** `chief-strategist` AND `content-creator`
  each have TWO active non-snapshot rows — which row the loader uses is
  loader-defined, and a cap set on only one would be roulette (347 therefore
  capped every uncapped row). RFC_006 SingleOwner territory; flagged here for
  whoever owns definition hygiene.
