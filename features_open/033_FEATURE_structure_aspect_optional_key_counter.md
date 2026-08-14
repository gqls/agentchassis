# 033 — a counter for site_specs structure-aspect opt-in key accumulation

**Filed:** 2026-08-14 · **Lane:** `loancalculator_couk` (planner half 2, migration 407)
**Raised by:** council `architecture` seat (MEDIUM, "worth a follow-up ticket, not a
blocker"), round 1 of correlation `508fe8eb-da12-4de6-833a-20eed750415f`.

## The gap

RFC_022's ruling on accumulated optional surface got its counter (migrations 402-405:
the N=10 budget over `RegisterActionInputSpec` declarations per shared action). That
counter's surface is **action input specs**. It deliberately does not see the OTHER
accumulation surface the same ruling worries about: **per-site opt-in keys on
`site_specs` aspects.**

The `structure` aspect reached **five** such keys on 2026-08-14: `url_shape` (BLD-018),
`honour_realised_identity` / `twin_identity_snap` / `stem_twin_snap` (PLAN-048), and
`plan_includes_tools` (PLAN-049). Every one is individually inert-by-default and
individually reviewed; the accumulation is exactly RFC_022's worry — ten individually
inert opt-in keys are a shared row nobody understands, and nothing counts the tenth.

## The shape of the fix

Same idiom as the 402-405 counter: a sweep that counts distinct opt-in keys read from
each `site_specs` aspect (grep the Go readers for `data->>'<key>'` /
`data ? '<key>'` against `aspect='structure'`, or maintain the census in PLAN-048's
key-inventory line and assert the two agree), with a budget that triggers a review of
the accumulated surface — never of reuse itself (owner 2026-08-14: shared actions are
estate design, not a smell; the same holds for a shared spec row).

## Interim census authority

Until built, the census lives in ONE place: PLAN-048's "key inventory for this aspect"
line in `register/site-plan-and-reconciler.md` (extended in place 2026-08-14 at the
council's reuse seat's direction — do not fork a parallel list).
