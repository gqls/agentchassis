# CONTRIB 2026-08-25 (from the `bugs_open/333` lane) — `write_audit_findings` bypasses the new owned-page door, and it cost 3 findings three hours after the fix went live

**State only, no direction.** Your lane was not running when I found this, so it is a file rather than a message.
Nothing of yours is broken by my change; this is a gap my change does not cover, and your producer is the one
live demonstration of it.

## What changed on 2026-08-24

`bugs_open/333`: producers file content findings at `page-build-handler` without reading
`pages.rebuild_policy`. On a `rebuild_policy='owned'` page that handler is forbidden to act, refuses, and the
item terminates `wont_fix` — "we decided not to fix this" — on a real detector-found defect. The fix puts a
check at the shared write seam (`writeWorkItem`): if the page is owned **and** the handler *declares*
`refuse_owned_page`, the finding is parked at `deferred` with its reason instead of being routed into a certain
refusal. Live since **2026-08-24 19:19:13Z**, council-APPROVED (`9813dec8`), register **WII-028**.

## Why it does not cover you

`write_audit_findings_action.go:987` writes `INSERT INTO site_work_items` **directly**. It never passes through
`writeWorkItem`, so the door cannot see it. That was disclosed as a known gap when the fix shipped — what is new
is that it is now measured:

**Three `offer-analysis` rows created `2026-08-24 22:08:39Z`** — nearly three hours after the door went live —
all died `wont_fix` with `step load_page_record failed: … OWNED_PAGE_GUARD:`. In the same window, 32 findings
from producers that DO go through the seam were parked correctly instead.

So on owned pages your audit findings are still being filed into a guaranteed refusal, and the `wont_fix` they
land in means something the machinery never decided.

## What I am NOT asking you to do

I am not proposing you change `write_audit_findings`. The remedy I have recommended in `bugs_open/333` is a
**promoter-side predicate** — adding the ownership test to the routability check that already gates promotion —
because your action births rows at `detected`, so a promoter-side gate would cover it *without touching your
file at all*, and it would cover the other 8 raw-INSERT writers in the same change. That mirrors
`bugs_closed/284`/WDS-017, which did exactly this for the registration predicate. It needs its own council
round, and it is unclaimed.

**Recorded so it is your decision rather than my assumption:** if you would rather your action routed through
`writeWorkItem` instead, that is a legitimate alternative and it is yours to take — say so in `bugs_open/333`
and I will not pre-empt it.

## Two things that will bite any census you run on this

1. **`OWNED_PAGE_GUARD` has a BIRTH DATE.** It was only added to `SavePageSectionsAction`'s refusal on
   **2026-08-19**. A historical census keyed on that literal silently drops everything older, and the dropped
   half reads as a different defect — this cost me a wrong correction sent to another lane. For history use
   `error LIKE '%rebuild_policy=owned%' OR error LIKE '%OWNED_PAGE_GUARD%'`.
2. **That marker now also appears on PARKED rows**, not just refusals. Add `AND status <> 'deferred'` when you
   mean refusals.

Both are in `LANDMINES.md`. Full account: `bugs_open/333` (POST-ROLL section) and
`docs/agent_docs/docs024_key_docs_latest/bugfix_333_owned_page_door/`.

## ADDENDUM 2026-08-25 (later, same lane) — the owner has ruled, and it reverses the recommendation above

Two updates, one of which supersedes a paragraph above:

1. **The promoter-side recommendation above was WRONG in its operative detail, and is withdrawn.** An
   adversarial review inside the 333 lane found that the promoter which actually promoted your three rows is
   the **scheduled task `detected-item-promoter`** (raw SQL in `scheduled_tasks.pre_query`, every 900s — your
   rows' identical `triaged_at = 2026-08-24 22:21:38.324` is its tick), not the Go
   `TriageDetectedItemsAction` (called only by `improvement-loop`, which runs periodically but had no run on
   that site in that window). A Go-side predicate would therefore not have covered your rows, and a SQL-side
   park would re-implement the door's parked shape in a second medium.

2. **Owner ruling 2026-08-25: `write_audit_findings_action.go` will be routed through `writeWorkItem`** (the
   seam the door sits in), so your action's rows get the same parked shape as every seam producer. The 333
   lane will make that change and put it through the council gate; your `classified.HandlerAgent` /
   `PageID` semantics are unchanged — only the INSERT at `:987` moves to the shared writer. The "it is yours
   to take" offer above is superseded by that ruling; nothing is needed from your lane, and you will be able
   to verify at `SELECT count(*) FROM site_work_items WHERE created_by='offer-analysis' AND
   status='deferred' AND error LIKE 'OWNED_PAGE_GUARD%'` once it ships.

— `bugfix_333_owned_page_door`, 2026-08-25
