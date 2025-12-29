// FILE: platform/orchestration/actions_list/local_actions.go
// A simple package with no dependencies
package actions_list

// also update registry with new actions

// LocalActions is just a list of action names that are executed locally
var LocalActions = map[string]bool{
	// Core workflow control
	"complete_workflow":  true,
	"await_response":     true,
	"evaluate_condition": true,

	// Agent management
	"spawn_agent":         true,
	"spawn_group":         true,
	"call_agent":          true,
	"discover_agents":     true,
	"start_orchestration": true,

	// Image handling
	"generate_image": true,

	// Web scraping
	"scrape_web":             true,
	"firecrawl_scrape":       true,
	"firecrawl_crawl":        true,
	"firecrawl_extract":      true,
	"validate_url":           true,
	"aggregate_scraped_data": true,
	"split_urls":             true,

	// Data operations
	"validate_input":   true,
	"transform_data":   true,
	"validate_schema":  true,
	"parse_json_field": true,
	"extract_field":    true,

	"calculate": true,

	// Web build
	"git_commit":            true,
	"git_commit_action":     true,
	"new_site_architect":    true,
	"assemble_from_library": true,
	// "assemble_full_page":        true,
	"assemble_page":           true,
	"assemble_multipage_site": true,
	// "assemble_html_parts":       true,
	"fetch_agent_questionnaire": true,
	// "wrap_multipage":            true,
	"loop":          true,
	"loop_complete": true,

	"insert_research_result":  true,
	"select_style_collection": true,
	"update_site_content":     true,
	"update_site_status":      true,
	"update_site_defaults":    true,
	"update_page_status":      true,
	"build_render_context":    true,
	"render_component":        true,
	"compile_page_sections":   true,
	"db_sync":                 true,
	"store_asset":             true,

	// Site and Link Management (Database Integration)
	"ensure_site_record":     true,
	"validate_site_plan":     true,
	"sync_pages_to_db":       true,
	"extract_and_sync_links": true,
	"update_site_timestamps": true,
	"get_navigation_from_db": true,

	// Data aggregation
	"aggregate_data":    true,
	"aggregate_webpage": true,

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
	"query_database":    true,

	// HTML operations
	"generate_html": true,
	"process_html":  true,
	"validate_html": true,

	// Workflow control
	"conditional_branch": true,
	"conditional":        true,
	"conditional_route":  true,

	// Planning and review
	"evaluate_task": true,

	// HITL / Approval actions
	"await_approval":               true,
	"process_approval_decision":    true,
	"create_approval_request":      true,
	"wait_for_approval_response":   true,
	"request_human_input":          true,
	"process_human_input_response": true,

	// Entity and storage actions
	"append_entity_state":      true,
	"read_latest_entity_state": true,
	"read_entity_history":      true,
	"read_my_state":            true,
	"write_my_state":           true,

	// Best Agent discovery actions
	"discover_best_agents":    true,
	"review_performance":      true,
	"approve_improvement":     true,
	"query_agent_definitions": true,

	// Legacy/duplicate
	"spawn_agent_k8s": true,
}

// IsLocalAction checks if an action is local (no dependencies)
func IsLocalAction(action string) bool {
	return LocalActions[action]
}
