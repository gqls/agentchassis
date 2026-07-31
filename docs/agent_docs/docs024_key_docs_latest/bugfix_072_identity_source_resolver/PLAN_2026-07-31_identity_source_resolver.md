# PLAN — `bugs_open/072`: the plan_sections resolver cannot see the canonical identity store

**Started 2026-07-31.** Thread: "bugfix 9". Bug: `bugs_open/072_HANDOFF_2026-07-25_contact_info_reads_flat_identity_keys_the_writer_nests.md`.

> **Number warning.** `072` is one of this repo's ambiguous numbers. The OTHER 072
> is `bugfix_072_component_css` (component markup without CSS) — **closed and
> live on v1.0.1171**. This directory is the *contact-info / identity source*
> case. Resolve by slug, never by number.

## What the bug file said, and what I found instead

The filed diagnosis is that `contact-info`'s `input_schema` sources four fields
from **flat** `site_specs.identity.{email,phone,address,hours}` while
`domain-research-classifier` writes them **nested** under `identity.contact.*`;
so the path never resolves and `email`'s `on_missing: needs_human_review`
withholds the whole section. Its remedy candidates were (1) repoint the schema at
the nested path, or (2) teach the resolver a nested fallback.

**The path mismatch is real. Both remedies fix nothing.** Measured 2026-07-31
against the live DB (query in `RUNBOOK`, §1):

| site | flat `identity.email` | nested `identity.contact.email` | `sites.email` |
|---|---|---|---|
| 7 working sites | populated | populated (6 of 7) | populated |
| gamesdesign.co.uk | — | **empty** | — |
| loancalculator.co.uk | — | *no `contact` key at all* | — |
| oufe.com | — | **empty** | **populated** |
| relojistas.com | — | **empty** | — |
| robot-hands.com | — | **empty** | **populated** |
| vetcomparison.uk | — | **empty** | **populated** |
| vonc.com | — | **empty** | **populated** |
| webdesign.co.uk | — | **empty** | **populated** |

The nested sub-object exists on 14 of 15 sites but its **values are null/empty on
exactly the 8 sites that fail**. So repointing the schema at
`identity.contact.email`, or adding a nested-only resolver fallback, resolves on
**0 of the 8 broken sites**.

The bug file's discriminator table (flat email present ⟺ contact-info rendered)
is correct as a measurement. Its causal reading is inverted: the sites that render
are the sites that have contact data *at all*; the rest have none in
`site_specs` — flat or nested.

**Where the data actually is:** `sites.email`, populated on **12 of 15** sites,
including **5 of the 8** that fail. The bug file records this in passing without
drawing the conclusion — *"the owner's phone had been written only to
`sites.phone`, which no component reads."* That sentence is the actual root cause.

## Root cause

`sourceResolver` (`platform/orchestration/actions/plan_sections_action.go`) can
resolve `site_specs.*`, `site_assets.*`, `pages.*`, `config.*` and `query.*`.
**No branch reads the `sites` row's own identity columns** (`email`, `phone`,
`contact_address`, `company_name`, `tagline`, `logo_text`, `logo_url`).

This is the **third instance of one class**, and the other two are already fixed
in the direction I am proposing:

| path | reads `sites.email`? | how |
|---|---|---|
| full writer render | **yes** | `loadSiteDataFull`, `render_site_components_action.go:337` |
| light section rerender | **yes** (fixed) | `buildRerenderBaseData`, `rerender_page_sections_action.go:590` — *"We now prefer the column … making both render paths agree"* (`bugs_open/006` §B) |
| **`plan_sections`** | **no** | ← this bug |

So the fix is not a new mechanism. It brings the one remaining path into line
with the two that already agree, on the store the platform already calls
canonical.

## Decision: what to build

**A bounded, explicit fallback chain in the `site_specs` branch of
`sourceResolver.resolve`, consulted only after the literal path misses.**

1. the writer's nested shape — `identity.<leaf>` → `identity.contact.<leaf>`
2. the canonical sites row — `identity.<leaf>` → `sites.<column>`

**Why this shape, and not the alternatives.**

- *Not* repointing the component schema (bug's candidate 1): fixes one component,
  0 of 8 sites, and leaves the next component to make the same mistake. The user's
  standing instruction is a framework fix over the individual case.
- *Not* a blind deep search of the aspect for the leaf name: two same-named keys
  at different depths become ambiguous, and the result depends on map iteration
  order. Enumerated instead — which is also the shape the neighbouring
  `site_assets` branch already uses (`imageryplan.ImageRoleForPath`, "literal key
  missed — try the image-role alias … exact keys always win").
- *Not* a new `site.*` source prefix: cleaner in the abstract, but it fixes
  nothing until every component schema is repointed, i.e. it is a migration
  dressed as a fix. It is also a new namespace on a shared mechanism —
  architecture-scope under the 2026-07-28 ruling — for no gain over reading the
  same columns behind the path components already declare.
- *Not* backfilling `site_specs.identity` from `sites` with a data migration:
  duplicates a fact into a second store and guarantees they drift. The platform
  has already ruled which store is canonical; the resolver should read it.

**Safety property the whole change rests on:** the literal path is tried first
and always wins, so **no path that resolves today changes its value**. The
fallback can only add resolution where the aspect held nothing. This is
test-asserted (`TestLiteralSpecPathAlwaysWinsOverBothAliases`) because it is the
property a reviewer would otherwise have to take on trust.

**Deliberately NOT COALESCEd** the way `loadSiteDataFull` is. That function needs
a non-empty string for a template, so it falls back `company_name → name →
domain`. Here an empty value must stay empty: the caller's question is whether the
field resolved *at all*, and substituting a domain for a missing company name
would satisfy a `needs_human_review` field with a value nobody supplied. Missing
stays missing; `on_missing` governs. Asserted by
`TestMissingIdentityFactStaysMissing`.

## What this fix does and does not buy

- **5 sites** gain a resolvable contact email (oufe, robot-hands, vetcomparison,
  vonc, webdesign) — `sites.email` populated, spec empty.
- **3 sites** still resolve nothing (gamesdesign, loancalculator, relojistas):
  no contact fact in any store. That is a **data gap, not a code defect**, and
  `relojistas` is an owner ruling of *no contact route at all*, so it must
  continue to resolve nothing. Correct behaviour, not a shortfall.
- `hours` resolves nowhere and has no column. `contact-info` declares it
  `skip_field`, which is right.
- The nested-shape step fixes **0 sites today** and is still the more important
  half: a *new* site is written by the classifier, nested-only, so it is broken
  by default. That is the recurrence the bug file correctly identified.

## Corrections to the originating brief

1. **The remedy in the bug file is wrong** (candidates 1 and 2 fix 0 of 8 sites).
   Recorded above with the measurement, and contributed back into the bug file.
2. **"8 of 13" then "8 of 14" is now 8 of 15** — a fifteenth site
   (`loancalculator.co.uk`) exists and has no `identity` aspect at all.
3. **The bug's "second defect: the drop is silent" is STALE.** `plan_sections`
   now always emits `sections_deferred` and `sections_skipped`
   (`plan_sections_action.go:922-924`, and the empty-result path at 695-697), and
   `persistSectionSkips` writes them durably. Verified by reading; a withheld
   section is recorded. No work needed.

## Phasing

- **P1** — measure (done), diagnose through the loop (done, corr `0f76987c`), fix
  + tests, council gate, commit. ← this session
- **P2** — the fix is inert until a chassis roll; the 5 sites then need their
  contact pages rebuilt to pick the block up. Verification recipe in `RUNBOOK` §4.
- **P3 (NOT this bug, filed separately)** — the census that found this also found
  that **74 of 100** distinct `site_specs.*` source paths across active components
  name an aspect that exists on **no site**. Different defect, different fix,
  own file.
