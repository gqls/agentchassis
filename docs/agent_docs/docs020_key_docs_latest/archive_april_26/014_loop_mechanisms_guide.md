# Loop Mechanisms in the Orchestration Framework

## Overview

Loops allow a workflow to iterate over a collection of items, executing a series of substeps for each item. Loops are not Go for-loops — they are dynamic workflow expansion. At runtime, the loop step injects N × M steps into the workflow plan (N items × M substeps per iteration), then the normal step-execution engine drives through them sequentially.

This guide covers everything needed to define, debug, and extend loops in agent workflows.

---

## How Loops Work — The Full Lifecycle

### Phase 1: Definition (SQL workflow)

A loop is defined as a step with `"action": "loop"` in the agent's workflow definition. It specifies what collection to iterate over, what variable name to use for the current item, and what substeps to run per iteration.

### Phase 2: LoopAction (Go action)

When `continueExecution` reaches the loop step, it calls `LoopAction` in `loop_actions.go`. This action does not execute the iterations — it reads the config, resolves the collection from `CollectedData`, parses the substeps into `models.Step` structs, determines their order, and returns a special result map with `"loop_action": true`.

### Phase 3: Loop Expansion (coordinator)

Back in `executeLocalAction` (coordinator.go), the code detects `loop_action: true` in the result and calls `handleLoopExpansion` in `loop_expansion_handler.go`. This is where the real work happens:

1. For each item in the collection, for each substep in order, a new step is injected into `state.WorkflowPlan.Steps` with a name following the pattern `{loop_name}_iter_{N}_{substep_name}`.

2. Each injected step gets iteration-specific config injected: `loop_iteration`, `loop_item_index`, `loop_var_name`, `loop_name`, `continue_on_error`.

3. Each item is stored in `CollectedData` at `{loop_name}_item_{N}`.

4. A `{loop_name}_complete` step is created with action `loop_complete` that aggregates results after all iterations finish.

5. `state.CurrentStep` is set to the first injected step.

6. State is persisted to the database.

7. `executeLocalAction` calls `continueExecution` recursively to begin executing the first iteration's substeps.

8. After the recursive call returns, `executeLocalAction` returns `ErrLoopExpansionHandled` — a sentinel that tells the outer `continueExecution` for-loop to exit cleanly without falling through to the loop step's `NextStep`.

### Phase 4: Iteration Execution

The injected steps execute as normal workflow steps. Before each loop substep runs, `setLoopVariable` sets the current item in `CollectedData` under the configured variable name (e.g. `current_section` or `current_item`). It also propagates outputs from earlier substeps in the same iteration so they're accessible under their base names (not just their iteration-suffixed names).

### Phase 5: Iteration Boundaries

When an iteration's last substep finishes, its `NextStep` points to either:
- The first substep of the next iteration (e.g. `{loop}_iter_{N+1}_{first_substep}`), or
- `{loop_name}_complete` if this was the last iteration.

This is resolved at expansion time by `resolveIterationNextStep`.

### Phase 6: Loop Complete

`LoopCompleteAction` runs as the final step. It iterates through `0..total_iterations-1`, collecting each iteration's substep outputs from `CollectedData` using the `substep_output_fields` list (set during expansion). It assembles these into an array of result objects and stores them under the loop step's `output_field`.

After `loop_complete`, execution continues to whatever `next_step` was defined on the original loop step.

---

## Defining a Loop in a Workflow

### Minimal Example

```json
"process_items": {
    "action": "loop",
    "config": {
        "iterate_over": "loaded_data.items",
        "loop_var": "current_item",
        "max_iterations": 20,
        "sub_workflow": {
            "start_step": "do_work",
            "steps": {
                "do_work": {
                    "action": "some_action",
                    "config": {
                        "input_from": "current_item.field_name"
                    },
                    "output_field": "work_result",
                    "description": "Process the item"
                }
            }
        }
    },
    "next_step": "after_loop",
    "description": "Process each item",
    "output_field": "items_processed"
}
```

### Config Fields

| Field | Required | Description |
|-------|----------|-------------|
| `iterate_over` or `items_field` | Yes | Dot-notation path into `CollectedData` to find the array. E.g. `"pending.items"` or `"section_components.components"`. |
| `loop_var` or `item_variable` | No (default: `"loop_item"`) | Name under which the current item is stored in `CollectedData` before each substep runs. |
| `max_iterations` | No (default: 20) | Safety limit. Collections larger than this are truncated. |
| `continue_on_error` | No (default: false) | When true, a failed substep records the error and skips to the next iteration instead of failing the workflow. |
| `allow_missing` | No (default: false) | When true, a missing collection at `iterate_over` returns an empty result rather than an error. |
| `sub_workflow.start_step` | Recommended | Name of the first substep. If omitted, the framework auto-detects by finding the substep with no incoming `next_step` references. Specify it explicitly to avoid ambiguity. |
| `sub_workflow.steps` | Yes | Map of substep definitions. Same structure as top-level workflow steps. |

The older `substeps` key (without the `sub_workflow` wrapper) also works but `sub_workflow` is the standard form used by workflow definitions.

### The Loop Step Itself

| Field | Description |
|-------|-------------|
| `next_step` | Where to go after the loop completes (after `loop_complete` runs). |
| `output_field` | The key in `CollectedData` where the aggregated results array is stored. |
| `description` | For logging. |

---

## Substep Definitions

Substeps are defined the same as regular workflow steps. Each substep has `action`, `config`, `output_field`, `next_step`, and `description`.

### How next_step Works Inside Substeps

Substep `next_step` values are resolved at expansion time:

1. **References another substep name** → Prefixed with iteration context. `"next_step": "generate_content"` becomes `{loop}_iter_{N}_generate_content`.

2. **Empty string or omitted** → This is the terminal substep of the iteration. It chains to the first substep of the next iteration, or to `{loop}_complete` on the last iteration.

3. **References a step name that is NOT a substep** → Left as-is. This would exit the loop early (unusual, but supported for edge cases like `"done"` mapping to `loop_complete`).

### How Config References Are Prefixed

Config fields that reference other substeps (`then_step`, `else_step`, `fallback_step`, `error_step`, `on_success`, `on_failure`) are also prefixed with the iteration index at expansion time. This means conditionals inside loops work correctly.

Config fields that reference data produced by other substeps (`content_from`, `context_from`, `data_from`, `source_field`, `input_from`, `result_from`, `content_field`, `commit_from`) have their values prefixed when they match a substep's `output_field`. For example, if substep `generate_content` has `output_field: "generated_content"`, and a later substep has `"content_from": "generated_content.result"`, the expansion rewrites this to `"content_from": "generated_content_{N}.result"` for iteration N.

### OutputField Suffixing

Each substep's `output_field` is made unique per iteration by `makeIterationOutputField`. A substep with `output_field: "claim_result"` stores its result at `claim_result_0`, `claim_result_1`, etc. in `CollectedData`.

However, `setLoopVariable` propagates these back to their base names before each substep runs. So within an iteration, a substep can reference `claim_result` without the suffix and it will resolve to the current iteration's value.

---

## Accessing Data Inside Loop Substeps

### The Loop Variable

Before each substep executes, `setLoopVariable` sets:
- `CollectedData[loop_var]` = the current item (e.g. `CollectedData["current_item"]` = the Nth item from the collection)
- `CollectedData["__current_loop_item_key__"]` = `"{loop_name}_item_{N}"` (for debugging)

### Outputs from Earlier Substeps in the Same Iteration

`setLoopVariable` also propagates iteration-specific outputs to their base names. If iteration 2's `claim` substep stored a result at `claim_result_2`, then before the `check_claim` substep runs, `setLoopVariable` copies `claim_result_2` to `claim_result`. This means substep configs can reference other substep outputs by their base `output_field` name without worrying about iteration suffixes.

### Outputs from the Parent Workflow

Substeps can access anything in `CollectedData` that was set before the loop started. For example, if a prior step stored `site_record`, any substep can reference `site_record.domain`.

### Outputs from Other Iterations

These are accessible via their suffixed keys: `claim_result_0`, `claim_result_1`, etc. But typically substeps only care about their own iteration's data.

---

## The Dispatch Loop Pattern

The most common loop pattern in the system is the dispatch loop: claim an item, spawn a handler, call the handler, mark complete. Here is the production example from `build-dispatch-loop`:

```json
"process_items": {
    "action": "loop",
    "config": {
        "items_field": "pending.items",
        "item_variable": "current_item",
        "max_iterations": 50,
        "continue_on_error": true,
        "sub_workflow": {
            "start_step": "claim",
            "steps": {
                "claim": {
                    "action": "claim_work_item",
                    "config": { "work_item_id": "current_item.id" },
                    "next_step": "check_claim",
                    "output_field": "claim_result",
                    "description": "Atomically claim item"
                },
                "check_claim": {
                    "action": "conditional",
                    "config": {
                        "condition": "claim_result.claimed == true",
                        "then_step": "spawn_handler",
                        "else_step": "done"
                    },
                    "description": "Skip if already claimed"
                },
                "spawn_handler": {
                    "action": "spawn_agent",
                    "config": {
                        "role": "handler",
                        "agent_type_field": "current_item.handler_agent"
                    },
                    "next_step": "call_handler",
                    "output_field": "handler_spawned",
                    "description": "Spawn handler (dynamic type per item)"
                },
                "call_handler": {
                    "action": "call_agent",
                    "config": {
                        "target_role": "handler",
                        "error_step": "mark_failed",
                        "input_mapping": {
                            "site_id": "current_item.site_id",
                            "domain": "input_data.domain",
                            "work_item_id": "current_item.id",
                            "item_type": "current_item.item_type",
                            "spec": "current_item.spec"
                        },
                        "timeout_seconds": 300
                    },
                    "next_step": "mark_complete",
                    "output_field": "handler_result",
                    "description": "Call handler agent"
                },
                "mark_complete": {
                    "action": "complete_work_item",
                    "config": {
                        "work_item_id": "current_item.id",
                        "result": "handler_result"
                    },
                    "next_step": "done",
                    "output_field": "item_completed",
                    "description": "Mark item complete"
                },
                "mark_failed": {
                    "action": "fail_work_item",
                    "config": {
                        "work_item_id": "current_item.id",
                        "error_message": "Handler failed"
                    },
                    "next_step": "done",
                    "output_field": "item_failed",
                    "description": "Mark item failed"
                },
                "done": {
                    "action": "loop_complete",
                    "description": "Item done"
                }
            }
        }
    },
    "next_step": "complete",
    "output_field": "items_processed",
    "description": "Process each item: claim → spawn → call → mark"
}
```

Key design points:
- `continue_on_error: true` means one failed item doesn't kill the whole batch.
- `claim_work_item` uses atomic database claims to prevent double-dispatch across concurrent instances.
- `spawn_handler` uses `agent_type_field` for dynamic type resolution — the handler agent type comes from the work item data, not the workflow definition.
- `call_handler` uses `target_role: "handler"` to find the just-spawned agent. Always use role-based lookup, never type-based.
- The `done` step with `action: "loop_complete"` is the iteration's terminal substep. Its empty `next_step` means the framework chains to the next iteration or to `{loop}_complete`.
- `error_step: "mark_failed"` on `call_handler` routes handler failures to mark the item as failed rather than crashing the loop.

---

## Conditionals Inside Loops

Loops can contain conditional substeps. The `then_step` and `else_step` values reference other substep names (without iteration prefix — the expansion handles prefixing).

```json
"check_render_mode": {
    "action": "conditional",
    "config": {
        "condition": "current_section.render_mode == 'agent'",
        "then_step": "call_researcher",
        "else_step": "render_from_template"
    },
    "description": "Route based on render mode"
}
```

At expansion time for iteration 2, this becomes:

```
Step: process_sections_loop_iter_2_check_render_mode
Config:
  then_step: process_sections_loop_iter_2_call_researcher
  else_step: process_sections_loop_iter_2_render_from_template
```

The conditional action returns a `next_step` override in its result, which the `continueExecution` for-loop picks up via `getNextStepFromResult`.

---

## Substep Ordering and start_step

The framework determines substep execution order by following `next_step` chains, starting from `start_step`. The algorithm (`buildSubstepOrder`):

1. Use `start_step` if specified and exists in the substeps map.
2. Otherwise auto-detect: find the substep that no other substep references as its `next_step` (the entry point with no incoming edges).
3. Follow the `next_step` chain from there, collecting each step.
4. Any substeps not in the chain are appended at the end (with a warning log).

The order matters because it determines which substep is "first" for each iteration (used for iteration boundary chaining) and which is "last" (its empty `next_step` triggers the cross-iteration chain).

When using conditionals, some substeps will not be in the main chain because they're only reachable via `then_step`/`else_step`. These show up in logs as "Substep not in chain, appending" — this is expected and harmless. The chain order is only used for determining the first substep and for `substep_output_fields` ordering in the `_complete` step.

**Recommendation:** Always specify `start_step` explicitly. Auto-detection depends on `next_step` chain analysis which can be ambiguous when conditionals create branching paths.

---

## Error Handling

### continue_on_error

When `continue_on_error: true` is set on the loop config:

1. If a substep fails, `shouldContinueLoopOnError` checks the step's config for the `continue_on_error` flag (propagated during expansion).
2. `skipToNextLoopIteration` records the error at `{loop_name}_iter_{N}_error` in `CollectedData`, increments `{loop_name}_error_count`, and advances `CurrentStep` to the next iteration's first substep (or `_complete` if it was the last iteration).
3. Execution continues from the new step.

### error_step on Individual Substeps

Substeps can define an `error_step` in their config (e.g. `"error_step": "mark_failed"`). This is handled by the normal `routeToErrorStepOrFail` mechanism. Since `error_step` values that reference other substeps are prefixed during expansion, the error routes correctly within the iteration.

### Async Failures (Timeouts)

When a `call_agent` substep times out, `handleRequestTimeout` uses `skipToNextLoopIterationForAsync` which is the same as `skipToNextLoopIteration` but also cleans up the awaited request.

---

## Loop Completion and Result Aggregation

The `{loop_name}_complete` step runs `LoopCompleteAction`. Its config contains:

```json
{
    "loop_name": "process_items",
    "total_iterations": 4,
    "substep_output_fields": ["claim_result", "handler_result", "item_completed"]
}
```

`substep_output_fields` is built automatically during expansion from the substeps that have non-empty `output_field` values. `LoopCompleteAction` iterates through each iteration index, looks up `{field}_{N}` in `CollectedData` for each field, and assembles per-iteration result objects.

The aggregated array is stored at the loop step's `output_field` (e.g. `items_processed`), making it available to subsequent workflow steps.

---

## The Race Condition and ErrLoopExpansionHandled

### The Problem

When `executeLocalAction` finishes loop expansion, it calls `continueExecution` recursively to begin executing iteration substeps. If a substep calls a child agent that responds very quickly (under ~1 second), the response handler goroutine can advance the workflow state in the database before the recursive `continueExecution` returns.

When the recursive call returns nil, the outer `continueExecution` for-loop resumes with stale context. It checks the DB for `AWAITING_RESPONSES` status — but the response handler already moved past that. It then falls through to the step-transition logic using `currentStepConfig.NextStep`, which is the original loop step's `NextStep` (e.g. `"complete"`). This skips all remaining iterations.

### The Fix

`executeLocalAction` returns `ErrLoopExpansionHandled` after the recursive `continueExecution` call. The outer `continueExecution` catches this sentinel before any other error handling and returns nil — never reaching the stale `NextStep` transition.

This is safe in all timing scenarios:
- **Fast response:** Thread B (response handler) already advanced the workflow. Outer returns nil. Thread B continues.
- **Slow response:** Recursive CE returns nil (state persisted as AWAITING). Outer returns nil via sentinel. Workflow resumes when response arrives.
- **Error:** If the recursive `continueExecution` returns a real error, that error is returned instead of the sentinel, and normal error handling applies.

### Implication for New Loops

You don't need to do anything special to handle this — the sentinel is in the framework code, not in workflow definitions. But be aware that loops with fast-responding child agents are the trigger. If you're debugging a loop that completes too early (skipping iterations), check for this pattern first.

---

## Naming Conventions

| Thing | Pattern | Example |
|-------|---------|---------|
| Loop step name | Descriptive, often ending in `_loop` | `process_items`, `process_sections_loop` |
| Injected step | `{loop_name}_iter_{N}_{substep_name}` | `process_items_iter_2_claim` |
| Completion step | `{loop_name}_complete` | `process_items_complete` |
| Item storage key | `{loop_name}_item_{N}` | `process_items_item_3` |
| Iteration output | `{output_field}_{N}` | `claim_result_2` |
| Error key | `{loop_name}_iter_{N}_error` | `process_items_iter_1_error` |
| Error count | `{loop_name}_error_count` | `process_items_error_count` |
| Loop metadata | `loop_metadata` | Shared key — beware if multiple loops exist |

### The loop_metadata Caveat

`loop_metadata` is a shared key in `CollectedData`. If your workflow has two sequential loops, the second loop's expansion overwrites `loop_metadata`. The `_complete` step config has its own `loop_name` and `total_iterations` (set during expansion), so `LoopCompleteAction` reads from config first and only falls back to `loop_metadata` for backward compatibility. This means multiple sequential loops work correctly — but don't rely on `loop_metadata` for anything except debugging.

---

## Diagnostic Queries

### Check what the expansion produced

```sql
SELECT
    jsonb_object_keys(workflow_plan->'Steps') as step_name
FROM orchestration_states
WHERE orchestration_id = '<id>'
ORDER BY step_name;
```

Filter to just the loop's steps:

```sql
SELECT
    key as step_name,
    value->>'action' as action,
    value->>'next_step' as next_step,
    value->>'output_field' as output_field
FROM orchestration_states,
     jsonb_each(workflow_plan->'Steps')
WHERE orchestration_id = '<id>'
  AND key LIKE 'process_items_iter_%'
ORDER BY key;
```

### Check CollectedData for iteration results

```sql
SELECT
    key,
    CASE
        WHEN length(value::text) > 200 THEN left(value::text, 200) || '...'
        ELSE value::text
    END as value_preview
FROM orchestration_states,
     jsonb_each(collected_data::jsonb)
WHERE orchestration_id = '<id>'
  AND (key LIKE '%_iter_%' OR key LIKE '%_item_%' OR key = 'loop_metadata')
ORDER BY key;
```

### Check for stuck loops

```sql
SELECT
    orchestration_id,
    current_step,
    status,
    last_activity,
    NOW() - last_activity as stale_for
FROM orchestration_states
WHERE current_step LIKE '%_iter_%'
  AND status = 'AWAITING_RESPONSES'
  AND last_activity < NOW() - INTERVAL '5 minutes';
```

---

## Checklist for Adding a New Loop

1. **Ensure the collection exists.** The step that produces the collection must store it at a `CollectedData` path that matches your `iterate_over`/`items_field`. Test that the path resolves — dot-notation traversal uses `getNestedValueForLoop`.

2. **Choose a meaningful loop_var name.** This is how substeps access the current item. Use something descriptive: `current_item`, `current_section`, `current_fix_item`.

3. **Define substeps with explicit next_step chains.** Each substep should have a `next_step` pointing to the next substep in the flow, except the terminal substep which should have an empty or omitted `next_step`.

4. **Specify start_step.** Don't rely on auto-detection.

5. **Set output_field on substeps that produce data.** Only substeps with `output_field` have their results tracked in the aggregation.

6. **Set output_field on the loop step itself.** This is where the aggregated results go.

7. **Set continue_on_error if appropriate.** For dispatch loops processing independent items, this is almost always what you want. For loops where iteration N depends on iteration N-1, leave it false.

8. **Set max_iterations.** The default of 20 is usually fine, but if you know you'll process more items (e.g. dispatch loops), set it higher. The production dispatch loop uses 50.

9. **Test with fast-responding handlers.** If your loop calls child agents, test with agents that respond in under a second to confirm the race condition fix (ErrLoopExpansionHandled) is working.

10. **Don't nest loops inside loops in the same workflow.** If you need nested iteration, spawn a sub-agent whose workflow contains its own loop. This keeps logs clear and responsibilities separate.

---

## Files Reference

| File | What It Does |
|------|-------------|
| `platform/orchestration/actions/loop_actions.go` | `LoopAction` — reads config, resolves collection, parses substeps, returns expansion data |
| `platform/orchestration/loop_expansion_handler.go` | `handleLoopExpansion` — injects steps into workflow plan, creates `_complete` step |
| `platform/orchestration/coordinator.go` | `executeLocalAction` — detects loop expansion, calls `handleLoopExpansion`, recursive `continueExecution`, returns `ErrLoopExpansionHandled` |
| `platform/orchestration/coordinator.go` | `continueExecution` — main execution loop, catches sentinel, manages state transitions |
| `platform/orchestration/coordinator.go` | `setLoopVariable` — sets current item and propagates iteration outputs |
| `platform/orchestration/loop_error_handler.go` | `skipToNextLoopIteration`, `shouldContinueLoopOnError`, `findFirstSubstep`, `parseLoopStepName` |
| `platform/orchestration/actions/loop_actions.go` | `LoopCompleteAction` — aggregates iteration results |


