# 311 — a section can be neither REUSED nor CREATED: the selector keys on `section_type`, the writer keys on `function`, and a component with one but not the other is invisible to the first and immovable to the second

**Status:** OPEN. Diagnosed 2026-08-18, `090` verdict **CONFIRMED on the first
iteration** (run correlation `8aa2e283-129f-41d1-93a0-6dcacbbabeae`, intake
`5f0798b3-b16c-4c98-903f-c2ef42ec1b8d`). Not fixed. Three sites are affected today
and the fleet buildout multiplies it.

**Symptom the owner saw:** *"remortgagecalculator.uk left out the actual tools."*
A site whose entire proposition is a calculator shipped with no calculator, and
nothing in the build reported a failure loudly enough to stop the deploy.

## What happens, in order

1. The planner plans a section named **`mortgages-repayment`** on `index`.
2. `SelectComponentByType` (`component_selector.go:186`) looks for a reusable
   component with `WHERE section_type = $1 AND component_level = 'section' AND
   is_active AND forked_from IS NULL`. The component that IS the mortgage
   repayment calculator — `Mortgages Repayment (loanandmortgagecalculator.co.uk)`,
   `function = 'mortgages-repayment'`, active, base — carries **`section_type`
   NULL**. No match. `resolveSectionComponent` returns `not_found`.
3. `not_found` raises **`needs_new_component`**, and the component-creator
   generates a fresh template.
4. `store_generated_component` looks for an existing row to overwrite —
   `WHERE function = $1 AND forked_from IS NULL`
   (`store_generated_component_action.go:229`). **This time it matches**, because
   the row it could not SELECT by `section_type` it can find by `function`. The
   store is now on the regeneration path against **another site's component**.
5. The field-contract guard (`store_generated_component_action.go:397-412`)
   refuses: the new schema drops the 8 field names
   (`button_1, heading_1, label_1..label_6`) that loanandmortgagecalculator.co.uk's
   stored `content_data` is keyed on. **The guard is correct** — overwriting would
   silently empty that site's live sections.
6. Three attempts, all rejected identically (12:51, 14:12, 19:02 on 2026-08-17).
   The section is left with `component_id = ''`, a `needs_section_data` item goes
   to `needs_human_review`, and **the page is built, deployed and served without
   it.**

The two keys disagree, so the component is in a state where it can be neither
reused (invisible to the selector) nor replaced (protected from the writer). No
retry can ever succeed and no amount of regeneration helps.

## Evidence (all measured 2026-08-18 unless stated)

**The colliding row belongs to a different site — the alternative explanation is
excluded.** This is the one question the diagnosis left open (a single site
retrying its own regeneration would look identical in the error log), and it is
answered by joining through `page_components`:

| function | rows | depended on by | requesting site(s) |
|---|---|---|---|
| `mortgages-repayment` | 1 (`b89f91e1`) | loanandmortgagecalculator.co.uk | **remortgagecalculator.uk** |
| `loans-credit-health-check` | 1 (`824e3309`) | loanandmortgagecalculator.co.uk | **loancalculator.co.uk**, **loanzy.uk** |

`content_components` has no `site_id`; ownership is only visible through
`page_components → pages → sites`, which is exactly why the writer's lookup cannot
see that it is about to overwrite a stranger.

**It is live right now, not historical.** `loans-credit-health-check` retried at
18:02, 18:07, 18:10, 18:14, 18:17, 18:21 and 18:25 on 2026-08-18, each time with
the identical 18-field rejection (`agent_error_log`, work item `8c8f5de5`).

**The class is 26 components wide, and that is a floor.** Of 140 active,
base-level `component_level='section'` rows, **26 carry no `section_type`** — every
one invisible to the selector and every one a landmine for the next site that
names its `function`. (89 have `section_type = function`; 25 have a different
one.) A further 79 `component_level='tool'` rows are invisible to this selector by
construction — it filters `component_level='section'` — and are reachable only via
the separate `add_tool` → `deploy_tool_to_site` fork-on-deploy path.

**The irony, and the cheapest possible repair for the pilot:** an active
section-level component `tool-mortgage-repayment` (`section_type` set, 10,760-char
template) has existed since 2026-05-06 and would have been selected had the
planner named that section type. The site did not need a new component at all.

## Why it is worse than one missing section

- **It fails silently at the site level.** The build's own queue recorded three
  failures and a `needs_human_review`, and the site deployed anyway. Nothing
  gates a deploy on "the sections I planned are all present".
- **It scales with the fleet.** The `PLAN_2026-08-18_first_50_build_order` builds
  50 finance sites that will plan overlapping section names by design — the
  second site to want any given tool is the one that breaks.
- **The failure is invisible in the artefact.** A page with 5 sections instead of
  6 looks finished. Only the plan says otherwise.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Make the writer's lookup site-aware and FORK instead of colliding**
   (`store_generated_component_action.go`). When the existing row is depended on
   by a site other than the requester, do not regenerate it — create a new row
   with `forked_from = <existing>.id` and a site-suffixed name, the convention
   `deploy_tool_action.go` already uses (`tool-llm-cost-calculator-…-webdesign-co-uk`).
   Closes the door: one site's component can no longer block another's, ever.
   Platform-scope, shared seam — needs a council round.
2. **Make the two keys agree.** Either the selector also matches on `function`
   when `section_type` is absent, or `section_type` is backfilled and made
   NOT NULL for `component_level='section'`. (1) without this still leaves 26
   unreusable components; this without (1) still lets the writer collide.
3. **Refuse to deploy a page whose planned sections did not all resolve.** The
   `needs_section_data` item already exists and already says `component_id: ""` —
   nothing reads it as a blocker. This does not fix the cause; it stops the cause
   reaching a live site silently, which is the half the owner actually saw.
4. Point the planner at existing section types (a prompt rule). Cheapest, and the
   weakest: it depends on the planner remembering, and "operators must remember X"
   is the shape this list exists to rank last.

## How to verify a fix

Pick a `function` that exists with a NULL `section_type` and is depended on by
site A; plan a section of that name on site B; assert site B ends with a
non-empty `component_id` AND that site A's `content_data` keys are untouched.
**Both halves matter** — a "fix" that lets B overwrite A's row passes the first
check and is the exact damage the guard was built to prevent.

## Related

- `bugs_open/275` — tool-suggester's LIMIT 30 hides most of the library. Adjacent
  (both are "the library is there and the site does not get it"), different cause.
- `bugs_closed/285` — tool-improver rewriting a shared template from a single
  finding. Same shared-row hazard, from the improvement side.
- `bugs_open/283` — interactive components cannot be reused on one page because
  element ids are literal. A different obstacle to the same goal (reuse).
- Register `CLC-*` (component library), `DIR-001` (directory pipeline).

## CONTRIBUTION 2026-08-18 (evening) — a third site, 7 of 7 tool functions lost, and the served artefact measured (from the `loanzy_uk_example_site` lane)

An independent reproduction, same day, different lane, **greenfield site with zero prior
components** — so nothing here depends on aged or hand-edited rows.

`[MEASURED]` `loanzy.uk` (site `55213ded-…`) was built from its domain name alone. Its plan
contained seven tool sections. **All seven `needs_new_component` items ended `failed`**: six at
`store_component` with this bug's exact refusal, one at `generate_template` with
`stop_reason=max_tokens` (a genuine truncation, unrelated). Verbatim, for
`loans-car-finance-calculator`:

> `rejected by pre-store validation: regeneration removes/renames 10 existing schema field(s)
> (button_1, button_2, heading_1, heading_2, label_1, label_2, label_3, label_4, label_5,
> para_1) that dependents' content_data is keyed on — overwriting would strand stored content`

**The incumbent is again `loanandmortgagecalculator.co.uk`** — identified by schema shape, since
it cannot be found by the name the writer used: `SELECT … WHERE input_schema::text LIKE
'%label_5%' AND '%button_2%' AND '%para_1%'` returns *"Loans Car Finance Calculator
(loanandmortgagecalculator.co.uk)"* and three siblings. Searching `content_components` for
`name LIKE 'loans-%'` returns **0 rows**, which is the §-mechanism showing through: the writer's
key and the stored name are not the same thing, so the component is unfindable by the identity
the failing build has.

**The consequence, measured at the served page rather than inferred:**
`https://loanzy.uk/tools/loan-comparison-calculator/index.html` returns **200, 22,600 bytes,
and contains ZERO `<input>` elements** (1 button, 2 inline script blocks). It is a calculator
page with no calculator — prose explaining a tool that is not there, live on the public web,
with no failure visible to a reader. That is this bug's *"built, deployed and served without
it"* clause, with a URL attached.

**Scale, stated because this lane's neighbour is about to multiply it:** the portfolio buildout
plans ~140 finance domains, and the L-family propositions share calculator functions by
construction (`loans-compare-loans`, `loans-overpayment-calculator`,
`loans-settlement-calculator`, `loans-interest-rate-stress-test` all failed here). **Whichever
site creates a function name first owns it, and every later site ships that tool hollow.**
7/7 is the rate for a site whose whole proposition is calculators — not a tail case.

No `090` filed: the mechanism is already proven in this file with a control; this contribution
asserts only an occurrence, its incumbent, and its served consequence, each read directly.
