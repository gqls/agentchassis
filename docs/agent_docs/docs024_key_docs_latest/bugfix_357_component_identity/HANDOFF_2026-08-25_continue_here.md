# HANDOFF — `bugs_open/357`, component identity — 2026-08-25

> ## ⚠ SUPERSEDED 2026-08-25 (evening) — read `HANDOFF_2026-08-25b_continue_here.md` first
>
> **This file's §4 ("THE OPEN QUESTION THAT MATTERS MOST") is ANSWERED and its premise was
> wrong in an instructive way.** It said phase 2's central claim was unproven because
> `adoption_candidates = 0` meant no page had ever arrived through the adopted route.
>
> **Two site adoptions that afternoon showed the route IS reached, with perfect inputs, and
> that the SAVE is refused afterwards by `save_page_sections`' prune floor** — 1 section
> against a 3- or 4-entry `pages.sections` plan written by the same action that chose the
> route. So `adopted = 0` was never evidence about phase 2 at all.
>
> **Phase 2 has since FIRED in production (12:24Z), twice, verified and serving.** The
> sections below on phases 0/2/3, the F2 guard, the arming state and the watch script all
> remain accurate — only §4's conclusion is retired.

**Read this first, then `bugs_open/357_HANDOFF_2026-08-22_a_whole_tool_page_is_stored_in_a_slot_that_claims_to_be_a_hero_component.md`.**

> **THE LANE IS NOT CLOSEABLE.** The bug's own complaint — 22 live rows declaring
> themselves the shared `hero` while storing a whole interactive tool — is **still
> true of all 22**, and one of them was re-minted this morning. What has changed is
> that the machinery the repair depends on is now built, armed and *proven*, and the
> repair itself is written and waiting on one precondition that has never been met.

---

## 1. State in one table

| | state | evidence |
|---|---|---|
| **Phase 0** — the provenance stamp actually reaches the database | **DONE and PROVEN AT VOLUME** | 571 saves since arming, **570 stamped**; the 1 unstamped is explained below and is the guard working |
| **F2 guard** — a carried tool must not acquire the *discarded* render's stamp | **PROVEN, with demand** (was "pending/vacuous" until 09:08 today) | see §3 — this is the day's important result |
| **Phase 2** — stop the producer mislabelling at birth | **ARMED, live, council-APPROVED — but its adoption path has NEVER FIRED** | `adopted=0`; `adoption_candidates=0` since arming |
| **Phase 3** — repair the 22 | **WRITTEN, COMMITTED, DELIBERATELY UNAPPLIED** | `sql_for_agents/578_retype_mislabelled_tool_rows_HOLD.sql` |
| **The bug itself** | **OPEN. 22 rows, one re-minted 2026-08-25 09:08** | population query in the bug file |

Fresh build `agent-chassis-67fd9c76f5` carries the capability — probed at the running
binary with a must-be-present control and a must-be-absent control, both correct.

---

## 2. What is armed, and how to undo it

`adopt_unidentified_fragments = true` on **all six** live `save_page_sections` steps
(applied by hand 2026-08-24 16:15Z on the owner's instruction), plus the
`adopted-fragment` component seeded.

```
577_seed_adopted_fragment_component_HOLD.sql          <- applied
579_enable_adopt_unidentified_fragments_HOLD.sql      <- applied
579_..._HOLD_ROLLBACK.sql                             <- the exact inverse, if needed
```

Verify the armed state independently rather than trusting the migration's output —
the recursive query is in `RUNBOOK_component_identity.md` under **ARMED 2026-08-24**.
Three of the six steps are nested inside `sub_workflow`s and a top-level `jsonb_each`
misses them.

⚠ **A disarm does NOT un-adopt rows already adopted** (there are none yet). See the
rollback file's own header.

---

## 3. THE DAY'S RESULT: the F2 guard passed a real test

Yesterday this read zero **with no demand** — no population row had been rebuilt, so
the zero meant nothing, and the bug file said so. At **2026-08-25 09:08:20** it got
its demand:

- `vetcomparison.uk/index`, slot `hero`, **re-minted through `save_page_sections`**
  (`content_brief` present) with the stored 11,326-byte tool spliced back in by
  Layer 2.
- **It received NO stamp** — and it is the **only** unstamped save of 571 since
  arming. Every other save that day was stamped.

That is the guard discriminating, not a mechanism that stamps nothing: a stamp on
that row would have named the *hero* template as the producer of a whole interactive
tool, which is the "worse than no stamp" case phase 0 was built to prevent. **Demand
present, correct refusal, everything else stamped.** Treat this as settled.

---

## 4. THE OPEN QUESTION THAT MATTERS MOST

**Phase 2 has never actually adopted anything, so its central claim is unproven in
production.**

`adoption_candidates = 0` since arming — no page has arrived through the route
adoption acts on (the no-`<section>` fallback, which stores a whole fragment as ONE
section and sets `SectionData.FallbackAdopted`). 571 ordinary saves went through the
armed seam and correctly produced no adoptions.

**And there is positive evidence of a second mint route that phase 2 does not
cover.** Two facts, both measured 2026-08-25:

1. The row re-minted this morning came through `save_page_sections` on a page with
   **four** rows — not the one-section fallback. It kept its `hero` binding because
   `carriedIdentity` deliberately only carries identity for `adopted-fragment` rows
   (the council's round-2 narrowing). **For an existing mislabelled row that is BY
   DESIGN** — phase 3 is its remedy, not phase 2.
2. But **every multi-row affected page has rows from DIFFERENT saves** (2 rows = 2
   distinct `created_at`; `vetcomparison/index` = 4 rows, 4 distinct times). The
   fallback emits exactly one section, so the multi-row pages did not get their hero
   row from a single fallback save. Rows are accumulating across saves rather than
   being replaced.

**What a successor must establish, and it is the top priority:** *by what route does
a NEW mislabelled row appear, and does phase 2 intercept that route?* Today the honest
answer is "unknown — the only route we have instrumented has never fired." Do not
assume phase 2 works because it is armed and approved; this lane has already lost a
day to a mechanism that was approved, rolled and doing nothing.

Suggested first step: `page_component_history` for `vetcomparison.uk/index` and one
`mortgagecalculator` tool page, to reconstruct the sequence of saves and see which
one introduces the `hero` binding.

---

## 5. Phase 3 — the repair — and its one blocking precondition

`578_retype_mislabelled_tool_rows_HOLD.sql`. Per row: `component_id` →
`adopted-fragment`, `content_data` → `{"body": <the stored bytes>}`,
`component_version_id` → the `{{.body}}` version. **Leaves `slot_name`, `position`,
`rendered_html`, `rendered_html_digest` and `pages.sections` untouched** — that is the
whole safety argument, because Layer 2 matches on slot-name equality and a rename
makes the next rebuild append the tool beside a fresh hero band.

Targets **all 22**, including the six `rebuild_policy='owned'` pages (owner
instruction 2026-08-24). Those six are the only rows phase 2 can **never** heal: the
owned-page guard returns at `save_page_sections_action.go:186` and adoption runs at
`:397`.

> **⚠ IT WILL REFUSE TO RUN TODAY, CORRECTLY.** Its precondition is that at least one
> **organically adopted row carrying a stamp** exists — proof that the shape it is
> about to create has been demonstrated in production. `adopted = 0`, so it raises and
> aborts. **Do not weaken that check to get it to run.** It is the owner's ruling
> ("once option 1 has been built" = live and readable) expressed as code.

So the order is: **make adoption fire once → verify that row → then run 578.**

---

## 6. Everything already proven, so nobody re-derives it

- The stamp is **true**, not merely present: of 245 stamped rows checked 08-24, 239
  named their component's current template and the 6 that did not were one component
  edited *after* those rows were written — a stale row, which is the mechanism
  working.
- Version churn is bounded: **1.00** `render_stamp` rows per component. Not a log.
- Nothing backfills: pre-roll cohort stamped stayed **0 of 987**. That asymmetry is
  the control.
- The tool pipeline (`create_tool_component` / `deploy_tool`) **types its rows
  correctly** with a bespoke component — it is not a source of this defect, and its
  rows can never be adopted (adoption lives inside `save_page_sections`).
- The three `loancash.co.uk` verbatim pages are excluded from the repair
  **structurally** — none is bound to `hero`.
- The six `owned` pages are **not** verbatim (`deploy_mode` = NONE on all six), so
  re-typing them cannot trigger the verbatim↔assembled flip.

---

## 7. Named follow-ups — real, and none of them blocking

1. **The rerender re-derives identity by NAME.** `resolveComponent`
   (`rerender_page_sections_action.go:377`) falls through to the slot-name map when
   `component_id` is empty, so a row that fails adoption gets re-bound to `hero` by
   the next rerender. Phase 2 *reduces* the mint; it does not eliminate it for that
   case.
2. **Five other `page_components` writers do not stamp.** First live instance seen
   08-24: `agritec.uk/tool-sfi26-revenue-stacker`, correctly typed by the tool
   pipeline and unstamped.
3. **`page_component_history` drops the stamp** — 15 columns, none of them
   `component_version_id`, so a row's provenance is lost the moment it is archived.
4. **`save_page_sections` registers no `ActionInputSpec`**, so
   `adopt_unidentified_fragments` is invisible to the RFC_022 optional-key budget —
   "not counted", which is not "under budget".

---

## 8. Watching it

`watch_357_adoption.sh` (in this directory, committed). Run it; a tick reads:

```
adopted=0 population=22 population_stamped=0 adoption_candidates=0 saves_since_arming=571
```

Every number needed to interpret a zero is on the line. **Read
`adoption_candidates` FIRST** — 0 means no qualifying page was saved, so
`adopted=0` says nothing about correctness.

Exit codes: `0` adoption observed · **`42` STOP condition** · `3` could not measure ·
`1` window elapsed. **42 deliberately, because `2` is what bash returns for a syntax
error** and the two were once indistinguishable. Do not edit the file while a copy is
running.

**The three STOP conditions, each invisible to "is the tool still there?":** a page's
row count going UP by one (the carry-forward landmine); a population row acquiring a
stamp (splice hygiene failed — now proven to hold, so a firing would be a regression);
a new fragment landing with `component_id` NULL.

---

## 9. Council and commits

All three rounds **APPROVED**: phase 0 `73a638c7`, phase 2 `74e4c1fd` (after one
REVISE whose two findings were both real), phase 1 `62aac6c2` (2026-08-22).

Register: **CLC-028** (the carry contract), **CLC-026** corrected to LIVE AND PROVEN.
Landmines: the severed-carrier class, and the `workItemTerminalStatuses` reading trap.
`WRONG_CALLS.md` entries **6–12** are from this lane; the last five share one shape —
*a real check standing in for a question it was never asked.*
