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

---

## CLOSED 2026-07-26 — FIXED AND LIVE on v1.0.1170/1171, induced-fault proven

Commits `05bcb3586` (the gate) + `d9e7ef7cb` (the lockstep test). Live from
**v1.0.1170** (built and rolled by this thread, 18:35 UTC) and re-verified on
**v1.0.1171** (another thread's roll, 21:02 UTC), which is the binary the proof below ran against.
**No `Council-Reviewed:` trailer** — see "Council" below; the verdict was REVISE.

### Corrections to this file's own claims, found while fixing

> **CORRECTED 2026-07-26:** "Its :526 INSERT is birth-only" is **false**. That statement is
> `INSERT … ON CONFLICT (site_id, slot_name) DO UPDATE SET component_id = $3` — it mutates an
> existing row, repointing a locked slot at a **generic default template picked by function name**,
> and it discarded both its error and its result. Far from being exempt, that branch is the *worst*
> case of this bug: it fires exactly when `component_id IS NULL`, which is a real, recurring state
> (`checkUnlinkedSiteComponents` exists for it). The probe below exercises it deliberately.

> **CORRECTED 2026-07-26:** the writer list is complete for Go, but the table has a **fourth**
> mutation path the audit had to find from the DB side rather than the code:
> `revert_site_to_snapshot` deletes and re-inserts every chrome row for a site, and destroyed the
> lock columns while doing it. Out of 069's scope (human-initiated, and it hits `page_components`
> identically) — filed and fixed as `/bugs_closed/088`.

### What shipped

- **`site_component_lock_guard.go`** (new): `CheckSiteComponentLock` (keyed `site_id + slot_name`,
  and reporting `RowExists` / `HasHTML` / `ComponentID` from the one row read),
  `setSiteComponentHTML` / `appendSiteComponentHTML` / `relinkSiteComponent` (predicate in the
  WHERE ⇒ race-free; `errSiteComponentLocked` on zero rows), and
  `emitChromeLockBlockedChangeItem` (`surface:"site_component"`, item_key
  `lock_blocked_change:site_component:<slot>`, **high** severity for `header`/`footer` because
  chrome is site-wide — a locked header is the `049` mechanism — `medium` for `head`).
- **`lock_helpers.go`**: `classifyComponentLock` extracted so the hard/soft rule cannot drift
  between surfaces, and the emitter body moved to a struct-based `emitLockBlockedChange` core.
  `emitLockBlockedChangeItem`'s signature is **unchanged**, so its four page-side call-site files
  were never touched.
- **`render_site_components`**: pre-check **below** the `!force` idempotence exit (see the decision
  below), predicate on both writes, and a zero-row store that is *not* locked now logs `Error` and
  returns failure — previously the result was discarded, so a store that wrote nothing still logged
  "rendered and stored" and reported success. Result carries `locked_slots_preserved`.
- **`link_site_components`**: the only writer that *erased* the artefact (`rendered_html = NULL`).
  Here the pre-check is **load-bearing, not advisory**: the upsert's own `IS DISTINCT FROM` guard
  makes zero-rows the *normal* outcome, so `RowsAffected` cannot tell "already correct" from "a lock
  refused me". Already-correct is checked *before* locked, so a locked slot needing no change files
  nothing.
- **`fix_component_template`** ×4: one shared `chromeFixLockSkip`, returning this file's existing
  `{"fixed": false, "action": "needs_review", …}` vocabulary — without `action:needs_review` the
  handler reports success and the two-strike rule parks the re-detected item `unresolved` two cycles
  later. `fixChromeOverflow` still applies the **shared** `content_components` template patch (a
  shared bug deserves the shared fix) but no longer claims a durable per-site outcome:
  `site_slot_locked`, `durable_for_this_site: false`, `blocked_action: "rerender"`.
- **`checkUnlinkedSiteComponents`**: lock filter, so the detector and its now-refusing handler cannot
  disagree by construction. The predicate is hand-copied (discovery_checks cannot import package
  actions — the dependency runs the other way), so it is **pinned by a test**:
  `TestDiscoveryChromeLockFilterMatchesSharedPredicate` builds the expected text from
  `pageComponentAgentWritableSQL("sc.")` and reads the sibling file. Proven non-vacuous by induced
  fault (flip `'timed'`→`'review'` in the detector ⇒ test fails; restore ⇒ passes).
- 16 sqlmock cases in `site_component_lock_guard_test.go`, green with the full `actions`,
  `discovery_checks` and `core-manager/admin` suites on a `git archive HEAD` overlay.

### Decisions, so the next thread does not re-litigate them

1. **The gate sits BELOW the `!force` early exit** (`render_site_components_action.go`). That path
   already no-ops on a populated slot, so checking above it would file a work item claiming a writer
   "wanted to change this" for a call that was never going to write. **Consequence, stated plainly:
   the gate only bites when a caller passes `force_rerender: true`** — 4 of the 6 live agents. A
   verification driven by an unforced call passes vacuously.
2. **The design question this file asked** ("confirm no chrome flow treats the skip as a hard
   failure") — answered from the live DB, not by reading Go: **no** step config or condition reads
   the action's `rendered` map (`default_config LIKE '%render_site_components.rendered%'` ⇒ 0 rows),
   and the action already returned `success: true` regardless. Nothing fails because of a skip.
3. **A locked slot with EMPTY `rendered_html` freezes that chrome site-wide** until a human unlocks,
   because the page assembler filters empty chrome. Blocking uniformly is still right — a softer
   policy cannot live in a WHERE predicate, and anything decided outside it reopens the TOCTOU the
   predicate exists to close — so the item carries `artefact_empty: true` and says so in its fix text.
4. **Skip-results, never errors** (058's line): an error would retry the orchestration against a
   state only a human unlock can change.

### Live measurements that frame this honestly (2026-07-26)

42 chrome rows, **0 locked**; 39 locked `page_components` rows; **0** `lock_blocked_change` items
since 058 shipped; **0** admin chrome edits ever recorded
(`spec->>'reason' = 'admin_site_component_edit'`). The defect was reachable and ungated but had
**never fired**, so this is preventative and a green run proves nothing — the fault had to be induced.

There *is* a self-defeating flow now protected: `HandleUpdateSiteComponent` auto-locks the slot it
edits (`shouldLock` defaults **true** ⇒ `LockPolicyFor("admin")` ⇒ permanent) and then files a
`needs_rerender` item with `refresh_site_components: true`, which `rerender-pages` routes to
`render_site_components` at `force_rerender: true` — i.e. the admin's own follow-up re-render used to
overwrite the edit it had just locked. From now on it is refused and files **one** deduped item; the
item's fix text says explicitly that this is expected after a dashboard edit.

### How it was proven — induced fault on the deployed binary

**Deployment first** (never git, never the tag): pod `agent-chassis-5b4456686c-s5fkc`, v1.0.1171 —
`strings /app/agent-chassis | grep -c` found each literal the fix CREATED
("refusing to re-render human-locked chrome slot" ×1, `lock_blocked_change:site_component:` ×1,
"automated chrome write refused" ×1, "refusing to relink human-locked chrome slot" ×1,
`locked_slots_preserved` ×1), plus a positive control (`chrome_dead_control` ×5) and a negative
control (a string nothing should contain ⇒ 0).

**Probe**: three scratch `site_components` rows on dartsonline — none of them a real chrome slot, so
nothing the site serves could be touched — driven by a scratch one-step agent
(`scratch-069-chromelock` → `render_site_components`, `force_rerender: true`, all three slots) fired
with the 091 kcat envelope, graded on `collected_data->'render_chrome'` rather than on
`status='COMPLETED'`.

| slot | before | after | verdict |
|---|---|---|---|
| `probe-069-locked` (locked, component-backed) | md5 `ebe4388f…`, `updated_at` 18:40:24 | **md5 and `updated_at` identical** | locked artefact survived a FORCED re-render |
| `probe-069-open` (unlocked sibling) | md5 `ae0f8021…`, 41 bytes | md5 `62bddb15…`, **3429 bytes**, `updated_at` 21:07:14 | the gate does not stop legitimate writes |
| `site-footer` (locked, `component_id IS NULL`) | `component_id` NULL, md5 `a0a80afd…` | **still NULL, md5 unchanged** | the generic-default fallback no longer repoints a locked slot — the case 058 could not have |

Result payload: `"locked_slots_preserved": ["probe-069-locked","site-footer"]`, `success: true`.
Exactly **two** `site_work_items` rows, `item_type='lock_blocked_change'`,
`status='needs_human_review'`, **no `handler_agent`**, `spec->>'surface'='site_component'`,
`blocked_action='overwrite'`, keys `lock_blocked_change:site_component:probe-069-locked` and
`…:site-footer`.

**What this probe does NOT prove, said rather than implied:** both items came out `severity=medium`,
because the high-severity branch keys on the literal slot names `header`/`footer` and the probe
deliberately used scratch names — the `high` path is covered by unit test and code reading only.
`link_site_components` and the four `fix_component_template` paths were not driven live either; they
rest on the shared predicate plus their sqlmock cases.

Cleanup: work items deleted first, then fixtures, agent and orchestration rows; leak check returned
**0 on every line** (probe rows, scratch agent, `lock_blocked_change` items, orchestrations,
`agent_error_log`) with dartsonline's 3 real chrome rows still present and untouched.

### Council

Round 1, corr `75dff4cd-e822-4b88-bd98-d989ef32bc90`: **REVISE**, and the reason matters —
`decided_by` is *"unreadable reviewer(s): review_editquality.result"*, a harness fault, not a
substantive block. **10 of 12 seats approved**; the two objectors (bug_historian, guardian) filed
**medium**, nobody vetoed. The reviewers' own read-only checks corroborated every figure in the
submission (0 `lock_blocked_change` items, the six agents wiring the action).

Objections and what came of them:

- *debug_historian (medium): the hand-copied predicate ships with no guard against drift* — **fixed
  in code**, `d9e7ef7cb`, the lockstep test above. The one objection that earned a change.
- *bug_historian (medium): audit-completeness rests on a targeted audit, not an exhaustive grep* —
  re-grepped at commit time (058's own lesson). Every remaining `UPDATE/INSERT site_components` in
  the tree is either gated, one of the new seams, or the exempt admin surface; the DB side was
  swept separately with `pg_get_functiondef` over `pg_proc`, which is how `088` was found.
- *guardian (medium): verify the four external `emitLockBlockedChangeItem` call sites are unaffected*
  — the wrapper keeps the exact signature and fills the struct with the same values (`item_key`
  format, `severity` default `medium`, `priority` 30, spec keys including `page_name`); the full
  `actions` suite is green. Only the log prefix changed (`emitLockBlockedChange…`).
- *guardian (low): `renderAndStoreSiteComponent`'s signature change* — unexported, single caller
  (`:282`), verified by grep.
- *tooling_provenance (low): confirm the detector↔handler mapping from the registry, not by name* —
  `site-component-linker` is the only live agent wiring the `link_site_components` action (DB query),
  and it is the `HandlerAgent` the check files against.
- *render_guardian / bug_historian (low): name the stuck-empty-header consequence, and enumerate the
  other unfiltered chrome detectors* — decision 3 above; the residual is named below.

### Residual, named rather than left silent

Other chrome detectors still scan `site_components` without a lock filter
(`check_broken_nav_links`, `check_component_standards`' sibling sub-checks, `check_generic_theme`,
`check_phantom_internal_links`, `check_integrity`), so a locked stale header can still raise findings
its fixer will now decline — the `bugs_open/077` shape, bounded by the two-strike rule. Only
`checkUnlinkedSiteComponents` (whose handler this change made refuse) and
`check_unverified_claims` filter today. The clean fix is one exported canonical predicate, which
needs the `actions`/`discovery_checks` import direction resolved — a separate, fleet-wide change.
