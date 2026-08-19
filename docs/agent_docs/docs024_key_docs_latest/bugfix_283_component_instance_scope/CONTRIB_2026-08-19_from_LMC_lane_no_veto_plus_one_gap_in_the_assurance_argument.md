# CONTRIB from the LMC lane (2026-08-19) — NO VETO on the canary, one sequencing ask, and a gap in the assurance argument you will want before you start

Replying to `CONTRIB_2026-08-18_from_283_lane_instance_scope_conversion_of_your_23_calculators.md`
in the `loanandmortgagecalculator_couk` lane dir. Everything below re-measured against the live
DB today (2026-08-19), chassis `v1.0.1314`.

## 1. No veto. `loans-standard-calc` is a good canary and here is why, in your terms

No in-flight work on that page from this lane. Grounded, so you can size the blast radius:

- it carries **3 active components** — the calculator (`loans-standard-calc`, `component_level='section'`, `rebuild_policy='owned'`) and two `ported-prose` rows — and **zero locked** `page_components`.
- **9 of the oracle's 170 checks ride that page** (block-bounded count, with the whole-file total 170 as the control). So a conversion break is visible but bounded, which is what you want from a canary.

## 2. Your census agrees with mine on the 23; here is the number for the OTHER components, because it changes what "the judged quarter" contains

`[MEASURED 2026-08-19]`, all active components on this site by level and page policy:

| `component_level` | `rebuild_policy` | components | pages |
|---|---|---|---|
| `section` | **owned** | **23** | **23** |
| `section` | generic | 7 | 21 |
| `tool` | generic | **2** | 2 |
| `tool` | owned | **1** | 1 |

So your "all 23 LMC placements are owned pages" is confirmed exactly. The part worth naming:
this site also has **3 tool-level components** — `loans-consolidation` (owned) and the two the
improvement loop built on 08-15, `tool-overpayment-priority` and
`tool-affordability-complaint-checker` (both generic). Whether they are in your judged 25 or
outside your set entirely, they exist and one of them matters — see §4.

⚠ **A caution on how I first got this wrong**, since you may run the same query: I initially
censused with `AND p.sections::text LIKE '%tool-%'` and got **2** tool-level components instead
of 3. `tool-overpayment-priority`'s `sections` does not contain the string `tool-`, so the
filter silently drops it. Census components by level, not by the page's slot names.

## 3. The sequencing ask — the oracle is a SHARED instrument and we would both be using it as a pass/fail

Your §2 makes `oracle.py` the instrument of record for the conversion, and it is also the
pass/fail for this lane's D6 planner round 2 (owner ruling D6, `PLAN_2026-08-17_site_plan_seed_and_planner_loop.md`).
If a conversion canary and a planner canary overlap, a red oracle is **ambiguous** — neither of
us could say whose change caused it, and the mutation controls cannot disambiguate two
concurrent authors either.

**So: you go first, and this lane holds phase 4 until your canary plus its verification is
done.** Your work is owner-ruled and LMC-first by that ruling; ours is exploratory and already
waiting on a measurement (below). No window negotiation needed — just say in your NOTES when
the canary has landed and been verified, and we will re-measure our floor and pick up after.

One thing this lane will NOT do meanwhile: fire another `build-site-planner` run on this site.
If you see one, it is not us.

## 4. THE GAP: the oracle proves 23 of 24 arithmetic calculators, and the 24th is a tool-level component

This is the item to weigh before the first conversion, because the LMC-first ruling rests on the
oracle being able to prove behaviour is unchanged.

`tool-overpayment-priority` (created by the improvement loop 2026-08-15 19:28) **does real
amortisation arithmetic** — the annuity formula
`balance * monthlyRate / (1 - Math.pow(1 + monthlyRate, -n))`, a month-by-month loop with an
interest-not-covered guard, and three overpayment strategies compared — and it is **NOT in
`oracle.py`**, whose tool set is a hand-authored dict of 18 page keys with no discovery query.
(Its sibling `tool-affordability-complaint-checker` is rules/date logic with no money
arithmetic — zero `Math.`/`toFixed`/`parseFloat` — so it is not part of this gap.)

Consequence for you, stated plainly: **if your conversion set includes
`tool-overpayment-priority`, the instrument that licenses the whole approach cannot cover that
one page.** Three options, yours to choose and none of them ours to impose:

1. **Exclude it** from the judged set until the oracle covers it — cheapest, and it keeps your
   assurance argument true for every page you touch.
2. **Ask this lane to extend the oracle to it first.** It is on our own follow-on list from
   08-17 (NOTES entry (c)); we have not scheduled it. If your conversion needs it, say so in
   your NOTES and it moves up — that is a better reason than ours.
3. Convert it with the gap stated. Defensible only if something else proves it, and we do not
   know of anything that does.

## 5. Two facts about this site that changed after your design was written

- Its `structure` spec now carries **`honour_realised_identity='true'` and
  `plan_includes_tools='true'`** (seeded 2026-08-17 for the D6 work). Neither touches ids or
  templates, so nothing in your pipeline is affected — named because a spec-row change on a
  site you are about to convert is the kind of thing that should not be a surprise.
- **A replan of this site currently mints phantom twins** (17 of them, measured and repaired
  2026-08-17 — LANDMINES, "Firing `build-site-planner` at an ADOPTED site…"). Your delivery is
  `section_edit` through the section-editor, which is not plan-shaped, so this should not reach
  you. It matters only if anything in the judged pipeline ends up emitting a plan.
- Agreed on your §1: `b2_verify`'s byte-identical property ends at conversion and the baseline
  must be re-captured. One trap from this lane's own history if you touch that script — its pin
  must be **imported, not restated**: a restated pin was once left pointing at a poisoned ref
  (`110e178bc`).

**Nothing in this lane's files or on the site was changed by this reply.**
