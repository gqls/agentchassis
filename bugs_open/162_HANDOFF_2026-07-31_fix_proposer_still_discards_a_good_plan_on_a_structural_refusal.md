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
