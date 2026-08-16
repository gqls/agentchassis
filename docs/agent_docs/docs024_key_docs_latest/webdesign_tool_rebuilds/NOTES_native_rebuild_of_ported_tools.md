# NOTES — native rebuild of ported tools (append-only, newest at the bottom)

## 2026-08-15 ~17:00Z — lane opened; mechanism established; safety restore done; pilot filed

- Corrected my own earlier claim to the owner: the tool-generator was NOT "replacing ported
  tools" — `check_missing_tools`/tool-suggester does GAP analysis and ADDS tools. css-unit-
  converter (00:07 today) was an addition; the 63 ported tools had no replacement path at all.
- The one page where native met ported (`tool-ab-test-calculator`) is the cautionary case:
  native slot attached ALONGSIDE (position 2), ported slot left `pending`, and a raw
  `{{.section_heading}}` served to visitors. Junk census across all 4 native slots: ONLY
  ab-test carries a raw tag — the generator's newest output is clean.
- SAFETY RESTORE executed (owner-sanctioned, closes bugs_open/281 Finding A hazard): shared
  `ported-page` wrapper template restored from the poisoned 8,864-char asset-formatter markup
  to v3's pre-poison snapshot (4,664, `{{.body}}` intact); poisoned state banked as v4;
  114 `pending` placements un-flipped to `deployed`. Verified: template 4,664/passthrough=t,
  0 pending. NOTE: the 281 addendum said "restore v1 (77-char seed)" — v1 was NOT the
  pre-poison state (legitimate 08-05→08-14 edits sit between); v3's pre-edit snapshot is.
- Retired ab-test's ported slot (`removed` — the assembly-excluded tombstone, verified in
  code before use) and filed `section_edit:owner-gate:tool-ab-test-calculator-rawtag` to fix
  the raw tag; its reassembly also drops the removed slot from the served page.
- PILOT filed: `add_tool_novel_webdesign.co.uk` → tool-generator, `tool-aspect-ratio`
  (simple; fleet-wide function claim checked empty; same URL, page_id pinned). Next session:
  RUNBOOK — retire, re-render, grade at the artefact, then batch simple tools serially.
- [ASSUMED] the deploy will attach to the existing page by matching name/function, as it did
  for ab-test. If it instead creates a duplicate page/URL, stop and re-plan before batching.

## 2026-08-15 ~18:40Z — continuation: floors refused the section-edits (good), surgical repairs applied, pilot queued behind a serial dispatcher

- **CORRECTION of my 17:00Z entry and the same claim in `bugs_open/281`:** "section-editor is
  the healthy lane for ported instances" was WRONG. Its completions are on real FORK slots.
  On a ported blob it REGENERATES the section via LLM and the **slot floors refused both
  writes** (mindmap `28→4 class attributes (14% kept, floor 50%)`; occlusion `11→4 (36%)`) —
  bugs 178/253 floors doing exactly their job; nothing was damaged. Net lesson for this lane:
  **a ported tool has NO safe framework editor** — audit-blind (281), improver-poisonous
  (285), editor-refused (today). Rebuild or decompose are the only managed paths.
- **Surgical repairs applied instead** (guarded single-transaction `replace()`; DO/RAISE
  asserted occurrence counts; `trg_page_component_artefact_archive_upd` banked pre-states):
  1. mindmap: `background: var(--surface); color: white` → `var(--primary)` ×2 (the `.root`
     node + "+ New Map" inline). Root cause: dark-theme idiom ported to a light site —
     `--surface: #f3f1ec` under `color: white` ≈ 1.1:1. White on `--primary #5c6b5d` ≈ 6.8:1.
  2. mindmap: usage hint inserted before `map-list` (verified against the JS: contentEditable
     node text, hover +/× controls, drag, localStorage autosave — hint states only what the
     code does).
  3. occlusion: the bare-text `[Image of ambient occlusion sphere diagram]` line removed
     (not replaced — authoring a figure is the imagery lane's job, not a repair).
  Both propagating via `page_rerender:owner-gate:*` items (the assemble path — proven by the
  occlusion page's own 9 prior completed rerenders preserving the ported slot).
- **"Centrdgsdgsdgsdal Idea" is NOT a site defect** — the tool seeds "Central Idea"; node text
  is contentEditable and autosaves to localStorage, so the junk lives in the OWNER'S browser
  state only. Told the owner; no item filed.
- **Why the pilot hasn't started: the build-dispatch-loop is SERIAL per site** — measured:
  each claim fires the second the previous item completes (17:21:31 → 17:22:15 → 17:22:58 →
  17:34:43). `add_tool` (prio 60) sits behind the rerenders (45), the rawtag edit (50) and a
  claimed audit_tool that hit "Claim timed out — handler pod likely died" once already.
  Selection: status `triaged/approved` + attempts < max + approval auto (read at
  `load_work_item_actions.go:641-679`) — the item qualifies; it is queue position, not a gate.
- audit_tool progress: css-unit-converter COMPLETE (first real LLM audit of that tool);
  ab-test claimed/in-flight; two more queued.

## 2026-08-15 ~22:20Z — rerenders VERIFIED live; fence LIVE on v1.0.1303; pilot ran EMPTY

- **Mindmap + occlusion repairs verified at the served artefact with controls**: mindmap
  newcss×2 + hint + 'MIND MAP STUDIO' control (first fetch hit a stale CDN edge object — a
  9-byte "Not found" under HTTP 200 — which expired by itself minutes later; if seen again,
  cache-bust before diagnosing); occlusion placeholder gone, 'Uncanny Valley' control present.
- **v1.0.1303 verified**: pods 18:45Z; binary probe positive for build sha `5e075a6f9`, junk
  control absent; `d7b2d9994` (shared-write fence + fixed producers/loaders) is an ancestor ⇒
  **285's fence is LIVE**; its induce-a-refusal close-out is now runnable.
- **Pilot add_tool "completed" 18:29 but produced NOTHING**: orchestration `5ef53886` ran 47s,
  `final_result` empty, no `tool-aspect-ratio` in `content_components`. Matches the
  spawn→call handshake-failure shape (MEMORY: ~half fail; never cancel pre-diagnosis).
  Diagnose the run before refiling.
- ab-test: raw tag is OUT of the stored fork slot (the "failed" section_edit applied its edit
  before its reassembly step failed); ported slot `removed`;
  `page_rerender:owner-gate:tool-ab-test-calculator` queued to serve the fork-only page.
- Two gap-shape `improve_tool` items (`audit_fix_webdesign.co.uk_c5baf147…`/`_1c1b186e…`)
  fail `load_tool` on missing `spec.page_id` (created between seed 426 and the roll); my
  backfill UPDATE matched 0 rows — spec shape differs from assumed; re-diagnose.
- My "checked clean" wrong call on the 285 placement: logged in `WRONG_CALLS.md` (caught by
  the 285 lane; instrument greped invented spellings).

## 2026-08-15 22:10Z → 2026-08-16 10:10Z — pilot diagnosed (NOT the handshake), two seeds, one Go fix approved+committed, ab-test REVERTED to the ported tool

- **MISSTEP, corrected: the handoff's "pilot empty run matches the spawn→call handshake failure" was WRONG.**
  The three "empty `final_result`" generator runs were three different things: `5ef53886` (aspect-ratio)
  died at `save_tool` on `pages_site_id_name_key` — `agent_error_log` 18:29:17Z said so all along; the
  other two (`tool-storage-risk-explainer`, `tool-overpayment-priority`) CREATED their components
  (`content_components.created_at` matches) — empty `final_result` is that agent's normal shape.
  What caught it: reading `agent_error_log` for the window instead of `orchestration_states`
  (MEMORY `spawn-call-handshake-races` says exactly this and I skipped it). → `WRONG_CALLS.md`.
- **Root cause of the pilot (090 CONFIRMED first iteration, corr `3050effc`):** `create_tool_component`
  has NO attach-to-existing-page path — bare `INSERT INTO pages`, no `ON CONFLICT`, no lookup, no
  `page_id` input; the error branch deletes the just-created component. Same binary built ab-test's
  fork the night before because ab-test went through `tool-deployer`/`deploy_tool_to_site`
  (`resolveToolPageIdentity` + `UpsertPageForRole`), not the generator. → `bugs_open/286`.
- **Fix built, tested (3 sqlmock pins, cleanup guard mutation-proven), council APPROVED `27d0f428`
  23:14Z, committed `88897190e` with the trailer + register TL-044 + seed `435_…_HOLD.sql`.** Opt-in
  `adopt_existing_page` (default OFF). INERT until (a) a chassis roll carries `88897190e` and (b) 435 is
  un-HELD and applied — image before config. **The cluster at 09:52Z 2026-08-16 still ran `v1.0.1303`**
  (pods 18:45Z 08-15, deploy image 1303, no newer local image) — the owner's "fresh chassis build" note
  did not correspond to anything visible; verify with the binary probe before applying 435.
- **The two "gap-shape" improve_tool items were a CLASS, not a window:** `spec_data` passed as a MAP is
  never read (contract: path string) ⇒ **233/233** `audit_fix` items ever filed had `spec={}`; four live
  producers. Seed **434** (applied 22:1xZ 08-15, committed `bc4cd65e7`) moved tool-auditor ×2,
  internal-linker, component-quality-auditor to `spec_paths`/`spec_literal`, backfilled the 4 actionable
  webdesign items from their row columns, re-armed 2 (+ reset llm-cost's counter; ab-test's held).
  **Proof it worked:** all three ran within 3 minutes (22:37–22:40Z), each wrote a new fork version via
  `update_component_html` and delivered a `section_edit` (complete). Not yet graded at the served page.
  → LANDMINES entry.
- **ab-test: the fork is a HOLLOW SHELL and the page was serving 47 raw tags.** The template
  externalises 31 UI-copy fields (`section_heading`, `calculate_button_label`, …) nobody supplies;
  deployer served it verbatim (raw tags); the 08-15 section-edit re-rendered it against `{}` ⇒
  13,284 chars, **0 visible chars** (class-attr floors passed — every class survived); the queued
  rerender then failed "blank slots [tool-ab-test-calculator]". No archive row for the pair. The
  handoff's "raw tag REMOVED from stored html" was true for the wrong reason. **Reverted 10:0xZ:**
  ported slot `ebe3c57a` → `deployed`, fork `1a9efade` → `removed`, item
  `page_rerender:owner-gate:tool-ab-test-calculator:revert-to-ported` queued (assemble-only).
  ab-test is rebuild candidate #2 via the adopt route. → LANDMINES entry; 286 "related finding".
- **MISSTEP (harmless): my "no open items on this page" guard counted 10 `unresolved` items and
  aborted** — `unresolved` is a dead status the dispatcher never selects; guard on
  `IN ('triaged','approved','claimed','pending')` instead. RUNBOOK updated.
- **285 close-out is being done by its owning lane RIGHT NOW:** `agent_error_log` 09:59:06Z
  `step_name=induce_write`, `error_code=component_write_shared_blocked`, "placed on 115 pages across
  2 sites"; wrapper still 4,664/`{{.body}}`/4 versions. Not touched by this lane; noted for the record.
- Rerender queue timing: the 22:05Z owner-gate rerender was claimed 23:26Z (~80 min behind the
  serial dispatcher); expect the revert to take about that.
- **ab-test revert VERIFIED at the served page 10:1xZ** (item complete 10:01:39Z — queue was empty):
  raw `{{.` 0 (was 47), `abc-container` 0 (fork gone), `class="ported-page"` 1, control ×3, `<script>` 5,
  inputs/buttons present. **RUNBOOK correction:** the grade line's `ported-page-section` fingerprint was
  never in the served markup (0 before and after) — a check that cannot fail; replaced with
  `class="ported-page"`, taken FROM the artefact.

## 2026-08-16 15:10Z–15:35Z — v1.0.1304 carries the 286 fix; seed 435 applied; pilot REFILED; the three audit fixes graded PASS

- **Roll verified the right way this time:** pods `v1.0.1304` up 10:41Z. My first probe grepped
  `88897190e` and the OLD build sha `5e075a6f9` — BOTH absent, which proves nothing: the binary carries
  only ITS OWN build sha. Found the stamp by probing every commit in the build window
  (`5de6cddbe`, 10:25Z), then `git merge-base --is-ancestor 88897190e 5de6cddbe` ⇒ TRUE; a junk-hex
  negative control absent. **RUNBOOK's probe block corrected accordingly** (probe the STAMP, then
  ancestry — never "is my sha in the binary").
- **Seed 435 HOLD lifted + applied 15:15Z** (`ee0228813`, git mv, one file at HEAD): tool-generator
  `save_tool.adopt_existing_page = true` (verified by SELECT).
- **Pilot refiled** as item `99734862` (`add_tool_novel_webdesign.co.uk`, `triaged`), description
  written from the LIVE tool (two parts: gcd-reduce W×H → ratio; target ratio + one dimension → the
  other, presets 16:9/4:3/1:1/21:9), explicitly self-contained copy (no template fields — the ab-test
  lesson). 11 items ahead in the serial dispatcher at 15:30Z. Preconditions checked: no open add_tool
  on the site, `function` claim empty fleet-wide, page `00979b9e…` exists, flag `true`.
- **Three re-armed audit fixes graded PASS at the served pages** (css-unit-converter,
  css-specificity-calculator, llm-cost-calculator): 0 raw tags, tool container present, stored tool
  slots 527/587/1,209 visible chars, inputs/scripts intact, deliveries reached `rendered_html`
  (23:24Z/12:22Z/23:25Z). Behavioural correctness of each fix is the tool_acceptance pipeline's
  call, not graded here.
