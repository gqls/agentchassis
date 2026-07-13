# Runbook — Phase 2: Provocation JS layer (static test data)

**Created:** 2026-06-25  •  **Updated:** 2026-07-04 (App. I gate results — zero planned sections)
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

### >>> YOU ARE HERE (2026-06-29 ~19:00)
**provocation-card is LIVE and fully working** (Path-2 loader: snippets.js commit
eb7f2ac + complete provocations.json). The remaining index shells are analysed
and triaged:
- **provocation-card** — done (Path-2). Durable: migrate to Path 1 after the
  extraction bug. Per-tool docs written (PLAN_/NOTES_provocation-card.md).
- **brief-explanation** — STATIC content → fix is REGENERATION with a real schema
  (content-writer fills at build), NOT a loader. Not yet done.
- **lobby-grid** — DYNAMIC (arena lobby / today's provocations grid) → loader
  candidate like provocation-card, but BLOCKED on two decisions (overlap with
  provocation-card's mini-lobby; v1 "rooms" semantics) + the shared extraction bug
  + Phase 3 data. Not yet done.

**NEXT: the extraction bug** (now the shared blocker). provocation-card AND
lobby-grid both have an inline `<script>` in html_template (`has_inline_script=TRUE`)
but empty `js_content`, so `/tools/assets/{function}.js` is never produced — their
built-in interactivity isn't deploying, and it blocks the Path-1 loader migration.
Leading hypothesis: the bug is in the component STORE path (separateInlineJS either
bailing on a malformed/truncated stored script, or never invoked on the creation
path that produced these Mode-B shells), not in the rerender-time reader. Needs:
the store action + separateInlineJS code, plus the SQL to confirm whether the
stored `<script>` is intact or truncated (has_script_close) and a healthy
comparison (gauntlet-interface / latest-news). See "EXTRACTION BUG" section below.

**Doc-storage route:** assessed (TOOL_DOCS_convention.md Appendix A) — files now
(library repo), hybrid later (NOTES→DB when agents write them; PLAN stays in git),
not a DB-first build.

**Manual proof track** (prove JS-fills-shell end-to-end on vonc):

- [x] **P2-1** Confirm shell DOM selectors — verified against live template
- [x] **P2-2a** Write sample `provocations.json` — validated
- [x] **P2-2b** Write loader JS — `node --check` passed
- [x] **P2-3** Insert `provocation-card-loader` into js_snippets — in DB, 4879 bytes, is_active
- [x] **P2-4** Triggered site-asset-renderer — ran clean (commit eb7f2ac), loader bundled
- [x] **P2-2c** provocations.json present in /data/ (interim hand-commit; proper pipeline = Phase 3)
- [x] **P2-5** Card fills COMPLETELY (headline/take/stats/lobby icons+titles+descriptions/CTAs all ✓). The earlier blank titles were stub test data; full JSON fixed it. Phase 2 mechanism proven.
- [ ] **P2-6** Extend to lobby-grid + brief-explanation

**Path decision** (resolve before the framework build, so we build the right thing):

- [x] **PD-1** ANSWERED — latest-news template ends `<script src="/tools/assets/latest-news.js">`; fetch lives in the component's extracted JS = **Path 1**.
- [x] **PD-2** ANSWERED — provocation-card has_inline_script=TRUE but js_content is EMPTY. The inline script was never extracted → /tools/assets/provocation-card.js isn't produced. **Separate bug** (see below).
- [x] **PD-3** DECIDED — **Path 1** (loader in component js_content → /tools/assets/{function}.js, auto-deploys on rerender), matching how news works. Path-2 snippet kept as working interim until the extraction bug is fixed and the loader is migrated.

**Framework fix track** (the actual deliverable — diagnosis done, build pending):

- [x] **FX-1** Diagnostic Set A — site-asset-renderer is NOT wired into the build (only webdesign-agent + rerender-site reference it). Gap 3 confirmed.
- [x] **FX-2** Diagnostic Set B — every existing js_snippet is a *generic* behaviour; `news-date-formatter` is a formatter not a fetcher. Raises the Path question above.
- [x] **FX-3** Read 022_dynamic_applications.md — Tier 1 doctrine = client-side data injection on static HTML; backend logic lives in agents. Option 2 is on-doctrine.
- [ ] **FX-4** Gap 1: build the data feed (render_provocations_section + provocation-orchestrator + scheduled task) — mirrors news pipeline
- [ ] **FX-5** Gap 2: make the pipeline auto-provision a data-driven component's loader + data feed at plan/create time
- [ ] **FX-6** Gap 3: PD-3 picked Path 1 → page rerender already ships component JS, so this drops OFF the critical path for provocation-card. Remains a latent issue for genuinely-generic Path-2 snippets; lower priority.
- [ ] **BUG-EXTRACT** provocation-card inline `<script>` not extracted to js_content (separateInlineJS / possibly-truncated stored template). Blocks Path 1 here — investigate before migrating the loader.


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
Agents are spawned via `spawn_agent`/`call_agent` from a parent workflow (only
long-running adapters are different). For a one-off manual run, post to the
generic entry point (`system.agent.generic.requests` with
`config.agent_type=site-asset-renderer`), exactly like the classifier smoke test.

A ready script is in outputs: **`trigger-asset-renderer-vonc.sh`** (site_id and
domain are vonc's; bash syntax checked). It posts the spawn message and prints
the log/orchestration-state commands to watch. site-asset-renderer runs
`render_js_snippets_for_site` (picks up `provocation-card-loader` automatically
via applies_to overlap) → `git_commit` writes `assets/js/snippets.js`.

The message body it sends:
```
{"action":"orchestrate","config":{"agent_type":"site-asset-renderer"},"input_data":{"site_id":"9ec3b9ee-5b08-461b-b4f8-9e1e03579c74","domain":"vonc.com"}}
```

After it runs, confirm the loader is in the bundle (look for `provocation-card-loader`
in the `snippet_names` log line), then once git→S3 deploys:
`curl -s https://vonc.com/assets/js/snippets.js | head -40`.

NOTE: triggering this by hand is the interim. The durable Gap 3 fix (FX-6) is to
auto-run it when snippets change — but only if PD-3 selects Path 2. If PD-3
selects Path 1, this manual trigger and FX-6 are both unnecessary for this case,
because the loader rides in the component's own JS asset on page rerender.

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

## PATH DECISION — where does the fetch logic belong? (resolve first)

The js_snippets inventory shows **every** pre-existing snippet is a *generic,
cross-component behaviour* (accordion, scroll-reveal, mobile-menu-toggle,
form-validation, counter-animate…), each ≤815 bytes and applying to 2–3 component
types. The news entry, `news-date-formatter` (506 bytes, applies to latest-news
**and** news-listing), is by its name a date helper — **not** a data fetcher.

Our `provocation-card-loader` is the outlier: specific to one component and 4879
bytes, and it is a fetch-and-render loader. That suggests it may be in the wrong
home. Two delivery paths exist in this system:

- **Path 1 — inline `<script>` in the component template.** Extracted to
  `content_components.js_content` and deployed as `/tools/assets/{function}.js`
  by `rerender_single_page_action.collectJSAssets` on **every page rerender**
  (automatic; no separate trigger). This is how all the working interactive
  components ship their JS.
- **Path 2 — a `js_snippets` row.** Bundled into `/assets/js/snippets.js` by
  `render_js_snippets_for_site`, committed by `site-asset-renderer` — which is
  **not** in the normal build/rerender flow (Gap 3). This is where we put the
  loader.

If news fetches its JSON via the `latest-news` component's own inline script
(Path 1), then the architecturally-consistent home for the provocation fetch is
the **provocation-card component's inline `<script>`** (it already has one for
card hover) — which removes the Gap 3 dependency entirely.

**Confirm before deciding:**

- **PD-1** — does `latest-news` fetch in an inline script?
  ```sql
  SELECT html_template FROM content_components
  WHERE function='latest-news' AND is_active=true AND forked_from IS NULL;
  ```
  Look for `fetch(` / `/data/latest-news.json` inside a `<script>` block.

- **PD-2** — is provocation-card's existing inline script already extracted/deployed?
  ```sql
  SELECT function, LENGTH(js_content) AS js_len,
         (html_template ILIKE '%<script%') AS has_inline_script
  FROM content_components
  WHERE function='provocation-card' AND is_active=true AND forked_from IS NULL;
  ```
  `js_len > 0` means `/tools/assets/provocation-card.js` already deploys on rerender.

- **PD-3** — decide:
  - **Path 1** (recommended if PD-1 confirms news does this): move the fetch-and-fill
    logic into the provocation-card component's inline `<script>` (merge with the
    card-activation IIFE), retire the js_snippet. Loader deploys on page rerender.
    Gap 3 / FX-6 unnecessary for this case.
  - **Path 2** (keep what we inserted): leave the js_snippet, and do FX-6 so it
    auto-deploys. Justified only if there's a reason to keep component-specific
    loaders out of the component template.

  Record the decision and reason in the running notes.

## EXTRACTION BUG — inline `<script>` not reaching js_content (NEXT TASK)

**One-paragraph explanation.** When a component template contains an inline
`<script>`, the pipeline is meant to extract that script into
`content_components.js_content` (the step `separateInlineJS`), so that on page
rerender `collectJSAssets` deploys it as `/tools/assets/{function}.js`. For
**provocation-card** and **lobby-grid**, the inline `<script>` is present in
`html_template` (`has_inline_script=TRUE`) but `js_content` is **empty** — the
extraction never landed, the per-component JS asset is never produced, and their
built-in interactivity (provocation-card's card-activation; lobby-grid's
hover/focus/entrance animation) isn't deploying. This also **blocks the Path-1
loader migration** (PD-3), since Path 1 depends on the loader living in
`js_content`.

**Where the bug likely lies.** Leading hypothesis: the **component store/save
path**, not the rerender-time reader. Either (a) `separateInlineJS` ran on a
malformed script and bailed — the earlier template dump showed provocation-card's
inline script truncated mid-function with a stray backslash, which would break a
regex/brace-matching extractor — or (b) these components were created/stored via a
path that wrote `html_template` without ever invoking extraction, consistent with
their shared Mode-B brokenness (bare `<no value>`, empty schema), which already
pointed at an abnormal creation/save path. Not committing to either until the code
and the stored-script state are seen. NOT the `collectJSAssets` reader (that just
reads `js_content`, which is correctly empty given nothing populated it).

**FINDINGS (2026-06-29) — root-caused; the code is fine, the data isn't.**
- `separateInlineJS` (store action, line 105, called unconditionally) and
  `collectJSAssets` (rerender reader) are both CORRECT. Not the bug.
- The has_script_close SQL split is decisive:
  - **lobby-grid**: `has_script_close=15644` (`</script>` present, script INTACT),
    yet the template still ends with the RAW IIFE (not a `<script src>` ref) and
    js_content is empty → **extraction never ran** on it (a working extraction
    would have replaced the raw script). Not a malformed-script bail.
  - **provocation-card**: `has_script_close=0` (no `</script>`), tail ends
    mid-function with a stray backslash → the stored script is genuinely
    **truncated in the DB** (separate corruption).
- Root cause: these Mode-B components were stored via a path that did NOT run
  separateInlineJS — most likely they **predate** its addition (or a seed/bulk
  path bypassed it) and were never regenerated. provocation-card's source was
  ADDITIONALLY truncated at generation, uncaught because validation checks
  unclosed `<style>` but not unclosed `<script>`.
- Structural gaps to fix so it can't recur: (1) add a `<script>` open/close
  balance check to store-path validation; (2) make separateInlineJS WARN on an
  unterminated `<script>` instead of silently returning empty.
- FIX direction (converges with Path-1): REGENERATE the three components through
  the current store path with a COMPLETE inline `<script>` (interactivity + data
  fetch-fill); separateInlineJS extracts → /tools/assets/{function}.js deploys on
  rerender; retire the Path-2 snippet. provocation-card MUST be regenerated
  (truncated source); lobby-grid is also Mode-B so regenerate it too for
  consistency.

**CONFIRMATION PENDING — healthy comparison (run this):**
```sql
SELECT function,
       LENGTH(COALESCE(js_content,'')) AS js_len,
       (html_template LIKE '%<script src=%') AS has_src_ref,
       (html_template LIKE '%<script>%')     AS has_raw_inline
FROM content_components
WHERE function IN ('gauntlet-interface','tool-archetype-taster-quiz','latest-news')
  AND is_active = true AND forked_from IS NULL
ORDER BY function;
```
Expect: js_len>0, has_src_ref=TRUE, has_raw_inline=FALSE for healthy components
(confirms the current path extracts correctly and the three shells are the
exception). latest-news already shows the src-ref pattern.

---

**Information needed to start (original — now mostly answered):**
1. **The store action that calls `separateInlineJS`** — likely
   `store_generated_component_action.go` or a content helper. Need to see whether
   extraction runs on INSERT, UPDATE, both, or only a specific creation path.
2. **The `separateInlineJS` implementation** — regex vs parser; what it does on a
   malformed/unterminated `<script>` (skip silently / error / partial).
3. **Is the stored script actually truncated, or was that display-only?** Run:
   ```sql
   SELECT function,
          RIGHT(html_template, 400) AS template_tail,
          POSITION('</script>' IN html_template) AS has_script_close
   FROM content_components
   WHERE function IN ('provocation-card','lobby-grid')
     AND is_active = true AND forked_from IS NULL;
   ```
   `has_script_close = 0` → stored `<script>` is genuinely unterminated →
   extraction skips it → root cause is upstream (whatever stored a truncated
   template). Non-zero → script intact → bug is in the extraction call/condition.
4. **A healthy comparison** (a component whose `js_content` IS populated):
   ```sql
   SELECT function, LENGTH(js_content) AS js_len,
          POSITION('</script>' IN html_template) AS has_script_close
   FROM content_components
   WHERE function IN ('gauntlet-interface','tool-archetype-taster-quiz','latest-news')
     AND is_active = true AND forked_from IS NULL;
   ```
   Contrasts a working extraction against the broken ones.

**Scope note:** this is `js-not-extracted` (taxonomy) and affects both
provocation-card and lobby-grid; root-causing it likely also restores the missing
card-hover/entrance interactivity. If the cause is the malformed stored template,
it overlaps `mode-b-template` — i.e. one creation/save bug with several surface
symptoms.

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

---

## Appendix E — Regenerating a component through component-creator (input-shape + contract finding)

*Added 2026-07-01. Relevant here because the durable Path-1 fix for provocation-card
and lobby-grid (the inline-`<script>` extraction bug) requires regenerating those
components through component-creator — the same path proven below on brief-explanation.*

### What component-creator does on regeneration
Looks up the existing row by the LLM's EMITTED `function` (forked_from IS NULL); if
found → `isRegeneration` → UPDATE IN PLACE (component_id preserved, FK-safe), snapshots
the old row to `component_versions` (MAX(version)+1), sets new html_template /
input_schema / js_content / render_mode / is_active=true, and raises ONE
`needs_rerender` per affected site. Status `regenerated` (003 §348).

### The invocation trap (proven on brief-explanation, 2026-07-01)
component-creator's WORKFLOW reads its inputs from `input_data.spec.*` (the handler
convention, 003 §984–995). But `call_agent` validates the target's
`input_contract.required` against the **top-level** extracted fields, *before*
invoking. component-creator requires `section_type`. Result:
- All fields TOP-LEVEL (081) → contract passes, but the workflow reads empty
  `input_data.spec.*` → generates a generic `general-hero` (a STRAY; store INSERTs
  because the emitted function doesn't match). 
- All fields nested under `spec` (082) → workflow would be happy, but the top-level
  contract check fails: `missing required fields: [section_type]. Provided fields:
  [domain site_id spec]` → agent never runs.
- **Working manual shape (083):** provide `section_type` BOTH top-level (for the
  contract) AND inside the full `spec` object (for the workflow); pin the function
  name in the description so the in-place UPDATE lands. Verified: brief-explanation
  58363894 updated in place, quality 100, 20 schema fields, has_no_value=f,
  active_rows=1, needs_rerender raised.

### Why the work-item path does NOT currently work for component-creator
`build-dispatch-loop`'s `call_handler` input_mapping (agent_definitions id 099b51e0)
flattens only work-item COLUMNS (site_id, domain, item_type, source, work_item_id) and
a few OPTIONAL (`?`) spec fields (issue?, page_name?, component_id?, reviewed_brief?),
plus the whole spec at `input_data.spec`. **It contains no `section_type`.** So a
component-creator dispatched via the loop gets `input_data.spec.section_type` (workflow
OK) but no top-level `section_type` → the same contract violation as 082. The generic
loop cannot satisfy a top-level-required contract without per-handler knowledge
(which 002 §414 explicitly avoids). PREDICTED, not yet run — confirm by dropping a
`needs_component_regeneration` item.

### Recommended fix (framework-level, best for the system)
Make the contract validator resolve each required field via the same convention
handlers consume by: accept the field if present TOP-LEVEL **or** at
`input_data.spec.{field}`. This aligns `call_agent` validation with 003's stated
"read spec fields from input_data.spec.*" rule, keeps the dispatch loop generic, and
removes the need for the 083 duplication. Do NOT add per-handler fields to the loop
mapping; do NOT enshrine the duplication as the rule. (Validator-checks-top-level-only
is inferred from runs 081/082/083 + the loop dump; confirm the check site in
call_agent.go before editing.)

### Interim recipe until the validator fix lands
Use the 083 manual shape (spawn generic orchestrator → spawn_agent component-creator →
call_agent with `section_type` top-level + full `spec`, function pinned). This is a
bespoke-caller workaround, not the canonical path.

### After regeneration — content-fill (do not skip, mind the hazard)
A regenerated component's new schema fields are EMPTY; the raised `needs_rerender`
only ASSEMBLES stored rendered HTML. To fill the fields, page-content-writer must
render the component with content for the target page (a page content-rebuild —
`needs_content_page`, 002 §684; confirm routing). HAZARD: a full-page content-rebuild
re-runs the writer over ALL sections (interactive-page clobber, 002 §498). On the
index that means hero + provocation-card. Re-emptying provocation-card at build time
is safe (the Path-2 loader refills it in-browser and the js_snippet survives), but
working sections must not be clobbered — render only the target component onto the
page, or confirm the others survive, before triggering.

---

## Appendix F — Index content-section deferral FIXED; provocation-card truncation repair + rebuild (YOU ARE HERE — 2026-07-02 ~19:50)

### Where we are
The 2026-07-02 full-index rebuild deferred brief-explanation, gauntlet-cta, system-stats
and dropped brief-explanation's page instance. Root cause (confirmed in
plan_sections_action.go `planSection`): each has a REQUIRED field whose data source
doesn't resolve and has no fallback → the required-field on_missing switch hits
`default:` ("Unknown on_missing — default to defer") → deferred → save_page_sections then
persists only the ready set, dropping the deferred instances. render_mode was a red
herring (deriveRenderMode returns `agent` merely when any field source=`llm`).

Two data gaps — both now fixed:
- **site_specs.cta.primary_url** (SHARED: gauntlet-cta.cta_primary_url +
  system-stats.cta_url). Was missing. FIXED — inserted a site_specs row aspect='cta',
  data={"primary_url":"/provocations/index.html",
  "secondary_url":"/tools/archetype-taster-quiz/index.html"}. [DONE]
  (resolver: ensureSpecs loads r.specs[aspect]=data; resolveSpecPath →
  r.specs["cta"]["primary_url"].)
- **site_assets.illustration** (brief-explanation.illustration_url). The resolver's
  ensureAssets only surfaces `hero` and `logo` (site_plan_imagery kinds hero/logo), so
  `site_assets.illustration` can NEVER resolve. FIXED — illustration_url set
  required=false + on_missing=skip_field (renders text-only). [DONE]

DISCOVERY: illustration assets DO exist (assets.illustration_game_master,
illustration_gauntlet_cta; purpose=illustration; active) but the resolver can't surface
them. Text-only is interim. FUTURE structural work: extend ensureAssets to surface
kind=illustration from site_plan_imagery+assets, OR set illustration_url fallback to a
specific illustration URL. (Note: gauntlet-cta's schema has no illustration field at all,
though an illustration_gauntlet_cta asset exists — asset/component mismatch to revisit.)

provocation-card truncation CONFIRMED (query C): html_template and page_components
rendered_html are byte-identical (md5 431095..., len 9994), both ending with a truncated
inline `<script>`: `...card.classList.remove('is-active');` then `});` followed by a stray backslash — stray backslash,
NO forEach close, NO IIFE close, NO `</script>`. An unclosed `<script>` swallows
`</main>`+footer → the live index renders short (no footer). PRE-EXISTING (the rebuild did
not change provocation-card — identical md5). This is the generation-truncation defect
(store validation checks unclosed `<style>` but not `<script>`).

### What to do next
1. Confirm the truncation marker is unique, then repair the TEMPLATE (source of truth —
   the rebuild re-renders provocation-card from it, so fixing the template is enough;
   the build overwrites rendered_html):
   ```sql
   SELECT (LENGTH(html_template) - LENGTH(REPLACE(html_template, E'});\\', ''))) / 4
          AS truncation_markers
   FROM content_components WHERE id='6163ff14-9f94-4962-aa19-d2718eabdeb1';   -- expect 1

   UPDATE content_components
   SET html_template = REPLACE(html_template, E'});\\', E'});\n  });\n})();\n</script>'),
       updated_at = NOW()
   WHERE id='6163ff14-9f94-4962-aa19-d2718eabdeb1'
   RETURNING RIGHT(html_template, 80) AS new_tail;   -- should end })();</script>
   ```
2. Re-run the index build (deferral now fixed → all 6 planned sections should render;
   provocation-card re-renders WHOLE from the fixed template):
   ```sql
   INSERT INTO site_work_items
     (site_id, source, item_type, summary, spec, handler_agent, status, created_by, item_key)
   VALUES
     ('9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'::uuid, 'manual', 'needs_page',
      'Rebuild index after deferral fixes + provocation-card repair',
      '{"domain":"vonc.com","page_id":"b4d24f8e-fccd-49df-9dad-aa56a0b20a68","filename":"index.html","page_name":"index"}'::jsonb,
      'page-build-handler', 'triaged', 'manual', 'manual-rebuild-index-' || gen_random_uuid());
   ```
3. Verify:
   ```sql
   SELECT pc.position, cc.function, pc.build_status,
          LENGTH(COALESCE(pc.rendered_html,'')) AS rendered_len,
          (pc.rendered_html LIKE '%})();%</script>%') AS pc_script_closed,
          pc.updated_at
   FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
   WHERE pc.page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'::uuid
   ORDER BY pc.position;
   ```
   Expect SIX rows incl brief-explanation, gauntlet-cta, system-stats with real content;
   provocation-card `pc_script_closed=t`; live index shows the footer again.
   If any of the 3 STILL defers, check its `query.*` fields (system-stats stat VALUES may
   be query-sourced, which this analysis didn't cover) — investigate, don't assume.
4. Separate framework track (do NOT fold into the above):
   - Harden store validation to catch an unclosed `<script>` (currently only `<style>`).
   - Extend the resolver to surface non-hero/logo asset kinds (illustration) so components
     can use illustration assets that already exist.

### Notes
- The re-run is also the TEST of the deferral fix (last run saved 3 sections; expect 6).
- provocation-card is runtime-filled by the Path-2 loader (snippets.js) — untouched by a
  page rebuild; the inline hover script we're repairing is separate/cosmetic.
- Do the template repair BEFORE the rebuild, or the rebuild re-renders the truncated
  provocation-card again.

### Appendix F — RESULT (2026-07-03 ~13:18) — DONE
Template repair: truncation_markers=1 (unique), REPLACE applied, new tail ends
`})();</script>`. [DONE]
Rebuild (needs_page, claimed by build-dispatch-loop, complete 13:17:56):
page_components now has SIX rows (was 3) — hero(2494), provocation-card(10015,
pc_script_closed=t), gauntlet-cta(8312), brief-explanation(10230), lobby-grid(15282),
system-stats(6562), all build_status=deployed, all updated 13:15.
=> Deferral fixes WORKED: brief-explanation, gauntlet-cta, system-stats rendered (were
   deferred/dropped last run). provocation-card truncation FIXED (script closed;
   rendered_len 9994→10015). Build passed validate_content (no unrendered placeholders).
Verify pending: live vonc.com/index.html (6 sections + footer restored + real copy);
DB check for empty headings / brand words on the 3 formerly-deferred sections.

REMAINING (separate tracks, not blockers):
 1. Wire illustrations — assets illustration_game_master / illustration_gauntlet_cta
    exist (files deployed + assets rows) but resolver ensureAssets surfaces only
    hero/logo. Options: extend ensureAssets to surface kind=illustration from
    site_plan_imagery+assets (structural, benefits all sites), OR set
    brief-explanation.illustration_url fallback to the illustration URL (quick, per-site).
    gauntlet-cta has NO illustration field despite illustration_gauntlet_cta existing —
    asset/component mismatch to revisit.
 2. Framework hardening: store validation should reject an unclosed `<script>` (currently
    only `<style>`) — this is what let provocation-card ship truncated. Plus the
    separateInlineJS truncation warning.
 3. provocation-card end-state: still Mode-B, filled at runtime by the Path-2 loader
    (snippets.js). Decide whether that stays the design or it moves to Path-1 (inline
    `<script>` extracted to js_content) like the other interactive components.
 4. Broader: other pages (about, archetypes, contact, provocations, tool pages) — audit
    for the same deferral/truncation classes now that the mechanism is understood.

---

## Appendix G — Live verification + two new findings (2026-07-03 ~13:25)

Content confirmed filled: brief-explanation / gauntlet-cta / system-stats all
has_empty_heading=f, has_brand_words=t. Live vonc.com/index.html (propagated to B2)
shows hero + gauntlet-cta + brief-explanation + system-stats with real copy. Deferral
fix + provocation-card truncation repair validated end-to-end.

FINDING 1 (priority — centerpiece regression): page_components has SIX rows but the
deployed HTML has only FOUR — hero(1), gauntlet-cta(3), brief-explanation(4),
system-stats(6). provocation-card(pos 2) and lobby-grid(pos 5) are ABSENT from the
deploy despite being page_components rows (deployed, fresh rendered_html @13:15). These
are the two Mode-B / inline-`<script>` sections; provocation-card is the daily-provocation
card (the site's centerpiece). On 2026-07-02 provocation-card WAS in the deployed HTML
(truncated shell); on 2026-07-03 it's gone. So the assembly/deploy step (page-rerender
deploy_page — "assemble from stored components") is dropping the two interactive/Mode-B
sections. HYPOTHESIS only (interactive-section handling at assembly) — confirm via the
rebuild deploy_result (what it assembled) or page-rerender assembly code before acting.
DO NOT assume. This is the first thing to look at.

FINDING 2 (content/schema quality): stat labels/suffixes show generic SaaS fallbacks
that don't fit Spark — gauntlet-cta "Happy Customers / Avg. Rating / Setup Time";
system-stats "14,203% / 61ms Avg. Split % / 9x". These are `static` fields carrying
generic default labels/suffixes; the LLM writes only the VALUES, so the fallbacks leak
through. Fix later at the component schema (Spark-appropriate static values, or make the
labels/suffixes llm-sourced). Applies to gauntlet-cta + system-stats.

STORE-VALIDATION HARDENING (definition, for the framework track):
store_generated_component's pre-store validation currently rejects `<no value>`
templates and checks unbalanced `<style>`, but NOT `<script>`. That gap let provocation-card
ship a truncated `<script>` (`});` followed by a stray backslash, no `</script>`) → passed validation → stored → broke
the page at render. Hardening = add a `<script>` open/close balance check alongside the
`<style>` one (reject/flag-for-regeneration on unclosed script) + a truncation warning in
separateInlineJS (unbalanced braces / trailing backslash). Prevents the CLASS of
"truncated template ships and breaks the page".

### Next tracks (do in turn; suggest FINDING 1 first)
 0. [suggested first] Investigate why provocation-card + lobby-grid are dropped from the
    deploy (page-rerender assembly / deploy_result). Centerpiece must be on the page.
 1. Wire illustrations (assets illustration_game_master / illustration_gauntlet_cta exist;
    resolver surfaces only hero/logo). Extend ensureAssets for kind=illustration, or set
    per-field fallback URLs.
 2. Store-validation hardening (above).
 3. Audit other pages (about, archetypes, contact, provocations, tool pages) for the same
    deferral / truncation / dropped-section classes.
 4. Stat-label fallbacks (FINDING 2) on gauntlet-cta + system-stats.

---

## Appendix H — lobby-grid build: where the three deliverables go + run order (2026-07-04)

Three files were produced for the lobby-grid runtime-fill build. Each has a DIFFERENT
destination — only one touches the live site, and not as a file.

### 1. lobby_grid_loader.js — SOURCE OF RECORD ONLY. Do NOT deploy this file anywhere.
Its full content is already embedded (dollar-quoted) inside lobby_grid_install.sql Part A.
Delivery to the live site is the Path-2 snippet pipeline, exactly like
provocation-card-loader:
  js_snippets row  →  site-asset-renderer bundles /assets/js/snippets.js  →  git commit to
  the 'sites' repo  →  GitHub Actions  →  Backblaze B2.
So: keep lobby_grid_loader.js in the library repo alongside provocation_card_loader.js
(versioned reference, per TOOL_DOCS_convention). If the loader ever needs editing, the edit
goes into the js_snippets ROW (UPDATE js_content) + re-trigger site-asset-renderer —
editing the .js file alone changes nothing on the site.

### 2. lobby_grid_install.sql — run against clients_db, in TWO parts (not in one go).
Where: the DB pod —
  kubectl -n ai-persona-system exec -it postgres-clients-0 -- psql -U clients_user -d clients_db
(or paste into your usual psql session). \d js_snippets first, per the schema rule.
  Part A (the js_snippet INSERT + verify SELECT)  →  run NOW.
  ── then trigger-asset-renderer-vonc.sh (bundles the loader into snippets.js) ──
  Part B (the data-runtime-fill marker: template + current index instance) →  run AFTER
  the renderer trigger. Expect template_marked=t and rendered_marked=t (one row each);
  a 0-row result means already-marked or literal mismatch — check before proceeding.
Afterwards keep the .sql file with the other numbered one-off scripts (081/082/083…) for
provenance; it is not re-run.

### 3. provocations.sample.json — its CONTENT becomes the live data file.
Destination: the 'sites' repo (github.com/gqls/sites) at  vonc.com/data/provocations.json
— i.e. REPLACE the existing provocations.json with this v2. Commit → GitHub Actions →
B2 → served at https://vonc.com/data/provocations.json.
Safe to replace: v2 is a SUPERSET — it keeps `today` (provocation-card's feed) and the old
`lobby` array (still read until the mini-lobby trim lands) and ADDS `arena` (lobby-grid's
feed). Interim hand-maintained data until the Phase-3 pipeline emits this file.

### Full run order (with the two triggers already in outputs)
 1. psql: lobby_grid_install.sql **Part A**  → verify js_len > 0 for 'lobby-grid-loader'.
 2. bash trigger-asset-renderer-vonc.sh      → snippets.js bundle commit.
 3. psql: lobby_grid_install.sql **Part B**  → template_marked=t, rendered_marked=t.
 4. git: commit provocations.sample.json content as vonc.com/data/provocations.json.
 5. bash rerender-index-vonc.sh              → index reassembled + deployed.
 6. Verify:
    curl -s https://vonc.com/index.html | grep -o 'data-component="[^"]*"'
      → now SIX incl. lobby-grid;
    curl -s https://vonc.com/data/provocations.json | grep -c '"arena"'  → 1;
    browser: the six arena cards fill (featured + wide styling, pulsing stat dots,
    cards navigate on click), provocation-card still fills, footer intact.
Then step 6 of PLAN_lobby-grid (next): trim provocation-card's mini-lobby.

### Appendix H — PROGRESS (2026-07-04 ~18:15)
- Step 1 DONE: js_snippet 'lobby-grid-loader' inserted (js_len 6743, applies_to ["lobby-grid"], active).
  The psql echo garbling in the paste was cosmetic (bracketed paste); the INSERT succeeded and the
  stored code is verified below. (6-byte length drift vs the source file — checked, not truncation.)
- Step 2 DONE: site-asset-renderer run COMPLETED (correlation 045afbf2, ~12s: load_site →
  render_js_snippets → deploy_js_snippets via git-adapter → complete). The shipped
  assets/js/snippets.js bundle now carries BOTH loaders; the lobby-grid-loader fragment extracted
  from the bundle passes a node syntax check as a complete IIFE (functions + graceful catch + init
  all present) — the shipped code is whole.
- NEXT: Step 3 marker SQL (lobby_grid_install.sql Part B — expect template_marked=t,
  rendered_marked=t) → Step 4 commit the v2 provocations.json (with `arena`) to the sites repo →
  Step 5 rerender-index-vonc.sh → Step 6 verify (six data-component values incl. lobby-grid; arena
  in the served JSON; six cards fill in a browser).

*(Formatting note, 2026-07-04: bare `<script>`-style tokens in this file's prose were breaking
markdown readers that parse inline HTML — the reader swallowed everything after the first bare tag,
mirroring the live-page bug itself. All such tokens are now backtick-wrapped and
backslash-terminated code spans reworded. If any other doc dies mid-render, the same sweep applies.)*

---

## Appendix I — Provocations page: 404 destination of all primary CTAs (opened 2026-07-04)

### Problem
B2 returns 404 NoSuchKey for `vonc.com/provocations/index.html`. That URL is the destination of
every primary action on the site: the hero "Enter the Gauntlet", the gauntlet-cta and system-stats
CTAs (`site_specs.cta.primary_url`), lobby-grid's "Enter the Arena" + all six arena card urls, and
provocation-card's primary CTA. The site's main action currently dead-ends.

### Evidence
`pages` row e4b3b195-919f-45ad-854e-201d3e846ea8: name `provocations-index`, url
`/provocations/index.html`, title "Provocations Archive | Spark", page_type `section-index`,
**build_status `planned`, updated 2026-06-22** — planned in the original build, never built.
(Schema note: `pages` has `name`, not `page_name` — corrected via `\d pages`; the earlier query
erred on the column, not the data.)

### Gate before building (do NOT blind-fire needs_page)
D1. Plan sections for `page_name='provocations-index'` in the current site_plan (what the page
    should contain).
D2. State of those components (`is_mode_b`, schema_field_count, render_mode) — Mode-B or
    unresolvable-required-field sections would be DEFERRED and, post assembler patch, empty ones
    are DROPPED → a blind build could deploy a husk of an archive page.
D3. Prior work items for this page (`page_id` or item_key match) — a failed item may state exactly
    why the pipeline never built it.

### Build (provisional — after the gate)
Mirror the proven index rebuild item, page_id authoritative:
```sql
INSERT INTO site_work_items
  (site_id, source, item_type, summary, spec, page_id, handler_agent, status, created_by, item_key)
VALUES
  ('9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'::uuid, 'manual', 'needs_page',
   'Build provocations-index (destination of all primary CTAs; currently 404)',
   '{"domain":"vonc.com","page_id":"e4b3b195-919f-45ad-854e-201d3e846ea8","filename":"provocations/index.html","page_name":"provocations-index"}'::jsonb,
   'e4b3b195-919f-45ad-854e-201d3e846ea8'::uuid,
   'page-build-handler', 'triaged', 'manual', 'manual-build-provocations-' || gen_random_uuid());
```
If D2 shows broken/dynamic sections: fix first (data gaps / runtime-fill marker + loader / component
regeneration), same playbook as the index — the archive's list section is likely the dynamic one and
may need the provocations feed (Phase-3 family).

### Verify
Work item → complete; page build_status → deployed;
`curl -s -o /dev/null -w "%{http_code}" https://vonc.com/provocations/index.html` → 200;
click-through from the arena cards + hero CTA lands on the archive; footer intact.

### Appendix I — GATE RESULTS (2026-07-04): zero planned sections; seven silent no-op completes
D1: 0 site_plan_sections for page_name='provocations-index' in the current plan (query shape proven
on the index page; one confirm pending: `SELECT DISTINCT sps.page_name FROM site_plan_sections sps
JOIN site_plans sp ON sp.id=sps.plan_id WHERE sp.site_id='9ec3b9ee-...' AND sp.is_current ORDER BY 1;`
plus `SELECT sections FROM pages WHERE id='e4b3b195-...';` — expect '[]').
D3: SEVEN items for the page, ALL complete, no errors (needs_page 06-22 +14s after page creation;
2× manual needs_page 06-26; 4× page_rerender 06-23→07-01) — and the pages row untouched throughout
(build_status='planned', updated_at = creation instant). The handler exits on the zero-sections path
before any page-writing step; the rerender skips deploy with no page_components. Silent no-op ×7.
=> Do NOT fire the provisional needs_page INSERT — it would be silent no-op #8. The fix is to give
the page PLAN SECTIONS first (header/hero + an archive LIST component, kind=dynamic, provocations
feed — section-descriptor/loader-builder design; pairs with the complex-tool loop chat), THEN build.
Framework prevention (guide §9 "Page build completes having built nothing"): planner ≥1-section
invariant at plan-store; handler zero-PLANNED vs zero-READY guard (fail/raise, never silent);
rerender skip-warn; auditor rules (linked+planned beyond threshold; post-deploy URL presence).
