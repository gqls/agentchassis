# PLAN 2026-09-03 — bugs_open/450: planned tool pages built as prose shells

Lane opened 2026-09-03 by the 450 fixing thread (session `bugs_open/450`). The bug was filed
2026-09-02 by the `portfolio_positioning` lane, which owns the **instance** cleanup only (owner
ruling 2026-09-03: build the 8 planned tools). No fixing thread existed for the **class** —
`scripts/who-owns.py 450` named the filing lane, and its handoff addressed the fix "to the fixing
thread". This lane is that thread.

Bug file:
`bugs_open/450_HANDOFF_2026-09-02_planned_tool_pages_are_built_as_prose_shells_by_the_link_repair_before_their_tools_exist.md`

## 1. The defect, in one paragraph

A site plan names tool pages (`pages.page_type='tool'`) with generic sections
(`hero-tool,generic-text-block`) **before the tool exists** — tools arrive hours-to-days later
from the design rotation, under names the planner never sees (seotools: 0 of 7 planned names
matched what `tool-deployer` eventually built). The plan-time hold that says "not the generic
builder" (`owned_page_review`) has **no consumer**. Five generic producers —
`unbuilt_internal_link` (first to fire on a fresh remake), `empty_section`, `needs_page`,
`needs_content_page`, `page_rerender` — route the page to `page-build-handler`, whose save guard
asks only `pages.rebuild_policy='owned'`. A planned tool page is `'generic'`, so the guard passes
it and the generic builder writes exactly what the plan said: prose about the tool. The URL then
serves 200 at full weight with the tool's own headline and **zero forms**, and every link-shaped
detector clears itself.

## 2. Why the obvious fix is the wrong one

The bug's candidate 2 reads "when the hold is filed, also set the page's `rebuild_policy`". Three
findings kill it as written:

- `pages.rebuild_policy` is CHECK-constrained to `'generic'|'owned'` (migration 164). A third
  value needs a constraint migration and is invisible to **every** existing reader, each of which
  is an independent `='owned'` literal (~12 of them: the guard, `ownedPageExclusionSQL` + its
  census, the `writeWorkItem` door, `loadVerbatimPageHTML`, reconcile, `UpdatePageStatus`, the
  fixer's livespec-pinned SQL, migration 410's router, …).
- **Nothing has ever UPDATEd the column.** Zero `UPDATE … rebuild_policy` statements exist in Go;
  the only writers are two INSERT paths (`adopt_verbatim.go`, `create_report_page_action.go`) and
  a handful of hand-run migrations. So "set it at plan time" would introduce the estate's first
  policy lifecycle — and there is no event today that would ever clear it when the tool lands.
- Reusing `'owned'` overloads a word that already means *verbatim/adopted* elsewhere: an `owned`
  page with one component is served from stored HTML rather than assembled
  (`loadVerbatimPageHTML`), and owned pages are excluded from build selection wholesale.

**So the predicate is DERIVED, not stored:** a page refuses generic builds when
`rebuild_policy='owned'` **OR** it is a *tool shell* — `page_type='tool'` with no live
`component_level='tool'` component. No migration, no new value, no lifecycle, and it
**self-clears** the moment the tool component is inserted.

## 3. Decisions and their reasons

| # | Decision | Reason |
|---|---|---|
| D1 | Derived predicate, not a stored policy | §2. Self-clearing; no lifecycle to forget. |
| D2 | One predicate in `owned_page_guard.go`, replacing `pageIsOwnedForGuard` (rename, not a sibling) | The file's own doctrine — "the ONLY place a pipeline may read ownership policy". Two predicates for "may the generic builder touch this page" is the drift class the `reuse_agent` seat objected to when this file was first written. The rename compile-forces all four call sites. |
| D3 | Close the door at `writeWorkItem` (file time) **and** at the handler/save/assemble seams | The write door costs nothing per finding and stops the class at source; but **339 `unbuilt_internal_link` rows are already queued**, and a write-time door cannot see them. `load_page_record`'s `refuse_owned_page` arm terminates those before any LLM spend. |
| D4 | Do NOT edit `check_phantom_internal_links.go` | Candidate 3's effect is achieved for **all five** producers at one seam. Editing the check binds one producer — "a guard only guards the door you walk through" — and `availableBuilders` is unreachable from `discovery_checks` (import cycle; bugs_open/220 deferred exactly this). |
| D5 | `'tool'` only in v1 | No `component_level='game'` exists, so a game arm could never self-clear; `entity-page` already routes to `capability_gap` via `builderForPageType`. The fragment is a function — a later arm is one disjunct. |
| D6 | Kill-switch `DISABLE_TOOL_SHELL_REFUSAL`, **armed by default**, disarming only the new arm | Owner ruling against default-OFF switches that rot unexercised; the growth-posture door's precedent. Independent rollback from the owned arm. |
| D7 | Plan-side gate is a **sibling key** `enforce_tool_sources` (default OFF), not a widening of `enforce_listing_sources` | The 444 session's CONTRIB asked for exactly this: independent rollback (a tool-arm misfire must not cost the live listing gate), and a tool page is not semantically a listing page. |
| D8 | The plan-side gate holds **empty-sectioned** tool pages too | The bug's own websitepromotion branch: a sectionless tool page does not "park harmlessly" — it produced 7 HITL-parked `unbuilt_internal_link` items plus a `needs_content_page`, recurring per remake, on a page row §7 proved no producer will ever fill. |
| D9 | Tool gate runs **before** the listing gate in `validate_site_plan` | Held tool children make the `/tools/` hub resolve zero children, so the listing gate holds the hub too — no phantom `/tools/` URLs at all. The reverse order ships an empty hub (a 444-class page). |

## 4. Phasing

1. **Commit 1 — the door-closer** (Go only). Derived predicate + the six consultation points +
   receipts carrying a `refusal_class` + mutation-proved tests + concept-register entry in the
   same commit (platform-seam ruling, condition 2).
2. **Commit 2 — the supply cut** (Go + migration). `enforce_tool_sources` sibling arm in a new
   `tool_item_sources.go`, finding code registered, migration arming `build-site-planner` on
   720's pattern.
3. **Commit 3 — docs and coordination.** Bug-file fix record + the emitter correction, CONTRIBs
   to the lanes that own the neighbouring bugs, 016b/LANDMINES pointers.

Commits 1 and 2 are independent: neither depends on the other landing, and the estate is safe with
either alone (1 without 2 = stubs still planned but never filled generically; 2 without 1 = no new
stubs, existing ones still reachable by the five producers).

## 5. Corrections to the originating brief

> **CORRECTION 2026-09-03 (this lane, at the code).** The bug file §2 attributes the
> `owned_page_review` hold to `validate_site_plan`. The action that writes that exact summary is
> **`ReconcileSitePlanAction`** (`reconcile_site_plan_action.go:270-300`), a later step of the
> same `build-site-planner` workflow; `sync_pages` mints the `pages` rows between the two. The
> distinction matters because it is why the hold cannot carry a page id at plan time — at
> `validate_plan` the row does not exist yet. Caught by reading the emitter rather than trusting
> the attribution.

> **CORRECTION 2026-09-03 (this lane).** A design note carried into this lane's own brief said
> `578_retype_mislabelled_tool_rows_HOLD.sql` retypes mislabelled **`pages.page_type`** rows. It
> does not — it retypes mislabelled **`page_components`** (tool bytes sitting in `hero` rows). No
> held migration retypes page rows. The misfire assessment in §3/D5 stands either way, but the
> evidence cited for it was wrong and must not be repeated in the council submission.

## 6. Open questions

- Does any live site carry a page wrongly typed `'tool'` whose generic rebuild is *wanted*? The
  refusal is loud (a review row per page), so the answer arrives as evidence rather than silence —
  but a census before the roll would size it. **[UNMEASURED as of 2026-09-03]**
- Residual, explicitly out of v1: `page_rerender` re-deploying the **existing** 61 shells
  (re-assembly of existing components is migration 164's sanctioned owned-page deploy path; gating
  it is the bug-210 family). Stated in the commit, not fixed by it.
