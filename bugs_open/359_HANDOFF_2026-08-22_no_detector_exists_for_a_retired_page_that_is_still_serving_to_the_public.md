# 359 — nothing detects a RETIRED page that is still serving to the public, and one has been live and unnoticed for at least 8 days

**Filed 2026-08-22** by the `bugfix_356_orphan_check_lifecycle_axis` lane, at the owner's
direction, on closing that lane. **Status: OPEN, UNOWNED.**

**This is an ABSENCE claim plus measured damage.** The absence is scoped and the damage is
first-hand with controls — see §2 and §5, and read §5 before treating the absence as settled.

## 1. The one-paragraph version

`pages.status = 'archived'` means *the platform has retired this page*. It does **not** mean the
page has stopped being served: retirement sets `status` and leaves the deployed artefact where
it is, so unpublishing is a separate act (`bugs_closed/098`). **Nothing anywhere checks whether
the two agree.** No discovery check, no cron, no scheduled task asks "is a page we retired still
answering 200 to the public?" — so the gap between *retired* and *retracted* is unobserved by
construction, and a page can sit in it indefinitely. One has.

## 2. The damage, measured 2026-08-22 with a fabricated-URL control per domain

37 pages are `status='archived'` with a non-null `deployed_at` on public domains. Sampled live
(control on the same domain each time, so the check could come out negative — and on
dartsonline it did):

```
CODE   BYTES     URL
404    2744      dartsonline.com/shipping-returns.html                      <- genuinely gone
404    2744      dartsonline.com/…-356-control.html                         (control, same size)
200    33326     loancalculator.co.uk/blog/loan-faqs.html                   <- ARCHIVED AND SERVING
404    1201      loancalculator.co.uk/blog/…-356-control.html               (control)
200    29997     loancalculator.co.uk/blog/jargon-buster.html               <- ARCHIVED AND SERVING
200    30997     robot-hands.com/gripper-catalog.html                       <- ARCHIVED AND SERVING
404    2886      robot-hands.com/…-356-control.html                         (control)
```

**The controls are what make this evidence rather than an observation:** dartsonline's archived
page returns 404 at exactly the same byte size as its control, so that domain's catch-all is
answering — and the three 200s are therefore real pages, not a permissive router.

### The load-bearing datum: one of them has been like this for at least 8 days, on the record

`robot-hands.com/gripper-catalog.html` is **30997 bytes today**. `bugs_open/266`'s note of
**2026-08-14** recorded it at **30997 bytes**, archived and serving, with the same control
pattern. Eight days, byte-identical, and **nothing raised anything** — no work item, no
`agent_error_log` row, no doc note. That is not a page slipping through a window; it is the
absence of any window at all.

⚠ **Do not read "3 of 7 sampled" as a rate.** The sample was chosen for recency and public
domains, not randomly, and the true denominator question ("how many of the 37 serve?") is the
first thing a fixing thread should measure properly — §6 candidate 1.

## 3. Why this is a real defect and not a tidy-up

- **It is the exact state `bugs_open/266` exists to prevent, surviving 266's own fix.** 266
  stopped archived pages being *re-deployed* (the `ARCHIVED_PAGE_GUARD` at `git_commit`, live
  and behaviourally proven). A guard on the deploy seam cannot retract a page that was already
  deployed before it was archived — so 266's fix is correct and this residue is outside it.
- **`bugs_closed/098` built the retraction, and nothing checks it RAN.** Retraction exists
  (`retract_page_deployment_action.go`, `retract_page_graph.go`) and is dispatched by hand or by
  a workflow. There is no reconciliation asking whether every archived page has actually been
  retracted, so a retraction that never fired, failed, or was never requested is indistinguishable
  from one that succeeded — a `complete` work item is not a retracted page.
- **It is publishing content we decided to stop publishing.** Whatever the reason for retiring a
  page — wrong, superseded, off-strategy, a claim we could not evidence — that reason still
  applies while it serves. `check_unverified_claims` deliberately keeps auditing archived pages
  for exactly this reason, which is an acknowledgement that the state is real and dangerous, not
  a substitute for detecting it.

## 4. Why nobody noticed — the two things that look like coverage and are not

1. **`check_unverified_claims` audits archived-and-serving pages** and is right to
   (`check_unverified_claims.go:458`, mutation-tested). But it audits their **copy**. It answers
   "is this page making an unsupported claim", never "should this page be here at all". A clean
   verdict from it is not evidence the page should be live.
2. **`ARCHIVED_PAGE_GUARD` fires and is counted** (`agent_error_log`,
   `ARCHIVED_PAGE_DEPLOY_REFUSED` — 20 rows on 2026-08-14). Those refusals look like the system
   handling archived pages, and they are — at the **write** seam. A page nobody is trying to
   rebuild generates no refusals and is invisible to that counter. **The quiet case is the
   undetected one**, which is the wrong way round.

## 5. The absence claim, scoped — this is what I checked, and what I did not

`[MEASURED 2026-08-22]`, and stated as scope rather than as a universal:

- **`discovery_checks/`**: no check has this subject. All 71 non-test files were read while
  filing `bugs_open/356`; the registry there classifies every one of the 51 that query `pages`.
- **Go, repo-wide**: 7 non-test files mention `status = 'archived'` / `archivedPageStatus`. Of
  those, 3 also perform HTTP — `retract_page_deployment_action.go` (performs retraction, does
  not look for the drift), `check_asset_reference_404.go` (asset references, and its page query
  is lifecycle-armed so it excludes archived pages), `refresh_evidence_fact_drift.go` (evidence
  facts). None asks this question.
- **Cluster**: `scheduled_tasks` has no row whose name matches `archiv`/`retract`/`unpublish`.

**What I did NOT check, so do not cite this as exhaustive:** SQL-only mechanisms with unrelated
names, anything in the admin dashboard or an adapter, and any detector whose name shares no
vocabulary with the concept. `prior_art_librarian`'s standing rule applies — **an absence is true
only when and where you looked**, and this is where I looked.

## 6. Fix candidates, ordered by what closes the door

1. **Measure the real population first.** Fetch all 37 with per-domain controls before designing
   anything. The three cases above may be three, or may be most of them, and the fix's shape
   (reconcile-and-retract vs flag-for-a-human) depends on which. Cheap, and it is the census arm
   any fix will be verified against.
2. **A discovery check, `archived_page_still_serving`** — the natural home, and the precedent is
   exact: `check_asset_reference_404` already does outbound fetches from the discovery path, so
   the "no outbound probe on the completion path" objection does not apply here. Route it to
   **`handler_agent: ""` initially** — a finding a human reads — because auto-retracting a live
   URL is not a thing to arm on first measurement.
   ⚠ Its page query must be **deliberately UNARMED** on the lifecycle axis and declared
   `PostureObserves` in `page_lifecycle_posture_test.go` (register WII-025) — a check whose
   entire subject is archived pages must see them. That registry will make you declare it.
3. **Or reconcile at the retraction seam** — for every archived page, assert a retraction was
   requested and completed. Cheaper (no fetch) and catches the "never asked" case, but it cannot
   see a retraction that reported success and did not work, which is `bugs_closed/098`'s own
   failure shape. **Weaker than 2; consider both, they answer different questions.**
4. **Do NOT auto-delete on detection.** A wrongly-archived page that is serving correctly is a
   real possibility (an archive is a status write, and `bugs_closed/215`'s canonicalisation
   collisions show statuses get set by accident), and un-publishing a good live page is the
   failure this estate calls "worse than the bug".

## 7. How to verify a fix

- **Census arm:** every archived page on a public domain either 404s or carries a recorded,
  dated reason for still serving. Today: **at least 3 serve with no such record.**
- **Disconfirming pair, and it must be a PAIR:** the detector flags
  `robot-hands.com/gripper-catalog.html` (archived, serving, 30997 B) **and does NOT flag**
  `dartsonline.com/shipping-returns.html` (archived, already 404). A detector that flags every
  archived page satisfies the first arm and is useless.
- **Control the instrument, not just the result:** a fabricated same-domain URL must 404 in the
  same run. Without it a permissive router makes every page look live — and on dartsonline the
  archived page and the control return the same 404 at the same byte count, which is what a
  correctly-absent page looks like.
- **Not valid:** a zero from the detector before it has ever flagged the three known-live cases.
  A silent detector and a clean estate are the same reading (016b §9).

## 8. Filing basis (owner ruling 2026-07-31)

**No `090` run, and the substitution is stated plainly.** This file asserts no new mechanism and
no cross-cutting root cause — the mechanism (`archived` ≠ retracted) is `bugs_closed/098`'s,
already diagnosed and fixed; what is new is an **absence** plus **first-hand artefact
measurement with controls**, both reproducible by the commands in §2. Two 090 runs from this
lane earlier today returned FAILED (kafka handshake race) and UNVERIFIABLE (scope-not-narrowing)
respectively, so a third on a claim the loop cannot settle — it has no outbound fetch — would
cost a round and answer nothing. **If a fixing thread asserts a root cause for WHY retraction did
not fire, that is a mechanism claim and should go through the loop first.**

## Related

`bugs_open/266` (archived pages re-deployed; its 2026-08-14 note is where `gripper-catalog` was
first recorded serving, and its deploy-seam guard is why this residue is out of its scope) ·
`bugs_closed/098` (archiving must retract — the retraction this file says nothing verifies) ·
`bugs_open/304` (retracting the last page of a site cannot unpublish it — an adjacent retraction
failure) · `bugs_open/356` §8 (where this gap was recorded as unowned, and the lane that found
it) · register **WII-025** (the posture registry a new check here must declare into) ·
`check_unverified_claims.go:458` (the one place that deliberately looks at these pages, and why
it is not coverage).

---

## 2026-08-26 — TAKEN by the `bugfix_359_archived_page_still_serving` lane. Re-measured; §6 candidate 1 is DISCHARGED

Ownership checked three ways before starting, all clear: `scripts/who-owns.py 359` returns
`likely OWNING workstream(s): (none identified)` (the only commit touching this file is the
filing commit `3f2891c75`); a grep of live session transcripts finds one hit, in the FILING
lane's own close-out ("filed `bugs_open/359`, ready to close"); and `site_work_items` holds no
open row of this shape.

### The bug is STILL VALID, and the population has MOVED

§6 candidate 1 asked for the real population before anything is designed. It is now a command —
**`scripts/audit-archived-still-serving.sh`** (`[--json] [<domain>…] | --self-test`), which
carries this file's §7 controls as its verdict logic rather than as advice.

`[MEASURED 2026-08-26]` **population 39 · 7 ARCHIVED AND SERVING · 32 correctly absent · 0
unjudgeable.** Every domain's invented-URL control returned 404 and every domain's known-good
`active`+shipped sibling returned 200, so both readings are real in both directions.

```
ai-agent-orchestration.com  /llm-cost-calculator.html
finetuning.uk               /tools/password-entropy.html            (archived 2026-08-25)
fundamentallyai.com         /blog/ai-readiness-checker-guide.html
fundamentallyai.com         /tools/llm-cost-calculator/index.html
leopardessconsulting.co.uk  /our-approach.html
robot-hands.com             /gripper-catalog.html                   (serving since ≥2026-08-14)
robot-hands.com             /news.html
```

**§2's "do not read 3 of 7 sampled as a rate" was right, and the true rate is 7 of 39 (18%).**

**The more useful correction is that this is a FLOW, not a backlog.** The two
`loancalculator.co.uk` blog pages this file recorded as serving on 2026-08-22 — `/blog/loan-faqs.html`
and `/blog/jargon-buster.html` — **both 404 today**, and **five of today's seven were not in that
sample at all**. So the set turns over inside four days. A one-off sweep of these seven would
read as a fix, change nothing, and be wrong again within a fortnight — which is the strongest
argument for §6 candidate 2 over any cleanup.

`robot-hands.com/gripper-catalog.html`, this file's load-bearing datum, has grown from eight days
to **twelve** and is unchanged.

### Three further measurements, because they change the fix's shape and its severity

1. **All seven are orphaned from their own site.** `site_nav_items` rows: 0 for all seven.
   `link_registry` inbound: 0 for all seven. They are reachable only by direct URL or by a search
   engine that indexed them before retirement — which is exactly why the class is invisible from
   inside the site.
2. **None of the seven is in its site's `sitemap.xml`** (checked live against each site's own
   sitemap, 25–51 `<loc>` entries each; 0 of 7 present). So we are not actively inviting
   indexing, we are failing to withdraw what was already indexed. That supports **medium**
   severity rather than high, and it is stated here so nobody has to re-derive it.
3. **None of the seven has a same-`url` sibling page row**, so the retraction action's
   active-page collision guard would refuse none of them. ⚠ `url` equality is the WEAK form of
   that test — the real rule is equality of the DERIVED FILE PATH
   (`datahelpers.PageFilePathFromURL`), because `/foo/` and `/foo/index.html` are one file. A
   detector must apply the derived-path form, not the string form.
4. **The re-deploy seam is closed, so a retraction will now stick.** `bugs_open/266`'s
   `ARCHIVED_PAGE_GUARD` is live at both deploy seams (`git_deployer_actions.go:81,103`,
   `v3_site_actions.go:899,911`, sharing `archived_page_guard.go`). This matters because
   LANDMINES records that retraction used to be **self-undoing** — delete the file, the next
   refresh republishes it, and a post-delete `curl` still shows 404 at the moment you look.

### One thing a fixing thread must know before touching the enablement half

Enabling a discovery check means adding its name to `agent_definitions`, and **an unregistered
name hard-fails the whole step** (`discovery_checks.go:198-216`, `bugs_open/149` B4), taking the
run's already-collected findings with it because the return precedes `tx.Commit()`. So the
migration must be held until the image carrying the Go file has rolled.

The build-enforced guard against that outage is `liveConfiguredChecks` in
`discovery_checks_registration_test.go`. `[MEASURED 2026-08-26]` **the live agents configure 82
distinct check names and that fixture asserts 63** — nineteen missing, including all three names
of a fifth agent (`acceptance-discovery-agent`) and `page_content_divergence`. All 82 resolve
today (verified by dumping `discovery_checks.Names()`), so this is an under-assertion and not a
live risk — but it is a 23% blind spot in the guard, and this lane refreshes it by union.

Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_359_archived_page_still_serving/`.
