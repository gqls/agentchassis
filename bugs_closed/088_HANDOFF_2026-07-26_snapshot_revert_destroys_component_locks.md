# Handoff — a snapshot revert silently destroys every component lock on the site

**Filed 2026-07-26**, found by the writer audit for `/bugs_open/069` (the chrome half of the
component lock gate). It is **not** a chrome defect: it hits `page_components` the same way, so it
belongs to neither `058` nor `069`, and it is fixed in the database rather than in Go.

> **Numbering note:** first drafted as `085`; `085`, `086` and `087` were taken by other sessions
> between the audit and the filing. This file is `088`.

## The defect

`revert_site_to_snapshot(p_snapshot_id, p_reverted_by)` restores a site by deleting and re-inserting
its content. Read from the **live** function body (`pg_get_functiondef`, 2026-07-26 — the repo's
`.sql` copies have drifted and disagree):

```sql
    -- ── 4. Restore site_components ─────────────────────────────────────
    DELETE FROM site_components WHERE site_id = v_snap.site_id;
    FOR v_sc IN SELECT * FROM jsonb_array_elements(v_snap.components_snapshot)
    LOOP
        INSERT INTO site_components (
            id, site_id, slot_name, component_id,
            rendered_html, content_data, build_status
        ) VALUES ( … );
    END LOOP;
```

**No lock column appears in that INSERT** — nor in the `page_components` one at section 2. So every
`locked_at / locked_by / lock_type / lock_expires_at` on the site becomes NULL. The capture side is
partial in a way that makes it worse, not better:

| table | captured by `take_site_snapshot` | restored by `revert_site_to_snapshot` |
|---|---|---|
| `site_components` | `locked_at`, `locked_by` (**not** `lock_type`, `lock_expires_at`) | none |
| `page_components` | none | none |

The `pre_revert` safety snapshot cannot rescue the locks, because it is taken by the same
`take_site_snapshot` and has the same hole.

## Why it matters

The lock columns are the **only** thing standing between a human-corrected artefact and automation.
`bugs_closed/058` (live v1.0.1165) and `bugs_open/069` both enforce exactly these four columns; a
revert quietly returns every protected row to "agent-writable" and nothing records that it happened.
The next ordinary rebuild then overwrites copy a human had locked, and the audit trail says the
rebuild was entitled to.

**Exposure, measured 2026-07-26:** 39 locked `page_components` rows live (all `lock_type`-stamped),
0 locked `site_components` rows. 11 snapshots exist; the most recent was taken 2026-06-24, so reverts
are rare — this is a live-but-cold defect, not a burning one.

Reachable from two live entry points, neither of which checks locks:

- `POST /admin/sites/:site_id/snapshots/:snapshot_id/revert` → `HandleRevertSnapshot`
  (`internal/core-manager/admin/snapshot_admin_handlers.go:282`)
- the registered `revert_site_snapshot` action
  (`platform/orchestration/actions/site_snapshots_actions.go:186`) — **no live `agent_definitions`
  row wires this action today** (checked), so in practice only a human can trigger it.

## The fix (chosen semantics)

**A revert restores content; it never locks or unlocks anything.**

- **Capture** all four lock columns for both tables, so a snapshot is a true record and the
  `pre_revert` safety copy is worth having.
- **Restore** by *preserving the current lock state* across the delete+insert, not by replaying the
  snapshot's. Replaying as-captured would silently RELEASE a lock added after the snapshot was
  taken — the very defect class being fixed here — and the approved lock policy
  (`031_LOCKS_should_locks_expire.md`, and 058's ruling) is that only a human unlock releases a lock.
  The content still comes from the snapshot: that is what the human asked for, and a revert is a
  human-initiated surface, exempt in the same way as the admin endpoints.

Keys used to carry the lock forward: `slot_name` for chrome (unique per site), and
`(page_id, slot_name)` for page components — `pages.id` survives the revert, so the key holds. Where
a page carries duplicate slot names, every matching row is re-locked: conservative, locks more rather
than fewer.

## How to verify

1. On a scratch site: lock one chrome slot and one page component; `SELECT take_site_snapshot(...)`;
   assert all four lock columns are present for both tables in the snapshot JSON.
2. Add a **second** lock *after* the snapshot (this is the case that fails under
   restore-as-captured).
3. `SELECT revert_site_to_snapshot(<snapshot>, 'verify')`.
4. Assert: both locks still present and still active; content restored from the snapshot; the
   `pre_revert` safety snapshot carries the lock columns too.
5. Leak-check the fixtures to zero.

## Related

- `/bugs_closed/058` — the page_components write-side lock gate this defect silently disarms.
- `/bugs_open/069` — the chrome half; this was found by its writer audit.
- `docs/agent_docs/sql_for_agents/115_locks.sql` — the migration that added `lock_type` /
  `lock_expires_at` to four tables (`page_components`, `site_components`, `site_plan_directives`,
  `assets`). **Residual worth a look, not audited here:** `site_plan_directives` and `assets` carry
  the same columns; nothing in this file says whether their writers honour them.

---

## FIXED & LIVE 2026-07-26 — migration `219_snapshot_revert_preserves_component_locks.sql`

DB config, so it is live the moment it is applied — no image roll, nothing inert.

**What shipped.** `CREATE OR REPLACE` of both functions, built by transforming the **live** bodies
(`pg_get_functiondef`) rather than retyping them, each substitution asserted to match exactly once:

- `take_site_snapshot` now captures `locked_at, locked_by, lock_type, lock_expires_at` for
  **both** `page_components` (which had none) and `site_components` (which had the first two).
- `revert_site_to_snapshot` reads the site's **current** lock state into a jsonb variable before each
  DELETE and re-applies it after the inserts — chrome keyed on `slot_name`, page components on
  `(page_id, slot_name)`, since `pages.id` survives a revert. The result JSON now reports
  `page_locks_preserved` / `chrome_locks_preserved`.
- The file's own `DO $guard$` block asserts the post-conditions inside the same transaction, so a
  half-applied file rolls itself back.

**Applied by hand** (`psql -f`), then `run-migrations.sh --record-only`: the runner's `--apply` runs
every pending file in order, and eight other threads' migrations were pending.

**Why preserve-current rather than restore-as-captured** — the decision, not a detail. Replaying the
snapshot's lock state would silently release any lock a human added *after* the snapshot was taken,
which is the same defect wearing different clothes. One rule instead: **a revert restores content; it
never locks or unlocks anything.** Content still comes from the snapshot, because that is what the
human asked for.

### How it was verified — induced fault, against the live functions

The whole test ran inside one transaction ending in `ROLLBACK`, so the real deployed functions were
exercised against the real schema and **no fixture was ever committed** (leak check afterwards: 0
rows on every line, and the 39 real locked `page_components` rows untouched). Every assertion RAISEs,
so silence would have been failure.

1. **Control first** — the oldest existing snapshot's chrome entry has keys
   `build_status, component_id, content_data, id, locked_at, locked_by, rendered_html, slot_name`
   and **no `lock_type`**. Without this the test would pass on a snapshot taken before the fix.
2. Fixture: a scratch site, one page with two components, two chrome slots; one row per table locked
   `permanent` **before** the snapshot.
3. After `take_site_snapshot`: all four lock columns present in the chrome capture *and* in the
   page-components capture, with the live lock recorded. → PASS
4. Then the discriminating step: a **second** lock added *after* the snapshot (page: permanent;
   chrome: `timed` +30d), plus content drift on every row.
5. `revert_site_to_snapshot` returned
   `"page_locks_preserved": 2, "chrome_locks_preserved": 2`, and afterwards: content restored to the
   snapshot's copy on both tables; the pre-snapshot locks intact; **the post-snapshot locks intact
   too, with the timed lock's type and expiry preserved**; exactly two locked page rows, so nothing
   was locked that was not locked before. → PASS
6. The auto `pre_revert` safety snapshot carries the lock columns as well, so the safety net is now a
   true record. → PASS

**Residual, named rather than left silent:** `115_locks.sql` added the same four columns to
`site_plan_directives` and `assets`. Nothing here audits whether *their* writers honour them.
