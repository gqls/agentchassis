# PLAN — bugs_open/045: a generic `hero-tool` component

**Started:** 2026-07-21 · **Branch:** `085_debug_and_feature_loops` · **Bug:** `bugs_open/045`

## The problem (as inherited, then re-verified)

`hero-tool` is a **generic** section request. The selector
(`component_selector.go` → `queryCandidates`) resolves it by
`section_type = 'hero-tool' AND component_level='section' AND is_active AND
forked_from IS NULL`, scores the matches, and returns `candidates[0]` (there is
**no score threshold** — if there is exactly one candidate it always wins). The
library held exactly one such row: `bayesian-ranking-hero-tool_pre_037`, a hero
for a Bayesian ranking product with 14 `source:static` fields whose Bayesian
fallbacks re-apply on every render and cannot be overridden by `content_data`.
So every tool page asking for `hero-tool` inherited "Start Ranking Free" /
"Calculate Rankings" / "Try the Bayesian Ranker".

**Not a planner bug** (the plan asked for the string `"hero-tool"`), **not a
selector bug** (it correctly resolved the only match) — a **library gap**.

## Design decisions (and why)

1. **Build a generic `hero-tool` component (candidate 1)** — `function =
   section_type = hero-tool`, `component_level='section'`. Every visible label is
   `source:llm` (speaks the page's own vocabulary) — **zero `source:static`
   product fallbacks**. CTA anchors **gated** `{{if .x_url}}`; `*_url` fields
   `source:renderer` (resolver-owned, never LLM-authored) → satisfies LNK-005 by
   construction and derives cleanly under schema-derived CTA pairing (023 class
   H). Trust stats are **optional + gated** with anti-fabrication guidance
   (bugs_open/043 — omit rather than invent). Same shape as `tool-guide-intro`
   after migration 179.
   - **No embedded product widget.** The Bayesian row baked a whole ranking
     widget into the hero's right column. The generic hero is a single-column
     intro only — the actual tool is a **separate `tool-<slug>` section** on the
     page (verified: gamesdesign's ranking tool is `tool-bayesian-ranking`,
     `component_level='tool'`, position 3). The hero widget merely duplicated it.

2. **Retire the Bayesian row from the pool (candidate 2), do NOT delete it.**
   Move its `section_type` `hero-tool` → `bayesian-ranking-hero-tool` (its own
   function). It stays `is_active=true`, **function unchanged**. After this the
   generic component is the **sole** hero-tool candidate → deterministic
   selection regardless of scoring. Deleting was refused: it is the sole active
   row for its function (023 R10), and 16 `_pre_037` rows are in the same state.

3. **Do these two atomically in one migration** (183) so there is never a moment
   with zero or two hero-tool candidates.

4. **Touch no deployed page.** The change is DB-config (live immediately) and
   only affects the NEXT plan/rebuild. `page_components` rows bake
   `component_id` + `rendered_html`; nothing here rewrites them. The one deployed
   Bayesian hero (gamesdesign) is untouched; on a future rebuild it goes generic
   and keeps its ranking function (separate section).

## Blast radius (measured 2026-07-21)

Three pages request `hero-tool`:
| site / page | build_status | placement | outcome of fix |
|---|---|---|---|
| finetuning.uk / ai-agent-roi-estimator | needs_rebuild | none | rebuild → generic hero ✔ |
| ai-agent-orchestration.com / agent-complexity-estimator | needs_rebuild | none | rebuild → generic hero ✔ |
| gamesdesign.co.uk / bayesian-ranking | **deployed** | Bayesian hero @pos1 | untouched now; future rebuild → generic (real tool is a separate section) |

37 active `page_type='tool'` pages fleet-wide are latent-exposed; all now covered.

## Status

- Migration **183 written, applied, all post-conditions green** (2026-07-21).
  DB now: `hero-tool` is the sole selector candidate (score 0.69); Bayesian row
  retired to `section_type='bayesian-ranking-hero-tool'`, still active.
- Remaining: rebuild-based behavioural verification (a rebuild is what arms the
  bug, so a clean rebuild is the definitive proof).

## Not done (deliberately, for a follow-up)

- **Candidate 4 (selection-sanity check)** — a build-time warn when the only
  candidate for a generic section name is a product-specific component. Cheap and
  worth it, but it is Go (needs an image roll) and is the shared branch with
  `bugs_open/039`. Left for a dedicated change; noted in the bug file.
