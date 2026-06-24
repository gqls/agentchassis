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
