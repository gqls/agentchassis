# PLAN — bugs_open/144: steps inside a `sub_workflow` are validated by nothing

**Started** 2026-07-29 · **Thread** bugsearch-7 · **Bug file**
`bugs_open/144_HANDOFF_2026-07-29_sub_workflow_steps_are_never_validated.md`
(filed by another thread, unowned when picked up — `scripts/who-owns.py 144` showed
only the filing commit and one docs commit, no owning workstream).

---

## 1. What is broken

`ValidateWorkflow` runs once, on the top-level workflow (`processor.go:276`). Steps
nested inside a loop's `sub_workflow` are extracted from config and executed directly
by the loop action (`loop_actions.go:70-88`), so every invariant the validator
enforces is unenforced for them.

The offline half — `scripts/audit-config-keys.sh` — walked
`default_config->'workflow'->'steps'`, the same top level. **Two halves blind in the
same direction agree with each other**, and consistent blindness reads exactly like
correctness. That is why this survived long enough to be found sideways, by a council
guardian objecting to a *different* claim.

## 2. What I measured before designing anything

All against live `agent_definitions` (`deleted_at IS NULL`, non-snapshot,
`is_active`), 2026-07-29. Commands in the RUNBOOK.

| | count |
|---|---|
| agents with a workflow | 178 |
| top-level steps | 1,236 (cross-checked: 1,173 object configs + 63 null = 1,236) |
| agents carrying a sub-workflow | 18 |
| sub-workflows | 19 |
| **nested steps** | **85** |
| top-level `(action, key)` pairs | 818 |
| nested `(action, key)` pairs | 66, of which **25 exist ONLY nested** |
| distinct nested actions | 20, **all `IsLocal: true`** |
| nested steps carrying a `topic` | **0** |
| nested pairs that would trip the strict-config rule | **0** |

The bug file said 64 nested steps and 24 nested-only pairs; I get 85 and 25. Not a
contradiction — I count `substeps` carriers as well as `sub_workflow` ones, and the
fleet moves. **Quote the date with the number.**

> The strict-config measurement used the SUPERSET (`opted_in`, i.e. every action that
> checks config at all, 63 of them) rather than the narrower `StrictConfig` set. That
> is the conservative direction — zero under the superset is zero under the subset —
> so the conclusion stands, but the figure is "0 of 66 pairs against 63 config-checking
> actions", not "against 63 strict actions".

## 3. Two things I got from reading the executor, not from the bug file

Both changed the design, and one of them would have been a fleet-wide outage.

**(a) A reference OUT of a sub-workflow is legitimate.** Loop expansion prefixes a
`next_step` only when it names a sibling substep; anything else passes through
untouched (`coordinator.go:4009-4014`) and may resolve against the enclosing plan —
or against `<loop>_complete`, which the expander *injects* (`loop_expansion_handler.go:192`)
and which exists in no definition. The bug's fix candidate 1 says "validate each as a
workflow in its own right"; done naively that makes every external reference a hard
error. **So an unresolved reference warns, and never fails.**

**(b) The runtime decoder drops fields.** `parseSubsteps` reads seven — action,
description, next_step, error_step, output_field, topic, config — and silently drops
`dependencies`, `sub_tasks`, `timeout`, `name`, `store_memory`, `target_agent_type`.
So a nested `dependencies` describes ordering that never happens. Decoding by JSON
round-trip into `models.Step` would have populated those fields and made the
validator vouch for behaviour the executor does not perform. **So the validator
mirrors `parseSubsteps` exactly and REPORTS the dropped keys rather than
pretend-enforcing them.**

## 4. Design

1. **`platform/validation/subworkflow.go`** — `subWorkflowsOf` (the single definition
   of where a nested step lives, mirroring the loop action's precedence: `substeps`
   wins, `sub_workflow` only if absent/empty), `DecodeSubWorkflowStep` (mirrors
   `parseSubsteps`, returns dropped keys), the recursion, and `WalkSteps` (exported).
2. **Severity split.** Hard error only where the fault is unambiguous at any depth:
   no action · remote action with no topic · nested `fan_out` (the decoder drops
   `sub_tasks`, so it would execute with none) · empty sub-workflow · `start_step`
   naming no step of that sub-workflow · cycle · nesting past a depth backstop ·
   unknown key on a `StrictConfig` action. Warning for everything validation cannot
   see: unresolved references, dropped fields, a step carrying both shapes.
3. **One config-key rule.** The bugs_open/101 check moves into `checkStepConfigKeys`,
   called from both depths. Two copies of a rule is how the two levels came to
   disagree in the first place.
4. **The audit uses the same traversal.** `cmd/config-key-audit --live-pairs` walks
   the definitions with `validation.WalkSteps`; `scripts/audit-config-keys.sh` feeds
   it the fetch and prints the coverage (top / nested / nested-only). The audit
   cannot go blind again without the validator going blind in the same commit.
5. **A dry-run harness ships with it** —
   `TestLiveDefinitionsPassSubWorkflowValidation`, skipped unless pointed at an
   export. It validates each live plan TWICE (whole, and with sub-workflows stripped)
   and reports only the difference, so a pre-existing rejection cannot be misread as
   damage from the patch.

## 5. The risk, and the answer to it

The bug names it exactly: *"a naive recursion could start rejecting workflows that run
today … Measure before shipping; that is the whole risk."*

**Measured: 0 of 178 live definitions newly rejected.** Three are rejected today and
would be rejected either way — `html-developer-chunked`, `multipage-wrapper`,
`html-assembler`, each whose top-level step names an action registered nowhere and
carries no topic. `bugs_closed/044` records this family as retired definitions still
flagged `is_active`. **Reported separately, not fixed here** (see NOTES; three live
builders still `spawn_agent`/`call_agent` at two of them).

The residual risk is honest and stated in the submission: DB config is live
immediately, so a nested step added *after* this measurement with a typo'd action is
now rejected where it previously ran and misbehaved. That is the intent — but it
means the first failure anybody sees will be a workflow that stopped, not one that
ran wrong.

## 6. Seam status

This changes what "validated" GUARANTEES fleet-wide, so under CLAUDE.md's 2026-07-29
ruling it is a platform seam. Condition (2) is met: registered as **WFA-003** in the
concept register in the same commit that ships it. Consumers named and told in the
submission rather than merely measured: `processor.go:276`, `agentbase/agent.go`,
`scripts/audit-config-keys.sh`. The venue question — council gate or RFC — was put to
the reviewers explicitly rather than assumed away.

Council submission: corr `9194bc97-8475-4022-b658-2ac64f06dd63` (2026-07-29).
