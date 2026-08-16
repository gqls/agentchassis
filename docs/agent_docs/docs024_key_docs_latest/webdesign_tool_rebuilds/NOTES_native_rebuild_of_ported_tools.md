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

## 2026-08-16 16:05Z–16:20Z — the pilot's adopt run GRADED PASS (steps 1+2); step 3 BLOCKED on a session permission

- **Step 1 — the adopt path did exactly what 286's fix promised. `[MEASURED]`**
  Item `99734862` claimed 15:47:33Z, `complete` 15:48:27Z, `error` empty (54 s end to end).
  Generator orch `72f0737e-5902-456e-b09e-dfd091b3aec1` (`tool-generator`, COMPLETED 15:48:28Z),
  `collected_data->'create_result'`:
  `page_adopted: true`, `page_id: 00979b9e-db47-4c26-819e-add95b0f8fd6` (the EXISTING page),
  `component_id: 96ac3a04-435e-46f8-8ffd-1549e3bc363f`, `generated: true`.
  `page_components` on `00979b9e` = exactly two rows: ported `f32583ed` (`ported-page`, position 0,
  `deployed`) and the new `d3bfd11a` (`tool-aspect-ratio`, position 2, `component_level='tool'`,
  `deployed`, 8,958 chars). `pages` gained **no** row named `tool-aspect-ratio`.
  ⇒ the collision that killed the 08-15 run (`pages_site_id_name_key`) is gone; 286 is proven live.
- **Step 2 — the generated component graded BEFORE any retire (the ab-test lesson). PASS. `[MEASURED]`**
  `content_components` `96ac3a04`: `html_template` 8,958 chars, `{{.` field count **0**,
  visible chars **469** (>300), `<script` **1**, `<script src=` **0**, `is_active`.
  **And read, not only counted** — the two parts the spec demanded are both there and real:
  `gcd()` reducing W×H to a simplified ratio plus a decimal, a `W:H` parser driving the
  missing-dimension calculation, inputs `ratio-width/ratio-height/target-ratio/dim-width/dim-height`,
  and the four preset buttons `16:9 / 4:3 / 1:1 / 21:9`. Self-contained copy, no template fields —
  the hollow-shell shape that sank the ab-test fork did not recur.
- **CORRECTION to the OWNER RULING's item 3, as written into PLAN and the 08-16 handoff.** Both say
  to "note the `page_component_history` archive row id per tool … that is what makes a bad
  reimplementation a one-statement revert." **There is no such row to note, and there never will be
  for a retire.** The archive trigger is
  `trg_page_component_artefact_archive_upd AFTER UPDATE OF rendered_html … WHEN (old.rendered_html
  IS NOT NULL AND new.rendered_html IS DISTINCT FROM old.rendered_html)` (plus an AFTER DELETE twin).
  A `build_status` flip touches neither condition. Measured: `page_component_history` holds **0 rows**
  for page `00979b9e`. This is also the unexplained observation in the 08-16 entry above
  ("No archive row for the pair" on ab-test) — recorded there as a fact, with its mechanism missing.
  **The revert handle is the ported `page_components` row itself**, which survives *because* we set
  status instead of deleting: id `f32583ed-7403-4c1a-a363-d1794bd78591`, `rendered_html` 5,850 chars,
  md5 `22edea99f47275258a03fa27d9c8d8d1`, pre-state `build_status='deployed'`, position 0.
  Revert = flip that one row back to `deployed` + re-render, exactly as the ab-test revert did.
  **Record the row id and the md5 per tool** — not an archive id.
- **The generator files its OWN rerender, with the right shape.** Item `937f3a52-bab4-4da2-84d7-42b4d0cf938f`
  (`page_rerender`, key `page_rerender_tool-aspect-ratio_<site>_assemble`, `triaged` 15:48:25Z), spec
  `{domain, page_id, filename, page_name}` — the assemble-only shape the RUNBOOK prescribes, no `reason`.
  So step 3 does **not** need an item filed; it needs the ported slot retired **before** that item is
  claimed. If it renders first, the page serves BOTH tools — the ab-test shape. 34 items ahead of it
  at 16:06Z; the queue was running `content_rewrite` items ~2–4 min each.
- **Two adopt-route side effects the PLAN does not mention** (neither harmful, both worth expecting on
  every subsequent rebuild): the generator created a **guide page** `cb5bd424-884c-4e0f-b047-9f1766357014`
  (`tool-aspect-ratio-guide`, `/guides/tool-aspect-ratio-guide.html`, 15:48:12Z) and reported
  `cross_links_added: 2`; and it filed `needs_content_page` (`tool_content:tool-aspect-ratio:…`) and
  `nav_drift` (`nav_rebuild:<site>`) items alongside the rerender. So "no new page" is true of the TOOL
  page only — the guide page is a new `pages` row, and it is the generator's normal output.
- **Baseline of the served page taken before touching anything `[MEASURED]`:**
  `https://webdesign.co.uk/tools/aspect-ratio/index.html` → 200, 13,113 B, `class="ported-page"` **1**,
  `{{\.` **0**, `<script` 5, new-tool marker `ratio-width` **0**.
  Note the trailing-slash form (`/tools/aspect-ratio/`) is a **404** — grade the `index.html` URL, or a
  clean `{{\.`-count passes against a 404 page that could never have contained a raw tag.
- **BLOCKED, and it is a session permission, not a platform fault:** the guarded retire UPDATE
  (RUNBOOK's per-page block, wrapped in a txn with two `DO`/`RAISE` post-asserts — exactly 1 removed
  `ported-page` slot, exactly 1 `deployed` slot left) was refused by this session's auto-mode
  classifier, in both heredoc and `psql -f` form. Reads against the live DB are unaffected. The SQL is
  ready at `<scratchpad>/retire_aspect.sql`. Owner asked to unblock; nothing else in step 3 is started.

## 2026-08-16 16:17Z — step 3a: the ported slot is RETIRED (owner unblocked the write)

- **Retire committed, guarded, `UPDATE 1`.** Ported slot `f32583ed-7403-4c1a-a363-d1794bd78591`
  `deployed` → `removed`; `rendered_html` **untouched — 5,850 chars, md5 `22edea99f47275258a03fa27d9c8d8d1`,
  byte-identical to the pre-state captured before the UPDATE.** That equality is the point of the
  check: it proves the revert handle survived the retire, so a bad reimplementation is one status
  flip away from the bytes that were actually served. Both in-transaction asserts passed (exactly 1
  `removed` `ported-page` slot; exactly 1 `deployed` slot remaining), then COMMIT.
- Post-state on page `00979b9e`: `ported-page` position 0 `removed` (5,850 chars retained) ·
  `tool-aspect-ratio` position 2 `deployed`, `component_level='tool'` (8,958 chars). One live tool.
- **We beat the generator's own rerender**, which is the ordering that matters: at 16:19Z item
  `937f3a52` is still `triaged` with **26 ahead** (was 34 at 16:06Z, 30 at 16:12Z ⇒ ~4 items / 6 min).
  So when it fires it assembles the page with the native tool only — one render, not a double-tool
  intermediate. **Generalises to every tool on this route:** the retire window is the queue depth
  behind the generator's rerender at the moment the build completes, not an unbounded interval.
- Not yet done: grade at the served artefact once `937f3a52` completes (require `http=200` on
  `/tools/aspect-ratio/index.html`, then `class="ported-page"` 0, `{{\.` 0, `ratio-width` ≥1,
  `<script` ≥1). Until that grades PASS the pilot is retired-but-unproven, not replaced.

## 2026-08-16 16:22Z — batch census for PLAN §2, taken while the rerender queues `[MEASURED]`

Read-only; nothing filed. All counts against the live DB, site `6b49db8e…`.

- **97 `ported-page` slots exist; only 63 were ever tools.** By top-level directory of `pages.url`:
  `/tools/` **64**, `/learn/` **32**, `/about/` **1**. And of the 64 under `/tools/`, one is
  `tools-index` (`/tools/index.html`) — the listing page, not a tool. So **63 ported tools**, which
  reconciles the PLAN's figure exactly, and the other 34 ported pages are out of this lane's scope.
  **Guard for the batch query: filter `p.name LIKE 'tool-%'`, not `p.url LIKE '/tools/%'`** — the
  latter files the index page as a tool.
- **62 remain deployed** (63 minus the pilot). Ported body sizes: min 3,806 B, avg 9,309 B,
  max 23,900 B (`tool-logic-architect`).
- **TRAP — there are TWO different sets of 13 here, and they are not the same 13.**
  `<script src=` in the ported slot's html: **13** tools. `content_data ? 'repair'`: **13** tools.
  **Intersection: 4.** So using the `repair` key as a cheap proxy for "the external-script class the
  PLAN warns about" mis-scopes **9 tools in each direction**. The external-script 13 is the set the
  PLAN and register **TL-032** mean (TL-032's landmine states "13 of webdesign.co.uk's 63 tools load
  relative `<script src>` files" — my independent count agrees, which is the corroboration, not the
  coincidence). The `repair` 13 is residue of the `webdesign_tools_repair` lane's runs.
  Derive the class you actually want; do not pick the other 13 because it is the same number.
  - external-script (the PLAN's class): micro-cms(4), vibe-equalizer(2), bayesian-rank,
    blueprint-compiler, community-growth, csp-builder, fluid-typography, head-architect,
    image-optimizer, performance-budget, recommender-engine, seo-schema, smart-contrast.
  - `repair`-key only (NOT external-script): animated-favicon, asset-formatter, insight-injector,
    logic-architect, mind-map, monolith-splitter, pasteboard, rls-architect, seo-injector.
- **Five native tool components now exist on the site**, and their placements matter for §4:
  `tool-aspect-ratio` deployed (the pilot) · `tool-css-specificity-calculator` deployed ·
  `tool-css-unit-converter` deployed · `tool-llm-cost-calculator` **approved, not deployed** ·
  `tool-ab-test-calculator` **`is_active` with placement `removed`** — the hollow fork.
  The three css/llm ones sit on pages with **no** live ported slot, so they were generator ADDITIONS,
  not replacements; only aspect-ratio is a replacement. ab-test's page correctly shows ported live,
  native gone — the 10:0xZ revert held.
- **The `already_exists` short-circuit is CONFIRMED by reading the source**, not inherited from the
  handoff: `create_tool_component_action.go` ~197–217 joins `content_components → page_components →
  pages` on `cc.function`, `cc.component_level='tool'`, `p.site_id` and **`cc.is_active=true` only —
  no `build_status` predicate**, `LIMIT 1`, and returns `{already_exists:true}` without writing.
  So the ab-test fork must be `is_active=false` before ab-test is refiled, exactly as the handoff says.
  Corollary the handoff does not draw: the same probe now makes **any** tool with a live native
  component un-refilable, which is the desired throttle — but it means a FAILED rebuild that still
  left a component behind must be deactivated before a retry, or the retry silently no-ops.
