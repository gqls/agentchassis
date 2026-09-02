# 315 — `pages.deployed_at` is stamped whether or not the object is written, and one page has now been skipped by FOUR completed rerenders

**Filed 2026-08-18** by the `webdesign_tool_rebuilds` lane.
~~**Status: CLOSED 2026-08-21 — fixed AND live AND proven at the artefact.**~~
**Status: REOPENED 2026-09-02 (owner ruling) — the STAMP half stayed fixed; the SKIP half is live. Handed to the `site_ai_agent_orchestration` lane (session "AI page 3").**

> ## ⚠ REOPENED 2026-09-02 — §3's shape is live on a different page, and the fix that closed this file cannot reach it
>
> Found by the `gamedesign.uk` lane's independent re-investigation (Fable, 2026-09-02) while
> tracing a fleet-wide born-empty rerender wave of **2026-04-18**; spot-checked and pinned by
> that lane. Owner ruling, same day: *"315 shaped defect — reopen it and hand it to that site's
> lane to fix."* Cross-reference: `bugs_open/432` §3a (where it was first recorded).
>
> **The instance, all `[MEASURED 2026-09-02]`:**
>
> `https://ai-agent-orchestration.com/roi-estimator.html` serves **200, 12,138 bytes,
> `<main>\n\n</main>`** — header, footer, nothing between. Same-domain invented-URL control:
> **404**. Row `94b5f2db-585b-4266-a378-ade5aee1d1a9`: `status=active`, `build_status=deployed`,
> `deployed_at=2026-05-02 17:14:18`, `content_hash IS NULL`, **0 `page_components` rows**.
> **Nine `page_rerender` items `complete`** on it, first filed 2026-08-26, latest 2026-09-02 —
> plus `content_rewrite:unresolved`, `content_rewrite:deferred`, `needs_copy_edit:deferred`,
> `canonical_mismatch:detected`, `head_essentials_missing:detected`. The object was born empty
> in the 04-18 wave (13 files across four sites that day) and has never held content.
>
> **What held from the close — read this before re-litigating candidates 1/2/4:** `deployed_at`
> did NOT move across nine rerenders (still 05-02), and `content_hash` was never written — so
> candidate 2's "refuse the stamp on a reported skip" is doing exactly its job. This is not a
> regression of the fix. It is the half the fix was never aimed at.
>
> **What is live:** a `page_rerender` on a page with **zero component rows** is a *guaranteed*
> skip — `assemblePage` returns `""` (`rerender_single_page_action.go`, `d777cb4d2`), the step
> returns `skipped: true`, `page-rerender`'s `check_skipped` routes to `complete_skipped`, and
> the item closes **`complete`**. The rotation re-files it; it completes again. Nine times. The
> page needs a **BUILD** (`needs_content_page` — there is nothing to assemble), not a rerender,
> and nothing routes a 0-component page from the rerender queue to the build queue. The
> `content_rewrite` items that *would* have written content are `unresolved`/`deferred`.
> Meanwhile an empty page serves 200 under an `active` row, invisible to `bugs_open/349`
> (that is the **404** shape — never deployed) and to `page_content_divergence` (the served
> bytes DO match what was last committed; the defect is that what was committed is empty).
>
> **Two more from the same 04-18 wave still serve empty** (Fable, 2026-09-02, not re-pinned by
> the reopening lane): `ai-agent-orchestration.com/llm-cost-calculator.html` — row `archived`,
> deployed 04-18, **`bugs_closed/359`'s class, still serving**; and
> `robot-hands.com/learning-center-article.html` — row url is `/blog/…`, served at root, a
> URL-shape mismatch a `pages.url` probe would miss. Other lanes' sites; named here, not owned here.
>
> **What closes the reopen** (proposed, for the owning lane to accept or amend):
> 1. `/roi-estimator.html` serves a non-empty `<main>` (built, not rerendered), with
>    `content_hash` set and `deployed_at` moved — verified at the served bytes with the control.
> 2. A `page_rerender` filed against a page with **0 component rows** either (a) refuses at
>    filing, (b) converts to `needs_content_page`, or (c) closes as something that is NOT
>    `complete` — the choice is the owning lane's; the test is that nine completions on an
>    untouched page cannot happen again. `bugs_open/213`'s lesson applies: do not add a second
>    producer on an existing `item_type` without naming it in the register.
> 3. A census: `SELECT p.url FROM pages p WHERE p.status='active' AND p.build_status='deployed'
>    AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id=p.id)` — every row is a
>    candidate for this shape. **Count it and date it**; this file has one instance, not a rate.
>
> Everything below this block is the file as closed on 2026-08-21, unchanged.

## REOPEN ACCEPTED + FIX BUILT 2026-09-02 (same day) — "AI page 3" session

**Criteria accepted as proposed, with two amendments**: (a) criterion 1's "built and
serving" is driven through the framework's OWN planning route (below) — the page's
SHAPE (the title says *estimator*: interactive tool vs content page) is the
planner's/owner's call, not this lane's to hardcode; (b) the same-wave
`llm-cost-calculator.html` (row `archived`, still serving empty — 359's class) is
recorded here as an OWED DECISION (retract the served file vs un-archive and build),
surfaced to the owner rather than acted on solo.

**Every reopen figure re-verified at the artefact before acting** (200/12,138 B,
literal `<main></main>`, control 404, row active/deployed/hash-NULL/0 components,
nine `page_rerender` complete). **Producer census**: all nine filed by
`rerender-pages`' `create_rerender_items`, ONE item_key, 08-26→09-02.

**Criterion 3 census, dated 2026-09-02.**
> **CORRECTED after council round 1 (debug_historian, right):** my first figure —
> "2 rows" — used `status='active' AND build_status='deployed'`, the estate's
> LIVENESS predicate, which is wrong for a census (it excludes served-but-not-
> 'deployed' and archived-but-served pages). On the honest predicate
> `deployed_at IS NOT NULL AND NOT EXISTS(components)`: **14 rows** —
> active/needs_rebuild **5** (`check_componentless_pages`' target, which is built
> but enabled in ZERO discovery agents), archived/deployed **4** + archived/
> needs_rebuild **4** (`bugs_closed/359`'s still-served class, other lanes' pages),
> active/deployed **1** (roi-estimator itself). A small population — not one
> instance, and not a rate.

**Prior art, and why roi-estimator falls through it** (council round 1, reuse +
prior_art seats — both correct to ask):
- `check_sectionless_pages` matches `sections=[]` but is gated on CURRENT-PLAN
  membership AND needs a same-role sibling to borrow layout from. roi-estimator is
  in neither — it never fires.
- `check_componentless_pages` needs sections PRESENT (an intact array to build
  from). roi-estimator has `sections=[]`, so it cannot match — and, separately, it
  is enabled in **zero** discovery agents fleet-wide (a real adjacent gap, surfaced
  as an owed decision, NOT bundled into this fix — it would not catch this page
  anyway).
So roi-estimator sits in the genuine gap between the two checks. This fix operates
at a different layer (convert the futile rerender at file/skip time), complementary
to both.

**Criterion 2: BOTH doors closed** (commit `8eca969cb` + the writeWorkItem refactor,
`Council-Submitted: 2be8ec34-d905-4afc-9cf4-227c2facbc14`):
- CONSUMER — `RerenderSinglePageAction`'s 0-component skip now files a deduped
  `needs_content_page` (stable key `needs_content_page:<page_id>`) **through
  `writeWorkItem`** — not a hand-rolled INSERT; the workItem struct's own comment
  records three call sites that hand-rolled this and were caught by two council
  seats, and the doors in that seam (owned-page routability, growth posture,
  whatever lands next) must reach this producer unremembered. Skip semantics and
  item completion untouched; `build_ask_filed` reported in the step result.
- PRODUCER — `create_rerender_items`' unscoped loop no longer files rerenders for
  0-component pages; it files the SAME deduped build ask and reports
  `empty_pages_converted_to_build_asks` in the step RESULT (the function's own
  `unknown_reason` precedent: results are read, pod logs scroll).
- Producer guard MUTATION-PROVED (disabled → the action-level sqlmock test fails on
  ordering; restored → green, `-count=1`). Consumer WIRING has no unit test
  (action-level scaffolding too heavy) — its proof is the live motivating case
  post-roll: re-fire a rerender at this page while unbuilt → the ask files/dedups.
  **Inert until an agent-chassis roll; verify at the per-service stamp.**

## llm-cost-calculator (secondary instance) — the retract-vs-rebuild question is MOOT; corrected facts 2026-09-02

The reopen block named `ai-agent-orchestration.com/llm-cost-calculator.html` (archived,
04-18, serving empty) as a retract-vs-rebuild decision. Owner input relayed via the
gamedesign.uk session (2026-09-02 ~19:15Z), as a PRINCIPLE not a site-ruling (the peer
flagged that "this site" in his sentence may have meant gamedesign.uk): *"we don't want
an llm cost calculator on this site necessarily unless the story is all about using llms
... in which case it is relevant."*

**Before acting on that principle I checked the artefact, and it changes the question.**
`[MEASURED 2026-09-02]` ai-agent-orchestration.com has THREE llm-cost-calculator pages:
- `/tools/tool-llm-cost-calculator.html` — **active, deployed 2026-09-02, 1 component,
  serves 67,613 B of real content.** The canonical tool ALREADY EXISTS, built today by
  another process. So the owner's principle ("belongs where the story is about LLMs" —
  and this site's is) is ALREADY SATISFIED. "Un-archive and build" rests on a false
  premise; do not.
- `/llm-cost-calculator.html` — archived, 04-18, 0 components, serves 200/empty. A stale
  duplicate at the OLD flat URL — **and the homepage AND /tools.html LINK to it.** A
  naive retract would 404 live navigation.
- `/guides/tool-llm-cost-calculator-guide.html` — active/needs_rebuild, 04-18, 0
  components, serves 200/empty. A third stale shell.

**So the real, narrower action** (a nav/duplicate cleanup, NOT a build): repoint the
homepage + tools-listing links from `/llm-cost-calculator.html` to the built
`/tools/tool-llm-cost-calculator.html`, then retract the two empty shells. That is a
content/design decision for the owner or the site's design lane, **not acted on by this
315-mechanism thread** — recorded here and surfaced. Filing a build for the flat URL
would only duplicate the tool that already exists.

These two shells are the SAME 04-18 born-empty wave as roi-estimator, so the mechanism
fix above stops FUTURE rerenders of them churning; the existing shells need the
nav-repoint+retract, which no mechanism should do automatically.

## COUNCIL: r1 REVISE (objections right, all answered) -> r2 APPROVED 2026-09-02

Corr `2be8ec34`. Round 1's two HIGH objections (ON CONFLICT convention, prior-art
gap) are GONE in r2 — the writeWorkItem refactor and the gap analysis answered
them. r2 approved with 3 advisories, none high; dispositions:
- *sketch still shows a raw ON CONFLICT* (editquality/guidelines/prior_art, medium)
  — a SUBMISSION-vs-CODE mismatch: my r2 rationale said the INSERT was gone but I
  left the stale literal in the plan SKETCH. VERIFIED AT THE CODE, not asserted:
  `grep -c 'ON CONFLICT' rerender_single_page_action.go` = **0**; the helper calls
  `writeWorkItem(..., dropOnConflict, ...)` (line ~1300), whose own INSERT carries
  the literal `ON CONFLICT (site_id, item_key) WHERE ...` matching idx_swi_dedup
  (load_work_item_actions.go:2170). The convention is satisfied by the seam.
- *consumer wiring has no unit test* (editquality, medium) — TRUE and stated: the
  producer conversion is action-level tested + mutation-proved; the consumer
  skip-branch wiring's proof is the live motivating case post-roll (re-fire a
  rerender at an unbuilt 0-component page -> the ask files/dedups). Heavy
  scaffolding for a full RerenderSinglePageAction test; deferred deliberately.
- *componentless zero-count needs SQL* (prior_art, low) — re-proven:
  `SELECT count(*) FROM agent_definitions WHERE default_config::text LIKE
  '%componentless_pages%' AND is_active ...` = **0**.
- throughput of the per-page EXISTS probe (guardian, low): one indexed existence
  check per page in a bounded loop; acceptable.

Both `Council-Submitted` commits (8eca969cb, 8a0b927f5) are credited by 098 now
the chain is approved.

**Criterion 1, driven to the framework's own edge** (2026-09-02 ~17:20Z):
- Hand-filed `needs_content_page` (`a423f7ea…`, the code path's stable key so later
  code filings dedup) ran within the hour and parked HONESTLY at
  `needs_human_review` — `mark_no_ready_sections`: the page has `sections=[]` and
  no spec; nothing to build yet. That is the adoption-pipeline S2 route working.
- So the missing artefact is a PLAN: `needs_content_planning` filed
  (`a8bbf171…`, status `detected` → the promoter routes it; a handlerless triaged
  insert is refused by `swi_no_handlerless_promotable`, correctly). Its spec
  cross-names the parked build ask. When the planner writes sections, the build
  ask can be re-driven and criterion 1 closes at the served bytes with the control.

**Not fixed here, stated plainly**: the empty page still SERVES 200 today — it
stays live until the plan→build chain lands. The reopen block's two other named
instances stay with their notes: llm-cost-calculator = owner decision (above);
robot-hands `learning-center-article` (url-shape mismatch) — this session also
wears the robot-hands hat and takes it into that lane's list.


> **WHAT CLOSED IT.** All four fix candidates are resolved or deliberately re-scoped:
>
> - **Candidate 1 (stamps before any deploy) — FIXED AND LIVE.** Migration `491` removed the two
>   pre-deploy stampers. "Stamps before any deploy" went 2-of-5 → **0**.
> - **Candidate 2 (the stamp ignores the deploy result) — FIXED AND LIVE.** The git-adapter now
>   returns `commit_sha` + `files_sha256` (RFC_038, register `DGH-013`), and `update_page_status`
>   refuses the stamp on a reported skip and records `pages.content_hash` — the sha256 of the bytes
>   actually committed. Migrations `494` (2026-08-20) and `547` (2026-08-21) armed all **six** live
>   deployed-stampers; **zero remain unarmed**.
> - **Candidate 4 (detect divergence) — BUILT, LIVE AND PROVEN.** `page_content_divergence`
>   (register `DGH-015`), shipped in chassis `v1.0.1322`, enabled by migration `526` at 19:23Z.
>   Proven by the discovery run's own record on a site with 21 judgeable pages:
>   `checks_run: [site_unreachable, page_content_divergence]`, `checks_failed: []`,
>   `checks_unregistered: []`.
> - **Candidate 3 (why one page fell out of the batch) — NOT ANSWERABLE FROM HERE, by design.** The
>   runner workflow lives in the private `gqls/sites` repo. Candidate 4 exists to *detect* that
>   failure from this side rather than explain it from theirs. **This is a scope decision, not an
>   omission.**
>
> **⚠ ADDENDUM 2026-08-22 — THE DETECTOR CAUGHT ONE, AND IT SHARPENS CANDIDATE 3. Closed stays closed.**
> `vetcomparison.uk/index.html`, published 2026-08-21 20:49:12Z, was observed serving superseded bytes
> at **1h04, 5h05 and 9h07** after the stamp — the served hash IDENTICAL across all three, so not a page
> caught mid-propagation — and matching again by 11h51. `deployed_at`/`updated_at` never moved, so
> **there was no redeploy: the same publish arrived 9–12 hours late on its own.**
>
> **That is a different fault from the one this file assumed.** §3's live instance looked like a page
> being *skipped* by the batch. This one was not skipped, it was *delayed* — which means the runner is
> not necessarily losing work, and a fix aimed at "why was this page dropped" may be aimed at the wrong
> thing. Still not diagnosable from this side (the workflow is in the private `gqls/sites` repo), but
> whoever has access now has a named case with timestamps instead of a class description. Detail:
> the lane's `NOTES` and register `DGH-015`.

> **THE STATE OF THE FLEET, re-runnable:** `[MEASURED 2026-08-21]` **253 pages carry a fingerprint**
> where all estate history had 0, and a fleet-wide sweep found every active hashed page serving bytes
> that hash exactly to its stored fingerprint. *"Is this page serving what we sent?"* is now one
> string comparison; on 2026-08-18 it took four steps and a judgement call.
>
> **⚠ TWO THINGS IN THIS FILE ARE REFUTED AND KEPT ANYWAY** — §2's table and §5's candidate 4 both use
> `deployed_at` vs origin `last-modified`. Run fleet-wide on 2026-08-19 that method returned **40 of 40
> "stale" on healthy pages, persisting 85 minutes**: a byte-identical rerender legitimately rewrites
> nothing, so the origin's mtime cannot answer this. §2's *conclusion* — that the column tracks "a
> rerender ran", not "bytes were written" — is **correct and is the finding that mattered**. Its
> *method* is not reusable. Only a content hash separates "never needed republishing" from "failed to
> republish", which is this bug's deep finding.
>
> **Residual, tracked elsewhere, none of it reopening this bug:** `PLAN` **D6** (make an unarmed
> stamper NULL the hash rather than leave a stale one — the backstop for a future unarmed stamper) and
> **D8** (widen the settle window to 60 min; the observed delivery tail is ~17 minutes, not the 14
> seconds first measured). Lane docs:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/`

## 1. The one-paragraph version

`webdesign.co.uk/tools/seo-injector/index.html` serves a tool that was replaced hours ago. The
database is correct, four `page_rerender` items have completed with no error, and
`pages.deployed_at` has been stamped fresh each time — while the **origin object has not been
rewritten since 14:12:06Z**. Nothing anywhere in the platform records that the publish did not
happen; `deployed_at` says it did.

## 2. THE MEASUREMENT DEFECT — `deployed_at` is not evidence of publication `[MEASURED 2026-08-18 20:46Z]`

Three pages, each stamped `deployed_at` within the last half hour, against the origin's own
`last-modified` (fetched with a cache-buster; `cf-cache-status: DYNAMIC`, so these are origin headers):

| page | `pages.deployed_at` | origin `last-modified` | serving correct content? |
|---|---|---|---|
| `tool-seo-injector` | 20:45:57 | **14:12:06** | **NO** |
| `tool-json-cleaner` | 20:45:06 | 19:08:55 | yes |
| `tool-smooth-shadow` | 20:15:29 | 19:08:54 | yes |

**All three are stale against their own `deployed_at`** — including the two that are serving correctly.
So the column tracks "a rerender ran", not "bytes were written". Anyone using it to answer *did this
page publish?* gets yes for a page that has not been touched in six hours.

Note the second row's evidence value: the two healthy pages share a `last-modified` **to the second**
(19:08:54 / 19:08:55), which says publication happens in **batches**, decoupled from the per-page
rerender that stamps `deployed_at`. That is the seam where a page can be dropped silently.

## 3. The live instance

- Page `3d1fbd02-ae36-436a-a281-539ac285d4aa`, `/tools/seo-injector/index.html`.
- **DB is correct:** ported slot `15b8323c` `build_status='removed'` (18:57:00); native slot
  `2100c25e` `deployed`, and its stored `rendered_html` **contains the new component's marker
  (`scriptOpenTag`) and does NOT contain the old tool's `b-type`**.
- **Four rerenders, all `complete`, all `error IS NULL`**, orchestrations COMPLETED with no
  `__step_error`: 15:18:58, 17:10:29, 20:12:06, and a purpose-built republish at **20:45:59** filed
  with a distinct `item_key` specifically to rule out dedup silently swallowing it.
- **Origin unchanged throughout:** `last-modified: Tue, 18 Aug 2026 14:12:06 GMT`, content still the
  ported tool (`class="ported-page"` 1, `scriptOpenTag` 0).
- **Isolated, not systemic:** four sibling pages rebuilt the same way today are all serving correctly
  (`html-minifier`, `svg-optimizer`, `json-cleaner`, `smooth-shadow`). The publish seam works; this
  page is being skipped by it.

## 4. Why it stayed invisible until now

Every layer below the artefact reports success: the work item is `complete`, the orchestration is
COMPLETED, `deployed_at` is fresh, and the database holds the right HTML. **This is CLAUDE.md's
"trust the rendered artefact, not the status" with all four lower layers green.** It was caught only
because this lane grades at the served bytes with a cache-buster — and note that an hour earlier the
identical *symptom* on a different page WAS just a stale edge cache, so the cheap explanation was
available and wrong here.

## 5. Fix candidates, ordered by what closes the door

1. **Make `deployed_at` mean what it says**: stamp it only after a confirmed object write, and record
   the written hash/etag alongside. `pages.content_hash` exists and is **empty on all three pages
   above** — populating it at publish time would make "is the origin current?" a comparison rather
   than an assumption.
2. **Fail the rerender when the publish writes nothing.** A completed item that produced no object is
   the defect; it should be `failed` with the reason, not `complete`.
3. **Find why this page is skipped by the batch** — the two healthy pages share a to-the-second
   publish time, so there is a batch boundary; this page is falling outside it. Start from the
   publisher's page selection, not from the rerender.
4. Alert on divergence: a periodic sweep comparing `deployed_at` to the origin's `last-modified` for
   deployed pages would have caught this the first time, at 15:18.

## 6. How to verify a fix

`curl -sI "https://<domain>/<url>?x=$RANDOM" | grep -i last-modified` must move forward after a
rerender completes, on the page named above. Negative control: a page nobody rerendered must NOT move.
**Always cache-bust** — `cf-cache-status: DYNAMIC` confirms you are reading the origin.

## 7. Related

- `docs024_key_docs_latest/webdesign_tool_rebuilds/NOTES_…` 2026-08-18 20:12Z and 20:46Z (full evidence)
- CLAUDE.md "Trust the rendered artefact, not the status" · `MEMORY/prove-a-deploy-at-the-artefact-index`
- The publish-seam canary from another lane (commit `a2a9912c2`, "served sha256 == pre-publish origin
  hash, published_hash written only after acceptance") — that canary proves the seam CAN work; this is
  a page it is not reaching, and the acceptance idea in it is fix candidate 1.

---

## Contribution, 2026-08-19 — a SECOND live instance, and the class is 42 pages across 14 sites

**Not a rival diagnosis and not a fix attempt.** Found by the `agentchassis-22` session while
measuring an unrelated question for the `bugfix_277_required_fields_repair` lane; verified and sized
by that lane. **Neither of us owns this bug and neither is picking it up** — this is here so it is not
lost in a NOTES appendix.

### A second instance of your §3, on a different site and a different site's lane

`vetcomparison.uk` — page `tool-compliance-deadline-calculator`:

| | |
|---|---|
| `pages.status` | **`active`** |
| `pages.build_status` | **`planned`** |
| `page_components` | **0** |
| created / last updated | 2026-07-17 / 2026-07-26 |
| served | **404** (byte-identical to a fabricated-URL control at 2,690 bytes, so the 404 is a real absence, not a fetch artefact) |
| `page_rerender` work items | **3, all `complete`** — 2026-08-11, 08-12, **08-18** |

**Three rerenders completed successfully on a page with nothing to render.** Same shape as your four,
one month older, and it has been `active` and unserved since 2026-07-17.

⚠ It also carries **4 `unbuilt_internal_link` items parked at `needs_human_review`** (2026-08-11).
So the estate *did* detect that links point at an unbuilt page and then stranded the finding — that
half is `bugs_open/083`'s disease, not yours, and is noted only so nobody reads the parked items as
this bug being handled.

### The class, measured — your §3 is not a singleton

[MEASURED 2026-08-19] pages with `status='active'` and **zero** `page_components`:

| `build_status` | pages | sites |
|---|---|---|
| `planned` | **42** | **14** |
| `needs_rebuild` | 11 | 6 |
| **`deployed`** | **2** | 2 |

**The 2 at `deployed` are the sharper version of your bug** — the estate believes those are published
and they have no components at all. The 42 at `planned` are the softer one: never built, but
`status='active'` and therefore link-target-eligible.

### And the detector that should see this files nothing

`diagnose_silent_check_action.go` already carries **two** checks for exactly these shapes:

- `gatherNavLinkedNeverBuilt` — `build_status='planned'` past a grace period, nav-linked, uncovered;
- `deployed_zero_components` — *"page built/deployed but serving zero components"*, and it is
  **`EmitDefault: false`**, described in its own registration as **REPORT-ONLY** because it *"may be
  a deliberate content removal"*.

[MEASURED 2026-08-19] `SELECT ... WHERE item_type ILIKE '%never_built%' OR '%nav_linked%'` returns
**zero rows fleet-wide, all time.** So the detection exists and has produced nothing — **undriven
rather than missing**, which is the distinction that decides whether the fix is code or a schedule.

⚠ **This matters to your §2 specifically:** you argue the measurement defect is the more important
finding. This supports that from a second direction — there are at least two checks that would have
surfaced this class, and one is deliberately silenced by an `EmitDefault: false` whose stated reason
("may be a deliberate content removal") is a judgement nobody has re-examined against a population of
2.

### What we are NOT claiming

The cause of any of it. We have not looked at why the rerenders complete, why the 42 sit at
`planned`, or whether `deployed_zero_components`' report-only default is still the right call. All
first-hand, all re-runnable, none diagnosed.

### ⚠ CORRECTION to the contribution above, same day — my count was 4x low, and the class figure is ~100x bigger

**The error:** I wrote *"3 `page_rerender` items all `complete`"* for
`tool-compliance-deadline-calculator`. That was **`site_work_items` only**, which the
`work-item-archiver` prunes to roughly a 7-day window. Over `site_work_items UNION ALL
site_work_items_archive` it is **13 completed rerenders**, not 3.

Caught by the `bugs_open/302 201` session, who volunteered the trap unprompted: the archive held
**20,184 rows against 10,689 live** when they measured it yesterday, and it had changed two of their
own figures by more than 20×.

⚠ **Which of my numbers were affected, stated so you do not have to guess:** the **42 / 11 / 2 page
counts are NOT affected** — they are computed over `pages` and `page_components`, which are not
archived. **Only the work-item counts were window-limited**, and they were the ones that made the
point.

### The class figure, re-measured over live + archive — and it is the strongest statement of your §2

Pages with `status='active'` and **zero `page_components`**, against every work item ever filed
against them:

| `build_status` | pages | sites | **COMPLETED work items** |
|---|---|---|---|
| `needs_rebuild` | 11 | 6 | **166** |
| `planned` | 42 | 14 | **130** |
| **`deployed`** | **2** | **2** | **35** |
| **total** | **55** | — | **331** |

**331 work items reported success against 55 pages that contain nothing.** The 35 against
`deployed` pages are the sharpest: the estate believes those two are published, they have no
components, and thirty-five items have completed against them.

### And WHY that is evidence about the standard rather than about these pages

From the `bugs_open/302 201` session, who own the adjacent guard and were precise about the boundary
(they declined to adopt this population as a test set for their own work, correctly, because their
guard pins a *predicate* and this is an *artefact* disagreement):

> the sweep completes on **positive orchestration evidence** — "the handler orchestration I
> dispatched reached COMPLETED" — which is explicitly **parity with the lost `mark_complete` write,
> not a stricter test**; migration `220`'s header says so in terms.

**A page with 13 completed rerenders and nothing rendered is exactly what that parity cannot
distinguish.** So this population is evidence about **whether positive-orchestration-evidence is a
sound completion standard at all** — which is your §2's thesis, arrived at from a third direction and
with a number attached.

**Still claiming no cause**, and still not ours: neither this lane, nor `agentchassis-22` who found
the first page, nor the `302/201` session is taking `315`.

---

## Contribution, 2026-08-19 (later) — diagnosis + a fix plan at the council. Three corrections to this file.

Picked up as a **fix** lane (`docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/`,
standing five present). The filing lane says twice it is not theirs to fix and `who-owns.py` confirms
no competing fix thread. **Nothing has shipped**; phases 0–2 are at the council gate as
`Council-Submitted: 377167cd-6324-4bc7-a866-87ad8c435132`.

### §2 is RIGHT, and its evidence table is two findings, not one

Register `DGH-009` (`docs026_concept_register/register/deployment-github.md:101`) records the
mechanism that explains the table:

> *"**`success:true` from the git-adapter is not evidence anything changed.** An unchanged file
> commits as an EMPTY commit and the adapter reports success with the file listed in `deploy_result`."*

A byte-identical rerender ⇒ empty commit ⇒ `b2 sync` rewrites no object ⇒ `last-modified` correctly
stays put. **So `tool-json-cleaner` and `tool-smooth-shadow` are not defects** — their bytes did not
need rewriting, and the column is not lying about them. Only the `seo-injector` row is the bug.

This does not weaken §2, it sharpens it: `deployed_at` honestly means *"a rerender ran and its
output was committed"*, has never meant *"the origin now serves these bytes"*, and the gap only does
damage when the bytes **did** change and still did not arrive.

**But it kills §5 candidate 4 as written.** A `deployed_at`-vs-`last-modified` sweep would have
convicted both healthy pages above. `[MEASURED 2026-08-19]` it is worse than that: 40 sampled
deployed pages ALL have an origin `last-modified` older than their own stamp, and **all 40 share one
three-second window** (09:33:56–58) while their stamps spread over the following hour — that is the
whole-domain `b2 sync` batch, seen directly. The comparison cannot separate *not synced yet* from
*will never sync*; only elapsed time can, and the known bad case took six hours. **Only an
intent-vs-reality content hash separates the three rows**, which makes candidate 1 a prerequisite
for candidate 4 rather than an alternative to it.

### §2's core claim is now measured at the config, not inferred from three pages

`[MEASURED 2026-08-19]`, joining every live `agent_definitions` step on `next_step` (**not** on
`jsonb_each` key order, which is arbitrary and reverses this particular answer): **19 `git_commit`
steps across 16 agents; 6 `update_page_status` steps across 6 agents.** Five stamp `deployed`:

| agent | preceded by | so the stamp is |
|---|---|---|
| `page-build-handler` | `save_page_sections` | **BEFORE any deploy is dispatched** |
| `tool-recreation-handler` | `save_page_sections` | **BEFORE any deploy is dispatched** |
| `page-rerender` | `git_commit` | after a commit whose result it discards |
| `report-builder` | `git_commit` | after a commit whose result it discards |
| `section-editor` | `git_commit` | after a commit it discards, then deploys again |

`deploy_result` appears **nowhere** in `v3_site_actions.go`. There is no arrangement of these five
under which the column could be evidence of publication.

### Correction to the 08-19 contribution's sizing — none of its three named instances is this bug

Checked at the DB and at the served artefact:

| named instance | measured | verdict |
|---|---|---|
| `vetcomparison.uk` `tool-compliance-deadline-calculator` | `build_status='planned'`, **`deployed_at` IS NULL**, still 404 today | real, but **nothing ever stamped it** — an active-never-built page, not a false claim |
| `idea.uk` `/tools.html#audience-check` | `deployed_at` **NULL**; a separate `/tools.html` row with 4 components serves 200 | a **phantom row whose `url` is a FRAGMENT** of another page |
| `ai-agent-orchestration.com` `/roi-estimator.html` | `deployed_at` 2026-05-02, **serves 200, rewritten today 08:37:59** | a stale duplicate row; the URL is live and current |

The "42 / 11 / 2" table sizes *componentless active pages* — a real, overlapping population already
targeted by `check_componentless_pages` — but it **does not size this bug**, and the two rows it
offers as the sharpest cases are the two that are not cases. Flagging it because it is the only
sizing this file carries and a fix aimed at it would be aimed at the wrong population.

### The two columns this needs already exist, and are dead

`[MEASURED 2026-08-19]` `pages.content_hash` **0 of 786**; `page_components.deploy_commit` **0 of
1,775**; and `grep -rn "deploy_commit" --include=*.go .` over the whole repo **including tests**
returns **zero lines**. `pages` and `site_work_items` have no commit/sha column at all.
`CommitToRepo` computes `newCommitSHA` and returns `repo.HTMLURL` — a per-repo constant — so the sha
never leaves the adapter.

`UpdatePageStatusInputSpec.RemovedConfigKeys["commit_from"]` already describes this feature and says
*"Implement it as a feature if wanted, do not re-add the key"*, and
`sql_for_agents/034_page_rerender_agent.sql:99` already promises
`"deploy_result": "git commit result with commit_sha"`. **So §5 candidate 1 is wiring up
designed-and-abandoned machinery, not inventing it** — which is why it is cheaper than it reads.

⚠ I corrected register `DGH-001` in place: it claimed *"Commit SHAs are recorded on pages and work
items for traceability"*, which is false in all three parts. Council seats read register entries as
ground truth, so that line would have drawn an objection to a proposal to add what it says exists.

### Two things I could NOT settle

- **§5 candidate 3 (why one page falls outside the batch) is not diagnosed.** The runner workflow
  lives in `gqls/sites/.github/workflows`; the repo is **private** (`api.github.com` → `Not Found`
  unauthenticated) and the chassis holds no B2 credentials, so neither the ref nor the bucket is
  readable from here. The proposed sweep is deliberately designed to *detect* that failure from this
  side without reading the runner.
- **The diagnosis loop returned no verdict.** Two dispatches
  (`6f900e18-2106-4145-a84c-811baeceaa0d`, `f1433782-6ba7-4304-a7f9-8bd830dfb7c9`) both died at the
  `verdict` step on the Anthropic usage cap. Per the owner ruling of 2026-07-31 I state the
  substitute plainly: every function above read at source, the workflow graph measured live, 744
  `deploy_result` rows censused over 7 days, 40 pages graded at the artefact with cache-busters, and
  the runner pods' own job logs read. If the loop ever completes and refutes any of this, it wins.

### Addendum, 2026-08-19 — I ran §5 candidate 4's own method and it returned 40 false positives

Worth its own note because it is the cheapest way to get this wrong, and I got it wrong first.

`[MEASURED 2026-08-19]` cache-busted `HEAD` on 40 live deployed pages, comparing origin
`last-modified` to `deployed_at` — candidate 4 exactly as written. Result: **40 of 40 "stale"**,
persisting for **85 minutes**, on `webdesign.co.uk`. It looked like a fleet-wide incident.

**All 40 were fine.** The check that settled it: pull `page_components.rendered_html` from the DB,
cut a 120-char needle from it, fetch the served page cache-busted, `grep -F`. The needle is present —
**the origin is serving the current database content.** The component had not changed since
2026-08-15; the origin was last written at 09:33:57 that morning; every rerender since produced
byte-identical output, committed an **empty commit** (register `DGH-009`), and `b2 sync` correctly
rewrote nothing. A deploy job even ran at 10:54:06Z and wrote nothing, as it should have.

**So candidate 4 cannot be built on timestamps, and a settle window does not rescue it** — 85 minutes
had elapsed and the honest answer was still "fine". A sweep built that way files 40 wrong items on
the busiest site in the estate.

**The generalisable point, which is really §2's point sharpened one more turn:** it took four steps
and a judgement call to answer *"is this one page correctly published?"*, because *"the bytes never
changed"* and *"the bytes changed and never arrived"* are **indistinguishable in every signal the
platform exposes** — same item status, same orchestration outcome, same `deployed_at`, same
`success: true` from the adapter, same unmoved `last-modified`. **The defect is not that pages fail
to publish; it is that nothing can tell those two apart.** Which is why candidate 1 is load-bearing
and candidate 4 is worthless without it.

## CONTRIBUTION 2026-08-20 (from the `bugfix_311_component_keys` lane) — the INVERSE case: published WITHOUT the stamp, and the rebuild flag left set

Found while chasing something else, and offered here rather than filed as a new number because it
is this file's defect class from the other side: **`pages`' status columns do not track the artefact
in either direction.** §2 records `deployed_at` stamped when nothing was written. This is the
opposite — the object WAS rewritten and the columns never moved.

**The case** [MEASURED 2026-08-20 15:28Z], and it is your own page:
`webdesign.co.uk` / `tool-ab-test-calculator`.

- `page_rerender` item `ad2a2dc4-4fbb-489f-9b74-bcd00e6f09ff`, handler **`page-rerender`**,
  `created_by='rerender-pages'`, **`status='complete'`** at 2026-08-19 21:36:17Z.
- The page **is** serving the rebuilt tool: 200, 16,172 bytes, **5 `<input>`**, and the markup is
  discriminated to the natively rebuilt component `8a315006` (its `a-visitors`/`b-visitors` ids
  appear in that row's `rendered_html` and in **neither** of the two `removed` slots — `id="verdict"`
  alone would not have told them apart, since the ported original carries it too).
- And yet: **`pages.build_status = 'needs_rebuild'`** and **`pages.deployed_at = 2026-08-14
  22:10:57`** — six days stale, through a successful republish.

**Why it matters to your §2 argument, and it strengthens it:** `deployed_at` is not merely
*optimistic*, it is **uncorrelated** — it can be fresh when nothing was written (§2) and stale when
something was. A reader cannot use it in either direction, and neither can a checker.

**The discriminator that makes it actionable:** the two rebuild paths behave differently.
Items filed at **`page-build-handler`** DO maintain the columns — five such items on `loanzy.uk`
this morning all flipped `build_status` to `deployed` with a fresh `deployed_at`. Items handled by
**`page-rerender`** apparently do not. If that holds beyond this one page, then "which handler
touched it last" predicts whether the row can be trusted, which is a cheaper fix surface than
auditing every writer. **Stated as a lead, not a finding — it rests on one page each way**, and the
census that would settle it is a comparison of `pages.updated_at` / `deployed_at` movement against
completed items grouped by `handler_agent`.

**One knock-on for anyone reading `build_status` as evidence:** a `needs_rebuild` row has **at least
four possible authors** (`refuseDeployStampOnSkip`, `UpdatePageStatusAction`'s shortfall arm,
`check_unresolved_sections`, `flagPagesForRebuild`) and **no attribution column**; a `090` run on
2026-08-20 (`e9555fad-5b25-46bc-9908-f40db98e16a4`) returned **UNVERIFIABLE** partly for that
reason, having found **zero** `agent_error_log` rows for three shortfall pages. So "the page is
flagged for rebuild" tells you neither who flagged it nor whether it has since been rebuilt.
