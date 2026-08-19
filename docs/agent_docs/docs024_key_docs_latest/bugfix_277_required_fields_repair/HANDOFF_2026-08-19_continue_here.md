# HANDOFF — 2026-08-19, fresh chat starts here: the lane is in its WAITING phase (three dated checks, then it closes), and RFC_030 is now the real work

**Supersedes `HANDOFF_2026-08-18b_continue_here.md`.** That file is still worth reading for the
`work-item-archiver` story and the `465`/`466` detail, but everything you need to ACT on is here.
Measured 2026-08-19 09:00–09:30Z. Read this from disk, then `NOTES_required_fields_repair.md` from
the bottom.

**One line: `083` and `277` are finished work waiting out their proving clocks (08-22 and ~08-24).
Nothing in them needs building. The largest real job on this lane is now `RFC_030`'s design round.**

## 0. Build state — verified at the BINARY, with controls

`agent-chassis` **`d3590ca4638d49bb6a3874db681814c4b0a99bbe`**, **158 commits** on from yesterday's
`0b185bad2` (HEAD 3 further on). `kafka-scheduler` reports the same sha from its own log line.

⚠ The startup `build provenance` line was **absent from `--tail=20000` on BOTH chassis pods** — that
is the documented "scrolled" case for a busy service, **not "unstamped"**. Probed `/proc/1/exe`
instead, always with a control in the same breath: sha **PRESENT**, current HEAD `db6ae7254`
**ABSENT**, yesterday's `0b185bad2` **ABSENT**. So a real build, not a same-tag cached no-op.

This roll carries **`480`'s Go half** and **`184` part 2** (the `literal_markdown` detector widened
to `md_link`). `465`/`466`/`471`/`472` are **config-only** and were already live — they need no roll.

## 1. What changed under this lane overnight, by OTHER sessions — read before you act

**Two owner decisions landed on 2026-08-18 evening, and one of them builds the thing this lane had
deliberately left open. Do not rebuild it.**

- **`480_owned_page_refusal_is_not_a_handler_failure.sql`** (applied 08-18 20:24; Go half
  `save_page_sections_action.go` + `v3_site_actions.go`, live in this roll). Owner decision №1:
  *"do not switch the handler off for this — write something other than `failed`."* An owned-page
  refusal now records **`wont_fix`**, which is in **neither** bucket of the promoter's success floor,
  so a protective refusal no longer votes in a competence measure it says nothing about. Shipped as
  an **opt-in field, unsafe default OFF** (`owned_page_refusal_status`) per the 2026-08-02 §2 ruling.
  Its author reached this from a different pair than I did (`phantom_internal_link`: 69% on generic
  pages, 0% on owned, blended 47%, ~134 findings queued behind it). **Two lanes converging from
  different evidence is why this was real.**
  - ⚠ **FORWARD-ONLY — it backfills nothing.** `literal_markdown → page-build-handler` re-measured
    today is **unchanged: 3 successes / 36 failures, 16 protective / 16 genuine, 8% raw, 16%
    corrected.** The ratchet stops tightening; it is not undone. **So `471`/`472`'s remedy text is
    still needed on 08-21.**
  - ⚠ **UNPROVEN, and the zero is readable only because of the control.** Zero `wont_fix` owned
    refusals since it applied — **and zero owned refusals written `failed` since the roll either.**
    That second half is the demand control: this is *"not yet exercised"*, not *"working"*.
    **OWED: grade `480` at the first real owned refusal.**
- **`479_escalation_reclaims_a_pair_that_has_since_qualified.sql`** (owner decision №2, *"fix the
  door"*). Closes `453`'s one-way door: an escalated row whose pair later qualifies rejoins the
  automated path. **My `471`/`472` text survived it** — its author used surgical replacement on three
  verbatim anchors, each asserted to occur exactly once, guarded on the whole body's md5, explicitly
  because *"that lane is iterating fast (465, 466, 471, 472 in one day)"*. Verified after the fact.
  **That is how to edit a live object another session is working on — copy it.**

## 2. THE THREE DATED CHECKS — all automatic, nothing to build

`held-pair-canary-escalation` is daily and fires at **12:57 UTC**. As of 09:23Z today it had last
run **08-18 12:57:48**, so none of these has happened yet. Instrument: the `pre_query_result` line in
`kafka-scheduler` logs (`grep detected-item-promoter` / the escalation task).

| when | expect | what it proves |
|---|---|---|
| **08-19 12:57** | **`escalated=0`, `watching=15`** | ⚠ **ZERO IS CORRECT — do not read it as a failed migration.** It is `466`(a) working (a `HAVING` that still speaks on an idle tick), and it grades the 08-18 date correction |
| **08-20 12:57** | `placeholder_contact → page-build-handler` (3 rows) escalates, **canary** wording | first real escalation this mechanism has ever produced |
| **08-21 12:57** | `literal_markdown` (10, **floor**), `dead_fragment_link` (1), `missing_conversion_path` (1) | first use of `471`/`472`'s corrected floor remedy — the one that says PARTITION THE FAILURES rather than "fix the handler" |

⚠ **Why 08-19 escalates nothing** (this bit me and is now a LANDMINE): a daily task with a 3-day
predicate delivers **3–4 days**, because it can only act on its own tick. `placeholder_contact`'s
oldest row is 08-16 **19:17**; at the 08-19 **12:57** tick it is 6h20m short of three days.
Conditional on the held set not changing — the clock keys on `min(created_at)` per PAIR, so if the
oldest row leaves `detected` the date moves **out**.

⚠ **Do not canary `missing_conversion_path → content-gap-planner`** — `bugs_open/255` owns it
(diagnosis CONFIRMED first iteration: routed at a handler that cannot read its spec). A canary would
record a documented routing defect as handler incompetence.

## 3. Closure — both bugs are finished work waiting out a clock

Neither needs building. Conditions are the bug files' own, not my judgement.

- **`277` → `bugs_closed/` ~2026-08-22.** Three conditions in its "Still open before this moves"
  section: (1) the churn guard's remaining days; (2) the two cancelled conversions re-raising — no
  `cancelled` rows of the type remain, so this depends on discovery rotation re-filing them, and
  **if not seen by ~08-22, re-file by hand**; (3) a named seam to watch, not to fix — `033`'s
  revalidator and this router both write `result`, and `mark_complete` REPLACES `result`. They
  compose correctly today; nothing guarantees it.
- **`083` → `bugs_closed/` ~2026-08-24.** Its own §5 says the fix is complete and proven (all three
  criteria met; the residual closed 08-18 and verified at the running service) and recommends closing
  **once `444`/`458` have sat their week** — both applied 08-17, so 08-24. Everything else `083`
  surfaced now has an owner: `453`'s one-way door → **`479`**; the refusal accounting → **`480`**;
  `literal_markdown`'s handler → `bugs_open/184`/`201`; the unstable key → `bugs_open/300`.
- ⚠ **Moving either file: name BOTH paths on the commit** — `git commit bugs_open/OLD.md
  bugs_closed/NEW.md -m "..."` — and verify at HEAD, not the tree:
  `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep <number>` must return exactly
  one line. `git mv` + a pathspec commit silently ships a COPY. **LANDMINE.**
- Health, so you can see the lane is sound while it waits: **1,123 promotions 08-14…08-18, 946
  complete vs 79 failed (92%)**.

## 4. THE REAL WORK — `RFC_030`, the router engine's design round

`docs/agent_docs/docs024_key_docs_latest/router_engine/` (standing five) ·
`docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_030_single_type_work_item_routers_want_one_engine.md`
(**Status: RULED 2026-08-15 by the owner — scheduled as its own lane**).

**Phase 1 (measure the live population per type) is DONE** and written up in that lane's NOTES.
**Phase 2 is a council design round on shape A vs B, submitted as an RFC-shaped design — BEFORE
building.** This IS architecture scope (a new shared mechanism), so it is the council's proper case,
unlike this lane's config-only migrations.

✅ **The blocker on phase 2 is CLEARED as of 2026-08-19: the PLAN's guarantee 8 is fixed.** It said
RFC_022's accumulation counter was *"unbuilt"*, which was false when written — RFC_022 is **CLOSED**,
the counter shipped 2026-08-13 (`cmd/config-key-audit --optional-key-budget`, register **WFA-013**),
the owner **ruled N = 10** on 08-14, and a daily CronJob has enforced it since 08-14. That is not a
citation fix: it converts guarantee 8 from "volunteer a count nobody consumes" into **a live budget
with a ruled threshold**, so the design round must ask whether the chosen shape makes each routed
type accumulate optional keys on one shared action — i.e. whether the engine walks toward N = 10 as
it succeeds. **[UNVERIFIED] which of A or B has that property is exactly what the round must
establish — do not assume.** Full correction, with the hand-maintained-literal trap and the parity
test to run, is in the PLAN at guarantee 8.

**Then:** build the engine; migrate **`410` first** (its 8 routes define the contract and its
44-item history is the regression fixture — census and canary evidence live in *this* lane's
directory), then `397`'s two; retire the three bespoke seeds (rollback files exist for all three);
update CQ-023 and IMG-071; register the engine.

## 5. Also owed, smaller

- **Grade `480`** at the first real owned refusal (§1). It is live and unexercised.
- **Council gate's config blind spot** — `097` scopes on `platform/`/`internal/`/`pkg/`, so every
  migration in this lane since `444` (`465`, `466`, `471`, `472`) has been unsubmittable. Another
  lane filed `landmine(297)`. Submitting `465` with `FORCE=1` is worth it — it changed what
  "lifetime" means for a shared gate.
- ⚠ **Migration numbers collide on this tree and MUST NOT be renumbered.** `453`, `454`, `462` and
  now `471` each exist twice; `462`'s two halves are one applied, one pending. **A number tells you
  neither author nor applied-state — ask the ledger by exact filename.**

## 6. The landmine family this lane keeps hitting — now SIX, and the shape has never varied

Every one is **a population or a domain assumed rather than enumerated**, and none was caught by
review (twelve seats approved `444`):

1. `failed` rows carry **no `completed_at`** — pair health keyed on it returns a uniform 100%.
2. `status` has **two** terminal success states (`complete`, `verified`).
3. The row set is **not stable** — rows leave `site_work_items`.
4. The row set is only a **~7-day window** — `work-item-archiver`; the archive is *bigger* than the
   live table.
5. **A control that cannot come out otherwise** — THREE tautological ones caught here. The test is
   not *"is this control true?"* but ***"could it ever have come out non-zero?"***
6. **The CLOCK** (new, 2026-08-18) — a daily task's "3-day limit" is 3–4 days; predicting from a
   DATE is off by a full tick, and the miss shows up as a silent zero.

Also live: **a same-tag rebuild ships the cached image** · **an aggregate-only SELECT with a `WHERE`
returns one row regardless — use `HAVING`** · **a pathspec commit still takes a same-file passenger**
· **backticks in `git commit -m` EXECUTE — use single quotes or `-F`** · **`EXPLAIN` proves SQL
parses, it cannot prove a path inside a string exists** (that shipped `bugs_open/295` into a live
payload 30 minutes after 295 moved to `bugs_closed/`).

## 7. Session-start checklist

`git log --oneline -10` · re-read this file **from disk** · **verify the chassis/scheduler sha before
trusting any Go claim, with a negative control** · `scripts/who-owns.py 277` and, **by slug**,
`detected_findings_never_reach_a_handler` (083 is an ambiguous number) · re-measure §2's held set
before believing its dates · then §2 if a tick is due, otherwise **§4**.
