# 181 — every detector that selects `build_status = 'deployed'` is blind to 28 live pages (35 counting archived-but-possibly-served)

**Filed:** 2026-08-03 by the `bugfix_175_page_role_upsert` lane, **at the council gate's
request** (round 2 on corr `e78c62e3-7f01-48f1-b083-924eaccd195a`, REVISE). Three seats —
`bug_historian`, `reuse_agent`, `debug_historian` — independently asked the same question
about `bugs_open/175`'s fix: *"does any other call site still hand-roll a
build_status/deployed_at liveness check?"* The audit they asked for produced this.

**Severity:** a **false-negative** class, not a corruption. Nothing is written wrongly;
some live pages are simply never looked at. Filed rather than fixed because the fix
changes what ~10 checks report fleet-wide, which is its own measurement.

**Status:** OPEN, unowned. Census done, no fix attempted.

---

## The mechanism

`pages.build_status = 'deployed'` reads as "this page is live". It is not.
A page flagged `needs_rebuild` **has been deployed and is still serving its previous
artefact** — that is `bugs_closed/037` in full, and `datahelpers/links_deployment_test.go`
has carried the warning since 2026-07-26:

> *"A page deployed once and later flagged needs_rebuild still serves its old artefact.
> Singling that status out would false-flag 34 live pages."*

The estate already has the correct predicate — `datahelpers.NeverDeployedPagePredicate`
(`deployed_at IS NULL AND COALESCE(build_status,'') <> 'deployed'`), negated for
"has shipped". `bugs_open/175` converged the two `pages` upsert helpers that make a
liveness *judgement* onto it. **The detectors were not converged, and they make the same
judgement for a different purpose.**

## The measurement (live, 2026-08-03)

```sql
-- CORRECTED 2026-08-03 — the first version of this census filtered `status='active'`
-- without ever enumerating that column. Enumerate first, then decide:
SELECT COALESCE(status,'(null)') AS status, count(*) AS pages,
       count(*) FILTER (WHERE COALESCE(build_status,'') <> 'deployed'
                          AND deployed_at IS NOT NULL) AS shipped_but_not_deployed_status
FROM pages GROUP BY 1 ORDER BY 2 DESC;
--  active   | 557 | 28
--  archived |  25 |  7
```

**28 ACTIVE pages have shipped and are being served, yet carry a `build_status` other than
`deployed`.** Every query whose WHERE clause is `p.build_status = 'deployed'` skips all 28.

> **CORRECTED 2026-08-03, at the council's `debug_historian` seat's insistence, and the
> correction is worth more than the number.** The original census read
> `... AND status='active'` and reported **28** as though that were the whole population. I
> scoped a blast radius by a status column **without first enumerating its values** — the
> generalised form of the "`sites.status` is informational, do not scope by it" trap.
> Enumerated: `pages.status` has two values, and **7 more rows are `archived` AND
> shipped-but-not-`deployed`**. Whether those 7 belong in the figure is a real question
> rather than a rounding error: `bugs_open/098` establishes that **archiving does not
> retract the live page**, so an archived page may well still be served — in which case a
> detector that skips it is blind to a live artefact, which is this bug. **So: 28 is the
> number for "pages the platform considers current"; 35 is the number for "pages that may
> still be on the wire". Both are stated because the difference is a decision (does a
> detector owe anything to an archived-but-served page?) and not mine to make silently.**

## The census — which sites make the judgement, and which do not

**Genuinely affected — a page-level selector meaning "the live pages" (5 discovery checks
plus these action-side queries):**

| site | what it skips |
|---|---|
| `discovery_checks/check_orphan_pages.go:200` | a shipped `needs_rebuild` page is never checked for orphanhood |
| `discovery_checks/check_unresolved_sections.go:43` | ditto, for unresolved sections |
| `discovery_checks/check_tool_acceptance_due.go:55` | a shipped tool page awaiting rebuild is never put to acceptance |
| `discovery_checks/check_page_component_status_drift.go:90` | drift on a shipped page is invisible |
| `discovery_checks/check_backend_entry_orphaned.go:119` | ditto |
| `maintenance_actions.go:723,750` · `render_news_section_html.go:77` · `request_render_audit_action.go:110` · `store_generated_component_action.go:843` · `render_directory_action.go:345` · `component_library.go:2378,2434` | listing / rendering / audit selectors on the same shape |

**NOT affected, checked rather than assumed:**

- **Every `pc.build_status` site** (`check_empty_sections`, `check_undeployed_assets`,
  `check_orphan_element_refs`, `check_image_url_404`, `check_dead_controls`,
  `check_placeholder_image_in_use`, `check_required_fields_missing`,
  `save_page_sections_action`, `fix_component_template_action`) — that is
  `page_components`, a different table and a different question. `check_dead_controls:70-74`
  already documents choosing component liveness over page liveness deliberately.
- **The other three `pages` upsert helpers** — `site_db_actions.go:1090 upsertPage`,
  `cmd/webdesignport/import.go:163 upsertPage`, `adopt_verbatim.go:470`. This was the
  council's actual question, and the answer is clean: **none of them makes a liveness
  judgement at all.** They *write* `build_status` as a value, or preserve the existing one
  (`CASE WHEN pages.build_status IS NULL THEN 'planned' ELSE pages.build_status END`).
  There is nothing to converge.
- `check_news_feed.go:690` (`COALESCE(build_status,'') <> 'deployed'`) — `bugs_closed/015`'s
  stranded-page predicate, deliberate and documented; `bugs_closed/081` records why.
- `create_tool_cross_link_items.go:458,471` — an `ORDER BY`, a ranking preference, harmless.

## The one that is logically identical, and is NOT drift

`queryresolve.FetchablePageEligibilitySQL` is
`AND (p.deployed_at IS NOT NULL OR p.build_status = 'deployed')` — **exactly
`NOT (NeverDeployedPagePredicate)`**, spelled out. It is not accidental: that file
documents a deliberate family of three eligibility predicates and says so —
*"The three are deliberately distinct; each comment says why, so the split does not read
as accidental drift."*

So this is a **naming** problem rather than a logic one: two constants in two packages
express one judgement and neither mentions the other. `175` added a cross-reference
comment to each as a stopgap. Whether they should be one constant is a real question and
is left open here rather than answered in passing.

## Fix candidates (none applied)

1. **Converge the affected selectors** onto `NOT (datahelpers.NeverDeployedPagePredicate)`.
   Correct, and it will make ~10 checks start reporting on up to 28 pages they have never
   seen. **Measure what each newly reports before shipping** — a check that suddenly files
   28 new items is indistinguishable from a broken check.
2. **Converge the constants** (`NeverDeployedPagePredicate` ↔ `FetchablePageEligibilitySQL`)
   into one home with the alias-prefix difference handled. Small, but it touches a
   deliberately-documented family — read `queryresolve.go:210-236` first.
3. **Detector only:** a `pattern-check` rule for a new `build_status = 'deployed'` on
   `pages`. Cheap and stops the class growing while 1 is decided.

## How to verify a fix

For each converged site, run the query **both ways** against production and diff the row
sets — the delta is exactly the pages that check has been blind to. Then confirm the new
rows are genuinely defective rather than newly-visible-and-fine, because the first
consequence of fixing a false-negative is a burst of findings.

## Related

- `bugs_closed/037` — `needs_rebuild` pages unprotected by the replan guard; the same
  predicate, filed once already.
- `bugs_closed/081` / `bugs_closed/175` — the two upsert arms, now converged (PBP-027).
- `datahelpers/links_deployment_test.go` — the test that already forbids naming the status.
- `LANDMINES.md`, "`pages.build_status = 'deployed'` is NOT 'is this page live'".
