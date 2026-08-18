# 301 — `page-build-handler` runs the LLM writer AND the link resolver BEFORE the owned-page guard, so the expensive work is done and then thrown away

**Filed 2026-08-18** by the `vigilant_designer_offer_analysis` lane. Found only because
`bugs_closed/295` made the refusals visible — before that fix this was invisible in every record.
**OPEN, unowned.**

**One line:** the ownership guard sits at `save_sections`, the LAST step of the workflow, so a
build targeting a `rebuild_policy='owned'` page runs `page-content-writer` (an LLM call) and
`internal-link-resolver` to completion first, and only then refuses. Measured overnight on one
site: **39 full chains run and discarded in ~2.5 hours.**

---

## The evidence

`[MEASURED 2026-08-18 12:20 UTC]` on webdesign.co.uk, window 02:30–05:00:

| orchestration owner | count | terminal step |
|---|---|---|
| `page-content-writer` | **39** | `complete` (COMPLETED) |
| `internal-link-resolver` | **39** | `complete` (COMPLETED) |
| **`page-build-handler`** | **39** | **`complete_error`** |

So the writer produced content 39 times and the resolver resolved links 39 times, and 39 builds
then ended in error. Over the same window the guard filed **38** new `owned_page_review` rows
(`refused_by='save_page_sections'`), and the work items show **21 `needs_page` + 17
`content_rewrite` failed** on that site.

**The ordering is structural, not incidental.** `page-build-handler`'s live workflow (read from
`agent_definitions`, 2026-08-17) is:
```
ensure_site_record → load_page_record → check_page_found → load_existing_content →
load_spec_sections → plan_sections → check_has_ready_sections → spawn_content_writer →
call_content_writer → check_content_produced → validate_content → save_sections → …
```
`load_page_record` already reads the page. **`rebuild_policy` is knowable at step 2 and is not
consulted until step 12.** The predicate is one function call (`pageIsOwnedForGuard`) against a
column on a row the workflow has already loaded.

**Scale beyond one night.** `owned_page_review` rows from this path, since the fix went live
2026-08-17 18:57: **59 rows across 5 sites in ~14 hours**, of which **49 are webdesign.co.uk —
half of that site's 97 owned pages.** Every one of those is a refusal, and on this route a refusal
means a writer run already happened.

⚠ **What is NOT established.** I have not proven that all 39 writer runs were for owned pages —
the counts are equal and the window matches, which is strong but is a correlation across three
aggregates, not a per-orchestration join. `[UNMEASURED]` the per-run linkage; the parent/child ids
are in `orchestration_states` and the join is straightforward for whoever picks this up, but
retention is ~24h so **it must be done on a fresh burst, not on this one.** Nor have I costed the
LLM spend — "39 writer runs" is a count, not a token figure.

## Why it matters beyond the waste

1. **Cost.** Each discarded chain is at least one content-writing LLM call, on a fleet that hit its
   Anthropic cap on 2026-08-14.
2. **It manufactures the queue noise it then reports.** Every refusal files an
   `owned_page_review` at `needs_human_review`. 59 rows in 14 hours is the guard working as
   designed — but a large share of them are for builds that should never have been attempted, so
   the human queue fills with reports of work nobody asked for.
3. **It is the same shape as `bugs_open/208`, one route over.** There the guard sat behind a git
   commit; here it sits behind an LLM call. 208's lesson was "move the refusal earlier"; this is
   the route that did not get moved.

## Fix candidates, ordered by what closes the door

1. **(Preferred) Refuse at `load_page_record`/`check_page_found`, not at `save_sections`.** The row
   is already loaded and `rebuild_policy` is on it. An owned page should take the error arm
   immediately, file the same `owned_page_review` row (`emitOwnedPageReviewItem` is already wired
   into the save path by `bugs_closed/295` and would move with it), and never reach the writer.
   ⚠ **Keep the save-path guard as well** — it is the backstop for any other caller, and removing
   it would re-open 295.
2. **(Cheaper, config-only, partial) Add an ownership condition to `check_has_ready_sections`.**
   No image roll. But it only helps the branch that goes through that check, and a config predicate
   duplicating a Go one is the drift class this estate keeps filing bugs about.
3. **(Upstream, the real repair) Stop filing generic content items against owned pages at all** —
   triage-time routing to `section_edit`, which completes on owned pages (18 times measured
   2026-08-17). This is `bugs_closed/295`'s untaken fix candidate 3 and it addresses the cause
   rather than the cost.

## How to verify a fix

Positive AND negative control, on a fresh burst:
- Dispatch a content item at a known **owned** page → expect **no** `page-content-writer`
  orchestration for it, an `owned_page_review` row still filed, and the item still `failed`.
- Dispatch one at a known **generic** page → expect the writer to run normally and the page to save.
Without the second, "no writer ran" is equally consistent with having broken the writer.

## Relates to

`bugs_closed/295` (made this visible; its fix candidate 3 is this file's candidate 3) ·
`bugs_open/208` (the sibling ordering defect on the rebuild route — guard behind a git commit) ·
`bugs_open/115` (findings that terminate nowhere — the queue-noise half) ·
LANDMINES `count(DISTINCT item_key)` entry (how to count these rows correctly)
