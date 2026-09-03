# HANDOFF — `bugfix_410_feed_phase_lock` — **THE LANE IS DONE.** Bug closed, fix proven, 24h change live and council-APPROVED

**Written 2026-09-03 ~09:30Z by the `bugfix_410_feed_phase_lock` lane. All times DB time
(`SELECT now()`). This is the cold-start doc — read this first.** It supersedes
`HANDOFF_2026-09-02_continue_here.md` and `HANDOFF_2026-08-26_continue_here.md`, both kept only as
the record of what was believed then.

⚠ **410 NAMES TWO UNRELATED BUGS.** This lane owns the **news-feed cadence phase lock**, at
`bugs_closed/410_HANDOFF_2026-08-26_next_fetch_at_stamped_at_fetch_time_phase_locks_every_six_hour_news_site_to_a_twelve_hour_cadence.md`.
The OTHER 410 — `..._three_seams_fail_toward_the_quiet_default...` — is a different case in
`bugs_open/` with its own lane (`bugfix_410_silent_scan_loss`). **Resolve by slug, never by number.**

---

## 1. Bottom line — nothing is outstanding

| | state |
|---|---|
| **The bug** (`bugs_closed/410`, phase-lock slug) | **CLOSED.** Fixed, live, and confirmed on real traffic on **three** separate occasions |
| **The owner's capacity decision** | **TAKEN AND LIVE** — cap left at 10, demand cut instead (migration 701, sources 6 h → 24 h) |
| **Migration 701** | **APPLIED, VERIFIED, and council-APPROVED** round 2 (corr `56c30292-3482-4d9c-8757-f287f1ef5a1b`) |
| **The prospective test** | **PASSED EXACTLY** — see §3 |
| **Chassis** | **v1.0.1356**, both replicas carry the look-ahead (probed with controls 2026-09-03) |

**The lane can be closed.** What remains below is (a) one optional observation completing the
cycle, (b) residuals that belong to other owners, and (c) traps for anyone who re-measures here.
None of it blocks.

## 2. What this lane did, in one screen

**The bug.** News sites should refresh every 6 h. The refresh pass ran on a fixed timetable, but
each fetch stamped `next_fetch_at = NOW() + fetch_interval` **from the moment of the fetch** —
10 s to 9 min after the timetable fired, because sites are processed sequentially. So a 6-hour
source fell due seconds *after* the next 6-hourly pass had already looked at it. It waited another
6 hours. The lock re-armed every cycle and **every run reported COMPLETED.** 12 of 14 sites.

**The fix** (council-approved round 1, corr `04c657d2`). A due look-ahead of **half the trigger
cadence** in *every* layer that asks "is this source due?" — serve on the **nearest** tick, not the
first tick strictly after. Cadence read live from `scheduled_tasks`; `COALESCE` falls back to
`interval '3 hours'`, never to a bare `NOW()`.

| layer | where | state |
|---|---|---|
| shared predicate | `platform/orchestration/actions/feed_due_lookahead.go` | LIVE |
| source-level (live) | `dispatch_feed_sources_action.go` due query | LIVE |
| source-level (dormant) | `feed_actions.go` `LoadDueSourcesAction` — 0 callers, fixed anyway | LIVE |
| site-level admission | migration `653_..._HOLD.sql` → `content-feed-trigger.find_news_sites` | APPLIED 2026-08-26 |

**The capacity follow-on** (council-approved round 2). The fix correctly made every 6h-only site
due at *every* pass, which pushed demand permanently above the `LIMIT 10` cap — 14 sites, ~9 h
effective cadence, 4 late by rotation each pass. **Owner ruled: leave the cap, cut demand.**
Migration **701** took sources 6 h → 24 h, changed the **column default** (the class fix — both
seeder `INSERT`s omit the column), and **spread the 14 sites across the four daily passes**.

> ⚠⚠ **THE TRIGGER CADENCE DELIBERATELY DID NOT MOVE, AND THIS IS LOAD-BEARING.**
> `fetch_interval` (24 h) ≠ cadence (6 h) is what makes the phase lock **structurally impossible**
> rather than merely fixed. Setting the cadence to 86400 s would re-create the precondition.
> 701's Guard 1 asserts 21600 s **at apply time only** — nothing enforces it afterwards (§5.1).
> And the 24 h interval **depends** on the look-ahead: without it, a source due at fetch+24 h is
> skipped by the tick that fires δ early and served one tick later — **24 h would silently become
> ~30 h.**

Commits: `201236b2a` `4da33f40e` `8ac0d0558` `f3ca6a9e5` `86aea3804` `b34c24f4c` `ac1727beb`
`703948e78` (closure) `d81396e2e` (migration 701) `e3b93aefe` `3b6d5ccc9` `f187a8709`.

## 3. The evidence the fix works — three independent confirmations

**Do not re-derive these; `orchestration_states` prunes at ~2 days and they will be gone.**

**(a) and (b), 2026-09-02 — the lower-bound test.** A site fetched in pass N cannot have been
fetched *before* it was dispatched in pass N, so its next due stamp is ≥ (pass-N dispatch) + 6 h.
If pass N+1 fires *before* that bound and admits it anyway, the old predicate cannot explain it.

| pass pair | sites served before their earliest possible due time |
|---|---|
| 02:57:57 → 08:58:27 | remortgagecalculator **+2 m 15 s**, idea.uk **+4 m 58 s**, vetcomparison **+10 m 57 s** |
| 08:58:27 → 14:58:58 | boxingonline **+2 m 33 s**, robot-hands **+3 m 58 s**, gaswholesalers **+6 m 50 s** |

`idea.uk` — the site the bug was filed on, skipped by 39 s on 08-26 — is among them. In each run
exactly one row was **vacuous** and discarded (mortgagecalculator −4 s, fundamentallyai −9 s);
that is not luck, see §6.

**(c) 2026-09-02/03 — the prospective slot test, recorded BEFORE the pass that tested it.**
Prediction written 18:23Z: *"the ~20:59Z pass dispatches EXACTLY ai-agent-orchestration,
fundamentallyai, mortgagecalculator, vetcomparison — and no others."*

**CONFIRMED exactly.** And the two following passes continued it without further prediction:

| pass | dispatched | slot |
|---|---|---|
| 09-02 20:59:03 | ai-agent-orchestration, fundamentallyai, mortgagecalculator, vetcomparison | **1 (4)** ✅ |
| 09-03 02:59:33 | boxingonline, gaswholesalers, relojistas, webdesign | **2 (4)** ✅ |
| 09-03 09:02:07 | dartsonline, idea.uk, remortgagecalculator | **3 (3)** ✅ |

## 4. The ONLY thing left, and it is optional

**Slot 4 has not yet served.** farmerinsurance.uk, loanandmortgagecalculator.co.uk,
robot-hands.com are due **09-03 14:58:57**. Watching that pass completes a full cycle across all
four slots. Everything it could show has already been shown three times over — **this is
tidiness, not evidence.** If it serves those three and no others, the lane is finished with a
complete cycle on the record.

```sql
SELECT s.domain, to_char(o.created_at,'DD HH24:MI:SS')
FROM orchestration_states o JOIN sites s ON s.id=o.site_id
WHERE o.owner_agent_type='content-feed-orchestrator'
  AND o.created_at > (SELECT last_triggered_at FROM scheduled_tasks WHERE name='content-feed-refresh')
ORDER BY o.created_at;
```

## 4b. The periodic health check — ⚠ THE VERSION IN THE 09-02 HANDOFF WAS WRONG. USE THIS ONE.

> **CORRECTED 2026-09-03.** The check I shipped yesterday returns **`slots = 56, busiest = 3`**
> against its own stated expectation of `slots = 4` — **a false alarm I planted, not degradation.**
> Two errors: (1) it counted `DISTINCT next_fetch_at` over **sources**, and once each source is
> stamped `NOW() + 24h` at its own second, one site fans out across 4–9 values; I derived the
> expected value from the transient post-migration moment when all stamps were artificially
> identical. (2) My first fix — bucketing by `floor((due − lt)/6h)` — was **also** wrong: it omits
> the look-ahead, so the boundary falls mid-pass and splits one pass across two buckets.

**A site is served by the first trigger `T >= due − lookahead`, so its serving pass is
`ceil((due − lookahead − last_triggered_at) / cadence)`.** Both cadence and look-ahead are read
live, so this does not go stale if either changes:

```sql
WITH t AS (SELECT last_triggered_at AS lt, interval_seconds AS secs,
                  make_interval(secs => interval_seconds / 2.0) AS lookahead
             FROM scheduled_tasks WHERE name='content-feed-refresh'),
site_due AS (SELECT cs.site_id, min(cs.next_fetch_at) AS due
               FROM content_sources cs WHERE cs.is_active GROUP BY cs.site_id)
SELECT ceil(extract(epoch from (sd.due - t.lookahead - t.lt)) / t.secs)::int AS served_by_pass,
       count(*) AS sites, string_agg(s.domain, ', ' ORDER BY s.domain) AS domains
FROM site_due sd CROSS JOIN t JOIN sites s ON s.id = sd.site_id
GROUP BY 1 ORDER BY 1;
```
`[MEASURED 2026-09-03 09:15Z]` **4 passes, 3 / 4 / 4 / 3 sites, busiest 4 vs cap 10** — each group
exactly its original slot. **Healthy.** Note it uses `min(next_fetch_at)` per **site**, because a
site is admitted when its *earliest* source is due.

**Why it can degrade, and there are TWO causes, not one:**
1. **A failed pass.** That slot's sites are not served, stay due, and are picked up by the next
   pass — where they re-stamp. The two slots merge permanently. Seen 09-01 20:57:41 (spawn
   handshake, served 1 site, reaped after 4 h).
2. **⚠ An out-of-band producer — the one I missed.** The checker layer can dispatch a
   `content-feed-orchestrator` run outside the trigger's slots. Observed 09-02 **22:13:01**, when
   no trigger fired: a `stale_news_section` repair for vetcomparison
   (`threshold_hours: 72`, `newest_item_age_hrs: 321`, different parent). Any such run re-stamps
   that site to *its* moment + 24 h. Here it was harmless (21:08 → 21:10:32, still slot 1), but a
   repair at, say, 05:00 would park a site between passes for good.

**The repair, if it degrades:** statement 3 of migration 701, run alone (the
`WITH slots AS … base AS …` UPDATE). Idempotent; re-spreads from the next expected pass.

## 5. Residuals — status of each, and who owns it

1. **Go↔config parity is not enforced on a schedule, and since 701 it guards TWO things.**
   (a) a future migration could rewrite `find_news_sites` without the look-ahead; (b) setting the
   cadence to 86400 s would silently restore the phase-lock precondition. Candidate: a
   `cmd/config-key-audit` mode asserting both. **NOT BUILT, genuinely unowned.** New platform
   mechanism, needs its own council round. Council flagged (a) and (b) as low-severity, "noting for
   tracking, not blocking".
2. **`interval '3 hours'` fallback** — cadence still 21600 s, so still correct. LANDMINE filed.
3. **`LoadDueSourcesAction` still callerless** — re-verified by full config-text census: **0**.
4. **Provocation twin — RESOLVED, NOT affected.** Its whole live workflow is
   `publish_feed → render_provocation_feed → complete`; no per-item due stamp exists, so no phase
   lock can form. Closed, not deferred.
5. **Cost/capacity — DECIDED, SHIPPED, APPROVED.** §2.
6. **Council round 2's medium objection, answered and recorded.** 701's `UPDATE` had no
   `is_active` filter while its rationale, Guard 2 and verify all reasoned about "73 active
   sources". `[MEASURED 2026-09-03]` there are **0 inactive rows** and the backup holds **73**, so
   the broader write was a no-op — **no damage, but the mismatch was real**: the write was
   unscoped, the disclosure and verification scoped. If 701 is ever re-run, add the filter *and*
   loosen the `slots <> 4` verify to `<= 4` (council's other low-severity point — a smaller estate
   would trip a correct spread).
7. **⚠ NOT this lane's, but found here:** vetcomparison.uk's newest news **item** is **13.4 days**
   old (`newest_item_age_hrs: 321`) against a 72 h threshold. That is a content-supply problem —
   the site has exactly **one** source — not a cadence problem. **701 neither causes nor fixes
   it.** Flagged so nobody attributes it to the 24 h change.

## 6. ⚠ TRAPS — read before re-measuring anything here

- **EXPECT EXACTLY ONE VACUOUS ROW PER PASS IN THE LOWER-BOUND TEST, AND IT IS THE FIRST SITE
  DISPATCHED.** Its fetch lag is smallest, so its due stamp lands closest to the next trigger and
  can fall either side. Observed both runs: mortgagecalculator **−4 s**, fundamentallyai **−9 s**.
  **Discard it** — it is served under either predicate.
- **A STALLED PASS MIMICS THIS BUG EXACTLY.** `FAILED` +
  `current_step=process_sites_iter_1_spawn_orchestrator` + `error` starting `reaper: stale
  EXECUTING_STEP` is the spawn→call handshake race, own owner. It gave 13 sites a 12 h gap on
  09-01. **Check `status` for every pass you count and exclude failures visibly.**
- **THE TRIGGER DRIFTS.** ~:46 (08-26) → ~:57 (09-02) → **~:59/:02** (09-03). **Never reuse a
  hardcoded window from any doc here — read the fire times.**
- **`orchestration_states` PRUNES AT ~2 DAYS.** Any pinned test set expires. This is what made the
  08-26 acceptance test unrunnable — and, by forcing a rebuild, is the only reason its design flaw
  was noticed.
- **AN ACCEPTANCE TEST WRITTEN BESIDE A FIX CAN BE UNSATISFIABLE BY THAT FIX.** The 08-26 handoff
  demanded every site from one pass reappear in the next, while its own residual 5 predicted the
  cap would start binding. `WRONG_CALLS.md` 2026-09-02 (a) + LANDMINE.
- **`_` IS A WILDCARD IN SQL `LIKE`.** `LIKE '%content_sources%'` matched `research_content.sources`
  and gave me a phantom third consumer, which I reported before catching it. Use `~` (POSIX, where
  `_` is literal) and give every census a positive control. **That trap is in `LANDMINES.md` three
  times and I hit it anyway** — the SessionStart hook only matches landmines against files already
  *dirty*, never a table or a SQL construct, so **grep LANDMINES for the construct yourself.**
  `WRONG_CALLS.md` 2026-09-02 (b).
- **`\d <table>` FIRST.** `scheduled_tasks` → `interval_seconds`; `orchestration_states` → `error`
  (not `error_message`). Both cost round-trips here.
- **`date -u -d '<naive timestamp>'` PARSES THE INPUT IN LOCAL TIME** — `-u` affects output only.
  Cost an hour of false "clock skew" on 08-26. **Append an explicit zone.**
- **DERIVING A CHECK'S EXPECTED VALUE FROM A TRANSIENT STATE.** §4b's first version expected
  `slots = 4` because that was true for the few hours when all stamps were artificially identical.
  **Ask what the value will be in steady state, not what it is right after your change.**

## 7. Verify the whole thing from cold, in four steps

```bash
# 1. config half
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query' FROM agent_definitions \
 WHERE type='content-feed-trigger' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
#    expect the look-ahead: NOW() + COALESCE((SELECT make_interval(secs => interval_seconds / 2.0) ...), interval '3 hours')

# 2. Go half — capability probe, ALWAYS with both controls
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- grep -acF 'make_interval(secs => interval_seconds / 2.0)' /proc/1/exe  # 2
kubectl -n ai-persona-system exec $POD -- grep -acF 'interval_seconds / 7.0' /proc/1/exe                         # 0
kubectl -n ai-persona-system exec $POD -- grep -acF 'DispatchFeedSourcesAction: dispatched ingester' /proc/1/exe # 1
```
3. **The 24 h state:** all active sources at `24:00:00`, column default `'24:00:00'::interval`.
4. **The spread:** §4b's corrected query — 4 passes, busiest ≤ 4.

⚠ The `build provenance` startup line scrolls out of `--tail=3000` within ~25 minutes on this
service. Empty there means "not in range", **not** "unstamped". Never substitute a *discovery*
grep for a 40-hex string — it matches Go's internal digit table.

## 8. Lane docs

`docs/agent_docs/docs024_key_docs_latest/bugfix_410_feed_phase_lock/`
- **`SUMMARY_2026-09-02_410_feed_phase_lock.md`** — milestone read-out, written to be read aloud
- `README_where_we_are.md` — the owner's plain-prose log (append only)
- `NOTES_410_feed_phase_lock.md` — technical log, every query and every misstep (append only)
- `RUNBOOK_410_feed_phase_lock.md` — the commands, incl. the acceptance test that works
- `PLAN_2026-08-26_...` — why fix candidates 2 and 3 were rejected
- `HANDOFF_2026-09-02_...`, `HANDOFF_2026-08-26_...` — **SUPERSEDED**, kept as the record

Related: `bugs_closed/410` (phase-lock slug) · **`bugs_open/316`** (owns the cap; CONTRIB filed
2026-09-02) · migrations `653` and **`701`** + their `_ROLLBACK`s · `LANDMINES.md` ·
`WRONG_CALLS.md` 2026-09-02 (a) and (b).
