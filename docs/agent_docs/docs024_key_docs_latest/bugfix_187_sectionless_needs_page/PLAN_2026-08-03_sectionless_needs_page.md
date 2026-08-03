# PLAN — bugfix 187: needs_page items minted for section-less pages park permanently

**Lane opened 2026-08-03 (evening).** Claim commit `411c6ab6f`. 090 diagnosis
run filed same evening, RUN correlation `b3dcb102-d4bf-44c1-b2a2-3068ce95acc6`
(per the 2026-07-31 owner ruling — filed BEFORE committing to the root cause;
first-hand verification below stands on its own, the loop pass corroborates).

## Root cause, per emitter (first-hand: code read at HEAD + all 28 rows measured live)

The handler contract (`load_page_sections_from_spec_action.go`): sections
resolve via (1) `site_plan_sections` for the CURRENT plan → (2) `site_specs`
`site_plan` aspect → (3) `pages.sections` → (4) same-role sibling synthesis,
which is **gated to pages in the current plan**. `ready_count == 0` routes to
`mark_no_ready_sections` (DB workflow config, migration `149_...noop_flags`),
which parks the item in `needs_human_review` with the census's error string.

| source | rows | code path | verdict |
|---|---|---|---|
| image-build-handler | 14 | `flag_page_image_rebuild_action.go:132-159` — emits from ONLY (site_id, page_name); its own header says "VERIFY BEFORE RELYING ON IT" and the assumption is now measured FALSE for section-less pages | **177's shape: unsatisfiable at birth** for pages resolving no sections. Emit-side guard. |
| page-rerender | 4 | `escalateRerenderToWriter`, `rerender_page_sections_action.go:803-830` — emits on a NULL `content_data` slot, from ONLY (site_id, page_name) | **177's shape** — a tool page's single widget slot can legitimately carry NULL `content_data` (components rendering from other than content_data are a named revalidator 'unknown' case), so the escalation asks the writer to rebuild from a section plan that does not exist. Guard. |
| reconcile_site_plan | 9 | `reconcile_site_plan_action.go:291-303` — hand-rolled INSERT; **has `planSections[planID][name]` in scope 2 lines above the emit** | **Split: leave the emitter alone.** Its parked rows are either since-BUILT (4: tungsten-guide, board-setup, cases-index, thames-water) or REAL gaps (5: pages the plan wanted with 0 sections + 0 plan rows now — bugs_closed/015's shape; the item is surfacing a defect and must NOT be suppressed). A guard here would silence 015-shape findings; the drain (below) handles the stale ones. |
| gemini-p7-verification | 1 | none — manual one-off enqueue (gemini_content_provider NOTES:607) | No emitter to fix. Page (grip-styles) is satisfiable NOW (3 declared + 3 plan rows, 0 slots) — real pending work; stays parked for the queue's owner. |
| json-leak-fix | 1 | none — manual batch 07-15, row already `rejected` | Dead. Nothing to do. |

**And the second half of the defect:** `reviewRevalidators`
(`revalidate_review_queue_action.go:149`) has NO `needs_page` entry — the bug
file's claim that "needs_page IS drainable by the revalidator" is **false**
(corrected in the bug file; WRONG_CALLS entry recorded). So items whose ask is
later satisfied by another route sit for ever: 4 of the 28 target pages are
fully built today and their items still park.

## Design (framework over case, per the architecture seat's own advisory)

The 177 close-out's `architecture` seat flagged: "a THIRD copy of the
satisfiability-mirror would be the moment to extract one shared resolver."
That moment is now — two more emitters need the same question answered.

**One shared, read-only resolver answers "what sections would the handler
see?" at BOTH ends of the item's life:**

- **Emit time** (guard): if the resolver finds nothing AND the page is not in
  the current plan (synthesis-eligible), skip the emit with an observable
  disposition — log + return-map, the 177 pattern. Plan membership counts as
  satisfiable even with zero explicit rows: the handler's fallback 4 can
  synthesise for plan members, so the guard must not out-guess it —
  conservative in the direction of EMITTING.
- **Revalidation time** (drain): a `needs_page` entry in `reviewRevalidators`
  using the same resolver. `resolved` ONLY on positive evidence the ask was
  satisfied (page exists, active, declares sections, and every declared
  section has a built `page_components` slot). Satisfiable-but-unbuilt →
  `still_holds` (with "satisfiable now" in evidence — visible to the queue's
  human). Section-less/ambiguous → `unknown`. This respects the mechanism's
  stated asymmetry (close only on positive evidence; never dispatch) — no new
  verdict vocabulary, no guarantee change, the map IS the designed extension
  point, so normal council gate, not RFC (owner ruling 2026-07-29 #1).

## Edits (≤8, council submission)

1. **New `platform/orchestration/actions/page_section_satisfiability.go`**:
   move + generalise `toolPageDeclaredSections` → unexported
   `declaredPageSections(ctx, dbq, logger, siteID, pageName) ([]string, string)`
   (loader fallbacks 1–3, read-only, first non-empty wins) + new
   `pageInCurrentPlan(ctx, dbq, siteID, pageName) bool` (mirrors fallback 4's
   gate: membership in `site_plan_pages` of the current plan). The OLD symbol
   name is fully removed (it becomes the pod-grep negative control).
2. **`tool_content_item.go`**: `raiseToolContentItem` calls the shared
   resolver; behaviour byte-identical; 177's 8 tests must pass unchanged
   (the proof of a pure extraction).
3. **`flag_page_image_rebuild_action.go`**: guard the needs_page emit —
   resolver finds nothing AND not in current plan → skip with disposition
   `skipped_sectionless_page` in the action's return map + INFO log; the
   `needs_rebuild` flag half of the action is untouched. Replace the
   "VERIFY BEFORE RELYING ON IT" header with the now-verified statement.
4. **`rerender_page_sections_action.go`**: same guard in
   `escalateRerenderToWriter`; skip is logged AND surfaced in the light
   rerender's output (a silent no-op must be observable — the 182 lesson).
5. **`revalidate_review_queue_action.go`**: register `needs_page` →
   `revalidateNeedsPage` using the shared resolver + slot-match query;
   verdicts per the design above. Unit tests alongside the existing
   `revalidate_review_queue_test.go` patterns.
6. **Tests**: sqlmock, mutation-hardened per the 177 lane's practice (each
   guard clause has a test that fails when the clause is inverted; unordered
   mocks — the 177 lane's own two-strike lesson).
7. **Sweep SQL** (`docs/agent_docs/sql_for_agents/NNN_187_sweep.sql`, applied
   post-roll): the measured unsatisfiable-at-birth rows (page resolves no
   sections, no plan membership, or page archived) → `wont_fix`, original
   error preserved in the reason (the 297 precedent — NOT `complete`: no work
   happened, and `complete` releases dependents on a lie). Built/satisfiable
   rows are left for the revalidator run (dry-run first, then apply) so the
   close carries its audit trail. Real-gap reconcile rows LEFT PARKED.
8. **Register + docs, same commit** (ordering-exemption condition 2): concept
   register entry for the shared resolver seam (producer set for `needs_page`
   named per RFC_010 ruling 1 — reconcile_site_plan `needs_page:<name>`,
   page-rerender `needs_page:<name>` co-dedup, image-build-handler
   `page_rerender:<name>`, plus two dead manual sources; landmine: item rows
   carry NULL `page_id` — join pages BY NAME); 016b §9 extension note; 033
   contribution note (needs_page now a covered type); bug-file close-out.

## Explicitly out of scope, named

- reconcile_site_plan's hand-rolled INSERT → `insertWorkItem` migration: same
  "tidy-up, not a defect" call the 177 close-out made for the companion-guide
  emits. Noted in the register entry.
- image-build-handler's `page_rerender:` item_key prefix (breaks the
  co-dedup design `needs_page:<name>`) — noted for the register; changing a
  dedup key mid-flight is its own change with its own blast radius.
- Re-dispatching grip-styles/brands-index/shop-index (satisfiable, unbuilt):
  real work the queue's owner (033) decides on; our change makes their state
  legible (`still_holds` + "satisfiable now" evidence), it does not act.
- Why tool pages' slots carry NULL content_data (the page-rerender trigger):
  if it recurs post-guard it is its own bug; do not fold it in here.

## Verify

- Unit: extraction proven by 177's tests unchanged; new guard tests + map
  entry tests, mutation-checked.
- Pod, both replicas, one exec: positive `declaredPageSections` (added),
  negative `toolPageDeclaredSections` (removed — non-zero on any pre-fix
  image, so the zero is disconfirmable).
- Live: next natural image-landing on a section-less page → skip logged, no
  new row (positive control: a sectioned page still mints). Revalidator
  dry-run names exactly the built rows as resolvable; apply closes them with
  evidence; re-run census — expect only the truthful rows (real gaps +
  satisfiable-pending) to remain parked.
