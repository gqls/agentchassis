# HANDOFF 2026-08-26 — `bugs_open/396`: both fixes are LIVE, EXERCISED and APPROVED. Nothing is blocked and nothing is owed.

**Read this box. Everything below it is background or recipe.**

> ## STATE
>
> **Both halves of the fix are in the running binary and the config is applied.** Chassis
> **`v1.0.1341`**, binary-proven on **both** replicas with a present-control and an absent-control
> in the same run.
>
> | piece | state |
> |---|---|
> | `status_override` allow-list — council **APPROVED** `9c16eb83` | **LIVE** |
> | `sites.lock_except_item_ids` — migration `632` | **APPLIED**, inert |
> | `honour_site_lock` arm in `LoadWorkItemsAction` | **LIVE** |
> | migration `633` (the held config half) | **APPLIED 2026-08-26**, hold condition met and proven |
> | park verb `park_work_items()` — migration `621`, WII-034 | **applied, DEMOTED** — see §5 |
>
> ## ✅ UPDATE 2026-08-26 09:0xZ — CREDITS RESTORED, BOTH OWED STEPS DONE
>
> The fleet came back at **08:58:28Z** (verified with `GROUP BY success`, not a bare count).
>
> **1. The exception list is EXERCISED.** Both predicates run verbatim against live data on
> `cv1.co.uk` (6 dispatchable items), inside a transaction that was **rolled back**:
> **unlocked → 6 items · locked, no exception → 0 · locked, one exception → exactly 1, the right
> one.** State B is the one that matters: the lock still holds with the exception column present.
> Negative control after rollback: site unlocked, exception list NULL, its queue never actually held.
> ⚠ **Still unproven: that the SCHEDULER picks such a site in production** — `find_dispatchable_site`
> is `ORDER BY created_at ASC` fleet-wide and 1,398 older items are queued, so it cannot be forced.
> **The gap is the tick, not the logic**, and it closes the first time a locked-with-exception site
> wins the ordering.
>
> **2. The council round is APPROVED** — `175df761` r2, **12 reviewers, 1 gating-level advisory,
> none high-severity**. r1's REVISE was answered in code rather than in prose.
>
> **NOTHING IS BLOCKED AND NOTHING IS OWED.** Two residuals are recorded and neither is urgent:
>
> - ⚠ **The two spellings of the lock rule are guarded only inside Go.** The approving round's own
>   advisory: `TestSiteLockExceptionSQLIsNotTheSelectorSpelling` cannot reach a **migration** author,
>   because migration SQL is text and is compiled against nothing. Their only guard is the
>   `sites.locked_at` entry in `LANDMINES.md`, which now carries both spellings verbatim and the
>   failure in each direction. **Anyone writing a migration that touches `find_dispatchable_site`
>   must read that entry first.**
> - **Nothing stops a raw `UPDATE … SET status='deferred'`.** Short of a trigger, nothing can — the
>   standing residual on `396`.
>
> **For the next submission, not this one:** the guardian noted that a `_HOLD` migration editing
> `agent_definitions` should carry `operation: config_change` in the plan, not `add`. Advisory, and
> worth getting right next time.

> ## ⛔ (SUPERSEDED — kept for the record) THE TWO THINGS OWED, AND WHY YOU COULD NOT DO THEM
>
> 1. **Exercise the exception list end to end** — lock a site, except one item, confirm that item
>    dispatches and its siblings do not. Recipe in §3.
> 2. **Resubmit the council round.** `175df761` r2 ran and died `complete_invalid`: *"no reviewer
>    produced a readable opinion (5 abstained, 12 unreadable) — a council with no opinions cannot
>    decide."* **It was never judged.** The JSON is already fixed and ready to fire (§3).
>
> **Both need a working fleet, and the fleet is down.** Anthropic **credits exhausted** — last
> success `2026-08-25 23:46:29Z`, first failure 41 s later, **631 consecutive failures**. The build
> queue is stalled at **1,399 `triaged` / 0 `claimed`**, oldest triaged 08-18.
>
> ⚠ **DO NOT RE-FILE THE OUTAGE.** Three other lanes diagnosed it hours before I did and the owner
> is already push-notified (`a12692649`, `d0d1deb42`, `60d2a2508`, `c4f52ea3a`). It is
> `bugs_open/243`'s class. **It needs an owner top-up, not a session.**
>
> **First move on a new session: check whether the fleet is back (§1). If it is, do §3 in order.
> If it is not, there is nothing to do on this lane.**

## 1. Is the fleet back? — the ONE query, and the trap in it

```sql
SELECT provider, success, count(*), to_char(max(created_at),'MM-DD HH24:MI') AS newest
FROM llm_call_log
WHERE provider='anthropic' AND created_at > now() - interval '30 minutes'
GROUP BY 1,2 ORDER BY 3 DESC;
```

⚠⚠ **`GROUP BY success` IS THE WHOLE QUERY. An ungrouped `count(*)` will tell you the fleet is
healthy while it is down.** I made exactly that error on 2026-08-26 and was one sentence from
filing the outage as its opposite: I saw `ai_endpoint_health` reading *unhealthy* and 569 calls
since, and read it as "the health row is stale". **All 569 had `success = false`.** A failing
endpoint produces *more* traffic, not less, because everything retries. `WRONG_CALLS.md` entry 8.

Corroborate at the queue, which is the honest downstream signal:

```sql
SELECT status, count(*) FROM site_work_items
WHERE status IN ('triaged','approved','claimed') GROUP BY 1;
```

**`claimed = 0` with a large `triaged` is the stall.** `claim_work_item` gates on
`ai_endpoint_health`, so a false row releases every claim fleet-wide.

## 2. What was actually built, and why it is not what this lane first proposed

**The bug:** work items parked at `status='deferred'` with a named `handler_agent` are selected by
nothing (`claim_work_item` takes `triaged`/`approved`; the promoter takes `detected`) **and** still
hold their `idx_swi_dedup` slot, so the detector cannot re-file and any other session dispatching
that page hits `23505` — a failure that reads as *"already queued"* and means *"queued and
abandoned"*. One such row blocked `bugs_open/328` for **22 days** and completed **2 minutes** after
being re-armed.

**This lane's first answer was wrong and the council caught it.** I built a park verb on the premise
that nothing could hold a site's queue. `sites.locked_at` / `locked_by` **already existed** — live
on 3 of 51 sites, gating `find_dispatchable_site`, with admin endpoints — and **it mutates no
work-item row**, so it strands nothing and the 22-day stall could not have happened under it.

**What was actually missing was narrower:** the lock is all-or-nothing. That gap is why the
`mortgagecalculator_couk_adoption` lane wrote a 15-second auto-defer backstop (its own handoff calls
the lock *"(a)"* and item status *"(b) the finer control"*), and that backstop minted 38 of the 52
unattributable rows.

**So the shipped fix is `sites.lock_except_item_ids`** — "hold this site EXCEPT these items",
honoured at both gates, **without changing any work item's status**.

## 3. THE RECIPE — do these in order, once the fleet is back

### (a) Resubmit the council round

The JSON is already corrected (duplicate paths merged, 6 edits):

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/deferred_work_item_park/submission_396_site_lock_exception_r2.json \
  175df761-d4bd-4db3-aa69-3a77a4b7fcd8
```

Passing the old correlation keeps the trail accumulating. **Budget ~30 min**, and a missing
orchestration row is usually latency — **but check `current_step` before assuming**: mine read
`complete_invalid`, which is not latency and not a content refusal either.

### (b) Exercise the exception list — the thing no test can prove

Everything so far is unit-tested and config-verified. **Nothing has watched it work.**

```sql
-- 1. pick a quiet site with at least 2 dispatchable items
SELECT s.domain, s.id, count(*) FILTER (WHERE w.status IN ('triaged','approved')) AS dispatchable
FROM sites s JOIN site_work_items w ON w.site_id = s.id
WHERE s.locked_at IS NULL GROUP BY 1,2 HAVING count(*) FILTER (WHERE w.status IN ('triaged','approved')) >= 2
ORDER BY 3 LIMIT 5;

-- 2. lock it, and except exactly ONE item
UPDATE sites SET locked_at = now(), locked_by = '396 exception-list acceptance test',
       lock_except_item_ids = ARRAY['<the one item id>']::uuid[]
WHERE id = '<site id>';

-- 3. WAIT for a dispatch tick, then assert BOTH halves
SELECT id, status FROM site_work_items WHERE site_id='<site id>' AND status IN ('triaged','approved','claimed','complete');
```

**The pass condition is TWO-SIDED and the second half is the whole test:**
the excepted item moves off `triaged` **AND every other item on that site does not**. Without the
second, a fix that simply ignored the lock would pass.

```sql
-- 4. ALWAYS release
UPDATE sites SET locked_at=NULL, locked_by=NULL, lock_except_item_ids=NULL WHERE id='<site id>';
```

⚠ **Use a site nobody is working.** `scripts/who-owns.py` and a glance at the lane dirs first.

## 4. ⚠⚠ THE TRAPS — read before touching any of this

- **THE LOCK IS ENFORCED AT EXACTLY ONE GATE.** `find_dispatchable_site` selects a **SITE**, not an
  item. `LoadWorkItemsAction`, which runs next, **has never checked `sites.locked_at`**. That is why
  the config half was held behind the binary: applied early it converts a full site hold into **no
  hold at all**, on exactly the sites somebody deliberately locked.
- **`load_work_item_actions.go:134` LOOKS like a second gate and is not.** It is inside
  **`WriteBuildItemsAction`**, and **its log line misnames its own function**:
  `"LoadWorkItemsAction: site is locked, skipping"`. That string cost me an hour.
- **`find_dispatchable_site`'s SQL is under `config.query`, NOT `config.pre_query`.** A migration
  reading the wrong key patches nothing and reports success.
- **The two spellings of the lock rule are deliberately different and must stay different.** The Go
  fragment is per-site and `$1`-parameterised; the selector is a cross-site scan with no `$1` and is
  spelled against the joined alias. `TestSiteLockExceptionSQLIsNotTheSelectorSpelling` fails if
  anyone merges them — a council reviewer already tried to, in their head.
- **`--record-only` REFUSES a `_HOLD` file** ("an UPPERCASE-suffixed sidecar"). Held migrations are
  invisible to `schema_migrations` unless you hand-write the row (`applied_by='hand-recorded'`,
  same as `610_..._HOLD`). **So "was the held half applied?" can only be answered from live config.**
- **An empty `grep` for `build provenance` means NOT IN RANGE, not unstamped** — it is a startup
  line and it scrolls. Use the binary probe, never `strings`, always with a present-control **and**
  an absent-control in the same run.

## 5. The park verb (`621` / WII-034) — applied, demoted, deliberately kept

It is **not** the answer to "hold a site" and its register entry says so. Two objections stand
against parking as a mechanism, both verified: a parked row **still holds its dedup slot**, and
`work_item_retraction.go:205` can **close a deliberate park without any unpark route**. It is kept
only for parking specific items on an **unlocked** site, where the alternative is the raw `UPDATE`
that caused this bug.

## 6. What is NOT done — stated so nobody reads silence as completion

- **No parked row has been released.** 52 are unstamped (mortgagecalculator 38, idea.uk 14, both
  sites now showing a zero live queue, so the holds look expired) and **62 are stamped, 60 of them
  carrying another lane's LIVE `"un-park after rebuild verify"` condition.** ⚠ **Do not sweep them
  — ask the holders.** `unpark_work_items` is scoped to one `parked_by` for exactly this reason.
- **Nothing stops a raw `UPDATE … SET status='deferred'`.** Short of a trigger, nothing can. Stated
  residual in `396` and WII-036.
- **The exception list has never been exercised.** §3(b).

## 7. Where everything lives

- **Bug:** `bugs_open/396_HANDOFF_2026-08-25_…md` — §6b is the corrected direction, §6c the live state.
- **This lane:** PLAN · NOTES (append-only, newest at the bottom — **the cold-start read**, and it
  carries all four missteps) · RUNBOOK (every query with its gotcha) · README_where_we_are (owner
  prose) · `submission_396_*.json`.
- **Register:** **WII-036** (the exception list), **WII-034** (the park verb, amended).
- **Code:** `work_items_common.go` (`siteLockExceptionSQL`, `workItemStatusOverrideAllowed`,
  `statusOverrideAllowed`), `load_work_item_actions.go` (`LoadWorkItemsAction`, `FailWorkItemAction`),
  `site_lock_exception_test.go`, `status_override_allowlist_test.go`.
- **Migrations:** `621` (+ROLLBACK), `632` (+ROLLBACK), `633_..._HOLD` (+ROLLBACK).
- **Councils:** `9c16eb83` APPROVED · `ed821065` REVISE (right, and acted on) · `175df761` r1 REVISE,
  r2 never judged.

## 8. The four missteps this lane made, because they are the transferable part

All four are in `WRONG_CALLS.md` and all four are **not looking**, not bad reasoning:

1. **Enumerated `spec`, never `result`** — 62 rows I called "no trace of any kind" were fully
   stamped. I had written the *"enumerate the stamps you don't know"* lesson two hours earlier and
   applied it to one column.
2. **Grepped the code for a writer, never the docs** — on a tree whose writers are sessions that
   write down what they did. The answer was in another lane's handoff, verbatim, for weeks.
3. **Asserted `sites.locked_at` didn't exist** — into a formal submission. It has its own landmine
   entry. So does the selector/loader divergence I then "discovered" by reading Go for twenty
   minutes.
4. **Counted calls and reported them as successes** — nearly filing a live outage as its opposite,
   with a matching landmine making the wrong reading feel confirmed.

**The check, now the first line of the RUNBOOK:** before writing *"nothing does X"* or *"X does not
exist"* — `grep -n "<symbol>" LANDMINES.md`, then `grep -rn "<symbol>" docs/`, **then** the source.
**An asserted absence is a claim about your search, not about the estate.**
