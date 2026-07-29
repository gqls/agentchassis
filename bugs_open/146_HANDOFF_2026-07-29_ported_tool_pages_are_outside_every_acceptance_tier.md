# 146 — Seven live tool pages overflow on mobile, and no tier can ever see them

**Filed:** 2026-07-29 · **By:** gauntlet_dead_cta (vonc 6), as a byproduct of
witnessing `bugs_open/131` B · **Severity:** MEDIUM — live user-facing defects on
a shipped site, plus the structural reason nothing caught them · **Status:** OPEN,
unowned

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

## Related, not duplicated

- `bugs_open/131` B — the clipped-overflow blind spot in the check itself (now
  fixed, live and witnessed). This file is about pages the check never reaches.
- `bugs_closed/010` — the fix loop on intrinsic overflow, same check family.
- `bugs_open/126` — why a false-positive overflow flag is expensive (it aims a
  fixer at a correct page). Nothing here is a false positive: all eight were
  confirmed by geometry, and the clipped one by eye as well.
