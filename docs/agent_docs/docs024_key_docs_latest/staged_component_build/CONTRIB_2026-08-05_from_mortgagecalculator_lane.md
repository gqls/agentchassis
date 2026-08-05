# CONTRIB — 12 mortgagecalculator.co.uk tools for your authoring backlog, goldens already captured

**From the `mortgagecalculator_couk_adoption` lane, 2026-08-05.** Your handoff says the
owner gave you the PLAN-authoring backlog (36 of 49 tools). This adds 12 tools that are
being recreated today — with the arithmetic half of their contracts **already captured**,
so authoring their fences is cheaper than the census suggests.

## What exists, where

- **`../mortgagecalculator_couk_adoption/acceptance/GOLDEN_2026-08-05_original_tools.json`**
  — all 12 original hand-built tools, captured on **four** vectors (defaults/double/half/
  **asym** — the new one; see below), answers included. This is the authority for what each
  rebuilt tool must compute. Emitted by `loancalculator_couk/toolgolden.py`.
- **`../mortgagecalculator_couk_adoption/acceptance/criteria/fact-finder.criteria.json`**
  — the one fence emittable from the originals (4 checks, 12 assertions), ready for a
  `doc_plans` row (`subject_type='tool'`, `subject_key='game-fact-finder'`) per your §9.
- **The other 11 refused emission for ONE uniform reason: every original calculate button
  has no `id`.** Their rebuilds (running now via `tool-recreation-handler`) should carry
  id-complete markup; re-emit from the rebuilt pages once they settle and the fences drop
  out mechanically — `toolgolden.py --emit-criteria <dir> <new urls>`.

## The harness changed today — read this before you re-capture anything

`toolgolden.py` gained a fourth **asym** vector (per-field factors) because the uniform
vectors falsely convict RATIO tools — investor.html's LTV/yield calculators are
scale-invariant, so "output identical across vectors" meant *correct*, not inert.
Non-regression proven against your sibling lane's `GOLDEN_2026-08-03b`: **11/11 MATCHES**
on the shared vectors. Pre-asym goldens still load, compare and emit (presence-guarded).
Full detail: `loancalculator_couk/NOTES` 2026-08-05 entry, TL-038 addendum.

## Known gaps, stated so they don't become surprises

- investor.html's golden covers the **yield calculator only** — the harness presses one
  button per page and the LTV half sits behind the second. A fence authored from it
  should either accept that scope or add a second-press check by hand.
- The goldens assert *"the arithmetic has not moved since capture"*, never *"the
  arithmetic is correct"* — TL-038's first landmine, unchanged.
