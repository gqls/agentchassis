# PLAN — pathfinding game page has no interactive surface (gamesdesign.co.uk)

**Status:** diagnosis-first. Do NOT trigger a recreation/rebuild for pathfinding until the candidate is pinned — the prior investigation showed blind re-triggering reproduces the empty result.

## Symptom (2026-06-14)
`https://gamesdesign.co.uk/games/pathfinding/index.html` renders the hero + a "How A* Actually Works" prose section and **no interactive game** — no `<canvas>`, no game `<script>`. The games-hub card promises the A* simulator ("Paint walls and terrain costs…"); the page doesn't contain it.

**Decisive asymmetry:** the OTHER games (auto-battler, economy-simulator, jelly-invaders, p2p-networking) DO still have their interactive surface (user-confirmed). So this is NOT the systemic recreation-loss defect that hit all five games in the June investigation — it is isolated to pathfinding. That narrows the candidates sharply (see §3).

**Relevant coincidence:** pathfinding was one of the §5 linking-rebuild targets (hero phantom + it's a game page). A `page-build-handler` rebuild runs `save_page_sections` (delete-and-reinsert). If pathfinding's widget was a `page_components` row not represented in the rebuild's section set, the rebuild could have CLOBBERED it. This makes candidate (b) the leading hypothesis — but confirm, don't assume.

## 1. This is the SAME problem we fixed before — prior art to reuse
From the May–June 2026 investigation (do not re-derive; reuse):
- `PLAN_tool_widget_clobber.md` — full investigation, queries A–R3, changelog steps 1–9.
- `HANDOFF_2026-05-26_tool_routing_fix_deployed.md` — what T1 fixed and what was left open.
- `016_debugging_guide_addendum_adopted_tools_no_widget.md` — assumption checklist + D/E/F diagnostic recipe + root cause.

Root-cause chain established then:
1. **Parse-stage loss:** adoption crawl captured markdown only, not JS (`007_adoption_pipeline_v4`). Interactive logic was lost before anything downstream could use it. FIXED: crawl config now requests `formats:["markdown","rawHtml"]`; a recrawl gets the JS.
2. **Vocabulary gap:** `page_type='game'` wasn't in the classifier vocabulary, so game pages didn't survive adoption. FIXED: `game` added.
3. **T1 routing (DEPLOYED):** `apply_adoption_plan_action.go` `buildPageFeatureMap` routes interactive pages → `tool-recreation-handler` (mode `recreate_tool`); discovery check `check_tool_recreation_needed.go`; canonical join `datahelpers/page_canonical.go` `CanonicalisePage`.
4. **Recreation-loss (T1 necessary, NOT sufficient):** correctly-routed games completed but had no widget. Candidate mechanisms (a)-(d) below. For the all-games-empty case this was never fully closed; for THIS single-page case the asymmetry rules most of them out.
5. **Clobber class, prior partial fix:** `save_page_sections_patch.go` added `findPreservedComponentIDs` — preserves `page_components` whose `content_components.input_schema` has `render_action` (e.g. latest-news). If the game widget's component does NOT carry `render_action`, this preservation does NOT cover it → a rebuild can still drop it. This is the most likely mechanism for a single clobbered game.

## 2. tool-recreation-handler shape (from prior art, confirm before relying)
Workflow: `recreate_tool → check_tool_completeness → validate_tool → save_page_sections → update_page_status → spawn_rerender`. Mode `recreate_tool` (NOT `recreate` — `load_existing_content` skips unless mode matches; this was a prior gotcha). Target page resolution is the suspect surface for candidate (a).

## 3. Candidate mechanisms — and what the asymmetry implies
Because the sibling games work, the field narrows to:
- **(b) Clobber — LEADING.** Widget existed on `games/pathfinding`, then a later rebuild (plausibly the §5 link-resolution rebuild) ran `save_page_sections` delete-and-reinsert and dropped it because the widget component wasn't in the new section set and isn't `render_action`-preserved. Evidence = `page_component_history` snapshot with `source='save_page_sections_overwrite'` (or similar) for this page.
- **(c) Never persisted for this page only.** Pathfinding's recreation specifically failed/timed-out while siblings succeeded. Evidence = the `needs_tool_recreation` item for pathfinding shows non-complete, or complete with no resulting widget and no clobber history.
- **(a) Mis-target — LESS LIKELY (would likely affect siblings too).** A `tool-game-pathfinding` twin holds the widget while the real page shows description. Evidence = twin page with interactive markup.
- **(d) Snippets false-negative — RULE OUT FIRST (cheapest).** The widget IS present but its `<script>` was extracted to `/assets/js/snippets.js`, so "no script in page HTML" misleads. The rendered page we have shows NO `<canvas>` and no widget container at all (not just missing script), so (d) is unlikely — but the deployed-snippets cross-check is one query and worth doing before any fix.

## 4. Diagnostic queries (scoped to pathfinding vs a working sibling) — RUN THESE FIRST
Adapted from the prior investigation's Steps A/B/C. Use `economy-simulator` as the known-good comparator (any working game serves).

**Q1 (decisive — page dump, pathfinding vs working sibling):**
```sql
SELECT p.name AS page, p.build_status, p.page_type, pc.position, pc.slot_name,
       cc.component_level, cc.function,
       length(pc.rendered_html) AS html_len,
       (pc.rendered_html LIKE '%<script%')                 AS inline_script,
       (pc.rendered_html ~* '<(canvas|form|input|button)') AS interactive_markup,
       LEFT(pc.rendered_html, 70) AS html_start
FROM pages p
JOIN sites s ON s.id = p.site_id
LEFT JOIN page_components pc ON pc.page_id = p.id
LEFT JOIN content_components cc ON cc.id = pc.component_id
WHERE s.domain='gamesdesign.co.uk'
  AND p.name IN ('game-pathfinding','pathfinding','game-economy-simulator','economy-simulator')
ORDER BY p.name, pc.position;
```
Reading: working sibling shows a row with `interactive_markup=t` (and a `component_level` of tool/game, or a widget container). If pathfinding LACKS that row but the sibling HAS it → widget is absent from pathfinding's components (→ clobber or never-persisted; Q3 disambiguates). If pathfinding HAS interactive markup but no inline `<script>` → candidate (d), check snippets. (Note: confirm pathfinding's actual page `name` — it may be `game-pathfinding` or `pathfinding`; the §4 audit listed it as `games/pathfinding` URL.)

**Q2 (mis-target — is there a twin?):**
```sql
SELECT p.name, p.url, p.page_type, p.build_status,
       wi.handler_agent, wi.created_by, wi.source, wi.status
FROM pages p
JOIN sites s ON s.id = p.site_id
LEFT JOIN site_work_items wi ON wi.page_id = p.id
WHERE s.domain='gamesdesign.co.uk'
  AND (p.name LIKE '%pathfinding%')
ORDER BY p.name;
```
A `tool-game-pathfinding` (or `tool-pathfinding`) row with `page_type=tool` holding the widget → candidate (a).

**Q3 (clobber evidence — decisive between b and c):**
```sql
SELECT pch.created_at, pch.source, length(pch.content_data::text) AS snap_len,
       (pch.content_data::text ~* '<(canvas|script|form|input|button)') AS snap_had_widget,
       LEFT(pch.content_data::text, 90) AS snap_start
FROM page_component_history pch
JOIN pages p ON p.id = pch.page_id
JOIN sites s ON s.id = p.site_id
WHERE s.domain='gamesdesign.co.uk' AND p.name IN ('game-pathfinding','pathfinding')
ORDER BY pch.created_at DESC;
```
A snapshot whose `snap_had_widget=t` with `source` indicating a save_page_sections overwrite, dated around the §5 rebuild → **(b) clobber confirmed**, and it identifies the rebuild as the culprit. No history with a widget at all → **(c) never persisted**.

**Q4 (snippets cross-check — rule out d cheaply):**
```sql
-- Does the deployed snippets bundle contain the pathfinding game logic?
-- (the widget's <script> may be extracted there). Inspect the deployed file:
--   curl -s https://gamesdesign.co.uk/assets/js/snippets.js | grep -i 'astar\|pathfind\|grid\|heuristic'
-- If the logic is present in snippets.js AND Q1 shows a widget container on the page,
-- the page may actually be interactive — re-examine the live page. If snippets.js
-- lacks it and the page has no container, (d) is ruled out.
```

## 5. Fix — decided by §4 outcome, not before
- **(b) clobber confirmed (Q3):** the fix is the clobber class, already partly designed. Either (i) make the game widget a first-class section in `site_plan` so `load_page_sections_from_spec` carries it through a rebuild (so `save_page_sections`'s reinsert includes it), OR (ii) extend `findPreservedComponentIDs` to preserve `component_level IN ('tool','game')` rows, not only `render_action` ones — so a content/linking rebuild can never drop a widget. (ii) is the smaller, safer change and directly prevents recurrence (incl. for the tools). See `PLAN_tool_widget_clobber.md` §5 (T4). Then re-run recreation for pathfinding ONCE to restore its widget, and confirm it survives a subsequent rebuild.
- **(c) never persisted (Q3 empty):** re-trigger recreation for pathfinding via `tool-recreation-handler` (mode `recreate_tool`), watch the `recreate_tool → check_tool_completeness → validate_tool → save_page_sections → spawn_rerender` chain + pod logs for where it fails; this is a single-page handler-side failure, not systemic.
- **(a) mis-target (Q2 twin):** fix where `tool-recreation-handler` resolves the target page (canonicalisation), redirect/remove the twin. Lower likelihood given siblings are fine.
- **(d) false-negative (Q4):** nothing to fix; correct the observation.

## 6. Restore + verify (after the fix)
1. Trigger recreation for pathfinding (mode `recreate_tool`) — single work item, like the §5 pattern but `handler_agent='tool-recreation-handler'`.
2. Confirm the deployed page renders the A* widget (canvas + controls).
3. **Regression guard (the point of the clobber fix):** run a content/linking rebuild on pathfinding AGAIN and confirm the widget SURVIVES — this proves (b) is actually fixed, not just papered over by re-recreation.
4. Re-check the other four games still have their widgets (the fix must not regress them).

## 7. Sequencing vs the linking batch
Independent of §5/§6 of the linking runbook. Do the linking close-out first (it's nearly done — two retries + audit). Then this. If the clobber fix (5(b)(ii)) lands, it ALSO protects every future linking/content rebuild from dropping widgets — so there's an argument to prioritise it before any further bulk rebuilds (e.g. the readopt), to avoid re-clobbering the working games.
