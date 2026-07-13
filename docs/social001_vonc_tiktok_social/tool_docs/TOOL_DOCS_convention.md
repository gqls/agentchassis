# Convention — Per-Tool Documentation That Travels With The Tool

**Created:** 2026-06-29
**Purpose:** Define a lightweight, consistent documentation practice so that
every tool/complex-component carries its own reasoning history — aims, the spec
it came from, choices presented, decisions made, bugs found, fixes applied, and
recurring problem categories — so that any future maintenance (by an agent or a
human) is fully informed and pulls in the same direction instead of re-deriving
lost context.

This sits below the GLOBAL debugging guide (016 / 016b) and the site-level
runbooks. Global docs hold cross-cutting patterns; site runbooks hold a site's
build/repair state; **per-tool docs hold one tool's whole story.**

---

## Why

Tools and games get complex, and bugs recur in categories (CSS-variable
mismatch, empty-shell, broken-template-slots, de-tool-on-rebuild, content-vs-
runtime mismatch, …). Without travelling documentation, each fix starts cold:
the fixer can't see why the tool was built this way, what was already tried, or
which decisions are deliberate vs accidental. The result is fixes that fight
earlier decisions. Travelling docs make the tool's intent and history available
at the moment of the fix.

The end-state vision: **domain name → imaginative spec → tools created and
maintained →** with documentation generated and accumulated automatically at
each step, so a tool's plan is written when it's created and its notes grow with
every change.

---

## The two documents per tool

Each tool/complex-component gets exactly two travelling docs, keyed by the
tool's identity (component `function`, e.g. `gauntlet-interface`):

### 1. PLAN — `PLAN_<function>.md`  (the intent; changes rarely)
Written when the tool is created; updated only when direction changes.
Contents:
- **Aim** — what the tool is for, in product terms.
- **Source spec** — the slice of the site spec / roadmap it derives from
  (paste or link the relevant `section_types`, `purpose`, `content_context`).
- **Behaviour contract** — what it does; for interactive tools, the states,
  inputs, outputs; for data-driven tools, the data source + the JSON/DOM
  contract it fills.
- **Delivery mechanism** — Path 1 (component inline `<script>` →
  `/tools/assets/{function}.js`, auto on rerender) vs Path 2 (library
  js_snippet → snippets.js) vs build-time content (content writer fills a
  schema). State which and WHY.
- **Dependencies** — data feeds, assets, other components, scheduled tasks.
- **Deliberate decisions** — choices that must NOT be "fixed" by a later pass
  (e.g. "intentionally JS-required because content is daily-regenerated").

### 2. NOTES — `NOTES_<function>.md`  (the history; append-only-ish)
A timestamped log. Every entry: date, what was observed, what was decided/done,
and WHY. Captures:
- **Choices presented + decision taken + rationale** (the fork and why we went
  the way we did).
- **Bugs found + fixes applied** (symptom → root cause → fix → verification).
- **Recurring problem categories** — tag each relevant entry with a category
  (see taxonomy below) so patterns surface.
- **Dead ends** — what was tried and abandoned, so it isn't retried.

---

## Problem-category taxonomy (tag NOTES entries)

Shared vocabulary so patterns are greppable across tools and can be rolled up
into the global debugging guide. Seed set (extend as needed):
- `css-variable-mismatch` — template uses a var name not in the stylesheet.
- `empty-shell` / `mode-b-template` — html_template is bare `<no value>`,
  names lost, empty schema; CSS selectors intact.
- `broken-template-slots` — `<no value>FIELD</no>` (Mode A, repairable).
- `content-vs-runtime-mismatch` — static content given a runtime loader, or
  daily data baked at build time.
- `detool-on-rebuild` — a content rebuild regenerates an interactive page from
  plan_sections and drops the tool.
- `js-not-extracted` — component has inline `<script>` but js_content empty, so
  `/tools/assets/{function}.js` isn't produced.
- `js-bundle-stale` — a js_snippet changed but snippets.js wasn't re-rendered
  (site-asset-renderer not triggered).
- `schema-template-drift` — input_schema and template slot set disagree.

---

## Where the docs live (so they travel)

Three layers, in order of how this should evolve:

1. **Now (this chat / manual):** markdown files named `PLAN_<function>.md` and
   `NOTES_<function>.md` in the outputs folder / project knowledge.
2. **DB as source of truth (recommended target):** a `tool_docs` table keyed by
   `function` (library-level, like the component itself), columns roughly:
   `function`, `doc_type` ('plan'|'notes'), `body` (markdown/text),
   `updated_at`. Library-level so the plan describes the tool TYPE and the notes
   accumulate both library fixes and per-site incidents (a per-site incident
   entry names the site_id). Travels through forks/regenerations because it's
   keyed by function, not by a transient component_id.
3. **Repo (optional mirror):** the pipeline could also commit a read-only copy
   under a `/.tooldocs/<function>.md` path for human browsing alongside the
   deployed site, but the DB row is authoritative.

---

## Pipeline integration (the vision — how this becomes automatic)

This is a feature for the tool lifecycle (docs 005 Tool Pipeline, 020 Tool
Lifecycle, 026 Component Regeneration). It is not built yet; this convention is
the spec.

- **On creation** (tool-suggester / tool-generator / component-creator): write
  the PLAN doc from the material the agent already has — the spec slice, the LLM
  reasoning for the design, the chosen delivery mechanism. The reasoning that is
  currently discarded after generation is exactly what the PLAN should capture.
- **On any modification** (component-template-fixer, component regeneration,
  manual fix): append a NOTES entry — symptom, root cause, fix, verification,
  category tag. An agent doing the fix writes the entry as part of its workflow.
- **On a new bug** (maintenance entry point): the maintaining agent LOADS
  PLAN + NOTES for that function FIRST, the same way we load site context — so
  the fix is informed by intent and history. This is the payoff.
- **Roll-up:** NOTES category tags feed the global debugging guide — when a
  category recurs across many tools, it graduates from local note to global
  pattern with a systemic fix (exactly how the css-variable-mismatch and
  empty-shell issues surfaced on vonc).

---

## Relationship to existing docs

- **Global debugging guide (016 / 016b):** cross-tool patterns and durable
  invariants. Per-tool NOTES feed UP into it via category tags.
- **Site runbooks (e.g. RUNBOOK_vonc_migrations.md):** a SITE's build/repair
  state across all its components. Per-tool docs are narrower (one tool, across
  all sites that use it).
- **Site running notes (e.g. RUNNING_NOTES_vonc.md):** chronological log of work
  on a site. When a tool matures, its tool-specific entries can be split out
  into `NOTES_<function>.md`.

---

## Applying it to the current work

Candidate tools on vonc/Spark that warrant their own travelling docs:
- `gauntlet-interface` (the daily Gauntlet — complex interactive)
- `tool-archetype-taster-quiz` (+ `archetype-result-card`)
- the **provocation daily-feed feature** (provocation-card + loader + the Phase 3
  pipeline) — currently documented site-level in RUNNING_NOTES_vonc.md and
  PLAN_spark_provocation_pipeline.md; these effectively ARE its plan + notes and
  can be renamed/split to the per-tool convention when convenient.

Suggested next step (if adopted): instantiate `PLAN_<function>.md` +
`NOTES_<function>.md` for these as we touch them, rather than retroactively, so
the practice starts without a big migration.
