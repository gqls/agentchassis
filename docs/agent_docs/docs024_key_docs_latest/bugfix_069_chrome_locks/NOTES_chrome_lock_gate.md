# NOTES — chrome lock gate (`bugs_open/069`) — append-only, newest at the bottom

## 2026-07-26 — session open, evidence gathered before touching code

**Ownership.** `who-owns.py 069` → "OWNED or recently active", which is a false positive here: the
only commit touching the file is `2c7fd3be9`, the 058 record that *filed* 069. 058 closed at
`3f66627bf` this morning saying 069 is a separate bug. No open `site_work_items` row mentions chrome
locks. Nobody is on it.

**Schema, checked live, not from the repo's `.sql` copies:**

```
\d site_components
  locked_at | locked_by | lock_type | lock_expires_at
  "chk_site_components_lock_type" CHECK (lock_type IS NULL OR lock_type IN ('permanent','timed','review'))
  "site_components_site_id_slot_name_key" UNIQUE (site_id, slot_name)
```

Identical to `page_components`. So `pageComponentAgentWritableSQL` is schema-valid here unchanged.

**Writers (`grep -rn "UPDATE site_components\|INSERT INTO site_components\|DELETE FROM site_components" --include=*.go`):**

```
fix_component_template_action.go:276,452,528,598
link_site_components_action.go:145
render_site_components_action.go:532,671
internal/core-manager/admin/page_admin_handlers.go:1116,1205,1232   (the human surface — exempt)
```

Plus, from the DB side: `revert_site_to_snapshot` is the only function that writes the table
(`pg_get_functiondef` over `pg_proc` — `take_site_snapshot` and `get_page_component` only read).

**Live counts:** 42 chrome rows, 0 locked. 931 page components, 39 locked, all `lock_type`-stamped.
0 `lock_blocked_change` items since 058 shipped. 0 admin chrome edits ever. The bug is real and
ungated but has never actually fired — so the proof has to induce it.

**The design question the handoff asks** ("confirm no chrome flow treats the skip as a hard
failure") — answered from the live DB rather than by reading Go:

```sql
SELECT count(*) FROM agent_definitions WHERE is_active AND NOT COALESCE(is_snapshot,false)
  AND deleted_at IS NULL AND default_config::text LIKE '%render_site_components.rendered%';
-- 0
```

Six live agents run the action; none reads its `rendered` map, and the action returns
`success: true` regardless. The only nearby conditional (`rerender-pages.check_refresh_components`)
decides whether it runs at all. Four of the six force; two do not.

### Misstep 1 — I nearly put the lock pre-check at the top of the function

The obvious placement is the first line of `renderAndStoreSiteComponent`. That is wrong: lines
474-489 are an idempotence gate that returns `true` **without writing** when `force == false` and the
slot already has HTML. A check above it would file a `lock_blocked_change` item — whose text claims
"an automated writer wanted to change this and was blocked" — for a call that was never going to
write, on every ordinary build of every site with a locked slot. The item would be a lie. Caught in
review before it was written; the check goes between 489 and 494, which also happens to cover the
`component_id IS NULL` fallback branch (the worst case) for free.

### Misstep 2 — I repeated the handoff's "birth-only" claim in my own first draft

`bugs_open/069` says the render action's INSERT "is birth-only", so I initially scoped it out. It is
`ON CONFLICT (site_id, slot_name) DO UPDATE SET component_id = $3` — a mutation, and the one that
repoints a locked slot at a **generic default template** picked by function name. Caught by reading
the statement instead of the sentence about it. Cheap check that would have caught it immediately:
read the SQL, not the prose describing the SQL — the same class as 016b's "grep the write side".

### Finding — the admin flow is self-defeating today

`HandleUpdateSiteComponent` (`page_admin_handlers.go:1084-1119`) defaults `shouldLock = true` (⇒
`LockPolicyFor("admin")` ⇒ permanent) **and** files a `needs_rerender` item with
`refresh_site_components: true`, handled by `rerender-pages`, which runs
`render_site_components` at `force_rerender: true`. So the admin's own follow-up re-render overwrites
the edit it just locked. `[MEASURED]` it has never happened — 0 such items exist — so this is a
reachable path, not an observed failure, and it must be written that way.

### Finding — a snapshot revert destroys the locks themselves (→ `bugs_open/085`)

Read from the live function bodies, because the repo's `.sql` copies are stale:

- `take_site_snapshot` captures chrome `locked_at, locked_by` but **not** `lock_type` /
  `lock_expires_at`, and captures **no** lock columns for `page_components`.
- `revert_site_to_snapshot` deletes and re-inserts both tables with no lock columns in either
  INSERT — so every lock becomes NULL, including the two that were captured. The `pre_revert` safety
  snapshot has the same hole, so it cannot restore what it did not record.

39 live locked page components are exposed. A subagent's audit reported this from the repo's `.sql`
files and got the capture set wrong (it said no lock columns are captured at all); the live
definition disagrees. Recorded because the correction is the useful part: **read `pg_get_functiondef`,
not `docs/agent_docs/sql_for_tables/*.sql` — the files have drifted.**

## 2026-07-26 evening — shipped, rolled, proven; and three things that went wrong on the way

**Shipped:** `05bcb3586` (the gate, 7 files), `d9e7ef7cb` (the lockstep test), `cc42407c1` (the tag
bump), plus `65864966a` for the snapshot half (`088` + migration 219).

### Misstep 3 — I built the image, and the tag had already moved twice under me

Planned to build `v1.0.1168`. By the time I got there the makefile said `v1.0.1169`, because another
thread had built 1168 *and* 1169 in the interim — and 1169 was built from `499a08398`, the **parent**
of my commit, so it did not contain the gate. Built `v1.0.1170` instead.
**The check that catches this:** `git log --oneline -1 -- makefile` and `docker images` before
assuming the tag in front of you is yours. A tag that exists is not a tag that contains your change.

### Misstep 4 — `HEAD~1` stopped meaning "before my fix" while I was using it

Extracted the council submission's `grounded_in` quotes with `git show HEAD~1:<file>`. Another
session committed between the extraction and the check, so `HEAD~1` silently became **my own
commit** and the quotes would have shown the POST-fix code as evidence of the pre-fix defect. Caught
by re-reading one quote and noticing it lacked the line I had added. Re-extracted against the
explicit SHA `499a08398`.
**The check:** in a shared tree, never cite `HEAD~1` — resolve `git rev-parse <mycommit>^` once and
use that literal SHA in the citation itself.

### Misstep 5 — the first live probe was eaten by another thread's roll

Fired the probe at 18:41 and it sat queued behind ~12 messages (one council-gate run was 47 minutes
in). While it waited, another session rolled the chassis to `v1.0.1171` at 21:02 — and the queued
message was gone: no `orchestration_states` row, ever. The `dispatch-queue-depth.sh` verdict
("QUEUED, not lost — do NOT re-fire") was *correct advice at the time* and still ended in a lost
message, because it cannot see a roll that happens later.
**The check, which is the exit test `bugfix_052` already wrote down:** "absence means queued" needs
*"has anything newer drained past me"* — and now also **"has the consumer restarted since I
published"**. Re-fired against 1171 and it ran in under 5 seconds.
Secondary note: the re-fire went out at 21:07:09, which was **253s** after the pod started — inside
the ~300s window where a dispatch is meant to be dropped. It ran anyway. So that landmine is about
the *spawn* path, not the generic orchestrate lane; do not treat the 300s rule as covering both.

### The proof

Deployment first, then the fault. Pod-grep on `agent-chassis-5b4456686c-s5fkc` (v1.0.1171) found all
five literals the fix creates, a positive control, and a zero for a negative control. Then three
scratch chrome slots on dartsonline (none of them a real slot) and a forced re-render:

- locked, component-backed → md5 **and** `updated_at` unchanged;
- unlocked sibling → rewritten, 41 → 3429 bytes;
- locked with `component_id IS NULL` → **still NULL**, so the generic-default fallback no longer
  repoints a locked slot (the case 058 could not have);
- `locked_slots_preserved: ["probe-069-locked","site-footer"]`, and exactly two
  `lock_blocked_change` items at `needs_human_review` with no handler and
  `surface=site_component`.

Cleanup: items first, then fixtures; leak check 0 on every line, dartsonline's 3 real chrome rows
untouched.

**What the probe did NOT prove** (recorded because the omission is invisible otherwise): both items
came out `severity=medium`, since the `high` branch keys on the literal slot names `header`/`footer`
and the probe used scratch names deliberately. `link_site_components` and the four
`fix_component_template` paths were not driven live either. Those rest on the shared predicate plus
their sqlmock cases.

### Council round 1 — REVISE, and why that is not a rejection

Corr `75dff4cd-e822-4b88-bd98-d989ef32bc90`. `decided_by: "unreadable reviewer(s):
review_editquality.result"` — the verdict was decided by a seat that failed to produce a readable
result, not by an objection. 10 of 12 approved; two mediums, no veto. Read `decided_by` **before**
reading the decision, exactly as the 029 thread found.

The one objection that earned code: debug_historian, that the hand-copied predicate in
`discovery_checks` shipped with nothing pinning it to the shared helper. Now
`TestDiscoveryChromeLockFilterMatchesSharedPredicate`, and I checked it can actually fail — flipping
`'timed'` to `'review'` in the detector made it fail, restoring made it pass. A lockstep test nobody
has ever seen fail is a claim, not a guard.
