# NOTES — bugfix 210 (append-only, newest at the bottom)

## 2026-08-08 — pickup, validation, research

- **Ownership checks before pickup**: `who-owns.py 210` → "OWNED or recently active" naming the
  208 lane — that lane FILED 210 and closed itself 2026-08-08 ("Still genuinely open for a
  future session: bugs_open/210" in its handoff). Live-transcript symbol grep
  (`pageSectionShortfall|pageHasComponents|OrdinarySkipStillStamps|ownedPageSkipReasonPrefix`,
  6h window): one session at 38 hits = the 208 lane's own closing session; its final message
  hands 210 off. Clean.
- **Bug still valid at HEAD** (5c138b279): skip refusal scoped to `OWNED_PAGE_GUARD` prefix
  only (`v3_site_actions.go:665-666`); both scope-pinning tests present.
- **MISSTEP (twice)**: wrote SQL against `agent_error_log` assuming `created_at` (it is
  `occurred_at`) and against `orchestration_states` assuming `id` (it is `orchestration_id`) —
  both without `\d` first, the exact CLAUDE.md "schema first" rule. Cost: two error round-trips.
  Cheap check that would have caught it: `\d <table>` before the first query against any table
  this session. → logged in WRONG_CALLS.md.
- **Live-window measurement**: 0 orchestration_states rows with
  `assembled_page.skipped='true'` — recorded as WEAK (terminal snapshot = last iteration only;
  the bug file already rejected the deployed_at-vs-components proxy as confounded). Frequency
  stays unknown; the fix's own error-log row is the counter.
- **Key discovery — the park item must be a RAW insert**: `writeWorkItem`'s two-strike counts
  the prior dishonestly-`complete` build items under `needs_page:<name>` and would brand the
  escalation `unresolved` at birth → terminal → does not hold `idx_swi_dedup`'s slot → bounds
  nothing. `emitOwnedPageReviewItem` is the sanctioned raw-insert template.
- **Drift found in `loadOpenPageItems`**: excludes 5 statuses where the canonical
  `workItemTerminalStatuses` has 7 — `cancelled` and `unresolved` are treated as OPEN by the
  reconciler while the dedup index frees them. `cancelled` contradicts migration 157's stated
  intent; `unresolved`-as-blocking is load-bearing anti-churn (keep, with comment). Split-brain
  scenario that decided it: a human cancels the park → planner path resumes, reconciler wedges
  forever.
- **Third producer on the `needs_page:` key namespace found live**: `needs_tool_recreation`
  items (mortgagecalculator, `needs_page:tool-overpayment` etc., 3 terminal rows/key). Their
  future inserts are blocked while a park is open for the same page — named as a consumer to
  tell (PLAN §consumers).
- **Corroboration**: my three-consumer measurement of `assemble_page` (config text match)
  agrees with LANDMINES.md:6019's nested `jsonb_path_query` walk — differently-shaped checks,
  so this is real agreement, not two-blind-checks-agreeing.

## 2026-08-08 — implementation

- Council submitted BEFORE implementation: corr `c9647117-3a4b-48a2-b34c-1ea25f4e1f7f`
  (5 edits, validated against the fixPlan schema client-side first). Verdict pending at
  commit time → commit carries `Council-Submitted:`.
- Implemented per PLAN: `page_build_failure_guard.go` (refusal/strike/park/auto-close),
  widened branch in `UpdatePageStatusAction` (owned branch untouched above it),
  `loadOpenPageItems` type+cancelled alignment, three tests replacing the scope pin.
- **All three mutations killed their named tests** (M1 skip-branch disabled →
  OrdinarySkipRefusesStamp failed on the unexpected stamp UPDATE; M2 park call removed →
  ThirdRefusalParksThePage failed on unmet park-INSERT expectation; M3 auto-close removed →
  SuccessfulStampClosesPark failed on unmet close-UPDATE expectation). Reverted; full
  `go test ./platform/orchestration/...` green, exit status checked via PIPESTATUS
  (the `| head && echo OK` landmine appended to LANDMINES the same day by another lane).
- Register: PBP-038 added; PBP-036's now-stale "keyed to THIS skip" bullet corrected
  visibly (strike-through + date), per the stale-status landmine.
- LANDMINES: two entries appended + `landmines-sync.py --apply` run (1350 rows owned).
- Consumers told by append-only notes: mortgagecalculator lane (ACTIVE session — used
  `cat >>` append rather than a whole-file Write to be collision-safe), feature_021 lane,
  208 lane handoff pointer.

## 2026-08-08 — post-commit

- Commit `2c3efc9f5`, 14 files, commit-scope report clean. Three shared files deliberately
  LEFT OUT (they carried other live sessions' appends): LANDMINES.md — my two entries were
  already swept to HEAD inside `f993554f6` (RSH-008 lane), nothing lost; WRONG_CALLS.md and
  the mortgagecalculator NOTES — my appends ride with those lanes' next commits.
- **Pattern-check advisory answered — the untouched twin `UpdatePageComponentsStatusAction`
  does NOT share the defect**: it stamps `page_components` review status (default 'approved')
  in the review flow, never writes `pages.build_status` or `built_from_plan_version`, and is
  not a step in the three assemble loops — so it cannot silence the reconciler, which is the
  whole of bug 210's mechanism. No change made there, deliberately.
- The other advisory (logged-model-output at v3_site_actions.go:5750) predates this lane's
  edits and belongs to whoever owns that call site — not smuggling a fix into this commit.
- Council run live at commit time (review_architecture executing ~17:34 UTC); verdict below
  when read.

## 2026-08-08 — council verdict READ: APPROVED, round 1, 4 advisory objections (none high)

Corr `c9647117`. Verdict `approved` at 17:36 UTC, ~10 minutes after submission (no dispatch
queueing this time). Every objection read; dispositions, each with its evidence:

1. **editquality (medium): "no edit touches the concept register."** True of the submission's
   edit list (platform files only); refuted by the COMMIT — `2c3efc9f5` carries PBP-038 +
   the index row, same commit as the code, which is the ruling's actual condition.
2. **guidelines (medium): raw ON CONFLICT vs the DELETE+INSERT contract.** Answered by
   induction THIS session, same design as 208's: in a rolled-back tx, my exact park INSERT
   gave `INSERT 0 1` then `INSERT 0 0` (deduped while open, bare conflict — no 42P10), then
   after driving the row terminal a re-insert succeeded (correct re-arm), 0 rows after
   ROLLBACK. The dedup-only intent needs no DELETE+INSERT: nothing is ever updated on
   conflict, and the open row WINNING is the design.
3. **guidelines (low): 'unresolved' left out of the reconciler's closed set.** Deliberate and
   comment-carried (code + register): freeing it re-emits every two-strike-parked page
   through a raw INSERT with no two-strike. Their alternative (check the park before treating
   unresolved as clear) adds a second query for the same behaviour.
4. **tooling_provenance (medium): no doc_notes trail.** The two LANDMINES entries were synced
   to doc_notes BEFORE the commit (`landmines-sync.py --apply`, 1350 rows), footprinted on
   `page_build_failure_guard.go`, `insertWorkItem`, `page_build_failed` and the `needs_page:`
   keys — that is the doc_notes mechanism the seat asked for.
5. **guardian (medium): can the tool-recreation lane tell "my dedup" from "someone's park"?**
   Yes — the discriminating query is in their NOTES and in LANDMINES: an open
   `page_build_failed` row at `needs_human_review` on the key IS the park, and its spec now
   carries `bug: bugs_open/210` (follow-up commit) per the mistyped_deployed_page convention.
6. **guardian (low): cancel-as-mute on the two older types.** MEASURED, and the seat was right
   to ask: 25 cancelled `needs_page` + 1 `owned_page_review` rows exist. Decomposition:
   14 sit on `deployed`+current pages (decideEmit `skip_built` — no effect), 3 are synthetic
   verify-keys matching no plan page (no effect), 1–2 produce cheap review items —
   **8 dartsonline pages (planned/needs_rebuild, cancelled 07-20) would genuinely re-emit
   LLM builds on that site's next reconcile.** Surfaced to the owner in README; the durable
   mute verb is `wont_fix`/`rejected` (closed in BOTH mechanisms), and re-statusing those 8
   is an operator call, not this lane's.
7. **debug_historian (medium): does an item-type/handler parity script choke?** No such
   script exists (grepped); the no-handler decision-item shape is established, deliberate
   precedent — the `mistyped_deployed_page` LANDMINE says "no handler by design". Its
   column trap (handler_agent is NOT NULL DEFAULT '', so IS NULL finds nothing) applies to
   the park row too and is why the spec carries the bug pointer.
8. **debug_historian (low): needs_rebuild selectors over-firing on parked pages.** Slot-keyed
   automatic producers are blocked by the park; the remaining selectors are demand-driven and
   every retry passes back through the guard (counted, bounded). The error-log counter is the
   watch instrument.
9. **prior_art_librarian (medium/low): blast-radius and 0-rows claims "unverified this
   session".** They WERE run this session against `agent_definitions` (both the action match
   and the `output_field` match — three agents) and `site_work_items` (0 `page_build_failed`
   rows); the grounded_in block said so. No further action.
10. **architecture (medium): size park inflow vs queue drain.** Post-roll watch item — the
    refusal counter is the sizing instrument (RUNBOOK query); expected inflow is bounded by
    bugs_open/202's quota-failure rate, and the seat itself says "not before merge".

**Third schema-first miss today** (`diagnosis_artifacts.content` — it is `body`) — after
writing the WRONG_CALLS entry about the first two. The tally is the point; entry updated.

## 2026-08-08 (later) — LIVE on v1.0.1268, pod-verified; and a correction

- **Fleet rolled to v1.0.1268.** All 12 chassis containers share ONE image digest
  (`4c08b8d8d7d6…`); 5 pods greped before the exec loop timed out, every one `1 1 0 3`:
  `DEPLOY_STAMP_REFUSED_ON_SKIP` 1 (commit 2c3efc9f5), `bugs_open/210` 1 (follow-up
  8d95779a2 — proves the roll is at/after the SECOND commit), fabricated
  `DEPLOY_STAMP_ZZFAKE_210` 0 (the grep can return zero), `OWNED_PAGE_GUARD` 3 (positive
  control). Digest identity generalises the 5 to all 12, the 208 practice.
- Baseline at verification time: **0 refusals, 0 parks** — correct this early; the counter
  is armed and starts measuring the bug's real frequency from now.
- > **CORRECTED 2026-08-08:** the unmute population is **SEVEN** dartsonline pages, not the
  > "8" published in yesterday's NOTES/README/bug file/commit message. The original query
  > output listed seven; the 8 was a transcription error at the summarising step, caught by
  > re-running the audit post-roll. WRONG_CALLS entry added. The seven: brands, brands-index,
  > grip-styles, guides, product-detail, shop, shop-index.
- `vonc.com/provocation` verified `rebuild_policy='owned'` (plan role blog-post) — its
  release produces ONE owned_page_review row via the reconciler's ownership branch, no LLM
  build. The single cancelled `owned_page_review` is oufe's `zz-canary-208` cleanup row —
  the page is in no plan, so releasing it does nothing.

## 2026-08-09 — owner decisions taken; the 7 re-muted (UPDATE 7, row identity verified)

- **Watch checked before anything else**: 0 refusals (`DEPLOY_STAMP_REFUSED_ON_SKIP`),
  0 parks (`page_build_failed`) — [MEASURED] ~09:45 UTC, ~16h post-roll. No first
  measurement yet; this is also the architecture seat's park-inflow-vs-drain sizing (both 0).
- **Owner decisions (AskUserQuestion this session):** (1) the 07-20 mute STANDS — re-mark
  the 7 as `wont_fix`; (2) the behavioural canary is AUTHORISED (live dispatch).
- **The audit returned EIGHT rows on re-run, not seven** — the snapshot went stale exactly
  as the RUNBOOK warns. The 8th is `new-arrivals`: its PAGE flipped to `needs_rebuild` at
  08:58:41 UTC today, flipped by whoever is actively driving dartsonline right now (a
  full-site `page_rerender` wave filed 09:03 UTC, pages deploying as I watched: contact
  09:29, about 09:25, index 09:16; `sale` flipped in the same second, and `sale` has no
  cancelled item so it never enters the audit). **Not the guard** — refusal count was 0.
  **Deliberately left alone**: not in the approved seven, actively owned by another lane,
  and a `wont_fix` on its 07-20 row would not stop a direct build anyway (the mute only
  blocks reconcile re-emission).
- **Checked before updating: reconcile had NOT yet fired** — zero new `needs_page` rows on
  dartsonline in 24h (the wave is `page_rerender` + one `chrome_divergence_overwritten`,
  different item_type, different keys, no collision with the 07-20 rows). The door closed
  before any replan tripped it.
- **The UPDATE**: 7 rows by id, `status='wont_fix'`, provenance merged into `result`
  (`remute: owner decision 2026-08-09 …`), `RETURNING` listed exactly brands, brands-index,
  grip-styles, guides, product-detail, shop, shop-index — row identity, not just `UPDATE 7`.
- **Verified by audit re-run**: the 7 are gone; remaining rows are `vonc.com/provocation`
  (left released per the owner decision's framing — one review item, no build),
  `new-arrivals` (above), and synthetic `verify*`/`page_rerender:*` keys matching no plan
  page (webdesign.uk rows newly present from other lanes' churn — same no-effect class).

## 2026-08-09 — the canary was AUTHORISED and then STOOD DOWN: the branch is not inducible

The owner authorised the live canary. Investigating the lever before firing (plan step 2a)
established it cannot be built as specified, so it was stood down with the owner's agreement.
The reasoning, because "we skipped the canary" is worthless without it:

- **[MEASURED 2026-08-09] There are exactly THREE doors into `assemble_page` fleet-wide, and
  all three are the same conditional.** Census over every active, non-snapshot definition —
  every step whose `next_step`/`then_step`/`else_step` is `assemble_page`:
  `site-work-orchestrator.check_review_approved` (then→assemble, else→`fail_item`),
  `pageflow-builder.check_review_approved` (else→`complete_page`),
  `page-rebuild.check_review_approved` (else→`complete_page`). No workflow and no loop starts
  there either (`page-rebuild` loop starts `plan_sections`, `pageflow-builder` at
  `write_page_content`, `site-work-orchestrator` at `spawn_handler`). Query in the RUNBOOK.
  Disconfirmable: a fourth door, or a loop entering at `assemble_page`, would have shown up.
- **Therefore the ordinary-skip arm is reachable only when the LLM reviewer APPROVES a failed
  or empty payload.** Both of the guard's triggers live *inside* `assemble_page`
  (`multipage_actions.go:91` upstream-failure, `:107` empty content) — downstream of
  `check_review_approved`. Any content failure I can induce deterministically diverts at that
  gate to `complete_page`/`fail_item` and never reaches `update_page_status` at all. A canary
  built on an induced failure would assert its own diversion, not the guard.
- **What that leaves is not nothing.** 208's canary already drove `assembled_page.skipped`
  into `UpdatePageStatusAction` through the real pipeline — the same `collected_data` key, the
  same shared entry (`v3_site_actions.go:667`, `upstreamAssemblySkipped`). The plumbing "a real
  run's skip flag reaches this guard" is behaviourally proven; 210 changed the branch taken
  *after* that entry. Unit tests cover the branch, each mutation-proven, plus the live DB
  induction of the park insert. **The uncovered gap is narrow and should be stated narrowly:
  no production run has yet been observed taking the non-owned arm.**
- **Rejected: a throwaway `agent_definitions` row** copying the loop from `assemble_page`
  onward with a deterministic failing step. It would be end-to-end through the real actions but
  no longer the real workflow, and it puts scratch config on a shared fleet mid-flight for a
  branch three other instruments already cover. Owner agreed.

### The measurement I ran that could not have come out otherwise (WRONG_CALLS entry added)

Trying to answer "has this branch ever fired historically", I queried
`orchestration_states.collected_data->'assembled_page'->>'skipped'` over all 4,457 rows
(2026-07-13 → today) and got **0**, with only ONE row carrying the key at all. That "0" is
worthless and I nearly wrote it down as evidence of rarity. Enumerating the keys instead of
path-reading them (`jsonb_object_keys` over `owner_agent_type='page-rebuild'`) shows why:
collected_data retains **only pre-loop steps** — `get_pages_to_rebuild`, `load_rebuild_context`,
`select_style_collection`, `site_record`, `pages_to_build` — and **no per-iteration loop state
whatsoever**. No `assembled_page`, no `page_content`, no `update_page_status`. A recursive
`$.**.assembled_page` walk finds it in `pageflow-builder` rows only (9, none skipped).
**So `orchestration_states` cannot answer the historical-frequency question for the two loops
that matter, and the refusal counter remains the only instrument** — which is what the bug file
already said, now with the reason. [MEASURED] the estate's own landmine ("a jsonb PATH read
cannot see the shape change underneath it — enumerate keys") caught this one exactly.

### A scope boundary nobody had written down: the guard does NOT cover `page-rerender`

Checked while establishing which paths reach the guard, and worth recording because the lane
docs' only mention of rerender is the confounded-proxy discussion:

- `upstreamAssemblySkipped` reads **`collected_data["assembled_page"]` and nothing else**
  (`owned_page_guard.go:308-319`). `page-rerender` renders via `rerender_single_page` into
  **`rendered_page`**, so a skipped rerender is invisible to the 210 guard — its
  `update_page_status` step (status `deployed`) would stamp normally.
- It does not, because that workflow carries its own **config-level** guard:
  `render_page → check_skipped` (`rendered_page.skipped == true` → `complete_skipped`, else
  `deploy_page`). Verified live 2026-08-09; `rerender_single_page_action.go:198-209` is the
  producer of that flag (`skipped:true`, `html:""`, on "no component rows").
- **No bug today** — the conditional is present and correct. But the rerender path's only
  protection against exactly bug 210's outcome is a workflow conditional, i.e. the fragility
  class 210 was filed to remove from the build path in code. Anyone who edits that workflow,
  or who reads PBP-038 as "the stamp is now refused on any skip", is one deletion away from
  reproducing 210 on the rerender path with no code guard underneath. LANDMINES entry added;
  PBP-038's scope line now states it.
