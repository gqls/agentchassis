# NOTES — bugs_open/285 tool-improver shared-template write (append-only, newest at the bottom)

## 2026-08-15 — pick, verify, research

- Ownership: `who-owns` → webdesign_uk lane (filer) + `bugfix_281_tool_audit_ported` (fix author,
  session 768d21d3, last active 18:57 local, finishing docs). Live-transcript grep: no other
  session on THIS 285 (390a1ae1 is on the OTHER 285, section-list lock-blind). Contributing into
  the bug file, not competing.
- **Fix NOT live** [MEASURED]: pods `agent-chassis-6f4688f88c-{sspqq,zmz4d}` image `v1.0.1302`,
  startTime 11:28Z; `25f92a967` committed 17:16Z, `d7b2d9994` 17:38Z. (A `grep -aq` probe of
  `/proc/1/exe` for those shas: absent — but note the "must be absent" control I used, 40 zeros,
  came back PRESENT, i.e. that control cannot fail; use a real-but-different sha.)
- Template + placements verified restored (4,664 chars, `{{.body}}`; 96+18 `deployed`, 1 `removed`).
- **`component_versions` v2 (7,714 chars, "Developer Resource Library", NO `{{.body}}`) is the
  08-05 POISON, not the recovery**: `store_generated_component` regen snapshots the CURRENT
  template BEFORE its UPDATE (`store_generated_component_action.go:439-451`). The regen OUTPUT is
  what v3 later snapshotted (4,664). Bug file/TL-042 read v2 as "recovery-by-coincidence" — wrong
  direction. → LANDMINE.
- **The 08-05 firing's real blast** [MEASURED]: poison → `component_template_corrupted:ported-page`
  (08-08 18:09) → component-creator regen → 3 `needs_rerender` (`component_regen_rerender`, reason
  `section_data_resolved`) → **154 `page_rerender`** across 3 sites → for owned ported pages the
  stub content_data lacks the schema's required llm `body` → **73 `needs_page` full LLM rebuilds,
  ALL FAILED** on `save_page_sections`'s owned-page guard ("rebuild_policy=owned … would clobber
  it"). The pages survived because of THAT guard, one layer down. So a shared-template regen is
  itself a fleet re-render trigger for every ported page.
- **CASUALTY FOUND** [MEASURED, then verified at the served page]: the improver's delivery item
  `section_edit_tool_fix_webdesign.co.uk_a7daa5c5…` (18:48:41Z, complete 18:51:59Z) targeted
  `learn-ai-builders-content-first` / slot `ported-page` — the OLD load_tool's `LIMIT 1` pick, not
  the tool's page. section-editor `content_edit` `field_updates {}` re-rendered the slot from the
  poisoned template: `rendered_html` 8,855 chars = wrapper CSS + `<article class="ported-page-content"> </article>` (EMPTY) + a fabricated "Related Downloads" list seeding `.asset-row` anchors to
  `/assets/downloads/{content-first-checklist.pdf,ai-content-brief-template.docx,prompt-library.md}`
  (non-existent). `curl https://webdesign.co.uk/learn/ai-builders/content-first.html` served it
  (200, 15,486 B, `portedPageAssetList` present). Fleet fingerprint sweep (`LIKE '%portedPageAssetList%'`
  over all page_components) → exactly this one row (positive control: it must appear, it did).
  The bug file's "checked clean" was a HEAD-of-row look; the poison's head IS the wrapper CSS.
  → WRONG_CALLS.
- Restore source proven: `page_component_history` `ab400131-2a41-434b-bd95-d44c9f064a32`
  (357 trigger, op=overwrite, 18:51:41Z, 3,781 chars); `sha256(rendered_html)` == the placement's
  `content_data->>'sha256'` (`a2d9fa85…`) — byte-exact webdesignport provenance match.
- **Seed 431 applied 18:18:31Z**: guarded restore (fingerprint + sha + no open rerender), digest
  NULL (ported bytes unstamped by design), reason-LESS `page_rerender`
  `f298cc52-dbf9-4993-ad82-7eee5dc57d97` queued at priority 20 / `triaged`. Dry-run in an aborted
  txn first (NOTICE lines identical). Poison archived by the trigger (1 row). **Served-page
  verification still owed** — see RUNBOOK.
- **"0 tool PLANs" (281 D1 premise) was measured on the wrong table** [MEASURED]: PLANs live in
  `doc_plans` (`check_tool_acceptance.go:590` reads it): **87 current tool PLANs; 14 of the 63
  ported webdesign tools have one** (asset-formatter included — that is why it could FAIL a
  criterion). `doc_notes … 'acceptance_criteria'` = 0 is true and irrelevant. D1's conclusion (no
  per-instance fixer exists) still stands; its "no PLAN to judge by" half does not. → WRONG_CALLS.
- **`fix_component_template` residual is mis-described** [read at HEAD]: its two `html_template`
  writes are `repair_template_slots` (component-scoped mechanical `<no value>x</no>`→`{{.x}}`, keyed
  by `spec.component_id`, no page_component_id) and `chrome_overflow_fix` (CSS APPEND to a chrome
  template reached via `site_components`, sharedness recorded in `shared_sites`). Neither is an LLM
  whole-rewrite from a per-page finding; neither restored the wrapper (v2/v3 show the regen did).
  The census line "takes page_component_id, reads the page's rendered_html" conflates the
  `align_slot_name`/`repair_page_component_status` metadata paths.
- Concurrent activity noticed: two ported instances (`tool-mind-map`, `learn-design-ambient-occlusion-css`)
  hand-edited via psql at 18:06:05Z today (archive rows, `application_name=psql`) — another lane's
  work, not touched.
- The improver's delivery step is a second per-instance hazard even with the pin: `section_edit`
  `content_edit` with empty `field_updates` RE-RENDERS the slot from template+content_data — for a
  ported instance that discards the instance's rendered_html. Any instance-scoped writeback must
  deliver via a reason-less `page_rerender` (assemble-only), never `section_edit`.

## 2026-08-16 — the roll landed; verify, induce, close, class-test

- **Fix LIVE** [MEASURED]: pods `agent-chassis-584b6fcf-{9mtqd,gz5bt}` `v1.0.1303`, started 18:45Z
  08-15. Stamp probe (`grep -aq <sha> /proc/1/exe` over the commits between 18:15Z and 18:50Z):
  the binary carries **`5e075a6f9`** — MY docs commit was HEAD at build time — and
  `git merge-base --is-ancestor d7b2d9994 5e075a6f9` / `25f92a967` both true. (The probe loop over
  12 candidates timed out after finding it: each exec ~10 s on the big binary — probe a SHORT
  window, and stop at the first hit.)
- **Casualty off the live site** [MEASURED at the served page]: `page_rerender:…:285-archive-restore`
  complete 18:22:24Z (08-15); `curl` → 200, 11,033 B, `portedPageAssetList` 0, `checklist.pdf` 0,
  h1 "The Content-First Strategy for Starter Sites" present. **My RUNBOOK positive control was
  wrong**: `class="article-content"` → 0 on the GOOD page (the ported markup uses `.article-content`
  only as a CSS selector; the h1 markup uses inline styles). Corrected in the RUNBOOK to the h1 text.
  Cheap check I skipped: run the positive control against the KNOWN-GOOD archived row before
  arming it (the row was right there in the DB). → WRONG_CALLS.
- **Third-firing watch clean**: 4 versions, template 4,664 `{{.body}}`, 114 deployed + 1 removed, no
  new item at `a7daa5c5…` since 17:00Z 08-15.
- **Induced refusal at the artefact** [MEASURED]: one-step ad-hoc orchestrate (`update_component_html`,
  `component_id: input_data.component_id`, `html_field: input_data.html`, `create_version:false`),
  payload = the CURRENT template byte-for-byte (md5 `6690f1aa…` both sides) so a non-firing fence
  would have written identical bytes rather than poison (it would still have flipped 115 → pending
  and added a version — the un-flip recipe was ready). Published via `kubectl run … kcat -P` with the
  JSON file on stdin (one line = one message). Result: orch `a9a824f5-cf9c-4fa1-b0a1-30ce7b99fe3b`
  **FAILED at `induce_write` in 0.4 s**; `agent_error_log` 09:59:06Z
  `component_write_shared_blocked` "section-level component placed on 115 pages across 2 sites";
  template `updated_at` unchanged; 0 pending; 4 versions. Close criterion 1 met.
- **Post-roll producers**: 2 `improve_tool` items since the roll, both real forks with per-page
  `audit_fix_<domain>_<page_id>` keys (seed 425's shape working); 0 `ported_tool_fix` yet — no
  design-discovery sweep has run on webdesign since the roll (last 08-14 16:04Z). 281's first-sweep
  census is still owed to THEIR lane, not forced here.
- **Class closure built + proven**: `component_template_writer_coverage_test.go` — same shape as
  `page_component_writer_coverage_test.go`; regex on the SQL, comments stripped
  (`withoutLineComments`), fan-out-intended map with reasons, the fenced writer must be SEEN.
  Mutations: (1) stub the fence call in `update_component_html_action.go` (compiling stub, not a
  rename — my first attempt was an undefined symbol and only broke the BUILD, which proves nothing)
  → FAIL "was not seen as a fenced writer"; (2) `zz_mutant_writer.go` with an unfenced UPDATE → FAIL
  naming it; restored → green. Council `d8668e1f-6272-4888-b116-19edbac283b2` submitted; committed
  `e2064f3bb` with `Council-Submitted:`.
- **Cross-session collision caught by `go vet`**: an UNTRACKED sibling test in another session
  (`agent_definition_nullable_columns_test.go`) declares an identical `stripLineComments`. HEAD was
  fine, the working tree was not, and whichever landed second would have broken HEAD. Renamed mine
  to `withoutLineComments` (`87aab3c82`). Then the package's test build broke anyway on ANOTHER
  lane's in-flight `component_instance_scope.go`/`_test.go` (InstanceToken signature) — verified my
  tests green against `git archive HEAD` in scratch (memory: shared-tree-wont-compile).
- **016b §9 correction** landed as a same-file passenger in another lane's commit `fa9c454ab`
  (they swept 016b minutes after I edited it) — nothing lost, noting for the record.
- **285 → `bugs_closed/`** (`a88090f4f`, both paths on the commit, exactly one line at HEAD).
  TL-042 status-update + index row (`67996ebf1`; index carried two other lanes' one-line rows).
- **Council `d8668e1f` APPROVED round 1** (10:04Z, ~6 min after submission — the "budget 30 min"
  figure is a queue-latency ceiling, not a floor): 10 approve, 1 object (editquality, medium),
  6 abstained. Acted on rather than defended: (i) each fan-out-intended entry now cites its
  InputSpec + the write it is exempted for — read at HEAD, all four are `Required [site_id]`
  sweeps (colour fixers select components IN USE on the site — a template shared with a second
  site is rewritten too, stated as intended for a declaration-level substitution; nav-link fixer
  is chrome via site_components) or a component-keyed regen; NONE takes a page input; (ii)
  subpackage sweep (`*/*.go`) — mutation-proven with a fake writer in `discovery_checks/` → FAIL;
  (iii) the guard header's stale "page-aware … open" line — the 281 lane had ALREADY corrected it
  (`fa661d5d2`, 11:06 local) after reading my contribution; I added the pointer that the test is
  now the census. bug_historian (low): call-exists ≠ verdict-obeyed — true, stated in the file;
  the behavioural half is `update_component_html_shared_fence_test.go`. Follow-up `194455b40`
  with `Council-Reviewed:`.
- 102_coverage_ratchet.txt gains the lane dir (pattern-check advisory on the first commit).
