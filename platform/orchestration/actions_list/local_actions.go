// FILE: platform/orchestration/actions_list/local_actions.go
// A simple package with no dependencies
package actions_list

// also update registry with new actions

// LocalActions is just a list of action names that are executed locally
var LocalActions = map[string]bool{
	// Core workflow control
	"complete_workflow": true,
	"await_response":    true,

	// Agent management
	"spawn_agent":         true,
	"spawn_group":         true,
	"call_agent":          true,
	"discover_agents":     true,
	"start_orchestration": true,

	// Data operations
	"validate_input":  true,
	"transform_data":  true,
	"validate_schema": true,
	"aggregate_data":  true,
	"calculate":       true,

	// LLM operations
	"execute_llm_prompt": true,

	// Memory operations
	"retrieve_memory": true,
	"store_memory":    true,
	"cache_lookup":    true,

	// External operations
	"send_notification": true,
	"http_request":      true,

	// Storage operations
	"validate_assets":   true,
	"deploy_to_hosting": true,
	"upload_to_s3":      true,
	"s3_upload":         true,
	"store_result":      true,
	"route_storage":     true,

	// HTML operations
	"generate_html": true,
	"process_html":  true,
	"validate_html": true,

	// Workflow control
	"conditional_branch": true,
	"conditional_route":  true,

	// Planning and review
	"plan_agent_team":       true,
	"review_performance":    true,
	"approve_agent_changes": true,
	"evaluate_task":         true,

	// Legacy/duplicate
	"spawn_agent_k8s": true,
}

// IsLocalAction checks if an action is local (no dependencies)
func IsLocalAction(action string) bool {
	return LocalActions[action]
}
