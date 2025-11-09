https://claude.ai/chat/c75db769-ab82-45f5-a4ed-8083bedc09ce
058_one_page_website

problem statement and background information
Agent Orchestration Architecture
Problem Statement

I would like to closely log and track the creation of agents, the messages (with headers and body) that are passed between agents.

Image creation is now working.

I'd like to now create a website with images and a separate about page and contact page. The images can be links in the html to the presigned urls.
The about page can be something like "this site was created by agents ... " etc or something along those lines.
the things we need include:

an agent group definition in the database with a workflow that can create the website.

agent definitions for each agent we use or create with their own workflows.

We can use the image adapter to create presigned urls of images using GenerateImageAction as we have been doing.

I'd like to break it down into several sections, with separate agents concentrating on each section. We can do detailed prompts for each section and we can, if we want to, spawn sub agents for e.g. research before writing the content.

We're using groups of workflows for the tasks. Agents should always respond to the responses topic of the caller and not to their own responses topic.

Please always work small steps at a time and think hard.

Please keep in mind the flow of work at all times and phrase your responses in this context including the function calls that lead up to the problem sometimes, so I can understand what's going on.

Please don't jump to conclusions.

Please don't create summary documents unless asked for, instead, moderately brief summaries at the end as you usually do will suffice.

Please don't say that any fixes are the final or last fixes. Don't use the words perfect, critical or excellent.

Please don't congratulate me but keep everything pragmatic and focused on the continuous debugging tasks we have ahead of us.

Every agent is an orchestrator.

Always look through existing code and see if we have similar functions or structs that already exist and that can be improved before creating new code.

Let's reuse (and alter if necessary) the functions and architecture that we already have before recreating similar structures. Please think hard about this every time.

Here is the new functionality (data_helpers.go)
// internal/backend/agent-chassis/platform/orchestration/data_helpers.go
package orchestration

import (
"fmt"
"os"
"strings"

	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// ============================================================================
// DATA HELPERS - Working with existing types.RequestMessage and types.ResponseMessage
// ============================================================================

// ============================================================================
// EXTRACTION FUNCTIONS - Get clean data from any message format
// ============================================================================

// ExtractDataFromMessage extracts clean data from ANY message format
// Works with types.RequestMessage, types.ResponseMessage, or raw maps
func ExtractDataFromMessage(source interface{}, logger *zap.Logger) map[string]interface{} {
if source == nil {
logger.Debug("ExtractDataFromMessage: source is nil")
return map[string]interface{}{}
}

	// Handle typed messages first
	switch msg := source.(type) {
	case *types.RequestMessage:
		return extractFromRequestMessage(msg, logger)
	case types.RequestMessage:
		return extractFromRequestMessage(&msg, logger)
	case *types.ResponseMessage:
		return extractFromResponseMessage(msg, logger)
	case types.ResponseMessage:
		return extractFromResponseMessage(&msg, logger)
	case map[string]interface{}:
		return extractFromRawMap(msg, logger)
	default:
		logger.Warn("ExtractDataFromMessage: unknown source type",
			zap.String("type", fmt.Sprintf("%T", source)))
		return map[string]interface{}{}
	}
}

// extractFromRequestMessage extracts data from a typed RequestMessage
func extractFromRequestMessage(msg *types.RequestMessage, logger *zap.Logger) map[string]interface{} {
if msg == nil || msg.Body == nil {
return map[string]interface{}{}
}

	// Body could be a map or another type
	switch body := msg.Body.(type) {
	case map[string]interface{}:
		// Look for data field first
		if data, ok := body["data"].(map[string]interface{}); ok {
			logger.Debug("extractFromRequestMessage: found body.data")
			return data
		}
		// Look for input_data field
		if data, ok := body["input_data"].(map[string]interface{}); ok {
			logger.Debug("extractFromRequestMessage: found body.input_data")
			return data
		}
		// Return cleaned body
		return cleanDataMap(body)
	default:
		logger.Debug("extractFromRequestMessage: body is not a map")
		return map[string]interface{}{}
	}
}

// extractFromResponseMessage extracts data from a typed ResponseMessage
func extractFromResponseMessage(msg *types.ResponseMessage, logger *zap.Logger) map[string]interface{} {
if msg == nil {
return map[string]interface{}{}
}

	// ResponseBody is a struct, not a pointer, so we work with it directly
	responseBody := msg.Body

	// Check if the Body field within ResponseBody has content
	if responseBody.Body == nil && responseBody.Error == nil {
		logger.Debug("extractFromResponseMessage: empty response body")
		return map[string]interface{}{}
	}

	// ResponseBody.Body contains the actual data
	if body, ok := responseBody.Body.(map[string]interface{}); ok {
		// Look for data field
		if data, ok := body["data"].(map[string]interface{}); ok {
			logger.Debug("extractFromResponseMessage: found body.data")
			return data
		}
		// Return the body itself if it's clean data
		logger.Debug("extractFromResponseMessage: using cleaned body")
		return cleanDataMap(body)
	}

	logger.Debug("extractFromResponseMessage: body is not a map")
	return map[string]interface{}{}
}

// extractFromRawMap handles raw map extraction (backward compatibility)
func extractFromRawMap(source map[string]interface{}, logger *zap.Logger) map[string]interface{} {
// Try different extraction strategies in order of preference

	// Strategy 1: Look for body.data (clean format)
	if data := extractFromPath(source, "body.data"); data != nil {
		logger.Debug("extractFromRawMap: found body.data")
		return data
	}

	// Strategy 2: Look for body.input_data (current format)
	if data := extractFromPath(source, "body.input_data"); data != nil {
		logger.Debug("extractFromRawMap: found body.input_data")
		return data
	}

	// Strategy 3: Look for input_data at any level
	if data := findDataField(source, "input_data"); data != nil {
		logger.Debug("extractFromRawMap: found input_data")
		return data
	}

	// Strategy 4: Look for data field at any level
	if data := findDataField(source, "data"); data != nil {
		logger.Debug("extractFromRawMap: found data")
		return data
	}

	// Strategy 5: Clean the body itself
	if body, ok := source["body"].(map[string]interface{}); ok {
		cleaned := cleanDataMap(body)
		if len(cleaned) > 0 {
			logger.Debug("extractFromRawMap: using cleaned body")
			return cleaned
		}
	}

	logger.Debug("extractFromRawMap: no data found")
	return map[string]interface{}{}
}

// ============================================================================
// MESSAGE BUILDING FUNCTIONS - Using existing types
// ============================================================================

// BuildRequestMessage creates a types.RequestMessage for sending to another agent
func BuildRequestMessage(
execCtx *types.ExecutionContext,
toAgentType string,
action string,
data map[string]interface{},
config map[string]interface{},
logger *zap.Logger,
) *types.RequestMessage {

	// Update context for this request
	execCtx.Action = action
	execCtx.ToAgentType = toAgentType

	// Build headers using existing method
	headers := execCtx.ToRequestHeaders()

	// Build body
	body := map[string]interface{}{
		"action": action,
		"data":   data,
	}

	if config != nil && len(config) > 0 {
		body["config"] = config
	}

	logger.Debug("BuildRequestMessage: created message",
		zap.String("action", action),
		zap.String("to_agent_type", toAgentType),
		zap.Int("data_fields", len(data)))

	return &types.RequestMessage{
		Headers: headers,
		Body:    body,
	}
}

// BuildResponseMessage creates a types.ResponseMessage for responding to parent
func BuildResponseMessage(
execCtx *types.ExecutionContext,
success bool,
responseData map[string]interface{},
errorInfo *types.ErrorInfo,
logger *zap.Logger,
) *types.ResponseMessage {

	// Create response context
	responseCtx := execCtx.CreateResponseContext(
		determineStatus(success, errorInfo),
		100, // fuel used - calculate properly in production
	)

	// Build headers using existing method
	headers := responseCtx.ToResponseHeaders()

	// Build body
	responseBody := types.ResponseBody{
		Success: success,
		Body:    map[string]interface{}{"data": responseData},
		Error:   errorInfo,
	}

	logger.Debug("BuildResponseMessage: created response",
		zap.Bool("success", success),
		zap.Int("data_fields", len(responseData)))

	return &types.ResponseMessage{
		Headers: headers,
		Body:    responseBody,
	}
}

// BuildInitializationRequest creates an initialization request for a spawned agent
func BuildInitializationRequest(
parentCtx *types.ExecutionContext,
childAgentType string,
role string,
initialData map[string]interface{},
agentConfig map[string]interface{},
logger *zap.Logger,
) *types.RequestMessage {

	// Create child context
	childCtx := parentCtx.CreateChildContext(childAgentType)
	childCtx.Action = "initialize"

	// Set functional role if specified
	if role != "" {
		childCtx.FunctionalRole = role
	}

	// Build headers
	headers := childCtx.ToRequestHeaders()

	// Build initialization body
	body := map[string]interface{}{
		"action":            "initialize",
		"is_initialization": true,
		"agent_info": map[string]interface{}{
			"agent_id":   childCtx.OrchestrationID,
			"agent_type": childAgentType,
			"agent_name": childCtx.OrchestrationName,
		},
		"data": initialData,
	}

	if role != "" {
		body["role"] = role
	}

	if agentConfig != nil {
		body["config"] = agentConfig
	}

	logger.Debug("BuildInitializationRequest: created",
		zap.String("child_type", childAgentType),
		zap.String("child_orch_id", childCtx.OrchestrationID))

	return &types.RequestMessage{
		Headers: headers,
		Body:    body,
	}
}

// ============================================================================
// COLLECTED DATA MANAGEMENT - Works with ExecutionContext
// ============================================================================

// BuildCollectedData builds CollectedData from an incoming message
func BuildCollectedData(
message interface{},
execCtx *types.ExecutionContext,
logger *zap.Logger,
) map[string]interface{} {

	collected := map[string]interface{}{
		"__execution_context__": execCtx,
	}

	// Extract data based on message type
	data := ExtractDataFromMessage(message, logger)
	if len(data) > 0 {
		collected["input_data"] = data
	}

	// Add action from context
	if execCtx.Action != "" {
		collected["action"] = execCtx.Action
	}

	// Extract config if it's a raw message with config
	if rawMsg, ok := message.(map[string]interface{}); ok {
		if config := extractSystemConfig(rawMsg); config != nil {
			collected["config"] = config
		}
		// Store raw for debugging
		collected["__raw_message__"] = rawMsg
	}

	// Add routing information
	collected["__my_requests_topic__"] = execCtx.RequestsTopic
	collected["__my_responses_topic__"] = execCtx.ResponsesTopic

	if execCtx.ReplyToTopic != "" {
		collected["__parent_responses_topic__"] = execCtx.ReplyToTopic
	}

	logger.Debug("BuildCollectedData: built structure",
		zap.Int("data_fields", len(data)),
		zap.String("action", execCtx.Action))

	return collected
}

// ============================================================================
// ORIGINAL FUNCTIONS - Maintained for backward compatibility
// ============================================================================

// NormalizeInputData extracts clean input_data from any source structure
func NormalizeInputData(source map[string]interface{}, logger *zap.Logger) map[string]interface{} {
logger.Info("NormalizeInputData: source:",
zap.Any("source is:", source))

	if source == nil {
		logger.Debug("NormalizeInputData: source is nil, returning empty map")
		return map[string]interface{}{}
	}

	return ExtractDataFromMessage(source, logger)
}

// NormalizeCollectedData builds a clean CollectedData structure from a message
func NormalizeCollectedData(
message map[string]interface{},
execCtx *types.ExecutionContext,
requestsTopic string,
logger *zap.Logger,
) map[string]interface{} {

	logger.Debug("NormalizeCollectedData: building clean structure")

	// Use the new builder
	collectedData := BuildCollectedData(message, execCtx, logger)

	// Override requests topic if provided
	if requestsTopic != "" {
		collectedData["__my_requests_topic__"] = requestsTopic
	}

	// Add parent responses topic from environment if needed
	if execCtx.ReplyToTopic == "" {
		if parentTopic := os.Getenv("PARENT_RESPONSES_TOPIC"); parentTopic != "" {
			execCtx.ReplyToTopic = parentTopic
			collectedData["__parent_responses_topic__"] = parentTopic
		}
	}

	// Extract additional fields from message if needed
	if agentConfig, ok := message["agent_config"].(map[string]interface{}); ok {
		collectedData["agent_config"] = agentConfig
	}

	if agentGroup, ok := message["agent_group"].(map[string]interface{}); ok {
		collectedData["agent_group"] = agentGroup
	}

	if agentType, ok := message["agent_type"].(string); ok {
		collectedData["agent_type"] = agentType
	}

	if prompt, ok := message["prompt"].(string); ok {
		collectedData["prompt"] = prompt
	}

	logger.Debug("NormalizeCollectedData: structure complete",
		zap.Int("fields", len(collectedData)))

	return collectedData
}

// NormalizeResponseData cleans response data before storing in CollectedData
func NormalizeResponseData(responseBody types.ResponseBody, logger *zap.Logger) map[string]interface{} {
if responseBody.Body == nil && responseBody.Error == nil {
logger.Debug("NormalizeResponseData: responseBody is nil")
return map[string]interface{}{}
}

	// ResponseBody.Body contains the actual data
	if body, ok := responseBody.Body.(map[string]interface{}); ok {
		if data, ok := body["data"].(map[string]interface{}); ok {
			return data
		}
		return cleanDataMap(body)
	}

	return map[string]interface{}{}
}

// GetInputData safely retrieves input_data from CollectedData
func GetInputData(collectedData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
if collectedData == nil {
logger.Warn("GetInputData: collectedData is nil")
return map[string]interface{}{}
}

	// First try the normalized location
	if data, ok := collectedData["input_data"].(map[string]interface{}); ok {
		return data
	}

	// Fallback to extraction
	return ExtractDataFromMessage(collectedData, logger)
}

// GetStepData safely retrieves data from a specific step's response
func GetStepData(collectedData map[string]interface{}, stepName string, logger *zap.Logger) (map[string]interface{}, bool) {
if collectedData == nil {
logger.Warn("GetStepData: collectedData is nil", zap.String("step", stepName))
return nil, false
}

	if stepData, ok := collectedData[stepName].(map[string]interface{}); ok {
		logger.Debug("GetStepData: found step data", zap.String("step", stepName))
		return stepData, true
	}

	logger.Debug("GetStepData: step data not found", zap.String("step", stepName))
	return nil, false
}

// GetMultipleStepData retrieves data from multiple steps
func GetMultipleStepData(collectedData map[string]interface{}, stepNames []string, logger *zap.Logger) map[string]interface{} {
if collectedData == nil {
logger.Warn("GetMultipleStepData: collectedData is nil")
return map[string]interface{}{}
}

	result := make(map[string]interface{})

	for _, stepName := range stepNames {
		if stepData, ok := collectedData[stepName].(map[string]interface{}); ok {
			result[stepName] = stepData
			logger.Debug("GetMultipleStepData: collected step data", zap.String("step", stepName))
		}
	}

	logger.Debug("GetMultipleStepData: collection complete",
		zap.Int("requested", len(stepNames)),
		zap.Int("found", len(result)))

	return result
}

// GetFieldFromPath retrieves a value using dot-notation path
func GetFieldFromPath(collectedData map[string]interface{}, path string, logger *zap.Logger) (interface{}, error) {
if collectedData == nil {
return nil, fmt.Errorf("collectedData is nil")
}

	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}

	val := getFieldByPath(collectedData, path)
	if val == nil {
		return nil, fmt.Errorf("field not found: %s", path)
	}

	logger.Debug("GetFieldFromPath: found value", zap.String("path", path))
	return val, nil
}

// GetFieldFromPathWithDefault retrieves a value with a fallback default
func GetFieldFromPathWithDefault(collectedData map[string]interface{}, path string, defaultValue interface{}, logger *zap.Logger) interface{} {
val, err := GetFieldFromPath(collectedData, path, logger)
if err != nil {
logger.Debug("GetFieldFromPathWithDefault: using default",
zap.String("path", path),
zap.Error(err))
return defaultValue
}
return val
}

// MergeInputData merges new input data with existing
func MergeInputData(existing, new map[string]interface{}, logger *zap.Logger) map[string]interface{} {
result := make(map[string]interface{})

	// Copy existing
	for k, v := range existing {
		result[k] = v
	}

	// Overlay new (overwrites)
	for k, v := range new {
		result[k] = v
	}

	logger.Debug("MergeInputData: merged data",
		zap.Int("existing_fields", len(existing)),
		zap.Int("new_fields", len(new)),
		zap.Int("result_fields", len(result)))

	return result
}

// UpdateCollectedData safely updates CollectedData with new information
func UpdateCollectedData(
collected map[string]interface{},
stepName string,
stepData map[string]interface{},
logger *zap.Logger,
) {
if collected == nil || stepData == nil {
return
}

	// Store step results
	collected[stepName] = stepData

	// Also merge into input_data if it's a data update
	if currentData, ok := collected["input_data"].(map[string]interface{}); ok {
		// Only merge actual data fields, not system fields
		cleanStepData := cleanDataMap(stepData)
		for k, v := range cleanStepData {
			currentData[k] = v
		}
		logger.Debug("UpdateCollectedData: merged step data",
			zap.String("step", stepName),
			zap.Int("merged_fields", len(cleanStepData)))
	}
}

// ============================================================================
// TRANSFORMATION FUNCTIONS
// ============================================================================

// TransformDataForAction prepares data for a specific action (dynamic)
func TransformDataForAction(
sourceData map[string]interface{},
transformSpec map[string]interface{},
logger *zap.Logger,
) map[string]interface{} {

	if transformSpec == nil {
		return sourceData
	}

	result := make(map[string]interface{})

	// Apply field mappings if specified
	if mappings, ok := transformSpec["field_mappings"].(map[string]interface{}); ok {
		for targetField, sourceField := range mappings {
			if sourceFieldStr, ok := sourceField.(string); ok {
				if val := getFieldByPath(sourceData, sourceFieldStr); val != nil {
					result[targetField] = val
				}
			}
		}
	}

	// Apply field filters if specified
	if fields, ok := transformSpec["include_fields"].([]interface{}); ok {
		for _, field := range fields {
			if fieldStr, ok := field.(string); ok {
				if val, exists := sourceData[fieldStr]; exists {
					result[fieldStr] = val
				}
			}
		}
	}

	// If no specific transformation, use source data
	if len(result) == 0 {
		return sourceData
	}

	// Add any additional fields specified
	if additions, ok := transformSpec["add_fields"].(map[string]interface{}); ok {
		for k, v := range additions {
			result[k] = v
		}
	}

	logger.Debug("TransformDataForAction: applied transformation",
		zap.Int("source_fields", len(sourceData)),
		zap.Int("result_fields", len(result)))

	return result
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// extractFromPath extracts data from a nested path like "body.data"
func extractFromPath(source map[string]interface{}, path string) map[string]interface{} {
current := source
segments := strings.Split(path, ".")

	for _, segment := range segments {
		if next, ok := current[segment].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}

	return current
}

// findDataField recursively finds a data field in the structure
func findDataField(source map[string]interface{}, fieldName string) map[string]interface{} {
// Check current level
if data, ok := source[fieldName].(map[string]interface{}); ok {
// Make sure it's actually data, not a wrapper
if !hasSystemFields(data) {
return data
}
// It might be nested, recurse
if nested := findDataField(data, fieldName); nested != nil {
return nested
}
}

	// Check one level deeper
	for key, value := range source {
		if key == "__raw_message__" || key == "__execution_context__" {
			continue
		}
		if nested, ok := value.(map[string]interface{}); ok {
			if result := findDataField(nested, fieldName); result != nil {
				return result
			}
		}
	}

	return nil
}

// cleanDataMap removes system fields from a map
func cleanDataMap(source map[string]interface{}) map[string]interface{} {
result := make(map[string]interface{})

	systemFields := map[string]bool{
		"action":                     true,
		"config":                     true,
		"agent_config":               true,
		"agent_group":                true,
		"__execution_context__":      true,
		"__raw_message__":            true,
		"__my_requests_topic__":      true,
		"__parent_responses_topic__": true,
		"headers":                    true,
		"is_initialization":          true,
		"agent_info":                 true,
		"role":                       true,
	}

	for k, v := range source {
		if !systemFields[k] {
			result[k] = v
		}
	}

	return result
}

// hasSystemFields checks if a map contains system fields
func hasSystemFields(data map[string]interface{}) bool {
systemFields := []string{"action", "config", "agent_config", "__execution_context__"}
for _, field := range systemFields {
if _, exists := data[field]; exists {
return true
}
}
return false
}

// extractAction finds the action from various possible locations
func extractAction(message map[string]interface{}) string {
// Try body.action first
if body, ok := message["body"].(map[string]interface{}); ok {
if action, ok := body["action"].(string); ok {
return action
}
}

	// Try top-level action
	if action, ok := message["action"].(string); ok {
		return action
	}

	// Try headers.action
	if headers, ok := message["headers"].(map[string]interface{}); ok {
		if action, ok := headers["action"].(string); ok {
			return action
		}
	}

	return ""
}

// extractSystemConfig extracts workflow/system configuration
func extractSystemConfig(message map[string]interface{}) map[string]interface{} {
// Try body.config first
if body, ok := message["body"].(map[string]interface{}); ok {
if config, ok := body["config"].(map[string]interface{}); ok {
if isSystemConfig(config) {
return config
}
}
}

	// Try top-level config
	if config, ok := message["config"].(map[string]interface{}); ok {
		if isSystemConfig(config) {
			return config
		}
	}

	return nil
}

// isSystemConfig checks if config is system/workflow config vs user data
func isSystemConfig(config map[string]interface{}) bool {
systemKeys := []string{"workflow", "group_type", "processing_mode", "ai_service"}
for _, key := range systemKeys {
if _, exists := config[key]; exists {
return true
}
}
return false
}

// getFieldByPath retrieves a field using dot notation
func getFieldByPath(data map[string]interface{}, path string) interface{} {
segments := strings.Split(path, ".")
var current interface{} = data

	for _, segment := range segments {
		if currentMap, ok := current.(map[string]interface{}); ok {
			current = currentMap[segment]
		} else {
			return nil
		}
	}

	return current
}

// determineStatus converts success/error info to status string
func determineStatus(success bool, errorInfo *types.ErrorInfo) string {
if success {
return "complete"
}
if errorInfo != nil && errorInfo.Recoverable {
return "error_recoverable"
}
return "error_unrecoverable"
}

--

previous initial message example:
=== run === robot hands website =====

# Generate IDs
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="robot-site-$(date +%H%M%S)"
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)
STEP_NAME="client_step_website_request"
CLIENT_ID="demo_client"

echo "========================================="
echo "Robot Hands Website Test"
echo "========================================="
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Request ID:       $REQUEST_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Name:             $ORCHESTRATION_NAME"
echo "  Client            $CLIENT_ID"
echo "  Message ID:       $MESSAGE_ID"
echo "  Agent ID:         $AGENT_ID"
echo "  Time:             $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================="

# Send message

kubectl -n kafka run -i --rm kcat-producer \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H orchestration_name=$ORCHESTRATION_NAME \
-H step_name=$STEP_NAME \
-H client_id=$CLIENT_ID \
-H message_type=request \
-H action=orchestrate \
-H from_agent_type=user \
-H from_agent_id=$AGENT_ID \
-H responses_topic=system.generic.responses <<EOF
{"action":"orchestrate","config":{"group_type":"robot-hands-website"},"input_data":{"business_type":"robotics automation company","business_name":"PrecisionBot Systems"}}
EOF

echo ""
echo "Message sent!"
echo ""
echo "Monitor:"
echo "  Orchestrator: kubectl logs -f deployment/agent-chassis -n agent-system | grep '$ORCHESTRATION_ID'"
echo "  Database: psql ... -c \"SELECT status, current_step FROM orchestration_states WHERE orchestration_id = '$ORCHESTRATION_ID';\""
echo ""
--

example agent group definition for robot hands website

                  id                  |        name         |     group_type      |                                                                agent_configs                                                                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               orchestration_workflow                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | usage_count | version |         created_at         |         updated_at         
--------------------------------------+---------------------+---------------------+----------------------------------------------------------------------------------------------------------------------------------------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+-------------+---------+----------------------------+----------------------------
d03fab55-b583-4228-907a-7b6040098dba | Robot Hands Website | robot-hands-website | [{"role": "hero_writer", "agent_type": "content-creator-hero-without-research"}, {"role": "image_creator", "agent_type": "image-generator"}] | {"steps": {"complete": {"action": "complete_workflow", "description": "Complete workflow"}, "generate_hero": {"action": "call_agent", "config": {"prompt": "Write a compelling hero section for {{.business_name}}, a {{.business_type}}. Focus on precision robotics and automation.", "agent_type": "content-creator-hero-without-research", "target_role": "hero_writer"}, "next_step": "generate_hero_image", "description": "Generate hero content"}, "spawn_hero_writer": {"action": "spawn_agent", "config": {"role": "hero_writer", "agent_type": "content-creator-hero-without-research"}, "next_step": "spawn_image_creator", "description": "Spawn hero content writer"}, "generate_hero_image": {"action": "call_agent", "config": {"width": 1920, "height": 1080, "prompt": "Professional photograph of precision robotic hands assembling electronic components, modern factory setting, dramatic lighting, photorealistic, 8k", "agent_type": "image-generator", "target_role": "image_creator"}, "next_step": "complete", "description": "Generate hero image"}, "spawn_image_creator": {"action": "spawn_agent", "config": {"role": "image_creator", "agent_type": "image-generator"}, "next_step": "generate_hero", "description": "Spawn image generator"}}, "start_step": "spawn_hero_writer"} |           0 |       1 | 2025-10-27 19:19:32.489062 | 2025-10-30 14:21:43.290942
(1 row)


--
example agent group definition for website builder

                  id                  |             name              |          group_type           |                                                                                                                                                                      agent_configs                                                                                                                                                                      |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    orchestration_workflow                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | usage_count | version |         created_at         |         updated_at      
3304b7b2-bc89-46f2-8859-b2060bbbd948 | Multi-Section Website Builder | multi-section-website-builder | [{"role": "hero_writer", "agent_type": "content-creator-hero"}, {"role": "features_writer", "agent_type": "content-creator-features"}, {"role": "testimonials_writer", "agent_type": "content-creator-testimonials"}, {"role": "contact_writer", "agent_type": "content-creator-contact"}, {"role": "cta_writer", "agent_type": "content-creator-cta"}] | {"steps": {"complete": {"action": "complete_workflow", "description": "Return complete website content"}, "generate_cta": {"action": "call_agent", "config": {"prompt": "Write a strong call-to-action section for {{.business_name}}. Include action text and a reason to act now.", "agent_type": "content-creator-cta", "target_role": "cta_writer"}, "next_step": "aggregate_all_sections", "description": "Generate CTA section"}, "generate_hero": {"action": "call_agent", "config": {"prompt": "Write a compelling hero section for {{.business_name}}, a {{.business_type}}. Include a powerful headline and subheadline.", "agent_type": "content-creator-hero", "target_role": "hero_writer"}, "next_step": "spawn_features_writer", "description": "Generate hero section"}, "generate_contact": {"action": "call_agent", "config": {"prompt": "Write a contact section for {{.business_name}}. Include inviting text that encourages customers to reach out.", "agent_type": "content-creator-contact", "target_role": "contact_writer"}, "next_step": "spawn_cta_writer", "description": "Generate contact section"}, "spawn_cta_writer": {"action": "spawn_agent", "config": {"role": "cta_writer", "agent_type": "content-creator-cta"}, "next_step": "generate_cta", "description": "Spawn agent for call-to-action section"}, "generate_features": {"action": "call_agent", "config": {"prompt": "List 3-4 key features or benefits of {{.business_name}}, a {{.business_type}}. Make each feature clear and customer-focused.", "agent_type": "content-creator-features", "target_role": "features_writer"}, "next_step": "spawn_testimonials_writer", "description": "Generate features section"}, "spawn_hero_writer": {"action": "spawn_agent", "config": {"role": "hero_writer", "agent_type": "content-creator-hero"}, "next_step": "generate_hero", "description": "Spawn agent for hero section"}, "spawn_contact_writer": {"action": "spawn_agent", "config": {"role": "contact_writer", "agent_type": "content-creator-contact"}, "next_step": "generate_contact", "description": "Spawn agent for contact section"}, "generate_testimonials": {"action": "call_agent", "config": {"prompt": "Create 2 realistic customer testimonials for {{.business_name}}, a {{.business_type}}. Make them authentic and specific.", "agent_type": "content-creator-testimonials", "target_role": "testimonials_writer"}, "next_step": "spawn_contact_writer", "description": "Generate testimonials section"}, "spawn_features_writer": {"action": "spawn_agent", "config": {"role": "features_writer", "agent_type": "content-creator-features"}, "next_step": "generate_features", "description": "Spawn agent for features section"}, "aggregate_all_sections": {"action": "aggregate_webpage", "config": {"wrapper": {"html_foot": "</body>\n</html>", "html_head": "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n<title>Website</title>\n<style>\nbody { font-family: system-ui, sans-serif; line-height: 1.6; max-width: 1200px; margin: 0 auto; padding: 20px; }\nsection { margin: 40px 0; padding: 20px; }\n.hero-section { background: #f5f5f5; }\n</style>\n</head>\n<body>"}, "section_order": ["generate_hero", "generate_features", "generate_testimonials", "generate_contact", "generate_cta"], "response_fields": ["generate_hero", "generate_features", "generate_testimonials", "generate_contact", "generate_cta"], "add_section_tags": true}, "next_step": "complete", "description": "Combine all website sections into HTML"}, "spawn_testimonials_writer": {"action": "spawn_agent", "config": {"role": "testimonials_writer", "agent_type": "content-creator-testimonials"}, "next_step": "generate_testimonials", "description": "Spawn agent for testimonials section"}}, "start_step": "spawn_hero_writer"} |           0 |       6 | 2025-10-13 16:26:06.877204 | 2025-10-16 13:38:55.358732

--

example content hero agent with sub research agent


                  id                  |         type         |    display_name     |                                            description                                            |  category   |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     default_config                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | is_active |          created_at           |          updated_at           | deleted_at |                     capabilities                      |       image_repository       | image_tag | command |                                          resources                                           |                                                                            topics                                                                            |                                            health_config                                            | env_vars | version | previous_version_id | task_workflow | orchestrator_workflow | orchestration_workflow |                delegation_preferences                 
--------------------------------------+----------------------+---------------------+---------------------------------------------------------------------------------------------------+-------------+-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+-----------+-------------------------------+-------------------------------+------------+-------------------------------------------------------+------------------------------+-----------+---------+----------------------------------------------------------------------------------------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------+-----------------------------------------------------------------------------------------------------+----------+---------+---------------------+---------------+-----------------------+------------------------+-------------------------------------------------------
d8b5e2c1-7539-4965-987b-25a4834ccf38 | content-creator-hero | Hero Section Writer | Specialized in writing compelling hero sections with powerful headlines and engaging subheadlines | data-driven | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return hero content"}, "call_researcher": {"action": "call_agent", "config": {"prompt": "Research background information about {{.business_type}} businesses", "agent_type": "content-researcher", "input_data": {"business_type": "{{.input_data.business_type}}"}, "target_role": "researcher"}, "next_step": "generate_hero_content", "description": "Get research data"}, "spawn_researcher": {"action": "spawn_agent", "config": {"role": "researcher", "agent_type": "content-researcher"}, "next_step": "call_researcher", "description": "Spawn research agent"}, "generate_hero_content": {"action": "execute_llm_prompt", "config": {"input_fields": ["call_researcher", "input_data"], "prompt_template": "Using this research: {{.call_researcher.result}}\n\nWrite a compelling hero section for {{.business_name}}, a {{.business_type}}. Include a powerful headline and engaging subheadline that captures attention and communicates the core value proposition."}, "next_step": "complete", "description": "Generate hero section with research"}}, "start_step": "spawn_researcher"}, "ai_service": {"model": "claude-3-5-sonnet-20241022", "provider": "anthropic", "api_key_env_var": "ANTHROPIC_API_KEY"}, "max_tokens": 2000, "temperature": 0.7, "processing_mode": "task"} | t         | 2025-10-14 19:00:09.806276+00 | 2025-10-30 18:08:00.611217+00 |            | ["content", "hero", "headlines", "value-proposition"] | docker.io/aqls/agent-chassis | v1.0.407  |         | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}} | {"error": "system.errors.content-creator-hero", "process": "system.agent.content-creator-hero.process", "response": "system.responses.content-creator-hero"} | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30} | []       |       1 |                     |               |                       |                        | {"fallback_to_self": true, "prefer_delegation": true}
(1 row)

--

full initiation script
# 1. FIRST - Stop all consumers by scaling down deployments
kubectl -n ai-persona-system scale deployment/agent-chassis --replicas=0
# Wait for pods to terminate
kubectl -n ai-persona-system wait --for=delete pod -l app=agent-chassis --timeout=60s

# 3. Clear database tables
kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c "TRUNCATE TABLE processed_messages, orchestration_states, pending_requests CASCADE;"

# 4. Delete ALL spawned agent jobs
kubectl -n ai-persona-system delete jobs -l spawned-by=orchestrator

# 5. List and clean up job topics (they follow pattern: job.<correlation>.<orchestration>.<step>)
# List all job topics
kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-topics.sh \
--bootstrap-server localhost:9092 \
--list | grep "^job\."

# Delete all job topics (be careful!)
kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- bash -c '
for topic in $(/opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list | grep "^job\."); do
echo "Deleting topic: $topic"
/opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic "$topic"
done
'

# delete initial topics
kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-topics.sh \
--bootstrap-server localhost:9092 \
--delete \
--topic system.agent.generic.requests

kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-topics.sh \
--bootstrap-server localhost:9092 \
--delete \
--topic system.agent.generic.responses

kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-topics.sh \
--bootstrap-server localhost:9092 \
--delete \
--topic system.responses.generic

# reset all offsets
echo 'resetting all offsets'
kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-consumer-groups.sh \
--bootstrap-server localhost:9092 \
--reset-offsets \
--to-earliest \
--all-groups   \
--all-topics    \
--execute

kubectl -n ai-persona-system delete jobs -l agent-type=calculator
kubectl -n ai-persona-system delete pods -l agent-type=calculator

# 6. Scale back up
kubectl -n ai-persona-system scale deployment/agent-chassis --replicas=1


# =======

sleep 5
kubectl -n ai-persona-system scale deployment/agent-chassis --replicas=0

kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-consumer-groups.sh \
--bootstrap-server localhost:9092 \
--reset-offsets \
--to-earliest \
--all-groups   \
--all-topics    \
--execute

kubectl -n ai-persona-system scale deployment/agent-chassis --replicas=1




====
=========




=== run === robot hands website =====

# Generate IDs
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="robot-site-$(date +%H%M%S)"
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)
STEP_NAME="client_step_website_request"
CLIENT_ID="demo_client"

echo "========================================="
echo "Robot Hands Website Test"
echo "========================================="
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Request ID:       $REQUEST_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Name:             $ORCHESTRATION_NAME"
echo "  Client            $CLIENT_ID"
echo "  Message ID:       $MESSAGE_ID"
echo "  Agent ID:         $AGENT_ID"
echo "  Time:             $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================="

# Send message

kubectl -n kafka run -i --rm kcat-producer \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H orchestration_name=$ORCHESTRATION_NAME \
-H step_name=$STEP_NAME \
-H client_id=$CLIENT_ID \
-H message_type=request \
-H action=orchestrate \
-H from_agent_type=user \
-H from_agent_id=$AGENT_ID \
-H responses_topic=system.generic.responses <<EOF
{"action":"orchestrate","config":{"group_type":"robot-hands-website"},"input_data":{"business_type":"robotics automation company","business_name":"PrecisionBot Systems"}}
EOF

echo ""
echo "Message sent!"
echo ""
echo "Monitor:"
echo "  Orchestrator: kubectl logs -f deployment/agent-chassis -n agent-system | grep '$ORCHESTRATION_ID'"
echo "  Database: psql ... -c \"SELECT status, current_step FROM orchestration_states WHERE orchestration_id = '$ORCHESTRATION_ID';\""
echo ""


=== run === multi section website builder =====

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="multi-site-$(date +%H%M%S)"
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)
STEP_NAME="client_step_website_request"
CLIENT_ID="demo_client"

echo "========================================="
echo "Multi-Section Website Builder Test"
echo "========================================="
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Request ID:       $REQUEST_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Name:             $ORCHESTRATION_NAME"
echo "  Time:             $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================="

kubectl -n kafka run -i --rm kcat-producer \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H orchestration_name=$ORCHESTRATION_NAME \
-H step_name=$STEP_NAME \
-H client_id=$CLIENT_ID \
-H message_type=request \
-H action=orchestrate \
-H from_agent_type=user \
-H from_agent_id=$AGENT_ID \
-H responses_topic=system.generic.responses <<EOF
{"action":"orchestrate","config":{"group_type":"multi-section-website-builder"},"input_data":{"business_type":"artisanal bakery","business_name":"Golden Crust Bakery"}}
EOF

echo ""
echo "Message sent. Waiting 20 seconds for processing..."
sleep 20

echo ""
echo "========================================="
echo "Checking Spawned Jobs"
echo "========================================="
kubectl -n ai-persona-system get jobs | grep content-creator | tail -10

echo ""
echo "========================================="
echo "Checking Spawned Pods"
echo "========================================="
kubectl -n ai-persona-system get pods | grep content-creator | tail -10

echo ""
echo "========================================="
echo "Checking Orchestration State in Database"
echo "========================================="
kubectl -n ai-persona-system exec -it postgres-clients-0 -- \
psql -U clients_user -d clients_db -c \
"SELECT
orchestration_id,
status,
current_step,
owner_agent_type,
to_char(created_at, 'HH24:MI:SS') as created,
to_char(updated_at, 'HH24:MI:SS') as updated
FROM orchestration_states
WHERE correlation_id = '$CORRELATION_ID'
ORDER BY created_at DESC
LIMIT 10;"

echo ""
echo "========================================="
echo "Checking for Any Errors in Orchestration"
echo "========================================="
kubectl -n ai-persona-system exec -it postgres-clients-0 -- \
psql -U clients_user -d clients_db -c \
"SELECT
orchestration_id,
status,
error
FROM orchestration_states
WHERE correlation_id = '$CORRELATION_ID'
AND (status = 'FAILED' OR error IS NOT NULL);"

echo ""
echo "========================================="
echo "Monitoring Response Topic"
echo "========================================="
echo "Listening for final response on system.responses.generic..."
echo "Will show correlation_id header and message body"
echo "Press Ctrl+C when you see the final response"
echo ""

kubectl -n kafka run -i --rm kcat-consumer \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -C \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.responses.generic \
-o end \
-e \
-f 'CORRELATION: %h{correlation_id}\nBODY: %s\n\n========================================\n\n'





make quick-agent-update ENVIRONMENT=production REGION=uk001
make build-backend push-backend deploy-core deploy-agents redeploy-agents ENVIRONMENT=production REGION=uk001

listen on a topic
kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 --   /opt/kafka/bin/kafka-console-consumer.sh   --bootstrap-server localhost:9092   --topic system.responses.generic   --from-beginning   --property print.headers=true

test message:
TIMESTAMP=$(date +%s)
ISOTIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="multi-site-$(date +%H%M%S)"
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)
STEP_NAME="client_step_website_request"
CLIENT_ID="demo_client"
TEST_MESSAGE='{"headers":{"correlation_id":"test-'${TIMESTAMP}'","client_id":"demo_client","message_id":"test-msg-'${TIMESTAMP}'","message_type":"request","action":"generate","timestamp":"'${ISOTIME}'","fuel_budget":1000},"body":{"action":"generate","input_data":{"prompt":"A beautiful sunset over mountains, photorealistic, 4k","width":1024,"height":1024,"style":""},"parent_responses_topic":"test.responses","reply_to_topic":"test.responses"}}'
echo "$TEST_MESSAGE" | kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-console-producer.sh \
--bootstrap-server localhost:9092 \
--topic system.adapter.image-generator.requests


all on one line
TIMESTAMP=$(date +%s); ISOTIME=$(date -u +%Y-%m-%dT%H:%M:%SZ); CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid); REQUEST_ID=$(cat /proc/sys/kernel/random/uuid); MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid); ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid); AGENT_ID=$(cat /proc/sys/kernel/random/uuid); ORCHESTRATION_NAME="multi-site-${TIMESTAMP}"; STEP_NAME="client_step_image_request"; CLIENT_ID="demo_client"; TEST_MESSAGE='{"headers":{"correlation_id":"'"${CORRELATION_ID}"'","request_id":"'"${REQUEST_ID}"'","message_id":"'"${MESSAGE_ID}"'","orchestration_id":"'"${ORCHESTRATION_ID}"'","orchestration_name":"'"${ORCHESTRATION_NAME}"'","step_name":"'"${STEP_NAME}"'","client_id":"'"${CLIENT_ID}"'","message_type":"request","action":"generate","timestamp":"'"${ISOTIME}"'","fuel_budget":1000,"from_agent_type":"user","from_agent_id":"user-001","from_agent_version":"1.0","to_agent_type":"image-generator","responses_topic":"test.responses","requests_topic":"system.adapter.image-generator.requests","status":"ok","retry_count":0,"is_error":"false"},"body":{"action":"generate","input_data":{"prompt":"A beautiful sunset over mountains, photorealistic, 4k","width":1024,"height":1024,"style":""},"parent_responses_topic":"test.responses","reply_to_topic":"test.responses"}}'; echo "$TEST_MESSAGE" | kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- /opt/kafka/bin/kafka-console-producer.sh --bootstrap-server localhost:9092 --topic system.adapter.image-generator.requests


SELECT * FROM agent_group_definitions WHERE group_type='robot-hands-website';

  
