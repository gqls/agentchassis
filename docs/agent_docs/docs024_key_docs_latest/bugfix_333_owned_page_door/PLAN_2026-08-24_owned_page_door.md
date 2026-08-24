# PLAN — 333: policy routability at the door

**Lane opened 2026-08-24.** Bug: `bugs_open/333_HANDOFF_2026-08-19_producers_route_content_findings_at_page_build_handler_without_reading_rebuild_policy_so_owned_pages_queue_findings_that_can_only_be_refused.md`

## What the bug is, in plain terms

A **work item** is the platform's durable record of "something is wrong with this page, and here is who
should fix it". The "who" is a column, `handler_agent`.

Some pages are marked `rebuild_policy='owned'`. That means "this page belongs to a tool or widget — the
generic page builder must not rewrite it", and the generic builder (`page-build-handler`) enforces that by
refusing the item outright.

The defect: about two dozen places in the code file content findings at `page-build-handler` **without ever
asking whether the page is owned**. So a real defect on an owned page is routed at the one handler that is
forbidden to touch it. It is refused, the item ends at `wont_fix` — a status meaning "we decided not to fix
this" — and the finding is gone. The detector finds it again next sweep, and the cycle repeats.

## The decision this lane takes

**Fix candidate 1 from the bug file: a check at the shared write seam (`writeWorkItem`), beside the one
`bugs_open/291` already put there for unregistered handlers.** Everything else in the bug's list either
repeats the predicate in ~26 places (candidate 3, which `bugs_open/266` rejected for good reason and the
26th producer arrives open anyway) or leaves the defect in place (candidate 4).

Three design decisions, and the reason for each:

1. **The door reads the handler's own DECLARATION, not a Go list of handler names.** `page-build-handler`
   opts into refusing owned pages through config (migration 488 sets `refuse_owned_page: true` on its
   `load_page_record` step). The door asks the database "does this handler declare that?" — so a handler
   that opts in later is covered without a code change, and a handler that does not (`page-rerender`, which
   completes 5,040 items on owned pages) is never touched.
2. **The parked row KEEPS its own item_type and item_key.** The first draft of this plan followed
   `bugs_closed/077`'s convention literally and re-typed the row to `capability_gap`. That was wrong, and
   the reason is worth stating: a detector retracts its own finding when it stops reproducing, and it does
   so by matching `(item_type, item_key)` (`resolveWorkItems`, `work_items_common.go:443-457`). A re-typed
   row matches nothing, so it would sit parked for ever once the page was fixed — the exact "nothing swept
   it, the items age for ever" hole two council seats caught on `bugs_open/342` a day earlier. We take 077's
   **signal** (`status='deferred'`, empty handler, `spec.gap_kind`/`builder_needed`) and leave the row's
   identity alone. The roadmap sweep reads `(item_type='capability_gap' OR status='deferred')`, so the
   consumer still sees it.
3. **`deferred`, not `blocked` and not `needs_human_review`.** `blocked` self-heals: the feasibility-recheck
   task promotes any blocked row whose handler exists, every 600 s. `needs_human_review` has 1,087 rows and
   no working surface (`bugs_open/033`). `deferred` is the estate's parking state: nothing promotes it, it
   holds its dedup slot so the detector cannot churn, and the roadmap sweep already reads it.

## Phasing

| phase | what | state |
|---|---|---|
| 0 | Ownership check, bug re-validation, census | done 2026-08-24 |
| 1 | `readRebuildPolicy` extraction (Tx-capable), the two shared renderers | |
| 2 | The door in `writeWorkItem` + `ownedPageParkedItem` (pure) | |
| 3 | Producer honesty (`create_work_item`, `raiseToolContentItem`) | |
| 4 | Tests incl. mutation proof + re-scripting the 5 existing seam tests | |
| 5 | Council round; commit with `Council-Submitted:` | |
| 6 | Register WII-028, LANDMINES, 016b §9, consumer CONTRIBs | |
| 7 | Post-roll verification (INERT until the chassis rolls) | |

## Corrections to the originating brief

> **CORRECTION 2026-08-24 (before any code was written):** the bug file's fix candidate 1 offers two demoted
> shapes — (a) `capability_gap` with a reason, or (b) keep the type at `needs_human_review`. **Neither is
> taken as written.** (a) breaks the retraction contract (above); (b) parks in a queue the owner has ruled
> should not fill. The shape shipped is a third: 077's signal on the finding's own identity. Recorded here
> rather than silently diverging.

> **CORRECTION 2026-08-24:** the bug file's §"The routing sites" says "30 literal sites in 25 files".
> Re-counted today: **49 matches of the literal in 26 files** (`grep -rn '"page-build-handler"' platform/
> internal/ --include=*.go | grep -v _test`), of which many are comments — the census below counts WRITERS,
> which is the number that matters, and it is 28 sites. Both figures are dated; neither is wrong for its date.
