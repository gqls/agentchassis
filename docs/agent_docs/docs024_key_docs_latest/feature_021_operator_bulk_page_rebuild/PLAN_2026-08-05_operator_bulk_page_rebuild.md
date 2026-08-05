# PLAN 2026-08-05 — `features_open/021`: operator-driven bulk page rebuild

**Feature file:** `features_open/021_FEATURE_operator_bulk_page_rebuild.md`.
**Raised:** 2026-07-25. **Picked up:** 2026-08-05, after a survey of `features_open/`
found its stated blocker (`bugs_open/070`) had closed 2026-07-27 and nobody had
picked it up since (`maintenance_queue` still showed the same 2 historic rows).

## The ask, restated

An operator with a legitimate reason to rebuild specific pages (new component
placed in the plan, rewritten writer prompt, corrected evidence base) has no
supported way to ask for it. The five-page fundamentallyai.com workaround
(2026-07-25) hand-mutated an unrelated old `site_work_items` row and collided
with the stale-item reaper. The original filing proposed building: (a) an
operator entry point, (b) sequencing, (c) a rebuild item type that "says it is
a rebuild," (d) recompose-vs-rerender intent, (e) dry-run-first.

## Correction to the original filing's premises — read this before (c) and the
## reaper dependency, both of which turned out not to apply the way filed

The filing reasoned from the operator's AD HOC workaround, which used
`site_work_items` directly. **The actual dormant paved road never touches
`site_work_items` at all.** Read `platform/orchestration/actions/maintenance_actions.go`
before assuming otherwise:

- `flagPagesForRebuild` (line 962) is a pure `UPDATE pages SET
  build_status='needs_rebuild'`. No `site_work_items` row is read or written by
  this path, ever.
- `stale-work-item-reaper`'s predicate (checked live, 2026-08-05:
  `SELECT pre_query FROM scheduled_tasks WHERE name='stale-work-item-reaper'`)
  only ages `site_work_items` rows (`WHERE status='triaged' AND pipeline='build'`).
  It cannot see `maintenance_queue`, which has its own status column.

**Consequence: (c) is not needed, and `bugs_open/070` is not a genuine
prerequisite for THIS path** (it remains a real prerequisite for the operator's
original by-hand workaround, which is a different, worse route this feature
makes unnecessary). This is a correction to the feature file's own reasoning,
not a disagreement with its goal — recorded here rather than silently, per this
repo's standing-docs discipline.

## What's actually built vs. what's missing (verified live, 2026-08-05)

| piece | state |
|---|---|
| `maintenance_queue` table | exists, correct shape, sensible defaults |
| `maintenance-triage` agent workflow (`scan_and_queue` → `check_dry_run` → `prepare_rebuild_dispatches` → `check_has_rebuilds` → `spawn_rebuilder` → `rebuild_loop`) | live, `is_active=true`, read in full — matches the feature file's description |
| `PrepareRebuildDispatchesAction` / `flagPagesForRebuild` | present, read in full, behave as documented |
| a `scheduled_tasks` row driving `maintenance-triage` | **none** — confirmed still true |
| **an operator entry point** | **built this session**: `scripts/rebuild_pages.sh` |

So the actual gap was narrower than the filing assumed: **one script**, not new
Go code. See `NOTES` for the full trail, including a bug the script's own first
test run caught (a `psql -t` / `INSERT ... RETURNING` command-tag leak) and a
design gap the same test caught (the existing `check_dry_run` branch previews
the automated scanner, never an operator-supplied page list).

## What is explicitly NOT done, and why

1. **Intent (recompose vs re-render) is not wired to any behaviour.**
   `page-rebuild`'s `build_pages_loop` always calls `plan_sections` + a
   content-writer step per page — there is no re-render-only fast path today.
   `features_open/012`'s `recompose_pages` vocabulary lives in a **different**
   pipeline (site-plan validation, `v3_site_actions.go`), not this one. The
   script writes `payload.intent` so a future Go change has somewhere to read
   it from, but every rebuild dispatched today is a full recompose. The
   ORIGINAL five-page fundamentallyai.com case wanted exactly that (new
   components + new copy), so this is not a blocker for the primary use case —
   but it means "just re-render existing content through fixed templates,
   cheaper than a full recompose" is NOT available yet if a future operator
   wants it. Building that would be a real Go change to `page-rebuild`'s
   workflow, needing its own scoping and (per this repo's platform-change
   norms) a council submission.
2. **No live (non-dry-run) dispatch has been fired.** `page-rebuild` sweeps in
   EVERY page on the target site already at `build_status='needs_rebuild'`,
   not only the ones named — confirmed live on gaswholesalers.com, which
   already carries 7 such pages from unrelated history. A real dispatch is a
   90-minute-per-site, content-regenerating, real-deploy action on a live
   commercial site. That is a deliberate stopping point for this session, not
   an oversight — see `HANDOFF` for exactly what to check before firing one.
3. **No topic/queue-impact assessment beyond a note.** The script uses the
   shared `system.agent.generic.requests` lane, same as most ad-hoc triggers.
   council-gate got its own topic (`bugs_open/096`) specifically because its
   runs take minutes and were head-of-line blocking the shared lane;
   `page-rebuild` calls run up to 90 minutes. Worth measuring impact after the
   first few real dispatches, not designing for in advance of any usage.

## Immediate next step (see HANDOFF for the full checklist)

Choose a real, deliberate first live-fire target — NOT gaswholesalers.com,
whose 7 pre-existing `needs_rebuild` pages have unknown history and would ride
along uninvited. Read the pre-flight sweep-in report for the chosen domain
first; only proceed with `DRY_RUN=0` once that report is understood and
accepted.
