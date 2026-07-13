# Running Notes — vonc.com build session

## Session started: 2026-06-22

---

## Root cause chain (in discovery order)

### 1. `write_site_spec` — spec_data string rejection
- **Found:** 2026-06-22 17:02–17:03
- **Symptom:** `persist_mission_brief` / `persist_roadmap_brief` failed:
  "spec_data must be a JSON object, got string"
- **Cause:** `WriteSiteSpecAction` hard type-asserted `spec_data` to
  `map[string]interface{}` with no string coercion path. The domain-submitter
  workflow resolves `input_data.mission_brief` which is a plain string.
- **Fix:** Coercion block in `WriteSiteSpecAction`: JSON string → parse;
  plain string → wrap as `{"text": value}`; object passes through.
- **File:** `platform/orchestration/actions/site_spec_actions.go`
- **Status:** Code delivered. Needs deployment.

### 2. `gauntlet-interface` template — `<no value>FIELD</no>` artifacts
- **Found:** 2026-06-23
- **Symptom:** Deployed HTML showed `eyebrow_label</no>`, `challenge_title</no>` etc.
- **Cause:** Template stored as rendered output with field name fallbacks preserved.
  Pattern: `<no value>FIELDNAME</no>` throughout `html_template`.
- **DB fix:** `regexp_replace` to rewrite `<no value>FIELD</no>` → `{{.FIELD}}`,
  then second pass to strip backslash artifact `{{\. → {{.`
- **Component ID:** `5da50747-7936-4b8f-a66d-c1ea98919c75`
- **Status:** Fixed in DB. Rerender completed 13:13.

### 3. `archetype-result-card` template — bare `<no value>` (no field names)
- **Found:** 2026-06-23
- **Symptom:** 30 `<no value>` artifacts, 0 `</no>` closing tags. Quality 30.
  "0 template variables". Template stored as fully-cleaned render output.
- **Cause:** Different failure mode — template rendered against empty context,
  `RenderTemplate` cleaned `<no value>` to empty string, that blank output
  stored as source. Field names irretrievably lost.
- **Fix:** `needs_component_regeneration` work item raised → `component-creator`
  regenerated from intact `input_schema`. Quality now 100, 28 template variables.
- **Component ID:** `2c7678fb-9940-428d-8b78-62e2510f6dbe`
- **Status:** Regenerated. `build_status = pending` on page_components rows —
  needs migration 003 to unblock rerender.

### 4. `render_mode = 'template'` hardcoded on all components (SYSTEMIC)
- **Found:** 2026-06-24
- **Symptom:** All section components have `render_mode = 'template'`,
  `template_variable_count = 0`. `check_render_mode` in page-content-writer
  always routes to `render_from_template`. LLM content generation path
  permanently unreachable.
- **Cause:** `StoreGeneratedComponentAction` INSERT hardcodes `'template'`
  as `render_mode` regardless of whether the schema has LLM fields.
  UPDATE path doesn't set `render_mode` at all.
- **Scope:** Library-wide. Affects every component ever created.
  Components with `actual_template_slots > 0` AND `llm_field_count > 0`
  need `render_mode = 'agent'` to receive LLM content.
- **Code fix:** `deriveRenderMode(inputSchemaJSON)` helper added.
  Called in both INSERT and UPDATE paths.
- **File:** `platform/orchestration/actions/store_generated_component_action.go`
- **DB migration:** Migration 002 — fix `check_render_mode` condition in `page-content-writer`
  agent_definition. NOT a component table change.
- **Status:** Code fix delivered (`deriveRenderMode` in `StoreGeneratedComponentAction`
  for future components). Agent_definition fix pending (migration 002).
- **NOTE:** `render_mode` on existing components does NOT need updating — the workflow
  routing reads `current_section.llm_field_specs` (set by plan_sections from schema),
  not `render_mode`. Migration 002 (render_mode sweep across 65 components) is DROPPED.

### 5. Broken components needing regeneration
Components with `actual_template_slots = 0` AND `llm_field_count > 0`
have empty templates — LLM schema fields but no `{{.field}}` slots.
These cannot render content regardless of `render_mode`.

| function | llm_fields | template_slots | action needed |
|---|---|---|---|
| `gauntlet-cta` | 16 | 0 | regenerate |
| `system-stats` | 24 | 0 | regenerate |
| `hero` | 4 | 6 | slots present — test after render_mode fix |
| `brief-explanation` | 0 | 0 | static — OK |
| `lobby-grid` | 0 | 0 | static — OK |
| `provocation-card` | 0 | 0 | static — OK |
| `gauntlet-interface` | 11 | 33 | slots present — test after render_mode fix |
| `archetype-result-card` | 17 | 29 | regenerated — test after migration 003 |
| `tool-archetype-taster-quiz` | 8 | 22 | slots present — test after render_mode fix |

### 6. Two manual rerender spec shape bug
- **Found:** 2026-06-23
- **Cause:** Manual `page_rerender` work items we inserted used
  `{"page_name": "..."}` only. `rerender_single_page` requires `page_id`
  (UUID), `domain`, `filename`, and `page_name`.
- **Fix:** Correct spec shape documented. See runbook.

### 7. Discovery checker gap
- **Found:** 2026-06-23
- **Fix:** `checkBrokenTemplateSlots` sub-check added to
  `check_component_standards.go`. `repair_template_slots` fix type added
  to `fix_component_template_action.go` with detection of Mode A
  (repairable) vs Mode B (needs regeneration).

---

## Two broken-template failure modes

**Mode A** (`<no value>FIELD</no>`): Field names survive as fallback text.
Repairable by `repair_template_slots` string substitution.

**Mode B** (bare `<no value>`): Field names lost. Template rendered against
empty context, `RenderTemplate` cleanup stripped everything.
Requires `needs_component_regeneration` → `component-creator`.

`repair_template_slots` now detects Mode B (no `</no>` tags) and returns
`action: needs_regeneration` rather than attempting a futile repair.

---

## Submission script issue
`004_submit_vonc_trigger.sh` builds `MESSAGE_BODY` via python3 but never
sends it — the `kcat` call mid-script uses a hardcoded inline `<<JSON` body.
Both payloads are identical for this run so no data was lost. Fix before
resubmitting with changes.

---

## Migration status (2026-06-24)

| Migration | Status | Notes |
|---|---|---|
| 001 | N/A | Handled via work items (gauntlet-cta, system-stats regen pending) |
| 002 | DONE | `check_render_mode` condition fixed in page-content-writer agent_definition |
| 003 | Pending | Unblock `pending` page_components |
| 004 | Pending | Rerenders queued (UUIDs needed fixing — see below) |

### Migration 002 outcome
`check_render_mode` condition changed from:
`current_section.component.render_mode == 'agent' OR current_section.component.needs_llm == true`
to:
`current_section.llm_field_specs != null`

This routes any section with LLM fields in its schema to content generation.
Takes effect immediately for all new page builds across all sites.

### Work items queued (2026-06-24) — UUIDs need fixing before claim
Three work items were inserted with placeholder `page_id` values.
Fix with UPDATE before workers claim them (see runbook fix queries).

| item | page_id to set |
|---|---|
| page_rerender tool-gauntlet | ecb637c1-845f-46bf-b174-9c92a43f9586 |
| page_rerender tool-archetype-taster-quiz | f1bc679f-5c48-46e8-9bb5-76cb8cf99ca5 |
| needs_page index | b4d24f8e-fccd-49df-9dad-aa56a0b20a68 |

### Components still needing regeneration
`gauntlet-cta` and `system-stats` have 0 template slots — need `needs_component_regeneration`
work items after code deployment. IDs:
- gauntlet-cta: `66bfd4a4-2163-4d34-be43-c42ee17e6af0`
- system-stats: `fdd92ad4-521a-4602-89cf-7ee1a66c10f1`

---

## Work item fix — 2026-06-24 (~11:46)

### Placeholder UUID problem

The three work items inserted for tool-gauntlet rerender, tool-archetype-taster-quiz
rerender, and index rebuild were inserted with literal placeholder strings as `page_id`
(`<tool-gauntlet id>` etc.) rather than real UUIDs.

The subsequent fix UPDATEs used the intended UUID value in the WHERE clause
(`spec->>'page_id' = 'ecb637c1-...'`) rather than the placeholder string
(`spec->>'page_id' = '<tool-gauntlet id>'`). This matched rows that already had
correct UUIDs and updated them (no-op), while the placeholder rows survived unchanged.
Result: `UPDATE 3` affected the wrong rows; placeholder rows remained broken.

Both tool page rerenders failed with "invalid UUID length: 18" and "invalid UUID length: 31"
(lengths of the placeholder strings). They reset to `triaged` and would retry.

**Resolution:** Delete and reinsert the two broken rerender items with correct UUIDs.
The index `needs_page` item had already been claimed correctly (UUID was correct there).

**Lesson:** When fixing placeholder values in jsonb specs, always filter on the
placeholder string itself, not the intended replacement value.

### Correct fix query pattern for placeholder UUIDs
```sql
-- Filter on the placeholder, not the target value
DELETE FROM site_work_items
WHERE spec->>'page_id' IN ('<tool-gauntlet id>', '<tool-archetype-taster-quiz id>');

-- Then reinsert with real UUIDs inline — no placeholder strings
```

### Current queue state (11:46)
- `needs_page index`: claimed at 11:46, in flight, no error
- `page_rerender tool-gauntlet`: triaged with error (placeholder), awaiting delete+reinsert
- `page_rerender tool-archetype-taster-quiz`: triaged with error (placeholder), awaiting delete+reinsert

---

## 2026-06-24 ~13:45 — Index rebuild result

Index page rerendered after `needs_page` rebuild completed (~11:46 claimed,
complete ~2 hours later).

**Hero section:** LLM routing fix working. Content generated correctly:
- Headline: "The world happened today. Your take is already late."
- Subheadline: Arena vocabulary, Gauntlet references, Archetype hook
- CTA: "Enter the Gauntlet" → `/contact.html` (unresolved CTA, expected)

**provocation-card, lobby-grid, brief-explanation:** Still empty shells.
These components have `llm_field_count = 0` AND `actual_template_slots = 0`.
The routing fix correctly sent them to `render_from_template`, but the
templates have no slots — nothing to fill from any source.

**Hypothesis:** These are intentionally runtime/JS-populated components.
The roadmap spec describes provocation-card as showing a live daily provocation
(AI-generated daily content, not build-time LLM). If so, they are correct as
static shells — JS populates them at page load. Needs confirmation from schema.

**Next:** Query schema for provocation-card, lobby-grid, brief-explanation
to determine whether they are intentionally static or need regeneration with
proper slots.

---

## 2026-06-24 ~14:00 — provocation-card, lobby-grid, brief-explanation confirmed static

Schema query returned `{}` (empty) for all three, zero template variables, zero slots.

These are **intentionally static shells** — no LLM fields were ever defined.
The provocation-card content is populated at runtime by the daily provocation
generation pipeline (JS from `/assets/js/snippets.js`). Not a build-time concern.

The index page is in the correct deployed state for V1:
- Hero: LLM content correct ✓
- provocation-card / lobby-grid / brief-explanation: static shells, JS-populated ✓
- gauntlet-cta / system-stats: still need regeneration (have schemas, zero template slots)

**Remaining component regeneration needed:**
- `gauntlet-cta` id: `66bfd4a4-2163-4d34-be43-c42ee17e6af0` — 16 LLM fields, 0 slots
- `system-stats` id: `fdd92ad4-521a-4602-89cf-7ee1a66c10f1` — 24 LLM fields, 0 slots

These need `needs_component_regeneration` work items raised after code deployment.
Until then the index gauntlet-cta and system-stats sections will render as empty shells.

**Separate pipeline concern (out of scope this session):**
Daily provocation generation pipeline to feed provocation-card, lobby-grid at runtime.

---

## 2026-06-24 ~15:30 — Index visual inspection

Screenshot shows two problems, both caused by missing CSS variable injection:

**Hero CTA button:** Dark blue/teal (`#0f3460` fallback) instead of magenta-pink (`#fc5c7d`).
`var(--accent-color)` not defined — theme CSS not injected.

**provocation-card section:** Bright solid violet instead of near-black arena feel.
`var(--color-primary, #1a1a2e)` resolving to bright violet, not `#0a0a0f`.
`var(--color-background)` not set.

**Root cause:** The `/* Theme-specific styles injected here: */` block in the `<head>`
is empty. CSS custom properties from the design spec (palette: `#7c3cff` primary,
`#fc5c7d` accent, `#0a0a0f` background) are not being injected at render time.

**Next step:** Check `resolved_composition` spec for `css_theme_id`, then check
whether `styles.css` has the correct variables or whether theme injection is failing.

---

## 2026-06-24 ~16:00 — CSS theme flow clarified

**webdesign-agent is not deprecated.** Per doc 027:

The pipeline is:
```
needs_composition (priority 7) → site-design-planner
    [depends_on gate]
needs_design (priority 8) → webdesign-agent
    → analyze_design (LLM → design_spec)
    → update_site (persists design_spec)
    → generate_css (render_css_from_spec — reads composition via FKs)
    → deploy_css (writes assets/css/styles.css to git)
    → [optionally fork_theme]
```

`css_themes.css_content` is **intentionally empty** — post-025 renderer reads
composition via FK chain at render time, not from stored css_content.
This is by design per `install_site_composition_action.go` line 210-212.

**styles.css was deployed 2 days ago** by webdesign-agent (commit msg:
"Update stylesheet via webdesign-agent").

**Remaining question:** What is in styles.css? Either:
(a) The file has wrong hex values (render_css_from_spec used wrong data)
(b) The `:root {}` block is correct but not loading (file 404 or specificity issue)

The page uses `var(--color-primary, #1a1a2e)` — the fallback is firing, not the
CSS variable. This means the variable either isn't defined or isn't reaching
the element.

**Next:** Check styles.css content via git fetch, and check
`sites.style_collection_id` is set correctly.

---

## 2026-06-24 ~16:30 — CSS root cause found: variable name mismatch

**styles.css is correct.** Has right values: `--color-primary: #7c3cff`,
`--color-accent: #fc5c7d`, `--color-background: #08080e`.
`needs_design` result confirmed correct design_spec.
`style_collection_id` and `css_theme_id` links confirmed correct.

**The problem is in the rendered component HTML, not the CSS.**

Hero CTA button uses `var(--accent-color, #0f3460)` in its inline styles.
The layout and styles.css define `--color-accent`, NOT `--accent-color`.
So the button falls back to `#0f3460` (dark blue) — wrong variable name.

Bright violet provocation-card section is actually correct — `--color-primary: #7c3cff`
is resolving properly. It looks jarring because of section ordering but the
colour value is right.

**Root causes:**
1. Hero component template generates `--accent-color` instead of `--color-accent`.
   Fix: update the hero component's html_template or regenerate it.
2. The jarring violet section appearance is a design/section issue, not a CSS bug.

**Next:** Check hero component template for `--accent-color` vs `--color-accent`.

---

## 2026-06-24 ~16:35 — Hero CSS variable fix applied

Snapshot: `858ebc23-3ba1-448b-890a-b370231ca659`

`UPDATE 1` — hero template fixed: `--primary-color` → `--color-primary`,
`--secondary-color` → `--color-secondary`.
Verify query returned 0 rows — clean.

Index rerender queued (`manual-rerender-index-hero-var-fix-*`).

Library-wide scan: only `archetype-grid` has a remaining non-standard-looking
variable (`--archetype-color`). Needs verification — likely intentional
per-archetype tinting, not a mis-named system variable.

After index rerender: hero gradient should show violet → deep violet → magenta-pink.
CTA button was already using `--color-accent` correctly; now the fallback hex
`#0f3460` matters less since the variable will resolve.

---

## 2026-06-24 ~16:40 — archetype-grid --archetype-color confirmed intentional

Context: `var(--archetype-color, var(--color-accent, #a78bfa))`

This is a per-archetype tinting variable set dynamically per card, with
`--color-accent` as a fallback and `#a78bfa` as a final fallback. Not a
mis-named system variable. No fix needed.

Library-wide CSS variable audit complete — hero was the only component
with mis-named system variables. No other issues found.

---

## 2026-06-24 ~16:50 — component-creator prompt patched: CSS variable naming

Snapshot: `snapshot_agent('component-creator', 'pre-css-variable-naming-fix')`
source_id: `23720180-7a39-4e3d-92e1-ebdbf95b57f4`

Section 7 of the `component-creator` prompt_template updated:
- Heading changed to "USE ONLY THESE NAMES"
- Added indented structure separating Palette from Layout tokens
- Added STRICT RULE paragraph explicitly prohibiting `--primary-color`,
  `--secondary-color`, `--accent-color`, `--background-color`, `--text-color`,
  `--border-color` — naming them as wrong and explaining why

`UPDATE 1` confirmed. Verified new heading present in DB.

All components generated from this point will use correct `--color-*` names.
Existing library components with wrong names: only `hero` (now fixed).

---

## 2026-06-25 ~16:30 — Blank cards root cause: empty-schema static-shell components

Full 8-page rebuild queued (`INSERT 0 8`). Hero now correct (magenta CTA, dark bg,
on-brand copy). But provocation-card, lobby-grid, brief-explanation render as
blank card outlines — empty input fields, unlabelled buttons, empty stat slots.

**Root cause confirmed:**
- `provocation-card`, `lobby-grid`, `brief-explanation` all have `input_schema = {}`
  (empty), zero template variables.
- NO js_snippets apply to them (query returned 0 rows).
- So there is NO content mechanism: not LLM-filled (no schema fields),
  not JS-filled (no snippets). They render hardcoded structure with no data path.

These components have card grids, input fields, and stat strips baked into the
HTML with no `{{.field}}` slots and no runtime JS. They were created as static
shells but nothing populates them.

**The about page shows the same pattern** — the `01 02 03` numbered cards,
comparison rows, and empty boxes are also empty-schema shell components.

**This is a structural issue, not a rebuild failure.** A rebuild cannot fill
components that have no fields and no JS. Two options:
1. Regenerate these components WITH proper input_schema (LLM fields) so the
   content writer fills them — makes them build-time content.
2. Build the runtime JS pipeline (daily provocation feed) + add js_snippets
   that populate them client-side — makes them runtime content.

Option 1 is simpler and fits the existing pipeline. Option 2 matches the
original "daily provocation" product vision but needs the data API built.

**Page structure note:** No `page_plan` aspect exists. site_specs aspects are:
briefing, classification, content_direction, design_intent, identity, mission,
resolved_composition, roadmap, strategy, submission. The per-page section
breakdown is stored elsewhere (pages table or within roadmap/strategy).

**Decision needed before fixing:** which option for the shell components.

---

## 2026-06-25 ~17:15 — Page plan located + Option 2 confirmed as intended design

**Page plan storage:** NOT in site_specs. It lives in:
`site_plans` (one per site, is_current) → `site_plan_sections` (plan_id, page_name, ordering, component_name).
vonc.com plan_id: `77493277-f510-47ea-aa27-8fca415743d6`.

**Full vonc.com section plan (22 sections across 8 pages):**
- index: hero, provocation-card, gauntlet-cta, brief-explanation, lobby-grid, system-stats
- about: hero-about, content-block-about, game-master-explanation, platform-comparison, differentiators, gauntlet-cta
- archetypes: hero, archetype-grid, archetype-combinations, call-to-action
- contact: hero-contact, contact-form
- tool-archetype-taster-quiz: tool-archetype-taster-quiz, archetype-result-card
- tool-gauntlet: gauntlet-interface, archetype-result-card

**Empty-shell components needing a content mechanism:** provocation-card,
lobby-grid, brief-explanation (all on index). These are the three with empty
schemas and no JS.

**Option 2 (runtime JS pipeline) confirmed as the ORIGINAL intended design.**
From the April "Community-driven AI chat platform concept" conversation, the
Spark v1 roadmap explicitly specifies:
- "Interactive elements (Gauntlet, AI sparring) use client-side JS calling backend APIs"
- "daily_provocation_generation_from_scraping"
- "daily static regeneration"
- "Everything is static HTML regenerated by the agent pipeline"

**The JS delivery mechanism already exists and works** (confirmed from
gaswholesalers + semantic-design conversations):
- `js_snippets` table, keyed by `applies_to` (component function names)
- `render_js_snippets_for_site` action → concatenates matching snippets → `/assets/js/snippets.js`
- `site-asset-renderer` agent renders + commits the bundle
- head component already has `<script src="/assets/js/snippets.js"></script>` (confirmed in vonc HTML)
- Matching is automatic at render time by `applies_to` overlap

**The news feed pipeline is a proven template for the data layer.** It does
exactly this pattern for news: scrape sources → score → produce
`/data/latest-news.json` → static site reads it. Components: content-feed-trigger,
content-feed-orchestrator, feed-ingester, feed-triage. Tables: content_sources,
content_feed_items. Scheduled: content-feed-refresh every 6 hours.

**The Spark provocation pipeline would mirror the news pipeline** with a
different LLM prompt (generate provocations + AI takes instead of summarising
news) producing `/data/provocations.json`.

**Pending check (queries built, not yet run):** whether ANY Spark content
pipeline already exists — content_sources, content_feed_items, scheduled_tasks,
agents, or js_snippets for provocation/spark/daily. Expectation: none exists.
Run the 5 diagnostic queries before building.

---

## 2026-06-25 ~18:00 — Phase 2 mechanism fully confirmed from source files

Read render_news_section_action.go, render_js_snippets_for_site_action.go, and
the agent_definitions for content-feed-orchestrator/trigger/ingester/triage +
site-asset-renderer. The full mechanism is now precisely understood.

**provocation-card shell current state:** re-rendered by the 8-page rebuild. The
template DOES have Go slots but content_data is empty, so every slot renders as
literal `<no value>`. The DOM structure is intact and matches our contract:
`.pc-eyebrow`, `.pc-headline#pc-headline` (with `<em>` support), `.pc-body`,
`.pc-btn-primary` (a[href]), `.pc-btn-secondary` (a[href]),
`.pc-stat-value`×3 + `.pc-stat-label`×3, `.pc-card`×4 each with
`.pc-card-icon`/`.pc-card-title`/`.pc-card-desc`. There is also an existing
inline `<script>` IIFE in the template that does hover/focus card activation —
our loader must coexist with it (different concern; ours fills content, theirs
handles interaction). NOTE the template's own script appears truncated in the
dump (ends mid-function with a stray backslash) — worth checking the live
rendered_html isn't broken, but this is the COMPONENT template not the snippet.

**site-asset-renderer (confirmed):** type `site-asset-renderer`, active.
Input: `site_id` required, `domain` optional. Workflow: ensure_site_record →
render_js_snippets_for_site → git_commit (commit msg "Update JS snippets bundle",
writes `assets/js/snippets.js`). timeout 120s. To trigger Phase 2 Step 4 we
just need to spawn this agent with site_id = vonc.

**render_js_snippets_for_site (confirmed):** loads the site's component function
list (from page_components + site_components), then selects js_snippets WHERE
is_active AND applies_to (jsonb array) overlaps that list, via
jsonb_array_elements_text. Concatenates into a bundle with header comments,
ordered by name. Empty bundle still written so the head <script src> never 404s.
So: our js_snippets row with applies_to ["provocation-card"] WILL be picked up
automatically because the index page uses the provocation-card component.

**git_commit pattern (confirmed):** takes a `files` map keyed by repo-relative
path (e.g. "data/latest-news.json", "assets/js/snippets.js") + domain_field +
commit_message. This is how ALL committed files reach the site repo for S3 deploy.
- Phase 2: commit hand-written `data/provocations.json` directly via a git_commit
  (or just place it in the repo).
- Phase 3: a new `render_provocations_section` action (mirror of
  render_news_section) produces the files map; git_commit writes it.

**News pipeline structure (the Phase 3 template), all confirmed active v1.0.1078:**
- content-feed-trigger: scheduled heartbeat (6h), finds news-recommended sites,
  spawns+calls content-feed-orchestrator per site (max 5). Uses scheduled_tasks
  row name='content-feed-refresh' (note: column is `name`, not `task_name` —
  that resolves the earlier scheduled_tasks query error).
- content-feed-orchestrator: seed_content_sources → dispatch_feed_sources
  (spawn feed-ingester per due source) → spawn+call feed-triage → 
  render_news_section → git_commit. 
- feed-ingester: route_by_type (rss/scrape/news_search/api_news) → normalize →
  write_feed_items → update_source_timestamps. REUSABLE AS-IS for scraping.
- feed-triage: load items → read_site_spec → LLM relevance+credibility scoring →
  apply_feed_scores. Our provocation-generator is the GENERATIVE analogue
  (turn topics into provocations, not just score them).

**render_news_section internals (the model for render_provocations_section):**
- Loads site domain, loads headline/subheadline from the page_component
  content_data (function='latest-news'), expires stale items, loads relevant
  feed items, builds JSON, returns files map {data/latest-news.json: ...}.
- Produces an archive JSON too if a news-index page exists (page_type='news-index').
  Spark parallel: produce a fuller provocations.json if provocations-index exists.
- KEY: the JSON shape is defined by a Go struct (newsJSONOutput/newsJSONItem).
  Our render action defines a provocationJSONOutput struct matching the DOM contract.

**Implication for build order:** Phase 2 is purely: (1) hand-write
data/provocations.json, (2) insert one js_snippets row (loader IIFE), (3) spawn
site-asset-renderer, (4) verify. No Go needed for Phase 2. Go (the render action +
provocation-generator agent) is Phase 3 only.

---

## 2026-06-25 ~18:30 — Reframing: fix in the framework, not manual pastes

User correction (important): the DATABASE is the source of truth, not the git
repo. So:
- The js_snippet belongs in the DB (DONE — provocation-card-loader inserted,
  is_active=t, applies_to ["provocation-card"], 4879 bytes). This is correct
  and systematic — render_js_snippets_for_site will bundle it.
- The DATA (provocations.json) should ALSO be produced from the DB and rendered
  out like any other asset — NOT hand-committed to the repo.
- Ideally BOTH the snippet AND the provocation-card component should have been
  generated dynamically when needed, rather than us patching after the fact.

Two distinct mechanisms, must not conflate:
1. JS SNIPPET DELIVERY: js_snippets row → render_js_snippets_for_site →
   assets/js/snippets.js (via site-asset-renderer + git_commit). This exists
   and works. Question: why did the provocation-card-loader snippet not already
   reach the site? Two sub-possibilities:
   (a) The snippet never existed until we just inserted it — i.e. nothing in the
       build pipeline creates loader snippets for JS-driven components. This is
       the likely structural gap.
   (b) The snippet existed but site-asset-renderer was never triggered for vonc
       after the component was added.
2. DATA DELIVERY: provocations.json. There is NO production mechanism yet — the
   news pipeline's render_news_section is the analogue but nothing equivalent
   exists for provocations. This is Phase 3 (the data pipeline).

KEY STRUCTURAL QUESTIONS to research (need data/code/docs):
- What in the build pipeline is responsible for ensuring a JS-driven component
  gets its loader snippet? (Probably nothing — that's the gap.)
- How does a component DECLARE that it needs a runtime data source + loader?
  (provocation-card has empty schema {} — it declares nothing. A proper design
  would have it declare its data dependency so the pipeline provisions both the
  snippet and the data feed.)
- When is site-asset-renderer triggered in the normal build flow? Is it wired
  into the build pipeline, or only invoked ad hoc / by webdesign-agent?
- How did the working tool components (gauntlet-interface etc.) get THEIR JS?
  They have inline <script> in the template (extracted to assets). The shells
  (provocation-card/lobby-grid) DON'T — they were built expecting an external
  loader that nothing produces.

This reframes Phase 2: rather than just proving JS-fills-shell manually, the
real fix is making the pipeline PRODUCE the loader + data automatically. Manual
steps become the throwaway proof; the framework fix is the deliverable.

---

## 2026-06-25 ~19:00 — Framework diagnosis COMPLETE: all three gaps confirmed

Read render_js_snippets_for_site_action.go (+test), render_news_section_action.go,
rerender_single_page_action.go, 022_dynamic_applications.md, js_snippets schema,
and the planner/asset agent definitions. The three gaps are now precisely pinned.

### js_snippets schema (confirmed)
Columns: id, name (varchar100 unique), description, js_content (NOT NULL),
semantic_tags jsonb, applies_to jsonb, dependencies jsonb, created_at, is_active
(NOT NULL default false). Our insert used name/description/applies_to/js_content/
is_active — all valid. (No site_id column — snippets are LIBRARY-WIDE, matched to
sites by applies_to ∩ site's component functions. So the SAME snippet auto-applies
to every site using provocation-card. Good.)

### THE TWO JS DELIVERY PATHS (this is the key architectural finding)
There are TWO separate JS delivery mechanisms, and the shells fell between them:

PATH 1 — per-component inline <script> → /tools/assets/{function}.js
- A component with inline <script> in its html_template has that script
  extracted to content_components.js_content (separateInlineJS).
- rerender_single_page_action.collectJSAssets reads cc.js_content for the
  page's components and emits files at `tools/assets/{function}.js`.
- This deploys AUTOMATICALLY on every page rerender — no separate trigger.
- This is how gauntlet-interface, tool-archetype-taster-quiz etc. get their JS:
  their interactivity is SELF-CONTAINED in the component.
- provocation-card ALSO has an inline <script> (the card hover/activation IIFE),
  so it DOES get a /tools/assets/provocation-card.js — but that script only does
  hover effects; it does NOT fetch data. The DATA-loading JS is separate.

PATH 2 — library js_snippets → /assets/js/snippets.js
- render_js_snippets_for_site bundles js_snippets WHERE applies_to ∩ site
  components, → /assets/js/snippets.js, committed by site-asset-renderer.
- This is the path our provocation-card-loader uses.
- CRITICAL: this path is NOT in the normal page build/rerender flow.

### GAP 3 CONFIRMED — site-asset-renderer is NOT wired into the build
Query: which agents reference site-asset-renderer / render_js_snippets?
RESULT: only `rerender-site` and `webdesign-agent` reference them. site-asset-
renderer's own description says "Triggered when js_snippets or component set
changes, or invoked by webdesign-agent."
- So /assets/js/snippets.js is rendered during INITIAL DESIGN (webdesign-agent)
  and on a full site rerender — but NOTHING re-runs it when a js_snippets row is
  added later. We added provocation-card-loader AFTER the site was built, so the
  bundle was never regenerated. THIS is why the JS didn't reach the site.
- The page rerenders we ran rebuild page HTML + /tools/assets/*.js (Path 1) but
  do NOT touch /assets/js/snippets.js (Path 2).
- No discovery check exists that detects "site has a js_snippet that applies but
  its snippets.js predates it" and spawns site-asset-renderer.

### GAP 2 CONFIRMED — nothing creates a loader snippet for a data-driven shell
- component-creator's contract (just re-read at v1.0.1080) has Tier A/B/C/D +
  renderer fields. A "renderer" source field is "filled at render time by JS or
  the renderer" with a fallback — but there is NO tier for "this component is
  populated from an external JSON feed at runtime, and needs a loader snippet +
  a data file produced by a pipeline." 
- So when a shell like provocation-card is created, NOTHING emits a companion
  js_snippet loader, and nothing registers that it needs a data feed. The shell
  is created with empty content and no delivery. Confirmed structural.
- How news does it: the 'latest-news' component pairs with (a) a js_snippet
  loader [Path 2] AND (b) render_news_section producing the JSON. Both were hand-
  built for news; there's no generic "data-driven component" abstraction that
  produces them automatically. (Need to confirm latest-news's snippet exists —
  run B3/B4 still.)

### GAP 1 CONFIRMED — no provocation data pipeline (as expected)
No content_sources / feed items / agents / scheduled task for Spark. The news
pipeline (render_news_section + content-feed-orchestrator + content-feed-refresh
scheduled task) is the template. scheduled_tasks keys on `name` (confirmed via
content-feed-trigger's UPDATE ... WHERE name='content-feed-refresh').

### assets table (confirmed) — NOT where snippets.js is tracked
The `assets` table is for IMAGES (asset_type, dimensions, origin_prompt, S3
storage_path etc.). snippets.js and tools/assets/*.js are committed straight to
git by git_commit; they are NOT rows in `assets`. So "has snippets.js ever been
produced for vonc" can't be answered from `assets` — check the git repo / the
site-asset-renderer run history instead.

### Doc 022 doctrine (confirms Option 2 is on-architecture)
Tier 1 (now): "Static Sites with Dynamic Components… Dynamic content injection
(RSS feeds, API-fetched data rendered client-side)… Still static HTML on
Cloudflare. The dynamic part runs in the browser." This is EXACTLY the
provocation pipeline. Principle 4: "Backend complexity lives in agents, not in
generated code." Principle 3: generated vs human content must be marked.

### WHERE THE FIX LANDS (informed by the above)
- Gap 3 (smallest, highest leverage, fixes the immediate bug for ALL sites):
  wire site-asset-renderer to run when snippets change. Two options:
  (a) a design-discovery-agent check (it already scans design-domain issues incl.
      "undeployed assets", "stale header/footer renders") — add a check
      "js_snippets bundle stale / missing for applicable snippet" → spawn
      site-asset-renderer. design-discovery-agent is the natural home.
  (b) make page build/rerender ALSO refresh snippets.js. Heavier; Path 1/Path 2
      separation exists deliberately, so (a) is cleaner.
- Gap 2 (medium): give component-creator (and the planner) a way to mark a
  component data-driven, so creating it ALSO seeds the loader js_snippet and
  registers the data feed need. This is a new field-source tier or component
  flag, aligned to doc 022's "dynamic component" tier.
- Gap 1 (largest, but well-templated): clone the news pipeline as the provocation
  data pipeline (render_provocations_section + provocation-orchestrator +
  scheduled task). Per PLAN_spark_provocation_pipeline.md.

ORDER: fix Gap 3 first (unblocks the JS we already inserted, benefits every site),
then prove the shell fills (P2-5), then Gap 1 (data), then Gap 2 (make it
automatic for the next site/component).

---

## 2026-06-25 ~19:30 — js_snippets inventory (B3/B4) reveals a path question

Full js_snippets inventory:
| name | applies_to | active | bytes |
|---|---|---|---|
| accordion | faq, accordion | f | 560 |
| copy-to-clipboard | code, share | f | 372 |
| counter-animate | stats, numbers, social-proof | f | 815 |
| form-validation | form, newsletter, contact | f | 671 |
| lazy-load-images | image, gallery | f | 509 |
| mobile-menu-toggle | navigation, header | f | 338 |
| news-date-formatter | latest-news, news-listing | t | 506 |
| provocation-card-loader | provocation-card | t | 4879 |
| scroll-reveal | section, card, feature | f | 407 |
| smooth-scroll | navigation, link | f | 307 |
| typing-effect | hero, headline | f | 538 |

KEY OBSERVATION — every pre-existing snippet is a GENERIC, cross-component
behaviour applying to 2-3 component types (accordion, scroll-reveal, mobile-menu,
form-validation, etc.). They are small (≤815 bytes). The news one,
`news-date-formatter` (506 bytes), applies to BOTH latest-news AND news-listing
and by its name is a DATE FORMATTER — a shared helper, NOT a data fetcher.

Our `provocation-card-loader` is the outlier: component-specific (applies to one
component), and 4879 bytes (6× the largest other). It is a FETCHER+RENDERER, not
a generic behaviour.

THIS RAISES A PATH QUESTION (do not assume — confirm):
- If `news-date-formatter` is only a formatter, then the news data FETCH must
  live somewhere else — most likely the `latest-news` COMPONENT's own inline
  <script>, which is extracted to content_components.js_content and deployed as
  /tools/assets/latest-news.js on every page rerender (Path 1, automatic).
- If that's how news works, then the architecturally-consistent home for the
  provocation FETCH logic is the provocation-card component's OWN inline <script>
  (Path 1), NOT a bespoke js_snippet (Path 2). provocation-card ALREADY has an
  inline <script> (card hover/activation) that becomes /tools/assets/
  provocation-card.js. Adding the fetch-and-fill there means:
    * it deploys automatically on page rerender (NO site-asset-renderer / Gap 3
      dependency at all)
    * it's co-located with the DOM it fills
    * it matches how every other interactive component ships its JS
    * js_snippets stays what it appears to be: a home for GENERIC behaviours

This would mean our Path-2 js_snippet (provocation-card-loader) is against the
grain — it works (once site-asset-renderer runs) but it's not where this kind of
logic belongs in this system.

TWO CONFIRMATION QUERIES NEEDED before committing to a path:
1. latest-news component template — does it have an inline <script> that
   fetch()es /data/latest-news.json? (Confirms news uses Path 1 for the fetch.)
   SELECT html_template FROM content_components
   WHERE function='latest-news' AND is_active=true AND forked_from IS NULL;
2. provocation-card's extracted js_content — is the existing inline script
   already in js_content (so /tools/assets/provocation-card.js deploys)?
   SELECT function, LENGTH(js_content) AS js_len,
          (html_template ILIKE '%<script%') AS has_inline
   FROM content_components
   WHERE function='provocation-card' AND is_active=true AND forked_from IS NULL;

PRAGMATIC INTERIM: the Path-2 snippet is already inserted. To unblock vonc now we
can still trigger site-asset-renderer and prove the fetch-fill works end-to-end
(P2-4/P2-5). If the confirmation shows Path 1 is the right home, we MOVE the
loader into the component's inline script later and retire the snippet — but the
JSON shape and fill logic we've written are reused verbatim either way.

design-discovery-agent (confirmed): runs run_discovery_checks with an array of
~15 named checks (stale_site_components, missing_css, undeployed_assets, etc.),
check_domain='design', writes findings to site_work_items, no LLM. A new check
slots into that array cleanly IF we go the Path-2/Gap-3 route. If we go Path 1,
Gap 3 is moot for this case (page rerender already ships the component JS).

---

## 2026-06-29 ~17:20 — PD queries answered + loader proven live + lobby-text gap

### site-asset-renderer ran clean
Triggered via the generic entry point (manual-assetrender-20260629). Orchestration
COMPLETED. render_js_snippets → git_commit (commit eb7f2ac "Update JS snippets
bundle"). snippet_count=1, snippet_names=[provocation-card-loader]. The git-adapter
(long-running adapter) handled the commit. /assets/js/snippets.js now live with the
loader. provocations.json present in /data/ (commit "Create provocations.json").

### PROOF: the loader works
Screenshot of vonc.com/index.html: the provocation-card now FILLS.
- Main provocation: "AI will <em>never</em> be funny on purpose." (em in accent pink) ✓
- AI take body: "The AI's take: humour needs a victim and a risk..." ✓
- Stat strip: 1,284 Positions Filed / 3h 12m Until Close / 62% Disagree ✓
- Lobby icons: 🔥 ⚡ 🧠 🎯 all four render ✓
- CTAs: "File Your Position →" + "See All Provocations" ✓
The Path-2 (js_snippet) delivery is confirmed working end-to-end.

### REMAINING GAP: lobby card title + desc blank (show as faint "...")
Icons render but .pc-card-title and .pc-card-desc are empty/"...". The loader sets
all three (icon, title, desc) in the SAME loop, so the loop runs and matches cards
(icons prove it). Therefore the committed provocations.json lobby entries have
`icon` but missing/empty `title`/`desc` — the committed file (commit "Create
provocations.json", predates this session's sample) is an earlier/stub version,
NOT the full provocations.sample.json we wrote (which has title+desc per entry).
NOT a loader bug, NOT a CSS bug — a data-content gap. Fix = the real data pipeline
(Phase 3) or, interim, commit the complete JSON.

### PD-1 ANSWERED — news uses Path 1 (inline script in the component) ✓✓✓
latest-news component html_template ends with:
  <script src="/tools/assets/latest-news.js"></script>
and its body is the DOM shell (news-container, news-footer, noscript fallback)
with NO inline fetch in the template itself — the fetch lives in
/tools/assets/latest-news.js, which is the component's extracted js_content
(Path 1). So NEWS DELIVERS ITS FETCH VIA PATH 1, not via a js_snippet. The
news-date-formatter snippet (Path 2) is only a helper, exactly as suspected.

THIS DECIDES PD-3: the architecturally-consistent home for the provocation
fetch-and-fill is the provocation-card COMPONENT's own JS (Path 1) — i.e.
content_components.js_content → /tools/assets/provocation-card.js, which deploys
automatically on every page rerender. NOT a bespoke js_snippet.

### PD-2 ANSWERED — but reveals a SEPARATE bug
provocation-card: has_inline_script = TRUE (the <script> is in html_template),
but js_len = NULL/empty. So the inline card-activation script was NEVER extracted
to js_content. That means:
- /tools/assets/provocation-card.js is NOT being produced (collectJSAssets reads
  js_content, which is empty).
- The card-hover/activation interactivity isn't deploying either.
- The extraction step (separateInlineJS, per component-creator contract "the
  pipeline will automatically extract inline <script> blocks into a separate JS
  asset") did NOT run for this component, OR ran before the script was added, OR
  the script is malformed so extraction skipped it.
- RECALL the html_template dump showed the inline <script> TRUNCATED mid-function
  with a stray backslash. If the stored template's script is actually broken/
  unterminated, separateInlineJS may have failed to extract it. NEEDS A LOOK.

### REVISED FRAMEWORK DIRECTION (Path 1 chosen)
Given PD-1 (news=Path1) and the doctrine of consistency:
- The provocation loader belongs in the provocation-card component's js_content,
  deployed as /tools/assets/provocation-card.js (Path 1, auto on rerender).
- This means Gap 3 (site-asset-renderer not auto-triggered) is NOT on the critical
  path for this case — Path 1 deploys via the normal page rerender. (Gap 3 is still
  a real latent issue for genuinely-generic snippets, but lower priority.)
- The js_snippet we inserted (provocation-card-loader, Path 2) WORKS and is what's
  live now — keep it as the working interim, but the durable home is Path 1.
- Gap 2 (auto-provision loader for data-driven components) becomes: component-creator
  should, for a data-driven component, put the fetch-fill loader in the component's
  own <script> (so it extracts to js_content), AND register the data feed need.
- SEPARATE BUG surfaced: provocation-card's inline script isn't extracted to
  js_content (js_len empty despite has_inline_script). Investigate separateInlineJS
  + the possibly-truncated stored template before relying on Path 1 here.

### NEXT ACTIONS (revised order)
1. Fix the lobby-text now: ensure committed provocations.json has title+desc per
   lobby entry (interim: commit the full sample; proper: Phase 3 generates it).
2. Investigate why provocation-card js_content is empty despite inline <script>
   (separateInlineJS / truncated template). This blocks Path 1 adoption.
3. Decide: migrate loader into component js_content (Path 1) once extraction works,
   then retire the Path-2 snippet — OR keep Path 2 as a deliberate exception.
4. Phase 3 data pipeline (render_provocations_section etc.) regardless of path.

---

## 2026-06-29 ~17:35 — Lobby text gap resolved; Phase 2 proven complete

Confirmed: the blank lobby titles/descriptions were the stub `"title":"..."` /
`"desc":"..."` in the committed provocations.json. Re-committing the FULL JSON
(real lobby content) fixed it. Screenshot shows all four lobby cards fully
populated: "Remote work killed mentorship" / "Privacy is already over" /
"Reading fiction makes you worse at facts" / "Most 'data-driven' decisions
aren't", each with its description. The headline, AI take, stats strip and CTAs
were already correct.

PHASE 2 IS PROVEN: the runtime JS layer (fetch /data/provocations.json → fill
the provocation-card shell) works end-to-end on the live site. No loader bug, no
CSS bug — the only issue was abbreviated test data, now corrected.

This validates the whole Option-2 approach: a static shell + a client-side loader
+ a JSON data file produces the daily-provocation card. Phase 3 (the pipeline that
GENERATES provocations.json from scraped/curated content) just needs to emit this
exact JSON shape.

### State of the two follow-ups
1. Lobby text — RESOLVED (data fix).
2. Path decision — DECIDED Path 1 (component js_content), but blocked on the
   extraction bug (provocation-card has_inline_script=TRUE, js_content empty).
   The Path-2 snippet remains live and working as the interim. Migrating to Path 1
   is not urgent now that the card works; it's a consistency/robustness improvement
   to schedule alongside fixing the extraction bug.

### Confirmed-good provocations.json shape (the Phase 3 generation target)
{
  "generated_at": ISO8601,
  "today": {
    "eyebrow", "headline" (may contain <em>), "body",
    "primary_cta": {"label","url"}, "secondary_cta": {"label","url"},
    "stats": [{"value","label"} x3]
  },
  "lobby": [{"icon","title","desc","url"} x4]
}
Loader fills: .pc-eyebrow, .pc-headline (innerHTML for em), .pc-body,
.pc-btn-primary/.pc-btn-secondary (href + label, SVG preserved),
.pc-stat-value/.pc-stat-label x3, .pc-card x4 (.pc-card-icon/.pc-card-title/
.pc-card-desc, + click/keyboard nav from url).

---

## 2026-06-29 ~18:00 — P2-6: brief-explanation analysed; lobby-grid still pending

### brief-explanation — it is STATIC; fix is regeneration, NOT a JS loader
Template inspected (full). Structure: `.be-eyebrow`, `.be-heading` (id be-heading,
supports `<mark>`), `.be-description`, an `.be-steps` ordered list of 3 ×
`.be-step` (each `.be-step-number` hardcoded 1/2/3 + `.be-step-text` with a
`<strong>` title and body text), `.be-stats` (3 × `.be-stat-value`/`.be-stat-label`),
`.be-cta-group` (`.be-btn-primary` + `.be-btn-secondary`, each href+label), and a
`.be-visual` figure with `<img src alt>` + `.be-visual-badge`.

This is a "what Spark is / how it works" explainer — STABLE BRAND CONTENT, not
daily-changing data. Per the roadmap, the about-style explainer copy doesn't change
day to day.

State: same Mode-B breakage as provocation-card — the stored html_template is full
of bare `<no value>` (field names lost) and input_schema is `{}`. CSS selectors are
intact.

DECISION: brief-explanation gets REGENERATED via component-creator with a real
input_schema, NOT a JS loader. Reasons:
- Static content must be in build-time HTML for SEO and to survive JS-off.
- A loader would render `<no value>` if fetch/JS failed.
- The content writer is the right producer for stable brand copy.
Proposed schema (Tier A voice unless noted):
  eyebrow, heading (may contain <mark>), description,
  step1_title, step1_text, step2_title, step2_text, step3_title, step3_text,
  stat1_value, stat1_label, stat2_value, stat2_label, stat3_value, stat3_label
    (stats: static/illustrative for launch — Tier A or Tier B with fallback),
  cta_primary_label, cta_primary_url, cta_secondary_label, cta_secondary_url
    (urls likely Tier B with sensible fallbacks, or Tier A),
  badge_text,
  image: Tier C from site illustration asset (site_record.illustration_url =
    /assets/images/illustration.jpg), alt text Tier A.
Then a needs_page rebuild fills it at build time.

IMPORTANT DISTINCTION (so the runbook's "Option 2 for empty shells" isn't
misapplied): Option 2 (runtime JS loader) is ONLY for DATA-DRIVEN shells whose
content changes daily (provocation-card; possibly lobby-grid). STATIC shells that
happen to be empty (brief-explanation) are fixed by REGENERATION with a proper
schema. Two different root-cause resolutions for the same surface symptom
(empty shell).

NOTE on provocation-card's underlying template: it is ALSO Mode-B broken; the JS
loader masks it by targeting selectors and overwriting textContent. If JS fails,
the user sees `<no value>`. A more robust version would regenerate provocation-card
to a clean template whose default (pre-JS) content is sensible/empty rather than
`<no value>` (renderer-source fields with fallbacks). Not urgent — logged as a
refinement.

### RECURRING-CATEGORY candidate: "Mode-B empty shell"
provocation-card, lobby-grid, brief-explanation ALL share: html_template = bare
`<no value>` (names lost) + input_schema `{}` + 0 slots, but intact CSS selectors.
This points to a component-CREATION-era fault (an early/!=current creation path
that stored rendered `<no value>` output as the template, losing slots+schema), OR
a save path that overwrote html_template with rendered output. Worth confirming as
a class, because the FIX differs by content type (dynamic→loader, static→regenerate)
but the CAUSE is one bug. Flag for the extraction-bug investigation that follows.

### lobby-grid — STILL PENDING (need the data)
Queries 1a (metadata) and 1b (lobby-grid html_template) were not run/pasted yet —
only 1c (brief-explanation) came back. Re-run 1a + 1b. Open question: lobby-grid is
likely a "today's rooms/provocations" grid (DYNAMIC → loader candidate, could read
the provocations.json `lobby` array or a rooms feed) — BUT provocation-card already
contains a mini-lobby (.pc-card × 4). So either lobby-grid is a fuller rooms list
(keep, give it a loader/feed) or it duplicates the mini-lobby (consider dropping it
from the index plan). Decide after seeing its template.

---

## 2026-06-29 ~18:10 — Adopting per-tool travelling documentation

Decision/direction (user): bake documentation into the tool lifecycle so each
tool carries its own reasoning history — domain → spec → tool creation →
maintenance, with a PLAN written at creation and NOTES accumulated on every
change, so bug-fixing a tool is always fully informed and fixes pull the same way.

Convention written: TOOL_DOCS_convention.md. Summary:
- Two docs per tool, keyed by component `function`:
  PLAN_<function>.md (intent: aim, source spec, behaviour + data/DOM contract,
  delivery mechanism Path1/Path2/build-time, dependencies, deliberate decisions)
  and NOTES_<function>.md (timestamped history: choices+decisions+rationale,
  bugs symptom→cause→fix→verify, recurring-category tags, dead ends).
- Problem-category taxonomy for tagging notes (css-variable-mismatch, empty-shell/
  mode-b-template, broken-template-slots, content-vs-runtime-mismatch,
  detool-on-rebuild, js-not-extracted, js-bundle-stale, schema-template-drift).
- Storage evolution: markdown now → DB `tool_docs` table keyed by function
  (source of truth, travels through forks/regenerations) → optional repo mirror.
- Pipeline integration (future feature, spec'd in the convention): tool-generator
  writes PLAN at creation (capturing the LLM design reasoning currently
  discarded); maintenance agents append NOTES; a bug entry-point loads PLAN+NOTES
  FIRST. Category tags roll up into the global debugging guide (016/016b).

Relationship: global debug guide = cross-tool patterns; site runbook = one
site's state; per-tool docs = one tool across all sites. Per-tool NOTES feed UP
into the global guide via category tags.

For the current work: gauntlet-interface, tool-archetype-taster-quiz, and the
provocation daily-feed feature are the candidates. The provocation feature is
already effectively documented at site level (RUNNING_NOTES_vonc +
PLAN_spark_provocation_pipeline) — those can be split to PLAN_/NOTES_ per
convention when convenient. Adopt going forward (instantiate as we touch each
tool) rather than a big retroactive migration.

---

## 2026-06-29 ~18:30 — lobby-grid analysed

Metadata (1a): brief-explanation and lobby-grid BOTH: schema {}, slots 0,
quality 50. brief-explanation has_inline_script=f, js_len=0. lobby-grid
has_inline_script=TRUE, js_len=0 → SAME extraction bug as provocation-card
(inline script not extracted to js_content → /tools/assets/lobby-grid.js not
produced; its hover/focus/entrance-animation effects aren't deploying).

lobby-grid structure (full template): header (`.lobby-grid-section__eyebrow`,
`__title` with `<em>`, `__subtitle`) + a `__grid` of 6 cards — 1 `--featured`
(spans 2 cols), 4 standard, 1 `--wide` (full row). Each card:
`__card-icon` (inline SVG), `__card-tag`, `__card-title`, `__card-desc`, and a
`__card-stat` line with a pulsing `__card-stat-dot` (live indicator). Then
`__cta` (`__cta-label` + `__cta-btn`). Mode-B broken (`<no value>`, empty schema).

NATURE: this is the ARENA LOBBY — "grid of active rooms" from the original Spark
design ("rooms not feeds"). DYNAMIC (today's active rooms/provocations) → loader
candidate, same pattern as provocation-card.

TWO DECISIONS FLAGGED (not resolved):
1. OVERLAP with provocation-card mini-lobby. provocation-card already has a 4-card
   mini-lobby ("today's other provocations"). lobby-grid is a fuller 6-card grid
   doing essentially the same job, richer. Recommendation: make lobby-grid THE
   "today's provocations grid" for the index, and trim provocation-card's mini-lobby
   to keep the hero card focused on the single daily provocation. Keeps one clear
   home for the list. DECISION NEEDED.
2. v1 SEMANTICS. "Live rooms" is a v3 concept (roadmap: v3 = live challenge rooms).
   In v1 (daily provocations + Gauntlet) there are no live rooms yet. Honest v1
   reading: lobby-grid shows TODAY'S PROVOCATIONS as enterable cards (each card =
   a provocation/topic), the "stat" line showing e.g. positions-filed rather than
   a live-room count. So its data = a richer provocations list (6 items with
   tag/title/desc/stat/icon), produced by the Phase 3 pipeline. Interim: extend
   provocations.json with a `rooms` (or reuse/expand `lobby`) array of up to 6
   richer entries.

DELIVERY: same as provocation-card — Path 2 (js_snippet) interim, Path 1
(component js_content) durable target, blocked on the extraction bug. Do
provocation-card and lobby-grid consistently (same path, same time).

So lobby-grid is NOT actioned yet: blocked on (a) the overlap/semantics decision,
(b) the extraction bug (shared, next item in our order), (c) Phase 3 data. Its
per-tool docs will capture these open decisions.

### Updated per-component plan for the index shells
- provocation-card: DONE via Path-2 loader (live, working). Durable: migrate to
  Path 1 after extraction bug fixed. [dynamic]
- lobby-grid: dynamic loader, BLOCKED on overlap decision + extraction bug + data.
  [dynamic]
- brief-explanation: REGENERATE with a real schema; content-writer fills at build.
  NOT a loader. [static]

---

## 2026-06-29 ~18:45 — Per-tool docs: started; doc-storage route assessed

DOC-STORAGE DECISION (critical assessment in TOOL_DOCS_convention.md Appendix A):
**Files now, hybrid later, structured for migration.** Do NOT build a
documentation database yet (zero docs instantiated = premature; humans authoring
now; git gives versioning+review free). Library-level tool docs live in the
agentchassis repo (e.g. /docs/tools/<function>/), NOT in per-site repos (a
component is cross-site). Keep NOTES entries import-shaped (uniform dated headers
+ a `Categories:` line). Introduce a DB layer ONLY when a trigger fires — agents
start writing notes, or cross-tool tag-queries become routine — and even then a
HYBRID: NOTES → a tool_doc_notes table (append rows + tag column for SQL roll-up);
PLAN → STAYS in git (versioned), maybe a thin DB index row. Avoid the worst option
(PLAN as an unversioned DB text column — discards git history for the artifact
whose job is preserving history). One-liner: git→hybrid, not DB-first.

STARTED INSTANTIATING per-tool docs (worked exemplar):
- PLAN_provocation-card.md — aim, source spec, behaviour + data/DOM contract,
  delivery (Path 2 live / Path 1 durable target), dependencies, deliberate
  decisions (JS-required by design), known limitation.
- NOTES_provocation-card.md — full timestamped history with category tags:
  empty-shell/mode-b-template → Option-2 decision → loader inserted → js-bundle-stale
  (manual site-asset-renderer trigger) → proven working → lobby data fixture fix →
  Path-1 decision (PD-1/2/3) → OPEN extraction bug.
provocation-card chosen as the first exemplar because it has the richest complete
story. lobby-grid and brief-explanation docs to follow as we action them (their
open decisions are already captured here in the site notes and will move into
their PLAN/NOTES).

CROSS-TOOL PATTERN already visible (argues for the eventual tag roll-up):
`js-not-extracted` now affects BOTH provocation-card and lobby-grid;
`mode-b-template`/`empty-shell` affects provocation-card, lobby-grid,
brief-explanation. These are the categories that should graduate into the global
debugging guide once the extraction bug is root-caused.

---

## 2026-06-29 ~19:30 — EXTRACTION BUG root-caused (split result is decisive)

Got the store action (store_generated_component_action.go) + rerender
(rerender_single_page_action.go) + the has_script_close SQL. Findings:

CODE IS CORRECT (not the bug):
- separateInlineJS (store action, line 105, called unconditionally): regex
  `(?s)<script\s*>(.*?)</script>` — REQUIRES a closing </script>. Removes the
  inline block, puts JS in js_content, and replaces it in the template with
  `<script src="/tools/assets/{function}.js">`. So a normally-stored component
  has a src-ref (NOT a raw inline script) + populated js_content.
- collectJSAssets (rerender): just reads js_content; correctly returns nothing
  when empty. Exonerated.

SQL SPLIT (decisive):
- lobby-grid: has_script_close=15644 (</script> PRESENT, script INTACT), BUT the
  stored template still ends with the RAW `})();</script>` IIFE — NOT a src-ref —
  and js_content is empty. A working extraction would have replaced the raw script
  and filled js_content. It didn't. => separateInlineJS NEVER RAN on lobby-grid.
  And since the script is intact, this is NOT a malformed-script bail.
- provocation-card: has_script_close=0 (NO </script>), template tail ends
  mid-function with a stray backslash => the stored script is genuinely TRUNCATED
  in the DB (a separate corruption on that row).

ROOT CAUSE (high confidence; one confirming query pending):
These Mode-B components were stored through a path that did NOT run
separateInlineJS — most likely they PREDATE its addition to the store action
(or a seed/bulk path bypassed it), and have never been regenerated since. So they
keep the old shape: raw inline script still in template, empty js_content, empty
schema, `<no value>` placeholders. lobby-grid's script happened to be complete;
provocation-card's was ALSO truncated at generation time (separate corruption).
Neither separateInlineJS nor collectJSAssets is buggy.

STRUCTURAL GAPS exposed (fix to prevent recurrence):
1. Store-path validation Check 2 flags unclosed <style> only — NO check for
   unclosed <script>. A template truncated mid-<script> (provocation-card) passes
   validation. ADD a <script> open/close balance check alongside the <style> one.
2. separateInlineJS SILENTLY returns empty on a <script> with no close. Make it
   WARN (logger.Info/Warn) when it sees a <script> opener with no matching close,
   so truncation is visible.

CONFIRMATION PENDING — healthy comparison (query 4, refined):
  SELECT function,
         LENGTH(COALESCE(js_content,'')) AS js_len,
         (html_template LIKE '%<script src=%') AS has_src_ref,
         (html_template LIKE '%<script>%')     AS has_raw_inline
  FROM content_components
  WHERE function IN ('gauntlet-interface','tool-archetype-taster-quiz','latest-news')
    AND is_active=true AND forked_from IS NULL
  ORDER BY function;
Expectation if theory holds: healthy components show js_len>0, has_src_ref=TRUE,
has_raw_inline=FALSE (raw script extracted out, replaced by src-ref). latest-news
already shows the src-ref pattern (its template ends `<script src="/tools/assets/
latest-news.js">`), so this mainly confirms gauntlet-interface + archetype-quiz.
If confirmed, the three Mode-B shells are the exception and the current path is
safe for NEW components.

FIX DIRECTION (structural; converges with Path-1 migration):
- Existing broken rows: REGENERATE the three through the current store path so
  separateInlineJS runs + schema/template rebuilt. lobby-grid could in principle
  be re-stored (script intact) but it's also Mode-B (empty schema) so full
  regeneration is the consistent fix. provocation-card MUST be regenerated
  (truncated source — nothing to re-extract).
- For Path 1: regenerate with a COMPLETE inline <script> containing BOTH the
  interactivity AND the data fetch-fill loader; separateInlineJS extracts it to
  js_content -> /tools/assets/{function}.js deploys on rerender; retire the
  Path-2 snippet afterwards.
- Harden validation (+<script> balance check) and separateInlineJS (+truncation
  warning) so new components can't regress.

---

## 2026-06-29 ~20:00 — separateInlineJS robustness Qs + locating the generation path

### separateInlineJS — attribute skip is deliberate and CORRECT (keep it)
Q: why does it only extract bare `<script>` (regex `<script\s*>`), skipping
attributed tags? A: good reason — skipping protects three cases that must NOT be
naively extracted:
- `<script src=...>`: external ref, no inline body; extracting breaks the ref.
- `<script type="application/ld+json">`: JSON-LD SEO data; must stay inline;
  extracting strips SEO + produces a .js file full of JSON.
- `<script type="module">`: module semantics; extracting + referencing via plain
  `<script src>` changes execution + breaks imports.
VERDICT: do NOT change it to extract everything. Mild gap: plain JS wrapped in
`<script defer>` / `type="text/javascript"` is safe to extract but currently
stays inline — BENIGN (still runs, just inline/duplicated), rare in generated
components. Optional improvement: log when it leaves an attributed `<script>`
inline (observability), alongside the unterminated-`<script>` warning already
proposed. Low priority; NOT our bug's cause.

### separateInlineJS — multiple `<script>` blocks: handled correctly
ReplaceAllStringFunc matches all; capture `(.*?)` is LAZY (+`(?s)`), so each
`<script>` pairs with its OWN nearest `</script>` (no over-match across two).
Bodies are combined: jsContent = Join(jsBlocks, "\n\n"); one `<script src>` ref
added. => When regenerating dynamic components we may use ONE `<script>` (both
interactivity + loader) OR TWO separate blocks; both end up combined in js_content.
Only real fragility: a literal `</script>` inside a JS string (spec requires
escaping as `<\/script>`); well-formed code avoids it; our components don't hit it.

### Generation path: NOT a rerender agent (confirmed from the attached defs)
All four rerender agents assemble/dispatch, none generates a template:
- page-rerender: rerender_single_page (assemble stored rendered_html) +
  rerender_page_sections (re-render from STORED content_data, explicitly no LLM).
- rerender-site: render_site_components + loop page-rerender.
- rerender-pages: create_rerender_items (+ render_site_components, snippets).
- site-asset-renderer: render_js_snippets only.
None calls the LLM-generation step or store_generated_component → none rebuilds an
html_template → none runs separateInlineJS. Content-fill (page-build-handler →
page-content-writer) is ALSO not it: it fills an EXISTING template, doesn't
regenerate it.

So the generation path = the component-generation agent that owns the LLM step
producing `generated_template` and hands it to StoreGeneratedComponentAction
(which runs separateInlineJS). It isn't named in the %build%/%render% lists.
LOCATING IT by behaviour (queries issued):
  SELECT type, agent_category, description FROM agent_definitions
  WHERE is_active=true AND default_config::text LIKE '%store_generated_component%'
  ORDER BY type;
  SELECT type, description FROM agent_definitions
  WHERE is_active=true AND (type LIKE '%component%' OR type LIKE '%section%'
        OR type LIKE '%tool%') ORDER BY type;
Then: find which WORK-ITEM type triggers a SINGLE named component regeneration
(vs only generating missing components during a full build), and point it at the
three shells.

### Consequence shaping the fix (Gap 2 dependency)
Regenerating re-runs the LLM. The creator must know provocation-card and
lobby-grid are DATA-DRIVEN so it emits the inline `<script>` with the fetch-fill
loader (old Gap 2 — creator had no data-driven tier). If it doesn't, regeneration
yields a clean template + interactivity script, and we fold the loader into that
script as a deliberate step. brief-explanation is static → plain regenerate with a
real schema suffices.

---

## 2026-06-29 ~20:30 — Generation path = component-creator; regeneration mechanics; Gap 2 confirmed

GENERATION PATH (answer to "rerender agent or build agent?"): NEITHER a rerender
agent NOR a site/page builder — it's **component-creator** (builder category). Only
agent calling store_generated_component (which runs separateInlineJS). Workflow:
ensure_site_record → read_site_spec → generate_template (execute_llm_prompt,
claude-sonnet-4-6, max_tokens 16000) → store_component → complete. Processes
`needs_new_component` work items. Input: required section_type; optional site_type,
page_context, description, design_direction, reference_content. Trigger pattern
(080 script): spawn_agent + call_agent via generic entry point with input_mapping.

REGENERATION MECHANICS (from store_generated_component_action.go, GOOD news):
- Looks up existing component by `function` (forked_from IS NULL, is_active DESC,
  updated_at DESC). If found → isRegeneration=true. (is_active filter removed
  2026-05-06 so a deactivated row still regenerates in place.)
- UPDATE IN PLACE: preserves component_id → all page_components/site_components FKs
  keep resolving, NO relink. Snapshots current row to component_versions (MAX+1)
  BEFORE update (history preserved). Sets html_template, input_schema, js_content
  (NULL if empty), is_dark_section, render_mode (derived), is_active=true.
- After UPDATE: markPagesPendingRebuild + raises ONE needs_rerender work item per
  affected site (deduped by site_id+item_key scoped to component_id) so rerender
  pipeline actually rebuilds. ScoreAndPersistComponent + markPagesForRebuild on
  both paths.
- So regenerating provocation-card/lobby-grid/brief-explanation is clean:
  in-place, FK-safe, history-kept, auto-rerendered.

PRE-STORE VALIDATION already blocks our bug class:
- REJECTS any template containing `<no value>` (comment explicitly describes the
  Mode-B cause: render output stored back as source, placeholders unrecoverable →
  regeneration required). This post-dates the broken components (fits root cause).
- Also checks placeholder/schema count parity, unclosed <style>, presence of
  <section>/<div>. STILL MISSING: unclosed <script> check (our hardening item).

GAP 2 CONFIRMED from the component-creator PROMPT:
Tiers are A (voice/llm), B (tunable labels/static+fallback), C (site data/
site_specs|site_assets), D (derived lists/query.{name} — resolved at PLAN time by
DB query), plus "renderer" source (single JS-filled value + fallback, e.g. timer).
There is NO tier for content fetched CLIENT-SIDE from a JSON file at RUNTIME (the
daily-feed pattern). So:
- brief-explanation (STATIC): regenerate as-is — Tier A/B/C cover it. No gap.
- provocation-card & lobby-grid (RUNTIME FEED): regenerating as-is would classify
  the provocation text as Tier A and bake a build-time provocation into the
  template — WRONG shape, loses the daily loader. Would fix <no value>/schema/
  extraction but regress the proven Option-2 design.

STRUCTURAL FIX for the dynamic two (proposed, pending decision): add a runtime-feed
tier to component-creator's prompt — Tier E, source "feed.{name}" — instructing the
LLM to emit a stable-selector DOM shell + an inline <script> loader that fetches
/data/{feed}.json and fills it, declaring the data contract. Mirror how Tier D
encodes a canonical range-block ("copy this form exactly"); hand the LLM a canonical
loader pattern based on our PROVEN provocation-card-loader. Then regeneration puts
the loader inline → separateInlineJS extracts → /tools/assets/{function}.js (Path 1),
retire the Path-2 snippet. Benefits all future daily-feed components.

SEQUENCING (proposed): (1) regenerate brief-explanation now (no gap, proves path);
(2) design Tier E prompt extension for the dynamic two — GET DECISION before writing.
Plus the <script>-balance validation hardening as a separate flagged change.

TWO DETERMINISM CONFIRMATIONS NEEDED before regenerating:
1. Regeneration keys on `function` from the LLM OUTPUT. The LLM must reproduce the
   SAME function name (provocation-card etc.) for in-place UPDATE; if it picks a
   different name, store_generated_component INSERTs a duplicate instead. Need to
   ensure the function name is pinned (pass intended function, or confirm the LLM
   reliably mirrors section_type→function). Risk is low for well-known types but
   real.
2. How is regeneration normally TRIGGERED for an existing low-quality component?
   `component-quality-auditor` "creates regeneration work items for low-quality
   ones" (our three are quality 50). Confirm its work-item shape — does it pass
   function/component_id for deterministic in-place regen, or just section_type?
   That's the more robust trigger than a raw needs_new_component.

---

## 2026-06-29 ~21:00 — Option A approved: regenerate brief-explanation (trigger ready)

QUALITY-AUDITOR finding (from its default_config): it creates
`needs_component_regeneration` items ONLY for components scoring **< 50**
(condition `quality_score < 50`), handler_agent=component-creator, spec_data
{function, component_id, quality_score, quality_issues}, item_key_prefix
quality_regen. Our three shells are EXACTLY 50 → NOT auto-picked-up (explains zero
queued items). So we trigger manually. The auditor's item shape DOES confirm the
designed regen path keys on `function` + routes to component-creator.

KEY MECHANIC (re)confirmed: store_generated_component looks up the existing row by
the LLM's EMITTED `function` (not by spec_data.function). So in-place UPDATE needs
the LLM to emit function='brief-explanation'. Passing section_type drives that
(reliable for a known name) but it's NOT pinned → must verify in-place vs duplicate.

TRIGGER BUILT: 081_regenerate_brief-explanation_vonc.sh
- Based on the project's 080 component-creator trigger (spawn_agent + call_agent
  via generic entry point, JSON embedded literally — NO jq dependency, matching 080).
- DELIBERATE ADDITIONS over 080 (flagged in-file): site_id + domain added to BOTH
  input_data and the call_creator input_mapping, so component-creator's
  ensure_site_record + read_site_spec resolve vonc and the LLM generates on-brand;
  plus a structure-focused description (eyebrow, heading w/ emphasis, description,
  EXACTLY 3 numbered steps, EXACTLY 3 stats, 2 CTAs, illustration + badge; dark
  section; voice fields as placeholders, image from site asset, stats as tunable
  labels) and a Spark design_direction (dark/game-energy, --color-* vars only).
- bash -n OK; embedded JSON parses OK.

PRE-REGEN BASELINE QUERY prepared (record id + active-row count before running) so
we can confirm the UPDATE is IN PLACE (same component_id, exactly one active row,
real schema, no <no value>, has {{placeholder ...}}).

TWO NUANCES FLAGGED for after the run:
1. Determinism: verify status='regenerated' (not 'created'), same component_id,
   active_rows=1. A 2nd row = LLM emitted a different function name → duplicate
   INSERT; deactivate dupe + re-run.
2. Content fill: regen auto-raises needs_rerender (ASSEMBLE only) — the new
   placeholders have no content_data yet, so the section may show fallbacks until
   the index is rebuilt by the content writer (needs_page → page-content-writer).
   We'll check which items the regen raised (query 3) before deciding whether to
   trigger a needs_page index rebuild.

DESIGN NOTE: regeneration produces a FRESH template (the LLM does not see the old
one) — the new markup/CSS will differ from the current broken shell. That's
expected and fine (the old template is unusable); the description pins the intended
structure so the result matches the brief-explanation design.

NEXT (after this proves out): design the Tier E runtime-feed prompt extension for
the two DYNAMIC shells (provocation-card, lobby-grid) before regenerating them.

---

## 2026-06-30 ~18:35 — brief-explanation regen MISFIRED (input not nested under spec); fix = 082

081 ran end-to-end (spawn → call → generate_template → store_component → complete)
so the pipeline works, BUT it produced a STRAY component, not a regeneration:
  stored_component: function='general-hero', display_name='General Hero',
  status='created', component_id=0ef52c95-2111-4e1f-a924-308d7a7eeab2, quality 100.
The existing brief-explanation (id 58363894-9db9-4d2f-81ac-c47b54d97fc3) is
UNTOUCHED (still quality 50, 0 slots, has_no_value=t, active_rows=1). No work items
raised (the stray is general-hero; nothing was waiting on that).

ROOT CAUSE — input shape (the caveat flagged before running, now confirmed):
component-creator reads input_data.spec.section_type / input_data.spec.description
(work-item convention: spec_data → input_data.spec). The 081 trigger, copied from
the 080 test script, mapped fields to input_data.section_type / input_data.description
(TOP-LEVEL). So the generate_template LLM prompt received an EMPTY section_type +
description and defaulted to the most generic component — a hero named
'general-hero'. Because the LLM emitted function='general-hero',
store_generated_component looked up function='general-hero', found nothing, and
INSERTed a new row (status created) instead of finding + updating the existing
'brief-explanation' row (lookup keys on function). Hence a stray + brief-explanation
untouched. (The section_type='brief-explanation' that still appears in the output is
recovered downstream; the decisive signals are the generic-hero template + the
mismatched function name.)
Evidence the cause is nesting: 080 maps "section_type"→input_data.section_type; the
callee reads input_data.spec.section_type; top-level keys don't resolve. So the
080 direct-call test pattern is INSUFFICIENT for component-creator — it needs the
inputs nested under spec. (Lesson: the 080 script likely never produced a correct
context-driven component either.)

CLEANUP: deactivate the stray general-hero (0ef52c95) after confirming it's
unreferenced (page_components/site_components/area_components all 0). SQL provided.
status='created' (new_version=1) confirms no prior 'general-hero' existed, so it's
definitively the row we just made — safe to deactivate.

FIX — 082_regenerate_brief-explanation_vonc.sh (supersedes 081):
- input_data now nests the creator inputs under a `spec` object:
  input_data.spec = {section_type, site_type, page_context, description,
  design_direction}.
- input_mapping maps "spec":"input_data.spec" (delivers the object where
  component-creator reads it), plus site_id/domain TOP-LEVEL (ensure_site_record
  reads input_data.site_id/domain, NOT spec).
- Added a line to the description pinning the function name to brief-explanation
  (belt-and-braces so the in-place UPDATE lands).
- bash -n OK; JSON parses; nesting + mapping asserted in-script.
Expected: LLM gets the real section_type+description → emits
function='brief-explanation' → store_generated_component UPDATE-in-place on
58363894 (status 'regenerated', active_rows=1, schema populated, no <no value>).

DURABLE LESSON (for the debugging guide / tool docs): manually triggering
component-creator via the generic spawn+call path requires the inputs NESTED under
`spec` (mapping "spec"→input_data.spec), mirroring the work-item spec_data→
input_data.spec convention. A flat input_mapping (080 pattern) yields an
empty-context generic generation. Category: workflow-variable-path / input-shape.

---

## 2026-07-01 ~10:10 — Regen 082 FAILED (contract violation); fix = 083 (both top-level + spec)

082 (all inputs nested under spec) FAILED at the call_agent stage:
  contract violation for agent 'component-creator': missing required fields:
  [section_type]. Provided fields: [domain site_id spec].
component-creator NEVER RAN (failure at call_agent extract/validate) → NO new stray.
(The 081 general-hero stray still needs the cleanup SQL from last turn if not done.)

MECHANISM now fully pinned by the two runs:
- call_agent extracts fields via input_mapping, then validates the target's
  input_contract.required against the TOP-LEVEL extracted keys, THEN invokes.
  → 081 (top-level) passed the contract; 082 (nested) failed it.
- The WORKFLOW steps read input_data.spec.* (work-item convention spec_data →
  input_data.spec). → 081 top-level left spec.* empty → generic hero.
So the contract and the workflow read the same field in DIFFERENT places. Neither
pure-top-level nor pure-nested satisfies both. This is a latent DESIGN SMELL in
component-creator (contract vs workflow disagree) — flagged for reconciliation.

FIX 083_regenerate_brief-explanation_vonc.sh (supersedes 082):
- Provides section_type BOTH top-level (contract) AND inside spec (workflow).
- input_mapping sources all one-level (input_data.section_type, input_data.spec) —
  the 082 log proved input_data.spec resolves as a source and arrives intact at
  component-creator, so one-level sources are safe.
- Function name pinned to brief-explanation in the description (in-place UPDATE).
- bash -n OK; JSON parses; structure asserted in-script.
Expected: real brief-explanation generated, function='brief-explanation', store
UPDATE-in-place on 58363894 (status 'regenerated', active_rows=1, schema populated,
has_no_value=f). Then a needs_page index rebuild fills the new fields.

DOCS updated per request:
- 016b debugging guide: new §9 entry "Manually invoking an agent via spawn+call —
  input_mapping must satisfy BOTH the input_contract (top-level) AND the workflow's
  field paths (usually input_data.spec.*)" (symptom a=contract violation,
  b=empty-context generation; diagnose; root cause; fix=both+spec, or use the
  work-item trigger; regeneration-keying note). NOTE: this was added to the working
  copy which ALSO contains the earlier extraction-bug §9 entry + negative-result
  heuristic + v3 changelog (the project 016b re-uploaded was the pre-edit original,
  so the working copy is the cumulative version to apply).
- Per-tool docs created: PLAN_brief-explanation.md + NOTES_brief-explanation.md
  (the 081→082→083 saga with category tags).

PREFERRED LONGER-TERM: drive component-creator via its DESIGNED work-item trigger
(needs_new_component / needs_component_regeneration) where the dispatch loop delivers
spec_data → input_data.spec consistently and this contract/spec split doesn't arise.
The 083 manual path is the pragmatic route for this one-off.

---

## 2026-07-01 ~12:46 — 083 SUCCEEDED (brief-explanation regenerated in place); build-dispatch-loop finding overturns the work-item recommendation

**083 result (correlation ecbca5cb-db7b-4817-a7aa-91b4b3111464):** brief-explanation
(id 58363894-9db9-4d2f-81ac-c47b54d97fc3, created_at UNCHANGED 2026-04-08,
updated_at 2026-07-01 12:46) UPDATED IN PLACE:
  quality 50 → 100 ; template_variable_count 0 → 20 ; schema_field_count 0 → 20 ;
  tmpl_len 8089 → 9760 ; has_no_value t → f ; active_rows = 1 (no duplicate).
Raised: needs_rerender `component_regen_rerender:58363894-...` (status triaged) —
matches the documented `regenerated` status behaviour (003 §348: snapshot to
component_versions, UPDATE in place, one needs_rerender per affected site).
=> The dual-placement fix WORKED: section_type TOP-LEVEL (satisfies the call_agent
   contract check) + full `spec` object (satisfies the workflow's input_data.spec.*
   reads). Function pinned in description → in-place UPDATE keyed on function.

**build-dispatch-loop dump (id 099b51e0) — the important finding.** Its call_handler
input_mapping is:
  {spec: current_item.spec, domain: input_data.domain, issue?: current_item.spec.issue,
   source: current_item.source, site_id: current_item.site_id,
   item_type: current_item.item_type, page_name?: current_item.spec.page_name,
   current_page: current_item.spec, work_item_id: current_item.id,
   component_id?: current_item.spec.component_id, reviewed_brief?: current_item.spec.reviewed_brief}
There is NO `section_type` (not even `section_type?`). So a component-creator
dispatched via the loop gets the whole spec at input_data.spec (workflow OK) but NO
TOP-LEVEL section_type → given the contract check validates required fields at
top-level (proven by 081/082/083), the WORK-ITEM PATH WOULD HIT THE SAME
`missing required fields: [section_type]` error as 082.
=> CORRECTION: my earlier "just use the work-item path, it's the clean route" was
   WRONG. The generic loop can't satisfy component-creator's top-level-required
   contract. component-creator is satisfiable only by a BESPOKE caller (the initial
   builder, or 083), not by the generic dispatch loop as it stands.

**What the loop flattens tells us the design intent.** It flattens only (a) always-
present work-item COLUMNS (site_id, domain, item_type, source, work_item_id) and
(b) a few spec fields — every one OPTIONAL (`?`). Per 002 §414 the loop is generic
("doesn't know what each handler needs"). So the intended contract is: REQUIRED
handler inputs come from input_data.spec; only optional conveniences are flattened.
component-creator requiring section_type TOP-LEVEL is the mismatch.

**FRAMEWORK-BEST FIX (not just what unblocks us).**
- NOT adding section_type to the loop mapping: non-optional would fail loudly for
  other handlers; `section_type?` violates "no ? on required" AND re-introduces
  per-handler knowledge into the generic loop (undoes 002 §414). Worst option.
- NOT blessing 083's duplication as the RULE: it's a bespoke-caller workaround the
  generic loop cannot perform; codifying it would enshrine something the system's
  own dispatch path can't satisfy.
- FIX IN THE CONTRACT VALIDATOR: 003 already says spec fields live at
  input_data.spec.* and reading the nested path is "the one contract-compliant way".
  The validator should honour the same convention — when checking required field X,
  accept X present TOP-LEVEL **or** at input_data.spec.X. Strictly more permissive in
  a principled direction (a genuinely-absent field still fails loudly); makes
  call_agent validation agree with how handlers consume inputs; lets component-creator
  (and any spec-required handler) dispatch through the generic loop with no
  per-handler mapping and no duplication.

**Honest caveats (on the record):**
- Validator-checks-top-level-only is INFERRED from 3 runs + the loop dump; not yet
  seen in call_agent.go. Exact change site needs confirming in code.
- "Work-item path fails for component-creator" is PREDICTED from the mapping, not run.
  Confirm (and get a regression test for the validator fix) by dropping a
  needs_component_regeneration item and watching it hit the same contract error.

**NEXT STEP for brief-explanation (with a hazard flag).** The component now has 20
schema fields but they are EMPTY. The raised needs_rerender only ASSEMBLES stored
rendered HTML (page_rerender does NOT re-render templates), so it will assemble the
OLD/empty rendering. To fill the 20 fields, page-content-writer must render the
component with LLM content for the index — i.e. a page content-rebuild
(needs_content_page per 002 §684; confirm the exact item_type/routing to
page-build-handler → page-content-writer).
HAZARD: a FULL index content-rebuild re-runs the writer over ALL index sections
(interactive-page clobber, 002 §498 / debugging-guide interactive-page entry). The
index carries hero (working) and provocation-card (Mode-B, filled at runtime by the
Path-2 loader). Re-emptying provocation-card at build time is fine (the loader fills
it in-browser) and the js_snippet loader survives, but the hero and any other
working section must not be clobbered. Before triggering: decide whether to render
ONLY brief-explanation onto the index (targeted) or accept a full-index rebuild after
confirming the working sections survive. Do NOT blind-fire a full page rebuild.

---

## 2026-07-01 ~13:10 — Validator gap CONFIRMED in code; patch written; full-index rebuild being prepped

**Confirmed in code (input_mapping.go + call_agent.go).** call_agent.extractDataForAgent
(~L974) resolves input_mapping → inputData, passes that SAME inputData to
ValidateInputContract (~L991), and returns it as the child's input_data (L481).
ValidateInputContract checks only `data[required]` at the TOP LEVEL — never
data["spec"][required]. So a required field delivered under the spec object (where
handlers read it, per 003) is reported missing. This is the exact mechanism behind
082's `missing required fields: [section_type]`, and why build-dispatch-loop (no
top-level section_type) can't dispatch component-creator.

**Patch written: PATCH_validate_input_contract.go.** ValidateInputContract now accepts
a required field if present top-level OR in data["spec"] (the documented
input_data.spec.* container). Strictly more permissive; genuinely-absent fields still
fail loudly; no loop-mapping or agent-def change. Uses logger.Info for the
spec-satisfied note (Debug is not surfaced). This is the framework fix; the 083
top-level duplication becomes unnecessary once applied.

**Full-index rebuild (chosen next move) — safety plan before firing.**
Index sections (6): hero (working), provocation-card (Mode-B shell, runtime-filled by
Path-2 loader), gauntlet-cta (empty shell), brief-explanation (regenerated, 20 empty
fields — TARGET), lobby-grid (Mode-B empty), system-stats (zero visible text).
Clobber mechanism (debugging guide, FIX PENDING): a page rebuild regenerates from
plan_sections and does NOT preserve interactive (inline-<script>) sections ABSENT from
the plan. Index risk = provocation-card / lobby-grid dropped if not in plan_sections;
hero must not degrade; provocation-card's .pc-* shell must survive for the loader.
GATING before trigger: (1) \d page_components + \d site_plans; (2) inventory current
index page_components; (3) compare against the index's plan_sections; (4) snapshot the
index (component_versions + site snapshot, doc 014) for reversibility.
Decision: full rebuild only if plan_sections contains all six (esp. provocation-card
+ lobby-grid); else add them to the plan first, or render only brief-explanation.
Post-rebuild verify: hero correct; provocation-card shell + loader still working;
brief-explanation now filled.

---

## 2026-07-01 ~13:25 — Index page_components inventory (pre-rebuild); 4 sections, not 6

page_id b4d24f8e-... has FOUR page_components (older notes said 6 — gauntlet-cta +
system-stats are NOT present as components; planned-but-never-rendered, revisit):
  1. hero (23f95f00) — template, deployed, not Mode-B, rendered 2263. Working.
  2. provocation-card (6163ff14) — template, deployed, has_inline_script, Mode-B,
     rendered 9994. Runtime-filled shell (.pc-* + Path-2 loader).
  3. brief-explanation (58363894) — AGENT, PENDING, NOT Mode-B, rendered 7869.
     Regen landed at page level (Mode-B now false, render_mode → agent), but
     build_status pending and the 7869 bytes are STALE (old broken rendering). Needs
     the build pipeline to render the new 20-field template. TARGET of the rebuild.
  4. lobby-grid (9304f14d) — template, deployed, has_inline_script, Mode-B,
     rendered 15282. Empty shell.

Schema captured (uploaded): page_components has position, build_status
(default pending), rendered_html, content_data, render_mode via content_components;
lock/deploy triggers (trigger_auto_lock_on_deploy auto-locks on deploy — relevant:
deployed components may be locked). site_plans is current+history; its sections live
in child table site_plan_sections (+ site_plan_pages), ON DELETE CASCADE.

REBUILD SAFETY GATE (before firing the full-index rebuild):
 (1 decisive) page-build-handler / page-content-writer workflow — does it render only
     `pending` page_components (=> only brief-explanation touched, neighbours safe) or
     regenerate the whole page from plan_sections (clobber risk)? Pulling the agent
     defs to settle this rather than guess.
 (2) index's site_plan_sections — what a plan-driven rebuild would render/discard;
     confirm provocation-card + lobby-grid are IN the plan (else discarded per the
     pending clobber).
 (3) backup: CREATE TABLE _vonc_index_pc_backup AS SELECT * ... page_id=index (4 rows)
     — reversibility if a clobber occurs.
Note trigger_auto_lock_on_deploy: deployed page_components may be LOCKED; a rebuild/
re-render may need to respect or clear locks (lock_type permanent|timed|review). Check
locked_at/lock_type on the 4 rows when we inventory — a lock could block re-render of
brief-explanation or protect the neighbours.
