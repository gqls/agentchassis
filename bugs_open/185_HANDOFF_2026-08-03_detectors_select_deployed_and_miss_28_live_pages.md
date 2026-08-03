# 185 — (RENUMBERED from 181 on 2026-08-03. The `bugfix_172_agent_state_cap` lane had
# already filed a different 181 — the code-lookup row caps case, committed `0a5ff9634`
# on 2026-08-02, hours before this one. Theirs was first; this one moved. Numbering is
# one sequence across both dirs and is never reassigned, so **resolve by slug** — and
# note 184 is ALSO doubly-used today, by two other lanes.)

# 185 — every detector that selects `build_status = 'deployed'` is blind to 28 live pages (35 counting archived-but-possibly-served)

**Filed:** 2026-08-03 (as 181; renumbered same day) by the `bugfix_175_page_role_upsert` lane, **at the council gate's
request** (round 2 on corr `e78c62e3-7f01-48f1-b083-924eaccd195a`, REVISE). Three seats —
`bug_historian`, `reuse_agent`, `debug_historian` — independently asked the same question
about `bugs_open/175`'s fix: *"does any other call site still hand-roll a
build_status/deployed_at liveness check?"* The audit they asked for produced this.

**Severity:** a **false-negative** class, not a corruption. Nothing is written wrongly;
some live pages are simply never looked at. Filed rather than fixed because the fix
changes what ~10 checks report fleet-wide, which is its own measurement.

**Status:** OPEN. **Picked up 2026-08-03 by the filing lane** — see § Progress at the foot.

**On the 2026-07-31 owner ruling (a cross-cutting root cause is not "filed" until it has
been through the `090` diagnosis loop, or the session states plainly why it substituted
first-hand verification):** `090` was **not** run, and the substitute is stated rather than
omitted. Every claim here is first-hand and mechanical — the census is `grep` output with
each site read and classified by hand, and the impact figure is a `GROUP BY` against the
live DB, re-run after the council caught the first version scoping on an unenumerated
status column. The loop's value is forming a cited theory when the cause is not where the
symptom is; here there is no theory to form — the predicate is visible in the source and
the missed rows are countable. What the loop could NOT have supplied, and what this fix
requires, is the both-ways row diff per call site in § Progress.

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


---

# PROGRESS 2026-08-03 — tranche 1 done: three detectors converged, two deliberately not

**Picked up by the filing lane.** Fix candidate 1, applied to the sites the measurement
justified and **only** those. Commit `171d52a1a`.

**Council: `66d07687-8e4a-46c5-aa0e-f056b96e1b2c`** (submitted 2026-08-03, verdict pending).

> **The commit's trailer is WRONG and this line is the repair.** `171d52a1a` carries
> `Council-Submitted: pending-185-tranche-1` — a placeholder typed before the submission
> ran, which resolves to nothing and will read as un-credited in the `098` coverage report.
> Forward-only forbids an amend, so the real correlation is recorded here instead.
> **The lesson is small and repeatable: submit FIRST, then commit with the id it printed.**
> The trailer is not a note to yourself, it is a join key.

## The seam, because the ALIAS was the whole reason this kept being re-typed

`datahelpers.NeverDeployedPagePredicate` could not be dropped into any query that aliases
its `pages` table — which is most of them — so each consumer wrote the expression out by
hand, and a hand-written copy is where `= 'deployed'` creeps back in. Two builders now:

```go
NeverDeployedPagePredicateFor(alias)   // "p" -> p.deployed_at IS NULL AND COALESCE(p.build_status,'') <> 'deployed'
PageHasShippedPredicateFor(alias)      // the parenthesised negation — the one detectors want
```

`NeverDeployedPagePredicate` is now **derived** from the first, so the bare and aliased
forms cannot drift. This is also the council's round-2 objection answered at the root:
the duplication existed *because* there was no aliased form to reuse.

## Converged — with the delta each one will actually report

The bug file's own rule was *"run the query both ways and diff the row sets… confirm the
new rows are genuinely defective rather than newly-visible-and-fine."* Done, and the first
answer was wrong in the useful direction:

| check | candidates newly visible | **findings it will actually file** | are they real? |
|---|---|---|---|
| `check_orphan_pages` | 22 | **2** | yes — `gaswholesalers.com` `tool-breakeven-volume-calculator-guide` and `tool-fuel-budget-forecaster-guide`: shipped, not in nav, nothing links them |
| `check_tool_acceptance_due` | 36 | **8** | yes — 6 `gamesdesign.co.uk` + 2 `vonc.com` tool pages, live and never acceptance-tested |
> **CORRECTED 2026-08-03 (live verification): the "8 findings" cell above was WRONG — those
> were 8 newly visible CANDIDATES, and the live run filed 0 items for them.** The check has
> three more gates after eligibility (a current `doc_plans` criteria fence, a 7-day verdict
> window, an open-run dedup), and 7 of the 8 have no current criteria plan — a Tier-2
> concern, correctly skipped. This is the same candidates-are-not-findings error corrected
> for the orphan row of this very table, made again two rows down. Logged in WRONG_CALLS.
| `check_backend_entry_orphaned` | 34 | scan widens by 34 components | unknown until it runs — its findings come from a live HTTP probe, so no honest prediction is possible |

**22 → 2 and 36 → 8 is the point.** The first query counted rows the check would *see*;
the findings are what survives the rest of its WHERE clause. Reporting "22 new orphan
findings" would have been wrong by an order of magnitude, and it is the same
candidates-are-not-findings error the estate keeps paying for.

## NOT converged, and this is a decision rather than an omission

- **`check_unresolved_sections`** — its filter is not a liveness test. It flips
  `deployed → needs_rebuild`; a page already `needs_rebuild` is already flagged, so
  converging adds `updated_at` churn and **no information** (and `updated_at` churn moves
  rows around `pre_query` orderings elsewhere).
- **`check_page_component_status_drift`** — here `= 'deployed'` is *correct*. Drift means
  "the page claims to be fully deployed but a component is not". A `needs_rebuild` page
  having non-deployed components is the **expected** state, so converging would
  manufacture 2 false positives. Measured, not assumed.

## Tranche 2, held back deliberately

`maintenance_actions.go:723,750` · `render_news_section_html.go:77` ·
`render_directory_action.go:345` · `request_render_audit_action.go:110` ·
`store_generated_component_action.go:843` · `component_library.go:2378,2434`.

These are **rerender queuers and listing selectors**, not detectors, and the question is
genuinely ambiguous: a shipped `needs_rebuild` page is serving a stale artefact, so a
rerender would fix it — but the page is already queued for a rebuild that will regenerate
it anyway. Converging them adds work items for pages already flagged. **That needs its own
measurement (how long does a `needs_rebuild` page actually wait?), which is a different
question from this one.** Held rather than swept in, per this file's own "do not make them
all identical".

## Still open

- Tranche 2 above.
- **Fix candidate 2** — merging `NeverDeployedPagePredicate` with
  `queryresolve.FetchablePageEligibilitySQL`. The aliased builder removes the *excuse* for
  the split (the alias) but not the split itself; `queryresolve` documents a deliberate
  family of three, so read `queryresolve.go:210-236` before touching it.
- **Nothing is live yet.** Committed only; ships on the next chassis roll.


---

# COUNCIL APPROVED 2026-08-03 (`66d07687`), and four seats found the same gap in my census

**Verdict: approved, 6 advisory objections, none high-severity.** Four seats
(`editquality`, `reuse_agent`, `debug_historian`, `prior_art_librarian`) independently
noted that the `LANDMINES` entry for this defect names three symbols my tranche-2 list did
not: `pageHasBeenLive`, `realisedPageIsBuilt`, `findStrandedNavPages`. Each checked:

| symbol | verdict |
|---|---|
| `pageHasBeenLive` | **GONE** — it WAS the hand-rolled predicate, deleted 2026-08-02. **The landmine's footprint was stale and sent four reviewers hunting a defect that no longer exists.** Corrected: the footprint now names the reusable builders instead. *When you fix the thing a landmine guards, re-read the landmine.* |
| `findStrandedNavPages` (`check_news_feed.go:690`) | **correct as-is** — `bugs_closed/015`'s stranded set, deliberately the NEGATIVE direction; `bugs_closed/081` records why. Now allow-listed *with that reason* rather than left implicit. |
| `realisedPageIsBuilt` (`v3_site_actions.go:5113`) | **still `status == "deployed"`, and it is NOT what `bugs_closed/037` fixed** — 037 shipped a separate `realisedPageCompositionIsPreserved`, which correctly covers `deployed` AND `needs_rebuild`. See below. |

## `realisedPageIsBuilt` — a content-injection path, with ~zero live exposure

Its two callers (`v3_site_actions.go:5031`, `:5064`) both ask one question: **is this
page's EMPTINESS authoritative?** — i.e. did it deploy empty because it renders through
another subsystem (a tool or blog-index page), or is it merely uncomposed? A shipped
`needs_rebuild` page's emptiness is authoritative by exactly the same argument, so the
narrow test misreads it as uncomposed and **lets the planner inject a generic layout into
a live page**. That is corruption, not a missed check.

**Measured before proposing anything:** pages that are shipped, not `deployed`, and
sectionless = **1 row**, and it is `archived` (`robot-hands.com/learning-center-article`).
So the live exposure today is effectively **nil**.

**Not fixed here, and the reason is mechanical rather than a judgement call:** the realised
page arrives as a `map[string]interface{}` that **carries no `deployed_at`** (checked —
the loading query does not select it). A correct fix therefore needs a query change plus a
Go change in the planner's hottest path, for one archived row. **Filed as tranche 3 in this
file rather than bolted onto a batch the council has already approved.**

## The other advisories

- **`bug_historian` [medium]: "nothing stops the next call site re-typing it by hand — which is
  verbatim how this bug arose."** Right, and now closed: `check_handrolled_shipped_predicate`
  in `scripts/pattern-check.py`. **MEASURED: 11 raw matches → 3 false positives allow-listed
  with reasons (an `ORDER BY` ranking twice, and `queryresolve`'s correct disjunct) → 8
  genuine, every one in the tranche-2 holdout set, 0 elsewhere.** The rule was also proved
  to FIRE on both a holdout and a synthetic new site, and to stay quiet on
  `page_components`, on writes, and on the correct spelling.
- **`guardian` [medium]: const→var is an exported-signature change asserted, not evidenced.**
  Evidence on the record: `grep -rn "NeverDeployedPagePredicate"` returns 30 references,
  every one either a comment, a `strings.Contains` test, or a runtime concatenation inside a
  function body; **zero appear in a `const (` block**, which is the only context a `var`
  would break.
- **`prior_art_librarian` / `reuse_agent` [low]: is `linkablePageStatusPredicate` the same
  judgement?** No — checked: it is `status NOT IN ('deleted','archived')`, the **page.status
  axis**, about whether a page may be OFFERED as a link target. This bug is the
  `build_status`/`deployed_at` axis, about whether a page is being SERVED. A page can be
  active and never shipped, or archived and still served (`bugs_open/098`). Different
  questions; both matter; neither substitutes.
- **`bug_historian` / `guardian` [low]: `check_backend_entry_orphaned`'s widened scan drives an
  unbounded live HTTP probe (0–34 findings) and is bundled with two measured edits.** Fair.
  It stays in the batch — it is read-only and the check already runs against these domains —
  but it is named here so a post-roll burst is attributable rather than mysterious.
- **`editquality` [low]: two overlapping guard tests on one property.**
  `links_deployment_test.go` guards the bare constant, the new file guards the builders.
  Deliberate: the constant is now derived, so its test is the one that would catch a
  hand-written literal creeping back into the derivation, while the builders' test is what
  new call sites will actually violate. Both cheap; neither redundant.

## Tranche 3, added by this round

`realisedPageIsBuilt` (above) — needs the loader to select `deployed_at` before the Go test
can be made honest. 1 archived row of exposure, so priority is low and the reason is
recorded rather than the work being silently dropped.

## Verification dispatched 2026-08-03 20:xx UTC — tranche 1 exercised live

Chassis `v1.0.1243` pod-verified carrying `NeverDeployedPagePredicateFor` /
`PageHasShippedPredicateFor` (both replicas). `improvement-loop` dispatched by hand
(`091_TRIGGER_improvement_loop_single_site.sh` pattern) against the three sites the
measurement predicted findings for:

| domain | site_id | correlation |
|---|---|---|
| gamesdesign.co.uk | `e33263f4-74f8-494f-b191-546845dbbddf` | `b8aeed5a-90cb-4753-9664-2af9219b4d1d` |
| gaswholesalers.com | `5fe15466-4e2e-4ff2-981e-98c1b7074002` | `d1641c1e-4b3e-4bc0-a13a-14ba46bfb954` |
| vonc.com | `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74` | `5029a34b-e90b-4c8e-9ee8-7ff36f1f2e4f` |

**Expect:** `orphan_blog_posts` on gaswholesalers.com (2, per the both-ways diff), and
`acceptance_run` on gamesdesign.co.uk (6) + vonc.com (2). Verify with:

```sql
SELECT s.domain, swi.item_type, swi.summary, swi.created_at
FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
WHERE swi.item_type IN ('orphan_blog_posts','acceptance_run','backend_entry_orphaned')
  AND swi.created_at > '2026-08-03 19:05:00'
ORDER BY swi.created_at DESC;
```


---

# TRANCHE 2 DECIDED AND DONE 2026-08-03 — the measurement inverted the holdout's premise

**Commit `9bd75a55f` · council `b563a61c-bbf3-40e6-9b2f-a32d6e71e964`** (submitted after
the commit, so the commit carries no trailer — deliberately no placeholder this time; the
correlation lives here, which is the honest forward record).

The holdout's stated reason was *"a shipped needs_rebuild page is already queued for a
rebuild that will regenerate it anyway"*. Measured (the question this file said tranche 2
needed):

```
41 shipped needs_rebuild pages now waiting:
p25 32.1h · p50 34.3h · p75 447.5h · p90 2412.1h · 13 over a week · 9 over a MONTH
```

**"It will rebuild soon anyway" is false for a quarter of the population** — the p75 waits
18+ days, the worst 102 days. During that window the page serves its stale artefact, so a
selector that skips it skips a live page for weeks. Decision per site, each with its own
evidence:

## Converged (3)

| site | delta today | why |
|---|---|---|
| `request_render_audit_action.go` | **36 pages / 8 sites** newly auditable | its own comment says the intent — "has a live URL" — and the predicate now says what the comment meant. A render audit that never photographs a page serving for 18 days is blind where it matters most |
| `render_news_section_html.go` (news rerender queuer) | 0 today, prospective | a shipped page carrying a news listing serves it; refreshing the listing on a page whose rebuild is weeks away is exactly the queuer's job |
| `render_directory_action.go` (directory queuer) | 0 today, prospective | same, byte-for-byte the news queuer's cousin |

**Collision with the 098 lane, handled rather than stumbled into:** the retraction lane
shipped `p.status = 'active'` into BOTH queuers the same day (an archived page was being
re-rendered twice daily), with tests pinning the literal old spelling. My convergence
landed on top of theirs — both filters now present — and their tests were updated to pin
the same intent under the new spelling, then **proved both ways**: reverting the artifact
predicate fails them, and dropping their status filter fails them. Neither lane's guard
was weakened by the other's edit, which is the same-file-passenger risk actually managed.

## Kept, with the decision recorded in the pattern-check allow-list (3 files)

- `maintenance_actions.go` — `findStalePages` flags pages for refresh; a `needs_rebuild`
  page is already flagged, converging double-queues it. `findPagesWithNoContent` overlaps
  `check_componentless_pages` (PBP-025), which covers the shipped-not-deployed case with
  its own deliberate predicate.
- `store_generated_component_action.go` — `markPagesForRebuild` flips
  `deployed → needs_rebuild`; a page already flagged gains nothing but `updated_at` churn.
- `component_library.go` — `GetHeaderNavFromPages`/`GetFooterNavFromPages` are **dead**:
  both call sites commented out, superseded by `nav_tables.go` which already uses the
  shared predicate. Left for the nav lane to delete, not converged.

**Post-tranche-2 fleet measurement: `check_handrolled_shipped_predicate` finds 0 genuine
hits.** Every `pages` liveness test in the tree is now either the shared predicate or a
recorded decision.

## Risk carried forward

The render audit change means up to 36 more pages enter audit rotation (capped at 25/site
per run, ordered by nav_order). If the audit starts photographing pages that then fail
checks, those findings are old blindness surfacing, not new defects — same caveat as
tranche 1's.


---

# TRANCHE 1 LIVE VERIFICATION COMPLETE 2026-08-03 — one exact proof, one falsified prediction, one pre-existing bug found

The three dispatched improvement loops completed (corrs `b8aeed5a` / `d1641c1e` / `5029a34b`).

## `check_orphan_pages` — PROVEN, with exact both-ways attribution

gaswholesalers.com had **no `orphan_blog_posts` item before** this run; it now has one
saying **"3 blog posts deployed but not linked"**. The check's own full query re-run both
ways at verification time: **NEW predicate → 3 blog orphans · OLD predicate → 0.** The
filed count matches the new predicate exactly, and matches the census (gaswholesalers has
exactly 3 shipped `needs_rebuild` blog-posts). This finding is caused by the convergence,
end to end: predicate → check → work item.

## `check_tool_acceptance_due` — predicate live, prediction falsified, and the reason is a finding

**0 of the 8 predicted tools filed.** Not a defect in the convergence — the check has
three further gates, and the census behind "8" modeled only eligibility:

- **7 of 8 have no current `doc_plans` criteria fence** → correctly skipped (that is
  Tier-2's `needs_criteria` concern, the check's own doc comment says so).
- **The 8th (`tool-archetype-taster-quiz`, vonc.com) is blocked by a pre-existing
  SUBJECT-KEY MISMATCH:** its plan is filed under `tool-archetype-taster-quiz`, but the
  component is `component_level='section'`, so `toolSubjectKeyExpr` strips the page
  prefix and looks up `archetype-taster-quiz` — no row. **Censused fleet-wide: exactly 1
  of 24 current tool plans is unreachable by the ladder's key.** Single row, not a class;
  it predates this change (the tool was invisible to the check entirely before the
  convergence, so the mismatch could never fire). Belongs to the tool-acceptance lane —
  recorded here, not fixed here.

The two acceptance items that DID file this run (`tool-spawn-rate-balancer`,
`gauntlet-round-record`) are ordinary due-runs on `deployed` pages — explicitly NOT
claimed as effects of this change.

## `check_backend_entry_orphaned` — ran, no findings

The widened scan (34 more components fleet-wide) produced no `backend_entry_orphaned`
items on the three sites. Its findings come from a live HTTP probe, so 0 is a valid
outcome, not evidence of a broken check — the probe found nothing orphaned on these
domains.

## The repeat error, named

The "8 findings" prediction was candidates-≠-findings again — the exact error corrected
for the orphan row of the same table, two rows up, in the same session. A both-ways diff
of a check's *eligibility clause* predicts what the check can SEE, never what it FILES;
every gate after eligibility must be either modeled or named as unmodeled. WRONG_CALLS
carries the tally entry.
