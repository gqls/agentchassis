# NOTES — bugs_open/045 hero-tool component (append-only, newest at bottom)

## 2026-07-21 — investigation

- Read the handoff (bugs_open/045) and the parent (023) + robot-hands + cta
  memories. 045 was split out of 023 as class F; 023 is owned by the
  cta_link_integrity workstream — I did **not** touch 023's scope.
- Traced the resolution path in code:
  - `plan_sections_action.go` Path 1 (`components[sectionName]`) keys by
    name/function — a `hero-tool` section MISSES it (Bayesian function is
    `bayesian-ranking-hero-tool`). So it falls to Path 2, the selector.
  - `resolveSectionComponent` → `SelectComponentByType` → `queryCandidates`
    (`component_selector.go:164`): matches on **`section_type = $1`**,
    `component_level='section'`, `is_active`, `forked_from IS NULL`; scores; takes
    `candidates[0]`. **No minimum-score cutoff** — a sole candidate always wins.
    This is THE fact that makes the fix deterministic once the Bayesian is retired.
- Live DB confirmed the handoff's root cause:
  - Only row with `section_type='hero-tool'` = `bayesian-ranking-hero-tool_pre_037`
    (id `7cd0408b...`), 14 `source:static` Bayesian fields, `suitable_site_types=[]`,
    `suitable_page_types=["bayesian-ranking"]`, `is_dark_section=t`.

### Correction to the handoff's blast-radius table (caught by re-measuring)
> **CORRECTED 2026-07-21:** the handoff named **two** armed pages and said the
> Bayesian row has **"0 page_components placements."** Live measurement found
> **THREE** pages requesting `hero-tool`, and **one placement**:
> `gamesdesign.co.uk/bayesian-ranking` is **deployed** with the Bayesian hero
> (`page_components` slot `bayesian-ranking-hero-tool`, 9405 B, position 1).
> The handoff's "0 placements / four removed" counted only the 023-scope pages
> (finetuning + orchestration + leopardess); gamesdesign was outside 023 and
> never cleaned. What caught it: querying `pages.sections LIKE '%hero-tool%'`
> across the whole fleet instead of trusting the two named pages.
> **This mattered:** gamesdesign is the ONE page where a Bayesian hero is
> *correct*, so the retire step had to be a supersede (keep active), never a
> delete, and I verified its ranking function survives (separate
> `tool-bayesian-ranking` section, position 3).

### Design settled
- Generic `hero-tool`: llm labels, gated renderer CTAs, optional gated
  anti-fabrication trust stats, single-column intro, no embedded widget.
- Retire Bayesian: `section_type` → `bayesian-ranking-hero-tool`, stay active.
- Both atomic in migration 183. Council gate does NOT apply (SQL/docs are refused
  client-side; scope is platform/internal/pkg Go). No Go change → no image roll.

### Applied
- `183_generic_hero_tool_component.sql` applied clean; NOTICE printed new id
  `0bf81196-e4e7-430b-bd5d-1585703678ae`; all 5 post-conditions passed.
- Post-apply: selector simulation for `section_type='hero-tool'` returns exactly
  ONE row (the generic, score 0.69); new template greps **0** Bayesian strings.

### Next
- Rebuild-based verification: a rebuild is what arms this bug, so a clean rebuild
  is the proof. The two armed pages are already `needs_rebuild`.
