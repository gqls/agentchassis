// Workflow Validator Tool
// Validates that workflow field references match contracts and action expectations
//
// Usage:
//   go run main.go -agents agents.json -actions action_specs.json
//
// Or with inline JSON:
//   go run main.go -agent-json '{"type":"pageflow-builder",...}'

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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
	Produces interface{} `json:"produces"` // can be string or map
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
	ReadsFrom   []string `json:"reads_from"`   // Config keys that specify input paths
	WritesTo    string   `json:"writes_to"`    // Usually "output_field"
	RequiredCfg []string `json:"required_cfg"` // Required config keys
	OptionalCfg []string `json:"optional_cfg"` // Optional config keys
}

// ValidationIssue represents a problem found
type ValidationIssue struct {
	AgentType string `json:"agent_type"`
	StepName  string `json:"step_name"`
	Severity  string `json:"severity"` // "error", "warning", "info"
	Category  string `json:"category"`
	Message   string `json:"message"`
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
		RequiredCfg: []string{"prompt_template"},
		OptionalCfg: []string{"input_fields", "output_format", "ai_service"},
	},
	"render_component": {
		Name:        "render_component",
		ReadsFrom:   []string{"component_from", "content_from", "context_from"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"component_function", "component_id", "component_from", "content_from", "context_from", "content_field"},
	},
	"git_commit": {
		Name:        "git_commit",
		ReadsFrom:   []string{"html_from", "content_field", "page_from", "site_id_from", "domain_field"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"html_from", "content_field", "page_from", "site_id_from", "domain_field", "files_field"},
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
	"update_site_defaults": {
		Name:        "update_site_defaults",
		ReadsFrom:   []string{"site_id_field", "defaults_field"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"site_id_field", "defaults_field", "defaults", "header_from", "footer_from"},
	},
	"update_page_status": {
		Name:        "update_page_status",
		ReadsFrom:   []string{"page_id_field", "page_id_from", "commit_from"},
		WritesTo:    "output_field",
		RequiredCfg: []string{"status"},
		OptionalCfg: []string{"page_id_field", "page_id_from", "commit_from", "site_id_field"},
	},
	"select_style_collection": {
		Name:        "select_style_collection",
		ReadsFrom:   []string{"site_id_field", "style_from", "domain_field"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"site_id_field", "style_collection_id", "style_from", "domain_field", "fallback_by_domain"},
	},
	"sync_pages_to_db": {
		Name:        "sync_pages_to_db",
		ReadsFrom:   []string{"input_fields"},
		WritesTo:    "output_field",
		RequiredCfg: []string{"input_fields"},
		OptionalCfg: []string{},
	},
	"get_pages_to_build": {
		Name:        "get_pages_to_build",
		ReadsFrom:   []string{"site_id_field"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"site_id_field", "build_statuses", "include_all"},
	},
	"ensure_site_record": {
		Name:        "ensure_site_record",
		ReadsFrom:   []string{"input_fields", "domain_field"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"input_fields", "domain_field", "store_brief_in_content_data"},
	},
	"loop": {
		Name:        "loop",
		ReadsFrom:   []string{"items_field", "iterate_over"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"items_field", "iterate_over", "item_variable", "loop_var", "max_iterations", "mode", "sub_workflow", "substeps"},
	},
	"assemble_page": {
		Name:        "assemble_page",
		ReadsFrom:   []string{"page_from", "content_from", "site_id_from", "content_field"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"page_from", "content_from", "site_id_from", "content_field", "add_navigation"},
	},
	"compile_page_sections": {
		Name:        "compile_page_sections",
		ReadsFrom:   []string{"sections_from", "sections_field", "page_from"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"sections_from", "sections_field", "page_from", "page_name", "include_research_ids"},
	},
	"build_render_context": {
		Name:        "build_render_context",
		ReadsFrom:   []string{"sources"},
		WritesTo:    "output_field",
		RequiredCfg: []string{"sources"},
		OptionalCfg: []string{},
	},
	"load_page_section_components": {
		Name:        "load_page_section_components",
		ReadsFrom:   []string{"sections_from", "page_from"},
		WritesTo:    "output_field",
		RequiredCfg: []string{},
		OptionalCfg: []string{"sections_from", "page_from", "include_templates", "include_input_schema"},
	},
	"store_asset": {
		Name:        "store_asset",
		ReadsFrom:   []string{"url_field", "site_id_field", "origin_prompt_field"},
		WritesTo:    "output_field",
		RequiredCfg: []string{"asset_type", "origin_type"},
		OptionalCfg: []string{"url_field", "site_id_field", "purpose", "brand_asset_key", "origin_prompt_field", "update_site_brand_assets"},
	},
	"conditional": {
		Name:        "conditional",
		ReadsFrom:   []string{"condition"},
		WritesTo:    "",
		RequiredCfg: []string{"condition", "then_step"},
		OptionalCfg: []string{"else_step"},
	},
	"complete_workflow": {
		Name:        "complete_workflow",
		ReadsFrom:   []string{"output_fields"},
		WritesTo:    "",
		RequiredCfg: []string{},
		OptionalCfg: []string{"output_fields", "output_field"},
	},
	"request_human_input": {
		Name:        "request_human_input",
		ReadsFrom:   []string{"data_field", "default_from"},
		WritesTo:    "output_field",
		RequiredCfg: []string{"request_type"},
		OptionalCfg: []string{"title", "message", "fields", "data_field", "skip_if", "timeout_seconds"},
	},
}

func main() {
	agentsFile := flag.String("agents", "", "Path to JSON file with agent definitions array")
	agentJSON := flag.String("agent-json", "", "Inline JSON for a single agent definition")
	actionsFile := flag.String("actions", "", "Path to JSON file with action specs (optional)")
	verbose := flag.Bool("verbose", false, "Show detailed output")
	flag.Parse()

	if *agentsFile == "" && *agentJSON == "" {
		fmt.Println("Usage: go run main.go -agents agents.json")
		fmt.Println("   or: go run main.go -agent-json '{...}'")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load action specs
	actionSpecs := defaultActionSpecs
	if *actionsFile != "" {
		data, err := os.ReadFile(*actionsFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading actions file: %v\n", err)
			os.Exit(1)
		}
		var customSpecs []ActionSpec
		if err := json.Unmarshal(data, &customSpecs); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing actions file: %v\n", err)
			os.Exit(1)
		}
		for _, spec := range customSpecs {
			actionSpecs[spec.Name] = spec
		}
	}

	// Load agent definitions
	var agents []AgentDefinition
	if *agentJSON != "" {
		var agent AgentDefinition
		if err := json.Unmarshal([]byte(*agentJSON), &agent); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing agent JSON: %v\n", err)
			os.Exit(1)
		}
		agents = append(agents, agent)
	} else {
		data, err := os.ReadFile(*agentsFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading agents file: %v\n", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(data, &agents); err != nil {
			// Try as single object
			var agent AgentDefinition
			if err := json.Unmarshal(data, &agent); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing agents file: %v\n", err)
				os.Exit(1)
			}
			agents = append(agents, agent)
		}
	}

	// Validate each agent
	var allIssues []ValidationIssue
	for _, agent := range agents {
		issues := validateAgent(agent, actionSpecs, *verbose)
		allIssues = append(allIssues, issues...)
	}

	// Report results
	reportIssues(allIssues, *verbose)
}

func validateAgent(agent AgentDefinition, actionSpecs map[string]ActionSpec, verbose bool) []ValidationIssue {
	var issues []ValidationIssue

	if verbose {
		fmt.Printf("\n=== Validating: %s ===\n", agent.Type)
	}

	// Parse workflow from default_config
	var config WorkflowConfig
	if err := json.Unmarshal(agent.DefaultConfig, &config); err != nil {
		issues = append(issues, ValidationIssue{
			AgentType: agent.Type,
			StepName:  "",
			Severity:  "error",
			Category:  "parse",
			Message:   fmt.Sprintf("Failed to parse default_config: %v", err),
		})
		return issues
	}

	if config.Workflow.Steps == nil {
		issues = append(issues, ValidationIssue{
			AgentType: agent.Type,
			StepName:  "",
			Severity:  "warning",
			Category:  "structure",
			Message:   "No workflow.steps found in default_config",
		})
		return issues
	}

	// Track what fields each step produces
	producedFields := make(map[string]string) // field -> step that produces it

	// Validate each step
	for stepName, step := range config.Workflow.Steps {
		stepIssues := validateStep(agent.Type, stepName, step, actionSpecs, producedFields, verbose)
		issues = append(issues, stepIssues...)

		// Record output field
		if step.OutputField != "" {
			producedFields[step.OutputField] = stepName
		}
	}

	// Validate input contract vs what workflow actually uses
	if agent.InputContract != nil {
		issues = append(issues, validateInputContract(agent, config.Workflow)...)
	}

	// Validate output contract vs what workflow produces
	if agent.OutputContract != nil {
		issues = append(issues, validateOutputContract(agent, config.Workflow, producedFields)...)
	}

	// Validate step flow (next_step references exist)
	issues = append(issues, validateStepFlow(agent.Type, config.Workflow)...)

	return issues
}

func validateStep(agentType, stepName string, step Step, actionSpecs map[string]ActionSpec, producedFields map[string]string, verbose bool) []ValidationIssue {
	var issues []ValidationIssue

	if verbose {
		fmt.Printf("  Step: %s (action: %s)\n", stepName, step.Action)
	}

	// Check if action is known
	spec, known := actionSpecs[step.Action]
	if !known {
		issues = append(issues, ValidationIssue{
			AgentType: agentType,
			StepName:  stepName,
			Severity:  "warning",
			Category:  "unknown_action",
			Message:   fmt.Sprintf("Action '%s' not in action specs - cannot validate config", step.Action),
		})
		return issues
	}

	// Check required config keys
	for _, req := range spec.RequiredCfg {
		if step.Config[req] == nil {
			issues = append(issues, ValidationIssue{
				AgentType: agentType,
				StepName:  stepName,
				Severity:  "error",
				Category:  "missing_config",
				Message:   fmt.Sprintf("Action '%s' requires config key '%s'", step.Action, req),
			})
		}
	}

	// Check field references in config
	for _, fieldKey := range spec.ReadsFrom {
		if fieldValue, ok := step.Config[fieldKey]; ok {
			// Check if it's a string path reference
			if pathStr, ok := fieldValue.(string); ok {
				// This is a field path - check if it references a produced field
				basePath := strings.Split(pathStr, ".")[0]
				if verbose {
					fmt.Printf("    Reads: %s = %s (base: %s)\n", fieldKey, pathStr, basePath)
				}
				// We could check if basePath is in producedFields or expected inputs
			}
			// Check if it's an array (input_fields)
			if fieldArray, ok := fieldValue.([]interface{}); ok {
				for _, f := range fieldArray {
					if pathStr, ok := f.(string); ok {
						basePath := strings.Split(pathStr, ".")[0]
						if verbose {
							fmt.Printf("    Reads: %s[] = %s (base: %s)\n", fieldKey, pathStr, basePath)
						}
					}
				}
			}
		}
	}

	// Check for common config mistakes
	issues = append(issues, checkCommonMistakes(agentType, stepName, step)...)

	return issues
}

func checkCommonMistakes(agentType, stepName string, step Step) []ValidationIssue {
	var issues []ValidationIssue

	// Check for .result suffix that might cause path issues
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
			if strings.HasSuffix(pathStr, ".response") {
				issues = append(issues, ValidationIssue{
					AgentType: agentType,
					StepName:  stepName,
					Severity:  "info",
					Category:  "path_pattern",
					Message:   fmt.Sprintf("Config '%s' uses '.response' suffix (%s) - verify this matches actual output structure", key, pathStr),
				})
			}
		}
	}

	// Check call_agent input_fields vs target agent's input_contract
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

	// Check if expected inputs are actually used somewhere in the workflow
	for expectedField := range agent.InputContract.Expects {
		found := false
		for stepName, step := range workflow.Steps {
			if fieldUsedInStep(expectedField, step) {
				found = true
				_ = stepName
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

	// Check if complete_workflow step references fields that are produced
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
								Message:   fmt.Sprintf("complete_workflow references '%s' but no step produces it (output_field)", fieldStr),
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

		// Check conditional then_step and else_step
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

	// Check start_step exists
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
	// Check config values for field references
	configJSON, _ := json.Marshal(step.Config)
	configStr := string(configJSON)

	// Simple check - field path appears in config
	basePath := strings.Split(fieldPath, ".")[0]
	return strings.Contains(configStr, basePath) || strings.Contains(configStr, fieldPath)
}

func reportIssues(issues []ValidationIssue, verbose bool) {
	if len(issues) == 0 {
		fmt.Println("\n✓ No issues found")
		return
	}

	// Group by severity
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
			fmt.Printf("  [%s] %s.%s: %s\n", e.Category, e.AgentType, step, e.Message)
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

	// Output as JSON for further processing
	if verbose {
		fmt.Println("\n--- JSON Output ---")
		jsonOut, _ := json.MarshalIndent(issues, "", "  ")
		fmt.Println(string(jsonOut))
	}
}
