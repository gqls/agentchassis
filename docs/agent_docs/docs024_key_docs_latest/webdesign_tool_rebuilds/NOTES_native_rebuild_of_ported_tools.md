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
