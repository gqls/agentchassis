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

### 2.4 The build's section authority is the site_plan, and it overwrites pages.sections (CONFIRMED in code)

`load_page_sections_from_spec_action.go` (`LoadPageSectionsFromSpecAction`), read 2026-05-26:

- It queries `site_specs WHERE aspect='site_plan' AND is_current=true`, finds the page entry by `name` inside `site_plan.pages[]`, and reads that entry's `sections`. This is the authoritative source.
- **It then syncs the plan's sections back into the `pages` table**: `UPDATE pages SET sections = $1 WHERE site_id=$2 AND name=$3 AND sections::text IS DISTINCT FROM $1`. So `pages.sections` is *overwritten from the plan* on every build where the plan has an entry.
- Only if the plan yields no sections does it fall back to `page_record.sections`, then `pages.sections`.

**Consequence for the fix:** setting `pages.sections` inside the tool action is futile when a plan entry exists — the next build's sync wipes it. The tool section must live in `site_plan` (planner/reconciler) or tool pages must be exempted from this sync.

**Puzzle this raises:** we observed `pages.sections = []`, yet a build produced three sections. If the plan had supplied them, the sync would have written them into `pages.sections` (it would not be empty). So either the plan has **no** entry for this page (and the three sections came from elsewhere — see §2.5), or a later process reset `pages.sections`. Unresolved until the §7 queries run. Do not assume the plan listed `[hero, generic-text-block, tool-list]`.

### 2.5 These pages are adoption-shaped — a second mechanism is in play

Doc `029` documents the gamesdesign.co.uk run specifically. The page name/url we queried (`tool-drop-rate-simulator`, `/tools/drop-rate-simulator/index.html`) is the **adoption-canonical** shape from `029`'s canonical table — **not** what `create_tool_component` produces (it would make `/tools/tool-drop-rate-simulator.html`). The hero CTAs we saw ("Launch Simulator" → `/contact.html`, "Browse Tools" → `/services.html`) match `B-029-2` exactly: leaked generic-component defaults on a page built from the plan with no real brief.

So there are **two distinct candidate mechanisms**, and which one applies to these pages is not yet known:

- **(M1) Clobber** — a widget was created via `create_tool_component`/`deploy_tool` and the content build deleted it (§2.1–2.3). This is a confirmed *defect in the code path*; whether it fired for these pages depends on whether a tool component was ever created for them.
- **(M2) Never generated** — these pages came from **adoption**, which captures text but not interactive JS (the "no parse stage" gap in `FOCUS_interactive_content_generation`). The planner then planned them as generic content pages, and **no widget was ever created** to clobber. The "description not tool" symptom would then be the absence of generation, not a deletion.

M1 and M2 are not mutually exclusive across the six tools. The §7 decisive queries disambiguate: if a tool component (script-bearing) exists for the site — linked or orphaned — M1 is in play; if none exists, M2 is the cause for those pages.

---

## 3. Root cause (defect confirmed; applicability to these pages pending §7)

A confirmed defect exists in the create/deploy → build path (M1): the tool widget is written into `page_components` as a side-effect, but the build pipeline **rebuilds `page_components` by DELETE+INSERT from a section list** whose authority is `site_specs.site_plan` (synced into `pages.sections`, falling back to it). The widget is not a member of that list, so the first `needs_content_page` build deletes it.

Two independent writers, one of which is authoritative and destructive:

1. **Create/deploy action** → side-writes the widget row (not in any section list).
2. **page-build-handler** (triggered by the action's own `needs_content_page` item) → rebuilds the page from the section list, wiping anything not in it.

The novel path (`create_tool_component`) is worst-hit: it never registers the tool in `pages.sections`, and (per §2.4) the sync would overwrite it anyway.

**Caveat (see §2.5):** the gamesdesign pages are adoption-shaped, so M2 (widget never generated) is an equally live cause for *these specific pages*. The §7 queries decide whether what we're looking at is M1, M2, or a mix — and that decision changes the fix from "stop the clobber" to "make the widget get generated/parsed in the first place."

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

- [x] **Read `load_page_sections_from_spec` — DONE (2026-05-26).** `site_plan` is authoritative *and* its sections are synced back into `pages.sections` on every build (§2.4). Fix must target `site_plan` (planner/reconciler) or exempt tool pages from the sync; setting `pages.sections` alone is futile.
- [ ] **GATING: clobber (M1) vs never-generated (M2) for these pages.** Does a script-bearing `component_level='tool'` component exist for gamesdesign (linked or orphaned)? Query D below. This is now the decision that selects the fix.
- [ ] Inspect `site_specs.site_plan` (`is_current = true`) for gamesdesign: is there an entry for `tool-drop-rate-simulator` (or `drop-rate-simulator`), and with what `sections`? Resolves the §2.4 puzzle and tells us what the planner emits for tool pages. Query A.
- [ ] Did the tool pipeline run for this site at all (`evaluate_tools`/`add_tool`, `tool-suggester`/`tool-generator`/`tool-deployer`)? Query E. Distinguishes "pipeline ran, widget clobbered" from "adoption only, never generated".
- [ ] Are there duplicate `pages` rows per tool slug (the `029` divergence: `tool-X` vs `X`)? Query C.
- [ ] Confirm a known **fork-path** tool elsewhere (e.g. gaswholesalers "Gas Unit Converter") has its widget present in deployed HTML — does the fork path survive today?
- [ ] Confirm `RenderComponentAction` emits a `render_mode='standalone'` component's `html_template` into `sections_metadata` so the widget re-inserts on save when it's a listed section.

---

## 7. Verification queries (read-only; schema verified 2026-05-26)

`content_components` has **no** `site_id`; scope tools via `page_components → pages.site_id`, or via the domain-slug suffix in `content_components.name` (the create/deploy naming convention, e.g. `…-gamesdesign-co-uk`).

**Decisive (run first): does a widget exist for this site?**

```sql
-- D) Any tool-level component for gamesdesign — linked or orphaned.
--    has_script=true on an UNLINKED row  => M1 (clobber): widget existed, build deleted it.
--    no rows at all                      => M2: widget never generated (adoption gap / suggester never ran).
SELECT cc.function, cc.name, cc.created_from,
       cc.forked_from IS NOT NULL              AS is_fork,
       cc.html_template LIKE '%<script%'       AS has_script,
       length(cc.html_template)                AS tmpl_len,
       EXISTS (
         SELECT 1 FROM page_components pc
         JOIN pages p ON p.id = pc.page_id
         JOIN sites s ON s.id = p.site_id
         WHERE pc.component_id = cc.id AND s.domain = 'gamesdesign.co.uk'
       )                                        AS linked_to_site_page
FROM content_components cc
WHERE cc.component_level = 'tool'
  AND cc.is_active = true
  AND cc.name LIKE '%gamesdesign%'
ORDER BY cc.created_from, cc.function;

-- E) Did the tool pipeline run for this site at all?
--    needs_content_page items from tool-generator/tool-deployer => M1 path executed.
--    only adoption items, no tool-* handlers                    => M2 (adoption only).
SELECT wi.item_type, wi.handler_agent, wi.status, wi.created_at, LEFT(wi.summary, 80) AS summary
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE s.domain = 'gamesdesign.co.uk'
  AND (wi.item_type IN ('evaluate_tools', 'add_tool', 'needs_content_page')
       OR wi.handler_agent IN ('tool-suggester', 'tool-generator', 'tool-deployer'))
ORDER BY wi.created_at DESC
LIMIT 50;
```

**Context queries:**

```sql
-- A) site_plan entry for this page — robust to tool- prefix. Resolves the §2.4 puzzle.
SELECT jsonb_path_query(ss.data, '$.pages[*] ? (@.name like_regex "drop-rate")')
FROM site_specs ss
JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'gamesdesign.co.uk'
  AND ss.aspect = 'site_plan'
  AND ss.is_current = true;

-- C) Duplicate page rows per tool slug (the 029 divergence: tool-X vs X).
SELECT p.name, p.url, p.page_type, p.build_status, p.sections
FROM pages p
JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'gamesdesign.co.uk'
  AND (p.name LIKE '%drop-rate%' OR p.url LIKE '%drop-rate%');

-- F) Recover a wiped widget from history (only meaningful if M1).
SELECT pch.created_at, LEFT(pch.content_data::text, 120) AS snapshot
FROM page_component_history pch
JOIN pages p ON p.id = pch.page_id
JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'gamesdesign.co.uk'
  AND p.name = 'tool-drop-rate-simulator'
  AND pch.source = 'save_page_sections_overwrite'
ORDER BY pch.created_at DESC;
```

Post-fix acceptance: a freshly built tool page has a `page_components` row with `component_level='tool'` carrying a `<script>` in `rendered_html`, and `site_plan` (hence `pages.sections`) lists that section.

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
- **2026-05-26 (step 2)** — Read `load_page_sections_from_spec`: `site_plan` is authoritative and **overwrites `pages.sections`** on every build → fix must target the planner/reconciler, not the action (§2.4). Read `029`: these pages are **adoption-shaped**, and the leaked `/services.html`/`/contact.html` CTAs match `B-029-2`. Raised a second mechanism **M2 (widget never generated via adoption)** alongside **M1 (clobber)** (§2.5). New gating question: M1 vs M2, decided by queries D/E. Next: run §7 queries D, E, A, C.
