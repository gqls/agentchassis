# PLAN 2026-08-16 — flag-only work items are promoted into the dispatch queue and stamped `blocked`

Lane opened 2026-08-16 by the `bugs_open/284` session. 284 was filed 2026-08-15 by
the `279` lane, which closed out explicitly saying the bug was **unowned and needed
a fresh session** (its own words, transcript 2026-08-16 16:5x). No competing work:
`scripts/who-owns.py 284` names no owning workstream, and a grep of every session
transcript touched in the last five hours found no other session on `capability_gap`
or on this file.

## What the bug is, in plain terms

The platform files two kinds of finding. Most are **jobs**: something is wrong and a
named agent can fix it, so the row carries a `handler_agent` and the dispatch loop
routes it. Some are **flags**: something is wrong and *nothing on the platform can
fix it* — a brand's palette is unreadable, a VM is down, a page references an image
that no asset deploys. A flag row deliberately names no handler; it is there for a
human to read.

The producers of flag rows believed that giving a row no handler was enough to keep
it out of the dispatch queue. It is not. A separate step promotes **every** row on a
site into the queue without looking at the handler, the dispatch loop then picks the
flag row up, discovers it has nobody to send it to, and stamps it `blocked` with the
error *"No handler_agent set — item cannot be routed to any agent"*.

So a row that says "a human should look at this" is rewritten to say "the machinery
tried to route this and failed". That is the defect: **a correct finding is recorded
as a failure**, and — because `blocked` still holds the row's de-duplication slot —
the check that found it can never file it again.

## The chain, each link read first-hand 2026-08-16

1. **Producers** build a `WorkItemSpec` with an empty `HandlerAgent` and
   `Status: "detected"`:
   - `discovery_checks/check_palette_contrast.go:120-132` — `capability_gap`
   - `discovery_checks/check_content_duplication.go:232-248` — `capability_gap`
   - `discovery_checks/check_image_url_404.go:256-278, 297-310, 330-348` —
     `image_url_404`; **omits the field entirely**, so Go zero-values it to `""`.
     A grep for `HandlerAgent: ""` cannot see this site.
   - `discovery_checks/check_site_unreachable.go:254-264`,
     `check_backend_unreachable.go:99-108` — alert-only, self-clearing.

   Each carries a comment asserting the row is deliberately non-dispatchable. The
   correct idiom is one file away: `discovery_checks/remit.go:163-207`
   (`CapabilityGapItem`) uses **`Status: "deferred"`** with the empty handler, and
   says why in the spec itself.

2. **Promoter.** `triage_detect_items_action.go:161-173` — `UPDATE site_work_items
   SET status='triaged', triaged_at=now(), spec=jsonb_set(…,'{original_pipeline}',…),
   pipeline=$2 WHERE site_id=$1 AND status='detected'`. No item_type filter, no
   handler filter. Live inside `improvement-loop` (`SingleOwner: true`).

3. **Loader + claim.** `load_work_item_actions.go` selects `status IN
   ('triaged','approved')`; `claim_work_item_action.go:96-105` claims on the same
   pair, then :159-180 reads `handler_agent`, finds it empty, sets `status='blocked'`
   and NULLs `claimed_by`/`claimed_at`.

4. **No way back.** `feasibility-recheck` (live scheduled task) un-blocks only
   `WHERE EXISTS (SELECT 1 FROM agent_definitions ad WHERE ad.type = wi.handler_agent …)`
   — no agent type is the empty string, so these rows are stuck for ever. `blocked`
   is in neither `workItemTerminalStatuses` (`work_items_common.go:40-55`) nor
   `idx_swi_dedup`'s excluded list, so each stuck row **holds the dedup slot** and the
   re-run finding is dropped.

5. **The platform already has the right predicate — in the OTHER promoter.** The live
   `detected-item-promoter` scheduled task promotes `detected`→`triaged` only where
   `COALESCE(handler_agent,'') <> ''` AND the handler is a live active agent row AND
   that `(item_type, handler_agent)` pair has ≥1 lifetime `complete`. One promoter is
   handler-aware; the Go one is not.

## `[MEASURED]` 2026-08-16 — the class is three times the size 284 recorded

284 counted 18 `capability_gap` rows. The same error string, fleet-wide:

```sql
SELECT item_type, status, left(error,55), count(*) FROM site_work_items
WHERE status='blocked' GROUP BY 1,2,3 ORDER BY 4 DESC;
```

| item_type | rows | sites | error |
|---|---|---|---|
| `image_url_404` | 40 | 15 | No handler_agent set |
| `capability_gap` | 18 | 14 | No handler_agent set |
| `needs_experience_plan` | 1 | 1 | No handler_agent set |
| `page_rerender` | 1 | 1 | No handler_agent set |
| `needs_human_review` | 5 | — | *Handler agent not registered: hitl-review* (different branch, out of scope) |

**60 rows, on at least 15 sites.** A further **37** rows sit at `detected` with an
empty handler today (`head_essentials_missing` 36 on one site, `image_url_404` 1) and
will be blocked the next time the improvement loop triages those sites.

The decisive join, which could have come out otherwise and did not: all 18 blocked
`capability_gap` rows carry `spec.original_pipeline` (written **only** by
`TriageDetectedItemsAction`) and a non-null `triaged_at`; all 19 correctly-`deferred`
rows carry neither, and 18 of those carry `spec.not_dispatchable` (written **only** by
`CapabilityGapItem`). Every blocked row has `claimed_by IS NULL` and
`attempt_count = 0` — exactly what the block branch's own NULL-ing leaves behind.

## Design position (fix candidates ranked by what they make unrepresentable)

Ordered by CLAUDE.md's rule — rank by what closes the door, not by what is quickest.

1. **Promoter guard (primary).** `TriageDetectedItemsAction` must not promote a row
   the claim path will refuse. The two are the same predicate one step apart; moving
   it earlier makes `blocked`-for-no-handler structurally unreachable from this path,
   whatever a producer does now or later. Open question the plan must answer, not
   assume: the scheduled promoter also demands a lifetime-`complete` pair, which
   inside the improvement loop would strand a brand-new handler's first item. The
   candidate predicate for THIS promoter is therefore the claim path's own two
   conditions and nothing more.
2. **Shared write door.** `writeWorkItem`/`insertWorkItem`
   (`load_work_item_actions.go:1316-1323`, ~20 callers) refusing "empty handler +
   dispatchable status" at insert time. Catches a producer that files born-`triaged`,
   which the promoter guard cannot see.
3. **Producer-side single expression of intent.** A comment repeated at six sites is
   what failed here — three of the six got it wrong. One constructor, or one named
   status choice, and the intent stops being folklore.
4. **DB CHECK constraint** making the pair unrepresentable. Real close-the-door, but
   **ordering hazard**: DB config is live immediately, Go is inert until the roll, so
   a constraint landing first makes the improvement loop's blanket UPDATE **error** on
   any site holding a flag-only `detected` row. Only after 1 has rolled.
5. **Repair the 60 rows** — after the door is closed, not before, or they re-block.

## Status

- `090` diagnosis run filed 2026-08-16, intake `ce78493f-…`, run correlation
  `d1477c1d-806d-…`. Filed **before** asserting the root cause, per the 2026-07-31
  owner ruling — this file's chain is first-hand, and the loop is the independent check.
- Fable was asked for the fix plan and returned nothing: *"You've reached your Fable 5
  limit"*. Planning proceeded here; re-run when the limit resets if it would add.
