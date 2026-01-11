How Conditional Branching Works
With evaluate_condition, your workflow can now branch based on conditions:

{
"steps": {
"check_condition": {
"action": "evaluate_condition",
"config": {
"condition": "{{.input_data.extract_structured}}",
"default": false
},
"next_step": {
"true": "do_structured_extraction",
"false": "do_basic_scrape"
}
}
}
}

The action:

Evaluates the template expression (e.g., {{.input_data.extract_structured}})
Returns a result with "result": true or "result": false
The orchestrator uses this to pick the next step from the next_step map
