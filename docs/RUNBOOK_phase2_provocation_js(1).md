# Runbook — Phase 2: Provocation JS layer (static test data)

**Created:** 2026-06-25  •  **Updated:** 2026-06-25 ~18:00
**Parent plan:** PLAN_spark_provocation_pipeline.md
**Parent runbook:** RUNBOOK_vonc_migrations.md (Track 2)

---

## What Phase 2 is

Prove the runtime JS layer works — JS reading a static JSON file and populating
the empty shell components — BEFORE building the data pipeline that generates
the JSON. This de-risks the hardest-to-test part (DOM population) using
hand-written data. If the shells fill correctly from a static
`/data/provocations.json`, the entire data pipeline just needs to produce that
JSON shape.

**Scope of Phase 2:** provocation-card on the index page only, first. Once that
works, extend to lobby-grid and brief-explanation.

---

## MECHANISM CONFIRMED (2026-06-25 ~18:00)

All of this is verified from source (render_js_snippets_for_site_action.go,
render_news_section_action.go, agent_definitions):

- **site-asset-renderer** agent (type `site-asset-renderer`, active): input
  `site_id` (required) + `domain` (optional). Runs render_js_snippets_for_site
  → git_commit (writes `assets/js/snippets.js`). 120s timeout. This is what we
  spawn in Step 4.
- **render_js_snippets_for_site** auto-selects js_snippets WHERE is_active AND
  `applies_to` overlaps the site's component functions. The index page uses the
  `provocation-card` component, so a snippet with `applies_to ["provocation-card"]`
  is picked up automatically — no per-site wiring.
- **git_commit** writes a `files` map keyed by repo path. `data/provocations.json`
  and `assets/js/snippets.js` both reach the repo this way.
- **The shell currently renders `<no value>`** in each slot (empty content_data).
  The loader overwrites these regardless, so that's harmless.
- **The template has its own card-activation IIFE** (hover/focus). Our loader is
  a separate snippet and only fills content — the two coexist.

## READY-MADE ARTIFACTS (in outputs)

- `provocations.sample.json` — validated sample data (1 today + 4 lobby).
- `provocation_card_loader.js` — the loader IIFE (node --check passed).
- `phase2_step3_insert_snippet.sql` — dollar-quoted INSERT for the js_snippets row.


**Out of scope (later phases):** the scraping/generation pipeline, the scheduled
task, the provocations-index page.

---

## The shell's DOM contract (what the JS targets)

The `provocation-card` shell (already deployed, empty) has this structure.
The JS fills these selectors — it does NOT change the structure or CSS.

Main provocation block:
- `.pc-eyebrow` — small label above headline (e.g. "TODAY'S PROVOCATION")
- `.pc-headline` (id `pc-headline`) — the provocation claim (may contain `<em>`)
- `.pc-body` — the AI's take / supporting text
- `.pc-btn-primary` — `href` + visible label (e.g. "File Your Position")
- `.pc-btn-secondary` — `href` + label (e.g. "See All Provocations")

Stat strip (3 stats):
- `.pc-stat-value` × 3 — the numbers (positions filed, time left, etc.)
- `.pc-stat-label` × 3 — the labels under each number

Mini-lobby ("today's other provocations", 4 cards):
- `.pc-card` × 4, each containing:
  - `.pc-card-icon` — emoji/symbol
  - `.pc-card-title` — short provocation title
  - `.pc-card-desc` — one-line description
  - the card itself may need an `href` (wrap in `<a>` or set a data attr)

**Action:** before writing the JS, re-fetch the live `provocation-card`
template to confirm these selectors are current (the shell may have been
re-rendered):
```sql
SELECT html_template FROM content_components
WHERE function = 'provocation-card' AND is_active = true AND forked_from IS NULL;
```

---

## Proposed JSON shape (`/data/provocations.json`)

Driven by the DOM contract above. Draft:
```json
{
  "generated_at": "2026-06-25T06:00:00Z",
  "today": {
    "eyebrow": "TODAY'S PROVOCATION",
    "headline": "AI will <em>never</em> be funny on purpose.",
    "body": "The AI's take: humour needs a victim and a risk...",
    "primary_cta": { "label": "File Your Position", "url": "/tools/gauntlet/index.html" },
    "secondary_cta": { "label": "See All Provocations", "url": "/provocations/index.html" },
    "stats": [
      { "value": "1,284", "label": "Positions Filed" },
      { "value": "3h 12m", "label": "Until Close" },
      { "value": "62%", "label": "Disagree" }
    ]
  },
  "lobby": [
    { "icon": "🔥", "title": "...", "desc": "...", "url": "/provocations/..." },
    { "icon": "⚡", "title": "...", "desc": "...", "url": "/provocations/..." },
    { "icon": "🧠", "title": "...", "desc": "...", "url": "/provocations/..." },
    { "icon": "🎯", "title": "...", "desc": "...", "url": "/provocations/..." }
  ]
}
```

The data pipeline (Phase 3) will produce exactly this shape. Finalise it here
in Phase 2 against the real DOM so Phase 3 has a fixed target.

---

## Steps

### [x] Step 1 — Confirm the shell DOM selectors are current — DONE
CONFIRMED from the live template dump: `.pc-eyebrow`, `.pc-headline#pc-headline`
(supports `<em>`), `.pc-body`, `.pc-btn-primary` (a[href]+SVG),
`.pc-btn-secondary` (a[href]), `.pc-stat-value`×3 + `.pc-stat-label`×3,
`.pc-card`×4 (each `.pc-card-icon`/`.pc-card-title`/`.pc-card-desc`). The loader
JS and sample JSON are written against exactly these selectors.

### [ ] Step 2 — Commit the test JSON to the repo
`provocations.sample.json` is ready (validated). Commit it to the vonc.com repo
at `data/provocations.json` (repo-relative path, same convention as
`data/latest-news.json`). Either:
- commit it directly to the repo, or
- run a `git_commit` action with files map `{"data/provocations.json": <contents>}`,
  domain `vonc.com`.
This is throwaway test data — Phase 3's render action will generate the real file.

### [ ] Step 3 — Insert the js_snippets row
`phase2_step3_insert_snippet.sql` is ready — a dollar-quoted INSERT for the
`provocation-card-loader` snippet (applies_to `["provocation-card"]`, the
validated loader IIFE as js_content, is_active true).

FIRST run `\d js_snippets` to confirm column names and any NOT NULL columns
without defaults. The render action reads name/description/js_content/applies_to/
is_active — confirm those exist as expected, then run the insert.

### [ ] Step 4 — Render and deploy snippets.js
Spawn `site-asset-renderer` for vonc.com. It runs render_js_snippets_for_site
(picks up our snippet automatically via applies_to overlap) → git_commit writes
`assets/js/snippets.js`. Input needed: `site_id`
(`9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`); domain optional.

Confirm HOW agents are spawned/triggered in this system before running — the
other pipelines use `spawn_agent`/`call_agent` from a parent workflow, or a
scheduled trigger. For a one-off manual run we need the correct invocation
(likely a work item the dispatcher routes to site-asset-renderer, or a direct
spawn message on its process topic). Determine the manual-trigger path, then
fire it with the vonc site_id.

### [ ] Step 5 — Verify in the browser
Load vonc.com/index.html. The provocation-card section should now show the
test provocation, stats, and 4 lobby cards instead of empty outlines.
Check the browser console for fetch errors.

### [ ] Step 6 — Extend to lobby-grid and brief-explanation
Once provocation-card works, repeat Steps 1–5 for the other two shells:
- `lobby-grid` — likely reads the same `lobby` array (or a larger room list)
- `brief-explanation` — static-ish; may just need its copy filled from JSON
  or may be genuinely static (confirm what it's meant to show)

Each gets its own js_snippets row (or one combined snippet with `applies_to`
listing all three functions).

---

## Done when

- provocation-card, lobby-grid, brief-explanation all populate from
  `/data/provocations.json` in the browser.
- The JSON shape is finalised and documented (Phase 3 will generate it).
- snippets.js is committed and loads without errors.

Then → Phase 3 (build the data pipeline that generates the JSON), per
PLAN_spark_provocation_pipeline.md.

---

## Constraints

- The JS must fail gracefully — if `/data/provocations.json` is missing or
  malformed, the shell stays as-is, no broken page.
- Scope JS to `[data-component="..."]` containers — don't touch other sections.
- No external CDN imports in the snippet (project rule).
- IIFE, no global pollution (matches existing js_snippets pattern).
- A `needs_page` rebuild of index re-renders the shell (empty) but does NOT
  remove the JS — snippets.js is a separate asset. So index rebuilds are safe.
- `logger.Info` not `logger.Debug` for any Go action involved.
