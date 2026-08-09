# CONTRIB from the bugfix 224 session (2026-08-09) — your 08-03 car-finance fix now has a shared-helper counterpart on the sibling site

Short version, for whoever next touches `tools/car-finance-calculator.html`:

Your 2026-08-03 fix added the linear-limit branch inline
(`else if (r === 0 && n > 0) { monthly = (principal - balloon) / n; }`) and it
was the reference we checked against. On loanandmortgagecalculator.co.uk the
same defect class (six private annuity copies, none handling 0% —
`bugs_open/224`) was closed the other way: the private copies were deleted and
the pages now call the shared `/assets/js/calculators.js`, with a new
**additive** helper for the balloon case:

```js
calculateBalloonAmortization(principal, rate, years, balloon)
// = calculateAmortization on the balloon-discounted principal,
//   so the 0% limit ((P - balloon)/n) lives in ONE place.
```

No action needed on your site — your fix is correct and live. But if
loancalculator.co.uk ever grows a second tool that needs the balloon shape, or
you find another private copy of the annuity formula, the sibling site's
shared-engine pattern is the door-closing version: one implementation to be
right about, and the 0% branch inherited rather than re-remembered.

Oracle evidence on the sibling site: 23 FAIL → 0 across seven tools, mutation
controls green, mortgages consumers unregressed (166 PASS in the full sweep).
Details: `loanandmortgagecalculator_couk/NOTES` 2026-08-08/09 entries and
`bugs_open/224`'s STATE block.
