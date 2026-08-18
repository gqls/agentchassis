# HANDOFF — `ai-agent-orchestration.com`, a rebuild the owner asked for, scoped but NOT started

> ## ⛔ SUPERSEDED 2026-08-18 by `HANDOFF_2026-08-18_continue_here.md` — READ THAT FIRST.
>
> This file remains the record of the rebuild **scoping** and its `bugs_closed/194` analysis, both
> of which still hold. **Every figure below is from 2026-08-05 ~10:30Z and is stale.** The ones
> that mattered, re-measured 2026-08-17/18:
>
> | this file says | actually, 13 days on |
> |---|---|
> | 31 NULL `content_data` across 10 pages | **15 across 8 pages** — partly repaired |
> | 42 queued `page_rerender` not moving | drained; the fleet-wide stall is `bugs_open/029`, **owned by the owner's own thread** |
> | scope is an owner decision, NOT taken | **taken 2026-08-17** — contrast fixed at the template (migrations `456`+`457`), `pricing` rebuild approved |
>
> **§4 of this file is still correct and still worth reading**: this site cannot serve
> `bugs_closed/194`'s check 3b, and that has not changed.

**Written 2026-08-05 21:00Z.** The owner raised this unprompted: *"if you need a site to work
on ai-agent-orchestration.com is a decent one, it needs a rebuild anyway there is a lot missing
on it."* It was offered as a target for `bugs_closed/194`'s acceptance check; **it cannot serve
that purpose** (see §4), but the rebuild itself is real, wanted, and **not started**.

**Nothing has been dispatched at this site by this lane.** All figures below are read-only,
measured 2026-08-05 ~10:30Z. Re-measure before acting — they are hours old and the site has
scheduled automation running against it.

## 1. Site facts

- `sites.id` = `2a8ebf9c-20a2-4c39-b191-840b012371da`, status `deployed`, **UNLOCKED**
  (`locked_at` NULL — unlike `mortgagecalculator.co.uk`).
- 38 pages: **32 `rebuild_policy='generic'`, 6 `owned`.** The `owned` guard blocks little here
  (it is what blocked `bugs_closed/087`'s run elsewhere — check it before any rebuild).
- Regular automation already touches it: `model-directory-publisher`, `feed-ingester`,
  `content-feed-orchestrator`, `build-dispatch-loop` all ran on 08-05.

## 2. What is actually wrong — "a lot missing", quantified

**(a) 31 of 106 `page_components` have `content_data` NULL, across 10 pages — and on 9 of those
10 it is EVERY component on the page** (3/3, 3/3, 5/5 …), not a partial loss. `blog` is the only
partial (2/3). Affected pages include `pricing` (5/5),
`ai-agent-observability-2025-what-teams-are-actually-monitoring`,
`building-a-hierarchical-agent-system-with-kafka-and-postgres`,
`deploying-ai-agents-kubernetes-practical-guide`,
`llm-provider-abstraction-production-agent-systems`,
`multi-agent-state-management-distributed-systems`,
`orchestrating-ai-agents-in-production-what-actually-breaks`,
`the-enterprise-ai-agent-adoption-gap-2025`, `tool-ai-agent-roi-estimator-guide`.

**This is `bugs_closed/194`'s damage class, live on this site.** These pages *serve correctly* —
the loss only bites when something tries to regenerate them, because `content_data` is the only
thing `rerender_page_sections` can rebuild a section from. Two of the site's own escalations say
so in as many words: `needs_page` items summarised *"a section had no stored content_data"* —
`containment-first-architecture` (07-31, still `needs_human_review`) and `services` (complete).

⚠ **Repair is re-running the build. NEVER restore from `page_component_history`** — its
`component_id` is NULLed by `ON DELETE SET NULL`, so pairing yesterday's content with today's
HTML makes the next rerender reinstate the old page (`bugs_closed/194` §4).

**(b) 5 pages have NO components at all**: `agent-complexity-estimator-guide`,
`ai-readiness-quiz-guide`, `tool-llm-cost-calculator-guide` (all `build_status='needs_rebuild'`),
plus **`llm-cost-calculator` and `roi-estimator`, which are marked `deployed` while holding
nothing.** Those last two are the sharpest "missing" — a deployed page with no content.

**(c) 42 queued `page_rerender` items not moving** (21 `detected`, 21 `unresolved`, handler
`page-rerender`). **[MEASURED, and deliberately NOT asserted as the cause]** overlap with the
NULL pages is **partial**: 10 of 21 detected and 9 of 21 unresolved sit on a page with a NULL
component. So roughly half the stuck queue is explained by the NULLs and **half is not** — I did
not establish why, and nobody should write "the NULLs caused the stuck queue" without doing so.

**(d) 123 other open work items**, incl. 22 `cta_names_unknown_destination` and 21
`required_fields_missing` at `needs_human_review`. 6 items are claimed by `build-dispatch-loop`
but parked at `needs_human_review` since 08-04 — **not in flight.**

## 3. How to do the rebuild — and the rule that constrains it

**OWNER RULING 2026-08-04, in CLAUDE.md: EVERY SITE GOES THROUGH THE FRAMEWORK. Never
hand-build one** — no hand-authored HTML to the bucket, however small or temporary. The
sanctioned route is to seed the site row/specs and dispatch
`scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh <domain> --email … --mission-file …`.
The ruling exists because a session hand-wrote the webdesign.uk shopfront; the fast path
produces an artefact nobody can audit and silently opts out of every pipeline control
(`evidence_base` gating, banned-claim sweeps, discovery checks, imagery style, rerender).

**Scope is an owner decision and was NOT taken.** Three shapes, cheapest first:
1. **The 5 empty pages only** — smallest spend, fixes the most visibly broken thing (two
   `deployed` pages serving nothing).
2. **The 10 NULL-`content_data` pages** — repairs the 194 damage so future rerenders work.
   Bigger spend; these pages currently *look* fine to a visitor.
3. **Full domain re-submission via 082** — the sanctioned "rebuild", largest spend, touches all
   32 `generic` pages.

⚠ **Whichever is chosen, a rebuild REGENERATES copy.** See `bugs_open/201`'s
`LANDMINES.md:4433` finding: a page-build repair does not edit, it rewrites — the existing prose
on those pages will be replaced, not corrected. On the 5 empty pages that is free (nothing to
lose). On the 10 NULL pages it is a real cost, because they currently serve good copy.
**That asymmetry is the main argument for doing (1) first and (2) deliberately.**

## 4. Why this site cannot serve `bugs_closed/194`'s check 3b — do not retry it

Measured against the predicate that actually gates `site-work-orchestrator`'s build loop
(`load_work_item_actions.go:623-661` + the step's filters:
`status IN ('triaged','approved') AND pipeline='build' AND handler_agent='page-content-writer'`),
this site returns **0**, failing on **two independent clauses**: not one of its items carries
that handler, and none is `triaged`/`approved`. A maintenance run here completes **green having
never executed the code under test**. Fleet-wide exactly one item qualifies, on a locked site.
Full write-up: `bugfix_194_sections_metadata_mapping/` (HANDOFF §3b, RUNBOOK R7) and
`WRONG_CALLS.md` 2026-08-05.

## 5. Before touching it

- Re-measure everything above; these numbers are from 08-05 ~10:30Z.
- Check nothing else is mid-flight: `scripts/who-owns.py`, the open-work-item query, **and**
  grep live `.jsonl` transcripts — commit history cannot show a session mid-fix.
- `bugfix_128_image_url_404` and `bugfix_149_nav_membership` both cite this domain; the most
  recent site-specific commit is `f7a2441c0` (08-03, "2 rerenders, nothing else").
- The site is UNLOCKED, so nothing will stop a dispatch — **that is a reason for care, not
  permission.**
