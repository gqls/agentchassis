# PLAN — Tool widget missing from deployed tool pages

**Status:** root cause confirmed for the novel-tool path; fix direction drafted, not implemented.
**Opened:** 2026-05-26
**Scope:** tool pages render a description but no interactive widget (e.g. gamesdesign.co.uk `/tools/drop-rate-simulator/`, `/tools/jump-physics/`).
**Constraint:** the adoption route touches this same build path in a parallel chat. Pull current versions of all named files before editing and coordinate. No variable renames except where explicitly intended and noted.

---

## 1. Symptom

Deployed tool pages contain a hero, a prose description (`section--generic` / `generic-text-block`), and a `tool-list`, but **no interactive widget** (no `<script>`-bearing tool section, no `data-component` for the tool). The page reads as a generic content page about the tool rather than the tool itself. The hero CTA ("Launch Simulator") points at `/contact.html`, which is a generic content-page pattern — a tell that the writer treated this as an ordinary page.

Observed identically across multiple tool pages → systematic pipeline behaviour, not per-tool LLM variance.

---

## 2. Confirmed findings (with evidence)

### 2.1 The widget is created, then deleted

For `tool-drop-rate-simulator` on gamesdesign.co.uk:

```
pages.sections = []          -- empty
page_components:
  pos 1  hero                (section)
  pos 2  generic-text-block  (section)   <- id 8d81e665…, the description
  pos 3  tool-list           (section)
```

No `component_level='tool'` row remains on the page. The widget that `create_tool_component` inserted at position 2 (with `build_status='deployed'`) is gone, and `generic-text-block` occupies position 2.

### 2.2 The clobber is a full DELETE + re-INSERT

`save_page_sections_action.go`, in `SavePageSectionsAction`:

```go
// Clear existing components for this page
DELETE FROM page_components WHERE page_id = $1   // line ~295
```

…then re-inserts only the sections handed to it by the content writer (positions renumbered `i+1`). Any `page_components` row not represented in the writer's output — including a side-written tool widget — is destroyed.

- Old content **is** snapshotted to `page_component_history` (`source = 'save_page_sections_overwrite'`) before the delete → widgets are recoverable.
- There is a **content-regression guard** (lines ~223–259) but it compares *visible text length* after stripping tags. A widget is mostly `<script>`/`<style>`/form markup with little visible text; the replacement (hero + prose + tool-list) has *more* text, so the guard does not trip. The guard is structurally blind to tools.

### 2.3 The two creation paths diverge

| | Novel path (`create_tool_component_action.go`) | Fork path (`deploy_tool_action.go`) |
|---|---|---|
| Sets `pages.sections`? | **No** — INSERT omits the column → defaults to `[]` | **Yes** — `["hero-tool","tool-guide-intro","<toolFunction>","tool-cta"]` |
| Tool is a first-class section? | No | Yes (the function name is in the list) |
| Side-writes `page_components` pos 2? | Yes (`build_status='deployed'`) | Yes (`build_status='deployed'`) |
| Queues `needs_content_page`? | Yes | Yes |
| `created_from` | `'generated'` | fork (`forked_from` set) |

gamesdesign's tools were created via the **novel path** (`created_from='generated'`), the more exposed of the two.

### 2.4 The build's section authority is the site_plan, not pages.sections

The page built three sections even though `pages.sections = []`. They therefore came from the page-build-handler step `load_spec_sections` (`load_page_sections_from_spec`), described as: *"Load sections from `site_specs.site_plan` (authoritative), fall back to `pages.sections`."* So the **site planner's `site_plan`** planned this tool page as `[hero, generic-text-block, tool-list]` with no widget, and that plan is what the build rebuilt from.

---

## 3. Root cause

The tool widget is written into `page_components` as a side-effect of the create/deploy action, but the build pipeline **rebuilds `page_components` by DELETE+INSERT from a section list** whose authority is `site_specs.site_plan` (falling back to `pages.sections`). The widget is not a member of that list, so the first `needs_content_page` build deletes it.

Two independent writers, one of which is authoritative and destructive:

1. **Create/deploy action** → side-writes the widget row (not in any section list).
2. **page-build-handler** (triggered by the action's own `needs_content_page` item) → rebuilds the page from the section list, wiping anything not in it.

The novel path is worst-hit: it never registers the tool in `pages.sections`, and the `site_plan` the build actually reads doesn't contain the widget either.

---

## 4. The design question (raised, and load-bearing)

> *"What is to say that the page wants these sections in the first place? It may not."*

Correct, and central. Both actions *assume* a tool page wants hero + guide-intro + CTA and queue a `needs_content_page` build to add them. But:

- That assumption is unverified. A tool page may want only the widget (plus a lean hero), with the educational content living in the **companion guide page**, which the same action already creates.
- The mechanism chosen to "add content around the tool" is the very thing that destroys the tool, because the widget isn't in the authoritative section list the build rebuilds from.
- The `content_guidance` literally says *"Do NOT regenerate the tool widget — it is already deployed at position 2,"* but the writer has no ability to honour that: `save_page_sections` deletes position 2 regardless.

So the canonical shape of a tool page is an explicit planning decision we need to make, not a default to inherit. This decision drives which fix below we adopt.

---

## 5. Fix options (structural-first; not yet implemented)

Guiding principle: **whatever the build rebuilds from must contain the tool widget as a first-class section.** Stop relying on a side-written row that a later rebuild can't see.

### Option 1 — Make the widget a first-class section in the authority *(preferred)*
- Decide the canonical tool-page section list (see Option 2 — they're paired).
- Ensure the **authority the build actually reads** carries the widget:
  - the planner emits a tool/embed section for `page_type='tool'` pages in `site_plan`, **or**
  - tool pages are made to prefer `pages.sections` over `site_plan`, **and**
  - the novel path sets `pages.sections` to include the tool function (matching what the fork path already does).
- Confirm `plan_sections` marks a standalone tool section "ready" (it should: `render_mode='standalone'`, empty `input_schema`, no `llm` fields) and that `RenderComponentAction`/compile emit its `html_template` into `sections_metadata` so `save_page_sections` re-inserts it.
- Result: the widget flows through plan → render → compile → save like any other section; nothing to clobber.

### Option 2 — Right-size the tool page *(pairs with Option 1; answers §4)*
- Choose the tool-page shape deliberately. Candidates:
  - **Lean:** `["hero-tool", "<toolFunction>"]` — widget plus a short hero; deep content lives in the guide page.
  - **Current fork shape:** `["hero-tool", "tool-guide-intro", "<toolFunction>", "tool-cta"]`.
- If a tool page does **not** want generic content, do not queue `needs_content_page` for it (or queue a minimal one). This removes the destructive build for parts the page doesn't want.

### Option 3 — Make `save_page_sections` structure-aware *(safety net, secondary)*
- Have the rebuild preserve/re-attach `page_components` whose linked `component_level='tool'`, **or** make the regression guard refuse to drop a `component_level='tool'` row unless the new section set also contains it.
- This is a guard against silent data loss, not the fix. Keep it, because two writers will keep colliding otherwise.

**Recommended:** Option 1 + Option 2 together (widget is a real section *and* the page shape is chosen on purpose), with Option 3 as a guard. Align the **novel path to the fork path**, then fix the shared `site_plan`-authority gap that exposes both.

---

## 6. Open questions to resolve before coding

- [ ] Read `load_page_sections_from_spec` — confirm the exact authority order and how a `page_type='tool'` page is selected from `site_plan`. **This decides whether the fix lives in the planner, the action, or both.**
- [ ] Inspect `site_specs.site_plan` (`is_current = true`) for gamesdesign: does it list `tool-drop-rate-simulator`, and with what section list? (Confirms whether the planner is the authority that dropped the widget.)
- [ ] Confirm a known **fork-path** tool elsewhere (e.g. gaswholesalers "Gas Unit Converter") actually has its widget present in deployed HTML — i.e. that the fork path survives today. If it does, that validates the novel-vs-fork divergence as the differentiator.
- [ ] Confirm `RenderComponentAction` emits a `render_mode='standalone'` component's `html_template` into `sections_metadata` (so the widget re-inserts on save when it's a listed section).

---

## 7. Verification queries (read-only; schema verified 2026-05-26)

`content_components` has **no** `site_id`; scope tools via `page_components → pages.site_id`.

```sql
-- A) site_plan section list for the tool page (where do the 3 sections come from?)
SELECT ss.aspect,
       jsonb_path_query(ss.data, '$.pages[*] ? (@.name == "tool-drop-rate-simulator")')
FROM site_specs ss
JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'gamesdesign.co.uk'
  AND ss.aspect = 'site_plan'
  AND ss.is_current = true;

-- B) Orphaned tool components (widget exists in content_components, not linked to any page)
SELECT cc.function, cc.created_from, cc.forked_from IS NOT NULL AS is_fork,
       length(cc.html_template) AS tmpl_len
FROM content_components cc
WHERE cc.component_level = 'tool'
  AND cc.is_active = true
  AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.component_id = cc.id);

-- C) Recover a wiped widget from history if needed
SELECT pch.created_at, LEFT(pch.content_data::text, 120) AS snapshot
FROM page_component_history pch
JOIN pages p ON p.id = pch.page_id
JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'gamesdesign.co.uk'
  AND p.name = 'tool-drop-rate-simulator'
  AND pch.source = 'save_page_sections_overwrite'
ORDER BY pch.created_at DESC;
```

Post-fix acceptance: a freshly built tool page has a `page_components` row with `component_level='tool'` carrying a `<script>` in `rendered_html`, and `pages.sections` / `site_plan` lists that section.

---

## 8. Secondary observations (note, do not fix in this pass)

- **`forked_from` NULL on novel tools.** `create_tool_component` omits `forked_from`, so generated tools are classified as *library* tools by the partial unique index `idx_cc_tool_function_unique (function) WHERE component_level='tool' AND forked_from IS NULL AND is_active`. Two sites generating the same function would collide. Latent; not today's bug.
- **`check_tool_health` blind spot.** Its INNER JOIN `content_components → page_components` means a tool with no linked `page_components` row (post-clobber) is invisible — the check reports "no tools" and passes. A detection check for "page_type='tool' page with zero `component_level='tool'` components" would have caught this.

---

## 9. Files in play (pull current revisions before editing)

| File | Role in the bug |
|---|---|
| `create_tool_component_action.go` | Novel path; does **not** set `pages.sections`; side-writes widget; queues `needs_content_page`. |
| `deploy_tool_action.go` | Fork path; sets `pages.sections` incl. tool function; otherwise same shape. |
| `save_page_sections_action.go` | DELETE+INSERT rebuild (the clobber); text-only regression guard. |
| `plan_sections_action.go` | Triages the section list; needs to keep a standalone tool section "ready". |
| `load_page_sections_from_spec` (not yet read) | Chooses section authority (`site_plan` vs `pages.sections`). |
| site-planner (agent_definition) | Emits the `site_plan` section list per page; likely needs a tool/embed section for `page_type='tool'`. |

---

## 10. Changelog

- **2026-05-26** — Investigation. Confirmed clobber via DB (`pages.sections=[]`, widget absent) and code (`DELETE FROM page_components`). Identified novel-vs-fork divergence and `site_plan` authority as root cause. Drafted fix options. Design question (does the page want these sections) folded in as a planning decision. Next: §6 open questions.
