# HANDOFF — `bugfix_410_feed_phase_lock` — BUG CLOSED · OWNER DECISIONS TAKEN AND LIVE (migration 701) · one confirmation left

**Written 2026-09-02 ~16:30Z, UPDATED ~18:45Z (migration 701 applied) by the `bugfix_410_feed_phase_lock` lane. All times in this lane are
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

> ### ⚠ UPDATED 2026-09-02 18:2xZ — THE OWNER DECISIONS IN §3 HAVE BEEN TAKEN AND ARE LIVE
>
> **Do not re-raise them.** The owner ruled: **leave the cap at 10**, and **cut demand instead —
> news sources move from a 6 h to a 24 h `fetch_interval`, sub-6h sources included.** Shipped as
> **migration 701**, applied and independently verified 2026-09-02 (commit `d81396e2e`, council
> corr `56c30292-3482-4d9c-8757-f287f1ef5a1b`, **verdict still owed a read**). §3 is rewritten
> below as the record of what was decided and why.
>
> **The trigger cadence deliberately did NOT move** — see §3. `24h ≠ 6h` is load-bearing: it makes
> this bug's phase lock *structurally impossible* rather than merely fixed.

**Nothing is required of anyone to keep it working.** What remains is:

- **one combined confirmation** (~20:59Z tonight, 10 minutes) — it settles BOTH the `v1.0.1354`
  check and 701's slot arithmetic in one pass. Prediction recorded in advance — §4;
- **one periodic check** that 701's pass-slot spread has not silently degraded — §4b;
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

## 3. ✅ THE DECISIONS — TAKEN 2026-09-02, LIVE. Record, not a question.

**Owner ruling: do NOT touch the cap. Cut demand instead — sources go from 6 h to 24 h.**
Shipped as **migration `701_content_sources_daily_fetch_interval.sql`** (+ `_ROLLBACK`),
applied by hand and independently verified 2026-09-02, commit `d81396e2e`.

### Why only ONE knob moved — this is the part to understand before touching anything here

"Reduce the frequency to 24 h" had two possible targets. They are not equivalent:

| | passes/day | sites due per pass vs cap 10 | outcome |
|---|---|---|---|
| **`fetch_interval` → 24h, cadence STAYS 21600 s** ✅ **shipped** | 4 | **~3.5** | cap never binds, **no cap change needed** |
| cadence → 86400 s | 1 | **14 → BINDS HARD** | 4 sites get nothing that day, wait 48 h; forces the deferred cap change |
| both → 24 h | 1 | **14 → BINDS HARD** | same starvation, **plus** re-creates `fetch_interval == cadence`, the exact precondition of this bug's phase lock |

> ⚠⚠ **THE `24h ≠ 6h` INEQUALITY IS LOAD-BEARING, NOT INCIDENTAL.** It is what makes the phase
> lock *structurally impossible* rather than merely fixed by the look-ahead. **If anyone ever sets
> `content-feed-refresh.interval_seconds` to 86400, this premise is void** — re-read
> `bugs_closed/410` §1 first. Migration 701's Guard 1 asserts 21600 s **at apply time only**;
> nothing enforces it afterwards. That is §5 residual 1 acquiring a second thing to guard.

### What 701 actually does, and why each part is load-bearing

1. **`ALTER COLUMN fetch_interval SET DEFAULT '24:00:00'`** — the **class fix**. `[MEASURED
   2026-09-02]` **both** `INSERT INTO content_sources` in `seed_content_sources_action.go`
   (`:284`, `:355`) **omit the column**, so new sources inherit the default. Without this line the
   change lasts until the next site build. No Go hardcodes the interval ⇒ **no image roll needed**.
2. **`UPDATE` the 73 existing rows** — including the sub-6h ones (relojistas 3h×2/4h×2,
   dartsonline 4h×2), per the owner. ⚠ **Those were the only always-due CONTROL sites in every
   cadence census in this lane. There is no sub-cadence control left.**
3. **Re-stamp `next_fetch_at` into four pass slots.** Without it all 14 sites are already due and
   settle **10-and-4** — and **10 exactly fills the cap**, so the fifteenth news site would restart
   contention immediately. Now **4/4/3/3**, six slots of headroom.

### Verified independently after apply (not trusting the in-transaction check)

| check | result |
|---|---|
| active sources by interval | **73 sources / 14 sites, all `24:00:00`** |
| **column default** | **`'24:00:00'::interval`** |
| slot spread | **4 / 4 / 3 / 3** sites; busiest **4** vs cap **10** |
| cadence (Guard 1's premise) | **21600** |
| backup for rollback | `bak_content_sources_fetch_interval_20260902`, **73 rows** |

Applied **by hand, alone** — NOT `run-migrations.sh --apply`, which takes **every** pending file
including other sessions'. Then registered with `--record-only` so it is not re-applied.

### Expected consequences — state these before someone files one as a regression

- Fetch volume **~180 → ~73 source-fetches/day** (~60% cut); every site refreshes **once per 24 h**
  (was ~9 h since the fix, 12 h before it).
- ⚠ **LCO-009 / `--capped-schedule-ordering` STOPS reporting cap hits.** Migration 653's header
  predicted the opposite and was right for the estate as it stood; 701 is why it stops. **A silent
  capped-schedule check from today is the EXPECTED result and is NOT evidence the check works** —
  if you need to know it still fires, give it a demand control rather than reading the zero.

### If you need to undo it

`701_..._ROLLBACK.sql`, by hand. It restores **per-row** values from the backup table, not the old
default — the pre-change intervals were **not uniform**, so resetting everything to 6 h would be a
silent data change dressed as a rollback. Note what rolling back re-enters: `fetch_interval ==
cadence` (the phase-lock precondition, survivable because the look-ahead is live) **and** the cap
contention of `bugs_open/316`.

## 4. THE ONE CONFIRMATION LEFT — ~20:59Z tonight, 10 minutes, settles TWO things at once

Tonight's pass is the first since **both** the new chassis build **and** migration 701. It tests
the fix on `v1.0.1354` *and* 701's slot arithmetic in one go.

**Prediction, recorded 2026-09-02 18:23Z BEFORE the pass that tests it.** The pass is expected at
**20:58:57Z**; its look-ahead reaches **23:58:57Z**; slot 2 is due **02:58:57Z**, a full **3 h
outside**. Therefore:

> **The ~20:59Z pass dispatches EXACTLY these four sites and no others:**
> **ai-agent-orchestration.com · fundamentallyai.com · mortgagecalculator.co.uk · vetcomparison.uk**
>
> **A fifth site** (especially a slot-2 one: boxingonline, gaswholesalers, relojistas, webdesign)
> **refutes the slot arithmetic.** **Fewer than four** means something is not reaching the
> decision — re-check the config half and the binary probe (§7) before anything else.

```sql
SELECT s.domain, to_char(o.created_at,'DD HH24:MI:SS')
FROM orchestration_states o JOIN sites s ON s.id=o.site_id
WHERE o.owner_agent_type='content-feed-orchestrator'
  AND o.created_at > (SELECT last_triggered_at FROM scheduled_tasks WHERE name='content-feed-refresh')
ORDER BY o.created_at;
```

The four slots, for reference:

| slot | due | sites |
|---|---|---|
| 1 | 09-02 20:58:57 | ai-agent-orchestration, fundamentallyai, mortgagecalculator, vetcomparison |
| 2 | 09-03 02:58:57 | boxingonline, gaswholesalers, relojistas, webdesign |
| 3 | 09-03 08:58:57 | dartsonline, idea.uk, remortgagecalculator |
| 4 | 09-03 14:58:57 | farmerinsurance, loanandmortgagecalculator, robot-hands |

**On the chassis half specifically:** `v1.0.1354` (rolled 15:39/15:53Z) was re-probed on **both**
replicas and both carry the look-ahead (`make_interval(secs => interval_seconds / 2.0)` → **2**;
negative control `interval_seconds / 7.0` → **0**; positive control → **1**). Every live
confirmation *before* tonight was produced by `v1.0.1352`, so the new binary carries the capability
but has not yet been exercised. Tonight gives it a turn.

⚠ **Provenance note:** the `build provenance` startup line was **already out of `--tail=3000`** on
pods up only ~25 minutes. An empty grep there means "not in range", **not** "unstamped". Do not
substitute a *discovery* grep for a 40-hex string — it matches Go's internal digit table and
returns the same wrong answer on every service.

## 4b. ⚠ THE ONE THING WORTH RE-CHECKING PERIODICALLY — 701's spread DEGRADES SILENTLY

**A failed pass permanently merges two slots, and nothing complains.** If a pass fails — as
`09-01 20:57:41` did (spawn handshake, served 1 site, reaped after 4 h) — that slot's sites are
not served, stay due, and are picked up by the **next** pass, where they re-stamp to *that* pass's
time. 4+4 = 8 is still under the cap, so nothing breaks; the headroom 701 bought is just spent,
one failure at a time, until the estate is back to piling into one or two passes.

```sql
SELECT count(DISTINCT next_fetch_at) AS slots, max(c) AS busiest FROM (
  SELECT next_fetch_at, count(DISTINCT site_id) c FROM content_sources
   WHERE is_active GROUP BY next_fetch_at) x;
```
**Expect `slots = 4`, `busiest <= 4`.** Fewer slots or a busier one means it has degraded. **The
fix is statement 3 of migration 701, run alone** (the `WITH slots AS ... base AS ...` UPDATE) — it
is idempotent and re-spreads from the next expected pass. This is not urgent and not a defect; it
is a property of spreading work with no scheduler that owns the spread.

## 5. Residuals at close — each with its actual status

1. **Go↔config predicate parity is NOT enforced against live config on a schedule — and since
   701 it has a SECOND thing to guard.** (a) A future migration could rewrite `find_news_sites`
   without the look-ahead and **nothing would fire**; the only signal would be sites drifting back
   to a phase-locked cadence. (b) **Since 701, the `24h ≠ 6h` inequality is load-bearing too** —
   setting `content-feed-refresh.interval_seconds` to 86400 would silently restore the phase-lock
   precondition, and 701's Guard 1 only checks it *at apply time*. Candidate: a
   `cmd/config-key-audit` mode asserting both — that any config query with a `next_fetch_at` due
   arm carries the look-ahead, and that no active source's `fetch_interval` equals the trigger
   cadence. **NOT BUILT, genuinely unowned.** It is a new platform mechanism needing a council
   round — which is why it was not built here, and why the bug was not held open for it.
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
5. **Cost/capacity** — **DECIDED AND SHIPPED**, see §3. Cap left at 10; demand cut instead via
   migration 701 (sources 6 h → 24 h). Fetch volume ~180 → ~73/day. ⚠ **Council verdict on 701
   (corr `56c30292-3482-4d9c-8757-f287f1ef5a1b`) is still owed a read** — the change is already
   live on the shared branch, so a REVISE must be acted on, not argued with:
   `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;`
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
