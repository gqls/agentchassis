# HANDOFF — start here: `bugs_open/326`, and the `345` adoption

> ## ⚠ CLOSE-OUT UPDATE, 2026-08-24 evening — read THIS table first; the sections below hold the reasoning
>
> The owner's ruling ("D + E now, census alongside") is **executed**. Everything below the
> table is context; the table is the whole remaining estate of this lane.
>
> | item | state | who owes what |
> |---|---|---|
> | **Migration 572** (customer path) | **LIVE + PROVEN end to end** (garden-tools.uk ran to `page_rerender complete` off a re-submission inside the window) | nobody — done |
> | **D** — within-cycle arm DEFERS (`f16c87beb`) | committed; **council APPROVED round 1** (`74d4fa7d`, 2 advisories, both acted on: doc_notes decision row written; recursive census re-run, still 0; the three-kill-switch relationship recorded). Inert until a roll | post-roll: binary-grep the NEW literal `deferred — terminal item too recent` (must-present) beside a must-absent control, per service |
> | **E** — 8 `recurrenceExpected` flags (`e4d20d97a` + `69dc4b653`) | committed; round 1 **REVISE** (process: I disclosed the ninth edit in risks instead of the edit list — conceded, it was a violation); **ROUND 2 IN FLIGHT**, same corr `7710367e`, run `751ff90e`, every objection answered with measurements | **the successor's ONE active duty: read round 2's verdict** (`SELECT body FROM doc_notes WHERE categories ? 'council-gate' AND body LIKE '%7710367e%' ORDER BY created_at DESC LIMIT 1`) and act on it — the code is on the shared branch either way. Post-roll E check is BEHAVIOURAL (a bool leaves no literal): a within-window repeat through a flagged producer inserts with `retry_after` NULL; through an unflagged one, defers |
> | **Migration 573 / `on_dedup`** (decision 2) | **UNRULED. 573 stays `_HOLD`.** The gate grep (`grep -rn '"on_dedup"' …/create_work_item_action.go`) still returns empty and MUST before any apply | the owner — yes (one small council round ships the key, then apply 573 after a roll) or no (delete 573) |
> | **14 undeclared config steps** | census recorded in NOTES, names listed; protected meanwhile by D's floor (delayed, never destroyed) | the owner names who tells the lanes (decision 3); the lanes classify |
> | **661-row two-strike landfill** (247 keys) | deliberately untouched; growth through the within-cycle arm ends when D rolls | RFC_010 / `bugs_open/033` D2 — already the owner's open decision |
> | **230 detector rows / fixer-lies class** | not this lane's | `bugs_open/352` |
> | **`retry_after` dual-meaning residual** | named in RFC_048 §4 and D's submission; a live-DB skip-then-serve test is the part sqlmock cannot supply | whoever next touches a `retry_after` reader |
> | **Kill switches near this column — THREE, all armed** | `DISABLE_ANTI_CHURN_DEFERRAL` (D: restores the silent drop), `DISABLE_WORK_ITEM_RETRY_BACKOFF` (failure ladder), `DISABLE_OWNED_PAGE_DOOR_DEMOTION` (333's door). Three writers gated; the shared READ predicate is unswitched | operators: reach for the right lever — the D one is the 326 rollback |
>
> The bug file itself stays in `bugs_open/` until D and E are **live on a roll** and the
> post-roll checks above pass — the estate's fixed-AND-live bar.

**Written 2026-08-24 by the `bugs_open/326` session.** Cold-start doc for this lane.
Read this first; the detail is in `NOTES_326_retry_the_front_door.md` (technical log, newest at
the bottom) and `README_where_we_are.md` (plain prose).

---

## ⚠ THE ONE THING THAT CAN DO DAMAGE: DO NOT APPLY MIGRATION 573

`docs/agent_docs/sql_for_agents/573_domain_submitter_refuses_to_report_success_over_nothing_HOLD.sql`

**A fresh chassis does NOT make this applicable, and that is the trap.** The natural reasoning —
*"573 was held for the roll, the fleet has rolled, so apply it"* — is wrong here, because the
code it depends on **was vetoed and never committed**. There is no build, present or future, that
carries `on_dedup` until `RFC_048` is decided and the patch is landed.

Applying it early does not degrade gracefully and does not fail once: `create_work_item` is
`StrictConfig` and `ValidateWorkflow` runs on **every message**, so an unrecognised key fails
**every `domain-submitter` run for as long as it is applied** — it takes the customer front door
down entirely.

**The check before you ever apply it**, and it must come back non-empty:
```bash
grep -rn '"on_dedup"' platform/orchestration/actions/create_work_item_action.go
```
Empty ⇒ the code is not in the tree ⇒ **573 stays held.** Verified empty as of 2026-08-24, and
`on_dedup` appears in **0** live agent definitions.

---

## Where 326 stands: FIXED, LIVE, and proven end to end

The customer path works. `docs/agent_docs/sql_for_agents/572_build_chain_declares_recurrence_expected.sql`
is **applied** (recorded via `--record-only`; the runner's `--apply` would have swept every other
lane's pending file).

**Verified at the artefact, twice, and it went further than the acceptance test asked.**
`garden-tools.uk` was a real greenfield build that died at hop two on an unrelated defect
(`bugs_open/376`). Re-submitted at **2h05m51s** — inside the 3.0h window that would have
swallowed it that morning — it queued a distinct row and then ran the whole pipeline:

```
needs_domain_research   complete 17:17:15   <- submission 1
needs_vertical_research failed   17:44:56   <- died here
needs_domain_research   complete 19:23:06   <- submission 2, THE FIX
needs_vertical_research complete 19:26:57
needs_strategy          claimed  20:05:55
needs_briefing          triaged  20:09:29
…                       page_rerender complete (2026-08-24)
```

Live state as of 2026-08-24: **5 of 5** build-chain steps declare `recurrence_expected: true`.

**The bug file is still OPEN and should stay open**, because what is fixed is the five build-chain
steps, not the mechanism. See the residual below.

### What was actually wrong (the filed root cause was incorrect)

`bugs_open/326` blamed `create_work_item` deduping on `item_key` "in ANY status". **`idx_swi_dedup`
excludes terminal statuses including `complete` and `cancelled`**, so a finished predecessor
cannot hold the slot. The real mechanism is the anti-churn brake *above* the insert
(`load_work_item_actions.go`, the block before `writeWorkItem`'s INSERT): under 3.0h since the
newest `complete`/`failed` sibling it returned **no row and no error**, reporting `deduped:true` —
indistinguishable from a genuine dedup. The correction, with evidence, is at the top of the bug
file.

---

## `RFC_048`: RULED AND EXECUTED (2026-08-24, later the same day)

> **SUPERSEDES the section below, which is kept for the reasoning.** The owner ruled
> **"D + E now, census alongside"**. Both are landed: **D** = `f16c87beb` (within-cycle arm
> defers, two-strike unchanged, kill switch `DISABLE_ANTI_CHURN_DEFERRAL` armed, council round
> `74d4fa7d`); **E** = `e4d20d97a` (eight `recurrenceExpected` declarations on Go action-request
> producers, council round `7710367e`). Go is inert until a roll — check the service's
> `build provenance` stamp before assuming either is live. The DO-NOT-APPLY on 573 above is
> UNCHANGED: decision 2 was not ruled, D deliberately did not add `on_dedup`, and the gate grep
> still returns empty.

## The decision that WAS waiting on the owner: `RFC_048` (historical)

`docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_048_the_anti_churn_brake_may_delay_work_but_may_not_destroy_it.md`
with `RFC_048_proposed_deferral.patch` beside it.

The framework-wide half — making both brake arms **defer** (`retry_after`) instead of dropping or
burying work — was **REJECTED by the council gate on a guardian hard veto**
(corr `f610741f-5054-41e8-b0b7-54915d79ba92`): a fleet-wide default-behaviour change bundled into
an urgent point fix, when migration 572 alone closed the bug. **The veto was right and has not
been contested.** CLAUDE.md: a scope veto is not answered by resubmitting with better
measurements.

The RFC costs three options (ship as proposed / opt-in per caller / census-then-ratchet), states
which one this session would pick and marks it as a view. **Do not commit the patch without a
ruling** — on this tree committing is shipping.

**Owed work if it ever lands**, recorded in §4: no test exercises `retry_after`'s three readers
(`claim_work_item_action.go`, the dispatch loader, `complete_work_item_verification.go`) against a
row deferred **without** a prior failure. `retry_after` would then carry two causal meanings;
RFC_043 is the contract it extends.

---

## The residual, and it is the real remaining exposure

**14 keyed `create_work_item` steps still declare nothing**, plus **36** non-test
`insertWorkItem` call sites nobody has audited. For any of those, a repeat request can still be
destroyed silently exactly as before.

```bash
./scripts/audit-undeclared-recurrence.sh          # names them; 14 of 194 agents as of 2026-08-24
./scripts/audit-undeclared-recurrence.sh --json
```

**The finding is a MISSING declaration, never a wrong guess** — an explicit `false` is clean.
`claims-auditor.request_claims_review` genuinely needs the counter (its revalidator-close loop
writes `complete` into the two-strike window by design), which is why the migration was scoped to
five and not thirteen. The other 14 belong to their own lanes; the census exists so they can be
told by name rather than guessed at.

**End state, if you want to close the class:** drive the census to zero, then make the declaration
*required* for keyed steps so the unclassified state stops being representable. Census now,
ratchet later — the RFC_022 sequence.

---

## `bugs_open/345`: adopted, with an unresolved ownership collision

**Committed** as `d14eae8ab` — a six-file change adding an 8th parameter (`stepConfig`) to
`applyWorkItemFailureLadder` for `bugs_open/345` candidate 2's repeat-failure opt-in. Verified
green against clean HEAD with `scripts/verify-head-builds.sh --with … --test`. **Inert:** `nil`
means "not opted in", and **0** live definitions name `stop_on_repeat_failure`.

**Three things a fresh session must know:**

1. **My commit message on `d14eae8ab` is WRONG and corrected in `8149aecb3`.** I wrote "the change
   was never incomplete; the list was". It was backwards. The original author left two *sibling*
   test files' positional sqlmock expectations at five; the `bugs_open/345` lane fixed them at
   13:58 today. Full correction:
   `docs/agent_docs/docs024_key_docs_latest/bugfix_345_adoption/CONTRIB_2026-08-24_correcting_my_own_345_commit_message.md`
2. **The defect worth keeping is theirs:** *a positional sqlmock declaration is an arity contract
   with no compiler behind it.* A new bind parameter compiles everywhere and fails only in sibling
   files that pin positions.
3. **OWNERSHIP COLLISION, unresolved and for the owner.** The `bugs_open/345` session reports
   being instructed by their user to adopt the same change about an hour earlier, and had already
   fixed, verified and **dispatched a council submission**
   (`f1f1fc37-35e9-45fd-88d7-fcc3ddcf9eb0`, cited on `8149aecb3` so the round is not wasted).
   Given the shared account, parallel instruction is the likely explanation. I committed before
   their hold request arrived; forward-only means I did not uncommit. **If that verdict returns
   REVISE/REJECTED it lands on committed code, and they are the submitter — take direction from
   them on content.**

---

## Commands that were hard to get right

Full list in `RUNBOOK_326_retry_the_front_door.md`. The four that cost most:

- **`scripts/verify-head-builds.sh [--with <file>] [--test]`** — never hand-roll
  `git archive HEAD | tar`. I did it eleven times and left **5.0 GB** of extracts on a disk at
  75%. Reap with `scripts/scratch-report.py --reap`.
- **Dating a re-submission after `orchestration_states` is reaped:** it keeps **exactly 24h,
  sliding**. Use `site_specs` (`aspect='submission'`, written *before* the deduping step).
- **Telling a real dedup from a brake suppression:** do NOT ask whether an open row holds the key
  *now* — that is a present-tense predicate about a past event and returns 0 either way. Query in
  `LANDMINES.md`, "…`deduped: true` does NOT mean an open item holds the key".
- **`097` prints its `SAVE: SUBMISSION_CORR=` receipt BEFORE it publishes**, and the pod name has
  one-second resolution — two sessions in the same second collide and nothing is published after
  a convincing summary. Read the trigger's LAST line, and confirm with a query on
  `fix_correlation_id`.

## What went wrong on my side, so you don't repeat it

Six recorded errors, all in `WRONG_CALLS.md` under 2026-08-23/24. Four shapes, and the shapes are
the useful part:

1. **Measurements that could not have come out otherwise** — a present-tense query about a past
   event; a one-level-deep step census that missed 8 of 19 nested steps; a `grep` with
   `2>/dev/null` from a `cd` that outlived its command.
2. **A correct measurement with a purpose invented afterwards** — I described a control as ruling
   out something the arithmetic already excluded.
3. **A correct measurement generalised past its population** — a `0/4` written up as "never", then
   hardened into "a defect". It was refuted 30 minutes later.
4. **A conclusion that inverted a correct correction** — the `345` commit message above.

**The rule that would have caught three of the four**, adopted from the `loanzy.uk` lane:
**if the evidence is a count, the claim must contain the count.** "4 of 5 draws contained the
refused host" needs no hedge and survives the next observation intact.

**And the honest note about how they were caught:** mostly not by me and not by a check. Two peer
sessions caught them, and two were only exposed because the system kept running and contradicted
me. That is luck standing in for method.
