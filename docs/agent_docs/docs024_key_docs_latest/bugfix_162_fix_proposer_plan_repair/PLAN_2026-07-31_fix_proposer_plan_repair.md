# PLAN — bugs_open/162: opt `fix-proposer` into the bounded plan-repair loop

**Started:** 2026-07-31 ~22:30 BST. **Lane:** picked up from `bugs_open/162`, which the
`bugfix_099_plan_refusal_recoverable` lane filed and explicitly left unowned
("`fix-proposer` belongs to the fix-loop lane"). Its session (`31f37a59`) had been idle
since 19:25 when I picked this up, and no other active transcript referenced 162.

## What the ticket asks for

One thing: `fix-proposer`'s `persist_plan` step runs the shared action
`diagnose_persist_fix_plan`. A structural validation failure there returns a bare error,
fails the step, routes to `complete_refused`, and **discards a completed, good plan** —
with `orchestration_states.error` NULL, so the run reads as clean on any dashboard keyed
on that column.

`bugs_open/099` candidate 2 already built the general fix: a bounded repair loop, live in
the chassis since v1.0.1215. It is **opt-in by design** — with no `repair_step` in the
step config the action behaves exactly as before. That opt-in is what let the fix ship
without changing the other consumers' guarantee, and it is also why `fix-proposer` was
still exposed.

So there was **no framework-vs-individual choice left to make in the fix itself**: 099
already made it, correctly and generally. The whole of 162 is "turn it on for the second
consumer", and the migration was already written, guarded and dry-run clean
(`docs/agent_docs/sql_for_agents/273_fix_proposer_plan_repair_loop.sql`).

## Decisions, and their reasons

1. **Apply 273 as written; do not rewrite it.** It was authored by the lane that built
   the mechanism, sized to this agent (`max_tokens` 8000 matching its own `propose`, not
   feature-designer's 32000), guarded against double application and against a changed
   step graph, and verified by a closing `DO` block. Re-deriving it would have been
   reuse-hostile and would have lost the sizing reasoning.

2. **Verify before applying, not after.** Confirmed the Go half is live on **both**
   chassis replicas by pod-grep with a positive AND a negative control in the same exec
   (`MEMORY.md`: a roll is not evidence your fix shipped). Confirmed no open
   `site_work_items` touch `fix-proposer`. Re-ran the dry run against the live row rather
   than trusting the file's "dry-run clean" claim from 2026-07-30.

3. **Verify the routing against the PARSER, not against the config's own prose.** The new
   router step is `conditional_branch` with condition `plan_persisted.plan_valid == true`.
   A config key that resolves to nothing looks exactly like a live one, so I read
   `conditional_branch_action.go` and confirmed: `resolveFieldValue` → `FindByPath` handles
   the dotted path from collected_data root, and `compareValues(nil, "true")` returns
   false — so an unresolvable field routes to `repair_plan`, never onward to a council.

4. **Then check the same thing at the DATA, because the parser check is not sufficient.**
   The real hazard was the opposite direction: if a step's result were *wrapped* (the way
   `execute_llm_prompt` with `output_format=json` leaves its object under
   `<output_field>.result`), the condition would never resolve and **every valid plan
   would route to repair** — an unbounded loop, because the repair counter increments on
   refusal artefacts and a misrouted valid plan writes none. Measured on live rows:
   `plan_persisted` is stored **unwrapped** at collected_data root, keys
   `[edit_count, files, persisted, plan_json, plan_valid, summary]`, `plan_valid=true`.
   Hazard excluded empirically, not by reading.

5. **Do NOT induce a refusal on `fix-proposer`.** See RUNBOOK for the method (arm
   `persist_plan.config.max_edits` to 1). I declined it deliberately: at the time of the
   change **three other lanes' `fix-proposer` runs were in flight**, and the arm is a
   fleet-wide edit to a shared live agent with a ~30-minute dispatch window. Tripping
   other lanes' runs to prove my own wiring is not a cost I get to spend on their behalf.
   What this leaves unproven is recorded honestly in NOTES and in the bug file — it is a
   gap, not a pass.

6. **The platform-code residual goes to the council, not into this commit unilaterally.**
   While verifying, I found that the shared action's `planRefusalErrorCode` comment states
   an invariant that is false for four of its five terminal exits. Fixing the two exits
   where it is safe means changing behaviour that four tests explicitly guard with the
   words "must not touch the DB", written days ago by the lane that owns the action. That
   is a scope judgement, not a cleanup — so it was submitted to the council gate
   (`SUBMISSION_CORR 7b1eb170-50b9-4c2f-b5d5-25fd6cf88c2b`) with the question posed
   explicitly in the rationale.

## Scope boundary — what this lane deliberately does NOT do

- **Does not opt `council-gate` in.** 162 records that as a positive choice: it is the
  gate that reviews changes to this very mechanism, and coupling it to the change it
  reviews buys nothing. It would also need `summary` added to the refusal result, because
  its prompt renders `plan_persisted.summary`.
- **Does not touch the not-opted-in code path** (`repair_step == ""` /
  `max_repair_attempts < 1`). That is another lane's documented deliberate opt-out.
- **Does not flip the opt-in to a default.** Per the owner ruling of 2026-07-29, changing
  what a shared mechanism GUARANTEES for consumers who did not ask is architecture-scope
  and needs an RFC, not a bug-fix commit.
- **Does not close the size-cap class** (`max_stages`, `max_total_edits`), whose refusal
  messages ask for less scope while the repair prompt forbids dropping scope. Inherited
  limitation, documented in 162 and in concept register `FIX-057`.
