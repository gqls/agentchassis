# PLAN 2026-08-06 — bug 205: the reaper resurrects deterministically-failing tasks forever, and the step that exposed it runs at a cap nobody chose

Owner thread: this session ("bugfix 205"). Bug file: `bugs_open/205_HANDOFF_2026-08-06_unconfigured_step_runs_at_the_hardcoded_2048_default_and_two_records_retry_forever.md`
(found and filed by the 183 lane, which deliberately did not fix it).

## What the diagnosis added to the bug file

The bug file's part 3 was explicitly `[INFERRED — NOT YET READ IN CODE]`: "the batch
selects records by not-yet-verified". **Read in code and config 2026-08-06 evening, the
real mechanism is different and worse:**

1. `LoadBusinessBatchAction` (`business_intel_actions.go:557-598`) claims
   `business_intel.collection_tasks` rows `status='pending'` → `in_progress`.
2. On verifier success, `StoreBusinessVerification` sets the task `completed`
   (`business_intel_actions.go:428-449`). **On failure there is no writer: the task
   stays `in_progress`. Nothing in Go ever writes `pending` back.** (Whole-repo grep:
   the only three writers of this table's `status` are the claim and two
   completed-variants.)
3. The **`stale-orchestration-reaper`** scheduled task (every 180 s, live row, no repo
   seed defines it) carries a `reset_tasks` CTE in its `pre_query`:
   `UPDATE business_intel.collection_tasks SET status='pending', started_at=NULL,
   orchestration_id=NULL WHERE status='in_progress' AND started_at < NOW() - INTERVAL
   '20 minutes'` — **unconditional: no retry ceiling, no backoff, no distinction
   between a dead claimant and a claimant that FAILED deterministically.**
   The table HAS a `retry_count` column; nothing anywhere increments it.
4. `vet-batch-verify` (every 300 s, gated on `pending > 0`) then re-claims and
   re-dispatches under a fresh correlation id. Loop period ≈ 20 min reap + ≤5 min
   schedule + batch time = the observed 25–55 min between identical failures.

**The class is 33 tasks, not 2 records.** [MEASURED 2026-08-06 ~20:30 UTC]
`orchestration_states`, last 24 h, `owner_agent_type='vet-practice-verifier'`:
**1,576 runs, 1,575 FAILED, 33 distinct task_ids** — each looping task re-dispatched
~50×/day. Only ONE of those (task `ea489aed-…`, business `926410bc-…`) reaches
`extract_and_reconcile` and burns an LLM call into the 2048 wall; the other 32 die
earlier at `scrape_website` on external API/URL errors, invisible to the
token-pressure check that found this bug. The bug file's second poisoned record
(prompt md5 `33749de2…`) no longer appears — only `105ca46f…` is looping now.

Prior art: 016b §9 already holds "a staleness reaper keyed on ROW AGE …" (2026-07-25,
`stale-work-item-reaper`) whose remedy names **attempt_count** — parking-with-a-counter
is established platform doctrine for reapers, not a new idea.

## The fix, ordered by what closes the door

**A. Reaper parks after N resets (DB config — live immediately, no roll).**
Rewrite the `reset_tasks` CTE: increment `retry_count` on every reset; once
`retry_count` reaches 5, set `status='failed'` with an `error_message` naming the
reaper and bug 205 instead of `pending`; back off re-eligibility via `scheduled_for =
NOW() + 20 min × retry_count` (the claim query already honours `scheduled_for`).
Report `tasks_parked` as its own counter column. This makes "re-dispatched forever"
unrepresentable for every failure mode — truncation, scrape errors, timeouts — not
just this bug's. A pod-crash victim still gets 5 spaced chances.

**B. The dormant re-arm door (Go — `ensure_collection_tasks.go`).**
The backfill's NOT EXISTS only blocks on `('pending','in_progress')`, so a parked
(`failed`) task would NOT block re-creation: if `vet-sweep-continue` (currently
disabled) is ever re-enabled, every parked business silently gets a fresh task with
`retry_count=0` and the loop reopens. Add `'failed'` to the blocking statuses. Zero
rows are `failed` today, so this is behaviourally inert until parking happens.

**C. WARN when a step's max_tokens resolves to the hardcoded transport default
(Go — `ai_actions.go`, the bug's candidate 1, framework-level).**
At the resolution fall-through (`ai_actions.go:358-363`), log at WARN with agent_type
+ step so an unsized step is visible the first time it runs, instead of after 64
truncations. Census (bug file): 8 of 126 active LLM steps have no cap at any level;
6 are dormant landmines.

**D. NOT done here — the step's own cap (owner call, per the bug file).**
~8000 would match the fleet mode; the bug file reserves cap raises to the owner and
the vet-intel lane. With A live the burn stops regardless; the parked task is the
prompt for that decision. Also out: `tolerate_truncation` (wrong for a reconciler —
silently-absent trailing fields would write a half-record into a practice directory).

## Sequencing

Go half (B+C) is inert until the next chassis roll; config half (A) is live on apply
and is the half that stops the burn. No ordering constraint between them (B matters
only when parking has happened AND the disabled sweep is re-enabled) — per the
2026-07-29 owner ruling, no ordering claim is made.

Immediately after applying A: seed `retry_count=5` on the 33 measured loopers (their
true attempt count is ~50; seeding lets the NEW mechanism park them on its next pass,
which both stops tonight's burn and behaviourally proves the parking branch live).

## Verification

- `llm_call_log`: no new `extract_and_reconcile` calls after the last parked cycle
  (require the quiet to coincide with `tasks_parked>0`, not with the sweep dying —
  the bug file's own caveat).
- `collection_tasks`: the 33 loopers sit at `status='failed'`, error naming 205.
- `orchestration_states`: vet-practice-verifier run rate collapses from ~65/h to 0.
- The WARN (C): after next roll, grep chassis logs for the new message on any
  unconfigured step that runs.

## Decisions and their reasons

- Park at **5** resets: a genuine crash victim rarely needs >2; a deterministic
  failure at 5 costs ~2 h of retries, vs. unbounded today. Not 3 — scrape targets
  can be transiently down for an hour.
- Park to `'failed'` (existing vocabulary of this table, zero current rows) rather
  than a new status value — no shared-vocabulary change, no consumer to notify;
  `cancelled` is taken by deliberate operator action (574 rows).
- Backoff via `scheduled_for`, not a new column — the claim predicate already reads it.
- Council: this plan (A+B+C) goes through the gate in one submission; Go committed
  with `Council-Submitted:` trailer, config applied with a row backup first
  (189 precedent).
