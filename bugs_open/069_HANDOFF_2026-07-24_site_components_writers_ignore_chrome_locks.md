# Handoff — site_components (chrome) writers ignore the lock columns, same defect as 058

**Filed 2026-07-24**, spun out of `/bugs_open/058` as its confirmed residual. 058 fixed the
write-side lock gate for **page_components**; the chrome table has the identical hole and was
deliberately left out of that commit to keep the reviewed change narrow.

## The defect

`site_components` carries `locked_at / locked_by / lock_type / lock_expires_at` (added by the same
applied 053/115 migration as page_components, with the same CHECK constraint), and the admin
surface sets them:

- `internal/core-manager/admin/page_admin_handlers.go` `HandleLockSiteComponent` /
  `HandleUnlockSiteComponent` (and the site-component edit endpoint's auto-lock-on-edit).

**No chrome writer reads them back.** Audited 2026-07-24:

- `fix_component_template_action.go` — four `UPDATE site_components SET rendered_html = …` sites
  (~:276, :452, :528, :598), keyed on site_id+slot_name. No lock check.
- `render_site_components_action.go:665` — `UPDATE site_components SET rendered_html = …,
  build_status='rendered'`. No lock check. (Its :526 INSERT is birth-only.)
- `link_site_components_action.go:145` — `INSERT … ON CONFLICT (site_id, slot_name) DO UPDATE`.
  No lock check.

So a human who locks a header/footer/head slot via the admin dashboard gets the same silent
overwrite 058 fixed for page sections: the next chrome re-render or template fix discards the
locked artefact.

## Why the fix is mechanical now

058 shipped the pieces (commit `82ae5a550`):

- `pageComponentAgentWritableSQL(alias)` (`lock_helpers.go`) — the predicate is table-generic
  (both tables share the column set); append it to the three UPDATE/upsert WHERE clauses above and
  check RowsAffected, exactly as done for `rebuild_blog_listing` / the colour fixers.
- `emitLockBlockedChangeItem(...)` — reuse for the signal; set `surface:"site_component"` in spec
  (the field already exists) and key the item on the slot.
- The admin endpoints already stamp `lock_type`/`lock_expires_at` on site_components locks (058's
  admin-handler edit covered both tables), so classification needs no backfill for new locks.

One design question the fixer must answer: `render_site_components` re-renders chrome from
templates on data changes — a locked slot that refuses re-render will serve stale nav until
unlocked. That is what the lock MEANS (058 took the same line for page sections, with the
`lock_blocked_change` item making the tension visible), but the fixer should confirm no chrome
flow treats the skip as a hard failure.

## How to verify

Mirror 058's recipe on a chrome slot: lock the footer row (`lock_type='permanent'` — NOT
`'admin'`, which violates the CHECK constraint), drive a chrome re-render, assert the locked row's
`rendered_html`/`updated_at` unchanged, an unlocked sibling slot re-rendered, and one
`lock_blocked_change` item with `surface:"site_component"`.

## Related

- `/bugs_open/058` — the page_components half: fix committed 2026-07-24 (`82ae5a550`), carries the
  shared predicate + emitter this fix should reuse, plus the corrected lock semantics
  (free-text `locked_by`, `lock_type` is the classifier, NULL type = hard).
- `/bugs_closed/053` (chrome nav fallback) / `/bugs_open/049` (stale chrome) — chrome re-render
  cadence context: chrome artefacts are cached and re-rendered rarely, so a locked-slot skip has a
  long shadow.
