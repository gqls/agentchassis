# 016 — Debugging Guide — Addendum: Tool pages render a description but no working widget

Slots into 016 after the page-build-handler log-trail section. Companion working doc: `PLAN_tool_widget_clobber.md`. Cross-refs: `029` (canonicalisation), `FOCUS_interactive_content_generation` (parse-stage gap).

---

## Symptom

A `page_type='tool'` page deploys with a hero, a prose `generic-text-block` describing the tool, and a `tool-list`, but **no interactive widget** — no `<script>`, no tool `data-component` in the body. The page reads as an article *about* the tool rather than the tool itself. Hero CTAs often point at generic leaked defaults (`/contact.html`, `/services.html` — the `B-029-2` symptom).

First seen: gamesdesign.co.uk `/tools/drop-rate-simulator/`, `/tools/jump-physics/` (2026-05-26).

---

## Assumption-checklist item (add to §0)

**19. "No widget on a tool page" is not necessarily a clobber — it may never have been generated.** Two distinct causes need different fixes; confirm which before touching code.

- **M1 — clobber.** A widget *was* created (via `create_tool_component`/`deploy_tool`, position 2, `build_status='deployed'`) and a later page rebuild deleted it. The clobber is real: `SavePageSectionsAction` does `DELETE FROM page_components WHERE page_id=$1` then re-inserts only the writer's sections, and the widget isn't in the section list it rebuilds from. Its text-only regression guard can't see a script-heavy widget, so it doesn't trip.
- **M2 — never generated.** The page came from **adoption recreate**. Adoption captures source *text* but not interactive JS (no parse stage), and nothing triggers the tool pipeline for adopted tool pages — so no widget was ever created. The "description not tool" symptom is the absence of generation, not a deletion.

Do not infer M1 from the rendered HTML alone. Run the three-query disambiguation below.

---

## Diagnostic recipe (read-only; ~30 seconds)

`content_components` has **no** `site_id`; scope tools via `page_components → pages.site_id` or the domain-slug suffix in `content_components.name`. `site_work_items` work category is `pipeline`, not `domain` (see 016 schema reminders).

```sql
-- D) Does a tool widget exist for the site at all (linked or orphaned)?
SELECT cc.function, cc.created_from,
       cc.html_template LIKE '%<script%' AS has_script,
       EXISTS (SELECT 1 FROM page_components pc
               JOIN pages p ON p.id = pc.page_id
               JOIN sites s ON s.id = p.site_id
               WHERE pc.component_id = cc.id AND s.domain = '<domain>') AS linked
FROM content_components cc
WHERE cc.component_level = 'tool' AND cc.is_active = true
  AND cc.name LIKE '%<domain-slug>%';

-- E) Did the tool pipeline run, or is it adoption-only?
SELECT wi.item_type, wi.handler_agent, wi.status, LEFT(wi.summary, 60) AS summary
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE s.domain = '<domain>'
  AND (wi.item_type IN ('evaluate_tools','add_tool','needs_content_page')
       OR wi.handler_agent IN ('tool-suggester','tool-generator','tool-deployer'))
ORDER BY wi.created_at DESC LIMIT 50;

-- F) Was a widget ever snapshotted before a delete? (only non-empty under M1)
SELECT pch.created_at, LEFT(pch.content_data::text, 120) AS snapshot
FROM page_component_history pch
JOIN pages p ON p.id = pch.page_id
JOIN sites s ON s.id = p.site_id
WHERE s.domain = '<domain>' AND p.name = '<tool-page-name>'
  AND pch.source = 'save_page_sections_overwrite'
ORDER BY pch.created_at DESC;
```

**Interpretation:**

| D | E | F | Verdict |
|---|---|---|---|
| script-bearing row, `linked=false` | `needs_content_page` from `tool-generator`/`tool-deployer` present | rows present | **M1** — widget existed, build deleted it. Recover from F. |
| no rows | only `page-build-handler` "Recreate …" items | no rows | **M2** — widget never generated (adoption only). |

---

## gamesdesign.co.uk result (2026-05-26) → M2

- **D: 0 rows.** No `component_level='tool'` component for the site — not linked, not orphaned.
- **E: adoption only.** Every item is `needs_content_page` from `page-build-handler`, summary "Recreate *X* page from gamesdesign.co.uk". No `evaluate_tools`/`add_tool`, no `tool-*` handlers.
- **F: 0 rows.** Nothing ever snapshotted → nothing was ever deleted.

The widget was never generated. The page is a faithful adoption recreate of captured text. This also explains why `pages.sections = []` survives: the recreate path feeds the writer the adopted content directly (`load_existing_content`, mode=recreate), so the `site_plan`/`pages.sections` authority is never populated for these pages.

---

## Root cause (M2) — CORRECTED after verification

An earlier draft of this entry said "no agent owns widget generation." That was **checked and found false** — do not repeat it. What actually exists:

- `apply_adoption_plan` routes adopted pages by interactivity: `if len(page.Features) > 0` → `needs_tool_recreation` / **`tool-recreation-handler`**; else → `needs_content_page` / `page-build-handler`.
- **`tool-recreation-handler`** is a real registered agent that LLM-recreates the widget (`recreate_tool`, marker `<!-- tool-recreation-complete -->`), gates on `check_tool_completeness`, and deploys via `page-rerender`.

So the widget *should* have been generated. The actual fault is a **misroute**: gamesdesign tool pages had `len(page.Features) == 0`, so adoption sent them down the static `page-build-handler` path (query E confirms: all `needs_content_page`, no `needs_tool_recreation`).

**Why `Features` was empty (prime suspect, structural — same class as `029`):** `buildPageFeatureMap` keys the feature map by the **raw** `fm["page"]` the adoption LLM wrote; the routing loop looks up `pageFeatures[pageName]` where `pageName` is **canonicalised** (`CanonicalisePage` adds a `tool-` prefix for tools). The key the map stored (`drop-rate-simulator`) never equals the key looked up (`tool-drop-rate-simulator`), so the lookup misses for every tool page even when the LLM detected the tool. Two sub-cases, distinguished by query G:

- **b1 — key desync** (LLM emitted features; canonicalised lookup misses). Definite bug regardless.
- **b2 — nothing emitted** (adoption analysis produced no `interactive_features`). Then the detection prompt is the target; the key fix alone won't help.

**RESOLVED 2026-05-26 → b1.** Query G (`rr.findings`) returned `interactive_features[].page` keys for all six tools (bare, e.g. `drop-rate-simulator`) and five games (prefixed, e.g. `game-p2p-networking`). The `tool` branch of `CanonicalisePage` adds `tool-`, so tool lookups miss; the `game` branch keeps the existing `game-`, so games match. The desync misroutes exactly the roles whose canonical prefix differs from the raw key. Fix: key the feature map by the canonical name in `buildPageFeatureMap` (it already receives `plan`, so it can resolve each feature's role from `plan["pages"]` and reuse `CanonicalisePage`).

Separately, `check_missing_tools` exists as a *site-level* `evaluate_tools → tool-suggester` trigger (0 tools → 7-day cooldown). It is **not** what fixes adopted tool pages: it counts tools site-wide (and via an INNER JOIN that can't see a tool *page* lacking a widget), and `tool-suggester` invents tools by vertical judgment rather than recreating *this* page's widget.

---

## Potential solutions

Ordered. The first is the M2 fix; the rest are surrounding structural cleanup.

1. **Fix the interactivity misroute (the handoff already exists).** Make detected tools actually reach `tool-recreation-handler`. Prime fix: canonicalise the feature-map key in `buildPageFeatureMap` (or match the raw name before `CanonicalisePage`), so `pageFeatures[pageName]` resolves for tool pages instead of missing on the `tool-` prefix. **Gate on query G first** — if adoption emitted no `interactive_features` at all (b2), fix the adoption detection prompt instead; the key fix won't help. Confirm `tool-recreation-handler` is `active` with a current image.
2. **Detection check (safety net).** A discovery check: `page_type='tool'` page with zero `component_level='tool'` `page_components` → emit a recreate/`add_tool` item. Closes the `check_tool_health` / `check_missing_tools` INNER-JOIN blind spot, which can't see a tool *page* missing its widget.
3. **Canonicalise tool page identity everywhere (overlap fix).** Route `create_tool_component` and `deploy_tool` page identity through `datahelpers.CanonicalisePage` (the `029` Phase-0 helper) so the three surfaces stop diverging. `create_tool_component` is the older surface and currently emits `/tools/<function>.html` with the `tool-` prefix embedded; it should take the canonical `/tools/<slug>/index.html`. This is the same canonicalisation discipline whose *absence* in `buildPageFeatureMap` caused the misroute. Tracked in `PLAN_tool_widget_clobber.md` (T3).
4. **Then fix the M1 clobber, or the recreated widget gets wiped.** Once `tool-recreation-handler` produces a widget, a later page rebuild will delete it unless the widget is a first-class section in `site_plan` (which `load_page_sections_from_spec` syncs into `pages.sections`). See `PLAN_tool_widget_clobber.md` §5.
5. **Longer-term: adoption parse stage.** Capture/regenerate the source widget's behaviour during the crawl rather than relying on full LLM recreation (`FOCUS_interactive_content_generation`, Path D).

---

## One-line takeaway

A tool page showing a description but no widget on an **adopted** site is most likely M2 (never generated) — but not because the responsibility is unowned. `tool-recreation-handler` owns it; the page was **misrouted** to `page-build-handler` because adoption's `Features` came back empty (prime suspect: `buildPageFeatureMap` keys by the raw page name while the route looks up the canonicalised one). Confirm with D/E/F, then G, before fixing.
