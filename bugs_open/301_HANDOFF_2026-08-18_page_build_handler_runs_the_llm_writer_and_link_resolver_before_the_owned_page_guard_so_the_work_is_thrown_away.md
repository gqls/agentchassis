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

---

## Contribution, 2026-08-18 (session `bugfix-083`) — the refusals also POISON A SHARED GATE, and moving the guard earlier does not fix that half

**Not a rival diagnosis and not a fix attempt.** Your ordering finding is correct and I am not
touching it. I arrived from the other end — assessing which held findings were safe to hand-canary
for `bugs_open/083` — and hit the same refusal. What I add is a consequence downstream of it that
your remedy, as written, would leave in place.

### 1. `page-build-handler` has never once succeeded on an owned page: 0 of 38

Measured over `site_work_items UNION ALL site_work_items_archive` (the archive holds 20,184 rows to
10,615 live, so a live-only count is a 7-day window — see `083`), joined to `pages.rebuild_policy`,
terminal outcomes only:

| item_type | `generic` pages | `owned` pages |
|---|---|---|
| `phantom_internal_link` | 101 ok / 46 failed = **69%** | **0 ok / 14 failed = 0%** |
| `empty_internal_href` | 9 / 2 = **82%** | **0 / 4 = 0%** |
| `literal_markdown` | 3 / 20 = 13% | **0 / 16 = 0%** |
| `placeholder_contact` | 0 / 2 | **0 / 4 = 0%** |
| **total, owned** | — | **0 ok / 38 failed** |

A uniform zero across four unrelated item types is not a handler having a bad run; it is a route
that cannot work. The live error is your guard speaking: *"page tool-archetype-taster-quiz is
rebuild_policy=owned (tool/widget-owned): a generic section save would clobber it. Use
`apply_section_edit` for targeted edits or the tool pipeline for rebuilds. Refusing to overwrite."*

### 2. There are ~134 more queued behind it

Non-terminal findings (`detected`/`needs_human_review`/`unresolved`/`failed`) routed at
`page-build-handler` and sitting on `owned` pages, right now: `content_rewrite` 46,
`literal_markdown` 41, `needs_content_page` 14, `phantom_internal_link` 11, `placeholder_contact` 10,
`empty_internal_href` 6, `empty_section` 5, `tone_shift` 1. **Every one is a guaranteed refusal.**

### 3. THE PART YOUR FIX WOULD NOT CATCH: the refusal is recorded as `failed`, and a shared gate reads that

`bugs_open/083`'s promoter now holds any `(item_type, handler_agent)` pair whose lifetime success
rate falls below **25%** over ≥5 terminal outcomes (migration `444`, corrected by `454`, scope
corrected by `465`). **That floor is computed per pair with no regard to `rebuild_policy`** — so
owned-page refusals, which are a routing defect, arrive at the gate indistinguishable from
incompetence and drag the pair's ratio toward the floor. When a pair crosses it, the promoter stops
dispatching the type **entirely — including the findings on `generic` pages that were succeeding.**

`phantom_internal_link` is the live illustration: **69% on generic pages, 47% overall.** It is not
held today, and the only thing keeping it above the line is that its owned rows are a minority. Add
enough and a 69%-effective repair path switches off.

**So moving the guard earlier stops the wasted LLM call (your finding) but not this**: an early
refusal still terminates the item, and if it still terminates it as `failed`, the floor still reads
it. **Worth deciding, alongside the reordering, what an owned-page refusal should WRITE** — a
distinct terminal status, or a `wont_fix` with the reason, or a re-route — so that "this handler may
not touch this page" stops being counted as "this handler cannot do this job". That is a one-word
choice in the guard and it protects a fleet-wide gate.

### 4. Two honest limits on the above

- **This does NOT rescue `literal_markdown`.** Excluding owned pages it is still 3 ok / 20 failed =
  **13%**, well under the floor. Its 10 currently-held rows are correctly held and its real defect is
  `bugs_open/184`. I checked specifically because "the owned pages explain the failures" was the
  conclusion I wanted, and for that pair it is false.
- **`placeholder_contact` is 0/2 on generic pages** — too small a sample to say whether it has an
  independent problem. It reads "never succeeded" partly because 10 of its 13 rows are on owned
  pages and never had a fair test.

### 5. On process

Per the 2026-07-31 ruling I am declaring a substitution rather than running `090`: the structural
claim here is one query plus your own guard's error text, both first-hand and both re-runnable — the
0/38 table above, and `rebuild_policy` read from `pages`. I am also not filing this as its own bug,
because `301` owns the `page-build-handler`/owned-guard interaction and a second file would drift
from it. Contributed rather than acted on: the status question in §3 is your call or the owner's.
