# PLAN — Tool widget missing from deployed tool pages

**Status:** live cause for gamesdesign confirmed as **M2 (widget never generated — adoption only)** via queries D/E/F (2026-05-26). **M1 (clobber)** is a confirmed *latent* defect that does not explain these pages but would bite once M2 is fixed. Fixes drafted, not implemented. See debugging-guide addendum `016_debugging_guide_addendum_adopted_tools_no_widget.md`.
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

### 2.8 BLOCKING (2026-05-26, step 8): correct routing still yields no widget; new `tool-game-*` duplicates

Query J (target preview) + K (game survival) returned more than expected:

- **K: all five games are widget-less** (`has_widget_component=f`, `has_script_section=f`) despite having routed correctly to `tool-recreation-handler` and completed (step 6, query I). So **correct routing does not currently produce a deployed widget.** T1 (routing fix) is necessary but **not sufficient** — there is a second, active defect downstream of routing.
- **J also surfaced five `tool-game-*` pages** (`page_type=tool`, `build_status=planned`) — duplicates of the five games as tools. New since earlier reads. Something turned each game into a planned tool page.
- State is moving under us: `tool-drop-rate-simulator` is now `deployed` (was `needs_rebuild` in query C); some tools `planned`, some `deployed`. Coordinate with the parallel adoption chat.

**Caveat — K is suggestive, not proof.** `has_script_section=f` can be a false negative: the snippets mechanism extracts inline `<script>` into `/assets/js/snippets.js`, and a recreated widget may sit at a non-tool/game `component_level` or have a NULL `component_id`. The §7 page-component dump (query L) is decisive.

**Consequence:** do NOT trigger recreation/backfill yet — it would reproduce the games' empty result on the tools. Diagnose the recreation-doesn't-land problem first (queries L/M/N).

**Candidate mechanisms (to test, not concluded):**
- Recreation built a separate `tool-game-*` page instead of populating the existing `game-*` page (target/canonicalisation defect in `tool-recreation-handler`).
- A widget landed then was clobbered by a later rebuild (M1 — would show in `page_component_history`).
- `tool-recreation-handler` completed without persisting a widget (handler-side defect).
- Snippets false-negative — widgets actually present (K artifact).

### 2.6 CORRECTION (2026-05-26, step 4): the handoff and the generating agent DO exist

An earlier claim in this plan — "no agent owns ensuring a tool page has a working widget; if adoption can't capture one, generate one" — was **checked and found false**. Verified against `apply_adoption_plan_action.go`, `check_tool_completeness_action.go`, and `bk_agent_definitions_backup.sql`:

- `apply_adoption_plan` routes adopted pages by interactivity: `if len(page.Features) > 0` → `needs_tool_recreation` / **`tool-recreation-handler`** (priority boosted, "tools are the site's value"); else → `needs_content_page` / `page-build-handler`.
- **`tool-recreation-handler`** is a real, registered agent ("Recreates interactive tools"). Its workflow: `recreate_tool` (`execute_llm_prompt`, emits `<!-- tool-recreation-complete -->` into `tool_recreation`) → `check_tool_completeness` (balanced `<script>`/`<style>`, marker present, length) → `spawn_rerender` → `page-rerender` deploy.

So the M2 outcome (no widget) is real, but the **cause is a misroute, not a missing owner**: gamesdesign's tool pages were routed `needs_content_page` (query E) because `len(page.Features) == 0`.

**Why `Features` was empty — prime suspect (structural, ties to 029):** `buildPageFeatureMap` keys the feature map by the **raw** `fm["page"]` from the adoption LLM's `interactive_features[]`; the routing loop looks up `pageFeatures[pageName]` where `pageName` is **canonicalised** by `CanonicalisePage` — which for a tool *adds a `tool-` prefix*. So the lookup misses for every tool page even when the LLM detected the tool. Two candidate reasons for empty `Features`, distinguished by the §7 query G:
- **(b1)** LLM emitted `interactive_features` but the un-canonicalised key never matches → miss. Definite bug regardless.
- **(b2)** Adoption analysis emitted no `interactive_features` → detection-prompt problem; key fix alone won't help.

### 2.7 RESOLVED (2026-05-26, step 5): b1 confirmed — prefix desync

Query G (`rr.findings`, not `rr.results`) returned `interactive_features[].page` keys for all six tools and five games. Tool keys are **bare** (`drop-rate-simulator`, `ttk-calculator`, …); game keys carry the **`game-` prefix** (`game-p2p-networking`, …).

Traced through the uploaded `CanonicalisePage`:
- `tool` branch returns `"tool-" + bare` → lookup key `tool-drop-rate-simulator`; feature map stored `drop-rate-simulator` → **miss** → empty `Features` → static route. (Confirms query C's URL `/tools/drop-rate-simulator/index.html`, whose bare slug = `drop-rate-simulator`.)
- `game` branch returns `"game-" + bare`; feature key already `game-…` → lookup **matches**.

So the desync misroutes **exactly the roles whose canonical prefix differs from the raw feature key** — tools gain `tool-` (miss); games already had `game-` (match). Corroborated by query E: it showed the six tools as `needs_content_page` and **no game pages**, consistent with games having taken the `tool-recreation-handler` route that E's filter excluded.

**Fix locus:** self-contained in `buildPageFeatureMap` (it already receives the whole `plan`, so it has `plan["pages"]` with `name`+`page_type`). Resolve each feature's role from `plan["pages"]` and key the map by the **canonical** name via `CanonicalisePage`, so it lands in the same space the loop looks up. No signature change; no change to `tool-recreation-handler`. Reuses the existing helper. Pending §7 query (1) for exact `raw_name`/`raw_type` to implement the key transform precisely, and query (2) to confirm games routed to `tool-recreation-handler`.

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

## 5b. Tasks

Concrete, ordered. T1 and T2 drafted as code (see outputs); pull current revisions before applying (adoption code is moving in parallel).

- [x] **T1 — Fix the adoption interactivity misroute (M2 fix). CODE DRAFTED (step 7).** Merged `apply_adoption_plan_action.go` produced: `buildPageFeatureMap` (now ~lines 718–793) keys by **canonical** name via `datahelpers.CanonicalisePage`, resolving each feature's role from `plan["pages"]`. Verified lines 1–717 and trailing helpers byte-identical to the uploaded source; only the one function changed; one noted local rename (`pageName`→`rawPage`). `adoptedPage.Features` (now line ~551) and `contentData` (line ~525) both read `pageFeatures[pageName]` (canonical), so the single change fixes routing + attachment. `strings`/`datahelpers` already imported. No Go toolchain in workspace — **compile/`gofmt`/vet on a Go machine before merge.**
- [x] **T1b — Backfill resolved as "run T2" (no separate emitter).** A second emitter would duplicate T2 and split responsibility, so backfill = T2's next discovery run picks up the six widget-less tool pages and emits `needs_tool_recreation` with canonical feature matching. Immediate path: trigger discovery for gamesdesign. A pure-SQL backfill is rejected — it would re-implement canonical feature-key matching in SQL (the exact fragility being removed). Read-only preview query (T2's target set) recorded in §7 (query J).
- [x] **T2 — Post-adoption detection check. CODE DRAFTED (step 7).** `check_tool_recreation_needed.go` (package `discovery_checks`): detects `page_type IN ('tool','game')`, `status='active'` pages with no widget (no tool/game component **or** no `<script>` section — robust to the recreated widget's `component_level`), sources `interactive_features` from adoption findings by canonical name (reuses `CanonicalisePage`), and emits `needs_tool_recreation` (item_key `needs_tool_recreation:<name>`, distinct from adoption's `needs_page:<name>`). Pages with no captured features are reported but deferred to generation (tool-suggester), not auto-recreated. 7-day per-page cooldown; runner handles dedup. **Open:** confirm the `component_level` recreated widgets use (T4 query answers this); confirm `tool-recreation-handler` is `active`.
- [ ] **T4 — M1 / recreation-loss is ACTIVE and BLOCKING (step 8).** Query K: the five correctly-routed games have no widget. So recreation→deploy itself doesn't land a widget — T1 alone won't fix the tools. Diagnose via §7 queries L (page dump), M (`tool-game-*` origin), N (recreation target + clobber history) before any trigger. Then fix the actual loss (mis-target / clobber / handler persist), which is prerequisite to a useful backfill.
- [ ] **T5 — `tool-game-*` duplicate pages (step 8).** Five `page_type=tool`, `planned` pages duplicating the games. Identify the creator (query M) and remove/redirect; likely the 029 role-divergence (game re-canonicalised as tool) or recreation mis-targeting. Pairs with T3.
- [ ] **T3 — Canonicalise tool page identity (overlap fix).** Route `create_tool_component` **and** `deploy_tool` page name/url/page_type through `datahelpers.CanonicalisePage` (the `029` Phase-0 helper), so the three surfaces (adoption, planner, tool actions) stop diverging. `create_tool_component` is the older surface: it currently builds `/tools/<function>.html` with the `tool-` prefix embedded; it should pass the bare slug + `Role="tool"` and take the canonical `/tools/<slug>/index.html`. *Note for `029`: Phase-0 deliverables listed only `apply_adoption_plan_action.go` and `site_db_actions.go` — the tool actions were missed and should be added there too.*
- [ ] **T4 — M1 clobber fix (sequenced after T1).** Once widgets are generated, make them survive rebuilds: the widget must be a first-class section in `site_plan` (which `load_page_sections_from_spec` syncs into `pages.sections`), not a side-written `page_components` row. See §5 Options 1–2. Optionally add Option 3 as a guard.

T1 unblocks the visible symptom. T3 is independent and can land alongside. T4 must precede any rollout of T1 at scale or the generated widgets will be wiped on the next build.

---

## 6. Open questions

- [x] **Read `load_page_sections_from_spec` — DONE (2026-05-26).** `site_plan` is authoritative *and* its sections are synced back into `pages.sections` on every build (§2.4). Fix must target `site_plan` (planner/reconciler) or exempt tool pages from the sync; setting `pages.sections` alone is futile.
- [x] **GATING: clobber (M1) vs never-generated (M2) — RESOLVED 2026-05-26: M2.** Query D = 0 tool components for the site; E = only `page-build-handler` "Recreate …" items (no `tool-*` handlers, no `add_tool`); F = no history snapshot. The widget was never generated; nothing was clobbered. M1 remains latent and applies *after* M2 is fixed.
- [ ] Inspect `site_specs.site_plan` (`is_current = true`) for gamesdesign: is there an entry for `tool-drop-rate-simulator` (or `drop-rate-simulator`), and with what `sections`? (Lower priority now M2 is confirmed; still useful for T1/T4 to know what the planner emits for tool pages.) Query A.
- [x] **Did the tool pipeline run? — RESOLVED 2026-05-26: no.** Query E shows only adoption recreate items. Folded into the gating finding.
- [x] **Duplicate `pages` rows per tool slug? — RESOLVED 2026-05-26: no.** Query C returned a single row for `tool-drop-rate-simulator` (`sections=[]`, `build_status=needs_rebuild`). No `tool-X` vs `X` divergence for this slug currently.
- [ ] Confirm a known **fork-path** tool elsewhere (e.g. gaswholesalers "Gas Unit Converter") has its widget present in deployed HTML — does the fork path survive today? (Validates M1 severity for T4.)
- [ ] Confirm `RenderComponentAction` emits a `render_mode='standalone'` component's `html_template` into `sections_metadata` so the widget re-inserts on save when it's a listed section. (Needed for T4.)

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

-- G) RESOLVED b1 (2026-05-26): adoption emitted interactive_features. Column is `findings` (not `results`).
--    Tool keys are bare ('drop-rate-simulator'); game keys carry 'game-' prefix.
SELECT jsonb_path_query(rr.findings, '$.interactive_features[*].page') AS feature_page_key
FROM research_results rr
JOIN sites s ON s.id = rr.site_id
WHERE s.domain = 'gamesdesign.co.uk'
  AND rr.research_agent_type ILIKE '%adopt%'
ORDER BY rr.created_at DESC;

-- H) Exact raw page names/types adoption used (to implement the canonical key transform precisely).
SELECT p->>'name' AS raw_name, p->>'page_type' AS raw_type
FROM research_results rr
JOIN sites s ON s.id = rr.site_id
CROSS JOIN LATERAL jsonb_array_elements(rr.findings->'pages') AS p
WHERE s.domain = 'gamesdesign.co.uk' AND rr.research_agent_type ILIKE '%adopt%'
ORDER BY rr.created_at DESC;

-- I) Mechanism check: did games route to tool-recreation-handler (prefix matched), and get widgets?
SELECT wi.item_type, wi.handler_agent, wi.status, LEFT(wi.summary, 60) AS summary
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE s.domain = 'gamesdesign.co.uk'
  AND (wi.item_type = 'needs_tool_recreation' OR wi.handler_agent = 'tool-recreation-handler')
ORDER BY wi.created_at DESC;

-- J) T1b backfill preview: which interactive pages T2 will target (widget-less). Expect the 6 tools, NOT games.
SELECT p.name, p.page_type, p.build_status
FROM pages p
JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'gamesdesign.co.uk'
  AND p.page_type IN ('tool','game')
  AND p.status = 'active'
  AND NOT EXISTS (
    SELECT 1 FROM page_components pc
    LEFT JOIN content_components cc ON cc.id = pc.component_id
    WHERE pc.page_id = p.id
      AND ((cc.component_level IN ('tool','game') AND cc.is_active = true)
           OR pc.rendered_html LIKE '%<script%')
  )
ORDER BY p.page_type, p.name;

-- K) T4: did the games' widgets survive (and what component_level do they use)?
SELECT p.name, p.build_status,
       EXISTS (SELECT 1 FROM page_components pc
               JOIN content_components cc ON cc.id = pc.component_id
               WHERE pc.page_id = p.id
                 AND cc.component_level IN ('tool','game') AND cc.is_active = true) AS has_widget_component,
       EXISTS (SELECT 1 FROM page_components pc
               WHERE pc.page_id = p.id AND pc.rendered_html LIKE '%<script%')      AS has_script_section
FROM pages p
JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'gamesdesign.co.uk' AND p.page_type = 'game' AND p.status = 'active'
ORDER BY p.name;

-- L) DECISIVE (step 8): what is actually on a game page vs its tool-game twin.
SELECT p.name AS page, p.build_status, pc.position, pc.slot_name,
       cc.component_level, cc.function,
       length(pc.rendered_html) AS html_len,
       (pc.rendered_html LIKE '%<script%')                  AS inline_script,
       (pc.rendered_html ~* '<(canvas|form|input|button)')  AS interactive_markup,
       LEFT(pc.rendered_html, 70) AS html_start
FROM pages p
JOIN sites s ON s.id = p.site_id
LEFT JOIN page_components pc ON pc.page_id = p.id
LEFT JOIN content_components cc ON cc.id = pc.component_id
WHERE s.domain = 'gamesdesign.co.uk'
  AND p.name IN ('game-jelly-invaders', 'tool-game-jelly-invaders')
ORDER BY p.name, pc.position;

-- M) Who created the tool-game-* duplicates?
SELECT p.name AS page, p.url, p.build_status,
       wi.item_type, wi.handler_agent, wi.source, wi.created_by, wi.status,
       LEFT(wi.summary, 55) AS summary
FROM pages p
JOIN sites s ON s.id = p.site_id
LEFT JOIN site_work_items wi ON wi.page_id = p.id
WHERE s.domain = 'gamesdesign.co.uk' AND p.name LIKE 'tool-game-%'
ORDER BY p.name, wi.created_at;

-- N1) What did the needs_tool_recreation items target, and how did they end?
SELECT COALESCE(p.name, '(none)') AS target_page, wi.status, wi.handler_agent,
       wi.created_at, wi.updated_at
FROM site_work_items wi
LEFT JOIN pages p ON p.id = wi.page_id
JOIN sites s ON s.id = wi.site_id
WHERE s.domain = 'gamesdesign.co.uk' AND wi.item_type = 'needs_tool_recreation'
ORDER BY wi.created_at DESC;

-- N2) Clobber evidence: was a widget ever on a game page, then overwritten?
SELECT pch.created_at, pch.source, LEFT(pch.content_data::text, 80) AS snapshot
FROM page_component_history pch
JOIN pages p ON p.id = pch.page_id
JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'gamesdesign.co.uk' AND p.name = 'game-jelly-invaders'
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

- **2026-05-26 (step 1)** — Investigation. Confirmed clobber via DB (`pages.sections=[]`, widget absent) and code (`DELETE FROM page_components`). Identified novel-vs-fork divergence and `site_plan` authority. Drafted fix options. Folded in the "does the page even want these sections" design question.
- **2026-05-26 (step 2)** — Read `load_page_sections_from_spec`: `site_plan` is authoritative and **overwrites `pages.sections`** on every build (§2.4) → fix must target the planner/reconciler, not the action. Read `029`: pages are **adoption-shaped**; leaked `/services.html`/`/contact.html` CTAs match `B-029-2`. Raised **M2 (never generated)** alongside **M1 (clobber)** (§2.5).
- **2026-05-26 (step 3)** — Ran D/E/F. **Gating resolved: M2, not M1.** No tool component (D=0), tool pipeline never ran (E = adoption recreate only), no clobber (F=0). §2.4 puzzle resolved: sections came from the adoption crawl via the recreate path. Added tasks T1–T4; wrote debugging-guide addendum.
- **2026-05-26 (step 4)** — Verified the "no agent owns widget generation" claim before acting; **it was false** (§2.6). `apply_adoption_plan` routes interactive pages to **`tool-recreation-handler`** (real agent: LLM-recreates widget, `check_tool_completeness`, `page-rerender`). Cause is a **misroute** — gamesdesign tools had empty `Features` → static path. Reframed T1 from "build a handoff" to "fix the misroute"; added query G (b1 vs b2). Noted `check_missing_tools` is a separate site-level trigger, not the cause.
- **2026-05-26 (step 5)** — Ran query G (column `findings`). **b1 confirmed** (§2.7): adoption emitted `interactive_features` for six tools + five games. Tool keys bare (`drop-rate-simulator`), game keys prefixed (`game-…`). Via uploaded `CanonicalisePage`: tool branch adds `tool-` → lookup `tool-drop-rate-simulator` ≠ stored `drop-rate-simulator` → miss → static route; games match (key already prefixed). Mechanism = prefix desync affecting only roles whose canonical prefix differs from the raw key. Fix locus pinned: `buildPageFeatureMap` should key by canonical name via `CanonicalisePage`. Next: §7 queries H (raw names) + I (games-route check), then implement T1.
- **2026-05-26 (step 6)** — Ran H + I. **Fully confirmed.** H: tools bare/`tool`, games prefixed/`game`. I: all five games routed `needs_tool_recreation → tool-recreation-handler` (complete); zero tools. Read current `apply_adoption_plan_action.go`: `buildPageFeatureMap` (721–742) keys raw; `adoptedPage.Features` (551) and `contentData` (525) both read `pageFeatures[pageName]` (canonical) — so one change fixes both.
- **2026-05-26 (step 7)** — Produced merged `apply_adoption_plan_action.go` (canonical-keyed `buildPageFeatureMap`; verified only that function changed) and new discovery check `check_tool_recreation_needed.go` (T2). Resolved T1b as "run T2" — no separate emitter (avoids duplicate ownership); rejected a SQL backfill (would re-implement canonical matching in SQL). Added §7 query J (T2 target preview) and K (T4 game-widget survival). Open before merge: compile/`gofmt`/`vet` on a Go machine; confirm `tool-recreation-handler` active.
- **2026-05-26 (step 8)** — Ran J + K. **Blocking finding:** all five games are widget-less (K: has_widget_component=f, has_script=f) despite correct routing + completed recreation (step 6) → recreation→deploy itself doesn't land a widget; T1 alone insufficient. J also surfaced five new `tool-game-*` planned tool pages (duplicates of the games). State moving (drop-rate now `deployed`). Held the trigger; recreation must be diagnosed first. Added §7 queries L (page dump, decisive), M (tool-game-* origin), N1/N2 (recreation target + clobber history); tasks T4 (now active/blocking) and T5 (tool-game-* duplicates). Caveat logged: K may be a snippets false-negative (query L settles it).
