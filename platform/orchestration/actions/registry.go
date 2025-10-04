package actions

import "github.com/gqls/agentchassis/platform/orchestration/actions_list"

// GlobalActionRegistry is the single source of truth for all available local actions
var GlobalActionRegistry = map[string]ActionFunc{
	// Core workflow control
	"complete_workflow": CompleteWorkflowAction,
	"await_response":    AwaitResponseAction,

	// Agent management
	"spawn_agent":         SpawnAgentAction,
	"spawn_group":         SpawnGroupAction,
	"call_agent":          CallAgentAction,
	"discover_agents":     DiscoverAgentsAction,
	"start_orchestration": StartOrchestrationAction,

	// Data operations
	"validate_input":  ValidateInputAction,
	"transform_data":  TransformDataAction,
	"validate_schema": ValidateSchemaAction,
	"aggregate_data":  AggregateDataAction,
	"calculate":       CalculateAction,

	// LLM operations
	"execute_llm_prompt": ExecuteLLMPromptActionFAKE,

	// Memory operations
	"retrieve_memory": RetrieveMemoryAction,
	"store_memory":    StoreMemoryAction,
	"cache_lookup":    CacheLookupAction,

	// External operations
	"send_notification": SendNotificationAction,
	"http_request":      HTTPRequestAction,

	// Storage operations
	"validate_assets":   ValidateAssetsAction,
	"deploy_to_hosting": DeployToHostingAction,
	"upload_to_s3":      UploadToS3Action,
	"s3_upload":         UploadToS3Action, // Alias
	"store_result":      StoreResultAction,
	"route_storage":     RouteStorageAction,

	// HTML operations
	"generate_html": GenerateHTMLAction,
	"process_html":  ProcessHTMLAction,
	"validate_html": ValidateHTMLAction,

	// Workflow control
	"conditional_branch": ConditionalBranchAction,
	"conditional_route":  ConditionalRouteAction,

	// Planning and review
	"plan_agent_team":       PlanAgentTeamAction,
	"review_performance":    ReviewPerformanceAction,
	"approve_agent_changes": ApproveAgentChangesAction,
	"evaluate_task":         EvaluateTaskAction,

	// Legacy/duplicate
	"spawn_agent_k8s": SpawnAgentAction, // Same as spawn_agent
}

// IsLocalAction checks if an action is available for local execution
// delegates to actions_list
func IsLocalAction(action string) bool {
	return actions_list.IsLocalAction(action)
}

// GetAction returns the action function if it exists
func GetAction(action string) (ActionFunc, bool) {
	fn, exists := GlobalActionRegistry[action]
	return fn, exists
}

// ListActions returns all available action names (useful for debugging/documentation)
func ListActions() []string {
	actions := make([]string, 0, len(GlobalActionRegistry))
	for name := range GlobalActionRegistry {
		actions = append(actions, name)
	}
	return actions
}
