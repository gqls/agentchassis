# CLOSED 2026-07-26 — Handoff — `site_work_items.updated_at` is not maintained, so "did this run?" reads as "no"

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
> **SUPERSEDED 2026-07-26 — fixed and live, see "RESOLVED" below.** The table below is kept as
> history (and remains the right method for **pre-2026-07-26 rows**, whose `updated_at` was
> deliberately not backfilled). For rows updated after the trigger went live, `updated_at` is now
> load-bearing and can be read as it is named.

Do not use `updated_at` on this table as evidence of anything. To establish whether a dispatched
item ran, check the artefact it was supposed to produce:

| item type | the artefact that proves it ran |
|---|---|
| `needs_site_plan` | a new `site_plans` row for the site, `is_current = true` |
| `needs_page` | `page_components` rows for the page with a fresh `created_at` |
| any | `completed_at`, and the matching `orchestration_states` row |

This is the same rule as CLAUDE.md § Debugging, "Trust the rendered artefact, not the status" —
here the *timestamp* is the thing lying rather than the status.

## RESOLVED — FIXED & LIVE 2026-07-26 (fix candidate 1, the trigger)

Live via **migration `216_site_work_items_touch_updated_at.sql`** (ledger-recorded). DB-only, so it
is live immediately — **no image roll needed**, no Go change, nothing inert.

```sql
CREATE TRIGGER trg_site_work_items_updated_at
BEFORE UPDATE ON site_work_items FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

**Reused `public.set_updated_at()` rather than creating `touch_updated_at()` as this file sketched.**
The function already existed and already backed `trg_<table>_updated_at` triggers on five tables
(`site_specs`, `layouts`, `site_plans`, `content_feed_items`, `model_lifecycle.training_runs`), so
this table now follows the established fleet convention instead of adding a second name for one
behaviour:

```sql
SELECT tgname, tgrelid::regclass FROM pg_trigger
WHERE tgfoid = 'set_updated_at'::regproc AND NOT tgisinternal;
-- trg_site_work_items_updated_at | site_work_items   <-- new, 6th
```

### Re-grounded before fixing (the filed figures had gone stale)

`site_work_items` is pruned, so the filed "4,156 of 4,643" no longer existed. Measured 2026-07-26,
same query: **134 of 941 complete rows had `updated_at > created_at` — 86% still wrong**, and still
zero triggers. The defect was unchanged; only the volume moved.

Also changed since filing: this file says "exactly one Go writer sets it, in an unrelated path". By
07-26 **six** paths set `updated_at = NOW()` explicitly (`claim_work_item_action.go:152,189`,
`complete_work_item_verification.go:250`, `load_work_item_actions.go:921`,
`plan_sections_action.go:1729`, `v3_site_actions.go:4524,4532`, `diagnose_*` actions,
`confirm_work_item_handler.go:221`). That is not a counter-argument to the trigger — it *is* the
"several writers must all remember" drift this file predicted, accumulating in real time. The
explicit writers coexist harmlessly (the trigger overwrites with the same transaction `NOW()`).

### Verified live — on the failing branch, not a happy path

The two **primary** status-transition writers set no `updated_at` and are the ones that were lying:

| path | statement | sets `updated_at`? |
|---|---|---|
| `claim_work_item_action.go:97` | `SET status='claimed', claimed_by, claimed_at` | **no** |
| `load_work_item_actions.go:802` (`CompleteWorkItemAction`) | `SET status='complete', result, completed_at, handled_by` | **no** |

An organic item **spanning the fix** proves it — no synthetic test, and it went through
`CompleteWorkItemAction` (`handled_by='build-dispatch-loop'`):

```
item_type    | page_rerender
created_at   | 2026-07-26 13:49:04.643017+00
claimed_at   | 2026-07-26 13:53:06.410941+00   <-- claim UPDATE, PRE-trigger: updated_at did not move
completed_at | 2026-07-26 13:53:46.338133+00   <-- completion UPDATE, POST-trigger (live 13:53:33)
updated_at   | 2026-07-26 13:53:46.338133+00   <-- moved, == completed_at to the microsecond
```

Before the fix this row would have read `updated_at = 13:49:04` — the exact misreading this case was
filed about.

Verification items 2 and 3 from "How to verify a fix" above:
- **Sweeper safety**: no reader of `site_work_items.updated_at` exists anywhere — not in Go
  (`platform/ internal/ pkg/ cmd/` hold only `SET` writers), not in the admin dashboard `.tsx`, not
  in `scripts/`, not in `scheduled_tasks.pre_query` or `agent_definitions`. The three sweepers whose
  config matches an `updated_at` grep read it on **other** tables (`sites`, `page_components`,
  `site_components`); `stale-work-item-reaper` filters on `created_at` and already *wrote*
  `updated_at` itself. No index involves the column (`idx_swi_dedup` keys `(site_id, item_key)`).
  Confirmed by behaviour too — both work-item sweepers completed cleanly on their first
  post-trigger tick (trigger live `13:53:33`; `last_completed_at == last_triggered_at` in each case):
  `build-pipeline-trigger` at `13:55:06`, `claimed-item-timeout` at `13:56:17`. `claimed-item-timeout`
  matters most here: its `completed_by_evidence` and `reset` branches both write
  `site_work_items` without setting `updated_at`, so they are newly trigger-covered, and its
  evidence checks read `updated_at` on `page_components`/`site_components` — different tables,
  unaffected. `stale-work-item-reaper` runs hourly, so its next tick falls after this close; it is
  covered by inspection (filters on `created_at`, already wrote `updated_at` itself) rather than
  by observation. [UNMEASURED — the one check not yet observed live]
- **No backfill**, as this file directs. The 134-of-941 historical figures above are unchanged by the
  fix; `completed_at` still carries the truth for old rows.

### Two traps worth keeping

1. **`now()` is fixed for a whole transaction.** The migration's guard live-fires the trigger on a
   probe row, and a probe inserted *at* `now()` would show `updated_at == created_at` even with the
   trigger working perfectly — both defaults and the trigger draw the same transaction timestamp. The
   probe is therefore **backdated an hour**, so the trigger's `now()` is the only thing that can move
   the column. Anyone testing a `touch_updated_at` trigger in one transaction will hit this and
   conclude, wrongly, that their trigger is dead.
2. The guard deliberately does **not** no-op-write a real historical row to test itself — that would
   stamp today's timestamp onto it, which is the same fabrication the no-backfill rule forbids. It
   inserts and deletes its own probe inside the transaction, so the row is never visible to another
   session and never survives the file.

**Number ambiguity:** a bare "035" in `016b` is ambiguous — `016b §9`'s "the 035 §1.5 bool trap" and
"the 003/035-envelope family" both refer to **`035_adapter_guide.md`**, not this bug. Resolve by slug.

## Related

- `/bugs_open/030` — dispatch-queue backlog; makes the wrong reading ("it never ran") plausible.
  (Itself closed 2026-07-26.)
- `/bugs_closed/001` — where this was hit; its "VERIFIED LIVE" section records the trap.
- `016b §9` — "A queued orchestration is indistinguishable from a dropped one", the same class of
  expensive misreading.
