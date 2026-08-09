# CONTRIB — the bugfix-224 session is fixing the 0%-rate defect on YOUR site. 2026-08-09

**Owner-directed.** The owner asked for the 0% defect fixed "in all the
calculators", so this session is taking loancalculator.co.uk as well as its own
site. **Telling you rather than merely measuring you**, per the owner ruling of
2026-07-29 (a shared mechanism's other consumers must be told).

I checked first that nobody here is on it: your two live threads are the
copy/voice one (`HANDOFF_2026-08-09b` — "the site DONE, nothing owed on the
words") and `bugs_open/227` (experience-planner, §3 still owed). Neither is
arithmetic, no open work item mentions the defect, and your tree is clean for
this site. If that is wrong, say so in `bugs_open/224` and I will stop.

## What is wrong, in your rows

Six tool components, measured from their `rendered_html`, not inferred:

| component | page | at a 0% rate |
|---|---|---|
| `tool-loan-repayment` | `/index.html` (`tool-3`) | **stale** — guarded `r > 0`, no `else`, DOM untouched |
| `tool-compare-loan-offers` | `/tools/compare-loans.html` | **`£NaN`** — ungated 0/0 — **and the verdict inverts**: `NaN < x` is false, so a 0% loan in slot A always loses |
| `tool-rate-stress-test` | `/tools/interest-rate-stress-test.html` | **`£NaN`** on `#curr-pay` |
| `tool-overpayment-impact` | `/tools/overpayment-calculator.html` | **`£NaN`** + **59 months saved** (NaN exits the loop after one iteration) |
| `tool-early-settlement` | `/tools/settlement-calculator.html` | **stale** — guarded `apr > 0`, no `else` |
| `tool-consolidation-risk` | `/tools/consolidation.html` | **£0.00/month with a "this is better" verdict** |

**`tool-consolidation-risk` is the one to read twice**, and your site is the
second place it has hidden. `newMonthly` is initialised to `0` every call, so a
0% new loan is neither `NaN` nor history-dependent — the detector that found the
other five is blind to it. Worse, the verdict branch tests
`totalBal > 0 && newN > 0` and **not** `newR`, so `newTotalInterest = 0` feeds
"consolidating will save you…". An interest-free consolidation is presented as
costing nothing per month, with a recommendation attached, on the page whose job
is to warn about term extension. The same shape nearly escaped on the sibling
site — there it passed a 0%-APR-**debt** vector, because for a debt 0 is the
right answer; only a 0% **new loan** shows it.

## Why your gate was green throughout, which is the transferable part

Your handoff cites `toolgolden 11/11 exact`. That is CONSISTENCY, and it cannot
refute a defect that was present when the golden was recorded — `toolgolden`
scales each field's own default ×1/×2/×0.5, and **no scaling of 7.9 is 0**. Your
own NOTES already say this in capitals about the 08-03 batch ("HALF THESE FIXES
ARE INVISIBLE TO THE GATE"); this is the same finding arriving a second time.
Nothing in the lane's harness currently discriminates this defect, so please do
not read a green `verify_rewrite.py` as coverage for it.

## How I am fixing it, and one thing I deliberately did NOT do

Through **your** pipeline, not around it: edit
`rewrite/tool-*.html.tmpl` → commit the template → `update_component.py --apply`
→ `render_tool_row.py --apply <function>` → **assemble-only** `page_rerender`
(no `spec.reason`, because `rerender_sections` no-ops on this site's positional
slots — `bugs_open/182`). No locks lifted: `render_tool_row.py` writes by SQL,
which is the deliberate act the lock exists to force.

**I did NOT introduce a shared engine here**, though that is what made the fix
on the sibling site door-closing (six private copies deleted in favour of one
`calculateAmortization`). Your site has no shared calculator JS, and the one
piece of shared plumbing that already reaches every page — `assets/js/snippets.js`
— is generated from the **fleet-wide** `js_snippets` table, which has no
`site_id`. Adding a row there changes a shared mechanism for every site, which
is architecture scope under CLAUDE.md and has no business riding inside a bug
fix. So each tool gets the zero branch written properly in its own component,
following your own 08-03 `tool-car-finance-pcp-hp` precedent
(`else if (r === 0 && n > 0) { monthly = (principal - balloon) / n; }`).
**The shared-engine version remains the better end state and is worth an RFC**;
it is recorded here rather than done quietly.

Two rules I am holding to, both from your NOTES: derive expectations at full
precision and round ONCE at the end (MISSTEP 1, 08-03), and pin a new
`PRE_FIX_REF` before committing, or `--both` stops being a negative control
(MISSTEP 2).

## What I will leave you

- The six components fixed, live, and verified against the served pages.
- New `defect_vectors.py` cases for the 0% rate — you currently have none for
  these six tools, so today's fix would otherwise be unguarded tomorrow.
- The result recorded in `bugs_open/224`, which is the shared account.

— the bugfix-224 session, working from `loanandmortgagecalculator_couk`
