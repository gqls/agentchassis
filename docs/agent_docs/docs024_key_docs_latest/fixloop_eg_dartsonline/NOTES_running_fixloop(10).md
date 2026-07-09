# RUNNING NOTES — Diagnosis→Fix Loop (v1)

Chronological; newest entries appended under DISCUSSION LOG; decisions
promoted to DECISIONS with rationale. Continues NOTES_running_fixloop(9).md
(entries up to 2026-07-07 are there and are not repeated).

## 2026-07-09 — new thread opened; pilot pre-check run FIRST

### Turn 1 — context loaded, pre-checks executed, symptom did not survive intact

Inherited state read: RUNBOOK_diagnosis_fix_loop(9).md, NOTES_running_fixloop(9).md,
HANDOFF_fixloop_thread(8).md, z_bundles/BUNDLE_fixloop_F0.md.

**Bundle confirmed deficient as the handoff warned.** BUNDLE_fixloop_F0.md
(199,579B) carries `## Schema — _none provided_`, `## Database capabilities —
_none provided_`, `## Runtime evidence — _not available in the thin slice_`.
It was generated without `-psql`, so it is code+docs only. For this pilot the
DB half is where the answer lives. Regenerate before any loop run that is
supposed to consume it.

**The three ★ pre-check queries were run against the live cluster** (psql via
`kubectl exec -n ai-persona-system postgres-clients-0`). One needed
correction: `site_work_items` has no `attempts` column — it is `attempt_count`.
The runbook's pre-check SQL is wrong on that column and is corrected in
RUNBOOK(10).

**The pre-check did not sharpen the symptom. It dissolved it, and then opened
a bigger one.** Findings, in the order they arrived:

1. dartsonline has exactly **one** guide-ish page row: `guides-index`,
   `page_type='section-index'`, `build_status='planned'`, `in_header=true`.
   There are **no** `page_type='guide'` rows at all.
2. gamesdesign.co.uk has 5 × `guide` + 1 × `guides-index` (`section-index`),
   **all `deployed`**. The differential is real.
3. Widening to the whole site: dartsonline's page matrix is
   `content` (3) and `landing` (2) → **deployed**; `blog-post` (4),
   `entity-directory` (2), `entity-page` (2), `section-index` (1), `tool` (1)
   → **all `planned`**. So this was never a guides bug. **Ten of fifteen
   pages were never built**, and guides is simply the one that got a nav link.
4. `blog-post` is *in* the hypothesised routing table and still was not built
   → the standing hypothesis was already in trouble at this point.
5. The `needs_page` work items for those pages **all exist and all completed**
   (`status='complete'`, `handler_agent='page-build-handler'`,
   `attempt_count=0`). Twenty-three of them. The system marked as complete the
   construction of pages it never constructed.
6. The `result` payloads split the set cleanly. The 5 deployed pages carry
   `{"response":{"deploy_result":{... "files":["/about.html"] ...}}}`.
   The 10 unbuilt pages carry `{"response":{"site_record":{...}}}` or a bare
   design-tokens blob (`{"spacing":…,"typography":…}`) — **no `deploy_result`
   field at all**.
7. `pages.sections` is the discriminator, and it is an **exact partition, 5 v 10,
   no exceptions**: `jsonb_array_length(sections) > 0` ⟺ `deployed`;
   `= 0` ⟺ `planned`. gamesdesign's guide and section-index pages each carry
   `sections` (2 apiece) and are deployed — **through the same
   `page-build-handler`**. The handler is not the discriminator. Sections are.
8. A name mismatch sits alongside it: the imagery flow emitted
   `page_rerender:guide-barrel-weight` … `guide-steel-tip-vs-soft-tip` items
   (slugs derived from hero assets `hero_guide_barrel_weight` etc.), but the
   plan and `pages` name those rows `barrel-weight`, `beginners`,
   `flight-shapes`, `steel-tip-vs-soft-tip`. Those rerenders targeted pages
   that do not exist under that name, no-op'd, and completed.

### Turn 1 (cont.) — mechanism found in code; standing hypothesis REFUTED

Read `reconcile_site_plan_action.go`, `load_work_item_actions.go`,
`populate_nav_tables_action.go`, and the **live** `page-build-handler`
workflow JSON out of `agent_definitions`.

**Cause B — the silent success.** The live `page-build-handler` workflow
(v1, active) runs `plan_sections` → `check_has_ready_sections`, a `conditional`
with `condition: "section_plan.ready_count > 0"` and
`else_step: "complete_error"`. `complete_error` is **`action: complete_workflow`**
— a *success* terminal — with
`success_message: "Content writer skipped — page has no sections defined"` and
`output_fields: ["page_content","site_record"]`. That output_fields list is
*exactly* the shape observed in the 10 unbuilt items' `result` (finding 6).
The dispatch loop then stamps `status='complete'`. The page row is never
touched: `updated_at == created_at` (12:39:05) on all ten.

**The platform already knows.** `load_work_item_actions.go:750-756`, a comment
on the completion guard, says in as many words: *"the dispatch loop calls
complete_work_item on every successful handler saga, and page-build-handler's
complete_error is a SUCCESS-labelled complete_workflow"* — and names the
remedy: *"mark_no_sections for a sectionless page with no sibling layout"*
would flag the item `needs_human_review`. **`mark_no_sections` does not exist.**
It is absent from the live workflow's 18 steps and appears nowhere in the repo
except that one comment (`grep -rn mark_no_sections` → 1 hit, the comment).
The guard faithfully preserves a flag that nothing ever sets.

**Cause C — nav grounded in the wrong column.**
`populate_nav_tables_action.go:242-243` selects
`FROM pages WHERE site_id = $1 AND status IN ('active','deployed','pending')`.
`pages.status` is a *lifecycle* column defaulting to `'active'` on insert.
`build_status` — the actual build state — is **never consulted**. So
`guides-index` (`build_status='planned'`, `status='active'`, `in_header=true`,
`page_type='section-index'`, which is not in the `neverPrimaryTypes` set
`{blog-post, tool, entity-page}`) is published straight into the primary nav.
That, precisely, is "the system linked to something it never built".

**Cause A — the planner under-populates sections.** `build-site-planner` wrote
15 plan pages into `site_plan_pages` but authored `sections` for only the 5
`content`/`landing` ones. Everything downstream is a consequence.

**THE STANDING HYPOTHESIS IS REFUTED — and it named the wrong file.** The
routing table is real but it is in **`WriteBuildItemsAction`**
(`load_work_item_actions.go:218-228`), *not* in `reconcile_site_plan`.
`guide` and `section-index` are indeed absent from it and `tool`/`entity-page`
are indeed commented out. But absence from that map **does not drop a page**:
`:239` defaults `handlerAgent = "page-build-handler"`, and `:283` logs
`"Unknown page_type, using page-build-handler"` before falling through. What
*does* drop pages is the separate `unavailableBuilders` map (`:233-237` —
`tool`, `entity-directory`, `entity-page`) which hits `continue // Skip — don't
create a dispatch work item`. And `reconcile_site_plan_action.go:213-217`
hardcodes `handler_agent='page-build-handler'` for **every** plan page with no
type switch whatsoever — so it has no `unavailableBuilders` guard and re-emits
items for the very types the other path deliberately skips.

So the guides nav link is caused by **B + C**, not by the routing table. The
hypothesis was directionally right that "nav outlives a dropped page" and
wrong about every mechanism it named. Evidence over assumption, exactly as the
intake instructed.

**Fourth finding (new, unlooked-for):** the two intake paths disagree.
`WriteBuildItemsAction` skips `tool`/`entity-directory`/`entity-page`;
`reconcile_site_plan` emits `needs_page` for them anyway. dartsonline's
`shop-index`/`brands-index` (`entity-directory`) have completed `needs_page`
items that could never have built. gamesdesign's items are `needs_content_page`
(the `availableBuilders` path); dartsonline's are `needs_page` (reconcile).
**That is the build-route variable the runbook told us to establish by
evidence rather than assume. Established.**

### Turn 1 — consequence for the pilot (a judgement the owner must rule on)

Loop-worthiness criterion 3 says a task is loop material only if it is *not*
answerable by one or two direct queries. This took roughly a dozen queries and
four file reads — more than two, but it never required a hypothesis-driven
re-scoping loop. **Candidate 3 has now gone the way of candidates 1 and 2: the
cheap pre-check answered it.** Three for three. That is not a failure of pilot
selection; it is the pre-check doctrine doing its job, and it says something
structural about this platform — its bugs are mostly *legible* to someone with
schema access and grep.

But the conclusion this time is the opposite of "reject the pilot", because
**we now hold ground truth with citations across all three tiers.** A loop
cannot be graded on a bug whose answer nobody knows. It can be graded exactly
here. RECOMMENDATION: promote this from *discovery* pilot to **known-answer
benchmark** — the loop runs on the original symptom string, blind, and we score
its output against the findings recorded above. See PLAN_fixloop_pilot.md §3.

## DECISIONS (with rationale)

### 2026-07-09 — pilot reframed as a known-answer benchmark (proposed, owner to confirm)
- The dartsonline guides symptom is fully diagnosed by hand, with static /
  live-data / runtime citations. Running the loop on it can no longer *discover*
  anything; it can **verify the loop**. That is more valuable at F0 than a
  discovery run, because F0's five criteria are all about the plumbing
  (intake, fetchable bundles, per-iteration notes, a cited mechanism) and only
  the fourth is about the answer — which we can now mark objectively.
- Scoring rubric pre-registered in PLAN_fixloop_pilot.md §3 so the grading
  cannot drift to fit whatever the loop happens to emit.

### 2026-07-09 — the bug itself is a platform defect, not a dartsonline defect
- Causes B and C are relay-level (a workflow definition; a nav action). Fixing
  dartsonline's plan fixes one site. Same shape as the roadmap gap that became
  builder item 6. The fix belongs in the platform; the F1 edit plan should
  target `check_has_ready_sections`/`complete_error` and
  `populate_nav_tables_action.go`, not the site's data.

### Carried forward unchanged from (9)
Q-A diagnosis_artifacts table (kind ∈ bundle|iteration_note), write-through
inside assemble. Q-B intake = `needs_diagnosis` item, `pipeline='diagnose'`,
null-site allowed. Q-C separate fixer agent, isolated write token, constrained
edit plan, gofmt+build gate. Q-D flag-based hard_veto; guideline-gap =
side-task. Q-F own working-notes storage, terminal note only into the tools
chat's doc_notes. Q-E/Q-G/Q-H remain open (F2).

## OPEN QUESTIONS RAISED THIS TURN
- Why does `/guides/index.html` render **blank** rather than 404, given no
  build ever ran? Something is serving a shell. Unresolved; low priority for
  the diagnosis, relevant to the fix's verification step.
- Should `reconcile_site_plan` grow the `unavailableBuilders` guard, or should
  `WriteBuildItemsAction` lose it? The two paths must agree; which way is a
  design decision for the builder thread.
- Was `mark_no_sections` ever written and removed, or only ever intended?
  Git archaeology would settle it and would feed F3's learning record.
