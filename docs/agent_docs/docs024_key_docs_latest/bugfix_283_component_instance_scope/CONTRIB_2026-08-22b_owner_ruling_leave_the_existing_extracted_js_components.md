# CONTRIB 2026-08-22b (from the `bugfix_311_component_keys` lane) — **OWNER RULING: "leave the existing components." The extracted-JS legacy set is NOT to be converted. Read this before you widen the sweeper.**

**Owner, 2026-08-22, answering the second of the two questions this lane put up:** *"leave the
existing components."*

Together with yesterday's ruling (*"the components that need scoping can be inline"*) the position
is now complete and has a clean seam:

| population | ruling |
|---|---|
| **NEW** components that need instance scope | keep the JS **inline**, run your birth gate on it (contrib 2026-08-22) |
| **EXISTING** components already shipping extracted `/tools/assets/<fn>.js` | **leave them.** No re-inlining, no conversion |

## The interaction you need before acting — this ruling and the sweeper widening pull opposite ways

My `CONTRIB_2026-08-21b` §1 recommended widening the sweeper's `getElementById`-only clause,
because it sees **8 of 39** and misses 31 (27 of which carry a script — 6 `querySelector`, 22
referencing a `.js` asset). **That widening is exactly what would drag the now-protected legacy set
into your conversion queue.** So the two must be paired deliberately:

- **Widen for VISIBILITY, not for automatic conversion.** The 22-referencing-an-asset population is
  the ruled-out set; if your sweep files `instance_scope_conversion` items for them it will be
  filing work the owner has declined, and someone will have to cancel them by hand.
- The cheap discriminator is already in the data: **a template that references
  `src="/tools/assets/….js"` is extracted-JS legacy**; one with inline `<script>` and literal ids is
  a genuine target. That split is one clause and it matches the ruling exactly.
- If you would rather not widen at all now, that is also consistent — the ruling means the legacy
  set has no remedy pending, so counting it only buys an accurate backlog figure.

## What the ruling costs, stated so it is not discovered later

Those components **cannot carry two instances on one page** — duplicate ids, and their external
script's lookups resolve to the first. Nothing is damaged today (each page binds one instance via
`page_components`), and the seven rows `311`'s diversion minted are in this set. The accepted
consequence: **if a page ever needs two of the same calculator, that component must be rebuilt as a
new inline one, not converted in place.** Worth a line in `RFC_034` so the next reader does not
re-open the question.

## Where this leaves the register

`RFC_034`'s remaining decisions are now closed on the owner's side. The only cross-lane item still
open between us is the one in `bugs_open/351` (the `</section>` predicate, `bugfix_198_roundtrip_writers`
lane) — and its **backfill-ordering** question, which the owner is holding until that fix lands.
