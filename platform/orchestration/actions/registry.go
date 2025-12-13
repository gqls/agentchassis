package actions

import (
	"github.com/gqls/agentchassis/platform/orchestration/actions_list"
)

// GlobalActionRegistry is the single source of truth for all available local actions
var GlobalActionRegistry = map[string]ActionFunc{
	// Core workflow control
	"complete_workflow":  CompleteWorkflowAction,
	"await_response":     AwaitResponseAction,
	"evaluate_condition": EvaluateConditionAction,

	// Agent management
	"spawn_agent":         SpawnAgentAction,
	"spawn_group":         SpawnGroupAction,
	"call_agent":          CallAgentAction,
	"discover_agents":     DiscoverAgentsAction,
	"start_orchestration": StartOrchestrationAction,

	// Image generation
	"generate_image": GenerateImageAction,

	// Web scraping
	"scrape_web":             WebscrapeAction,        // Main action
	"firecrawl_scrape":       FirecrawlScrapeAction,  // Single page
	"firecrawl_crawl":        FirecrawlCrawlAction,   // Multi-page crawl
	"firecrawl_extract":      FirecrawlExtractAction, // Structured extraction
	"validate_url":           ValidateURLAction,      // URL validation
	"aggregate_scraped_data": AggregateScrapedDataAction,
	"split_urls":             SplitURLsAction,

	// Data operations
	"validate_input":  ValidateInputAction,
	"transform_data":  TransformDataAction,
	"validate_schema": ValidateSchemaAction,

	// maths actions
	"calculate": CalculateAction,

	// Data aggregation
	"aggregate_data":    AggregateDataAction,
	"aggregate_webpage": AggregateWebpageAction,

	// LLM operations
	//"execute_llm_prompt": ExecuteLLMPromptActionFAKE,
	"execute_llm_prompt": ExecuteLLMPromptAction,

	// Web build
	"git_commit":         GitCommitAction,
	"git_commit_action":  GitCommitAction,
	"new_site_architect": AssembleFromLibraryAction,
	// "assemble_from_library":     AssembleFromLibraryAction,
	// "assemble_full_page":        AssembleFullPageAction,
	"assemble_page":           AssemblePageAction,
	"assemble_multipage_site": AssembleMultipageSiteAction,
	// "assemble_html_parts":       AssemblePageAction,
	"fetch_agent_questionnaire": FetchAgentQuestionnaireAction,
	// "wrap_multipage":            AssembleMultipageSiteAction,
	"loop":

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

	// Planning and review and agent discovery
	"evaluate_task": EvaluateTaskAction,

	// HITL / Approval actions
	"await_approval":               AwaitApprovalAction,
	"process_approval_decision":    ProcessApprovalDecisionAction,
	"process_data":                 ProcessApprovalDecisionAction, // for one of the workflows - prob delete this later
	"create_approval_request":      CreateApprovalRequestAction,
	"wait_for_approval_response":   WaitForApprovalResponseAction,
	"request_human_input":          RequestHumanInputAction,
	"process_human_input_response": ProcessHumanInputResponseAction,

	// Entity and storage actions
	"append_entity_state":      AppendEntityStateAction,
	"read_latest_entity_state": ReadLatestEntityStateAction,
	"read_entity_history":      ReadEntityHistoryAction,
	"read_my_state":            ReadMyStateAction,
	"write_my_state":           WriteMyStateAction,

	// Best Agent discovery actions
	"discover_best_agents":    DiscoverBestAgentsAction,
	"review_performance":      ReviewPerformanceAction,
	"approve_improvement":     ApproveImprovementAction,
	"query_agent_definitions": QueryAgentDefinitionsAction,

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
