# PLAN 2026-09-02 — gamesdesign.co.uk lane opens; first task is the brand rename

## Lane

This directory is the workstream for **gamesdesign.co.uk** (site id
`e33263f4-74f8-494f-b191-546845dbbddf`). Opened 2026-09-02 when the owner named a
session "gamesdesign.co.uk" and told it to announce itself to the Portfolio
positioning lane. Until then no thread owned the site — positioning's NOTES called it
"the healthy sibling, another lane's build" and had been holding evidence for it.

Positioning coordinates to honour (register rows GD1/GD2, per the P5 twin split):
**this site is the authority seat** — free browser tools + guides + learning, for solo
devs/students/small teams. gamedesign.uk (rebuilding, its own session) takes the
professional/studio-practice side (editorial, process, workflow) and is the compatible
future home of any paid tier. **Do not drift this site into studio-practice editorial
ground.** The literal name "GameDesign.uk Pro" appears in NO current spec of this site
(peer-verified 2026-09-02) — not a recorded fact.

## Task 1 — the brand rename (owner ruling 2026-09-02)

The ruling, relayed by both the Portfolio positioning and gamedesign.uk sessions:
**gamesdesign.co.uk must stop calling itself "GameDesign.uk"** — that brand belongs to
gamedesign.uk, and two live domains cannot share one brand.

Root cause (class bug, filed by the gamedesign.uk session as
`bugs_open/439_HANDOFF_2026-09-02_adoption_carries_the_source_sites_brand_name_verbatim_into_a_destination_on_a_different_domain.md`):
the June adoption inherited the crawled site's identity verbatim
(`company_name: 'GameDesign.uk'`, `adopted_from: 'gamedesign.uk'`), and
content-gap-planner carried it forward on 2026-08-17. `sites.company_name` /
`logo_text` / `tagline` are all EMPTY, so the spec was the only source. 439 proposes
the seam fix (apply_adoption_plan reconciling against destination_domain); **the seam
is not this lane's to fix, the instance was.**

### Decisions

- **New name: "GamesDesign.co.uk"** — owner's explicit choice 2026-09-02 from
  positioning's recommendation (domain-as-brand, the estate convention). GD1 to be
  moved recommended→decided by the positioning lane (told 2026-09-02).
- **Case-SENSITIVE replace** of the exact string `GameDesign.uk` →
  `GamesDesign.co.uk`, so lowercase `gamedesign.uk` survives everywhere: it is a
  historical fact (`identity.adopted_from`) and a real cross-link
  (guide-p2p-architecture → https://gamedesign.uk/games/p2p-networking/index.html).
- **Specs superseded, never updated in place** (partial unique index on
  `(site_id, aspect) WHERE is_current`; retire-then-insert as separate statements in
  one transaction — a chained CTE hits the index).
- **All carrying stores fixed in one pass**, or the next planner run re-derives the
  old name from the one missed: specs AND site_plan_pages AND pages AND
  page_components content_data. rendered_html also swapped mechanically so the stored
  artefact is consistent even before rerenders land.
- **Rerender, not regenerate**: a rerender merges content_data (the fix survives); a
  regeneration replaces it. Dispatched per page (not site-wide `rerender-pages`)
  because the site-wide route queues hours behind the estate and `bugs_open/315`
  records a page skipped by four rerenders — per-page dispatch gives per-page
  correlation ids and per-page served verification.

### Measured footprint (2026-09-02, live DB, before the fix)

| store | hits |
|---|---|
| site_specs (current) | 4 rows: briefing, design_intent, identity, tools |
| site_plan_pages.title | 22 rows |
| pages.title | 23 rows (21 as explicit `\| `/`- ` suffix) |
| pages.meta_description | 1 row |
| page_components.content_data | 30 rows across 21 pages |
| page_components.rendered_html | 10 rows |
| site_components / site_plans / plan sections / imagery / directives | 0 |
| served homepage | 30 occurrences incl. `<title>`, og:title, JSON-LD, tool-card titles |

Union of pages needing redeploy: **32** active pages of 49.

### Verification plan

- In-transaction `DO` guard: zero residuals in every store, exactly 4 current specs
  carrying the new name (demand control), `adopted_from` still lowercase. (PASSED at
  execution.)
- Post-rerender: per served page, cache-busted — old string count 0 (case-sensitive),
  new string present (demand control), and the lowercase cross-link on
  guide-p2p-architecture still present (control that the replace stayed narrow).
- NOT claimed: any fleet-wide seam verification — per 439 §6, my rows are manually
  renamed, so a seam fix must be verified on a NEW cross-domain adoption, not here.
