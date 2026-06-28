# Runbook — Phase 2: Provocation JS layer (static test data)

**Created:** 2026-06-25
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

### [ ] Step 1 — Confirm the shell DOM selectors are current
Run the `provocation-card` template query above. Confirm `.pc-headline`,
`.pc-body`, `.pc-eyebrow`, `.pc-stat-value`/`.pc-stat-label`, `.pc-card*`
selectors exist as expected. Adjust the JSON shape / JS selectors if the shell
changed.

### [ ] Step 2 — Hand-write the test JSON and commit it
Create `/data/provocations.json` in the vonc.com repo with 1 today + 4 lobby
entries (sample content is fine). Commit it directly, or via a `git_commit`
action. This is throwaway test data — real content comes from the pipeline later.

How to place it: the file lives at the site's web root under `/data/`. Confirm
how other JSON data files (e.g. `latest-news.json`) are committed — same path
convention and same git mechanism.

### [ ] Step 3 — Write the js_snippets row for provocation-card
Insert a row into `js_snippets`:
- `name`: e.g. `provocation-card-loader`
- `applies_to`: JSONB array including `"provocation-card"`
- `js_content`: an IIFE that:
  1. `fetch('/data/provocations.json')`
  2. on success, find `[data-component="provocation-card"]` and fill the
     selectors from the DOM contract
  3. fail gracefully (leave shell as-is, console.warn) if fetch fails
- `is_active`: true

Check the `js_snippets` schema first:
```sql
\d js_snippets
```
Confirm column names (`applies_to`, `js_content`, `is_active`, `dependencies`,
`semantic_tags`) before inserting.

### [ ] Step 4 — Render and deploy snippets.js
Trigger `site-asset-renderer` for vonc.com so `render_js_snippets_for_site`
picks up the new snippet (its `applies_to` matches the index's
provocation-card component) and commits `/assets/js/snippets.js`.

Spawn it via a work item (confirm the item_type/handler `site-asset-renderer`
expects):
```sql
-- Verify the trigger mechanism first
SELECT type, default_config->'workflow'->'steps' AS steps
FROM agent_definitions WHERE type = 'site-asset-renderer';

-- Then queue (item_type TBD from how it's normally triggered — likely a
-- direct spawn or an 'undeployed_asset' / asset-render work item)
```

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
