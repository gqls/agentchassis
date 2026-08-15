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
