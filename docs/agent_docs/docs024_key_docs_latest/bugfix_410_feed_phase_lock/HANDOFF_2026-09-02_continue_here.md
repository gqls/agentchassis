# HANDOFF — `bugfix_410_feed_phase_lock` — the BUG IS CLOSED; the LANE has one optional check and two owner decisions

**Written 2026-09-02 ~16:30Z by the `bugfix_410_feed_phase_lock` lane. All times in this lane are
DB time (`SELECT now()`). This is the cold-start doc — read this first, and read it INSTEAD of
`HANDOFF_2026-08-26_continue_here.md`, which is superseded and kept only as the record of what was
believed on the 26th.**

⚠ **410 NAMES TWO UNRELATED BUGS.** This lane owns the **news-feed cadence phase lock**, now at
`bugs_closed/410_HANDOFF_2026-08-26_next_fetch_at_stamped_at_fetch_time_phase_locks_every_six_hour_news_site_to_a_twelve_hour_cadence.md`.
The OTHER 410 — `..._three_seams_fail_toward_the_quiet_default...` — is a different case in
`bugs_open/` with its own lane (`bugfix_410_silent_scan_loss`). **Resolve by slug, never by number.**

---

## 1. Bottom line

**The bug is FIXED, LIVE, and CONFIRMED WORKING on real traffic — twice, on two independent pass
pairs and two disjoint site sets.** It was closed 2026-09-02 and the file moved to `bugs_closed/`.

**Nothing is required of anyone to keep it working.** What remains is:

- **one optional confirmation** (~20:59Z tonight, 10 minutes) that the *new* chassis build
  `v1.0.1354` exercises it live, not just carries it — §4;
- **two decisions that are the owner's**, both about capacity and spend, neither of which a
  session should take — §3;
- **one unbuilt guard** that is genuinely unowned and is not this lane's to hold open — §5.

## 2. What the bug was and what shipped (one paragraph each)

**The bug.** News sites are meant to refresh every 6 hours. A refresh pass runs on a fixed
timetable, drifting only seconds. But when a pass actually fetched a source it stamped
`next_fetch_at = NOW() + fetch_interval` — counted from **the fetch**, which happens 10 s to 9 min
after the timetable fired, because sites are processed one after another. So a 6-hour source came
due seconds *after* the next 6-hourly pass had already looked at it and moved on. It waited
another 6 hours. The lock re-armed every cycle and **every run reported COMPLETED.** 12 of 14 news
sites, since the trigger was armed.

**The fix (candidate 1).** A due look-ahead of **half the trigger cadence**, applied in *every*
layer that asks "is this source due?" — serve on the **nearest** grid tick, not the first tick
strictly after. Cadence is read **live** from `scheduled_tasks`; `COALESCE` falls back to
`interval '3 hours'` (half of today's 21600 s) so a renamed task degrades to the designed value,
never to the bare `NOW()` that caused the lock.

| layer | where | state |
|---|---|---|
| shared predicate | `platform/orchestration/actions/feed_due_lookahead.go` (`feedDueLookaheadSQL`, `feedSourceDuePredicate`) | LIVE |
| source-level (live path) | `dispatch_feed_sources_action.go` due query | LIVE |
| source-level (dormant) | `feed_actions.go` `LoadDueSourcesAction` — zero live callers, fixed so a future caller inherits it | LIVE |
| site-level admission | migration `653_content_feed_due_lookahead_HOLD.sql` → `content-feed-trigger.find_news_sites` | APPLIED 2026-08-26 20:52Z |

Council **APPROVED round 1**, corr `04c657d2-cbee-4528-b124-b53a747d2e96`. Commits: `201236b2a`
(fix), `4da33f40e`, `8ac0d0558`, `f3ca6a9e5`, `86aea3804`, `b34c24f4c`, `ac1727beb`, **`703948e78`**
(closure).

## 3. ⚠ THE DECISIONS — both the owner's, neither taken

### Decision A — raise the per-pass cap from 10, or accept a ~9-hour cadence

**What the cap IS.** The trigger's site-selection query ends `ORDER BY due_at ASC NULLS LAST,
domain ASC LIMIT 10`. That `LIMIT 10` is how many sites one refresh pass will take. It is a plain
number written into the configuration.

**The rule that governs it.** The cap was set as a capacity/spend control (`bugs_open/316`, fix
556): when more sites are due than the cap admits, the longest-waiting go first and the rest wait
for the next pass. Changing it is a spend decision, explicitly reserved to the owner.

**How this case measures against it.** `[MEASURED 2026-09-02]` **14** eligible news sites; **12**
are 6-hour-only; **2** (dartsonline, relojistas) hold a sub-6h source, are due at *every* pass and
take 2 of the 10 slots every time. That leaves **8 slots for 12 sites**. Across the three healthy
passes in retained history: 36 demanded, 24 served — **exactly 24 observed**, spread evenly
(max−min = 1). So:

| | before the fix | today | designed |
|---|---|---|---|
| effective refresh interval, 6h-only site | 12 h | **≈9 h** | 6 h |

**The decision:** raise `LIMIT` to **≥14** and every site gets the intended 6 hours; leave it at 10
and four sites are a pass late, by rotation, for ever. **The cost of raising it is roughly doubling
feed ingestion** (~+100 ingester runs/day) — the spend flagged on 08-26. Nothing is broken either
way; 4 sites currently sit 9 h 24 m–9 h 32 m stale.

⚠ **If you raise it, the number goes stale by ADDITION.** The cap is a literal while the eligible
site count grows with every news site added: the right answer was "~12" on 2026-08-26 and is **14**
today. Consider deriving it, or leaving explicit headroom, rather than pinning today's count.

### Decision B — is ~9 hours acceptable as the steady state, or does this need a bug of its own?

If Decision A is "leave it at 10", then the estate's stated design (6-hourly news) and its actual
behaviour (~9-hourly) differ permanently, and **only `bugs_open/316` records that.** The choice is
whether that is an accepted trade-off written into 316, or a defect someone is expected to close.
**A session should not decide this** — it is the difference between a documented capacity limit
and an open obligation.

**Where it is recorded:** CONTRIB filed into `bugs_open/316` on 2026-09-02 with all the numbers.
316 is literally titled *"the news feed cap starves the alphabetical tail and the queue is 2×
oversubscribed"*, so it is the right home either way.

## 4. The ONE optional check left — 10 minutes, ~20:59Z tonight

**A fresh chassis build `v1.0.1354` rolled at 15:39/15:53Z.** Both replicas were re-probed and
**both carry the look-ahead** (`make_interval(secs => interval_seconds / 2.0)` → **2**; negative
control `interval_seconds / 7.0` → **0**; positive control → **1**).

**But it rolled AFTER the 14:58:58 pass, so every live confirmation to date was produced by
`v1.0.1352`.** The new binary carries the capability; it has not yet been *exercised*.

> **The check:** after the ~20:59Z pass, re-run the acceptance test (RUNBOOK, "THE ACCEPTANCE
> TEST") on the pair **14:58:58 → ~20:59**. Expect ~3 discriminating sites served. Read the actual
> fire times first — do not reuse the numbers above.

**This is corroboration, not a blocker.** The capability probe already answers the question that
matters, and the fix's Go half is deterministic given the predicate is in the binary. If it comes
back clean, the lane has nothing left but §3 and §5.

⚠ **Provenance note:** the `build provenance` startup line was **already out of `--tail=3000`** on
pods up only ~25 minutes. An empty grep there means "not in range", **not** "unstamped". Do not
substitute a *discovery* grep for a 40-hex string — it matches Go's internal digit table and
returns the same wrong answer on every service.

## 5. Residuals at close — each with its actual status

1. **Go↔config predicate parity is NOT enforced against live config on a schedule.** A future
   migration could rewrite `find_news_sites` without the look-ahead and **nothing would fire**;
   the only signal would be sites drifting back to 12 hours. Candidate: a `cmd/config-key-audit`
   mode asserting that any config query with a `next_fetch_at` due arm carries the look-ahead.
   **NOT BUILT, genuinely unowned.** It is a new platform mechanism and needs a council round —
   which is exactly why it was not built here, and why the bug was not held open for it.
2. **The `interval '3 hours'` fallback goes stale in BOTH layers if the cadence changes.** Cadence
   still 21600 s as of 2026-09-02, so it is still the designed value. 653's guard catches drift
   only at apply time. LANDMINE filed (footprint `feed_due_lookahead.go`). **Unchanged.**
3. **`LoadDueSourcesAction` remains callerless** — fixed and guarded, but dead code. Re-verified
   2026-09-02 with the full config-text census the council asked for (`default_config::text LIKE
   '%load_due_sources%'` over all rows incl. snapshots and nested steps): **0**.
4. **The provocation twin — RESOLVED 2026-09-02, and it is NOT affected.**
   `provocation-feed-refresh` (21600 s, enabled) targets `provocation-feed-publisher`, whose entire
   live workflow is one step: `publish_feed` → `render_provocation_feed` → `complete`. It holds
   **no per-item due stamp** (`next_fetch_at` appears nowhere outside the content-feed paths) and
   writes only `scheduled_tasks.last_completed_at`. **No due predicate ⇒ no phase lock possible.**
   This residual is closed, not deferred.
5. **Cost/capacity** — §3, the owner's.
6. **The bug file stays OPEN until…** — **DISCHARGED.** Its first condition (the acceptance test)
   is met; its second (four run-hours/day) is unreachable at the current cap and was **reassigned
   to `bugs_open/316`**, not dropped.

## 6. ⚠ TRAPS — read these before re-measuring anything here

- **A STALLED PASS MIMICS THIS BUG EXACTLY.** `09-01 20:57:41` **FAILED** —
  `current_step=process_sites_iter_1_spawn_orchestrator`, `error` = *"reaper: stale EXECUTING_STEP
  for >4h"* — after serving **one** site, costing the other 13 a whole pass, i.e. a 12-hour gap
  with nothing to do with the phase lock. That is the known spawn→call handshake race (own owner).
  **Check `orchestration_states.status` for every pass you count, and exclude failures visibly.**
- **THE ACCEPTANCE TEST IN THE 08-26 HANDOFF WAS UNSATISFIABLE — do not copy its shape.** It
  demanded that *every* site from one pass reappear in the next, while §6 residual 5 of the *same
  document* predicted the fix would push demand past the cap and displace the surplus correctly.
  Measured: **4 of 8** discriminating sites were correctly capped out of the very pass that proved
  the fix works. `WRONG_CALLS.md` 2026-09-02 + LANDMINE *"a fix that relieves a bottleneck raises
  demand on the NEXT constraint"*. **Use the lower-bound test in the RUNBOOK instead.**
- **EXPECT EXACTLY ONE VACUOUS ROW PER PASS, AND IT IS THE FIRST SITE DISPATCHED.** Its fetch lag
  is smallest, so its due stamp lands closest to the next trigger and can fall either side.
  Observed both runs: `mortgagecalculator` −4 s, `fundamentallyai` −9 s. **Discard it** — it is
  served under either predicate, and counting it is the mistake this lane already made once.
- **THE TRIGGER DRIFTS.** ~:46 on 08-26 → **~:57/:58** on 09-02, ~12 min in a week. **Never reuse a
  hardcoded window from any doc here — read the fire times.**
- **`orchestration_states` PRUNES AT ~2 DAYS.** Any pinned test set expires. This is what made the
  08-26 test unrunnable — and, by forcing a rebuild from live data, is the only reason its design
  flaw was ever noticed.
- **`\d <table>` FIRST.** `scheduled_tasks`'s column is `interval_seconds` (not `schedule*`);
  `orchestration_states`'s is `error` (not `error_message`). Both cost round-trips here.
- **`date -u -d '<naive timestamp>'` PARSES THE INPUT IN LOCAL TIME** — `-u` affects output only.
  On this BST box a watcher armed for "20:53 UTC" fired at 19:53Z, and the off-by-one was
  misread as cluster clock skew (retracted; all three clocks agree within 4 s). **Append an
  explicit zone:** `date -d '… UTC' +%s`.

## 7. How to verify the whole thing from cold, in four commands

```bash
# 1. both halves present?  (config)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query' FROM agent_definitions \
 WHERE type='content-feed-trigger' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
#    expect: next_fetch_at <= NOW() + COALESCE((SELECT make_interval(secs => interval_seconds / 2.0) ...), interval '3 hours')

# 2. both halves present?  (Go — capability probe, ALWAYS with both controls)
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- grep -acF 'make_interval(secs => interval_seconds / 2.0)' /proc/1/exe  # 2
kubectl -n ai-persona-system exec $POD -- grep -acF 'interval_seconds / 7.0' /proc/1/exe                         # 0
kubectl -n ai-persona-system exec $POD -- grep -acF 'DispatchFeedSourcesAction: dispatched ingester' /proc/1/exe # 1

# 3. read the real fire times (never reuse a doc's)
#    then 4. run RUNBOOK "THE ACCEPTANCE TEST" on the newest healthy pass pair.
```

## 8. Lane docs

`docs/agent_docs/docs024_key_docs_latest/bugfix_410_feed_phase_lock/`
- **`SUMMARY_2026-09-02_410_feed_phase_lock.md`** — the milestone read-out, written to be read aloud
- `README_where_we_are.md` — the owner's plain-prose log (append only, newest at bottom)
- `NOTES_410_feed_phase_lock.md` — technical log, every query and every misstep (append only)
- `RUNBOOK_410_feed_phase_lock.md` — **the commands, including the acceptance test that works**
- `PLAN_2026-08-26_410_feed_phase_lock.md` — why candidates 2 and 3 were rejected
- `HANDOFF_2026-08-26_continue_here.md` — **SUPERSEDED**, kept as the record of the 26th

Related: `bugs_closed/410_…phase_locks_every_six_hour_news_site…` · **`bugs_open/316`** (owns the
cap) · `LANDMINES.md` (two-layer due predicate; the bottleneck/acceptance-test entry) ·
`WRONG_CALLS.md` 2026-09-02 · migration `653` + its `_ROLLBACK`.
