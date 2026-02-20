// FILE: platform/orchestration/actions/deploy_image_asset_action.go
// Deploys an image asset: downloads from storage, optimizes, commits to git

package actions

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// DeployImageAssetInputSpec defines the standard input contract.
//
// New callers (asset-deployer) use input_fields: ["s3_uri", "deploy_path", "purpose", "domain"]
// Existing callers (pageflow-builder) use config keys: "uri_field", "domain_field", "purpose"
// The Deprecated map bridges old config keys to new field names.
var DeployImageAssetInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"domain"},
	Optional: []string{"s3_uri", "deploy_path", "purpose"},
	Defaults: map[string]interface{}{
		"purpose": "hero",
	},
	Deprecated: map[string]string{
		"uri_field":         "s3_uri",
		"domain_field":      "domain",
		"purpose_field":     "purpose",
		"deploy_path_field": "deploy_path",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("deploy_image_asset", DeployImageAssetInputSpec)
}

// DeployImageAssetAction downloads an image from storage, optimizes it, and commits via git
//
// Inputs (via ActionInputSpec):
//   - domain (required): site domain for git commit
//   - s3_uri (optional): storage URI — if not provided, falls back to findStorageURI lookup
//   - purpose (optional, default "hero"): image purpose — controls resize dimensions via ImagePurposes
//   - deploy_path (optional): override output path in git (e.g. "assets/images/departments/dept-foo.jpg")
func DeployImageAssetAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "deploy_image_asset"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Extract inputs using standard pattern (checklist: ActionInputSpec)
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		config,
		DeployImageAssetInputSpec,
		logger,
	)
	if err != nil {
		// Don't hard-fail — fall back to legacy extraction for backward compat
		logger.Warn("ExtractActionInputs returned error, falling back to legacy extraction",
			zap.Error(err))
		inputs = &datahelpers.ActionInputs{Values: map[string]interface{}{}}
	}

	// Resolve purpose
	purpose := inputs.Get("purpose")
	if purpose == "" {
		if p, ok := config["purpose"].(string); ok && p != "" {
			// Legacy: static purpose in config (pageflow-builder pattern)
			purpose = p
		} else {
			purpose = "hero"
		}
	}

	// Resolve storage URI
	// Priority: inputs.Get("s3_uri") (from input_fields) → findStorageURI (legacy lookup)
	storageURI := inputs.Get("s3_uri")
	if storageURI == "" {
		storageURI = findStorageURI(params.CollectedData, config, purpose, logger)
	}
	if storageURI == "" {
		logger.Warn("No storage URI found for image asset",
			zap.String("purpose", purpose),
			zap.Int("collected_data_key_count", len(params.CollectedData)))
		return map[string]interface{}{
			"deployed": false,
			"skipped":  true,
			"reason":   "no storage URI found for " + purpose,
		}, nil
	}

	// Resolve domain
	// Priority: inputs.Get("domain") (from input_fields) → findDomain (legacy lookup)
	domain := inputs.Get("domain")
	if domain == "" {
		domain = findDomain(params.CollectedData, config)
	}
	if domain == "" {
		return nil, fmt.Errorf("domain not found")
	}

	// Get storage client
	if params.StorageClient == nil {
		return nil, fmt.Errorf("storage client not available")
	}

	s3Client, ok := params.StorageClient.(*storage.S3Client)
	if !ok {
		return nil, fmt.Errorf("storage client is not S3Client type")
	}

	logger.Info("Deploying image asset",
		zap.String("purpose", purpose),
		zap.String("storage_uri", storageURI),
		zap.String("domain", domain))

	// Download, optimize, and prepare paths using storage package
	processed, err := s3Client.DownloadOptimizeAndPrepare(ctx, storageURI, purpose, logger)
	if err != nil {
		logger.Error("Failed to process image", zap.Error(err))
		return map[string]interface{}{
			"deployed": false,
			"error":    err.Error(),
		}, nil
	}

	// Allow deploy_path to override the output path.
	// Purpose still controls resize dimensions; deploy_path controls where the file goes.
	// Priority: inputs.Get("deploy_path") (from input_fields) → config["deploy_path"] (static)
	deployPath := inputs.Get("deploy_path")
	if deployPath == "" {
		if dp, ok := config["deploy_path"].(string); ok && dp != "" {
			deployPath = dp
		}
	}
	if deployPath != "" {
		processed.Paths = storage.AssetPaths{
			FilePath:    deployPath,
			RelativeURL: "/" + deployPath,
			Filename:    filepath.Base(deployPath),
		}
		logger.Info("Using custom deploy_path",
			zap.String("deploy_path", deployPath))
	}

	logger.Info("Image processed",
		zap.Int("size_bytes", len(processed.Data)),
		zap.String("content_type", processed.ContentType),
		zap.String("output_path", processed.Paths.FilePath))

	// Build files map for git commit (base64 encoded for binary)
	filesMap := map[string]interface{}{
		processed.Paths.FilePath: map[string]interface{}{
			"content":  base64.StdEncoding.EncodeToString(processed.Data),
			"encoding": "base64",
		},
	}

	// Send to git adapter
	result, err := sendGitCommitRequest(ctx, params, domain, filesMap, purpose, logger)
	if err != nil {
		return nil, err
	}

	// Add image URL to result
	result["image_url"] = processed.Paths.RelativeURL
	result["output_path"] = processed.Paths.FilePath
	result["size_bytes"] = len(processed.Data)

	return result, nil
}

// findStorageURI looks for the storage URI in multiple locations
func findStorageURI(collectedData map[string]interface{}, config map[string]interface{}, purpose string, logger *zap.Logger) string {
	// Priority 1: Config-specified field
	if uriField, ok := config["uri_field"].(string); ok && uriField != "" {
		if uri := datahelpers.ExtractNestedFieldString(collectedData, uriField); uri != "" {
			logger.Debug("Found URI via config uri_field", zap.String("path", uriField))
			return uri
		}
	}

	// Priority 2: {purpose}_uri (set by StoreAssetAction)
	if uri := datahelpers.ExtractNestedFieldString(collectedData, purpose+"_uri"); uri != "" {
		logger.Debug("Found URI at purpose_uri", zap.String("path", purpose+"_uri"))
		return uri
	}

	// Priority 3: {purpose}_result.image_uri pattern
	resultField := purpose + "_result.image_uri"
	if uri := datahelpers.ExtractNestedFieldString(collectedData, resultField); uri != "" {
		logger.Debug("Found URI at result.image_uri", zap.String("path", resultField))
		return uri
	}

	// Priority 4: Deep nested from image-generator agent response
	// Pattern: {purpose}_result.response.generate.response.image_uri
	deepField := purpose + "_result.response.generate.response.image_uri"
	if uri := datahelpers.ExtractNestedFieldString(collectedData, deepField); uri != "" {
		logger.Debug("Found URI at deep nested path", zap.String("path", deepField))
		return uri
	}

	// Priority 5: Slightly less nested pattern
	// Pattern: {purpose}_result.response.image_uri
	lessDeepField := purpose + "_result.response.image_uri"
	if uri := datahelpers.ExtractNestedFieldString(collectedData, lessDeepField); uri != "" {
		logger.Debug("Found URI at response.image_uri", zap.String("path", lessDeepField))
		return uri
	}

	// Priority 6: Check generate_hero_image step directly (step name pattern)
	generateStepField := "generate_" + purpose + "_image.response.generate.response.image_uri"
	if uri := datahelpers.ExtractNestedFieldString(collectedData, generateStepField); uri != "" {
		logger.Debug("Found URI at generate step", zap.String("path", generateStepField))
		return uri
	}

	// Priority 7: {purpose}_stored.asset_url if it's a storage URI
	storedField := purpose + "_stored.asset_url"
	if url := datahelpers.ExtractNestedFieldString(collectedData, storedField); storage.IsS3URI(url) {
		logger.Debug("Found URI at stored.asset_url", zap.String("path", storedField))
		return url
	}

	// Log what we searched for debugging
	logger.Debug("Storage URI not found, searched paths",
		zap.String("purpose", purpose),
		zap.Strings("searched", []string{
			purpose + "_uri",
			purpose + "_result.image_uri",
			purpose + "_result.response.generate.response.image_uri",
			purpose + "_result.response.image_uri",
			"generate_" + purpose + "_image.response.generate.response.image_uri",
			purpose + "_stored.asset_url",
		}))

	return ""
}

// findDomain extracts domain from collected data
func findDomain(collectedData map[string]interface{}, config map[string]interface{}) string {
	// Try config-specified field
	domainField := "site_record.domain"
	if df, ok := config["domain_field"].(string); ok && df != "" {
		domainField = df
	}

	if domain := datahelpers.ExtractNestedFieldString(collectedData, domainField); domain != "" {
		return domain
	}

	// Fallback to input_data.domain
	return datahelpers.ExtractNestedFieldString(collectedData, "input_data.domain")
}

// sendGitCommitRequest sends the commit request to git-adapter
func sendGitCommitRequest(ctx context.Context, params ActionParams, domain string, filesMap map[string]interface{}, purpose string, logger *zap.Logger) (map[string]interface{}, error) {
	if params.Producer == nil {
		return nil, fmt.Errorf("kafka producer not available")
	}

	// Get execution context
	correlationID := params.ExecutionContext.CorrelationID
	orchestrationID := params.ExecutionContext.OrchestrationID
	clientID := params.ExecutionContext.ClientID
	responsesTopic := params.ExecutionContext.ResponsesTopic

	// Fallback for responses topic
	if responsesTopic == "" {
		if alt, ok := params.CollectedData["__my_responses_topic__"].(string); ok && alt != "" {
			responsesTopic = alt
		}
	}
	if responsesTopic == "" {
		if parent, ok := params.CollectedData["__parent_responses_topic__"].(string); ok && parent != "" {
			responsesTopic = parent
		}
	}

	newRequestID := uuid.New().String()

	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":    correlationID,
			"orchestration_id":  orchestrationID,
			"client_id":         clientID,
			"step_name":         params.ExecutionContext.StepName,
			"request_id":        newRequestID,
			"message_type":      "request",
			"sender_agent_type": params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":   orchestrationID,
			"sender_pod_name":   params.ExecutionContext.Sender.PodName,
			"responses_topic":   responsesTopic,
		},
		"body": map[string]interface{}{
			"action": "commit",
			"data": map[string]interface{}{
				"repo_name":      "sites",
				"domain":         domain,
				"files":          filesMap,
				"commit_message": fmt.Sprintf("Deploy %s image for %s", purpose, domain),
			},
		},
	}

	// Serialize and send
	requestBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal git request: %w", err)
	}

	adapterTopic := "system.adapter.git.requests"
	headers := map[string]string{
		"correlation_id":   correlationID,
		"orchestration_id": orchestrationID,
		"request_id":       newRequestID,
		"message_type":     "request",
	}

	err = params.Producer.Produce(ctx, adapterTopic, headers, []byte(correlationID), requestBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to send git commit request: %w", err)
	}

	logger.Info("Image deploy request sent to git adapter",
		zap.String("request_id", newRequestID),
		zap.String("domain", domain))

	return map[string]interface{}{
		"deployed":       true,
		"await_response": true,
		"request_id":     newRequestID,
		"purpose":        purpose,
		"domain":         domain,
	}, nil
}
