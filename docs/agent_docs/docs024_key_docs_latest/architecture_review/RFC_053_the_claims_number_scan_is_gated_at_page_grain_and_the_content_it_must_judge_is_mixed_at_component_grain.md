# RFC_053 — the claims number scan is gated at PAGE grain, and the content it must judge is mixed at COMPONENT grain

**Status:** OPEN — raised 2026-08-24 by the `bugs_open/364` lane, at the explicit request of the
council's `architecture` seat on round `b8df25dc-7d19-48b9-9b52-b93b25523d4a` (APPROVED; the seat
scored the shipped fix a clean `point_fix` needing no RFC, and asked separately that *"the
ClaimSurface component-grain RFC should actually get filed, not just referenced in a commit
message — three occurrences of the same shape is the threshold where 'filed in rationale' should
become 'filed in architecture_review/'"*). Not a blocker to anything shipped.

## The shape, in one line

**Three bugs (`073`, `102`, `364`) have now been closed by widening the same page-grain allow-list,
and each one's own write-up says page grain is the wrong grain.**

## What exists today

`datahelpers.ClaimSurface{PageType}` + `ProseNumbersAreClaims()` (`claims.go`) gate
`ScanUnregisteredNumbers` — and **only** that scan. Verified 2026-08-24 by enumeration, because a
council seat asked the same question and its premise was wrong: `editorialPageTypes` has exactly
**one** reader (`ProseNumbersAreClaims`), which has exactly **one** caller (the guard at
`claims.go:868`). `ScanBannedClaims`, `ScanAllBannedClaims(WithSuppressed)` and `ScanStatClaims`
take no `ClaimSurface` at all and cannot consult it.

`editorialPageTypes` now holds eight page types. Five were added by `bugs_open/102` (2026-07-28);
three — `adoption-tracker`, `protocol-tracker`, `model-directory` — by `bugs_open/364`
(2026-08-24), and **that addition knowingly fails the map's own two-part membership bar.**

## Why page grain is the wrong grain, measured

The bar says a member must have *"a body that is never marketing"*. The tracker pages do not:

| slot | content | what the gate does |
|---|---|---|
| `hero` | *"Multi-agent systems deployed to production in days, not months"* | **silenced** |
| `*-listing` | `rollout_scope Over 80% of Fortune 500 deploying active agents … source` | correctly silenced |
| `call-to-action` | *"We run over 1,600 orchestrations a day across 13 live production systems"* | **silenced** |

A single page mixes the site's own first-person marketing voice with an aggregated table of third
parties' figures. **Page type cannot express that, and the number of page types that mix this way
is growing** — trackers, directories, comparison pages, entity pages. The map either keeps growing
one measured false positive at a time, or the grain changes.

`claims.go`'s own comment on the `report` page type has been asking for this since 2026-07-28:
excluding a page type there *"would fix those by coincidence, not by mechanism."*

## The question for the architecture track

1. **What is the trusted signal, and where does it live?** The candidates, in the order they were
   considered and why the first two are already out:
   - **slot name** (`*-listing` ⇒ third-party) — **measured and REFUTED.** `case-studies-list` on
     the same site is a list of *our own* work ("orchestrates 30+ specialised agents", "under 4
     hours"): three genuine claims that this rule would blind. The suffix does not discriminate.
   - **a marker in the rendered HTML** (`data-claims-scope="third-party"`) — **rejected without
     measuring, on principle.** The HTML is LLM-generated, so the thing being policed would be able
     to emit its own exemption. Any such declaration must come from trusted DB-side structure.
   - **a declared property on the component definition** (`content_components.function` /
     `category` / a new column) — the surviving candidate. `claims_stats.go` is the precedent
     (*"a figure in a `stat1_value` field is a published quantitative claim BY CONSTRUCTION"*), and
     `claims_regulated.go` is the precedent for the self-vs-third-party distinction specifically
     (first-person by design, measured 8/8 must-catch and 10/10 must-allow, including
     *"Nationwide Building Society is authorised and regulated by the FCA"* passing).
2. **How does the signal reach the gate that actually refuses?** This is the real engineering
   question and it is unevenly hard. Traced 2026-08-24: **4 of the 5 `ClaimSurface` construction
   points can pass component identity in one line** — `check_unverified_claims.go:511` and `:553`
   (both already hold `slotName`), `save_sections_claims_guard.go:151` (`sections[i].ComponentName`
   *is* the slot name, and the loop is already per-section), and `cmd/claimscan/main.go:162` (`slot`
   already parsed from TSV field 2). The struct uses keyed literals, so widening is
   source-compatible. **The fifth is `validate_page_content.go:377` — the build gate, i.e. the only
   one that refuses a page — and it scans whole-page HTML with no per-component split.** Component
   identity is reachable there via `sections_metadata` in `CollectedData` (the pattern
   `validate_page_content_stats.go:86-119` already uses), **but only `page-build-handler` supplies
   it**; the other three callers of that gate would resolve UNKNOWN.
3. **Does UNKNOWN stay noisy?** It must, and this is the property most at risk in a refactor. The
   current design is explicit that the zero value means UNKNOWN and is **scanned**, because *"a
   scanner that has gone quiet and one that is broken look identical from the outside"*. A
   component-grain gate has strictly more unknowns than a page-grain one (site chrome, a new
   template, a component built before the property existed). **The safe default must remain
   "scan", and that means the fix reduces false positives only where a declaration exists** —
   which is the honest cost of doing it properly.

## Scope ruling this RFC should settle

Under RFC_022's 2026-08-11 narrowing this is plausibly **not** architecture-scope: it would be
opt-in, its unsafe side (going quiet) would be the default-OFF side, and no live consumer names it
at the point it ships. **But that must be asserted with the consumer enumeration, not without it** —
asserting it without the query is itself the objection. This RFC exists partly so that enumeration
is done once, in public, rather than inside a bug patch.

Note also the optional-key budget (WFA-013, N=10): if the declaration lands as an action input key
rather than a column, it counts.

## What it costs to leave it

**Bounded and visible, which is why the seat did not raise severity**: one map entry per newly
discovered page type, each requiring its own measured false positive first. The compounding cost is
different and less visible — every page type added to the map takes its hero and CTA with it, so
the estate's first-person claim coverage shrinks by whole pages each time, silently, and only the
pinned test (`TestTrackerPagesGiveUpTheirFirstPersonClaims`) records that it happened.

## Measured, dated — all `[MEASURED 2026-08-24]`

- Fleet claims census, 19 opted-in sites, ~1,457 live components, each against its **own current**
  register, export asserted row-for-row against the DB: **44 findings.** 20 third-party figures in
  aggregated listings (**zero precision**), 16 genuine first-person claims, 5 counting-fact drift
  (`bugs_open/386`), 3 formula/legal/hypothetical.
- After the page-grain interim: ai-agent-orchestration.com **36 → 16**, zero survivors on any
  tracker page, all 16 genuine claims retained.
- Damage the interim addresses: `agent_error_log` — **70 refused build attempts, 23 distinct pages,
  8 sites, 60 days**, still daily; `model-directory` refused 20 times since 2026-07-29.
- Page types NOT added, on the same bar that keeps `blog-index` out: `entity-directory` (4 pages),
  `entity-page` (21 pages) — **zero** measured findings; analogy is not a measurement.
- A second, independent defect found while testing the interim and fixed as a council fast-follow
  (`0f9f7f3ff`): `businessClaimContextRe` carried `orchestration` **singular**, so the CTA claim in
  the table above had never been scanned at all. **That gate is an allow-list of NOUNS with the same
  unbounded-miss property as the unit allow-list** — and a miss there is silent, where a false
  positive is loud. Whoever takes this RFC should decide whether the noun list is in scope too;
  the same "wrong instrument" argument applies to it.

## Relations

`bugs_open/364` (§5b–5d hold the full reasoning), `bugs_closed/073`, `bugs_closed/102`;
concept register **CLM-016** (extended 2026-08-24), CLM-014, CLM-019;
`LANDMINES.md` — *"The claims number scan is now SILENT on three whole page types"*;
council round `b8df25dc-7d19-48b9-9b52-b93b25523d4a`.
