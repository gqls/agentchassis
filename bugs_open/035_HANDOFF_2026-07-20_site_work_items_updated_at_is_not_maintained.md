# Handoff — `site_work_items.updated_at` is not maintained, so "did this run?" reads as "no"

**Filed 2026-07-20**, hit while verifying `/bugs_open/001` live. Small, but it costs time on exactly
the occasions when someone is trying to establish whether a dispatched job actually did anything —
and it makes a completed job look like a dropped one.

## The defect

`site_work_items` has an `updated_at` column that almost nothing writes. A work item can be created,
claimed, processed and completed, and `updated_at` will still equal `created_at`.

```sql
SELECT count(*) AS complete_rows,
       count(completed_at) AS has_completed_at,
       count(*) FILTER (WHERE updated_at > created_at) AS updated_at_moved
FROM site_work_items WHERE status='complete';
--  4643 | 4622 | 487
```

**4,156 of 4,643 completed items (89.5%) never had `updated_at` touched.** There is no trigger:

```sql
SELECT tgname FROM pg_trigger
WHERE tgrelid='site_work_items'::regclass AND NOT tgisinternal;  -- 0 rows
```

And exactly one Go writer sets it, in an unrelated path
(`diagnose_silent_check_action.go:423`). The status-transition writers do not.

## Why it bites

The natural "did my dispatch do anything?" query is `SELECT status, updated_at FROM site_work_items
WHERE id = …`, and it returns something that looks decisive and is not:

```
  status  |          updated_at
----------+-------------------------------
 complete | 2026-07-20 09:39:38.118429+00   <-- identical to created_at
```

On 2026-07-20 that read as *"the item was never picked up; it has been sitting there since I
inserted it"*. The re-plan had in fact run to completion and written a new `site_plans` row six
minutes earlier. The misreading is especially easy because `/bugs_open/030` (dispatch-queue backlog)
makes "nothing has happened yet" a genuinely common state, so the wrong conclusion is plausible —
and the documented remedy for a dropped dispatch is to resubmit, which is a credit spend
(`016b §9`, "A queued orchestration is indistinguishable from a dropped one").

`completed_at` IS reliably written (4,622 of 4,643) — so the information exists, just not where the
column name promises it.

## Fix candidates

1. **A trigger, so it cannot drift again** (preferred — this is one statement and covers every
   writer, present and future):
   ```sql
   CREATE OR REPLACE FUNCTION touch_updated_at() RETURNS trigger AS $$
   BEGIN NEW.updated_at = now(); RETURN NEW; END $$ LANGUAGE plpgsql;
   CREATE TRIGGER site_work_items_touch_updated_at
     BEFORE UPDATE ON site_work_items FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
   ```
   Check first whether such a function already exists for another table, and whether any code
   *relies* on `updated_at` being immutable (grep for `updated_at` in work-item queries — the
   dispatch/claim paths and any dedup or age-based sweep are what to look at, since a suddenly-moving
   column could change which rows a sweeper considers stale).
2. **Set it in the status-transition writers.** More faithful to how the rest of the platform writes
   (`UPDATE … SET status = $1, updated_at = NOW()`), but it is the same "several writers must all
   remember" shape that this platform keeps getting bitten by, and a new writer will forget.
3. **Do neither, and remove the column** — if `completed_at` plus the status is genuinely all
   anyone needs, a column that lies is worse than no column. Requires checking no dashboard or
   report reads it.

Prefer 1. This is a small enough surface that 2 is defensible, but the column has already drifted to
89.5% wrong, which is the argument against relying on discipline.

## How to verify a fix

1. Insert a work item, let it be claimed and completed, assert `updated_at > created_at`.
2. Assert the claim/dispatch path still selects the same rows it did before (a moving `updated_at`
   must not change which items a sweeper treats as stale).
3. Backfill is **not** needed and should not be attempted — `completed_at` already carries the truth
   for historical rows, and inventing an `updated_at` for 4,156 rows would fabricate timestamps.

## Until it is fixed

Do not use `updated_at` on this table as evidence of anything. To establish whether a dispatched
item ran, check the artefact it was supposed to produce:

| item type | the artefact that proves it ran |
|---|---|
| `needs_site_plan` | a new `site_plans` row for the site, `is_current = true` |
| `needs_page` | `page_components` rows for the page with a fresh `created_at` |
| any | `completed_at`, and the matching `orchestration_states` row |

This is the same rule as CLAUDE.md § Debugging, "Trust the rendered artefact, not the status" —
here the *timestamp* is the thing lying rather than the status.

## Related

- `/bugs_open/030` — dispatch-queue backlog; makes the wrong reading ("it never ran") plausible.
- `/bugs_open/001` — where this was hit; its "VERIFIED LIVE" section records the trap.
- `016b §9` — "A queued orchestration is indistinguishable from a dropped one", the same class of
  expensive misreading.
