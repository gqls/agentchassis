# PLAN — bugs_open/210: a content-failed page build is stamped `deployed`, and the rebuild request is forgotten

**Lane opened 2026-08-08.** Owner thread: this one. Bug file:
`bugs_open/210_HANDOFF_2026-08-06_a_content_failed_page_build_is_stamped_deployed_and_the_rebuild_request_is_forgotten.md`
(filed by the 208 lane as the follow-up its fix deliberately excluded; 208 closed 2026-08-08,
this bug confirmed unowned before pickup — who-owns + live-transcript symbol grep, both clean).

## The defect (verified at HEAD 5c138b279, 2026-08-08)

`UpdatePageStatusAction` (`v3_site_actions.go:664-677`) refuses the `deployed` stamp after an
assembly skip **only when the skip carries the `OWNED_PAGE_GUARD` prefix** (bugs_open/208's
guard). An ordinary skip — content generation failed, or content empty — passes through:
`pageHasComponents` and `pageSectionShortfall` are satisfied by the *previous* build's output,
so the page is stamped `build_status='deployed'`, `deployed_at=NOW()`,
`built_from_plan_version=<current plan>`. The rebuild request is forgotten: `decideEmit`
returns `skip_built` forever after, and the page serves stale content while the fleet believes
it is current.

## Decisions and their reasons

1. **Fix candidate 1 from the bug file** (refuse the stamp on any assembly skip + explicit,
   bounded requeue), because it is the only candidate that answers the retry-loop objection
   instead of inheriting it (candidate 3) or adding a status-vocabulary value that every
   `build_status IN (...)` enumeration must learn (candidate 2).
2. **The retry bound is a three-strikes rule implemented at the refusal site**, counted from
   `agent_error_log` refusal rows for the page since its last successful deploy (7-day cap).
   Strikes 1–2: flip to `needs_rebuild` + NULL the plan stamp (the honest state; the existing
   producers re-emit and retry). Strike 3: additionally park the page behind an open
   `page_build_failed` work item, `status='needs_human_review'`, **no handler**, with
   **`item_key='needs_page:<name>'`** — the *shared page-slot key* — so `idx_swi_dedup`
   blocks every automatic producer of that slot until a human closes it.
   Why `agent_error_log` and not terminal work items as a counter: a `complete` work item
   cannot be distinguished from a genuinely successful build (that indistinguishability IS
   bug 210), and the error-log row doubles as the permanent measurement signal the bug file's
   "How to measure" section asks for (candidate 2 there).
3. **The escalation insert is RAW SQL** (the `emitOwnedPageReviewItem` pattern), deliberately
   NOT `insertWorkItem`: the two-strike block in `writeWorkItem` counts the prior (dishonestly
   `complete`) build items under the same key and would brand the escalation `unresolved` at
   birth — terminal, so it would never hold the dedup slot and the design would silently not
   bound anything.
4. **On a successful `deployed` stamp, auto-close any open `page_build_failed` item** for the
   page. Success is the definitive evidence the parked condition is resolved; without this the
   park outlives its truth (the a-handoff-outlives-the-work class).
5. **`loadOpenPageItems` learns the new type and stops treating `cancelled` as open.**
   Adding `'page_build_failed'` to its type filter makes the reconciler's `skippedQueued`
   counter honest (otherwise it would count a blocked page as "emitted" while the bare
   `ON CONFLICT DO NOTHING` silently dropped the insert — the 091 dishonesty family).
   `'cancelled'` joins its closed set because migration 157 already ruled a cancelled row must
   not hold the dedup slot, and today the reconciler alone still honours it — a human
   cancelling the park would free the planner path but wedge the reconciler path forever.
   **`'unresolved'` deliberately stays blocking** for the reconciler: freeing it would make the
   reconciler re-emit every two-strike-parked page on every run through its raw INSERT (no
   two-strike), reintroducing the exact loop this fix closes.
6. **Scope pins updated, not deleted**: `TestUpdatePageStatus_OrdinarySkipStillStamps` was
   written to make widening a decision rather than a side effect — it becomes the test of the
   new refusal. `TestSavePageSections_OrdinarySkipIsNotClaimed` is UNCHANGED: the save path
   still writes sections metadata on an ordinary skip (content_data is what a later re-render
   regenerates from; stopping those writes is not part of this fix).

## Measured facts the plan rests on (all 2026-08-08, live cluster + HEAD)

- Exactly **three live agents** produce `assembled_page`: `pageflow-builder`, `page-rebuild`,
  `site-work-orchestrator` (config text match; independently corroborated by LANDMINES.md:6019's
  nested `jsonb_path_query` walk — two differently-shaped checks agree). The six other agents
  running `update_page_status` never carry the key → the widened guard is a no-op for them.
- `idx_swi_dedup` = `(site_id, item_key) WHERE status NOT IN (7 terminal statuses)` — matches
  `workItemTerminalStatuses` exactly. `loadOpenPageItems` excludes only 5 (drift: treats
  `unresolved`, `cancelled` as open).
- The `needs_page:<name>` key namespace already has **three producers**:
  `ReconcileSitePlanAction` (item_type `needs_page`, raw INSERT), `WriteBuildItemsAction`
  (item_type `needs_content_page`, via `insertWorkItem`, two-strike active), and the
  tool-recreation lane (item_type `needs_tool_recreation`, observed live on
  mortgagecalculator.co.uk, 3 terminal rows per key).
- `item_type='page_build_failed'` count in live `site_work_items`: **0** (name is free).
- `agent_error_log` retains from 2026-07-09 (~1 month) — long enough for a 7-day strike window.
  Writer: `agenterrors.Write` (RFC_012 leaf package, the ONE writer).
- No live instance of the 210 shape in the current ~24h `orchestration_states` window
  (`collected_data->'assembled_page'->>'skipped'='true'` → 0 rows). **[WEAK — terminal states
  hold only the LAST loop iteration's snapshot, so a mid-loop skip followed by a later page's
  success is invisible to this query.]** Frequency remains unmeasured until the fix's own
  error-log rows start counting it — which is decision 2's second purpose.
- Retry pacing if unbounded: `build-pipeline-trigger` fires every 120 s; reconcile runs inside
  `build-site-planner` (demand-driven). So the unbounded loop is slow, real, and silent.

## Consumers to TELL (owner ruling 2026-07-29 §3)

- **feature_021 operator bulk page rebuild** — their entry point's guarantee changes: a bulk
  rebuild whose content step fails now leaves `needs_rebuild` (and can park after 3 failures)
  instead of silently stamping `deployed`. Note goes in their NOTES file.
- **mortgagecalculator tool-recreation lane** — an OPEN `page_build_failed` park under
  `needs_page:<name>` blocks a `needs_tool_recreation` insert for the same page slot, and
  `insertWorkItem`'s false return reads as "already covered" (LANDMINES :1440 family). The park
  is `needs_human_review` + visible, and a successful deploy auto-closes it; but the lane
  should know the block exists.
- **208 lane** (closed) — its handoff names 210 as the successor; close-out note added there.

## Edits (≤8, the council-submission shape)

1. `platform/orchestration/actions/page_build_failure_guard.go` — NEW: refusal helper
   (error-log row → needs_rebuild flip → strike count → raw park insert), auto-close helper,
   constants. Doc header carries decisions 2–4.
2. `platform/orchestration/actions/v3_site_actions.go` — widen the skip branch in
   `UpdatePageStatusAction` (owned-prefix branch unchanged above it); call auto-close after a
   successful deployed stamp.
3. `platform/orchestration/actions/reconcile_site_plan_action.go` — `loadOpenPageItems`: add
   `'page_build_failed'` to the type filter; add `'cancelled'` to the closed set; comment why
   `'unresolved'` stays.
4. `platform/orchestration/actions/owned_page_guard_test.go` — repurpose
   `TestUpdatePageStatus_OrdinarySkipStillStamps` → refusal test; add strike-3 park test and
   auto-close test. (Mutate-to-prove during dev: invert the guard, watch each fail.)
5. `docs/agent_docs/docs026_concept_register/register/page-build-pipeline.md` — register the
   `needs_page:<name>` shared page-slot key (producer set + shapes) and the `page_build_failed`
   type, same commit as the seam (CLAUDE.md condition 2).
6. `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md` — the "raw insert or the park is
   stillborn" trap + the shared-slot block on tool-recreation inserts; run landmines-sync.
7. `bugs_open/210_...md` — status update: fix built, decisions, what proves it live.
8. Consumer notes (feature_021, mortgagecalculator lane docs).

## Not in scope

- Making the failing workflows FAIL their work items (the `complete`-on-failure dishonesty is
  bugs_open/099's family; this fix makes the page state honest without re-plumbing workflow
  error semantics).
- bugs_closed/037's boundary (a replan may take a needs_rebuild page's composition) — our fix
  moves pages INTO that state honestly; the 037 question is unchanged and stays with 037.
- Any change to `SavePageSectionsAction` scope (see decision 6).
