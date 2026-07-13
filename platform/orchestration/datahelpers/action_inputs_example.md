// FILE: Example of using ActionInputSpec in an action
// This shows the before/after comparison

// ============================================================
// BEFORE: Every action has this boilerplate
// ============================================================

func RerenderSinglePageAction_OLD(ctx context.Context, params ActionParams) (interface{}, error) {
config := params.StepConfig.Config

	// Boilerplate: Parse input_fields from config
	inputFields := []string{"page_id", "site_id", "domain"}
	if fields, ok := config["input_fields"].([]interface{}); ok {
		inputFields = make([]string, len(fields))
		for i, f := range fields {
			inputFields[i], _ = f.(string)
		}
	}

	// Boilerplate: Call ExtractFields
	extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	// Boilerplate: Try flat field, then nested (backward compat)
	var pageIDStr string
	if s, ok := extracted["page_id"].(string); ok && s != "" {
		pageIDStr = s
	} else if currentPage, ok := extracted["current_page"].(map[string]interface{}); ok {
		pageIDStr, _ = currentPage["page_id"].(string)
	}
	if pageIDStr == "" {
		return nil, fmt.Errorf("page_id not found in input")
	}

	// Boilerplate: Same for site_id...
	var siteIDStr string
	if s, ok := extracted["site_id"].(string); ok && s != "" {
		siteIDStr = s
	} else if rerenderPages, ok := extracted["rerender_pages"].(map[string]interface{}); ok {
		siteIDStr, _ = rerenderPages["site_id"].(string)
	}
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found in input")
	}

	// ... 50+ lines of extraction before real work starts
}

// ============================================================
// AFTER: Using ActionInputSpec
// ============================================================

// Define the spec once (can be at package level or in init())
var RerenderSinglePageInputSpec = datahelpers.ActionInputSpec{
                                Required: []string{"page_id", "site_id"},
                                Optional: []string{"domain", "max_nav_items"},
                                Defaults: map[string]interface{}{
                                "max_nav_items": 6,
                            },
// Map old config patterns to new field names
Deprecated: map[string]string{
                "page_id_field": "page_id",
                "site_id_field": "site_id",
                "domain_field":  "domain",
                },
}

func init() {
// Register for documentation/contract generation
datahelpers.RegisterActionInputSpec("rerender_single_page", RerenderSinglePageInputSpec)
}

func RerenderSinglePageAction_NEW(ctx context.Context, params ActionParams) (interface{}, error) {
// One call extracts everything, handles all patterns, validates required fields
inputs, err := datahelpers.ExtractActionInputs(
            params.CollectedData,
            params.StepConfig.Config,
            RerenderSinglePageInputSpec,
            params.Logger,
            )
if err != nil {
return nil, fmt.Errorf("input extraction failed: %w", err)
}

	// Clean access to values
	pageIDStr := inputs.Get("page_id")
	siteIDStr := inputs.Get("site_id")
	domain := inputs.Get("domain")
	maxNavItems := inputs.GetInt("max_nav_items", 6)

	// Parse UUIDs
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid page_id: %w", err)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// ... actual business logic starts here, ~40 lines saved
}

// ============================================================
// WORKFLOW CONFIG: Same as before, but simpler
// ============================================================

/*
Workflow step (preferred pattern):
{
"action": "rerender_single_page",
"config": {
"input_fields": ["page_id", "site_id", "domain"]
}
}

Workflow step (deprecated but still works):
{
"action": "rerender_single_page",
"config": {
"page_id_field": "current_page.page_id",
"site_id_field": "rerender_pages.site_id"
}
}

Both work, but deprecated pattern logs a warning:
WARN: Using deprecated config pattern
deprecated_key=page_id_field
path=current_page.page_id
use_instead=input_fields: ["page_id"]
*/

// ============================================================
// CALL_AGENT: input_mapping unchanged
// ============================================================

/*
The call_agent action uses input_mapping to prepare data for the child agent.
This is a different layer - it maps caller's data to child's input_data.

{
"action": "call_agent",
"config": {
"agent_type": "page-rerender",
"input_mapping": {
"page_id": "current_page.page_id",
"site_id": "rerender_pages.site_id",
"domain": "rerender_pages.domain"
}
}
}

This results in the child agent receiving:
{
"input_data": {
"page_id": "uuid-123",
"site_id": "uuid-456",
"domain": "example.com"
}
}

The child's action then uses input_fields to extract from input_data.
*/

// ============================================================
// LAYERED RESPONSIBILITY
// ============================================================

/*
Layer 1: call_agent with input_mapping
- Caller's responsibility
- Maps caller's data paths to child's input_data keys
- "I'm passing these values to the child"

Layer 2: Action with input_fields
- Action's responsibility
- Declares what fields the action needs
- "I need these values to do my work"
- Uses ExtractActionInputs() which handles:
    * Looking in input_data
    * Looking in CollectedData directly
    * Deprecated *_field patterns (with warnings)
    * Required field validation

Layer 3: Deprecated *_field patterns
- Legacy support only
- Will log warnings
- Will be removed in future version
  */