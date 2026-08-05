# SESSION HANDOFF — 2026-08-05 21:00Z · five open threads, each with its own handoff

Written to continue in a fresh chat. **This file is an index with enough state to route; each
thread's own handoff is the authority.** Nothing here needs re-deriving.

Session ran 10:16Z → 21:00Z. Started on `bugfix_194`'s two owed checks; that led to
`bugs_open/201`, which the owner then chose as the work. **Chassis rolled twice today:
`v1.0.1252` at 09:10Z, `v1.0.1254` at 20:41Z (current, both replicas).**

## Commits from this session (all on `087_towards_multiple_domains`)

| commit | what |
|---|---|
| `131fabc59` | 194: candidate-list correction + `WRONG_CALLS` entry |
| `d513e1dca` | 194: the site-lock finding (a third silent vacuous-pass path) |
| `bf4f4e5ca` | 194: 3b blocked behind 201; CONTRIB into `bugs_open/201` |
| `37afbb847` | **201 fix-1** — three checks re-routed (`Council-Submitted:`) |
| `9eadf4908` | 201 lane docs (PLAN/RUNBOOK/NOTES/README) |
| `bbdfa2842` | 201 verdict follow-through + **RFC_014** |
| `ae6d8d062` | `WRONG_CALLS`: I grepped LANDMINES for the wrong end of the fix |

One `LANDMINES.md` entry (verification runs having three silent ways to pass vacuously) was
swept into another lane's commit `438834174` as a same-file passenger — nothing lost.

---

## THREAD 1 — `bugs_closed/194` check 3a · **effectively passing, formal read due 2026-08-06 09:10Z**
**Handoff:** `bugfix_194_sections_metadata_mapping/HANDOFF_2026-08-05_continue_here.md` §3a

At T+11h38m: **0 `CONTENT_DATA_REGRESSION`**, positive control 476 `TIMEOUT`, and — the part
that was missing this morning — **real traffic: `page-rerender` +120, `page-build-handler` +34**
on the fixed binary. This morning's zero was near-vacuous (1 run and 0 runs respectively) and
was recorded as such; that objection is now answered.

**To do:** the formal read tomorrow ≥09:10Z, **with the run counts in the same query** — a zero
without traffic is the trap this whole section exists for. Note the window spans `v1.0.1252` and
`v1.0.1254`; both carry the fix.

## THREAD 2 — `bugs_open/201` · **OPEN. Fix-1 live on v1.0.1254, UNPROVEN**
**Handoff:** `bugfix_201_page_content_writer_dispatch/HANDOFF_2026-08-05_continue_here.md`

Council **APPROVED** (`71523705-…`, 15 reviewers, 5 advisory, none high). Three discovery checks
re-pointed off direct `page-content-writer` dispatch, which hard-failed 11 of 11 because a
discovery spec carries no `sections` key.

**The single most important thing for the next session:** **a pod-grep cannot verify this
change** — it swaps one pre-existing string literal for another. My first RUNBOOK R7 grepped a
Go *comment* and returned 0 on a build that has the fix; it is corrected. Use **R7b**: newly
filed `literal_markdown`/`placeholder_contact` items must carry `page-build-handler`. **Measured
20:48Z: zero rows — the checks have not fired.** `quality-discovery-agent` (22 runs) last ran
12:14Z, before the roll. **Zero rows is not a pass.**

**Also owed:** symptom 2 (`mark_complete` trusts `handler_result` blindly — an item reached
`complete` having written nothing), deliberately left until after fix-1 per 201 §2.

**Known cost, not a bug:** the repair **rewrites the section and loses prior prose**
(`LANDMINES.md:4433`). Do not "fix" it with `mode=recreate` — that feeds a stale adoption-crawl
snapshot. Expect it during verification.

## THREAD 3 — `RFC_014` · **owner decision, nothing blocked on it**
**Handoff:** `architecture_review/RFC_014_handleragent_is_a_stringly_typed_routing_contract.md`

Raised by the council's `architecture` seat in the approved round: this was the **fifth** site
where a check named a `HandlerAgent` that could not consume the spec it files. The only guard
checks the agent **exists**, never that it can **consume the shape** — so site six passes CI too.
Three costed options; the cheap floor (a narrower legal-direct-dispatch set) is hours of work.
The seat said "ship it" — this is not a reason to revisit 201.

## THREAD 4 — `mortgagecalculator.co.uk`'s lock · **owner decision, and NOT mine to release**
**Owning lane:** `bugs_open/183` / the mortgagecalculator adoption lane. No handoff written by
me deliberately — I do not own it and should not open a competing lane.

`sites.locked_at` = 2026-08-03 10:30Z, `locked_by` = *"mortgagecalculator-adoption-lane:
composition+design done; held pending owner decision on page rebuilds"* — i.e. held pending a
decision on precisely the operation that would unblock 194's 3b.

⚠ **Do not release it to run a check.** `aee11cb90` is the incident where a live homepage was
rebuilt under a held lock on that same site; the lock is the control added afterwards. Also note
`load_work_items` returns **success with zero items** (`skipped_reason: site_locked`) for a
locked site — silently indistinguishable from an idle one.

## THREAD 5 — `ai-agent-orchestration.com` rebuild · **scoped, NOT started**
**Handoff:** `site_ai_agent_orchestration/HANDOFF_2026-08-05_rebuild_scope.md`

The owner asked for this. Site is **UNLOCKED**. Quantified: **31 of 106 components have NULL
`content_data` across 10 pages — on 9 of them every component on the page**; **5 pages have no
components at all**, two of which are marked `deployed`; 42 stuck `page_rerender` items (overlap
with the NULL pages is **partial** — ~half — and I did **not** establish the other half's cause).

Three scope options costed in the handoff. **Must go through the framework** (owner ruling
2026-08-04 — never hand-build). ⚠ A rebuild **regenerates copy**, so it is free on the 5 empty
pages and a real cost on the 10 that currently serve good prose. That asymmetry is the argument
for doing the empty ones first.

**It cannot serve 194's check 3b** — 0 qualifying items on two independent clauses. Do not retry.

---

## Cross-cutting things this session learned the hard way

- **A count of work that *looks* ready is not a count of work the loader will load.** The 194
  handoff's seven-site candidate list was measured against a different predicate than the one
  gating the loop; all seven were 0. (`WRONG_CALLS.md`)
- **Grep `LANDMINES.md` for the symbol you are ROUTING TO, not just the one you are FIXING.** A
  fix has two ends; I searched the broken one thoroughly and treated the destination as safe
  because it runs in production. "It works in production" answers *does it run*, never *what
  does it do to my case*. (`WRONG_CALLS.md`)
- **A pod-grep only verifies a change that ADDS or REMOVES a string.** A re-point between two
  existing literals is invisible to it — and a grep for a Go *comment* always returns 0.
- **Read the schema first.** Two SQL errors this session (`orchestration_states.agent_type` is
  `owner_agent_type`; `jsonb_object_keys` cannot sit bare beside an aggregate). Both were loud;
  the same haste against a column that exists but means something else is the expensive version.
- **A top-level `#>> '{workflow,steps,…}'` silently misses steps nested in a loop
  `sub_workflow`** — and an empty result reads exactly like "no such step". Use
  `jsonb_path_query(…, '$.**.steps')`. Hit twice today.
- **`who-owns.py` can name your own lane** if you have just written mentions into its docs, and
  it cannot see a session mid-fix at all. Grep live `.jsonl` transcripts too — two other sessions
  were in `page-content-writer` code today.
