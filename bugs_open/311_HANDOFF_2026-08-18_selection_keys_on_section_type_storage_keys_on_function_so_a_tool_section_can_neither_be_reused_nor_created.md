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
2. ~~**Make the two keys agree.** Either the selector also matches on `function`
   when `section_type` is absent, or `section_type` is backfilled and made
   NOT NULL for `component_level='section'`. (1) without this still leaves 26
   unreusable components; this without (1) still lets the writer collide.~~
   > **CORRECTED 2026-08-19 (fix lane): REFUTED by reading the resolution chain
   > end-to-end — see the 2026-08-19 contribution below.** Path 1 of
   > `plan_sections` ALREADY matches by `function` (before the selector), so
   > a function-match fallback adds nothing, and the backfill is a no-op for
   > guard-passing rows and actively harmful for guard-dropped ones (it turns
   > a self-healing `not_found` into a `selector_error` → "pass the section to
   > the content writer as-is" silent degrade). The narrow surviving piece:
   > the regen UPDATE now self-heals `section_type = COALESCE(section_type, …)`.
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

## CROSS-LANE 2026-08-19 — `RFC_036` is THIS WALL at the tool level, with the fix already written out, and neither file knew about the other

`architecture_review/RFC_036_tool_function_uniqueness_is_fleet_wide_but_tool_identity_is_per_site.md`
(filed independently, `webdesign_uk` tool-rebuild lane) states the same underlying design fact
this bug is an instance of: **a component's `function` is unique fleet-wide while its identity
is per-site.** As of this note, `grep 311` in RFC_036 and `grep 036` in this file both returned
**0** — two lanes walked into the same wall from opposite ends and neither could see the other.

**They are NOT the same defect, and it matters which you are looking at** [MEASURED 2026-08-19]:

| | this bug (311) | RFC_036 |
|---|---|---|
| `component_level` of the blocked rows | **`section`** | **`tool`** |
| what refuses | the **regeneration field-contract guard**, `store_generated_component_action.go:397-412` | the **unique index** `idx_cc_tool_function_unique` (`WHERE component_level='tool' AND forked_from IS NULL AND is_active`) |
| the writer | `store_generated_component` | `create_tool_component` |
| how it presents | `needs_new_component` **`failed`**, section left with `component_id=''` | a **`complete`** work item (RFC_036 §title) |

Confirmed by querying the incumbents directly: `mortgages-repayment`,
`loans-credit-health-check` and `loans-car-finance-calculator` are all
`component_level='section'`, so `idx_cc_tool_function_unique` — which is partial on
`component_level='tool'` — **cannot** be what refuses them. Conversely
`tool-ab-test-calculator` and `tool-meme-generator` are `tool`-level and are not touched by the
schema-field guard.

**The remedy is the same shape in both, on two different writers, and that is the point.**
RFC_036 §9.3 says: before the INSERT in `create_tool_component_action.go`, look up a library
tool claiming this `function` and, if one exists, **set the new row's `forked_from` to its id** —
because a site-specific build of a tool the library also offers *is* a site copy, which is what
`forked_from` means everywhere else, and `deploy_tool_to_site` already creates exactly that shape
(`deploy_tool_action.go:294-312`). **That is fix candidate 1 of this file, written out by
someone else for the neighbouring writer.**

**So whoever builds either should build both**, or say plainly which writer they left alone.
Fixing `create_tool_component` and not `store_generated_component` leaves 311 live for every
section-level tool — which is the 7-of-7 case on `loanzy.uk` and the calculator page serving
**zero `<input>` elements** recorded in the contribution above. Both are shared-seam changes on
the component write path: **council gate + a chassis roll**, and they would sensibly go as one
submission rather than two.

Not filed as a duplicate and not merged: RFC_036 has an owner direction and a costed path of its
own, and this file has a `090` verdict and three sites. They should stay separate documents that
cite each other, which they now do.

## CONTRIBUTION 2026-08-19 (fix lane) — the mechanism refined, candidate 1 BUILT for the section writer, candidate 2 refuted

Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_311_component_keys/`.

### The refinement: the incumbents are found by function and then THROWN AWAY by the template guard

This file's step 2 says the selector's `section_type` key is why the incumbent is never
reused. That is only the second half of the miss. Read end-to-end, the resolution order in
`plan_sections` is: Path 0 (stored `page_components` identity) → **Path 1: name/function
lookup** (`loadSectionComponents`, `v3_site_actions.go`, pass 2 matches
`content_components.function`, no `component_level` filter) → Path 2 (the selector) → Path 3
(`needs_new_component`). The incumbents (`function='mortgages-repayment'` etc.) ARE found at
Path 1 — and then dropped by `componentInfoFromRaw` → `componentTemplateValid` →
`sectionTemplateValid` (`plan_sections_action.go`), which requires a `</section>` substring.
All three are hand-seeded (`created_from='manual'`, 2026-08-13/15) **tool-shaped templates
ending `</script>`** — measured, none contains `</section>`, all >100 chars — so the loader
reads them as truncated and falls through. THAT drop is what reaches the selector; the NULL
`section_type` then completes the miss. Independent artefact: work item `3d775f99`
(2026-08-15) defers with *"stored component 824e3309 … failed the template guard"* verbatim.

Consequence (correction marked at candidate 2 above): **backfilling `section_type` fixes
nothing** — a guard-passing row already resolves at Path 1 by function; a guard-dropped row
with a backfilled `section_type` would be SELECTED by the selector and then dropped by
`loadSingleComponentSchema`, turning `not_found` (self-healing: raises the work item) into
`selector_error` (silent: *"passing section to content writer as-is"*).

A `090` on this refinement was filed (intake `1306e72c`, run `f1433782`); the run **failed on
infrastructure** — the fleet's Anthropic API returned *"You have reached your specified API
usage limits"* at the verdict step (so did the neighbouring run `6f900e18`) — and the intake
sits at `triaged` for automatic re-dispatch. Stated substitute verification (per the
2026-07-31 ruling): the verbatim guard message on `3d775f99` naming the exact component id;
the measured `</section>` absence on all three templates against the guard's
`strings.Contains` predicate; the resolution chain read function-by-function and cited above.

### What is BUILT (candidate 1, section writer) — commit alongside this note

`resolveStorageIdentity` (`platform/orchestration/actions/component_storage_identity.go`,
register **CLC-020**): `store_generated_component` now takes a dependent census
(`page_components→pages` UNION `site_components`) before treating a function-matched row as
its regeneration target. Foreign dependents → the write DIVERTS to a fresh **base** row
`<function>-<domainSlug(requester)>` with `section_type` = the requested section name —
selector-visible, so the requesting page's rebuild links it and later sites REUSE it instead
of failing (the library heals itself at first collision). Own-site collisions, and callers
with no `input_data.site_id`, keep today's semantics; base+scoped both foreign refuses
loudly. The regen UPDATE self-heals `section_type = COALESCE(section_type, requested)`.
Mutation-proven tests (`component_storage_identity_test.go`): deleting the diversion routes
the foreign-collision test into an uncovered incumbent UPDATE and fails.

Why NOT `forked_from = incumbent` here (vs. RFC_036 §9.3 at tool level, deliberately
different): every section selection path filters `forked_from IS NULL`, so a fork is
invisible to the rebuild that must link it; and the generation is fresh, not the incumbent's
lineage. At tool level the fork escapes a partial unique index and the deploy path links
pages itself — the same wall wants different bricks on the two writers.

### Stated plainly, per the cross-lane note: which writer was left alone

**`create_tool_component` (RFC_036's writer) is untouched by this round.** RFC_036 is OPEN,
its lane proposed no code, and the owner holds a contained interim there. When that half is
built, `foreignDependents` (CLC-020) is the reusable census. Also left open: candidate 3
(nothing gates a deploy on planned-sections-present — the silent-ship half the owner saw;
`check_unresolved_sections` re-arms pages but nothing blocks the deploy), and the three
tool-shaped section-level incumbent rows themselves (mis-shelved; their sites serve stored
`rendered_html`; repair belongs to their lane, through the framework per the RFC_034 bar).

### Verification once a chassis image ships this (the file's own recipe, made concrete)

Re-drive one failed `needs_new_component` (loanzy.uk, `loans-credit-health-check`) and
assert BOTH halves: a new base row `function='loans-credit-health-check-loanzy-uk'`,
`section_type='loans-credit-health-check'`, AND incumbent `824e3309`'s `md5(html_template)`
unchanged with its dependents' `content_data` keys untouched; then the loanzy page links a
non-empty `component_id` after its rebuild. The `COMPONENT_COLLISION_DIVERTED` row in
`agent_error_log` is the queryable demand signal.
