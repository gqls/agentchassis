# 097 — CTA integrity misses in-body card links to pages that were never built

**Filed** 2026-07-26 from the oufe.com workstream.
**Severity** medium — publicly visible broken links on a live site, including a
main navigation item.
**Status** OPEN.
**Related** `bugs_open/052` (listings re-advertise unbuilt pages) — same family,
different surface: 052 is about generated listings, this is about hand-planned
in-body cards. Also `bugs_open/049` (broken links fleet-wide).

## Symptom

oufe.com went live with **six broken content links on the homepage**, and the
header's own **Cases** nav item 404ing. The detector raised exactly two items:

```
unresolved_cta | Unresolved CTA on index ('hero'): no real-page destination for secondary_cta_url
unresolved_cta | Unresolved CTA on index ('call-to-action'): no real-page destination for secondary_cta_url
```

Those two were real and correctly found. The six that were missed:

| card link | target |
|---|---|
| `/restructuring-plan` | never planned |
| `/creditor-waterfall` | never planned |
| `/cases/thames-water` | page exists but its url is `/blog/thames-water.html` |
| `/tools` | never planned |
| `/framework` | never planned |
| `/cases` | page exists but its url is `/cases/index.html` |

## Cause

The check covers the *named CTA fields* of components that declare them —
`cta_url`, `secondary_cta_url`, `primary_cta_url`. It does not walk **arrays of
child objects** inside `content_data`, and `info-card-grid` keeps its
destinations at `content_data.cards[*].link_url`.

Two of the six are especially worth noting, because they show the failure is not
only "page missing":

- `/cases/thames-water` and `/cases` were written from **the plan's intent**, not
  from the page record. The pages exist; their urls are `/blog/thames-water.html`
  and `/cases/index.html`. So a link can be broken even when its target page is
  built and deployed — checking "does a page with this name exist" is not
  sufficient, the check has to resolve against `pages.url`.

## Why the templates did not save us

The component templates are already fail-closed and correct:

```
info-card-grid:  {{if .link_url}}<a href="{{.link_url}}">…</a>{{end}}
hero:            {{if and .cta_text .cta_url}}<a href="{{.cta_url}}">…</a>{{end}}
```

A card with **no** url renders as plain text — good. But a card with a url that
points nowhere renders a perfectly well-formed anchor. The template can only
protect against a missing destination, never against a wrong one. That check has
to live where page urls are known.

## Fix candidates, ordered by what closes the door

1. **Resolve every internal href in rendered component HTML against `pages.url`**
   at the same point the CTA check runs. This catches the whole class regardless
   of which field or nesting the link came from, and it catches wrong-url as well
   as missing-page. It is also the only candidate that would have caught all six.
2. **Walk arrays in `content_data`** for `*_url` / `link_url` keys, in addition to
   the top-level named fields. Cheaper, but it enumerates shapes and the next
   component with a new nesting reopens the gap.
3. Add `link_url` to the named-field list. Fixes this one component only.

Candidate 1 makes the bad state detectable wherever it appears; 2 and 3 both
require someone to remember to extend a list, which is the shape that produced
this bug.

## How to verify a fix

On a site with a card grid, point one card at a path with no page and another at
a page whose `url` differs from the path used. Run the check. Both must be
flagged. Then confirm a correct link is **not** flagged — a checker that fires on
everything is as useless as one that fires on nothing.

**Induce both faults.** A clean report on a healthy site proves only that the
check runs.

## Second instance, different site — found 2026-07-26 while closing `bugs_closed/052`

`robot-hands.com/learning-center.html` is live, linked from the site's own
navigation, and carries five in-body category links that all 404. Verified with
`curl`, not inferred:

```
404  https://robot-hands.com/learning-center/calculators
404  https://robot-hands.com/learning-center/technology-guides
```

(the other three — `/learning-center/application-guides`,
`/learning-center/comparisons`, `/learning-center/specification-workflow` — are the
same shape and were not individually fetched. `[UNVERIFIED]`)

Two things this instance adds to the oufe.com evidence above:

- **It is not confined to one site or one component library.** Same failure, a
  different site built at a different time, so the missed-nesting cause is
  platform-wide rather than an oufe-specific component choice.
- **These links have no `.html` extension**, while every real page on the site is
  served at `<name>.html`. So the destination was written from the plan's intent
  rather than from `pages.url` — the same sub-cause as `/cases/thames-water` above,
  and further evidence for fix candidate 1 (resolve rendered hrefs against
  `pages.url`) over candidates 2 and 3, which would not catch a link whose *target
  page exists under a different url*.

Not investigated further here — this file owns the mechanism; recording the instance
so the fix's verification set includes a second site.

---

## Triage 2026-07-27, post-roll (v1.0.1174) — oufe is clean, robot-hands is not, and the detector gap is untouched

Verification sweep, not a fix. Both instances re-probed live over HTTPS.

**oufe.com's six are gone.** The homepage now emits only resolvable targets —
`/about.html`, `/cases/index.html`, `/contact.html`, `/index.html` (200 confirmed on the
two checked). `/restructuring-plan`, `/creditor-waterfall`, `/cases/thames-water`,
`/tools`, `/framework` and bare `/cases` are all absent from the served markup.

**robot-hands.com's are all still live, and there is a sixth this file missed.** Every one
individually probed, so the `[UNVERIFIED]` marker on three of them above is now discharged:

```
404  /learning-center/calculators            404  /learning-center/comparisons
404  /learning-center/technology-guides      404  /learning-center/specification-workflow
404  /learning-center/application-guides     404  /matchmatrix/methodology     <- NEW
```

`/matchmatrix/methodology` is the same shape from a different hub, and note the site *does*
serve `/matchmatrix-methodology.html` (200) — so it is the "target exists at a different
url" sub-class, not a pure phantom. That strengthens fix candidate 1 over 2 and 3 again.

### What changed underneath this bug, and what did not

`bugs_open/079`'s `RepairPageLinks` is now live and firing at the build gate
(`validate_page_content.go:356`), and it resolves **every** internal `<a href>` in the
assembled page against `pages.url` — which is literally this file's fix candidate 1, arriving
from another lane and at a different stage. On the numbers above it would have handled
oufe's six: bare `/cases` normalises to `/cases/index.html` and **rewrites**; the four true
phantoms **unlink**. `/cases/thames-water` (real page at `/blog/thames-water.html`) would
unlink rather than repoint — the 404 dies, the intended destination is still lost.

**Two reasons this file stays OPEN:**

1. **The detector is unchanged.** `ctaFieldNames`
   (`resolve_internal_links_action.go:98-105`) still lists six components, still has no
   `info-card-grid`, and nothing walks arrays in `content_data` — grep for `link_url` in that
   action returns nothing. So `unresolved_cta` still misses exactly what it missed, and the
   only thing now catching the class is a *repair* at a later stage that silently deletes the
   link instead of reporting it. Detection and repair are not substitutes.
2. **A write-path repair does not reach a deployed page.** robot-hands' six have been live
   since before the fix and will stay live until that page is rebuilt. There is no sweep that
   would find them (`bugs_open/083`: `improvement-sweep` off since 2026-05).

**Sharpest next action:** decide whether candidate 1 is now redundant at the gate and should
instead be implemented where this file said — at the CTA check — so the finding is *reported*
as well as repaired. Then a rebuild of `robot-hands.com/learning-center.html` clears the live
damage. Neither needs a diagnosis run; the cause is fully known.

> ## CORRECTION 2026-07-27 (bugs thread) — **"the cause is fully known" does not survive one query, and a diagnosis run is now filed**
>
> The line directly above says no diagnosis run is needed. Two measurements taken while
> picking this up say otherwise, and they are the reason a run was spent:
>
> **1. The page was REBUILT AND DEPLOYED on 2026-07-25, and kept its 404s.**
>
> ```
>      domain      |      name       | build_status | deployed_at
> -----------------+-----------------+--------------+------------
>  robot-hands.com | learning-center | deployed     | 2026-07-25
>  robot-hands.com | matchmatrix     | deployed     | 2026-07-24
> ```
>
> This is not a stale page that predates the fixes. It went through a build after
> `bugs_open/079`'s `RepairPageLinks` was described as "live and firing at the build gate",
> and the six unresolvable hrefs were neither rewritten, nor unlinked, nor reported. The
> triage above reasoned that the live damage persists because "a write-path repair does not
> reach a deployed page" — true in general, but **it does not explain this page**, which was
> written after the repair existed.
>
> **2. robot-hands.com has ZERO `phantom_internal_link` findings, ever.**
>
> ```sql
> SELECT s.domain, swi.item_type, swi.status, count(*)
> FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
> WHERE swi.item_type IN ('phantom_internal_link','unresolved_cta','misdirected_cta')
> GROUP BY 1,2,3;
> --  robot-hands.com | unresolved_cta | needs_human_review | 3   (07-03..07-06)
> --  ...and no phantom_internal_link row for this domain at all
> ```
>
> Five other domains DO carry `phantom_internal_link` rows (ai-agent-orchestration 7,
> idea.uk 9, relojistas 5, vonc 1), so the check is not globally dead — it just never
> produced a finding here. The newest such row fleet-wide is **2026-07-24**, one day BEFORE
> this page's deploy.
>
> **Why this matters for the fix.** This file's candidates all assume the gap is the
> *detector's field enumeration* (`ctaFieldNames` not walking arrays). That may still be
> true, but three mechanisms claim to cover this page and it passed all three, and nobody
> has established which was actually on the build path for the 07-25 deploy. Choosing
> between candidates 1/2/3 before knowing that is choosing a fix for a mechanism that may
> not be the one that failed — the failure mode CLAUDE.md's 2026-07-19 correction is about.
>
> **Diagnosis run filed:** correlation `3002a141-cb3d-42f2-b4d6-a288177d06b5`, asking which
> of `resolve_internal_links_action.go` (`ctaFieldNames`),
> `discovery_checks/check_phantom_internal_links.go` and `datahelpers/link_repair.go`
> (`RepairPageLinks`) was on the path for that build, and why the hrefs survived it.
> **No code was written for this bug pending that answer.**
>
> > **BLOCKED 2026-07-27 21:55 — both diagnosis attempts are dead, and the cause is
> > `bugs_open/029`.**
> >
> > - **Attempt 1** (`3002a141`, fired 19:04) produced three evidence bundles and stopped
> >   at **19:25:10**. I killed it myself by rolling the chassis at 19:22 — logged in
> >   `WRONG_CALLS.md`, my fault, not this bug's.
> > - **Attempt 2** (`e1aa4695`, fired 20:47) **FAILED** at `spawn_diagnoser` 20:55:21:
> >   *"Request 01a36f8c-b85a-4e1a-9bec-a556a7f78ef3 timed out after 3 retries"*. Its
> >   worker pod came up at 20:48:27Z and has idled ever since, logging only "No activity
> >   for 5 minutes" — a live hung spawn. Full timeline contributed to `bugs_open/029`,
> >   and the pod deliberately left running there as a specimen.
> >
> > So the mechanism question this file needs answered — **which of the three mechanisms
> > was actually on the 07-25 build path** — is unanswered and cannot be asked again until
> > 029 lets a diagnoser run. Do not pick between fix candidates 1/2/3 before it is.
> >
> > **UPDATE 2026-07-28 — attempt 3 RAN, and returned `UNVERIFIABLE`. The blocker is now
> > `bugs_open/108`, not 029.**
> >
> > Correlation `914dc844-7dad-4d5a-8d1b-a9c4296880c4`, five iterations, ~254KB of
> > bundles, no answer. Two of its findings are worth keeping regardless:
> >
> > - **The phantom targets have no `pages` rows at all** —
> >   `/learning-center/technology-guides` and `/learning-center/application-guides`
> >   return 0 rows. So these are pure phantoms, not the "target exists at a different
> >   url" sub-class that `/matchmatrix/methodology` belongs to.
> > - **No `phantom_internal_link` OR `unbuilt_internal_link` work item exists for this
> >   site**, on a widened query across any `page_id`. The earlier finding holds under a
> >   broader predicate than I used.
> >
> > **Why it could not conclude.** It asked whether `RepairPageLinks` exists — load-bearing,
> > because the whole question is whether the gate's repair step was on the path — and the
> > code index answered *"0 rows … The query was RUN and found nothing; this is not an
> > unanswered question."* That is a **false zero**: the symbol is at
> > `link_repair.go:139` and is called at `validate_page_content.go:357`. The index holds
> > exactly one commit, `e19aa5d` of 2026-07-24, now **970 commits behind HEAD**, and
> > `link_repair.go` was added on 2026-07-26 — two days later. Full measurement in
> > `bugs_open/108`, which this run has now given a measured cost.
> >
> > **So the sequencing is: 108 before this.** A fourth diagnosis run against the same
> > index will fail the same way. Until the index is current, this question has to be
> > answered by hand — read the page-build pipeline's step list for the 07-25 deploy and
> > check whether `validate_page_content` was on it — or not asked of the loop at all.

---

**2026-07-28 ~11:45 (bugs thread 2): diagnosis re-fired — the 108 blocker is GONE.**
`bugs_closed/108` closed today: the code index now describes `d98010e8b` (2026-07-28)
with `RepairPageLinks` indexed (`symbol` hit count 1), and the freshness banner states
what it mirrors. The question is unchanged: which of the three mechanisms was on the
07-25 build path. New run: correlation `9543aaf1-765e-4ae6-abb9-dbbb4ac13e83`. Read it
via `diagnosis_artifacts` / `collected_data->'verdict'`; note a code-tier run with no
explicit subject writes no `doc_notes` row.

---

## 2026-07-28 12:4x — ATTEMPT 4: CONFIRMED. The answer: NONE of the three mechanisms is on the bulk path.

Diagnosis `9543aaf1` (COMPLETED 11:44, five citations, both evidence families):

- The 07-25 rebuild is `site_work_items` id `2647398d-f07c-4ea9-b635-bd831c441952`
  (source=`rerender-pages`, handler=`page-rerender`, page_id `69d73b19`, completed
  2026-07-25T01:44:25Z) — the **bulk** `RerenderSitePagesAction → rerenderSinglePage`
  path, not single-page, not discovery-triggered.
- `rerenderSinglePage`'s full post-processing chain is: `InjectHeader → InjectFooter →
  rerenderInjectContactInfo → rerenderCleanDoubleDoctype → StripToolDocHeader`.
  **`RepairPageLinks` is called nowhere in it** — it lives only inside
  `ValidatePageContentAction` (`validate_page_content.go`), wired to the separate
  `page-build-handler/validate_content` step, which this pipeline never invokes.
  Neither is `PrepareLinkContextAction` nor any discovery link checker. Broken links in
  the assembled sections pass straight to deploy with nothing to rewrite or block them.
- The cited code itself records the drift class: the bulk path also lacks the
  single-page path's `collectJSAssets` ("js_content assets are only emitted by
  single-page rerenders") — the two paths have been diverging for a while.

**Why attempt 3 failed and 4 succeeded:** attempt 3 (corr `914dc844`) was defeated by
the stale code index (`bugs_closed/108`) — `RepairPageLinks` postdated the indexed
commit and the confident zero sent it chasing a casing theory. The index was fixed and
reindexed at `d98010e8b` this morning; attempt 4 cited the call site directly.

**Fix direction (framework, not per-case):** the defect is a SIBLING-PATH DRIFT — the
single-page path has the rigorous post-processing; the bulk path re-implements a
subset. The fix that makes recurrence unrepresentable is ONE shared post-processing
pipeline both paths call (repair + validation included), not adding one missing call
to the bulk copy (which leaves the next divergence open). Landmine from memory: do NOT
wire `InjectLinkConstraints` — it is a dead duplicate (see bugfix 092 notes).

---

## 2026-07-28 13:1x — THE REPAIR HALF IS FIXED, LIVE, AND VERIFIED ON THE DAMAGED PAGE. The detector half stays open.

**Built** (commit `c18f6f430`, live chassis `v1.0.1187`, digest-verified after a live
same-tag collision at 1186): `repairOutboundPageLinks` — one shared seam applying the
build gate's `RepairPageLinks` (same `loadValidPagePaths` index) at
HTML-leaves-for-deploy on BOTH rerender paths. `writeLinkRepairLog` generalised with a
`linkRepairOrigin` so the durable record names WHICH path repaired; the gate's rows are
byte-identical to before.

**Verified end to end on the live damage** (single-page rerender corr `c9dc739a`,
COMPLETED 13:11):

| check | result |
|---|---|
| `agent_error_log` row `agent_type='page-rerender'`, code `CONTENT_LINK_REPAIR_DETAIL` | **"Repaired 6 dead internal link(s): 0 rewritten, 6 removed", page learning-center** — a row shape structurally impossible on any pre-1187 binary |
| served `learning-center.html`, the five `/learning-center/*` hrefs | **5 → 0** |
| served `learning-center.html`, `/matchmatrix/methodology` href | **1 → 0** |
| page integrity | 30,103 bytes, anchor text preserved as plain text (unlink keeps body prose) |

The remaining three pages carrying `/matchmatrix/methodology` (`gripper-catalog`,
`pneumatic-vs-electric-grippers`, `selection-guide` — a wider spread than this file
knew) are being cleared by the same dispatch.

**Correction to the diagnosis, which changes nothing about the fix:**
`RerenderSitePagesAction` — the "bulk path" the verdict named — is **unregistered dead
code**: no registry entry, no `agent_definitions` reference, no callers. The live bulk
mechanism is `create_rerender_items` → per-page work items → `page-rerender` running
`RerenderSinglePageAction`. So the 07-25 rebuild went through the SINGLE-page action,
which equally lacked the repair. The mechanism finding ("no rerender path validates or
repairs") holds; the function attribution was wrong. The fix covered both paths plus
the dead one, so the correction costs nothing — but it is why fixing the drift CLASS
beats fixing the named site.

**Council: could not sit.** Submission `fcd4322b` died at its first seat — the
Anthropic API spend cap is exhausted until 2026-08-01 00:00 UTC. The gate is advisory;
committed without trailer per standing practice. **Resubmit with
`RESUBMIT_CORR=fcd4322b` when LLM access returns.**

**WHY THIS FILE STAYS OPEN — the headline defect is untouched:** `ctaFieldNames`
(`resolve_internal_links_action.go`) still enumerates named CTA fields and walks no
`content_data` arrays, so `unresolved_cta` still cannot SEE `info-card-grid` card
links; and `check_phantom_internal_links` has still never produced a finding for
robot-hands.com (mechanism unestablished — possibly 083's dead sweep). Detection and
repair are not substitutes: the repair now deletes a phantom link silently, and the
authoring defect that wrote it goes unreported. Fix candidate 1 (resolve rendered
hrefs against `pages.url` at the CHECK, reporting findings) remains the open work.

**Site-wide sweep result (all four affected pages rebuilt through the repaired path,
13:1x–13:2x):** learning-center 0 rewritten/6 unlinked · gripper-catalog 1/4 ·
pneumatic-vs-electric-grippers 0/6 · selection-guide 2/4 — **23 dead links repaired**,
both repair arms exercised live (3 rewrites to real stored URLs, 20 phantoms unlinked).
All four pages re-probed over HTTPS: 200, zero remaining `/matchmatrix/methodology` or
`/learning-center/*` phantom hrefs. The live damage this file tracked is CLEARED; the
detector gap (headline) remains the open work.

---

## 2026-07-28 15:2x — council APPROVED on resubmit; the site's FIRST discovery audit ran; one live phantom remains and it is another family's bug

**Council trail `fcd4322b`, round 2 (post cap-raise): APPROVED** — 11 reviewers, 0
unreadable. The code had already shipped (advisory gate, spend-cap window), so the
commit stays trailer-less — a known 098 false negative. Objections acted on:
bug_historian's medium (fail-open skip left no durable record) fixed forward in
`e60f5ef59` — the skip branch now writes `CONTENT_LINK_REPAIR_SKIPPED` to
`agent_error_log` (inert until the next roll). Its second medium stands as an **open
sub-question**: are the two rerender actions plus the build gate the COMPLETE set of
paths that ship page HTML? Not claimed; the origin-stamped log at least makes every
covered path visible.

**The dispatcher gap, proven and then exercised.** `check_phantom_internal_links`
(LNK-009) is live in `completeness-discovery-agent` but has NO fleet dispatcher — the
four covered domains each got a hand-fired one-shot; robot-hands had simply never been
audited. Fired its first: **24 phantom findings across 5 pages** (plus 8
cta_names_unknown_destination, 8 needs_internal_links, 20 page_rerender items, nav
drift). Note the check reads **stored** `page_components.rendered_html`, so its
findings describe authoring debt (`content_data` still holds the bad hrefs); the four
repaired pages are outbound-clean while their stored rows stay dirty by design. The
queue rows stand as the record — the `detected`-unclaimable drain is `083`'s open
problem, deliberately not worked around here.

**The one remaining LIVE phantom is not this bug.** `learning-center-index`
(`/learning-center/index.html`) serves a **May deploy**: `build_status='needs_rebuild'`
since then, components untouched since 07-03; the stitch-rerender legitimately
no-opped (empty final_result). Its phantom href targets `learning-center-article` —
**`status='archived'` yet once-deployed and now 404** (`bugs_open/098`'s exact shape:
archiving does not undeploy). A stale section-index page needing a real rebuild is the
`080`/`081` family with its standing owner call. Routed there; not brute-forced here.

**Remaining open work in this file:** the `ctaFieldNames` enumeration gap (candidate
1/2 at the CHECK stage), the deploy-path completeness sub-question above, and a
fleet dispatcher decision for the discovery audit (owner-adjacent: it files rows into
a queue with no drain — `083`/`033`).
