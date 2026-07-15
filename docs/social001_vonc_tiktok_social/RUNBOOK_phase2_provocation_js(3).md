# Runbook — Phase 2: Provocation JS layer (static test data)

**Created:** 2026-06-25  •  **Updated:** 2026-06-25 ~18:00
**Parent plan:** PLAN_spark_provocation_pipeline.md
**Parent runbook:** RUNBOOK_vonc_migrations.md (Track 2)


---

## THE PROBLEM (read this first if you're new to this document)

Three components on vonc.com — **provocation-card**, **lobby-grid**, and
**brief-explanation** — render as empty card outlines (blank inputs, unlabelled
buttons, empty stat slots). They were built as *static shells*: HTML structure
with no `{{.field}}` template slots and no content. They are meant to be filled
at runtime by JavaScript that reads a daily-regenerated JSON data file
(`/data/provocations.json`) — the Spark product's "daily provocation" mechanism.

But nothing fills them, because **three things were never built**:

1. **The data feed** — no pipeline produces `provocations.json`. (The news feed
   pipeline does exactly this for news; the provocation version doesn't exist.)
2. **The loader JS** — no `js_snippets` row existed to fetch the JSON and fill
   the shells. (We have now inserted `provocation-card-loader` by hand as a
   proof; the framework should generate it.)
3. **The deploy trigger** — even with a snippet in the DB, nothing re-ran
   `site-asset-renderer` to rebuild `assets/js/snippets.js` for the site.

**The database is the source of truth, not the git repo.** So the correct fix
is NOT to paste JSON/JS into the repo — it is to make the pipeline PRODUCE the
loader snippet and the data feed from the DB and render them out like any other
asset, ideally generated dynamically when the component is planned. The manual
steps below are a one-time proof that the loader + JSON shape are correct; the
real deliverable is the framework fix (see "FRAMEWORK FIX" section near the end).

---

## STATUS AT A GLANCE

Manual proof track (prove the JS-fills-shell mechanism works end-to-end):

<ul>
<li><input type="checkbox" checked disabled> <b>P2-1</b> Confirm shell DOM selectors (done — selectors verified against live template)</li>
<li><input type="checkbox" checked disabled> <b>P2-2a</b> Write sample <code>provocations.json</code> (done — validated)</li>
<li><input type="checkbox" checked disabled> <b>P2-2b</b> Write loader JS (done — node --check passed)</li>
<li><input type="checkbox" checked disabled> <b>P2-3</b> Insert <code>provocation-card-loader</code> into js_snippets (done — in DB, 4879 bytes, is_active)</li>
<li><input type="checkbox" disabled> <b>P2-2c</b> Get <code>provocations.json</code> onto the site the proper way (see "DATA DEPLOY" — do NOT hand-commit)</li>
<li><input type="checkbox" disabled> <b>P2-4</b> Trigger <code>site-asset-renderer</code> for vonc → rebuild snippets.js</li>
<li><input type="checkbox" disabled> <b>P2-5</b> Verify the provocation-card fills in the browser</li>
<li><input type="checkbox" disabled> <b>P2-6</b> Extend loader to lobby-grid + brief-explanation</li>
</ul>

Framework fix track (the actual deliverable — diagnose then build):

<ul>
<li><input type="checkbox" disabled> <b>FX-1</b> Run Diagnostic Set A — is site-asset-renderer wired into the build, or ad hoc?</li>
<li><input type="checkbox" disabled> <b>FX-2</b> Run Diagnostic Set B — is loader-snippet-per-component an existing pattern? How does 'latest-news' deliver its JS?</li>
<li><input type="checkbox" disabled> <b>FX-3</b> Read 022_dynamic_applications.md + the site-planner definition — how should a data-driven component declare its data dependency?</li>
<li><input type="checkbox" disabled> <b>FX-4</b> Gap 1 fix: build the data feed (render_provocations_section + provocation-orchestrator + scheduled task) — mirrors news pipeline</li>
<li><input type="checkbox" disabled> <b>FX-5</b> Gap 2 fix: make the pipeline auto-create the loader snippet for data-driven components</li>
<li><input type="checkbox" disabled> <b>FX-6</b> Gap 3 fix: auto-trigger site-asset-renderer when a site's js_snippets set changes</li>
</ul>


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

### P2-1 — Confirm the shell DOM selectors are current  ✅ DONE
CONFIRMED from the live template dump: `.pc-eyebrow`, `.pc-headline#pc-headline`
(supports `<em>`), `.pc-body`, `.pc-btn-primary` (a[href]+SVG),
`.pc-btn-secondary` (a[href]), `.pc-stat-value`×3 + `.pc-stat-label`×3,
`.pc-card`×4 (each `.pc-card-icon`/`.pc-card-title`/`.pc-card-desc`). The loader
JS and sample JSON are written against exactly these selectors.

### P2-2c — Get the test JSON onto the site (DATA DEPLOY — see below, do not hand-commit)
`provocations.sample.json` is ready (validated). Commit it to the vonc.com repo
at `data/provocations.json` (repo-relative path, same convention as
`data/latest-news.json`). Either:
- commit it directly to the repo, or
- run a `git_commit` action with files map `{"data/provocations.json": <contents>}`,
  domain `vonc.com`.
This is throwaway test data — Phase 3's render action will generate the real file.

### P2-3 — Insert the js_snippets row  ✅ DONE (snippet now in DB)
`phase2_step3_insert_snippet.sql` is ready — a dollar-quoted INSERT for the
`provocation-card-loader` snippet (applies_to `["provocation-card"]`, the
validated loader IIFE as js_content, is_active true).

FIRST run `\d js_snippets` to confirm column names and any NOT NULL columns
without defaults. The render action reads name/description/js_content/applies_to/
is_active — confirm those exist as expected, then run the insert.

### P2-4 — Render and deploy snippets.js (trigger site-asset-renderer)
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

### P2-5 — Verify in the browser
Load vonc.com/index.html. The provocation-card section should now show the
test provocation, stats, and 4 lobby cards instead of empty outlines.
Check the browser console for fetch errors.

### P2-6 — Extend to lobby-grid and brief-explanation
Once provocation-card works, repeat Steps 1–5 for the other two shells:
- `lobby-grid` — likely reads the same `lobby` array (or a larger room list)
- `brief-explanation` — static-ish; may just need its copy filled from JSON
  or may be genuinely static (confirm what it's meant to show)

Each gets its own js_snippets row (or one combined snippet with `applies_to`
listing all three functions).

---

## DATA DEPLOY — the proper way to get provocations.json onto the site

**Do not hand-commit `provocations.json` to the repo.** The database is the
source of truth; data files are rendered from the DB and committed by an action,
exactly like the news feed does it.

The news pipeline is the proven pattern (render_news_section_action.go):
- A render action reads rows from the DB (content_feed_items for news), shapes
  them into a Go struct, marshals to JSON, and returns a `files` map:
  `{"data/latest-news.json": "<json>"}`.
- A `git_commit` step in the orchestrator writes that map to the repo.
- No human touches the file; it regenerates every pipeline cycle.

So the proper provocation equivalent (this is FX-4, Phase 3):
1. **`render_provocations_section`** Go action — mirror of render_news_section.
   Reads provocation rows from the DB (a provocations table, or content_feed_items
   tagged as provocations), shapes them to the `provocationJSONOutput` struct that
   matches the loader's DOM contract, returns
   `{"data/provocations.json": "<json>"}`.
2. **`provocation-orchestrator`** agent — clone of content-feed-orchestrator:
   seed sources → dispatch ingesters → spawn provocation-generator (LLM:
   topics→provocations) → render_provocations_section → git_commit.
3. **scheduled task** (scheduled_tasks row, daily) — clone of
   `content-feed-refresh`. Note the real column is `name` (not `task_name`).

**For the P2 manual proof ONLY** (until FX-4 exists), to avoid hand-editing the
repo, the cleanest interim is a tiny throwaway render action OR a one-off
git_commit invocation with the sample JSON as the files map — but recognise this
is scaffolding. The moment render_provocations_section exists, the data file is
produced from the DB and this manual step is retired. If a quick visual proof is
needed now, commit the sample once, verify the loader works, then delete it and
let the real pipeline own the file.

---

## FRAMEWORK FIX — the actual deliverable (three gaps)

The blank cards are not a one-site bug; they expose three framework gaps. Each
has a distinct fix. Diagnose (FX-1..3) before building (FX-4..6).

### DIAGNOSIS CONFIRMED (2026-06-25 ~19:00) — two JS delivery paths

There are TWO separate JS delivery mechanisms; the data-driven shells fell
between them:

- **Path 1 — per-component inline script → `/tools/assets/{function}.js`.**
  A component's inline `<script>` is extracted to `content_components.js_content`
  and deployed by `rerender_single_page_action.collectJSAssets` on every page
  rerender. AUTOMATIC. This is how the working interactive components (gauntlet-
  interface etc.) get their JS. provocation-card has an inline script too, but it
  only does card hover effects — it does NOT fetch data.
- **Path 2 — library `js_snippets` → `/assets/js/snippets.js`.**
  `render_js_snippets_for_site` bundles snippets matching the site's components;
  `site-asset-renderer` commits the bundle. This is the path our loader uses.
  **It is NOT part of the normal page build/rerender flow.**

### Gap 3 — site-asset-renderer not re-triggered (FX-6) [SMALLEST, DO FIRST]
CONFIRMED: only `rerender-site` and `webdesign-agent` reference site-asset-
renderer / render_js_snippets. So `/assets/js/snippets.js` is built during
INITIAL DESIGN and on a full site rerender — but nothing re-runs it when a
js_snippets row is added later. We added provocation-card-loader after the
build, so the bundle was never regenerated. **This is the direct cause of the JS
not reaching the site.** The page rerenders we ran touch Path 1, never Path 2.

Fix (preferred): add a check to **design-discovery-agent** (it already scans for
"undeployed assets" and "stale header/footer renders" — same shape). The check:
a site has an active js_snippet whose applies_to matches one of the site's
component functions, but the deployed snippets.js is missing or predates the
snippet → spawn site-asset-renderer for that site. Confirm design-discovery-
agent's structure (query in the notes) before writing it.

For the immediate vonc unblock: spawn site-asset-renderer for vonc manually
(P2-4) — that regenerates snippets.js with our loader included.

### Gap 2 — loader snippet + data-feed not auto-created (FX-5) [MEDIUM]
CONFIRMED: component-creator has Tier A/B/C/D + "renderer" fields, but NO tier
for "this component is populated from an external JSON feed at runtime and needs
(i) a loader snippet and (ii) a data file produced by a pipeline." So creating a
shell like provocation-card emits no loader and registers no data need. news's
'latest-news' has BOTH a loader snippet and render_news_section, but both were
hand-built — there's no generic data-driven-component abstraction.
Decision input still needed: Diagnostic B3/B4 (does latest-news have a loader
snippet?). If yes, copy that pattern; the fix is a new component flag/tier that,
at creation/plan time, seeds the loader js_snippet and registers the feed.

### Gap 1 — no data feed (FX-4) [LARGEST, WELL-TEMPLATED]
CONFIRMED: no Spark content_sources/feed items/agents/scheduled task. Fix = clone
the news pipeline (render_provocations_section + provocation-orchestrator +
scheduled task `name`='provocation-refresh'), per PLAN_spark_provocation_pipeline.md.

### On-doctrine check (022_dynamic_applications.md)
Tier 1 (now) explicitly includes "Dynamic content injection (RSS feeds,
API-fetched data rendered client-side)… still static HTML… the dynamic part runs
in the browser." Option 2 is the documented approach, not a workaround.
Principle 4: backend complexity lives in agents, not generated code — so the
provocation generation logic belongs in an agent, the site just fetches JSON.

### How a component should DECLARE it is data-driven (the root design point)
provocation-card has `input_schema = {}` — it declares nothing, so the pipeline
can't know it needs a feed + loader. The durable fix is a schema convention: a
data-driven component declares its runtime data dependency (e.g. a
`data_source` block naming the feed + the loader), so that at plan time the
pipeline provisions the feed, the loader snippet, and the render/commit/schedule
wiring automatically. 022_dynamic_applications.md (FX-3) should be the spec this
aligns to; if it doesn't cover runtime-data components, this convention is a
proposed addition to that doc.

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
