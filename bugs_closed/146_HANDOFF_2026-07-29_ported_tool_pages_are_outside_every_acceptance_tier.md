# 146 — Seven live tool pages overflow on mobile, and no tier can ever see them

> # CLOSED 2026-08-18 — fixed, council-APPROVED r1, LIVE on v1.0.1310, and INDUCED in production the same evening
>
> The whole chain ran, with evidence at each step:
>
> 1. **Binary-verified on BOTH replicas of v1.0.1310** (pods 18:00Z):
>    `routePortedAcceptanceFailure` → 2 hits in `/proc/1/exe` each, invented-symbol
>    control → 0.
> 2. **Council APPROVED round 1** (corr `d2edf61d`, 13 approve / 4 advisory objections,
>    none high; all four mediums answered with evidence — lane NOTES).
> 3. **Production induction 19:24Z** (item `cc37a4d7`, the due-sweep's exact ported spec
>    shape, hand-filed `triaged`): the real Tier-4 run FAILED mobile-fit on the live page
>    (div#preview-card 380px forced by div.card-img 324px — the same defect this file
>    measured on 07-29), and the new arm filed
>    **`ported_tool_fix:tool_acceptance_tier4:vibe-equalizer:6b49db8e…` at
>    `needs_human_review`** with the full actionable spec: subject, page, shared
>    component id, failing checks, the criteria, the mobile screenshot (durable s3 URI +
>    presigned view), `overflow_forced_by` and the fix hint. The verdict note's `Fix:`
>    line records the filing — the first ported Tier-4 failure ever to route anywhere.
>    (Before this fix the identical failure was eaten FOUR times: pasteboard ×2,
>    vibe-equalizer ×2, 2026-08-05 and 08-14.)
>
> **What lives on, and where** (none of it gates this file):
> - **The rewrite programme** — OWNER RULING 2026-08-18: all ported tools are rewritten
>   natively, fleet-wide (72 clause-b pages / 6 sites censused in the lane NOTES). The
>   webdesign_tool_rebuilds lane owns the proven machinery; the seven pages of this
>   file's title are theirs, with the re-measured evidence pointed into their NOTES.
> - **The induced item itself** stays in the review queue as a REAL finding — it is the
>   worked example of the route and queue-ordering signal for the rebuilds.
> - **Named residual**: a failing ported verdict with no `spec.component_id`
>   (pre-seed-425/bespoke shapes), a transient lookup error, or a renamed fork still
>   takes the old manual arm, deliberately (bug_historian's advisory, dispositioned).
> - The needs_human_review drain cadence belongs to 033/083/291, unchanged.
>
> Lane: `docs024_key_docs_latest/bugfix_146_ported_tools_acceptance/`. Commits
> `1549dc58b` (fix+tests+register+016b, `Council-Submitted:`), `ec02ac8bd` (verdict
> read + advisories), `55c0f3fb8` (ruling + live proof), plus lane docs.

**Filed:** 2026-07-29 · **By:** gauntlet_dead_cta (vonc 6), as a byproduct of
witnessing `bugs_open/131` B · **Severity:** MEDIUM — live user-facing defects on
a shipped site, plus the structural reason nothing caught them · **Status:** OPEN,
**taken up 2026-08-17** (lane: `docs024_key_docs_latest/bugfix_146_ported_tools_acceptance/`)
— re-validated VALID the same day; read the dated section at the foot before the body,
the population numbers above it are stale.

## The symptom, measured

Scanning every deployed tool page fleet-wide at the browser-runner's own mobile
profile (390×844) with the `no_horizontal_overflow` clause **extracted from
`internal/adapters/browserrunner/run_checks_action.go` at runtime**:

**86 clean · 8 overflowing · 0 errors.** One (`tool-loot-table-balancer`) is the
clipped-overflow case that closed 131-B. **The other seven are all on
webdesign.co.uk**, and all are ordinary `scrollWidth` overflow that even the OLD
clause detects:

| page | over | widest offender | cause reported by the clause |
|---|---|---|---|
| `/tools/smooth-shadow/` | 293px | `section (330px)` | `flex-wrap:nowrap` |
| `/tools/recommender-engine/` | 196px | `h3 (565px)` | fixed width 565px |
| `/tools/layout-generator/` | 119px | `div (488px)` | fixed width 438px |
| `/tools/css-variables/` | 95px | `div.section-box (464px)` | fixed width 414px |
| `/tools/social-card/` | 87px | `strong (455px)` | fixed width 455px |
| `/tools/blob-maker/` | 33px | `div (402px)` | fixed width 352px |
| `/tools/vibe-equalizer/` | 11px | `a.card (380px)` | fixed width 380px |

Reproduce: `docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/p4_sources/scan_clipped_tools_2026-07-29.py`
(takes a file of URLs; full output committed beside it).

## Why nothing caught them — the structural half

`tool_acceptance_due` (`platform/orchestration/actions/discovery_checks/check_tool_acceptance_due.go:41-55`)
only considers components with **`cc.component_level = 'tool'`**. Measured on the
live DB:

```sql
-- 17 rows: the ENTIRE eligible population fleet-wide
SELECT s.domain, cc.function, (dp.body LIKE '%```criteria%') AS has_criteria
FROM content_components cc
JOIN page_components pc ON pc.component_id = cc.id
JOIN pages p ON pc.page_id = p.id
JOIN sites s ON s.id = p.site_id
LEFT JOIN doc_plans dp ON dp.subject_type='tool' AND dp.subject_key=cc.function AND dp.is_current
WHERE cc.component_level='tool' AND cc.is_active AND p.status='active' AND p.build_status='deployed';
```

- **17** deployed tool components fleet-wide; **8** have a current PLAN with a
  ```` ```criteria ```` fence, so **8** are eligible for Tier 4 at all.
- **webdesign.co.uk contributes ZERO.** Every one of its 64 `/tools/*`
  page-components is `component_level = 'section'` — these are **ported** pages
  (the clause attributes several offenders to `component: "ported-page"`), so
  `tool_acceptance_due` cannot see them, no PLAN exists, no criteria exist, and
  neither Tier 2 nor Tier 4 will ever run against them.

So the defects are not "missed" — the pages were never in any tier's population.
A porting path produced live tool pages that no acceptance tier covers.

> **[CORRECTED before filing]** My first framing was *"the acceptance lane has
> run 4 times ever across 94 tool pages"*. Both halves misled: 94 was pages whose
> NAME starts with `tool-` (it includes guides and ported pages), not the lane's
> population, which is 8. Four runs against 8 tools on a 7-day cooldown is not the
> scandal the ratio implied. **The denominator came from my question, not from the
> mechanism** — the same shape as `[[narrow-filter-defines-the-conclusion]]`.
> The real finding survived the correction and is narrower: a whole site's tools
> are structurally outside the tiers.

## Fix candidates, ordered by what closes the door

1. **Make the porting path emit tool-level components (or a criteria fence) for
   anything served at `/tools/`.** Closes it at the source: a ported tool then
   joins the same population as a generated one and cannot be silently uncovered.
   Needs a decision on whether ported pages are *deliberately* out of scope —
   that is an owner/architecture call, not a code question.
2. **Widen the discovery check's population** (e.g. tool-level OR a page whose
   URL is under `/tools/` on a site that has ported tools). Cheaper, but leaves
   "which pages are tools" as a heuristic in two places.
3. **Fix the seven pages page-side and accept the coverage gap.** Repairs the
   symptom only; the next ported tool arrives uncovered. `[NOT RECOMMENDED alone]`
   — "operators must remember to check ported pages" is the defect, not the fix.

## How to verify a fix

Re-run the scanner over the same URL list: the seven must go clean at 390px.
For the structural half, the eligible-population query above must return the
webdesign.co.uk tools (candidate 1 or 2), not just the current 17.

---

# RE-VALIDATED 2026-08-17 and FIXED IN CODE 2026-08-18 — OPEN until the roll + live induction

Taken up by the `bugfix_146_ported_tools_acceptance` lane (full trail in
`docs024_key_docs_latest/bugfix_146_ported_tools_acceptance/`). Two corrections to this
file's own mechanism section, then what shipped:

> **CORRECTED 2026-08-18 — the structural half above was fixed THREE HOURS after this
> file was filed, and the file never said so.** `ac9f75a0c` (2026-07-29 17:19, TL-033)
> replaced the `component_level='tool'` gate with the shared `toolEligibilityWhere`
> (tool_eligibility.go), which admits ported sole-component `page_type='tool'` pages,
> keyed by page-name stem. Fix-candidate 2 above therefore shipped the day of filing.
> Re-running this file's own SQL today measures a predicate 19 days dead — one session
> did exactly that and recorded it (WRONG_CALLS 2026-08-17) before the code read caught it.

**What actually kept the symptom alive** (all 7 pages re-scanned live 2026-08-17 with the
extracted adapter clause: **0 clean · 7 FLAGGED · 0 errors**, same culprits):

1. **No fence → no run.** Tier 4 is criteria-gated; 6 of the 7 have no current
   ```criteria``` PLAN. 48 of 67 ported pages fleet-wide are unfenced (webdesign 45,
   loancash 3). Fencing policy is an OWNER DECISION, costed in the lane PLAN §"Door 1" —
   deliberately not taken unilaterally.
2. **Fenced + FAILED → silent sink.** The Tier-4 judge re-derived its component by
   `cc.function = <subject key>`, which no ported instance satisfies, so a failing ported
   verdict wrote "route this manually" and filed nothing — `pasteboard` and
   `vibe-equalizer` (fence present) each FAILED twice, 2026-08-05 and 08-14, zero items.
   This is `bugs_closed/281` Finding B, which that lane left open and unowned.

**FIXED IN CODE (door 2): `1549dc58b`, 2026-08-18** — `JudgeAcceptanceResultsAction` is
now the THIRD producer of `ported_tool_fix` (after tool_health and Tier-2 acceptance),
firing only when the run item's own `spec.component_id` resolves to an active non-tool
component; key `ported_tool_fix:tool_acceptance_tier4:<subjectKey>:<siteID>`, handler-less
`needs_human_review`, refresh-and-merge dedup. Register TL-042 updated in the same commit.
**Council APPROVED at ROUND 1** — corr `d2edf61d-87af-4195-bcce-c5717afc2d9e`, 2026-08-18,
"approved with 4 advisory objection(s) — none high-severity"; all four mediums answered
with evidence the same day (lane NOTES § "council APPROVED ROUND 1" holds each
disposition, including the guidelines seat's DELETE+INSERT claim being contradicted by
this file's own pinned dedup contract). The commit carries `Council-Submitted:`
(pre-verdict, forward-only); 098 credits it automatically. **Go is inert until the next
chassis roll** (fleet is on v1.0.1309 = `f0117fb8`, which predates this).

Named residual (bug_historian's advisory): a failing ported verdict whose run item lacks
`spec.component_id` (pre-seed-425 / bespoke shapes only — every current producer writes
it), a transient component lookup error, or a renamed fork still takes the old
"route this manually" arm, deliberately — a fourth handler-less queue for a shrinking
legacy shape would be noise, and the 033/083 cadence work owns the drains that exist.

> **OWNER RULING 2026-08-18: "Ported tools should all be rewritten and be fully
> customisable by the framework like any other tool."** This answers the fence question
> (door 1) as none-of-the-above — native rebuilds fleet-wide, the webdesign_tool_rebuilds
> approach extended. Measured backlog: 72 clause-b pages / 6 sites (webdesign 56,
> gamesdesign 6, loancash 3, loanandmortgagecalculator 3, vonc 3, mortgagecalculator 1).
> On loancash the owner's "was a new build, should be framework-built" is half right,
> measured: 2 of its tools ARE framework forks (2026-08-12, fenced, mobile-checked);
> 3 calculators + the tools-index came in as verbatim ports during the 2026-08-01
> adoption and join the rewrite list. The judge fix below stays the safety net while any
> port still serves.

**To close this file** (the fixed-AND-live bar):
1. ✅ **DONE 2026-08-18** — v1.0.1310 (pods 18:00Z) carries it: `routePortedAcceptanceFailure`
   present in BOTH replicas' `/proc/1/exe` (2 hits each), invented-symbol control 0.
2. The standing failing case IS the induction — **FIRED 2026-08-18 18:4xZ** rather than
   waiting for the 7-day cadence: hand-inserted the due-sweep's exact ported item shape
   (item `cc37a4d7-5019-401b-bb5b-7db5a23c4c2b`, `triaged` — a hand-filed item at
   `detected` is invisible to the dispatcher). Expected: mobile-fit FAILS (live page still
   overflowed 08-17) → `ported_tool_fix:tool_acceptance_tier4:vibe-equalizer:…` appears
   with the overflow attribution. ⚠ `tool_acceptance_run.sh` cannot fire a ported tool
   (resolves the page by `cc.function`, which is `ported-page` for every port) — the
   hand-insert with the proven 08-14 spec shape is the route until someone teaches the
   script the eligibility predicate.
3. The seven pages themselves are the **webdesign_tool_rebuilds** lane's (owner-directed
   native replacement; 4 of 63 done as of 2026-08-17) — this file does not gate on them.

## Related, not duplicated

- `bugs_open/131` B — the clipped-overflow blind spot in the check itself (now
  fixed, live and witnessed). This file is about pages the check never reaches.
- `bugs_closed/010` — the fix loop on intrinsic overflow, same check family.
- `bugs_open/126` — why a false-positive overflow flag is expensive (it aims a
  fixer at a correct page). Nothing here is a false positive: all eight were
  confirmed by geometry, and the clipped one by eye as well.
