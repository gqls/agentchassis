# HANDOFF — brochure component library / fundamentallyai.com — 2026-07-28b (afternoon)

**Cold-start document, supersedes `HANDOFF_2026-07-28_continue_here.md`** (read that one
next anyway — its §3 cap story, §4b owner rule, §4c–4e design directions and landmines
L11–L17 all still stand; only its §4 action list is superseded here).

## What changed this afternoon (evidence: NOTES 07-28 afternoon entry)

- **§4.1–4.4 of the morning list are DONE.** Canary complete; capabilities chart correct
  (2 charts, 085 held); link crawl run; **both `needs_imagery` items complete with stored
  assets**. The llm-cost-calculator hero is live and serving
  (`/assets/images/content-hero-llm-cost-calculator.jpg`, 200, no new broken refs).
- **`bugs_open/079` is REOPENED — the link repair's output is DISCARDED at save.**
  The gate repaired all 9 invented links on the 10:05Z build; `save_page_sections`
  persisted the unrepaired sections 400ms later (structured `sections_metadata` path
  wins; `clean_html` is fallback-only). Full mechanism + fix candidates in the 079
  REOPENED banner. Consequence: **there is NO downstream mitigation for invented links
  or srcs — 092 (writer never gets constraints) is now the live front line.**
- **`/assets/illustrations/` is pure invention** (1 component fleet-wide — the regressed
  carousel). The §4b regression is one defect with the links, filed into 071/092.
- **Anthropic spend cap EXHAUSTED until 08-01 00:00 UTC** — councils + diagnosis loop
  die on `provider=anthropic`; failures show `COMPLETED` with the refusal in
  `__step_error` (099 family). Gemini text/image (this site's lanes) work fine.

## Constants

Unchanged — see the morning handoff §1. DB access, site/plan/page ids, `.html`-only hrefs.

## Next actions, in order

> **EVENING SESSION 2026-07-28 (same thread, post-cap-raise): items 1, 2 and 4 are DONE;
> item 5's §4d is done-and-corrected (see item 5).** The Anthropic cap was raised by the
> owner ~14:50 BST — the "after 08-01" framing below is stale. What happened, with
> evidence in NOTES evening entries and `bugs_open/079`:
> - **Item 1 DONE.** The selector's generated file was ALREADY published (only the
>   reference was missing; tool pages here are `generic`, not 'owned' — the no-op is the
>   EMPTY `pages.sections`). Hero placed by scoped edit of the single component's
>   `rendered_html` + assemble-only 049b rerender; served page verified, image itself
>   eyeballed. `56fbcc9a` completed.
> - **Item 2 DONE — it was one revert, not thirteen repairs.** Both corrupted components'
>   pre-regression `content_data` was sitting in `page_component_history` (displaced at
>   the 10:45 overwrite): real jpgs, real fragment links. Restored both (corrupted state
>   archived, source `operator_restore_pre_regression_2026-07-28`), re-rendered LLM-free
>   (`section_data_resolved` via a `page_rerender` work item — **the INSERT template
>   needs `handler_agent='page-rerender'` or it hard-blocks**; cta NOTES corrected).
>   Served: 0 phantoms, 0 invented svgs, charts intact. Residuals, all pre-existing:
>   favicon 404; the fragment links' anchors don't exist (no `id=` on the page, true of
>   the last-good state too); `icon-cap-review.jpg` flaps 200/404/200.
> - **Item 4 DONE early.** 090 verification fired 16:38Z (branch had to be PUSHED first —
>   the loop clones from origin), verdict **UNVERIFIABLE, no refutation**; chasing its one
>   new citation yielded a third corroborating site (gamesdesign bayesian-ranking:
>   repair recomputed today, store unchanged since 07-21). Full addendum in
>   `bugs_open/079`, which also names the standing reproductions now that capabilities
>   is clean. Verdict location: child orchestration `collected_data->'verdict'`, NOT
>   diagnosis_artifacts. Intake auto-closed by the enabled `diagnose-pipeline-trigger`.
> - **Still open for a next session: item 3 (`bugs_open/128`, read, still unowned);
>   §4c/§4e design directions; the 079 PLATFORM fix (candidate 1) — now the queue's
>   biggest lever, Anthropic lane permitting; the anchors-don't-exist residual if the
>   fragment links should actually scroll.**

1. **The selector tool's hero is generated but NOT placed — and the generic build
   cannot place it.** Asset `d76f0282` exists (item `539893ae` complete 14:42Z). Work
   item `56fbcc9a` ("Re-render tool-model-approach-selector after its image asset
   landed") was retried at ~14:45Z WITH the asset present and no-opped again
   ("no sections ready to build") — so the missing-image premise was wrong.
   > **CORRECTED 2026-07-28 (same session that wrote the first version of this
   > item):** the calculator's hero needed NO re-render — its page already carried
   > the `content-hero-llm-cost-calculator.jpg` reference and generating the file
   > made it 200. The selector page carries **no hero `<img>` at all**, so it needs
   > content added, and it is a tool page — the generic `page-build-handler` finds
   > empty spec sections and no-ops (tool pages belong to the tool pipeline;
   > `save_page_sections` refuses `rebuild_policy='owned'` pages by design).
   Route it through the tool pipeline or a scoped section edit, not another retry of
   `56fbcc9a` (attempt 1 of 3 burned proving this; it is parked `needs_human_review`
   again, which is the correct resting state until the right route is chosen).
2. **The capabilities page still serves 13 broken refs** (9 links + 4 carousel svgs).
   Repair now has a decision to make that the morning handoff couldn't see: a hand
   repair still won't survive a rebuild (071's proven point), and the automated repair
   is proven vacuous (079 REOPENED). Options: (a) fix 079 candidate 1 (repair inside
   `save_page_sections`) then re-render; (b) scoped section re-render with corrected
   content_data for the carousel (LLM-free, safe, but same invention next rebuild);
   (c) wait for 092. The owner's "replace before deleting" rule applies to the images.
3. **`bugs_open/128`** (`image_url_404` never makes an HTTP request) — untouched, still
   worth a thread.
4. **After 08-01:** fire the owed 090 diagnosis verification on the 079 reopen claim,
   and any council submissions queued during the cap.
5. **Owner design directions** (§4c carousels/cliffhanger, §4e shapes registry) — the
   remaining open design surface. §4d is DONE and half-corrected (evening session):
   the join check exists (`sql_for_agents/256`, dual-column) and returns **zero
   missing** — §4d's "four wrong names" was a `function`-only join; all four resolve
   via `content_components.section_type`, so step 2 is MOOT. Step 3 (bind a site) is
   the experience-register thread's — its patterns were updated twice today; do not
   touch its rows. §4c/§4e remain untouched.

## Landmine added this session

- **L18 — a repair that logs durably still may not persist.** `CONTENT_LINK_REPAIR_DETAIL`
  rows are records of what the gate COMPUTED, not of what was saved. Two
  representations of the page travel the build (`clean_html` + `sections_metadata`);
  writes to the one that loses are silently vacuous. Verify content fixes against the
  saved `page_components` row, then the served page — never the action's return map.
  (016b §9 has the general pattern.)
