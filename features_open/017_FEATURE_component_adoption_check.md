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
