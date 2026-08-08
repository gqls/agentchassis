# HANDOFF 2026-08-08b — continue here (`bugs_open/206`, directory-build-handler)

**Supersedes `HANDOFF_2026-08-08_continue_here.md`** (keep it — accurate through
round 3's submission; every one of its numbered steps has now been either done
or overtaken by events below). **Read `NOTES_directory_build_handler.md`'s
2026-08-08b/c/d entries first** — full evidence trail, predictions-vs-outcomes,
both live-fire defects.

## State in one paragraph

Council round 3 **APPROVED** (09:18Z; three low advisories, none gating — use
`Council-Reviewed: 5b8e4cf7-31c3-4793-a550-d6b9be1f00e8` on this lane's
commits; `5e2697205`/`a3509b77f` already carry it). Migrations **325, 326,
336, 337 all applied and recorded** (each has its own DO/RAISE guard; 336 and
337 are live-fire corrections to 326's `call_page_build_handler`
input_mapping — see below). Go code confirmed on **v1.0.1266** (roll landed
~16:00Z; pod-grepped BOTH replicas: `ensure_page_section_layout` 5,
`business_directory` 3, `directory-build-handler` 1, round-2 marker
`zero-business result` 1, negative control 0). `directory-build-handler.image_tag`
in `agent_definitions` still says v1.0.1264 — harmless label, update it when
convenient. The two target pages are **NOT yet deployed**; both `needs_page`
items are mid-retry (see "Where it stands right now"). The owner's
improvement-loop question is fully answered and written up (NOTES 2026-08-08c,
`bugs_open/220`).

## What happened today (compressed)

1. **Round 3 read: APPROVED.** Migrations 325+326 applied by hand, recorded;
   `image_tag` set to the then-live v1.0.1264 after pod-grep.
2. **Owner asked whether the improvement loop would have caught bug 206's
   problems. Experiment run** (one-shot improvement-loop over vetcomparison.uk,
   corr `867d6054-...`, predictions registered in NOTES before firing):
   - Detection: YES — historically (2026-08-02, three `unbuilt_internal_link`
     items) and again live (7 more + 1 `empty_internal_href`).
   - Fix: NO — **`bugs_open/220` filed**: the dispatcher maps
     `spec.page_name` (the page CONTAINING a broken link), never the item's
     `page_id` (the TARGET the check filed against), so it rebuilt the wrong
     pages, marked the items `complete`, and will re-detect forever. Live
     re-demonstration is in the bug file (a §9 pattern + a footprinted
     LANDMINES entry shipped with it; landmine synced to doc_notes, verifier
     fired, corr `42546607-...`, verdict unread — check
     `doc_notes categories ? 'landmine-verification'` if curious).
   - **Better than predicted**: the loop REVIVED both parked `needs_page` rows
     (`incomplete_page_group` re-mint + `refreshOpenWorkItem`, the 091/184
     machinery) and re-routed directory-index to `directory-build-handler`
     via the builder map — the new capability was picked up by the loop
     itself, no hand SQL.
   - I cancelled the loop's 5 remaining duplicate wrong-page rebuild items
     (error text on each points at 220) and bumped the two `needs_page` rows
     to priority 95 to get ahead of a 14-item rerender wave.
3. **Two config defects in migration 326 found by the live dispatches, both
   fixed by migration, both committed:**
   - **336**: input_mapping KEYS carried the `input_data.` prefix → child
     rejected the call (contract violation: missing site_id/domain). Keys are
     the CHILD's field names; values are parent dot-paths.
   - **337**: delegation passed only site_id/domain/page_name;
     `update_status` (which runs BEFORE `deploy_page` — content saved,
     nothing deployed) resolves the page via `input_data.spec.page_name` /
     `current_page.name`, both normally supplied by the dispatcher's own
     mapping (`"spec"/"current_page": "current_item.spec"`). 337 mirrors
     those keys into the delegation. Transferable lesson in NOTES 2026-08-08d:
     a delegating agent must satisfy the target's WHOLE input contract
     (every consumer step), not just its entry step.
   - `ensure_page_section_layout` itself **worked in production**:
     `site_plan_sections` now carries `directory-index: [hero,
     directory-listing]` — written by the loop's own dispatch, first
     production use.
4. **Fresh roll to v1.0.1266 landed ~16:00Z** (owner). Both items were safely
   `triaged` (not in flight). I HELD them (`needs_human_review`, error text
   says why) through the ~300s post-restart silent-spawn-drop window, with a
   background job that re-triages both at pod age ≥330s. One likely roll
   orphan NOT in this lane: `audit_tool` item `8bd00299` claimed 15:56:25Z —
   left for the claimed-item-timeout reaper.

## Where it stands right now (re-verify, do not trust)

- `715ec305` needs_page:directory-index — handler `directory-build-handler`,
  **attempt 2 of 3 used**; the remaining attempt runs with the full 337
  mapping. Attempt 2 proved the chain through save_sections; the remaining
  steps (update_status → spawn_rerender_agent → deploy_page) read only
  child-produced fields (pre-checked, NOTES 08-08d).
- `2f50bfda` needs_page:guides-index — **re-routed to `directory-build-handler`**
  (deviation from the 08-08 handoff's Step 5, which said "handler unchanged";
  the live system itself refuted that — the loop re-dispatched it on bare
  page-build-handler and it no-op'd again). Attempt 1 of 3 used (pre-337
  snapshot, failed at update_status as expected).
- Both were re-triaged automatically after the drop window — **verify**:
  ```sql
  SELECT id, status, handler_agent, attempt_count, error FROM site_work_items
  WHERE id IN ('715ec305-1de1-4901-b988-b4880d58cce9','2f50bfda-0e2f-4a2d-bb14-22f76114f092');
  ```
  If still `needs_human_review` with the HELD error text, the background
  re-triage died — re-run its UPDATE by hand (status='triaged', error=NULL).
- `build-pipeline-trigger` (120s) dispatches them unaided. Builds take ~2-5
  min each. Watch pages, not statuses:
  ```sql
  SELECT name, build_status, deployed_at FROM pages
  WHERE site_id='72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND name IN ('directory-index','guides-index');
  ```

## Remaining steps

1. **Confirm both builds complete and pages deploy.** Then verify the
   ARTEFACTS: `curl -s https://vetcomparison.uk/directory/index.html` — real
   business names/postcodes (an empty list = the ensure_layout resolver found
   no directory-export-json config — a different failure, distinguish it);
   `/guides/index.html` — the three real guide titles (guide-cma-compliance,
   guide-cma-market-investigation, guide-independent-strategy), not invented
   ones. Also `last-modified` fresh on both.
   - If directory-index's LAST attempt fails: read the new error first. You
     may reset `attempt_count` (operator action, precedent in this lane) —
     but only after diagnosing; a third distinct config gap in the delegation
     would want its own migration like 336/337.
2. **Record closure evidence INSIDE `bugs_open/206`** as a dated section —
   **do NOT move the file to bugs_closed/** (owner direction 2026-08-06,
   overrides CLAUDE.md's "fixed AND live → bugs_closed" — see memory
   `owner-keeps-fixed-bugs-in-bugs-open`).
3. **Concept register `BLD-017`** (register/build-pipeline.md): update status
   to deployed/proven-live with the evidence (work item ids, orchestration,
   curl). Mention 336/337 in the entry — the seed file 326 in the repo does
   NOT match the live row's mapping (the migrations corrected it); a reader
   of 326 alone inherits the defect. Consider whether 326's file deserves a
   pointer comment (forward-only: add a header note, don't rewrite history).
4. **Tell the owner plainly** whether the original complaint (alphabetical
   vet list / "Search the directory" 404 from `features_open/021`) is now
   resolved on the live site.
5. **Clean up experiment residue** (optional, low priority): the loop's
   rerender wave items for the two pages may rerender them again post-deploy
   (harmless); the two `unbuilt_internal_link` items that completed
   wrongly (3f066b90, 4ba1d4dd) are terminal — leave them, they're bug 220's
   evidence. The cancelled five carry pointers to 220.
6. **`bugs_open/220`** is filed, evidenced, reproduced — its fix (dispatcher
   honours `page_id` + verification at mark_complete) is a SEPARATE lane;
   don't fold it into 206. `entity-page`/`practice` stays deliberately
   unbuilt (P1 crawl 10/~2,109 as of 08-06).

## Cost/context note

This session ran long (experiment + two live-fire fixes + filings). If you are
continuing in a fresh chat: NOTES 2026-08-08b/c/d + this file + `bugs_open/220`
carry everything; nothing lives only in the old chat. The 090/council trail for
this lane: corr `5b8e4cf7-31c3-4793-a550-d6b9be1f00e8` (rounds 1-3, approved).
