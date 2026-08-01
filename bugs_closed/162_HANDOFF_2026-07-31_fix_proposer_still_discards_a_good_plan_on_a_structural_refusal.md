# 162 — fix-proposer still discards a completed plan on a structural refusal, and the fix is written and waiting

**Filed:** 2026-07-31, by the `bugfix_099_plan_refusal_recoverable` lane, **at the
council gate's request**. Its `bug_historian` seat objected that leaving the sibling
consumer on the old behaviour with only a doc pointer is *"exactly the kind of artifact
that reads as 'handled' in a diff without being handled in production"*, and asked for
a trackable item rather than a comment in another bug's file. It is right.

**Severity:** medium. Not an outage. It silently destroys a completed, good plan and
spends the full proposer cost to do it. Identical in kind to `bugs_open/099`, which is
where the mechanism is documented.

**Status:** OPEN, **unowned by this lane on purpose** — `fix-proposer` belongs to the
fix-loop lane. **The fix is written, dry-run clean, and one command away.**

## The defect

`fix-proposer`'s `persist_plan` step runs the shared action
`diagnose_persist_fix_plan`. When a plan fails structural validation that action
returns a bare error, which fails the step, routes to `complete_refused`, and
**discards the plan** — with `orchestration_states.error` NULL, so the run reads as
clean on any dashboard keyed on that column. The reason lives only in
`collected_data->>'__step_error'`.

This is `bugs_open/099` exactly, one consumer over.

## Why it is still open when 099's fix is live

`099` candidate 2 made the refusal recoverable, and it is **live** (chassis
`v1.0.1215`, verified by pod-grep on both replicas). But the mechanism is **opt-in by
design**: with no `repair_step` in the step config the action behaves exactly as it
did before. That opt-in is what let the fix ship without touching the other two
consumers' guarantee — and it is also why `fix-proposer` is still exposed.

Only `feature-designer` is opted in (migration `272`, applied 2026-07-31).

## The fix, already written

`docs/agent_docs/sql_for_agents/273_fix_proposer_plan_repair_loop.sql`

- **Dry-run clean** against the live row (piped with `COMMIT` → `ROLLBACK`: guards
  passed, three `UPDATE 1`s, verification block passed, rolled back). So it is proven
  applicable, not merely drafted.
- Guarded (refuses a second application; refuses if `persist_plan.next_step` is no
  longer `select_panel`), snapshotted, and verified by a closing `DO` block.
- Sized to **this** agent: `max_tokens` 8000 matching its own `propose` step, not
  feature-designer's 32000 — a fix plan is one ≤8-edit plan, a feature plan is up to
  six staged ones. Prompt uses `{{.diagnosis_row.conclusion}}` because this agent has
  no `spec_row`.

**To apply:** confirm the chassis carries the Go half (`grep -ac plan_validation_refusal`
on the pod, with a control in the same exec), then run the file. Rollback instructions
are in its header — and note the `snapshot_agent` trap recorded in `LANDMINES.md`:
the 2-arg form writes to `agent_definitions_backup`, and a restore must order by
`snapshot_taken_at`, never `created_at`.

## Known limitation, inherited from 099

The repair loop closes the **structural** class of validator rule (duplicate file in a
stage, modify-before-add, forward `depends_on`, per-stage edit cap, …). It does **not**
close the **size caps** (`max_stages`, `max_total_edits`), whose messages ask for less
scope while the repair prompt forbids dropping scope — such a refusal burns its one
repair round and goes terminal, exactly where the platform lands today. See
`bugs_open/099`'s correction of 2026-07-31 and concept register `FIX-057`.

## Also still exposed, and deliberately

`council-gate → persist_submission` uses the same action and is likewise not opted in.
That one is a positive choice rather than an omission: it is the gate that reviews
changes to this very mechanism, and coupling it to the change it reviews buys nothing.
Opting it in would additionally need `summary` added to the refusal result, because
its prompt renders `plan_persisted.summary` (measured 2026-07-31) and the refusal path
does not return it.

## References

- `bugs_open/099_HANDOFF_2026-07-26_feature_designer_plans_die_on_a_rule_it_is_never_told.md`
  — the mechanism, the live evidence, and the corrections.
- `docs024_key_docs_latest/bugfix_099_plan_refusal_recoverable/` — PLAN, RUNBOOK
  (including how to induce a refusal and the four conditions that make the test
  discriminating), NOTES, README, SUMMARY.
- Concept register `FIX-057`.
- Council correlation `f4a4628f-3b90-4054-a875-f2cf72b83e72` — the round that asked
  for this file.

---

# CLOSED 2026-07-31 — 273 applied, verified at the live row, `bugs_open` → `bugs_closed`

Closed by the `bugfix_162_fix_proposer_plan_repair` lane. Working docs (the standing
five, including the missteps):
`docs/agent_docs/docs024_key_docs_latest/bugfix_162_fix_proposer_plan_repair/`.

## What was done

`sql_for_agents/273_fix_proposer_plan_repair_loop.sql` applied at ~22:40 BST, as written
— not re-derived. Config, so **live immediately**; no roll needed.

**Why this meets the "fixed AND live" bar.** The defect's precondition is literally
`repairStep == ""` (`diagnose_persist_fix_plan_action.go`, the early return in
`planValidationRefusal`). That is now **false** at the live row, so the discard path is
unreachable for this agent. Verified at the artefact, not at the migration's own
verification block:

- `persist_plan` → `check_plan_valid`, `repair_step=repair_plan`, `max_repair_attempts=1`
- `check_plan_valid` → then `select_panel` / else `repair_plan`
- `repair_plan` → `persist_plan`, `output_field=proposal` (the same field `propose`
  writes, so `plan_field: proposal.result` reads the repaired plan)
- re-running the migration now raises `273: already applied`
- Go half live on **both** replicas: `plan_validation_refusal` ×2, positive control ×11,
  negative control 0, all in the same exec

## The check that was worth more than the fix

The router reads `plan_persisted.plan_valid`. Had step results been *wrapped* — the way
`execute_llm_prompt` leaves its object under `<output_field>.result` — the field would
never resolve, `compareValues(nil,"true")` is false, and **every valid plan would route
to repair**. That loop is *not* bounded: the repair counter counts refusal artefacts, and
a misrouted valid plan writes none, so it would spin to `fuel_budget`. Excluded by
measurement, not by reading — live rows carry `plan_persisted` **unwrapped** at
collected_data root, keys `[edit_count, files, persisted, plan_json, plan_valid, summary]`,
`plan_valid=true`.

## What is NOT proven, deliberately

**[UNVERIFIED] the route `persist_plan → check_plan_valid → repair_plan → persist_plan` has
not been induced on `fix-proposer` itself.** The method is in this lane's RUNBOOK §5 (arm
`persist_plan.config.max_edits` to 1). It was declined because the arm is a fleet-wide
edit to a shared live agent with a ~30-minute dispatch window, and **three other lanes'
`fix-proposer` runs were in flight** at the time. It is proven on `feature-designer`,
which runs the identical shared Go code and the same-shaped graph (2026-07-31: 3 refusals,
2 routed to repair, 1 exhausted terminal). Run it in a quiet window to close the gap.

**[UNMEASURED, AND UNMEASURABLE IN RETROSPECT] how often this bit.** `orchestration_states`
retention is ~1 day, and more fundamentally the defect **destroys its own evidence**: a
non-opted-in refusal writes no artefact, no `agent_error_log` row, and leaves
`orchestration_states.error` NULL. There is no population to count. Any figure quoted for
this bug's historical impact is fabricated.

## Residual shipped alongside — COMMITTED, NOT LIVE

Found while verifying, reviewed as its own question, and **not part of this bug's defect**:
`planValidationRefusal` has five terminal exits and only one wrote the operator-facing
`agent_error_log` row, while the comment on `planRefusalErrorCode` told the reader that
row's absence was dispositive (3 rows fleet-wide, all `feature-designer`, all from that one
exit). Commit `417d6fd87` records the refusal on the two **bookkeeping-failure** exits and
corrects the comment. Council **APPROVED**, 9 reviewers, 3 advisory objections, all
discharged (corr `7b1eb170-50b9-4c2f-b5d5-25fd6cf88c2b`).

**That commit is inert until the next chassis roll.** It does not hold this bug open,
because it is not this bug's defect — but do not read it as live.

Deliberately NOT changed, both documented in the code: the not-opted-in exit
(`council-gate`'s positive choice) and the no-orchestration-id exit (recording there writes
`orchestration_id` NULL, which the same file records at :341-344 as itself the defect).

## The finding a later thread should actually care about

The containment test the council's guardian seat asked for **shipped inert**, and so had
three pre-existing assertions beside it. `mock.ExpectationsWereMet()` reports expectations
*registered and not consumed* — register none and it is true by construction; it never sees
an *unexpected* call. Proven by mutation: `recordPlanRefusal` moved above the opt-in check
and **all four "must not touch the DB" assertions still passed**. Replaced with an observed
logger (`dbTouchWatcher`); the same mutation now fails three tests, and the three older
guards were retrofitted. See `LANDMINES.md` (footprint `sqlmock` / `ExpectationsWereMet`)
and `WRONG_CALLS.md`. **32 other files in the tree use the same API and are [UNMEASURED].**

## Still open elsewhere, unchanged by this

- `council-gate → persist_submission` remains not opted in — a positive choice, and opting
  it in would additionally need `summary` added to the refusal result.
- The **size-cap** class (`max_stages`, `max_total_edits`) is still not repairable: those
  messages ask for less scope while the repair prompt forbids dropping it. Inherited from
  `bugs_open/099`; see concept register `FIX-057`.
- A fourth consumer exists that `162` did not name: `council-gate-036scratch`. Not a
  defect — `is_active=false`, 0 runs in 30 days.
