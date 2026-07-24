# 016 FEATURE — brief-fidelity audit ("did the machine build what was asked?")

**Raised:** 2026-07-24, by the owner, reviewing fundamentallyai.com ("First, it
is nothing like the brief"). **Priority 1 of the site-quality automation set**
(016/017/018/019). **Status:** specified, not built.
**Owner:** brochure_component_library workstream (first user), generic by design.

## The gap

Nothing in the platform compares the **rendered site** against its own
`mission_brief` / `design_intent`. The build pipeline verifies completeness
(pages built, sections filled, images present) and correctness (links, claims,
placeholders) — never **fidelity**: fundamentallyai's brief asked for
Bain/BCG/McKinsey-register visual language, card carousels, hover-zoom imagery,
swipeable strips, code-rendered charts, image-rich pages, and layouts that vary
per page. The pipeline delivered standard text-heavy sections, sparse imagery,
zero charts — and every existing check passed. "Nothing like the brief" was
invisible to the machine because no check reads the brief.

Evidence this class is real: fundamentallyai built overnight 2026-07-20 with
`design_intent` explicitly encoding carousels/charts/line-illustration; the
first human look (owner, 2026-07-24) found the mismatch no automation had.

## What it is

A per-site audit agent (`brief-fidelity-auditor`) that:
1. Loads `site_specs` (`mission_brief`, `design_intent`, `content_direction`) +
   the rendered pages (HTML + **screenshots** — the P3 screenshot machinery
   exists) + the page/component inventory.
2. Grades, against the brief's own promises, e.g.: promised component classes
   present (carousel? chart? interactive)? imagery density vs "image-rich"?
   layout variety across pages vs "vary the design"? palette/typography register
   vs the named reference style? every grade cites the brief line it grades.
3. Files `site_work_items` (`item_type='brief_fidelity_gap'`, one per concrete
   gap, severity by how load-bearing the promise was) — findings become
   routable work, not a chat report.
4. Runs post-build (after first full build completes) and on demand; NOT on the
   high-frequency sweep (it is an LLM audit, costlier than the mechanical checks).

**Model: Gemini** (owner call 2026-07-24 — use Gemini as the design-capable
model for now and evaluate; same model family as the swapped content-writer).

## Notes / constraints

- The audit judges against the SITE'S OWN brief, never a generic taste standard
  (that's 018's job). A site whose brief asked for austere text pages must PASS
  with austere text pages.
- Grades must cite the brief verbatim (the claims-verification discipline
  applied to design promises).
- Screenshots matter: HTML inspection alone cannot see rendered density/variety.
- Relates: 017 (component adoption — one specific fidelity dimension made
  mechanical), 018 (taste critique — no brief needed), features_open/012's
  lesson (a gate that counts rows misses what the rows LOOK like).
