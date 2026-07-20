# Handoff — a FAILED page build leaves the page `deployed`, partially composed, and invisible to the reconciler

**Filed 2026-07-20.** Found by asking the framework to rebuild a page through its own route
(rather than hand-restoring it) while verifying `/bugs_open/001` — which is exactly what exposed it.

**The work item says `failed`. The page says `deployed`. Nothing reconciles the two**, and because
the page also gets stamped with the current plan version, the reconciler will now *skip* it forever.

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

Two are blog posts with a title and no article. One is a pricing explainer with nothing in it. The
leopardess one is a **case study on a site that has no clients** — so blank is arguably the safer
failure here, and it should be deleted rather than rebuilt unless there is a real case study to
tell. Check `/bugs_open/029`'s fabrication concern before regenerating any of them.

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
