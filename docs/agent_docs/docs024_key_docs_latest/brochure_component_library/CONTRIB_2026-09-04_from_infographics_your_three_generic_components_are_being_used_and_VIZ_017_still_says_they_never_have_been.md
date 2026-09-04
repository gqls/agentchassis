# CONTRIB 2026-09-04 → `brochure_component_library`, from the new `infographics` lane: **your three generic components went from zero to eighteen instances in four days, and VIZ-017 still says no page has ever been built with any of them**

**From:** `docs/agent_docs/docs024_key_docs_latest/infographics/` — opened today at the owner's
direction. This is good news plus one stale line, and the stale line has already cost something.

---

## 1. VIZ-017's `verify-later` question is ANSWERED, and the answer is yes

Your entry (2026-08-24) closes: *"**Live, but UNEXERCISED: no page has yet been built with any of
them**"*, with a `verify-later` of *"whether a real build selects any of them"*.

`[MEASURED 2026-09-04]`:

| component | instances | sites | first use | last use |
|---|---|---|---|---|
| `checklist` | **9** | 3 | 2026-09-02 | 2026-09-04 |
| `comparison-table` | **7** | 4 | 2026-09-02 | 2026-09-03 |
| `period-calendar` | **2** | 2 | 2026-08-31 | 2026-09-04 |

**Every first use postdates the entry.** All `build_status='deployed'`. Verified at the served
artefact with a control, not at the row: `websitepromotion.co.uk/blog/website-launch-promotion-checklist.html`
→ HTTP 200, 80,415 B, `checklist__item` / `__body` / `__footnote` markup, 48 `<li>`; an invented
sibling path on the same domain → **404**.

Together with `mechanism-flow` (14/10), `evidence-chart` (10/5) and `evidence-timeseries` (3/3), the
set is **45 instances across 17 domains** and accelerating: ≤3/day through August, **4 on 09-02, 15 on
09-03, 9 by midday 09-04**.

**Your reuse discipline is visible in the data.** `comparison-table`'s named-cells decision — a
missing value renders empty *in place* rather than shifting the grid — is exactly what is holding up
across three seotools comparison pages built by a planner nobody briefed. And dropping the fourth
"steps" component because `mechanism-flow` already existed looks right: `mechanism-flow` is the
most-used of all six.

I have corrected VIZ-017 visibly (strike-through + dated block, per the register-status landmine)
rather than only noting it here. **Correct it in your own docs too if you carry the claim** — I have
not touched anything outside the register entry.

## 2. Why the stale line was not free

Register status lines are read by council seats as ground truth, and this one was load-bearing in a
live owner-facing question.

Since 2026-08-31 the owner has asked twice for explanatory imagery to replace explanatory copy, and
four consecutive sessions across three lanes investigated *"why does the estate have no
infographics"*. Every one of them measured `site_plan_imagery.kind='infographic'` — **1 row in all
fleet history** — and reported that the capability is undriven. **None ran a query against your
components**, and VIZ-017 told anyone who did look that they had never been used.

The structural reason, which is not anyone's carelessness: **your components are named for their
shape, never their function.** No grep or SQL containing the word "infographic" can reach
`comparison-table` or `mechanism-flow` — not in the DB, not in Go, not in the docs. I have filed that
as a landmine (`LANDMINES.md`, *"How many infographics does the estate have?"*) with both query arms,
and it cross-references your register file as the source of truth for what route B contains.

**The practical ask:** when you add a structured component, the landmine's route-B list and
`infographics/RUNBOOK_infographics.md` §1 both carry a hand-maintained name list. A component nobody
adds makes the census read low — the same failure one level up. Ping this lane, or edit either list
directly.

## 3. Where the boundary sits

Written up in `infographics/PLAN_2026-09-04_infographics.md` §4:

- **You build route B's components.** This lane does not build components and has said so in its
  will-not-do list.
- **This lane owns the selection rule** — for a given explanatory need, prose vs a generated picture
  vs one of yours — and routes work to you.
- **The contradiction I am carrying, which touches VIZ-005:** the live planner prompt (mig 718) tells
  the planner to use a *diffusion* `infographic` **"for numbers"**, while IMG-046 says a diffusion
  infographic *"must never carry real numbers"* and VIZ-005 draws exactly your boundary — generated
  images explain, code-rendered output states. **VIZ-005 is `designed, not built (the rule is stated;
  nothing enforces it)`, and the live prompt is currently the counter-example.** I have flagged it in
  IMG-046 and not resolved it; which side is wrong is an owner decision.

## 4. One question for you, asked rather than assumed

`[UNMEASURED]` **How does a component become reachable to the planner for an *explanatory need*, as
opposed to being chosen because a page type implies it?** VIZ-017 credits PLAN-053 /
`component_expresses` with making the planner able to *see* what a component expresses. I have not
read that derivation and will not assert anything about it. If the 09-02→09-04 inflection is
`component_expresses` finally biting, that is your result and it should be recorded as one — I have
three candidates for that inflection (mig 718, your components landing, the 641 planner work) and
they are **confounded by date**, so I am naming them as a question, not testing them.

— the `infographics` lane, 2026-09-04
