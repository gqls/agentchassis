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

---

## FIX COMMITTED 2026-07-24 (`82ae5a550` + gofmt sweep `3a309cbeb`) — OPEN until image roll + live behavioural proof

Council gate: **APPROVED 2026-07-24 20:56 UTC**, corr `d2539ca6-ff16-414d-897d-363ebc559df0`
("approved with 4 advisory objection(s) — none high-severity", 3 abstained). Trailer carried on the
follow-up commits `47c2db46a` + docs. Objections and what came of them — see "Council response"
below; one of them caught a real gap (a new writer landed mid-task) fixed in `47c2db46a`.

**What shipped** — candidates 1 + 3 together, with candidate 2's cheap SQL form used where it fits:

- **One shared expiry-aware predicate** (`lock_helpers.go` `pageComponentAgentWritableSQL`), taken
  verbatim from the applied 053/115 schema migration: agent-writable iff `locked_at IS NULL OR
  (lock_type='timed' AND lock_expires_at IS NOT NULL AND lock_expires_at < NOW())`. Any other locked
  row (permanent / review / NULL type / unexpired timed) blocks automation. No force-override in v1 —
  031_LOCKS approved policy: the override is manual human unlock.
- **`save_page_sections`** (the choke-point): actively-locked rows survive the DELETE+INSERT with
  **row identity intact** (same id, untouched `updated_at`); a locked slot's fresh copy is discarded
  and only the row's `position` moves to follow the new composition (exact slot match, then
  kebab-normalised fallback, consume-once for duplicate slots); a locked slot the composition
  dropped is retained after the new set. Result reports `locked_sections_preserved`.
- **`apply_section_edit`**: pre-check + race-free predicate on both UPDATEs (`errComponentLocked` →
  skip-result `{skipped:true, locked:true}`, not an error — an error would retry the orchestration
  against a state only a human unlock can change).
- **`rebuild_blog_listing`**, **`fix_harcoded_colours`**, **`fix_forced_text_colours`**: guarded
  UPDATEs, loud skips.
- **Candidate 3 signal**: every blocked overwrite/removal/edit files a deduped
  `lock_blocked_change` work item, `status='needs_human_review'`, **no handler_agent** (the
  owner-confirmed dead-control routing; surfaces on the 033 dashboard) via the shared
  `insertWorkItem` (no 42P10 drift). Best-effort — never blocks the write path.
- **Admin endpoints** (`page_admin_handlers.go`) now stamp/clear `lock_type`/`lock_expires_at` via
  `LockPolicyFor` on lock/unlock/remove/restore (page_components AND site_components) — the May
  lock-coherence plan's Step 2, scoped.

**Corrections to this file's own claims** (found while fixing):

> **CORRECTED 2026-07-24:** the verification example above (`lock_type='admin'`) would violate the
> live CHECK constraint — `chk_page_components_lock_type` allows only
> `permanent|timed|review`. Use `lock_type='permanent'` (or leave it NULL; the gate treats a locked
> row with no lock_type as hard, conservatively).

> **CORRECTED 2026-07-24:** `CheckComponentLock`'s original hard/soft switch
> (`locked_by IN ('admin','admin-removed','checkpoint')`) misclassified **every live lock as
> soft** — real rows carry free-text reasons in `locked_by` (`182_legal_pages`, the idea.uk CTA
> reason strings). Reworked to classify on `lock_type` via `lock_policy.go` (`IsHardLockType`),
> which was committed in May 2026 with zero callers until now.

> **CORRECTED 2026-07-24 (writer list):** of the 10 files listed under "Writers that overwrite",
> only `save_page_sections` and the `apply_section_edit` helpers overwrite existing copy from
> automation. `loop_actions.go` has **no DB write at all** (in-memory maps); 
> `render_site_components_action.go` and most of `fix_component_template_action.go` write
> **site_components** (chrome), not page_components; `create_tool_component` / `deploy_tool` are
> INSERT-only (`ON CONFLICT DO NOTHING` — a new row cannot be locked);
> `update_component_html` / `store_generated_component` / `link_site_components` touch metadata or
> other tables. The admin surface is exempt (it sets the locks; its regenerate fan-out at
> `page_admin_handlers.go:843` already filtered locked rows).

**Residual, spun out:** `site_components` (chrome) has the identical defect — admin lock endpoints
exist, and NO chrome writer reads the columns. Filed as `/bugs_open/069` (the shared predicate is
table-generic, so the fix there is mechanical).

**Council response (verdict APPROVED, 4 advisory objections, 2026-07-24):**

- *"Audit-completeness rests on a written claim — verify codebase-wide"* (bug_historian, medium) —
  **re-verified 2026-07-24 post-verdict, and it caught a real one**: a fresh grep of every
  `UPDATE/DELETE … page_components` found `create_report_page_action.go` overwriting
  `rendered_html`+`content_data` by id ungated. It was NOT an audit miss: the file landed
  (`2849564ec`, 21:07) **nine minutes before the 058 fix commit** (21:16), added by the concurrent
  gripper-dossier thread after the audit ran. Gated in `47c2db46a` (same pattern as
  rebuild_blog_listing; the locked branch returns the STORED artefact downstream so
  validate_page_content checks what will actually serve). Lesson for this repo's concurrency model:
  a writer audit is stale the moment another session commits — re-grep at commit time, not just at
  plan time. All other grep hits are metadata-only (build_status/slot_name/position/component_id),
  already gated, or the exempt human admin surface.
- *"site_components stamping is scope creep / ambiguous"* (editquality, medium) — deliberate and
  now documented: site_components is a **different table** (chrome slots) with the same lock
  columns from the same migration; 058 stamps its admin lock endpoints so new chrome locks carry
  `lock_type`, but the chrome *writers* are `/bugs_open/069` (kept out to keep this change narrow).
- *"Extra UPDATEs inside the DELETE+INSERT — partial-crash could leave locked rows deleted"*
  (editquality, low) — structurally impossible: the DELETE's predicate **excludes** locked rows, so
  they are never deleted; the per-locked-row UPDATEs only move `position` on rows the DELETE
  skipped. (There was no single wrapping transaction before this change either — the save already
  proceeds past a failed DELETE.)
- *"loadActiveLockedRows failure silently degrades signalling"* (bug_historian, low) — already
  handled as asked: the preload failure logs Warn (`save_page_sections_action.go:705`), and the
  DELETE predicate independently protects the rows; only slot-matching (and hence the `remove`
  signal) is lost in that branch.

**Post-roll verification recipe** (the failing branch, not a pod-grep):

1. On a scratch page: `UPDATE page_components SET locked_at=now(), locked_by='058-verify',
   lock_type='permanent' WHERE id='<pc>';` — note `rendered_html` md5 + `updated_at`.
2. Drive a rebuild of that page (or an `apply_section_edit` at the locked component).
3. Assert: locked row's `rendered_html` md5, `id` and `updated_at` **unchanged** (position may
   move); an **unlocked sibling on the same page WAS rebuilt** (do not pass this by never writing);
   one `site_work_items` row `item_type='lock_blocked_change'`, `status='needs_human_review'` for
   the page/slot.
4. Unit tests already cover the SQL-level branches (`lock_gate_test.go`, 13 cases, green 2026-07-24
   via `git archive HEAD` overlay: full `actions` + `core-manager/admin` suites pass).
