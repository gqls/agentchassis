# PLAN 2026-09-03 — bug 161's residual: the register that protects nothing

**Lane:** resumed `bugs_closed/161`. **Product of this lane:** `bugs_open/456`, commits
`3f221f99f` / `e5b41dc31` / `fef2ced9c`, register `CLM-031` + `CLM-032`, RFC_025 §12.

## The brief, and how it changed under me

Asked to check whether `bugs_open/161` was still valid, resume it if no thread was active, and
prefer a framework-wide fix to an individual one.

**161 turned out to be closed, fixed and still working** — re-verified at the artefact, not
carried forward. So the question became: what did its close-out leave behind? The answer was
not the residual the close-out named.

> **CORRECTION to my own first framing, made mid-session and worth keeping.** I began planning
> around the 27 artifact-sourced facts the sweep cannot count — a real gap, and the one 161
> predicted. That would have been a nudge and a counter: worth doing, low stakes. Running the
> real parser over all 27 live registers to size that population is what turned up the actual
> defect: **two registers were not parsing at all**, so two sites' entire claims layer,
> `banned_claims` included, had been off for over a week. **The measurement I ran to scope a
> small finding is what found the large one.** I would not have run it if I had trusted the
> close-out's census instead of re-taking it.

## Decisions, and the reasons

1. **Fix the parse, not the two rows.** Repairing finetuning.uk's and noted.co.uk's facts would
   have made the symptom go away and left the mechanism live for the next author. The framework
   fix is that **a malformed fact costs that fact** — bans, attestations and citation codes
   decode regardless. Ordered by what makes the bad state unrepresentable, as CLAUDE.md asks.
2. **Do NOT give `Value` a text shape in this change**, though that is the honest end state.
   It is a shared-vocabulary addition and belongs in its own round; with the parse fixed a text
   fact is skipped and reported rather than catastrophic, so the pressure is off. Named as an
   open option in RFC_025 §12.5 rather than left as a gap. The architecture seat independently
   set the trigger for promoting it: **if the item type's volume is still growing when next
   measured, that is the RFC.**
3. **Check the arming hazard BEFORE proposing, not after.** Making these registers parse arms
   10 banned claims at once, and 161's own landmine says a banned claim is BLOCKER severity —
   repair copy first, then arm. Measured: both sites' deployed text scores zero violations. The
   limit is stated everywhere it is quoted (stored `rendered_html`, tags stripped, so `<title>`
   and JSON-LD are outside it).
4. **Two commits, not one, because of a peer.** `refresh_evidence_base_action.go` carried
   another session's uncommitted `livespec` call site. Committing it would have taken their
   half-written work as a passenger and **broken HEAD for the fleet**, since `make build-*`
   builds from committed HEAD. Held the sweep half until their helper landed (`1802359a6`).
5. **Contribute into `bugs_open/288`, do not compete.** `register_guards_code_phase_b` owns that
   bug, RFC_025 stage 2 and this function. Its planned next step (Phase 3b) is untouched.

## What is deliberately NOT in scope

Bulk-retyping the 27 unverifiable facts (ratified as per-site and human-paced); repairing the
two malformed rows (their lanes'); generalising the tolerant decode to other `site_specs`
aspects (open, and explicitly not cleared — see `456` §9); the `needs_human_review` queue depth
(1,389 rows, `bugs_open/033`).

## Verification design

Nothing here is provable by unit test alone, because the failure is *a check that does not run*.
The controls, in order of strength: a **before/after run of the real parser on the live
register** with one fact repaired and nothing else changed (could have come out otherwise);
**mutation tests** run red, not merely written; and post-roll, a **census demand control** — 25
registers unchanged, 2 must flip — plus the work items actually appearing. `456` §8 holds the
commands.
