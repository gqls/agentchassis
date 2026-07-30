# HANDOFF — brochure component library / fundamentallyai.com — 2026-07-30

**Cold-start document.** Written because this thread's context is large and the owner
asked for a clean start point. Read this file, then follow its pointers — do not re-read
the whole NOTES file unless you need a specific detail it names.

## What this session actually was

Started from `bugs_open/079` (phantom links). That bug was **closed by another thread**
before this session did anything — nothing was owed on it. Everything below is what
this session did instead, in order.

## DONE and verified — do not redo

1. **079 confirmed closed elsewhere.** No action needed.
2. **Dead CSS background fixed** (`hero.jpg` on 3 pages) — commit `69a93dffc`.
3. **§4c carousel built**: `teaser-reveal-panel`, a new component implementing the
   pre-existing `experience_patterns` row `teaser-detail-deeplink`. Live on
   `index.html` originally, now on 4 pages (see below). CLC-012 in the concept register.
4. **Four owner polish decisions actioned**: decision-record overclaim softened on 7
   pages/12 instances; 10 dead `#fragment` links (5 rewritten, 5 unlinked); em-dashes
   fixed in 3 templates + the tool-generator prompt.
5. **Site-wide repeated-capability-list content found and fixed.** 18 sections across 5
   pages independently restated the same 9 company facts (home/capabilities looking
   "very similar" was the visible tip of this). Fixed by keeping one list per page.
   **Root cause filed as `bugs_open/151`** — the per-section content writer gets the
   whole site's fact pool with no record of what a sibling section already said, and the
   once-per-site planner (the only place with cross-page visibility) doesn't guard
   against it. **Not built** — a real platform change (touches the planner + the write
   path), owner's call on priority.
6. **"Carousels on almost every block, with images"**: `teaser-reveal-panel` gained
   optional `image_url`/`image_alt`, rolled out to `index.html` (retrofit),
   `capabilities.html` (replaced `services-grid`), `multi-agent-review-council.html`
   (replaced `info-card-grid`), `model-fine-tuning.html` (**merged** two card-grid
   sections that were independently restating 4-of-6 facts each — the 151 pattern again,
   page-local). All images already existed on the site (~25 generated assets); none
   newly generated.
7. **Owner feedback round 1** (padding / visual ellipsis / open-state merge / real
   carousel arrows): fixed. The visible "…" is CSS-drawn (`content:"\2026"` on
   `::after`), **never a stored character** — a literal one is exactly what this
   component's own no-ellipsis rule forbids (a truncation checker reads a trailing
   ellipsis as damage). Arrows reuse `hero-card-carousel`'s exact `goTo`/`nearestIndex`
   pattern rather than a second implementation.
8. **THE IMPORTANT ONE — owner feedback round 2 found a bug that had been live since
   Round 1.** Arrows didn't scroll, opening one card didn't close another. Root cause:
   `<script src=".../snippets.js">` sits in `<head>`, non-deferred, executing **before**
   the panel's own markup exists in `<body>` — so `behaviour.js`'s DOM query always found
   zero panels and the whole file silently no-opped, every load, since the very first
   version. **Every prior "verified" check on this component — the render harness,
   `probe_reveal_open_state.py` — exercised static markup or forced DOM state directly.
   None of them ever called `.click()`, so none could have caught this.** Found only by
   simulating real clicks in headless Chromium. Fixed with the same
   `document.readyState` → `DOMContentLoaded` guard `hero-card-carousel`'s own snippet
   already has (6 of 7 active `js_snippets` already follow this convention — checked
   fleet-wide, **not a platform gap**, just this component missed it; `news-date-formatter`
   is the one other exception, unchecked). Also redesigned the open state to literally
   **replace** the whole closed block with one merged paragraph (hook+continuation+body,
   no line break at the old cut point) instead of appending body under a still-visible
   cut-off line, and fixed a genuine padding mismatch between closed/open states.
9. **Portfolio logos wired in** (this session, after kubeconfig refresh): relojistas,
   idea.uk and leopardessconsulting's real logos (verified live, visually inspected)
   now render in a consistent white chip above each portfolio card's title. Added
   `logo_url`/`logo_alt` to `portfolio-showcase`'s schema+template AND to
   `site_specs.portfolio.projects` (the source of truth per the component's own
   `source:` declaration) AND to the live `content_data`. Verified by screenshot.

**Commits, in order:** `69a93dffc`, `d42ebc44f`, `2d9ed63b8`, `4988b8674`, `2e97f8929`
(151 filed), `4b2e157df`, `dbd59bbde`, `763ceb81c`, `1ec9e8cf6`, `dec1807aa`, `3fec63705`
(summary), plus this session's uncommitted-as-of-writing portfolio-logo work (commit
before you do anything else if it isn't already committed — check `git log` first).

**Full technical log:** `NOTES_brochure_component_library.md`, entries from
`## 2026-07-29 (later)` (line 2713) onward. **Plain-prose owner log:**
`README_where_we_are.md`, same date range. **Milestone read-out:**
`SUMMARY_2026-07-30_the_panel_is_finished_and_two_new_fronts_open.md`.

## READ THIS BEFORE DOING ANYTHING ON STEP 5 — it is already done

The owner asked (this thread) to evaluate how the framework's tool-building pipeline
could be improved to support hand-building a complicated component/tool with equivalent
rigour. **A parallel thread already answered this**, using this session's own
`teaser-reveal-panel` build (rounds 1–5, including the click-bug in item 8 above) as its
evidence base. Do not re-derive it.

- **The proposal**: `docs024_key_docs_latest/staged_component_build/PROPOSAL_2026-07-30_step_by_step_build_with_stage_gates.md`
  — Part 1 is an evidenced, commit-by-commit provenance of the carousel; Part 2 is the
  design. Core finding: the platform already has almost everything needed (travelling
  docs tables `doc_plans`/`doc_notes`, the Tiers 0–4 verification ladder,
  `browser-runner-adapter` with a real `interaction`/`text_matches` check type already
  proven end-to-end by a same-day pilot) — **for tools, not for components**. The gap is
  wiring, not construction. Proposes an 8-stage ladder (S0 shape → S7 regress); **S6
  ("does it operate when driven, in real Chromium") is the stage that would have caught
  item 8's bug**, and is the single most important stage in the ladder.
- **The feature anchor**: `features_open/027_FEATURE_staged_part_build_with_stage_gates.md`.
  Line 7: **"Owner wants the work done in a SEPARATE thread."** — same instruction the
  owner just gave this thread. Do not build the staged-gate system here either.
- Cross-references worth knowing: `webdesign_tools_repair/REPORT_2026-07-29_concepts_for_a_working_tools_chain.md`
  (the tools-chain gap analysis, G1–G5, that the proposal's reuse inventory is built on);
  `docs024_key_docs_latest/travelling_docs/OVERVIEW_self_verifying_tools.md` (the
  plain-language tour of the tool-doc system — **this is "the doc traveller"** the owner
  keeps referring to); `RUNBOOK_travelling_docs(38).md` §0 (the historical Stage 0–6
  rollout tracker for tools, the direct precedent for the new proposal's ladder).
  **Landmine**: that directory has 39 numbered copies of each travelling doc
  (`RUNBOOK_travelling_docs(1..39).md`) — highest number is current; the bare
  `RUNBOOK_travelling_docs.md` is NOT the latest.

**What this means for you, next thread:** read the proposal and the feature file. If the
owner wants step 5 progressed, it means either (a) deciding the open questions in the
proposal's own "Open questions the next thread should decide" section (six of them,
listed there, e.g. one checker-agent-per-stage vs. one parameterised agent; who fires
S6 and when; whether a gate can refuse vs. only report), or (b) starting the ordered
build list in Part 2 (component travelling docs first — "without this there is nothing
for any stage gate to read"). It does not mean writing another proposal.

## NOT DONE — the actual next build

**An interactive tools page, with sliders/inputs, built on real live platform numbers**
(pages/sites hosted; council decisions by verdict and by reviewer seat). Owner-confirmed
requirements, verbatim from this thread:

- **"The tool wants to have some interactivity like a slider or two or inputs."** — not
  a passive dashboard. Style should match the two existing tool pages on this site
  (`/tools/llm-cost-calculator/`, `/tools/model-approach-selector/`) — **look at how
  those are actually built first** (their `content_components` rows, `page_type`,
  whether they're `tool-*` function components or something else) before designing a
  third pattern. Not yet investigated this session — do it first.
- **Real numbers, not invented ones.** Candidate sources already known to exist:
  `sites`/`pages` counts; council verdicts — likely `doc_notes` filtered by
  `categories ? 'council-gate'` (see CLAUDE.md's own council-gate section for the
  query shape) or `orchestration_states` for the underlying runs. **Query these fresh —
  don't trust any count already written down in this handoff or elsewhere, they go
  stale within days** (established pattern this whole session).
- **"Please document the tool builds as you go as per the doc traveller."** Given the
  step-5 finding above, the honest options are: (a) hand-maintain a PLAN+NOTES pair for
  this tool in the travelling-docs STYLE (files, like `teaser-reveal-panel`'s own
  `components/` directory already does) since the DB-table version
  (`doc_plans`/`doc_notes`) is explicitly step-5's *unbuilt* proposal for components; or
  (b) check whether `doc_plans`/`doc_notes` already accept an arbitrary subject key
  cheaply enough to use directly even before step 5's wiring is built — this is
  literally the proposal's own **first UNVERIFIED item**: *"whether `doc_plans` fits a
  component's needs without schema change... it is one query and it gates the whole
  design, so it is written as the next thread's first action."* Worth running that one
  query before deciding which of (a)/(b) to do.
- **Known hazard for ANY interactive component on this platform**:
  `TL-001`/`save_page_sections_action.go`'s "INTERACTIVITY REGRESSION BLOCKED" guard
  (confirmed present in the current source this session) protects a tool's widget from
  being silently wiped by a routine content rebuild — but it is a guard against DELETION,
  not a guarantee of correctness. Given item 8 above, **build this tool's own render
  harness + mutants (S2), and prove S6 (real clicks in a real browser) before calling it
  done** — do not repeat the mistake that cost this session five rounds. If the S6
  wiring from step 5 doesn't exist yet, hand-roll a real-click test the way
  `render_teaser_reveal_panel.go` + a Puppeteer-less headless-Chromium click script did
  this session (`scripts/` in this same directory has the working pattern — search NOTES
  for "real click" to find the exact technique, since file:// + injected `<script>` +
  `--dump-dom` was the only combination that reliably worked in this sandbox; a plain
  `--screenshot` on a `file://` copy did NOT reliably paint remote images — use a live
  `https://` fetch + tall `--window-size` + crop for any visual/screenshot verification
  instead).

## Portfolio logos — one loose end

Verify the portfolio-logo commit actually landed (`git log --oneline -3` first thing) —
this was done right before this handoff was written and may not be committed yet
depending on when you're reading this.

## Landmines carried from this whole session (still true, not re-explained here)

- A section's placement lives in **three places** — `site_plan_sections`,
  `pages.sections`, `page_components.slot_name` — miss one and a `complete` re-render
  silently drops the section.
- `assets.url` can show a broken placeholder while the real stable
  `/assets/images/<key>.jpg` path (note: hyphens, not underscores) is live 200 — check
  the stable path, never the tracking column.
- Kubeconfig token expires ~every 3 days; the owner refreshes it (just done, so you have
  a few days from 2026-07-30).
- `js_snippets` and several `content_components` rows are technically fleet-shared
  (visible to every site) but currently used by only 1 site each — editing their
  template/JS is low-risk in practice, but always check `usage_count`/distinct-site-count
  before assuming that.
- Any DB write touching more than one table (placement, content + archive) should go in
  one transaction with a verifying `SELECT` before `COMMIT` — this session hit one real
  SQL bug (`pages.id` vs `page_components.page_id` column confusion) and the transaction
  rolled back cleanly with zero side effects, which is the whole reason to do it this way.

## Standing docs in this directory (all current as of this handoff)

`PLAN_2026-07-29_teaser_reveal_panel.md` (design), `RUNBOOK_brochure_component_library.md`
(commands + gotchas), `NOTES_brochure_component_library.md` (full technical log),
`README_where_we_are.md` (plain-prose owner log), `SUMMARY_2026-07-30_...md` (milestone
read-out), this file. `components/teaser-reveal-panel/` and
`components/portfolio-showcase/` hold the version-controlled source for both components
touched this session.
