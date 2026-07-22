# Handoff — the page rebuild/render path does not honour `page_components` locks

**Filed 2026-07-22**, spun out of `/bugs_open/038` as its candidate 3. It is an **independent**
defect: 038's fix stops an *unchanged* deployed page being rebuilt at all, but when the plan
*genuinely* re-composes a page (the correct-to-rebuild case), the rebuild silently overwrites
human-locked component copy.

## The defect

`page_components` carries `locked_at` / `locked_by` / `lock_type` for exactly one purpose: a human
reviews a component, corrects its copy, and locks it so automated writers leave it alone. The
platform even ships the helper for this:

```
platform/orchestration/actions/lock_helpers.go — CheckComponentLock(ctx, db, componentID, logger)
  → ComponentLockStatus{IsLocked, LockedBy, LockedAt, IsHard}
  // "for execution agents that write to page_components and need to respect human locks"
```

**It has zero callers.** Grep across `platform/` (2026-07-22): the only matches for
`CheckComponentLock` are its own definition and doc-comment. No execution/rebuild writer of
`page_components` consults it. The discovery *checks* filter locks in SQL (`AND pc.locked_at IS
NULL`), but the *write* path — the one that regenerates a page's components on rebuild — does not.

Writers that overwrite `page_components` and should be gated (non-exhaustive; grep
`INSERT INTO page_components` / `rendered_html`):

```
render_site_components_action.go   rerender_pages_actions.go   rerender_single_page_action.go
create_tool_component_action.go    deploy_tool_action.go       link_site_components_action.go
fix_component_template_action.go   fix_harcoded_colours_action.go   fix_forced_text_colours_action.go
loop_actions.go
```

## Why it matters

Human-reviewed, locked copy is discarded the moment a re-plan legitimately re-composes the page (or
any of the fix/rerender paths above runs against it). This is the same class as `/bugs_open/029`
(unrequested regeneration) and the reason `/bugs_open/033` (a review-diff surface) exists — but here
the signal the human already gave (`locked_at`) is being ignored outright.

## Fix candidates

1. **Gate every `page_components` writer on the existing helper.** Before overwriting a row, call
   `CheckComponentLock`; if `IsHard` (admin / admin-removed / checkpoint) skip that component unless
   the work item carries `force=true`, mirroring the helper's own doc-comment. Soft (deploy) locks
   may be overwritable — decide per lock_type. The helper already classifies hard vs soft.
2. **Filter at selection time** where a writer picks components to (re)render, adding
   `AND locked_at IS NULL` the way the discovery checks already do — cheaper for the bulk
   render/rerender paths, but each writer must be audited individually.
3. Whichever is chosen, **emit a signal when a lock blocks a wanted change** (a `needs_human_review`
   / the 033 diff), so "the plan wants to change a locked page" surfaces instead of silently
   no-op-ing. A silent skip trades one silent failure for another.

## How to verify a fix

1. Lock a `page_components` row (`UPDATE page_components SET locked_at=now(), lock_type='admin' …`),
   then drive a rebuild/rerender of its page, and assert the locked row's `rendered_html` and
   `updated_at` are **unchanged** (the artefact, not the status).
2. Assert an **un**locked sibling on the same page IS still rebuilt (do not fix this by never
   writing).
3. If candidate 3 is taken, assert the blocked change produces an actionable item, not a silent skip.

## Related

- `/bugs_open/038` — origin; its candidate 1 (stop rebuilding *unchanged* pages) is fixed and live.
  This is the orthogonal half: honour locks when a page *is* legitimately rebuilt.
- `/bugs_open/029` — `tool-suggester` regenerating pages; same consequence, different trigger.
- `/bugs_open/033` — the missing human-review surface; the natural home for candidate 3's signal.
