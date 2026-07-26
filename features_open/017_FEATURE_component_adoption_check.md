# 017 FEATURE — component-adoption check (registered-but-never-selected; interactivity promised, none used)

**Raised:** 2026-07-24, by the owner, reviewing fundamentallyai.com ("I can't
see the interactive components"). **Priority 2 of the site-quality automation
set.** **Status:** specified, not built.

## The gap

Five new interactive components were built and registered
(`hero-card-carousel`, `image-hover-card-grid`, `swipeable-insight-carousel`,
`stat-band`, `people-feature-block`) and the planners load the registry
dynamically (`load_component_library` → `AvailableFunctions`) — but the
**planner never selected any of them**; every instance on fundamentallyai was
hand-placed. Nothing detects either failure mode:
1. **Dormant components** — registered, active, never chosen by any planner on
   any site (the exact shape of the dormant-agents inventory, `bugs_open/044`,
   applied to `content_components`).
2. **Interactivity mismatch** — a site whose `design_intent`/brief promises
   carousels/interactive elements, whose built pages use none (a mechanical,
   cheap special case of 016).

## What it is

Two cheap mechanical checks (no LLM):
1. `check_component_adoption` (fleet-level report, dormant-agents style):
   for each active `section` component, when registered vs. count of
   `page_components`/`site_plan_sections` referencing it, split hand-placed vs
   planner-selected (planner-selected = appears in `site_plan_sections`).
   Flags never-planner-selected components older than N days. CAVEAT from 044:
   usage_count on content_components is dead (0/162, bugs_open/060) — count from
   the join tables, not the counter.
2. `check_interactivity_promised` (per-site discovery check): `design_intent`
   mentions carousel/interactive/swipe/chart (word-class list, not exact
   strings) AND no page_component maps to a component with matching
   `semantic_tags` → work item.

## Why it matters beyond this site

A registered component the planner never picks is inert inventory — the
"prompt-seam" landmine this workstream recorded on day one. The likely root
cause (to confirm when building): the planner LLM chooses from
`AvailableFunctions` but with no steering toward new/interactive types, it
defaults to the familiar ones it has seen in every prior plan. The check makes
that drift visible; fixing selection (prompt weighting, `suitable_site_types`
gating, or exemplars) is follow-on work the check will measure.

---

## Contribution from the `hero_tool_component_045` workstream (2026-07-26) — a third check, the inverse of check 1

**Why it is landing here rather than in its own file:** this was fix candidate 4 of
`bugs_closed/045`, deliberately deferred because it is Go (needs an image roll) and was to
ride with `bugs_open/039`, the sibling branch of the same selector. **039 has since closed**,
so it had no home. It is the same report, over the same table, from the same direction as
check 1 — so it belongs to whoever builds 017, not to a competing feature file. No work
started; observation and spec only.

**3. `check_generic_section_sole_candidate` — a generic section name whose only candidate is
product-specific.**

Check 1 asks *"which components does nobody ever select?"*. This asks the mirror question:
*"which section names have so few candidates that selection is not really happening?"* Both
are cheap `content_components` reports and share the same shape, so build them together.

The failure it would have caught is `045` itself, on day one. A page plan asked for the
generic section `hero-tool`; the active library held **exactly one** component with
`section_type='hero-tool'`, hard-wired to a Bayesian ranking product with 14 `source:static`
labels; `SelectComponentByType` has **no minimum-score threshold**, so a sole candidate always
wins however badly it fits. Result: an LLM cost calculator shipped with *"Start Ranking Free"*
and *"Try the Bayesian Ranker"*, on live customer sites, for weeks. Every component behaved
correctly — the library was simply missing a neutral option, and nothing anywhere says so.

**Cheap version (the one worth building):** for each distinct `section_type` requested by any
`pages.sections` / `site_plan_sections` row, count active `component_level='section'`
candidates. Flag where the count is **1** and that sole candidate looks product-specific.
Two mechanical tells, no LLM needed:
- it carries `source:static` **value** fields (frozen text the page's own `content_data`
  cannot override — that is the whole mechanism), especially several of them; and/or
- its `suitable_page_types` names a single narrow product (`["bayesian-ranking"]`) while the
  section name requesting it is generic.

**Two facts to build against, both measured in 045 and worth not re-deriving:**
- **A section name resolves by `content_components.section_type`, NOT by `function` or
  `name`** — and `page_components.slot_name` is written from the *component's* `function`,
  so the slot name on a built page is not the section name that was requested. Reading
  `slot_name` as the request is a real trap: on `gamesdesign.co.uk/bayesian-ranking` the plan
  says `"hero-tool"` and the slot says `bayesian-ranking-hero-tool`.
- **`usage_count` is dead** (0/162, `bugs_open/060`) — same caveat as check 1; count from the
  join tables.

**Scope note so this stays cheap:** a sole candidate is not a defect by itself. Most sections
have one sensible component and that is fine. The signal is *sole candidate* **plus**
*frozen product-specific vocabulary* — that pairing is what makes the request unanswerable
rather than merely unambiguous. Report it, don't block on it.
