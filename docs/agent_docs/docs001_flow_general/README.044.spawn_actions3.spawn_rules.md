The naming conventions are now important because we're using them to find spawned agents:
Required Patterns:

Spawn steps: Must start with spawn_ followed by a descriptor

Examples: spawn_adder, spawn_multiplier, spawn_calculator
The suffix after spawn_ helps identify the role


Action steps: Can use perform_, execute_, process_, etc.

Examples: perform_addition, execute_calculation, process_data
These reference spawned agents by role


Step uniqueness: Each step name must be unique within a workflow

Use 3-letter suffixes if needed: validate_input_aaa, validate_output_bbb


Workflow Creation Rules Summary
When creating workflows:
{
"steps": {
// Spawn steps - prefix with "spawn_"
"spawn_validator_aaa": {
"action": "spawn_agent",
"config": {"agent_type": "validator", "role": "input_validator"}
},

    // Action steps - use descriptive names
    "validate_user_input_aab": {
      "action": "call_agent",
      "config": {"target_role": "input_validator"}
    },
    
    // Ensure uniqueness with suffixes
    "transform_data_aac": {
      "action": "call_agent",
      "config": {"target_role": "transformer"}
    }
}
}