# 018 FEATURE — specialist-model design critic (screenshot-based taste review)

**Raised:** 2026-07-24, by the owner ("the design of the site and the cards can
be much improved… we could use better, specialist design models").
**Priority 3 of the site-quality automation set.** **Status:** specified, not built.

## The gap

The existing visual-design-auditor checks mechanics (contrast, spacing tokens,
responsive breakpoints, dark-section rules) — it has no **taste**. Nothing
evaluates whether a rendered page looks like quality design by current
professional standards: hierarchy, rhythm, whitespace balance, card
composition, imagery treatment cohesion. "The cards could be much improved" is
invisible to every current check.

## What it is

A per-site design-critique agent:
1. Input: **rendered screenshots** per page (desktop + mobile viewports — the
   screenshot machinery proven in P3), plus the design_intent palette/register
   for context.
2. **Model: Gemini** (owner call 2026-07-24 — try Gemini as the design model
   and see if it works better; revisit specialist alternatives later if not).
   The model prompt asks for a graded critique per page — hierarchy, density,
   card composition, imagery use, distinctiveness vs the named reference
   register — each finding tied to a specific page + viewport + region, each
   with a concrete, actionable change (not "make it better").
3. Output: `site_work_items` (`item_type='design_critique'`) sized so the
   improvement loop / design agents can act on them one at a time.
4. Cadence: post-build + on-demand + after major re-plans. Not per-render.

## Two roles, same model (build later, review first)

- **Review-time** (this feature): critique what was built.
- **Compose-time** (follow-on once review proves out): use the same specialist
  model in the webdesign/site-design-planner path to *produce* better
  compositions, not just complain about them. Kept out of scope here so the
  critic can be shipped without touching the build pipeline.

## Notes

- Findings must be concrete enough to act on mechanically (name the component,
  the property, the direction) — a vague "feels dated" work item is noise.
- Relates: 016 (fidelity-to-brief; this one is taste-with-no-brief-needed),
  the imagery-best-in-class workstream (imagery treatment cohesion), and the
  content-writer style work (copy register is part of perceived design quality).
