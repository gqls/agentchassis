# 070 — `stale-work-item-reaper` parks fresh requests: it keys on row age, not time spent in `triaged`

**Filed:** 2026-07-25, from the fundamentallyai.com stage-2 rollout.
**Severity:** low blast radius, high confusion cost. It does not lose work
silently in the normal (loader-created) path; it reliably kills **re-queued**
work items and mislabels them as stale.
**Status:** OPEN — diagnosed, verified live, not fixed. **DB-config only: the fix
is live immediately, no image roll.**
**Blocks:** `features_open/021` (operator bulk page rebuild) treats this as a
prerequisite.

## Symptom

Re-queueing an existing build work item (`status='triaged'`) has it flipped to
`unresolved` within the hour, with its summary prefixed `[stale: triaged 48h+]`
— even though it entered `triaged` minutes ago. Repeat, and the prefix
accumulates. Three rows on fundamentallyai.com now read:

```
[stale: triaged 48h+] [stale: triaged 48h+] Build index page (not_built)
```

The label is false twice over: the item had not been triaged for 48h, and the
page in question is deployed, not `not_built`.

## Root cause

`scheduled_tasks.pre_query` for `stale-work-item-reaper` (verified live
2026-07-25, `enabled=true`, `interval_seconds=3600`):

```sql
UPDATE site_work_items
SET status = 'unresolved',
    summary = '[stale: triaged 48h+] ' || summary,
    updated_at = NOW()
WHERE status = 'triaged'
  AND pipeline = 'build'
  AND created_at < NOW() - INTERVAL '48 hours'   -- <-- row birth, not triage entry
  AND claimed_at IS NULL
```

The task's own description states the intent: *"Marks triaged work items as
unresolved if they have been waiting 48h+ **without being claimed**."* The
predicate measures something else — how long ago the **row was created**. For an
item created and triaged in one go and never touched again, the two coincide,
which is why this has never bitten the loader-created path. For a re-queued item
they diverge completely: a row created 2026-07-20 is *born* eligible, so any
re-queue of it is reapable on the reaper's very next tick.

Source of the definition: `docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql:1137`
(the live row matches the file).

### Why it usually *looks* fine, and when it doesn't

`build-pipeline-trigger` runs every **120s** while the reaper runs every
**3600s**, so a single re-queued item is normally claimed long before the reaper
sees it (`claimed_at IS NULL` then excludes it permanently). Observed both ways
on the same site, same day:

- **Sequential re-queue → survives.** `needs_page:capabilities` (created
  2026-07-20 21:27) re-queued 2026-07-25 ~08:09, claimed 08:11, completed 08:17.
- **Batch re-queue → parked.** Four pages queued together: the first is claimed,
  the rest wait behind the in-flight build [INFERRED: single-flight per site,
  downstream of the trigger — the trigger's pre-query only filters
  `sites.locked_at IS NULL`; I did not trace where the serialisation lives, see
  `bugs_open/029`/`030`]. A page build takes tens of minutes, so siblings stay
  unclaimed past the reaper's hour and are parked. Twice — hence the doubled
  prefix.

So the defect is a **race the re-queuer cannot win for a batch**, not a
deterministic kill.

## Fix candidates

**Candidate 1 (recommended, minimal): key on `updated_at`.**

```sql
  AND updated_at < NOW() - INTERVAL '48 hours'
```

An item sitting untouched in `triaged` has `updated_at` = the moment it was set
`triaged`, which *is* the stated intent. No schema change, no image roll.
**Risk to weigh before applying:** anything that touches the row for unrelated
reasons bumps `updated_at` and postpones reaping, so a genuinely wedged item
that is periodically written to would never be parked — the queue-deadlock this
task exists to prevent. Whether that happens in practice is measurable before
changing anything:

```sql
-- do triaged build items get touched after entering triaged?
SELECT count(*) FILTER (WHERE updated_at > created_at + INTERVAL '1 minute') AS touched,
       count(*) AS total
FROM site_work_items WHERE status='triaged' AND pipeline='build';
```

**Candidate 2 (durable): a `triaged_at` column,** set whenever status becomes
`triaged`, reaped on that. Precise and immune to unrelated writes, but needs a
migration plus every writer that sets `triaged` to maintain it — and there are
several. Worth it only if candidate 1's measurement shows real churn.

**Candidate 3 (independent of both, and worth doing anyway): stop mutating the
summary.** Prefixing in place is lossy and cumulative. Record the reaping in a
dedicated column or a note, and leave the human-readable summary alone. The
present behaviour makes an already-confusing row unreadable, and there is no way
back to the original text.

## Verification (for whoever fixes it)

Induce the fault, don't just watch a happy path (`[[verify-the-failing-branch]]`):

1. Pick a build item with `created_at` older than 48h and `claimed_at IS NULL`.
2. Set `status='triaged'`, and **disable `build-pipeline-trigger`** for the test
   so the claim cannot win the race and mask the result.
3. Wait for one reaper tick (≤1h) — pre-fix it is parked, post-fix it is not.
4. Re-enable the trigger. Confirm the genuine case still works: an item whose
   `updated_at` is 48h+ old and unclaimed is still parked.

Do not test by re-queueing and watching it succeed — that only proves the
trigger won the race.

## Notes

- `bugs_closed/048` concerns the **same scheduled task** (its no-op pre-query
  starving the `maintenance` concurrency group) but a different defect. Read it
  before editing the row; the fix there established that this task's pre-query
  must return no rows when there is nothing to do.
- The three parked rows on fundamentallyai.com are evidence — leave them until
  the fix is verified, then clean up the corrupted summaries by hand (there is
  no automated way back).

## Reproduced live, unprompted, on the failing branch (2026-07-25 15:32 UTC)

Recorded the same day the bug was filed, from the same site's queue — this is the
induced-fault evidence the Verification section asks for, obtained by accident:

| what | when (UTC) |
|---|---|
| row `dd50beb1` (`needs_page:self-correction-leopardessconsulting`) created | 2026-07-20 21:27 |
| re-queued to `triaged` by hand | 2026-07-25 ~15:20 |
| reaper parked it `unresolved`, prefix `[stale: triaged 48h+]` | 2026-07-25 **15:32** |

**Twelve minutes** between entering `triaged` and being labelled "triaged 48h+".
`build-pipeline-trigger` did not get to it first because the site had another
build in flight. Nothing about this run was set up to test the reaper.

**The workaround from Fix candidate 3's neighbourhood works and is legal.**
Instead of resurrecting the historic row, INSERT a new one:

- `unresolved` IS in `idx_swi_dedup`'s terminal-status exclusion list
  (`complete, verified, rejected, wont_fix, failed, unresolved, cancelled`), so a
  second row with the **same `item_key`** inserts cleanly beside the parked one.
  No constraint is being worked around.
- The fresh `created_at` puts it outside the reaper's predicate entirely, and the
  summary can describe the rebuild actually requested rather than the 2026-07-20
  first build.
- **Landmine:** `site_work_items.created_by` is NOT NULL with no default, so a
  `INSERT … SELECT` copy of an existing row must name it explicitly or fail with
  23502.
- Caveat repeated from §Fix candidates: a hand-written INSERT bypasses the
  Go-side two-strike suppression in `insertWorkItem`. Deliberate here; know that
  you are doing it.

This makes candidate 1 (`updated_at`) less urgent for operators — there is a
working path — but no less correct: the label the reaper writes is false, and it
writes it onto rows that are actively being worked.
