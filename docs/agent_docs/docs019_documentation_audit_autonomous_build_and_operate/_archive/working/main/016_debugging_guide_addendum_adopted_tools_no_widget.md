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

## Root cause (M2)

A responsibility seam with no owner:

- **Adoption (recreate)** owns reproducing captured content. It did its job — it reproduced the text. It cannot capture interactive JS and does not hand off to anything that can.
- **Tool pipeline** (`tool-suggester` → `tool-generator`/`tool-deployer`) owns building widgets. Nothing triggered it for this site.
- **No agent owns** "this is a tool page; it must have a working widget; if adoption couldn't capture one, generate one." The widget falls down the gap.

Separately, page **identity** is built by three surfaces independently — `apply_adoption_plan`, `sync_pages_to_db` (planner), and `create_tool_component` — producing divergent URLs (`/tools/drop-rate-simulator/index.html` vs `create_tool_component`'s `/tools/tool-drop-rate-simulator.html`). `029` routed the first two through `CanonicalisePage`; the tool actions were never added. That's overlap producing divergence, alongside the widget gap.

---

## Potential solutions

Ordered; first two are the M2 fix, the rest are the surrounding structural cleanup.

1. **Adoption → tool-pipeline handoff (owns the gap).** When adoption classifies a source page as `page_type='tool'`, emit an `add_tool` / `evaluate_tools` work item (or a flag the tool pipeline reads) so `tool-generator` produces a widget. Clear ownership: adoption *emits*, tool-generator *owns generation*. Keeps responsibilities distinct (no widget logic inside adoption).
2. **Post-adoption detection check (safety net).** A discovery check that finds `page_type='tool'` pages with zero `component_level='tool'` `page_components` and emits `add_tool`. This also closes the `check_tool_health` blind spot (its INNER JOIN `content_components → page_components` reports "no tools" and passes when a tool page has no linked tool component).
3. **Canonicalise tool page identity (overlap fix).** Route `create_tool_component` and `deploy_tool` page name/url/page_type through `datahelpers.CanonicalisePage` so all three surfaces agree. `create_tool_component` is the older surface and currently builds `/tools/<function>.html` with the `tool-` prefix embedded — change it to pass the bare slug + `Role="tool"` and take the canonical `/tools/<slug>/index.html`. Tracked as a task in `PLAN_tool_widget_clobber.md`.
4. **Then fix M1, or the new widget gets wiped.** Once a widget is generated for an adopted tool page, the next rebuild will clobber it unless the widget is a first-class section in `site_plan` (which `load_page_sections_from_spec` syncs into `pages.sections`). M2 and M1 are sequential: generate the widget, then make it survive. See `PLAN_tool_widget_clobber.md` §5.
5. **Longer-term: adoption parse stage.** Capture/regenerate the source widget's behaviour during adoption rather than only its text (`FOCUS_interactive_content_generation`, Path D). Reduces reliance on regeneration when the source already had a good tool.

---

## One-line takeaway

A tool page showing a description but no widget on an **adopted** site is most likely M2 (never generated), not a clobber — confirm with queries D/E/F before assuming, because the fix is "generate + hand off," not "stop the delete."
