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
