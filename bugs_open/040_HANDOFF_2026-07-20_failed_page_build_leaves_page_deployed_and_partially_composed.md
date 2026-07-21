# Handoff — a FAILED page build leaves the page `deployed`, partially composed, and invisible to the reconciler

> **NUMBER COLLISION (2026-07-20, same day):** another thread filed a different
> `040` (`kafka_dial_timeouts_fleetwide_intermittent`, which cites itself as
> **040-kafka-dial**). Numbers are never reassigned — resolve by slug
> (`bugs_closed/README.md` duplicate-numbers table). Cite this case as
> **040-partial-build**.
>
> The two are related in substance, not just in number: 040-kafka-dial is a
> plausible *cause* of the spawn timeout that triggered this case (see "What
> happened" below), and `/bugs_open/003` is the mechanism connecting them.

**Filed 2026-07-20.** Found by asking the framework to rebuild a page through its own route
(rather than hand-restoring it) while verifying `/bugs_open/001` — which is exactly what exposed it.

**The work item says `failed`. The page says `deployed`. Nothing reconciles the two**, and because
the page also gets stamped with the current plan version, the reconciler will now *skip* it forever.

---

## FIX APPLIED 2026-07-21 — candidate 1 (structural), ALREADY LIVE on v1.0.1146

**Status: FIX LIVE, behavioural firing not yet observed. 040 STAYS OPEN** (the section-DROP root
cause is separate and unfixed — see below). This is the structural fix: a page must not be stamped
`deployed` + `built_from_plan_version` unless every planned section was written.

> **It went live faster than expected — via another session's build, not mine.** I edited
> `v3_site_actions.go`, and before I committed it a concurrent add-all sweep committed it into
> `fe2ba5e52` ("v1.0.1146 - sweep"), which then built and deployed the chassis image. So the Go
> source shipped in **v1.0.1146** without my ever running a build — the exact
> "committed code rides ANYONE's next sweep build, and builds are from HEAD" lesson. **Verified live
> against the running pod binary** (not the tag): `strings /app/agent-chassis | grep -c "build is
> short of its plan"` → 1, `"planned sections rendered"` → 1, positive control `"no rendered
> components"` → 2. This commit carries only the **test + docs**; the Go source is in `fe2ba5e52`.
>
> **What is proven vs not.** LIVE in the binary ✓; decision logic unit-tested ✓
> (`page_section_shortfall_test.go`: refuse 6/5, allow 4/4, allow 5>3, allow 0/0); SQL live-validated
> ✓ (dartsonline → `testimonials` missing; 28 pages fleet-wide; `gaswholesalers.com/services` 4=4 NOT
> flagged). **NOT yet proven: an actual live partial build being refused** — the guard is passive
> until a build with a shortfall reaches the deploy mark, and none has since the 12:15:20Z roll
> (`kubectl logs <chassis pod> | grep "short of its plan"` → 0). This is deployment, not correctness
> (the "verify the failing branch" rule). To catch the first natural firing:
> `kubectl -n ai-persona-system logs -l app=agent-chassis --since=1h | grep -i "short of its plan"`.
> A forced proof (rebuild a known-partial page and assert it flips to `needs_rebuild` with the stamp
> cleared) is deliberately NOT done here — it dispatches into the contended queue (`bugs_open/030`,
> ~30-min latency) and, for a deterministic-drop page, starts the rebuild churn described below.

**Where.** `platform/orchestration/actions/v3_site_actions.go`, `UpdatePageStatusAction` (the
`update_page_status` action — the `update_status` step that every deploy path runs *before* the
git deploy). The existing "Option B" guard already refused a **0-component** deploy; this extends
the same choke point to refuse a **partial** deploy. New helper `pageSectionShortfall` returns
`(planned, rendered)`; the guard flips the page to `needs_rebuild` and clears the stamp when
`rendered < planned`, exactly like the 0-component path. Test:
`platform/orchestration/actions/page_section_shortfall_test.go`.

**The load-bearing decision: count ROWS, not section names.** My first instinct was to match each
planned section name against `page_components.slot_name` / `content_components.function` (the
`/bugs_open/039` Part-1 pattern, and candidate 3's suggestion). **I validated it fleet-wide before
shipping and it was WRONG** — it flagged **74** deployed pages where the count method flags **28**.
The extra ~46 are false positives: `pages.sections` names do **not** reliably equal the live
`slot_name`/`function`. Concrete, from the live DB 2026-07-21:

- `gaswholesalers.com/services` — planned `["services-hero","services-grid","features","call_to_action"]`
  vs live slots `["hero-services","services-grid","features","call-to-action"]`. Two names diverge —
  **word order** (`services-hero`↔`hero-services`) and **underscore vs hyphen**
  (`call_to_action`↔`call-to-action`) — on a page that is complete and serving 4 real components
  (1275/3473/4925/2151 bytes). Per-name matching would refuse it and drive it into a rebuild loop.

So the guard compares **row count** (`count(page_components) < count(sections − suppressed_sections)`),
which is the signal the fleet sweep below already validated. Trade-off, stated plainly: the count
method has a **false negative** the name method would not — a page with a duplicate component *and* a
missing section counts equal and slips through. That is acceptable for a deploy gate, where a false
*positive* (refusing a healthy page) is far more damaging than a false negative (the fleet sweep and
`incomplete_page_group`/`empty_sections` discovery checks are the other layers). See
`016b_debugging_guide` §9 for the transferable pattern.

**Suppressed sections are excluded** from the planned count, so a deliberately-dropped section
(`pages.suppressed_sections`) is never read as a shortfall. `sections=[]` pages (tools/blog-index
rendered by another subsystem — `/bugs_open/050`) have `planned=0` and are never refused.

**Positioning verified.** In all four deploy callers (`page-build-handler`, `page-rerender`,
`section-editor`, `tool-recreation-handler`) a component-writing step (`save_sections`/`apply_edit`)
runs *before* `update_status`, so a shortfall at the guard is real, not a mid-build race. This is the
same precondition the live 0-component guard already relies on.

**Fleet consequence now live.** ~28 already-partial deployed pages (the sweep below) will, on their
next build/edit/rerender (the fix is live as of the 12:15:20Z roll), be refused and flipped to
`needs_rebuild`, so the reconciler re-emits them. That is the intended healing. **Caveat (honest):**
if the underlying section-drop is
*deterministic* for a page (dartsonline dropped `testimonials` on two separate builds — root cause
still UNKNOWN, see the CORRECTED block below), that page will re-enter the build queue each reconcile
cycle rather than converge. That is a **slow drip gated by reconcile cadence + work-item dedup**, and
it is strictly better than a silent permanent five-sixths page — the loop is the visible signal that
the separate section-drop defect needs its own diagnosis. **Candidate 2** (propagate the orchestration
error onto the work item's `error`) is NOT done here — it is now smaller than filed (the error is
already durable in `agent_error_log` since v1.0.1140; it is a join, not a missing record) and is a
cheap independent follow-up.

**How to verify** (against a real build): see "How to verify a fix" below, plus re-run the fleet
sweep and confirm no *healthy* page (e.g. `gaswholesalers.com/services`, 4=4) was flipped to
`needs_rebuild`.

**Residual / next investigation (NOT this fix).** The fix stops a partial build being *recorded* as
a complete, plan-stamped success; it does **not** stop the partial build being *produced*. Why
`testimonials` is dropped by a build that reports complete is still **UNKNOWN** (see the CORRECTED
2026-07-20 block below — three hypotheses tested and refuted). Until that is found, the fix converts a
silent permanent five-sixths page into a visible one that keeps asking to rebuild — which for a
*deterministic*-drop page is a slow reconcile-paced loop. Finding the section-drop cause is now the
highest-value follow-on, and this fix is what makes it visible instead of silent.

---

## What happened (dartsonline.com `index`, 2026-07-20)

A `needs_page` item for `index` was handed to the framework. The build wrote five components, then
its spawned child request timed out and the orchestration ended in `complete_error`:

```sql
SELECT status, current_step, collected_data->'__step_error'->>'message'
FROM orchestration_states WHERE orchestration_id='5c930b26-cf9a-4779-a19d-1215c8ad1de6';
-- COMPLETED | complete_error | Request 57508ea9-…-74821de917b8 timed out after 3 retries
```

The item was correctly marked failed:

```
status        | failed
attempt_count | 1
max_attempts  | 3
error         | (empty)
result        | {"completed_by_step": "mark_item_failed",
                "completed_by_orchestration_id": "5c930b26-…"}
```

**But the page was left deployed anyway:**

```sql
SELECT build_status, sections, built_from_plan_version FROM pages WHERE name='index' AND site_id='5fe8785b-…';
-- deployed | ["hero","product-grid","category-listing","features","call-to-action","testimonials"]
--          | 5d438145-0d2d-4023-9761-98ab4d06318c
```

and only **five of the six** planned sections exist, each marked `build_status='deployed'`:

| position | function | bytes |
|---|---|---|
| 1 | hero | 2466 |
| 2 | product-grid | 3055 |
| 3 | category-listing | **364 — hollow** |
| 4 | features | 4079 |
| 5 | call-to-action | 2347 |
| — | **testimonials — ABSENT** | — |

`testimonials` is not suppressed (`pages.suppressed_sections = []`) and the component exists and is
active (`content_components`: `name='testimonials', function='testimonials', is_active=t`). It was
simply never written — almost certainly the request that timed out.

> ### CORRECTED 2026-07-20 18:00 UTC, after the v1.0.1140 deploy — two of my own claims were wrong
>
> **(a) "almost certainly the request that timed out" is REFUTED.** The page was rebuilt again at
> **15:18:17** by an `image-build-handler` item (*"Re-render index after its image asset landed"*,
> reason `image_landed`) which reported **`status='complete'`** at 15:19:02. That build **also
> produced 5 of 6 sections** — `testimonials` missing again, `category-listing` again a 364-byte
> shell. **A successful build drops the section too**, so the timeout never explained it.
>
> Also, the timeout's failing step was **`deploy_page`**, i.e. *after* component writing. It could
> not have eaten a component that the earlier steps were responsible for producing. I had asserted
> a cause that the step name alone contradicts.
>
> Two other hypotheses tested and also refuted: the `testimonials` template passes
> `sectionTemplateValid` (it contains `</section>`), so it is not being dropped by
> `loadComponentSchemas`' truncation skip; and it is not in `sections_deferred` or
> `sections_skipped`. **The cause of the dropped section is currently UNKNOWN** — that is the honest
> state, and it wants its own investigation rather than another guess.
>
> **(b) "the cause is one join away but not written where a human or a sweep would look" is TOO
> STRONG.** The failure *was* durably recorded, in `agent_error_log`, with the message, the step and
> the work item id:
> ```sql
> SELECT agent_type, step_name, error_message, work_item_id FROM agent_error_log
> WHERE orchestration_id='5c930b26-cf9a-4779-a19d-1215c8ad1de6';
> -- page-build-handler | deploy_page | Request 57508ea9-… timed out after 3 retries | f98331dd-…
> ```
> The narrower true statement: **`site_work_items.error` is empty**, so the work item does not carry
> its own failure, but the platform did record it in a queryable table (32,398 rows since
> 2026-04-02). Fix candidate 2 below is therefore smaller than written — it is a join, not a
> missing record. v1.0.1140 adds `platform/orchestration/agent_error_log.go`, a further writer for
> that table from `routeToErrorStep` / `notifyParentOfFailure`.
>
> ### And the core defect is now DEMONSTRATED rather than predicted
>
> After the 15:18 rebuild the page is `deployed` **and stamped with the CURRENT plan**
> (`built_from_plan_version = dcc7834e`, the second re-plan's id). So `decideEmit` now genuinely
> returns `skip_built` for it. **dartsonline's homepage is a permanent five-sixths page that no
> longer asks to be built**, and this time it got there via a build that reported success. That is
> the whole bug, observed end to end, without needing the failure at all — which makes the title's
> emphasis on "FAILED" too narrow. **A partial build reports success and strands the page either
> way.**

`category-listing`'s 364 bytes are an empty shell (empty `<h2>`, `href="#"`, empty grid) — the
`/bugs_open/039` empty-section shape, and there is a matching `empty_section` work item for this
exact page and section sitting `detected` since 2026-07-14.

## Why it is worse than a lost section

**The page is now invisible to the reconciler.** `decideEmit`
(`reconcile_site_plan_action.go:341`) returns `skip_built` when the page is `deployed` **and**
`built_from_plan_version == planID`. Both are now true. So a reconcile against plan `5d438145` will
skip `index` as finished — a five-sixths page that no longer asks to be built. The failure marked
the *item*, and the item is terminal; nothing will revisit the page.

Note the interaction with `/bugs_open/038`: a *new* plan would flip it back to `stale` and rebuild
it. So the page happens to be recoverable by re-planning the whole site — but only as a side effect
of a different defect, and not by any mechanism that knows this page is incomplete.

**Two secondary observations**, not the main defect but worth checking with it:
- `attempt_count = 1`, `max_attempts = 3`, status terminal `failed` — **no retry happened**. Whether
  the retry budget is honoured on this path is unverified; if it is not, that is a second defect.
- `error` on the work item is **empty**, though the orchestration recorded a clear message. The
  cause is one join away but not written where a human or a sweep would look. Same shape as
  `/bugs_open/034` (validation errors dropped with no durable record).

## Fix candidates

1. **A page must not be `deployed` unless every planned section was written.** At the end of the
   build, compare the written `page_components` against the plan's section list; on a shortfall,
   leave `build_status` as-is (or set `needs_rebuild`) and do not stamp `built_from_plan_version`.
   The stamp is the part that makes the damage permanent, so at minimum: **never stamp on a failed
   or partial build.**
2. **Propagate the orchestration error onto the work item.** `mark_item_failed` has the failing
   orchestration id in `result`; write its `__step_error.message` into `error` so the failure is
   legible without a manual join.
3. **A completeness check that compares plan to artefact.** `pages.sections` vs the page's
   `page_components` (matching on `function` — see `/bugs_open/039` Part 1) would catch this class
   fleet-wide regardless of which build path dropped the section. Worth checking whether
   `complete_work_item_verification.go` is the right home; it exists for saga no-ops of this shape.

Candidate 1 is the structural fix. 2 is cheap and independently worth doing.

## How to verify a fix

1. Force a page build to fail partway (kill the handler pod mid-build, or point a section at a
   missing component). Assert the page does **not** end `deployed`, and `built_from_plan_version`
   is **not** stamped.
2. Assert the reconciler then re-emits that page rather than skipping it.
3. Assert the work item's `error` is non-empty and names the real failure.
4. Re-run the fleet sweep below and assert it shrinks.

## Fleet sweep — this is NOT an isolated page (run 2026-07-20)

```sql
SELECT s.domain, p.name,
       jsonb_array_length(p.sections) AS planned,
       count(pc.id) AS rendered
FROM pages p JOIN sites s ON s.id=p.site_id
LEFT JOIN page_components pc ON pc.page_id=p.id
WHERE p.build_status='deployed' AND jsonb_array_length(COALESCE(p.sections,'[]'))>0
GROUP BY 1,2,3 HAVING count(pc.id) < jsonb_array_length(p.sections);
```

**25 deployed pages across 6 sites are short of their plan; 39 sections missing in total; 4 pages
have ZERO components.** Worst offenders: `gaswholesalers.com/wholesale-pricing-explained` (7 planned,
0 rendered), `leopardessconsulting.co.uk/case-study-multi-agent-cost-control-platform` (4, 0), two
`finetuning.uk` blog posts (3, 0 each). `leopardessconsulting.co.uk` accounts for 11 of the 25.

> **Read this before chasing the list.** The query measures the **DB record**, not the live page,
> and the two can diverge — a page deployed months ago still serves its old file even if its
> `page_components` rows were later removed. So a row here is a **signal to triage, not proven live
> damage**. I checked one before believing it, and recommend the same for each:
>
> `gaswholesalers.com/wholesale-pricing-explained.html` — `deployed_at` 2026-05-06, 0 component
> rows. It serves **HTTP 200, 11,872 bytes**, which at first looks like a healthy page and a false
> positive. It is not: the body has **zero `<section>` elements, zero `<h1>`, zero `<h2>`, and its
> `<main>` contains 0 characters of text.** The 11.8KB is entirely header/footer/nav chrome. **A
> live, blank page marked `deployed`.**
>
> So for this one the DB signal and the live artefact agree.

### All 25 checked live (2026-07-20) — the split is 4 broken, 21 degraded

Every row above was fetched and its `<main>` text measured. **The DB signal is a poor predictor of
severity: being short by 2 of 7 sections can mean a totally blank page, or a perfectly serviceable
one.** Do not triage from the DB counts alone.

**Group A — 4 pages, genuinely blank and live** (`<main>` = 0 characters; all are the zero-component
rows). These are unambiguous damage:

| page | sections planned |
|---|---|
| `finetuning.uk/blog/chatgpt-has-your-data-does-that-matter.html` | 3 |
| `finetuning.uk/blog/what-is-rag-and-do-small-businesses-need-it.html` | 3 |
| `gaswholesalers.com/wholesale-pricing-explained.html` | 7 |
| `leopardessconsulting.co.uk/case-study-multi-agent-cost-control-platform.html` | 4 |

Two are blog posts with a title and no article. One is a pricing explainer with nothing in it.

> **CORRECTED 2026-07-20 by the site owner.** I wrote here that the leopardess entry was "a case
> study on a site that has no clients", so blank was the safer failure and it should be deleted.
> **The premise was wrong.** The owner's case studies describe **real work — systems actually built
> and running — they are simply not *client* case studies.** The distinction matters: "no clients"
> was read as "nothing to write about", which is false. That mistaken premise also appears in
> `/bugs_open/001`'s FRESH EVIDENCE section ("this site has no clients and no case studies") and
> should not be carried forward from there either.
>
> Checking the live site after the correction changed the conclusion completely — see below.

**The leopardess case-studies section is NOT broken and must not be deleted.** `/case-studies.html`
is `deployed`, in the header and footer nav, and serves 8,054 characters of honest, specific copy:
h1 *"Systems running in production, not proposals waiting for a budget."*, then four real systems —
Companies House record checking, news trust-ranking, no-code interactive tool generation, and "the
platform that built this website". That is exactly the owner's framing (real work, not client work)
and it reads well.

**What is actually wrong is a set of orphaned detail pages.** Eight `case-study-*` detail pages exist
in `pages`, and:

- **nothing links to any of them** — checked `/`, `/case-studies.html`, `/who-we-help.html`,
  `/use-cases.html`, `/how-it-works.html`, `/insights.html`, `/our-approach.html`: zero `href`s to
  `/case-study-*`. They are also `in_header=false, in_footer=false` and **absent from
  `sitemap.xml`** (27 URLs, none of them).
- **seven of them serve HTTP 200 with a completely empty `<main>`**; the eighth
  (`document-intelligence-pipeline`) 404s.

So the index tells the story inline and never needed detail pages, while seven blank shells sit at
live URLs reachable only by typing them directly. They are a tidy-up, not an outage — and deleting
them is the likely right call precisely *because* the section above them is already complete.
Rebuilding them instead would re-run the content writer over case-study material, which is the
`/bugs_open/029` fabrication path on the one site where that history is worst.

**Do not action the deletion without the owner's word** — this is live outward-facing content.

**Group B — 21 pages, short by one or two sections but serving real content** (`<main>` from 1.6k to
17k characters). e.g. `leopardessconsulting.co.uk/our-approach.html` renders 5 of 6 sections and
11,650 characters. These are degraded, not broken; each needs a per-page judgement about whether the
missing section mattered (a CTA and a FAQ are not the same loss). Eleven of the 21 are leopardess.

Caveat on the method: counting `<section>` tags is unreliable across templates — the four
`gamesdesign.co.uk/games/*` pages report zero `<section>` elements while serving 12k–17k characters
of real content in different markup. `<main>` text length is the trustworthy signal; the section
count is not.

## Related

- `/bugs_open/028` (slug `page_build_noop_reports_complete_and_deploys_borrowed_components`) — the
  closest relative: a no-op reporting `complete` and deploying borrowed components. **Distinct**:
  there the item wrongly said `complete`; here the item correctly says `failed` and the *page* state
  disagrees with it. Same family (page state not derived from build reality), different mechanism.
- `/bugs_open/003` — spawned children losing their response is the timeout that triggered this. That
  bug explains the *failure*; this one is about what the failure leaves behind.
- `/bugs_open/038` — why re-planning the site would mask this by rebuilding the page anyway.
- `/bugs_open/039` — the hollow `category-listing`, and the function/name matching needed for fix 3.
- `/bugs_open/034` — the same "error existed but was never written durably" shape.
