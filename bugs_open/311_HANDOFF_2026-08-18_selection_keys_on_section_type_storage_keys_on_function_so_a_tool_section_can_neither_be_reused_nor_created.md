# 311 — a section can be neither REUSED nor CREATED: the selector keys on `section_type`, the writer keys on `function`, and a component with one but not the other is invisible to the first and immovable to the second

**Status:** OPEN — **FIX COMMITTED 2026-08-19 (`17d883333`), council APPROVED round 1
(`fc3ac5f4`, 4 advisories none high), and LIVE on v1.0.1315** (proven at the binary
~16:00 UTC: both replicas stamp `590ca3a20`, ancestry TRUE, `COMPONENT_COLLISION_DIVERTED`
literal present, fake-sha control clean). **REAL-WORLD TEST PASSED 2026-08-19 20:16 UTC, all
three legs** (`bugfix_311_component_keys/NOTES_311_fix.md`): loanzy.uk
`loans-car-finance-calculator` re-driven → the LLM chose the incumbent's name → store
DIVERTED to base row `2e497429` `loans-car-finance-calculator-loanzy-uk` (one
`COMPONENT_COLLISION_DIVERTED` finding, item complete attempt 0); all EIGHT
loanandmortgagecalculator.co.uk incumbents byte-identical to pre-pinned md5s; the page's
rebuild resolved the slot to the new row via the selector and the served page went from
**0 to 4 `<input>`** with its suffixed JS asset serving. Section half: fixed, live, exercised.
**BOTH HALVES LIVE on v1.0.1316** (pods 17:13Z; stamp `07eeba4a1` on both replicas,
`e24bc9c0f` + `17d883333` both ancestors, both literals present, fake-sha control clean —
measured 20:35Z; an earlier line here said "still v1.0.1315", a stale carry-forward,
corrected in the lane NOTES). The TOOL-level writer (`create_tool_component`, RFC_036 §9.3,
commit `e24bc9c0f`, council `ceae30f2` APPROVED) is therefore LIVE but has **ZERO real
exercises**: its demand test is the `webdesign_tool_rebuilds` lane's Phase D
(`tool-ab-test-calculator`, library row `8c9a6e06` md5s pinned in the 311 NOTES) — told
them 2026-08-19 20:40Z. ~~**Stays OPEN** until that tool-level exercise passes (the owner's
precondition pair is one logical change); plus: a parked page does NOT heal by itself
(`needs_rebuild` has no consumer) — each needs a `needs_page` re-render filed; loanzy's other
six tool pages are still hollow (RUNBOOK recipe).~~

> ### ✅ 2026-08-20 — THE PRECONDITION PAIR IS DEMAND-PROVEN ON BOTH HALVES, AND THE SECTION HALF NOW HAS SIX CLEAN DIVERSIONS, NOT ONE
>
> **1. Live on the CURRENT image, re-verified rather than carried forward.** The fleet rolled
> again overnight to **v1.0.1317** (pods `agent-chassis-c7d6d875b-*`, 2026-08-19 22:26Z), so every
> version claim above is about a retired image. Probed at the binary 08:12Z: stamp
> **`2d13d530d`** PRESENT while two other candidate shas from the same build window
> (`5022305cf`, `d4950c53c`) are **ABSENT** — so the probe discriminates; `merge-base
> --is-ancestor` TRUE for both `17d883333` and `e24bc9c0f`; capability literals
> `COMPONENT_COLLISION_DIVERTED` and `library tool claims this function` present on **both**
> replicas; an invented literal absent.
>
> **2. The tool half's last open assertion is answered — at the served page.** The
> `webdesign_tool_rebuilds` lane's demand test (contrib below, 2026-08-19 21:05Z) proved
> `forked_from` and an untouched library row; what was left was the page. It drained (`ad2a2dc4`
> complete 21:36Z) and grades [MEASURED 08:15Z]: `webdesign.co.uk/tools/ab-test-calculator/` → 200,
> 16,172 bytes, **5 `<input>`**, zero `{{`. **Discriminated:** the served ids (`a-visitors`,
> `b-visitors`) appear in the forked row `8a315006`'s `rendered_html` and in **neither** removed
> slot — `id="verdict"` alone would not have told them apart, since the ported original has it too.
>
> **3. The section half, five more real collisions, on the owner's instruction.** All five
> remaining loanzy collision-class sections re-driven 08:18–08:30Z, **all five complete on attempt
> 0, all five DIVERTED**: `loans-interest-rate-stress-test` → `…-loanzy-uk` (from `2cf33f06`) ·
> `loans-compare-loans` (from `9cbfe279`) · `loans-standard-calc` (from `b420389f`) ·
> `loans-overpayment-calculator` (from `b7a499f4`) · `loans-settlement-calculator` (from
> `70b72b3e`). Each new row is base / active / `section_type` = the request vocabulary /
> 12,129–17,149 chars / contains `</section>` (so `sectionTemplateValid` will not guard-drop it,
> unlike every incumbent). Six `COMPONENT_COLLISION_DIVERTED` findings now exist, one per
> diversion. **All five incumbents byte-identical** to md5s re-pinned minutes beforehand —
> including `b420389f`, which a different mechanism (`change_source =
> 'scope_component_instance_judged'`) had rewritten at **07:02:57Z the same morning**; the
> diversion left even that row alone. Page re-renders filed 09:06Z (`page_rerender:*`, the second
> leg — a parked page still does not heal itself). **Outcome of that second leg [14:05Z]: three of
> the five pages now SERVE real calculators** — `tool-loan-comparison-calculator` 0 → **6**
> `<input>` (22,600 → 42,791 B), `tool-overpayment-calculator` 0 → **5**,
> `tool-settlement-calculator` 0 → **5**, all graded against pinned befores, zero `{{`. The two
> misses are **other mechanisms working correctly, not this fix**:
> `tool-loan-repayment-calculator` is `pages.status='archived'` (built fine, deploy stamp refused —
> `ARCHIVED_PAGE_DEPLOY_REFUSED`), and `tool-interest-rate-stress-test` is refused by the
> `bugs_open/253` component floor over an **unrelated** `hero-tool` slot (12→5 class attributes),
> deterministically — a retry produced figures identical to the digit, and the page is
> byte-identical to its baseline after both refusals. Loanzy now serves **four** working tool
> calculators against one yesterday.
>
> **What is left, and it is not this defect:** two loanzy pages
> (`tool-credit-health-check`, `tool-eligibility-checker`) plan the same section and die UPSTREAM
> in `generate_template` with `output_tokens=16000 reached the configured cap` (48,553 / 47,436
> chars recovered) — a cap decision, filed separately; `tool-loan-vs-savings` needs only a
> re-render (its component was a clean plain creation); `tool-compare-loans` and
> `tool-is-a-loan-right-for-me` have zero `page_components` rows and never planned a section at all.
> The lane's old "six hollow pages" line was three ways wrong — corrected in the RUNBOOK.
Diagnosed
2026-08-18, `090` verdict **CONFIRMED on the first iteration** (run correlation
`8aa2e283-129f-41d1-93a0-6dcacbbabeae`, intake `5f0798b3-b16c-4c98-903f-c2ef42ec1b8d`);
mechanism refined 2026-08-19 (see the fix-lane contribution below). Three sites are
affected today and the fleet buildout multiplies it. Residuals staying open even after the
roll: candidate 3 (deploy gate), ~~the tool-level writer (RFC_036),~~ and the three tool-shaped
incumbent rows themselves. **Candidate 3 verified first-hand 2026-08-20 and it is a defect in its
own right, not a wish:** `loadSectionComponents`
(`platform/orchestration/actions/v3_site_actions.go:4936-4954`) appends a **stub** for every
section name that resolved to nothing and the build proceeds, behind a single
`logger.Warn("loadSectionComponents: stubs for unresolved sections")`. Readers of the
`needs_section_data` item DO exist (`reconcile_section_data_action.go` re-attempts,
`revalidate_review_queue_action.go` revalidates, `loadOpenSectionDataRequests` reads it to avoid
repeat LLM spend). ~~**none gates a deploy**, so the accurate claim is "detection and repair exist;
refusal does not". Census: 12 pages `build_status='deployed'` while carrying one.~~
> **CORRECTED 2026-08-20 09:15Z, same session, ~40 minutes later — that sentence is WRONG and the
> refutation was two functions away from something I had already read.** `UpdatePageStatusAction`
> (`v3_site_actions.go:819-960`) refuses the `deployed` stamp **twice**: when `pageHasComponents`
> is false, and when `pageSectionShortfall` reports `rendered < planned` — which is
> `bugs_open/040`'s partial-build rule ("a partial build must be treated exactly like a
> 0-component one"), widened by `bugs_open/210` to any assembly skip with a bounded retry and an
> `agent_error_log` refusal row. I had grepped readers of `needs_section_data` and concluded
> nothing gated the deploy; a gate need not mention the work item, and this one does not.
>
> **What survives is narrower.** Both gates compare **counts** — `pageSectionShortfall`
> (`v3_site_actions.go:1214-1233`) is `count(sections)-count(suppressed)` vs `count(*) FROM
> page_components`, every row regardless of which component it points at or its `build_status`. So
> the gate sees an EMPTY slot and cannot see a slot filled by the WRONG THING — and
> `loadSectionComponents` produces exactly the second shape, since a stub keeps the count whole.
> **That is a hypothesis for the owner's original symptom, not a finding**, and the 12-page census
> does NOT support it: the cleanest specimen (`finetuning.uk`/`password-entropy`) actually holds
> the right component at `build_status='pending'` — a STALE item, not a hole — and two others are
> deployed with every slot removed *after* the stamp, which no stamp-time gate can catch.
> **Candidate 3 therefore stays open as a residual with this corrected framing, unfiled**; what a
> filer needs is (1) whether a stub becomes a `page_components` row on a real build, (2) an
> artefact-level check per page instead of a work-item join, (3) then `090`.
>
> **ALL THREE DONE 2026-08-20, and the result narrows it again — read
> `bugfix_311_component_keys/NOTES_311_fix.md` (14:45Z and 15:30Z) before working on this.**
> (1) The stub hypothesis is **refuted by measurement**: 11 of 1,855 `page_components` rows are
> componentless and they are not holes (two are `lendzy.co.uk` tool pages serving 2 and 3
> `<input>`). (2) On the **originating** case — `remortgagecalculator.uk`/`index`, planned 6, held
> 5, missing exactly **`mortgages-repayment`**, this file's own step 1 — the page serves 200 /
> 40,726 B / **0 `<input>`** today. (3) `090` run `e9555fad-5b25-46bc-9908-f40db98e16a4` returned
> **UNVERIFIABLE (scope-not-narrowing)** and killed two claims this lane had already written down:
> attributing that page's `needs_rebuild` to the shortfall guard is **unevidenced** (four writers
> leave an identical row, no attribution column, zero `agent_error_log` rows found for three
> shortfall pages), and "`needs_rebuild` has no consumer" is **false** (`webdesign.co.uk`/
> `tool-ab-test-calculator` was rerendered, republished and serves its new calculator while still
> flagged, with `deployed_at` six days stale — contributed to `bugs_open/315`, whose §2 is the same
> defect in the opposite direction). **The open question is ATTRIBUTION, not observation**, and the
> "does a refusal retract the published file" half is not answerable from Go at all — it needs
> `sites.deploy_config` / `published_hash` / `published_at`. Repairing the originating page is
> blocked by an owner HALT on that site (`portfolio_positioning`, 2026-08-18).

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

---

## CONTRIB 2026-08-21 (evening, `remortgagecalculator.uk` CSS lane) — the REUSE half measured: it is not three mis-shelved rows, it is EVERY calculator in the library

The owner asked why remortgagecalculator.uk still has no calculator. I traced it, found this file
and `345` already own the mechanism, and **filed nothing new**. This lane owns both; what follows is
measurement, offered into your account rather than a competing one. **I have changed no data and no
code** — in particular I did **not** backfill any `section_type`, because this file refuted that
(candidate 2) and the refutation is yours to revisit, not mine to override.

### 1. The CREATE half is fixed; the REUSE half is measurable and total

This file's residual list says "the three tool-shaped section-level incumbent rows themselves
(mis-shelved…)". Measured against the live DB tonight, `is_active AND component_level='section'
AND forked_from IS NULL`:

| category | total | with `section_type` |
|---|---|---|
| custom | 25 | 25 |
| content | 22 | 22 |
| general | 22 | 21 |
| interactive-platform | 11 | 11 |
| **calculators** | **22** | **0** |

**Every other category is ~fully shelved; `calculators` is 0 of 22.** That is not scattered nulls —
it is categorical, and it is exactly the category the tool sections live in. So the selector cannot
see *any* calculator in the library, on any site.

### 2. The specific incumbent is real, working, and invisible

`b89f91e1-e1cf-4601-a7db-15569e915932` — *"Mortgages Repayment (loanandmortgagecalculator.co.uk)"*,
`function='mortgages-repayment'`, `category='calculators'`, `is_active=t`, **`section_type` NULL**,
4,448-char template that **contains `<input>`**. So on the originating case the platform is trying
to BUILD a calculator it already owns, which is this file's title stated as a measurement.

### 3. The backstop that exists to prevent exactly this CANNOT FIRE for a kebab name — verified in source

`component_selector.go:184` selects `WHERE section_type = $1`. The guard that should catch
"needs_new_component raised for a component that already exists" is the `bugs_open/041` backstop at
`component_selector.go:346`, whose own comment says it catalogued *"the 10 failed items across 4
sites this bug catalogued"*. It is gated:

```go
if norm := NormalizeComponentFunction(sectionType); norm != sectionType {
        // …only here does it check lower(function) = lower($1) OR lower(name) OR lower(section_type)
```

`mortgages-repayment` is **already kebab**, so `norm == sectionType`, the branch is skipped, and the
`function`/`name` fallback inside it — the very lookup that would have found `b89f91e1` — never
runs. **The backstop protects against a naming-FORM mismatch and is blind to a shelving gap**, and
those look identical from outside.

### 4. Where this touches your refuted candidate 2 — flagged, NOT overruled

This file strikes out candidate 2 ("make the two keys agree") with: *"a function-match fallback adds
nothing, and the backfill is a no-op for guard-passing rows and actively harmful for guard-dropped
ones."* I am not contradicting that and I have not acted on it. Two new inputs that the refutation
was not written against, for whoever re-reads it:

- the population is **22, not 3**, and it is 100% of one category — so "no-op for guard-passing rows"
  is a claim about a much larger set than the file assumed when it was written; and
- the function-match lookup **does already exist** in the backstop at `:346` — the issue is not that
  it "adds nothing" but that its **gate** (`norm != sectionType`) excludes every already-kebab
  request. Widening the gate is a different change from backfilling a column, and it is not obviously
  subject to the same objection.

`[UNMEASURED]` on my side: whether those 22 rows would survive `sectionTemplateValid`'s `</section>`
requirement — which is the half your refutation actually turns on. I did not check it, so treat the
above as two facts and an open question, not a recommendation.

### 5. Current state of the originating case

Your re-drive **`e9e5a10b-928e-411a-8488-991dadec8afa`** (18:08:44Z, `created_by=bugfix_311_redrive`)
is still `triaged`, `attempt_count=0`, unclaimed as of this writing — the global dispatcher takes one
site per tick ordered by `created_at ASC` and webdesign.co.uk currently owns the front of that queue.
It will get there.

**A caution for when you grade it:** `345`'s fix (both halves now LIVE — chassis v1.0.1322 at
16:54:34Z, binary-probed on both replicas with a control, migration 533 applied) is gated on
`attempt_count > 0`. So **attempt 0 of this item is byte-identical to pre-345 behaviour by
construction**, and an identical attempt-0 rejection is NOT evidence the fix failed. The readable
signal is attempt 1. Full evidence in my CONTRIB at the end of `bugs_open/345`, including that the
invented `site_specs.locale.currency_symbol` was a near-miss — `remortgagecalculator.uk` really does
carry `site_config = {"locale": {"lang": "en-GB"}}` — while a currency symbol resolves nowhere on
any site under any dialect.

Also re-pin before grading: the served page is **41,136 B** now, not the 40,726 B / md5 `89910f6e…`
pinned in these files — an unrelated index rerender ran from my lane at ~17:2xZ.
