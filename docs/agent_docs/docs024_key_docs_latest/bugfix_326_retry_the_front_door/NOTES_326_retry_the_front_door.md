# NOTES — bugs_open/326, the front door cannot be re-used

Append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## 2026-08-23 ~17:00Z — session start, ownership and validity

`scripts/who-owns.py 326` says OWNED-or-recently-active, but the only commit it
finds is `ace492170`, the filing itself. The filing lane is
`docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/` (handoff
`HANDOFF_2026-08-19_fixing_the_one_shot_route.md`, commit `1eb446f89`). That lane
owns the ACCOUNT of the one-shot route; nobody has taken the FIX. The diagnosis
queue (`site_work_items` `needs_diagnosis` / `awaiting_diagnosis`) was **empty**
`[MEASURED 2026-08-23]` before I filed into it.

090 filed anyway, because I was about to assert a structural root cause that
CONTRADICTS the filed one: intake `df0b3d97-bb22-4b42-a85c-030ff42536cf`,
run `655c8508-a265-4926-8625-b54beff26869`.

## 2026-08-23 — the filed root cause is WRONG, and the live index says so

`bugs_open/326` states: *"`create_work_item`'s step matches the previous attempt's
item by `item_key` — regardless of that item's status, including `complete` and
`cancelled`"*. That is not what the estate does.

`\d site_work_items` on the live DB `[MEASURED 2026-08-23]`:

```
"idx_swi_dedup" UNIQUE, btree (site_id, item_key)
  WHERE item_key IS NOT NULL
    AND (status <> ALL (ARRAY['complete','verified','rejected','wont_fix',
                              'failed','unresolved','cancelled']))
```

`complete` and `cancelled` are BOTH excluded from the index. A `complete`
predecessor cannot hold the dedup slot, and the `ON CONFLICT (site_id, item_key)
… DO NOTHING` in `writeWorkItem` (`platform/orchestration/actions/load_work_item_actions.go:1650`)
names exactly that predicate. So the index arm is innocent.

## 2026-08-23 — the real mechanism: the ANTI-CHURN block, two arms

`writeWorkItem` runs a block BEFORE the insert, for any item with a non-empty
`itemKey` and `recurrenceExpected == false`
(`load_work_item_actions.go:1516`). It counts siblings on
`(site_id, item_key)` at status `complete`/`failed` within 7 days:

- **Arm A — within-cycle suppression.** `newestAge < 3.0` hours ⇒
  `return workItemWrite{}, nil`. **No row, no error.** The caller gets
  `Inserted:false`, which `create_work_item` reports to config as
  `deduped: true` — byte-identical to a genuine open-item dedup.
- **Arm B — two-strike.** `terminalCount >= 2` ⇒ the arriving item's status is
  rewritten to `unresolved` and its summary prefixed `[unresolved after N attempts]`.
  `unresolved` is in `workItemTerminalStatuses` AND is not in
  `workItemDispatchableStatuses` (`triaged`, `approved` only), so the row is born
  **terminal, undispatchable, and outside the dedup index** — every later attempt
  repeats the same fate.

## 2026-08-23 — MISSTEP: my first live check could not have come out otherwise

I ran a query classifying every `deduped:true` result in retained
`orchestration_states` by whether an OPEN row holds that `(site_id, item_key)`
**now**, and got `open_holder = 0` for both distinct keys — and read that as
"the index cannot have been the mechanism". **It establishes nothing.** The
holder was open AT THE TIME and had since completed. Redone with timestamps
(`ev.updated_at` vs the row's `created_at`/`completed_at`), all **36** dedup
events resolve as legitimate index dedups: the newest dedup on each key is
~20 seconds after the row was created and ~5 minutes before it completed.

That is the right answer and a useful NEGATIVE CONTROL — Arm C works as designed
and is distinguishable — but I nearly banked the wrong one. Logged in
`WRONG_CALLS.md`.

## 2026-08-23 — the timing that settles it: loanzy.uk, three submissions

`domain-submitter` writes a `submission` spec BEFORE the `create_work_item` step,
so `site_specs` dates every submission even after `orchestration_states` is
reaped (retention is ~24h; the bug's correlation
`3296ac3a-30bd-4db5-84ae-16260394f3bc` is long gone).

```sql
SELECT created_at, is_current FROM site_specs
WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='submission'
ORDER BY created_at;
```

| # | submission spec | outcome |
|---|---|---|
| 1 | **12:53:00.04Z** | filed `research_loanzy.uk` at 12:53:00.25, ran, `complete` at 13:36:36 |
| 2 | **15:21:17.23Z** | **the deduped one** — `inserted:false`, no row |
| 3 | **20:16:12.61Z** | filed a new `research_loanzy.uk` at 20:16:13.85 |

Submission 2 landed **2 hours 28 minutes** after the terminal sibling was created.
Arm A's window is **3.0 hours**. `[MEASURED 2026-08-23]` — and the measurement
could have come out otherwise: at 3h01m the insert would have succeeded and there
would have been no bug to file.

**Consequence the bug file does not have.** The 78-row hand-rename of `item_key`s
was **not** what made submission 3 possible `[INFERRED 2026-08-23 from the code
path + these timings — to be proven with a test, not asserted]`: by 20:16 the
newest terminal sibling was 7.4 hours old, `terminalCount` was 1, and `complete`
is outside `idx_swi_dedup`. Waiting until 15:53Z would have done the same job.
The documented recovery procedure is hand-surgery for a three-hour timer.

## 2026-08-23 — what IS permanent, and it is Arm B

Arm A expires. Arm B does not, and nothing in the bug file mentions it.
Two terminal attempts on one key inside 7 days ⇒ **every** further attempt is born
`unresolved`: terminal, undispatchable, invisible to the dedup index, and — for
build-pipeline stage types — drained by nothing. `reviewRevalidators`
(`revalidate_review_queue_action.go:277`) covers six types
(`unresolved_cta`, `required_fields_missing`, `needs_section_data`, `needs_page`,
`voice_tells`, unverified claims). **None** of `needs_domain_research`,
`needs_strategy`, `needs_briefing`, `needs_site_plan` is among them.

Fleet-wide `[MEASURED 2026-08-23]`:

```sql
SELECT count(*) FILTER (WHERE summary LIKE '[unresolved after%') AS born_two_strike,
       count(*) AS all_unresolved FROM site_work_items WHERE status='unresolved';
--  635 | 747
```

**635 of 747** `unresolved` rows carry the two-strike stamp. The biggest
populations are `page_rerender` (212) and `improve_tool` (205) — both ACTION
REQUESTS, which is the class `recurrenceExpected` exists to exempt.

## 2026-08-23 — the class, measured at the config

> **CORRECTED 2026-08-23, same session — the count below was 17 and it was an
> UNDERCOUNT, from a walk that only descended one level.** My first census did
> `jsonb_each(default_config->'workflow'->'steps')`, which never enters
> `sub_workflow`. A properly recursive walk gives **22** `create_work_item` steps,
> **21** with an `item_key_prefix`, **19** of those undeclared and **2** declaring
> `recurrence_expected: true`. **Eight of the nineteen are nested**, so the flat
> walk missed nearly half of them — including both `tool-auditor` steps, both
> `tool-suggester` steps, `internal-linker` and `component-quality-auditor`.
> This is `bugs_open/144`'s cost reproduced exactly, in the same session that
> then wrote a check whose header cites it. What caught it: the Go audit
> (`config-key-audit --undeclared-recurrence`, which uses `validation.WalkSteps`)
> reported **19** against my **15-ish**, and the disagreement was the tell.
> The two now agree — an independent cross-check from a different code path.

Live `create_work_item` steps carrying an `item_key_prefix`, and whether they
opt out of anti-churn `[SUPERSEDED — see the correction above; 17 was a flat walk]`:

- `recurrence_expected: true` on **2 of 17** — `improvement-loop.record_not_converging`
  and `tool-improver.create_rerender_item`.
- The entire one-shot build chain is in the other 15, unexempted:
  `domain-submitter` (`research`) → `domain-research-classifier` (`vertical_research`)
  → `vertical-exemplar-researcher` (`strategy`) → `domain-strategist` (`briefing`)
  → `build-briefing-agent` (`site_plan`).

Go-side: `insertWorkItem` has **36** non-test call sites and `writeWorkItem` **8**
`[MEASURED 2026-08-23]`; **10** places set `recurrenceExpected: true`.
(Census date recorded per the owner ruling of 2026-08-22 — re-run
`git log --since=2026-08-23 --diff-filter=A -- platform/orchestration/actions/`
before quoting these.)

## 2026-08-23 — the 090 loop produced NO VERDICT, and I am saying so rather than skipping it

Intake `df0b3d97-bb22-4b42-a85c-030ff42536cf`, run `655c8508-a265-4926-8625-b54beff26869`.
It ran to `COMPLETED` and wrote **5 bundles and nothing else** — `SELECT DISTINCT kind`
over `diagnosis_artifacts` returns `bundle` alone; no `iteration_note`, no
`council_report`, no `doc_notes` row.

This is the documented shape, and the SessionStart hook warned me about it before I
filed: **`load_work_item_actions.go` is 82,669 bytes**, and a 090 whose scope lands in
a file over ~60KB returns bundles and no verdict. The discriminating check (only
readable after the first bundle) confirmed it on iteration 2:

```
_(body omitted — 25397 chars, and 50544 of the 60000-char body budget is already
  spent. It was found; it did not fit.)_
```

**Demand control, so the silence means something:** `doc_notes` took **21 rows** in the
same two-hour window. The recorder was alive; it simply had nothing to say about my run.

Iteration 2 *did* render both deciding arms (`newestAge < 3.0` and `terminalCount >= 2`
both present in the bundle body), so the diagnoser saw the mechanism — it just never
wrote a conclusion.

**So I take the owner ruling's escape hatch EXPLICITLY** (2026-07-31: a `bugs_open/`
file asserting a cross-cutting root cause is not filed until it has been through the
loop, *or the filing session states plainly why it substituted equivalent first-hand
verification*). Substituted here, and this is the substitute: the live index predicate
read from `\d site_work_items`; the whole anti-churn block read in place; the
`site_specs` submission timings; the 36-event negative control on the dedup arm; and
the five mutations below. Every one is in this file with its query.

## 2026-08-23 — mutation proofs, because a passing test proves nothing on its own

Run against a `git archive HEAD` extract with only this lane's files overlaid (the
shared tree does not compile — see below). All five caught:

| # | mutation | caught by |
|---|---|---|
| M1 | remove the `retry_after` column append | arg-count mismatch, 4 tests |
| M2 | restore the legacy DROP on arm A | `..._WithinCycleWindow_DefersRatherThanDrops` |
| M3 | restore the `unresolved` BRAND on arm B | `..._TwoCompletedPredecessors_DeferRatherThanPoison` |
| M4 | delete the kill switch | both `TestAntiChurn_KillSwitch_...` subtests |
| M5 | flat 3h interval instead of the window REMAINDER | 4 tests, incl. all three boundary cases |

## 2026-08-23 — three shared-tree hazards, all verified first-hand

1. **`load_work_item_actions.go` is dirty by a THIRD session** — `+4/−1` at
   `FailWorkItemAction:1307`, `bugs_open/345` candidate 2. Flagged by the loanzy.uk
   lane; confirmed with `git diff --numstat` rather than taken on report.
2. **The working tree does not compile**, and this is worse than the first item.
   `go build ./platform/orchestration/actions/` exits **1**: `applyCTARecompute` at
   `rerender_page_sections_action.go:550,552` is mid-signature-change by another
   session. `git archive HEAD` of the same package exits **0**. So `make build-*`
   (committed HEAD) is fine and **anyone who builds an image today will not notice**,
   while anyone running `go test` on this package is blocked — an asymmetry that reads
   exactly like "my test setup is broken". Worked around with a HEAD-overlay script;
   see the RUNBOOK.
   - The 345 hunk is also **half-written**: the caller passes an 8th argument to
     `applyWorkItemFailureLadder`, whose callee is still the 7-argument HEAD version. So
     overlaying my own copy of the file onto HEAD *also* fails to build until their hunk
     is stripped from the build copy. Stripped there, never in the tree.
3. **`cmd/config-key-audit` tests are broken at HEAD+tree too** — `livedeclarations_test.go`
   references `livespec.DeferredDeclarations`, which another session's uncommitted
   `platform/livespec/livespec.go` no longer defines. `go vet` on a clean HEAD extract
   exits 0, so it is theirs, not mine. My own audit tests were run on an overlay.

## 2026-08-23 — the council submission was DISPATCHED TWICE, and the first one never left

`097` printed its `SAVE: SUBMISSION_CORR=94c196fa-…` summary and then failed:

```
Error from server (AlreadyExists): pods "kcat-cgate-1787507905" already exists
```

The publish is at `097:263`, **after** the summary at `:260`, so the script prints a
convincing receipt and then drops the message. The pod name is
`kcat-cgate-$(date +%s)` — **one-second resolution** — and another session ran `097` in
the same second.

**This is NOT the "a missing orchestration row is latency, do not retry" case**, and the
difference is worth stating because the standing guidance points the other way: there I
had a *positive failure signal* — `kubectl run` returned non-zero, so the container never
started and `kcat` never ran. Retrying cost nothing because nothing had been dispatched.
Re-run correlation `f610741f-5054-41e8-b0b7-54915d79ba92`, confirmed live at
`review_editquality` by querying `orchestration_states` on `fix_correlation_id`.
`94c196fa` should be treated as never having existed.

## 2026-08-23 — from the loanzy.uk lane (live greenfield build, garden-tools.uk)

- `research_garden-tools.uk`: `created_at` **17:17:15.482Z**, `completed_at`
  **17:44:59.593Z** — **27m44s apart**. The brake keys on `MAX(created_at)`, so the
  window closes at **20:17:15Z**, not 20:45. **An operator reasoning from `completed_at`
  gets the boundary wrong in the UNSAFE direction on a long-running item** — on a page
  build that gap could be hours. Worth its place in the bug file.
- Time-to-first-agent was **24m52s** (submit 17:17:18Z → claim 17:42:10Z), which is
  queue depth ÷ tick rate.
- `build-pipeline-trigger`'s `find_dispatchable_site` is **FIFO by work-item
  `created_at`**, one site per ~90s tick, and **a site with ANY `claimed` item is
  invisible** until it clears. So a re-submission staged while the site has something
  claimed will not run and **will look exactly like suppression**. Any Arm A
  verification must snapshot whether the site had a claimed item at that instant, or
  the negative result is uninterpretable.
- **REFUTED, by that lane, about its own earlier claim to me:** the picker does *not*
  walk sites in ascending `site_id`. That was 14 consecutive ordered samples that
  stopped being ordered twenty minutes later; the selector never mentions `site_id`
  except as a final tie-break. I had already written the `site_id` version into my plan
  and have removed it. Recording it here because I nearly carried a refuted mechanism
  into a bug file on a peer's say-so — **a subagent's or a peer's report is another
  doc**, and this one was corrected by its own author before I could ground it.
