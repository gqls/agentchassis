https://claude.ai/chat/d6b7bd1a-59b0-49f2-b5ab-ecfa9bfcfde8

LOOP ACTION INTEGRATION GUIDE

Files created:
1. loop_action.go - Main loop action that prepares expansion
2. loop_expansion_handler.go - Coordinator handler that injects steps
3. loop_complete_action.go - Aggregates results at end of loop

STEP 1: Add loop_action.go to platform/actions/
Location: platform/actions/loop_action.go

STEP 2: Add loop_complete_action.go to platform/actions/
Location: platform/actions/loop_complete_action.go

STEP 3: Add loop_expansion_handler.go to platform/orchestration/
Location: platform/orchestration/loop_expansion_handler.go

STEP 4: Register actions in action registry
File: platform/actions/registry.go (or wherever actions are registered)

Find the line:
"loop":

Change to:
"loop":          LoopAction,
"loop_complete": LoopCompleteAction,

That `GlobalActionRegistry` entry is the ONLY registration needed. (This step used
to say "also add to the LocalActions list in the actions_list package" — that list
was already dead when `bugs_open/017` was diagnosed, and has been deleted. An action
missing from `GlobalActionRegistry` fails validation as "requires a topic";
`registry_parity_test.go` now fails the build instead.)

STEP 5: Integrate coordinator handler
File: platform/orchestration/coordinator.go

In executeLocalAction function, after action execution, add:

    // Execute the action
    result, err := executeAction(ctx, handler, params, contextLogger)
    if err != nil {
        return err
    }

    // NEW: Check if result is a loop expansion
    if resultMap, ok := result.(map[string]interface{}); ok {
        if isLoop, _ := resultMap["loop_action"].(bool); isLoop {
            contextLogger.Info("Detected loop action, expanding workflow")
            if err := s.handleLoopExpansion(state, resultMap, contextLogger); err != nil {
                return fmt.Errorf("failed to expand loop: %w", err)
            }
            // Loop expansion sets state.CurrentStep to first iteration step
            // Save state and continue
            if err := s.repo.UpdateState(ctx, state); err != nil {
                return fmt.Errorf("failed to save state after loop expansion: %w", err)
            }
            return s.continueExecution(ctx, state, headers)
        }
    }

    // Rest of existing code...

STEP 6: Set loop variable before step execution
File: platform/orchestration/coordinator.go

In executeLocalAction function, BEFORE executing the action handler, add:

    // NEW: Set loop variable if this is a loop iteration step
    s.setLoopVariable(state, step, contextLogger)

    // Then execute action as normal
    result, err := executeAction(ctx, handler, params, contextLogger)


HOW IT WORKS:

1. Loop action executes and returns expansion data
2. Coordinator detects loop_action: true in result
3. handleLoopExpansion injects N * M steps into workflow:
    - generate_pages_sequential_iter_0_research_page
    - generate_pages_sequential_iter_0_plan_page
    - generate_pages_sequential_iter_0_generate_content
    - generate_pages_sequential_iter_1_research_page
    - ...
    - generate_pages_sequential_complete

4. Each iteration step has loop_iteration and loop_var_name in config
5. Before executing each step, setLoopVariable sets current_page = items[N]
6. Substeps execute normally (can be async call_agent, etc.)
7. Loop completes when generate_pages_sequential_complete runs
8. Aggregated results stored in output_field

TESTING:

Use the workflow from project knowledge:
{
"action": "loop",
"config": {
"iterate_over": "flow_definition.pages",
"loop_var": "current_page",
"max_iterations": 20,
"substeps": {
"research_page": {
"action": "call_agent",
"config": {
"agent_type": "content-researcher",
"input_fields": ["current_page", "flow_definition"],
"timeout_seconds": 60
},
"output_field": "page_research"
},
"generate_page_content": {
"action": "call_agent",
"config": {
"agent_type": "content-creator",
"input_fields": ["current_page", "page_research"],
"timeout_seconds": 180
},
"output_field": "page_html"
}
}
},
"next_step": "assemble_site",
"output_field": "all_pages"
}

Expected behavior:
- If flow_definition.pages = [page1, page2, page3]
- Creates 6 steps: 3 iterations × 2 substeps
- Each iteration sets current_page to page1, page2, page3
- Results aggregated in all_pages

==============

I've created the loop action implementation with three components:

**1. loop_action.go** - Parses config and returns expansion data
**2. loop_expansion_handler.go** - Injects N×M steps into the workflow
**3. loop_complete_action.go** - Aggregates results

**Key design points:**

- Loop expands to: `loopname_iter_0_substep1` → `loopname_iter_0_substep2` → `loopname_iter_1_substep1` → ... → `loopname_complete`
- Each iteration gets its item in `current_page` (or whatever loop_var is set to)
- Works with async actions (call_agent, etc.) since each substep is a normal workflow step
- State saved between steps, so recoverable mid-loop
- Max 20 iterations safety limit

**Integration needs:**
1. Add files to platform/actions and platform/orchestration
2. Register `LoopAction` and `LoopCompleteAction` in registry
3. Add loop detection and handler call in coordinator's executeLocalAction
4. Add setLoopVariable call before step execution

Review the files and let me know if you want any changes before we integrate into the actual codebase.

