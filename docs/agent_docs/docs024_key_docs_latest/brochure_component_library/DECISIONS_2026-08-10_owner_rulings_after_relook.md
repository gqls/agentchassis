# Owner rulings 2026-08-10 — the four §6 decisions, after the re-look

Companion to `HANDOFF_2026-08-10_fact_assignment_front_continue_here.md` §6.
That section put four decisions to the owner. Before ruling, the owner asked for
a fresh-eyes pass over the advice; the pass **changed or sharpened three of the
four recommendations**, and the owner then ruled on all four (2026-08-10, in
session). This file records what the re-look corrected, each ruling, and the
work each ruling creates. The re-look corrections are recorded here *because*
they were corrections — the first walkthrough's versions were wrong in ways the
next reader should not inherit.

## 1. What the re-look corrected (evidence inline)

1. **Decision 1's proposed condition was already satisfied.** The first
   walkthrough recommended ratifying the 215 lossy merge "on condition the Warn
   log carries enough to reconstruct what was dropped". Reading the shipped
   code (`14b1cff28`, `dedupePlanPageRows`): the Warn already logs both raw
   names and both full section lists. **The real defect is one level down: a
   chassis Warn is not a record.** An active chassis pod retains **under one
   second** of log (measured 2026-08-08, bugs_open/136 §11: oldest retrievable
   line 0.4 s old; a warn line proven unreachable ~2 minutes after firing on a
   Running pod). "Logs the other at Warn" is therefore, in practice, "records
   it nowhere", and the lossy branch's firing rate is unmeasurable.
2. **Decision 3's "no new damage" framing was wrong in one direction.**
   The first walkthrough called the prose escape "permissive — worst case is
   today's behaviour". Reading the mechanism
   (`v3_site_actions.go:3105`, `recomposePagesFromSpec` `:5844`):
   `recompose_pages` works by *releasing* a page from the preserve guard so the
   planner's freely-proposed composition governs it — but the release happens
   in **validate**, while seed 362 instructs the **planner**, which still sees
   the page's realised sections and will re-emit them verbatim. Release a page
   that is re-emitted verbatim and the redesign **silently no-ops**: no error,
   page unchanged, operator none the wiser. That is a new failure mode 362
   introduces, and the seed's own header concedes the cause ("the planner is
   not told which pages those are"). Dropping the escape would be worse still —
   it would leave `recompose_pages` released-but-never-deviated-from, i.e.
   completely dead.
3. **Decision 2's "aspirational clause" concern was overstated, and the real
   contingency is sharper.** Branch B of the v4 prompt only renders once
   scoping is live, so "other sections on this site carry them" is true when
   served — *except* in exactly the §3.5 hole (a dropped/malformed `facts` key
   is indistinguishable from a legitimately empty assignment). The error
   direction there is omission, never fabrication, which is what lets the
   compliance read certify now rather than wait (see §2, Decision 2).

## 2. The rulings

### Decision 3 — seed 362's redesign escape (ruled: ship with the prose escape, plus detection, plus the registered follow-up)
**Ruling:** option 1 as re-presented — 362 ships with the prose escape
unchanged; a **detection line** is added so the new silent no-op becomes
visible (in validate: when a page named in `recompose_pages` comes back with a
composition identical to its realised one, record it durably — this merges with
the council's existing "loud-signal on a recompose drop" follow-up already
recorded on `features_open/012`); and the **permanent fix is registered as a
follow-up** on `features_open/012`: surface `recompose_pages` to the planner as
a field, the RFC_010 §2 shape ("a comment is not a control on a tree this many
sessions share"). Until the field exists, a `recompose_pages` caller must also
state the redesign in the briefing the planner sees — recorded on 012, and to
be added to `LANDMINES.md` **when 362 applies** (the trap is not live before
then).

### Decision 2 — the compliance read of writer prompt v4 (ruled: owner delegated the read to the session, acting as a lawyer)
**Ruling:** rather than performing the read personally now, the owner directed
the session to perform it "acting as a lawyer". Done — the full opinion is
`COMPLIANCE_READ_2026-08-10_writer_prompt_v4.md` (this directory), with the
operative branch text quoted in full and an **owner countersign line**, since
the compliance seat's ask was for a *human* read and seed 330's header makes
that an apply precondition. **Verdict in one line: the three-way branch — the
whole of 330's delta — is certified; no overclaimed-reliability phrasing; four
pre-existing findings recorded as follow-ups, none introduced by 330, none
blocking.** The one quality condition: the §3.5 `FACT_ASSIGNMENT_ABSENT` fix
(already work-queue item 2) is what makes Branch B's "other sections carry
them" reliably true; it should land in the same round.

### Decision 4 — apply-order enforcement for 362→328→330 (ruled: no tooling; self-guarding seeds)
**Ruling:** no new tooling. Every partial-apply combination was re-derived and
is inert (362 alone is beneficial; 328 alone feeds sections nothing reads; 330
alone falls through to today's site-wide branch), so the hazard fails safe.
Instead: seeds 328 and 330 each gain a `DO`/`RAISE` precondition block
asserting their predecessor applied (a verify block of bare `SELECT`s cannot
stop a `COMMIT` — the estate's own landmine). **Sequencing constraint: do not
author the guards until the §3.4 census settles which agent's `plan_sections`
serves the writer path** — its answer may re-target 328, and a guard written
now would encode a precondition that might be wrong.

### Decision 1 — the 215 lossy-merge policy (ruled: ratified, with a durable-record condition)
**Ruling:** richer-wins stands. The code read supported it: richer-wins
discards strictly less than keep-first in every case; an earlier keep-first
draft was caught by its own test
(`TestDedupePlanPageRows_RicherComposedWinsEvenWhenSecond`); backfill fills
blanks only; the observed collision shape (composed + stub) takes the lossless
Info branch. **Condition:** the lossy branch's Warn is upgraded to a durable
`agent_error_log` record (the `FACT_CARRY_UNMATCHED_SECTION` pattern), because
chassis log rotation is sub-second and the branch is otherwise unobservable.
That converts the standing policy question into a measurable one: **richer-
wins, durably recorded, revisit if the count is ever non-zero.**

## 3. Work these rulings create (joins the front's queue)

In addition to the handoff §5 queue (unchanged: §3.4 query first, then the
§3.5 fix, then resubmit with `RESUBMIT_CORR=a06ff850-aff6-4ed0-8e0a-93d57b0cbc45`,
submit BEFORE committing, `Council-Submitted:` trailer):

1. **Detection line** for the recompose verbatim no-op (Decision 3) — small Go
   change in validate, same file/front as the §3.5 fix; natural to build and
   mutation-test alongside it and carry in the same resubmission.
2. **`agent_error_log` record** for the 215 lossy branch (Decision 1) — small
   Go change in `dedupePlanPageRows`'s Warn arm; belongs to the 215 lane's
   file, coordinate rather than compete (`scripts/who-owns.py 215` before
   touching).
3. **`DO`/`RAISE` self-guards** in seeds 328/330 (Decision 4) — authored only
   after the §3.4 answer lands.
4. **LANDMINES entry** for the recompose no-op — written when 362 applies, not
   before.
5. **Owner countersign** on the compliance read — before 330 applies (its
   header's precondition).
6. Follow-up seeds from the compliance read's pre-existing findings (§3.4–3.7
   there: invented-commitments clause; edit-mode legacy-claims wording;
   testimonial trade-dress check) — recorded, unscheduled, not this round's.
