// Workflow Validator Tool v2
// Validates that workflow field references match contracts and action expectations
// NOW INCLUDES: call_agent response wrapping detection
//
// The key insight: certain actions (call_agent, spawn_agent) wrap their responses
// under a ".response" key. This validator detects when workflows reference these
// outputs without the required wrapper.
//
// Usage:
//   go run main.go -agents agents.json
//   go run main.go -agent-json '{"type":"pageflow-builder",...}'

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// AgentDefinition represents a row from agent_definitions table
type AgentDefinition struct {
	ID             string          `json:"id"`
	Type           string          `json:"type"`
	DisplayName    string          `json:"display_name"`
	DefaultConfig  json.RawMessage `json:"default_config"`
	InputContract  *InputContract  `json:"input_contract"`
	OutputContract *OutputContract `json:"output_contract"`
}

type InputContract struct {
	Expects  map[string]string `json:"expects"`
	Required []string          `json:"required"`
}

type OutputContract struct {
	Produces interface{} `json:"produces"`
	Format   interface{} `json:"format"`
}

type WorkflowConfig struct {
	Workflow       Workflow `json:"workflow"`
	ProcessingMode string   `json:"processing_mode"`
}

type Workflow struct {
	StartStep string          `json:"start_step"`
	Steps     map[string]Step `json:"steps"`
}

type Step struct {
	Action      string                 `json:"action"`
	Config      map[string]interface{} `json:"config"`
	NextStep    string                 `json:"next_step"`
	OutputField string                 `json:"output_field"`
	Description string                 `json:"description"`
}

// ActionSpec defines what fields an action reads and writes
type ActionSpec struct {
	Name        string   `json:"name"`
	ReadsFrom   []string `json:"reads_from"`
	WritesTo    string   `json:"writes_to"`
	RequiredCfg []string `json:"required_cfg"`
	OptionalCfg []string `json:"optional_cfg"`
}

// DataFlowEntry tracks what data is available and how it's wrapped
type DataFlowEntry struct {
	StepName    string // Step that produced this
	Action      string // Action type that produced it
	OutputField string // The field name it's stored under
	WrapsIn     string // If action wraps response (e.g., "response" for call_agent)
}

// ValidationIssue represents a problem found
type ValidationIssue struct {
	AgentType string `json:"agent_type"`
	StepName  string `json:"step_name"`
	Severity  string `json:"severity"`
	Category  string `json:"category"`
	Message   string `json:"message"`
	Suggested string `json:"suggested,omitempty"`
}

// ============================================================================
// ACTION WRAPPING BEHAVIOR - THE KEY INSIGHT
// ============================================================================
// This is derived from applyResponseToState() in the Go codebase:
//
//	if step.Action == "spawn_agent" || step.Action == "call_agent" {
//	    existingData["response"] = normalisedData
//	}
//
// This means when you reference call_agent output, the actual data is under .response
// ============================================================================
var actionWrappingBehavior = map[string]string{
	"call_agent":  "response", // call_agent stores actual response under .response
	"spawn_agent": "response", // spawn_agent also wraps under .response
}

// Built-in action specifications
var defaultActionSpecs = map[string]ActionSpec{
	"call_agent": {
		Name:        "call_agent",
		ReadsFrom:   []string{"input_fields"},
		WritesTo:    "output_field",
		RequiredCfg: []string{"agent_type", "target_role"},
		OptionalCfg: []string{"input_fields", "timeout_seconds"},
	},
	"spawn_agent": {
		Name:        "spawn_agent",
		ReadsFrom:   []string{"input_fields"},
		WritesTo:    "output_field",
		RequiredCfg: []string{"role", "agent_type"},
		OptionalCfg: []string{"input_fields", "agent_type_field"},
	},
	"execute_llm_prompt": {
		Name:        "execute_llm_prompt",
		ReadsFrom:   []string{"input_fields", "prompt_template"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"input_fields", "output_format", "ai_service", "prompt_template"},
	},
	"conditional": {
		Name:        "conditional",
		ReadsFrom:   []string{"condition"},
		WritesTo:    "",
		RequiredCfg: []string{"condition", "then_step"},
		OptionalCfg: []string{"else_step"},
	},
	"loop": {
		Name:        "loop",
		ReadsFrom:   []string{"items_field", "iterate_over"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"items_field", "iterate_over", "item_variable", "loop_var", "max_iterations", "mode", "sub_workflow", "substeps"},
	},
	"complete_workflow": {
		Name:        "complete_workflow",
		ReadsFrom:   []string{"output_fields", "output"},
		WritesTo:    "",
		RequiredCfg: []string{},
		OptionalCfg: []string{"output_fields", "output_field", "output"},
	},
	"sync_pages_to_db": {
		Name:        "sync_pages_to_db",
		ReadsFrom:   []string{"input_fields"},
		WritesTo:    "output_field",
		RequiredCfg: []string{"input_fields"},
		OptionalCfg: []string{},
	},
	"store_asset": {
		Name:        "store_asset",
		ReadsFrom:   []string{"url_field", "site_id_field", "origin_prompt_field"},
		WritesTo:    "output_field",
		RequiredCfg: []string{"asset_type", "origin_type"},
		OptionalCfg: []string{"url_field", "site_id_field", "purpose", "brand_asset_key", "origin_prompt_field"},
	},
	"select_style_collection": {
		Name:        "select_style_collection",
		ReadsFrom:   []string{"site_id_field", "style_from", "domain_field"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"site_id_field", "style_collection_id", "style_from", "domain_field", "fallback_by_domain"},
	},
	"git_commit": {
		Name:        "git_commit",
		ReadsFrom:   []string{"html_from", "content_field", "page_from", "site_id_from"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"html_from", "content_field", "page_from", "site_id_from", "domain_field", "files_field"},
	},
	"assemble_page": {
		Name:        "assemble_page",
		ReadsFrom:   []string{"page_from", "content_from", "site_id_from"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"page_from", "content_from", "site_id_from", "content_field", "add_navigation"},
	},
	"update_page_status": {
		Name:        "update_page_status",
		ReadsFrom:   []string{"page_id_from", "commit_from"},
		WritesTo:    "output_field",
		RequiredCfg: []string{"status"},
		OptionalCfg: []string{"page_id_field", "page_id_from", "commit_from", "site_id_field"},
	},
	"update_site_status": {
		Name:        "update_site_status",
		ReadsFrom:   []string{"site_id_field"},
		WritesTo:    "output_field",
		RequiredCfg: []string{"status"},
		OptionalCfg: []string{"site_id_field", "deployed_at"},
	},
	"update_site_content": {
		Name:        "update_site_content",
		ReadsFrom:   []string{"site_id_field", "content_field"},
		WritesTo:    "output_field",
		RequiredCfg: []string{"content_field"},
		OptionalCfg: []string{"site_id_field", "merge"},
	},
	"ensure_site_record": {
		Name:        "ensure_site_record",
		ReadsFrom:   []string{"input_fields", "domain_field"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"input_fields", "domain_field", "store_brief_in_content_data"},
	},
	"request_human_input": {
		Name:        "request_human_input",
		ReadsFrom:   []string{"data_field", "default_from"},
		WritesTo:    "output_field",
		RequiredCfg: []string{"request_type"},
		OptionalCfg: []string{"title", "message", "fields", "data_field", "timeout_seconds"},
	},
}

func main() {
	agentsFile := flag.String("agents", "", "Path to JSON file with agent definitions array")
	agentJSON := flag.String("agent-json", "", "Inline JSON for a single agent definition")
	verbose := flag.Bool("verbose", false, "Show detailed output")
	flag.Parse()

	if *agentsFile == "" && *agentJSON == "" {
		fmt.Println("Usage: go run main.go -agents agents.json")
		fmt.Println("   or: go run main.go -agent-json '{...}'")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load agents
	var agents []AgentDefinition
	if *agentJSON != "" {
		var agent AgentDefinition
		if err := json.Unmarshal([]byte(*agentJSON), &agent); err != nil {
			fmt.Printf("Error parsing agent JSON: %v\n", err)
			os.Exit(1)
		}
		agents = append(agents, agent)
	} else {
		data, err := os.ReadFile(*agentsFile)
		if err != nil {
			fmt.Printf("Error reading agents file: %v\n", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(data, &agents); err != nil {
			fmt.Printf("Error parsing agents file: %v\n", err)
			os.Exit(1)
		}
	}

	// Validate each agent
	var allIssues []ValidationIssue
	for _, agent := range agents {
		fmt.Printf("Validating: %s\n", agent.Type)
		issues := validateAgent(agent, defaultActionSpecs)
		allIssues = append(allIssues, issues...)
	}

	reportIssues(allIssues, *verbose)

	// Exit with error code if errors found
	for _, issue := range allIssues {
		if issue.Severity == "error" {
			os.Exit(1)
		}
	}
}

func validateAgent(agent AgentDefinition, actionSpecs map[string]ActionSpec) []ValidationIssue {
	var issues []ValidationIssue

	// Parse workflow config
	var config WorkflowConfig
	if err := json.Unmarshal(agent.DefaultConfig, &config); err != nil {
		issues = append(issues, ValidationIssue{
			AgentType: agent.Type,
			Severity:  "error",
			Category:  "parse_error",
			Message:   fmt.Sprintf("Failed to parse default_config: %v", err),
		})
		return issues
	}

	if len(config.Workflow.Steps) == 0 {
		issues = append(issues, ValidationIssue{
			AgentType: agent.Type,
			Severity:  "warning",
			Category:  "empty_workflow",
			Message:   "Workflow has no steps defined",
		})
		return issues
	}

	// Build data flow map - track what each step produces and how it wraps
	dataFlow := buildDataFlowMap(config.Workflow)

	// Track produced fields
	producedFields := make(map[string]string)
	for stepName, step := range config.Workflow.Steps {
		if step.OutputField != "" {
			producedFields[step.OutputField] = stepName
		}
	}

	// Validate each step
	for stepName, step := range config.Workflow.Steps {
		stepIssues := validateStep(agent.Type, stepName, step, actionSpecs, producedFields, dataFlow)
		issues = append(issues, stepIssues...)
	}

	// Validate workflow flow
	flowIssues := validateStepFlow(agent.Type, config.Workflow)
	issues = append(issues, flowIssues...)

	// Validate contracts
	if agent.InputContract != nil {
		contractIssues := validateInputContract(agent, config.Workflow)
		issues = append(issues, contractIssues...)
	}

	outputIssues := validateOutputContract(agent, config.Workflow, producedFields)
	issues = append(issues, outputIssues...)

	// KEY VALIDATION: Check path references against data flow for wrapper issues
	pathIssues := validateDataPathReferences(agent.Type, config.Workflow, dataFlow)
	issues = append(issues, pathIssues...)

	return issues
}

// buildDataFlowMap creates a map of output_field -> DataFlowEntry
func buildDataFlowMap(workflow Workflow) map[string]DataFlowEntry {
	dataFlow := make(map[string]DataFlowEntry)

	for stepName, step := range workflow.Steps {
		if step.OutputField == "" {
			continue
		}

		entry := DataFlowEntry{
			StepName:    stepName,
			Action:      step.Action,
			OutputField: step.OutputField,
		}

		// Check if this action wraps its output
		if wrapper, wraps := actionWrappingBehavior[step.Action]; wraps {
			entry.WrapsIn = wrapper
		}

		dataFlow[step.OutputField] = entry
	}

	return dataFlow
}

// validateDataPathReferences checks all path references against known data flow
// THIS IS THE KEY FUNCTION that catches missing .response wrappers
func validateDataPathReferences(agentType string, workflow Workflow, dataFlow map[string]DataFlowEntry) []ValidationIssue {
	var issues []ValidationIssue

	for stepName, step := range workflow.Steps {
		// Extract all path references from this step's config
		paths := extractAllPaths(step.Config)

		// Check condition strings for conditionals
		if step.Action == "conditional" {
			if condition, ok := step.Config["condition"].(string); ok {
				conditionPaths := extractPathsFromCondition(condition)
				paths = append(paths, conditionPaths...)
			}
		}

		// Check template strings for paths like {{site_plan.validated_plan}}
		templatePaths := extractTemplatePathsFromConfig(step.Config)
		paths = append(paths, templatePaths...)

		for _, path := range paths {
			issue := checkPathAgainstDataFlow(agentType, stepName, path, dataFlow)
			if issue != nil {
				issues = append(issues, *issue)
			}
		}

		// Special check for loop iterate_over
		if step.Action == "loop" {
			if iterateOver, ok := step.Config["iterate_over"].(string); ok {
				issue := checkPathAgainstDataFlow(agentType, stepName, iterateOver, dataFlow)
				if issue != nil {
					issues = append(issues, *issue)
				}
			}
		}
	}

	return issues
}

// extractAllPaths extracts field path references from a config map
func extractAllPaths(config map[string]interface{}) []string {
	var paths []string

	for _, value := range config {
		switch v := value.(type) {
		case string:
			if looksLikePath(v) {
				paths = append(paths, v)
			}
		case map[string]interface{}:
			nested := extractAllPaths(v)
			paths = append(paths, nested...)
		case []interface{}:
			for _, item := range v {
				if str, ok := item.(string); ok && looksLikePath(str) {
					paths = append(paths, str)
				}
				if m, ok := item.(map[string]interface{}); ok {
					nested := extractAllPaths(m)
					paths = append(paths, nested...)
				}
			}
		}
	}

	return paths
}

// extractTemplatePathsFromConfig finds {{.path}} patterns in config values
func extractTemplatePathsFromConfig(config map[string]interface{}) []string {
	var paths []string
	configJSON, _ := json.Marshal(config)
	configStr := string(configJSON)

	// Find all {{.path}} or {{path}} patterns
	re := regexp.MustCompile(`\{\{\.?([a-zA-Z_][a-zA-Z0-9_.]*)\}\}`)
	matches := re.FindAllStringSubmatch(configStr, -1)
	for _, match := range matches {
		if len(match) > 1 && strings.Contains(match[1], ".") {
			paths = append(paths, match[1])
		}
	}

	return paths
}

// looksLikePath checks if a string looks like a field path reference
func looksLikePath(s string) bool {
	if !strings.Contains(s, ".") {
		return false
	}
	if strings.HasPrefix(s, "http") {
		return false
	}
	if strings.Contains(s, " ") && !strings.Contains(s, "{{") {
		return false
	}
	// Skip comparison operators (conditions handled separately)
	if strings.Contains(s, "==") || strings.Contains(s, "!=") {
		return false
	}
	return true
}

// extractPathsFromCondition extracts field paths from condition expressions
func extractPathsFromCondition(condition string) []string {
	var paths []string

	// Split by OR/AND
	parts := regexp.MustCompile(`\s+(?:OR|AND)\s+`).Split(condition, -1)

	for _, part := range parts {
		// Extract the left side of comparisons
		re := regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_.]*)\s*(?:==|!=|>=|<=|>|<)`)
		matches := re.FindAllStringSubmatch(part, -1)
		for _, match := range matches {
			if len(match) > 1 && strings.Contains(match[1], ".") {
				paths = append(paths, match[1])
			}
		}
	}

	return paths
}

// checkPathAgainstDataFlow verifies a path reference is valid given known data flow
func checkPathAgainstDataFlow(agentType, stepName, path string, dataFlow map[string]DataFlowEntry) *ValidationIssue {
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return nil
	}

	rootField := parts[0]

	// Check if we have data flow info for this root field
	entry, exists := dataFlow[rootField]
	if !exists {
		// Could be input_data, current_page (loop var), or other implicit fields
		return nil
	}

	// If this action wraps its output, check if the path includes the wrapper
	if entry.WrapsIn != "" {
		secondPart := parts[1]

		// Check if the path is missing the wrapper
		if secondPart != entry.WrapsIn {
			// This is the bug! Path references data without going through wrapper
			suggestedPath := rootField + "." + entry.WrapsIn + "." + strings.Join(parts[1:], ".")

			return &ValidationIssue{
				AgentType: agentType,
				StepName:  stepName,
				Severity:  "error",
				Category:  "missing_response_wrapper",
				Message: fmt.Sprintf(
					"Path '%s' references output of %s step '%s', but %s wraps responses under '.%s'. "+
						"Actual path should be '%s'",
					path, entry.Action, entry.StepName, entry.Action, entry.WrapsIn, suggestedPath,
				),
				Suggested: suggestedPath,
			}
		}
	}

	return nil
}

func validateStep(agentType, stepName string, step Step, actionSpecs map[string]ActionSpec, producedFields map[string]string, dataFlow map[string]DataFlowEntry) []ValidationIssue {
	var issues []ValidationIssue

	spec, hasSpec := actionSpecs[step.Action]
	if !hasSpec {
		issues = append(issues, ValidationIssue{
			AgentType: agentType,
			StepName:  stepName,
			Severity:  "warning",
			Category:  "unknown_action",
			Message:   fmt.Sprintf("Unknown action '%s' - cannot validate config", step.Action),
		})
		return issues
	}

	// Check required config
	for _, required := range spec.RequiredCfg {
		if _, exists := step.Config[required]; !exists {
			issues = append(issues, ValidationIssue{
				AgentType: agentType,
				StepName:  stepName,
				Severity:  "error",
				Category:  "missing_config",
				Message:   fmt.Sprintf("Action '%s' requires config '%s'", step.Action, required),
			})
		}
	}

	// Check for potentially problematic path patterns
	for key, value := range step.Config {
		if pathStr, ok := value.(string); ok {
			if strings.HasSuffix(pathStr, ".result") {
				issues = append(issues, ValidationIssue{
					AgentType: agentType,
					StepName:  stepName,
					Severity:  "info",
					Category:  "path_pattern",
					Message:   fmt.Sprintf("Config '%s' uses '.result' suffix (%s) - verify this matches actual output structure", key, pathStr),
				})
			}
		}
	}

	// Check call_agent input_fields
	if step.Action == "call_agent" {
		if inputFields, ok := step.Config["input_fields"].([]interface{}); ok {
			if len(inputFields) == 0 {
				issues = append(issues, ValidationIssue{
					AgentType: agentType,
					StepName:  stepName,
					Severity:  "warning",
					Category:  "call_agent",
					Message:   "call_agent has empty input_fields - child agent may not receive necessary data",
				})
			}
		}
	}

	// Check loop config
	if step.Action == "loop" {
		hasItemsField := step.Config["items_field"] != nil
		hasIterateOver := step.Config["iterate_over"] != nil
		if !hasItemsField && !hasIterateOver {
			issues = append(issues, ValidationIssue{
				AgentType: agentType,
				StepName:  stepName,
				Severity:  "error",
				Category:  "loop_config",
				Message:   "loop action requires 'items_field' or 'iterate_over' in config",
			})
		}
	}

	return issues
}

func validateInputContract(agent AgentDefinition, workflow Workflow) []ValidationIssue {
	var issues []ValidationIssue

	if agent.InputContract.Expects == nil {
		return issues
	}

	for expectedField := range agent.InputContract.Expects {
		found := false
		for _, step := range workflow.Steps {
			if fieldUsedInStep(expectedField, step) {
				found = true
				break
			}
		}
		if !found {
			issues = append(issues, ValidationIssue{
				AgentType: agent.Type,
				StepName:  "",
				Severity:  "warning",
				Category:  "unused_input",
				Message:   fmt.Sprintf("Input contract expects '%s' but it's not used in any workflow step", expectedField),
			})
		}
	}

	return issues
}

func validateOutputContract(agent AgentDefinition, workflow Workflow, producedFields map[string]string) []ValidationIssue {
	var issues []ValidationIssue

	for stepName, step := range workflow.Steps {
		if step.Action == "complete_workflow" {
			if outputFields, ok := step.Config["output_fields"].([]interface{}); ok {
				for _, field := range outputFields {
					if fieldStr, ok := field.(string); ok {
						if _, produced := producedFields[fieldStr]; !produced {
							issues = append(issues, ValidationIssue{
								AgentType: agent.Type,
								StepName:  stepName,
								Severity:  "warning",
								Category:  "missing_output",
								Message:   fmt.Sprintf("complete_workflow references '%s' but no step produces it", fieldStr),
							})
						}
					}
				}
			}
		}
	}

	return issues
}

func validateStepFlow(agentType string, workflow Workflow) []ValidationIssue {
	var issues []ValidationIssue

	for stepName, step := range workflow.Steps {
		if step.NextStep != "" {
			if _, exists := workflow.Steps[step.NextStep]; !exists {
				issues = append(issues, ValidationIssue{
					AgentType: agentType,
					StepName:  stepName,
					Severity:  "error",
					Category:  "broken_flow",
					Message:   fmt.Sprintf("next_step '%s' does not exist", step.NextStep),
				})
			}
		}

		if step.Action == "conditional" {
			if thenStep, ok := step.Config["then_step"].(string); ok {
				if _, exists := workflow.Steps[thenStep]; !exists {
					issues = append(issues, ValidationIssue{
						AgentType: agentType,
						StepName:  stepName,
						Severity:  "error",
						Category:  "broken_flow",
						Message:   fmt.Sprintf("then_step '%s' does not exist", thenStep),
					})
				}
			}
			if elseStep, ok := step.Config["else_step"].(string); ok {
				if _, exists := workflow.Steps[elseStep]; !exists {
					issues = append(issues, ValidationIssue{
						AgentType: agentType,
						StepName:  stepName,
						Severity:  "error",
						Category:  "broken_flow",
						Message:   fmt.Sprintf("else_step '%s' does not exist", elseStep),
					})
				}
			}
		}
	}

	if workflow.StartStep != "" {
		if _, exists := workflow.Steps[workflow.StartStep]; !exists {
			issues = append(issues, ValidationIssue{
				AgentType: agentType,
				StepName:  "",
				Severity:  "error",
				Category:  "broken_flow",
				Message:   fmt.Sprintf("start_step '%s' does not exist", workflow.StartStep),
			})
		}
	}

	return issues
}

func fieldUsedInStep(fieldPath string, step Step) bool {
	configJSON, _ := json.Marshal(step.Config)
	configStr := string(configJSON)
	basePath := strings.Split(fieldPath, ".")[0]
	return strings.Contains(configStr, basePath) || strings.Contains(configStr, fieldPath)
}

func reportIssues(issues []ValidationIssue, verbose bool) {
	if len(issues) == 0 {
		fmt.Println("\n✓ No issues found")
		return
	}

	errors := []ValidationIssue{}
	warnings := []ValidationIssue{}
	infos := []ValidationIssue{}

	for _, issue := range issues {
		switch issue.Severity {
		case "error":
			errors = append(errors, issue)
		case "warning":
			warnings = append(warnings, issue)
		case "info":
			infos = append(infos, issue)
		}
	}

	fmt.Printf("\n=== Validation Results ===\n")
	fmt.Printf("Errors: %d, Warnings: %d, Info: %d\n\n", len(errors), len(warnings), len(infos))

	if len(errors) > 0 {
		fmt.Println("❌ ERRORS:")
		for _, e := range errors {
			step := e.StepName
			if step == "" {
				step = "(workflow)"
			}
			fmt.Printf("  [%s] %s.%s:\n", e.Category, e.AgentType, step)
			fmt.Printf("      %s\n", e.Message)
			if e.Suggested != "" {
				fmt.Printf("      SUGGESTED FIX: %s\n", e.Suggested)
			}
		}
		fmt.Println()
	}

	if len(warnings) > 0 {
		fmt.Println("⚠️  WARNINGS:")
		for _, w := range warnings {
			step := w.StepName
			if step == "" {
				step = "(workflow)"
			}
			fmt.Printf("  [%s] %s.%s: %s\n", w.Category, w.AgentType, step, w.Message)
		}
		fmt.Println()
	}

	if verbose && len(infos) > 0 {
		fmt.Println("ℹ️  INFO:")
		for _, i := range infos {
			step := i.StepName
			if step == "" {
				step = "(workflow)"
			}
			fmt.Printf("  [%s] %s.%s: %s\n", i.Category, i.AgentType, step, i.Message)
		}
		fmt.Println()
	}

	if verbose {
		fmt.Println("\n--- JSON Output ---")
		jsonOut, _ := json.MarshalIndent(issues, "", "  ")
		fmt.Println(string(jsonOut))
	}
}
