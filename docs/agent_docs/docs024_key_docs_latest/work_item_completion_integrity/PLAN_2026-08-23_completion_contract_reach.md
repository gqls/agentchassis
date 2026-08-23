# PLAN 2026-08-23 — the completion contract's REACH (the follow-through from `bugs_closed/344`)

**Lane revived 2026-08-23** after dormancy since 2026-08-08. Same subject as ever: an item's
recorded status must mean what it says.

## What prompted this

A session was asked to fix `bugs_open/344`. **There is no such file** — 344 was closed 2026-08-21
(`06a9d9c41`), fixed and live. Re-verifying rather than trusting it confirmed the closure holds
(details and the demand control in `bugs_closed/344` §5(a)). What was left open is a different
question: **344 established a contract and shipped it to two writers; about eleven can reach a
`triaged` row.**

The contract: *a work item whose retry the failure ladder has already scheduled must not be
stamped `complete`.*

## The design decisions, and their reasons

1. **Go convergence, not a database trigger.** A `BEFORE UPDATE` trigger on `site_work_items`
   would bind every writer at once and `RETURN NULL` would reproduce the `rows==0` outcome callers
   already handle (migration `216` is the precedent on this very table). **Rejected** on RFC_043
   Q3, owner 2026-08-21: *"Shared contracts live in GO."* A trigger gives the contract a second
   home. The counter-argument — that it is a relocation, not a mirror, and would let the SQL copy
   in `claimed-item-timeout` retire — is genuine and **architecture-scope**; it needs an RFC, not a
   bug patch. Recorded in the bug file so it is not re-derived from scratch.
2. **Order by what closes the door**, per the standing rule. That put the ladder's stamp fidelity
   first (it is the substrate every other guard reads), then the writers, then the vocabulary.
3. **Human dispositions are a stated exemption, not an omission.** The admin handlers complete
   `WHERE id = $1` with no status guard. A human overriding a machine-scheduled retry is the point
   of a disposition; silently refusing it is a different design question. Named, not fixed.

## What shipped, and what is blocked

**Shipped** — `markOriginalComplete` (`apply_gap_plan_action.go`), commit `2dd05c5b2`, council corr
`af5135d6-8ca2-4453-b33e-a299dcd6a622`. Chosen as the first increment because it is the **only**
`complete` writer that names `triaged` in its own WHERE, so it selects for exactly what the ladder
writes. Fixed both halves: the missing predicate, and the discarded error+rowcount that made any
refusal invisible. Four mutation proofs, byte-identical restore, green on clean HEAD.

**Blocked, and this is the resizing** — the other three gaps live in `v3_site_actions.go`,
`load_work_item_actions.go` and `work_item_failure_ladder.go`, all carrying other sessions' large,
currently-RED uncommitted work (one touched **one minute** before the check). Editing them takes
half-finished work as a same-file passenger, which a pathspec commit cannot prevent. **Recorded in
`bugs_closed/344` §5(e) with the constraint stated, so the next session can re-check and proceed.**

## Corrections made in flight (kept, not edited away)

- **`WORK_ITEM_BURST_COOLDOWN_MINUTES=0` is NOT a live disarm path** — `envInt` requires `n > 0`
  and falls back to the default. I listed it as one first. The statement-level hazard survives for a
  caller passing 0 directly; the env route is closed.
- **The `completion_skipped` 22/0 split is not evidence of demand** — the marker self-erases on the
  next success, and it erases the very reason being counted. Both in `WRONG_CALLS.md`.

## Where this goes next

1. Unblock and finish gaps (b)/(c)/(d) when those files are quiet — the bug file names each.
2. RFC_043 Q2 (converge the three guard lists) is **still unowned**; the instrument for its
   disconfirming question is now written up in the RFC and the bug file.
3. Gap D's fix — one append-only `agent_error_log` line — is what makes any of this measurable.
