# 467 — a re-plan cannot add ANY new page to a site of 20+ pages, silently

**Filed 2026-09-03 by session `463`, found while fixing `bugs_open/463`. UNOWNED.**
**Severity: on `[MEASURED 2026-09-03]` 26 of 42 sites with pages — 62% — a re-plan can
add nothing at all. The only observable is one `logger.Warn` on a pod that keeps under a
second of log.**

## 1. The defect, in one branch

`truncatePreservingRealised` (`platform/orchestration/actions/v3_site_actions.go`), called
from `ValidateSitePlanAction` after reconciliation:

```go
	if len(keep) >= maxPages {
		logger.Warn("validate: preserved pages exceed max_pages; keeping all preserved, dropping all net-new",
			zap.Int("preserved", len(keep)), zap.Int("max_pages", maxPages))
		return keep
	}
	budget := maxPages - len(keep)
```

`keep` is every page in the preservation set — `noCurrentPlanFlag(rm) ||
realisedPageCompositionIsPreserved(rm)`, i.e. every `deployed` or `needs_rebuild` page.
`netNew` is everything else, which is exactly the set of pages the planner has just
proposed and the site does not have yet.

So once a site's built page count reaches the cap, **`budget` is never computed and every
newly proposed page is discarded** — not the excess, all of them. The cap comes from the
`validate_plan` step config, defaulting to 20:

```go
	maxPages := 20
	if mp, ok := config["max_pages"].(float64); ok {
		maxPages = int(mp)
	}
```

Migration 053 sets `"max_pages": 20`, so the configured value and the default coincide today.

## 2. Measured, and the measurement could have come out otherwise

`[MEASURED 2026-09-03]`, live DB — sites by count of pages in the preservation set
(`build_status IN ('deployed','needs_rebuild')`):

| | |
|---|---|
| sites with any pages | **42** |
| **at or over the cap — every net-new page dropped** | **26** |
| within 5 of the cap | 5 |
| largest site | **151** pages (webdesign.co.uk) |

The twelve largest: webdesign.co.uk 151, loanandmortgagecalculator.co.uk 62, finetuning.uk
57, leopardessconsulting.co.uk 53, robot-hands.com 50, gamesdesign.co.uk 49,
ai-agent-orchestration.com 47, loancalculator.co.uk 44, gaswholesalers.com 44,
mortgagecalculator.co.uk 43, dartsonline.com 43, seotools.co.uk 42.

```sql
WITH p AS (
  SELECT s.domain, count(*) FILTER (WHERE pg.build_status IN ('deployed','needs_rebuild')) AS preserved
    FROM sites s JOIN pages pg ON pg.site_id = s.id
   GROUP BY s.domain)
SELECT count(*) AS sites, count(*) FILTER (WHERE preserved >= 20) AS at_or_over_cap FROM p;
```

Had the estate's sites been mostly small this would have returned 2 or 3 and the bug would
be theoretical. It returned 26.

## 3. Why this is not the cap doing its job

The cap exists (`bugs_open/001` fix step 3) to stop an LLM re-proposing 80 pages and
evicting live ones. Preserving the built pages is right. **Dropping the entire net-new set
rather than filling a budget is not the same decision**, and nothing states it as intended:
the branch reads as an edge case for `budget <= 0`, and the `Warn` calls it "dropping all
net-new" without saying that on a mature site this is the *normal* path, not the edge.

The consequence is that page growth on this estate stops dead at 20 built pages, for every
site, through any number of re-plans — and the sites past that point are precisely the
established, revenue-carrying ones.

## 4. Why nobody has noticed

Same shape as `bugs_open/463`, which is how it was found:

- The plan still *succeeds*. `validate_plan` returns a page array, the orchestration
  completes, `capability_gaps_emitted` is 0.
- The only record is one `logger.Warn` in a chassis pod that retains under a second of log
  (`bugs_open/136` §11).
- No durable row is written — no `agent_error_log`, no `capability_gap`, no work item.
- On a mature site, "the re-plan added nothing" looks exactly like "the planner proposed
  nothing new", which is a perfectly ordinary outcome.

**Not** covered by the `428` lane's `recommended_type_reconciliation.go`, which classifies
by stage and would bucket this as `dropped_in_validation` — good — but only for a page type
that appears in the strategy's `recommended_page_types` and has **no** surviving page of
that type. A site at the cap already has pages of most types, so the type survives and the
check stays quiet.

## 5. Fix candidates, ordered by what closes the door

1. **Make the bad state unrepresentable: the cap should bound what a re-plan ADDS, not
   what a site may contain.** `maxPages` was written when a plan was the whole site; on a
   site with a current plan the preserved set is not a proposal to be capped, it is a fact.
   Split the two: preserve everything preserved, and apply the budget to `netNew` only
   (`netNew[:min(maxNewPages, len(netNew))]`). This needs an owner ruling on what
   `maxNewPages` should be, which is the real question the current code answers by accident.
2. **Fail loudly whichever rule wins.** A page dropped between `plan_site` and the plan
   writer must leave a durable record naming it. `reconcileCounts.DroppedPages` (added by
   `bugs_open/463`'s fix, commit `9b540c2e6`) is the shape; this pass should populate it
   too, with `Pass: "truncate"`.
3. Do **not** simply raise `max_pages`. It moves the wall, and the wall is silent — a site
   at 51 pages with a cap of 50 has the same bug and the same absence of evidence.

## 6. How to verify a fix

Assert at the step boundary, not the served site. On a site with ≥20 preserved pages,
re-plan and compare:

```sql
SELECT jsonb_array_length(collected_data->'plan_site'->'result'->'pages')  AS proposed,
       jsonb_array_length(collected_data->'validate_plan'->'pages')        AS survived
  FROM orchestration_states WHERE correlation_id='<corr>';
```

Today `survived` equals the preserved count exactly and `proposed − survived` is the whole
net-new set. The pass condition is that a named new page proposed by `plan_site` is present
in `validate_plan`'s output and then in `site_plan_pages`.

## 7. Relationship to 463

Found while fixing `bugs_open/463` (Pass C deleting a section index's new children), and
they are the same class in series: **463 stops a new child reaching the plan on any site;
467 stops any new page reaching the plan on a site of 20+.** So on 26 of 42 sites, 463's
fix alone will not fill an empty hub — 467 will eat the children immediately afterwards.
Whoever verifies 463 at the artefact must pick a site under the cap, or fix this first.
gamedesign.uk (9 pages) is under it, which is why 463's own case is verifiable today.
