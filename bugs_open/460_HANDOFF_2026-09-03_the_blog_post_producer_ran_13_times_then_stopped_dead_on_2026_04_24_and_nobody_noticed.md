# 460 — the blog-post producer (`blog-content-planner`) ran 13 times, then stopped dead on 2026-04-24, and nothing has noticed in four months

**Filed 2026-09-03 ~11:50Z by the `gamedesign.uk` lane. UNOWNED — filed as an evidenced OPEN
QUESTION, not a diagnosis.**

> **THIS FILE ASSERTS NO ROOT CAUSE.** Every number below is first-hand and dated; the *why* is
> exactly what is missing, and I have not investigated it. Per CLAUDE.md's 2026-07-31 ruling, the
> "file through `090` before asserting a cross-cutting cause" trigger does not fire on a filing
> that deliberately asserts none — but it **will** fire on whoever writes the cause into this file.
> Read §6 before you do.

## 1. What the mechanism is, in plain terms

`blog-content-planner` is a live agent whose job is to plan and create a site's articles. It runs
the `create_blog_posts` action, which writes `blog-post` page rows directly into `pages`. It is the
second of the estate's two routes to an article — the other being `build-site-planner` planning
`blog-post` pages inline as part of the site plan.

Both routes are real and both have produced live articles. This file is about the second one
having gone silent.

## 2. The measurement (all `[MEASURED 2026-09-03 ~11:10–11:50Z]`, first-hand)

**The agent is live and wired:**
- `platform/orchestration/actions/create_blog_posts_action.go` — `INSERT INTO pages (… page_type …)`
  with `page_type` from `datahelpers.CanonicalisePage` on a `blog-post` role.
- Registered at `platform/orchestration/actions/registry.go:720` and via
  `RegisterActionInputSpec("create_blog_posts", …)`.
- Exactly **one** live agent definition names it: `blog-content-planner` — `is_active = true`,
  `is_snapshot = false`, `deleted_at IS NULL`. Its steps: `load_specs`, `check_existing_posts`,
  `plan_posts`, `create_post_pages`, `ensure_site_record`, `complete`.
- A discovery check exists to drive it: `platform/orchestration/actions/discovery_checks/check_empty_blog.go`
  files `item_type='needs_blog_posts'`, `HandlerAgent: "blog-content-planner"`, summary *"Blog page
  exists but no blog posts — needs initial content planned by blog-content-planner"*.

**It ran, and it worked:**

| | |
|---|---|
| `needs_blog_posts` items, all history (live + archive) | **14** |
| — `complete` | **13** |
| — `wont_fix` | 1 |
| first item | **2026-03-14** |
| **last item** | **2026-04-24** |
| sites served | `finetuning.uk`, `ai-agent-orchestration.com`, `leopardessconsulting.co.uk`, + 3 rows whose `site_id` no longer has a `sites` row |
| `llm_call_log` for `agent_type='blog-content-planner'`, all history | **10 calls, 2026-04-03 10:05:59Z → 2026-04-24 14:42:00Z** |

**Nothing since 2026-04-24 — on either instrument, independently.** The last work item and the last
LLM call agree to the day. That is 4 months and 10 days as of filing.

## 3. What is NOT claimed — the census that stops this being tidier than the evidence

I checked whether the sites this producer served are the sites with the biggest article archives.
**Partly, and the exceptions matter:**

| site | active `blog-post` pages | first created | served by this producer? |
|---|---|---|---|
| webdesign.co.uk | 52 | 2026-08-05 | **no** — no `needs_blog_posts` row ever |
| dartsonline.com | 23 | 2026-07-06 | **no** — same |
| finetuning.uk | 22 | **2026-03-14** | yes (items 03-14 → 04-24) |
| ai-agent-orchestration.com | 18 | **2026-03-14** | yes (items 03-14 → 04-24) |
| leopardessconsulting.co.uk | 14 | 2026-04-04 | yes (item 04-23) |

So **the planner route is demonstrably alive and producing** (webdesign.co.uk's 52 posts are all
August; dartsonline's run to September), and this producer is **not** the only source of articles.
The two sites whose first article lands exactly on 2026-03-14 — the day of the first
`needs_blog_posts` items — are suggestive, not proof: I have not joined individual pages to
individual runs. **Do not write "this producer made those archives" without that join.**

## 4. Why it is worth someone's hour

1. **A driven-then-stopped mechanism is a different investigation from a never-driven one.** It
   worked 13 times. Something changed on or about 2026-04-24 and no alarm exists for it.
2. **It is the missing producer that two other open bugs keep naming.** `bugs_open/444`'s gate
   files `capability_gap` rows with `builder_needed=blog_posts` for listing pages with nothing to
   list — gamedesign.uk got one at 2026-09-03 10:40:18Z. This agent is what would satisfy them.
3. **`build-site-planner` cites it as a reason to plan no articles** (`bugs_open/428`). On three
   sites in the trailing 30 days the planner deferred article pages to an editorial pass —
   gamedesign.uk verbatim: *"satisfied by the blog infrastructure; individual posts are not planned
   as static pages here."* **The planner was pointing at something real that had stopped running.**
   Migrations **730/731** (2026-09-03, designblog.co.uk lane) now tell the planner in its own prompt
   that no later pass runs and name this agent's dormancy with its date. **If this producer is
   revived, that rule text becomes stale and must be revisited** — it is dated precisely so that is
   checkable.
4. **Reviving it would be an alternative or complementary route to rule 20's launch posts**, which
   currently make the planner solely responsible for a site's first articles.

## 5. ⚠ TWO ROLLING-WINDOW TRAPS — both would have made this bug invisible

Both fired on real sessions today. Neither produces an error.

1. **`site_work_items` is not all of history.** `[MEASURED 11:45Z]` the live table holds 29,680 rows
   (2026-03-15 → 2026-09-03) and **`site_work_items_archive` holds 33,350** (2026-02-22 →
   2026-08-26). A closed row is archived out of the live table. Querying `site_work_items` alone for
   `item_type='needs_blog_posts'` returns **0 rows, fleet-wide, all statuses** — which reads as
   *"this was never used"* and is the exact opposite of the truth. **All 14 rows are in the
   archive.** I nearly filed the false version; the live-table zero is what made me check.
2. **`orchestration_states` is a ~24-HOUR window.** `[MEASURED 11:10Z]` 9,163 rows spanning
   2026-09-02 10:41Z → 2026-09-03 10:59Z. A `SELECT … WHERE owner_agent_type='blog-content-planner'`
   there returns 0 and proves **nothing whatever** about April. The `bugs_open/427` lane hit this
   one first and caught themselves.

**The instrument with real memory is `llm_call_log`** — kept verbatim as the training corpus (owner
ruling 2026-08-22), so it reaches back to April. Use it, and the archive table, for any "has X ever
run" question.

## 6. Where to start (nothing here is a hypothesis, only a place to look)

- **What dispatched it?** ZERO rows in live+archive carry `handler_agent='blog-content-planner'`
  *outside* those 14 `needs_blog_posts` items, so `check_empty_blog` is the only known driver. Is
  that check still RUN? Is it in a discovery suite that still executes, and does it still reach a
  dispatcher rather than filing in `record` mode (RFC_056's circuit breaker, 2026-08-25, holds
  LLM-audit findings back from auto-dispatch — **note it postdates the stop by four months, so it
  cannot be the cause**)?
- **Did the agent or its trigger change?** `git log` the action, the check, and any seed touching
  `blog-content-planner`; `agent_definitions` snapshots around 2026-04-24.
- **Did the sites simply stop qualifying?** `check_empty_blog` fires on "blog page exists but no
  blog posts". If every site that had a blog page got its posts in March/April, the check would go
  quiet **legitimately** — and this whole file would be a non-bug. **That is the cheapest thing to
  test first, and the most likely benign explanation.** Test it by finding sites with a blog/section
  index and zero posts TODAY (gamedesign.uk is one) and asking why no item was filed for them.
- **`bugs_open/443`** covers a different defect in the same file (`create_blog_posts` writes
  `pages.sections`, the materialised cache, never `site_plan_sections`, the authority) — read it
  before changing anything there, and do not confuse the two.

## 7. Ownership / routing

**Unowned.** Raised by the `gamedesign.uk` lane as a by-product of `bugs_open/428`'s planner-refusal
work; that lane owns a site, not this class, and is not taking it. The `bugs_open/427` lane found
the producer and holds the 428 residual; the `designblog.co.uk` lane owns migrations 730/731 whose
rule text names this agent's dormancy and would need revisiting if it is revived. Both have been
told. `scripts/who-owns.py 460` before routing work at it.
